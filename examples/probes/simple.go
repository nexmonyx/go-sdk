package main

import (
	"context"
	"fmt"
	"log"

	"github.com/nexmonyx/go-sdk"
)

func main() {
	// Use dev API and provided credentials
	client, err := nexmonyx.NewClient(&nexmonyx.Config{
		BaseURL: "https://api-dev.nexmonyx.com",
		Auth: nexmonyx.AuthConfig{
			APIKey:    "03nwb4Ql",
			APISecret: "SiYt4GaHh7xX3Pb87i9PwBQ6",
		},
		Debug: true,
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Get available regions first
	fmt.Println("Getting available regions...")
	regions, err := client.Probes.GetAvailableRegions(ctx)
	if err != nil {
		log.Printf("Failed to get regions: %v", err)
		// Use default region
		regions = []*nexmonyx.MonitoringRegion{{Code: "NYC3"}}
	}
	
	// Use first available region
	regionCodes := []string{}
	if len(regions) > 0 {
		regionCodes = append(regionCodes, regions[0].Code)
		fmt.Printf("Using region: %s\n", regions[0].Code)
	}

	// Try different probe types
	probeTypes := []struct {
		Type   string
		Target string
		Name   string
	}{
		{"http", "https://example.com", "Test HTTP Probe"},
		{"icmp", "8.8.8.8", "Test ICMP Probe"},
		{"heartbeat", "https://example.com/heartbeat", "Test Heartbeat Probe"},
		{"tcp", "example.com", "Test TCP Probe"},
	}

	for _, pt := range probeTypes {
		fmt.Printf("\n=== Creating %s ===\n", pt.Name)
		
		probe, err := client.Probes.CreateSimpleProbe(ctx, pt.Name, pt.Type, pt.Target, regionCodes)
		if err != nil {
			fmt.Printf("❌ Failed to create %s probe: %v\n", pt.Type, err)
		} else {
			fmt.Printf("✅ Created %s probe: %s (ID: %d)\n", pt.Type, probe.Name, probe.ID)
		}
	}

	// List all probes
	fmt.Println("\n=== Listing All Probes ===")
	probes, _, err := client.Probes.List(ctx, nil)
	if err != nil {
		log.Printf("Failed to list probes: %v", err)
	} else {
		fmt.Printf("Found %d probes:\n", len(probes))
		for _, p := range probes {
			fmt.Printf("- %s (ID: %d): %s - %s (Enabled: %v)\n", p.Name, p.ID, p.Type, p.Target, p.Enabled)
		}
	}
}