package integration

import (
	"context"
	"errors"
	"testing"
	"time"

	nexmonyx "github.com/nexmonyx/go-sdk/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNetworkFailureRecovery tests SDK behavior during network issues
func TestNetworkFailureRecovery(t *testing.T) {
	skipIfShort(t)

	t.Run("TimeoutHandling", func(t *testing.T) {
		// Create client with very short timeout
		client, err := nexmonyx.NewClient(&nexmonyx.Config{
			BaseURL: "http://192.0.2.1:9999", // Non-routable IP (TEST-NET-1)
			Timeout: 500 * time.Millisecond,  // Very short timeout
			Auth: nexmonyx.AuthConfig{
				Token: "test-token",
			},
		})
		require.NoError(t, err, "Failed to create client")

		// Should timeout quickly
		ctx := context.Background()
		start := time.Now()
		_, _, err = client.Servers.List(ctx, &nexmonyx.ListOptions{Page: 1, Limit: 10})
		duration := time.Since(start)

		require.Error(t, err, "Request should timeout")
		assert.Less(t, duration, 2*time.Second, "Should timeout quickly")

		t.Logf("Request timed out after %v as expected", duration)
	})

	t.Run("ConnectionRefused", func(t *testing.T) {
		// Create client pointing to closed port
		client, err := nexmonyx.NewClient(&nexmonyx.Config{
			BaseURL: "http://localhost:9999", // Nothing listening
			Timeout: 2 * time.Second,
			Auth: nexmonyx.AuthConfig{
				Token: "test-token",
			},
		})
		require.NoError(t, err, "Failed to create client")

		_, _, err = client.Servers.List(context.Background(), &nexmonyx.ListOptions{Page: 1, Limit: 10})
		require.Error(t, err, "Should fail with connection error")

		t.Logf("Connection refused as expected: %v", err)
	})

	t.Run("DNSFailure", func(t *testing.T) {
		// Create client with non-existent domain
		client, err := nexmonyx.NewClient(&nexmonyx.Config{
			BaseURL: "http://this-domain-absolutely-does-not-exist-12345.invalid",
			Timeout: 2 * time.Second,
			Auth: nexmonyx.AuthConfig{
				Token: "test-token",
			},
		})
		require.NoError(t, err, "Failed to create client")

		_, _, err = client.Servers.List(context.Background(), &nexmonyx.ListOptions{Page: 1, Limit: 10})
		require.Error(t, err, "Should fail with DNS error")

		t.Logf("DNS failure as expected: %v", err)
	})
}

// TestAPIRateLimiting tests rate limit handling
func TestAPIRateLimiting(t *testing.T) {
	skipIfShort(t)

	t.Run("RateLimitSimulation", func(t *testing.T) {
		// Note: Mock server doesn't implement rate limiting
		// This test would work against dev API with actual rate limits
		t.Skip("Mock server doesn't implement rate limiting - test against dev API")
	})
}

// TestContextCancellation tests context cancellation handling
func TestContextCancellation(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("CancelDuringRequest", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		// Cancel immediately
		cancel()

		_, _, err := env.Client.Servers.List(ctx, &nexmonyx.ListOptions{Page: 1, Limit: 10})
		require.Error(t, err, "Request should be canceled")
		assert.True(t, errors.Is(err, context.Canceled), "Error should be context.Canceled")

		t.Logf("Context cancellation handled correctly")
	})

	t.Run("TimeoutContext", func(t *testing.T) {
		// Create context with very short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		time.Sleep(10 * time.Millisecond) // Ensure timeout passes

		_, _, err := env.Client.Servers.List(ctx, &nexmonyx.ListOptions{Page: 1, Limit: 10})
		require.Error(t, err, "Request should timeout")
		assert.True(t, errors.Is(err, context.DeadlineExceeded), "Error should be context.DeadlineExceeded")

		t.Logf("Context deadline exceeded handled correctly")
	})

	t.Run("ValidContextSucceeds", func(t *testing.T) {
		// Create context with reasonable timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		servers, _, err := env.Client.Servers.List(ctx, &nexmonyx.ListOptions{Page: 1, Limit: 10})
		require.NoError(t, err, "Request should succeed with valid context")
		assert.NotNil(t, servers, "Servers should not be nil")

		t.Logf("Valid context allows request to succeed")
	})
}

// TestResourceNotFound tests 404 handling
func TestResourceNotFound(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("ServerNotFound", func(t *testing.T) {
		_, err := env.Client.Servers.GetByUUID(env.Ctx, "non-existent-server-uuid-12345")
		require.Error(t, err, "Should return error for non-existent server")

		// Check if it's a NotFoundError
		_, isNotFound := err.(*nexmonyx.NotFoundError)
		assert.True(t, isNotFound, "Error should be NotFoundError type")

		t.Logf("NotFoundError correctly returned: %v", err)
	})

	t.Run("OrganizationNotFound", func(t *testing.T) {
		_, err := env.Client.Organizations.Get(env.Ctx, "non-existent-org-uuid-12345")
		require.Error(t, err, "Should return error for non-existent organization")

		t.Logf("Organization not found error: %v", err)
	})
}

// TestValidationErrors tests 400 validation error handling
func TestValidationErrors(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("MissingRequiredField", func(t *testing.T) {
		// Create server without required hostname
		server := &nexmonyx.Server{
			OrganizationID: 1,
			MainIP:         "192.168.1.100",
			// Missing required hostname
		}

		_, err := env.Client.Servers.Create(env.Ctx, server)
		require.Error(t, err, "Should fail without required hostname")

		t.Logf("Validation error for missing hostname: %v", err)
	})

	t.Run("InvalidDataFormat", func(t *testing.T) {
		// Create server with invalid data
		server := &nexmonyx.Server{
			Hostname:       "", // Empty hostname
			OrganizationID: 1,
			MainIP:         "192.168.1.100",
		}

		_, err := env.Client.Servers.Create(env.Ctx, server)
		require.Error(t, err, "Should fail with empty hostname")

		t.Logf("Validation error for empty hostname: %v", err)
	})
}

// TestPartialFailures tests handling of partial operation failures
func TestPartialFailures(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("BulkOperationPartialFailure", func(t *testing.T) {
		// Create multiple servers, some should succeed, some fail
		servers := []*nexmonyx.Server{
			{Hostname: "valid-server-1", OrganizationID: 1, MainIP: "192.168.1.1"},
			{Hostname: "", OrganizationID: 1, MainIP: "192.168.1.2"},            // Invalid - no hostname
			{Hostname: "valid-server-2", OrganizationID: 1, MainIP: "192.168.1.3"},
			{OrganizationID: 1, MainIP: "192.168.1.4"},                          // Invalid - no hostname
		}

		successCount := 0
		failureCount := 0
		createdUUIDs := []string{}

		for i, server := range servers {
			created, err := env.Client.Servers.Create(env.Ctx, server)
			if err != nil {
				failureCount++
				t.Logf("Server %d failed as expected: %v", i+1, err)
			} else {
				successCount++
				createdUUIDs = append(createdUUIDs, created.ServerUUID)
				t.Logf("Server %d created successfully: %s", i+1, created.ServerUUID)
			}
		}

		assert.Greater(t, successCount, 0, "Some servers should succeed")
		assert.Greater(t, failureCount, 0, "Some servers should fail")

		t.Logf("Bulk operation: %d succeeded, %d failed", successCount, failureCount)

		// Clean up successful creations
		for _, uuid := range createdUUIDs {
			env.Client.Servers.Delete(env.Ctx, uuid)
		}
	})
}

// TestConcurrentRequests tests handling of concurrent API requests
func TestConcurrentRequests(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("MultipleConcurrentGets", func(t *testing.T) {
		// Make multiple concurrent requests
		concurrency := 5
		done := make(chan error, concurrency)

		for i := 0; i < concurrency; i++ {
			go func(index int) {
				_, _, err := env.Client.Servers.List(env.Ctx, &nexmonyx.ListOptions{
					Page:  1,
					Limit: 10,
				})
				done <- err
			}(i)
		}

		// Wait for all requests to complete
		errorCount := 0
		for i := 0; i < concurrency; i++ {
			err := <-done
			if err != nil {
				errorCount++
				t.Logf("Request %d failed: %v", i+1, err)
			}
		}

		assert.Equal(t, 0, errorCount, "All concurrent requests should succeed")
		t.Logf("All %d concurrent requests completed successfully", concurrency)
	})
}

// TestInvalidJSONResponse tests handling of malformed API responses
func TestInvalidJSONResponse(t *testing.T) {
	skipIfShort(t)

	t.Run("MalformedResponse", func(t *testing.T) {
		// Note: Mock server returns valid JSON
		// This test would work against a mock that returns invalid JSON
		t.Skip("Mock server returns valid JSON - would need special mock for this test")
	})
}

// TestHTTPStatusCodes tests handling of various HTTP status codes
func TestHTTPStatusCodes(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("404NotFound", func(t *testing.T) {
		_, err := env.Client.Servers.GetByUUID(env.Ctx, "does-not-exist")
		require.Error(t, err, "Should return error for 404")

		t.Logf("404 handled correctly: %v", err)
	})

	t.Run("400BadRequest", func(t *testing.T) {
		// Send invalid data
		server := &nexmonyx.Server{
			// Missing required fields
			OrganizationID: 1,
		}

		_, err := env.Client.Servers.Create(env.Ctx, server)
		require.Error(t, err, "Should return error for invalid request")

		t.Logf("400 handled correctly: %v", err)
	})

	t.Run("401Unauthorized", func(t *testing.T) {
		// Create client with invalid auth
		badClient, err := nexmonyx.NewClient(&nexmonyx.Config{
			BaseURL: env.BaseURL,
			Auth: nexmonyx.AuthConfig{
				Token: "invalid-token",
			},
		})
		require.NoError(t, err, "Failed to create client")

		_, _, err = badClient.Servers.List(env.Ctx, &nexmonyx.ListOptions{Page: 1, Limit: 10})
		require.Error(t, err, "Should return error for unauthorized")

		t.Logf("401 handled correctly: %v", err)
	})
}

// TestRetryLogic tests automatic retry behavior (if implemented)
func TestRetryLogic(t *testing.T) {
	skipIfShort(t)

	t.Run("RetryOn5xxError", func(t *testing.T) {
		// Note: Would need mock server to simulate 503 with Retry-After
		t.Skip("Retry logic testing requires special mock server setup")
	})
}

// TestErrorMessages tests that error messages are informative
func TestErrorMessages(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("ErrorContainsUsefulInfo", func(t *testing.T) {
		_, err := env.Client.Servers.GetByUUID(env.Ctx, "non-existent")
		require.Error(t, err, "Should return error")

		errMsg := err.Error()
		assert.NotEmpty(t, errMsg, "Error message should not be empty")
		assert.Contains(t, errMsg, "Server not found", "Error should mention what wasn't found")

		t.Logf("Error message: %s", errMsg)
	})
}
