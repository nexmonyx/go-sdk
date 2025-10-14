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

// =============================================================================
// Unified API Keys Tests
// =============================================================================

func TestAPIKeysService_CreateUnified(t *testing.T) {
	tests := []struct {
		name           string
		request        *CreateUnifiedAPIKeyRequest
		responseStatus int
		responseBody   interface{}
		wantErr        bool
		validateResult func(*testing.T, *CreateUnifiedAPIKeyResponse)
	}{
		{
			name: "successful user key creation",
			request: &CreateUnifiedAPIKeyRequest{
				Name:         "Test User Key",
				Description:  "Test Description",
				Type:         APIKeyTypeUser,
				Capabilities: []string{"read", "write"},
			},
			responseStatus: http.StatusOK,
			responseBody: StandardResponse{
				Status:  "success",
				Message: "API key created successfully",
				Data: &CreateUnifiedAPIKeyResponse{
					Key: &UnifiedAPIKey{
						GormModel: GormModel{ID: 1},
						Name:      "Test User Key",
						Type:      APIKeyTypeUser,
						Status:    APIKeyStatusActive,
					},
					KeyValue: "nxm_test_key_123456",
				},
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *CreateUnifiedAPIKeyResponse) {
				assert.NotNil(t, result)
				assert.NotNil(t, result.Key)
				assert.Equal(t, "Test User Key", result.Key.Name)
				assert.Equal(t, APIKeyTypeUser, result.Key.Type)
				assert.Equal(t, "nxm_test_key_123456", result.KeyValue)
			},
		},
		{
			name: "successful monitoring agent key creation",
			request: &CreateUnifiedAPIKeyRequest{
				Name:         "Monitoring Agent",
				Description:  "Regional monitoring agent",
				Type:         APIKeyTypeMonitoringAgent,
				NamespaceName:    "monitoring",
				AgentType:    "probe-executor",
				RegionCode:   "us-east-1",
				Capabilities: []string{"submit_metrics", "fetch_probes"},
			},
			responseStatus: http.StatusOK,
			responseBody: StandardResponse{
				Status: "success",
				Data: &CreateUnifiedAPIKeyResponse{
					Key: &UnifiedAPIKey{
						GormModel:  GormModel{ID: 2},
						Name:       "Monitoring Agent",
						Type:       APIKeyTypeMonitoringAgent,
						NamespaceName:  "monitoring",
						AgentType:  "probe-executor",
						RegionCode: "us-east-1",
						Status:     APIKeyStatusActive,
					},
					KeyValue: "nxm_monitoring_agent_key",
				},
			},
			validateResult: func(t *testing.T, result *CreateUnifiedAPIKeyResponse) {
				assert.Equal(t, APIKeyTypeMonitoringAgent, result.Key.Type)
				assert.Equal(t, "monitoring", result.Key.NamespaceName)
				assert.Equal(t, "us-east-1", result.Key.RegionCode)
			},
		},
		{
			name: "api error response",
			request: &CreateUnifiedAPIKeyRequest{
				Name: "Invalid Key",
				Type: APIKeyTypeUser,
			},
			responseStatus: http.StatusBadRequest,
			responseBody: ErrorResponse{
				Status:  "error",
				Message: "validation failed",
				Error:   "description is required",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/v2/api-keys", r.URL.Path)

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL})
			result, err := client.APIKeys.CreateUnified(context.Background(), tt.request)

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

func TestAPIKeysService_GetUnified(t *testing.T) {
	tests := []struct {
		name           string
		keyID          string
		responseStatus int
		responseBody   interface{}
		wantErr        bool
		expectedKey    *UnifiedAPIKey
	}{
		{
			name:           "successful get",
			keyID:          "key-123",
			responseStatus: http.StatusOK,
			responseBody: StandardResponse{
				Status: "success",
				Data: &UnifiedAPIKey{
					GormModel: GormModel{ID: 1},
					KeyID:     "key-123",
					Name:      "Test Key",
					Type:      APIKeyTypeUser,
					Status:    APIKeyStatusActive,
				},
			},
			expectedKey: &UnifiedAPIKey{
				GormModel: GormModel{ID: 1},
				KeyID:     "key-123",
				Name:      "Test Key",
				Type:      APIKeyTypeUser,
				Status:    APIKeyStatusActive,
			},
		},
		{
			name:           "key not found",
			keyID:          "nonexistent",
			responseStatus: http.StatusNotFound,
			responseBody: ErrorResponse{
				Status:  "error",
				Message: "key not found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, tt.keyID)

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL})
			result, err := client.APIKeys.GetUnified(context.Background(), tt.keyID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedKey.Name, result.Name)
				assert.Equal(t, tt.expectedKey.Type, result.Type)
			}
		})
	}
}

func TestAPIKeysService_ListUnified(t *testing.T) {
	tests := []struct {
		name           string
		options        *ListUnifiedAPIKeysOptions
		responseStatus int
		responseBody   interface{}
		wantErr        bool
		validateResult func(*testing.T, []*UnifiedAPIKey, *PaginationMeta)
	}{
		{
			name: "list all keys with pagination",
			options: &ListUnifiedAPIKeysOptions{
				ListOptions: ListOptions{
					Page:  1,
					Limit: 10,
				},
			},
			responseStatus: http.StatusOK,
			responseBody: PaginatedResponse{
				Status:  "success",
				Message: "Keys retrieved successfully",
				Data: &[]*UnifiedAPIKey{
					{GormModel: GormModel{ID: 1}, Name: "Key 1", Type: APIKeyTypeUser},
					{GormModel: GormModel{ID: 2}, Name: "Key 2", Type: APIKeyTypeAdmin},
				},
				Meta: &PaginationMeta{
					TotalItems:       2,
					Page:        1,
					Limit:       10,
					TotalPages:  1,
					HasMore: false,
				},
			},
			validateResult: func(t *testing.T, keys []*UnifiedAPIKey, meta *PaginationMeta) {
				assert.Len(t, keys, 2)
				assert.Equal(t, 2, meta.TotalItems)
				assert.Equal(t, 1, meta.Page)
			},
		},
		{
			name: "filter by type",
			options: &ListUnifiedAPIKeysOptions{
				ListOptions: ListOptions{Page: 1, Limit: 10},
				Type:        APIKeyTypeMonitoringAgent,
			},
			responseStatus: http.StatusOK,
			responseBody: PaginatedResponse{
				Status: "success",
				Data: &[]*UnifiedAPIKey{
					{GormModel: GormModel{ID: 1}, Name: "Agent 1", Type: APIKeyTypeMonitoringAgent},
				},
				Meta: &PaginationMeta{TotalItems: 1, Page: 1, Limit: 10},
			},
			validateResult: func(t *testing.T, keys []*UnifiedAPIKey, meta *PaginationMeta) {
				assert.Len(t, keys, 1)
				assert.Equal(t, APIKeyTypeMonitoringAgent, keys[0].Type)
			},
		},
		{
			name: "filter by status",
			options: &ListUnifiedAPIKeysOptions{
				ListOptions: ListOptions{Page: 1, Limit: 10},
				Status:      APIKeyStatusActive,
			},
			responseStatus: http.StatusOK,
			responseBody: PaginatedResponse{
				Status: "success",
				Data: &[]*UnifiedAPIKey{
					{GormModel: GormModel{ID: 1}, Status: APIKeyStatusActive},
				},
				Meta: &PaginationMeta{TotalItems: 1},
			},
			validateResult: func(t *testing.T, keys []*UnifiedAPIKey, meta *PaginationMeta) {
				assert.Equal(t, APIKeyStatusActive, keys[0].Status)
			},
		},
		{
			name: "filter by region",
			options: &ListUnifiedAPIKeysOptions{
				ListOptions: ListOptions{Page: 1, Limit: 10},
				RegionCode:  "us-west-2",
			},
			responseStatus: http.StatusOK,
			responseBody: PaginatedResponse{
				Status: "success",
				Data: &[]*UnifiedAPIKey{
					{GormModel: GormModel{ID: 1}, RegionCode: "us-west-2"},
				},
				Meta: &PaginationMeta{TotalItems: 1},
			},
			validateResult: func(t *testing.T, keys []*UnifiedAPIKey, meta *PaginationMeta) {
				assert.Equal(t, "us-west-2", keys[0].RegionCode)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/v2/api-keys", r.URL.Path)

				// Validate query parameters
				if tt.options != nil {
					if tt.options.Type != "" {
						assert.Equal(t, string(tt.options.Type), r.URL.Query().Get("type"))
					}
					if tt.options.Status != "" {
						assert.Equal(t, string(tt.options.Status), r.URL.Query().Get("status"))
					}
					if tt.options.RegionCode != "" {
						assert.Equal(t, tt.options.RegionCode, r.URL.Query().Get("region_code"))
					}
				}

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL})
			keys, meta, err := client.APIKeys.ListUnified(context.Background(), tt.options)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, keys, meta)
				}
			}
		})
	}
}

func TestAPIKeysService_UpdateUnified(t *testing.T) {
	tests := []struct {
		name           string
		keyID          string
		request        *UpdateUnifiedAPIKeyRequest
		responseStatus int
		responseBody   interface{}
		wantErr        bool
	}{
		{
			name:  "successful update",
			keyID: "key-123",
			request: &UpdateUnifiedAPIKeyRequest{
				Name:        stringPtr("Updated Name"),
				Description: stringPtr("Updated Description"),
			},
			responseStatus: http.StatusOK,
			responseBody: StandardResponse{
				Status: "success",
				Data: &UnifiedAPIKey{
					GormModel:   GormModel{ID: 1},
					KeyID:       "key-123",
					Name:        "Updated Name",
					Description: "Updated Description",
				},
			},
		},
		{
			name:  "update capabilities",
			keyID: "key-456",
			request: &UpdateUnifiedAPIKeyRequest{
				Capabilities: []string{"read", "write", "delete"},
			},
			responseStatus: http.StatusOK,
			responseBody: StandardResponse{
				Status: "success",
				Data: &UnifiedAPIKey{
					GormModel:    GormModel{ID: 2},
					KeyID:        "key-456",
					Capabilities: []string{"read", "write", "delete"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "PUT", r.Method)
				assert.Contains(t, r.URL.Path, tt.keyID)

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL})
			result, err := client.APIKeys.UpdateUnified(context.Background(), tt.keyID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestAPIKeysService_DeleteUnified(t *testing.T) {
	tests := []struct {
		name           string
		keyID          string
		responseStatus int
		wantErr        bool
	}{
		{
			name:           "successful delete",
			keyID:          "key-123",
			responseStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "key not found",
			keyID:          "nonexistent",
			responseStatus: http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "DELETE", r.Method)
				assert.Contains(t, r.URL.Path, tt.keyID)

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(StandardResponse{Status: "success"})
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL})
			err := client.APIKeys.DeleteUnified(context.Background(), tt.keyID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAPIKeysService_RevokeUnified(t *testing.T) {
	tests := []struct {
		name           string
		keyID          string
		responseStatus int
		wantErr        bool
	}{
		{
			name:           "successful revoke",
			keyID:          "key-123",
			responseStatus: http.StatusOK,
		},
		{
			name:           "already revoked",
			keyID:          "key-456",
			responseStatus: http.StatusConflict,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Contains(t, r.URL.Path, "revoke")

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(StandardResponse{Status: "success"})
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL})
			err := client.APIKeys.RevokeUnified(context.Background(), tt.keyID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAPIKeysService_RegenerateUnified(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.URL.Path, "regenerate")

		json.NewEncoder(w).Encode(StandardResponse{
			Status: "success",
			Data: &CreateUnifiedAPIKeyResponse{
				Key: &UnifiedAPIKey{
					GormModel: GormModel{ID: 1},
					KeyID:     "key-123",
					Name:      "Regenerated Key",
				},
				KeyValue: "nxm_new_key_789",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, err := client.APIKeys.RegenerateUnified(context.Background(), "key-123")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "nxm_new_key_789", result.KeyValue)
}

// =============================================================================
// Organization-scoped Operations
// =============================================================================

func TestAPIKeysService_CreateForOrganization(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.URL.Path, "/organizations/")
		assert.Contains(t, r.URL.Path, "/api-keys")

		json.NewEncoder(w).Encode(StandardResponse{
			Status: "success",
			Data: &CreateUnifiedAPIKeyResponse{
				Key: &UnifiedAPIKey{
					GormModel:      GormModel{ID: 1},
					Name:           "Org Key",
					OrganizationID: 123,
				},
				KeyValue: "org_key_123",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, err := client.APIKeys.CreateForOrganization(context.Background(), "org-123", &CreateUnifiedAPIKeyRequest{
		Name: "Org Key",
		Type: APIKeyTypeUser,
	})

	require.NoError(t, err)
	assert.Equal(t, "Org Key", result.Key.Name)
}

func TestAPIKeysService_ListForOrganization(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/organizations/")

		json.NewEncoder(w).Encode(PaginatedResponse{
			Status: "success",
			Data: &[]*UnifiedAPIKey{
				{GormModel: GormModel{ID: 1}, Name: "Org Key 1"},
			},
			Meta: &PaginationMeta{TotalItems: 1},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	keys, meta, err := client.APIKeys.ListForOrganization(context.Background(), "org-123", nil)

	require.NoError(t, err)
	assert.Len(t, keys, 1)
	assert.Equal(t, 1, meta.TotalItems)
}

// =============================================================================
// Admin Operations
// =============================================================================

func TestAPIKeysService_AdminCreateUnified(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v2/admin/api-keys", r.URL.Path)

		json.NewEncoder(w).Encode(StandardResponse{
			Status: "success",
			Data: &CreateUnifiedAPIKeyResponse{
				Key: &UnifiedAPIKey{
					GormModel: GormModel{ID: 1},
					Name:      "Admin Key",
					Type:      APIKeyTypeAdmin,
				},
				KeyValue: "admin_key_secret",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, err := client.APIKeys.AdminCreateUnified(context.Background(), &CreateUnifiedAPIKeyRequest{
		Name: "Admin Key",
		Type: APIKeyTypeAdmin,
	})

	require.NoError(t, err)
	assert.Equal(t, APIKeyTypeAdmin, result.Key.Type)
}

func TestAPIKeysService_AdminListUnified(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v2/admin/api-keys", r.URL.Path)

		json.NewEncoder(w).Encode(PaginatedResponse{
			Status: "success",
			Data: &[]*UnifiedAPIKey{
				{GormModel: GormModel{ID: 1}, Type: APIKeyTypeAdmin},
				{GormModel: GormModel{ID: 2}, Type: APIKeyTypeUser},
			},
			Meta: &PaginationMeta{TotalItems: 2},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	keys, meta, err := client.APIKeys.AdminListUnified(context.Background(), nil)

	require.NoError(t, err)
	assert.Len(t, keys, 2)
	assert.Equal(t, 2, meta.TotalItems)
}

// =============================================================================
// Legacy API Operations
// =============================================================================

func TestAPIKeysService_Create_Legacy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(StandardResponse{
			Status: "success",
			Data: &CreateUnifiedAPIKeyResponse{
				Key: &UnifiedAPIKey{
					GormModel: GormModel{ID: 1},
					Name:      "Legacy Key",
				},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, err := client.APIKeys.Create(context.Background(), &APIKey{
		Name:        "Legacy Key",
		Description: "Legacy Description",
	})

	require.NoError(t, err)
	assert.Equal(t, "Legacy Key", result.Name)
}

func TestAPIKeysService_Get_Legacy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(StandardResponse{
			Status: "success",
			Data: &UnifiedAPIKey{
				GormModel: GormModel{ID: 1},
				KeyID:     "key-123",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, err := client.APIKeys.Get(context.Background(), "key-123")

	require.NoError(t, err)
	assert.Equal(t, "key-123", result.KeyID)
}

func TestAPIKeysService_List_Legacy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(PaginatedResponse{
			Status: "success",
			Data:   &[]*UnifiedAPIKey{{GormModel: GormModel{ID: 1}}},
			Meta:   &PaginationMeta{TotalItems: 1},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	keys, meta, err := client.APIKeys.List(context.Background(), &ListOptions{Page: 1, Limit: 10})

	require.NoError(t, err)
	assert.Len(t, keys, 1)
	assert.Equal(t, 1, meta.TotalItems)
}

func TestAPIKeysService_Update_Legacy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(StandardResponse{
			Status: "success",
			Data: &UnifiedAPIKey{
				GormModel: GormModel{ID: 1},
				Name:      "Updated",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, err := client.APIKeys.Update(context.Background(), "key-123", &APIKey{Name: "Updated"})

	require.NoError(t, err)
	assert.Equal(t, "Updated", result.Name)
}

func TestAPIKeysService_Delete_Legacy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		json.NewEncoder(w).Encode(StandardResponse{Status: "success"})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	err := client.APIKeys.Delete(context.Background(), "key-123")

	assert.NoError(t, err)
}

func TestAPIKeysService_Revoke_Legacy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		json.NewEncoder(w).Encode(StandardResponse{Status: "success"})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	err := client.APIKeys.Revoke(context.Background(), "key-123")

	assert.NoError(t, err)
}

func TestAPIKeysService_Regenerate_Legacy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(StandardResponse{
			Status: "success",
			Data: &CreateUnifiedAPIKeyResponse{
				Key: &UnifiedAPIKey{
					GormModel: GormModel{ID: 1},
				},
				KeyValue: "new_key",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, err := client.APIKeys.Regenerate(context.Background(), "key-123")

	require.NoError(t, err)
	assert.NotNil(t, result)
}

// =============================================================================
// Specialized Key Creation Helpers
// =============================================================================

func TestAPIKeysService_CreateUserKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(StandardResponse{
			Status: "success",
			Data: &CreateUnifiedAPIKeyResponse{
				Key: &UnifiedAPIKey{
					GormModel: GormModel{ID: 1},
					Name:      "User Key",
					Type:      APIKeyTypeUser,
				},
				KeyValue: "user_key_123",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, err := client.APIKeys.CreateUserKey(context.Background(), "User Key", "Description", []string{"read"})

	require.NoError(t, err)
	assert.Equal(t, APIKeyTypeUser, result.Key.Type)
}

func TestAPIKeysService_CreateAdminKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(StandardResponse{
			Status: "success",
			Data: &CreateUnifiedAPIKeyResponse{
				Key: &UnifiedAPIKey{
					GormModel: GormModel{ID: 1},
					Type:      APIKeyTypeAdmin,
				},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, err := client.APIKeys.CreateAdminKey(context.Background(), "Admin Key", "Description", []string{"*"}, 123)

	require.NoError(t, err)
	assert.Equal(t, APIKeyTypeAdmin, result.Key.Type)
}

func TestAPIKeysService_CreateMonitoringAgentKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(StandardResponse{
			Status: "success",
			Data: &CreateUnifiedAPIKeyResponse{
				Key: &UnifiedAPIKey{
					GormModel:  GormModel{ID: 1},
					Type:       APIKeyTypeMonitoringAgent,
					NamespaceName:  "monitoring",
					AgentType:  "probe-executor",
					RegionCode: "us-east-1",
				},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, err := client.APIKeys.CreateMonitoringAgentKey(
		context.Background(),
		"Agent Key",
		"Description",
		"monitoring",
		"probe-executor",
		"us-east-1",
		[]string{"submit_metrics"},
	)

	require.NoError(t, err)
	assert.Equal(t, APIKeyTypeMonitoringAgent, result.Key.Type)
	assert.Equal(t, "us-east-1", result.Key.RegionCode)
}

func TestAPIKeysService_CreateRegistrationKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(StandardResponse{
			Status: "success",
			Data: &CreateUnifiedAPIKeyResponse{
				Key: &UnifiedAPIKey{
					GormModel: GormModel{ID: 1},
					Type:      APIKeyTypeRegistration,
				},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, err := client.APIKeys.CreateRegistrationKey(context.Background(), "Registration Key", "Description", 123)

	require.NoError(t, err)
	assert.Equal(t, APIKeyTypeRegistration, result.Key.Type)
}

// =============================================================================
// Key Validation and Information Helpers
// =============================================================================

func TestAPIKeysService_ValidateKey(t *testing.T) {
	tests := []struct {
		name        string
		keyStatus   APIKeyStatus
		wantErr     bool
		errContains string
	}{
		{
			name:      "active key valid",
			keyStatus: APIKeyStatusActive,
			wantErr:   false,
		},
		{
			name:        "revoked key invalid",
			keyStatus:   APIKeyStatusRevoked,
			wantErr:     true,
			errContains: "not active",
		},
		{
			name:        "expired key invalid",
			keyStatus:   APIKeyStatusExpired,
			wantErr:     true,
			errContains: "not active",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				json.NewEncoder(w).Encode(StandardResponse{
					Status: "success",
					Data: &UnifiedAPIKey{
						GormModel: GormModel{ID: 1},
						Status:    tt.keyStatus,
					},
				})
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL})
			result, err := client.APIKeys.ValidateKey(context.Background(), "key-123")

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestAPIKeysService_GetKeysByType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, string(APIKeyTypeMonitoringAgent), r.URL.Query().Get("type"))
		json.NewEncoder(w).Encode(PaginatedResponse{
			Status: "success",
			Data:   &[]*UnifiedAPIKey{{GormModel: GormModel{ID: 1}, Type: APIKeyTypeMonitoringAgent}},
			Meta:   &PaginationMeta{TotalItems: 1},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	keys, meta, err := client.APIKeys.GetKeysByType(context.Background(), APIKeyTypeMonitoringAgent, &ListOptions{})

	require.NoError(t, err)
	assert.Len(t, keys, 1)
	assert.Equal(t, APIKeyTypeMonitoringAgent, keys[0].Type)
	assert.Equal(t, 1, meta.TotalItems)
}

func TestAPIKeysService_GetActiveKeys(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, string(APIKeyStatusActive), r.URL.Query().Get("status"))
		json.NewEncoder(w).Encode(PaginatedResponse{
			Status: "success",
			Data:   &[]*UnifiedAPIKey{{GormModel: GormModel{ID: 1}, Status: APIKeyStatusActive}},
			Meta:   &PaginationMeta{TotalItems: 1},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	keys, _, err := client.APIKeys.GetActiveKeys(context.Background(), &ListOptions{})

	require.NoError(t, err)
	assert.Len(t, keys, 1)
	assert.Equal(t, APIKeyStatusActive, keys[0].Status)
}

func TestAPIKeysService_GetMonitoringAgentKeys(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/organizations/")
		json.NewEncoder(w).Encode(PaginatedResponse{
			Status: "success",
			Data:   &[]*UnifiedAPIKey{{GormModel: GormModel{ID: 1}, Type: APIKeyTypeMonitoringAgent}},
			Meta:   &PaginationMeta{TotalItems: 1},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	keys, _, err := client.APIKeys.GetMonitoringAgentKeys(context.Background(), "org-123", &ListOptions{})

	require.NoError(t, err)
	assert.Len(t, keys, 1)
	assert.Equal(t, APIKeyTypeMonitoringAgent, keys[0].Type)
}

func TestAPIKeysService_GetRegistrationKeys(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/admin/api-keys", r.URL.Path)
		assert.Equal(t, string(APIKeyTypeRegistration), r.URL.Query().Get("type"))
		json.NewEncoder(w).Encode(PaginatedResponse{
			Status: "success",
			Data:   &[]*UnifiedAPIKey{{GormModel: GormModel{ID: 1}, Type: APIKeyTypeRegistration}},
			Meta:   &PaginationMeta{TotalItems: 1},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	keys, _, err := client.APIKeys.GetRegistrationKeys(context.Background(), &ListOptions{})

	require.NoError(t, err)
	assert.Len(t, keys, 1)
	assert.Equal(t, APIKeyTypeRegistration, keys[0].Type)
}

// =============================================================================
// Helper Functions
// =============================================================================

