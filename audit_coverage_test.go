package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// GetAuditLogs coverage tests - error paths and edge cases

func TestAuditService_GetAuditLogs_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Internal server error",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	logs, meta, err := client.Audit.GetAuditLogs(context.Background(), nil, nil)
	assert.Error(t, err)
	assert.Nil(t, logs)
	assert.Nil(t, meta)
}

func TestAuditService_GetAuditLogs_AllFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/audit/logs", r.URL.Path)

		// Verify all filter parameters
		q := r.URL.Query()
		assert.Equal(t, "1", q.Get("user_id"))
		assert.Equal(t, "create", q.Get("action"))
		assert.Equal(t, "server", q.Get("resource_type"))
		assert.Equal(t, "res-123", q.Get("resource_id"))
		assert.Equal(t, "2024-01-01", q.Get("start_date"))
		assert.Equal(t, "2024-12-31", q.Get("end_date"))
		assert.Equal(t, "high", q.Get("severity"))
		assert.Equal(t, "192.168.1.1", q.Get("ip_address"))
		assert.Equal(t, "1", q.Get("page"))
		assert.Equal(t, "25", q.Get("limit"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []AuditLog{},
			"meta": PaginationMeta{Page: 1, Limit: 25, TotalItems: 0, TotalPages: 0},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	opts := &PaginationOptions{Page: 1, Limit: 25}
	filters := map[string]interface{}{
		"user_id":       uint(1),
		"action":        "create",
		"resource_type": "server",
		"resource_id":   "res-123",
		"start_date":    "2024-01-01",
		"end_date":      "2024-12-31",
		"severity":      "high",
		"ip_address":    "192.168.1.1",
	}

	logs, meta, err := client.Audit.GetAuditLogs(context.Background(), opts, filters)
	assert.NoError(t, err)
	assert.NotNil(t, logs)
	assert.NotNil(t, meta)
}

func TestAuditService_GetAuditLogs_EmptyFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify filters with empty/zero values are not included
		q := r.URL.Query()
		assert.Empty(t, q.Get("user_id"))
		assert.Empty(t, q.Get("action"))
		assert.Empty(t, q.Get("resource_type"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []AuditLog{},
			"meta": PaginationMeta{},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	filters := map[string]interface{}{
		"user_id":       uint(0), // Zero value should be ignored
		"action":        "",      // Empty string should be ignored
		"resource_type": "",      // Empty string should be ignored
	}

	logs, meta, err := client.Audit.GetAuditLogs(context.Background(), nil, filters)
	assert.NoError(t, err)
	assert.NotNil(t, logs)
	assert.NotNil(t, meta)
}

// GetAuditLog coverage tests

func TestAuditService_GetAuditLog_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Audit log not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	log, err := client.Audit.GetAuditLog(context.Background(), 999)
	assert.Error(t, err)
	assert.Nil(t, log)
}

// ExportAuditLogs coverage tests - different formats and filters

func TestAuditService_ExportAuditLogs_CSV(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/audit/logs/export", r.URL.Path)

		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.NoError(t, err)
		assert.Equal(t, "csv", body["format"])

		w.Header().Set("Content-Type", "text/csv")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("id,action,user\n1,create,admin"))
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	data, err := client.Audit.ExportAuditLogs(context.Background(), "csv", nil)
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Contains(t, string(data), "id,action,user")
}

func TestAuditService_ExportAuditLogs_WithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.NoError(t, err)

		// Verify all filter fields
		assert.Equal(t, "json", body["format"])
		assert.Equal(t, float64(1), body["user_id"])
		assert.Equal(t, "delete", body["action"])
		assert.Equal(t, "vm", body["resource_type"])
		assert.Equal(t, "2024-01-01", body["start_date"])
		assert.Equal(t, "2024-12-31", body["end_date"])
		assert.Equal(t, "critical", body["severity"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"id": 1, "action": "delete"},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	filters := map[string]interface{}{
		"user_id":       uint(1),
		"action":        "delete",
		"resource_type": "vm",
		"start_date":    "2024-01-01",
		"end_date":      "2024-12-31",
		"severity":      "critical",
	}

	data, err := client.Audit.ExportAuditLogs(context.Background(), "json", filters)
	assert.NoError(t, err)
	assert.NotNil(t, data)
}

func TestAuditService_ExportAuditLogs_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Invalid format",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	data, err := client.Audit.ExportAuditLogs(context.Background(), "invalid", nil)
	assert.Error(t, err)
	assert.Nil(t, data)
}

// GetAuditStatistics coverage tests - with and without dates

func TestAuditService_GetAuditStatistics_WithDates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/audit/statistics", r.URL.Path)

		q := r.URL.Query()
		assert.Equal(t, "2024-01-01", q.Get("start_date"))
		assert.Equal(t, "2024-12-31", q.Get("end_date"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"total_logs": 100,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	stats, err := client.Audit.GetAuditStatistics(context.Background(), "2024-01-01", "2024-12-31")
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 100, stats.TotalLogs)
}

func TestAuditService_GetAuditStatistics_NoDates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assert.Empty(t, q.Get("start_date"))
		assert.Empty(t, q.Get("end_date"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"total_logs": 500,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	stats, err := client.Audit.GetAuditStatistics(context.Background(), "", "")
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 500, stats.TotalLogs)
}

func TestAuditService_GetAuditStatistics_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Failed to compute statistics",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	stats, err := client.Audit.GetAuditStatistics(context.Background(), "", "")
	assert.Error(t, err)
	assert.Nil(t, stats)
}

// GetUserAuditHistory coverage tests - with dates and pagination

func TestAuditService_GetUserAuditHistory_WithAllParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/audit/users/123/history", r.URL.Path)

		q := r.URL.Query()
		assert.Equal(t, "1", q.Get("page"))
		assert.Equal(t, "10", q.Get("limit"))
		assert.Equal(t, "2024-01-01", q.Get("start_date"))
		assert.Equal(t, "2024-12-31", q.Get("end_date"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []AuditLog{},
			"meta": PaginationMeta{Page: 1, Limit: 10},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	opts := &PaginationOptions{Page: 1, Limit: 10}
	logs, meta, err := client.Audit.GetUserAuditHistory(context.Background(), 123, opts, "2024-01-01", "2024-12-31")
	assert.NoError(t, err)
	assert.NotNil(t, logs)
	assert.NotNil(t, meta)
}

func TestAuditService_GetUserAuditHistory_NoDates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assert.Empty(t, q.Get("start_date"))
		assert.Empty(t, q.Get("end_date"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []AuditLog{},
			"meta": PaginationMeta{},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	logs, meta, err := client.Audit.GetUserAuditHistory(context.Background(), 123, nil, "", "")
	assert.NoError(t, err)
	assert.NotNil(t, logs)
	assert.NotNil(t, meta)
}

func TestAuditService_GetUserAuditHistory_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Access denied",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	logs, meta, err := client.Audit.GetUserAuditHistory(context.Background(), 123, nil, "", "")
	assert.Error(t, err)
	assert.Nil(t, logs)
	assert.Nil(t, meta)
}
