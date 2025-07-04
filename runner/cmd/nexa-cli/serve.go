package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/NexaAI/nexa-sdk/internal/config"
	nexa_sdk "github.com/NexaAI/nexa-sdk/nexa-sdk"
	"github.com/NexaAI/nexa-sdk/server"
)

// serve creates a command to start the Nexa AI service server.
// This command initializes the Nexa SDK, starts the HTTP server for AI services,
// and properly cleans up resources when the server shuts down.
// The server provides REST API endpoints for AI model inference and management.
// Usage: nexa serve
func serve() *cobra.Command {
	serveCmd := &cobra.Command{}
	serveCmd.Use = "serve"
	serveCmd.Short = "Run the Nexa AI Service"

	serveCmd.Run = func(cmd *cobra.Command, args []string) {
		fmt.Printf("Start Nexa AI Server on http://%s\n", config.Get().Host)

		// Initialize SDK resources and prepare AI models for serving
		nexa_sdk.Init()

		// Start the HTTP server (blocks until shutdown signal received)
		// This serves REST API endpoints for model inference
		server.Serve()

		// Clean up SDK resources and free memory when server stops
		nexa_sdk.DeInit()
	}

	serveCmd.Flags().SortFlags = false
	serveCmd.Flags().String("host", "127.0.0.1:18181", "Default server address (env: NEXA_HOST)")
	serveCmd.Flags().Int("keepalive", 300, "Keepalive seconds (env: NEXA_KEEPALIVE)")

	viper.BindPFlag("host", serveCmd.Flags().Lookup("host"))
	viper.BindPFlag("keepalive", serveCmd.Flags().Lookup("keepalive"))

	return serveCmd
}
