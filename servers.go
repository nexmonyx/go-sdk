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
		Path:   fmt.Sprintf("/api/v1/servers/%s", id),
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
		Path:   fmt.Sprintf("/api/v1/servers/uuid/%s", uuid),
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
		Path:   "/api/v1/servers",
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
		Path:   "/api/v1/servers",
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
		Path:   fmt.Sprintf("/api/v1/servers/%s", id),
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
		Path:   fmt.Sprintf("/api/v1/servers/%s", id),
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
		Path:   "/api/v1/servers/register",
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
		Path:   fmt.Sprintf("/api/v1/servers/%s/heartbeat", uuid),
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
		Path:   fmt.Sprintf("/api/v1/servers/%s/metrics", id),
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
		Path:   fmt.Sprintf("/api/v1/servers/%s/alerts", id),
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
		Path:   fmt.Sprintf("/api/v1/servers/%s/tags", id),
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
		Path:   fmt.Sprintf("/api/v1/servers/%s/execute", id),
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
		Path:   fmt.Sprintf("/api/v1/servers/%s/system-info", id),
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