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

// TestProbeAlertsService_List_Handler tests the List method for probe alerts
func TestProbeAlertsService_List_Handler(t *testing.T) {
	tests := []struct {
		name       string
		opts       *ProbeAlertListOptions
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, []*ProbeAlert, *PaginationMeta)
	}{
		{
			name: "success - list all probe alerts",
			opts: &ProbeAlertListOptions{
				ListOptions: ListOptions{Page: 1, Limit: 25},
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"alerts": []map[string]interface{}{
						{
							"id":       1,
							"probe_id": 10,
							"name":     "HTTP Probe Alert",
							"status":   "active",
							"message":  "Probe failed to connect",
							"conditions": map[string]interface{}{
								"failure_threshold":  3,
								"recovery_threshold": 2,
							},
							"triggered_at":      "2025-10-17T10:00:00Z",
							"notification_sent": true,
							"created_at":         "2025-10-17T10:00:00Z",
							"updated_at":         "2025-10-17T10:00:00Z",
						},
						{
							"id":       2,
							"probe_id": 11,
							"name":     "ICMP Probe Alert",
							"status":   "resolved",
							"message":  "Probe recovered",
							"conditions": map[string]interface{}{
								"failure_threshold":  3,
								"recovery_threshold": 2,
							},
							"triggered_at":      "2025-10-17T09:00:00Z",
							"resolved_at":       "2025-10-17T09:30:00Z",
							"notification_sent": true,
							"created_at":         "2025-10-17T09:00:00Z",
							"updated_at":         "2025-10-17T09:30:00Z",
						},
					},
					"pagination": map[string]interface{}{
						"page":  1,
						"limit": 25,
						"total": 2,
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alerts []*ProbeAlert, meta *PaginationMeta) {
				assert.Len(t, alerts, 2)
				assert.Equal(t, uint(1), alerts[0].ID)
				assert.Equal(t, "HTTP Probe Alert", alerts[0].Name)
				assert.Equal(t, "active", alerts[0].Status)
				assert.NotNil(t, meta)
			},
		},
		{
			name: "success - list with status filter (active)",
			opts: &ProbeAlertListOptions{
				ListOptions: ListOptions{Page: 1, Limit: 10},
				Status:      "active",
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"alerts": []map[string]interface{}{
						{
							"id":       1,
							"probe_id": 10,
							"name":     "Active Alert",
							"status":   "active",
							"message":  "Probe failed",
							"conditions": map[string]interface{}{
								"failure_threshold":  3,
								"recovery_threshold": 2,
							},
							"triggered_at":      "2025-10-17T10:00:00Z",
							"notification_sent": true,
							"created_at":         "2025-10-17T10:00:00Z",
							"updated_at":         "2025-10-17T10:00:00Z",
						},
					},
					"pagination": map[string]interface{}{
						"page":  1,
						"limit": 10,
						"total": 1,
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alerts []*ProbeAlert, meta *PaginationMeta) {
				assert.Len(t, alerts, 1)
				assert.Equal(t, "active", alerts[0].Status)
			},
		},
		{
			name: "success - list with probe_id filter",
			opts: &ProbeAlertListOptions{
				ListOptions: ListOptions{Page: 1, Limit: 10},
				ProbeID:     5,
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"alerts":     []map[string]interface{}{},
					"pagination": map[string]interface{}{"page": 1, "limit": 10, "total": 0},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alerts []*ProbeAlert, meta *PaginationMeta) {
				assert.Len(t, alerts, 0)
			},
		},
		{
			name:       "success - empty list",
			opts:       &ProbeAlertListOptions{ListOptions: ListOptions{Page: 1, Limit: 25}},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"alerts":     []map[string]interface{}{},
					"pagination": map[string]interface{}{"page": 1, "limit": 25, "total": 0},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alerts []*ProbeAlert, meta *PaginationMeta) {
				assert.Len(t, alerts, 0)
				assert.NotNil(t, meta)
			},
		},
		{
			name:       "validation error - invalid page",
			opts:       &ProbeAlertListOptions{ListOptions: ListOptions{Page: -1, Limit: 25}},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"error": "page must be positive",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized - invalid token",
			opts:       &ProbeAlertListOptions{ListOptions: ListOptions{Page: 1, Limit: 25}},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"error": "invalid token",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			opts:       &ProbeAlertListOptions{ListOptions: ListOptions{Page: 1, Limit: 25}},
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
				assert.Equal(t, "/v1/probe-alerts", r.URL.Path)

				// Validate query parameters
				query := r.URL.Query()
				if tt.opts != nil && tt.opts.Page > 0 {
					assert.Equal(t, "1", query.Get("page"))
				}
				if tt.opts != nil && tt.opts.Status != "" {
					assert.Equal(t, tt.opts.Status, query.Get("status"))
				}
				if tt.opts != nil && tt.opts.ProbeID > 0 {
					assert.Equal(t, "5", query.Get("probe_id"))
				}

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
			alerts, meta, err := client.ProbeAlerts.List(ctx, tt.opts)

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

// TestProbeAlertsService_Get_Handler tests the Get method for individual probe alerts
func TestProbeAlertsService_Get_Handler(t *testing.T) {
	tests := []struct {
		name       string
		alertID    uint
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *ProbeAlert)
	}{
		{
			name:       "success - get active probe alert",
			alertID:    1,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"alert": map[string]interface{}{
						"id":       1,
						"probe_id": 10,
						"name":     "HTTP Connection Failed",
						"status":   "active",
						"message":  "Failed to connect to target",
						"conditions": map[string]interface{}{
							"failure_threshold":  3,
							"recovery_threshold": 2,
						},
						"triggered_at":      "2025-10-17T10:15:00Z",
						"notification_sent": true,
						"created_at":         "2025-10-17T10:15:00Z",
						"updated_at":         "2025-10-17T10:15:00Z",
					},
					"probe": map[string]interface{}{
						"id":   10,
						"uuid": "probe-uuid-123",
						"name": "Production HTTP Check",
						"type": "http",
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alert *ProbeAlert) {
				assert.Equal(t, uint(1), alert.ID)
				assert.Equal(t, uint(10), alert.ProbeID)
				assert.Equal(t, "HTTP Connection Failed", alert.Name)
				assert.Equal(t, "active", alert.Status)
				assert.True(t, alert.NotificationSent)
			},
		},
		{
			name:       "success - get acknowledged probe alert",
			alertID:    2,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"alert": map[string]interface{}{
						"id":       2,
						"probe_id": 11,
						"name":     "DNS Resolution Failed",
						"status":   "acknowledged",
						"message":  "DNS server unreachable",
						"conditions": map[string]interface{}{
							"failure_threshold":  3,
							"recovery_threshold": 2,
						},
						"triggered_at":      "2025-10-17T09:00:00Z",
						"acknowledged_by":   uint(5),
						"acknowledged_at":   "2025-10-17T09:30:00Z",
						"notification_sent": true,
						"created_at":         "2025-10-17T09:00:00Z",
						"updated_at":         "2025-10-17T09:30:00Z",
					},
					"probe": map[string]interface{}{
						"id":   11,
						"uuid": "probe-uuid-456",
						"name": "DNS Health Check",
						"type": "dns",
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alert *ProbeAlert) {
				assert.Equal(t, uint(2), alert.ID)
				assert.Equal(t, "acknowledged", alert.Status)
				assert.NotNil(t, alert.AcknowledgedBy)
				assert.Equal(t, uint(5), *alert.AcknowledgedBy)
			},
		},
		{
			name:       "success - get resolved probe alert",
			alertID:    3,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"alert": map[string]interface{}{
						"id":       3,
						"probe_id": 12,
						"name":     "ICMP Timeout",
						"status":   "resolved",
						"message":  "Probe recovered",
						"conditions": map[string]interface{}{
							"failure_threshold":  3,
							"recovery_threshold": 2,
						},
						"triggered_at": "2025-10-17T08:00:00Z",
						"resolved_at":  "2025-10-17T08:45:00Z",
						"resolution":   "Service restarted",
						"created_at":   "2025-10-17T08:00:00Z",
						"updated_at":   "2025-10-17T08:45:00Z",
					},
					"probe": map[string]interface{}{
						"id":   12,
						"uuid": "probe-uuid-789",
						"name": "Server Ping",
						"type": "icmp",
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alert *ProbeAlert) {
				assert.Equal(t, uint(3), alert.ID)
				assert.Equal(t, "resolved", alert.Status)
				assert.NotNil(t, alert.ResolvedAt)
				assert.NotNil(t, alert.Resolution)
				assert.Equal(t, "Service restarted", *alert.Resolution)
			},
		},
		{
			name:       "not found - alert does not exist",
			alertID:    999,
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"error": "probe alert not found",
			},
			wantErr: true,
		},
		{
			name:       "forbidden - no access to alert",
			alertID:    1,
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"error": "insufficient permissions",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized - invalid token",
			alertID:    1,
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"error": "invalid token",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			alertID:    1,
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
				assert.Contains(t, r.URL.Path, "/v1/probe-alerts/")

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
			alert, err := client.ProbeAlerts.Get(ctx, tt.alertID)

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

// TestProbeAlertsService_Acknowledge_Handler tests the Acknowledge method
func TestProbeAlertsService_Acknowledge_Handler(t *testing.T) {
	tests := []struct {
		name       string
		alertID    uint
		note       string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *ProbeAlert)
	}{
		{
			name:       "success - acknowledge with note",
			alertID:    1,
			note:       "Investigating the issue",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"alert": map[string]interface{}{
						"id":       1,
						"probe_id": 10,
						"name":     "HTTP Alert",
						"status":   "acknowledged",
						"message":  "Connection failed",
						"conditions": map[string]interface{}{
							"failure_threshold":  3,
							"recovery_threshold": 2,
						},
						"triggered_at":      "2025-10-17T10:00:00Z",
						"acknowledged_by":   uint(5),
						"acknowledged_at":   "2025-10-17T10:05:00Z",
						"notification_sent": true,
						"created_at":         "2025-10-17T10:00:00Z",
						"updated_at":         "2025-10-17T10:05:00Z",
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alert *ProbeAlert) {
				assert.Equal(t, "acknowledged", alert.Status)
				assert.NotNil(t, alert.AcknowledgedBy)
				assert.NotNil(t, alert.AcknowledgedAt)
			},
		},
		{
			name:       "success - acknowledge without note",
			alertID:    2,
			note:       "",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"alert": map[string]interface{}{
						"id":       2,
						"status":   "acknowledged",
						"acknowledged_by":   uint(5),
						"acknowledged_at":   "2025-10-17T10:10:00Z",
						"notification_sent": true,
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alert *ProbeAlert) {
				assert.Equal(t, "acknowledged", alert.Status)
			},
		},
		{
			name:       "not found - alert does not exist",
			alertID:    999,
			note:       "Note",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"error": "probe alert not found",
			},
			wantErr: true,
		},
		{
			name:       "conflict - already resolved",
			alertID:    1,
			note:       "Note",
			mockStatus: http.StatusConflict,
			mockBody: map[string]interface{}{
				"error": "cannot acknowledge a resolved alert",
			},
			wantErr: true,
		},
		{
			name:       "forbidden - insufficient permissions",
			alertID:    1,
			note:       "Note",
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"error": "insufficient permissions",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized - invalid token",
			alertID:    1,
			note:       "Note",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"error": "invalid token",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			alertID:    1,
			note:       "Note",
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
				assert.Contains(t, r.URL.Path, "/v1/probe-alerts/")
				assert.Contains(t, r.URL.Path, "/acknowledge")

				// Validate request body if note is provided
				if tt.note != "" {
					var body map[string]string
					err := json.NewDecoder(r.Body).Decode(&body)
					require.NoError(t, err)
					assert.Equal(t, tt.note, body["note"])
				}

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
			alert, err := client.ProbeAlerts.Acknowledge(ctx, tt.alertID, tt.note)

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

// TestProbeAlertsService_Resolve_Handler tests the Resolve method
func TestProbeAlertsService_Resolve_Handler(t *testing.T) {
	tests := []struct {
		name       string
		alertID    uint
		resolution string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *ProbeAlert)
	}{
		{
			name:       "success - resolve with resolution",
			alertID:    1,
			resolution: "Issue fixed by restarting service",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"alert": map[string]interface{}{
						"id":       1,
						"probe_id": 10,
						"name":     "HTTP Alert",
						"status":   "resolved",
						"message":  "Connection recovered",
						"conditions": map[string]interface{}{
							"failure_threshold":  3,
							"recovery_threshold": 2,
						},
						"triggered_at":      "2025-10-17T10:00:00Z",
						"resolved_at":       "2025-10-17T10:20:00Z",
						"resolution":        "Issue fixed by restarting service",
						"notification_sent": true,
						"created_at":         "2025-10-17T10:00:00Z",
						"updated_at":         "2025-10-17T10:20:00Z",
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alert *ProbeAlert) {
				assert.Equal(t, "resolved", alert.Status)
				assert.NotNil(t, alert.ResolvedAt)
				assert.NotNil(t, alert.Resolution)
				assert.Equal(t, "Issue fixed by restarting service", *alert.Resolution)
			},
		},
		{
			name:       "success - resolve without resolution",
			alertID:    2,
			resolution: "",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"alert": map[string]interface{}{
						"id":          2,
						"status":      "resolved",
						"resolved_at": "2025-10-17T10:30:00Z",
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alert *ProbeAlert) {
				assert.Equal(t, "resolved", alert.Status)
			},
		},
		{
			name:       "not found - alert does not exist",
			alertID:    999,
			resolution: "Resolution",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"error": "probe alert not found",
			},
			wantErr: true,
		},
		{
			name:       "conflict - already resolved",
			alertID:    1,
			resolution: "Resolution",
			mockStatus: http.StatusConflict,
			mockBody: map[string]interface{}{
				"error": "alert is already resolved",
			},
			wantErr: true,
		},
		{
			name:       "forbidden - insufficient permissions",
			alertID:    1,
			resolution: "Resolution",
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"error": "insufficient permissions",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized - invalid token",
			alertID:    1,
			resolution: "Resolution",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"error": "invalid token",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			alertID:    1,
			resolution: "Resolution",
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
				assert.Contains(t, r.URL.Path, "/v1/probe-alerts/")
				assert.Contains(t, r.URL.Path, "/resolve")

				// Validate request body if resolution is provided
				if tt.resolution != "" {
					var body map[string]string
					err := json.NewDecoder(r.Body).Decode(&body)
					require.NoError(t, err)
					assert.Equal(t, tt.resolution, body["resolution"])
				}

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
			alert, err := client.ProbeAlerts.Resolve(ctx, tt.alertID, tt.resolution)

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

// TestProbeAlertsService_ListAdmin_Handler tests the ListAdmin method for admin access
func TestProbeAlertsService_ListAdmin_Handler(t *testing.T) {
	tests := []struct {
		name       string
		opts       *AdminProbeAlertListOptions
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, []*AdminProbeAlert, *PaginationMeta)
	}{
		{
			name: "success - admin list all alerts",
			opts: &AdminProbeAlertListOptions{
				ListOptions: ListOptions{Page: 1, Limit: 25},
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"alerts": []map[string]interface{}{
						{
							"id":                  1,
							"probe_id":            10,
							"name":                "Alert 1",
							"status":              "active",
							"message":             "Failed",
							"organization_id":     1,
							"organization_name":   "Org A",
							"probe_name":          "HTTP Check",
							"probe_type":          "http",
							"probe_target":        "https://example.com",
							"notification_sent":   true,
							"triggered_at":        "2025-10-17T10:00:00Z",
							"created_at":          "2025-10-17T10:00:00Z",
							"updated_at":          "2025-10-17T10:00:00Z",
							"conditions":          map[string]interface{}{"failure_threshold": 3},
						},
						{
							"id":                  2,
							"probe_id":            11,
							"name":                "Alert 2",
							"status":              "resolved",
							"message":             "Recovered",
							"organization_id":     2,
							"organization_name":   "Org B",
							"probe_name":          "DNS Check",
							"probe_type":          "dns",
							"probe_target":        "8.8.8.8",
							"notification_sent":   true,
							"triggered_at":        "2025-10-17T09:00:00Z",
							"resolved_at":         "2025-10-17T09:30:00Z",
							"created_at":          "2025-10-17T09:00:00Z",
							"updated_at":          "2025-10-17T09:30:00Z",
							"conditions":          map[string]interface{}{"failure_threshold": 3},
						},
					},
					"pagination": map[string]interface{}{
						"page":  1,
						"limit": 25,
						"total": 2,
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alerts []*AdminProbeAlert, meta *PaginationMeta) {
				assert.Len(t, alerts, 2)
				assert.Equal(t, uint(1), alerts[0].ID)
				assert.Equal(t, "Org A", alerts[0].OrganizationName)
				assert.Equal(t, "http", alerts[0].ProbeType)
				assert.NotNil(t, meta)
			},
		},
		{
			name: "success - admin list with status filter",
			opts: &AdminProbeAlertListOptions{
				ListOptions: ListOptions{Page: 1, Limit: 10},
				Status:      "active",
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"alerts": []map[string]interface{}{
						{
							"id":                1,
							"status":            "active",
							"organization_id":   1,
							"organization_name": "Org A",
						},
					},
					"pagination": map[string]interface{}{
						"page":  1,
						"limit": 10,
						"total": 1,
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alerts []*AdminProbeAlert, meta *PaginationMeta) {
				assert.Len(t, alerts, 1)
				assert.Equal(t, "active", alerts[0].Status)
			},
		},
		{
			name: "success - admin list with org filter",
			opts: &AdminProbeAlertListOptions{
				ListOptions:    ListOptions{Page: 1, Limit: 10},
				OrganizationID: 1,
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"alerts":     []map[string]interface{}{},
					"pagination": map[string]interface{}{"page": 1, "limit": 10, "total": 0},
				},
			},
			wantErr: false,
		},
		{
			name:       "forbidden - not admin",
			opts:       &AdminProbeAlertListOptions{ListOptions: ListOptions{Page: 1, Limit: 25}},
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"error": "admin access required",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized - invalid token",
			opts:       &AdminProbeAlertListOptions{ListOptions: ListOptions{Page: 1, Limit: 25}},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"error": "invalid token",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/v1/admin/probe-alerts", r.URL.Path)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL:    server.URL,
				Auth:       AuthConfig{Token: "admin-token"},
				RetryCount: 0,
			})
			require.NoError(t, err)

			ctx := context.Background()
			alerts, meta, err := client.ProbeAlerts.ListAdmin(ctx, tt.opts)

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

// BenchmarkProbeAlertsService benchmarks probe alerts operations
func BenchmarkProbeAlertsService(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"alerts": []map[string]interface{}{
					{
						"id":       1,
						"probe_id": 10,
						"name":     "Test Alert",
						"status":   "active",
						"conditions": map[string]interface{}{
							"failure_threshold":  3,
							"recovery_threshold": 2,
						},
					},
				},
				"alert": map[string]interface{}{
					"id":       1,
					"status":   "acknowledged",
					"acknowledged_by": uint(5),
				},
				"pagination": map[string]interface{}{
					"page":  1,
					"limit": 25,
					"total": 1,
				},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{
		BaseURL:    server.URL,
		Auth:       AuthConfig{Token: "test"},
		RetryCount: 0,
	})

	ctx := context.Background()

	b.ResetTimer()
	b.Run("List", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			client.ProbeAlerts.List(ctx, &ProbeAlertListOptions{
				ListOptions: ListOptions{Page: 1, Limit: 25},
			})
		}
	})

	b.Run("Get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			client.ProbeAlerts.Get(ctx, 1)
		}
	})

	b.Run("Acknowledge", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			client.ProbeAlerts.Acknowledge(ctx, 1, "Note")
		}
	})

	b.Run("Resolve", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			client.ProbeAlerts.Resolve(ctx, 1, "Resolution")
		}
	})
}
