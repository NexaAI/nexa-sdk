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
	"sync"
	"sync/atomic"
	"time"

	"github.com/bytedance/sonic"
	"golang.org/x/sync/errgroup"
	"resty.dev/v3"

	"github.com/NexaAI/nexa-sdk/runner/internal/types"
)

const ProgressSuffix = ".progress"

type chunkProgress struct {
	bitmap    []uint64
	fileSize  int64
	chunkSize int64
	path      string
	dirty     atomic.Int32
}

func (p *chunkProgress) nChunks() int {
	return int((p.fileSize + p.chunkSize - 1) / p.chunkSize)
}

func loadProgress(path string, fileSize, chunkSize int64) (*chunkProgress, error) {
	if chunkSize <= 0 || fileSize <= 0 {
		return nil, nil
	}
	nChunks := int((fileSize + chunkSize - 1) / chunkSize)
	if nChunks == 0 {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil
	}
	const headerSize = 16
	if len(data) < headerSize {
		return nil, nil
	}
	gotSize := binary.LittleEndian.Uint64(data[0:8])
	gotChunk := binary.LittleEndian.Uint64(data[8:16])
	if uint64(fileSize) != gotSize || uint64(chunkSize) != gotChunk {
		return nil, nil
	}
	words := (nChunks + 63) / 64
	if len(data) < headerSize+words*8 {
		return nil, nil
	}
	p := &chunkProgress{
		bitmap:    make([]uint64, words),
		fileSize:  fileSize,
		chunkSize: chunkSize,
		path:      path,
	}
	for i := range words {
		p.bitmap[i] = binary.LittleEndian.Uint64(data[headerSize+i*8:])
	}
	return p, nil
}

func (p *chunkProgress) isDone(i int) bool {
	if i < 0 || i >= p.nChunks() {
		return false
	}
	w := atomic.LoadUint64(&p.bitmap[i/64])
	return (w & (1 << (i % 64))) != 0
}

func (p *chunkProgress) setDone(i int) {
	if i < 0 || i >= p.nChunks() {
		return
	}
	atomic.OrUint64(&p.bitmap[i/64], 1<<(i%64))
	p.dirty.Store(1)
}

func (p *chunkProgress) save() error {
	if !p.dirty.CompareAndSwap(1, 0) {
		return nil
	}
	words := len(p.bitmap)
	buf := make([]byte, 16+words*8)
	binary.LittleEndian.PutUint64(buf[0:8], uint64(p.fileSize))
	binary.LittleEndian.PutUint64(buf[8:16], uint64(p.chunkSize))
	for i := range words {
		binary.LittleEndian.PutUint64(buf[16+i*8:], atomic.LoadUint64(&p.bitmap[i]))
	}
	tmpPath := p.path + ".tmp"
	if err := os.WriteFile(tmpPath, buf, 0o644); err != nil {
		return err
	}
	return os.Rename(tmpPath, p.path)
}

func (p *chunkProgress) anyDone() bool {
	for i := 0; i < len(p.bitmap); i++ {
		if atomic.LoadUint64(&p.bitmap[i]) != 0 {
			return true
		}
	}
	return false
}

func (p *chunkProgress) countDoneAndSize(chunkSizeFunc func(i int) int64) (n int, size int64) {
	for i := 0; i < p.nChunks(); i++ {
		if p.isDone(i) {
			n++
			size += chunkSizeFunc(i)
		}
	}
	return n, size
}

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

		type fileWork struct {
			f         ModelFileInfo
			chunkSize int64
			nChunks   int
			progress  *chunkProgress
		}
		var workList []fileWork
		var downloaded int64
		var progressList []*chunkProgress

		for _, f := range files {
			chunkSize := max(minChunkSize, f.Size/128)
			nChunks := 0
			if chunkSize > 0 {
				nChunks = int((f.Size + chunkSize - 1) / chunkSize)
			}
			progressPath := filepath.Join(outputPath, f.Name+ProgressSuffix)
			var progress *chunkProgress
			if nChunks > 0 {
				if p, _ := loadProgress(progressPath, f.Size, chunkSize); p != nil {
					progress = p
				} else {
					progress = &chunkProgress{
						bitmap:    make([]uint64, (nChunks+63)/64),
						fileSize:  f.Size,
						chunkSize: chunkSize,
						path:      progressPath,
					}
				}
				nDone, size := progress.countDoneAndSize(func(i int) int64 {
					offset := int64(i) * chunkSize
					return min(chunkSize, f.Size-offset)
				})
				downloaded += size
				if nDone > 0 {
					slog.Info("resuming download", "file", f.Name, "chunks_done", nDone, "bytes_done", size, "total_chunks", nChunks)
				}
				progressList = append(progressList, progress)
			}
			workList = append(workList, fileWork{f: f, chunkSize: chunkSize, nChunks: nChunks, progress: progress})
		}

		done := make(chan struct{})
		defer close(done)
		go func() {
			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					for _, p := range progressList {
						_ = p.save()
					}
				}
			}
		}()

		g, gctx := errgroup.WithContext(ctx)
		g.SetLimit(maxConcurrency)

		for _, w := range workList {
			f, chunkSize, nChunks, prog := w.f, w.chunkSize, w.nChunks, w.progress
			if err := os.MkdirAll(filepath.Dir(filepath.Join(outputPath, f.Name)), 0o755); err != nil {
				errCh <- fmt.Errorf("failed to create directory: %v, %s", err, f.Name)
				return
			}
			if nChunks == 0 {
				continue
			}
			if prog != nil && prog.anyDone() {
				outPath := filepath.Join(outputPath, f.Name)
				slog.Debug("opening output file for resume", "file", f.Name)
				file, err := os.OpenFile(outPath, os.O_RDWR|os.O_CREATE, 0o644)
				if err != nil {
					errCh <- err
					return
				}
				fi, _ := file.Stat()
				if fi == nil || fi.Size() < f.Size {
					if err := file.Truncate(f.Size); err != nil {
						file.Close()
						errCh <- err
						return
					}
				}
				file.Close()
			}
			slog.Info("Download file", "name", f.Name, "size", f.Size, "chunkSize", chunkSize)

			for i := range nChunks {
				if prog != nil && prog.isDone(i) {
					continue
				}
				offset := int64(i) * chunkSize
				limit := min(chunkSize, f.Size-offset)
				task := downloadTask{
					OutputPath: outputPath,
					ModelName:  modelName,
					FileName:   f.Name,
					Offset:     offset,
					Limit:      limit,
				}
				idx, t := i, task
				g.Go(func() error {
					if err := doTask(gctx, hub, t); err != nil {
						slog.Error("Download task failed", "task", t, "error", err)
						return err
					}
					if prog != nil {
						prog.setDone(idx)
					}
					resCh <- types.DownloadInfo{
						TotalDownloaded: atomic.AddInt64(&downloaded, t.Limit),
						TotalSize:       totalSize,
					}
					return nil
				})
			}
		}

		if err := g.Wait(); err != nil {
			errCh <- err
			return
		}
		for _, w := range workList {
			if w.progress != nil {
				if err := os.Remove(w.progress.path); err != nil {
					slog.Debug("remove progress file", "path", w.progress.path, "error", err)
				}
			}
		}
		slog.Info("download complete", "model", modelName, "outputPath", outputPath)
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
