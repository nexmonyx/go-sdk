package nexmonyx

import (
	"context"
	"fmt"
)

// ProbesService is defined in client.go

// Create creates a new probe
func (s *ProbesService) Create(ctx context.Context, req *ProbeCreateRequest) (*MonitoringProbe, error) {
	// Convert ProbeCreateRequest to map to match API expectations
	config := make(map[string]interface{})

	// Based on probe type, set appropriate config fields
	switch req.Type {
	case "icmp":
		if req.Target != "" {
			config["host"] = req.Target
		}
	case "http", "https":
		if req.Target != "" {
			config["url"] = req.Target
		}
	case "tcp":
		if req.Target != "" {
			config["host"] = req.Target
		}
		if tcpPort, ok := req.Configuration["port"]; ok {
			config["port"] = tcpPort
		}
	case "heartbeat":
		if req.Target != "" {
			config["url"] = req.Target
		}
	}

	// Add any additional config from the request
	if req.Configuration != nil {
		for k, v := range req.Configuration {
			config[k] = v
		}
	}

	// Create the request body matching API expectations
	body := map[string]interface{}{
		"name":      req.Name,
		"type":      req.Type,
		"frequency": req.Interval,             // API expects "frequency", SDK has "interval"
		"regions":   []string{req.RegionCode}, // Convert single region to array
		"enabled":   req.Enabled,
		"config":    config,
	}

	// Add description if provided
	if req.Name != "" {
		body["description"] = req.Name // Use name as description if not provided
	}

	var result struct {
		Status string `json:"status"`
		Data   struct {
			Probe MonitoringProbe `json:"probe"`
		} `json:"data"`
		Message string `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/probes",
		Body:   body,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}

	return &result.Data.Probe, nil
}

// List returns all probes
func (s *ProbesService) List(ctx context.Context, opts *ListOptions) ([]*MonitoringProbe, *PaginationMeta, error) {
	var resp PaginatedResponse
	var probes []*MonitoringProbe
	resp.Data = &probes

	req := &Request{
		Method: "GET",
		Path:   "/v2/probes",
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

// Get retrieves a probe by UUID
func (s *ProbesService) Get(ctx context.Context, uuid string) (*MonitoringProbe, error) {
	var resp StandardResponse
	resp.Data = &MonitoringProbe{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v2/probes/%s", uuid),
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

// Update updates a probe
func (s *ProbesService) Update(ctx context.Context, uuid string, req *ProbeUpdateRequest) (*MonitoringProbe, error) {
	// Build update request body
	body := make(map[string]interface{})

	if req.Name != nil {
		body["name"] = *req.Name
	}
	if req.Enabled != nil {
		body["enabled"] = *req.Enabled
	}
	if req.Interval != nil {
		body["frequency"] = *req.Interval
	}
	if req.Configuration != nil {
		body["config"] = req.Configuration
	}

	var resp StandardResponse
	resp.Data = &MonitoringProbe{}

	_, err := s.client.Do(ctx, &Request{
		Method: "PATCH",
		Path:   fmt.Sprintf("/v2/probes/%s", uuid),
		Body:   body,
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

// Delete removes a probe
func (s *ProbesService) Delete(ctx context.Context, uuid string) error {
	_, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v2/probes/%s", uuid),
	})
	return err
}

// GetHealth returns the health status of a probe
func (s *ProbesService) GetHealth(ctx context.Context, uuid string) (*ProbeHealth, error) {
	var result struct {
		Status string       `json:"status"`
		Data   *ProbeHealth `json:"data"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/probes/%s/health", uuid),
		Result: &result,
	})
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

// ListResults returns probe execution results
func (s *ProbesService) ListResults(ctx context.Context, uuid string, opts *ProbeResultListOptions) ([]*ProbeResult, *PaginationMeta, error) {
	return s.client.Monitoring.ListProbeResults(ctx, opts)
}

// GetAvailableRegions returns available monitoring regions
func (s *ProbesService) GetAvailableRegions(ctx context.Context) ([]*MonitoringRegion, error) {
	var result struct {
		Status string              `json:"status"`
		Data   []*MonitoringRegion `json:"data"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/monitoring/regions",
		Result: &result,
	})
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

// GetAvailableProbeTypes returns available probe types
func (s *ProbesService) GetAvailableProbeTypes(ctx context.Context) ([]string, error) {
	// For now, return static list
	return []string{"icmp", "http", "https", "tcp", "heartbeat"}, nil
}

// CreateSimpleProbe creates a probe with simpler parameters
func (s *ProbesService) CreateSimpleProbe(ctx context.Context, name, probeType, target string, regions []string) (*MonitoringProbe, error) {
	// Convert to API format
	config := make(map[string]interface{})

	// Based on probe type, set appropriate config fields
	switch probeType {
	case "icmp":
		config["host"] = target
	case "http", "https":
		config["url"] = target
	case "tcp":
		config["host"] = target
		config["port"] = 80 // Default port
	case "heartbeat":
		config["url"] = target
	}

	// Create the request body matching API expectations
	body := map[string]interface{}{
		"name":        name,
		"description": name,
		"type":        probeType,
		"frequency":   300, // 5 minutes default
		"regions":     regions,
		"enabled":     true,
		"config":      config,
	}

	var result struct {
		Status string `json:"status"`
		Data   struct {
			Probe MonitoringProbe `json:"probe"`
		} `json:"data"`
		Message string `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/probes",
		Body:   body,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}

	return &result.Data.Probe, nil
}

// ProbeHealth represents probe health status
type ProbeHealth struct {
	ProbeUUID       string               `json:"probe_uuid"`
	Name            string               `json:"name"`
	Type            string               `json:"type"`
	Target          string               `json:"target"`
	Enabled         bool                 `json:"enabled"`
	LastStatus      string               `json:"last_status"`
	LastRun         string               `json:"last_run,omitempty"`
	HealthScore     float64              `json:"health_score"`
	Availability24h float64              `json:"availability_24h"`
	AverageResponse int                  `json:"average_response_ms"`
	RegionStatus    []RegionHealthStatus `json:"region_status,omitempty"`
}

// RegionHealthStatus represents health status for a specific region
type RegionHealthStatus struct {
	Region          string  `json:"region"`
	RegionName      string  `json:"region_name"`
	LastStatus      string  `json:"last_status"`
	LastRun         string  `json:"last_run,omitempty"`
	Availability24h float64 `json:"availability_24h"`
	AverageResponse int     `json:"average_response_ms"`
}

// ========================================
// CONTROLLER-SPECIFIC METHODS
// ========================================
// These methods are used by probe-controller for orchestration.
// They require monitoring key authentication and are designed for
// internal controller-to-API communication.

// ListByOrganization retrieves all probes for a specific organization
// This method is used by probe-controller to load probe assignments on startup.
// It requires monitoring key authentication and is typically called once during
// controller initialization to populate the probe scheduler.
func (s *ProbesService) ListByOrganization(ctx context.Context, organizationID uint) ([]*MonitoringProbe, error) {
	var result struct {
		Status string             `json:"status"`
		Data   []*MonitoringProbe `json:"data"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/controllers/probes/list",
		Query:  map[string]string{"org_id": fmt.Sprintf("%d", organizationID)},
		Result: &result,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list probes for organization %d: %w", organizationID, err)
	}

	return result.Data, nil
}

// GetByUUID retrieves a specific probe by its UUID
// This method is used by probe-controller to fetch probe details.
// It's a convenience wrapper around the standard Get method for consistency
// with other controller-specific methods.
func (s *ProbesService) GetByUUID(ctx context.Context, probeUUID string) (*MonitoringProbe, error) {
	probe, err := s.Get(ctx, probeUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get probe %s: %w", probeUUID, err)
	}
	return probe, nil
}

// GetRegionalResults retrieves recent regional execution results for a probe
// This method is used by consensus engine to calculate global status.
// It returns the latest execution result from each monitoring region for the
// specified probe, which the consensus engine uses to determine overall probe health.
func (s *ProbesService) GetRegionalResults(ctx context.Context, probeUUID string) ([]RegionalResult, error) {
	var result struct {
		Status string           `json:"status"`
		Data   []RegionalResult `json:"data"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/controllers/probes/%s/regional-results", probeUUID),
		Result: &result,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get regional results for probe %s: %w", probeUUID, err)
	}

	return result.Data, nil
}

// UpdateControllerStatus updates the probe status from controller
// This method is used by probe-controller to update probe execution status.
// Valid status values are: "up", "down", "degraded", "unknown"
// This should be called after the consensus engine calculates the global status.
func (s *ProbesService) UpdateControllerStatus(ctx context.Context, probeUUID string, status string) error {
	// Validate status
	validStatuses := map[string]bool{
		"up":       true,
		"down":     true,
		"degraded": true,
		"unknown":  true,
	}
	if !validStatuses[status] {
		return fmt.Errorf("invalid status '%s' for probe %s: must be one of: up, down, degraded, unknown", status, probeUUID)
	}

	body := map[string]interface{}{
		"status": status,
	}

	var result struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/controllers/probes/%s/status", probeUUID),
		Body:   body,
		Result: &result,
	})
	if err != nil {
		return fmt.Errorf("failed to update status to '%s' for probe %s: %w", status, probeUUID, err)
	}

	return nil
}

// GetProbeConfig retrieves probe configuration including consensus type
// This method is used by consensus engine to determine consensus algorithm
//
// UUID Handling: MonitoringProbe uses GormModel which only has an ID field.
// The probeUUID parameter is used to set ProbeConfiguration.ProbeUUID since
// the MonitoringProbe struct doesn't contain a UUID field.
//
// Default Consensus: If the probe's config doesn't specify a consensus_type,
// this method defaults to "majority" consensus. Valid consensus types are:
// - "majority": Probe is UP if more than 50% of regions report UP
// - "all": Probe is UP only if ALL regions report UP
// - "any": Probe is UP if ANY region reports UP
func (s *ProbesService) GetProbeConfig(ctx context.Context, probeUUID string) (*ProbeConfiguration, error) {
	probe, err := s.GetByUUID(ctx, probeUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get config for probe %s: %w", probeUUID, err)
	}

	// Convert MonitoringProbe to ProbeConfiguration
	config := &ProbeConfiguration{
		ProbeUUID:     probeUUID, // Use the provided UUID parameter (see UUID Handling above)
		Name:          probe.Name,
		Type:          probe.Type,
		Target:        probe.Target,
		Interval:      probe.Interval,
		Timeout:       probe.Timeout,
		Regions:       probe.Regions,
		ConsensusType: "majority", // Default consensus type (see Default Consensus above)
	}

	// Extract consensus type from probe config if present
	if probe.Config != nil {
		if consensusType, ok := probe.Config["consensus_type"].(string); ok {
			config.ConsensusType = consensusType
		}
	}

	return config, nil
}

// RecordConsensusResult stores a consensus calculation result
// This method is used by consensus engine to persist global status.
// It records the complete consensus calculation including regional results,
// statistics, and whether an alert should be triggered. This creates a historical
// record for reporting and debugging consensus decisions.
func (s *ProbesService) RecordConsensusResult(ctx context.Context, result *ConsensusResultRequest) error {
	if result == nil {
		return fmt.Errorf("consensus result cannot be nil")
	}
	if result.ProbeUUID == "" {
		return fmt.Errorf("probe UUID is required in consensus result")
	}

	var resp struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/controllers/probes/consensus-results",
		Body:   result,
		Result: &resp,
	})
	if err != nil {
		return fmt.Errorf("failed to record consensus result for probe %s: %w", result.ProbeUUID, err)
	}

	return nil
}

// GetActiveProbes retrieves all active (enabled) probes for an organization
// This method is used by probe-controller to get probes that should be actively monitored.
// It filters for enabled=true and deleted_at IS NULL, returning only probes that
// should be executing according to their configured schedules.
//
// This is a convenience method that filters the results of ListByOrganization to
// return only probes that are currently active and should be scheduled for execution.
func (s *ProbesService) GetActiveProbes(ctx context.Context, organizationID uint) ([]*MonitoringProbe, error) {
	// Get all probes for the organization
	allProbes, err := s.ListByOrganization(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to list probes for organization %d: %w", organizationID, err)
	}

	// Filter for active (enabled) probes only
	activeProbes := make([]*MonitoringProbe, 0, len(allProbes))
	for _, probe := range allProbes {
		if probe.Enabled {
			activeProbes = append(activeProbes, probe)
		}
	}

	return activeProbes, nil
}

// SubmitResult submits a single probe execution result
// This is a convenience wrapper around Monitoring.SubmitResults for submitting a
// single probe result. For bulk submissions, use client.Monitoring.SubmitResults directly.
//
// Example usage:
//
//	result := &nexmonyx.ProbeExecutionResult{
//	    ProbeID:      probe.ID,
//	    ProbeUUID:    probe.ProbeUUID,
//	    ExecutedAt:   time.Now(),
//	    Region:       "us-east-1",
//	    Status:       "success",
//	    ResponseTime: 150,
//	    StatusCode:   200,
//	}
//	err := client.Probes.SubmitResult(ctx, result)
func (s *ProbesService) SubmitResult(ctx context.Context, result *ProbeExecutionResult) error {
	if result == nil {
		return fmt.Errorf("probe result cannot be nil")
	}
	if result.ProbeUUID == "" {
		return fmt.Errorf("probe UUID is required in result")
	}

	// Use the Monitoring service's SubmitResults method for a single result
	return s.client.Monitoring.SubmitResults(ctx, []ProbeExecutionResult{*result})
}

// ========================================
// CONTROLLER-SPECIFIC TYPES
// ========================================

// RegionalResult represents the status from a single monitoring region
type RegionalResult struct {
	Region       string  `json:"region"`
	Status       string  `json:"status"`
	ResponseTime int     `json:"response_time"`
	ErrorMessage *string `json:"error_message,omitempty"`
	CheckedAt    string  `json:"checked_at"`
}

// ProbeConfiguration represents probe configuration for controller use
type ProbeConfiguration struct {
	ProbeUUID     string   `json:"probe_uuid"`
	Name          string   `json:"name"`
	Type          string   `json:"type"`
	Target        string   `json:"target"`
	Interval      int      `json:"interval"`
	Timeout       int      `json:"timeout"`
	Regions       []string `json:"regions"`
	ConsensusType string   `json:"consensus_type"` // "majority", "all", "any"
}

// ConsensusResultRequest represents a consensus result to be recorded
type ConsensusResultRequest struct {
	ProbeUUID           string           `json:"probe_uuid"`
	OrganizationID      uint             `json:"organization_id"`
	ConsensusType       string           `json:"consensus_type"`
	GlobalStatus        string           `json:"global_status"`
	RegionResults       []RegionalResult `json:"region_results"`
	TotalRegions        int              `json:"total_regions"`
	UpRegions           int              `json:"up_regions"`
	DownRegions         int              `json:"down_regions"`
	DegradedRegions     int              `json:"degraded_regions"`
	UnknownRegions      int              `json:"unknown_regions"`
	ShouldAlert         bool             `json:"should_alert"`
	AverageResponseTime int              `json:"average_response_time"`
	MinResponseTime     int              `json:"min_response_time"`
	MaxResponseTime     int              `json:"max_response_time"`
	UptimePercentage    float64          `json:"uptime_percentage"`
}
