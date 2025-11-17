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

// TestServiceMonitoringService_SubmitServiceData tests submitting service monitoring data
func TestServiceMonitoringService_SubmitServiceData(t *testing.T) {
	tests := []struct {
		name           string
		serverID       string
		mockStatus     int
		mockResponse   interface{}
		wantErr        bool
		errType        string
		clientAuthFunc func(*Config)
	}{
		{
			name:       "success - submit service data",
			serverID:   "server-001",
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": map[string]interface{}{"submitted": true},
			},
			wantErr: false,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - invalid server id (400)",
			serverID:       "invalid-server",
			mockStatus:     http.StatusBadRequest,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			errType:        "ValidationError",
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - unauthorized (401)",
			serverID:       "server-001",
			mockStatus:     http.StatusUnauthorized,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			errType:        "UnauthorizedError",
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "invalid-key", APISecret: "invalid-secret"}
			},
		},
		{
			name:           "error - forbidden (403)",
			serverID:       "server-001",
			mockStatus:     http.StatusForbidden,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			errType:        "ForbiddenError",
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - server error (500)",
			serverID:       "server-001",
			mockStatus:     http.StatusInternalServerError,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			errType:        "APIError",
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - service unavailable (503)",
			serverID:       "server-001",
			mockStatus:     http.StatusServiceUnavailable,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			errType:        "APIError",
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			config := &Config{
				BaseURL: server.URL,
			}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			require.NoError(t, err)

			services := NewServiceInfo()
			services.AddService(&ServiceMonitoringInfo{
				Name:    "nginx",
				State:   "active",
				SubState: "running",
			})

			err = client.ServiceMonitoring.SubmitServiceData(context.Background(), tt.serverID, services)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestServiceMonitoringService_SubmitServiceMetrics tests submitting service metrics
func TestServiceMonitoringService_SubmitServiceMetrics(t *testing.T) {
	tests := []struct {
		name           string
		serverID       string
		mockStatus     int
		mockResponse   interface{}
		wantErr        bool
		errType        string
		clientAuthFunc func(*Config)
	}{
		{
			name:       "success - submit service metrics",
			serverID:   "server-001",
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": map[string]interface{}{"metricsCount": 5},
			},
			wantErr: false,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - invalid metrics format (400)",
			serverID:       "server-001",
			mockStatus:     http.StatusBadRequest,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			errType:        "ValidationError",
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - unauthorized (401)",
			serverID:       "server-001",
			mockStatus:     http.StatusUnauthorized,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			errType:        "UnauthorizedError",
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "invalid-key", APISecret: "invalid-secret"}
			},
		},
		{
			name:           "error - forbidden (403)",
			serverID:       "server-001",
			mockStatus:     http.StatusForbidden,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			errType:        "ForbiddenError",
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - server error (500)",
			serverID:       "server-001",
			mockStatus:     http.StatusInternalServerError,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			errType:        "APIError",
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - too many requests (429)",
			serverID:       "server-001",
			mockStatus:     http.StatusTooManyRequests,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			errType:        "RateLimitError",
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			config := &Config{
				BaseURL: server.URL,
			}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			require.NoError(t, err)

			now := time.Now()
			metrics := []*ServiceMetrics{
				{
					ServiceName:  "nginx",
					Timestamp:    now,
					CPUPercent:   25.5,
					MemoryRSS:    1024000,
					ProcessCount: 1,
					ThreadCount:  4,
				},
			}

			err = client.ServiceMonitoring.SubmitServiceMetrics(context.Background(), tt.serverID, metrics)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestServiceMonitoringService_GetServerServices tests getting server services
func TestServiceMonitoringService_GetServerServices(t *testing.T) {
	tests := []struct {
		name           string
		serverID       string
		mockStatus     int
		mockResponse   interface{}
		wantErr        bool
		wantServices   int
		checkFunc      func(*testing.T, *ServiceStatusResponse)
		clientAuthFunc func(*Config)
	}{
		{
			name:       "success - get service status",
			serverID:   "server-001",
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"server_uuid":  "server-001",
				"last_updated": time.Now().Format(time.RFC3339),
				"services": []map[string]interface{}{
					{
						"name":      "nginx",
						"state":     "active",
						"sub_state": "running",
						"main_pid":  1234,
					},
					{
						"name":      "mysql",
						"state":     "active",
						"sub_state": "running",
						"main_pid":  5678,
					},
				},
				"summary": map[string]interface{}{
					"total":    2,
					"active":   2,
					"inactive": 0,
					"failed":   0,
				},
			},
			wantErr:      false,
			wantServices: 2,
			checkFunc: func(t *testing.T, resp *ServiceStatusResponse) {
				assert.Equal(t, "server-001", resp.ServerUUID)
				assert.Equal(t, 2, resp.Summary.Total)
				assert.Equal(t, 2, resp.Summary.Active)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:       "success - different server id",
			serverID:   "server-789",
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"server_uuid":  "server-789",
				"last_updated": time.Now().Format(time.RFC3339),
				"services":     []interface{}{},
				"summary": map[string]interface{}{
					"total":    0,
					"active":   0,
					"inactive": 0,
					"failed":   0,
				},
			},
			wantErr:      false,
			wantServices: 0,
			checkFunc: func(t *testing.T, resp *ServiceStatusResponse) {
				assert.Equal(t, "server-789", resp.ServerUUID)
				assert.Equal(t, 0, resp.Summary.Total)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - not found (404)",
			serverID:       "nonexistent-server",
			mockStatus:     http.StatusNotFound,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			wantServices:   0,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - unauthorized (401)",
			serverID:       "server-001",
			mockStatus:     http.StatusUnauthorized,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			wantServices:   0,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "invalid-key", APISecret: "invalid-secret"}
			},
		},
		{
			name:           "error - forbidden (403)",
			serverID:       "server-001",
			mockStatus:     http.StatusForbidden,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			wantServices:   0,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - server error (500)",
			serverID:       "server-001",
			mockStatus:     http.StatusInternalServerError,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			wantServices:   0,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - service unavailable (503)",
			serverID:       "server-001",
			mockStatus:     http.StatusServiceUnavailable,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			wantServices:   0,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			config := &Config{
				BaseURL: server.URL,
			}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			require.NoError(t, err)

			resp, err := client.ServiceMonitoring.GetServerServices(context.Background(), tt.serverID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.checkFunc != nil {
					tt.checkFunc(t, resp)
				}
			}
		})
	}
}

// TestServiceMonitoringService_GetServiceHistory tests getting service history
func TestServiceMonitoringService_GetServiceHistory(t *testing.T) {
	tests := []struct {
		name           string
		serverID       string
		serviceName    string
		mockStatus     int
		mockResponse   interface{}
		wantErr        bool
		wantEntries    int
		checkFunc      func(*testing.T, *ServiceHistoryResponse)
		clientAuthFunc func(*Config)
	}{
		{
			name:        "success - get service history with pagination",
			serverID:    "server-001",
			serviceName: "nginx",
			mockStatus:  http.StatusOK,
			mockResponse: map[string]interface{}{
				"server_uuid":  "server-001",
				"service_name": "nginx",
				"history": []map[string]interface{}{
					{
						"timestamp":    "2024-10-17T10:00:00Z",
						"state":        "active",
						"sub_state":    "running",
						"cpu_percent":  25.5,
						"memory_bytes": 1024000,
					},
					{
						"timestamp":    "2024-10-17T11:00:00Z",
						"state":        "active",
						"sub_state":    "running",
						"cpu_percent":  30.2,
						"memory_bytes": 1124000,
					},
				},
			},
			wantErr:     false,
			wantEntries: 2,
			checkFunc: func(t *testing.T, resp *ServiceHistoryResponse) {
				assert.Equal(t, "server-001", resp.ServerUUID)
				assert.Equal(t, "nginx", resp.ServiceName)
				assert.Equal(t, 2, len(resp.History))
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:        "success - different time ranges",
			serverID:    "server-002",
			serviceName: "mysql",
			mockStatus:  http.StatusOK,
			mockResponse: map[string]interface{}{
				"server_uuid":  "server-002",
				"service_name": "mysql",
				"history": []map[string]interface{}{
					{
						"timestamp":    "2024-10-16T00:00:00Z",
						"state":        "active",
						"sub_state":    "running",
						"cpu_percent":  15.0,
						"memory_bytes": 2048000,
					},
				},
			},
			wantErr:     false,
			wantEntries: 1,
			checkFunc: func(t *testing.T, resp *ServiceHistoryResponse) {
				assert.Equal(t, "server-002", resp.ServerUUID)
				assert.Equal(t, "mysql", resp.ServiceName)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - not found (404)",
			serverID:       "nonexistent-server",
			serviceName:    "nginx",
			mockStatus:     http.StatusNotFound,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			wantEntries:    0,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - invalid parameters (400)",
			serverID:       "server-001",
			serviceName:    "nginx",
			mockStatus:     http.StatusBadRequest,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			wantEntries:    0,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - unauthorized (401)",
			serverID:       "server-001",
			serviceName:    "nginx",
			mockStatus:     http.StatusUnauthorized,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			wantEntries:    0,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "invalid-key", APISecret: "invalid-secret"}
			},
		},
		{
			name:           "error - forbidden (403)",
			serverID:       "server-001",
			serviceName:    "nginx",
			mockStatus:     http.StatusForbidden,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			wantEntries:    0,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - server error (500)",
			serverID:       "server-001",
			serviceName:    "nginx",
			mockStatus:     http.StatusInternalServerError,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			wantEntries:    0,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - service unavailable (503)",
			serverID:       "server-001",
			serviceName:    "nginx",
			mockStatus:     http.StatusServiceUnavailable,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			wantEntries:    0,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			config := &Config{
				BaseURL: server.URL,
			}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			require.NoError(t, err)

			opts := &ListOptions{Page: 1, Limit: 10}
			resp, err := client.ServiceMonitoring.GetServiceHistory(context.Background(), tt.serverID, tt.serviceName, opts)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.checkFunc != nil {
					tt.checkFunc(t, resp)
				}
			}
		})
	}
}

// TestServiceMonitoringService_RestartService tests restarting a service
func TestServiceMonitoringService_RestartService(t *testing.T) {
	tests := []struct {
		name           string
		serverID       string
		serviceName    string
		mockStatus     int
		mockResponse   interface{}
		wantErr        bool
		clientAuthFunc func(*Config)
	}{
		{
			name:       "success - restart service",
			serverID:   "server-001",
			serviceName: "nginx",
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": map[string]interface{}{"restart_initiated": true},
			},
			wantErr: false,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - unauthorized (401)",
			serverID:       "server-001",
			serviceName:    "nginx",
			mockStatus:     http.StatusUnauthorized,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "invalid-key", APISecret: "invalid-secret"}
			},
		},
		{
			name:           "error - forbidden (403)",
			serverID:       "server-001",
			serviceName:    "nginx",
			mockStatus:     http.StatusForbidden,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - not found (404)",
			serverID:       "nonexistent-server",
			serviceName:    "nginx",
			mockStatus:     http.StatusNotFound,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - server error (500)",
			serverID:       "server-001",
			serviceName:    "nginx",
			mockStatus:     http.StatusInternalServerError,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - conflict (409)",
			serverID:       "server-001",
			serviceName:    "nginx",
			mockStatus:     http.StatusConflict,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			config := &Config{
				BaseURL: server.URL,
			}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			require.NoError(t, err)

			err = client.ServiceMonitoring.RestartService(context.Background(), tt.serverID, tt.serviceName)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestServiceMonitoringService_GetServiceLogs tests retrieving service logs
func TestServiceMonitoringService_GetServiceLogs(t *testing.T) {
	tests := []struct {
		name           string
		serverID       string
		serviceName    string
		mockStatus     int
		mockResponse   interface{}
		wantErr        bool
		wantLogCount   int
		checkFunc      func(*testing.T, []ServiceLogEntry)
		clientAuthFunc func(*Config)
	}{
		{
			name:        "success - get logs with pagination",
			serverID:    "server-001",
			serviceName: "nginx",
			mockStatus:  http.StatusOK,
			mockResponse: []map[string]interface{}{
				{
					"timestamp": "2024-10-17T10:00:00Z",
					"level":     "info",
					"message":   "Server started",
				},
				{
					"timestamp": "2024-10-17T11:00:00Z",
					"level":     "warning",
					"message":   "Memory usage high",
				},
			},
			wantErr:     false,
			wantLogCount: 2,
			checkFunc: func(t *testing.T, logs []ServiceLogEntry) {
				assert.Equal(t, 2, len(logs))
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:        "success - logs with filters applied",
			serverID:    "server-002",
			serviceName: "mysql",
			mockStatus:  http.StatusOK,
			mockResponse: []map[string]interface{}{
				{
					"timestamp": "2024-10-17T10:30:00Z",
					"level":     "error",
					"message":   "Connection timeout",
				},
			},
			wantErr:      false,
			wantLogCount: 1,
			checkFunc: func(t *testing.T, logs []ServiceLogEntry) {
				assert.Equal(t, 1, len(logs))
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - not found (404)",
			serverID:       "nonexistent-server",
			serviceName:    "nginx",
			mockStatus:     http.StatusNotFound,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			wantLogCount:   0,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - unauthorized (401)",
			serverID:       "server-001",
			serviceName:    "nginx",
			mockStatus:     http.StatusUnauthorized,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			wantLogCount:   0,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "invalid-key", APISecret: "invalid-secret"}
			},
		},
		{
			name:           "error - forbidden (403)",
			serverID:       "server-001",
			serviceName:    "nginx",
			mockStatus:     http.StatusForbidden,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			wantLogCount:   0,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - server error (500)",
			serverID:       "server-001",
			serviceName:    "nginx",
			mockStatus:     http.StatusInternalServerError,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			wantLogCount:   0,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - bad request (400)",
			serverID:       "server-001",
			serviceName:    "nginx",
			mockStatus:     http.StatusBadRequest,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			wantLogCount:   0,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			config := &Config{
				BaseURL: server.URL,
			}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			require.NoError(t, err)

			opts := &ListOptions{Page: 1, Limit: 50}
			logs, err := client.ServiceMonitoring.GetServiceLogs(context.Background(), tt.serverID, tt.serviceName, opts)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, logs)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, logs)
				assert.Equal(t, tt.wantLogCount, len(logs))
				if tt.checkFunc != nil {
					tt.checkFunc(t, logs)
				}
			}
		})
	}
}

// TestServiceMonitoringService_GetFailedServices tests getting failed services
func TestServiceMonitoringService_GetFailedServices(t *testing.T) {
	tests := []struct {
		name            string
		organizationID  string
		mockStatus      int
		mockResponse    interface{}
		wantErr         bool
		wantServices    int
		checkFunc       func(*testing.T, []*ServiceMonitoringInfo)
		clientAuthFunc  func(*Config)
	}{
		{
			name:           "success - get failed services",
			organizationID: "org-001",
			mockStatus:     http.StatusOK,
			mockResponse: []map[string]interface{}{
				{
					"name":      "nginx",
					"state":     "failed",
					"sub_state": "dead",
					"main_pid":  0,
				},
				{
					"name":      "redis",
					"state":     "failed",
					"sub_state": "dead",
					"main_pid":  0,
				},
			},
			wantErr:      false,
			wantServices: 2,
			checkFunc: func(t *testing.T, services []*ServiceMonitoringInfo) {
				assert.Equal(t, 2, len(services))
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "success - empty list",
			organizationID: "org-002",
			mockStatus:     http.StatusOK,
			mockResponse:   []interface{}{},
			wantErr:        false,
			wantServices:   0,
			checkFunc: func(t *testing.T, services []*ServiceMonitoringInfo) {
				assert.Equal(t, 0, len(services))
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - not found (404)",
			organizationID: "nonexistent-org",
			mockStatus:     http.StatusNotFound,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			wantServices:   0,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - unauthorized (401)",
			organizationID: "org-001",
			mockStatus:     http.StatusUnauthorized,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			wantServices:   0,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "invalid-key", APISecret: "invalid-secret"}
			},
		},
		{
			name:           "error - forbidden (403)",
			organizationID: "org-001",
			mockStatus:     http.StatusForbidden,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			wantServices:   0,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - server error (500)",
			organizationID: "org-001",
			mockStatus:     http.StatusInternalServerError,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			wantServices:   0,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			config := &Config{
				BaseURL: server.URL,
			}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			require.NoError(t, err)

			services, err := client.ServiceMonitoring.GetFailedServices(context.Background(), tt.organizationID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, services)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, services)
				assert.Equal(t, tt.wantServices, len(services))
				if tt.checkFunc != nil {
					tt.checkFunc(t, services)
				}
			}
		})
	}
}

// TestServiceMonitoringService_CreateServiceAlert tests creating service alerts
func TestServiceMonitoringService_CreateServiceAlert(t *testing.T) {
	tests := []struct {
		name           string
		organizationID string
		mockStatus     int
		mockResponse   interface{}
		wantErr        bool
		clientAuthFunc func(*Config)
	}{
		{
			name:           "success - create alert",
			organizationID: "org-001",
			mockStatus:     http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": map[string]interface{}{"alert_id": "alert-123"},
			},
			wantErr: false,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - invalid config (400)",
			organizationID: "org-001",
			mockStatus:     http.StatusBadRequest,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - unauthorized (401)",
			organizationID: "org-001",
			mockStatus:     http.StatusUnauthorized,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "invalid-key", APISecret: "invalid-secret"}
			},
		},
		{
			name:           "error - forbidden (403)",
			organizationID: "org-001",
			mockStatus:     http.StatusForbidden,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - server error (500)",
			organizationID: "org-001",
			mockStatus:     http.StatusInternalServerError,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - conflict (409)",
			organizationID: "org-001",
			mockStatus:     http.StatusConflict,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			config := &Config{
				BaseURL: server.URL,
			}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			require.NoError(t, err)

			alertConfig := ServiceAlertConfig{
				Name:            "nginx-failure-alert",
				Description:     "Alert when nginx fails",
				ServicePatterns: []string{"nginx"},
				Conditions:      []string{"state=failed"},
				Severity:        "critical",
				Enabled:         true,
			}

			err = client.ServiceMonitoring.CreateServiceAlert(context.Background(), tt.organizationID, alertConfig)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestServiceMonitoringService_SubmitServiceLogs tests submitting service logs
func TestServiceMonitoringService_SubmitServiceLogs(t *testing.T) {
	tests := []struct {
		name           string
		serverID       string
		mockStatus     int
		mockResponse   interface{}
		wantErr        bool
		clientAuthFunc func(*Config)
	}{
		{
			name:       "success - submit logs",
			serverID:   "server-001",
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": map[string]interface{}{"logsSubmitted": true},
			},
			wantErr: false,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - invalid log format (400)",
			serverID:       "server-001",
			mockStatus:     http.StatusBadRequest,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - unauthorized (401)",
			serverID:       "server-001",
			mockStatus:     http.StatusUnauthorized,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "invalid-key", APISecret: "invalid-secret"}
			},
		},
		{
			name:           "error - forbidden (403)",
			serverID:       "server-001",
			mockStatus:     http.StatusForbidden,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - server error (500)",
			serverID:       "server-001",
			mockStatus:     http.StatusInternalServerError,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - payload too large (413)",
			serverID:       "server-001",
			mockStatus:     http.StatusRequestEntityTooLarge,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			config := &Config{
				BaseURL: server.URL,
			}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			require.NoError(t, err)

			now := time.Now()
			logs := map[string][]ServiceLogEntry{
				"nginx": {
					{
						Timestamp: now,
						Level:     "info",
						Message:   "Server started",
					},
					{
						Timestamp: now.Add(time.Minute),
						Level:     "warning",
						Message:   "Memory usage high",
					},
				},
			}

			err = client.ServiceMonitoring.SubmitServiceLogs(context.Background(), tt.serverID, logs)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
