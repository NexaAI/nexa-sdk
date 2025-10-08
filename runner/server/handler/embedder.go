package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go"

	"github.com/NexaAI/nexa-sdk/runner/internal/config"
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
	param := struct {
		openai.EmbeddingNewParams
		KeepAlive *int64 `json:"keep_alive"`
	}{}

	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	keepAlive := config.Get().KeepAlive
	if param.KeepAlive != nil {
		keepAlive = *param.KeepAlive
	}
	p, err := service.KeepAliveGet[nexa_sdk.Embedder](
		string(param.Model),
		types.ModelParam{},
		false,
		keepAlive,
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
		texts = []string{param.Input.OfString.String()}
	}

	numTexts := len(texts)
	if numTexts == 0 {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "no text content found in input"})
		return
	}

	// Create embedder input
	embedInput := nexa_sdk.EmbedderEmbedInput{
		Texts: texts,
		Config: &nexa_sdk.EmbeddingConfig{
			BatchSize:       int32(numTexts),
			Normalize:       true,
			NormalizeMethod: "l2",
		},
	}

	res, err := p.Embed(embedInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	dimOutput, err := p.EmbeddingDimension()
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	embeddingDim := int(dimOutput.Dimension)
	embeddings := make([]openai.Embedding, numTexts)

	// Convert embeddings to the format expected by OpenAI API
	// res.Embeddings is a flat array of float32 values
	// We need to group them by the number of texts
	for i := range numTexts {
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

	response := openai.CreateEmbeddingResponse{
		Data:  embeddings,
		Model: param.Model,
		Usage: openai.CreateEmbeddingResponseUsage{
			PromptTokens: res.ProfileData.PromptTokens,
			TotalTokens:  res.ProfileData.TotalTokens(),
		},
	}

	c.JSON(http.StatusOK, response)
}
