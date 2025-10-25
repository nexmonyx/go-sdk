// +build integration

package nexmonyx

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration tests require the following environment variables:
// - NEXMONYX_API_URL: The base URL of the Nexmonyx API (e.g., https://api.nexmonyx.com)
// - NEXMONYX_AUTH_TOKEN: A valid JWT token for authentication
// - NEXMONYX_SERVER_UUID: (Optional) Server UUID for agent tests
// - NEXMONYX_DEBUG: (Optional) Set to "true" for debug logging
//
// Run with: go test -tags=integration -v -timeout 30m

// getTestConfig returns a client configuration from environment variables
func getTestConfig(t *testing.T) *Config {
	apiURL := os.Getenv("NEXMONYX_API_URL")
	authToken := os.Getenv("NEXMONYX_AUTH_TOKEN")

	if apiURL == "" {
		t.Skip("NEXMONYX_API_URL not set - skipping integration tests")
	}
	if authToken == "" {
		t.Skip("NEXMONYX_AUTH_TOKEN not set - skipping integration tests")
	}

	config := &Config{
		BaseURL: apiURL,
		Auth: AuthConfig{
			Token: authToken,
		},
	}

	if os.Getenv("NEXMONYX_DEBUG") == "true" {
		config.Debug = true
	}

	return config
}

// TestIntegration_UserProfile tests retrieving the current user profile
func TestIntegration_UserProfile(t *testing.T) {
	config := getTestConfig(t)
	client, err := NewClient(config)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("Get current user", func(t *testing.T) {
		user, err := client.Users.GetCurrent(ctx)
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.NotZero(t, user.ID)
		assert.NotEmpty(t, user.Email)
		t.Logf("Current user: %s (ID: %d)", user.Email, user.ID)
	})
}

// TestIntegration_Organizations tests organization operations
func TestIntegration_Organizations(t *testing.T) {
	config := getTestConfig(t)
	client, err := NewClient(config)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("List organizations", func(t *testing.T) {
		orgs, meta, err := client.Organizations.List(ctx, &ListOptions{
			Page:  1,
			Limit: 10,
		})
		require.NoError(t, err)
		assert.NotNil(t, orgs)
		assert.NotNil(t, meta)
		assert.GreaterOrEqual(t, meta.TotalItems, 0)

		if len(orgs) > 0 {
			t.Logf("Found %d organizations, first org: %s (ID: %d)",
				len(orgs), orgs[0].Name, orgs[0].ID)
		}
	})

	t.Run("Get specific organization", func(t *testing.T) {
		orgs, _, err := client.Organizations.List(ctx, &ListOptions{Limit: 1})
		require.NoError(t, err)

		if len(orgs) > 0 {
			orgUUID := orgs[0].UUID
			org, err := client.Organizations.Get(ctx, orgUUID)
			require.NoError(t, err)
			assert.NotNil(t, org)
			assert.Equal(t, orgUUID, org.UUID)
			t.Logf("Retrieved organization: %s", org.Name)
		} else {
			t.Skip("No organizations available for testing")
		}
	})
}

// TestIntegration_Servers tests server operations
func TestIntegration_Servers(t *testing.T) {
	config := getTestConfig(t)
	client, err := NewClient(config)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("List servers", func(t *testing.T) {
		servers, meta, err := client.Servers.List(ctx, &ListOptions{
			Page:  1,
			Limit: 10,
		})
		require.NoError(t, err)
		assert.NotNil(t, servers)
		assert.NotNil(t, meta)

		if len(servers) > 0 {
			t.Logf("Found %d servers, first server: %s (UUID: %s)",
				len(servers), servers[0].Hostname, servers[0].ServerUUID)
		}
	})

	t.Run("Get server details", func(t *testing.T) {
		servers, _, err := client.Servers.List(ctx, &ListOptions{Limit: 1})
		require.NoError(t, err)

		if len(servers) > 0 {
			serverUUID := servers[0].ServerUUID
			server, err := client.Servers.Get(ctx, serverUUID)
			require.NoError(t, err)
			assert.NotNil(t, server)
			assert.Equal(t, serverUUID, server.ServerUUID)
			t.Logf("Retrieved server: %s", server.Hostname)
		} else {
			t.Skip("No servers available for testing")
		}
	})
}

// TestIntegration_ServerAgent tests agent-specific operations
func TestIntegration_ServerAgent(t *testing.T) {
	serverUUID := os.Getenv("NEXMONYX_SERVER_UUID")
	if serverUUID == "" {
		t.Skip("NEXMONYX_SERVER_UUID not set - skipping agent tests")
	}

	config := getTestConfig(t)
	client, err := NewClient(config)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("Submit heartbeat", func(t *testing.T) {
		err := client.Servers.SendHeartbeat(ctx, serverUUID)
		assert.NoError(t, err)
		t.Logf("Heartbeat sent successfully for server %s", serverUUID)
	})
}

// TestIntegration_Alerts tests alert operations
func TestIntegration_Alerts(t *testing.T) {
	config := getTestConfig(t)
	client, err := NewClient(config)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("List alerts", func(t *testing.T) {
		alerts, meta, err := client.Alerts.List(ctx, &ListOptions{
			Page:  1,
			Limit: 10,
		})
		require.NoError(t, err)
		assert.NotNil(t, alerts)
		assert.NotNil(t, meta)
		t.Logf("Found %d alerts (Total: %d)", len(alerts), meta.TotalItems)

		if len(alerts) > 0 {
			t.Logf("First alert: %s (Type: %s, Severity: %s)",
				alerts[0].Name, alerts[0].Type, alerts[0].Severity)
		}
	})
}

// TestIntegration_Pagination tests pagination across different endpoints
func TestIntegration_Pagination(t *testing.T) {
	config := getTestConfig(t)
	client, err := NewClient(config)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("Paginate through servers", func(t *testing.T) {
		page1, meta1, err := client.Servers.List(ctx, &ListOptions{
			Page:  1,
			Limit: 5,
		})
		require.NoError(t, err)
		assert.NotNil(t, meta1)

		if meta1.TotalPages > 1 {
			page2, meta2, err := client.Servers.List(ctx, &ListOptions{
				Page:  2,
				Limit: 5,
			})
			require.NoError(t, err)
			assert.NotNil(t, page2)
			assert.NotNil(t, meta2)

			assert.Equal(t, 2, meta2.Page)
			assert.Equal(t, meta1.TotalItems, meta2.TotalItems)

			t.Logf("Page 1: %d items, Page 2: %d items, Total: %d",
				len(page1), len(page2), meta1.TotalItems)
		} else {
			t.Log("Only one page available")
		}
	})
}

// TestIntegration_ErrorHandling tests API error responses
func TestIntegration_ErrorHandling(t *testing.T) {
	config := getTestConfig(t)
	client, err := NewClient(config)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("Not found error", func(t *testing.T) {
		_, err := client.Servers.Get(ctx, "non-existent-uuid-12345")
		assert.Error(t, err)

		if apiErr, ok := err.(*APIError); ok {
			t.Logf("Correctly received not found error: %s", apiErr.Message)
		}
	})
}

// TestIntegration_ContextCancellation tests context handling
func TestIntegration_ContextCancellation(t *testing.T) {
	config := getTestConfig(t)
	client, err := NewClient(config)
	require.NoError(t, err)

	t.Run("Cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, _, err := client.Servers.List(ctx, nil)
		assert.Error(t, err)
		t.Logf("Context cancellation handled correctly: %v", err)
	})

	t.Run("Timeout context", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		time.Sleep(10 * time.Millisecond)

		_, _, err := client.Servers.List(ctx, nil)
		assert.Error(t, err)
		t.Logf("Context timeout handled correctly: %v", err)
	})
}
