package nexmonyx

import (
	"context"
	"fmt"
	"strconv"
)

// TagsService handles tag-related operations
type TagsService struct {
	client *Client
}

// List retrieves a list of tags with optional filtering
// Authentication: JWT Token required
// Endpoint: GET /v1/tags
// Parameters:
//   - opts: Filtering and pagination options (namespace, source, key, page, limit)
func (s *TagsService) List(ctx context.Context, opts *TagListOptions) ([]*Tag, *PaginationMeta, error) {
	var resp struct {
		Data       []*Tag          `json:"data"`
		Pagination *PaginationMeta `json:"pagination"`
		Status     string          `json:"status"`
		Message    string          `json:"message"`
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/tags",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return resp.Data, resp.Pagination, nil
}

// Create creates a new tag
// Authentication: JWT Token required
// Endpoint: POST /v1/tags
// Parameters:
//   - req: Tag creation request with namespace, key, value, and optional description
func (s *TagsService) Create(ctx context.Context, req *TagCreateRequest) (*Tag, error) {
	var resp struct {
		Data    *Tag   `json:"data"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/tags",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// GetServerTags retrieves all tags assigned to a specific server
// Authentication: JWT Token required
// Endpoint: GET /v1/server/{serverID}/tags
// Parameters:
//   - serverID: Server UUID
func (s *TagsService) GetServerTags(ctx context.Context, serverID string) ([]*ServerTag, error) {
	var resp struct {
		Data    []*ServerTag `json:"data"`
		Status  string       `json:"status"`
		Message string       `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/server/%s/tags", serverID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// AssignTagsToServer assigns one or more tags to a server
// Authentication: JWT Token required
// Endpoint: POST /v1/server/{serverID}/tags
// Parameters:
//   - serverID: Server UUID
//   - req: Tag assignment request with array of tag IDs
func (s *TagsService) AssignTagsToServer(ctx context.Context, serverID string, req *TagAssignRequest) (*TagAssignmentResult, error) {
	var resp struct {
		Data    *TagAssignmentResult `json:"data"`
		Status  string               `json:"status"`
		Message string               `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/server/%s/tags", serverID),
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// RemoveTagFromServer removes a specific tag from a server
// Authentication: JWT Token required
// Endpoint: DELETE /v1/server/{serverID}/tags/{tagID}
// Parameters:
//   - serverID: Server UUID
//   - tagID: Tag ID to remove
func (s *TagsService) RemoveTagFromServer(ctx context.Context, serverID string, tagID uint) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v1/server/%s/tags/%d", serverID, tagID),
		Result: &resp,
	})
	return err
}

// GetServersByTag retrieves all servers that have a specific tag assigned
// Authentication: JWT Token required
// Endpoint: GET /v1/tags/{tagID}/servers
// Parameters:
//   - tagID: Tag ID to get servers for
//
// Returns the tag information and list of servers with that tag
func (s *TagsService) GetServersByTag(ctx context.Context, tagID uint) (*GetServersByTagResponse, error) {
	var resp struct {
		Data    *GetServersByTagResponse `json:"data"`
		Status  string                   `json:"status"`
		Message string                   `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/tags/%d/servers", tagID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// ============================================================================
// Tag Namespace Methods
// ============================================================================

// CreateNamespace creates a new tag namespace with validation rules
// Authentication: JWT Token required
// Endpoint: POST /v1/tag-namespaces
// Parameters:
//   - req: Namespace creation request with namespace name, type, patterns, and validation rules
func (s *TagsService) CreateNamespace(ctx context.Context, req *TagNamespaceCreateRequest) (*TagNamespace, error) {
	var resp struct {
		Data    *TagNamespace `json:"data"`
		Status  string        `json:"status"`
		Message string        `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/tags/namespaces",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// ListNamespaces retrieves all namespaces for the organization with optional filtering
// Authentication: JWT Token required
// Endpoint: GET /v1/tags/namespaces
// Parameters:
//   - opts: Filtering options (type, parent, active, search, hierarchy)
//
// Returns namespaces array and total count
func (s *TagsService) ListNamespaces(ctx context.Context, opts *TagNamespaceListOptions) ([]*TagNamespace, int, error) {
	var resp struct {
		Data struct {
			Namespaces []*TagNamespace `json:"namespaces"`
			Total      int             `json:"total"`
		} `json:"data"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/tags/namespaces",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	return resp.Data.Namespaces, resp.Data.Total, nil
}

// SetNamespacePermissions sets user or role permissions for a namespace
// Authentication: JWT Token required
// Endpoint: POST /v1/tags/namespaces/{namespace}/permissions
// Parameters:
//   - namespace: Namespace name
//   - req: Permission request with user_id OR role_name and permission flags
//
// Note: Either UserID or RoleName must be provided, but not both
// WARNING: API endpoint not yet implemented - this method will return 404
func (s *TagsService) SetNamespacePermissions(ctx context.Context, namespace string, req *TagNamespacePermissionRequest) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/tags/namespaces/%s/permissions", namespace),
		Body:   req,
		Result: &resp,
	})
	return err
}

// ============================================================================
// Tag Inheritance Methods
// ============================================================================

// CreateInheritanceRule creates a new tag inheritance rule for automatic propagation
// Authentication: JWT Token required
// Endpoint: POST /v1/tag-inheritance/rules
// Parameters:
//   - req: Rule creation request with source/target types, patterns, and conditions
//
// Returns the created inheritance rule with metadata
func (s *TagsService) CreateInheritanceRule(ctx context.Context, req *TagInheritanceRuleCreateRequest) (*TagInheritanceRule, error) {
	var resp struct {
		Data    *TagInheritanceRule `json:"data"`
		Status  string              `json:"status"`
		Message string              `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/tag-inheritance/rules",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// SetOrganizationTag sets a tag at the organization level with inheritance options
// Authentication: JWT Token required
// Endpoint: POST /v1/tag-inheritance/organization-tags
// Parameters:
//   - req: Organization tag request with tag ID and inheritance settings
//
// Returns the organization tag with inheritance configuration
func (s *TagsService) SetOrganizationTag(ctx context.Context, req *OrganizationTagRequest) (*OrganizationTag, error) {
	var resp struct {
		Data    *OrganizationTag `json:"data"`
		Status  string           `json:"status"`
		Message string           `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/tag-inheritance/organization-tags",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// ListOrganizationTags retrieves all tags set at the organization level
// Authentication: JWT Token required
// Endpoint: GET /v1/tag-inheritance/organization-tags
// Parameters:
//   - opts: Filtering options (inherit_only)
//
// Returns array of organization tags and total count
func (s *TagsService) ListOrganizationTags(ctx context.Context, opts *OrganizationTagListOptions) ([]*OrganizationTag, int, error) {
	var resp struct {
		Data struct {
			Tags  []*OrganizationTag `json:"tags"`
			Total int                `json:"total"`
		} `json:"data"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/tag-inheritance/organization-tags",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	return resp.Data.Tags, resp.Data.Total, nil
}

// RemoveOrganizationTag removes a tag from the organization level
// Authentication: JWT Token required
// Endpoint: DELETE /v1/tag-inheritance/organization-tags/{tagID}
// Parameters:
//   - tagID: Tag ID to remove from organization level
//
// This stops inheritance of the tag to all servers
func (s *TagsService) RemoveOrganizationTag(ctx context.Context, tagID uint) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v1/tag-inheritance/organization-tags/%d", tagID),
		Result: &resp,
	})
	return err
}

// CreateServerRelationship creates a parent-child relationship between servers
// Authentication: JWT Token required
// Endpoint: POST /v1/tag-inheritance/server-relationships
// Parameters:
//   - req: Server relationship request with parent/child server IDs and inheritance settings
//
// Returns the created relationship with server information
func (s *TagsService) CreateServerRelationship(ctx context.Context, req *ServerRelationshipRequest) (*ServerParentRelationship, error) {
	var resp struct {
		Data    *ServerParentRelationship `json:"data"`
		Status  string                    `json:"status"`
		Message string                    `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/tag-inheritance/server-relationships",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// ListServerRelationships retrieves all server parent-child relationships
// Authentication: JWT Token required
// Endpoint: GET /v1/tag-inheritance/server-relationships
// Parameters:
//   - opts: Filtering options (server_id, relation_type, inherit_only)
//
// Returns array of server relationships and total count
func (s *TagsService) ListServerRelationships(ctx context.Context, opts *ServerRelationshipListOptions) ([]*ServerParentRelationship, int, error) {
	var resp struct {
		Data struct {
			Relationships []*ServerParentRelationship `json:"relationships"`
			Total         int                         `json:"total"`
		} `json:"data"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/tag-inheritance/server-relationships",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	return resp.Data.Relationships, resp.Data.Total, nil
}

// DeleteServerRelationship deletes a server parent-child relationship
// Authentication: JWT Token required
// Endpoint: DELETE /v1/tag-inheritance/server-relationships/{id}
// Parameters:
//   - relationshipID: Relationship ID to delete
//
// This stops tag inheritance between the parent and child servers
func (s *TagsService) DeleteServerRelationship(ctx context.Context, relationshipID uint) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v1/tag-inheritance/server-relationships/%d", relationshipID),
		Result: &resp,
	})
	return err
}

// ============================================================================
// Tag History Methods
// ============================================================================

// GetTagHistory retrieves the complete tag change history for a specific server
// Authentication: JWT Token required
// Endpoint: GET /v1/servers/{serverID}/tags/history
// Parameters:
//   - serverID: Server ID to retrieve history for
//   - opts: Query options for filtering and pagination (action, namespace, source, date range, page, limit)
//
// Returns paginated list of tag history entries with pagination metadata
func (s *TagsService) GetTagHistory(ctx context.Context, serverID uint, opts *TagHistoryQueryParams) ([]*TagHistoryResponse, *PaginationMeta, error) {
	var resp struct {
		Data       []*TagHistoryResponse `json:"data"`
		Pagination *PaginationMeta       `json:"pagination"`
		Status     string                `json:"status"`
		Message    string                `json:"message"`
		Meta       *PaginationMeta       `json:"meta"` // Alternative location for pagination
	}

	req := &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/servers/%d/tags/history", serverID),
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	// Handle both pagination and meta fields (API might use either)
	pagination := resp.Pagination
	if pagination == nil {
		pagination = resp.Meta
	}

	return resp.Data, pagination, nil
}

// GetTagHistorySummary retrieves aggregated statistics about tag changes for a specific server
// Authentication: JWT Token required
// Endpoint: GET /v1/servers/{serverID}/tags/history/summary
// Parameters:
//   - serverID: Server ID to retrieve summary for
//
// Returns aggregated statistics including:
//   - Total changes count
//   - Changes grouped by action type (added, removed, modified)
//   - Changes grouped by namespace
//   - Most active users (top 5)
//   - Recent activity stats (24h, 7d, 30d)
func (s *TagsService) GetTagHistorySummary(ctx context.Context, serverID uint) (*TagHistorySummary, error) {
	var resp struct {
		Data    *TagHistorySummary `json:"data"`
		Status  string             `json:"status"`
		Message string             `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/servers/%d/tags/history/summary", serverID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// ============================================================================
// Bulk Tag Operations
// ============================================================================

// BulkCreateTags creates multiple tags in a single operation
// Authentication: JWT Token required
// Endpoint: POST /v1/bulk/tags
// Parameters:
//   - req: Bulk tag creation request with array of tags to create
//
// Returns result with created tags, skipped tags (already exist), and counts
func (s *TagsService) BulkCreateTags(ctx context.Context, req *BulkTagCreateRequest) (*BulkTagCreateResult, error) {
	var resp struct {
		Data    *BulkTagCreateResult `json:"data"`
		Status  string               `json:"status"`
		Message string               `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/bulk/tags",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// BulkAssignTags assigns multiple tags to multiple servers in a single operation
// Authentication: JWT Token required
// Endpoint: POST /v1/bulk/tags/assign
// Parameters:
//   - req: Bulk assignment request with server IDs and tag IDs
//
// Returns result with assigned, skipped, and total assignment counts
func (s *TagsService) BulkAssignTags(ctx context.Context, req *BulkTagAssignRequest) (*BulkTagAssignResult, error) {
	var resp struct {
		Data    *BulkTagAssignResult `json:"data"`
		Status  string               `json:"status"`
		Message string               `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/bulk/tags/assign",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// AssignTagsToGroups assigns multiple servers to multiple groups in a single operation
// Authentication: JWT Token required
// Endpoint: POST /v1/bulk/groups/assign
// Parameters:
//   - req: Bulk group assignment request with server IDs and group IDs
//
// Returns result with assigned, skipped, and total assignment counts
// Note: Skips smart (automatic) groups automatically
func (s *TagsService) AssignTagsToGroups(ctx context.Context, req *BulkGroupAssignRequest) (*BulkGroupAssignResult, error) {
	var resp struct {
		Data    *BulkGroupAssignResult `json:"data"`
		Status  string                 `json:"status"`
		Message string                 `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/bulk/groups/assign",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// ============================================================================
// Tag Detection Rules
// ============================================================================

// ListTagDetectionRules retrieves all tag detection rules for the organization
//
// Endpoint: GET /v1/tag-rules
// Authentication: JWT Token or API Key required
//
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - opts: Optional filtering parameters (enabled status, namespace, pagination)
//
// Returns:
//   - rules: List of tag detection rules matching the filters
//   - total: Total number of rules matching the filters
//   - err: Error if the request fails
func (s *TagsService) ListTagDetectionRules(ctx context.Context, opts *TagDetectionRuleListOptions) ([]*TagDetectionRule, int, error) {
	var resp struct {
		Data       []*TagDetectionRule `json:"data"`
		Status     string              `json:"status"`
		Message    string              `json:"message"`
		Pagination *PaginationMeta     `json:"pagination"`
		Meta       *PaginationMeta     `json:"meta"` // Alternative location
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/tag-rules",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	totalItems := 0
	if resp.Pagination != nil {
		totalItems = resp.Pagination.TotalItems
	} else if resp.Meta != nil {
		totalItems = resp.Meta.TotalItems
	}

	return resp.Data, totalItems, nil
}

// CreateDefaultRules creates default tag detection rules for the organization
//
// Endpoint: POST /v1/tag-rules/defaults
// Authentication: JWT Token or API Key required
//
// This creates a standard set of detection rules for common infrastructure patterns.
//
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//
// Returns:
//   - result: Result containing the count of created rules
//   - err: Error if the request fails
func (s *TagsService) CreateDefaultRules(ctx context.Context) (*DefaultRulesCreateResult, error) {
	var resp struct {
		Data    *DefaultRulesCreateResult `json:"data"`
		Status  string                    `json:"status"`
		Message string                    `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/tag-rules/defaults",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// EvaluateRules evaluates tag detection rules for specified servers
//
// Endpoint: POST /v1/tag-rules/evaluate
// Authentication: JWT Token or API Key required
//
// This triggers asynchronous evaluation of tag detection rules. Rules are evaluated
// in the background and matching tags are automatically assigned to servers.
//
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - req: Request specifying which servers to evaluate (specific IDs or all servers)
//
// Returns:
//   - result: Result containing the count of servers queued for processing
//   - err: Error if the request fails
func (s *TagsService) EvaluateRules(ctx context.Context, req *EvaluateRulesRequest) (*EvaluateRulesResult, error) {
	var resp struct {
		Data    *EvaluateRulesResult `json:"data"`
		Status  string               `json:"status"`
		Message string               `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/tag-rules/evaluate",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// CreateTagDetectionRule creates a new tag detection rule
//
// Endpoint: POST /v1/tag-rules
// Authentication: JWT Token or API Key required
//
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - req: Rule creation request with name, conditions, tag to apply, priority, etc.
//
// Returns:
//   - rule: The created tag detection rule
//   - err: Error if the request fails
func (s *TagsService) CreateTagDetectionRule(ctx context.Context, req *TagDetectionRuleCreateRequest) (*TagDetectionRule, error) {
	var resp struct {
		Data    *TagDetectionRule `json:"data"`
		Status  string            `json:"status"`
		Message string            `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/tag-rules",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// GetTagDetectionRule retrieves a single tag detection rule by ID
//
// Endpoint: GET /v1/tag-rules/{id}
// Authentication: JWT Token or API Key required
//
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - ruleID: Rule ID to retrieve
//
// Returns:
//   - rule: The tag detection rule
//   - err: Error if the request fails
func (s *TagsService) GetTagDetectionRule(ctx context.Context, ruleID uint) (*TagDetectionRule, error) {
	var resp struct {
		Data    *TagDetectionRule `json:"data"`
		Status  string            `json:"status"`
		Message string            `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/tag-rules/%d", ruleID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// UpdateTagDetectionRule updates an existing tag detection rule
//
// Endpoint: PUT /v1/tag-rules/{id}
// Authentication: JWT Token or API Key required
//
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - ruleID: Rule ID to update
//   - req: Update request with modified fields
//
// Returns:
//   - rule: The updated tag detection rule
//   - err: Error if the request fails
func (s *TagsService) UpdateTagDetectionRule(ctx context.Context, ruleID uint, req *TagDetectionRuleUpdateRequest) (*TagDetectionRule, error) {
	var resp struct {
		Data    *TagDetectionRule `json:"data"`
		Status  string            `json:"status"`
		Message string            `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/tag-rules/%d", ruleID),
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// DeleteTagDetectionRule deletes a tag detection rule
//
// Endpoint: DELETE /v1/tag-rules/{id}
// Authentication: JWT Token or API Key required
//
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - ruleID: Rule ID to delete
//
// Returns:
//   - err: Error if the request fails
func (s *TagsService) DeleteTagDetectionRule(ctx context.Context, ruleID uint) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v1/tag-rules/%d", ruleID),
		Result: &resp,
	})
	return err
}

// ExecuteTagDetectionRule manually executes a tag detection rule against servers
//
// Endpoint: POST /v1/tag-rules/{id}/execute
// Authentication: JWT Token or API Key required
//
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - ruleID: Rule ID to execute
//   - req: Execution options (server IDs to evaluate, dry run mode)
//
// Returns:
//   - result: Execution result with matched servers and applied tags
//   - err: Error if the request fails
func (s *TagsService) ExecuteTagDetectionRule(ctx context.Context, ruleID uint, req *ExecuteRuleRequest) (*ExecuteRuleResponse, error) {
	var resp struct {
		Data    *ExecuteRuleResponse `json:"data"`
		Status  string               `json:"status"`
		Message string               `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/tag-rules/%d/execute", ruleID),
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// GetTagDetectionRuleExecutions retrieves the execution history for a rule
//
// Endpoint: GET /v1/tag-rules/{id}/executions
// Authentication: JWT Token or API Key required
//
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - ruleID: Rule ID to get executions for
//   - opts: Pagination options (page, limit)
//
// Returns:
//   - result: Execution history with server tags applied by this rule
//   - err: Error if the request fails
func (s *TagsService) GetTagDetectionRuleExecutions(ctx context.Context, ruleID uint, opts *RuleExecutionListOptions) (*RuleExecutionsResponse, error) {
	var resp struct {
		Data       *RuleExecutionsResponse `json:"data"`
		Status     string                  `json:"status"`
		Message    string                  `json:"message"`
		Pagination *PaginationMeta         `json:"pagination"`
	}

	req := &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/tag-rules/%d/executions", ruleID),
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// GetTagDetectionRuleStatistics retrieves aggregated statistics for tag detection rules
//
// Endpoint: GET /v1/tag-rules/statistics
// Authentication: JWT Token or API Key required
//
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - opts: Filter options (enabled status, namespace)
//
// Returns:
//   - stats: Aggregated rule statistics
//   - err: Error if the request fails
func (s *TagsService) GetTagDetectionRuleStatistics(ctx context.Context, opts *RuleStatisticsOptions) (*RuleStatisticsResponse, error) {
	var resp struct {
		Data    *RuleStatisticsResponse `json:"data"`
		Status  string                  `json:"status"`
		Message string                  `json:"message"`
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/tag-rules/statistics",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// ============================================================================
// Tag CRUD Methods
// ============================================================================

// GetTag retrieves a single tag by ID
// Authentication: JWT Token or API Key required
// Endpoint: GET /v1/tags/{id}
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - tagID: Tag ID to retrieve
//
// Returns the tag details including server count
func (s *TagsService) GetTag(ctx context.Context, tagID uint) (*Tag, error) {
	var resp struct {
		Data    *Tag   `json:"data"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/tags/%d", tagID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// UpdateTag updates an existing tag's description
// Authentication: JWT Token required
// Endpoint: PUT /v1/tags/{id}
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - tagID: Tag ID to update
//   - req: Update request with new description
//
// Returns the updated tag
func (s *TagsService) UpdateTag(ctx context.Context, tagID uint, req *TagUpdateRequest) (*Tag, error) {
	var resp struct {
		Data    *Tag   `json:"data"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/tags/%d", tagID),
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// DeleteTag deletes a tag by ID
// Authentication: JWT Token required
// Endpoint: DELETE /v1/tags/{id}
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - tagID: Tag ID to delete
//   - cascade: If true, removes tag from all servers. If false, fails if tag is assigned (default: true)
//
// Returns deletion result with count of removed server associations
func (s *TagsService) DeleteTag(ctx context.Context, tagID uint, cascade bool) (*TagDeleteResult, error) {
	var resp struct {
		Data    *TagDeleteResult `json:"data"`
		Status  string           `json:"status"`
		Message string           `json:"message"`
	}

	req := &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v1/tags/%d", tagID),
		Result: &resp,
	}

	// Add cascade query parameter
	req.Query = map[string]string{
		"cascade": strconv.FormatBool(cascade),
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// ============================================================================
// Namespace CRUD Methods
// ============================================================================

// UpdateNamespace updates an existing namespace definition
// Authentication: JWT Token required
// Endpoint: PUT /v1/tags/namespaces/{id}
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - namespaceID: Namespace ID to update
//   - req: Update request with new configuration
//
// Returns the updated namespace
func (s *TagsService) UpdateNamespace(ctx context.Context, namespaceID uint, req *NamespaceUpdateRequest) (*TagNamespace, error) {
	var resp struct {
		Data    *TagNamespace `json:"data"`
		Status  string        `json:"status"`
		Message string        `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/tags/namespaces/%d", namespaceID),
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// DeleteNamespace deletes a namespace definition
// Authentication: JWT Token required
// Endpoint: DELETE /v1/tags/namespaces/{id}
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - namespaceID: Namespace ID to delete
//   - cascade: If true, deletes associated tags and child namespaces. If false, fails if namespace has dependencies (default: false)
//
// Returns deletion result with counts of deleted resources
func (s *TagsService) DeleteNamespace(ctx context.Context, namespaceID uint, cascade bool) (*NamespaceDeleteResult, error) {
	var resp struct {
		Data    *NamespaceDeleteResult `json:"data"`
		Status  string                 `json:"status"`
		Message string                 `json:"message"`
	}

	req := &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v1/tags/namespaces/%d", namespaceID),
		Result: &resp,
	}

	// Add cascade query parameter
	req.Query = map[string]string{
		"cascade": strconv.FormatBool(cascade),
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// ============================================================================
// Tag-to-Tag Inheritance Methods
// ============================================================================

// SetTagInheritance creates a tag-to-tag inheritance relationship
// Authentication: JWT Token required
// Endpoint: POST /v1/tags/{id}/inherit
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - childTagID: Child tag ID that will inherit from parent
//   - parentTagID: Parent tag ID to inherit from
//
// Creates inheritance with circular reference and depth limit (max 10 levels) checks
// Returns the created inheritance relationship
func (s *TagsService) SetTagInheritance(ctx context.Context, childTagID uint, parentTagID uint) (*TagInheritanceRelationship, error) {
	var resp struct {
		Data    *TagInheritanceRelationship `json:"data"`
		Status  string                      `json:"status"`
		Message string                      `json:"message"`
	}

	reqBody := map[string]interface{}{
		"parent_tag_id": parentTagID,
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/tags/%d/inherit", childTagID),
		Body:   reqBody,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// GetInheritedTags retrieves the complete inheritance chain for a tag
// Authentication: JWT Token required
// Endpoint: GET /v1/tags/{id}/inherited
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - tagID: Tag ID to get inheritance chain for
//
// Returns inheritance chain from immediate parent to root, with depth and count
func (s *TagsService) GetInheritedTags(ctx context.Context, tagID uint) (*TagInheritanceChain, error) {
	var resp struct {
		Data    *TagInheritanceChain `json:"data"`
		Status  string               `json:"status"`
		Message string               `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/tags/%d/inherited", tagID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// RemoveTagInheritance removes tag-to-tag inheritance relationships
// Authentication: JWT Token required
// Endpoint: DELETE /v1/tags/{id}/inherit
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - childTagID: Child tag ID to remove inheritance from
//   - cascade: If true, removes all descendant relationships. If false, only removes direct parent relationship (default: false)
//
// Returns count of deleted inheritance relationships
func (s *TagsService) RemoveTagInheritance(ctx context.Context, childTagID uint, cascade bool) (*TagInheritanceDeleteResult, error) {
	var resp struct {
		Data    *TagInheritanceDeleteResult `json:"data"`
		Status  string                      `json:"status"`
		Message string                      `json:"message"`
	}

	req := &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v1/tags/%d/inherit", childTagID),
		Result: &resp,
	}

	// Add cascade query parameter
	req.Query = map[string]string{
		"cascade": strconv.FormatBool(cascade),
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}
