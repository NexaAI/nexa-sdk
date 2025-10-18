// Copyright (c) 2025 Nexa AI
//
// LICENSE NOTICE - DUAL LICENSING:
// - NPU models and inference: CC-BY-NC 4.0 (NON-COMMERCIAL USE ONLY)
// - GPU/CPU models and inference: Apache 2.0 (FREE FOR ALL USE)

package store

import "errors"

// File lock related errors
var (
	ErrModelNameEmpty = errors.New("model name is empty")
	ErrModelLocked    = errors.New("model is currently locked by another process")
)
