package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Comprehensive coverage tests for organizations.go
// Focus on improving methods at 87.5% and 90%

// Get tests (87.5% - type assertion line)

func TestOrganizationsService_Get_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/organizations/999", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Organization not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	org, err := client.Organizations.Get(context.Background(), "999")
	assert.Error(t, err)
	assert.Nil(t, org)
}

// List tests (90% - nil options edge case)

func TestOrganizationsService_List_NilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/organizations", r.URL.Path)

		// Verify no query parameters when opts is nil
		assert.Empty(t, r.URL.Query().Get("page"))
		assert.Empty(t, r.URL.Query().Get("limit"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []*Organization{},
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
	orgs, meta, err := client.Organizations.List(context.Background(), nil)
	assert.NoError(t, err)
	assert.NotNil(t, orgs)
	assert.NotNil(t, meta)
	assert.Len(t, orgs, 0)
}

func TestOrganizationsService_List_Error(t *testing.T) {
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
	orgs, meta, err := client.Organizations.List(context.Background(), nil)
	assert.Error(t, err)
	assert.Nil(t, orgs)
	assert.Nil(t, meta)
}

// Create tests (87.5% - type assertion line)

func TestOrganizationsService_Create_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/organizations", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Invalid organization data",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	org, err := client.Organizations.Create(context.Background(), &Organization{
		Name: "Test Org",
	})
	assert.Error(t, err)
	assert.Nil(t, org)
}

// Update tests (87.5% - type assertion line)

func TestOrganizationsService_Update_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/organizations/999", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Organization not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	org, err := client.Organizations.Update(context.Background(), "999", &Organization{
		Name: "Updated Org",
	})
	assert.Error(t, err)
	assert.Nil(t, org)
}

// GetByUUID tests (87.5% - type assertion line)

func TestOrganizationsService_GetByUUID_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/organizations/uuid/org-uuid-123", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id":   1,
				"uuid": "org-uuid-123",
				"name": "Test Organization",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	org, err := client.Organizations.GetByUUID(context.Background(), "org-uuid-123")
	assert.NoError(t, err)
	assert.NotNil(t, org)
}

func TestOrganizationsService_GetByUUID_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Organization not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	org, err := client.Organizations.GetByUUID(context.Background(), "nonexistent")
	assert.Error(t, err)
	assert.Nil(t, org)
}

// UpdateSettings tests (87.5% - type assertion line)

func TestOrganizationsService_UpdateSettings_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/organizations/1/settings", r.URL.Path)

		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.NoError(t, err)
		assert.Equal(t, "dark", body["theme"])
		assert.Equal(t, "US/Pacific", body["timezone"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id": 1,
				"settings": map[string]interface{}{
					"theme":    "dark",
					"timezone": "US/Pacific",
				},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	org, err := client.Organizations.UpdateSettings(context.Background(), "1", map[string]interface{}{
		"theme":    "dark",
		"timezone": "US/Pacific",
	})
	assert.NoError(t, err)
	assert.NotNil(t, org)
}

func TestOrganizationsService_UpdateSettings_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Invalid settings",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	org, err := client.Organizations.UpdateSettings(context.Background(), "1", map[string]interface{}{
		"invalid": "setting",
	})
	assert.Error(t, err)
	assert.Nil(t, org)
}

func TestOrganizationsService_UpdateSettings_EmptySettings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.NoError(t, err)
		assert.Len(t, body, 0)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id":       1,
				"settings": map[string]interface{}{},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	org, err := client.Organizations.UpdateSettings(context.Background(), "1", map[string]interface{}{})
	assert.NoError(t, err)
	assert.NotNil(t, org)
}

// Success path tests for remaining methods

func TestOrganizationsService_Create_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id":   1,
				"name": "New Organization",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	org, err := client.Organizations.Create(context.Background(), &Organization{
		Name: "New Organization",
	})
	assert.NoError(t, err)
	assert.NotNil(t, org)
}

func TestOrganizationsService_Update_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id":   1,
				"name": "Updated Organization",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	org, err := client.Organizations.Update(context.Background(), "1", &Organization{
		Name: "Updated Organization",
	})
	assert.NoError(t, err)
	assert.NotNil(t, org)
}

// GetServers, GetUsers, GetAlerts, GetBilling nil options tests

func TestOrganizationsService_GetServers_NilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/organizations/1/servers", r.URL.Path)
		assert.Empty(t, r.URL.Query().Get("page"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []*Server{},
			"meta": PaginationMeta{},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	servers, meta, err := client.Organizations.GetServers(context.Background(), "1", nil)
	assert.NoError(t, err)
	assert.NotNil(t, servers)
	assert.NotNil(t, meta)
}

func TestOrganizationsService_GetUsers_NilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/organizations/1/users", r.URL.Path)
		assert.Empty(t, r.URL.Query().Get("page"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []*User{},
			"meta": PaginationMeta{},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	users, meta, err := client.Organizations.GetUsers(context.Background(), "1", nil)
	assert.NoError(t, err)
	assert.NotNil(t, users)
	assert.NotNil(t, meta)
}

func TestOrganizationsService_GetAlerts_NilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/organizations/1/alerts", r.URL.Path)
		assert.Empty(t, r.URL.Query().Get("page"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []*Alert{},
			"meta": PaginationMeta{},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	alerts, meta, err := client.Organizations.GetAlerts(context.Background(), "1", nil)
	assert.NoError(t, err)
	assert.NotNil(t, alerts)
	assert.NotNil(t, meta)
}

func TestOrganizationsService_GetBilling_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/organizations/1/billing", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"subscription": "premium",
				"billing_email": "billing@example.com",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	billing, err := client.Organizations.GetBilling(context.Background(), "1")
	assert.NoError(t, err)
	assert.NotNil(t, billing)
	assert.Equal(t, "premium", billing["subscription"])
}

func TestOrganizationsService_GetBilling_Error(t *testing.T) {
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
	billing, err := client.Organizations.GetBilling(context.Background(), "1")
	assert.Error(t, err)
	assert.Nil(t, billing)
}
