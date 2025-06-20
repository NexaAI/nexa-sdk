package main

import "github.com/spf13/cobra"

// TODO: fill description

// root creates the main Nexa CLI command with all subcommands.
// It sets up the command tree structure for model management,
// inference, and server operations.
func root() *cobra.Command {
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
	return rootCmd
}

// main is the entry point that executes the root command.
func main() {
	root().Execute()
}
