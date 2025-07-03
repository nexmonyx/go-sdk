package nexmonyx

import (
	"context"
	"fmt"
)

// CreateJob creates a new background job
func (s *BackgroundJobsService) CreateJob(ctx context.Context, req *CreateBackgroundJobRequest) (*BackgroundJob, *Response, error) {
	var resp StandardResponse
	resp.Data = &BackgroundJob{}

	apiResp, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/api/v1/background-jobs",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, nil, err
	}

	if job, ok := resp.Data.(*BackgroundJob); ok {
		return job, apiResp, nil
	}
	return nil, apiResp, fmt.Errorf("unexpected response type")
}

// Get retrieves a background job by ID
func (s *BackgroundJobsService) Get(ctx context.Context, jobID uint) (*BackgroundJob, *Response, error) {
	var resp StandardResponse
	resp.Data = &BackgroundJob{}

	apiResp, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/api/v1/background-jobs/%d", jobID),
		Result: &resp,
	})
	if err != nil {
		return nil, nil, err
	}

	if job, ok := resp.Data.(*BackgroundJob); ok {
		return job, apiResp, nil
	}
	return nil, apiResp, fmt.Errorf("unexpected response type")
}

// List retrieves a list of background jobs
func (s *BackgroundJobsService) List(ctx context.Context, opts *ListJobsOptions) ([]*BackgroundJob, *PaginationMeta, error) {
	var resp PaginatedResponse
	var jobs []*BackgroundJob
	resp.Data = &jobs

	req := &Request{
		Method: "GET",
		Path:   "/api/v1/background-jobs",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return jobs, resp.Meta, nil
}

// Cancel cancels a background job
func (s *BackgroundJobsService) Cancel(ctx context.Context, jobID uint) (*Response, error) {
	var resp StandardResponse

	apiResp, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/api/v1/background-jobs/%d/cancel", jobID),
		Result: &resp,
	})
	return apiResp, err
}

// Retry retries a failed background job
func (s *BackgroundJobsService) Retry(ctx context.Context, jobID string) (*BackgroundJob, error) {
	var resp StandardResponse
	resp.Data = &BackgroundJob{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/api/v1/background-jobs/%s/retry", jobID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if job, ok := resp.Data.(*BackgroundJob); ok {
		return job, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetStatus retrieves the status of a background job
func (s *BackgroundJobsService) GetStatus(ctx context.Context, jobID string) (*JobStatus, error) {
	var resp StandardResponse
	resp.Data = &JobStatus{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/api/v1/background-jobs/%s/status", jobID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if status, ok := resp.Data.(*JobStatus); ok {
		return status, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// BackgroundJob represents a background job
type BackgroundJob struct {
	ID             uint                   `json:"id"`
	Type           string                 `json:"type"`
	Status         string                 `json:"status"` // pending, running, completed, failed, cancelled
	Progress       int                    `json:"progress"` // 0-100
	Priority       int                    `json:"priority"` // 1 (low), 2 (normal), 3 (high)
	ProgressText   string                 `json:"progress_text,omitempty"`
	OrganizationID uint                   `json:"organization_id"`
	UserID         uint                   `json:"user_id"`
	CreatedAt      *CustomTime            `json:"created_at"`
	StartedAt      *CustomTime            `json:"started_at,omitempty"`
	CompletedAt    *CustomTime            `json:"completed_at,omitempty"`
	FailedAt       *CustomTime            `json:"failed_at,omitempty"`
	RetryCount     int                    `json:"retry_count"`
	MaxRetries     int                    `json:"max_retries"`
	Payload        map[string]interface{} `json:"payload,omitempty"`
	Result         map[string]interface{} `json:"result,omitempty"`
	Error          string                 `json:"error,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// CreateBackgroundJobRequest represents a request to create a background job
type CreateBackgroundJobRequest struct {
	Type       string                 `json:"type"`
	Priority   int                    `json:"priority"` // 1 (low), 2 (normal), 3 (high)
	Payload    map[string]interface{} `json:"payload,omitempty"`
	MaxRetries int                    `json:"max_retries,omitempty"`
	ScheduleAt *CustomTime            `json:"schedule_at,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ListJobsOptions represents options for listing background jobs
type ListJobsOptions struct {
	Page   int    `url:"page,omitempty"`
	Limit  int    `url:"limit,omitempty"`
	Type   string `url:"type,omitempty"`
	Status string `url:"status,omitempty"`
	UserID uint   `url:"user_id,omitempty"`
}

// ToQuery converts ListJobsOptions to query parameters
func (o *ListJobsOptions) ToQuery() map[string]string {
	params := make(map[string]string)
	if o.Page > 0 {
		params["page"] = fmt.Sprintf("%d", o.Page)
	}
	if o.Limit > 0 {
		params["limit"] = fmt.Sprintf("%d", o.Limit)
	}
	if o.Type != "" {
		params["type"] = o.Type
	}
	if o.Status != "" {
		params["status"] = o.Status
	}
	if o.UserID > 0 {
		params["user_id"] = fmt.Sprintf("%d", o.UserID)
	}
	return params
}

// JobStatus represents the status of a background job
type JobStatus struct {
	ID         string                 `json:"id"`
	Status     string                 `json:"status"`
	Progress   int                    `json:"progress"`
	Message    string                 `json:"message,omitempty"`
	Steps      []JobStep              `json:"steps,omitempty"`
	Metrics    map[string]interface{} `json:"metrics,omitempty"`
	UpdatedAt  *CustomTime            `json:"updated_at"`
}

// JobStep represents a step in a background job
type JobStep struct {
	Name      string      `json:"name"`
	Status    string      `json:"status"`
	StartedAt *CustomTime `json:"started_at"`
	EndedAt   *CustomTime `json:"ended_at,omitempty"`
	Duration  float64     `json:"duration,omitempty"` // seconds
	Error     string      `json:"error,omitempty"`
}

// IsComplete returns true if the job is complete (succeeded or failed)
func (j *BackgroundJob) IsComplete() bool {
	return j.Status == "completed" || j.Status == "failed" || j.Status == "cancelled"
}

// IsRunning returns true if the job is currently running
func (j *BackgroundJob) IsRunning() bool {
	return j.Status == "running"
}

// IsFailed returns true if the job failed
func (j *BackgroundJob) IsFailed() bool {
	return j.Status == "failed"
}

// CreateDataExportJob creates a data export background job
func (s *BackgroundJobsService) CreateDataExportJob(ctx context.Context, organizationID uint, exportFormat string, dataTypes []string) (*BackgroundJob, *Response, error) {
	return s.CreateJob(ctx, &CreateBackgroundJobRequest{
		Type:     "data_export",
		Priority: 2,
		Payload: map[string]interface{}{
			"organization_id": organizationID,
			"export_format":   exportFormat,
			"data_types":      dataTypes,
		},
	})
}

// CreateReportGenerationJob creates a report generation background job
func (s *BackgroundJobsService) CreateReportGenerationJob(ctx context.Context, organizationID uint, reportType string, period string, serverIDs []uint) (*BackgroundJob, *Response, error) {
	payload := map[string]interface{}{
		"organization_id": organizationID,
		"report_type":     reportType,
		"period":          period,
		"server_ids":      serverIDs,
	}
	
	return s.CreateJob(ctx, &CreateBackgroundJobRequest{
		Type:     "report_generation",
		Priority: 2,
		Payload:  payload,
	})
}

// CreateAlertDigestJob creates an alert digest background job
func (s *BackgroundJobsService) CreateAlertDigestJob(ctx context.Context, organizationID uint, period string, recipientEmails []string) (*BackgroundJob, *Response, error) {
	return s.CreateJob(ctx, &CreateBackgroundJobRequest{
		Type:     "alert_digest",
		Priority: 1,
		Payload: map[string]interface{}{
			"organization_id":  organizationID,
			"period":           period,
			"recipient_emails": recipientEmails,
		},
	})
}