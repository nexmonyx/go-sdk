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

func TestHealthService_GetSystemHealthOverview(t *testing.T) {
	tests := []struct {
		name       string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *SystemHealthOverview)
	}{
		{
			name:       "successful get system health overview",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "System health overview retrieved successfully",
				Data: &SystemHealthOverview{
					OverallStatus:   "healthy",
					OverallScore:    95.5,
					TotalServices:   10,
					HealthyServices: 9,
					WarningServices: 1,
					CriticalServices: 0,
					UnknownServices: 0,
					ServiceBreakdown: []ServiceBreakdown{
						{
							ServiceName:      "api",
							Status:           "healthy",
							Score:            98.0,
							CheckCount:       100,
							LastCheckTime:    "2025-10-21T00:00:00Z",
							ResponseTime:     45.5,
							UptimePercentage: 99.9,
						},
					},
					CategoryBreakdown: []CategoryBreakdown{
						{
							Category:     "database",
							HealthyCount: 3,
							WarningCount: 0,
							CriticalCount: 0,
							TotalCount:   3,
							AverageScore: 98.0,
						},
					},
					RecentTrends: []ServiceTrend{
						{
							ServiceName:    "api",
							TrendDirection: "improving",
							ScoreChange:    2.5,
							PeriodHours:    24,
						},
					},
					AlertSummary: AlertSummary{
						TotalAlerts:          5,
						CriticalAlerts:       1,
						WarningAlerts:        4,
						UnacknowledgedAlerts: 2,
					},
					RecentIncidents: []IncidentLog{
						{
							ServiceName:  "database",
							IncidentType: "connection_timeout",
							Severity:     "warning",
							StartTime:    "2025-10-20T12:00:00Z",
							EndTime:      "2025-10-20T12:05:00Z",
							Duration:     "5m",
							Resolved:     true,
						},
					},
					Timestamp: "2025-10-21T00:00:00Z",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, overview *SystemHealthOverview) {
				assert.NotNil(t, overview)
				assert.Equal(t, "healthy", overview.OverallStatus)
				assert.Equal(t, 95.5, overview.OverallScore)
				assert.Equal(t, 10, overview.TotalServices)
				assert.Equal(t, 9, overview.HealthyServices)
				assert.Len(t, overview.ServiceBreakdown, 1)
				assert.Equal(t, "api", overview.ServiceBreakdown[0].ServiceName)
				assert.Len(t, overview.CategoryBreakdown, 1)
				assert.Len(t, overview.RecentTrends, 1)
				assert.Equal(t, 5, overview.AlertSummary.TotalAlerts)
				assert.Len(t, overview.RecentIncidents, 1)
			},
		},
		{
			name:       "critical system status",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "System health overview retrieved successfully",
				Data: &SystemHealthOverview{
					OverallStatus:    "critical",
					OverallScore:     45.0,
					TotalServices:    10,
					HealthyServices:  5,
					WarningServices:  2,
					CriticalServices: 3,
					UnknownServices:  0,
					Timestamp:        "2025-10-21T00:00:00Z",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, overview *SystemHealthOverview) {
				assert.NotNil(t, overview)
				assert.Equal(t, "critical", overview.OverallStatus)
				assert.Equal(t, 45.0, overview.OverallScore)
				assert.Equal(t, 3, overview.CriticalServices)
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
			name:       "internal server error",
			mockStatus: http.StatusInternalServerError,
			mockBody: ErrorResponse{
				Status:  "error",
				Message: "Failed to retrieve system health overview",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/v1/health/system/overview", r.URL.Path)

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

			overview, err := client.Health.GetSystemHealthOverview(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, overview)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, overview)
				}
			}
		})
	}
}

func TestHealthService_GetServiceHealthHistory(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
		startTime   string
		endTime     string
		granularity string
		mockStatus  int
		mockBody    interface{}
		wantErr     bool
		checkFunc   func(*testing.T, *HealthMetricsHistory)
	}{
		{
			name:        "successful get service health history",
			serviceName: "api",
			startTime:   "2025-10-20T00:00:00Z",
			endTime:     "2025-10-21T00:00:00Z",
			granularity: "hour",
			mockStatus:  http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Service health history retrieved successfully",
				Data: &HealthMetricsHistory{
					ServiceName: "api",
					StartTime:   "2025-10-20T00:00:00Z",
					EndTime:     "2025-10-21T00:00:00Z",
					Granularity: "hour",
					DataPoints: []HealthMetricPoint{
						{
							Timestamp:       "2025-10-20T00:00:00Z",
							HealthyCount:    90,
							WarningCount:    5,
							CriticalCount:   3,
							UnknownCount:    2,
							TotalChecks:     100,
							AvgResponseTime: 45.5,
							Availability:    95.0,
							ErrorRate:       5.0,
						},
						{
							Timestamp:       "2025-10-20T01:00:00Z",
							HealthyCount:    92,
							WarningCount:    4,
							CriticalCount:   2,
							UnknownCount:    2,
							TotalChecks:     100,
							AvgResponseTime: 42.3,
							Availability:    96.0,
							ErrorRate:       4.0,
						},
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, history *HealthMetricsHistory) {
				assert.NotNil(t, history)
				assert.Equal(t, "api", history.ServiceName)
				assert.Equal(t, "hour", history.Granularity)
				assert.Len(t, history.DataPoints, 2)
				assert.Equal(t, 90, history.DataPoints[0].HealthyCount)
				assert.Equal(t, 45.5, history.DataPoints[0].AvgResponseTime)
			},
		},
		{
			name:        "service not found",
			serviceName: "nonexistent",
			startTime:   "",
			endTime:     "",
			granularity: "",
			mockStatus:  http.StatusNotFound,
			mockBody: ErrorResponse{
				Status:  "error",
				Message: "Service not found",
			},
			wantErr: true,
		},
		{
			name:        "invalid time range",
			serviceName: "api",
			startTime:   "invalid",
			endTime:     "invalid",
			granularity: "hour",
			mockStatus:  http.StatusBadRequest,
			mockBody: ErrorResponse{
				Status:  "error",
				Message: "Invalid time format",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/health/services/")
				assert.Contains(t, r.URL.Path, "/history")

				// Check query parameters
				if tt.startTime != "" {
					assert.Equal(t, tt.startTime, r.URL.Query().Get("start_time"))
				}
				if tt.endTime != "" {
					assert.Equal(t, tt.endTime, r.URL.Query().Get("end_time"))
				}
				if tt.granularity != "" {
					assert.Equal(t, tt.granularity, r.URL.Query().Get("granularity"))
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

			history, err := client.Health.GetServiceHealthHistory(
				context.Background(),
				tt.serviceName,
				tt.startTime,
				tt.endTime,
				tt.granularity,
			)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, history)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, history)
				}
			}
		})
	}
}

func TestHealthService_GetServiceHealthScore(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
		mockStatus  int
		mockBody    interface{}
		wantErr     bool
		checkFunc   func(*testing.T, *ServiceHealthScore)
	}{
		{
			name:        "successful get service health score",
			serviceName: "api",
			mockStatus:  http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Service health score retrieved successfully",
				Data: &ServiceHealthScore{
					ServiceName: "api",
					Score:       95.5,
					Timestamp:   "2025-10-21T00:00:00Z",
					Interpretation: map[string]string{
						"90-100": "excellent",
						"80-89":  "good",
						"70-79":  "warning",
						"0-69":   "critical",
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, score *ServiceHealthScore) {
				assert.NotNil(t, score)
				assert.Equal(t, "api", score.ServiceName)
				assert.Equal(t, 95.5, score.Score)
				assert.NotEmpty(t, score.Timestamp)
				assert.Len(t, score.Interpretation, 4)
				assert.Equal(t, "excellent", score.Interpretation["90-100"])
			},
		},
		{
			name:        "low health score",
			serviceName: "database",
			mockStatus:  http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Service health score retrieved successfully",
				Data: &ServiceHealthScore{
					ServiceName: "database",
					Score:       45.0,
					Timestamp:   "2025-10-21T00:00:00Z",
					Interpretation: map[string]string{
						"90-100": "excellent",
						"80-89":  "good",
						"70-79":  "warning",
						"0-69":   "critical",
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, score *ServiceHealthScore) {
				assert.NotNil(t, score)
				assert.Equal(t, "database", score.ServiceName)
				assert.Equal(t, 45.0, score.Score)
			},
		},
		{
			name:        "service not found",
			serviceName: "nonexistent",
			mockStatus:  http.StatusNotFound,
			mockBody: ErrorResponse{
				Status:  "error",
				Message: "Service not found",
			},
			wantErr: true,
		},
		{
			name:        "forbidden access",
			serviceName: "api",
			mockStatus:  http.StatusForbidden,
			mockBody: ErrorResponse{
				Status:  "error",
				Message: "Access denied",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/health/services/")
				assert.Contains(t, r.URL.Path, "/score")

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

			score, err := client.Health.GetServiceHealthScore(context.Background(), tt.serviceName)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, score)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, score)
				}
			}
		})
	}
}
