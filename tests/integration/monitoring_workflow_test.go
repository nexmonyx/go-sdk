package integration

import (
	"fmt"
	"testing"

	nexmonyx "github.com/nexmonyx/go-sdk/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProbeLifecycleWorkflow tests the complete probe lifecycle from creation to deletion
func TestProbeLifecycleWorkflow(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("CompleteProbeLifecycle", func(t *testing.T) {
		// Step 1: Create a new probe
		probe := &nexmonyx.MonitoringProbe{
			Name:           "Test Website Monitor",
			Description:    "Probe for integration testing",
			Type:           "https",
			Target:         "https://test.example.com",
			Interval:       60,
			Timeout:        10,
			Enabled:        true,
			OrganizationID: 1,
			Regions:        []string{"us-east"},
			Config: map[string]interface{}{
				"method":          "GET",
				"expected_status": 200,
				"verify_ssl":      true,
			},
			AlertConfig: &nexmonyx.ProbeAlertConfig{
				Enabled:          true,
				FailureThreshold: 3,
				SuccessThreshold: 2,
				Channels:         []string{"email"},
			},
			Tags: []string{"test", "integration"},
		}

		created, err := env.Client.Monitoring.CreateProbe(env.Ctx, probe)
		require.NoError(t, err, "Failed to create probe")
		require.NotNil(t, created, "Created probe should not be nil")
		require.NotEmpty(t, created.ProbeUUID, "Probe UUID should not be empty")

		t.Logf("Step 1: Created probe with UUID: %s", created.ProbeUUID)

		// Step 2: Retrieve probe and verify
		retrieved, err := env.Client.Monitoring.GetProbe(env.Ctx, created.ProbeUUID)
		require.NoError(t, err, "Failed to retrieve probe")
		require.NotNil(t, retrieved, "Retrieved probe should not be nil")
		assert.Equal(t, created.ProbeUUID, retrieved.ProbeUUID, "Probe UUIDs should match")
		assert.Equal(t, "Test Website Monitor", retrieved.Name, "Probe name should match")

		t.Logf("Step 2: Retrieved probe details successfully")

		// Step 3: Update probe settings
		updated := &nexmonyx.MonitoringProbe{
			Name:        "Test Website Monitor",
			Description: "Updated probe description",
			Interval:    30,
			Timeout:     15,
			Target:      "https://updated.example.com",
		}
		updated.ID = created.ID

		updatedProbe, err := env.Client.Monitoring.UpdateProbe(env.Ctx, created.ProbeUUID, updated)
		require.NoError(t, err, "Failed to update probe")
		assert.Equal(t, 30, updatedProbe.Interval, "Interval should be updated")
		assert.Equal(t, "Updated probe description", updatedProbe.Description, "Description should be updated")

		t.Logf("Step 3: Updated probe settings")

		// Step 4: List probes and verify our probe is included
		probes, meta, err := env.Client.Monitoring.ListProbes(env.Ctx, &nexmonyx.ListOptions{
			Page:  1,
			Limit: 25,
		})
		require.NoError(t, err, "Failed to list probes")
		require.NotNil(t, probes, "Probes list should not be nil")
		assertPaginationValid(t, meta)

		found := false
		for _, p := range probes {
			if p.ProbeUUID == created.ProbeUUID {
				found = true
				break
			}
		}
		assert.True(t, found, "Created probe should be in probes list")

		t.Logf("Step 4: Verified probe appears in list")

		// Step 5: Delete probe
		err = env.Client.Monitoring.DeleteProbe(env.Ctx, created.ProbeUUID)
		require.NoError(t, err, "Failed to delete probe")

		t.Logf("Step 5: Deleted probe successfully")

		// Step 6: Verify probe is deleted
		deleted, err := env.Client.Monitoring.GetProbe(env.Ctx, created.ProbeUUID)
		require.Error(t, err, "Should return error when getting deleted probe")
		assert.Nil(t, deleted, "Deleted probe should be nil")

		t.Logf("Step 6: Verified probe deletion")
	})
}

// TestProbeTypesWorkflow tests different probe types
func TestProbeTypesWorkflow(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("CreateProbesWithDifferentTypes", func(t *testing.T) {
		probeTypes := []struct {
			name       string
			probeType  string
			target     string
			config     map[string]interface{}
		}{
			{
				name:      "HTTPS Health Check",
				probeType: "https",
				target:    "https://example.com",
				config: map[string]interface{}{
					"method":          "GET",
					"expected_status": 200,
				},
			},
			{
				name:      "TCP Port Check",
				probeType: "tcp",
				target:    "example.com:443",
				config: map[string]interface{}{
					"port": 443,
				},
			},
			{
				name:      "DNS Resolution Check",
				probeType: "dns",
				target:    "example.com",
				config: map[string]interface{}{
					"record_type": "A",
				},
			},
		}

		createdProbes := make([]*nexmonyx.MonitoringProbe, 0, len(probeTypes))

		for _, pt := range probeTypes {
			probe := &nexmonyx.MonitoringProbe{
				Name:           pt.name,
				Type:           pt.probeType,
				Target:         pt.target,
				Interval:       60,
				Timeout:        10,
				Enabled:        true,
				OrganizationID: 1,
				Config:         pt.config,
			}

			created, err := env.Client.Monitoring.CreateProbe(env.Ctx, probe)
			require.NoError(t, err, "Failed to create %s probe", pt.probeType)
			createdProbes = append(createdProbes, created)

			t.Logf("Created %s probe: %s", pt.probeType, created.ProbeUUID)
		}

		assert.Equal(t, len(probeTypes), len(createdProbes), "Should create all probe types")

		// Clean up
		for _, probe := range createdProbes {
			env.Client.Monitoring.DeleteProbe(env.Ctx, probe.ProbeUUID)
		}

		t.Logf("Tested %d probe types successfully", len(probeTypes))
	})
}

// TestBulkProbeOperations tests creating and managing multiple probes
func TestBulkProbeOperations(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("CreateMultipleProbes", func(t *testing.T) {
		probeCount := 3
		createdProbes := make([]*nexmonyx.MonitoringProbe, 0, probeCount)

		// Create multiple probes
		for i := 0; i < probeCount; i++ {
			probe := &nexmonyx.MonitoringProbe{
				Name:           fmt.Sprintf("Bulk Test Probe %d", i+1),
				Description:    fmt.Sprintf("Probe %d for bulk testing", i+1),
				Type:           "https",
				Target:         fmt.Sprintf("https://test%d.example.com", i+1),
				Interval:       60 + (i * 10),
				Timeout:        10,
				Enabled:        true,
				OrganizationID: 1,
				Regions:        []string{"us-east"},
			}

			created, err := env.Client.Monitoring.CreateProbe(env.Ctx, probe)
			require.NoError(t, err, "Failed to create probe %d", i+1)
			createdProbes = append(createdProbes, created)
		}

		t.Logf("Created %d probes", len(createdProbes))
		assert.Equal(t, probeCount, len(createdProbes), "Should have created all probes")

		// Verify all probes exist
		for _, probe := range createdProbes {
			retrieved, err := env.Client.Monitoring.GetProbe(env.Ctx, probe.ProbeUUID)
			require.NoError(t, err, "Should be able to retrieve probe")
			assert.NotNil(t, retrieved, "Retrieved probe should not be nil")
		}

		t.Logf("Verified all %d probes exist", len(createdProbes))

		// Clean up - delete all created probes
		for i, probe := range createdProbes {
			t.Logf("Deleting probe %d/%d with UUID: %s", i+1, len(createdProbes), probe.ProbeUUID)
			err := env.Client.Monitoring.DeleteProbe(env.Ctx, probe.ProbeUUID)
			require.NoError(t, err, "Should be able to delete probe %d (UUID: %s)", i+1, probe.ProbeUUID)
		}

		t.Logf("Cleaned up all %d probes", len(createdProbes))
	})
}

// TestProbeSearchAndFiltering tests probe search and filtering capabilities
func TestProbeSearchAndFiltering(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("SearchByType", func(t *testing.T) {
		// List all probes
		probes, meta, err := env.Client.Monitoring.ListProbes(env.Ctx, &nexmonyx.ListOptions{
			Page:  1,
			Limit: 25,
		})

		require.NoError(t, err, "Failed to list probes")
		require.NotNil(t, probes, "Probes list should not be nil")
		assertPaginationValid(t, meta)

		// Check for probes with different types from fixtures
		hasHTTPS := false
		hasTCP := false

		for _, probe := range probes {
			if probe.Type == "https" {
				hasHTTPS = true
			}
			if probe.Type == "tcp" {
				hasTCP = true
			}
		}

		t.Logf("Found probes - HTTPS: %v, TCP: %v", hasHTTPS, hasTCP)
		assert.True(t, hasHTTPS || hasTCP, "Should find probes with different types")
	})
}

// TestProbePagination tests pagination of probe lists
func TestProbePagination(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("PaginateThroughProbes", func(t *testing.T) {
		// Get first page
		page1, meta1, err := env.Client.Monitoring.ListProbes(env.Ctx, &nexmonyx.ListOptions{
			Page:  1,
			Limit: 2,
		})
		require.NoError(t, err, "Failed to get first page")
		require.NotNil(t, page1, "First page should not be nil")
		assertPaginationValid(t, meta1)

		t.Logf("Page 1: %d probes, Total: %d, Total Pages: %d",
			len(page1), meta1.TotalItems, meta1.TotalPages)

		// If there are multiple pages, get the second page
		if meta1.TotalPages > 1 {
			page2, meta2, err := env.Client.Monitoring.ListProbes(env.Ctx, &nexmonyx.ListOptions{
				Page:  2,
				Limit: 2,
			})
			require.NoError(t, err, "Failed to get second page")
			require.NotNil(t, page2, "Second page should not be nil")
			assertPaginationValid(t, meta2)
			assert.Equal(t, 2, meta2.Page, "Should be on page 2")

			t.Logf("Page 2: %d probes", len(page2))

			// Verify pages have different probes
			if len(page1) > 0 && len(page2) > 0 {
				assert.NotEqual(t, page1[0].ProbeUUID, page2[0].ProbeUUID,
					"Different pages should have different probes")
			}
		}
	})
}

// TestProbeValidation tests probe creation validation
func TestProbeValidation(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("CreateProbeWithoutName", func(t *testing.T) {
		// Try to create probe without required name
		probe := &nexmonyx.MonitoringProbe{
			Type:           "https",
			Target:         "https://example.com",
			Interval:       60,
			OrganizationID: 1,
		}

		_, err := env.Client.Monitoring.CreateProbe(env.Ctx, probe)
		require.Error(t, err, "Should fail without name")
		t.Logf("Correctly rejected probe without name")
	})

	t.Run("CreateProbeWithoutType", func(t *testing.T) {
		// Try to create probe without required type
		probe := &nexmonyx.MonitoringProbe{
			Name:           "Test Probe",
			Target:         "https://example.com",
			Interval:       60,
			OrganizationID: 1,
		}

		_, err := env.Client.Monitoring.CreateProbe(env.Ctx, probe)
		require.Error(t, err, "Should fail without type")
		t.Logf("Correctly rejected probe without type")
	})

	t.Run("CreateProbeWithoutTarget", func(t *testing.T) {
		// Try to create probe without required target
		probe := &nexmonyx.MonitoringProbe{
			Name:           "Test Probe",
			Type:           "https",
			Interval:       60,
			OrganizationID: 1,
		}

		_, err := env.Client.Monitoring.CreateProbe(env.Ctx, probe)
		require.Error(t, err, "Should fail without target")
		t.Logf("Correctly rejected probe without target")
	})

	t.Run("CreateProbeWithValidData", func(t *testing.T) {
		// Create probe with all required fields
		probe := &nexmonyx.MonitoringProbe{
			Name:           "Validation Test Probe",
			Description:    "Probe for validation testing",
			Type:           "https",
			Target:         "https://valid.example.com",
			Interval:       60,
			Timeout:        10,
			Enabled:        true,
			OrganizationID: 1,
		}

		created, err := env.Client.Monitoring.CreateProbe(env.Ctx, probe)
		require.NoError(t, err, "Should succeed with valid data")
		require.NotNil(t, created, "Created probe should not be nil")

		// Clean up
		env.Client.Monitoring.DeleteProbe(env.Ctx, created.ProbeUUID)

		t.Logf("Successfully created probe with valid data")
	})
}

// TestProbeRegionsWorkflow tests probes with regional execution
func TestProbeRegionsWorkflow(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("ProbeWithMultipleRegions", func(t *testing.T) {
		probe := &nexmonyx.MonitoringProbe{
			Name:           "Multi-Region Probe",
			Description:    "Probe executing from multiple regions",
			Type:           "https",
			Target:         "https://global.example.com",
			Interval:       60,
			Timeout:        10,
			Enabled:        true,
			OrganizationID: 1,
			Regions:        []string{"us-east", "us-west", "eu-west"},
		}

		created, err := env.Client.Monitoring.CreateProbe(env.Ctx, probe)
		require.NoError(t, err, "Failed to create multi-region probe")
		defer env.Client.Monitoring.DeleteProbe(env.Ctx, created.ProbeUUID)

		assert.Equal(t, 3, len(created.Regions), "Should have 3 regions")
		assert.Contains(t, created.Regions, "us-east", "Should include us-east region")
		assert.Contains(t, created.Regions, "us-west", "Should include us-west region")
		assert.Contains(t, created.Regions, "eu-west", "Should include eu-west region")

		t.Logf("Created probe with %d regions: %v", len(created.Regions), created.Regions)
	})
}

// TestProbeAlertConfigWorkflow tests probe alert configuration
func TestProbeAlertConfigWorkflow(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("ProbeWithAlertConfig", func(t *testing.T) {
		probe := &nexmonyx.MonitoringProbe{
			Name:           "Alert-Enabled Probe",
			Type:           "https",
			Target:         "https://critical.example.com",
			Interval:       30,
			Timeout:        5,
			Enabled:        true,
			OrganizationID: 1,
			AlertConfig: &nexmonyx.ProbeAlertConfig{
				Enabled:           true,
				FailureThreshold:  2,
				SuccessThreshold:  1,
				NotificationDelay: 60,
				Channels:          []string{"email", "slack", "pagerduty"},
				Recipients:        []string{"ops@example.com"},
			},
		}

		created, err := env.Client.Monitoring.CreateProbe(env.Ctx, probe)
		require.NoError(t, err, "Failed to create probe with alert config")
		defer env.Client.Monitoring.DeleteProbe(env.Ctx, created.ProbeUUID)

		require.NotNil(t, created.AlertConfig, "Alert config should not be nil")
		assert.True(t, created.AlertConfig.Enabled, "Alerts should be enabled")
		assert.Equal(t, 2, created.AlertConfig.FailureThreshold, "Failure threshold should match")
		assert.Equal(t, 3, len(created.AlertConfig.Channels), "Should have 3 notification channels")

		t.Logf("Created probe with alert config - Threshold: %d, Channels: %v",
			created.AlertConfig.FailureThreshold, created.AlertConfig.Channels)
	})
}
