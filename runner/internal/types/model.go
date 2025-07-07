package types

type ModelType string

const (
	ModelTypeLLM      = "llm"
	ModelTypeVLM      = "vlm"
	ModelTypeEmbedder = "embedder"
	ModelTypeReranker = "reranker"
	ModelTypeImageGen = "image_gen"
)

type ModelManifest struct {
	Name  string
	Size  int64
	Quant string

	ModelFile  string
	MMProjFile string
	ExtraFiles []string
}

type ModelParam struct {
	CtxLen int32
	Device *string
}

type DownloadInfo struct {
	CurrentName       string
	CurrentSize       int64
	CurrentDownloaded int64
	TotalSize         int64
	TotalDownloaded   int64
}
