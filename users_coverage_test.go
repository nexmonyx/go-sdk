package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Comprehensive coverage tests for users.go
// Focus on improving UpdatePermissions (77.8%), UpdatePreferences (75.0%),
// and other methods at 87.5%

// Error path tests for methods at 87.5%

func TestUsersService_Get_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/users/999", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "User not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	user, err := client.Users.Get(context.Background(), "999")
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestUsersService_GetByEmail_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/users/email/notfound@example.com", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "User not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	user, err := client.Users.GetByEmail(context.Background(), "notfound@example.com")
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestUsersService_Create_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/users", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Invalid user data",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	user, err := client.Users.Create(context.Background(), &User{
		FirstName: "Test",
		LastName:  "User",
		Email:     "invalid-email",
	})
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestUsersService_Update_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/users/999", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "User not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	user, err := client.Users.Update(context.Background(), "999", &User{
		FirstName: "Updated",
		LastName:  "Name",
	})
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestUsersService_GetCurrent_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/users/me", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Unauthorized",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	user, err := client.Users.GetCurrent(context.Background())
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestUsersService_Enable_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/users/999/enable", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "User not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	user, err := client.Users.Enable(context.Background(), "999")
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestUsersService_Disable_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/users/999/disable", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "User not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	user, err := client.Users.Disable(context.Background(), "999")
	assert.Error(t, err)
	assert.Nil(t, user)
}

// UpdateRole tests (88.9% - needs error path)

func TestUsersService_UpdateRole_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/users/1/role", r.URL.Path)

		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.NoError(t, err)
		assert.Equal(t, "admin", body["role"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id":   1,
				"role": "admin",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	user, err := client.Users.UpdateRole(context.Background(), "1", "admin")
	assert.NoError(t, err)
	assert.NotNil(t, user)
}

func TestUsersService_UpdateRole_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Invalid role",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	user, err := client.Users.UpdateRole(context.Background(), "1", "invalid_role")
	assert.Error(t, err)
	assert.Nil(t, user)
}

// UpdatePermissions tests (77.8% - needs success and error paths)

func TestUsersService_UpdatePermissions_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/users/1/permissions", r.URL.Path)

		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.NoError(t, err)
		perms, ok := body["permissions"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, perms, 3)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id":          1,
				"permissions": []string{"read:servers", "write:servers", "delete:servers"},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	user, err := client.Users.UpdatePermissions(context.Background(), "1", []string{
		"read:servers",
		"write:servers",
		"delete:servers",
	})
	assert.NoError(t, err)
	assert.NotNil(t, user)
}

func TestUsersService_UpdatePermissions_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Insufficient permissions",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	user, err := client.Users.UpdatePermissions(context.Background(), "1", []string{"admin:all"})
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestUsersService_UpdatePermissions_EmptyPermissions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)

		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.NoError(t, err)
		perms, ok := body["permissions"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, perms, 0)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id":          1,
				"permissions": []string{},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	user, err := client.Users.UpdatePermissions(context.Background(), "1", []string{})
	assert.NoError(t, err)
	assert.NotNil(t, user)
}

// UpdatePreferences tests (75.0% - needs success and error paths)

func TestUsersService_UpdatePreferences_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/users/1/preferences", r.URL.Path)

		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.NoError(t, err)
		assert.Equal(t, "dark", body["theme"])
		assert.Equal(t, "en", body["language"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id": 1,
				"preferences": map[string]interface{}{
					"theme":    "dark",
					"language": "en",
				},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	user, err := client.Users.UpdatePreferences(context.Background(), "1", map[string]interface{}{
		"theme":    "dark",
		"language": "en",
	})
	assert.NoError(t, err)
	assert.NotNil(t, user)
}

func TestUsersService_UpdatePreferences_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Invalid preferences",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	user, err := client.Users.UpdatePreferences(context.Background(), "1", map[string]interface{}{
		"invalid": "value",
	})
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestUsersService_UpdatePreferences_EmptyPreferences(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)

		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.NoError(t, err)
		assert.Len(t, body, 0)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id":          1,
				"preferences": map[string]interface{}{},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	user, err := client.Users.UpdatePreferences(context.Background(), "1", map[string]interface{}{})
	assert.NoError(t, err)
	assert.NotNil(t, user)
}

func TestUsersService_UpdatePreferences_ComplexPreferences(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)

		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.NoError(t, err)
		assert.Equal(t, "dark", body["theme"])

		notifications, ok := body["notifications"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, true, notifications["email"])
		assert.Equal(t, false, notifications["sms"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id": 1,
				"preferences": map[string]interface{}{
					"theme": "dark",
					"notifications": map[string]interface{}{
						"email": true,
						"sms":   false,
					},
				},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	user, err := client.Users.UpdatePreferences(context.Background(), "1", map[string]interface{}{
		"theme": "dark",
		"notifications": map[string]interface{}{
			"email": true,
			"sms":   false,
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, user)
}

// List tests (90.0% - needs edge case for nil options)

func TestUsersService_List_NilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/users", r.URL.Path)

		// Verify no query parameters when opts is nil
		assert.Empty(t, r.URL.Query().Get("page"))
		assert.Empty(t, r.URL.Query().Get("limit"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []*User{},
			"meta": PaginationMeta{
				Page:       1,
				Limit:      25,
				TotalItems: 0,
				TotalPages: 0,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	users, meta, err := client.Users.List(context.Background(), nil)
	assert.NoError(t, err)
	assert.NotNil(t, users)
	assert.NotNil(t, meta)
	assert.Len(t, users, 0)
}

// Additional success path tests

func TestUsersService_Enable_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/users/1/enable", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id":      1,
				"enabled": true,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	user, err := client.Users.Enable(context.Background(), "1")
	assert.NoError(t, err)
	assert.NotNil(t, user)
}

func TestUsersService_Disable_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/users/1/disable", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id":      1,
				"enabled": false,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	user, err := client.Users.Disable(context.Background(), "1")
	assert.NoError(t, err)
	assert.NotNil(t, user)
}
