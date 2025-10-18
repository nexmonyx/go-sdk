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

// TestServersService_Create_Handler tests creating a new server
func TestServersService_Create_Handler(t *testing.T) {
	tests := []struct {
		name           string
		mockStatus     int
		mockResponse   interface{}
		wantErr        bool
		checkFunc      func(*testing.T, *Server)
		clientAuthFunc func(*Config)
	}{
		{
			name:       "success - create server",
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"id":       1,
					"uuid":     "server-uuid-001",
					"hostname": "web-server-01",
					"main_ip":  "192.168.1.100",
					"os":       "Ubuntu",
					"os_version": "22.04",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, server *Server) {
				assert.Equal(t, "web-server-01", server.Hostname)
				assert.Equal(t, "192.168.1.100", server.MainIP)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - bad request (400)",
			mockStatus:     http.StatusBadRequest,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - unauthorized (401)",
			mockStatus:     http.StatusUnauthorized,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "invalid-key", APISecret: "invalid-secret"}
			},
		},
		{
			name:           "error - forbidden (403)",
			mockStatus:     http.StatusForbidden,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - conflict (409)",
			mockStatus:     http.StatusConflict,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - server error (500)",
			mockStatus:     http.StatusInternalServerError,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			config := &Config{
				BaseURL: server.URL,
			}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			require.NoError(t, err)

			newServer := &Server{
				Hostname: "web-server-01",
				MainIP:   "192.168.1.100",
				OS:       "Ubuntu",
				OSVersion: "22.04",
			}

			result, err := client.Servers.Create(context.Background(), newServer)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestServersService_Get_Handler tests retrieving a server by ID/UUID
func TestServersService_Get_Handler(t *testing.T) {
	tests := []struct {
		name           string
		serverID       string
		mockStatus     int
		mockResponse   interface{}
		wantErr        bool
		checkFunc      func(*testing.T, *Server)
		clientAuthFunc func(*Config)
	}{
		{
			name:       "success - get server by UUID",
			serverID:   "server-uuid-001",
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"id":       1,
					"server_uuid":     "server-uuid-001",
					"hostname": "web-server-01",
					"main_ip":  "192.168.1.100",
					"os":       "Ubuntu",
					"os_version": "22.04",
					"cpu_cores": 8,
					"total_memory_gb": 32,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, server *Server) {
				assert.Equal(t, "server-uuid-001", server.ServerUUID)
				assert.Equal(t, "web-server-01", server.Hostname)
				assert.Equal(t, 8, server.CPUCores)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - not found (404)",
			serverID:       "nonexistent-uuid",
			mockStatus:     http.StatusNotFound,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - unauthorized (401)",
			serverID:       "server-uuid-001",
			mockStatus:     http.StatusUnauthorized,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "invalid-key", APISecret: "invalid-secret"}
			},
		},
		{
			name:           "error - forbidden (403)",
			serverID:       "server-uuid-001",
			mockStatus:     http.StatusForbidden,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - server error (500)",
			serverID:       "server-uuid-001",
			mockStatus:     http.StatusInternalServerError,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			config := &Config{
				BaseURL: server.URL,
			}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			require.NoError(t, err)

			result, err := client.Servers.Get(context.Background(), tt.serverID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestServersService_GetByUUID_Handler tests retrieving a server by UUID
func TestServersService_GetByUUID_Handler(t *testing.T) {
	tests := []struct {
		name           string
		serverUUID     string
		mockStatus     int
		mockResponse   interface{}
		wantErr        bool
		checkFunc      func(*testing.T, *Server)
		clientAuthFunc func(*Config)
	}{
		{
			name:       "success - get server by UUID directly",
			serverUUID: "server-uuid-002",
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"id":       2,
					"server_uuid":     "server-uuid-002",
					"hostname": "db-server-01",
					"main_ip":  "192.168.1.101",
					"os":       "CentOS",
					"os_version": "8.5",
					"environment": "production",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, server *Server) {
				assert.Equal(t, "server-uuid-002", server.ServerUUID)
				assert.Equal(t, "db-server-01", server.Hostname)
				assert.Equal(t, "production", server.Environment)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - not found (404)",
			serverUUID:     "invalid-uuid",
			mockStatus:     http.StatusNotFound,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - unauthorized (401)",
			serverUUID:     "server-uuid-002",
			mockStatus:     http.StatusUnauthorized,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "invalid-key", APISecret: "invalid-secret"}
			},
		},
		{
			name:           "error - server error (500)",
			serverUUID:     "server-uuid-002",
			mockStatus:     http.StatusInternalServerError,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			config := &Config{
				BaseURL: server.URL,
			}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			require.NoError(t, err)

			result, err := client.Servers.GetByUUID(context.Background(), tt.serverUUID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestServersService_List_Handler tests listing servers with pagination
func TestServersService_List_Handler(t *testing.T) {
	tests := []struct {
		name           string
		mockStatus     int
		mockResponse   interface{}
		wantErr        bool
		wantCount      int
		checkFunc      func(*testing.T, []*Server, *PaginationMeta)
		clientAuthFunc func(*Config)
	}{
		{
			name:       "success - list servers with pagination",
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id":       1,
						"uuid":     "server-uuid-001",
						"hostname": "web-server-01",
						"main_ip":  "192.168.1.100",
						"os":       "Ubuntu",
					},
					{
						"id":       2,
						"uuid":     "server-uuid-002",
						"hostname": "web-server-02",
						"main_ip":  "192.168.1.101",
						"os":       "Ubuntu",
					},
				},
				"meta": map[string]interface{}{
					"total_items": 10,
					"total_pages": 5,
					"page":        1,
					"limit":       2,
				},
			},
			wantErr:   false,
			wantCount: 2,
			checkFunc: func(t *testing.T, servers []*Server, meta *PaginationMeta) {
				assert.Equal(t, 2, len(servers))
				assert.Equal(t, "web-server-01", servers[0].Hostname)
				assert.Equal(t, "web-server-02", servers[1].Hostname)
				assert.Equal(t, 10, meta.TotalItems)
				assert.Equal(t, 5, meta.TotalPages)
				assert.Equal(t, 1, meta.Page)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:       "success - empty list",
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": []interface{}{},
				"meta": map[string]interface{}{
					"total_items": 0,
					"total_pages": 0,
					"page":        1,
					"limit":       25,
				},
			},
			wantErr:   false,
			wantCount: 0,
			checkFunc: func(t *testing.T, servers []*Server, meta *PaginationMeta) {
				assert.Equal(t, 0, len(servers))
				assert.Equal(t, 0, meta.TotalItems)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:       "success - list with filters",
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id":       3,
						"uuid":     "server-uuid-003",
						"hostname": "prod-server-01",
						"environment": "production",
						"os": "CentOS",
					},
				},
				"meta": map[string]interface{}{
					"total_items": 1,
					"total_pages": 1,
					"page":        1,
					"limit":       25,
				},
			},
			wantErr:   false,
			wantCount: 1,
			checkFunc: func(t *testing.T, servers []*Server, meta *PaginationMeta) {
				assert.Equal(t, 1, len(servers))
				assert.Equal(t, "production", servers[0].Environment)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - unauthorized (401)",
			mockStatus:     http.StatusUnauthorized,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			wantCount:      0,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "invalid-key", APISecret: "invalid-secret"}
			},
		},
		{
			name:           "error - forbidden (403)",
			mockStatus:     http.StatusForbidden,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			wantCount:      0,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - bad request (400)",
			mockStatus:     http.StatusBadRequest,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			wantCount:      0,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - server error (500)",
			mockStatus:     http.StatusInternalServerError,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			wantCount:      0,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			config := &Config{
				BaseURL: server.URL,
			}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			require.NoError(t, err)

			opts := &ListOptions{Page: 1, Limit: 25}
			servers, meta, err := client.Servers.List(context.Background(), opts)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, servers)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, servers)
				assert.Equal(t, tt.wantCount, len(servers))
				if tt.checkFunc != nil {
					tt.checkFunc(t, servers, meta)
				}
			}
		})
	}
}

// TestServersService_Update_Handler tests updating an existing server
func TestServersService_Update_Handler(t *testing.T) {
	tests := []struct {
		name           string
		serverID       string
		mockStatus     int
		mockResponse   interface{}
		wantErr        bool
		checkFunc      func(*testing.T, *Server)
		clientAuthFunc func(*Config)
	}{
		{
			name:       "success - update server details",
			serverID:   "server-uuid-001",
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"id":           1,
					"server_uuid":  "server-uuid-001",
					"hostname":     "web-server-updated",
					"main_ip":      "192.168.1.150",
					"environment":  "staging",
					"location":     "US-WEST",
					"os":           "Ubuntu",
					"os_version":   "22.04",
					"cpu_cores":    16,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, server *Server) {
				assert.Equal(t, "web-server-updated", server.Hostname)
				assert.Equal(t, "staging", server.Environment)
				assert.Equal(t, 16, server.CPUCores)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - bad request (400)",
			serverID:       "server-uuid-001",
			mockStatus:     http.StatusBadRequest,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - not found (404)",
			serverID:       "nonexistent-uuid",
			mockStatus:     http.StatusNotFound,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - unauthorized (401)",
			serverID:       "server-uuid-001",
			mockStatus:     http.StatusUnauthorized,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "invalid-key", APISecret: "invalid-secret"}
			},
		},
		{
			name:           "error - forbidden (403)",
			serverID:       "server-uuid-001",
			mockStatus:     http.StatusForbidden,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - conflict (409)",
			serverID:       "server-uuid-001",
			mockStatus:     http.StatusConflict,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - server error (500)",
			serverID:       "server-uuid-001",
			mockStatus:     http.StatusInternalServerError,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Contains(t, []string{"PUT", "PATCH", "POST"}, r.Method)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			config := &Config{
				BaseURL: server.URL,
			}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			require.NoError(t, err)

			updatedServer := &Server{
				Hostname:    "web-server-updated",
				MainIP:      "192.168.1.150",
				Environment: "staging",
				Location:    "US-WEST",
				CPUCores:    16,
			}

			result, err := client.Servers.Update(context.Background(), tt.serverID, updatedServer)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestServersService_Delete_Handler tests deleting a server
func TestServersService_Delete_Handler(t *testing.T) {
	tests := []struct {
		name           string
		serverID       string
		mockStatus     int
		mockResponse   interface{}
		wantErr        bool
		clientAuthFunc func(*Config)
	}{
		{
			name:         "success - delete server",
			serverID:     "server-uuid-001",
			mockStatus:   http.StatusOK,
			mockResponse: map[string]interface{}{},
			wantErr:      false,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:         "success - delete returns 204 No Content",
			serverID:     "server-uuid-002",
			mockStatus:   http.StatusNoContent,
			mockResponse: nil,
			wantErr:      false,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - not found (404)",
			serverID:       "nonexistent-uuid",
			mockStatus:     http.StatusNotFound,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - unauthorized (401)",
			serverID:       "server-uuid-001",
			mockStatus:     http.StatusUnauthorized,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "invalid-key", APISecret: "invalid-secret"}
			},
		},
		{
			name:           "error - forbidden (403) - insufficient permissions",
			serverID:       "server-uuid-001",
			mockStatus:     http.StatusForbidden,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - conflict (409)",
			serverID:       "server-uuid-001",
			mockStatus:     http.StatusConflict,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - server error (500)",
			serverID:       "server-uuid-001",
			mockStatus:     http.StatusInternalServerError,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "DELETE", r.Method)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}))
			defer server.Close()

			config := &Config{
				BaseURL: server.URL,
			}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			require.NoError(t, err)

			err = client.Servers.Delete(context.Background(), tt.serverID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestServersService_Register_Handler tests registering a new server
func TestServersService_Register_Handler(t *testing.T) {
	tests := []struct {
		name           string
		hostname       string
		organizationID uint
		mockStatus     int
		mockResponse   interface{}
		wantErr        bool
		checkFunc      func(*testing.T, *Server)
		clientAuthFunc func(*Config)
	}{
		{
			name:           "success - register server",
			hostname:       "new-server-01",
			organizationID: 1,
			mockStatus:     http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"id":              1,
					"server_uuid":     "new-server-uuid-001",
					"hostname":        "new-server-01",
					"organization_id": 1,
					"main_ip":         "192.168.1.200",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, server *Server) {
				assert.Equal(t, "new-server-01", server.Hostname)
				assert.Equal(t, "new-server-uuid-001", server.ServerUUID)
				assert.Equal(t, uint(1), server.OrganizationID)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "success - register server with different org",
			hostname:       "prod-server-01",
			organizationID: 5,
			mockStatus:     http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"id":              2,
					"server_uuid":     "prod-server-uuid-001",
					"hostname":        "prod-server-01",
					"organization_id": 5,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, server *Server) {
				assert.Equal(t, "prod-server-01", server.Hostname)
				assert.Equal(t, uint(5), server.OrganizationID)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - bad request (400)",
			hostname:       "invalid@hostname",
			organizationID: 1,
			mockStatus:     http.StatusBadRequest,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - unauthorized (401)",
			hostname:       "new-server-01",
			organizationID: 1,
			mockStatus:     http.StatusUnauthorized,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "invalid-key", APISecret: "invalid-secret"}
			},
		},
		{
			name:           "error - forbidden (403)",
			hostname:       "new-server-01",
			organizationID: 1,
			mockStatus:     http.StatusForbidden,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - conflict (409) - hostname already exists",
			hostname:       "existing-server",
			organizationID: 1,
			mockStatus:     http.StatusConflict,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - organization not found (404)",
			hostname:       "new-server-01",
			organizationID: 999,
			mockStatus:     http.StatusNotFound,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:           "error - server error (500)",
			hostname:       "new-server-01",
			organizationID: 1,
			mockStatus:     http.StatusInternalServerError,
			mockResponse:   map[string]interface{}{},
			wantErr:        true,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			config := &Config{
				BaseURL: server.URL,
			}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			require.NoError(t, err)

			result, err := client.Servers.Register(context.Background(), tt.hostname, tt.organizationID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// Phase 1C: GetMetrics and GetAlerts Handler Tests

func TestServersService_GetMetrics_Handler(t *testing.T) {
	tests := []struct {
		name         string
		serverID     string
		opts         *ListOptions
		mockStatus   int
		mockResponse interface{}
		wantErr      bool
		wantLen      int
		checkFunc    func(*testing.T, []*Metric, *PaginationMeta)
		clientAuthFunc func(*Config)
	}{
		{
			name:       "success - get metrics with data",
			serverID:   "server-123",
			opts:       nil,
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"server_uuid": "server-123",
						"name":        "cpu_usage",
						"value":       45.5,
						"unit":        "percent",
						"timestamp":   "2021-10-01T12:00:00Z",
					},
					{
						"server_uuid": "server-123",
						"name":        "memory_usage",
						"value":       62.3,
						"unit":        "percent",
						"timestamp":   "2021-10-01T12:01:00Z",
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
			checkFunc: func(t *testing.T, metrics []*Metric, meta *PaginationMeta) {
				if len(metrics) > 0 && metrics[0] != nil {
					assert.Equal(t, "cpu_usage", metrics[0].Name)
					assert.Equal(t, float64(45.5), metrics[0].Value)
				}
				assert.Equal(t, 2, meta.TotalItems)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:       "success - get metrics with pagination",
			serverID:   "server-456",
			opts:       &ListOptions{Page: 2, Limit: 10},
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"server_uuid": "server-456",
						"name":        "disk_usage",
						"value":       55.0,
						"unit":        "percent",
						"timestamp":   "2021-10-02T14:30:00Z",
					},
				},
				"meta": map[string]interface{}{
					"total_items": 50,
					"limit":       10,
					"page":        2,
					"per_page":    10,
				},
			},
			wantErr: false,
			wantLen: 1,
			checkFunc: func(t *testing.T, metrics []*Metric, meta *PaginationMeta) {
				assert.Equal(t, 50, meta.TotalItems)
				assert.Equal(t, 10, meta.Limit)
				assert.Equal(t, 2, meta.Page)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:       "success - empty metrics list",
			serverID:   "server-789",
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
			checkFunc: func(t *testing.T, metrics []*Metric, meta *PaginationMeta) {
				assert.Equal(t, 0, meta.TotalItems)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:         "error - bad request (400)",
			serverID:     "invalid-id",
			opts:         nil,
			mockStatus:   http.StatusBadRequest,
			mockResponse: map[string]interface{}{"error": "invalid server id"},
			wantErr:      true,
			wantLen:      0,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:         "error - unauthorized (401)",
			serverID:     "server-123",
			opts:         nil,
			mockStatus:   http.StatusUnauthorized,
			mockResponse: map[string]interface{}{"error": "invalid api key"},
			wantErr:      true,
			wantLen:      0,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "invalid-key", APISecret: "invalid-secret"}
			},
		},
		{
			name:         "error - forbidden (403)",
			serverID:     "server-123",
			opts:         nil,
			mockStatus:   http.StatusForbidden,
			mockResponse: map[string]interface{}{"error": "insufficient permissions"},
			wantErr:      true,
			wantLen:      0,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:         "error - not found (404)",
			serverID:     "nonexistent-server",
			opts:         nil,
			mockStatus:   http.StatusNotFound,
			mockResponse: map[string]interface{}{"error": "server not found"},
			wantErr:      true,
			wantLen:      0,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:         "error - server error (500)",
			serverID:     "server-123",
			opts:         nil,
			mockStatus:   http.StatusInternalServerError,
			mockResponse: map[string]interface{}{"error": "internal server error"},
			wantErr:      true,
			wantLen:      0,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:         "error - service unavailable (503)",
			serverID:     "server-123",
			opts:         nil,
			mockStatus:   http.StatusServiceUnavailable,
			mockResponse: map[string]interface{}{"error": "service temporarily unavailable"},
			wantErr:      true,
			wantLen:      0,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method and path
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/server/"+tt.serverID+"/metrics")

				// Set status and response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			// Create client with mock server
			config := &Config{
				BaseURL: server.URL,
			}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			assert.NoError(t, err)

			// Call method
			metrics, meta, err := client.Servers.GetMetrics(context.Background(), tt.serverID, tt.opts)

			// Verify results
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, metrics)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantLen, len(metrics))
				if tt.checkFunc != nil {
					tt.checkFunc(t, metrics, meta)
				}
			}
		})
	}
}

func TestServersService_GetAlerts_Handler(t *testing.T) {
	tests := []struct {
		name         string
		serverID     string
		opts         *ListOptions
		mockStatus   int
		mockResponse interface{}
		wantErr      bool
		wantLen      int
		checkFunc    func(*testing.T, []*Alert, *PaginationMeta)
		clientAuthFunc func(*Config)
	}{
		{
			name:       "success - get alerts with data",
			serverID:   "server-123",
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
						"server_id":        1,
					},
					{
						"id":               2,
						"name":             "High Memory Usage",
						"type":             "metric_threshold",
						"metric_name":      "memory_usage",
						"condition":        ">",
						"threshold":        85.0,
						"organization_id":  1,
						"server_id":        1,
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
				if len(alerts) > 0 && alerts[0] != nil {
					assert.Equal(t, "High CPU Usage", alerts[0].Name)
					assert.Equal(t, "metric_threshold", alerts[0].Type)
				}
				assert.Equal(t, 2, meta.TotalItems)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:       "success - get alerts with pagination",
			serverID:   "server-456",
			opts:       &ListOptions{Page: 1, Limit: 20},
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
						"server_id":        1,
					},
				},
				"meta": map[string]interface{}{
					"total_items": 35,
					"limit":       20,
					"page":        1,
					"per_page":    20,
				},
			},
			wantErr: false,
			wantLen: 1,
			checkFunc: func(t *testing.T, alerts []*Alert, meta *PaginationMeta) {
				assert.Equal(t, 35, meta.TotalItems)
				assert.Equal(t, 20, meta.Limit)
				assert.Equal(t, 1, meta.Page)
			},
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:       "success - empty alerts list",
			serverID:   "server-789",
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
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:         "error - bad request (400)",
			serverID:     "invalid-id",
			opts:         nil,
			mockStatus:   http.StatusBadRequest,
			mockResponse: map[string]interface{}{"error": "invalid server id"},
			wantErr:      true,
			wantLen:      0,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:         "error - unauthorized (401)",
			serverID:     "server-123",
			opts:         nil,
			mockStatus:   http.StatusUnauthorized,
			mockResponse: map[string]interface{}{"error": "invalid api key"},
			wantErr:      true,
			wantLen:      0,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "invalid-key", APISecret: "invalid-secret"}
			},
		},
		{
			name:         "error - forbidden (403)",
			serverID:     "server-123",
			opts:         nil,
			mockStatus:   http.StatusForbidden,
			mockResponse: map[string]interface{}{"error": "insufficient permissions"},
			wantErr:      true,
			wantLen:      0,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:         "error - not found (404)",
			serverID:     "nonexistent-server",
			opts:         nil,
			mockStatus:   http.StatusNotFound,
			mockResponse: map[string]interface{}{"error": "server not found"},
			wantErr:      true,
			wantLen:      0,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:         "error - server error (500)",
			serverID:     "server-123",
			opts:         nil,
			mockStatus:   http.StatusInternalServerError,
			mockResponse: map[string]interface{}{"error": "internal server error"},
			wantErr:      true,
			wantLen:      0,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
		{
			name:         "error - service unavailable (503)",
			serverID:     "server-123",
			opts:         nil,
			mockStatus:   http.StatusServiceUnavailable,
			mockResponse: map[string]interface{}{"error": "service temporarily unavailable"},
			wantErr:      true,
			wantLen:      0,
			checkFunc:    nil,
			clientAuthFunc: func(c *Config) {
				c.Auth = AuthConfig{APIKey: "test-key", APISecret: "test-secret"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method and path
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/server/"+tt.serverID+"/alerts")

				// Set status and response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			// Create client with mock server
			config := &Config{
				BaseURL: server.URL,
			}
			tt.clientAuthFunc(config)

			client, err := NewClient(config)
			assert.NoError(t, err)

			// Call method
			alerts, meta, err := client.Servers.GetAlerts(context.Background(), tt.serverID, tt.opts)

			// Verify results
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
