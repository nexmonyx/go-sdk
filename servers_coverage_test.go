package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Coverage tests for servers.go focusing on debug paths and low-coverage methods

// Heartbeat tests - coverage improvement from 33.3% to 100%

func TestServersService_Heartbeat_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/heartbeat", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Heartbeat received",
		})
	}))
	defer server.Close()

	// Test without debug mode
	client, _ := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{ServerUUID: "test-uuid", ServerSecret: "test-secret"},
	})
	err := client.Servers.Heartbeat(context.Background())
	assert.NoError(t, err)
}

func TestServersService_Heartbeat_WithDebug(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Heartbeat received",
		})
	}))
	defer server.Close()

	// Test with debug mode to cover debug logging paths
	client, _ := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{ServerUUID: "test-uuid", ServerSecret: "test-secret"},
		Debug:   true,
	})
	err := client.Servers.Heartbeat(context.Background())
	assert.NoError(t, err)
}

func TestServersService_Heartbeat_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Server error",
		})
	}))
	defer server.Close()

	// Test error path without debug
	client, _ := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{ServerUUID: "test-uuid", ServerSecret: "test-secret"},
	})
	err := client.Servers.Heartbeat(context.Background())
	assert.Error(t, err)
}

func TestServersService_Heartbeat_ErrorWithDebug(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Bad request",
		})
	}))
	defer server.Close()

	// Test error path with debug mode
	client, _ := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{ServerUUID: "test-uuid", ServerSecret: "test-secret"},
		Debug:   true,
	})
	err := client.Servers.Heartbeat(context.Background())
	assert.Error(t, err)
}

// HeartbeatWithVersion tests - coverage improvement from 36.8% to 100%

func TestServersService_HeartbeatWithVersion_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/heartbeat", r.URL.Path)

		var body map[string]string
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.NoError(t, err)
		assert.Equal(t, "v1.2.3", body["agent_version"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Heartbeat with version received",
		})
	}))
	defer server.Close()

	// Test without debug mode
	client, _ := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{ServerUUID: "test-uuid", ServerSecret: "test-secret"},
	})
	err := client.Servers.HeartbeatWithVersion(context.Background(), "v1.2.3")
	assert.NoError(t, err)
}

func TestServersService_HeartbeatWithVersion_WithDebug(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Heartbeat received",
		})
	}))
	defer server.Close()

	// Test with debug mode
	client, _ := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{ServerUUID: "test-uuid", ServerSecret: "test-secret"},
		Debug:   true,
	})
	err := client.Servers.HeartbeatWithVersion(context.Background(), "v2.0.0")
	assert.NoError(t, err)
}

func TestServersService_HeartbeatWithVersion_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Unauthorized",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{ServerUUID: "test-uuid", ServerSecret: "test-secret"},
	})
	err := client.Servers.HeartbeatWithVersion(context.Background(), "v1.0.0")
	assert.Error(t, err)
}

func TestServersService_HeartbeatWithVersion_ErrorWithDebug(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Forbidden",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{ServerUUID: "test-uuid", ServerSecret: "test-secret"},
		Debug:   true,
	})
	err := client.Servers.HeartbeatWithVersion(context.Background(), "v1.0.0")
	assert.Error(t, err)
}

// UpdateDetails tests - coverage improvement from 18.9% to 100%

func TestServersService_UpdateDetails_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/server/test-uuid/details", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id":          1,
				"server_uuid": "test-uuid",
				"hostname":    "updated-hostname",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{ServerUUID: "test-uuid", ServerSecret: "test-secret"},
	})

	req := &ServerDetailsUpdateRequest{
		Hostname:     "updated-hostname",
		OS:           "Ubuntu",
		OSVersion:    "22.04",
		OSArch:       "x86_64",
		CPUModel:     "Intel Xeon",
		CPUCores:     8,
		MemoryTotal:  16777216000,
		StorageTotal: 1099511627776,
	}

	result, err := client.Servers.UpdateDetails(context.Background(), "test-uuid", req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "updated-hostname", result.Hostname)
}

func TestServersService_UpdateDetails_WithDebug(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id":          1,
				"server_uuid": "test-uuid",
				"hostname":    "test-server",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{ServerUUID: "test-uuid", ServerSecret: "test-secret"},
		Debug:   true,
	})

	req := &ServerDetailsUpdateRequest{
		Hostname:     "test-server",
		OS:           "Ubuntu",
		OSVersion:    "22.04",
		OSArch:       "x86_64",
		CPUModel:     "Intel Xeon",
		CPUCores:     16,
		MemoryTotal:  33554432000,
		StorageTotal: 2199023255552,
	}

	result, err := client.Servers.UpdateDetails(context.Background(), "test-uuid", req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestServersService_UpdateDetails_WithHardwareDetails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id":          1,
				"server_uuid": "test-uuid",
				"hostname":    "test-server",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{ServerUUID: "test-uuid", ServerSecret: "test-secret"},
		Debug:   true,
	})

	req := &ServerDetailsUpdateRequest{
		Hostname:     "test-server",
		OS:           "Ubuntu",
		OSVersion:    "22.04",
		OSArch:       "x86_64",
		CPUModel:     "Intel Xeon",
		CPUCores:     8,
		MemoryTotal:  16777216000,
		StorageTotal: 1099511627776,
		Hardware: &HardwareDetails{
			CPU: []ServerCPUInfo{
				{
					Manufacturer:  "Intel",
					ModelName:     "Xeon E5-2680",
					PhysicalCores: 8,
					LogicalCores:  16,
					L3Cache:       20480,
				},
			},
			Memory: &ServerMemoryInfo{
				TotalSize:  16777216000,
				MemoryType: "DDR4",
				Speed:      2400,
			},
			Network: []ServerNetworkInterfaceInfo{
				{
					Name:         "eth0",
					HardwareAddr: "00:11:22:33:44:55",
					SpeedMbps:    1000,
				},
			},
			Disks: []ServerDiskInfo{
				{
					Device:    "/dev/sda",
					DiskModel: "Samsung SSD",
					Type:      "ssd",
					Size:      1099511627776,
				},
			},
		},
	}

	result, err := client.Servers.UpdateDetails(context.Background(), "test-uuid", req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestServersService_UpdateDetails_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Server not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{ServerUUID: "test-uuid", ServerSecret: "test-secret"},
	})

	req := &ServerDetailsUpdateRequest{
		Hostname: "test-server",
	}

	result, err := client.Servers.UpdateDetails(context.Background(), "nonexistent-uuid", req)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestServersService_UpdateDetails_ErrorWithDebug(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Invalid request",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{ServerUUID: "test-uuid", ServerSecret: "test-secret"},
		Debug:   true,
	})

	req := &ServerDetailsUpdateRequest{
		Hostname: "test-server",
	}

	result, err := client.Servers.UpdateDetails(context.Background(), "test-uuid", req)
	assert.Error(t, err)
	assert.Nil(t, result)
}

// UpdateInfo tests - coverage improvement from 27.8% to 100%

func TestServersService_UpdateInfo_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/server/test-uuid/info", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id":          1,
				"server_uuid": "test-uuid",
				"hostname":    "info-updated",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{ServerUUID: "test-uuid", ServerSecret: "test-secret"},
	})

	req := &ServerDetailsUpdateRequest{
		Hostname:     "info-updated",
		OS:           "Debian",
		OSVersion:    "11",
		OSArch:       "amd64",
		CPUModel:     "AMD EPYC",
		CPUCores:     12,
		MemoryTotal:  32212254720,
		StorageTotal: 2199023255552,
	}

	result, err := client.Servers.UpdateInfo(context.Background(), "test-uuid", req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "info-updated", result.Hostname)
}

func TestServersService_UpdateInfo_WithDebug(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id":          1,
				"server_uuid": "test-uuid",
				"hostname":    "debug-server",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{ServerUUID: "test-uuid", ServerSecret: "test-secret"},
		Debug:   true,
	})

	req := &ServerDetailsUpdateRequest{
		Hostname:     "debug-server",
		OS:           "CentOS",
		OSVersion:    "8",
		OSArch:       "x86_64",
		CPUModel:     "Intel Core",
		CPUCores:     4,
		MemoryTotal:  8589934592,
		StorageTotal: 549755813888,
	}

	result, err := client.Servers.UpdateInfo(context.Background(), "test-uuid", req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestServersService_UpdateInfo_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Unauthorized",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{ServerUUID: "wrong-uuid", ServerSecret: "wrong-secret"},
	})

	req := &ServerDetailsUpdateRequest{
		Hostname: "test-server",
	}

	result, err := client.Servers.UpdateInfo(context.Background(), "test-uuid", req)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestServersService_UpdateInfo_ErrorWithDebug(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Forbidden",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{ServerUUID: "test-uuid", ServerSecret: "test-secret"},
		Debug:   true,
	})

	req := &ServerDetailsUpdateRequest{
		Hostname: "test-server",
	}

	result, err := client.Servers.UpdateInfo(context.Background(), "test-uuid", req)
	assert.Error(t, err)
	assert.Nil(t, result)
}

// Additional edge case tests for existing methods

func TestServersService_GetByUUID_TypeAssertion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Return response that won't unmarshal to Server
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   "not-a-server",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, err := client.Servers.GetByUUID(context.Background(), "test-uuid")
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestServersService_Create_TypeAssertion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   12345,
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, err := client.Servers.Create(context.Background(), &Server{Hostname: "test"})
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestServersService_List_NilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v2/servers", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   []*Server{},
			"meta": PaginationMeta{
				Page:       1,
				Limit:      25,
				TotalItems: 0,
				TotalPages: 0,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	servers, meta, err := client.Servers.List(context.Background(), nil)
	assert.NoError(t, err)
	assert.NotNil(t, servers)
	assert.NotNil(t, meta)
}

func TestServersService_UpdateServer_TypeAssertion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   []string{"not", "a", "server"},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, err := client.Servers.UpdateServer(context.Background(), "test-uuid", &ServerUpdateRequest{})
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestServersService_UpdateDetails_TypeAssertion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   123456, // Invalid type
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, err := client.Servers.UpdateDetails(context.Background(), "test-uuid", &ServerDetailsUpdateRequest{})
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestServersService_UpdateInfo_TypeAssertion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   false,
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, err := client.Servers.UpdateInfo(context.Background(), "test-uuid", &ServerDetailsUpdateRequest{})
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestServersService_GetDetails_TypeAssertion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   nil,
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, err := client.Servers.GetDetails(context.Background(), "test-uuid")
	assert.Error(t, err)
	assert.Nil(t, result)
}
