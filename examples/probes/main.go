package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/nexmonyx/go-sdk"
)

func main() {
	// Create client with API key authentication
	client, err := nexmonyx.NewClient(&nexmonyx.Config{
		BaseURL: "https://api.nexmonyx.com",
		Auth: nexmonyx.AuthConfig{
			APIKey:    os.Getenv("NEXMONYX_API_KEY"),
			APISecret: os.Getenv("NEXMONYX_API_SECRET"),
		},
		Debug: true,
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Example 1: List available probe types
	fmt.Println("=== Available Probe Types ===")
	probeTypes, err := client.Probes.GetAvailableProbeTypes(ctx)
	if err != nil {
		log.Printf("Failed to get probe types: %v", err)
	} else {
		for _, pt := range probeTypes {
			fmt.Printf("- %s\n", pt)
		}
	}

	// Example 2: List available regions
	fmt.Println("\n=== Available Monitoring Regions ===")
	regions, err := client.Probes.GetAvailableRegions(ctx)
	if err != nil {
		log.Printf("Failed to get regions: %v", err)
	} else {
		for _, region := range regions {
			fmt.Printf("- %s: %s, %s (%s)\n", region.Code, region.Name, region.Country, region.Continent)
		}
	}

	// Example 3: Create a new probe
	fmt.Println("\n=== Creating a New Probe ===")
	newProbe := &nexmonyx.ProbeCreateRequest{
		Name:        "Example HTTP Probe",
		Description: "Monitor example.com availability",
		Type:        nexmonyx.ProbeTypeHTTP,
		Scope:       nexmonyx.ProbeScopePublic,
		Target:      "https://example.com",
		Interval:    300, // 5 minutes
		Timeout:     30,  // 30 seconds
		Enabled:     true,
		Config: nexmonyx.ProbeConfig{
			Method:             strPtr("GET"),
			ExpectedStatusCode: intPtr(200),
			FollowRedirects:    boolPtr(true),
			ValidateCert:       boolPtr(true),
		},
		Regions:        []string{"NYC3", "SFO3"}, // New York and San Francisco
		AlertThreshold: 3,
		AlertEnabled:   true,
	}

	probe, err := client.Probes.Create(ctx, newProbe)
	if err != nil {
		log.Printf("Failed to create probe: %v", err)
	} else {
		fmt.Printf("Created probe: %s (UUID: %s)\n", probe.Name, probe.ProbeUUID)
	}

	// Example 4: List all probes
	fmt.Println("\n=== Listing All Probes ===")
	enabledFilter := true
	probes, _, err := client.Probes.List(ctx, &nexmonyx.ProbeListOptions{
		Enabled: &enabledFilter,
		ListOptions: nexmonyx.ListOptions{
			Page:  1,
			Limit: 10,
		},
	})
	if err != nil {
		log.Printf("Failed to list probes: %v", err)
	} else {
		for _, p := range probes {
			fmt.Printf("- %s (%s): %s - %s\n", p.Name, p.ProbeUUID, p.Type, p.Target)
		}
	}

	// Example 5: Get probe health
	if len(probes) > 0 {
		fmt.Println("\n=== Probe Health Status ===")
		health, err := client.Probes.GetHealth(ctx, probes[0].ProbeUUID)
		if err != nil {
			log.Printf("Failed to get probe health: %v", err)
		} else {
			fmt.Printf("Probe: %s\n", health.Name)
			fmt.Printf("Health Score: %.2f%%\n", health.HealthScore)
			fmt.Printf("24h Availability: %.2f%%\n", health.Availability24h)
			fmt.Printf("Average Response: %dms\n", health.AverageResponse)
			fmt.Printf("Last Status: %s\n", health.LastStatus)
			
			if len(health.RegionStatus) > 0 {
				fmt.Println("\nRegional Status:")
				for _, rs := range health.RegionStatus {
					fmt.Printf("  - %s (%s): %s, Availability: %.2f%%\n", 
						rs.Region, rs.RegionName, rs.LastStatus, rs.Availability24h)
				}
			}
		}
	}

	// Example 6: Update a probe
	if len(probes) > 0 {
		fmt.Println("\n=== Updating a Probe ===")
		updateReq := &nexmonyx.ProbeUpdateRequest{
			Description: strPtr("Updated description"),
			Interval:    intPtr(600), // Change to 10 minutes
		}

		updated, err := client.Probes.Update(ctx, probes[0].ProbeUUID, updateReq)
		if err != nil {
			log.Printf("Failed to update probe: %v", err)
		} else {
			fmt.Printf("Updated probe: %s - New interval: %d seconds\n", updated.Name, updated.Interval)
		}
	}

	// Example 7: Get probe results
	if len(probes) > 0 {
		fmt.Println("\n=== Recent Probe Results ===")
		results, _, err := client.Probes.ListResults(ctx, probes[0].ProbeUUID, &nexmonyx.ProbeResultListOptions{
			ListOptions: nexmonyx.ListOptions{
				Limit: 5,
			},
		})
		if err != nil {
			log.Printf("Failed to get probe results: %v", err)
		} else {
			for _, result := range results {
				fmt.Printf("- %s: %s (Response: %dms) - %s\n", 
					result.ExecutedAt.Format("2006-01-02 15:04:05"), 
					result.Status, 
					result.ResponseTime,
					result.Region)
			}
		}
	}
}

// Helper functions
func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}