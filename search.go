package nexmonyx

import (
	"context"
	"fmt"
)

// SearchService handles search operations across servers, tags, and other resources
type SearchService struct {
	client *Client
}

// SearchServers performs a comprehensive search across servers
// Authentication: JWT Token required
// Endpoint: GET /v1/search/servers
// Parameters:
//   - query: Search query string (searches name, UUID, hostname, IP addresses, tags)
//   - opts: Optional pagination options
//   - filters: Optional filters (location, environment, status, classification)
// Returns: Array of SearchResult objects with pagination metadata
func (s *SearchService) SearchServers(ctx context.Context, query string, opts *PaginationOptions, filters map[string]interface{}) ([]SearchResult, *PaginationMeta, error) {
	var resp struct {
		Data []SearchResult  `json:"data"`
		Meta *PaginationMeta `json:"meta"`
	}

	queryParams := make(map[string]string)
	if query != "" {
		queryParams["query"] = query
	}
	if opts != nil {
		if opts.Page > 0 {
			queryParams["page"] = fmt.Sprintf("%d", opts.Page)
		}
		if opts.Limit > 0 {
			queryParams["limit"] = fmt.Sprintf("%d", opts.Limit)
		}
	}

	// Add filter parameters
	if filters != nil {
		if location, ok := filters["location"].(string); ok && location != "" {
			queryParams["location"] = location
		}
		if environment, ok := filters["environment"].(string); ok && environment != "" {
			queryParams["environment"] = environment
		}
		if status, ok := filters["status"].(string); ok && status != "" {
			queryParams["status"] = status
		}
		if classification, ok := filters["classification"].(string); ok && classification != "" {
			queryParams["classification"] = classification
		}
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/search/servers",
		Result: &resp,
	}
	if len(queryParams) > 0 {
		req.Query = queryParams
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return resp.Data, resp.Meta, nil
}

// SearchTags searches for tags across the system
// Authentication: JWT Token required
// Endpoint: GET /v1/search/tags
// Parameters:
//   - query: Search query string (searches tag names and descriptions)
//   - opts: Optional pagination options
//   - filters: Optional filters (tag_type, scope)
// Returns: Array of TagSearchResult objects with pagination metadata
func (s *SearchService) SearchTags(ctx context.Context, query string, opts *PaginationOptions, filters map[string]interface{}) ([]TagSearchResult, *PaginationMeta, error) {
	var resp struct {
		Data []TagSearchResult `json:"data"`
		Meta *PaginationMeta   `json:"meta"`
	}

	queryParams := make(map[string]string)
	if query != "" {
		queryParams["query"] = query
	}
	if opts != nil {
		if opts.Page > 0 {
			queryParams["page"] = fmt.Sprintf("%d", opts.Page)
		}
		if opts.Limit > 0 {
			queryParams["limit"] = fmt.Sprintf("%d", opts.Limit)
		}
	}

	// Add filter parameters
	if filters != nil {
		if tagType, ok := filters["tag_type"].(string); ok && tagType != "" {
			queryParams["tag_type"] = tagType
		}
		if scope, ok := filters["scope"].(string); ok && scope != "" {
			queryParams["scope"] = scope
		}
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/search/tags",
		Result: &resp,
	}
	if len(queryParams) > 0 {
		req.Query = queryParams
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return resp.Data, resp.Meta, nil
}

// GetTagStatistics retrieves comprehensive statistics about tag usage
// Authentication: JWT Token required
// Endpoint: GET /v1/search/tags/statistics
// Parameters:
//   - tagType: Optional filter by tag type (manual, auto, system)
//   - scope: Optional filter by scope (organization, user, server)
// Returns: TagStatistics object with comprehensive tag usage data
func (s *SearchService) GetTagStatistics(ctx context.Context, tagType string, scope string) (*TagStatistics, error) {
	var resp struct {
		Data    *TagStatistics `json:"data"`
		Status  string         `json:"status"`
		Message string         `json:"message"`
	}

	queryParams := make(map[string]string)
	if tagType != "" {
		queryParams["tag_type"] = tagType
	}
	if scope != "" {
		queryParams["scope"] = scope
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/search/tags/statistics",
		Result: &resp,
	}
	if len(queryParams) > 0 {
		req.Query = queryParams
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}
