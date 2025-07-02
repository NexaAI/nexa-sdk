package store

import (
	"os"
	"path"

	"github.com/gofrs/flock"
)

func (s *Store) TryLockModel(modelName string) error {
	if modelName == "" {
		return ErrModelLocked
	}

	encName := s.encodeName(modelName)
	modelDir := path.Join(s.GetModelsDir(), encName)

	if err := os.MkdirAll(modelDir, 0o770); err != nil {
		return err
	}

	lockPath := path.Join(modelDir, ".model.lock")

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
