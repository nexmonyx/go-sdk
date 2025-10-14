package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIncidentsService_CreateIncident(t *testing.T) {
	tests := []struct {
		name           string
		request        CreateIncidentRequest
		responseStatus int
		responseBody   interface{}
		wantErr        bool
		validateResult func(*testing.T, *Incident)
	}{
		{
			name: "create critical incident",
			request: CreateIncidentRequest{
				Title:       "Database Down",
				Description: "Production database is not responding",
				Severity:    IncidentSeverityCritical,
				ServerID:    uintPtr(123),
			},
			responseStatus: http.StatusOK,
			responseBody: struct {
				Status  string     `json:"status"`
				Message string     `json:"message"`
				Data    *Incident  `json:"data"`
			}{
				Status:  "success",
				Message: "Incident created successfully",
				Data: &Incident{
					GormModel:   GormModel{ID: 1},
					Title:       "Database Down",
					Description: "Production database is not responding",
					Severity:    IncidentSeverityCritical,
					Status:      IncidentStatusActive,
					Source:      IncidentSourceAlert,
					SourceID:    uintPtr(123),
				},
			},
			validateResult: func(t *testing.T, inc *Incident) {
				assert.Equal(t, "Database Down", inc.Title)
				assert.Equal(t, IncidentSeverityCritical, inc.Severity)
				assert.Equal(t, IncidentStatusActive, inc.Status)
				assert.NotNil(t, inc.SourceID)
			},
		},
		{
			name: "create warning incident",
			request: CreateIncidentRequest{
				Title:       "High CPU Usage",
				Description: "CPU usage above 90% for 5 minutes",
				Severity:    IncidentSeverityWarning,
				ServerID:    uintPtr(456),
				Metadata: map[string]interface{}{
					"cpu_usage": 92.5,
					"duration":  300,
				},
			},
			responseStatus: http.StatusOK,
			responseBody: struct {
				Status  string    `json:"status"`
				Message string    `json:"message"`
				Data    *Incident `json:"data"`
			}{
				Status: "success",
				Data: &Incident{
					GormModel:   GormModel{ID: 2},
					Title:       "High CPU Usage",
					Severity:    IncidentSeverityWarning,
					Status:      IncidentStatusActive,
				},
			},
		},
		{
			name: "create probe failure incident",
			request: CreateIncidentRequest{
				Title:       "Probe Failed",
				Description: "HTTP probe timeout",
				Severity:    IncidentSeverityCritical,
				ProbeID:     uintPtr(789),
			},
			responseStatus: http.StatusOK,
			responseBody: struct {
				Status  string    `json:"status"`
				Message string    `json:"message"`
				Data    *Incident `json:"data"`
			}{
				Status: "success",
				Data: &Incident{
					GormModel:   GormModel{ID: 3},
					Title:       "Probe Failed",
					Severity:    IncidentSeverityCritical,
					Source:      IncidentSourceProbe,
					SourceID:    uintPtr(789),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/v1/incidents", r.URL.Path)

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL})
			result, err := client.Incidents.CreateIncident(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}

func TestIncidentsService_GetIncident(t *testing.T) {
	tests := []struct {
		name           string
		incidentID     uint
		responseStatus int
		responseBody   interface{}
		wantErr        bool
	}{
		{
			name:           "get existing incident",
			incidentID:     123,
			responseStatus: http.StatusOK,
			responseBody: struct {
				Status  string    `json:"status"`
				Message string    `json:"message"`
				Data    *Incident `json:"data"`
			}{
				Status: "success",
				Data: &Incident{
					GormModel:   GormModel{ID: 123},
					Title:       "Test Incident",
					Severity:    IncidentSeverityCritical,
					Status:      IncidentStatusActive,
				},
			},
		},
		{
			name:           "incident not found",
			incidentID:     999,
			responseStatus: http.StatusNotFound,
			responseBody: struct {
				Status  string `json:"status"`
				Message string `json:"message"`
				Data    *Incident `json:"data"`
			}{
				Status:  "error",
				Message: "incident not found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/incidents/")

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL})
			result, err := client.Incidents.GetIncident(context.Background(), tt.incidentID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.incidentID, result.ID)
			}
		})
	}
}

func TestIncidentsService_UpdateIncident(t *testing.T) {
	tests := []struct {
		name           string
		incidentID     uint
		request        UpdateIncidentRequest
		responseStatus int
		wantErr        bool
	}{
		{
			name:       "update status to resolved",
			incidentID: 123,
			request: UpdateIncidentRequest{
				Status: IncidentStatusResolved,
			},
			responseStatus: http.StatusOK,
		},
		{
			name:       "update status to acknowledged",
			incidentID: 456,
			request: UpdateIncidentRequest{
				Status: IncidentStatusAcknowledged,
			},
			responseStatus: http.StatusOK,
		},
		{
			name:       "update description",
			incidentID: 789,
			request: UpdateIncidentRequest{
				Description: "Updated description with more details",
			},
			responseStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "PUT", r.Method)
				assert.Contains(t, r.URL.Path, "/incidents/")

				var req UpdateIncidentRequest
				json.NewDecoder(r.Body).Decode(&req)
				assert.Equal(t, tt.request.Status, req.Status)

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(struct {
					Status  string    `json:"status"`
					Message string    `json:"message"`
					Data    *Incident `json:"data"`
				}{
					Status: "success",
					Data: &Incident{
						GormModel: GormModel{ID: tt.incidentID},
						Status:    tt.request.Status,
					},
				})
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL})
			result, err := client.Incidents.UpdateIncident(context.Background(), tt.incidentID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.request.Status, result.Status)
			}
		})
	}
}

func TestIncidentsService_ListIncidents(t *testing.T) {
	tests := []struct {
		name           string
		options        *IncidentListOptions
		responseStatus int
		responseBody   interface{}
		validateResult func(*testing.T, *IncidentListResponse)
	}{
		{
			name: "list all incidents",
			options: &IncidentListOptions{
				ListOptions: ListOptions{Page: 1, Limit: 10},
			},
			responseStatus: http.StatusOK,
			responseBody: struct {
				Status  string                `json:"status"`
				Message string                `json:"message"`
				Data    *IncidentListResponse `json:"data"`
			}{
				Status: "success",
				Data: &IncidentListResponse{
					Incidents: []Incident{
						{GormModel: GormModel{ID: 1}, Title: "Incident 1", Severity: IncidentSeverityCritical},
						{GormModel: GormModel{ID: 2}, Title: "Incident 2", Severity: IncidentSeverityWarning},
					},
					Total: 2,
					Page:  1,
					Limit: 10,
					Pages: 1,
				},
			},
			validateResult: func(t *testing.T, resp *IncidentListResponse) {
				assert.Len(t, resp.Incidents, 2)
				assert.Equal(t, int64(2), resp.Total)
			},
		},
		{
			name: "filter by severity",
			options: &IncidentListOptions{
				ListOptions: ListOptions{Page: 1, Limit: 10},
				Severity:    string(IncidentSeverityCritical),
			},
			responseStatus: http.StatusOK,
			responseBody: struct {
				Status  string                `json:"status"`
				Message string                `json:"message"`
				Data    *IncidentListResponse `json:"data"`
			}{
				Status: "success",
				Data: &IncidentListResponse{
					Incidents: []Incident{
						{GormModel: GormModel{ID: 1}, Severity: IncidentSeverityCritical},
					},
					Total: 1,
				},
			},
			validateResult: func(t *testing.T, resp *IncidentListResponse) {
				assert.Len(t, resp.Incidents, 1)
				assert.Equal(t, IncidentSeverityCritical, resp.Incidents[0].Severity)
			},
		},
		{
			name: "filter by server",
			options: &IncidentListOptions{
				ServerID: 123,
			},
			responseStatus: http.StatusOK,
			responseBody: struct {
				Status  string                `json:"status"`
				Message string                `json:"message"`
				Data    *IncidentListResponse `json:"data"`
			}{
				Status: "success",
				Data: &IncidentListResponse{
					Incidents: []Incident{
						{GormModel: GormModel{ID: 1}, Source: IncidentSourceAlert, SourceID: uintPtr(123)},
					},
					Total: 1,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/v1/incidents", r.URL.Path)

				if tt.options != nil {
					if tt.options.Severity != "" {
						assert.Equal(t, tt.options.Severity, r.URL.Query().Get("severity"))
					}
					if tt.options.ServerID > 0 {
						assert.NotEmpty(t, r.URL.Query().Get("server_id"))
					}
				}

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL})
			result, err := client.Incidents.ListIncidents(context.Background(), tt.options)

			require.NoError(t, err)
			assert.NotNil(t, result)
			if tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

func TestIncidentsService_GetRecentIncidents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/incidents/recent", r.URL.Path)

		limit := r.URL.Query().Get("limit")
		severity := r.URL.Query().Get("severity")

		incidents := []Incident{
			{GormModel: GormModel{ID: 1}, Title: "Recent 1"},
			{GormModel: GormModel{ID: 2}, Title: "Recent 2"},
		}

		if severity == string(IncidentSeverityCritical) {
			incidents = []Incident{
				{GormModel: GormModel{ID: 1}, Severity: IncidentSeverityCritical},
			}
		}

		json.NewEncoder(w).Encode(struct {
			Status  string `json:"status"`
			Message string `json:"message"`
			Data    struct {
				Incidents []Incident `json:"incidents"`
				Total     int64      `json:"total"`
				Limit     int        `json:"limit"`
			} `json:"data"`
		}{
			Status: "success",
			Data: struct {
				Incidents []Incident `json:"incidents"`
				Total     int64      `json:"total"`
				Limit     int        `json:"limit"`
			}{
				Incidents: incidents,
				Total:     int64(len(incidents)),
				Limit:     10,
			},
		})

		if limit != "" {
			assert.NotEmpty(t, limit)
		}
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	t.Run("get recent without filters", func(t *testing.T) {
		incidents, err := client.Incidents.GetRecentIncidents(context.Background(), 10, "")
		require.NoError(t, err)
		assert.Len(t, incidents, 2)
	})

	t.Run("get recent with severity filter", func(t *testing.T) {
		incidents, err := client.Incidents.GetRecentIncidents(context.Background(), 10, string(IncidentSeverityCritical))
		require.NoError(t, err)
		assert.Len(t, incidents, 1)
	})
}

func TestIncidentsService_GetIncidentStats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/incidents/stats", r.URL.Path)

		json.NewEncoder(w).Encode(struct {
			Status  string         `json:"status"`
			Message string         `json:"message"`
			Data    *IncidentStats `json:"data"`
		}{
			Status: "success",
			Data: &IncidentStats{
				TotalCount:     50,
				ActiveCount:    15,
				RecentCount:    50,
				RecentResolved: 30,
				BySeverity: map[string]int{
					"critical": 8,
					"warning":  5,
					"info":     2,
				},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	stats, err := client.Incidents.GetIncidentStats(context.Background())

	require.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 15, stats.ActiveCount)
	assert.Equal(t, int64(30), stats.RecentResolved)
	assert.Equal(t, 8, stats.BySeverity["critical"])
}

func TestIncidentsService_ResolveIncident(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)

		var req UpdateIncidentRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, IncidentStatusResolved, req.Status)

		json.NewEncoder(w).Encode(struct {
			Status  string    `json:"status"`
			Message string    `json:"message"`
			Data    *Incident `json:"data"`
		}{
			Status: "success",
			Data: &Incident{
				GormModel: GormModel{ID: 123},
				Status:    IncidentStatusResolved,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, err := client.Incidents.ResolveIncident(context.Background(), 123)

	require.NoError(t, err)
	assert.Equal(t, IncidentStatusResolved, result.Status)
}

func TestIncidentsService_AcknowledgeIncident(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)

		var req UpdateIncidentRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, IncidentStatusAcknowledged, req.Status)

		json.NewEncoder(w).Encode(struct {
			Status  string    `json:"status"`
			Message string    `json:"message"`
			Data    *Incident `json:"data"`
		}{
			Status: "success",
			Data: &Incident{
				GormModel: GormModel{ID: 456},
				Status:    IncidentStatusAcknowledged,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, err := client.Incidents.AcknowledgeIncident(context.Background(), 456)

	require.NoError(t, err)
	assert.Equal(t, IncidentStatusAcknowledged, result.Status)
}

func TestIncidentsService_CreateIncidentFromAlert(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CreateIncidentRequest
		json.NewDecoder(r.Body).Decode(&req)

		assert.Contains(t, req.Title, "Alert:")
		assert.NotNil(t, req.Metadata)
		assert.Equal(t, "alert", req.Metadata["source"])

		json.NewEncoder(w).Encode(struct {
			Status  string    `json:"status"`
			Message string    `json:"message"`
			Data    *Incident `json:"data"`
		}{
			Status: "success",
			Data: &Incident{
				GormModel: GormModel{ID: 1},
				Title:     req.Title,
				Source:    IncidentSourceAlert,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, err := client.Incidents.CreateIncidentFromAlert(
		context.Background(),
		123,
		456,
		"CPU Alert",
		IncidentSeverityCritical,
		uintPtr(789),
		"CPU usage exceeded 90%",
	)

	require.NoError(t, err)
	assert.Contains(t, result.Title, "Alert:")
	assert.Equal(t, IncidentSourceAlert, result.Source)
}

func TestIncidentsService_CreateIncidentFromProbe(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CreateIncidentRequest
		json.NewDecoder(r.Body).Decode(&req)

		assert.Contains(t, req.Title, "Probe Failure:")
		assert.Equal(t, IncidentSeverityCritical, req.Severity)

		json.NewEncoder(w).Encode(struct {
			Status  string    `json:"status"`
			Message string    `json:"message"`
			Data    *Incident `json:"data"`
		}{
			Status: "success",
			Data: &Incident{
				GormModel: GormModel{ID: 1},
				Title:     req.Title,
				Source:    IncidentSourceProbe,
				Severity:  IncidentSeverityCritical,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, err := client.Incidents.CreateIncidentFromProbe(
		context.Background(),
		123,
		456,
		"HTTP Probe",
		"Timeout after 30 seconds",
	)

	require.NoError(t, err)
	assert.Contains(t, result.Title, "Probe Failure:")
	assert.Equal(t, IncidentSeverityCritical, result.Severity)
}

func TestIncidentsService_ResolveIncidentFromAlert(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		if r.Method == "GET" {
			// List incidents
			json.NewEncoder(w).Encode(struct {
				Status  string                `json:"status"`
				Message string                `json:"message"`
				Data    *IncidentListResponse `json:"data"`
			}{
				Status: "success",
				Data: &IncidentListResponse{
					Incidents: []Incident{
						{
							GormModel: GormModel{ID: 1},
							Source:    IncidentSourceAlert,
							SourceID:  uintPtr(123),
							Status:    IncidentStatusActive,
						},
					},
					Total: 1,
				},
			})
		} else if r.Method == "PUT" {
			// Resolve incident
			json.NewEncoder(w).Encode(struct {
				Status  string    `json:"status"`
				Message string    `json:"message"`
				Data    *Incident `json:"data"`
			}{
				Status: "success",
				Data: &Incident{
					GormModel: GormModel{ID: 1},
					Status:    IncidentStatusResolved,
				},
			})
		}
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	err := client.Incidents.ResolveIncidentFromAlert(context.Background(), 123)

	require.NoError(t, err)
	assert.Equal(t, 2, callCount) // List + Resolve
}

func TestIncidentsService_ResolveIncidentFromProbe(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		if r.Method == "GET" {
			json.NewEncoder(w).Encode(struct {
				Status  string                `json:"status"`
				Message string                `json:"message"`
				Data    *IncidentListResponse `json:"data"`
			}{
				Status: "success",
				Data: &IncidentListResponse{
					Incidents: []Incident{
						{
							GormModel: GormModel{ID: 1},
							Source:    IncidentSourceProbe,
							SourceID:  uintPtr(456),
							Status:    IncidentStatusActive,
						},
					},
					Total: 1,
				},
			})
		} else if r.Method == "PUT" {
			json.NewEncoder(w).Encode(struct {
				Status  string    `json:"status"`
				Message string    `json:"message"`
				Data    *Incident `json:"data"`
			}{
				Status: "success",
				Data: &Incident{
					GormModel: GormModel{ID: 1},
					Status:    IncidentStatusResolved,
				},
			})
		}
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	err := client.Incidents.ResolveIncidentFromProbe(context.Background(), 456)

	require.NoError(t, err)
	assert.Equal(t, 2, callCount)
}

// Helper function
