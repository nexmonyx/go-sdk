package nexmonyx

import (
	"context"
	"fmt"
)

// SchedulesService handles schedule management API operations
// Schedules are used by the scheduler-controller to automate recurring tasks
type SchedulesService struct {
	client *Client
}

// ScheduleTargetType represents the type of target for a schedule
type ScheduleTargetType string

const (
	ScheduleTargetJob          ScheduleTargetType = "job"
	ScheduleTargetReport       ScheduleTargetType = "report"
	ScheduleTargetMaintenance  ScheduleTargetType = "maintenance"
	ScheduleTargetNotification ScheduleTargetType = "notification"
	ScheduleTargetCleanup      ScheduleTargetType = "cleanup"
	ScheduleTargetHealthCheck  ScheduleTargetType = "health_check"
	ScheduleTargetCustom       ScheduleTargetType = "custom"
)

// ScheduleRetryPolicy represents the retry policy for a schedule
type ScheduleRetryPolicy string

const (
	ScheduleRetryExponential ScheduleRetryPolicy = "exponential"
	ScheduleRetryLinear      ScheduleRetryPolicy = "linear"
	ScheduleRetryFixed       ScheduleRetryPolicy = "fixed"
)

// ScheduleStatus represents the status of a schedule
type ScheduleStatus string

const (
	ScheduleStatusActive     ScheduleStatus = "active"
	ScheduleStatusPaused     ScheduleStatus = "paused"
	ScheduleStatusCompleted  ScheduleStatus = "completed"
	ScheduleStatusError      ScheduleStatus = "error"
)

// ScheduleExecutionStatus represents the status of a schedule execution
type ScheduleExecutionStatus string

const (
	ScheduleExecutionPending   ScheduleExecutionStatus = "pending"
	ScheduleExecutionRunning   ScheduleExecutionStatus = "running"
	ScheduleExecutionCompleted ScheduleExecutionStatus = "completed"
	ScheduleExecutionFailed    ScheduleExecutionStatus = "failed"
	ScheduleExecutionRetrying  ScheduleExecutionStatus = "retrying"
	ScheduleExecutionCancelled ScheduleExecutionStatus = "cancelled"
	ScheduleExecutionTimedOut  ScheduleExecutionStatus = "timed_out"
)

// Schedule represents a scheduler-controller schedule
type Schedule struct {
	ID             uint                   `json:"id"`
	ScheduleUUID   string                 `json:"schedule_uuid"`
	OrganizationID uint                   `json:"organization_id"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description,omitempty"`
	CronExpression string                 `json:"cron_expression"`
	Timezone       string                 `json:"timezone"`
	TargetType     ScheduleTargetType     `json:"target_type"`
	TargetConfig   map[string]interface{} `json:"target_config"`
	Enabled        bool                   `json:"enabled"`
	MaxRetries     int                    `json:"max_retries"`
	RetryPolicy    ScheduleRetryPolicy    `json:"retry_policy"`
	TimeoutMinutes int                    `json:"timeout_minutes"`
	Status         ScheduleStatus         `json:"status"`
	NextRunAt      *string                `json:"next_run_at,omitempty"`
	LastRunAt      *string                `json:"last_run_at,omitempty"`
	LastRunStatus  *string                `json:"last_run_status,omitempty"`
	LastRunError   *string                `json:"last_run_error,omitempty"`
	RunCount       int64                  `json:"run_count"`
	SuccessCount   int64                  `json:"success_count"`
	FailureCount   int64                  `json:"failure_count"`
	CreatedAt      string                 `json:"created_at"`
	UpdatedAt      string                 `json:"updated_at"`
	CreatedByID    *uint                  `json:"created_by_id,omitempty"`
	CreatedByEmail *string                `json:"created_by_email,omitempty"`
	UpdatedByID    *uint                  `json:"updated_by_id,omitempty"`
	UpdatedByEmail *string                `json:"updated_by_email,omitempty"`
}

// ScheduleExecution represents a single execution of a schedule
type ScheduleExecution struct {
	ID                 uint                    `json:"id"`
	ExecutionUUID      string                  `json:"execution_uuid"`
	ScheduleID         uint                    `json:"schedule_id"`
	OrganizationID     uint                    `json:"organization_id"`
	TriggerTime        string                  `json:"trigger_time"`
	ScheduledTime      string                  `json:"scheduled_time"`
	JitterAppliedMs    int                     `json:"jitter_applied_ms"`
	Status             ScheduleExecutionStatus `json:"status"`
	StartedAt          *string                 `json:"started_at,omitempty"`
	CompletedAt        *string                 `json:"completed_at,omitempty"`
	DurationMs         *int                    `json:"duration_ms,omitempty"`
	TargetResourceID   *string                 `json:"target_resource_id,omitempty"`
	TargetResourceType *string                 `json:"target_resource_type,omitempty"`
	RetryCount         int                     `json:"retry_count"`
	ErrorMessage       *string                 `json:"error_message,omitempty"`
	ErrorDetails       map[string]interface{}  `json:"error_details,omitempty"`
	Result             map[string]interface{}  `json:"result,omitempty"`
	ManualTrigger      bool                    `json:"manual_trigger"`
	ManualTriggerReason *string                `json:"manual_trigger_reason,omitempty"`
	TriggeredByID      *uint                   `json:"triggered_by_id,omitempty"`
	TriggeredByEmail   *string                 `json:"triggered_by_email,omitempty"`
	CreatedAt          string                  `json:"created_at"`
}

// ScheduleStatistics represents aggregated statistics for a schedule
type ScheduleStatistics struct {
	ScheduleID           uint    `json:"schedule_id"`
	Period               string  `json:"period"`
	TotalExecutions      int64   `json:"total_executions"`
	SuccessfulExecutions int64   `json:"successful_executions"`
	FailedExecutions     int64   `json:"failed_executions"`
	SuccessRate          float64 `json:"success_rate"`
	AvgDurationMs        int64   `json:"avg_duration_ms"`
	MinDurationMs        int64   `json:"min_duration_ms"`
	MaxDurationMs        int64   `json:"max_duration_ms"`
	LastExecutionAt      *string `json:"last_execution_at,omitempty"`
	NextExecutionAt      *string `json:"next_execution_at,omitempty"`
}

// NextRunPreview represents a preview of a scheduled run time
type NextRunPreview struct {
	RunTime   string `json:"run_time"`
	RunNumber int    `json:"run_number"`
}

// NextRunsResponse represents the response for next runs preview
type NextRunsResponse struct {
	ScheduleID     uint             `json:"schedule_id"`
	CronExpression string           `json:"cron_expression"`
	Timezone       string           `json:"timezone"`
	NextRuns       []NextRunPreview `json:"next_runs"`
}

// ValidateCronResponse represents the response for cron validation
type ValidateCronResponse struct {
	Valid       bool     `json:"valid"`
	Expression  string   `json:"expression"`
	Description string   `json:"description,omitempty"`
	NextRuns    []string `json:"next_runs,omitempty"`
	Error       string   `json:"error,omitempty"`
}

// TimezoneInfo represents timezone information
type TimezoneInfo struct {
	Name   string `json:"name"`
	Offset string `json:"offset"`
	Region string `json:"region,omitempty"`
}

// CreateScheduleRequest represents a request to create a new schedule
type CreateScheduleRequest struct {
	Name           string                 `json:"name"`
	Description    string                 `json:"description,omitempty"`
	CronExpression string                 `json:"cron_expression"`
	Timezone       string                 `json:"timezone,omitempty"`
	TargetType     ScheduleTargetType     `json:"target_type"`
	TargetConfig   map[string]interface{} `json:"target_config"`
	Enabled        *bool                  `json:"enabled,omitempty"`
	MaxRetries     *int                   `json:"max_retries,omitempty"`
	RetryPolicy    string                 `json:"retry_policy,omitempty"`
	TimeoutMinutes *int                   `json:"timeout_minutes,omitempty"`
}

// UpdateScheduleRequest represents a request to update an existing schedule
type UpdateScheduleRequest struct {
	Name           *string                `json:"name,omitempty"`
	Description    *string                `json:"description,omitempty"`
	CronExpression *string                `json:"cron_expression,omitempty"`
	Timezone       *string                `json:"timezone,omitempty"`
	TargetType     *string                `json:"target_type,omitempty"`
	TargetConfig   map[string]interface{} `json:"target_config,omitempty"`
	Enabled        *bool                  `json:"enabled,omitempty"`
	MaxRetries     *int                   `json:"max_retries,omitempty"`
	RetryPolicy    *string                `json:"retry_policy,omitempty"`
	TimeoutMinutes *int                   `json:"timeout_minutes,omitempty"`
}

// TriggerScheduleRequest represents a request to manually trigger a schedule
type TriggerScheduleRequest struct {
	SkipJitter bool   `json:"skip_jitter,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

// ValidateCronRequest represents a request to validate a cron expression
type ValidateCronRequest struct {
	CronExpression string `json:"cron_expression"`
	Timezone       string `json:"timezone,omitempty"`
	PreviewCount   int    `json:"preview_count,omitempty"`
}

// ListSchedulesOptions represents options for filtering schedule listings
type ListSchedulesOptions struct {
	Page       int    `url:"page,omitempty"`
	PageSize   int    `url:"page_size,omitempty"`
	Status     string `url:"status,omitempty"`
	TargetType string `url:"target_type,omitempty"`
	Enabled    *bool  `url:"enabled,omitempty"`
	Search     string `url:"search,omitempty"`
}

// ToQuery converts ListSchedulesOptions to query parameters
func (o *ListSchedulesOptions) ToQuery() map[string]string {
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
	if o.TargetType != "" {
		params["target_type"] = o.TargetType
	}
	if o.Enabled != nil {
		params["enabled"] = fmt.Sprintf("%t", *o.Enabled)
	}
	if o.Search != "" {
		params["search"] = o.Search
	}
	return params
}

// ListExecutionsOptions represents options for filtering execution listings
type ListExecutionsOptions struct {
	Page     int    `url:"page,omitempty"`
	PageSize int    `url:"page_size,omitempty"`
	Status   string `url:"status,omitempty"`
}

// ToQuery converts ListExecutionsOptions to query parameters
func (o *ListExecutionsOptions) ToQuery() map[string]string {
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
	return params
}

// PaginatedSchedulesResponse wraps a paginated schedules response
type PaginatedSchedulesResponse struct {
	Schedules  []Schedule     `json:"schedules"`
	Pagination PaginationMeta `json:"pagination"`
}

// PaginatedScheduleExecutionsResponse wraps a paginated schedule executions response
type PaginatedScheduleExecutionsResponse struct {
	Executions []ScheduleExecution `json:"executions"`
	Pagination PaginationMeta      `json:"pagination"`
}

// =============================================================================
// Schedule CRUD Operations
// =============================================================================

// CreateSchedule creates a new schedule
// Authentication: JWT Token or Unified API Key required
// Endpoint: POST /v1/schedules
func (s *SchedulesService) CreateSchedule(ctx context.Context, req *CreateScheduleRequest) (*Schedule, *Response, error) {
	var resp struct {
		Status  string   `json:"status"`
		Message string   `json:"message"`
		Data    Schedule `json:"data"`
	}

	apiResp, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/schedules",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, nil, err
	}

	return &resp.Data, apiResp, nil
}

// ListSchedules retrieves a paginated list of schedules
// Authentication: JWT Token or Unified API Key required
// Endpoint: GET /v1/schedules
func (s *SchedulesService) ListSchedules(ctx context.Context, opts *ListSchedulesOptions) (*PaginatedSchedulesResponse, *Response, error) {
	var resp struct {
		Status  string       `json:"status"`
		Message string       `json:"message"`
		Data    []Schedule   `json:"data"`
		Meta    PaginationMeta `json:"meta"`
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/schedules",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	apiResp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return &PaginatedSchedulesResponse{
		Schedules:  resp.Data,
		Pagination: resp.Meta,
	}, apiResp, nil
}

// GetSchedule retrieves a specific schedule by ID
// Authentication: JWT Token or Unified API Key required
// Endpoint: GET /v1/schedules/{id}
func (s *SchedulesService) GetSchedule(ctx context.Context, scheduleID uint) (*Schedule, *Response, error) {
	var resp struct {
		Status  string   `json:"status"`
		Message string   `json:"message"`
		Data    Schedule `json:"data"`
	}

	apiResp, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/schedules/%d", scheduleID),
		Result: &resp,
	})
	if err != nil {
		return nil, nil, err
	}

	return &resp.Data, apiResp, nil
}

// UpdateSchedule updates an existing schedule
// Authentication: JWT Token or Unified API Key required
// Endpoint: PUT /v1/schedules/{id}
func (s *SchedulesService) UpdateSchedule(ctx context.Context, scheduleID uint, req *UpdateScheduleRequest) (*Schedule, *Response, error) {
	var resp struct {
		Status  string   `json:"status"`
		Message string   `json:"message"`
		Data    Schedule `json:"data"`
	}

	apiResp, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/schedules/%d", scheduleID),
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, nil, err
	}

	return &resp.Data, apiResp, nil
}

// DeleteSchedule deletes a schedule
// Authentication: JWT Token or Unified API Key required
// Endpoint: DELETE /v1/schedules/{id}
func (s *SchedulesService) DeleteSchedule(ctx context.Context, scheduleID uint) (*Response, error) {
	var resp struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	apiResp, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v1/schedules/%d", scheduleID),
		Result: &resp,
	})
	return apiResp, err
}

// =============================================================================
// Schedule Control Operations
// =============================================================================

// EnableSchedule enables a schedule
// Authentication: JWT Token or Unified API Key required
// Endpoint: POST /v1/schedules/{id}/enable
func (s *SchedulesService) EnableSchedule(ctx context.Context, scheduleID uint) (*Schedule, *Response, error) {
	var resp struct {
		Status  string   `json:"status"`
		Message string   `json:"message"`
		Data    Schedule `json:"data"`
	}

	apiResp, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/schedules/%d/enable", scheduleID),
		Result: &resp,
	})
	if err != nil {
		return nil, nil, err
	}

	return &resp.Data, apiResp, nil
}

// DisableSchedule disables a schedule
// Authentication: JWT Token or Unified API Key required
// Endpoint: POST /v1/schedules/{id}/disable
func (s *SchedulesService) DisableSchedule(ctx context.Context, scheduleID uint) (*Schedule, *Response, error) {
	var resp struct {
		Status  string   `json:"status"`
		Message string   `json:"message"`
		Data    Schedule `json:"data"`
	}

	apiResp, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/schedules/%d/disable", scheduleID),
		Result: &resp,
	})
	if err != nil {
		return nil, nil, err
	}

	return &resp.Data, apiResp, nil
}

// TriggerSchedule manually triggers a schedule execution
// Authentication: JWT Token or Unified API Key required
// Endpoint: POST /v1/schedules/{id}/trigger
func (s *SchedulesService) TriggerSchedule(ctx context.Context, scheduleID uint, req *TriggerScheduleRequest) (*ScheduleExecution, *Response, error) {
	var resp struct {
		Status  string            `json:"status"`
		Message string            `json:"message"`
		Data    ScheduleExecution `json:"data"`
	}

	apiResp, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/schedules/%d/trigger", scheduleID),
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, nil, err
	}

	return &resp.Data, apiResp, nil
}

// =============================================================================
// Schedule History and Preview
// =============================================================================

// GetExecutions retrieves execution history for a schedule
// Authentication: JWT Token or Unified API Key required
// Endpoint: GET /v1/schedules/{id}/executions
func (s *SchedulesService) GetExecutions(ctx context.Context, scheduleID uint, opts *ListExecutionsOptions) (*PaginatedScheduleExecutionsResponse, *Response, error) {
	var resp struct {
		Status  string              `json:"status"`
		Message string              `json:"message"`
		Data    []ScheduleExecution `json:"data"`
		Meta    PaginationMeta      `json:"meta"`
	}

	req := &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/schedules/%d/executions", scheduleID),
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	apiResp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return &PaginatedScheduleExecutionsResponse{
		Executions: resp.Data,
		Pagination: resp.Meta,
	}, apiResp, nil
}

// GetStatistics retrieves statistics for a schedule
// Authentication: JWT Token or Unified API Key required
// Endpoint: GET /v1/schedules/{id}/statistics
func (s *SchedulesService) GetStatistics(ctx context.Context, scheduleID uint, period string) (*ScheduleStatistics, *Response, error) {
	var resp struct {
		Status  string             `json:"status"`
		Message string             `json:"message"`
		Data    ScheduleStatistics `json:"data"`
	}

	req := &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/schedules/%d/statistics", scheduleID),
		Result: &resp,
	}

	if period != "" {
		req.Query = map[string]string{"period": period}
	}

	apiResp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return &resp.Data, apiResp, nil
}

// GetNextRuns retrieves a preview of upcoming run times for a schedule
// Authentication: JWT Token or Unified API Key required
// Endpoint: GET /v1/schedules/{id}/next-runs
func (s *SchedulesService) GetNextRuns(ctx context.Context, scheduleID uint, count int) (*NextRunsResponse, *Response, error) {
	var resp struct {
		Status  string           `json:"status"`
		Message string           `json:"message"`
		Data    NextRunsResponse `json:"data"`
	}

	req := &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/schedules/%d/next-runs", scheduleID),
		Result: &resp,
	}

	if count > 0 {
		req.Query = map[string]string{"count": fmt.Sprintf("%d", count)}
	}

	apiResp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return &resp.Data, apiResp, nil
}

// =============================================================================
// Execution Management
// =============================================================================

// ExecutionCallbackRequest represents a request to update execution status via callback
type ExecutionCallbackRequest struct {
	Status             string                 `json:"status"`
	TargetResourceID   string                 `json:"target_resource_id,omitempty"`
	TargetResourceType string                 `json:"target_resource_type,omitempty"`
	ErrorMessage       string                 `json:"error_message,omitempty"`
	ErrorDetails       map[string]interface{} `json:"error_details,omitempty"`
	Result             map[string]interface{} `json:"result,omitempty"`
	DurationMs         *int                   `json:"duration_ms,omitempty"`
}

// UpdateExecutionRequest represents a request to update an execution record
type UpdateExecutionRequest struct {
	Status             *string                `json:"status,omitempty"`
	TargetResourceID   *string                `json:"target_resource_id,omitempty"`
	TargetResourceType *string                `json:"target_resource_type,omitempty"`
	ErrorMessage       *string                `json:"error_message,omitempty"`
	ErrorDetails       map[string]interface{} `json:"error_details,omitempty"`
	Result             map[string]interface{} `json:"result,omitempty"`
	DurationMs         *int                   `json:"duration_ms,omitempty"`
}

// ExecutionCallback reports execution completion/failure via callback
// This is used by job/report controllers to report back execution status
// Authentication: JWT Token or Unified API Key required
// Endpoint: POST /v1/schedules/{id}/executions/{executionId}/callback
func (s *SchedulesService) ExecutionCallback(ctx context.Context, scheduleID, executionID uint, req *ExecutionCallbackRequest) (*ScheduleExecution, *Response, error) {
	var resp struct {
		Status  string            `json:"status"`
		Message string            `json:"message"`
		Data    ScheduleExecution `json:"data"`
	}

	apiResp, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/schedules/%d/executions/%d/callback", scheduleID, executionID),
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, nil, err
	}

	return &resp.Data, apiResp, nil
}

// UpdateExecution updates an execution record
// Authentication: JWT Token or Unified API Key required
// Endpoint: PATCH /v1/schedules/{id}/executions/{executionId}
func (s *SchedulesService) UpdateExecution(ctx context.Context, scheduleID, executionID uint, req *UpdateExecutionRequest) (*ScheduleExecution, *Response, error) {
	var resp struct {
		Status  string            `json:"status"`
		Message string            `json:"message"`
		Data    ScheduleExecution `json:"data"`
	}

	apiResp, err := s.client.Do(ctx, &Request{
		Method: "PATCH",
		Path:   fmt.Sprintf("/v1/schedules/%d/executions/%d", scheduleID, executionID),
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, nil, err
	}

	return &resp.Data, apiResp, nil
}

// GetExecution retrieves a specific execution record
// Authentication: JWT Token or Unified API Key required
// Endpoint: GET /v1/schedules/{id}/executions/{executionId}
func (s *SchedulesService) GetExecution(ctx context.Context, scheduleID, executionID uint) (*ScheduleExecution, *Response, error) {
	var resp struct {
		Status  string            `json:"status"`
		Message string            `json:"message"`
		Data    ScheduleExecution `json:"data"`
	}

	apiResp, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/schedules/%d/executions/%d", scheduleID, executionID),
		Result: &resp,
	})
	if err != nil {
		return nil, nil, err
	}

	return &resp.Data, apiResp, nil
}

// =============================================================================
// Utility Operations
// =============================================================================

// ValidateCron validates a cron expression and optionally previews run times
// Authentication: JWT Token or Unified API Key required
// Endpoint: POST /v1/schedules/validate-cron
func (s *SchedulesService) ValidateCron(ctx context.Context, req *ValidateCronRequest) (*ValidateCronResponse, *Response, error) {
	var resp struct {
		Status  string               `json:"status"`
		Message string               `json:"message"`
		Data    ValidateCronResponse `json:"data"`
	}

	apiResp, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/schedules/validate-cron",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, nil, err
	}

	return &resp.Data, apiResp, nil
}

// ListTimezones retrieves a list of supported timezones
// Authentication: JWT Token or Unified API Key required
// Endpoint: GET /v1/schedules/timezones
func (s *SchedulesService) ListTimezones(ctx context.Context, region string) ([]TimezoneInfo, *Response, error) {
	var resp struct {
		Status  string         `json:"status"`
		Message string         `json:"message"`
		Data    []TimezoneInfo `json:"data"`
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/schedules/timezones",
		Result: &resp,
	}

	if region != "" {
		req.Query = map[string]string{"region": region}
	}

	apiResp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return resp.Data, apiResp, nil
}

// =============================================================================
// Helper Methods
// =============================================================================

// IsEnabled returns true if the schedule is enabled
func (s *Schedule) IsEnabled() bool {
	return s.Enabled && s.Status == ScheduleStatusActive
}

// IsActive returns true if the schedule is in active status
func (s *Schedule) IsActive() bool {
	return s.Status == ScheduleStatusActive
}

// IsPaused returns true if the schedule is paused
func (s *Schedule) IsPaused() bool {
	return s.Status == ScheduleStatusPaused
}

// HasErrors returns true if the schedule has errors
func (s *Schedule) HasErrors() bool {
	return s.Status == ScheduleStatusError || s.LastRunError != nil
}

// GetSuccessRate returns the success rate as a percentage
func (s *Schedule) GetSuccessRate() float64 {
	if s.RunCount == 0 {
		return 0
	}
	return float64(s.SuccessCount) / float64(s.RunCount) * 100
}

// IsComplete returns true if the execution is in a terminal state
func (e *ScheduleExecution) IsComplete() bool {
	return e.Status == ScheduleExecutionCompleted ||
		e.Status == ScheduleExecutionFailed ||
		e.Status == ScheduleExecutionCancelled ||
		e.Status == ScheduleExecutionTimedOut
}

// IsSuccessful returns true if the execution completed successfully
func (e *ScheduleExecution) IsSuccessful() bool {
	return e.Status == ScheduleExecutionCompleted
}

// IsFailed returns true if the execution failed
func (e *ScheduleExecution) IsFailed() bool {
	return e.Status == ScheduleExecutionFailed || e.Status == ScheduleExecutionTimedOut
}

// IsRunning returns true if the execution is currently running
func (e *ScheduleExecution) IsRunning() bool {
	return e.Status == ScheduleExecutionRunning
}
