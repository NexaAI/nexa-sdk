package server

import (
	"log/slog"
	"os"

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

	cfg := config.Get()
	var err error

	// Determine whether to serve over HTTPS
	if cfg.EnableHTTPS {
		certFile := cfg.CertFile
		keyFile := cfg.KeyFile

		// Verify that certificate and key files exist
		if _, err := os.Stat(certFile); os.IsNotExist(err) {
			slog.Error("HTTPS Certificate file not found", "cert", certFile)
			return
		}
		if _, err := os.Stat(keyFile); os.IsNotExist(err) {
			slog.Error("HTTPS Key file not found", "key", keyFile)
			return
		}

		slog.Info("HTTPS enabled", "cert", certFile, "key", keyFile)
		// slog.Info("Localhosting on https://" + cfg.Host + "/docs/ui")
		err = engine.RunTLS(cfg.Host, certFile, keyFile)
	} else {
		slog.Info("Localhosting on http://" + cfg.Host + "/docs/ui")
		err = engine.Run(cfg.Host)
	}

	if err != nil {
		slog.Error("HTTP/HTTPS Server Error", "err", err)
	}
}
