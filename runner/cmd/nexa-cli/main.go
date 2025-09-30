package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/NexaAI/nexa-sdk/runner/internal/config"
	"github.com/NexaAI/nexa-sdk/runner/internal/model_hub"
	"github.com/NexaAI/nexa-sdk/runner/internal/render"
	"github.com/NexaAI/nexa-sdk/runner/internal/store"
)

var (
	skipUpdate  bool
	skipMigrate bool
)

// RootCmd creates the main Nexa CLI command with all subcommands.
// It sets up the command tree structure for model management,
// inference, and server operations.
func RootCmd() *cobra.Command {
	cobra.EnableCommandSorting = false

	rootCmd := &cobra.Command{
		Use: "nexa",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// log
			applyLogLevel()

			subCmd := cmd.CalledAs()

			// force check migrate
			if !skipMigrate && slices.Contains([]string{"infer", "fc", "functioncall", "serve", "run"}, subCmd) {
				if err := checkMigrate(); err != nil {
					fmt.Println(render.GetTheme().Error.Sprintf("Migrate error: %s", err))
					os.Exit(1)
				}
			}

			// skip update check
			if !skipUpdate {
				if !slices.Contains([]string{"version", "update", "migrate"}, subCmd) {
					go checkForUpdate(false)
				}
				notifyUpdate()
			}

			// license
			license, err := store.Get().ConfigGet("license")
			if err != nil || license == "" {
				slog.Warn("license is not set", "err", err)
			} else if err := os.Setenv("NEXA_TOKEN", license); err != nil {
				panic(err)
			}

			return nil
		},
	}
	rootCmd.PersistentFlags().BoolVarP(&skipUpdate, "skip-update", "", false, "Skip checking for updates")
	rootCmd.PersistentFlags().BoolVarP(&skipMigrate, "skip-migrate", "", false, "Skip checking for model migrations")

	rootCmd.AddGroup(
		&cobra.Group{ID: "model", Title: "Model Commands"},
		&cobra.Group{ID: "inference", Title: "Inference Commands"},
		&cobra.Group{ID: "management", Title: "Management Commands"},
	)

	rootCmd.AddCommand(
		pull(), remove(), clean(), list(),
		infer(), functionCall(),
		serve(), run(),
		_config(),
		version(), update(),
	)

	return rootCmd
}

func checkDependency() {
	if _, err := exec.LookPath("sox"); err != nil {
		fmt.Println(render.GetTheme().Warning.Sprintf("Sox is not installed, some features may not work. Try:"))
		switch runtime.GOOS {
		case "darwin":
			fmt.Println(render.GetTheme().Warning.Sprintf("  brew install sox"))
		case "linux":
			fmt.Println(render.GetTheme().Warning.Sprintf("  sudo apt install sox       # Debian/Ubuntu"))
			fmt.Println(render.GetTheme().Warning.Sprintf("  sudo yum install sox       # RHEL/CentOS/Fedora"))
		case "windows":
			fmt.Println(render.GetTheme().Warning.Sprintf("  winget install --id=ChrisBagwell.SoX -e"))
			fmt.Println(render.GetTheme().Warning.Sprintf("Then restart your terminal to make sure sox is in PATH"))
		default:
			fmt.Println(render.GetTheme().Warning.Sprintf("Please install it manually for your OS: %s\n", runtime.GOOS))
		}
	}

	if _, err := exec.LookPath("ffmpeg"); err != nil {
		fmt.Println(render.GetTheme().Warning.Sprintf("FFmpeg is not installed, some features may not work. Try:"))
		switch runtime.GOOS {
		case "darwin":
			fmt.Println(render.GetTheme().Warning.Sprintf("  brew install ffmpeg"))
		case "linux":
			fmt.Println(render.GetTheme().Warning.Sprintf("  sudo apt install ffmpeg    # Debian/Ubuntu"))
			fmt.Println(render.GetTheme().Warning.Sprintf("  sudo yum install ffmpeg    # RHEL/CentOS/Fedora"))
		case "windows":
			fmt.Println(render.GetTheme().Warning.Sprintf("  winget install --id=BtbN.FFmpeg.GPL -e"))
			fmt.Println(render.GetTheme().Warning.Sprintf("Then restart your terminal to make sure ffmpeg is in PATH"))
		default:
			fmt.Println(render.GetTheme().Warning.Sprintf("Please install it manually for your OS: %s\n", runtime.GOOS))
		}
	}
}

func normalizeModelName(name string) string {
	// support shortcuts
	if actualName, exists := config.GetModelMapping(name); exists {
		return actualName
	}

	// support qwen3 -> NexaAI/qwen3
	if !strings.Contains(name, "/") {
		return "NexaAI/" + name
	}

	// support https://huggingface.co/Qwen/Qwen3-0.6B-GGUF -> Qwen/Qwen3-0.6B-GGUF
	if strings.HasPrefix(name, model_hub.HF_ENDPOINT) {
		return strings.TrimPrefix(name, model_hub.HF_ENDPOINT+"/")
	}

	return name
}

// main is the entry point that executes the root command.
func main() {
	if err := RootCmd().Execute(); err != nil {
		slog.Error("nexa-cli failed", "err", err)
		os.Exit(1)
	}
}
