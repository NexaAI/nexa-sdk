package types

type ModelType string

const (
	ModelTypeLLM      = "llm"
	ModelTypeVLM      = "vlm"
	ModelTypeEmbedder = "embedder"
	ModelTypeReranker = "reranker"
)

type ModelManifest struct {
	Name          string `json:"Name"`
	Size          int64
	ModelType     ModelType
	ModelFile     string
	TokenizerFile string
	ExtraFiles    []string
}

type ModelParam struct {
	Device *string
	CtxLen int32
}

type DownloadInfo struct {
	CurrentSize       int64
	CurrentDownloaded int64
	CurrentName       string
}
