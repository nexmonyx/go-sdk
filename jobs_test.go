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

// =============================================================================
// Core Job Operations Tests
// =============================================================================

func TestJobsService_CreateJob(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/jobs", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		var reqBody CreateJobRequest
		json.NewDecoder(r.Body).Decode(&reqBody)
		assert.Equal(t, "Daily Backup", reqBody.Name)
		assert.Equal(t, "script", reqBody.Type)
		assert.Equal(t, 2, reqBody.Priority)

		response := struct {
			Status  string        `json:"status"`
			Message string        `json:"message"`
			Data    ControllerJob `json:"data"`
		}{
			Status:  "success",
			Message: "Job created successfully",
			Data: ControllerJob{
				ID:             "job-123",
				Name:           reqBody.Name,
				Type:           reqBody.Type,
				Status:         "pending",
				Priority:       reqBody.Priority,
				TimeoutSeconds: 3600,
				MaxRetries:     3,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
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

	job, _, err := client.Jobs.CreateJob(context.Background(), &CreateJobRequest{
		Name:     "Daily Backup",
		Type:     "script",
		Priority: 2,
		Payload: map[string]interface{}{
			"script": "/usr/local/bin/backup.sh",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "job-123", job.ID)
	assert.Equal(t, "Daily Backup", job.Name)
	assert.Equal(t, "script", job.Type)
	assert.Equal(t, "pending", job.Status)
	assert.Equal(t, 2, job.Priority)
}

func TestJobsService_ListJobs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/jobs", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "20", r.URL.Query().Get("page_size"))
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		response := struct {
			Status  string `json:"status"`
			Message string `json:"message"`
			Data    struct {
				Jobs       []ControllerJob `json:"jobs"`
				Pagination PaginationMeta  `json:"pagination"`
			} `json:"data"`
		}{
			Status:  "success",
			Message: "Jobs retrieved",
			Data: struct {
				Jobs       []ControllerJob `json:"jobs"`
				Pagination PaginationMeta  `json:"pagination"`
			}{
				Jobs: []ControllerJob{
					{
						ID:       "job-1",
						Name:     "Backup Job",
						Type:     "script",
						Status:   "completed",
						Priority: 3,
					},
					{
						ID:       "job-2",
						Name:     "Report Job",
						Type:     "report_generation",
						Status:   "running",
						Priority: 2,
					},
				},
				Pagination: PaginationMeta{
					Page:       1,
					PerPage:    20,
					TotalItems: 2,
					TotalPages: 1,
				},
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

	result, _, err := client.Jobs.ListJobs(context.Background(), &ListControllerJobsOptions{
		Page:     1,
		PageSize: 20,
	})
	require.NoError(t, err)
	assert.Len(t, result.Jobs, 2)
	assert.Equal(t, "job-1", result.Jobs[0].ID)
	assert.Equal(t, "Backup Job", result.Jobs[0].Name)
	assert.Equal(t, "completed", result.Jobs[0].Status)
	assert.Equal(t, 2, result.Pagination.TotalItems)
}

func TestJobsService_ListJobs_WithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/jobs", r.URL.Path)
		assert.Equal(t, "running", r.URL.Query().Get("status"))
		assert.Equal(t, "script", r.URL.Query().Get("type"))
		assert.Equal(t, "2", r.URL.Query().Get("priority"))

		response := struct {
			Status  string `json:"status"`
			Message string `json:"message"`
			Data    struct {
				Jobs       []ControllerJob `json:"jobs"`
				Pagination PaginationMeta  `json:"pagination"`
			} `json:"data"`
		}{
			Status:  "success",
			Message: "Jobs retrieved",
			Data: struct {
				Jobs       []ControllerJob `json:"jobs"`
				Pagination PaginationMeta  `json:"pagination"`
			}{
				Jobs: []ControllerJob{
					{
						ID:       "job-filtered",
						Name:     "Filtered Job",
						Type:     "script",
						Status:   "running",
						Priority: 2,
					},
				},
				Pagination: PaginationMeta{
					TotalItems: 1,
					TotalPages: 1,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	result, _, err := client.Jobs.ListJobs(context.Background(), &ListControllerJobsOptions{
		Status:   "running",
		Type:     "script",
		Priority: 2,
	})
	require.NoError(t, err)
	assert.Len(t, result.Jobs, 1)
	assert.Equal(t, "running", result.Jobs[0].Status)
	assert.Equal(t, "script", result.Jobs[0].Type)
}

func TestJobsService_GetJob(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/jobs/job-123", r.URL.Path)

		response := struct {
			Status  string        `json:"status"`
			Message string        `json:"message"`
			Data    ControllerJob `json:"data"`
		}{
			Status:  "success",
			Message: "Job retrieved",
			Data: ControllerJob{
				ID:             "job-123",
				Name:           "Test Job",
				Type:           "api_call",
				Status:         "completed",
				Priority:       3,
				TimeoutSeconds: 300,
				MaxRetries:     5,
				RetryCount:     0,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	job, _, err := client.Jobs.GetJob(context.Background(), "job-123")
	require.NoError(t, err)
	assert.Equal(t, "job-123", job.ID)
	assert.Equal(t, "Test Job", job.Name)
	assert.Equal(t, "api_call", job.Type)
	assert.Equal(t, "completed", job.Status)
}

func TestJobsService_GetJob_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Job not found",
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	_, _, err = client.Jobs.GetJob(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestJobsService_UpdateJob(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/jobs/job-123", r.URL.Path)

		var reqBody UpdateJobRequest
		json.NewDecoder(r.Body).Decode(&reqBody)
		assert.Equal(t, "Updated Job", reqBody.Name)

		response := struct {
			Status  string        `json:"status"`
			Message string        `json:"message"`
			Data    ControllerJob `json:"data"`
		}{
			Status:  "success",
			Message: "Job updated",
			Data: ControllerJob{
				ID:       "job-123",
				Name:     "Updated Job",
				Type:     "script",
				Status:   "pending",
				Priority: 1,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	priority := 1
	job, _, err := client.Jobs.UpdateJob(context.Background(), "job-123", &UpdateJobRequest{
		Name:     "Updated Job",
		Priority: &priority,
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated Job", job.Name)
	assert.Equal(t, 1, job.Priority)
}

func TestJobsService_DeleteJob(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/jobs/job-123", r.URL.Path)

		response := struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		}{
			Status:  "success",
			Message: "Job deleted",
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

	_, err = client.Jobs.DeleteJob(context.Background(), "job-123")
	require.NoError(t, err)
}

func TestJobsService_CancelJob(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/jobs/job-123/cancel", r.URL.Path)

		response := struct {
			Status  string        `json:"status"`
			Message string        `json:"message"`
			Data    ControllerJob `json:"data"`
		}{
			Status:  "success",
			Message: "Job cancelled",
			Data: ControllerJob{
				ID:     "job-123",
				Name:   "Cancelled Job",
				Type:   "script",
				Status: "cancelled",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	job, _, err := client.Jobs.CancelJob(context.Background(), "job-123")
	require.NoError(t, err)
	assert.Equal(t, "cancelled", job.Status)
}

func TestJobsService_RetryJob(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/jobs/job-123/retry", r.URL.Path)

		response := struct {
			Status  string        `json:"status"`
			Message string        `json:"message"`
			Data    ControllerJob `json:"data"`
		}{
			Status:  "success",
			Message: "Job retry initiated",
			Data: ControllerJob{
				ID:         "job-123",
				Name:       "Retried Job",
				Type:       "script",
				Status:     "queued",
				RetryCount: 1,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	job, _, err := client.Jobs.RetryJob(context.Background(), "job-123")
	require.NoError(t, err)
	assert.Equal(t, "queued", job.Status)
	assert.Equal(t, 1, job.RetryCount)
}

// =============================================================================
// Execution & Statistics Tests
// =============================================================================

func TestJobsService_GetJobExecutions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/jobs/job-123/executions", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "10", r.URL.Query().Get("page_size"))

		durationMs := int64(5000)
		success := true
		response := struct {
			Status  string `json:"status"`
			Message string `json:"message"`
			Data    struct {
				Executions []JobExecution `json:"executions"`
				Pagination PaginationMeta `json:"pagination"`
			} `json:"data"`
		}{
			Status:  "success",
			Message: "Executions retrieved",
			Data: struct {
				Executions []JobExecution `json:"executions"`
				Pagination PaginationMeta `json:"pagination"`
			}{
				Executions: []JobExecution{
					{
						ID:            "exec-1",
						JobID:         "job-123",
						AttemptNumber: 1,
						Status:        "completed",
						StartedAt:     "2024-01-15T10:00:00Z",
						CompletedAt:   "2024-01-15T10:00:05Z",
						DurationMs:    &durationMs,
						Success:       &success,
						WorkerID:      "worker-1",
					},
				},
				Pagination: PaginationMeta{
					TotalItems: 1,
					TotalPages: 1,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	result, _, err := client.Jobs.GetJobExecutions(context.Background(), "job-123", 1, 10)
	require.NoError(t, err)
	assert.Len(t, result.Executions, 1)
	assert.Equal(t, "exec-1", result.Executions[0].ID)
	assert.Equal(t, "completed", result.Executions[0].Status)
	assert.True(t, *result.Executions[0].Success)
}

func TestJobsService_GetJobStatistics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/jobs/statistics", r.URL.Path)

		response := struct {
			Status  string        `json:"status"`
			Message string        `json:"message"`
			Data    JobStatistics `json:"data"`
		}{
			Status:  "success",
			Message: "Statistics retrieved",
			Data: JobStatistics{
				Summary: JobSummary{
					TotalJobs:    100,
					Pending:      5,
					Queued:       10,
					Running:      15,
					Completed24h: 50,
					Failed24h:    5,
					DLQCount:     3,
				},
				ByType: map[string]TypeStats{
					"script": {
						Total:         60,
						SuccessRate:   95.5,
						AvgDurationMs: 5000,
					},
					"api_call": {
						Total:         40,
						SuccessRate:   98.0,
						AvgDurationMs: 1000,
					},
				},
				ByPriority: map[string]int{
					"1": 20,
					"2": 30,
					"3": 50,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	stats, _, err := client.Jobs.GetJobStatistics(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 100, stats.Summary.TotalJobs)
	assert.Equal(t, 15, stats.Summary.Running)
	assert.Equal(t, 95.5, stats.ByType["script"].SuccessRate)
	assert.Equal(t, 50, stats.ByPriority["3"])
}

// =============================================================================
// Dead Letter Queue Tests
// =============================================================================

func TestJobsService_ListDeadLetterQueue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/jobs/deadletter", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "false", r.URL.Query().Get("resolved"))

		response := struct {
			Status  string `json:"status"`
			Message string `json:"message"`
			Data    struct {
				Entries    []DeadLetterEntry `json:"entries"`
				Pagination PaginationMeta    `json:"pagination"`
			} `json:"data"`
		}{
			Status:  "success",
			Message: "Dead letter entries retrieved",
			Data: struct {
				Entries    []DeadLetterEntry `json:"entries"`
				Pagination PaginationMeta    `json:"pagination"`
			}{
				Entries: []DeadLetterEntry{
					{
						ID:            "dlq-1",
						JobID:         "job-456",
						JobType:       "script",
						JobName:       "Failed Script",
						FailureReason: "max_retries_exceeded",
						LastError:     "Connection timeout",
						CreatedAt:     "2024-01-15T10:00:00Z",
						ExpiresAt:     "2024-01-22T10:00:00Z",
						Resolved:      false,
					},
				},
				Pagination: PaginationMeta{
					TotalItems: 1,
					TotalPages: 1,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	resolved := false
	result, _, err := client.Jobs.ListDeadLetterQueue(context.Background(), &ListDeadLetterOptions{
		Page:     1,
		Resolved: &resolved,
	})
	require.NoError(t, err)
	assert.Len(t, result.Entries, 1)
	assert.Equal(t, "dlq-1", result.Entries[0].ID)
	assert.Equal(t, "max_retries_exceeded", result.Entries[0].FailureReason)
	assert.False(t, result.Entries[0].Resolved)
}

func TestJobsService_RetryDeadLetterEntry(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/jobs/deadletter/dlq-1/retry", r.URL.Path)

		response := struct {
			Status  string        `json:"status"`
			Message string        `json:"message"`
			Data    ControllerJob `json:"data"`
		}{
			Status:  "success",
			Message: "DLQ entry retried",
			Data: ControllerJob{
				ID:     "job-new-123",
				Name:   "Retried from DLQ",
				Type:   "script",
				Status: "queued",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	job, _, err := client.Jobs.RetryDeadLetterEntry(context.Background(), "dlq-1")
	require.NoError(t, err)
	assert.Equal(t, "job-new-123", job.ID)
	assert.Equal(t, "queued", job.Status)
}

// =============================================================================
// Job Types Tests
// =============================================================================

func TestJobsService_ListJobTypes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/jobs/types", r.URL.Path)
		assert.Equal(t, "true", r.URL.Query().Get("include_system"))

		response := struct {
			Status  string `json:"status"`
			Message string `json:"message"`
			Data    struct {
				JobTypes   []JobType      `json:"job_types"`
				Pagination PaginationMeta `json:"pagination"`
			} `json:"data"`
		}{
			Status:  "success",
			Message: "Job types retrieved",
			Data: struct {
				JobTypes   []JobType      `json:"job_types"`
				Pagination PaginationMeta `json:"pagination"`
			}{
				JobTypes: []JobType{
					{
						ID:                 "script",
						Name:               "script",
						DisplayName:        "Script Execution",
						Description:        "Run shell scripts",
						DefaultTimeoutSecs: 3600,
						DefaultMaxRetries:  3,
						DefaultPriority:    3,
						MaxConcurrent:      10,
						IsSystem:           true,
						IsEnabled:          true,
					},
					{
						ID:                 "api_call",
						Name:               "api_call",
						DisplayName:        "API Call",
						Description:        "Make HTTP requests",
						DefaultTimeoutSecs: 300,
						DefaultMaxRetries:  5,
						DefaultPriority:    3,
						MaxConcurrent:      50,
						IsSystem:           true,
						IsEnabled:          true,
					},
				},
				Pagination: PaginationMeta{
					TotalItems: 2,
					TotalPages: 1,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	result, _, err := client.Jobs.ListJobTypes(context.Background(), &ListJobTypesOptions{
		IncludeSystem: true,
	})
	require.NoError(t, err)
	assert.Len(t, result.JobTypes, 2)
	assert.Equal(t, "script", result.JobTypes[0].Name)
	assert.True(t, result.JobTypes[0].IsSystem)
}

func TestJobsService_CreateJobType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/jobs/types", r.URL.Path)

		var reqBody CreateJobTypeRequest
		json.NewDecoder(r.Body).Decode(&reqBody)
		assert.Equal(t, "custom_type", reqBody.Name)

		response := struct {
			Status  string  `json:"status"`
			Message string  `json:"message"`
			Data    JobType `json:"data"`
		}{
			Status:  "success",
			Message: "Job type created",
			Data: JobType{
				ID:                 "custom_type",
				Name:               "custom_type",
				DisplayName:        "Custom Type",
				Description:        "My custom job type",
				DefaultTimeoutSecs: 600,
				DefaultMaxRetries:  3,
				DefaultPriority:    3,
				MaxConcurrent:      5,
				IsSystem:           false,
				IsEnabled:          true,
			},
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

	jobType, _, err := client.Jobs.CreateJobType(context.Background(), &CreateJobTypeRequest{
		Name:               "custom_type",
		DisplayName:        "Custom Type",
		Description:        "My custom job type",
		DefaultTimeoutSecs: 600,
		MaxConcurrent:      5,
	})
	require.NoError(t, err)
	assert.Equal(t, "custom_type", jobType.Name)
	assert.False(t, jobType.IsSystem)
	assert.True(t, jobType.IsEnabled)
}

func TestJobsService_GetJobType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/jobs/types/script", r.URL.Path)

		response := struct {
			Status  string  `json:"status"`
			Message string  `json:"message"`
			Data    JobType `json:"data"`
		}{
			Status:  "success",
			Message: "Job type retrieved",
			Data: JobType{
				ID:                 "script",
				Name:               "script",
				DisplayName:        "Script Execution",
				DefaultTimeoutSecs: 3600,
				DefaultMaxRetries:  3,
				IsSystem:           true,
				IsEnabled:          true,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	jobType, _, err := client.Jobs.GetJobType(context.Background(), "script")
	require.NoError(t, err)
	assert.Equal(t, "script", jobType.Name)
	assert.Equal(t, 3600, jobType.DefaultTimeoutSecs)
	assert.True(t, jobType.IsSystem)
}

// =============================================================================
// Job Templates Tests
// =============================================================================

func TestJobsService_ListJobTemplates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/jobs/templates", r.URL.Path)
		assert.Equal(t, "true", r.URL.Query().Get("active_only"))

		response := struct {
			Status  string `json:"status"`
			Message string `json:"message"`
			Data    struct {
				Templates  []JobTemplate  `json:"templates"`
				Pagination PaginationMeta `json:"pagination"`
			} `json:"data"`
		}{
			Status:  "success",
			Message: "Templates retrieved",
			Data: struct {
				Templates  []JobTemplate  `json:"templates"`
				Pagination PaginationMeta `json:"pagination"`
			}{
				Templates: []JobTemplate{
					{
						ID:              "tmpl-1",
						Name:            "Daily Backup Template",
						Description:     "Template for daily backups",
						JobType:         "script",
						DefaultPriority: 2,
						UsageCount:      50,
						IsActive:        true,
						CreatedBy:       "admin",
					},
				},
				Pagination: PaginationMeta{
					TotalItems: 1,
					TotalPages: 1,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	result, _, err := client.Jobs.ListJobTemplates(context.Background(), &ListJobTemplatesOptions{
		ActiveOnly: true,
	})
	require.NoError(t, err)
	assert.Len(t, result.Templates, 1)
	assert.Equal(t, "Daily Backup Template", result.Templates[0].Name)
	assert.True(t, result.Templates[0].IsActive)
}

func TestJobsService_CreateJobTemplate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/jobs/templates", r.URL.Path)

		var reqBody CreateJobTemplateRequest
		json.NewDecoder(r.Body).Decode(&reqBody)
		assert.Equal(t, "Backup Template", reqBody.Name)
		assert.Equal(t, "script", reqBody.JobType)

		response := struct {
			Status  string      `json:"status"`
			Message string      `json:"message"`
			Data    JobTemplate `json:"data"`
		}{
			Status:  "success",
			Message: "Template created",
			Data: JobTemplate{
				ID:              "tmpl-new",
				Name:            reqBody.Name,
				Description:     reqBody.Description,
				JobType:         reqBody.JobType,
				DefaultPriority: 3,
				UsageCount:      0,
				IsActive:        true,
			},
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

	template, _, err := client.Jobs.CreateJobTemplate(context.Background(), &CreateJobTemplateRequest{
		Name:        "Backup Template",
		Description: "Template for backups",
		JobType:     "script",
		Parameters: []TemplateParam{
			{Name: "target", Type: "string", Required: true},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "tmpl-new", template.ID)
	assert.Equal(t, "Backup Template", template.Name)
	assert.True(t, template.IsActive)
}

func TestJobsService_GetJobTemplate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/jobs/templates/tmpl-123", r.URL.Path)

		response := struct {
			Status  string      `json:"status"`
			Message string      `json:"message"`
			Data    JobTemplate `json:"data"`
		}{
			Status:  "success",
			Message: "Template retrieved",
			Data: JobTemplate{
				ID:          "tmpl-123",
				Name:        "Test Template",
				JobType:     "script",
				UsageCount:  25,
				IsActive:    true,
				CreatedBy:   "admin",
				Parameters: []TemplateParam{
					{Name: "target", Type: "string", Required: true},
					{Name: "timeout", Type: "int", Required: false, DefaultValue: 300},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	template, _, err := client.Jobs.GetJobTemplate(context.Background(), "tmpl-123")
	require.NoError(t, err)
	assert.Equal(t, "tmpl-123", template.ID)
	assert.Len(t, template.Parameters, 2)
	assert.Equal(t, "target", template.Parameters[0].Name)
	assert.True(t, template.Parameters[0].Required)
}

func TestJobsService_UpdateJobTemplate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/jobs/templates/tmpl-123", r.URL.Path)

		response := struct {
			Status  string      `json:"status"`
			Message string      `json:"message"`
			Data    JobTemplate `json:"data"`
		}{
			Status:  "success",
			Message: "Template updated",
			Data: JobTemplate{
				ID:          "tmpl-123",
				Name:        "Updated Template",
				Description: "Updated description",
				IsActive:    true,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	name := "Updated Template"
	description := "Updated description"
	template, _, err := client.Jobs.UpdateJobTemplate(context.Background(), "tmpl-123", &UpdateJobTemplateRequest{
		Name:        &name,
		Description: &description,
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated Template", template.Name)
}

func TestJobsService_DeleteJobTemplate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/jobs/templates/tmpl-123", r.URL.Path)

		response := struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		}{
			Status:  "success",
			Message: "Template deleted",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	_, err = client.Jobs.DeleteJobTemplate(context.Background(), "tmpl-123")
	require.NoError(t, err)
}

func TestJobsService_CreateJobFromTemplate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/jobs/from-template/tmpl-123", r.URL.Path)

		var reqBody CreateJobFromTemplateRequest
		json.NewDecoder(r.Body).Decode(&reqBody)
		assert.Equal(t, "Job from Template", reqBody.Name)

		response := struct {
			Status  string        `json:"status"`
			Message string        `json:"message"`
			Data    ControllerJob `json:"data"`
		}{
			Status:  "success",
			Message: "Job created from template",
			Data: ControllerJob{
				ID:       "job-from-tmpl",
				Name:     reqBody.Name,
				Type:     "script",
				Status:   "pending",
				Priority: 3,
			},
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

	job, _, err := client.Jobs.CreateJobFromTemplate(context.Background(), "tmpl-123", &CreateJobFromTemplateRequest{
		Name: "Job from Template",
		PayloadOverrides: map[string]interface{}{
			"target": "/backup/destination",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "job-from-tmpl", job.ID)
	assert.Equal(t, "Job from Template", job.Name)
	assert.Equal(t, "pending", job.Status)
}

// =============================================================================
// Helper Method Tests
// =============================================================================

func TestControllerJob_IsComplete(t *testing.T) {
	tests := []struct {
		status   string
		expected bool
	}{
		{"completed", true},
		{"failed", true},
		{"cancelled", true},
		{"dlq", true},
		{"pending", false},
		{"queued", false},
		{"running", false},
		{"retrying", false},
	}

	for _, tc := range tests {
		job := &ControllerJob{Status: tc.status}
		assert.Equal(t, tc.expected, job.IsComplete(), "status: %s", tc.status)
	}
}

func TestControllerJob_IsRunning(t *testing.T) {
	tests := []struct {
		status   string
		expected bool
	}{
		{"running", true},
		{"pending", false},
		{"queued", false},
		{"completed", false},
		{"failed", false},
	}

	for _, tc := range tests {
		job := &ControllerJob{Status: tc.status}
		assert.Equal(t, tc.expected, job.IsRunning(), "status: %s", tc.status)
	}
}

func TestControllerJob_IsFailed(t *testing.T) {
	tests := []struct {
		status   string
		expected bool
	}{
		{"failed", true},
		{"completed", false},
		{"running", false},
		{"dlq", false},
	}

	for _, tc := range tests {
		job := &ControllerJob{Status: tc.status}
		assert.Equal(t, tc.expected, job.IsFailed(), "status: %s", tc.status)
	}
}

func TestControllerJob_CanRetry(t *testing.T) {
	tests := []struct {
		status   string
		expected bool
	}{
		{"failed", true},
		{"cancelled", true},
		{"dlq", true},
		{"completed", false},
		{"running", false},
		{"pending", false},
	}

	for _, tc := range tests {
		job := &ControllerJob{Status: tc.status}
		assert.Equal(t, tc.expected, job.CanRetry(), "status: %s", tc.status)
	}
}

func TestControllerJob_CanCancel(t *testing.T) {
	tests := []struct {
		status   string
		expected bool
	}{
		{"pending", true},
		{"queued", true},
		{"running", true},
		{"retrying", true},
		{"completed", false},
		{"failed", false},
		{"cancelled", false},
		{"dlq", false},
	}

	for _, tc := range tests {
		job := &ControllerJob{Status: tc.status}
		assert.Equal(t, tc.expected, job.CanCancel(), "status: %s", tc.status)
	}
}

// =============================================================================
// Options ToQuery Tests
// =============================================================================

func TestListControllerJobsOptions_ToQuery(t *testing.T) {
	opts := &ListControllerJobsOptions{
		Page:          2,
		PageSize:      50,
		Status:        "running",
		Type:          "script",
		Priority:      1,
		ScheduleID:    "sched-123",
		CreatedAfter:  "2024-01-01",
		CreatedBefore: "2024-01-31",
	}

	query := opts.ToQuery()
	assert.Equal(t, "2", query["page"])
	assert.Equal(t, "50", query["page_size"])
	assert.Equal(t, "running", query["status"])
	assert.Equal(t, "script", query["type"])
	assert.Equal(t, "1", query["priority"])
	assert.Equal(t, "sched-123", query["schedule_id"])
	assert.Equal(t, "2024-01-01", query["created_after"])
	assert.Equal(t, "2024-01-31", query["created_before"])
}

func TestListDeadLetterOptions_ToQuery(t *testing.T) {
	resolved := true
	opts := &ListDeadLetterOptions{
		Page:     1,
		PageSize: 20,
		Resolved: &resolved,
	}

	query := opts.ToQuery()
	assert.Equal(t, "1", query["page"])
	assert.Equal(t, "20", query["page_size"])
	assert.Equal(t, "true", query["resolved"])
}

func TestListJobTypesOptions_ToQuery(t *testing.T) {
	opts := &ListJobTypesOptions{
		Page:          1,
		PageSize:      10,
		IncludeSystem: true,
		EnabledOnly:   true,
	}

	query := opts.ToQuery()
	assert.Equal(t, "1", query["page"])
	assert.Equal(t, "10", query["page_size"])
	assert.Equal(t, "true", query["include_system"])
	assert.Equal(t, "true", query["enabled_only"])
}

func TestListJobTemplatesOptions_ToQuery(t *testing.T) {
	opts := &ListJobTemplatesOptions{
		Page:       1,
		PageSize:   25,
		JobType:    "script",
		ActiveOnly: true,
		SortBy:     "created_at",
		SortOrder:  "desc",
	}

	query := opts.ToQuery()
	assert.Equal(t, "1", query["page"])
	assert.Equal(t, "25", query["page_size"])
	assert.Equal(t, "script", query["job_type"])
	assert.Equal(t, "true", query["active_only"])
	assert.Equal(t, "created_at", query["sort_by"])
	assert.Equal(t, "desc", query["sort_order"])
}

// =============================================================================
// Error Handling Tests
// =============================================================================

func TestJobsService_CreateJob_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Invalid or missing authentication",
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "invalid-token"},
	})
	require.NoError(t, err)

	_, _, err = client.Jobs.CreateJob(context.Background(), &CreateJobRequest{
		Name: "Test",
		Type: "script",
	})
	assert.Error(t, err)
}

func TestJobsService_CreateJob_ValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Validation failed: name is required",
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	_, _, err = client.Jobs.CreateJob(context.Background(), &CreateJobRequest{
		Type: "script", // Missing name
	})
	assert.Error(t, err)
}

func TestJobsService_CancelJob_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Job cannot be cancelled in current state",
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	_, _, err = client.Jobs.CancelJob(context.Background(), "job-completed")
	assert.Error(t, err)
}

func TestJobsService_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Internal server error",
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	_, _, err = client.Jobs.GetJob(context.Background(), "job-123")
	assert.Error(t, err)
}
