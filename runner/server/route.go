package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/NexaAI/nexa-sdk/server/handler"
	"github.com/NexaAI/nexa-sdk/server/middleware"
	swagger_ui "github.com/NexaAI/nexa-sdk/server/swagger-ui"
)

// @BasePath /v1

// http://localhost:18181/docs/ui/
func RegisterSwagger(r *gin.Engine) {
	g := r.Group("/docs")

	g.GET("/swagger.json", swagger_ui.SwaggerJSONHandler())

	g.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/docs/ui/")
	})

	g.StaticFS("/ui", swagger_ui.FS)
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
}
