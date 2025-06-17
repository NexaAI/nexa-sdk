package store

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/bytedance/sonic"

	"github.com/NexaAI/nexa-sdk/internal/types"
)

func (s *Store) List() ([]types.Model, error) {
	dir := s.modelDir()
	names, e := os.ReadDir(dir)
	if e != nil {
		return nil, e
	}

	res := make([]types.Model, 0, len(names))
	for _, name := range names {
		if !name.IsDir() {
			log.Printf("parse [%s] error: %s", name.Name(), fmt.Errorf("not a dir"))
			continue
		}
		data, e := os.ReadFile(path.Join(dir, name.Name(), "manifest"))
		if e != nil {
			log.Printf("parse [%s] error: %s", name.Name(), e)
			continue
		}

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

func (s *Store) Remove(name string) error {
	return os.RemoveAll(path.Join(s.modelDir(), s.encodeName(name)))
}

func (s *Store) Clean() error {
	return os.RemoveAll(s.modelDir())
}
