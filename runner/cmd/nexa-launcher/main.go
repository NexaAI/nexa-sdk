package main

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"syscall"
)

var (
	filePath string
	libPath  string
	binPath  string
)

func initPath() {
	var err error
	filePath, err = os.Executable()
	if err != nil {
		panic(err)
	}

	libPath = path.Join(path.Dir(filePath), "lib")
	binPath = path.Join(path.Dir(filePath), "nexa-cli")
}

func setRuntimeEnv() {
	switch runtime.GOOS {
	case "windows":
		panic("unsupport os :" + runtime.GOOS)

	case "linux":
		rpath := os.Getenv("LD_LIBRARY_PATH")
		if rpath == "" {
			rpath = libPath
		} else {
			rpath = fmt.Sprintf("%s:%s", libPath, rpath)
		}
		os.Setenv("LD_LIBRARY_PATH", rpath)

	case "darwin":
		rpath := os.Getenv("DYLD_LIBRARY_PATH")
		if rpath == "" {
			rpath = libPath
		} else {
			rpath = fmt.Sprintf("%s:%s", libPath, rpath)
		}
		os.Setenv("DYLD_LIBRARY_PATH", rpath)

	default:
		panic("unsupport os :" + runtime.GOOS)
	}
}

func main() {
	initPath()
	setRuntimeEnv()
	os.Args[0] = binPath
	fmt.Printf("execute %s %v\n", binPath, os.Args)
	err := syscall.Exec(binPath, os.Args, os.Environ())
	if err != nil {
		panic(err)
	}
}
