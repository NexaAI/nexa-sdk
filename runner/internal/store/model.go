package store

import "path"

// ModelfilePath returns the full path to a model's data file
func (s *Store) ModelfilePath(name string) string {
	return path.Join(s.home, "models", s.encodeName(name), "modelfile")
}
