package types

import "io"

type FuncReadCloser struct {
	Raw io.ReadCloser
	F   func(p []byte)
}

func (w *FuncReadCloser) Read(p []byte) (n int, err error) {
	n, err = w.Raw.Read(p)
	w.F(p[:n])
	return
}

func (w *FuncReadCloser) Close() error {
	return w.Raw.Close()
}
