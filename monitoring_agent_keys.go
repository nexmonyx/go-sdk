package nexmonyx

import (
	"context"
	"fmt"
)

// MonitoringAgentKey represents a monitoring agent API key
type MonitoringAgentKey struct {
	GormModel
	KeyID           string          `json:"key_id"`
	KeyPrefix       string          `json:"key_prefix"`       // First few characters for display
	OrganizationID  uint            `json:"organization_id"`
	Organization    *Organization   `json:"organization,omitempty"`
	RemoteClusterID *uint           `json:"remote_cluster_id,omitempty"`
	RemoteCluster   *RemoteCluster  `json:"remote_cluster,omitempty"`
	Namespace       string          `json:"namespace"`
	Description     string          `json:"description"`
	Capabilities    string          `json:"capabilities,omitempty"`
	Status          string          `json:"status"` // active, revoked
	LastUsedAt      *CustomTime     `json:"last_used_at,omitempty"`
	UsageCount      int             `json:"usage_count"`
}


// CreateMonitoringAgentKeyRequest represents the request to create a monitoring agent key
type CreateMonitoringAgentKeyRequest struct {
	OrganizationID  uint   `json:"organization_id"`
	RemoteClusterID *uint  `json:"remote_cluster_id,omitempty"`
	Description     string `json:"description"`
	Capabilities    string `json:"capabilities,omitempty"`
}

// CreateMonitoringAgentKeyResponse represents the response when creating a monitoring agent key
type CreateMonitoringAgentKeyResponse struct {
	KeyID     string              `json:"key_id"`
	SecretKey string              `json:"secret_key"`
	FullToken string              `json:"full_token"`
	Key       *MonitoringAgentKey `json:"key"`
}

// ListMonitoringAgentKeysOptions represents options for listing monitoring agent keys
type ListMonitoringAgentKeysOptions struct {
	Page        int    `json:"page,omitempty"`
	Limit       int    `json:"limit,omitempty"`
	Namespace   string `json:"namespace,omitempty"`
	Enabled     *bool  `json:"enabled,omitempty"`
	ClusterID   *uint  `json:"cluster_id,omitempty"`
}

// ToQuery converts options to query parameters
func (o *ListMonitoringAgentKeysOptions) ToQuery() map[string]string {
	query := make(map[string]string)
	if o.Page > 0 {
		query["page"] = fmt.Sprintf("%d", o.Page)
	}
	if o.Limit > 0 {
		query["limit"] = fmt.Sprintf("%d", o.Limit)
	}
	if o.Namespace != "" {
		query["namespace"] = o.Namespace
	}
	if o.Enabled != nil {
		query["enabled"] = fmt.Sprintf("%t", *o.Enabled)
	}
	if o.ClusterID != nil {
		query["cluster_id"] = fmt.Sprintf("%d", *o.ClusterID)
	}
	return query
}

// Admin Methods

// CreateMonitoringAgentKey creates a new monitoring agent key (admin only)
func (s *MonitoringAgentKeysService) CreateAdmin(ctx context.Context, req *CreateMonitoringAgentKeyRequest) (*CreateMonitoringAgentKeyResponse, error) {
	var resp StandardResponse
	result := &CreateMonitoringAgentKeyResponse{}
	resp.Data = result

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/admin/monitoring-agent-keys",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Customer Organization Methods

// Create creates a new monitoring agent key for the organization
func (s *MonitoringAgentKeysService) Create(ctx context.Context, organizationID string, description string) (*CreateMonitoringAgentKeyResponse, error) {
	req := &CreateMonitoringAgentKeyRequest{
		Description: description,
	}

	var resp StandardResponse
	result := &CreateMonitoringAgentKeyResponse{}
	resp.Data = result

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/organizations/%s/monitoring-agent-keys", organizationID),
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// List retrieves monitoring agent keys for an organization
func (s *MonitoringAgentKeysService) List(ctx context.Context, organizationID string, opts *ListMonitoringAgentKeysOptions) ([]*MonitoringAgentKey, *PaginationMeta, error) {
	var resp struct {
		StandardResponse
		Keys       []*MonitoringAgentKey `json:"keys"`
		Pagination *PaginationMeta       `json:"pagination"`
	}

	req := &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/organizations/%s/monitoring-agent-keys", organizationID),
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return resp.Keys, resp.Pagination, nil
}

// Revoke revokes a monitoring agent key
func (s *MonitoringAgentKeysService) Revoke(ctx context.Context, organizationID, keyID string) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/organizations/%s/monitoring-agent-keys/%s/revoke", organizationID, keyID),
		Result: &resp,
	})
	return err
}

// IsActive returns true if the monitoring agent key is active
func (k *MonitoringAgentKey) IsActive() bool {
	return k.Status == "active"
}

// IsRevoked returns true if the monitoring agent key is revoked
func (k *MonitoringAgentKey) IsRevoked() bool {
	return k.Status == "revoked"
}