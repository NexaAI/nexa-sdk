// Copyright (c) 2025 Nexa AI
//
// LICENSE NOTICE - DUAL LICENSING:
// - NPU models and inference: CC-BY-NC 4.0 (NON-COMMERCIAL USE ONLY)
// - GPU/CPU models and inference: Apache 2.0 (FREE FOR ALL USE)

package store

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/gofrs/flock"
)

func (s *Store) LockModel(modelName string) error {
	slog.Debug("LockModel", "modelName", modelName)

	if modelName == "" {
		return ErrModelNameEmpty
	}

	modelsDir := s.ModelDirPath()
	os.MkdirAll(filepath.Join(modelsDir, modelName), 0o770)
	lockPath := filepath.Join(modelsDir, modelName+".lock")

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
	slog.Debug("UnLockModel", "modelName", modelName)

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
