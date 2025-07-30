package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nexmonyx/go-sdk/v2"
)

// BasicMonitoringAgent demonstrates the basic functionality of a monitoring agent
// using the Nexmonyx Go SDK with MON_ key authentication
func main() {
	// Get monitoring key from environment
	monitoringKey := os.Getenv("NEXMONYX_MONITORING_KEY")
	if monitoringKey == "" {
		log.Fatal("NEXMONYX_MONITORING_KEY environment variable is required")
	}

	// Get API endpoint (defaults to production if not set)
	apiEndpoint := os.Getenv("NEXMONYX_API_ENDPOINT")
	if apiEndpoint == "" {
		apiEndpoint = "https://api.nexmonyx.com"
	}

	// Get region from environment (defaults to us-east-1)
	region := os.Getenv("NEXMONYX_REGION")
	if region == "" {
		region = "us-east-1"
	}

	fmt.Printf("Starting monitoring agent for region: %s\n", region)

	// Create monitoring agent client
	client, err := nexmonyx.NewMonitoringAgentClient(&nexmonyx.Config{
		BaseURL: apiEndpoint,
		Auth: nexmonyx.AuthConfig{
			MonitoringKey: monitoringKey,
		},
		Debug: os.Getenv("DEBUG") == "true",
	})
	if err != nil {
		log.Fatalf("Failed to create monitoring agent client: %v", err)
	}

	ctx := context.Background()

	// Test authentication with health check
	fmt.Println("Testing authentication...")
	if err := client.HealthCheck(ctx); err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}
	fmt.Println("✓ Authentication successful")

	// Get assigned probes for our region
	fmt.Printf("Fetching assigned probes for region: %s\n", region)
	probes, err := client.Monitoring.GetAssignedProbes(ctx, region)
	if err != nil {
		log.Fatalf("Failed to get assigned probes: %v", err)
	}

	fmt.Printf("✓ Found %d assigned probes\n", len(probes))
	for i, probe := range probes {
		fmt.Printf("  %d. %s (%s) -> %s (interval: %ds)\n", 
			i+1, probe.Name, probe.Type, probe.Target, probe.Interval)
	}

	// Create node info for heartbeat
	nodeInfo := nexmonyx.NodeInfo{
		AgentID:      "monitoring-agent-example",
		AgentVersion: "1.0.0",
		Region:       region,
		Hostname:     getHostname(),
		IPAddress:    getLocalIP(),
		Status:       "healthy",
		Uptime:       time.Hour * 2, // Example: agent has been running for 2 hours
		LastSeen:     time.Now(),
		ProbesAssigned: len(probes),
		SupportedTypes: []string{"http", "https", "tcp", "icmp"},
		MaxConcurrency: 10,
		Environment:    "production",
	}

	// Send heartbeat
	fmt.Println("Sending heartbeat...")
	if err := client.Monitoring.Heartbeat(ctx, nodeInfo); err != nil {
		log.Fatalf("Failed to send heartbeat: %v", err)
	}
	fmt.Println("✓ Heartbeat sent successfully")

	// Simulate executing probes and submitting results
	if len(probes) > 0 {
		fmt.Println("Simulating probe execution...")
		results := simulateProbeExecution(probes)
		
		if err := client.Monitoring.SubmitResults(ctx, results); err != nil {
			log.Fatalf("Failed to submit probe results: %v", err)
		}
		
		fmt.Printf("✓ Submitted %d probe results\n", len(results))
	}

	fmt.Println("Monitoring agent example completed successfully!")
}

// simulateProbeExecution simulates executing probes and returns mock results
func simulateProbeExecution(probes []*nexmonyx.ProbeAssignment) []nexmonyx.ProbeExecutionResult {
	var results []nexmonyx.ProbeExecutionResult
	
	for _, probe := range probes {
		// Simulate successful probe execution
		result := nexmonyx.ProbeExecutionResult{
			ProbeID:       probe.ProbeID,
			ProbeUUID:     probe.ProbeUUID,
			ExecutedAt:    time.Now(),
			Region:        probe.Region,
			Status:        "success",
			ResponseTime:  150 + (len(probe.Target) % 100), // Mock response time
			StatusCode:    200,
			DNSTime:       10,
			ConnectTime:   30,
			TLSTime:       20,
			FirstByteTime: 80,
			TotalTime:     150,
			ResponseSize:  1024,
		}
		
		// Occasionally simulate failures for demonstration
		if probe.ProbeID%7 == 0 { // Every 7th probe "fails"
			result.Status = "failed"
			result.StatusCode = 500
			result.Error = "Connection timeout"
			result.ResponseTime = probe.Timeout * 1000 // Convert to milliseconds
		}
		
		results = append(results, result)
	}
	
	return results
}

// getHostname returns the system hostname
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

// getLocalIP returns a mock local IP address
func getLocalIP() string {
	// In a real implementation, you'd get the actual local IP
	return "10.0.1.100"
}