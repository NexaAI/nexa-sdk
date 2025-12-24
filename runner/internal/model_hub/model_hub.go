// Copyright 2024-2025 Nexa AI, Inc.
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
				slog.Info("Download file", "name", f.Name, "size", f.Size, "chunkSize", chunkSize)
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
