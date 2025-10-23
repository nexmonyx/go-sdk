package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestControllersService_SubmitControllerHeartbeat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/controllers/test-controller/heartbeat", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Heartbeat received",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	req := &ControllerHeartbeatRequest{
		Version:  "1.0.0",
		Status:   "healthy",
		IsLeader: true,
	}
	err := client.Controllers.SubmitControllerHeartbeat(context.Background(), "test-controller", req)
	assert.NoError(t, err)
}

func TestControllersService_SubmitControllerHeartbeat_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Internal server error",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	req := &ControllerHeartbeatRequest{
		Version: "1.0.0",
		Status:  "healthy",
	}
	err := client.Controllers.SubmitControllerHeartbeat(context.Background(), "test-controller", req)
	assert.Error(t, err)
}

func TestControllersService_GetControllersSummary(t *testing.T) {
	now := time.Now()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/controllers/summary", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"total_controllers":   5,
				"active_controllers":  4,
				"healthy_controllers": 3,
				"controllers": map[string]interface{}{
					"alert-controller": map[string]interface{}{
						"name":           "alert-controller",
						"status":         "healthy",
						"version":        "1.0.0",
						"is_leader":      true,
						"last_seen":      now.Format(time.RFC3339),
						"instance_count": 2,
					},
				},
				"last_updated": now.Format(time.RFC3339),
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	summary, err := client.Controllers.GetControllersSummary(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, 5, summary.TotalControllers)
	assert.Equal(t, 4, summary.ActiveControllers)
	assert.Equal(t, 3, summary.HealthyControllers)
	assert.NotNil(t, summary.Controllers)
}

func TestControllersService_GetControllersSummary_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Failed to retrieve summary",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	summary, err := client.Controllers.GetControllersSummary(context.Background())
	assert.Error(t, err)
	assert.Nil(t, summary)
}

func TestControllersService_GetControllerStatus(t *testing.T) {
	now := time.Now()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/controllers/alert-controller/status", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"controller_name": "alert-controller",
				"status":          "healthy",
				"version":         "1.0.0",
				"is_leader":       true,
				"instances": []map[string]interface{}{
					{
						"hostname":        "pod-1",
						"version":         "1.0.0",
						"status":          "healthy",
						"is_leader":       true,
						"last_heartbeat":  now.Format(time.RFC3339),
						"processed_items": 1000,
						"error_count":     5,
					},
				},
				"metadata": map[string]interface{}{
					"region": "us-east-1",
				},
				"last_updated": now.Format(time.RFC3339),
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	status, err := client.Controllers.GetControllerStatus(context.Background(), "alert-controller")
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, "alert-controller", status.ControllerName)
	assert.Equal(t, "healthy", status.Status)
	assert.Equal(t, "1.0.0", status.Version)
	assert.True(t, status.IsLeader)
	assert.NotNil(t, status.Instances)
	assert.Len(t, status.Instances, 1)
}

func TestControllersService_GetControllerStatus_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Controller not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	status, err := client.Controllers.GetControllerStatus(context.Background(), "nonexistent")
	assert.Error(t, err)
	assert.Nil(t, status)
}

func TestControllersService_DeleteController(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/controllers/old-controller", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Controller deleted",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	err := client.Controllers.DeleteController(context.Background(), "old-controller")
	assert.NoError(t, err)
}

func TestControllersService_DeleteController_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Admin access required",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	err := client.Controllers.DeleteController(context.Background(), "some-controller")
	assert.Error(t, err)
}

func TestControllersService_GetAlertControllerStatus(t *testing.T) {
	now := time.Now()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/controllers/alerts/status", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"service":               "alert-controller",
				"version":               "1.0.0",
				"status":                "healthy",
				"is_leader":             true,
				"active_rules":          25,
				"active_instances":      150,
				"processed_alerts":      10000,
				"notifications_sent":    5000,
				"last_processed_at":     now.Format(time.RFC3339),
				"queue_depth":           50,
				"processing_rate":       125.5,
				"error_count":           10,
				"health_checks_passed":  95,
				"health_checks_failed":  5,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	status, err := client.Controllers.GetAlertControllerStatus(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, "alert-controller", status.Service)
	assert.Equal(t, "1.0.0", status.Version)
	assert.Equal(t, "healthy", status.Status)
	assert.True(t, status.IsLeader)
	assert.Equal(t, 25, status.ActiveRules)
	assert.Equal(t, 150, status.ActiveInstances)
	assert.Equal(t, int64(10000), status.ProcessedAlerts)
	assert.Equal(t, int64(5000), status.NotificationsSent)
}

func TestControllersService_GetAlertControllerStatus_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Service unavailable",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	status, err := client.Controllers.GetAlertControllerStatus(context.Background())
	assert.Error(t, err)
	assert.Nil(t, status)
}

func TestControllersService_GetAlertControllerLeaderStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/controllers/alerts/leader/status", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"is_leader":    true,
				"leader_id":    "pod-1",
				"election_age": 3600,
				"renew_time":   time.Now().Format(time.RFC3339),
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	leaderStatus, err := client.Controllers.GetAlertControllerLeaderStatus(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, leaderStatus)
	assert.True(t, leaderStatus["is_leader"].(bool))
	assert.Equal(t, "pod-1", leaderStatus["leader_id"])
}

func TestControllersService_GetAlertControllerLeaderStatus_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Failed to get leader status",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	leaderStatus, err := client.Controllers.GetAlertControllerLeaderStatus(context.Background())
	assert.Error(t, err)
	assert.Nil(t, leaderStatus)
}

func TestControllersService_NetworkError(t *testing.T) {
	client, _ := NewClient(&Config{BaseURL: "http://invalid-server:9999"})

	// Test all methods with network error
	err := client.Controllers.SubmitControllerHeartbeat(context.Background(), "test", &ControllerHeartbeatRequest{})
	assert.Error(t, err)

	_, err = client.Controllers.GetControllersSummary(context.Background())
	assert.Error(t, err)

	_, err = client.Controllers.GetControllerStatus(context.Background(), "test")
	assert.Error(t, err)

	err = client.Controllers.DeleteController(context.Background(), "test")
	assert.Error(t, err)

	_, err = client.Controllers.GetAlertControllerStatus(context.Background())
	assert.Error(t, err)

	_, err = client.Controllers.GetAlertControllerLeaderStatus(context.Background())
	assert.Error(t, err)
}
