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

func TestSearchService_SearchServers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/search/servers", r.URL.Path)
		assert.Equal(t, "web", r.URL.Query().Get("query"))
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "20", r.URL.Query().Get("limit"))
		assert.Equal(t, "production", r.URL.Query().Get("environment"))
		assert.Equal(t, "online", r.URL.Query().Get("status"))
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		response := struct {
			Data []SearchResult  `json:"data"`
			Meta *PaginationMeta `json:"meta"`
		}{
			Data: []SearchResult{
				{
					ServerID:       1,
					ServerUUID:     "uuid-123",
					ServerName:     "web-server-01",
					Hostname:       "web01.example.com",
					OrganizationID: 10,
					Location:       "us-east-1",
					Environment:    "production",
					Classification: "critical",
					Status:         "online",
					IPAddresses:    []string{"10.0.1.10", "192.168.1.10"},
					Tags:           []string{"web", "production", "critical"},
					RelevanceScore: 0.95,
					MatchedFields:  []string{"name", "tags"},
					LastSeenAt:     &CustomTime{Time: time.Now()},
					CreatedAt:      CustomTime{Time: time.Now()},
				},
				{
					ServerID:       2,
					ServerUUID:     "uuid-456",
					ServerName:     "web-server-02",
					Hostname:       "web02.example.com",
					OrganizationID: 10,
					Location:       "us-west-2",
					Environment:    "production",
					Status:         "online",
					IPAddresses:    []string{"10.0.2.10"},
					Tags:           []string{"web", "production"},
					RelevanceScore: 0.87,
					MatchedFields:  []string{"tags"},
					CreatedAt:      CustomTime{Time: time.Now()},
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

	results, meta, err := client.Search.SearchServers(context.Background(),
		"web",
		&PaginationOptions{Page: 1, Limit: 20},
		map[string]interface{}{
			"environment": "production",
			"status":      "online",
		})
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "web-server-01", results[0].ServerName)
	assert.Equal(t, float64(0.95), results[0].RelevanceScore)
	assert.Len(t, results[0].Tags, 3)
	assert.Contains(t, results[0].MatchedFields, "name")
	assert.NotNil(t, meta)
	assert.Equal(t, 2, meta.TotalItems)
}

func TestSearchService_SearchServers_WithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/search/servers", r.URL.Path)
		assert.Equal(t, "database", r.URL.Query().Get("query"))
		assert.Equal(t, "us-east-1", r.URL.Query().Get("location"))
		assert.Equal(t, "production", r.URL.Query().Get("environment"))
		assert.Equal(t, "critical", r.URL.Query().Get("classification"))

		response := struct {
			Data []SearchResult  `json:"data"`
			Meta *PaginationMeta `json:"meta"`
		}{
			Data: []SearchResult{
				{
					ServerID:       5,
					ServerUUID:     "uuid-789",
					ServerName:     "db-master-01",
					Location:       "us-east-1",
					Environment:    "production",
					Classification: "critical",
					Status:         "online",
					Tags:           []string{"database", "postgresql", "master"},
					RelevanceScore: 0.98,
					CreatedAt:      CustomTime{Time: time.Now()},
				},
			},
			Meta: &PaginationMeta{
				Page:       1,
				PerPage:    20,
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

	results, meta, err := client.Search.SearchServers(context.Background(),
		"database",
		nil,
		map[string]interface{}{
			"location":       "us-east-1",
			"environment":    "production",
			"classification": "critical",
		})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "db-master-01", results[0].ServerName)
	assert.Equal(t, "critical", results[0].Classification)
	assert.NotNil(t, meta)
}

func TestSearchService_SearchTags(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/search/tags", r.URL.Path)
		assert.Equal(t, "prod", r.URL.Query().Get("query"))
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "50", r.URL.Query().Get("limit"))
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		response := struct {
			Data []TagSearchResult `json:"data"`
			Meta *PaginationMeta   `json:"meta"`
		}{
			Data: []TagSearchResult{
				{
					TagID:          1,
					TagName:        "production",
					TagType:        "manual",
					Scope:          "organization",
					Description:    "Production environment servers",
					Color:          "#FF0000",
					UsageCount:     45,
					ServerCount:    42,
					RelevanceScore: 0.92,
					MatchedFields:  []string{"name"},
					CreatedAt:      CustomTime{Time: time.Now()},
					UpdatedAt:      CustomTime{Time: time.Now()},
				},
				{
					TagID:          2,
					TagName:        "prod-critical",
					TagType:        "auto",
					Scope:          "server",
					Description:    "Critical production systems",
					UsageCount:     12,
					ServerCount:    12,
					RelevanceScore: 0.85,
					MatchedFields:  []string{"name"},
					CreatedAt:      CustomTime{Time: time.Now()},
					UpdatedAt:      CustomTime{Time: time.Now()},
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

	results, meta, err := client.Search.SearchTags(context.Background(),
		"prod",
		&PaginationOptions{Page: 1, Limit: 50},
		nil)
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "production", results[0].TagName)
	assert.Equal(t, "manual", results[0].TagType)
	assert.Equal(t, 45, results[0].UsageCount)
	assert.Equal(t, 42, results[0].ServerCount)
	assert.NotNil(t, meta)
	assert.Equal(t, 2, meta.TotalItems)
}

func TestSearchService_SearchTags_WithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/search/tags", r.URL.Path)
		assert.Equal(t, "system", r.URL.Query().Get("query"))
		assert.Equal(t, "auto", r.URL.Query().Get("tag_type"))
		assert.Equal(t, "server", r.URL.Query().Get("scope"))

		response := struct {
			Data []TagSearchResult `json:"data"`
			Meta *PaginationMeta   `json:"meta"`
		}{
			Data: []TagSearchResult{
				{
					TagID:          10,
					TagName:        "system-generated",
					TagType:        "auto",
					Scope:          "server",
					UsageCount:     150,
					ServerCount:    120,
					RelevanceScore: 0.88,
					CreatedAt:      CustomTime{Time: time.Now()},
					UpdatedAt:      CustomTime{Time: time.Now()},
				},
			},
			Meta: &PaginationMeta{
				Page:       1,
				PerPage:    20,
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

	results, meta, err := client.Search.SearchTags(context.Background(),
		"system",
		nil,
		map[string]interface{}{
			"tag_type": "auto",
			"scope":    "server",
		})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "system-generated", results[0].TagName)
	assert.Equal(t, "auto", results[0].TagType)
	assert.Equal(t, "server", results[0].Scope)
	assert.NotNil(t, meta)
}

func TestSearchService_GetTagStatistics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/search/tags/statistics", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		response := struct {
			Data    *TagStatistics `json:"data"`
			Status  string         `json:"status"`
			Message string         `json:"message"`
		}{
			Data: &TagStatistics{
				TotalTags:  125,
				ManualTags: 75,
				AutoTags:   40,
				SystemTags: 10,
				TagsByScope: map[string]int{
					"organization": 60,
					"user":         30,
					"server":       35,
				},
				MostUsedTags: []TagUsageStats{
					{
						TagID:       1,
						TagName:     "production",
						TagType:     "manual",
						UsageCount:  150,
						ServerCount: 120,
						LastUsedAt:  CustomTime{Time: time.Now()},
					},
					{
						TagID:       2,
						TagName:     "critical",
						TagType:     "manual",
						UsageCount:  95,
						ServerCount: 85,
						LastUsedAt:  CustomTime{Time: time.Now()},
					},
				},
				RecentlyCreated: []TagSearchResult{
					{
						TagID:       100,
						TagName:     "new-app",
						TagType:     "manual",
						Scope:       "organization",
						UsageCount:  5,
						ServerCount: 5,
						CreatedAt:   CustomTime{Time: time.Now()},
						UpdatedAt:   CustomTime{Time: time.Now()},
					},
				},
				UnusedTags:       15,
				AveragePerServer: 3.5,
			},
			Status:  "success",
			Message: "Tag statistics retrieved successfully",
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

	stats, err := client.Search.GetTagStatistics(context.Background(), "", "")
	require.NoError(t, err)
	assert.Equal(t, 125, stats.TotalTags)
	assert.Equal(t, 75, stats.ManualTags)
	assert.Equal(t, 40, stats.AutoTags)
	assert.Equal(t, 10, stats.SystemTags)
	assert.Len(t, stats.TagsByScope, 3)
	assert.Equal(t, 60, stats.TagsByScope["organization"])
	assert.Len(t, stats.MostUsedTags, 2)
	assert.Equal(t, "production", stats.MostUsedTags[0].TagName)
	assert.Equal(t, 150, stats.MostUsedTags[0].UsageCount)
	assert.Len(t, stats.RecentlyCreated, 1)
	assert.Equal(t, 15, stats.UnusedTags)
	assert.Equal(t, 3.5, stats.AveragePerServer)
}

func TestSearchService_GetTagStatistics_WithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/search/tags/statistics", r.URL.Path)
		assert.Equal(t, "manual", r.URL.Query().Get("tag_type"))
		assert.Equal(t, "organization", r.URL.Query().Get("scope"))

		response := struct {
			Data    *TagStatistics `json:"data"`
			Status  string         `json:"status"`
			Message string         `json:"message"`
		}{
			Data: &TagStatistics{
				TotalTags:  60,
				ManualTags: 60,
				AutoTags:   0,
				SystemTags: 0,
				TagsByScope: map[string]int{
					"organization": 60,
				},
				MostUsedTags: []TagUsageStats{
					{
						TagID:       1,
						TagName:     "production",
						TagType:     "manual",
						UsageCount:  150,
						ServerCount: 120,
						LastUsedAt:  CustomTime{Time: time.Now()},
					},
				},
				RecentlyCreated:  []TagSearchResult{},
				UnusedTags:       5,
				AveragePerServer: 2.8,
			},
			Status:  "success",
			Message: "Tag statistics retrieved successfully",
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

	stats, err := client.Search.GetTagStatistics(context.Background(), "manual", "organization")
	require.NoError(t, err)
	assert.Equal(t, 60, stats.TotalTags)
	assert.Equal(t, 60, stats.ManualTags)
	assert.Equal(t, 0, stats.AutoTags)
	assert.Len(t, stats.TagsByScope, 1)
}

func TestSearchService_ErrorHandling(t *testing.T) {
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

			_, _, err = client.Search.SearchServers(context.Background(), "test", nil, nil)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
