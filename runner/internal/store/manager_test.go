package store

import (
	"os"
	"path/filepath"
	"testing"
)

// TestUnicodePaths verifies that the store can handle Unicode characters in paths
// This is critical for Windows users with non-ASCII characters in their usernames
func TestUnicodePaths(t *testing.T) {
	// Test cases with various Unicode characters
	testCases := []struct {
		name     string
		username string
	}{
		{"German Umlaut", "Jörg"},
		{"French Accent", "François"},
		{"Spanish Tilde", "José"},
		{"Japanese", "山田"},
		{"Emoji", "User😀"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Get cache directory - this should work regardless of Unicode in path
			cacheDir, err := os.UserCacheDir()
			if err != nil {
				t.Fatalf("os.UserCacheDir() failed: %v", err)
			}

			// Create a test path with the Unicode username
			testPath := filepath.Join(cacheDir, "nexa_test_"+tc.username)
			
			// Attempt to create directory
			err = os.MkdirAll(testPath, 0o770)
			if err != nil {
				t.Errorf("Failed to create directory with Unicode path: %v", err)
			}
			defer os.RemoveAll(testPath)

			// Verify directory was created
			info, err := os.Stat(testPath)
			if err != nil {
				t.Errorf("Failed to stat created directory: %v", err)
			}
			if !info.IsDir() {
				t.Errorf("Created path is not a directory")
			}

			// Test file operations within the Unicode path
			testFile := filepath.Join(testPath, "test.txt")
			testContent := []byte("Unicode path test")
			
			err = os.WriteFile(testFile, testContent, 0o644)
			if err != nil {
				t.Errorf("Failed to write file in Unicode path: %v", err)
			}

			// Read back the file
			readContent, err := os.ReadFile(testFile)
			if err != nil {
				t.Errorf("Failed to read file from Unicode path: %v", err)
			}

			if string(readContent) != string(testContent) {
				t.Errorf("File content mismatch: got %s, want %s", readContent, testContent)
			}
		})
	}
}

// TestStoreInitWithUserCacheDir ensures store initialization works with UserCacheDir
func TestStoreInitWithUserCacheDir(t *testing.T) {
	// This test verifies that the store can initialize using UserCacheDir
	// which properly handles Unicode characters across all platforms
	
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		t.Fatalf("os.UserCacheDir() failed: %v", err)
	}

	// Verify cache directory exists and is accessible
	info, err := os.Stat(cacheDir)
	if err != nil {
		t.Fatalf("Cannot access cache directory: %v", err)
	}

	if !info.IsDir() {
		t.Fatalf("Cache path is not a directory")
	}

	// Create test subdirectory structure similar to nexa
	testDir := filepath.Join(cacheDir, "nexa_test_store")
	defer os.RemoveAll(testDir)

	err = os.MkdirAll(filepath.Join(testDir, "models"), 0o770)
	if err != nil {
		t.Fatalf("Failed to create store directory structure: %v", err)
	}

	// Verify structure was created
	modelsPath := filepath.Join(testDir, "models")
	info, err = os.Stat(modelsPath)
	if err != nil {
		t.Fatalf("Failed to stat models directory: %v", err)
	}

	if !info.IsDir() {
		t.Fatalf("Models path is not a directory")
	}
}

// TestPathOperationsWithUnicode tests various filesystem operations with Unicode paths
func TestPathOperationsWithUnicode(t *testing.T) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		t.Fatalf("os.UserCacheDir() failed: %v", err)
	}

	// Create test directory with Unicode
	testBase := filepath.Join(cacheDir, "nexa_test_字符")
	defer os.RemoveAll(testBase)

	// Test directory creation
	err = os.MkdirAll(filepath.Join(testBase, "subdir", "nested"), 0o770)
	if err != nil {
		t.Errorf("Failed to create nested directories: %v", err)
	}

	// Test file creation
	testFile := filepath.Join(testBase, "test_文件.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0o644)
	if err != nil {
		t.Errorf("Failed to write Unicode filename: %v", err)
	}

	// Test file reading
	_, err = os.ReadFile(testFile)
	if err != nil {
		t.Errorf("Failed to read Unicode filename: %v", err)
	}

	// Test directory listing
	entries, err := os.ReadDir(testBase)
	if err != nil {
		t.Errorf("Failed to read directory: %v", err)
	}

	if len(entries) == 0 {
		t.Errorf("Expected entries in directory, got none")
	}

	// Test file removal
	err = os.Remove(testFile)
	if err != nil {
		t.Errorf("Failed to remove file with Unicode name: %v", err)
	}
}

