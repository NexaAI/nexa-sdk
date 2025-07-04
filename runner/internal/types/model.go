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
	Name string `json:"Name"`
	Size int64

	ModelFile  string
	MMProjFile string
	ExtraFiles []string
}

type ModelParam struct {
	CtxLen int32
	Device *string
}

type DownloadInfo struct {
	CurrentSize       int64
	CurrentDownloaded int64
	CurrentName       string
}
