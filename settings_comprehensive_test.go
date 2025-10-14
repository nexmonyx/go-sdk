package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSettingsService_Get(t *testing.T) {
	tests := []struct {
		name           string
		orgID          string
		responseStatus int
		responseBody   interface{}
		wantErr        bool
		validateResult func(*testing.T, *Settings)
	}{
		{
			name:           "get complete settings",
			orgID:          "org-123",
			responseStatus: http.StatusOK,
			responseBody: StandardResponse{
				Status: "success",
				Data: &Settings{
					OrganizationID: 123,
					GeneralSettings: &GeneralSettings{
						TimeZone:         "America/New_York",
						DateFormat:       "YYYY-MM-DD",
						TimeFormat:       "HH:mm:ss",
						Language:         "en",
						DefaultDashboard: "overview",
						SessionTimeout:   30,
					},
					SecuritySettings: &SecuritySettings{
						Require2FA:        true,
						AllowAPIKeys:      true,
						APIKeyExpiration:  90,
						AuditLogRetention: 365,
					},
					NotificationSettings: &NotificationSettings{
						EmailEnabled:     true,
						EmailRecipients:  []string{"admin@example.com"},
						SlackEnabled:     true,
						WebhooksEnabled:  true,
					},
				},
			},
			validateResult: func(t *testing.T, s *Settings) {
				assert.Equal(t, uint(123), s.OrganizationID)
				assert.NotNil(t, s.GeneralSettings)
				assert.Equal(t, "America/New_York", s.GeneralSettings.TimeZone)
				assert.True(t, s.SecuritySettings.Require2FA)
				assert.True(t, s.NotificationSettings.EmailEnabled)
			},
		},
		{
			name:           "organization not found",
			orgID:          "nonexistent",
			responseStatus: http.StatusNotFound,
			responseBody: ErrorResponse{
				Status:  "error",
				Message: "organization not found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/organizations/")
				assert.Contains(t, r.URL.Path, "/settings")

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL})
			result, err := client.Settings.Get(context.Background(), tt.orgID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}

func TestSettingsService_Update(t *testing.T) {
	tests := []struct {
		name           string
		orgID          string
		settings       *Settings
		responseStatus int
		wantErr        bool
		validateResult func(*testing.T, *Settings)
	}{
		{
			name:  "update general settings",
			orgID: "org-123",
			settings: &Settings{
				OrganizationID: 123,
				GeneralSettings: &GeneralSettings{
					TimeZone:       "UTC",
					DateFormat:     "DD/MM/YYYY",
					Language:       "fr",
					SessionTimeout: 60,
				},
			},
			responseStatus: http.StatusOK,
			validateResult: func(t *testing.T, s *Settings) {
				assert.Equal(t, "UTC", s.GeneralSettings.TimeZone)
				assert.Equal(t, "fr", s.GeneralSettings.Language)
			},
		},
		{
			name:  "update security settings",
			orgID: "org-456",
			settings: &Settings{
				OrganizationID: 456,
				SecuritySettings: &SecuritySettings{
					Require2FA:        true,
					AllowAPIKeys:      false,
					APIKeyExpiration:  30,
					AuditLogRetention: 180,
				},
			},
			responseStatus: http.StatusOK,
			validateResult: func(t *testing.T, s *Settings) {
				assert.True(t, s.SecuritySettings.Require2FA)
				assert.False(t, s.SecuritySettings.AllowAPIKeys)
			},
		},
		{
			name:  "update notification settings",
			orgID: "org-789",
			settings: &Settings{
				OrganizationID: 789,
				NotificationSettings: &NotificationSettings{
					EmailEnabled:    true,
					EmailRecipients: []string{"ops@example.com", "dev@example.com"},
					SlackEnabled:    true,
					SlackWebhook:    "https://hooks.slack.com/services/XXX",
				},
			},
			responseStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "PUT", r.Method)
				assert.Contains(t, r.URL.Path, "/organizations/")
				assert.Contains(t, r.URL.Path, "/settings")

				var req Settings
				json.NewDecoder(r.Body).Decode(&req)

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(StandardResponse{
					Status: "success",
					Data:   tt.settings,
				})
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL})
			result, err := client.Settings.Update(context.Background(), tt.orgID, tt.settings)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}

func TestSettingsService_GetNotificationSettings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/settings/notifications")

		json.NewEncoder(w).Encode(StandardResponse{
			Status: "success",
			Data: &NotificationSettings{
				EmailEnabled:     true,
				EmailRecipients:  []string{"admin@example.com"},
				SlackEnabled:     true,
				SlackWebhook:     "https://hooks.slack.com/services/TEST",
				PagerDutyEnabled: true,
				WebhooksEnabled:  true,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, err := client.Settings.GetNotificationSettings(context.Background(), "org-123")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.EmailEnabled)
	assert.True(t, result.SlackEnabled)
	assert.Len(t, result.EmailRecipients, 1)
}

func TestSettingsService_UpdateNotificationSettings(t *testing.T) {
	tests := []struct {
		name     string
		settings *NotificationSettings
		validate func(*testing.T, *NotificationSettings)
	}{
		{
			name: "enable all channels",
			settings: &NotificationSettings{
				EmailEnabled:     true,
				EmailRecipients:  []string{"ops@example.com"},
				SlackEnabled:     true,
				SlackWebhook:     "https://hooks.slack.com/xxx",
				PagerDutyEnabled: true,
				WebhooksEnabled:  true,
			},
			validate: func(t *testing.T, ns *NotificationSettings) {
				assert.True(t, ns.EmailEnabled)
				assert.True(t, ns.SlackEnabled)
				assert.True(t, ns.PagerDutyEnabled)
			},
		},
		{
			name: "disable email, enable slack only",
			settings: &NotificationSettings{
				EmailEnabled: false,
				SlackEnabled: true,
				SlackWebhook: "https://hooks.slack.com/yyy",
			},
			validate: func(t *testing.T, ns *NotificationSettings) {
				assert.False(t, ns.EmailEnabled)
				assert.True(t, ns.SlackEnabled)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "PUT", r.Method)
				assert.Contains(t, r.URL.Path, "/settings/notifications")

				var req NotificationSettings
				json.NewDecoder(r.Body).Decode(&req)

				json.NewEncoder(w).Encode(StandardResponse{
					Status: "success",
					Data:   tt.settings,
				})
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL})
			result, err := client.Settings.UpdateNotificationSettings(context.Background(), "org-123", tt.settings)

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestSettingsService_CompleteSettingsWorkflow(t *testing.T) {
	// Test: Get -> Modify -> Update -> Get again
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		if r.Method == "GET" {
			json.NewEncoder(w).Encode(StandardResponse{
				Status: "success",
				Data: &Settings{
					OrganizationID: 123,
					GeneralSettings: &GeneralSettings{
						TimeZone: "UTC",
						Language: "en",
					},
				},
			})
		} else if r.Method == "PUT" {
			json.NewEncoder(w).Encode(StandardResponse{
				Status: "success",
				Data: &Settings{
					OrganizationID: 123,
					GeneralSettings: &GeneralSettings{
						TimeZone: "America/New_York",
						Language: "en",
					},
				},
			})
		}
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	// Get current settings
	settings, err := client.Settings.Get(context.Background(), "org-123")
	require.NoError(t, err)
	assert.Equal(t, "UTC", settings.GeneralSettings.TimeZone)

	// Update settings
	settings.GeneralSettings.TimeZone = "America/New_York"
	updated, err := client.Settings.Update(context.Background(), "org-123", settings)
	require.NoError(t, err)
	assert.Equal(t, "America/New_York", updated.GeneralSettings.TimeZone)

	// Verify update
	final, err := client.Settings.Get(context.Background(), "org-123")
	require.NoError(t, err)
	assert.Equal(t, "America/New_York", final.GeneralSettings.TimeZone)

	assert.Equal(t, 3, callCount)
}

func TestSettingsService_SecurityPolicies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(StandardResponse{
			Status: "success",
			Data: &Settings{
				OrganizationID: 123,
				SecuritySettings: &SecuritySettings{
					Require2FA:       true,
					AllowAPIKeys:     true,
					APIKeyExpiration: 90,
					AllowedDomains:   []string{"example.com", "test.com"},
					SSOEnabled:       true,
					SSOProvider:      "okta",
					SSOConfig: map[string]interface{}{
						"domain": "company.okta.com",
					},
				},
				GeneralSettings: &GeneralSettings{
					PasswordPolicy: &PasswordPolicy{
						MinLength:           12,
						RequireUppercase:    true,
						RequireLowercase:    true,
						RequireNumbers:      true,
						RequireSpecialChars: true,
						ExpirationDays:      90,
						HistoryCount:        5,
					},
				},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	settings, err := client.Settings.Get(context.Background(), "org-123")

	require.NoError(t, err)
	assert.True(t, settings.SecuritySettings.Require2FA)
	assert.True(t, settings.SecuritySettings.SSOEnabled)
	assert.Equal(t, "okta", settings.SecuritySettings.SSOProvider)
	assert.NotNil(t, settings.GeneralSettings.PasswordPolicy)
	assert.Equal(t, 12, settings.GeneralSettings.PasswordPolicy.MinLength)
}

func TestSettingsService_MonitoringSettings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(StandardResponse{
			Status: "success",
			Data: &Settings{
				OrganizationID: 123,
				MonitoringSettings: &MonitoringSettings{
					MetricsRetention:    30,
					DefaultInterval:     60,
					EnableAutoDiscovery: true,
					AutoDiscoveryFilters: []string{
						"env:production",
						"type:web",
					},
					DefaultAlertThresholds: map[string]float64{
						"cpu":    80.0,
						"memory": 85.0,
						"disk":   90.0,
					},
				},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	settings, err := client.Settings.Get(context.Background(), "org-123")

	require.NoError(t, err)
	assert.Equal(t, 30, settings.MonitoringSettings.MetricsRetention)
	assert.True(t, settings.MonitoringSettings.EnableAutoDiscovery)
	assert.Len(t, settings.MonitoringSettings.AutoDiscoveryFilters, 2)
	assert.Equal(t, 80.0, settings.MonitoringSettings.DefaultAlertThresholds["cpu"])
}

func TestSettingsService_IntegrationSettings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(StandardResponse{
			Status: "success",
			Data: &Settings{
				OrganizationID: 123,
				IntegrationSettings: &IntegrationSettings{
					AWSEnabled: true,
					AWSConfig: map[string]interface{}{
						"access_key_id": "AKIAIOSFODNN7EXAMPLE",
						"region":        "us-east-1",
					},
					DatadogEnabled: true,
					DatadogConfig: map[string]interface{}{
						"api_key": "xxx",
						"app_key": "yyy",
					},
				},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	settings, err := client.Settings.Get(context.Background(), "org-123")

	require.NoError(t, err)
	assert.True(t, settings.IntegrationSettings.AWSEnabled)
	assert.True(t, settings.IntegrationSettings.DatadogEnabled)
	assert.NotNil(t, settings.IntegrationSettings.AWSConfig)
}

func TestSettingsService_CustomSettings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(StandardResponse{
			Status: "success",
			Data: &Settings{
				OrganizationID: 123,
				CustomSettings: map[string]interface{}{
					"feature_flags": map[string]bool{
						"new_dashboard": true,
						"beta_features": false,
					},
					"custom_branding": map[string]string{
						"logo_url":    "https://example.com/logo.png",
						"primary_color": "#007bff",
					},
				},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	settings, err := client.Settings.Get(context.Background(), "org-123")

	require.NoError(t, err)
	assert.NotNil(t, settings.CustomSettings)
	assert.Contains(t, settings.CustomSettings, "feature_flags")
	assert.Contains(t, settings.CustomSettings, "custom_branding")
}
