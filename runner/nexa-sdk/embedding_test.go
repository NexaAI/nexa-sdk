package nexa_sdk

import (
	"log/slog"
	"testing"
)

var embedder *Embedder

func initEmbedder() {
	slog.Debug("initEmbedder called")

	var err error

	// input := EmbedderCreateInput{
	// 	ModelPath: "modelfiles/llama_cpp/jina-embeddings-v2-small-en-Q4_K_M.gguf",
	// 	PluginID:  "llama_cpp",
	// }

	input := EmbedderCreateInput{
		ModelPath: "modelfiles/mlx/jina-v2-fp16-mlx/model.safetensors",
		PluginID:  "mlx",
	}

	embedder, err = NewEmbedder(input)
	if err != nil {
		panic("Error creating Embedder: " + err.Error())
	}
}

func deinitEmbedder() {
	embedder.Destroy()
}

func TestEmbedderEmbeddingDimension(t *testing.T) {
	output, err := embedder.EmbeddingDimension()
	if err != nil {
		t.Errorf("EmbeddingDimension failed: %v", err)
		return
	}

	t.Logf("Embedding dimension: %d", output.Dimension)
}

func TestEmbedderEmbed(t *testing.T) {
	input := EmbedderEmbedInput{
		Texts: []string{
			"Hello, this is a test sentence.",
			"Another test sentence for embedding.",
			"Third sentence to test batch processing.",
		},
		Config: &EmbeddingConfig{
			BatchSize:       2,
			Normalize:       true,
			NormalizeMethod: "l2",
		},
	}

	output, err := embedder.Embed(input)
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	t.Logf("Embed completed successfully")
	t.Logf("Total embeddings: %d", len(output.Embeddings))

	if len(output.Embeddings) > 0 {
		// Safely access first 10 elements or all elements if less than 10
		displayCount := 10
		if len(output.Embeddings) < displayCount {
			displayCount = len(output.Embeddings)
		}
		t.Logf("First embedding values: %v", output.Embeddings[:displayCount])
	}
}
