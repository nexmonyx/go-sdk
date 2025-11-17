package nexmonyx

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAPIKeysService_CreateUnifiedComprehensive tests the CreateUnified method with various scenarios
func TestAPIKeysService_CreateUnifiedComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		request    *CreateUnifiedAPIKeyRequest
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *CreateUnifiedAPIKeyResponse)
	}{
		{
			name: "success - full user key creation",
			request: &CreateUnifiedAPIKeyRequest{
				Name:         "User API Key",
				Description:  "Test user API key",
				Type:         APIKeyTypeUser,
				Capabilities: []string{"read", "write"},
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"key": map[string]interface{}{
						"id":              123,
						"key_id":          "key-123",
						"name":            "User API Key",
						"type":            "user",
						"capabilities":    []string{"read", "write"},
						"organization_id": 1,
					},
					"secret": "secret-abc123",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, resp *CreateUnifiedAPIKeyResponse) {
				assert.NotNil(t, resp.Key)
				assert.Equal(t, "secret-abc123", resp.Secret)
			},
		},
		{
			name: "success - monitoring agent key",
			request: &CreateUnifiedAPIKeyRequest{
				Name:               "Monitoring Agent Key",
				Type:               APIKeyTypeMonitoringAgent,
				NamespaceName:      "prod",
				AgentType:          "prometheus",
				RegionCode:         "us-east-1",
				AllowedProbeScopes: []string{"metrics:submit"},
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"key": map[string]interface{}{
						"id":              456,
						"key_id":          "key-456",
						"name":            "Monitoring Agent Key",
						"type":            "monitoring_agent",
						"namespace_name":  "prod",
						"agent_type":      "prometheus",
						"region_code":     "us-east-1",
						"organization_id": 1,
					},
					"secret": "secret-def456",
				},
			},
			wantErr: false,
		},
		{
			name: "validation error - missing name",
			request: &CreateUnifiedAPIKeyRequest{
				Type: APIKeyTypeUser,
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Name is required",
			},
			wantErr: true,
		},
		{
			name: "validation error - invalid type",
			request: &CreateUnifiedAPIKeyRequest{
				Name: "Test Key",
				Type: "invalid-type",
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Invalid key type",
			},
			wantErr: true,
		},
		{
			name: "unauthorized",
			request: &CreateUnifiedAPIKeyRequest{
				Name: "Test Key",
				Type: APIKeyTypeUser,
			},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name: "forbidden - insufficient permissions",
			request: &CreateUnifiedAPIKeyRequest{
				Name: "Admin Key",
				Type: APIKeyTypeAdmin,
			},
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Insufficient permissions to create admin key",
			},
			wantErr: true,
		},
		{
			name: "server error",
			request: &CreateUnifiedAPIKeyRequest{
				Name: "Test Key",
				Type: APIKeyTypeUser,
			},
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
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/v2/api-keys", r.URL.Path)

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

			resp, err := client.APIKeys.CreateUnified(ctx, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.checkFunc != nil {
					tt.checkFunc(t, resp)
				}
			}
		})
	}
}

// TestAPIKeysService_GetUnifiedComprehensive tests the GetUnified method with various scenarios
func TestAPIKeysService_GetUnifiedComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		keyID      string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *UnifiedAPIKey)
	}{
		{
			name:       "success - full key details",
			keyID:      "key-123",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":              123,
					"key_id":          "key-123",
					"name":            "Test Key",
					"type":            "user",
					"status":          "active",
					"capabilities":    []string{"read", "write"},
					"organization_id": 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, key *UnifiedAPIKey) {
				assert.Equal(t, "Test Key", key.Name)
			},
		},
		{
			name:       "success - monitoring agent key",
			keyID:      "key-456",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":              456,
					"key_id":          "key-456",
					"type":            "monitoring_agent",
					"status":          "active",
					"namespace_name":  "prod",
					"agent_type":      "prometheus",
					"region_code":     "us-east-1",
					"organization_id": 1,
				},
			},
			wantErr: false,
		},
		{
			name:       "not found",
			keyID:      "nonexistent",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "API key not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			keyID:      "key-123",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "forbidden",
			keyID:      "key-123",
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Access denied",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			keyID:      "key-123",
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
				assert.Contains(t, r.URL.Path, "/v2/api-keys/")

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

			key, err := client.APIKeys.GetUnified(ctx, tt.keyID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, key)
				if tt.checkFunc != nil {
					tt.checkFunc(t, key)
				}
			}
		})
	}
}

// TestAPIKeysService_ListUnifiedComprehensive tests the ListUnified method with various scenarios
func TestAPIKeysService_ListUnifiedComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		opts       *ListUnifiedAPIKeysOptions
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, []*UnifiedAPIKey, *PaginationMeta)
	}{
		{
			name: "success - with options",
			opts: &ListUnifiedAPIKeysOptions{
				ListOptions: ListOptions{Page: 1, Limit: 25},
				Type:        APIKeyTypeUser,
				Status:      APIKeyStatusActive,
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": []map[string]interface{}{
					{"id": 1, "name": "Key 1", "type": "user"},
					{"id": 2, "name": "Key 2", "type": "user"},
				},
				"meta": map[string]interface{}{
					"page":        1,
					"limit":       25,
					"total_items": 2,
					"total_pages": 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, keys []*UnifiedAPIKey, meta *PaginationMeta) {
				assert.Len(t, keys, 2)
				assert.Equal(t, 1, meta.Page)
				assert.Equal(t, 2, meta.TotalItems)
			},
		},
		{
			name:       "success - nil options",
			opts:       nil,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data":   []map[string]interface{}{},
				"meta": map[string]interface{}{
					"page":        1,
					"limit":       25,
					"total_items": 0,
					"total_pages": 0,
				},
			},
			wantErr: false,
		},
		{
			name: "success - filter by agent type",
			opts: &ListUnifiedAPIKeysOptions{
				ListOptions: ListOptions{Page: 1, Limit: 10},
				Type:        APIKeyTypeMonitoringAgent,
				AgentType:   "prometheus",
				RegionCode:  "us-east-1",
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": []map[string]interface{}{
					{
						"id":              1,
						"key_id":          "key-1",
						"type":            "monitoring_agent",
						"agent_type":      "prometheus",
						"region_code":     "us-east-1",
						"organization_id": 1,
					},
				},
				"meta": map[string]interface{}{
					"page":        1,
					"limit":       10,
					"total_items": 1,
					"total_pages": 1,
				},
			},
			wantErr: false,
		},
		{
			name: "unauthorized",
			opts: &ListUnifiedAPIKeysOptions{
				ListOptions: ListOptions{Page: 1, Limit: 25},
			},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name: "forbidden",
			opts: &ListUnifiedAPIKeysOptions{
				ListOptions: ListOptions{Page: 1, Limit: 25},
			},
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Access denied",
			},
			wantErr: true,
		},
		{
			name: "server error",
			opts: &ListUnifiedAPIKeysOptions{
				ListOptions: ListOptions{Page: 1, Limit: 25},
			},
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
				assert.Equal(t, "/v2/api-keys", r.URL.Path)

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

			keys, meta, err := client.APIKeys.ListUnified(ctx, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, keys)
				assert.NotNil(t, meta)
				if tt.checkFunc != nil {
					tt.checkFunc(t, keys, meta)
				}
			}
		})
	}
}

// TestAPIKeysService_UpdateUnifiedComprehensive tests the UpdateUnified method with various scenarios
func TestAPIKeysService_UpdateUnifiedComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		keyID      string
		request    *UpdateUnifiedAPIKeyRequest
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *UnifiedAPIKey)
	}{
		{
			name:  "success - update name",
			keyID: "key-123",
			request: &UpdateUnifiedAPIKeyRequest{
				Name: stringPtr("Updated Key Name"),
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":              123,
					"key_id":          "key-123",
					"name":            "Updated Key Name",
					"organization_id": 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, key *UnifiedAPIKey) {
				assert.Equal(t, "Updated Key Name", key.Name)
			},
		},
		{
			name:  "success - update capabilities",
			keyID: "key-123",
			request: &UpdateUnifiedAPIKeyRequest{
				Capabilities: []string{"read", "write", "delete"},
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":              123,
					"key_id":          "key-123",
					"capabilities":    []string{"read", "write", "delete"},
					"organization_id": 1,
				},
			},
			wantErr: false,
		},
		{
			name:  "validation error - invalid capabilities",
			keyID: "key-123",
			request: &UpdateUnifiedAPIKeyRequest{
				Capabilities: []string{"invalid_capability"},
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Invalid capability",
			},
			wantErr: true,
		},
		{
			name:  "not found",
			keyID: "nonexistent",
			request: &UpdateUnifiedAPIKeyRequest{
				Name: stringPtr("Updated"),
			},
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "API key not found",
			},
			wantErr: true,
		},
		{
			name:  "unauthorized",
			keyID: "key-123",
			request: &UpdateUnifiedAPIKeyRequest{
				Name: stringPtr("Updated"),
			},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:  "forbidden",
			keyID: "key-123",
			request: &UpdateUnifiedAPIKeyRequest{
				Name: stringPtr("Updated"),
			},
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Access denied",
			},
			wantErr: true,
		},
		{
			name:  "server error",
			keyID: "key-123",
			request: &UpdateUnifiedAPIKeyRequest{
				Name: stringPtr("Updated"),
			},
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
				assert.Equal(t, "PUT", r.Method)
				assert.Contains(t, r.URL.Path, "/v2/api-keys/")

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

			key, err := client.APIKeys.UpdateUnified(ctx, tt.keyID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, key)
				if tt.checkFunc != nil {
					tt.checkFunc(t, key)
				}
			}
		})
	}
}

// TestAPIKeysService_DeleteUnifiedComprehensive tests the DeleteUnified method with various scenarios
func TestAPIKeysService_DeleteUnifiedComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		keyID      string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:       "success",
			keyID:      "key-123",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status":  "success",
				"message": "API key deleted",
			},
			wantErr: false,
		},
		{
			name:       "not found",
			keyID:      "nonexistent",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "API key not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			keyID:      "key-123",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "forbidden",
			keyID:      "key-123",
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Cannot delete this API key",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			keyID:      "key-123",
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
				assert.Equal(t, "DELETE", r.Method)
				assert.Contains(t, r.URL.Path, "/v2/api-keys/")

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

			err = client.APIKeys.DeleteUnified(ctx, tt.keyID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestAPIKeysService_RevokeUnifiedComprehensive tests the RevokeUnified method with various scenarios
func TestAPIKeysService_RevokeUnifiedComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		keyID      string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:       "success",
			keyID:      "key-123",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status":  "success",
				"message": "API key revoked",
			},
			wantErr: false,
		},
		{
			name:       "not found",
			keyID:      "nonexistent",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "API key not found",
			},
			wantErr: true,
		},
		{
			name:       "already revoked",
			keyID:      "key-123",
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "API key already revoked",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			keyID:      "key-123",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "forbidden",
			keyID:      "key-123",
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Cannot revoke this API key",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			keyID:      "key-123",
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
				assert.Equal(t, "POST", r.Method)
				assert.Contains(t, r.URL.Path, "/revoke")

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

			err = client.APIKeys.RevokeUnified(ctx, tt.keyID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestAPIKeysService_RegenerateUnifiedComprehensive tests the RegenerateUnified method with various scenarios
func TestAPIKeysService_RegenerateUnifiedComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		keyID      string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *CreateUnifiedAPIKeyResponse)
	}{
		{
			name:       "success",
			keyID:      "key-123",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"key": map[string]interface{}{
						"id":              123,
						"key_id":          "key-123",
						"name":            "Test Key",
						"organization_id": 1,
					},
					"secret": "new-secret-xyz789",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, resp *CreateUnifiedAPIKeyResponse) {
				assert.NotNil(t, resp.Key)
				assert.Equal(t, "new-secret-xyz789", resp.Secret)
			},
		},
		{
			name:       "not found",
			keyID:      "nonexistent",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "API key not found",
			},
			wantErr: true,
		},
		{
			name:       "revoked key cannot be regenerated",
			keyID:      "key-123",
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Cannot regenerate revoked key",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			keyID:      "key-123",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "forbidden",
			keyID:      "key-123",
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Access denied",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			keyID:      "key-123",
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
				assert.Equal(t, "POST", r.Method)
				assert.Contains(t, r.URL.Path, "/regenerate")

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

			resp, err := client.APIKeys.RegenerateUnified(ctx, tt.keyID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.checkFunc != nil {
					tt.checkFunc(t, resp)
				}
			}
		})
	}
}


// TestAPIKeysService_CreateForOrganizationComprehensive tests the CreateForOrganization method
func TestAPIKeysService_CreateForOrganizationComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		orgID      string
		request    *CreateUnifiedAPIKeyRequest
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *CreateUnifiedAPIKeyResponse)
	}{
		{
			name:  "success - organization key",
			orgID: "org-123",
			request: &CreateUnifiedAPIKeyRequest{
				Name:        "Org API Key",
				Description: "Organization-scoped key",
				Type:        APIKeyTypeUser,
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"key": map[string]interface{}{
						"id":              123,
						"key_id":          "key-123",
						"name":            "Org API Key",
						"organization_id": 123,
					},
					"secret": "secret-org123",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, resp *CreateUnifiedAPIKeyResponse) {
				assert.NotNil(t, resp.Key)
				assert.Equal(t, "secret-org123", resp.Secret)
			},
		},
		{
			name:  "validation error - missing name",
			orgID: "org-123",
			request: &CreateUnifiedAPIKeyRequest{
				Type: APIKeyTypeUser,
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Name is required",
			},
			wantErr: true,
		},
		{
			name:  "not found - organization doesn't exist",
			orgID: "org-nonexistent",
			request: &CreateUnifiedAPIKeyRequest{
				Name: "Test Key",
				Type: APIKeyTypeUser,
			},
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Organization not found",
			},
			wantErr: true,
		},
		{
			name:  "unauthorized",
			orgID: "org-123",
			request: &CreateUnifiedAPIKeyRequest{
				Name: "Test Key",
				Type: APIKeyTypeUser,
			},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:  "forbidden - not org member",
			orgID: "org-123",
			request: &CreateUnifiedAPIKeyRequest{
				Name: "Test Key",
				Type: APIKeyTypeUser,
			},
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Not a member of this organization",
			},
			wantErr: true,
		},
		{
			name:  "server error",
			orgID: "org-123",
			request: &CreateUnifiedAPIKeyRequest{
				Name: "Test Key",
				Type: APIKeyTypeUser,
			},
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
				assert.Equal(t, "POST", r.Method)
				assert.Contains(t, r.URL.Path, "/v2/organizations/")
				assert.Contains(t, r.URL.Path, "/api-keys")

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

			resp, err := client.APIKeys.CreateForOrganization(ctx, tt.orgID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.checkFunc != nil {
					tt.checkFunc(t, resp)
				}
			}
		})
	}
}

// TestAPIKeysService_ListForOrganizationComprehensive tests the ListForOrganization method
func TestAPIKeysService_ListForOrganizationComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		orgID      string
		opts       *ListUnifiedAPIKeysOptions
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, []*UnifiedAPIKey, *PaginationMeta)
	}{
		{
			name:  "success - with options",
			orgID: "org-123",
			opts: &ListUnifiedAPIKeysOptions{
				ListOptions: ListOptions{Page: 1, Limit: 25},
				Type:        APIKeyTypeUser,
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": []map[string]interface{}{
					{"id": 1, "name": "Key 1", "organization_id": 123},
					{"id": 2, "name": "Key 2", "organization_id": 123},
				},
				"meta": map[string]interface{}{
					"page":        1,
					"limit":       25,
					"total_items": 2,
					"total_pages": 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, keys []*UnifiedAPIKey, meta *PaginationMeta) {
				assert.Len(t, keys, 2)
				assert.Equal(t, 1, meta.Page)
			},
		},
		{
			name:  "success - nil options",
			orgID: "org-123",
			opts:  nil,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data":   []map[string]interface{}{},
				"meta": map[string]interface{}{
					"page":        1,
					"limit":       25,
					"total_items": 0,
					"total_pages": 0,
				},
			},
			wantErr: false,
		},
		{
			name:  "not found - organization doesn't exist",
			orgID: "org-nonexistent",
			opts: &ListUnifiedAPIKeysOptions{
				ListOptions: ListOptions{Page: 1, Limit: 25},
			},
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Organization not found",
			},
			wantErr: true,
		},
		{
			name:  "unauthorized",
			orgID: "org-123",
			opts: &ListUnifiedAPIKeysOptions{
				ListOptions: ListOptions{Page: 1, Limit: 25},
			},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:  "forbidden",
			orgID: "org-123",
			opts: &ListUnifiedAPIKeysOptions{
				ListOptions: ListOptions{Page: 1, Limit: 25},
			},
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Access denied",
			},
			wantErr: true,
		},
		{
			name:  "server error",
			orgID: "org-123",
			opts: &ListUnifiedAPIKeysOptions{
				ListOptions: ListOptions{Page: 1, Limit: 25},
			},
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
				assert.Contains(t, r.URL.Path, "/v2/organizations/")
				assert.Contains(t, r.URL.Path, "/api-keys")

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

			keys, meta, err := client.APIKeys.ListForOrganization(ctx, tt.orgID, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, keys)
				assert.NotNil(t, meta)
				if tt.checkFunc != nil {
					tt.checkFunc(t, keys, meta)
				}
			}
		})
	}
}

// TestAPIKeysService_AdminCreateUnifiedComprehensive tests the AdminCreateUnified method
func TestAPIKeysService_AdminCreateUnifiedComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		request    *CreateUnifiedAPIKeyRequest
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *CreateUnifiedAPIKeyResponse)
	}{
		{
			name: "success - admin key creation",
			request: &CreateUnifiedAPIKeyRequest{
				Name:        "Admin Key",
				Description: "Admin API key",
				Type:        APIKeyTypeAdmin,
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"key": map[string]interface{}{
						"id":              123,
						"key_id":          "key-123",
						"name":            "Admin Key",
						"type":            "admin",
						"organization_id": 1,
					},
					"secret": "secret-admin123",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, resp *CreateUnifiedAPIKeyResponse) {
				assert.NotNil(t, resp.Key)
				assert.Equal(t, "secret-admin123", resp.Secret)
			},
		},
		{
			name: "success - registration key",
			request: &CreateUnifiedAPIKeyRequest{
				Name: "Registration Key",
				Type: APIKeyTypeRegistration,
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"key": map[string]interface{}{
						"id":              456,
						"key_id":          "key-456",
						"type":            "registration",
						"organization_id": 1,
					},
					"secret": "secret-reg456",
				},
			},
			wantErr: false,
		},
		{
			name: "validation error - missing name",
			request: &CreateUnifiedAPIKeyRequest{
				Type: APIKeyTypeAdmin,
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Name is required",
			},
			wantErr: true,
		},
		{
			name: "unauthorized",
			request: &CreateUnifiedAPIKeyRequest{
				Name: "Admin Key",
				Type: APIKeyTypeAdmin,
			},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name: "forbidden - not admin",
			request: &CreateUnifiedAPIKeyRequest{
				Name: "Admin Key",
				Type: APIKeyTypeAdmin,
			},
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Admin access required",
			},
			wantErr: true,
		},
		{
			name: "server error",
			request: &CreateUnifiedAPIKeyRequest{
				Name: "Admin Key",
				Type: APIKeyTypeAdmin,
			},
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
				assert.Equal(t, "POST", r.Method)
				assert.Contains(t, r.URL.Path, "/v2/admin/api-keys")

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

			resp, err := client.APIKeys.AdminCreateUnified(ctx, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.checkFunc != nil {
					tt.checkFunc(t, resp)
				}
			}
		})
	}
}

// TestAPIKeysService_AdminListUnifiedComprehensive tests the AdminListUnified method
func TestAPIKeysService_AdminListUnifiedComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		opts       *ListUnifiedAPIKeysOptions
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, []*UnifiedAPIKey, *PaginationMeta)
	}{
		{
			name: "success - all keys",
			opts: &ListUnifiedAPIKeysOptions{
				ListOptions: ListOptions{Page: 1, Limit: 50},
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": []map[string]interface{}{
					{"id": 1, "type": "user"},
					{"id": 2, "type": "admin"},
					{"id": 3, "type": "monitoring_agent"},
				},
				"meta": map[string]interface{}{
					"page":        1,
					"limit":       50,
					"total_items": 3,
					"total_pages": 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, keys []*UnifiedAPIKey, meta *PaginationMeta) {
				assert.Len(t, keys, 3)
				assert.Equal(t, 3, meta.TotalItems)
			},
		},
		{
			name: "success - filter by type",
			opts: &ListUnifiedAPIKeysOptions{
				ListOptions: ListOptions{Page: 1, Limit: 25},
				Type:        APIKeyTypeAdmin,
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": []map[string]interface{}{
					{"id": 1, "type": "admin"},
				},
				"meta": map[string]interface{}{
					"page":        1,
					"limit":       25,
					"total_items": 1,
					"total_pages": 1,
				},
			},
			wantErr: false,
		},
		{
			name: "success - nil options",
			opts: nil,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data":   []map[string]interface{}{},
				"meta": map[string]interface{}{
					"page":        1,
					"limit":       25,
					"total_items": 0,
					"total_pages": 0,
				},
			},
			wantErr: false,
		},
		{
			name: "unauthorized",
			opts: &ListUnifiedAPIKeysOptions{
				ListOptions: ListOptions{Page: 1, Limit: 25},
			},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name: "forbidden - not admin",
			opts: &ListUnifiedAPIKeysOptions{
				ListOptions: ListOptions{Page: 1, Limit: 25},
			},
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Admin access required",
			},
			wantErr: true,
		},
		{
			name: "server error",
			opts: &ListUnifiedAPIKeysOptions{
				ListOptions: ListOptions{Page: 1, Limit: 25},
			},
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
				assert.Contains(t, r.URL.Path, "/v2/admin/api-keys")

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

			keys, meta, err := client.APIKeys.AdminListUnified(ctx, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, keys)
				assert.NotNil(t, meta)
				if tt.checkFunc != nil {
					tt.checkFunc(t, keys, meta)
				}
			}
		})
	}
}

// TestAPIKeysService_ValidateKeyComprehensive tests the ValidateKey method
func TestAPIKeysService_ValidateKeyComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		keyID      string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *UnifiedAPIKey)
	}{
		{
			name:       "success - active key",
			keyID:      "key-123",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":              123,
					"key_id":          "key-123",
					"name":            "Test Key",
					"status":          "active",
					"organization_id": 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, key *UnifiedAPIKey) {
				assert.NotEmpty(t, key.Name)
			},
		},
		{
			name:       "not found",
			keyID:      "nonexistent",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "API key not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			keyID:      "key-123",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			keyID:      "key-123",
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
				assert.Contains(t, r.URL.Path, "/v2/api-keys/")

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

			key, err := client.APIKeys.ValidateKey(ctx, tt.keyID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, key)
				if tt.checkFunc != nil {
					tt.checkFunc(t, key)
				}
			}
		})
	}
}

// TestAPIKeysService_NetworkErrors tests handling of network-level errors
func TestAPIKeysService_NetworkErrors(t *testing.T) {
	tests := []struct {
		name          string
		setupServer   func() string
		setupContext  func() context.Context
		operation     string
		expectError   bool
		errorContains string
	}{
		{
			name: "connection refused - server not listening",
			setupServer: func() string {
				return "http://127.0.0.1:9999"
			},
			setupContext: func() context.Context {
				ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
				return ctx
			},
			operation:     "list",
			expectError:   true,
			errorContains: "connection refused",
		},
		{
			name: "connection timeout - unreachable host",
			setupServer: func() string {
				return "http://192.0.2.1:8080"
			},
			setupContext: func() context.Context {
				ctx, _ := context.WithTimeout(context.Background(), 100*time.Millisecond)
				return ctx
			},
			operation:     "get",
			expectError:   true,
			errorContains: "context deadline exceeded",
		},
		{
			name: "DNS failure - invalid hostname",
			setupServer: func() string {
				return "http://this-domain-does-not-exist-12345.invalid"
			},
			setupContext: func() context.Context {
				ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
				return ctx
			},
			operation:     "create",
			expectError:   true,
			errorContains: "no such host",
		},
		{
			name: "read timeout - server accepts but doesn't respond",
			setupServer: func() string {
				listener, _ := net.Listen("tcp", "127.0.0.1:0")
				go func() {
					defer listener.Close()
					conn, err := listener.Accept()
					if err != nil {
						return
					}
					time.Sleep(5 * time.Second)
					conn.Close()
				}()
				return "http://" + listener.Addr().String()
			},
			setupContext: func() context.Context {
				ctx, _ := context.WithTimeout(context.Background(), 500*time.Millisecond)
				return ctx
			},
			operation:     "update",
			expectError:   true,
			errorContains: "context deadline exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serverURL := tt.setupServer()
			ctx := tt.setupContext()

			client, err := NewClient(&Config{
				BaseURL:    serverURL,
				Auth:       AuthConfig{Token: "test-token"},
				RetryCount: 0,
				Timeout:    2 * time.Second,
			})
			require.NoError(t, err)

			var apiErr error
			switch tt.operation {
			case "list":
				_, _, apiErr = client.APIKeys.List(ctx, nil)
			case "get":
				_, apiErr = client.APIKeys.Get(ctx, "test-id")
			case "create":
				apiKey := &APIKey{Name: "test"}
				_, apiErr = client.APIKeys.Create(ctx, apiKey)
			case "update":
				apiKey := &APIKey{Name: "updated"}
				_, apiErr = client.APIKeys.Update(ctx, "test-id", apiKey)
			}

			if tt.expectError {
				assert.Error(t, apiErr)
				if tt.errorContains != "" {
					assert.Contains(t, apiErr.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, apiErr)
			}
		})
	}
}

// TestAPIKeysService_ConcurrentOperations tests concurrent operations on API keys
func TestAPIKeysService_ConcurrentOperations(t *testing.T) {
	tests := []struct {
		name              string
		concurrencyLevel  int
		operationsPerGoro int
		operation         string
		mockStatus        int
		mockBody          interface{}
	}{
		{
			name:              "concurrent List - low concurrency",
			concurrencyLevel:  10,
			operationsPerGoro: 5,
			operation:         "list",
			mockStatus:        http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": []map[string]interface{}{
					{
						"id":     1,
						"key_id": "key-1",
						"name":   "Test Key",
						"type":   "user",
					},
				},
				"meta": map[string]interface{}{"total_items": 1},
			},
		},
		{
			name:              "concurrent GetUnified - medium concurrency",
			concurrencyLevel:  50,
			operationsPerGoro: 2,
			operation:         "get",
			mockStatus:        http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":              1,
					"key_id":          "key-1",
					"name":            "Test Key",
					"type":            "user",
					"organization_id": 1,
				},
			},
		},
		{
			name:              "concurrent CreateUnified - medium concurrency",
			concurrencyLevel:  30,
			operationsPerGoro: 2,
			operation:         "create",
			mockStatus:        http.StatusCreated,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"key": map[string]interface{}{
						"id":              2,
						"key_id":          "key-2",
						"name":            "New Key",
						"type":            "user",
						"organization_id": 1,
					},
					"secret": "new-secret-123",
				},
			},
		},
		{
			name:              "high concurrency stress - mixed operations",
			concurrencyLevel:  100,
			operationsPerGoro: 1,
			operation:         "list",
			mockStatus:        http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data":   []map[string]interface{}{},
				"meta":   map[string]interface{}{"total_items": 0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			successCount := int64(0)
			errorCount := int64(0)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

			var wg sync.WaitGroup
			startTime := time.Now()

			for i := 0; i < tt.concurrencyLevel; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()

					for j := 0; j < tt.operationsPerGoro; j++ {
						var apiErr error

						switch tt.operation {
						case "list":
							_, _, apiErr = client.APIKeys.ListUnified(context.Background(), nil)
						case "get":
							_, apiErr = client.APIKeys.GetUnified(context.Background(), "key-1")
						case "create":
							req := &CreateUnifiedAPIKeyRequest{Name: "Test Key", Type: APIKeyTypeUser}
							_, apiErr = client.APIKeys.CreateUnified(context.Background(), req)
						case "update":
							req := &UpdateUnifiedAPIKeyRequest{Name: stringPtr("Updated")}
							_, apiErr = client.APIKeys.UpdateUnified(context.Background(), "key-1", req)
						}

						if apiErr != nil {
							atomic.AddInt64(&errorCount, 1)
						} else {
							atomic.AddInt64(&successCount, 1)
						}
					}
				}(i)
			}

			wg.Wait()
			duration := time.Since(startTime)

			totalOps := int64(tt.concurrencyLevel * tt.operationsPerGoro)
			assert.Equal(t, totalOps, successCount+errorCount, "Total operations should equal success + error count")
			assert.Equal(t, int64(0), errorCount, "Expected no errors in concurrent operations")
			assert.Equal(t, totalOps, successCount, "All operations should succeed")

			t.Logf("Completed %d operations in %v (%.2f ops/sec)",
				totalOps, duration, float64(totalOps)/duration.Seconds())
		})
	}
}
