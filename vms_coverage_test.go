package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Comprehensive coverage tests for vms.go
// Focus on improving Create (80%), Get (80%), Start (80%), Stop (87.5%), Restart (75%)

// Create tests (80% - needs error path)

func TestVMsService_Create_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/vms", r.URL.Path)

		var body VMConfiguration
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.NoError(t, err)
		assert.Equal(t, "test-vm", body.Name)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":     1,
				"name":   "test-vm",
				"status": "creating",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	vm, err := client.VMs.Create(context.Background(), &VMConfiguration{
		Name:      "test-vm",
		CPUCores:  2,
		MemoryMB:  4096,
		StorageGB: 50,
	})
	assert.NoError(t, err)
	assert.NotNil(t, vm)
}

func TestVMsService_Create_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Invalid VM configuration",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	vm, err := client.VMs.Create(context.Background(), &VMConfiguration{
		Name: "test-vm",
	})
	assert.Error(t, err)
	assert.Nil(t, vm)
}

// Get tests (80% - needs error path)

func TestVMsService_Get_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v1/vms/1", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":     1,
				"name":   "test-vm",
				"status": "running",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	vm, err := client.VMs.Get(context.Background(), 1)
	assert.NoError(t, err)
	assert.NotNil(t, vm)
}

func TestVMsService_Get_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "VM not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	vm, err := client.VMs.Get(context.Background(), 999)
	assert.Error(t, err)
	assert.Nil(t, vm)
}

// Start tests (80% - needs error path)

func TestVMsService_Start_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v2/organizations/1/virtual-machines/10/start", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"operation_id": "op-123",
				"status":       "pending",
				"action":       "start",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	op, err := client.VMs.Start(context.Background(), 1, 10)
	assert.NoError(t, err)
	assert.NotNil(t, op)
}

func TestVMsService_Start_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "VM is already running",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	op, err := client.VMs.Start(context.Background(), 1, 10)
	assert.Error(t, err)
	assert.Nil(t, op)
}

// Stop tests (87.5% - needs error path and force flag test)

func TestVMsService_Stop_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v2/organizations/1/virtual-machines/10/stop", r.URL.Path)

		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.NoError(t, err)
		// Force should not be in body when false
		_, hasForce := body["force"]
		assert.False(t, hasForce)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"operation_id": "op-124",
				"status":       "pending",
				"action":       "stop",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	op, err := client.VMs.Stop(context.Background(), 1, 10, false)
	assert.NoError(t, err)
	assert.NotNil(t, op)
}

func TestVMsService_Stop_WithForce(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)

		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.NoError(t, err)
		// Force should be in body when true
		assert.Equal(t, true, body["force"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"operation_id": "op-125",
				"status":       "pending",
				"action":       "stop",
				"force":        true,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	op, err := client.VMs.Stop(context.Background(), 1, 10, true)
	assert.NoError(t, err)
	assert.NotNil(t, op)
}

func TestVMsService_Stop_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "VM is already stopped",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	op, err := client.VMs.Stop(context.Background(), 1, 10, false)
	assert.Error(t, err)
	assert.Nil(t, op)
}

// Restart tests (75% - needs error path and force flag test)

func TestVMsService_Restart_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v2/organizations/1/virtual-machines/10/restart", r.URL.Path)

		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.NoError(t, err)
		// Force should not be in body when false
		_, hasForce := body["force"]
		assert.False(t, hasForce)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"operation_id": "op-126",
				"status":       "pending",
				"action":       "restart",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	op, err := client.VMs.Restart(context.Background(), 1, 10, false)
	assert.NoError(t, err)
	assert.NotNil(t, op)
}

func TestVMsService_Restart_WithForce(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)

		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.NoError(t, err)
		// Force should be in body when true
		assert.Equal(t, true, body["force"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"operation_id": "op-127",
				"status":       "pending",
				"action":       "restart",
				"force":        true,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	op, err := client.VMs.Restart(context.Background(), 1, 10, true)
	assert.NoError(t, err)
	assert.NotNil(t, op)
}

func TestVMsService_Restart_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "VM is in an invalid state for restart",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	op, err := client.VMs.Restart(context.Background(), 1, 10, false)
	assert.Error(t, err)
	assert.Nil(t, op)
}

// List tests - edge cases

func TestVMsService_List_NilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v1/vms", r.URL.Path)

		// Verify no query parameters when opts is nil
		assert.Empty(t, r.URL.Query().Get("page"))
		assert.Empty(t, r.URL.Query().Get("limit"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []VirtualMachine{},
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
	vms, meta, err := client.VMs.List(context.Background(), nil)
	assert.NoError(t, err)
	assert.NotNil(t, vms)
	assert.NotNil(t, meta)
	assert.Len(t, vms, 0)
}

func TestVMsService_List_WithPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)

		q := r.URL.Query()
		assert.Equal(t, "2", q.Get("page"))
		assert.Equal(t, "50", q.Get("limit"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":     1,
					"name":   "vm-1",
					"status": "running",
				},
			},
			"meta": PaginationMeta{
				Page:       2,
				Limit:      50,
				TotalItems: 100,
				TotalPages: 2,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	vms, meta, err := client.VMs.List(context.Background(), &PaginationOptions{
		Page:  2,
		Limit: 50,
	})
	assert.NoError(t, err)
	assert.NotNil(t, vms)
	assert.NotNil(t, meta)
	assert.Len(t, vms, 1)
	assert.Equal(t, 2, meta.Page)
	assert.Equal(t, 50, meta.Limit)
}

// Delete tests

func TestVMsService_Delete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/v2/organizations/1/virtual-machines/10", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "VM deleted successfully",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	err := client.VMs.Delete(context.Background(), 1, 10)
	assert.NoError(t, err)
}

func TestVMsService_Delete_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "VM not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	err := client.VMs.Delete(context.Background(), 1, 999)
	assert.Error(t, err)
}
