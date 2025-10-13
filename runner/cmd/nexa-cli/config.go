package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/NexaAI/nexa-sdk/runner/internal/render"
	"github.com/NexaAI/nexa-sdk/runner/internal/store"
)

func _config() *cobra.Command {
	cmd := &cobra.Command{
		GroupID: "management",
		Use:     "config",
		Short:   "Manage Nexa CLI configuration",
		Long:    "Commands to manage Nexa CLI configuration, including setting and getting configuration values.",
	}

	cmd.AddCommand(
		configGetCmd(),
		configSetCmd(),
		configListCmd(),
	)

	return cmd
}

func configGetCmd() *cobra.Command {
	license := &cobra.Command{
		Use: "license",
		Run: func(cmd *cobra.Command, args []string) {
			value, err := store.Get().ConfigGet("license")
			if err != nil {
				fmt.Println(render.GetTheme().Error.Sprintf("Failed to get configuration: %s", err))
				os.Exit(1)
			}
			fmt.Println(render.GetTheme().Info.Sprintf("%s", value))
		},
	}

	getCmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Long:  "Retrieve the value of a specific configuration key.",
	}
	getCmd.AddCommand(license)
	return getCmd
}

func configSetCmd() *cobra.Command {
	license := &cobra.Command{
		Use:  "license",
		Args: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			s := store.Get()
			key := "license"
			value := args[0]

			if err := s.ConfigSet(key, value); err != nil {
				fmt.Println(render.GetTheme().Error.Sprintf("Failed to set configuration: %s", err))
				os.Exit(1)
			}
		},
	}

	setCmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Long:  "Set a specific configuration key to a new value.",
	}
	setCmd.AddCommand(license)
	return setCmd
}

func configListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all configuration values",
		Long:  "Display all configuration keys and their corresponding values.",
		Run: func(cmd *cobra.Command, args []string) {
			s := store.Get()

			for _, key := range []string{"license"} {
				value, _ := s.ConfigGet(key)
				fmt.Println(render.GetTheme().Info.Sprintf("%s: %s", key, value))
			}
		},
	}
}
