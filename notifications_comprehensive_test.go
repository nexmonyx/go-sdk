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

func TestNotificationsService_SendNotification(t *testing.T) {
	tests := []struct {
		name           string
		request        *NotificationRequest
		responseStatus int
		responseBody   interface{}
		wantErr        bool
		validateResult func(*testing.T, *NotificationResponse)
	}{
		{
			name: "successful email notification",
			request: &NotificationRequest{
				OrganizationID: 123,
				Subject:        "Test Alert",
				Content:        "This is a test notification",
				ContentType:    "text",
				Priority:       NotificationPriorityHigh,
				ChannelTypes:   []string{"email"},
			},
			responseStatus: http.StatusOK,
			responseBody: StandardResponse{
				Status:  "success",
				Message: "Notification sent successfully",
				Data: &NotificationResponse{
					ID:     123,
					Status: "sent",
					ChannelsUsed: []ChannelUsageInfo{
						{ChannelType: "email", Status: "sent"},
					},
				},
			},
			validateResult: func(t *testing.T, resp *NotificationResponse) {
				assert.Equal(t, uint(123), resp.ID)
				assert.Equal(t, "sent", resp.Status)
				assert.Len(t, resp.ChannelsUsed, 1)
			},
		},
		{
			name: "multi-channel notification",
			request: &NotificationRequest{
				OrganizationID: 456,
				Subject:        "Critical Alert",
				Content:        "<h1>System Critical</h1><p>Immediate action required</p>",
				ContentType:    "html",
				Priority:       NotificationPriorityCritical,
				ChannelTypes:   []string{"email", "slack", "webhook"},
				Metadata: map[string]interface{}{
					"server_id": 789,
					"alert_id":  101,
				},
			},
			responseStatus: http.StatusOK,
			responseBody: StandardResponse{
				Status: "success",
				Data: &NotificationResponse{
					ID:     456,
					Status: "sent",
					ChannelsUsed: []ChannelUsageInfo{
						{ChannelType: "email", Status: "sent"},
						{ChannelType: "slack", Status: "sent"},
						{ChannelType: "webhook", Status: "sent"},
					},
				},
			},
			validateResult: func(t *testing.T, resp *NotificationResponse) {
				assert.Len(t, resp.ChannelsUsed, 3)
			},
		},
		{
			name: "partial failure notification",
			request: &NotificationRequest{
				OrganizationID: 789,
				Subject:        "Warning",
				Content:        "Partial notification test",
				Priority:       NotificationPriorityNormal,
				ChannelTypes:   []string{"email", "slack", "sms"},
			},
			responseStatus: http.StatusOK,
			responseBody: StandardResponse{
				Status: "partial_success",
				Data: &NotificationResponse{
					ID:     789,
					Status: "partial",
					ChannelsUsed: []ChannelUsageInfo{
						{ChannelType: "email", Status: "sent"},
						{ChannelType: "slack", Status: "sent"},
						{ChannelType: "sms", Status: "failed", Error: "SMS service unavailable"},
					},
				},
			},
			validateResult: func(t *testing.T, resp *NotificationResponse) {
				assert.Equal(t, "partial", resp.Status)
				assert.Len(t, resp.ChannelsUsed, 3)
				// Verify one channel failed
				failedCount := 0
				for _, ch := range resp.ChannelsUsed {
					if ch.Status == "failed" {
						failedCount++
					}
				}
				assert.Equal(t, 1, failedCount)
			},
		},
		{
			name: "validation error",
			request: &NotificationRequest{
				OrganizationID: 0,
				Subject:        "",
				Content:        "",
			},
			responseStatus: http.StatusBadRequest,
			responseBody: ErrorResponse{
				Status:  "error",
				Message: "validation failed",
				Error:   "organization_id, subject, and content are required",
			},
			wantErr: true,
		},
		{
			name: "notification with custom templates",
			request: &NotificationRequest{
				OrganizationID: 111,
				Subject:        "Custom Template",
				Content:        "Template content",
				Priority:       NotificationPriorityLow,
				Metadata: map[string]interface{}{
					"template_id": "quota_alert",
					"template_vars": map[string]string{
						"org_name": "ACME Corp",
						"usage":    "95%",
					},
				},
			},
			responseStatus: http.StatusOK,
			responseBody: StandardResponse{
				Status: "success",
				Data: &NotificationResponse{
					ID:     111,
					Status: "sent",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/v1/notifications/send", r.URL.Path)

				// Verify request body
				var req NotificationRequest
				err := json.NewDecoder(r.Body).Decode(&req)
				require.NoError(t, err)
				assert.Equal(t, tt.request.OrganizationID, req.OrganizationID)
				assert.Equal(t, tt.request.Subject, req.Subject)

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL})
			result, err := client.Notifications.SendNotification(context.Background(), tt.request)

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

func TestNotificationsService_SendBatchNotifications(t *testing.T) {
	tests := []struct {
		name           string
		request        *BatchNotificationRequest
		responseStatus int
		responseBody   interface{}
		wantErr        bool
		validateResult func(*testing.T, *BatchNotificationResponse)
	}{
		{
			name: "successful batch send",
			request: &BatchNotificationRequest{
				Notifications: []NotificationRequest{
					{
						OrganizationID: 123,
						Subject:        "Alert 1",
						Content:        "Content 1",
						Priority:       NotificationPriorityHigh,
					},
					{
						OrganizationID: 123,
						Subject:        "Alert 2",
						Content:        "Content 2",
						Priority:       NotificationPriorityNormal,
					},
					{
						OrganizationID: 456,
						Subject:        "Alert 3",
						Content:        "Content 3",
						Priority:       NotificationPriorityLow,
					},
				},
			},
			responseStatus: http.StatusOK,
			responseBody: StandardResponse{
				Status:  "success",
				Message: "Batch notifications sent",
				Data: &BatchNotificationResponse{
					TotalRequested: 3,
					TotalAccepted:      3,
					TotalRejected:    0,
					Results: []NotificationResponse{
						{ID: 1, Status: "sent"},
						{ID: 2, Status: "sent"},
						{ID: 3, Status: "sent"},
					},
				},
			},
			validateResult: func(t *testing.T, resp *BatchNotificationResponse) {
				assert.Equal(t, 3, resp.TotalRequested)
				assert.Equal(t, 3, resp.TotalAccepted)
				assert.Equal(t, 0, resp.TotalRejected)
				assert.Len(t, resp.Results, 3)
			},
		},
		{
			name: "partial batch success",
			request: &BatchNotificationRequest{
				Notifications: []NotificationRequest{
					{OrganizationID: 123, Subject: "Valid 1", Content: "Content 1"},
					{OrganizationID: 0, Subject: "", Content: ""},
					{OrganizationID: 456, Subject: "Valid 2", Content: "Content 2"},
				},
			},
			responseStatus: http.StatusOK,
			responseBody: StandardResponse{
				Status: "partial_success",
				Data: &BatchNotificationResponse{
					TotalRequested: 3,
					TotalAccepted:      2,
					TotalRejected:    1,
					Results: []NotificationResponse{
						{ID: 1, Status: "sent"},
						{ID: 0, Status: "failed"},
						{ID: 3, Status: "sent"},
					},
					Errors: []string{"index 1: validation failed"},
				},
			},
			validateResult: func(t *testing.T, resp *BatchNotificationResponse) {
				assert.Equal(t, 2, resp.TotalAccepted)
				assert.Equal(t, 1, resp.TotalRejected)
				assert.Len(t, resp.Errors, 1)
			},
		},
		{
			name: "empty batch",
			request: &BatchNotificationRequest{
				Notifications: []NotificationRequest{},
			},
			responseStatus: http.StatusBadRequest,
			responseBody: ErrorResponse{
				Status:  "error",
				Message: "batch is empty",
			},
			wantErr: true,
		},
		{
			name: "batch with scheduling",
			request: &BatchNotificationRequest{
				Notifications: []NotificationRequest{
					{OrganizationID: 123, Subject: "Scheduled", Content: "Future notification"},
				},
			},
			responseStatus: http.StatusOK,
			responseBody: StandardResponse{
				Status: "success",
				Data: &BatchNotificationResponse{
					TotalRequested: 1,
					TotalAccepted:  1,
					TotalRejected:  0,
					Results: []NotificationResponse{
						{ID: 999, Status: "scheduled"},
					},
				},
			},
			validateResult: func(t *testing.T, resp *BatchNotificationResponse) {
				assert.Equal(t, 1, resp.TotalAccepted)
				assert.Equal(t, "scheduled", resp.Results[0].Status)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/v1/notifications/send/batch", r.URL.Path)

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL})
			result, err := client.Notifications.SendBatchNotifications(context.Background(), tt.request)

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

func TestNotificationsService_GetNotificationStatus(t *testing.T) {
	tests := []struct {
		name           string
		request        *NotificationStatusRequest
		responseStatus int
		responseBody   interface{}
		wantErr        bool
		validateResult func(*testing.T, *NotificationStatusResponse)
	}{
		{
			name: "get single notification status",
			request: &NotificationStatusRequest{
				NotificationIDs: []string{"notif-123"},
			},
			responseStatus: http.StatusOK,
			responseBody: StandardResponse{
				Status: "success",
				Data: &NotificationStatusResponse{
					Statuses: []NotificationStatus{
						{
							ID: "notif-123",
							Status:         "delivered",
							SentAt:         "2024-01-15T10:00:00Z",
							DeliveredAt:    "2024-01-15T10:00:05Z",
							Channels: map[string]ChannelStatus{
								"email": {Status: "delivered", DeliveredAt: "2024-01-15T10:00:05Z"},
							},
						},
					},
				},
			},
			validateResult: func(t *testing.T, resp *NotificationStatusResponse) {
				assert.Len(t, resp.Statuses, 1)
				assert.Equal(t, "delivered", resp.Statuses[0].Status)
				assert.Contains(t, resp.Statuses[0].Channels, "email")
			},
		},
		{
			name: "get multiple notification statuses",
			request: &NotificationStatusRequest{
				NotificationIDs: []string{"notif-1", "notif-2", "notif-3"},
			},
			responseStatus: http.StatusOK,
			responseBody: StandardResponse{
				Status: "success",
				Data: &NotificationStatusResponse{
					Statuses: []NotificationStatus{
						{ID: 1, Status: "sent"},
						{ID: 2, Status: "delivered"},
						{ID: 3, Status: "failed"},
					},
				},
			},
			validateResult: func(t *testing.T, resp *NotificationStatusResponse) {
				assert.Len(t, resp.Statuses, 3)
				assert.Equal(t, "sent", resp.Statuses[0].Status)
				assert.Equal(t, "delivered", resp.Statuses[1].Status)
				assert.Equal(t, "failed", resp.Statuses[2].Status)
			},
		},
		{
			name: "notification not found",
			request: &NotificationStatusRequest{
				NotificationIDs: []string{"nonexistent"},
			},
			responseStatus: http.StatusNotFound,
			responseBody: ErrorResponse{
				Status:  "error",
				Message: "notification not found",
			},
			wantErr: true,
		},
		{
			name: "get status with channel details",
			request: &NotificationStatusRequest{
				NotificationIDs:    []string{"notif-456"},
				IncludeChannelInfo: true,
			},
			responseStatus: http.StatusOK,
			responseBody: StandardResponse{
				Status: "success",
				Data: &NotificationStatusResponse{
					Statuses: []NotificationStatus{
						{
							ID: "notif-456",
							Status:         "partial",
							Channels: map[string]ChannelStatus{
								"email": {Status: "delivered", DeliveredAt: "2024-01-15T10:00:05Z"},
								"slack": {Status: "delivered", DeliveredAt: "2024-01-15T10:00:03Z"},
								"sms":   {Status: "failed", Error: "invalid phone number"},
							},
						},
					},
				},
			},
			validateResult: func(t *testing.T, resp *NotificationStatusResponse) {
				channels := resp.Statuses[0].Channels
				assert.Len(t, channels, 3)
				assert.Equal(t, "delivered", channels["email"].Status)
				assert.Equal(t, "failed", channels["sms"].Status)
				assert.NotEmpty(t, channels["sms"].Error)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/v1/notifications/status", r.URL.Path)

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL})
			result, err := client.Notifications.GetNotificationStatus(context.Background(), tt.request)

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

func TestNotificationsService_SendQuotaAlert(t *testing.T) {
	tests := []struct {
		name           string
		orgID          uint
		subject        string
		content        string
		priority       NotificationPriority
		metadata       map[string]interface{}
		responseStatus int
		responseBody   interface{}
		wantErr        bool
		validateResult func(*testing.T, *NotificationResponse)
	}{
		{
			name:     "quota warning alert",
			orgID:    123,
			subject:  "Quota Warning: 80% Usage",
			content:  "<p>Your organization has reached 80% of quota limit</p>",
			priority: NotificationPriorityNormal,
			metadata: map[string]interface{}{
				"quota_type":    "storage",
				"current_usage": 800,
				"quota_limit":   1000,
				"usage_percent": 80.0,
			},
			responseStatus: http.StatusOK,
			responseBody: StandardResponse{
				Status: "success",
				Data: &NotificationResponse{
					ID: "quota-alert-123",
					Status:         "sent",
					SentChannels:   []string{"email"},
				},
			},
			validateResult: func(t *testing.T, resp *NotificationResponse) {
				assert.Equal(t, "quota-alert-123", resp.NotificationID)
				assert.Equal(t, "sent", resp.Status)
			},
		},
		{
			name:     "quota critical alert",
			orgID:    456,
			subject:  "Quota Critical: 95% Usage",
			content:  "<p>URGENT: Your organization has reached 95% of quota limit</p>",
			priority: NotificationPriorityCritical,
			metadata: map[string]interface{}{
				"quota_type":    "cpu",
				"current_usage": 950,
				"quota_limit":   1000,
				"usage_percent": 95.0,
			},
			responseStatus: http.StatusOK,
			responseBody: StandardResponse{
				Status: "success",
				Data: &NotificationResponse{
					ID: "quota-alert-456",
					Status:         "sent",
					SentChannels:   []string{"email", "slack"},
				},
			},
			validateResult: func(t *testing.T, resp *NotificationResponse) {
				assert.Len(t, resp.SentChannels, 2)
			},
		},
		{
			name:           "quota exceeded alert",
			orgID:          789,
			subject:        "Quota Exceeded: 100% Usage",
			content:        "<p>Your organization has exceeded the quota limit</p>",
			priority:       NotificationPriorityCritical,
			metadata:       map[string]interface{}{"quota_type": "memory"},
			responseStatus: http.StatusOK,
			responseBody: StandardResponse{
				Status: "success",
				Data: &NotificationResponse{
					ID: "quota-alert-789",
					Status:         "sent",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/v1/notifications/send", r.URL.Path)

				// Verify request body
				var req NotificationRequest
				err := json.NewDecoder(r.Body).Decode(&req)
				require.NoError(t, err)
				assert.Equal(t, tt.orgID, req.OrganizationID)
				assert.Equal(t, tt.subject, req.Subject)
				assert.Equal(t, tt.content, req.Content)
				assert.Equal(t, tt.priority, req.Priority)
				assert.Equal(t, "html", req.ContentType)

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL})
			result, err := client.Notifications.SendQuotaAlert(
				context.Background(),
				tt.orgID,
				tt.subject,
				tt.content,
				tt.priority,
				tt.metadata,
			)

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

// =============================================================================
// Integration and Edge Cases
// =============================================================================

func TestNotificationsService_NotificationWorkflow(t *testing.T) {
	// Test complete notification workflow: send -> get status
	notificationID := "workflow-notif-123"

	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		switch r.URL.Path {
		case "/v1/notifications/send":
			json.NewEncoder(w).Encode(StandardResponse{
				Status: "success",
				Data: &NotificationResponse{
					ID: notificationID,
					Status:         "sent",
				},
			})

		case "/v1/notifications/status":
			json.NewEncoder(w).Encode(StandardResponse{
				Status: "success",
				Data: &NotificationStatusResponse{
					Statuses: []NotificationStatus{
						{
							ID: notificationID,
							Status:         "delivered",
							SentAt:         "2024-01-15T10:00:00Z",
							DeliveredAt:    "2024-01-15T10:00:05Z",
						},
					},
				},
			})
		}
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	// Step 1: Send notification
	sendResp, err := client.Notifications.SendNotification(context.Background(), &NotificationRequest{
		OrganizationID: 123,
		Subject:        "Test",
		Content:        "Content",
		Priority:       NotificationPriorityNormal,
	})
	require.NoError(t, err)
	assert.Equal(t, notificationID, sendResp.NotificationID)

	// Step 2: Get status
	statusResp, err := client.Notifications.GetNotificationStatus(context.Background(), &NotificationStatusRequest{
		NotificationIDs: []string{notificationID},
	})
	require.NoError(t, err)
	assert.Len(t, statusResp.Statuses, 1)
	assert.Equal(t, "delivered", statusResp.Statuses[0].Status)

	assert.Equal(t, 2, callCount)
}

func TestNotificationsService_ConcurrentNotifications(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(StandardResponse{
			Status: "success",
			Data: &NotificationResponse{
				ID: "concurrent-notif",
				Status:         "sent",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	done := make(chan bool)
	for i := 0; i < 5; i++ {
		go func(index int) {
			_, err := client.Notifications.SendNotification(context.Background(), &NotificationRequest{
				OrganizationID: uint(100 + index),
				Subject:        "Concurrent Test",
				Content:        "Content",
				Priority:       NotificationPriorityLow,
			})
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	for i := 0; i < 5; i++ {
		<-done
	}
}

func TestNotificationsService_LargeBatch(t *testing.T) {
	batchSize := 100
	notifications := make([]NotificationRequest, batchSize)
	for i := 0; i < batchSize; i++ {
		notifications[i] = NotificationRequest{
			OrganizationID: uint(i + 1),
			Subject:        "Batch Test",
			Content:        "Content",
			Priority:       NotificationPriorityLow,
		}
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req BatchNotificationRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Len(t, req.Notifications, batchSize)

		results := make([]NotificationResponse, batchSize)
		for i := 0; i < batchSize; i++ {
			results[i] = NotificationResponse{
				ID: "batch-notif-" + string(rune(i)),
				Status:         "sent",
			}
		}

		json.NewEncoder(w).Encode(StandardResponse{
			Status: "success",
			Data: &BatchNotificationResponse{
				TotalRequested: batchSize,
				TotalAccepted:      batchSize,
				Results:        results,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, err := client.Notifications.SendBatchNotifications(context.Background(), &BatchNotificationRequest{
		Notifications: notifications,
	})

	require.NoError(t, err)
	assert.Equal(t, batchSize, result.TotalSent)
	assert.Len(t, result.Results, batchSize)
}

func TestNotificationsService_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
			return
		}
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.Notifications.SendNotification(ctx, &NotificationRequest{
		OrganizationID: 123,
		Subject:        "Test",
		Content:        "Content",
	})

	assert.Error(t, err)
}

func TestNotificationsService_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  "error",
			Message: "validation failed",
			Error:   "subject is required",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	_, err := client.Notifications.SendNotification(context.Background(), &NotificationRequest{
		OrganizationID: 123,
		Content:        "Content without subject",
	})

	assert.Error(t, err)
}
