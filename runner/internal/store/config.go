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
)

func (s *Store) ConfigGet(key string) (string, error) {
	slog.Debug("ConfigGet", "key", key)
	data, err := os.ReadFile(filepath.Join(s.home, "config"))
	return string(data), err
}

func (s *Store) ConfigSet(key string, value string) error {
	slog.Debug("ConfigSet", "key", key, "value", value)
	return os.WriteFile(filepath.Join(s.home, "config"), []byte(value), 0o600)
}
