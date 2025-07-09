package store

import (
	"path"

	"github.com/gofrs/flock"
)

func (s *Store) LockModel(modelName string) error {
	if modelName == "" {
		return ErrModelNameEmpty
	}

	encName := s.encodeName(modelName)
	modelsDir := s.GetModelsDir()
	lockPath := path.Join(modelsDir, "."+encName+".lock")

	fl := flock.New(lockPath)

	locked, err := fl.TryLock()
	if err != nil {
		return err
	}
	if !locked {
		return ErrModelLocked
	}

	s.modelLocks.Store(modelName, fl)
	return nil
}

func (s *Store) UnlockModel(modelName string) error {
	if modelName == "" {
		return nil
	}

	if info, ok := s.modelLocks.Load(modelName); ok {
		fl := info.(*flock.Flock)
		fl.Unlock()
		s.modelLocks.Delete(modelName)
	}

	return nil
}
