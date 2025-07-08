package nexmonyx

import (
	"context"
	"fmt"
)

// ProbeAlertsService handles communication with the probe alerts endpoints
type ProbeAlertsService struct {
	client *Client
}

// ProbeAlert represents a monitoring probe alert
type ProbeAlert struct {
	ID               uint                 `json:"id"`
	ProbeID          uint                 `json:"probe_id"`
	Name             string               `json:"name"`
	Status           string               `json:"status"` // active, acknowledged, resolved
	Message          string               `json:"message"`
	Conditions       ProbeAlertConditions `json:"conditions"`
	TriggeredAt      *CustomTime          `json:"triggered_at"`
	AcknowledgedBy   *uint                `json:"acknowledged_by,omitempty"`
	AcknowledgedAt   *CustomTime          `json:"acknowledged_at,omitempty"`
	ResolvedAt       *CustomTime          `json:"resolved_at,omitempty"`
	Resolution       *string              `json:"resolution,omitempty"`
	NotificationSent bool                 `json:"notification_sent"`
	CreatedAt        *CustomTime          `json:"created_at"`
	UpdatedAt        *CustomTime          `json:"updated_at"`
}

// ProbeAlertConditions represents the conditions that triggered an alert
type ProbeAlertConditions struct {
	FailureThreshold  int `json:"failure_threshold"`
	RecoveryThreshold int `json:"recovery_threshold"`
}

// ProbeAlertListOptions represents options for listing probe alerts
type ProbeAlertListOptions struct {
	ListOptions
	Status  string // Filter by status
	ProbeID int    // Filter by probe ID
}

// ToQuery converts ProbeAlertListOptions to query parameters
func (opts *ProbeAlertListOptions) ToQuery() map[string]string {
	params := opts.ListOptions.ToQuery()

	if opts.Status != "" {
		params["status"] = opts.Status
	}
	if opts.ProbeID > 0 {
		params["probe_id"] = fmt.Sprintf("%d", opts.ProbeID)
	}

	return params
}

// List retrieves all probe alerts for the organization
func (s *ProbeAlertsService) List(ctx context.Context, opts *ProbeAlertListOptions) ([]*ProbeAlert, *PaginationMeta, error) {
	var resp struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Alerts     []*ProbeAlert   `json:"alerts"`
			Pagination *PaginationMeta `json:"pagination"`
		} `json:"data"`
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/probe-alerts",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return resp.Data.Alerts, resp.Data.Pagination, nil
}

// Get retrieves a specific probe alert by ID
func (s *ProbeAlertsService) Get(ctx context.Context, id uint) (*ProbeAlert, error) {
	var resp struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Alert *ProbeAlert `json:"alert"`
			Probe struct {
				ID   uint   `json:"id"`
				UUID string `json:"uuid"`
				Name string `json:"name"`
				Type string `json:"type"`
			} `json:"probe"`
		} `json:"data"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/probe-alerts/%d", id),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data.Alert, nil
}

// Acknowledge acknowledges a probe alert
func (s *ProbeAlertsService) Acknowledge(ctx context.Context, id uint, note string) (*ProbeAlert, error) {
	var resp struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Alert *ProbeAlert `json:"alert"`
		} `json:"data"`
	}

	body := map[string]string{}
	if note != "" {
		body["note"] = note
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/probe-alerts/%d/acknowledge", id),
		Body:   body,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data.Alert, nil
}

// Resolve resolves a probe alert
func (s *ProbeAlertsService) Resolve(ctx context.Context, id uint, resolution string) (*ProbeAlert, error) {
	var resp struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Alert *ProbeAlert `json:"alert"`
		} `json:"data"`
	}

	body := map[string]string{}
	if resolution != "" {
		body["resolution"] = resolution
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/probe-alerts/%d/resolve", id),
		Body:   body,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data.Alert, nil
}

// AdminProbeAlert represents a probe alert with additional organization information (admin view)
type AdminProbeAlert struct {
	ProbeAlert
	OrganizationName string `json:"organization_name"`
	OrganizationID   uint   `json:"organization_id"`
	ProbeName        string `json:"probe_name"`
	ProbeType        string `json:"probe_type"`
	ProbeTarget      string `json:"probe_target"`
}

// AdminProbeAlertListOptions represents options for listing probe alerts as admin
type AdminProbeAlertListOptions struct {
	ListOptions
	Status         string // Filter by status
	OrganizationID int    // Filter by organization
}

// ToQuery converts AdminProbeAlertListOptions to query parameters
func (opts *AdminProbeAlertListOptions) ToQuery() map[string]string {
	params := opts.ListOptions.ToQuery()

	if opts.Status != "" {
		params["status"] = opts.Status
	}
	if opts.OrganizationID > 0 {
		params["organization_id"] = fmt.Sprintf("%d", opts.OrganizationID)
	}

	return params
}

// ListAdmin retrieves all probe alerts across all organizations (admin only)
func (s *ProbeAlertsService) ListAdmin(ctx context.Context, opts *AdminProbeAlertListOptions) ([]*AdminProbeAlert, *PaginationMeta, error) {
	var resp struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Alerts     []*AdminProbeAlert `json:"alerts"`
			Pagination *PaginationMeta    `json:"pagination"`
		} `json:"data"`
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/admin/probe-alerts",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return resp.Data.Alerts, resp.Data.Pagination, nil
}
