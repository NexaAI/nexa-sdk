package config

var modelMapping = map[string]string{
	"qwen3":     "Qwen/Qwen3-4B-GGUF",
	"qwen2vl":   "ggml-org/Qwen2-VL-2B-Instruct-GGUF",
	"qwen2.5vl": "Qwen/Qwen2.5-VL-3B-Instruct",
	"gemma3":    "ggml-org/gemma-3-4b-it-GGUF",
	"smolvlm":   "ggml-org/SmolVLM-500M-Instruct-GGUF",
}

func GetModelMapping(shortcut string) (string, bool) {
	actualPath, exists := modelMapping[shortcut]
	return actualPath, exists
}
