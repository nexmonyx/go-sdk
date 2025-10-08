package nexmonyx

import (
	"context"
	"fmt"
)

// AnalyticsService handles analytics-related operations
// All analytics endpoints use the /v2/analytics prefix
type AnalyticsService struct {
	client *Client
}

// AI Analytics Methods
// These methods provide AI-powered insights and analysis

// GetCapabilities retrieves available AI analytics features
// Authentication: JWT Token required
// Endpoint: GET /v2/analytics/ai/capabilities
// Returns: Available AI features and their status
func (s *AnalyticsService) GetCapabilities(ctx context.Context) (*AICapabilities, error) {
	var resp struct {
		Data    *AICapabilities `json:"data"`
		Status  string          `json:"status"`
		Message string          `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v2/analytics/ai/capabilities",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// AnalyzeMetrics analyzes metrics using AI
// Authentication: JWT Token required
// Endpoint: POST /v2/analytics/ai/analyze
// Parameters:
//   - req: Analysis request with metrics data and context parameters
// Returns: AI-powered analysis results with insights and recommendations
func (s *AnalyticsService) AnalyzeMetrics(ctx context.Context, req *AIAnalysisRequest) (*AIAnalysisResult, error) {
	var resp struct {
		Data    *AIAnalysisResult `json:"data"`
		Status  string            `json:"status"`
		Message string            `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v2/analytics/ai/analyze",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// GetServiceStatus retrieves AI service health status
// Authentication: JWT Token required
// Endpoint: GET /v2/analytics/ai/status
// Returns: AI service health and availability status
func (s *AnalyticsService) GetServiceStatus(ctx context.Context) (*AIServiceStatus, error) {
	var resp struct {
		Data    *AIServiceStatus `json:"data"`
		Status  string           `json:"status"`
		Message string           `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v2/analytics/ai/status",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// Hardware Analytics Methods
// These methods provide hardware health, trends, and predictions

// GetHardwareTrends retrieves historical hardware trends for a server
// Authentication: JWT Token required
// Endpoint: GET /v2/analytics/hardware/trends/{uuid}
// Parameters:
//   - serverUUID: Server UUID
//   - startTime: Start of time range (RFC3339 format)
//   - endTime: End of time range (RFC3339 format)
//   - metricTypes: Optional comma-separated metric types (cpu, memory, disk, network)
// Returns: Historical trends with aggregated metrics
func (s *AnalyticsService) GetHardwareTrends(ctx context.Context, serverUUID, startTime, endTime string, metricTypes ...string) (*HardwareTrends, error) {
	var resp struct {
		Data    *HardwareTrends `json:"data"`
		Status  string          `json:"status"`
		Message string          `json:"message"`
	}

	query := map[string]string{
		"start_time": startTime,
		"end_time":   endTime,
	}
	if len(metricTypes) > 0 {
		query["metric_types"] = metricTypes[0]
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v2/analytics/hardware/trends/" + serverUUID,
		Query:  query,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// GetHardwareHealth retrieves current hardware health score and diagnostics
// Authentication: JWT Token required
// Endpoint: GET /v2/analytics/hardware/health/{uuid}
// Parameters:
//   - serverUUID: Server UUID
// Returns: Current health score, diagnostics, and component status
func (s *AnalyticsService) GetHardwareHealth(ctx context.Context, serverUUID string) (*HardwareHealth, error) {
	var resp struct {
		Data    *HardwareHealth `json:"data"`
		Status  string          `json:"status"`
		Message string          `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v2/analytics/hardware/health/" + serverUUID,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// GetHardwarePredictions retrieves predictive analytics for hardware failures
// Authentication: JWT Token required
// Endpoint: GET /v2/analytics/hardware/predictions/{uuid}
// Parameters:
//   - serverUUID: Server UUID
//   - horizon: Optional prediction horizon in days (default: 30)
// Returns: Predicted failure probabilities and recommended actions
func (s *AnalyticsService) GetHardwarePredictions(ctx context.Context, serverUUID string, horizon int) (*HardwarePrediction, error) {
	var resp struct {
		Data    *HardwarePrediction `json:"data"`
		Status  string              `json:"status"`
		Message string              `json:"message"`
	}

	query := make(map[string]string)
	if horizon > 0 {
		query["horizon"] = fmt.Sprintf("%d", horizon)
	}

	req := &Request{
		Method: "GET",
		Path:   "/v2/analytics/hardware/predictions/" + serverUUID,
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

// Fleet Analytics Methods
// These methods provide organization-wide fleet statistics and metrics

// GetFleetOverview retrieves organization-wide fleet statistics
// Authentication: JWT Token required
// Endpoint: GET /v2/analytics/fleet/overview
// Parameters: Optional query parameters for filtering and aggregation
// Returns: Fleet-wide statistics including server counts, health distribution, resource utilization
func (s *AnalyticsService) GetFleetOverview(ctx context.Context) (*FleetOverview, error) {
	var resp struct {
		Data    *FleetOverview `json:"data"`
		Status  string         `json:"status"`
		Message string         `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v2/analytics/fleet/overview",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// GetOrganizationDashboard retrieves comprehensive dashboard data
// Authentication: JWT Token required
// Endpoint: GET /v2/analytics/fleet/dashboard
// Returns: Comprehensive dashboard with aggregated metrics, server health distribution,
//          alerts summary, and trending data
func (s *AnalyticsService) GetOrganizationDashboard(ctx context.Context) (*OrganizationDashboard, error) {
	var resp struct {
		Data    *OrganizationDashboard `json:"data"`
		Status  string                 `json:"status"`
		Message string                 `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v2/analytics/fleet/dashboard",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// Advanced Analytics Methods
// These methods provide correlation analysis and dependency graphs

// AnalyzeCorrelations analyzes metric correlations across servers
// Authentication: JWT Token required
// Endpoint: POST /v2/analytics/correlation/analyze
// Parameters:
//   - req: Correlation analysis request with metric selection and time ranges
// Returns: Correlation results showing relationships between metrics
func (s *AnalyticsService) AnalyzeCorrelations(ctx context.Context, req *CorrelationAnalysisRequest) (*CorrelationResult, error) {
	var resp struct {
		Data    *CorrelationResult `json:"data"`
		Status  string             `json:"status"`
		Message string             `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v2/analytics/correlation/analyze",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// BuildDependencyGraph builds dependency graph for infrastructure
// Authentication: JWT Token required
// Endpoint: GET /v2/analytics/graph/dependencies
// Parameters: Optional query parameters for graph visualization
// Returns: Dependency graph showing relationships between servers, services, and components
func (s *AnalyticsService) BuildDependencyGraph(ctx context.Context) (*DependencyGraph, error) {
	var resp struct {
		Data    *DependencyGraph `json:"data"`
		Status  string           `json:"status"`
		Message string           `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v2/analytics/graph/dependencies",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}
