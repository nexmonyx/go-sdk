package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUsersService_GetComprehensive tests the Get method with various scenarios
func TestUsersService_GetComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *User)
	}{
		{
			name:       "success - full user data",
			userID:     "user-123",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":    123,
					"email": "user@example.com",
					"first_name": "John",
					"last_name": "Doe",
					"display_name": "John Doe",
					"role": "admin",
					"created_at": "2024-01-01T00:00:00Z",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, user *User) {
				assert.Equal(t, uint(123), user.ID)
				assert.Equal(t, "user@example.com", user.Email)
				assert.Equal(t, "John", user.FirstName)
				assert.Equal(t, "Doe", user.LastName)
			},
		},
		{
			name:       "success - minimal user data",
			userID:     "user-456",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":    456,
					"email": "minimal@example.com",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, user *User) {
				assert.Equal(t, uint(456), user.ID)
				assert.Equal(t, "minimal@example.com", user.Email)
			},
		},
		{
			name:       "not found",
			userID:     "user-999",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "User not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			userID:     "user-123",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "forbidden",
			userID:     "user-123",
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Access denied",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			userID:     "user-123",
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
				assert.Contains(t, r.URL.Path, tt.userID)
				assert.Contains(t, r.URL.Path, "/v1/users/")

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

			result, err := client.Users.Get(ctx, tt.userID)

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

// TestUsersService_GetByEmailComprehensive tests the GetByEmail method
func TestUsersService_GetByEmailComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		email      string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *User)
	}{
		{
			name:       "success - found by email",
			email:      "user@example.com",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":    123,
					"email": "user@example.com",
					"first_name": "John",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, user *User) {
				assert.Equal(t, "user@example.com", user.Email)
			},
		},
		{
			name:       "not found",
			email:      "notfound@example.com",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "User not found",
			},
			wantErr: true,
		},
		{
			name:       "invalid email format",
			email:      "invalid-email",
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Invalid email format",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			email:      "user@example.com",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			email:      "user@example.com",
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
				assert.Contains(t, r.URL.Path, "/v1/users/email/")

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

			result, err := client.Users.GetByEmail(ctx, tt.email)

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

// Test continues with List, Create, Update, Delete, GetCurrent, UpdateRole, UpdatePermissions, UpdatePreferences, ResetPassword, Enable, Disable

// TestUsersService_ListComprehensive tests the List method
func TestUsersService_ListComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		opts       *ListOptions
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, []*User, *PaginationMeta)
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
						"id":    1,
						"email": "user1@example.com",
					},
					{
						"id":    2,
						"email": "user2@example.com",
					},
				},
				"meta": map[string]interface{}{
					"page":        1,
					"limit":       10,
					"total_items": 2,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, users []*User, meta *PaginationMeta) {
				assert.Len(t, users, 2)
				assert.NotNil(t, meta)
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
				assert.Contains(t, r.URL.Path, "/v1/users")

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

			users, meta, err := client.Users.List(ctx, tt.opts)

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

// TestUsersService_CreateComprehensive tests the Create method
func TestUsersService_CreateComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		user       *User
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *User)
	}{
		{
			name: "success - full user",
			user: &User{
				Email:     "newuser@example.com",
				FirstName: "New",
				LastName:  "User",
			},
			mockStatus: http.StatusCreated,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":         100,
					"email":      "newuser@example.com",
					"first_name": "New",
					"last_name":  "User",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, user *User) {
				assert.Equal(t, "newuser@example.com", user.Email)
			},
		},
		{
			name: "validation error - missing email",
			user: &User{
				FirstName: "No",
				LastName:  "Email",
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Email is required",
			},
			wantErr: true,
		},
		{
			name: "validation error - duplicate email",
			user: &User{
				Email: "existing@example.com",
			},
			mockStatus: http.StatusConflict,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "User with this email already exists",
			},
			wantErr: true,
		},
		{
			name: "unauthorized",
			user: &User{
				Email: "test@example.com",
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
			user: &User{
				Email: "test@example.com",
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
			user: &User{
				Email: "test@example.com",
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
				assert.Contains(t, r.URL.Path, "/v1/users")

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

			result, err := client.Users.Create(ctx, tt.user)

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

// TestUsersService_UpdateComprehensive tests the Update method
func TestUsersService_UpdateComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		user       *User
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:   "success - full update",
			userID: "user-123",
			user: &User{
				FirstName: "Updated",
				LastName:  "Name",
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":         123,
					"first_name": "Updated",
					"last_name":  "Name",
				},
			},
			wantErr: false,
		},
		{
			name:   "not found",
			userID: "user-999",
			user: &User{
				FirstName: "Test",
			},
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "User not found",
			},
			wantErr: true,
		},
		{
			name:   "unauthorized",
			userID: "user-123",
			user: &User{
				FirstName: "Test",
			},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:   "forbidden",
			userID: "user-123",
			user: &User{
				FirstName: "Test",
			},
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Access denied",
			},
			wantErr: true,
		},
		{
			name:   "server error",
			userID: "user-123",
			user: &User{
				FirstName: "Test",
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
				assert.Contains(t, r.URL.Path, tt.userID)

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

			_, err = client.Users.Update(ctx, tt.userID, tt.user)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Continue with remaining functions...

// TestUsersService_DeleteComprehensive tests the Delete method
func TestUsersService_DeleteComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:       "success",
			userID:     "user-123",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status":  "success",
				"message": "User deleted",
			},
			wantErr: false,
		},
		{
			name:       "not found",
			userID:     "user-999",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "User not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			userID:     "user-123",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "forbidden",
			userID:     "user-123",
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Insufficient permissions",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			userID:     "user-123",
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
				assert.Contains(t, r.URL.Path, tt.userID)

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

			err = client.Users.Delete(ctx, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestUsersService_GetCurrentComprehensive tests the GetCurrent method
func TestUsersService_GetCurrentComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *User)
	}{
		{
			name:       "success - full current user",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":    123,
					"email": "currentuser@example.com",
					"first_name": "Current",
					"last_name": "User",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, user *User) {
				assert.Equal(t, "currentuser@example.com", user.Email)
			},
		},
		{
			name:       "unauthorized",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "server error",
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
				assert.Contains(t, r.URL.Path, "/v1/users/me")

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

			result, err := client.Users.GetCurrent(ctx)

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

// TestUsersService_UpdateRoleComprehensive tests the UpdateRole method
func TestUsersService_UpdateRoleComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		role       string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:       "success - update to admin",
			userID:     "user-123",
			role:       "admin",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":   123,
					"role": "admin",
				},
			},
			wantErr: false,
		},
		{
			name:       "validation error - invalid role",
			userID:     "user-123",
			role:       "invalid-role",
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Invalid role",
			},
			wantErr: true,
		},
		{
			name:       "not found",
			userID:     "user-999",
			role:       "admin",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "User not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			userID:     "user-123",
			role:       "admin",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "forbidden",
			userID:     "user-123",
			role:       "admin",
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Insufficient permissions",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			userID:     "user-123",
			role:       "admin",
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
				assert.Contains(t, r.URL.Path, tt.userID)
				assert.Contains(t, r.URL.Path, "/role")

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

			_, err = client.Users.UpdateRole(ctx, tt.userID, tt.role)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestUsersService_UpdatePermissionsComprehensive tests the UpdatePermissions method
func TestUsersService_UpdatePermissionsComprehensive(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		permissions []string
		mockStatus  int
		mockBody    interface{}
		wantErr     bool
	}{
		{
			name:        "success - update permissions",
			userID:      "user-123",
			permissions: []string{"read", "write", "delete"},
			mockStatus:  http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":          123,
					"permissions": []string{"read", "write", "delete"},
				},
			},
			wantErr: false,
		},
		{
			name:        "success - empty permissions",
			userID:      "user-123",
			permissions: []string{},
			mockStatus:  http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":          123,
					"permissions": []string{},
				},
			},
			wantErr: false,
		},
		{
			name:        "not found",
			userID:      "user-999",
			permissions: []string{"read"},
			mockStatus:  http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "User not found",
			},
			wantErr: true,
		},
		{
			name:        "unauthorized",
			userID:      "user-123",
			permissions: []string{"read"},
			mockStatus:  http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:        "forbidden",
			userID:      "user-123",
			permissions: []string{"read"},
			mockStatus:  http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Insufficient permissions",
			},
			wantErr: true,
		},
		{
			name:        "server error",
			userID:      "user-123",
			permissions: []string{"read"},
			mockStatus:  http.StatusInternalServerError,
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
				assert.Contains(t, r.URL.Path, tt.userID)
				assert.Contains(t, r.URL.Path, "/permissions")

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

			_, err = client.Users.UpdatePermissions(ctx, tt.userID, tt.permissions)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestUsersService_UpdatePreferencesComprehensive tests the UpdatePreferences method
func TestUsersService_UpdatePreferencesComprehensive(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		preferences map[string]interface{}
		mockStatus  int
		mockBody    interface{}
		wantErr     bool
	}{
		{
			name:   "success - update preferences",
			userID: "user-123",
			preferences: map[string]interface{}{
				"theme":    "dark",
				"language": "en",
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id": 123,
					"preferences": map[string]interface{}{
						"theme":    "dark",
						"language": "en",
					},
				},
			},
			wantErr: false,
		},
		{
			name:        "success - empty preferences",
			userID:      "user-123",
			preferences: map[string]interface{}{},
			mockStatus:  http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":          123,
					"preferences": map[string]interface{}{},
				},
			},
			wantErr: false,
		},
		{
			name:   "not found",
			userID: "user-999",
			preferences: map[string]interface{}{
				"theme": "dark",
			},
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "User not found",
			},
			wantErr: true,
		},
		{
			name:   "unauthorized",
			userID: "user-123",
			preferences: map[string]interface{}{
				"theme": "dark",
			},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:   "forbidden",
			userID: "user-123",
			preferences: map[string]interface{}{
				"theme": "dark",
			},
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Access denied",
			},
			wantErr: true,
		},
		{
			name:   "server error",
			userID: "user-123",
			preferences: map[string]interface{}{
				"theme": "dark",
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
				assert.Contains(t, r.URL.Path, tt.userID)
				assert.Contains(t, r.URL.Path, "/preferences")

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

			_, err = client.Users.UpdatePreferences(ctx, tt.userID, tt.preferences)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestUsersService_ResetPasswordComprehensive tests the ResetPassword method
func TestUsersService_ResetPasswordComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		email      string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:       "success - password reset sent",
			email:      "user@example.com",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status":  "success",
				"message": "Password reset email sent",
			},
			wantErr: false,
		},
		{
			name:       "not found - email doesn't exist",
			email:      "notfound@example.com",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "User not found",
			},
			wantErr: true,
		},
		{
			name:       "validation error - invalid email",
			email:      "invalid-email",
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Invalid email format",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			email:      "user@example.com",
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Failed to send reset email",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Contains(t, r.URL.Path, "/reset-password")

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

			err = client.Users.ResetPassword(ctx, tt.email)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestUsersService_EnableComprehensive tests the Enable method
func TestUsersService_EnableComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:       "success - user enabled",
			userID:     "user-123",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":      123,
					"enabled": true,
				},
			},
			wantErr: false,
		},
		{
			name:       "not found",
			userID:     "user-999",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "User not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			userID:     "user-123",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "forbidden",
			userID:     "user-123",
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Insufficient permissions",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			userID:     "user-123",
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
				assert.Contains(t, r.URL.Path, tt.userID)
				assert.Contains(t, r.URL.Path, "/enable")

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

			_, err = client.Users.Enable(ctx, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestUsersService_DisableComprehensive tests the Disable method
func TestUsersService_DisableComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:       "success - user disabled",
			userID:     "user-123",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":      123,
					"enabled": false,
				},
			},
			wantErr: false,
		},
		{
			name:       "not found",
			userID:     "user-999",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "User not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			userID:     "user-123",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "forbidden",
			userID:     "user-123",
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Insufficient permissions",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			userID:     "user-123",
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
				assert.Contains(t, r.URL.Path, tt.userID)
				assert.Contains(t, r.URL.Path, "/disable")

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

			_, err = client.Users.Disable(ctx, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
