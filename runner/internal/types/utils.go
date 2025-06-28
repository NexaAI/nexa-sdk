package types

import "io"

type TeeReadCloserF struct {
	Raw     io.ReadCloser
	WriterF func(p []byte) (int, error)
}

func (w *TeeReadCloserF) Read(p []byte) (n int, err error) {
	n, err = w.Raw.Read(p)
	w.WriterF(p[:n])
	return
}

func (w *TeeReadCloserF) Close() error {
	return w.Raw.Close()
}
