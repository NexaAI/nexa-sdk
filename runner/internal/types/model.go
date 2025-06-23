package types

type Model struct {
	Name string
	Size uint64
}

type DownloadInfo struct {
	FileSize   uint64
	Downloaded uint64
}

type FuncWriter struct {
	f func([]byte) (int, error)
}

func (w *FuncWriter) Write(p []byte) (n int, err error) {
	return w.f(p)
}
