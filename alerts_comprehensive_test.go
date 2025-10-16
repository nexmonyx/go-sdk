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

// TestAlertsService_CreateComprehensive tests the Create method with various scenarios
func TestAlertsService_CreateComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		alert      *Alert
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Alert)
	}{
		{
			name: "success - full alert creation",
			alert: &Alert{
				Name:           "High CPU Usage",
				Description:    "Alert when CPU exceeds threshold",
				OrganizationID: 1,
				Type:           "metric",
				MetricName:     "cpu_usage",
				Condition:      "greater_than",
				Threshold:      80.0,
				Duration:       300,
				Frequency:      60,
				Enabled:        true,
				Severity:       "warning",
				Channels:       []string{"email", "slack"},
			},
			mockStatus: http.StatusCreated,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":              1,
					"name":            "High CPU Usage",
					"description":     "Alert when CPU exceeds threshold",
					"organization_id": 1,
					"type":            "metric",
					"metric_name":     "cpu_usage",
					"condition":       "greater_than",
					"threshold":       80.0,
					"duration":        300,
					"frequency":       60,
					"enabled":         true,
					"status":          "active",
					"severity":        "warning",
					"channels":        []string{"email", "slack"},
					"created_at":      "2024-01-15T10:00:00Z",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alert *Alert) {
				assert.Equal(t, "High CPU Usage", alert.Name)
				assert.Equal(t, "cpu_usage", alert.MetricName)
				assert.Equal(t, 80.0, alert.Threshold)
				assert.True(t, alert.Enabled)
			},
		},
		{
			name: "validation error - missing name",
			alert: &Alert{
				OrganizationID: 1,
				Type:           "metric",
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Name is required",
			},
			wantErr: true,
		},
		{
			name: "validation error - invalid threshold",
			alert: &Alert{
				Name:           "Test Alert",
				OrganizationID: 1,
				Type:           "metric",
				Threshold:      -10.0,
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Invalid threshold value",
			},
			wantErr: true,
		},
		{
			name: "unauthorized",
			alert: &Alert{
				Name:           "Test Alert",
				OrganizationID: 1,
			},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name: "forbidden",
			alert: &Alert{
				Name:           "Test Alert",
				OrganizationID: 1,
			},
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Access denied",
			},
			wantErr: true,
		},
		{
			name: "server error",
			alert: &Alert{
				Name:           "Test Alert",
				OrganizationID: 1,
			},
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL:    server.URL,
				Auth:       AuthConfig{Token: "test-token"},
				RetryCount: 0,
			})
			require.NoError(t, err)

			ctx := context.Background()
			if tt.wantErr && tt.mockStatus >= 500 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
			}

			result, err := client.Alerts.Create(ctx, tt.alert)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestAlertsService_GetComprehensive tests the Get method with various scenarios
func TestAlertsService_GetComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Alert)
	}{
		{
			name: "success - get alert details",
			id:   "alert-id-123",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":              1,
					"name":            "CPU Alert",
					"description":     "CPU monitoring alert",
					"organization_id": 1,
					"type":            "metric",
					"metric_name":     "cpu_usage",
					"threshold":       80.0,
					"enabled":         true,
					"status":          "active",
					"severity":        "warning",
					"created_at":      "2024-01-15T10:00:00Z",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alert *Alert) {
				assert.Equal(t, "CPU Alert", alert.Name)
				assert.Equal(t, "cpu_usage", alert.MetricName)
				assert.True(t, alert.Enabled)
			},
		},
		{
			name:       "not found",
			id:         "non-existent",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Alert not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			id:         "alert-id-123",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "forbidden",
			id:         "alert-id-123",
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Access denied",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			id:         "alert-id-123",
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL:    server.URL,
				Auth:       AuthConfig{Token: "test-token"},
				RetryCount: 0,
			})
			require.NoError(t, err)

			ctx := context.Background()
			if tt.wantErr && tt.mockStatus >= 500 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
			}

			result, err := client.Alerts.Get(ctx, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestAlertsService_ListComprehensive tests the List method with various scenarios
func TestAlertsService_ListComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		opts       *ListOptions
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, []*Alert, *PaginationMeta)
	}{
		{
			name: "success - list all alerts",
			opts: &ListOptions{Page: 1, Limit: 25},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": []map[string]interface{}{
					{
						"id":          1,
						"name":        "CPU Alert",
						"type":        "metric",
						"enabled":     true,
						"status":      "active",
						"severity":    "warning",
					},
					{
						"id":          2,
						"name":        "Memory Alert",
						"type":        "metric",
						"enabled":     true,
						"status":      "active",
						"severity":    "critical",
					},
				},
				"meta": map[string]interface{}{
					"page":        1,
					"limit":       25,
					"total_items": 2,
					"total_pages": 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alerts []*Alert, meta *PaginationMeta) {
				assert.Len(t, alerts, 2)
				assert.Equal(t, 2, meta.TotalItems)
			},
		},
		{
			name:       "success - empty results",
			opts:       &ListOptions{Page: 1, Limit: 25},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data":   []map[string]interface{}{},
				"meta": map[string]interface{}{
					"page":        1,
					"limit":       25,
					"total_items": 0,
					"total_pages": 0,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alerts []*Alert, meta *PaginationMeta) {
				assert.Len(t, alerts, 0)
				assert.Equal(t, 0, meta.TotalItems)
			},
		},
		{
			name:       "unauthorized",
			opts:       &ListOptions{Page: 1, Limit: 25},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "forbidden",
			opts:       &ListOptions{Page: 1, Limit: 25},
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Access denied",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			opts:       &ListOptions{Page: 1, Limit: 25},
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL:    server.URL,
				Auth:       AuthConfig{Token: "test-token"},
				RetryCount: 0,
			})
			require.NoError(t, err)

			ctx := context.Background()
			if tt.wantErr && tt.mockStatus >= 500 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
			}

			alerts, meta, err := client.Alerts.List(ctx, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, alerts)
				if tt.checkFunc != nil {
					tt.checkFunc(t, alerts, meta)
				}
			}
		})
	}
}

// TestAlertsService_EnableDisableComprehensive tests Enable and Disable methods
func TestAlertsService_EnableDisableComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		method     string // "enable" or "disable"
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Alert)
	}{
		{
			name:   "success - enable alert",
			id:     "alert-id-123",
			method: "enable",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":      1,
					"name":    "Test Alert",
					"enabled": true,
					"status":  "active",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alert *Alert) {
				assert.True(t, alert.Enabled)
			},
		},
		{
			name:   "success - disable alert",
			id:     "alert-id-123",
			method: "disable",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":      1,
					"name":    "Test Alert",
					"enabled": false,
					"status":  "inactive",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alert *Alert) {
				assert.False(t, alert.Enabled)
			},
		},
		{
			name:       "not found - enable",
			id:         "non-existent",
			method:     "enable",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Alert not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized - disable",
			id:         "alert-id-123",
			method:     "disable",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "server error - enable",
			id:         "alert-id-123",
			method:     "enable",
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL:    server.URL,
				Auth:       AuthConfig{Token: "test-token"},
				RetryCount: 0,
			})
			require.NoError(t, err)

			ctx := context.Background()
			if tt.wantErr && tt.mockStatus >= 500 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
			}

			var result *Alert
			if tt.method == "enable" {
				result, err = client.Alerts.Enable(ctx, tt.id)
			} else {
				result, err = client.Alerts.Disable(ctx, tt.id)
			}

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestAlertsService_DeleteComprehensive tests the Delete method with various scenarios
func TestAlertsService_DeleteComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:       "success - delete alert",
			id:         "alert-id-123",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status":  "success",
				"message": "Alert deleted successfully",
			},
			wantErr: false,
		},
		{
			name:       "success - no content",
			id:         "alert-id-456",
			mockStatus: http.StatusNoContent,
			mockBody:   nil,
			wantErr:    false,
		},
		{
			name:       "not found",
			id:         "non-existent",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Alert not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			id:         "alert-id-123",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "forbidden",
			id:         "alert-id-123",
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Access denied",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			id:         "alert-id-123",
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "DELETE", r.Method)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				if tt.mockBody != nil {
					json.NewEncoder(w).Encode(tt.mockBody)
				}
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL:    server.URL,
				Auth:       AuthConfig{Token: "test-token"},
				RetryCount: 0,
			})
			require.NoError(t, err)

			ctx := context.Background()
			if tt.wantErr && tt.mockStatus >= 500 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
			}

			err = client.Alerts.Delete(ctx, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
