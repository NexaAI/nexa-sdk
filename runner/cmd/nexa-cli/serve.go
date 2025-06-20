package main

import (
	"github.com/spf13/cobra"

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
		// Initialize SDK resources and prepare AI models for serving
		nexa_sdk.Init()

		// Start the HTTP server (blocks until shutdown signal received)
		// This serves REST API endpoints for model inference
		server.Serve()

		// Clean up SDK resources and free memory when server stops
		nexa_sdk.DeInit()
	}
	return serveCmd
}
