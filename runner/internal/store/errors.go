package store

import "errors"

// File lock related errors
var (
	ErrModelNameEmpty = errors.New("model name is empty")
	ErrModelLocked    = errors.New("model is currently locked by another process")
)

