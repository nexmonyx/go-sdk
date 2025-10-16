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

// TestBackgroundJobsService_CreateJobComprehensive tests the CreateJob method
func TestBackgroundJobsService_CreateJobComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		request    *CreateBackgroundJobRequest
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *BackgroundJob)
	}{
		{
			name: "success - create export job",
			request: &CreateBackgroundJobRequest{
				Type:     "data_export",
				Priority: 2,
				Payload: map[string]interface{}{
					"format": "json",
					"types":  []string{"servers", "alerts"},
				},
			},
			mockStatus: http.StatusCreated,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":              1,
					"type":            "data_export",
					"status":          "pending",
					"progress":        0,
					"priority":        2,
					"organization_id": 1,
					"user_id":         1,
					"retry_count":     0,
					"max_retries":     3,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, job *BackgroundJob) {
				assert.Equal(t, uint(1), job.ID)
				assert.Equal(t, "data_export", job.Type)
				assert.Equal(t, "pending", job.Status)
			},
		},
		{
			name: "validation error - missing type",
			request: &CreateBackgroundJobRequest{
				Priority: 2,
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Job type is required",
			},
			wantErr: true,
		},
		{
			name: "unauthorized",
			request: &CreateBackgroundJobRequest{
				Type: "data_export",
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
			request: &CreateBackgroundJobRequest{
				Type: "data_export",
			},
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

			result, _, err := client.BackgroundJobs.CreateJob(ctx, tt.request)

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

// TestBackgroundJobsService_GetComprehensive tests the Get method
func TestBackgroundJobsService_GetComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		jobID      uint
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *BackgroundJob)
	}{
		{
			name:       "success - get running job",
			jobID:      1,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":              1,
					"type":            "data_export",
					"status":          "running",
					"progress":        45,
					"progress_text":   "Processing servers data",
					"organization_id": 1,
					"user_id":         1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, job *BackgroundJob) {
				assert.Equal(t, uint(1), job.ID)
				assert.Equal(t, "running", job.Status)
				assert.Equal(t, 45, job.Progress)
			},
		},
		{
			name:       "not found",
			jobID:      999,
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Job not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			jobID:      1,
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			jobID:      1,
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

			result, _, err := client.BackgroundJobs.Get(ctx, tt.jobID)

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

// TestBackgroundJobsService_ListComprehensive tests the List method
func TestBackgroundJobsService_ListComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		opts       *ListJobsOptions
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, []*BackgroundJob, *PaginationMeta)
	}{
		{
			name: "success - list all jobs",
			opts: &ListJobsOptions{Page: 1, Limit: 25},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": []map[string]interface{}{
					{
						"id":     1,
						"type":   "data_export",
						"status": "completed",
					},
					{
						"id":     2,
						"type":   "report_generation",
						"status": "running",
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
			checkFunc: func(t *testing.T, jobs []*BackgroundJob, meta *PaginationMeta) {
				assert.Len(t, jobs, 2)
				assert.Equal(t, 2, meta.TotalItems)
			},
		},
		{
			name:       "success - empty results",
			opts:       &ListJobsOptions{Page: 1, Limit: 25},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data":   []map[string]interface{}{},
				"meta": map[string]interface{}{
					"page":        1,
					"limit":       25,
					"total_items": 0,
					"total_pages": 0,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, jobs []*BackgroundJob, meta *PaginationMeta) {
				assert.Len(t, jobs, 0)
				assert.Equal(t, 0, meta.TotalItems)
			},
		},
		{
			name:       "unauthorized",
			opts:       &ListJobsOptions{Page: 1, Limit: 25},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			opts:       &ListJobsOptions{Page: 1, Limit: 25},
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

			jobs, meta, err := client.BackgroundJobs.List(ctx, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, jobs)
				if tt.checkFunc != nil {
					tt.checkFunc(t, jobs, meta)
				}
			}
		})
	}
}

// TestBackgroundJobsService_CancelComprehensive tests the Cancel method
func TestBackgroundJobsService_CancelComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		jobID      uint
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:       "success - cancel job",
			jobID:      1,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status":  "success",
				"message": "Job cancelled successfully",
			},
			wantErr: false,
		},
		{
			name:       "not found",
			jobID:      999,
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Job not found",
			},
			wantErr: true,
		},
		{
			name:       "conflict - already completed",
			jobID:      1,
			mockStatus: http.StatusConflict,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Cannot cancel completed job",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			jobID:      1,
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			jobID:      1,
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

			_, err = client.BackgroundJobs.Cancel(ctx, tt.jobID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestBackgroundJobsService_GetStatusComprehensive tests the GetStatus method
func TestBackgroundJobsService_GetStatusComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		jobID      string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *JobStatus)
	}{
		{
			name:       "success - running job",
			jobID:      "job-123",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"status":        "running",
					"progress":      65,
					"progress_text": "Processing alerts data",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, status *JobStatus) {
				assert.Equal(t, "running", status.Status)
				assert.Equal(t, 65, status.Progress)
			},
		},
		{
			name:       "success - completed job",
			jobID:      "job-456",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"status":   "completed",
					"progress": 100,
				},
			},
			wantErr: false,
		},
		{
			name:       "not found",
			jobID:      "non-existent",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Job not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			jobID:      "job-123",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			jobID:      "job-123",
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

			result, err := client.BackgroundJobs.GetStatus(ctx, tt.jobID)

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
