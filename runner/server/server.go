package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"golang.ngrok.com/ngrok/v2"

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

	RegisterRoot(engine)
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
	} else if cfg.UseNgrok {
		slog.Info("Ngrok HTTPS enabled")

		// Create a custom agent with explicit authtoken
		agent, err := ngrok.NewAgent(
			ngrok.WithAuthtoken("30w3s7lTj2h3NMEBTyQ1sQ2pSGU_5iAXnXDWJS8VwbsQEcNws"),
		)
		if err != nil {
			slog.Error("Failed to create ngrok agent", "err", err)
			return
		}

		// Connect the agent
		if err := agent.Connect(context.Background()); err != nil {
			slog.Error("Failed to connect ngrok agent", "err", err)
			return
		}

		// Create listener using the agent
		listener, err := agent.Listen(context.Background())
		if err != nil {
			slog.Error("Failed to create ngrok tunnel", "err", err)
			return
		}

		slog.Info("API documentation available at", "url", listener.URL().String()+"/docs/ui")
		err = http.Serve(listener, engine)
		if err != nil {
			slog.Error("Failed to serve ngrok tunnel", "err", err)
			return
		}
	} else {
		slog.Info("Localhosting on http://" + cfg.Host + "/docs/ui")
		err = engine.Run(cfg.Host)
	}

	if err != nil {
		slog.Error("HTTP/HTTPS Server Error", "err", err)
	}
}
