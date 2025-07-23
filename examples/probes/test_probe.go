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

	// List available regions
	fmt.Println("=== Available Monitoring Regions ===")
	regions, err := client.Probes.GetAvailableRegions(ctx)
	if err != nil {
		log.Printf("Failed to get regions: %v", err)
	} else {
		for _, region := range regions {
			fmt.Printf("- %s: %s, %s (%s)\n", region.Code, region.Name, region.Country, region.Continent)
		}
	}

	// Create a test ICMP probe
	fmt.Println("\n=== Creating Test ICMP Probe ===")
	
	icmpProbe := &nexmonyx.ProbeCreateRequest{
		Name:        "Test ICMP Probe",
		Description: "Testing ICMP probe creation",
		Type:        "icmp",
		Target:      "8.8.8.8", // This will be put in config["host"] by the service
		Frequency:   300, // 5 minutes
		Regions:     []string{"NYC3"},
		Enabled:     true,
	}

	probe, err := client.Probes.Create(ctx, icmpProbe)
	if err != nil {
		log.Printf("Failed to create ICMP probe: %v", err)
		
		// Try HTTP probe instead
		fmt.Println("\n=== Trying HTTP Probe ===")
		httpProbe := &nexmonyx.ProbeCreateRequest{
			Name:        "Test HTTP Probe",
			Description: "Testing HTTP probe creation",
			Type:        "http",
			Target:      "https://example.com",
			Frequency:   300,
			Regions:     []string{"NYC3"},
			Enabled:     true,
		}
		
		probe, err = client.Probes.Create(ctx, httpProbe)
		if err != nil {
			log.Printf("Failed to create HTTP probe: %v", err)
			
			// Try heartbeat probe
			fmt.Println("\n=== Trying Heartbeat Probe ===")
			heartbeatProbe := &nexmonyx.ProbeCreateRequest{
				Name:        "Test Heartbeat Probe",
				Description: "Testing heartbeat probe creation",
				Type:        "heartbeat",
				Target:      "https://example.com/heartbeat",
				Frequency:   300,
				Regions:     []string{"NYC3"},
				Enabled:     true,
			}
			
			probe, err = client.Probes.Create(ctx, heartbeatProbe)
			if err != nil {
				log.Printf("Failed to create heartbeat probe: %v", err)
			} else {
				fmt.Printf("Created heartbeat probe: %s (ID: %d)\n", probe.Name, probe.ID)
			}
		} else {
			fmt.Printf("Created HTTP probe: %s (ID: %d)\n", probe.Name, probe.ID)
		}
	} else {
		fmt.Printf("Created ICMP probe: %s (ID: %d)\n", probe.Name, probe.ID)
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