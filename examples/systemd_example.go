//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	nexmonyx "github.com/nexmonyx/go-sdk"
)

// This example demonstrates how to use the Systemd service in the Nexmonyx SDK
func main() {
	// Example 1: Agent submitting systemd service data
	agentExample()

	// Example 2: User querying systemd service data
	userExample()

	// Example 3: Advanced querying and filtering
	advancedExample()
}

func agentExample() {
	fmt.Println("\n=== Agent Example: Submitting Systemd Service Data ===")

	// Create client with server credentials (for agents)
	client, err := nexmonyx.NewClient(&nexmonyx.Config{
		BaseURL: "https://api.nexmonyx.com",
		Auth: nexmonyx.AuthConfig{
			ServerUUID:   os.Getenv("NEXMONYX_SERVER_UUID"),
			ServerSecret: os.Getenv("NEXMONYX_SERVER_SECRET"),
		},
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Prepare systemd service data
	services := []nexmonyx.SystemdServiceInfo{
		{
			Name:                 "nginx.service",
			UnitType:             "service",
			Description:          "A high performance web server and a reverse proxy server",
			LoadState:            "loaded",
			ActiveState:          "active",
			SubState:             "running",
			UnitState:            "enabled",
			MainPID:              12345,
			Type:                 "forking",
			User:                 "www-data",
			Group:                "www-data",
			WorkingDir:           "/",
			ExecStart:            []string{"/usr/sbin/nginx -g 'daemon on; master_process on;'"},
			ExecReload:           []string{"/usr/sbin/nginx -g 'daemon on; master_process on;' -s reload"},
			ExecStop:             []string{"/usr/sbin/nginx -s quit"},
			MemoryCurrent:        104857600, // 100MB
			CPUUsageNSec:         5000000000,
			CPUUsagePercent:      2.5,
			TasksCurrent:         5,
			RestartCount:         0,
			StartupDuration:      0.125,
			HealthScore:          95,
			PrivateTmp:           true,
			ProtectSystem:        "full",
			ProtectHome:          "true",
			NoNewPrivileges:      true,
			DetectionMethod:      "systemctl",
			ActiveEnterTimestamp: time.Now().Add(-24 * time.Hour),
		},
		{
			Name:        "postgresql.service",
			UnitType:    "service",
			Description: "PostgreSQL RDBMS",
			LoadState:   "loaded",
			ActiveState: "active",
			SubState:    "running",
			UnitState:   "enabled",
			MainPID:     67890,
			Type:        "notify",
			User:        "postgres",
			Group:       "postgres",
			WorkingDir:  "/var/lib/postgresql",
			ExecStart:   []string{"/usr/lib/postgresql/14/bin/postgres -D /var/lib/postgresql/14/main"},
			Environment: []string{
				"PGDATA=/var/lib/postgresql/14/main",
				"LANG=en_US.UTF-8",
			},
			Wants:                []string{"network.target"},
			After:                []string{"network.target"},
			MemoryCurrent:        536870912, // 512MB
			CPUUsageNSec:         20000000000,
			CPUUsagePercent:      5.2,
			TasksCurrent:         15,
			RestartCount:         0,
			StartupDuration:      2.456,
			HealthScore:          90,
			PrivateNetwork:       false,
			PrivateTmp:           true,
			ProtectSystem:        "full",
			ProtectHome:          "true",
			NoNewPrivileges:      true,
			DetectionMethod:      "systemctl",
			ActiveEnterTimestamp: time.Now().Add(-7 * 24 * time.Hour),
		},
		{
			Name:                "redis.service",
			UnitType:            "service",
			Description:         "Advanced key-value store",
			LoadState:           "loaded",
			ActiveState:         "failed",
			SubState:            "failed",
			UnitState:           "enabled",
			Type:                "notify",
			ExitCode:            1,
			ExitStatus:          "1/FAILURE",
			Result:              "exit-code",
			StatusText:          "Fatal error, can't open config file '/etc/redis/redis.conf'",
			HealthScore:         0,
			RestartCount:        3,
			DetectionMethod:     "systemctl",
			ActiveExitTimestamp: time.Now().Add(-10 * time.Minute),
		},
	}

	// Prepare system statistics
	systemStats := &nexmonyx.SystemdSystemStats{
		TotalUnits:         250,
		ServiceUnits:       120,
		SocketUnits:        45,
		TargetUnits:        30,
		TimerUnits:         15,
		MountUnits:         20,
		DeviceUnits:        10,
		ScopeUnits:         5,
		SliceUnits:         5,
		ActiveUnits:        230,
		InactiveUnits:      17,
		FailedUnits:        3,
		EnabledUnits:       180,
		DisabledUnits:      70,
		MaskedUnits:        0,
		SystemStartupTime:  15.789,
		LastBootTime:       time.Now().Add(-30 * 24 * time.Hour),
		SystemManagerPID:   1,
		SystemState:        "degraded",
		TotalMemoryUsage:   2147483648, // 2GB
		TotalCPUUsage:      25.5,
		TotalTaskCount:     450,
		OverallHealthScore: 85,
		CriticalServices:   []string{"redis.service", "mysql.service"},
		SystemdIssues: []string{
			"3 units failed",
			"System in degraded state",
		},
		RecentFailures: 3,
	}

	// Create request
	request := &nexmonyx.SystemdServiceRequest{
		ServerUUID:  os.Getenv("NEXMONYX_SERVER_UUID"),
		CollectedAt: time.Now().Format(time.RFC3339),
		Services:    services,
		SystemStats: systemStats,
	}

	// Submit the data
	ctx := context.Background()
	err = client.Systemd.Submit(ctx, request)
	if err != nil {
		log.Fatalf("Failed to submit systemd data: %v", err)
	}

	fmt.Printf("Successfully submitted systemd data\n")
}

func userExample() {
	fmt.Println("\n=== User Example: Querying Systemd Service Data ===")

	// Create client with user token
	client, err := nexmonyx.NewClient(&nexmonyx.Config{
		BaseURL: "https://api.nexmonyx.com",
		Auth: nexmonyx.AuthConfig{
			Token: os.Getenv("NEXMONYX_AUTH_TOKEN"),
		},
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	serverUUID := "550e8400-e29b-41d4-a716-446655440000"

	// Example 1: Get latest systemd services
	fmt.Println("\n1. Getting latest systemd services...")
	services, err := client.Systemd.Get(ctx, serverUUID)
	if err != nil {
		log.Printf("Failed to get latest services: %v", err)
	} else {
		fmt.Printf("Found %d services\n", len(services))
		for _, svc := range services {
			fmt.Printf("  - %s: %s/%s (Health: %d%%)\n",
				svc.Name, svc.ActiveState, svc.SubState, svc.HealthScore)
		}
	}

	// Example 2: Get services with list options
	fmt.Println("\n2. Getting services with filters...")
	opts := &nexmonyx.ListOptions{
		StartDate: time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
		EndDate:   time.Now().Format(time.RFC3339),
		Limit:     100,
	}
	services, meta, err := client.Systemd.List(ctx, opts)
	if err != nil {
		log.Printf("Failed to get services: %v", err)
	} else {
		fmt.Printf("Found %d services in the last 24 hours (page %d of %d)\n", 
			len(services), meta.Page, meta.TotalPages)
	}

	// Example 3: Get specific service
	fmt.Println("\n3. Getting specific service...")
	service, err := client.Systemd.GetServiceByName(ctx, serverUUID, "nginx.service")
	if err != nil {
		log.Printf("Failed to get nginx.service: %v", err)
	} else {
		fmt.Printf("Service: %s\n", service.Name)
		fmt.Printf("  Status: %s/%s\n", service.ActiveState, service.SubState)
		fmt.Printf("  Type: %s\n", service.Type)
		fmt.Printf("  PID: %d\n", service.MainPID)
		fmt.Printf("  Memory: %.2f MB\n", float64(service.MemoryCurrent)/1024/1024)
		fmt.Printf("  CPU: %.2f%%\n", service.CPUUsagePercent)
		fmt.Printf("  Healthy: %v\n", service.IsHealthy())
	}

	// Example 4: Get system statistics
	fmt.Println("\n4. Getting system statistics...")
	stats, err := client.Systemd.GetSystemStats(ctx, serverUUID)
	if err != nil {
		log.Printf("Failed to get system stats: %v", err)
	} else {
		fmt.Printf("System Statistics:\n")
		fmt.Printf("  Total Units: %d\n", stats.TotalUnits)
		fmt.Printf("  Service Units: %d\n", stats.ServiceUnits)
		fmt.Printf("  Active Units: %d\n", stats.ActiveUnits)
		fmt.Printf("  Failed Units: %d\n", stats.FailedUnits)
		fmt.Printf("  System State: %s\n", stats.SystemState)
		fmt.Printf("  System Healthy: %v\n", stats.IsHealthy())
		fmt.Printf("  System Degraded: %v\n", stats.IsDegraded())
	}

	// Example 5: Get service history
	fmt.Println("\n5. Getting service history...")
	opts = &nexmonyx.ListOptions{
		Filters: map[string]string{
			"server_uuid":  serverUUID,
			"service_name": "nginx.service",
		},
		StartDate: time.Now().Add(-7 * 24 * time.Hour).Format(time.RFC3339),
		EndDate:   time.Now().Format(time.RFC3339),
		Limit:     10,
	}
	services, pagination, err := client.Systemd.List(ctx, opts)
	if err != nil {
		log.Printf("Failed to get service history: %v", err)
	} else {
		fmt.Printf("Service history for nginx.service:\n")
		for _, svc := range services {
			fmt.Printf("  %s: %s/%s\n",
				"recent",
				svc.ActiveState,
				svc.SubState)
		}
		if pagination != nil {
			fmt.Printf("\nPagination: Page %d of %d (Total: %d)\n",
				pagination.Page, pagination.TotalPages, pagination.TotalItems)
		}
	}
}

func advancedExample() {
	fmt.Println("\n=== Advanced Example: Complex Queries and Analysis ===")

	// Create client
	client, err := nexmonyx.NewClient(&nexmonyx.Config{
		BaseURL: "https://api.nexmonyx.com",
		Auth: nexmonyx.AuthConfig{
			Token: os.Getenv("NEXMONYX_AUTH_TOKEN"),
		},
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Example 1: Query all failed services
	fmt.Println("\n1. Querying all failed services...")
	opts := &nexmonyx.ListOptions{
		Filters: map[string]string{
			"active_state": "failed",
		},
		Limit: 50,
	}
	failedServices, _, err := client.Systemd.List(ctx, opts)
	if err != nil {
		log.Printf("Failed to query services: %v", err)
	} else {
		fmt.Printf("Found %d failed services:\n", len(failedServices))
		for _, svc := range failedServices {
			fmt.Printf("  - %s: %s (Exit Code: %d)\n",
				svc.Name, svc.StatusText, svc.ExitCode)
		}
	}

	// Example 2: Query timer units
	fmt.Println("\n2. Querying timer units...")
	opts = &nexmonyx.ListOptions{
		Filters: map[string]string{
			"unit_type": "timer",
		},
		Limit: 20,
	}
	timers, _, err := client.Systemd.List(ctx, opts)
	if err != nil {
		log.Printf("Failed to query timers: %v", err)
	} else {
		fmt.Printf("Found %d timer units:\n", len(timers))
		for _, timer := range timers {
			if timer.NextElapseTime != nil {
				fmt.Printf("  - %s: Next run at %s\n",
					timer.Name, timer.NextElapseTime.Format("2006-01-02 15:04:05"))
			} else {
				fmt.Printf("  - %s: No next run scheduled\n", timer.Name)
			}
		}
	}

	// Example 3: Analyze service health across servers
	fmt.Println("\n3. Analyzing service health across multiple servers...")
	serverUUIDs := []string{
		"550e8400-e29b-41d4-a716-446655440001",
		"550e8400-e29b-41d4-a716-446655440002",
		"550e8400-e29b-41d4-a716-446655440003",
	}

	for _, uuid := range serverUUIDs {
		services, err := client.Systemd.Get(ctx, uuid)
		if err != nil {
			log.Printf("Failed to get services for %s: %v", uuid, err)
			continue
		}

		healthyCount := 0
		failedCount := 0
		for _, svc := range services {
			if svc.IsHealthy() {
				healthyCount++
			}
			if svc.IsFailed() {
				failedCount++
			}
		}

		fmt.Printf("\nServer %s:\n", uuid)
		fmt.Printf("  Total Services: %d\n", len(services))
		fmt.Printf("  Healthy: %d\n", healthyCount)
		fmt.Printf("  Failed: %d\n", failedCount)
		
		// Get stats separately
		stats, err := client.Systemd.GetSystemStats(ctx, uuid)
		if err == nil && stats != nil {
			fmt.Printf("  System State: %s\n", stats.SystemState)
			fmt.Printf("  Overall Health: %d%%\n", stats.OverallHealthScore)
		}
	}

	// Example 4: Monitor critical services
	fmt.Println("\n4. Monitoring critical services...")
	criticalServices := []string{"nginx.service", "postgresql.service", "redis.service"}

	for _, serviceName := range criticalServices {
		fmt.Printf("\nChecking %s across all servers...\n", serviceName)

		opts := &nexmonyx.ListOptions{
			Filters: map[string]string{
				"service_name": serviceName,
			},
			StartDate: time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			EndDate:   time.Now().Format(time.RFC3339),
			Limit:     100,
		}

		history, _, err := client.Systemd.List(ctx, opts)
		if err != nil {
			log.Printf("Failed to get history for %s: %v", serviceName, err)
			continue
		}

		// Analyze state changes
		stateChanges := 0
		var lastState string
		for _, svc := range history {
			if lastState != "" && lastState != svc.ActiveState {
				stateChanges++
			}
			lastState = svc.ActiveState
		}

		fmt.Printf("  State changes in last hour: %d\n", stateChanges)
		if len(history) > 0 {
			current := history[0]
			fmt.Printf("  Current state: %s/%s\n", current.ActiveState, current.SubState)
			fmt.Printf("  Restart count: %d\n", current.RestartCount)
		}
	}

	// Example 5: Working with additional info
	fmt.Println("\n5. Working with additional service information...")
	service, err := client.Systemd.GetServiceByName(ctx, serverUUIDs[0], "custom-app.service")
	if err == nil && service != nil && service.AdditionalInfo != nil {
		fmt.Printf("Custom app service additional info:\n")

		// Check for custom fields
		if val, ok := service.GetAdditionalInfo("custom_metric"); ok {
			fmt.Printf("  Custom Metric: %v\n", val)
		}

		if val, ok := service.GetAdditionalInfo("deployment_version"); ok {
			fmt.Printf("  Deployment Version: %v\n", val)
		}

		// Print all additional info
		for key, value := range service.AdditionalInfo {
			fmt.Printf("  %s: %v\n", key, value)
		}
	} else if err != nil {
		log.Printf("Failed to get custom-app.service: %v", err)
	}
}
