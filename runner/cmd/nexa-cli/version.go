// Copyright 2024-2025 Nexa AI, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
		GroupID: "management",
		Use:     "version",
		Short:   "show nexasdk version",
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
