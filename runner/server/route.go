package server

import (
	"github.com/gin-gonic/gin"

	"github.com/NexaAI/nexa-sdk/server/handler"
)

func RegisterAPIv1(r *gin.Engine) {
	g := r.Group("/v1")

	g.POST("/completions", handler.Completions)
	g.POST("/chat/completions", handler.ChatCompletions)
}
