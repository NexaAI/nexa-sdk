package main

import (
	"fmt"
	"log/slog"

	nexa_sdk "github.com/NexaAI/nexa-sdk/runner/nexa-sdk"
	"github.com/spf13/cobra"
)

var Version = "Unknown"

func version() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "show nexasdk version",
	}

	versionCmd.Run = func(cmd *cobra.Command, args []string) {
		fmt.Println("NexaSDK Bridge Version: " + nexa_sdk.Version())
		fmt.Println("NexaSDK CLI Version:    " + Version)
		fmt.Println()
		fmt.Println("Available Plugins:")

		nexa_sdk.Init()
		defer nexa_sdk.DeInit()

		pls, e := nexa_sdk.GetPluginList()
		if e != nil {
			slog.Error("GetPluginList error", "err", e)
		}
		for _, p := range pls.PluginIDs {
			fmt.Println("  - " + p)
		}
	}

	return versionCmd
}
