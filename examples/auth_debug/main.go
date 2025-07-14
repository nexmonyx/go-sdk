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
		debug        = flag.Bool("debug", false, "Enable debug mode")
		testAuth     = flag.Bool("test-auth", false, "Run authentication header tests")
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
		fmt.Println("Usage: auth_debug -uuid <SERVER_UUID> -secret <SERVER_SECRET> [-api <API_URL>] [-debug] [-test-auth]")
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

	// Run authentication header tests if requested
	if *testAuth {
		fmt.Println("Running authentication header tests...")
		if err := client.DebugAuthHeaders(ctx); err != nil {
			log.Fatalf("Authentication test failed: %v", err)
		}
		return
	}

	// Test heartbeat endpoint
	fmt.Println("Testing heartbeat endpoint with SDK...")
	fmt.Printf("Server UUID: %s\n", *serverUUID)
	fmt.Printf("API URL: %s\n", *apiURL)
	fmt.Println()

	// Enable debug mode to see headers
	if !*debug {
		fmt.Println("Tip: Use -debug flag to see request headers")
		fmt.Println()
	}

	// Attempt to send heartbeat
	err = client.Servers.Heartbeat(ctx)
	if err != nil {
		fmt.Printf("❌ Heartbeat failed: %v\n", err)

		// Provide diagnostic information
		fmt.Println("\nDiagnostic Information:")
		fmt.Println("- Ensure the server UUID and secret are correct")
		fmt.Println("- Verify the API endpoint is accessible")
		fmt.Println("- Run with -test-auth flag to test different header formats")
		fmt.Println("- Run with -debug flag to see request details")

		os.Exit(1)
	}

	fmt.Println("✅ Heartbeat successful!")
	fmt.Println("\nAuthentication is working correctly.")
}
