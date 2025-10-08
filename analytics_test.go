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

// AI Analytics Tests

func TestAnalyticsService_GetCapabilities(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v2/analytics/ai/capabilities", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		response := StandardResponse{
			Status:  "success",
			Message: "AI capabilities retrieved successfully",
			Data: &AICapabilities{
				AnomalyDetection:    true,
				PredictiveAnalytics: true,
				RootCauseAnalysis:   true,
				CapacityPlanning:    false,
				AvailableModels:     []string{"anomaly-v1", "prediction-v2"},
				Status:              "operational",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	capabilities, err := client.Analytics.GetCapabilities(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, capabilities)
	assert.True(t, capabilities.AnomalyDetection)
	assert.True(t, capabilities.PredictiveAnalytics)
	assert.Equal(t, "operational", capabilities.Status)
	assert.Len(t, capabilities.AvailableModels, 2)
}

func TestAnalyticsService_AnalyzeMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v2/analytics/ai/analyze", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		var req AIAnalysisRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "anomaly", req.AnalysisType)

		response := StandardResponse{
			Status:  "success",
			Message: "Analysis completed successfully",
			Data: &AIAnalysisResult{
				AnalysisID: "analysis-123",
				Timestamp:  CustomTime{Time: time.Now()},
				Insights: []AIInsight{
					{
						Type:        "warning",
						Title:       "High CPU Usage Detected",
						Description: "CPU usage consistently above 90%",
						Severity:    "high",
						Confidence:  0.95,
					},
				},
				Confidence: 0.92,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	req := &AIAnalysisRequest{
		ServerUUIDs:  []string{"server-1"},
		MetricTypes:  []string{"cpu", "memory"},
		AnalysisType: "anomaly",
		TimeRange: TimeRange{
			Start: time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			End:   time.Now().Format(time.RFC3339),
		},
	}

	result, err := client.Analytics.AnalyzeMetrics(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "analysis-123", result.AnalysisID)
	assert.Len(t, result.Insights, 1)
	assert.Equal(t, "warning", result.Insights[0].Type)
}

func TestAnalyticsService_GetServiceStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v2/analytics/ai/status", r.URL.Path)

		response := StandardResponse{
			Status:  "success",
			Message: "AI service status retrieved",
			Data: &AIServiceStatus{
				Status:          "operational",
				LastCheck:       CustomTime{Time: time.Now()},
				ModelsAvailable: 5,
				QueueLength:     2,
				AverageLatency:  125.5,
				Uptime:          99.9,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	status, err := client.Analytics.GetServiceStatus(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, "operational", status.Status)
	assert.Equal(t, 5, status.ModelsAvailable)
}

// Hardware Analytics Tests

func TestAnalyticsService_GetHardwareTrends(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v2/analytics/hardware/trends/server-uuid-123", r.URL.Path)
		assert.Equal(t, "2024-01-01T00:00:00Z", r.URL.Query().Get("start_time"))
		assert.Equal(t, "2024-01-02T00:00:00Z", r.URL.Query().Get("end_time"))

		response := StandardResponse{
			Status:  "success",
			Message: "Hardware trends retrieved",
			Data: &HardwareTrends{
				ServerUUID: "server-uuid-123",
				StartTime:  CustomTime{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
				EndTime:    CustomTime{Time: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)},
				CPUTrend: MetricTrendData{
					Average:    45.5,
					Minimum:    20.0,
					Maximum:    85.0,
					Growth:     5.2,
					Volatility: 12.3,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	trends, err := client.Analytics.GetHardwareTrends(
		context.Background(),
		"server-uuid-123",
		"2024-01-01T00:00:00Z",
		"2024-01-02T00:00:00Z",
	)
	require.NoError(t, err)
	assert.NotNil(t, trends)
	assert.Equal(t, "server-uuid-123", trends.ServerUUID)
	assert.Equal(t, 45.5, trends.CPUTrend.Average)
}

func TestAnalyticsService_GetHardwareHealth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v2/analytics/hardware/health/server-uuid-456", r.URL.Path)

		response := StandardResponse{
			Status:  "success",
			Message: "Hardware health retrieved",
			Data: &HardwareHealth{
				ServerUUID:   "server-uuid-456",
				OverallScore: 85,
				LastCheck:    CustomTime{Time: time.Now()},
				ComponentScores: ComponentHealthMap{
					CPU:     90,
					Memory:  85,
					Disk:    80,
					Network: 95,
				},
				Issues: []HealthIssue{
					{
						Component:   "disk",
						Severity:    "warning",
						Description: "Disk usage at 75%",
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	health, err := client.Analytics.GetHardwareHealth(context.Background(), "server-uuid-456")
	require.NoError(t, err)
	assert.NotNil(t, health)
	assert.Equal(t, 85, health.OverallScore)
	assert.Equal(t, 90, health.ComponentScores.CPU)
	assert.Len(t, health.Issues, 1)
}

func TestAnalyticsService_GetHardwarePredictions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v2/analytics/hardware/predictions/server-uuid-789", r.URL.Path)
		assert.Equal(t, "60", r.URL.Query().Get("horizon"))

		response := StandardResponse{
			Status:  "success",
			Message: "Hardware predictions retrieved",
			Data: &HardwarePrediction{
				ServerUUID:         "server-uuid-789",
				PredictionHorizon:  60,
				FailureProbability: 0.15,
				ComponentPredictions: []ComponentPrediction{
					{
						Component:          "disk",
						FailureProbability: 0.25,
						WarningLevel:       "medium",
					},
				},
				ConfidenceLevel: 0.88,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	predictions, err := client.Analytics.GetHardwarePredictions(context.Background(), "server-uuid-789", 60)
	require.NoError(t, err)
	assert.NotNil(t, predictions)
	assert.Equal(t, 60, predictions.PredictionHorizon)
	assert.Equal(t, 0.15, predictions.FailureProbability)
	assert.Len(t, predictions.ComponentPredictions, 1)
}

// Fleet Analytics Tests

func TestAnalyticsService_GetFleetOverview(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v2/analytics/fleet/overview", r.URL.Path)

		response := StandardResponse{
			Status:  "success",
			Message: "Fleet overview retrieved",
			Data: &FleetOverview{
				TotalServers:    100,
				ActiveServers:   95,
				InactiveServers: 5,
				HealthDistribution: HealthDistribution{
					Healthy:  70,
					Warning:  20,
					Critical: 5,
					Unknown:  5,
				},
				ResourceUtilization: ResourceUtilization{
					AverageCPU:    55.5,
					AverageMemory: 65.2,
					AverageDisk:   45.0,
					TotalStorage:  5000.0,
				},
				LastUpdated: CustomTime{Time: time.Now()},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	overview, err := client.Analytics.GetFleetOverview(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, overview)
	assert.Equal(t, 100, overview.TotalServers)
	assert.Equal(t, 95, overview.ActiveServers)
	assert.Equal(t, 70, overview.HealthDistribution.Healthy)
}

func TestAnalyticsService_GetOrganizationDashboard(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v2/analytics/fleet/dashboard", r.URL.Path)

		response := StandardResponse{
			Status:  "success",
			Message: "Dashboard data retrieved",
			Data: &OrganizationDashboard{
				FleetOverview: FleetOverview{
					TotalServers:  50,
					ActiveServers: 48,
				},
				RecentAlerts: []DashboardAlert{
					{
						AlertID:  1,
						Severity: "high",
						Title:    "High CPU Usage",
						Timestamp: CustomTime{Time: time.Now()},
					},
				},
				TrendingMetrics: []TrendingMetric{
					{
						MetricType: "cpu",
						Value:      65.5,
						Change:     5.2,
						Trend:      "up",
					},
				},
				LastUpdated: CustomTime{Time: time.Now()},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	dashboard, err := client.Analytics.GetOrganizationDashboard(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, dashboard)
	assert.Equal(t, 50, dashboard.FleetOverview.TotalServers)
	assert.Len(t, dashboard.RecentAlerts, 1)
	assert.Len(t, dashboard.TrendingMetrics, 1)
}

// Advanced Analytics Tests

func TestAnalyticsService_AnalyzeCorrelations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v2/analytics/correlation/analyze", r.URL.Path)

		var req CorrelationAnalysisRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Contains(t, req.MetricTypes, "cpu")

		response := StandardResponse{
			Status:  "success",
			Message: "Correlation analysis completed",
			Data: &CorrelationResult{
				Correlations: []MetricCorrelation{
					{
						Metric1:     "cpu",
						Metric2:     "memory",
						Coefficient: 0.75,
						Strength:    "strong",
					},
				},
				MetricLabels: []string{"cpu", "memory", "disk"},
				AnalyzedAt:   CustomTime{Time: time.Now()},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	req := &CorrelationAnalysisRequest{
		MetricTypes: []string{"cpu", "memory", "disk"},
		TimeRange: TimeRange{
			Start: time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			End:   time.Now().Format(time.RFC3339),
		},
		Method: "pearson",
	}

	result, err := client.Analytics.AnalyzeCorrelations(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Correlations, 1)
	assert.Equal(t, 0.75, result.Correlations[0].Coefficient)
}

func TestAnalyticsService_BuildDependencyGraph(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v2/analytics/graph/dependencies", r.URL.Path)

		response := StandardResponse{
			Status:  "success",
			Message: "Dependency graph generated",
			Data: &DependencyGraph{
				Nodes: []DependencyNode{
					{
						ID:     "server-1",
						Type:   "server",
						Name:   "web-server-1",
						Status: "healthy",
					},
					{
						ID:     "db-1",
						Type:   "database",
						Name:   "postgres-primary",
						Status: "healthy",
					},
				},
				Edges: []DependencyEdge{
					{
						From: "server-1",
						To:   "db-1",
						Type: "depends_on",
					},
				},
				GeneratedAt: CustomTime{Time: time.Now()},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	graph, err := client.Analytics.BuildDependencyGraph(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, graph)
	assert.Len(t, graph.Nodes, 2)
	assert.Len(t, graph.Edges, 1)
	assert.Equal(t, "server", graph.Nodes[0].Type)
}

// Error Handling Tests

func TestAnalyticsService_ErrorHandling(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		responseBody string
		shouldError  bool
	}{
		{
			name:         "Unauthorized",
			statusCode:   401,
			responseBody: `{"status":"error","error":"unauthorized","message":"Authentication required"}`,
			shouldError:  true,
		},
		{
			name:         "Forbidden",
			statusCode:   403,
			responseBody: `{"status":"error","error":"forbidden","message":"Insufficient permissions"}`,
			shouldError:  true,
		},
		{
			name:         "Not Found",
			statusCode:   404,
			responseBody: `{"status":"error","error":"not_found","message":"Resource not found"}`,
			shouldError:  true,
		},
		{
			name:         "Internal Server Error",
			statusCode:   500,
			responseBody: `{"status":"error","error":"internal_server_error","message":"Service error"}`,
			shouldError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			// Test various methods with error
			_, err = client.Analytics.GetCapabilities(context.Background())
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
