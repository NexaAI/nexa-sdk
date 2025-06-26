package types

type Model struct {
	Name      string
	Size      int64
	ModelFile string
}

type ModelFile struct {
	Name string
	Size int64
}

type DownloadInfo struct {
	TotalSize         int64
	TotalDownloaded   int64
	CurrentSize       int64
	CurrentDownloaded int64
	CurrentName       string
}
