package nexmonyx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Error path tests to improve coverage from 87.5% to 100%
// Note: Like other services, the type assertion lines may be unreachable
// since JSON unmarshaling validates types first.

func TestAlertsService_Create_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/alerts/rules", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Invalid alert configuration",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	alert, err := client.Alerts.Create(context.Background(), &Alert{
		Name:      "Test Alert",
		Condition: "cpu > 90",
	})
	assert.Error(t, err)
	assert.Nil(t, alert)
}

func TestAlertsService_Get_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/alerts/rules/999", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Alert not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	alert, err := client.Alerts.Get(context.Background(), "999")
	assert.Error(t, err)
	assert.Nil(t, alert)
}

func TestAlertsService_Update_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/alerts/rules/1", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Invalid update data",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	alert, err := client.Alerts.Update(context.Background(), "1", &Alert{
		Name:      "Updated Alert",
		Condition: "invalid",
	})
	assert.Error(t, err)
	assert.Nil(t, alert)
}

func TestAlertsService_Enable_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/alerts/999/enable", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Alert not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	alert, err := client.Alerts.Enable(context.Background(), "999")
	assert.Error(t, err)
	assert.Nil(t, alert)
}

func TestAlertsService_Disable_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/alerts/999/disable", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Alert not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	alert, err := client.Alerts.Disable(context.Background(), "999")
	assert.Error(t, err)
	assert.Nil(t, alert)
}

func TestAlertsService_Test_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/alerts/1/test", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Cannot test alert",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, err := client.Alerts.Test(context.Background(), "1")
	assert.Error(t, err)
	assert.Nil(t, result)
}

// Edge case tests for List and GetHistory with nil options

func TestAlertsService_List_NilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/alerts/rules", r.URL.Path)

		// Verify no query parameters when opts is nil
		assert.Empty(t, r.URL.Query().Get("page"))
		assert.Empty(t, r.URL.Query().Get("limit"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []*Alert{},
			"meta": PaginationMeta{
				Page:       1,
				Limit:      25,
				TotalItems: 0,
				TotalPages: 0,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	alerts, meta, err := client.Alerts.List(context.Background(), nil)
	assert.NoError(t, err)
	assert.NotNil(t, alerts)
	assert.NotNil(t, meta)
	assert.Len(t, alerts, 0)
}

func TestAlertsService_GetHistory_NilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/alerts/1/history", r.URL.Path)

		// Verify no query parameters when opts is nil
		assert.Empty(t, r.URL.Query().Get("page"))
		assert.Empty(t, r.URL.Query().Get("limit"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []*AlertHistoryEntry{},
			"meta": PaginationMeta{
				Page:       1,
				Limit:      25,
				TotalItems: 0,
				TotalPages: 0,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	history, meta, err := client.Alerts.GetHistory(context.Background(), "1", nil)
	assert.NoError(t, err)
	assert.NotNil(t, history)
	assert.NotNil(t, meta)
	assert.Len(t, history, 0)
}

func TestAlertsService_ListChannels_NilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/alerts/channels", r.URL.Path)

		// Verify no query parameters when opts is nil
		assert.Empty(t, r.URL.Query().Get("page"))
		assert.Empty(t, r.URL.Query().Get("limit"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []*AlertChannel{},
			"meta": PaginationMeta{
				Page:       1,
				Limit:      25,
				TotalItems: 0,
				TotalPages: 0,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	channels, meta, err := client.Alerts.ListChannels(context.Background(), nil)
	assert.NoError(t, err)
	assert.NotNil(t, channels)
	assert.NotNil(t, meta)
	assert.Len(t, channels, 0)
}

// Success path tests with various configurations

func TestAlertsService_Create_Comprehensive(t *testing.T) {
	tests := []struct {
		name       string
		alert      *Alert
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Alert)
	}{
		{
			name: "success - CPU alert",
			alert: &Alert{
				Name:      "CPU Alert",
				Condition: "cpu > 90",
				Enabled:   true,
				Severity:  "critical",
			},
			mockStatus: http.StatusCreated,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":        1,
					"name":      "CPU Alert",
					"condition": "cpu > 90",
					"enabled":   true,
					"severity":  "critical",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alert *Alert) {
				assert.Equal(t, "CPU Alert", alert.Name)
				assert.Equal(t, "cpu > 90", alert.Condition)
				assert.True(t, alert.Enabled)
			},
		},
		{
			name: "success - memory alert with threshold",
			alert: &Alert{
				Name:      "Memory Alert",
				Condition: "memory > 85",
				Enabled:   true,
				Severity:  "warning",
			},
			mockStatus: http.StatusCreated,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":        2,
					"name":      "Memory Alert",
					"condition": "memory > 85",
					"enabled":   true,
					"severity":  "warning",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alert *Alert) {
				assert.Equal(t, "Memory Alert", alert.Name)
				assert.Equal(t, "warning", alert.Severity)
			},
		},
		{
			name: "validation error - missing name",
			alert: &Alert{
				Condition: "cpu > 90",
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Alert name is required",
			},
			wantErr: true,
		},
		{
			name: "validation error - invalid condition",
			alert: &Alert{
				Name:      "Test Alert",
				Condition: "invalid syntax",
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Invalid alert condition syntax",
			},
			wantErr: true,
		},
		{
			name: "unauthorized",
			alert: &Alert{
				Name:      "CPU Alert",
				Condition: "cpu > 90",
			},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Unauthorized",
			},
			wantErr: true,
		},
		{
			name: "forbidden",
			alert: &Alert{
				Name:      "CPU Alert",
				Condition: "cpu > 90",
			},
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Insufficient permissions",
			},
			wantErr: true,
		},
		{
			name: "server error",
			alert: &Alert{
				Name:      "CPU Alert",
				Condition: "cpu > 90",
			},
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/v1/alerts/rules", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL:    server.URL,
				RetryCount: 0,
			})

			result, err := client.Alerts.Create(context.Background(), tt.alert)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

func TestAlertsService_Get_Comprehensive(t *testing.T) {
	tests := []struct {
		name       string
		alertID    string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Alert)
	}{
		{
			name:       "success - CPU alert",
			alertID:    "1",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":        1,
					"name":      "CPU Alert",
					"condition": "cpu > 90",
					"enabled":   true,
					"severity":  "critical",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alert *Alert) {
				assert.Equal(t, uint(1), alert.ID)
				assert.Equal(t, "CPU Alert", alert.Name)
				assert.True(t, alert.Enabled)
			},
		},
		{
			name:       "success - disk alert",
			alertID:    "2",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":        2,
					"name":      "Disk Alert",
					"condition": "disk > 80",
					"enabled":   false,
					"severity":  "warning",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alert *Alert) {
				assert.Equal(t, "Disk Alert", alert.Name)
				assert.Equal(t, "warning", alert.Severity)
				assert.False(t, alert.Enabled)
			},
		},
		{
			name:       "not found",
			alertID:    "999",
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Alert not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			alertID:    "1",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Unauthorized",
			},
			wantErr: true,
		},
		{
			name:       "forbidden",
			alertID:    "1",
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Insufficient permissions",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			alertID:    "1",
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, fmt.Sprintf("/v1/alerts/rules/%s", tt.alertID), r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL:    server.URL,
				RetryCount: 0,
			})

			result, err := client.Alerts.Get(context.Background(), tt.alertID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

func TestAlertsService_List_Comprehensive(t *testing.T) {
	tests := []struct {
		name       string
		options    *ListOptions
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, []*Alert, *PaginationMeta)
	}{
		{
			name:       "success - multiple alerts",
			options:    &ListOptions{Page: 1, Limit: 10},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": []map[string]interface{}{
					{"id": 1, "name": "CPU Alert", "condition": "cpu > 90", "enabled": true},
					{"id": 2, "name": "Memory Alert", "condition": "memory > 85", "enabled": true},
					{"id": 3, "name": "Disk Alert", "condition": "disk > 80", "enabled": false},
				},
				"meta": map[string]interface{}{
					"total_items": 3,
					"total_pages": 1,
					"page":        1,
					"limit":       10,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alerts []*Alert, meta *PaginationMeta) {
				assert.Len(t, alerts, 3)
				assert.NotNil(t, meta)
				assert.Equal(t, 3, meta.TotalItems)
				assert.Equal(t, "CPU Alert", alerts[0].Name)
			},
		},
		{
			name:       "success - empty list",
			options:    &ListOptions{Page: 1, Limit: 25},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": []map[string]interface{}{},
				"meta": map[string]interface{}{
					"total_items": 0,
					"total_pages": 0,
					"page":        1,
					"limit":       25,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alerts []*Alert, meta *PaginationMeta) {
				assert.Len(t, alerts, 0)
				assert.Equal(t, 0, meta.TotalItems)
			},
		},
		{
			name:       "success - with pagination",
			options:    &ListOptions{Page: 2, Limit: 25},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": []map[string]interface{}{
					{"id": 26, "name": "Alert 26", "enabled": true},
				},
				"meta": map[string]interface{}{
					"total_items": 50,
					"total_pages": 2,
					"page":        2,
					"limit":       25,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alerts []*Alert, meta *PaginationMeta) {
				assert.Len(t, alerts, 1)
				assert.Equal(t, 2, meta.Page)
			},
		},
		{
			name:       "unauthorized",
			options:    &ListOptions{Page: 1, Limit: 25},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Unauthorized",
			},
			wantErr: true,
		},
		{
			name:       "forbidden",
			options:    &ListOptions{Page: 1, Limit: 25},
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Insufficient permissions",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			options:    &ListOptions{Page: 1, Limit: 25},
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/v1/alerts/rules", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL:    server.URL,
				RetryCount: 0,
			})

			result, meta, err := client.Alerts.List(context.Background(), tt.options)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.NotNil(t, meta)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result, meta)
				}
			}
		})
	}
}

func TestAlertsService_Update_Comprehensive(t *testing.T) {
	tests := []struct {
		name       string
		alertID    string
		alert      *Alert
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Alert)
	}{
		{
			name:    "success - update name and condition",
			alertID: "1",
			alert: &Alert{
				Name:      "Updated CPU Alert",
				Condition: "cpu > 95",
				Enabled:   true,
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":        1,
					"name":      "Updated CPU Alert",
					"condition": "cpu > 95",
					"enabled":   true,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alert *Alert) {
				assert.Equal(t, uint(1), alert.ID)
				assert.Equal(t, "Updated CPU Alert", alert.Name)
				assert.Equal(t, "cpu > 95", alert.Condition)
			},
		},
		{
			name:    "success - disable alert",
			alertID: "2",
			alert: &Alert{
				Enabled: false,
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"data": map[string]interface{}{
					"id":      2,
					"enabled": false,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alert *Alert) {
				assert.False(t, alert.Enabled)
			},
		},
		{
			name:    "not found",
			alertID: "999",
			alert: &Alert{
				Name: "Updated Alert",
			},
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Alert not found",
			},
			wantErr: true,
		},
		{
			name:    "validation error",
			alertID: "1",
			alert: &Alert{
				Condition: "invalid syntax",
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Invalid alert condition",
			},
			wantErr: true,
		},
		{
			name:    "unauthorized",
			alertID: "1",
			alert: &Alert{
				Name: "Updated Alert",
			},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Unauthorized",
			},
			wantErr: true,
		},
		{
			name:    "forbidden",
			alertID: "1",
			alert: &Alert{
				Name: "Updated Alert",
			},
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Insufficient permissions",
			},
			wantErr: true,
		},
		{
			name:    "server error",
			alertID: "1",
			alert: &Alert{
				Name: "Updated Alert",
			},
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "PUT", r.Method)
				assert.Equal(t, fmt.Sprintf("/v1/alerts/rules/%s", tt.alertID), r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL:    server.URL,
				RetryCount: 0,
			})

			result, err := client.Alerts.Update(context.Background(), tt.alertID, tt.alert)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

func TestAlertsService_Enable_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/alerts/1/enable", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id":      1,
				"enabled": true,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	alert, err := client.Alerts.Enable(context.Background(), "1")
	assert.NoError(t, err)
	assert.NotNil(t, alert)
}

func TestAlertsService_Disable_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/alerts/1/disable", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id":      1,
				"enabled": false,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	alert, err := client.Alerts.Disable(context.Background(), "1")
	assert.NoError(t, err)
	assert.NotNil(t, alert)
}

func TestAlertsService_Test_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/alerts/1/test", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"success":   true,
				"triggered": true,
				"message":   "Alert would trigger",
				"value":     95.5,
				"threshold": 90.0,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, err := client.Alerts.Test(context.Background(), "1")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.True(t, result.Triggered)
}
