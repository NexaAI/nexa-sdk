package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go"

	"github.com/NexaAI/nexa-sdk/internal/types"
	nexa_sdk "github.com/NexaAI/nexa-sdk/nexa-sdk"
	"github.com/NexaAI/nexa-sdk/server/service"
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
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

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
