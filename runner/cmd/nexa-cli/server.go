package main

import "github.com/spf13/cobra"

func serve() *cobra.Command {
	serveCmd := &cobra.Command{}
	serveCmd.Use = "serve"
	serveCmd.Short = "Run the Nexa AI Service"

	serveCmd.Run = func(cmd *cobra.Command, args []string) {

	}
	return serveCmd
}
