package server

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/NexaAI/nexa-sdk/runner/internal/config"
	"github.com/NexaAI/nexa-sdk/runner/internal/render"
	"github.com/NexaAI/nexa-sdk/runner/server/service"
)

// @Title		Nexa AI Server
// @Version	0.0.0
// @BasePath	/v1
func Serve() {
	service.Init()
	defer service.DeInit()

	gin.SetMode(gin.ReleaseMode)
	engine := gin.Default()

	RegisterRoot(engine)
	RegisterAPIv1(engine)
	RegisterSwagger(engine)

	cfg := config.Get()
	var err error

	// Determine whether to serve over HTTPS
	if cfg.HTTPS {
		certFile := cfg.CertFile
		keyFile := cfg.KeyFile

		// Verify that certificate and key files exist
		if _, err := os.Stat(certFile); os.IsNotExist(err) {
			fmt.Println(render.GetTheme().Error.Sprintf("HTTPS Certificate file not found: %s", certFile))
			return
		}
		if _, err := os.Stat(keyFile); os.IsNotExist(err) {
			fmt.Println(render.GetTheme().Error.Sprintf("HTTPS Key file not found: %s", keyFile))
			return
		}

		fmt.Println(render.GetTheme().Info.Sprintf("HTTPS enabled: cert=%s key=%s", certFile, keyFile))
		// fmt.Println(render.GetTheme().Info.Sprintf("Localhosting on https://%s/docs/ui", cfg.Host))
		err = engine.RunTLS(cfg.Host, certFile, keyFile)
	} else {
		fmt.Println(render.GetTheme().Info.Sprintf("Localhosting on http://%s/docs/ui", cfg.Host))
		err = engine.Run(cfg.Host)
	}

	if err != nil {
		fmt.Println(render.GetTheme().Error.Sprintf("HTTP/HTTPS Server Error: %v", err))
	}
}
