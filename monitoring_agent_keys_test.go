package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestMonitoringAgentKeysService_Create(t *testing.T) {
	setup()
	defer teardown()

	// Expected request
	expectedReq := &CreateMonitoringAgentKeyRequest{
		Description:        "Test private agent",
		NamespaceName:      "test-private-agent",
		AgentType:          "private",
		RegionCode:         "NYC3",
		AllowedProbeScopes: []string{"public", "private"},
	}

	// Mock response
	mockResp := &CreateMonitoringAgentKeyResponse{
		KeyID:              "mag_test123",
		SecretKey:          "secret123",
		FullToken:          "mag_test123.secret123",
		AgentType:          "private",
		AllowedProbeScopes: []string{"public", "private"},
		Key: &MonitoringAgentKey{
			KeyID:              "mag_test123",
			OrganizationID:     114,
			Description:        "Test private agent",
			NamespaceName:      "test-private-agent",
			AgentType:          "private",
			RegionCode:         "NYC3",
			AllowedProbeScopes: `["public","private"]`,
			Status:             "active",
		},
	}

	mux.HandleFunc("/v1/organizations/114/monitoring-agent-keys", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		var req CreateMonitoringAgentKeyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}

		// Verify organization ID is cleared for org endpoints
		if req.OrganizationID != 0 {
			t.Errorf("Expected OrganizationID to be 0, got %d", req.OrganizationID)
		}

		// Verify other fields
		if req.Description != expectedReq.Description {
			t.Errorf("Expected Description %s, got %s", expectedReq.Description, req.Description)
		}
		if req.AgentType != expectedReq.AgentType {
			t.Errorf("Expected AgentType %s, got %s", expectedReq.AgentType, req.AgentType)
		}

		writeJSON(w, StandardResponse{
			Success: true,
			Message: "Monitoring agent key created successfully",
			Data:    mockResp,
		})
	})

	resp, err := client.MonitoringAgentKeys.Create(context.Background(), "114", expectedReq)
	if err != nil {
		t.Fatal(err)
	}

	if resp.KeyID != mockResp.KeyID {
		t.Errorf("Expected KeyID %s, got %s", mockResp.KeyID, resp.KeyID)
	}
	if resp.FullToken != mockResp.FullToken {
		t.Errorf("Expected FullToken %s, got %s", mockResp.FullToken, resp.FullToken)
	}
	if resp.AgentType != mockResp.AgentType {
		t.Errorf("Expected AgentType %s, got %s", mockResp.AgentType, resp.AgentType)
	}
}

func TestMonitoringAgentKeysService_CreateAdmin(t *testing.T) {
	setup()
	defer teardown()

	// Expected request
	expectedReq := &CreateMonitoringAgentKeyRequest{
		OrganizationID:     114,
		Description:        "Admin test agent",
		NamespaceName:      "admin-test-agent",
		AgentType:          "public",
		RegionCode:         "NYC3",
		AllowedProbeScopes: []string{"public"},
	}

	mux.HandleFunc("/v1/admin/monitoring-agent-keys", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		var req CreateMonitoringAgentKeyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}

		// Verify organization ID is included for admin endpoints
		if req.OrganizationID != expectedReq.OrganizationID {
			t.Errorf("Expected OrganizationID %d, got %d", expectedReq.OrganizationID, req.OrganizationID)
		}

		writeJSON(w, StandardResponse{
			Success: true,
			Message: "Monitoring agent key created successfully",
			Data: &CreateMonitoringAgentKeyResponse{
				KeyID:              "mag_admin123",
				SecretKey:          "adminsecret123",
				FullToken:          "mag_admin123.adminsecret123",
				AgentType:          "public",
				AllowedProbeScopes: []string{"public"},
			},
		})
	})

	resp, err := client.MonitoringAgentKeys.CreateAdmin(context.Background(), expectedReq)
	if err != nil {
		t.Fatal(err)
	}

	if resp.KeyID != "mag_admin123" {
		t.Errorf("Expected KeyID mag_admin123, got %s", resp.KeyID)
	}
}

func TestNewPublicAgentKeyRequest(t *testing.T) {
	req := NewPublicAgentKeyRequest("Test Public Agent", "test-public", "NYC3")
	
	if req.AgentType != "public" {
		t.Errorf("Expected AgentType 'public', got %s", req.AgentType)
	}
	if req.RegionCode != "NYC3" {
		t.Errorf("Expected RegionCode 'NYC3', got %s", req.RegionCode)
	}
	if len(req.AllowedProbeScopes) != 1 || req.AllowedProbeScopes[0] != "public" {
		t.Errorf("Expected AllowedProbeScopes ['public'], got %v", req.AllowedProbeScopes)
	}
}

func TestNewPrivateAgentKeyRequest(t *testing.T) {
	req := NewPrivateAgentKeyRequest("Test Private Agent", "test-private", "NYC3")
	
	if req.AgentType != "private" {
		t.Errorf("Expected AgentType 'private', got %s", req.AgentType)
	}
	if len(req.AllowedProbeScopes) != 2 {
		t.Errorf("Expected 2 allowed probe scopes, got %d", len(req.AllowedProbeScopes))
	}
}

func TestMonitoringAgentKey_IsPublic(t *testing.T) {
	key := &MonitoringAgentKey{AgentType: "public"}
	if !key.IsPublic() {
		t.Error("Expected IsPublic() to return true for public agent")
	}
	if key.IsPrivate() {
		t.Error("Expected IsPrivate() to return false for public agent")
	}
}

func TestMonitoringAgentKey_IsPrivate(t *testing.T) {
	key := &MonitoringAgentKey{AgentType: "private"}
	if !key.IsPrivate() {
		t.Error("Expected IsPrivate() to return true for private agent")
	}
	if key.IsPublic() {
		t.Error("Expected IsPublic() to return false for private agent")
	}
}