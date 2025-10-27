package server

import (
	"net/http"

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

	// ==== legacy ====
	g.POST("/completions", func(c *gin.Context) {
		c.JSON(http.StatusGone, map[string]any{"error": "this endpoint is deprecated, please use /chat/completions instead"})
	})

	// ==== openai compatible ====
	g.POST("/chat/completions", handler.ChatCompletions)
	g.POST("/embeddings", handler.Embeddings)
	g.POST("/audio/speech", handler.Speech)
	g.POST("/audio/transcriptions", handler.Transcriptions)
	g.POST("/images/generations", handler.ImageGenerations)
	// ==== nexa specific ====
	g.POST("/audio/diarize", handler.Diarize)
	g.POST("/reranking", handler.Reranking)
	g.POST("/cv", handler.CV)

	// ==== model management ====
	g.GET("/models/*model", handler.RetrieveModel)
	g.GET("/models", handler.ListModels)
}
