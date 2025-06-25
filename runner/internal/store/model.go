package store

import (
	"encoding/base64"
	"path"
)

// encodeName encodes model names to safe filesystem names using base64
func (s *Store) encodeName(name string) string {
	return base64.URLEncoding.EncodeToString([]byte(name))
}

// ModelfilePath returns the full path to a model's data file
func (s *Store) ModelfilePath(name string) string {
	return path.Join(s.home, "models", s.encodeName(name), "modelfile")
}

// modelDir returns the path to the models directory
func (s *Store) CachefilePath(name string) string {
	return path.Join(s.home, "cache", name)
}
