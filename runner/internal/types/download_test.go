package types

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadProgress_NewAndSave(t *testing.T) {
	tmpDir := t.TempDir()
	
	progress := NewDownloadProgress("test-model", 1000000)
	progress.AddFile("model.gguf", 1000000)
	
	if err := progress.Save(tmpDir); err != nil {
		t.Fatalf("Failed to save progress: %v", err)
	}
	
	// Verify file exists
	progressFile := filepath.Join(tmpDir, ".download_progress.json")
	if _, err := os.Stat(progressFile); os.IsNotExist(err) {
		t.Fatal("Progress file was not created")
	}
}

func TestDownloadProgress_LoadAndResume(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create and save progress
	progress := NewDownloadProgress("test-model", 1000000)
	progress.AddFile("model.gguf", 1000000)
	progress.Files["model.gguf"].MarkRangeComplete(0, 500000)
	
	if err := progress.Save(tmpDir); err != nil {
		t.Fatalf("Failed to save progress: %v", err)
	}
	
	// Load progress
	loaded, err := LoadDownloadProgress(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load progress: %v", err)
	}
	
	if loaded == nil {
		t.Fatal("Loaded progress is nil")
	}
	
	if loaded.ModelName != "test-model" {
		t.Errorf("Expected model name 'test-model', got '%s'", loaded.ModelName)
	}
	
	if loaded.Files["model.gguf"].Downloaded != 500000 {
		t.Errorf("Expected 500000 bytes downloaded, got %d", loaded.Files["model.gguf"].Downloaded)
	}
}

func TestFileProgress_MarkRangeComplete(t *testing.T) {
	fp := &FileProgress{
		FileName:       "test.bin",
		TotalSize:      1000,
		CompletedRange: make([]CompletedRange, 0),
	}
	
	// Mark first range
	fp.MarkRangeComplete(0, 100)
	if fp.Downloaded != 100 {
		t.Errorf("Expected 100 bytes downloaded, got %d", fp.Downloaded)
	}
	
	// Mark second range
	fp.MarkRangeComplete(200, 300)
	if fp.Downloaded != 200 {
		t.Errorf("Expected 200 bytes downloaded, got %d", fp.Downloaded)
	}
	
	// Mark overlapping range (should merge)
	fp.MarkRangeComplete(100, 200)
	if fp.Downloaded != 300 {
		t.Errorf("Expected 300 bytes downloaded after merge, got %d", fp.Downloaded)
	}
	
	// Should have merged into single range
	if len(fp.CompletedRange) != 1 {
		t.Errorf("Expected 1 range after merge, got %d", len(fp.CompletedRange))
	}
	
	if fp.CompletedRange[0].Start != 0 || fp.CompletedRange[0].End != 300 {
		t.Errorf("Expected range 0-300, got %d-%d", fp.CompletedRange[0].Start, fp.CompletedRange[0].End)
	}
}

func TestFileProgress_GetMissingRanges(t *testing.T) {
	fp := &FileProgress{
		FileName:       "test.bin",
		TotalSize:      1000,
		CompletedRange: make([]CompletedRange, 0),
	}
	
	// No downloads yet
	missing := fp.GetMissingRanges(100)
	if len(missing) != 1 {
		t.Fatalf("Expected 1 missing range, got %d", len(missing))
	}
	if missing[0].Start != 0 || missing[0].End != 1000 {
		t.Errorf("Expected missing range 0-1000, got %d-%d", missing[0].Start, missing[0].End)
	}
	
	// Download middle section
	fp.MarkRangeComplete(400, 600)
	missing = fp.GetMissingRanges(100)
	
	if len(missing) != 2 {
		t.Fatalf("Expected 2 missing ranges, got %d", len(missing))
	}
	
	// First gap: 0-400
	if missing[0].Start != 0 || missing[0].End != 400 {
		t.Errorf("Expected first gap 0-400, got %d-%d", missing[0].Start, missing[0].End)
	}
	
	// Second gap: 600-1000
	if missing[1].Start != 600 || missing[1].End != 1000 {
		t.Errorf("Expected second gap 600-1000, got %d-%d", missing[1].Start, missing[1].End)
	}
}

func TestFileProgress_IsComplete(t *testing.T) {
	fp := &FileProgress{
		FileName:       "test.bin",
		TotalSize:      1000,
		CompletedRange: make([]CompletedRange, 0),
	}
	
	if fp.IsComplete() {
		t.Error("File should not be complete initially")
	}
	
	// Download entire file
	fp.MarkRangeComplete(0, 1000)
	
	if !fp.IsComplete() {
		t.Error("File should be complete after downloading entire range")
	}
	
	if fp.Downloaded != 1000 {
		t.Errorf("Expected 1000 bytes downloaded, got %d", fp.Downloaded)
	}
}

func TestFileProgress_MergeAdjacentRanges(t *testing.T) {
	fp := &FileProgress{
		FileName:       "test.bin",
		TotalSize:      1000,
		CompletedRange: make([]CompletedRange, 0),
	}
	
	// Download in reverse order
	fp.MarkRangeComplete(600, 800)
	fp.MarkRangeComplete(400, 600)
	fp.MarkRangeComplete(200, 400)
	fp.MarkRangeComplete(0, 200)
	
	// All ranges should be merged into one
	if len(fp.CompletedRange) != 1 {
		t.Errorf("Expected 1 merged range, got %d", len(fp.CompletedRange))
	}
	
	if fp.CompletedRange[0].Start != 0 || fp.CompletedRange[0].End != 800 {
		t.Errorf("Expected merged range 0-800, got %d-%d", fp.CompletedRange[0].Start, fp.CompletedRange[0].End)
	}
	
	if fp.Downloaded != 800 {
		t.Errorf("Expected 800 bytes downloaded, got %d", fp.Downloaded)
	}
}

func TestCleanupProgress(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create progress file
	progress := NewDownloadProgress("test-model", 1000000)
	if err := progress.Save(tmpDir); err != nil {
		t.Fatalf("Failed to save progress: %v", err)
	}
	
	// Verify it exists
	progressFile := filepath.Join(tmpDir, ".download_progress.json")
	if _, err := os.Stat(progressFile); os.IsNotExist(err) {
		t.Fatal("Progress file should exist before cleanup")
	}
	
	// Cleanup
	if err := CleanupProgress(tmpDir); err != nil {
		t.Fatalf("Failed to cleanup progress: %v", err)
	}
	
	// Verify it's gone
	if _, err := os.Stat(progressFile); !os.IsNotExist(err) {
		t.Error("Progress file should be deleted after cleanup")
	}
}

func TestDownloadProgress_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()
	
	progress := NewDownloadProgress("test-model", 3000000)
	progress.AddFile("model1.gguf", 1000000)
	progress.AddFile("model2.gguf", 1000000)
	progress.AddFile("model3.gguf", 1000000)
	
	// Download parts of each file
	progress.Files["model1.gguf"].MarkRangeComplete(0, 500000)
	progress.Files["model2.gguf"].MarkRangeComplete(0, 750000)
	progress.Files["model3.gguf"].MarkRangeComplete(0, 1000000) // Complete
	
	// Save and reload
	if err := progress.Save(tmpDir); err != nil {
		t.Fatalf("Failed to save progress: %v", err)
	}
	
	loaded, err := LoadDownloadProgress(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load progress: %v", err)
	}
	
	// Verify all files
	if loaded.Files["model1.gguf"].Downloaded != 500000 {
		t.Errorf("model1 expected 500000, got %d", loaded.Files["model1.gguf"].Downloaded)
	}
	
	if loaded.Files["model2.gguf"].Downloaded != 750000 {
		t.Errorf("model2 expected 750000, got %d", loaded.Files["model2.gguf"].Downloaded)
	}
	
	if !loaded.Files["model3.gguf"].IsComplete() {
		t.Error("model3 should be complete")
	}
}

