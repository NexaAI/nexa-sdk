package main

import (
	"github.com/spf13/cobra"
)

func server() *cobra.Command {
	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "Run the Nexa AI Service",
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
	return serverCmd
}

func root() *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "nexa",
	}
	rootCmd.AddCommand(server())
	rootCmd.AddCommand(infer())
	return rootCmd
}

func main() {
	root().Execute()
}
