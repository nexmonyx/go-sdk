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

// DeleteReport deletes a specific report
// Authentication: JWT Token required
// Endpoint: DELETE /v1/reports/{id}
// Parameters:
//   - reportID: Report ID
// Returns: Success confirmation
func (s *ReportingService) DeleteReport(ctx context.Context, reportID uint) error {
	_, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v1/reports/%d", reportID),
	})
	return err
}

// GetReportStatus retrieves the status of a specific report
// Authentication: JWT Token required
// Endpoint: GET /v1/reports/{id}/status
// Parameters:
//   - reportID: Report ID
// Returns: ReportStatus object with current progress
func (s *ReportingService) GetReportStatus(ctx context.Context, reportID uint) (*ReportStatus, error) {
	var resp struct {
		Data    *ReportStatus `json:"data"`
		Status  string        `json:"status"`
		Message string        `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/reports/%d/status", reportID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// GetSchedule retrieves details of a specific schedule
// Authentication: JWT Token required
// Endpoint: GET /v1/reports/schedules/{id}
// Parameters:
//   - scheduleID: Schedule ID
// Returns: ReportSchedule object with full details
func (s *ReportingService) GetSchedule(ctx context.Context, scheduleID uint) (*ReportSchedule, error) {
	var resp struct {
		Data    *ReportSchedule `json:"data"`
		Status  string          `json:"status"`
		Message string          `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/reports/schedules/%d", scheduleID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// UpdateSchedule updates an existing schedule
// Authentication: JWT Token required
// Endpoint: PUT /v1/reports/schedules/{id}
// Parameters:
//   - scheduleID: Schedule ID
//   - update: UpdateReportScheduleRequest with fields to update
// Returns: Updated ReportSchedule object
func (s *ReportingService) UpdateSchedule(ctx context.Context, scheduleID uint, update *UpdateReportScheduleRequest) (*ReportSchedule, error) {
	var resp struct {
		Data    *ReportSchedule `json:"data"`
		Status  string          `json:"status"`
		Message string          `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/reports/schedules/%d", scheduleID),
		Body:   update,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// ============================================================================
// Template Operations
// ============================================================================

// ListTemplates retrieves a list of available report templates
// Authentication: JWT Token required
// Endpoint: GET /v1/reports/templates
// Parameters:
//   - opts: Optional pagination options
//   - templateType: Optional filter by template type (health, alert, inventory, uptime, custom)
//   - includeSystem: Include system templates (default: true)
// Returns: Array of ReportTemplate objects with pagination metadata
func (s *ReportingService) ListTemplates(ctx context.Context, opts *PaginationOptions, templateType string, includeSystem *bool) ([]ReportTemplate, *PaginationMeta, error) {
	var resp struct {
		Data []ReportTemplate `json:"data"`
		Meta *PaginationMeta  `json:"meta"`
	}

	query := make(map[string]string)
	if opts != nil {
		if opts.Page > 0 {
			query["page"] = fmt.Sprintf("%d", opts.Page)
		}
		if opts.Limit > 0 {
			query["per_page"] = fmt.Sprintf("%d", opts.Limit)
		}
	}
	if templateType != "" {
		query["type"] = templateType
	}
	if includeSystem != nil {
		query["include_system"] = fmt.Sprintf("%t", *includeSystem)
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/reports/templates",
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

// GetTemplate retrieves details of a specific template
// Authentication: JWT Token required
// Endpoint: GET /v1/reports/templates/{id}
// Parameters:
//   - templateID: Template ID
// Returns: ReportTemplate object with full details
func (s *ReportingService) GetTemplate(ctx context.Context, templateID uint) (*ReportTemplate, error) {
	var resp struct {
		Data    *ReportTemplate `json:"data"`
		Status  string          `json:"status"`
		Message string          `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/reports/templates/%d", templateID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// CreateTemplate creates a new report template
// Authentication: JWT Token required
// Endpoint: POST /v1/reports/templates
// Parameters:
//   - template: CreateTemplateRequest with template definition
// Returns: Created ReportTemplate object
func (s *ReportingService) CreateTemplate(ctx context.Context, template *CreateTemplateRequest) (*ReportTemplate, error) {
	var resp struct {
		Data    *ReportTemplate `json:"data"`
		Status  string          `json:"status"`
		Message string          `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/reports/templates",
		Body:   template,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// UpdateTemplate updates an existing template
// Authentication: JWT Token required
// Endpoint: PUT /v1/reports/templates/{id}
// Parameters:
//   - templateID: Template ID
//   - update: UpdateTemplateRequest with fields to update
// Returns: Updated ReportTemplate object
func (s *ReportingService) UpdateTemplate(ctx context.Context, templateID uint, update *UpdateTemplateRequest) (*ReportTemplate, error) {
	var resp struct {
		Data    *ReportTemplate `json:"data"`
		Status  string          `json:"status"`
		Message string          `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/reports/templates/%d", templateID),
		Body:   update,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// DeleteTemplate deletes a specific template
// Authentication: JWT Token required
// Endpoint: DELETE /v1/reports/templates/{id}
// Parameters:
//   - templateID: Template ID
// Returns: Success confirmation
func (s *ReportingService) DeleteTemplate(ctx context.Context, templateID uint) error {
	_, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v1/reports/templates/%d", templateID),
	})
	return err
}

// PreviewTemplate generates a preview of a template
// Authentication: JWT Token required
// Endpoint: POST /v1/reports/templates/{id}/preview
// Parameters:
//   - templateID: Template ID
//   - preview: TemplatePreviewRequest with preview parameters
// Returns: TemplatePreviewResponse with sample data
func (s *ReportingService) PreviewTemplate(ctx context.Context, templateID uint, preview *TemplatePreviewRequest) (*TemplatePreviewResponse, error) {
	var resp struct {
		Data    *TemplatePreviewResponse `json:"data"`
		Status  string                   `json:"status"`
		Message string                   `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/reports/templates/%d/preview", templateID),
		Body:   preview,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// ============================================================================
// Quick Reports
// ============================================================================

// QuickReportOptions contains common options for quick reports
type QuickReportOptions struct {
	Format    string           // Output format (json, pdf, csv, xlsx)
	TimeRange *ReportTimeRange // Time range for the report
}

// GetHealthSummary generates a quick health summary report
// Authentication: JWT Token required
// Endpoint: GET /v1/reports/quick/health-summary
// Parameters:
//   - opts: Optional quick report options
// Returns: HealthSummaryResponse with server health data
func (s *ReportingService) GetHealthSummary(ctx context.Context, opts *QuickReportOptions) (*HealthSummaryResponse, error) {
	var resp struct {
		Data    *HealthSummaryResponse `json:"data"`
		Status  string                 `json:"status"`
		Message string                 `json:"message"`
	}

	query := make(map[string]string)
	if opts != nil {
		if opts.Format != "" {
			query["format"] = opts.Format
		}
		if opts.TimeRange != nil {
			if opts.TimeRange.StartDate != "" {
				query["start_date"] = opts.TimeRange.StartDate
			}
			if opts.TimeRange.EndDate != "" {
				query["end_date"] = opts.TimeRange.EndDate
			}
		}
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/reports/quick/health-summary",
		Result: &resp,
	}
	if len(query) > 0 {
		req.Query = query
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// AlertHistoryOptions contains options for alert history reports
type AlertHistoryOptions struct {
	QuickReportOptions
	Severity string // Filter by severity (critical, warning, info)
	Status   string // Filter by status (active, resolved)
}

// GetAlertHistory generates a quick alert history report
// Authentication: JWT Token required
// Endpoint: GET /v1/reports/quick/alert-history
// Parameters:
//   - opts: Optional alert history options
// Returns: AlertHistoryResponse with alert data
func (s *ReportingService) GetAlertHistory(ctx context.Context, opts *AlertHistoryOptions) (*AlertHistoryResponse, error) {
	var resp struct {
		Data    *AlertHistoryResponse `json:"data"`
		Status  string                `json:"status"`
		Message string                `json:"message"`
	}

	query := make(map[string]string)
	if opts != nil {
		if opts.Format != "" {
			query["format"] = opts.Format
		}
		if opts.Severity != "" {
			query["severity"] = opts.Severity
		}
		if opts.Status != "" {
			query["status"] = opts.Status
		}
		if opts.TimeRange != nil {
			if opts.TimeRange.StartDate != "" {
				query["start_date"] = opts.TimeRange.StartDate
			}
			if opts.TimeRange.EndDate != "" {
				query["end_date"] = opts.TimeRange.EndDate
			}
		}
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/reports/quick/alert-history",
		Result: &resp,
	}
	if len(query) > 0 {
		req.Query = query
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// ServerInventoryOptions contains options for server inventory reports
type ServerInventoryOptions struct {
	QuickReportOptions
	Status string // Filter by server status (online, offline)
	OS     string // Filter by operating system
}

// GetServerInventory generates a quick server inventory report
// Authentication: JWT Token required
// Endpoint: GET /v1/reports/quick/server-inventory
// Parameters:
//   - opts: Optional server inventory options
// Returns: ServerInventoryResponse with server data
func (s *ReportingService) GetServerInventory(ctx context.Context, opts *ServerInventoryOptions) (*ServerInventoryResponse, error) {
	var resp struct {
		Data    *ServerInventoryResponse `json:"data"`
		Status  string                   `json:"status"`
		Message string                   `json:"message"`
	}

	query := make(map[string]string)
	if opts != nil {
		if opts.Format != "" {
			query["format"] = opts.Format
		}
		if opts.Status != "" {
			query["status"] = opts.Status
		}
		if opts.OS != "" {
			query["os"] = opts.OS
		}
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/reports/quick/server-inventory",
		Result: &resp,
	}
	if len(query) > 0 {
		req.Query = query
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// GetUptimeReport generates a quick uptime report
// Authentication: JWT Token required
// Endpoint: GET /v1/reports/quick/uptime
// Parameters:
//   - opts: Optional quick report options
// Returns: UptimeReportResponse with uptime data
func (s *ReportingService) GetUptimeReport(ctx context.Context, opts *QuickReportOptions) (*UptimeReportResponse, error) {
	var resp struct {
		Data    *UptimeReportResponse `json:"data"`
		Status  string                `json:"status"`
		Message string                `json:"message"`
	}

	query := make(map[string]string)
	if opts != nil {
		if opts.Format != "" {
			query["format"] = opts.Format
		}
		if opts.TimeRange != nil {
			if opts.TimeRange.StartDate != "" {
				query["start_date"] = opts.TimeRange.StartDate
			}
			if opts.TimeRange.EndDate != "" {
				query["end_date"] = opts.TimeRange.EndDate
			}
		}
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/reports/quick/uptime",
		Result: &resp,
	}
	if len(query) > 0 {
		req.Query = query
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}
