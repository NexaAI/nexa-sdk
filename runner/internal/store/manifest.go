package store

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/bytedance/sonic"

	"github.com/NexaAI/nexa-sdk/internal/types"
)

// List returns all locally stored models by reading their manifest files
func (s *Store) List() ([]types.Model, error) {
	dir := s.modelDir()
	names, e := os.ReadDir(dir)
	if e != nil {
		return nil, e
	}

	// Parse each model directory's manifest
	res := make([]types.Model, 0, len(names))
	for _, name := range names {
		// Skip non-directory entries
		if !name.IsDir() {
			log.Printf("parse [%s] error: %s", name.Name(), fmt.Errorf("not a dir"))
			continue
		}

		// Read manifest file
		data, e := os.ReadFile(path.Join(dir, name.Name(), "manifest"))
		if e != nil {
			log.Printf("parse [%s] error: %s", name.Name(), e)
			continue
		}

		// Parse manifest JSON
		model := types.Model{}
		e = sonic.Unmarshal(data, &model)
		if e != nil {
			log.Printf("parse [%s] error: %s", name.Name(), e)
			continue
		}

		res = append(res, model)
	}

	return res, nil
}

// Remove deletes a specific model and all its files
func (s *Store) Remove(name string) error {
	return os.RemoveAll(path.Join(s.modelDir(), s.encodeName(name)))
}

// Clean removes all stored models and the models directory
func (s *Store) Clean() error {
	return os.RemoveAll(s.modelDir())
}
