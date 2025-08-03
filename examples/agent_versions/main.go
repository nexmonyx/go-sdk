package main

import (
	"context"
	"fmt"
	"log"

	"github.com/nexmonyx/go-sdk"
)

func main() {
	// Create client with API key authentication
	client, err := nexmonyx.NewClient(&nexmonyx.Config{
		BaseURL: "https://api.nexmonyx.com",
		Auth: nexmonyx.AuthConfig{
			UnifiedAPIKey: "your-api-key-here",
		},
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Example 1: Register a new agent version
	fmt.Println("=== Registering Agent Version ===")
	versionReq := &nexmonyx.AgentVersionRequest{
		Version:     "v1.3.4",
		Environment: "production",
		Platform:    "linux",
		Architectures: []string{"amd64", "arm64"},
		DownloadURLs: map[string]string{
			"amd64": "https://cdn.nexmonyx.com/agent/v1.3.4/nexmonyx-agent-amd64",
			"arm64": "https://cdn.nexmonyx.com/agent/v1.3.4/nexmonyx-agent-arm64",
		},
		UpdaterURLs: map[string]string{
			"amd64": "https://cdn.nexmonyx.com/agent/v1.3.4/nexmonyx-updater-amd64",
			"arm64": "https://cdn.nexmonyx.com/agent/v1.3.4/nexmonyx-updater-arm64",
		},
		ReleaseNotes:      "Bug fixes and performance improvements",
		MinimumAPIVersion: "1.0.0",
	}

	// Register the version (simple registration without response)
	err = client.AgentVersions.RegisterVersion(ctx, versionReq)
	if err != nil {
		log.Printf("Failed to register version: %v", err)
	} else {
		fmt.Printf("Successfully registered version %s\n", versionReq.Version)
	}

	// Example 2: Create version and get response
	fmt.Println("\n=== Creating Agent Version with Response ===")
	createdVersion, err := client.AgentVersions.CreateVersion(ctx, versionReq)
	if err != nil {
		log.Printf("Failed to create version: %v", err)
	} else {
		fmt.Printf("Created version: %s (ID: %d)\n", createdVersion.Version, createdVersion.ID)
	}

	// Example 3: List existing versions
	fmt.Println("\n=== Listing Agent Versions ===")
	versions, meta, err := client.AgentVersions.ListVersions(ctx, &nexmonyx.ListOptions{
		Limit: 10,
	})
	if err != nil {
		log.Printf("Failed to list versions: %v", err)
	} else {
		fmt.Printf("Found %d versions (page %d of %d):\n", len(versions), meta.Page, meta.TotalPages)
		for _, version := range versions {
			fmt.Printf("  - %s (%s) - %s\n", version.Version, version.Platform, version.ReleaseNotes)
		}
	}

	// Example 4: Get specific version
	fmt.Println("\n=== Getting Specific Version ===")
	version, err := client.AgentVersions.GetVersion(ctx, "v1.3.4")
	if err != nil {
		log.Printf("Failed to get version: %v", err)
	} else {
		fmt.Printf("Version: %s\n", version.Version)
		fmt.Printf("Platform: %s\n", version.Platform)
		fmt.Printf("Architectures: %v\n", version.Architectures)
		fmt.Printf("Release Date: %v\n", version.ReleaseDate)
	}

	// Example 5: Using admin methods (requires admin privileges)
	fmt.Println("\n=== Admin Methods Example ===")
	adminVersion, err := client.AgentVersions.AdminCreateVersion(ctx, "v1.3.5-admin", "Created via admin API")
	if err != nil {
		log.Printf("Failed to create admin version: %v", err)
	} else {
		fmt.Printf("Admin created version: %s (ID: %d)\n", adminVersion.Version, adminVersion.ID)

		// Add binaries using admin method
		for _, arch := range []string{"amd64", "arm64"} {
			binaryReq := &nexmonyx.AgentBinaryRequest{
				Platform:     "linux",
				Architecture: arch,
				DownloadURL:  fmt.Sprintf("https://cdn.nexmonyx.com/agent/v1.3.5/nexmonyx-agent-%s", arch),
				FileHash:     fmt.Sprintf("sha256:placeholder-%s", arch),
			}

			err = client.AgentVersions.AdminAddBinary(ctx, adminVersion.ID, binaryReq)
			if err != nil {
				log.Printf("Failed to add binary for %s: %v", arch, err)
			} else {
				fmt.Printf("Added binary for %s\n", arch)
			}
		}
	}
}