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
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/lmittmann/tint"
)

const MODEL_NAME = "NexaAI/OmniNeural-4B"

func TestMain(m *testing.M) {
	slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		AddSource: true,
		Level:     slog.LevelDebug,
	})))

	// only test huggingface
	hubs = hubs[3:]

	os.Exit(m.Run())
}

func TestChunkProgress_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "x.progress")
	fileSize := int64(320)
	chunkSize := int64(100)
	nChunks := 4
	words := (nChunks + 63) / 64
	progress := &chunkProgress{
		bitmap:    make([]uint64, words),
		fileSize:  fileSize,
		chunkSize: chunkSize,
		path:      path,
	}
	progress.setDone(0)
	progress.setDone(2)
	if err := progress.save(); err != nil {
		t.Fatal(err)
	}
	loaded, err := loadProgress(path, fileSize, chunkSize)
	if err != nil || loaded == nil {
		t.Fatalf("loadProgress after save: %v, %v", loaded, err)
	}
	if !loaded.isDone(0) || loaded.isDone(1) || !loaded.isDone(2) || loaded.isDone(3) {
		t.Errorf("loaded progress: isDone(0)=true, isDone(1)=false, isDone(2)=true, isDone(3)=false")
	}
	chunkSizeFunc := func(i int) int64 {
		offset := int64(i) * chunkSize
		if offset+chunkSize > fileSize {
			return fileSize - offset
		}
		return chunkSize
	}
	n, size := loaded.countDoneAndSize(chunkSizeFunc)
	if n != 2 || size != 200 {
		t.Errorf("countDoneAndSize: want n=2 size=200, got n=%d size=%d", n, size)
	}
}

func TestModelInfo(t *testing.T) {
	data, _, err := ModelInfo(context.Background(), MODEL_NAME)
	if err != nil {
		t.Error(err)
	}
	t.Log(data)
}

func TestGetFileContent(t *testing.T) {
	data, err := GetFileContent(context.Background(), MODEL_NAME, ".gitattributes")
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("GetFileContent:\n%s", data)
}

func TestDownload(t *testing.T) {
	files, _, err := ModelInfo(context.Background(), MODEL_NAME)
	if err != nil {
		t.Error(err)
		return
	}

	resCh, errCh := StartDownload(context.Background(), MODEL_NAME, "/tmp/OmniNeural-4B", files)
	for p := range resCh {
		t.Logf("Downloaded: %d / %d", p.TotalDownloaded, p.TotalSize)
	}
	for e := range errCh {
		t.Error(e)
	}

	os.RemoveAll("/tmp/OmniNeural-4B/")
}

func BenchmarkDownload(b *testing.B) {
	files, _, err := ModelInfo(context.Background(), "ggml-org/embeddinggemma-300M-qat-q4_0-GGUF")
	if err != nil {
		b.Error(err)
		return
	}

	resCh, errCh := StartDownload(context.Background(), "ggml-org/embeddinggemma-300M-qat-q4_0-GGUF", "/tmp/embeddinggemma-300M-qat-q4_0-GGUF", files)
	for p := range resCh {
		b.Logf("Downloaded: %d / %d", p.TotalDownloaded, p.TotalSize)
	}
	for e := range errCh {
		b.Error(e)
	}
}
