package nexmonyx

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewServiceInfo tests the NewServiceInfo constructor
func TestNewServiceInfo(t *testing.T) {
	serviceInfo := NewServiceInfo()

	require.NotNil(t, serviceInfo)
	assert.NotNil(t, serviceInfo.Services)
	assert.NotNil(t, serviceInfo.Metrics)
	assert.NotNil(t, serviceInfo.Logs)
	assert.Len(t, serviceInfo.Services, 0)
	assert.Len(t, serviceInfo.Metrics, 0)
	assert.Len(t, serviceInfo.Logs, 0)
}

// TestServiceInfo_AddService tests adding services to ServiceInfo
func TestServiceInfo_AddService(t *testing.T) {
	tests := []struct {
		name            string
		initialServices []*ServiceMonitoringInfo
		serviceToAdd    *ServiceMonitoringInfo
		expectedCount   int
	}{
		{
			name:            "add service to empty list",
			initialServices: nil,
			serviceToAdd: &ServiceMonitoringInfo{
				Name:  "nginx",
				State: "active",
			},
			expectedCount: 1,
		},
		{
			name: "add service to existing list",
			initialServices: []*ServiceMonitoringInfo{
				{Name: "ssh", State: "active"},
			},
			serviceToAdd: &ServiceMonitoringInfo{
				Name:  "nginx",
				State: "active",
			},
			expectedCount: 2,
		},
		{
			name:            "add service with detailed info",
			initialServices: nil,
			serviceToAdd: &ServiceMonitoringInfo{
				Name:          "postgresql",
				State:         "active",
				SubState:      "running",
				Description:   "PostgreSQL Database",
				MemoryCurrent: 1024000,
				CPUUsageNSec:  5000000000,
				RestartCount:  0,
			},
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceInfo := &ServiceInfo{
				Services: tt.initialServices,
			}

			serviceInfo.AddService(tt.serviceToAdd)

			assert.Len(t, serviceInfo.Services, tt.expectedCount)
			assert.Contains(t, serviceInfo.Services, tt.serviceToAdd)
		})
	}
}

// TestServiceInfo_AddMetrics tests adding metrics to ServiceInfo
func TestServiceInfo_AddMetrics(t *testing.T) {
	tests := []struct {
		name           string
		initialMetrics []*ServiceMetrics
		metricsToAdd   *ServiceMetrics
		expectedCount  int
	}{
		{
			name:           "add metrics to empty list",
			initialMetrics: nil,
			metricsToAdd: &ServiceMetrics{
				ServiceName:  "nginx",
				CPUPercent:   25.5,
				MemoryRSS:    500000,
				ProcessCount: 4,
				ThreadCount:  16,
			},
			expectedCount: 1,
		},
		{
			name: "add metrics to existing list",
			initialMetrics: []*ServiceMetrics{
				{ServiceName: "ssh", CPUPercent: 1.5},
			},
			metricsToAdd: &ServiceMetrics{
				ServiceName:  "nginx",
				CPUPercent:   25.5,
				MemoryRSS:    500000,
				ProcessCount: 4,
				ThreadCount:  16,
			},
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceInfo := &ServiceInfo{
				Metrics: tt.initialMetrics,
			}

			serviceInfo.AddMetrics(tt.metricsToAdd)

			assert.Len(t, serviceInfo.Metrics, tt.expectedCount)
			assert.Contains(t, serviceInfo.Metrics, tt.metricsToAdd)
		})
	}
}

// TestServiceInfo_AddLogEntry tests adding log entries
func TestServiceInfo_AddLogEntry(t *testing.T) {
	tests := []struct {
		name         string
		initialLogs  map[string][]ServiceLogEntry
		serviceName  string
		logEntry     ServiceLogEntry
		expectedLogs int
	}{
		{
			name:        "add log entry to empty map",
			initialLogs: nil,
			serviceName: "nginx",
			logEntry: ServiceLogEntry{
				Timestamp: time.Now(),
				Level:     "info",
				Message:   "Service started",
			},
			expectedLogs: 1,
		},
		{
			name: "add log entry to existing service",
			initialLogs: map[string][]ServiceLogEntry{
				"nginx": {
					{Level: "info", Message: "First log"},
				},
			},
			serviceName: "nginx",
			logEntry: ServiceLogEntry{
				Timestamp: time.Now(),
				Level:     "error",
				Message:   "Error occurred",
			},
			expectedLogs: 2,
		},
		{
			name: "add log entry to new service",
			initialLogs: map[string][]ServiceLogEntry{
				"nginx": {{Level: "info", Message: "Nginx log"}},
			},
			serviceName: "postgresql",
			logEntry: ServiceLogEntry{
				Timestamp: time.Now(),
				Level:     "warning",
				Message:   "Slow query detected",
			},
			expectedLogs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceInfo := &ServiceInfo{
				Logs: tt.initialLogs,
			}

			serviceInfo.AddLogEntry(tt.serviceName, tt.logEntry)

			assert.Len(t, serviceInfo.Logs[tt.serviceName], tt.expectedLogs)
			assert.Contains(t, serviceInfo.Logs[tt.serviceName], tt.logEntry)
		})
	}
}

// TestServiceInfo_GetServiceByName tests service lookup by name
func TestServiceInfo_GetServiceByName(t *testing.T) {
	nginxService := &ServiceMonitoringInfo{Name: "nginx", State: "active"}
	sshService := &ServiceMonitoringInfo{Name: "ssh", State: "active"}

	tests := []struct {
		name         string
		services     []*ServiceMonitoringInfo
		searchName   string
		expectedName string
		shouldFind   bool
	}{
		{
			name:         "find existing service",
			services:     []*ServiceMonitoringInfo{nginxService, sshService},
			searchName:   "nginx",
			expectedName: "nginx",
			shouldFind:   true,
		},
		{
			name:       "service not found",
			services:   []*ServiceMonitoringInfo{nginxService, sshService},
			searchName: "postgresql",
			shouldFind: false,
		},
		{
			name:       "empty service list",
			services:   []*ServiceMonitoringInfo{},
			searchName: "nginx",
			shouldFind: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceInfo := &ServiceInfo{
				Services: tt.services,
			}

			result := serviceInfo.GetServiceByName(tt.searchName)

			if tt.shouldFind {
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedName, result.Name)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

// TestServiceInfo_GetFailedServices tests filtering failed services
func TestServiceInfo_GetFailedServices(t *testing.T) {
	tests := []struct {
		name          string
		services      []*ServiceMonitoringInfo
		expectedCount int
		expectedNames []string
	}{
		{
			name: "multiple failed services",
			services: []*ServiceMonitoringInfo{
				{Name: "nginx", State: "active"},
				{Name: "postgresql", State: "failed"},
				{Name: "redis", State: "failed"},
				{Name: "ssh", State: "active"},
			},
			expectedCount: 2,
			expectedNames: []string{"postgresql", "redis"},
		},
		{
			name: "no failed services",
			services: []*ServiceMonitoringInfo{
				{Name: "nginx", State: "active"},
				{Name: "ssh", State: "active"},
			},
			expectedCount: 0,
			expectedNames: []string{},
		},
		{
			name:          "empty service list",
			services:      []*ServiceMonitoringInfo{},
			expectedCount: 0,
			expectedNames: []string{},
		},
		{
			name: "all services failed",
			services: []*ServiceMonitoringInfo{
				{Name: "nginx", State: "failed"},
				{Name: "postgresql", State: "failed"},
			},
			expectedCount: 2,
			expectedNames: []string{"nginx", "postgresql"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceInfo := &ServiceInfo{
				Services: tt.services,
			}

			failed := serviceInfo.GetFailedServices()

			assert.Len(t, failed, tt.expectedCount)
			for _, expectedName := range tt.expectedNames {
				found := false
				for _, service := range failed {
					if service.Name == expectedName {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected to find service: %s", expectedName)
			}
		})
	}
}

// TestServiceInfo_GetActiveServices tests filtering active services
func TestServiceInfo_GetActiveServices(t *testing.T) {
	tests := []struct {
		name          string
		services      []*ServiceMonitoringInfo
		expectedCount int
		expectedNames []string
	}{
		{
			name: "mixed states",
			services: []*ServiceMonitoringInfo{
				{Name: "nginx", State: "active"},
				{Name: "postgresql", State: "failed"},
				{Name: "redis", State: "active"},
				{Name: "memcached", State: "inactive"},
			},
			expectedCount: 2,
			expectedNames: []string{"nginx", "redis"},
		},
		{
			name: "all active",
			services: []*ServiceMonitoringInfo{
				{Name: "nginx", State: "active"},
				{Name: "ssh", State: "active"},
			},
			expectedCount: 2,
			expectedNames: []string{"nginx", "ssh"},
		},
		{
			name: "no active services",
			services: []*ServiceMonitoringInfo{
				{Name: "nginx", State: "failed"},
				{Name: "ssh", State: "inactive"},
			},
			expectedCount: 0,
			expectedNames: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceInfo := &ServiceInfo{
				Services: tt.services,
			}

			active := serviceInfo.GetActiveServices()

			assert.Len(t, active, tt.expectedCount)
			for _, expectedName := range tt.expectedNames {
				found := false
				for _, service := range active {
					if service.Name == expectedName {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected to find service: %s", expectedName)
			}
		})
	}
}

// TestServiceInfo_GetServiceMetrics tests retrieving latest metrics for a service
func TestServiceInfo_GetServiceMetrics(t *testing.T) {
	now := time.Now()
	oldTime := now.Add(-1 * time.Hour)

	tests := []struct {
		name            string
		metrics         []*ServiceMetrics
		serviceName     string
		expectedCPU     float64
		shouldFindany   bool
	}{
		{
			name: "get latest metrics when multiple exist",
			metrics: []*ServiceMetrics{
				{ServiceName: "nginx", Timestamp: oldTime, CPUPercent: 10.0},
				{ServiceName: "nginx", Timestamp: now, CPUPercent: 25.5},
				{ServiceName: "ssh", Timestamp: now, CPUPercent: 1.5},
			},
			serviceName:   "nginx",
			expectedCPU:   25.5,
			shouldFindany: true,
		},
		{
			name: "service not found",
			metrics: []*ServiceMetrics{
				{ServiceName: "nginx", Timestamp: now, CPUPercent: 25.5},
			},
			serviceName:   "postgresql",
			shouldFindany: false,
		},
		{
			name: "single metrics for service",
			metrics: []*ServiceMetrics{
				{ServiceName: "redis", Timestamp: now, CPUPercent: 5.2},
			},
			serviceName:   "redis",
			expectedCPU:   5.2,
			shouldFindany: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceInfo := &ServiceInfo{
				Metrics: tt.metrics,
			}

			result := serviceInfo.GetServiceMetrics(tt.serviceName)

			if tt.shouldFindany {
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedCPU, result.CPUPercent)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

// TestServiceInfo_CountServicesByState tests counting services by state
func TestServiceInfo_CountServicesByState(t *testing.T) {
	tests := []struct {
		name          string
		services      []*ServiceMonitoringInfo
		expectedCount map[string]int
	}{
		{
			name: "mixed states",
			services: []*ServiceMonitoringInfo{
				{Name: "nginx", State: "active"},
				{Name: "postgresql", State: "active"},
				{Name: "redis", State: "failed"},
				{Name: "memcached", State: "inactive"},
				{Name: "mongodb", State: "active"},
			},
			expectedCount: map[string]int{
				"active":   3,
				"failed":   1,
				"inactive": 1,
			},
		},
		{
			name:          "empty service list",
			services:      []*ServiceMonitoringInfo{},
			expectedCount: map[string]int{},
		},
		{
			name: "all same state",
			services: []*ServiceMonitoringInfo{
				{Name: "nginx", State: "active"},
				{Name: "ssh", State: "active"},
			},
			expectedCount: map[string]int{
				"active": 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceInfo := &ServiceInfo{
				Services: tt.services,
			}

			counts := serviceInfo.CountServicesByState()

			assert.Equal(t, tt.expectedCount, counts)
		})
	}
}

// TestServiceInfo_GetServiceLogs tests retrieving logs for a service
func TestServiceInfo_GetServiceLogs(t *testing.T) {
	tests := []struct {
		name          string
		logs          map[string][]ServiceLogEntry
		serviceName   string
		expectedCount int
	}{
		{
			name: "get logs for existing service",
			logs: map[string][]ServiceLogEntry{
				"nginx": {
					{Level: "info", Message: "Started"},
					{Level: "error", Message: "Failed to connect"},
				},
				"ssh": {
					{Level: "info", Message: "Connection accepted"},
				},
			},
			serviceName:   "nginx",
			expectedCount: 2,
		},
		{
			name: "service with no logs",
			logs: map[string][]ServiceLogEntry{
				"nginx": {},
			},
			serviceName:   "nginx",
			expectedCount: 0,
		},
		{
			name: "service not in map",
			logs: map[string][]ServiceLogEntry{
				"nginx": {{Level: "info", Message: "Test"}},
			},
			serviceName:   "postgresql",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceInfo := &ServiceInfo{
				Logs: tt.logs,
			}

			logs := serviceInfo.GetServiceLogs(tt.serviceName)

			assert.Len(t, logs, tt.expectedCount)
		})
	}
}

// TestServiceInfo_GetErrorLogs tests filtering error logs
func TestServiceInfo_GetErrorLogs(t *testing.T) {
	tests := []struct {
		name               string
		logs               map[string][]ServiceLogEntry
		expectedServices   []string
		expectedErrorCount map[string]int
	}{
		{
			name: "multiple services with errors",
			logs: map[string][]ServiceLogEntry{
				"nginx": {
					{Level: "info", Message: "Started"},
					{Level: "error", Message: "Connection failed"},
					{Level: "err", Message: "Timeout"},
				},
				"ssh": {
					{Level: "info", Message: "OK"},
					{Level: "error", Message: "Authentication failed"},
				},
				"postgresql": {
					{Level: "info", Message: "Running"},
				},
			},
			expectedServices: []string{"nginx", "ssh"},
			expectedErrorCount: map[string]int{
				"nginx": 2,
				"ssh":   1,
			},
		},
		{
			name: "no error logs",
			logs: map[string][]ServiceLogEntry{
				"nginx": {
					{Level: "info", Message: "Started"},
					{Level: "warning", Message: "High load"},
				},
			},
			expectedServices:   []string{},
			expectedErrorCount: map[string]int{},
		},
		{
			name: "case insensitive error detection",
			logs: map[string][]ServiceLogEntry{
				"nginx": {
					{Level: "ERROR", Message: "Uppercase error"},
					{Level: "Error", Message: "Capitalized error"},
					{Level: "ERR", Message: "Uppercase err"},
				},
			},
			expectedServices: []string{"nginx"},
			expectedErrorCount: map[string]int{
				"nginx": 3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceInfo := &ServiceInfo{
				Logs: tt.logs,
			}

			errorLogs := serviceInfo.GetErrorLogs()

			assert.Len(t, errorLogs, len(tt.expectedServices))
			for service, expectedCount := range tt.expectedErrorCount {
				assert.Len(t, errorLogs[service], expectedCount,
					"Service %s should have %d error logs", service, expectedCount)
			}
		})
	}
}

// TestServiceInfo_CalculateTotalMemoryUsage tests memory calculation
func TestServiceInfo_CalculateTotalMemoryUsage(t *testing.T) {
	tests := []struct {
		name          string
		services      []*ServiceMonitoringInfo
		expectedTotal uint64
	}{
		{
			name: "multiple services with memory",
			services: []*ServiceMonitoringInfo{
				{Name: "nginx", MemoryCurrent: 1024000},
				{Name: "postgresql", MemoryCurrent: 5120000},
				{Name: "redis", MemoryCurrent: 512000},
			},
			expectedTotal: 6656000,
		},
		{
			name:          "empty service list",
			services:      []*ServiceMonitoringInfo{},
			expectedTotal: 0,
		},
		{
			name: "services with zero memory",
			services: []*ServiceMonitoringInfo{
				{Name: "nginx", MemoryCurrent: 0},
				{Name: "ssh", MemoryCurrent: 0},
			},
			expectedTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceInfo := &ServiceInfo{
				Services: tt.services,
			}

			total := serviceInfo.CalculateTotalMemoryUsage()

			assert.Equal(t, tt.expectedTotal, total)
		})
	}
}

// TestServiceInfo_CalculateTotalCPUTime tests CPU time calculation
func TestServiceInfo_CalculateTotalCPUTime(t *testing.T) {
	tests := []struct {
		name                string
		services            []*ServiceMonitoringInfo
		expectedNanoseconds uint64
	}{
		{
			name: "multiple services with CPU time",
			services: []*ServiceMonitoringInfo{
				{Name: "nginx", CPUUsageNSec: 1000000000},    // 1 second
				{Name: "postgresql", CPUUsageNSec: 5000000000}, // 5 seconds
				{Name: "redis", CPUUsageNSec: 500000000},     // 0.5 seconds
			},
			expectedNanoseconds: 6500000000, // 6.5 seconds
		},
		{
			name:                "empty service list",
			services:            []*ServiceMonitoringInfo{},
			expectedNanoseconds: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceInfo := &ServiceInfo{
				Services: tt.services,
			}

			total := serviceInfo.CalculateTotalCPUTime()

			assert.Equal(t, time.Duration(tt.expectedNanoseconds), total)
		})
	}
}

// TestServiceInfo_GetHighMemoryServices tests filtering high memory services
func TestServiceInfo_GetHighMemoryServices(t *testing.T) {
	tests := []struct {
		name          string
		services      []*ServiceMonitoringInfo
		threshold     uint64
		expectedCount int
		expectedNames []string
	}{
		{
			name: "some services above threshold",
			services: []*ServiceMonitoringInfo{
				{Name: "nginx", MemoryCurrent: 1000000},      // 1 MB
				{Name: "postgresql", MemoryCurrent: 10000000}, // 10 MB
				{Name: "redis", MemoryCurrent: 500000},       // 0.5 MB
			},
			threshold:     2000000, // 2 MB threshold
			expectedCount: 1,
			expectedNames: []string{"postgresql"},
		},
		{
			name: "no services above threshold",
			services: []*ServiceMonitoringInfo{
				{Name: "nginx", MemoryCurrent: 1000000},
				{Name: "ssh", MemoryCurrent: 500000},
			},
			threshold:     5000000,
			expectedCount: 0,
			expectedNames: []string{},
		},
		{
			name: "all services above threshold",
			services: []*ServiceMonitoringInfo{
				{Name: "nginx", MemoryCurrent: 5000000},
				{Name: "postgresql", MemoryCurrent: 10000000},
			},
			threshold:     1000000,
			expectedCount: 2,
			expectedNames: []string{"nginx", "postgresql"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceInfo := &ServiceInfo{
				Services: tt.services,
			}

			highMem := serviceInfo.GetHighMemoryServices(tt.threshold)

			assert.Len(t, highMem, tt.expectedCount)
			for _, expectedName := range tt.expectedNames {
				found := false
				for _, service := range highMem {
					if service.Name == expectedName {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected to find service: %s", expectedName)
			}
		})
	}
}

// TestServiceInfo_GetRecentlyRestartedServices tests filtering restarted services
func TestServiceInfo_GetRecentlyRestartedServices(t *testing.T) {
	tests := []struct {
		name          string
		services      []*ServiceMonitoringInfo
		expectedCount int
		expectedNames []string
	}{
		{
			name: "some services with restarts",
			services: []*ServiceMonitoringInfo{
				{Name: "nginx", RestartCount: 2},
				{Name: "postgresql", RestartCount: 0},
				{Name: "redis", RestartCount: 5},
			},
			expectedCount: 2,
			expectedNames: []string{"nginx", "redis"},
		},
		{
			name: "no services with restarts",
			services: []*ServiceMonitoringInfo{
				{Name: "nginx", RestartCount: 0},
				{Name: "ssh", RestartCount: 0},
			},
			expectedCount: 0,
			expectedNames: []string{},
		},
		{
			name: "all services with restarts",
			services: []*ServiceMonitoringInfo{
				{Name: "nginx", RestartCount: 1},
				{Name: "postgresql", RestartCount: 3},
			},
			expectedCount: 2,
			expectedNames: []string{"nginx", "postgresql"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceInfo := &ServiceInfo{
				Services: tt.services,
			}

			restarted := serviceInfo.GetRecentlyRestartedServices()

			assert.Len(t, restarted, tt.expectedCount)
			for _, expectedName := range tt.expectedNames {
				found := false
				for _, service := range restarted {
					if service.Name == expectedName {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected to find service: %s", expectedName)
			}
		})
	}
}

// TestNewServiceMonitoringConfig tests the config constructor
func TestNewServiceMonitoringConfig(t *testing.T) {
	config := NewServiceMonitoringConfig()

	require.NotNil(t, config)
	assert.True(t, config.Enabled)
	assert.NotNil(t, config.IncludeServices)
	assert.NotNil(t, config.ExcludeServices)
	assert.NotNil(t, config.IncludePatterns)
	assert.NotNil(t, config.ExcludePatterns)
	assert.True(t, config.CollectMetrics)
	assert.True(t, config.CollectLogs)
	assert.Equal(t, 100, config.LogLines)
	assert.Equal(t, "60s", config.MetricsInterval)
	assert.Equal(t, "/var/lib/nexmonyx/service-log-state", config.LogStateFile)
}

// TestServiceMonitoringConfig_ShouldMonitorService tests service filtering logic
func TestServiceMonitoringConfig_ShouldMonitorService(t *testing.T) {
	tests := []struct {
		name           string
		config         *ServiceMonitoringConfig
		serviceName    string
		shouldMonitor  bool
	}{
		{
			name: "explicit exclude takes precedence",
			config: &ServiceMonitoringConfig{
				IncludeServices: []string{"nginx"},
				ExcludeServices: []string{"nginx"},
			},
			serviceName:   "nginx",
			shouldMonitor: false,
		},
		{
			name: "exclude pattern matches",
			config: &ServiceMonitoringConfig{
				ExcludePatterns: []string{"*-debug", "test-*"},
			},
			serviceName:   "nginx-debug",
			shouldMonitor: false,
		},
		{
			name: "explicit include",
			config: &ServiceMonitoringConfig{
				IncludeServices: []string{"nginx", "ssh"},
			},
			serviceName:   "nginx",
			shouldMonitor: true,
		},
		{
			name: "include pattern matches",
			config: &ServiceMonitoringConfig{
				IncludePatterns: []string{"ssh*", "systemd*"},
			},
			serviceName:   "sshd",
			shouldMonitor: true,
		},
		{
			name: "no rules - default to monitor",
			config: &ServiceMonitoringConfig{
				IncludeServices: []string{},
				ExcludeServices: []string{},
				IncludePatterns: []string{},
				ExcludePatterns: []string{},
			},
			serviceName:   "nginx",
			shouldMonitor: true,
		},
		{
			name: "has include patterns but service doesn't match",
			config: &ServiceMonitoringConfig{
				IncludePatterns: []string{"ssh*"},
			},
			serviceName:   "nginx",
			shouldMonitor: false,
		},
		{
			name: "include pattern with wildcards on both ends matches",
			config: &ServiceMonitoringConfig{
				IncludePatterns: []string{"*daemon*", "ssh*"},
			},
			serviceName:   "my-daemon-service",
			shouldMonitor: true,
		},
		{
			name: "exclude pattern with wildcards on both ends matches",
			config: &ServiceMonitoringConfig{
				ExcludePatterns: []string{"*temp*"},
			},
			serviceName:   "my-temp-service",
			shouldMonitor: false,
		},
		{
			name: "include pattern with exact match (no wildcards)",
			config: &ServiceMonitoringConfig{
				IncludePatterns: []string{"nginx", "ssh*"},
			},
			serviceName:   "nginx",
			shouldMonitor: true,
		},
		{
			name: "exclude pattern with exact match (no wildcards)",
			config: &ServiceMonitoringConfig{
				ExcludePatterns: []string{"test-service"},
			},
			serviceName:   "test-service",
			shouldMonitor: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.ShouldMonitorService(tt.serviceName)
			assert.Equal(t, tt.shouldMonitor, result)
		})
	}
}

// TestCreateServiceMetrics tests metrics creation
func TestCreateServiceMetrics(t *testing.T) {
	metrics := CreateServiceMetrics("nginx", 25.5, 1024000, 4, 16)

	require.NotNil(t, metrics)
	assert.Equal(t, "nginx", metrics.ServiceName)
	assert.Equal(t, 25.5, metrics.CPUPercent)
	assert.Equal(t, uint64(1024000), metrics.MemoryRSS)
	assert.Equal(t, 4, metrics.ProcessCount)
	assert.Equal(t, 16, metrics.ThreadCount)
	assert.False(t, metrics.Timestamp.IsZero())
}

// TestCreateServiceLogEntry tests log entry creation
func TestCreateServiceLogEntry(t *testing.T) {
	entry := CreateServiceLogEntry("error", "Connection failed")

	require.NotNil(t, entry)
	assert.Equal(t, "error", entry.Level)
	assert.Equal(t, "Connection failed", entry.Message)
	assert.NotNil(t, entry.Fields)
	assert.False(t, entry.Timestamp.IsZero())
}

// TestFormatServiceUptime tests uptime formatting
func TestFormatServiceUptime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		activeSince *time.Time
		expected    string
		checkPrefix bool
		prefix      string
	}{
		{
			name:        "nil active since",
			activeSince: nil,
			expected:    "N/A",
		},
		{
			name: "multiple days",
			activeSince: func() *time.Time {
				t := now.Add(-50 * time.Hour)
				return &t
			}(),
			checkPrefix: true,
			prefix:      "2d",
		},
		{
			name: "hours only",
			activeSince: func() *time.Time {
				t := now.Add(-5 * time.Hour)
				return &t
			}(),
			checkPrefix: true,
			prefix:      "5h",
		},
		{
			name: "minutes only",
			activeSince: func() *time.Time {
				t := now.Add(-30 * time.Minute)
				return &t
			}(),
			expected: "30m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatServiceUptime(tt.activeSince)

			if tt.checkPrefix {
				assert.Contains(t, result, tt.prefix)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestGetServiceHealth tests health scoring
func TestGetServiceHealth(t *testing.T) {
	tests := []struct {
		name           string
		service        *ServiceMonitoringInfo
		expectedHealth int
	}{
		{
			name: "active running with no restarts",
			service: &ServiceMonitoringInfo{
				State:        "active",
				SubState:     "running",
				RestartCount: 0,
			},
			expectedHealth: 100,
		},
		{
			name: "active running with restarts",
			service: &ServiceMonitoringInfo{
				State:        "active",
				SubState:     "running",
				RestartCount: 3,
			},
			expectedHealth: 70, // 100 - (3 * 10)
		},
		{
			name: "active but not running",
			service: &ServiceMonitoringInfo{
				State:    "active",
				SubState: "exited",
			},
			expectedHealth: 75,
		},
		{
			name: "inactive service",
			service: &ServiceMonitoringInfo{
				State: "inactive",
			},
			expectedHealth: 50,
		},
		{
			name: "failed service",
			service: &ServiceMonitoringInfo{
				State: "failed",
			},
			expectedHealth: 0,
		},
		{
			name: "unknown state",
			service: &ServiceMonitoringInfo{
				State: "unknown",
			},
			expectedHealth: 25,
		},
		{
			name: "many restarts (health floor at 0)",
			service: &ServiceMonitoringInfo{
				State:        "active",
				SubState:     "running",
				RestartCount: 20, // Would be 100 - 200 = -100, but floored at 0
			},
			expectedHealth: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			health := GetServiceHealth(tt.service)
			assert.Equal(t, tt.expectedHealth, health)
		})
	}
}
