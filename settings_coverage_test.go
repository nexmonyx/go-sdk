package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSettingsService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/v1/organizations/")
		assert.Contains(t, r.URL.Path, "/settings")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"organization_id": 1,
				"general":         map[string]interface{}{"timezone": "UTC"},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	settings, err := client.Settings.Get(context.Background(), "org-123")
	assert.NoError(t, err)
	assert.NotNil(t, settings)
}

func TestSettingsService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Contains(t, r.URL.Path, "/v1/organizations/")
		assert.Contains(t, r.URL.Path, "/settings")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"organization_id": 1,
				"general":         map[string]interface{}{"timezone": "America/New_York"},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	settings := &Settings{
		OrganizationID: 1,
		GeneralSettings: &GeneralSettings{
			TimeZone: "America/New_York",
		},
	}
	updated, err := client.Settings.Update(context.Background(), "org-123", settings)
	assert.NoError(t, err)
	assert.NotNil(t, updated)
}

func TestSettingsService_GetNotificationSettings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/settings/notifications")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"email_enabled": true,
				"slack_enabled": false,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	settings, err := client.Settings.GetNotificationSettings(context.Background(), "org-123")
	assert.NoError(t, err)
	assert.NotNil(t, settings)
}

func TestSettingsService_UpdateNotificationSettings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Contains(t, r.URL.Path, "/settings/notifications")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"email_enabled": true,
				"slack_enabled": true,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	settings := &NotificationSettings{
		EmailEnabled: true,
		SlackEnabled: true,
	}
	updated, err := client.Settings.UpdateNotificationSettings(context.Background(), "org-123", settings)
	assert.NoError(t, err)
	assert.NotNil(t, updated)
}

func TestSettingsService_Get_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Organization not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	settings, err := client.Settings.Get(context.Background(), "invalid-org")
	assert.Error(t, err)
	assert.Nil(t, settings)
}

func TestSettingsService_Update_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Invalid settings data",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	settings := &Settings{OrganizationID: 1}
	updated, err := client.Settings.Update(context.Background(), "org-123", settings)
	assert.Error(t, err)
	assert.Nil(t, updated)
}

func TestSettingsService_NotificationSettings_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Server error",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	_, err := client.Settings.GetNotificationSettings(context.Background(), "org-123")
	assert.Error(t, err)

	notifSettings := &NotificationSettings{}
	_, err = client.Settings.UpdateNotificationSettings(context.Background(), "org-123", notifSettings)
	assert.Error(t, err)
}

func TestSettingsService_NetworkError(t *testing.T) {
	client, _ := NewClient(&Config{BaseURL: "http://invalid-server:9999"})

	_, err := client.Settings.Get(context.Background(), "org-123")
	assert.Error(t, err)

	_, err = client.Settings.Update(context.Background(), "org-123", &Settings{})
	assert.Error(t, err)

	_, err = client.Settings.GetNotificationSettings(context.Background(), "org-123")
	assert.Error(t, err)

	_, err = client.Settings.UpdateNotificationSettings(context.Background(), "org-123", &NotificationSettings{})
	assert.Error(t, err)
}
