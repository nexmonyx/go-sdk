package nexmonyx

import (
	"context"
)

// RegionsService handles monitoring region operations
type RegionsService struct {
	client *Client
}

// PublicRegion represents a simplified region for public monitoring
// This matches the response from GET /v1/regions endpoint
type PublicRegion struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	Country   string `json:"country"`
	Continent string `json:"continent"`
}

// List returns all available public monitoring regions
// GET /v1/regions
func (s *RegionsService) List(ctx context.Context) ([]*PublicRegion, error) {
	var result struct {
		Status string          `json:"status"`
		Data   []*PublicRegion `json:"data"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/regions",
		Result: &result,
	})
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}
