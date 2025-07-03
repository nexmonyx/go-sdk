package nexmonyx

import (
	"context"
	"fmt"
	"time"
)

// List retrieves a list of providers
func (s *ProvidersService) List(ctx context.Context, opts *ProviderListOptions) (*ProviderListResponse, *PaginationMeta, error) {
	var resp PaginatedResponse
	var providerResponse ProviderListResponse
	resp.Data = &providerResponse

	req := &Request{
		Method: "GET",
		Path:   "/v1/providers",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return &providerResponse, resp.Meta, nil
}

// Create creates a new provider
func (s *ProvidersService) Create(ctx context.Context, req *ProviderCreateRequest) (*Provider, *Response, error) {
	var resp StandardResponse
	resp.Data = &Provider{}

	apiResp, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/providers",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, nil, err
	}

	if provider, ok := resp.Data.(*Provider); ok {
		return provider, apiResp, nil
	}
	return nil, apiResp, fmt.Errorf("unexpected response type")
}

// Get retrieves a provider by ID
func (s *ProvidersService) Get(ctx context.Context, providerID string) (*Provider, error) {
	var resp StandardResponse
	resp.Data = &Provider{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/providers/%s", providerID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if provider, ok := resp.Data.(*Provider); ok {
		return provider, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// Update updates a provider
func (s *ProvidersService) Update(ctx context.Context, providerID string, req *ProviderUpdateRequest) (*Provider, error) {
	var resp StandardResponse
	resp.Data = &Provider{}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/providers/%s", providerID),
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if provider, ok := resp.Data.(*Provider); ok {
		return provider, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// Delete deletes a provider
func (s *ProvidersService) Delete(ctx context.Context, providerID string) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v1/providers/%s", providerID),
		Result: &resp,
	})
	return err
}

// Sync syncs a provider
func (s *ProvidersService) Sync(ctx context.Context, providerID string) (*SyncResponse, error) {
	var resp StandardResponse
	resp.Data = &SyncResponse{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/providers/%s/sync", providerID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if syncResp, ok := resp.Data.(*SyncResponse); ok {
		return syncResp, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// Provider represents a cloud provider
type Provider struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	ProviderType string    `json:"provider_type"`
	Status       string    `json:"status"`
	VMCount      int       `json:"vm_count"`
	Region       string    `json:"region,omitempty"`
	Description  string    `json:"description,omitempty"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
}

// ProviderListResponse represents a response containing providers
type ProviderListResponse struct {
	Providers  []Provider `json:"providers"`
	Total      int        `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	TotalPages int        `json:"total_pages"`
}

// ProviderCreateRequest represents a request to create a provider
type ProviderCreateRequest struct {
	Name         string                 `json:"name"`
	ProviderType string                 `json:"provider_type"`
	Region       string                 `json:"region,omitempty"`
	Description  string                 `json:"description,omitempty"`
	Credentials  map[string]interface{} `json:"credentials,omitempty"`
}

// ProviderUpdateRequest represents a request to update a provider
type ProviderUpdateRequest struct {
	Name         string                 `json:"name,omitempty"`
	Description  string                 `json:"description,omitempty"`
	Credentials  map[string]interface{} `json:"credentials,omitempty"`
}

// SyncResponse represents the response from a sync operation
type SyncResponse struct {
	ID          string                 `json:"id,omitempty"`
	Success     bool                   `json:"success"`
	Message     string                 `json:"message"`
	Status      string                 `json:"status,omitempty"`
	ProviderID  string                 `json:"provider_id"`
	StartedAt   string                 `json:"started_at,omitempty"`
	SyncedAt    time.Time              `json:"synced_at"`
	SyncedVMs   int                    `json:"synced_vms"`
	VMsFound    int                    `json:"vms_found,omitempty"`
	VMsAdded    int                    `json:"vms_added,omitempty"`
	VMsUpdated  int                    `json:"vms_updated,omitempty"`
	VMsRemoved  int                    `json:"vms_removed,omitempty"`
	SyncResults map[string]interface{} `json:"sync_results,omitempty"`
}

// ProviderListOptions represents options for listing providers
type ProviderListOptions struct {
	Page     int    `url:"page,omitempty"`
	PageSize int    `url:"page_size,omitempty"`
	Status   string `url:"status,omitempty"`
	Type     string `url:"type,omitempty"`
}

// ToQuery converts options to query parameters
func (o *ProviderListOptions) ToQuery() map[string]string {
	params := make(map[string]string)
	if o.Page > 0 {
		params["page"] = fmt.Sprintf("%d", o.Page)
	}
	if o.PageSize > 0 {
		params["page_size"] = fmt.Sprintf("%d", o.PageSize)
	}
	if o.Status != "" {
		params["status"] = o.Status
	}
	if o.Type != "" {
		params["type"] = o.Type
	}
	return params
}