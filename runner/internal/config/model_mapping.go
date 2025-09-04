package config

var modelMapping = map[string]string{
	"qwen3":       "NexaAI/Qwen3-4B-GGUF",
	"qwen2vl":     "ggml-org/Qwen2-VL-2B-Instruct-GGUF",
	"qwen2.5vl":   "Qwen/Qwen2.5-VL-3B-Instruct",
	"gemma3":      "ggml-org/gemma-3-4b-it-GGUF",
	"smolvlm":     "ggml-org/SmolVLM-500M-Instruct-GGUF",
	"gpt-oss":     "NexaAI/gpt-oss-20b-GGUF",
	"gpt-oss-mlx": "NexaAI/gpt-oss-20b-MLX-4bit",

	// QNN
	"omni-neural":   "NexaAI/OmniNeural-4B",
	"qwen3-npu":     "NexaAI/qwen3-1.7B-npu",
	"qwen3-4B-npu":  "NexaAI/qwen3-4B-npu",
	"paddleocr-npu": "NexaAI/paddleocr-npu",
	"yolov12-npu":   "NexaAI/yolov12-npu",
	"parakeet-npu":  "NexaAI/parakeet-tdt-0.6b-v3-npu",
	"llama3-1B-npu": "NexaAI/llama3-1B-npu",
	"llama3-3B-npu": "NexaAI/llama3-3B-npu",

	"omni-neural-npu-encrypt": "nexaml/omni-neural-npu-encrypt",
	"qwen3-1.7B-npu-encrypt":  "nexaml/qwen3-1.7B-npu-encrypt",
	"qwen3-4B-npu-encrypt":    "nexaml/qwen3-4B-npu-encrypt",
	"paddleocr-npu-encrypt":   "nexaml/paddleocr-npu-encrypt",
	"yolov12-npu-encrypt":     "nexaml/yolov12-npu-encrypt",
}

func GetModelMapping(shortcut string) (string, bool) {
	actualPath, exists := modelMapping[shortcut]
	return actualPath, exists
}
