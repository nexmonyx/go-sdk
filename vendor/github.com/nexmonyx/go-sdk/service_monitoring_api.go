package nexmonyx

import (
	"context"
	"fmt"
)

// ServiceMonitoringService handles service monitoring related API operations
type ServiceMonitoringService struct {
	client *Client
}

// ServiceMonitoringRequest represents a request to submit service monitoring data
type ServiceMonitoringRequest struct {
	ServerUUID string       `json:"server_uuid"`
	Services   *ServiceInfo `json:"services"`
}

// ServiceMetricsRequest represents a request to submit service metrics
type ServiceMetricsRequest struct {
	ServerUUID string            `json:"server_uuid"`
	Metrics    []*ServiceMetrics `json:"metrics"`
}

// ServiceLogsRequest represents a request to submit service logs
type ServiceLogsRequest struct {
	ServerUUID string                        `json:"server_uuid"`
	Logs       map[string][]ServiceLogEntry `json:"logs"`
}

// ServiceStatusResponse represents the response from getting service status
type ServiceStatusResponse struct {
	ServerUUID   string                 `json:"server_uuid"`
	LastUpdated  string                 `json:"last_updated"`
	Services     []*ServiceMonitoringInfo  `json:"services"`
	Summary      ServiceStatusSummary   `json:"summary"`
}

// ServiceStatusSummary provides a summary of service states
type ServiceStatusSummary struct {
	Total        int            `json:"total"`
	Active       int            `json:"active"`
	Inactive     int            `json:"inactive"`
	Failed       int            `json:"failed"`
	StateCounts  map[string]int `json:"state_counts"`
}

// ServiceHistoryResponse represents historical service data
type ServiceHistoryResponse struct {
	ServerUUID   string                `json:"server_uuid"`
	ServiceName  string                `json:"service_name"`
	History      []ServiceHistoryEntry `json:"history"`
}

// ServiceHistoryEntry represents a point in service history
type ServiceHistoryEntry struct {
	Timestamp    string  `json:"timestamp"`
	State        string  `json:"state"`
	SubState     string  `json:"sub_state"`
	CPUPercent   float64 `json:"cpu_percent,omitempty"`
	MemoryBytes  uint64  `json:"memory_bytes,omitempty"`
	RestartCount int     `json:"restart_count,omitempty"`
}

// SubmitServiceData submits service monitoring data
func (s *ServiceMonitoringService) SubmitServiceData(ctx context.Context, serverID string, services *ServiceInfo) error {
	req := &ServiceMonitoringRequest{
		ServerUUID: serverID,
		Services:   services,
	}

	path := fmt.Sprintf("/v1/servers/%s/services", serverID)
	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   path,
		Body:   req,
	})
	return err
}

// SubmitServiceMetrics submits service metrics as time-series data
func (s *ServiceMonitoringService) SubmitServiceMetrics(ctx context.Context, serverID string, metrics []*ServiceMetrics) error {
	req := &ServiceMetricsRequest{
		ServerUUID: serverID,
		Metrics:    metrics,
	}

	path := fmt.Sprintf("/v1/servers/%s/services/metrics", serverID)
	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   path,
		Body:   req,
	})
	return err
}

// SubmitServiceLogs submits service logs
func (s *ServiceMonitoringService) SubmitServiceLogs(ctx context.Context, serverID string, logs map[string][]ServiceLogEntry) error {
	req := &ServiceLogsRequest{
		ServerUUID: serverID,
		Logs:       logs,
	}

	path := fmt.Sprintf("/v1/servers/%s/services/logs", serverID)
	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   path,
		Body:   req,
	})
	return err
}

// GetServerServices gets the current service status for a server
func (s *ServiceMonitoringService) GetServerServices(ctx context.Context, serverID string) (*ServiceStatusResponse, error) {
	path := fmt.Sprintf("/v1/servers/%s/services", serverID)
	
	var response ServiceStatusResponse
	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   path,
		Result: &response,
	})
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// GetServiceHistory gets historical data for a specific service
func (s *ServiceMonitoringService) GetServiceHistory(ctx context.Context, serverID string, serviceName string, opts *ListOptions) (*ServiceHistoryResponse, error) {
	path := fmt.Sprintf("/v1/servers/%s/services/%s/history", serverID, serviceName)
	
	var query map[string]string
	if opts != nil {
		query = opts.ToQuery()
	}

	var response ServiceHistoryResponse
	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   path,
		Query:  query,
		Result: &response,
	})
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// RestartService requests a service restart (requires appropriate permissions)
func (s *ServiceMonitoringService) RestartService(ctx context.Context, serverID string, serviceName string) error {
	path := fmt.Sprintf("/v1/servers/%s/services/%s/restart", serverID, serviceName)
	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   path,
	})
	return err
}

// GetServiceLogs retrieves logs for a specific service
func (s *ServiceMonitoringService) GetServiceLogs(ctx context.Context, serverID string, serviceName string, opts *ListOptions) ([]ServiceLogEntry, error) {
	path := fmt.Sprintf("/v1/servers/%s/services/%s/logs", serverID, serviceName)
	
	var query map[string]string
	if opts != nil {
		query = opts.ToQuery()
	}

	var logs []ServiceLogEntry
	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   path,
		Query:  query,
		Result: &logs,
	})
	if err != nil {
		return nil, err
	}

	return logs, nil
}

// GetFailedServices returns all failed services across servers in an organization
func (s *ServiceMonitoringService) GetFailedServices(ctx context.Context, organizationID string) ([]*ServiceMonitoringInfo, error) {
	path := fmt.Sprintf("/v1/organizations/%s/services/failed", organizationID)
	
	var services []*ServiceMonitoringInfo
	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   path,
		Result: &services,
	})
	if err != nil {
		return nil, err
	}

	return services, nil
}

// CreateServiceAlert creates an alert rule for service monitoring
func (s *ServiceMonitoringService) CreateServiceAlert(ctx context.Context, organizationID string, alertConfig ServiceAlertConfig) error {
	path := fmt.Sprintf("/v1/organizations/%s/alerts/services", organizationID)
	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   path,
		Body:   alertConfig,
	})
	return err
}

// ServiceAlertConfig represents configuration for service alerts
type ServiceAlertConfig struct {
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	ServicePatterns []string `json:"service_patterns"`
	Conditions      []string `json:"conditions"` // e.g., "state=failed", "restart_count>3"
	Severity        string   `json:"severity"`    // info, warning, critical
	Enabled         bool     `json:"enabled"`
}