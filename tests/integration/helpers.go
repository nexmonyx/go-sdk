package integration

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	nexmonyx "github.com/nexmonyx/go-sdk/v2"
	"github.com/stretchr/testify/require"
)

// TestEnvironment holds the test environment configuration
type TestEnvironment struct {
	Client    *nexmonyx.Client
	MockAPI   *MockAPIServer
	BaseURL   string
	AuthToken string
	Ctx       context.Context
}

// isDevMode checks if we're running against dev API
func isDevMode() bool {
	return os.Getenv("INTEGRATION_TEST_MODE") == "dev"
}

// setupIntegrationTest initializes test environment (supports both mock and dev modes)
func setupIntegrationTest(t *testing.T) *TestEnvironment {
	t.Helper()

	var client *nexmonyx.Client
	var mockAPI *MockAPIServer
	var baseURL string
	var authToken string

	if isDevMode() {
		// Dev API mode - use real API server
		baseURL = os.Getenv("INTEGRATION_TEST_API_URL")
		if baseURL == "" {
			t.Fatal("INTEGRATION_TEST_API_URL must be set in dev mode")
		}

		authToken = os.Getenv("INTEGRATION_TEST_AUTH_TOKEN")
		if authToken == "" {
			t.Fatal("INTEGRATION_TEST_AUTH_TOKEN must be set in dev mode")
		}

		var err error
		client, err = nexmonyx.NewClient(&nexmonyx.Config{
			BaseURL: baseURL,
			Auth: nexmonyx.AuthConfig{
				Token: authToken,
			},
			Timeout: getTestTimeout(),
			Debug:   os.Getenv("INTEGRATION_TEST_DEBUG") == "true",
		})
		require.NoError(t, err, "Failed to create SDK client for dev API")

		t.Logf("Integration test environment (DEV MODE) initialized with API: %s", baseURL)
	} else {
		// Mock mode - use mock API server
		mockAPI = NewMockAPIServer(t)
		baseURL = mockAPI.Server.URL
		authToken = "test-token"

		var err error
		client, err = nexmonyx.NewClient(&nexmonyx.Config{
			BaseURL: baseURL,
			Auth: nexmonyx.AuthConfig{
				Token: authToken,
			},
			Timeout: getTestTimeout(),
			Debug:   os.Getenv("INTEGRATION_TEST_DEBUG") == "true",
		})
		require.NoError(t, err, "Failed to create SDK client for mock API")

		t.Logf("Integration test environment (MOCK MODE) initialized at %s", baseURL)
	}

	ctx := context.Background()

	env := &TestEnvironment{
		Client:    client,
		MockAPI:   mockAPI,
		BaseURL:   baseURL,
		AuthToken: authToken,
		Ctx:       ctx,
	}

	return env
}

// teardownIntegrationTest cleans up the test environment
func teardownIntegrationTest(t *testing.T, env *TestEnvironment) {
	t.Helper()

	if env.MockAPI != nil {
		// Mock mode - close mock server
		env.MockAPI.Close()
		t.Log("Integration test environment (MOCK MODE) cleaned up")
	} else {
		// Dev mode - perform cleanup operations on dev API
		cleanupDevResources(t, env)
		t.Log("Integration test environment (DEV MODE) cleaned up")
	}
}

// cleanupDevResources removes test resources from dev API
func cleanupDevResources(t *testing.T, env *TestEnvironment) {
	t.Helper()

	// Delete any servers created during tests with "test-" prefix
	servers, _, err := env.Client.Servers.List(env.Ctx, &nexmonyx.ListOptions{
		Search: "test-",
		Limit:  100,
	})
	if err != nil {
		t.Logf("Warning: Failed to list test servers for cleanup: %v", err)
		return
	}

	for _, server := range servers {
		if strings.HasPrefix(server.Hostname, "test-") ||
			strings.HasPrefix(server.Hostname, "workflow-test-") ||
			strings.HasPrefix(server.Hostname, "alert-test-") ||
			strings.HasPrefix(server.Hostname, "org-test-") {
			err := env.Client.Servers.Delete(env.Ctx, server.ServerUUID)
			if err != nil {
				t.Logf("Warning: Failed to delete test server %s: %v", server.ServerUUID, err)
			} else {
				t.Logf("Cleaned up test server: %s", server.ServerUUID)
			}
		}
	}
}

// loadFixture loads a JSON fixture file and returns the parsed data
func loadFixture(t *testing.T, filename string) interface{} {
	t.Helper()

	// Get the directory of the current file
	_, currentFile, _, ok := runtime.Caller(0)
	require.True(t, ok, "Failed to get current file path")

	fixturesDir := filepath.Join(filepath.Dir(currentFile), "fixtures")
	fixturePath := filepath.Join(fixturesDir, filename)

	data, err := os.ReadFile(fixturePath)
	require.NoError(t, err, "Failed to read fixture file %s", filename)

	var result interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err, "Failed to parse fixture file %s", filename)

	t.Logf("Loaded fixture: %s (%d bytes)", filename, len(data))
	return result
}

// Test Data Creation Helpers

// createTestServer creates a test server via the API
func createTestServer(t *testing.T, env *TestEnvironment, hostname string) *nexmonyx.Server {
	t.Helper()

	server := &nexmonyx.Server{
		Hostname:       hostname,
		OrganizationID: 1,
		MainIP:         "192.168.1.200",
		Location:       "Test-Location",
		Environment:    "testing",
		Classification: "test",
	}

	created, err := env.Client.Servers.Create(env.Ctx, server)
	require.NoError(t, err, "Failed to create test server")
	require.NotNil(t, created, "Server should not be nil")

	t.Logf("Created test server: %s (UUID: %s)", hostname, created.ServerUUID)
	return created
}

// createTestOrganization creates a test organization via the API
// assertValidUUID checks if a string is a valid UUID format
func assertValidUUID(t *testing.T, uuid string, fieldName string) {
	t.Helper()

	require.NotEmpty(t, uuid, "%s should not be empty", fieldName)
	require.Greater(t, len(uuid), 0, "%s should have length > 0", fieldName)

	// Basic UUID format check (either UUID v4 format or our custom format)
	// We accept both "server-001" style and "550e8400-e29b-41d4-a716-446655440000" style
	require.True(t, len(uuid) >= 8, "%s should have length >= 8 (got: %s)", fieldName, uuid)
}

// assertPaginationValid checks if pagination metadata is valid
func assertPaginationValid(t *testing.T, meta *nexmonyx.PaginationMeta) {
	t.Helper()

	require.NotNil(t, meta, "Pagination metadata should not be nil")
	require.Greater(t, meta.TotalItems, 0, "TotalItems should be greater than 0")
	require.Greater(t, meta.Page, 0, "Page should be greater than 0")
	require.Greater(t, meta.Limit, 0, "Limit should be greater than 0")
	require.Greater(t, meta.TotalPages, 0, "TotalPages should be greater than 0")
	require.LessOrEqual(t, meta.Page, meta.TotalPages,
		"Page should be less than or equal to TotalPages")
}

// Utility Functions

// skipIfShort skips the test if running in short mode
func skipIfShort(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
}

// getTestTimeout returns the timeout duration for integration tests
func getTestTimeout() time.Duration {
	timeoutStr := os.Getenv("INTEGRATION_TEST_TIMEOUT")
	if timeoutStr == "" {
		return 30 * time.Second // Default timeout
	}

	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return 30 * time.Second
	}
	return timeout
}

// Metrics Helpers

// createTestMetricsPayload creates a sample comprehensive metrics payload
func createTestMetricsPayload(serverUUID string) *nexmonyx.ComprehensiveMetricsRequest {
	return &nexmonyx.ComprehensiveMetricsRequest{
		ServerUUID:  serverUUID,
		CollectedAt: time.Now().Format(time.RFC3339),
		CPU: &nexmonyx.CPUMetrics{
			UsagePercent:  45.2,
			CoreCount:     8,
			LoadAverage1:  2.1,
			LoadAverage5:  1.8,
			LoadAverage15: 1.5,
		},
		Memory: &nexmonyx.MemoryMetrics{
			TotalBytes:     16777216000,
			UsedBytes:      8388608000,
			AvailableBytes: 8388608000,
			UsagePercent:   50.0,
		},
		Disks: []nexmonyx.DiskMetrics{
			{
				Mountpoint:   "/",
				TotalBytes:   107374182400,
				UsedBytes:    64424509440,
				FreeBytes:    42949672960,
				UsagePercent: 60.0,
			},
		},
		Network: []nexmonyx.NetworkMetrics{
			{
				Interface:   "eth0",
				BytesSent:   1073741824,
				BytesRecv:   2147483648,
				PacketsSent: 1000000,
				PacketsRecv: 1500000,
				ErrorsIn:    0,
				ErrorsOut:   0,
			},
		},
	}
}
