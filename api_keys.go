package nexmonyx

import (
	"context"
	"fmt"
)

// =============================================================================
// Unified API Keys Service
// =============================================================================

// CreateUnified creates a new unified API key
func (s *APIKeysService) CreateUnified(ctx context.Context, req *CreateUnifiedAPIKeyRequest) (*CreateUnifiedAPIKeyResponse, error) {
	var resp StandardResponse
	result := &CreateUnifiedAPIKeyResponse{}
	resp.Data = result

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v2/api-keys",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetUnified retrieves a unified API key by ID
func (s *APIKeysService) GetUnified(ctx context.Context, keyID string) (*UnifiedAPIKey, error) {
	var resp StandardResponse
	resp.Data = &UnifiedAPIKey{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v2/api-keys/%s", keyID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if key, ok := resp.Data.(*UnifiedAPIKey); ok {
		return key, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// ListUnified retrieves a list of unified API keys
func (s *APIKeysService) ListUnified(ctx context.Context, opts *ListUnifiedAPIKeysOptions) ([]*UnifiedAPIKey, *PaginationMeta, error) {
	var resp PaginatedResponse
	var keys []*UnifiedAPIKey
	resp.Data = &keys

	req := &Request{
		Method: "GET",
		Path:   "/v2/api-keys",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ListOptions.ToQuery()
		
		// Add unified-specific query parameters
		if opts.Type != "" {
			req.Query["type"] = string(opts.Type)
		}
		if opts.Status != "" {
			req.Query["status"] = string(opts.Status)
		}
		if opts.UserID != 0 {
			req.Query["user_id"] = fmt.Sprintf("%d", opts.UserID)
		}
		if opts.AgentType != "" {
			req.Query["agent_type"] = opts.AgentType
		}
		if opts.RegionCode != "" {
			req.Query["region_code"] = opts.RegionCode
		}
		if opts.Namespace != "" {
			req.Query["namespace"] = opts.Namespace
		}
		if opts.Capability != "" {
			req.Query["capability"] = opts.Capability
		}
		if opts.Tag != "" {
			req.Query["tag"] = opts.Tag
		}
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return keys, resp.Meta, nil
}

// UpdateUnified updates a unified API key
func (s *APIKeysService) UpdateUnified(ctx context.Context, keyID string, req *UpdateUnifiedAPIKeyRequest) (*UnifiedAPIKey, error) {
	var resp StandardResponse
	resp.Data = &UnifiedAPIKey{}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v2/api-keys/%s", keyID),
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if key, ok := resp.Data.(*UnifiedAPIKey); ok {
		return key, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// DeleteUnified deletes a unified API key
func (s *APIKeysService) DeleteUnified(ctx context.Context, keyID string) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v2/api-keys/%s", keyID),
		Result: &resp,
	})
	return err
}

// RevokeUnified revokes a unified API key
func (s *APIKeysService) RevokeUnified(ctx context.Context, keyID string) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v2/api-keys/%s/revoke", keyID),
		Result: &resp,
	})
	return err
}

// EnableUnified enables a disabled API key
func (s *APIKeysService) EnableUnified(ctx context.Context, keyID string) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v2/admin/api-keys/%s/enable", keyID),
		Result: &resp,
	})
	return err
}

// DisableUnified disables an active API key without revoking it
func (s *APIKeysService) DisableUnified(ctx context.Context, keyID string) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v2/admin/api-keys/%s/disable", keyID),
		Result: &resp,
	})
	return err
}

// RegenerateUnified regenerates a unified API key
func (s *APIKeysService) RegenerateUnified(ctx context.Context, keyID string) (*CreateUnifiedAPIKeyResponse, error) {
	var resp StandardResponse
	result := &CreateUnifiedAPIKeyResponse{}
	resp.Data = result

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v2/api-keys/%s/regenerate", keyID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// =============================================================================
// Organization-scoped API Key operations
// =============================================================================

// CreateForOrganization creates a new API key for a specific organization
func (s *APIKeysService) CreateForOrganization(ctx context.Context, orgID string, req *CreateUnifiedAPIKeyRequest) (*CreateUnifiedAPIKeyResponse, error) {
	var resp StandardResponse
	result := &CreateUnifiedAPIKeyResponse{}
	resp.Data = result

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v2/organizations/%s/api-keys", orgID),
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ListForOrganization retrieves API keys for a specific organization
func (s *APIKeysService) ListForOrganization(ctx context.Context, orgID string, opts *ListUnifiedAPIKeysOptions) ([]*UnifiedAPIKey, *PaginationMeta, error) {
	var resp PaginatedResponse
	var keys []*UnifiedAPIKey
	resp.Data = &keys

	req := &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v2/organizations/%s/api-keys", orgID),
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ListOptions.ToQuery()
		
		// Add unified-specific query parameters
		if opts.Type != "" {
			req.Query["type"] = string(opts.Type)
		}
		if opts.Status != "" {
			req.Query["status"] = string(opts.Status)
		}
		if opts.UserID != 0 {
			req.Query["user_id"] = fmt.Sprintf("%d", opts.UserID)
		}
		if opts.AgentType != "" {
			req.Query["agent_type"] = opts.AgentType
		}
		if opts.RegionCode != "" {
			req.Query["region_code"] = opts.RegionCode
		}
		if opts.Namespace != "" {
			req.Query["namespace"] = opts.Namespace
		}
		if opts.Capability != "" {
			req.Query["capability"] = opts.Capability
		}
		if opts.Tag != "" {
			req.Query["tag"] = opts.Tag
		}
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return keys, resp.Meta, nil
}

// =============================================================================
// Admin API Key operations
// =============================================================================

// AdminCreateUnified creates a new unified API key with admin privileges
func (s *APIKeysService) AdminCreateUnified(ctx context.Context, req *CreateUnifiedAPIKeyRequest) (*CreateUnifiedAPIKeyResponse, error) {
	var resp StandardResponse
	result := &CreateUnifiedAPIKeyResponse{}
	resp.Data = result

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v2/admin/api-keys",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// AdminListUnified retrieves all unified API keys (admin only)
func (s *APIKeysService) AdminListUnified(ctx context.Context, opts *ListUnifiedAPIKeysOptions) ([]*UnifiedAPIKey, *PaginationMeta, error) {
	var resp PaginatedResponse
	var keys []*UnifiedAPIKey
	resp.Data = &keys

	req := &Request{
		Method: "GET",
		Path:   "/v2/admin/api-keys",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ListOptions.ToQuery()
		
		// Add unified-specific query parameters
		if opts.Type != "" {
			req.Query["type"] = string(opts.Type)
		}
		if opts.Status != "" {
			req.Query["status"] = string(opts.Status)
		}
		if opts.UserID != 0 {
			req.Query["user_id"] = fmt.Sprintf("%d", opts.UserID)
		}
		if opts.AgentType != "" {
			req.Query["agent_type"] = opts.AgentType
		}
		if opts.RegionCode != "" {
			req.Query["region_code"] = opts.RegionCode
		}
		if opts.Namespace != "" {
			req.Query["namespace"] = opts.Namespace
		}
		if opts.Capability != "" {
			req.Query["capability"] = opts.Capability
		}
		if opts.Tag != "" {
			req.Query["tag"] = opts.Tag
		}
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return keys, resp.Meta, nil
}

// =============================================================================
// Legacy API Key operations (backward compatibility)
// =============================================================================

// Create creates a new API key (legacy interface)
func (s *APIKeysService) Create(ctx context.Context, apiKey *APIKey) (*APIKey, error) {
	// Convert legacy APIKey to CreateUnifiedAPIKeyRequest
	req := &CreateUnifiedAPIKeyRequest{
		Name:         apiKey.Name,
		Description:  apiKey.Description,
		Type:         APIKeyTypeUser, // Default to user type for legacy keys
		Capabilities: apiKey.Scopes,  // Map scopes to capabilities
	}

	// Create the unified key
	resp, err := s.CreateUnified(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Key, nil
}

// Get retrieves an API key by ID (legacy interface)
func (s *APIKeysService) Get(ctx context.Context, id string) (*APIKey, error) {
	return s.GetUnified(ctx, id)
}

// List retrieves a list of API keys (legacy interface)
func (s *APIKeysService) List(ctx context.Context, opts *ListOptions) ([]*APIKey, *PaginationMeta, error) {
	unifiedOpts := &ListUnifiedAPIKeysOptions{
		ListOptions: *opts,
	}
	
	return s.ListUnified(ctx, unifiedOpts)
}

// Update updates an API key (legacy interface)
func (s *APIKeysService) Update(ctx context.Context, id string, apiKey *APIKey) (*APIKey, error) {
	req := &UpdateUnifiedAPIKeyRequest{
		Name:         &apiKey.Name,
		Description:  &apiKey.Description,
		Capabilities: apiKey.Scopes, // Map scopes to capabilities
	}
	
	return s.UpdateUnified(ctx, id, req)
}

// Delete deletes an API key (legacy interface)
func (s *APIKeysService) Delete(ctx context.Context, id string) error {
	return s.DeleteUnified(ctx, id)
}

// Revoke revokes an API key (legacy interface)
func (s *APIKeysService) Revoke(ctx context.Context, id string) error {
	return s.RevokeUnified(ctx, id)
}

// Regenerate regenerates an API key (legacy interface)
func (s *APIKeysService) Regenerate(ctx context.Context, id string) (*APIKey, error) {
	resp, err := s.RegenerateUnified(ctx, id)
	if err != nil {
		return nil, err
	}
	
	return resp.Key, nil
}

// =============================================================================
// Specialized key creation helpers
// =============================================================================

// CreateUserKey creates a new user API key with standard user capabilities
func (s *APIKeysService) CreateUserKey(ctx context.Context, name, description string, capabilities []string) (*CreateUnifiedAPIKeyResponse, error) {
	req := NewUserAPIKey(name, description, capabilities)
	return s.CreateUnified(ctx, req)
}

// CreateAdminKey creates a new admin API key with elevated capabilities
func (s *APIKeysService) CreateAdminKey(ctx context.Context, name, description string, capabilities []string, orgID uint) (*CreateUnifiedAPIKeyResponse, error) {
	req := NewAdminAPIKey(name, description, capabilities, orgID)
	return s.AdminCreateUnified(ctx, req)
}

// CreateMonitoringAgentKey creates a new monitoring agent key
func (s *APIKeysService) CreateMonitoringAgentKey(ctx context.Context, name, description, namespace, agentType, regionCode string, allowedScopes []string) (*CreateUnifiedAPIKeyResponse, error) {
	req := NewMonitoringAgentKey(name, description, namespace, agentType, regionCode, allowedScopes)
	return s.CreateUnified(ctx, req)
}

// CreateRegistrationKey creates a new server registration key
func (s *APIKeysService) CreateRegistrationKey(ctx context.Context, name, description string, orgID uint) (*CreateUnifiedAPIKeyResponse, error) {
	req := NewRegistrationKey(name, description, orgID)
	return s.AdminCreateUnified(ctx, req)
}

// =============================================================================
// Key validation and information helpers
// =============================================================================

// ValidateKey validates a key and returns its information
func (s *APIKeysService) ValidateKey(ctx context.Context, keyID string) (*UnifiedAPIKey, error) {
	key, err := s.GetUnified(ctx, keyID)
	if err != nil {
		return nil, err
	}

	if !key.IsActive() {
		return nil, fmt.Errorf("API key is not active")
	}

	return key, nil
}

// GetKeysByType retrieves all keys of a specific type
func (s *APIKeysService) GetKeysByType(ctx context.Context, keyType APIKeyType, opts *ListOptions) ([]*UnifiedAPIKey, *PaginationMeta, error) {
	unifiedOpts := &ListUnifiedAPIKeysOptions{
		ListOptions: *opts,
		Type:        keyType,
	}
	
	return s.ListUnified(ctx, unifiedOpts)
}

// GetActiveKeys retrieves all active keys
func (s *APIKeysService) GetActiveKeys(ctx context.Context, opts *ListOptions) ([]*UnifiedAPIKey, *PaginationMeta, error) {
	unifiedOpts := &ListUnifiedAPIKeysOptions{
		ListOptions: *opts,
		Status:      APIKeyStatusActive,
	}
	
	return s.ListUnified(ctx, unifiedOpts)
}

// GetMonitoringAgentKeys retrieves all monitoring agent keys for an organization
func (s *APIKeysService) GetMonitoringAgentKeys(ctx context.Context, orgID string, opts *ListOptions) ([]*UnifiedAPIKey, *PaginationMeta, error) {
	unifiedOpts := &ListUnifiedAPIKeysOptions{
		ListOptions: *opts,
		Type:        APIKeyTypeMonitoringAgent,
	}
	
	return s.ListForOrganization(ctx, orgID, unifiedOpts)
}

// GetRegistrationKeys retrieves all registration keys (admin only)
func (s *APIKeysService) GetRegistrationKeys(ctx context.Context, opts *ListOptions) ([]*UnifiedAPIKey, *PaginationMeta, error) {
	unifiedOpts := &ListUnifiedAPIKeysOptions{
		ListOptions: *opts,
		Type:        APIKeyTypeRegistration,
	}
	
	return s.AdminListUnified(ctx, unifiedOpts)
}