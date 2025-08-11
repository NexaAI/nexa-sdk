package nexa_sdk

import (
	"os"
	"testing"
)

// TestMain sets up the test environment and runs all tests
// It initializes the SDK, creates an LLM instance, runs tests, then cleans up
func TestMain(m *testing.M) {
	// slog.SetLogLoggerLevel(slog.LevelDebug)

	Init()
	defer DeInit()

	// initLLM()
	// defer deinitLLM()

	// initVLM()
	// defer deinitVLM()

	// initTTS()
	// defer deinitTTS()

	// initASR()
	// defer deinitASR()

	// initEmbedder()
	// defer deinitEmbedder()

	// initCV()
	// defer deinitCV()

	initReranker()
	defer deinitReranker()

	code := m.Run()
	os.Exit(code)
}
