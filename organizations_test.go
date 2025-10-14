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

// TestOrganizationsService_Get tests the Get method
func TestOrganizationsService_Get(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Organization)
	}{
		{
			name:       "successful get",
			id:         "org-123",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status: "success",
				Data: &Organization{
					GormModel:   GormModel{ID: 1},
					Name:        "Test Organization",
					Description: "A test organization",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, org *Organization) {
				assert.NotNil(t, org)
				assert.Equal(t, "Test Organization", org.Name)
				assert.Equal(t, "A test organization", org.Description)
			},
		},
		{
			name:       "organization not found",
			id:         "nonexistent",
			mockStatus: http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Organization not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			id:         "org-123",
			mockStatus: http.StatusUnauthorized,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Unauthorized",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/organizations/")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			result, err := client.Organizations.Get(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestOrganizationsService_GetByUUID tests the GetByUUID method
func TestOrganizationsService_GetByUUID(t *testing.T) {
	tests := []struct {
		name       string
		uuid       string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Organization)
	}{
		{
			name:       "successful get by UUID",
			uuid:       "550e8400-e29b-41d4-a716-446655440000",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status: "success",
				Data: &Organization{
					GormModel: GormModel{ID: 1},
					UUID:      "550e8400-e29b-41d4-a716-446655440000",
					Name:      "Test Organization",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, org *Organization) {
				assert.NotNil(t, org)
				assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", org.UUID)
			},
		},
		{
			name:       "organization not found",
			uuid:       "nonexistent-uuid",
			mockStatus: http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Organization not found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/organizations/uuid/")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			result, err := client.Organizations.GetByUUID(context.Background(), tt.uuid)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestOrganizationsService_List tests the List method
func TestOrganizationsService_List(t *testing.T) {
	tests := []struct {
		name       string
		opts       *ListOptions
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, []*Organization, *PaginationMeta)
	}{
		{
			name: "successful list with pagination",
			opts: &ListOptions{
				Page:  1,
				Limit: 10,
			},
			mockStatus: http.StatusOK,
			mockBody: PaginatedResponse{
				Status: "success",
				Data: []*Organization{
					{
						GormModel: GormModel{ID: 1},
						Name:      "Organization One",
					},
					{
						GormModel: GormModel{ID: 2},
						Name:      "Organization Two",
					},
				},
				Meta: &PaginationMeta{
					Page:       1,
					Limit:      10,
					TotalItems: 2,
					TotalPages: 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, orgs []*Organization, meta *PaginationMeta) {
				assert.NotNil(t, orgs)
				assert.Len(t, orgs, 2)
				assert.Equal(t, "Organization One", orgs[0].Name)
				assert.NotNil(t, meta)
				assert.Equal(t, 2, meta.TotalItems)
			},
		},
		{
			name:       "empty list",
			opts:       nil,
			mockStatus: http.StatusOK,
			mockBody: PaginatedResponse{
				Status: "success",
				Data:   []*Organization{},
				Meta: &PaginationMeta{
					Page:       1,
					Limit:      25,
					TotalItems: 0,
					TotalPages: 0,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, orgs []*Organization, meta *PaginationMeta) {
				assert.NotNil(t, orgs)
				assert.Len(t, orgs, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/v1/organizations", r.URL.Path)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			result, meta, err := client.Organizations.List(context.Background(), tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Nil(t, meta)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result, meta)
				}
			}
		})
	}
}

// TestOrganizationsService_Create tests the Create method
func TestOrganizationsService_Create(t *testing.T) {
	tests := []struct {
		name       string
		org        *Organization
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Organization)
	}{
		{
			name: "successful create",
			org: &Organization{
				Name:        "New Organization",
				Description: "A new organization",
			},
			mockStatus: http.StatusCreated,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Organization created successfully",
				Data: &Organization{
					GormModel:   GormModel{ID: 1},
					Name:        "New Organization",
					Description: "A new organization",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, org *Organization) {
				assert.NotNil(t, org)
				assert.Equal(t, uint(1), org.ID)
				assert.Equal(t, "New Organization", org.Name)
			},
		},
		{
			name: "validation error",
			org: &Organization{
				Name: "", // Empty name
			},
			mockStatus: http.StatusBadRequest,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Validation failed",
				Error:   "Name is required",
			},
			wantErr: true,
		},
		{
			name: "duplicate name",
			org: &Organization{
				Name: "Existing Org",
			},
			mockStatus: http.StatusConflict,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Organization name already exists",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/v1/organizations", r.URL.Path)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			result, err := client.Organizations.Create(context.Background(), tt.org)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestOrganizationsService_Update tests the Update method
func TestOrganizationsService_Update(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		org        *Organization
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Organization)
	}{
		{
			name: "successful update",
			id:   "org-123",
			org: &Organization{
				Name: "Updated Organization",
			},
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Organization updated successfully",
				Data: &Organization{
					GormModel: GormModel{ID: 1},
					Name:      "Updated Organization",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, org *Organization) {
				assert.NotNil(t, org)
				assert.Equal(t, "Updated Organization", org.Name)
			},
		},
		{
			name: "organization not found",
			id:   "nonexistent",
			org: &Organization{
				Name: "Test",
			},
			mockStatus: http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Organization not found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "PUT", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/organizations/")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			result, err := client.Organizations.Update(context.Background(), tt.id, tt.org)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestOrganizationsService_Delete tests the Delete method
func TestOrganizationsService_Delete(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:       "successful delete",
			id:         "org-123",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Organization deleted successfully",
			},
			wantErr: false,
		},
		{
			name:       "organization not found",
			id:         "nonexistent",
			mockStatus: http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Organization not found",
			},
			wantErr: true,
		},
		{
			name:       "forbidden - cannot delete",
			id:         "org-123",
			mockStatus: http.StatusForbidden,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Cannot delete organization with active servers",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "DELETE", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/organizations/")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			err = client.Organizations.Delete(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestOrganizationsService_GetServers tests the GetServers method
func TestOrganizationsService_GetServers(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		opts       *ListOptions
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, []*Server, *PaginationMeta)
	}{
		{
			name: "successful get servers",
			id:   "org-123",
			opts: &ListOptions{
				Page:  1,
				Limit: 10,
			},
			mockStatus: http.StatusOK,
			mockBody: PaginatedResponse{
				Status: "success",
				Data: []*Server{
					{
						GormModel:  GormModel{ID: 1},
						ServerUUID: "server-1",
						Hostname:   "web-server-01",
					},
					{
						GormModel:  GormModel{ID: 2},
						ServerUUID: "server-2",
						Hostname:   "db-server-01",
					},
				},
				Meta: &PaginationMeta{
					Page:       1,
					Limit:      10,
					TotalItems: 2,
					TotalPages: 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, servers []*Server, meta *PaginationMeta) {
				assert.NotNil(t, servers)
				assert.Len(t, servers, 2)
				assert.Equal(t, "web-server-01", servers[0].Hostname)
				assert.NotNil(t, meta)
			},
		},
		{
			name:       "organization not found",
			id:         "nonexistent",
			opts:       nil,
			mockStatus: http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Organization not found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/servers")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			result, meta, err := client.Organizations.GetServers(context.Background(), tt.id, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Nil(t, meta)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result, meta)
				}
			}
		})
	}
}

// TestOrganizationsService_GetUsers tests the GetUsers method
func TestOrganizationsService_GetUsers(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		opts       *ListOptions
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, []*User, *PaginationMeta)
	}{
		{
			name: "successful get users",
			id:   "org-123",
			opts: &ListOptions{
				Page:  1,
				Limit: 10,
			},
			mockStatus: http.StatusOK,
			mockBody: PaginatedResponse{
				Status: "success",
				Data: []*User{
					{
						GormModel: GormModel{ID: 1},
						Email:     "user1@example.com",
						FirstName: "User",
						LastName:  "One",
					},
					{
						GormModel: GormModel{ID: 2},
						Email:     "user2@example.com",
						FirstName: "User",
						LastName:  "Two",
					},
				},
				Meta: &PaginationMeta{
					Page:       1,
					Limit:      10,
					TotalItems: 2,
					TotalPages: 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, users []*User, meta *PaginationMeta) {
				assert.NotNil(t, users)
				assert.Len(t, users, 2)
				assert.Equal(t, "user1@example.com", users[0].Email)
				assert.NotNil(t, meta)
			},
		},
		{
			name:       "organization not found",
			id:         "nonexistent",
			opts:       nil,
			mockStatus: http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Organization not found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/users")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			result, meta, err := client.Organizations.GetUsers(context.Background(), tt.id, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Nil(t, meta)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result, meta)
				}
			}
		})
	}
}

// TestOrganizationsService_GetAlerts tests the GetAlerts method
func TestOrganizationsService_GetAlerts(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		opts       *ListOptions
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, []*Alert, *PaginationMeta)
	}{
		{
			name: "successful get alerts",
			id:   "org-123",
			opts: &ListOptions{
				Page:  1,
				Limit: 10,
			},
			mockStatus: http.StatusOK,
			mockBody: PaginatedResponse{
				Status: "success",
				Data: []*Alert{
					{
						GormModel: GormModel{ID: 1},
						Name:      "High CPU Alert",
						Type:      "metric",
					},
					{
						GormModel: GormModel{ID: 2},
						Name:      "Low Memory Alert",
						Type:      "metric",
					},
				},
				Meta: &PaginationMeta{
					Page:       1,
					Limit:      10,
					TotalItems: 2,
					TotalPages: 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alerts []*Alert, meta *PaginationMeta) {
				assert.NotNil(t, alerts)
				assert.Len(t, alerts, 2)
				assert.Equal(t, "High CPU Alert", alerts[0].Name)
				assert.NotNil(t, meta)
			},
		},
		{
			name:       "organization not found",
			id:         "nonexistent",
			opts:       nil,
			mockStatus: http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Organization not found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/alerts")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			result, meta, err := client.Organizations.GetAlerts(context.Background(), tt.id, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Nil(t, meta)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result, meta)
				}
			}
		})
	}
}

// TestOrganizationsService_UpdateSettings tests the UpdateSettings method
func TestOrganizationsService_UpdateSettings(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		settings   map[string]interface{}
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Organization)
	}{
		{
			name: "successful settings update",
			id:   "org-123",
			settings: map[string]interface{}{
				"timezone": "America/New_York",
				"theme":    "dark",
			},
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status: "success",
				Data: &Organization{
					GormModel: GormModel{ID: 1},
					Name:      "Test Organization",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, org *Organization) {
				assert.NotNil(t, org)
			},
		},
		{
			name: "organization not found",
			id:   "nonexistent",
			settings: map[string]interface{}{
				"theme": "dark",
			},
			mockStatus: http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Organization not found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "PUT", r.Method)
				assert.Contains(t, r.URL.Path, "/settings")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			result, err := client.Organizations.UpdateSettings(context.Background(), tt.id, tt.settings)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestOrganizationsService_GetBilling tests the GetBilling method
func TestOrganizationsService_GetBilling(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, map[string]interface{})
	}{
		{
			name:       "successful get billing",
			id:         "org-123",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status: "success",
				Data: map[string]interface{}{
					"subscription_status": "active",
					"plan":                "enterprise",
					"billing_email":       "billing@example.com",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, billing map[string]interface{}) {
				assert.NotNil(t, billing)
				assert.Equal(t, "active", billing["subscription_status"])
			},
		},
		{
			name:       "organization not found",
			id:         "nonexistent",
			mockStatus: http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Organization not found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/billing")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			result, err := client.Organizations.GetBilling(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestOrganizationJSON tests JSON marshaling and unmarshaling of Organization
func TestOrganizationJSON(t *testing.T) {
	org := &Organization{
		GormModel:   GormModel{ID: 1},
		UUID:        "550e8400-e29b-41d4-a716-446655440000",
		Name:        "Test Organization",
		Description: "A test organization",
	}

	// Marshal to JSON
	data, err := json.Marshal(org)
	require.NoError(t, err)
	assert.Contains(t, string(data), "Test Organization")
	assert.Contains(t, string(data), "A test organization")

	// Unmarshal from JSON
	var decoded Organization
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, org.Name, decoded.Name)
	assert.Equal(t, org.Description, decoded.Description)
	assert.Equal(t, org.UUID, decoded.UUID)
}
