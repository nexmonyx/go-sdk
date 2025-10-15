package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Helper method tests - IsComplete, IsRunning, IsFailed

func TestBackgroundJob_IsComplete(t *testing.T) {
	completedJob := &BackgroundJob{Status: "completed"}
	assert.True(t, completedJob.IsComplete())

	cancelledJob := &BackgroundJob{Status: "cancelled"}
	assert.True(t, cancelledJob.IsComplete())

	failedJob := &BackgroundJob{Status: "failed"}
	assert.True(t, failedJob.IsComplete())

	pendingJob := &BackgroundJob{Status: "pending"}
	assert.False(t, pendingJob.IsComplete())

	runningJob := &BackgroundJob{Status: "running"}
	assert.False(t, runningJob.IsComplete())
}

func TestBackgroundJob_IsRunning(t *testing.T) {
	runningJob := &BackgroundJob{Status: "running"}
	assert.True(t, runningJob.IsRunning())

	completedJob := &BackgroundJob{Status: "completed"}
	assert.False(t, completedJob.IsRunning())

	pendingJob := &BackgroundJob{Status: "pending"}
	assert.False(t, pendingJob.IsRunning())

	failedJob := &BackgroundJob{Status: "failed"}
	assert.False(t, failedJob.IsRunning())
}

func TestBackgroundJob_IsFailed(t *testing.T) {
	failedJob := &BackgroundJob{Status: "failed"}
	assert.True(t, failedJob.IsFailed())

	completedJob := &BackgroundJob{Status: "completed"}
	assert.False(t, completedJob.IsFailed())

	runningJob := &BackgroundJob{Status: "running"}
	assert.False(t, runningJob.IsFailed())

	pendingJob := &BackgroundJob{Status: "pending"}
	assert.False(t, pendingJob.IsFailed())
}

// Error path tests for 75% coverage methods

func TestBackgroundJobsService_CreateJob_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Invalid job type",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	job, _, err := client.BackgroundJobs.CreateJob(context.Background(), &CreateBackgroundJobRequest{
		Type: "invalid",
	})
	assert.Error(t, err)
	assert.Nil(t, job)
}

func TestBackgroundJobsService_Get_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Job not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	job, _, err := client.BackgroundJobs.Get(context.Background(), 999)
	assert.Error(t, err)
	assert.Nil(t, job)
}

func TestBackgroundJobsService_Retry_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Cannot retry running job",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	job, err := client.BackgroundJobs.Retry(context.Background(), "1")
	assert.Error(t, err)
	assert.Nil(t, job)
}

func TestBackgroundJobsService_GetStatus_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Job not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	status, err := client.BackgroundJobs.GetStatus(context.Background(), "999")
	assert.Error(t, err)
	assert.Nil(t, status)
}

func TestBackgroundJobsService_UpdateJobStatus_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Cannot update job status",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	job, _, err := client.BackgroundJobs.UpdateJobStatus(context.Background(), 1, &BackgroundJobStatusUpdateRequest{Status: "running"})
	assert.Error(t, err)
	assert.Nil(t, job)
}

func TestBackgroundJobsService_UpdateJobProgress_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Invalid progress value",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	job, _, err := client.BackgroundJobs.UpdateJobProgress(context.Background(), 1, &BackgroundJobProgressUpdateRequest{Progress: 150})
	assert.Error(t, err)
	assert.Nil(t, job)
}

func TestBackgroundJobsService_CompleteJob_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Job is not running",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	job, _, err := client.BackgroundJobs.CompleteJob(context.Background(), 1, map[string]interface{}{"status": "done"})
	assert.Error(t, err)
	assert.Nil(t, job)
}

func TestBackgroundJobsService_FailJob_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Cannot fail completed job",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	job, _, err := client.BackgroundJobs.FailJob(context.Background(), 1, "Test error")
	assert.Error(t, err)
	assert.Nil(t, job)
}

// These tests cover edge cases that are already well-tested in the existing test file
