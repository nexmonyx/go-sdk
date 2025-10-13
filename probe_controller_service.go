package nexmonyx

import (
	"context"
	"fmt"
	"time"
)

// ProbeControllerService handles probe controller operations for orchestrating
// probe executions across regional monitoring nodes. This service manages
// assignments, regional results, consensus calculations, and health state tracking.
//
// This service is designed for internal controller-to-API communication and requires
// appropriate authentication (monitoring key or API key/secret).
type ProbeControllerService struct {
	client *Client
}

// ProbeControllerAssignment represents a probe execution assignment that links
// a probe to a specific monitoring node in a region for execution.
type ProbeControllerAssignment struct {
	ID               uint       `json:"id"`
	ProbeID          uint       `json:"probe_id"`
	ProbeUUID        string     `json:"probe_uuid"`
	MonitoringNodeID *uint      `json:"monitoring_node_id"`
	Region           string     `json:"region"`
	AssignedAt       time.Time  `json:"assigned_at"`
	LastExecution    *time.Time `json:"last_execution"`
	Status           string     `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// ProbeControllerAssignmentCreateRequest contains the fields required to create
// a new probe execution assignment.
type ProbeControllerAssignmentCreateRequest struct {
	ProbeID          uint   `json:"probe_id"`
	ProbeUUID        string `json:"probe_uuid"`
	MonitoringNodeID *uint  `json:"monitoring_node_id,omitempty"`
	Region           string `json:"region"`
	Status           string `json:"status,omitempty"`
}

// ProbeControllerAssignmentUpdateRequest contains the fields that can be updated
// on an existing probe execution assignment.
type ProbeControllerAssignmentUpdateRequest struct {
	MonitoringNodeID *uint  `json:"monitoring_node_id,omitempty"`
	Status           string `json:"status,omitempty"`
}

// ProbeControllerAssignmentListOptions provides filtering options when listing
// probe execution assignments. All fields are optional - use nil to skip filtering.
type ProbeControllerAssignmentListOptions struct {
	ProbeUUID        *string
	Region           *string
	Status           *string
	MonitoringNodeID *uint
}

// ProbeControllerRegionalResult represents the execution result from a single
// monitoring region for a probe. These results are aggregated by the consensus
// engine to determine global probe status.
type ProbeControllerRegionalResult struct {
	ID               uint       `json:"id"`
	ProbeUUID        string     `json:"probe_uuid"`
	Region           string     `json:"region"`
	Status           string     `json:"status"`
	ResponseTime     *int       `json:"response_time"`
	Success          bool       `json:"success"`
	ErrorMessage     *string    `json:"error_message"`
	IsCustomerRegion bool       `json:"is_customer_region"`
	AgentID          *string    `json:"agent_id"`
	Timestamp        time.Time  `json:"timestamp"`
	ExpiresAt        *time.Time `json:"expires_at"`
	CreatedAt        time.Time  `json:"created_at"`
}

// ProbeControllerRegionalResultStoreRequest contains the fields required to store
// a regional probe execution result. The TTLSeconds field controls how long the
// result is kept before expiration.
type ProbeControllerRegionalResultStoreRequest struct {
	ProbeUUID        string  `json:"probe_uuid"`
	Region           string  `json:"region"`
	Status           string  `json:"status"`
	ResponseTime     *int    `json:"response_time,omitempty"`
	Success          bool    `json:"success"`
	ErrorMessage     *string `json:"error_message,omitempty"`
	IsCustomerRegion bool    `json:"is_customer_region"`
	AgentID          *string `json:"agent_id,omitempty"`
	TTLSeconds       int     `json:"ttl_seconds,omitempty"`
}

// ProbeControllerRegionalResultListOptions provides filtering options when listing
// regional probe execution results. All fields are optional - use nil to skip filtering.
type ProbeControllerRegionalResultListOptions struct {
	Region           *string
	Status           *string
	IsCustomerRegion *bool
	Since            *string
}

// ProbeControllerConsensusResult represents the aggregated consensus calculation
// from all regional probe execution results. This determines the global status
// and whether alerts should be triggered.
type ProbeControllerConsensusResult struct {
	ID              uint      `json:"id"`
	ProbeID         uint      `json:"probe_id"`
	ProbeUUID       string    `json:"probe_uuid"`
	GlobalStatus    string    `json:"global_status"`
	ConsensusType   string    `json:"consensus_type"`
	ShouldAlert     bool      `json:"should_alert"`
	UpRegions       int       `json:"up_regions"`
	DownRegions     int       `json:"down_regions"`
	DegradedRegions int       `json:"degraded_regions"`
	UnknownRegions  int       `json:"unknown_regions"`
	TotalRegions    int       `json:"total_regions"`
	ConsensusRatio  float64   `json:"consensus_ratio"`
	CalculatedAt    time.Time `json:"calculated_at"`
	AlertTriggered  bool      `json:"alert_triggered"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ProbeControllerConsensusResultStoreRequest contains the fields required to store
// a consensus calculation result. This is typically called after the consensus engine
// aggregates all regional results for a probe.
type ProbeControllerConsensusResultStoreRequest struct {
	ProbeID         uint    `json:"probe_id"`
	ProbeUUID       string  `json:"probe_uuid"`
	GlobalStatus    string  `json:"global_status"`
	ConsensusType   string  `json:"consensus_type"`
	ShouldAlert     bool    `json:"should_alert"`
	UpRegions       int     `json:"up_regions"`
	DownRegions     int     `json:"down_regions"`
	DegradedRegions int     `json:"degraded_regions"`
	UnknownRegions  int     `json:"unknown_regions"`
	TotalRegions    int     `json:"total_regions"`
	ConsensusRatio  float64 `json:"consensus_ratio"`
	AlertTriggered  bool    `json:"alert_triggered"`
}

// ConsensusHistoryOptions provides filtering options when retrieving historical
// consensus results for a probe. All fields are optional - use nil to skip filtering.
type ConsensusHistoryOptions struct {
	Since *string
	Limit *int
}

// ProbeControllerHealthState represents a key-value health state entry for
// the probe controller. This can be used to track controller status, metrics,
// or operational information.
type ProbeControllerHealthState struct {
	ID        uint      `json:"id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ProbeControllerHealthUpdateRequest contains the fields required to update
// or create a health state entry.
type ProbeControllerHealthUpdateRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// CreateAssignment creates a new probe execution assignment, linking a probe to a
// specific monitoring node in a region. This is used by the controller to distribute
// probe execution work across the monitoring infrastructure.
//
// Example:
//
//	assignment, err := client.ProbeController.CreateAssignment(ctx, &nexmonyx.ProbeControllerAssignmentCreateRequest{
//	    ProbeID:   123,
//	    ProbeUUID: "probe-uuid-here",
//	    Region:    "us-east-1",
//	    Status:    "active",
//	})
func (s *ProbeControllerService) CreateAssignment(ctx context.Context, req *ProbeControllerAssignmentCreateRequest) (*ProbeControllerAssignment, error) {
	// Validate required fields
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	if req.ProbeID == 0 {
		return nil, fmt.Errorf("probe_id is required")
	}
	if req.ProbeUUID == "" {
		return nil, fmt.Errorf("probe_uuid is required")
	}
	if req.Region == "" {
		return nil, fmt.Errorf("region is required")
	}

	var result struct {
		Status  string                     `json:"status"`
		Data    *ProbeControllerAssignment `json:"data"`
		Message string                     `json:"message"`
	}
	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/controllers/probe/assignments",
		Body:   req,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

// ListAssignments retrieves probe execution assignments with optional filtering.
// Use the options parameter to filter by probe UUID, region, status, or monitoring node.
//
// Example:
//
//	assignments, err := client.ProbeController.ListAssignments(ctx, &nexmonyx.ProbeControllerAssignmentListOptions{
//	    ProbeUUID: "probe-uuid-here",
//	    Region:    "us-east-1",
//	    Status:    "active",
//	})
func (s *ProbeControllerService) ListAssignments(ctx context.Context, opts *ProbeControllerAssignmentListOptions) ([]*ProbeControllerAssignment, error) {
	var result struct {
		Status  string                       `json:"status"`
		Data    []*ProbeControllerAssignment `json:"data"`
		Message string                       `json:"message"`
	}
	query := make(map[string]string)
	if opts != nil {
		if opts.ProbeUUID != nil && *opts.ProbeUUID != "" {
			query["probe_uuid"] = *opts.ProbeUUID
		}
		if opts.Region != nil && *opts.Region != "" {
			query["region"] = *opts.Region
		}
		if opts.Status != nil && *opts.Status != "" {
			query["status"] = *opts.Status
		}
		if opts.MonitoringNodeID != nil {
			query["monitoring_node_id"] = fmt.Sprintf("%d", *opts.MonitoringNodeID)
		}
	}
	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/controllers/probe/assignments",
		Query:  query,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

// UpdateAssignment updates an existing probe execution assignment. This can be used
// to change the assigned monitoring node or update the assignment status.
//
// Example:
//
//	newNodeID := uint(456)
//	assignment, err := client.ProbeController.UpdateAssignment(ctx, assignmentID, &nexmonyx.ProbeControllerAssignmentUpdateRequest{
//	    MonitoringNodeID: &newNodeID,
//	    Status:           "paused",
//	})
func (s *ProbeControllerService) UpdateAssignment(ctx context.Context, id uint, req *ProbeControllerAssignmentUpdateRequest) (*ProbeControllerAssignment, error) {
	// Validate required fields
	if id == 0 {
		return nil, fmt.Errorf("assignment id is required")
	}
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	var result struct {
		Status  string                     `json:"status"`
		Data    *ProbeControllerAssignment `json:"data"`
		Message string                     `json:"message"`
	}
	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/controllers/probe/assignments/%d", id),
		Body:   req,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

// DeleteAssignment removes a probe execution assignment. This unassigns the probe
// from its monitoring node in the specified region. The deleted assignment is returned.
//
// Example:
//
//	deletedAssignment, err := client.ProbeController.DeleteAssignment(ctx, assignmentID)
func (s *ProbeControllerService) DeleteAssignment(ctx context.Context, id uint) (*ProbeControllerAssignment, error) {
	// Validate required fields
	if id == 0 {
		return nil, fmt.Errorf("assignment id is required")
	}

	var result struct {
		Status  string                     `json:"status"`
		Data    *ProbeControllerAssignment `json:"data"`
		Message string                     `json:"message"`
	}
	_, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v1/controllers/probe/assignments/%d", id),
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

// StoreRegionalResult stores a probe execution result from a specific region.
// These regional results are used by the consensus engine to calculate global probe status.
// The TTLSeconds field controls how long the result is retained.
//
// Example:
//
//	responseTime := 150
//	result, err := client.ProbeController.StoreRegionalResult(ctx, &nexmonyx.ProbeControllerRegionalResultStoreRequest{
//	    ProbeUUID:        "probe-uuid-here",
//	    Region:           "us-east-1",
//	    Status:           "up",
//	    ResponseTime:     &responseTime,
//	    Success:          true,
//	    IsCustomerRegion: false,
//	    TTLSeconds:       3600, // 1 hour
//	})
func (s *ProbeControllerService) StoreRegionalResult(ctx context.Context, req *ProbeControllerRegionalResultStoreRequest) (*ProbeControllerRegionalResult, error) {
	// Validate required fields
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	if req.ProbeUUID == "" {
		return nil, fmt.Errorf("probe_uuid is required")
	}
	if req.Region == "" {
		return nil, fmt.Errorf("region is required")
	}
	if req.Status == "" {
		return nil, fmt.Errorf("status is required")
	}

	var result struct {
		Status  string                         `json:"status"`
		Data    *ProbeControllerRegionalResult `json:"data"`
		Message string                         `json:"message"`
	}
	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/controllers/probe/results/regional",
		Body:   req,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

// GetRegionalResults retrieves regional probe execution results for a specific probe.
// This is used by the consensus engine to aggregate results from all regions and
// calculate global probe status. Use options to filter by region, status, or customer regions.
//
// Example:
//
//	results, err := client.ProbeController.GetRegionalResults(ctx, "probe-uuid-here", &nexmonyx.ProbeControllerRegionalResultListOptions{
//	    Region: "us-east-1",
//	    Since:  "2024-01-01T00:00:00Z",
//	})
func (s *ProbeControllerService) GetRegionalResults(ctx context.Context, probeUUID string, opts *ProbeControllerRegionalResultListOptions) ([]*ProbeControllerRegionalResult, error) {
	// Validate required fields
	if probeUUID == "" {
		return nil, fmt.Errorf("probe_uuid is required")
	}

	var result struct {
		Status  string                           `json:"status"`
		Data    []*ProbeControllerRegionalResult `json:"data"`
		Message string                           `json:"message"`
	}
	query := make(map[string]string)
	if opts != nil {
		if opts.Region != nil && *opts.Region != "" {
			query["region"] = *opts.Region
		}
		if opts.Status != nil && *opts.Status != "" {
			query["status"] = *opts.Status
		}
		if opts.IsCustomerRegion != nil {
			query["is_customer_region"] = fmt.Sprintf("%t", *opts.IsCustomerRegion)
		}
		if opts.Since != nil && *opts.Since != "" {
			query["since"] = *opts.Since
		}
	}
	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/controllers/probe/results/regional/%s", probeUUID),
		Query:  query,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

// StoreConsensusResult stores a consensus calculation result after aggregating all
// regional execution results for a probe. This determines the global probe status
// and whether alerts should be triggered.
//
// Example:
//
//	consensus, err := client.ProbeController.StoreConsensusResult(ctx, &nexmonyx.ProbeControllerConsensusResultStoreRequest{
//	    ProbeID:         123,
//	    ProbeUUID:       "probe-uuid-here",
//	    GlobalStatus:    "up",
//	    ConsensusType:   "majority",
//	    ShouldAlert:     false,
//	    UpRegions:       3,
//	    DownRegions:     0,
//	    TotalRegions:    3,
//	    ConsensusRatio:  1.0,
//	    AlertTriggered:  false,
//	})
func (s *ProbeControllerService) StoreConsensusResult(ctx context.Context, req *ProbeControllerConsensusResultStoreRequest) (*ProbeControllerConsensusResult, error) {
	// Validate required fields
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	if req.ProbeID == 0 {
		return nil, fmt.Errorf("probe_id is required")
	}
	if req.ProbeUUID == "" {
		return nil, fmt.Errorf("probe_uuid is required")
	}
	if req.GlobalStatus == "" {
		return nil, fmt.Errorf("global_status is required")
	}
	if req.ConsensusType == "" {
		return nil, fmt.Errorf("consensus_type is required")
	}

	var result struct {
		Status  string                          `json:"status"`
		Data    *ProbeControllerConsensusResult `json:"data"`
		Message string                          `json:"message"`
	}
	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/controllers/probe/results/consensus",
		Body:   req,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

// GetConsensusHistory retrieves historical consensus calculation results for a probe.
// This can be used to analyze probe health trends over time and understand alerting patterns.
//
// Example:
//
//	history, err := client.ProbeController.GetConsensusHistory(ctx, "probe-uuid-here", &nexmonyx.ConsensusHistoryOptions{
//	    Since: "2024-01-01T00:00:00Z",
//	    Limit: 100,
//	})
func (s *ProbeControllerService) GetConsensusHistory(ctx context.Context, probeUUID string, opts *ConsensusHistoryOptions) ([]*ProbeControllerConsensusResult, error) {
	// Validate required fields
	if probeUUID == "" {
		return nil, fmt.Errorf("probe_uuid is required")
	}

	var result struct {
		Status  string                            `json:"status"`
		Data    []*ProbeControllerConsensusResult `json:"data"`
		Message string                            `json:"message"`
	}
	query := make(map[string]string)
	if opts != nil {
		if opts.Since != nil && *opts.Since != "" {
			query["since"] = *opts.Since
		}
		if opts.Limit != nil && *opts.Limit > 0 {
			query["limit"] = fmt.Sprintf("%d", *opts.Limit)
		}
	}
	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/controllers/probe/results/consensus/%s", probeUUID),
		Query:  query,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

// UpdateHealthState updates or creates a health state entry for the probe controller.
// This can be used to track controller status, metrics, or operational information.
// Each entry is identified by a unique key.
//
// Example:
//
//	state, err := client.ProbeController.UpdateHealthState(ctx, &nexmonyx.ProbeControllerHealthUpdateRequest{
//	    Key:   "controller_status",
//	    Value: "healthy",
//	})
func (s *ProbeControllerService) UpdateHealthState(ctx context.Context, req *ProbeControllerHealthUpdateRequest) (*ProbeControllerHealthState, error) {
	// Validate required fields
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	if req.Key == "" {
		return nil, fmt.Errorf("key is required")
	}

	var result struct {
		Status  string                      `json:"status"`
		Data    *ProbeControllerHealthState `json:"data"`
		Message string                      `json:"message"`
	}
	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   "/v1/controllers/probe/health",
		Body:   req,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

// GetHealthStates retrieves all health state entries for the probe controller.
// This returns all key-value pairs tracking controller status and operational metrics.
//
// Example:
//
//	states, err := client.ProbeController.GetHealthStates(ctx)
//	for _, state := range states {
//	    fmt.Printf("Key: %s, Value: %s, Updated: %s\n", state.Key, state.Value, state.UpdatedAt)
//	}
func (s *ProbeControllerService) GetHealthStates(ctx context.Context) ([]*ProbeControllerHealthState, error) {
	var result struct {
		Status  string                        `json:"status"`
		Data    []*ProbeControllerHealthState `json:"data"`
		Message string                        `json:"message"`
	}
	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/controllers/probe/health",
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}
