package main

import "github.com/spf13/cobra"

// TODO: fill description
func root() *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "nexa",
	}

	rootCmd.AddCommand(pull())
	rootCmd.AddCommand(remove())
	rootCmd.AddCommand(clean())
	rootCmd.AddCommand(list())

	rootCmd.AddCommand(infer())

	rootCmd.AddCommand(serve())
	return rootCmd
}

func main() {
	root().Execute()
}
