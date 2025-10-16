package nexmonyx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProbeControllerService_CreateAssignment tests the CreateAssignment method
func TestProbeControllerService_CreateAssignment(t *testing.T) {
	tests := []struct {
		name           string
		request        *ProbeControllerAssignmentCreateRequest
		serverResponse ProbeControllerAssignment
		serverStatus   int
		expectError    bool
	}{
		{
			name: "successful assignment creation",
			request: &ProbeControllerAssignmentCreateRequest{
				ProbeID:   123,
				ProbeUUID: "probe-uuid-123",
				Region:    "us-east-1",
				Status:    "active",
			},
			serverResponse: ProbeControllerAssignment{
				ID:        1,
				ProbeID:   123,
				ProbeUUID: "probe-uuid-123",
				Region:    "us-east-1",
				Status:    "active",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name: "assignment with monitoring node",
			request: &ProbeControllerAssignmentCreateRequest{
				ProbeID:          456,
				ProbeUUID:        "probe-uuid-456",
				Region:           "eu-west-1",
				MonitoringNodeID: uintPtr(789),
				Status:           "pending",
			},
			serverResponse: ProbeControllerAssignment{
				ID:               2,
				ProbeID:          456,
				ProbeUUID:        "probe-uuid-456",
				Region:           "eu-west-1",
				MonitoringNodeID: uintPtr(789),
				Status:           "pending",
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			},
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name: "server error",
			request: &ProbeControllerAssignmentCreateRequest{
				ProbeID:   999,
				ProbeUUID: "invalid",
				Region:    "invalid",
			},
			serverStatus: http.StatusInternalServerError,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					t.Errorf("Expected POST request, got %s", r.Method)
				}
				if r.URL.Path != "/v1/controllers/probe/assignments" {
					t.Errorf("Expected path /v1/controllers/probe/assignments, got %s", r.URL.Path)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				if tt.serverStatus == http.StatusOK {
					response := struct {
						Status  string                        `json:"status"`
						Data    *ProbeControllerAssignment    `json:"data"`
						Message string                        `json:"message"`
					}{
						Status:  "success",
						Data:    &tt.serverResponse,
						Message: "Assignment created successfully",
					}
					json.NewEncoder(w).Encode(response)
				} else {
					json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
				}
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})

			result, err := client.ProbeController.CreateAssignment(context.Background(), tt.request)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Error("Expected result but got nil")
					return
				}
				if result.ProbeID != tt.serverResponse.ProbeID {
					t.Errorf("Expected ProbeID %d, got %d", tt.serverResponse.ProbeID, result.ProbeID)
				}
				if result.Region != tt.serverResponse.Region {
					t.Errorf("Expected Region %s, got %s", tt.serverResponse.Region, result.Region)
				}
			}
		})
	}
}

// TestProbeControllerService_CreateAssignment_ValidationErrors tests validation error handling
func TestProbeControllerService_CreateAssignment_ValidationErrors(t *testing.T) {
	tests := []struct {
		name           string
		request        *ProbeControllerAssignmentCreateRequest
		responseStatus int
		responseError  map[string]interface{}
		expectError    bool
		errorMessage   string
	}{
		{
			name:         "nil_request_validation",
			request:      nil,
			expectError:  true,
			errorMessage: "request cannot be nil",
		},
		{
			name: "missing_probe_id",
			request: &ProbeControllerAssignmentCreateRequest{
				ProbeID:   0, // Missing/zero
				ProbeUUID: "probe-uuid-123",
				Region:    "us-east-1",
			},
			expectError:  true,
			errorMessage: "probe_id is required",
		},
		{
			name: "missing_probe_uuid",
			request: &ProbeControllerAssignmentCreateRequest{
				ProbeID:   1,
				ProbeUUID: "", // Missing/empty
				Region:    "us-east-1",
			},
			expectError:  true,
			errorMessage: "probe_uuid is required",
		},
		{
			name: "missing_region",
			request: &ProbeControllerAssignmentCreateRequest{
				ProbeID:   1,
				ProbeUUID: "probe-uuid-123",
				Region:    "", // Missing/empty
			},
			expectError:  true,
			errorMessage: "region is required",
		},
		{
			name: "unauthorized_401",
			request: &ProbeControllerAssignmentCreateRequest{
				ProbeID:   1,
				ProbeUUID: "probe-uuid-123",
				Region:    "us-east-1",
			},
			responseStatus: http.StatusUnauthorized,
			responseError: map[string]interface{}{
				"status": "error",
				"error":  "Unauthorized access",
			},
			expectError: true,
		},
		{
			name: "server_error_500",
			request: &ProbeControllerAssignmentCreateRequest{
				ProbeID:   1,
				ProbeUUID: "probe-uuid-123",
				Region:    "us-east-1",
			},
			responseStatus: http.StatusInternalServerError,
			responseError: map[string]interface{}{
				"status": "error",
				"error":  "Internal server error",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/v1/controllers/probe/assignments", r.URL.Path)

				w.Header().Set("Content-Type", "application/json")

				if tt.responseStatus != 0 {
					w.WriteHeader(tt.responseStatus)
					if tt.responseError != nil {
						json.NewEncoder(w).Encode(tt.responseError)
					}
				} else {
					w.WriteHeader(http.StatusCreated)
					response := StandardResponse{
						Status:  "success",
						Message: "Assignment created",
						Data: &ProbeControllerAssignment{
							ID:        1,
							ProbeID:   tt.request.ProbeID,
							ProbeUUID: tt.request.ProbeUUID,
							Region:    tt.request.Region,
							Status:    "active",
						},
					}
					json.NewEncoder(w).Encode(response)
				}
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{MonitoringKey: "test-monitoring-key"},
			})
			require.NoError(t, err)

			result, err := client.ProbeController.CreateAssignment(context.Background(), tt.request)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

// TestProbeControllerService_ListAssignments tests the ListAssignments method
func TestProbeControllerService_ListAssignments(t *testing.T) {
	tests := []struct {
		name           string
		options        *ProbeControllerAssignmentListOptions
		serverResponse []ProbeControllerAssignment
		expectError    bool
		validateQuery  func(*testing.T, map[string][]string)
	}{
		{
			name:    "list all assignments",
			options: nil,
			serverResponse: []ProbeControllerAssignment{
				{ID: 1, ProbeUUID: "probe-1", Region: "us-east-1", Status: "active"},
				{ID: 2, ProbeUUID: "probe-2", Region: "eu-west-1", Status: "pending"},
			},
			expectError: false,
		},
		{
			name: "filter by probe UUID",
			options: &ProbeControllerAssignmentListOptions{
				ProbeUUID: stringPtr("probe-123"),
			},
			serverResponse: []ProbeControllerAssignment{
				{ID: 1, ProbeUUID: "probe-123", Region: "us-east-1", Status: "active"},
			},
			expectError: false,
			validateQuery: func(t *testing.T, query map[string][]string) {
				if query["probe_uuid"][0] != "probe-123" {
					t.Errorf("Expected probe_uuid=probe-123, got %s", query["probe_uuid"][0])
				}
			},
		},
		{
			name: "filter by multiple criteria",
			options: &ProbeControllerAssignmentListOptions{
				ProbeUUID: stringPtr("probe-456"),
				Region:    stringPtr("ap-southeast-1"),
				Status:    stringPtr("active"),
			},
			serverResponse: []ProbeControllerAssignment{
				{ID: 3, ProbeUUID: "probe-456", Region: "ap-southeast-1", Status: "active"},
			},
			expectError: false,
			validateQuery: func(t *testing.T, query map[string][]string) {
				if query["probe_uuid"][0] != "probe-456" {
					t.Errorf("Expected probe_uuid=probe-456")
				}
				if query["region"][0] != "ap-southeast-1" {
					t.Errorf("Expected region=ap-southeast-1")
				}
				if query["status"][0] != "active" {
					t.Errorf("Expected status=active")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("Expected GET request, got %s", r.Method)
				}
				if tt.validateQuery != nil {
					tt.validateQuery(t, r.URL.Query())
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				response := struct {
					Status  string                         `json:"status"`
					Data    []*ProbeControllerAssignment   `json:"data"`
					Message string                         `json:"message"`
				}{
					Status:  "success",
					Data:    make([]*ProbeControllerAssignment, len(tt.serverResponse)),
					Message: "Assignments retrieved successfully",
				}
				for i := range tt.serverResponse {
					response.Data[i] = &tt.serverResponse[i]
				}
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})

			results, err := client.ProbeController.ListAssignments(context.Background(), tt.options)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(results) != len(tt.serverResponse) {
					t.Errorf("Expected %d results, got %d", len(tt.serverResponse), len(results))
				}
			}
		})
	}
}

// TestProbeControllerService_UpdateAssignment tests the UpdateAssignment method
func TestProbeControllerService_UpdateAssignment(t *testing.T) {
	tests := []struct {
		name           string
		assignmentID   uint
		request        *ProbeControllerAssignmentUpdateRequest
		serverResponse ProbeControllerAssignment
		expectError    bool
	}{
		{
			name:         "update status only",
			assignmentID: 1,
			request: &ProbeControllerAssignmentUpdateRequest{
				Status: "paused",
			},
			serverResponse: ProbeControllerAssignment{
				ID:     1,
				Status: "paused",
			},
			expectError: false,
		},
		{
			name:         "update monitoring node",
			assignmentID: 2,
			request: &ProbeControllerAssignmentUpdateRequest{
				MonitoringNodeID: uintPtr(999),
			},
			serverResponse: ProbeControllerAssignment{
				ID:               2,
				MonitoringNodeID: uintPtr(999),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "PUT" {
					t.Errorf("Expected PUT request, got %s", r.Method)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				response := struct {
					Status  string                        `json:"status"`
					Data    *ProbeControllerAssignment    `json:"data"`
					Message string                        `json:"message"`
				}{
					Status:  "success",
					Data:    &tt.serverResponse,
					Message: "Assignment updated successfully",
				}
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})

			result, err := client.ProbeController.UpdateAssignment(context.Background(), tt.assignmentID, tt.request)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result.ID != tt.serverResponse.ID {
					t.Errorf("Expected ID %d, got %d", tt.serverResponse.ID, result.ID)
				}
			}
		})
	}
}

// TestProbeControllerService_UpdateAssignment_ErrorHandling tests error handling for UpdateAssignment
func TestProbeControllerService_UpdateAssignment_ErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		assignmentID   uint
		request        *ProbeControllerAssignmentUpdateRequest
		responseStatus int
		responseError  map[string]interface{}
		expectError    bool
		errorMessage   string
	}{
		{
			name:         "zero_assignment_id_validation",
			assignmentID: 0,
			request: &ProbeControllerAssignmentUpdateRequest{
				Status: "active",
			},
			expectError:  true,
			errorMessage: "assignment id is required",
		},
		{
			name:         "nil_request_validation",
			assignmentID: 1,
			request:      nil,
			expectError:  true,
			errorMessage: "request cannot be nil",
		},
		{
			name:         "assignment_not_found_404",
			assignmentID: 9999,
			request: &ProbeControllerAssignmentUpdateRequest{
				Status: "active",
			},
			responseStatus: http.StatusNotFound,
			responseError: map[string]interface{}{
				"status": "error",
				"error":  "Assignment not found",
			},
			expectError: true,
		},
		{
			name:         "unauthorized_401",
			assignmentID: 1,
			request: &ProbeControllerAssignmentUpdateRequest{
				Status: "active",
			},
			responseStatus: http.StatusUnauthorized,
			responseError: map[string]interface{}{
				"status": "error",
				"error":  "Unauthorized access",
			},
			expectError: true,
		},
		{
			name:         "validation_error_400",
			assignmentID: 1,
			request: &ProbeControllerAssignmentUpdateRequest{
				Status: "invalid-status",
			},
			responseStatus: http.StatusBadRequest,
			responseError: map[string]interface{}{
				"status": "error",
				"error":  "Invalid status value",
			},
			expectError: true,
		},
		{
			name:         "server_error_500",
			assignmentID: 1,
			request: &ProbeControllerAssignmentUpdateRequest{
				Status: "active",
			},
			responseStatus: http.StatusInternalServerError,
			responseError: map[string]interface{}{
				"status": "error",
				"error":  "Internal server error",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "PUT", r.Method)
				expectedPath := fmt.Sprintf("/v1/controllers/probe/assignments/%d", tt.assignmentID)
				assert.Equal(t, expectedPath, r.URL.Path)

				w.Header().Set("Content-Type", "application/json")

				if tt.responseStatus != 0 {
					w.WriteHeader(tt.responseStatus)
					if tt.responseError != nil {
						json.NewEncoder(w).Encode(tt.responseError)
					}
				} else {
					w.WriteHeader(http.StatusOK)
					response := StandardResponse{
						Status:  "success",
						Message: "Assignment updated",
						Data: &ProbeControllerAssignment{
							ID:     tt.assignmentID,
							Status: tt.request.Status,
						},
					}
					json.NewEncoder(w).Encode(response)
				}
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{MonitoringKey: "test-monitoring-key"},
			})
			require.NoError(t, err)

			result, err := client.ProbeController.UpdateAssignment(context.Background(), tt.assignmentID, tt.request)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

// TestProbeControllerService_DeleteAssignment tests the DeleteAssignment method
func TestProbeControllerService_DeleteAssignment(t *testing.T) {
	assignmentID := uint(123)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/controllers/probe/assignments/123" {
			t.Errorf("Expected path /v1/controllers/probe/assignments/123, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		deletedAssignment := ProbeControllerAssignment{
			ID:        123,
			ProbeUUID: "deleted-probe",
			Status:    "deleted",
		}
		response := struct {
			Status  string                        `json:"status"`
			Data    *ProbeControllerAssignment    `json:"data"`
			Message string                        `json:"message"`
		}{
			Status:  "success",
			Data:    &deletedAssignment,
			Message: "Assignment deleted successfully",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, _ := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})

	result, err := client.ProbeController.DeleteAssignment(context.Background(), assignmentID)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.ID != assignmentID {
		t.Errorf("Expected ID %d, got %d", assignmentID, result.ID)
	}
}

// TestProbeControllerService_StoreRegionalResult tests the StoreRegionalResult method
func TestProbeControllerService_StoreRegionalResult(t *testing.T) {
	tests := []struct {
		name           string
		request        *ProbeControllerRegionalResultStoreRequest
		serverResponse ProbeControllerRegionalResult
		expectError    bool
	}{
		{
			name: "successful result storage",
			request: &ProbeControllerRegionalResultStoreRequest{
				ProbeUUID:        "probe-123",
				Region:           "us-east-1",
				Status:           "up",
				ResponseTime:     intPtr(150),
				Success:          true,
				IsCustomerRegion: false,
				TTLSeconds:       3600,
			},
			serverResponse: ProbeControllerRegionalResult{
				ID:               1,
				ProbeUUID:        "probe-123",
				Region:           "us-east-1",
				Status:           "up",
				ResponseTime:     intPtr(150),
				Success:          true,
				IsCustomerRegion: false,
				Timestamp:        time.Now(),
			},
			expectError: false,
		},
		{
			name: "failed probe result",
			request: &ProbeControllerRegionalResultStoreRequest{
				ProbeUUID:        "probe-456",
				Region:           "eu-west-1",
				Status:           "down",
				Success:          false,
				ErrorMessage:     stringPtr("Connection timeout"),
				IsCustomerRegion: true,
				TTLSeconds:       7200,
			},
			serverResponse: ProbeControllerRegionalResult{
				ID:               2,
				ProbeUUID:        "probe-456",
				Region:           "eu-west-1",
				Status:           "down",
				Success:          false,
				ErrorMessage:     stringPtr("Connection timeout"),
				IsCustomerRegion: true,
				Timestamp:        time.Now(),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					t.Errorf("Expected POST request, got %s", r.Method)
				}
				if r.URL.Path != "/v1/controllers/probe/results/regional" {
					t.Errorf("Expected path /v1/controllers/probe/results/regional, got %s", r.URL.Path)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				response := struct {
					Status  string                           `json:"status"`
					Data    *ProbeControllerRegionalResult   `json:"data"`
					Message string                           `json:"message"`
				}{
					Status:  "success",
					Data:    &tt.serverResponse,
					Message: "Regional result stored successfully",
				}
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})

			result, err := client.ProbeController.StoreRegionalResult(context.Background(), tt.request)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result.ProbeUUID != tt.serverResponse.ProbeUUID {
					t.Errorf("Expected ProbeUUID %s, got %s", tt.serverResponse.ProbeUUID, result.ProbeUUID)
				}
				if result.Region != tt.serverResponse.Region {
					t.Errorf("Expected Region %s, got %s", tt.serverResponse.Region, result.Region)
				}
				if result.Success != tt.serverResponse.Success {
					t.Errorf("Expected Success %v, got %v", tt.serverResponse.Success, result.Success)
				}
			}
		})
	}
}

// TestProbeControllerService_StoreRegionalResult_ValidationErrors tests validation error handling
func TestProbeControllerService_StoreRegionalResult_ValidationErrors(t *testing.T) {
	tests := []struct {
		name           string
		request        *ProbeControllerRegionalResultStoreRequest
		responseStatus int
		responseError  map[string]interface{}
		expectError    bool
		errorMessage   string
	}{
		{
			name:         "nil_request_validation",
			request:      nil,
			expectError:  true,
			errorMessage: "request cannot be nil",
		},
		{
			name: "missing_probe_uuid",
			request: &ProbeControllerRegionalResultStoreRequest{
				ProbeUUID:    "", // Missing/empty
				Region:       "us-east-1",
				Status:       "up",
				Success:      true,
				TTLSeconds:   3600,
			},
			expectError:  true,
			errorMessage: "probe_uuid is required",
		},
		{
			name: "missing_region",
			request: &ProbeControllerRegionalResultStoreRequest{
				ProbeUUID:    "probe-123",
				Region:       "", // Missing/empty
				Status:       "up",
				Success:      true,
				TTLSeconds:   3600,
			},
			expectError:  true,
			errorMessage: "region is required",
		},
		{
			name: "missing_status",
			request: &ProbeControllerRegionalResultStoreRequest{
				ProbeUUID:    "probe-123",
				Region:       "us-east-1",
				Status:       "", // Missing/empty
				Success:      true,
				TTLSeconds:   3600,
			},
			expectError:  true,
			errorMessage: "status is required",
		},
		{
			name: "unauthorized_401",
			request: &ProbeControllerRegionalResultStoreRequest{
				ProbeUUID:    "probe-123",
				Region:       "us-east-1",
				Status:       "up",
				Success:      true,
				TTLSeconds:   3600,
			},
			responseStatus: http.StatusUnauthorized,
			responseError: map[string]interface{}{
				"status": "error",
				"error":  "Unauthorized access",
			},
			expectError: true,
		},
		{
			name: "duplicate_result_409",
			request: &ProbeControllerRegionalResultStoreRequest{
				ProbeUUID:    "probe-123",
				Region:       "us-east-1",
				Status:       "up",
				Success:      true,
				TTLSeconds:   3600,
			},
			responseStatus: http.StatusConflict,
			responseError: map[string]interface{}{
				"status": "error",
				"error":  "Duplicate result for this timestamp",
			},
			expectError: true,
		},
		{
			name: "server_error_500",
			request: &ProbeControllerRegionalResultStoreRequest{
				ProbeUUID:    "probe-123",
				Region:       "us-east-1",
				Status:       "up",
				Success:      true,
				TTLSeconds:   3600,
			},
			responseStatus: http.StatusInternalServerError,
			responseError: map[string]interface{}{
				"status": "error",
				"error":  "Internal server error",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/v1/controllers/probe/results/regional", r.URL.Path)

				w.Header().Set("Content-Type", "application/json")

				if tt.responseStatus != 0 {
					w.WriteHeader(tt.responseStatus)
					if tt.responseError != nil {
						json.NewEncoder(w).Encode(tt.responseError)
					}
				} else {
					w.WriteHeader(http.StatusOK)
					response := StandardResponse{
						Status:  "success",
						Message: "Regional result stored",
						Data: &ProbeControllerRegionalResult{
							ID:        1,
							ProbeUUID: tt.request.ProbeUUID,
							Region:    tt.request.Region,
							Status:    tt.request.Status,
							Success:   tt.request.Success,
							Timestamp: time.Now(),
						},
					}
					json.NewEncoder(w).Encode(response)
				}
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{MonitoringKey: "test-monitoring-key"},
			})
			require.NoError(t, err)

			result, err := client.ProbeController.StoreRegionalResult(context.Background(), tt.request)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

// TestProbeControllerService_GetRegionalResults tests the GetRegionalResults method
func TestProbeControllerService_GetRegionalResults(t *testing.T) {
	tests := []struct {
		name           string
		probeUUID      string
		options        *ProbeControllerRegionalResultListOptions
		serverResponse []ProbeControllerRegionalResult
		expectError    bool
	}{
		{
			name:      "get all results for probe",
			probeUUID: "probe-123",
			options:   nil,
			serverResponse: []ProbeControllerRegionalResult{
				{ID: 1, ProbeUUID: "probe-123", Region: "us-east-1", Status: "up", Success: true},
				{ID: 2, ProbeUUID: "probe-123", Region: "eu-west-1", Status: "up", Success: true},
			},
			expectError: false,
		},
		{
			name:      "filter by region",
			probeUUID: "probe-456",
			options: &ProbeControllerRegionalResultListOptions{
				Region: stringPtr("ap-southeast-1"),
			},
			serverResponse: []ProbeControllerRegionalResult{
				{ID: 3, ProbeUUID: "probe-456", Region: "ap-southeast-1", Status: "down", Success: false},
			},
			expectError: false,
		},
		{
			name:      "filter by customer regions",
			probeUUID: "probe-789",
			options: &ProbeControllerRegionalResultListOptions{
				IsCustomerRegion: boolPtr(true),
			},
			serverResponse: []ProbeControllerRegionalResult{
				{ID: 4, ProbeUUID: "probe-789", Region: "custom-region-1", IsCustomerRegion: true, Success: true},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("Expected GET request, got %s", r.Method)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				response := struct {
					Status  string                             `json:"status"`
					Data    []*ProbeControllerRegionalResult   `json:"data"`
					Message string                             `json:"message"`
				}{
					Status:  "success",
					Data:    make([]*ProbeControllerRegionalResult, len(tt.serverResponse)),
					Message: "Regional results retrieved successfully",
				}
				for i := range tt.serverResponse {
					response.Data[i] = &tt.serverResponse[i]
				}
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})

			results, err := client.ProbeController.GetRegionalResults(context.Background(), tt.probeUUID, tt.options)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(results) != len(tt.serverResponse) {
					t.Errorf("Expected %d results, got %d", len(tt.serverResponse), len(results))
				}
			}
		})
	}
}

// TestProbeControllerService_StoreConsensusResult tests the StoreConsensusResult method
func TestProbeControllerService_StoreConsensusResult(t *testing.T) {
	tests := []struct {
		name           string
		request        *ProbeControllerConsensusResultStoreRequest
		serverResponse ProbeControllerConsensusResult
		expectError    bool
	}{
		{
			name: "majority consensus - probe up",
			request: &ProbeControllerConsensusResultStoreRequest{
				ProbeID:         123,
				ProbeUUID:       "probe-123",
				GlobalStatus:    "up",
				ConsensusType:   "majority",
				ShouldAlert:     false,
				UpRegions:       3,
				DownRegions:     0,
				DegradedRegions: 0,
				UnknownRegions:  0,
				TotalRegions:    3,
				ConsensusRatio:  1.0,
				AlertTriggered:  false,
			},
			serverResponse: ProbeControllerConsensusResult{
				ID:              1,
				ProbeID:         123,
				ProbeUUID:       "probe-123",
				GlobalStatus:    "up",
				ConsensusType:   "majority",
				ShouldAlert:     false,
				UpRegions:       3,
				DownRegions:     0,
				TotalRegions:    3,
				ConsensusRatio:  1.0,
				AlertTriggered:  false,
				CalculatedAt:    time.Now(),
			},
			expectError: false,
		},
		{
			name: "majority consensus - probe down with alert",
			request: &ProbeControllerConsensusResultStoreRequest{
				ProbeID:         456,
				ProbeUUID:       "probe-456",
				GlobalStatus:    "down",
				ConsensusType:   "majority",
				ShouldAlert:     true,
				UpRegions:       0,
				DownRegions:     3,
				DegradedRegions: 0,
				UnknownRegions:  0,
				TotalRegions:    3,
				ConsensusRatio:  0.0,
				AlertTriggered:  true,
			},
			serverResponse: ProbeControllerConsensusResult{
				ID:             2,
				ProbeID:        456,
				ProbeUUID:      "probe-456",
				GlobalStatus:   "down",
				ConsensusType:  "majority",
				ShouldAlert:    true,
				UpRegions:      0,
				DownRegions:    3,
				TotalRegions:   3,
				ConsensusRatio: 0.0,
				AlertTriggered: true,
				CalculatedAt:   time.Now(),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					t.Errorf("Expected POST request, got %s", r.Method)
				}
				if r.URL.Path != "/v1/controllers/probe/results/consensus" {
					t.Errorf("Expected path /v1/controllers/probe/results/consensus, got %s", r.URL.Path)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				response := struct {
					Status  string                            `json:"status"`
					Data    *ProbeControllerConsensusResult   `json:"data"`
					Message string                            `json:"message"`
				}{
					Status:  "success",
					Data:    &tt.serverResponse,
					Message: "Consensus result stored successfully",
				}
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})

			result, err := client.ProbeController.StoreConsensusResult(context.Background(), tt.request)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result.ProbeUUID != tt.serverResponse.ProbeUUID {
					t.Errorf("Expected ProbeUUID %s, got %s", tt.serverResponse.ProbeUUID, result.ProbeUUID)
				}
				if result.GlobalStatus != tt.serverResponse.GlobalStatus {
					t.Errorf("Expected GlobalStatus %s, got %s", tt.serverResponse.GlobalStatus, result.GlobalStatus)
				}
				if result.ShouldAlert != tt.serverResponse.ShouldAlert {
					t.Errorf("Expected ShouldAlert %v, got %v", tt.serverResponse.ShouldAlert, result.ShouldAlert)
				}
			}
		})
	}
}

// TestProbeControllerService_StoreConsensusResult_ValidationErrors tests validation error handling
func TestProbeControllerService_StoreConsensusResult_ValidationErrors(t *testing.T) {
	tests := []struct {
		name           string
		request        *ProbeControllerConsensusResultStoreRequest
		responseStatus int
		responseError  map[string]interface{}
		expectError    bool
		errorMessage   string
	}{
		{
			name:         "nil_request_validation",
			request:      nil,
			expectError:  true,
			errorMessage: "request cannot be nil",
		},
		{
			name: "missing_probe_id",
			request: &ProbeControllerConsensusResultStoreRequest{
				ProbeID:       0, // Missing/zero
				ProbeUUID:     "probe-123",
				GlobalStatus:  "up",
				ConsensusType: "majority",
			},
			expectError:  true,
			errorMessage: "probe_id is required",
		},
		{
			name: "missing_probe_uuid",
			request: &ProbeControllerConsensusResultStoreRequest{
				ProbeID:       123,
				ProbeUUID:     "", // Missing/empty
				GlobalStatus:  "up",
				ConsensusType: "majority",
			},
			expectError:  true,
			errorMessage: "probe_uuid is required",
		},
		{
			name: "missing_global_status",
			request: &ProbeControllerConsensusResultStoreRequest{
				ProbeID:       123,
				ProbeUUID:     "probe-123",
				GlobalStatus:  "", // Missing/empty
				ConsensusType: "majority",
			},
			expectError:  true,
			errorMessage: "global_status is required",
		},
		{
			name: "missing_consensus_type",
			request: &ProbeControllerConsensusResultStoreRequest{
				ProbeID:       123,
				ProbeUUID:     "probe-123",
				GlobalStatus:  "up",
				ConsensusType: "", // Missing/empty
			},
			expectError:  true,
			errorMessage: "consensus_type is required",
		},
		{
			name: "unauthorized_401",
			request: &ProbeControllerConsensusResultStoreRequest{
				ProbeID:       123,
				ProbeUUID:     "probe-123",
				GlobalStatus:  "up",
				ConsensusType: "majority",
			},
			responseStatus: http.StatusUnauthorized,
			responseError: map[string]interface{}{
				"status": "error",
				"error":  "Unauthorized access",
			},
			expectError: true,
		},
		{
			name: "invalid_consensus_data_400",
			request: &ProbeControllerConsensusResultStoreRequest{
				ProbeID:         123,
				ProbeUUID:       "probe-123",
				GlobalStatus:    "up",
				ConsensusType:   "invalid-type",
				DegradedRegions: -1, // Invalid negative value
			},
			responseStatus: http.StatusBadRequest,
			responseError: map[string]interface{}{
				"status": "error",
				"error":  "Invalid consensus data",
			},
			expectError: true,
		},
		{
			name: "server_error_500",
			request: &ProbeControllerConsensusResultStoreRequest{
				ProbeID:       123,
				ProbeUUID:     "probe-123",
				GlobalStatus:  "up",
				ConsensusType: "majority",
			},
			responseStatus: http.StatusInternalServerError,
			responseError: map[string]interface{}{
				"status": "error",
				"error":  "Internal server error",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/v1/controllers/probe/results/consensus", r.URL.Path)

				w.Header().Set("Content-Type", "application/json")

				if tt.responseStatus != 0 {
					w.WriteHeader(tt.responseStatus)
					if tt.responseError != nil {
						json.NewEncoder(w).Encode(tt.responseError)
					}
				} else {
					w.WriteHeader(http.StatusOK)
					response := StandardResponse{
						Status:  "success",
						Message: "Consensus result stored",
						Data: &ProbeControllerConsensusResult{
							ID:           1,
							ProbeID:      tt.request.ProbeID,
							ProbeUUID:    tt.request.ProbeUUID,
							GlobalStatus: tt.request.GlobalStatus,
							CalculatedAt: time.Now(),
						},
					}
					json.NewEncoder(w).Encode(response)
				}
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{MonitoringKey: "test-monitoring-key"},
			})
			require.NoError(t, err)

			result, err := client.ProbeController.StoreConsensusResult(context.Background(), tt.request)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

// TestProbeControllerService_GetConsensusHistory tests the GetConsensusHistory method
func TestProbeControllerService_GetConsensusHistory(t *testing.T) {
	probeUUID := "probe-123"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		query := r.URL.Query()
		if limit := query.Get("limit"); limit != "" && limit != "100" {
			t.Errorf("Expected limit=100, got %s", limit)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		consensusResults := []*ProbeControllerConsensusResult{
			{ID: 1, ProbeUUID: probeUUID, GlobalStatus: "up", ConsensusType: "majority"},
			{ID: 2, ProbeUUID: probeUUID, GlobalStatus: "down", ConsensusType: "majority"},
			{ID: 3, ProbeUUID: probeUUID, GlobalStatus: "up", ConsensusType: "majority"},
		}
		response := struct {
			Status  string                              `json:"status"`
			Data    []*ProbeControllerConsensusResult   `json:"data"`
			Message string                              `json:"message"`
		}{
			Status:  "success",
			Data:    consensusResults,
			Message: "Consensus history retrieved successfully",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, _ := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})

	limit := 100
	results, err := client.ProbeController.GetConsensusHistory(context.Background(), probeUUID, &ConsensusHistoryOptions{
		Limit: &limit,
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}
	for _, result := range results {
		if result.ProbeUUID != probeUUID {
			t.Errorf("Expected ProbeUUID %s, got %s", probeUUID, result.ProbeUUID)
		}
	}
}

// TestProbeControllerService_UpdateHealthState tests the UpdateHealthState method
func TestProbeControllerService_UpdateHealthState(t *testing.T) {
	tests := []struct {
		name           string
		request        *ProbeControllerHealthUpdateRequest
		serverResponse ProbeControllerHealthState
		expectError    bool
	}{
		{
			name: "update controller status",
			request: &ProbeControllerHealthUpdateRequest{
				Key:   "controller_status",
				Value: "healthy",
			},
			serverResponse: ProbeControllerHealthState{
				ID:        1,
				Key:       "controller_status",
				Value:     "healthy",
				UpdatedAt: time.Now(),
			},
			expectError: false,
		},
		{
			name: "update probe count metric",
			request: &ProbeControllerHealthUpdateRequest{
				Key:   "active_probes",
				Value: "42",
			},
			serverResponse: ProbeControllerHealthState{
				ID:        2,
				Key:       "active_probes",
				Value:     "42",
				UpdatedAt: time.Now(),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "PUT" {
					t.Errorf("Expected PUT request, got %s", r.Method)
				}
				if r.URL.Path != "/v1/controllers/probe/health" {
					t.Errorf("Expected path /v1/controllers/probe/health, got %s", r.URL.Path)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				response := struct {
					Status  string                        `json:"status"`
					Data    *ProbeControllerHealthState   `json:"data"`
					Message string                        `json:"message"`
				}{
					Status:  "success",
					Data:    &tt.serverResponse,
					Message: "Health state updated successfully",
				}
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})

			result, err := client.ProbeController.UpdateHealthState(context.Background(), tt.request)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result.Key != tt.serverResponse.Key {
					t.Errorf("Expected Key %s, got %s", tt.serverResponse.Key, result.Key)
				}
				if result.Value != tt.serverResponse.Value {
					t.Errorf("Expected Value %s, got %s", tt.serverResponse.Value, result.Value)
				}
			}
		})
	}
}

// TestProbeControllerService_UpdateHealthState_ErrorHandling tests error scenarios for UpdateHealthState
func TestProbeControllerService_UpdateHealthState_ErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		request        *ProbeControllerHealthUpdateRequest
		responseStatus int
		responseError  map[string]interface{}
		expectError    bool
	}{
		{
			name:           "nil_request_validation",
			request:        nil,
			responseStatus: http.StatusOK, // Won't be called
			expectError:    true,          // nil request causes validation error
		},
		{
			name: "server_error_500",
			request: &ProbeControllerHealthUpdateRequest{
				Key:   "controller_status",
				Value: "healthy",
			},
			responseStatus: http.StatusInternalServerError,
			responseError: map[string]interface{}{
				"status": "error",
				"error":  "Internal server error",
			},
			expectError: true,
		},
		{
			name: "unauthorized_401",
			request: &ProbeControllerHealthUpdateRequest{
				Key:   "active_probes",
				Value: "100",
			},
			responseStatus: http.StatusUnauthorized,
			responseError: map[string]interface{}{
				"status": "error",
				"error":  "Unauthorized",
			},
			expectError: true,
		},
		{
			name: "validation_error_400",
			request: &ProbeControllerHealthUpdateRequest{
				Key:   "", // Empty key
				Value: "test",
			},
			responseStatus: http.StatusBadRequest,
			responseError: map[string]interface{}{
				"status": "error",
				"error":  "Invalid key",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "PUT" {
					t.Errorf("Expected PUT request, got %s", r.Method)
				}
				if r.URL.Path != "/v1/controllers/probe/health" {
					t.Errorf("Expected path /v1/controllers/probe/health, got %s", r.URL.Path)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.responseStatus)

				if tt.expectError {
					json.NewEncoder(w).Encode(tt.responseError)
				} else {
					response := struct {
						Status  string                        `json:"status"`
						Data    *ProbeControllerHealthState   `json:"data"`
						Message string                        `json:"message"`
					}{
						Status: "success",
						Data: &ProbeControllerHealthState{
							ID:        1,
							Key:       tt.request.Key,
							Value:     tt.request.Value,
							UpdatedAt: time.Now(),
						},
						Message: "Health state updated successfully",
					}
					json.NewEncoder(w).Encode(response)
				}
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})

			result, err := client.ProbeController.UpdateHealthState(context.Background(), tt.request)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if result != nil {
					t.Error("Expected nil result on error")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == nil && tt.request != nil {
					t.Error("Expected non-nil result")
				}
			}
		})
	}
}

// TestProbeControllerService_GetHealthStates tests the GetHealthStates method
func TestProbeControllerService_GetHealthStates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/controllers/probe/health" {
			t.Errorf("Expected path /v1/controllers/probe/health, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		healthStates := []*ProbeControllerHealthState{
			{ID: 1, Key: "controller_status", Value: "healthy", UpdatedAt: time.Now()},
			{ID: 2, Key: "active_probes", Value: "42", UpdatedAt: time.Now()},
			{ID: 3, Key: "last_sync", Value: "2024-01-01T00:00:00Z", UpdatedAt: time.Now()},
		}
		response := struct {
			Status  string                          `json:"status"`
			Data    []*ProbeControllerHealthState   `json:"data"`
			Message string                          `json:"message"`
		}{
			Status:  "success",
			Data:    healthStates,
			Message: "Health states retrieved successfully",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, _ := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})

	results, err := client.ProbeController.GetHealthStates(context.Background())

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	expectedKeys := map[string]bool{"controller_status": true, "active_probes": true, "last_sync": true}
	for _, result := range results {
		if !expectedKeys[result.Key] {
			t.Errorf("Unexpected key: %s", result.Key)
		}
	}
}

// Helper functions for test data
func uintPtr(v uint) *uint {
	return &v
}
