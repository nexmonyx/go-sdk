package nexmonyx

import (
	"context"
	"encoding/json"
	"fmt"
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

// TestOrganizationsService_GetComprehensive tests the Get method with various scenarios
func TestOrganizationsService_GetComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		orgID      string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Organization)
	}{
		{
			name:       "success - full organization data",
			orgID:      "org-123",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id": 123,
					"uuid":        "uuid-123",
					"name":        "Test Organization",
					"description": "Test org description",
					"created_at":  "2024-01-01T00:00:00Z",
					"updated_at":  "2024-01-01T00:00:00Z",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, org *Organization) {
				assert.Equal(t, uint(123), org.ID)
				assert.Equal(t, "uuid-123", org.UUID)
				assert.Equal(t, "Test Organization", org.Name)
				assert.Equal(t, "Test org description", org.Description)
			},
		},
		{
			name:       "success - minimal organization data",
			orgID:      "org-456",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id": 456,
					"uuid": "uuid-456",
					"name": "Minimal Org",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, org *Organization) {
				assert.Equal(t, uint(456), org.ID)
				assert.Equal(t, "Minimal Org", org.Name)
			},
		},
		{
			name:       "not found",
			orgID:      "org-999",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Organization not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			orgID:      "org-123",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "forbidden",
			orgID:      "org-123",
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Access denied",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			orgID:      "org-123",
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Internal server error",
			},
			wantErr: true,
		},
		{
			name:       "empty organization ID",
			orgID:      "",
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Organization ID required",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				if tt.orgID != "" {
					assert.Contains(t, r.URL.Path, tt.orgID)
				}
				assert.Contains(t, r.URL.Path, "/v1/organizations/")

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

			result, err := client.Organizations.Get(ctx, tt.orgID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestOrganizationsService_ListComprehensive tests the List method with various scenarios
func TestOrganizationsService_ListComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		opts       *ListOptions
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, []*Organization, *PaginationMeta)
	}{
		{
			name: "success - with pagination",
			opts: &ListOptions{
				Page:  1,
				Limit: 10,
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": []map[string]interface{}{
					{
						"id": 1,
						"uuid": "uuid-1",
						"name": "Organization 1",
					},
					{
						"id": 2,
						"uuid": "uuid-2",
						"name": "Organization 2",
					},
				},
				"meta": map[string]interface{}{
					"page": 1,
					"limit":     10,
					"total_items":        2,
					"total_pages":  1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, orgs []*Organization, meta *PaginationMeta) {
				assert.Len(t, orgs, 2)
				assert.Equal(t, uint(1), orgs[0].ID)
				assert.Equal(t, uint(2), orgs[1].ID)
				assert.NotNil(t, meta)
				assert.Equal(t, 1, meta.Page)
				assert.Equal(t, 2, meta.TotalItems)
			},
		},
		{
			name: "success - with search",
			opts: &ListOptions{
				Search: "Test",
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": []map[string]interface{}{
					{
						"id": 1,
						"name": "Test Organization",
					},
				},
				"meta": map[string]interface{}{
					"total_items": 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, orgs []*Organization, meta *PaginationMeta) {
				assert.Len(t, orgs, 1)
				assert.Contains(t, orgs[0].Name, "Test")
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
					"total_items": 0,
				},
			},
			wantErr: false,
		},
		{
			name: "success - empty result",
			opts: &ListOptions{Page: 1, Limit: 10},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data":   []map[string]interface{}{},
				"meta": map[string]interface{}{
					"total_items": 0,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, orgs []*Organization, meta *PaginationMeta) {
				assert.Len(t, orgs, 0)
			},
		},
		{
			name:       "unauthorized",
			opts:       &ListOptions{Page: 1},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "forbidden",
			opts:       &ListOptions{Page: 1},
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Access denied",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			opts:       &ListOptions{Page: 1},
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
				assert.Contains(t, r.URL.Path, "/v1/organizations")

				if tt.opts != nil {
					if tt.opts.Page > 0 {
						assert.Equal(t, fmt.Sprintf("%d", tt.opts.Page), r.URL.Query().Get("page"))
					}
					if tt.opts.Limit > 0 {
						assert.Equal(t, fmt.Sprintf("%d", tt.opts.Limit), r.URL.Query().Get("limit"))
					}
					if tt.opts.Search != "" {
						assert.Equal(t, tt.opts.Search, r.URL.Query().Get("search"))
					}
				}

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

			orgs, meta, err := client.Organizations.List(ctx, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, orgs)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, orgs)
				if tt.checkFunc != nil {
					tt.checkFunc(t, orgs, meta)
				}
			}
		})
	}
}

// TestOrganizationsService_CreateComprehensive tests the Create method with various scenarios
func TestOrganizationsService_CreateComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		org        *Organization
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Organization)
	}{
		{
			name: "success - full organization",
			org: &Organization{
				Name:        "New Organization",
				Description: "A new test organization",
			},
			mockStatus: http.StatusCreated,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id": 100,
					"uuid":        "uuid-new",
					"name":        "New Organization",
					"description": "A new test organization",
					"created_at":  "2024-01-01T00:00:00Z",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, org *Organization) {
				assert.Equal(t, uint(100), org.ID)
				assert.Equal(t, "New Organization", org.Name)
			},
		},
		{
			name: "success - minimal organization",
			org: &Organization{
				Name: "Minimal Org",
			},
			mockStatus: http.StatusCreated,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id": 101,
					"name": "Minimal Org",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, org *Organization) {
				assert.Equal(t, uint(101), org.ID)
			},
		},
		{
			name: "validation error - missing name",
			org: &Organization{
				Description: "No name",
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Name is required",
			},
			wantErr: true,
		},
		{
			name: "validation error - duplicate name",
			org: &Organization{
				Name: "Existing Org",
			},
			mockStatus: http.StatusConflict,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Organization with this name already exists",
			},
			wantErr: true,
		},
		{
			name: "unauthorized",
			org: &Organization{
				Name: "Test Org",
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
			org: &Organization{
				Name: "Test Org",
			},
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Insufficient permissions",
			},
			wantErr: true,
		},
		{
			name: "server error",
			org: &Organization{
				Name: "Test Org",
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
				assert.Contains(t, r.URL.Path, "/v1/organizations")

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

			result, err := client.Organizations.Create(ctx, tt.org)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestOrganizationsService_UpdateComprehensive tests the Update method with various scenarios
func TestOrganizationsService_UpdateComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		orgID      string
		org        *Organization
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Organization)
	}{
		{
			name:  "success - full update",
			orgID: "org-123",
			org: &Organization{
				Name:        "Updated Organization",
				Description: "Updated description",
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id": 123,
					"name":        "Updated Organization",
					"description": "Updated description",
					"updated_at":  "2024-01-02T00:00:00Z",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, org *Organization) {
				assert.Equal(t, uint(123), org.ID)
				assert.Equal(t, "Updated Organization", org.Name)
				assert.Equal(t, "Updated description", org.Description)
			},
		},
		{
			name:  "success - partial update",
			orgID: "org-456",
			org: &Organization{
				Description: "Only description updated",
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id": 456,
					"description": "Only description updated",
				},
			},
			wantErr: false,
		},
		{
			name:  "not found",
			orgID: "org-999",
			org: &Organization{
				Name: "Test",
			},
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Organization not found",
			},
			wantErr: true,
		},
		{
			name:  "validation error",
			orgID: "org-123",
			org: &Organization{
				Name: "", // Invalid empty name
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Invalid organization data",
			},
			wantErr: true,
		},
		{
			name:  "unauthorized",
			orgID: "org-123",
			org: &Organization{
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
			name:  "forbidden",
			orgID: "org-123",
			org: &Organization{
				Name: "Test",
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
			org: &Organization{
				Name: "Test",
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
				assert.Contains(t, r.URL.Path, tt.orgID)
				assert.Contains(t, r.URL.Path, "/v1/organizations/")

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

			result, err := client.Organizations.Update(ctx, tt.orgID, tt.org)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestOrganizationsService_DeleteComprehensive tests the Delete method with various scenarios
func TestOrganizationsService_DeleteComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		orgID      string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:       "success",
			orgID:      "org-123",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status":  "success",
				"message": "Organization deleted successfully",
			},
			wantErr: false,
		},
		{
			name:       "success - no content",
			orgID:      "org-456",
			mockStatus: http.StatusNoContent,
			mockBody:   nil,
			wantErr:    false,
		},
		{
			name:       "not found",
			orgID:      "org-999",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Organization not found",
			},
			wantErr: true,
		},
		{
			name:       "conflict - has dependencies",
			orgID:      "org-123",
			mockStatus: http.StatusConflict,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Cannot delete organization with active servers",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			orgID:      "org-123",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "forbidden",
			orgID:      "org-123",
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Insufficient permissions",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			orgID:      "org-123",
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
				assert.Contains(t, r.URL.Path, tt.orgID)
				assert.Contains(t, r.URL.Path, "/v1/organizations/")

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

			err = client.Organizations.Delete(ctx, tt.orgID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestOrganizationsService_GetByUUIDComprehensive tests the GetByUUID method with various scenarios
func TestOrganizationsService_GetByUUIDComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		uuid       string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Organization)
	}{
		{
			name:       "success - full data",
			uuid:       "uuid-123",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id": 123,
					"uuid":        "uuid-123",
					"name":        "Test Organization",
					"description": "Found by UUID",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, org *Organization) {
				assert.Equal(t, "uuid-123", org.UUID)
				assert.Equal(t, "Test Organization", org.Name)
			},
		},
		{
			name:       "not found",
			uuid:       "uuid-999",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Organization not found",
			},
			wantErr: true,
		},
		{
			name:       "invalid uuid format",
			uuid:       "invalid",
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Invalid UUID format",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			uuid:       "uuid-123",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "forbidden",
			uuid:       "uuid-123",
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Access denied",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			uuid:       "uuid-123",
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
				assert.Contains(t, r.URL.Path, tt.uuid)
				assert.Contains(t, r.URL.Path, "/v1/organizations/uuid/")

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

			result, err := client.Organizations.GetByUUID(ctx, tt.uuid)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestOrganizationsService_GetServersComprehensive tests the GetServers method with various scenarios
func TestOrganizationsService_GetServersComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		orgID      string
		opts       *ListOptions
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, []*Server, *PaginationMeta)
	}{
		{
			name:  "success - with servers",
			orgID: "org-123",
			opts: &ListOptions{
				Page:  1,
				Limit: 10,
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": []map[string]interface{}{
					{
						"id": 1,
						"uuid":     "srv-uuid-1",
						"hostname": "web-server-1",
					},
					{
						"id": 2,
						"uuid":     "srv-uuid-2",
						"hostname": "web-server-2",
					},
				},
				"meta": map[string]interface{}{
					"total_items": 2,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, servers []*Server, meta *PaginationMeta) {
				assert.Len(t, servers, 2)
				assert.Equal(t, uint(1), servers[0].ID)
				assert.NotNil(t, meta)
			},
		},
		{
			name:       "success - nil options",
			orgID:      "org-123",
			opts:       nil,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data":   []map[string]interface{}{},
				"meta": map[string]interface{}{
					"total_items": 0,
				},
			},
			wantErr: false,
		},
		{
			name:  "success - empty result",
			orgID: "org-456",
			opts:  &ListOptions{Page: 1},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data":   []map[string]interface{}{},
				"meta": map[string]interface{}{
					"total_items": 0,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, servers []*Server, meta *PaginationMeta) {
				assert.Len(t, servers, 0)
			},
		},
		{
			name:  "not found",
			orgID: "org-999",
			opts:  &ListOptions{Page: 1},
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
			opts:  &ListOptions{Page: 1},
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
			opts:  &ListOptions{Page: 1},
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
			opts:  &ListOptions{Page: 1},
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
				assert.Contains(t, r.URL.Path, tt.orgID)
				assert.Contains(t, r.URL.Path, "/v1/organizations/")
				assert.Contains(t, r.URL.Path, "/servers")

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

			servers, meta, err := client.Organizations.GetServers(ctx, tt.orgID, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, servers)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, servers)
				if tt.checkFunc != nil {
					tt.checkFunc(t, servers, meta)
				}
			}
		})
	}
}

// TestOrganizationsService_GetUsersComprehensive tests the GetUsers method with various scenarios
func TestOrganizationsService_GetUsersComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		orgID      string
		opts       *ListOptions
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, []*User, *PaginationMeta)
	}{
		{
			name:  "success - with users",
			orgID: "org-123",
			opts: &ListOptions{
				Page:  1,
				Limit: 10,
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": []map[string]interface{}{
					{
						"id": 1,
						"email": "user1@example.com",
						"name":  "User One",
					},
					{
						"id": 2,
						"email": "user2@example.com",
						"name":  "User Two",
					},
				},
				"meta": map[string]interface{}{
					"total_items": 2,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, users []*User, meta *PaginationMeta) {
				assert.Len(t, users, 2)
				assert.Equal(t, uint(1), users[0].ID)
				assert.NotNil(t, meta)
			},
		},
		{
			name:       "success - nil options",
			orgID:      "org-123",
			opts:       nil,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data":   []map[string]interface{}{},
				"meta": map[string]interface{}{
					"total_items": 0,
				},
			},
			wantErr: false,
		},
		{
			name:  "not found",
			orgID: "org-999",
			opts:  &ListOptions{Page: 1},
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
			opts:  &ListOptions{Page: 1},
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
			opts:  &ListOptions{Page: 1},
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
			opts:  &ListOptions{Page: 1},
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
				assert.Contains(t, r.URL.Path, tt.orgID)
				assert.Contains(t, r.URL.Path, "/v1/organizations/")
				assert.Contains(t, r.URL.Path, "/users")

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

			users, meta, err := client.Organizations.GetUsers(ctx, tt.orgID, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, users)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, users)
				if tt.checkFunc != nil {
					tt.checkFunc(t, users, meta)
				}
			}
		})
	}
}

// TestOrganizationsService_GetAlertsComprehensive tests the GetAlerts method with various scenarios
func TestOrganizationsService_GetAlertsComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		orgID      string
		opts       *ListOptions
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, []*Alert, *PaginationMeta)
	}{
		{
			name:  "success - with alerts",
			orgID: "org-123",
			opts: &ListOptions{
				Page:  1,
				Limit: 10,
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": []map[string]interface{}{
					{
						"id": 1,
						"name":        "CPU Alert",
						"description": "High CPU usage",
					},
					{
						"id": 2,
						"name":        "Memory Alert",
						"description": "High memory usage",
					},
				},
				"meta": map[string]interface{}{
					"total_items": 2,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alerts []*Alert, meta *PaginationMeta) {
				assert.Len(t, alerts, 2)
				assert.Equal(t, uint(1), alerts[0].ID)
				assert.NotNil(t, meta)
			},
		},
		{
			name:       "success - nil options",
			orgID:      "org-123",
			opts:       nil,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data":   []map[string]interface{}{},
				"meta": map[string]interface{}{
					"total_items": 0,
				},
			},
			wantErr: false,
		},
		{
			name:  "not found",
			orgID: "org-999",
			opts:  &ListOptions{Page: 1},
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
			opts:  &ListOptions{Page: 1},
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
			opts:  &ListOptions{Page: 1},
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
			opts:  &ListOptions{Page: 1},
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
				assert.Contains(t, r.URL.Path, tt.orgID)
				assert.Contains(t, r.URL.Path, "/v1/organizations/")
				assert.Contains(t, r.URL.Path, "/alerts")

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

			alerts, meta, err := client.Organizations.GetAlerts(ctx, tt.orgID, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, alerts)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, alerts)
				if tt.checkFunc != nil {
					tt.checkFunc(t, alerts, meta)
				}
			}
		})
	}
}

// TestOrganizationsService_UpdateSettingsComprehensive tests the UpdateSettings method with various scenarios
func TestOrganizationsService_UpdateSettingsComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		orgID      string
		settings   map[string]interface{}
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Organization)
	}{
		{
			name:  "success - full settings update",
			orgID: "org-123",
			settings: map[string]interface{}{
				"notifications_enabled": true,
				"alert_threshold":       80,
				"timezone":              "UTC",
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id": 123,
					"name": "Test Organization",
					"settings": map[string]interface{}{
						"notifications_enabled": true,
						"alert_threshold":       80,
						"timezone":              "UTC",
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, org *Organization) {
				assert.Equal(t, uint(123), org.ID)
			},
		},
		{
			name:  "success - single setting",
			orgID: "org-456",
			settings: map[string]interface{}{
				"timezone": "America/New_York",
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id": 456,
				},
			},
			wantErr: false,
		},
		{
			name:     "validation error - invalid settings",
			orgID:    "org-123",
			settings: map[string]interface{}{
				"invalid_key": "invalid_value",
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Invalid settings key",
			},
			wantErr: true,
		},
		{
			name:     "empty settings",
			orgID:    "org-123",
			settings: map[string]interface{}{},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Settings cannot be empty",
			},
			wantErr: true,
		},
		{
			name:  "not found",
			orgID: "org-999",
			settings: map[string]interface{}{
				"timezone": "UTC",
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
			settings: map[string]interface{}{
				"timezone": "UTC",
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
			settings: map[string]interface{}{
				"timezone": "UTC",
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
			settings: map[string]interface{}{
				"timezone": "UTC",
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
				assert.Contains(t, r.URL.Path, tt.orgID)
				assert.Contains(t, r.URL.Path, "/v1/organizations/")
				assert.Contains(t, r.URL.Path, "/settings")

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

			result, err := client.Organizations.UpdateSettings(ctx, tt.orgID, tt.settings)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestOrganizationsService_GetBillingComprehensive tests the GetBilling method with various scenarios
func TestOrganizationsService_GetBillingComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		orgID      string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, map[string]interface{})
	}{
		{
			name:       "success - full billing data",
			orgID:      "org-123",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"plan":         "premium",
					"status":       "active",
					"next_billing": "2024-02-01",
					"amount":       99.99,
					"currency":     "USD",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, billing map[string]interface{}) {
				assert.Equal(t, "premium", billing["plan"])
				assert.Equal(t, "active", billing["status"])
			},
		},
		{
			name:       "success - basic billing data",
			orgID:      "org-456",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"plan":   "free",
					"status": "active",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, billing map[string]interface{}) {
				assert.Equal(t, "free", billing["plan"])
			},
		},
		{
			name:       "not found",
			orgID:      "org-999",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Organization not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			orgID:      "org-123",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "forbidden",
			orgID:      "org-123",
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Access denied to billing information",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			orgID:      "org-123",
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
				assert.Contains(t, r.URL.Path, tt.orgID)
				assert.Contains(t, r.URL.Path, "/v1/organizations/")
				assert.Contains(t, r.URL.Path, "/billing")

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

			result, err := client.Organizations.GetBilling(ctx, tt.orgID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestOrganizationsService_NetworkErrors tests handling of network-level errors
func TestOrganizationsService_NetworkErrors(t *testing.T) {
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
				// Return URL on port that nothing is listening on
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
				// Use non-routable IP (RFC 5737 TEST-NET-1)
				return "http://192.0.2.1:8080"
			},
			setupContext: func() context.Context {
				// Very short timeout to fail fast
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
				// Use guaranteed non-existent domain
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
				// Create server that accepts connections but never responds
				listener, _ := net.Listen("tcp", "127.0.0.1:0")
				go func() {
					defer listener.Close()
					conn, err := listener.Accept()
					if err != nil {
						return
					}
					// Accept connection but never read/write - just hold it open
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
				RetryCount: 0, // Critical: prevent retry delays
				Timeout:    2 * time.Second,
			})
			require.NoError(t, err)

			var apiErr error
			switch tt.operation {
			case "list":
				_, _, apiErr = client.Organizations.List(ctx, nil)
			case "get":
				_, apiErr = client.Organizations.Get(ctx, "org-uuid")
			case "create":
				org := &Organization{Name: "Test Org"}
				_, apiErr = client.Organizations.Create(ctx, org)
			case "update":
				org := &Organization{Name: "Updated"}
				_, apiErr = client.Organizations.Update(ctx, "org-uuid", org)
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

// TestOrganizationsService_ConcurrentOperations tests concurrent operations on organizations
func TestOrganizationsService_ConcurrentOperations(t *testing.T) {
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
						"id":   1,
						"uuid": "org-1",
						"name": "Organization 1",
					},
				},
				"meta": map[string]interface{}{"total": 1},
			},
		},
		{
			name:              "concurrent Get - medium concurrency",
			concurrencyLevel:  50,
			operationsPerGoro: 2,
			operation:         "get",
			mockStatus:        http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":   1,
					"uuid": "org-1",
					"name": "Organization 1",
				},
			},
		},
		{
			name:              "concurrent Create - medium concurrency",
			concurrencyLevel:  30,
			operationsPerGoro: 2,
			operation:         "create",
			mockStatus:        http.StatusCreated,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":   2,
					"uuid": "org-new",
					"name": "New Organization",
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
				"meta":   map[string]interface{}{"total": 0},
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
				RetryCount: 0, // Critical: prevent retry delays in tests
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
							_, _, apiErr = client.Organizations.List(context.Background(), nil)
						case "get":
							_, apiErr = client.Organizations.Get(context.Background(), "org-1")
						case "create":
							org := &Organization{Name: "Test Organization"}
							_, apiErr = client.Organizations.Create(context.Background(), org)
						case "update":
							org := &Organization{Name: "Updated"}
							_, apiErr = client.Organizations.Update(context.Background(), "org-1", org)
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

			// Assertions
			totalOps := int64(tt.concurrencyLevel * tt.operationsPerGoro)
			assert.Equal(t, totalOps, successCount+errorCount, "Total operations should equal success + error count")
			assert.Equal(t, int64(0), errorCount, "Expected no errors in concurrent operations")
			assert.Equal(t, totalOps, successCount, "All operations should succeed")

			// Log performance metrics
			t.Logf("Completed %d operations in %v (%.2f ops/sec)",
				totalOps, duration, float64(totalOps)/duration.Seconds())
		})
	}
}
