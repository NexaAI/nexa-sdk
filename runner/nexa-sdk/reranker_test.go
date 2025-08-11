package nexa_sdk

import (
	"log/slog"
	"testing"
)

var reranker *Reranker

func initReranker() {
	slog.Debug("initReranker called")

	var err error

	// Note: Replace with actual reranker model path when available
	input := RerankerCreateInput{
		ModelPath: "modelfiles/mlx/jina-v2-rerank-mlx/jina-reranker-v2-base-multilingual-f16.safetensors",
		PluginID:  "mlx",
	}

	reranker, err = NewReranker(input)
	if err != nil {
		slog.Debug("Reranker model not available, skipping test")
		return
	}
}

func deinitReranker() {
	if reranker != nil {
		reranker.Destroy()
	}
}

func TestRerankerRerank(t *testing.T) {
	if reranker == nil {
		t.Skip("Reranker not initialized, skipping test")
	}

	input := RerankerRerankInput{
		Query: "What is machine learning?",
		Documents: []string{
			"Machine learning is a subset of artificial intelligence.",
			"Machine learning algorithms learn patterns from data.",
			"The weather is sunny today.",
			"Deep learning is a type of machine learning.",
		},
		Config: &RerankConfig{
			BatchSize:       4,
			Normalize:       true,
			NormalizeMethod: "softmax",
		},
	}

	output, err := reranker.Rerank(input)
	if err != nil {
		t.Fatalf("Rerank failed: %v", err)
	}

	t.Logf("Rerank completed successfully")
	t.Logf("Total scores: %d", len(output.Scores))

	for i, score := range output.Scores {
		t.Logf("Document %d score: %.6f", i+1, score)
	}

	t.Logf("Profile data: %#v", output.ProfileData)
}
