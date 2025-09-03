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

	"github.com/NexaAI/nexa-sdk/runner/internal/types"
)

type ModelFileInfo struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

type ModelHub interface {
	CheckAvailable(ctx context.Context, modelName string) error
	ModelInfo(ctx context.Context, modelName string) ([]ModelFileInfo, error)
	GetFileContent(ctx context.Context, modelName, fileName string, offset, limit int64, writer io.Writer) error
}

var hubs = []ModelHub{
	NewVocles(),
	//NewHuggingFace(),
}

var errNotAvailable = fmt.Errorf("no available model hub")

// list model files

func ModelInfo(ctx context.Context, modelName string) ([]ModelFileInfo, *types.ModelManifest, error) {
	slog.Debug("fetching model info", "model", modelName)

	for _, hub := range hubs {
		if err := hub.CheckAvailable(ctx, modelName); err != nil {
			slog.Warn("hub not available, try next", "hub", reflect.TypeOf(hub), "err", err)
			continue
		}
		slog.Info("hub available", "hub", reflect.TypeOf(hub))

		files, err := hub.ModelInfo(ctx, modelName)
		if err != nil {
			return nil, nil, err
		}

		// if !slices.Contains(files, "nexa.manifest") {
		// }

		return files, nil, nil

	}
	return nil, nil, errNotAvailable
}

// Get single file content

func GetFileContent(ctx context.Context, modelName, fileName string) ([]byte, error) {
	slog.Debug("fetching file content", "model", modelName, "file", fileName)

	for _, hub := range hubs {
		if err := hub.CheckAvailable(ctx, modelName); err != nil {
			slog.Warn("hub not available, try next", "hub", reflect.TypeOf(hub), "err", err)
			continue
		}
		slog.Info("hub available", "hub", reflect.TypeOf(hub))

		buf := bytes.NewBuffer(nil)

		if err := hub.GetFileContent(ctx, modelName, fileName, 0, 0, buf); err != nil {
			return nil, err
		}

		return buf.Bytes(), nil
	}
	return nil, errNotAvailable
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
	minChunkSize   = 8 * 1024 * 1024 // 8MB
	maxConcurrency = 8
)

func StartDownload(ctx context.Context, modelName, outputPath string, files []ModelFileInfo) (resChan chan types.DownloadInfo, errChan chan error) {
	slog.Debug("Starting download", "model", modelName, "outputPath", outputPath, "files", files)

	var hub ModelHub
	for _, h := range hubs {
		if err := h.CheckAvailable(ctx, modelName); err != nil {
			slog.Warn("hub not available, try next", "hub", hub, "err", err)
		} else {
			hub = h
			break
		}
	}

	resCh := make(chan types.DownloadInfo)
	errCh := make(chan error, maxConcurrency)

	// no available hub
	if hub == nil {
		close(resCh)
		errCh <- errNotAvailable
		close(errCh)
		return resCh, errCh
	}

	go func() {
		defer close(errCh)
		defer close(resCh)

		var downloaded int64
		var totalSize int64
		for _, f := range files {
			totalSize += f.Size
		}

		// create tasks
		tasks := make(chan downloadTask, maxConcurrency)
		nctx, cancel := context.WithCancel(ctx)

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

				// create task
				task := downloadTask{
					OutputPath: outputPath,
					ModelName:  modelName,
					FileName:   f.Name,
				}

				// enqueue tasks
				chunkSize := max(minChunkSize, f.Size/128)
				slog.Info("Downlaod file", "name", f.Name, "size", f.Size, "chunkSize", chunkSize)
				for task.Offset = 0; task.Offset < f.Size; task.Offset += chunkSize {
					task.Limit = min(chunkSize, f.Size-task.Offset)

					// send chunk
					select {
					case tasks <- task:
					case <-nctx.Done():
						slog.Warn("download canceled", "error", nctx.Err())
						return
					}
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

					resCh <- types.DownloadInfo{
						TotalDownloaded: atomic.AddInt64(&downloaded, task.Limit),
						TotalSize:       totalSize,
					}
				}
			}()
		}

		wg1.Wait()
		close(tasks)
		wg2.Wait()
		cancel()
	}()

	return resCh, errCh
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
