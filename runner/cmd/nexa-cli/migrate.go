package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/NexaAI/nexa-sdk/runner/internal/model_hub"
	"github.com/NexaAI/nexa-sdk/runner/internal/render"
	"github.com/NexaAI/nexa-sdk/runner/internal/store"
	"github.com/NexaAI/nexa-sdk/runner/internal/types"
)

func checkMigrate() error {
	currentVersion := Version
	v, err := os.ReadFile(filepath.Join(store.Get().DataPath(), "last_version"))
	lastVersion := string(v)
	slog.Info("checkMigrate", "current", currentVersion, "version", lastVersion, "err", err)
	if err != nil && !os.IsNotExist(err) {
		fmt.Println(render.GetTheme().Error.Sprintf("Failed to read last_version: %v", err))
		return err
	}

	if currentVersion != lastVersion {
		fmt.Println(render.GetTheme().Warning.Sprintf(`
A new version of Nexa CLI is detected. Please run "nexa migrate" to migrate your models.
Use "nexa migrate --help" to see more options.

		`))
		return errors.New("need migrate")
	}

	return nil
}

func migrate() *cobra.Command {
	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate models from previous versions",
	}

	skip := migrateCmd.Flags().Bool("skip", false, "force skip migration")
	deleteOnly := migrateCmd.Flags().Bool("delete-only", false, "delete outdated models instead of migrating")

	migrateCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if *skip {
			fmt.Println(render.GetTheme().Warning.Sprintf("Migration skipped. Please make sure your models are compatible with the current version."))
			return os.WriteFile(filepath.Join(store.Get().DataPath(), "last_version"), []byte(Version), 0o600)
		}

		s := store.Get()

		models, err := s.List()
		if err != nil {
			return err
		}

		for _, model := range models {
			// check if model has update
			spin := render.NewSpinner("Checking model " + model.Name)
			spin.Start()
			files, hmf, err := model_hub.ModelInfo(context.Background(), model.Name)
			spin.Stop()

			if err != nil {
				fmt.Println(render.GetTheme().Error.Sprintf("Failed to get model info for %s: %v", model.Name, err))
				return err
			}

			if model.MinSDKVersion == hmf.MinSDKVersion {
				continue
			}

			if !isValidVersion(hmf.MinSDKVersion) {
				fmt.Println(render.GetTheme().Error.Sprintf("Model %s requires NexaSDK CLI version %s or higher. Please upgrade your NexaSDK CLI.", model.Name, hmf.MinSDKVersion))
				return nil
			}

			// start migrate
			if err := s.Remove(model.Name); err != nil {
				fmt.Println(render.GetTheme().Error.Sprintf("Failed to delete model %s: %v", model.Name, err))
				return err
			}

			if *deleteOnly {
				continue
			}

			fmt.Println(render.GetTheme().Info.Sprintf("Start pull model %s", model.Name))

			// TODO: support multi quant
			if false {
				// newManifest, err := chooseQuantFiles(*mf)
				// if err != nil {
				// 	return
				// }
				// pgCh, errCh := s.PullExtraQuant(context.TODO(), *mf, *newManifest)
				// bar := render.NewProgressBar(newManifest.GetSize()-mf.GetSize(), "downloading")
				//
				// for pg := range pgCh {
				// 	bar.Set(pg.TotalDownloaded)
				// }
				// bar.Exit()
				//
				// for err := range errCh {
				// 	bar.Clear()
				// 	fmt.Println(render.GetTheme().Error.Sprintf("Error: %s", err))
				// }
			} else {
				var manifest types.ModelManifest

				manifest.ModelName = hmf.ModelName
				manifest.PluginId = hmf.PluginId
				manifest.ModelType = hmf.ModelType
				manifest.MinSDKVersion = hmf.MinSDKVersion

				if manifest.ModelName == "" {
					manifest.ModelName = model.ModelName
				}
				if manifest.PluginId == "" {
					manifest.PluginId = model.PluginId
				}
				if manifest.ModelType == "" {
					manifest.ModelType = model.ModelType
				}

				err := chooseFiles(model.Name, files, &manifest)
				if err != nil {
					fmt.Println(render.GetTheme().Error.Sprintf("Error: %s", err))
					return err
				}

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

		return os.WriteFile(filepath.Join(store.Get().DataPath(), "last_version"), []byte(Version), 0o600)
	}
	return migrateCmd
}
