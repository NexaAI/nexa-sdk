package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/NexaAI/nexa-sdk/runner/internal/types"
	nexa_sdk "github.com/NexaAI/nexa-sdk/runner/nexa-sdk"
	"github.com/NexaAI/nexa-sdk/runner/server/service"
)

type RerankingRequest struct {
	Model     string   `json:"model" default:"Qwen/Qwen3-0.6B-GGUF"`
	Query     string   `json:"query" default:"hi"`
	Documents []string `json:"documents" default:"hello,world"`
}

type RerankResponse struct {
	Result []float32 `json:"result" default:"[0.1,0.2,0.3]"`
}

// @Router			/reranking [post]
// @Summary		Reranks the given documents for the given query.
// @Description	Reranks the given documents for the given query.
// @Accept			json
// @Param			request	body	RerankingRequest	true	"Reranking request"
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

	rerankInput := nexa_sdk.RerankerRerankInput{
		Query:     param.Query,
		Documents: param.Documents,
		Config: &nexa_sdk.RerankConfig{
			BatchSize:       int32(len(param.Documents)),
			Normalize:       true,
			NormalizeMethod: "softmax",
		},
	}

	res, err := p.Rerank(rerankInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
	} else {
		c.JSON(http.StatusOK, RerankResponse{Result: res.Scores})
	}
}
