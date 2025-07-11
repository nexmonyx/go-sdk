package nexmonyx

import (
	"context"
	"fmt"
)

// ControllersService handles communication with controller-related endpoints
// This service provides methods for controller health monitoring, heartbeat functionality,
// and controller registration with the Nexmonyx API.

// SendHeartbeat sends a heartbeat from a controller to the API
func (s *ControllersService) SendHeartbeat(ctx context.Context, req *ControllerHeartbeatRequest) error {
	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/controllers/heartbeat",
		Body:   req,
	})
	return err
}

// RegisterController registers a new controller with the API
func (s *ControllersService) RegisterController(ctx context.Context, controllerID, version string) error {
	req := map[string]interface{}{
		"controller_id": controllerID,
		"version":       version,
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/controllers/register",
		Body:   req,
	})
	return err
}

// GetControllerStatus retrieves the status of a specific controller
func (s *ControllersService) GetControllerStatus(ctx context.Context, controllerID string) (*ControllerHealthInfo, error) {
	var resp StandardResponse
	resp.Data = &ControllerHealthInfo{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/controllers/%s/status", controllerID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if status, ok := resp.Data.(*ControllerHealthInfo); ok {
		return status, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// ListControllers retrieves a list of all registered controllers
func (s *ControllersService) ListControllers(ctx context.Context) ([]ControllerHealthInfo, error) {
	var resp StandardResponse
	var controllers []ControllerHealthInfo
	resp.Data = &controllers

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/controllers",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if data, ok := resp.Data.(*[]ControllerHealthInfo); ok {
		return *data, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// UpdateControllerStatus updates the status of a specific controller
func (s *ControllersService) UpdateControllerStatus(ctx context.Context, controllerID string, status string) error {
	req := map[string]interface{}{
		"status": status,
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/controllers/%s/status", controllerID),
		Body:   req,
	})
	return err
}

// DeregisterController removes a controller from the API registry
func (s *ControllersService) DeregisterController(ctx context.Context, controllerID string) error {
	_, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v1/controllers/%s", controllerID),
	})
	return err
}
