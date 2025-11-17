package nexmonyx

import (
	"context"
	"encoding/json"
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

// TestUsersService_NetworkErrors tests handling of network-level errors
func TestUsersService_NetworkErrors(t *testing.T) {
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
				_, _, apiErr = client.Users.List(ctx, nil)
			case "get":
				_, apiErr = client.Users.Get(ctx, "user-uuid")
			case "create":
				_, apiErr = client.Users.Create(ctx, &User{Email: "test@example.com"})
			case "update":
				_, apiErr = client.Users.Update(ctx, "user-uuid", &User{Email: "updated@example.com"})
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

// TestUsersService_ConcurrentOperations tests concurrent operations on users
func TestUsersService_ConcurrentOperations(t *testing.T) {
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
						"id":    1,
						"uuid":  "user-1",
						"email": "user1@example.com",
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
					"id":    1,
					"uuid":  "user-1",
					"email": "user1@example.com",
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
					"id":    2,
					"uuid":  "user-new",
					"email": "newuser@example.com",
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
							_, _, apiErr = client.Users.List(context.Background(), nil)
						case "get":
							_, apiErr = client.Users.Get(context.Background(), "user-1")
						case "create":
							_, apiErr = client.Users.Create(context.Background(), &User{Email: "test@example.com"})
						case "update":
							_, apiErr = client.Users.Update(context.Background(), "user-1", &User{Email: "updated@example.com"})
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
