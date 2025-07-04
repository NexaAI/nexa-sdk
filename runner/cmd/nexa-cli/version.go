package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = "unknown"

func version() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "show nexasdk version",
	}

	versionCmd.Run = func(cmd *cobra.Command, args []string) {
		fmt.Println("NexaSDK Version:     ")
		fmt.Println("NexaSDK Cli Version: " + Version)
	}

	return versionCmd
}
