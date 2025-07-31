package main

import (
	"context"
	"fmt"
	"log"
	"os"

	nexmonyx "github.com/nexmonyx/go-sdk/v2"
)

func main() {
	// Example of using the unified API key system

	// Create a client with admin authentication (using JWT token)
	adminClient, err := nexmonyx.NewClient(&nexmonyx.Config{
		BaseURL: "https://api.nexmonyx.com",
		Auth: nexmonyx.AuthConfig{
			Token: os.Getenv("NEXMONYX_ADMIN_TOKEN"),
		},
		Debug: true,
	})
	if err != nil {
		log.Fatalf("Failed to create admin client: %v", err)
	}

	ctx := context.Background()

	// Example 1: Create a user API key
	fmt.Println("=== Creating User API Key ===")
	userKeyReq := nexmonyx.NewUserAPIKey(
		"My User API Key",
		"API key for user operations",
		[]string{nexmonyx.CapabilityServersRead, nexmonyx.CapabilityMetricsRead},
	)

	userKeyResp, err := adminClient.APIKeys.CreateUnified(ctx, userKeyReq)
	if err != nil {
		log.Printf("Failed to create user API key: %v", err)
	} else {
		fmt.Printf("Created user API key: %s\n", userKeyResp.KeyID)
		fmt.Printf("Key value (save this): %s\n", userKeyResp.KeyValue)
		fmt.Printf("Secret (save this): %s\n", userKeyResp.Secret)
	}

	// Example 2: Create a monitoring agent key
	fmt.Println("\n=== Creating Monitoring Agent Key ===")
	agentKeyReq := nexmonyx.NewMonitoringAgentKey(
		"Private Monitoring Agent",
		"Agent for internal monitoring",
		"production",
		"private",
		"us-east-1",
		[]string{"public", "private"},
	)

	agentKeyResp, err := adminClient.APIKeys.CreateUnified(ctx, agentKeyReq)
	if err != nil {
		log.Printf("Failed to create monitoring agent key: %v", err)
	} else {
		fmt.Printf("Created monitoring agent key: %s\n", agentKeyResp.KeyID)
		fmt.Printf("Full token (save this): %s\n", agentKeyResp.FullToken)
	}

	// Example 3: Create a registration key
	fmt.Println("\n=== Creating Registration Key ===")
	regKeyReq := nexmonyx.NewRegistrationKey(
		"Server Registration Key",
		"For registering new servers",
		1, // Organization ID
	)

	regKeyResp, err := adminClient.APIKeys.AdminCreateUnified(ctx, regKeyReq)
	if err != nil {
		log.Printf("Failed to create registration key: %v", err)
	} else {
		fmt.Printf("Created registration key: %s\n", regKeyResp.KeyID)
		fmt.Printf("Registration token: %s\n", regKeyResp.FullToken)
	}

	// Example 4: Using the unified API key for authentication
	if userKeyResp != nil {
		fmt.Println("\n=== Using User API Key ===")
		
		// Create a client with the user API key
		userClient := adminClient.WithUnifiedAPIKeyAndSecret(userKeyResp.KeyValue, userKeyResp.Secret)
		
		// List servers using the user key
		servers, meta, err := userClient.Servers.List(ctx, &nexmonyx.ListOptions{
			Page:  1,
			Limit: 10,
		})
		if err != nil {
			log.Printf("Failed to list servers with user key: %v", err)
		} else {
			fmt.Printf("Listed %d servers (total %d)\n", len(servers), meta.TotalItems)
		}
	}

	// Example 5: Using monitoring agent key
	if agentKeyResp != nil {
		fmt.Println("\n=== Using Monitoring Agent Key ===")
		
		// Create a client with the monitoring agent key (bearer token)
		agentClient := adminClient.WithUnifiedAPIKey(agentKeyResp.FullToken)
		
		// This would typically be used for probe execution or metrics submission
		fmt.Printf("Monitoring agent client created with key: %s\n", agentKeyResp.KeyID)
		
		// Example: Check if the agent client is working
		_ = agentClient // Use the variable to avoid unused error
	}

	// Example 6: Using registration key for server registration
	if regKeyResp != nil {
		fmt.Println("\n=== Using Registration Key ===")
		
		// Create a client with the registration key
		regClient := adminClient.WithRegistrationKey(regKeyResp.FullToken)
		
		// Register a new server
		serverReq := &nexmonyx.ServerCreateRequest{
			Hostname:       "test-server-001",
			MainIP:         "192.168.1.100",
			OS:             "Linux",
			OSVersion:      "Ubuntu 22.04",
			OSArch:         "x86_64",
			SerialNumber:   "TEST001",
			MacAddress:     "aa:bb:cc:dd:ee:ff",
			Environment:    "testing",
			Location:       "Test Lab",
			Classification: "test",
		}

		server, err := regClient.Servers.RegisterWithKey(ctx, regKeyResp.FullToken, serverReq)
		if err != nil {
			log.Printf("Failed to register server: %v", err)
		} else {
			fmt.Printf("Registered server: %s (UUID: %s)\n", server.Hostname, server.ServerUUID)
		}
	}

	// Example 7: List and manage API keys
	fmt.Println("\n=== Managing API Keys ===")
	
	// List all API keys (admin only)
	keys, meta, err := adminClient.APIKeys.AdminListUnified(ctx, &nexmonyx.ListUnifiedAPIKeysOptions{
		ListOptions: nexmonyx.ListOptions{Page: 1, Limit: 10},
		Status:      nexmonyx.APIKeyStatusActive,
	})
	if err != nil {
		log.Printf("Failed to list API keys: %v", err)
	} else {
		fmt.Printf("Found %d active API keys (total %d)\n", len(keys), meta.TotalItems)
		for _, key := range keys {
			fmt.Printf("- %s (%s): %s\n", key.Name, key.Type, key.KeyID)
		}
	}

	// Example 8: Key validation and capabilities
	if len(keys) > 0 {
		key := keys[0]
		fmt.Printf("\n=== Key Information: %s ===\n", key.Name)
		fmt.Printf("Type: %s\n", key.Type)
		fmt.Printf("Status: %s\n", key.Status)
		fmt.Printf("Active: %t\n", key.IsActive())
		fmt.Printf("Can register servers: %t\n", key.CanRegisterServers())
		fmt.Printf("Is monitoring agent: %t\n", key.IsMonitoringAgent())
		fmt.Printf("Auth method: %s\n", key.GetAuthenticationMethod())
		fmt.Printf("Capabilities: %v\n", key.Capabilities)
		
		// Check specific capabilities
		if key.HasCapability(nexmonyx.CapabilityServersRead) {
			fmt.Println("✓ Can read servers")
		}
		if key.HasCapability(nexmonyx.CapabilityMetricsSubmit) {
			fmt.Println("✓ Can submit metrics")
		}
	}

	fmt.Println("\n=== Example completed ===")
}