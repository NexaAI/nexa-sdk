package types

type ModelType string

const (
	ModelTypeLLM      ModelType = "llm"
	ModelTypeVLM      ModelType = "vlm"
	ModelTypeEmbedder ModelType = "embedder"
	ModelTypeReranker ModelType = "reranker"
	ModelTypeImageGen ModelType = "image_gen"
	ModelTypeTTS      ModelType = "tts"
	ModelTypeASR      ModelType = "asr"
	ModelTypeCV       ModelType = "cv"
)

type ModeFileInfo struct {
	Name       string
	Downloaded bool
	Size       int64
}

type ModelManifest struct {
	Name      string
	ModelType ModelType
	PluginId  string

	ModelFile     map[string]ModeFileInfo // quant -> modelfile
	MMProjFile    ModeFileInfo
	TokenizerFile ModeFileInfo
	ExtraFiles    []ModeFileInfo
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
	if m.TokenizerFile.Downloaded {
		count += m.TokenizerFile.Size
	}
	for _, v := range m.ExtraFiles {
		if v.Downloaded {
			count += v.Size
		}
	}

	return count
}

type ModelParam struct {
	NCtx int32
}

type DownloadInfo struct {
	CurrentName       string
	CurrentSize       int64
	CurrentDownloaded int64
	TotalSize         int64
	TotalDownloaded   int64
}
