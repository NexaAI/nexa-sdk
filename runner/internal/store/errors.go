package store

import "errors"

// File lock related errors
var (
	ErrModelLocked = errors.New("model is currently locked by another process")
	ErrStoreLocked = errors.New("store is currently locked by another process")
) 