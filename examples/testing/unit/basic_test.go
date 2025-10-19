package main

import (
	"testing"

	"github.com/nexmonyx/go-sdk/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ExampleClientCreation demonstrates basic client creation
// This is a simple unit test showing how to create a Nexmonyx SDK client
func TestClientCreation(t *testing.T) {
	// Create a new client with JWT token authentication
	client, err := nexmonyx.NewClient(&nexmonyx.Config{
		BaseURL: "https://api.example.com",
		Auth: nexmonyx.AuthConfig{
			Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
		},
	})

	// Assert no error occurred during client creation
	require.NoError(t, err, "Client creation should succeed")

	// Assert client is not nil
	assert.NotNil(t, client, "Client should be created")
}

// TestClientWithAPIKeyAuth demonstrates client creation with API key authentication
func TestClientWithAPIKeyAuth(t *testing.T) {
	client, err := nexmonyx.NewClient(&nexmonyx.Config{
		BaseURL: "https://api.example.com",
		Auth: nexmonyx.AuthConfig{
			APIKey:    "test-key-123",
			APISecret: "test-secret-456",
		},
	})

	require.NoError(t, err)
	assert.NotNil(t, client)
}

// TestClientWithServerCredentials demonstrates agent authentication
func TestClientWithServerCredentials(t *testing.T) {
	client, err := nexmonyx.NewClient(&nexmonyx.Config{
		BaseURL: "https://api.example.com",
		Auth: nexmonyx.AuthConfig{
			ServerUUID:   "server-123",
			ServerSecret: "secret-456",
		},
	})

	require.NoError(t, err)
	assert.NotNil(t, client)
}

// TestClientAuthMethodSwitching demonstrates changing authentication method
func TestClientAuthMethodSwitching(t *testing.T) {
	client, err := nexmonyx.NewClient(&nexmonyx.Config{
		BaseURL: "https://api.example.com",
		Auth: nexmonyx.AuthConfig{
			Token: "initial-token",
		},
	})

	require.NoError(t, err)

	// Switch to API key authentication
	newClient := client.WithAPIKey("new-key", "new-secret")
	assert.NotNil(t, newClient)
}

// TestResponseTypeParsing demonstrates parsing different response types
func TestResponseTypeParsing(t *testing.T) {
	// This example shows how to structure tests for different response types
	tests := []struct {
		name          string
		responseType  string
		expectedError bool
	}{
		{"Standard Response", "success", false},
		{"Error Response", "error", true},
		{"Empty Response", "empty", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test logic here
			_ = tt.responseType
		})
	}
}

// TestErrorHandling demonstrates how to handle different error types
func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		errorType   string
		expectError bool
	}{
		{"Validation Error", 400, "validation", true},
		{"Unauthorized", 401, "auth", true},
		{"Forbidden", 403, "permission", true},
		{"Not Found", 404, "notfound", true},
		{"Rate Limited", 429, "ratelimit", true},
		{"Server Error", 500, "server", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test error handling for each status code
			_ = tt.statusCode
			_ = tt.errorType
		})
	}
}

// TestListOptions demonstrates pagination options
func TestListOptions(t *testing.T) {
	opts := &nexmonyx.ListOptions{
		Page:   1,
		Limit:  25,
		Search: "test",
		Sort:   "name",
		Order:  "asc",
	}

	// Convert to query parameters
	query := opts.ToQuery()

	assert.Equal(t, "1", query["page"])
	assert.Equal(t, "25", query["limit"])
	assert.Equal(t, "test", query["search"])
	assert.Equal(t, "name", query["sort"])
	assert.Equal(t, "asc", query["order"])
}

// TestTableDrivenTests demonstrates the table-driven testing pattern
func TestTableDrivenTests(t *testing.T) {
	testCases := []struct {
		name     string
		input    interface{}
		expected string
		wantErr  bool
	}{
		{"Valid input", "test", "test", false},
		{"Empty input", "", "", false},
		{"Nil input", nil, "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test implementation
			_ = tc.input
			_ = tc.expected
			_ = tc.wantErr
		})
	}
}

// BenchmarkClientCreation demonstrates benchmarking client creation
// Run with: go test -bench=BenchmarkClientCreation -benchmem ./unit
func BenchmarkClientCreation(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = nexmonyx.NewClient(&nexmonyx.Config{
			BaseURL: "https://api.example.com",
			Auth: nexmonyx.AuthConfig{
				Token: "test-token",
			},
		})
	}
}
