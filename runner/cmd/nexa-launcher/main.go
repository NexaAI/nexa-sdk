package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	baseDir  string
	backends []string
	backend  string
)

func initPath() {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	baseDir = filepath.Dir(exe)
	if filepath.Base(baseDir) == "bin" {
		baseDir = filepath.Dir(baseDir)
	}

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
	default:
		panic("unsupported OS: " + runtime.GOOS)
	}
}

func detectBackend() {
	for i := range len(os.Args) {
		switch os.Args[i] {
		case "-b", "--backend":
			if i+1 >= len(os.Args) {
				panic("must specify <backend> after --backend")
			}
			backend = os.Args[i+1]
			os.Args = append(os.Args[:i], os.Args[i+2:]...)
			return
		}
	}

	// detect mlx
	backend = backends[0]
	if len(os.Args) > 1 {
		for _, arg := range os.Args[1:] {
			if strings.Contains(strings.ToLower(arg), "mlx") {
				backend = "mlx"
			}
		}
	}
}

func main() {
	initPath()
	detectBackend()
	setRuntimeEnv()

	binPath := filepath.Join(baseDir, "nexa-cli")
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
