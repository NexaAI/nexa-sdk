package types

type FuncWriter struct {
	F func([]byte) (int, error)
}

func (w FuncWriter) Write(p []byte) (n int, err error) {
	return w.F(p)
}
