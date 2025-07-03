package nexmonyx

import (
	"context"
	"fmt"
)

// CreateAlert creates a new alert
func (s *AlertsService) Create(ctx context.Context, alert *Alert) (*Alert, error) {
	var resp StandardResponse
	resp.Data = &Alert{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/api/v1/alerts",
		Body:   alert,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if created, ok := resp.Data.(*Alert); ok {
		return created, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetAlert retrieves an alert by ID
func (s *AlertsService) Get(ctx context.Context, id string) (*Alert, error) {
	var resp StandardResponse
	resp.Data = &Alert{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/api/v1/alerts/%s", id),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if alert, ok := resp.Data.(*Alert); ok {
		return alert, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// ListAlerts retrieves a list of alerts
func (s *AlertsService) List(ctx context.Context, opts *ListOptions) ([]*Alert, *PaginationMeta, error) {
	var resp PaginatedResponse
	var alerts []*Alert
	resp.Data = &alerts

	req := &Request{
		Method: "GET",
		Path:   "/api/v1/alerts",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return alerts, resp.Meta, nil
}

// UpdateAlert updates an existing alert
func (s *AlertsService) Update(ctx context.Context, id string, alert *Alert) (*Alert, error) {
	var resp StandardResponse
	resp.Data = &Alert{}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/api/v1/alerts/%s", id),
		Body:   alert,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if updated, ok := resp.Data.(*Alert); ok {
		return updated, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// DeleteAlert deletes an alert
func (s *AlertsService) Delete(ctx context.Context, id string) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/api/v1/alerts/%s", id),
		Result: &resp,
	})
	return err
}

// EnableAlert enables an alert
func (s *AlertsService) Enable(ctx context.Context, id string) (*Alert, error) {
	var resp StandardResponse
	resp.Data = &Alert{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/api/v1/alerts/%s/enable", id),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if alert, ok := resp.Data.(*Alert); ok {
		return alert, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// DisableAlert disables an alert
func (s *AlertsService) Disable(ctx context.Context, id string) (*Alert, error) {
	var resp StandardResponse
	resp.Data = &Alert{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/api/v1/alerts/%s/disable", id),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if alert, ok := resp.Data.(*Alert); ok {
		return alert, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetAlertHistory retrieves alert history
func (s *AlertsService) GetHistory(ctx context.Context, id string, opts *ListOptions) ([]*AlertHistoryEntry, *PaginationMeta, error) {
	var resp PaginatedResponse
	var history []*AlertHistoryEntry
	resp.Data = &history

	req := &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/api/v1/alerts/%s/history", id),
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return history, resp.Meta, nil
}

// TestAlert tests an alert configuration
func (s *AlertsService) Test(ctx context.Context, id string) (*AlertTestResult, error) {
	var resp StandardResponse
	resp.Data = &AlertTestResult{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/api/v1/alerts/%s/test", id),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if result, ok := resp.Data.(*AlertTestResult); ok {
		return result, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// AcknowledgeAlert acknowledges an alert
func (s *AlertsService) Acknowledge(ctx context.Context, id string, message string) error {
	var resp StandardResponse

	body := map[string]interface{}{
		"message": message,
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/api/v1/alerts/%s/acknowledge", id),
		Body:   body,
		Result: &resp,
	})
	return err
}

// AlertHistoryEntry represents an alert history entry
type AlertHistoryEntry struct {
	ID          uint                   `json:"id"`
	AlertID     uint                   `json:"alert_id"`
	TriggeredAt *CustomTime            `json:"triggered_at"`
	ResolvedAt  *CustomTime            `json:"resolved_at,omitempty"`
	Status      string                 `json:"status"`
	Value       float64                `json:"value"`
	Threshold   float64                `json:"threshold"`
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// AlertTestResult represents the result of an alert test
type AlertTestResult struct {
	Success   bool                   `json:"success"`
	Triggered bool                   `json:"triggered"`
	Message   string                 `json:"message"`
	Value     float64                `json:"value,omitempty"`
	Threshold float64                `json:"threshold,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Errors    []string               `json:"errors,omitempty"`
}
