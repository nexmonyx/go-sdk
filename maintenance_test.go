package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Maintenance Window CRUD Operations Tests
// =============================================================================

func TestMaintenanceWindowsService_CreateMaintenanceWindow_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/maintenance-windows", r.URL.Path)

		var reqBody CreateMaintenanceWindowRequest
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		require.NoError(t, err)
		assert.Equal(t, "Server Maintenance", reqBody.Name)
		assert.Equal(t, "2024-01-15T02:00:00Z", reqBody.StartsAt)
		assert.Equal(t, "2024-01-15T04:00:00Z", reqBody.EndsAt)
		assert.True(t, reqBody.SuppressAlerts)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Maintenance window created successfully",
			"data": map[string]interface{}{
				"id":               1,
				"window_uuid":      "mw-uuid-123",
				"organization_id":  1,
				"name":             "Server Maintenance",
				"description":      "Monthly server maintenance",
				"starts_at":        "2024-01-15T02:00:00Z",
				"ends_at":          "2024-01-15T04:00:00Z",
				"duration":         120,
				"status":           "scheduled",
				"suppress_alerts":  true,
				"pause_monitoring": false,
				"notify_before":    30,
				"server_filter":    map[string]interface{}{"tags": []string{"production"}},
				"created_at":       "2024-01-01T00:00:00Z",
				"updated_at":       "2024-01-01T00:00:00Z",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	window, resp, err := client.MaintenanceWindows.CreateMaintenanceWindow(context.Background(), &CreateMaintenanceWindowRequest{
		Name:           "Server Maintenance",
		Description:    "Monthly server maintenance",
		StartsAt:       "2024-01-15T02:00:00Z",
		EndsAt:         "2024-01-15T04:00:00Z",
		SuppressAlerts: true,
		ServerFilter:   map[string]interface{}{"tags": []string{"production"}},
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, window)
	assert.Equal(t, uint(1), window.ID)
	assert.Equal(t, "Server Maintenance", window.Name)
	assert.Equal(t, MaintenanceWindowStatusScheduled, window.Status)
	assert.True(t, window.SuppressAlerts)
	assert.Equal(t, 120, window.Duration)
}

func TestMaintenanceWindowsService_CreateMaintenanceWindow_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/maintenance-windows", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Start time must be before end time",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	window, _, err := client.MaintenanceWindows.CreateMaintenanceWindow(context.Background(), &CreateMaintenanceWindowRequest{
		Name:         "Bad Window",
		StartsAt:     "2024-01-15T04:00:00Z",
		EndsAt:       "2024-01-15T02:00:00Z", // End before start
		ServerFilter: map[string]interface{}{},
	})

	assert.Error(t, err)
	assert.Nil(t, window)
}

func TestMaintenanceWindowsService_ListMaintenanceWindows_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/maintenance-windows", r.URL.Path)
		assert.Equal(t, "scheduled", r.URL.Query().Get("status"))
		assert.Equal(t, "1", r.URL.Query().Get("page"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Maintenance windows retrieved",
			"data": map[string]interface{}{
				"windows": []map[string]interface{}{
					{
						"id":               1,
						"window_uuid":      "mw-uuid-1",
						"organization_id":  1,
						"name":             "Window 1",
						"starts_at":        "2024-01-15T02:00:00Z",
						"ends_at":          "2024-01-15T04:00:00Z",
						"duration":         120,
						"status":           "scheduled",
						"suppress_alerts":  true,
						"pause_monitoring": false,
						"notify_before":    30,
						"server_filter":    map[string]interface{}{},
						"created_at":       "2024-01-01T00:00:00Z",
						"updated_at":       "2024-01-01T00:00:00Z",
					},
					{
						"id":               2,
						"window_uuid":      "mw-uuid-2",
						"organization_id":  1,
						"name":             "Window 2",
						"starts_at":        "2024-01-20T00:00:00Z",
						"ends_at":          "2024-01-20T02:00:00Z",
						"duration":         120,
						"status":           "scheduled",
						"suppress_alerts":  true,
						"pause_monitoring": false,
						"notify_before":    60,
						"server_filter":    map[string]interface{}{},
						"created_at":       "2024-01-01T00:00:00Z",
						"updated_at":       "2024-01-01T00:00:00Z",
					},
				},
				"total":       2,
				"page":        1,
				"limit":       10,
				"total_pages": 1,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, resp, err := client.MaintenanceWindows.ListMaintenanceWindows(context.Background(), &ListMaintenanceWindowsOptions{
		Page:   1,
		Status: "scheduled",
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, result)
	assert.Len(t, result.Windows, 2)
	assert.Equal(t, "Window 1", result.Windows[0].Name)
	assert.Equal(t, "Window 2", result.Windows[1].Name)
	assert.Equal(t, int64(2), result.Total)
}

func TestMaintenanceWindowsService_ListMaintenanceWindows_NilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/maintenance-windows", r.URL.Path)
		assert.Empty(t, r.URL.RawQuery)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Maintenance windows retrieved",
			"data": map[string]interface{}{
				"windows":     []map[string]interface{}{},
				"total":       0,
				"page":        1,
				"limit":       10,
				"total_pages": 0,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, resp, err := client.MaintenanceWindows.ListMaintenanceWindows(context.Background(), nil)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, result)
}

func TestMaintenanceWindowsService_ListMaintenanceWindows_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Database connection failed",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, _, err := client.MaintenanceWindows.ListMaintenanceWindows(context.Background(), nil)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestMaintenanceWindowsService_GetMaintenanceWindow_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/maintenance-windows/123", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Maintenance window retrieved",
			"data": map[string]interface{}{
				"id":               123,
				"window_uuid":      "mw-uuid-123",
				"organization_id":  1,
				"name":             "Server Maintenance",
				"description":      "Scheduled maintenance for database upgrades",
				"starts_at":        "2024-01-15T02:00:00Z",
				"ends_at":          "2024-01-15T04:00:00Z",
				"duration":         120,
				"status":           "scheduled",
				"suppress_alerts":  true,
				"pause_monitoring": true,
				"notify_before":    60,
				"server_filter":    map[string]interface{}{"server_ids": []int{1, 2, 3}},
				"actions": []map[string]interface{}{
					{"type": "notify", "config": map[string]interface{}{"channel": "slack"}, "order": 1},
				},
				"created_at":       "2024-01-01T00:00:00Z",
				"updated_at":       "2024-01-01T00:00:00Z",
				"created_by_id":    1,
				"created_by_email": "admin@example.com",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	window, resp, err := client.MaintenanceWindows.GetMaintenanceWindow(context.Background(), 123)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, window)
	assert.Equal(t, uint(123), window.ID)
	assert.Equal(t, "Server Maintenance", window.Name)
	assert.Equal(t, "Scheduled maintenance for database upgrades", window.Description)
	assert.True(t, window.SuppressAlerts)
	assert.True(t, window.PauseMonitoring)
	assert.Equal(t, 60, window.NotifyBefore)
}

func TestMaintenanceWindowsService_GetMaintenanceWindow_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/maintenance-windows/999", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Maintenance window not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	window, _, err := client.MaintenanceWindows.GetMaintenanceWindow(context.Background(), 999)

	assert.Error(t, err)
	assert.Nil(t, window)
}

func TestMaintenanceWindowsService_UpdateMaintenanceWindow_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/maintenance-windows/123", r.URL.Path)

		var reqBody UpdateMaintenanceWindowRequest
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		require.NoError(t, err)
		assert.NotNil(t, reqBody.Name)
		assert.Equal(t, "Updated Maintenance", *reqBody.Name)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Maintenance window updated",
			"data": map[string]interface{}{
				"id":               123,
				"window_uuid":      "mw-uuid-123",
				"organization_id":  1,
				"name":             "Updated Maintenance",
				"starts_at":        "2024-01-15T02:00:00Z",
				"ends_at":          "2024-01-15T04:00:00Z",
				"duration":         120,
				"status":           "scheduled",
				"suppress_alerts":  true,
				"pause_monitoring": false,
				"notify_before":    30,
				"server_filter":    map[string]interface{}{},
				"created_at":       "2024-01-01T00:00:00Z",
				"updated_at":       "2024-01-02T00:00:00Z",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	newName := "Updated Maintenance"
	window, resp, err := client.MaintenanceWindows.UpdateMaintenanceWindow(context.Background(), 123, &UpdateMaintenanceWindowRequest{
		Name: &newName,
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, window)
	assert.Equal(t, "Updated Maintenance", window.Name)
}

func TestMaintenanceWindowsService_UpdateMaintenanceWindow_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Cannot update active maintenance window",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	newName := "Failed Update"
	window, _, err := client.MaintenanceWindows.UpdateMaintenanceWindow(context.Background(), 123, &UpdateMaintenanceWindowRequest{
		Name: &newName,
	})

	assert.Error(t, err)
	assert.Nil(t, window)
}

func TestMaintenanceWindowsService_DeleteMaintenanceWindow_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/maintenance-windows/123", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Maintenance window deleted",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	resp, err := client.MaintenanceWindows.DeleteMaintenanceWindow(context.Background(), 123)

	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestMaintenanceWindowsService_DeleteMaintenanceWindow_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Maintenance window not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	_, err := client.MaintenanceWindows.DeleteMaintenanceWindow(context.Background(), 999)

	assert.Error(t, err)
}

// =============================================================================
// Maintenance Window Special Operations Tests
// =============================================================================

func TestMaintenanceWindowsService_GetActiveMaintenanceWindows_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/maintenance-windows/active", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Active maintenance windows retrieved",
			"data": map[string]interface{}{
				"windows": []map[string]interface{}{
					{
						"id":               1,
						"window_uuid":      "mw-uuid-1",
						"organization_id":  1,
						"name":             "Active Maintenance 1",
						"starts_at":        "2024-01-01T00:00:00Z",
						"ends_at":          "2024-01-01T02:00:00Z",
						"duration":         120,
						"status":           "active",
						"suppress_alerts":  true,
						"pause_monitoring": true,
						"notify_before":    30,
						"server_filter":    map[string]interface{}{},
						"activated_at":     "2024-01-01T00:00:00Z",
						"created_at":       "2023-12-31T00:00:00Z",
						"updated_at":       "2024-01-01T00:00:00Z",
					},
				},
				"count": 1,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, resp, err := client.MaintenanceWindows.GetActiveMaintenanceWindows(context.Background())

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, result)
	assert.Len(t, result.Windows, 1)
	assert.Equal(t, 1, result.Count)
	assert.Equal(t, MaintenanceWindowStatusActive, result.Windows[0].Status)
}

func TestMaintenanceWindowsService_GetActiveMaintenanceWindows_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/maintenance-windows/active", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "No active maintenance windows",
			"data": map[string]interface{}{
				"windows": []map[string]interface{}{},
				"count":   0,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, resp, err := client.MaintenanceWindows.GetActiveMaintenanceWindows(context.Background())

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, result)
	assert.Empty(t, result.Windows)
	assert.Equal(t, 0, result.Count)
}

func TestMaintenanceWindowsService_GetActiveMaintenanceWindows_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Failed to retrieve active windows",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, _, err := client.MaintenanceWindows.GetActiveMaintenanceWindows(context.Background())

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestMaintenanceWindowsService_GetUpcomingMaintenanceWindows_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/maintenance-windows/upcoming", r.URL.Path)
		assert.Equal(t, "7", r.URL.Query().Get("days"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Upcoming maintenance windows retrieved",
			"data": map[string]interface{}{
				"windows": []map[string]interface{}{
					{
						"id":               1,
						"window_uuid":      "mw-uuid-1",
						"organization_id":  1,
						"name":             "Upcoming Maintenance 1",
						"starts_at":        "2024-01-15T02:00:00Z",
						"ends_at":          "2024-01-15T04:00:00Z",
						"duration":         120,
						"status":           "scheduled",
						"suppress_alerts":  true,
						"pause_monitoring": false,
						"notify_before":    30,
						"server_filter":    map[string]interface{}{},
						"created_at":       "2024-01-01T00:00:00Z",
						"updated_at":       "2024-01-01T00:00:00Z",
					},
					{
						"id":               2,
						"window_uuid":      "mw-uuid-2",
						"organization_id":  1,
						"name":             "Upcoming Maintenance 2",
						"starts_at":        "2024-01-20T00:00:00Z",
						"ends_at":          "2024-01-20T02:00:00Z",
						"duration":         120,
						"status":           "scheduled",
						"suppress_alerts":  true,
						"pause_monitoring": false,
						"notify_before":    60,
						"server_filter":    map[string]interface{}{},
						"created_at":       "2024-01-01T00:00:00Z",
						"updated_at":       "2024-01-01T00:00:00Z",
					},
				},
				"count":     2,
				"days_from": 7,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, resp, err := client.MaintenanceWindows.GetUpcomingMaintenanceWindows(context.Background(), &GetUpcomingOptions{
		Days: 7,
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, result)
	assert.Len(t, result.Windows, 2)
	assert.Equal(t, 2, result.Count)
	assert.Equal(t, 7, result.DaysFrom)
}

func TestMaintenanceWindowsService_GetUpcomingMaintenanceWindows_NilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/maintenance-windows/upcoming", r.URL.Path)
		assert.Empty(t, r.URL.Query().Get("days"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Upcoming maintenance windows retrieved",
			"data": map[string]interface{}{
				"windows":   []map[string]interface{}{},
				"count":     0,
				"days_from": 30,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, resp, err := client.MaintenanceWindows.GetUpcomingMaintenanceWindows(context.Background(), nil)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, result)
}

func TestMaintenanceWindowsService_GetUpcomingMaintenanceWindows_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Failed to retrieve upcoming windows",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, _, err := client.MaintenanceWindows.GetUpcomingMaintenanceWindows(context.Background(), nil)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestMaintenanceWindowsService_CancelMaintenanceWindow_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/maintenance-windows/123/cancel", r.URL.Path)

		var reqBody CancelMaintenanceWindowRequest
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		require.NoError(t, err)
		assert.Equal(t, "Emergency change of plans", reqBody.Reason)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Maintenance window cancelled",
			"data": map[string]interface{}{
				"id":               123,
				"window_uuid":      "mw-uuid-123",
				"organization_id":  1,
				"name":             "Cancelled Maintenance",
				"starts_at":        "2024-01-15T02:00:00Z",
				"ends_at":          "2024-01-15T04:00:00Z",
				"duration":         120,
				"status":           "cancelled",
				"suppress_alerts":  true,
				"pause_monitoring": false,
				"notify_before":    30,
				"server_filter":    map[string]interface{}{},
				"cancelled_at":     "2024-01-14T00:00:00Z",
				"cancel_reason":    "Emergency change of plans",
				"created_at":       "2024-01-01T00:00:00Z",
				"updated_at":       "2024-01-14T00:00:00Z",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	window, resp, err := client.MaintenanceWindows.CancelMaintenanceWindow(context.Background(), 123, &CancelMaintenanceWindowRequest{
		Reason: "Emergency change of plans",
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, window)
	assert.Equal(t, uint(123), window.ID)
	assert.Equal(t, MaintenanceWindowStatusCancelled, window.Status)
	assert.NotNil(t, window.CancelledAt)
	assert.NotNil(t, window.CancelReason)
	assert.Equal(t, "Emergency change of plans", *window.CancelReason)
}

func TestMaintenanceWindowsService_CancelMaintenanceWindow_NoReason(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/maintenance-windows/123/cancel", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Maintenance window cancelled",
			"data": map[string]interface{}{
				"id":               123,
				"window_uuid":      "mw-uuid-123",
				"organization_id":  1,
				"name":             "Cancelled Maintenance",
				"starts_at":        "2024-01-15T02:00:00Z",
				"ends_at":          "2024-01-15T04:00:00Z",
				"duration":         120,
				"status":           "cancelled",
				"suppress_alerts":  true,
				"pause_monitoring": false,
				"notify_before":    30,
				"server_filter":    map[string]interface{}{},
				"cancelled_at":     "2024-01-14T00:00:00Z",
				"created_at":       "2024-01-01T00:00:00Z",
				"updated_at":       "2024-01-14T00:00:00Z",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	window, resp, err := client.MaintenanceWindows.CancelMaintenanceWindow(context.Background(), 123, &CancelMaintenanceWindowRequest{})

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, window)
	assert.Equal(t, MaintenanceWindowStatusCancelled, window.Status)
}

func TestMaintenanceWindowsService_CancelMaintenanceWindow_AlreadyCompleted(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Cannot cancel completed maintenance window",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	window, _, err := client.MaintenanceWindows.CancelMaintenanceWindow(context.Background(), 123, &CancelMaintenanceWindowRequest{})

	assert.Error(t, err)
	assert.Nil(t, window)
}

func TestMaintenanceWindowsService_CancelMaintenanceWindow_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Maintenance window not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	window, _, err := client.MaintenanceWindows.CancelMaintenanceWindow(context.Background(), 999, &CancelMaintenanceWindowRequest{})

	assert.Error(t, err)
	assert.Nil(t, window)
}

// =============================================================================
// ToQuery Method Tests
// =============================================================================

func TestListMaintenanceWindowsOptions_ToQuery(t *testing.T) {
	opts := &ListMaintenanceWindowsOptions{
		Page:     1,
		Limit:    20,
		Status:   "scheduled",
		FromDate: "2024-01-01",
		ToDate:   "2024-01-31",
	}

	query := opts.ToQuery()

	assert.Equal(t, "1", query["page"])
	assert.Equal(t, "20", query["limit"])
	assert.Equal(t, "scheduled", query["status"])
	assert.Equal(t, "2024-01-01", query["from_date"])
	assert.Equal(t, "2024-01-31", query["to_date"])
}

func TestListMaintenanceWindowsOptions_ToQuery_Empty(t *testing.T) {
	opts := &ListMaintenanceWindowsOptions{}
	query := opts.ToQuery()

	assert.Empty(t, query)
}

func TestListMaintenanceWindowsOptions_ToQuery_Partial(t *testing.T) {
	opts := &ListMaintenanceWindowsOptions{
		Page:   2,
		Status: "active",
	}

	query := opts.ToQuery()

	assert.Equal(t, "2", query["page"])
	assert.Equal(t, "active", query["status"])
	assert.Empty(t, query["limit"])
	assert.Empty(t, query["from_date"])
	assert.Empty(t, query["to_date"])
}

func TestGetUpcomingOptions_ToQuery(t *testing.T) {
	opts := &GetUpcomingOptions{
		Days: 14,
	}

	query := opts.ToQuery()

	assert.Equal(t, "14", query["days"])
}

func TestGetUpcomingOptions_ToQuery_Empty(t *testing.T) {
	opts := &GetUpcomingOptions{}
	query := opts.ToQuery()

	assert.Empty(t, query)
}

func TestGetUpcomingOptions_ToQuery_ZeroDays(t *testing.T) {
	opts := &GetUpcomingOptions{
		Days: 0,
	}

	query := opts.ToQuery()

	assert.Empty(t, query)
}

// =============================================================================
// MaintenanceAction Tests
// =============================================================================

func TestMaintenanceAction_Struct(t *testing.T) {
	action := MaintenanceAction{
		Type: "notify",
		Config: map[string]interface{}{
			"channel": "slack",
			"message": "Maintenance starting",
		},
		Order: 1,
	}

	assert.Equal(t, "notify", action.Type)
	assert.Equal(t, "slack", action.Config["channel"])
	assert.Equal(t, 1, action.Order)
}

// =============================================================================
// MaintenanceWindowStatus Tests
// =============================================================================

func TestMaintenanceWindowStatus_Constants(t *testing.T) {
	assert.Equal(t, MaintenanceWindowStatus("scheduled"), MaintenanceWindowStatusScheduled)
	assert.Equal(t, MaintenanceWindowStatus("active"), MaintenanceWindowStatusActive)
	assert.Equal(t, MaintenanceWindowStatus("completed"), MaintenanceWindowStatusCompleted)
	assert.Equal(t, MaintenanceWindowStatus("cancelled"), MaintenanceWindowStatusCancelled)
}

// =============================================================================
// Integration-style Tests
// =============================================================================

func TestMaintenanceWindowsService_FullWorkflow(t *testing.T) {
	// Simulate a full workflow: create -> get -> update -> cancel
	createCalled := false
	getCalled := false
	updateCalled := false
	cancelCalled := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == "POST" && r.URL.Path == "/v1/maintenance-windows":
			createCalled = true
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "success",
				"message": "Created",
				"data": map[string]interface{}{
					"id":               100,
					"window_uuid":      "mw-workflow-test",
					"organization_id":  1,
					"name":             "Workflow Test",
					"starts_at":        "2024-01-15T02:00:00Z",
					"ends_at":          "2024-01-15T04:00:00Z",
					"duration":         120,
					"status":           "scheduled",
					"suppress_alerts":  true,
					"pause_monitoring": false,
					"notify_before":    30,
					"server_filter":    map[string]interface{}{},
					"created_at":       "2024-01-01T00:00:00Z",
					"updated_at":       "2024-01-01T00:00:00Z",
				},
			})

		case r.Method == "GET" && r.URL.Path == "/v1/maintenance-windows/100":
			getCalled = true
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "success",
				"message": "Retrieved",
				"data": map[string]interface{}{
					"id":               100,
					"window_uuid":      "mw-workflow-test",
					"organization_id":  1,
					"name":             "Workflow Test",
					"starts_at":        "2024-01-15T02:00:00Z",
					"ends_at":          "2024-01-15T04:00:00Z",
					"duration":         120,
					"status":           "scheduled",
					"suppress_alerts":  true,
					"pause_monitoring": false,
					"notify_before":    30,
					"server_filter":    map[string]interface{}{},
					"created_at":       "2024-01-01T00:00:00Z",
					"updated_at":       "2024-01-01T00:00:00Z",
				},
			})

		case r.Method == "PUT" && r.URL.Path == "/v1/maintenance-windows/100":
			updateCalled = true
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "success",
				"message": "Updated",
				"data": map[string]interface{}{
					"id":               100,
					"window_uuid":      "mw-workflow-test",
					"organization_id":  1,
					"name":             "Updated Workflow Test",
					"starts_at":        "2024-01-15T02:00:00Z",
					"ends_at":          "2024-01-15T04:00:00Z",
					"duration":         120,
					"status":           "scheduled",
					"suppress_alerts":  true,
					"pause_monitoring": false,
					"notify_before":    30,
					"server_filter":    map[string]interface{}{},
					"created_at":       "2024-01-01T00:00:00Z",
					"updated_at":       "2024-01-02T00:00:00Z",
				},
			})

		case r.Method == "POST" && r.URL.Path == "/v1/maintenance-windows/100/cancel":
			cancelCalled = true
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "success",
				"message": "Cancelled",
				"data": map[string]interface{}{
					"id":               100,
					"window_uuid":      "mw-workflow-test",
					"organization_id":  1,
					"name":             "Updated Workflow Test",
					"starts_at":        "2024-01-15T02:00:00Z",
					"ends_at":          "2024-01-15T04:00:00Z",
					"duration":         120,
					"status":           "cancelled",
					"suppress_alerts":  true,
					"pause_monitoring": false,
					"notify_before":    30,
					"server_filter":    map[string]interface{}{},
					"cancelled_at":     "2024-01-03T00:00:00Z",
					"cancel_reason":    "Test complete",
					"created_at":       "2024-01-01T00:00:00Z",
					"updated_at":       "2024-01-03T00:00:00Z",
				},
			})
		}
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	ctx := context.Background()

	// Step 1: Create
	window, _, err := client.MaintenanceWindows.CreateMaintenanceWindow(ctx, &CreateMaintenanceWindowRequest{
		Name:         "Workflow Test",
		StartsAt:     "2024-01-15T02:00:00Z",
		EndsAt:       "2024-01-15T04:00:00Z",
		ServerFilter: map[string]interface{}{},
	})
	require.NoError(t, err)
	assert.Equal(t, uint(100), window.ID)

	// Step 2: Get
	window, _, err = client.MaintenanceWindows.GetMaintenanceWindow(ctx, 100)
	require.NoError(t, err)
	assert.Equal(t, "Workflow Test", window.Name)

	// Step 3: Update
	newName := "Updated Workflow Test"
	window, _, err = client.MaintenanceWindows.UpdateMaintenanceWindow(ctx, 100, &UpdateMaintenanceWindowRequest{
		Name: &newName,
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated Workflow Test", window.Name)

	// Step 4: Cancel
	window, _, err = client.MaintenanceWindows.CancelMaintenanceWindow(ctx, 100, &CancelMaintenanceWindowRequest{
		Reason: "Test complete",
	})
	require.NoError(t, err)
	assert.Equal(t, MaintenanceWindowStatusCancelled, window.Status)

	// Verify all steps were called
	assert.True(t, createCalled, "Create should have been called")
	assert.True(t, getCalled, "Get should have been called")
	assert.True(t, updateCalled, "Update should have been called")
	assert.True(t, cancelCalled, "Cancel should have been called")
}
