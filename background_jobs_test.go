package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackgroundJobsService_CreateJob(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/background-jobs", r.URL.Path)

		var req CreateBackgroundJobRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "data_export", req.Type)
		assert.Equal(t, 2, req.Priority)
		assert.NotNil(t, req.Payload)

		response := map[string]interface{}{
			"success": true,
			"data": BackgroundJob{
				ID:       1,
				Type:     req.Type,
				Status:   "pending",
				Priority: req.Priority,
				Payload:  req.Payload,
			},
			"message": "Background job created successfully",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})
	require.NoError(t, err)

	// Test creating a job
	ctx := context.Background()
	job, resp, err := client.BackgroundJobs.CreateJob(ctx, &CreateBackgroundJobRequest{
		Type:     "data_export",
		Priority: 2,
		Payload: map[string]interface{}{
			"organization_id": 1,
			"export_format":   "json",
			"data_types":      []string{"servers", "users"},
		},
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, job)
	assert.Equal(t, uint(1), job.ID)
	assert.Equal(t, "data_export", job.Type)
	assert.Equal(t, "pending", job.Status)
}

func TestBackgroundJobsService_ListJobs(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/background-jobs", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "10", r.URL.Query().Get("limit"))
		assert.Equal(t, "data_export", r.URL.Query().Get("type"))

		jobs := []*BackgroundJob{
			{
				ID:       1,
				Type:     "data_export",
				Status:   "completed",
				Priority: 2,
			},
			{
				ID:       2,
				Type:     "data_export",
				Status:   "running",
				Priority: 3,
			},
		}

		response := map[string]interface{}{
			"success": true,
			"data":    jobs,
			"meta": map[string]interface{}{
				"total":       2,
				"page":        1,
				"limit":       10,
				"total_pages": 1,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})
	require.NoError(t, err)

	// Test listing jobs
	ctx := context.Background()
	jobs, resp, err := client.BackgroundJobs.List(ctx, &ListJobsOptions{
		Page:  1,
		Limit: 10,
		Type:  "data_export",
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, jobs)
	assert.Len(t, jobs, 2)
	assert.Equal(t, uint(1), jobs[0].ID)
	assert.Equal(t, "data_export", jobs[0].Type)
}

func TestBackgroundJobsService_GetJobByID(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/background-jobs/123", r.URL.Path)

		response := map[string]interface{}{
			"success": true,
			"data": BackgroundJob{
				ID:           123,
				Type:         "report_generation",
				Status:       "running",
				Priority:     2,
				Progress:     50,
				ProgressText: "Generating report",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})
	require.NoError(t, err)

	// Test getting a job
	ctx := context.Background()
	job, resp, err := client.BackgroundJobs.Get(ctx, 123)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, job)
	assert.Equal(t, uint(123), job.ID)
	assert.Equal(t, "report_generation", job.Type)
	assert.Equal(t, "running", job.Status)
	assert.Equal(t, 50, job.Progress)
	assert.Equal(t, "Generating report", job.ProgressText)
}

func TestBackgroundJobsService_CancelJob(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/background-jobs/456/cancel", r.URL.Path)

		response := map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"job_id": 456,
				"status": "cancelled",
			},
			"message": "Background job cancelled successfully",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})
	require.NoError(t, err)

	// Test cancelling a job
	ctx := context.Background()
	resp, err := client.BackgroundJobs.Cancel(ctx, 456)

	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestBackgroundJobsService_ConvenienceMethods(t *testing.T) {
	callCount := 0
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/background-jobs", r.URL.Path)

		var req CreateBackgroundJobRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		// Verify the request based on job type
		switch callCount {
		case 0: // Data export
			assert.Equal(t, "data_export", req.Type)
			assert.Equal(t, "json", req.Payload["export_format"])
		case 1: // Report generation
			assert.Equal(t, "report_generation", req.Type)
			assert.Equal(t, "uptime", req.Payload["report_type"])
		case 2: // Alert digest
			assert.Equal(t, "alert_digest", req.Type)
			assert.Equal(t, "daily", req.Payload["period"])
		}

		callCount++

		response := map[string]interface{}{
			"success": true,
			"data": BackgroundJob{
				ID:       uint(callCount),
				Type:     req.Type,
				Status:   "pending",
				Priority: req.Priority,
				Payload:  req.Payload,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL:    server.URL,
		HTTPClient: &http.Client{},
	})
	require.NoError(t, err)

	ctx := context.Background()

	// Test data export job
	job1, _, err := client.BackgroundJobs.CreateDataExportJob(ctx, 1, "json", []string{"servers", "users"})
	require.NoError(t, err)
	assert.Equal(t, "data_export", job1.Type)

	// Test report generation job
	job2, _, err := client.BackgroundJobs.CreateReportGenerationJob(ctx, 1, "uptime", "monthly", []uint{1, 2, 3})
	require.NoError(t, err)
	assert.Equal(t, "report_generation", job2.Type)

	// Test alert digest job
	job3, _, err := client.BackgroundJobs.CreateAlertDigestJob(ctx, 1, "daily", []string{"admin@example.com"})
	require.NoError(t, err)
	assert.Equal(t, "alert_digest", job3.Type)

	assert.Equal(t, 3, callCount)
}
