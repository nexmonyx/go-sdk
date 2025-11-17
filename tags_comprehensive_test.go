package nexmonyx

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTagsService_ListComprehensive tests the List method
func TestTagsService_ListComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		opts       *TagListOptions
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, []*Tag, *PaginationMeta)
	}{
		{
			name: "success - list all tags",
			opts: &TagListOptions{Page: 1, Limit: 25},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id":              1,
						"organization_id": 1,
						"namespace":       "environment",
						"key":             "env",
						"value":           "production",
						"source":          "manual",
						"description":     "Production environment",
						"server_count":    10,
						"created_at":      "2024-01-15T10:00:00Z",
					},
					{
						"id":           2,
						"namespace":    "department",
						"key":          "dept",
						"value":        "engineering",
						"source":       "automatic",
						"server_count": 5,
					},
				},
				"pagination": map[string]interface{}{
					"page":        1,
					"limit":       25,
					"total_items": 2,
					"total_pages": 1,
				},
				"status": "success",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, tags []*Tag, meta *PaginationMeta) {
				assert.Len(t, tags, 2)
				assert.Equal(t, "environment", tags[0].Namespace)
				assert.Equal(t, "production", tags[0].Value)
				assert.Equal(t, int64(10), tags[0].ServerCount)
				assert.Equal(t, 2, meta.TotalItems)
			},
		},
		{
			name:       "success - empty list",
			opts:       &TagListOptions{Page: 1, Limit: 25},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": []map[string]interface{}{},
				"pagination": map[string]interface{}{
					"page":        1,
					"limit":       25,
					"total_items": 0,
					"total_pages": 0,
				},
				"status": "success",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, tags []*Tag, meta *PaginationMeta) {
				assert.Len(t, tags, 0)
				assert.Equal(t, 0, meta.TotalItems)
			},
		},
		{
			name: "success - with namespace filter",
			opts: &TagListOptions{
				Page:      1,
				Limit:     10,
				Namespace: "environment",
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": []map[string]interface{}{
					{"id": 1, "namespace": "environment", "key": "env", "value": "production", "server_count": 10},
				},
				"pagination": map[string]interface{}{
					"page":        1,
					"limit":       10,
					"total_items": 1,
					"total_pages": 1,
				},
				"status": "success",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, tags []*Tag, meta *PaginationMeta) {
				assert.Len(t, tags, 1)
				assert.Equal(t, "environment", tags[0].Namespace)
			},
		},
		{
			name:       "unauthorized",
			opts:       &TagListOptions{Page: 1, Limit: 25},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			opts:       &TagListOptions{Page: 1, Limit: 25},
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
				assert.Equal(t, "/v1/tags", r.URL.Path)

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

			tags, meta, err := client.Tags.List(ctx, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, tags)
				if tt.checkFunc != nil {
					tt.checkFunc(t, tags, meta)
				}
			}
		})
	}
}

// TestTagsService_CreateComprehensive tests the Create method
func TestTagsService_CreateComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		request    *TagCreateRequest
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Tag)
	}{
		{
			name: "success - create tag with description",
			request: &TagCreateRequest{
				Namespace:   "environment",
				Key:         "env",
				Value:       "production",
				Description: "Production environment tag",
			},
			mockStatus: http.StatusCreated,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":              1,
					"organization_id": 1,
					"namespace":       "environment",
					"key":             "env",
					"value":           "production",
					"source":          "manual",
					"description":     "Production environment tag",
					"server_count":    0,
					"created_at":      "2024-01-15T10:00:00Z",
					"updated_at":      "2024-01-15T10:00:00Z",
				},
				"status": "success",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, tag *Tag) {
				assert.Equal(t, uint(1), tag.ID)
				assert.Equal(t, "environment", tag.Namespace)
				assert.Equal(t, "production", tag.Value)
				assert.Equal(t, "Production environment tag", tag.Description)
			},
		},
		{
			name: "success - minimal tag",
			request: &TagCreateRequest{
				Namespace: "department",
				Key:       "dept",
				Value:     "engineering",
			},
			mockStatus: http.StatusCreated,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":           2,
					"namespace":    "department",
					"key":          "dept",
					"value":        "engineering",
					"source":       "manual",
					"server_count": 0,
				},
				"status": "success",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, tag *Tag) {
				assert.Equal(t, "department", tag.Namespace)
				assert.Equal(t, "engineering", tag.Value)
			},
		},
		{
			name: "validation error - missing namespace",
			request: &TagCreateRequest{
				Key:   "env",
				Value: "production",
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Namespace is required",
			},
			wantErr: true,
		},
		{
			name: "validation error - missing key",
			request: &TagCreateRequest{
				Namespace: "environment",
				Value:     "production",
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Key is required",
			},
			wantErr: true,
		},
		{
			name: "conflict - tag already exists",
			request: &TagCreateRequest{
				Namespace: "environment",
				Key:       "env",
				Value:     "production",
			},
			mockStatus: http.StatusConflict,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Tag with this namespace, key, and value already exists",
			},
			wantErr: true,
		},
		{
			name: "unauthorized",
			request: &TagCreateRequest{
				Namespace: "environment",
				Key:       "env",
				Value:     "production",
			},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name: "server error",
			request: &TagCreateRequest{
				Namespace: "environment",
				Key:       "env",
				Value:     "production",
			},
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Failed to create tag",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/v1/tags", r.URL.Path)

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

			result, err := client.Tags.Create(ctx, tt.request)

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

// TestTagsService_GetServerTagsComprehensive tests the GetServerTags method
func TestTagsService_GetServerTagsComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		serverID   string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, []*ServerTag)
	}{
		{
			name:       "success - get server tags",
			serverID:   "server-uuid-123",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id":              1,
						"server_id":       1,
						"server_uuid":     "server-uuid-123",
						"tag_id":          1,
						"namespace":       "environment",
						"key":             "env",
						"value":           "production",
						"source":          "manual",
						"inherited":       false,
						"inherited_from":  "",
						"assigned_at":     "2024-01-15T10:00:00Z",
					},
					{
						"id":             2,
						"server_id":      1,
						"tag_id":         2,
						"namespace":      "department",
						"key":            "dept",
						"value":          "engineering",
						"source":         "automatic",
						"inherited":      true,
						"inherited_from": "organization",
					},
				},
				"status": "success",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, tags []*ServerTag) {
				assert.Len(t, tags, 2)
				assert.Equal(t, "environment", tags[0].Namespace)
				assert.Equal(t, "production", tags[0].Value)
				assert.Equal(t, "manual", tags[0].Source)
				assert.Equal(t, "automatic", tags[1].Source)
			},
		},
		{
			name:       "success - empty tags",
			serverID:   "server-uuid-456",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data":   []map[string]interface{}{},
				"status": "success",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, tags []*ServerTag) {
				assert.Len(t, tags, 0)
			},
		},
		{
			name:       "not found",
			serverID:   "non-existent-uuid",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Server not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			serverID:   "server-uuid-123",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			serverID:   "server-uuid-123",
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

			result, err := client.Tags.GetServerTags(ctx, tt.serverID)

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

// TestTagsService_AssignTagsToServerComprehensive tests the AssignTagsToServer method
func TestTagsService_AssignTagsToServerComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		serverID   string
		request    *TagAssignRequest
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *TagAssignmentResult)
	}{
		{
			name:     "success - assign tags",
			serverID: "server-uuid-123",
			request: &TagAssignRequest{
				TagIDs: []uint{1, 2, 3},
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"assigned":         3,
					"already_assigned": 0,
					"total":            3,
				},
				"status": "success",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, result *TagAssignmentResult) {
				assert.Equal(t, 3, result.Assigned)
				assert.Equal(t, 0, result.AlreadyAssigned)
				assert.Equal(t, 3, result.Total)
			},
		},
		{
			name:     "success - partial assignment",
			serverID: "server-uuid-123",
			request: &TagAssignRequest{
				TagIDs: []uint{1, 2, 3, 4},
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"assigned":         2,
					"already_assigned": 2,
					"total":            4,
				},
				"status": "success",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, result *TagAssignmentResult) {
				assert.Equal(t, 2, result.Assigned)
				assert.Equal(t, 2, result.AlreadyAssigned)
				assert.Equal(t, 4, result.Total)
			},
		},
		{
			name:     "validation error - empty tag list",
			serverID: "server-uuid-123",
			request: &TagAssignRequest{
				TagIDs: []uint{},
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "At least one tag ID is required",
			},
			wantErr: true,
		},
		{
			name:     "not found - server",
			serverID: "non-existent-uuid",
			request: &TagAssignRequest{
				TagIDs: []uint{1, 2},
			},
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Server not found",
			},
			wantErr: true,
		},
		{
			name:     "unauthorized",
			serverID: "server-uuid-123",
			request: &TagAssignRequest{
				TagIDs: []uint{1, 2},
			},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:     "server error",
			serverID: "server-uuid-123",
			request: &TagAssignRequest{
				TagIDs: []uint{1, 2},
			},
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Failed to assign tags",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)

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

			result, err := client.Tags.AssignTagsToServer(ctx, tt.serverID, tt.request)

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

// TestTagsService_RemoveTagFromServerComprehensive tests the RemoveTagFromServer method
func TestTagsService_RemoveTagFromServerComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		serverID   string
		tagID      uint
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:       "success - remove tag",
			serverID:   "server-uuid-123",
			tagID:      1,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status":  "success",
				"message": "Tag removed from server",
			},
			wantErr: false,
		},
		{
			name:       "success - no content",
			serverID:   "server-uuid-123",
			tagID:      2,
			mockStatus: http.StatusNoContent,
			mockBody:   nil,
			wantErr:    false,
		},
		{
			name:       "not found - tag not assigned",
			serverID:   "server-uuid-123",
			tagID:      999,
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Tag not assigned to server",
			},
			wantErr: true,
		},
		{
			name:       "not found - server",
			serverID:   "non-existent-uuid",
			tagID:      1,
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Server not found",
			},
			wantErr: true,
		},
		{
			name:       "conflict - inherited tag",
			serverID:   "server-uuid-123",
			tagID:      1,
			mockStatus: http.StatusConflict,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Cannot remove inherited tag",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			serverID:   "server-uuid-123",
			tagID:      1,
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			serverID:   "server-uuid-123",
			tagID:      1,
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Failed to remove tag",
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

			err = client.Tags.RemoveTagFromServer(ctx, tt.serverID, tt.tagID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestTagsService_NetworkErrors tests handling of network-level errors
func TestTagsService_NetworkErrors(t *testing.T) {
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
			operation:     "create",
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
			operation:     "list",
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
			operation:     "create",
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
				_, _, apiErr = client.Tags.List(ctx, nil)
			case "create":
				req := &TagCreateRequest{Key: "test", Value: "test"}
				_, apiErr = client.Tags.Create(ctx, req)
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
