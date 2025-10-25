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

// Comprehensive tests for organization-level health controller methods (Task #2627)
// Methods tested:
// 1. CreateHealthCheckDefinition
// 2. ListHealthCheckDefinitions
// 3. GetHealthCheckDefinition
// 4. UpdateHealthCheckDefinition
// 5. DeleteHealthCheckDefinition
// 6. SubmitHealthCheckResult
// 7. GetOrganizationHealthStatus
// 8. ListHealthAlerts

// ==================== CreateHealthCheckDefinition Tests ====================

func TestHealthService_CreateHealthCheckDefinition_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/health/definitions", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(StandardResponse{
			Status:  "success",
			Message: "Health check definition created successfully",
			Data: &HealthCheckDefinitionResponse{
				ID:           1,
				CheckType:    "database",
				IntervalSeconds: 60,
				TimeoutSeconds:  30,
				TargetName:   "primary-db",
				TargetConfig: map[string]interface{}{"host": "db.example.com"},
				Thresholds:   map[string]interface{}{"response_time_ms": 500},
				Enabled:         true,
				CreatedAt:    time.Now().Format(time.RFC3339),
				UpdatedAt:    time.Now().Format(time.RFC3339),
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	definition, err := client.Health.CreateHealthCheckDefinition(context.Background(), &CreateHealthCheckDefinitionRequest{
		CheckType:       "database",
		CheckName:       "Primary Database Check",
		IntervalSeconds: 60,
		TimeoutSeconds:  30,
		TargetName:      "primary-db",
		TargetConfig:    map[string]interface{}{"host": "db.example.com"},
		Thresholds:      map[string]interface{}{"response_time_ms": 500},
	})

	assert.NoError(t, err)
	assert.NotNil(t, definition)
	assert.Equal(t, uint64(1), definition.ID)
	assert.Equal(t, "database", definition.CheckType)
	assert.Equal(t, "primary-db", definition.TargetName)
	assert.True(t, definition.Enabled)
}

func TestHealthService_CreateHealthCheckDefinition_ValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  "error",
			Message: "Validation failed: check_type is required",
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	definition, err := client.Health.CreateHealthCheckDefinition(context.Background(), &CreateHealthCheckDefinitionRequest{
		TargetName: "test-target",
		// Missing required CheckType
	})

	assert.Error(t, err)
	assert.Nil(t, definition)
}

func TestHealthService_CreateHealthCheckDefinition_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  "error",
			Message: "Unauthorized",
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	definition, err := client.Health.CreateHealthCheckDefinition(context.Background(), &CreateHealthCheckDefinitionRequest{
		CheckType:       "api",
		CheckName:       "API Check",
		IntervalSeconds: 30,
		TargetName:      "test-api",
	})

	assert.Error(t, err)
	assert.Nil(t, definition)
}

func TestHealthService_CreateHealthCheckDefinition_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  "error",
			Message: "Internal server error",
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	definition, err := client.Health.CreateHealthCheckDefinition(context.Background(), &CreateHealthCheckDefinitionRequest{
		CheckType:       "database",
		CheckName:       "DB Check",
		IntervalSeconds: 60,
		TargetName:      "test-db",
	})

	assert.Error(t, err)
	assert.Nil(t, definition)
}

func TestHealthService_CreateHealthCheckDefinition_TypeAssertion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		// Return wrong type to trigger type assertion error
		json.NewEncoder(w).Encode(StandardResponse{
			Status:  "success",
			Message: "Created",
			Data:    "invalid-type", // Should be HealthCheckDefinitionResponse
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	definition, err := client.Health.CreateHealthCheckDefinition(context.Background(), &CreateHealthCheckDefinitionRequest{
		CheckType:       "api",
		CheckName:       "Test API",
		IntervalSeconds: 30,
		TargetName:      "test",
	})

	assert.Error(t, err)
	assert.Nil(t, definition)
	// Type assertion error happens at JSON unmarshal level in resty client
}

// ==================== ListHealthCheckDefinitions Tests ====================

func TestHealthService_ListHealthCheckDefinitions_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/health/definitions", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(StandardResponse{
			Status:  "success",
			Message: "Health check definitions retrieved successfully",
			Data: &ListHealthCheckDefinitionsResponse{
				Total: 2,
				Definitions: []HealthCheckDefinitionResponse{
					{
						ID:              1,
						CheckType:       "database",
						IntervalSeconds: 60,
						TargetName:      "primary-db",
						Enabled:         true,
					},
					{
						ID:              2,
						CheckType:       "api",
						IntervalSeconds: 30,
						TargetName:      "api-endpoint",
						Enabled:         true,
					},
				},
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	list, err := client.Health.ListHealthCheckDefinitions(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, list)
	assert.Equal(t, int64(2), list.Total)
	assert.Len(t, list.Definitions, 2)
	assert.Equal(t, "database", list.Definitions[0].CheckType)
	assert.Equal(t, "api", list.Definitions[1].CheckType)
}

func TestHealthService_ListHealthCheckDefinitions_EmptyList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(StandardResponse{
			Status:  "success",
			Message: "Health check definitions retrieved successfully",
			Data: &ListHealthCheckDefinitionsResponse{
				Total:       0,
				Definitions: []HealthCheckDefinitionResponse{},
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	list, err := client.Health.ListHealthCheckDefinitions(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, list)
	assert.Equal(t, int64(0), list.Total)
	assert.Empty(t, list.Definitions)
}

func TestHealthService_ListHealthCheckDefinitions_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  "error",
			Message: "Unauthorized",
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	list, err := client.Health.ListHealthCheckDefinitions(context.Background())

	assert.Error(t, err)
	assert.Nil(t, list)
}

func TestHealthService_ListHealthCheckDefinitions_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  "error",
			Message: "Internal server error",
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	list, err := client.Health.ListHealthCheckDefinitions(context.Background())

	assert.Error(t, err)
	assert.Nil(t, list)
}

// ==================== GetHealthCheckDefinition Tests ====================

func TestHealthService_GetHealthCheckDefinition_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/health/definitions/1", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(StandardResponse{
			Status:  "success",
			Message: "Health check definition retrieved successfully",
			Data: &HealthCheckDefinitionResponse{
				ID:           1,
				CheckType:    "database",
				IntervalSeconds: 60,
				TimeoutSeconds:  30,
				TargetName:   "primary-db",
				TargetConfig: map[string]interface{}{"host": "db.example.com"},
				Thresholds:   map[string]interface{}{"response_time_ms": 500},
				Enabled:         true,
				CreatedAt:    time.Now().Format(time.RFC3339),
				UpdatedAt:    time.Now().Format(time.RFC3339),
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	definition, err := client.Health.GetHealthCheckDefinition(context.Background(), 1)

	assert.NoError(t, err)
	assert.NotNil(t, definition)
	assert.Equal(t, uint64(1), definition.ID)
	assert.Equal(t, "database", definition.CheckType)
	assert.Equal(t, "primary-db", definition.TargetName)
}

func TestHealthService_GetHealthCheckDefinition_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  "error",
			Message: "Health check definition not found",
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	definition, err := client.Health.GetHealthCheckDefinition(context.Background(), 999)

	assert.Error(t, err)
	assert.Nil(t, definition)
}

func TestHealthService_GetHealthCheckDefinition_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  "error",
			Message: "Unauthorized",
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	definition, err := client.Health.GetHealthCheckDefinition(context.Background(), 1)

	assert.Error(t, err)
	assert.Nil(t, definition)
}

func TestHealthService_GetHealthCheckDefinition_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  "error",
			Message: "Access denied",
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	definition, err := client.Health.GetHealthCheckDefinition(context.Background(), 1)

	assert.Error(t, err)
	assert.Nil(t, definition)
}

func TestHealthService_GetHealthCheckDefinition_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  "error",
			Message: "Internal server error",
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	definition, err := client.Health.GetHealthCheckDefinition(context.Background(), 1)

	assert.Error(t, err)
	assert.Nil(t, definition)
}

// ==================== UpdateHealthCheckDefinition Tests ====================

func TestHealthService_UpdateHealthCheckDefinition_Success_AllFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/health/definitions/1", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(StandardResponse{
			Status:  "success",
			Message: "Health check definition updated successfully",
			Data: &HealthCheckDefinitionResponse{
				ID:         1,
				CheckType:  "database",
				IntervalSeconds: 120,
				TimeoutSeconds:  60,
				TargetName: "primary-db-updated",
				Enabled:     false,
				UpdatedAt:  time.Now().Format(time.RFC3339),
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	definition, err := client.Health.UpdateHealthCheckDefinition(context.Background(), 1, &CreateHealthCheckDefinitionRequest{
		CheckType:  "database",
		IntervalSeconds: 120,
		TimeoutSeconds:  60,
		TargetName: "primary-db-updated",
	})

	assert.NoError(t, err)
	assert.NotNil(t, definition)
	assert.Equal(t, uint64(1), definition.ID)
	assert.Equal(t, 120, definition.IntervalSeconds)
	assert.Equal(t, "primary-db-updated", definition.TargetName)
}

func TestHealthService_UpdateHealthCheckDefinition_Success_PartialFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(StandardResponse{
			Status:  "success",
			Message: "Health check definition updated successfully",
			Data: &HealthCheckDefinitionResponse{
				ID:        1,
				CheckType: "api",
				IntervalSeconds: 90,
				Enabled: true,
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	definition, err := client.Health.UpdateHealthCheckDefinition(context.Background(), 1, &CreateHealthCheckDefinitionRequest{
		IntervalSeconds: 90,
	})

	assert.NoError(t, err)
	assert.NotNil(t, definition)
	assert.Equal(t, 90, definition.IntervalSeconds)
}

func TestHealthService_UpdateHealthCheckDefinition_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  "error",
			Message: "Health check definition not found",
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	definition, err := client.Health.UpdateHealthCheckDefinition(context.Background(), 999, &CreateHealthCheckDefinitionRequest{
		IntervalSeconds: 120,
	})

	assert.Error(t, err)
	assert.Nil(t, definition)
}

func TestHealthService_UpdateHealthCheckDefinition_ValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  "error",
			Message: "Validation failed: interval must be positive",
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	definition, err := client.Health.UpdateHealthCheckDefinition(context.Background(), 1, &CreateHealthCheckDefinitionRequest{
		IntervalSeconds: -10,
	})

	assert.Error(t, err)
	assert.Nil(t, definition)
}

func TestHealthService_UpdateHealthCheckDefinition_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  "error",
			Message: "Unauthorized",
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	definition, err := client.Health.UpdateHealthCheckDefinition(context.Background(), 1, &CreateHealthCheckDefinitionRequest{
		IntervalSeconds: 120,
	})

	assert.Error(t, err)
	assert.Nil(t, definition)
}

// ==================== DeleteHealthCheckDefinition Tests ====================

func TestHealthService_DeleteHealthCheckDefinition_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/health/definitions/1", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(StandardResponse{
			Status:  "success",
			Message: "Health check definition deleted successfully",
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	err = client.Health.DeleteHealthCheckDefinition(context.Background(), 1)

	assert.NoError(t, err)
}

func TestHealthService_DeleteHealthCheckDefinition_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  "error",
			Message: "Health check definition not found",
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	err = client.Health.DeleteHealthCheckDefinition(context.Background(), 999)

	assert.Error(t, err)
}

func TestHealthService_DeleteHealthCheckDefinition_Forbidden_InUse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  "error",
			Message: "Cannot delete health check definition with active results",
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	err = client.Health.DeleteHealthCheckDefinition(context.Background(), 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Cannot delete health check definition")
}

func TestHealthService_DeleteHealthCheckDefinition_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  "error",
			Message: "Unauthorized",
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	err = client.Health.DeleteHealthCheckDefinition(context.Background(), 1)

	assert.Error(t, err)
}

// ==================== SubmitHealthCheckResult Tests ====================

func TestHealthService_SubmitHealthCheckResult_Success_Healthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/health/results", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(StandardResponse{
			Status:  "success",
			Message: "Health check result submitted successfully",
			Data: &HealthCheckResultResponse{
				ID:                1,
				DefinitionID:      10,
				Status:            "healthy",
				Score:             100,
				ResponseTimeMs:    45,
				Message:           "All systems operational",
				ConsecutiveFailures: 0,
				CreatedAt:         time.Now().Format(time.RFC3339),
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	result, err := client.Health.SubmitHealthCheckResult(context.Background(), &SubmitHealthCheckResultRequest{
		DefinitionID:   10,
		Status:         "healthy",
		Score:          100,
		ResponseTimeMs: 45,
		Message:        "All systems operational",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint64(1), result.ID)
	assert.Equal(t, uint64(10), result.DefinitionID)
	assert.Equal(t, "healthy", result.Status)
	assert.Equal(t, 100, result.Score)
	assert.Equal(t, 0, result.ConsecutiveFailures)
}

func TestHealthService_SubmitHealthCheckResult_Success_CriticalWithDetails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(StandardResponse{
			Status:  "success",
			Message: "Health check result submitted successfully",
			Data: &HealthCheckResultResponse{
				ID:                2,
				DefinitionID:      10,
				Status:            "critical",
				Score:             0,
				ResponseTimeMs:    5000,
				Message:           "Database connection timeout",
				Details:           map[string]interface{}{"error": "connection refused", "host": "db.example.com"},
				ConsecutiveFailures: 3,
				CreatedAt:         time.Now().Format(time.RFC3339),
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	result, err := client.Health.SubmitHealthCheckResult(context.Background(), &SubmitHealthCheckResultRequest{
		DefinitionID:   10,
		Status:         "critical",
		Score:          0,
		ResponseTimeMs: 5000,
		Message:        "Database connection timeout",
		Details:        map[string]interface{}{"error": "connection refused", "host": "db.example.com"},
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "critical", result.Status)
	assert.Equal(t, 0, result.Score)
	assert.Equal(t, 3, result.ConsecutiveFailures)
	assert.NotNil(t, result.Details)
}

func TestHealthService_SubmitHealthCheckResult_ValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  "error",
			Message: "Validation failed: definition_id is required",
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	result, err := client.Health.SubmitHealthCheckResult(context.Background(), &SubmitHealthCheckResultRequest{
		Status: "healthy",
		// Missing required DefinitionID
	})

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestHealthService_SubmitHealthCheckResult_DefinitionNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  "error",
			Message: "Health check definition not found",
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	result, err := client.Health.SubmitHealthCheckResult(context.Background(), &SubmitHealthCheckResultRequest{
		DefinitionID: 999,
		Status:       "healthy",
		Score:        100,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestHealthService_SubmitHealthCheckResult_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  "error",
			Message: "Unauthorized",
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	result, err := client.Health.SubmitHealthCheckResult(context.Background(), &SubmitHealthCheckResultRequest{
		DefinitionID: 10,
		Status:       "healthy",
		Score:        100,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
}

// ==================== GetOrganizationHealthStatus Tests ====================

func TestHealthService_GetOrganizationHealthStatus_Success_Healthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/health/status", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(StandardResponse{
			Status:  "success",
			Message: "Organization health status retrieved successfully",
			Data: map[string]interface{}{
				"organization_id":      1,
				"overall_status":       "healthy",
				"overall_score":        98,
				"database_status":      "healthy",
				"database_score":       100,
				"api_status":           "healthy",
				"api_score":            95,
				"resource_status":      "healthy",
				"resource_score":       99,
				"microservice_status":  "healthy",
				"microservice_score":   97,
				"uptime_percent":    99.8,
				"last_evaluated_at":     time.Now().Format(time.RFC3339),
				"created_at":           time.Now().Format(time.RFC3339),
				"updated_at":           time.Now().Format(time.RFC3339),
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	status, err := client.Health.GetOrganizationHealthStatus(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, "healthy", status.OverallStatus)
	assert.Equal(t, 98, status.OverallScore)
	assert.Equal(t, "healthy", status.DatabaseStatus)
	assert.Equal(t, 99.8, status.UptimePercent)
}

func TestHealthService_GetOrganizationHealthStatus_Success_Degraded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(StandardResponse{
			Status:  "success",
			Message: "Organization health status retrieved successfully",
			Data: map[string]interface{}{
				"organization_id":      1,
				"overall_status":      "degraded",
				"overall_score":       75,
				"database_status":     "healthy",
				"database_score":      100,
				"api_status":          "degraded",
				"api_score":           60,
				"resource_status":     "healthy",
				"resource_score":      90,
				"microservice_status": "degraded",
				"microservice_score":  70,
				"uptime_percent":   95.5,
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	status, err := client.Health.GetOrganizationHealthStatus(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, "degraded", status.OverallStatus)
	assert.Equal(t, 75, status.OverallScore)
	assert.Equal(t, "degraded", status.APIStatus)
	assert.Equal(t, 95.5, status.UptimePercent)
}

func TestHealthService_GetOrganizationHealthStatus_Success_Critical(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(StandardResponse{
			Status:  "success",
			Message: "Organization health status retrieved successfully",
			Data: map[string]interface{}{
				"organization_id":      1,
				"overall_status":      "critical",
				"overall_score":       25,
				"database_status":     "critical",
				"database_score":      0,
				"api_status":          "degraded",
				"api_score":           50,
				"resource_status":     "healthy",
				"resource_score":      90,
				"microservice_status": "critical",
				"microservice_score":  10,
				"uptime_percent":   85.2,
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	status, err := client.Health.GetOrganizationHealthStatus(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, "critical", status.OverallStatus)
	assert.Equal(t, 25, status.OverallScore)
	assert.Equal(t, "critical", status.DatabaseStatus)
	assert.Equal(t, 85.2, status.UptimePercent)
}

func TestHealthService_GetOrganizationHealthStatus_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  "error",
			Message: "Unauthorized",
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	status, err := client.Health.GetOrganizationHealthStatus(context.Background())

	assert.Error(t, err)
	assert.Nil(t, status)
}

func TestHealthService_GetOrganizationHealthStatus_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  "error",
			Message: "Internal server error",
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	status, err := client.Health.GetOrganizationHealthStatus(context.Background())

	assert.Error(t, err)
	assert.Nil(t, status)
}

// ==================== ListHealthAlerts Tests ====================

func TestHealthService_ListHealthAlerts_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/health/alerts", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(StandardResponse{
			Status:  "success",
			Message: "Health alerts retrieved successfully",
			Data: &ListHealthAlertsResponse{
				Total: int64(2),
				Alerts: []HealthAlertResponse{
					{
						ID:           1,
						DefinitionID: 10,
						// AlertType:    "threshold_exceeded",
						Severity:     "critical",
						Title:        "Database response time critical",
						Description:  "Response time exceeded threshold of 500ms",
						Acknowledged: false,
						Status:      "active",
						CreatedAt:    time.Now().Format(time.RFC3339),
					},
					{
						ID:           2,
						DefinitionID: 11,
						// AlertType:    "consecutive_failures",
						Severity:     "warning",
						Title:        "API health check degraded",
						Description:  "3 consecutive failures detected",
						Acknowledged: true,
						Status:      "active",
						CreatedAt:    time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
					},
				},
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	alerts, err := client.Health.ListHealthAlerts(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, alerts)
	assert.Equal(t, int64(2), alerts.Total)
	assert.Len(t, alerts.Alerts, 2)
	assert.Equal(t, "critical", alerts.Alerts[0].Severity)
	assert.False(t, alerts.Alerts[0].Acknowledged)
	assert.Equal(t, "warning", alerts.Alerts[1].Severity)
	assert.True(t, alerts.Alerts[1].Acknowledged)
}

func TestHealthService_ListHealthAlerts_EmptyList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(StandardResponse{
			Status:  "success",
			Message: "Health alerts retrieved successfully",
			Data: &ListHealthAlertsResponse{
				Total:  int64(0),
				Alerts: []HealthAlertResponse{},
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	alerts, err := client.Health.ListHealthAlerts(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, alerts)
	assert.Equal(t, int64(0), alerts.Total)
	assert.Empty(t, alerts.Alerts)
}

func TestHealthService_ListHealthAlerts_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  "error",
			Message: "Unauthorized",
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	alerts, err := client.Health.ListHealthAlerts(context.Background())

	assert.Error(t, err)
	assert.Nil(t, alerts)
}

func TestHealthService_ListHealthAlerts_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  "error",
			Message: "Internal server error",
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	alerts, err := client.Health.ListHealthAlerts(context.Background())

	assert.Error(t, err)
	assert.Nil(t, alerts)
}
