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

// ============================================================================
// Tag Management Tests
// ============================================================================

func TestTagsService_ListTags(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/tags", r.URL.Path)
		assert.Equal(t, "env", r.URL.Query().Get("namespace"))
		assert.Equal(t, "manual", r.URL.Query().Get("source"))
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		tags := []Tag{
			{
				ID:             1,
				OrganizationID: 100,
				Namespace:      "env",
				Key:            "environment",
				Value:          "production",
				Source:         "manual",
				CreatedAt:      CustomTime{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
				UpdatedAt:      CustomTime{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
			},
		}
		response := StandardResponse{
			Status:  "success",
			Message: "Tags retrieved successfully",
			Data:    tags,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	opts := &TagListOptions{
		Namespace: "env",
		Source:    "manual",
		Page:      1,
		Limit:     50,
	}
	tags, meta, err := client.Tags.List(context.Background(), opts)
	require.NoError(t, err)
	assert.Len(t, tags, 1)
	assert.Equal(t, "env", tags[0].Namespace)
	assert.Equal(t, "environment", tags[0].Key)
	_ = meta // Pagination metadata
}

func TestTagsService_CreateTag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/tags", r.URL.Path)

		var req TagCreateRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "env", req.Namespace)
		assert.Equal(t, "environment", req.Key)

		response := StandardResponse{
			Status:  "success",
			Message: "Tag created successfully",
			Data: &Tag{
				ID:             1,
				OrganizationID: 100,
				Namespace:      req.Namespace,
				Key:            req.Key,
				Value:          req.Value,
				CreatedAt:      CustomTime{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
				UpdatedAt:      CustomTime{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	req := &TagCreateRequest{
		Namespace: "env",
		Key:       "environment",
		Value:     "staging",
	}
	tag, err := client.Tags.Create(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "env", tag.Namespace)
	assert.Equal(t, "environment", tag.Key)
}

func TestTagsService_GetServerTags(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/server/server-uuid-123/tags", r.URL.Path)

		tags := []ServerTag{
			{
				ID:         1,
				TagID:      10,
				Namespace:  "env",
				Key:        "environment",
				Value:      "production",
				Source:     "manual",
				AssignedAt: CustomTime{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
			},
		}
		response := StandardResponse{
			Status:  "success",
			Message: "Server tags retrieved successfully",
			Data:    tags,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	tags, err := client.Tags.GetServerTags(context.Background(), "server-uuid-123")
	require.NoError(t, err)
	assert.Len(t, tags, 1)
	assert.Equal(t, uint(10), tags[0].TagID)
}

func TestTagsService_AssignTagsToServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/servers/server-uuid-456/tags", r.URL.Path)

		var req TagAssignRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Len(t, req.TagIDs, 2)

		response := StandardResponse{
			Status:  "success",
			Message: "Tags assigned successfully",
			Data: &TagAssignmentResult{
				Assigned:        2,
				AlreadyAssigned: 0,
				Total:           2,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	req := &TagAssignRequest{TagIDs: []uint{1, 2}}
	result, err := client.Tags.AssignTagsToServer(context.Background(), "server-uuid-456", req)
	require.NoError(t, err)
	assert.Equal(t, 2, result.Assigned)
	assert.Equal(t, 2, result.Total)
}

func TestTagsService_RemoveTagFromServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/servers/server-uuid-789/tags/456", r.URL.Path)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	err = client.Tags.RemoveTagFromServer(context.Background(), "server-uuid-789", 456)
	require.NoError(t, err)
}

func TestTagsService_GetServersByTag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/tags/123/servers", r.URL.Path)

		data := GetServersByTagResponse{
			TagID:       123,
			TagKey:      "environment",
			TagValue:    "production",
			Namespace:   "infra",
			ServerCount: 2,
			Servers: []TagServerInfo{
				{
					ID:         1,
					ServerUUID: "server-uuid-001",
					Name:       "web-server-1",
					Hostname:   "web-1.example.com",
					MainIP:     "192.168.1.10",
					Os:         "Ubuntu",
					OsVersion:  "22.04",
					Status:     "online",
					AssignedAt: CustomTime{time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)},
					Source:     "manual",
				},
				{
					ID:         2,
					ServerUUID: "server-uuid-002",
					Name:       "web-server-2",
					Hostname:   "web-2.example.com",
					MainIP:     "192.168.1.11",
					Os:         "Ubuntu",
					OsVersion:  "22.04",
					Status:     "online",
					AssignedAt: CustomTime{time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC)},
					Source:     "automatic",
				},
			},
		}
		response := StandardResponse{
			Status:  "success",
			Message: "Servers fetched successfully",
			Data:    data,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	result, err := client.Tags.GetServersByTag(context.Background(), 123)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, uint(123), result.TagID)
	assert.Equal(t, "environment", result.TagKey)
	assert.Equal(t, "production", result.TagValue)
	assert.Equal(t, "infra", result.Namespace)
	assert.Equal(t, 2, result.ServerCount)
	assert.Len(t, result.Servers, 2)

	// Verify first server
	assert.Equal(t, uint(1), result.Servers[0].ID)
	assert.Equal(t, "server-uuid-001", result.Servers[0].ServerUUID)
	assert.Equal(t, "web-server-1", result.Servers[0].Name)
	assert.Equal(t, "online", result.Servers[0].Status)
	assert.Equal(t, "manual", result.Servers[0].Source)

	// Verify second server
	assert.Equal(t, uint(2), result.Servers[1].ID)
	assert.Equal(t, "server-uuid-002", result.Servers[1].ServerUUID)
	assert.Equal(t, "automatic", result.Servers[1].Source)
}

func TestTagsService_GetServersByTag_EmptyResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/tags/456/servers", r.URL.Path)

		data := GetServersByTagResponse{
			TagID:       456,
			TagKey:      "environment",
			TagValue:    "staging",
			Namespace:   "infra",
			ServerCount: 0,
			Servers:     []TagServerInfo{},
		}
		response := StandardResponse{
			Status:  "success",
			Message: "Servers fetched successfully",
			Data:    data,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	result, err := client.Tags.GetServersByTag(context.Background(), 456)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, uint(456), result.TagID)
	assert.Equal(t, 0, result.ServerCount)
	assert.Len(t, result.Servers, 0)
}

// ============================================================================
// Namespace Tests
// ============================================================================

func TestTagsService_CreateNamespace(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/tag-namespaces", r.URL.Path)

		var req TagNamespaceCreateRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "env", req.Namespace)

		response := StandardResponse{
			Status:  "success",
			Message: "Namespace created successfully",
			Data: &TagNamespace{
				ID:          1,
				Namespace:   req.Namespace,
				Description: req.Description,
				Type:        req.Type,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	req := &TagNamespaceCreateRequest{
		Namespace:   "env",
		Description: "Environment tags",
		Type:        "system",
	}
	ns, err := client.Tags.CreateNamespace(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "env", ns.Namespace)
}

func TestTagsService_ListNamespaces(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/tag-namespaces", r.URL.Path)

		namespaces := []TagNamespace{
			{
				ID:          1,
				Namespace:   "env",
				Description: "Environment tags",
				Type:        "system",
				CreatedAt:   CustomTime{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
				UpdatedAt:   CustomTime{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
			},
		}
		response := StandardResponse{
			Status:  "success",
			Message: "Namespaces retrieved successfully",
			Data: struct {
				Namespaces []TagNamespace `json:"namespaces"`
				Total      int            `json:"total"`
			}{
				Namespaces: namespaces,
				Total:      len(namespaces),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	namespaces, total, err := client.Tags.ListNamespaces(context.Background(), nil)
	require.NoError(t, err)
	assert.Len(t, namespaces, 1)
	assert.Equal(t, 1, total)
}

func TestTagsService_SetNamespacePermissions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/tag-namespaces/env/permissions", r.URL.Path)

		var req TagNamespacePermissionRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.True(t, req.CanCreate)
		assert.True(t, req.CanRead)

		response := StandardResponse{
			Status:  "success",
			Message: "Permissions updated successfully",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	req := &TagNamespacePermissionRequest{
		RoleName:  "admin",
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: false,
	}
	err = client.Tags.SetNamespacePermissions(context.Background(), "env", req)
	require.NoError(t, err)
}

// ============================================================================
// Inheritance Tests
// ============================================================================

func TestTagsService_CreateInheritanceRule(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/tag-inheritance/rules", r.URL.Path)

		var req TagInheritanceRuleCreateRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "Auto-tag rule", req.Name)

		response := StandardResponse{
			Status:  "success",
			Message: "Inheritance rule created successfully",
			Data: &TagInheritanceRule{
				ID:             1,
				OrganizationID: 100,
				Name:           req.Name,
				SourceType:     req.SourceType,
				TargetType:     req.TargetType,
				CreatedAt:      CustomTime{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
				UpdatedAt:      CustomTime{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	req := &TagInheritanceRuleCreateRequest{
		Name:       "Auto-tag rule",
		SourceType: "organization",
		TargetType: "all_servers",
	}
	rule, err := client.Tags.CreateInheritanceRule(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "Auto-tag rule", rule.Name)
}

func TestTagsService_SetOrganizationTag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/tag-inheritance/organization-tags", r.URL.Path)

		var req OrganizationTagRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, uint(10), req.TagID)

		response := StandardResponse{
			Status:  "success",
			Message: "Organization tag set successfully",
			Data: &OrganizationTag{
				ID:             1,
				OrganizationID: 100,
				InheritToAll:   req.InheritToAll,
				CreatedAt:      CustomTime{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
				UpdatedAt:      CustomTime{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	req := &OrganizationTagRequest{
		TagID:        10,
		InheritToAll: true,
	}
	orgTag, err := client.Tags.SetOrganizationTag(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, orgTag.InheritToAll)
}

func TestTagsService_ListOrganizationTags(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/tag-inheritance/organization-tags", r.URL.Path)

		orgTags := []OrganizationTag{
			{
				ID:             1,
				OrganizationID: 100,
				InheritToAll:   true,
				CreatedAt:      CustomTime{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
				UpdatedAt:      CustomTime{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
			},
		}
		response := StandardResponse{
			Status:  "success",
			Message: "Organization tags retrieved successfully",
			Data: struct {
				Tags  []OrganizationTag `json:"tags"`
				Total int               `json:"total"`
			}{
				Tags:  orgTags,
				Total: len(orgTags),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	orgTags, total, err := client.Tags.ListOrganizationTags(context.Background(), nil)
	require.NoError(t, err)
	assert.Len(t, orgTags, 1)
	assert.Equal(t, 1, total)
}

func TestTagsService_RemoveOrganizationTag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/tag-inheritance/organization-tags/123", r.URL.Path)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	err = client.Tags.RemoveOrganizationTag(context.Background(), 123)
	require.NoError(t, err)
}

func TestTagsService_CreateServerRelationship(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/tag-inheritance/server-relationships", r.URL.Path)

		var req ServerRelationshipRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "parent-uuid-100", req.ParentServerID)

		response := StandardResponse{
			Status:  "success",
			Message: "Server relationship created successfully",
			Data: &ServerParentRelationship{
				ID:             1,
				OrganizationID: 100,
				RelationType:   req.RelationType,
				InheritTags:    req.InheritTags,
				CreatedAt:      CustomTime{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
				UpdatedAt:      CustomTime{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	req := &ServerRelationshipRequest{
		ParentServerID: "parent-uuid-100",
		ChildServerID:  "child-uuid-200",
		RelationType:   "vm_host",
		InheritTags:    true,
	}
	rel, err := client.Tags.CreateServerRelationship(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "vm_host", rel.RelationType)
}

func TestTagsService_ListServerRelationships(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/tag-inheritance/server-relationships", r.URL.Path)

		relationships := []ServerParentRelationship{
			{
				ID:             1,
				OrganizationID: 100,
				RelationType:   "vm_host",
				InheritTags:    true,
				CreatedAt:      CustomTime{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
				UpdatedAt:      CustomTime{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
			},
		}
		response := StandardResponse{
			Status:  "success",
			Message: "Server relationships retrieved successfully",
			Data: struct {
				Relationships []ServerParentRelationship `json:"relationships"`
				Total         int                        `json:"total"`
			}{
				Relationships: relationships,
				Total:         len(relationships),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	rels, total, err := client.Tags.ListServerRelationships(context.Background(), nil)
	require.NoError(t, err)
	assert.Len(t, rels, 1)
	assert.Equal(t, 1, total)
}

func TestTagsService_DeleteServerRelationship(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/tag-inheritance/server-relationships/123", r.URL.Path)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	err = client.Tags.DeleteServerRelationship(context.Background(), 123)
	require.NoError(t, err)
}

// ============================================================================
// History Tests
// ============================================================================

func TestTagsService_GetTagHistory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/servers/123/tags/history", r.URL.Path)
		assert.Equal(t, "added", r.URL.Query().Get("action"))

		envNamespace := "env"
		history := []*TagHistoryResponse{
			{
				ID:     "1",
				Action: "added",
				Tag: TagHistoryTag{
					ID:        10,
					Namespace: &envNamespace,
					Key:       "environment",
					Value:     "production",
				},
				Timestamp: "2025-01-01T00:00:00Z",
			},
		}
		response := struct {
			Data       []*TagHistoryResponse `json:"data"`
			Status     string                `json:"status"`
			Message    string                `json:"message"`
			Pagination *PaginationMeta       `json:"pagination"`
		}{
			Data:    history,
			Status:  "success",
			Message: "Tag history retrieved successfully",
			Pagination: &PaginationMeta{
				Page:       1,
				Limit:      50,
				TotalItems: 1,
				TotalPages: 1,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	opts := &TagHistoryQueryParams{
		Action: "added",
		Page:   1,
		Limit:  50,
	}
	history, meta, err := client.Tags.GetTagHistory(context.Background(), 123, opts)
	require.NoError(t, err)
	assert.Len(t, history, 1)
	assert.Equal(t, "added", history[0].Action)
	assert.NotNil(t, meta)
	assert.Equal(t, 1, meta.TotalItems)
}

func TestTagsService_GetTagHistorySummary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/servers/123/tags/history/summary", r.URL.Path)

		response := StandardResponse{
			Status:  "success",
			Message: "Tag history summary retrieved successfully",
			Data: &TagHistorySummary{
				TotalChanges: 10,
				ChangesByAction: map[string]int{
					"added":   5,
					"removed": 3,
					"updated": 2,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	summary, err := client.Tags.GetTagHistorySummary(context.Background(), 123)
	require.NoError(t, err)
	assert.Equal(t, 10, summary.TotalChanges)
	assert.Equal(t, 5, summary.ChangesByAction["added"])
}

// ============================================================================
// Bulk Operations Tests
// ============================================================================

func TestTagsService_BulkCreateTags(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/bulk/tags", r.URL.Path)

		var req BulkTagCreateRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Len(t, req.Tags, 2)

		response := StandardResponse{
			Status:  "success",
			Message: "Bulk tag creation completed",
			Data: &BulkTagCreateResult{
				CreatedCount: 2,
				SkippedCount: 0,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	req := &BulkTagCreateRequest{
		Tags: []BulkTagCreateItem{
			{Namespace: "env", Key: "environment", Value: "prod"},
			{Namespace: "env", Key: "region", Value: "us-east-1"},
		},
	}
	result, err := client.Tags.BulkCreateTags(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 2, result.CreatedCount)
}

func TestTagsService_BulkAssignTags(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/bulk/tags/assign", r.URL.Path)

		var req BulkTagAssignRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Len(t, req.ServerIDs, 3)
		assert.Len(t, req.TagIDs, 2)

		response := StandardResponse{
			Status:  "success",
			Message: "Bulk tag assignment completed",
			Data: &BulkTagAssignResult{
				Assigned: 6,
				Skipped:  0,
				Total:    6,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	req := &BulkTagAssignRequest{
		ServerIDs: []string{"server1", "server2", "server3"},
		TagIDs:    []uint{1, 2},
	}
	result, err := client.Tags.BulkAssignTags(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 6, result.Assigned)
}

func TestTagsService_AssignTagsToGroups(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/bulk/groups/assign", r.URL.Path)

		var req BulkGroupAssignRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Len(t, req.ServerIDs, 5)
		assert.Len(t, req.GroupIDs, 2)

		response := StandardResponse{
			Status:  "success",
			Message: "Servers assigned to groups successfully",
			Data: &BulkGroupAssignResult{
				Assigned: 10,
				Skipped:  0,
				Total:    10,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	req := &BulkGroupAssignRequest{
		ServerIDs: []string{"s1", "s2", "s3", "s4", "s5"},
		GroupIDs:  []uint{1, 2},
	}
	result, err := client.Tags.AssignTagsToGroups(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 10, result.Assigned)
}

// ============================================================================
// Rule Detection Tests
// ============================================================================

func TestTagsService_ListTagDetectionRules(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/tag-rules", r.URL.Path)
		assert.Equal(t, "true", r.URL.Query().Get("enabled"))

		rules := []*TagDetectionRule{
			{
				ID:             1,
				OrganizationID: 100,
				Name:           "Auto-detect production",
				Namespace:      "env",
				TagKey:         "environment",
				TagValue:       "production",
				Enabled:        true,
			},
		}
		response := struct {
			Data       []*TagDetectionRule `json:"data"`
			Status     string              `json:"status"`
			Message    string              `json:"message"`
			Pagination *PaginationMeta     `json:"pagination"`
		}{
			Data:    rules,
			Status:  "success",
			Message: "Tag detection rules retrieved successfully",
			Pagination: &PaginationMeta{
				Page:       1,
				Limit:      50,
				TotalItems: 1,
				TotalPages: 1,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	enabled := true
	opts := &TagDetectionRuleListOptions{
		Enabled: &enabled,
		Page:    1,
		Limit:   50,
	}
	rules, total, err := client.Tags.ListTagDetectionRules(context.Background(), opts)
	require.NoError(t, err)
	assert.Len(t, rules, 1)
	assert.True(t, rules[0].Enabled)
	assert.Equal(t, 1, total)
}

func TestTagsService_CreateDefaultRules(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/tag-rules/defaults", r.URL.Path)

		response := StandardResponse{
			Status:  "success",
			Message: "Default rules created successfully",
			Data: &DefaultRulesCreateResult{
				CreatedCount: 5,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	result, err := client.Tags.CreateDefaultRules(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 5, result.CreatedCount)
}

func TestTagsService_EvaluateRules(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/tag-rules/evaluate", r.URL.Path)

		var req EvaluateRulesRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Len(t, req.ServerIDs, 3)

		response := StandardResponse{
			Status:  "success",
			Message: "Rule evaluation queued",
			Data: &EvaluateRulesResult{
				ProcessingCount: 3,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	req := &EvaluateRulesRequest{
		ServerIDs: []string{"server1", "server2", "server3"},
	}
	result, err := client.Tags.EvaluateRules(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 3, result.ProcessingCount)
}

// ============================================================================
// Error Handling Tests
// ============================================================================

func TestTagsService_ErrorHandling(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		responseBody string
		shouldError  bool
	}{
		{
			name:         "Unauthorized",
			statusCode:   401,
			responseBody: `{"status":"error","error":"unauthorized","message":"Authentication required"}`,
			shouldError:  true,
		},
		{
			name:         "Forbidden",
			statusCode:   403,
			responseBody: `{"status":"error","error":"forbidden","message":"Access denied"}`,
			shouldError:  true,
		},
		{
			name:         "Not Found",
			statusCode:   404,
			responseBody: `{"status":"error","error":"not_found","message":"Tag not found"}`,
			shouldError:  true,
		},
		{
			name:         "Internal Server Error",
			statusCode:   500,
			responseBody: `{"status":"error","error":"internal_server_error","message":"Database error"}`,
			shouldError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			_, _, err = client.Tags.List(context.Background(), nil)
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// Query Parameter Tests
// ============================================================================

func TestTagListOptions_ToQuery(t *testing.T) {
	opts := &TagListOptions{
		Namespace: "env",
		Source:    "manual",
		Key:       "environment",
		Page:      2,
		Limit:     100,
	}

	query := opts.ToQuery()
	assert.Equal(t, "env", query["namespace"])
	assert.Equal(t, "manual", query["source"])
	assert.Equal(t, "environment", query["key"])
	assert.Equal(t, "2", query["page"])
	assert.Equal(t, "100", query["limit"])
}

func TestTagHistoryQueryParams_ToQuery(t *testing.T) {
	now := time.Now()
	opts := &TagHistoryQueryParams{
		Action:    "added",
		Namespace: "env",
		Source:    "manual",
		TagID:     123,
		StartDate: now.Format(time.RFC3339),
		EndDate:   now.Add(24 * time.Hour).Format(time.RFC3339),
		Page:      1,
		Limit:     50,
	}

	query := opts.ToQuery()
	assert.Equal(t, "added", query["action"])
	assert.Equal(t, "env", query["namespace"])
	assert.Equal(t, "manual", query["source"])
	assert.Equal(t, "123", query["tag_id"])
	assert.NotEmpty(t, query["start_date"])
	assert.NotEmpty(t, query["end_date"])
	assert.Equal(t, "1", query["page"])
	assert.Equal(t, "50", query["limit"])
}

func TestTagDetectionRuleListOptions_ToQuery(t *testing.T) {
	enabled := true
	opts := &TagDetectionRuleListOptions{
		Enabled:   &enabled,
		Namespace: "env",
		Page:      1,
		Limit:     50,
	}

	query := opts.ToQuery()
	assert.Equal(t, "true", query["enabled"])
	assert.Equal(t, "env", query["namespace"])
	assert.Equal(t, "1", query["page"])
	assert.Equal(t, "50", query["limit"])
}

// ============================================================================
// Tag CRUD Tests (Task #3892)
// ============================================================================

func TestTagsService_GetTag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/tags/123", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		response := StandardResponse{
			Status:  "success",
			Message: "Tag retrieved successfully",
			Data: &Tag{
				ID:             123,
				OrganizationID: 100,
				Namespace:      "env",
				Key:            "environment",
				Value:          "production",
				Description:    "Production environment tag",
				Source:         "manual",
				ServerCount:    15,
				CreatedAt:      CustomTime{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
				UpdatedAt:      CustomTime{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	tag, err := client.Tags.GetTag(context.Background(), 123)
	require.NoError(t, err)
	assert.Equal(t, uint(123), tag.ID)
	assert.Equal(t, "env", tag.Namespace)
	assert.Equal(t, "environment", tag.Key)
	assert.Equal(t, "production", tag.Value)
	assert.Equal(t, "Production environment tag", tag.Description)
	assert.Equal(t, int64(15), tag.ServerCount)
}

func TestTagsService_GetTag_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"error":   "not_found",
			"message": "Tag not found",
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	_, err = client.Tags.GetTag(context.Background(), 999)
	assert.Error(t, err)
}

func TestTagsService_UpdateTag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/tags/123", r.URL.Path)

		var req TagUpdateRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "Updated production environment", req.Description)

		response := StandardResponse{
			Status:  "success",
			Message: "Tag updated successfully",
			Data: &Tag{
				ID:             123,
				OrganizationID: 100,
				Namespace:      "env",
				Key:            "environment",
				Value:          "production",
				Description:    req.Description,
				ServerCount:    15,
				CreatedAt:      CustomTime{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
				UpdatedAt:      CustomTime{time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	req := &TagUpdateRequest{
		Description: "Updated production environment",
	}
	tag, err := client.Tags.UpdateTag(context.Background(), 123, req)
	require.NoError(t, err)
	assert.Equal(t, uint(123), tag.ID)
	assert.Equal(t, "Updated production environment", tag.Description)
}

func TestTagsService_DeleteTag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/tags/123", r.URL.Path)
		assert.Equal(t, "true", r.URL.Query().Get("cascade"))

		response := StandardResponse{
			Status:  "success",
			Message: "Tag deleted successfully",
			Data: &TagDeleteResult{
				TagID:              123,
				RemovedAssociations: 15,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	result, err := client.Tags.DeleteTag(context.Background(), 123, true)
	require.NoError(t, err)
	assert.Equal(t, uint(123), result.TagID)
	assert.Equal(t, int64(15), result.RemovedAssociations)
}

func TestTagsService_DeleteTag_NoCascade(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/tags/123", r.URL.Path)
		assert.Equal(t, "false", r.URL.Query().Get("cascade"))

		response := StandardResponse{
			Status:  "success",
			Message: "Tag deleted successfully",
			Data: &TagDeleteResult{
				TagID:              123,
				RemovedAssociations: 0,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	result, err := client.Tags.DeleteTag(context.Background(), 123, false)
	require.NoError(t, err)
	assert.Equal(t, uint(123), result.TagID)
	assert.Equal(t, int64(0), result.RemovedAssociations)
}

// ============================================================================
// Namespace CRUD Tests (Task #3892)
// ============================================================================

func TestTagsService_UpdateNamespace(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/tags/namespaces/123", r.URL.Path)

		var req NamespaceUpdateRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "Updated environment namespace description", req.Description)
		assert.Equal(t, "^[a-z0-9-]+$", req.KeyPattern)

		response := StandardResponse{
			Status:  "success",
			Message: "Namespace updated successfully",
			Data: &TagNamespace{
				ID:               123,
				Namespace:        "env",
				Description:      req.Description,
				Type:             "system",
				KeyPattern:       req.KeyPattern,
				ValuePattern:     req.ValuePattern,
				AllowedValues:    req.AllowedValues,
				RequiresApproval: false,
				IsActive:         true,
				CreatedAt:        CustomTime{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
				UpdatedAt:        CustomTime{time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	req := &NamespaceUpdateRequest{
		Description:  "Updated environment namespace description",
		KeyPattern:   "^[a-z0-9-]+$",
		ValuePattern: "^[a-zA-Z0-9-]+$",
		AllowedValues: []string{"production", "staging", "development"},
	}
	ns, err := client.Tags.UpdateNamespace(context.Background(), 123, req)
	require.NoError(t, err)
	assert.Equal(t, uint(123), ns.ID)
	assert.Equal(t, "Updated environment namespace description", ns.Description)
	assert.Equal(t, "^[a-z0-9-]+$", ns.KeyPattern)
}

func TestTagsService_DeleteNamespace(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/tags/namespaces/123", r.URL.Path)
		assert.Equal(t, "true", r.URL.Query().Get("cascade"))

		response := StandardResponse{
			Status:  "success",
			Message: "Namespace deleted successfully",
			Data: &NamespaceDeleteResult{
				NamespaceID:        123,
				DeletedTags:        25,
				DeletedChildren:    3,
				DeletedPermissions: 5,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	result, err := client.Tags.DeleteNamespace(context.Background(), 123, true)
	require.NoError(t, err)
	assert.Equal(t, uint(123), result.NamespaceID)
	assert.Equal(t, int64(25), result.DeletedTags)
	assert.Equal(t, int64(3), result.DeletedChildren)
	assert.Equal(t, int64(5), result.DeletedPermissions)
}

func TestTagsService_DeleteNamespace_NoCascade(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/tags/namespaces/123", r.URL.Path)
		assert.Equal(t, "false", r.URL.Query().Get("cascade"))

		response := StandardResponse{
			Status:  "success",
			Message: "Namespace deleted successfully",
			Data: &NamespaceDeleteResult{
				NamespaceID:        123,
				DeletedTags:        0,
				DeletedChildren:    0,
				DeletedPermissions: 0,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	result, err := client.Tags.DeleteNamespace(context.Background(), 123, false)
	require.NoError(t, err)
	assert.Equal(t, uint(123), result.NamespaceID)
	assert.Equal(t, int64(0), result.DeletedTags)
}

// ============================================================================
// Tag-to-Tag Inheritance Tests (Task #3892)
// ============================================================================

func TestTagsService_SetTagInheritance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/tags/456/inherit", r.URL.Path)

		var reqBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&reqBody)
		assert.Equal(t, float64(123), reqBody["parent_tag_id"])

		response := StandardResponse{
			Status:  "success",
			Message: "Tag inheritance set successfully",
			Data: &TagInheritanceRelationship{
				ID:             1,
				OrganizationID: 100,
				ParentTag: TagInfo{
					ID:          123,
					Namespace:   "env",
					Key:         "environment",
					Value:       "production",
					Description: "Production environment",
				},
				ChildTag: TagInfo{
					ID:          456,
					Namespace:   "env",
					Key:         "environment",
					Value:       "prod-us-west",
					Description: "Production US West",
				},
				CreatedBy: &UserInfo{
					ID:    10,
					Email: "admin@example.com",
				},
				CreatedAt: "2025-01-01T00:00:00Z",
				UpdatedAt: "2025-01-01T00:00:00Z",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	rel, err := client.Tags.SetTagInheritance(context.Background(), 456, 123)
	require.NoError(t, err)
	assert.Equal(t, uint(1), rel.ID)
	assert.Equal(t, uint(123), rel.ParentTag.ID)
	assert.Equal(t, uint(456), rel.ChildTag.ID)
	assert.Equal(t, "production", rel.ParentTag.Value)
	assert.Equal(t, "prod-us-west", rel.ChildTag.Value)
}

func TestTagsService_SetTagInheritance_CircularReference(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"error":   "bad_request",
			"message": "Circular reference detected",
			"details": "This inheritance would create a circular reference",
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	_, err = client.Tags.SetTagInheritance(context.Background(), 123, 456)
	assert.Error(t, err)
}

func TestTagsService_GetInheritedTags(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/tags/456/inherited", r.URL.Path)

		response := StandardResponse{
			Status:  "success",
			Message: "Inherited tags retrieved successfully",
			Data: &TagInheritanceChain{
				TagID: 456,
				InheritanceChain: []TagInfo{
					{
						ID:          123,
						Namespace:   "env",
						Key:         "environment",
						Value:       "production",
						Description: "Production environment",
					},
					{
						ID:          789,
						Namespace:   "env",
						Key:         "environment",
						Value:       "root",
						Description: "Root environment",
					},
				},
				InheritanceDepth:  2,
				TotalInheritances: 2,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	chain, err := client.Tags.GetInheritedTags(context.Background(), 456)
	require.NoError(t, err)
	assert.Equal(t, uint(456), chain.TagID)
	assert.Equal(t, 2, chain.InheritanceDepth)
	assert.Equal(t, 2, chain.TotalInheritances)
	assert.Len(t, chain.InheritanceChain, 2)
	assert.Equal(t, uint(123), chain.InheritanceChain[0].ID)
	assert.Equal(t, "production", chain.InheritanceChain[0].Value)
}

func TestTagsService_GetInheritedTags_NoInheritance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := StandardResponse{
			Status:  "success",
			Message: "Inherited tags retrieved successfully",
			Data: &TagInheritanceChain{
				TagID:             456,
				InheritanceChain:  []TagInfo{},
				InheritanceDepth:  0,
				TotalInheritances: 0,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	chain, err := client.Tags.GetInheritedTags(context.Background(), 456)
	require.NoError(t, err)
	assert.Equal(t, 0, chain.InheritanceDepth)
	assert.Len(t, chain.InheritanceChain, 0)
}

func TestTagsService_RemoveTagInheritance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/tags/456/inherit", r.URL.Path)
		assert.Equal(t, "true", r.URL.Query().Get("cascade"))

		response := StandardResponse{
			Status:  "success",
			Message: "Tag inheritance removed successfully",
			Data: &TagInheritanceDeleteResult{
				DeletedRelationships: 3,
				Cascade:              true,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	result, err := client.Tags.RemoveTagInheritance(context.Background(), 456, true)
	require.NoError(t, err)
	assert.Equal(t, 3, result.DeletedRelationships)
	assert.True(t, result.Cascade)
}

func TestTagsService_RemoveTagInheritance_NoCascade(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/tags/456/inherit", r.URL.Path)
		assert.Equal(t, "false", r.URL.Query().Get("cascade"))

		response := StandardResponse{
			Status:  "success",
			Message: "Tag inheritance removed successfully",
			Data: &TagInheritanceDeleteResult{
				DeletedRelationships: 1,
				Cascade:              false,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	result, err := client.Tags.RemoveTagInheritance(context.Background(), 456, false)
	require.NoError(t, err)
	assert.Equal(t, 1, result.DeletedRelationships)
	assert.False(t, result.Cascade)
}

func TestTagsService_RemoveTagInheritance_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"error":   "not_found",
			"message": "No inheritance relationship found",
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	_, err = client.Tags.RemoveTagInheritance(context.Background(), 999, false)
	assert.Error(t, err)
}
