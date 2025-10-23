package nexmonyx

import (
	"context"
	"fmt"
)

// GetOrganization retrieves an organization by ID
func (s *OrganizationsService) Get(ctx context.Context, id string) (*Organization, error) {
	var resp StandardResponse
	resp.Data = &Organization{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/organizations/%s", id),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if org, ok := resp.Data.(*Organization); ok {
		return org, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// ListOrganizations retrieves a list of organizations
func (s *OrganizationsService) List(ctx context.Context, opts *ListOptions) ([]*Organization, *PaginationMeta, error) {
	var resp PaginatedResponse
	var orgs []*Organization
	resp.Data = &orgs

	req := &Request{
		Method: "GET",
		Path:   "/v1/organizations",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return orgs, resp.Meta, nil
}

// CreateOrganization creates a new organization
func (s *OrganizationsService) Create(ctx context.Context, org *Organization) (*Organization, error) {
	var resp StandardResponse
	resp.Data = &Organization{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/organizations",
		Body:   org,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if created, ok := resp.Data.(*Organization); ok {
		return created, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// UpdateOrganization updates an existing organization
func (s *OrganizationsService) Update(ctx context.Context, id string, org *Organization) (*Organization, error) {
	var resp StandardResponse
	resp.Data = &Organization{}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/organizations/%s", id),
		Body:   org,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if updated, ok := resp.Data.(*Organization); ok {
		return updated, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// DeleteOrganization deletes an organization
func (s *OrganizationsService) Delete(ctx context.Context, id string) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v1/organizations/%s", id),
		Result: &resp,
	})
	return err
}

// GetOrganizationByUUID retrieves an organization by UUID
func (s *OrganizationsService) GetByUUID(ctx context.Context, uuid string) (*Organization, error) {
	var resp StandardResponse
	resp.Data = &Organization{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/organizations/uuid/%s", uuid),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if org, ok := resp.Data.(*Organization); ok {
		return org, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetOrganizationServers retrieves servers for an organization
func (s *OrganizationsService) GetServers(ctx context.Context, id string, opts *ListOptions) ([]*Server, *PaginationMeta, error) {
	var resp PaginatedResponse
	var servers []*Server
	resp.Data = &servers

	req := &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/organizations/%s/servers", id),
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

// GetOrganizationUsers retrieves users for an organization
func (s *OrganizationsService) GetUsers(ctx context.Context, id string, opts *ListOptions) ([]*User, *PaginationMeta, error) {
	var resp PaginatedResponse
	var users []*User
	resp.Data = &users

	req := &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/organizations/%s/users", id),
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return users, resp.Meta, nil
}

// GetOrganizationAlerts retrieves alerts for an organization
func (s *OrganizationsService) GetAlerts(ctx context.Context, id string, opts *ListOptions) ([]*Alert, *PaginationMeta, error) {
	var resp PaginatedResponse
	var alerts []*Alert
	resp.Data = &alerts

	req := &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/organizations/%s/alerts", id),
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

// UpdateOrganizationSettings updates organization settings
func (s *OrganizationsService) UpdateSettings(ctx context.Context, id string, settings map[string]interface{}) (*Organization, error) {
	var resp StandardResponse
	resp.Data = &Organization{}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/organizations/%s/settings", id),
		Body:   settings,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if org, ok := resp.Data.(*Organization); ok {
		return org, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetOrganizationBilling retrieves billing information for an organization
func (s *OrganizationsService) GetBilling(ctx context.Context, id string) (map[string]interface{}, error) {
	var resp StandardResponse
	var billing map[string]interface{}
	resp.Data = &billing

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/organizations/%s/billing", id),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return billing, nil
}

// V2 Organization Infrastructure Methods

// GetAgentEndpoint retrieves the UUID-based agent ingestion endpoint for an organization
func (s *OrganizationsService) GetAgentEndpoint(ctx context.Context, id string) (map[string]interface{}, error) {
	var resp StandardResponse
	var endpoint map[string]interface{}
	resp.Data = &endpoint

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v2/organizations/%s/agent-endpoint", id),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return endpoint, nil
}

// GetInfrastructureStatus retrieves the infrastructure provisioning status for an organization
func (s *OrganizationsService) GetInfrastructureStatus(ctx context.Context, id string) (map[string]interface{}, error) {
	var resp StandardResponse
	var status map[string]interface{}
	resp.Data = &status

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v2/organizations/%s/infrastructure-status", id),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return status, nil
}

// TriggerInfrastructureProvisioning triggers infrastructure provisioning for an organization
func (s *OrganizationsService) TriggerInfrastructureProvisioning(ctx context.Context, id string) (map[string]interface{}, error) {
	var resp StandardResponse
	var result map[string]interface{}
	resp.Data = &result

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v2/organizations/%s/trigger-provisioning", id),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}
