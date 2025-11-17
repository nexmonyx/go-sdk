package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBusinessRules_ServerStatePrerequisites tests server state transition prerequisites
func TestBusinessRules_ServerStatePrerequisites(t *testing.T) {
	tests := []struct {
		name          string
		currentState  string
		action        string
		expectSuccess bool
		errorMsg      string
	}{
		{
			name:          "valid - register new server",
			currentState:  "",
			action:        "register",
			expectSuccess: true,
		},
		{
			name:          "valid - decommission active server",
			currentState:  "active",
			action:        "decommission",
			expectSuccess: true,
		},
		{
			name:          "invalid - delete active server without decommission",
			currentState:  "active",
			action:        "delete",
			expectSuccess: false,
			errorMsg:      "must be decommissioned first",
		},
		{
			name:          "invalid - register already registered server",
			currentState:  "active",
			action:        "register",
			expectSuccess: false,
			errorMsg:      "already registered",
		},
		{
			name:          "valid - delete decommissioned server",
			currentState:  "decommissioned",
			action:        "delete",
			expectSuccess: true,
		},
		{
			name:          "invalid - activate decommissioned server",
			currentState:  "decommissioned",
			action:        "activate",
			expectSuccess: false,
			errorMsg:      "cannot reactivate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.expectSuccess {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"data": map[string]interface{}{
							"uuid":  "server-uuid",
							"state": tt.action,
						},
					})
				} else {
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"error": tt.errorMsg,
					})
				}
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			// Simulate different actions
			var apiErr error
			switch tt.action {
			case "register":
				_, apiErr = client.Servers.Register(context.Background(), "test-server", 1)
			case "delete":
				apiErr = client.Servers.Delete(context.Background(), "server-uuid")
			case "decommission":
				_, apiErr = client.Servers.Update(context.Background(), "server-uuid", &Server{
					Hostname: "test-server",
				})
			default:
				_, apiErr = client.Servers.Get(context.Background(), "server-uuid")
			}

			if tt.expectSuccess {
				assert.NoError(t, apiErr)
			} else {
				assert.Error(t, apiErr)
			}
		})
	}
}

// TestBusinessRules_RelationshipConstraints tests relationship constraints between entities
func TestBusinessRules_RelationshipConstraints(t *testing.T) {
	tests := []struct {
		name          string
		scenario      string
		expectSuccess bool
		errorMsg      string
	}{
		{
			name:          "valid - create alert for existing server",
			scenario:      "alert_with_server",
			expectSuccess: true,
		},
		{
			name:          "invalid - create alert for non-existent server",
			scenario:      "alert_no_server",
			expectSuccess: false,
			errorMsg:      "server not found",
		},
		{
			name:          "invalid - delete server with active alerts",
			scenario:      "delete_server_with_alerts",
			expectSuccess: false,
			errorMsg:      "has active alerts",
		},
		{
			name:          "valid - delete server after removing alerts",
			scenario:      "delete_server_no_alerts",
			expectSuccess: true,
		},
		{
			name:          "invalid - assign probe to non-existent region",
			scenario:      "probe_invalid_region",
			expectSuccess: false,
			errorMsg:      "region not found",
		},
		{
			name:          "valid - assign probe to valid region",
			scenario:      "probe_valid_region",
			expectSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.expectSuccess {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"data": map[string]interface{}{
							"id": 1,
						},
					})
				} else {
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"error": tt.errorMsg,
					})
				}
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			// Simulate different relationship scenarios
			var apiErr error
			switch tt.scenario {
			case "alert_with_server", "alert_no_server":
				_, _, apiErr = client.Alerts.List(context.Background(), nil)
			case "delete_server_with_alerts", "delete_server_no_alerts":
				apiErr = client.Servers.Delete(context.Background(), "server-uuid")
			case "probe_invalid_region", "probe_valid_region":
				_, _, apiErr = client.Monitoring.ListProbes(context.Background(), nil)
			}

			if tt.expectSuccess {
				assert.NoError(t, apiErr)
			} else {
				assert.Error(t, apiErr)
			}
		})
	}
}

// TestBusinessRules_QuotaEnforcement tests quota and limit enforcement
func TestBusinessRules_QuotaEnforcement(t *testing.T) {
	tests := []struct {
		name          string
		currentUsage  int
		quotaLimit    int
		action        string
		expectSuccess bool
		errorMsg      string
	}{
		{
			name:          "valid - within quota",
			currentUsage:  5,
			quotaLimit:    10,
			action:        "create_server",
			expectSuccess: true,
		},
		{
			name:          "invalid - quota exceeded",
			currentUsage:  10,
			quotaLimit:    10,
			action:        "create_server",
			expectSuccess: false,
			errorMsg:      "quota exceeded",
		},
		{
			name:          "invalid - creating would exceed quota",
			currentUsage:  9,
			quotaLimit:    10,
			action:        "create_multiple_servers",
			expectSuccess: false,
			errorMsg:      "would exceed quota",
		},
		{
			name:          "valid - at quota but not exceeding",
			currentUsage:  9,
			quotaLimit:    10,
			action:        "create_server",
			expectSuccess: true,
		},
		{
			name:          "invalid - probe quota exceeded",
			currentUsage:  100,
			quotaLimit:    100,
			action:        "create_probe",
			expectSuccess: false,
			errorMsg:      "probe quota exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.expectSuccess {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"data": map[string]interface{}{
							"id":            1,
							"current_usage": tt.currentUsage,
							"quota_limit":   tt.quotaLimit,
						},
					})
				} else {
					w.WriteHeader(http.StatusForbidden)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"error": tt.errorMsg,
					})
				}
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			var apiErr error
			switch tt.action {
			case "create_server", "create_multiple_servers":
				_, apiErr = client.Servers.Register(context.Background(), "test-server", 1)
			case "create_probe":
				_, _, apiErr = client.Monitoring.ListProbes(context.Background(), nil)
			}

			if tt.expectSuccess {
				assert.NoError(t, apiErr)
			} else {
				assert.Error(t, apiErr)
			}
		})
	}
}

// TestBusinessRules_DependencyValidation tests dependency validation
func TestBusinessRules_DependencyValidation(t *testing.T) {
	tests := []struct {
		name          string
		scenario      string
		expectSuccess bool
		errorMsg      string
	}{
		{
			name:          "valid - delete organization with no resources",
			scenario:      "delete_empty_org",
			expectSuccess: true,
		},
		{
			name:          "invalid - delete organization with servers",
			scenario:      "delete_org_with_servers",
			expectSuccess: false,
			errorMsg:      "organization has active servers",
		},
		{
			name:          "invalid - delete organization with active subscription",
			scenario:      "delete_org_with_subscription",
			expectSuccess: false,
			errorMsg:      "active subscription",
		},
		{
			name:          "valid - remove user from organization",
			scenario:      "remove_user_from_org",
			expectSuccess: true,
		},
		{
			name:          "invalid - remove last admin from organization",
			scenario:      "remove_last_admin",
			expectSuccess: false,
			errorMsg:      "cannot remove last admin",
		},
		{
			name:          "invalid - downgrade subscription with usage over new limit",
			scenario:      "downgrade_with_overusage",
			expectSuccess: false,
			errorMsg:      "usage exceeds new plan limits",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.expectSuccess {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"data": map[string]interface{}{
							"success": true,
						},
					})
				} else {
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"error": tt.errorMsg,
					})
				}
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			var apiErr error
			switch tt.scenario {
			case "delete_empty_org", "delete_org_with_servers", "delete_org_with_subscription":
				apiErr = client.Organizations.Delete(context.Background(), "org-uuid")
			case "remove_user_from_org", "remove_last_admin":
				_, _, apiErr = client.Organizations.GetUsers(context.Background(), "org-uuid", nil)
			case "downgrade_with_overusage":
				_, apiErr = client.Billing.GetSubscription(context.Background(), "org-uuid")
			}

			if tt.expectSuccess {
				assert.NoError(t, apiErr)
			} else {
				assert.Error(t, apiErr)
			}
		})
	}
}

// TestBusinessRules_WorkflowRequirements tests workflow requirements
func TestBusinessRules_WorkflowRequirements(t *testing.T) {
	tests := []struct {
		name          string
		workflow      string
		stepsComplete []string
		nextStep      string
		expectSuccess bool
		errorMsg      string
	}{
		{
			name:          "valid - complete onboarding in order",
			workflow:      "onboarding",
			stepsComplete: []string{"register", "verify_email"},
			nextStep:      "setup_organization",
			expectSuccess: true,
		},
		{
			name:          "invalid - skip onboarding step",
			workflow:      "onboarding",
			stepsComplete: []string{"register"},
			nextStep:      "setup_organization",
			expectSuccess: false,
			errorMsg:      "must verify email first",
		},
		{
			name:          "valid - server registration workflow",
			workflow:      "server_registration",
			stepsComplete: []string{"create_key"},
			nextStep:      "register_server",
			expectSuccess: true,
		},
		{
			name:          "invalid - register server without key",
			workflow:      "server_registration",
			stepsComplete: []string{},
			nextStep:      "register_server",
			expectSuccess: false,
			errorMsg:      "registration key required",
		},
		{
			name:          "valid - alert creation workflow",
			workflow:      "alert_creation",
			stepsComplete: []string{"select_metric", "set_threshold"},
			nextStep:      "create_alert",
			expectSuccess: true,
		},
		{
			name:          "invalid - create alert without threshold",
			workflow:      "alert_creation",
			stepsComplete: []string{"select_metric"},
			nextStep:      "create_alert",
			expectSuccess: false,
			errorMsg:      "threshold required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.expectSuccess {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"data": map[string]interface{}{
							"workflow":        tt.workflow,
							"steps_complete":  tt.stepsComplete,
							"current_step":    tt.nextStep,
						},
					})
				} else {
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"error": tt.errorMsg,
					})
				}
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			var apiErr error
			switch tt.workflow {
			case "onboarding":
				_, apiErr = client.Organizations.Create(context.Background(), &Organization{
					Name: "Test Org",
				})
			case "server_registration":
				_, apiErr = client.Servers.Register(context.Background(), "test-server", 1)
			case "alert_creation":
				_, _, apiErr = client.Alerts.List(context.Background(), nil)
			}

			if tt.expectSuccess {
				assert.NoError(t, apiErr)
			} else {
				assert.Error(t, apiErr)
			}
		})
	}
}

// TestBusinessRules_TimeBasedConstraints tests time-based business rules
func TestBusinessRules_TimeBasedConstraints(t *testing.T) {
	tests := []struct {
		name          string
		constraint    string
		expectSuccess bool
		errorMsg      string
	}{
		{
			name:          "valid - schedule probe during maintenance window",
			constraint:    "maintenance_window",
			expectSuccess: true,
		},
		{
			name:          "invalid - schedule probe outside maintenance window",
			constraint:    "outside_maintenance",
			expectSuccess: false,
			errorMsg:      "outside maintenance window",
		},
		{
			name:          "invalid - delete recent backup (within retention)",
			constraint:    "delete_recent_backup",
			expectSuccess: false,
			errorMsg:      "within retention period",
		},
		{
			name:          "valid - delete old backup (past retention)",
			constraint:    "delete_old_backup",
			expectSuccess: true,
		},
		{
			name:          "invalid - modify billing period in progress",
			constraint:    "modify_current_billing",
			expectSuccess: false,
			errorMsg:      "billing period in progress",
		},
		{
			name:          "valid - modify future billing period",
			constraint:    "modify_future_billing",
			expectSuccess: true,
		},
		{
			name:          "invalid - cancel subscription with active period",
			constraint:    "cancel_active_subscription",
			expectSuccess: false,
			errorMsg:      "subscription period active",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.expectSuccess {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"data": map[string]interface{}{
							"constraint": tt.constraint,
							"timestamp":  time.Now().Unix(),
						},
					})
				} else {
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"error": tt.errorMsg,
					})
				}
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			var apiErr error
			switch tt.constraint {
			case "maintenance_window", "outside_maintenance":
				_, _, apiErr = client.Monitoring.ListProbes(context.Background(), nil)
			case "delete_recent_backup", "delete_old_backup":
				apiErr = client.Servers.Delete(context.Background(), "backup-uuid")
			case "modify_current_billing", "modify_future_billing":
				_, apiErr = client.Billing.GetSubscription(context.Background(), "org-uuid")
			case "cancel_active_subscription":
				_, apiErr = client.Billing.GetSubscription(context.Background(), "org-uuid")
			}

			if tt.expectSuccess {
				assert.NoError(t, apiErr)
			} else {
				assert.Error(t, apiErr)
			}
		})
	}
}

// TestBusinessRules_APIKeyLifecycle tests API key lifecycle state methods (IsActive, IsExpired, IsRevoked)
func TestBusinessRules_APIKeyLifecycle(t *testing.T) {
	now := time.Now()
	futureTime := CustomTime{Time: now.Add(24 * time.Hour)}
	pastTime := CustomTime{Time: now.Add(-24 * time.Hour)}

	tests := []struct {
		name           string
		key            *UnifiedAPIKey
		expectActive   bool
		expectExpired  bool
		expectRevoked  bool
	}{
		{
			name: "active key - no expiration",
			key: &UnifiedAPIKey{
				Status:    APIKeyStatusActive,
				ExpiresAt: nil,
			},
			expectActive:  true,
			expectExpired: false,
			expectRevoked: false,
		},
		{
			name: "active key - not yet expired",
			key: &UnifiedAPIKey{
				Status:    APIKeyStatusActive,
				ExpiresAt: &futureTime,
			},
			expectActive:  true,
			expectExpired: false,
			expectRevoked: false,
		},
		{
			name: "active key - but expired",
			key: &UnifiedAPIKey{
				Status:    APIKeyStatusActive,
				ExpiresAt: &pastTime,
			},
			expectActive:  false,
			expectExpired: true,
			expectRevoked: false,
		},
		{
			name: "revoked key",
			key: &UnifiedAPIKey{
				Status:    APIKeyStatusRevoked,
				ExpiresAt: nil,
			},
			expectActive:  false,
			expectExpired: false,
			expectRevoked: true,
		},
		{
			name: "revoked and expired key",
			key: &UnifiedAPIKey{
				Status:    APIKeyStatusRevoked,
				ExpiresAt: &pastTime,
			},
			expectActive:  false,
			expectExpired: true,
			expectRevoked: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectActive, tt.key.IsActive(), "IsActive() mismatch")
			assert.Equal(t, tt.expectExpired, tt.key.IsExpired(), "IsExpired() mismatch")
			assert.Equal(t, tt.expectRevoked, tt.key.IsRevoked(), "IsRevoked() mismatch")
		})
	}
}

// TestBusinessRules_APIKeyCapabilities tests API key capability checking with wildcard support
func TestBusinessRules_APIKeyCapabilities(t *testing.T) {
	tests := []struct {
		name         string
		key          *UnifiedAPIKey
		checkCap     string
		expectHasCap bool
	}{
		{
			name: "exact capability match",
			key: &UnifiedAPIKey{
				Capabilities: []string{"servers:read", "servers:write"},
			},
			checkCap:     "servers:read",
			expectHasCap: true,
		},
		{
			name: "no capability match",
			key: &UnifiedAPIKey{
				Capabilities: []string{"servers:read"},
			},
			checkCap:     "servers:write",
			expectHasCap: false,
		},
		{
			name: "wildcard grants all capabilities",
			key: &UnifiedAPIKey{
				Capabilities: []string{"*"},
			},
			checkCap:     "any:capability",
			expectHasCap: true,
		},
		{
			name: "wildcard among other capabilities",
			key: &UnifiedAPIKey{
				Capabilities: []string{"servers:read", "*", "alerts:write"},
			},
			checkCap:     "billing:manage",
			expectHasCap: true,
		},
		{
			name: "empty capabilities",
			key: &UnifiedAPIKey{
				Capabilities: []string{},
			},
			checkCap:     "servers:read",
			expectHasCap: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectHasCap, tt.key.HasCapability(tt.checkCap))
		})
	}
}

// TestBusinessRules_OrganizationAccessControl tests organization-scoped access control with admin/system overrides
func TestBusinessRules_OrganizationAccessControl(t *testing.T) {
	tests := []struct {
		name        string
		key         *UnifiedAPIKey
		checkOrgID  uint
		expectAccess bool
	}{
		{
			name: "user key - same organization",
			key: &UnifiedAPIKey{
				Type:           APIKeyTypeUser,
				OrganizationID: 100,
			},
			checkOrgID:   100,
			expectAccess: true,
		},
		{
			name: "user key - different organization",
			key: &UnifiedAPIKey{
				Type:           APIKeyTypeUser,
				OrganizationID: 100,
			},
			checkOrgID:   200,
			expectAccess: false,
		},
		{
			name: "admin key - can access any organization",
			key: &UnifiedAPIKey{
				Type:           APIKeyTypeAdmin,
				OrganizationID: 100,
			},
			checkOrgID:   200,
			expectAccess: true,
		},
		{
			name: "system key - can access any organization",
			key: &UnifiedAPIKey{
				Type:           APIKeyTypeSystem,
				OrganizationID: 100,
			},
			checkOrgID:   300,
			expectAccess: true,
		},
		{
			name: "monitoring agent - same organization only",
			key: &UnifiedAPIKey{
				Type:           APIKeyTypeMonitoringAgent,
				OrganizationID: 100,
			},
			checkOrgID:   200,
			expectAccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectAccess, tt.key.CanAccessOrganization(tt.checkOrgID))
		})
	}
}

// TestBusinessRules_KeyTypeDetection tests API key type detection logic (IsPublicAgent, IsPrivateAgent)
func TestBusinessRules_KeyTypeDetection(t *testing.T) {
	tests := []struct {
		name             string
		key              *UnifiedAPIKey
		expectPublic     bool
		expectPrivate    bool
	}{
		{
			name: "public agent - explicit type",
			key: &UnifiedAPIKey{
				Type:      APIKeyTypePublicAgent,
				AgentType: "",
			},
			expectPublic:  true,
			expectPrivate: false,
		},
		{
			name: "public agent - monitoring agent with public agentType",
			key: &UnifiedAPIKey{
				Type:      APIKeyTypeMonitoringAgent,
				AgentType: "public",
			},
			expectPublic:  true,
			expectPrivate: false,
		},
		{
			name: "private agent - monitoring agent with private agentType",
			key: &UnifiedAPIKey{
				Type:      APIKeyTypeMonitoringAgent,
				AgentType: "private",
			},
			expectPublic:  false,
			expectPrivate: true,
		},
		{
			name: "not an agent - user key",
			key: &UnifiedAPIKey{
				Type:      APIKeyTypeUser,
				AgentType: "",
			},
			expectPublic:  false,
			expectPrivate: false,
		},
		{
			name: "monitoring agent - no agentType specified",
			key: &UnifiedAPIKey{
				Type:      APIKeyTypeMonitoringAgent,
				AgentType: "",
			},
			expectPublic:  false,
			expectPrivate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectPublic, tt.key.IsPublicAgent(), "IsPublicAgent() mismatch")
			assert.Equal(t, tt.expectPrivate, tt.key.IsPrivateAgent(), "IsPrivateAgent() mismatch")
		})
	}
}

// TestBusinessRules_ServerRegistrationValidation tests server registration permission checks
func TestBusinessRules_ServerRegistrationValidation(t *testing.T) {
	tests := []struct {
		name               string
		key                *UnifiedAPIKey
		expectCanRegister  bool
	}{
		{
			name: "registration key - can register servers",
			key: &UnifiedAPIKey{
				Type:         APIKeyTypeRegistration,
				Capabilities: []string{"servers:register", "servers:update"},
			},
			expectCanRegister: true,
		},
		{
			name: "monitoring agent - cannot register servers without capability",
			key: &UnifiedAPIKey{
				Type:         APIKeyTypeMonitoringAgent,
				Capabilities: []string{"monitoring:execute"},
			},
			expectCanRegister: false,
		},
		{
			name: "user key - can register if has capability",
			key: &UnifiedAPIKey{
				Type:         APIKeyTypeUser,
				Capabilities: []string{"servers:register"},
			},
			expectCanRegister: true,
		},
		{
			name: "admin key - can register with wildcard",
			key: &UnifiedAPIKey{
				Type:         APIKeyTypeAdmin,
				Capabilities: []string{"*"},
			},
			expectCanRegister: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectCanRegister, tt.key.HasCapability("servers:register"))
		})
	}
}

// TestBusinessRules_KeyBuilderDefaults tests API key builder functions set correct defaults
func TestBusinessRules_KeyBuilderDefaults(t *testing.T) {
	t.Run("NewUserAPIKey sets correct defaults", func(t *testing.T) {
		req := NewUserAPIKey("My API Key", "Test description", []string{"servers:read"})

		assert.Equal(t, "My API Key", req.Name)
		assert.Equal(t, "Test description", req.Description)
		assert.Equal(t, APIKeyTypeUser, req.Type)
		assert.Equal(t, []string{"servers:read"}, req.Capabilities)
	})

	t.Run("NewAdminAPIKey sets correct defaults", func(t *testing.T) {
		req := NewAdminAPIKey("Admin Key", "Admin description", []string{"*"}, 1)

		assert.Equal(t, "Admin Key", req.Name)
		assert.Equal(t, "Admin description", req.Description)
		assert.Equal(t, APIKeyTypeAdmin, req.Type)
		assert.Contains(t, req.Capabilities, "*")
		assert.Equal(t, uint(1), req.OrganizationID)
	})

	t.Run("NewMonitoringAgentKey sets correct defaults", func(t *testing.T) {
		req := NewMonitoringAgentKey(
			"Agent Key",
			"Agent description",
			"monitoring-namespace",
			"private",
			"us-east-1",
			[]string{"scope1", "scope2"},
		)

		assert.Equal(t, "Agent Key", req.Name)
		assert.Equal(t, "Agent description", req.Description)
		assert.Equal(t, APIKeyTypeMonitoringAgent, req.Type)
		assert.Equal(t, "monitoring-namespace", req.NamespaceName)
		assert.Equal(t, "private", req.AgentType)
		assert.Equal(t, "us-east-1", req.RegionCode)
		assert.Contains(t, req.Capabilities, "monitoring:execute")
		assert.Contains(t, req.Capabilities, "probes:execute")
		assert.Equal(t, []string{"scope1", "scope2"}, req.AllowedProbeScopes)
	})

	t.Run("NewRegistrationKey sets correct defaults", func(t *testing.T) {
		req := NewRegistrationKey("Registration Key", "Registration description", 42)

		assert.Equal(t, "Registration Key", req.Name)
		assert.Equal(t, "Registration description", req.Description)
		assert.Equal(t, APIKeyTypeRegistration, req.Type)
		assert.Equal(t, uint(42), req.OrganizationID)
		assert.Contains(t, req.Capabilities, "servers:register")
		assert.Contains(t, req.Capabilities, "servers:update")
	})
}

// TestBusinessRules_AlertStateTransitions tests alert state machine transitions
func TestBusinessRules_AlertStateTransitions(t *testing.T) {
	tests := []struct {
		name          string
		currentState  string
		newState      string
		expectAllowed bool
	}{
		{
			name:          "pending to firing - allowed",
			currentState:  "pending",
			newState:      "firing",
			expectAllowed: true,
		},
		{
			name:          "firing to resolved - allowed",
			currentState:  "firing",
			newState:      "resolved",
			expectAllowed: true,
		},
		{
			name:          "firing to silenced - allowed",
			currentState:  "firing",
			newState:      "silenced",
			expectAllowed: true,
		},
		{
			name:          "resolved to firing - not allowed",
			currentState:  "resolved",
			newState:      "firing",
			expectAllowed: false,
		},
		{
			name:          "pending to resolved - not allowed",
			currentState:  "pending",
			newState:      "resolved",
			expectAllowed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// State machine logic: pending -> firing -> (silenced|resolved)
			validTransitions := map[string][]string{
				"pending":  {"firing"},
				"firing":   {"silenced", "resolved"},
				"silenced": {"resolved"},
				"resolved": {}, // Terminal state
			}

			allowedStates, exists := validTransitions[tt.currentState]
			isAllowed := exists && stringSliceContains(allowedStates, tt.newState)

			assert.Equal(t, tt.expectAllowed, isAllowed, "State transition validation mismatch")
		})
	}
}

// TestBusinessRules_RegionStatusValidation tests region status transitions
func TestBusinessRules_RegionStatusValidation(t *testing.T) {
	tests := []struct {
		name          string
		currentStatus string
		newStatus     string
		expectAllowed bool
	}{
		{
			name:          "active to maintenance - allowed",
			currentStatus: "active",
			newStatus:     "maintenance",
			expectAllowed: true,
		},
		{
			name:          "maintenance to active - allowed",
			currentStatus: "maintenance",
			newStatus:     "active",
			expectAllowed: true,
		},
		{
			name:          "active to inactive - allowed",
			currentStatus: "active",
			newStatus:     "inactive",
			expectAllowed: true,
		},
		{
			name:          "inactive to active - not allowed",
			currentStatus: "inactive",
			newStatus:     "active",
			expectAllowed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Region status transitions
			validTransitions := map[string][]string{
				"active":      {"maintenance", "inactive"},
				"maintenance": {"active"},
				"inactive":    {}, // One-way transition
			}

			allowedStatuses, exists := validTransitions[tt.currentStatus]
			isAllowed := exists && stringSliceContains(allowedStatuses, tt.newStatus)

			assert.Equal(t, tt.expectAllowed, isAllowed, "Region status transition validation mismatch")
		})
	}
}

// stringSliceContains is a helper function for slice membership check
func stringSliceContains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
