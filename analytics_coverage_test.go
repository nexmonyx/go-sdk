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

func TestAnalyticsService_GetCapabilities(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v2/analytics/ai/capabilities", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"anomaly_detection":    true,
				"predictive_analytics": true,
				"root_cause_analysis":  false,
				"capacity_planning":    true,
				"available_models":     []string{"model-v1", "model-v2"},
				"status":               "operational",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	capabilities, err := client.Analytics.GetCapabilities(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	assert.NotNil(t, capabilities, "capabilities should not be nil")
	assert.True(t, capabilities.AnomalyDetection)
	assert.Equal(t, "operational", capabilities.Status)
}

func TestAnalyticsService_AnalyzeMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v2/analytics/ai/analyze", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"analysis_id": "analysis-123",
				"timestamp":   "2024-01-01T00:00:00Z",
				"insights": []map[string]interface{}{
					{
						"type":        "warning",
						"title":       "CPU Spike Detected",
						"description": "Unusual CPU usage pattern detected",
						"severity":    "high",
						"confidence":  0.95,
					},
				},
				"confidence": 0.95,
			},
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
	assert.Equal(t, 0.95, result.Confidence)
	assert.Len(t, result.Insights, 1)
}

func TestAnalyticsService_GetServiceStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v2/analytics/ai/status", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"status":            "operational",
				"last_check":        "2024-01-01T00:00:00Z",
				"models_available":  5,
				"queue_length":      10,
				"average_latency_ms": 150.5,
				"uptime_percentage": 99.9,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	status, err := client.Analytics.GetServiceStatus(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, "operational", status.Status)
	assert.Equal(t, 99.9, status.Uptime)
}

func TestAnalyticsService_GetHardwareTrends(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/v2/analytics/hardware/trends/")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"server_uuid": "server-123",
				"start_time":  "2024-01-01T00:00:00Z",
				"end_time":    "2024-01-07T00:00:00Z",
				"cpu_trend": map[string]interface{}{
					"average":    50.5,
					"minimum":    20.0,
					"maximum":    95.0,
					"trend_line": []map[string]interface{}{
						{"timestamp": "2024-01-01T00:00:00Z", "value": 50.0},
					},
					"growth_percentage": 5.5,
					"volatility":        2.3,
				},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	trends, err := client.Analytics.GetHardwareTrends(context.Background(), "server-123", "2024-01-01T00:00:00Z", "2024-01-07T00:00:00Z", "cpu")
	assert.NoError(t, err)
	assert.NotNil(t, trends)
	assert.Equal(t, "server-123", trends.ServerUUID)
}

func TestAnalyticsService_GetHardwareHealth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/v2/analytics/hardware/health/")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"server_uuid":   "server-123",
				"overall_score": 95.0,
				"status":        "healthy",
				"last_check":    "2024-01-01T00:00:00Z",
				"component_scores": map[string]interface{}{
					"cpu":     98.0,
					"memory":  95.0,
					"disk":    92.0,
					"network": 97.0,
				},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	health, err := client.Analytics.GetHardwareHealth(context.Background(), "server-123")
	assert.NoError(t, err)
	assert.NotNil(t, health)
	assert.Equal(t, "server-123", health.ServerUUID)
}

func TestAnalyticsService_GetHardwarePredictions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/v2/analytics/hardware/predictions/")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"server_uuid":          "server-123",
				"prediction_horizon":   30,
				"failure_probability":  0.05,
				"confidence":           0.92,
				"predicted_failures":   []map[string]interface{}{},
				"recommended_actions":  []string{"Monitor disk usage", "Schedule maintenance"},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	prediction, err := client.Analytics.GetHardwarePredictions(context.Background(), "server-123", 30)
	assert.NoError(t, err)
	assert.NotNil(t, prediction)
	assert.Equal(t, "server-123", prediction.ServerUUID)
	assert.Equal(t, 0.05, prediction.FailureProbability)
}

func TestAnalyticsService_GetFleetOverview(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v2/analytics/fleet/overview", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"total_servers":   100,
				"active_servers":  95,
				"inactive_servers": 5,
				"health_distribution": map[string]interface{}{
					"healthy":  90,
					"warning":  8,
					"critical": 2,
					"unknown":  0,
				},
				"resource_utilization": map[string]interface{}{
					"average_cpu":    55.5,
					"average_memory": 65.2,
					"average_disk":   45.8,
					"total_storage_gb": 10000.0,
				},
				"last_updated": "2024-01-01T00:00:00Z",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	overview, err := client.Analytics.GetFleetOverview(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, overview)
	assert.Equal(t, 100, overview.TotalServers)
	assert.Equal(t, 95, overview.ActiveServers)
}

func TestAnalyticsService_GetOrganizationDashboard(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v2/analytics/fleet/dashboard", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"fleet_overview": map[string]interface{}{
					"total_servers":   100,
					"active_servers":  95,
					"inactive_servers": 5,
					"health_distribution": map[string]interface{}{
						"healthy":  90,
						"warning":  8,
						"critical": 2,
						"unknown":  0,
					},
					"resource_utilization": map[string]interface{}{
						"average_cpu":    55.5,
						"average_memory": 65.2,
						"average_disk":   45.8,
						"total_storage_gb": 10000.0,
					},
					"last_updated": "2024-01-01T00:00:00Z",
				},
				"recent_alerts": []map[string]interface{}{
					{
						"alert_id":  1,
						"severity":  "critical",
						"title":     "High CPU Usage",
						"timestamp": "2024-01-01T00:00:00Z",
					},
				},
				"trending_metrics": []map[string]interface{}{},
				"last_updated": "2024-01-01T00:00:00Z",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	dashboard, err := client.Analytics.GetOrganizationDashboard(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, dashboard)
	assert.Equal(t, 100, dashboard.FleetOverview.TotalServers)
	assert.Len(t, dashboard.RecentAlerts, 1)
}

func TestAnalyticsService_AnalyzeCorrelations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v2/analytics/correlation/analyze", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"correlations": []map[string]interface{}{
					{
						"metric1":     "cpu",
						"metric2":     "memory",
						"coefficient": 0.85,
						"p_value":     0.001,
					},
				},
				"metric_labels": []string{"cpu", "memory"},
				"analyzed_at":   "2024-01-01T00:00:00Z",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	req := &CorrelationAnalysisRequest{ServerUUIDs: []string{"server-1", "server-2"}}
	result, err := client.Analytics.AnalyzeCorrelations(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Correlations, 1)
}

func TestAnalyticsService_BuildDependencyGraph(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v2/analytics/graph/dependencies", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"nodes": []map[string]interface{}{
					{"id": "node1", "type": "server", "name": "Web Server 1", "status": "healthy"},
					{"id": "node2", "type": "server", "name": "DB Server 1", "status": "healthy"},
				},
				"edges": []map[string]interface{}{
					{"from": "node1", "to": "node2", "type": "depends_on", "weight": 1},
				},
				"generated_at": "2024-01-01T00:00:00Z",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	graph, err := client.Analytics.BuildDependencyGraph(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, graph)
	assert.Len(t, graph.Nodes, 2)
	assert.Len(t, graph.Edges, 1)
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

	// Disable retries to prevent timeout
	client, _ := NewClient(&Config{BaseURL: server.URL, RetryCount: 0})

	// Use short timeout context to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := client.Analytics.GetCapabilities(ctx)
	assert.Error(t, err)

	_, err = client.Analytics.AnalyzeMetrics(ctx, &AIAnalysisRequest{})
	assert.Error(t, err)

	_, err = client.Analytics.GetServiceStatus(ctx)
	assert.Error(t, err)

	_, err = client.Analytics.GetHardwareTrends(ctx, "server-123", "2024-01-01", "2024-01-07")
	assert.Error(t, err)

	_, err = client.Analytics.GetHardwareHealth(ctx, "server-123")
	assert.Error(t, err)

	_, err = client.Analytics.GetHardwarePredictions(ctx, "server-123", 30)
	assert.Error(t, err)

	_, err = client.Analytics.GetFleetOverview(ctx)
	assert.Error(t, err)

	_, err = client.Analytics.GetOrganizationDashboard(ctx)
	assert.Error(t, err)

	_, err = client.Analytics.AnalyzeCorrelations(ctx, &CorrelationAnalysisRequest{})
	assert.Error(t, err)

	_, err = client.Analytics.BuildDependencyGraph(ctx)
	assert.Error(t, err)
}
