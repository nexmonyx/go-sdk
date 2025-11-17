package nexmonyx

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==========================================
// Comprehensive Monitoring Service Tests
// ==========================================

func TestMonitoringService_CreateProbe_Comprehensive(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Contains(t, r.URL.Path, "/v1/monitoring/probes")

			probe := &MonitoringProbe{
				ProbeUUID: "probe-123",
				Name:      "Test HTTP Probe",
				Type:      "http",
				Target:    "https://example.com",
				Interval:  60,
				Timeout:   30,
				Enabled:   true,
			}

			response := struct {
				Data *MonitoringProbe `json:"data"`
			}{Data: probe}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client, _ := NewClient(&Config{
			BaseURL: server.URL,
			Auth:    AuthConfig{APIKey: "test-key", APISecret: "test-secret"},
		})

		result, err := client.Monitoring.CreateProbe(context.Background(), &MonitoringProbe{
			Name: "Test HTTP Probe",
		})

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "probe-123", result.ProbeUUID)
		assert.Equal(t, "http", result.Type)
	})

	t.Run("Error - Bad Request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid probe"})
		}))
		defer server.Close()

		client, _ := NewClient(&Config{
			BaseURL: server.URL,
			Auth:    AuthConfig{APIKey: "test-key", APISecret: "test-secret"},
		})

		_, err := client.Monitoring.CreateProbe(context.Background(), &MonitoringProbe{})
		assert.Error(t, err)
	})
}

func TestMonitoringService_GetProbe_Comprehensive(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.URL.Path, "/v1/monitoring/probes/probe-123")

			probe := &MonitoringProbe{
				ProbeUUID:      "probe-123",
				Name:           "Test Probe",
				Type:           "https",
				Target:         "https://api.example.com",
				Interval:       300,
				OrganizationID: 1,
			}

			response := struct {
				Data *MonitoringProbe `json:"data"`
			}{Data: probe}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client, _ := NewClient(&Config{
			BaseURL: server.URL,
			Auth:    AuthConfig{APIKey: "test-key", APISecret: "test-secret"},
		})

		result, err := client.Monitoring.GetProbe(context.Background(), "probe-123")

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "probe-123", result.ProbeUUID)
		assert.Equal(t, "https", result.Type)
	})

	t.Run("Error - Not Found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "probe not found"})
		}))
		defer server.Close()

		client, _ := NewClient(&Config{
			BaseURL: server.URL,
			Auth:    AuthConfig{APIKey: "test-key", APISecret: "test-secret"},
		})

		_, err := client.Monitoring.GetProbe(context.Background(), "non-existent")
		assert.Error(t, err)
	})
}

func TestMonitoringService_ListProbes_Comprehensive(t *testing.T) {
	t.Run("Success - Multiple Probes", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.URL.Path, "/v1/monitoring/probes")

			probes := []*MonitoringProbe{
				{ProbeUUID: "probe-1", Name: "Probe 1", Type: "http"},
				{ProbeUUID: "probe-2", Name: "Probe 2", Type: "tcp"},
				{ProbeUUID: "probe-3", Name: "Probe 3", Type: "icmp"},
			}

			response := struct {
				Data []*MonitoringProbe `json:"data"`
				Meta *PaginationMeta    `json:"meta"`
			}{
				Data: probes,
				Meta: &PaginationMeta{TotalItems: 3, TotalPages: 1, Page: 1, Limit: 10},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client, _ := NewClient(&Config{
			BaseURL: server.URL,
			Auth:    AuthConfig{APIKey: "test-key", APISecret: "test-secret"},
		})

		probes, meta, err := client.Monitoring.ListProbes(context.Background(), &ListOptions{Page: 1, Limit: 10})

		require.NoError(t, err)
		assert.Len(t, probes, 3)
		assert.NotNil(t, meta)
		assert.Equal(t, 3, meta.TotalItems)
	})

	t.Run("Success - Empty List", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := struct {
				Data []*MonitoringProbe `json:"data"`
				Meta *PaginationMeta    `json:"meta"`
			}{
				Data: []*MonitoringProbe{},
				Meta: &PaginationMeta{TotalItems: 0, TotalPages: 0, Page: 1, Limit: 10},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client, _ := NewClient(&Config{
			BaseURL: server.URL,
			Auth:    AuthConfig{APIKey: "test-key", APISecret: "test-secret"},
		})

		probes, meta, err := client.Monitoring.ListProbes(context.Background(), nil)

		require.NoError(t, err)
		assert.Len(t, probes, 0)
		assert.NotNil(t, meta)
	})
}

func TestMonitoringService_UpdateProbe_Comprehensive(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "PUT", r.Method)
			assert.Contains(t, r.URL.Path, "/v1/monitoring/probes/probe-123")

			probe := &MonitoringProbe{
				ProbeUUID: "probe-123",
				Name:      "Updated Probe",
				Interval:  600,
				Enabled:   false,
			}

			response := struct {
				Data *MonitoringProbe `json:"data"`
			}{Data: probe}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client, _ := NewClient(&Config{
			BaseURL: server.URL,
			Auth:    AuthConfig{APIKey: "test-key", APISecret: "test-secret"},
		})

		result, err := client.Monitoring.UpdateProbe(context.Background(), "probe-123", &MonitoringProbe{
			Name:     "Updated Probe",
			Interval: 600,
			Enabled:  false,
		})

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "probe-123", result.ProbeUUID)
		assert.Equal(t, "Updated Probe", result.Name)
	})
}

func TestMonitoringService_DeleteProbe_Comprehensive(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "DELETE", r.Method)
			assert.Contains(t, r.URL.Path, "/v1/monitoring/probes/probe-123")

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"message": "deleted"})
		}))
		defer server.Close()

		client, _ := NewClient(&Config{
			BaseURL: server.URL,
			Auth:    AuthConfig{APIKey: "test-key", APISecret: "test-secret"},
		})

		err := client.Monitoring.DeleteProbe(context.Background(), "probe-123")
		require.NoError(t, err)
	})
}

func TestMonitoringService_GetProbeResults_Comprehensive(t *testing.T) {
	now := time.Now()
	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.URL.Path, "/v1/monitoring/probes/probe-123/results")

			results := []*ProbeTestResult{
				{ProbeUUID: "probe-123", Status: "up", ResponseTime: 100, ExecutedAt: &CustomTime{Time: now}},
				{ProbeUUID: "probe-123", Status: "up", ResponseTime: 120, ExecutedAt: &CustomTime{Time: now.Add(-1 * time.Minute)}},
			}

			response := struct {
				Data []*ProbeTestResult `json:"data"`
				Meta *PaginationMeta    `json:"meta"`
			}{
				Data: results,
				Meta: &PaginationMeta{TotalItems: 2, TotalPages: 1, Page: 1, Limit: 10},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client, _ := NewClient(&Config{
			BaseURL: server.URL,
			Auth:    AuthConfig{APIKey: "test-key", APISecret: "test-secret"},
		})

		results, meta, err := client.Monitoring.GetProbeResults(context.Background(), "probe-123", nil)

		require.NoError(t, err)
		assert.Len(t, results, 2)
		assert.NotNil(t, meta)
	})
}

func TestMonitoringService_GetAgents_Comprehensive(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.URL.Path, "/v1/monitoring/agents")

			agents := []*MonitoringAgent{
				{UUID: "agent-1", Name: "Agent 1", Status: "active"},
				{UUID: "agent-2", Name: "Agent 2", Status: "inactive"},
			}

			response := struct {
				Data []*MonitoringAgent `json:"data"`
				Meta *PaginationMeta    `json:"meta"`
			}{
				Data: agents,
				Meta: &PaginationMeta{TotalItems: 2, TotalPages: 1, Page: 1, Limit: 20},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client, _ := NewClient(&Config{
			BaseURL: server.URL,
			Auth:    AuthConfig{APIKey: "test-key", APISecret: "test-secret"},
		})

		agents, meta, err := client.Monitoring.GetAgents(context.Background(), &ListOptions{Page: 1, Limit: 20})

		require.NoError(t, err)
		assert.Len(t, agents, 2)
		assert.NotNil(t, meta)
	})
}

func TestMonitoringService_GetStatus_Comprehensive(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.URL.Path, "/v1/monitoring/organizations/org-123/status")

			status := &MonitoringStatus{
				ActiveProbes:  10,
				TotalProbes:   15,
				ActiveAgents:  3,
				TotalAgents:   5,
				HealthyProbes: 8,
				FailingProbes: 2,
				ProbesByType: map[string]int{
					"http":  6,
					"https": 4,
				},
			}

			response := struct {
				Data *MonitoringStatus `json:"data"`
			}{Data: status}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client, _ := NewClient(&Config{
			BaseURL: server.URL,
			Auth:    AuthConfig{APIKey: "test-key", APISecret: "test-secret"},
		})

		result, err := client.Monitoring.GetStatus(context.Background(), "org-123")

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 10, result.ActiveProbes)
		assert.Equal(t, 15, result.TotalProbes)
	})
}

func TestMonitoringService_GetAgentStatus_Comprehensive(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.URL.Path, "/v1/monitoring/agents/agent-123/status")

			status := &AgentStatusResponse{
				AgentID:       "agent-123",
				Status:        "healthy",
				LastHeartbeat: &CustomTime{Time: time.Now()},
				Uptime:        86400.5,
				ProbesRunning: 25,
				ProbesFailed:  2,
				ProbesSuccess: 150,
				ErrorRate:     1.3,
			}

			response := struct {
				Data *AgentStatusResponse `json:"data"`
			}{Data: status}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client, _ := NewClient(&Config{
			BaseURL: server.URL,
			Auth:    AuthConfig{APIKey: "test-key", APISecret: "test-secret"},
		})

		result, err := client.Monitoring.GetAgentStatus(context.Background(), "agent-123")

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "agent-123", result.AgentID)
		assert.Equal(t, "healthy", result.Status)
	})
}

func TestMonitoringService_UpdateAgent_Comprehensive(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "PUT", r.Method)
			assert.Contains(t, r.URL.Path, "/v1/monitoring/agents/agent-123")

			agent := &MonitoringAgent{
				UUID:   "agent-123",
				Name:   "Updated Agent",
				Status: "inactive",
			}

			response := struct {
				Data *MonitoringAgent `json:"data"`
			}{Data: agent}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client, _ := NewClient(&Config{
			BaseURL: server.URL,
			Auth:    AuthConfig{APIKey: "test-key", APISecret: "test-secret"},
		})

		result, err := client.Monitoring.UpdateAgent(context.Background(), "agent-123", map[string]interface{}{
			"name":   "Updated Agent",
			"status": "inactive",
		})

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "agent-123", result.UUID)
	})
}

func TestMonitoringService_GetAgent_Comprehensive(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.URL.Path, "/v1/monitoring/agents/agent-123")

			agent := &MonitoringAgent{
				UUID:   "agent-123",
				Name:   "Test Agent",
				Status: "active",
			}

			response := struct {
				Data *MonitoringAgent `json:"data"`
			}{Data: agent}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client, _ := NewClient(&Config{
			BaseURL: server.URL,
			Auth:    AuthConfig{APIKey: "test-key", APISecret: "test-secret"},
		})

		result, err := client.Monitoring.GetAgent(context.Background(), "agent-123")

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "agent-123", result.UUID)
	})
}

func TestMonitoringService_DeleteAgent_Comprehensive(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "DELETE", r.Method)
			assert.Contains(t, r.URL.Path, "/v1/monitoring/agents/agent-123")

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client, _ := NewClient(&Config{
			BaseURL: server.URL,
			Auth:    AuthConfig{APIKey: "test-key", APISecret: "test-secret"},
		})

		err := client.Monitoring.DeleteAgent(context.Background(), "agent-123")
		require.NoError(t, err)
	})
}

func TestMonitoringService_GetProbeMetrics_Comprehensive(t *testing.T) {
	t.Run("Success - With TimeRange", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.URL.Path, "/v1/monitoring/probes/test-uuid/metrics")
			assert.Equal(t, "2023-01-01T00:00:00Z", r.URL.Query().Get("start"))
			assert.Equal(t, "2023-01-07T23:59:59Z", r.URL.Query().Get("end"))

			metrics := &ProbeMetrics{
				ProbeUUID:        "test-uuid",
				AvgResponseTime:  175.5,
				UptimePercentage: 99.5,
				TotalChecks:      1000,
				SuccessfulChecks: 995,
				FailedChecks:     5,
				LastCheck:        &CustomTime{Time: time.Now()},
				LastStatus:       "up",
			}

			response := struct {
				Data *ProbeMetrics `json:"data"`
			}{Data: metrics}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client, _ := NewClient(&Config{
			BaseURL: server.URL,
			Auth:    AuthConfig{APIKey: "test-key", APISecret: "test-secret"},
		})

		timeRange := &TimeRange{
			Start: "2023-01-01T00:00:00Z",
			End:   "2023-01-07T23:59:59Z",
		}

		result, err := client.Monitoring.GetProbeMetrics(context.Background(), "test-uuid", timeRange)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "test-uuid", result.ProbeUUID)
		assert.Equal(t, 99.5, result.UptimePercentage)
	})

	t.Run("Success - Without TimeRange", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Empty(t, r.URL.Query().Get("start"))
			assert.Empty(t, r.URL.Query().Get("end"))

			metrics := &ProbeMetrics{
				ProbeUUID:       "test-uuid",
				AvgResponseTime: 100.0,
			}

			response := struct {
				Data *ProbeMetrics `json:"data"`
			}{Data: metrics}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client, _ := NewClient(&Config{
			BaseURL: server.URL,
			Auth:    AuthConfig{APIKey: "test-key", APISecret: "test-secret"},
		})

		result, err := client.Monitoring.GetProbeMetrics(context.Background(), "test-uuid")

		require.NoError(t, err)
		assert.Equal(t, "test-uuid", result.ProbeUUID)
	})
}

func TestMonitoringService_GetAssignedProbes_Comprehensive(t *testing.T) {
	t.Run("Success - With Region", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.URL.Path, "/v1/monitoring/probes")
			assert.Equal(t, "us-east-1", r.URL.Query().Get("region"))

			assignments := []*ProbeAssignment{
				{ProbeUUID: "probe-1", Name: "Probe 1", Type: "http", Region: "us-east-1"},
				{ProbeUUID: "probe-2", Name: "Probe 2", Type: "tcp", Region: "us-east-1"},
			}

			response := struct {
				Data []*ProbeAssignment `json:"data"`
			}{Data: assignments}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client, _ := NewClient(&Config{
			BaseURL: server.URL,
			Auth:    AuthConfig{APIKey: "test-key", APISecret: "test-secret"},
		})

		result, err := client.Monitoring.GetAssignedProbes(context.Background(), "us-east-1")

		require.NoError(t, err)
		assert.Len(t, result, 2)
	})
}

func TestMonitoringService_SubmitResults_Comprehensive(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Contains(t, r.URL.Path, "/v1/monitoring/results")

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "success"})
		}))
		defer server.Close()

		client, _ := NewClient(&Config{
			BaseURL: server.URL,
			Auth:    AuthConfig{APIKey: "test-key", APISecret: "test-secret"},
		})

		results := []ProbeExecutionResult{
			{
				ProbeUUID:    "probe-1",
				ExecutedAt:   time.Now(),
				Region:       "us-east-1",
				Status:       "success",
				ResponseTime: 150,
				StatusCode:   200,
			},
		}

		err := client.Monitoring.SubmitResults(context.Background(), results)
		require.NoError(t, err)
	})
}

func TestMonitoringService_Heartbeat_Comprehensive(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Contains(t, r.URL.Path, "/v1/monitoring/heartbeat")

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "received"})
		}))
		defer server.Close()

		client, _ := NewClient(&Config{
			BaseURL: server.URL,
			Auth:    AuthConfig{APIKey: "test-key", APISecret: "test-secret"},
		})

		nodeInfo := NodeInfo{
			AgentID:      "agent-123",
			AgentVersion: "1.0.0",
			Status:       "healthy",
			LastSeen:     time.Now(),
		}

		err := client.Monitoring.Heartbeat(context.Background(), nodeInfo)
		require.NoError(t, err)
	})
}

func TestMonitoringService_ListProbeResults_Comprehensive(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.URL.Path, "/v1/monitoring/probe-results")
			assert.Equal(t, "test-probe-uuid", r.URL.Query().Get("probe_uuid"))

			results := []*ProbeResult{
				{ProbeUUID: "test-probe-uuid", Status: "up", ResponseTime: 150},
				{ProbeUUID: "test-probe-uuid", Status: "up", ResponseTime: 200},
			}

			response := struct {
				Data []*ProbeResult  `json:"data"`
				Meta *PaginationMeta `json:"meta"`
			}{
				Data: results,
				Meta: &PaginationMeta{TotalItems: 2, TotalPages: 1, Page: 1, Limit: 10},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client, _ := NewClient(&Config{
			BaseURL: server.URL,
			Auth:    AuthConfig{APIKey: "test-key", APISecret: "test-secret"},
		})

		opts := &ProbeResultListOptions{
			ProbeUUID: "test-probe-uuid",
			ListOptions: ListOptions{
				Page:  1,
				Limit: 10,
			},
		}

		results, meta, err := client.Monitoring.ListProbeResults(context.Background(), opts)

		require.NoError(t, err)
		assert.Len(t, results, 2)
		assert.NotNil(t, meta)
	})
}

func TestMonitoringAgentListOptions_ToQuery_Comprehensive(t *testing.T) {
	t.Run("All Options Set", func(t *testing.T) {
		enabled := true
		opts := &MonitoringAgentListOptions{
			Status:  "active",
			Region:  "us-east-1",
			Type:    "private",
			Enabled: &enabled,
			ListOptions: ListOptions{
				Page:   1,
				Limit:  25,
				Search: "test",
			},
		}

		result := opts.ToQuery()

		assert.Equal(t, "1", result["page"])
		assert.Equal(t, "25", result["limit"])
		assert.Equal(t, "test", result["search"])
		assert.Equal(t, "active", result["status"])
		assert.Equal(t, "us-east-1", result["region"])
		assert.Equal(t, "private", result["type"])
		assert.Equal(t, "true", result["enabled"])
	})
}

func TestMonitoringDeploymentListOptions_ToQuery_Comprehensive(t *testing.T) {
	t.Run("All Options Set", func(t *testing.T) {
		opts := &MonitoringDeploymentListOptions{
			Environment: "production",
			Region:      "eu-west-1",
			Status:      "active",
			ListOptions: ListOptions{
				Page:  1,
				Limit: 20,
			},
		}

		result := opts.ToQuery()

		assert.Equal(t, "1", result["page"])
		assert.Equal(t, "20", result["limit"])
		assert.Equal(t, "production", result["environment"])
		assert.Equal(t, "eu-west-1", result["region"])
		assert.Equal(t, "active", result["status"])
	})
}

func TestProbeResultListOptions_ToQuery_Comprehensive(t *testing.T) {
	t.Run("All Options Set", func(t *testing.T) {
		opts := &ProbeResultListOptions{
			ProbeUUID: "probe-uuid-123",
			Status:    "up",
			Region:    "us-east-1",
			ListOptions: ListOptions{
				Page:  1,
				Limit: 50,
			},
		}

		result := opts.ToQuery()

		assert.Equal(t, "1", result["page"])
		assert.Equal(t, "50", result["limit"])
		assert.Equal(t, "probe-uuid-123", result["probe_uuid"])
		assert.Equal(t, "up", result["status"])
		assert.Equal(t, "us-east-1", result["region"])
	})
}

// TestMonitoringService_NetworkErrors tests handling of network-level errors
func TestMonitoringService_NetworkErrors(t *testing.T) {
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
				_, _, apiErr = client.Monitoring.ListProbes(ctx, nil)
			case "get":
				_, apiErr = client.Monitoring.GetProbe(ctx, "test-id")
			case "create":
				probe := &MonitoringProbe{Name: "test"}
				_, apiErr = client.Monitoring.CreateProbe(ctx, probe)
			case "update":
				probe := &MonitoringProbe{Name: "updated"}
				_, apiErr = client.Monitoring.UpdateProbe(ctx, "test-id", probe)
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
