package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Comprehensive coverage tests for probe_alerts.go (currently at 0%)

// ToQuery tests for ProbeAlertListOptions

func TestProbeAlertListOptions_ToQuery_AllFilters(t *testing.T) {
	opts := &ProbeAlertListOptions{
		ListOptions: ListOptions{
			Page:  2,
			Limit: 50,
		},
		Status:  "active",
		ProbeID: 123,
	}

	query := opts.ToQuery()
	assert.Equal(t, "active", query["status"])
	assert.Equal(t, "123", query["probe_id"])
	assert.Equal(t, "2", query["page"])
	assert.Equal(t, "50", query["limit"])
}

func TestProbeAlertListOptions_ToQuery_MinimalFilters(t *testing.T) {
	opts := &ProbeAlertListOptions{
		ListOptions: ListOptions{
			Page: 1,
		},
		Status:  "",
		ProbeID: 0,
	}

	query := opts.ToQuery()
	assert.Equal(t, "1", query["page"])
	// Empty values should not be in query
	assert.Empty(t, query["status"])
	assert.Empty(t, query["probe_id"])
}

// List tests

func TestProbeAlertsService_List_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/probe-alerts", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Probe alerts retrieved",
			"data": map[string]interface{}{
				"alerts": []map[string]interface{}{
					{
						"id":        1,
						"probe_id":  100,
						"name":      "High Response Time",
						"status":    "active",
						"message":   "Response time exceeded threshold",
					},
				},
				"pagination": map[string]interface{}{
					"page":        1,
					"limit":       25,
					"total_items": 1,
					"total_pages": 1,
				},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	alerts, meta, err := client.ProbeAlerts.List(context.Background(), &ProbeAlertListOptions{
		ListOptions: ListOptions{Page: 1, Limit: 25},
		Status:      "active",
	})
	assert.NoError(t, err)
	assert.NotNil(t, alerts)
	assert.NotNil(t, meta)
	assert.Len(t, alerts, 1)
	assert.Equal(t, uint(1), alerts[0].ID)
	assert.Equal(t, "active", alerts[0].Status)
}

func TestProbeAlertsService_List_NilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/probe-alerts", r.URL.Path)

		// Verify no query parameters when opts is nil
		assert.Empty(t, r.URL.Query().Get("page"))
		assert.Empty(t, r.URL.Query().Get("status"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Probe alerts retrieved",
			"data": map[string]interface{}{
				"alerts":     []*ProbeAlert{},
				"pagination": PaginationMeta{Page: 1, Limit: 25},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	alerts, meta, err := client.ProbeAlerts.List(context.Background(), nil)
	assert.NoError(t, err)
	assert.NotNil(t, alerts)
	assert.NotNil(t, meta)
}

func TestProbeAlertsService_List_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Internal server error",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	alerts, meta, err := client.ProbeAlerts.List(context.Background(), nil)
	assert.Error(t, err)
	assert.Nil(t, alerts)
	assert.Nil(t, meta)
}

// Get tests

func TestProbeAlertsService_Get_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/probe-alerts/1", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Probe alert retrieved",
			"data": map[string]interface{}{
				"alert": map[string]interface{}{
					"id":        1,
					"probe_id":  100,
					"name":      "High Response Time",
					"status":    "active",
					"message":   "Response time exceeded threshold",
				},
				"probe": map[string]interface{}{
					"id":   100,
					"uuid": "probe-uuid-123",
					"name": "API Health Check",
					"type": "http",
				},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	alert, err := client.ProbeAlerts.Get(context.Background(), 1)
	assert.NoError(t, err)
	assert.NotNil(t, alert)
	assert.Equal(t, uint(1), alert.ID)
	assert.Equal(t, "active", alert.Status)
}

func TestProbeAlertsService_Get_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Probe alert not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	alert, err := client.ProbeAlerts.Get(context.Background(), 999)
	assert.Error(t, err)
	assert.Nil(t, alert)
}

// Acknowledge tests

func TestProbeAlertsService_Acknowledge_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/probe-alerts/1/acknowledge", r.URL.Path)

		var body map[string]string
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.NoError(t, err)
		assert.Equal(t, "Investigating the issue", body["note"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Probe alert acknowledged",
			"data": map[string]interface{}{
				"alert": map[string]interface{}{
					"id":     1,
					"status": "acknowledged",
				},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	alert, err := client.ProbeAlerts.Acknowledge(context.Background(), 1, "Investigating the issue")
	assert.NoError(t, err)
	assert.NotNil(t, alert)
	assert.Equal(t, "acknowledged", alert.Status)
}

func TestProbeAlertsService_Acknowledge_EmptyNote(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)

		var body map[string]string
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.NoError(t, err)
		// Empty note should not be in body
		_, hasNote := body["note"]
		assert.False(t, hasNote)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Probe alert acknowledged",
			"data": map[string]interface{}{
				"alert": map[string]interface{}{
					"id":     1,
					"status": "acknowledged",
				},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	alert, err := client.ProbeAlerts.Acknowledge(context.Background(), 1, "")
	assert.NoError(t, err)
	assert.NotNil(t, alert)
}

func TestProbeAlertsService_Acknowledge_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Probe alert not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	alert, err := client.ProbeAlerts.Acknowledge(context.Background(), 999, "Note")
	assert.Error(t, err)
	assert.Nil(t, alert)
}

// Resolve tests

func TestProbeAlertsService_Resolve_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/probe-alerts/1/resolve", r.URL.Path)

		var body map[string]string
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.NoError(t, err)
		assert.Equal(t, "Issue fixed by restarting service", body["resolution"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Probe alert resolved",
			"data": map[string]interface{}{
				"alert": map[string]interface{}{
					"id":     1,
					"status": "resolved",
				},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	alert, err := client.ProbeAlerts.Resolve(context.Background(), 1, "Issue fixed by restarting service")
	assert.NoError(t, err)
	assert.NotNil(t, alert)
	assert.Equal(t, "resolved", alert.Status)
}

func TestProbeAlertsService_Resolve_EmptyResolution(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)

		var body map[string]string
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.NoError(t, err)
		// Empty resolution should not be in body
		_, hasResolution := body["resolution"]
		assert.False(t, hasResolution)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Probe alert resolved",
			"data": map[string]interface{}{
				"alert": map[string]interface{}{
					"id":     1,
					"status": "resolved",
				},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	alert, err := client.ProbeAlerts.Resolve(context.Background(), 1, "")
	assert.NoError(t, err)
	assert.NotNil(t, alert)
}

func TestProbeAlertsService_Resolve_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Probe alert not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	alert, err := client.ProbeAlerts.Resolve(context.Background(), 999, "Resolution")
	assert.Error(t, err)
	assert.Nil(t, alert)
}

// ToQuery tests for AdminProbeAlertListOptions

func TestAdminProbeAlertListOptions_ToQuery_AllFilters(t *testing.T) {
	opts := &AdminProbeAlertListOptions{
		ListOptions: ListOptions{
			Page:  3,
			Limit: 100,
		},
		Status:         "resolved",
		OrganizationID: 456,
	}

	query := opts.ToQuery()
	assert.Equal(t, "resolved", query["status"])
	assert.Equal(t, "456", query["organization_id"])
	assert.Equal(t, "3", query["page"])
	assert.Equal(t, "100", query["limit"])
}

func TestAdminProbeAlertListOptions_ToQuery_MinimalFilters(t *testing.T) {
	opts := &AdminProbeAlertListOptions{
		ListOptions: ListOptions{
			Page: 1,
		},
		Status:         "",
		OrganizationID: 0,
	}

	query := opts.ToQuery()
	assert.Equal(t, "1", query["page"])
	// Empty values should not be in query
	assert.Empty(t, query["status"])
	assert.Empty(t, query["organization_id"])
}

// ListAdmin tests

func TestProbeAlertsService_ListAdmin_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/admin/probe-alerts", r.URL.Path)

		q := r.URL.Query()
		assert.Equal(t, "active", q.Get("status"))
		assert.Equal(t, "100", q.Get("organization_id"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Admin probe alerts retrieved",
			"data": map[string]interface{}{
				"alerts": []map[string]interface{}{
					{
						"id":                1,
						"probe_id":          100,
						"name":              "High Response Time",
						"status":            "active",
						"organization_id":   100,
						"organization_name": "Acme Corp",
						"probe_name":        "API Health Check",
						"probe_type":        "http",
						"probe_target":      "https://api.example.com",
					},
				},
				"pagination": map[string]interface{}{
					"page":        1,
					"limit":       25,
					"total_items": 1,
					"total_pages": 1,
				},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	alerts, meta, err := client.ProbeAlerts.ListAdmin(context.Background(), &AdminProbeAlertListOptions{
		ListOptions:    ListOptions{Page: 1, Limit: 25},
		Status:         "active",
		OrganizationID: 100,
	})
	assert.NoError(t, err)
	assert.NotNil(t, alerts)
	assert.NotNil(t, meta)
	assert.Len(t, alerts, 1)
	assert.Equal(t, uint(1), alerts[0].ID)
	assert.Equal(t, "Acme Corp", alerts[0].OrganizationName)
}

func TestProbeAlertsService_ListAdmin_NilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/admin/probe-alerts", r.URL.Path)

		// Verify no query parameters when opts is nil
		assert.Empty(t, r.URL.Query().Get("page"))
		assert.Empty(t, r.URL.Query().Get("status"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Admin probe alerts retrieved",
			"data": map[string]interface{}{
				"alerts":     []*AdminProbeAlert{},
				"pagination": PaginationMeta{Page: 1, Limit: 25},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	alerts, meta, err := client.ProbeAlerts.ListAdmin(context.Background(), nil)
	assert.NoError(t, err)
	assert.NotNil(t, alerts)
	assert.NotNil(t, meta)
}

func TestProbeAlertsService_ListAdmin_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Insufficient permissions",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	alerts, meta, err := client.ProbeAlerts.ListAdmin(context.Background(), nil)
	assert.Error(t, err)
	assert.Nil(t, alerts)
	assert.Nil(t, meta)
}
