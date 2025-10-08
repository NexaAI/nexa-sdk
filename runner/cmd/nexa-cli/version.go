package main

import (
	"fmt"
	"log/slog"

	goversion "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
)

var Version string

func version() *cobra.Command {
	versionCmd := &cobra.Command{
		GroupID: "management",
		Use:     "version",
		Short:   "show nexasdk version",
	}

	versionCmd.Run = func(cmd *cobra.Command, args []string) {
		fmt.Println("Linux Arm64 Latest")
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
