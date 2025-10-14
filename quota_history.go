package nexmonyx

import (
	"context"
	"fmt"
	"time"
)

// QuotaHistoryService handles quota usage history operations
type QuotaHistoryService service

// RecordQuotaUsage records quota usage metrics in batch
// Authentication: Admin JWT Token or API Key required
// Endpoint: POST /v1/admin/quota-history/record
// Parameters:
//   - records: Array of quota usage records to store
//
// This method is used by org-management-controller to submit quota usage metrics
// to the API for historical tracking and trend analysis.
func (s *QuotaHistoryService) RecordQuotaUsage(ctx context.Context, records []QuotaUsageRecord) error {
	var resp StandardResponse

	req := QuotaUsageRecordRequest{
		Records: records,
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/admin/quota-history/record",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return err
	}

	return nil
}

// GetHistoricalUsage retrieves historical quota usage records
// Authentication: Admin JWT Token or API Key required
// Endpoint: GET /v1/admin/quota-history/:org_id/usage
// Parameters:
//   - orgID: Organization ID to retrieve usage for
//   - resourceType: Optional resource type filter (cpu, memory, storage, etc.)
//   - startDate: Start of the time range (default: 7 days ago)
//   - endDate: End of the time range (default: now)
func (s *QuotaHistoryService) GetHistoricalUsage(ctx context.Context, orgID uint, resourceType string, startDate, endDate time.Time) ([]QuotaUsageHistory, error) {
	var resp StandardResponse
	var history []QuotaUsageHistory
	resp.Data = &history

	// Build query parameters
	query := make(map[string]string)
	if resourceType != "" {
		query["resource_type"] = resourceType
	}
	if !startDate.IsZero() {
		query["start_date"] = startDate.Format(time.RFC3339)
	}
	if !endDate.IsZero() {
		query["end_date"] = endDate.Format(time.RFC3339)
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/admin/quota-history/%d/usage", orgID),
		Query:  query,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return history, nil
}

// GetAverageUtilization calculates average utilization for a resource type
// Authentication: Admin JWT Token or API Key required
// Endpoint: GET /v1/admin/quota-history/:org_id/average-utilization
// Parameters:
//   - orgID: Organization ID
//   - resourceType: Resource type (cpu, memory, storage, etc.)
//   - startDate: Start of the time range (default: 7 days ago)
//   - endDate: End of the time range (default: now)
func (s *QuotaHistoryService) GetAverageUtilization(ctx context.Context, orgID uint, resourceType string, startDate, endDate time.Time) (*AverageUtilizationResponse, error) {
	var resp StandardResponse
	resp.Data = &AverageUtilizationResponse{}

	// Build query parameters
	query := make(map[string]string)
	query["resource_type"] = resourceType
	if !startDate.IsZero() {
		query["start_date"] = startDate.Format(time.RFC3339)
	}
	if !endDate.IsZero() {
		query["end_date"] = endDate.Format(time.RFC3339)
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/admin/quota-history/%d/average-utilization", orgID),
		Query:  query,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if avgUtil, ok := resp.Data.(*AverageUtilizationResponse); ok {
		return avgUtil, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetPeakUtilization retrieves the highest utilization record for a resource type
// Authentication: Admin JWT Token or API Key required
// Endpoint: GET /v1/admin/quota-history/:org_id/peak-utilization
// Parameters:
//   - orgID: Organization ID
//   - resourceType: Resource type (cpu, memory, storage, etc.)
//   - startDate: Start of the time range (default: 7 days ago)
//   - endDate: End of the time range (default: now)
func (s *QuotaHistoryService) GetPeakUtilization(ctx context.Context, orgID uint, resourceType string, startDate, endDate time.Time) (*QuotaUsageHistory, error) {
	var resp StandardResponse
	resp.Data = &QuotaUsageHistory{}

	// Build query parameters
	query := make(map[string]string)
	query["resource_type"] = resourceType
	if !startDate.IsZero() {
		query["start_date"] = startDate.Format(time.RFC3339)
	}
	if !endDate.IsZero() {
		query["end_date"] = endDate.Format(time.RFC3339)
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/admin/quota-history/%d/peak-utilization", orgID),
		Query:  query,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if peakRecord, ok := resp.Data.(*QuotaUsageHistory); ok {
		return peakRecord, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetDailyAggregates retrieves daily aggregated quota usage statistics
// Authentication: Admin JWT Token or API Key required
// Endpoint: GET /v1/admin/quota-history/:org_id/daily-aggregates
// Parameters:
//   - orgID: Organization ID
//   - resourceType: Resource type (cpu, memory, storage, etc.)
//   - startDate: Start of the time range (default: 30 days ago)
//   - endDate: End of the time range (default: now)
func (s *QuotaHistoryService) GetDailyAggregates(ctx context.Context, orgID uint, resourceType string, startDate, endDate time.Time) ([]DailyAggregateResponse, error) {
	var resp StandardResponse
	var aggregates []DailyAggregateResponse
	resp.Data = &aggregates

	// Build query parameters
	query := make(map[string]string)
	query["resource_type"] = resourceType
	if !startDate.IsZero() {
		query["start_date"] = startDate.Format(time.RFC3339)
	}
	if !endDate.IsZero() {
		query["end_date"] = endDate.Format(time.RFC3339)
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/admin/quota-history/%d/daily-aggregates", orgID),
		Query:  query,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return aggregates, nil
}

// GetResourceSummary retrieves summary statistics for all resource types
// Authentication: Admin JWT Token or API Key required
// Endpoint: GET /v1/admin/quota-history/:org_id/resource-summary
// Parameters:
//   - orgID: Organization ID
//   - startDate: Start date for averages (default: 7 days ago)
//   - endDate: End date for averages (default: now)
func (s *QuotaHistoryService) GetResourceSummary(ctx context.Context, orgID uint, startDate, endDate time.Time) ([]ResourceSummaryResponse, error) {
	var resp StandardResponse
	var summaries []ResourceSummaryResponse
	resp.Data = &summaries

	// Build query parameters
	query := make(map[string]string)
	if !startDate.IsZero() {
		query["start_date"] = startDate.Format(time.RFC3339)
	}
	if !endDate.IsZero() {
		query["end_date"] = endDate.Format(time.RFC3339)
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/admin/quota-history/%d/resource-summary", orgID),
		Query:  query,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return summaries, nil
}

// GetUsageTrend analyzes quota usage trends using linear regression
// Authentication: Admin JWT Token or API Key required
// Endpoint: GET /v1/admin/quota-history/:org_id/usage-trend
// Parameters:
//   - orgID: Organization ID
//   - resourceType: Resource type (cpu, memory, storage, etc.)
//   - days: Number of days to analyze (default: 7)
func (s *QuotaHistoryService) GetUsageTrend(ctx context.Context, orgID uint, resourceType string, days int) (*UsageTrendResponse, error) {
	var resp StandardResponse
	resp.Data = &UsageTrendResponse{}

	// Build query parameters
	query := make(map[string]string)
	query["resource_type"] = resourceType
	if days > 0 {
		query["days"] = fmt.Sprintf("%d", days)
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/admin/quota-history/%d/usage-trend", orgID),
		Query:  query,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if trendResponse, ok := resp.Data.(*UsageTrendResponse); ok {
		return trendResponse, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// DetectUsagePatterns analyzes quota usage to detect anomalies and patterns
// Authentication: Admin JWT Token or API Key required
// Endpoint: GET /v1/admin/quota-history/:org_id/detect-patterns
// Parameters:
//   - orgID: Organization ID
//   - days: Number of days to analyze (default: 7)
func (s *QuotaHistoryService) DetectUsagePatterns(ctx context.Context, orgID uint, days int) (*UsagePatternsResponse, error) {
	var resp StandardResponse
	resp.Data = &UsagePatternsResponse{}

	// Build query parameters
	query := make(map[string]string)
	if days > 0 {
		query["days"] = fmt.Sprintf("%d", days)
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/admin/quota-history/%d/detect-patterns", orgID),
		Query:  query,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if patternsResponse, ok := resp.Data.(*UsagePatternsResponse); ok {
		return patternsResponse, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// CleanupOldRecords deletes quota history records older than specified retention days
// Authentication: Admin JWT Token or API Key required
// Endpoint: DELETE /v1/admin/quota-history/:org_id/cleanup
// Parameters:
//   - orgID: Organization ID
//   - retentionDays: Number of days to retain (default: 90, minimum: 7)
func (s *QuotaHistoryService) CleanupOldRecords(ctx context.Context, orgID uint, retentionDays int) (int, error) {
	var resp struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    struct {
			DeletedCount   int    `json:"deleted_count"`
			RetentionDays  int    `json:"retention_days"`
			CutoffDate     string `json:"cutoff_date"`
		} `json:"data"`
	}

	// Build query parameters
	query := make(map[string]string)
	if retentionDays > 0 {
		query["retention_days"] = fmt.Sprintf("%d", retentionDays)
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v1/admin/quota-history/%d/cleanup", orgID),
		Query:  query,
		Result: &resp,
	})
	if err != nil {
		return 0, err
	}

	return resp.Data.DeletedCount, nil
}
