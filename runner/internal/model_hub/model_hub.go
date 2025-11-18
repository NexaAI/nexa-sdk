package model_hub

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/NexaAI/nexa-sdk/runner/internal/types"
	"github.com/bytedance/sonic"
)

type ModelFileInfo struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

type ModelHub interface {
	CheckAvailable(ctx context.Context, modelName string) error
	MaxConcurrency() int
	ModelInfo(ctx context.Context, modelName string) ([]ModelFileInfo, error)
	GetFileContent(ctx context.Context, modelName, fileName string, offset, limit int64, writer io.Writer) error
}

var hubs = []ModelHub{
	NewVolces(false),
	NewS3(),
	NewHuggingFace(),
}

var errUnavailable = fmt.Errorf("no model hub contains the model")

// Specify hub to use
func SetHub(h ModelHub) {
	hubs = []ModelHub{h}
}

// list model files

func ModelInfo(ctx context.Context, modelName string) ([]ModelFileInfo, *types.ModelManifest, error) {
	slog.Debug("fetching model info", "model", modelName)

	hub, err := getHub(ctx, modelName)
	if err != nil {
		return nil, nil, err
	}

	files, err := hub.ModelInfo(ctx, modelName)
	if err != nil {
		return nil, nil, err
	}

	// check manifest available
	const manifestFile = "nexa.manifest"
	var hasManifest bool
	for i := 0; i < len(files); i++ {
		if files[i].Name == manifestFile {
			files = append(files[:i], files[i+1:]...)
			hasManifest = true
			break
		}
	}
	if !hasManifest {
		return files, nil, nil
	}

	// parse manifest
	data, err := GetFileContent(ctx, modelName, manifestFile)
	if err != nil {
		slog.Warn("failed to get manifest file, ignore", "error", err)
		return nil, nil, err
	}

	var manifest types.ModelManifest
	if err := sonic.Unmarshal(data, &manifest); err != nil {
		slog.Warn("failed to parse manifest file, ignore", "error", err)
		return nil, nil, err
	}

	return files, &manifest, nil

}

// Get single file content

func GetFileContent(ctx context.Context, modelName, fileName string) ([]byte, error) {
	slog.Debug("fetching file content", "model", modelName, "file", fileName)

	hub, err := getHub(ctx, modelName)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	if err := hub.GetFileContent(ctx, modelName, fileName, 0, 0, buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Batch download

type downloadTask struct {
	OutputPath string

	ModelName string
	FileName  string
	Offset    int64
	Limit     int64
}

const (
	minChunkSize = 16 * 1024 * 1024 // 16MiB
)

func StartDownload(ctx context.Context, modelName, outputPath string, files []ModelFileInfo) (resChan chan types.DownloadInfo, errChan chan error) {
	slog.Info("Starting download", "model", modelName, "outputPath", outputPath, "files", files)

	hub, err := getHub(ctx, modelName)

	if err != nil {
		resCh := make(chan types.DownloadInfo)
		errCh := make(chan error, 1)
		close(resCh)
		errCh <- err
		close(errCh)
		return resCh, errCh
	}

	maxConcurrency := hub.MaxConcurrency()
	resCh := make(chan types.DownloadInfo)
	errCh := make(chan error, maxConcurrency)

	slog.Info("GetHub", "hub", reflect.TypeOf(hub), "maxConcurrency", maxConcurrency)

	go func() {
		defer close(errCh)
		defer close(resCh)

		var totalSize int64
		for _, f := range files {
			totalSize += f.Size
		}

		// Load existing download progress if available
		progress, err := types.LoadDownloadProgress(outputPath)
		if err != nil {
			slog.Warn("Failed to load download progress, starting fresh", "error", err)
			progress = nil
		}

		// Initialize progress if not exists
		if progress == nil {
			progress = types.NewDownloadProgress(modelName, totalSize)
			for _, f := range files {
				progress.AddFile(f.Name, f.Size)
			}
			slog.Info("Starting fresh download", "total_size", totalSize)
		} else {
			slog.Info("Resuming download", "already_downloaded", progress.Downloaded, "total_size", totalSize)
		}

		var downloaded int64 = progress.Downloaded

		// create tasks
		tasks := make(chan downloadTask, maxConcurrency)
		nctx, cancel := context.WithCancel(ctx)

		// Mutex for progress updates
		var progressMutex sync.Mutex

		var wg1 sync.WaitGroup
		wg1.Add(1)
		go func() {
			defer wg1.Done()
			for _, f := range files {
				err := os.MkdirAll(filepath.Dir(filepath.Join(outputPath, f.Name)), 0o755)
				if err != nil {
					errCh <- fmt.Errorf("failed to create directory: %v, %s", err, f.Name)
					cancel()
					return
				}

				// Get file progress
				fileProgress := progress.Files[f.Name]
				if fileProgress == nil {
					fileProgress = &types.FileProgress{
						FileName:       f.Name,
						TotalSize:      f.Size,
						CompletedRange: make([]types.CompletedRange, 0),
					}
					progress.Files[f.Name] = fileProgress
				}

				// Skip if file already complete
				if fileProgress.IsComplete() {
					slog.Info("File already downloaded, skipping", "file", f.Name)
					continue
				}

				// Get missing ranges
				chunkSize := max(minChunkSize, f.Size/128)
				missingRanges := fileProgress.GetMissingRanges(chunkSize)
				
				slog.Info("Download file", "name", f.Name, "size", f.Size, "chunkSize", chunkSize, "missingRanges", len(missingRanges))

				// Create tasks for missing ranges only
				for _, missingRange := range missingRanges {
					for offset := missingRange.Start; offset < missingRange.End; offset += chunkSize {
						task := downloadTask{
							OutputPath: outputPath,
							ModelName:  modelName,
							FileName:   f.Name,
							Offset:     offset,
							Limit:      min(chunkSize, missingRange.End-offset),
						}

						// send chunk
						select {
						case tasks <- task:
						case <-nctx.Done():
							slog.Warn("download canceled", "error", nctx.Err())
							return
						}
					}
				}
			}
		}()

		// Periodic progress saver
		progressTicker := time.NewTicker(5 * time.Second)
		defer progressTicker.Stop()

		saveChan := make(chan struct{}, 1)
		go func() {
			for {
				select {
				case <-progressTicker.C:
					select {
					case saveChan <- struct{}{}:
					default:
					}
				case <-nctx.Done():
					return
				}
			}
		}()

		// concurrent control
		var wg2 sync.WaitGroup
		for range maxConcurrency {
			wg2.Add(1)
			go func() {
				defer wg2.Done()

				for task := range tasks {
					err := doTask(nctx, hub, task)
					if err != nil {
						slog.Error("Download task failed", "task", task, "error", err)
						errCh <- err
						cancel()
						return
					}

					// Update progress
					progressMutex.Lock()
					fileProgress := progress.Files[task.FileName]
					fileProgress.MarkRangeComplete(task.Offset, task.Offset+task.Limit)
					progress.Downloaded = calculateTotalDownloaded(progress)
					progressMutex.Unlock()

					resCh <- types.DownloadInfo{
						TotalDownloaded: atomic.AddInt64(&downloaded, task.Limit),
						TotalSize:       totalSize,
					}

					// Try to save progress periodically
					select {
					case <-saveChan:
						progressMutex.Lock()
						if err := progress.Save(outputPath); err != nil {
							slog.Warn("Failed to save progress", "error", err)
						} else {
							slog.Debug("Progress saved", "downloaded", progress.Downloaded)
						}
						progressMutex.Unlock()
					default:
					}
				}
			}()
		}

		wg1.Wait()
		close(tasks)
		wg2.Wait()

		// Final progress save
		progressMutex.Lock()
		if err := progress.Save(outputPath); err != nil {
			slog.Warn("Failed to save final progress", "error", err)
		}
		progressMutex.Unlock()

		// Cleanup progress file on successful completion
		if progress.Downloaded >= totalSize {
			if err := types.CleanupProgress(outputPath); err != nil {
				slog.Warn("Failed to cleanup progress file", "error", err)
			} else {
				slog.Info("Download completed, progress file cleaned up")
			}
		}

		cancel()
	}()

	return resCh, errCh
}

// calculateTotalDownloaded sums up downloaded bytes across all files
func calculateTotalDownloaded(progress *types.DownloadProgress) int64 {
	var total int64
	for _, fp := range progress.Files {
		total += fp.Downloaded
	}
	return total
}

func getHub(ctx context.Context, modelName string) (ModelHub, error) {
	for _, h := range hubs {
		if err := h.CheckAvailable(ctx, modelName); err != nil {
			slog.Warn("hub not available, try next", "hub", reflect.TypeOf(h), "err", err)
		} else {
			slog.Info("hub available", "hub", reflect.TypeOf(h))
			return h, nil
		}
	}
	return nil, errUnavailable
}

func doTask(ctx context.Context, hub ModelHub, task downloadTask) error {
	slog.Debug("Downloading chunk", "OutputPath", task.OutputPath, "model", task.ModelName, "file", task.FileName, "offset", task.Offset, "limit", task.Limit)

	file, err := os.OpenFile(filepath.Join(task.OutputPath, task.FileName), os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}

	_, err = file.Seek(task.Offset, io.SeekStart)
	if err != nil {
		return err
	}

	err = hub.GetFileContent(ctx, task.ModelName, task.FileName, task.Offset, task.Limit, file)
	if err != nil {
		file.Close()
		return err
	}

	return file.Close()
}
