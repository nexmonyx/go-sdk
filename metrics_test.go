package nexmonyx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMetricsBuilder tests the metrics builder functionality
func TestMetricsBuilder(t *testing.T) {
	builder := NewMetricsBuilder("test-host")

	metrics := builder.
		WithAgentVersion("2.0.0").
		WithCollectionDuration(0.125).
		WithErrorCount(0).
		WithCPUMetrics(&TimescaleCPUMetrics{
			UsagePercent: 45.5,
			LoadAverage: &LoadAverage{
				Load1:  1.2,
				Load5:  1.5,
				Load15: 1.8,
			},
		}).
		WithMemoryMetrics(&TimescaleMemoryMetrics{
			Total:       16777216000,
			Used:        8388608000,
			UsedPercent: 50.0,
		}).
		Build()

	assert.Equal(t, "test-host", metrics.Hostname)
	assert.Equal(t, "2.0.0", metrics.AgentVersion)
	assert.Equal(t, 0.125, metrics.CollectionDuration)
	assert.Equal(t, 0, metrics.ErrorCount)
	assert.NotNil(t, metrics.Metrics.CPU)
	assert.Equal(t, 45.5, metrics.Metrics.CPU.UsagePercent)
	assert.NotNil(t, metrics.Metrics.Memory)
	assert.Equal(t, uint64(16777216000), metrics.Metrics.Memory.Total)
}

// TestConvertLegacyToTimescaleMetrics tests the legacy format conversion
func TestConvertLegacyToTimescaleMetrics(t *testing.T) {
	legacy := &ComprehensiveMetricsRequest{
		ServerUUID:  "test-uuid",
		CollectedAt: time.Now().Format(time.RFC3339),
		SystemInfo: &SystemInfo{
			Hostname:        "test-host",
			OS:              "linux",
			OSVersion:       "Ubuntu 22.04",
			KernelVersion:   "5.15.0",
			CPUArchitecture: "x86_64",
			Uptime:          3600,
		},
		CPU: &CPUMetrics{
			UsagePercent:  45.5,
			LoadAverage1:  1.2,
			LoadAverage5:  1.5,
			LoadAverage15: 1.8,
			CoreCount:     4,
			PerCoreUsage:  []float64{40.1, 45.2, 48.3, 48.4},
		},
		Memory: &MemoryMetrics{
			TotalBytes:       16777216000,
			UsedBytes:        8388608000,
			FreeBytes:        8388608000,
			AvailableBytes:   8388608000,
			UsagePercent:     50.0,
			SwapTotalBytes:   4194304000,
			SwapUsedBytes:    1048576000,
			SwapFreeBytes:    3145728000,
			SwapUsagePercent: 25.0,
		},
	}

	converted := ConvertLegacyToTimescaleMetrics(legacy)

	assert.Equal(t, "test-host", converted.Hostname)
	assert.NotNil(t, converted.Metrics.CPU)
	assert.Equal(t, 45.5, converted.Metrics.CPU.UsagePercent)
	assert.Equal(t, 1.2, converted.Metrics.CPU.LoadAverage.Load1)
	assert.Len(t, converted.Metrics.CPU.PerCPU, 4)
	assert.Equal(t, "0", converted.Metrics.CPU.PerCPU[0].Core)
	assert.Equal(t, 40.1, converted.Metrics.CPU.PerCPU[0].UsagePercent)

	assert.NotNil(t, converted.Metrics.Memory)
	assert.Equal(t, uint64(16777216000), converted.Metrics.Memory.Total)
	assert.Equal(t, 50.0, converted.Metrics.Memory.UsedPercent)

	assert.NotNil(t, converted.Metrics.System)
	assert.Equal(t, "test-host", converted.Metrics.System.Host.Hostname)
	assert.Equal(t, "linux", converted.Metrics.System.Host.OS)
}

// TestSubmitComprehensiveToTimescale tests submitting metrics to TimescaleDB
func TestSubmitComprehensiveToTimescale(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/metrics/comprehensive", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "test-uuid", r.Header.Get("X-Server-UUID"))
		assert.Equal(t, "test-secret", r.Header.Get("X-Server-Secret"))

		var body ComprehensiveMetricsSubmission
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)

		assert.Equal(t, "test-host", body.Hostname)
		assert.NotNil(t, body.Metrics.CPU)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Metrics stored successfully",
		})
	}))
	defer server.Close()

	// Create client
	config := &Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			ServerUUID:   "test-uuid",
			ServerSecret: "test-secret",
		},
	}
	client, err := NewClient(config)
	require.NoError(t, err)

	// Create metrics
	metrics := &ComprehensiveMetricsSubmission{
		Timestamp: time.Now().Unix(),
		Hostname:  "test-host",
		Metrics: &ComprehensiveMetricsPayload{
			CPU: &TimescaleCPUMetrics{
				UsagePercent: 45.5,
			},
		},
	}

	// Submit metrics
	err = client.Metrics.SubmitComprehensiveToTimescale(context.Background(), metrics)
	assert.NoError(t, err)
}

// TestGetLatestMetrics tests retrieving latest metrics
func TestGetLatestMetrics(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/servers/test-uuid/metrics/latest", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.Header.Get("Authorization"), "Bearer")

		metricsData := TimescaleMetricsResponse{
			ServerUUID: "test-uuid",
			Timestamp:  time.Now().Format(time.RFC3339),
			Metrics: &ComprehensiveMetricsTimescale{
				ServerUUID:         "test-uuid",
				Timestamp:          time.Now(),
				CPUUsagePercent:    floatPtr(45.5),
				MemoryUsagePercent: floatPtr(60.2),
			},
			Source: "timescaledb",
		}

		response := StandardResponse{
			Status: "success",
			Data:   metricsData,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with JWT auth
	config := &Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-jwt-token",
		},
	}
	client, err := NewClient(config)
	require.NoError(t, err)

	// Get latest metrics
	result, err := client.Metrics.GetLatestMetrics(context.Background(), "test-uuid")
	require.NoError(t, err)

	assert.Equal(t, "test-uuid", result.ServerUUID)
	assert.Equal(t, "timescaledb", result.Source)
	assert.NotNil(t, result.Metrics)
	assert.Equal(t, 45.5, *result.Metrics.CPUUsagePercent)
	assert.Equal(t, 60.2, *result.Metrics.MemoryUsagePercent)
}

// TestGetMetricsRange tests retrieving metrics range with comprehensive coverage
func TestGetMetricsRange(t *testing.T) {
	tests := []struct {
		name           string
		serverUUID     string
		startTime      string
		endTime        string
		limit          int
		responseStatus int
		responseData   interface{} // Can be TimescaleMetricsRangeResponse or map[string]interface{}
		responseError  bool
		expectError    bool
		expectedCount  int
	}{
		{
			name:           "successful_metrics_retrieval",
			serverUUID:     "test-uuid",
			startTime:      "2023-01-01T00:00:00Z",
			endTime:        "2023-01-01T01:00:00Z",
			limit:          100,
			responseStatus: http.StatusOK,
			responseData: TimescaleMetricsRangeResponse{
				ServerUUID: "test-uuid",
				StartTime:  "2023-01-01T00:00:00Z",
				EndTime:    "2023-01-01T01:00:00Z",
				Count:      2,
				Metrics: []*ComprehensiveMetricsTimescale{
					{ServerUUID: "test-uuid", CPUUsagePercent: floatPtr(45.5)},
					{ServerUUID: "test-uuid", CPUUsagePercent: floatPtr(48.2)},
				},
				Source: "timescaledb",
			},
			expectError:   false,
			expectedCount: 2,
		},
		{
			name:           "zero_limit_parameter",
			serverUUID:     "test-uuid-2",
			startTime:      "2023-01-01T00:00:00Z",
			endTime:        "2023-01-01T02:00:00Z",
			limit:          0, // Test zero limit (should be omitted from query)
			responseStatus: http.StatusOK,
			responseData: TimescaleMetricsRangeResponse{
				ServerUUID: "test-uuid-2",
				Count:      5,
				Metrics:    []*ComprehensiveMetricsTimescale{},
				Source:     "timescaledb",
			},
			expectError:   false,
			expectedCount: 5,
		},
		{
			name:           "map_data_type_fallback",
			serverUUID:     "test-uuid-3",
			startTime:      "2023-01-01T00:00:00Z",
			endTime:        "2023-01-01T01:00:00Z",
			limit:          50,
			responseStatus: http.StatusOK,
			// Return as map[string]interface{} to test fallback conversion path
			responseData: map[string]interface{}{
				"server_uuid": "test-uuid-3",
				"start_time":  "2023-01-01T00:00:00Z",
				"end_time":    "2023-01-01T01:00:00Z",
				"count":       float64(3), // JSON numbers are float64
				"metrics":     []interface{}{},
				"source":      "timescaledb",
			},
			expectError:   false,
			expectedCount: 3,
		},
		{
			name:           "server_error",
			serverUUID:     "error-uuid",
			startTime:      "2023-01-01T00:00:00Z",
			endTime:        "2023-01-01T01:00:00Z",
			limit:          100,
			responseStatus: http.StatusInternalServerError,
			responseError:  true,
			expectError:    true,
		},
		{
			name:           "unauthorized_error",
			serverUUID:     "unauth-uuid",
			startTime:      "2023-01-01T00:00:00Z",
			endTime:        "2023-01-01T01:00:00Z",
			limit:          100,
			responseStatus: http.StatusUnauthorized,
			responseError:  true,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/v2/servers/"+tt.serverUUID+"/metrics/range", r.URL.Path)
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.Header.Get("Authorization"), "Bearer")

				// Check query parameters
				query := r.URL.Query()
				assert.Equal(t, tt.startTime, query.Get("start_time"))
				assert.Equal(t, tt.endTime, query.Get("end_time"))

				// Check limit parameter handling
				if tt.limit > 0 {
					assert.Equal(t, fmt.Sprintf("%d", tt.limit), query.Get("limit"))
				} else {
					assert.Empty(t, query.Get("limit"), "Limit should not be in query when zero")
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.responseStatus)

				if tt.responseError {
					json.NewEncoder(w).Encode(map[string]string{
						"status": "error",
						"error":  "Internal server error",
					})
				} else {
					response := StandardResponse{
						Status: "success",
						Data:   tt.responseData,
					}
					json.NewEncoder(w).Encode(response)
				}
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-jwt-token"},
			})
			require.NoError(t, err)

			result, err := client.Metrics.GetMetricsRange(
				context.Background(),
				tt.serverUUID,
				tt.startTime,
				tt.endTime,
				tt.limit,
			)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result, "Result should not be nil")
				assert.Equal(t, tt.serverUUID, result.ServerUUID)
				assert.Equal(t, tt.expectedCount, result.Count)
			}
		})
	}
}

// TestMetricsAggregator tests the metrics aggregator
func TestMetricsAggregator(t *testing.T) {
	now := time.Now()
	metrics := []ComprehensiveMetricsTimescale{
		{
			Timestamp:          now.Add(-30 * time.Minute),
			CPUUsagePercent:    floatPtr(40.0),
			MemoryUsagePercent: floatPtr(50.0),
		},
		{
			Timestamp:          now.Add(-20 * time.Minute),
			CPUUsagePercent:    floatPtr(45.0),
			MemoryUsagePercent: floatPtr(55.0),
		},
		{
			Timestamp:          now.Add(-10 * time.Minute),
			CPUUsagePercent:    floatPtr(50.0),
			MemoryUsagePercent: floatPtr(60.0),
		},
	}

	aggregator := NewMetricsAggregator(metrics)

	// Test average calculations
	assert.Equal(t, 45.0, aggregator.AverageCPUUsage())
	assert.Equal(t, 55.0, aggregator.AverageMemoryUsage())

	// Test max calculations
	assert.Equal(t, 50.0, aggregator.MaxCPUUsage())
	assert.Equal(t, 60.0, aggregator.MaxMemoryUsage())

	// Test time range
	start, end := aggregator.TimeRange()
	assert.Equal(t, metrics[0].Timestamp, start)
	assert.Equal(t, metrics[2].Timestamp, end)
}

// Helper function to create float64 pointer
func floatPtr(f float64) *float64 {
	return &f
}
