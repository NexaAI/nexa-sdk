package main

import (
	"fmt"

	nexa_sdk "github.com/NexaAI/nexa-sdk/nexa-sdk"
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
	}

	return versionCmd
}
