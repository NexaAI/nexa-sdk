package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
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
		s := store.NewStore()

		// make nexaml repo as default
		if !strings.Contains(args[0], "/") {
			args[0] += "nexaml/"
		}

		// download manifest
		spin := spinner.New(spinner.CharSets[39], 100*time.Millisecond, spinner.WithSuffix("download manifest..."))
		spin.Start()
		manifest, files, err := s.HFRepoFiles(context.TODO(), args[0])
		spin.Stop()
		if err != nil {
			fmt.Println(text.FgRed.Sprintf("Get manifest from huggingface error: %s", err))
			return
		}

		opt := store.PullOption{}
		if manifest != nil {
			// use preset manifest
			opt.ModelType = manifest.ModelType
			opt.ModelFile = manifest.ModelFile
			opt.TokenizerFile = manifest.TokenizerFile
			opt.ExtraFiles = manifest.ExtraFiles
		} else {
			// interactive choose
			var err error
			opt.ModelType, opt.ModelFile, opt.TokenizerFile, opt.ExtraFiles, err = chooseFiles(files)
			if err != nil {
				return
			}
		}

		// TODO: replace with go-pretty
		pgCh, errCh := s.Pull(context.TODO(), args[0], opt)
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
	removeCmd := &cobra.Command{}
	removeCmd.Use = "remove"

	removeCmd.Args = cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)

	removeCmd.Run = func(cmd *cobra.Command, args []string) {
		s := store.NewStore()
		// make nexaml repo as default
		if !strings.Contains(args[0], "/") {
			args[0] += "nexaml/"
		}

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
		tw.AppendHeader(table.Row{"NAME", "TYPE", "SIZE"})
		for _, model := range models {
			tw.AppendRow(table.Row{model.Name, model.ModelType, model.Size})
		}
		tw.Render()
	}

	return listCmd
}
