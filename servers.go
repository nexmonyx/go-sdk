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
		Path:   fmt.Sprintf("/v1/server/%s", id),
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
		Path:   fmt.Sprintf("/v1/server/uuid/%s", uuid),
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
		Path:   fmt.Sprintf("/v1/server/%s", id),
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
		Path:   fmt.Sprintf("/v1/server/%s", id),
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
		Path:   "/v1/server/register",
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
		Path:   fmt.Sprintf("/v1/server/%s/heartbeat", uuid),
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
		Path:   fmt.Sprintf("/v1/server/%s/metrics", id),
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
		Path:   fmt.Sprintf("/v1/server/%s/alerts", id),
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
		Path:   fmt.Sprintf("/v1/server/%s/tags", id),
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
		Path:   fmt.Sprintf("/v1/server/%s/execute", id),
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
		Path:   fmt.Sprintf("/v1/server/%s/system-info", id),
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
		Path:   "/v1/server/register",
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
	if s.client.config.Debug {
		fmt.Printf("[DEBUG] Heartbeat: Starting heartbeat request\n")
		fmt.Printf("[DEBUG] Heartbeat: Endpoint: POST /v1/heartbeat\n")
		fmt.Printf("[DEBUG] Heartbeat: Using server UUID: %s\n", s.client.config.Auth.ServerUUID)
	}

	var resp StandardResponse

	httpResp, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/heartbeat",
		Result: &resp,
	})
	
	if s.client.config.Debug {
		if err != nil {
			fmt.Printf("[DEBUG] Heartbeat: Request failed with error: %v\n", err)
		} else {
			fmt.Printf("[DEBUG] Heartbeat: Request successful\n")
			fmt.Printf("[DEBUG] Heartbeat: Response status: %s\n", resp.Status)
			fmt.Printf("[DEBUG] Heartbeat: Response message: %s\n", resp.Message)
			if httpResp != nil {
				fmt.Printf("[DEBUG] Heartbeat: HTTP Status Code: %d\n", httpResp.StatusCode)
			}
		}
	}

	return err
}

// HeartbeatWithVersion sends a heartbeat with agent version from the authenticated server
func (s *ServersService) HeartbeatWithVersion(ctx context.Context, agentVersion string) error {
	if s.client.config.Debug {
		fmt.Printf("[DEBUG] HeartbeatWithVersion: Starting heartbeat request with version\n")
		fmt.Printf("[DEBUG] HeartbeatWithVersion: Endpoint: POST /v1/heartbeat\n")
		fmt.Printf("[DEBUG] HeartbeatWithVersion: Agent version: %s\n", agentVersion)
		fmt.Printf("[DEBUG] HeartbeatWithVersion: Using server UUID: %s\n", s.client.config.Auth.ServerUUID)
	}

	var resp StandardResponse

	body := map[string]string{
		"agent_version": agentVersion,
	}

	if s.client.config.Debug {
		fmt.Printf("[DEBUG] HeartbeatWithVersion: Request body: %+v\n", body)
	}

	httpResp, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/heartbeat",
		Body:   body,
		Result: &resp,
	})
	
	if s.client.config.Debug {
		if err != nil {
			fmt.Printf("[DEBUG] HeartbeatWithVersion: Request failed with error: %v\n", err)
		} else {
			fmt.Printf("[DEBUG] HeartbeatWithVersion: Request successful\n")
			fmt.Printf("[DEBUG] HeartbeatWithVersion: Response status: %s\n", resp.Status)
			fmt.Printf("[DEBUG] HeartbeatWithVersion: Response message: %s\n", resp.Message)
			if httpResp != nil {
				fmt.Printf("[DEBUG] HeartbeatWithVersion: HTTP Status Code: %d\n", httpResp.StatusCode)
			}
		}
	}

	return err
}

// UpdateServer updates server information
func (s *ServersService) UpdateServer(ctx context.Context, serverUUID string, req *ServerUpdateRequest) (*Server, error) {
	var resp StandardResponse
	resp.Data = &Server{}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/server/%s", serverUUID),
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
	endpoint := fmt.Sprintf("/v1/server/%s/details", serverUUID)
	
	if s.client.config.Debug {
		fmt.Printf("[DEBUG] UpdateDetails: Starting server details update\n")
		fmt.Printf("[DEBUG] UpdateDetails: Endpoint: PUT %s\n", endpoint)
		fmt.Printf("[DEBUG] UpdateDetails: Server UUID: %s\n", serverUUID)
		fmt.Printf("[DEBUG] UpdateDetails: Request data:\n")
		if req != nil {
			fmt.Printf("[DEBUG]   Hostname: %s\n", req.Hostname)
			fmt.Printf("[DEBUG]   OS: %s\n", req.OS)
			fmt.Printf("[DEBUG]   OS Version: %s\n", req.OSVersion)
			fmt.Printf("[DEBUG]   OS Arch: %s\n", req.OSArch)
			fmt.Printf("[DEBUG]   CPUModel: %s\n", req.CPUModel)
			fmt.Printf("[DEBUG]   CPUCores: %d\n", req.CPUCores)
			fmt.Printf("[DEBUG]   MemoryTotal: %d\n", req.MemoryTotal)
			fmt.Printf("[DEBUG]   StorageTotal: %d\n", req.StorageTotal)
		}
		fmt.Printf("[DEBUG] UpdateDetails: Using authentication - Server UUID: %s\n", s.client.config.Auth.ServerUUID)
	}

	var resp StandardResponse
	resp.Data = &Server{}

	httpResp, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   endpoint,
		Body:   req,
		Result: &resp,
	})
	
	if s.client.config.Debug {
		if err != nil {
			fmt.Printf("[DEBUG] UpdateDetails: Request failed with error: %v\n", err)
		} else {
			fmt.Printf("[DEBUG] UpdateDetails: Request successful\n")
			fmt.Printf("[DEBUG] UpdateDetails: Response status: %s\n", resp.Status)
			fmt.Printf("[DEBUG] UpdateDetails: Response message: %s\n", resp.Message)
			if httpResp != nil {
				fmt.Printf("[DEBUG] UpdateDetails: HTTP Status Code: %d\n", httpResp.StatusCode)
			}
			if server, ok := resp.Data.(*Server); ok && server != nil {
				fmt.Printf("[DEBUG] UpdateDetails: Server ID: %d\n", server.ID)
				fmt.Printf("[DEBUG] UpdateDetails: Server UUID: %s\n", server.ServerUUID)
				fmt.Printf("[DEBUG] UpdateDetails: Server Hostname: %s\n", server.Hostname)
			}
		}
	}

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
	endpoint := fmt.Sprintf("/v1/server/%s/info", serverUUID)
	
	if s.client.config.Debug {
		fmt.Printf("[DEBUG] UpdateInfo: Starting server info update\n")
		fmt.Printf("[DEBUG] UpdateInfo: Endpoint: PUT %s\n", endpoint)
		fmt.Printf("[DEBUG] UpdateInfo: Server UUID: %s\n", serverUUID)
		fmt.Printf("[DEBUG] UpdateInfo: Request data:\n")
		if req != nil {
			fmt.Printf("[DEBUG]   Hostname: %s\n", req.Hostname)
			fmt.Printf("[DEBUG]   OS: %s\n", req.OS)
			fmt.Printf("[DEBUG]   OS Version: %s\n", req.OSVersion)
			fmt.Printf("[DEBUG]   OS Arch: %s\n", req.OSArch)
			fmt.Printf("[DEBUG]   CPUModel: %s\n", req.CPUModel)
			fmt.Printf("[DEBUG]   CPUCores: %d\n", req.CPUCores)
			fmt.Printf("[DEBUG]   MemoryTotal: %d\n", req.MemoryTotal)
			fmt.Printf("[DEBUG]   StorageTotal: %d\n", req.StorageTotal)
		}
		fmt.Printf("[DEBUG] UpdateInfo: Using authentication - Server UUID: %s\n", s.client.config.Auth.ServerUUID)
	}

	var resp StandardResponse
	resp.Data = &Server{}

	httpResp, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   endpoint,
		Body:   req,
		Result: &resp,
	})
	
	if s.client.config.Debug {
		if err != nil {
			fmt.Printf("[DEBUG] UpdateInfo: Request failed with error: %v\n", err)
		} else {
			fmt.Printf("[DEBUG] UpdateInfo: Request successful\n")
			fmt.Printf("[DEBUG] UpdateInfo: Response status: %s\n", resp.Status)
			fmt.Printf("[DEBUG] UpdateInfo: Response message: %s\n", resp.Message)
			if httpResp != nil {
				fmt.Printf("[DEBUG] UpdateInfo: HTTP Status Code: %d\n", httpResp.StatusCode)
			}
			if server, ok := resp.Data.(*Server); ok && server != nil {
				fmt.Printf("[DEBUG] UpdateInfo: Server ID: %d\n", server.ID)
				fmt.Printf("[DEBUG] UpdateInfo: Server UUID: %s\n", server.ServerUUID)
				fmt.Printf("[DEBUG] UpdateInfo: Server Hostname: %s\n", server.Hostname)
			}
		}
	}

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
	endpoint := fmt.Sprintf("/v1/server/%s/heartbeat", serverUUID)
	
	if s.client.config.Debug {
		fmt.Printf("[DEBUG] UpdateHeartbeat: Starting heartbeat update\n")
		fmt.Printf("[DEBUG] UpdateHeartbeat: Endpoint: PUT %s\n", endpoint)
		fmt.Printf("[DEBUG] UpdateHeartbeat: Server UUID: %s\n", serverUUID)
		fmt.Printf("[DEBUG] UpdateHeartbeat: Using authentication - Server UUID: %s\n", s.client.config.Auth.ServerUUID)
	}

	var resp StandardResponse

	httpResp, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   endpoint,
		Result: &resp,
	})
	
	if s.client.config.Debug {
		if err != nil {
			fmt.Printf("[DEBUG] UpdateHeartbeat: Request failed with error: %v\n", err)
		} else {
			fmt.Printf("[DEBUG] UpdateHeartbeat: Request successful\n")
			fmt.Printf("[DEBUG] UpdateHeartbeat: Response status: %s\n", resp.Status)
			fmt.Printf("[DEBUG] UpdateHeartbeat: Response message: %s\n", resp.Message)
			if httpResp != nil {
				fmt.Printf("[DEBUG] UpdateHeartbeat: HTTP Status Code: %d\n", httpResp.StatusCode)
			}
		}
	}

	return err
}

// GetHeartbeat retrieves heartbeat information for a server
func (s *ServersService) GetHeartbeat(ctx context.Context, serverUUID string) (*HeartbeatResponse, error) {
	endpoint := fmt.Sprintf("/v1/server/%s/heartbeat", serverUUID)
	
	if s.client.config.Debug {
		fmt.Printf("[DEBUG] GetHeartbeat: Starting heartbeat retrieval\n")
		fmt.Printf("[DEBUG] GetHeartbeat: Endpoint: GET %s\n", endpoint)
		fmt.Printf("[DEBUG] GetHeartbeat: Server UUID: %s\n", serverUUID)
		fmt.Printf("[DEBUG] GetHeartbeat: Using authentication - Server UUID: %s\n", s.client.config.Auth.ServerUUID)
	}

	var resp StandardResponse
	resp.Data = &HeartbeatResponse{}

	httpResp, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   endpoint,
		Result: &resp,
	})
	
	if s.client.config.Debug {
		if err != nil {
			fmt.Printf("[DEBUG] GetHeartbeat: Request failed with error: %v\n", err)
		} else {
			fmt.Printf("[DEBUG] GetHeartbeat: Request successful\n")
			fmt.Printf("[DEBUG] GetHeartbeat: Response status: %s\n", resp.Status)
			fmt.Printf("[DEBUG] GetHeartbeat: Response message: %s\n", resp.Message)
			if httpResp != nil {
				fmt.Printf("[DEBUG] GetHeartbeat: HTTP Status Code: %d\n", httpResp.StatusCode)
			}
			if heartbeat, ok := resp.Data.(*HeartbeatResponse); ok && heartbeat != nil {
				fmt.Printf("[DEBUG] GetHeartbeat: Server UUID: %s\n", heartbeat.ServerUUID)
				fmt.Printf("[DEBUG] GetHeartbeat: Server Status: %s\n", heartbeat.ServerStatus)
				if heartbeat.LastHeartbeat != nil {
					fmt.Printf("[DEBUG] GetHeartbeat: Last Heartbeat: %v\n", *heartbeat.LastHeartbeat)
				}
				fmt.Printf("[DEBUG] GetHeartbeat: Heartbeat Count: %d\n", heartbeat.HeartbeatCount)
			}
		}
	}

	if err != nil {
		return nil, err
	}

	if heartbeat, ok := resp.Data.(*HeartbeatResponse); ok {
		return heartbeat, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}
