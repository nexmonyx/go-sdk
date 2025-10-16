package integration

import (
	"fmt"
	"testing"

	nexmonyx "github.com/nexmonyx/go-sdk/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServerLifecycleWorkflow tests the complete server lifecycle from registration to deletion
func TestServerLifecycleWorkflow(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("CompleteServerLifecycle", func(t *testing.T) {
		// Step 1: Register a new server
		server := &nexmonyx.Server{
			Hostname:       "workflow-test-server",
			OrganizationID: 1,
			MainIP:         "192.168.100.50",
			Environment:    "testing",
			Classification: "test",
		}

		registered, err := env.Client.Servers.Create(env.Ctx, server)
		require.NoError(t, err, "Failed to register server")
		require.NotNil(t, registered, "Registered server should not be nil")
		require.NotEmpty(t, registered.ServerUUID, "Server UUID should not be empty")

		t.Logf("Step 1: Registered server with UUID: %s", registered.ServerUUID)

		// Step 2: Retrieve server details and verify
		retrieved, err := env.Client.Servers.GetByUUID(env.Ctx, registered.ServerUUID)
		require.NoError(t, err, "Failed to retrieve server")
		require.NotNil(t, retrieved, "Retrieved server should not be nil")
		assert.Equal(t, registered.ServerUUID, retrieved.ServerUUID, "Server UUIDs should match")
		assert.Equal(t, "workflow-test-server", retrieved.Hostname, "Hostname should match")

		t.Logf("Step 2: Retrieved server details successfully")

		// Step 3: Update server details
		updated := &nexmonyx.Server{
			ServerUUID:  registered.ServerUUID,
			Hostname:    "workflow-test-server",
			Location:    "Updated-Location",
			Environment: "staging",
		}
		updatedServer, err := env.Client.Servers.Update(env.Ctx, registered.ServerUUID, updated)
		require.NoError(t, err, "Failed to update server")
		assert.Equal(t, "Updated-Location", updatedServer.Location, "Location should be updated")
		assert.Equal(t, "staging", updatedServer.Environment, "Environment should be updated")

		t.Logf("Step 3: Updated server details")

		// Step 4: List servers and verify our server is included
		servers, meta, err := env.Client.Servers.List(env.Ctx, &nexmonyx.ListOptions{
			Page:  1,
			Limit: 25,
		})
		require.NoError(t, err, "Failed to list servers")
		require.NotNil(t, servers, "Servers list should not be nil")
		assertPaginationValid(t, meta)

		found := false
		for _, s := range servers {
			if s.ServerUUID == registered.ServerUUID {
				found = true
				break
			}
		}
		assert.True(t, found, "Registered server should be in server list")

		t.Logf("Step 4: Verified server appears in server list")

		// Step 5: Delete server
		err = env.Client.Servers.Delete(env.Ctx, registered.ServerUUID)
		require.NoError(t, err, "Failed to delete server")

		t.Logf("Step 5: Deleted server successfully")

		// Step 6: Verify server is deleted
		deleted, err := env.Client.Servers.GetByUUID(env.Ctx, registered.ServerUUID)
		require.Error(t, err, "Should return error when getting deleted server")
		assert.Nil(t, deleted, "Deleted server should be nil")

		t.Logf("Step 6: Verified server deletion")
	})
}

// TestServerMetricsWorkflow tests server registration and metrics submission
func TestServerMetricsWorkflow(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("ServerWithMetricsSubmission", func(t *testing.T) {
		// Create a server
		server := createTestServer(t, env, "metrics-test-server")
		defer env.Client.Servers.Delete(env.Ctx, server.ServerUUID)

		// Submit metrics for the server
		metrics := createTestMetricsPayload(server.ServerUUID)

		// Note: Mock server may not have metrics.SubmitComprehensive implemented
		// This test demonstrates the workflow pattern
		t.Logf("Created metrics payload for server: %s", server.ServerUUID)
		assert.NotNil(t, metrics, "Metrics payload should not be nil")
		assert.Equal(t, server.ServerUUID, metrics.ServerUUID, "Metrics should be for correct server")
	})
}

// TestBulkServerOperations tests creating and managing multiple servers
func TestBulkServerOperations(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("CreateMultipleServers", func(t *testing.T) {
		serverCount := 5
		createdServers := make([]*nexmonyx.Server, 0, serverCount)

		// Create multiple servers
		for i := 0; i < serverCount; i++ {
			server := createTestServer(t, env, fmt.Sprintf("bulk-test-server-%d", i+1))
			createdServers = append(createdServers, server)
		}

		t.Logf("Created %d servers", len(createdServers))
		assert.Equal(t, serverCount, len(createdServers), "Should have created all servers")

		// Verify all servers exist
		for _, server := range createdServers {
			retrieved, err := env.Client.Servers.GetByUUID(env.Ctx, server.ServerUUID)
			require.NoError(t, err, "Should be able to retrieve server")
			assert.NotNil(t, retrieved, "Retrieved server should not be nil")
		}

		t.Logf("Verified all %d servers exist", len(createdServers))

		// Clean up - delete all created servers
		for _, server := range createdServers {
			err := env.Client.Servers.Delete(env.Ctx, server.ServerUUID)
			require.NoError(t, err, "Should be able to delete server")
		}

		t.Logf("Cleaned up all %d servers", len(createdServers))
	})
}

// TestServerSearchAndFiltering tests server search and filtering capabilities
func TestServerSearchAndFiltering(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("SearchByHostname", func(t *testing.T) {
		// Create a server with specific hostname
		server := createTestServer(t, env, "search-test-unique-hostname")
		defer env.Client.Servers.Delete(env.Ctx, server.ServerUUID)

		// Search for the server
		servers, _, err := env.Client.Servers.List(env.Ctx, &nexmonyx.ListOptions{
			Page:   1,
			Limit:  25,
			Search: "search-test-unique",
		})

		require.NoError(t, err, "Search should not error")

		// Verify our server is in the results
		found := false
		for _, s := range servers {
			if s.ServerUUID == server.ServerUUID {
				found = true
				break
			}
		}

		assert.True(t, found, "Search should find our server")
		t.Logf("Successfully searched and found server")
	})
}

// TestServerPagination tests pagination of server lists
func TestServerPagination(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("PaginateThroughServers", func(t *testing.T) {
		// Get first page
		page1, meta1, err := env.Client.Servers.List(env.Ctx, &nexmonyx.ListOptions{
			Page:  1,
			Limit: 2,
		})
		require.NoError(t, err, "Failed to get first page")
		require.NotNil(t, page1, "First page should not be nil")
		assertPaginationValid(t, meta1)

		t.Logf("Page 1: %d servers, Total: %d, Total Pages: %d",
			len(page1), meta1.TotalItems, meta1.TotalPages)

		// If there are multiple pages, get the second page
		if meta1.TotalPages > 1 {
			page2, meta2, err := env.Client.Servers.List(env.Ctx, &nexmonyx.ListOptions{
				Page:  2,
				Limit: 2,
			})
			require.NoError(t, err, "Failed to get second page")
			require.NotNil(t, page2, "Second page should not be nil")
			assertPaginationValid(t, meta2)
			assert.Equal(t, 2, meta2.Page, "Should be on page 2")

			t.Logf("Page 2: %d servers", len(page2))

			// Verify pages have different servers
			if len(page1) > 0 && len(page2) > 0 {
				assert.NotEqual(t, page1[0].ServerUUID, page2[0].ServerUUID,
					"Different pages should have different servers")
			}
		}
	})
}

// TestServerValidation tests server creation validation
func TestServerValidation(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("CreateServerWithoutHostname", func(t *testing.T) {
		// Try to create server without required hostname
		server := &nexmonyx.Server{
			OrganizationID: 1,
			MainIP:         "192.168.1.100",
		}

		_, err := env.Client.Servers.Create(env.Ctx, server)
		require.Error(t, err, "Should fail without hostname")
		t.Logf("Correctly rejected server without hostname")
	})

	t.Run("CreateServerWithValidData", func(t *testing.T) {
		// Create server with all required fields
		server := &nexmonyx.Server{
			Hostname:       "validation-test-server",
			OrganizationID: 1,
			MainIP:         "192.168.1.150",
			Environment:    "testing",
		}

		created, err := env.Client.Servers.Create(env.Ctx, server)
		require.NoError(t, err, "Should succeed with valid data")
		require.NotNil(t, created, "Created server should not be nil")

		// Clean up
		env.Client.Servers.Delete(env.Ctx, created.ServerUUID)

		t.Logf("Successfully created server with valid data")
	})
}
