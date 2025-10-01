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

type ModelFileInfo struct {
	Name       string
	Downloaded bool
	Size       int64
}

type ModelManifest struct {
	Name          string // OrgName/RepoName
	ModelName     string // model arch name like "qwen3-4b", "yolov12", etc.
	ModelType     ModelType
	PluginId      string
	DeviceId      string
	MinSDKVersion string

	ModelFile     map[string]ModelFileInfo // quant -> modelfile
	MMProjFile    ModelFileInfo
	TokenizerFile ModelFileInfo
	ExtraFiles    []ModelFileInfo
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
	NCtx       int32
	NGpuLayers int32

	// npu only
	SystemPrompt string
}

type DownloadInfo struct {
	// CurrentFileName   string
	// CurrentDownloaded int64
	// CurrentSize       int64
	TotalDownloaded int64
	TotalSize       int64
}
