package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
)

var (
	filePath string
	libPath  string
	binPath  string
)

func initPath() {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	filePath = exe
	baseDir := filepath.Dir(filePath)

	libPath = filepath.Join(baseDir, "lib")
	binPath = filepath.Join(baseDir, "nexa-cli")
}

func prependPath(envKey, newPath string) {
	old := os.Getenv(envKey)
	if old == "" {
		_ = os.Setenv(envKey, newPath)
	} else {
		sep := string(os.PathListSeparator) // ';' on Windows, ':' on *nix
		_ = os.Setenv(envKey, newPath+sep+old)
	}
}

func setRuntimeEnv() {
	switch runtime.GOOS {
	case "windows":
		prependPath("PATH", libPath)
	case "linux":
		prependPath("LD_LIBRARY_PATH", libPath)
	case "darwin":
		prependPath("DYLD_LIBRARY_PATH", libPath)
	default:
		panic("unsupported OS: " + runtime.GOOS)
	}
}

func main() {
	initPath()
	setRuntimeEnv()

	cmd := exec.Command(binPath, os.Args[1:]...)
	cmd.Env = os.Environ()
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	cSignal := make(chan os.Signal, 1)
	signal.Notify(cSignal, os.Interrupt)
	go func() {
		for range cSignal {
		}
	}()

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "run %s failed: %v\n", binPath, err)
		os.Exit(1)
	}
}
