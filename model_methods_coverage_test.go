package nexmonyx

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCustomTime_UnmarshalJSON_InvalidFormat tests the error path for invalid time formats
func TestCustomTime_UnmarshalJSON_InvalidFormat(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{
			name:      "invalid format - returns error",
			input:     `"not-a-valid-time-format"`,
			expectErr: true,
		},
		{
			name:      "invalid format - wrong structure",
			input:     `"12345-67-89"`,
			expectErr: true,
		},
		{
			name:      "null value - no error",
			input:     `"null"`,
			expectErr: false,
		},
		{
			name:      "empty string - no error",
			input:     `""`,
			expectErr: false,
		},
		{
			name:      "valid RFC3339 - no error",
			input:     `"2023-10-14T12:30:45Z"`,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ct CustomTime
			err := json.Unmarshal([]byte(tt.input), &ct)

			if tt.expectErr {
				assert.Error(t, err, "expected error for invalid time format")
			} else {
				assert.NoError(t, err, "expected no error for valid/null/empty input")
			}
		})
	}
}

// TestNewUserAPIKey tests the NewUserAPIKey constructor
func TestNewUserAPIKey(t *testing.T) {
	name := "test-key"
	description := "Test API Key"
	capabilities := []string{"read", "write"}

	req := NewUserAPIKey(name, description, capabilities)

	assert.NotNil(t, req)
	assert.Equal(t, name, req.Name)
	assert.Equal(t, description, req.Description)
	assert.Equal(t, APIKeyTypeUser, req.Type)
	assert.Equal(t, capabilities, req.Capabilities)
}

// TestNewAdminAPIKey tests the NewAdminAPIKey constructor
func TestNewAdminAPIKey(t *testing.T) {
	name := "admin-key"
	description := "Admin API Key"
	capabilities := []string{"admin:read", "admin:write"}
	orgID := uint(123)

	req := NewAdminAPIKey(name, description, capabilities, orgID)

	assert.NotNil(t, req)
	assert.Equal(t, name, req.Name)
	assert.Equal(t, description, req.Description)
	assert.Equal(t, APIKeyTypeAdmin, req.Type)
	assert.Equal(t, capabilities, req.Capabilities)
	assert.Equal(t, orgID, req.OrganizationID)
}

// TestNewMonitoringAgentKey tests the NewMonitoringAgentKey constructor
func TestNewMonitoringAgentKey(t *testing.T) {
	name := "monitor-key"
	description := "Monitoring Agent Key"
	namespace := "production"
	agentType := "prometheus"
	regionCode := "us-east-1"
	allowedScopes := []string{"metrics:read"}

	req := NewMonitoringAgentKey(name, description, namespace, agentType, regionCode, allowedScopes)

	assert.NotNil(t, req)
	assert.Equal(t, name, req.Name)
	assert.Equal(t, description, req.Description)
	assert.Equal(t, APIKeyTypeMonitoringAgent, req.Type)
	assert.Equal(t, namespace, req.NamespaceName)
	assert.Equal(t, agentType, req.AgentType)
	assert.Equal(t, regionCode, req.RegionCode)
	assert.Equal(t, allowedScopes, req.AllowedProbeScopes)
}

// TestNewRegistrationKey tests the NewRegistrationKey constructor
func TestNewRegistrationKey(t *testing.T) {
	name := "registration-key"
	description := "Server Registration Key"
	orgID := uint(456)

	req := NewRegistrationKey(name, description, orgID)

	assert.NotNil(t, req)
	assert.Equal(t, name, req.Name)
	assert.Equal(t, description, req.Description)
	assert.Equal(t, APIKeyTypeRegistration, req.Type)
	assert.Equal(t, orgID, req.OrganizationID)
}

// TestNewServerDetailsUpdateRequest tests the NewServerDetailsUpdateRequest constructor
func TestNewServerDetailsUpdateRequest(t *testing.T) {
	req := NewServerDetailsUpdateRequest()

	assert.NotNil(t, req)
	// Constructor returns an empty struct, just verify it's not nil
}

// TestOrganizationTagListOptions_ToQuery tests the OrganizationTagListOptions ToQuery method
func TestOrganizationTagListOptions_ToQuery(t *testing.T) {
	opts := &OrganizationTagListOptions{
		InheritOnly: true,
	}

	query := opts.ToQuery()

	assert.NotNil(t, query)
	assert.Equal(t, "true", query["inherit_only"])
}

// TestOrganizationTagListOptions_ToQuery_False tests with InheritOnly=false
func TestOrganizationTagListOptions_ToQuery_False(t *testing.T) {
	opts := &OrganizationTagListOptions{
		InheritOnly: false,
	}

	query := opts.ToQuery()

	assert.NotNil(t, query)
	// When false, inherit_only shouldn't be in the query
	_, exists := query["inherit_only"]
	assert.False(t, exists)
}

// TestServerRelationshipListOptions_ToQuery tests the ServerRelationshipListOptions ToQuery method
func TestServerRelationshipListOptions_ToQuery(t *testing.T) {
	opts := &ServerRelationshipListOptions{
		ServerID:     "server-123",
		RelationType: "parent",
		InheritOnly:  true,
	}

	query := opts.ToQuery()

	assert.NotNil(t, query)
	assert.Equal(t, "server-123", query["server_id"])
	assert.Equal(t, "parent", query["relation_type"])
	assert.Equal(t, "true", query["inherit_only"])
}

// TestServerRelationshipListOptions_ToQuery_Empty tests with zero values
func TestServerRelationshipListOptions_ToQuery_Empty(t *testing.T) {
	opts := &ServerRelationshipListOptions{}

	query := opts.ToQuery()

	assert.NotNil(t, query)
	// Empty struct should produce empty query or only false values omitted
	_, hasServerID := query["server_id"]
	assert.False(t, hasServerID)
}

// TestTagNamespaceListOptions_ToQuery_ActiveFalse tests with Active=false
func TestTagNamespaceListOptions_ToQuery_ActiveFalse(t *testing.T) {
	activeFalse := false
	opts := &TagNamespaceListOptions{
		Type:      "custom",
		Parent:    "parent-namespace",
		Active:    &activeFalse,
		Search:    "test",
		Hierarchy: true,
	}

	query := opts.ToQuery()

	assert.NotNil(t, query)
	assert.Equal(t, "custom", query["type"])
	assert.Equal(t, "parent-namespace", query["parent"])
	assert.Equal(t, "false", query["active"]) // Testing the false branch
	assert.Equal(t, "test", query["search"])
	assert.Equal(t, "true", query["hierarchy"])
}

// TestTagDetectionRuleListOptions_ToQuery_EnabledFalse tests with Enabled=false
func TestTagDetectionRuleListOptions_ToQuery_EnabledFalse(t *testing.T) {
	enabledFalse := false
	opts := &TagDetectionRuleListOptions{
		Enabled:   &enabledFalse,
		Namespace: "production",
		Page:      2,
		Limit:     50,
	}

	query := opts.ToQuery()

	assert.NotNil(t, query)
	assert.Equal(t, "false", query["enabled"]) // Testing the false branch
	assert.Equal(t, "production", query["namespace"])
	assert.Equal(t, "2", query["page"])
	assert.Equal(t, "50", query["limit"])
}
