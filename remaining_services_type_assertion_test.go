package nexmonyx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRemainingServices_TypeAssertionErrors tests the "unexpected response type" error paths
// for the remaining service methods at 87.5% coverage
func TestRemainingServices_TypeAssertionErrors(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(w http.ResponseWriter, r *http.Request)
		testMethod func(t *testing.T, client *Client)
	}{
		// agent_versions.go
		{
			name: "AgentVersions.CreateVersion - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				req := &AgentVersionRequest{
					Version:      "1.0.0",
					ReleaseNotes: "Test version",
				}
				result, err := client.AgentVersions.CreateVersion(context.Background(), req)
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "unexpected response type")
			},
		},
		{
			name: "AgentVersions.GetVersion - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				result, err := client.AgentVersions.GetVersion(context.Background(), "1.0.0")
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "unexpected response type")
			},
		},
		// api_keys.go
		{
			name: "APIKeys.GetUnified - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				result, err := client.APIKeys.GetUnified(context.Background(), "key-123")
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "unexpected response type")
			},
		},
		{
			name: "APIKeys.UpdateUnified - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				name := "Updated Key"
				req := &UpdateUnifiedAPIKeyRequest{
					Name: &name,
				}
				result, err := client.APIKeys.UpdateUnified(context.Background(), "key-123", req)
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "unexpected response type")
			},
		},
		// billing.go
		{
			name: "Billing.GetBillingInfo - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				result, err := client.Billing.GetBillingInfo(context.Background(), "org-uuid")
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "unexpected response type")
			},
		},
		{
			name: "Billing.GetSubscription - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				result, err := client.Billing.GetSubscription(context.Background(), "org-uuid")
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "unexpected response type")
			},
		},
		// billing_usage.go
		{
			name: "BillingUsage.GetMyCurrentUsage - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				result, err := client.BillingUsage.GetMyCurrentUsage(context.Background())
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "unexpected response type")
			},
		},
		{
			name: "BillingUsage.GetMyUsageSummary - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				result, err := client.BillingUsage.GetMyUsageSummary(context.Background(), time.Time{}, time.Time{})
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "unexpected response type")
			},
		},
		// controllers.go
		{
			name: "Controllers.GetControllerStatus - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				result, err := client.Controllers.GetControllerStatus(context.Background(), "us-east-1")
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "unexpected response type")
			},
		},
		// health.go
		{
			name: "Health.GetHealth - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				result, err := client.Health.GetHealth(context.Background())
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
