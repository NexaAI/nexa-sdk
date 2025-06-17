package main

import (
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/NexaAI/nexa-sdk/internal/store"
)

func pull() *cobra.Command {
	pullCmd := &cobra.Command{}
	pullCmd.Use = "pull"

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
