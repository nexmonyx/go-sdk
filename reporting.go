package nexmonyx

import (
	"context"
	"fmt"
)

// ReportingService handles report generation and scheduling operations
type ReportingService struct {
	client *Client
}

// GenerateReport generates a new report with specified configuration
// Authentication: JWT Token required
// Endpoint: POST /v1/reports/generate
// Parameters:
//   - reportType: Type of report (usage, performance, compliance, billing)
//   - config: Report configuration including parameters and filters
// Returns: Report object with generation status
func (s *ReportingService) GenerateReport(ctx context.Context, config *ReportConfiguration) (*Report, error) {
	var resp struct {
		Data    *Report `json:"data"`
		Status  string  `json:"status"`
		Message string  `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/reports/generate",
		Body:   config,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// ListReports retrieves a list of reports with optional filtering
// Authentication: JWT Token required
// Endpoint: GET /v1/reports
// Parameters:
//   - opts: Optional pagination options
//   - status: Optional status filter (pending, generating, completed, failed)
// Returns: Array of Report objects with pagination metadata
func (s *ReportingService) ListReports(ctx context.Context, opts *PaginationOptions, status string) ([]Report, *PaginationMeta, error) {
	var resp struct {
		Data []Report         `json:"data"`
		Meta *PaginationMeta  `json:"meta"`
	}

	query := make(map[string]string)
	if opts != nil {
		if opts.Page > 0 {
			query["page"] = fmt.Sprintf("%d", opts.Page)
		}
		if opts.Limit > 0 {
			query["limit"] = fmt.Sprintf("%d", opts.Limit)
		}
	}
	if status != "" {
		query["status"] = status
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/reports",
		Result: &resp,
	}
	if len(query) > 0 {
		req.Query = query
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return resp.Data, resp.Meta, nil
}

// GetReport retrieves details of a specific report
// Authentication: JWT Token required
// Endpoint: GET /v1/reports/{id}
// Parameters:
//   - reportID: Report ID
// Returns: Report object with full details
func (s *ReportingService) GetReport(ctx context.Context, reportID uint) (*Report, error) {
	var resp struct {
		Data    *Report `json:"data"`
		Status  string  `json:"status"`
		Message string  `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/reports/%d", reportID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// DownloadReport downloads the generated report file
// Authentication: JWT Token required
// Endpoint: GET /v1/reports/{id}/download
// Parameters:
//   - reportID: Report ID
// Returns: Report file content as byte array
func (s *ReportingService) DownloadReport(ctx context.Context, reportID uint) ([]byte, error) {
	// Note: This endpoint returns raw file content, not JSON
	resp, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/reports/%d/download", reportID),
	})
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

// ScheduleReport creates a scheduled report with recurring execution
// Authentication: JWT Token required
// Endpoint: POST /v1/reports/schedule
// Parameters:
//   - schedule: ReportSchedule object with cron expression and configuration
// Returns: Created ReportSchedule object
func (s *ReportingService) ScheduleReport(ctx context.Context, schedule *ReportSchedule) (*ReportSchedule, error) {
	var resp struct {
		Data    *ReportSchedule `json:"data"`
		Status  string          `json:"status"`
		Message string          `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/reports/schedule",
		Body:   schedule,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// ListSchedules retrieves a list of scheduled reports
// Authentication: JWT Token required
// Endpoint: GET /v1/reports/schedules
// Parameters:
//   - opts: Optional pagination options
//   - enabled: Optional filter for enabled/disabled schedules (nil = all)
// Returns: Array of ReportSchedule objects with pagination metadata
func (s *ReportingService) ListSchedules(ctx context.Context, opts *PaginationOptions, enabled *bool) ([]ReportSchedule, *PaginationMeta, error) {
	var resp struct {
		Data []ReportSchedule `json:"data"`
		Meta *PaginationMeta  `json:"meta"`
	}

	query := make(map[string]string)
	if opts != nil {
		if opts.Page > 0 {
			query["page"] = fmt.Sprintf("%d", opts.Page)
		}
		if opts.Limit > 0 {
			query["limit"] = fmt.Sprintf("%d", opts.Limit)
		}
	}
	if enabled != nil {
		query["enabled"] = fmt.Sprintf("%t", *enabled)
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/reports/schedules",
		Result: &resp,
	}
	if len(query) > 0 {
		req.Query = query
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return resp.Data, resp.Meta, nil
}

// DeleteSchedule deletes a scheduled report
// Authentication: JWT Token required
// Endpoint: DELETE /v1/reports/schedules/{id}
// Parameters:
//   - scheduleID: Schedule ID
// Returns: Success confirmation
func (s *ReportingService) DeleteSchedule(ctx context.Context, scheduleID uint) error {
	_, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v1/reports/schedules/%d", scheduleID),
	})
	return err
}
