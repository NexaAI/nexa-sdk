package config

import (
	"runtime"
	"sync"
)

// key: shortcut
// value:
//   - list of (os, arch, actual model_name)
//   - os and arch can be empty string to match any
//   - the later value will override the previous one if both match
var modelMappingSrc = map[string][][3]string{
	"qwen3": {
		{"", "", "NexaAI/Qwen3-4B-GGUF"},
	},
	"qwen2vl": {
		{"", "", "ggml-org/Qwen2-VL-2B-Instruct-GGUF"},
	},
	"qwen2.5vl": {
		{"", "", "unsloth/Qwen2.5-VL-3B-Instruct-GGUF"},
		{"darwin", "arm64", "Qwen/Qwen2.5-VL-3B-Instruct"},
	},
	"qwen3vl": {
		{"windows", "amd64", "NexaAI/qwen3vl-GGUF"},
		{"windows", "arm64", "NexaAI/qwen3vl-npu"},
		{"darwin", "arm64", "NexaAI/qwen3vl-mlx-4bit"},
	},
	"gemma3": {
		{"", "", "ggml-org/gemma-3-4b-it-GGUF"},
	},
	"smolvlm": {
		{"", "", "ggml-org/SmolVLM-500M-Instruct-GGUF"},
	},
	"gpt-oss": {
		{"", "", "NexaAI/gpt-oss-20b-GGUF"},
	},
	"gpt-oss-mlx": {
		{"darwin", "arm64", "NexaAI/gpt-oss-20b-MLX-4bit"},
	},
	"omni-neural": {
		{"windows", "arm64", "NexaAI/OmniNeural-4B"},
	},
}

var modelMappingInit sync.Once

var modelMapping = make(map[string]string)

func GetModelMapping(shortcut string) (string, bool) {
	modelMappingInit.Do(func() {
		for model, aliases := range modelMappingSrc {
			for _, entry := range aliases {
				if (entry[0] == "" || entry[0] == runtime.GOOS) && (entry[1] == "" || entry[1] == runtime.GOARCH) {
					modelMapping[model] = entry[2]
				}
			}
		}
	})

	actualPath, exists := modelMapping[shortcut]
	return actualPath, exists
}
