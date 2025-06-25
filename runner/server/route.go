package server

import (
	"github.com/gin-gonic/gin"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware

	docs "github.com/NexaAI/nexa-sdk/server/docs"
	"github.com/NexaAI/nexa-sdk/server/handler"
	"github.com/NexaAI/nexa-sdk/server/middleware"
)

// @BasePath /v1

// http://localhost:18181/docs/index.html
func RegisterSwagger(r *gin.Engine) {
	docs.SwaggerInfo.BasePath = "/v1"
	g := r.Group("/docs")

	g.GET("/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

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
