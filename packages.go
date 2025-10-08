package nexmonyx

import (
	"context"
)

// PackagesService handles organization package/tier management and limits
type PackagesService struct {
	client *Client
}

// GetAvailablePackageTiers retrieves information about all available package tiers
// Authentication: Public (no authentication required)
// Endpoint: GET /v1/package/tiers
// Returns: Map of available package tiers with their features and limits
func (s *PackagesService) GetAvailablePackageTiers(ctx context.Context) (map[string]interface{}, error) {
	var resp struct {
		Data    map[string]interface{} `json:"data"`
		Status  string                 `json:"status"`
		Message string                 `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/package/tiers",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// GetOrganizationPackage retrieves the current package information for the authenticated organization
// Authentication: JWT Token required
// Endpoint: GET /v1/organization/package
// Returns: OrganizationPackage with current limits and usage information
func (s *PackagesService) GetOrganizationPackage(ctx context.Context) (*OrganizationPackage, error) {
	var resp struct {
		Data    *OrganizationPackage `json:"data"`
		Status  string               `json:"status"`
		Message string               `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/organization/package",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// UpgradeOrganizationPackage upgrades the organization to a new package tier
// Authentication: JWT Token required
// Endpoint: POST /v1/organization/package/upgrade
// Parameters:
//   - req: Upgrade request containing new tier and optional payment information
// Returns: Updated OrganizationPackage
func (s *PackagesService) UpgradeOrganizationPackage(ctx context.Context, req *PackageUpgradeRequest) (*OrganizationPackage, error) {
	var resp struct {
		Data    *OrganizationPackage `json:"data"`
		Status  string               `json:"status"`
		Message string               `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/organization/package/upgrade",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// ValidateProbeConfig validates if a probe configuration is allowed under the current package limits
// Authentication: JWT Token required
// Endpoint: POST /v1/organization/package/validate-probe-config
// Parameters:
//   - req: Probe configuration to validate
// Returns: Validation result indicating if configuration is allowed
func (s *PackagesService) ValidateProbeConfig(ctx context.Context, req *ProbeConfigValidationRequest) (*ProbeConfigValidationResult, error) {
	var resp struct {
		Data    *ProbeConfigValidationResult `json:"data"`
		Status  string                       `json:"status"`
		Message string                       `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/organization/package/validate-probe-config",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}
