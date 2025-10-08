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

func TestBillingUsageService_GetMyCurrentUsage(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/billing/usage/current", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		// Send response
		response := StandardResponse{
			Status:  "success",
			Message: "Current usage metrics retrieved successfully",
			Data: &OrganizationUsageMetrics{
				ID:               1,
				OrganizationID:   100,
				ActiveAgentCount: 25,
				TotalAgentCount:  30,
				StorageUsedGB:    150.5,
				RetentionDays:    30,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-jwt-token",
		},
	})
	require.NoError(t, err)

	// Test GetMyCurrentUsage
	usage, err := client.BillingUsage.GetMyCurrentUsage(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, usage)
	assert.Equal(t, uint(100), usage.OrganizationID)
	assert.Equal(t, 25, usage.ActiveAgentCount)
	assert.Equal(t, 30, usage.TotalAgentCount)
	assert.Equal(t, 150.5, usage.StorageUsedGB)
}

func TestBillingUsageService_GetMyUsageHistory(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/billing/usage/history", r.URL.Path)
		assert.NotEmpty(t, r.URL.Query().Get("start_date"))
		assert.NotEmpty(t, r.URL.Query().Get("end_date"))
		assert.Equal(t, "daily", r.URL.Query().Get("interval"))

		// Send response
		history := []UsageMetricsHistory{
			{
				ID:               1,
				OrganizationID:   100,
				ActiveAgentCount: 25,
				StorageUsedGB:    150.5,
				RetentionDays:    30,
			},
			{
				ID:               2,
				OrganizationID:   100,
				ActiveAgentCount: 24,
				StorageUsedGB:    148.2,
				RetentionDays:    30,
			},
		}
		response := StandardResponse{
			Status:  "success",
			Message: "Usage history retrieved successfully",
			Data:    history,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-jwt-token",
		},
	})
	require.NoError(t, err)

	// Test GetMyUsageHistory
	startDate := time.Now().AddDate(0, 0, -7)
	endDate := time.Now()
	history, err := client.BillingUsage.GetMyUsageHistory(context.Background(), startDate, endDate, "daily")
	require.NoError(t, err)
	assert.Len(t, history, 2)
	assert.Equal(t, uint(100), history[0].OrganizationID)
	assert.Equal(t, 25, history[0].ActiveAgentCount)
}

func TestBillingUsageService_GetMyUsageSummary(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/billing/usage/summary", r.URL.Path)
		assert.NotEmpty(t, r.URL.Query().Get("start_date"))
		assert.NotEmpty(t, r.URL.Query().Get("end_date"))

		// Send response
		response := StandardResponse{
			Status:  "success",
			Message: "Usage summary calculated successfully",
			Data: &UsageSummary{
				OrganizationID:    100,
				AverageAgentCount: 24.5,
				MaxAgentCount:     30,
				AverageStorageGB:  149.3,
				MaxStorageGB:      155.2,
				TotalDataPoints:   30,
				RetentionDays:     30,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-jwt-token",
		},
	})
	require.NoError(t, err)

	// Test GetMyUsageSummary
	startDate := time.Now().AddDate(0, -1, 0)
	endDate := time.Now()
	summary, err := client.BillingUsage.GetMyUsageSummary(context.Background(), startDate, endDate)
	require.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, uint(100), summary.OrganizationID)
	assert.Equal(t, 24.5, summary.AverageAgentCount)
	assert.Equal(t, 30, summary.MaxAgentCount)
	assert.Equal(t, 30, summary.TotalDataPoints)
}

func TestBillingUsageService_GetOrgCurrentUsage(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/admin/billing/organizations/100/usage", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		// Send response
		response := StandardResponse{
			Status:  "success",
			Message: "Current usage metrics retrieved successfully",
			Data: &OrganizationUsageMetrics{
				ID:               1,
				OrganizationID:   100,
				ActiveAgentCount: 25,
				TotalAgentCount:  30,
				StorageUsedGB:    150.5,
				RetentionDays:    30,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with admin auth
	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "admin-jwt-token",
		},
	})
	require.NoError(t, err)

	// Test GetOrgCurrentUsage
	usage, err := client.BillingUsage.GetOrgCurrentUsage(context.Background(), 100)
	require.NoError(t, err)
	assert.NotNil(t, usage)
	assert.Equal(t, uint(100), usage.OrganizationID)
	assert.Equal(t, 25, usage.ActiveAgentCount)
}

func TestBillingUsageService_GetOrgUsageHistory(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/admin/billing/organizations/100/usage/history", r.URL.Path)
		assert.Equal(t, "monthly", r.URL.Query().Get("interval"))

		// Send response
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
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "admin-jwt-token",
		},
	})
	require.NoError(t, err)

	// Test GetOrgUsageHistory
	startDate := time.Now().AddDate(0, -3, 0)
	endDate := time.Now()
	history, err := client.BillingUsage.GetOrgUsageHistory(context.Background(), 100, startDate, endDate, "monthly")
	require.NoError(t, err)
	assert.Len(t, history, 1)
	assert.Equal(t, uint(100), history[0].OrganizationID)
}

func TestBillingUsageService_GetOrgUsageSummary(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/admin/billing/organizations/100/usage/summary", r.URL.Path)

		// Send response
		response := StandardResponse{
			Status:  "success",
			Message: "Usage summary calculated successfully",
			Data: &UsageSummary{
				OrganizationID:        100,
				AverageAgentCount:     24.5,
				MaxAgentCount:         30,
				BillingRecommendation: "Consider Enterprise tier",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "admin-jwt-token",
		},
	})
	require.NoError(t, err)

	// Test GetOrgUsageSummary
	startDate := time.Now().AddDate(0, -1, 0)
	endDate := time.Now()
	summary, err := client.BillingUsage.GetOrgUsageSummary(context.Background(), 100, startDate, endDate)
	require.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, uint(100), summary.OrganizationID)
	assert.Equal(t, "Consider Enterprise tier", summary.BillingRecommendation)
}

func TestBillingUsageService_GetAllUsageOverview(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/admin/billing/usage/overview", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "50", r.URL.Query().Get("limit"))

		// Send response
		overview := &OrganizationUsageOverview{
			TotalOrganizations: 10,
			TotalActiveAgents:  250,
			TotalStorageGB:     1500.0,
			Organizations: []OrganizationUsageMetrics{
				{
					ID:               1,
					OrganizationID:   100,
					ActiveAgentCount: 25,
					StorageUsedGB:    150.5,
				},
				{
					ID:               2,
					OrganizationID:   101,
					ActiveAgentCount: 30,
					StorageUsedGB:    200.0,
				},
			},
		}
		response := map[string]interface{}{
			"status":  "success",
			"message": "Usage overview retrieved successfully",
			"data":    overview,
			"pagination": map[string]interface{}{
				"page":        1,
				"limit":       50,
				"total_items": 10,
				"total_pages": 1,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "admin-jwt-token",
		},
	})
	require.NoError(t, err)

	// Test GetAllUsageOverview
	opts := &ListOptions{
		Page:  1,
		Limit: 50,
	}
	overview, meta, err := client.BillingUsage.GetAllUsageOverview(context.Background(), opts)
	require.NoError(t, err)
	assert.NotNil(t, overview)
	assert.NotNil(t, meta)
	assert.Equal(t, 10, overview.TotalOrganizations)
	assert.Equal(t, 250, overview.TotalActiveAgents)
	assert.Len(t, overview.Organizations, 2)
	assert.Equal(t, 1, meta.Page)
	assert.Equal(t, 50, meta.Limit)
}

func TestBillingUsageService_ErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		responseBody   string
		shouldError    bool
	}{
		{
			name:         "Unauthorized",
			statusCode:   401,
			responseBody: `{"status":"error","error":"unauthorized","message":"Authentication required"}`,
			shouldError:  true,
		},
		{
			name:         "Forbidden",
			statusCode:   403,
			responseBody: `{"status":"error","error":"forbidden","message":"Admin access required"}`,
			shouldError:  true,
		},
		{
			name:         "Not Found",
			statusCode:   404,
			responseBody: `{"status":"error","error":"not_found","message":"Usage metrics not found"}`,
			shouldError:  true,
		},
		{
			name:         "Internal Server Error",
			statusCode:   500,
			responseBody: `{"status":"error","error":"internal_server_error","message":"Database error"}`,
			shouldError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			// Create client
			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth: AuthConfig{
					Token: "test-token",
				},
			})
			require.NoError(t, err)

			// Test GetMyCurrentUsage with error
			_, err = client.BillingUsage.GetMyCurrentUsage(context.Background())
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
