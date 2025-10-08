package nexmonyx

import (
	"context"
	"fmt"
)

// ServerGroupsService handles server group management operations
type ServerGroupsService struct {
	client *Client
}

// CreateGroup creates a new server group
// Authentication: JWT Token required
// Endpoint: POST /v1/groups
// Parameters:
//   - name: Group name
//   - description: Optional group description
//   - tags: Optional tags for group organization
// Returns: Created ServerGroup object
func (s *ServerGroupsService) CreateGroup(ctx context.Context, name string, description string, tags []string) (*ServerGroup, error) {
	var resp struct {
		Data    *ServerGroup `json:"data"`
		Status  string       `json:"status"`
		Message string       `json:"message"`
	}

	body := map[string]interface{}{
		"name": name,
	}
	if description != "" {
		body["description"] = description
	}
	if len(tags) > 0 {
		body["tags"] = tags
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/groups",
		Body:   body,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// ListGroups retrieves a list of server groups
// Authentication: JWT Token required
// Endpoint: GET /v1/groups
// Parameters:
//   - opts: Optional pagination options
//   - nameFilter: Optional filter for group name search
//   - tags: Optional filter by tags
// Returns: Array of ServerGroup objects with pagination metadata
func (s *ServerGroupsService) ListGroups(ctx context.Context, opts *PaginationOptions, nameFilter string, tags []string) ([]ServerGroup, *PaginationMeta, error) {
	var resp struct {
		Data []ServerGroup   `json:"data"`
		Meta *PaginationMeta `json:"meta"`
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
	if nameFilter != "" {
		query["name"] = nameFilter
	}
	if len(tags) > 0 {
		// Join tags with comma for query parameter
		tagStr := ""
		for i, tag := range tags {
			if i > 0 {
				tagStr += ","
			}
			tagStr += tag
		}
		query["tags"] = tagStr
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/groups",
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

// AddServersToGroup adds servers to a group
// Authentication: JWT Token required
// Endpoint: POST /v1/groups/{groupID}/servers
// Parameters:
//   - groupID: Group ID
//   - serverIDs: Array of server IDs to add
//   - serverUUIDs: Array of server UUIDs to add
// Returns: Number of servers added
func (s *ServerGroupsService) AddServersToGroup(ctx context.Context, groupID uint, serverIDs []uint, serverUUIDs []string) (int, error) {
	var resp struct {
		Data struct {
			ServersAdded int `json:"servers_added"`
		} `json:"data"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	body := map[string]interface{}{}
	if len(serverIDs) > 0 {
		body["server_ids"] = serverIDs
	}
	if len(serverUUIDs) > 0 {
		body["server_uuids"] = serverUUIDs
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/groups/%d/servers", groupID),
		Body:   body,
		Result: &resp,
	})
	if err != nil {
		return 0, err
	}

	return resp.Data.ServersAdded, nil
}

// GetGroupServers retrieves servers in a specific group
// Authentication: JWT Token required
// Endpoint: GET /v1/groups/{groupID}/servers
// Parameters:
//   - groupID: Group ID
//   - opts: Optional pagination options
//   - status: Optional filter by server status
//   - tags: Optional filter by server tags
// Returns: Array of ServerGroupMembership objects with pagination metadata
func (s *ServerGroupsService) GetGroupServers(ctx context.Context, groupID uint, opts *PaginationOptions, status string, tags []string) ([]ServerGroupMembership, *PaginationMeta, error) {
	var resp struct {
		Data []ServerGroupMembership `json:"data"`
		Meta *PaginationMeta         `json:"meta"`
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
	if status != "" {
		query["status"] = status
	}
	if len(tags) > 0 {
		tagStr := ""
		for i, tag := range tags {
			if i > 0 {
				tagStr += ","
			}
			tagStr += tag
		}
		query["tags"] = tagStr
	}

	req := &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/groups/%d/servers", groupID),
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
