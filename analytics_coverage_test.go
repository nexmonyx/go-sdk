package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnalyticsService_GetCapabilities(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v2/analytics/ai/capabilities", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   map[string]interface{}{"features": []string{"anomaly_detection", "prediction"}},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	capabilities, err := client.Analytics.GetCapabilities(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, capabilities)
}

func TestAnalyticsService_AnalyzeMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v2/analytics/ai/analyze", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   map[string]interface{}{"insights": []string{"spike_detected"}, "confidence": 0.95},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	req := &AIAnalysisRequest{
		ServerUUIDs:  []string{"server-123"},
		MetricTypes:  []string{"cpu"},
		AnalysisType: "anomaly",
	}
	result, err := client.Analytics.AnalyzeMetrics(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestAnalyticsService_GetServiceStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v2/analytics/ai/status", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   map[string]interface{}{"status": "healthy", "uptime": 99.9},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	status, err := client.Analytics.GetServiceStatus(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, status)
}

func TestAnalyticsService_GetHardwareTrends(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/v2/analytics/hardware/trends/")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   map[string]interface{}{"trends": []map[string]interface{}{{"date": "2024-01-01", "value": 50}}},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	trends, err := client.Analytics.GetHardwareTrends(context.Background(), "server-123", "2024-01-01T00:00:00Z", "2024-01-07T00:00:00Z", "cpu")
	assert.NoError(t, err)
	assert.NotNil(t, trends)
}

func TestAnalyticsService_GetHardwareHealth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/v2/analytics/hardware/health/")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   map[string]interface{}{"health_score": 95, "status": "healthy"},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	health, err := client.Analytics.GetHardwareHealth(context.Background(), "server-123")
	assert.NoError(t, err)
	assert.NotNil(t, health)
}

func TestAnalyticsService_GetHardwarePredictions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/v2/analytics/hardware/predictions/")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   map[string]interface{}{"failure_probability": 0.05, "predicted_failure_date": "2024-12-31"},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	prediction, err := client.Analytics.GetHardwarePredictions(context.Background(), "server-123", 30)
	assert.NoError(t, err)
	assert.NotNil(t, prediction)
}

func TestAnalyticsService_GetFleetOverview(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v2/analytics/fleet/overview", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   map[string]interface{}{"total_servers": 100, "healthy": 95, "degraded": 5},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	overview, err := client.Analytics.GetFleetOverview(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, overview)
}

func TestAnalyticsService_GetOrganizationDashboard(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v2/analytics/fleet/dashboard", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   map[string]interface{}{"servers": 100, "alerts": 5, "uptime": 99.9},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	dashboard, err := client.Analytics.GetOrganizationDashboard(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, dashboard)
}

func TestAnalyticsService_AnalyzeCorrelations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v2/analytics/correlation/analyze", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   map[string]interface{}{"correlations": []map[string]interface{}{{"metric1": "cpu", "metric2": "memory", "coefficient": 0.85}}},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	req := &CorrelationAnalysisRequest{ServerUUIDs: []string{"server-1", "server-2"}}
	result, err := client.Analytics.AnalyzeCorrelations(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestAnalyticsService_BuildDependencyGraph(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v2/analytics/graph/dependencies", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   map[string]interface{}{"nodes": []string{"node1", "node2"}, "edges": []map[string]string{{"from": "node1", "to": "node2"}}},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	graph, err := client.Analytics.BuildDependencyGraph(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, graph)
}

func TestAnalyticsService_Errors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Internal error",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	_, err := client.Analytics.GetCapabilities(context.Background())
	assert.Error(t, err)

	_, err = client.Analytics.AnalyzeMetrics(context.Background(), &AIAnalysisRequest{})
	assert.Error(t, err)

	_, err = client.Analytics.GetServiceStatus(context.Background())
	assert.Error(t, err)

	_, err = client.Analytics.GetHardwareTrends(context.Background(), "server-123", "2024-01-01", "2024-01-07")
	assert.Error(t, err)

	_, err = client.Analytics.GetHardwareHealth(context.Background(), "server-123")
	assert.Error(t, err)

	_, err = client.Analytics.GetHardwarePredictions(context.Background(), "server-123", 30)
	assert.Error(t, err)

	_, err = client.Analytics.GetFleetOverview(context.Background())
	assert.Error(t, err)

	_, err = client.Analytics.GetOrganizationDashboard(context.Background())
	assert.Error(t, err)

	_, err = client.Analytics.AnalyzeCorrelations(context.Background(), &CorrelationAnalysisRequest{})
	assert.Error(t, err)

	_, err = client.Analytics.BuildDependencyGraph(context.Background())
	assert.Error(t, err)
}
