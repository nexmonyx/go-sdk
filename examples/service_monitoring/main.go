package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nexmonyx/go-sdk/v2"
)

func main() {
	// Initialize the SDK client with server credentials
	client, err := nexmonyx.NewClient(&nexmonyx.Config{
		BaseURL: "https://api.nexmonyx.com",
		Auth: nexmonyx.AuthConfig{
			ServerUUID:   "your-server-uuid",
			ServerSecret: "your-server-secret",
		},
		Debug: true,
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Example 1: Submit service monitoring data as part of comprehensive metrics
	submitServiceMonitoringData(ctx, client)

	// Example 2: Monitor critical services
	monitorCriticalServices()

	// Example 3: Analyze service logs
	analyzeServiceLogs()

	// Example 4: Track service resource usage
	trackServiceResources()
}

func submitServiceMonitoringData(ctx context.Context, client *nexmonyx.Client) {
	fmt.Println("=== Submitting Service Monitoring Data ===")

	// Create service info
	serviceInfo := nexmonyx.NewServiceInfo()

	// Add SSH service
	activeSince := time.Now().Add(-24 * time.Hour)
	serviceInfo.AddService(&nexmonyx.ServiceMonitoringInfo{
		Name:          "ssh.service",
		State:         "active",
		SubState:      "running",
		LoadState:     "loaded",
		Description:   "OpenBSD Secure Shell server",
		MainPID:       163754,
		MemoryCurrent: 4308992,
		CPUUsageNSec:  890000000,
		TasksCurrent:  1,
		RestartCount:  0,
		ActiveSince:   &activeSince,
	})

	// Add Nginx service (failed state)
	serviceInfo.AddService(&nexmonyx.ServiceMonitoringInfo{
		Name:         "nginx.service",
		State:        "failed",
		SubState:     "failed",
		LoadState:    "loaded",
		Description:  "The nginx HTTP and reverse proxy server",
		MainPID:      0,
		RestartCount: 3,
	})

	// Add Cron service
	cronActiveSince := time.Now().Add(-48 * time.Hour)
	serviceInfo.AddService(&nexmonyx.ServiceMonitoringInfo{
		Name:          "cron.service",
		State:         "active",
		SubState:      "running",
		LoadState:     "loaded",
		Description:   "Regular background program processing daemon",
		MainPID:       13969,
		MemoryCurrent: 679936,
		CPUUsageNSec:  1250000000,
		TasksCurrent:  1,
		ActiveSince:   &cronActiveSince,
	})

	// Add service metrics
	serviceInfo.AddMetrics(nexmonyx.CreateServiceMetrics(
		"ssh.service",
		0.1,      // CPU percent
		4308992,  // Memory RSS
		1,        // Process count
		1,        // Thread count
	))

	serviceInfo.AddMetrics(nexmonyx.CreateServiceMetrics(
		"cron.service",
		0.0,
		679936,
		1,
		1,
	))

	// Add service logs
	serviceInfo.AddLogEntry("ssh.service", nexmonyx.ServiceLogEntry{
		Timestamp: time.Now().Add(-5 * time.Minute),
		Level:     "info",
		Message:   "Accepted publickey for user from 192.168.1.100 port 52847 ssh2",
		Fields: map[string]string{
			"pid":  "163754",
			"unit": "ssh.service",
		},
	})

	serviceInfo.AddLogEntry("nginx.service", nexmonyx.ServiceLogEntry{
		Timestamp: time.Now().Add(-1 * time.Hour),
		Level:     "error",
		Message:   "nginx: [emerg] bind() to 0.0.0.0:80 failed (98: Address already in use)",
		Fields: map[string]string{
			"unit": "nginx.service",
		},
	})

	// Display service summary
	fmt.Printf("Total services monitored: %d\n", len(serviceInfo.Services))
	stateCounts := serviceInfo.CountServicesByState()
	for state, count := range stateCounts {
		fmt.Printf("  %s: %d\n", state, count)
	}

	failedServices := serviceInfo.GetFailedServices()
	if len(failedServices) > 0 {
		fmt.Printf("\nFailed services:\n")
		for _, service := range failedServices {
			fmt.Printf("  - %s: %s\n", service.Name, service.Description)
		}
	}

	// Submit as part of comprehensive metrics
	metricsRequest := &nexmonyx.ComprehensiveMetricsRequest{
		ServerUUID:  "your-server-uuid",
		CollectedAt: time.Now().UTC().Format(time.RFC3339),
		Services:    serviceInfo,
		SystemInfo: &nexmonyx.SystemInfo{
			Hostname:      "server-01",
			OS:            "Linux",
			OSVersion:     "Ubuntu 22.04",
			Architecture:  "x86_64",
			Uptime:        172800, // 2 days
			Processes:     150,
		},
	}

	err := client.Metrics.SubmitComprehensive(ctx, metricsRequest)
	if err != nil {
		log.Printf("Failed to submit metrics: %v", err)
	} else {
		fmt.Println("Successfully submitted service monitoring data")
	}
}

func monitorCriticalServices() {
	fmt.Println("\n=== Monitoring Critical Services ===")

	// Define critical services to monitor
	criticalServices := []string{
		"ssh.service",
		"systemd-resolved.service",
		"systemd-networkd.service",
		"cron.service",
	}

	// Create monitoring configuration
	config := nexmonyx.NewServiceMonitoringConfig()
	config.IncludeServices = criticalServices
	config.CollectMetrics = true
	config.CollectLogs = true

	// Simulate monitoring critical services
	serviceInfo := nexmonyx.NewServiceInfo()

	// Add some test services
	for _, serviceName := range criticalServices {
		if config.ShouldMonitorService(serviceName) {
			fmt.Printf("Monitoring: %s\n", serviceName)
			
			// In a real implementation, you would query systemd for actual data
			serviceInfo.AddService(&nexmonyx.ServiceMonitoringInfo{
				Name:      serviceName,
				State:     "active",
				SubState:  "running",
				LoadState: "loaded",
			})
		}
	}

	// Check service health
	for _, service := range serviceInfo.Services {
		health := nexmonyx.GetServiceHealth(service)
		fmt.Printf("  %s health: %d%%\n", service.Name, health)
	}
}

func analyzeServiceLogs() {
	fmt.Println("\n=== Analyzing Service Logs ===")

	serviceInfo := nexmonyx.NewServiceInfo()

	// Add sample logs
	serviceInfo.AddLogEntry("nginx.service", nexmonyx.CreateServiceLogEntry(
		"error",
		"upstream timed out (110: Connection timed out) while reading response header",
	))

	serviceInfo.AddLogEntry("mysql.service", nexmonyx.CreateServiceLogEntry(
		"warning",
		"Aborted connection 12345 to db: 'app_db' user: 'app_user' host: 'localhost'",
	))

	serviceInfo.AddLogEntry("ssh.service", nexmonyx.CreateServiceLogEntry(
		"info",
		"Server listening on 0.0.0.0 port 22",
	))

	serviceInfo.AddLogEntry("nginx.service", nexmonyx.CreateServiceLogEntry(
		"error",
		"open() \"/var/cache/nginx/proxy_temp/1/02/0000000021\" failed (13: Permission denied)",
	))

	// Get all error logs
	errorLogs := serviceInfo.GetErrorLogs()
	fmt.Printf("Found %d services with errors:\n", len(errorLogs))
	
	for serviceName, logs := range errorLogs {
		fmt.Printf("\n%s errors:\n", serviceName)
		for _, log := range logs {
			fmt.Printf("  [%s] %s\n", log.Timestamp.Format("15:04:05"), log.Message)
		}
	}
}

func trackServiceResources() {
	fmt.Println("\n=== Tracking Service Resource Usage ===")

	serviceInfo := nexmonyx.NewServiceInfo()

	// Add services with varying resource usage
	services := []struct {
		name   string
		memory uint64
		cpu    uint64
	}{
		{"mysql.service", 2147483648, 5000000000},      // 2GB memory, 5s CPU
		{"nginx.service", 134217728, 1000000000},       // 128MB memory, 1s CPU
		{"redis.service", 536870912, 2000000000},       // 512MB memory, 2s CPU
		{"elasticsearch.service", 4294967296, 10000000000}, // 4GB memory, 10s CPU
	}

	for _, svc := range services {
		activeSince := time.Now().Add(-72 * time.Hour)
		serviceInfo.AddService(&nexmonyx.ServiceMonitoringInfo{
			Name:          svc.name,
			State:         "active",
			SubState:      "running",
			MemoryCurrent: svc.memory,
			CPUUsageNSec:  svc.cpu,
			ActiveSince:   &activeSince,
		})
	}

	// Calculate totals
	totalMemory := serviceInfo.CalculateTotalMemoryUsage()
	totalCPU := serviceInfo.CalculateTotalCPUTime()

	fmt.Printf("Total memory usage: %.2f GB\n", float64(totalMemory)/(1024*1024*1024))
	fmt.Printf("Total CPU time: %s\n", totalCPU)

	// Find high memory services (> 1GB)
	highMemServices := serviceInfo.GetHighMemoryServices(1024 * 1024 * 1024)
	fmt.Printf("\nServices using > 1GB memory:\n")
	for _, service := range highMemServices {
		fmt.Printf("  %s: %.2f GB (uptime: %s)\n", 
			service.Name, 
			float64(service.MemoryCurrent)/(1024*1024*1024),
			nexmonyx.FormatServiceUptime(service.ActiveSince))
	}

	// Add metrics for trend analysis
	for _, service := range serviceInfo.Services {
		cpuPercent := float64(service.CPUUsageNSec) / float64(time.Since(*service.ActiveSince).Nanoseconds()) * 100
		serviceInfo.AddMetrics(&nexmonyx.ServiceMetrics{
			ServiceName:  service.Name,
			Timestamp:    time.Now(),
			CPUPercent:   cpuPercent,
			MemoryRSS:    service.MemoryCurrent,
			ProcessCount: 1,
			ThreadCount:  int(service.TasksCurrent),
		})
	}

	// Display latest metrics
	fmt.Printf("\nLatest metrics:\n")
	for _, service := range serviceInfo.Services {
		metrics := serviceInfo.GetServiceMetrics(service.Name)
		if metrics != nil {
			fmt.Printf("  %s: CPU %.2f%%, Memory %.2f GB\n",
				service.Name,
				metrics.CPUPercent,
				float64(metrics.MemoryRSS)/(1024*1024*1024))
		}
	}
}