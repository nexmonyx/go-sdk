package integration

import (
	"testing"

	nexmonyx "github.com/nexmonyx/go-sdk/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServersIntegration tests the complete server lifecycle
func TestServersIntegration(t *testing.T) {
	skipIfShort(t)

	// Setup test environment
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("ListServers", func(t *testing.T) {
		// List all servers
		servers, meta, err := env.Client.Servers.List(env.Ctx, &nexmonyx.ListOptions{
			Page:  1,
			Limit: 25,
		})

		require.NoError(t, err, "Failed to list servers")
		require.NotNil(t, servers, "Servers list should not be nil")
		require.NotEmpty(t, servers, "Servers list should not be empty")

		// Verify pagination metadata
		assertPaginationValid(t, meta)

		// Verify at least one server from fixtures is present
		assert.Greater(t, len(servers), 0, "Should have at least one server")

		// Verify first server has valid fields
		firstServer := servers[0]
		assertValidUUID(t, firstServer.ServerUUID, "Server UUID")
		assert.NotEmpty(t, firstServer.Hostname, "Server hostname should not be empty")
		// Note: CreatedAt and UpdatedAt are *CustomTime, not time.Time
		assert.NotNil(t, firstServer.CreatedAt, "Server created_at should not be nil")
		assert.NotNil(t, firstServer.UpdatedAt, "Server updated_at should not be nil")
	})

	t.Run("GetServerByUUID", func(t *testing.T) {
		// Get a specific server by UUID (from fixtures)
		server, err := env.Client.Servers.Get(env.Ctx, "server-001")

		require.NoError(t, err, "Failed to get server")
		require.NotNil(t, server, "Server should not be nil")

		// Verify server details
		assert.Equal(t, "server-001", server.ServerUUID)
		assert.Equal(t, "web-server-01", server.Hostname)
		assert.Equal(t, "192.168.1.100", server.MainIP)
		assert.Equal(t, "US-East", server.Location)
		assert.Equal(t, "production", server.Environment)
		assert.Equal(t, "active", server.Status)
	})

	t.Run("GetServerNotFound", func(t *testing.T) {
		// Try to get a non-existent server
		server, err := env.Client.Servers.Get(env.Ctx, "non-existent-uuid")

		require.Error(t, err, "Should return error for non-existent server")
		assert.Nil(t, server, "Server should be nil for not found")

		// Verify error is NotFoundError
		_, isNotFound := err.(*nexmonyx.NotFoundError)
		assert.True(t, isNotFound, "Error should be NotFoundError")
	})

	t.Run("CreateServer", func(t *testing.T) {
		// Create a new server
		newServer := &nexmonyx.Server{
			Hostname:       "test-server-new",
			OrganizationID: 1,
			MainIP:         "192.168.1.250",
			Location:       "Test-Region",
			Environment:    "testing",
			Classification: "test",
		}

		server, err := env.Client.Servers.Create(env.Ctx, newServer)

		require.NoError(t, err, "Failed to create server")
		require.NotNil(t, server, "Created server should not be nil")

		// Verify created server has correct fields
		assert.NotEmpty(t, server.ServerUUID, "Server UUID should not be empty")
		assert.Equal(t, newServer.Hostname, server.Hostname)
		assert.Equal(t, newServer.MainIP, server.MainIP)
		assert.Equal(t, newServer.Location, server.Location)
		assert.Equal(t, newServer.Environment, server.Environment)
		assert.Equal(t, "active", server.Status, "New server should be active")
		// Note: CreatedAt and UpdatedAt are *CustomTime, not time.Time
		assert.NotNil(t, server.CreatedAt, "Server created_at should not be nil")
		assert.NotNil(t, server.UpdatedAt, "Server updated_at should not be nil")

		t.Logf("Created server: %s (UUID: %s)", server.Hostname, server.ServerUUID)
	})

	t.Run("UpdateServer", func(t *testing.T) {
		// First, create a server to update
		created := createTestServer(t, env, "test-server-to-update")

		// Update the server
		updateReq := &nexmonyx.Server{
			ServerUUID:  created.ServerUUID,
			Hostname:    "test-server-updated",
			Location:    "Updated-Region",
			Environment: "staging",
		}

		updated, err := env.Client.Servers.Update(env.Ctx, created.ServerUUID, updateReq)

		require.NoError(t, err, "Failed to update server")
		require.NotNil(t, updated, "Updated server should not be nil")

		// Verify updated fields
		assert.Equal(t, created.ServerUUID, updated.ServerUUID, "ServerUUID should not change")
		assert.Equal(t, "test-server-updated", updated.Hostname)
		assert.Equal(t, "Updated-Region", updated.Location)
		assert.Equal(t, "staging", updated.Environment)
		// Note: Cannot directly compare *CustomTime with After - just verify not nil
		assert.NotNil(t, updated.UpdatedAt, "UpdatedAt should not be nil")
		assert.NotNil(t, updated.CreatedAt, "CreatedAt should not be nil")

		t.Logf("Updated server: %s", updated.ServerUUID)
	})

	t.Run("DeleteServer", func(t *testing.T) {
		// First, create a server to delete
		created := createTestServer(t, env, "test-server-to-delete")

		// Delete the server
		err := env.Client.Servers.Delete(env.Ctx, created.ServerUUID)
		require.NoError(t, err, "Failed to delete server")

		// Verify server is deleted by trying to get it
		deleted, err := env.Client.Servers.Get(env.Ctx, created.ServerUUID)
		require.Error(t, err, "Should return error when getting deleted server")
		assert.Nil(t, deleted, "Deleted server should be nil")

		t.Logf("Deleted server: %s", created.ServerUUID)
	})

	t.Run("SearchServers", func(t *testing.T) {
		// Search for servers with a query
		servers, meta, err := env.Client.Servers.List(env.Ctx, &nexmonyx.ListOptions{
			Page:   1,
			Limit:  25,
			Search: "web-server",
		})

		require.NoError(t, err, "Failed to search servers")
		require.NotNil(t, servers, "Servers list should not be nil")

		// Verify search results
		if len(servers) > 0 {
			assertPaginationValid(t, meta)

			// Verify at least one result contains the search term
			found := false
			for _, server := range servers {
				if containsIgnoreCase(server.Hostname, "web-server") {
					found = true
					break
				}
			}
			assert.True(t, found, "At least one server should match search term")
		}
	})

	t.Run("PaginationTest", func(t *testing.T) {
		// Test pagination by requesting different pages
		page1, meta1, err := env.Client.Servers.List(env.Ctx, &nexmonyx.ListOptions{
			Page:  1,
			Limit: 2,
		})
		require.NoError(t, err, "Failed to list servers page 1")
		require.NotNil(t, page1, "Page 1 should not be nil")

		assertPaginationValid(t, meta1)
		assert.LessOrEqual(t, len(page1), 2, "Page 1 should have at most 2 servers")

		// If there are more pages, fetch page 2
		if meta1.TotalPages > 1 {
			page2, meta2, err := env.Client.Servers.List(env.Ctx, &nexmonyx.ListOptions{
				Page:  2,
				Limit: 2,
			})
			require.NoError(t, err, "Failed to list servers page 2")
			require.NotNil(t, page2, "Page 2 should not be nil")

			assertPaginationValid(t, meta2)
			assert.Equal(t, 2, meta2.Page, "Should be on page 2")

			// Verify pages have different servers
			if len(page1) > 0 && len(page2) > 0 {
				assert.NotEqual(t, page1[0].ServerUUID, page2[0].ServerUUID,
					"Page 1 and Page 2 should have different servers")
			}
		}
	})
}

// TestServersAuthentication tests authentication requirements
func TestServersAuthentication(t *testing.T) {
	skipIfShort(t)

	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("UnauthenticatedRequest", func(t *testing.T) {
		// Create a client without authentication
		unauthClient, err := nexmonyx.NewClient(&nexmonyx.Config{
			BaseURL: env.BaseURL,
			// No auth config
		})
		require.NoError(t, err, "Failed to create unauthenticated client")

		// Try to list servers without authentication
		servers, _, err := unauthClient.Servers.List(env.Ctx, &nexmonyx.ListOptions{
			Page:  1,
			Limit: 25,
		})

		// Should return an error (unauthorized)
		require.Error(t, err, "Unauthenticated request should fail")
		assert.Nil(t, servers, "Servers should be nil for unauthenticated request")

		// Verify error is UnauthorizedError
		_, isUnauthorized := err.(*nexmonyx.UnauthorizedError)
		assert.True(t, isUnauthorized, "Error should be UnauthorizedError")
	})

	t.Run("InvalidToken", func(t *testing.T) {
		// Create a client with invalid token
		invalidClient, err := nexmonyx.NewClient(&nexmonyx.Config{
			BaseURL: env.BaseURL,
			Auth: nexmonyx.AuthConfig{
				Token: "invalid-token",
			},
		})
		require.NoError(t, err, "Failed to create client with invalid token")

		// Try to list servers with invalid token
		servers, _, err := invalidClient.Servers.List(env.Ctx, &nexmonyx.ListOptions{
			Page:  1,
			Limit: 25,
		})

		// Should return an error (unauthorized)
		require.Error(t, err, "Invalid token request should fail")
		assert.Nil(t, servers, "Servers should be nil for invalid token")
	})
}

// Helper function for case-insensitive contains check
func containsIgnoreCase(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	return contains(s, substr)
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexOfSubstring(s, substr) >= 0
}

func indexOfSubstring(s, substr string) int {
	n := len(substr)
	for i := 0; i <= len(s)-n; i++ {
		if s[i:i+n] == substr {
			return i
		}
	}
	return -1
}
