package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Additional coverage tests to reach 100% on background_jobs.go
// Focuses on specialized job creation methods and ToQuery methods

// List tests with nil options and comprehensive query parameter coverage

func TestBackgroundJobsService_List_NilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/background-jobs", r.URL.Path)

		// Verify no query parameters when opts is nil
		assert.Empty(t, r.URL.Query().Get("page"))
		assert.Empty(t, r.URL.Query().Get("limit"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []*BackgroundJob{},
			"meta": PaginationMeta{Page: 1, Limit: 25},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	jobs, meta, err := client.BackgroundJobs.List(context.Background(), nil)
	assert.NoError(t, err)
	assert.NotNil(t, jobs)
	assert.NotNil(t, meta)
}

func TestBackgroundJobsService_List_AllFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)

		q := r.URL.Query()
		assert.Equal(t, "2", q.Get("page"))
		assert.Equal(t, "50", q.Get("limit"))
		assert.Equal(t, "data_export", q.Get("type"))
		assert.Equal(t, "completed", q.Get("status"))
		assert.Equal(t, "123", q.Get("user_id"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []*BackgroundJob{},
			"meta": PaginationMeta{Page: 2, Limit: 50},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	jobs, meta, err := client.BackgroundJobs.List(context.Background(), &ListJobsOptions{
		Page:   2,
		Limit:  50,
		Type:   "data_export",
		Status: "completed",
		UserID: 123,
	})
	assert.NoError(t, err)
	assert.NotNil(t, jobs)
	assert.NotNil(t, meta)
	assert.Equal(t, 2, meta.Page)
}

func TestBackgroundJobsService_List_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Unauthorized",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	jobs, meta, err := client.BackgroundJobs.List(context.Background(), nil)
	assert.Error(t, err)
	assert.Nil(t, jobs)
	assert.Nil(t, meta)
}

// Specialized job creation method tests

func TestBackgroundJobsService_CreateDataExportJob_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/background-jobs", r.URL.Path)

		var body CreateBackgroundJobRequest
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.NoError(t, err)
		assert.Equal(t, "data_export", body.Type)
		assert.Equal(t, 2, body.Priority)
		assert.NotNil(t, body.Payload)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id":     1,
				"type":   "data_export",
				"status": "pending",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	job, resp, err := client.BackgroundJobs.CreateDataExportJob(
		context.Background(),
		1,
		"csv",
		[]string{"servers", "alerts"},
	)
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.NotNil(t, resp)
}

func TestBackgroundJobsService_CreateDataExportJob_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Invalid export format",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	job, _, err := client.BackgroundJobs.CreateDataExportJob(
		context.Background(),
		1,
		"invalid",
		[]string{"servers"},
	)
	assert.Error(t, err)
	assert.Nil(t, job)
}

func TestBackgroundJobsService_CreateReportGenerationJob_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)

		var body CreateBackgroundJobRequest
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.NoError(t, err)
		assert.Equal(t, "report_generation", body.Type)
		assert.Equal(t, 2, body.Priority)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id":     2,
				"type":   "report_generation",
				"status": "pending",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	job, resp, err := client.BackgroundJobs.CreateReportGenerationJob(
		context.Background(),
		1,
		"uptime",
		"monthly",
		[]uint{1, 2, 3},
	)
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.NotNil(t, resp)
}

func TestBackgroundJobsService_CreateReportGenerationJob_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Invalid report type",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	job, _, err := client.BackgroundJobs.CreateReportGenerationJob(
		context.Background(),
		1,
		"invalid",
		"monthly",
		[]uint{1},
	)
	assert.Error(t, err)
	assert.Nil(t, job)
}

func TestBackgroundJobsService_CreateAlertDigestJob_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)

		var body CreateBackgroundJobRequest
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.NoError(t, err)
		assert.Equal(t, "alert_digest", body.Type)
		assert.Equal(t, 1, body.Priority)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id":     3,
				"type":   "alert_digest",
				"status": "pending",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	job, resp, err := client.BackgroundJobs.CreateAlertDigestJob(
		context.Background(),
		1,
		"daily",
		[]string{"admin@example.com"},
	)
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.NotNil(t, resp)
}

func TestBackgroundJobsService_CreateAlertDigestJob_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Invalid recipient emails",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	job, _, err := client.BackgroundJobs.CreateAlertDigestJob(
		context.Background(),
		1,
		"daily",
		[]string{},
	)
	assert.Error(t, err)
	assert.Nil(t, job)
}

// Cancel tests

func TestBackgroundJobsService_Cancel_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/background-jobs/1/cancel", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Job cancelled",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	resp, err := client.BackgroundJobs.Cancel(context.Background(), 1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestBackgroundJobsService_Cancel_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Cannot cancel completed job",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	resp, err := client.BackgroundJobs.Cancel(context.Background(), 1)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

// GetPendingJobs tests with comprehensive options coverage

func TestBackgroundJobsService_GetPendingJobs_NilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/background-jobs/pending", r.URL.Path)

		// Verify no query parameters when opts is nil
		assert.Empty(t, r.URL.Query().Get("limit"))
		assert.Empty(t, r.URL.Query().Get("immediate_only"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []*BackgroundJob{},
			"meta": PaginationMeta{},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	jobs, meta, err := client.BackgroundJobs.GetPendingJobs(context.Background(), nil)
	assert.NoError(t, err)
	assert.NotNil(t, jobs)
	assert.NotNil(t, meta)
}

func TestBackgroundJobsService_GetPendingJobs_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)

		q := r.URL.Query()
		assert.Equal(t, "10", q.Get("limit"))
		assert.Equal(t, "true", q.Get("immediate_only"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []*BackgroundJob{
				{ID: 1, Type: "data_export", Status: "pending"},
			},
			"meta": PaginationMeta{Limit: 10},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	jobs, meta, err := client.BackgroundJobs.GetPendingJobs(context.Background(), &GetPendingJobsOptions{
		Limit:         10,
		ImmediateOnly: true,
	})
	assert.NoError(t, err)
	assert.NotNil(t, jobs)
	assert.NotNil(t, meta)
	assert.Len(t, jobs, 1)
}

func TestBackgroundJobsService_GetPendingJobs_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Insufficient permissions",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	jobs, meta, err := client.BackgroundJobs.GetPendingJobs(context.Background(), nil)
	assert.Error(t, err)
	assert.Nil(t, jobs)
	assert.Nil(t, meta)
}

// Success path tests for complete coverage

func TestBackgroundJobsService_CreateJob_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id":     1,
				"type":   "custom_job",
				"status": "pending",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	job, resp, err := client.BackgroundJobs.CreateJob(context.Background(), &CreateBackgroundJobRequest{
		Type:     "custom_job",
		Priority: 2,
	})
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.NotNil(t, resp)
}

func TestBackgroundJobsService_Get_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/background-jobs/1", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id":     1,
				"status": "running",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	job, resp, err := client.BackgroundJobs.Get(context.Background(), 1)
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.NotNil(t, resp)
}

func TestBackgroundJobsService_Retry_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id":     1,
				"status": "pending",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	job, err := client.BackgroundJobs.Retry(context.Background(), "1")
	assert.NoError(t, err)
	assert.NotNil(t, job)
}

func TestBackgroundJobsService_GetStatus_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id":       "1",
				"status":   "running",
				"progress": 50,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	status, err := client.BackgroundJobs.GetStatus(context.Background(), "1")
	assert.NoError(t, err)
	assert.NotNil(t, status)
}

func TestBackgroundJobsService_UpdateJobStatus_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id":     1,
				"status": "running",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	job, resp, err := client.BackgroundJobs.UpdateJobStatus(context.Background(), 1, &BackgroundJobStatusUpdateRequest{
		Status: "running",
	})
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.NotNil(t, resp)
}

func TestBackgroundJobsService_UpdateJobProgress_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id":       1,
				"progress": 75,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	job, resp, err := client.BackgroundJobs.UpdateJobProgress(context.Background(), 1, &BackgroundJobProgressUpdateRequest{
		Progress:     75,
		ProgressText: "Processing...",
	})
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.NotNil(t, resp)
}

func TestBackgroundJobsService_CompleteJob_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id":     1,
				"status": "completed",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	job, resp, err := client.BackgroundJobs.CompleteJob(context.Background(), 1, map[string]interface{}{
		"records_processed": 1000,
	})
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.NotNil(t, resp)
}

func TestBackgroundJobsService_FailJob_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id":     1,
				"status": "failed",
				"error":  "Processing error",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	job, resp, err := client.BackgroundJobs.FailJob(context.Background(), 1, "Processing error")
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.NotNil(t, resp)
}
