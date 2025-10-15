package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestQuotaHistoryService_RecordQuotaUsage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/admin/quota-history/record", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Quota usage recorded",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	records := []QuotaUsageRecord{
		{OrganizationID: 1, ResourceType: "cpu", UsedAmount: 50, HardLimit: 100, CollectedAt: time.Now()},
	}
	err := client.QuotaHistory.RecordQuotaUsage(context.Background(), records)
	assert.NoError(t, err)
}

func TestQuotaHistoryService_GetHistoricalUsage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/v1/admin/quota-history/")
		assert.Contains(t, r.URL.Path, "/usage")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   []map[string]interface{}{{"resource_type": "cpu", "usage": 50}},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	start := time.Now().Add(-7 * 24 * time.Hour)
	end := time.Now()
	history, err := client.QuotaHistory.GetHistoricalUsage(context.Background(), 1, "cpu", start, end)
	assert.NoError(t, err)
	assert.NotNil(t, history)
}

func TestQuotaHistoryService_GetAverageUtilization(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/average-utilization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   map[string]interface{}{"average": 75.5, "resource_type": "cpu"},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	start := time.Now().Add(-7 * 24 * time.Hour)
	end := time.Now()
	avg, err := client.QuotaHistory.GetAverageUtilization(context.Background(), 1, "cpu", start, end)
	assert.NoError(t, err)
	assert.NotNil(t, avg)
}

func TestQuotaHistoryService_GetPeakUtilization(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/peak-utilization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   map[string]interface{}{"peak": 95.0, "resource_type": "memory"},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	start := time.Now().Add(-30 * 24 * time.Hour)
	end := time.Now()
	peak, err := client.QuotaHistory.GetPeakUtilization(context.Background(), 1, "memory", start, end)
	assert.NoError(t, err)
	assert.NotNil(t, peak)
}

func TestQuotaHistoryService_GetDailyAggregates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/daily-aggregates")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": []map[string]interface{}{
				{"date": "2024-01-01", "avg_usage": 50.0},
				{"date": "2024-01-02", "avg_usage": 55.0},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	start := time.Now().Add(-30 * 24 * time.Hour)
	end := time.Now()
	aggregates, err := client.QuotaHistory.GetDailyAggregates(context.Background(), 1, "storage", start, end)
	assert.NoError(t, err)
	assert.NotNil(t, aggregates)
}

func TestQuotaHistoryService_GetResourceSummary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/resource-summary")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": []map[string]interface{}{
				{"resource_type": "cpu", "avg_usage": 60.0},
				{"resource_type": "memory", "avg_usage": 70.0},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	start := time.Now().Add(-7 * 24 * time.Hour)
	end := time.Now()
	summary, err := client.QuotaHistory.GetResourceSummary(context.Background(), 1, start, end)
	assert.NoError(t, err)
	assert.NotNil(t, summary)
}

func TestQuotaHistoryService_GetUsageTrend(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/usage-trend")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   map[string]interface{}{"trend": "increasing", "slope": 2.5},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	trend, err := client.QuotaHistory.GetUsageTrend(context.Background(), 1, "cpu", 7)
	assert.NoError(t, err)
	assert.NotNil(t, trend)
}

func TestQuotaHistoryService_DetectUsagePatterns(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/detect-patterns")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"organization_id": 1,
				"analysis_date":   "2024-01-01",
				"patterns": []map[string]interface{}{
					{
						"pattern_type":      "high_utilization",
						"description":       "High CPU usage detected",
						"severity":          "warning",
						"affected_resource": "cpu",
						"detected_value":    85.5,
						"threshold_value":   80.0,
						"recommendation":    "Consider upgrading CPU resources",
					},
				},
				"pattern_count": 1,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	patterns, err := client.QuotaHistory.DetectUsagePatterns(context.Background(), 1, 7)
	assert.NoError(t, err)
	assert.NotNil(t, patterns)
}

func TestQuotaHistoryService_CleanupOldRecords(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Contains(t, r.URL.Path, "/cleanup")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Cleanup complete",
			"data": map[string]interface{}{
				"deleted_count":  150,
				"retention_days": 90,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	deletedCount, err := client.QuotaHistory.CleanupOldRecords(context.Background(), 1, 90)
	assert.NoError(t, err)
	assert.Equal(t, 150, deletedCount)
}

func TestQuotaHistoryService_Errors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Internal error",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	now := time.Now()

	err := client.QuotaHistory.RecordQuotaUsage(context.Background(), []QuotaUsageRecord{})
	assert.Error(t, err)

	_, err = client.QuotaHistory.GetHistoricalUsage(context.Background(), 1, "cpu", now, now)
	assert.Error(t, err)

	_, err = client.QuotaHistory.GetAverageUtilization(context.Background(), 1, "cpu", now, now)
	assert.Error(t, err)

	_, err = client.QuotaHistory.GetPeakUtilization(context.Background(), 1, "cpu", now, now)
	assert.Error(t, err)

	_, err = client.QuotaHistory.GetDailyAggregates(context.Background(), 1, "cpu", now, now)
	assert.Error(t, err)

	_, err = client.QuotaHistory.GetResourceSummary(context.Background(), 1, now, now)
	assert.Error(t, err)

	_, err = client.QuotaHistory.GetUsageTrend(context.Background(), 1, "cpu", 7)
	assert.Error(t, err)

	_, err = client.QuotaHistory.DetectUsagePatterns(context.Background(), 1, 7)
	assert.Error(t, err)

	_, err = client.QuotaHistory.CleanupOldRecords(context.Background(), 1, 90)
	assert.Error(t, err)
}
