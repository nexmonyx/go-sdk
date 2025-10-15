package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIncidentsService_CreateIncident(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/incidents", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   map[string]interface{}{"id": 1, "title": "Test Incident"},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	req := CreateIncidentRequest{
		Title:       "Test Incident",
		Description: "Test description",
		Severity:    IncidentSeverityCritical,
	}
	incident, err := client.Incidents.CreateIncident(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, incident)
}

func TestIncidentsService_GetIncident(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/v1/incidents/")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   map[string]interface{}{"id": 1, "title": "Test Incident"},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	incident, err := client.Incidents.GetIncident(context.Background(), 1)
	assert.NoError(t, err)
	assert.NotNil(t, incident)
}

func TestIncidentsService_UpdateIncident(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Contains(t, r.URL.Path, "/v1/incidents/")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   map[string]interface{}{"id": 1, "title": "Updated Incident"},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	req := UpdateIncidentRequest{Status: IncidentStatusResolved}
	incident, err := client.Incidents.UpdateIncident(context.Background(), 1, req)
	assert.NoError(t, err)
	assert.NotNil(t, incident)
}

func TestIncidentsService_ListIncidents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/incidents", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"incidents": []map[string]interface{}{{"id": 1}},
				"total":     1,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	opts := &IncidentListOptions{Status: "active", Severity: "critical"}
	response, err := client.Incidents.ListIncidents(context.Background(), opts)
	assert.NoError(t, err)
	assert.NotNil(t, response)
}

func TestIncidentsService_GetRecentIncidents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/incidents/recent", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"incidents": []map[string]interface{}{{"id": 1}},
				"total":     1,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	incidents, err := client.Incidents.GetRecentIncidents(context.Background(), 10, "critical")
	assert.NoError(t, err)
	assert.NotNil(t, incidents)
}

func TestIncidentsService_GetIncidentStats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/incidents/stats", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   map[string]interface{}{"total": 100, "active": 10},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	stats, err := client.Incidents.GetIncidentStats(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, stats)
}

func TestIncidentsService_ResolveIncident(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   map[string]interface{}{"id": 1, "status": "resolved"},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	incident, err := client.Incidents.ResolveIncident(context.Background(), 1)
	assert.NoError(t, err)
	assert.NotNil(t, incident)
}

func TestIncidentsService_AcknowledgeIncident(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   map[string]interface{}{"id": 1, "status": "acknowledged"},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	incident, err := client.Incidents.AcknowledgeIncident(context.Background(), 1)
	assert.NoError(t, err)
	assert.NotNil(t, incident)
}

func TestIncidentsService_CreateIncidentFromAlert(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   map[string]interface{}{"id": 1},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	serverID := uint(123)
	incident, err := client.Incidents.CreateIncidentFromAlert(context.Background(), 1, 100, "Alert 1", IncidentSeverityCritical, &serverID, "Test alert")
	assert.NoError(t, err)
	assert.NotNil(t, incident)
}

func TestIncidentsService_CreateIncidentFromProbe(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   map[string]interface{}{"id": 1},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	incident, err := client.Incidents.CreateIncidentFromProbe(context.Background(), 1, 200, "Probe 1", "Probe failed")
	assert.NoError(t, err)
	assert.NotNil(t, incident)
}

func TestIncidentsService_ResolveIncidentFromAlert(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"incidents": []map[string]interface{}{},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	err := client.Incidents.ResolveIncidentFromAlert(context.Background(), 100)
	assert.NoError(t, err)
}

func TestIncidentsService_ResolveIncidentFromProbe(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"incidents": []map[string]interface{}{},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	err := client.Incidents.ResolveIncidentFromProbe(context.Background(), 200)
	assert.NoError(t, err)
}

func TestIncidentsService_Errors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Internal error",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	_, err := client.Incidents.CreateIncident(context.Background(), CreateIncidentRequest{})
	assert.Error(t, err)

	_, err = client.Incidents.GetIncident(context.Background(), 1)
	assert.Error(t, err)

	_, err = client.Incidents.ListIncidents(context.Background(), nil)
	assert.Error(t, err)

	_, err = client.Incidents.GetRecentIncidents(context.Background(), 0, "")
	assert.Error(t, err)

	_, err = client.Incidents.GetIncidentStats(context.Background())
	assert.Error(t, err)
}
