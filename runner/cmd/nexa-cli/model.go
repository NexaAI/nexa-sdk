package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/dustin/go-humanize"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/schollz/progressbar/v3"
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
		name := normalizeModelName(args[0])

		s := store.Get()

		if _, err := s.GetManifest(name); err == nil {
			fmt.Println(text.FgBlue.Sprint("Already downloaded, if you want to use another quant, please manual remove first"))
			return
		}

		// download manifest
		spin := spinner.New(
			spinner.CharSets[39],
			100*time.Millisecond,
			spinner.WithSuffix("download manifest from: "+name),
		)
		spin.Start()
		files, err := s.HFModelInfo(context.TODO(), name)
		spin.Stop()
		if err != nil {
			fmt.Println(text.FgRed.Sprintf("Get manifest from huggingface error: %s", err))
			return
		}

		manifest, err := chooseFiles(name, files)
		if err != nil {
			return
		}

		// TODO: replace with go-pretty
		pgCh, errCh := s.Pull(context.TODO(), manifest)
		bar := progressbar.NewOptions64(
			manifest.Size,
			progressbar.OptionSetDescription("downloading"),
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionShowBytes(true),
			progressbar.OptionShowTotalBytes(true),
			progressbar.OptionSetWidth(10),
			progressbar.OptionThrottle(65*time.Millisecond),
			progressbar.OptionShowCount(),
			progressbar.OptionOnCompletion(func() {
				fmt.Fprint(os.Stderr, "\n")
			}),
			progressbar.OptionSpinnerType(14),
			progressbar.OptionFullWidth(),
			progressbar.OptionSetRenderBlankState(true),
			progressbar.OptionUseANSICodes(true),
		)

		for pg := range pgCh {
			bar.Set64(pg.TotalDownloaded)
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
		name := normalizeModelName(args[0])

		s := store.Get()
		e := s.Remove(name)
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
		s := store.Get()
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
		s := store.Get()
		models, e := s.List()
		if e != nil {
			fmt.Println(e)
			return
		}

		// Create formatted table output
		tw := table.NewWriter()
		tw.SetOutputMirror(os.Stdout)
		tw.SetStyle(table.StyleLight)
		tw.AppendHeader(table.Row{"NAME", "QUANT", "SIZE"})
		for _, model := range models {
			tw.AppendRow(table.Row{model.Name, model.Quant, humanize.IBytes(uint64(model.Size))})
		}
		tw.Render()
	}

	return listCmd
}
