package main

import (
	"fmt"
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

	// LLM backend mapping - if any of these model names are found in args, use the corresponding backend
	fmt.Println("os.Args[1:]", os.Args[1:])
	llmBackendMap := map[string][]string{
		"qwen3-4B": {
			"qwen3-4B-npu",
			"jan-v1-4B-npu",
			"Qwen3-4B-Thinking-2507-npu",
			"Qwen3-4B-Instruct-2507-npu",
		},
		"llama3-3B-turbo": {
			"Llama3.2-3B-NPU-Turbo",
		},
		"llama3-3B": {
			"Llama3.2-3B-NPU",
		},
	}

	llmBackend := "qwen3" // default backend
	for backend, modelNames := range llmBackendMap {
		for _, arg := range os.Args[1:] {
			for _, modelName := range modelNames {
				if strings.Contains(arg, modelName) {
					llmBackend = backend
					fmt.Println("llmBackend", llmBackend)
					goto foundLLMBackend
				}
			}
		}
	}
foundLLMBackend:
	llmBackend = filepath.Join(baseDir, llmBackend)

	// CV backend mapping
	cvBackendMap := map[string][]string{
		"paddleocr": {
			"paddleocr-npu",
		},
	}

	cvBackend := "yolov12" // default backend
	for backend, modelNames := range cvBackendMap {
		for _, arg := range os.Args[1:] {
			for _, modelName := range modelNames {
				if strings.Contains(arg, modelName) {
					cvBackend = backend
					goto foundCVBackend
				}
			}
		}
	}
foundCVBackend:
	cvBackend = filepath.Join(baseDir, cvBackend)

	switch runtime.GOOS {
	case "windows":
		prependPath("PATH", llmBackend)
		prependPath("PATH", cvBackend)
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
