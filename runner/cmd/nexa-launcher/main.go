package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
)

var (
	baseDir string
	binPath string

	backends []string
	backend  string
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

	backend = filepath.Join(baseDir, "lib", backend)
	switch runtime.GOOS {
	case "windows":
		prependPath("PATH", backend)
	case "linux":
		prependPath("LD_LIBRARY_PATH", backend)
	case "darwin":
		prependPath("DYLD_LIBRARY_PATH", backend)
		if strings.Contains(backend, "mlx") {
			binPath = filepath.Join(filepath.Dir(binPath), "nexa-cli-mlx")
		}
	default:
		panic("unsupported OS: " + runtime.GOOS)
	}
}

func detectBackend() {
	var libName string
	switch runtime.GOOS {
	case "windows":
		libName = "nexa_bridge.dll"
	case "linux":
		libName = "libnexa_bridge.so"
	case "darwin":
		libName = "libnexa_bridge.dylib"
	default:
		panic("unsupported OS: " + runtime.GOOS)
	}

	// detect all backend
	libDir := filepath.Join(baseDir, "lib")
	dirs, err := os.ReadDir(libDir)
	if err != nil {
		panic(err)
	}
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}

		bridgePath := filepath.Join(libDir, dir.Name(), libName)
		_, err := os.Stat(bridgePath)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			panic(err)
		}
		backends = append(backends, dir.Name())
	}
	if len(backends) == 0 {
		panic("no backend found")
	}

	backend = backends[0]

	for i := range len(os.Args) {
		switch os.Args[i] {
		case "-b", "--backend":
			if i+1 >= len(os.Args) {
				fmt.Println("must specify <backend> after --backend")
				os.Exit(1)
			}
			backend = os.Args[i+1]
			if !slices.Contains(backends, backend) {
				fmt.Println(fmt.Sprint("backend not found, exist backends: ", backends))
				os.Exit(1)
			}
			os.Args = append(os.Args[:i], os.Args[i+2:]...)
			return
		}
	}

	// Special handling for MLX models
	if len(os.Args) > 1 {
		for _, arg := range os.Args[1:] {
			if strings.Contains(strings.ToLower(arg), "mlx") {
				if runtime.GOOS != "darwin" {
					fmt.Printf("ℹ️ MLX models work on Apple-silicon Macs. You're on %s—choose a GGUF model or try on an M-series Mac.\n", runtime.GOOS)
					os.Exit(1)
				}
				backend = "mlx"
			}
		}
	}
}

func main() {
	initPath()
	detectBackend()
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
