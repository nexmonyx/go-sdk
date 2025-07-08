package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProbeAlertsService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/probe-alerts" {
			t.Errorf("Expected /v1/probe-alerts, got %s", r.URL.Path)
		}

		// Check authentication
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected Bearer test-token, got %s", r.Header.Get("Authorization"))
		}

		// Check query parameters
		query := r.URL.Query()
		if query.Get("status") != "active" {
			t.Errorf("Expected status=active, got %s", query.Get("status"))
		}
		if query.Get("probe_id") != "123" {
			t.Errorf("Expected probe_id=123, got %s", query.Get("probe_id"))
		}

		response := map[string]interface{}{
			"status":  "success",
			"message": "Probe alerts retrieved successfully",
			"data": map[string]interface{}{
				"alerts": []ProbeAlert{
					{
						ID:      1,
						ProbeID: 123,
						Name:    "High Response Time",
						Status:  "active",
						Message: "Response time exceeded 500ms",
						Conditions: ProbeAlertConditions{
							FailureThreshold:  3,
							RecoveryThreshold: 2,
						},
						TriggeredAt:      &CustomTime{Time: time.Now()},
						NotificationSent: true,
						CreatedAt:        &CustomTime{Time: time.Now()},
						UpdatedAt:        &CustomTime{Time: time.Now()},
					},
				},
				"pagination": &PaginationMeta{
					Page:       1,
					Limit:      10,
					TotalItems: 1,
					TotalPages: 1,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	opts := &ProbeAlertListOptions{
		Status:  "active",
		ProbeID: 123,
		ListOptions: ListOptions{
			Page:  1,
			Limit: 10,
		},
	}

	alerts, meta, err := client.ProbeAlerts.List(context.Background(), opts)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(alerts) != 1 {
		t.Errorf("Expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].ID != 1 {
		t.Errorf("Expected alert ID 1, got %d", alerts[0].ID)
	}
	if alerts[0].Status != "active" {
		t.Errorf("Expected status 'active', got '%s'", alerts[0].Status)
	}
	if meta.TotalItems != 1 {
		t.Errorf("Expected total items 1, got %d", meta.TotalItems)
	}
}

func TestProbeAlertsService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/probe-alerts/1" {
			t.Errorf("Expected /v1/probe-alerts/1, got %s", r.URL.Path)
		}

		response := map[string]interface{}{
			"status":  "success",
			"message": "Probe alert retrieved successfully",
			"data": map[string]interface{}{
				"alert": ProbeAlert{
					ID:      1,
					ProbeID: 123,
					Name:    "High Response Time",
					Status:  "active",
					Message: "Response time exceeded 500ms",
					Conditions: ProbeAlertConditions{
						FailureThreshold:  3,
						RecoveryThreshold: 2,
					},
					TriggeredAt:      &CustomTime{Time: time.Now()},
					NotificationSent: true,
					CreatedAt:        &CustomTime{Time: time.Now()},
					UpdatedAt:        &CustomTime{Time: time.Now()},
				},
				"probe": map[string]interface{}{
					"id":   123,
					"uuid": "probe-uuid-123",
					"name": "Example API Health Check",
					"type": "http",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	alert, err := client.ProbeAlerts.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if alert.ID != 1 {
		t.Errorf("Expected alert ID 1, got %d", alert.ID)
	}
	if alert.Status != "active" {
		t.Errorf("Expected status 'active', got '%s'", alert.Status)
	}
}

func TestProbeAlertsService_Acknowledge(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/probe-alerts/1/acknowledge" {
			t.Errorf("Expected /v1/probe-alerts/1/acknowledge, got %s", r.URL.Path)
		}

		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["note"] != "Investigating the issue" {
			t.Errorf("Expected note 'Investigating the issue', got '%s'", body["note"])
		}

		response := map[string]interface{}{
			"status":  "success",
			"message": "Probe alert acknowledged",
			"data": map[string]interface{}{
				"alert": ProbeAlert{
					ID:             1,
					ProbeID:        123,
					Name:           "High Response Time",
					Status:         "acknowledged",
					Message:        "Response time exceeded 500ms",
					AcknowledgedBy: &[]uint{1}[0],
					AcknowledgedAt: &CustomTime{Time: time.Now()},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	alert, err := client.ProbeAlerts.Acknowledge(context.Background(), 1, "Investigating the issue")
	if err != nil {
		t.Fatalf("Acknowledge failed: %v", err)
	}

	if alert.Status != "acknowledged" {
		t.Errorf("Expected status 'acknowledged', got '%s'", alert.Status)
	}
	if alert.AcknowledgedBy == nil {
		t.Errorf("Expected AcknowledgedBy to be set")
	}
}

func TestProbeAlertsService_Resolve(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/probe-alerts/1/resolve" {
			t.Errorf("Expected /v1/probe-alerts/1/resolve, got %s", r.URL.Path)
		}

		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["resolution"] != "Server resources scaled up" {
			t.Errorf("Expected resolution 'Server resources scaled up', got '%s'", body["resolution"])
		}

		response := map[string]interface{}{
			"status":  "success",
			"message": "Probe alert resolved",
			"data": map[string]interface{}{
				"alert": ProbeAlert{
					ID:         1,
					ProbeID:    123,
					Name:       "High Response Time",
					Status:     "resolved",
					Message:    "Response time exceeded 500ms",
					ResolvedAt: &CustomTime{Time: time.Now()},
					Resolution: &[]string{"Server resources scaled up"}[0],
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	alert, err := client.ProbeAlerts.Resolve(context.Background(), 1, "Server resources scaled up")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if alert.Status != "resolved" {
		t.Errorf("Expected status 'resolved', got '%s'", alert.Status)
	}
	if alert.Resolution == nil || *alert.Resolution != "Server resources scaled up" {
		t.Errorf("Expected resolution to be set")
	}
}

func TestProbeAlertsService_ListAdmin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/admin/probe-alerts" {
			t.Errorf("Expected /v1/admin/probe-alerts, got %s", r.URL.Path)
		}

		// Check query parameters
		query := r.URL.Query()
		if query.Get("organization_id") != "456" {
			t.Errorf("Expected organization_id=456, got %s", query.Get("organization_id"))
		}

		response := map[string]interface{}{
			"status":  "success",
			"message": "Admin probe alerts retrieved successfully",
			"data": map[string]interface{}{
				"alerts": []AdminProbeAlert{
					{
						ProbeAlert: ProbeAlert{
							ID:      1,
							ProbeID: 123,
							Name:    "High Response Time",
							Status:  "active",
							Message: "Response time exceeded 500ms",
						},
						OrganizationName: "Example Corp",
						OrganizationID:   456,
						ProbeName:        "Example API Health Check",
						ProbeType:        "http",
						ProbeTarget:      "https://api.example.com/health",
					},
				},
				"pagination": &PaginationMeta{
					Page:       1,
					Limit:      10,
					TotalItems: 1,
					TotalPages: 1,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	opts := &AdminProbeAlertListOptions{
		OrganizationID: 456,
		ListOptions: ListOptions{
			Page:  1,
			Limit: 10,
		},
	}

	alerts, meta, err := client.ProbeAlerts.ListAdmin(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListAdmin failed: %v", err)
	}

	if len(alerts) != 1 {
		t.Errorf("Expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].OrganizationID != 456 {
		t.Errorf("Expected organization ID 456, got %d", alerts[0].OrganizationID)
	}
	if alerts[0].OrganizationName != "Example Corp" {
		t.Errorf("Expected organization name 'Example Corp', got '%s'", alerts[0].OrganizationName)
	}
	if meta.TotalItems != 1 {
		t.Errorf("Expected total items 1, got %d", meta.TotalItems)
	}
}