package server

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/NexaAI/nexa-sdk/internal/config"
	"github.com/NexaAI/nexa-sdk/server/service"
)

// @Title		Nexa AI Server
// @Version	0.0.0
// @BasePath	/v1
func Serve() {
	service.Init()
	defer service.DeInit()

	gin.SetMode(gin.ReleaseMode)
	engine := gin.Default()

	RegisterAPIv1(engine)
	RegisterSwagger(engine)

	// NEXA_HOST=127.0.0.1:18181 nexa serve
	err := engine.Run(config.Get().Host)
	if err != nil {
		slog.Error("HTTP Server Error", "err", err)
	}
}
