package nexmonyx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Tests for MonitoringAgentKey helper methods

func TestMonitoringAgentKey_IsActive(t *testing.T) {
	// Active key
	activeKey := &MonitoringAgentKey{
		Status: "active",
	}
	assert.True(t, activeKey.IsActive())

	// Revoked key
	revokedKey := &MonitoringAgentKey{
		Status: "revoked",
	}
	assert.False(t, revokedKey.IsActive())

	// Other status
	otherKey := &MonitoringAgentKey{
		Status: "pending",
	}
	assert.False(t, otherKey.IsActive())
}

func TestMonitoringAgentKey_IsRevoked(t *testing.T) {
	// Revoked key
	revokedKey := &MonitoringAgentKey{
		Status: "revoked",
	}
	assert.True(t, revokedKey.IsRevoked())

	// Active key
	activeKey := &MonitoringAgentKey{
		Status: "active",
	}
	assert.False(t, activeKey.IsRevoked())

	// Other status
	otherKey := &MonitoringAgentKey{
		Status: "expired",
	}
	assert.False(t, otherKey.IsRevoked())
}

func TestMonitoringAgentKey_IsPublic(t *testing.T) {
	// Public agent
	publicKey := &MonitoringAgentKey{
		AgentType: "public",
	}
	assert.True(t, publicKey.IsPublic())

	// Private agent
	privateKey := &MonitoringAgentKey{
		AgentType: "private",
	}
	assert.False(t, privateKey.IsPublic())

	// Other type
	otherKey := &MonitoringAgentKey{
		AgentType: "hybrid",
	}
	assert.False(t, otherKey.IsPublic())
}

func TestMonitoringAgentKey_IsPrivate(t *testing.T) {
	// Private agent
	privateKey := &MonitoringAgentKey{
		AgentType: "private",
	}
	assert.True(t, privateKey.IsPrivate())

	// Public agent
	publicKey := &MonitoringAgentKey{
		AgentType: "public",
	}
	assert.False(t, publicKey.IsPrivate())

	// Other type
	otherKey := &MonitoringAgentKey{
		AgentType: "custom",
	}
	assert.False(t, otherKey.IsPrivate())
}

// Tests for helper request constructors

func TestNewPublicAgentKeyRequest(t *testing.T) {
	req := NewPublicAgentKeyRequest("Test Public Agent", "production", "us-west-1")

	assert.NotNil(t, req)
	assert.Equal(t, "Test Public Agent", req.Description)
	assert.Equal(t, "production", req.NamespaceName)
	assert.Equal(t, "public", req.AgentType)
	assert.Equal(t, "us-west-1", req.RegionCode)
	assert.Equal(t, []string{"public"}, req.AllowedProbeScopes)
}

func TestNewPublicAgentKeyRequest_EmptyValues(t *testing.T) {
	req := NewPublicAgentKeyRequest("", "", "")

	assert.NotNil(t, req)
	assert.Empty(t, req.Description)
	assert.Empty(t, req.NamespaceName)
	assert.Equal(t, "public", req.AgentType)
	assert.Empty(t, req.RegionCode)
	assert.Equal(t, []string{"public"}, req.AllowedProbeScopes)
}

func TestNewPrivateAgentKeyRequest(t *testing.T) {
	req := NewPrivateAgentKeyRequest("Test Private Agent", "staging", "eu-central-1")

	assert.NotNil(t, req)
	assert.Equal(t, "Test Private Agent", req.Description)
	assert.Equal(t, "staging", req.NamespaceName)
	assert.Equal(t, "private", req.AgentType)
	assert.Equal(t, "eu-central-1", req.RegionCode)
	assert.Equal(t, []string{"public", "private"}, req.AllowedProbeScopes)
}

func TestNewPrivateAgentKeyRequest_EmptyRegion(t *testing.T) {
	req := NewPrivateAgentKeyRequest("Private Agent", "development", "")

	assert.NotNil(t, req)
	assert.Equal(t, "Private Agent", req.Description)
	assert.Equal(t, "development", req.NamespaceName)
	assert.Equal(t, "private", req.AgentType)
	assert.Empty(t, req.RegionCode) // Region is optional for private agents
	assert.Equal(t, []string{"public", "private"}, req.AllowedProbeScopes)
}

func TestNewPrivateAgentKeyRequest_EmptyValues(t *testing.T) {
	req := NewPrivateAgentKeyRequest("", "", "")

	assert.NotNil(t, req)
	assert.Empty(t, req.Description)
	assert.Empty(t, req.NamespaceName)
	assert.Equal(t, "private", req.AgentType)
	assert.Empty(t, req.RegionCode)
	assert.Equal(t, []string{"public", "private"}, req.AllowedProbeScopes)
}

// Test ToQuery with various options

func TestListMonitoringAgentKeysOptions_ToQuery_AllFields(t *testing.T) {
	enabled := true
	clusterID := uint(10)

	opts := &ListMonitoringAgentKeysOptions{
		Page:      2,
		Limit:     50,
		Namespace: "production",
		Enabled:   &enabled,
		ClusterID: &clusterID,
	}

	query := opts.ToQuery()

	assert.Equal(t, "2", query["page"])
	assert.Equal(t, "50", query["limit"])
	assert.Equal(t, "production", query["namespace"])
	assert.Equal(t, "true", query["enabled"])
	assert.Equal(t, "10", query["cluster_id"])
}

func TestListMonitoringAgentKeysOptions_ToQuery_EnabledFalse(t *testing.T) {
	enabled := false

	opts := &ListMonitoringAgentKeysOptions{
		Enabled: &enabled,
	}

	query := opts.ToQuery()

	assert.Equal(t, "false", query["enabled"])
	_, hasPage := query["page"]
	assert.False(t, hasPage)
}

func TestListMonitoringAgentKeysOptions_ToQuery_EmptyOptions(t *testing.T) {
	opts := &ListMonitoringAgentKeysOptions{}

	query := opts.ToQuery()

	// Empty options should produce empty query (no zero values added)
	_, hasPage := query["page"]
	assert.False(t, hasPage)
	_, hasLimit := query["limit"]
	assert.False(t, hasLimit)
	_, hasNamespace := query["namespace"]
	assert.False(t, hasNamespace)
	_, hasEnabled := query["enabled"]
	assert.False(t, hasEnabled)
	_, hasClusterID := query["cluster_id"]
	assert.False(t, hasClusterID)
}

func TestListMonitoringAgentKeysOptions_ToQuery_PartialOptions(t *testing.T) {
	opts := &ListMonitoringAgentKeysOptions{
		Page:      1,
		Namespace: "staging",
	}

	query := opts.ToQuery()

	assert.Equal(t, "1", query["page"])
	assert.Equal(t, "staging", query["namespace"])
	_, hasLimit := query["limit"]
	assert.False(t, hasLimit)
	_, hasEnabled := query["enabled"]
	assert.False(t, hasEnabled)
}

// Integration test combining helper methods

func TestMonitoringAgentKey_Helpers_Integration(t *testing.T) {
	// Create a public active key
	publicKey := &MonitoringAgentKey{
		Status:    "active",
		AgentType: "public",
	}

	assert.True(t, publicKey.IsActive())
	assert.False(t, publicKey.IsRevoked())
	assert.True(t, publicKey.IsPublic())
	assert.False(t, publicKey.IsPrivate())

	// Create a private revoked key
	privateKey := &MonitoringAgentKey{
		Status:    "revoked",
		AgentType: "private",
	}

	assert.False(t, privateKey.IsActive())
	assert.True(t, privateKey.IsRevoked())
	assert.False(t, privateKey.IsPublic())
	assert.True(t, privateKey.IsPrivate())
}

func TestRequestConstructors_Integration(t *testing.T) {
	// Test public request constructor
	publicReq := NewPublicAgentKeyRequest(
		"Public Monitoring Agent",
		"production",
		"us-east-1",
	)

	assert.Equal(t, "public", publicReq.AgentType)
	assert.Equal(t, []string{"public"}, publicReq.AllowedProbeScopes)
	assert.NotEmpty(t, publicReq.RegionCode)

	// Test private request constructor
	privateReq := NewPrivateAgentKeyRequest(
		"Private Monitoring Agent",
		"production",
		"eu-west-1",
	)

	assert.Equal(t, "private", privateReq.AgentType)
	assert.Equal(t, []string{"public", "private"}, privateReq.AllowedProbeScopes)
	assert.NotEmpty(t, privateReq.RegionCode)

	// Verify they are different
	assert.NotEqual(t, publicReq.AgentType, privateReq.AgentType)
	assert.NotEqual(t, publicReq.AllowedProbeScopes, privateReq.AllowedProbeScopes)
}
