package nexmonyx

import (
	"context"
	"fmt"
	"time"
)

// CreateProbe creates a new monitoring probe
func (s *MonitoringService) CreateProbe(ctx context.Context, probe *MonitoringProbe) (*MonitoringProbe, error) {
	var resp StandardResponse
	resp.Data = &MonitoringProbe{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/api/v1/monitoring/probes",
		Body:   probe,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if created, ok := resp.Data.(*MonitoringProbe); ok {
		return created, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetProbe retrieves a monitoring probe by ID
func (s *MonitoringService) GetProbe(ctx context.Context, id string) (*MonitoringProbe, error) {
	var resp StandardResponse
	resp.Data = &MonitoringProbe{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/api/v1/monitoring/probes/%s", id),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if probe, ok := resp.Data.(*MonitoringProbe); ok {
		return probe, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// ListProbes retrieves a list of monitoring probes
func (s *MonitoringService) ListProbes(ctx context.Context, opts *ListOptions) ([]*MonitoringProbe, *PaginationMeta, error) {
	var resp PaginatedResponse
	var probes []*MonitoringProbe
	resp.Data = &probes

	req := &Request{
		Method: "GET",
		Path:   "/api/v1/monitoring/probes",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return probes, resp.Meta, nil
}

// UpdateProbe updates a monitoring probe
func (s *MonitoringService) UpdateProbe(ctx context.Context, id string, probe *MonitoringProbe) (*MonitoringProbe, error) {
	var resp StandardResponse
	resp.Data = &MonitoringProbe{}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/api/v1/monitoring/probes/%s", id),
		Body:   probe,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if updated, ok := resp.Data.(*MonitoringProbe); ok {
		return updated, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// DeleteProbe deletes a monitoring probe
func (s *MonitoringService) DeleteProbe(ctx context.Context, id string) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/api/v1/monitoring/probes/%s", id),
		Result: &resp,
	})
	return err
}

// GetProbeResults retrieves test results for a probe
func (s *MonitoringService) GetProbeResults(ctx context.Context, probeID string, opts *ListOptions) ([]*ProbeTestResult, *PaginationMeta, error) {
	var resp PaginatedResponse
	var results []*ProbeTestResult
	resp.Data = &results

	req := &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/api/v1/monitoring/probes/%s/results", probeID),
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return results, resp.Meta, nil
}

// GetMonitoringAgents retrieves monitoring agents
func (s *MonitoringService) GetAgents(ctx context.Context, opts *ListOptions) ([]*MonitoringAgent, *PaginationMeta, error) {
	var resp PaginatedResponse
	var agents []*MonitoringAgent
	resp.Data = &agents

	req := &Request{
		Method: "GET",
		Path:   "/api/v1/monitoring/agents",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return agents, resp.Meta, nil
}

// GetMonitoringStatus retrieves monitoring status for an organization
func (s *MonitoringService) GetStatus(ctx context.Context, organizationID string) (*MonitoringStatus, error) {
	var resp StandardResponse
	resp.Data = &MonitoringStatus{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/api/v1/monitoring/organizations/%s/status", organizationID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if status, ok := resp.Data.(*MonitoringStatus); ok {
		return status, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// TestProbe manually triggers a probe test
func (s *MonitoringService) TestProbe(ctx context.Context, probeID string) (*ProbeTestResult, error) {
	var resp StandardResponse
	resp.Data = &ProbeTestResult{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/api/v1/monitoring/probes/%s/test", probeID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if result, ok := resp.Data.(*ProbeTestResult); ok {
		return result, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// MonitoringProbe represents a monitoring probe configuration
type MonitoringProbe struct {
	GormModel
	Name           string                 `json:"name"`
	Description    string                 `json:"description,omitempty"`
	Type           string                 `json:"type"` // http, https, tcp, icmp, dns
	Target         string                 `json:"target"`
	Interval       int                    `json:"interval"` // seconds
	Timeout        int                    `json:"timeout"`  // seconds
	Enabled        bool                   `json:"enabled"`
	OrganizationID uint                   `json:"organization_id"`
	ServerID       *uint                  `json:"server_id,omitempty"`
	Regions        []string               `json:"regions,omitempty"`
	Config         map[string]interface{} `json:"config,omitempty"`
	AlertConfig    *ProbeAlertConfig      `json:"alert_config,omitempty"`
	Tags           []string               `json:"tags,omitempty"`
}

// ProbeAlertConfig represents alert configuration for a probe
type ProbeAlertConfig struct {
	Enabled           bool     `json:"enabled"`
	FailureThreshold  int      `json:"failure_threshold"`
	SuccessThreshold  int      `json:"success_threshold"`
	NotificationDelay int      `json:"notification_delay,omitempty"`
	Channels          []string `json:"channels,omitempty"`
	Recipients        []string `json:"recipients,omitempty"`
}

// MonitoringStatus represents the monitoring status for an organization
type MonitoringStatus struct {
	ActiveProbes    int                    `json:"active_probes"`
	TotalProbes     int                    `json:"total_probes"`
	ActiveAgents    int                    `json:"active_agents"`
	TotalAgents     int                    `json:"total_agents"`
	HealthyProbes   int                    `json:"healthy_probes"`
	FailingProbes   int                    `json:"failing_probes"`
	ProbesByType    map[string]int         `json:"probes_by_type"`
	ProbesByRegion  map[string]int         `json:"probes_by_region"`
	RecentIncidents []MonitoringIncident   `json:"recent_incidents,omitempty"`
	Metrics         map[string]interface{} `json:"metrics,omitempty"`
}

// MonitoringIncident represents a monitoring incident
type MonitoringIncident struct {
	ID         uint        `json:"id"`
	ProbeID    uint        `json:"probe_id"`
	ProbeName  string      `json:"probe_name"`
	StartedAt  *CustomTime `json:"started_at"`
	ResolvedAt *CustomTime `json:"resolved_at,omitempty"`
	Duration   int         `json:"duration,omitempty"`
	Status     string      `json:"status"`
	Reason     string      `json:"reason"`
	Details    string      `json:"details,omitempty"`
}

// GetAgentStatus retrieves the status of a monitoring agent
func (s *MonitoringService) GetAgentStatus(ctx context.Context, agentID string) (*AgentStatusResponse, error) {
	var resp StandardResponse
	resp.Data = &AgentStatusResponse{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/api/v1/monitoring/agents/%s/status", agentID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if status, ok := resp.Data.(*AgentStatusResponse); ok {
		return status, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// AgentStatusResponse represents the status response for a monitoring agent
type AgentStatusResponse struct {
	AgentID       string      `json:"agent_id"`
	Status        string      `json:"status"`
	LastHeartbeat *CustomTime `json:"last_heartbeat"`
	Uptime        float64     `json:"uptime"`
	ProbesRunning int         `json:"probes_running"`
	ProbesFailed  int64       `json:"probes_failed"`
	ProbesSuccess int64       `json:"probes_success"`
	ErrorRate     float64     `json:"error_rate"`
}

// ListAgents retrieves a list of monitoring agents
func (s *MonitoringService) ListAgents(ctx context.Context, opts *MonitoringAgentListOptions) ([]*MonitoringAgent, *PaginationMeta, error) {
	var resp PaginatedResponse
	var agents []*MonitoringAgent
	resp.Data = &agents

	req := &Request{
		Method: "GET",
		Path:   "/api/v1/monitoring/agents",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return agents, resp.Meta, nil
}

// RegisterAgent registers a new monitoring agent
func (s *MonitoringService) RegisterAgent(ctx context.Context, registration *AgentRegistration) (*MonitoringAgent, error) {
	var resp StandardResponse
	resp.Data = &MonitoringAgent{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/api/v1/monitoring/agents",
		Body:   registration,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if agent, ok := resp.Data.(*MonitoringAgent); ok {
		return agent, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// MonitoringAgentListOptions represents options for listing monitoring agents
type MonitoringAgentListOptions struct {
	ListOptions
	Status  string `url:"status,omitempty"`
	Region  string `url:"region,omitempty"`
	Type    string `url:"type,omitempty"`
	Enabled *bool  `url:"enabled,omitempty"`
}

// ToQuery converts options to query parameters
func (o *MonitoringAgentListOptions) ToQuery() map[string]string {
	params := o.ListOptions.ToQuery()
	if o.Status != "" {
		params["status"] = o.Status
	}
	if o.Region != "" {
		params["region"] = o.Region
	}
	if o.Type != "" {
		params["type"] = o.Type
	}
	if o.Enabled != nil {
		params["enabled"] = fmt.Sprintf("%t", *o.Enabled)
	}
	return params
}

// AgentRegistration represents a monitoring agent registration request
type AgentRegistration struct {
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Region       string                 `json:"region"`
	Location     string                 `json:"location,omitempty"`
	Provider     string                 `json:"provider,omitempty"`
	Version      string                 `json:"version"`
	Capabilities []string               `json:"capabilities"`
	Config       map[string]interface{} `json:"config,omitempty"`
	MaxProbes    int                    `json:"max_probes,omitempty"`
}

// MonitoringDeployment represents a monitoring deployment
type MonitoringDeployment struct {
	ID             uint      `json:"id"`
	OrganizationID uint      `json:"organization_id"`
	Region         string    `json:"region"`
	NamespaceName  string    `json:"namespace_name"`
	DeploymentName string    `json:"deployment_name"`
	Status         string    `json:"status"`
	ErrorCount     int       `json:"error_count"`
	CurrentVersion string    `json:"current_version,omitempty"`
	TargetVersion  string    `json:"target_version,omitempty"`
	CreatedAt      time.Time `json:"created_at,omitempty"`
	UpdatedAt      time.Time `json:"updated_at,omitempty"`
}

// ListDeployments retrieves a list of monitoring deployments
func (s *MonitoringService) ListDeployments(ctx context.Context, opts *MonitoringDeploymentListOptions) ([]*MonitoringDeployment, *PaginationMeta, error) {
	var resp PaginatedResponse
	var deployments []*MonitoringDeployment
	resp.Data = &deployments

	req := &Request{
		Method: "GET",
		Path:   "/api/v1/monitoring/deployments",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return deployments, resp.Meta, nil
}

// MonitoringDeploymentListOptions represents options for listing monitoring deployments
type MonitoringDeploymentListOptions struct {
	ListOptions
	Environment string `url:"environment,omitempty"`
	Region      string `url:"region,omitempty"`
	Status      string `url:"status,omitempty"`
}

// ToQuery converts options to query parameters
func (o *MonitoringDeploymentListOptions) ToQuery() map[string]string {
	params := o.ListOptions.ToQuery()
	if o.Environment != "" {
		params["environment"] = o.Environment
	}
	if o.Region != "" {
		params["region"] = o.Region
	}
	if o.Status != "" {
		params["status"] = o.Status
	}
	return params
}

// ProbeResult represents a probe test result
type ProbeResult struct {
	ProbeID      uint                `json:"probe_id"`
	ProbeUUID    string              `json:"probe_uuid"`
	Region       string              `json:"region"`
	Status       string              `json:"status"`
	ResponseTime int                 `json:"response_time"`
	ExecutedAt   *CustomTime         `json:"executed_at"`
	StatusCode   int                 `json:"status_code,omitempty"`
	Error        string              `json:"error,omitempty"`
	Details      *ProbeResultDetails `json:"details,omitempty"`
}

// ProbeResultDetails represents detailed probe result information
type ProbeResultDetails struct {
	StatusCode   *int  `json:"status_code,omitempty"`
	ResponseSize *int  `json:"response_size,omitempty"`
	ContentMatch *bool `json:"content_match,omitempty"`
	DNSTime      *int  `json:"dns_time,omitempty"`
	ConnectTime  *int  `json:"connect_time,omitempty"`
	TLSTime      *int  `json:"tls_time,omitempty"`
}

// ListProbeResults retrieves a list of probe results
func (s *MonitoringService) ListProbeResults(ctx context.Context, opts *ProbeResultListOptions) ([]*ProbeResult, *PaginationMeta, error) {
	var resp PaginatedResponse
	var results []*ProbeResult
	resp.Data = &results

	req := &Request{
		Method: "GET",
		Path:   "/api/v1/monitoring/probe-results",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return results, resp.Meta, nil
}

// ProbeResultListOptions represents options for listing probe results
type ProbeResultListOptions struct {
	ListOptions
	ProbeUUID string `url:"probe_uuid,omitempty"`
	Status    string `url:"status,omitempty"`
	Region    string `url:"region,omitempty"`
}

// ToQuery converts options to query parameters
func (o *ProbeResultListOptions) ToQuery() map[string]string {
	params := o.ListOptions.ToQuery()
	if o.ProbeUUID != "" {
		params["probe_uuid"] = o.ProbeUUID
	}
	if o.Status != "" {
		params["status"] = o.Status
	}
	if o.Region != "" {
		params["region"] = o.Region
	}
	return params
}

// ProbeMetrics represents probe metrics
type ProbeMetrics struct {
	ProbeID          uint        `json:"probe_id,omitempty"`
	ProbeUUID        string      `json:"probe_uuid"`
	Region           string      `json:"region,omitempty"`
	Uptime           float64     `json:"uptime,omitempty"`
	UptimePercentage float64     `json:"uptime_percentage,omitempty"`
	AvgResponseTime  float64     `json:"avg_response_time"`
	SuccessRate      float64     `json:"success_rate,omitempty"`
	TotalTests       int64       `json:"total_tests,omitempty"`
	TotalChecks      int64       `json:"total_checks,omitempty"`
	SuccessfulTests  int64       `json:"successful_tests,omitempty"`
	SuccessfulChecks int64       `json:"successful_checks,omitempty"`
	FailedTests      int64       `json:"failed_tests,omitempty"`
	FailedChecks     int64       `json:"failed_checks,omitempty"`
	LastCheck        *CustomTime `json:"last_check,omitempty"`
	LastStatus       string      `json:"last_status,omitempty"`
}

// GetProbeMetrics retrieves metrics for a specific probe
func (s *MonitoringService) GetProbeMetrics(ctx context.Context, probeUUID string, timeRange ...*TimeRange) (*ProbeMetrics, error) {
	var resp StandardResponse
	resp.Data = &ProbeMetrics{}

	req := &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/api/v1/monitoring/probes/%s/metrics", probeUUID),
		Result: &resp,
	}

	if len(timeRange) > 0 && timeRange[0] != nil {
		req.Query = map[string]string{
			"start": timeRange[0].Start,
			"end":   timeRange[0].End,
		}
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	if metrics, ok := resp.Data.(*ProbeMetrics); ok {
		return metrics, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// ProbeRequest represents a request to create or update a monitoring probe
type ProbeRequest struct {
	Name           string       `json:"name"`
	Description    string       `json:"description,omitempty"`
	Type           string       `json:"type"`
	Scope          string       `json:"scope,omitempty"`
	Target         string       `json:"target"`
	Interval       int          `json:"interval"`
	Timeout        int          `json:"timeout"`
	Enabled        bool         `json:"enabled"`
	Config         *ProbeConfig `json:"config,omitempty"`
	Regions        []string     `json:"regions,omitempty"`
	AlertThreshold int          `json:"alert_threshold,omitempty"`
	AlertEnabled   bool         `json:"alert_enabled,omitempty"`
}

// ProbeConfig represents the configuration for a monitoring probe
type ProbeConfig struct {
	Method             *string           `json:"method,omitempty"`
	ExpectedStatusCode *int              `json:"expected_status_code,omitempty"`
	FollowRedirects    *bool             `json:"follow_redirects,omitempty"`
	Headers            map[string]string `json:"headers,omitempty"`
	Body               *string           `json:"body,omitempty"`
	UserAgent          *string           `json:"user_agent,omitempty"`
	Keyword            *string           `json:"keyword,omitempty"`
	Port               *int              `json:"port,omitempty"`
}

// ProbeAlertChannel represents an alert channel for a probe
type ProbeAlertChannel struct {
	ProbeID uint         `json:"probe_id"`
	Type    string       `json:"type"`
	Name    string       `json:"name"`
	Enabled bool         `json:"enabled"`
	Config  *AlertConfig `json:"config,omitempty"`
}

// AlertConfig represents alert configuration
type AlertConfig struct {
	Recipients []string `json:"recipients,omitempty"`
	Webhook    string   `json:"webhook,omitempty"`
	SlackToken string   `json:"slack_token,omitempty"`
	Channel    string   `json:"channel,omitempty"`
}
