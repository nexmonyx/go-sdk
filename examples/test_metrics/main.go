package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	nexmonyx "github.com/nexmonyx/go-sdk"
)

func main() {
	// Command line flags
	var (
		apiURL       = flag.String("api", "https://api.nexmonyx.com", "API endpoint URL")
		serverUUID   = flag.String("uuid", "", "Server UUID")
		serverSecret = flag.String("secret", "", "Server Secret")
		debug        = flag.Bool("debug", false, "Enable debug mode")
	)
	flag.Parse()

	// Check for environment variables if flags not provided
	if *serverUUID == "" {
		*serverUUID = os.Getenv("NEXMONYX_SERVER_UUID")
	}
	if *serverSecret == "" {
		*serverSecret = os.Getenv("NEXMONYX_SERVER_SECRET")
	}

	// Validate required parameters
	if *serverUUID == "" || *serverSecret == "" {
		fmt.Println("Usage: test_metrics -uuid <SERVER_UUID> -secret <SERVER_SECRET> [-api <API_URL>] [-debug]")
		fmt.Println("\nAlternatively, set environment variables:")
		fmt.Println("  NEXMONYX_SERVER_UUID")
		fmt.Println("  NEXMONYX_SERVER_SECRET")
		os.Exit(1)
	}

	// Create client with server credentials
	client, err := nexmonyx.NewClient(&nexmonyx.Config{
		BaseURL: *apiURL,
		Auth: nexmonyx.AuthConfig{
			ServerUUID:   *serverUUID,
			ServerSecret: *serverSecret,
		},
		Debug: *debug,
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test metrics submission
	fmt.Println("Testing metrics submission with fixed headers...")
	fmt.Printf("Server UUID: %s\n", *serverUUID)
	fmt.Printf("API URL: %s\n", *apiURL)
	fmt.Printf("SDK Version: %s\n", nexmonyx.Version)
	fmt.Println()

	metrics := &nexmonyx.ComprehensiveMetricsRequest{
		ServerUUID:  *serverUUID,
		CollectedAt: time.Now().UTC().Format(time.RFC3339),
		CPU: &nexmonyx.CPUMetrics{
			UsagePercent: 45.5,
		},
		Memory: &nexmonyx.MemoryMetrics{
			TotalBytes:     8589934592, // 8GB
			AvailableBytes: 4294967296, // 4GB
			UsedBytes:      4294967296, // 4GB
			FreeBytes:      4294967296, // 4GB
		},
	}

	err = client.Metrics.SubmitComprehensive(ctx, metrics)
	if err != nil {
		fmt.Printf("❌ Metrics submission failed: %v\n", err)
		fmt.Println("\nThis indicates the authentication headers issue is NOT fixed.")
		os.Exit(1)
	}

	fmt.Println("✅ Metrics submission successful!")
	fmt.Println("\nThe SDK is now correctly using 'Server-UUID' and 'Server-Secret' headers.")
	fmt.Println("Authentication is working correctly for the metrics endpoint.")
}
