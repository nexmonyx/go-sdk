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

func TestControllersService_SendHeartbeat(t *testing.T) {
	tests := []struct {
		name           string
		request        *ControllerHeartbeatRequest
		responseStatus int
		responseBody   interface{}
		wantErr        bool
		errContains    string
	}{
		{
			name: "successful heartbeat",
			request: &ControllerHeartbeatRequest{
				ControllerID: "controller-1",
				Status:       "healthy",
				Version:      "1.0.0",
			},
			responseStatus: http.StatusOK,
			responseBody: StandardResponse{
				Status:  "success",
				Message: "Heartbeat received",
			},
			wantErr: false,
		},
		{
			name: "heartbeat with metrics",
			request: &ControllerHeartbeatRequest{
				ControllerID: "controller-2",
				Status:       "healthy",
				Version:      "1.2.0",
				Metadata: map[string]interface{}{
					"cpu_usage":    45.2,
					"memory_usage": 60.5,
					"goroutines":   150,
				},
			},
			responseStatus: http.StatusOK,
			responseBody: StandardResponse{
				Status:  "success",
				Message: "Heartbeat with metrics received",
			},
			wantErr: false,
		},
		{
			name: "missing controller name",
			request: &ControllerHeartbeatRequest{
				Status:  "healthy",
				Version: "1.0.0",
			},
			wantErr:     true,
			errContains: "controller name is required",
		},
		{
			name: "controller not found",
			request: &ControllerHeartbeatRequest{
				ControllerID: "nonexistent",
				Status:         "healthy",
			},
			responseStatus: http.StatusNotFound,
			responseBody: ErrorResponse{
				Status:  "error",
				Message: "controller not found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip HTTP call if validation error expected
			if tt.errContains != "" && tt.responseStatus == 0 {
				client, _ := NewClient(&Config{BaseURL: "http://localhost"})
				err := client.Controllers.SendHeartbeat(context.Background(), tt.request)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Contains(t, r.URL.Path, "/controllers/")
				assert.Contains(t, r.URL.Path, "/heartbeat")

				// Decode and verify request body
				var req ControllerHeartbeatRequest
				err := json.NewDecoder(r.Body).Decode(&req)
				require.NoError(t, err)
				assert.Equal(t, tt.request.ControllerID, req.ControllerID)

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL})
			err := client.Controllers.SendHeartbeat(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestControllersService_RegisterController(t *testing.T) {
	tests := []struct {
		name           string
		controllerID   string
		version        string
		responseStatus int
		responseBody   interface{}
		wantErr        bool
	}{
		{
			name:           "successful registration",
			controllerID:   "controller-1",
			version:        "1.0.0",
			responseStatus: http.StatusOK,
			responseBody: StandardResponse{
				Status:  "success",
				Message: "Controller registered successfully",
			},
			wantErr: false,
		},
		{
			name:           "registration with new version",
			controllerID:   "controller-2",
			version:        "2.0.0",
			responseStatus: http.StatusOK,
			responseBody: StandardResponse{
				Status:  "success",
				Message: "Controller registered",
			},
			wantErr: false,
		},
		{
			name:           "duplicate registration",
			controllerID:   "controller-existing",
			version:        "1.0.0",
			responseStatus: http.StatusConflict,
			responseBody: ErrorResponse{
				Status:  "error",
				Message: "controller already registered",
			},
			wantErr: true,
		},
		{
			name:           "invalid version",
			controllerID:   "controller-3",
			version:        "invalid",
			responseStatus: http.StatusBadRequest,
			responseBody: ErrorResponse{
				Status:  "error",
				Message: "invalid version format",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/v1/controllers/register", r.URL.Path)

				// Verify request body
				var req map[string]interface{}
				err := json.NewDecoder(r.Body).Decode(&req)
				require.NoError(t, err)
				assert.Equal(t, tt.controllerID, req["controller_id"])
				assert.Equal(t, tt.version, req["version"])

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL})
			err := client.Controllers.RegisterController(context.Background(), tt.controllerID, tt.version)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestControllersService_GetControllerStatus(t *testing.T) {
	tests := []struct {
		name           string
		controllerID   string
		responseStatus int
		responseBody   interface{}
		wantErr        bool
		validateResult func(*testing.T, *ControllerHealthInfo)
	}{
		{
			name:           "healthy controller",
			controllerID:   "controller-1",
			responseStatus: http.StatusOK,
			responseBody: StandardResponse{
				Status:  "success",
				Message: "Controller status retrieved",
				Data: &ControllerHealthInfo{
					Status:  "healthy",
					Version: "1.0.0",
					Uptime:  3600,
				},
			},
			wantErr: false,
			validateResult: func(t *testing.T, info *ControllerHealthInfo) {
				assert.Equal(t, "healthy", info.Status)
				assert.Equal(t, "1.0.0", info.Version)
				assert.Equal(t, time.Duration(3600), info.Uptime)
			},
		},
		{
			name:           "degraded controller",
			controllerID:   "controller-2",
			responseStatus: http.StatusOK,
			responseBody: StandardResponse{
				Status: "success",
				Data: &ControllerHealthInfo{
					Status:  "degraded",
					Version: "1.1.0",
				},
			},
			validateResult: func(t *testing.T, info *ControllerHealthInfo) {
				assert.Equal(t, "degraded", info.Status)
			},
		},
		{
			name:           "controller not found",
			controllerID:   "nonexistent",
			responseStatus: http.StatusNotFound,
			responseBody: ErrorResponse{
				Status:  "error",
				Message: "controller not found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/controllers/")
				assert.Contains(t, r.URL.Path, "/status")
				assert.Contains(t, r.URL.Path, tt.controllerID)

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL})
			result, err := client.Controllers.GetControllerStatus(context.Background(), tt.controllerID)

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

func TestControllersService_ListControllers(t *testing.T) {
	tests := []struct {
		name           string
		responseStatus int
		responseBody   interface{}
		wantErr        bool
		validateResult func(*testing.T, []ControllerHealthInfo)
	}{
		{
			name:           "list all controllers",
			responseStatus: http.StatusOK,
			responseBody: StandardResponse{
				Status:  "success",
				Message: "Controllers retrieved",
				Data: &[]ControllerHealthInfo{
					{
						Status:  "healthy",
						Version: "1.0.0",
						Uptime:  3600,
					},
					{
						Status:  "degraded",
						Version: "1.1.0",
						Uptime:  7200,
					},
					{
						Status:  "healthy",
						Version: "2.0.0",
						Uptime:  1800,
					},
				},
			},
			wantErr: false,
			validateResult: func(t *testing.T, controllers []ControllerHealthInfo) {
				assert.Len(t, controllers, 3)
				assert.Equal(t, "healthy", controllers[0].Status)
				assert.Equal(t, "degraded", controllers[1].Status)
			},
		},
		{
			name:           "empty list",
			responseStatus: http.StatusOK,
			responseBody: StandardResponse{
				Status:  "success",
				Message: "No controllers found",
				Data:    &[]ControllerHealthInfo{},
			},
			wantErr: false,
			validateResult: func(t *testing.T, controllers []ControllerHealthInfo) {
				assert.Len(t, controllers, 0)
			},
		},
		{
			name:           "server error",
			responseStatus: http.StatusInternalServerError,
			responseBody: ErrorResponse{
				Status:  "error",
				Message: "internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/v1/controllers", r.URL.Path)

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL})
			result, err := client.Controllers.ListControllers(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}

func TestControllersService_UpdateControllerStatus(t *testing.T) {
	tests := []struct {
		name           string
		controllerID   string
		status         string
		responseStatus int
		responseBody   interface{}
		wantErr        bool
	}{
		{
			name:           "update to healthy",
			controllerID:   "controller-1",
			status:         "healthy",
			responseStatus: http.StatusOK,
			responseBody: StandardResponse{
				Status:  "success",
				Message: "Status updated successfully",
			},
			wantErr: false,
		},
		{
			name:           "update to degraded",
			controllerID:   "controller-2",
			status:         "degraded",
			responseStatus: http.StatusOK,
			responseBody: StandardResponse{
				Status:  "success",
				Message: "Status updated to degraded",
			},
			wantErr: false,
		},
		{
			name:           "update to unhealthy",
			controllerID:   "controller-3",
			status:         "unhealthy",
			responseStatus: http.StatusOK,
			responseBody: StandardResponse{
				Status:  "success",
				Message: "Status updated to unhealthy",
			},
			wantErr: false,
		},
		{
			name:           "controller not found",
			controllerID:   "nonexistent",
			status:         "healthy",
			responseStatus: http.StatusNotFound,
			responseBody: ErrorResponse{
				Status:  "error",
				Message: "controller not found",
			},
			wantErr: true,
		},
		{
			name:           "invalid status",
			controllerID:   "controller-4",
			status:         "invalid_status",
			responseStatus: http.StatusBadRequest,
			responseBody: ErrorResponse{
				Status:  "error",
				Message: "invalid status value",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "PUT", r.Method)
				assert.Contains(t, r.URL.Path, "/controllers/")
				assert.Contains(t, r.URL.Path, "/status")
				assert.Contains(t, r.URL.Path, tt.controllerID)

				// Verify request body
				var req map[string]interface{}
				err := json.NewDecoder(r.Body).Decode(&req)
				require.NoError(t, err)
				assert.Equal(t, tt.status, req["status"])

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL})
			err := client.Controllers.UpdateControllerStatus(context.Background(), tt.controllerID, tt.status)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestControllersService_DeregisterController(t *testing.T) {
	tests := []struct {
		name           string
		controllerID   string
		responseStatus int
		responseBody   interface{}
		wantErr        bool
	}{
		{
			name:           "successful deregistration",
			controllerID:   "controller-1",
			responseStatus: http.StatusOK,
			responseBody: StandardResponse{
				Status:  "success",
				Message: "Controller deregistered successfully",
			},
			wantErr: false,
		},
		{
			name:           "controller not found",
			controllerID:   "nonexistent",
			responseStatus: http.StatusNotFound,
			responseBody: ErrorResponse{
				Status:  "error",
				Message: "controller not found",
			},
			wantErr: true,
		},
		{
			name:           "controller in use",
			controllerID:   "active-controller",
			responseStatus: http.StatusConflict,
			responseBody: ErrorResponse{
				Status:  "error",
				Message: "controller is currently active and cannot be deregistered",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "DELETE", r.Method)
				assert.Contains(t, r.URL.Path, "/controllers/")
				assert.Contains(t, r.URL.Path, tt.controllerID)

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL})
			err := client.Controllers.DeregisterController(context.Background(), tt.controllerID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// =============================================================================
// Integration Scenarios
// =============================================================================

func TestControllersService_CompleteLifecycle(t *testing.T) {
	// This test simulates the complete lifecycle of a controller:
	// 1. Register
	// 2. Send heartbeat
	// 3. Update status
	// 4. Get status
	// 5. Deregister

	controllerID := "test-controller-lifecycle"
	version := "1.0.0"

	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		switch {
		case r.Method == "POST" && r.URL.Path == "/v1/controllers/register":
			// Step 1: Register
			json.NewEncoder(w).Encode(StandardResponse{Status: "success"})

		case r.Method == "POST" && r.URL.Path == "/v1/controllers/"+controllerID+"/heartbeat":
			// Step 2: Heartbeat
			json.NewEncoder(w).Encode(StandardResponse{Status: "success"})

		case r.Method == "PUT" && r.URL.Path == "/v1/controllers/"+controllerID+"/status":
			// Step 3: Update status
			json.NewEncoder(w).Encode(StandardResponse{Status: "success"})

		case r.Method == "GET" && r.URL.Path == "/v1/controllers/"+controllerID+"/status":
			// Step 4: Get status
			json.NewEncoder(w).Encode(StandardResponse{
				Status: "success",
				Data: &ControllerHealthInfo{
					Status:       "healthy",
					Version:      version,
				},
			})

		case r.Method == "DELETE" && r.URL.Path == "/v1/controllers/"+controllerID:
			// Step 5: Deregister
			json.NewEncoder(w).Encode(StandardResponse{Status: "success"})

		default:
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	// Step 1: Register
	err := client.Controllers.RegisterController(context.Background(), controllerID, version)
	require.NoError(t, err)

	// Step 2: Send heartbeat
	err = client.Controllers.SendHeartbeat(context.Background(), &ControllerHeartbeatRequest{
		ControllerID: controllerID,
		Status:       "healthy",
		Version:      version,
	})
	require.NoError(t, err)

	// Step 3: Update status
	err = client.Controllers.UpdateControllerStatus(context.Background(), controllerID, "healthy")
	require.NoError(t, err)

	// Step 4: Get status
	status, err := client.Controllers.GetControllerStatus(context.Background(), controllerID)
	require.NoError(t, err)
	assert.Equal(t, "healthy", status.Status)

	// Step 5: Deregister
	err = client.Controllers.DeregisterController(context.Background(), controllerID)
	require.NoError(t, err)

	// Verify all steps were called
	assert.Equal(t, 5, callCount)
}

func TestControllersService_MultipleHeartbeats(t *testing.T) {
	// Test sending multiple heartbeats in sequence
	heartbeatCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		heartbeatCount++
		assert.Equal(t, "POST", r.Method)
		json.NewEncoder(w).Encode(StandardResponse{Status: "success"})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	// Send multiple heartbeats
	for i := 0; i < 5; i++ {
		err := client.Controllers.SendHeartbeat(context.Background(), &ControllerHeartbeatRequest{
			ControllerID: "test-controller",
			Status:       "healthy",
			Version:      "1.0.0",
		})
		require.NoError(t, err)
	}

	assert.Equal(t, 5, heartbeatCount)
}

func TestControllersService_ConcurrentStatusUpdates(t *testing.T) {
	// Test concurrent status updates for different controllers
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(StandardResponse{Status: "success"})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	// Update status for multiple controllers concurrently
	done := make(chan bool)

	for i := 1; i <= 3; i++ {
		go func(id int) {
			controllerID := "controller-" + string(rune('0'+id))
			err := client.Controllers.UpdateControllerStatus(context.Background(), controllerID, "healthy")
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}
}

// =============================================================================
// Edge Cases and Error Handling
// =============================================================================

func TestControllersService_EmptyControllerID(t *testing.T) {
	client, _ := NewClient(&Config{BaseURL: "http://localhost"})

	err := client.Controllers.SendHeartbeat(context.Background(), &ControllerHeartbeatRequest{
		ControllerID: "",
		Status:       "healthy",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "controller name is required")
}

func TestControllersService_NilRequest(t *testing.T) {
	client, _ := NewClient(&Config{BaseURL: "http://localhost"})

	err := client.Controllers.SendHeartbeat(context.Background(), nil)

	require.Error(t, err)
}

func TestControllersService_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Delay to allow context cancellation
		select {
		case <-r.Context().Done():
			return
		}
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := client.Controllers.SendHeartbeat(ctx, &ControllerHeartbeatRequest{
		ControllerID: "test",
		Status:       "healthy",
	})

	assert.Error(t, err)
}
