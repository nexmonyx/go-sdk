package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestMonitoringService_TestProbe(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/v1/monitoring/probes/test-uuid/test") {
			t.Errorf("Expected /v1/monitoring/probes/test-uuid/test, got %s", r.URL.Path)
		}

		result := ProbeTestResult{
			ProbeUUID:    "test-uuid",
			Target:       "https://example.com",
			Type:         "http",
			Region:       "us-east-1",
			Status:       "up",
			ResponseTime: 150,
			ExecutedAt:   &CustomTime{Time: time.Now()},
		}

		response := struct {
			Data ProbeTestResult `json:"data"`
		}{Data: result}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			APIKey:    "test-key",
			APISecret: "test-secret",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	result, err := client.Monitoring.TestProbe(context.Background(), "test-uuid")
	if err != nil {
		t.Fatalf("TestProbe failed: %v", err)
	}

	if result.ProbeUUID != "test-uuid" {
		t.Errorf("Expected probe UUID 'test-uuid', got '%s'", result.ProbeUUID)
	}
	if result.Status != "up" {
		t.Errorf("Expected status 'up', got '%s'", result.Status)
	}
}

func TestMonitoringService_ListAgents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/v1/monitoring/agents") {
			t.Errorf("Expected /v1/monitoring/agents, got %s", r.URL.Path)
		}

		// Check query parameters
		query := r.URL.Query()
		if query.Get("type") != "private" {
			t.Errorf("Expected type=private, got %s", query.Get("type"))
		}

		agents := []MonitoringAgent{
			{
				UUID:   "agent-uuid-1",
				Name:   "test-agent-1",
				Status: "active",
			},
			{
				UUID:   "agent-uuid-2",
				Name:   "test-agent-2",
				Status: "active",
			},
		}

		response := struct {
			Data []MonitoringAgent `json:"data"`
			Meta *PaginationMeta   `json:"meta"`
		}{
			Data: agents,
			Meta: &PaginationMeta{
				TotalItems: 2,
				TotalPages: 1,
				Page:       1,
				Limit:      10,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			APIKey:    "test-key",
			APISecret: "test-secret",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	opts := &MonitoringAgentListOptions{
		Type: "private",
		ListOptions: ListOptions{
			Page:  1,
			Limit: 10,
		},
	}

	agents, meta, err := client.Monitoring.ListAgents(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListAgents failed: %v", err)
	}

	if len(agents) != 2 {
		t.Errorf("Expected 2 agents, got %d", len(agents))
	}
	if agents[0].UUID != "agent-uuid-1" {
		t.Errorf("Expected first agent UUID 'agent-uuid-1', got '%s'", agents[0].UUID)
	}
	if meta.TotalItems != 2 {
		t.Errorf("Expected total items 2, got %d", meta.TotalItems)
	}
}

func TestMonitoringService_RegisterAgent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/v1/monitoring/agents") {
			t.Errorf("Expected /v1/monitoring/agents, got %s", r.URL.Path)
		}

		var req AgentRegistration
		json.NewDecoder(r.Body).Decode(&req)

		if req.Name != "test-agent" {
			t.Errorf("Expected agent name 'test-agent', got '%s'", req.Name)
		}

		agent := MonitoringAgent{
			UUID:   "new-agent-uuid",
			Name:   req.Name,
			Status: "active",
		}

		response := struct {
			Data MonitoringAgent `json:"data"`
		}{Data: agent}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			APIKey:    "test-key",
			APISecret: "test-secret",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	req := &AgentRegistration{
		Name:         "test-agent",
		Type:         "private",
		Region:       "us-east-1",
		MaxProbes:    100,
		Capabilities: []string{"probe:read", "probe:write"},
	}

	agent, err := client.Monitoring.RegisterAgent(context.Background(), req)
	if err != nil {
		t.Fatalf("RegisterAgent failed: %v", err)
	}

	if agent.UUID != "new-agent-uuid" {
		t.Errorf("Expected agent UUID 'new-agent-uuid', got '%s'", agent.UUID)
	}
	if agent.Name != "test-agent" {
		t.Errorf("Expected agent name 'test-agent', got '%s'", agent.Name)
	}
}

func TestMonitoringService_ListDeployments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/deployments") {
			t.Errorf("Expected path to contain /deployments, got %s", r.URL.Path)
		}

		// Check query parameters
		query := r.URL.Query()
		if query.Get("environment") != "dev" {
			t.Errorf("Expected environment=dev, got %s", query.Get("environment"))
		}

		deployments := []MonitoringDeployment{
			{
				OrganizationID: 1,
				Region:         "us-east-1",
				NamespaceName:  "nexmonyx-dev",
				DeploymentName: "monitoring-agent",
				Status:         "active",
				ErrorCount:     0,
			},
		}

		response := struct {
			Data []MonitoringDeployment `json:"data"`
			Meta *PaginationMeta        `json:"meta"`
		}{
			Data: deployments,
			Meta: &PaginationMeta{
				TotalItems: 1,
				TotalPages: 1,
				Page:       1,
				Limit:      10,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			APIKey:    "test-key",
			APISecret: "test-secret",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	opts := &MonitoringDeploymentListOptions{
		Environment: "dev", // This filter is on the ListOptions, not the struct itself
		ListOptions: ListOptions{
			Page:  1,
			Limit: 10,
		},
	}

	deployments, meta, err := client.Monitoring.ListDeployments(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListDeployments failed: %v", err)
	}

	if len(deployments) != 1 {
		t.Errorf("Expected 1 deployment, got %d", len(deployments))
	}
	// The Environment field is not in the MonitoringDeployment struct from monitoring_deployments.go
	// if deployments[0].Environment != "dev" {
	// 	t.Errorf("Expected environment 'dev', got '%s'", deployments[0].Environment)
	// }
	if meta.TotalItems != 1 {
		t.Errorf("Expected total items 1, got %d", meta.TotalItems)
	}
}

func TestMonitoringService_ListProbeResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/v1/monitoring/probe-results") {
			t.Errorf("Expected /v1/monitoring/probe-results, got %s", r.URL.Path)
		}

		// Check query parameters
		query := r.URL.Query()
		if query.Get("probe_uuid") != "test-probe-uuid" {
			t.Errorf("Expected probe_uuid=test-probe-uuid, got %s", query.Get("probe_uuid"))
		}

		results := []ProbeResult{
			{
				ProbeID:      1,
				ProbeUUID:    "test-probe-uuid",
				Region:       "us-east-1",
				Status:       "up",
				ResponseTime: 150,
				ExecutedAt:   &CustomTime{Time: time.Now()},
			},
			{
				ProbeID:      1,
				ProbeUUID:    "test-probe-uuid",
				Region:       "us-west-2",
				Status:       "up",
				ResponseTime: 200,
				ExecutedAt:   &CustomTime{Time: time.Now().Add(-5 * time.Minute)},
			},
		}

		response := struct {
			Data []ProbeResult   `json:"data"`
			Meta *PaginationMeta `json:"meta"`
		}{
			Data: results,
			Meta: &PaginationMeta{
				TotalItems: 2,
				TotalPages: 1,
				Page:       1,
				Limit:      10,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			APIKey:    "test-key",
			APISecret: "test-secret",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	opts := &ProbeResultListOptions{
		ProbeUUID: "test-probe-uuid",
		ListOptions: ListOptions{
			Page:  1,
			Limit: 10,
		},
	}

	results, meta, err := client.Monitoring.ListProbeResults(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListProbeResults failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
	if results[0].ProbeUUID != "test-probe-uuid" {
		t.Errorf("Expected probe UUID 'test-probe-uuid', got '%s'", results[0].ProbeUUID)
	}
	if meta.TotalItems != 2 {
		t.Errorf("Expected total items 2, got %d", meta.TotalItems)
	}
}

func TestMonitoringService_GetProbeMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/v1/monitoring/probes/test-uuid/metrics") {
			t.Errorf("Expected /v1/monitoring/probes/test-uuid/metrics, got %s", r.URL.Path)
		}

		metrics := ProbeMetrics{
			ProbeID:          1,
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
			Data ProbeMetrics `json:"data"`
		}{Data: metrics}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			APIKey:    "test-key",
			APISecret: "test-secret",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	timeRange := &TimeRange{
		Start: "2023-01-01T00:00:00Z",
		End:   "2023-01-07T23:59:59Z",
	}

	metrics, err := client.Monitoring.GetProbeMetrics(context.Background(), "test-uuid", timeRange)
	if err != nil {
		t.Fatalf("GetProbeMetrics failed: %v", err)
	}

	if metrics.ProbeUUID != "test-uuid" {
		t.Errorf("Expected probe UUID 'test-uuid', got '%s'", metrics.ProbeUUID)
	}
	if metrics.UptimePercentage != 99.5 {
		t.Errorf("Expected uptime 99.5%%, got %f%%", metrics.UptimePercentage)
	}
	if metrics.TotalChecks != 1000 {
		t.Errorf("Expected 1000 total checks, got %d", metrics.TotalChecks)
	}
}
