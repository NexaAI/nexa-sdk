package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/NexaAI/nexa-sdk/runner/internal/config"
	nexa_sdk "github.com/NexaAI/nexa-sdk/runner/nexa-sdk"
	"github.com/NexaAI/nexa-sdk/runner/server"
)

// serve creates a command to start the Nexa AI service server.
// This command initializes the Nexa SDK, starts the HTTP server for AI services,
// and properly cleans up resources when the server shuts down.
// The server provides REST API endpoints for AI model inference and management.
// Usage: nexa serve
func serve() *cobra.Command {
	serveCmd := &cobra.Command{
		GroupID: "inference",
		Use:     "serve",
		Short:   "Run the Nexa AI Service",
	}

	serveCmd.Flags().SortFlags = false
	serveCmd.Flags().String("host", "127.0.0.1:18181", "Default server address (env: NEXA_HOST)")
	serveCmd.Flags().Int("keepalive", 300, "Keepalive seconds (env: NEXA_KEEPALIVE)")
	// HTTPS / TLS flags
	serveCmd.Flags().Bool("https", false, "Enable HTTPS/TLS (env: NEXA_ENABLEHTTPS)")
	serveCmd.Flags().String("certfile", "cert.pem", "TLS certificate file path (env: NEXA_CERTFILE)")
	serveCmd.Flags().String("keyfile", "key.pem", "TLS private key file path (env: NEXA_KEYFILE)")
	serveCmd.Flags().Bool("ngrok", false, "Use ngrok for public HTTPS tunnel (env: NEXA_NGROK)")

	viper.BindPFlag("host", serveCmd.Flags().Lookup("host"))
	viper.BindPFlag("keepalive", serveCmd.Flags().Lookup("keepalive"))
	viper.BindPFlag("enablehttps", serveCmd.Flags().Lookup("https"))
	viper.BindPFlag("certfile", serveCmd.Flags().Lookup("certfile"))
	viper.BindPFlag("keyfile", serveCmd.Flags().Lookup("keyfile"))
	viper.BindPFlag("ngrok", serveCmd.Flags().Lookup("ngrok"))

	serveCmd.Run = func(cmd *cobra.Command, args []string) {
		config.Refresh()

		checkDependency()
		nexa_sdk.Init()

		server.Serve()

		nexa_sdk.DeInit()
	}

	return serveCmd
}
