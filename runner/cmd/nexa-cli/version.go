package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = "Unknown"

func version() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "show nexasdk version",
	}

	versionCmd.Run = func(cmd *cobra.Command, args []string) {
		fmt.Println("Linux ARM64 Latest")
	}

	return versionCmd
}
