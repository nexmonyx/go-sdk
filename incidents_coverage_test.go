package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIncidentsService_CreateIncident(t *testing.T) {
	tests := []struct {
		name       string
		req        CreateIncidentRequest
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name: "successful create",
			req: CreateIncidentRequest{
				Title:       "Test Incident",
				Description: "Test description",
				Severity:    IncidentSeverityCritical,
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data":   map[string]interface{}{"id": 1, "title": "Test Incident"},
			},
			wantErr: false,
		},
		{
			name: "validation error - missing title",
			req: CreateIncidentRequest{
				Description: "Test description",
				Severity:    IncidentSeverityCritical,
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "title is required",
			},
			wantErr: true,
		},
		{
			name: "unauthorized",
			req: CreateIncidentRequest{
				Title:    "Test Incident",
				Severity: IncidentSeverityCritical,
			},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "authentication required",
			},
			wantErr: true,
		},
		{
			name: "forbidden",
			req: CreateIncidentRequest{
				Title:    "Test Incident",
				Severity: IncidentSeverityCritical,
			},
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "insufficient permissions",
			},
			wantErr: true,
		},
		{
			name: "internal server error",
			req: CreateIncidentRequest{
				Title:    "Test Incident",
				Severity: IncidentSeverityCritical,
			},
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/v1/incidents", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL, RetryCount: 0})
			incident, err := client.Incidents.CreateIncident(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, incident)
			}
		})
	}
}

func TestIncidentsService_GetIncident(t *testing.T) {
	tests := []struct {
		name       string
		incidentID uint
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:       "successful get",
			incidentID: 1,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data":   map[string]interface{}{"id": 1, "title": "Test Incident"},
			},
			wantErr: false,
		},
		{
			name:       "incident not found",
			incidentID: 999,
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "incident not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			incidentID: 1,
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "authentication required",
			},
			wantErr: true,
		},
		{
			name:       "forbidden",
			incidentID: 1,
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "insufficient permissions",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/incidents/")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL, RetryCount: 0})
			incident, err := client.Incidents.GetIncident(context.Background(), tt.incidentID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, incident)
			}
		})
	}
}

func TestIncidentsService_UpdateIncident(t *testing.T) {
	tests := []struct {
		name       string
		incidentID uint
		req        UpdateIncidentRequest
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:       "successful update",
			incidentID: 1,
			req:        UpdateIncidentRequest{Status: IncidentStatusResolved},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data":   map[string]interface{}{"id": 1, "title": "Updated Incident"},
			},
			wantErr: false,
		},
		{
			name:       "incident not found",
			incidentID: 999,
			req:        UpdateIncidentRequest{Status: IncidentStatusResolved},
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "incident not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			incidentID: 1,
			req:        UpdateIncidentRequest{Status: IncidentStatusResolved},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "authentication required",
			},
			wantErr: true,
		},
		{
			name:       "forbidden",
			incidentID: 1,
			req:        UpdateIncidentRequest{Status: IncidentStatusResolved},
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "insufficient permissions",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "PUT", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/incidents/")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL, RetryCount: 0})
			incident, err := client.Incidents.UpdateIncident(context.Background(), tt.incidentID, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, incident)
			}
		})
	}
}

func TestIncidentsService_ListIncidents(t *testing.T) {
	tests := []struct {
		name       string
		opts       *IncidentListOptions
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:       "successful list",
			opts:       &IncidentListOptions{Status: "active", Severity: "critical"},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"incidents": []map[string]interface{}{{"id": 1}},
					"total":     1,
				},
			},
			wantErr: false,
		},
		{
			name:       "empty list",
			opts:       &IncidentListOptions{Status: "resolved"},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"incidents": []map[string]interface{}{},
					"total":     0,
				},
			},
			wantErr: false,
		},
		{
			name:       "unauthorized",
			opts:       &IncidentListOptions{Status: "active"},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "authentication required",
			},
			wantErr: true,
		},
		{
			name:       "forbidden",
			opts:       &IncidentListOptions{Status: "active"},
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "insufficient permissions",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/v1/incidents", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL, RetryCount: 0})
			response, err := client.Incidents.ListIncidents(context.Background(), tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
			}
		})
	}
}

func TestIncidentsService_GetRecentIncidents(t *testing.T) {
	tests := []struct {
		name       string
		limit      int
		severity   string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:       "successful get recent incidents",
			limit:      10,
			severity:   "critical",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"incidents": []map[string]interface{}{{"id": 1, "title": "Incident 1"}},
					"total":     1,
				},
			},
			wantErr: false,
		},
		{
			name:       "empty results",
			limit:      10,
			severity:   "low",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"incidents": []map[string]interface{}{},
					"total":     0,
				},
			},
			wantErr: false,
		},
		{
			name:       "unauthorized",
			limit:      10,
			severity:   "critical",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "authentication required",
			},
			wantErr: true,
		},
		{
			name:       "forbidden",
			limit:      10,
			severity:   "critical",
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "insufficient permissions",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			limit:      10,
			severity:   "critical",
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/v1/incidents/recent", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL, RetryCount: 0})
			incidents, err := client.Incidents.GetRecentIncidents(context.Background(), tt.limit, tt.severity)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, incidents)
			}
		})
	}
}

func TestIncidentsService_GetIncidentStats(t *testing.T) {
	tests := []struct {
		name       string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:       "successful get stats",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data":   map[string]interface{}{"total": 100, "active": 10, "resolved": 90},
			},
			wantErr: false,
		},
		{
			name:       "empty stats",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data":   map[string]interface{}{"total": 0, "active": 0, "resolved": 0},
			},
			wantErr: false,
		},
		{
			name:       "unauthorized",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "authentication required",
			},
			wantErr: true,
		},
		{
			name:       "forbidden",
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "insufficient permissions",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/v1/incidents/stats", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL, RetryCount: 0})
			stats, err := client.Incidents.GetIncidentStats(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, stats)
			}
		})
	}
}

func TestIncidentsService_ResolveIncident(t *testing.T) {
	tests := []struct {
		name       string
		incidentID uint
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:       "successful resolve",
			incidentID: 1,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data":   map[string]interface{}{"id": 1, "status": "resolved"},
			},
			wantErr: false,
		},
		{
			name:       "incident not found",
			incidentID: 999,
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "incident not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			incidentID: 1,
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "authentication required",
			},
			wantErr: true,
		},
		{
			name:       "forbidden",
			incidentID: 1,
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "insufficient permissions",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			incidentID: 1,
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "PUT", r.Method)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL, RetryCount: 0})
			incident, err := client.Incidents.ResolveIncident(context.Background(), tt.incidentID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, incident)
			}
		})
	}
}

func TestIncidentsService_AcknowledgeIncident(t *testing.T) {
	tests := []struct {
		name       string
		incidentID uint
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:       "successful acknowledge",
			incidentID: 1,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data":   map[string]interface{}{"id": 1, "status": "acknowledged"},
			},
			wantErr: false,
		},
		{
			name:       "incident not found",
			incidentID: 999,
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "incident not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			incidentID: 1,
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "authentication required",
			},
			wantErr: true,
		},
		{
			name:       "forbidden",
			incidentID: 1,
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "insufficient permissions",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			incidentID: 1,
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL, RetryCount: 0})
			incident, err := client.Incidents.AcknowledgeIncident(context.Background(), tt.incidentID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, incident)
			}
		})
	}
}

func TestIncidentsService_CreateIncidentFromAlert(t *testing.T) {
	tests := []struct {
		name        string
		orgID       uint
		alertID     uint
		title       string
		severity    IncidentSeverity
		serverID    *uint
		description string
		mockStatus  int
		mockBody    interface{}
		wantErr     bool
	}{
		{
			name:        "successful create from alert",
			orgID:       1,
			alertID:     100,
			title:       "Alert 1",
			severity:    IncidentSeverityCritical,
			serverID:    func() *uint { id := uint(123); return &id }(),
			description: "Test alert",
			mockStatus:  http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data":   map[string]interface{}{"id": 1, "title": "Alert 1"},
			},
			wantErr: false,
		},
		{
			name:        "create without server ID",
			orgID:       1,
			alertID:     100,
			title:       "Alert 2",
			severity:    IncidentSeverityWarning,
			serverID:    nil,
			description: "Alert without server",
			mockStatus:  http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data":   map[string]interface{}{"id": 2},
			},
			wantErr: false,
		},
		{
			name:        "unauthorized",
			orgID:       1,
			alertID:     100,
			title:       "Alert",
			severity:    IncidentSeverityCritical,
			serverID:    nil,
			description: "Test",
			mockStatus:  http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "authentication required",
			},
			wantErr: true,
		},
		{
			name:        "forbidden",
			orgID:       1,
			alertID:     100,
			title:       "Alert",
			severity:    IncidentSeverityCritical,
			serverID:    nil,
			description: "Test",
			mockStatus:  http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "insufficient permissions",
			},
			wantErr: true,
		},
		{
			name:        "server error",
			orgID:       1,
			alertID:     100,
			title:       "Alert",
			severity:    IncidentSeverityCritical,
			serverID:    nil,
			description: "Test",
			mockStatus:  http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL, RetryCount: 0})
			incident, err := client.Incidents.CreateIncidentFromAlert(context.Background(), tt.orgID, tt.alertID, tt.title, tt.severity, tt.serverID, tt.description)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, incident)
			}
		})
	}
}

func TestIncidentsService_CreateIncidentFromProbe(t *testing.T) {
	tests := []struct {
		name        string
		orgID       uint
		probeID     uint
		title       string
		description string
		mockStatus  int
		mockBody    interface{}
		wantErr     bool
	}{
		{
			name:        "successful create from probe",
			orgID:       1,
			probeID:     200,
			title:       "Probe 1",
			description: "Probe failed",
			mockStatus:  http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data":   map[string]interface{}{"id": 1, "title": "Probe 1"},
			},
			wantErr: false,
		},
		{
			name:        "create with minimal description",
			orgID:       1,
			probeID:     201,
			title:       "Probe Check Failed",
			description: "Down",
			mockStatus:  http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data":   map[string]interface{}{"id": 2},
			},
			wantErr: false,
		},
		{
			name:        "unauthorized",
			orgID:       1,
			probeID:     200,
			title:       "Probe",
			description: "Test",
			mockStatus:  http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "authentication required",
			},
			wantErr: true,
		},
		{
			name:        "forbidden",
			orgID:       1,
			probeID:     200,
			title:       "Probe",
			description: "Test",
			mockStatus:  http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "insufficient permissions",
			},
			wantErr: true,
		},
		{
			name:        "server error",
			orgID:       1,
			probeID:     200,
			title:       "Probe",
			description: "Test",
			mockStatus:  http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL, RetryCount: 0})
			incident, err := client.Incidents.CreateIncidentFromProbe(context.Background(), tt.orgID, tt.probeID, tt.title, tt.description)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, incident)
			}
		})
	}
}

func TestIncidentsService_ResolveIncidentFromAlert(t *testing.T) {
	tests := []struct {
		name       string
		alertID    uint
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:       "successful resolve - no incidents",
			alertID:    100,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"incidents": []map[string]interface{}{},
				},
			},
			wantErr: false,
		},
		{
			name:       "unauthorized",
			alertID:    100,
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "authentication required",
			},
			wantErr: true,
		},
		{
			name:       "forbidden",
			alertID:    100,
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "insufficient permissions",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			alertID:    100,
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL, RetryCount: 0})
			err := client.Incidents.ResolveIncidentFromAlert(context.Background(), tt.alertID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIncidentsService_ResolveIncidentFromProbe(t *testing.T) {
	tests := []struct {
		name       string
		probeID    uint
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:       "successful resolve - no incidents",
			probeID:    200,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"incidents": []map[string]interface{}{},
				},
			},
			wantErr: false,
		},
		{
			name:       "unauthorized",
			probeID:    200,
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "authentication required",
			},
			wantErr: true,
		},
		{
			name:       "forbidden",
			probeID:    200,
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "insufficient permissions",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			probeID:    200,
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL, RetryCount: 0})
			err := client.Incidents.ResolveIncidentFromProbe(context.Background(), tt.probeID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIncidentsService_ResolveIncidentFromAlert_WithIncidents(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if r.Method == "GET" && r.URL.Path == "/v1/incidents" {
			// ListIncidents call
			sourceID := uint(100)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"incidents": []map[string]interface{}{
						{
							"id":        1,
							"source":    "alert",
							"source_id": sourceID,
							"status":    "active",
						},
						{
							"id":        2,
							"source":    "probe",
							"source_id": uint(999),
							"status":    "active",
						},
					},
					"total": 2,
					"page":  1,
					"limit": 25,
					"pages": 1,
				},
			})
		} else if r.Method == "PUT" && strings.Contains(r.URL.Path, "/incidents/") {
			// UpdateIncident call (called by ResolveIncident)
			callCount++
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "success",
				"data":   map[string]interface{}{"id": 1, "status": "resolved"},
			})
		}
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	err := client.Incidents.ResolveIncidentFromAlert(context.Background(), 100)
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount, "Should resolve exactly 1 incident")
}

func TestIncidentsService_ResolveIncidentFromAlert_ListError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Failed to list incidents",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL, RetryCount: 0})
	err := client.Incidents.ResolveIncidentFromAlert(context.Background(), 100)
	assert.Error(t, err)
}

func TestIncidentsService_ResolveIncidentFromAlert_ResolveError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == "GET" {
			// ListIncidents succeeds
			sourceID := uint(100)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"incidents": []map[string]interface{}{
						{
							"id":        1,
							"source":    "alert",
							"source_id": sourceID,
							"status":    "active",
						},
					},
				},
			})
		} else if r.Method == "PUT" {
			// UpdateIncident fails (called by ResolveIncident)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "error",
				"message": "Failed to resolve incident",
			})
		}
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL, RetryCount: 0})
	err := client.Incidents.ResolveIncidentFromAlert(context.Background(), 100)
	assert.Error(t, err)
}

func TestIncidentsService_ResolveIncidentFromProbe_WithIncidents(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if r.Method == "GET" && r.URL.Path == "/v1/incidents" {
			// ListIncidents call
			sourceID := uint(200)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"incidents": []map[string]interface{}{
						{
							"id":        10,
							"source":    "probe",
							"source_id": sourceID,
							"status":    "active",
						},
						{
							"id":        11,
							"source":    "probe",
							"source_id": sourceID,
							"status":    "active",
						},
						{
							"id":        12,
							"source":    "alert",
							"source_id": uint(999),
							"status":    "active",
						},
					},
					"total": 3,
					"page":  1,
					"limit": 25,
					"pages": 1,
				},
			})
		} else if r.Method == "PUT" && strings.Contains(r.URL.Path, "/incidents/") {
			// UpdateIncident call (called by ResolveIncident)
			callCount++
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "success",
				"data":   map[string]interface{}{"id": callCount, "status": "resolved"},
			})
		}
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	err := client.Incidents.ResolveIncidentFromProbe(context.Background(), 200)
	assert.NoError(t, err)
	assert.Equal(t, 2, callCount, "Should resolve exactly 2 incidents")
}

func TestIncidentsService_ResolveIncidentFromProbe_ListError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Access denied",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	err := client.Incidents.ResolveIncidentFromProbe(context.Background(), 200)
	assert.Error(t, err)
}

func TestIncidentsService_ResolveIncidentFromProbe_ResolveError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == "GET" {
			// ListIncidents succeeds
			sourceID := uint(200)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"incidents": []map[string]interface{}{
						{
							"id":        10,
							"source":    "probe",
							"source_id": sourceID,
							"status":    "active",
						},
					},
				},
			})
		} else if r.Method == "PUT" {
			// UpdateIncident fails (called by ResolveIncident)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "error",
				"message": "Cannot resolve incident",
			})
		}
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	err := client.Incidents.ResolveIncidentFromProbe(context.Background(), 200)
	assert.Error(t, err)
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
