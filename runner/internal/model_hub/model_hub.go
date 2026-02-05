// Copyright 2024-2026 Nexa AI, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model_hub

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bytedance/sonic"
	"golang.org/x/sync/errgroup"
	"resty.dev/v3"

	"github.com/NexaAI/nexa-sdk/runner/internal/types"
)

type ModelFileInfo struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

type ModelHub interface {
	ChinaMainlandOnly() bool
	MaxConcurrency() int
	CheckAvailable(ctx context.Context, modelName string) error
	ModelInfo(ctx context.Context, modelName string) ([]ModelFileInfo, error)
	GetFileContent(ctx context.Context, modelName, fileName string, offset, limit int64, writer io.Writer) error
}

var hubs = []ModelHub{
	NewVolces(),
	NewModelScope(),
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

// Binary marker file format for tracking completed chunks.
// Layout: [magic 4B][fileSize 8B][chunkSize 8B][totalChunks 4B][chunk0 1B][chunk1 1B]...
// Each chunk byte: 0x00 = pending, 0x01 = complete.
// Concurrent goroutines write to distinct byte offsets via WriteAt, so no mutex is needed.
const (
	markerMagic     = "NXCK"
	markerHeaderLen = 4 + 8 + 8 + 4 // 24 bytes
)

// chunkTracker tracks per-chunk completion state using an in-memory byte slice
// backed by a binary marker file. Each goroutine writes to its own chunk offset,
// eliminating the need for locks.
type chunkTracker struct {
	filePath    string
	fileSize    int64
	chunkSize   int64
	totalChunks int
	chunks      []byte   // 0x00 = pending, 0x01 = complete
	file        *os.File // open handle for single-byte WriteAt calls
}

// loadOrCreateTracker loads an existing binary marker file or creates a fresh one.
// Discards incompatible markers (different fileSize/chunkSize/totalChunks) and corrupted data.
func loadOrCreateTracker(markerPath string, fileSize, chunkSize int64, totalChunks int) *chunkTracker {
	ct := &chunkTracker{
		filePath:    markerPath,
		fileSize:    fileSize,
		chunkSize:   chunkSize,
		totalChunks: totalChunks,
		chunks:      make([]byte, totalChunks),
	}

	data, err := os.ReadFile(markerPath)
	if err == nil && ct.parseMarker(data) {
		completed := 0
		for _, b := range ct.chunks {
			if b == 0x01 {
				completed++
			}
		}
		slog.Info("Loaded chunk marker", "path", markerPath, "completed", completed, "total", totalChunks)
	} else {
		// Create fresh marker file (header + zeroed chunk bytes)
		ct.initMarkerFile()
	}

	// Open file for subsequent single-byte writes
	if f, err := os.OpenFile(markerPath, os.O_WRONLY, 0o644); err == nil {
		ct.file = f
	}

	return ct
}

// parseMarker validates and loads a binary marker file into the in-memory chunks slice.
func (ct *chunkTracker) parseMarker(data []byte) bool {
	if len(data) < markerHeaderLen+ct.totalChunks {
		slog.Warn("Corrupted chunk marker (too short), starting fresh", "path", ct.filePath)
		return false
	}
	if string(data[:4]) != markerMagic {
		slog.Warn("Corrupted chunk marker (bad magic), starting fresh", "path", ct.filePath)
		return false
	}

	storedFileSize := int64(binary.LittleEndian.Uint64(data[4:12]))
	storedChunkSize := int64(binary.LittleEndian.Uint64(data[12:20]))
	storedTotalChunks := int(binary.LittleEndian.Uint32(data[20:24]))

	if storedFileSize != ct.fileSize || storedChunkSize != ct.chunkSize || storedTotalChunks != ct.totalChunks {
		slog.Warn("Incompatible chunk marker, starting fresh",
			"path", ct.filePath,
			"markerFileSize", storedFileSize, "expectedFileSize", ct.fileSize,
			"markerChunkSize", storedChunkSize, "expectedChunkSize", ct.chunkSize)
		return false
	}

	copy(ct.chunks, data[markerHeaderLen:markerHeaderLen+ct.totalChunks])
	return true
}

// initMarkerFile writes a fresh binary marker file (header + zeroed chunk bytes).
func (ct *chunkTracker) initMarkerFile() {
	buf := make([]byte, markerHeaderLen+ct.totalChunks)
	copy(buf[:4], markerMagic)
	binary.LittleEndian.PutUint64(buf[4:12], uint64(ct.fileSize))
	binary.LittleEndian.PutUint64(buf[12:20], uint64(ct.chunkSize))
	binary.LittleEndian.PutUint32(buf[20:24], uint32(ct.totalChunks))

	if err := os.WriteFile(ct.filePath, buf, 0o644); err != nil {
		slog.Warn("Failed to create chunk marker file", "path", ct.filePath, "error", err)
	}
}

// isComplete checks if a specific chunk has been completed.
func (ct *chunkTracker) isComplete(chunkID int) bool {
	if chunkID < 0 || chunkID >= ct.totalChunks {
		return false
	}
	return ct.chunks[chunkID] == 0x01
}

// allComplete checks if all chunks have been completed.
func (ct *chunkTracker) allComplete() bool {
	for _, b := range ct.chunks {
		if b != 0x01 {
			return false
		}
	}
	return true
}

// markComplete marks a chunk as done in memory and writes a single byte to disk.
// Safe for concurrent use: each goroutine writes to a distinct chunkID offset.
func (ct *chunkTracker) markComplete(chunkID int) error {
	if chunkID < 0 || chunkID >= ct.totalChunks {
		return fmt.Errorf("chunkID %d out of range [0, %d)", chunkID, ct.totalChunks)
	}
	ct.chunks[chunkID] = 0x01

	if ct.file == nil {
		return nil
	}
	_, err := ct.file.WriteAt([]byte{0x01}, int64(markerHeaderLen+chunkID))
	if err != nil {
		return fmt.Errorf("failed to write chunk marker: %w", err)
	}
	return nil
}

// close closes the marker file handle without deleting the file.
// Safe to call multiple times.
func (ct *chunkTracker) close() {
	if ct.file != nil {
		ct.file.Close()
		ct.file = nil
	}
}

// remove closes the file handle and deletes the marker file.
func (ct *chunkTracker) remove() error {
	ct.close()
	return os.Remove(ct.filePath)
}

// completedBytes returns the total bytes represented by completed chunks.
func (ct *chunkTracker) completedBytes() int64 {
	var total int64
	for i, b := range ct.chunks {
		if b == 0x01 {
			if i == ct.totalChunks-1 {
				// Last chunk may be smaller
				total += ct.fileSize - int64(i)*ct.chunkSize
			} else {
				total += ct.chunkSize
			}
		}
	}
	return total
}

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

		var downloaded int64
		var totalSize int64
		for _, f := range files {
			totalSize += f.Size
		}

		// create tasks
		g, gctx := errgroup.WithContext(ctx)
		g.SetLimit(maxConcurrency)

		var trackers []*chunkTracker
		defer func() {
			for _, tr := range trackers {
				tr.close()
			}
		}()

		for _, f := range files {
			if err := os.MkdirAll(filepath.Dir(filepath.Join(outputPath, f.Name)), 0o755); err != nil {
				errCh <- fmt.Errorf("failed to create directory: %v, %s", err, f.Name)
				return
			}

			filePath := filepath.Join(outputPath, f.Name)
			markerPath := filePath + ".chunks"

			// Check if file exists at expected size with NO marker file → skip
			// (backwards-compatible with pre-marker downloads)
			info, statErr := os.Stat(filePath)
			_, markerErr := os.Stat(markerPath)
			if statErr == nil && info.Size() == f.Size && os.IsNotExist(markerErr) {
				slog.Info("File already complete, skipping", "path", filePath, "size", f.Size)
				atomic.AddInt64(&downloaded, f.Size)
				resCh <- types.DownloadInfo{
					TotalDownloaded: atomic.LoadInt64(&downloaded),
					TotalSize:       totalSize,
				}
				continue
			}

			// Check if file is larger than expected → remove file + any stale marker
			if statErr == nil && info.Size() > f.Size {
				slog.Warn("File larger than expected, removing", "path", filePath, "currentSize", info.Size(), "expectedSize", f.Size)
				if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
					errCh <- fmt.Errorf("failed to remove corrupted file: %v, %s", err, f.Name)
					return
				}
				_ = os.Remove(markerPath)
			}

			// Compute chunk parameters and create tracker
			chunkSize := max(minChunkSize, f.Size/128)
			totalChunks := int((f.Size + chunkSize - 1) / chunkSize)
			if totalChunks == 0 {
				totalChunks = 1
			}
			tracker := loadOrCreateTracker(markerPath, f.Size, chunkSize, totalChunks)
			trackers = append(trackers, tracker)

			slog.Info("Download file", "name", f.Name, "size", f.Size, "chunkSize", chunkSize, "totalChunks", totalChunks)

			// If all chunks are already complete, skip
			if tracker.allComplete() {
				slog.Info("All chunks complete, skipping", "path", filePath)
				atomic.AddInt64(&downloaded, f.Size)
				resCh <- types.DownloadInfo{
					TotalDownloaded: atomic.LoadInt64(&downloaded),
					TotalSize:       totalSize,
				}
				_ = tracker.remove()
				trackers = slices.Delete(trackers, len(trackers)-1, len(trackers))
				continue
			}

			// Add completed chunk bytes to progress counter
			completedBytes := tracker.completedBytes()
			if completedBytes > 0 {
				atomic.AddInt64(&downloaded, completedBytes)
				slog.Info("Resuming download", "path", filePath, "completedBytes", completedBytes, "totalBytes", f.Size)
			}

			// Schedule only incomplete chunks
			for chunkID := 0; chunkID < totalChunks; chunkID++ {
				if tracker.isComplete(chunkID) {
					continue
				}

				offset := int64(chunkID) * chunkSize
				limit := min(chunkSize, f.Size-offset)
				cid := chunkID
				tr := tracker

				task := downloadTask{
					OutputPath: outputPath,
					ModelName:  modelName,
					FileName:   f.Name,
					Offset:     offset,
					Limit:      limit,
				}

				g.Go(func() error {
					if err := doTask(gctx, hub, task); err != nil {
						slog.Error("Download task failed", "task", task, "error", err)
						return err
					}

					if err := tr.markComplete(cid); err != nil {
						slog.Warn("Failed to persist chunk marker", "chunkID", cid, "error", err)
					}

					resCh <- types.DownloadInfo{
						TotalDownloaded: atomic.AddInt64(&downloaded, task.Limit),
						TotalSize:       totalSize,
					}
					return nil
				})
			}
		}

		if err := g.Wait(); err != nil {
			errCh <- err
		} else {
			// All downloads succeeded, remove marker files
			for _, tr := range trackers {
				_ = tr.remove()
			}
		}
	}()

	return resCh, errCh
}

var (
	chinaMainlandCheck sync.Once
	isChinaMainland    bool
)

func checkChinaMainland() bool {
	chinaMainlandCheck.Do(func() {
		client := resty.New()
		client.SetTimeout(2 * time.Second)
		defer client.Close()

		for _, ep := range [][]string{
			{"http://ip-api.com/json", "countryCode"},
			{"https://ipapi.co/json", "country_code"},
			{"https://ipinfo.io/json", "country"},
		} {
			res, err := client.R().
				// EnableDebug().
				Get(ep[0])
			if err != nil {
				continue
			}

			n, err := sonic.GetFromString(res.String(), ep[1])
			if err != nil {
				continue
			}

			code, err := n.String()
			if err != nil {
				continue
			}

			slog.Info("Detected country code", "endpoint", ep[0], "code", code)
			isChinaMainland = code == "CN"
			return
		}
		slog.Error("Detect country code failed")
	})
	return isChinaMainland
}

func getHub(ctx context.Context, modelName string) (ModelHub, error) {
	// if only one hub specified, check availability first
	if len(hubs) == 1 {
		h := hubs[0]
		slog.Info("specified single hub", "hub", reflect.TypeOf(h))
		return h, h.CheckAvailable(ctx, modelName)
	}

	// try each hub
	for _, h := range hubs {
		if h.ChinaMainlandOnly() && !checkChinaMainland() {
			slog.Info("skip china mainland only hub", "hub", reflect.TypeOf(h))
			continue
		}
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

func code2error(client *resty.Client, response *resty.Response) error {
	switch response.StatusCode() {
	case http.StatusOK:
		return nil
	case http.StatusNotFound, http.StatusUnauthorized:
		return fmt.Errorf("model not found, please check the model name or auth token")
	default:
		return fmt.Errorf("HTTPError: %s", response.Status())
	}
}
