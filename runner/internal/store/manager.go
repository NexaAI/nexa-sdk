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
	homeDir, e := os.UserHomeDir()
	if e != nil {
		panic(e)
	}

	// Set nexa cache directory
	s.home = filepath.Join(homeDir, ".cache", "nexa.ai", "nexa_sdk")

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
	modelsDir := s.ModelDirPath()

	entries, err := os.ReadDir(modelsDir)
	if err != nil {
		return
	}

	// org name
	for _, entry := range entries {
		// try to remove the lock file
		//if entry.Type().IsRegular() && strings.HasSuffix(entry.Name(), ".lock") {
		//	os.Remove(filepath.Join(modelsDir, entry.Name()))
		//}

		if !entry.IsDir() {
			continue
		}

		// repo name
		subentries, err := os.ReadDir(filepath.Join(modelsDir, entry.Name()))
		if err != nil {
			continue
		}
		for _, subentry := range subentries {
			if !subentry.IsDir() {
				continue
			}
			name := entry.Name() + "/" + subentry.Name()
			slog.Info("Checking model directory", "name", name)
			if s.isCorruptedModelDirectory(name) {
				if err := s.LockModel(name); err != nil {
					slog.Warn("Skipping cleanup of directory", "name", name, "err", err)
					continue
				}

				slog.Info("Cleaning corrupted model directory", "name", name)
				if err := os.RemoveAll(s.ModelfilePath(name, "")); err != nil {
					slog.Error("Failed to remove corrupted directory", "name", name, "err", err)
				}

				s.UnlockModel(name)
			}
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
