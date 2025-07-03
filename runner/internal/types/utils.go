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
