package nexa_sdk

import (
	"os"
	"path"
	"testing"
)

var (
	// nexaPath holds the base directory for Nexa cache and models
	nexaPath string
)

// TestMain sets up the test environment and runs all tests
// It initializes the SDK, creates an LLM instance, runs tests, then cleans up
func TestMain(m *testing.M) {
	// Get user cache directory and set Nexa path
	cache, _ := os.UserCacheDir()
	nexaPath = path.Join(cache, "nexa")

	// Initialize SDK and LLM for testing
	Init()
	initLLM()

	// Run all tests
	code := m.Run()

	// Clean up resources after testing
	deinitLLM()
	DeInit()
	os.Exit(code)
}
