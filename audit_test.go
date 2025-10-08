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

func TestAuditService_GetAuditLogs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/audit/logs", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "50", r.URL.Query().Get("limit"))
		assert.Equal(t, "create", r.URL.Query().Get("action"))
		assert.Equal(t, "server", r.URL.Query().Get("resource_type"))
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		userID := uint(10)
		response := struct {
			Data []AuditLog      `json:"data"`
			Meta *PaginationMeta `json:"meta"`
		}{
			Data: []AuditLog{
				{
					ID:             1,
					OrganizationID: 100,
					UserID:         &userID,
					UserEmail:      "admin@example.com",
					UserName:       "Admin User",
					Action:         "create",
					ResourceType:   "server",
					ResourceID:     "server-123",
					ResourceName:   "web-server-01",
					Description:    "Created new server",
					IPAddress:      "192.168.1.100",
					UserAgent:      "Mozilla/5.0",
					Severity:       "info",
					Status:         "success",
					DurationMs:     150,
					CreatedAt:      CustomTime{Time: time.Now()},
				},
				{
					ID:             2,
					OrganizationID: 100,
					UserID:         &userID,
					UserEmail:      "admin@example.com",
					Action:         "create",
					ResourceType:   "server",
					ResourceID:     "server-456",
					ResourceName:   "db-server-01",
					Description:    "Created database server",
					Severity:       "info",
					Status:         "success",
					CreatedAt:      CustomTime{Time: time.Now()},
				},
			},
			Meta: &PaginationMeta{
				Page:       1,
				PerPage:    50,
				TotalItems: 2,
				TotalPages: 1,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	logs, meta, err := client.Audit.GetAuditLogs(context.Background(),
		&PaginationOptions{Page: 1, Limit: 50},
		map[string]interface{}{
			"action":        "create",
			"resource_type": "server",
		})
	require.NoError(t, err)
	assert.Len(t, logs, 2)
	assert.Equal(t, "create", logs[0].Action)
	assert.Equal(t, "server", logs[0].ResourceType)
	assert.Equal(t, "admin@example.com", logs[0].UserEmail)
	assert.Equal(t, "success", logs[0].Status)
	assert.NotNil(t, meta)
	assert.Equal(t, 2, meta.TotalItems)
}

func TestAuditService_GetAuditLogs_WithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/audit/logs", r.URL.Path)
		assert.Equal(t, "10", r.URL.Query().Get("user_id"))
		assert.Equal(t, "delete", r.URL.Query().Get("action"))
		assert.Equal(t, "critical", r.URL.Query().Get("severity"))
		assert.Equal(t, "2024-01-01T00:00:00Z", r.URL.Query().Get("start_date"))

		userID := uint(10)
		response := struct {
			Data []AuditLog      `json:"data"`
			Meta *PaginationMeta `json:"meta"`
		}{
			Data: []AuditLog{
				{
					ID:             5,
					OrganizationID: 100,
					UserID:         &userID,
					UserEmail:      "admin@example.com",
					Action:         "delete",
					ResourceType:   "server",
					ResourceID:     "server-999",
					ResourceName:   "old-server",
					Description:    "Deleted old server",
					Severity:       "critical",
					Status:         "success",
					CreatedAt:      CustomTime{Time: time.Now()},
				},
			},
			Meta: &PaginationMeta{
				Page:       1,
				PerPage:    20,
				TotalItems: 1,
				TotalPages: 1,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	logs, meta, err := client.Audit.GetAuditLogs(context.Background(),
		nil,
		map[string]interface{}{
			"user_id":    uint(10),
			"action":     "delete",
			"severity":   "critical",
			"start_date": "2024-01-01T00:00:00Z",
		})
	require.NoError(t, err)
	assert.Len(t, logs, 1)
	assert.Equal(t, "delete", logs[0].Action)
	assert.Equal(t, "critical", logs[0].Severity)
	assert.NotNil(t, meta)
}

func TestAuditService_GetAuditLog(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/audit/logs/123", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		userID := uint(10)
		response := struct {
			Data    *AuditLog `json:"data"`
			Status  string    `json:"status"`
			Message string    `json:"message"`
		}{
			Data: &AuditLog{
				ID:             123,
				OrganizationID: 100,
				UserID:         &userID,
				UserEmail:      "admin@example.com",
				UserName:       "Admin User",
				Action:         "update",
				ResourceType:   "alert",
				ResourceID:     "alert-456",
				ResourceName:   "CPU Alert",
				Description:    "Updated alert threshold",
				IPAddress:      "192.168.1.100",
				UserAgent:      "Mozilla/5.0",
				Severity:       "info",
				Status:         "success",
				DurationMs:     75,
				Location:       "US-East",
				DeviceType:     "desktop",
				ComplianceFlags: []string{"SOC2", "GDPR"},
				CreatedAt:      CustomTime{Time: time.Now()},
			},
			Status:  "success",
			Message: "Audit log retrieved successfully",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	log, err := client.Audit.GetAuditLog(context.Background(), 123)
	require.NoError(t, err)
	assert.Equal(t, uint(123), log.ID)
	assert.Equal(t, "update", log.Action)
	assert.Equal(t, "alert", log.ResourceType)
	assert.Equal(t, "admin@example.com", log.UserEmail)
	assert.Equal(t, "US-East", log.Location)
	assert.Equal(t, "desktop", log.DeviceType)
	assert.Len(t, log.ComplianceFlags, 2)
	assert.Contains(t, log.ComplianceFlags, "SOC2")
	assert.Contains(t, log.ComplianceFlags, "GDPR")
}

func TestAuditService_ExportAuditLogs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/audit/logs/export", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		var reqBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&reqBody)
		assert.Equal(t, "csv", reqBody["format"])
		assert.Equal(t, "create", reqBody["action"])

		// Return CSV data
		w.Header().Set("Content-Type", "text/csv")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ID,Action,Resource,User,Timestamp\n1,create,server,admin@example.com,2024-01-01T00:00:00Z\n"))
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	data, err := client.Audit.ExportAuditLogs(context.Background(),
		"csv",
		map[string]interface{}{
			"action": "create",
		})
	require.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, string(data), "ID,Action,Resource")
}

func TestAuditService_GetAuditStatistics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/audit/statistics", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		response := struct {
			Data    *AuditStatistics `json:"data"`
			Status  string           `json:"status"`
			Message string           `json:"message"`
		}{
			Data: &AuditStatistics{
				TotalLogs:    1500,
				TotalUsers:   25,
				TotalActions: 8,
				ActionBreakdown: map[string]int{
					"create": 600,
					"update": 450,
					"delete": 200,
					"login":  250,
				},
				ResourceBreakdown: map[string]int{
					"server":       700,
					"user":         400,
					"alert":        200,
					"organization": 200,
				},
				SeverityBreakdown: map[string]int{
					"info":     1200,
					"warning":  200,
					"critical": 100,
				},
				StatusBreakdown: map[string]int{
					"success": 1400,
					"failure": 100,
				},
				TopUsers: []AuditUserActivity{
					{
						UserID:         1,
						UserEmail:      "admin@example.com",
						UserName:       "Admin User",
						ActionCount:    500,
						FailedAttempts: 5,
						LastActivity:   CustomTime{Time: time.Now()},
						TopActions:     []string{"create", "update", "delete"},
					},
					{
						UserID:         2,
						UserEmail:      "user@example.com",
						UserName:       "Regular User",
						ActionCount:    200,
						FailedAttempts: 2,
						LastActivity:   CustomTime{Time: time.Now()},
						TopActions:     []string{"update", "read"},
					},
				},
				TopActions: []AuditActionCount{
					{
						Action:      "create",
						Count:       600,
						SuccessRate: 98.5,
					},
					{
						Action:      "update",
						Count:       450,
						SuccessRate: 99.1,
					},
				},
				TopResources: []AuditResourceCount{
					{
						ResourceType: "server",
						AccessCount:  700,
					},
					{
						ResourceType: "user",
						AccessCount:  400,
					},
				},
				FailedAttempts:    100,
				CriticalEvents:    100,
				ComplianceBreakdown: map[string]int{
					"SOC2": 800,
					"GDPR": 700,
				},
				AverageDurationMs: 125.5,
				TimeRange: AuditTimeRange{
					StartDate:  CustomTime{Time: time.Now().AddDate(0, -1, 0)},
					EndDate:    CustomTime{Time: time.Now()},
					DurationMs: 2592000000, // 30 days
				},
			},
			Status:  "success",
			Message: "Audit statistics retrieved successfully",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	stats, err := client.Audit.GetAuditStatistics(context.Background(), "", "")
	require.NoError(t, err)
	assert.Equal(t, 1500, stats.TotalLogs)
	assert.Equal(t, 25, stats.TotalUsers)
	assert.Equal(t, 8, stats.TotalActions)
	assert.Len(t, stats.ActionBreakdown, 4)
	assert.Equal(t, 600, stats.ActionBreakdown["create"])
	assert.Len(t, stats.ResourceBreakdown, 4)
	assert.Equal(t, 700, stats.ResourceBreakdown["server"])
	assert.Len(t, stats.SeverityBreakdown, 3)
	assert.Equal(t, 1200, stats.SeverityBreakdown["info"])
	assert.Len(t, stats.TopUsers, 2)
	assert.Equal(t, "admin@example.com", stats.TopUsers[0].UserEmail)
	assert.Equal(t, 500, stats.TopUsers[0].ActionCount)
	assert.Len(t, stats.TopActions, 2)
	assert.Equal(t, "create", stats.TopActions[0].Action)
	assert.Equal(t, float64(98.5), stats.TopActions[0].SuccessRate)
	assert.Equal(t, 100, stats.FailedAttempts)
	assert.Equal(t, 100, stats.CriticalEvents)
	assert.Equal(t, 125.5, stats.AverageDurationMs)
}

func TestAuditService_GetUserAuditHistory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/audit/users/10/history", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "100", r.URL.Query().Get("limit"))
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		userID := uint(10)
		response := struct {
			Data []AuditLog      `json:"data"`
			Meta *PaginationMeta `json:"meta"`
		}{
			Data: []AuditLog{
				{
					ID:             1,
					OrganizationID: 100,
					UserID:         &userID,
					UserEmail:      "user@example.com",
					UserName:       "Test User",
					Action:         "login",
					ResourceType:   "session",
					Description:    "User logged in",
					IPAddress:      "192.168.1.50",
					Severity:       "info",
					Status:         "success",
					CreatedAt:      CustomTime{Time: time.Now()},
				},
				{
					ID:             2,
					OrganizationID: 100,
					UserID:         &userID,
					UserEmail:      "user@example.com",
					Action:         "update",
					ResourceType:   "profile",
					Description:    "Updated user profile",
					Severity:       "info",
					Status:         "success",
					CreatedAt:      CustomTime{Time: time.Now().Add(-1 * time.Hour)},
				},
			},
			Meta: &PaginationMeta{
				Page:       1,
				PerPage:    100,
				TotalItems: 2,
				TotalPages: 1,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	logs, meta, err := client.Audit.GetUserAuditHistory(context.Background(),
		10,
		&PaginationOptions{Page: 1, Limit: 100},
		"",
		"")
	require.NoError(t, err)
	assert.Len(t, logs, 2)
	assert.Equal(t, "login", logs[0].Action)
	assert.Equal(t, "user@example.com", logs[0].UserEmail)
	assert.Equal(t, uint(10), *logs[0].UserID)
	assert.NotNil(t, meta)
	assert.Equal(t, 2, meta.TotalItems)
}

func TestAuditService_ErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		expectedError bool
	}{
		{
			name:          "Unauthorized",
			statusCode:    http.StatusUnauthorized,
			expectedError: true,
		},
		{
			name:          "Forbidden",
			statusCode:    http.StatusForbidden,
			expectedError: true,
		},
		{
			name:          "Not Found",
			statusCode:    http.StatusNotFound,
			expectedError: true,
		},
		{
			name:          "Internal Server Error",
			statusCode:    http.StatusInternalServerError,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(StandardResponse{
					Status:  "error",
					Message: "Error occurred",
				})
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			_, _, err = client.Audit.GetAuditLogs(context.Background(), nil, nil)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
