package nexmonyx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNotificationsService_SendNotification_TableDriven demonstrates table-driven test pattern
func TestNotificationsService_SendNotification_TableDriven(t *testing.T) {
	tests := []struct {
		name       string
		request    *NotificationRequest
		setupMock  func(w http.ResponseWriter, r *http.Request)
		wantErr    bool
		checkResp  func(t *testing.T, resp *NotificationResponse)
		checkError func(t *testing.T, err error)
	}{
		{
			name: "success - basic notification",
			request: &NotificationRequest{
				OrganizationID: 1,
				Subject:        "Test Notification",
				Content:        "This is a test",
				ContentType:    "text",
				Priority:       NotificationPriorityNormal,
			},
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/v1/notifications/send", r.URL.Path)

				helper := NewTestResponseHelper()
				helper.WriteSuccessResponse(w, map[string]interface{}{
					"notification_id": "notif-123",
					"status":          "sent",
				})
			},
			wantErr: false,
			checkResp: func(t *testing.T, resp *NotificationResponse) {
				assert.NotNil(t, resp)
			},
		},
		{
			name: "success - high priority with metadata",
			request: &NotificationRequest{
				OrganizationID: 1,
				Subject:        "Critical Alert",
				Content:        "<p>Critical system alert</p>",
				ContentType:    "html",
				Priority:       NotificationPriorityCritical,
				Metadata: map[string]interface{}{
					"alert_type": "system",
					"severity":   "critical",
				},
			},
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				helper := NewTestResponseHelper()
				helper.WriteSuccessResponse(w, map[string]interface{}{
					"notification_id": "notif-456",
					"status":          "sent",
					"channels_used": []map[string]interface{}{
						{
							"channel_id":   1,
							"channel_name": "Email",
							"channel_type": "email",
							"status":       "sent",
							"recipient":    "user@example.com",
						},
						{
							"channel_id":   2,
							"channel_name": "Slack",
							"channel_type": "slack",
							"status":       "sent",
							"recipient":    "#alerts",
						},
					},
				})
			},
			wantErr: false,
			checkResp: func(t *testing.T, resp *NotificationResponse) {
				assert.NotNil(t, resp)
			},
		},
		{
			name: "validation error - missing subject",
			request: &NotificationRequest{
				OrganizationID: 1,
				Content:        "Content without subject",
			},
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				helper := NewTestResponseHelper()
				helper.WriteValidationError(w, map[string][]string{
					"subject": {"Subject is required"},
				})
			},
			wantErr: true,
			checkError: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "unauthorized - invalid token",
			request: &NotificationRequest{
				OrganizationID: 1,
				Subject:        "Test",
				Content:        "Test content",
			},
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				NewErrorScenarios().WriteUnauthorizedError(w)
			},
			wantErr: true,
			checkError: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "rate limit exceeded",
			request: &NotificationRequest{
				OrganizationID: 1,
				Subject:        "Test",
				Content:        "Test content",
			},
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				helper := NewTestResponseHelper().WithRateLimit(1000, 1000)
				helper.WriteRateLimitError(w)
			},
			wantErr: true,
			checkError: func(t *testing.T, err error) {
				assert.Error(t, err)
				// Could check for specific rate limit error type here
			},
		},
		{
			name: "server error",
			request: &NotificationRequest{
				OrganizationID: 1,
				Subject:        "Test",
				Content:        "Test content",
			},
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				NewErrorScenarios().WriteServerError(w)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.setupMock))
			defer server.Close()

			client, err := NewClient(&Config{BaseURL: server.URL})
			require.NoError(t, err)

			resp, err := client.Notifications.SendNotification(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
				if tt.checkError != nil {
					tt.checkError(t, err)
				}
			} else {
				assert.NoError(t, err)
				if tt.checkResp != nil {
					tt.checkResp(t, resp)
				}
			}
		})
	}
}

// TestNotificationsService_SendBatchNotifications_TableDriven demonstrates batch operations with table-driven tests
func TestNotificationsService_SendBatchNotifications_TableDriven(t *testing.T) {
	tests := []struct {
		name       string
		request    *BatchNotificationRequest
		setupMock  func(w http.ResponseWriter, r *http.Request)
		wantErr    bool
		checkResp  func(t *testing.T, resp *BatchNotificationResponse)
	}{
		{
			name: "success - multiple notifications",
			request: &BatchNotificationRequest{
				Notifications: []NotificationRequest{
					{
						OrganizationID: 1,
						Subject:        "Notification 1",
						Content:        "Content 1",
					},
					{
						OrganizationID: 1,
						Subject:        "Notification 2",
						Content:        "Content 2",
					},
					{
						OrganizationID: 1,
						Subject:        "Notification 3",
						Content:        "Content 3",
					},
				},
			},
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/v1/notifications/send/batch", r.URL.Path)

				helper := NewTestResponseHelper()
				helper.WriteSuccessResponse(w, map[string]interface{}{
					"total_requested": 3,
					"total_accepted":  3,
					"total_rejected":  0,
				})
			},
			wantErr: false,
			checkResp: func(t *testing.T, resp *BatchNotificationResponse) {
				assert.NotNil(t, resp)
			},
		},
		{
			name: "partial success - some notifications failed",
			request: &BatchNotificationRequest{
				Notifications: []NotificationRequest{
					{OrganizationID: 1, Subject: "Valid 1", Content: "Content 1"},
					{OrganizationID: 1, Subject: "Valid 2", Content: "Content 2"},
				},
			},
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				helper := NewTestResponseHelper()
				helper.WriteSuccessResponse(w, map[string]interface{}{
					"total_requested": 2,
					"total_accepted":  1,
					"total_rejected":  1,
					"errors": []string{
						"Notification 1 failed: Invalid email configuration - No recipients configured",
					},
				})
			},
			wantErr: false,
			checkResp: func(t *testing.T, resp *BatchNotificationResponse) {
				assert.NotNil(t, resp)
			},
		},
		{
			name: "validation error - empty request",
			request: &BatchNotificationRequest{
				Notifications: []NotificationRequest{},
			},
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				helper := NewTestResponseHelper()
				helper.WriteValidationError(w, map[string][]string{
					"notifications": {"At least one notification is required"},
				})
			},
			wantErr: true,
		},
		{
			name: "server error",
			request: &BatchNotificationRequest{
				Notifications: []NotificationRequest{
					{OrganizationID: 1, Subject: "Test", Content: "Test"},
				},
			},
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				NewErrorScenarios().WriteServerError(w)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.setupMock))
			defer server.Close()

			client, err := NewClient(&Config{BaseURL: server.URL})
			require.NoError(t, err)

			resp, err := client.Notifications.SendBatchNotifications(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				if tt.checkResp != nil {
					tt.checkResp(t, resp)
				}
			}
		})
	}
}

// TestNotificationsService_GetNotificationStatus_TableDriven demonstrates status checking with table-driven tests
func TestNotificationsService_GetNotificationStatus_TableDriven(t *testing.T) {
	tests := []struct {
		name      string
		request   *NotificationStatusRequest
		setupMock func(w http.ResponseWriter, r *http.Request)
		wantErr   bool
		checkResp func(t *testing.T, resp *NotificationStatusResponse)
	}{
		{
			name: "success - single notification",
			request: &NotificationStatusRequest{
				NotificationIDs: []uint{123},
			},
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				helper := NewTestResponseHelper()
				helper.WriteSuccessResponse(w, map[string]interface{}{
					"notifications": []map[string]interface{}{
						{
							"notification_id": 123,
							"status":          "delivered",
							"delivered_at":    "2023-01-01T12:00:00Z",
						},
					},
				})
			},
			wantErr: false,
			checkResp: func(t *testing.T, resp *NotificationStatusResponse) {
				assert.NotNil(t, resp)
			},
		},
		{
			name: "success - multiple notifications with different statuses",
			request: &NotificationStatusRequest{
				NotificationIDs: []uint{123, 456, 789},
			},
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				helper := NewTestResponseHelper()
				helper.WriteSuccessResponse(w, map[string]interface{}{
					"notifications": []map[string]interface{}{
						{"notification_id": 123, "status": "delivered"},
						{"notification_id": 456, "status": "pending"},
						{"notification_id": 789, "status": "failed"},
					},
				})
			},
			wantErr: false,
			checkResp: func(t *testing.T, resp *NotificationStatusResponse) {
				assert.NotNil(t, resp)
			},
		},
		{
			name: "not found - invalid notification ID",
			request: &NotificationStatusRequest{
				NotificationIDs: []uint{999999},
			},
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				NewErrorScenarios().WriteNotFoundError(w, "Notification")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.setupMock))
			defer server.Close()

			client, err := NewClient(&Config{BaseURL: server.URL})
			require.NoError(t, err)

			resp, err := client.Notifications.GetNotificationStatus(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				if tt.checkResp != nil {
					tt.checkResp(t, resp)
				}
			}
		})
	}
}

// TestNotificationsService_SendQuotaAlert_TableDriven demonstrates quota alerts with table-driven tests
func TestNotificationsService_SendQuotaAlert_TableDriven(t *testing.T) {
	tests := []struct {
		name         string
		orgID        uint
		subject      string
		content      string
		priority     NotificationPriority
		metadata     map[string]interface{}
		setupMock    func(w http.ResponseWriter, r *http.Request)
		wantErr      bool
	}{
		{
			name:     "success - normal priority quota alert",
			orgID:    1,
			subject:  "Quota Alert",
			content:  "<p>You have reached 80% of your quota</p>",
			priority: NotificationPriorityNormal,
			metadata: map[string]interface{}{
				"quota_type": "servers",
				"usage":      "80%",
			},
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				helper := NewTestResponseHelper()
				helper.WriteSuccessResponse(w, map[string]interface{}{
					"notification_id": "quota-alert-123",
					"status":          "sent",
				})
			},
			wantErr: false,
		},
		{
			name:     "success - critical priority quota alert",
			orgID:    1,
			subject:  "Critical Quota Alert",
			content:  "<p>You have exceeded your quota</p>",
			priority: NotificationPriorityCritical,
			metadata: map[string]interface{}{
				"quota_type": "servers",
				"usage":      "105%",
				"action":     "service_limited",
			},
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				helper := NewTestResponseHelper()
				helper.WriteSuccessResponse(w, map[string]interface{}{
					"notification_id": "quota-alert-456",
					"status":          "sent",
				})
			},
			wantErr: false,
		},
		{
			name:     "success - nil metadata",
			orgID:    1,
			subject:  "Simple Alert",
			content:  "Simple quota alert",
			priority: NotificationPriorityNormal,
			metadata: nil,
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				helper := NewTestResponseHelper()
				helper.WriteSuccessResponse(w, map[string]interface{}{
					"notification_id": "alert-789",
					"status":          "sent",
				})
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.setupMock))
			defer server.Close()

			client, err := NewClient(&Config{BaseURL: server.URL})
			require.NoError(t, err)

			resp, err := client.Notifications.SendQuotaAlert(
				context.Background(),
				tt.orgID,
				tt.subject,
				tt.content,
				tt.priority,
				tt.metadata,
			)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}
