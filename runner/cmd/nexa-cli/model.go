package main

import (
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/NexaAI/nexa-sdk/internal/store"
)

// pull creates a command to download and cache a model by name.
// Usage: nexa pull <model-name>
func pull() *cobra.Command {
	pullCmd := &cobra.Command{}
	pullCmd.Use = "pull <model-name>"

	pullCmd.Short = "Pull model from HuggingFace"
	pullCmd.Long = "Download and cache a model by name."

	pullCmd.Args = cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)

	pullCmd.Run = func(cmd *cobra.Command, args []string) {
		s := store.NewStore()
		e := s.Pull(args[0])
		if e != nil {
			fmt.Println(e)
		}
	}

	return pullCmd
}

// remove creates a command to delete a cached model by name.
// Usage: nexa remove <model-name>
func remove() *cobra.Command {
	removeCmd := &cobra.Command{}
	removeCmd.Use = "remove"

	removeCmd.Args = cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)

	removeCmd.Run = func(cmd *cobra.Command, args []string) {
		s := store.NewStore()
		e := s.Remove(args[0])
		if e != nil {
			fmt.Println(e)
		}
	}

	return removeCmd
}

// clean creates a command to remove all cached models and free up storage.
// Usage: nexa clean
func clean() *cobra.Command {
	cleanCmd := &cobra.Command{}
	cleanCmd.Use = "clean"

	cleanCmd.Run = func(cmd *cobra.Command, args []string) {
		s := store.NewStore()
		e := s.Clean()
		if e != nil {
			fmt.Println(e)
		}
	}

	return cleanCmd
}

// list creates a command to display all cached models in a formatted table.
// Shows model names and their storage sizes.
// Usage: nexa list
func list() *cobra.Command {
	listCmd := &cobra.Command{}
	listCmd.Use = "list"

	listCmd.Run = func(cmd *cobra.Command, args []string) {
		s := store.NewStore()
		models, e := s.List()
		if e != nil {
			fmt.Println(e)
			return
		}

		// Create formatted table output
		tw := table.NewWriter()
		tw.SetOutputMirror(os.Stdout)
		tw.SetStyle(table.StyleLight)
		tw.AppendHeader(table.Row{"NAME", "SIZE"})
		for _, model := range models {
			tw.AppendRow(table.Row{model.Name, model.Size})
		}
		tw.Render()
	}

	return listCmd
}
