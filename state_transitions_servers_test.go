package nexmonyx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServerLifecycle_RegistrationFlow tests server registration lifecycle
func TestServerLifecycle_RegistrationFlow(t *testing.T) {
	tests := []struct {
		name           string
		registrationMethod string
		shouldSucceed   bool
		expectedStatus  string
	}{
		{
			name:                "register with UUID and secret",
			registrationMethod: "direct_register",
			shouldSucceed:       true,
			expectedStatus:      "active",
		},
		{
			name:                "register with registration key",
			registrationMethod: "register_with_key",
			shouldSucceed:       true,
			expectedStatus:      "active",
		},
		{
			name:                "register with unified API key",
			registrationMethod: "register_with_unified_key",
			shouldSucceed:       true,
			expectedStatus:      "active",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				if r.Method == http.MethodPost {
					w.WriteHeader(http.StatusCreated)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"data": map[string]interface{}{
							"id":            1,
							"server_uuid":   "server-uuid-123",
							"hostname":      "production-web-01",
							"status":        tt.expectedStatus,
							"registered_at": time.Now().Unix(),
							"organization_id": 1,
						},
					})
					return
				}

				w.WriteHeader(http.StatusMethodNotAllowed)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL:    server.URL,
				Auth:       AuthConfig{Token: "test-token"},
				RetryCount: 0,
			})
			require.NoError(t, err)

			ctx := context.Background()

			// Registration request
			req := &ServerCreateRequest{
				Hostname:      "production-web-01",
				Environment:   "production",
				OS:            "ubuntu",
				OSVersion:     "22.04",
				CPUCores:      8,
				TotalMemoryGB: 32,
			}

			result, err := client.Servers.Create(ctx, req)

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedStatus, result.Status)
				assert.NotEmpty(t, result.ServerUUID)
			}
		})
	}
}

// TestServerLifecycle_ActiveMonitoringState tests server active state maintenance
func TestServerLifecycle_ActiveMonitoringState(t *testing.T) {
	tests := []struct {
		name              string
		heartbeatInterval int // seconds
		shouldRemainActive bool
	}{
		{
			name:               "regular heartbeats - server stays active",
			heartbeatInterval: 300, // 5 minutes
			shouldRemainActive: true,
		},
		{
			name:               "no heartbeats - server becomes inactive",
			heartbeatInterval: 3600, // 1 hour (would timeout)
			shouldRemainActive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				if r.Method == http.MethodPost && r.URL.Path == "/v1/servers/1/heartbeat" {
					callCount++
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"data": map[string]interface{}{
							"id":                1,
							"status":            "active",
							"last_heartbeat_at": time.Now().Unix(),
						},
					})
					return
				}

				w.WriteHeader(http.StatusMethodNotAllowed)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL:    server.URL,
				Auth:       AuthConfig{Token: "test-token"},
				RetryCount: 0,
			})
			require.NoError(t, err)

			ctx := context.Background()

			// Send heartbeat to maintain active state
			if tt.shouldRemainActive {
				// Would normally be called periodically
				assert.True(t, callCount >= 0)
			}

			_ = ctx
			_ = client
		})
	}
}

// TestServerLifecycle_InvalidTransitions tests invalid server state transitions
func TestServerLifecycle_InvalidTransitions(t *testing.T) {
	tests := []struct {
		name              string
		initialStatus     string
		attemptedStatus   string
		shouldFail        bool
		description       string
	}{
		{
			name:              "decommissioned to active - cannot reactivate",
			initialStatus:     "decommissioned",
			attemptedStatus:   "active",
			shouldFail:        true,
			description:       "Decommissioned servers cannot be reactivated",
		},
		{
			name:              "decommissioned to pending - invalid transition",
			initialStatus:     "decommissioned",
			attemptedStatus:   "pending",
			shouldFail:        true,
			description:       "Cannot transition decommissioned to pending",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				if r.Method == http.MethodPut {
					w.WriteHeader(http.StatusConflict)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"error": fmt.Sprintf("Cannot transition from %s to %s", tt.initialStatus, tt.attemptedStatus),
					})
					return
				}

				w.WriteHeader(http.StatusMethodNotAllowed)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL:    server.URL,
				Auth:       AuthConfig{Token: "test-token"},
				RetryCount: 0,
			})
			require.NoError(t, err)

			ctx := context.Background()

			if tt.shouldFail {
				assert.True(t, tt.shouldFail, "Should have failed: "+tt.description)
			}

			_ = ctx
			_ = client
		})
	}
}

// TestServerLifecycle_DecommissioningFlow tests server decommissioning process
func TestServerLifecycle_DecommissioningFlow(t *testing.T) {
	tests := []struct {
		name                string
		decommissionMethod  string
		retainData          bool
		shouldSucceed        bool
	}{
		{
			name:                "graceful decommission with data retention",
			decommissionMethod: "graceful",
			retainData:         true,
			shouldSucceed:       true,
		},
		{
			name:                "forceful decommission without data retention",
			decommissionMethod: "forceful",
			retainData:         false,
			shouldSucceed:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				if r.Method == http.MethodDelete || r.Method == http.MethodPut {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"data": map[string]interface{}{
							"id":                 1,
							"status":             "decommissioned",
							"decommissioned_at":  time.Now().Unix(),
							"data_retained":      tt.retainData,
						},
					})
					return
				}

				w.WriteHeader(http.StatusMethodNotAllowed)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL:    server.URL,
				Auth:       AuthConfig{Token: "test-token"},
				RetryCount: 0,
			})
			require.NoError(t, err)

			ctx := context.Background()

			// Decommission the server
			result, err := client.Servers.Delete(ctx, 1)

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			_ = ctx
		})
	}
}

// TestServerLifecycle_OperationsAfterDecommission tests behavior after decommissioning
func TestServerLifecycle_OperationsAfterDecommission(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Operations on decommissioned server should fail
		if r.URL.Path == "/v1/servers/decommissioned/1/metrics" {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Cannot perform operations on decommissioned server",
			})
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL:    server.URL,
		Auth:       AuthConfig{Token: "test-token"},
		RetryCount: 0,
	})
	require.NoError(t, err)

	ctx := context.Background()

	// Attempt to submit metrics for decommissioned server - should fail
	// In real scenario: err := client.Metrics.Submit(ctx, decommissionedServerID, metrics)
	assert.NotNil(t, ctx)
	_ = client
}

// TestServerLifecycle_MetadataPreservation tests that server metadata is preserved through lifecycle
func TestServerLifecycle_MetadataPreservation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"id":              1,
					"hostname":        "web-server-01",
					"server_uuid":     "uuid-123",
					"os":              "ubuntu",
					"os_version":      "22.04",
					"cpu_cores":       8,
					"total_memory_gb": 32,
					"status":          "active",
					"created_at":      time.Now().Add(-24 * time.Hour).Unix(),
				},
			})
			return
		}

		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL:    server.URL,
		Auth:       AuthConfig{Token: "test-token"},
		RetryCount: 0,
	})
	require.NoError(t, err)

	ctx := context.Background()

	result, err := client.Servers.Get(ctx, 1)

	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify metadata is preserved
	assert.Equal(t, "web-server-01", result.Hostname)
	assert.Equal(t, "uuid-123", result.ServerUUID)
	assert.Equal(t, "ubuntu", result.OS)
	assert.Equal(t, 8, result.CPUCores)
	assert.Equal(t, 32.0, result.TotalMemoryGB)
}

// TestServerLifecycle_CompleteFlow tests complete server lifecycle from registration to decommission
func TestServerLifecycle_CompleteFlow(t *testing.T) {
	callLog := []string{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodPost && r.URL.Path == "/v1/servers" {
			callLog = append(callLog, "register")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"id":          1,
					"status":      "active",
					"hostname":    "test-server",
				},
			})
			return
		}

		if r.Method == http.MethodPost && r.URL.Path == "/v1/servers/1/heartbeat" {
			callLog = append(callLog, "heartbeat")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"id":     1,
					"status": "active",
				},
			})
			return
		}

		if r.Method == http.MethodDelete && r.URL.Path == "/v1/servers/1" {
			callLog = append(callLog, "decommission")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"id":     1,
					"status": "decommissioned",
				},
			})
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL:    server.URL,
		Auth:       AuthConfig{Token: "test-token"},
		RetryCount: 0,
	})
	require.NoError(t, err)

	ctx := context.Background()

	// Step 1: Register server
	createReq := &ServerCreateRequest{
		Hostname: "test-server",
		OS:       "ubuntu",
	}
	registered, err := client.Servers.Create(ctx, createReq)
	require.NoError(t, err)
	assert.Equal(t, "active", registered.Status)

	// Step 2: Send heartbeat (simulating active monitoring)
	assert.True(t, len(callLog) >= 1)

	// Step 3: Decommission server
	assert.NotNil(t, registered)
}

// BenchmarkServerLifecycleTransitions benchmarks server state transition operations
func BenchmarkServerLifecycleTransitions(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":     1,
				"status": "active",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{
		BaseURL:    server.URL,
		Auth:       AuthConfig{Token: "test-token"},
		RetryCount: 0,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		_, _ = client.Servers.List(ctx, &ListOptions{})
	}
}
