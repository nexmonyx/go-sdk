package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Error path tests to improve coverage from 75% to 87.5%+

func TestMonitoringService_UpdateProbe_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Invalid probe configuration",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	probe, err := client.Monitoring.UpdateProbe(context.Background(), "probe-1", &MonitoringProbe{
		Name: "Updated Probe",
	})
	assert.Error(t, err)
	assert.Nil(t, probe)
}

func TestMonitoringService_GetStatus_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Monitoring service unavailable",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	status, err := client.Monitoring.GetStatus(context.Background(), "probe-1")
	assert.Error(t, err)
	assert.Nil(t, status)
}

func TestMonitoringService_TestProbe_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Invalid probe configuration",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, err := client.Monitoring.TestProbe(context.Background(), "probe-1")
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestMonitoringService_GetAgentStatus_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Agent not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	status, err := client.Monitoring.GetAgentStatus(context.Background(), "agent-999")
	assert.Error(t, err)
	assert.Nil(t, status)
}

func TestMonitoringService_RegisterAgent_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Agent already registered",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	agent, err := client.Monitoring.RegisterAgent(context.Background(), &AgentRegistration{
		Name:   "agent-123",
		Type:   "monitoring",
		Region: "us-east-1",
	})
	assert.Error(t, err)
	assert.Nil(t, agent)
}

func TestMonitoringService_UpdateAgent_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Agent not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	agent, err := client.Monitoring.UpdateAgent(context.Background(), "agent-999", map[string]interface{}{
		"status": "active",
	})
	assert.Error(t, err)
	assert.Nil(t, agent)
}

func TestMonitoringService_GetAgent_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Agent not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	agent, err := client.Monitoring.GetAgent(context.Background(), "agent-999")
	assert.Error(t, err)
	assert.Nil(t, agent)
}

// ToQuery edge case for ProbeResultListOptions

func TestProbeResultListOptions_ToQuery_AllFilters(t *testing.T) {
	opts := &ProbeResultListOptions{
		ListOptions: ListOptions{
			Page:  2,
			Limit: 50,
		},
		ProbeUUID: "probe-uuid-123",
		Status:    "failed",
		Region:    "us-east-1",
	}

	query := opts.ToQuery()
	assert.Equal(t, "probe-uuid-123", query["probe_uuid"])
	assert.Equal(t, "failed", query["status"])
	assert.Equal(t, "2", query["page"])
	assert.Equal(t, "50", query["limit"])
	assert.Equal(t, "us-east-1", query["region"])
}

func TestProbeResultListOptions_ToQuery_MinimalFilters(t *testing.T) {
	opts := &ProbeResultListOptions{
		ListOptions: ListOptions{
			Page: 1,
		},
		ProbeUUID: "probe-123",
	}

	query := opts.ToQuery()
	assert.Equal(t, "probe-123", query["probe_uuid"])
	assert.Equal(t, "1", query["page"])
	// Empty values should not be in query
	assert.Empty(t, query["status"])
	assert.Empty(t, query["region"])
}
