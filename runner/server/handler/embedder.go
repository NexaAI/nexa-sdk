package handler

import (
	"net/http"

	"github.com/NexaAI/nexa-sdk/internal/store"
	nexa_sdk "github.com/NexaAI/nexa-sdk/nexa-sdk"
	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go"
)

// curl -v http://localhost:18181/v1/embeddings -d '{ "model": "Qwen/Qwen3-0.6B-GGUF", "input": ["hello","world"] }'
func Embeddings(c *gin.Context) {
	param := openai.EmbeddingNewParams{}
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	s := store.NewStore()

	file, err := s.ModelfilePath(param.Model)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	p := nexa_sdk.NewEmbedder(file, nil, nil)
	defer p.Destroy()

	res, err := p.Embed(param.Input.OfArrayOfStrings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
	} else {
		embeddings := make([]float64, len(res))
		for i := range res {
			embeddings[i] = float64(res[i])
		}
		c.JSON(http.StatusOK, openai.Embedding{Embedding: embeddings})
	}
}
