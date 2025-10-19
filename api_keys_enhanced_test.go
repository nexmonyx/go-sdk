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

// TestAPIKeysService_CapabilityValidation tests capability validation and scope checking
func TestAPIKeysService_CapabilityValidation(t *testing.T) {
	tests := []struct {
		name       string
		keyID      string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *UnifiedAPIKey)
	}{
		{
			name:       "success - wildcard capabilities",
			keyID:      "key-wildcard",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":              1,
					"key_id":          "key-wildcard",
					"name":            "Wildcard Key",
					"type":            "user",
					"status":          "active",
					"capabilities":    []string{"*"},
					"organization_id": 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, key *UnifiedAPIKey) {
				assert.True(t, key.HasCapability("anything"))
				assert.True(t, key.HasCapability("servers:read"))
				assert.True(t, key.HasCapability("admin:write"))
			},
		},
		{
			name:       "success - multiple specific capabilities",
			keyID:      "key-multi",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":              2,
					"key_id":          "key-multi",
					"name":            "Multi-Capability Key",
					"type":            "user",
					"status":          "active",
					"capabilities":    []string{"servers:read", "servers:write", "metrics:submit", "probes:execute"},
					"organization_id": 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, key *UnifiedAPIKey) {
				assert.True(t, key.HasCapability("servers:read"))
				assert.True(t, key.HasCapability("servers:write"))
				assert.True(t, key.HasCapability("metrics:submit"))
				assert.False(t, key.HasCapability("admin:write"))
			},
		},
		{
			name:       "success - registration key capabilities",
			keyID:      "key-reg",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":              3,
					"key_id":          "key-reg",
					"name":            "Registration Key",
					"type":            "registration",
					"status":          "active",
					"capabilities":    []string{"servers:register", "servers:update"},
					"organization_id": 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, key *UnifiedAPIKey) {
				assert.True(t, key.CanRegisterServers())
				assert.True(t, key.HasCapability("servers:register"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL:    server.URL,
				Auth:       AuthConfig{Token: "test-token"},
				RetryCount: 0,
			})
			require.NoError(t, err)

			key, err := client.APIKeys.GetUnified(context.Background(), tt.keyID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, key)
				if tt.checkFunc != nil {
					tt.checkFunc(t, key)
				}
			}
		})
	}
}

// TestAPIKeysService_ExpirationHandling tests key expiration logic and edge cases
func TestAPIKeysService_ExpirationHandling(t *testing.T) {
	tests := []struct {
		name       string
		keyID      string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *UnifiedAPIKey)
	}{
		{
			name:       "success - active key with no expiration",
			keyID:      "key-noexp",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":              1,
					"key_id":          "key-noexp",
					"name":            "Persistent Key",
					"type":            "user",
					"status":          "active",
					"expires_at":      nil,
					"organization_id": 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, key *UnifiedAPIKey) {
				assert.True(t, key.IsActive())
				assert.Nil(t, key.ExpiresAt)
			},
		},
		{
			name:       "success - active key expiring in future",
			keyID:      "key-future",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":              2,
					"key_id":          "key-future",
					"name":            "Future Expiration Key",
					"type":            "user",
					"status":          "active",
					"expires_at":      "2099-12-31T23:59:59Z",
					"organization_id": 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, key *UnifiedAPIKey) {
				assert.True(t, key.IsActive())
				assert.NotNil(t, key.ExpiresAt)
			},
		},
		{
			name:       "success - expired key marked inactive",
			keyID:      "key-expired",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":              3,
					"key_id":          "key-expired",
					"name":            "Expired Key",
					"type":            "user",
					"status":          "expired",
					"expires_at":      "2020-01-01T00:00:00Z",
					"organization_id": 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, key *UnifiedAPIKey) {
				assert.False(t, key.IsActive())
			},
		},
		{
			name:       "success - revoked key",
			keyID:      "key-revoked",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":              4,
					"key_id":          "key-revoked",
					"name":            "Revoked Key",
					"type":            "user",
					"status":          "revoked",
					"revoked_at":      "2025-10-15T10:00:00Z",
					"organization_id": 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, key *UnifiedAPIKey) {
				assert.False(t, key.IsActive())
				assert.Equal(t, APIKeyStatusRevoked, key.Status)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL:    server.URL,
				Auth:       AuthConfig{Token: "test-token"},
				RetryCount: 0,
			})
			require.NoError(t, err)

			key, err := client.APIKeys.GetUnified(context.Background(), tt.keyID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, key)
				if tt.checkFunc != nil {
					tt.checkFunc(t, key)
				}
			}
		})
	}
}

// TestAPIKeysService_KeyTypeSpecificBehavior tests behavior specific to each key type
func TestAPIKeysService_KeyTypeSpecificBehavior(t *testing.T) {
	tests := []struct {
		name       string
		keyID      string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *UnifiedAPIKey)
	}{
		{
			name:       "success - monitoring agent key properties",
			keyID:      "key-agent",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":                    1,
					"key_id":                "key-agent",
					"name":                  "Monitoring Agent",
					"type":                  "monitoring_agent",
					"status":                "active",
					"namespace_name":        "production",
					"agent_type":            "prometheus",
					"region_code":           "us-east-1",
					"allowed_probe_scopes":  []string{"metrics:submit", "health:report"},
					"organization_id":       1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, key *UnifiedAPIKey) {
				assert.True(t, key.IsMonitoringAgent())
				assert.Equal(t, "bearer", key.GetAuthenticationMethod())
				assert.Equal(t, "prometheus", key.AgentType)
				assert.Equal(t, "us-east-1", key.RegionCode)
			},
		},
		{
			name:       "success - registration key properties",
			keyID:      "key-registration",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":              1,
					"key_id":          "key-registration",
					"name":            "Server Registration",
					"type":            "registration",
					"status":          "active",
					"capabilities":    []string{"servers:register", "servers:update"},
					"organization_id": 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, key *UnifiedAPIKey) {
				assert.True(t, key.CanRegisterServers())
				assert.Equal(t, "headers", key.GetAuthenticationMethod())
			},
		},
		{
			name:       "success - admin key properties",
			keyID:      "key-admin",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":              1,
					"key_id":          "key-admin",
					"name":            "Admin Access",
					"type":            "admin",
					"status":          "active",
					"capabilities":    []string{"admin:read", "admin:write", "audit:read"},
					"organization_id": 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, key *UnifiedAPIKey) {
				assert.Equal(t, APIKeyTypeAdmin, key.Type)
				assert.True(t, key.HasCapability("admin:read"))
				assert.True(t, key.HasCapability("admin:write"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL:    server.URL,
				Auth:       AuthConfig{Token: "test-token"},
				RetryCount: 0,
			})
			require.NoError(t, err)

			key, err := client.APIKeys.GetUnified(context.Background(), tt.keyID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, key)
				if tt.checkFunc != nil {
					tt.checkFunc(t, key)
				}
			}
		})
	}
}

// TestAPIKeysService_AdvancedErrorScenarios tests additional error handling scenarios
func TestAPIKeysService_AdvancedErrorScenarios(t *testing.T) {
	tests := []struct {
		name       string
		operation  string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		errType    string
	}{
		{
			name:       "error - rate limit exceeded",
			operation:  "list",
			mockStatus: http.StatusTooManyRequests,
			mockBody: map[string]interface{}{
				"error": "rate limit exceeded",
			},
			wantErr: true,
			errType: "rate limit",
		},
		{
			name:       "error - conflict - duplicate key name",
			operation:  "create",
			mockStatus: http.StatusConflict,
			mockBody: map[string]interface{}{
				"error": "duplicate key name",
			},
			wantErr: true,
			errType: "conflict",
		},
		{
			name:       "error - invalid request format",
			operation:  "create",
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"error": "invalid request format",
			},
			wantErr: true,
			errType: "validation",
		},
		{
			name:       "error - service unavailable",
			operation:  "list",
			mockStatus: http.StatusServiceUnavailable,
			mockBody: map[string]interface{}{
				"error": "service temporarily unavailable",
			},
			wantErr: true,
			errType: "service unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			var ctx context.Context
			var cancel context.CancelFunc

			if tt.wantErr && tt.mockStatus >= 500 {
				ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
			} else {
				ctx = context.Background()
			}

			if tt.operation == "create" {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tt.mockStatus)
					json.NewEncoder(w).Encode(tt.mockBody)
				}))
			} else {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tt.mockStatus)
					json.NewEncoder(w).Encode(tt.mockBody)
				}))
			}
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL:    server.URL,
				Auth:       AuthConfig{Token: "test-token"},
				RetryCount: 0,
			})
			require.NoError(t, err)

			if tt.operation == "create" {
				_, err = client.APIKeys.CreateUnified(ctx, &CreateUnifiedAPIKeyRequest{
					Name: "Test",
					Type: APIKeyTypeUser,
				})
			} else {
				_, _, err = client.APIKeys.ListUnified(ctx, nil)
			}

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
