package nexmonyx_test

import (
	"context"
	"fmt"
	"log"
	"time"

	nexmonyx "github.com/nexmonyx/go-sdk"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// RequestIDKey is the context key for request ID
	RequestIDKey contextKey = "request_id"
)

// Example_basicUsage demonstrates basic SDK usage with JWT authentication
func Example_basicUsage() {
	config := &nexmonyx.Config{
		BaseURL: "https://api.nexmonyx.com",
		Auth: nexmonyx.AuthConfig{
			Token: "your-jwt-token",
		},
		Debug: true,
	}

	client, err := nexmonyx.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// List users (example - replace with actual user ID when available)
	users, _, err := client.Users.List(ctx, &nexmonyx.ListOptions{Page: 1, Limit: 10})
	if err != nil {
		log.Fatal(err)
	}

	if len(users) > 0 {
		fmt.Printf("First user: %s (%s)\n", users[0].Email, users[0].FirstName+" "+users[0].LastName)
	}

	// List organizations
	orgs, _, err := client.Organizations.List(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	for _, org := range orgs {
		fmt.Printf("Organization: %s (UUID: %s)\n", org.Name, org.UUID)
	}
}

// Example_serverAgent demonstrates agent-style usage with server credentials
func Example_serverAgent() {
	config := &nexmonyx.Config{
		BaseURL: "https://api.nexmonyx.com",
		Auth: nexmonyx.AuthConfig{
			ServerUUID:   "your-server-uuid",
			ServerSecret: "your-server-secret",
		},
	}

	client, err := nexmonyx.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// List servers
	servers, _, err := client.Servers.List(ctx, &nexmonyx.ListOptions{Page: 1, Limit: 10})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d servers\n", len(servers))

	// Submit comprehensive metrics
	metrics := &nexmonyx.ComprehensiveMetricsRequest{
		ServerUUID:  "your-server-uuid",
		CollectedAt: time.Now().Format(time.RFC3339),
		SystemInfo: &nexmonyx.SystemInfo{
			Hostname:      "web-server-01",
			OS:            "Ubuntu",
			OSVersion:     "22.04 LTS",
			KernelVersion: "5.15.0-72-generic",
			Uptime:        3600,
		},
		CPU: &nexmonyx.CPUMetrics{
			UsagePercent:  45.2,
			LoadAverage1:  1.2,
			LoadAverage5:  1.5,
			LoadAverage15: 1.8,
			CoreCount:     4,
			ThreadCount:   8,
		},
		Memory: &nexmonyx.MemoryMetrics{
			TotalBytes:       8589934592, // 8GB
			UsedBytes:        3865470976, // ~3.6GB
			FreeBytes:        4724463616, // ~4.4GB
			AvailableBytes:   4724463616,
			UsagePercent:     45.1,
			SwapTotalBytes:   2147483648, // 2GB
			SwapUsedBytes:    0,
			SwapFreeBytes:    2147483648,
			SwapUsagePercent: 0,
		},
	}

	err = client.Metrics.SubmitComprehensive(ctx, metrics)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Metrics submitted successfully")
}

// Example_apiKeyAuth demonstrates API key authentication
func Example_apiKeyAuth() {
	config := &nexmonyx.Config{
		BaseURL: "https://api.nexmonyx.com",
		Auth: nexmonyx.AuthConfig{
			APIKey:    "your-api-key",
			APISecret: "your-api-secret",
		},
	}

	client, err := nexmonyx.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// List servers with pagination
	opts := &nexmonyx.ListOptions{
		Page:  1,
		Limit: 10,
		Sort:  "hostname",
		Order: "asc",
	}

	servers, meta, err := client.Servers.List(ctx, opts)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d servers (page %d of %d)\n", meta.TotalItems, meta.Page, meta.TotalPages)
	for _, server := range servers {
		fmt.Printf("Server: %s (%s) - %s\n", server.Hostname, server.ServerUUID, server.Status)
	}
}

// Example_organizationManagement demonstrates organization management
func Example_organizationManagement() {
	config := &nexmonyx.Config{
		BaseURL: "https://api.nexmonyx.com",
		Auth: nexmonyx.AuthConfig{
			Token: "your-jwt-token",
		},
	}

	client, err := nexmonyx.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Create organization
	orgReq := &nexmonyx.Organization{
		Name: "My Test Organization",
	}

	org, err := client.Organizations.Create(ctx, orgReq)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created organization: %s (UUID: %s)\n", org.Name, org.UUID)

	// Get organization users
	users, _, err := client.Organizations.GetUsers(ctx, org.UUID, &nexmonyx.ListOptions{Page: 1, Limit: 10})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Organization has %d users\n", len(users))
}

// Example_monitoringAndAlerts demonstrates monitoring and alerting
func Example_monitoringAndAlerts() {
	config := &nexmonyx.Config{
		BaseURL: "https://api.nexmonyx.com",
		Auth: nexmonyx.AuthConfig{
			Token: "your-jwt-token",
		},
	}

	client, err := nexmonyx.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Create probe
	probeReq := &nexmonyx.MonitoringProbe{
		Name:     "Website Health Check",
		Type:     "http",
		Target:   "https://example.com",
		Interval: 60, // seconds
		Timeout:  30, // seconds
		Config: map[string]interface{}{
			"method":           "GET",
			"expected_status":  200,
			"follow_redirects": true,
		},
		Tags: []string{"production", "website"},
	}

	probe, err := client.Monitoring.CreateProbe(ctx, probeReq)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created probe: %s (ID: %d)\n", probe.Name, probe.ID)

	// Create alert
	alertReq := &nexmonyx.Alert{
		Name:        "High CPU Usage",
		Description: "Alert when CPU usage exceeds 80%",
		MetricName:  "cpu",
		Condition:   "greater_than",
		Threshold:   80.0,
		Duration:    300, // 5 minutes
		Severity:    "warning",
	}

	alert, err := client.Alerts.Create(ctx, alertReq)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created alert: %s (ID: %d)\n", alert.Name, alert.ID)
}

// Example_metricsQuery demonstrates querying metrics
func Example_metricsQuery() {
	config := &nexmonyx.Config{
		BaseURL: "https://api.nexmonyx.com",
		Auth: nexmonyx.AuthConfig{
			Token: "your-jwt-token",
		},
	}

	client, err := nexmonyx.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	serverUUID := "your-server-uuid"

	// Get server metrics for the last hour
	_ = &nexmonyx.TimeRange{
		Start: time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
		End:   time.Now().Format(time.RFC3339),
	}

	metrics, _, err := client.Servers.GetMetrics(ctx, serverUUID, &nexmonyx.ListOptions{Page: 1, Limit: 10})
	if err != nil {
		log.Fatal(err)
	}

	if len(metrics) > 0 {
		fmt.Printf("Retrieved %d metrics for server\n", len(metrics))
	}

	// Query available metrics
	query := &nexmonyx.MetricsQuery{
		ServerUUIDs: []string{serverUUID},
		MetricNames: []string{"cpu"},
		Limit:       10,
		StartTime:   time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
		EndTime:     time.Now().Format(time.RFC3339),
	}
	allMetrics, err := client.Metrics.Query(ctx, query)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d metrics\n", len(allMetrics))
}

// Example_errorHandling demonstrates error handling
func Example_errorHandling() {
	config := &nexmonyx.Config{
		BaseURL: "https://api.nexmonyx.com",
		Auth: nexmonyx.AuthConfig{
			Token: "invalid-token",
		},
	}

	client, err := nexmonyx.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// This will return an error
	_, _, err = client.Users.List(ctx, &nexmonyx.ListOptions{Page: 1, Limit: 1})
	if err != nil {
		switch e := err.(type) {
		case *nexmonyx.APIError:
			fmt.Printf("API Error: %s - %s\n", e.ErrorCode, e.Message)
		case *nexmonyx.UnauthorizedError:
			fmt.Printf("Unauthorized: %s\n", e.Message)
		case *nexmonyx.NotFoundError:
			fmt.Printf("Not found: %s %s\n", e.Resource, e.ID)
		case *nexmonyx.RateLimitError:
			fmt.Printf("Rate limited: %s\n", e.Message)
		default:
			fmt.Printf("Unknown error: %v\n", err)
		}
	}
}

// Example_customHeaders demonstrates using custom headers
func Example_customHeaders() {
	config := &nexmonyx.Config{
		BaseURL: "https://api.nexmonyx.com",
		Auth: nexmonyx.AuthConfig{
			Token: "your-jwt-token",
		},
		Headers: map[string]string{
			"X-Custom-Header":  "custom-value",
			"X-Client-Version": "1.0.0",
		},
	}

	client, err := nexmonyx.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Add request-specific context
	ctx = context.WithValue(ctx, RequestIDKey, "req-12345")

	users, _, err := client.Users.List(ctx, &nexmonyx.ListOptions{Page: 1, Limit: 1})
	if err != nil {
		log.Fatal(err)
	}

	if len(users) > 0 {
		fmt.Printf("User: %s\n", users[0].Email)
	}
}
