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
		Path:   "/api/v1/health",
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
		Path:   "/api/v1/health/detailed",
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
		Path:   "/api/v1/health/checks",
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
		Path:   fmt.Sprintf("/api/v1/health/checks/%d", id),
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
		Path:   "/api/v1/health/checks",
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
		Path:   fmt.Sprintf("/api/v1/health/checks/%d", id),
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
		Path:   fmt.Sprintf("/api/v1/health/checks/%d", id),
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
		Path:   "/api/v1/health/history",
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
