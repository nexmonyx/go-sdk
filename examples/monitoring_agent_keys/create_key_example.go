package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/nexmonyx/go-sdk"
)

func main() {
	// Create client
	client, err := nexmonyx.NewClient(&nexmonyx.Config{
		BaseURL: "https://api.nexmonyx.com",
		Auth: nexmonyx.AuthConfig{
			Token: os.Getenv("NEXMONYX_TOKEN"), // Use JWT token
		},
	})
	if err != nil {
		log.Fatal("Failed to create client:", err)
	}

	ctx := context.Background()
	orgID := "114" // Your organization ID

	// Example 1: Create a private monitoring agent key
	fmt.Println("Creating private monitoring agent key...")
	privateKeyReq := nexmonyx.NewPrivateAgentKeyRequest(
		"My Private Monitoring Agent",
		"private-agent-1",
		"NYC3", // Region code is optional for private agents
	)

	privateKey, err := client.MonitoringAgentKeys.Create(ctx, orgID, privateKeyReq)
	if err != nil {
		log.Fatal("Failed to create private key:", err)
	}

	fmt.Printf("Private Agent Key Created:\n")
	fmt.Printf("  Key ID: %s\n", privateKey.KeyID)
	fmt.Printf("  Full Token: %s\n", privateKey.FullToken)
	fmt.Printf("  Agent Type: %s\n", privateKey.AgentType)
	fmt.Printf("  Allowed Scopes: %v\n", privateKey.AllowedProbeScopes)
	fmt.Println()

	// Example 2: Create a public monitoring agent key
	fmt.Println("Creating public monitoring agent key...")
	publicKeyReq := nexmonyx.NewPublicAgentKeyRequest(
		"NYC3 Public Monitoring Agent",
		"public-agent-nyc3",
		"NYC3", // Region code is REQUIRED for public agents
	)

	publicKey, err := client.MonitoringAgentKeys.Create(ctx, orgID, publicKeyReq)
	if err != nil {
		log.Fatal("Failed to create public key:", err)
	}

	fmt.Printf("Public Agent Key Created:\n")
	fmt.Printf("  Key ID: %s\n", publicKey.KeyID)
	fmt.Printf("  Full Token: %s\n", publicKey.FullToken)
	fmt.Printf("  Agent Type: %s\n", publicKey.AgentType)
	fmt.Printf("  Allowed Scopes: %v\n", publicKey.AllowedProbeScopes)
	fmt.Println()

	// Example 3: List monitoring agent keys
	fmt.Println("Listing monitoring agent keys...")
	keys, _, err := client.MonitoringAgentKeys.List(ctx, orgID, nil)
	if err != nil {
		log.Fatal("Failed to list keys:", err)
	}

	fmt.Printf("Found %d monitoring agent keys:\n", len(keys))
	for _, key := range keys {
		fmt.Printf("  - %s (%s): %s - Type: %s, Region: %s\n", 
			key.KeyID, 
			key.Status, 
			key.Description,
			key.AgentType,
			key.RegionCode,
		)
	}

	// Example 4: Admin endpoint (requires admin token)
	if os.Getenv("NEXMONYX_ADMIN_TOKEN") != "" {
		// Create admin client
		adminClient, err := nexmonyx.NewClient(&nexmonyx.Config{
			BaseURL: "https://api.nexmonyx.com",
			Auth: nexmonyx.AuthConfig{
				Token: os.Getenv("NEXMONYX_ADMIN_TOKEN"),
			},
		})
		if err != nil {
			log.Fatal("Failed to create admin client:", err)
		}

		fmt.Println("\nCreating key via admin endpoint...")
		adminReq := &nexmonyx.CreateMonitoringAgentKeyRequest{
			OrganizationID:     114,
			Description:        "Admin-created monitoring agent",
			NamespaceName:      "admin-agent-1",
			AgentType:          "private",
			RegionCode:         "NYC3",
			AllowedProbeScopes: []string{"public", "private"},
		}

		adminKey, err := adminClient.MonitoringAgentKeys.CreateAdmin(ctx, adminReq)
		if err != nil {
			log.Printf("Failed to create admin key: %v", err)
		} else {
			fmt.Printf("Admin Key Created:\n")
			fmt.Printf("  Key ID: %s\n", adminKey.KeyID)
			fmt.Printf("  Full Token: %s\n", adminKey.FullToken)
		}
	}
}