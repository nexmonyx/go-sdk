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

func TestQuotaHistoryService_RecordQuotaUsage(t *testing.T) {
	tests := []struct {
		name           string
		records        []QuotaUsageRecord
		responseStatus int
		wantErr        bool
	}{
		{
			name: "record single usage",
			records: []QuotaUsageRecord{
				{
					OrganizationID: 123,
					ResourceType:   "cpu",
					Usage:          75.5,
					Quota:          100.0,
					Timestamp:      time.Now(),
				},
			},
			responseStatus: http.StatusOK,
		},
		{
			name: "record multiple resources",
			records: []QuotaUsageRecord{
				{OrganizationID: 123, ResourceType: "cpu", Usage: 80, Quota: 100},
				{OrganizationID: 123, ResourceType: "memory", Usage: 60, Quota: 100},
				{OrganizationID: 123, ResourceType: "storage", Usage: 500, Quota: 1000},
			},
			responseStatus: http.StatusOK,
		},
		{
			name:           "empty records",
			records:        []QuotaUsageRecord{},
			responseStatus: http.StatusBadRequest,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/v1/admin/quota-history/record", r.URL.Path)

				var req QuotaUsageRecordRequest
				json.NewDecoder(r.Body).Decode(&req)
				assert.Len(t, req.Records, len(tt.records))

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(StandardResponse{Status: "success"})
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL})
			err := client.QuotaHistory.RecordQuotaUsage(context.Background(), tt.records)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestQuotaHistoryService_GetHistoricalUsage(t *testing.T) {
	tests := []struct {
		name         string
		orgID        uint
		resourceType string
		startDate    time.Time
		endDate      time.Time
		validateResp func(*testing.T, []QuotaUsageHistory)
	}{
		{
			name:         "get CPU usage history",
			orgID:        123,
			resourceType: "cpu",
			startDate:    time.Now().AddDate(0, 0, -7),
			endDate:      time.Now(),
			validateResp: func(t *testing.T, history []QuotaUsageHistory) {
				assert.NotEmpty(t, history)
				for _, h := range history {
					assert.Equal(t, "cpu", h.ResourceType)
				}
			},
		},
		{
			name:         "get all resource types",
			orgID:        456,
			resourceType: "",
			startDate:    time.Now().AddDate(0, 0, -30),
			endDate:      time.Now(),
			validateResp: func(t *testing.T, history []QuotaUsageHistory) {
				assert.NotEmpty(t, history)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/quota-history/")
				assert.Contains(t, r.URL.Path, "/usage")

				if tt.resourceType != "" {
					assert.Equal(t, tt.resourceType, r.URL.Query().Get("resource_type"))
				}

				history := []QuotaUsageHistory{
					{
						OrganizationID: tt.orgID,
						ResourceType:   tt.resourceType,
						Usage:          75.0,
						Quota:          100.0,
						RecordedAt:     time.Now(),
					},
				}

				json.NewEncoder(w).Encode(StandardResponse{
					Status: "success",
					Data:   &history,
				})
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL})
			result, err := client.QuotaHistory.GetHistoricalUsage(
				context.Background(),
				tt.orgID,
				tt.resourceType,
				tt.startDate,
				tt.endDate,
			)

			require.NoError(t, err)
			if tt.validateResp != nil {
				tt.validateResp(t, result)
			}
		})
	}
}

func TestQuotaHistoryService_GetAverageUtilization(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/average-utilization")
		assert.Equal(t, "cpu", r.URL.Query().Get("resource_type"))

		json.NewEncoder(w).Encode(StandardResponse{
			Status: "success",
			Data: &AverageUtilizationResponse{
				OrganizationID:  123,
				ResourceType:    "cpu",
				AverageUsage:    72.5,
				AverageQuota:    100.0,
				Utilization:     72.5,
				DataPoints:      168,
				StartDate:       time.Now().AddDate(0, 0, -7),
				EndDate:         time.Now(),
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, err := client.QuotaHistory.GetAverageUtilization(
		context.Background(),
		123,
		"cpu",
		time.Now().AddDate(0, 0, -7),
		time.Now(),
	)

	require.NoError(t, err)
	assert.Equal(t, "cpu", result.ResourceType)
	assert.Equal(t, 72.5, result.AverageUsage)
	assert.Equal(t, 168, result.DataPoints)
}

func TestQuotaHistoryService_GetPeakUtilization(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/peak-utilization")

		peakTime := time.Now().Add(-2 * time.Hour)
		json.NewEncoder(w).Encode(StandardResponse{
			Status: "success",
			Data: &QuotaUsageHistory{
				OrganizationID: 123,
				ResourceType:   "memory",
				Usage:          95.0,
				Quota:          100.0,
				RecordedAt:     peakTime,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, err := client.QuotaHistory.GetPeakUtilization(
		context.Background(),
		123,
		"memory",
		time.Now().AddDate(0, 0, -7),
		time.Now(),
	)

	require.NoError(t, err)
	assert.Equal(t, "memory", result.ResourceType)
	assert.Equal(t, 95.0, result.Usage)
}

func TestQuotaHistoryService_GetDailyAggregates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/daily-aggregates")

		aggregates := []DailyAggregateResponse{
			{
				Date:         "2024-01-15",
				ResourceType: "cpu",
				AvgUsage:     70.0,
				MinUsage:     45.0,
				MaxUsage:     95.0,
				DataPoints:   24,
			},
			{
				Date:         "2024-01-14",
				ResourceType: "cpu",
				AvgUsage:     65.0,
				MinUsage:     40.0,
				MaxUsage:     90.0,
				DataPoints:   24,
			},
		}

		json.NewEncoder(w).Encode(StandardResponse{
			Status: "success",
			Data:   &aggregates,
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, err := client.QuotaHistory.GetDailyAggregates(
		context.Background(),
		123,
		"cpu",
		time.Now().AddDate(0, 0, -30),
		time.Now(),
	)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "2024-01-15", result[0].Date)
	assert.Equal(t, 70.0, result[0].AvgUsage)
}

func TestQuotaHistoryService_GetResourceSummary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/resource-summary")

		summaries := []ResourceSummaryResponse{
			{
				ResourceType:   "cpu",
				CurrentUsage:   75.0,
				CurrentQuota:   100.0,
				AverageUsage:   70.0,
				PeakUsage:      95.0,
				Utilization:    75.0,
			},
			{
				ResourceType:   "memory",
				CurrentUsage:   60.0,
				CurrentQuota:   100.0,
				AverageUsage:   55.0,
				PeakUsage:      85.0,
				Utilization:    60.0,
			},
		}

		json.NewEncoder(w).Encode(StandardResponse{
			Status: "success",
			Data:   &summaries,
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, err := client.QuotaHistory.GetResourceSummary(
		context.Background(),
		123,
		time.Now().AddDate(0, 0, -7),
		time.Now(),
	)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "cpu", result[0].ResourceType)
	assert.Equal(t, 75.0, result[0].CurrentUsage)
	assert.Equal(t, "memory", result[1].ResourceType)
}

func TestQuotaHistoryService_GetUsageTrend(t *testing.T) {
	tests := []struct {
		name         string
		orgID        uint
		resourceType string
		days         int
		validateResp func(*testing.T, *UsageTrendResponse)
	}{
		{
			name:         "increasing trend",
			orgID:        123,
			resourceType: "cpu",
			days:         7,
			validateResp: func(t *testing.T, trend *UsageTrendResponse) {
				assert.Equal(t, "cpu", trend.ResourceType)
				assert.Equal(t, "increasing", trend.Trend)
				assert.Greater(t, trend.Slope, 0.0)
			},
		},
		{
			name:         "decreasing trend",
			orgID:        456,
			resourceType: "memory",
			days:         30,
			validateResp: func(t *testing.T, trend *UsageTrendResponse) {
				assert.Equal(t, "decreasing", trend.Trend)
				assert.Less(t, trend.Slope, 0.0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/usage-trend")

				trend := "increasing"
				slope := 0.5
				if tt.name == "decreasing trend" {
					trend = "decreasing"
					slope = -0.3
				}

				json.NewEncoder(w).Encode(StandardResponse{
					Status: "success",
					Data: &UsageTrendResponse{
						OrganizationID: tt.orgID,
						ResourceType:   tt.resourceType,
						Trend:          trend,
						Slope:          slope,
						DataPoints:     tt.days * 24,
						StartDate:      time.Now().AddDate(0, 0, -tt.days),
						EndDate:        time.Now(),
					},
				})
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL})
			result, err := client.QuotaHistory.GetUsageTrend(
				context.Background(),
				tt.orgID,
				tt.resourceType,
				tt.days,
			)

			require.NoError(t, err)
			if tt.validateResp != nil {
				tt.validateResp(t, result)
			}
		})
	}
}

func TestQuotaHistoryService_DetectUsagePatterns(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/detect-patterns")
		assert.Equal(t, "7", r.URL.Query().Get("days"))

		json.NewEncoder(w).Encode(StandardResponse{
			Status: "success",
			Data: &UsagePatternsResponse{
				OrganizationID: 123,
				Patterns: []UsagePattern{
					{
						ResourceType: "cpu",
						PatternType:  "spike",
						Description:  "Unusual spike detected at 14:00",
						Severity:     "warning",
						DetectedAt:   time.Now().Add(-2 * time.Hour),
						Metadata: map[string]interface{}{
							"peak_value": 95.0,
							"avg_value":  70.0,
						},
					},
					{
						ResourceType: "memory",
						PatternType:  "gradual_increase",
						Description:  "Steady increase over 7 days",
						Severity:     "info",
						DetectedAt:   time.Now(),
					},
				},
				AnalyzedPeriod: "7 days",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, err := client.QuotaHistory.DetectUsagePatterns(context.Background(), 123, 7)

	require.NoError(t, err)
	assert.Equal(t, uint(123), result.OrganizationID)
	assert.Len(t, result.Patterns, 2)
	assert.Equal(t, "spike", result.Patterns[0].PatternType)
	assert.Equal(t, "gradual_increase", result.Patterns[1].PatternType)
}

func TestQuotaHistoryService_CleanupOldRecords(t *testing.T) {
	tests := []struct {
		name           string
		orgID          uint
		retentionDays  int
		expectedDelete int
		responseStatus int
		wantErr        bool
	}{
		{
			name:           "cleanup with 90 day retention",
			orgID:          123,
			retentionDays:  90,
			expectedDelete: 1500,
			responseStatus: http.StatusOK,
		},
		{
			name:           "cleanup with 30 day retention",
			orgID:          456,
			retentionDays:  30,
			expectedDelete: 5000,
			responseStatus: http.StatusOK,
		},
		{
			name:           "minimum retention days",
			orgID:          789,
			retentionDays:  7,
			expectedDelete: 100,
			responseStatus: http.StatusOK,
		},
		{
			name:           "invalid retention days",
			orgID:          999,
			retentionDays:  3,
			responseStatus: http.StatusBadRequest,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "DELETE", r.Method)
				assert.Contains(t, r.URL.Path, "/cleanup")

				if tt.retentionDays > 0 {
					assert.NotEmpty(t, r.URL.Query().Get("retention_days"))
				}

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(struct {
					Status  string `json:"status"`
					Message string `json:"message"`
					Data    struct {
						DeletedCount  int    `json:"deleted_count"`
						RetentionDays int    `json:"retention_days"`
						CutoffDate    string `json:"cutoff_date"`
					} `json:"data"`
				}{
					Status:  "success",
					Message: "Cleanup completed",
					Data: struct {
						DeletedCount  int    `json:"deleted_count"`
						RetentionDays int    `json:"retention_days"`
						CutoffDate    string `json:"cutoff_date"`
					}{
						DeletedCount:  tt.expectedDelete,
						RetentionDays: tt.retentionDays,
						CutoffDate:    time.Now().AddDate(0, 0, -tt.retentionDays).Format(time.RFC3339),
					},
				})
			}))
			defer server.Close()

			client, _ := NewClient(&Config{BaseURL: server.URL})
			deletedCount, err := client.QuotaHistory.CleanupOldRecords(
				context.Background(),
				tt.orgID,
				tt.retentionDays,
			)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedDelete, deletedCount)
			}
		})
	}
}

// Integration test - complete quota tracking workflow
func TestQuotaHistoryService_CompleteWorkflow(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		switch {
		case r.Method == "POST" && r.URL.Path == "/v1/admin/quota-history/record":
			// Record usage
			json.NewEncoder(w).Encode(StandardResponse{Status: "success"})

		case r.Method == "GET" && r.URL.Path == "/v1/admin/quota-history/123/usage":
			// Get historical usage
			history := []QuotaUsageHistory{
				{OrganizationID: 123, ResourceType: "cpu", Usage: 75.0},
			}
			json.NewEncoder(w).Encode(StandardResponse{Status: "success", Data: &history})

		case r.Method == "GET" && r.URL.Path == "/v1/admin/quota-history/123/usage-trend":
			// Get trend
			json.NewEncoder(w).Encode(StandardResponse{
				Status: "success",
				Data:   &UsageTrendResponse{ResourceType: "cpu", Trend: "stable"},
			})

		case r.Method == "GET" && r.URL.Path == "/v1/admin/quota-history/123/detect-patterns":
			// Detect patterns
			json.NewEncoder(w).Encode(StandardResponse{
				Status: "success",
				Data:   &UsagePatternsResponse{Patterns: []UsagePattern{}},
			})
		}
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	// 1. Record usage
	err := client.QuotaHistory.RecordQuotaUsage(context.Background(), []QuotaUsageRecord{
		{OrganizationID: 123, ResourceType: "cpu", Usage: 75.0, Quota: 100.0},
	})
	require.NoError(t, err)

	// 2. Get historical usage
	history, err := client.QuotaHistory.GetHistoricalUsage(context.Background(), 123, "cpu", time.Now().AddDate(0, 0, -7), time.Now())
	require.NoError(t, err)
	assert.NotEmpty(t, history)

	// 3. Analyze trend
	trend, err := client.QuotaHistory.GetUsageTrend(context.Background(), 123, "cpu", 7)
	require.NoError(t, err)
	assert.Equal(t, "stable", trend.Trend)

	// 4. Detect patterns
	patterns, err := client.QuotaHistory.DetectUsagePatterns(context.Background(), 123, 7)
	require.NoError(t, err)
	assert.NotNil(t, patterns)

	assert.Equal(t, 4, callCount)
}
