package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIKeysService_CreateUnified(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v2/api-keys", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"key":    map[string]interface{}{"id": 1, "name": "Test Key"},
				"secret": "secret-value",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	req := &CreateUnifiedAPIKeyRequest{
		Name:        "Test Key",
		Description: "Test Description",
		Type:        APIKeyTypeUser,
	}
	response, err := client.APIKeys.CreateUnified(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, response)
}

func TestAPIKeysService_GetUnified(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/v2/api-keys/")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   map[string]interface{}{"id": 1, "name": "Test Key"},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	key, err := client.APIKeys.GetUnified(context.Background(), "key-123")
	assert.NoError(t, err)
	assert.NotNil(t, key)
}

func TestAPIKeysService_ListUnified(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v2/api-keys", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   []map[string]interface{}{{"id": 1, "name": "Key 1"}, {"id": 2, "name": "Key 2"}},
			"meta":   map[string]interface{}{"page": 1, "limit": 25, "total": 2},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	keys, meta, err := client.APIKeys.ListUnified(context.Background(), &ListUnifiedAPIKeysOptions{
		ListOptions: ListOptions{Page: 1, Limit: 25},
		Type:        APIKeyTypeUser,
		Status:      APIKeyStatusActive,
	})
	assert.NoError(t, err)
	assert.NotNil(t, keys)
	assert.NotNil(t, meta)
}

func TestAPIKeysService_UpdateUnified(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Contains(t, r.URL.Path, "/v2/api-keys/")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   map[string]interface{}{"id": 1, "name": "Updated Key"},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	name := "Updated Key"
	req := &UpdateUnifiedAPIKeyRequest{Name: &name}
	key, err := client.APIKeys.UpdateUnified(context.Background(), "key-123", req)
	assert.NoError(t, err)
	assert.NotNil(t, key)
}

func TestAPIKeysService_DeleteUnified(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Contains(t, r.URL.Path, "/v2/api-keys/")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	err := client.APIKeys.DeleteUnified(context.Background(), "key-123")
	assert.NoError(t, err)
}

func TestAPIKeysService_RevokeUnified(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.URL.Path, "/revoke")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	err := client.APIKeys.RevokeUnified(context.Background(), "key-123")
	assert.NoError(t, err)
}

func TestAPIKeysService_RegenerateUnified(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.URL.Path, "/regenerate")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   map[string]interface{}{"key": map[string]interface{}{"id": 1}},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	response, err := client.APIKeys.RegenerateUnified(context.Background(), "key-123")
	assert.NoError(t, err)
	assert.NotNil(t, response)
}

func TestAPIKeysService_CreateForOrganization(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/v2/organizations/")
		assert.Contains(t, r.URL.Path, "/api-keys")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   map[string]interface{}{"key": map[string]interface{}{"id": 1}},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	req := &CreateUnifiedAPIKeyRequest{Name: "Org Key"}
	response, err := client.APIKeys.CreateForOrganization(context.Background(), "org-123", req)
	assert.NoError(t, err)
	assert.NotNil(t, response)
}

func TestAPIKeysService_ListForOrganization(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/v2/organizations/")
		assert.Contains(t, r.URL.Path, "/api-keys")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   []map[string]interface{}{{"id": 1, "name": "Key 1"}},
			"meta":   map[string]interface{}{"page": 1},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	keys, meta, err := client.APIKeys.ListForOrganization(context.Background(), "org-123", nil)
	assert.NoError(t, err)
	assert.NotNil(t, keys)
	assert.NotNil(t, meta)
}

func TestAPIKeysService_AdminCreateUnified(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/admin/api-keys", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   map[string]interface{}{"key": map[string]interface{}{"id": 1, "name": "Admin Key"}},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	req := &CreateUnifiedAPIKeyRequest{Name: "Admin Key"}
	response, err := client.APIKeys.AdminCreateUnified(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, response)
}

func TestAPIKeysService_AdminListUnified(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/admin/api-keys", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   []map[string]interface{}{{"id": 1, "name": "Key 1"}},
			"meta":   map[string]interface{}{"page": 1},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	keys, meta, err := client.APIKeys.AdminListUnified(context.Background(), nil)
	assert.NoError(t, err)
	assert.NotNil(t, keys)
	assert.NotNil(t, meta)
}

func TestAPIKeysService_LegacyMethods(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   map[string]interface{}{"id": 1, "name": "Legacy Key"},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	// Test Create (legacy)
	apiKey := &APIKey{Name: "Legacy Key", Description: "Test"}
	_, err := client.APIKeys.Create(context.Background(), apiKey)
	assert.NoError(t, err)

	// Test Get (legacy)
	_, err = client.APIKeys.Get(context.Background(), "key-123")
	assert.NoError(t, err)

	// Test Update (legacy)
	_, err = client.APIKeys.Update(context.Background(), "key-123", apiKey)
	assert.NoError(t, err)
}

func TestAPIKeysService_SpecializedHelpers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   map[string]interface{}{"key": map[string]interface{}{"id": 1}},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	_, err := client.APIKeys.CreateUserKey(context.Background(), "User Key", "desc", []string{"read"})
	assert.NoError(t, err)

	_, err = client.APIKeys.CreateAdminKey(context.Background(), "Admin Key", "desc", []string{"admin"}, 1)
	assert.NoError(t, err)

	_, err = client.APIKeys.CreateMonitoringAgentKey(context.Background(), "Agent Key", "desc", "ns", "agent", "us-east-1", []string{"metrics"})
	assert.NoError(t, err)

	_, err = client.APIKeys.CreateRegistrationKey(context.Background(), "Reg Key", "desc", 1)
	assert.NoError(t, err)
}

func TestAPIKeysService_Validators(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   map[string]interface{}{"id": 1, "status": "active"},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	_, err := client.APIKeys.ValidateKey(context.Background(), "key-123")
	assert.NoError(t, err)
}

func TestAPIKeysService_FilterHelpers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   []map[string]interface{}{{"id": 1, "name": "Key 1"}},
			"meta":   map[string]interface{}{"page": 1},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	_, _, err := client.APIKeys.GetKeysByType(context.Background(), APIKeyTypeUser, &ListOptions{})
	assert.NoError(t, err)

	_, _, err = client.APIKeys.GetActiveKeys(context.Background(), &ListOptions{})
	assert.NoError(t, err)

	_, _, err = client.APIKeys.GetMonitoringAgentKeys(context.Background(), "org-123", &ListOptions{})
	assert.NoError(t, err)

	_, _, err = client.APIKeys.GetRegistrationKeys(context.Background(), &ListOptions{})
	assert.NoError(t, err)
}

// Test legacy API key methods
func TestAPIKeysService_LegacyList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   []map[string]interface{}{{"id": 1, "name": "Test Key"}},
			"meta":   map[string]interface{}{"total": 1, "page": 1},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	keys, meta, err := client.APIKeys.List(context.Background(), &ListOptions{Page: 1, Limit: 10})
	assert.NoError(t, err)
	assert.NotNil(t, keys)
	assert.NotNil(t, meta)
}

func TestAPIKeysService_LegacyDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "API key deleted",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	err := client.APIKeys.Delete(context.Background(), "key-123")
	assert.NoError(t, err)
}

func TestAPIKeysService_LegacyRevoke(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.URL.Path, "/revoke")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "API key revoked",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	err := client.APIKeys.Revoke(context.Background(), "key-123")
	assert.NoError(t, err)
}

func TestAPIKeysService_LegacyRegenerate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.URL.Path, "/regenerate")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"key": map[string]interface{}{
					"id":     1,
					"key":    "new-regenerated-key",
					"secret": "new-secret",
					"name":   "Regenerated Key",
				},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	key, err := client.APIKeys.Regenerate(context.Background(), "key-123")
	assert.NoError(t, err)
	assert.NotNil(t, key)
}

func TestAPIKeysService_LegacyRegenerate_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "API key not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	key, err := client.APIKeys.Regenerate(context.Background(), "invalid-key")
	assert.Error(t, err)
	assert.Nil(t, key)
}

func TestAPIKeysService_ValidateKey_Active(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/v2/api-keys/")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   map[string]interface{}{"id": 1, "name": "Active Key", "status": "active"},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	key, err := client.APIKeys.ValidateKey(context.Background(), "key-123")
	assert.NoError(t, err)
	assert.NotNil(t, key)
}

func TestAPIKeysService_ValidateKey_Inactive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/v2/api-keys/")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   map[string]interface{}{"id": 1, "name": "Inactive Key", "status": "revoked"},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	key, err := client.APIKeys.ValidateKey(context.Background(), "key-123")
	assert.Error(t, err)
	assert.Nil(t, key)
	assert.Contains(t, err.Error(), "not active")
}

func TestAPIKeysService_ValidateKey_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "API key not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	key, err := client.APIKeys.ValidateKey(context.Background(), "invalid-key")
	assert.Error(t, err)
	assert.Nil(t, key)
}

func TestAPIKeysService_ListForOrganization_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/v2/organizations/")
		assert.Contains(t, r.URL.Path, "/api-keys")

		// Verify query parameters
		query := r.URL.Query()
		assert.Equal(t, "user", query.Get("type"))
		assert.Equal(t, "active", query.Get("status"))
		assert.Equal(t, "1", query.Get("page"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   []map[string]interface{}{{"id": 1, "name": "Org Key"}},
			"meta":   map[string]interface{}{"page": 1, "limit": 25, "total": 1},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	opts := &ListUnifiedAPIKeysOptions{
		ListOptions: ListOptions{Page: 1, Limit: 25},
		Type:        APIKeyTypeUser,
		Status:      APIKeyStatusActive,
	}
	keys, meta, err := client.APIKeys.ListForOrganization(context.Background(), "org-123", opts)
	assert.NoError(t, err)
	assert.NotNil(t, keys)
	assert.NotNil(t, meta)
}

func TestAPIKeysService_ListForOrganization_AllFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)

		// Verify all query parameters are set
		query := r.URL.Query()
		assert.Equal(t, "monitoring_agent", query.Get("type"))
		assert.Equal(t, "active", query.Get("status"))
		assert.Equal(t, "100", query.Get("user_id"))
		assert.Equal(t, "probe", query.Get("agent_type"))
		assert.Equal(t, "us-east-1", query.Get("region_code"))
		assert.Equal(t, "production", query.Get("namespace"))
		assert.Equal(t, "metrics:read", query.Get("capability"))
		assert.Equal(t, "critical", query.Get("tag"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   []map[string]interface{}{},
			"meta":   map[string]interface{}{"page": 1, "limit": 25, "total": 0},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	opts := &ListUnifiedAPIKeysOptions{
		ListOptions: ListOptions{Page: 1, Limit: 25},
		Type:        APIKeyTypeMonitoringAgent,
		Status:      APIKeyStatusActive,
		UserID:      100,
		AgentType:   "probe",
		RegionCode:  "us-east-1",
		Namespace:   "production",
		Capability:  "metrics:read",
		Tag:         "critical",
	}
	keys, meta, err := client.APIKeys.ListForOrganization(context.Background(), "org-123", opts)
	assert.NoError(t, err)
	assert.NotNil(t, keys)
	assert.NotNil(t, meta)
}

func TestAPIKeysService_ListForOrganization_NilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   []map[string]interface{}{{"id": 1}},
			"meta":   map[string]interface{}{"page": 1},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	keys, meta, err := client.APIKeys.ListForOrganization(context.Background(), "org-123", nil)
	assert.NoError(t, err)
	assert.NotNil(t, keys)
	assert.NotNil(t, meta)
}

func TestAPIKeysService_ListForOrganization_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Access denied",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	keys, meta, err := client.APIKeys.ListForOrganization(context.Background(), "org-123", nil)
	assert.Error(t, err)
	assert.Nil(t, keys)
	assert.Nil(t, meta)
}

func TestAPIKeysService_AdminListUnified_WithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v2/admin/api-keys", r.URL.Path)

		// Verify query parameters
		query := r.URL.Query()
		assert.Equal(t, "admin", query.Get("type"))
		assert.Equal(t, "active", query.Get("status"))
		assert.Equal(t, "50", query.Get("user_id"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   []map[string]interface{}{{"id": 1, "name": "Admin Key"}},
			"meta":   map[string]interface{}{"page": 1, "limit": 25, "total": 1},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	opts := &ListUnifiedAPIKeysOptions{
		ListOptions: ListOptions{Page: 1, Limit: 25},
		Type:        APIKeyTypeAdmin,
		Status:      APIKeyStatusActive,
		UserID:      50,
	}
	keys, meta, err := client.APIKeys.AdminListUnified(context.Background(), opts)
	assert.NoError(t, err)
	assert.NotNil(t, keys)
	assert.NotNil(t, meta)
	assert.Len(t, keys, 1)
}

func TestAPIKeysService_AdminListUnified_AllFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)

		// Verify all filters are properly set
		query := r.URL.Query()
		assert.Equal(t, "registration", query.Get("type"))
		assert.Equal(t, "revoked", query.Get("status"))
		assert.Equal(t, "200", query.Get("user_id"))
		assert.Equal(t, "controller", query.Get("agent_type"))
		assert.Equal(t, "eu-west-1", query.Get("region_code"))
		assert.Equal(t, "staging", query.Get("namespace"))
		assert.Equal(t, "admin:write", query.Get("capability"))
		assert.Equal(t, "deprecated", query.Get("tag"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   []map[string]interface{}{},
			"meta":   map[string]interface{}{"page": 1, "limit": 25, "total": 0},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	opts := &ListUnifiedAPIKeysOptions{
		ListOptions: ListOptions{Page: 1, Limit: 25},
		Type:        APIKeyTypeRegistration,
		Status:      APIKeyStatusRevoked,
		UserID:      200,
		AgentType:   "controller",
		RegionCode:  "eu-west-1",
		Namespace:   "staging",
		Capability:  "admin:write",
		Tag:         "deprecated",
	}
	keys, meta, err := client.APIKeys.AdminListUnified(context.Background(), opts)
	assert.NoError(t, err)
	assert.NotNil(t, keys)
	assert.NotNil(t, meta)
}

func TestAPIKeysService_AdminListUnified_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Unauthorized",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	keys, meta, err := client.APIKeys.AdminListUnified(context.Background(), nil)
	assert.Error(t, err)
	assert.Nil(t, keys)
	assert.Nil(t, meta)
}
