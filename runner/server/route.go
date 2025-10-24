package server

import (
	"github.com/gin-gonic/gin"

	"github.com/NexaAI/nexa-sdk/runner/server/docs"
	"github.com/NexaAI/nexa-sdk/runner/server/handler"
	"github.com/NexaAI/nexa-sdk/runner/server/middleware"
)

func RegisterRoot(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		c.String(200, "Nexa SDK is running")
	})
}

// http://localhost:18181/docs/ui/
func RegisterSwagger(r *gin.Engine) {
	g := r.Group("/docs")
	g.GET("/swagger.yaml", docs.SwaggerYAMLHandler())
	g.StaticFS("/ui", docs.FS)
}

func RegisterAPIv1(r *gin.Engine) {
	g := r.Group("/v1")
	g.Use(middleware.GIL)

	g.POST("/completions", handler.Completions)
	g.POST("/chat/completions", handler.ChatCompletions)
	g.POST("/embeddings", handler.Embeddings)
	g.POST("/images/generations", handler.ImageGenerations)
	g.POST("/reranking", handler.Reranking)

	g.GET("/models/*model", handler.RetrieveModel)
	g.GET("/models", handler.ListModels)
}
