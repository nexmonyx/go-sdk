package integration

import (
	"os"
	"testing"
)

// TestMain is the entry point for all integration tests
func TestMain(m *testing.M) {
	// Check if integration tests should run
	if os.Getenv("INTEGRATION_TESTS") == "" {
		// Skip integration tests by default
		os.Exit(0)
	}

	// Set default test mode to mock if not specified
	if os.Getenv("INTEGRATION_TEST_MODE") == "" {
		os.Setenv("INTEGRATION_TEST_MODE", "mock")
	}

	// Run tests
	code := m.Run()

	// Exit with test result code
	os.Exit(code)
}

// Example: To run integration tests:
// INTEGRATION_TESTS=true go test -v ./tests/integration/...
//
// With debug logging:
// INTEGRATION_TESTS=true INTEGRATION_TEST_DEBUG=true go test -v ./tests/integration/...
//
// With custom timeout:
// INTEGRATION_TESTS=true INTEGRATION_TEST_TIMEOUT=60s go test -v ./tests/integration/...
