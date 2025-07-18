package types

type ModelType string

const (
	ModelTypeLLM      = "llm"
	ModelTypeVLM      = "vlm"
	ModelTypeEmbedder = "embedder"
	ModelTypeReranker = "reranker"
	ModelTypeImageGen = "image_gen"
	ModelTypeTTS      = "tts"
	ModelTypeASR      = "asr"
)

type ModeFileInfo struct {
	Name       string
	Downloaded bool
	Size       int64
}

type ModelManifest struct {
	Name string

	ModelFile  map[string]ModeFileInfo // quant -> modelfile
	MMProjFile ModeFileInfo
	ExtraFiles []ModeFileInfo
}

func (m ModelManifest) GetSize() int64 {
	var count int64

	for _, v := range m.ModelFile {
		if v.Downloaded {
			count += v.Size
		}
	}
	if m.MMProjFile.Downloaded {
		count += m.MMProjFile.Size
	}
	for _, v := range m.ExtraFiles {
		if v.Downloaded {
			count += v.Size
		}
	}

	return count
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
