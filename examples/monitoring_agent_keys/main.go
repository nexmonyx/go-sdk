package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/nexmonyx/go-sdk"
)

func main() {
	// This example demonstrates how to manage monitoring agent keys
	// Both admin operations (for region enrollment) and customer operations (self-service)

	// For admin operations, use JWT token authentication
	adminConfig := &nexmonyx.Config{
		BaseURL: "https://api-dev.nexmonyx.com", // Use dev environment
		Auth: nexmonyx.AuthConfig{
			Token: os.Getenv("NEXMONYX_JWT_TOKEN"), // JWT token for admin access
		},
		Debug: true, // Enable debug logging
	}

	adminClient, err := nexmonyx.NewClient(adminConfig)
	if err != nil {
		log.Fatalf("Failed to create admin client: %v", err)
	}

	ctx := context.Background()

	fmt.Println("=== Admin Operations ===")

	// Admin: Create monitoring agent key for region enrollment
	adminKeyReq := &nexmonyx.CreateMonitoringAgentKeyRequest{
		OrganizationID:  1, // Target organization ID
		RemoteClusterID: nil, // No cluster restriction
		Description:     "Test region monitoring agent key",
		Capabilities:    "probe_execution,heartbeat",
	}

	fmt.Printf("Creating admin monitoring agent key...\n")
	agentKeyResp, err := adminClient.MonitoringAgentKeys.CreateAdmin(ctx, adminKeyReq)
	if err != nil {
		log.Printf("Failed to create admin monitoring agent key: %v", err)
	} else {
		fmt.Printf("✅ Created monitoring agent key: %s\n", agentKeyResp.FullToken)
		fmt.Printf("   Key ID: %s\n", agentKeyResp.KeyID)
		fmt.Printf("   Description: %s\n", agentKeyResp.Key.Description)
	}

	fmt.Println("\n=== Customer Operations ===")

	// For customer operations, use API key authentication or JWT token
	customerConfig := &nexmonyx.Config{
		BaseURL: "https://api-dev.nexmonyx.com",
		Auth: nexmonyx.AuthConfig{
			// You can use either JWT token or API key/secret
			Token: os.Getenv("NEXMONYX_JWT_TOKEN"),
			// Or: APIKey/APISecret for service authentication
			// APIKey:    os.Getenv("NEXMONYX_API_KEY"),
			// APISecret: os.Getenv("NEXMONYX_API_SECRET"),
		},
		Debug: true,
	}

	customerClient, err := nexmonyx.NewClient(customerConfig)
	if err != nil {
		log.Fatalf("Failed to create customer client: %v", err)
	}

	organizationID := "1" // Replace with actual organization UUID/ID

	// Customer: Create monitoring agent key for their own use
	fmt.Printf("Creating customer monitoring agent key...\n")
	customerKeyResp, err := customerClient.MonitoringAgentKeys.Create(ctx, 
		organizationID, 
		"Development environment monitoring")
	if err != nil {
		log.Printf("Failed to create customer monitoring agent key: %v", err)
	} else {
		fmt.Printf("✅ Created customer monitoring agent key: %s\n", customerKeyResp.FullToken)
	}

	// Customer: List monitoring agent keys
	fmt.Printf("\nListing monitoring agent keys...\n")
	keys, pagination, err := customerClient.MonitoringAgentKeys.List(ctx, organizationID, &nexmonyx.ListMonitoringAgentKeysOptions{
		Page:  1,
		Limit: 10,
	})
	if err != nil {
		log.Printf("Failed to list monitoring agent keys: %v", err)
	} else {
		fmt.Printf("Found %d monitoring agent keys (page %d of %d):\n", 
			len(keys), pagination.Page, pagination.TotalPages)
		for i, key := range keys {
			fmt.Printf("  %d. %s - %s (%s)\n", 
				i+1, key.KeyPrefix, key.Description, key.Status)
		}
	}

	// Example: Revoke a monitoring agent key (if needed)
	if len(keys) > 0 && keys[0].Status == "active" {
		fmt.Printf("\nRevoking monitoring agent key: %s...\n", keys[0].KeyID)
		err = customerClient.MonitoringAgentKeys.Revoke(ctx, organizationID, keys[0].KeyID)
		if err != nil {
			log.Printf("Failed to revoke monitoring agent key: %v", err)
		} else {
			fmt.Printf("✅ Successfully revoked monitoring agent key\n")
		}
	}

	fmt.Println("\n=== Usage Tips ===")
	fmt.Println("1. Admin operations require JWT token with admin privileges")
	fmt.Println("2. Customer operations work with JWT token or API key/secret")
	fmt.Println("3. The full token format is: mag_<keyID>.<secretKey>")
	fmt.Println("4. Use the full token in monitoring-agent configuration")
	fmt.Println("5. Keys can be filtered by namespace, cluster, or status")
}