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

// TestVMsService_ListComprehensive tests the List method
func TestVMsService_ListComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		opts       *PaginationOptions
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, []VirtualMachine, *PaginationMeta)
	}{
		{
			name: "success - list all VMs",
			opts: &PaginationOptions{Page: 1, Limit: 25},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id":              1,
						"organization_id": 1,
						"name":            "web-vm-01",
						"description":     "Web server VM",
						"status":          "running",
						"cpu_cores":       4,
						"memory_mb":       8192,
						"storage_gb":      100,
						"ip_address":      "10.0.1.10",
						"os_type":         "linux",
						"os_version":      "Ubuntu 22.04",
						"created_at":      "2024-01-15T10:00:00Z",
					},
					{
						"id":              2,
						"organization_id": 1,
						"name":            "db-vm-01",
						"status":          "stopped",
						"cpu_cores":       8,
						"memory_mb":       16384,
						"storage_gb":      500,
					},
				},
				"meta": map[string]interface{}{
					"page":        1,
					"limit":       25,
					"total_items": 2,
					"total_pages": 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, vms []VirtualMachine, meta *PaginationMeta) {
				assert.Len(t, vms, 2)
				assert.Equal(t, "web-vm-01", vms[0].Name)
				assert.Equal(t, "running", vms[0].Status)
				assert.Equal(t, 4, vms[0].CPUCores)
				assert.Equal(t, 2, meta.TotalItems)
			},
		},
		{
			name:       "success - empty list",
			opts:       &PaginationOptions{Page: 1, Limit: 25},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": []map[string]interface{}{},
				"meta": map[string]interface{}{
					"page":        1,
					"limit":       25,
					"total_items": 0,
					"total_pages": 0,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, vms []VirtualMachine, meta *PaginationMeta) {
				assert.Len(t, vms, 0)
				assert.Equal(t, 0, meta.TotalItems)
			},
		},
		{
			name:       "success - with pagination",
			opts:       &PaginationOptions{Page: 2, Limit: 10},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": []map[string]interface{}{
					{"id": 11, "name": "vm-11", "status": "running", "cpu_cores": 2, "memory_mb": 4096, "storage_gb": 50},
				},
				"meta": map[string]interface{}{
					"page":        2,
					"limit":       10,
					"total_items": 15,
					"total_pages": 2,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, vms []VirtualMachine, meta *PaginationMeta) {
				assert.Len(t, vms, 1)
				assert.Equal(t, 2, meta.Page)
			},
		},
		{
			name:       "unauthorized",
			opts:       &PaginationOptions{Page: 1, Limit: 25},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			opts:       &PaginationOptions{Page: 1, Limit: 25},
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/api/v1/vms", r.URL.Path)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL:    server.URL,
				Auth:       AuthConfig{Token: "test-token"},
				RetryCount: 0,
			})
			require.NoError(t, err)

			ctx := context.Background()
			if tt.wantErr && tt.mockStatus >= 500 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
			}

			vms, meta, err := client.VMs.List(ctx, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, vms)
				if tt.checkFunc != nil {
					tt.checkFunc(t, vms, meta)
				}
			}
		})
	}
}

// TestVMsService_CreateComprehensive tests the Create method
func TestVMsService_CreateComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		config     *VMConfiguration
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *VirtualMachine)
	}{
		{
			name: "success - create VM with full config",
			config: &VMConfiguration{
				Name:        "production-web-vm",
				Description: "Production web server VM",
				CPUCores:    8,
				MemoryMB:    16384,
				StorageGB:   200,
				OSType:      "linux",
			},
			mockStatus: http.StatusCreated,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":              1,
					"organization_id": 1,
					"name":            "production-web-vm",
					"description":     "Production web server VM",
					"status":          "provisioning",
					"cpu_cores":       8,
					"memory_mb":       16384,
					"storage_gb":      200,
					"os_type":         "linux",
					"created_at":      "2024-01-15T10:00:00Z",
					"updated_at":      "2024-01-15T10:00:00Z",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, vm *VirtualMachine) {
				assert.Equal(t, uint(1), vm.ID)
				assert.Equal(t, "production-web-vm", vm.Name)
				assert.Equal(t, "provisioning", vm.Status)
				assert.Equal(t, 8, vm.CPUCores)
				assert.Equal(t, 16384, vm.MemoryMB)
			},
		},
		{
			name: "success - minimal VM config",
			config: &VMConfiguration{
				Name:      "test-vm",
				CPUCores:  2,
				MemoryMB:  4096,
				StorageGB: 50,
			},
			mockStatus: http.StatusCreated,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":         2,
					"name":       "test-vm",
					"status":     "provisioning",
					"cpu_cores":  2,
					"memory_mb":  4096,
					"storage_gb": 50,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, vm *VirtualMachine) {
				assert.Equal(t, "test-vm", vm.Name)
				assert.Equal(t, 2, vm.CPUCores)
			},
		},
		{
			name: "validation error - missing required fields",
			config: &VMConfiguration{
				Name: "incomplete-vm",
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "CPU cores, memory, and storage are required",
			},
			wantErr: true,
		},
		{
			name: "validation error - invalid resource values",
			config: &VMConfiguration{
				Name:      "invalid-vm",
				CPUCores:  0,
				MemoryMB:  100,
				StorageGB: -10,
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Invalid resource specifications",
			},
			wantErr: true,
		},
		{
			name: "unauthorized",
			config: &VMConfiguration{
				Name:      "test-vm",
				CPUCores:  2,
				MemoryMB:  4096,
				StorageGB: 50,
			},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name: "server error",
			config: &VMConfiguration{
				Name:      "test-vm",
				CPUCores:  2,
				MemoryMB:  4096,
				StorageGB: 50,
			},
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Failed to provision VM",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/api/v1/vms", r.URL.Path)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL:    server.URL,
				Auth:       AuthConfig{Token: "test-token"},
				RetryCount: 0,
			})
			require.NoError(t, err)

			ctx := context.Background()
			if tt.wantErr && tt.mockStatus >= 500 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
			}

			result, err := client.VMs.Create(ctx, tt.config)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestVMsService_GetComprehensive tests the Get method
func TestVMsService_GetComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		vmID       uint
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *VirtualMachine)
	}{
		{
			name:       "success - get running VM",
			vmID:       1,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":              1,
					"organization_id": 1,
					"name":            "web-vm-01",
					"description":     "Production web server",
					"status":          "running",
					"cpu_cores":       8,
					"memory_mb":       16384,
					"storage_gb":      200,
					"ip_address":      "10.0.1.10",
					"mac_address":     "00:1A:2B:3C:4D:5E",
					"os_type":         "linux",
					"os_version":      "Ubuntu 22.04",
					"created_at":      "2024-01-15T10:00:00Z",
					"started_at":      "2024-01-15T10:05:00Z",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, vm *VirtualMachine) {
				assert.Equal(t, uint(1), vm.ID)
				assert.Equal(t, "web-vm-01", vm.Name)
				assert.Equal(t, "running", vm.Status)
				assert.Equal(t, "10.0.1.10", vm.IPAddress)
				assert.Equal(t, 8, vm.CPUCores)
			},
		},
		{
			name:       "success - get stopped VM",
			vmID:       2,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":         2,
					"name":       "test-vm",
					"status":     "stopped",
					"cpu_cores":  2,
					"memory_mb":  4096,
					"storage_gb": 50,
					"stopped_at": "2024-01-15T12:00:00Z",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, vm *VirtualMachine) {
				assert.Equal(t, uint(2), vm.ID)
				assert.Equal(t, "stopped", vm.Status)
			},
		},
		{
			name:       "not found",
			vmID:       999,
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Virtual machine not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			vmID:       1,
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			vmID:       1,
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL:    server.URL,
				Auth:       AuthConfig{Token: "test-token"},
				RetryCount: 0,
			})
			require.NoError(t, err)

			ctx := context.Background()
			if tt.wantErr && tt.mockStatus >= 500 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
			}

			result, err := client.VMs.Get(ctx, tt.vmID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestVMsService_DeleteComprehensive tests the Delete method
func TestVMsService_DeleteComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		orgID      uint
		vmID       uint
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:       "success - delete VM",
			orgID:      1,
			vmID:       1,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status":  "success",
				"message": "VM deleted successfully",
			},
			wantErr: false,
		},
		{
			name:       "success - no content",
			orgID:      1,
			vmID:       2,
			mockStatus: http.StatusNoContent,
			mockBody:   nil,
			wantErr:    false,
		},
		{
			name:       "not found",
			orgID:      1,
			vmID:       999,
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Virtual machine not found",
			},
			wantErr: true,
		},
		{
			name:       "conflict - VM still running",
			orgID:      1,
			vmID:       1,
			mockStatus: http.StatusConflict,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Cannot delete running VM. Stop it first.",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			orgID:      1,
			vmID:       1,
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "forbidden - insufficient permissions",
			orgID:      1,
			vmID:       1,
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Insufficient permissions to delete VM",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			orgID:      1,
			vmID:       1,
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Failed to delete VM",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "DELETE", r.Method)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				if tt.mockBody != nil {
					json.NewEncoder(w).Encode(tt.mockBody)
				}
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL:    server.URL,
				Auth:       AuthConfig{Token: "test-token"},
				RetryCount: 0,
			})
			require.NoError(t, err)

			ctx := context.Background()
			if tt.wantErr && tt.mockStatus >= 500 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
			}

			err = client.VMs.Delete(ctx, tt.orgID, tt.vmID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestVMsService_StartComprehensive tests the Start method
func TestVMsService_StartComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		orgID      uint
		vmID       uint
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *VMOperation)
	}{
		{
			name:       "success - start VM",
			orgID:      1,
			vmID:       1,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":             1,
					"vm_id":          1,
					"operation_type": "start",
					"status":         "in_progress",
					"progress":       25,
					"message":        "Starting virtual machine",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, op *VMOperation) {
				assert.Equal(t, uint(1), op.ID)
				assert.Equal(t, "start", op.OperationType)
				assert.Equal(t, "in_progress", op.Status)
				assert.Equal(t, 25, op.Progress)
			},
		},
		{
			name:       "not found",
			orgID:      1,
			vmID:       999,
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Virtual machine not found",
			},
			wantErr: true,
		},
		{
			name:       "conflict - already running",
			orgID:      1,
			vmID:       1,
			mockStatus: http.StatusConflict,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "VM is already running",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			orgID:      1,
			vmID:       1,
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			orgID:      1,
			vmID:       1,
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Failed to start VM",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL:    server.URL,
				Auth:       AuthConfig{Token: "test-token"},
				RetryCount: 0,
			})
			require.NoError(t, err)

			ctx := context.Background()
			if tt.wantErr && tt.mockStatus >= 500 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
			}

			result, err := client.VMs.Start(ctx, tt.orgID, tt.vmID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestVMsService_StopComprehensive tests the Stop method
func TestVMsService_StopComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		orgID      uint
		vmID       uint
		force      bool
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *VMOperation)
	}{
		{
			name:       "success - graceful stop",
			orgID:      1,
			vmID:       1,
			force:      false,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":             1,
					"vm_id":          1,
					"operation_type": "stop",
					"status":         "in_progress",
					"progress":       10,
					"message":        "Gracefully stopping VM",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, op *VMOperation) {
				assert.Equal(t, "stop", op.OperationType)
				assert.Equal(t, "in_progress", op.Status)
			},
		},
		{
			name:       "success - force stop",
			orgID:      1,
			vmID:       1,
			force:      true,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":             2,
					"vm_id":          1,
					"operation_type": "stop",
					"status":         "completed",
					"progress":       100,
					"message":        "VM force stopped",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, op *VMOperation) {
				assert.Equal(t, "completed", op.Status)
				assert.Equal(t, 100, op.Progress)
			},
		},
		{
			name:       "conflict - already stopped",
			orgID:      1,
			vmID:       1,
			force:      false,
			mockStatus: http.StatusConflict,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "VM is already stopped",
			},
			wantErr: true,
		},
		{
			name:       "not found",
			orgID:      1,
			vmID:       999,
			force:      false,
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Virtual machine not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			orgID:      1,
			vmID:       1,
			force:      false,
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			orgID:      1,
			vmID:       1,
			force:      false,
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Failed to stop VM",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL:    server.URL,
				Auth:       AuthConfig{Token: "test-token"},
				RetryCount: 0,
			})
			require.NoError(t, err)

			ctx := context.Background()
			if tt.wantErr && tt.mockStatus >= 500 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
			}

			result, err := client.VMs.Stop(ctx, tt.orgID, tt.vmID, tt.force)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestVMsService_RestartComprehensive tests the Restart method
func TestVMsService_RestartComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		orgID      uint
		vmID       uint
		force      bool
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *VMOperation)
	}{
		{
			name:       "success - graceful restart",
			orgID:      1,
			vmID:       1,
			force:      false,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":             1,
					"vm_id":          1,
					"operation_type": "restart",
					"status":         "in_progress",
					"progress":       50,
					"message":        "Restarting VM",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, op *VMOperation) {
				assert.Equal(t, "restart", op.OperationType)
				assert.Equal(t, "in_progress", op.Status)
				assert.Equal(t, 50, op.Progress)
			},
		},
		{
			name:       "success - force restart",
			orgID:      1,
			vmID:       1,
			force:      true,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":             2,
					"vm_id":          1,
					"operation_type": "restart",
					"status":         "completed",
					"progress":       100,
					"message":        "VM force restarted",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, op *VMOperation) {
				assert.Equal(t, "completed", op.Status)
			},
		},
		{
			name:       "conflict - VM not running",
			orgID:      1,
			vmID:       1,
			force:      false,
			mockStatus: http.StatusConflict,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Cannot restart stopped VM",
			},
			wantErr: true,
		},
		{
			name:       "not found",
			orgID:      1,
			vmID:       999,
			force:      false,
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Virtual machine not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			orgID:      1,
			vmID:       1,
			force:      false,
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			orgID:      1,
			vmID:       1,
			force:      false,
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Failed to restart VM",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL:    server.URL,
				Auth:       AuthConfig{Token: "test-token"},
				RetryCount: 0,
			})
			require.NoError(t, err)

			ctx := context.Background()
			if tt.wantErr && tt.mockStatus >= 500 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
			}

			result, err := client.VMs.Restart(ctx, tt.orgID, tt.vmID, tt.force)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}
