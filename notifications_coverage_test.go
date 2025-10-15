package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotificationsService_SendNotification(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/notifications/send", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"notification_id": "notif-123",
				"status":          "sent",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	req := &NotificationRequest{
		OrganizationID: 1,
		Subject:        "Test Notification",
		Content:        "This is a test",
		ContentType:    "text",
		Priority:       NotificationPriorityHigh,
	}
	response, err := client.Notifications.SendNotification(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, response)
}

func TestNotificationsService_SendNotification_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Invalid notification request",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	req := &NotificationRequest{}
	response, err := client.Notifications.SendNotification(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestNotificationsService_SendBatchNotifications(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/notifications/send/batch", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"total":      2,
				"successful": 2,
				"failed":     0,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	req := &BatchNotificationRequest{
		Notifications: []NotificationRequest{
			{Subject: "Test 1", Content: "Content 1"},
			{Subject: "Test 2", Content: "Content 2"},
		},
	}
	response, err := client.Notifications.SendBatchNotifications(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, response)
}

func TestNotificationsService_SendBatchNotifications_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Batch send failed",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	req := &BatchNotificationRequest{}
	response, err := client.Notifications.SendBatchNotifications(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestNotificationsService_GetNotificationStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/notifications/status", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"notification_id": "notif-123",
				"status":          "delivered",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	req := &NotificationStatusRequest{
		NotificationIDs: []uint{123},
	}
	response, err := client.Notifications.GetNotificationStatus(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, response)
}

func TestNotificationsService_GetNotificationStatus_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Notification not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	req := &NotificationStatusRequest{NotificationIDs: []uint{999}}
	response, err := client.Notifications.GetNotificationStatus(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestNotificationsService_SendQuotaAlert(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/notifications/send", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"notification_id": "quota-alert-123",
				"status":          "sent",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	metadata := map[string]interface{}{
		"quota_type": "cpu",
		"usage":      "95%",
	}
	response, err := client.Notifications.SendQuotaAlert(
		context.Background(),
		1,
		"Quota Alert",
		"<p>Quota exceeded</p>",
		NotificationPriorityCritical,
		metadata,
	)
	assert.NoError(t, err)
	assert.NotNil(t, response)
}

func TestNotificationsService_SendQuotaAlert_WithNilMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   map[string]interface{}{"notification_id": "alert-456"},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	response, err := client.Notifications.SendQuotaAlert(
		context.Background(),
		1,
		"Simple Alert",
		"Content",
		NotificationPriorityNormal,
		nil,
	)
	assert.NoError(t, err)
	assert.NotNil(t, response)
}

func TestNotificationsService_AllMethods_NetworkError(t *testing.T) {
	client, _ := NewClient(&Config{BaseURL: "http://invalid-server:9999"})

	_, err := client.Notifications.SendNotification(context.Background(), &NotificationRequest{})
	assert.Error(t, err)

	_, err = client.Notifications.SendBatchNotifications(context.Background(), &BatchNotificationRequest{})
	assert.Error(t, err)

	_, err = client.Notifications.GetNotificationStatus(context.Background(), &NotificationStatusRequest{})
	assert.Error(t, err)

	_, err = client.Notifications.SendQuotaAlert(context.Background(), 1, "Test", "Test", NotificationPriorityLow, nil)
	assert.Error(t, err)
}
