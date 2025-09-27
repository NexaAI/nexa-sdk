package main

import (
	"fmt"
	"log/slog"

	goversion "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"

	nexa_sdk "github.com/NexaAI/nexa-sdk/runner/nexa-sdk"
)

var Version string

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

func isValidVersion(minVersion string) bool {
	// community repo or dev version
	if minVersion == "" || Version == "" {
		return true
	}

	slog.Debug("check version", "minVersion", minVersion, "curVersion", Version)
	minV, err := goversion.NewVersion(minVersion)
	if err != nil {
		panic(err)
	}
	curV, err := goversion.NewVersion(Version)
	if err != nil {
		panic(err)
	}
	return curV.GreaterThanOrEqual(minV)
}
