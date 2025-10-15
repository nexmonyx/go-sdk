package nexmonyx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerGroupsService_CreateGroup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/groups", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		var reqBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&reqBody)
		assert.Equal(t, "Production Servers", reqBody["name"])
		assert.Equal(t, "All production environment servers", reqBody["description"])

		response := StandardResponse{
			Status:  "success",
			Message: "Group created successfully",
			Data: &ServerGroup{
				ID:             1,
				OrganizationID: 10,
				Name:           reqBody["name"].(string),
				Description:    reqBody["description"].(string),
				ServerCount:    0,
				Tags:           []string{"production", "critical"},
				CreatedAt:      CustomTime{Time: time.Now()},
				UpdatedAt:      CustomTime{Time: time.Now()},
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

	group, err := client.ServerGroups.CreateGroup(context.Background(),
		"Production Servers",
		"All production environment servers",
		[]string{"production", "critical"})
	require.NoError(t, err)
	assert.Equal(t, "Production Servers", group.Name)
	assert.Equal(t, 0, group.ServerCount)
	assert.Len(t, group.Tags, 2)
}

func TestServerGroupsService_ListGroups(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/groups", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "20", r.URL.Query().Get("limit"))
		assert.Equal(t, "prod", r.URL.Query().Get("name"))
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		response := struct {
			Data []ServerGroup   `json:"data"`
			Meta *PaginationMeta `json:"meta"`
		}{
			Data: []ServerGroup{
				{
					ID:             1,
					OrganizationID: 10,
					Name:           "Production Servers",
					Description:    "All production environment servers",
					ServerCount:    25,
					Tags:           []string{"production"},
					CreatedAt:      CustomTime{Time: time.Now()},
					UpdatedAt:      CustomTime{Time: time.Now()},
				},
				{
					ID:             2,
					OrganizationID: 10,
					Name:           "Production Database Servers",
					Description:    "Database servers in production",
					ServerCount:    5,
					Tags:           []string{"production", "database"},
					CreatedAt:      CustomTime{Time: time.Now()},
					UpdatedAt:      CustomTime{Time: time.Now()},
				},
			},
			Meta: &PaginationMeta{
				Page:       1,
				PerPage:    20,
				TotalItems: 2,
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

	groups, meta, err := client.ServerGroups.ListGroups(context.Background(),
		&PaginationOptions{Page: 1, Limit: 20},
		"prod",
		nil)
	require.NoError(t, err)
	assert.Len(t, groups, 2)
	assert.Equal(t, "Production Servers", groups[0].Name)
	assert.Equal(t, 25, groups[0].ServerCount)
	assert.NotNil(t, meta)
	assert.Equal(t, 2, meta.TotalItems)
}

func TestServerGroupsService_AddServersToGroup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/groups/1/servers", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		var reqBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&reqBody)
		assert.NotNil(t, reqBody["server_uuids"])

		response := StandardResponse{
			Status:  "success",
			Message: "Servers added to group",
			Data: map[string]interface{}{
				"servers_added": 3,
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

	count, err := client.ServerGroups.AddServersToGroup(context.Background(),
		1,
		nil,
		[]string{"uuid-1", "uuid-2", "uuid-3"})
	require.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestServerGroupsService_GetGroupServers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/groups/1/servers", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "50", r.URL.Query().Get("limit"))
		assert.Equal(t, "online", r.URL.Query().Get("status"))
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		response := struct {
			Data []ServerGroupMembership `json:"data"`
			Meta *PaginationMeta         `json:"meta"`
		}{
			Data: []ServerGroupMembership{
				{
					GroupID:      1,
					GroupName:    "Production Servers",
					ServerID:     101,
					ServerUUID:   "uuid-1",
					ServerName:   "web-server-01",
					ServerStatus: "online",
					AddedAt:      CustomTime{Time: time.Now()},
				},
				{
					GroupID:      1,
					GroupName:    "Production Servers",
					ServerID:     102,
					ServerUUID:   "uuid-2",
					ServerName:   "web-server-02",
					ServerStatus: "online",
					AddedAt:      CustomTime{Time: time.Now()},
				},
			},
			Meta: &PaginationMeta{
				Page:       1,
				PerPage:    50,
				TotalItems: 2,
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

	members, meta, err := client.ServerGroups.GetGroupServers(context.Background(),
		1,
		&PaginationOptions{Page: 1, Limit: 50},
		"online",
		nil)
	require.NoError(t, err)
	assert.Len(t, members, 2)
	assert.Equal(t, "web-server-01", members[0].ServerName)
	assert.Equal(t, "online", members[0].ServerStatus)
	assert.NotNil(t, meta)
	assert.Equal(t, 2, meta.TotalItems)
}

// TestServerGroupsService_GetGroupServers_ErrorScenarios tests error handling for GetGroupServers
func TestServerGroupsService_GetGroupServers_ErrorScenarios(t *testing.T) {
	tests := []struct {
		name           string
		groupID        uint
		options        *PaginationOptions
		status         string
		tags           []string
		responseStatus int
		responseData   []ServerGroupMembership
		responseMeta   *PaginationMeta
		responseError  bool
		expectError    bool
		expectedCount  int
	}{
		{
			name:           "empty_results",
			groupID:        uint(2),
			options:        &PaginationOptions{Page: 1, Limit: 25},
			status:         "offline",
			tags:           nil,
			responseStatus: http.StatusOK,
			responseData:   []ServerGroupMembership{}, // Empty results
			responseMeta: &PaginationMeta{
				CurrentPage: 1,
				TotalItems:  0,
				TotalPages:  0,
			},
			expectError:   false,
			expectedCount: 0,
		},
		{
			name:           "server_error_500",
			groupID:        uint(3),
			options:        &PaginationOptions{Page: 1, Limit: 10},
			status:         "",
			tags:           nil,
			responseStatus: http.StatusInternalServerError,
			responseError:  true,
			expectError:    true,
		},
		{
			name:           "unauthorized_401",
			groupID:        uint(4),
			options:        &PaginationOptions{Page: 1, Limit: 10},
			status:         "",
			tags:           nil,
			responseStatus: http.StatusUnauthorized,
			responseError:  true,
			expectError:    true,
		},
		{
			name:           "group_not_found_404",
			groupID:        uint(999),
			options:        &PaginationOptions{Page: 1, Limit: 10},
			status:         "",
			tags:           nil,
			responseStatus: http.StatusNotFound,
			responseError:  true,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, fmt.Sprintf("/v1/groups/%d/servers", tt.groupID), r.URL.Path)

				// Verify query parameters
				query := r.URL.Query()
				if tt.options != nil {
					assert.Equal(t, fmt.Sprintf("%d", tt.options.Page), query.Get("page"))
					assert.Equal(t, fmt.Sprintf("%d", tt.options.Limit), query.Get("limit"))
				}
				if tt.status != "" {
					assert.Equal(t, tt.status, query.Get("status"))
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.responseStatus)

				if tt.responseError {
					json.NewEncoder(w).Encode(map[string]interface{}{
						"status": "error",
						"error":  "Server error",
					})
				} else {
					response := struct {
						Data []ServerGroupMembership `json:"data"`
						Meta *PaginationMeta         `json:"meta"`
					}{
						Data: tt.responseData,
						Meta: tt.responseMeta,
					}
					json.NewEncoder(w).Encode(response)
				}
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			members, meta, err := client.ServerGroups.GetGroupServers(
				context.Background(),
				tt.groupID,
				tt.options,
				tt.status,
				tt.tags,
			)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, members, tt.expectedCount)
				if tt.expectedCount > 0 && meta != nil {
					assert.NotNil(t, meta)
					assert.Equal(t, tt.responseMeta.TotalItems, meta.TotalItems)
				}
			}
		})
	}
}

func TestServerGroupsService_ErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		expectedError bool
	}{
		{
			name:          "Unauthorized",
			statusCode:    http.StatusUnauthorized,
			expectedError: true,
		},
		{
			name:          "Forbidden",
			statusCode:    http.StatusForbidden,
			expectedError: true,
		},
		{
			name:          "Not Found",
			statusCode:    http.StatusNotFound,
			expectedError: true,
		},
		{
			name:          "Internal Server Error",
			statusCode:    http.StatusInternalServerError,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(StandardResponse{
					Status:  "error",
					Message: "Error occurred",
				})
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			_, _, err = client.ServerGroups.ListGroups(context.Background(), nil, "", nil)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
