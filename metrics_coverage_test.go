package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Coverage tests for metrics.go
// Most methods already at 100%, focusing on GetStatus (87.5%) and builder methods (0%)

// GetStatus tests (87.5% - type assertion line)

func TestMetricsService_GetStatus_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/metrics/server-uuid-123/status", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"server_uuid":         "server-uuid-123",
				"collection_enabled":  true,
				"collection_interval": 60,
				"error_count":         0,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	status, err := client.Metrics.GetStatus(context.Background(), "server-uuid-123")
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, "server-uuid-123", status.ServerUUID)
	assert.True(t, status.CollectionEnabled)
}

func TestMetricsService_GetStatus_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Server not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	status, err := client.Metrics.GetStatus(context.Background(), "nonexistent")
	assert.Error(t, err)
	assert.Nil(t, status)
}

// MetricsBuilder tests (0% - all builder methods)

func TestNewMetricsBuilder(t *testing.T) {
	builder := NewMetricsBuilder("test-host")
	assert.NotNil(t, builder)
	assert.NotNil(t, builder.metrics)
	assert.Equal(t, "test-host", builder.metrics.Hostname)
	assert.NotNil(t, builder.metrics.CollectedAt)
	assert.NotNil(t, builder.metrics.Metrics)
}

func TestMetricsBuilder_WithAgentVersion(t *testing.T) {
	builder := NewMetricsBuilder("test-host")
	result := builder.WithAgentVersion("v1.2.3")

	assert.Equal(t, builder, result) // Fluent interface returns self
	assert.Equal(t, "v1.2.3", builder.metrics.AgentVersion)
}

func TestMetricsBuilder_WithCollectionDuration(t *testing.T) {
	builder := NewMetricsBuilder("test-host")
	result := builder.WithCollectionDuration(1.25)

	assert.Equal(t, builder, result)
	assert.Equal(t, 1.25, builder.metrics.CollectionDuration)
}

func TestMetricsBuilder_WithErrorCount(t *testing.T) {
	builder := NewMetricsBuilder("test-host")
	result := builder.WithErrorCount(3)

	assert.Equal(t, builder, result)
	assert.Equal(t, 3, builder.metrics.ErrorCount)
}

func TestMetricsBuilder_WithCPUMetrics(t *testing.T) {
	builder := NewMetricsBuilder("test-host")
	cpuMetrics := &TimescaleCPUMetrics{
		UsagePercent:  75.5,
		UserPercent:   50.0,
		SystemPercent: 25.5,
		IdlePercent:   24.5,
	}

	result := builder.WithCPUMetrics(cpuMetrics)

	assert.Equal(t, builder, result)
	assert.Equal(t, cpuMetrics, builder.metrics.Metrics.CPU)
	assert.Equal(t, 75.5, builder.metrics.Metrics.CPU.UsagePercent)
}

func TestMetricsBuilder_WithMemoryMetrics(t *testing.T) {
	builder := NewMetricsBuilder("test-host")
	memoryMetrics := &TimescaleMemoryMetrics{
		Total:       16777216000,
		Available:   8388608000,
		Used:        8388608000,
		UsedPercent: 50.0,
	}

	result := builder.WithMemoryMetrics(memoryMetrics)

	assert.Equal(t, builder, result)
	assert.Equal(t, memoryMetrics, builder.metrics.Metrics.Memory)
	assert.Equal(t, 50.0, builder.metrics.Metrics.Memory.UsedPercent)
}

func TestMetricsBuilder_Build(t *testing.T) {
	builder := NewMetricsBuilder("test-host")
	metrics := builder.Build()

	assert.NotNil(t, metrics)
	assert.Equal(t, "test-host", metrics.Hostname)
	assert.NotNil(t, metrics.CollectedAt)
}

func TestMetricsBuilder_FluentChaining(t *testing.T) {
	metrics := NewMetricsBuilder("test-host").
		WithAgentVersion("v2.0.0").
		WithCollectionDuration(2.5).
		WithErrorCount(0).
		WithCPUMetrics(&TimescaleCPUMetrics{
			UsagePercent: 80.0,
			UserPercent:  60.0,
		}).
		WithMemoryMetrics(&TimescaleMemoryMetrics{
			Total:       32000000000,
			Used:        16000000000,
			UsedPercent: 50.0,
		}).
		Build()

	assert.NotNil(t, metrics)
	assert.Equal(t, "test-host", metrics.Hostname)
	assert.Equal(t, "v2.0.0", metrics.AgentVersion)
	assert.Equal(t, 2.5, metrics.CollectionDuration)
	assert.Equal(t, 0, metrics.ErrorCount)
	assert.NotNil(t, metrics.Metrics.CPU)
	assert.Equal(t, 80.0, metrics.Metrics.CPU.UsagePercent)
	assert.NotNil(t, metrics.Metrics.Memory)
	assert.Equal(t, 50.0, metrics.Metrics.Memory.UsedPercent)
}

func TestMetricsBuilder_BuildMinimal(t *testing.T) {
	metrics := NewMetricsBuilder("minimal-host").Build()

	assert.NotNil(t, metrics)
	assert.Equal(t, "minimal-host", metrics.Hostname)
	assert.Empty(t, metrics.AgentVersion)
	assert.Equal(t, 0.0, metrics.CollectionDuration)
	assert.Equal(t, 0, metrics.ErrorCount)
	assert.NotNil(t, metrics.Metrics)
}
