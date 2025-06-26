package main

import (
	"io"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/NexaAI/nexa-sdk/internal/config"
)

// TODO: fill description

// RootCmd creates the main Nexa CLI command with all subcommands.
// It sets up the command tree structure for model management,
// inference, and server operations.
func RootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "nexa",
	}

	// Model management commands
	rootCmd.AddCommand(pull())
	rootCmd.AddCommand(remove())
	rootCmd.AddCommand(clean())
	rootCmd.AddCommand(list())

	// Inference command for interactive chat
	rootCmd.AddCommand(infer())

	// Server command to run AI service
	rootCmd.AddCommand(serve())
	rootCmd.AddCommand(run())

	rootCmd.AddCommand(tool())
	return rootCmd
}

// main is the entry point that executes the root command.
func main() {
	if config.Get().Debug {
		log.SetOutput(os.Stderr)
	} else {
		log.SetOutput(io.Discard)
	}
	RootCmd().Execute()
}
