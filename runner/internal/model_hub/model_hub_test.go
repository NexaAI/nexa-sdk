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

func TestCheckExistingFile(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "checkExistingFile_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("file doesn't exist returns 0", func(t *testing.T) {
		result := checkExistingFile(tmpDir+"/nonexistent.txt", 1000)
		if result != 0 {
			t.Errorf("Expected 0 for nonexistent file, got %d", result)
		}
	})

	t.Run("file complete returns size", func(t *testing.T) {
		filePath := tmpDir + "/complete.txt"
		expectedSize := int64(100)
		content := make([]byte, expectedSize)
		if err := os.WriteFile(filePath, content, 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		result := checkExistingFile(filePath, expectedSize)
		if result != expectedSize {
			t.Errorf("Expected %d for complete file, got %d", expectedSize, result)
		}
	})

	t.Run("file partial returns partial size", func(t *testing.T) {
		filePath := tmpDir + "/partial.txt"
		partialSize := int64(50)
		expectedSize := int64(100)
		content := make([]byte, partialSize)
		if err := os.WriteFile(filePath, content, 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		result := checkExistingFile(filePath, expectedSize)
		if result != partialSize {
			t.Errorf("Expected %d for partial file, got %d", partialSize, result)
		}
	})

	t.Run("file too large returns -1", func(t *testing.T) {
		filePath := tmpDir + "/toolarge.txt"
		actualSize := int64(150)
		expectedSize := int64(100)
		content := make([]byte, actualSize)
		if err := os.WriteFile(filePath, content, 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		result := checkExistingFile(filePath, expectedSize)
		if result != -1 {
			t.Errorf("Expected -1 for file larger than expected, got %d", result)
		}
	})

	t.Run("empty file returns 0", func(t *testing.T) {
		filePath := tmpDir + "/empty.txt"
		if err := os.WriteFile(filePath, []byte{}, 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		result := checkExistingFile(filePath, 100)
		if result != 0 {
			t.Errorf("Expected 0 for empty file, got %d", result)
		}
	})
}
