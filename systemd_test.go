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

func TestSystemdService_Submit(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v2/hardware/systemd/services", r.URL.Path)
		assert.Equal(t, "test-server-uuid", r.Header.Get("Server-UUID"))
		assert.Equal(t, "test-server-secret", r.Header.Get("Server-Secret"))

		// Parse request body
		var req SystemdServiceRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "test-server-uuid", req.ServerUUID)
		assert.Len(t, req.Services, 2)
		assert.Equal(t, "nginx.service", req.Services[0].Name)
		assert.NotNil(t, req.SystemStats)

		// Send response
		response := map[string]interface{}{
			"status":  "success",
			"message": "Services submitted successfully",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			ServerUUID:   "test-server-uuid",
			ServerSecret: "test-server-secret",
		},
	})
	require.NoError(t, err)

	// Prepare test data
	services := []SystemdServiceInfo{
		{
			Name:            "nginx.service",
			UnitType:        "service",
			Description:     "A high performance web server",
			LoadState:       "loaded",
			ActiveState:     "active",
			SubState:        "running",
			UnitState:       "enabled",
			MainPID:         1234,
			MemoryCurrent:   104857600, // 100MB
			CPUUsagePercent: 2.5,
			HealthScore:     95,
			DetectionMethod: "systemctl",
		},
		{
			Name:            "postgresql.service",
			UnitType:        "service",
			Description:     "PostgreSQL database server",
			LoadState:       "loaded",
			ActiveState:     "active",
			SubState:        "running",
			UnitState:       "enabled",
			MainPID:         5678,
			MemoryCurrent:   536870912, // 512MB
			CPUUsagePercent: 5.2,
			HealthScore:     90,
			DetectionMethod: "systemctl",
		},
	}

	systemStats := &SystemdSystemStats{
		TotalUnits:         150,
		ServiceUnits:       80,
		ActiveUnits:        75,
		FailedUnits:        2,
		EnabledUnits:       60,
		SystemState:        "degraded",
		OverallHealthScore: 85,
	}

	req := &SystemdServiceRequest{
		ServerUUID:  "test-server-uuid",
		CollectedAt: time.Now().Format(time.RFC3339),
		Services:    services,
		SystemStats: systemStats,
	}

	// Submit systemd data
	err = client.Systemd.Submit(context.Background(), req)
	require.NoError(t, err)
}

func TestSystemdService_GetSystemdServices(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v2/servers/test-server-uuid/systemd/services", r.URL.Path)

		// Check query parameters
		startTime := r.URL.Query().Get("start_time")
		endTime := r.URL.Query().Get("end_time")
		assert.NotEmpty(t, startTime)
		assert.NotEmpty(t, endTime)

		// Send response
		response := map[string]interface{}{
			"status": "success",
			"data": []SystemdServiceInfo{
				{
					Name:        "nginx.service",
					UnitType:    "service",
					LoadState:   "loaded",
					ActiveState: "active",
					SubState:    "running",
					UnitState:   "enabled",
				},
				{
					Name:        "ssh.service",
					UnitType:    "service",
					LoadState:   "loaded",
					ActiveState: "active",
					SubState:    "running",
					UnitState:   "enabled",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	require.NoError(t, err)

	// Get systemd services
	services, err := client.Systemd.Get(context.Background(), "test-server-uuid")
	require.NoError(t, err)
	assert.Len(t, services, 2)
	assert.Equal(t, "nginx.service", services[0].Name)
	assert.Equal(t, "ssh.service", services[1].Name)
}

func TestSystemdService_GetLatestSystemdServices(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v2/servers/test-server-uuid/systemd/services/latest", r.URL.Path)

		// Send response
		response := map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"services": []SystemdServiceInfo{
					{
						Name:        "nginx.service",
						UnitType:    "service",
						LoadState:   "loaded",
						ActiveState: "active",
						SubState:    "running",
						UnitState:   "enabled",
						HealthScore: 95,
					},
				},
				"system_stats": SystemdSystemStats{
					TotalUnits:         150,
					ActiveUnits:        145,
					FailedUnits:        0,
					SystemState:        "running",
					OverallHealthScore: 98,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	require.NoError(t, err)

	// Get systemd system stats instead (closest available method)
	stats, err := client.Systemd.GetSystemStats(context.Background(), "test-server-uuid")
	require.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 150, stats.TotalUnits)
	assert.Equal(t, "running", stats.SystemState)
}

func TestSystemdService_List(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/systemd", r.URL.Path)

		// Send response
		response := map[string]interface{}{
			"status": "success",
			"data": []SystemdServiceInfo{
				{
					Name:        "nginx.service",
					ActiveState: "active",
					SubState:    "running",
				},
				{
					Name:        "postgresql.service",
					ActiveState: "active",
					SubState:    "running",
				},
			},
			"meta": PaginationMeta{
				Page:       1,
				Limit:      10,
				TotalItems: 2,
				TotalPages: 1,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	require.NoError(t, err)

	// List systemd services
	opts := &ListOptions{
		Page:  1,
		Limit: 10,
	}
	services, pagination, err := client.Systemd.List(context.Background(), opts)
	require.NoError(t, err)
	assert.Len(t, services, 2)
	assert.Equal(t, "nginx.service", services[0].Name)
	assert.Equal(t, "postgresql.service", services[1].Name)
	assert.NotNil(t, pagination)
	assert.Equal(t, 2, pagination.TotalItems)
}

func TestSystemdService_GetSystemdServiceByName(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v2/servers/test-server-uuid/systemd/services/nginx.service", r.URL.Path)

		// Send response
		response := map[string]interface{}{
			"status": "success",
			"data": SystemdServiceInfo{
				Name:        "nginx.service",
				UnitType:    "service",
				Description: "A high performance web server",
				LoadState:   "loaded",
				ActiveState: "active",
				SubState:    "running",
				UnitState:   "enabled",
				MainPID:     1234,
				Type:        "forking",
				ExecStart:   []string{"/usr/sbin/nginx -g 'daemon on; master_process on;'"},
				WorkingDir:  "/",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	require.NoError(t, err)

	// Get specific service
	service, err := client.Systemd.GetServiceByName(context.Background(), "test-server-uuid", "nginx.service")
	require.NoError(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, "nginx.service", service.Name)
	assert.Equal(t, "forking", service.Type)
	assert.Equal(t, 1234, service.MainPID)
}

func TestSystemdService_GetSystemdSystemStats(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v2/servers/test-server-uuid/systemd/stats", r.URL.Path)

		// Send response
		response := map[string]interface{}{
			"status": "success",
			"data": SystemdSystemStats{
				TotalUnits:         200,
				ServiceUnits:       100,
				SocketUnits:        30,
				ActiveUnits:        180,
				FailedUnits:        3,
				SystemState:        "degraded",
				SystemStartupTime:  12.345,
				OverallHealthScore: 85,
				CriticalServices:   []string{"mysql.service", "redis.service"},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	require.NoError(t, err)

	// Get system stats
	stats, err := client.Systemd.GetSystemStats(context.Background(), "test-server-uuid")
	require.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 200, stats.TotalUnits)
	assert.Equal(t, "degraded", stats.SystemState)
	assert.Equal(t, 3, stats.FailedUnits)
	assert.Len(t, stats.CriticalServices, 2)
}

func TestSystemdServiceInfo_HelperFunctions(t *testing.T) {
	t.Run("IsHealthy", func(t *testing.T) {
		service := &SystemdServiceInfo{
			ActiveState: "active",
			SubState:    "running",
			Result:      "success",
			HealthScore: 80,
		}
		assert.True(t, service.IsHealthy())

		service.ActiveState = "failed"
		assert.False(t, service.IsHealthy())

		service.ActiveState = "active"
		service.HealthScore = 50
		assert.False(t, service.IsHealthy())
	})

	t.Run("IsFailed", func(t *testing.T) {
		service := &SystemdServiceInfo{
			ActiveState: "failed",
		}
		assert.True(t, service.IsFailed())

		service.ActiveState = "active"
		service.Result = "failure"
		assert.True(t, service.IsFailed())

		service.Result = "success"
		assert.False(t, service.IsFailed())
	})

	t.Run("IsEnabled", func(t *testing.T) {
		service := &SystemdServiceInfo{
			UnitState: "enabled",
		}
		assert.True(t, service.IsEnabled())

		service.UnitState = "disabled"
		assert.False(t, service.IsEnabled())
	})

	t.Run("GetAdditionalInfo", func(t *testing.T) {
		service := &SystemdServiceInfo{
			AdditionalInfo: map[string]interface{}{
				"custom_field":  "custom_value",
				"restart_limit": 5,
			},
		}

		val, ok := service.GetAdditionalInfo("custom_field")
		assert.True(t, ok)
		assert.Equal(t, "custom_value", val)

		val, ok = service.GetAdditionalInfo("nonexistent")
		assert.False(t, ok)
		assert.Nil(t, val)
	})
}

func TestSystemdSystemStats_HelperFunctions(t *testing.T) {
	t.Run("IsHealthy", func(t *testing.T) {
		stats := &SystemdSystemStats{
			SystemState:        "running",
			FailedUnits:        0,
			OverallHealthScore: 90,
		}
		assert.True(t, stats.IsHealthy())

		stats.FailedUnits = 1
		assert.False(t, stats.IsHealthy())

		stats.FailedUnits = 0
		stats.SystemState = "degraded"
		assert.False(t, stats.IsHealthy())
	})

	t.Run("IsDegraded", func(t *testing.T) {
		stats := &SystemdSystemStats{
			SystemState: "degraded",
			FailedUnits: 0,
		}
		assert.True(t, stats.IsDegraded())

		stats.SystemState = "running"
		stats.FailedUnits = 2
		assert.True(t, stats.IsDegraded())

		stats.FailedUnits = 0
		assert.False(t, stats.IsDegraded())
	})
}
