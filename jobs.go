package nexmonyx

import (
	"context"
	"fmt"
	"time"
)

// ControllerJob represents a job managed by the job-controller
type ControllerJob struct {
	ID             string                 `json:"id"`              // UUID
	Name           string                 `json:"name"`            // Human-readable name
	Type           string                 `json:"type"`            // Job type: script, api_call, data_process, etc.
	Status         string                 `json:"status"`          // pending, queued, running, completed, failed, retrying, cancelled, dlq
	Priority       int                    `json:"priority"`        // 1-10, lower is higher priority
	Description    string                 `json:"description,omitempty"`
	TimeoutSeconds int                    `json:"timeout_seconds"` // Execution timeout
	MaxRetries     int                    `json:"max_retries"`     // Maximum retry attempts
	RetryCount     int                    `json:"retry_count"`     // Current retry count
	RetryPolicy    *RetryPolicy           `json:"retry_policy,omitempty"`
	Payload        map[string]interface{} `json:"payload,omitempty"`
	Result         map[string]interface{} `json:"result,omitempty"`
	ScheduleID     string                 `json:"schedule_id,omitempty"`
	ScheduledAt    *time.Time             `json:"scheduled_at,omitempty"`
	QueuedAt       *time.Time             `json:"queued_at,omitempty"`
	StartedAt      *time.Time             `json:"started_at,omitempty"`
	CompletedAt    *time.Time             `json:"completed_at,omitempty"`
	LastError      string                 `json:"last_error,omitempty"`
	CreatedBy      string                 `json:"created_by,omitempty"`
	Tags           map[string]string      `json:"tags,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// RetryPolicy represents the retry configuration for a job
type RetryPolicy struct {
	Type            string  `json:"type"`               // exponential, linear, fixed, none
	BackoffMs       int     `json:"backoff_ms"`         // Initial backoff in milliseconds
	MaxBackoffMs    int     `json:"max_backoff_ms"`     // Maximum backoff cap
	BackoffMulti    float64 `json:"backoff_multiplier"` // Multiplier for exponential backoff
}

// JobExecution represents a single execution attempt of a job
type JobExecution struct {
	ID            string     `json:"id"`             // Execution UUID
	JobID         string     `json:"job_id"`         // Parent job UUID
	AttemptNumber int        `json:"attempt_number"` // 1-based attempt number
	Status        string     `json:"status"`         // running, completed, failed
	StartedAt     string     `json:"started_at"`
	CompletedAt   string     `json:"completed_at,omitempty"`
	DurationMs    *int64     `json:"duration_ms,omitempty"`
	Success       *bool      `json:"success,omitempty"`
	Output        string     `json:"output,omitempty"`
	ExitCode      *int       `json:"exit_code,omitempty"`
	ErrorMessage  *string    `json:"error_message,omitempty"`
	WorkerID      string     `json:"worker_id,omitempty"`
	Hostname      string     `json:"hostname,omitempty"`
}

// DeadLetterEntry represents a job in the dead letter queue
type DeadLetterEntry struct {
	ID            string                   `json:"id"`             // Entry UUID
	JobID         string                   `json:"job_id"`         // Original job UUID
	JobType       string                   `json:"job_type"`
	JobName       string                   `json:"job_name"`
	FailureReason string                   `json:"failure_reason"`
	LastError     string                   `json:"last_error,omitempty"`
	Payload       map[string]interface{}   `json:"payload,omitempty"`
	RetryHistory  []map[string]interface{} `json:"retry_history,omitempty"`
	CreatedAt     string                   `json:"created_at"`
	ExpiresAt     string                   `json:"expires_at"`
	Resolved      bool                     `json:"resolved"`
}

// JobStatistics represents aggregated job statistics
type JobStatistics struct {
	Summary    JobSummary             `json:"summary"`
	ByType     map[string]TypeStats   `json:"by_type"`
	ByPriority map[string]int         `json:"by_priority"`
}

// JobSummary contains high-level job counts
type JobSummary struct {
	TotalJobs    int `json:"total_jobs"`
	Pending      int `json:"pending"`
	Queued       int `json:"queued"`
	Running      int `json:"running"`
	Completed24h int `json:"completed_24h"`
	Failed24h    int `json:"failed_24h"`
	DLQCount     int `json:"dlq_count"`
}

// TypeStats contains statistics for a job type
type TypeStats struct {
	Total         int     `json:"total"`
	SuccessRate   float64 `json:"success_rate"`
	AvgDurationMs int64   `json:"avg_duration_ms"`
}

// CreateJobRequest represents a request to create a new job
type CreateJobRequest struct {
	Name           string                 `json:"name"`
	Type           string                 `json:"type"`
	Description    string                 `json:"description,omitempty"`
	Priority       int                    `json:"priority,omitempty"`       // Default: 3
	TimeoutSeconds int                    `json:"timeout_seconds,omitempty"` // Default: 3600
	MaxRetries     int                    `json:"max_retries,omitempty"`    // Default: 3
	RetryPolicy    *RetryPolicy           `json:"retry_policy,omitempty"`
	Payload        map[string]interface{} `json:"payload,omitempty"`
	ScheduleID     string                 `json:"schedule_id,omitempty"`
	ScheduledAt    *time.Time             `json:"scheduled_at,omitempty"`
	Tags           map[string]string      `json:"tags,omitempty"`
}

// UpdateJobRequest represents a request to update an existing job
type UpdateJobRequest struct {
	Name           string                 `json:"name,omitempty"`
	Description    string                 `json:"description,omitempty"`
	Priority       *int                   `json:"priority,omitempty"`
	TimeoutSeconds *int                   `json:"timeout_seconds,omitempty"`
	MaxRetries     *int                   `json:"max_retries,omitempty"`
	RetryPolicy    *RetryPolicy           `json:"retry_policy,omitempty"`
	Payload        map[string]interface{} `json:"payload,omitempty"`
	ScheduledAt    *time.Time             `json:"scheduled_at,omitempty"`
	Tags           map[string]string      `json:"tags,omitempty"`
}

// ListJobsOptions represents options for filtering job listings
type ListControllerJobsOptions struct {
	Page         int    `url:"page,omitempty"`
	PageSize     int    `url:"page_size,omitempty"`
	Status       string `url:"status,omitempty"`
	Type         string `url:"type,omitempty"`
	Priority     int    `url:"priority,omitempty"`
	ScheduleID   string `url:"schedule_id,omitempty"`
	CreatedAfter string `url:"created_after,omitempty"`
	CreatedBefore string `url:"created_before,omitempty"`
}

// ToQuery converts ListControllerJobsOptions to query parameters
func (o *ListControllerJobsOptions) ToQuery() map[string]string {
	params := make(map[string]string)
	if o.Page > 0 {
		params["page"] = fmt.Sprintf("%d", o.Page)
	}
	if o.PageSize > 0 {
		params["page_size"] = fmt.Sprintf("%d", o.PageSize)
	}
	if o.Status != "" {
		params["status"] = o.Status
	}
	if o.Type != "" {
		params["type"] = o.Type
	}
	if o.Priority > 0 {
		params["priority"] = fmt.Sprintf("%d", o.Priority)
	}
	if o.ScheduleID != "" {
		params["schedule_id"] = o.ScheduleID
	}
	if o.CreatedAfter != "" {
		params["created_after"] = o.CreatedAfter
	}
	if o.CreatedBefore != "" {
		params["created_before"] = o.CreatedBefore
	}
	return params
}

// ListDeadLetterOptions represents options for filtering dead letter queue listings
type ListDeadLetterOptions struct {
	Page     int  `url:"page,omitempty"`
	PageSize int  `url:"page_size,omitempty"`
	Resolved *bool `url:"resolved,omitempty"`
}

// ToQuery converts ListDeadLetterOptions to query parameters
func (o *ListDeadLetterOptions) ToQuery() map[string]string {
	params := make(map[string]string)
	if o.Page > 0 {
		params["page"] = fmt.Sprintf("%d", o.Page)
	}
	if o.PageSize > 0 {
		params["page_size"] = fmt.Sprintf("%d", o.PageSize)
	}
	if o.Resolved != nil {
		params["resolved"] = fmt.Sprintf("%t", *o.Resolved)
	}
	return params
}

// JobsResponse wraps a single job response
type JobsResponse struct {
	Job ControllerJob `json:"data"`
}

// PaginatedJobsResponse wraps a paginated jobs response
type PaginatedJobsResponse struct {
	Jobs       []ControllerJob `json:"jobs"`
	Pagination PaginationMeta  `json:"pagination"`
}

// PaginatedExecutionsResponse wraps a paginated executions response
type PaginatedExecutionsResponse struct {
	Executions []JobExecution `json:"executions"`
	Pagination PaginationMeta `json:"pagination"`
}

// PaginatedDeadLetterResponse wraps a paginated dead letter queue response
type PaginatedDeadLetterResponse struct {
	Entries    []DeadLetterEntry `json:"entries"`
	Pagination PaginationMeta    `json:"pagination"`
}

// CreateJob creates a new job
// Authentication: JWT Token or Unified API Key required
// Endpoint: POST /v1/jobs
func (s *JobsService) CreateJob(ctx context.Context, req *CreateJobRequest) (*ControllerJob, *Response, error) {
	var resp struct {
		Status  string        `json:"status"`
		Message string        `json:"message"`
		Data    ControllerJob `json:"data"`
	}

	apiResp, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/jobs",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, nil, err
	}

	return &resp.Data, apiResp, nil
}

// ListJobs retrieves a paginated list of jobs
// Authentication: JWT Token or Unified API Key required
// Endpoint: GET /v1/jobs
func (s *JobsService) ListJobs(ctx context.Context, opts *ListControllerJobsOptions) (*PaginatedJobsResponse, *Response, error) {
	var resp struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Jobs       []ControllerJob `json:"jobs"`
			Pagination PaginationMeta  `json:"pagination"`
		} `json:"data"`
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/jobs",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	apiResp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return &PaginatedJobsResponse{
		Jobs:       resp.Data.Jobs,
		Pagination: resp.Data.Pagination,
	}, apiResp, nil
}

// GetJob retrieves a specific job by ID
// Authentication: JWT Token or Unified API Key required
// Endpoint: GET /v1/jobs/{id}
func (s *JobsService) GetJob(ctx context.Context, jobID string) (*ControllerJob, *Response, error) {
	var resp struct {
		Status  string        `json:"status"`
		Message string        `json:"message"`
		Data    ControllerJob `json:"data"`
	}

	apiResp, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/jobs/%s", jobID),
		Result: &resp,
	})
	if err != nil {
		return nil, nil, err
	}

	return &resp.Data, apiResp, nil
}

// UpdateJob updates an existing job
// Authentication: JWT Token or Unified API Key required
// Endpoint: PUT /v1/jobs/{id}
func (s *JobsService) UpdateJob(ctx context.Context, jobID string, req *UpdateJobRequest) (*ControllerJob, *Response, error) {
	var resp struct {
		Status  string        `json:"status"`
		Message string        `json:"message"`
		Data    ControllerJob `json:"data"`
	}

	apiResp, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/jobs/%s", jobID),
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, nil, err
	}

	return &resp.Data, apiResp, nil
}

// DeleteJob deletes or cancels a job
// Authentication: JWT Token or Unified API Key required
// Endpoint: DELETE /v1/jobs/{id}
func (s *JobsService) DeleteJob(ctx context.Context, jobID string) (*Response, error) {
	var resp struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	apiResp, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v1/jobs/%s", jobID),
		Result: &resp,
	})
	return apiResp, err
}

// CancelJob cancels a running or queued job
// Authentication: JWT Token or Unified API Key required
// Endpoint: POST /v1/jobs/{id}/cancel
func (s *JobsService) CancelJob(ctx context.Context, jobID string) (*ControllerJob, *Response, error) {
	var resp struct {
		Status  string        `json:"status"`
		Message string        `json:"message"`
		Data    ControllerJob `json:"data"`
	}

	apiResp, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/jobs/%s/cancel", jobID),
		Result: &resp,
	})
	if err != nil {
		return nil, nil, err
	}

	return &resp.Data, apiResp, nil
}

// RetryJob retries a failed, cancelled, or DLQ job
// Authentication: JWT Token or Unified API Key required
// Endpoint: POST /v1/jobs/{id}/retry
func (s *JobsService) RetryJob(ctx context.Context, jobID string) (*ControllerJob, *Response, error) {
	var resp struct {
		Status  string        `json:"status"`
		Message string        `json:"message"`
		Data    ControllerJob `json:"data"`
	}

	apiResp, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/jobs/%s/retry", jobID),
		Result: &resp,
	})
	if err != nil {
		return nil, nil, err
	}

	return &resp.Data, apiResp, nil
}

// GetJobExecutions retrieves the execution history for a job
// Authentication: JWT Token or Unified API Key required
// Endpoint: GET /v1/jobs/{id}/executions
func (s *JobsService) GetJobExecutions(ctx context.Context, jobID string, page, pageSize int) (*PaginatedExecutionsResponse, *Response, error) {
	var resp struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Executions []JobExecution `json:"executions"`
			Pagination PaginationMeta `json:"pagination"`
		} `json:"data"`
	}

	query := make(map[string]string)
	if page > 0 {
		query["page"] = fmt.Sprintf("%d", page)
	}
	if pageSize > 0 {
		query["page_size"] = fmt.Sprintf("%d", pageSize)
	}

	req := &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/jobs/%s/executions", jobID),
		Result: &resp,
	}
	if len(query) > 0 {
		req.Query = query
	}

	apiResp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return &PaginatedExecutionsResponse{
		Executions: resp.Data.Executions,
		Pagination: resp.Data.Pagination,
	}, apiResp, nil
}

// ListDeadLetterQueue retrieves the dead letter queue entries
// Authentication: JWT Token or Unified API Key required
// Endpoint: GET /v1/jobs/deadletter
func (s *JobsService) ListDeadLetterQueue(ctx context.Context, opts *ListDeadLetterOptions) (*PaginatedDeadLetterResponse, *Response, error) {
	var resp struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Entries    []DeadLetterEntry `json:"entries"`
			Pagination PaginationMeta    `json:"pagination"`
		} `json:"data"`
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/jobs/deadletter",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	apiResp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return &PaginatedDeadLetterResponse{
		Entries:    resp.Data.Entries,
		Pagination: resp.Data.Pagination,
	}, apiResp, nil
}

// RetryDeadLetterEntry creates a new job from a dead letter queue entry
// Authentication: JWT Token or Unified API Key required
// Endpoint: POST /v1/jobs/deadletter/{id}/retry
func (s *JobsService) RetryDeadLetterEntry(ctx context.Context, entryID string) (*ControllerJob, *Response, error) {
	var resp struct {
		Status  string        `json:"status"`
		Message string        `json:"message"`
		Data    ControllerJob `json:"data"`
	}

	apiResp, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/jobs/deadletter/%s/retry", entryID),
		Result: &resp,
	})
	if err != nil {
		return nil, nil, err
	}

	return &resp.Data, apiResp, nil
}

// GetJobStatistics retrieves aggregated job statistics
// Authentication: JWT Token or Unified API Key required
// Endpoint: GET /v1/jobs/statistics
func (s *JobsService) GetJobStatistics(ctx context.Context) (*JobStatistics, *Response, error) {
	var resp struct {
		Status  string        `json:"status"`
		Message string        `json:"message"`
		Data    JobStatistics `json:"data"`
	}

	apiResp, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/jobs/statistics",
		Result: &resp,
	})
	if err != nil {
		return nil, nil, err
	}

	return &resp.Data, apiResp, nil
}

// IsComplete returns true if the job is in a terminal state
func (j *ControllerJob) IsComplete() bool {
	return j.Status == "completed" || j.Status == "failed" || j.Status == "cancelled" || j.Status == "dlq"
}

// IsRunning returns true if the job is currently running
func (j *ControllerJob) IsRunning() bool {
	return j.Status == "running"
}

// IsFailed returns true if the job failed
func (j *ControllerJob) IsFailed() bool {
	return j.Status == "failed"
}

// CanRetry returns true if the job can be retried
func (j *ControllerJob) CanRetry() bool {
	return j.Status == "failed" || j.Status == "cancelled" || j.Status == "dlq"
}

// CanCancel returns true if the job can be cancelled
func (j *ControllerJob) CanCancel() bool {
	return j.Status == "pending" || j.Status == "queued" || j.Status == "running" || j.Status == "retrying"
}

// =============================================================================
// Job Types API (Task #3906)
// =============================================================================

// JobType represents a job type in the registry
type JobType struct {
	ID                 string `json:"id,omitempty"`
	Name               string `json:"name"`
	DisplayName        string `json:"display_name"`
	Description        string `json:"description,omitempty"`
	DefaultTimeoutSecs int    `json:"default_timeout_seconds"`
	DefaultMaxRetries  int    `json:"default_max_retries"`
	DefaultPriority    int    `json:"default_priority"`
	MaxConcurrent      int    `json:"max_concurrent"`
	IsSystem           bool   `json:"is_system"`
	IsEnabled          bool   `json:"is_enabled"`
	CreatedAt          string `json:"created_at,omitempty"`
}

// CreateJobTypeRequest represents a request to create a custom job type
type CreateJobTypeRequest struct {
	Name               string `json:"name"`
	DisplayName        string `json:"display_name"`
	Description        string `json:"description,omitempty"`
	DefaultTimeoutSecs int    `json:"default_timeout_seconds,omitempty"`
	DefaultMaxRetries  int    `json:"default_max_retries,omitempty"`
	DefaultPriority    int    `json:"default_priority,omitempty"`
	MaxConcurrent      int    `json:"max_concurrent,omitempty"`
}

// ListJobTypesOptions represents options for listing job types
type ListJobTypesOptions struct {
	Page          int  `url:"page,omitempty"`
	PageSize      int  `url:"page_size,omitempty"`
	IncludeSystem bool `url:"include_system,omitempty"`
	EnabledOnly   bool `url:"enabled_only,omitempty"`
}

// ToQuery converts ListJobTypesOptions to query parameters
func (o *ListJobTypesOptions) ToQuery() map[string]string {
	params := make(map[string]string)
	if o.Page > 0 {
		params["page"] = fmt.Sprintf("%d", o.Page)
	}
	if o.PageSize > 0 {
		params["page_size"] = fmt.Sprintf("%d", o.PageSize)
	}
	if o.IncludeSystem {
		params["include_system"] = "true"
	}
	if o.EnabledOnly {
		params["enabled_only"] = "true"
	}
	return params
}

// PaginatedJobTypesResponse wraps a paginated job types response
type PaginatedJobTypesResponse struct {
	JobTypes   []JobType      `json:"job_types"`
	Pagination PaginationMeta `json:"pagination"`
}

// ListJobTypes retrieves all available job types
// Authentication: JWT Token or Unified API Key required
// Endpoint: GET /v1/jobs/types
func (s *JobsService) ListJobTypes(ctx context.Context, opts *ListJobTypesOptions) (*PaginatedJobTypesResponse, *Response, error) {
	var resp struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    struct {
			JobTypes   []JobType      `json:"job_types"`
			Pagination PaginationMeta `json:"pagination"`
		} `json:"data"`
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/jobs/types",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	apiResp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return &PaginatedJobTypesResponse{
		JobTypes:   resp.Data.JobTypes,
		Pagination: resp.Data.Pagination,
	}, apiResp, nil
}

// CreateJobType creates a new custom job type
// Authentication: JWT Token or Unified API Key required
// Endpoint: POST /v1/jobs/types
func (s *JobsService) CreateJobType(ctx context.Context, req *CreateJobTypeRequest) (*JobType, *Response, error) {
	var resp struct {
		Status  string  `json:"status"`
		Message string  `json:"message"`
		Data    JobType `json:"data"`
	}

	apiResp, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/jobs/types",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, nil, err
	}

	return &resp.Data, apiResp, nil
}

// GetJobType retrieves a specific job type by name
// Authentication: JWT Token or Unified API Key required
// Endpoint: GET /v1/jobs/types/{type}
func (s *JobsService) GetJobType(ctx context.Context, typeName string) (*JobType, *Response, error) {
	var resp struct {
		Status  string  `json:"status"`
		Message string  `json:"message"`
		Data    JobType `json:"data"`
	}

	apiResp, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/jobs/types/%s", typeName),
		Result: &resp,
	})
	if err != nil {
		return nil, nil, err
	}

	return &resp.Data, apiResp, nil
}

// =============================================================================
// Job Templates API (Task #3906)
// =============================================================================

// JobTemplate represents a reusable job template
type JobTemplate struct {
	ID                 string                 `json:"id"`
	OrganizationID     uint                   `json:"organization_id"`
	Name               string                 `json:"name"`
	Description        string                 `json:"description,omitempty"`
	JobType            string                 `json:"job_type"`
	DefaultPayload     map[string]interface{} `json:"default_payload,omitempty"`
	DefaultPriority    int                    `json:"default_priority"`
	DefaultTimeoutSecs int                    `json:"default_timeout_seconds"`
	DefaultMaxRetries  int                    `json:"default_max_retries"`
	Parameters         []TemplateParam        `json:"parameters,omitempty"`
	Tags               map[string]string      `json:"tags,omitempty"`
	UsageCount         int                    `json:"usage_count"`
	IsActive           bool                   `json:"is_active"`
	CreatedBy          string                 `json:"created_by"`
	CreatedAt          string                 `json:"created_at"`
	UpdatedAt          string                 `json:"updated_at"`
}

// TemplateParam represents a configurable parameter in a job template
type TemplateParam struct {
	Name         string      `json:"name"`
	Description  string      `json:"description,omitempty"`
	Type         string      `json:"type"` // string, int, bool, json
	Required     bool        `json:"required"`
	DefaultValue interface{} `json:"default_value,omitempty"`
}

// CreateJobTemplateRequest represents a request to create a job template
type CreateJobTemplateRequest struct {
	Name               string                 `json:"name"`
	Description        string                 `json:"description,omitempty"`
	JobType            string                 `json:"job_type"`
	DefaultPayload     map[string]interface{} `json:"default_payload,omitempty"`
	DefaultPriority    int                    `json:"default_priority,omitempty"`
	DefaultTimeoutSecs int                    `json:"default_timeout_seconds,omitempty"`
	DefaultMaxRetries  int                    `json:"default_max_retries,omitempty"`
	Parameters         []TemplateParam        `json:"parameters,omitempty"`
	Tags               map[string]string      `json:"tags,omitempty"`
}

// UpdateJobTemplateRequest represents a request to update a job template
type UpdateJobTemplateRequest struct {
	Name               *string                `json:"name,omitempty"`
	Description        *string                `json:"description,omitempty"`
	DefaultPayload     map[string]interface{} `json:"default_payload,omitempty"`
	DefaultPriority    *int                   `json:"default_priority,omitempty"`
	DefaultTimeoutSecs *int                   `json:"default_timeout_seconds,omitempty"`
	DefaultMaxRetries  *int                   `json:"default_max_retries,omitempty"`
	Parameters         []TemplateParam        `json:"parameters,omitempty"`
	Tags               map[string]string      `json:"tags,omitempty"`
	IsActive           *bool                  `json:"is_active,omitempty"`
}

// CreateJobFromTemplateRequest represents a request to create a job from a template
type CreateJobFromTemplateRequest struct {
	Name             string                 `json:"name"`
	PayloadOverrides map[string]interface{} `json:"payload_overrides,omitempty"`
	PriorityOverride *int                   `json:"priority_override,omitempty"`
	ScheduledAt      *time.Time             `json:"scheduled_at,omitempty"`
}

// ListJobTemplatesOptions represents options for listing job templates
type ListJobTemplatesOptions struct {
	Page       int    `url:"page,omitempty"`
	PageSize   int    `url:"page_size,omitempty"`
	JobType    string `url:"job_type,omitempty"`
	ActiveOnly bool   `url:"active_only,omitempty"`
	SortBy     string `url:"sort_by,omitempty"`
	SortOrder  string `url:"sort_order,omitempty"`
}

// ToQuery converts ListJobTemplatesOptions to query parameters
func (o *ListJobTemplatesOptions) ToQuery() map[string]string {
	params := make(map[string]string)
	if o.Page > 0 {
		params["page"] = fmt.Sprintf("%d", o.Page)
	}
	if o.PageSize > 0 {
		params["page_size"] = fmt.Sprintf("%d", o.PageSize)
	}
	if o.JobType != "" {
		params["job_type"] = o.JobType
	}
	if o.ActiveOnly {
		params["active_only"] = "true"
	}
	if o.SortBy != "" {
		params["sort_by"] = o.SortBy
	}
	if o.SortOrder != "" {
		params["sort_order"] = o.SortOrder
	}
	return params
}

// PaginatedJobTemplatesResponse wraps a paginated job templates response
type PaginatedJobTemplatesResponse struct {
	Templates  []JobTemplate  `json:"templates"`
	Pagination PaginationMeta `json:"pagination"`
}

// ListJobTemplates retrieves a paginated list of job templates
// Authentication: JWT Token or Unified API Key required
// Endpoint: GET /v1/jobs/templates
func (s *JobsService) ListJobTemplates(ctx context.Context, opts *ListJobTemplatesOptions) (*PaginatedJobTemplatesResponse, *Response, error) {
	var resp struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Templates  []JobTemplate  `json:"templates"`
			Pagination PaginationMeta `json:"pagination"`
		} `json:"data"`
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/jobs/templates",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	apiResp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return &PaginatedJobTemplatesResponse{
		Templates:  resp.Data.Templates,
		Pagination: resp.Data.Pagination,
	}, apiResp, nil
}

// CreateJobTemplate creates a new job template
// Authentication: JWT Token or Unified API Key required
// Endpoint: POST /v1/jobs/templates
func (s *JobsService) CreateJobTemplate(ctx context.Context, req *CreateJobTemplateRequest) (*JobTemplate, *Response, error) {
	var resp struct {
		Status  string      `json:"status"`
		Message string      `json:"message"`
		Data    JobTemplate `json:"data"`
	}

	apiResp, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/jobs/templates",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, nil, err
	}

	return &resp.Data, apiResp, nil
}

// GetJobTemplate retrieves a specific job template by ID
// Authentication: JWT Token or Unified API Key required
// Endpoint: GET /v1/jobs/templates/{id}
func (s *JobsService) GetJobTemplate(ctx context.Context, templateID string) (*JobTemplate, *Response, error) {
	var resp struct {
		Status  string      `json:"status"`
		Message string      `json:"message"`
		Data    JobTemplate `json:"data"`
	}

	apiResp, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/jobs/templates/%s", templateID),
		Result: &resp,
	})
	if err != nil {
		return nil, nil, err
	}

	return &resp.Data, apiResp, nil
}

// UpdateJobTemplate updates an existing job template
// Authentication: JWT Token or Unified API Key required
// Endpoint: PUT /v1/jobs/templates/{id}
func (s *JobsService) UpdateJobTemplate(ctx context.Context, templateID string, req *UpdateJobTemplateRequest) (*JobTemplate, *Response, error) {
	var resp struct {
		Status  string      `json:"status"`
		Message string      `json:"message"`
		Data    JobTemplate `json:"data"`
	}

	apiResp, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/jobs/templates/%s", templateID),
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, nil, err
	}

	return &resp.Data, apiResp, nil
}

// DeleteJobTemplate deletes a job template
// Authentication: JWT Token or Unified API Key required
// Endpoint: DELETE /v1/jobs/templates/{id}
func (s *JobsService) DeleteJobTemplate(ctx context.Context, templateID string) (*Response, error) {
	var resp struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	apiResp, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v1/jobs/templates/%s", templateID),
		Result: &resp,
	})
	return apiResp, err
}

// CreateJobFromTemplate creates a new job from a template
// Authentication: JWT Token or Unified API Key required
// Endpoint: POST /v1/jobs/from-template/{id}
func (s *JobsService) CreateJobFromTemplate(ctx context.Context, templateID string, req *CreateJobFromTemplateRequest) (*ControllerJob, *Response, error) {
	var resp struct {
		Status  string        `json:"status"`
		Message string        `json:"message"`
		Data    ControllerJob `json:"data"`
	}

	apiResp, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/jobs/from-template/%s", templateID),
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, nil, err
	}

	return &resp.Data, apiResp, nil
}
