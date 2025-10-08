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
		assert.Equal(t, "/v1/servers/server-uuid-123/tags", r.URL.Path)

		tags := []ServerTag{
			{
				ID:        1,
				TagID:     10,
				Namespace: "env",
				Key:       "environment",
				Value:     "production",
				Source:    "manual",
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
			},
		}
		response := StandardResponse{
			Status:  "success",
			Message: "Namespaces retrieved successfully",
			Data:    namespaces,
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
	assert.Equal(t, 0, total) // No pagination in response
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
			},
		}
		response := StandardResponse{
			Status:  "success",
			Message: "Organization tags retrieved successfully",
			Data:    orgTags,
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
	assert.Equal(t, 0, total)
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
			},
		}
		response := StandardResponse{
			Status:  "success",
			Message: "Server relationships retrieved successfully",
			Data:    relationships,
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
	assert.Equal(t, 0, total)
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
