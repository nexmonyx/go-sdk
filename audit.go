package nexmonyx

import (
	"context"
	"fmt"
)

// AuditService handles audit log operations and compliance tracking
type AuditService struct {
	client *Client
}

// GetAuditLogs retrieves audit logs with comprehensive filtering
// Authentication: JWT Token required
// Endpoint: GET /v1/audit/logs
// Parameters:
//   - opts: Optional pagination options
//   - filters: Optional filters (user_id, action, resource_type, start_date, end_date, severity)
// Returns: Array of AuditLog objects with pagination metadata
func (s *AuditService) GetAuditLogs(ctx context.Context, opts *PaginationOptions, filters map[string]interface{}) ([]AuditLog, *PaginationMeta, error) {
	var resp struct {
		Data []AuditLog      `json:"data"`
		Meta *PaginationMeta `json:"meta"`
	}

	queryParams := make(map[string]string)
	if opts != nil {
		if opts.Page > 0 {
			queryParams["page"] = fmt.Sprintf("%d", opts.Page)
		}
		if opts.Limit > 0 {
			queryParams["limit"] = fmt.Sprintf("%d", opts.Limit)
		}
	}

	// Add filter parameters
	if filters != nil {
		if userID, ok := filters["user_id"].(uint); ok && userID > 0 {
			queryParams["user_id"] = fmt.Sprintf("%d", userID)
		}
		if action, ok := filters["action"].(string); ok && action != "" {
			queryParams["action"] = action
		}
		if resourceType, ok := filters["resource_type"].(string); ok && resourceType != "" {
			queryParams["resource_type"] = resourceType
		}
		if resourceID, ok := filters["resource_id"].(string); ok && resourceID != "" {
			queryParams["resource_id"] = resourceID
		}
		if startDate, ok := filters["start_date"].(string); ok && startDate != "" {
			queryParams["start_date"] = startDate
		}
		if endDate, ok := filters["end_date"].(string); ok && endDate != "" {
			queryParams["end_date"] = endDate
		}
		if severity, ok := filters["severity"].(string); ok && severity != "" {
			queryParams["severity"] = severity
		}
		if ipAddress, ok := filters["ip_address"].(string); ok && ipAddress != "" {
			queryParams["ip_address"] = ipAddress
		}
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/audit/logs",
		Result: &resp,
	}
	if len(queryParams) > 0 {
		req.Query = queryParams
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return resp.Data, resp.Meta, nil
}

// GetAuditLog retrieves a specific audit log entry by ID
// Authentication: JWT Token required
// Endpoint: GET /v1/audit/logs/{id}
// Parameters:
//   - id: Audit log ID
// Returns: AuditLog object with full details
func (s *AuditService) GetAuditLog(ctx context.Context, id uint) (*AuditLog, error) {
	var resp struct {
		Data    *AuditLog `json:"data"`
		Status  string    `json:"status"`
		Message string    `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/audit/logs/%d", id),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// ExportAuditLogs exports audit logs in the specified format
// Authentication: JWT Token required
// Endpoint: POST /v1/audit/logs/export
// Parameters:
//   - format: Export format (csv, json, pdf)
//   - filters: Optional filters (same as GetAuditLogs)
// Returns: Exported audit logs as byte array
func (s *AuditService) ExportAuditLogs(ctx context.Context, format string, filters map[string]interface{}) ([]byte, error) {
	body := map[string]interface{}{
		"format": format,
	}

	// Add filters to request body
	if filters != nil {
		if userID, ok := filters["user_id"].(uint); ok && userID > 0 {
			body["user_id"] = userID
		}
		if action, ok := filters["action"].(string); ok && action != "" {
			body["action"] = action
		}
		if resourceType, ok := filters["resource_type"].(string); ok && resourceType != "" {
			body["resource_type"] = resourceType
		}
		if startDate, ok := filters["start_date"].(string); ok && startDate != "" {
			body["start_date"] = startDate
		}
		if endDate, ok := filters["end_date"].(string); ok && endDate != "" {
			body["end_date"] = endDate
		}
		if severity, ok := filters["severity"].(string); ok && severity != "" {
			body["severity"] = severity
		}
	}

	resp, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/audit/logs/export",
		Body:   body,
	})
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

// GetAuditStatistics retrieves comprehensive audit activity statistics
// Authentication: JWT Token required
// Endpoint: GET /v1/audit/statistics
// Parameters:
//   - startDate: Optional start date filter (ISO 8601 format)
//   - endDate: Optional end date filter (ISO 8601 format)
// Returns: AuditStatistics object with activity breakdown
func (s *AuditService) GetAuditStatistics(ctx context.Context, startDate string, endDate string) (*AuditStatistics, error) {
	var resp struct {
		Data    *AuditStatistics `json:"data"`
		Status  string           `json:"status"`
		Message string           `json:"message"`
	}

	queryParams := make(map[string]string)
	if startDate != "" {
		queryParams["start_date"] = startDate
	}
	if endDate != "" {
		queryParams["end_date"] = endDate
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/audit/statistics",
		Result: &resp,
	}
	if len(queryParams) > 0 {
		req.Query = queryParams
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// GetUserAuditHistory retrieves audit history for a specific user
// Authentication: JWT Token required
// Endpoint: GET /v1/audit/users/{userID}/history
// Parameters:
//   - userID: User ID
//   - opts: Optional pagination options
//   - startDate: Optional start date filter
//   - endDate: Optional end date filter
// Returns: Array of AuditLog objects with pagination metadata
func (s *AuditService) GetUserAuditHistory(ctx context.Context, userID uint, opts *PaginationOptions, startDate string, endDate string) ([]AuditLog, *PaginationMeta, error) {
	var resp struct {
		Data []AuditLog      `json:"data"`
		Meta *PaginationMeta `json:"meta"`
	}

	queryParams := make(map[string]string)
	if opts != nil {
		if opts.Page > 0 {
			queryParams["page"] = fmt.Sprintf("%d", opts.Page)
		}
		if opts.Limit > 0 {
			queryParams["limit"] = fmt.Sprintf("%d", opts.Limit)
		}
	}
	if startDate != "" {
		queryParams["start_date"] = startDate
	}
	if endDate != "" {
		queryParams["end_date"] = endDate
	}

	req := &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/audit/users/%d/history", userID),
		Result: &resp,
	}
	if len(queryParams) > 0 {
		req.Query = queryParams
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return resp.Data, resp.Meta, nil
}
