package store

import (
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/gofrs/flock"
)

type Store struct {
	home       string
	modelLocks sync.Map // Model locks mapping map[string]*flock.Flock
}

var (
	instance *Store
	once     sync.Once
)

// Get returns the singleton instance of Store
func Get() *Store {
	once.Do(func() {
		instance = &Store{}
		instance.init()
	})
	return instance
}

// init sets up the store's directory structure
func (s *Store) init() {
	// Get user's cache directory (OS-specific)
	// On Windows: C:\Users\<username>\AppData\Local
	// On macOS: ~/Library/Caches
	// On Linux: ~/.cache
	// This properly handles Unicode characters in usernames across all platforms
	cacheDir, e := os.UserCacheDir()
	if e != nil {
		panic(e)
	}

	// Set nexa cache directory
	s.home = filepath.Join(cacheDir, "nexa.ai", "nexa_sdk")

	// Create models directory structure
	for _, d := range []string{"models"} {
		e = os.MkdirAll(filepath.Join(s.home, d), 0o770)
		if e != nil {
			panic(e)
		}
	}

	s.cleanCorruptedDirectories()
}

func (s *Store) Close() error {
	s.modelLocks.Range(func(key, value any) bool {
		fl := value.(*flock.Flock)
		if fl != nil {
			fl.Unlock()
		}
		s.modelLocks.Delete(key)
		return true
	})

	return nil
}

func (s *Store) cleanCorruptedDirectories() {
	models, err := s.scanModelDir()
	if err != nil {
		slog.Error("Failed to scan model directory", "err", err)
		return
	}

	for _, models := range models {
		slog.Info("Checking model directory", "name", models)
		if s.isCorruptedModelDirectory(models) {
			if err := s.LockModel(models); err != nil {
				slog.Warn("Skipping cleanup of directory", "name", models, "err", err)
				continue
			}

			slog.Info("Cleaning corrupted model directory", "name", models)
			if err := os.RemoveAll(s.ModelfilePath(models, "")); err != nil {
				slog.Error("Failed to remove corrupted directory", "name", models, "err", err)
			}

			s.UnlockModel(models)
		}
	}
}

func (s *Store) isCorruptedModelDirectory(name string) bool {
	manifestPath := s.ModelfilePath(name, "nexa.manifest")
	if _, err := os.Stat(manifestPath); err != nil {
		slog.Info("Cleaning corrupted model directory", "name", err)
		return true
	}

	// TDOD: Check Manifest file should be valid JSON and parseable

	return false
}
