package nexmonyx

import (
	"testing"
	"time"
)

func TestUnifiedAPIKey(t *testing.T) {
	// Test basic unified API key functionality
	key := &UnifiedAPIKey{
		KeyID:        "test-key-123",
		Name:         "Test Key",
		Description:  "Test unified API key",
		Type:         APIKeyTypeUser,
		Capabilities: []string{"servers:read", "metrics:read"},
		Status:       APIKeyStatusActive,
		OrganizationID: 1,
	}

	t.Run("IsActive", func(t *testing.T) {
		if !key.IsActive() {
			t.Error("Expected key to be active")
		}

		// Test with expired key
		expiredTime := CustomTime{Time: time.Now().Add(-1 * time.Hour)}
		key.ExpiresAt = &expiredTime
		if key.IsActive() {
			t.Error("Expected expired key to not be active")
		}

		// Reset for other tests
		key.ExpiresAt = nil
	})

	t.Run("HasCapability", func(t *testing.T) {
		if !key.HasCapability("servers:read") {
			t.Error("Expected key to have servers:read capability")
		}

		if key.HasCapability("servers:write") {
			t.Error("Expected key to not have servers:write capability")
		}

		// Test wildcard
		key.Capabilities = []string{"*"}
		if !key.HasCapability("anything") {
			t.Error("Expected key with wildcard to have any capability")
		}
	})

	t.Run("CanRegisterServers", func(t *testing.T) {
		// Test registration key
		regKey := &UnifiedAPIKey{
			Type:   APIKeyTypeRegistration,
			Status: APIKeyStatusActive,
		}
		if !regKey.CanRegisterServers() {
			t.Error("Expected registration key to be able to register servers")
		}

		// Test key with servers:register capability
		key.Type = APIKeyTypeUser
		key.Capabilities = []string{"servers:register"}
		if !key.CanRegisterServers() {
			t.Error("Expected key with servers:register capability to be able to register servers")
		}
	})

	t.Run("IsMonitoringAgent", func(t *testing.T) {
		monitoringKey := &UnifiedAPIKey{
			Type:   APIKeyTypeMonitoringAgent,
			Status: APIKeyStatusActive,
		}
		if !monitoringKey.IsMonitoringAgent() {
			t.Error("Expected monitoring agent key to be identified as monitoring agent")
		}

		publicAgentKey := &UnifiedAPIKey{
			Type:   APIKeyTypePublicAgent,
			Status: APIKeyStatusActive,
		}
		if !publicAgentKey.IsMonitoringAgent() {
			t.Error("Expected public agent key to be identified as monitoring agent")
		}
	})

	t.Run("GetAuthenticationMethod", func(t *testing.T) {
		// Test monitoring agent key (should use bearer)
		monitoringKey := &UnifiedAPIKey{
			Type: APIKeyTypeMonitoringAgent,
		}
		if monitoringKey.GetAuthenticationMethod() != "bearer" {
			t.Error("Expected monitoring agent key to use bearer authentication")
		}

		// Test registration key (should use headers)
		regKey := &UnifiedAPIKey{
			Type: APIKeyTypeRegistration,
		}
		if regKey.GetAuthenticationMethod() != "headers" {
			t.Error("Expected registration key to use headers authentication")
		}

		// Test user key (should use headers)
		userKey := &UnifiedAPIKey{
			Type: APIKeyTypeUser,
		}
		if userKey.GetAuthenticationMethod() != "headers" {
			t.Error("Expected user key to use headers authentication")
		}
	})
}

func TestAPIKeyHelpers(t *testing.T) {
	t.Run("NewUserAPIKey", func(t *testing.T) {
		req := NewUserAPIKey("Test User Key", "Test description", []string{"servers:read"})
		if req.Type != APIKeyTypeUser {
			t.Error("Expected user API key type")
		}
		if req.Name != "Test User Key" {
			t.Error("Expected correct name")
		}
		if len(req.Capabilities) != 1 || req.Capabilities[0] != "servers:read" {
			t.Error("Expected correct capabilities")
		}
	})

	t.Run("NewMonitoringAgentKey", func(t *testing.T) {
		req := NewMonitoringAgentKey("Agent Key", "Test agent", "test-ns", "private", "us-east-1", []string{"public", "private"})
		if req.Type != APIKeyTypeMonitoringAgent {
			t.Error("Expected monitoring agent key type")
		}
		if req.AgentType != "private" {
			t.Error("Expected private agent type")
		}
		if req.RegionCode != "us-east-1" {
			t.Error("Expected correct region code")
		}
		if len(req.AllowedProbeScopes) != 2 {
			t.Error("Expected correct allowed probe scopes")
		}
	})

	t.Run("NewRegistrationKey", func(t *testing.T) {
		req := NewRegistrationKey("Registration Key", "For server registration", 123)
		if req.Type != APIKeyTypeRegistration {
			t.Error("Expected registration key type")
		}
		if req.OrganizationID != 123 {
			t.Error("Expected correct organization ID")
		}
		if len(req.Capabilities) != 2 {
			t.Error("Expected servers:register and servers:update capabilities")
		}
	})
}

func TestAPIKeyTypes(t *testing.T) {
	// Test all API key types are defined
	types := []APIKeyType{
		APIKeyTypeUser,
		APIKeyTypeAdmin,
		APIKeyTypeMonitoringAgent,
		APIKeyTypeSystem,
		APIKeyTypePublicAgent,
		APIKeyTypeRegistration,
		APIKeyTypeOrgMonitoring,
	}

	for _, keyType := range types {
		if string(keyType) == "" {
			t.Errorf("API key type %v should not be empty", keyType)
		}
	}
}

func TestAPIKeyStatuses(t *testing.T) {
	// Test all API key statuses are defined
	statuses := []APIKeyStatus{
		APIKeyStatusActive,
		APIKeyStatusRevoked,
		APIKeyStatusExpired,
		APIKeyStatusPending,
	}

	for _, status := range statuses {
		if string(status) == "" {
			t.Errorf("API key status %v should not be empty", status)
		}
	}
}

func TestCapabilityConstants(t *testing.T) {
	// Test that capability constants are defined
	capabilities := []string{
		CapabilityServersRead,
		CapabilityServersWrite,
		CapabilityServersRegister,
		CapabilityServersDelete,
		CapabilityServersAll,
		CapabilityMonitoringRead,
		CapabilityMonitoringWrite,
		CapabilityMonitoringExecute,
		CapabilityMonitoringAll,
		CapabilityProbesRead,
		CapabilityProbesWrite,
		CapabilityProbesExecute,
		CapabilityProbesAll,
		CapabilityMetricsRead,
		CapabilityMetricsWrite,
		CapabilityMetricsSubmit,
		CapabilityMetricsAll,
		CapabilityOrganizationRead,
		CapabilityOrganizationWrite,
		CapabilityOrganizationAll,
		CapabilityAdminRead,
		CapabilityAdminWrite,
		CapabilityAdminAll,
		CapabilityAll,
	}

	for _, capability := range capabilities {
		if capability == "" {
			t.Error("Capability constant should not be empty")
		}
	}
}

func TestBackwardCompatibility(t *testing.T) {
	// Test that APIKey type alias works
	var key APIKey = UnifiedAPIKey{
		KeyID:  "test-123",
		Name:   "Test",
		Status: APIKeyStatusActive,
		Type:   APIKeyTypeUser,
	}

	if key.KeyID != "test-123" {
		t.Error("APIKey type alias should work")
	}

	if !key.IsActive() {
		t.Error("APIKey methods should work through type alias")
	}
}