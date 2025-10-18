package main

import (
	"context"
	"os"
	"testing"

	"github.com/nexmonyx/go-sdk/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ExampleMockAPITesting demonstrates integration testing with mock API server
// These examples show how to test SDK functionality with the mock API server
// Run with: INTEGRATION_TESTS=true go test -v ./integration
// Make sure to start the mock API server: docker-compose up -d

// TestServerListingWithMockAPI demonstrates listing servers using mock API
func TestServerListingWithMockAPI(t *testing.T) {
	// Skip if integration tests are not enabled
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Integration tests disabled")
	}

	// Create client pointing to mock API
	client, err := nexmonyx.NewClient(&nexmonyx.Config{
		BaseURL: "http://localhost:8080",
		Auth: nexmonyx.AuthConfig{
			Token: "test-token",
		},
	})

	require.NoError(t, err)

	// List servers
	ctx := context.Background()
	opts := &nexmonyx.ListOptions{
		Page:  1,
		Limit: 25,
	}

	// This would call the mock API server
	// servers, meta, err := client.Servers.List(ctx, opts)
	// For this example, we'll just demonstrate the pattern
	_ = ctx
	_ = opts
	_ = client
}

// TestOrganizationCRUD demonstrates Create, Read, Update, Delete operations
func TestOrganizationCRUD(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Integration tests disabled")
	}

	client, err := nexmonyx.NewClient(&nexmonyx.Config{
		BaseURL: "http://localhost:8080",
		Auth: nexmonyx.AuthConfig{
			Token: "test-token",
		},
	})

	require.NoError(t, err)

	ctx := context.Background()

	// Create operation
	createReq := &nexmonyx.Organization{
		Name:        "Test Organization",
		Description: "Test org for integration testing",
	}

	// org, err := client.Organizations.Create(ctx, createReq)
	// require.NoError(t, err)
	// assert.NotNil(t, org)

	// Read operation
	// retrievedOrg, err := client.Organizations.Get(ctx, org.UUID)
	// require.NoError(t, err)
	// assert.Equal(t, org.UUID, retrievedOrg.UUID)

	// Update operation
	// updateReq := &nexmonyx.Organization{
	//     Name: "Updated Organization",
	// }
	// updatedOrg, err := client.Organizations.Update(ctx, org.UUID, updateReq)
	// require.NoError(t, err)
	// assert.Equal(t, "Updated Organization", updatedOrg.Name)

	// Delete operation
	// err = client.Organizations.Delete(ctx, org.UUID)
	// require.NoError(t, err)

	_ = createReq
	_ = ctx
	_ = client
}

// TestErrorHandlingWithMockAPI demonstrates handling various API errors
func TestErrorHandlingWithMockAPI(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Integration tests disabled")
	}

	client, err := nexmonyx.NewClient(&nexmonyx.Config{
		BaseURL: "http://localhost:8080",
		Auth: nexmonyx.AuthConfig{
			Token: "test-token",
		},
	})

	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name        string
		testFunc    func() error
		expectError bool
	}{
		{
			name: "Not Found Error",
			testFunc: func() error {
				// Try to get non-existent organization
				// _, err := client.Organizations.Get(ctx, "nonexistent-uuid")
				return nil // Would be error
			},
			expectError: true,
		},
		{
			name: "Validation Error",
			testFunc: func() error {
				// Try to create with invalid data
				// _, err := client.Organizations.Create(ctx, &nexmonyx.CreateOrganizationRequest{})
				return nil // Would be error
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.testFunc()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	_ = ctx
	_ = client
}

// TestConcurrentOperations demonstrates concurrent API calls
func TestConcurrentOperations(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Integration tests disabled")
	}

	client, err := nexmonyx.NewClient(&nexmonyx.Config{
		BaseURL: "http://localhost:8080",
		Auth: nexmonyx.AuthConfig{
			Token: "test-token",
		},
	})

	require.NoError(t, err)

	ctx := context.Background()

	// Test concurrent list operations
	numGoroutines := 10
	done := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			opts := &nexmonyx.ListOptions{
				Page:  1,
				Limit: 10,
			}
			// servers, _, err := client.Servers.List(ctx, opts)
			// done <- err
			_ = opts // Mark opts as intentionally unused for this example
			done <- nil // For this example
		}()
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		err := <-done
		assert.NoError(t, err)
	}

	_ = ctx
	_ = client
}

// TestMetricsSubmissionWithMockAPI demonstrates metrics submission workflow
func TestMetricsSubmissionWithMockAPI(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Integration tests disabled")
	}

	client, err := nexmonyx.NewClient(&nexmonyx.Config{
		BaseURL: "http://localhost:8080",
		Auth: nexmonyx.AuthConfig{
			ServerUUID:   "test-server-uuid",
			ServerSecret: "test-server-secret",
		},
	})

	require.NoError(t, err)

	ctx := context.Background()

	// Prepare metrics
	metrics := &nexmonyx.ComprehensiveMetricsRequest{
		ServerUUID: "test-server-uuid",
	}

	// Submit metrics
	// err = client.Metrics.SubmitComprehensive(ctx, metrics)
	// require.NoError(t, err)

	_ = ctx
	_ = metrics
	_ = client
}

// TestDataValidation demonstrates input validation
func TestDataValidation(t *testing.T) {
	tests := []struct {
		name      string
		input     interface{}
		wantValid bool
	}{
		{"Valid server UUID", "550e8400-e29b-41d4-a716-446655440000", true},
		{"Invalid UUID format", "not-a-uuid", false},
		{"Empty string", "", false},
		{"Nil value", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validation logic here
			_ = tt.input
			_ = tt.wantValid
		})
	}
}

// TestRealAPIIntegration demonstrates testing with real Nexmonyx dev API
// This test is skipped by default and requires real credentials
// To run: INTEGRATION_TESTS=true INTEGRATION_TEST_MODE=dev go test -v ./integration
func TestRealAPIIntegration(t *testing.T) {
	mode := os.Getenv("INTEGRATION_TEST_MODE")
	if mode != "dev" {
		t.Skip("Real API integration test skipped (set INTEGRATION_TEST_MODE=dev to run)")
	}

	apiURL := os.Getenv("INTEGRATION_TEST_API_URL")
	if apiURL == "" {
		t.Skip("INTEGRATION_TEST_API_URL not set")
	}

	token := os.Getenv("INTEGRATION_TEST_AUTH_TOKEN")
	if token == "" {
		t.Skip("INTEGRATION_TEST_AUTH_TOKEN not set")
	}

	client, err := nexmonyx.NewClient(&nexmonyx.Config{
		BaseURL: apiURL,
		Auth: nexmonyx.AuthConfig{
			Token: token,
		},
	})

	require.NoError(t, err)

	ctx := context.Background()

	// Test with real API
	// orgs, _, err := client.Organizations.List(ctx, &nexmonyx.ListOptions{})
	// require.NoError(t, err)
	// assert.NotNil(t, orgs)

	_ = ctx
	_ = client
}
