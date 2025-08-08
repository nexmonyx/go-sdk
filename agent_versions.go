package nexmonyx

import (
	"context"
	"fmt"
)

// RegisterVersion registers a new agent version
func (s *AgentVersionsService) RegisterVersion(ctx context.Context, req *AgentVersionRequest) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/agent/versions",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return err
	}

	return nil
}

// CreateVersion creates a new agent version and returns the created version
func (s *AgentVersionsService) CreateVersion(ctx context.Context, req *AgentVersionRequest) (*AgentVersion, error) {
	var resp StandardResponse
	resp.Data = &AgentVersion{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/agent/versions",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if version, ok := resp.Data.(*AgentVersion); ok {
		return version, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetVersion retrieves an agent version by version string
func (s *AgentVersionsService) GetVersion(ctx context.Context, version string) (*AgentVersion, error) {
	var resp StandardResponse
	resp.Data = &AgentVersion{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/agent/versions/%s", version),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if agentVersion, ok := resp.Data.(*AgentVersion); ok {
		return agentVersion, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// ListVersions retrieves a list of agent versions
func (s *AgentVersionsService) ListVersions(ctx context.Context, opts *ListOptions) ([]*AgentVersion, *PaginationMeta, error) {
	var resp PaginatedResponse
	var versions []*AgentVersion
	resp.Data = &versions

	req := &Request{
		Method: "GET",
		Path:   "/v1/agent/versions",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return versions, resp.Meta, nil
}

// AddBinary adds a binary for an agent version
func (s *AgentVersionsService) AddBinary(ctx context.Context, versionID uint, req *AgentBinaryRequest) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/agent/versions/%d/binaries", versionID),
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return err
	}

	return nil
}

// Admin Methods - these use the admin endpoints as seen in the register-version.sh script

// AdminCreateVersion creates a new agent version using admin endpoints
func (s *AgentVersionsService) AdminCreateVersion(ctx context.Context, version, notes string) (*AgentVersion, error) {
	var resp StandardResponse
	resp.Data = &AgentVersion{}

	body := map[string]interface{}{
		"Version": version,
		"Notes": notes,
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/admin/agent-versions",
		Body:   body,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if agentVersion, ok := resp.Data.(*AgentVersion); ok {
		return agentVersion, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// AdminAddBinary adds a binary for an agent version using admin endpoints
func (s *AgentVersionsService) AdminAddBinary(ctx context.Context, versionID uint, req *AgentBinaryRequest) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/admin/agent-versions/%d/binaries", versionID),
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return err
	}

	return nil
}