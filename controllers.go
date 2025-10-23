package nexmonyx

import (
	"context"
	"fmt"
)

// SubmitControllerHeartbeat submits a heartbeat for a controller
func (s *ControllersService) SubmitControllerHeartbeat(ctx context.Context, controllerName string, req *ControllerHeartbeatRequest) error {
	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/controllers/%s/heartbeat", controllerName),
		Body:   req,
	})
	return err
}

// GetControllersSummary retrieves a summary of all controllers
func (s *ControllersService) GetControllersSummary(ctx context.Context) (*ControllersSummaryResponse, error) {
	var resp StandardResponse
	resp.Data = &ControllersSummaryResponse{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/controllers/summary",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if summary, ok := resp.Data.(*ControllersSummaryResponse); ok {
		return summary, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetControllerStatus retrieves the status of a specific controller
func (s *ControllersService) GetControllerStatus(ctx context.Context, controllerName string) (*ControllerStatusResponse, error) {
	var resp StandardResponse
	resp.Data = &ControllerStatusResponse{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/controllers/%s/status", controllerName),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if status, ok := resp.Data.(*ControllerStatusResponse); ok {
		return status, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// DeleteController deletes a controller record (admin only)
func (s *ControllersService) DeleteController(ctx context.Context, controllerName string) error {
	_, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v1/controllers/%s", controllerName),
	})
	return err
}

// GetAlertControllerStatus retrieves the alert controller status
func (s *ControllersService) GetAlertControllerStatus(ctx context.Context) (*AlertControllerStatusResponse, error) {
	var resp StandardResponse
	resp.Data = &AlertControllerStatusResponse{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/controllers/alerts/status",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if status, ok := resp.Data.(*AlertControllerStatusResponse); ok {
		return status, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetAlertControllerLeaderStatus retrieves the alert controller leader election status
func (s *ControllersService) GetAlertControllerLeaderStatus(ctx context.Context) (map[string]interface{}, error) {
	var resp StandardResponse
	var leaderStatus map[string]interface{}
	resp.Data = &leaderStatus

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/controllers/alerts/leader/status",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return leaderStatus, nil
}

// Controller Management Types

// ControllersSummaryResponse represents the summary of all controllers
type ControllersSummaryResponse struct {
	TotalControllers   int                              `json:"total_controllers"`
	ActiveControllers  int                              `json:"active_controllers"`
	HealthyControllers int                              `json:"healthy_controllers"`
	Controllers        map[string]ControllerSummaryInfo `json:"controllers"`
	LastUpdated        *CustomTime                      `json:"last_updated"`
}

// ControllerSummaryInfo represents summary information for a single controller
type ControllerSummaryInfo struct {
	Name          string      `json:"name"`
	Status        string      `json:"status"`
	Version       string      `json:"version"`
	IsLeader      bool        `json:"is_leader"`
	LastSeen      *CustomTime `json:"last_seen"`
	InstanceCount int         `json:"instance_count,omitempty"`
}

// ControllerStatusResponse represents detailed status for a specific controller
type ControllerStatusResponse struct {
	ControllerName string                 `json:"controller_name"`
	Status         string                 `json:"status"`
	Version        string                 `json:"version"`
	IsLeader       bool                   `json:"is_leader"`
	Instances      []ControllerInstance   `json:"instances"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	LastUpdated    *CustomTime            `json:"last_updated"`
}

// ControllerInstance represents a single instance of a controller
type ControllerInstance struct {
	Hostname       string                 `json:"hostname"`
	Version        string                 `json:"version"`
	Status         string                 `json:"status"`
	IsLeader       bool                   `json:"is_leader"`
	LastHeartbeat  *CustomTime            `json:"last_heartbeat"`
	ProcessedItems int64                  `json:"processed_items,omitempty"`
	ErrorCount     int64                  `json:"error_count,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// AlertControllerStatusResponse represents the alert controller status
type AlertControllerStatusResponse struct {
	Service            string      `json:"service"`
	Version            string      `json:"version"`
	Status             string      `json:"status"`
	IsLeader           bool        `json:"is_leader"`
	ActiveRules        int         `json:"active_rules"`
	ActiveInstances    int         `json:"active_instances"`
	ProcessedAlerts    int64       `json:"processed_alerts"`
	NotificationsSent  int64       `json:"notifications_sent"`
	LastProcessedAt    *CustomTime `json:"last_processed_at,omitempty"`
	QueueDepth         int         `json:"queue_depth,omitempty"`
	ProcessingRate     float64     `json:"processing_rate,omitempty"`
	ErrorCount         int64       `json:"error_count"`
	HealthChecksPassed int         `json:"health_checks_passed"`
	HealthChecksFailed int         `json:"health_checks_failed"`
}
