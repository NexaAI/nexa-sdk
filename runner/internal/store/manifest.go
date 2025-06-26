package store

import (
	"fmt"
	"os"
	"path"

	"github.com/bytedance/sonic"

	"github.com/NexaAI/nexa-sdk/internal/types"
)

// List returns all locally stored models by reading their manifest files
func (s *Store) List() ([]types.Model, error) {
	dir := path.Join(s.home, "models")
	names, e := os.ReadDir(dir)
	if e != nil {
		return nil, e
	}

	// Parse each model directory's manifest
	res := make([]types.Model, 0, len(names))
	for _, name := range names {
		model, err := s.getManifest(name.Name())
		if err != nil {
			fmt.Printf("getManifest error: %s\n", err)
			continue
		}

		res = append(res, *model)
	}

	return res, nil
}

// Remove deletes a specific model and all its files
func (s *Store) Remove(name string) error {
	return os.RemoveAll(path.Join(s.home, "models", s.encodeName(name)))
}

// Clean removes all stored models and the models directory
func (s *Store) Clean() error {
	return os.RemoveAll(path.Join(s.home, "models"))
}

// ModelfilePath returns the full path to a model's data file
func (s *Store) ModelfilePath(name string) (string, error) {
	model, err := s.getManifest(s.encodeName(name))
	if err != nil {
		return "", err
	}
	return path.Join(s.home, "models", s.encodeName(name), model.ModelFile), nil
}

func (s *Store) getManifest(encName string) (*types.Model, error) {
	dir := path.Join(s.home, "models")
	// Read manifest file
	data, e := os.ReadFile(path.Join(dir, encName, "nexa.manifest"))
	if e != nil {
		return nil, e
	}

	// Parse manifest JSON
	model := types.Model{}
	e = sonic.Unmarshal(data, &model)
	if e != nil {
		return nil, e
	}
	return &model, nil
}
