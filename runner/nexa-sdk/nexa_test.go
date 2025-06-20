package nexa_sdk

import (
	"os"
	"path"
	"testing"
)

var (
	nexaPath string
)

func TestMain(m *testing.M) {
	cache, _ := os.UserCacheDir()
	nexaPath = path.Join(cache, "nexa")

	Init()
	initLLM()

	code := m.Run()

	deinitLLM()
	DeInit()
	os.Exit(code)
}
