
package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Phase 3A: Metrics Submission and Query Handler Tests

func TestMetricsService_Submit_Handler(t *testing.T) {
	tests := []struct {
		name           string
		serverUUID     string
		metrics        []*Metric
		mockStatus     int
		mockResponse   interface{}
		wantErr        bool
		clientAuthFunc func(*Config)
	}{
		{
			name:       "success - submit metrics",
			serverUUID: "server-uuid-001",
			metrics: []*Metric{
				{
					Name:  "cpu_usage",
					Value: 45.5,
					Unit:  "percent",
				},
				{
					Name:  "memory_usage",
					Value: 62.3,
					Unit:  "percent",
				},
			},
			mockStatus:   http.StatusOK,
			mockResponse: map[string]interface{}{},
			wantErr:      false,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{ServerUUID: "server-uuid-001", ServerSecret: "secret"}
			},
		},
		{
			name:       "success - submit single metric",
			serverUUID: "server-uuid-002",
			metrics: []*Metric{
				{
					Name:  "disk_usage",
					Value: 75.0,
					Unit:  "percent",
				},
			},
			mockStatus:   http.StatusOK,
			mockResponse: map[string]interface{}{},
			wantErr:      false,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{ServerUUID: "server-uuid-002", ServerSecret: "secret"}
			},
		},
		{
			name:         "error - bad request (400)",
			serverUUID:   "server-uuid-001",
			metrics:      []*Metric{},
			mockStatus:   http.StatusBadRequest,
			mockResponse: map[string]interface{}{"error": "metrics required"},
			wantErr:      true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{ServerUUID: "server-uuid-001", ServerSecret: "secret"}
			},
		},
		{
			name:       "error - unauthorized (401)",
			serverUUID: "server-uuid-001",
			metrics: []*Metric{
				{Name: "cpu_usage", Value: 50.0},
			},
			mockStatus:   http.StatusUnauthorized,
			mockResponse: map[string]interface{}{"error": "invalid credentials"},
			wantErr:      true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{ServerUUID: "invalid", ServerSecret: "invalid"}
			},
		},
		{
			name:       "error - not found (404)",
			serverUUID: "invalid-uuid",
			metrics: []*Metric{
				{Name: "cpu_usage", Value: 50.0},
			},
			mockStatus:   http.StatusNotFound,
			mockResponse: map[string]interface{}{"error": "server not found"},
			wantErr:      true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{ServerUUID: "invalid-uuid", ServerSecret: "secret"}
			},
		},
		{
			name:       "error - conflict (409)",
			serverUUID: "server-uuid-001",
			metrics: []*Metric{
				{Name: "invalid_metric", Value: 999.0},
			},
			mockStatus:   http.StatusConflict,
			mockResponse: map[string]interface{}{"error": "metric validation failed"},
			wantErr:      true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{ServerUUID: "server-uuid-001", ServerSecret: "secret"}
			},
		},
		{
			name:       "error - server error (500)",
			serverUUID: "server-uuid-001",
			metrics: []*Metric{
				{Name: "cpu_usage", Value: 50.0},
			},
			mockStatus:   http.StatusInternalServerError,
			mockResponse: map[string]interface{}{"error": "internal server error"},
			wantErr:      true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{ServerUUID: "server-uuid-001", ServerSecret: "secret"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/metrics")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			config := &Config{BaseURL: server.URL}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			assert.NoError(t, err)

			err = client.Metrics.Submit(context.Background(), tt.serverUUID, tt.metrics)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMetricsService_SubmitComprehensive_Handler(t *testing.T) {
	tests := []struct {
		name           string
		input          *ComprehensiveMetricsRequest
		mockStatus     int
		mockResponse   interface{}
		wantErr        bool
		clientAuthFunc func(*Config)
	}{
		{
			name: "success - submit comprehensive metrics",
			input: &ComprehensiveMetricsRequest{
				ServerUUID:  "server-uuid-001",
				CollectedAt: "2023-10-01T12:00:00Z",
				CPU: &CPUMetrics{
					UsagePercent: 45.5,
					CoreCount:    4,
					LoadAverage1: 1.2,
				},
				Memory: &MemoryMetrics{
					UsedBytes:    8192,
					TotalBytes:   16384,
					UsagePercent: 50.0,
				},
			},
			mockStatus:   http.StatusOK,
			mockResponse: map[string]interface{}{},
			wantErr:      false,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{ServerUUID: "server-uuid-001", ServerSecret: "secret"}
			},
		},
		{
			name: "success - submit with disk metrics",
			input: &ComprehensiveMetricsRequest{
				ServerUUID:  "server-uuid-002",
				CollectedAt: "2023-10-01T12:00:00Z",
				Disks: []DiskMetrics{
					{
						Mountpoint:   "/",
						UsedBytes:    100,
						TotalBytes:   500,
						UsagePercent: 20.0,
					},
				},
			},
			mockStatus:   http.StatusOK,
			mockResponse: map[string]interface{}{},
			wantErr:      false,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{ServerUUID: "server-uuid-002", ServerSecret: "secret"}
			},
		},
		{
			name:         "error - bad request (400)",
			input:        &ComprehensiveMetricsRequest{},
			mockStatus:   http.StatusBadRequest,
			mockResponse: map[string]interface{}{"error": "invalid metrics"},
			wantErr:      true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{ServerUUID: "server-uuid-001", ServerSecret: "secret"}
			},
		},
		{
			name: "error - unauthorized (401)",
			input: &ComprehensiveMetricsRequest{
				ServerUUID: "server-uuid-001",
				CollectedAt: "2023-10-01T12:00:00Z",
			},
			mockStatus:   http.StatusUnauthorized,
			mockResponse: map[string]interface{}{"error": "invalid credentials"},
			wantErr:      true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{ServerUUID: "invalid", ServerSecret: "invalid"}
			},
		},
		{
			name: "error - server error (500)",
			input: &ComprehensiveMetricsRequest{
				ServerUUID: "server-uuid-001",
				CollectedAt: "2023-10-01T12:00:00Z",
			},
			mockStatus:   http.StatusInternalServerError,
			mockResponse: map[string]interface{}{"error": "internal server error"},
			wantErr:      true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{ServerUUID: "server-uuid-001", ServerSecret: "secret"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Contains(t, r.URL.Path, "/v2/metrics/comprehensive")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			config := &Config{BaseURL: server.URL}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			assert.NoError(t, err)

			err = client.Metrics.SubmitComprehensive(context.Background(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMetricsService_SubmitAggregatedMetrics_Handler(t *testing.T) {
	tests := []struct {
		name           string
		input          *AggregatedMetricsRequest
		mockStatus     int
		mockResponse   interface{}
		wantErr        bool
		clientAuthFunc func(*Config)
	}{
		{
			name: "success - submit aggregated metrics",
			input: &AggregatedMetricsRequest{
				ServerUUID:  "server-uuid-001",
				CollectedAt: "2023-10-01T12:00:00Z",
				CPU: &CPUAggregation{
					UsagePercent:   45.5,
					MaxCorePercent: 95.0,
					MinCorePercent: 10.0,
				},
				Memory: &MemoryAggregation{
					UsedBytes:   5242880,
					TotalBytes:  8388608,
					UsedPercent: 62.3,
				},
			},
			mockStatus:   http.StatusOK,
			mockResponse: map[string]interface{}{},
			wantErr:      false,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{ServerUUID: "server-uuid-001", ServerSecret: "secret"}
			},
		},
		{
			name: "success - submit with CPU aggregation",
			input: &AggregatedMetricsRequest{
				ServerUUID:  "server-uuid-002",
				CollectedAt: "2023-10-01T12:00:00Z",
				CPU: &CPUAggregation{
					UsagePercent: 50.0,
				},
			},
			mockStatus:   http.StatusOK,
			mockResponse: map[string]interface{}{},
			wantErr:      false,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{ServerUUID: "server-uuid-001", ServerSecret: "secret"}
			},
		},
		{
			name:         "error - bad request (400)",
			input:        &AggregatedMetricsRequest{},
			mockStatus:   http.StatusBadRequest,
			mockResponse: map[string]interface{}{"error": "invalid aggregated metrics"},
			wantErr:      true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{ServerUUID: "server-uuid-001", ServerSecret: "secret"}
			},
		},
		{
			name: "error - unauthorized (401)",
			input: &AggregatedMetricsRequest{
				ServerUUID: "server-uuid-001",
				CollectedAt: "2023-10-01T12:00:00Z",
			},
			mockStatus:   http.StatusUnauthorized,
			mockResponse: map[string]interface{}{"error": "invalid credentials"},
			wantErr:      true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{ServerUUID: "invalid", ServerSecret: "invalid"}
			},
		},
		{
			name: "error - conflict (409)",
			input: &AggregatedMetricsRequest{
				ServerUUID: "server-uuid-001",
				CollectedAt: "2023-10-01T12:00:00Z",
			},
			mockStatus:   http.StatusConflict,
			mockResponse: map[string]interface{}{"error": "duplicate metrics"},
			wantErr:      true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{ServerUUID: "server-uuid-001", ServerSecret: "secret"}
			},
		},
		{
			name: "error - server error (500)",
			input: &AggregatedMetricsRequest{
				ServerUUID: "server-uuid-001",
				CollectedAt: "2023-10-01T12:00:00Z",
			},
			mockStatus:   http.StatusInternalServerError,
			mockResponse: map[string]interface{}{"error": "internal server error"},
			wantErr:      true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{ServerUUID: "server-uuid-001", ServerSecret: "secret"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/metrics/aggregated")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			config := &Config{BaseURL: server.URL}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			assert.NoError(t, err)

			err = client.Metrics.SubmitAggregatedMetrics(context.Background(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMetricsService_Query_Handler(t *testing.T) {
	tests := []struct {
		name           string
		query          *MetricsQuery
		mockStatus     int
		mockResponse   interface{}
		wantErr        bool
		wantLen        int
		checkFunc      func(*testing.T, []*Metric)
		clientAuthFunc func(*Config)
	}{
		{
			name: "success - query metrics",
			query: &MetricsQuery{
				ServerUUIDs: []string{"server-uuid-001"},
				MetricNames: []string{"cpu_usage", "memory_usage"},
				StartTime:   "2023-10-01T00:00:00Z",
				EndTime: "2023-10-02T00:00:00Z",
			},
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"server_uuid": "server-uuid-001",
						"name":        "cpu_usage",
						"value":       45.5,
						"unit":        "percent",
						"timestamp":   "2023-10-01T12:00:00Z",
					},
					{
						"server_uuid": "server-uuid-001",
						"name":        "memory_usage",
						"value":       62.3,
						"unit":        "percent",
						"timestamp":   "2023-10-01T12:00:00Z",
					},
				},
			},
			wantErr: false,
			wantLen: 2,
			checkFunc: func(t *testing.T, metrics []*Metric) {
				if len(metrics) > 0 && metrics[0] != nil {
					assert.Equal(t, "cpu_usage", metrics[0].Name)
				}
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name: "success - query with aggregation",
			query: &MetricsQuery{
				ServerUUIDs: []string{"server-uuid-001"},
				MetricNames: []string{"cpu_usage"},
				StartTime:   "2023-10-01T00:00:00Z",
				EndTime: "2023-10-02T00:00:00Z",
				Aggregation: "avg",
				GroupBy:     "1h",
			},
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"name":  "cpu_usage_avg",
						"value": 50.0,
					},
				},
			},
			wantErr: false,
			wantLen: 1,
			checkFunc: func(t *testing.T, metrics []*Metric) {
				assert.Equal(t, 1, len(metrics))
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - bad request (400)",
			query:        &MetricsQuery{},
			mockStatus:   http.StatusBadRequest,
			mockResponse: map[string]interface{}{"error": "invalid query parameters"},
			wantErr:      true,
			wantLen:      0,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name: "error - unauthorized (401)",
			query: &MetricsQuery{
				ServerUUIDs: []string{"server-uuid-001"},
				StartTime:   "2023-10-01T00:00:00Z",
				EndTime: "2023-10-02T00:00:00Z",
			},
			mockStatus:   http.StatusUnauthorized,
			mockResponse: map[string]interface{}{"error": "invalid token"},
			wantErr:      true,
			wantLen:      0,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "invalid-token"}
			},
		},
		{
			name: "error - not found (404)",
			query: &MetricsQuery{
				ServerUUIDs: []string{"invalid-uuid"},
				StartTime:   "2023-10-01T00:00:00Z",
				EndTime: "2023-10-02T00:00:00Z",
			},
			mockStatus:   http.StatusNotFound,
			mockResponse: map[string]interface{}{"error": "server not found"},
			wantErr:      true,
			wantLen:      0,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name: "error - server error (500)",
			query: &MetricsQuery{
				ServerUUIDs: []string{"server-uuid-001"},
				StartTime:   "2023-10-01T00:00:00Z",
				EndTime: "2023-10-02T00:00:00Z",
			},
			mockStatus:   http.StatusInternalServerError,
			mockResponse: map[string]interface{}{"error": "internal server error"},
			wantErr:      true,
			wantLen:      0,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/metrics/query")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			config := &Config{BaseURL: server.URL}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			assert.NoError(t, err)

			metrics, err := client.Metrics.Query(context.Background(), tt.query)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, metrics)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantLen, len(metrics))
				if tt.checkFunc != nil {
					tt.checkFunc(t, metrics)
				}
			}
		})
	}
}

// Phase 3B: Metrics Retrieval Handler Tests

func TestMetricsService_Get_Handler(t *testing.T) {
	tests := []struct {
		name           string
		serverUUID     string
		opts           *ListOptions
		mockStatus     int
		mockResponse   interface{}
		wantErr        bool
		wantLen        int
		checkFunc      func(*testing.T, []*Metric, *PaginationMeta)
		clientAuthFunc func(*Config)
	}{
		{
			name:       "success - get server metrics",
			serverUUID: "server-uuid-001",
			opts:       nil,
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"server_uuid": "server-uuid-001",
						"name":        "cpu_usage",
						"value":       45.5,
						"unit":        "percent",
						"timestamp":   "2023-10-01T12:00:00Z",
					},
					{
						"server_uuid": "server-uuid-001",
						"name":        "memory_usage",
						"value":       62.3,
						"unit":        "percent",
						"timestamp":   "2023-10-01T12:00:00Z",
					},
				},
				"meta": map[string]interface{}{
					"total_items": 2,
					"limit":       25,
					"page":        1,
				},
			},
			wantErr: false,
			wantLen: 2,
			checkFunc: func(t *testing.T, metrics []*Metric, meta *PaginationMeta) {
				assert.Equal(t, 2, meta.TotalItems)
				if len(metrics) > 0 && metrics[0] != nil {
					assert.Equal(t, "cpu_usage", metrics[0].Name)
				}
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:       "success - get with pagination",
			serverUUID: "server-uuid-001",
			opts:       &ListOptions{Page: 2, Limit: 10},
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"server_uuid": "server-uuid-001",
						"name":        "disk_usage",
						"value":       75.0,
						"unit":        "percent",
					},
				},
				"meta": map[string]interface{}{
					"total_items": 50,
					"limit":       10,
					"page":        2,
					"per_page":    10,
				},
			},
			wantErr: false,
			wantLen: 1,
			checkFunc: func(t *testing.T, metrics []*Metric, meta *PaginationMeta) {
				assert.Equal(t, 50, meta.TotalItems)
				assert.Equal(t, 2, meta.Page)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:       "success - empty metrics",
			serverUUID: "server-uuid-002",
			opts:       nil,
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": []map[string]interface{}{},
				"meta": map[string]interface{}{
					"total_items": 0,
					"limit":       25,
					"page":        1,
				},
			},
			wantErr: false,
			wantLen: 0,
			checkFunc: func(t *testing.T, metrics []*Metric, meta *PaginationMeta) {
				assert.Equal(t, 0, meta.TotalItems)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - not found (404)",
			serverUUID:   "invalid-uuid",
			opts:         nil,
			mockStatus:   http.StatusNotFound,
			mockResponse: map[string]interface{}{"error": "server not found"},
			wantErr:      true,
			wantLen:      0,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - unauthorized (401)",
			serverUUID:   "server-uuid-001",
			opts:         nil,
			mockStatus:   http.StatusUnauthorized,
			mockResponse: map[string]interface{}{"error": "invalid token"},
			wantErr:      true,
			wantLen:      0,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "invalid-token"}
			},
		},
		{
			name:         "error - server error (500)",
			serverUUID:   "server-uuid-001",
			opts:         nil,
			mockStatus:   http.StatusInternalServerError,
			mockResponse: map[string]interface{}{"error": "internal server error"},
			wantErr:      true,
			wantLen:      0,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/metrics/server/"+tt.serverUUID)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			config := &Config{BaseURL: server.URL}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			assert.NoError(t, err)

			metrics, meta, err := client.Metrics.Get(context.Background(), tt.serverUUID, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, metrics)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantLen, len(metrics))
				if tt.checkFunc != nil {
					tt.checkFunc(t, metrics, meta)
				}
			}
		})
	}
}

func TestMetricsService_GetSummary_Handler(t *testing.T) {
	tests := []struct {
		name           string
		serverUUID     string
		timeRange      *QueryTimeRange
		mockStatus     int
		mockResponse   interface{}
		wantErr        bool
		checkFunc      func(*testing.T, map[string]interface{})
		clientAuthFunc func(*Config)
	}{
		{
			name:       "success - get metrics summary",
			serverUUID: "server-uuid-001",
			timeRange: &QueryTimeRange{
				Start: time.Unix(1697472000, 0),
				End:   time.Unix(1697558400, 0),
			},
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"cpu_usage_avg":          45.5,
					"cpu_usage_max":          95.0,
					"cpu_usage_min":          10.0,
					"memory_usage_avg":       62.3,
					"memory_usage_max":       90.0,
					"disk_usage_current":     75.0,
					"uptime_percent":         99.9,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, summary map[string]interface{}) {
				assert.NotNil(t, summary)
				assert.Equal(t, 45.5, summary["cpu_usage_avg"])
				assert.Equal(t, 99.9, summary["uptime_percent"])
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:       "success - without time range",
			serverUUID: "server-uuid-002",
			timeRange:  nil,
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"cpu_usage_avg": 55.0,
					"memory_usage_avg": 70.0,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, summary map[string]interface{}) {
				assert.NotNil(t, summary)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:       "error - not found (404)",
			serverUUID: "invalid-uuid",
			timeRange:  nil,
			mockStatus: http.StatusNotFound,
			mockResponse: map[string]interface{}{"error": "server not found"},
			wantErr:    true,
			checkFunc:  nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:       "error - unauthorized (401)",
			serverUUID: "server-uuid-001",
			timeRange:  nil,
			mockStatus: http.StatusUnauthorized,
			mockResponse: map[string]interface{}{"error": "invalid token"},
			wantErr:    true,
			checkFunc:  nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "invalid-token"}
			},
		},
		{
			name:       "error - server error (500)",
			serverUUID: "server-uuid-001",
			timeRange:  nil,
			mockStatus: http.StatusInternalServerError,
			mockResponse: map[string]interface{}{"error": "internal server error"},
			wantErr:    true,
			checkFunc:  nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/metrics/server/"+tt.serverUUID+"/summary")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			config := &Config{BaseURL: server.URL}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			assert.NoError(t, err)

			summary, err := client.Metrics.GetSummary(context.Background(), tt.serverUUID, tt.timeRange)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, summary)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, summary)
				if tt.checkFunc != nil {
					tt.checkFunc(t, summary)
				}
			}
		})
	}
}

func TestMetricsService_GetAggregated_Handler(t *testing.T) {
	tests := []struct {
		name           string
		aggregation    *MetricsAggregation
		mockStatus     int
		mockResponse   interface{}
		wantErr        bool
		checkFunc      func(*testing.T, map[string]interface{})
		clientAuthFunc func(*Config)
	}{
		{
			name: "success - aggregate metrics",
			aggregation: &MetricsAggregation{
				ServerUUIDs: []string{"server-uuid-001", "server-uuid-002"},
				MetricNames: []string{"cpu_usage", "memory_usage"},
				StartTime:   "2023-10-01T00:00:00Z",
				EndTime: "2023-10-02T00:00:00Z",
				GroupBy:     []string{"server_uuid"},
				Function:    "avg",
				Interval:    "1h",
			},
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"results": []map[string]interface{}{
						{
							"server_uuid":  "server-uuid-001",
							"cpu_usage":    45.5,
							"memory_usage": 62.3,
						},
						{
							"server_uuid":  "server-uuid-002",
							"cpu_usage":    55.0,
							"memory_usage": 70.0,
						},
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, result map[string]interface{}) {
				assert.NotNil(t, result)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name: "success - sum aggregation",
			aggregation: &MetricsAggregation{
				MetricNames: []string{"network_bytes_sent"},
				StartTime:   "2023-10-01T00:00:00Z",
				EndTime: "2023-10-02T00:00:00Z",
				GroupBy:     []string{"server_uuid"},
				Function:    "sum",
			},
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"total_bytes": 1000000,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, result map[string]interface{}) {
				assert.NotNil(t, result)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - bad request (400)",
			aggregation:  &MetricsAggregation{},
			mockStatus:   http.StatusBadRequest,
			mockResponse: map[string]interface{}{"error": "invalid aggregation parameters"},
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name: "error - unauthorized (401)",
			aggregation: &MetricsAggregation{
				MetricNames: []string{"cpu_usage"},
				StartTime:   "2023-10-01T00:00:00Z",
				EndTime: "2023-10-02T00:00:00Z",
				Function:    "avg",
			},
			mockStatus:   http.StatusUnauthorized,
			mockResponse: map[string]interface{}{"error": "invalid token"},
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "invalid-token"}
			},
		},
		{
			name: "error - server error (500)",
			aggregation: &MetricsAggregation{
				MetricNames: []string{"cpu_usage"},
				StartTime:   "2023-10-01T00:00:00Z",
				EndTime: "2023-10-02T00:00:00Z",
				Function:    "avg",
			},
			mockStatus:   http.StatusInternalServerError,
			mockResponse: map[string]interface{}{"error": "internal server error"},
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/metrics/aggregate")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			config := &Config{BaseURL: server.URL}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			assert.NoError(t, err)

			result, err := client.Metrics.GetAggregated(context.Background(), tt.aggregation)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// Phase 3C: Metrics Status and Export Handler Tests

func TestMetricsService_GetStatus_Handler(t *testing.T) {
	tests := []struct {
		name           string
		serverUUID     string
		mockStatus     int
		mockResponse   interface{}
		wantErr        bool
		checkFunc      func(*testing.T, *MetricsStatus)
		clientAuthFunc func(*Config)
	}{
		{
			name:       "success - get metrics status",
			serverUUID: "server-uuid-001",
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"server_uuid":          "server-uuid-001",
					"collection_enabled":   true,
					"last_collection":      "2023-10-01T12:00:00Z",
					"next_collection":      "2023-10-01T12:05:00Z",
					"collection_interval":  300,
					"error_count":          0,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, status *MetricsStatus) {
				assert.Equal(t, "server-uuid-001", status.ServerUUID)
				assert.Equal(t, true, status.CollectionEnabled)
				assert.Equal(t, 300, status.CollectionInterval)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:       "success - get disabled status",
			serverUUID: "server-uuid-002",
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"server_uuid":         "server-uuid-002",
					"collection_enabled":  false,
					"collection_interval": 0,
					"error_count":         5,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, status *MetricsStatus) {
				assert.Equal(t, "server-uuid-002", status.ServerUUID)
				assert.Equal(t, false, status.CollectionEnabled)
				assert.Equal(t, 5, status.ErrorCount)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - not found (404)",
			serverUUID:   "invalid-uuid",
			mockStatus:   http.StatusNotFound,
			mockResponse: map[string]interface{}{"error": "server not found"},
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - unauthorized (401)",
			serverUUID:   "server-uuid-001",
			mockStatus:   http.StatusUnauthorized,
			mockResponse: map[string]interface{}{"error": "invalid token"},
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "invalid-token"}
			},
		},
		{
			name:         "error - server error (500)",
			serverUUID:   "server-uuid-001",
			mockStatus:   http.StatusInternalServerError,
			mockResponse: map[string]interface{}{"error": "internal server error"},
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/metrics/"+tt.serverUUID+"/status")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			config := &Config{BaseURL: server.URL}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			assert.NoError(t, err)

			status, err := client.Metrics.GetStatus(context.Background(), tt.serverUUID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, status)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, status)
				if tt.checkFunc != nil {
					tt.checkFunc(t, status)
				}
			}
		})
	}
}

func TestMetricsService_Export_Handler(t *testing.T) {
	tests := []struct {
		name           string
		export         *MetricsExport
		mockStatus     int
		mockResponse   []byte
		wantErr        bool
		checkFunc      func(*testing.T, []byte)
		clientAuthFunc func(*Config)
	}{
		{
			name: "success - export as CSV",
			export: &MetricsExport{
				ServerUUIDs: []string{"server-uuid-001"},
				MetricNames: []string{"cpu_usage", "memory_usage"},
				StartTime:   "2023-10-01T00:00:00Z",
				EndTime: "2023-10-02T00:00:00Z",
				Format:      "csv",
			},
			mockStatus:   http.StatusOK,
			mockResponse: []byte("server_uuid,metric_name,value,timestamp\nserver-uuid-001,cpu_usage,45.5,2023-10-01T12:00:00Z"),
			wantErr:      false,
			checkFunc: func(t *testing.T, data []byte) {
				assert.NotNil(t, data)
				assert.Contains(t, string(data), "server_uuid")
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name: "success - export as JSON",
			export: &MetricsExport{
				ServerUUIDs: []string{"server-uuid-001"},
				StartTime:   "2023-10-01T00:00:00Z",
				EndTime: "2023-10-02T00:00:00Z",
				Format:      "json",
			},
			mockStatus: http.StatusOK,
			mockResponse: []byte(`{
				"metrics": [
					{"server_uuid":"server-uuid-001","name":"cpu_usage","value":45.5}
				]
			}`),
			wantErr: false,
			checkFunc: func(t *testing.T, data []byte) {
				assert.NotNil(t, data)
				assert.Contains(t, string(data), "metrics")
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - bad request (400)",
			export:       &MetricsExport{},
			mockStatus:   http.StatusBadRequest,
			mockResponse: []byte(`{"error":"invalid export parameters"}`),
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name: "error - unauthorized (401)",
			export: &MetricsExport{
				ServerUUIDs: []string{"server-uuid-001"},
				StartTime:   "2023-10-01T00:00:00Z",
				EndTime: "2023-10-02T00:00:00Z",
				Format:      "csv",
			},
			mockStatus:   http.StatusUnauthorized,
			mockResponse: []byte(`{"error":"invalid token"}`),
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "invalid-token"}
			},
		},
		{
			name: "error - forbidden (403)",
			export: &MetricsExport{
				ServerUUIDs: []string{"server-uuid-001"},
				StartTime:   "2023-10-01T00:00:00Z",
				EndTime: "2023-10-02T00:00:00Z",
				Format:      "csv",
			},
			mockStatus:   http.StatusForbidden,
			mockResponse: []byte(`{"error":"insufficient permissions"}`),
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name: "error - server error (500)",
			export: &MetricsExport{
				ServerUUIDs: []string{"server-uuid-001"},
				StartTime:   "2023-10-01T00:00:00Z",
				EndTime: "2023-10-02T00:00:00Z",
				Format:      "csv",
			},
			mockStatus:   http.StatusInternalServerError,
			mockResponse: []byte(`{"error":"internal server error"}`),
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/metrics/export")

				w.WriteHeader(tt.mockStatus)
				w.Write(tt.mockResponse)
			}))
			defer server.Close()

			config := &Config{BaseURL: server.URL}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			assert.NoError(t, err)

			data, err := client.Metrics.Export(context.Background(), tt.export)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, data)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, data)
				if tt.checkFunc != nil {
					tt.checkFunc(t, data)
				}
			}
		})
	}
}
