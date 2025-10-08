package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVMsService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v1/vms", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "10", r.URL.Query().Get("limit"))
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		response := struct {
			Data []VirtualMachine `json:"data"`
			Meta *PaginationMeta  `json:"meta"`
		}{
			Data: []VirtualMachine{
				{
					ID:             1,
					OrganizationID: 10,
					Name:           "web-server-01",
					Status:         "running",
					CPUCores:       4,
					MemoryMB:       8192,
					StorageGB:      100,
					IPAddress:      "192.168.1.100",
					OSType:         "linux",
					OSVersion:      "Ubuntu 22.04",
					CreatedAt:      CustomTime{Time: time.Now()},
					UpdatedAt:      CustomTime{Time: time.Now()},
				},
			},
			Meta: &PaginationMeta{
				Page:       1,
				PerPage:    10,
				TotalItems: 1,
				TotalPages: 1,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	vms, meta, err := client.VMs.List(context.Background(), &PaginationOptions{Page: 1, Limit: 10})
	require.NoError(t, err)
	assert.Len(t, vms, 1)
	assert.Equal(t, "web-server-01", vms[0].Name)
	assert.Equal(t, "running", vms[0].Status)
	assert.NotNil(t, meta)
	assert.Equal(t, 1, meta.TotalItems)
}

func TestVMsService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/vms", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		var reqBody VMConfiguration
		json.NewDecoder(r.Body).Decode(&reqBody)
		assert.Equal(t, "test-vm", reqBody.Name)
		assert.Equal(t, 2, reqBody.CPUCores)
		assert.Equal(t, 4096, reqBody.MemoryMB)

		response := struct {
			Data *VirtualMachine `json:"data"`
		}{
			Data: &VirtualMachine{
				ID:             1,
				OrganizationID: 10,
				Name:           reqBody.Name,
				Status:         "stopped",
				CPUCores:       reqBody.CPUCores,
				MemoryMB:       reqBody.MemoryMB,
				StorageGB:      reqBody.StorageGB,
				OSType:         reqBody.OSType,
				CreatedAt:      CustomTime{Time: time.Now()},
				UpdatedAt:      CustomTime{Time: time.Now()},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	config := &VMConfiguration{
		Name:      "test-vm",
		CPUCores:  2,
		MemoryMB:  4096,
		StorageGB: 50,
		OSType:    "linux",
	}

	vm, err := client.VMs.Create(context.Background(), config)
	require.NoError(t, err)
	assert.Equal(t, "test-vm", vm.Name)
	assert.Equal(t, "stopped", vm.Status)
	assert.Equal(t, 2, vm.CPUCores)
}

func TestVMsService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v1/vms/1", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		response := struct {
			Data *VirtualMachine `json:"data"`
		}{
			Data: &VirtualMachine{
				ID:             1,
				OrganizationID: 10,
				Name:           "web-server-01",
				Description:    "Production web server",
				Status:         "running",
				CPUCores:       4,
				MemoryMB:       8192,
				StorageGB:      100,
				IPAddress:      "192.168.1.100",
				MACAddress:     "00:16:3e:00:00:01",
				OSType:         "linux",
				OSVersion:      "Ubuntu 22.04",
				CreatedAt:      CustomTime{Time: time.Now()},
				UpdatedAt:      CustomTime{Time: time.Now()},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	vm, err := client.VMs.Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "web-server-01", vm.Name)
	assert.Equal(t, "running", vm.Status)
	assert.Equal(t, "192.168.1.100", vm.IPAddress)
}

func TestVMsService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/v2/organizations/10/virtual-machines/1", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	err = client.VMs.Delete(context.Background(), 10, 1)
	require.NoError(t, err)
}

func TestVMsService_Start(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v2/organizations/10/virtual-machines/1/start", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		response := struct {
			Data *VMOperation `json:"data"`
		}{
			Data: &VMOperation{
				ID:            1,
				VMID:          1,
				OperationType: "start",
				Status:        "in_progress",
				Progress:      10,
				Message:       "Starting virtual machine",
				CreatedAt:     CustomTime{Time: time.Now()},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	operation, err := client.VMs.Start(context.Background(), 10, 1)
	require.NoError(t, err)
	assert.Equal(t, "start", operation.OperationType)
	assert.Equal(t, "in_progress", operation.Status)
	assert.Equal(t, 10, operation.Progress)
}

func TestVMsService_Stop(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v2/organizations/10/virtual-machines/1/stop", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		var reqBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&reqBody)
		forceStop, _ := reqBody["force"].(bool)

		response := struct {
			Data *VMOperation `json:"data"`
		}{
			Data: &VMOperation{
				ID:            2,
				VMID:          1,
				OperationType: "stop",
				Status:        "in_progress",
				Progress:      0,
				Message:       map[bool]string{true: "Force stopping VM", false: "Gracefully stopping VM"}[forceStop],
				CreatedAt:     CustomTime{Time: time.Now()},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	// Test graceful stop
	operation, err := client.VMs.Stop(context.Background(), 10, 1, false)
	require.NoError(t, err)
	assert.Equal(t, "stop", operation.OperationType)
	assert.Equal(t, "Gracefully stopping VM", operation.Message)

	// Test force stop
	operation, err = client.VMs.Stop(context.Background(), 10, 1, true)
	require.NoError(t, err)
	assert.Equal(t, "stop", operation.OperationType)
	assert.Equal(t, "Force stopping VM", operation.Message)
}

func TestVMsService_Restart(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v2/organizations/10/virtual-machines/1/restart", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		response := struct {
			Data *VMOperation `json:"data"`
		}{
			Data: &VMOperation{
				ID:            3,
				VMID:          1,
				OperationType: "restart",
				Status:        "in_progress",
				Progress:      25,
				Message:       "Restarting virtual machine",
				CreatedAt:     CustomTime{Time: time.Now()},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	operation, err := client.VMs.Restart(context.Background(), 10, 1, false)
	require.NoError(t, err)
	assert.Equal(t, "restart", operation.OperationType)
	assert.Equal(t, "in_progress", operation.Status)
	assert.Equal(t, 25, operation.Progress)
}

func TestVMsService_ErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		expectedError  bool
	}{
		{
			name:          "Unauthorized",
			statusCode:    http.StatusUnauthorized,
			expectedError: true,
		},
		{
			name:          "Forbidden",
			statusCode:    http.StatusForbidden,
			expectedError: true,
		},
		{
			name:          "Not Found",
			statusCode:    http.StatusNotFound,
			expectedError: true,
		},
		{
			name:          "Internal Server Error",
			statusCode:    http.StatusInternalServerError,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(StandardResponse{
					Status:  "error",
					Message: "Error occurred",
				})
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			_, _, err = client.VMs.List(context.Background(), nil)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
