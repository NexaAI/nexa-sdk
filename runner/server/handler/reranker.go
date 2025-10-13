package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/NexaAI/nexa-sdk/runner/internal/types"
	nexa_sdk "github.com/NexaAI/nexa-sdk/runner/nexa-sdk"
	"github.com/NexaAI/nexa-sdk/runner/server/service"
)

type RerankingRequest struct {
	Model           string   `json:"model" binding:"required"`
	Query           string   `json:"query" binding:"required"`
	Documents       []string `json:"documents" binding:"required"`
	NormalizeMethod string   `json:"normalize_method" binding:"required"`
	Normalize       bool     `json:"normalize" binding:"required"`
}

type RerankResponse struct {
	Result []float32 `json:"result"`
}

// @Router			/reranking [post]
// @Summary		Rerank documents based on their relevance to a query
// @Description	Reranks a list of documents according to their semantic relevance to the provided query. Returns relevance scores for each document. Higher scores indicate greater relevance.
// @Tags			Reranking
// @Accept			json
// @Produce		json
// @Param			request	body		RerankingRequest	true	"Reranking request with model, query, documents and normalization settings"
// @Success		200		{object}	RerankResponse		"Successfully reranked documents with relevance scores"
// @Failure		400		{object}	map[string]any		"Invalid request parameters"
// @Failure		500		{object}	map[string]any		"Internal server error"
func Reranking(c *gin.Context) {
	param := RerankingRequest{}
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	p, err := service.KeepAliveGet[nexa_sdk.Reranker](
		string(param.Model),
		types.ModelParam{},
		false,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	res, err := p.Rerank(nexa_sdk.RerankerRerankInput{
		Query:     param.Query,
		Documents: param.Documents,
		Config: &nexa_sdk.RerankConfig{
			BatchSize:       int32(len(param.Documents)),
			Normalize:       param.Normalize,
			NormalizeMethod: param.NormalizeMethod,
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
	} else {
		c.JSON(http.StatusOK, RerankResponse{Result: res.Scores})
	}
}
