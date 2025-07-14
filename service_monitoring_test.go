package nexmonyx_test

import (
	"testing"
	"time"

	"github.com/nexmonyx/go-sdk"
	"github.com/stretchr/testify/assert"
)

func TestServiceInfo(t *testing.T) {
	t.Run("NewServiceInfo", func(t *testing.T) {
		info := nexmonyx.NewServiceInfo()
		assert.NotNil(t, info)
		assert.NotNil(t, info.Services)
		assert.NotNil(t, info.Metrics)
		assert.NotNil(t, info.Logs)
		assert.Equal(t, 0, len(info.Services))
	})

	t.Run("AddService", func(t *testing.T) {
		info := nexmonyx.NewServiceInfo()
		activeSince := time.Now()
		
		service := &nexmonyx.ServiceMonitoringInfo{
			Name:          "test.service",
			State:         "active",
			SubState:      "running",
			LoadState:     "loaded",
			Description:   "Test Service",
			MainPID:       1234,
			MemoryCurrent: 1048576,
			ActiveSince:   &activeSince,
		}
		
		info.AddService(service)
		assert.Equal(t, 1, len(info.Services))
		assert.Equal(t, "test.service", info.Services[0].Name)
	})

	t.Run("GetServiceByName", func(t *testing.T) {
		info := nexmonyx.NewServiceInfo()
		
		info.AddService(&nexmonyx.ServiceMonitoringInfo{Name: "service1"})
		info.AddService(&nexmonyx.ServiceMonitoringInfo{Name: "service2"})
		
		found := info.GetServiceByName("service1")
		assert.NotNil(t, found)
		assert.Equal(t, "service1", found.Name)
		
		notFound := info.GetServiceByName("service3")
		assert.Nil(t, notFound)
	})

	t.Run("GetFailedServices", func(t *testing.T) {
		info := nexmonyx.NewServiceInfo()
		
		info.AddService(&nexmonyx.ServiceMonitoringInfo{Name: "service1", State: "active"})
		info.AddService(&nexmonyx.ServiceMonitoringInfo{Name: "service2", State: "failed"})
		info.AddService(&nexmonyx.ServiceMonitoringInfo{Name: "service3", State: "failed"})
		info.AddService(&nexmonyx.ServiceMonitoringInfo{Name: "service4", State: "inactive"})
		
		failed := info.GetFailedServices()
		assert.Equal(t, 2, len(failed))
		for _, service := range failed {
			assert.Equal(t, "failed", service.State)
		}
	})

	t.Run("CountServicesByState", func(t *testing.T) {
		info := nexmonyx.NewServiceInfo()
		
		info.AddService(&nexmonyx.ServiceMonitoringInfo{Name: "s1", State: "active"})
		info.AddService(&nexmonyx.ServiceMonitoringInfo{Name: "s2", State: "active"})
		info.AddService(&nexmonyx.ServiceMonitoringInfo{Name: "s3", State: "failed"})
		info.AddService(&nexmonyx.ServiceMonitoringInfo{Name: "s4", State: "inactive"})
		
		counts := info.CountServicesByState()
		assert.Equal(t, 2, counts["active"])
		assert.Equal(t, 1, counts["failed"])
		assert.Equal(t, 1, counts["inactive"])
	})

	t.Run("CalculateTotalMemoryUsage", func(t *testing.T) {
		info := nexmonyx.NewServiceInfo()
		
		info.AddService(&nexmonyx.ServiceMonitoringInfo{Name: "s1", MemoryCurrent: 1000000})
		info.AddService(&nexmonyx.ServiceMonitoringInfo{Name: "s2", MemoryCurrent: 2000000})
		info.AddService(&nexmonyx.ServiceMonitoringInfo{Name: "s3", MemoryCurrent: 3000000})
		
		total := info.CalculateTotalMemoryUsage()
		assert.Equal(t, uint64(6000000), total)
	})
}

func TestServiceMetrics(t *testing.T) {
	t.Run("AddMetrics", func(t *testing.T) {
		info := nexmonyx.NewServiceInfo()
		
		metrics := &nexmonyx.ServiceMetrics{
			ServiceName:  "test.service",
			Timestamp:    time.Now(),
			CPUPercent:   5.5,
			MemoryRSS:    2048576,
			ProcessCount: 1,
			ThreadCount:  4,
		}
		
		info.AddMetrics(metrics)
		assert.Equal(t, 1, len(info.Metrics))
		assert.Equal(t, "test.service", info.Metrics[0].ServiceName)
		assert.Equal(t, 5.5, info.Metrics[0].CPUPercent)
	})

	t.Run("GetServiceMetrics", func(t *testing.T) {
		info := nexmonyx.NewServiceInfo()
		now := time.Now()
		
		// Add older metrics
		info.AddMetrics(&nexmonyx.ServiceMetrics{
			ServiceName: "test.service",
			Timestamp:   now.Add(-1 * time.Hour),
			CPUPercent:  10.0,
		})
		
		// Add newer metrics
		info.AddMetrics(&nexmonyx.ServiceMetrics{
			ServiceName: "test.service",
			Timestamp:   now,
			CPUPercent:  5.0,
		})
		
		// Should return the newer metrics
		latest := info.GetServiceMetrics("test.service")
		assert.NotNil(t, latest)
		assert.Equal(t, 5.0, latest.CPUPercent)
	})
}

func TestServiceLogs(t *testing.T) {
	t.Run("AddLogEntry", func(t *testing.T) {
		info := nexmonyx.NewServiceInfo()
		
		logEntry := nexmonyx.ServiceLogEntry{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   "Service started",
			Fields: map[string]string{
				"pid": "1234",
			},
		}
		
		info.AddLogEntry("test.service", logEntry)
		logs := info.GetServiceLogs("test.service")
		assert.Equal(t, 1, len(logs))
		assert.Equal(t, "Service started", logs[0].Message)
	})

	t.Run("GetErrorLogs", func(t *testing.T) {
		info := nexmonyx.NewServiceInfo()
		
		// Add various log levels
		info.AddLogEntry("service1", nexmonyx.ServiceLogEntry{Level: "info", Message: "Info message"})
		info.AddLogEntry("service1", nexmonyx.ServiceLogEntry{Level: "error", Message: "Error message"})
		info.AddLogEntry("service2", nexmonyx.ServiceLogEntry{Level: "warning", Message: "Warning message"})
		info.AddLogEntry("service2", nexmonyx.ServiceLogEntry{Level: "err", Message: "Err message"})
		info.AddLogEntry("service3", nexmonyx.ServiceLogEntry{Level: "ERROR", Message: "ERROR message"})
		
		errorLogs := info.GetErrorLogs()
		assert.Equal(t, 3, len(errorLogs))
		assert.Equal(t, 1, len(errorLogs["service1"]))
		assert.Equal(t, 1, len(errorLogs["service2"]))
		assert.Equal(t, 1, len(errorLogs["service3"]))
	})
}

func TestServiceMonitoringConfig(t *testing.T) {
	t.Run("NewServiceMonitoringConfig", func(t *testing.T) {
		config := nexmonyx.NewServiceMonitoringConfig()
		assert.True(t, config.Enabled)
		assert.True(t, config.CollectMetrics)
		assert.True(t, config.CollectLogs)
		assert.Equal(t, 100, config.LogLines)
		assert.Equal(t, "60s", config.MetricsInterval)
		assert.Contains(t, config.IncludePatterns, "ssh*")
		assert.Contains(t, config.ExcludePatterns, "*.scope")
	})

	t.Run("ShouldMonitorService", func(t *testing.T) {
		config := nexmonyx.NewServiceMonitoringConfig()
		
		// Test default patterns
		assert.True(t, config.ShouldMonitorService("ssh.service"))
		assert.True(t, config.ShouldMonitorService("systemd-resolved.service"))
		assert.False(t, config.ShouldMonitorService("user@1000.service"))
		assert.False(t, config.ShouldMonitorService("test-debug"))
		
		// Test explicit includes
		config.IncludeServices = []string{"nginx.service", "mysql.service"}
		assert.True(t, config.ShouldMonitorService("nginx.service"))
		assert.True(t, config.ShouldMonitorService("mysql.service"))
		
		// Test explicit excludes
		config.ExcludeServices = []string{"nginx.service"}
		assert.False(t, config.ShouldMonitorService("nginx.service"))
		assert.True(t, config.ShouldMonitorService("mysql.service"))
	})
}

func TestServiceHelpers(t *testing.T) {
	t.Run("GetServiceHealth", func(t *testing.T) {
		tests := []struct {
			name     string
			service  nexmonyx.ServiceMonitoringInfo
			expected int
		}{
			{
				name: "Active running with no restarts",
				service: nexmonyx.ServiceMonitoringInfo{
					State:        "active",
					SubState:     "running",
					RestartCount: 0,
				},
				expected: 100,
			},
			{
				name: "Active running with restarts",
				service: nexmonyx.ServiceMonitoringInfo{
					State:        "active",
					SubState:     "running",
					RestartCount: 3,
				},
				expected: 70,
			},
			{
				name: "Active but not running",
				service: nexmonyx.ServiceMonitoringInfo{
					State:    "active",
					SubState: "exited",
				},
				expected: 75,
			},
			{
				name: "Failed service",
				service: nexmonyx.ServiceMonitoringInfo{
					State: "failed",
				},
				expected: 0,
			},
			{
				name: "Inactive service",
				service: nexmonyx.ServiceMonitoringInfo{
					State: "inactive",
				},
				expected: 50,
			},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				health := nexmonyx.GetServiceHealth(&tt.service)
				assert.Equal(t, tt.expected, health)
			})
		}
	})

	t.Run("FormatServiceUptime", func(t *testing.T) {
		now := time.Now()
		
		// Test various uptimes
		tests := []struct {
			name     string
			since    *time.Time
			expected string
		}{
			{
				name:     "Nil time",
				since:    nil,
				expected: "N/A",
			},
			{
				name:     "Minutes only",
				since:    func() *time.Time { t := now.Add(-30 * time.Minute); return &t }(),
				expected: "30m",
			},
			{
				name:     "Hours and minutes",
				since:    func() *time.Time { t := now.Add(-90 * time.Minute); return &t }(),
				expected: "1h 30m",
			},
			{
				name:     "Days, hours and minutes",
				since:    func() *time.Time { t := now.Add(-25 * time.Hour); return &t }(),
				expected: "1d 1h 0m",
			},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uptime := nexmonyx.FormatServiceUptime(tt.since)
				assert.Equal(t, tt.expected, uptime)
			})
		}
	})

	t.Run("CreateServiceMetrics", func(t *testing.T) {
		metrics := nexmonyx.CreateServiceMetrics("test.service", 10.5, 1048576, 2, 8)
		
		assert.Equal(t, "test.service", metrics.ServiceName)
		assert.Equal(t, 10.5, metrics.CPUPercent)
		assert.Equal(t, uint64(1048576), metrics.MemoryRSS)
		assert.Equal(t, 2, metrics.ProcessCount)
		assert.Equal(t, 8, metrics.ThreadCount)
		assert.WithinDuration(t, time.Now(), metrics.Timestamp, 1*time.Second)
	})

	t.Run("CreateServiceLogEntry", func(t *testing.T) {
		log := nexmonyx.CreateServiceLogEntry("error", "Test error message")
		
		assert.Equal(t, "error", log.Level)
		assert.Equal(t, "Test error message", log.Message)
		assert.NotNil(t, log.Fields)
		assert.WithinDuration(t, time.Now(), log.Timestamp, 1*time.Second)
	})
}