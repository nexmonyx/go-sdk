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

// TestClustersService_CreateClusterComprehensive tests the CreateCluster method
func TestClustersService_CreateClusterComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		request    *ClusterCreateRequest
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Cluster)
	}{
		{
			name: "success - create cluster with full config",
			request: &ClusterCreateRequest{
				Name:         "production-k8s",
				APIServerURL: "https://k8s.example.com:6443",
				Token:        "test-service-account-token",
				CACert:       "-----BEGIN CERTIFICATE-----\ntest-ca-cert\n-----END CERTIFICATE-----",
			},
			mockStatus: http.StatusCreated,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":              1,
					"name":            "production-k8s",
					"api_server_url":  "https://k8s.example.com:6443",
					"status":          "online",
					"node_count":      5,
					"pod_count":       100,
					"is_active":       true,
					"created_at":      "2024-01-15T10:00:00Z",
					"updated_at":      "2024-01-15T10:00:00Z",
				},
				"status": "success",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, cluster *Cluster) {
				assert.Equal(t, uint(1), cluster.ID)
				assert.Equal(t, "production-k8s", cluster.Name)
				assert.Equal(t, "https://k8s.example.com:6443", cluster.APIServerURL)
				assert.Equal(t, "online", cluster.Status)
				assert.True(t, cluster.IsActive)
			},
		},
		{
			name: "success - minimal cluster config",
			request: &ClusterCreateRequest{
				Name:         "staging-k8s",
				APIServerURL: "https://staging.example.com:6443",
				Token:        "staging-token",
			},
			mockStatus: http.StatusCreated,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":             2,
					"name":           "staging-k8s",
					"api_server_url": "https://staging.example.com:6443",
					"status":         "unknown",
					"node_count":     0,
					"pod_count":      0,
					"is_active":      true,
				},
				"status": "success",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, cluster *Cluster) {
				assert.Equal(t, "staging-k8s", cluster.Name)
				assert.Equal(t, "unknown", cluster.Status)
			},
		},
		{
			name: "validation error - missing name",
			request: &ClusterCreateRequest{
				APIServerURL: "https://k8s.example.com:6443",
				Token:        "test-token",
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Cluster name is required",
			},
			wantErr: true,
		},
		{
			name: "validation error - missing API server URL",
			request: &ClusterCreateRequest{
				Name:  "test-cluster",
				Token: "test-token",
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "API server URL is required",
			},
			wantErr: true,
		},
		{
			name: "validation error - invalid API server URL",
			request: &ClusterCreateRequest{
				Name:         "test-cluster",
				APIServerURL: "not-a-valid-url",
				Token:        "test-token",
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Invalid API server URL format",
			},
			wantErr: true,
		},
		{
			name: "conflict - cluster name already exists",
			request: &ClusterCreateRequest{
				Name:         "production-k8s",
				APIServerURL: "https://k8s.example.com:6443",
				Token:        "test-token",
			},
			mockStatus: http.StatusConflict,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Cluster with this name already exists",
			},
			wantErr: true,
		},
		{
			name: "unauthorized",
			request: &ClusterCreateRequest{
				Name:         "test-cluster",
				APIServerURL: "https://k8s.example.com:6443",
				Token:        "test-token",
			},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name: "forbidden - insufficient permissions",
			request: &ClusterCreateRequest{
				Name:         "test-cluster",
				APIServerURL: "https://k8s.example.com:6443",
				Token:        "test-token",
			},
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Admin privileges required",
			},
			wantErr: true,
		},
		{
			name: "server error",
			request: &ClusterCreateRequest{
				Name:         "test-cluster",
				APIServerURL: "https://k8s.example.com:6443",
				Token:        "test-token",
			},
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Failed to create cluster",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/v1/admin/clusters", r.URL.Path)

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

			result, err := client.Clusters.CreateCluster(ctx, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestClustersService_ListClustersComprehensive tests the ListClusters method
func TestClustersService_ListClustersComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		opts       *PaginationOptions
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, []Cluster, *PaginationMeta)
	}{
		{
			name: "success - list all clusters",
			opts: &PaginationOptions{Page: 1, Limit: 25},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id":             1,
						"name":           "production-k8s",
						"api_server_url": "https://k8s-prod.example.com:6443",
						"status":         "online",
						"node_count":     10,
						"pod_count":      250,
						"is_active":      true,
						"created_at":     "2024-01-15T10:00:00Z",
					},
					{
						"id":             2,
						"name":           "staging-k8s",
						"api_server_url": "https://k8s-staging.example.com:6443",
						"status":         "online",
						"node_count":     3,
						"pod_count":      50,
						"is_active":      true,
					},
				},
				"meta": map[string]interface{}{
					"page":        1,
					"limit":       25,
					"total_items": 2,
					"total_pages": 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, clusters []Cluster, meta *PaginationMeta) {
				assert.Len(t, clusters, 2)
				assert.Equal(t, "production-k8s", clusters[0].Name)
				assert.Equal(t, "online", clusters[0].Status)
				assert.Equal(t, 10, clusters[0].NodeCount)
				assert.Equal(t, 2, meta.TotalItems)
			},
		},
		{
			name:       "success - empty list",
			opts:       &PaginationOptions{Page: 1, Limit: 25},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": []map[string]interface{}{},
				"meta": map[string]interface{}{
					"page":        1,
					"limit":       25,
					"total_items": 0,
					"total_pages": 0,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, clusters []Cluster, meta *PaginationMeta) {
				assert.Len(t, clusters, 0)
				assert.Equal(t, 0, meta.TotalItems)
			},
		},
		{
			name:       "success - with pagination",
			opts:       &PaginationOptions{Page: 2, Limit: 10},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": []map[string]interface{}{
					{"id": 11, "name": "dev-k8s-11", "status": "online", "node_count": 2, "pod_count": 20, "is_active": true},
				},
				"meta": map[string]interface{}{
					"page":        2,
					"limit":       10,
					"total_items": 15,
					"total_pages": 2,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, clusters []Cluster, meta *PaginationMeta) {
				assert.Len(t, clusters, 1)
				assert.Equal(t, 2, meta.Page)
			},
		},
		{
			name:       "unauthorized",
			opts:       &PaginationOptions{Page: 1, Limit: 25},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "forbidden - insufficient permissions",
			opts:       &PaginationOptions{Page: 1, Limit: 25},
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Admin privileges required",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			opts:       &PaginationOptions{Page: 1, Limit: 25},
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
				assert.Equal(t, "/v1/admin/clusters", r.URL.Path)

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

			clusters, meta, err := client.Clusters.ListClusters(ctx, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, clusters)
				if tt.checkFunc != nil {
					tt.checkFunc(t, clusters, meta)
				}
			}
		})
	}
}

// TestClustersService_GetClusterComprehensive tests the GetCluster method
func TestClustersService_GetClusterComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		clusterID  uint
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Cluster)
	}{
		{
			name:       "success - get online cluster",
			clusterID:  1,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":              1,
					"name":            "production-k8s",
					"api_server_url":  "https://k8s-prod.example.com:6443",
					"status":          "online",
					"last_checked":    "2024-01-15T10:30:00Z",
					"last_connected":  "2024-01-15T10:29:00Z",
					"node_count":      10,
					"pod_count":       250,
					"is_active":       true,
					"created_at":      "2024-01-10T10:00:00Z",
					"updated_at":      "2024-01-15T10:30:00Z",
				},
				"status": "success",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, cluster *Cluster) {
				assert.Equal(t, uint(1), cluster.ID)
				assert.Equal(t, "production-k8s", cluster.Name)
				assert.Equal(t, "online", cluster.Status)
				assert.Equal(t, 10, cluster.NodeCount)
				assert.True(t, cluster.IsActive)
			},
		},
		{
			name:       "success - get offline cluster with error",
			clusterID:  2,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":             2,
					"name":           "failing-k8s",
					"api_server_url": "https://k8s-fail.example.com:6443",
					"status":         "error",
					"error_message":  "Connection timeout: failed to connect to API server",
					"last_checked":   "2024-01-15T10:00:00Z",
					"node_count":     0,
					"pod_count":      0,
					"is_active":      true,
				},
				"status": "success",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, cluster *Cluster) {
				assert.Equal(t, "failing-k8s", cluster.Name)
				assert.Equal(t, "error", cluster.Status)
				assert.Equal(t, "Connection timeout: failed to connect to API server", cluster.ErrorMessage)
			},
		},
		{
			name:       "not found",
			clusterID:  999,
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Cluster not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			clusterID:  1,
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "forbidden - insufficient permissions",
			clusterID:  1,
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Admin privileges required",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			clusterID:  1,
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

			result, err := client.Clusters.GetCluster(ctx, tt.clusterID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestClustersService_UpdateClusterComprehensive tests the UpdateCluster method
func TestClustersService_UpdateClusterComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		clusterID  uint
		request    *ClusterUpdateRequest
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Cluster)
	}{
		{
			name:      "success - update cluster name",
			clusterID: 1,
			request: &ClusterUpdateRequest{
				Name: stringPtr("production-k8s-updated"),
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":             1,
					"name":           "production-k8s-updated",
					"api_server_url": "https://k8s-prod.example.com:6443",
					"status":         "online",
					"node_count":     10,
					"pod_count":      250,
					"is_active":      true,
					"updated_at":     "2024-01-15T11:00:00Z",
				},
				"status": "success",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, cluster *Cluster) {
				assert.Equal(t, "production-k8s-updated", cluster.Name)
			},
		},
		{
			name:      "success - update API server URL and token",
			clusterID: 1,
			request: &ClusterUpdateRequest{
				APIServerURL: stringPtr("https://k8s-new.example.com:6443"),
				Token:        stringPtr("new-service-account-token"),
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":             1,
					"name":           "production-k8s",
					"api_server_url": "https://k8s-new.example.com:6443",
					"status":         "online",
					"node_count":     10,
					"pod_count":      250,
					"is_active":      true,
				},
				"status": "success",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, cluster *Cluster) {
				assert.Equal(t, "https://k8s-new.example.com:6443", cluster.APIServerURL)
			},
		},
		{
			name:      "success - disable cluster monitoring",
			clusterID: 1,
			request: &ClusterUpdateRequest{
				IsActive: boolPtr(false),
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":             1,
					"name":           "production-k8s",
					"api_server_url": "https://k8s-prod.example.com:6443",
					"status":         "offline",
					"node_count":     10,
					"pod_count":      250,
					"is_active":      false,
				},
				"status": "success",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, cluster *Cluster) {
				assert.False(t, cluster.IsActive)
			},
		},
		{
			name:      "validation error - invalid API server URL",
			clusterID: 1,
			request: &ClusterUpdateRequest{
				APIServerURL: stringPtr("not-a-valid-url"),
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Invalid API server URL format",
			},
			wantErr: true,
		},
		{
			name:      "not found",
			clusterID: 999,
			request: &ClusterUpdateRequest{
				Name: stringPtr("test"),
			},
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Cluster not found",
			},
			wantErr: true,
		},
		{
			name:      "conflict - name already exists",
			clusterID: 1,
			request: &ClusterUpdateRequest{
				Name: stringPtr("staging-k8s"),
			},
			mockStatus: http.StatusConflict,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Cluster with this name already exists",
			},
			wantErr: true,
		},
		{
			name:      "unauthorized",
			clusterID: 1,
			request: &ClusterUpdateRequest{
				Name: stringPtr("test"),
			},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:      "forbidden - insufficient permissions",
			clusterID: 1,
			request: &ClusterUpdateRequest{
				Name: stringPtr("test"),
			},
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Admin privileges required",
			},
			wantErr: true,
		},
		{
			name:      "server error",
			clusterID: 1,
			request: &ClusterUpdateRequest{
				Name: stringPtr("test"),
			},
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Failed to update cluster",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "PUT", r.Method)

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

			result, err := client.Clusters.UpdateCluster(ctx, tt.clusterID, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestClustersService_DeleteClusterComprehensive tests the DeleteCluster method
func TestClustersService_DeleteClusterComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		clusterID  uint
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:       "success - delete cluster",
			clusterID:  1,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status":  "success",
				"message": "Cluster deleted successfully",
			},
			wantErr: false,
		},
		{
			name:       "success - no content",
			clusterID:  2,
			mockStatus: http.StatusNoContent,
			mockBody:   nil,
			wantErr:    false,
		},
		{
			name:       "not found",
			clusterID:  999,
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Cluster not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			clusterID:  1,
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "forbidden - insufficient permissions",
			clusterID:  1,
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Admin privileges required",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			clusterID:  1,
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Failed to delete cluster",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "DELETE", r.Method)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				if tt.mockBody != nil {
					json.NewEncoder(w).Encode(tt.mockBody)
				}
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

			err = client.Clusters.DeleteCluster(ctx, tt.clusterID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}


// TestClustersService_NetworkErrors tests handling of network-level errors
func TestClustersService_NetworkErrors(t *testing.T) {
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
				_, _, apiErr = client.Clusters.ListClusters(ctx, nil)
			case "get":
				_, apiErr = client.Clusters.GetCluster(ctx, 1)
			case "create":
				req := &ClusterCreateRequest{Name: "test"}
				_, apiErr = client.Clusters.CreateCluster(ctx, req)
			case "update":
				req := &ClusterUpdateRequest{Name: stringPtr("updated")}
				_, apiErr = client.Clusters.UpdateCluster(ctx, 1, req)
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

// TestClustersService_ConcurrentOperations tests concurrent operations on clusters
func TestClustersService_ConcurrentOperations(t *testing.T) {
	tests := []struct {
		name              string
		concurrencyLevel  int
		operationsPerGoro int
		operation         string
		mockStatus        int
		mockBody          interface{}
	}{
		{
			name:              "concurrent ListClusters - low concurrency",
			concurrencyLevel:  10,
			operationsPerGoro: 5,
			operation:         "list",
			mockStatus:        http.StatusOK,
			mockBody: map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id":             1,
						"name":           "production-k8s",
						"api_server_url": "https://k8s.example.com:6443",
						"status":         "online",
						"is_active":      true,
					},
				},
				"meta": map[string]interface{}{"total_items": 1},
			},
		},
		{
			name:              "concurrent GetCluster - medium concurrency",
			concurrencyLevel:  50,
			operationsPerGoro: 2,
			operation:         "get",
			mockStatus:        http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":             1,
					"name":           "production-k8s",
					"api_server_url": "https://k8s.example.com:6443",
					"status":         "online",
					"is_active":      true,
				},
			},
		},
		{
			name:              "concurrent CreateCluster - medium concurrency",
			concurrencyLevel:  30,
			operationsPerGoro: 2,
			operation:         "create",
			mockStatus:        http.StatusCreated,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":             2,
					"name":           "test-k8s",
					"api_server_url": "https://test.example.com:6443",
					"status":         "unknown",
					"is_active":      true,
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
				"data": []map[string]interface{}{},
				"meta": map[string]interface{}{"total_items": 0},
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
				RetryCount: 0,
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
							_, _, apiErr = client.Clusters.ListClusters(context.Background(), nil)
						case "get":
							_, apiErr = client.Clusters.GetCluster(context.Background(), 1)
						case "create":
							req := &ClusterCreateRequest{Name: "test-k8s", APIServerURL: "https://test.example.com:6443"}
							_, apiErr = client.Clusters.CreateCluster(context.Background(), req)
						case "update":
							req := &ClusterUpdateRequest{Name: stringPtr("updated-k8s")}
							_, apiErr = client.Clusters.UpdateCluster(context.Background(), 1, req)
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

			totalOps := int64(tt.concurrencyLevel * tt.operationsPerGoro)
			assert.Equal(t, totalOps, successCount+errorCount, "Total operations should equal success + error count")
			assert.Equal(t, int64(0), errorCount, "Expected no errors in concurrent operations")
			assert.Equal(t, totalOps, successCount, "All operations should succeed")

			t.Logf("Completed %d operations in %v (%.2f ops/sec)",
				totalOps, duration, float64(totalOps)/duration.Seconds())
		})
	}
}

