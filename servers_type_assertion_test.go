package nexmonyx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServersService_TypeAssertionErrors tests the "unexpected response type" error paths
// for methods at 87.5% coverage in servers.go
func TestServersService_TypeAssertionErrors(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(w http.ResponseWriter, r *http.Request)
		testMethod func(t *testing.T, client *Client)
	}{
		{
			name: "GetByUUID - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				result, err := client.Servers.GetByUUID(context.Background(), "server-uuid-123")
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "unexpected response type")
			},
		},
		{
			name: "Create - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				server := &Server{Hostname: "test-server"}
				result, err := client.Servers.Create(context.Background(), server)
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "unexpected response type")
			},
		},
		{
			name: "Update - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				server := &Server{Hostname: "updated-server"}
				result, err := client.Servers.Update(context.Background(), "server-123", server)
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "unexpected response type")
			},
		},
		{
			name: "UpdateDetails - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				req := &ServerDetailsUpdateRequest{Hostname: "updated"}
				result, err := client.Servers.UpdateDetails(context.Background(), "server-123", req)
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "unexpected response type")
			},
		},
		{
			name: "Register - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				result, err := client.Servers.Register(context.Background(), "test-server", 1)
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "unexpected response type")
			},
		},
		{
			name: "UpdateTags - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				result, err := client.Servers.UpdateTags(context.Background(), "server-123", []string{"tag1", "tag2"})
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "unexpected response type")
			},
		},
		{
			name: "GetSystemInfo - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				result, err := client.Servers.GetSystemInfo(context.Background(), "server-123")
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "unexpected response type")
			},
		},
		{
			name: "RegisterWithKeyFull - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				req := &ServerCreateRequest{
					Hostname: "test-server",
					MainIP:   "192.168.1.1",
				}
				result, err := client.Servers.RegisterWithKeyFull(context.Background(), "test-key", req)
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "unexpected response type")
			},
		},
		{
			name: "RegisterWithUnifiedKeyFull - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				key := &UnifiedAPIKey{
					Type:       APIKeyTypeRegistration,
					Key:        "test-key",
					FullToken:  "test-token",
					Status:     APIKeyStatusActive,
					Scopes:     []string{"servers:register"},
				}
				req := &ServerCreateRequest{
					Hostname: "test-server",
					MainIP:   "192.168.1.1",
				}
				result, err := client.Servers.RegisterWithUnifiedKeyFull(context.Background(), key, req)
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "unexpected response type")
			},
		},
		{
			name: "UpdateServer - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				req := &ServerUpdateRequest{
					Hostname: "updated-server",
				}
				result, err := client.Servers.UpdateServer(context.Background(), "server-123", req)
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "unexpected response type")
			},
		},
		{
			name: "UpdateInfo - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				req := &ServerDetailsUpdateRequest{Hostname: "updated"}
				result, err := client.Servers.UpdateInfo(context.Background(), "server-123", req)
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "unexpected response type")
			},
		},
		{
			name: "GetDetails - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				result, err := client.Servers.GetDetails(context.Background(), "server-123")
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "unexpected response type")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.setupMock))
			defer server.Close()

			client, err := NewClient(&Config{BaseURL: server.URL})
			require.NoError(t, err)

			// Execute the test method
			tt.testMethod(t, client)
		})
	}
}
