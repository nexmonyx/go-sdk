package nexmonyx

import (
	"context"
	"fmt"
	"time"
)

// CreateAPIKey creates a new API key
func (s *APIKeysService) Create(ctx context.Context, apiKey *APIKey) (*APIKey, error) {
	var resp StandardResponse
	resp.Data = &APIKey{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/api-keys",
		Body:   apiKey,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if created, ok := resp.Data.(*APIKey); ok {
		return created, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetAPIKey retrieves an API key by ID
func (s *APIKeysService) Get(ctx context.Context, id string) (*APIKey, error) {
	var resp StandardResponse
	resp.Data = &APIKey{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/api-keys/%s", id),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if apiKey, ok := resp.Data.(*APIKey); ok {
		return apiKey, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// ListAPIKeys retrieves a list of API keys
func (s *APIKeysService) List(ctx context.Context, opts *ListOptions) ([]*APIKey, *PaginationMeta, error) {
	var resp PaginatedResponse
	var apiKeys []*APIKey
	resp.Data = &apiKeys

	req := &Request{
		Method: "GET",
		Path:   "/v1/api-keys",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return apiKeys, resp.Meta, nil
}

// UpdateAPIKey updates an API key
func (s *APIKeysService) Update(ctx context.Context, id string, apiKey *APIKey) (*APIKey, error) {
	var resp StandardResponse
	resp.Data = &APIKey{}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/api-keys/%s", id),
		Body:   apiKey,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if updated, ok := resp.Data.(*APIKey); ok {
		return updated, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// DeleteAPIKey deletes an API key
func (s *APIKeysService) Delete(ctx context.Context, id string) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v1/api-keys/%s", id),
		Result: &resp,
	})
	return err
}

// RevokeAPIKey revokes an API key
func (s *APIKeysService) Revoke(ctx context.Context, id string) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/api-keys/%s/revoke", id),
		Result: &resp,
	})
	return err
}

// RegenerateAPIKey regenerates an API key
func (s *APIKeysService) Regenerate(ctx context.Context, id string) (*APIKey, error) {
	var resp StandardResponse
	resp.Data = &APIKey{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/api-keys/%s/regenerate", id),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if apiKey, ok := resp.Data.(*APIKey); ok {
		return apiKey, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// APIKey represents an API key
type APIKey struct {
	GormModel
	Name             string                 `json:"name"`
	Description      string                 `json:"description,omitempty"`
	Key              string                 `json:"key,omitempty"`    // Only returned on creation
	Secret           string                 `json:"secret,omitempty"` // Only returned on creation
	KeyPrefix        string                 `json:"key_prefix"`       // First few characters of the key
	OrganizationID   uint                   `json:"organization_id"`
	UserID           uint                   `json:"user_id"`
	User             *User                  `json:"user,omitempty"`
	Scopes           []string               `json:"scopes"`
	Status           string                 `json:"status"` // active, revoked, expired
	ExpiresAt        *CustomTime            `json:"expires_at,omitempty"`
	LastUsedAt       *CustomTime            `json:"last_used_at,omitempty"`
	LastUsedIP       string                 `json:"last_used_ip,omitempty"`
	UsageCount       int                    `json:"usage_count"`
	RateLimitPerHour int                    `json:"rate_limit_per_hour,omitempty"`
	AllowedIPs       []string               `json:"allowed_ips,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// IsActive returns true if the API key is active
func (k *APIKey) IsActive() bool {
	return k.Status == "active" && (k.ExpiresAt == nil || k.ExpiresAt.After(time.Now()))
}

// IsExpired returns true if the API key is expired
func (k *APIKey) IsExpired() bool {
	return k.ExpiresAt != nil && k.ExpiresAt.Before(time.Now())
}

// HasScope returns true if the API key has the specified scope
func (k *APIKey) HasScope(scope string) bool {
	for _, s := range k.Scopes {
		if s == scope || s == "*" {
			return true
		}
	}
	return false
}
