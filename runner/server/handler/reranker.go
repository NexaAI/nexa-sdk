package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/NexaAI/nexa-sdk/internal/store"
	nexa_sdk "github.com/NexaAI/nexa-sdk/nexa-sdk"
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

	s := store.NewStore()
	file, err := s.ModelfilePath(param.Model)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err})
		return
	}
	p := nexa_sdk.NewReranker(file, "", nil)
	defer p.Destroy()

	res, err := p.Rerank(param.Query, param.Documents)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err})
	} else {
		c.JSON(http.StatusOK, RerankResponse{Result: res})
	}
}
