package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAPIKeysService_CreateAPIKey tests API key creation with various scenarios
func TestAPIKeysService_CreateAPIKey(t *testing.T) {
	tests := []struct {
		name       string
		keyReq     *CreateUnifiedAPIKeyRequest
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *CreateUnifiedAPIKeyResponse)
	}{
		{
			name: "success - create user API key",
			keyReq: &CreateUnifiedAPIKeyRequest{
				Name:        "User Key",
				Description: "Test user key",
				Type:        APIKeyTypeUser,
				Capabilities: []string{"servers:read", "metrics:submit"},
			},
			mockStatus: http.StatusCreated,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"key": map[string]interface{}{
						"id":           1,
						"key_id":       "usr_abc123def456",
						"name":         "User Key",
						"type":         "user",
						"status":       "active",
						"capabilities": []string{"servers:read", "metrics:submit"},
					},
					"secret": "secret_xyz789uvw123",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, resp *CreateUnifiedAPIKeyResponse) {
				assert.NotNil(t, resp.Key)
				assert.Equal(t, "User Key", resp.Key.Name)
				assert.Equal(t, APIKeyTypeUser, resp.Key.Type)
				assert.Equal(t, "secret_xyz789uvw123", resp.Secret)
				assert.Len(t, resp.Key.Capabilities, 2)
			},
		},
		{
			name: "success - create monitoring agent key",
			keyReq: &CreateUnifiedAPIKeyRequest{
				Name:           "Agent Key",
				Description:    "Monitoring agent key",
				Type:           APIKeyTypeMonitoringAgent,
				NamespaceName:  "production",
				AgentType:      "prometheus",
				RegionCode:     "us-east-1",
				Capabilities:   []string{"metrics:submit", "health:report"},
			},
			mockStatus: http.StatusCreated,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"key": map[string]interface{}{
						"id":                   1,
						"key_id":               "agent_prom123456",
						"name":                 "Agent Key",
						"type":                 "monitoring_agent",
						"status":               "active",
						"namespace_name":       "production",
						"agent_type":           "prometheus",
						"region_code":          "us-east-1",
						"allowed_probe_scopes": []string{"metrics:submit", "health:report"},
					},
					"secret": "secret_agent789",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, resp *CreateUnifiedAPIKeyResponse) {
				assert.NotNil(t, resp.Key)
				assert.Equal(t, "Agent Key", resp.Key.Name)
				assert.Equal(t, APIKeyTypeMonitoringAgent, resp.Key.Type)
				assert.Equal(t, "prometheus", resp.Key.AgentType)
				assert.Equal(t, "us-east-1", resp.Key.RegionCode)
			},
		},
		{
			name: "success - create registration key",
			keyReq: &CreateUnifiedAPIKeyRequest{
				Name:        "Registration Key",
				Description: "Server registration key",
				Type:        APIKeyTypeRegistration,
				Capabilities: []string{"servers:register", "servers:update"},
			},
			mockStatus: http.StatusCreated,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"key": map[string]interface{}{
						"id":           1,
						"key_id":       "reg_12345678",
						"name":         "Registration Key",
						"type":         "registration",
						"status":       "active",
						"capabilities": []string{"servers:register", "servers:update"},
					},
					"secret": "secret_reg123",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, resp *CreateUnifiedAPIKeyResponse) {
				assert.NotNil(t, resp.Key)
				assert.Equal(t, APIKeyTypeRegistration, resp.Key.Type)
				assert.True(t, resp.Key.CanRegisterServers())
			},
		},
		{
			name: "error - missing required name",
			keyReq: &CreateUnifiedAPIKeyRequest{
				Name:        "",
				Type:        APIKeyTypeUser,
				Capabilities: []string{"servers:read"},
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status": "error",
				"error":  "name is required",
			},
			wantErr: true,
			checkFunc: nil,
		},
		{
			name: "error - invalid key type",
			keyReq: &CreateUnifiedAPIKeyRequest{
				Name: "Invalid Key",
				Type: APIKeyType("invalid_type"),
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status": "error",
				"error":  "invalid key type",
			},
			wantErr: true,
			checkFunc: nil,
		},
		{
			name: "error - unauthorized",
			keyReq: &CreateUnifiedAPIKeyRequest{
				Name: "Unauthorized Key",
				Type: APIKeyTypeUser,
			},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status": "error",
				"error":  "unauthorized",
			},
			wantErr: true,
			checkFunc: nil,
		},
		{
			name: "error - forbidden - insufficient permissions",
			keyReq: &CreateUnifiedAPIKeyRequest{
				Name: "Forbidden Key",
				Type: APIKeyTypeAdmin,
			},
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status": "error",
				"error":  "insufficient permissions to create admin key",
			},
			wantErr: true,
			checkFunc: nil,
		},
		{
			name: "error - conflict - duplicate name",
			keyReq: &CreateUnifiedAPIKeyRequest{
				Name:        "Duplicate Key",
				Type:        APIKeyTypeUser,
				Capabilities: []string{"servers:read"},
			},
			mockStatus: http.StatusConflict,
			mockBody: map[string]interface{}{
				"status": "error",
				"error":  "key name already exists",
			},
			wantErr: true,
			checkFunc: nil,
		},
		{
			name: "error - internal server error",
			keyReq: &CreateUnifiedAPIKeyRequest{
				Name: "Server Error Key",
				Type: APIKeyTypeUser,
			},
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status": "error",
				"error":  "internal server error",
			},
			wantErr: true,
			checkFunc: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
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
			resp, err := client.APIKeys.CreateUnified(ctx, tt.keyReq)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, resp)
				if tt.checkFunc != nil {
					tt.checkFunc(t, resp)
				}
			}
		})
	}
}

// TestAPIKeysService_ListAPIKeys tests API key listing with various filters and pagination
func TestAPIKeysService_ListAPIKeys(t *testing.T) {
	tests := []struct {
		name       string
		opts       *ListUnifiedAPIKeysOptions
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, []*UnifiedAPIKey, *PaginationMeta)
	}{
		{
			name:       "success - list all keys",
			opts:       nil,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": []map[string]interface{}{
					{
						"id":           1,
						"key_id":       "usr_abc123",
						"name":         "User Key 1",
						"type":         "user",
						"status":       "active",
						"capabilities": []string{"servers:read"},
					},
					{
						"id":           2,
						"key_id":       "usr_def456",
						"name":         "User Key 2",
						"type":         "user",
						"status":       "active",
						"capabilities": []string{"servers:read", "metrics:submit"},
					},
				},
				"meta": map[string]interface{}{
					"page":         1,
					"per_page":     25,
					"total_items":  2,
					"total_pages":  1,
					"limit":        25,
					"current_page": 1,
					"first_page":   1,
					"last_page":    1,
					"from":         1,
					"to":           2,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, keys []*UnifiedAPIKey, meta *PaginationMeta) {
				assert.Len(t, keys, 2)
				assert.Equal(t, "User Key 1", keys[0].Name)
				assert.Equal(t, "User Key 2", keys[1].Name)
				if meta != nil {
					assert.Equal(t, 1, meta.Page)
					assert.Equal(t, 25, meta.PerPage)
					assert.Equal(t, 2, meta.TotalItems)
				}
			},
		},
		{
			name: "success - filter by key type",
			opts: &ListUnifiedAPIKeysOptions{
				ListOptions: ListOptions{Page: 1, Limit: 10},
				Type:        APIKeyTypeUser,
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": []map[string]interface{}{
					{
						"id":           1,
						"key_id":       "usr_abc123",
						"name":         "User Key",
						"type":         "user",
						"status":       "active",
						"capabilities": []string{"servers:read"},
					},
				},
				"meta": map[string]interface{}{
					"page":         1,
					"per_page":     10,
					"limit":        10,
					"total_items":  1,
					"total_pages":  1,
					"current_page": 1,
					"first_page":   1,
					"last_page":    1,
					"from":         1,
					"to":           1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, keys []*UnifiedAPIKey, meta *PaginationMeta) {
				assert.Len(t, keys, 1)
				assert.Equal(t, APIKeyTypeUser, keys[0].Type)
			},
		},
		{
			name: "success - filter by status",
			opts: &ListUnifiedAPIKeysOptions{
				ListOptions: ListOptions{Limit: 10},
				Status:      APIKeyStatusActive,
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": []map[string]interface{}{
					{
						"id":     1,
						"key_id": "usr_active",
						"name":   "Active Key",
						"type":   "user",
						"status": "active",
					},
				},
				"meta": map[string]interface{}{
					"page":         1,
					"per_page":     10,
					"limit":        10,
					"total_items":  1,
					"total_pages":  1,
					"current_page": 1,
					"first_page":   1,
					"last_page":    1,
					"from":         1,
					"to":           1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, keys []*UnifiedAPIKey, meta *PaginationMeta) {
				assert.Len(t, keys, 1)
				assert.Equal(t, APIKeyStatusActive, keys[0].Status)
			},
		},
		{
			name: "success - pagination",
			opts: &ListUnifiedAPIKeysOptions{
				ListOptions: ListOptions{Page: 2, Limit: 5},
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": []map[string]interface{}{
					{
						"id":           6,
						"key_id":       "key_6",
						"name":         "Key 6",
						"type":         "user",
						"status":       "active",
						"capabilities": []string{},
					},
				},
				"meta": map[string]interface{}{
					"page":         2,
					"per_page":     5,
					"limit":        5,
					"total_items":  10,
					"total_pages":  2,
					"current_page": 2,
					"first_page":   1,
					"last_page":    2,
					"from":         6,
					"to":           10,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, keys []*UnifiedAPIKey, meta *PaginationMeta) {
				assert.Len(t, keys, 1)
				if meta != nil {
					assert.Equal(t, 2, meta.Page)
					assert.Equal(t, 5, meta.PerPage)
					assert.Equal(t, 10, meta.TotalItems)
					assert.Equal(t, 2, meta.TotalPages)
				}
			},
		},
		{
			name: "success - empty list",
			opts: &ListUnifiedAPIKeysOptions{
				ListOptions: ListOptions{Limit: 25},
				Type:        APIKeyTypeMonitoringAgent,
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data":   []map[string]interface{}{},
				"meta": map[string]interface{}{
					"page":         1,
					"per_page":     25,
					"limit":        25,
					"total_items":  0,
					"total_pages":  0,
					"current_page": 1,
					"first_page":   1,
					"last_page":    0,
					"from":         0,
					"to":           0,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, keys []*UnifiedAPIKey, meta *PaginationMeta) {
				assert.Len(t, keys, 0)
			},
		},
		{
			name:       "error - unauthorized",
			opts:       nil,
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status": "error",
				"error":  "unauthorized",
			},
			wantErr: true,
			checkFunc: nil,
		},
		{
			name:       "error - forbidden",
			opts:       nil,
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status": "error",
				"error":  "access denied",
			},
			wantErr: true,
			checkFunc: nil,
		},
		{
			name:       "error - internal server error",
			opts:       nil,
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status": "error",
				"error":  "internal server error",
			},
			wantErr: true,
			checkFunc: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
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
			keys, meta, err := client.APIKeys.ListUnified(ctx, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, keys)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, keys)
				if tt.checkFunc != nil {
					tt.checkFunc(t, keys, meta)
				}
			}
		})
	}
}

// TestAPIKeysService_RevokeAPIKey tests API key revocation
func TestAPIKeysService_RevokeAPIKey(t *testing.T) {
	tests := []struct {
		name       string
		keyID      string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:       "success - revoke active key",
			keyID:      "usr_abc123",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data":   map[string]interface{}{},
			},
			wantErr: false,
		},
		{
			name:       "success - revoke already revoked key",
			keyID:      "usr_already_revoked",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data":   map[string]interface{}{},
			},
			wantErr: false,
		},
		{
			name:       "error - key not found",
			keyID:      "nonexistent_key",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status": "error",
				"error":  "key not found",
			},
			wantErr: true,
		},
		{
			name:       "error - unauthorized",
			keyID:      "usr_abc123",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status": "error",
				"error":  "unauthorized",
			},
			wantErr: true,
		},
		{
			name:       "error - forbidden - cannot revoke key",
			keyID:      "usr_abc123",
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status": "error",
				"error":  "insufficient permissions to revoke this key",
			},
			wantErr: true,
		},
		{
			name:       "error - internal server error",
			keyID:      "usr_abc123",
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status": "error",
				"error":  "internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Contains(t, r.URL.Path, tt.keyID)
				assert.Contains(t, r.URL.Path, "revoke")
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
			err = client.APIKeys.RevokeUnified(ctx, tt.keyID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestAPIKeysService_RegenerateAPIKey tests API key regeneration
func TestAPIKeysService_RegenerateAPIKey(t *testing.T) {
	tests := []struct {
		name       string
		keyID      string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *CreateUnifiedAPIKeyResponse)
	}{
		{
			name:       "success - regenerate active key",
			keyID:      "usr_abc123",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"key": map[string]interface{}{
						"id":           1,
						"key_id":       "usr_abc123",
						"name":         "Regenerated Key",
						"type":         "user",
						"status":       "active",
						"capabilities": []string{"servers:read"},
					},
					"secret": "new_secret_xyz789",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, resp *CreateUnifiedAPIKeyResponse) {
				assert.NotNil(t, resp.Key)
				assert.Equal(t, "usr_abc123", resp.Key.KeyID)
				assert.Equal(t, "Regenerated Key", resp.Key.Name)
				assert.Equal(t, "new_secret_xyz789", resp.Secret)
			},
		},
		{
			name:       "success - regenerate monitoring agent key",
			keyID:      "agent_prom123456",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"key": map[string]interface{}{
						"id":             1,
						"key_id":         "agent_prom123456",
						"name":           "Monitoring Agent",
						"type":           "monitoring_agent",
						"status":         "active",
						"namespace_name": "production",
						"agent_type":     "prometheus",
						"region_code":    "us-east-1",
					},
					"secret": "agent_secret_new",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, resp *CreateUnifiedAPIKeyResponse) {
				assert.NotNil(t, resp.Key)
				assert.Equal(t, APIKeyTypeMonitoringAgent, resp.Key.Type)
				assert.Equal(t, "agent_secret_new", resp.Secret)
			},
		},
		{
			name:       "error - key not found",
			keyID:      "nonexistent_key",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status": "error",
				"error":  "key not found",
			},
			wantErr: true,
			checkFunc: nil,
		},
		{
			name:       "error - cannot regenerate revoked key",
			keyID:      "usr_revoked",
			mockStatus: http.StatusConflict,
			mockBody: map[string]interface{}{
				"status": "error",
				"error":  "cannot regenerate revoked key",
			},
			wantErr: true,
			checkFunc: nil,
		},
		{
			name:       "error - unauthorized",
			keyID:      "usr_abc123",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status": "error",
				"error":  "unauthorized",
			},
			wantErr: true,
			checkFunc: nil,
		},
		{
			name:       "error - forbidden",
			keyID:      "usr_abc123",
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status": "error",
				"error":  "insufficient permissions to regenerate this key",
			},
			wantErr: true,
			checkFunc: nil,
		},
		{
			name:       "error - internal server error",
			keyID:      "usr_abc123",
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status": "error",
				"error":  "internal server error",
			},
			wantErr: true,
			checkFunc: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Contains(t, r.URL.Path, tt.keyID)
				assert.Contains(t, r.URL.Path, "regenerate")
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
			resp, err := client.APIKeys.RegenerateUnified(ctx, tt.keyID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, resp)
				if tt.checkFunc != nil {
					tt.checkFunc(t, resp)
				}
			}
		})
	}
}
