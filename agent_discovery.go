package nexmonyx

import (
	"context"
	"fmt"
)

// AgentDiscoveryService handles agent discovery operations for dynamic ingestor URL discovery.
// This service enables agents to query the discovery API to obtain their assigned ingestor
// WebSocket URL, supporting load balancing, failover, and geographic routing.
type AgentDiscoveryService struct {
	client *Client
}

// DiscoveryResponse represents the response from the agent discovery endpoint.
// It contains the assigned ingestor URL, fallback URLs for resilience, and caching parameters.
type DiscoveryResponse struct {
	// IngestorURL is the primary WebSocket URL the agent should connect to
	IngestorURL string `json:"ingestor_url"`

	// FallbackURLs is a list of alternative ingestor URLs to use if the primary fails
	FallbackURLs []string `json:"fallback_urls"`

	// TTLSeconds indicates how long (in seconds) the agent should cache this discovery response
	TTLSeconds int `json:"ttl_seconds"`

	// CheckIntervalSeconds indicates how often (in seconds) the agent should re-query the discovery API
	CheckIntervalSeconds int `json:"check_interval_seconds"`

	// AssignedPod is the specific ingestor pod name assigned to this agent (optional, for load balancing)
	AssignedPod string `json:"assigned_pod,omitempty"`

	// AssignedRegion is the geographic region of the assigned ingestor (optional, for geographic routing)
	AssignedRegion string `json:"assigned_region,omitempty"`

	// OrganizationTier indicates the organization's service tier (e.g., "shared", "dedicated")
	OrganizationTier string `json:"organization_tier"`
}

// Discover queries the discovery API to obtain the assigned ingestor URL for the agent.
//
// This method requires server credential authentication (ServerUUID and ServerSecret in AuthConfig).
// The discovery endpoint uses server credentials to identify the agent and return the appropriate
// ingestor URL based on load balancing, organization tier, and geographic routing rules.
//
// Parameters:
//   - ctx: Context for timeout and cancellation control
//
// Returns:
//   - *DiscoveryResponse: The discovery response containing ingestor URL and caching parameters
//   - error: Error if the request fails (authentication, network, server error, etc.)
//
// Example usage:
//
//	client, _ := nexmonyx.NewClient(&nexmonyx.Config{
//	    BaseURL: "https://api.nexmonyx.com",
//	    Auth: nexmonyx.AuthConfig{
//	        ServerUUID:   "550e8400-e29b-41d4-a716-446655440000",
//	        ServerSecret: "your-server-secret",
//	    },
//	})
//
//	discovery, err := client.AgentDiscovery.Discover(ctx)
//	if err != nil {
//	    log.Fatalf("Discovery failed: %v", err)
//	}
//
//	log.Printf("Connect to: %s", discovery.IngestorURL)
//	log.Printf("Cache TTL: %d seconds", discovery.TTLSeconds)
func (s *AgentDiscoveryService) Discover(ctx context.Context) (*DiscoveryResponse, error) {
	var resp StandardResponse
	resp.Data = &DiscoveryResponse{}

	// The discovery endpoint expects Server-UUID and Server-Secret headers (without X- prefix)
	// We need to override the default X-Server-UUID/X-Server-Secret headers set by the client
	headers := make(map[string]string)
	if s.client.config.Auth.ServerUUID != "" {
		headers["Server-UUID"] = s.client.config.Auth.ServerUUID
	}
	if s.client.config.Auth.ServerSecret != "" {
		headers["Server-Secret"] = s.client.config.Auth.ServerSecret
	}

	_, err := s.client.Do(ctx, &Request{
		Method:  "GET",
		Path:    "/v1/agents/discovery",
		Headers: headers,
		Result:  &resp,
	})
	if err != nil {
		return nil, err
	}

	// Type assert the response data to DiscoveryResponse
	if discovery, ok := resp.Data.(*DiscoveryResponse); ok {
		return discovery, nil
	}

	return nil, fmt.Errorf("unexpected response type: expected *DiscoveryResponse, got %T", resp.Data)
}
