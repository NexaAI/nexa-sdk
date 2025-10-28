package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go"

	"github.com/NexaAI/nexa-sdk/runner/internal/types"
	nexa_sdk "github.com/NexaAI/nexa-sdk/runner/nexa-sdk"
	"github.com/NexaAI/nexa-sdk/runner/server/service"
)

type EmbeddingNewParams openai.EmbeddingNewParams

type EmbeddingRequest struct {
	EmbeddingNewParams

	TaskType string `json:"task_type"`
}

func defaultEmbeddingRequest() EmbeddingRequest {
	return EmbeddingRequest{
		TaskType: "default",
	}
}

func Embeddings(c *gin.Context) {
	param := defaultEmbeddingRequest()
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	slog.Info("Embeddings request received", "param", param)

	p, err := service.KeepAliveGet[nexa_sdk.Embedder](
		string(param.Model),
		types.ModelParam{},
		false,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error(), "code": nexa_sdk.SDKErrorCode(err)})
		return
	}

	// Convert input to the format expected by the embedder
	var texts []string
	switch {
	case param.Input.OfString.Value != "":
		texts = []string{param.Input.OfString.String()}
	case param.Input.OfArrayOfStrings != nil:
		texts = param.Input.OfArrayOfStrings
	default:
		c.JSON(http.StatusBadRequest, map[string]any{"error": "input must be a string or an array of strings"})
		return
	}

	slog.Debug("Embeddings called", "model", param.Model, "num_texts", len(texts))
	if len(texts) == 0 {
		c.JSON(http.StatusOK, nil)
		return
	}

	// Create embedder input
	embedInput := nexa_sdk.EmbedderEmbedInput{
		Texts:    texts,
		Config:   &nexa_sdk.EmbeddingConfig{},
		TaskType: param.TaskType,
	}

	res, err := p.Embed(embedInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error(), "code": nexa_sdk.SDKErrorCode(err)})
		return
	}

	embeddings := make([]openai.Embedding, len(texts))

	// Convert embeddings to the format expected by OpenAI API
	for i := range len(texts) {
		embeddingSlice := res.Embeddings[i]

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
