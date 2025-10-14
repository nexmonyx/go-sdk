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

// ========================================
// STANDARD SERVICE METHODS TESTS
// ========================================

// TestProbesService_Create tests the Create method
func TestProbesService_Create(t *testing.T) {
	tests := []struct {
		name        string
		request     *ProbeCreateRequest
		statusCode  int
		response    *MonitoringProbe
		expectError bool
	}{
		{
			name: "create icmp probe successfully",
			request: &ProbeCreateRequest{
				Name:       "ICMP Probe Test",
				Type:       "icmp",
				Target:     "8.8.8.8",
				RegionCode: "us-east-1",
				Interval:   60,
				Enabled:    true,
			},
			statusCode: http.StatusOK,
			response: &MonitoringProbe{
				GormModel: GormModel{ID: 1},
				Name:      "ICMP Probe Test",
				Type:      "icmp",
				Target:    "8.8.8.8",
				Enabled:   true,
				Interval:  60,
			},
			expectError: false,
		},
		{
			name: "create http probe with configuration",
			request: &ProbeCreateRequest{
				Name:       "HTTP Probe Test",
				Type:       "http",
				Target:     "https://example.com",
				RegionCode: "us-west-2",
				Interval:   300,
				Enabled:    true,
				Configuration: map[string]interface{}{
					"method":  "GET",
					"timeout": 10,
				},
			},
			statusCode: http.StatusOK,
			response: &MonitoringProbe{
				GormModel: GormModel{ID: 2},
				Name:      "HTTP Probe Test",
				Type:      "http",
				Target:    "https://example.com",
				Enabled:   true,
				Interval:  300,
			},
			expectError: false,
		},
		{
			name: "create tcp probe with port",
			request: &ProbeCreateRequest{
				Name:       "TCP Probe Test",
				Type:       "tcp",
				Target:     "database.example.com",
				RegionCode: "eu-west-1",
				Interval:   120,
				Enabled:    true,
				Configuration: map[string]interface{}{
					"port": 5432,
				},
			},
			statusCode: http.StatusOK,
			response: &MonitoringProbe{
				GormModel: GormModel{ID: 3},
				Name:      "TCP Probe Test",
				Type:      "tcp",
				Target:    "database.example.com",
				Enabled:   true,
				Interval:  120,
			},
			expectError: false,
		},
		{
			name: "create heartbeat probe",
			request: &ProbeCreateRequest{
				Name:       "Heartbeat Probe",
				Type:       "heartbeat",
				Target:     "https://heartbeat.example.com",
				RegionCode: "ap-south-1",
				Interval:   600,
				Enabled:    true,
			},
			statusCode: http.StatusOK,
			response: &MonitoringProbe{
				GormModel: GormModel{ID: 4},
				Name:      "Heartbeat Probe",
				Type:      "heartbeat",
				Target:    "https://heartbeat.example.com",
				Enabled:   true,
				Interval:  600,
			},
			expectError: false,
		},
		{
			name: "create https probe",
			request: &ProbeCreateRequest{
				Name:       "HTTPS Probe Test",
				Type:       "https",
				Target:     "https://secure.example.com",
				RegionCode: "us-east-1",
				Interval:   180,
				Enabled:    false,
			},
			statusCode: http.StatusOK,
			response: &MonitoringProbe{
				GormModel: GormModel{ID: 5},
				Name:      "HTTPS Probe Test",
				Type:      "https",
				Target:    "https://secure.example.com",
				Enabled:   false,
				Interval:  180,
			},
			expectError: false,
		},
		{
			name: "create probe with empty target",
			request: &ProbeCreateRequest{
				Name:       "Empty Target Probe",
				Type:       "icmp",
				Target:     "",
				RegionCode: "us-east-1",
				Interval:   60,
				Enabled:    true,
			},
			statusCode:  http.StatusBadRequest,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/probes")

				// Verify request body structure
				var body map[string]interface{}
				err := json.NewDecoder(r.Body).Decode(&body)
				require.NoError(t, err)

				if tt.statusCode == http.StatusOK {
					assert.Equal(t, tt.request.Name, body["name"])
					assert.Equal(t, tt.request.Type, body["type"])
					assert.Equal(t, float64(tt.request.Interval), body["frequency"])
					assert.Equal(t, tt.request.Enabled, body["enabled"])

					response := struct {
						Status string `json:"status"`
						Data   struct {
							Probe MonitoringProbe `json:"probe"`
						} `json:"data"`
						Message string `json:"message"`
					}{
						Status: "success",
						Data: struct {
							Probe MonitoringProbe `json:"probe"`
						}{Probe: *tt.response},
						Message: "Probe created successfully",
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tt.statusCode)
					json.NewEncoder(w).Encode(response)
				} else {
					w.WriteHeader(tt.statusCode)
					json.NewEncoder(w).Encode(map[string]string{
						"status":  "error",
						"message": "validation failed",
					})
				}
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			probe, err := client.Probes.Create(context.Background(), tt.request)
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, probe)
			assert.Equal(t, tt.response.Name, probe.Name)
			assert.Equal(t, tt.response.Type, probe.Type)
			assert.Equal(t, tt.response.Target, probe.Target)
			assert.Equal(t, tt.response.Enabled, probe.Enabled)
		})
	}
}

// TestProbesService_List tests the List method
func TestProbesService_List(t *testing.T) {
	tests := []struct {
		name        string
		options     *ListOptions
		probes      []*MonitoringProbe
		meta        *PaginationMeta
		expectError bool
	}{
		{
			name: "list probes with pagination",
			options: &ListOptions{
				Page:  1,
				Limit: 10,
			},
			probes: []*MonitoringProbe{
				{
					GormModel: GormModel{ID: 1},
					Name:      "Probe 1",
					Type:      "http",
					Enabled:   true,
				},
				{
					GormModel: GormModel{ID: 2},
					Name:      "Probe 2",
					Type:      "icmp",
					Enabled:   true,
				},
			},
			meta: &PaginationMeta{
				TotalItems:  2,
				PerPage:     10,
				CurrentPage: 1,
				LastPage:    1,
			},
			expectError: false,
		},
		{
			name:    "list all probes without options",
			options: nil,
			probes: []*MonitoringProbe{
				{
					GormModel: GormModel{ID: 1},
					Name:      "Probe 1",
					Type:      "tcp",
					Enabled:   false,
				},
			},
			meta: &PaginationMeta{
				TotalItems:  1,
				PerPage:     25,
				CurrentPage: 1,
				LastPage:    1,
			},
			expectError: false,
		},
		{
			name: "empty result set",
			options: &ListOptions{
				Page:  1,
				Limit: 10,
			},
			probes:      []*MonitoringProbe{},
			meta:        &PaginationMeta{TotalItems: 0, PerPage: 10, CurrentPage: 1, LastPage: 1},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)

				response := struct {
					Status string             `json:"status"`
					Data   []*MonitoringProbe `json:"data"`
					Meta   *PaginationMeta    `json:"meta"`
				}{
					Status: "success",
					Data:   tt.probes,
					Meta:   tt.meta,
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			probes, meta, err := client.Probes.List(context.Background(), tt.options)
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, len(tt.probes), len(probes))
			assert.NotNil(t, meta)
			if tt.meta != nil {
				assert.Equal(t, tt.meta.TotalItems, meta.TotalItems)
				assert.Equal(t, tt.meta.CurrentPage, meta.CurrentPage)
			}
		})
	}
}

// TestProbesService_Get tests the Get method
func TestProbesService_Get(t *testing.T) {
	tests := []struct {
		name        string
		uuid        string
		probe       *MonitoringProbe
		statusCode  int
		expectError bool
	}{
		{
			name: "get probe successfully",
			uuid: "probe-uuid-123",
			probe: &MonitoringProbe{
				GormModel: GormModel{ID: 1},
				Name:      "Test Probe",
				Type:      "http",
				Target:    "https://example.com",
				Enabled:   true,
				Interval:  300,
			},
			statusCode:  http.StatusOK,
			expectError: false,
		},
		{
			name:        "probe not found",
			uuid:        "non-existent-uuid",
			statusCode:  http.StatusNotFound,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, tt.uuid)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)

				if tt.statusCode == http.StatusOK {
					response := struct {
						Status string           `json:"status"`
						Data   *MonitoringProbe `json:"data"`
					}{
						Status: "success",
						Data:   tt.probe,
					}
					json.NewEncoder(w).Encode(response)
				} else {
					json.NewEncoder(w).Encode(map[string]string{
						"status":  "error",
						"message": "probe not found",
					})
				}
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			probe, err := client.Probes.Get(context.Background(), tt.uuid)
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, probe)
			assert.Equal(t, tt.probe.Name, probe.Name)
			assert.Equal(t, tt.probe.Type, probe.Type)
		})
	}
}

// TestProbesService_Update tests the Update method
func TestProbesService_Update(t *testing.T) {
	// Using helper functions from monitoring_types_test.go
	tests := []struct {
		name        string
		uuid        string
		request     *ProbeUpdateRequest
		response    *MonitoringProbe
		statusCode  int
		expectError bool
	}{
		{
			name: "update probe name",
			uuid: "probe-uuid-123",
			request: &ProbeUpdateRequest{
				Name: stringPtr("Updated Probe Name"),
			},
			response: &MonitoringProbe{
				GormModel: GormModel{ID: 1},
				Name:      "Updated Probe Name",
				Type:      "http",
				Enabled:   true,
				Interval:  300,
			},
			statusCode:  http.StatusOK,
			expectError: false,
		},
		{
			name: "update probe enabled status",
			uuid: "probe-uuid-456",
			request: &ProbeUpdateRequest{
				Enabled: boolPtr(false),
			},
			response: &MonitoringProbe{
				GormModel: GormModel{ID: 2},
				Name:      "Test Probe",
				Type:      "icmp",
				Enabled:   false,
				Interval:  60,
			},
			statusCode:  http.StatusOK,
			expectError: false,
		},
		{
			name: "update probe interval",
			uuid: "probe-uuid-789",
			request: &ProbeUpdateRequest{
				Interval: intPtr(600),
			},
			response: &MonitoringProbe{
				GormModel: GormModel{ID: 3},
				Name:      "Test Probe",
				Type:      "tcp",
				Enabled:   true,
				Interval:  600,
			},
			statusCode:  http.StatusOK,
			expectError: false,
		},
		{
			name: "update multiple fields",
			uuid: "probe-uuid-abc",
			request: &ProbeUpdateRequest{
				Name:     stringPtr("Multi-Update Probe"),
				Enabled:  boolPtr(true),
				Interval: intPtr(180),
				Configuration: map[string]interface{}{
					"timeout": 30,
				},
			},
			response: &MonitoringProbe{
				GormModel: GormModel{ID: 4},
				Name:      "Multi-Update Probe",
				Type:      "http",
				Enabled:   true,
				Interval:  180,
				Config: map[string]interface{}{
					"timeout": 30,
				},
			},
			statusCode:  http.StatusOK,
			expectError: false,
		},
		{
			name:        "update non-existent probe",
			uuid:        "non-existent-uuid",
			request:     &ProbeUpdateRequest{Name: stringPtr("Failed Update")},
			statusCode:  http.StatusNotFound,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Contains(t, []string{"PUT", "PATCH"}, r.Method)
				assert.Contains(t, r.URL.Path, tt.uuid)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)

				if tt.statusCode == http.StatusOK {
					response := struct {
						Status string           `json:"status"`
						Data   *MonitoringProbe `json:"data"`
					}{
						Status: "success",
						Data:   tt.response,
					}
					json.NewEncoder(w).Encode(response)
				} else {
					json.NewEncoder(w).Encode(map[string]string{
						"status":  "error",
						"message": "probe not found",
					})
				}
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			probe, err := client.Probes.Update(context.Background(), tt.uuid, tt.request)
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, probe)
			if tt.request.Name != nil {
				assert.Equal(t, *tt.request.Name, probe.Name)
			}
			if tt.request.Enabled != nil {
				assert.Equal(t, *tt.request.Enabled, probe.Enabled)
			}
			if tt.request.Interval != nil {
				assert.Equal(t, *tt.request.Interval, probe.Interval)
			}
		})
	}
}

// TestProbesService_Delete tests the Delete method
func TestProbesService_Delete(t *testing.T) {
	tests := []struct {
		name        string
		uuid        string
		statusCode  int
		expectError bool
	}{
		{
			name:        "delete probe successfully",
			uuid:        "probe-uuid-123",
			statusCode:  http.StatusOK,
			expectError: false,
		},
		{
			name:        "delete non-existent probe",
			uuid:        "non-existent-uuid",
			statusCode:  http.StatusNotFound,
			expectError: true,
		},
		{
			name:        "delete with server error",
			uuid:        "probe-uuid-error",
			statusCode:  http.StatusInternalServerError,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "DELETE", r.Method)
				assert.Contains(t, r.URL.Path, tt.uuid)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)

				if tt.statusCode == http.StatusOK {
					json.NewEncoder(w).Encode(map[string]string{
						"status":  "success",
						"message": "Probe deleted successfully",
					})
				} else {
					json.NewEncoder(w).Encode(map[string]string{
						"status":  "error",
						"message": "operation failed",
					})
				}
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			err = client.Probes.Delete(context.Background(), tt.uuid)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestProbesService_GetHealth tests the GetHealth method
func TestProbesService_GetHealth(t *testing.T) {
	tests := []struct {
		name        string
		uuid        string
		health      *ProbeHealth
		statusCode  int
		expectError bool
	}{
		{
			name: "get health with all metrics",
			uuid: "probe-uuid-123",
			health: &ProbeHealth{
				ProbeUUID:       "probe-uuid-123",
				Name:            "Test Probe",
				Type:            "http",
				Target:          "https://example.com",
				Enabled:         true,
				LastStatus:      "up",
				LastRun:         time.Now().Format(time.RFC3339),
				HealthScore:     95.5,
				Availability24h: 99.8,
				AverageResponse: 150,
				RegionStatus: []RegionHealthStatus{
					{
						Region:          "us-east-1",
						RegionName:      "US East",
						LastStatus:      "up",
						Availability24h: 100.0,
						AverageResponse: 120,
					},
					{
						Region:          "eu-west-1",
						RegionName:      "EU West",
						LastStatus:      "up",
						Availability24h: 99.5,
						AverageResponse: 180,
					},
				},
			},
			statusCode:  http.StatusOK,
			expectError: false,
		},
		{
			name: "get health for degraded probe",
			uuid: "probe-uuid-456",
			health: &ProbeHealth{
				ProbeUUID:       "probe-uuid-456",
				Name:            "Degraded Probe",
				Type:            "icmp",
				Target:          "8.8.8.8",
				Enabled:         true,
				LastStatus:      "degraded",
				HealthScore:     60.0,
				Availability24h: 75.0,
				AverageResponse: 300,
			},
			statusCode:  http.StatusOK,
			expectError: false,
		},
		{
			name:        "probe not found",
			uuid:        "non-existent-uuid",
			statusCode:  http.StatusNotFound,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/probes/"+tt.uuid+"/health")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)

				if tt.statusCode == http.StatusOK {
					response := struct {
						Status string       `json:"status"`
						Data   *ProbeHealth `json:"data"`
					}{
						Status: "success",
						Data:   tt.health,
					}
					json.NewEncoder(w).Encode(response)
				} else {
					json.NewEncoder(w).Encode(map[string]string{
						"status":  "error",
						"message": "probe not found",
					})
				}
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			health, err := client.Probes.GetHealth(context.Background(), tt.uuid)
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, health)
			assert.Equal(t, tt.health.ProbeUUID, health.ProbeUUID)
			assert.Equal(t, tt.health.Name, health.Name)
			assert.Equal(t, tt.health.LastStatus, health.LastStatus)
			assert.Equal(t, tt.health.HealthScore, health.HealthScore)
			assert.Equal(t, tt.health.Availability24h, health.Availability24h)
			assert.Equal(t, len(tt.health.RegionStatus), len(health.RegionStatus))
		})
	}
}

// TestProbesService_ListResults tests the ListResults method
func TestProbesService_ListResults(t *testing.T) {
	tests := []struct {
		name        string
		uuid        string
		options     *ProbeResultListOptions
		results     []*ProbeResult
		meta        *PaginationMeta
		expectError bool
	}{
		{
			name: "list results with pagination",
			uuid: "probe-uuid-123",
			options: &ProbeResultListOptions{
				ListOptions: ListOptions{
					Page:  1,
					Limit: 10,
				},
			},
			results: []*ProbeResult{
				{
					ProbeID:      1,
					Region:       "us-east-1",
					Status:       "success",
					ResponseTime: 150,
					StatusCode:   200,
				},
				{
					ProbeID:      1,
					Region:       "eu-west-1",
					Status:       "success",
					ResponseTime: 200,
					StatusCode:   200,
				},
			},
			meta: &PaginationMeta{
				TotalItems:  2,
				PerPage:     10,
				CurrentPage: 1,
				LastPage:    1,
			},
			expectError: false,
		},
		{
			name:    "empty results",
			uuid:    "probe-uuid-empty",
			options: nil,
			results: []*ProbeResult{},
			meta: &PaginationMeta{
				TotalItems:  0,
				PerPage:     25,
				CurrentPage: 1,
				LastPage:    1,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)

				response := struct {
					Status string          `json:"status"`
					Data   []*ProbeResult  `json:"data"`
					Meta   *PaginationMeta `json:"meta"`
				}{
					Status: "success",
					Data:   tt.results,
					Meta:   tt.meta,
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			results, meta, err := client.Probes.ListResults(context.Background(), tt.uuid, tt.options)
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, len(tt.results), len(results))
			assert.NotNil(t, meta)
			if tt.meta != nil {
				assert.Equal(t, tt.meta.TotalItems, meta.TotalItems)
			}
		})
	}
}

// TestProbesService_GetAvailableRegions tests the GetAvailableRegions method
func TestProbesService_GetAvailableRegions(t *testing.T) {
	tests := []struct {
		name        string
		regions     []*MonitoringRegion
		statusCode  int
		expectError bool
	}{
		{
			name: "get available regions successfully",
			regions: []*MonitoringRegion{
				{
					Code:     "us-east-1",
					Name:     "US East (N. Virginia)",
					Location: "Virginia, USA",
					Status:   "active",
				},
				{
					Code:     "eu-west-1",
					Name:     "EU West (Ireland)",
					Location: "Dublin, Ireland",
					Status:   "active",
				},
				{
					Code:     "ap-south-1",
					Name:     "Asia Pacific (Mumbai)",
					Location: "Mumbai, India",
					Status:   "active",
				},
			},
			statusCode:  http.StatusOK,
			expectError: false,
		},
		{
			name:        "empty regions list",
			regions:     []*MonitoringRegion{},
			statusCode:  http.StatusOK,
			expectError: false,
		},
		{
			name:        "server error",
			statusCode:  http.StatusInternalServerError,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/monitoring/regions")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)

				if tt.statusCode == http.StatusOK {
					response := struct {
						Status string              `json:"status"`
						Data   []*MonitoringRegion `json:"data"`
					}{
						Status: "success",
						Data:   tt.regions,
					}
					json.NewEncoder(w).Encode(response)
				} else {
					json.NewEncoder(w).Encode(map[string]string{
						"status":  "error",
						"message": "server error",
					})
				}
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			regions, err := client.Probes.GetAvailableRegions(context.Background())
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, len(tt.regions), len(regions))
			for i, region := range regions {
				assert.Equal(t, tt.regions[i].Code, region.Code)
				assert.Equal(t, tt.regions[i].Name, region.Name)
				assert.Equal(t, tt.regions[i].Location, region.Location)
			}
		})
	}
}

// TestProbesService_GetAvailableProbeTypes tests the GetAvailableProbeTypes method
func TestProbesService_GetAvailableProbeTypes(t *testing.T) {
	t.Run("get available probe types", func(t *testing.T) {
		client, err := NewClient(&Config{
			BaseURL: "https://api.nexmonyx.com",
			Auth:    AuthConfig{Token: "test-token"},
		})
		require.NoError(t, err)

		types, err := client.Probes.GetAvailableProbeTypes(context.Background())
		require.NoError(t, err)
		assert.NotEmpty(t, types)
		assert.Contains(t, types, "icmp")
		assert.Contains(t, types, "http")
		assert.Contains(t, types, "https")
		assert.Contains(t, types, "tcp")
		assert.Contains(t, types, "heartbeat")
		assert.Equal(t, 5, len(types))
	})
}

// TestProbesService_CreateSimpleProbe tests the CreateSimpleProbe method
func TestProbesService_CreateSimpleProbe(t *testing.T) {
	tests := []struct {
		name        string
		probeName   string
		probeType   string
		target      string
		regions     []string
		response    *MonitoringProbe
		statusCode  int
		expectError bool
	}{
		{
			name:      "create simple icmp probe",
			probeName: "Simple ICMP Probe",
			probeType: "icmp",
			target:    "8.8.8.8",
			regions:   []string{"us-east-1"},
			response: &MonitoringProbe{
				GormModel: GormModel{ID: 1},
				Name:      "Simple ICMP Probe",
				Type:      "icmp",
				Target:    "8.8.8.8",
				Enabled:   true,
				Interval:  300,
			},
			statusCode:  http.StatusOK,
			expectError: false,
		},
		{
			name:      "create simple http probe",
			probeName: "Simple HTTP Probe",
			probeType: "http",
			target:    "https://example.com",
			regions:   []string{"us-east-1", "eu-west-1"},
			response: &MonitoringProbe{
				GormModel: GormModel{ID: 2},
				Name:      "Simple HTTP Probe",
				Type:      "http",
				Target:    "https://example.com",
				Enabled:   true,
				Interval:  300,
			},
			statusCode:  http.StatusOK,
			expectError: false,
		},
		{
			name:      "create simple tcp probe",
			probeName: "Simple TCP Probe",
			probeType: "tcp",
			target:    "database.example.com",
			regions:   []string{"ap-south-1"},
			response: &MonitoringProbe{
				GormModel: GormModel{ID: 3},
				Name:      "Simple TCP Probe",
				Type:      "tcp",
				Target:    "database.example.com",
				Enabled:   true,
				Interval:  300,
			},
			statusCode:  http.StatusOK,
			expectError: false,
		},
		{
			name:      "create simple https probe",
			probeName: "Simple HTTPS Probe",
			probeType: "https",
			target:    "https://secure.example.com",
			regions:   []string{"us-west-2"},
			response: &MonitoringProbe{
				GormModel: GormModel{ID: 4},
				Name:      "Simple HTTPS Probe",
				Type:      "https",
				Target:    "https://secure.example.com",
				Enabled:   true,
				Interval:  300,
			},
			statusCode:  http.StatusOK,
			expectError: false,
		},
		{
			name:      "create simple heartbeat probe",
			probeName: "Simple Heartbeat Probe",
			probeType: "heartbeat",
			target:    "https://heartbeat.example.com",
			regions:   []string{"us-east-1", "eu-west-1", "ap-south-1"},
			response: &MonitoringProbe{
				GormModel: GormModel{ID: 5},
				Name:      "Simple Heartbeat Probe",
				Type:      "heartbeat",
				Target:    "https://heartbeat.example.com",
				Enabled:   true,
				Interval:  300,
			},
			statusCode:  http.StatusOK,
			expectError: false,
		},
		{
			name:        "validation error",
			probeName:   "",
			probeType:   "http",
			target:      "",
			regions:     []string{},
			statusCode:  http.StatusBadRequest,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/probes")

				// Verify request body
				var body map[string]interface{}
				err := json.NewDecoder(r.Body).Decode(&body)
				require.NoError(t, err)

				if tt.statusCode == http.StatusOK {
					assert.Equal(t, tt.probeName, body["name"])
					assert.Equal(t, tt.probeType, body["type"])
					assert.Equal(t, float64(300), body["frequency"])
					assert.Equal(t, true, body["enabled"])

					// Verify config based on probe type
					config, ok := body["config"].(map[string]interface{})
					assert.True(t, ok)
					switch tt.probeType {
					case "icmp":
						assert.Equal(t, tt.target, config["host"])
					case "http", "https", "heartbeat":
						assert.Equal(t, tt.target, config["url"])
					case "tcp":
						assert.Equal(t, tt.target, config["host"])
						assert.NotNil(t, config["port"])
					}

					response := struct {
						Status string `json:"status"`
						Data   struct {
							Probe MonitoringProbe `json:"probe"`
						} `json:"data"`
						Message string `json:"message"`
					}{
						Status: "success",
						Data: struct {
							Probe MonitoringProbe `json:"probe"`
						}{Probe: *tt.response},
						Message: "Probe created successfully",
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tt.statusCode)
					json.NewEncoder(w).Encode(response)
				} else {
					w.WriteHeader(tt.statusCode)
					json.NewEncoder(w).Encode(map[string]string{
						"status":  "error",
						"message": "validation failed",
					})
				}
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			probe, err := client.Probes.CreateSimpleProbe(context.Background(), tt.probeName, tt.probeType, tt.target, tt.regions)
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, probe)
			assert.Equal(t, tt.response.Name, probe.Name)
			assert.Equal(t, tt.response.Type, probe.Type)
			assert.Equal(t, tt.response.Target, probe.Target)
			assert.Equal(t, tt.response.Enabled, probe.Enabled)
		})
	}
}

// TestProbesService_ListByOrganization tests the ListByOrganization method
func TestProbesService_ListByOrganization(t *testing.T) {
	tests := []struct {
		name           string
		orgID          uint
		responseProbes []*MonitoringProbe
		expectError    bool
	}{
		{
			name:  "successful list with multiple probes",
			orgID: 1,
			responseProbes: []*MonitoringProbe{
				{
					GormModel: GormModel{ID: 1},
					Name:      "test-probe-1",
					Type:      "http",
					Target:    "https://example.com",
					Enabled:   true,
					Interval:  60,
				},
				{
					GormModel: GormModel{ID: 2},
					Name:      "test-probe-2",
					Type:      "icmp",
					Target:    "8.8.8.8",
					Enabled:   true,
					Interval:  30,
				},
			},
			expectError: false,
		},
		{
			name:           "empty result for organization",
			orgID:          999,
			responseProbes: []*MonitoringProbe{},
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/controllers/probes/list")

				// Verify org_id query parameter
				orgID := r.URL.Query().Get("org_id")
				assert.NotEmpty(t, orgID)

				response := struct {
					Status string             `json:"status"`
					Data   []*MonitoringProbe `json:"data"`
				}{
					Status: "success",
					Data:   tt.responseProbes,
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth: AuthConfig{
					MonitoringKey: "test-monitoring-key",
				},
			})
			require.NoError(t, err)

			probes, err := client.Probes.ListByOrganization(context.Background(), tt.orgID)
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, len(tt.responseProbes), len(probes))

			for i, probe := range probes {
				assert.Equal(t, tt.responseProbes[i].Name, probe.Name)
			}
		})
	}
}

// TestProbesService_GetByUUID tests the GetByUUID method
func TestProbesService_GetByUUID(t *testing.T) {
	tests := []struct {
		name        string
		probeUUID   string
		probe       *MonitoringProbe
		expectError bool
	}{
		{
			name:      "successful get probe",
			probeUUID: "test-probe-uuid",
			probe: &MonitoringProbe{
				GormModel: GormModel{ID: 1},
				Name:      "test-probe",
				Type:      "http",
				Target:    "https://example.com",
				Enabled:   true,
				Interval:  60,
				Regions:   []string{"us-east-1", "eu-west-1"},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)

				response := struct {
					Status string           `json:"status"`
					Data   *MonitoringProbe `json:"data"`
				}{
					Status: "success",
					Data:   tt.probe,
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth: AuthConfig{
					MonitoringKey: "test-monitoring-key",
				},
			})
			require.NoError(t, err)

			probe, err := client.Probes.GetByUUID(context.Background(), tt.probeUUID)
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.probe.Name, probe.Name)
			assert.Equal(t, tt.probe.Type, probe.Type)
		})
	}
}

// TestProbesService_GetRegionalResults tests the GetRegionalResults method
func TestProbesService_GetRegionalResults(t *testing.T) {
	tests := []struct {
		name           string
		probeUUID      string
		regionalResult []RegionalResult
		expectError    bool
	}{
		{
			name:      "successful get regional results with multiple regions",
			probeUUID: "test-probe-uuid",
			regionalResult: []RegionalResult{
				{
					Region:       "us-east-1",
					Status:       "up",
					ResponseTime: 150,
					CheckedAt:    time.Now().Format(time.RFC3339),
				},
				{
					Region:       "eu-west-1",
					Status:       "up",
					ResponseTime: 200,
					CheckedAt:    time.Now().Format(time.RFC3339),
				},
				{
					Region:       "ap-south-1",
					Status:       "down",
					ResponseTime: 0,
					ErrorMessage: stringPtr("Connection timeout"),
					CheckedAt:    time.Now().Format(time.RFC3339),
				},
			},
			expectError: false,
		},
		{
			name:           "empty results",
			probeUUID:      "probe-with-no-results",
			regionalResult: []RegionalResult{},
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/controllers/probes/"+tt.probeUUID+"/regional-results")

				response := struct {
					Status string           `json:"status"`
					Data   []RegionalResult `json:"data"`
				}{
					Status: "success",
					Data:   tt.regionalResult,
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth: AuthConfig{
					MonitoringKey: "test-monitoring-key",
				},
			})
			require.NoError(t, err)

			results, err := client.Probes.GetRegionalResults(context.Background(), tt.probeUUID)
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, len(tt.regionalResult), len(results))

			for i, result := range results {
				assert.Equal(t, tt.regionalResult[i].Region, result.Region)
				assert.Equal(t, tt.regionalResult[i].Status, result.Status)
			}
		})
	}
}

// TestProbesService_UpdateControllerStatus tests the UpdateControllerStatus method
func TestProbesService_UpdateControllerStatus(t *testing.T) {
	tests := []struct {
		name        string
		probeUUID   string
		status      string
		expectError bool
	}{
		{
			name:        "update status to up",
			probeUUID:   "test-probe-uuid",
			status:      "up",
			expectError: false,
		},
		{
			name:        "update status to down",
			probeUUID:   "test-probe-uuid",
			status:      "down",
			expectError: false,
		},
		{
			name:        "update status to degraded",
			probeUUID:   "test-probe-uuid",
			status:      "degraded",
			expectError: false,
		},
		{
			name:        "update status to unknown",
			probeUUID:   "test-probe-uuid",
			status:      "unknown",
			expectError: false,
		},
		{
			name:        "invalid status",
			probeUUID:   "test-probe-uuid",
			status:      "invalid-status",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For invalid status test, skip server setup and test validation directly
			if tt.status == "invalid-status" {
				client, err := NewClient(&Config{
					BaseURL: "https://api.nexmonyx.com",
					Auth: AuthConfig{
						MonitoringKey: "test-monitoring-key",
					},
				})
				require.NoError(t, err)

				err = client.Probes.UpdateControllerStatus(context.Background(), tt.probeUUID, tt.status)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid status")
				return
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "PUT", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/controllers/probes/"+tt.probeUUID+"/status")

				// Verify request body
				var body map[string]interface{}
				json.NewDecoder(r.Body).Decode(&body)
				assert.Equal(t, tt.status, body["status"])

				response := struct {
					Status  string `json:"status"`
					Message string `json:"message"`
				}{
					Status:  "success",
					Message: "Status updated successfully",
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth: AuthConfig{
					MonitoringKey: "test-monitoring-key",
				},
			})
			require.NoError(t, err)

			err = client.Probes.UpdateControllerStatus(context.Background(), tt.probeUUID, tt.status)
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

// TestProbesService_GetProbeConfig tests the GetProbeConfig method
func TestProbesService_GetProbeConfig(t *testing.T) {
	tests := []struct {
		name              string
		probeUUID         string
		probe             *MonitoringProbe
		expectedConsensus string
		expectError       bool
	}{
		{
			name:      "probe with explicit consensus type",
			probeUUID: "test-probe-uuid-1",
			probe: &MonitoringProbe{
				GormModel: GormModel{ID: 1},
				Name:      "test-probe-1",
				Type:      "http",
				Target:    "https://example.com",
				Enabled:   true,
				Interval:  60,
				Timeout:   10,
				Regions:   []string{"us-east-1", "eu-west-1", "ap-south-1"},
				Config: map[string]interface{}{
					"consensus_type": "all",
				},
			},
			expectedConsensus: "all",
			expectError:       false,
		},
		{
			name:      "probe without consensus type (defaults to majority)",
			probeUUID: "test-probe-uuid-2",
			probe: &MonitoringProbe{
				GormModel: GormModel{ID: 2},
				Name:      "test-probe-2",
				Type:      "icmp",
				Target:    "8.8.8.8",
				Enabled:   true,
				Interval:  30,
				Timeout:   5,
				Regions:   []string{"us-east-1", "eu-west-1"},
				Config:    map[string]interface{}{},
			},
			expectedConsensus: "majority",
			expectError:       false,
		},
		{
			name:      "probe with any consensus type",
			probeUUID: "test-probe-uuid-3",
			probe: &MonitoringProbe{
				GormModel: GormModel{ID: 3},
				Name:      "test-probe-3",
				Type:      "tcp",
				Target:    "database.example.com:5432",
				Enabled:   true,
				Interval:  120,
				Timeout:   15,
				Regions:   []string{"us-east-1"},
				Config: map[string]interface{}{
					"consensus_type": "any",
				},
			},
			expectedConsensus: "any",
			expectError:       false,
		},
		{
			name:      "probe with nil config",
			probeUUID: "test-probe-uuid-4",
			probe: &MonitoringProbe{
				GormModel: GormModel{ID: 4},
				Name:      "test-probe-4",
				Type:      "http",
				Target:    "https://example.com",
				Enabled:   true,
				Interval:  60,
				Timeout:   10,
				Regions:   []string{"us-east-1"},
				Config:    nil,
			},
			expectedConsensus: "majority",
			expectError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)

				response := struct {
					Status string           `json:"status"`
					Data   *MonitoringProbe `json:"data"`
				}{
					Status: "success",
					Data:   tt.probe,
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth: AuthConfig{
					MonitoringKey: "test-monitoring-key",
				},
			})
			require.NoError(t, err)

			config, err := client.Probes.GetProbeConfig(context.Background(), tt.probeUUID)
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.probeUUID, config.ProbeUUID)
			assert.Equal(t, tt.expectedConsensus, config.ConsensusType)
			assert.Equal(t, tt.probe.Name, config.Name)
			assert.Equal(t, tt.probe.Type, config.Type)
			assert.Equal(t, len(tt.probe.Regions), len(config.Regions))
		})
	}
}

// TestProbesService_RecordConsensusResult tests the RecordConsensusResult method
func TestProbesService_RecordConsensusResult(t *testing.T) {
	tests := []struct {
		name        string
		result      *ConsensusResultRequest
		expectError bool
	}{
		{
			name: "successful consensus result recording",
			result: &ConsensusResultRequest{
				ProbeUUID:      "test-probe-uuid",
				OrganizationID: 1,
				ConsensusType:  "majority",
				GlobalStatus:   "up",
				RegionResults: []RegionalResult{
					{
						Region:       "us-east-1",
						Status:       "up",
						ResponseTime: 150,
						CheckedAt:    time.Now().Format(time.RFC3339),
					},
					{
						Region:       "eu-west-1",
						Status:       "up",
						ResponseTime: 200,
						CheckedAt:    time.Now().Format(time.RFC3339),
					},
				},
				TotalRegions:        2,
				UpRegions:           2,
				DownRegions:         0,
				DegradedRegions:     0,
				UnknownRegions:      0,
				ShouldAlert:         false,
				AverageResponseTime: 175,
				MinResponseTime:     150,
				MaxResponseTime:     200,
				UptimePercentage:    100.0,
			},
			expectError: false,
		},
		{
			name: "consensus result with down status",
			result: &ConsensusResultRequest{
				ProbeUUID:      "test-probe-uuid-2",
				OrganizationID: 1,
				ConsensusType:  "all",
				GlobalStatus:   "down",
				RegionResults: []RegionalResult{
					{
						Region:       "us-east-1",
						Status:       "up",
						ResponseTime: 150,
						CheckedAt:    time.Now().Format(time.RFC3339),
					},
					{
						Region:       "eu-west-1",
						Status:       "down",
						ResponseTime: 0,
						ErrorMessage: stringPtr("Connection timeout"),
						CheckedAt:    time.Now().Format(time.RFC3339),
					},
				},
				TotalRegions:        2,
				UpRegions:           1,
				DownRegions:         1,
				DegradedRegions:     0,
				UnknownRegions:      0,
				ShouldAlert:         true,
				AverageResponseTime: 75,
				MinResponseTime:     0,
				MaxResponseTime:     150,
				UptimePercentage:    50.0,
			},
			expectError: false,
		},
		{
			name:        "nil consensus result",
			result:      nil,
			expectError: true,
		},
		{
			name: "empty probe UUID",
			result: &ConsensusResultRequest{
				ProbeUUID:      "",
				OrganizationID: 1,
				ConsensusType:  "majority",
				GlobalStatus:   "up",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For validation error tests, skip server setup
			if tt.result == nil || tt.result.ProbeUUID == "" {
				client, err := NewClient(&Config{
					BaseURL: "https://api.nexmonyx.com",
					Auth: AuthConfig{
						MonitoringKey: "test-monitoring-key",
					},
				})
				require.NoError(t, err)

				err = client.Probes.RecordConsensusResult(context.Background(), tt.result)
				assert.Error(t, err)
				return
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/controllers/probes/consensus-results")

				// Verify request body
				var body ConsensusResultRequest
				json.NewDecoder(r.Body).Decode(&body)
				assert.Equal(t, tt.result.ProbeUUID, body.ProbeUUID)
				assert.Equal(t, tt.result.GlobalStatus, body.GlobalStatus)

				response := struct {
					Status  string `json:"status"`
					Message string `json:"message"`
				}{
					Status:  "success",
					Message: "Consensus result recorded successfully",
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth: AuthConfig{
					MonitoringKey: "test-monitoring-key",
				},
			})
			require.NoError(t, err)

			err = client.Probes.RecordConsensusResult(context.Background(), tt.result)
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}
