package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/NexaAI/nexa-sdk/runner/internal/types"
	nexa_sdk "github.com/NexaAI/nexa-sdk/runner/nexa-sdk"
	"github.com/NexaAI/nexa-sdk/runner/server/service"
)

type RerankingRequest struct {
	Model           string   `json:"model" binding:"required"`
	Query           string   `json:"query"`
	Documents       []string `json:"documents"`
	BatchSize       int32    `json:"batch_size"`
	NormalizeMethod string   `json:"normalize_method"`
	Normalize       bool     `json:"normalize"`
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

	slog.Info("Reranking request received",
		"model", param.Model,
		"query", param.Query,
		"documents", param.Documents,
	)

	p, err := service.KeepAliveGet[nexa_sdk.Reranker](
		string(param.Model),
		types.ModelParam{},
		false,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error(), "code": nexa_sdk.SDKErrorCode(err)})
		return
	}

	if param.Query == "" || len(param.Documents) == 0 {
		if param.Query != "" || len(param.Documents) != 0 {
			c.JSON(http.StatusBadRequest, map[string]any{"error": "both query and documents must be provided"})
			return
		}
		c.JSON(http.StatusOK, nil)
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
