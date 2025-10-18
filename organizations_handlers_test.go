package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Phase 2A: Organizations CRUD Handler Tests

func TestOrganizationsService_Create_Handler(t *testing.T) {
	tests := []struct {
		name           string
		input          *Organization
		mockStatus     int
		mockResponse   interface{}
		wantErr        bool
		checkFunc      func(*testing.T, *Organization)
		clientAuthFunc func(*Config)
	}{
		{
			name: "success - create organization",
			input: &Organization{
				Name:        "Tech Corp",
				Description: "A technology corporation",
			},
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"id":          1,
					"uuid":        "org-uuid-001",
					"name":        "Tech Corp",
					"description": "A technology corporation",
					"created_at":  "2021-10-01T10:00:00Z",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, org *Organization) {
				assert.Equal(t, "Tech Corp", org.Name)
				assert.Equal(t, "A technology corporation", org.Description)
				assert.Equal(t, "org-uuid-001", org.UUID)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name: "success - create organization with minimal fields",
			input: &Organization{
				Name: "New Org",
			},
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"id":   2,
					"uuid": "org-uuid-002",
					"name": "New Org",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, org *Organization) {
				assert.Equal(t, "New Org", org.Name)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - bad request (400)",
			input:        &Organization{},
			mockStatus:   http.StatusBadRequest,
			mockResponse: map[string]interface{}{"error": "name is required"},
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name: "error - unauthorized (401)",
			input: &Organization{
				Name: "Unauthorized Org",
			},
			mockStatus:   http.StatusUnauthorized,
			mockResponse: map[string]interface{}{"error": "invalid token"},
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "invalid-token"}
			},
		},
		{
			name: "error - conflict (409) - duplicate name",
			input: &Organization{
				Name: "Existing Org",
			},
			mockStatus:   http.StatusConflict,
			mockResponse: map[string]interface{}{"error": "organization name already exists"},
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name: "error - server error (500)",
			input: &Organization{
				Name: "Tech Corp",
			},
			mockStatus:   http.StatusInternalServerError,
			mockResponse: map[string]interface{}{"error": "internal server error"},
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/organizations")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			config := &Config{BaseURL: server.URL}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			assert.NoError(t, err)

			org, err := client.Organizations.Create(context.Background(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, org)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, org)
				if tt.checkFunc != nil {
					tt.checkFunc(t, org)
				}
			}
		})
	}
}

func TestOrganizationsService_Get_Handler(t *testing.T) {
	tests := []struct {
		name           string
		orgID          string
		mockStatus     int
		mockResponse   interface{}
		wantErr        bool
		checkFunc      func(*testing.T, *Organization)
		clientAuthFunc func(*Config)
	}{
		{
			name:       "success - get organization",
			orgID:      "1",
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"id":          1,
					"uuid":        "org-uuid-001",
					"name":        "Tech Corp",
					"description": "A technology corporation",
					"created_at":  "2021-10-01T10:00:00Z",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, org *Organization) {
				assert.Equal(t, "Tech Corp", org.Name)
				assert.Equal(t, uint(1), org.ID)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - not found (404)",
			orgID:        "999",
			mockStatus:   http.StatusNotFound,
			mockResponse: map[string]interface{}{"error": "organization not found"},
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - unauthorized (401)",
			orgID:        "1",
			mockStatus:   http.StatusUnauthorized,
			mockResponse: map[string]interface{}{"error": "invalid token"},
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "invalid-token"}
			},
		},
		{
			name:         "error - forbidden (403)",
			orgID:        "2",
			mockStatus:   http.StatusForbidden,
			mockResponse: map[string]interface{}{"error": "insufficient permissions"},
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - server error (500)",
			orgID:        "1",
			mockStatus:   http.StatusInternalServerError,
			mockResponse: map[string]interface{}{"error": "internal server error"},
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/organizations/"+tt.orgID)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			config := &Config{BaseURL: server.URL}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			assert.NoError(t, err)

			org, err := client.Organizations.Get(context.Background(), tt.orgID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, org)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, org)
				if tt.checkFunc != nil {
					tt.checkFunc(t, org)
				}
			}
		})
	}
}

func TestOrganizationsService_GetByUUID_Handler(t *testing.T) {
	tests := []struct {
		name           string
		uuid           string
		mockStatus     int
		mockResponse   interface{}
		wantErr        bool
		checkFunc      func(*testing.T, *Organization)
		clientAuthFunc func(*Config)
	}{
		{
			name:       "success - get organization by UUID",
			uuid:       "org-uuid-001",
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"id":          1,
					"uuid":        "org-uuid-001",
					"name":        "Tech Corp",
					"description": "A technology corporation",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, org *Organization) {
				assert.Equal(t, "org-uuid-001", org.UUID)
				assert.Equal(t, "Tech Corp", org.Name)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - not found (404)",
			uuid:         "invalid-uuid",
			mockStatus:   http.StatusNotFound,
			mockResponse: map[string]interface{}{"error": "organization not found"},
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - unauthorized (401)",
			uuid:         "org-uuid-001",
			mockStatus:   http.StatusUnauthorized,
			mockResponse: map[string]interface{}{"error": "invalid token"},
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "invalid-token"}
			},
		},
		{
			name:         "error - forbidden (403)",
			uuid:         "org-uuid-002",
			mockStatus:   http.StatusForbidden,
			mockResponse: map[string]interface{}{"error": "insufficient permissions"},
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - server error (500)",
			uuid:         "org-uuid-001",
			mockStatus:   http.StatusInternalServerError,
			mockResponse: map[string]interface{}{"error": "internal server error"},
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/organizations/uuid/"+tt.uuid)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			config := &Config{BaseURL: server.URL}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			assert.NoError(t, err)

			org, err := client.Organizations.GetByUUID(context.Background(), tt.uuid)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, org)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, org)
				if tt.checkFunc != nil {
					tt.checkFunc(t, org)
				}
			}
		})
	}
}

func TestOrganizationsService_List_Handler(t *testing.T) {
	tests := []struct {
		name           string
		opts           *ListOptions
		mockStatus     int
		mockResponse   interface{}
		wantErr        bool
		wantLen        int
		checkFunc      func(*testing.T, []*Organization, *PaginationMeta)
		clientAuthFunc func(*Config)
	}{
		{
			name:       "success - list organizations",
			opts:       nil,
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id":   1,
						"uuid": "org-uuid-001",
						"name": "Tech Corp",
					},
					{
						"id":   2,
						"uuid": "org-uuid-002",
						"name": "Finance Inc",
					},
				},
				"meta": map[string]interface{}{
					"total_items": 2,
					"limit":       25,
					"page":        1,
				},
			},
			wantErr: false,
			wantLen: 2,
			checkFunc: func(t *testing.T, orgs []*Organization, meta *PaginationMeta) {
				assert.Equal(t, 2, meta.TotalItems)
				if len(orgs) > 0 && orgs[0] != nil {
					assert.Equal(t, "Tech Corp", orgs[0].Name)
				}
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:       "success - list with pagination",
			opts:       &ListOptions{Page: 2, Limit: 10},
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id":   11,
						"uuid": "org-uuid-011",
						"name": "Research Lab",
					},
				},
				"meta": map[string]interface{}{
					"total_items": 25,
					"limit":       10,
					"page":        2,
					"per_page":    10,
				},
			},
			wantErr: false,
			wantLen: 1,
			checkFunc: func(t *testing.T, orgs []*Organization, meta *PaginationMeta) {
				assert.Equal(t, 25, meta.TotalItems)
				assert.Equal(t, 2, meta.Page)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:       "success - empty list",
			opts:       nil,
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": []map[string]interface{}{},
				"meta": map[string]interface{}{
					"total_items": 0,
					"limit":       25,
					"page":        1,
				},
			},
			wantErr: false,
			wantLen: 0,
			checkFunc: func(t *testing.T, orgs []*Organization, meta *PaginationMeta) {
				assert.Equal(t, 0, meta.TotalItems)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - unauthorized (401)",
			opts:         nil,
			mockStatus:   http.StatusUnauthorized,
			mockResponse: map[string]interface{}{"error": "invalid token"},
			wantErr:      true,
			wantLen:      0,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "invalid-token"}
			},
		},
		{
			name:         "error - forbidden (403)",
			opts:         nil,
			mockStatus:   http.StatusForbidden,
			mockResponse: map[string]interface{}{"error": "insufficient permissions"},
			wantErr:      true,
			wantLen:      0,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - bad request (400)",
			opts:         &ListOptions{Page: -1},
			mockStatus:   http.StatusBadRequest,
			mockResponse: map[string]interface{}{"error": "invalid pagination parameters"},
			wantErr:      true,
			wantLen:      0,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - server error (500)",
			opts:         nil,
			mockStatus:   http.StatusInternalServerError,
			mockResponse: map[string]interface{}{"error": "internal server error"},
			wantErr:      true,
			wantLen:      0,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/organizations")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			config := &Config{BaseURL: server.URL}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			assert.NoError(t, err)

			orgs, meta, err := client.Organizations.List(context.Background(), tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, orgs)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantLen, len(orgs))
				if tt.checkFunc != nil {
					tt.checkFunc(t, orgs, meta)
				}
			}
		})
	}
}

func TestOrganizationsService_Update_Handler(t *testing.T) {
	tests := []struct {
		name           string
		orgID          string
		input          *Organization
		mockStatus     int
		mockResponse   interface{}
		wantErr        bool
		checkFunc      func(*testing.T, *Organization)
		clientAuthFunc func(*Config)
	}{
		{
			name:  "success - update organization",
			orgID: "1",
			input: &Organization{
				Name:        "Tech Corp Updated",
				Description: "Updated description",
			},
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"id":          1,
					"uuid":        "org-uuid-001",
					"name":        "Tech Corp Updated",
					"description": "Updated description",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, org *Organization) {
				assert.Equal(t, "Tech Corp Updated", org.Name)
				assert.Equal(t, "Updated description", org.Description)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:  "success - partial update",
			orgID: "2",
			input: &Organization{
				Name: "Finance Inc Updated",
			},
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"id":   2,
					"uuid": "org-uuid-002",
					"name": "Finance Inc Updated",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, org *Organization) {
				assert.Equal(t, "Finance Inc Updated", org.Name)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - bad request (400)",
			orgID:        "1",
			input:        &Organization{Name: ""},
			mockStatus:   http.StatusBadRequest,
			mockResponse: map[string]interface{}{"error": "name is required"},
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:  "error - not found (404)",
			orgID: "999",
			input: &Organization{
				Name: "Updated",
			},
			mockStatus:   http.StatusNotFound,
			mockResponse: map[string]interface{}{"error": "organization not found"},
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:  "error - unauthorized (401)",
			orgID: "1",
			input: &Organization{
				Name: "Updated",
			},
			mockStatus:   http.StatusUnauthorized,
			mockResponse: map[string]interface{}{"error": "invalid token"},
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "invalid-token"}
			},
		},
		{
			name:  "error - forbidden (403)",
			orgID: "2",
			input: &Organization{
				Name: "Updated",
			},
			mockStatus:   http.StatusForbidden,
			mockResponse: map[string]interface{}{"error": "insufficient permissions"},
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:  "error - server error (500)",
			orgID: "1",
			input: &Organization{
				Name: "Updated",
			},
			mockStatus:   http.StatusInternalServerError,
			mockResponse: map[string]interface{}{"error": "internal server error"},
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Contains(t, []string{"PUT", "PATCH", "POST"}, r.Method)
				assert.Contains(t, r.URL.Path, "/v1/organizations/"+tt.orgID)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			config := &Config{BaseURL: server.URL}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			assert.NoError(t, err)

			org, err := client.Organizations.Update(context.Background(), tt.orgID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, org)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, org)
				if tt.checkFunc != nil {
					tt.checkFunc(t, org)
				}
			}
		})
	}
}

func TestOrganizationsService_Delete_Handler(t *testing.T) {
	tests := []struct {
		name           string
		orgID          string
		mockStatus     int
		mockResponse   interface{}
		wantErr        bool
		clientAuthFunc func(*Config)
	}{
		{
			name:         "success - delete organization (200 OK)",
			orgID:        "1",
			mockStatus:   http.StatusOK,
			mockResponse: map[string]interface{}{},
			wantErr:      false,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "success - delete organization (204 No Content)",
			orgID:        "2",
			mockStatus:   http.StatusNoContent,
			mockResponse: nil,
			wantErr:      false,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - not found (404)",
			orgID:        "999",
			mockStatus:   http.StatusNotFound,
			mockResponse: map[string]interface{}{"error": "organization not found"},
			wantErr:      true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - unauthorized (401)",
			orgID:        "1",
			mockStatus:   http.StatusUnauthorized,
			mockResponse: map[string]interface{}{"error": "invalid token"},
			wantErr:      true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "invalid-token"}
			},
		},
		{
			name:         "error - forbidden (403)",
			orgID:        "2",
			mockStatus:   http.StatusForbidden,
			mockResponse: map[string]interface{}{"error": "insufficient permissions"},
			wantErr:      true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - conflict (409) - has dependencies",
			orgID:        "1",
			mockStatus:   http.StatusConflict,
			mockResponse: map[string]interface{}{"error": "organization has active servers"},
			wantErr:      true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - server error (500)",
			orgID:        "1",
			mockStatus:   http.StatusInternalServerError,
			mockResponse: map[string]interface{}{"error": "internal server error"},
			wantErr:      true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "DELETE", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/organizations/"+tt.orgID)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}))
			defer server.Close()

			config := &Config{BaseURL: server.URL}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			assert.NoError(t, err)

			err = client.Organizations.Delete(context.Background(), tt.orgID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Phase 2B: Organizations Resource Retrieval Handler Tests

func TestOrganizationsService_GetServers_Handler(t *testing.T) {
	tests := []struct {
		name           string
		orgID          string
		opts           *ListOptions
		mockStatus     int
		mockResponse   interface{}
		wantErr        bool
		wantLen        int
		checkFunc      func(*testing.T, []*Server, *PaginationMeta)
		clientAuthFunc func(*Config)
	}{
		{
			name:       "success - get organization servers",
			orgID:      "1",
			opts:       nil,
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id":              1,
						"server_uuid":     "server-uuid-001",
						"hostname":        "web-server-01",
						"organization_id": 1,
						"main_ip":         "192.168.1.100",
					},
					{
						"id":              2,
						"server_uuid":     "server-uuid-002",
						"hostname":        "db-server-01",
						"organization_id": 1,
						"main_ip":         "192.168.1.101",
					},
				},
				"meta": map[string]interface{}{
					"total_items": 2,
					"limit":       25,
					"page":        1,
				},
			},
			wantErr: false,
			wantLen: 2,
			checkFunc: func(t *testing.T, servers []*Server, meta *PaginationMeta) {
				assert.Equal(t, 2, meta.TotalItems)
				if len(servers) > 0 && servers[0] != nil {
					assert.Equal(t, "web-server-01", servers[0].Hostname)
				}
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:       "success - get servers with pagination",
			orgID:      "1",
			opts:       &ListOptions{Page: 1, Limit: 10},
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id":              3,
						"server_uuid":     "server-uuid-003",
						"hostname":        "app-server-01",
						"organization_id": 1,
						"main_ip":         "192.168.1.102",
					},
				},
				"meta": map[string]interface{}{
					"total_items": 15,
					"limit":       10,
					"page":        1,
					"per_page":    10,
				},
			},
			wantErr: false,
			wantLen: 1,
			checkFunc: func(t *testing.T, servers []*Server, meta *PaginationMeta) {
				assert.Equal(t, 15, meta.TotalItems)
				assert.Equal(t, 10, meta.Limit)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:       "success - empty servers list",
			orgID:      "2",
			opts:       nil,
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": []map[string]interface{}{},
				"meta": map[string]interface{}{
					"total_items": 0,
					"limit":       25,
					"page":        1,
				},
			},
			wantErr: false,
			wantLen: 0,
			checkFunc: func(t *testing.T, servers []*Server, meta *PaginationMeta) {
				assert.Equal(t, 0, meta.TotalItems)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - not found (404)",
			orgID:        "999",
			opts:         nil,
			mockStatus:   http.StatusNotFound,
			mockResponse: map[string]interface{}{"error": "organization not found"},
			wantErr:      true,
			wantLen:      0,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - unauthorized (401)",
			orgID:        "1",
			opts:         nil,
			mockStatus:   http.StatusUnauthorized,
			mockResponse: map[string]interface{}{"error": "invalid token"},
			wantErr:      true,
			wantLen:      0,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "invalid-token"}
			},
		},
		{
			name:         "error - forbidden (403)",
			orgID:        "2",
			opts:         nil,
			mockStatus:   http.StatusForbidden,
			mockResponse: map[string]interface{}{"error": "insufficient permissions"},
			wantErr:      true,
			wantLen:      0,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - server error (500)",
			orgID:        "1",
			opts:         nil,
			mockStatus:   http.StatusInternalServerError,
			mockResponse: map[string]interface{}{"error": "internal server error"},
			wantErr:      true,
			wantLen:      0,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/organizations/"+tt.orgID+"/servers")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			config := &Config{BaseURL: server.URL}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			assert.NoError(t, err)

			servers, meta, err := client.Organizations.GetServers(context.Background(), tt.orgID, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, servers)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantLen, len(servers))
				if tt.checkFunc != nil {
					tt.checkFunc(t, servers, meta)
				}
			}
		})
	}
}

func TestOrganizationsService_GetUsers_Handler(t *testing.T) {
	tests := []struct {
		name           string
		orgID          string
		opts           *ListOptions
		mockStatus     int
		mockResponse   interface{}
		wantErr        bool
		wantLen        int
		checkFunc      func(*testing.T, []*User, *PaginationMeta)
		clientAuthFunc func(*Config)
	}{
		{
			name:       "success - get organization users",
			orgID:      "1",
			opts:       nil,
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id":    1,
						"uuid":  "user-uuid-001",
						"email": "admin@example.com",
						"name":  "Admin User",
						"role":  "admin",
					},
					{
						"id":    2,
						"uuid":  "user-uuid-002",
						"email": "user@example.com",
						"name":  "Regular User",
						"role":  "member",
					},
				},
				"meta": map[string]interface{}{
					"total_items": 2,
					"limit":       25,
					"page":        1,
				},
			},
			wantErr: false,
			wantLen: 2,
			checkFunc: func(t *testing.T, users []*User, meta *PaginationMeta) {
				assert.Equal(t, 2, meta.TotalItems)
				if len(users) > 0 && users[0] != nil {
					assert.Equal(t, "admin@example.com", users[0].Email)
				}
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:       "success - get users with pagination",
			orgID:      "1",
			opts:       &ListOptions{Page: 2, Limit: 10},
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id":    11,
						"uuid":  "user-uuid-011",
						"email": "viewer@example.com",
						"name":  "Viewer User",
						"role":  "viewer",
					},
				},
				"meta": map[string]interface{}{
					"total_items": 25,
					"limit":       10,
					"page":        2,
					"per_page":    10,
				},
			},
			wantErr: false,
			wantLen: 1,
			checkFunc: func(t *testing.T, users []*User, meta *PaginationMeta) {
				assert.Equal(t, 25, meta.TotalItems)
				assert.Equal(t, 2, meta.Page)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:       "success - empty users list",
			orgID:      "3",
			opts:       nil,
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": []map[string]interface{}{},
				"meta": map[string]interface{}{
					"total_items": 0,
					"limit":       25,
					"page":        1,
				},
			},
			wantErr: false,
			wantLen: 0,
			checkFunc: func(t *testing.T, users []*User, meta *PaginationMeta) {
				assert.Equal(t, 0, meta.TotalItems)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - not found (404)",
			orgID:        "999",
			opts:         nil,
			mockStatus:   http.StatusNotFound,
			mockResponse: map[string]interface{}{"error": "organization not found"},
			wantErr:      true,
			wantLen:      0,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - unauthorized (401)",
			orgID:        "1",
			opts:         nil,
			mockStatus:   http.StatusUnauthorized,
			mockResponse: map[string]interface{}{"error": "invalid token"},
			wantErr:      true,
			wantLen:      0,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "invalid-token"}
			},
		},
		{
			name:         "error - forbidden (403)",
			orgID:        "2",
			opts:         nil,
			mockStatus:   http.StatusForbidden,
			mockResponse: map[string]interface{}{"error": "insufficient permissions"},
			wantErr:      true,
			wantLen:      0,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - server error (500)",
			orgID:        "1",
			opts:         nil,
			mockStatus:   http.StatusInternalServerError,
			mockResponse: map[string]interface{}{"error": "internal server error"},
			wantErr:      true,
			wantLen:      0,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/organizations/"+tt.orgID+"/users")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			config := &Config{BaseURL: server.URL}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			assert.NoError(t, err)

			users, meta, err := client.Organizations.GetUsers(context.Background(), tt.orgID, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, users)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantLen, len(users))
				if tt.checkFunc != nil {
					tt.checkFunc(t, users, meta)
				}
			}
		})
	}
}

func TestOrganizationsService_GetAlerts_Handler(t *testing.T) {
	tests := []struct {
		name           string
		orgID          string
		opts           *ListOptions
		mockStatus     int
		mockResponse   interface{}
		wantErr        bool
		wantLen        int
		checkFunc      func(*testing.T, []*Alert, *PaginationMeta)
		clientAuthFunc func(*Config)
	}{
		{
			name:       "success - get organization alerts",
			orgID:      "1",
			opts:       nil,
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id":               1,
						"name":             "High CPU Usage",
						"type":             "metric_threshold",
						"metric_name":      "cpu_usage",
						"condition":        ">",
						"threshold":        80.0,
						"organization_id":  1,
					},
					{
						"id":               2,
						"name":             "High Memory Usage",
						"type":             "metric_threshold",
						"metric_name":      "memory_usage",
						"condition":        ">",
						"threshold":        85.0,
						"organization_id":  1,
					},
				},
				"meta": map[string]interface{}{
					"total_items": 2,
					"limit":       25,
					"page":        1,
				},
			},
			wantErr: false,
			wantLen: 2,
			checkFunc: func(t *testing.T, alerts []*Alert, meta *PaginationMeta) {
				assert.Equal(t, 2, meta.TotalItems)
				if len(alerts) > 0 && alerts[0] != nil {
					assert.Equal(t, "High CPU Usage", alerts[0].Name)
				}
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:       "success - get alerts with pagination",
			orgID:      "1",
			opts:       &ListOptions{Page: 1, Limit: 15},
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id":               5,
						"name":             "Disk Usage Alert",
						"type":             "metric_threshold",
						"metric_name":      "disk_usage",
						"condition":        ">",
						"threshold":        90.0,
						"organization_id":  1,
					},
				},
				"meta": map[string]interface{}{
					"total_items": 30,
					"limit":       15,
					"page":        1,
					"per_page":    15,
				},
			},
			wantErr: false,
			wantLen: 1,
			checkFunc: func(t *testing.T, alerts []*Alert, meta *PaginationMeta) {
				assert.Equal(t, 30, meta.TotalItems)
				assert.Equal(t, 1, meta.Page)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:       "success - empty alerts list",
			orgID:      "4",
			opts:       nil,
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": []map[string]interface{}{},
				"meta": map[string]interface{}{
					"total_items": 0,
					"limit":       25,
					"page":        1,
				},
			},
			wantErr: false,
			wantLen: 0,
			checkFunc: func(t *testing.T, alerts []*Alert, meta *PaginationMeta) {
				assert.Equal(t, 0, meta.TotalItems)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - not found (404)",
			orgID:        "999",
			opts:         nil,
			mockStatus:   http.StatusNotFound,
			mockResponse: map[string]interface{}{"error": "organization not found"},
			wantErr:      true,
			wantLen:      0,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - unauthorized (401)",
			orgID:        "1",
			opts:         nil,
			mockStatus:   http.StatusUnauthorized,
			mockResponse: map[string]interface{}{"error": "invalid token"},
			wantErr:      true,
			wantLen:      0,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "invalid-token"}
			},
		},
		{
			name:         "error - forbidden (403)",
			orgID:        "2",
			opts:         nil,
			mockStatus:   http.StatusForbidden,
			mockResponse: map[string]interface{}{"error": "insufficient permissions"},
			wantErr:      true,
			wantLen:      0,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - server error (500)",
			orgID:        "1",
			opts:         nil,
			mockStatus:   http.StatusInternalServerError,
			mockResponse: map[string]interface{}{"error": "internal server error"},
			wantErr:      true,
			wantLen:      0,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/organizations/"+tt.orgID+"/alerts")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			config := &Config{BaseURL: server.URL}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			assert.NoError(t, err)

			alerts, meta, err := client.Organizations.GetAlerts(context.Background(), tt.orgID, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, alerts)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantLen, len(alerts))
				if tt.checkFunc != nil {
					tt.checkFunc(t, alerts, meta)
				}
			}
		})
	}
}

// Phase 2C: Organizations Special Operations Handler Tests

func TestOrganizationsService_UpdateSettings_Handler(t *testing.T) {
	tests := []struct {
		name           string
		orgID          string
		settings       map[string]interface{}
		mockStatus     int
		mockResponse   interface{}
		wantErr        bool
		checkFunc      func(*testing.T, *Organization)
		clientAuthFunc func(*Config)
	}{
		{
			name:  "success - update organization settings",
			orgID: "1",
			settings: map[string]interface{}{
				"notification_enabled": true,
				"timezone":             "UTC",
			},
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"id":       1,
					"uuid":     "org-uuid-001",
					"name":     "Tech Corp",
					"settings": map[string]interface{}{"notification_enabled": true},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, org *Organization) {
				assert.Equal(t, "Tech Corp", org.Name)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:  "success - update single setting",
			orgID: "2",
			settings: map[string]interface{}{
				"max_users": 100,
			},
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"id":       2,
					"uuid":     "org-uuid-002",
					"name":     "Finance Inc",
					"settings": map[string]interface{}{"max_users": 100},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, org *Organization) {
				assert.Equal(t, "Finance Inc", org.Name)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - bad request (400)",
			orgID:        "1",
			settings:     map[string]interface{}{"invalid_setting": "value"},
			mockStatus:   http.StatusBadRequest,
			mockResponse: map[string]interface{}{"error": "invalid settings"},
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - not found (404)",
			orgID:        "999",
			settings:     map[string]interface{}{"timezone": "UTC"},
			mockStatus:   http.StatusNotFound,
			mockResponse: map[string]interface{}{"error": "organization not found"},
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - unauthorized (401)",
			orgID:        "1",
			settings:     map[string]interface{}{"timezone": "UTC"},
			mockStatus:   http.StatusUnauthorized,
			mockResponse: map[string]interface{}{"error": "invalid token"},
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "invalid-token"}
			},
		},
		{
			name:         "error - forbidden (403)",
			orgID:        "2",
			settings:     map[string]interface{}{"timezone": "UTC"},
			mockStatus:   http.StatusForbidden,
			mockResponse: map[string]interface{}{"error": "insufficient permissions"},
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - server error (500)",
			orgID:        "1",
			settings:     map[string]interface{}{"timezone": "UTC"},
			mockStatus:   http.StatusInternalServerError,
			mockResponse: map[string]interface{}{"error": "internal server error"},
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Contains(t, []string{"PUT", "PATCH", "POST"}, r.Method)
				assert.Contains(t, r.URL.Path, "/v1/organizations/"+tt.orgID+"/settings")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			config := &Config{BaseURL: server.URL}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			assert.NoError(t, err)

			org, err := client.Organizations.UpdateSettings(context.Background(), tt.orgID, tt.settings)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, org)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, org)
				if tt.checkFunc != nil {
					tt.checkFunc(t, org)
				}
			}
		})
	}
}

func TestOrganizationsService_GetBilling_Handler(t *testing.T) {
	tests := []struct {
		name           string
		orgID          string
		mockStatus     int
		mockResponse   interface{}
		wantErr        bool
		checkFunc      func(*testing.T, map[string]interface{})
		clientAuthFunc func(*Config)
	}{
		{
			name:       "success - get billing information",
			orgID:      "1",
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"subscription_id":   "sub-001",
					"plan":              "professional",
					"status":            "active",
					"next_billing_date": "2025-11-17",
					"amount":            99.99,
					"currency":          "USD",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, billing map[string]interface{}) {
				assert.NotNil(t, billing)
				assert.Equal(t, "sub-001", billing["subscription_id"])
				assert.Equal(t, "professional", billing["plan"])
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:       "success - get billing with payment method",
			orgID:      "2",
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"subscription_id": "sub-002",
					"plan":            "enterprise",
					"status":          "active",
					"payment_method":  "credit_card",
					"last_4_digits":   "1234",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, billing map[string]interface{}) {
				assert.NotNil(t, billing)
				assert.Equal(t, "enterprise", billing["plan"])
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - not found (404)",
			orgID:        "999",
			mockStatus:   http.StatusNotFound,
			mockResponse: map[string]interface{}{"error": "organization not found"},
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - unauthorized (401)",
			orgID:        "1",
			mockStatus:   http.StatusUnauthorized,
			mockResponse: map[string]interface{}{"error": "invalid token"},
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "invalid-token"}
			},
		},
		{
			name:         "error - forbidden (403)",
			orgID:        "2",
			mockStatus:   http.StatusForbidden,
			mockResponse: map[string]interface{}{"error": "insufficient permissions"},
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - service unavailable (503)",
			orgID:        "1",
			mockStatus:   http.StatusServiceUnavailable,
			mockResponse: map[string]interface{}{"error": "service temporarily unavailable"},
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
		{
			name:         "error - server error (500)",
			orgID:        "1",
			mockStatus:   http.StatusInternalServerError,
			mockResponse: map[string]interface{}{"error": "internal server error"},
			wantErr:      true,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{Token: "jwt-token"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/organizations/"+tt.orgID+"/billing")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			config := &Config{BaseURL: server.URL}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			assert.NoError(t, err)

			billing, err := client.Organizations.GetBilling(context.Background(), tt.orgID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, billing)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, billing)
				if tt.checkFunc != nil {
					tt.checkFunc(t, billing)
				}
			}
		})
	}
}
