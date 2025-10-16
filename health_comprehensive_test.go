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

// TestHealthService_GetHealthComprehensive tests the GetHealth method
func TestHealthService_GetHealthComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *HealthStatus)
	}{
		{
			name:       "success - system healthy",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"status":  "ok",
					"healthy": true,
					"version": "1.0.0",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, health *HealthStatus) {
				assert.Equal(t, "ok", health.Status)
				assert.True(t, health.Healthy)
				assert.Equal(t, "1.0.0", health.Version)
			},
		},
		{
			name:       "success - system degraded",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"status":  "degraded",
					"healthy": false,
					"version": "1.0.0",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, health *HealthStatus) {
				assert.Equal(t, "degraded", health.Status)
				assert.False(t, health.Healthy)
			},
		},
		{
			name:       "unauthorized",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "server error",
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
			if tt.wantErr && tt.mockStatus >= 500 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
			}

			result, err := client.Health.GetHealth(ctx)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestHealthService_GetHealthDetailedComprehensive tests the GetHealthDetailed method
func TestHealthService_GetHealthDetailedComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *DetailedHealthStatus)
	}{
		{
			name:       "success - detailed health",
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"status":  "ok",
					"healthy": true,
					"version": "1.0.0",
					"uptime":  3600,
					"services": map[string]interface{}{
						"api": map[string]interface{}{
							"status":  "healthy",
							"healthy": true,
						},
						"database": map[string]interface{}{
							"status":  "healthy",
							"healthy": true,
						},
					},
					"database": map[string]interface{}{
						"connected":  true,
						"latency_ms": 5,
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, health *DetailedHealthStatus) {
				assert.Equal(t, "ok", health.Status)
				assert.True(t, health.Healthy)
				assert.Equal(t, int64(3600), health.Uptime)
				assert.NotNil(t, health.Services)
			},
		},
		{
			name:       "unauthorized",
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "server error",
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
			if tt.wantErr && tt.mockStatus >= 500 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
			}

			result, err := client.Health.GetHealthDetailed(ctx)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestHealthService_CreateHealthCheckComprehensive tests the Create method
func TestHealthService_CreateHealthCheckComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		request    *CreateHealthCheckRequest
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *HealthCheck)
	}{
		{
			name: "success - create health check",
			request: &CreateHealthCheckRequest{
				ServerID:         1,
				CheckName:        "CPU Check",
				CheckType:        "cpu",
				CheckDescription: "Monitor CPU usage",
				IsEnabled:        true,
				CheckInterval:    5,
				CheckTimeout:     30,
			},
			mockStatus: http.StatusCreated,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":                     1,
					"server_id":              1,
					"check_name":             "CPU Check",
					"check_type":             "cpu",
					"check_description":      "Monitor CPU usage",
					"is_enabled":             true,
					"check_interval_minutes": 5,
					"check_timeout_seconds":  30,
					"last_status":            "pending",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, check *HealthCheck) {
				assert.Equal(t, uint(1), check.ID)
				assert.Equal(t, "CPU Check", check.CheckName)
				assert.True(t, check.IsEnabled)
			},
		},
		{
			name: "validation error - missing name",
			request: &CreateHealthCheckRequest{
				ServerID:  1,
				CheckType: "cpu",
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Check name is required",
			},
			wantErr: true,
		},
		{
			name: "unauthorized",
			request: &CreateHealthCheckRequest{
				ServerID:  1,
				CheckName: "Test",
				CheckType: "cpu",
			},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name: "server error",
			request: &CreateHealthCheckRequest{
				ServerID:  1,
				CheckName: "Test",
				CheckType: "cpu",
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
			if tt.wantErr && tt.mockStatus >= 500 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
			}

			result, err := client.Health.Create(ctx, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestHealthService_GetHealthCheckComprehensive tests the Get method
func TestHealthService_GetHealthCheckComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		id         uint
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *HealthCheck)
	}{
		{
			name:       "success - get health check",
			id:         1,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":                     1,
					"server_id":              1,
					"check_name":             "CPU Check",
					"check_type":             "cpu",
					"is_enabled":             true,
					"check_interval_minutes": 5,
					"last_status":            "healthy",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, check *HealthCheck) {
				assert.Equal(t, uint(1), check.ID)
				assert.Equal(t, "CPU Check", check.CheckName)
				assert.True(t, check.IsEnabled)
			},
		},
		{
			name:       "not found",
			id:         999,
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Health check not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			id:         1,
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			id:         1,
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
			if tt.wantErr && tt.mockStatus >= 500 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
			}

			result, err := client.Health.Get(ctx, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestHealthService_DeleteHealthCheckComprehensive tests the Delete method
func TestHealthService_DeleteHealthCheckComprehensive(t *testing.T) {
	tests := []struct {
		name       string
		id         uint
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:       "success - delete health check",
			id:         1,
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status":  "success",
				"message": "Health check deleted",
			},
			wantErr: false,
		},
		{
			name:       "success - no content",
			id:         2,
			mockStatus: http.StatusNoContent,
			mockBody:   nil,
			wantErr:    false,
		},
		{
			name:       "not found",
			id:         999,
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Health check not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			id:         1,
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			id:         1,
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
				assert.Equal(t, "DELETE", r.Method)

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
			if tt.wantErr && tt.mockStatus >= 500 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
			}

			err = client.Health.Delete(ctx, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
