package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"

	"github.com/NexaAI/nexa-sdk/internal/store"
	"github.com/NexaAI/nexa-sdk/internal/types"
)

// pull creates a command to download and cache a model by name.
// Usage: nexa pull <model-name>
func pull() *cobra.Command {
	pullCmd := &cobra.Command{}
	pullCmd.Use = "pull <model-name>"

	pullCmd.Short = "Pull model from HuggingFace"
	pullCmd.Long = "Download and cache a model by name."

	pullCmd.Args = cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)
	_type := pullCmd.Flags().StringP("type", "t", types.ModelTypeLLM, "specify the model type, must be one of <llm|vlm|embed|rerank>")
	model := pullCmd.Flags().StringP("model", "m", "", "specify the main model file")
	tokenizer := pullCmd.Flags().StringP("tokenizer", "k", "", "specify the tokenizer file")
	extra := pullCmd.Flags().StringSliceP("extra-files", "e", nil, "specify extra files need download")
	all := pullCmd.Flags().BoolP("all", "a", false, "download all file even specify the model file")

	pullCmd.Run = func(cmd *cobra.Command, args []string) {
		s := store.NewStore()

		// TODO: replace with go-pretty
		pgCh, errCh := s.Pull(context.TODO(), args[0], store.PullOption{
			ModelType: types.ModelType(*_type),
			Model:     *model,
			Tokenizer: *tokenizer,
			Extra:     *extra,
			ALl:       *all,
		})
		bar := progressbar.DefaultBytes(-1, "downloading")
		for pg := range pgCh {
			if pg.CurrentSize != bar.GetMax64() {
				bar.Reset()
				bar.Describe("download " + pg.CurrentName)
				bar.ChangeMax64(pg.CurrentSize)
			}
			bar.Set64(pg.CurrentDownloaded)
		}
		bar.Exit()

		for err := range errCh {
			bar.Clear()
			fmt.Println(text.FgRed.Sprintf("Error: %s", err))
		}
	}

	return pullCmd
}

// remove creates a command to delete a cached model by name.
// Usage: nexa remove <model-name>
func remove() *cobra.Command {
	removeCmd := &cobra.Command{
		Use:   "remove <model-name>",
		Short: "Remove cached model",
		Long:  "Delete a cached model by name. This will remove the model files from the local cache.",
	}

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
	cleanCmd := &cobra.Command{
		Use:   "clean",
		Short: "remove all cached models",
		Long:  "Remove all cached models and free up storage. This will delete all model files from the local cache.",
	}

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
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all cached models",
		Long:  "Display all cached models in a formatted table, showing model names, types, and sizes.",
	}

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
		tw.AppendHeader(table.Row{"NAME", "TYPE", "SIZE"})
		for _, model := range models {
			tw.AppendRow(table.Row{model.Name, model.ModelType, model.Size})
		}
		tw.Render()
	}

	return listCmd
}
