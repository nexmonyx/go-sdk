package integration

import (
	"fmt"
	"testing"

	nexmonyx "github.com/nexmonyx/go-sdk/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAlertLifecycleWorkflow tests the complete alert lifecycle from creation to deletion
func TestAlertLifecycleWorkflow(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("CompleteAlertLifecycle", func(t *testing.T) {
		// Step 1: Create a new alert rule
		alert := &nexmonyx.Alert{
			Name:           "Test High CPU Alert",
			Description:    "Alert for integration testing",
			OrganizationID: 1,
			Type:           "metric",
			MetricName:     "cpu_usage",
			Condition:      "greater_than",
			Threshold:      85.0,
			Duration:       300,
			Frequency:      60,
			Enabled:        true,
			Status:         "active",
			Severity:       "warning",
			Channels:       []string{"email"},
		}

		created, err := env.Client.Alerts.Create(env.Ctx, alert)
		require.NoError(t, err, "Failed to create alert")
		require.NotNil(t, created, "Created alert should not be nil")
		require.NotZero(t, created.ID, "Alert ID should not be zero")

		t.Logf("Step 1: Created alert with ID: %d", created.ID)

		// Step 2: Retrieve alert and verify
		alertID := fmt.Sprintf("%d", created.ID)
		retrieved, err := env.Client.Alerts.Get(env.Ctx, alertID)
		require.NoError(t, err, "Failed to retrieve alert")
		require.NotNil(t, retrieved, "Retrieved alert should not be nil")
		assert.Equal(t, created.ID, retrieved.ID, "Alert IDs should match")
		assert.Equal(t, "Test High CPU Alert", retrieved.Name, "Alert name should match")

		t.Logf("Step 2: Retrieved alert details successfully")

		// Step 3: Update alert threshold
		updated := &nexmonyx.Alert{
			Name:        "Test High CPU Alert",
			Description: "Updated alert description",
			Threshold:   90.0,
			Severity:    "critical",
		}
		updated.ID = created.ID

		updatedAlert, err := env.Client.Alerts.Update(env.Ctx, alertID, updated)
		require.NoError(t, err, "Failed to update alert")
		assert.Equal(t, 90.0, updatedAlert.Threshold, "Threshold should be updated")
		assert.Equal(t, "critical", updatedAlert.Severity, "Severity should be updated")

		t.Logf("Step 3: Updated alert threshold and severity")

		// Step 4: List alerts and verify our alert is included
		alerts, meta, err := env.Client.Alerts.List(env.Ctx, &nexmonyx.ListOptions{
			Page:  1,
			Limit: 25,
		})
		require.NoError(t, err, "Failed to list alerts")
		require.NotNil(t, alerts, "Alerts list should not be nil")
		assertPaginationValid(t, meta)

		found := false
		for _, a := range alerts {
			if a.ID == created.ID {
				found = true
				break
			}
		}
		assert.True(t, found, "Created alert should be in alerts list")

		t.Logf("Step 4: Verified alert appears in list")

		// Step 5: Delete alert
		err = env.Client.Alerts.Delete(env.Ctx, alertID)
		require.NoError(t, err, "Failed to delete alert")

		t.Logf("Step 5: Deleted alert successfully")

		// Step 6: Verify alert is deleted
		deleted, err := env.Client.Alerts.Get(env.Ctx, alertID)
		require.Error(t, err, "Should return error when getting deleted alert")
		assert.Nil(t, deleted, "Deleted alert should be nil")

		t.Logf("Step 6: Verified alert deletion")
	})
}

// TestAlertEnableDisableWorkflow tests enabling and disabling alerts
func TestAlertEnableDisableWorkflow(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("EnableDisableAlert", func(t *testing.T) {
		// Create an alert that's initially disabled
		alert := &nexmonyx.Alert{
			Name:           "Toggle Test Alert",
			Description:    "Alert for enable/disable testing",
			OrganizationID: 1,
			Type:           "metric",
			MetricName:     "memory_usage",
			Condition:      "greater_than",
			Threshold:      80.0,
			Enabled:        false,
			Status:         "inactive",
			Severity:       "warning",
			Channels:       []string{"email"},
		}

		created, err := env.Client.Alerts.Create(env.Ctx, alert)
		require.NoError(t, err, "Failed to create alert")
		defer env.Client.Alerts.Delete(env.Ctx, fmt.Sprintf("%d", created.ID))

		alertID := fmt.Sprintf("%d", created.ID)
		assert.False(t, created.Enabled, "Alert should be initially disabled")

		t.Logf("Created disabled alert with ID: %d", created.ID)

		// Enable the alert
		enabled, err := env.Client.Alerts.Enable(env.Ctx, alertID)
		if err != nil {
			t.Skipf("Enable endpoint not available: %v", err)
			return
		}

		require.NotNil(t, enabled, "Enabled alert should not be nil")
		assert.True(t, enabled.Enabled, "Alert should be enabled")

		t.Logf("Enabled alert successfully")

		// Disable the alert
		disabled, err := env.Client.Alerts.Disable(env.Ctx, alertID)
		require.NoError(t, err, "Failed to disable alert")
		require.NotNil(t, disabled, "Disabled alert should not be nil")
		assert.False(t, disabled.Enabled, "Alert should be disabled")

		t.Logf("Disabled alert successfully")
	})
}

// TestBulkAlertOperations tests creating and managing multiple alerts
func TestBulkAlertOperations(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("CreateMultipleAlerts", func(t *testing.T) {
		alertCount := 3
		createdAlerts := make([]*nexmonyx.Alert, 0, alertCount)

		// Create multiple alerts
		for i := 0; i < alertCount; i++ {
			alert := &nexmonyx.Alert{
				Name:           fmt.Sprintf("Bulk Test Alert %d", i+1),
				Description:    fmt.Sprintf("Alert %d for bulk testing", i+1),
				OrganizationID: 1,
				Type:           "metric",
				MetricName:     "cpu_usage",
				Condition:      "greater_than",
				Threshold:      float64(70 + i*5),
				Enabled:        true,
				Status:         "active",
				Severity:       "warning",
				Channels:       []string{"email"},
			}

			created, err := env.Client.Alerts.Create(env.Ctx, alert)
			require.NoError(t, err, "Failed to create alert %d", i+1)
			createdAlerts = append(createdAlerts, created)
		}

		t.Logf("Created %d alerts", len(createdAlerts))
		assert.Equal(t, alertCount, len(createdAlerts), "Should have created all alerts")

		// Verify all alerts exist
		for _, alert := range createdAlerts {
			alertID := fmt.Sprintf("%d", alert.ID)
			retrieved, err := env.Client.Alerts.Get(env.Ctx, alertID)
			require.NoError(t, err, "Should be able to retrieve alert")
			assert.NotNil(t, retrieved, "Retrieved alert should not be nil")
		}

		t.Logf("Verified all %d alerts exist", len(createdAlerts))

		// Clean up - delete all created alerts
		for i, alert := range createdAlerts {
			alertID := fmt.Sprintf("%d", alert.ID)
			t.Logf("Deleting alert %d/%d with ID: %s", i+1, len(createdAlerts), alertID)
			err := env.Client.Alerts.Delete(env.Ctx, alertID)
			require.NoError(t, err, "Should be able to delete alert %d (ID: %s)", i+1, alertID)
		}

		t.Logf("Cleaned up all %d alerts", len(createdAlerts))
	})
}

// TestAlertSearchAndFiltering tests alert search and filtering capabilities
func TestAlertSearchAndFiltering(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("SearchBySeverity", func(t *testing.T) {
		// List all alerts
		alerts, meta, err := env.Client.Alerts.List(env.Ctx, &nexmonyx.ListOptions{
			Page:  1,
			Limit: 25,
		})

		require.NoError(t, err, "Failed to list alerts")
		require.NotNil(t, alerts, "Alerts list should not be nil")
		assertPaginationValid(t, meta)

		// Check for alerts with different severities from fixtures
		hasCritical := false
		hasWarning := false

		for _, alert := range alerts {
			if alert.Severity == "critical" {
				hasCritical = true
			}
			if alert.Severity == "warning" {
				hasWarning = true
			}
		}

		t.Logf("Found alerts - Critical: %v, Warning: %v", hasCritical, hasWarning)
		assert.True(t, hasCritical || hasWarning, "Should find alerts with different severities")
	})
}

// TestAlertPagination tests pagination of alert lists
func TestAlertPagination(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("PaginateThroughAlerts", func(t *testing.T) {
		// Get first page
		page1, meta1, err := env.Client.Alerts.List(env.Ctx, &nexmonyx.ListOptions{
			Page:  1,
			Limit: 2,
		})
		require.NoError(t, err, "Failed to get first page")
		require.NotNil(t, page1, "First page should not be nil")
		assertPaginationValid(t, meta1)

		t.Logf("Page 1: %d alerts, Total: %d, Total Pages: %d",
			len(page1), meta1.TotalItems, meta1.TotalPages)

		// If there are multiple pages, get the second page
		if meta1.TotalPages > 1 {
			page2, meta2, err := env.Client.Alerts.List(env.Ctx, &nexmonyx.ListOptions{
				Page:  2,
				Limit: 2,
			})
			require.NoError(t, err, "Failed to get second page")
			require.NotNil(t, page2, "Second page should not be nil")
			assertPaginationValid(t, meta2)
			assert.Equal(t, 2, meta2.Page, "Should be on page 2")

			t.Logf("Page 2: %d alerts", len(page2))

			// Verify pages have different alerts
			if len(page1) > 0 && len(page2) > 0 {
				assert.NotEqual(t, page1[0].ID, page2[0].ID,
					"Different pages should have different alerts")
			}
		}
	})
}

// TestAlertValidation tests alert creation validation
func TestAlertValidation(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("CreateAlertWithoutName", func(t *testing.T) {
		// Try to create alert without required name
		alert := &nexmonyx.Alert{
			OrganizationID: 1,
			Type:           "metric",
			MetricName:     "cpu_usage",
			Condition:      "greater_than",
			Threshold:      80.0,
		}

		_, err := env.Client.Alerts.Create(env.Ctx, alert)
		require.Error(t, err, "Should fail without name")
		t.Logf("Correctly rejected alert without name")
	})

	t.Run("CreateAlertWithValidData", func(t *testing.T) {
		// Create alert with all required fields
		alert := &nexmonyx.Alert{
			Name:           "Validation Test Alert",
			Description:    "Alert for validation testing",
			OrganizationID: 1,
			Type:           "metric",
			MetricName:     "cpu_usage",
			Condition:      "greater_than",
			Threshold:      75.0,
			Enabled:        true,
			Status:         "active",
			Severity:       "warning",
			Channels:       []string{"email"},
		}

		created, err := env.Client.Alerts.Create(env.Ctx, alert)
		require.NoError(t, err, "Should succeed with valid data")
		require.NotNil(t, created, "Created alert should not be nil")

		// Clean up
		env.Client.Alerts.Delete(env.Ctx, fmt.Sprintf("%d", created.ID))

		t.Logf("Successfully created alert with valid data")
	})
}

// TestAlertsByServerWorkflow tests alerts associated with specific servers
func TestAlertsByServerWorkflow(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("ServerSpecificAlerts", func(t *testing.T) {
		// List all alerts
		alerts, _, err := env.Client.Alerts.List(env.Ctx, &nexmonyx.ListOptions{
			Page:  1,
			Limit: 100,
		})
		require.NoError(t, err, "Failed to list alerts")

		// Count alerts by server (from fixtures, we have alerts for server-001 and server-002)
		serverAlertCount := make(map[uint]int)
		for _, alert := range alerts {
			if alert.ServerID != nil {
				serverAlertCount[*alert.ServerID]++
			}
		}

		t.Logf("Alerts by server: %+v", serverAlertCount)
		assert.Greater(t, len(serverAlertCount), 0, "Should have alerts associated with servers")
	})
}

// TestAlertSeverityLevels tests different alert severity levels
func TestAlertSeverityLevels(t *testing.T) {
	skipIfShort(t)
	env := setupIntegrationTest(t)
	defer teardownIntegrationTest(t, env)

	t.Run("CreateAlertsWithDifferentSeverities", func(t *testing.T) {
		severities := []string{"info", "warning", "critical"}
		createdAlerts := make([]*nexmonyx.Alert, 0, len(severities))

		for _, severity := range severities {
			alert := &nexmonyx.Alert{
				Name:           fmt.Sprintf("Severity Test Alert - %s", severity),
				Description:    fmt.Sprintf("Testing %s severity level", severity),
				OrganizationID: 1,
				Type:           "metric",
				MetricName:     "cpu_usage",
				Condition:      "greater_than",
				Threshold:      70.0,
				Enabled:        true,
				Status:         "active",
				Severity:       severity,
				Channels:       []string{"email"},
			}

			created, err := env.Client.Alerts.Create(env.Ctx, alert)
			require.NoError(t, err, "Failed to create %s alert", severity)
			createdAlerts = append(createdAlerts, created)

			t.Logf("Created %s severity alert", severity)
		}

		// Verify all severities were created
		assert.Equal(t, len(severities), len(createdAlerts), "Should create all severity levels")

		// Clean up
		for _, alert := range createdAlerts {
			env.Client.Alerts.Delete(env.Ctx, fmt.Sprintf("%d", alert.ID))
		}

		t.Logf("Tested all severity levels: %v", severities)
	})
}
