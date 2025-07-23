package nexmonyx

import (
	"context"
	"fmt"
)

// ProbesService handles probe-related API operations
type ProbesService struct {
	client *Client
}

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