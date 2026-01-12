// Copyright 2024-2026 Nexa AI, Inc.
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
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
)

var (
	baseDir string
	binPath string
)

func initPath() {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	realExe, err := filepath.EvalSymlinks(exe)
	if err != nil {
		realExe = exe
	}
	baseDir = filepath.Dir(realExe)
	binPath = filepath.Join(baseDir, "nexa-cli")
	if filepath.Base(baseDir) == "bin" {
		baseDir = filepath.Dir(baseDir)
	}
}

func main() {
	initPath()

	cmd := exec.Command(binPath, os.Args[1:]...)
	cmd.Env = os.Environ()
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	cSignal := make(chan os.Signal, 1)
	signal.Notify(cSignal, os.Interrupt)
	go func() {
		for range cSignal {
		}
	}()

	cmd.Run()
	os.Exit(cmd.ProcessState.ExitCode())
}
