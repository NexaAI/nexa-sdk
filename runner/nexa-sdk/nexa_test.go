package nexa_sdk

import (
	"os"
	"path"
	"testing"
)

// nexaPath holds the base directory for Nexa cache and models
var nexaPath string

// TestMain sets up the test environment and runs all tests
// It initializes the SDK, creates an LLM instance, runs tests, then cleans up
func TestMain(m *testing.M) {
	// Get user cache directory and set Nexa path
	cache, _ := os.UserCacheDir()
	nexaPath = path.Join(cache, "nexa")

	// Initialize SDK and modules for testing
	Init()
	defer DeInit()
	initLLM()
	defer deinitLLM()

	// Run all tests
	code := m.Run()

	os.Exit(code)
}
