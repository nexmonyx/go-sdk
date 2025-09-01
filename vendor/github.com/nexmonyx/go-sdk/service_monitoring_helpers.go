package nexmonyx

import (
	"fmt"
	"strings"
	"time"
)

// NewServiceInfo creates a new ServiceInfo instance
func NewServiceInfo() *ServiceInfo {
	return &ServiceInfo{
		Services: make([]*ServiceMonitoringInfo, 0),
		Metrics:  make([]*ServiceMetrics, 0),
		Logs:     make(map[string][]ServiceLogEntry),
	}
}

// AddService adds a systemd service to the service info
func (s *ServiceInfo) AddService(service *ServiceMonitoringInfo) {
	if s.Services == nil {
		s.Services = make([]*ServiceMonitoringInfo, 0)
	}
	s.Services = append(s.Services, service)
}

// AddMetrics adds metrics for a service
func (s *ServiceInfo) AddMetrics(metrics *ServiceMetrics) {
	if s.Metrics == nil {
		s.Metrics = make([]*ServiceMetrics, 0)
	}
	s.Metrics = append(s.Metrics, metrics)
}

// AddLogEntry adds a log entry for a specific service
func (s *ServiceInfo) AddLogEntry(serviceName string, logEntry ServiceLogEntry) {
	if s.Logs == nil {
		s.Logs = make(map[string][]ServiceLogEntry)
	}
	s.Logs[serviceName] = append(s.Logs[serviceName], logEntry)
}

// GetServiceByName returns a service by its name
func (s *ServiceInfo) GetServiceByName(name string) *ServiceMonitoringInfo {
	for _, service := range s.Services {
		if service.Name == name {
			return service
		}
	}
	return nil
}

// GetFailedServices returns all services in failed state
func (s *ServiceInfo) GetFailedServices() []*ServiceMonitoringInfo {
	var failed []*ServiceMonitoringInfo
	for _, service := range s.Services {
		if service.State == "failed" {
			failed = append(failed, service)
		}
	}
	return failed
}

// GetActiveServices returns all active services
func (s *ServiceInfo) GetActiveServices() []*ServiceMonitoringInfo {
	var active []*ServiceMonitoringInfo
	for _, service := range s.Services {
		if service.State == "active" {
			active = append(active, service)
		}
	}
	return active
}

// GetServiceMetrics returns the latest metrics for a specific service
func (s *ServiceInfo) GetServiceMetrics(serviceName string) *ServiceMetrics {
	var latest *ServiceMetrics
	for _, metrics := range s.Metrics {
		if metrics.ServiceName == serviceName {
			if latest == nil || metrics.Timestamp.After(latest.Timestamp) {
				latest = metrics
			}
		}
	}
	return latest
}

// CountServicesByState returns a map of service states to counts
func (s *ServiceInfo) CountServicesByState() map[string]int {
	counts := make(map[string]int)
	for _, service := range s.Services {
		counts[service.State]++
	}
	return counts
}

// GetServiceLogs returns logs for a specific service
func (s *ServiceInfo) GetServiceLogs(serviceName string) []ServiceLogEntry {
	return s.Logs[serviceName]
}

// GetErrorLogs returns all error-level logs across all services
func (s *ServiceInfo) GetErrorLogs() map[string][]ServiceLogEntry {
	errorLogs := make(map[string][]ServiceLogEntry)
	for serviceName, logs := range s.Logs {
		for _, log := range logs {
			if strings.ToLower(log.Level) == "error" || strings.ToLower(log.Level) == "err" {
				errorLogs[serviceName] = append(errorLogs[serviceName], log)
			}
		}
	}
	return errorLogs
}

// CalculateTotalMemoryUsage calculates total memory usage across all services
func (s *ServiceInfo) CalculateTotalMemoryUsage() uint64 {
	var total uint64
	for _, service := range s.Services {
		total += service.MemoryCurrent
	}
	return total
}

// CalculateTotalCPUTime calculates total CPU time across all services
func (s *ServiceInfo) CalculateTotalCPUTime() time.Duration {
	var totalNanoseconds uint64
	for _, service := range s.Services {
		totalNanoseconds += service.CPUUsageNSec
	}
	return time.Duration(totalNanoseconds) * time.Nanosecond
}

// GetHighMemoryServices returns services using more than the specified memory threshold
func (s *ServiceInfo) GetHighMemoryServices(thresholdBytes uint64) []*ServiceMonitoringInfo {
	var highMemServices []*ServiceMonitoringInfo
	for _, service := range s.Services {
		if service.MemoryCurrent > thresholdBytes {
			highMemServices = append(highMemServices, service)
		}
	}
	return highMemServices
}

// GetRecentlyRestartedServices returns services that have been restarted
func (s *ServiceInfo) GetRecentlyRestartedServices() []*ServiceMonitoringInfo {
	var restarted []*ServiceMonitoringInfo
	for _, service := range s.Services {
		if service.RestartCount > 0 {
			restarted = append(restarted, service)
		}
	}
	return restarted
}

// NewServiceMonitoringConfig creates a new service monitoring configuration with defaults
func NewServiceMonitoringConfig() *ServiceMonitoringConfig {
	return &ServiceMonitoringConfig{
		Enabled:         true,
		IncludeServices: []string{},
		ExcludeServices: []string{},
		IncludePatterns: []string{"ssh*", "systemd*", "network*", "cron*"},
		ExcludePatterns: []string{"*-debug", "test-*", "*.scope", "*.slice"},
		CollectMetrics:  true,
		CollectLogs:     true,
		LogLines:        100,
		MetricsInterval: "60s",
		LogStateFile:    "/var/lib/nexmonyx/service-log-state",
	}
}

// ShouldMonitorService determines if a service should be monitored based on configuration
func (c *ServiceMonitoringConfig) ShouldMonitorService(serviceName string) bool {
	// Check explicit excludes first
	for _, exclude := range c.ExcludeServices {
		if serviceName == exclude {
			return false
		}
	}
	
	// Check exclude patterns
	for _, pattern := range c.ExcludePatterns {
		if matchPattern(serviceName, pattern) {
			return false
		}
	}
	
	// Check explicit includes
	for _, include := range c.IncludeServices {
		if serviceName == include {
			return true
		}
	}
	
	// Check include patterns
	for _, pattern := range c.IncludePatterns {
		if matchPattern(serviceName, pattern) {
			return true
		}
	}
	
	// If no patterns match and we have include patterns, don't monitor
	if len(c.IncludePatterns) > 0 || len(c.IncludeServices) > 0 {
		return false
	}
	
	// Default to monitoring if no rules apply
	return true
}

// matchPattern performs simple glob pattern matching
func matchPattern(name, pattern string) bool {
	// Simple implementation - can be enhanced
	if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
		return strings.Contains(name, pattern[1:len(pattern)-1])
	} else if strings.HasPrefix(pattern, "*") {
		return strings.HasSuffix(name, pattern[1:])
	} else if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(name, pattern[:len(pattern)-1])
	}
	return name == pattern
}

// CreateServiceMetrics creates a ServiceMetrics instance from current data
func CreateServiceMetrics(serviceName string, cpuPercent float64, memoryRSS uint64, processCount, threadCount int) *ServiceMetrics {
	return &ServiceMetrics{
		ServiceName:  serviceName,
		Timestamp:    time.Now(),
		CPUPercent:   cpuPercent,
		MemoryRSS:    memoryRSS,
		ProcessCount: processCount,
		ThreadCount:  threadCount,
	}
}

// CreateServiceLogEntry creates a ServiceLogEntry
func CreateServiceLogEntry(level, message string) ServiceLogEntry {
	return ServiceLogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Fields:    make(map[string]string),
	}
}

// FormatServiceUptime formats the service uptime in a human-readable format
func FormatServiceUptime(activeSince *time.Time) string {
	if activeSince == nil {
		return "N/A"
	}
	
	duration := time.Since(*activeSince)
	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60
	
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

// GetServiceHealth returns a health score (0-100) based on service state
func GetServiceHealth(service *ServiceMonitoringInfo) int {
	switch service.State {
	case "active":
		if service.SubState == "running" {
			// Penalize for restarts
			health := 100 - (service.RestartCount * 10)
			if health < 0 {
				health = 0
			}
			return health
		}
		return 75 // Active but not running
	case "inactive":
		return 50 // Service is stopped
	case "failed":
		return 0 // Service has failed
	default:
		return 25 // Unknown state
	}
}