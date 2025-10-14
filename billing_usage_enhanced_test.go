package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Additional tests to improve billing_usage.go coverage from 60.5% to 70%+

func TestBillingUsageService_RecordUsageMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/admin/usage-metrics/record", r.URL.Path)

		var req UsageMetricsRecordRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, uint(100), req.OrganizationID)
		assert.Equal(t, 25, req.ActiveAgentCount)

		response := StandardResponse{
			Status:  "success",
			Message: "Usage metrics recorded successfully",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "admin-token",
		},
	})
	require.NoError(t, err)

	metrics := &UsageMetricsRecordRequest{
		OrganizationID:   100,
		ActiveAgentCount: 25,
		TotalAgentCount:  30,
		StorageUsedBytes: 161620273152,
		RetentionDays:    30,
		CollectedAt:      time.Now(),
	}

	err = client.BillingUsage.RecordUsageMetrics(context.Background(), metrics)
	require.NoError(t, err)
}

func TestBillingUsageService_GetOrgAgentCounts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/admin/usage-metrics/100/agent-counts", r.URL.Path)

		response := StandardResponse{
			Status:  "success",
			Message: "Agent counts retrieved successfully",
			Data: &AgentCountsResponse{
				OrganizationID: 100,
				ActiveCount:    25,
				TotalCount:     30,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "admin-token",
		},
	})
	require.NoError(t, err)

	counts, err := client.BillingUsage.GetOrgAgentCounts(context.Background(), 100)
	require.NoError(t, err)
	assert.NotNil(t, counts)
	assert.Equal(t, uint(100), counts.OrganizationID)
	assert.Equal(t, 25, counts.ActiveCount)
	assert.Equal(t, 30, counts.TotalCount)
}

func TestBillingUsageService_GetOrgStorageUsage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/admin/usage-metrics/100/storage", r.URL.Path)

		response := StandardResponse{
			Status:  "success",
			Message: "Storage usage retrieved successfully",
			Data: &StorageUsageResponse{
				OrganizationID: 100,
				StorageBytes:   161620273152,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "admin-token",
		},
	})
	require.NoError(t, err)

	storage, err := client.BillingUsage.GetOrgStorageUsage(context.Background(), 100)
	require.NoError(t, err)
	assert.NotNil(t, storage)
	assert.Equal(t, uint(100), storage.OrganizationID)
	assert.Equal(t, int64(161620273152), storage.StorageBytes)
}

func TestBillingUsageService_GetMyUsageHistory_EmptyDates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/billing/usage/history", r.URL.Path)
		// When dates are empty (zero value), they should not be in query params
		assert.Empty(t, r.URL.Query().Get("start_date"))
		assert.Empty(t, r.URL.Query().Get("end_date"))

		history := []UsageMetricsHistory{
			{
				ID:               1,
				OrganizationID:   100,
				ActiveAgentCount: 25,
				StorageUsedGB:    150.5,
			},
		}

		response := StandardResponse{
			Status:  "success",
			Message: "Usage history retrieved successfully",
			Data:    history,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-jwt-token",
		},
	})
	require.NoError(t, err)

	// Pass zero values for times
	history, err := client.BillingUsage.GetMyUsageHistory(context.Background(), time.Time{}, time.Time{}, "")
	require.NoError(t, err)
	assert.Len(t, history, 1)
}

func TestBillingUsageService_GetOrgUsageHistory_WithInterval(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/admin/billing/organizations/100/usage/history", r.URL.Path)
		assert.Equal(t, "hourly", r.URL.Query().Get("interval"))

		history := []UsageMetricsHistory{
			{
				ID:               1,
				OrganizationID:   100,
				ActiveAgentCount: 25,
			},
		}

		response := StandardResponse{
			Status: "success",
			Data:   history,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "admin-token",
		},
	})
	require.NoError(t, err)

	history, err := client.BillingUsage.GetOrgUsageHistory(context.Background(), 100, time.Time{}, time.Time{}, "hourly")
	require.NoError(t, err)
	assert.Len(t, history, 1)
}

func TestBillingUsageService_GetAllUsageOverview_NilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/admin/billing/usage/overview", r.URL.Path)

		overview := &OrganizationUsageOverview{
			TotalOrganizations: 5,
			TotalActiveAgents:  100,
			TotalStorageGB:     500.0,
		}

		response := map[string]interface{}{
			"status": "success",
			"data":   overview,
			"pagination": map[string]interface{}{
				"page":  1,
				"limit": 25,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "admin-token",
		},
	})
	require.NoError(t, err)

	// Pass nil options
	overview, meta, err := client.BillingUsage.GetAllUsageOverview(context.Background(), nil)
	require.NoError(t, err)
	assert.NotNil(t, overview)
	assert.NotNil(t, meta)
	assert.Equal(t, 5, overview.TotalOrganizations)
}
