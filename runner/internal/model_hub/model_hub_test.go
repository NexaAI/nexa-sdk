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
	"encoding/binary"
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
		defer tracker.close()

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
		defer tracker.close()

		if err := tracker.markComplete(3); err != nil {
			t.Fatalf("markComplete failed: %v", err)
		}
		if err := tracker.markComplete(7); err != nil {
			t.Fatalf("markComplete failed: %v", err)
		}

		// Reload from disk
		tracker2 := loadOrCreateTracker(markerPath, 1000, 100, 10)
		defer tracker2.close()
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
		defer tracker.close()
		if err := tracker.markComplete(5); err != nil {
			t.Fatalf("markComplete failed: %v", err)
		}

		// Reload with different file size
		tracker2 := loadOrCreateTracker(markerPath, 2000, 100, 10)
		defer tracker2.close()
		if tracker2.isComplete(5) {
			t.Error("Expected chunk 5 to be incomplete after file size change")
		}
	})

	t.Run("discards incompatible markers with different chunk size", func(t *testing.T) {
		markerPath := filepath.Join(tmpDir, "incompat_csize.chunks")
		tracker := loadOrCreateTracker(markerPath, 1000, 100, 10)
		defer tracker.close()
		if err := tracker.markComplete(5); err != nil {
			t.Fatalf("markComplete failed: %v", err)
		}

		// Reload with different chunk size
		tracker2 := loadOrCreateTracker(markerPath, 1000, 200, 5)
		defer tracker2.close()
		if tracker2.isComplete(5) {
			t.Error("Expected chunk 5 to be incomplete after chunk size change")
		}
	})

	t.Run("handles corrupted marker file gracefully", func(t *testing.T) {
		markerPath := filepath.Join(tmpDir, "corrupted.chunks")
		if err := os.WriteFile(markerPath, []byte("garbage data that is not a valid binary marker"), 0o644); err != nil {
			t.Fatalf("Failed to write corrupted marker: %v", err)
		}

		tracker := loadOrCreateTracker(markerPath, 1000, 100, 10)
		defer tracker.close()
		if tracker.allComplete() {
			t.Error("Expected fresh tracker from corrupted marker")
		}
		for i := 0; i < 10; i++ {
			if tracker.isComplete(i) {
				t.Errorf("Expected chunk %d to be incomplete from corrupted marker", i)
			}
		}
	})

	t.Run("allComplete returns true when all chunks marked", func(t *testing.T) {
		markerPath := filepath.Join(tmpDir, "allcomplete.chunks")
		totalChunks := 5
		tracker := loadOrCreateTracker(markerPath, 500, 100, totalChunks)
		defer tracker.close()

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

	t.Run("concurrent markComplete is safe and persisted", func(t *testing.T) {
		markerPath := filepath.Join(tmpDir, "concurrent.chunks")
		totalChunks := 50
		tracker := loadOrCreateTracker(markerPath, int64(totalChunks)*100, 100, totalChunks)
		defer tracker.close()

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

		// Verify all chunks are marked in memory
		for i := 0; i < totalChunks; i++ {
			if !tracker.isComplete(i) {
				t.Errorf("Expected chunk %d to be complete after concurrent marking", i)
			}
		}

		// Reload from disk to verify WriteAt persistence
		tracker2 := loadOrCreateTracker(markerPath, int64(totalChunks)*100, 100, totalChunks)
		defer tracker2.close()
		for i := 0; i < totalChunks; i++ {
			if !tracker2.isComplete(i) {
				t.Errorf("Expected chunk %d to be persisted on disk after concurrent marking", i)
			}
		}
	})

	t.Run("markComplete is idempotent", func(t *testing.T) {
		markerPath := filepath.Join(tmpDir, "idempotent.chunks")
		tracker := loadOrCreateTracker(markerPath, 1000, 100, 10)
		defer tracker.close()

		if err := tracker.markComplete(3); err != nil {
			t.Fatalf("first markComplete failed: %v", err)
		}
		if err := tracker.markComplete(3); err != nil {
			t.Fatalf("second markComplete failed: %v", err)
		}

		// Reload and verify only counted once
		tracker2 := loadOrCreateTracker(markerPath, 1000, 100, 10)
		defer tracker2.close()
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

	t.Run("markComplete rejects out-of-range chunkID", func(t *testing.T) {
		markerPath := filepath.Join(tmpDir, "bounds.chunks")
		tracker := loadOrCreateTracker(markerPath, 1000, 100, 10)
		defer tracker.close()

		if err := tracker.markComplete(-1); err == nil {
			t.Error("Expected error for negative chunkID")
		}
		if err := tracker.markComplete(10); err == nil {
			t.Error("Expected error for chunkID == totalChunks")
		}
		if err := tracker.markComplete(100); err == nil {
			t.Error("Expected error for chunkID >> totalChunks")
		}
	})

	t.Run("completedBytes calculates correctly", func(t *testing.T) {
		markerPath := filepath.Join(tmpDir, "bytes.chunks")
		// 250 byte file, 100 byte chunks → 3 chunks (100, 100, 50)
		tracker := loadOrCreateTracker(markerPath, 250, 100, 3)
		defer tracker.close()

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

	t.Run("binary marker file has correct on-disk layout", func(t *testing.T) {
		markerPath := filepath.Join(tmpDir, "layout.chunks")
		var fileSize int64 = 5000
		var chunkSize int64 = 500
		totalChunks := 10

		tracker := loadOrCreateTracker(markerPath, fileSize, chunkSize, totalChunks)
		defer tracker.close()
		if err := tracker.markComplete(2); err != nil {
			t.Fatalf("markComplete failed: %v", err)
		}
		if err := tracker.markComplete(7); err != nil {
			t.Fatalf("markComplete failed: %v", err)
		}

		// Read raw file and validate structure
		data, err := os.ReadFile(markerPath)
		if err != nil {
			t.Fatalf("Failed to read marker file: %v", err)
		}

		expectedLen := markerHeaderLen + totalChunks
		if len(data) != expectedLen {
			t.Fatalf("Expected marker file size %d, got %d", expectedLen, len(data))
		}

		// Validate header
		if string(data[:4]) != markerMagic {
			t.Errorf("Expected magic %q, got %q", markerMagic, string(data[:4]))
		}
		if got := int64(binary.LittleEndian.Uint64(data[4:12])); got != fileSize {
			t.Errorf("Expected fileSize %d, got %d", fileSize, got)
		}
		if got := int64(binary.LittleEndian.Uint64(data[12:20])); got != chunkSize {
			t.Errorf("Expected chunkSize %d, got %d", chunkSize, got)
		}
		if got := int(binary.LittleEndian.Uint32(data[20:24])); got != totalChunks {
			t.Errorf("Expected totalChunks %d, got %d", totalChunks, got)
		}

		// Validate chunk bytes
		chunkData := data[markerHeaderLen:]
		for i, b := range chunkData {
			if i == 2 || i == 7 {
				if b != 0x01 {
					t.Errorf("Expected chunk %d to be 0x01, got 0x%02x", i, b)
				}
			} else {
				if b != 0x00 {
					t.Errorf("Expected chunk %d to be 0x00, got 0x%02x", i, b)
				}
			}
		}
	})

	t.Run("gracefully migrates from old JSON marker format", func(t *testing.T) {
		markerPath := filepath.Join(tmpDir, "old_json.chunks")
		// Simulate an old JSON-format marker file from a previous version
		oldJSON := `{"file_size":1000,"chunk_size":100,"total_chunks":10,"completed_chunks":[3,7]}`
		if err := os.WriteFile(markerPath, []byte(oldJSON), 0o644); err != nil {
			t.Fatalf("Failed to write old JSON marker: %v", err)
		}

		// Load should detect bad magic and start fresh
		tracker := loadOrCreateTracker(markerPath, 1000, 100, 10)
		defer tracker.close()
		for i := 0; i < 10; i++ {
			if tracker.isComplete(i) {
				t.Errorf("Expected chunk %d to be incomplete after migration from JSON", i)
			}
		}

		// Verify the file was overwritten with valid binary format
		data, err := os.ReadFile(markerPath)
		if err != nil {
			t.Fatalf("Failed to read migrated marker: %v", err)
		}
		if string(data[:4]) != markerMagic {
			t.Errorf("Expected migrated file to have binary magic, got %q", string(data[:4]))
		}
	})

	t.Run("close is safe to call multiple times", func(t *testing.T) {
		markerPath := filepath.Join(tmpDir, "multiclose.chunks")
		tracker := loadOrCreateTracker(markerPath, 1000, 100, 10)

		tracker.close()
		tracker.close() // should not panic

		// After close, markComplete should still work in memory (file == nil path)
		if err := tracker.markComplete(0); err != nil {
			t.Fatalf("markComplete after close failed: %v", err)
		}
		if !tracker.isComplete(0) {
			t.Error("Expected chunk 0 complete in memory after close")
		}
	})
}
