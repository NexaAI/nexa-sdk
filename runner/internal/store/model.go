package store

import (
	"encoding/base64"
	"path"
)

// encodeName encodes model names to safe filesystem names using base64
func (s *Store) encodeName(name string) string {
	return base64.URLEncoding.EncodeToString([]byte(name))
}

// modelDir returns the path to the models directory
func (s *Store) CachefilePath(name string) string {
	return path.Join(s.home, "cache", name)
}
