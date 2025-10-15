package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatabaseService_CreateOrganizationSchema(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/admin/database/schemas", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"schema_name": "org_123",
				"exists":      true,
				"message":     "Schema created successfully",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	response, err := client.Database.CreateOrganizationSchema(context.Background(), 123)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "org_123", response.SchemaName)
}

func TestDatabaseService_CreateOrganizationSchema_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Failed to create schema",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	response, err := client.Database.CreateOrganizationSchema(context.Background(), 123)
	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestDatabaseService_DeleteOrganizationSchema(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Contains(t, r.URL.Path, "/v1/admin/database/schemas/")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"schema_name": "org_456",
				"exists":      false,
				"message":     "Schema deleted successfully",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	response, err := client.Database.DeleteOrganizationSchema(context.Background(), 456)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "org_456", response.SchemaName)
}

func TestDatabaseService_DeleteOrganizationSchema_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Schema not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	response, err := client.Database.DeleteOrganizationSchema(context.Background(), 999)
	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestDatabaseService_CheckSchemaExists(t *testing.T) {
	tests := []struct {
		name       string
		exists     bool
		statusCode int
	}{
		{
			name:       "schema exists",
			exists:     true,
			statusCode: http.StatusOK,
		},
		{
			name:       "schema does not exist",
			exists:     false,
			statusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/admin/database/schemas/")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"status": "success",
					"data": map[string]interface{}{
						"schema_name": "org_789",
						"exists":      tt.exists,
						"message":     "Check complete",
					},
				})
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL})
			exists, err := client.Database.CheckSchemaExists(context.Background(), 789)
			assert.NoError(t, err)
			assert.Equal(t, tt.exists, exists)
		})
	}
}

func TestDatabaseService_CheckSchemaExists_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Access denied",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	exists, err := client.Database.CheckSchemaExists(context.Background(), 999)
	assert.Error(t, err)
	assert.False(t, exists)
}

func TestDatabaseService_AllMethods_NetworkError(t *testing.T) {
	// Test network error handling by using invalid URL
	client, _ := NewClient(&Config{BaseURL: "http://invalid-server-that-does-not-exist:9999"})

	_, err := client.Database.CreateOrganizationSchema(context.Background(), 1)
	assert.Error(t, err)

	_, err = client.Database.DeleteOrganizationSchema(context.Background(), 1)
	assert.Error(t, err)

	_, err = client.Database.CheckSchemaExists(context.Background(), 1)
	assert.Error(t, err)
}
