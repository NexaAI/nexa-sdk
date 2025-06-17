package store

import (
	"encoding/base64"
	"os"
	"path"
)

// TODO: file lock
// TODO: clean bad dir on init

// │.
// │└─ models
// │   └─ model_name
// │      ├─ modelfile
// │      └─ manifest
type Store struct {
	home string
}

func NewStore() Store {
	s := Store{}
	s.init()
	return s
}

func (s *Store) init() {
	cacheDir, e := os.UserCacheDir()
	if e != nil {
		panic(e)
	}

	s.home = path.Join(cacheDir, "nexa")

	e = os.MkdirAll(s.modelDir(), 0o770)
	if e != nil {
		panic(e)
	}
}

func (s *Store) modelDir() string {
	return path.Join(s.home, "models")
}

func (s *Store) encodeName(name string) string {
	return base64.URLEncoding.EncodeToString([]byte(name))
}
