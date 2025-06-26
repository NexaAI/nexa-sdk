package types

type Model struct {
	Name      string
	Size      uint64
	ModelFile string
}

type ModelFile struct {
	Name string
	Size uint64
}

type DownloadInfo struct {
	Size       uint64
	Downloaded uint64
}
