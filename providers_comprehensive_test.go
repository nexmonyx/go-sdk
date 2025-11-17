package nexmonyx

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProvidersService_ListComprehensive tests the List method
func TestProvidersService_ListComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		opts       *ProviderListOptions
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *ProviderListResponse, *PaginationMeta)
	}{
		{
			name: "success - list all providers",
			opts: &ProviderListOptions{Page: 1, PageSize: 25},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"providers": []map[string]interface{}{
						{
							"id":            "provider-aws-1",
							"name":          "AWS Production",
							"provider_type": "aws",
							"status":        "active",
							"vm_count":      150,
							"region":        "us-east-1",
							"created_at":    "2024-01-15T10:00:00Z",
						},
						{
							"id":            "provider-azure-1",
							"name":          "Azure Staging",
							"provider_type": "azure",
							"status":        "active",
							"vm_count":      75,
							"region":        "eastus",
						},
					},
					"total":       2,
					"page":        1,
					"page_size":   25,
					"total_pages": 1,
				},
				"meta": map[string]interface{}{
					"page":        1,
					"limit":       25,
					"total_items": 2,
					"total_pages": 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, resp *ProviderListResponse, meta *PaginationMeta) {
				assert.Len(t, resp.Providers, 2)
				assert.Equal(t, "AWS Production", resp.Providers[0].Name)
				assert.Equal(t, "aws", resp.Providers[0].ProviderType)
				assert.Equal(t, 150, resp.Providers[0].VMCount)
				assert.Equal(t, 2, resp.Total)
			},
		},
		{
			name:       "success - empty list",
			opts:       &ProviderListOptions{Page: 1, PageSize: 25},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"providers":   []map[string]interface{}{},
					"total":       0,
					"page":        1,
					"page_size":   25,
					"total_pages": 0,
				},
				"meta": map[string]interface{}{
					"page":        1,
					"limit":       25,
					"total_items": 0,
					"total_pages": 0,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, resp *ProviderListResponse, meta *PaginationMeta) {
				assert.Len(t, resp.Providers, 0)
				assert.Equal(t, 0, resp.Total)
			},
		},
		{
			name: "success - with type filter",
			opts: &ProviderListOptions{Page: 1, PageSize: 10, Type: "aws"},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"providers": []map[string]interface{}{
						{"id": "provider-aws-1", "name": "AWS Production", "provider_type": "aws", "status": "active", "vm_count": 150},
					},
					"total":       1,
					"page":        1,
					"page_size":   10,
					"total_pages": 1,
				},
				"meta": map[string]interface{}{
					"page":        1,
					"limit":       10,
					"total_items": 1,
					"total_pages": 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, resp *ProviderListResponse, meta *PaginationMeta) {
				assert.Len(t, resp.Providers, 1)
				assert.Equal(t, "aws", resp.Providers[0].ProviderType)
			},
		},
		{
			name:       "unauthorized",
			opts:       &ProviderListOptions{Page: 1, PageSize: 25},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			opts:       &ProviderListOptions{Page: 1, PageSize: 25},
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
				assert.Equal(t, "/v1/providers", r.URL.Path)

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

			providers, meta, err := client.Providers.List(ctx, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, providers)
				if tt.checkFunc != nil {
					tt.checkFunc(t, providers, meta)
				}
			}
		})
	}
}

// TestProvidersService_CreateComprehensive tests the Create method
func TestProvidersService_CreateComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		request    *ProviderCreateRequest
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Provider)
	}{
		{
			name: "success - create AWS provider",
			request: &ProviderCreateRequest{
				Name:         "AWS Production",
				ProviderType: "aws",
				Region:       "us-east-1",
				Description:  "Production AWS environment",
				Credentials: map[string]interface{}{
					"access_key_id":     "AKIA...",
					"secret_access_key": "secret...",
				},
			},
			mockStatus: http.StatusCreated,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":            "provider-aws-1",
					"name":          "AWS Production",
					"provider_type": "aws",
					"status":        "active",
					"vm_count":      0,
					"region":        "us-east-1",
					"description":   "Production AWS environment",
					"created_at":    "2024-01-15T10:00:00Z",
					"updated_at":    "2024-01-15T10:00:00Z",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, provider *Provider) {
				assert.Equal(t, "provider-aws-1", provider.ID)
				assert.Equal(t, "AWS Production", provider.Name)
				assert.Equal(t, "aws", provider.ProviderType)
				assert.Equal(t, "active", provider.Status)
			},
		},
		{
			name: "success - create Azure provider",
			request: &ProviderCreateRequest{
				Name:         "Azure Staging",
				ProviderType: "azure",
				Region:       "eastus",
				Credentials: map[string]interface{}{
					"tenant_id":       "tenant-123",
					"client_id":       "client-456",
					"client_secret":   "secret-789",
					"subscription_id": "sub-012",
				},
			},
			mockStatus: http.StatusCreated,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":            "provider-azure-1",
					"name":          "Azure Staging",
					"provider_type": "azure",
					"status":        "active",
					"vm_count":      0,
					"region":        "eastus",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, provider *Provider) {
				assert.Equal(t, "Azure Staging", provider.Name)
				assert.Equal(t, "azure", provider.ProviderType)
			},
		},
		{
			name: "validation error - missing name",
			request: &ProviderCreateRequest{
				ProviderType: "aws",
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Provider name is required",
			},
			wantErr: true,
		},
		{
			name: "validation error - missing provider type",
			request: &ProviderCreateRequest{
				Name: "Test Provider",
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Provider type is required",
			},
			wantErr: true,
		},
		{
			name: "validation error - invalid provider type",
			request: &ProviderCreateRequest{
				Name:         "Test Provider",
				ProviderType: "invalid-type",
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Invalid provider type",
			},
			wantErr: true,
		},
		{
			name: "validation error - missing credentials",
			request: &ProviderCreateRequest{
				Name:         "AWS Production",
				ProviderType: "aws",
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Provider credentials are required",
			},
			wantErr: true,
		},
		{
			name: "conflict - provider name already exists",
			request: &ProviderCreateRequest{
				Name:         "AWS Production",
				ProviderType: "aws",
				Credentials:  map[string]interface{}{"key": "value"},
			},
			mockStatus: http.StatusConflict,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Provider with this name already exists",
			},
			wantErr: true,
		},
		{
			name: "unauthorized",
			request: &ProviderCreateRequest{
				Name:         "Test Provider",
				ProviderType: "aws",
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
			request: &ProviderCreateRequest{
				Name:         "Test Provider",
				ProviderType: "aws",
			},
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Failed to create provider",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/v1/providers", r.URL.Path)

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

			result, _, err := client.Providers.Create(ctx, tt.request)

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

// TestProvidersService_GetComprehensive tests the Get method
func TestProvidersService_GetComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		providerID string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Provider)
	}{
		{
			name:       "success - get active provider",
			providerID: "provider-aws-1",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":            "provider-aws-1",
					"name":          "AWS Production",
					"provider_type": "aws",
					"status":        "active",
					"vm_count":      150,
					"region":        "us-east-1",
					"description":   "Production AWS environment",
					"created_at":    "2024-01-10T10:00:00Z",
					"updated_at":    "2024-01-15T10:30:00Z",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, provider *Provider) {
				assert.Equal(t, "provider-aws-1", provider.ID)
				assert.Equal(t, "AWS Production", provider.Name)
				assert.Equal(t, "active", provider.Status)
				assert.Equal(t, 150, provider.VMCount)
			},
		},
		{
			name:       "success - get inactive provider",
			providerID: "provider-azure-1",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":            "provider-azure-1",
					"name":          "Azure Dev",
					"provider_type": "azure",
					"status":        "inactive",
					"vm_count":      0,
					"region":        "eastus",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, provider *Provider) {
				assert.Equal(t, "Azure Dev", provider.Name)
				assert.Equal(t, "inactive", provider.Status)
			},
		},
		{
			name:       "not found",
			providerID: "non-existent",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Provider not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			providerID: "provider-aws-1",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			providerID: "provider-aws-1",
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

			result, err := client.Providers.Get(ctx, tt.providerID)

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

// TestProvidersService_UpdateComprehensive tests the Update method
func TestProvidersService_UpdateComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		providerID string
		request    *ProviderUpdateRequest
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Provider)
	}{
		{
			name:       "success - update name",
			providerID: "provider-aws-1",
			request: &ProviderUpdateRequest{
				Name: "AWS Production Updated",
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":            "provider-aws-1",
					"name":          "AWS Production Updated",
					"provider_type": "aws",
					"status":        "active",
					"vm_count":      150,
					"region":        "us-east-1",
					"updated_at":    "2024-01-15T11:00:00Z",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, provider *Provider) {
				assert.Equal(t, "AWS Production Updated", provider.Name)
			},
		},
		{
			name:       "success - update description",
			providerID: "provider-aws-1",
			request: &ProviderUpdateRequest{
				Description: "Updated description for production AWS",
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":            "provider-aws-1",
					"name":          "AWS Production",
					"provider_type": "aws",
					"status":        "active",
					"vm_count":      150,
					"description":   "Updated description for production AWS",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, provider *Provider) {
				assert.Equal(t, "Updated description for production AWS", provider.Description)
			},
		},
		{
			name:       "success - update credentials",
			providerID: "provider-aws-1",
			request: &ProviderUpdateRequest{
				Credentials: map[string]interface{}{
					"access_key_id":     "NEW_KEY",
					"secret_access_key": "NEW_SECRET",
				},
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":            "provider-aws-1",
					"name":          "AWS Production",
					"provider_type": "aws",
					"status":        "active",
					"vm_count":      150,
				},
			},
			wantErr: false,
		},
		{
			name:       "not found",
			providerID: "non-existent",
			request: &ProviderUpdateRequest{
				Name: "Test",
			},
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Provider not found",
			},
			wantErr: true,
		},
		{
			name:       "conflict - name already exists",
			providerID: "provider-aws-1",
			request: &ProviderUpdateRequest{
				Name: "Azure Staging",
			},
			mockStatus: http.StatusConflict,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Provider with this name already exists",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			providerID: "provider-aws-1",
			request: &ProviderUpdateRequest{
				Name: "Test",
			},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			providerID: "provider-aws-1",
			request: &ProviderUpdateRequest{
				Name: "Test",
			},
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Failed to update provider",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "PUT", r.Method)

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

			result, err := client.Providers.Update(ctx, tt.providerID, tt.request)

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

// TestProvidersService_DeleteComprehensive tests the Delete method
func TestProvidersService_DeleteComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		providerID string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:       "success - delete provider",
			providerID: "provider-aws-1",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status":  "success",
				"message": "Provider deleted successfully",
			},
			wantErr: false,
		},
		{
			name:       "success - no content",
			providerID: "provider-azure-1",
			mockStatus: http.StatusNoContent,
			mockBody:   nil,
			wantErr:    false,
		},
		{
			name:       "not found",
			providerID: "non-existent",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Provider not found",
			},
			wantErr: true,
		},
		{
			name:       "conflict - provider has active VMs",
			providerID: "provider-aws-1",
			mockStatus: http.StatusConflict,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Cannot delete provider with active VMs",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			providerID: "provider-aws-1",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			providerID: "provider-aws-1",
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Failed to delete provider",
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

			err = client.Providers.Delete(ctx, tt.providerID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestProvidersService_SyncComprehensive tests the Sync method
func TestProvidersService_SyncComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		providerID string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *SyncResponse)
	}{
		{
			name:       "success - sync with new VMs",
			providerID: "provider-aws-1",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"success":     true,
					"message":     "Successfully synced provider",
					"provider_id": "provider-aws-1",
					"synced_at":   "2024-01-15T10:30:00Z",
					"synced_vms":  150,
					"vms_found":   155,
					"vms_added":   5,
					"vms_updated": 145,
					"vms_removed": 0,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, sync *SyncResponse) {
				assert.True(t, sync.Success)
				assert.Equal(t, 150, sync.SyncedVMs)
				assert.Equal(t, 5, sync.VMsAdded)
			},
		},
		{
			name:       "success - sync with removed VMs",
			providerID: "provider-azure-1",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"success":     true,
					"message":     "Sync completed",
					"provider_id": "provider-azure-1",
					"synced_at":   "2024-01-15T10:35:00Z",
					"synced_vms":  70,
					"vms_found":   70,
					"vms_added":   0,
					"vms_updated": 70,
					"vms_removed": 5,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, sync *SyncResponse) {
				assert.True(t, sync.Success)
				assert.Equal(t, 5, sync.VMsRemoved)
			},
		},
		{
			name:       "not found",
			providerID: "non-existent",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Provider not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			providerID: "provider-aws-1",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			providerID: "provider-aws-1",
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Failed to sync provider",
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

			result, err := client.Providers.Sync(ctx, tt.providerID)

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

// TestProvidersService_NetworkErrors tests handling of network-level errors
func TestProvidersService_NetworkErrors(t *testing.T) {
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
				_, _, apiErr = client.Providers.List(ctx, nil)
			case "get":
				_, apiErr = client.Providers.Get(ctx, "test-id")
			case "create":
				req := &ProviderCreateRequest{Name: "test", ProviderType: "test"}
				_, _, apiErr = client.Providers.Create(ctx, req)
			case "update":
				req := &ProviderUpdateRequest{Name: "updated"}
				_, apiErr = client.Providers.Update(ctx, "test-id", req)
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
