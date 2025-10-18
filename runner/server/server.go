// Copyright (c) 2025 Nexa AI
//
// LICENSE NOTICE - DUAL LICENSING:
// - NPU models and inference: CC-BY-NC 4.0 (NON-COMMERCIAL USE ONLY)
// - GPU/CPU models and inference: Apache 2.0 (FREE FOR ALL USE)

package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"golang.ngrok.com/ngrok/v2"

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
	if cfg.EnableHTTPS {
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
		fmt.Println(render.GetTheme().Info.Sprintf("Localhosting on http://%s/docs/ui", cfg.Host))
		err = engine.Run(cfg.Host)
	}

	if err != nil {
		fmt.Println(render.GetTheme().Error.Sprintf("HTTP/HTTPS Server Error: %v", err))
	}
}
