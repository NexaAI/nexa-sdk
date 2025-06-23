package server

import (
	"github.com/gin-gonic/gin"

	"github.com/NexaAI/nexa-sdk/server/handler"
	"github.com/NexaAI/nexa-sdk/server/middleware"
)

func RegisterAPIv1(r *gin.Engine) {
	g := r.Group("/v1")

	g.Use(middleware.GIL)

	g.POST("/saveKVCache", handler.SaveKVCache)
	g.POST("/loadKVCache", handler.LoadKVCache)

	g.POST("/completions", handler.Completions)
	g.POST("/chat/completions", handler.ChatCompletions)
}
