package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	nexmonyx "github.com/nexmonyx/go-sdk"
)

func main() {
	// Command line flags
	var (
		apiURL       = flag.String("api", "https://api.nexmonyx.com", "API endpoint URL")
		serverUUID   = flag.String("uuid", "", "Server UUID")
		serverSecret = flag.String("secret", "", "Server Secret")
		testType     = flag.String("test", "heartbeat", "Test type: heartbeat, update-details, update-info, get-heartbeat")
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
		fmt.Println("Usage: debug_heartbeat -uuid <SERVER_UUID> -secret <SERVER_SECRET> [-api <API_URL>] [-test <TYPE>]")
		fmt.Println("\nAlternatively, set environment variables:")
		fmt.Println("  NEXMONYX_SERVER_UUID")
		fmt.Println("  NEXMONYX_SERVER_SECRET")
		fmt.Println("\nTest types:")
		fmt.Println("  heartbeat       - Test basic heartbeat")
		fmt.Println("  update-details  - Test server details update")
		fmt.Println("  update-info     - Test server info update")
		fmt.Println("  get-heartbeat   - Test heartbeat retrieval")
		os.Exit(1)
	}

	fmt.Println("========================================")
	fmt.Println("Nexmonyx SDK Debug Test")
	fmt.Println("========================================")
	fmt.Printf("API URL: %s\n", *apiURL)
	fmt.Printf("Server UUID: %s\n", *serverUUID)
	fmt.Printf("Test Type: %s\n", *testType)
	fmt.Println("Debug Mode: ENABLED")
	fmt.Println("========================================")

	// Create client with debug mode enabled
	client, err := nexmonyx.NewClient(&nexmonyx.Config{
		BaseURL: *apiURL,
		Auth: nexmonyx.AuthConfig{
			ServerUUID:   *serverUUID,
			ServerSecret: *serverSecret,
		},
		Debug: true, // Enable debug logging
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	switch *testType {
	case "heartbeat":
		testHeartbeat(ctx, client)
	case "update-details":
		testUpdateDetails(ctx, client, *serverUUID)
	case "update-info":
		testUpdateInfo(ctx, client, *serverUUID)
	default:
		log.Fatalf("Unknown test type: %s", *testType)
	}
}

func testHeartbeat(ctx context.Context, client *nexmonyx.Client) {
	fmt.Println("\n=== Testing Heartbeat ===")

	// Test basic heartbeat
	err := client.Servers.Heartbeat(ctx)
	if err != nil {
		fmt.Printf("\n❌ Heartbeat failed: %v\n", err)
	} else {
		fmt.Printf("\n✅ Heartbeat successful!\n")
	}

	// Test heartbeat with version
	fmt.Println("\n=== Testing Heartbeat with Version ===")
	err = client.Servers.HeartbeatWithVersion(ctx, "v1.0.0")
	if err != nil {
		fmt.Printf("\n❌ Heartbeat with version failed: %v\n", err)
	} else {
		fmt.Printf("\n✅ Heartbeat with version successful!\n")
	}
}

func testUpdateDetails(ctx context.Context, client *nexmonyx.Client, serverUUID string) {
	fmt.Println("\n=== Testing Update Details ===")

	details := &nexmonyx.ServerDetailsUpdateRequest{
		Hostname:     "debug-test-server",
		OS:           "Ubuntu 22.04 LTS",
		OSVersion:    "5.15.0-88-generic",
		OSArch:       "x86_64",
		CPUModel:     "Intel(R) Core(TM) i7-9700K CPU @ 3.60GHz",
		CPUCores:     8,
		MemoryTotal:  16384 * 1024 * 1024,      // Convert MB to bytes
		StorageTotal: 500 * 1024 * 1024 * 1024, // Convert GB to bytes
	}

	server, err := client.Servers.UpdateDetails(ctx, serverUUID, details)
	if err != nil {
		fmt.Printf("\n❌ Update details failed: %v\n", err)
	} else {
		fmt.Printf("\n✅ Update details successful!\n")
		if server != nil {
			fmt.Printf("Server ID: %d\n", server.ID)
			fmt.Printf("Server UUID: %s\n", server.ServerUUID)
			fmt.Printf("Server Hostname: %s\n", server.Hostname)
		}
	}
}

func testUpdateInfo(ctx context.Context, client *nexmonyx.Client, serverUUID string) {
	fmt.Println("\n=== Testing Update Info ===")

	info := &nexmonyx.ServerDetailsUpdateRequest{
		Hostname:     "debug-test-server-info",
		OS:           "Ubuntu 22.04 LTS",
		OSVersion:    "5.15.0-88-generic",
		OSArch:       "x86_64",
		CPUModel:     "Intel(R) Core(TM) i7-9700K CPU @ 3.60GHz",
		CPUCores:     8,
		MemoryTotal:  16384 * 1024 * 1024,      // Convert MB to bytes
		StorageTotal: 500 * 1024 * 1024 * 1024, // Convert GB to bytes
	}

	server, err := client.Servers.UpdateInfo(ctx, serverUUID, info)
	if err != nil {
		fmt.Printf("\n❌ Update info failed: %v\n", err)
	} else {
		fmt.Printf("\n✅ Update info successful!\n")
		if server != nil {
			fmt.Printf("Server ID: %d\n", server.ID)
			fmt.Printf("Server UUID: %s\n", server.ServerUUID)
			fmt.Printf("Server Hostname: %s\n", server.Hostname)
		}
	}
}

