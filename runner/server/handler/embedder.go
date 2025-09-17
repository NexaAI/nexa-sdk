package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go"

	"github.com/NexaAI/nexa-sdk/runner/internal/types"
	nexa_sdk "github.com/NexaAI/nexa-sdk/runner/nexa-sdk"
	"github.com/NexaAI/nexa-sdk/runner/server/service"
)

// @Router			/embeddings [post]
// @Summary		Creates an embedding for the given input.
// @Description	Creates an embedding for the given input.
// @Accept			json
// @Param			request	body	openai.EmbeddingNewParams	true	"Embedding request"
func Embeddings(c *gin.Context) {
	param := openai.EmbeddingNewParams{}
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	p, err := service.KeepAliveGet[nexa_sdk.Embedder](
		string(param.Model),
		types.ModelParam{},
		c.GetHeader("Nexa-KeepCache") != "true",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	// Convert input to the format expected by the embedder
	var texts []string
	if param.Input.OfArrayOfStrings != nil {
		texts = param.Input.OfArrayOfStrings
	} else {
		// Try to get string input
		texts = []string{param.Input.OfString.String()}
	}

	// Create embedder input
	embedInput := nexa_sdk.EmbedderEmbedInput{
		Texts: texts,
		Config: &nexa_sdk.EmbeddingConfig{
			BatchSize:       int32(len(texts)),
			Normalize:       true,
			NormalizeMethod: "l2",
		},
	}

	res, err := p.Embed(embedInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	// Convert embeddings to the format expected by OpenAI API
	// res.Embeddings is a flat array of float32 values
	// We need to group them by the number of texts
	embeddingDim := len(res.Embeddings) / len(texts)
	embeddings := make([]openai.Embedding, len(texts))
	
	for i := 0; i < len(texts); i++ {
		start := i * embeddingDim
		end := start + embeddingDim
		embeddingSlice := res.Embeddings[start:end]
		
		// Convert float32 to float64 for OpenAI API compatibility
		embeddingFloat64 := make([]float64, len(embeddingSlice))
		for j, val := range embeddingSlice {
			embeddingFloat64[j] = float64(val)
		}
		
		embeddings[i] = openai.Embedding{
			Embedding: embeddingFloat64,
			Index:     int64(i),
		}
	}

	response := map[string]interface{}{
		"object": "list",
		"data":   embeddings,
		"model":  param.Model,
		"usage": openai.CompletionUsage{
			PromptTokens: int64(len(texts)),
			TotalTokens:  int64(len(texts)),
		},
	}

	c.JSON(http.StatusOK, response)
}
