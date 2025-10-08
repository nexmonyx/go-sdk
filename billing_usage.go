package nexmonyx

import (
	"context"
	"fmt"
	"time"
)

// GetMyCurrentUsage retrieves the current usage metrics for the authenticated user's organization
// Authentication: JWT Token required
// Endpoint: GET /v1/billing/usage/current
func (s *BillingUsageService) GetMyCurrentUsage(ctx context.Context) (*OrganizationUsageMetrics, error) {
	var resp StandardResponse
	resp.Data = &OrganizationUsageMetrics{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/billing/usage/current",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if usage, ok := resp.Data.(*OrganizationUsageMetrics); ok {
		return usage, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetMyUsageHistory retrieves historical usage metrics for the authenticated user's organization
// Authentication: JWT Token required
// Endpoint: GET /v1/billing/usage/history
// Parameters:
//   - startDate: Start of the time range (default: 30 days ago)
//   - endDate: End of the time range (default: now)
//   - interval: Aggregation interval - "hourly", "daily", or "monthly" (default: "daily")
func (s *BillingUsageService) GetMyUsageHistory(ctx context.Context, startDate, endDate time.Time, interval string) ([]UsageMetricsHistory, error) {
	var resp StandardResponse
	var history []UsageMetricsHistory
	resp.Data = &history

	// Build query parameters
	query := make(map[string]string)
	if !startDate.IsZero() {
		query["start_date"] = startDate.Format(time.RFC3339)
	}
	if !endDate.IsZero() {
		query["end_date"] = endDate.Format(time.RFC3339)
	}
	if interval != "" {
		query["interval"] = interval
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/billing/usage/history",
		Query:  query,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return history, nil
}

// GetMyUsageSummary retrieves aggregated usage summary for the authenticated user's organization
// Authentication: JWT Token required
// Endpoint: GET /v1/billing/usage/summary
// Parameters:
//   - startDate: Start of the time range (default: 1 month ago)
//   - endDate: End of the time range (default: now)
func (s *BillingUsageService) GetMyUsageSummary(ctx context.Context, startDate, endDate time.Time) (*UsageSummary, error) {
	var resp StandardResponse
	resp.Data = &UsageSummary{}

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
		Path:   "/v1/billing/usage/summary",
		Query:  query,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if summary, ok := resp.Data.(*UsageSummary); ok {
		return summary, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetOrgCurrentUsage retrieves the current usage metrics for a specific organization (admin only)
// Authentication: Admin JWT Token or API Key required
// Endpoint: GET /v1/admin/billing/organizations/:id/usage
// Parameters:
//   - orgID: Organization ID to retrieve usage for
func (s *BillingUsageService) GetOrgCurrentUsage(ctx context.Context, orgID uint) (*OrganizationUsageMetrics, error) {
	var resp StandardResponse
	resp.Data = &OrganizationUsageMetrics{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/admin/billing/organizations/%d/usage", orgID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if usage, ok := resp.Data.(*OrganizationUsageMetrics); ok {
		return usage, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetOrgUsageHistory retrieves historical usage metrics for a specific organization (admin only)
// Authentication: Admin JWT Token or API Key required
// Endpoint: GET /v1/admin/billing/organizations/:id/usage/history
// Parameters:
//   - orgID: Organization ID to retrieve usage for
//   - startDate: Start of the time range (default: 30 days ago)
//   - endDate: End of the time range (default: now)
//   - interval: Aggregation interval - "hourly", "daily", or "monthly" (default: "daily")
func (s *BillingUsageService) GetOrgUsageHistory(ctx context.Context, orgID uint, startDate, endDate time.Time, interval string) ([]UsageMetricsHistory, error) {
	var resp StandardResponse
	var history []UsageMetricsHistory
	resp.Data = &history

	// Build query parameters
	query := make(map[string]string)
	if !startDate.IsZero() {
		query["start_date"] = startDate.Format(time.RFC3339)
	}
	if !endDate.IsZero() {
		query["end_date"] = endDate.Format(time.RFC3339)
	}
	if interval != "" {
		query["interval"] = interval
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/admin/billing/organizations/%d/usage/history", orgID),
		Query:  query,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return history, nil
}

// GetOrgUsageSummary retrieves aggregated usage summary for a specific organization (admin only)
// Authentication: Admin JWT Token or API Key required
// Endpoint: GET /v1/admin/billing/organizations/:id/usage/summary
// Parameters:
//   - orgID: Organization ID to retrieve usage for
//   - startDate: Start of the time range (default: 30 days ago)
//   - endDate: End of the time range (default: now)
func (s *BillingUsageService) GetOrgUsageSummary(ctx context.Context, orgID uint, startDate, endDate time.Time) (*UsageSummary, error) {
	var resp StandardResponse
	resp.Data = &UsageSummary{}

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
		Path:   fmt.Sprintf("/v1/admin/billing/organizations/%d/usage/summary", orgID),
		Query:  query,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if summary, ok := resp.Data.(*UsageSummary); ok {
		return summary, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetAllUsageOverview retrieves usage overview for all organizations (admin only)
// Authentication: Admin JWT Token or API Key required
// Endpoint: GET /v1/admin/billing/usage/overview
// Parameters:
//   - opts: Pagination options (page, limit)
//
// Returns organization usage metrics with pagination metadata.
func (s *BillingUsageService) GetAllUsageOverview(ctx context.Context, opts *ListOptions) (*OrganizationUsageOverview, *PaginationMeta, error) {
	var resp struct {
		Status     string                      `json:"status"`
		Message    string                      `json:"message"`
		Data       *OrganizationUsageOverview  `json:"data"`
		Pagination *PaginationMeta             `json:"pagination"`
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/admin/billing/usage/overview",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return resp.Data, resp.Pagination, nil
}
