package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	nexa_sdk "github.com/NexaAI/nexa-sdk/nexa-sdk"
	"github.com/NexaAI/nexa-sdk/server/service"
)

type RerankingRequest struct {
	Model     string   `json:"model"`
	Query     string   `json:"query"`
	Documents []string `json:"documents"`
}

type RerankResponse struct {
	Result []float32 `json:"result"`
}

// curl -v http://localhost:18181/v1/reranking -d '{ "model": "Qwen/Qwen3-0.6B-GGUF", "query" : "hi", "documents": ["hello","world"] }'
func Reranking(c *gin.Context) {
	param := RerankingRequest{}
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	p, err := service.KeepAliveGet[nexa_sdk.Reranker](
		string(param.Model),
		service.ModelParam{},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	res, err := p.Rerank(param.Query, param.Documents)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
	} else {
		c.JSON(http.StatusOK, RerankResponse{Result: res})
	}
}
