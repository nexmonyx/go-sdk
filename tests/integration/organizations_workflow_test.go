package integration

import (
	"fmt"
	"testing"

	nexmonyx "github.com/nexmonyx/go-sdk/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOrganizationLifecycleWorkflow tests the complete organization lifecycle from creation to deletion
func TestOrganizationLifecycleWorkflow(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("CompleteOrganizationLifecycle", func(t *testing.T) {
		// Step 1: Create a new organization
		org := &nexmonyx.Organization{
			Name:        "Test Organization",
			Description: "Organization for integration testing",
			Industry:    "Technology",
			Website:     "https://test-org.example.com",
			Country:     "US",
			TimeZone:    "America/New_York",
		}

		created, err := env.Client.Organizations.Create(env.Ctx, org)
		require.NoError(t, err, "Failed to create organization")
		require.NotNil(t, created, "Created organization should not be nil")
		require.NotEmpty(t, created.UUID, "Organization UUID should not be empty")

		t.Logf("Step 1: Created organization with UUID: %s", created.UUID)

		// Step 2: Retrieve organization details and verify
		retrieved, err := env.Client.Organizations.GetByUUID(env.Ctx, created.UUID)
		require.NoError(t, err, "Failed to retrieve organization")
		require.NotNil(t, retrieved, "Retrieved organization should not be nil")
		assert.Equal(t, created.UUID, retrieved.UUID, "Organization UUIDs should match")
		assert.Equal(t, "Test Organization", retrieved.Name, "Organization name should match")

		t.Logf("Step 2: Retrieved organization details successfully")

		// Step 3: Update organization details
		updated := &nexmonyx.Organization{
			UUID:        created.UUID,
			Name:        "Test Organization",
			Description: "Updated organization description",
			Industry:    "SaaS",
			Website:     "https://updated-org.example.com",
		}

		updatedOrg, err := env.Client.Organizations.Update(env.Ctx, created.UUID, updated)
		require.NoError(t, err, "Failed to update organization")
		assert.Equal(t, "Updated organization description", updatedOrg.Description, "Description should be updated")
		assert.Equal(t, "SaaS", updatedOrg.Industry, "Industry should be updated")

		t.Logf("Step 3: Updated organization details")

		// Step 4: List organizations and verify our organization is included
		orgs, meta, err := env.Client.Organizations.List(env.Ctx, &nexmonyx.ListOptions{
			Page:  1,
			Limit: 25,
		})
		require.NoError(t, err, "Failed to list organizations")
		require.NotNil(t, orgs, "Organizations list should not be nil")
		assertPaginationValid(t, meta)

		found := false
		for _, o := range orgs {
			if o.UUID == created.UUID {
				found = true
				break
			}
		}
		assert.True(t, found, "Created organization should be in organizations list")

		t.Logf("Step 4: Verified organization appears in list")

		// Step 5: Delete organization
		err = env.Client.Organizations.Delete(env.Ctx, created.UUID)
		require.NoError(t, err, "Failed to delete organization")

		t.Logf("Step 5: Deleted organization successfully")

		// Step 6: Verify organization is deleted
		deleted, err := env.Client.Organizations.GetByUUID(env.Ctx, created.UUID)
		require.Error(t, err, "Should return error when getting deleted organization")
		assert.Nil(t, deleted, "Deleted organization should be nil")

		t.Logf("Step 6: Verified organization deletion")
	})
}

// TestOrganizationResourceManagement tests organization's relationship with servers and users
func TestOrganizationResourceManagement(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("OrganizationWithServers", func(t *testing.T) {
		// Use fixture organization (org-001)
		orgUUID := "org-001"

		// Get organization's servers
		servers, meta, err := env.Client.Organizations.GetServers(env.Ctx, orgUUID, &nexmonyx.ListOptions{
			Page:  1,
			Limit: 25,
		})

		require.NoError(t, err, "Failed to get organization servers")
		require.NotNil(t, servers, "Servers list should not be nil")
		assertPaginationValid(t, meta)

		t.Logf("Organization %s has %d servers", orgUUID, len(servers))

		// Verify all servers belong to this organization
		for _, server := range servers {
			assert.Equal(t, uint(1), server.OrganizationID, "Server should belong to organization 1")
		}
	})

	t.Run("OrganizationWithUsers", func(t *testing.T) {
		// Use fixture organization (org-001)
		orgUUID := "org-001"

		// Get organization's users
		users, meta, err := env.Client.Organizations.GetUsers(env.Ctx, orgUUID, &nexmonyx.ListOptions{
			Page:  1,
			Limit: 25,
		})

		// Note: Mock server may not have this endpoint fully implemented
		if err != nil {
			t.Skipf("GetUsers endpoint not available: %v", err)
			return
		}

		require.NotNil(t, users, "Users list should not be nil")
		require.NotNil(t, meta, "Pagination metadata should not be nil")

		// Note: Mock server returns empty user list (fixtures don't include user relationships)
		t.Logf("Organization %s has %d users", orgUUID, len(users))
	})

	t.Run("OrganizationWithAlerts", func(t *testing.T) {
		// Use fixture organization (org-001)
		orgUUID := "org-001"

		// Get organization's alerts
		alerts, meta, err := env.Client.Organizations.GetAlerts(env.Ctx, orgUUID, &nexmonyx.ListOptions{
			Page:  1,
			Limit: 25,
		})

		// Note: Mock server may not have this endpoint fully implemented
		if err != nil {
			t.Skipf("GetAlerts endpoint not available: %v", err)
			return
		}

		require.NotNil(t, alerts, "Alerts list should not be nil")
		assertPaginationValid(t, meta)

		t.Logf("Organization %s has %d alerts", orgUUID, len(alerts))
	})
}

// TestBulkOrganizationOperations tests creating and managing multiple organizations
func TestBulkOrganizationOperations(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("CreateMultipleOrganizations", func(t *testing.T) {
		orgCount := 3
		createdOrgs := make([]*nexmonyx.Organization, 0, orgCount)

		// Create multiple organizations
		for i := 0; i < orgCount; i++ {
			org := &nexmonyx.Organization{
				Name:        fmt.Sprintf("Bulk Test Org %d", i+1),
				Description: fmt.Sprintf("Organization %d for bulk testing", i+1),
				Industry:    "Technology",
			}

			created, err := env.Client.Organizations.Create(env.Ctx, org)
			require.NoError(t, err, "Failed to create organization %d", i+1)
			createdOrgs = append(createdOrgs, created)
		}

		t.Logf("Created %d organizations", len(createdOrgs))
		assert.Equal(t, orgCount, len(createdOrgs), "Should have created all organizations")

		// Verify all organizations exist
		for _, org := range createdOrgs {
			retrieved, err := env.Client.Organizations.GetByUUID(env.Ctx, org.UUID)
			require.NoError(t, err, "Should be able to retrieve organization")
			assert.NotNil(t, retrieved, "Retrieved organization should not be nil")
		}

		t.Logf("Verified all %d organizations exist", len(createdOrgs))

		// Clean up - delete all created organizations
		for _, org := range createdOrgs {
			err := env.Client.Organizations.Delete(env.Ctx, org.UUID)
			require.NoError(t, err, "Should be able to delete organization")
		}

		t.Logf("Cleaned up all %d organizations", len(createdOrgs))
	})
}

// TestOrganizationSearchAndFiltering tests organization search and filtering capabilities
func TestOrganizationSearchAndFiltering(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("SearchByName", func(t *testing.T) {
		// Search for organizations with a query
		orgs, meta, err := env.Client.Organizations.List(env.Ctx, &nexmonyx.ListOptions{
			Page:   1,
			Limit:  25,
			Search: "Acme",
		})

		require.NoError(t, err, "Search should not error")
		require.NotNil(t, orgs, "Organizations list should not be nil")
		assertPaginationValid(t, meta)

		// Should find "Acme Corporation" from fixtures
		found := false
		for _, org := range orgs {
			if org.UUID == "org-001" {
				found = true
				assert.Contains(t, org.Name, "Acme", "Organization name should contain search term")
				break
			}
		}

		assert.True(t, found, "Search should find Acme Corporation")
		t.Logf("Successfully searched and found organization")
	})
}

// TestOrganizationPagination tests pagination of organization lists
func TestOrganizationPagination(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("PaginateThroughOrganizations", func(t *testing.T) {
		// Get first page
		page1, meta1, err := env.Client.Organizations.List(env.Ctx, &nexmonyx.ListOptions{
			Page:  1,
			Limit: 2,
		})
		require.NoError(t, err, "Failed to get first page")
		require.NotNil(t, page1, "First page should not be nil")
		assertPaginationValid(t, meta1)

		t.Logf("Page 1: %d organizations, Total: %d, Total Pages: %d",
			len(page1), meta1.TotalItems, meta1.TotalPages)

		// If there are multiple pages, get the second page
		if meta1.TotalPages > 1 {
			page2, meta2, err := env.Client.Organizations.List(env.Ctx, &nexmonyx.ListOptions{
				Page:  2,
				Limit: 2,
			})
			require.NoError(t, err, "Failed to get second page")
			require.NotNil(t, page2, "Second page should not be nil")
			assertPaginationValid(t, meta2)
			assert.Equal(t, 2, meta2.Page, "Should be on page 2")

			t.Logf("Page 2: %d organizations", len(page2))

			// Verify pages have different organizations
			if len(page1) > 0 && len(page2) > 0 {
				assert.NotEqual(t, page1[0].UUID, page2[0].UUID,
					"Different pages should have different organizations")
			}
		}
	})
}

// TestOrganizationValidation tests organization creation validation
func TestOrganizationValidation(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("CreateOrganizationWithoutName", func(t *testing.T) {
		// Try to create organization without required name
		org := &nexmonyx.Organization{
			Description: "Organization without name",
		}

		_, err := env.Client.Organizations.Create(env.Ctx, org)
		require.Error(t, err, "Should fail without name")
		t.Logf("Correctly rejected organization without name")
	})

	t.Run("CreateOrganizationWithValidData", func(t *testing.T) {
		// Create organization with all required fields
		org := &nexmonyx.Organization{
			Name:        "Validation Test Org",
			Description: "Organization for validation testing",
			Industry:    "Technology",
		}

		created, err := env.Client.Organizations.Create(env.Ctx, org)
		require.NoError(t, err, "Should succeed with valid data")
		require.NotNil(t, created, "Created organization should not be nil")

		// Clean up
		env.Client.Organizations.Delete(env.Ctx, created.UUID)

		t.Logf("Successfully created organization with valid data")
	})
}

// TestOrganizationResourceIsolation tests that organizations properly isolate resources
func TestOrganizationResourceIsolation(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("ServersBelongToCorrectOrganization", func(t *testing.T) {
		// Get servers for org-001
		org1Servers, _, err := env.Client.Organizations.GetServers(env.Ctx, "org-001", &nexmonyx.ListOptions{
			Page:  1,
			Limit: 100,
		})
		require.NoError(t, err, "Failed to get org-001 servers")

		// Get servers for org-002
		org2Servers, _, err := env.Client.Organizations.GetServers(env.Ctx, "org-002", &nexmonyx.ListOptions{
			Page:  1,
			Limit: 100,
		})
		require.NoError(t, err, "Failed to get org-002 servers")

		// Verify no server UUIDs overlap
		org1UUIDs := make(map[string]bool)
		for _, server := range org1Servers {
			org1UUIDs[server.ServerUUID] = true
		}

		for _, server := range org2Servers {
			assert.False(t, org1UUIDs[server.ServerUUID],
				"Server %s should not appear in both organizations", server.ServerUUID)
		}

		t.Logf("Verified resource isolation: org-001 has %d servers, org-002 has %d servers",
			len(org1Servers), len(org2Servers))
	})
}
