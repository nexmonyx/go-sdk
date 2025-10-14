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

// TestAlertsService_Create tests the Create method
func TestAlertsService_Create(t *testing.T) {
	tests := []struct {
		name       string
		alert      *Alert
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Alert)
	}{
		{
			name: "successful create",
			alert: &Alert{
				Name:        "High CPU Alert",
				Description: "Alert when CPU usage exceeds threshold",
				Type:        "metric",
				MetricName:  "cpu.usage",
				Condition:   "gt",
				Threshold:   80.0,
				Duration:    300,
				Frequency:   60,
				Severity:    "warning",
				Enabled:     true,
			},
			mockStatus: http.StatusCreated,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Alert created successfully",
				Data: &Alert{
					GormModel: GormModel{
						ID: 1,
					},
					Name:        "High CPU Alert",
					Description: "Alert when CPU usage exceeds threshold",
					Type:        "metric",
					MetricName:  "cpu.usage",
					Condition:   "gt",
					Threshold:   80.0,
					Duration:    300,
					Frequency:   60,
					Severity:    "warning",
					Enabled:     true,
					Status:      "active",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alert *Alert) {
				assert.NotNil(t, alert)
				assert.Equal(t, uint(1), alert.ID)
				assert.Equal(t, "High CPU Alert", alert.Name)
				assert.Equal(t, "metric", alert.Type)
				assert.Equal(t, "cpu.usage", alert.MetricName)
				assert.Equal(t, 80.0, alert.Threshold)
				assert.Equal(t, "active", alert.Status)
			},
		},
		{
			name: "validation error",
			alert: &Alert{
				Name: "", // Empty name should cause validation error
			},
			mockStatus: http.StatusBadRequest,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Validation failed",
				Error:   "Alert name is required",
			},
			wantErr: true,
		},
		{
			name: "unauthorized",
			alert: &Alert{
				Name:       "Test Alert",
				Type:       "metric",
				MetricName: "cpu.usage",
			},
			mockStatus: http.StatusUnauthorized,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Unauthorized",
			},
			wantErr: true,
		},
		{
			name: "server error",
			alert: &Alert{
				Name:       "Test Alert",
				Type:       "metric",
				MetricName: "cpu.usage",
			},
			mockStatus: http.StatusInternalServerError,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method and path
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/v1/alerts/rules", r.URL.Path)

				// Verify request body if provided
				if tt.alert != nil {
					var receivedAlert Alert
					err := json.NewDecoder(r.Body).Decode(&receivedAlert)
					require.NoError(t, err)
					assert.Equal(t, tt.alert.Name, receivedAlert.Name)
					assert.Equal(t, tt.alert.Type, receivedAlert.Type)
				}

				// Send mock response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			// Create client
			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			// Call Create
			result, err := client.Alerts.Create(context.Background(), tt.alert)

			// Check error
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestAlertsService_Get tests the Get method
func TestAlertsService_Get(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Alert)
	}{
		{
			name:       "successful get",
			id:         "1",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status: "success",
				Data: &Alert{
					GormModel: GormModel{
						ID: 1,
					},
					Name:        "High CPU Alert",
					Description: "Alert when CPU usage exceeds threshold",
					Type:        "metric",
					MetricName:  "cpu.usage",
					Condition:   "gt",
					Threshold:   80.0,
					Enabled:     true,
					Status:      "active",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alert *Alert) {
				assert.NotNil(t, alert)
				assert.Equal(t, uint(1), alert.ID)
				assert.Equal(t, "High CPU Alert", alert.Name)
				assert.Equal(t, "metric", alert.Type)
				assert.Equal(t, "active", alert.Status)
			},
		},
		{
			name:       "alert not found",
			id:         "999",
			mockStatus: http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Alert not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			id:         "1",
			mockStatus: http.StatusUnauthorized,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Unauthorized",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method and path
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/alerts/rules/")

				// Send mock response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			// Create client
			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			// Call Get
			result, err := client.Alerts.Get(context.Background(), tt.id)

			// Check error
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestAlertsService_List tests the List method
func TestAlertsService_List(t *testing.T) {
	tests := []struct {
		name       string
		opts       *ListOptions
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, []*Alert, *PaginationMeta)
	}{
		{
			name: "successful list with pagination",
			opts: &ListOptions{
				Page:  1,
				Limit: 10,
			},
			mockStatus: http.StatusOK,
			mockBody: PaginatedResponse{
				Status: "success",
				Data: []*Alert{
					{
						GormModel:  GormModel{ID: 1},
						Name:       "High CPU Alert",
						Type:       "metric",
						MetricName: "cpu.usage",
						Enabled:    true,
					},
					{
						GormModel:  GormModel{ID: 2},
						Name:       "Low Memory Alert",
						Type:       "metric",
						MetricName: "memory.available",
						Enabled:    true,
					},
				},
				Meta: &PaginationMeta{
					Page:       1,
					Limit:      10,
					TotalItems: 2,
					TotalPages: 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alerts []*Alert, meta *PaginationMeta) {
				assert.NotNil(t, alerts)
				assert.Len(t, alerts, 2)
				assert.Equal(t, "High CPU Alert", alerts[0].Name)
				assert.Equal(t, "Low Memory Alert", alerts[1].Name)
				assert.NotNil(t, meta)
				assert.Equal(t, 1, meta.Page)
				assert.Equal(t, 2, meta.TotalItems)
			},
		},
		{
			name: "list with search",
			opts: &ListOptions{
				Search: "CPU",
			},
			mockStatus: http.StatusOK,
			mockBody: PaginatedResponse{
				Status: "success",
				Data: []*Alert{
					{
						GormModel:  GormModel{ID: 1},
						Name:       "High CPU Alert",
						Type:       "metric",
						MetricName: "cpu.usage",
					},
				},
				Meta: &PaginationMeta{
					Page:       1,
					Limit:      25,
					TotalItems: 1,
					TotalPages: 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alerts []*Alert, meta *PaginationMeta) {
				assert.NotNil(t, alerts)
				assert.Len(t, alerts, 1)
				assert.Contains(t, alerts[0].Name, "CPU")
			},
		},
		{
			name:       "empty list",
			opts:       nil,
			mockStatus: http.StatusOK,
			mockBody: PaginatedResponse{
				Status: "success",
				Data:   []*Alert{},
				Meta: &PaginationMeta{
					Page:       1,
					Limit:      25,
					TotalItems: 0,
					TotalPages: 0,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alerts []*Alert, meta *PaginationMeta) {
				assert.NotNil(t, alerts)
				assert.Len(t, alerts, 0)
			},
		},
		{
			name:       "unauthorized",
			opts:       nil,
			mockStatus: http.StatusUnauthorized,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Unauthorized",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method and path
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/v1/alerts/rules", r.URL.Path)

				// Verify query parameters
				if tt.opts != nil {
					query := r.URL.Query()
					if tt.opts.Search != "" {
						assert.Equal(t, tt.opts.Search, query.Get("search"))
					}
				}

				// Send mock response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			// Create client
			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			// Call List
			result, meta, err := client.Alerts.List(context.Background(), tt.opts)

			// Check error
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Nil(t, meta)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result, meta)
				}
			}
		})
	}
}

// TestAlertsService_Update tests the Update method
func TestAlertsService_Update(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		alert      *Alert
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Alert)
	}{
		{
			name: "successful update",
			id:   "1",
			alert: &Alert{
				Name:      "Updated CPU Alert",
				Threshold: 90.0,
				Enabled:   true,
			},
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Alert updated successfully",
				Data: &Alert{
					GormModel: GormModel{ID: 1},
					Name:      "Updated CPU Alert",
					Threshold: 90.0,
					Enabled:   true,
					Status:    "active",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alert *Alert) {
				assert.NotNil(t, alert)
				assert.Equal(t, "Updated CPU Alert", alert.Name)
				assert.Equal(t, 90.0, alert.Threshold)
			},
		},
		{
			name: "alert not found",
			id:   "999",
			alert: &Alert{
				Name: "Updated Alert",
			},
			mockStatus: http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Alert not found",
			},
			wantErr: true,
		},
		{
			name: "validation error",
			id:   "1",
			alert: &Alert{
				Threshold: -1.0, // Invalid threshold
			},
			mockStatus: http.StatusBadRequest,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Validation failed",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method and path
				assert.Equal(t, "PUT", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/alerts/rules/")

				// Verify request body
				if tt.alert != nil {
					var receivedAlert Alert
					err := json.NewDecoder(r.Body).Decode(&receivedAlert)
					require.NoError(t, err)
				}

				// Send mock response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			// Create client
			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			// Call Update
			result, err := client.Alerts.Update(context.Background(), tt.id, tt.alert)

			// Check error
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestAlertsService_Delete tests the Delete method
func TestAlertsService_Delete(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:       "successful delete",
			id:         "1",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Alert deleted successfully",
			},
			wantErr: false,
		},
		{
			name:       "alert not found",
			id:         "999",
			mockStatus: http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Alert not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			id:         "1",
			mockStatus: http.StatusUnauthorized,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Unauthorized",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method and path
				assert.Equal(t, "DELETE", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/alerts/rules/")

				// Send mock response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			// Create client
			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			// Call Delete
			err = client.Alerts.Delete(context.Background(), tt.id)

			// Check error
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestAlertsService_Enable tests the Enable method
func TestAlertsService_Enable(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Alert)
	}{
		{
			name:       "successful enable",
			id:         "1",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Alert enabled successfully",
				Data: &Alert{
					GormModel: GormModel{ID: 1},
					Name:      "Test Alert",
					Enabled:   true,
					Status:    "active",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alert *Alert) {
				assert.NotNil(t, alert)
				assert.True(t, alert.Enabled)
				assert.Equal(t, "active", alert.Status)
			},
		},
		{
			name:       "alert not found",
			id:         "999",
			mockStatus: http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Alert not found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method and path
				assert.Equal(t, "POST", r.Method)
				assert.Contains(t, r.URL.Path, "/enable")

				// Send mock response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			// Create client
			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			// Call Enable
			result, err := client.Alerts.Enable(context.Background(), tt.id)

			// Check error
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestAlertsService_Disable tests the Disable method
func TestAlertsService_Disable(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Alert)
	}{
		{
			name:       "successful disable",
			id:         "1",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Alert disabled successfully",
				Data: &Alert{
					GormModel: GormModel{ID: 1},
					Name:      "Test Alert",
					Enabled:   false,
					Status:    "disabled",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, alert *Alert) {
				assert.NotNil(t, alert)
				assert.False(t, alert.Enabled)
				assert.Equal(t, "disabled", alert.Status)
			},
		},
		{
			name:       "alert not found",
			id:         "999",
			mockStatus: http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Alert not found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method and path
				assert.Equal(t, "POST", r.Method)
				assert.Contains(t, r.URL.Path, "/disable")

				// Send mock response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			// Create client
			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			// Call Disable
			result, err := client.Alerts.Disable(context.Background(), tt.id)

			// Check error
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestAlertsService_GetHistory tests the GetHistory method
func TestAlertsService_GetHistory(t *testing.T) {
	now := time.Now()
	customNow := &CustomTime{Time: now}

	tests := []struct {
		name       string
		id         string
		opts       *ListOptions
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, []*AlertHistoryEntry, *PaginationMeta)
	}{
		{
			name: "successful get history",
			id:   "1",
			opts: &ListOptions{
				Page:  1,
				Limit: 10,
			},
			mockStatus: http.StatusOK,
			mockBody: PaginatedResponse{
				Status: "success",
				Data: []*AlertHistoryEntry{
					{
						ID:          1,
						AlertID:     1,
						TriggeredAt: customNow,
						ResolvedAt:  customNow,
						Status:      "resolved",
						Value:       85.5,
						Threshold:   80.0,
						Message:     "CPU usage exceeded threshold",
					},
					{
						ID:          2,
						AlertID:     1,
						TriggeredAt: customNow,
						Status:      "triggered",
						Value:       82.0,
						Threshold:   80.0,
						Message:     "CPU usage exceeded threshold",
					},
				},
				Meta: &PaginationMeta{
					Page:       1,
					Limit:      10,
					TotalItems: 2,
					TotalPages: 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, history []*AlertHistoryEntry, meta *PaginationMeta) {
				assert.NotNil(t, history)
				assert.Len(t, history, 2)
				assert.Equal(t, "resolved", history[0].Status)
				assert.Equal(t, "triggered", history[1].Status)
				assert.NotNil(t, meta)
			},
		},
		{
			name:       "empty history",
			id:         "1",
			opts:       nil,
			mockStatus: http.StatusOK,
			mockBody: PaginatedResponse{
				Status: "success",
				Data:   []*AlertHistoryEntry{},
				Meta: &PaginationMeta{
					Page:       1,
					Limit:      25,
					TotalItems: 0,
					TotalPages: 0,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, history []*AlertHistoryEntry, meta *PaginationMeta) {
				assert.NotNil(t, history)
				assert.Len(t, history, 0)
			},
		},
		{
			name:       "alert not found",
			id:         "999",
			opts:       nil,
			mockStatus: http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Alert not found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method and path
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/history")

				// Send mock response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			// Create client
			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			// Call GetHistory
			result, meta, err := client.Alerts.GetHistory(context.Background(), tt.id, tt.opts)

			// Check error
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Nil(t, meta)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result, meta)
				}
			}
		})
	}
}

// TestAlertsService_Test tests the Test method
func TestAlertsService_Test(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *AlertTestResult)
	}{
		{
			name:       "successful test - triggered",
			id:         "1",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Alert test completed",
				Data: &AlertTestResult{
					Success:   true,
					Triggered: true,
					Message:   "Alert would be triggered",
					Value:     85.5,
					Threshold: 80.0,
					Details: map[string]interface{}{
						"metric": "cpu.usage",
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, result *AlertTestResult) {
				assert.NotNil(t, result)
				assert.True(t, result.Success)
				assert.True(t, result.Triggered)
				assert.Equal(t, "Alert would be triggered", result.Message)
				assert.Equal(t, 85.5, result.Value)
			},
		},
		{
			name:       "successful test - not triggered",
			id:         "1",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Alert test completed",
				Data: &AlertTestResult{
					Success:   true,
					Triggered: false,
					Message:   "Alert would not be triggered",
					Value:     70.0,
					Threshold: 80.0,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, result *AlertTestResult) {
				assert.NotNil(t, result)
				assert.True(t, result.Success)
				assert.False(t, result.Triggered)
			},
		},
		{
			name:       "test failed",
			id:         "1",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Alert test failed",
				Data: &AlertTestResult{
					Success:   false,
					Triggered: false,
					Message:   "Failed to fetch metrics",
					Errors:    []string{"Metric not found", "Connection timeout"},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, result *AlertTestResult) {
				assert.NotNil(t, result)
				assert.False(t, result.Success)
				assert.Len(t, result.Errors, 2)
			},
		},
		{
			name:       "alert not found",
			id:         "999",
			mockStatus: http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Alert not found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method and path
				assert.Equal(t, "POST", r.Method)
				assert.Contains(t, r.URL.Path, "/test")

				// Send mock response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			// Create client
			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			// Call Test
			result, err := client.Alerts.Test(context.Background(), tt.id)

			// Check error
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestAlertsService_Acknowledge tests the Acknowledge method
func TestAlertsService_Acknowledge(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		message    string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:       "successful acknowledgment",
			id:         "1",
			message:    "Investigating the issue",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Alert acknowledged successfully",
			},
			wantErr: false,
		},
		{
			name:       "acknowledgment with empty message",
			id:         "1",
			message:    "",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Alert acknowledged successfully",
			},
			wantErr: false,
		},
		{
			name:       "alert not found",
			id:         "999",
			message:    "Test message",
			mockStatus: http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Alert not found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method and path
				assert.Equal(t, "POST", r.Method)
				assert.Contains(t, r.URL.Path, "/acknowledge")

				// Verify request body
				var body map[string]interface{}
				err := json.NewDecoder(r.Body).Decode(&body)
				require.NoError(t, err)
				assert.Equal(t, tt.message, body["message"])

				// Send mock response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			// Create client
			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			// Call Acknowledge
			err = client.Alerts.Acknowledge(context.Background(), tt.id, tt.message)

			// Check error
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestAlertsService_ListChannels tests the ListChannels method
func TestAlertsService_ListChannels(t *testing.T) {
	tests := []struct {
		name       string
		opts       *ListOptions
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, []*AlertChannel, *PaginationMeta)
	}{
		{
			name: "successful list channels",
			opts: &ListOptions{
				Page:  1,
				Limit: 10,
			},
			mockStatus: http.StatusOK,
			mockBody: PaginatedResponse{
				Status: "success",
				Data: []*AlertChannel{
					{
						ID:   1,
						Name: "Email Notifications",
						Type: "email",
						Configuration: map[string]interface{}{
							"recipients": []string{"admin@example.com"},
						},
						Enabled: true,
					},
					{
						ID:   2,
						Name: "Slack Alerts",
						Type: "slack",
						Configuration: map[string]interface{}{
							"webhook_url": "https://hooks.slack.com/services/xxx",
						},
						Enabled: true,
					},
				},
				Meta: &PaginationMeta{
					Page:       1,
					Limit:      10,
					TotalItems: 2,
					TotalPages: 1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, channels []*AlertChannel, meta *PaginationMeta) {
				assert.NotNil(t, channels)
				assert.Len(t, channels, 2)
				assert.Equal(t, "Email Notifications", channels[0].Name)
				assert.Equal(t, "email", channels[0].Type)
				assert.Equal(t, "Slack Alerts", channels[1].Name)
				assert.Equal(t, "slack", channels[1].Type)
				assert.NotNil(t, meta)
			},
		},
		{
			name:       "empty channels list",
			opts:       nil,
			mockStatus: http.StatusOK,
			mockBody: PaginatedResponse{
				Status: "success",
				Data:   []*AlertChannel{},
				Meta: &PaginationMeta{
					Page:       1,
					Limit:      25,
					TotalItems: 0,
					TotalPages: 0,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, channels []*AlertChannel, meta *PaginationMeta) {
				assert.NotNil(t, channels)
				assert.Len(t, channels, 0)
			},
		},
		{
			name:       "unauthorized",
			opts:       nil,
			mockStatus: http.StatusUnauthorized,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Unauthorized",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method and path
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/v1/alerts/channels", r.URL.Path)

				// Send mock response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			// Create client
			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			// Call ListChannels
			result, meta, err := client.Alerts.ListChannels(context.Background(), tt.opts)

			// Check error
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Nil(t, meta)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result, meta)
				}
			}
		})
	}
}

// TestAlertJSON tests JSON marshaling and unmarshaling of Alert
func TestAlertJSON(t *testing.T) {
	alert := &Alert{
		GormModel:   GormModel{ID: 1},
		Name:        "High CPU Alert",
		Description: "Alert when CPU usage exceeds threshold",
		Type:        "metric",
		MetricName:  "cpu.usage",
		Condition:   "gt",
		Threshold:   80.0,
		Duration:    300,
		Frequency:   60,
		Severity:    "warning",
		Enabled:     true,
		Status:      "active",
	}

	// Marshal to JSON
	data, err := json.Marshal(alert)
	require.NoError(t, err)
	assert.Contains(t, string(data), "High CPU Alert")

	// Unmarshal from JSON
	var decoded Alert
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, alert.Name, decoded.Name)
	assert.Equal(t, alert.Type, decoded.Type)
	assert.Equal(t, alert.Threshold, decoded.Threshold)
}

// TestAlertHistoryEntryJSON tests JSON marshaling and unmarshaling of AlertHistoryEntry
func TestAlertHistoryEntryJSON(t *testing.T) {
	now := time.Now()
	entry := &AlertHistoryEntry{
		ID:          1,
		AlertID:     1,
		TriggeredAt: &CustomTime{Time: now},
		ResolvedAt:  &CustomTime{Time: now},
		Status:      "resolved",
		Value:       85.5,
		Threshold:   80.0,
		Message:     "CPU usage exceeded threshold",
		Details: map[string]interface{}{
			"host": "server-01",
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(entry)
	require.NoError(t, err)
	assert.Contains(t, string(data), "CPU usage exceeded threshold")

	// Unmarshal from JSON
	var decoded AlertHistoryEntry
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, entry.Status, decoded.Status)
	assert.Equal(t, entry.Value, decoded.Value)
}

// TestAlertTestResultJSON tests JSON marshaling and unmarshaling of AlertTestResult
func TestAlertTestResultJSON(t *testing.T) {
	result := &AlertTestResult{
		Success:   true,
		Triggered: true,
		Message:   "Alert would be triggered",
		Value:     85.5,
		Threshold: 80.0,
		Details: map[string]interface{}{
			"metric": "cpu.usage",
		},
		Errors: []string{},
	}

	// Marshal to JSON
	data, err := json.Marshal(result)
	require.NoError(t, err)
	assert.Contains(t, string(data), "Alert would be triggered")

	// Unmarshal from JSON
	var decoded AlertTestResult
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, result.Success, decoded.Success)
	assert.Equal(t, result.Triggered, decoded.Triggered)
	assert.Equal(t, result.Value, decoded.Value)
}

// TestAlertChannelJSON tests JSON marshaling and unmarshaling of AlertChannel
func TestAlertChannelJSON(t *testing.T) {
	channel := &AlertChannel{
		ID:   1,
		Name: "Email Notifications",
		Type: "email",
		Configuration: map[string]interface{}{
			"recipients": []string{"admin@example.com"},
		},
		Enabled: true,
	}

	// Marshal to JSON
	data, err := json.Marshal(channel)
	require.NoError(t, err)
	assert.Contains(t, string(data), "Email Notifications")

	// Unmarshal from JSON
	var decoded AlertChannel
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, channel.Name, decoded.Name)
	assert.Equal(t, channel.Type, decoded.Type)
	assert.Equal(t, channel.Enabled, decoded.Enabled)
}
