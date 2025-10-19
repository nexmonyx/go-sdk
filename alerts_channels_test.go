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

// TestAlertsService_CreateChannel tests the CreateChannel method with various scenarios
func TestAlertsService_CreateChannel(t *testing.T) {
	tests := []struct {
		name       string
		channel    *AlertChannel
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *AlertChannel)
	}{
		{
			name: "success - create email channel",
			channel: &AlertChannel{
				Name:    "Email Alerts",
				Type:    "email",
				Enabled: true,
				Configuration: map[string]interface{}{
					"recipients": []string{"admin@example.com", "ops@example.com"},
					"reply_to":   "noreply@example.com",
					"template":   "default",
				},
			},
			mockStatus: http.StatusCreated,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":              1,
					"name":            "Email Alerts",
					"type":            "email",
					"enabled":         true,
					"organization_id": 1,
					"configuration": map[string]interface{}{
						"recipients": []interface{}{"admin@example.com", "ops@example.com"},
						"reply_to":   "noreply@example.com",
						"template":   "default",
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, ch *AlertChannel) {
				assert.Equal(t, "Email Alerts", ch.Name)
				assert.Equal(t, "email", ch.Type)
				assert.True(t, ch.Enabled)
			},
		},
		{
			name: "success - create slack channel",
			channel: &AlertChannel{
				Name:    "Slack Notifications",
				Type:    "slack",
				Enabled: true,
				Configuration: map[string]interface{}{
					"webhook_url":      "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX",
					"channel_override": "#alerts",
					"attachments":      true,
				},
			},
			mockStatus: http.StatusCreated,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":              2,
					"name":            "Slack Notifications",
					"type":            "slack",
					"enabled":         true,
					"organization_id": 1,
					"configuration": map[string]interface{}{
						"webhook_url":      "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX",
						"channel_override": "#alerts",
						"attachments":      true,
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, ch *AlertChannel) {
				assert.Equal(t, "Slack Notifications", ch.Name)
				assert.Equal(t, "slack", ch.Type)
			},
		},
		{
			name: "success - create webhook channel",
			channel: &AlertChannel{
				Name:    "Custom Webhook",
				Type:    "webhook",
				Enabled: true,
				Configuration: map[string]interface{}{
					"endpoint": "https://api.example.com/webhook",
					"auth_headers": map[string]interface{}{
						"Authorization": "Bearer token123",
						"X-API-Key":     "key123",
					},
					"payload_template": "json",
				},
			},
			mockStatus: http.StatusCreated,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":              3,
					"name":            "Custom Webhook",
					"type":            "webhook",
					"enabled":         true,
					"organization_id": 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, ch *AlertChannel) {
				assert.Equal(t, "Custom Webhook", ch.Name)
				assert.Equal(t, "webhook", ch.Type)
			},
		},
		{
			name: "success - create pagerduty channel",
			channel: &AlertChannel{
				Name:    "PagerDuty Integration",
				Type:    "pagerduty",
				Enabled: true,
				Configuration: map[string]interface{}{
					"integration_key": "abc123xyz456",
					"severity_mapping": map[string]interface{}{
						"critical": "critical",
						"warning":  "warning",
						"info":     "info",
					},
				},
			},
			mockStatus: http.StatusCreated,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":              4,
					"name":            "PagerDuty Integration",
					"type":            "pagerduty",
					"enabled":         true,
					"organization_id": 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, ch *AlertChannel) {
				assert.Equal(t, "PagerDuty Integration", ch.Name)
				assert.Equal(t, "pagerduty", ch.Type)
			},
		},
		{
			name: "validation error - missing name",
			channel: &AlertChannel{
				Type:    "email",
				Enabled: true,
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"error": "name is required",
			},
			wantErr: true,
		},
		{
			name: "validation error - missing type",
			channel: &AlertChannel{
				Name:    "Test Channel",
				Enabled: true,
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"error": "type is required",
			},
			wantErr: true,
		},
		{
			name: "validation error - invalid type",
			channel: &AlertChannel{
				Name:    "Test Channel",
				Type:    "invalid_type",
				Enabled: true,
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"error": "invalid channel type",
			},
			wantErr: true,
		},
		{
			name: "validation error - missing configuration",
			channel: &AlertChannel{
				Name:    "Test Channel",
				Type:    "email",
				Enabled: true,
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"error": "configuration is required",
			},
			wantErr: true,
		},
		{
			name: "validation error - invalid email configuration",
			channel: &AlertChannel{
				Name:    "Bad Email",
				Type:    "email",
				Enabled: true,
				Configuration: map[string]interface{}{
					"recipients": "not-an-array",
				},
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"error": "recipients must be an array of email addresses",
			},
			wantErr: true,
		},
		{
			name: "validation error - invalid slack webhook",
			channel: &AlertChannel{
				Name:    "Bad Slack",
				Type:    "slack",
				Enabled: true,
				Configuration: map[string]interface{}{
					"webhook_url": "invalid-url",
				},
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"error": "invalid slack webhook URL",
			},
			wantErr: true,
		},
		{
			name: "validation error - invalid pagerduty key",
			channel: &AlertChannel{
				Name:    "Bad PagerDuty",
				Type:    "pagerduty",
				Enabled: true,
				Configuration: map[string]interface{}{
					"integration_key": "",
				},
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"error": "integration_key is required for pagerduty channels",
			},
			wantErr: true,
		},
		{
			name: "unauthorized - invalid token",
			channel: &AlertChannel{
				Name: "Test",
				Type: "email",
			},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"error": "invalid authentication token",
			},
			wantErr: true,
		},
		{
			name: "forbidden - insufficient permissions",
			channel: &AlertChannel{
				Name: "Test",
				Type: "email",
			},
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"error": "insufficient permissions to create channels",
			},
			wantErr: true,
		},
		{
			name: "conflict - duplicate channel name",
			channel: &AlertChannel{
				Name: "Existing Channel",
				Type: "email",
			},
			mockStatus: http.StatusConflict,
			mockBody: map[string]interface{}{
				"error": "channel name already exists",
			},
			wantErr: true,
		},
		{
			name: "server error",
			channel: &AlertChannel{
				Name: "Test",
				Type: "email",
			},
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"error": "internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/v1/alerts/channels", r.URL.Path)

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

			ctx := context.Background()
			channel, err := client.Alerts.CreateChannel(ctx, tt.channel)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, channel)
				if tt.checkFunc != nil {
					tt.checkFunc(t, channel)
				}
			}
		})
	}
}

// TestAlertsService_GetChannel tests the GetChannel method
func TestAlertsService_GetChannel(t *testing.T) {
	tests := []struct {
		name       string
		channelID  string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *AlertChannel)
	}{
		{
			name:       "success - get email channel",
			channelID:  "1",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":              1,
					"name":            "Email Alerts",
					"type":            "email",
					"enabled":         true,
					"organization_id": 1,
					"configuration": map[string]interface{}{
						"recipients": []interface{}{"admin@example.com"},
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, ch *AlertChannel) {
				assert.Equal(t, "Email Alerts", ch.Name)
				assert.Equal(t, "email", ch.Type)
			},
		},
		{
			name:       "success - get slack channel",
			channelID:  "2",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":              2,
					"name":            "Slack Notifications",
					"type":            "slack",
					"enabled":         true,
					"organization_id": 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, ch *AlertChannel) {
				assert.Equal(t, "Slack Notifications", ch.Name)
				assert.Equal(t, "slack", ch.Type)
			},
		},
		{
			name:       "success - get webhook channel",
			channelID:  "3",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":              3,
					"name":            "Custom Webhook",
					"type":            "webhook",
					"enabled":         true,
					"organization_id": 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, ch *AlertChannel) {
				assert.Equal(t, "Custom Webhook", ch.Name)
				assert.Equal(t, "webhook", ch.Type)
			},
		},
		{
			name:       "not found - invalid channel id",
			channelID:  "999",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"error": "channel not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			channelID:  "1",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"error": "invalid token",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			channelID:  "1",
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"error": "internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)

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

			ctx := context.Background()
			channel, err := client.Alerts.GetChannel(ctx, tt.channelID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, channel)
				if tt.checkFunc != nil {
					tt.checkFunc(t, channel)
				}
			}
		})
	}
}

// TestAlertsService_UpdateChannel tests the UpdateChannel method
func TestAlertsService_UpdateChannel(t *testing.T) {
	tests := []struct {
		name       string
		channelID  string
		channel    *AlertChannel
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *AlertChannel)
	}{
		{
			name:      "success - update channel name",
			channelID: "1",
			channel: &AlertChannel{
				Name: "Updated Email Alerts",
				Type: "email",
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":      1,
					"name":    "Updated Email Alerts",
					"type":    "email",
					"enabled": true,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, ch *AlertChannel) {
				assert.Equal(t, "Updated Email Alerts", ch.Name)
			},
		},
		{
			name:      "success - update channel configuration",
			channelID: "1",
			channel: &AlertChannel{
				Name: "Email Alerts",
				Type: "email",
				Configuration: map[string]interface{}{
					"recipients": []string{"newemail@example.com"},
				},
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":   1,
					"name": "Email Alerts",
					"type": "email",
				},
			},
			wantErr: false,
		},
		{
			name:      "success - enable channel",
			channelID: "1",
			channel: &AlertChannel{
				Enabled: true,
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":      1,
					"enabled": true,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, ch *AlertChannel) {
				assert.True(t, ch.Enabled)
			},
		},
		{
			name:      "success - disable channel",
			channelID: "1",
			channel: &AlertChannel{
				Enabled: false,
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":      1,
					"enabled": false,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, ch *AlertChannel) {
				assert.False(t, ch.Enabled)
			},
		},
		{
			name:      "success - update email recipients",
			channelID: "1",
			channel: &AlertChannel{
				Configuration: map[string]interface{}{
					"recipients": []string{"admin1@example.com", "admin2@example.com", "admin3@example.com"},
				},
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id": 1,
				},
			},
			wantErr: false,
		},
		{
			name:      "success - update slack webhook",
			channelID: "2",
			channel: &AlertChannel{
				Configuration: map[string]interface{}{
					"webhook_url": "https://hooks.slack.com/services/NEW/WEBHOOK/URL",
				},
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id": 2,
				},
			},
			wantErr: false,
		},
		{
			name:      "success - update webhook headers",
			channelID: "3",
			channel: &AlertChannel{
				Configuration: map[string]interface{}{
					"auth_headers": map[string]interface{}{
						"Authorization": "Bearer newtoken",
					},
				},
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id": 3,
				},
			},
			wantErr: false,
		},
		{
			name:      "validation error - invalid configuration",
			channelID: "1",
			channel: &AlertChannel{
				Configuration: map[string]interface{}{
					"recipients": "invalid",
				},
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"error": "invalid configuration",
			},
			wantErr: true,
		},
		{
			name:      "not found - invalid channel id",
			channelID: "999",
			channel: &AlertChannel{
				Name: "Test",
			},
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"error": "channel not found",
			},
			wantErr: true,
		},
		{
			name:      "unauthorized",
			channelID: "1",
			channel: &AlertChannel{
				Name: "Test",
			},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"error": "invalid token",
			},
			wantErr: true,
		},
		{
			name:      "server error",
			channelID: "1",
			channel: &AlertChannel{
				Name: "Test",
			},
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"error": "internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPut, r.Method)

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

			ctx := context.Background()
			channel, err := client.Alerts.UpdateChannel(ctx, tt.channelID, tt.channel)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, channel)
				if tt.checkFunc != nil {
					tt.checkFunc(t, channel)
				}
			}
		})
	}
}

// TestAlertsService_DeleteChannel tests the DeleteChannel method
func TestAlertsService_DeleteChannel(t *testing.T) {
	tests := []struct {
		name       string
		channelID  string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:       "success - delete channel (204)",
			channelID:  "1",
			mockStatus: http.StatusNoContent,
			mockBody:   nil,
			wantErr:    false,
		},
		{
			name:       "success - delete channel (200 with body)",
			channelID:  "2",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "deleted",
			},
			wantErr: false,
		},
		{
			name:       "not found - invalid channel id",
			channelID:  "999",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"error": "channel not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			channelID:  "1",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"error": "invalid token",
			},
			wantErr: true,
		},
		{
			name:       "forbidden - channel in use",
			channelID:  "1",
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"error": "cannot delete channel in use by active alerts",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			channelID:  "1",
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"error": "internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodDelete, r.Method)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				if tt.mockBody != nil {
					json.NewEncoder(w).Encode(tt.mockBody)
				}
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL:    server.URL,
				Auth:       AuthConfig{Token: "test-token"},
				RetryCount: 0,
			})
			require.NoError(t, err)

			ctx := context.Background()
			err = client.Alerts.DeleteChannel(ctx, tt.channelID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestAlertsService_TestChannel tests the TestChannel method
func TestAlertsService_TestChannel(t *testing.T) {
	tests := []struct {
		name       string
		channelID  string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *ChannelTestResult)
	}{
		{
			name:       "success - test email channel",
			channelID:  "1",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"success": true,
					"message": "Test email sent successfully",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, result *ChannelTestResult) {
				assert.True(t, result.Success)
				assert.Contains(t, result.Message, "success")
			},
		},
		{
			name:       "success - test slack channel",
			channelID:  "2",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"success": true,
					"message": "Test message posted to Slack",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, result *ChannelTestResult) {
				assert.True(t, result.Success)
			},
		},
		{
			name:       "success - test webhook channel",
			channelID:  "3",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"success": true,
					"message": "Test webhook delivered",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, result *ChannelTestResult) {
				assert.True(t, result.Success)
			},
		},
		{
			name:       "success - test pagerduty channel",
			channelID:  "4",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"success": true,
					"message": "Test incident created in PagerDuty",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, result *ChannelTestResult) {
				assert.True(t, result.Success)
			},
		},
		{
			name:       "failure - invalid email configuration",
			channelID:  "1",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"success": false,
					"message": "Failed to send test email",
					"errors":  []string{"invalid email address"},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, result *ChannelTestResult) {
				assert.False(t, result.Success)
				assert.NotEmpty(t, result.Errors)
			},
		},
		{
			name:       "failure - slack webhook unreachable",
			channelID:  "2",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"success": false,
					"message": "Slack webhook unreachable",
					"errors":  []string{"connection timeout"},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, result *ChannelTestResult) {
				assert.False(t, result.Success)
			},
		},
		{
			name:       "not found - invalid channel id",
			channelID:  "999",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"error": "channel not found",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			channelID:  "1",
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"error": "internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)

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

			ctx := context.Background()
			result, err := client.Alerts.TestChannel(ctx, tt.channelID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestAlertsService_ListChannels_Comprehensive tests the ListChannels method comprehensively
func TestAlertsService_ListChannels_Comprehensive(t *testing.T) {
	tests := []struct {
		name       string
		opts       *ListOptions
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, []*AlertChannel, *PaginationMeta)
	}{
		{
			name:       "success - list all channels",
			opts:       nil,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"id":   1,
						"name": "Email Alerts",
						"type": "email",
					},
					map[string]interface{}{
						"id":   2,
						"name": "Slack Notifications",
						"type": "slack",
					},
					map[string]interface{}{
						"id":   3,
						"name": "Custom Webhook",
						"type": "webhook",
					},
				},
				"pagination": map[string]interface{}{
					"page":  1,
					"limit": 25,
					"total": 3,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, channels []*AlertChannel, meta *PaginationMeta) {
				assert.Len(t, channels, 3)
				if meta != nil {
					assert.Equal(t, 1, meta.Page)
					assert.Equal(t, 25, meta.Limit)
				}
			},
		},
		{
			name: "success - list with pagination",
			opts: &ListOptions{
				Page:  1,
				Limit: 10,
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"id":   1,
						"name": "Email Alerts",
						"type": "email",
					},
				},
				"pagination": map[string]interface{}{
					"page":  1,
					"limit": 10,
					"total": 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, channels []*AlertChannel, meta *PaginationMeta) {
				assert.Len(t, channels, 1)
			},
		},
		{
			name:       "success - empty list",
			opts:       nil,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": []interface{}{},
				"pagination": map[string]interface{}{
					"page":  1,
					"limit": 25,
					"total": 0,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, channels []*AlertChannel, meta *PaginationMeta) {
				assert.Len(t, channels, 0)
			},
		},
		{
			name: "success - filter by type",
			opts: &ListOptions{
				Search: "type:email",
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"id":   1,
						"name": "Email Alerts",
						"type": "email",
					},
				},
				"pagination": map[string]interface{}{
					"page":  1,
					"limit": 25,
					"total": 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, channels []*AlertChannel, meta *PaginationMeta) {
				assert.Len(t, channels, 1)
				assert.Equal(t, "email", channels[0].Type)
			},
		},
		{
			name: "success - search by name",
			opts: &ListOptions{
				Search: "Slack",
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"id":   2,
						"name": "Slack Notifications",
						"type": "slack",
					},
				},
				"pagination": map[string]interface{}{
					"page":  1,
					"limit": 25,
					"total": 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, channels []*AlertChannel, meta *PaginationMeta) {
				assert.Len(t, channels, 1)
				assert.Equal(t, "Slack Notifications", channels[0].Name)
			},
		},
		{
			name: "validation error - invalid pagination",
			opts: &ListOptions{
				Page:  0,
				Limit: 0,
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"error": "invalid pagination parameters",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			opts:       nil,
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"error": "invalid token",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			opts:       nil,
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"error": "internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)

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

			ctx := context.Background()
			channels, meta, err := client.Alerts.ListChannels(ctx, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, channels, meta)
				}
			}
		})
	}
}
