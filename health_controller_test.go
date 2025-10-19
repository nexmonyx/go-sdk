package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthService_GetAllControllerHealthStatus_Success(t *testing.T) {
	// Setup mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/health/controllers/status" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			response := StandardResponse{
				Status: "success",
				Data: &ControllerHealthStatusResponse{
					Controllers: map[string]ControllerStatus{
						"org-management-controller": {
							Status:      "healthy",
							Message:     "Controller is healthy",
							Details:     map[string]string{"endpoint": "http://localhost:8084/health"},
							LastUpdated: "2025-10-19T05:30:00Z",
							Duration:    "125ms",
						},
					},
					Total:     1,
					Timestamp: "2025-10-19T05:30:00Z",
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL: mockServer.URL,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test GetAllControllerHealthStatus
	status, err := client.Health.GetAllControllerHealthStatus(context.Background())
	if err != nil {
		t.Fatalf("Failed to get controller health status: %v", err)
	}

	if status == nil {
		t.Fatal("Expected status response, got nil")
	}

	if status.Total != 1 {
		t.Errorf("Expected Total=1, got %d", status.Total)
	}

	if len(status.Controllers) != 1 {
		t.Errorf("Expected 1 controller, got %d", len(status.Controllers))
	}

	controller, ok := status.Controllers["org-management-controller"]
	if !ok {
		t.Fatal("Expected org-management-controller in response")
	}

	if controller.Status != "healthy" {
		t.Errorf("Expected status='healthy', got '%s'", controller.Status)
	}
}

func TestHealthService_GetAllControllerHealthStatus_Error(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status":"error","message":"Internal server error"}`))
	}))
	defer mockServer.Close()

	client, err := NewClient(&Config{
		BaseURL: mockServer.URL,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.Health.GetAllControllerHealthStatus(context.Background())
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestHealthService_GetControllerHealthStatus_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/health/controllers/org-management-controller/status" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			response := StandardResponse{
				Status: "success",
				Data: &ControllerHealthDetailResponse{
					ControllerName: "org-management-controller",
					Status:         "healthy",
					Message:        "Controller is healthy",
					Details: map[string]string{
						"endpoint":     "http://localhost:8084/health",
						"status_code":  "200",
						"response_time_ms": "125",
					},
					LastUpdated:    "2025-10-19T05:30:00Z",
					Duration:       "125ms",
					ResponseTimeMs: 125,
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	client, err := NewClient(&Config{
		BaseURL: mockServer.URL,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test GetControllerHealthStatus
	detail, err := client.Health.GetControllerHealthStatus(context.Background(), "org-management-controller")
	if err != nil {
		t.Fatalf("Failed to get controller health status: %v", err)
	}

	if detail == nil {
		t.Fatal("Expected detail response, got nil")
	}

	if detail.ControllerName != "org-management-controller" {
		t.Errorf("Expected ControllerName='org-management-controller', got '%s'", detail.ControllerName)
	}

	if detail.Status != "healthy" {
		t.Errorf("Expected status='healthy', got '%s'", detail.Status)
	}

	if detail.ResponseTimeMs != 125 {
		t.Errorf("Expected ResponseTimeMs=125, got %d", detail.ResponseTimeMs)
	}
}

func TestHealthService_GetControllerHealthStatus_WithName(t *testing.T) {
	testControllerName := "test-controller"
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/v1/health/controllers/" + testControllerName + "/status"
		if r.URL.Path == expectedPath && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			response := StandardResponse{
				Status: "success",
				Data: &ControllerHealthDetailResponse{
					ControllerName: testControllerName,
					Status:         "warning",
					Message:        "Controller has warnings",
					Details:        map[string]string{"reason": "high_response_time"},
					LastUpdated:    "2025-10-19T05:30:00Z",
					Duration:       "5200ms",
					ResponseTimeMs: 5200,
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	client, err := NewClient(&Config{
		BaseURL: mockServer.URL,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	detail, err := client.Health.GetControllerHealthStatus(context.Background(), testControllerName)
	if err != nil {
		t.Fatalf("Failed to get controller health status: %v", err)
	}

	if detail.Status != "warning" {
		t.Errorf("Expected status='warning', got '%s'", detail.Status)
	}
}

func TestHealthService_GetControllerHealthStatus_NotFound(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"status":"error","message":"Controller not found"}`))
	}))
	defer mockServer.Close()

	client, err := NewClient(&Config{
		BaseURL: mockServer.URL,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.Health.GetControllerHealthStatus(context.Background(), "nonexistent-controller")
	if err == nil {
		t.Fatal("Expected error for nonexistent controller, got nil")
	}
}

func TestControllerHealthStatusResponse_Structure(t *testing.T) {
	// Verify response structure
	resp := &ControllerHealthStatusResponse{
		Controllers: map[string]ControllerStatus{
			"test": {
				Status:      "healthy",
				Message:     "OK",
				Details:     map[string]string{"key": "value"},
				LastUpdated: "2025-10-19T05:30:00Z",
				Duration:    "100ms",
			},
		},
		Total:     1,
		Timestamp: "2025-10-19T05:30:00Z",
	}

	// Marshal to JSON to verify structure
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	// Unmarshal to verify JSON format
	var unmarshalled ControllerHealthStatusResponse
	err = json.Unmarshal(data, &unmarshalled)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if unmarshalled.Total != resp.Total {
		t.Errorf("Total mismatch after marshal/unmarshal")
	}
}

func TestControllerHealthDetailResponse_Structure(t *testing.T) {
	// Verify detail response structure
	detail := &ControllerHealthDetailResponse{
		ControllerName: "test-controller",
		Status:         "healthy",
		Message:        "All systems operational",
		Details: map[string]string{
			"cpu":    "45%",
			"memory": "60%",
		},
		LastUpdated:    "2025-10-19T05:30:00Z",
		Duration:       "125ms",
		ResponseTimeMs: 125,
	}

	// Marshal to JSON to verify structure
	data, err := json.Marshal(detail)
	if err != nil {
		t.Fatalf("Failed to marshal detail response: %v", err)
	}

	// Unmarshal to verify JSON format
	var unmarshalled ControllerHealthDetailResponse
	err = json.Unmarshal(data, &unmarshalled)
	if err != nil {
		t.Fatalf("Failed to unmarshal detail response: %v", err)
	}

	if unmarshalled.ResponseTimeMs != detail.ResponseTimeMs {
		t.Errorf("ResponseTimeMs mismatch after marshal/unmarshal")
	}

	if unmarshalled.ControllerName != detail.ControllerName {
		t.Errorf("ControllerName mismatch after marshal/unmarshal")
	}
}
