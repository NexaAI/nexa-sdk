package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/NexaAI/nexa-sdk/runner/internal/model_hub"
	"github.com/NexaAI/nexa-sdk/runner/internal/render"
	"github.com/NexaAI/nexa-sdk/runner/internal/store"
	"github.com/NexaAI/nexa-sdk/runner/internal/types"
	"github.com/jedib0t/go-pretty/v6/table"
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
		// skip when no model cached
		models, err := store.Get().List()
		if err != nil {
			return err
		}
		if len(models) == 0 {
			return finishMigrate()
		}

		fmt.Print(render.GetTheme().Warning.Sprintf("A new version of Nexa CLI is detected. Start migrate now ? [Y/n] "))
		input, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		input = strings.TrimSpace(input)
		input = strings.ToLower(input)
		if input == "y" || input == "" {
			if err := startMigrate(); err != nil {
				return err
			}
			return finishMigrate()
		} else {
			fmt.Println()
			fmt.Println(render.GetTheme().Warning.Sprintf("Migration cancelled. You cannot use old models until migration is completed."))
			fmt.Println(render.GetTheme().Warning.Sprintf("You can run `nexa clean` to remove all old models."))
			fmt.Println()
			return errors.New("canceled")
		}
	}

	return nil
}

func finishMigrate() error {
	err := os.WriteFile(filepath.Join(store.Get().DataPath(), "last_version"), []byte(Version), 0o600)
	if err != nil {
		fmt.Println(render.GetTheme().Error.Sprintf("Failed to write last_version: %v", err))
		return err
	}
	fmt.Println(render.GetTheme().Success.Sprintf("Migration completed."))
	return nil
}

type MigrateResult struct {
	ModelName string
	Status    string
}

const (
	StatusSkip    = "skip"
	StatusSuccess = "success"
	StatusFailed  = "failed"
)

func startMigrate() error {

	s := store.Get()

	models, err := s.List()
	if err != nil {
		fmt.Println(render.GetTheme().Error.Sprintf("Failed to get model list: %v", err))
		return err
	}

	result := make([]MigrateResult, len(models))

	for i, model := range models {
		// check if model has update
		spin := render.NewSpinner("Checking model " + model.Name)
		spin.Start()
		files, hmf, err := model_hub.ModelInfo(context.Background(), model.Name)
		spin.Stop()

		if err != nil {
			fmt.Println(render.GetTheme().Error.Sprintf("Failed to get model info for %s: %v", model.Name, err))
			return err
		}

		if hmf == nil || model.MinSDKVersion == hmf.MinSDKVersion {
			result[i] = MigrateResult{ModelName: model.Name, Status: StatusSkip}
			continue
		}

		if !isValidVersion(hmf.MinSDKVersion) {
			fmt.Println(render.GetTheme().Error.Sprintf("Model %s requires NexaSDK CLI version %s or higher. Please upgrade your NexaSDK CLI.", model.Name, hmf.MinSDKVersion))
			return fmt.Errorf("model %s requires CLI version %s or higher", model.Name, hmf.MinSDKVersion)
		}

		// start migrate
		if err := s.Remove(model.Name); err != nil {
			fmt.Println(render.GetTheme().Error.Sprintf("Failed to delete model %s: %v", model.Name, err))
			return err
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
			result[i] = MigrateResult{ModelName: model.Name, Status: StatusSuccess}

			for err := range errCh {
				bar.Clear()
				fmt.Println(render.GetTheme().Error.Sprintf("Error: %s", err))
				result[i] = MigrateResult{ModelName: model.Name, Status: StatusFailed}
			}
		}
	}

	hasError := false
	tw := table.NewWriter()
	tw.SetOutputMirror(os.Stdout)
	tw.SetStyle(table.StyleLight)
	tw.AppendHeader(table.Row{"NAME", "STATUS"})
	for _, r := range result {
		switch r.Status {
		case StatusSkip:
			tw.AppendRow(table.Row{r.ModelName, render.GetTheme().Info.Sprintf("already up to date")})
		case StatusSuccess:
			tw.AppendRow(table.Row{r.ModelName, render.GetTheme().Success.Sprintf("success")})
		case StatusFailed:
			hasError = true
			tw.AppendRow(table.Row{r.ModelName, render.GetTheme().Warning.Sprintf("failed")})
		}
	}
	tw.Render()
	if hasError {
		fmt.Println(render.GetTheme().Warning.Sprintf("Some models failed to migrate. You need to download them manually."))
	}

	return nil
}
