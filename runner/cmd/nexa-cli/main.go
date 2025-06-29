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
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if config.Get().Debug {
				log.SetOutput(os.Stderr)
			} else {
				log.SetOutput(io.Discard)
			}
			return nil
		},
	}

	rootCmd.AddCommand(
		pull(), remove(), clean(), list(),
		infer(),
		serve(), run(),
		tool(),
	)

	return rootCmd
}

// main is the entry point that executes the root command.
func main() {
	if err := RootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
