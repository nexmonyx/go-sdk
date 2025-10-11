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
		"name":        req.Name,
		"type":        req.Type,
		"frequency":   req.Interval, // API expects "frequency", SDK has "interval"
		"regions":     []string{req.RegionCode}, // Convert single region to array
		"enabled":     req.Enabled,
		"config":      config,
	}
	
	// Add description if provided
	if req.Name != "" {
		body["description"] = req.Name // Use name as description if not provided
	}
	
	var result struct {
		Status  string           `json:"status"`
		Data    struct {
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
	return s.client.Monitoring.ListProbes(ctx, opts)
}

// Get retrieves a probe by UUID
func (s *ProbesService) Get(ctx context.Context, uuid string) (*MonitoringProbe, error) {
	return s.client.Monitoring.GetProbe(ctx, uuid)
}

// Update updates a probe
func (s *ProbesService) Update(ctx context.Context, uuid string, req *ProbeUpdateRequest) (*MonitoringProbe, error) {
	probe := &MonitoringProbe{}
	
	if req.Name != nil {
		probe.Name = *req.Name
	}
	if req.Enabled != nil {
		probe.Enabled = *req.Enabled
	}
	if req.Interval != nil {
		probe.Interval = *req.Interval
	}
	if req.Configuration != nil {
		probe.Config = req.Configuration
	}
	
	return s.client.Monitoring.UpdateProbe(ctx, uuid, probe)
}

// Delete removes a probe
func (s *ProbesService) Delete(ctx context.Context, uuid string) error {
	return s.client.Monitoring.DeleteProbe(ctx, uuid)
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
		Status string               `json:"status"`
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
		Status  string           `json:"status"`
		Data    struct {
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
// These methods are used by probe-controller for orchestration

// ListByOrganization retrieves all probes for a specific organization
// This method is used by probe-controller to load probe assignments on startup
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
		return nil, err
	}

	return result.Data, nil
}

// GetByUUID retrieves a specific probe by its UUID
// This method is used by probe-controller to fetch probe details
func (s *ProbesService) GetByUUID(ctx context.Context, probeUUID string) (*MonitoringProbe, error) {
	// Reuse the existing Get method which calls the monitoring service
	return s.Get(ctx, probeUUID)
}

// GetRegionalResults retrieves recent regional execution results for a probe
// This method is used by consensus engine to calculate global status
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
		return nil, err
	}

	return result.Data, nil
}

// UpdateControllerStatus updates the probe status from controller
// This method is used by probe-controller to update probe execution status
func (s *ProbesService) UpdateControllerStatus(ctx context.Context, probeUUID string, status string) error {
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

	return err
}

// GetProbeConfig retrieves probe configuration including consensus type
// This method is used by consensus engine to determine consensus algorithm
func (s *ProbesService) GetProbeConfig(ctx context.Context, probeUUID string) (*ProbeConfiguration, error) {
	probe, err := s.GetByUUID(ctx, probeUUID)
	if err != nil {
		return nil, err
	}

	// Convert MonitoringProbe to ProbeConfiguration
	config := &ProbeConfiguration{
		ProbeUUID:     probeUUID, // Use the provided UUID parameter
		Name:          probe.Name,
		Type:          probe.Type,
		Target:        probe.Target,
		Interval:      probe.Interval,
		Timeout:       probe.Timeout,
		Regions:       probe.Regions,
		ConsensusType: "majority", // Default consensus type
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
// This method is used by consensus engine to persist global status
func (s *ProbesService) RecordConsensusResult(ctx context.Context, result *ConsensusResultRequest) error {
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

	return err
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