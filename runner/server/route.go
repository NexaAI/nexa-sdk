package server

import (
	"github.com/gin-gonic/gin"

	"github.com/NexaAI/nexa-sdk/server/docs"
	"github.com/NexaAI/nexa-sdk/server/handler"
	"github.com/NexaAI/nexa-sdk/server/middleware"
)

func RegisterRoot(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		c.String(200, "Nexa SDK is running")
	})
}

// @BasePath /v1

// http://localhost:18181/docs/ui/
func RegisterSwagger(r *gin.Engine) {
	g := r.Group("/docs")

	g.GET("/swagger.yaml", docs.SwaggerYAMLHandler())
	g.StaticFS("/ui", docs.FS)
}
func RegisterAPIv1(r *gin.Engine) {
	g := r.Group("/v1")

	g.Use(middleware.GIL)

	//g.POST("/saveKVCache", handler.SaveKVCache)
	//g.POST("/loadKVCache", handler.LoadKVCache)

	g.POST("/completions", handler.Completions)
	g.POST("/chat/completions", handler.ChatCompletions)

	g.POST("/embeddings", handler.Embeddings)

	g.POST("/reranking", handler.Reranking)

	g.GET("/models/*model", handler.RetrieveModel)
	g.POST("/models", handler.PullModel)
}
