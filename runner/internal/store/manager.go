package store

import (
	"encoding/base64"
	"log/slog"
	"os"
	"path"
	"sync"

	"github.com/gofrs/flock"
)

// Model directory structure:
// │.
// │└─ models
// │   └─ model_name (base64 encoded)
// │      ├─ modelfile (actual model data)
// │      └─ manifest (model metadata)

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
	cacheDir, e := os.UserCacheDir()
	if e != nil {
		panic(e)
	}

	// Set nexa cache directory
	s.home = path.Join(cacheDir, "nexa")

	// Create models directory structure
	for _, d := range []string{"models", "cache"} {
		e = os.MkdirAll(path.Join(s.home, d), 0o770)
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

func (s *Store) GetModelsDir() string {
	return path.Join(s.home, "models")
}

func (s *Store) cleanCorruptedDirectories() {
	modelsDir := s.GetModelsDir()

	entries, err := os.ReadDir(modelsDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		// try to remove the lock file
		// if entry.Type().IsRegular() && strings.HasSuffix(entry.Name(), ".lock") {
		// 	os.Remove(path.Join(modelsDir, entry.Name()))
		// }

		if !entry.IsDir() {
			continue
		}

		dirName := entry.Name()
		dirPath := path.Join(modelsDir, dirName)
		if s.isCorruptedModelDirectory(dirName, dirPath) {
			modelName, err := base64.URLEncoding.DecodeString(dirName)
			if err != nil {
				slog.Warn("Cleaning invalid name model directory", "dirName", dirName)
				if err := os.RemoveAll(dirPath); err != nil {
					slog.Warn("Failed to remove corrupted directory", "dirname", dirName, "err", err)
				}
				continue
			}

			if err := s.LockModel(string(modelName)); err != nil {
				slog.Warn("Skipping cleanup of directory", "dirName", dirName, "err", err)
				continue
			}

			slog.Info("Cleaning corrupted model directory", "dirname", dirName)
			if err := os.RemoveAll(dirPath); err != nil {
				slog.Error("Failed to remove corrupted directory", "dirName", dirName, "err", err)
			}

			s.UnlockModel(string(modelName))
		}
	}
}

func (s *Store) isCorruptedModelDirectory(dirName, dirPath string) bool {
	if _, err := base64.URLEncoding.DecodeString(dirName); err != nil {
		return true
	}

	manifestPath := path.Join(dirPath, "nexa.manifest")
	if _, err := os.Stat(manifestPath); err != nil {
		return true
	}

	// TDOD: Check Manifest file should be valid JSON and parseable

	return false
}
