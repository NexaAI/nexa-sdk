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
		panic(err)
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
