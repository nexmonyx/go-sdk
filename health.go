package nexmonyx

import (
	"context"
	"fmt"
	"time"
)

// GetHealth retrieves the health status of the API
func (s *HealthService) GetHealth(ctx context.Context) (*HealthStatus, error) {
	var resp StandardResponse
	resp.Data = &HealthStatus{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/healthz",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if health, ok := resp.Data.(*HealthStatus); ok {
		return health, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetHealthDetailed retrieves detailed health status
func (s *HealthService) GetHealthDetailed(ctx context.Context) (*DetailedHealthStatus, error) {
	var resp StandardResponse
	resp.Data = &DetailedHealthStatus{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/health/detailed",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if health, ok := resp.Data.(*DetailedHealthStatus); ok {
		return health, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// List retrieves all health checks
func (s *HealthService) List(ctx context.Context, opts *HealthCheckListOptions) ([]HealthCheck, *PaginationMeta, error) {
	var resp PaginatedResponse
	resp.Data = &[]HealthCheck{}

	query := make(map[string]string)
	if opts != nil {
		if opts.ServerID != nil {
			query["server_id"] = fmt.Sprintf("%d", *opts.ServerID)
		}
		if opts.CheckType != nil {
			query["check_type"] = *opts.CheckType
		}
		if opts.IsEnabled != nil {
			if *opts.IsEnabled {
				query["is_enabled"] = "true"
			} else {
				query["is_enabled"] = "false"
			}
		}
		if opts.ListOptions.Page > 0 {
			query["page"] = fmt.Sprintf("%d", opts.ListOptions.Page)
		}
		if opts.ListOptions.Limit > 0 {
			query["limit"] = fmt.Sprintf("%d", opts.ListOptions.Limit)
		}
		if opts.ListOptions.Sort != "" {
			query["sort"] = opts.ListOptions.Sort
		}
		if opts.ListOptions.Order != "" {
			query["order"] = opts.ListOptions.Order
		}
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/health/checks",
		Query:  query,
		Result: &resp,
	})
	if err != nil {
		return nil, nil, err
	}

	if checks, ok := resp.Data.(*[]HealthCheck); ok {
		return *checks, resp.Meta, nil
	}
	return nil, nil, fmt.Errorf("unexpected response type")
}

// Get retrieves a specific health check by ID
func (s *HealthService) Get(ctx context.Context, id uint) (*HealthCheck, error) {
	var resp StandardResponse
	resp.Data = &HealthCheck{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/health/checks/%d", id),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if check, ok := resp.Data.(*HealthCheck); ok {
		return check, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// Create creates a new health check
func (s *HealthService) Create(ctx context.Context, req *CreateHealthCheckRequest) (*HealthCheck, error) {
	var resp StandardResponse
	resp.Data = &HealthCheck{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/health/checks",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if check, ok := resp.Data.(*HealthCheck); ok {
		return check, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// Update updates an existing health check
func (s *HealthService) Update(ctx context.Context, id uint, req *UpdateHealthCheckRequest) (*HealthCheck, error) {
	var resp StandardResponse
	resp.Data = &HealthCheck{}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/health/checks/%d", id),
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if check, ok := resp.Data.(*HealthCheck); ok {
		return check, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// Delete removes a health check
func (s *HealthService) Delete(ctx context.Context, id uint) error {
	_, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v1/health/checks/%d", id),
	})
	return err
}

// GetHistory retrieves health check history
func (s *HealthService) GetHistory(ctx context.Context, opts *HealthCheckHistoryListOptions) ([]HealthCheckHistory, *PaginationMeta, error) {
	var resp PaginatedResponse
	resp.Data = &[]HealthCheckHistory{}

	query := make(map[string]string)
	if opts != nil {
		if opts.HealthCheckID != nil {
			query["health_check_id"] = fmt.Sprintf("%d", *opts.HealthCheckID)
		}
		if opts.ServerID != nil {
			query["server_id"] = fmt.Sprintf("%d", *opts.ServerID)
		}
		if opts.Status != nil {
			query["status"] = *opts.Status
		}
		if opts.FromDate != nil {
			query["from_date"] = opts.FromDate.Format(time.RFC3339)
		}
		if opts.ToDate != nil {
			query["to_date"] = opts.ToDate.Format(time.RFC3339)
		}
		if opts.ListOptions.Page > 0 {
			query["page"] = fmt.Sprintf("%d", opts.ListOptions.Page)
		}
		if opts.ListOptions.Limit > 0 {
			query["limit"] = fmt.Sprintf("%d", opts.ListOptions.Limit)
		}
		if opts.ListOptions.Sort != "" {
			query["sort"] = opts.ListOptions.Sort
		}
		if opts.ListOptions.Order != "" {
			query["order"] = opts.ListOptions.Order
		}
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/health/history",
		Query:  query,
		Result: &resp,
	})
	if err != nil {
		return nil, nil, err
	}

	if history, ok := resp.Data.(*[]HealthCheckHistory); ok {
		return *history, resp.Meta, nil
	}
	return nil, nil, fmt.Errorf("unexpected response type")
}

// Organization-Level Health Check Methods (Task 307)

// CreateHealthCheckDefinition creates a new organization-level health check definition
func (s *HealthService) CreateHealthCheckDefinition(ctx context.Context, req *CreateHealthCheckDefinitionRequest) (*HealthCheckDefinitionResponse, error) {
	var resp StandardResponse
	resp.Data = &HealthCheckDefinitionResponse{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/health/definitions",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if def, ok := resp.Data.(*HealthCheckDefinitionResponse); ok {
		return def, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// ListHealthCheckDefinitions retrieves all organization-level health check definitions
func (s *HealthService) ListHealthCheckDefinitions(ctx context.Context) (*ListHealthCheckDefinitionsResponse, error) {
	var resp StandardResponse
	resp.Data = &ListHealthCheckDefinitionsResponse{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/health/definitions",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if list, ok := resp.Data.(*ListHealthCheckDefinitionsResponse); ok {
		return list, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetHealthCheckDefinition retrieves a specific health check definition by ID
func (s *HealthService) GetHealthCheckDefinition(ctx context.Context, id uint64) (*HealthCheckDefinitionResponse, error) {
	var resp StandardResponse
	resp.Data = &HealthCheckDefinitionResponse{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/health/definitions/%d", id),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if def, ok := resp.Data.(*HealthCheckDefinitionResponse); ok {
		return def, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// UpdateHealthCheckDefinition updates an existing health check definition
func (s *HealthService) UpdateHealthCheckDefinition(ctx context.Context, id uint64, req *CreateHealthCheckDefinitionRequest) (*HealthCheckDefinitionResponse, error) {
	var resp StandardResponse
	resp.Data = &HealthCheckDefinitionResponse{}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/health/definitions/%d", id),
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if def, ok := resp.Data.(*HealthCheckDefinitionResponse); ok {
		return def, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// DeleteHealthCheckDefinition removes a health check definition
func (s *HealthService) DeleteHealthCheckDefinition(ctx context.Context, id uint64) error {
	_, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v1/health/definitions/%d", id),
	})
	return err
}

// SubmitHealthCheckResult submits a health check result for a defined health check
func (s *HealthService) SubmitHealthCheckResult(ctx context.Context, req *SubmitHealthCheckResultRequest) (*HealthCheckResultResponse, error) {
	var resp StandardResponse
	resp.Data = &HealthCheckResultResponse{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/health/results",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if result, ok := resp.Data.(*HealthCheckResultResponse); ok {
		return result, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetOrganizationHealthStatus retrieves aggregated health status for the organization
func (s *HealthService) GetOrganizationHealthStatus(ctx context.Context) (*HealthStatusAggregateResponse, error) {
	var resp StandardResponse
	resp.Data = &HealthStatusAggregateResponse{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/health/status",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if status, ok := resp.Data.(*HealthStatusAggregateResponse); ok {
		return status, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// ListHealthAlerts retrieves all health alerts for the organization
func (s *HealthService) ListHealthAlerts(ctx context.Context) (*ListHealthAlertsResponse, error) {
	var resp StandardResponse
	resp.Data = &ListHealthAlertsResponse{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/health/alerts",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if alerts, ok := resp.Data.(*ListHealthAlertsResponse); ok {
		return alerts, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// Cross-Controller Health Monitoring Methods

// GetAllControllerHealthStatus retrieves health status for all monitored controllers
func (s *HealthService) GetAllControllerHealthStatus(ctx context.Context) (*ControllerHealthStatusResponse, error) {
	var resp StandardResponse
	resp.Data = &ControllerHealthStatusResponse{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/health/controllers/status",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if status, ok := resp.Data.(*ControllerHealthStatusResponse); ok {
		return status, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetControllerHealthStatus retrieves health status for a specific controller
func (s *HealthService) GetControllerHealthStatus(ctx context.Context, controllerName string) (*ControllerHealthDetailResponse, error) {
	var resp StandardResponse
	resp.Data = &ControllerHealthDetailResponse{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/health/controllers/%s/status", controllerName),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if detail, ok := resp.Data.(*ControllerHealthDetailResponse); ok {
		return detail, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetSystemHealthOverview retrieves comprehensive system-wide health metrics
func (s *HealthService) GetSystemHealthOverview(ctx context.Context) (*SystemHealthOverview, error) {
	var resp StandardResponse
	resp.Data = &SystemHealthOverview{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/health/system/overview",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if overview, ok := resp.Data.(*SystemHealthOverview); ok {
		return overview, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetServiceHealthHistory retrieves historical health metrics for a specific service
func (s *HealthService) GetServiceHealthHistory(ctx context.Context, serviceName string, startTime, endTime, granularity string) (*HealthMetricsHistory, error) {
	var resp StandardResponse
	resp.Data = &HealthMetricsHistory{}

	query := make(map[string]string)
	if startTime != "" {
		query["start_time"] = startTime
	}
	if endTime != "" {
		query["end_time"] = endTime
	}
	if granularity != "" {
		query["granularity"] = granularity
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/health/services/%s/history", serviceName),
		Query:  query,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if history, ok := resp.Data.(*HealthMetricsHistory); ok {
		return history, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetServiceHealthScore retrieves the health score (0-100) for a specific service
func (s *HealthService) GetServiceHealthScore(ctx context.Context, serviceName string) (*ServiceHealthScore, error) {
	var resp StandardResponse
	resp.Data = &ServiceHealthScore{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/health/services/%s/score", serviceName),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if score, ok := resp.Data.(*ServiceHealthScore); ok {
		return score, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// HealthStatus represents the basic health status
type HealthStatus struct {
	Status    string      `json:"status"`
	Healthy   bool        `json:"healthy"`
	Version   string      `json:"version"`
	Timestamp *CustomTime `json:"timestamp"`
}

// DetailedHealthStatus represents detailed health information
type DetailedHealthStatus struct {
	Status    string                   `json:"status"`
	Healthy   bool                     `json:"healthy"`
	Version   string                   `json:"version"`
	Timestamp *CustomTime              `json:"timestamp"`
	Uptime    int64                    `json:"uptime"`
	Services  map[string]ServiceHealth `json:"services"`
	Database  *DatabaseHealth          `json:"database"`
	Redis     *RedisHealth             `json:"redis"`
	Metrics   map[string]interface{}   `json:"metrics"`
}

// ServiceHealth represents the health of a service
type ServiceHealth struct {
	Healthy      bool   `json:"healthy"`
	Status       string `json:"status"`
	Message      string `json:"message,omitempty"`
	ResponseTime int    `json:"response_time,omitempty"` // milliseconds
}

// DatabaseHealth represents database health
type DatabaseHealth struct {
	Healthy         bool   `json:"healthy"`
	ConnectionCount int    `json:"connection_count"`
	MaxConnections  int    `json:"max_connections"`
	ResponseTime    int    `json:"response_time"` // milliseconds
	Version         string `json:"version"`
}

// RedisHealth represents Redis health
type RedisHealth struct {
	Healthy          bool   `json:"healthy"`
	Connected        bool   `json:"connected"`
	ResponseTime     int    `json:"response_time"` // milliseconds
	MemoryUsage      int64  `json:"memory_usage"`
	ConnectedClients int    `json:"connected_clients"`
	Version          string `json:"version"`
}

// HealthCheck represents a health check definition
type HealthCheck struct {
	ID                  uint                   `json:"id"`
	ServerID            uint                   `json:"server_id"`
	CheckName           string                 `json:"check_name"`
	CheckType           string                 `json:"check_type"`
	CheckDescription    string                 `json:"check_description,omitempty"`
	IsEnabled           bool                   `json:"is_enabled"`
	CheckInterval       int                    `json:"check_interval_minutes"` // minutes
	CheckTimeout        int                    `json:"check_timeout_seconds"`  // seconds
	MaxRetries          int                    `json:"max_retries"`
	RetryInterval       int                    `json:"retry_interval_seconds"` // seconds
	CheckData           map[string]interface{} `json:"check_data,omitempty"`
	Threshold           map[string]interface{} `json:"threshold,omitempty"`
	LastCheckAt         *CustomTime            `json:"last_check_at,omitempty"`
	NextCheckAt         time.Time              `json:"next_check_at"`
	LastStatus          string                 `json:"last_status,omitempty"`
	LastScore           int                    `json:"last_score,omitempty"`
	ConsecutiveFailures int                    `json:"consecutive_failures"`
	CreatedAt           *CustomTime            `json:"created_at"`
	UpdatedAt           *CustomTime            `json:"updated_at"`

	// Related data
	Server *Server `json:"server,omitempty"`
}

// HealthCheckHistory represents a health check result entry
type HealthCheckHistory struct {
	ID            uint                   `json:"id"`
	HealthCheckID uint                   `json:"health_check_id"`
	ServerID      uint                   `json:"server_id"`
	Status        string                 `json:"status"`           // healthy, warning, critical
	Score         int                    `json:"score"`            // 0-100
	ResponseTime  int64                  `json:"response_time_ms"` // milliseconds
	ErrorMessage  string                 `json:"error_message,omitempty"`
	CheckData     map[string]interface{} `json:"check_data,omitempty"`
	Attempt       int                    `json:"attempt"` // retry attempt number
	CreatedAt     *CustomTime            `json:"created_at"`

	// Related data
	HealthCheck *HealthCheck `json:"health_check,omitempty"`
	Server      *Server      `json:"server,omitempty"`
}

// HealthCheckListOptions represents options for listing health checks
type HealthCheckListOptions struct {
	ServerID    *uint   `json:"server_id,omitempty"`
	CheckType   *string `json:"check_type,omitempty"`
	IsEnabled   *bool   `json:"is_enabled,omitempty"`
	ListOptions ListOptions
}

// HealthCheckHistoryListOptions represents options for listing health check history
type HealthCheckHistoryListOptions struct {
	HealthCheckID *uint      `json:"health_check_id,omitempty"`
	ServerID      *uint      `json:"server_id,omitempty"`
	Status        *string    `json:"status,omitempty"`
	FromDate      *time.Time `json:"from_date,omitempty"`
	ToDate        *time.Time `json:"to_date,omitempty"`
	ListOptions   ListOptions
}

// CreateHealthCheckRequest represents a request to create a health check
type CreateHealthCheckRequest struct {
	ServerID         uint                   `json:"server_id" validate:"required"`
	CheckName        string                 `json:"check_name" validate:"required"`
	CheckType        string                 `json:"check_type" validate:"required"`
	CheckDescription string                 `json:"check_description,omitempty"`
	IsEnabled        bool                   `json:"is_enabled"`
	CheckInterval    int                    `json:"check_interval_minutes" validate:"min=1"`
	CheckTimeout     int                    `json:"check_timeout_seconds" validate:"min=1"`
	MaxRetries       int                    `json:"max_retries" validate:"min=0"`
	RetryInterval    int                    `json:"retry_interval_seconds" validate:"min=1"`
	CheckData        map[string]interface{} `json:"check_data,omitempty"`
	Threshold        map[string]interface{} `json:"threshold,omitempty"`
}

// UpdateHealthCheckRequest represents a request to update a health check
type UpdateHealthCheckRequest struct {
	CheckName        *string                `json:"check_name,omitempty"`
	CheckType        *string                `json:"check_type,omitempty"`
	CheckDescription *string                `json:"check_description,omitempty"`
	IsEnabled        *bool                  `json:"is_enabled,omitempty"`
	CheckInterval    *int                   `json:"check_interval_minutes,omitempty"`
	CheckTimeout     *int                   `json:"check_timeout_seconds,omitempty"`
	MaxRetries       *int                   `json:"max_retries,omitempty"`
	RetryInterval    *int                   `json:"retry_interval_seconds,omitempty"`
	CheckData        map[string]interface{} `json:"check_data,omitempty"`
	Threshold        map[string]interface{} `json:"threshold,omitempty"`
}

// Cross-Controller Health Response Types

// ControllerHealthStatusResponse represents the response for all controller health statuses
type ControllerHealthStatusResponse struct {
	Controllers map[string]ControllerStatus `json:"controllers"`
	Total       int                         `json:"total"`
	Timestamp   string                      `json:"timestamp"`
}

// ControllerStatus represents the health status of a single controller
type ControllerStatus struct {
	Status      string            `json:"status"`
	Message     string            `json:"message"`
	Details     map[string]string `json:"details"`
	LastUpdated string            `json:"last_updated"`
	Duration    string            `json:"duration"`
}

// ControllerHealthDetailResponse represents detailed health information for a specific controller
type ControllerHealthDetailResponse struct {
	ControllerName string            `json:"controller_name"`
	Status         string            `json:"status"`
	Message        string            `json:"message"`
	Details        map[string]string `json:"details"`
	LastUpdated    string            `json:"last_updated"`
	Duration       string            `json:"duration"`
	ResponseTimeMs int64             `json:"response_time_ms"`
}

// Health Controller Specific Methods

// GetHealthControllerInfo retrieves basic information about the health controller service
func (s *HealthService) GetHealthControllerInfo(ctx context.Context) (*ControllerHealthInfo, error) {
	var resp StandardResponse
	resp.Data = &ControllerHealthInfo{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/health/info",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if info, ok := resp.Data.(*ControllerHealthInfo); ok {
		return info, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetHealthControllerStatus retrieves the current status of the health controller
func (s *HealthService) GetHealthControllerStatus(ctx context.Context) (*ControllerStatus, error) {
	var resp StandardResponse
	resp.Data = &ControllerStatus{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/admin/health-controller/status",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if status, ok := resp.Data.(*ControllerStatus); ok {
		return status, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetHealthControllerLeaderStatus retrieves the current leader election status for health controller
func (s *HealthService) GetHealthControllerLeaderStatus(ctx context.Context) (map[string]interface{}, error) {
	var resp StandardResponse
	var leaderStatus map[string]interface{}
	resp.Data = &leaderStatus

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/admin/health-controller/leader-status",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return leaderStatus, nil
}

// GetAdminSystemHealthStatus retrieves comprehensive system health status with pagination
func (s *HealthService) GetAdminSystemHealthStatus(ctx context.Context, opts *SystemHealthStatusOptions) ([]SystemHealthResult, *PaginationMeta, error) {
	var resp PaginatedResponse
	resp.Data = &[]SystemHealthResult{}

	query := make(map[string]string)
	if opts != nil {
		if opts.Status != "" {
			query["status"] = opts.Status
		}
		if opts.Category != "" {
			query["category"] = opts.Category
		}
		if opts.ListOptions.Page > 0 {
			query["page"] = fmt.Sprintf("%d", opts.ListOptions.Page)
		}
		if opts.ListOptions.Limit > 0 {
			query["limit"] = fmt.Sprintf("%d", opts.ListOptions.Limit)
		}
		if opts.ListOptions.Sort != "" {
			query["sort"] = opts.ListOptions.Sort
		}
		if opts.ListOptions.Order != "" {
			query["order"] = opts.ListOptions.Order
		}
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/admin/health/system-status",
		Query:  query,
		Result: &resp,
	})
	if err != nil {
		return nil, nil, err
	}

	if results, ok := resp.Data.(*[]SystemHealthResult); ok {
		return *results, resp.Meta, nil
	}
	return nil, nil, fmt.Errorf("unexpected response type")
}

// TriggerManualHealthCheck manually triggers a health check or group of health checks
func (s *HealthService) TriggerManualHealthCheck(ctx context.Context, req *TriggerHealthCheckRequest) (*TriggerHealthCheckResponse, error) {
	var resp StandardResponse
	resp.Data = &TriggerHealthCheckResponse{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/admin/health/trigger-check",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if result, ok := resp.Data.(*TriggerHealthCheckResponse); ok {
		return result, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// Health Controller Types



// SystemHealthResult represents a single health check result
type SystemHealthResult struct {
	CheckName    string                 `json:"check_name"`
	Category     string                 `json:"category"`
	Status       string                 `json:"status"`
	Message      string                 `json:"message,omitempty"`
	ResponseTime int64                  `json:"response_time_ms"`
	LastChecked  *CustomTime            `json:"last_checked"`
	Details      map[string]interface{} `json:"details,omitempty"`
}

// SystemHealthStatusOptions represents options for querying system health status
type SystemHealthStatusOptions struct {
	Status      string      `json:"status,omitempty"`
	Category    string      `json:"category,omitempty"`
	ListOptions ListOptions `json:"list_options,omitempty"`
}

// TriggerHealthCheckRequest represents a request to trigger health checks
type TriggerHealthCheckRequest struct {
	CheckName string            `json:"check_name,omitempty"`
	Category  string            `json:"category,omitempty"`
	Timeout   int               `json:"timeout,omitempty"`
	Options   map[string]string `json:"options,omitempty"`
}

// TriggerHealthCheckResponse represents the response from triggering health checks
type TriggerHealthCheckResponse struct {
	TriggeredChecks int                  `json:"triggered_checks"`
	Results         []SystemHealthResult `json:"results"`
}
