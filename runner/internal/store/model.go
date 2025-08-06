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
	orgs, e := os.ReadDir(s.ModelDirPath())
	if e != nil {
		return nil, e
	}

	// Parse each model directory's manifest
	res := make([]types.ModelManifest, 0)
	for _, org := range orgs {
		if !org.IsDir() {
			continue
		}

		repos, e := os.ReadDir(filepath.Join(s.ModelDirPath(), org.Name()))
		if e != nil {
			continue
		}

		for _, repo := range repos {
			if !repo.IsDir() {
				continue
			}

			model, err := s.GetManifest(org.Name() + "/" + repo.Name())
			if err != nil {
				slog.Warn("GetManifest Error", "err", err)
				continue
			}

			res = append(res, *model)
		}
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

	modelsDir := s.ModelDirPath()
	entries, err := os.ReadDir(modelsDir)
	if err != nil {
		// If models directory doesn't exist, nothing to clean
		if !os.IsNotExist(err) {
			slog.Warn("Failed to read model path", "err", err)
		}
		return 0
	}

	// Get list of all model names to remove
	var modelNames []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		modelNames = append(modelNames, entry.Name())
	}

	// Remove each model using the Remove function
	// This ensures proper lock handling and consistency
	count := 0
	for _, modelName := range modelNames {
		if err := s.Remove(modelName); err != nil {
			slog.Warn("Failed to remove model", "modelName", modelName, "err", err)
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
