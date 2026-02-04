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
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
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

func TestChunkTracker(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "chunkTracker_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("fresh tracker has no completed chunks", func(t *testing.T) {
		markerPath := filepath.Join(tmpDir, "fresh.chunks")
		tracker := loadOrCreateTracker(markerPath, 1000, 100, 10)

		for i := 0; i < 10; i++ {
			if tracker.isComplete(i) {
				t.Errorf("Expected chunk %d to be incomplete on fresh tracker", i)
			}
		}
		if tracker.allComplete() {
			t.Error("Expected allComplete to be false on fresh tracker")
		}
	})

	t.Run("markComplete persists to disk and can be reloaded", func(t *testing.T) {
		markerPath := filepath.Join(tmpDir, "persist.chunks")
		tracker := loadOrCreateTracker(markerPath, 1000, 100, 10)

		if err := tracker.markComplete(3); err != nil {
			t.Fatalf("markComplete failed: %v", err)
		}
		if err := tracker.markComplete(7); err != nil {
			t.Fatalf("markComplete failed: %v", err)
		}

		// Reload from disk
		tracker2 := loadOrCreateTracker(markerPath, 1000, 100, 10)
		if !tracker2.isComplete(3) {
			t.Error("Expected chunk 3 to be complete after reload")
		}
		if !tracker2.isComplete(7) {
			t.Error("Expected chunk 7 to be complete after reload")
		}
		if tracker2.isComplete(0) {
			t.Error("Expected chunk 0 to be incomplete after reload")
		}
	})

	t.Run("discards incompatible markers with different file size", func(t *testing.T) {
		markerPath := filepath.Join(tmpDir, "incompat_fsize.chunks")
		tracker := loadOrCreateTracker(markerPath, 1000, 100, 10)
		if err := tracker.markComplete(5); err != nil {
			t.Fatalf("markComplete failed: %v", err)
		}

		// Reload with different file size
		tracker2 := loadOrCreateTracker(markerPath, 2000, 100, 10)
		if tracker2.isComplete(5) {
			t.Error("Expected chunk 5 to be incomplete after file size change")
		}
	})

	t.Run("discards incompatible markers with different chunk size", func(t *testing.T) {
		markerPath := filepath.Join(tmpDir, "incompat_csize.chunks")
		tracker := loadOrCreateTracker(markerPath, 1000, 100, 10)
		if err := tracker.markComplete(5); err != nil {
			t.Fatalf("markComplete failed: %v", err)
		}

		// Reload with different chunk size
		tracker2 := loadOrCreateTracker(markerPath, 1000, 200, 5)
		if tracker2.isComplete(5) {
			t.Error("Expected chunk 5 to be incomplete after chunk size change")
		}
	})

	t.Run("handles corrupted JSON gracefully", func(t *testing.T) {
		markerPath := filepath.Join(tmpDir, "corrupted.chunks")
		if err := os.WriteFile(markerPath, []byte("not valid json{{{"), 0o644); err != nil {
			t.Fatalf("Failed to write corrupted marker: %v", err)
		}

		tracker := loadOrCreateTracker(markerPath, 1000, 100, 10)
		if tracker.allComplete() {
			t.Error("Expected fresh tracker from corrupted JSON")
		}
		for i := 0; i < 10; i++ {
			if tracker.isComplete(i) {
				t.Errorf("Expected chunk %d to be incomplete from corrupted JSON", i)
			}
		}
	})

	t.Run("allComplete returns true when all chunks marked", func(t *testing.T) {
		markerPath := filepath.Join(tmpDir, "allcomplete.chunks")
		totalChunks := 5
		tracker := loadOrCreateTracker(markerPath, 500, 100, totalChunks)

		for i := 0; i < totalChunks; i++ {
			if tracker.allComplete() {
				t.Errorf("Expected allComplete to be false before marking chunk %d", i)
			}
			if err := tracker.markComplete(i); err != nil {
				t.Fatalf("markComplete(%d) failed: %v", i, err)
			}
		}

		if !tracker.allComplete() {
			t.Error("Expected allComplete to be true after marking all chunks")
		}
	})

	t.Run("remove deletes marker file", func(t *testing.T) {
		markerPath := filepath.Join(tmpDir, "removable.chunks")
		tracker := loadOrCreateTracker(markerPath, 1000, 100, 10)
		if err := tracker.markComplete(0); err != nil {
			t.Fatalf("markComplete failed: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(markerPath); err != nil {
			t.Fatalf("Expected marker file to exist: %v", err)
		}

		if err := tracker.remove(); err != nil {
			t.Fatalf("remove failed: %v", err)
		}

		if _, err := os.Stat(markerPath); !os.IsNotExist(err) {
			t.Error("Expected marker file to be deleted")
		}
	})

	t.Run("concurrent markComplete is safe", func(t *testing.T) {
		markerPath := filepath.Join(tmpDir, "concurrent.chunks")
		totalChunks := 50
		tracker := loadOrCreateTracker(markerPath, int64(totalChunks)*100, 100, totalChunks)

		var wg sync.WaitGroup
		wg.Add(totalChunks)
		for i := 0; i < totalChunks; i++ {
			go func(id int) {
				defer wg.Done()
				if err := tracker.markComplete(id); err != nil {
					t.Errorf("concurrent markComplete(%d) failed: %v", id, err)
				}
			}(i)
		}
		wg.Wait()

		if !tracker.allComplete() {
			t.Error("Expected allComplete to be true after concurrent marking")
		}

		// Verify all chunks are marked
		for i := 0; i < totalChunks; i++ {
			if !tracker.isComplete(i) {
				t.Errorf("Expected chunk %d to be complete after concurrent marking", i)
			}
		}
	})

	t.Run("markComplete is idempotent", func(t *testing.T) {
		markerPath := filepath.Join(tmpDir, "idempotent.chunks")
		tracker := loadOrCreateTracker(markerPath, 1000, 100, 10)

		if err := tracker.markComplete(3); err != nil {
			t.Fatalf("first markComplete failed: %v", err)
		}
		if err := tracker.markComplete(3); err != nil {
			t.Fatalf("second markComplete failed: %v", err)
		}

		// Reload and verify only counted once
		tracker2 := loadOrCreateTracker(markerPath, 1000, 100, 10)
		count := 0
		for i := 0; i < 10; i++ {
			if tracker2.isComplete(i) {
				count++
			}
		}
		if count != 1 {
			t.Errorf("Expected 1 completed chunk, got %d", count)
		}
	})

	t.Run("completedBytes calculates correctly", func(t *testing.T) {
		markerPath := filepath.Join(tmpDir, "bytes.chunks")
		// 250 byte file, 100 byte chunks → 3 chunks (100, 100, 50)
		tracker := loadOrCreateTracker(markerPath, 250, 100, 3)

		if err := tracker.markComplete(0); err != nil {
			t.Fatalf("markComplete failed: %v", err)
		}
		if b := tracker.completedBytes(); b != 100 {
			t.Errorf("Expected 100 bytes, got %d", b)
		}

		if err := tracker.markComplete(2); err != nil {
			t.Fatalf("markComplete failed: %v", err)
		}
		// Last chunk is 250 - 2*100 = 50
		if b := tracker.completedBytes(); b != 150 {
			t.Errorf("Expected 150 bytes, got %d", b)
		}

		if err := tracker.markComplete(1); err != nil {
			t.Fatalf("markComplete failed: %v", err)
		}
		if b := tracker.completedBytes(); b != 250 {
			t.Errorf("Expected 250 bytes, got %d", b)
		}
	})
}

// Ensure unused imports are referenced for the test file.
var _ = fmt.Sprintf
