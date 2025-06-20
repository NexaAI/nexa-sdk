package server

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/NexaAI/nexa-sdk/internal/config"
)

func Serve() {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.Default()

	RegisterAPIV1(engine)

	// NEXA_HOST=127.0.0.1:18181 nexa serve
	err := engine.Run(config.Get().Host)
	if err != nil {
		fmt.Printf("HTTP Server Error: %s", err)
	}
}
