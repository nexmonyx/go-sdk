package nexmonyx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAlertsService_TypeAssertionErrors tests the "unexpected response type" error paths
// for methods at 87.5% coverage. These tests cover the type assertion fallback that occurs
// when the API returns malformed data.
func TestAlertsService_TypeAssertionErrors(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(w http.ResponseWriter, r *http.Request)
		testMethod func(t *testing.T, client *Client)
	}{
		{
			name: "Create - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				// Return success but with null data, which will cause type assertion to fail
				// The JSON unmarshals successfully but resp.Data ends up as nil instead of *Alert
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","message":"created","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				alert := &Alert{
					Name:        "Test Alert",
					MetricName:  "cpu_usage",
					Condition:   "greater_than",
					Threshold:   80.0,
					Severity:    "warning",
					Description: "CPU usage alert",
				}
				result, err := client.Alerts.Create(context.Background(), alert)
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "unexpected response type")
			},
		},
		{
			name: "Get - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				result, err := client.Alerts.Get(context.Background(), "alert-123")
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
				alert := &Alert{Name: "Updated Alert"}
				result, err := client.Alerts.Update(context.Background(), "alert-123", alert)
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "unexpected response type")
			},
		},
		{
			name: "Enable - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				result, err := client.Alerts.Enable(context.Background(), "alert-123")
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "unexpected response type")
			},
		},
		{
			name: "Disable - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				result, err := client.Alerts.Disable(context.Background(), "alert-123")
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "unexpected response type")
			},
		},
		{
			name: "Test - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				result, err := client.Alerts.Test(context.Background(), "alert-123")
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "unexpected response type")
			},
		},
		{
			name: "CreateChannel - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				channel := &AlertChannel{
					Name: "Test Channel",
					Type: "email",
					Configuration: map[string]interface{}{
						"recipients": []string{"test@example.com"},
					},
				}
				result, err := client.Alerts.CreateChannel(context.Background(), channel)
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "unexpected response type")
			},
		},
		{
			name: "GetChannel - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				result, err := client.Alerts.GetChannel(context.Background(), "channel-123")
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "unexpected response type")
			},
		},
		{
			name: "UpdateChannel - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				channel := &AlertChannel{Name: "Updated Channel"}
				result, err := client.Alerts.UpdateChannel(context.Background(), "channel-123", channel)
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "unexpected response type")
			},
		},
		{
			name: "TestChannel - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				result, err := client.Alerts.TestChannel(context.Background(), "channel-123")
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
