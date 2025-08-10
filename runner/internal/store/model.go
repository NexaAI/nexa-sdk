package store

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/bytedance/sonic"

	"github.com/NexaAI/nexa-sdk/runner/internal/types"
)

// List returns all locally stored models by reading their manifest files
func (s *Store) List() ([]types.ModelManifest, error) {
	res := make([]types.ModelManifest, 0)
	models, err := s.scanModelDir()
	if err != nil {
		return nil, err
	}
	for _, model := range models {
		// Parse each model directory's manifest
		model, err := s.GetManifest(model)
		if err != nil {
			slog.Warn("GetManifest Error", "err", err)
			continue
		}

		res = append(res, *model)
	}

	return res, nil
}

// Remove deletes a specific model and all its files
func (s *Store) Remove(name string) error {
	slog.Debug("Remove model", "model", name)

	err := s.LockModel(name)
	if err != nil {
		return err
	}
	defer s.UnlockModel(name)
	return os.RemoveAll(filepath.Join(s.home, "models", name))
}

// Clean removes all stored models and the models directory
func (s *Store) Clean() int {
	slog.Debug("Start clean model")

	models, err := s.scanModelDir()
	if err != nil {
		return 0
	}

	// Get list of all model names to remove
	count := 0
	for _, model := range models {
		if err := s.Remove(model); err != nil {
			slog.Warn("Failed to remove model", "model", model, "err", err)
			continue
		}
		count += 1
	}

	return count
}

func (s *Store) GetManifest(name string) (*types.ModelManifest, error) {
	err := s.LockModel(name)
	if err != nil {
		return nil, err
	}
	defer s.UnlockModel(name)

	dir := filepath.Join(s.home, "models")
	// Read manifest file
	data, e := os.ReadFile(filepath.Join(dir, name, "nexa.manifest"))
	if e != nil {
		return nil, e
	}

	// Parse manifest JSON
	model := types.ModelManifest{}
	e = sonic.Unmarshal(data, &model)
	if e != nil {
		return nil, e
	}
	return &model, nil
}

func (s *Store) ModelDirPath() string {
	return filepath.Join(s.home, "models")
}

// ModelfilePath returns the full path to a model's data file
func (s *Store) ModelfilePath(name string, file string) string {
	return filepath.Join(s.home, "models", name, file)
}

func (s *Store) scanModelDir() ([]string, error) {
	orgs, e := os.ReadDir(s.ModelDirPath())
	if e != nil {
		slog.Warn("Failed to read model directory", "err", e)
		return nil, e
	}

	// Parse each model directory's manifest
	res := make([]string, 0)
	for _, org := range orgs {
		if !org.IsDir() {
			continue
		}

		repos, e := os.ReadDir(filepath.Join(s.ModelDirPath(), org.Name()))
		if e != nil {
			slog.Warn("Failed to read model subdirectory", "org", org.Name(), "err", e)
			continue
		}

		for _, repo := range repos {
			if !repo.IsDir() {
				continue
			}

			res = append(res, org.Name()+"/"+repo.Name())
		}
	}

	return res, nil
}
