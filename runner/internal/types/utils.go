// Copyright (c) 2025 Nexa AI
//
// LICENSE NOTICE - DUAL LICENSING:
// - NPU models and inference: CC-BY-NC 4.0 (NON-COMMERCIAL USE ONLY)
// - GPU/CPU models and inference: Apache 2.0 (FREE FOR ALL USE)

package types

import "io"

type TeeReadF struct {
	Raw     io.Reader
	WriterF func(p []byte) (int, error)
}

func (w *TeeReadF) Read(p []byte) (n int, err error) {
	n, err = w.Raw.Read(p)
	w.WriterF(p[:n])
	return
}
