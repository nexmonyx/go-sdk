package integration

import (
	"context"
	"testing"

	nexmonyx "github.com/nexmonyx/go-sdk/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJWTAuthentication tests JWT token-based authentication
func TestJWTAuthentication(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("ValidJWTToken", func(t *testing.T) {
		// Create client with valid JWT token
		client, err := nexmonyx.NewClient(&nexmonyx.Config{
			BaseURL: env.BaseURL,
			Auth: nexmonyx.AuthConfig{
				Token: "test-token",
			},
		})
		require.NoError(t, err, "Failed to create client with JWT token")

		// Test authenticated request
		servers, _, err := client.Servers.List(context.Background(), &nexmonyx.ListOptions{
			Page:  1,
			Limit: 10,
		})
		require.NoError(t, err, "Failed to list servers with valid token")
		assert.NotNil(t, servers, "Servers list should not be nil")

		t.Logf("Successfully authenticated with JWT token and retrieved %d servers", len(servers))
	})

	t.Run("InvalidJWTToken", func(t *testing.T) {
		// Create client with invalid token
		client, err := nexmonyx.NewClient(&nexmonyx.Config{
			BaseURL: env.BaseURL,
			Auth: nexmonyx.AuthConfig{
				Token: "invalid-token-12345",
			},
		})
		require.NoError(t, err, "Failed to create client")

		// Should fail with unauthorized error
		_, _, err = client.Servers.List(context.Background(), &nexmonyx.ListOptions{
			Page:  1,
			Limit: 10,
		})
		require.Error(t, err, "Should fail with invalid token")

		t.Logf("Correctly rejected invalid JWT token: %v", err)
	})

	t.Run("MissingToken", func(t *testing.T) {
		// Create client without auth config
		client, err := nexmonyx.NewClient(&nexmonyx.Config{
			BaseURL: env.BaseURL,
			// No auth config
		})
		require.NoError(t, err, "Failed to create client")

		// Should fail with unauthorized error
		_, _, err = client.Servers.List(context.Background(), &nexmonyx.ListOptions{
			Page:  1,
			Limit: 10,
		})
		require.Error(t, err, "Should fail without authentication")

		t.Logf("Correctly rejected request without authentication: %v", err)
	})
}

// TestAPIKeyAuthentication tests API key/secret authentication
func TestAPIKeyAuthentication(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("ValidAPIKeySecret", func(t *testing.T) {
		// Note: Mock server uses Bearer token auth
		// In real dev API, this would test API key/secret auth
		t.Skip("Mock server doesn't support API key auth - test against dev API")
	})

	t.Run("InvalidAPIKey", func(t *testing.T) {
		// Would test with invalid API key
		t.Skip("Mock server doesn't support API key auth - test against dev API")
	})
}

// TestServerCredentialsAuthentication tests server UUID/secret authentication
func TestServerCredentialsAuthentication(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("ValidServerCredentials", func(t *testing.T) {
		// Note: Mock server uses Bearer token auth
		// In real dev API, this would test server credentials auth
		t.Skip("Mock server doesn't support server credentials auth - test against dev API")
	})

	t.Run("InvalidServerCredentials", func(t *testing.T) {
		// Would test with invalid server credentials
		t.Skip("Mock server doesn't support server credentials auth - test against dev API")
	})
}

// TestAuthenticationHeaders tests that correct headers are sent
func TestAuthenticationHeaders(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("BearerTokenHeader", func(t *testing.T) {
		// Verify Authorization: Bearer <token> header is sent
		client, err := nexmonyx.NewClient(&nexmonyx.Config{
			BaseURL: env.BaseURL,
			Auth: nexmonyx.AuthConfig{
				Token: "test-token",
			},
			Debug: false, // Set to true to see request/response details
		})
		require.NoError(t, err, "Failed to create client")

		// Make request - headers are automatically added by SDK
		_, _, err = client.Servers.List(context.Background(), &nexmonyx.ListOptions{
			Page:  1,
			Limit: 5,
		})
		require.NoError(t, err, "Request should succeed with proper headers")

		t.Logf("Authentication headers correctly sent and accepted")
	})
}

// TestMultipleAuthMethods tests that only one auth method can be used at a time
func TestMultipleAuthMethods(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("TokenTakesPrecedence", func(t *testing.T) {
		// If multiple auth methods are provided, token should take precedence
		client, err := nexmonyx.NewClient(&nexmonyx.Config{
			BaseURL: env.BaseURL,
			Auth: nexmonyx.AuthConfig{
				Token:        "test-token", // This should be used
				APIKey:       "some-api-key",
				APISecret:    "some-secret",
				ServerUUID:   "some-uuid",
				ServerSecret: "some-server-secret",
			},
		})
		require.NoError(t, err, "Failed to create client")

		// Should authenticate successfully with token
		_, _, err = client.Servers.List(context.Background(), &nexmonyx.ListOptions{
			Page:  1,
			Limit: 5,
		})
		require.NoError(t, err, "Should authenticate with token")

		t.Logf("Token auth takes precedence when multiple methods provided")
	})
}

// TestAuthenticationWithDifferentEndpoints tests auth across different API endpoints
func TestAuthenticationWithDifferentEndpoints(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("AuthWorksAcrossEndpoints", func(t *testing.T) {
		client, err := nexmonyx.NewClient(&nexmonyx.Config{
			BaseURL: env.BaseURL,
			Auth: nexmonyx.AuthConfig{
				Token: "test-token",
			},
		})
		require.NoError(t, err, "Failed to create client")

		ctx := context.Background()

		// Test servers endpoint
		servers, _, err := client.Servers.List(ctx, &nexmonyx.ListOptions{Page: 1, Limit: 5})
		require.NoError(t, err, "Servers endpoint should work")
		assert.NotNil(t, servers)

		// Test organizations endpoint
		orgs, _, err := client.Organizations.List(ctx, &nexmonyx.ListOptions{Page: 1, Limit: 5})
		require.NoError(t, err, "Organizations endpoint should work")
		assert.NotNil(t, orgs)

		// Test health endpoint
		health, err := client.Health.GetHealth(ctx)
		require.NoError(t, err, "Health check endpoint should work")
		assert.NotNil(t, health)

		t.Logf("Authentication works across multiple endpoints")
	})
}

// TestReauthentication tests behavior when re-creating client with same credentials
func TestReauthentication(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("MultipleClientsWithSameCredentials", func(t *testing.T) {
		ctx := context.Background()

		// Create first client
		client1, err := nexmonyx.NewClient(&nexmonyx.Config{
			BaseURL: env.BaseURL,
			Auth: nexmonyx.AuthConfig{
				Token: "test-token",
			},
		})
		require.NoError(t, err, "Failed to create first client")

		// Create second client with same credentials
		client2, err := nexmonyx.NewClient(&nexmonyx.Config{
			BaseURL: env.BaseURL,
			Auth: nexmonyx.AuthConfig{
				Token: "test-token",
			},
		})
		require.NoError(t, err, "Failed to create second client")

		// Both clients should work
		servers1, _, err := client1.Servers.List(ctx, &nexmonyx.ListOptions{Page: 1, Limit: 5})
		require.NoError(t, err, "First client should work")
		assert.NotNil(t, servers1)

		servers2, _, err := client2.Servers.List(ctx, &nexmonyx.ListOptions{Page: 1, Limit: 5})
		require.NoError(t, err, "Second client should work")
		assert.NotNil(t, servers2)

		t.Logf("Multiple clients with same credentials work correctly")
	})
}
