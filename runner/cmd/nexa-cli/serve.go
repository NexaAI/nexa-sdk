package main

import (
	"github.com/spf13/cobra"

	nexa_sdk "github.com/NexaAI/nexa-sdk/nexa-sdk"
	"github.com/NexaAI/nexa-sdk/server"
)

func serve() *cobra.Command {
	serveCmd := &cobra.Command{}
	serveCmd.Use = "serve"
	serveCmd.Short = "Run the Nexa AI Service"

	serveCmd.Run = func(cmd *cobra.Command, args []string) {
		nexa_sdk.Init()
		server.Serve()
		nexa_sdk.DeInit()
	}
	return serveCmd
}
