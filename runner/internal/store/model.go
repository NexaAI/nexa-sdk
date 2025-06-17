package store

import "path"

func (s *Store) ModelfilePath(name string) string {
	return path.Join(s.home, "models", s.encodeName(name), "model")
}
