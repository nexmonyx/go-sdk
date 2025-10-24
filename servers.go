package nexmonyx

import (
	"context"
	"fmt"
)

// GetServer retrieves a server by ID (deprecated - use GetByUUID instead)
// This method assumes the ID is actually a UUID
func (s *ServersService) Get(ctx context.Context, id string) (*Server, error) {
	// Redirect to GetByUUID since the API only supports UUID-based access
	return s.GetByUUID(ctx, id)
}

// GetServerByUUID retrieves a server by UUID
func (s *ServersService) GetByUUID(ctx context.Context, uuid string) (*Server, error) {
	var resp StandardResponse
	resp.Data = &Server{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/server/%s/details", uuid),
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
		Path:   "/v2/servers",
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

// ListInScope retrieves servers matching alert rule scope filters
func (s *ServersService) ListInScope(ctx context.Context, filters *ScopeFilters) ([]*Server, error) {
	var resp StandardResponse
	var servers []*Server
	resp.Data = &servers

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/servers/in-scope",
		Body:   filters,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return servers, nil
}

// CreateServer registers a new server (deprecated - use RegisterWithKey instead)
// This method is deprecated as server creation now requires registration keys
func (s *ServersService) Create(ctx context.Context, server *Server) (*Server, error) {
	// For backward compatibility, try to register with empty key
	// This will likely fail unless the API supports it
	var resp StandardResponse
	resp.Data = &Server{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/register",
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

// UpdateServer updates an existing server (deprecated - use UpdateDetails instead)
// This method assumes the ID is actually a UUID and uses the details endpoint
func (s *ServersService) Update(ctx context.Context, id string, server *Server) (*Server, error) {
	// Convert to server update request and use UpdateDetails
	// Map all relevant fields from Server to ServerDetailsUpdateRequest
	req := &ServerDetailsUpdateRequest{
		Hostname:       server.Hostname,
		MainIP:         server.MainIP,
		Environment:    server.Environment,
		Location:       server.Location,
		Classification: server.Classification,
		OS:             server.OS,
		OSVersion:      server.OSVersion,
		OSArch:         server.OSArch,
		CPUModel:       server.CPUModel,
		CPUCores:       server.CPUCores,
		// Note: Server uses TotalMemoryGB/TotalDiskGB while UpdateRequest uses MemoryTotal/StorageTotal
		// These are intentionally not mapped as they have different units (GB vs bytes)
	}
	return s.UpdateDetails(ctx, id, req)
}

// DeleteServer deletes a server (requires admin permissions)
// This method assumes the ID is actually a UUID and uses the admin endpoint
func (s *ServersService) Delete(ctx context.Context, id string) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v1/admin/server/%s", id),
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
		Path:   "/v1/register",
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
// Deprecated: UUID parameter is ignored. Use Heartbeat() method instead which uses server credentials from client config.
func (s *ServersService) SendHeartbeat(ctx context.Context, uuid string) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/heartbeat",
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
		Path:   "/v1/register",
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

// =============================================================================
// Unified Registration Key Methods
// =============================================================================

// RegisterWithUnifiedKey registers a new server using a unified registration key
func (s *ServersService) RegisterWithUnifiedKey(ctx context.Context, key *UnifiedAPIKey, req *ServerCreateRequest) (*Server, error) {
	if !key.CanRegisterServers() {
		return nil, fmt.Errorf("API key does not have server registration capabilities")
	}

	resp, err := s.RegisterWithUnifiedKeyFull(ctx, key, req)
	if err != nil {
		return nil, err
	}
	return resp.Server, nil
}

// RegisterWithUnifiedKeyFull registers a new server using a unified registration key and returns the full response
func (s *ServersService) RegisterWithUnifiedKeyFull(ctx context.Context, key *UnifiedAPIKey, req *ServerCreateRequest) (*ServerRegistrationResponse, error) {
	if !key.CanRegisterServers() {
		return nil, fmt.Errorf("API key does not have server registration capabilities")
	}

	var headers map[string]string
	switch key.GetAuthenticationMethod() {
	case "headers":
		if key.Type == APIKeyTypeRegistration {
			headers = map[string]string{
				"X-Registration-Key": key.FullToken,
			}
		} else {
			headers = map[string]string{
				"Access-Key":    key.Key,
				"Access-Secret": key.Secret,
			}
		}
	case "bearer":
		headers = map[string]string{
			"Authorization": "Bearer " + key.FullToken,
		}
	}

	var resp StandardResponse
	resp.Data = &ServerRegistrationResponse{}

	_, err := s.client.Do(ctx, &Request{
		Method:  "POST",
		Path:    "/v1/register",
		Headers: headers,
		Body:    req,
		Result:  &resp,
	})
	if err != nil {
		return nil, err
	}

	if regResp, ok := resp.Data.(*ServerRegistrationResponse); ok {
		return regResp, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// ValidateRegistrationKey validates that a unified API key can be used for server registration
func (s *ServersService) ValidateRegistrationKey(ctx context.Context, key *UnifiedAPIKey) error {
	if !key.IsActive() {
		return fmt.Errorf("registration key is not active")
	}

	if !key.CanRegisterServers() {
		return fmt.Errorf("API key does not have server registration capabilities")
	}

	// For registration keys, we might also want to check organization access
	// This would depend on the API implementation
	return nil
}

// RegisterServerQuick is a convenience method that creates a registration key and immediately uses it
func (s *ServersService) RegisterServerQuick(ctx context.Context, orgID uint, hostname, mainIP string) (*ServerRegistrationResponse, error) {
	// This would require admin privileges to create a registration key
	// For now, return an error indicating this needs to be done in two steps
	return nil, fmt.Errorf("quick registration not supported - create a registration key first using APIKeys.CreateRegistrationKey()")
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

	// Use the admin endpoint since there's no general server update endpoint
	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/admin/server/%s", serverUUID),
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
			
			// Debug enhanced hardware details
			if req.HasHardwareDetails() {
				fmt.Printf("[DEBUG] UpdateDetails: Enhanced hardware details present\n")
				if len(req.Hardware.CPU) > 0 {
					fmt.Printf("[DEBUG]   CPU Count: %d\n", len(req.Hardware.CPU))
					for i, cpu := range req.Hardware.CPU {
						fmt.Printf("[DEBUG]   CPU[%d]: %s %s (%d cores)\n", i, cpu.Manufacturer, cpu.ModelName, cpu.PhysicalCores)
					}
				}
				if req.Hardware.Memory != nil {
					fmt.Printf("[DEBUG]   Memory: %d bytes total, %s type\n", req.Hardware.Memory.TotalSize, req.Hardware.Memory.MemoryType)
				}
				if len(req.Hardware.Network) > 0 {
					fmt.Printf("[DEBUG]   Network Interfaces: %d\n", len(req.Hardware.Network))
					for i, iface := range req.Hardware.Network {
						fmt.Printf("[DEBUG]   Interface[%d]: %s (%s) - %d Mbps\n", i, iface.Name, iface.HardwareAddr, iface.SpeedMbps)
					}
				}
				if req.HasDisks() {
					fmt.Printf("[DEBUG]   Disks: %d\n", len(req.Hardware.Disks))
					for i, disk := range req.Hardware.Disks {
						fmt.Printf("[DEBUG]   Disk[%d]: %s (%s) %s - %d bytes\n", i, disk.Device, disk.DiskModel, disk.Type, disk.Size)
					}
				}
			} else {
				fmt.Printf("[DEBUG] UpdateDetails: No enhanced hardware details\n")
			}
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



