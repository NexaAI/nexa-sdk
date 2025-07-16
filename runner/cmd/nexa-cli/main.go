package main

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/NexaAI/nexa-sdk/internal/store"
)

// RootCmd creates the main Nexa CLI command with all subcommands.
// It sets up the command tree structure for model management,
// inference, and server operations.
func RootCmd() *cobra.Command {
	cobra.EnableCommandSorting = false

	rootCmd := &cobra.Command{
		Use: "nexa",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			applyLogLevel()
			return nil
		},
	}

	rootCmd.AddCommand(
		pull(), remove(), clean(), list(),
		infer(), genImage(),
		serve(), run(),
		version(),
	)

	return rootCmd
}

func normalizeModelName(name string) string {
	// support qwen3 -> nexaml/qwen3
	if !strings.Contains(name, "/") {
		return "nexaml/" + name
	}

	// support https://huggingface.co/Qwen/Qwen3-0.6B-GGUF -> Qwen/Qwen3-0.6B-GGUF
	if strings.HasPrefix(name, store.HF_ENDPOINT) {
		return strings.TrimPrefix(name, store.HF_ENDPOINT+"/")
	}

	return name
}

// main is the entry point that executes the root command.
func main() {
	if err := RootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
