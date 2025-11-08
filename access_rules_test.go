package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAccessRulesService_Create(t *testing.T) {
	mockRule := &AccessRule{
		ID:        1,
		RuleType:  AccessRuleTypeIP,
		Action:    AccessRuleActionWhitelist,
		Target:    "192.168.1.1",
		Duration:  AccessRuleDurationPermanent,
		Reason:    "Test whitelist rule",
		CreatedBy: 1,
		Active:    true,
		IsExpired: false,
		IsActive:  true,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/admin/access-rules" {
			t.Errorf("Expected path /v1/admin/access-rules, got %s", r.URL.Path)
		}

		resp := StandardResponse{
			Status:  "success",
			Message: "Access rule created successfully",
			Data:    mockRule,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		jsonBytes, _ := json.Marshal(resp)
		t.Logf("Mock server returning: %s", string(jsonBytes))
		w.Write(jsonBytes)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	req := &AccessRuleCreateRequest{
		RuleType: AccessRuleTypeIP,
		Action:   AccessRuleActionWhitelist,
		Target:   "192.168.1.1",
		Duration: AccessRuleDurationPermanent,
		Reason:   "Test whitelist rule",
	}

	rule, err := client.AccessRules.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	t.Logf("Returned rule: %+v", rule)
	if rule.ID != 1 {
		t.Errorf("Expected ID 1, got %d", rule.ID)
	}
	if rule.RuleType != AccessRuleTypeIP {
		t.Errorf("Expected rule_type 'ip', got %s", rule.RuleType)
	}
	if rule.Action != AccessRuleActionWhitelist {
		t.Errorf("Expected action 'whitelist', got %s", rule.Action)
	}
	if rule.Target != "192.168.1.1" {
		t.Errorf("Expected target '192.168.1.1', got %s", rule.Target)
	}
}

func TestAccessRulesService_List(t *testing.T) {
	mockRules := []*AccessRule{
		{
			ID:        1,
			RuleType:  AccessRuleTypeIP,
			Action:    AccessRuleActionWhitelist,
			Target:    "192.168.1.1",
			Duration:  AccessRuleDurationPermanent,
			Reason:    "Test rule",
			CreatedBy: 1,
			Active:    true,
			IsExpired: false,
			IsActive:  true,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/admin/access-rules" {
			t.Errorf("Expected path /v1/admin/access-rules, got %s", r.URL.Path)
		}

		// Check query parameters
		query := r.URL.Query()
		if query.Get("active") != "true" {
			t.Errorf("Expected active=true query parameter")
		}

		resp := ListAccessRulesResponse{
			Rules:      mockRules,
			TotalCount: 1,
			Limit:      50,
			Offset:     0,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	active := true
	opts := &ListAccessRulesOptions{
		Active: &active,
	}

	resp, err := client.AccessRules.List(context.Background(), opts)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if resp.TotalCount != 1 {
		t.Errorf("Expected total_count 1, got %d", resp.TotalCount)
	}
	if len(resp.Rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(resp.Rules))
	}
	if resp.Rules[0].ID != 1 {
		t.Errorf("Expected rule ID 1, got %d", resp.Rules[0].ID)
	}
}

func TestAccessRulesService_Get(t *testing.T) {
	mockRule := &AccessRule{
		ID:        1,
		RuleType:  AccessRuleTypeIP,
		Action:    AccessRuleActionBlacklist,
		Target:    "10.0.0.1",
		Duration:  AccessRuleDurationTemporary,
		Reason:    "Test blacklist rule",
		CreatedBy: 1,
		Active:    true,
		IsExpired: false,
		IsActive:  true,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/admin/access-rules/1" {
			t.Errorf("Expected path /v1/admin/access-rules/1, got %s", r.URL.Path)
		}

		resp := StandardResponse{
			Status:  "success",
			Message: "Access rule retrieved successfully",
			Data:    mockRule,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	rule, err := client.AccessRules.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if rule.ID != 1 {
		t.Errorf("Expected ID 1, got %d", rule.ID)
	}
	if rule.RuleType != AccessRuleTypeIP {
		t.Errorf("Expected rule_type 'ip', got %s", rule.RuleType)
	}
	if rule.Action != AccessRuleActionBlacklist {
		t.Errorf("Expected action 'blacklist', got %s", rule.Action)
	}
}

func TestAccessRulesService_Update(t *testing.T) {
	mockRule := &AccessRule{
		ID:        1,
		RuleType:  AccessRuleTypeIP,
		Action:    AccessRuleActionWhitelist,
		Target:    "192.168.1.1",
		Duration:  AccessRuleDurationPermanent,
		Reason:    "Updated reason",
		CreatedBy: 1,
		Active:    false,
		IsExpired: false,
		IsActive:  false,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/admin/access-rules/1" {
			t.Errorf("Expected path /v1/admin/access-rules/1, got %s", r.URL.Path)
		}

		resp := StandardResponse{
			Status:  "success",
			Message: "Access rule updated successfully",
			Data:    mockRule,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	active := false
	req := &AccessRuleUpdateRequest{
		Active: &active,
	}

	rule, err := client.AccessRules.Update(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if rule.Active != false {
		t.Errorf("Expected active=false, got %v", rule.Active)
	}
}

func TestAccessRulesService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/admin/access-rules/1" {
			t.Errorf("Expected path /v1/admin/access-rules/1, got %s", r.URL.Path)
		}

		resp := StandardResponse{
			Status:  "success",
			Message: "Access rule deleted successfully",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	err = client.AccessRules.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestAccessRulesService_QuickBlockIP(t *testing.T) {
	mockRule := &AccessRule{
		ID:        2,
		RuleType:  AccessRuleTypeIP,
		Action:    AccessRuleActionBlacklist,
		Target:    "10.0.0.100",
		Duration:  AccessRuleDurationTemporary,
		Reason:    "Quick block test",
		CreatedBy: 1,
		Active:    true,
		IsExpired: false,
		IsActive:  true,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/admin/access-rules/quick/block-ip" {
			t.Errorf("Expected path /v1/admin/access-rules/quick/block-ip, got %s", r.URL.Path)
		}

		resp := StandardResponse{
			Status:  "success",
			Message: "IP blocked successfully",
			Data:    mockRule,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	hours := 24
	req := &QuickBlockIPRequest{
		IP:       "10.0.0.100",
		Duration: AccessRuleDurationTemporary,
		Hours:    &hours,
		Reason:   "Quick block test",
	}

	rule, err := client.AccessRules.QuickBlockIP(context.Background(), req)
	if err != nil {
		t.Fatalf("QuickBlockIP failed: %v", err)
	}

	if rule.Target != "10.0.0.100" {
		t.Errorf("Expected target '10.0.0.100', got %s", rule.Target)
	}
	if rule.Action != AccessRuleActionBlacklist {
		t.Errorf("Expected action 'blacklist', got %s", rule.Action)
	}
}

func TestAccessRulesService_BulkOperation(t *testing.T) {
	mockResp := &BulkOperationResponse{
		Action:   "activate",
		RuleIDs:  []uint{1, 2, 3},
		Affected: 3,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/admin/access-rules/bulk" {
			t.Errorf("Expected path /v1/admin/access-rules/bulk, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResp)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	req := &AccessRuleBulkRequest{
		Action:  "activate",
		RuleIDs: []uint{1, 2, 3},
		Reason:  "Bulk activation test",
	}

	resp, err := client.AccessRules.BulkOperation(context.Background(), req)
	if err != nil {
		t.Fatalf("BulkOperation failed: %v", err)
	}

	if resp.Affected != 3 {
		t.Errorf("Expected affected=3, got %d", resp.Affected)
	}
	if resp.Action != "activate" {
		t.Errorf("Expected action='activate', got %s", resp.Action)
	}
}

func TestAccessRuleConstants(t *testing.T) {
	// Test AccessRuleType constants
	if AccessRuleTypeIP != "ip" {
		t.Errorf("Expected AccessRuleTypeIP='ip', got %s", AccessRuleTypeIP)
	}
	if AccessRuleTypeUser != "user" {
		t.Errorf("Expected AccessRuleTypeUser='user', got %s", AccessRuleTypeUser)
	}
	if AccessRuleTypeCIDR != "cidr" {
		t.Errorf("Expected AccessRuleTypeCIDR='cidr', got %s", AccessRuleTypeCIDR)
	}
	if AccessRuleTypeASN != "asn" {
		t.Errorf("Expected AccessRuleTypeASN='asn', got %s", AccessRuleTypeASN)
	}

	// Test AccessRuleAction constants
	if AccessRuleActionWhitelist != "whitelist" {
		t.Errorf("Expected AccessRuleActionWhitelist='whitelist', got %s", AccessRuleActionWhitelist)
	}
	if AccessRuleActionBlacklist != "blacklist" {
		t.Errorf("Expected AccessRuleActionBlacklist='blacklist', got %s", AccessRuleActionBlacklist)
	}

	// Test AccessRuleDuration constants
	if AccessRuleDurationTemporary != "temporary" {
		t.Errorf("Expected AccessRuleDurationTemporary='temporary', got %s", AccessRuleDurationTemporary)
	}
	if AccessRuleDurationPermanent != "permanent" {
		t.Errorf("Expected AccessRuleDurationPermanent='permanent', got %s", AccessRuleDurationPermanent)
	}
}
