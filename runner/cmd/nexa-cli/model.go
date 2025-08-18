package main

import (
	"context"
	"fmt"
	"os"

	"github.com/dustin/go-humanize"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/NexaAI/nexa-sdk/runner/internal/render"
	"github.com/NexaAI/nexa-sdk/runner/internal/store"
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

		mf, err := s.GetManifest(name)
		if err == nil {
			downloaded := true
			for _, f := range mf.ModelFile {
				if !f.Downloaded {
					downloaded = false
					break
				}
			}

			if downloaded {
				fmt.Println(render.GetTheme().Info.Sprint("Already downloaded all quant"))
				return
			}
		}

		spin := render.NewSpinner("download manifest from: " + name)
		spin.Start()
		files, err := s.HFModelInfo(context.TODO(), name)
		spin.Stop()
		if err != nil {
			fmt.Println(render.GetTheme().Error.Sprintf("Get manifest from huggingface error: %s", err))
			return
		}

		if mf != nil {
			newManifest, err := chooseQuantFiles(*mf)
			if err != nil {
				return
			}
			// TODO: replace with go-pretty
			pgCh, errCh := s.PullExtraQuant(context.TODO(), *mf, *newManifest)
			bar := render.NewProgressBar(newManifest.GetSize()-mf.GetSize(), "downloading")

			for pg := range pgCh {
				bar.Set(pg.TotalDownloaded)
			}
			bar.Exit()

			for err := range errCh {
				bar.Clear()
				fmt.Println(render.GetTheme().Error.Sprintf("Error: %s", err))
			}
		} else {
			modelType, err := chooseModelTypeByName(name)
			if err != nil {
				return
			}

			manifest, err := chooseFiles(name, files)
			if err != nil {
				return
			}
			manifest.ModelType = modelType

			// TODO: replace with go-pretty
			pgCh, errCh := s.Pull(context.TODO(), manifest)
			bar := render.NewProgressBar(manifest.GetSize(), "downloading")

			for pg := range pgCh {
				bar.Set(pg.TotalDownloaded)
			}
			bar.Exit()

			for err := range errCh {
				bar.Clear()
				fmt.Println(render.GetTheme().Error.Sprintf("Error: %s", err))
			}
		}
	}

	return pullCmd
}

// remove creates a command to delete a cached model by name.
// Usage: nexa remove <model-name>
func remove() *cobra.Command {
	removeCmd := &cobra.Command{
		Use:     "remove <model-name>",
		Aliases: []string{"rm"},
		Short:   "Remove cached model",
		Long:    "Delete a cached model by name. This will remove the model files from the local cache.",
	}

	removeCmd.Args = cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)

	removeCmd.Run = func(cmd *cobra.Command, args []string) {
		name := normalizeModelName(args[0])

		s := store.Get()
		e := s.Remove(name)
		if e != nil {
			fmt.Println(e)
		} else {
			fmt.Printf("✔  Removed %s\n", name)
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
		c := s.Clean()
		fmt.Printf("✔  Removed %d models\n", c)
	}

	return cleanCmd
}

// list creates a command to display all cached models in a formatted table.
// Shows model names and their storage sizes.
// Usage: nexa list
func list() *cobra.Command {
	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all cached models",
		Long:    "Display all cached models in a formatted table, showing model names, types, and sizes.",
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
		tw.AppendHeader(table.Row{"NAME", "TYPE", "PLUGIN", "SIZE"})
		for _, model := range models {
			tw.AppendRow(table.Row{model.Name, model.ModelType, model.PluginId, humanize.IBytes(uint64(model.GetSize()))})
		}
		tw.Render()
	}

	return listCmd
}
