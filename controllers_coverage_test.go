package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestControllersService_SendHeartbeat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.URL.Path, "/v1/controllers/")
		assert.Contains(t, r.URL.Path, "/heartbeat")
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
		ControllerName: "test-controller",
		Status:         "healthy",
		Version:        "1.0.0",
	}
	err := client.Controllers.SendHeartbeat(context.Background(), req)
	assert.NoError(t, err)
}

func TestControllersService_SendHeartbeat_NoName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	req := &ControllerHeartbeatRequest{}
	err := client.Controllers.SendHeartbeat(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "controller name is required")
}

func TestControllersService_RegisterController(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/controllers/register", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Controller registered",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	err := client.Controllers.RegisterController(context.Background(), "controller-123", "v1.0.0")
	assert.NoError(t, err)
}

func TestControllersService_GetControllerStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/v1/controllers/")
		assert.Contains(t, r.URL.Path, "/status")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"controller_id": "controller-123",
				"status":        "healthy",
				"uptime":        3600,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	status, err := client.Controllers.GetControllerStatus(context.Background(), "controller-123")
	assert.NoError(t, err)
	assert.NotNil(t, status)
}

func TestControllersService_ListControllers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/controllers", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": []map[string]interface{}{
				{"controller_id": "controller-1", "status": "healthy"},
				{"controller_id": "controller-2", "status": "degraded"},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	controllers, err := client.Controllers.ListControllers(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, controllers)
}

func TestControllersService_UpdateControllerStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Contains(t, r.URL.Path, "/v1/controllers/")
		assert.Contains(t, r.URL.Path, "/status")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Status updated",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	err := client.Controllers.UpdateControllerStatus(context.Background(), "controller-123", "degraded")
	assert.NoError(t, err)
}

func TestControllersService_DeregisterController(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Contains(t, r.URL.Path, "/v1/controllers/")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Controller deregistered",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	err := client.Controllers.DeregisterController(context.Background(), "controller-123")
	assert.NoError(t, err)
}

func TestControllersService_Errors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Internal error",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	err := client.Controllers.SendHeartbeat(context.Background(), &ControllerHeartbeatRequest{ControllerName: "test"})
	assert.Error(t, err)

	err = client.Controllers.RegisterController(context.Background(), "controller-123", "v1.0.0")
	assert.Error(t, err)

	_, err = client.Controllers.GetControllerStatus(context.Background(), "controller-123")
	assert.Error(t, err)

	_, err = client.Controllers.ListControllers(context.Background())
	assert.Error(t, err)

	err = client.Controllers.UpdateControllerStatus(context.Background(), "controller-123", "degraded")
	assert.Error(t, err)

	err = client.Controllers.DeregisterController(context.Background(), "controller-123")
	assert.Error(t, err)
}

func TestControllersService_NetworkError(t *testing.T) {
	client, _ := NewClient(&Config{BaseURL: "http://invalid-server:9999"})

	err := client.Controllers.SendHeartbeat(context.Background(), &ControllerHeartbeatRequest{ControllerName: "test"})
	assert.Error(t, err)

	err = client.Controllers.RegisterController(context.Background(), "controller-123", "v1.0.0")
	assert.Error(t, err)

	_, err = client.Controllers.GetControllerStatus(context.Background(), "controller-123")
	assert.Error(t, err)
}
