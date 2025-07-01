package store

import (
	"encoding/base64"
	"fmt"
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

		name, err := base64.URLEncoding.DecodeString(encName.Name())
		if err != nil {
			fmt.Printf("getManifest error: %s\n", err)
			continue
		}

		model, err := s.GetManifest(string(name))
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

func (s *Store) GetManifest(name string) (*types.ModelManifest, error) {
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
