package main

import (
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
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

func setRuntimeEnv() {
	prependPath := func(envKey, newPath string) {
		old := os.Getenv(envKey)
		if old == "" {
			_ = os.Setenv(envKey, newPath)
		} else {
			sep := string(os.PathListSeparator) // ';' on Windows, ':' on *nix
			_ = os.Setenv(envKey, newPath+sep+old)
		}
	}

	backend := "yolov12"
	for _, arg := range os.Args[1:] {
		if strings.Contains(arg, "paddleocr") {
			backend = "paddleocr"
			break
		}
	}

	backend = filepath.Join(baseDir, backend)
	switch runtime.GOOS {
	case "windows":
		prependPath("PATH", backend)
	case "linux":
		prependPath("LD_LIBRARY_PATH", backend)
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

	cmd.Run()
	os.Exit(cmd.ProcessState.ExitCode())
}
