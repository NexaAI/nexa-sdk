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
