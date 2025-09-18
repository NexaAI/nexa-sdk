package config

var modelMapping = map[string]string{
	"qwen3":       "NexaAI/Qwen3-4B-GGUF",
	"qwen2vl":     "ggml-org/Qwen2-VL-2B-Instruct-GGUF",
	"qwen2.5vl":   "Qwen/Qwen2.5-VL-3B-Instruct",
	"gemma3":      "ggml-org/gemma-3-4b-it-GGUF",
	"smolvlm":     "ggml-org/SmolVLM-500M-Instruct-GGUF",
	"gpt-oss":     "NexaAI/gpt-oss-20b-GGUF",
	"gpt-oss-mlx": "NexaAI/gpt-oss-20b-MLX-4bit",
	"omni-neural": "NexaAI/OmniNeural-4B",
}

func GetModelMapping(shortcut string) (string, bool) {
	actualPath, exists := modelMapping[shortcut]
	return actualPath, exists
}
