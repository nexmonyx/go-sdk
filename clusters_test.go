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

func TestClustersService_CreateCluster(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/admin/clusters", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		var reqBody ClusterCreateRequest
		json.NewDecoder(r.Body).Decode(&reqBody)
		assert.Equal(t, "Production Cluster", reqBody.Name)
		assert.Equal(t, "https://k8s.example.com:6443", reqBody.APIServerURL)
		assert.NotEmpty(t, reqBody.Token)

		response := struct {
			Data    *Cluster `json:"data"`
			Status  string   `json:"status"`
			Message string   `json:"message"`
		}{
			Data: &Cluster{
				ID:            1,
				Name:          reqBody.Name,
				APIServerURL:  reqBody.APIServerURL,
				Token:         reqBody.Token,
				CACert:        reqBody.CACert,
				Status:        "unknown",
				IsActive:      true,
				NodeCount:     0,
				PodCount:      0,
				CreatedAt:     CustomTime{Time: time.Now()},
				UpdatedAt:     CustomTime{Time: time.Now()},
			},
			Status:  "success",
			Message: "Cluster created successfully",
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

	cluster, err := client.Clusters.CreateCluster(context.Background(), &ClusterCreateRequest{
		Name:         "Production Cluster",
		APIServerURL: "https://k8s.example.com:6443",
		Token:        "sa-token-xyz",
		CACert:       "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
	})
	require.NoError(t, err)
	assert.Equal(t, "Production Cluster", cluster.Name)
	assert.Equal(t, "https://k8s.example.com:6443", cluster.APIServerURL)
	assert.Equal(t, "unknown", cluster.Status)
	assert.True(t, cluster.IsActive)
}

func TestClustersService_ListClusters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/admin/clusters", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "50", r.URL.Query().Get("limit"))
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		response := struct {
			Data []Cluster       `json:"data"`
			Meta *PaginationMeta `json:"meta"`
		}{
			Data: []Cluster{
				{
					ID:            1,
					Name:          "Production Cluster",
					APIServerURL:  "https://k8s-prod.example.com:6443",
					Status:        "online",
					IsActive:      true,
					NodeCount:     5,
					PodCount:      150,
					LastConnected: &CustomTime{Time: time.Now()},
					CreatedAt:     CustomTime{Time: time.Now()},
					UpdatedAt:     CustomTime{Time: time.Now()},
				},
				{
					ID:            2,
					Name:          "Staging Cluster",
					APIServerURL:  "https://k8s-staging.example.com:6443",
					Status:        "online",
					IsActive:      true,
					NodeCount:     3,
					PodCount:      75,
					LastConnected: &CustomTime{Time: time.Now()},
					CreatedAt:     CustomTime{Time: time.Now()},
					UpdatedAt:     CustomTime{Time: time.Now()},
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

	clusters, meta, err := client.Clusters.ListClusters(context.Background(),
		&PaginationOptions{Page: 1, Limit: 50})
	require.NoError(t, err)
	assert.Len(t, clusters, 2)
	assert.Equal(t, "Production Cluster", clusters[0].Name)
	assert.Equal(t, "online", clusters[0].Status)
	assert.Equal(t, 5, clusters[0].NodeCount)
	assert.Equal(t, 150, clusters[0].PodCount)
	assert.Equal(t, "Staging Cluster", clusters[1].Name)
	assert.Equal(t, 3, clusters[1].NodeCount)
	assert.NotNil(t, meta)
	assert.Equal(t, 2, meta.TotalItems)
}

func TestClustersService_GetCluster(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/admin/clusters/123", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		lastChecked := CustomTime{Time: time.Now().Add(-5 * time.Minute)}
		lastConnected := CustomTime{Time: time.Now().Add(-1 * time.Minute)}

		response := struct {
			Data    *Cluster `json:"data"`
			Status  string   `json:"status"`
			Message string   `json:"message"`
		}{
			Data: &Cluster{
				ID:            123,
				Name:          "Production Cluster",
				APIServerURL:  "https://k8s-prod.example.com:6443",
				Token:         "sa-token-xyz",
				CACert:        "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
				Status:        "online",
				LastChecked:   &lastChecked,
				LastConnected: &lastConnected,
				NodeCount:     5,
				PodCount:      150,
				IsActive:      true,
				CreatedAt:     CustomTime{Time: time.Now()},
				UpdatedAt:     CustomTime{Time: time.Now()},
			},
			Status:  "success",
			Message: "Cluster retrieved successfully",
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

	cluster, err := client.Clusters.GetCluster(context.Background(), 123)
	require.NoError(t, err)
	assert.Equal(t, uint(123), cluster.ID)
	assert.Equal(t, "Production Cluster", cluster.Name)
	assert.Equal(t, "https://k8s-prod.example.com:6443", cluster.APIServerURL)
	assert.Equal(t, "online", cluster.Status)
	assert.Equal(t, 5, cluster.NodeCount)
	assert.Equal(t, 150, cluster.PodCount)
	assert.True(t, cluster.IsActive)
	assert.NotEmpty(t, cluster.Token)
	assert.NotEmpty(t, cluster.CACert)
}

func TestClustersService_UpdateCluster(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/admin/clusters/456", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		var reqBody ClusterUpdateRequest
		json.NewDecoder(r.Body).Decode(&reqBody)
		assert.NotNil(t, reqBody.Name)
		assert.Equal(t, "Updated Cluster Name", *reqBody.Name)

		response := struct {
			Data    *Cluster `json:"data"`
			Status  string   `json:"status"`
			Message string   `json:"message"`
		}{
			Data: &Cluster{
				ID:            456,
				Name:          *reqBody.Name,
				APIServerURL:  "https://k8s-prod.example.com:6443",
				Token:         "sa-token-xyz",
				Status:        "online",
				NodeCount:     5,
				PodCount:      150,
				IsActive:      true,
				CreatedAt:     CustomTime{Time: time.Now()},
				UpdatedAt:     CustomTime{Time: time.Now()},
			},
			Status:  "success",
			Message: "Cluster updated successfully",
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

	updatedName := "Updated Cluster Name"
	cluster, err := client.Clusters.UpdateCluster(context.Background(), 456, &ClusterUpdateRequest{
		Name: &updatedName,
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated Cluster Name", cluster.Name)
	assert.Equal(t, uint(456), cluster.ID)
}

func TestClustersService_DeleteCluster(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/admin/clusters/789", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		response := struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		}{
			Status:  "success",
			Message: "Cluster deleted successfully",
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

	err = client.Clusters.DeleteCluster(context.Background(), 789)
	require.NoError(t, err)
}

func TestClustersService_ErrorHandling(t *testing.T) {
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

			_, _, err = client.Clusters.ListClusters(context.Background(), nil)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
