package nexmonyx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestServersService_GetByUUID tests retrieving a server by UUID
func TestServersService_GetByUUID(t *testing.T) {
	tests := []struct {
		name       string
		uuid       string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Server)
	}{
		{
			name:       "successful get by UUID",
			uuid:       "550e8400-e29b-41d4-a716-446655440000",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Server retrieved successfully",
				Data: &Server{
					ServerUUID: "550e8400-e29b-41d4-a716-446655440000",
					Hostname:   "web-server-01",
					MainIP:     "192.168.1.100",
					Status:     "active",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, s *Server) {
				if s.Hostname != "web-server-01" {
					t.Errorf("Expected hostname 'web-server-01', got '%s'", s.Hostname)
				}
				if s.Status != "active" {
					t.Errorf("Expected status 'active', got '%s'", s.Status)
				}
			},
		},
		{
			name:       "server not found",
			uuid:       "invalid-uuid",
			mockStatus: http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Server not found",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			mockStatus: http.StatusInternalServerError,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != fmt.Sprintf("/v1/server/%s/details", tt.uuid) {
					t.Errorf("Expected path '/v1/server/%s/details', got '%s'", tt.uuid, r.URL.Path)
				}
				if r.Method != http.MethodGet {
					t.Errorf("Expected method GET, got %s", r.Method)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})

			result, err := client.Servers.GetByUUID(context.Background(), tt.uuid)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetByUUID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

// TestServersService_Get tests the deprecated Get method
func TestServersService_Get(t *testing.T) {
	t.Run("redirects to GetByUUID", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(StandardResponse{
				Status: "success",
				Data: &Server{
					ServerUUID: "550e8400-e29b-41d4-a716-446655440000",
					Hostname:   "test-server",
				},
			})
		}))
		defer server.Close()

		client, _ := NewClient(&Config{
			BaseURL: server.URL,
			Auth:    AuthConfig{Token: "test-token"},
		})

		result, err := client.Servers.Get(context.Background(), "550e8400-e29b-41d4-a716-446655440000")
		if err != nil {
			t.Errorf("Get() error = %v, want nil", err)
		}
		if result.Hostname != "test-server" {
			t.Errorf("Expected hostname 'test-server', got '%s'", result.Hostname)
		}
	})
}

// TestServersService_List tests listing servers with pagination
func TestServersService_List(t *testing.T) {
	tests := []struct {
		name       string
		opts       *ListOptions
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, []*Server, *PaginationMeta)
	}{
		{
			name: "successful list with pagination",
			opts: &ListOptions{Page: 1, Limit: 25},
			mockStatus: http.StatusOK,
			mockBody: PaginatedResponse{
				Status:  "success",
				Message: "Servers retrieved successfully",
				Data: &[]*Server{
					{Hostname: "server-01", Status: "active"},
					{Hostname: "server-02", Status: "inactive"},
				},
				Meta: &PaginationMeta{
					Page:       1,
					TotalPages: 1,
					TotalItems: 2,
					Limit:      25,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, servers []*Server, meta *PaginationMeta) {
				if len(servers) != 2 {
					t.Errorf("Expected 2 servers, got %d", len(servers))
				}
				if meta.TotalItems != 2 {
					t.Errorf("Expected TotalItems 2, got %d", meta.TotalItems)
				}
			},
		},
		{
			name:       "empty list",
			opts:       &ListOptions{Page: 1, Limit: 25},
			mockStatus: http.StatusOK,
			mockBody: PaginatedResponse{
				Status:  "success",
				Message: "No servers found",
				Data:    &[]*Server{},
				Meta: &PaginationMeta{
					Page:       1,
					TotalPages: 0,
					TotalItems: 0,
					Limit:      25,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, servers []*Server, meta *PaginationMeta) {
				if len(servers) != 0 {
					t.Errorf("Expected 0 servers, got %d", len(servers))
				}
			},
		},
		{
			name:       "with search filter",
			opts:       &ListOptions{Search: "web-server", Page: 1, Limit: 10},
			mockStatus: http.StatusOK,
			mockBody: PaginatedResponse{
				Status:  "success",
				Message: "Servers retrieved successfully",
				Data: &[]*Server{
					{Hostname: "web-server-01", Status: "active"},
				},
				Meta: &PaginationMeta{
					Page:       1,
					TotalPages: 1,
					TotalItems: 1,
					Limit:      10,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v2/servers" {
					t.Errorf("Expected path '/v2/servers', got '%s'", r.URL.Path)
				}
				if r.Method != http.MethodGet {
					t.Errorf("Expected method GET, got %s", r.Method)
				}

				// Check query parameters
				if tt.opts != nil && tt.opts.Search != "" {
					if r.URL.Query().Get("search") != tt.opts.Search {
						t.Errorf("Expected search '%s', got '%s'", tt.opts.Search, r.URL.Query().Get("search"))
					}
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})

			servers, meta, err := client.Servers.List(context.Background(), tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, servers, meta)
			}
		})
	}
}

// TestServersService_Delete tests deleting a server
func TestServersService_Delete(t *testing.T) {
	tests := []struct {
		name       string
		serverID   string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:       "successful delete",
			serverID:   "550e8400-e29b-41d4-a716-446655440000",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Server deleted successfully",
			},
			wantErr: false,
		},
		{
			name:       "server not found",
			serverID:   "invalid-uuid",
			mockStatus: http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Server not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			serverID:   "550e8400-e29b-41d4-a716-446655440000",
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
				expectedPath := fmt.Sprintf("/v1/admin/server/%s", tt.serverID)
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path '%s', got '%s'", expectedPath, r.URL.Path)
				}
				if r.Method != http.MethodDelete {
					t.Errorf("Expected method DELETE, got %s", r.Method)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})

			err := client.Servers.Delete(context.Background(), tt.serverID)

			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestServersService_Register tests server registration
func TestServersService_Register(t *testing.T) {
	tests := []struct {
		name       string
		hostname   string
		orgID      uint
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Server)
	}{
		{
			name:       "successful registration",
			hostname:   "new-server-01",
			orgID:      1,
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Server registered successfully",
				Data: &Server{
					ServerUUID:     "550e8400-e29b-41d4-a716-446655440000",
					Hostname:       "new-server-01",
					OrganizationID: 1,
					Status:         "active",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, s *Server) {
				if s.Hostname != "new-server-01" {
					t.Errorf("Expected hostname 'new-server-01', got '%s'", s.Hostname)
				}
				if s.OrganizationID != 1 {
					t.Errorf("Expected organization ID 1, got %d", s.OrganizationID)
				}
			},
		},
		{
			name:       "registration failure",
			hostname:   "duplicate-server",
			orgID:      1,
			mockStatus: http.StatusConflict,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Server already exists",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/register" {
					t.Errorf("Expected path '/v1/register', got '%s'", r.URL.Path)
				}
				if r.Method != http.MethodPost {
					t.Errorf("Expected method POST, got %s", r.Method)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})

			result, err := client.Servers.Register(context.Background(), tt.hostname, tt.orgID)

			if (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

// TestServersService_Heartbeat tests sending heartbeat
func TestServersService_Heartbeat(t *testing.T) {
	tests := []struct {
		name       string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:       "successful heartbeat",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Heartbeat received",
			},
			wantErr: false,
		},
		{
			name:       "heartbeat failure",
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
				if r.URL.Path != "/v1/heartbeat" {
					t.Errorf("Expected path '/v1/heartbeat', got '%s'", r.URL.Path)
				}
				if r.Method != http.MethodPost {
					t.Errorf("Expected method POST, got %s", r.Method)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth: AuthConfig{
					ServerUUID:   "550e8400-e29b-41d4-a716-446655440000",
					ServerSecret: "test-secret",
				},
			})

			err := client.Servers.Heartbeat(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("Heartbeat() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestServersService_HeartbeatWithVersion tests sending heartbeat with agent version
func TestServersService_HeartbeatWithVersion(t *testing.T) {
	tests := []struct {
		name         string
		agentVersion string
		mockStatus   int
		mockBody     interface{}
		wantErr      bool
	}{
		{
			name:         "successful heartbeat with version",
			agentVersion: "1.2.3",
			mockStatus:   http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Heartbeat received",
			},
			wantErr: false,
		},
		{
			name:         "empty version",
			agentVersion: "",
			mockStatus:   http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Heartbeat received",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/heartbeat" {
					t.Errorf("Expected path '/v1/heartbeat', got '%s'", r.URL.Path)
				}
				if r.Method != http.MethodPost {
					t.Errorf("Expected method POST, got %s", r.Method)
				}

				// Verify body contains agent_version
				var body map[string]string
				json.NewDecoder(r.Body).Decode(&body)
				if body["agent_version"] != tt.agentVersion {
					t.Errorf("Expected agent_version '%s', got '%s'", tt.agentVersion, body["agent_version"])
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth: AuthConfig{
					ServerUUID:   "550e8400-e29b-41d4-a716-446655440000",
					ServerSecret: "test-secret",
				},
			})

			err := client.Servers.HeartbeatWithVersion(context.Background(), tt.agentVersion)

			if (err != nil) != tt.wantErr {
				t.Errorf("HeartbeatWithVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestServersService_UpdateDetails tests updating server details
func TestServersService_UpdateDetails(t *testing.T) {
	tests := []struct {
		name       string
		serverUUID string
		req        *ServerDetailsUpdateRequest
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Server)
	}{
		{
			name:       "successful update",
			serverUUID: "550e8400-e29b-41d4-a716-446655440000",
			req: &ServerDetailsUpdateRequest{
				Hostname:     "updated-server",
				OS:           "Ubuntu",
				OSVersion:    "22.04",
				CPUModel:     "Intel Xeon",
				CPUCores:     8,
				MemoryTotal:  16384,
				StorageTotal: 500000,
			},
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Server details updated",
				Data: &Server{
					ServerUUID:  "550e8400-e29b-41d4-a716-446655440000",
					Hostname:    "updated-server",
					OS:          "Ubuntu",
					OSVersion:   "22.04",
					CPUModel:    "Intel Xeon",
					CPUCores:    8,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, s *Server) {
				if s.Hostname != "updated-server" {
					t.Errorf("Expected hostname 'updated-server', got '%s'", s.Hostname)
				}
				if s.CPUCores != 8 {
					t.Errorf("Expected CPUCores 8, got %d", s.CPUCores)
				}
			},
		},
		{
			name:       "update failure",
			serverUUID: "invalid-uuid",
			req: &ServerDetailsUpdateRequest{
				Hostname: "test-server",
			},
			mockStatus: http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Server not found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := fmt.Sprintf("/v1/server/%s/details", tt.serverUUID)
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path '%s', got '%s'", expectedPath, r.URL.Path)
				}
				if r.Method != http.MethodPut {
					t.Errorf("Expected method PUT, got %s", r.Method)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth: AuthConfig{
					ServerUUID:   "550e8400-e29b-41d4-a716-446655440000",
					ServerSecret: "test-secret",
				},
			})

			result, err := client.Servers.UpdateDetails(context.Background(), tt.serverUUID, tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateDetails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

// TestServersService_UpdateInfo tests updating server info
func TestServersService_UpdateInfo(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/v1/server/550e8400-e29b-41d4-a716-446655440000/info" {
				t.Errorf("Unexpected path: %s", r.URL.Path)
			}
			if r.Method != http.MethodPut {
				t.Errorf("Expected PUT, got %s", r.Method)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(StandardResponse{
				Status:  "success",
				Message: "Server info updated",
				Data: &Server{
					ServerUUID: "550e8400-e29b-41d4-a716-446655440000",
					Hostname:   "updated-server",
				},
			})
		}))
		defer server.Close()

		client, _ := NewClient(&Config{
			BaseURL: server.URL,
			Auth: AuthConfig{
				ServerUUID:   "550e8400-e29b-41d4-a716-446655440000",
				ServerSecret: "test-secret",
			},
		})

		req := &ServerDetailsUpdateRequest{
			Hostname: "updated-server",
		}

		result, err := client.Servers.UpdateInfo(context.Background(), "550e8400-e29b-41d4-a716-446655440000", req)
		if err != nil {
			t.Errorf("UpdateInfo() error = %v, want nil", err)
		}
		if result.Hostname != "updated-server" {
			t.Errorf("Expected hostname 'updated-server', got '%s'", result.Hostname)
		}
	})
}

// TestServersService_GetDetails tests retrieving server details
func TestServersService_GetDetails(t *testing.T) {
	tests := []struct {
		name       string
		serverUUID string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Server)
	}{
		{
			name:       "successful get details",
			serverUUID: "550e8400-e29b-41d4-a716-446655440000",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Server details retrieved",
				Data: &Server{
					ServerUUID: "550e8400-e29b-41d4-a716-446655440000",
					Hostname:   "web-server-01",
					OS:         "Ubuntu",
					OSVersion:  "22.04",
					CPUCores:   4,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, s *Server) {
				if s.Hostname != "web-server-01" {
					t.Errorf("Expected hostname 'web-server-01', got '%s'", s.Hostname)
				}
				if s.OS != "Ubuntu" {
					t.Errorf("Expected OS 'Ubuntu', got '%s'", s.OS)
				}
			},
		},
		{
			name:       "server not found",
			serverUUID: "invalid-uuid",
			mockStatus: http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Server not found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := fmt.Sprintf("/v1/server/%s/details", tt.serverUUID)
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path '%s', got '%s'", expectedPath, r.URL.Path)
				}
				if r.Method != http.MethodGet {
					t.Errorf("Expected method GET, got %s", r.Method)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})

			result, err := client.Servers.GetDetails(context.Background(), tt.serverUUID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetDetails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

// TestServersService_RegisterWithKey tests registering with a registration key
func TestServersService_RegisterWithKey(t *testing.T) {
	tests := []struct {
		name            string
		registrationKey string
		req             *ServerCreateRequest
		mockStatus      int
		mockBody        interface{}
		wantErr         bool
		checkFunc       func(*testing.T, *Server)
	}{
		{
			name:            "successful registration with key",
			registrationKey: "reg-key-123",
			req: &ServerCreateRequest{
				Hostname: "new-server",
				MainIP:   "192.168.1.100",
			},
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Server registered",
				Data: &ServerRegistrationResponse{
					Server: &Server{
						ServerUUID:     "550e8400-e29b-41d4-a716-446655440000",
						Hostname:       "new-server",
						MainIP:         "192.168.1.100",
						OrganizationID: 1,
						Status:         "active",
					},
					ServerSecret: "generated-secret",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, s *Server) {
				if s.Hostname != "new-server" {
					t.Errorf("Expected hostname 'new-server', got '%s'", s.Hostname)
				}
				if s.Status != "active" {
					t.Errorf("Expected status 'active', got '%s'", s.Status)
				}
			},
		},
		{
			name:            "invalid registration key",
			registrationKey: "invalid-key",
			req: &ServerCreateRequest{
				Hostname: "new-server",
			},
			mockStatus: http.StatusUnauthorized,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Invalid registration key",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/register" {
					t.Errorf("Expected path '/v1/register', got '%s'", r.URL.Path)
				}
				if r.Method != http.MethodPost {
					t.Errorf("Expected method POST, got %s", r.Method)
				}

				// Verify registration key header
				if r.Header.Get("X-Registration-Key") != tt.registrationKey {
					t.Errorf("Expected registration key '%s', got '%s'", tt.registrationKey, r.Header.Get("X-Registration-Key"))
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
			})

			result, err := client.Servers.RegisterWithKey(context.Background(), tt.registrationKey, tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("RegisterWithKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

// TestServersService_GetSystemInfo tests retrieving system information
func TestServersService_GetSystemInfo(t *testing.T) {
	tests := []struct {
		name       string
		serverID   string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *SystemInfo)
	}{
		{
			name:       "successful get system info",
			serverID:   "550e8400-e29b-41d4-a716-446655440000",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "System info retrieved",
				Data: &SystemInfo{
					Hostname:  "web-server-01",
					Uptime:    86400,
					Processes: 120,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, info *SystemInfo) {
				if info.Hostname != "web-server-01" {
					t.Errorf("Expected hostname 'web-server-01', got '%s'", info.Hostname)
				}
				if info.Uptime != 86400 {
					t.Errorf("Expected uptime 86400, got %d", info.Uptime)
				}
				if info.Processes != 120 {
					t.Errorf("Expected processes 120, got %d", info.Processes)
				}
			},
		},
		{
			name:       "system info not found",
			serverID:   "invalid-uuid",
			mockStatus: http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Server not found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := fmt.Sprintf("/v1/server/%s/system-info", tt.serverID)
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path '%s', got '%s'", expectedPath, r.URL.Path)
				}
				if r.Method != http.MethodGet {
					t.Errorf("Expected method GET, got %s", r.Method)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})

			result, err := client.Servers.GetSystemInfo(context.Background(), tt.serverID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetSystemInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}
