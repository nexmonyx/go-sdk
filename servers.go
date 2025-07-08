package nexmonyx

import (
	"context"
	"fmt"
)

// GetServer retrieves a server by ID
func (s *ServersService) Get(ctx context.Context, id string) (*Server, error) {
	var resp StandardResponse
	resp.Data = &Server{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/servers/%s", id),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if server, ok := resp.Data.(*Server); ok {
		return server, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetServerByUUID retrieves a server by UUID
func (s *ServersService) GetByUUID(ctx context.Context, uuid string) (*Server, error) {
	var resp StandardResponse
	resp.Data = &Server{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/servers/uuid/%s", uuid),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if server, ok := resp.Data.(*Server); ok {
		return server, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// ListServers retrieves a list of servers
func (s *ServersService) List(ctx context.Context, opts *ListOptions) ([]*Server, *PaginationMeta, error) {
	var resp PaginatedResponse
	var servers []*Server
	resp.Data = &servers

	req := &Request{
		Method: "GET",
		Path:   "/v1/servers",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return servers, resp.Meta, nil
}

// CreateServer registers a new server
func (s *ServersService) Create(ctx context.Context, server *Server) (*Server, error) {
	var resp StandardResponse
	resp.Data = &Server{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/servers",
		Body:   server,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if created, ok := resp.Data.(*Server); ok {
		return created, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// UpdateServer updates an existing server
func (s *ServersService) Update(ctx context.Context, id string, server *Server) (*Server, error) {
	var resp StandardResponse
	resp.Data = &Server{}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/servers/%s", id),
		Body:   server,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if updated, ok := resp.Data.(*Server); ok {
		return updated, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// DeleteServer deletes a server
func (s *ServersService) Delete(ctx context.Context, id string) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v1/servers/%s", id),
		Result: &resp,
	})
	return err
}

// RegisterServer registers a new server with credentials
func (s *ServersService) Register(ctx context.Context, hostname string, organizationID uint) (*Server, error) {
	var resp StandardResponse
	resp.Data = &Server{}

	body := map[string]interface{}{
		"hostname":        hostname,
		"organization_id": organizationID,
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/servers/register",
		Body:   body,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if server, ok := resp.Data.(*Server); ok {
		return server, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// SendHeartbeat sends a heartbeat for a server
func (s *ServersService) SendHeartbeat(ctx context.Context, uuid string) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/servers/%s/heartbeat", uuid),
		Result: &resp,
	})
	return err
}

// GetServerMetrics retrieves metrics for a server
func (s *ServersService) GetMetrics(ctx context.Context, id string, opts *ListOptions) ([]*Metric, *PaginationMeta, error) {
	var resp PaginatedResponse
	var metrics []*Metric
	resp.Data = &metrics

	req := &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/servers/%s/metrics", id),
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return metrics, resp.Meta, nil
}

// GetServerAlerts retrieves alerts for a server
func (s *ServersService) GetAlerts(ctx context.Context, id string, opts *ListOptions) ([]*Alert, *PaginationMeta, error) {
	var resp PaginatedResponse
	var alerts []*Alert
	resp.Data = &alerts

	req := &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/servers/%s/alerts", id),
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return alerts, resp.Meta, nil
}

// UpdateServerTags updates tags for a server
func (s *ServersService) UpdateTags(ctx context.Context, id string, tags []string) (*Server, error) {
	var resp StandardResponse
	resp.Data = &Server{}

	body := map[string]interface{}{
		"tags": tags,
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/servers/%s/tags", id),
		Body:   body,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if server, ok := resp.Data.(*Server); ok {
		return server, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// ExecuteCommand executes a command on a server
func (s *ServersService) ExecuteCommand(ctx context.Context, id string, command string) (map[string]interface{}, error) {
	var resp StandardResponse
	var result map[string]interface{}
	resp.Data = &result

	body := map[string]interface{}{
		"command": command,
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/servers/%s/execute", id),
		Body:   body,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetServerSystemInfo retrieves system information for a server
func (s *ServersService) GetSystemInfo(ctx context.Context, id string) (*SystemInfo, error) {
	var resp StandardResponse
	resp.Data = &SystemInfo{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/servers/%s/system-info", id),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if info, ok := resp.Data.(*SystemInfo); ok {
		return info, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// RegisterWithKey registers a new server with a registration key
func (s *ServersService) RegisterWithKey(ctx context.Context, registrationKey string, req *ServerCreateRequest) (*Server, error) {
	resp, err := s.RegisterWithKeyFull(ctx, registrationKey, req)
	if err != nil {
		return nil, err
	}
	return resp.Server, nil
}

// RegisterWithKeyFull registers a new server with a registration key and returns the full response
func (s *ServersService) RegisterWithKeyFull(ctx context.Context, registrationKey string, req *ServerCreateRequest) (*ServerRegistrationResponse, error) {
	var resp StandardResponse
	resp.Data = &ServerRegistrationResponse{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/servers/register",
		Headers: map[string]string{
			"X-Registration-Key": registrationKey,
		},
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if regResp, ok := resp.Data.(*ServerRegistrationResponse); ok {
		return regResp, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// Heartbeat sends a heartbeat from the authenticated server
func (s *ServersService) Heartbeat(ctx context.Context) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/heartbeat",
		Result: &resp,
	})
	return err
}

// HeartbeatWithVersion sends a heartbeat with agent version from the authenticated server
func (s *ServersService) HeartbeatWithVersion(ctx context.Context, agentVersion string) error {
	var resp StandardResponse

	body := map[string]string{
		"agent_version": agentVersion,
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/heartbeat",
		Body:   body,
		Result: &resp,
	})
	return err
}

// UpdateServer updates server information
func (s *ServersService) UpdateServer(ctx context.Context, serverUUID string, req *ServerUpdateRequest) (*Server, error) {
	var resp StandardResponse
	resp.Data = &Server{}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/servers/%s", serverUUID),
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if server, ok := resp.Data.(*Server); ok {
		return server, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// UpdateDetails updates detailed server information including hardware info
func (s *ServersService) UpdateDetails(ctx context.Context, serverUUID string, req *ServerDetailsUpdateRequest) (*Server, error) {
	var resp StandardResponse
	resp.Data = &Server{}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/server/%s/details", serverUUID),
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if server, ok := resp.Data.(*Server); ok {
		return server, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// UpdateInfo updates server information (alias for UpdateDetails)
func (s *ServersService) UpdateInfo(ctx context.Context, serverUUID string, req *ServerDetailsUpdateRequest) (*Server, error) {
	var resp StandardResponse
	resp.Data = &Server{}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/server/%s/info", serverUUID),
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if server, ok := resp.Data.(*Server); ok {
		return server, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetDetails retrieves server details
func (s *ServersService) GetDetails(ctx context.Context, serverUUID string) (*Server, error) {
	var resp StandardResponse
	resp.Data = &Server{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/server/%s/details", serverUUID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if server, ok := resp.Data.(*Server); ok {
		return server, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetFullDetails retrieves comprehensive server details including CPU information
// This endpoint requires JWT authentication with servers:read permission
func (s *ServersService) GetFullDetails(ctx context.Context, serverUUID string) (*Server, error) {
	var resp StandardResponse
	resp.Data = &Server{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/server/%s/full-details", serverUUID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if server, ok := resp.Data.(*Server); ok {
		return server, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// UpdateHeartbeat updates the heartbeat for a specific server
func (s *ServersService) UpdateHeartbeat(ctx context.Context, serverUUID string) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/server/%s/heartbeat", serverUUID),
		Result: &resp,
	})
	return err
}

// GetHeartbeat retrieves heartbeat information for a server
func (s *ServersService) GetHeartbeat(ctx context.Context, serverUUID string) (*HeartbeatResponse, error) {
	var resp StandardResponse
	resp.Data = &HeartbeatResponse{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/server/%s/heartbeat", serverUUID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if heartbeat, ok := resp.Data.(*HeartbeatResponse); ok {
		return heartbeat, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}
