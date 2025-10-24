package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAgentDiscoveryService_Discover(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse interface{}
		statusCode     int
		serverUUID     string
		serverSecret   string
		wantErr        bool
		errContains    string
		validate       func(t *testing.T, resp *DiscoveryResponse)
	}{
		{
			name: "successful discovery",
			serverResponse: StandardResponse{
				Status:  "success",
				Message: "Discovery successful",
				Data: map[string]interface{}{
					"ingestor_url":           "wss://ingestor-pod-1.nexmonyx.com",
					"fallback_urls":          []interface{}{"wss://ingestor-pod-2.nexmonyx.com", "wss://ingestor-pod-3.nexmonyx.com"},
					"ttl_seconds":            3600,
					"check_interval_seconds": 300,
					"assigned_pod":           "ingestor-pod-1",
					"assigned_region":        "us-west-2",
					"organization_tier":      "shared",
				},
			},
			statusCode:   http.StatusOK,
			serverUUID:   "550e8400-e29b-41d4-a716-446655440000",
			serverSecret: "test-server-secret",
			wantErr:      false,
			validate: func(t *testing.T, resp *DiscoveryResponse) {
				if resp.IngestorURL != "wss://ingestor-pod-1.nexmonyx.com" {
					t.Errorf("IngestorURL = %v, want wss://ingestor-pod-1.nexmonyx.com", resp.IngestorURL)
				}
				if len(resp.FallbackURLs) != 2 {
					t.Errorf("FallbackURLs length = %d, want 2", len(resp.FallbackURLs))
				}
				if resp.TTLSeconds != 3600 {
					t.Errorf("TTLSeconds = %d, want 3600", resp.TTLSeconds)
				}
				if resp.CheckIntervalSeconds != 300 {
					t.Errorf("CheckIntervalSeconds = %d, want 300", resp.CheckIntervalSeconds)
				}
				if resp.AssignedPod != "ingestor-pod-1" {
					t.Errorf("AssignedPod = %v, want ingestor-pod-1", resp.AssignedPod)
				}
				if resp.AssignedRegion != "us-west-2" {
					t.Errorf("AssignedRegion = %v, want us-west-2", resp.AssignedRegion)
				}
				if resp.OrganizationTier != "shared" {
					t.Errorf("OrganizationTier = %v, want shared", resp.OrganizationTier)
				}
			},
		},
		{
			name: "minimal response (optional fields empty)",
			serverResponse: StandardResponse{
				Status:  "success",
				Message: "Discovery successful",
				Data: map[string]interface{}{
					"ingestor_url":           "wss://shared.nexmonyx.com",
					"fallback_urls":          []interface{}{"wss://fallback.nexmonyx.com"},
					"ttl_seconds":            1800,
					"check_interval_seconds": 600,
					"organization_tier":      "free",
				},
			},
			statusCode:   http.StatusOK,
			serverUUID:   "test-uuid",
			serverSecret: "test-secret",
			wantErr:      false,
			validate: func(t *testing.T, resp *DiscoveryResponse) {
				if resp.IngestorURL != "wss://shared.nexmonyx.com" {
					t.Errorf("IngestorURL = %v, want wss://shared.nexmonyx.com", resp.IngestorURL)
				}
				if resp.AssignedPod != "" {
					t.Errorf("AssignedPod = %v, want empty string", resp.AssignedPod)
				}
				if resp.AssignedRegion != "" {
					t.Errorf("AssignedRegion = %v, want empty string", resp.AssignedRegion)
				}
			},
		},
		{
			name: "unauthorized - invalid credentials",
			serverResponse: map[string]interface{}{
				"status":  "error",
				"error":   "unauthorized",
				"message": "Invalid server credentials",
				"details": "Server UUID or secret is incorrect",
			},
			statusCode:   http.StatusUnauthorized,
			serverUUID:   "invalid-uuid",
			serverSecret: "invalid-secret",
			wantErr:      true,
			errContains:  "Invalid server credentials",
		},
		{
			name: "forbidden - server disabled",
			serverResponse: map[string]interface{}{
				"status":  "error",
				"error":   "forbidden",
				"message": "Server is disabled",
				"details": "This server has been disabled by the organization",
			},
			statusCode:   http.StatusForbidden,
			serverUUID:   "disabled-uuid",
			serverSecret: "test-secret",
			wantErr:      true,
			errContains:  "Server is disabled",
		},
		{
			name: "forbidden - organization disabled",
			serverResponse: map[string]interface{}{
				"status":  "error",
				"error":   "forbidden",
				"message": "Organization is disabled",
				"details": "Your organization has been disabled",
			},
			statusCode:   http.StatusForbidden,
			serverUUID:   "test-uuid",
			serverSecret: "test-secret",
			wantErr:      true,
			errContains:  "Organization is disabled",
		},
		{
			name: "not found - server not found",
			serverResponse: map[string]interface{}{
				"status":  "error",
				"error":   "not_found",
				"message": "Server not found",
				"details": "No server exists with the provided UUID",
			},
			statusCode:   http.StatusNotFound,
			serverUUID:   "nonexistent-uuid",
			serverSecret: "test-secret",
			wantErr:      true,
			errContains:  "Server not found",
		},
		{
			name: "internal server error",
			serverResponse: map[string]interface{}{
				"status":  "error",
				"error":   "internal_server_error",
				"message": "Failed to retrieve organization",
				"details": "Database connection failed",
			},
			statusCode:   http.StatusInternalServerError,
			serverUUID:   "test-uuid",
			serverSecret: "test-secret",
			wantErr:      true,
			errContains:  "Failed to retrieve organization",
		},
		{
			name: "service unavailable",
			serverResponse: map[string]interface{}{
				"status":  "error",
				"error":   "service_unavailable",
				"message": "Service temporarily unavailable",
				"details": "Please try again later",
			},
			statusCode:   http.StatusServiceUnavailable,
			serverUUID:   "test-uuid",
			serverSecret: "test-secret",
			wantErr:      true,
			errContains:  "Service temporarily unavailable",
		},
		{
			name: "missing server credentials",
			serverResponse: map[string]interface{}{
				"status":  "error",
				"error":   "unauthorized",
				"message": "Server credentials required",
				"details": "Server-UUID and Server-Secret headers required",
			},
			statusCode:   http.StatusUnauthorized,
			serverUUID:   "", // Empty credentials
			serverSecret: "",
			wantErr:      true,
			errContains:  "Server credentials required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify headers are sent correctly (without X- prefix)
				if tt.serverUUID != "" {
					if got := r.Header.Get("Server-UUID"); got != tt.serverUUID {
						t.Errorf("Server-UUID header = %v, want %v", got, tt.serverUUID)
					}
				}
				if tt.serverSecret != "" {
					if got := r.Header.Get("Server-Secret"); got != tt.serverSecret {
						t.Errorf("Server-Secret header = %v, want %v", got, tt.serverSecret)
					}
				}

				// Verify endpoint
				if r.URL.Path != "/v1/agents/discovery" {
					t.Errorf("Request path = %v, want /v1/agents/discovery", r.URL.Path)
				}

				// Verify method
				if r.Method != http.MethodGet {
					t.Errorf("Request method = %v, want GET", r.Method)
				}

				// Send response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(tt.serverResponse)
			}))
			defer server.Close()

			// Create client
			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth: AuthConfig{
					ServerUUID:   tt.serverUUID,
					ServerSecret: tt.serverSecret,
				},
			})
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			// Call Discover
			ctx := context.Background()
			resp, err := client.AgentDiscovery.Discover(ctx)

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("Discover() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If error expected, verify error message
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !contains(err.Error(), tt.errContains) {
					t.Errorf("Discover() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			// Validate successful response
			if tt.validate != nil {
				tt.validate(t, resp)
			}
		})
	}
}

func TestAgentDiscoveryService_Discover_ContextCancellation(t *testing.T) {
	// Create test server with delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(StandardResponse{
			Status:  "success",
			Message: "Discovery successful",
			Data: map[string]interface{}{
				"ingestor_url":           "wss://test.nexmonyx.com",
				"fallback_urls":          []interface{}{},
				"ttl_seconds":            3600,
				"check_interval_seconds": 300,
				"organization_tier":      "shared",
			},
		})
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			ServerUUID:   "test-uuid",
			ServerSecret: "test-secret",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Create context with immediate cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Call Discover - should fail due to cancelled context
	_, err = client.AgentDiscovery.Discover(ctx)
	if err == nil {
		t.Error("Discover() expected error for cancelled context, got nil")
	}
}

func TestAgentDiscoveryService_Discover_ContextTimeout(t *testing.T) {
	// Create test server with long delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(StandardResponse{
			Status:  "success",
			Message: "Discovery successful",
			Data: map[string]interface{}{
				"ingestor_url":           "wss://test.nexmonyx.com",
				"fallback_urls":          []interface{}{},
				"ttl_seconds":            3600,
				"check_interval_seconds": 300,
				"organization_tier":      "shared",
			},
		})
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			ServerUUID:   "test-uuid",
			ServerSecret: "test-secret",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Call Discover - should timeout
	_, err = client.AgentDiscovery.Discover(ctx)
	if err == nil {
		t.Error("Discover() expected timeout error, got nil")
	}
}

func TestAgentDiscoveryService_Discover_InvalidResponseType(t *testing.T) {
	// Create test server returning unexpected type
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Return a string instead of DiscoveryResponse
		json.NewEncoder(w).Encode(StandardResponse{
			Status:  "success",
			Message: "Discovery successful",
			Data:    "invalid-type", // Wrong type - should be map/object
		})
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			ServerUUID:   "test-uuid",
			ServerSecret: "test-secret",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Call Discover - should fail JSON unmarshaling
	ctx := context.Background()
	_, err = client.AgentDiscovery.Discover(ctx)
	if err == nil {
		t.Error("Discover() expected error for invalid response type, got nil")
	}
	// SDK returns JSON unmarshaling error when type is wrong
	if !contains(err.Error(), "json: cannot unmarshal") && !contains(err.Error(), "unexpected response type") {
		t.Errorf("Discover() error = %v, want error containing unmarshaling or type error", err)
	}
}
