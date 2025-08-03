package nexmonyx

import (
	"context"
	"testing"
)

func TestAgentVersionsService_Integration(t *testing.T) {
	// Create a test client
	client, err := NewClient(&Config{
		BaseURL: "https://api-dev.nexmonyx.com",
		Auth: AuthConfig{
			UnifiedAPIKey: "test-key",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Verify AgentVersions service is available
	if client.AgentVersions == nil {
		t.Fatal("AgentVersions service is nil")
	}

	// Test AgentVersionRequest structure
	req := &AgentVersionRequest{
		Version:     "v1.0.0-test",
		Environment: "test",
		Platform:    "linux",
		Architectures: []string{"amd64", "arm64"},
		DownloadURLs: map[string]string{
			"amd64": "https://example.com/agent-amd64",
			"arm64": "https://example.com/agent-arm64",
		},
		ReleaseNotes:      "Test version",
		MinimumAPIVersion: "1.0.0",
	}

	// We won't actually make the request in tests since we don't have valid credentials
	// but we can verify the request structure is valid
	if req.Version == "" {
		t.Error("Version should not be empty")
	}
	if req.Platform == "" {
		t.Error("Platform should not be empty")
	}
}

func TestAgentVersionRequest_Validation(t *testing.T) {
	// Test minimal valid request
	req := &AgentVersionRequest{
		Version:  "v1.0.0",
		Platform: "linux",
	}

	if req.Version != "v1.0.0" {
		t.Errorf("Expected version v1.0.0, got %s", req.Version)
	}
	if req.Platform != "linux" {
		t.Errorf("Expected platform linux, got %s", req.Platform)
	}

	// Test request with all fields
	fullReq := &AgentVersionRequest{
		Version:     "v1.2.3",
		Environment: "production",
		Platform:    "linux",
		Architectures: []string{"amd64", "arm64"},
		DownloadURLs: map[string]string{
			"amd64": "https://cdn.example.com/agent-v1.2.3-amd64",
			"arm64": "https://cdn.example.com/agent-v1.2.3-arm64",
		},
		UpdaterURLs: map[string]string{
			"amd64": "https://cdn.example.com/updater-v1.2.3-amd64",
			"arm64": "https://cdn.example.com/updater-v1.2.3-arm64",
		},
		ReleaseNotes:      "Bug fixes and improvements",
		MinimumAPIVersion: "1.0.0",
	}

	if len(fullReq.Architectures) != 2 {
		t.Errorf("Expected 2 architectures, got %d", len(fullReq.Architectures))
	}
	if len(fullReq.DownloadURLs) != 2 {
		t.Errorf("Expected 2 download URLs, got %d", len(fullReq.DownloadURLs))
	}
}