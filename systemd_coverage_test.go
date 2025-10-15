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

// Comprehensive coverage tests for systemd.go
// Focuses on achieving 100% coverage with correct API paths

// Submit tests

func TestSystemdService_Submit_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/systemd", r.URL.Path)

		var body SystemdServiceRequest
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.NoError(t, err)
		assert.Equal(t, "test-server-uuid", body.ServerUUID)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Systemd data submitted",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	err := client.Systemd.Submit(context.Background(), &SystemdServiceRequest{
		ServerUUID:  "test-server-uuid",
		CollectedAt: "2024-01-01T00:00:00Z",
		Services:    []SystemdServiceInfo{{Name: "nginx.service", ActiveState: "active"}},
	})
	assert.NoError(t, err)
}

func TestSystemdService_Submit_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Invalid systemd data",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	err := client.Systemd.Submit(context.Background(), &SystemdServiceRequest{
		ServerUUID: "invalid",
	})
	assert.Error(t, err)
}

// Get tests

func TestSystemdService_Get_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/systemd/test-server-uuid", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": []map[string]interface{}{
				{
					"name":         "nginx.service",
					"active_state": "active",
					"sub_state":    "running",
				},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	services, err := client.Systemd.Get(context.Background(), "test-server-uuid")
	assert.NoError(t, err)
	assert.NotNil(t, services)
	assert.Len(t, services, 1)
}

func TestSystemdService_Get_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Server not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	services, err := client.Systemd.Get(context.Background(), "nonexistent")
	assert.Error(t, err)
	assert.Nil(t, services)
}

// List tests

func TestSystemdService_List_NilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/systemd", r.URL.Path)

		// Verify no query parameters when opts is nil
		assert.Empty(t, r.URL.Query().Get("page"))
		assert.Empty(t, r.URL.Query().Get("limit"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []*SystemdServiceInfo{},
			"meta": PaginationMeta{Page: 1, Limit: 25},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	services, meta, err := client.Systemd.List(context.Background(), nil)
	assert.NoError(t, err)
	assert.NotNil(t, services)
	assert.NotNil(t, meta)
}

func TestSystemdService_List_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)

		q := r.URL.Query()
		assert.Equal(t, "2", q.Get("page"))
		assert.Equal(t, "50", q.Get("limit"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []*SystemdServiceInfo{
				{Name: "nginx.service", ActiveState: "active"},
			},
			"meta": PaginationMeta{Page: 2, Limit: 50},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	services, meta, err := client.Systemd.List(context.Background(), &ListOptions{
		Page:  2,
		Limit: 50,
	})
	assert.NoError(t, err)
	assert.NotNil(t, services)
	assert.NotNil(t, meta)
	assert.Len(t, services, 1)
}

func TestSystemdService_List_Error(t *testing.T) {
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
	services, meta, err := client.Systemd.List(context.Background(), nil)
	assert.Error(t, err)
	assert.Nil(t, services)
	assert.Nil(t, meta)
}

// GetServiceByName tests

func TestSystemdService_GetServiceByName_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/systemd/test-server-uuid/service/nginx.service", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"name":         "nginx.service",
				"active_state": "active",
				"sub_state":    "running",
				"health_score": 95,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	service, err := client.Systemd.GetServiceByName(context.Background(), "test-server-uuid", "nginx.service")
	assert.NoError(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, "nginx.service", service.Name)
}

func TestSystemdService_GetServiceByName_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Service not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	service, err := client.Systemd.GetServiceByName(context.Background(), "test-server-uuid", "nonexistent.service")
	assert.Error(t, err)
	assert.Nil(t, service)
}

// GetSystemStats tests

func TestSystemdService_GetSystemStats_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/systemd/test-server-uuid/stats", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"total_units":          150,
				"service_units":        50,
				"active_units":         120,
				"failed_units":         0,
				"system_state":         "running",
				"overall_health_score": 95,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	stats, err := client.Systemd.GetSystemStats(context.Background(), "test-server-uuid")
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 150, stats.TotalUnits)
	assert.Equal(t, "running", stats.SystemState)
}

func TestSystemdService_GetSystemStats_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Server not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	stats, err := client.Systemd.GetSystemStats(context.Background(), "nonexistent")
	assert.Error(t, err)
	assert.Nil(t, stats)
}

// SystemdServiceInfo helper method tests

func TestSystemdServiceInfo_IsHealthy(t *testing.T) {
	// Healthy service
	healthyService := &SystemdServiceInfo{
		ActiveState: "active",
		SubState:    "running",
		HealthScore: 85,
	}
	assert.True(t, healthyService.IsHealthy())

	// Not healthy - wrong active state
	notHealthy1 := &SystemdServiceInfo{
		ActiveState: "inactive",
		SubState:    "running",
		HealthScore: 85,
	}
	assert.False(t, notHealthy1.IsHealthy())

	// Not healthy - wrong sub state
	notHealthy2 := &SystemdServiceInfo{
		ActiveState: "active",
		SubState:    "dead",
		HealthScore: 85,
	}
	assert.False(t, notHealthy2.IsHealthy())

	// Not healthy - low health score
	notHealthy3 := &SystemdServiceInfo{
		ActiveState: "active",
		SubState:    "running",
		HealthScore: 50,
	}
	assert.False(t, notHealthy3.IsHealthy())
}

func TestSystemdServiceInfo_IsFailed(t *testing.T) {
	// Failed active state
	failedService1 := &SystemdServiceInfo{
		ActiveState: "failed",
		SubState:    "dead",
	}
	assert.True(t, failedService1.IsFailed())

	// Failed sub state
	failedService2 := &SystemdServiceInfo{
		ActiveState: "active",
		SubState:    "failed",
	}
	assert.True(t, failedService2.IsFailed())

	// Not failed
	healthyService := &SystemdServiceInfo{
		ActiveState: "active",
		SubState:    "running",
	}
	assert.False(t, healthyService.IsFailed())
}

func TestSystemdServiceInfo_IsEnabled(t *testing.T) {
	// Enabled service
	enabledService := &SystemdServiceInfo{
		UnitState: "enabled",
	}
	assert.True(t, enabledService.IsEnabled())

	// Disabled service
	disabledService := &SystemdServiceInfo{
		UnitState: "disabled",
	}
	assert.False(t, disabledService.IsEnabled())

	// Static service
	staticService := &SystemdServiceInfo{
		UnitState: "static",
	}
	assert.False(t, staticService.IsEnabled())
}

func TestSystemdServiceInfo_GetAdditionalInfo(t *testing.T) {
	// With additional info
	service := &SystemdServiceInfo{
		AdditionalInfo: map[string]interface{}{
			"custom_key":   "custom_value",
			"retry_count":  5,
			"last_restart": "2024-01-01T00:00:00Z",
		},
	}

	val, ok := service.GetAdditionalInfo("custom_key")
	assert.True(t, ok)
	assert.Equal(t, "custom_value", val)

	val2, ok2 := service.GetAdditionalInfo("retry_count")
	assert.True(t, ok2)
	assert.Equal(t, 5, val2)

	// Non-existent key
	val3, ok3 := service.GetAdditionalInfo("nonexistent")
	assert.False(t, ok3)
	assert.Nil(t, val3)

	// Nil additional info
	serviceNil := &SystemdServiceInfo{
		AdditionalInfo: nil,
	}
	val4, ok4 := serviceNil.GetAdditionalInfo("any_key")
	assert.False(t, ok4)
	assert.Nil(t, val4)
}

// SystemdSystemStats helper method tests

func TestSystemdSystemStats_IsHealthy(t *testing.T) {
	// Healthy system
	healthyStats := &SystemdSystemStats{
		SystemState:        "running",
		FailedUnits:        0,
		OverallHealthScore: 95,
	}
	assert.True(t, healthyStats.IsHealthy())

	// Not healthy - wrong system state
	notHealthy1 := &SystemdSystemStats{
		SystemState:        "degraded",
		FailedUnits:        0,
		OverallHealthScore: 95,
	}
	assert.False(t, notHealthy1.IsHealthy())

	// Not healthy - has failed units
	notHealthy2 := &SystemdSystemStats{
		SystemState:        "running",
		FailedUnits:        5,
		OverallHealthScore: 95,
	}
	assert.False(t, notHealthy2.IsHealthy())

	// Not healthy - low health score
	notHealthy3 := &SystemdSystemStats{
		SystemState:        "running",
		FailedUnits:        0,
		OverallHealthScore: 70,
	}
	assert.False(t, notHealthy3.IsHealthy())
}

func TestSystemdSystemStats_IsDegraded(t *testing.T) {
	// Degraded system state
	degraded1 := &SystemdSystemStats{
		SystemState: "degraded",
		FailedUnits: 0,
	}
	assert.True(t, degraded1.IsDegraded())

	// Has failed units
	degraded2 := &SystemdSystemStats{
		SystemState: "running",
		FailedUnits: 3,
	}
	assert.True(t, degraded2.IsDegraded())

	// Not degraded
	healthy := &SystemdSystemStats{
		SystemState: "running",
		FailedUnits: 0,
	}
	assert.False(t, healthy.IsDegraded())
}

// Comprehensive integration test

func TestSystemdService_CompleteWorkflow(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == "POST" && r.URL.Path == "/v1/systemd":
			// Submit
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "success",
			})

		case r.Method == "GET" && r.URL.Path == "/v1/systemd/test-server-uuid":
			// Get services
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "success",
				"data": []map[string]interface{}{
					{"name": "nginx.service", "active_state": "active"},
				},
			})

		case r.Method == "GET" && r.URL.Path == "/v1/systemd/test-server-uuid/stats":
			// Get stats
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"total_units":  150,
					"system_state": "running",
				},
			})

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	// 1. Submit systemd data
	err := client.Systemd.Submit(context.Background(), &SystemdServiceRequest{
		ServerUUID:  "test-server-uuid",
		CollectedAt: time.Now().Format(time.RFC3339),
		Services:    []SystemdServiceInfo{{Name: "nginx.service"}},
	})
	assert.NoError(t, err)

	// 2. Get services
	services, err := client.Systemd.Get(context.Background(), "test-server-uuid")
	assert.NoError(t, err)
	assert.Len(t, services, 1)

	// 3. Get system stats
	stats, err := client.Systemd.GetSystemStats(context.Background(), "test-server-uuid")
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 150, stats.TotalUnits)
}
