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

func TestHealthService_GetHealth(t *testing.T) {
	tests := []struct {
		name       string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *HealthStatus)
	}{
		{
			name:       "successful get health",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Health retrieved successfully",
				Data: &HealthStatus{
					Status:    "ok",
					Healthy:   true,
					Version:   "1.0.0",
					Timestamp: &CustomTime{Time: time.Now()},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, health *HealthStatus) {
				assert.NotNil(t, health)
				assert.Equal(t, "ok", health.Status)
				assert.True(t, health.Healthy)
				assert.Equal(t, "1.0.0", health.Version)
				assert.NotNil(t, health.Timestamp)
			},
		},
		{
			name:       "unhealthy status",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Health retrieved successfully",
				Data: &HealthStatus{
					Status:    "degraded",
					Healthy:   false,
					Version:   "1.0.0",
					Timestamp: &CustomTime{Time: time.Now()},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, health *HealthStatus) {
				assert.NotNil(t, health)
				assert.Equal(t, "degraded", health.Status)
				assert.False(t, health.Healthy)
			},
		},
		{
			name:       "unauthorized access",
			mockStatus: http.StatusUnauthorized,
			mockBody: ErrorResponse{
				Status:  "error",
				Message: "Unauthorized",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			mockStatus: http.StatusInternalServerError,
			mockBody: ErrorResponse{
				Status:  "error",
				Message: "Internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/v1/healthz", r.URL.Path)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			health, err := client.Health.GetHealth(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, health)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, health)
				}
			}
		})
	}
}

func TestHealthService_GetHealthDetailed(t *testing.T) {
	tests := []struct {
		name       string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *DetailedHealthStatus)
	}{
		{
			name:       "successful get detailed health",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Detailed health retrieved successfully",
				Data: &DetailedHealthStatus{
					Status:    "ok",
					Healthy:   true,
					Version:   "1.0.0",
					Timestamp: &CustomTime{Time: time.Now()},
					Uptime:    3600,
					Services: map[string]ServiceHealth{
						"api": {
							Healthy:      true,
							Status:       "ok",
							ResponseTime: 50,
						},
						"database": {
							Healthy:      true,
							Status:       "ok",
							ResponseTime: 10,
						},
					},
					Database: &DatabaseHealth{
						Healthy:         true,
						ConnectionCount: 5,
						MaxConnections:  100,
						ResponseTime:    10,
						Version:         "PostgreSQL 15.0",
					},
					Redis: &RedisHealth{
						Healthy:          true,
						Connected:        true,
						ResponseTime:     5,
						MemoryUsage:      1024000,
						ConnectedClients: 10,
						Version:          "7.0.0",
					},
					Metrics: map[string]interface{}{
						"requests_per_second": 100,
						"error_rate":          0.01,
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, health *DetailedHealthStatus) {
				assert.NotNil(t, health)
				assert.Equal(t, "ok", health.Status)
				assert.True(t, health.Healthy)
				assert.Equal(t, "1.0.0", health.Version)
				assert.Equal(t, int64(3600), health.Uptime)
				assert.Len(t, health.Services, 2)
				assert.NotNil(t, health.Database)
				assert.True(t, health.Database.Healthy)
				assert.NotNil(t, health.Redis)
				assert.True(t, health.Redis.Healthy)
				assert.NotEmpty(t, health.Metrics)
			},
		},
		{
			name:       "degraded health with failing services",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Detailed health retrieved successfully",
				Data: &DetailedHealthStatus{
					Status:    "degraded",
					Healthy:   false,
					Version:   "1.0.0",
					Timestamp: &CustomTime{Time: time.Now()},
					Uptime:    3600,
					Services: map[string]ServiceHealth{
						"api": {
							Healthy:      false,
							Status:       "error",
							Message:      "High response time",
							ResponseTime: 5000,
						},
					},
					Database: &DatabaseHealth{
						Healthy:         true,
						ConnectionCount: 90,
						MaxConnections:  100,
						ResponseTime:    50,
						Version:         "PostgreSQL 15.0",
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, health *DetailedHealthStatus) {
				assert.NotNil(t, health)
				assert.Equal(t, "degraded", health.Status)
				assert.False(t, health.Healthy)
				assert.NotEmpty(t, health.Services)
				apiHealth := health.Services["api"]
				assert.False(t, apiHealth.Healthy)
				assert.Equal(t, "error", apiHealth.Status)
			},
		},
		{
			name:       "unauthorized access",
			mockStatus: http.StatusUnauthorized,
			mockBody: ErrorResponse{
				Status:  "error",
				Message: "Unauthorized",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/v1/health/detailed", r.URL.Path)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			health, err := client.Health.GetHealthDetailed(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, health)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, health)
				}
			}
		})
	}
}

func TestHealthService_List(t *testing.T) {
	serverID := uint(1)
	checkType := "http"
	isEnabled := true

	tests := []struct {
		name       string
		opts       *HealthCheckListOptions
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, []HealthCheck, *PaginationMeta)
	}{
		{
			name: "successful list with options",
			opts: &HealthCheckListOptions{
				ServerID:  &serverID,
				CheckType: &checkType,
				IsEnabled: &isEnabled,
				ListOptions: ListOptions{
					Page:  1,
					Limit: 25,
				},
			},
			mockStatus: http.StatusOK,
			mockBody: PaginatedResponse{
				Status:  "success",
				Message: "Health checks retrieved successfully",
				Data: &[]HealthCheck{
					{
						ID:              1,
						ServerID:        1,
						CheckName:       "HTTP Health Check",
						CheckType:       "http",
						IsEnabled:       true,
						CheckInterval:   5,
						CheckTimeout:    30,
						MaxRetries:      3,
						RetryInterval:   10,
						LastStatus:      "healthy",
						LastScore:       100,
						CreatedAt:       &CustomTime{Time: time.Now()},
						UpdatedAt:       &CustomTime{Time: time.Now()},
					},
				},
				Meta: &PaginationMeta{
					Page:       1,
					TotalPages: 1,
					TotalItems: 1,
					Limit:      25,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, checks []HealthCheck, meta *PaginationMeta) {
				assert.Len(t, checks, 1)
				assert.Equal(t, uint(1), checks[0].ID)
				assert.Equal(t, "HTTP Health Check", checks[0].CheckName)
				assert.Equal(t, "http", checks[0].CheckType)
				assert.True(t, checks[0].IsEnabled)
				assert.NotNil(t, meta)
				assert.Equal(t, 1, meta.Page)
				assert.Equal(t, 1, meta.TotalItems)
			},
		},
		{
			name:       "list without options",
			opts:       nil,
			mockStatus: http.StatusOK,
			mockBody: PaginatedResponse{
				Status:  "success",
				Message: "Health checks retrieved successfully",
				Data:    &[]HealthCheck{},
				Meta: &PaginationMeta{
					Page:       1,
					TotalPages: 0,
					TotalItems: 0,
					Limit:      25,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, checks []HealthCheck, meta *PaginationMeta) {
				assert.Empty(t, checks)
				assert.NotNil(t, meta)
			},
		},
		{
			name: "unauthorized access",
			opts: nil,
			mockStatus: http.StatusUnauthorized,
			mockBody: ErrorResponse{
				Status:  "error",
				Message: "Unauthorized",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/v1/health/checks", r.URL.Path)

				if tt.opts != nil {
					query := r.URL.Query()
					if tt.opts.ServerID != nil {
						assert.Equal(t, "1", query.Get("server_id"))
					}
					if tt.opts.CheckType != nil {
						assert.Equal(t, "http", query.Get("check_type"))
					}
					if tt.opts.IsEnabled != nil {
						assert.Equal(t, "true", query.Get("is_enabled"))
					}
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			checks, meta, err := client.Health.List(context.Background(), tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, checks)
				assert.Nil(t, meta)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, checks, meta)
				}
			}
		})
	}
}

func TestHealthService_Get(t *testing.T) {
	tests := []struct {
		name       string
		id         uint
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *HealthCheck)
	}{
		{
			name:       "successful get",
			id:         1,
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Health check retrieved successfully",
				Data: &HealthCheck{
					ID:              1,
					ServerID:        1,
					CheckName:       "HTTP Health Check",
					CheckType:       "http",
					CheckDescription: "Checks HTTP endpoint",
					IsEnabled:       true,
					CheckInterval:   5,
					CheckTimeout:    30,
					MaxRetries:      3,
					RetryInterval:   10,
					CheckData: map[string]interface{}{
						"url":    "https://example.com",
						"method": "GET",
					},
					LastStatus:      "healthy",
					LastScore:       100,
					CreatedAt:       &CustomTime{Time: time.Now()},
					UpdatedAt:       &CustomTime{Time: time.Now()},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, check *HealthCheck) {
				assert.NotNil(t, check)
				assert.Equal(t, uint(1), check.ID)
				assert.Equal(t, "HTTP Health Check", check.CheckName)
				assert.Equal(t, "http", check.CheckType)
				assert.True(t, check.IsEnabled)
				assert.NotEmpty(t, check.CheckData)
			},
		},
		{
			name:       "not found",
			id:         999,
			mockStatus: http.StatusNotFound,
			mockBody: ErrorResponse{
				Status:  "error",
				Message: "Health check not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized access",
			id:         1,
			mockStatus: http.StatusUnauthorized,
			mockBody: ErrorResponse{
				Status:  "error",
				Message: "Unauthorized",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/health/checks/")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			check, err := client.Health.Get(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, check)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, check)
				}
			}
		})
	}
}

func TestHealthService_Create(t *testing.T) {
	tests := []struct {
		name       string
		req        *CreateHealthCheckRequest
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *HealthCheck)
	}{
		{
			name: "successful create",
			req: &CreateHealthCheckRequest{
				ServerID:         1,
				CheckName:        "New HTTP Check",
				CheckType:        "http",
				CheckDescription: "Monitors web endpoint",
				IsEnabled:        true,
				CheckInterval:    5,
				CheckTimeout:     30,
				MaxRetries:       3,
				RetryInterval:    10,
				CheckData: map[string]interface{}{
					"url":    "https://example.com",
					"method": "GET",
				},
			},
			mockStatus: http.StatusCreated,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Health check created successfully",
				Data: &HealthCheck{
					ID:               1,
					ServerID:         1,
					CheckName:        "New HTTP Check",
					CheckType:        "http",
					CheckDescription: "Monitors web endpoint",
					IsEnabled:        true,
					CheckInterval:    5,
					CheckTimeout:     30,
					MaxRetries:       3,
					RetryInterval:    10,
					CheckData: map[string]interface{}{
						"url":    "https://example.com",
						"method": "GET",
					},
					CreatedAt: &CustomTime{Time: time.Now()},
					UpdatedAt: &CustomTime{Time: time.Now()},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, check *HealthCheck) {
				assert.NotNil(t, check)
				assert.Equal(t, uint(1), check.ID)
				assert.Equal(t, "New HTTP Check", check.CheckName)
				assert.Equal(t, "http", check.CheckType)
				assert.True(t, check.IsEnabled)
			},
		},
		{
			name: "validation error",
			req: &CreateHealthCheckRequest{
				CheckName: "Invalid Check",
				// Missing required ServerID and CheckType
			},
			mockStatus: http.StatusBadRequest,
			mockBody: ErrorResponse{
				Status:  "error",
				Message: "Validation failed",
			},
			wantErr: true,
		},
		{
			name: "unauthorized access",
			req: &CreateHealthCheckRequest{
				ServerID:      1,
				CheckName:     "Test Check",
				CheckType:     "http",
				CheckInterval: 5,
				CheckTimeout:  30,
			},
			mockStatus: http.StatusUnauthorized,
			mockBody: ErrorResponse{
				Status:  "error",
				Message: "Unauthorized",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/v1/health/checks", r.URL.Path)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			check, err := client.Health.Create(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, check)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, check)
				}
			}
		})
	}
}

func TestHealthService_Update(t *testing.T) {
	checkName := "Updated HTTP Check"
	isEnabled := false
	checkInterval := 10

	tests := []struct {
		name       string
		id         uint
		req        *UpdateHealthCheckRequest
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *HealthCheck)
	}{
		{
			name: "successful update",
			id:   1,
			req: &UpdateHealthCheckRequest{
				CheckName:     &checkName,
				IsEnabled:     &isEnabled,
				CheckInterval: &checkInterval,
			},
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Health check updated successfully",
				Data: &HealthCheck{
					ID:            1,
					ServerID:      1,
					CheckName:     "Updated HTTP Check",
					CheckType:     "http",
					IsEnabled:     false,
					CheckInterval: 10,
					CheckTimeout:  30,
					MaxRetries:    3,
					RetryInterval: 10,
					UpdatedAt:     &CustomTime{Time: time.Now()},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, check *HealthCheck) {
				assert.NotNil(t, check)
				assert.Equal(t, uint(1), check.ID)
				assert.Equal(t, "Updated HTTP Check", check.CheckName)
				assert.False(t, check.IsEnabled)
				assert.Equal(t, 10, check.CheckInterval)
			},
		},
		{
			name: "not found",
			id:   999,
			req: &UpdateHealthCheckRequest{
				CheckName: &checkName,
			},
			mockStatus: http.StatusNotFound,
			mockBody: ErrorResponse{
				Status:  "error",
				Message: "Health check not found",
			},
			wantErr: true,
		},
		{
			name: "unauthorized access",
			id:   1,
			req: &UpdateHealthCheckRequest{
				CheckName: &checkName,
			},
			mockStatus: http.StatusUnauthorized,
			mockBody: ErrorResponse{
				Status:  "error",
				Message: "Unauthorized",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "PUT", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/health/checks/")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			check, err := client.Health.Update(context.Background(), tt.id, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, check)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, check)
				}
			}
		})
	}
}

func TestHealthService_Delete(t *testing.T) {
	tests := []struct {
		name       string
		id         uint
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name:       "successful delete",
			id:         1,
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Health check deleted successfully",
			},
			wantErr: false,
		},
		{
			name:       "not found",
			id:         999,
			mockStatus: http.StatusNotFound,
			mockBody: ErrorResponse{
				Status:  "error",
				Message: "Health check not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized access",
			id:         1,
			mockStatus: http.StatusUnauthorized,
			mockBody: ErrorResponse{
				Status:  "error",
				Message: "Unauthorized",
			},
			wantErr: true,
		},
		{
			name:       "forbidden - health check in use",
			id:         1,
			mockStatus: http.StatusForbidden,
			mockBody: ErrorResponse{
				Status:  "error",
				Message: "Cannot delete health check with active monitoring",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "DELETE", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/health/checks/")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			err = client.Health.Delete(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHealthService_GetHistory(t *testing.T) {
	healthCheckID := uint(1)
	serverID := uint(1)
	status := "healthy"
	fromDate := time.Now().Add(-24 * time.Hour)
	toDate := time.Now()

	tests := []struct {
		name       string
		opts       *HealthCheckHistoryListOptions
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, []HealthCheckHistory, *PaginationMeta)
	}{
		{
			name: "successful get history with options",
			opts: &HealthCheckHistoryListOptions{
				HealthCheckID: &healthCheckID,
				ServerID:      &serverID,
				Status:        &status,
				FromDate:      &fromDate,
				ToDate:        &toDate,
				ListOptions: ListOptions{
					Page:  1,
					Limit: 50,
				},
			},
			mockStatus: http.StatusOK,
			mockBody: PaginatedResponse{
				Status:  "success",
				Message: "Health check history retrieved successfully",
				Data: &[]HealthCheckHistory{
					{
						ID:            1,
						HealthCheckID: 1,
						ServerID:      1,
						Status:        "healthy",
						Score:         100,
						ResponseTime:  50,
						Attempt:       1,
						CreatedAt:     &CustomTime{Time: time.Now()},
					},
					{
						ID:            2,
						HealthCheckID: 1,
						ServerID:      1,
						Status:        "healthy",
						Score:         98,
						ResponseTime:  55,
						Attempt:       1,
						CreatedAt:     &CustomTime{Time: time.Now().Add(-5 * time.Minute)},
					},
				},
				Meta: &PaginationMeta{
					Page:       1,
					TotalPages: 1,
					TotalItems: 2,
					Limit:      50,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, history []HealthCheckHistory, meta *PaginationMeta) {
				assert.Len(t, history, 2)
				assert.Equal(t, uint(1), history[0].ID)
				assert.Equal(t, "healthy", history[0].Status)
				assert.Equal(t, 100, history[0].Score)
				assert.NotNil(t, meta)
				assert.Equal(t, 2, meta.TotalItems)
			},
		},
		{
			name:       "get history without options",
			opts:       nil,
			mockStatus: http.StatusOK,
			mockBody: PaginatedResponse{
				Status:  "success",
				Message: "Health check history retrieved successfully",
				Data:    &[]HealthCheckHistory{},
				Meta: &PaginationMeta{
					Page:       1,
					TotalPages: 0,
					TotalItems: 0,
					Limit:      25,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, history []HealthCheckHistory, meta *PaginationMeta) {
				assert.Empty(t, history)
				assert.NotNil(t, meta)
			},
		},
		{
			name: "history with error results",
			opts: &HealthCheckHistoryListOptions{
				HealthCheckID: &healthCheckID,
				Status:        &status,
			},
			mockStatus: http.StatusOK,
			mockBody: PaginatedResponse{
				Status:  "success",
				Message: "Health check history retrieved successfully",
				Data: &[]HealthCheckHistory{
					{
						ID:            3,
						HealthCheckID: 1,
						ServerID:      1,
						Status:        "critical",
						Score:         0,
						ResponseTime:  5000,
						ErrorMessage:  "Connection timeout",
						Attempt:       3,
						CreatedAt:     &CustomTime{Time: time.Now()},
					},
				},
				Meta: &PaginationMeta{
					Page:       1,
					TotalPages: 1,
					TotalItems: 1,
					Limit:      25,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, history []HealthCheckHistory, meta *PaginationMeta) {
				assert.Len(t, history, 1)
				assert.Equal(t, "critical", history[0].Status)
				assert.Equal(t, 0, history[0].Score)
				assert.Equal(t, "Connection timeout", history[0].ErrorMessage)
				assert.Equal(t, 3, history[0].Attempt)
			},
		},
		{
			name: "unauthorized access",
			opts: nil,
			mockStatus: http.StatusUnauthorized,
			mockBody: ErrorResponse{
				Status:  "error",
				Message: "Unauthorized",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/v1/health/history", r.URL.Path)

				if tt.opts != nil {
					query := r.URL.Query()
					if tt.opts.HealthCheckID != nil {
						assert.Equal(t, "1", query.Get("health_check_id"))
					}
					if tt.opts.ServerID != nil {
						assert.Equal(t, "1", query.Get("server_id"))
					}
					if tt.opts.Status != nil {
						assert.Equal(t, "healthy", query.Get("status"))
					}
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			history, meta, err := client.Health.GetHistory(context.Background(), tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, history)
				assert.Nil(t, meta)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, history, meta)
				}
			}
		})
	}
}

func TestHealthStatusJSON(t *testing.T) {
	now := time.Now()
	healthStatus := HealthStatus{
		Status:    "ok",
		Healthy:   true,
		Version:   "1.0.0",
		Timestamp: &CustomTime{Time: now},
	}

	// Test marshaling
	data, err := json.Marshal(healthStatus)
	require.NoError(t, err)
	assert.Contains(t, string(data), "ok")
	assert.Contains(t, string(data), "1.0.0")

	// Test unmarshaling
	var decoded HealthStatus
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "ok", decoded.Status)
	assert.True(t, decoded.Healthy)
	assert.Equal(t, "1.0.0", decoded.Version)
}

func TestHealthCheckJSON(t *testing.T) {
	now := time.Now()
	healthCheck := HealthCheck{
		ID:              1,
		ServerID:        1,
		CheckName:       "Test Check",
		CheckType:       "http",
		IsEnabled:       true,
		CheckInterval:   5,
		CheckTimeout:    30,
		MaxRetries:      3,
		RetryInterval:   10,
		LastStatus:      "healthy",
		LastScore:       100,
		CreatedAt:       &CustomTime{Time: now},
		UpdatedAt:       &CustomTime{Time: now},
	}

	// Test marshaling
	data, err := json.Marshal(healthCheck)
	require.NoError(t, err)
	assert.Contains(t, string(data), "Test Check")
	assert.Contains(t, string(data), "http")

	// Test unmarshaling
	var decoded HealthCheck
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, uint(1), decoded.ID)
	assert.Equal(t, "Test Check", decoded.CheckName)
	assert.Equal(t, "http", decoded.CheckType)
	assert.True(t, decoded.IsEnabled)
}
