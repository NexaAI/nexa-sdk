package types

type ModelType string

const (
	ModelTypeLLM      = "llm"
	ModelTypeVLM      = "vlm"
	ModelTypeEmbedder = "embedder"
	ModelTypeReranker = "reranker"
)

type Model struct {
	Name          string
	Size          int64
	ModelType     ModelType
	ModelFile     string
	TokenizerFile string
}

type ModelParam struct {
	Device *string
	CtxLen int32
}

type DownloadInfo struct {
	TotalSize         int64
	TotalDownloaded   int64
	CurrentSize       int64
	CurrentDownloaded int64
	CurrentName       string
}
