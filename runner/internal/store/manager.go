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
// │   └─ model_name (base64 encoded)
// │      ├─ modelfile (actual model data)
// │      └─ manifest (model metadata)
//
// TODO: implement file locking for concurrent access
// TODO: clean corrupted directories on initialization
type Store struct {
	home string // base cache directory path
}

// NewStore creates and initializes a new Store instance
func NewStore() Store {
	s := Store{}
	s.init()
	return s
}

// init sets up the store's directory structure
func (s *Store) init() {
	// Get user's cache directory (OS-specific)
	cacheDir, e := os.UserCacheDir()
	if e != nil {
		panic(e)
	}

	// Set nexa cache directory
	s.home = path.Join(cacheDir, "nexa")

	// Create models directory structure
	e = os.MkdirAll(s.modelDir(), 0o770)
	if e != nil {
		panic(e)
	}
}

// modelDir returns the path to the models directory
func (s *Store) modelDir() string {
	return path.Join(s.home, "models")
}

// encodeName encodes model names to safe filesystem names using base64
func (s *Store) encodeName(name string) string {
	return base64.URLEncoding.EncodeToString([]byte(name))
}
