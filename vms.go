package nexmonyx

import (
	"context"
	"fmt"
)

// VMsService handles virtual machine management operations
type VMsService struct {
	client *Client
}

// List retrieves a list of virtual machines
// Authentication: JWT Token required
// Endpoint: GET /api/v1/vms
// Parameters:
//   - opts: Optional pagination options
// Returns: Array of VirtualMachine objects with pagination metadata
func (s *VMsService) List(ctx context.Context, opts *PaginationOptions) ([]VirtualMachine, *PaginationMeta, error) {
	var resp struct {
		Data []VirtualMachine `json:"data"`
		Meta *PaginationMeta  `json:"meta"`
	}

	query := make(map[string]string)
	if opts != nil {
		if opts.Page > 0 {
			query["page"] = fmt.Sprintf("%d", opts.Page)
		}
		if opts.Limit > 0 {
			query["limit"] = fmt.Sprintf("%d", opts.Limit)
		}
	}

	req := &Request{
		Method: "GET",
		Path:   "/api/v1/vms",
		Result: &resp,
	}
	if len(query) > 0 {
		req.Query = query
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return resp.Data, resp.Meta, nil
}

// Create creates a new virtual machine
// Authentication: JWT Token required
// Endpoint: POST /api/v1/vms
// Parameters:
//   - config: VM configuration (name, CPU, memory, storage, etc.)
// Returns: Created VirtualMachine object
func (s *VMsService) Create(ctx context.Context, config *VMConfiguration) (*VirtualMachine, error) {
	var resp struct {
		Data *VirtualMachine `json:"data"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/api/v1/vms",
		Body:   config,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// Get retrieves details of a specific virtual machine
// Authentication: JWT Token required
// Endpoint: GET /api/v1/vms/{id}
// Parameters:
//   - vmID: Virtual machine ID
// Returns: VirtualMachine object with full details
func (s *VMsService) Get(ctx context.Context, vmID uint) (*VirtualMachine, error) {
	var resp struct {
		Data *VirtualMachine `json:"data"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/api/v1/vms/%d", vmID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// Delete deletes a virtual machine
// Authentication: JWT Token required
// Endpoint: DELETE /api/v2/organizations/{orgId}/virtual-machines/{vmId}
// Parameters:
//   - orgID: Organization ID
//   - vmID: Virtual machine ID
// Returns: Success confirmation
func (s *VMsService) Delete(ctx context.Context, orgID uint, vmID uint) error {
	_, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/api/v2/organizations/%d/virtual-machines/%d", orgID, vmID),
	})
	return err
}

// Start starts a stopped virtual machine
// Authentication: JWT Token required
// Endpoint: POST /api/v2/organizations/{orgId}/virtual-machines/{vmId}/start
// Parameters:
//   - orgID: Organization ID
//   - vmID: Virtual machine ID
// Returns: VMOperation object with operation status
func (s *VMsService) Start(ctx context.Context, orgID uint, vmID uint) (*VMOperation, error) {
	var resp struct {
		Data *VMOperation `json:"data"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/api/v2/organizations/%d/virtual-machines/%d/start", orgID, vmID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// Stop stops a running virtual machine
// Authentication: JWT Token required
// Endpoint: POST /api/v2/organizations/{orgId}/virtual-machines/{vmId}/stop
// Parameters:
//   - orgID: Organization ID
//   - vmID: Virtual machine ID
//   - force: Optional force flag for immediate shutdown (default: graceful)
// Returns: VMOperation object with operation status
func (s *VMsService) Stop(ctx context.Context, orgID uint, vmID uint, force bool) (*VMOperation, error) {
	var resp struct {
		Data *VMOperation `json:"data"`
	}

	body := map[string]interface{}{}
	if force {
		body["force"] = true
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/api/v2/organizations/%d/virtual-machines/%d/stop", orgID, vmID),
		Body:   body,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// Restart restarts a virtual machine
// Authentication: JWT Token required
// Endpoint: POST /api/v2/organizations/{orgId}/virtual-machines/{vmId}/restart
// Parameters:
//   - orgID: Organization ID
//   - vmID: Virtual machine ID
//   - force: Optional force flag for immediate restart (default: graceful)
// Returns: VMOperation object with operation status
func (s *VMsService) Restart(ctx context.Context, orgID uint, vmID uint, force bool) (*VMOperation, error) {
	var resp struct {
		Data *VMOperation `json:"data"`
	}

	body := map[string]interface{}{}
	if force {
		body["force"] = true
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/api/v2/organizations/%d/virtual-machines/%d/restart", orgID, vmID),
		Body:   body,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}
