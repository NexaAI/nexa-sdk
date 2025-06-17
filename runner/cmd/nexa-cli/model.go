package main

import (
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/NexaAI/nexa-sdk/internal/store"
)

func pull() *cobra.Command {
	pullCmd := &cobra.Command{}
	pullCmd.Use = "pull"

	return pullCmd
}

func remove() *cobra.Command {
	removeCmd := &cobra.Command{}
	removeCmd.Use = "remove"

	removeCmd.Args = cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)

	removeCmd.RunE = func(cmd *cobra.Command, args []string) error {
		s := store.NewStore()
		return s.Remove(args[0])
	}

	return removeCmd
}

func clean() *cobra.Command {
	cleanCmd := &cobra.Command{}
	cleanCmd.Use = "clean"

	cleanCmd.RunE = func(cmd *cobra.Command, args []string) error {
		s := store.NewStore()
		return s.Clean()
	}

	return cleanCmd
}

func list() *cobra.Command {
	listCmd := &cobra.Command{}
	listCmd.Use = "list"

	listCmd.RunE = func(cmd *cobra.Command, args []string) error {
		s := store.NewStore()
		models, e := s.List()
		if e != nil {
			return e
		}

		tw := table.NewWriter()
		tw.SetOutputMirror(os.Stdout)
		tw.SetStyle(table.StyleLight)
		tw.AppendHeader(table.Row{"NAME", "SIZE"})
		for _, model := range models {
			tw.AppendRow(table.Row{model.Name, -1})
		}
		tw.Render()

		return nil
	}

	return listCmd
}
