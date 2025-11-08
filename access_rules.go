package nexmonyx

import (
	"context"
	"fmt"
	"time"
)

// AccessRuleType defines the type of target for an access rule
type AccessRuleType string

const (
	AccessRuleTypeIP   AccessRuleType = "ip"
	AccessRuleTypeUser AccessRuleType = "user"
	AccessRuleTypeCIDR AccessRuleType = "cidr"
	AccessRuleTypeASN  AccessRuleType = "asn"
)

// AccessRuleAction defines the action to take for matching requests
type AccessRuleAction string

const (
	AccessRuleActionWhitelist AccessRuleAction = "whitelist"
	AccessRuleActionBlacklist AccessRuleAction = "blacklist"
)

// AccessRuleDuration defines whether a rule is temporary or permanent
type AccessRuleDuration string

const (
	AccessRuleDurationTemporary AccessRuleDuration = "temporary"
	AccessRuleDurationPermanent AccessRuleDuration = "permanent"
)

// AccessRule represents a whitelist or blacklist rule
type AccessRule struct {
	ID        uint               `json:"id"`
	RuleType  AccessRuleType     `json:"rule_type"`
	Action    AccessRuleAction   `json:"action"`
	Target    string             `json:"target"`
	Duration  AccessRuleDuration `json:"duration"`
	ExpiresAt *CustomTime        `json:"expires_at,omitempty"`
	Reason    string             `json:"reason"`
	CreatedBy uint               `json:"created_by"`
	CreatedAt *CustomTime        `json:"created_at"`
	UpdatedAt *CustomTime        `json:"updated_at"`
	Active    bool               `json:"active"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`

	// Computed fields
	IsExpired bool  `json:"is_expired"`
	IsActive  bool  `json:"is_active"`
	Creator   *User `json:"creator,omitempty"`
}

// AccessRuleCreateRequest represents a request to create an access rule
type AccessRuleCreateRequest struct {
	RuleType  AccessRuleType         `json:"rule_type"`
	Action    AccessRuleAction       `json:"action"`
	Target    string                 `json:"target"`
	Duration  AccessRuleDuration     `json:"duration"`
	ExpiresAt *time.Time             `json:"expires_at,omitempty"`
	Reason    string                 `json:"reason"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// AccessRuleUpdateRequest represents a request to update an access rule
type AccessRuleUpdateRequest struct {
	Active    *bool                  `json:"active,omitempty"`
	ExpiresAt *time.Time             `json:"expires_at,omitempty"`
	Reason    *string                `json:"reason,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// AccessRuleBulkRequest represents a request for bulk operations
type AccessRuleBulkRequest struct {
	Action  string `json:"action"` // activate, deactivate, delete
	RuleIDs []uint `json:"rule_ids"`
	Reason  string `json:"reason"`
}

// QuickBlockIPRequest represents a quick action to block an IP
type QuickBlockIPRequest struct {
	IP       string             `json:"ip"`
	Duration AccessRuleDuration `json:"duration"`
	Hours    *int               `json:"hours,omitempty"` // For temporary blocks
	Reason   string             `json:"reason"`
}

// QuickWhitelistUserRequest represents a quick action to whitelist a user
type QuickWhitelistUserRequest struct {
	UserID   uint               `json:"user_id"`
	Duration AccessRuleDuration `json:"duration"`
	Hours    *int               `json:"hours,omitempty"` // For temporary whitelist
	Reason   string             `json:"reason"`
}

// ListAccessRulesOptions represents options for filtering access rules
type ListAccessRulesOptions struct {
	RuleType       *AccessRuleType     `json:"rule_type,omitempty"`
	Action         *AccessRuleAction   `json:"action,omitempty"`
	Target         *string             `json:"target,omitempty"`
	Active         *bool               `json:"active,omitempty"`
	Duration       *AccessRuleDuration `json:"duration,omitempty"`
	IncludeDeleted bool                `json:"include_deleted,omitempty"`
	Limit          int                 `json:"limit,omitempty"`
	Offset         int                 `json:"offset,omitempty"`
}

// ListAccessRulesResponse represents the paginated response for listing access rules
type ListAccessRulesResponse struct {
	Rules      []*AccessRule `json:"rules"`
	TotalCount int64         `json:"total_count"`
	Limit      int           `json:"limit"`
	Offset     int           `json:"offset"`
}

// BulkOperationResponse represents the response for bulk operations
type BulkOperationResponse struct {
	Action      string      `json:"action"`
	RuleIDs     []uint      `json:"rule_ids"`
	Affected    int         `json:"affected"`
	ProcessedAt *CustomTime `json:"processed_at"`
}

// AccessRulesService handles access rule operations
type AccessRulesService struct {
	client *Client
}

// Create creates a new access rule
func (s *AccessRulesService) Create(ctx context.Context, req *AccessRuleCreateRequest) (*AccessRule, error) {
	var resp StandardResponse
	resp.Data = &AccessRule{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/admin/access-rules",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if rule, ok := resp.Data.(*AccessRule); ok {
		return rule, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// List retrieves a paginated list of access rules with optional filtering
func (s *AccessRulesService) List(ctx context.Context, opts *ListAccessRulesOptions) (*ListAccessRulesResponse, error) {
	var resp ListAccessRulesResponse

	req := &Request{
		Method: "GET",
		Path:   "/v1/admin/access-rules",
		Result: &resp,
	}

	// Build query parameters
	query := make(map[string]string)
	if opts != nil {
		if opts.RuleType != nil {
			query["rule_type"] = string(*opts.RuleType)
		}
		if opts.Action != nil {
			query["action"] = string(*opts.Action)
		}
		if opts.Target != nil {
			query["target"] = *opts.Target
		}
		if opts.Active != nil {
			if *opts.Active {
				query["active"] = "true"
			} else {
				query["active"] = "false"
			}
		}
		if opts.Duration != nil {
			query["duration"] = string(*opts.Duration)
		}
		if opts.IncludeDeleted {
			query["include_deleted"] = "true"
		}
		if opts.Limit > 0 {
			query["limit"] = fmt.Sprintf("%d", opts.Limit)
		}
		if opts.Offset > 0 {
			query["offset"] = fmt.Sprintf("%d", opts.Offset)
		}
	}
	req.Query = query

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// Get retrieves a specific access rule by ID
func (s *AccessRulesService) Get(ctx context.Context, ruleID uint) (*AccessRule, error) {
	var resp StandardResponse
	resp.Data = &AccessRule{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/admin/access-rules/%d", ruleID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if rule, ok := resp.Data.(*AccessRule); ok {
		return rule, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// Update updates an existing access rule
func (s *AccessRulesService) Update(ctx context.Context, ruleID uint, req *AccessRuleUpdateRequest) (*AccessRule, error) {
	var resp StandardResponse
	resp.Data = &AccessRule{}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/admin/access-rules/%d", ruleID),
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if rule, ok := resp.Data.(*AccessRule); ok {
		return rule, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// Delete performs a soft delete on an access rule
func (s *AccessRulesService) Delete(ctx context.Context, ruleID uint) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v1/admin/access-rules/%d", ruleID),
		Result: &resp,
	})
	return err
}

// QuickBlockIP performs a quick action to block an IP address
func (s *AccessRulesService) QuickBlockIP(ctx context.Context, req *QuickBlockIPRequest) (*AccessRule, error) {
	var resp StandardResponse
	resp.Data = &AccessRule{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/admin/access-rules/quick/block-ip",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if rule, ok := resp.Data.(*AccessRule); ok {
		return rule, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// QuickUnblockIP performs a quick action to unblock an IP address
func (s *AccessRulesService) QuickUnblockIP(ctx context.Context, ip string) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/admin/access-rules/quick/unblock-ip",
		Body:   map[string]string{"ip": ip},
		Result: &resp,
	})
	return err
}

// QuickWhitelistUser performs a quick action to whitelist a user
func (s *AccessRulesService) QuickWhitelistUser(ctx context.Context, req *QuickWhitelistUserRequest) (*AccessRule, error) {
	var resp StandardResponse
	resp.Data = &AccessRule{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/admin/access-rules/quick/whitelist-user",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if rule, ok := resp.Data.(*AccessRule); ok {
		return rule, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// BulkOperation performs bulk operations on multiple access rules
func (s *AccessRulesService) BulkOperation(ctx context.Context, req *AccessRuleBulkRequest) (*BulkOperationResponse, error) {
	var resp BulkOperationResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/admin/access-rules/bulk",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
