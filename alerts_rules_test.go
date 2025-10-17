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

// TestAlertsService_Create_Rules tests the Create method with alert rule validation
func TestAlertsService_Create_Rules(t *testing.T) {
	tests := []struct {
		name       string
		alert      *Alert
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Alert)
	}{
		{
			name: "success - create cpu threshold alert (greater_than)",
			alert: &Alert{
				Name:           "High CPU Usage",
				Description:    "Alert when CPU exceeds 80%",
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
				"data": map[string]interface{}{
					"id":              1,
					"name":            "High CPU Usage",
					"metric_name":     "cpu_usage",
					"condition":       "greater_than",
					"threshold":       80.0,
					"duration":        300,
					"frequency":       60,
					"severity":        "warning",
					"enabled":         true,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alert *Alert) {
				assert.Equal(t, "High CPU Usage", alert.Name)
				assert.Equal(t, "cpu_usage", alert.MetricName)
				assert.Equal(t, "greater_than", alert.Condition)
				assert.Equal(t, 80.0, alert.Threshold)
			},
		},
		{
			name: "success - create memory alert (less_than)",
			alert: &Alert{
				Name:           "Low Memory Alert",
				MetricName:     "memory_available",
				Condition:      "less_than",
				Threshold:      20.0,
				Duration:       300,
				Frequency:      60,
				Severity:       "critical",
			},
			mockStatus: http.StatusCreated,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":          2,
					"metric_name": "memory_available",
					"condition":   "less_than",
					"threshold":   20.0,
					"severity":    "critical",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alert *Alert) {
				assert.Equal(t, "less_than", alert.Condition)
				assert.Equal(t, 20.0, alert.Threshold)
			},
		},
		{
			name: "success - create disk alert (greater_than_or_equal)",
			alert: &Alert{
				Name:       "Disk Usage High",
				MetricName: "disk_usage",
				Condition:  "greater_than_or_equal",
				Threshold:  90.0,
			},
			mockStatus: http.StatusCreated,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":          3,
					"condition":   "greater_than_or_equal",
					"threshold":   90.0,
				},
			},
			wantErr: false,
		},
		{
			name: "success - create network latency alert (less_than_or_equal)",
			alert: &Alert{
				Name:       "High Latency",
				MetricName: "network_latency",
				Condition:  "less_than_or_equal",
				Threshold:  100.0,
			},
			mockStatus: http.StatusCreated,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":        4,
					"condition": "less_than_or_equal",
				},
			},
			wantErr: false,
		},
		{
			name: "success - create response time alert (equal)",
			alert: &Alert{
				Name:       "Timeout Alert",
				MetricName: "response_time",
				Condition:  "equal",
				Threshold:  5000.0,
			},
			mockStatus: http.StatusCreated,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id": 5,
				},
			},
			wantErr: false,
		},
		{
			name: "validation error - missing name",
			alert: &Alert{
				MetricName: "cpu_usage",
				Condition:  "greater_than",
				Threshold:  80.0,
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"error": "name is required",
			},
			wantErr: true,
		},
		{
			name: "validation error - missing metric name",
			alert: &Alert{
				Name:      "Test Alert",
				Condition: "greater_than",
				Threshold: 80.0,
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"error": "metric_name is required",
			},
			wantErr: true,
		},
		{
			name: "validation error - invalid condition",
			alert: &Alert{
				Name:       "Test Alert",
				MetricName: "cpu_usage",
				Condition:  "invalid_op",
				Threshold:  80.0,
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"error": "invalid condition operator",
			},
			wantErr: true,
		},
		{
			name: "validation error - negative threshold",
			alert: &Alert{
				Name:       "Test Alert",
				MetricName: "cpu_usage",
				Condition:  "greater_than",
				Threshold:  -10.0,
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"error": "threshold must be positive",
			},
			wantErr: true,
		},
		{
			name: "unauthorized - invalid token",
			alert: &Alert{
				Name:       "Test",
				MetricName: "cpu",
				Condition:  "greater_than",
				Threshold:  80.0,
			},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"error": "invalid token",
			},
			wantErr: true,
		},
		{
			name: "server error",
			alert: &Alert{
				Name:       "Test",
				MetricName: "cpu",
				Condition:  "greater_than",
				Threshold:  80.0,
			},
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"error": "internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/v1/alerts/rules", r.URL.Path)

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
			alert, err := client.Alerts.Create(ctx, tt.alert)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, alert)
				if tt.checkFunc != nil {
					tt.checkFunc(t, alert)
				}
			}
		})
	}
}

// TestAlertsService_Get_Rules tests the Get method for alert rules
func TestAlertsService_Get_Rules(t *testing.T) {
	tests := []struct {
		name       string
		ruleID     string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Alert)
	}{
		{
			name:       "success - get cpu alert rule",
			ruleID:     "1",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":            1,
					"name":          "High CPU Alert",
					"metric_name":   "cpu_usage",
					"condition":     "greater_than",
					"threshold":     80.0,
					"duration":      300,
					"frequency":     60,
					"severity":      "warning",
					"enabled":       true,
					"channels":      []string{"email"},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alert *Alert) {
				assert.Equal(t, "High CPU Alert", alert.Name)
				assert.Equal(t, "cpu_usage", alert.MetricName)
				assert.Equal(t, "greater_than", alert.Condition)
				assert.Equal(t, 80.0, alert.Threshold)
			},
		},
		{
			name:       "success - get memory alert rule",
			ruleID:     "2",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":          2,
					"name":        "Low Memory Alert",
					"metric_name": "memory_available",
					"condition":   "less_than",
					"threshold":   20.0,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alert *Alert) {
				assert.Equal(t, "Low Memory Alert", alert.Name)
				assert.Equal(t, "less_than", alert.Condition)
			},
		},
		{
			name:       "success - get disk alert rule",
			ruleID:     "3",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":          3,
					"name":        "Disk Usage High",
					"metric_name": "disk_usage",
					"condition":   "greater_than_or_equal",
					"threshold":   90.0,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alert *Alert) {
				assert.Equal(t, "greater_than_or_equal", alert.Condition)
			},
		},
		{
			name:       "not found - invalid rule id",
			ruleID:     "999",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"error": "alert rule not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			ruleID:     "1",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"error": "invalid token",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			ruleID:     "1",
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"error": "internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)

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
			alert, err := client.Alerts.Get(ctx, tt.ruleID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, alert)
				if tt.checkFunc != nil {
					tt.checkFunc(t, alert)
				}
			}
		})
	}
}

// TestAlertsService_Update_Rules tests the Update method for alert rules
func TestAlertsService_Update_Rules(t *testing.T) {
	tests := []struct {
		name       string
		ruleID     string
		alert      *Alert
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Alert)
	}{
		{
			name:   "success - update threshold value",
			ruleID: "1",
			alert: &Alert{
				Name:      "High CPU Alert",
				Threshold: 85.0,
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":        1,
					"name":      "High CPU Alert",
					"threshold": 85.0,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alert *Alert) {
				assert.Equal(t, 85.0, alert.Threshold)
			},
		},
		{
			name:   "success - change condition operator",
			ruleID: "1",
			alert: &Alert{
				Condition: "greater_than_or_equal",
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":        1,
					"condition": "greater_than_or_equal",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alert *Alert) {
				assert.Equal(t, "greater_than_or_equal", alert.Condition)
			},
		},
		{
			name:   "success - update metric name",
			ruleID: "1",
			alert: &Alert{
				MetricName: "memory_usage",
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":            1,
					"metric_name":   "memory_usage",
				},
			},
			wantErr: false,
		},
		{
			name:   "success - update severity",
			ruleID: "1",
			alert: &Alert{
				Severity: "critical",
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":       1,
					"severity": "critical",
				},
			},
			wantErr: false,
		},
		{
			name:   "success - disable rule",
			ruleID: "1",
			alert: &Alert{
				Enabled: false,
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":      1,
					"enabled": false,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alert *Alert) {
				assert.False(t, alert.Enabled)
			},
		},
		{
			name:   "success - update channels",
			ruleID: "1",
			alert: &Alert{
				Channels: []string{"email", "slack", "webhook"},
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":       1,
					"channels": []string{"email", "slack", "webhook"},
				},
			},
			wantErr: false,
		},
		{
			name:   "validation error - invalid threshold",
			ruleID: "1",
			alert: &Alert{
				Threshold: -50.0,
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"error": "threshold must be positive",
			},
			wantErr: true,
		},
		{
			name:   "not found - invalid rule id",
			ruleID: "999",
			alert: &Alert{
				Name: "Test",
			},
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"error": "alert rule not found",
			},
			wantErr: true,
		},
		{
			name:   "unauthorized",
			ruleID: "1",
			alert: &Alert{
				Name: "Test",
			},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"error": "invalid token",
			},
			wantErr: true,
		},
		{
			name:   "server error",
			ruleID: "1",
			alert: &Alert{
				Name: "Test",
			},
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"error": "internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPut, r.Method)

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
			alert, err := client.Alerts.Update(ctx, tt.ruleID, tt.alert)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, alert)
				if tt.checkFunc != nil {
					tt.checkFunc(t, alert)
				}
			}
		})
	}
}

// TestAlertsService_Delete_Rules tests the Delete method for alert rules
func TestAlertsService_Delete_Rules(t *testing.T) {
	tests := []struct {
		name       string
		ruleID     string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:       "success - delete rule (204)",
			ruleID:     "1",
			mockStatus: http.StatusNoContent,
			mockBody:   nil,
			wantErr:    false,
		},
		{
			name:       "success - delete rule (200)",
			ruleID:     "2",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "deleted",
			},
			wantErr: false,
		},
		{
			name:       "not found - invalid rule id",
			ruleID:     "999",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"error": "alert rule not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			ruleID:     "1",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"error": "invalid token",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			ruleID:     "1",
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"error": "internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodDelete, r.Method)

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
			err = client.Alerts.Delete(ctx, tt.ruleID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestAlertsService_List_Rules tests the List method for alert rules
func TestAlertsService_List_Rules(t *testing.T) {
	tests := []struct {
		name       string
		opts       *ListOptions
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, []*Alert, *PaginationMeta)
	}{
		{
			name:       "success - list all rules",
			opts:       nil,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"id":          1,
						"name":        "High CPU Alert",
						"metric_name": "cpu_usage",
						"condition":   "greater_than",
						"threshold":   80.0,
					},
					map[string]interface{}{
						"id":          2,
						"name":        "Low Memory Alert",
						"metric_name": "memory_available",
						"condition":   "less_than",
						"threshold":   20.0,
					},
					map[string]interface{}{
						"id":          3,
						"name":        "Disk Usage High",
						"metric_name": "disk_usage",
						"condition":   "greater_than_or_equal",
						"threshold":   90.0,
					},
				},
				"pagination": map[string]interface{}{
					"page":  1,
					"limit": 25,
					"total": 3,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alerts []*Alert, meta *PaginationMeta) {
				assert.Len(t, alerts, 3)
				if meta != nil {
					assert.Equal(t, 1, meta.Page)
				}
			},
		},
		{
			name: "success - list with pagination",
			opts: &ListOptions{
				Page:  1,
				Limit: 10,
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"id":   1,
						"name": "High CPU Alert",
					},
				},
				"pagination": map[string]interface{}{
					"page":  1,
					"limit": 10,
					"total": 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alerts []*Alert, meta *PaginationMeta) {
				assert.Len(t, alerts, 1)
			},
		},
		{
			name:       "success - empty list",
			opts:       nil,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": []interface{}{},
				"pagination": map[string]interface{}{
					"page":  1,
					"limit": 25,
					"total": 0,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alerts []*Alert, meta *PaginationMeta) {
				assert.Len(t, alerts, 0)
			},
		},
		{
			name: "success - filter by metric",
			opts: &ListOptions{
				Search: "metric:cpu_usage",
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"id":          1,
						"name":        "High CPU Alert",
						"metric_name": "cpu_usage",
					},
				},
				"pagination": map[string]interface{}{
					"page":  1,
					"limit": 25,
					"total": 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alerts []*Alert, meta *PaginationMeta) {
				assert.Len(t, alerts, 1)
				assert.Equal(t, "cpu_usage", alerts[0].MetricName)
			},
		},
		{
			name:       "unauthorized",
			opts:       nil,
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"error": "invalid token",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			opts:       nil,
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"error": "internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)

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
			alerts, meta, err := client.Alerts.List(ctx, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, alerts, meta)
				}
			}
		})
	}
}
