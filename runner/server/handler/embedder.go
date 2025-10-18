// Copyright (c) 2025 Nexa AI
//
// LICENSE NOTICE - DUAL LICENSING:
// - NPU models and inference: CC-BY-NC 4.0 (NON-COMMERCIAL USE ONLY)
// - GPU/CPU models and inference: Apache 2.0 (FREE FOR ALL USE)

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go"

	"github.com/NexaAI/nexa-sdk/runner/internal/types"
	nexa_sdk "github.com/NexaAI/nexa-sdk/runner/nexa-sdk"
	"github.com/NexaAI/nexa-sdk/runner/server/service"
)

func Embeddings(c *gin.Context) {
	param := openai.EmbeddingNewParams{}
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

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
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error(), "code": nexa_sdk.SDKErrorCode(err)})
		return
	}

	dimOutput, err := p.EmbeddingDimension()
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error(), "code": nexa_sdk.SDKErrorCode(err)})
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
