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

// Comprehensive coverage tests for health.go (system health service)
// Focus on improving List (84.6%) and GetHistory (89.3%)

// GetHealth tests (87.5% - type assertion)

func TestHealthService_GetHealth_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Service unavailable",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	health, err := client.Health.GetHealth(context.Background())
	assert.Error(t, err)
	assert.Nil(t, health)
}

// GetHealthDetailed tests (87.5% - type assertion)

func TestHealthService_GetHealthDetailed_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Service unavailable",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	health, err := client.Health.GetHealthDetailed(context.Background())
	assert.Error(t, err)
	assert.Nil(t, health)
}

// List tests (84.6% - needs comprehensive filter coverage)

func TestHealthService_List_AllFilters(t *testing.T) {
	serverID := uint(10)
	checkType := "disk"
	isEnabled := true

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/health/checks", r.URL.Path)

		q := r.URL.Query()
		assert.Equal(t, "10", q.Get("server_id"))
		assert.Equal(t, "disk", q.Get("check_type"))
		assert.Equal(t, "true", q.Get("is_enabled"))
		assert.Equal(t, "2", q.Get("page"))
		assert.Equal(t, "50", q.Get("limit"))
		assert.Equal(t, "name", q.Get("sort"))
		assert.Equal(t, "asc", q.Get("order"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []HealthCheck{},
			"meta": PaginationMeta{Page: 2, Limit: 50},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	checks, meta, err := client.Health.List(context.Background(), &HealthCheckListOptions{
		ServerID:  &serverID,
		CheckType: &checkType,
		IsEnabled: &isEnabled,
		ListOptions: ListOptions{
			Page:  2,
			Limit: 50,
			Sort:  "name",
			Order: "asc",
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, checks)
	assert.NotNil(t, meta)
}

func TestHealthService_List_IsEnabledFalse(t *testing.T) {
	isEnabled := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assert.Equal(t, "false", q.Get("is_enabled"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []HealthCheck{},
			"meta": PaginationMeta{},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	checks, meta, err := client.Health.List(context.Background(), &HealthCheckListOptions{
		IsEnabled: &isEnabled,
	})
	assert.NoError(t, err)
	assert.NotNil(t, checks)
	assert.NotNil(t, meta)
}

func TestHealthService_List_NilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.URL.Query().Get("server_id"))
		assert.Empty(t, r.URL.Query().Get("check_type"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []HealthCheck{},
			"meta": PaginationMeta{},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	checks, meta, err := client.Health.List(context.Background(), nil)
	assert.NoError(t, err)
	assert.NotNil(t, checks)
	assert.NotNil(t, meta)
}

func TestHealthService_List_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Unauthorized",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	checks, meta, err := client.Health.List(context.Background(), nil)
	assert.Error(t, err)
	assert.Nil(t, checks)
	assert.Nil(t, meta)
}

// Get tests (87.5% - type assertion)

func TestHealthService_Get_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Health check not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	check, err := client.Health.Get(context.Background(), 999)
	assert.Error(t, err)
	assert.Nil(t, check)
}

// Create tests (87.5% - type assertion)

func TestHealthService_Create_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Invalid health check configuration",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	check, err := client.Health.Create(context.Background(), &CreateHealthCheckRequest{
		ServerID:      1,
		CheckName:     "Test Check",
		CheckType:     "disk",
		CheckInterval: 5,
		CheckTimeout:  30,
	})
	assert.Error(t, err)
	assert.Nil(t, check)
}

// Update tests (87.5% - type assertion)

func TestHealthService_Update_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Health check not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	checkName := "Updated Check"
	check, err := client.Health.Update(context.Background(), 999, &UpdateHealthCheckRequest{
		CheckName: &checkName,
	})
	assert.Error(t, err)
	assert.Nil(t, check)
}

// GetHistory tests (89.3% - needs comprehensive filter coverage)

func TestHealthService_GetHistory_AllFilters(t *testing.T) {
	healthCheckID := uint(5)
	serverID := uint(10)
	status := "critical"
	fromDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	toDate := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/health/history", r.URL.Path)

		q := r.URL.Query()
		assert.Equal(t, "5", q.Get("health_check_id"))
		assert.Equal(t, "10", q.Get("server_id"))
		assert.Equal(t, "critical", q.Get("status"))
		assert.NotEmpty(t, q.Get("from_date"))
		assert.NotEmpty(t, q.Get("to_date"))
		assert.Equal(t, "3", q.Get("page"))
		assert.Equal(t, "100", q.Get("limit"))
		assert.Equal(t, "created_at", q.Get("sort"))
		assert.Equal(t, "desc", q.Get("order"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []HealthCheckHistory{},
			"meta": PaginationMeta{Page: 3, Limit: 100},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	history, meta, err := client.Health.GetHistory(context.Background(), &HealthCheckHistoryListOptions{
		HealthCheckID: &healthCheckID,
		ServerID:      &serverID,
		Status:        &status,
		FromDate:      &fromDate,
		ToDate:        &toDate,
		ListOptions: ListOptions{
			Page:  3,
			Limit: 100,
			Sort:  "created_at",
			Order: "desc",
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, history)
	assert.NotNil(t, meta)
}

func TestHealthService_GetHistory_NilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.URL.Query().Get("health_check_id"))
		assert.Empty(t, r.URL.Query().Get("status"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []HealthCheckHistory{},
			"meta": PaginationMeta{},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	history, meta, err := client.Health.GetHistory(context.Background(), nil)
	assert.NoError(t, err)
	assert.NotNil(t, history)
	assert.NotNil(t, meta)
}

func TestHealthService_GetHistory_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Insufficient permissions",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	history, meta, err := client.Health.GetHistory(context.Background(), nil)
	assert.Error(t, err)
	assert.Nil(t, history)
	assert.Nil(t, meta)
}

// Success path tests

func TestHealthService_Create_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id":             1,
				"check_name":     "Disk Check",
				"check_type":     "disk",
				"server_id":      1,
				"check_interval": 5,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	check, err := client.Health.Create(context.Background(), &CreateHealthCheckRequest{
		ServerID:      1,
		CheckName:     "Disk Check",
		CheckType:     "disk",
		CheckInterval: 5,
		CheckTimeout:  30,
	})
	assert.NoError(t, err)
	assert.NotNil(t, check)
}

func TestHealthService_Update_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"id":         1,
				"check_name": "Updated Disk Check",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	checkName := "Updated Disk Check"
	check, err := client.Health.Update(context.Background(), 1, &UpdateHealthCheckRequest{
		CheckName: &checkName,
	})
	assert.NoError(t, err)
	assert.NotNil(t, check)
}
