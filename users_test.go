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

// TestUsersService_Get tests the Get method
func TestUsersService_Get(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *User)
	}{
		{
			name:       "successful get",
			id:         "user-123",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status: "success",
				Data: &User{
					GormModel: GormModel{ID: 1},
					Email:     "john.doe@example.com",
					FirstName: "John",
					LastName:  "Doe",
					Role:      "admin",
					IsActive:  true,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, user *User) {
				assert.NotNil(t, user)
				assert.Equal(t, "john.doe@example.com", user.Email)
				assert.Equal(t, "John", user.FirstName)
				assert.Equal(t, "admin", user.Role)
			},
		},
		{
			name:       "user not found",
			id:         "nonexistent",
			mockStatus: http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "User not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			id:         "user-123",
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
				assert.Contains(t, r.URL.Path, "/v1/users/")

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

			result, err := client.Users.Get(context.Background(), tt.id)

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

// TestUsersService_GetByEmail tests the GetByEmail method
func TestUsersService_GetByEmail(t *testing.T) {
	tests := []struct {
		name       string
		email      string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *User)
	}{
		{
			name:       "successful get by email",
			email:      "john.doe@example.com",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status: "success",
				Data: &User{
					GormModel: GormModel{ID: 1},
					Email:     "john.doe@example.com",
					FirstName: "John",
					LastName:  "Doe",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, user *User) {
				assert.NotNil(t, user)
				assert.Equal(t, "john.doe@example.com", user.Email)
			},
		},
		{
			name:       "user not found",
			email:      "nonexistent@example.com",
			mockStatus: http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "User not found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/users/email/")

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

			result, err := client.Users.GetByEmail(context.Background(), tt.email)

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

// TestUsersService_GetCurrent tests the GetCurrent method
func TestUsersService_GetCurrent(t *testing.T) {
	tests := []struct {
		name       string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *User)
	}{
		{
			name:       "successful get current user",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status: "success",
				Data: &User{
					GormModel: GormModel{ID: 1},
					Email:     "current.user@example.com",
					FirstName: "Current",
					LastName:  "User",
					Role:      "user",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, user *User) {
				assert.NotNil(t, user)
				assert.Equal(t, "current.user@example.com", user.Email)
			},
		},
		{
			name:       "unauthorized",
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
				assert.Equal(t, "/v1/users/me", r.URL.Path)

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

			result, err := client.Users.GetCurrent(context.Background())

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

// TestUsersService_List tests the List method
func TestUsersService_List(t *testing.T) {
	tests := []struct {
		name       string
		opts       *ListOptions
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, []*User, *PaginationMeta)
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
				assert.Equal(t, 2, meta.TotalItems)
			},
		},
		{
			name:       "empty list",
			opts:       nil,
			mockStatus: http.StatusOK,
			mockBody: PaginatedResponse{
				Status: "success",
				Data:   []*User{},
				Meta: &PaginationMeta{
					Page:       1,
					Limit:      25,
					TotalItems: 0,
					TotalPages: 0,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, users []*User, meta *PaginationMeta) {
				assert.NotNil(t, users)
				assert.Len(t, users, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/v1/users", r.URL.Path)

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

			result, meta, err := client.Users.List(context.Background(), tt.opts)

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

// TestUsersService_Create tests the Create method
func TestUsersService_Create(t *testing.T) {
	tests := []struct {
		name       string
		user       *User
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *User)
	}{
		{
			name: "successful create",
			user: &User{
				Email:     "newuser@example.com",
				FirstName: "New",
				LastName:  "User",
				Role:      "user",
			},
			mockStatus: http.StatusCreated,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "User created successfully",
				Data: &User{
					GormModel: GormModel{ID: 1},
					Email:     "newuser@example.com",
					FirstName: "New",
					LastName:  "User",
					Role:      "user",
					IsActive:  true,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, user *User) {
				assert.NotNil(t, user)
				assert.Equal(t, uint(1), user.ID)
				assert.Equal(t, "newuser@example.com", user.Email)
			},
		},
		{
			name: "validation error",
			user: &User{
				Email: "", // Empty email
			},
			mockStatus: http.StatusBadRequest,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Validation failed",
				Error:   "Email is required",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/v1/users", r.URL.Path)

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

			result, err := client.Users.Create(context.Background(), tt.user)

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

// TestUsersService_Update tests the Update method
func TestUsersService_Update(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		user       *User
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *User)
	}{
		{
			name: "successful update",
			id:   "user-123",
			user: &User{
				FirstName: "Updated",
				LastName:  "Name",
			},
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "User updated successfully",
				Data: &User{
					GormModel: GormModel{ID: 1},
					Email:     "user@example.com",
					FirstName: "Updated",
					LastName:  "Name",
					IsActive:  true,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, user *User) {
				assert.NotNil(t, user)
				assert.Equal(t, "Updated", user.FirstName)
			},
		},
		{
			name: "user not found",
			id:   "nonexistent",
			user: &User{
				FirstName: "Test",
			},
			mockStatus: http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "User not found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "PUT", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/users/")

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

			result, err := client.Users.Update(context.Background(), tt.id, tt.user)

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

// TestUsersService_Delete tests the Delete method
func TestUsersService_Delete(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:       "successful delete",
			id:         "user-123",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "User deleted successfully",
			},
			wantErr: false,
		},
		{
			name:       "user not found",
			id:         "nonexistent",
			mockStatus: http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "User not found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "DELETE", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/users/")

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

			err = client.Users.Delete(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestUsersService_UpdateRole tests the UpdateRole method
func TestUsersService_UpdateRole(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		role       string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *User)
	}{
		{
			name:       "successful role update",
			id:         "user-123",
			role:       "admin",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status: "success",
				Data: &User{
					GormModel: GormModel{ID: 1},
					Email:     "user@example.com",
					Role:      "admin",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, user *User) {
				assert.NotNil(t, user)
				assert.Equal(t, "admin", user.Role)
			},
		},
		{
			name:       "invalid role",
			id:         "user-123",
			role:       "invalid",
			mockStatus: http.StatusBadRequest,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Invalid role",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "PUT", r.Method)
				assert.Contains(t, r.URL.Path, "/role")

				var body map[string]interface{}
				json.NewDecoder(r.Body).Decode(&body)
				assert.Equal(t, tt.role, body["role"])

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

			result, err := client.Users.UpdateRole(context.Background(), tt.id, tt.role)

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

// TestUsersService_UpdatePermissions tests the UpdatePermissions method
func TestUsersService_UpdatePermissions(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		permissions []string
		mockStatus  int
		mockBody    interface{}
		wantErr     bool
		checkFunc   func(*testing.T, *User)
	}{
		{
			name:        "successful permissions update",
			id:          "user-123",
			permissions: []string{"read", "write", "delete"},
			mockStatus:  http.StatusOK,
			mockBody: StandardResponse{
				Status: "success",
				Data: &User{
					GormModel:   GormModel{ID: 1},
					Email:       "user@example.com",
					Permissions: []string{"read", "write", "delete"},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, user *User) {
				assert.NotNil(t, user)
				assert.Len(t, user.Permissions, 3)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "PUT", r.Method)
				assert.Contains(t, r.URL.Path, "/permissions")

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

			result, err := client.Users.UpdatePermissions(context.Background(), tt.id, tt.permissions)

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

// TestUsersService_UpdatePreferences tests the UpdatePreferences method
func TestUsersService_UpdatePreferences(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		preferences map[string]interface{}
		mockStatus  int
		mockBody    interface{}
		wantErr     bool
	}{
		{
			name: "successful preferences update",
			id:   "user-123",
			preferences: map[string]interface{}{
				"theme":    "dark",
				"language": "en",
			},
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status: "success",
				Data: &User{
					GormModel: GormModel{ID: 1},
					Email:     "user@example.com",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "PUT", r.Method)
				assert.Contains(t, r.URL.Path, "/preferences")

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

			result, err := client.Users.UpdatePreferences(context.Background(), tt.id, tt.preferences)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

// TestUsersService_ResetPassword tests the ResetPassword method
func TestUsersService_ResetPassword(t *testing.T) {
	tests := []struct {
		name       string
		email      string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:       "successful password reset request",
			email:      "user@example.com",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Password reset email sent",
			},
			wantErr: false,
		},
		{
			name:       "user not found",
			email:      "nonexistent@example.com",
			mockStatus: http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "User not found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/v1/users/reset-password", r.URL.Path)

				var body map[string]interface{}
				json.NewDecoder(r.Body).Decode(&body)
				assert.Equal(t, tt.email, body["email"])

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

			err = client.Users.ResetPassword(context.Background(), tt.email)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestUsersService_Enable tests the Enable method
func TestUsersService_Enable(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *User)
	}{
		{
			name:       "successful enable",
			id:         "user-123",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status: "success",
				Data: &User{
					GormModel: GormModel{ID: 1},
					Email:     "user@example.com",
					IsActive:  true,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, user *User) {
				assert.NotNil(t, user)
				assert.True(t, user.IsActive)
			},
		},
		{
			name:       "user not found",
			id:         "nonexistent",
			mockStatus: http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "User not found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Contains(t, r.URL.Path, "/enable")

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

			result, err := client.Users.Enable(context.Background(), tt.id)

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

// TestUsersService_Disable tests the Disable method
func TestUsersService_Disable(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *User)
	}{
		{
			name:       "successful disable",
			id:         "user-123",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status: "success",
				Data: &User{
					GormModel: GormModel{ID: 1},
					Email:     "user@example.com",
					IsActive:  false,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, user *User) {
				assert.NotNil(t, user)
				assert.False(t, user.IsActive)
			},
		},
		{
			name:       "user not found",
			id:         "nonexistent",
			mockStatus: http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "User not found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Contains(t, r.URL.Path, "/disable")

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

			result, err := client.Users.Disable(context.Background(), tt.id)

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

// TestUserJSON tests JSON marshaling and unmarshaling of User
func TestUserJSON(t *testing.T) {
	user := &User{
		GormModel:   GormModel{ID: 1},
		Email:       "test@example.com",
		FirstName:   "Test",
		LastName:    "User",
		Role:        "admin",
		Permissions: []string{"read", "write"},
		IsActive:    true,
	}

	// Marshal to JSON
	data, err := json.Marshal(user)
	require.NoError(t, err)
	assert.Contains(t, string(data), "test@example.com")
	assert.Contains(t, string(data), "admin")

	// Unmarshal from JSON
	var decoded User
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, user.Email, decoded.Email)
	assert.Equal(t, user.FirstName, decoded.FirstName)
	assert.Equal(t, user.Role, decoded.Role)
	assert.Equal(t, user.IsActive, decoded.IsActive)
}
