package nexmonyx

import (
	"context"
	"fmt"
)

// MaintenanceWindowsService handles maintenance window API operations
// Maintenance windows are used to schedule planned maintenance periods with alert suppression
type MaintenanceWindowsService struct {
	client *Client
}

// MaintenanceWindowStatus represents the status of a maintenance window
type MaintenanceWindowStatus string

const (
	MaintenanceWindowStatusScheduled MaintenanceWindowStatus = "scheduled"
	MaintenanceWindowStatusActive    MaintenanceWindowStatus = "active"
	MaintenanceWindowStatusCompleted MaintenanceWindowStatus = "completed"
	MaintenanceWindowStatusCancelled MaintenanceWindowStatus = "cancelled"
)

// MaintenanceAction represents an action to take during a maintenance window
type MaintenanceAction struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config,omitempty"`
}

// ScheduledMaintenanceWindow represents a scheduled maintenance window
type ScheduledMaintenanceWindow struct {
	ID              uint                    `json:"id"`
	WindowUUID      string                  `json:"window_uuid"`
	OrganizationID  uint                    `json:"organization_id"`
	ScheduleID      *uint                   `json:"schedule_id,omitempty"`
	Name            string                  `json:"name"`
	Description     string                  `json:"description,omitempty"`
	StartsAt        string                  `json:"starts_at"`
	EndsAt          string                  `json:"ends_at"`
	Duration        int                     `json:"duration"`
	Status          MaintenanceWindowStatus `json:"status"`
	SuppressAlerts  bool                    `json:"suppress_alerts"`
	PauseMonitoring bool                    `json:"pause_monitoring"`
	NotifyBefore    int                     `json:"notify_before"`
	ServerFilter    map[string]interface{}  `json:"server_filter"`
	Actions         []MaintenanceAction     `json:"actions,omitempty"`
	AlertSilenceID  *uint                   `json:"alert_silence_id,omitempty"`
	ActivatedAt     *string                 `json:"activated_at,omitempty"`
	CompletedAt     *string                 `json:"completed_at,omitempty"`
	CancelledAt     *string                 `json:"cancelled_at,omitempty"`
	CancelReason    *string                 `json:"cancel_reason,omitempty"`
	CreatedAt       string                  `json:"created_at"`
	UpdatedAt       string                  `json:"updated_at"`
	CreatedByID     *uint                   `json:"created_by_id,omitempty"`
	CreatedByEmail  *string                 `json:"created_by_email,omitempty"`
	UpdatedByID     *uint                   `json:"updated_by_id,omitempty"`
	UpdatedByEmail  *string                 `json:"updated_by_email,omitempty"`
}

// CreateMaintenanceWindowRequest represents the request to create a maintenance window
type CreateMaintenanceWindowRequest struct {
	Name            string                 `json:"name"`
	Description     string                 `json:"description,omitempty"`
	StartsAt        string                 `json:"starts_at"`
	EndsAt          string                 `json:"ends_at"`
	SuppressAlerts  bool                   `json:"suppress_alerts"`
	PauseMonitoring bool                   `json:"pause_monitoring"`
	NotifyBefore    int                    `json:"notify_before,omitempty"`
	ServerFilter    map[string]interface{} `json:"server_filter"`
	Actions         []MaintenanceAction    `json:"actions,omitempty"`
	ScheduleID      *uint                  `json:"schedule_id,omitempty"`
}

// UpdateMaintenanceWindowRequest represents the request to update a maintenance window
type UpdateMaintenanceWindowRequest struct {
	Name            *string                 `json:"name,omitempty"`
	Description     *string                 `json:"description,omitempty"`
	StartsAt        *string                 `json:"starts_at,omitempty"`
	EndsAt          *string                 `json:"ends_at,omitempty"`
	SuppressAlerts  *bool                   `json:"suppress_alerts,omitempty"`
	PauseMonitoring *bool                   `json:"pause_monitoring,omitempty"`
	NotifyBefore    *int                    `json:"notify_before,omitempty"`
	ServerFilter    *map[string]interface{} `json:"server_filter,omitempty"`
	Actions         *[]MaintenanceAction    `json:"actions,omitempty"`
}

// CancelMaintenanceWindowRequest represents the request to cancel a maintenance window
type CancelMaintenanceWindowRequest struct {
	Reason string `json:"reason,omitempty"`
}

// ListMaintenanceWindowsOptions represents options for filtering maintenance window listings
type ListMaintenanceWindowsOptions struct {
	Page     int    `url:"page,omitempty"`
	Limit    int    `url:"limit,omitempty"`
	Status   string `url:"status,omitempty"`
	FromDate string `url:"from_date,omitempty"`
	ToDate   string `url:"to_date,omitempty"`
}

// ToQuery converts ListMaintenanceWindowsOptions to query parameters
func (o *ListMaintenanceWindowsOptions) ToQuery() map[string]string {
	params := make(map[string]string)
	if o.Page > 0 {
		params["page"] = fmt.Sprintf("%d", o.Page)
	}
	if o.Limit > 0 {
		params["limit"] = fmt.Sprintf("%d", o.Limit)
	}
	if o.Status != "" {
		params["status"] = o.Status
	}
	if o.FromDate != "" {
		params["from_date"] = o.FromDate
	}
	if o.ToDate != "" {
		params["to_date"] = o.ToDate
	}
	return params
}

// GetUpcomingOptions represents options for getting upcoming maintenance windows
type GetUpcomingOptions struct {
	Days int `url:"days,omitempty"`
}

// ToQuery converts GetUpcomingOptions to query parameters
func (o *GetUpcomingOptions) ToQuery() map[string]string {
	params := make(map[string]string)
	if o.Days > 0 {
		params["days"] = fmt.Sprintf("%d", o.Days)
	}
	return params
}

// MaintenanceWindowsListResponse wraps a paginated maintenance windows response
type MaintenanceWindowsListResponse struct {
	Windows    []ScheduledMaintenanceWindow `json:"windows"`
	Total      int64                        `json:"total"`
	Page       int                          `json:"page"`
	Limit      int                          `json:"limit"`
	TotalPages int                          `json:"total_pages"`
}

// ActiveWindowsResponse wraps an active maintenance windows response
type ActiveWindowsResponse struct {
	Windows []ScheduledMaintenanceWindow `json:"windows"`
	Count   int                          `json:"count"`
}

// UpcomingWindowsResponse wraps an upcoming maintenance windows response
type UpcomingWindowsResponse struct {
	Windows  []ScheduledMaintenanceWindow `json:"windows"`
	Count    int                          `json:"count"`
	DaysFrom int                          `json:"days_from"`
}

// =============================================================================
// Maintenance Window CRUD Operations
// =============================================================================

// CreateMaintenanceWindow creates a new maintenance window
// Authentication: JWT Token or Unified API Key required
// Endpoint: POST /v1/maintenance-windows
func (s *MaintenanceWindowsService) CreateMaintenanceWindow(ctx context.Context, req *CreateMaintenanceWindowRequest) (*ScheduledMaintenanceWindow, *Response, error) {
	var resp struct {
		Status  string                     `json:"status"`
		Message string                     `json:"message"`
		Data    ScheduledMaintenanceWindow `json:"data"`
	}

	apiResp, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/maintenance-windows",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, nil, err
	}

	return &resp.Data, apiResp, nil
}

// ListMaintenanceWindows retrieves a paginated list of maintenance windows
// Authentication: JWT Token or Unified API Key required
// Endpoint: GET /v1/maintenance-windows
func (s *MaintenanceWindowsService) ListMaintenanceWindows(ctx context.Context, opts *ListMaintenanceWindowsOptions) (*MaintenanceWindowsListResponse, *Response, error) {
	var resp struct {
		Status  string                         `json:"status"`
		Message string                         `json:"message"`
		Data    MaintenanceWindowsListResponse `json:"data"`
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/maintenance-windows",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	apiResp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return &resp.Data, apiResp, nil
}

// GetMaintenanceWindow retrieves a specific maintenance window by ID
// Authentication: JWT Token or Unified API Key required
// Endpoint: GET /v1/maintenance-windows/{id}
func (s *MaintenanceWindowsService) GetMaintenanceWindow(ctx context.Context, windowID uint) (*ScheduledMaintenanceWindow, *Response, error) {
	var resp struct {
		Status  string                     `json:"status"`
		Message string                     `json:"message"`
		Data    ScheduledMaintenanceWindow `json:"data"`
	}

	apiResp, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/maintenance-windows/%d", windowID),
		Result: &resp,
	})
	if err != nil {
		return nil, nil, err
	}

	return &resp.Data, apiResp, nil
}

// UpdateMaintenanceWindow updates an existing maintenance window
// Authentication: JWT Token or Unified API Key required
// Endpoint: PUT /v1/maintenance-windows/{id}
func (s *MaintenanceWindowsService) UpdateMaintenanceWindow(ctx context.Context, windowID uint, req *UpdateMaintenanceWindowRequest) (*ScheduledMaintenanceWindow, *Response, error) {
	var resp struct {
		Status  string                     `json:"status"`
		Message string                     `json:"message"`
		Data    ScheduledMaintenanceWindow `json:"data"`
	}

	apiResp, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/maintenance-windows/%d", windowID),
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, nil, err
	}

	return &resp.Data, apiResp, nil
}

// DeleteMaintenanceWindow deletes a maintenance window (soft delete)
// Authentication: JWT Token or Unified API Key required
// Endpoint: DELETE /v1/maintenance-windows/{id}
func (s *MaintenanceWindowsService) DeleteMaintenanceWindow(ctx context.Context, windowID uint) (*Response, error) {
	var resp struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	apiResp, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v1/maintenance-windows/%d", windowID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return apiResp, nil
}

// =============================================================================
// Maintenance Window Special Operations
// =============================================================================

// GetActiveMaintenanceWindows retrieves all currently active maintenance windows
// Authentication: JWT Token or Unified API Key required
// Endpoint: GET /v1/maintenance-windows/active
func (s *MaintenanceWindowsService) GetActiveMaintenanceWindows(ctx context.Context) (*ActiveWindowsResponse, *Response, error) {
	var resp struct {
		Status  string                `json:"status"`
		Message string                `json:"message"`
		Data    ActiveWindowsResponse `json:"data"`
	}

	apiResp, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/maintenance-windows/active",
		Result: &resp,
	})
	if err != nil {
		return nil, nil, err
	}

	return &resp.Data, apiResp, nil
}

// GetUpcomingMaintenanceWindows retrieves upcoming scheduled maintenance windows
// Authentication: JWT Token or Unified API Key required
// Endpoint: GET /v1/maintenance-windows/upcoming
func (s *MaintenanceWindowsService) GetUpcomingMaintenanceWindows(ctx context.Context, opts *GetUpcomingOptions) (*UpcomingWindowsResponse, *Response, error) {
	var resp struct {
		Status  string                  `json:"status"`
		Message string                  `json:"message"`
		Data    UpcomingWindowsResponse `json:"data"`
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/maintenance-windows/upcoming",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	apiResp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return &resp.Data, apiResp, nil
}

// CancelMaintenanceWindow cancels a scheduled or active maintenance window
// Authentication: JWT Token or Unified API Key required
// Endpoint: POST /v1/maintenance-windows/{id}/cancel
func (s *MaintenanceWindowsService) CancelMaintenanceWindow(ctx context.Context, windowID uint, req *CancelMaintenanceWindowRequest) (*ScheduledMaintenanceWindow, *Response, error) {
	var resp struct {
		Status  string                     `json:"status"`
		Message string                     `json:"message"`
		Data    ScheduledMaintenanceWindow `json:"data"`
	}

	apiResp, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/maintenance-windows/%d/cancel", windowID),
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, nil, err
	}

	return &resp.Data, apiResp, nil
}
