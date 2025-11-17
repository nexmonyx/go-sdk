package nexmonyx

import (
	"context"
	"encoding/json"
	"net"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMetricsService_Submit tests the Submit method with various scenarios
func TestMetricsService_Submit(t *testing.T) {
	tests := []struct {
		name           string
		serverUUID     string
		metrics        []*Metric
		authConfig     AuthConfig
		serverStatus   int
		serverResponse map[string]interface{}
		expectError    bool
		description    string
	}{
		{
			name:       "successful_submission_with_jwt",
			serverUUID: "server-uuid-123",
			metrics: []*Metric{
				{
					ServerUUID: "server-uuid-123",
					Name:       "cpu.usage",
					Value:      45.5,
					Unit:       "percent",
					Timestamp:  time.Now(),
				},
				{
					ServerUUID: "server-uuid-123",
					Name:       "memory.usage",
					Value:      60.2,
					Unit:       "percent",
					Timestamp:  time.Now(),
				},
			},
			authConfig: AuthConfig{
				Token: "test-jwt-token",
			},
			serverStatus: http.StatusOK,
			serverResponse: map[string]interface{}{
				"status":  "success",
				"message": "Metrics submitted successfully",
			},
			expectError: false,
			description: "Should successfully submit metrics with JWT authentication",
		},
		{
			name:       "successful_submission_with_server_auth",
			serverUUID: "server-uuid-456",
			metrics: []*Metric{
				{
					ServerUUID: "server-uuid-456",
					Name:       "disk.usage",
					Value:      75.0,
					Unit:       "percent",
					Timestamp:  time.Now(),
				},
			},
			authConfig: AuthConfig{
				ServerUUID:   "server-uuid-456",
				ServerSecret: "server-secret",
			},
			serverStatus: http.StatusAccepted,
			serverResponse: map[string]interface{}{
				"status":  "success",
				"message": "Metrics queued for processing",
			},
			expectError: false,
			description: "Should successfully submit metrics with server authentication",
		},
		{
			name:       "error_bad_request",
			serverUUID: "server-uuid-789",
			metrics: []*Metric{
				{
					ServerUUID: "server-uuid-789",
					Name:       "invalid.metric",
					Value:      -1.0,
					Timestamp:  time.Now(),
				},
			},
			authConfig: AuthConfig{
				Token: "test-jwt-token",
			},
			serverStatus: http.StatusBadRequest,
			serverResponse: map[string]interface{}{
				"status":  "error",
				"message": "Invalid metric data",
			},
			expectError: true,
			description: "Should handle bad request errors",
		},
		{
			name:       "error_unauthorized",
			serverUUID: "server-uuid-999",
			metrics: []*Metric{
				{
					ServerUUID: "server-uuid-999",
					Name:       "cpu.usage",
					Value:      50.0,
					Timestamp:  time.Now(),
				},
			},
			authConfig: AuthConfig{
				Token: "invalid-token",
			},
			serverStatus: http.StatusUnauthorized,
			serverResponse: map[string]interface{}{
				"status":  "error",
				"message": "Unauthorized",
			},
			expectError: true,
			description: "Should handle unauthorized errors",
		},
		{
			name:       "empty_metrics_array",
			serverUUID: "server-uuid-empty",
			metrics:    []*Metric{},
			authConfig: AuthConfig{
				Token: "test-jwt-token",
			},
			serverStatus: http.StatusOK,
			serverResponse: map[string]interface{}{
				"status":  "success",
				"message": "No metrics to process",
			},
			expectError: false,
			description: "Should handle empty metrics array",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request
				assert.Equal(t, "/v1/metrics", r.URL.Path)
				assert.Equal(t, "POST", r.Method)

				// Decode body
				var body map[string]interface{}
				err := json.NewDecoder(r.Body).Decode(&body)
				require.NoError(t, err)

				// Verify server UUID in body
				assert.Equal(t, tt.serverUUID, body["server_uuid"])

				// Verify metrics array
				metricsArray, ok := body["metrics"].([]interface{})
				assert.True(t, ok)
				assert.Len(t, metricsArray, len(tt.metrics))

				// Send response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				json.NewEncoder(w).Encode(tt.serverResponse)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    tt.authConfig,
			})
			require.NoError(t, err)

			err = client.Metrics.Submit(context.Background(), tt.serverUUID, tt.metrics)

			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

// TestMetricsService_SubmitComprehensive tests the SubmitComprehensive method
func TestMetricsService_SubmitComprehensive(t *testing.T) {
	tests := []struct {
		name           string
		metrics        *ComprehensiveMetricsRequest
		authConfig     AuthConfig
		serverStatus   int
		expectError    bool
		expectedUUID   string
		description    string
	}{
		{
			name: "comprehensive_metrics_with_all_data",
			metrics: &ComprehensiveMetricsRequest{
				ServerUUID:  "server-uuid-comprehensive",
				CollectedAt: time.Now().Format(time.RFC3339),
				SystemInfo: &SystemInfo{
					Hostname:        "test-server",
					OS:              "linux",
					OSVersion:       "Ubuntu 22.04",
					KernelVersion:   "5.15.0",
					CPUArchitecture: "x86_64",
					Uptime:          3600,
				},
				CPU: &CPUMetrics{
					UsagePercent:  45.5,
					CoreCount:     8,
					LoadAverage1:  1.2,
					LoadAverage5:  1.5,
					LoadAverage15: 1.8,
					PerCoreUsage:  []float64{40.0, 45.0, 50.0, 48.0, 44.0, 46.0, 49.0, 47.0},
				},
				Memory: &MemoryMetrics{
					TotalBytes:       16777216000,
					UsedBytes:        8388608000,
					FreeBytes:        8388608000,
					AvailableBytes:   8388608000,
					UsagePercent:     50.0,
					SwapTotalBytes:   4194304000,
					SwapUsedBytes:    1048576000,
					SwapUsagePercent: 25.0,
				},
			},
			authConfig: AuthConfig{
				Token: "test-jwt-token",
			},
			serverStatus:   http.StatusOK,
			expectError:    false,
			expectedUUID:   "server-uuid-comprehensive",
			description:    "Should submit comprehensive metrics with all data",
		},
		{
			name: "auto_populate_server_uuid",
			metrics: &ComprehensiveMetricsRequest{
				ServerUUID:  "", // Empty, should be auto-populated
				CollectedAt: time.Now().Format(time.RFC3339),
				CPU: &CPUMetrics{
					UsagePercent: 50.0,
				},
			},
			authConfig: AuthConfig{
				ServerUUID:   "auto-populated-uuid",
				ServerSecret: "server-secret",
			},
			serverStatus:   http.StatusAccepted,
			expectError:    false,
			expectedUUID:   "auto-populated-uuid",
			description:    "Should auto-populate server UUID from client config",
		},
		{
			name: "error_server_not_found",
			metrics: &ComprehensiveMetricsRequest{
				ServerUUID:  "non-existent-uuid",
				CollectedAt: time.Now().Format(time.RFC3339),
			},
			authConfig: AuthConfig{
				Token: "test-jwt-token",
			},
			serverStatus:   http.StatusNotFound,
			expectError:    true,
			expectedUUID:   "non-existent-uuid",
			description:    "Should handle server not found error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/v2/metrics/comprehensive", r.URL.Path)
				assert.Equal(t, "POST", r.Method)

				var body ComprehensiveMetricsRequest
				err := json.NewDecoder(r.Body).Decode(&body)
				require.NoError(t, err)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"status":  "success",
					"message": "Metrics processed",
				})
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    tt.authConfig,
			})
			require.NoError(t, err)

			err = client.Metrics.SubmitComprehensive(context.Background(), tt.metrics)

			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				assert.Equal(t, tt.expectedUUID, tt.metrics.ServerUUID, "ServerUUID should be set correctly")
			}
		})
	}
}

// TestMetricsService_SubmitAggregatedMetrics tests the SubmitAggregatedMetrics method
func TestMetricsService_SubmitAggregatedMetrics(t *testing.T) {
	tests := []struct {
		name         string
		metrics      *AggregatedMetricsRequest
		authConfig   AuthConfig
		serverStatus int
		expectError  bool
		description  string
	}{
		{
			name: "successful_aggregated_submission",
			metrics: &AggregatedMetricsRequest{
				ServerUUID:  "server-uuid-agg",
				CollectedAt: time.Now().Format(time.RFC3339),
				CPU: &CPUAggregation{
					UsagePercent:  45.5,
					LoadAverage1:  1.2,
					LoadAverage5:  1.5,
					LoadAverage15: 1.8,
					CoreCount:     8,
				},
				Memory: &MemoryAggregation{
					TotalBytes:      16777216000,
					UsedBytes:       8388608000,
					FreeBytes:       8388608000,
					UsedPercent:     50.0,
					SwapTotalBytes:  4194304000,
					SwapUsedBytes:   1048576000,
					SwapUsedPercent: 25.0,
				},
			},
			authConfig: AuthConfig{
				Token: "test-jwt-token",
			},
			serverStatus: http.StatusOK,
			expectError:  false,
			description:  "Should submit aggregated metrics successfully",
		},
		{
			name: "auto_set_server_uuid_aggregated",
			metrics: &AggregatedMetricsRequest{
				ServerUUID:  "",
				CollectedAt: time.Now().Format(time.RFC3339),
				CPU: &CPUAggregation{
					UsagePercent: 50.0,
				},
			},
			authConfig: AuthConfig{
				ServerUUID:   "auto-agg-uuid",
				ServerSecret: "server-secret",
			},
			serverStatus: http.StatusAccepted,
			expectError:  false,
			description:  "Should auto-set server UUID for aggregated metrics",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/v1/metrics/aggregated", r.URL.Path)
				assert.Equal(t, "POST", r.Method)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"status": "success",
				})
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    tt.authConfig,
			})
			require.NoError(t, err)

			err = client.Metrics.SubmitAggregatedMetrics(context.Background(), tt.metrics)

			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

// TestMetricsService_Query tests the Query method
func TestMetricsService_Query(t *testing.T) {
	tests := []struct {
		name         string
		query        *MetricsQuery
		authConfig   AuthConfig
		serverStatus int
		mockMetrics  []*Metric
		expectError  bool
		description  string
	}{
		{
			name: "query_with_time_range",
			query: &MetricsQuery{
				ServerUUIDs: []string{"server-1", "server-2"},
				MetricNames: []string{"cpu.usage", "memory.usage"},
				StartTime:   time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
				EndTime:     time.Now().Format(time.RFC3339),
				Limit:       100,
			},
			authConfig: AuthConfig{
				Token: "test-jwt-token",
			},
			serverStatus: http.StatusOK,
			mockMetrics: []*Metric{
				{
					ServerUUID: "server-1",
					Name:       "cpu.usage",
					Value:      45.5,
					Unit:       "percent",
					Timestamp:  time.Now(),
				},
				{
					ServerUUID: "server-1",
					Name:       "memory.usage",
					Value:      60.2,
					Unit:       "percent",
					Timestamp:  time.Now(),
				},
			},
			expectError: false,
			description: "Should query metrics with time range filter",
		},
		{
			name: "query_with_aggregation",
			query: &MetricsQuery{
				ServerUUIDs: []string{"server-1"},
				MetricNames: []string{"cpu.usage"},
				StartTime:   time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
				EndTime:     time.Now().Format(time.RFC3339),
				GroupBy:     "hour",
				Aggregation: "avg",
			},
			authConfig: AuthConfig{
				Token: "test-jwt-token",
			},
			serverStatus: http.StatusOK,
			mockMetrics: []*Metric{
				{
					ServerUUID: "server-1",
					Name:       "cpu.usage",
					Value:      50.0,
					Unit:       "percent",
					Timestamp:  time.Now(),
				},
			},
			expectError: false,
			description: "Should query metrics with aggregation",
		},
		{
			name: "query_with_filters",
			query: &MetricsQuery{
				ServerUUIDs: []string{"server-1"},
				MetricNames: []string{"disk.usage"},
				StartTime:   time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
				EndTime:     time.Now().Format(time.RFC3339),
				Filters: map[string]interface{}{
					"device": "/dev/sda1",
					"mount":  "/",
				},
			},
			authConfig: AuthConfig{
				Token: "test-jwt-token",
			},
			serverStatus: http.StatusOK,
			mockMetrics: []*Metric{
				{
					ServerUUID: "server-1",
					Name:       "disk.usage",
					Value:      75.0,
					Unit:       "percent",
					Timestamp:  time.Now(),
					Tags: map[string]string{
						"device": "/dev/sda1",
						"mount":  "/",
					},
				},
			},
			expectError: false,
			description: "Should query metrics with custom filters",
		},
		{
			name: "query_error_bad_request",
			query: &MetricsQuery{
				StartTime: "invalid-time",
				EndTime:   "invalid-time",
			},
			authConfig: AuthConfig{
				Token: "test-jwt-token",
			},
			serverStatus: http.StatusBadRequest,
			mockMetrics:  nil,
			expectError:  true,
			description:  "Should handle bad request errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/v1/metrics/query", r.URL.Path)
				assert.Equal(t, "POST", r.Method)

				var body MetricsQuery
				err := json.NewDecoder(r.Body).Decode(&body)
				require.NoError(t, err)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)

				if tt.serverStatus == http.StatusOK {
					json.NewEncoder(w).Encode(map[string]interface{}{
						"status": "success",
						"data":   tt.mockMetrics,
					})
				} else {
					json.NewEncoder(w).Encode(map[string]interface{}{
						"status":  "error",
						"message": "Query error",
					})
				}
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    tt.authConfig,
			})
			require.NoError(t, err)

			metrics, err := client.Metrics.Query(context.Background(), tt.query)

			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				assert.Equal(t, len(tt.mockMetrics), len(metrics))
			}
		})
	}
}

// TestMetricsService_Get tests the Get method
func TestMetricsService_Get(t *testing.T) {
	tests := []struct {
		name         string
		serverUUID   string
		opts         *ListOptions
		authConfig   AuthConfig
		serverStatus int
		mockMetrics  []*Metric
		mockMeta     *PaginationMeta
		expectError  bool
		description  string
	}{
		{
			name:       "get_metrics_with_pagination",
			serverUUID: "server-uuid-123",
			opts: &ListOptions{
				Page:  1,
				Limit: 25,
			},
			authConfig: AuthConfig{
				Token: "test-jwt-token",
			},
			serverStatus: http.StatusOK,
			mockMetrics: []*Metric{
				{
					ServerUUID: "server-uuid-123",
					Name:       "cpu.usage",
					Value:      45.5,
					Timestamp:  time.Now(),
				},
			},
			mockMeta: &PaginationMeta{
				CurrentPage: 1,
				TotalPages:  5,
				TotalItems:  125,
				PerPage:     25,
			},
			expectError: false,
			description: "Should get metrics with pagination",
		},
		{
			name:       "get_metrics_with_search",
			serverUUID: "server-uuid-456",
			opts: &ListOptions{
				Page:   1,
				Limit:  50,
				Search: "cpu",
			},
			authConfig: AuthConfig{
				Token: "test-jwt-token",
			},
			serverStatus: http.StatusOK,
			mockMetrics: []*Metric{
				{
					ServerUUID: "server-uuid-456",
					Name:       "cpu.usage",
					Value:      50.0,
					Timestamp:  time.Now(),
				},
				{
					ServerUUID: "server-uuid-456",
					Name:       "cpu.temperature",
					Value:      65.0,
					Timestamp:  time.Now(),
				},
			},
			mockMeta: &PaginationMeta{
				CurrentPage: 1,
				TotalPages:  1,
				TotalItems:  2,
				PerPage:     50,
			},
			expectError: false,
			description: "Should get metrics with search filter",
		},
		{
			name:       "get_metrics_no_options",
			serverUUID: "server-uuid-789",
			opts:       nil,
			authConfig: AuthConfig{
				Token: "test-jwt-token",
			},
			serverStatus: http.StatusOK,
			mockMetrics: []*Metric{
				{
					ServerUUID: "server-uuid-789",
					Name:       "memory.usage",
					Value:      60.0,
					Timestamp:  time.Now(),
				},
			},
			mockMeta: &PaginationMeta{
				CurrentPage: 1,
				TotalPages:  1,
				TotalItems:  1,
				PerPage:     10,
			},
			expectError: false,
			description: "Should get metrics without options",
		},
		{
			name:       "get_metrics_server_not_found",
			serverUUID: "non-existent-uuid",
			opts:       nil,
			authConfig: AuthConfig{
				Token: "test-jwt-token",
			},
			serverStatus: http.StatusNotFound,
			mockMetrics:  nil,
			mockMeta:     nil,
			expectError:  true,
			description:  "Should handle server not found error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/v1/metrics/server/"+tt.serverUUID, r.URL.Path)
				assert.Equal(t, "GET", r.Method)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)

				if tt.serverStatus == http.StatusOK {
					json.NewEncoder(w).Encode(map[string]interface{}{
						"status": "success",
						"data":   tt.mockMetrics,
						"meta":   tt.mockMeta,
					})
				} else {
					json.NewEncoder(w).Encode(map[string]interface{}{
						"status":  "error",
						"message": "Server not found",
					})
				}
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    tt.authConfig,
			})
			require.NoError(t, err)

			metrics, meta, err := client.Metrics.Get(context.Background(), tt.serverUUID, tt.opts)

			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, metrics)
				assert.Nil(t, meta)
			} else {
				assert.NoError(t, err, tt.description)
				assert.Equal(t, len(tt.mockMetrics), len(metrics))
				if tt.mockMeta != nil {
					assert.Equal(t, tt.mockMeta.CurrentPage, meta.CurrentPage)
					assert.Equal(t, tt.mockMeta.TotalItems, meta.TotalItems)
				}
			}
		})
	}
}

// TestMetricsService_GetSummary tests the GetSummary method
func TestMetricsService_GetSummary(t *testing.T) {
	tests := []struct {
		name         string
		serverUUID   string
		timeRange    *QueryTimeRange
		authConfig   AuthConfig
		serverStatus int
		mockSummary  map[string]interface{}
		expectError  bool
		description  string
	}{
		{
			name:       "get_summary_with_time_range",
			serverUUID: "server-uuid-123",
			timeRange: &QueryTimeRange{
				Start: time.Now().Add(-24 * time.Hour),
				End:   time.Now(),
			},
			authConfig: AuthConfig{
				Token: "test-jwt-token",
			},
			serverStatus: http.StatusOK,
			mockSummary: map[string]interface{}{
				"cpu": map[string]interface{}{
					"avg": 45.5,
					"min": 20.0,
					"max": 80.0,
				},
				"memory": map[string]interface{}{
					"avg": 60.2,
					"min": 50.0,
					"max": 75.0,
				},
			},
			expectError: false,
			description: "Should get summary with time range",
		},
		{
			name:       "get_summary_without_time_range",
			serverUUID: "server-uuid-456",
			timeRange:  nil,
			authConfig: AuthConfig{
				Token: "test-jwt-token",
			},
			serverStatus: http.StatusOK,
			mockSummary: map[string]interface{}{
				"cpu": map[string]interface{}{
					"current": 50.0,
				},
			},
			expectError: false,
			description: "Should get summary without time range",
		},
		{
			name:       "get_summary_server_not_found",
			serverUUID: "non-existent-uuid",
			timeRange:  nil,
			authConfig: AuthConfig{
				Token: "test-jwt-token",
			},
			serverStatus: http.StatusNotFound,
			mockSummary:  nil,
			expectError:  true,
			description:  "Should handle server not found error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/v1/metrics/server/"+tt.serverUUID+"/summary", r.URL.Path)
				assert.Equal(t, "GET", r.Method)

				if tt.timeRange != nil {
					query := r.URL.Query()
					assert.NotEmpty(t, query.Get("start"))
					assert.NotEmpty(t, query.Get("end"))
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)

				if tt.serverStatus == http.StatusOK {
					json.NewEncoder(w).Encode(map[string]interface{}{
						"status": "success",
						"data":   tt.mockSummary,
					})
				} else {
					json.NewEncoder(w).Encode(map[string]interface{}{
						"status":  "error",
						"message": "Server not found",
					})
				}
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    tt.authConfig,
			})
			require.NoError(t, err)

			summary, err := client.Metrics.GetSummary(context.Background(), tt.serverUUID, tt.timeRange)

			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, summary)
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, summary)
				if tt.mockSummary != nil {
					assert.Equal(t, len(tt.mockSummary), len(summary))
				}
			}
		})
	}
}

// TestMetricsService_GetAggregated tests the GetAggregated method
func TestMetricsService_GetAggregated(t *testing.T) {
	tests := []struct {
		name         string
		aggregation  *MetricsAggregation
		authConfig   AuthConfig
		serverStatus int
		mockResult   map[string]interface{}
		expectError  bool
		description  string
	}{
		{
			name: "aggregate_average_cpu",
			aggregation: &MetricsAggregation{
				ServerUUIDs: []string{"server-1", "server-2"},
				MetricNames: []string{"cpu.usage"},
				StartTime:   time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
				EndTime:     time.Now().Format(time.RFC3339),
				GroupBy:     []string{"server_uuid"},
				Function:    "avg",
				Interval:    "1h",
			},
			authConfig: AuthConfig{
				Token: "test-jwt-token",
			},
			serverStatus: http.StatusOK,
			mockResult: map[string]interface{}{
				"server-1": 45.5,
				"server-2": 50.2,
			},
			expectError: false,
			description: "Should aggregate metrics with average function",
		},
		{
			name: "aggregate_max_memory",
			aggregation: &MetricsAggregation{
				ServerUUIDs: []string{"server-1"},
				MetricNames: []string{"memory.usage"},
				StartTime:   time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
				EndTime:     time.Now().Format(time.RFC3339),
				GroupBy:     []string{"timestamp"},
				Function:    "max",
				Interval:    "5m",
			},
			authConfig: AuthConfig{
				Token: "test-jwt-token",
			},
			serverStatus: http.StatusOK,
			mockResult: map[string]interface{}{
				"max_value": 75.0,
			},
			expectError: false,
			description: "Should aggregate metrics with max function",
		},
		{
			name: "aggregate_error_bad_request",
			aggregation: &MetricsAggregation{
				MetricNames: []string{},
				StartTime:   "invalid",
				EndTime:     "invalid",
				GroupBy:     []string{},
				Function:    "invalid",
			},
			authConfig: AuthConfig{
				Token: "test-jwt-token",
			},
			serverStatus: http.StatusBadRequest,
			mockResult:   nil,
			expectError:  true,
			description:  "Should handle bad request errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/v1/metrics/aggregate", r.URL.Path)
				assert.Equal(t, "POST", r.Method)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)

				if tt.serverStatus == http.StatusOK {
					json.NewEncoder(w).Encode(map[string]interface{}{
						"status": "success",
						"data":   tt.mockResult,
					})
				} else {
					json.NewEncoder(w).Encode(map[string]interface{}{
						"status":  "error",
						"message": "Bad request",
					})
				}
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    tt.authConfig,
			})
			require.NoError(t, err)

			result, err := client.Metrics.GetAggregated(context.Background(), tt.aggregation)

			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, result)
				if tt.mockResult != nil {
					assert.Equal(t, len(tt.mockResult), len(result))
				}
			}
		})
	}
}

// TestMetricsService_Export tests the Export method
func TestMetricsService_Export(t *testing.T) {
	tests := []struct {
		name         string
		export       *MetricsExport
		authConfig   AuthConfig
		serverStatus int
		mockData     []byte
		expectError  bool
		description  string
	}{
		{
			name: "export_csv_format",
			export: &MetricsExport{
				ServerUUIDs: []string{"server-1", "server-2"},
				MetricNames: []string{"cpu.usage", "memory.usage"},
				StartTime:   time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
				EndTime:     time.Now().Format(time.RFC3339),
				Format:      "csv",
			},
			authConfig: AuthConfig{
				Token: "test-jwt-token",
			},
			serverStatus: http.StatusOK,
			mockData:     []byte("timestamp,server_uuid,metric_name,value\n2023-01-01T00:00:00Z,server-1,cpu.usage,45.5\n"),
			expectError:  false,
			description:  "Should export metrics in CSV format",
		},
		{
			name: "export_json_format",
			export: &MetricsExport{
				ServerUUIDs: []string{"server-1"},
				MetricNames: []string{"cpu.usage"},
				StartTime:   time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
				EndTime:     time.Now().Format(time.RFC3339),
				Format:      "json",
			},
			authConfig: AuthConfig{
				Token: "test-jwt-token",
			},
			serverStatus: http.StatusOK,
			mockData:     []byte(`[{"timestamp":"2023-01-01T00:00:00Z","server_uuid":"server-1","metric_name":"cpu.usage","value":45.5}]`),
			expectError:  false,
			description:  "Should export metrics in JSON format",
		},
		{
			name: "export_prometheus_format",
			export: &MetricsExport{
				ServerUUIDs: []string{"server-1"},
				StartTime:   time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
				EndTime:     time.Now().Format(time.RFC3339),
				Format:      "prometheus",
			},
			authConfig: AuthConfig{
				Token: "test-jwt-token",
			},
			serverStatus: http.StatusOK,
			mockData:     []byte("# TYPE cpu_usage gauge\ncpu_usage{server_uuid=\"server-1\"} 45.5\n"),
			expectError:  false,
			description:  "Should export metrics in Prometheus format",
		},
		{
			name: "export_error_invalid_format",
			export: &MetricsExport{
				ServerUUIDs: []string{"server-1"},
				StartTime:   time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
				EndTime:     time.Now().Format(time.RFC3339),
				Format:      "invalid",
			},
			authConfig: AuthConfig{
				Token: "test-jwt-token",
			},
			serverStatus: http.StatusBadRequest,
			mockData:     nil,
			expectError:  true,
			description:  "Should handle invalid format error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/v1/metrics/export", r.URL.Path)
				assert.Equal(t, "POST", r.Method)

				w.WriteHeader(tt.serverStatus)

				if tt.serverStatus == http.StatusOK {
					w.Write(tt.mockData)
				} else {
					json.NewEncoder(w).Encode(map[string]interface{}{
						"status":  "error",
						"message": "Invalid format",
					})
				}
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    tt.authConfig,
			})
			require.NoError(t, err)

			data, err := client.Metrics.Export(context.Background(), tt.export)

			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, data)
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, data)
				assert.Equal(t, tt.mockData, data)
			}
		})
	}
}

// TestMetricsService_GetStatus tests the GetStatus method
func TestMetricsService_GetStatus(t *testing.T) {
	tests := []struct {
		name         string
		serverUUID   string
		authConfig   AuthConfig
		serverStatus int
		mockStatus   *MetricsStatus
		expectError  bool
		description  string
	}{
		{
			name:       "get_status_enabled",
			serverUUID: "server-uuid-123",
			authConfig: AuthConfig{
				Token: "test-jwt-token",
			},
			serverStatus: http.StatusOK,
			mockStatus: &MetricsStatus{
				ServerUUID:         "server-uuid-123",
				CollectionEnabled:  true,
				CollectionInterval: 60,
				ErrorCount:         0,
			},
			expectError: false,
			description: "Should get metrics status when enabled",
		},
		{
			name:       "get_status_disabled_with_errors",
			serverUUID: "server-uuid-456",
			authConfig: AuthConfig{
				Token: "test-jwt-token",
			},
			serverStatus: http.StatusOK,
			mockStatus: &MetricsStatus{
				ServerUUID:         "server-uuid-456",
				CollectionEnabled:  false,
				CollectionInterval: 60,
				ErrorCount:         5,
				LastError:          "Connection timeout",
			},
			expectError: false,
			description: "Should get metrics status with errors",
		},
		{
			name:       "get_status_server_not_found",
			serverUUID: "non-existent-uuid",
			authConfig: AuthConfig{
				Token: "test-jwt-token",
			},
			serverStatus: http.StatusNotFound,
			mockStatus:   nil,
			expectError:  true,
			description:  "Should handle server not found error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/v1/metrics/"+tt.serverUUID+"/status", r.URL.Path)
				assert.Equal(t, "GET", r.Method)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)

				if tt.serverStatus == http.StatusOK {
					json.NewEncoder(w).Encode(map[string]interface{}{
						"status": "success",
						"data":   tt.mockStatus,
					})
				} else {
					json.NewEncoder(w).Encode(map[string]interface{}{
						"status":  "error",
						"message": "Server not found",
					})
				}
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    tt.authConfig,
			})
			require.NoError(t, err)

			status, err := client.Metrics.GetStatus(context.Background(), tt.serverUUID)

			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, status)
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, status)
				assert.Equal(t, tt.mockStatus.ServerUUID, status.ServerUUID)
				assert.Equal(t, tt.mockStatus.CollectionEnabled, status.CollectionEnabled)
			}
		})
	}
}

// TestMetricsService_GetServerMetrics tests the GetServerMetrics method
func TestMetricsService_GetServerMetrics(t *testing.T) {
	tests := []struct {
		name         string
		serverUUID   string
		metricName   string
		timeRange    *TimeRange
		authConfig   AuthConfig
		serverStatus int
		mockMetrics  []interface{}
		expectError  bool
		description  string
	}{
		{
			name:       "get_specific_metric_with_time_range",
			serverUUID: "server-uuid-123",
			metricName: "cpu.usage",
			timeRange: &TimeRange{
				Start: time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
				End:   time.Now().Format(time.RFC3339),
			},
			authConfig: AuthConfig{
				Token: "test-jwt-token",
			},
			serverStatus: http.StatusOK,
			mockMetrics: []interface{}{
				map[string]interface{}{
					"timestamp": time.Now().Format(time.RFC3339),
					"value":     45.5,
				},
				map[string]interface{}{
					"timestamp": time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
					"value":     48.2,
				},
			},
			expectError: false,
			description: "Should get specific metric with time range",
		},
		{
			name:       "get_metric_without_time_range",
			serverUUID: "server-uuid-456",
			metricName: "memory.usage",
			timeRange:  nil,
			authConfig: AuthConfig{
				Token: "test-jwt-token",
			},
			serverStatus: http.StatusOK,
			mockMetrics: []interface{}{
				map[string]interface{}{
					"timestamp": time.Now().Format(time.RFC3339),
					"value":     60.0,
				},
			},
			expectError: false,
			description: "Should get metric without time range",
		},
		{
			name:       "get_metric_not_found",
			serverUUID: "server-uuid-789",
			metricName: "nonexistent.metric",
			timeRange:  nil,
			authConfig: AuthConfig{
				Token: "test-jwt-token",
			},
			serverStatus: http.StatusNotFound,
			mockMetrics:  nil,
			expectError:  true,
			description:  "Should handle metric not found error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/v1/metrics/server/"+tt.serverUUID, r.URL.Path)
				assert.Equal(t, "GET", r.Method)

				query := r.URL.Query()
				assert.Equal(t, tt.metricName, query.Get("metric"))

				if tt.timeRange != nil {
					assert.Equal(t, tt.timeRange.Start, query.Get("start"))
					assert.Equal(t, tt.timeRange.End, query.Get("end"))
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)

				if tt.serverStatus == http.StatusOK {
					json.NewEncoder(w).Encode(map[string]interface{}{
						"status": "success",
						"data":   tt.mockMetrics,
					})
				} else {
					json.NewEncoder(w).Encode(map[string]interface{}{
						"status":  "error",
						"message": "Not found",
					})
				}
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    tt.authConfig,
			})
			require.NoError(t, err)

			metrics, err := client.Metrics.GetServerMetrics(context.Background(), tt.serverUUID, tt.metricName, tt.timeRange)

			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, metrics)
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, metrics)
				assert.Equal(t, len(tt.mockMetrics), len(metrics))
			}
		})
	}
}

// TestMetricsAggregator_NewMetricsAggregator tests NewMetricsAggregator with different input types
func TestMetricsAggregator_NewMetricsAggregator(t *testing.T) {
	now := time.Now()
	cpuPercent := 45.0
	memPercent := 60.0

	tests := []struct {
		name        string
		input       []interface{}
		expectedLen int
		description string
	}{
		{
			name: "create_with_slice_of_pointers",
			input: []interface{}{
				[]*ComprehensiveMetricsTimescale{
					{
						Timestamp:          now,
						CPUUsagePercent:    &cpuPercent,
						MemoryUsagePercent: &memPercent,
					},
				},
			},
			expectedLen: 1,
			description: "Should create aggregator with slice of pointers",
		},
		{
			name: "create_with_slice_of_values",
			input: []interface{}{
				[]ComprehensiveMetricsTimescale{
					{
						Timestamp:          now,
						CPUUsagePercent:    &cpuPercent,
						MemoryUsagePercent: &memPercent,
					},
				},
			},
			expectedLen: 1,
			description: "Should create aggregator with slice of values",
		},
		{
			name: "create_with_single_pointer",
			input: []interface{}{
				&ComprehensiveMetricsTimescale{
					Timestamp:          now,
					CPUUsagePercent:    &cpuPercent,
					MemoryUsagePercent: &memPercent,
				},
			},
			expectedLen: 1,
			description: "Should create aggregator with single pointer",
		},
		{
			name: "create_with_single_value",
			input: []interface{}{
				ComprehensiveMetricsTimescale{
					Timestamp:          now,
					CPUUsagePercent:    &cpuPercent,
					MemoryUsagePercent: &memPercent,
				},
			},
			expectedLen: 1,
			description: "Should create aggregator with single value",
		},
		{
			name:        "create_with_no_input",
			input:       []interface{}{},
			expectedLen: 0,
			description: "Should create empty aggregator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aggregator := NewMetricsAggregator(tt.input...)
			assert.NotNil(t, aggregator)
			assert.Equal(t, tt.expectedLen, len(aggregator.metrics), tt.description)
		})
	}
}

// TestMetricsAggregator_AddMetrics tests AddMetrics method
func TestMetricsAggregator_AddMetrics(t *testing.T) {
	now := time.Now()
	cpuPercent := 45.0
	memPercent := 60.0

	aggregator := NewMetricsAggregator()
	assert.Equal(t, 0, len(aggregator.metrics))

	// Add metrics
	aggregator.AddMetrics(
		&ComprehensiveMetricsTimescale{
			Timestamp:          now,
			CPUUsagePercent:    &cpuPercent,
			MemoryUsagePercent: &memPercent,
		},
		&ComprehensiveMetricsTimescale{
			Timestamp:          now.Add(1 * time.Minute),
			CPUUsagePercent:    &cpuPercent,
			MemoryUsagePercent: &memPercent,
		},
	)

	assert.Equal(t, 2, len(aggregator.metrics))
}

// TestMetricsAggregator_Aggregate tests Aggregate method
func TestMetricsAggregator_Aggregate(t *testing.T) {
	now := time.Now()
	cpuPercent := 45.0

	aggregator := NewMetricsAggregator(&ComprehensiveMetricsTimescale{
		Timestamp:       now,
		CPUUsagePercent: &cpuPercent,
	})

	result := aggregator.WithGroupBy("server_uuid").WithFunction("avg").Aggregate()

	assert.NotNil(t, result)
	assert.Equal(t, 1, result["count"])
	assert.Equal(t, "server_uuid", result["groupBy"])
	assert.Equal(t, "avg", result["function"])
}

// TestMetricsService_Context_Cancellation tests context cancellation
func TestMetricsService_Context_Cancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-jwt-token",
		},
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	metrics := []*Metric{
		{
			// ServerUUID parameter provided to Submit() method
			Name:       "cpu.usage",
			Value:      50.0,
			Timestamp:  time.Now(),
		},
	}

	err = client.Metrics.Submit(ctx, "test-uuid", metrics)
	assert.Error(t, err, "Should return error on context cancellation")
}

// TestMetricsService_API_Key_Authentication tests API key authentication
func TestMetricsService_API_Key_Authentication(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.Header.Get("Access-Key"), "test-api-key")
		assert.Contains(t, r.Header.Get("Access-Secret"), "test-api-secret")

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			APIKey:    "test-api-key",
			APISecret: "test-api-secret",
		},
	})
	require.NoError(t, err)

	metrics := []*Metric{
		{
			// ServerUUID parameter provided to Submit() method
			Name:       "cpu.usage",
			Value:      50.0,
			Timestamp:  time.Now(),
		},
	}

	err = client.Metrics.Submit(context.Background(), "test-uuid", metrics)
	assert.NoError(t, err)
}

// TestMetricsService_Multiple_Servers tests metrics for multiple servers
func TestMetricsService_Multiple_Servers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		metricsArray := body["metrics"].([]interface{})
		assert.GreaterOrEqual(t, len(metricsArray), 2)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-jwt-token",
		},
	})
	require.NoError(t, err)

	metrics := []*Metric{
		{
			ServerUUID: "server-1",
			Name:       "cpu.usage",
			Value:      45.0,
			Timestamp:  time.Now(),
		},
		{
			ServerUUID: "server-2",
			Name:       "cpu.usage",
			Value:      50.0,
			Timestamp:  time.Now(),
		},
		{
			ServerUUID: "server-3",
			Name:       "cpu.usage",
			Value:      55.0,
			Timestamp:  time.Now(),
		},
	}

	err = client.Metrics.Submit(context.Background(), "batch-submit", metrics)
	assert.NoError(t, err)
}

// TestMetricsService_Edge_Cases tests edge cases
func TestMetricsService_Edge_Cases(t *testing.T) {
	t.Run("nil_metrics_array", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
		}))
		defer server.Close()

		client, err := NewClient(&Config{
			BaseURL: server.URL,
			Auth: AuthConfig{
				Token: "test-jwt-token",
			},
		})
		require.NoError(t, err)

		err = client.Metrics.Submit(context.Background(), "test-uuid", nil)
		assert.NoError(t, err)
	})

	t.Run("empty_server_uuid", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "error",
				"message": "Server UUID required",
			})
		}))
		defer server.Close()

		client, err := NewClient(&Config{
			BaseURL: server.URL,
			Auth: AuthConfig{
				Token: "test-jwt-token",
			},
		})
		require.NoError(t, err)

		metrics := []*Metric{{Name: "cpu.usage", Value: 50.0}}
		err = client.Metrics.Submit(context.Background(), "", metrics)
		assert.Error(t, err)
	})

	t.Run("very_large_metrics_array", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
		}))
		defer server.Close()

		client, err := NewClient(&Config{
			BaseURL: server.URL,
			Auth: AuthConfig{
				Token: "test-jwt-token",
			},
		})
		require.NoError(t, err)

		// Create 1000 metrics
		metrics := make([]*Metric, 1000)
		for i := 0; i < 1000; i++ {
			metrics[i] = &Metric{
				// ServerUUID parameter provided to Submit() method
				Name:       "cpu.usage",
				Value:      float64(i % 100),
				Timestamp:  time.Now(),
			}
		}

		err = client.Metrics.Submit(context.Background(), "test-uuid", metrics)
		assert.NoError(t, err)
	})
}
// TestMetricsService_GetLatestMetrics tests the GetLatestMetrics method with various scenarios
func TestMetricsService_GetLatestMetrics(t *testing.T) {
	tests := []struct {
		name       string
		serverUUID string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *TimescaleMetricsResponse)
	}{
		{
			name:       "success - with full metrics",
			serverUUID: "server-123",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"server_uuid": "server-123",
					"timestamp":   "2024-01-01T12:00:00Z",
					"metrics": map[string]interface{}{
						"server_uuid":          "server-123",
						"collected_at":         "2024-01-01T12:00:00Z",
						"agent_version":        "1.0.0",
						"collection_duration":  1.5,
						"cpu_usage_percent":    45.5,
						"memory_usage_percent": 62.3,
					},
					"source": "timescaledb",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, resp *TimescaleMetricsResponse) {
				assert.Equal(t, "server-123", resp.ServerUUID)
				assert.Equal(t, "2024-01-01T12:00:00Z", resp.Timestamp)
				assert.NotNil(t, resp.Metrics)
				assert.Equal(t, "timescaledb", resp.Source)
			},
		},
		{
			name:       "success - minimal metrics",
			serverUUID: "server-456",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"server_uuid": "server-456",
					"timestamp":   "2024-01-01T13:00:00Z",
					"metrics": map[string]interface{}{
						"server_uuid":         "server-456",
						"collected_at":        "2024-01-01T13:00:00Z",
						"agent_version":       "1.0.1",
						"collection_duration": 0.8,
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, resp *TimescaleMetricsResponse) {
				assert.Equal(t, "server-456", resp.ServerUUID)
				assert.NotNil(t, resp.Metrics)
			},
		},
		{
			name:       "not found - server doesn't exist",
			serverUUID: "nonexistent",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Server not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			serverUUID: "server-123",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "forbidden - no access",
			serverUUID: "server-789",
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Access denied",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			serverUUID: "server-123",
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Internal server error",
			},
			wantErr: true,
		},
		{
			name:       "empty server UUID",
			serverUUID: "",
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Server UUID required",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				if tt.serverUUID != "" {
					assert.Contains(t, r.URL.Path, tt.serverUUID)
				}
				assert.Contains(t, r.URL.Path, "/v2/servers/")
				assert.Contains(t, r.URL.Path, "/metrics/latest")

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

			// Use timeout context for error scenarios with 500 status
			ctx := context.Background()
			if tt.wantErr && tt.mockStatus >= 500 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
			}

			result, err := client.Metrics.GetLatestMetrics(ctx, tt.serverUUID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestMetricsService_GetMetricsRange tests the GetMetricsRange method with various scenarios
func TestMetricsService_GetMetricsRange(t *testing.T) {
	tests := []struct {
		name       string
		serverUUID string
		startTime  string
		endTime    string
		limit      int
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *TimescaleMetricsRangeResponse)
	}{
		{
			name:       "success - with limit",
			serverUUID: "server-123",
			startTime:  "2024-01-01T00:00:00Z",
			endTime:    "2024-01-01T23:59:59Z",
			limit:      100,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"server_uuid": "server-123",
					"start_time":  "2024-01-01T00:00:00Z",
					"end_time":    "2024-01-01T23:59:59Z",
					"metrics": []map[string]interface{}{
						{
							"server_uuid":          "server-123",
							"collected_at":         "2024-01-01T12:00:00Z",
							"agent_version":        "1.0.0",
							"collection_duration":  1.2,
							"cpu_usage_percent":    45.5,
							"memory_usage_percent": 62.3,
						},
						{
							"server_uuid":          "server-123",
							"collected_at":         "2024-01-01T13:00:00Z",
							"agent_version":        "1.0.0",
							"collection_duration":  1.1,
							"cpu_usage_percent":    48.2,
							"memory_usage_percent": 64.1,
						},
					},
					"count":  2,
					"source": "timescaledb",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, resp *TimescaleMetricsRangeResponse) {
				assert.Equal(t, "server-123", resp.ServerUUID)
				assert.Equal(t, "2024-01-01T00:00:00Z", resp.StartTime)
				assert.Equal(t, "2024-01-01T23:59:59Z", resp.EndTime)
				assert.Len(t, resp.Metrics, 2)
				assert.Equal(t, 2, resp.Count)
				assert.Equal(t, "timescaledb", resp.Source)
			},
		},
		{
			name:       "success - without limit (unlimited)",
			serverUUID: "server-456",
			startTime:  "2024-01-01T00:00:00Z",
			endTime:    "2024-01-02T00:00:00Z",
			limit:      0, // No limit
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"server_uuid": "server-456",
					"start_time":  "2024-01-01T00:00:00Z",
					"end_time":    "2024-01-02T00:00:00Z",
					"metrics":     []map[string]interface{}{},
					"count":       0,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, resp *TimescaleMetricsRangeResponse) {
				assert.Equal(t, "server-456", resp.ServerUUID)
				assert.Len(t, resp.Metrics, 0)
				assert.Equal(t, 0, resp.Count)
			},
		},
		{
			name:       "success - large result set",
			serverUUID: "server-789",
			startTime:  "2024-01-01T00:00:00Z",
			endTime:    "2024-01-07T00:00:00Z",
			limit:      1000,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"server_uuid": "server-789",
					"start_time":  "2024-01-01T00:00:00Z",
					"end_time":    "2024-01-07T00:00:00Z",
					"metrics": func() []map[string]interface{} {
						metrics := make([]map[string]interface{}, 500)
						for i := 0; i < 500; i++ {
							metrics[i] = map[string]interface{}{
								"server_uuid":         "server-789",
								"collected_at":        "2024-01-01T12:00:00Z",
								"agent_version":       "1.0.0",
								"collection_duration": 1.0,
							}
						}
						return metrics
					}(),
					"count": 500,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, resp *TimescaleMetricsRangeResponse) {
				assert.Equal(t, "server-789", resp.ServerUUID)
				assert.Len(t, resp.Metrics, 500)
				assert.Equal(t, 500, resp.Count)
			},
		},
		{
			name:       "validation error - invalid time range",
			serverUUID: "server-123",
			startTime:  "invalid-time",
			endTime:    "2024-01-01T23:59:59Z",
			limit:      100,
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Invalid time format",
			},
			wantErr: true,
		},
		{
			name:       "not found - server doesn't exist",
			serverUUID: "nonexistent",
			startTime:  "2024-01-01T00:00:00Z",
			endTime:    "2024-01-01T23:59:59Z",
			limit:      100,
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Server not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			serverUUID: "server-123",
			startTime:  "2024-01-01T00:00:00Z",
			endTime:    "2024-01-01T23:59:59Z",
			limit:      100,
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "forbidden",
			serverUUID: "server-123",
			startTime:  "2024-01-01T00:00:00Z",
			endTime:    "2024-01-01T23:59:59Z",
			limit:      100,
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Access denied",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			serverUUID: "server-123",
			startTime:  "2024-01-01T00:00:00Z",
			endTime:    "2024-01-01T23:59:59Z",
			limit:      100,
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
				assert.Contains(t, r.URL.Path, tt.serverUUID)
				assert.Contains(t, r.URL.Path, "/v2/servers/")
				assert.Contains(t, r.URL.Path, "/metrics/range")

				// Verify query parameters
				assert.Equal(t, tt.startTime, r.URL.Query().Get("start_time"))
				assert.Equal(t, tt.endTime, r.URL.Query().Get("end_time"))
				if tt.limit > 0 {
					assert.Equal(t, fmt.Sprintf("%d", tt.limit), r.URL.Query().Get("limit"))
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

			// Use timeout context for error scenarios with 500 status
			ctx := context.Background()
			if tt.wantErr && tt.mockStatus >= 500 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
			}

			result, err := client.Metrics.GetMetricsRange(ctx, tt.serverUUID, tt.startTime, tt.endTime, tt.limit)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}
// TestMetricsService_SubmitComprehensiveToTimescale tests the SubmitComprehensiveToTimescale method with various scenarios
func TestMetricsService_SubmitComprehensiveToTimescale(t *testing.T) {
	tests := []struct {
		name       string
		metrics    *ComprehensiveMetricsSubmission
		serverAuth *AuthConfig // Optional server auth to test auto-population
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *http.Request)
	}{
		{
			name: "success - full metrics submission",
			metrics: &ComprehensiveMetricsSubmission{
				Metrics: &ComprehensiveMetricsPayload{
					ServerUUID:  "server-123",
					CollectedAt: time.Now().Format(time.RFC3339),
					CPU: &TimescaleCPUMetrics{
						UsagePercent: 45.5,
					},
					Memory: &TimescaleMemoryMetrics{
						UsedPercent: 62.3,
					},
				},
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status":  "success",
				"message": "Metrics submitted successfully",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Contains(t, r.URL.Path, "/v2/metrics/comprehensive")
			},
		},
		{
			name: "success - with server auth auto-population",
			metrics: &ComprehensiveMetricsSubmission{
				Metrics: &ComprehensiveMetricsPayload{
					ServerUUID:  "", // Empty, should be auto-populated
					CollectedAt: time.Now().Format(time.RFC3339),
				},
			},
			serverAuth: &AuthConfig{
				ServerUUID:   "auto-server-uuid",
				ServerSecret: "auto-secret",
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status":  "success",
				"message": "Metrics submitted",
			},
			wantErr: false,
		},
		{
			name: "success - minimal metrics",
			metrics: &ComprehensiveMetricsSubmission{
				Metrics: &ComprehensiveMetricsPayload{
					ServerUUID:  "server-456",
					CollectedAt: time.Now().Format(time.RFC3339),
				},
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
			},
			wantErr: false,
		},
		{
			name: "validation error - missing server UUID",
			metrics: &ComprehensiveMetricsSubmission{
				Metrics: &ComprehensiveMetricsPayload{
					ServerUUID:  "",
					CollectedAt: time.Now().Format(time.RFC3339),
				},
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Server UUID required",
			},
			wantErr: true,
		},
		{
			name: "validation error - invalid metrics data",
			metrics: &ComprehensiveMetricsSubmission{
				Metrics: &ComprehensiveMetricsPayload{
					ServerUUID:  "server-123",
					CollectedAt: "invalid-date",
				},
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Invalid metrics data",
			},
			wantErr: true,
		},
		{
			name: "unauthorized",
			metrics: &ComprehensiveMetricsSubmission{
				Metrics: &ComprehensiveMetricsPayload{
					ServerUUID:  "server-123",
					CollectedAt: time.Now().Format(time.RFC3339),
				},
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
			metrics: &ComprehensiveMetricsSubmission{
				Metrics: &ComprehensiveMetricsPayload{
					ServerUUID:  "server-123",
					CollectedAt: time.Now().Format(time.RFC3339),
				},
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
			metrics: &ComprehensiveMetricsSubmission{
				Metrics: &ComprehensiveMetricsPayload{
					ServerUUID:  "server-123",
					CollectedAt: time.Now().Format(time.RFC3339),
				},
			},
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Internal server error",
			},
			wantErr: true,
		},
		{
			name:       "nil metrics",
			metrics:    &ComprehensiveMetricsSubmission{Metrics: nil},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Metrics required",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var lastRequest *http.Request
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				lastRequest = r
				if tt.checkFunc != nil {
					tt.checkFunc(t, r)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			auth := AuthConfig{Token: "test-token"}
			if tt.serverAuth != nil {
				auth = *tt.serverAuth
			}

			client, err := NewClient(&Config{
				BaseURL:    server.URL,
				Auth:       auth,
				RetryCount: 0,
			})
			require.NoError(t, err)

			// Use timeout context for error scenarios with 500 status
			ctx := context.Background()
			if tt.wantErr && tt.mockStatus >= 500 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
			}

			err = client.Metrics.SubmitComprehensiveToTimescale(ctx, tt.metrics)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, lastRequest)
			}
		})
	}
}

// TestMetricsService_NetworkErrors tests handling of network-level errors
func TestMetricsService_NetworkErrors(t *testing.T) {
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
			operation:     "get",
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
			operation:     "submit",
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
			operation:     "query",
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
			operation:     "aggregate",
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
			case "get":
				_, _, apiErr = client.Metrics.Get(ctx, "test-uuid", nil)
			case "submit":
				metrics := []*Metric{{Name: "test"}}
				apiErr = client.Metrics.Submit(ctx, "test-uuid", metrics)
			case "query":
				query := &MetricsQuery{ServerUUIDs: []string{"test-uuid"}}
				_, apiErr = client.Metrics.Query(ctx, query)
			case "aggregate":
				agg := &MetricsAggregation{Function: "avg"}
				_, apiErr = client.Metrics.GetAggregated(ctx, agg)
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
