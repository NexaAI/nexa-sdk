package store

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
	"path"

	"github.com/bytedance/sonic"

	"github.com/NexaAI/nexa-sdk/internal/types"
)

// List returns all locally stored models by reading their manifest files
func (s *Store) List() ([]types.ModelManifest, error) {
	dir := path.Join(s.home, "models")
	names, e := os.ReadDir(dir)
	if e != nil {
		return nil, e
	}

	// Parse each model directory's manifest
	res := make([]types.ModelManifest, 0, len(names))
	for _, encName := range names {
		if !encName.IsDir() {
			continue
		}

		name, err := base64.URLEncoding.DecodeString(encName.Name())
		if err != nil {
			slog.Warn("GetManifest Error", "err", err)
			continue
		}

		model, err := s.GetManifest(string(name))
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
	slog.Info("Remove model", "model", name)

	err := s.LockModel(name)
	if err != nil {
		return err
	}
	defer s.UnlockModel(name)
	return os.RemoveAll(path.Join(s.home, "models", s.encodeName(name)))
}

// Clean removes all stored models and the models directory
func (s *Store) Clean() error {
	slog.Info("Start clean model")

	modelsDir := s.GetModelsDir()
	entries, err := os.ReadDir(modelsDir)
	if err != nil {
		// If models directory doesn't exist, nothing to clean
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// Get list of all model names to remove
	var modelNames []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		modelName, err := base64.URLEncoding.DecodeString(entry.Name())
		if err != nil {
			continue // Skip corrupted directory names
		}

		modelNames = append(modelNames, string(modelName))
	}

	// Remove each model using the Remove function
	// This ensures proper lock handling and consistency
	for _, modelName := range modelNames {
		if err := s.Remove(modelName); err != nil {
			return fmt.Errorf("failed to remove model '%s': %w", modelName, err)
		}
	}

	return nil
}

func (s *Store) GetManifest(name string) (*types.ModelManifest, error) {
	err := s.LockModel(name)
	if err != nil {
		return nil, err
	}
	defer s.UnlockModel(name)

	dir := path.Join(s.home, "models")
	// Read manifest file
	data, e := os.ReadFile(path.Join(dir, s.encodeName(name), "nexa.manifest"))
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

// ModelfilePath returns the full path to a model's data file
func (s *Store) ModelfilePath(name string, file string) string {
	return path.Join(s.home, "models", s.encodeName(name), file)
}

// encodeName encodes model names to safe filesystem names using base64
func (s *Store) encodeName(name string) string {
	return base64.URLEncoding.EncodeToString([]byte(name))
}

// modelDir returns the path to the models directory
func (s *Store) CachefilePath(name string) string {
	return path.Join(s.home, "cache", name)
}

// modelDir returns the path to the models directory
func (s *Store) HistoryFilePath() string {
	return path.Join(s.home, "history")
}
