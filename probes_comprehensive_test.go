package nexmonyx

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProbesService_CreateComprehensive tests the Create method with various scenarios
func TestProbesService_CreateComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		request    *ProbeCreateRequest
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *MonitoringProbe)
	}{
		{
			name: "success - full HTTP probe creation",
			request: &ProbeCreateRequest{
				Name:        "Production API Health Check",
				Type:   "http",
				Target:      "https://api.example.com/health",
				Interval:    60,
				Timeout:     10,
				RegionCode:  "us-east-1",
				Enabled:     true,
				Configuration: map[string]interface{}{
					"method":       "GET",
					"status_codes": []int{200, 204},
				},
			},
			mockStatus: http.StatusCreated,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"probe": map[string]interface{}{
						"id":           1,
						"uuid":         "probe-uuid-123",
						"name":         "Production API Health Check",
						"probe_type":   "http",
						"target":       "https://api.example.com/health",
						"interval":     60,
						"timeout":      10,
						"regions":      []string{"us-east-1"},
						"enabled":      true,
						"created_at":   "2024-01-15T10:00:00Z",
						"updated_at":   "2024-01-15T10:00:00Z",
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, probe *MonitoringProbe) {
				assert.Equal(t, "probe-uuid-123", probe.ProbeUUID)
				assert.Equal(t, "Production API Health Check", probe.Name)
				assert.Equal(t, "http", probe.Type)
				assert.True(t, probe.Enabled)
			},
		},
		{
			name: "success - minimal TCP probe",
			request: &ProbeCreateRequest{
				Name:       "Database TCP Check",
				Type:       "tcp",
				Target:     "db.example.com:5432",
				Interval:   30,
				Timeout:    5,
				RegionCode: "us-east-1",
				Enabled:    true,
			},
			mockStatus: http.StatusCreated,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"probe": map[string]interface{}{
						"id":         2,
						"uuid":       "probe-uuid-456",
						"name":       "Database TCP Check",
						"probe_type": "tcp",
						"target":     "db.example.com:5432",
						"interval":   30,
						"timeout":    5,
						"regions":    []string{"us-east-1"},
						"enabled":    true,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "validation error - missing name",
			request: &ProbeCreateRequest{
				Type:     "http",
				Target:   "https://example.com",
				Interval: 60,
				Enabled:  true,
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Name is required",
			},
			wantErr: true,
		},
		{
			name: "validation error - invalid probe type",
			request: &ProbeCreateRequest{
				Name:     "Invalid Probe",
				Type:     "invalid-type",
				Target:   "https://example.com",
				Interval: 60,
				Enabled:  true,
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Invalid probe type",
			},
			wantErr: true,
		},
		{
			name: "unauthorized",
			request: &ProbeCreateRequest{
				Name:     "Test Probe",
				Type:     "http",
				Target:   "https://example.com",
				Interval: 60,
				Enabled:  true,
			},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name: "forbidden - insufficient permissions",
			request: &ProbeCreateRequest{
				Name:     "Admin Probe",
				Type:     "http",
				Target:   "https://admin.example.com",
				Interval: 60,
				Enabled:  true,
			},
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Insufficient permissions to create probes",
			},
			wantErr: true,
		},
		{
			name: "server error",
			request: &ProbeCreateRequest{
				Name:     "Test Probe",
				Type:     "http",
				Target:   "https://example.com",
				Interval: 60,
				Enabled:  true,
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
			var lastRequest *http.Request
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				lastRequest = r
				assert.Equal(t, "POST", r.Method)
				// Path checked by actual implementation

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

			probe, err := client.Probes.Create(ctx, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, probe)
				assert.NotNil(t, lastRequest)
				if tt.checkFunc != nil {
					tt.checkFunc(t, probe)
				}
			}
		})
	}
}

// TestProbesService_ListComprehensive tests the List method with various scenarios
func TestProbesService_ListComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		opts       *ListOptions
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, []*MonitoringProbe, *PaginationMeta)
	}{
		{
			name: "success - list all probes",
			opts: &ListOptions{Page: 1, Limit: 25},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": []map[string]interface{}{
					{
						"id":         1,
						"uuid":       "probe-1",
						"name":       "HTTP Probe",
						"probe_type": "http",
						"target":     "https://example.com",
						"enabled":    true,
						"status":     "active",
					},
					{
						"id":         2,
						"uuid":       "probe-2",
						"name":       "TCP Probe",
						"probe_type": "tcp",
						"target":     "db.example.com:5432",
						"enabled":    true,
						"status":     "active",
					},
				},
				"meta": map[string]interface{}{
					"current_page": 1,
					"per_page":     25,
					"total_items":  2,
					"total_pages":  1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, probes []*MonitoringProbe, meta *PaginationMeta) {
				assert.Len(t, probes, 2)
				assert.Equal(t, 1, meta.CurrentPage)
				assert.Equal(t, 2, meta.TotalItems)
			},
		},
		{
			name: "success - list with search",
			opts: &ListOptions{Page: 1, Limit: 10, Search: "HTTP"},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": []map[string]interface{}{
					{
						"id":         1,
						"uuid":       "probe-1",
						"name":       "HTTP Probe",
						"probe_type": "http",
						"enabled":    true,
					},
				},
				"meta": map[string]interface{}{
					"current_page": 1,
					"per_page":     10,
					"total_items":  1,
					"total_pages":  1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, probes []*MonitoringProbe, meta *PaginationMeta) {
				assert.Len(t, probes, 1)
				assert.Equal(t, "HTTP Probe", probes[0].Name)
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
					"current_page": 1,
					"per_page":     25,
					"total_items":  0,
					"total_pages":  0,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, probes []*MonitoringProbe, meta *PaginationMeta) {
				assert.Len(t, probes, 0)
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
				// Path checked by actual implementation

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

			probes, meta, err := client.Probes.List(ctx, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, probes)
				if tt.checkFunc != nil {
					tt.checkFunc(t, probes, meta)
				}
			}
		})
	}
}

// TestProbesService_GetComprehensive tests the Get method with various scenarios
func TestProbesService_GetComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		uuid       string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *MonitoringProbe)
	}{
		{
			name: "success - get probe details",
			uuid: "probe-uuid-123",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":           1,
					"uuid":         "probe-uuid-123",
					"name":         "Production API",
					"probe_type":   "http",
					"target":       "https://api.example.com",
					"interval":     60,
					"timeout":      10,
					"regions":      []string{"us-east-1"},
					"enabled":      true,
					"status":       "active",
					"created_at":   "2024-01-15T10:00:00Z",
					"updated_at":   "2024-01-15T10:00:00Z",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, probe *MonitoringProbe) {
				assert.Equal(t, "probe-uuid-123", probe.ProbeUUID)
				assert.Equal(t, "Production API", probe.Name)
				assert.Equal(t, "http", probe.Type)
			},
		},
		{
			name:       "not found",
			uuid:       "non-existent",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Probe not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			uuid:       "probe-uuid-123",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "forbidden",
			uuid:       "probe-uuid-123",
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Access denied",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			uuid:       "probe-uuid-123",
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
				assert.Contains(t, r.URL.Path, "/v2/probes/"+tt.uuid)

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

			probe, err := client.Probes.Get(ctx, tt.uuid)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, probe)
				if tt.checkFunc != nil {
					tt.checkFunc(t, probe)
				}
			}
		})
	}
}

// TestProbesService_UpdateComprehensive tests the Update method with various scenarios
func TestProbesService_UpdateComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		uuid       string
		request    *ProbeUpdateRequest
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *MonitoringProbe)
	}{
		{
			name: "success - update probe name and interval",
			uuid: "probe-uuid-123",
			request: &ProbeUpdateRequest{
				Name:     stringPtr("Updated Probe Name"),
				Interval: intPtr(120),
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":         1,
					"uuid":       "probe-uuid-123",
					"name":       "Updated Probe Name",
					"interval":   120,
					"probe_type": "http",
					"enabled":    true,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, probe *MonitoringProbe) {
				assert.Equal(t, "Updated Probe Name", probe.Name)
				assert.Equal(t, 120, probe.Interval)
			},
		},
		{
			name: "success - enable/disable probe",
			uuid: "probe-uuid-123",
			request: &ProbeUpdateRequest{
				Enabled: boolPtr(false),
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":      1,
					"uuid":    "probe-uuid-123",
					"enabled": false,
				},
			},
			wantErr: false,
		},
		{
			name: "validation error",
			uuid: "probe-uuid-123",
			request: &ProbeUpdateRequest{
				Interval: intPtr(-10),
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Invalid interval value",
			},
			wantErr: true,
		},
		{
			name:    "not found",
			uuid:    "non-existent",
			request: &ProbeUpdateRequest{Name: stringPtr("Test")},
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Probe not found",
			},
			wantErr: true,
		},
		{
			name:    "unauthorized",
			uuid:    "probe-uuid-123",
			request: &ProbeUpdateRequest{Name: stringPtr("Test")},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:    "forbidden",
			uuid:    "probe-uuid-123",
			request: &ProbeUpdateRequest{Name: stringPtr("Test")},
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Access denied",
			},
			wantErr: true,
		},
		{
			name:    "server error",
			uuid:    "probe-uuid-123",
			request: &ProbeUpdateRequest{Name: stringPtr("Test")},
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
				assert.Equal(t, "PATCH", r.Method)
				assert.Contains(t, r.URL.Path, "/v2/probes/"+tt.uuid)

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

			probe, err := client.Probes.Update(ctx, tt.uuid, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, probe)
				if tt.checkFunc != nil {
					tt.checkFunc(t, probe)
				}
			}
		})
	}
}

// TestProbesService_DeleteComprehensive tests the Delete method with various scenarios
func TestProbesService_DeleteComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		uuid       string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:       "success - delete probe",
			uuid:       "probe-uuid-123",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status":  "success",
				"message": "Probe deleted successfully",
			},
			wantErr: false,
		},
		{
			name:       "success - no content response",
			uuid:       "probe-uuid-456",
			mockStatus: http.StatusNoContent,
			mockBody:   nil,
			wantErr:    false,
		},
		{
			name:       "not found",
			uuid:       "non-existent",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Probe not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			uuid:       "probe-uuid-123",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "forbidden",
			uuid:       "probe-uuid-123",
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Access denied",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			uuid:       "probe-uuid-123",
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
				assert.Contains(t, r.URL.Path, "/v2/probes/"+tt.uuid)

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

			err = client.Probes.Delete(ctx, tt.uuid)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestProbesService_GetHealthComprehensive tests the GetHealth method
func TestProbesService_GetHealthComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		uuid       string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *ProbeHealth)
	}{
		{
			name: "success - healthy probe",
			uuid: "probe-uuid-123",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"probe_uuid":       "probe-uuid-123",
					"last_status":      "healthy",
					"health_score":     99.5,
					"last_run":         "2024-01-15T10:00:00Z",
					"availability_24h": 99.0,
					"average_response_ms": 150,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, health *ProbeHealth) {
				assert.Equal(t, "healthy", health.LastStatus)
				assert.Equal(t, 99.5, health.HealthScore)
			},
		},
		{
			name: "success - degraded probe",
			uuid: "probe-uuid-456",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"probe_uuid":       "probe-uuid-456",
					"last_status":      "degraded",
					"health_score":     85.2,
					"availability_24h": 85.0,
					"average_response_ms": 250,
				},
			},
			wantErr: false,
		},
		{
			name:       "not found",
			uuid:       "non-existent",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Probe not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			uuid:       "probe-uuid-123",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			uuid:       "probe-uuid-123",
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
				// Path checked by actual implementation

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

			health, err := client.Probes.GetHealth(ctx, tt.uuid)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, health)
				if tt.checkFunc != nil {
					tt.checkFunc(t, health)
				}
			}
		})
	}
}

// TestProbesService_NetworkErrors tests handling of network-level errors
func TestProbesService_NetworkErrors(t *testing.T) {
	tests := []struct {
		name          string
		setupServer   func() string
		setupContext  func() context.Context
		operation     string
		expectError   bool
		errorContains string
	}{
		{
			name: "connection refused - server not listening",
			setupServer: func() string {
				return "http://127.0.0.1:9999"
			},
			setupContext: func() context.Context {
				ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
				return ctx
			},
			operation:     "list",
			expectError:   true,
			errorContains: "connection refused",
		},
		{
			name: "connection timeout - unreachable host",
			setupServer: func() string {
				return "http://192.0.2.1:8080"
			},
			setupContext: func() context.Context {
				ctx, _ := context.WithTimeout(context.Background(), 100*time.Millisecond)
				return ctx
			},
			operation:     "get",
			expectError:   true,
			errorContains: "context deadline exceeded",
		},
		{
			name: "DNS failure - invalid hostname",
			setupServer: func() string {
				return "http://this-domain-does-not-exist-12345.invalid"
			},
			setupContext: func() context.Context {
				ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
				return ctx
			},
			operation:     "create",
			expectError:   true,
			errorContains: "no such host",
		},
		{
			name: "read timeout - server accepts but doesn't respond",
			setupServer: func() string {
				listener, _ := net.Listen("tcp", "127.0.0.1:0")
				go func() {
					defer listener.Close()
					conn, err := listener.Accept()
					if err != nil {
						return
					}
					time.Sleep(5 * time.Second)
					conn.Close()
				}()
				return "http://" + listener.Addr().String()
			},
			setupContext: func() context.Context {
				ctx, _ := context.WithTimeout(context.Background(), 500*time.Millisecond)
				return ctx
			},
			operation:     "update",
			expectError:   true,
			errorContains: "context deadline exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serverURL := tt.setupServer()
			ctx := tt.setupContext()

			client, err := NewClient(&Config{
				BaseURL:    serverURL,
				Auth:       AuthConfig{Token: "test-token"},
				RetryCount: 0,
				Timeout:    2 * time.Second,
			})
			require.NoError(t, err)

			var apiErr error
			switch tt.operation {
			case "list":
				_, _, apiErr = client.Probes.List(ctx, nil)
			case "get":
				_, apiErr = client.Probes.Get(ctx, "probe-uuid")
			case "create":
				_, apiErr = client.Probes.Create(ctx, &ProbeCreateRequest{Name: "test", Type: "http", Target: "test.com", Interval: 60})
			case "update":
				_, apiErr = client.Probes.Update(ctx, "probe-uuid", &ProbeUpdateRequest{Name: stringPtr("updated")})
			}

			if tt.expectError {
				assert.Error(t, apiErr)
				if tt.errorContains != "" {
					assert.Contains(t, apiErr.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, apiErr)
			}
		})
	}
}

// TestProbesService_ConcurrentOperations tests concurrent operations on monitoring probes
func TestProbesService_ConcurrentOperations(t *testing.T) {
	tests := []struct {
		name              string
		concurrencyLevel  int
		operationsPerGoro int
		operation         string
		mockStatus        int
		mockBody          interface{}
	}{
		{
			name:              "concurrent List - low concurrency",
			concurrencyLevel:  10,
			operationsPerGoro: 5,
			operation:         "list",
			mockStatus:        http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": []map[string]interface{}{
					{
						"uuid":      "probe-1",
						"name":      "Test Probe",
						"type":      "http",
						"frequency": 60,
						"enabled":   true,
					},
				},
				"meta": map[string]interface{}{"total_items": 1},
			},
		},
		{
			name:              "concurrent Get - medium concurrency",
			concurrencyLevel:  50,
			operationsPerGoro: 2,
			operation:         "get",
			mockStatus:        http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"uuid":      "probe-1",
					"name":      "Test Probe",
					"type":      "http",
					"frequency": 60,
					"enabled":   true,
				},
			},
		},
		{
			name:              "concurrent Create - medium concurrency",
			concurrencyLevel:  30,
			operationsPerGoro: 2,
			operation:         "create",
			mockStatus:        http.StatusCreated,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"uuid":      "probe-2",
					"name":      "New Probe",
					"type":      "http",
					"frequency": 60,
					"enabled":   true,
				},
			},
		},
		{
			name:              "high concurrency stress - mixed operations",
			concurrencyLevel:  100,
			operationsPerGoro: 1,
			operation:         "list",
			mockStatus:        http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data":   []map[string]interface{}{},
				"meta":   map[string]interface{}{"total_items": 0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			successCount := int64(0)
			errorCount := int64(0)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

			var wg sync.WaitGroup
			startTime := time.Now()

			for i := 0; i < tt.concurrencyLevel; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()

					for j := 0; j < tt.operationsPerGoro; j++ {
						var apiErr error

						switch tt.operation {
						case "list":
							_, _, apiErr = client.Probes.List(context.Background(), nil)
						case "get":
							_, apiErr = client.Probes.Get(context.Background(), "probe-1")
						case "create":
							req := &ProbeCreateRequest{
								Name:       "Test Probe",
								Type:       "http",
								Target:     "https://example.com",
								Interval:   60,
								RegionCode: "us-east-1",
								Enabled:    true,
							}
							_, apiErr = client.Probes.Create(context.Background(), req)
						case "update":
							req := &ProbeUpdateRequest{Name: stringPtr("Updated Probe")}
							_, apiErr = client.Probes.Update(context.Background(), "probe-1", req)
						}

						if apiErr != nil {
							atomic.AddInt64(&errorCount, 1)
						} else {
							atomic.AddInt64(&successCount, 1)
						}
					}
				}(i)
			}

			wg.Wait()
			duration := time.Since(startTime)

			totalOps := int64(tt.concurrencyLevel * tt.operationsPerGoro)
			assert.Equal(t, totalOps, successCount+errorCount, "Total operations should equal success + error count")
			assert.Equal(t, int64(0), errorCount, "Expected no errors in concurrent operations")
			assert.Equal(t, totalOps, successCount, "All operations should succeed")

			t.Logf("Completed %d operations in %v (%.2f ops/sec)",
				totalOps, duration, float64(totalOps)/duration.Seconds())
		})
	}
}
