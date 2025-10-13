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
	BatchSize       int32    `json:"batch_size" binding:"required"`
	NormalizeMethod string   `json:"normalize_method" binding:"required"`
	Normalize       bool     `json:"normalize" binding:"required"`
}

type RerankResponse struct {
	Result []float32 `json:"result"`
}

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
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error(), "code": nexa_sdk.SDKErrorCode(err)})
		return
	}

	res, err := p.Rerank(nexa_sdk.RerankerRerankInput{
		Query:     param.Query,
		Documents: param.Documents,
		Config: &nexa_sdk.RerankConfig{
			BatchSize:       param.BatchSize,
			Normalize:       param.Normalize,
			NormalizeMethod: param.NormalizeMethod,
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error(), "code": nexa_sdk.SDKErrorCode(err)})
	} else {
		c.JSON(http.StatusOK, RerankResponse{Result: res.Scores})
	}
}
