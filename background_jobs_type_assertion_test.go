package nexmonyx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBackgroundJobsService_TypeAssertionErrors tests the "unexpected response type" error paths
// for methods at 87.5% coverage in background_jobs.go
func TestBackgroundJobsService_TypeAssertionErrors(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(w http.ResponseWriter, r *http.Request)
		testMethod func(t *testing.T, client *Client)
	}{
		{
			name: "CreateJob - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				req := &CreateBackgroundJobRequest{
					Type:     "test_job",
					Priority: 2,
				}
				result, _, err := client.BackgroundJobs.CreateJob(context.Background(), req)
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
				result, _, err := client.BackgroundJobs.Get(context.Background(), 123)
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "unexpected response type")
			},
		},
		{
			name: "Retry - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				result, err := client.BackgroundJobs.Retry(context.Background(), "job-123")
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "unexpected response type")
			},
		},
		{
			name: "GetStatus - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				result, err := client.BackgroundJobs.GetStatus(context.Background(), "job-123")
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "unexpected response type")
			},
		},
		{
			name: "UpdateJobStatus - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				req := &BackgroundJobStatusUpdateRequest{
					Status: "running",
				}
				result, _, err := client.BackgroundJobs.UpdateJobStatus(context.Background(), 123, req)
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "unexpected response type")
			},
		},
		{
			name: "UpdateJobProgress - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				req := &BackgroundJobProgressUpdateRequest{
					Progress:     50,
					ProgressText: "Half done",
				}
				result, _, err := client.BackgroundJobs.UpdateJobProgress(context.Background(), 123, req)
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "unexpected response type")
			},
		},
		{
			name: "CompleteJob - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				result, _, err := client.BackgroundJobs.CompleteJob(context.Background(), 123, map[string]interface{}{
					"success": true,
				})
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "unexpected response type")
			},
		},
		{
			name: "FailJob - unexpected response type (nil data)",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":null}`))
			},
			testMethod: func(t *testing.T, client *Client) {
				result, _, err := client.BackgroundJobs.FailJob(context.Background(), 123, "Test error")
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
