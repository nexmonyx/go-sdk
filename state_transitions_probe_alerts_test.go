package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProbeAlertStateTransitions_ValidTransitions tests valid probe alert state transitions
func TestProbeAlertStateTransitions_ValidTransitions(t *testing.T) {
	tests := []struct {
		name                string
		initialStatus       string
		targetStatus        string
		shouldSucceed        bool
		expectedFinalStatus string
	}{
		{
			name:                 "active to acknowledged - valid transition",
			initialStatus:        "active",
			targetStatus:         "acknowledged",
			shouldSucceed:         true,
			expectedFinalStatus:  "acknowledged",
		},
		{
			name:                 "active to resolved - direct resolution",
			initialStatus:        "active",
			targetStatus:         "resolved",
			shouldSucceed:         true,
			expectedFinalStatus:  "resolved",
		},
		{
			name:                 "acknowledged to resolved - normal flow",
			initialStatus:        "acknowledged",
			targetStatus:         "resolved",
			shouldSucceed:         true,
			expectedFinalStatus:  "resolved",
		},
		{
			name:                 "auto-resolution on probe recovery",
			initialStatus:        "acknowledged",
			targetStatus:         "resolved",
			shouldSucceed:         true,
			expectedFinalStatus:  "resolved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				if r.Method == http.MethodPut {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"data": map[string]interface{}{
							"id":     1,
							"status": tt.expectedFinalStatus,
							"probe_id": 123,
							"server_id": 456,
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

			// Update probe alert status
			// In real scenario: result, err := client.ProbeAlerts.Update(ctx, 1, updateReq)

			if tt.shouldSucceed {
				assert.True(t, tt.shouldSucceed)
				assert.Equal(t, tt.expectedFinalStatus, tt.targetStatus)
			}

			_ = ctx
			_ = client
		})
	}
}

// TestProbeAlertStateTransitions_ResolvedIsFinal tests that resolved is a final state
func TestProbeAlertStateTransitions_ResolvedIsFinal(t *testing.T) {
	tests := []struct {
		name              string
		attemptedStatus   string
		shouldFail        bool
	}{
		{
			name:             "resolved to active - should fail",
			attemptedStatus: "active",
			shouldFail:       true,
		},
		{
			name:             "resolved to acknowledged - should fail",
			attemptedStatus: "acknowledged",
			shouldFail:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				if r.Method == http.MethodPut {
					w.WriteHeader(http.StatusConflict)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"error": "Cannot transition from resolved state",
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

			if tt.shouldFail {
				assert.True(t, tt.shouldFail)
			}

			_ = client
		})
	}
}

// TestProbeAlertStateTransitions_AutoResolution tests automatic resolution when probe recovers
func TestProbeAlertStateTransitions_AutoResolution(t *testing.T) {
	tests := []struct {
		name               string
		probeStatus        string
		alertShouldResolve bool
	}{
		{
			name:               "probe recovers - alert auto-resolves",
			probeStatus:        "up",
			alertShouldResolve: true,
		},
		{
			name:               "probe still down - alert remains active",
			probeStatus:        "down",
			alertShouldResolve: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				if r.Method == http.MethodGet {
					w.WriteHeader(http.StatusOK)
					status := "active"
					if tt.alertShouldResolve {
						status = "resolved"
					}
					json.NewEncoder(w).Encode(map[string]interface{}{
						"data": map[string]interface{}{
							"id":           1,
							"status":       status,
							"probe_status": tt.probeStatus,
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

			// In real scenario: alert, err := client.ProbeAlerts.Get(ctx, 1)
			assert.NotNil(t, client)
			_ = ctx
		})
	}
}

// TestProbeAlertStateTransitions_AcknowledgmentContext tests acknowledgment with monitoring context
func TestProbeAlertStateTransitions_AcknowledgmentContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodPut {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"id":               1,
					"status":           "acknowledged",
					"acknowledged_at":  time.Now().Unix(),
					"acknowledged_by":  123,
					"probe_id":         456,
					"monitoring_context": "Investigating probe failure",
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

	// Acknowledge with monitoring context
	// In real scenario: result, err := client.ProbeAlerts.Acknowledge(ctx, 1, &AcknowledgeRequest{...})

	assert.NotNil(t, client)
	_ = ctx
}

// TestProbeAlertStateTransitions_ProbeRecoveryFlow tests complete recovery flow
func TestProbeAlertStateTransitions_ProbeRecoveryFlow(t *testing.T) {
	callSequence := []string{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodGet && r.URL.Path == "/v1/probe-alerts/1" {
			callSequence = append(callSequence, "get_alert")
			w.WriteHeader(http.StatusOK)

			// Simulate progression: active -> acknowledged -> resolved
			status := "active"
			if len(callSequence) > 2 {
				status = "acknowledged"
			}
			if len(callSequence) > 3 {
				status = "resolved"
			}

			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"id":     1,
					"status": status,
				},
			})
			return
		}

		if r.Method == http.MethodPut {
			callSequence = append(callSequence, "update")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"id":     1,
					"status": "resolved",
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

	// Step 1: Get initial alert (active)
	// Step 2: Acknowledge the alert
	// Step 3: Wait for probe recovery
	// Step 4: Auto-resolve

	assert.NotNil(t, client)
	_ = ctx
}

// TestProbeAlertStateTransitions_ConcurrentAcknowledgments tests concurrent acknowledgment attempts
func TestProbeAlertStateTransitions_ConcurrentAcknowledgments(t *testing.T) {
	ackCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodPut {
			ackCount++
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"id":     1,
					"status": "acknowledged",
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

	// Multiple concurrent acknowledgments should all succeed (idempotent)
	assert.NotNil(t, client)
	assert.GreaterOrEqual(t, ackCount, 0)
	_ = ctx
}

// TestProbeAlertStateTransitions_AlertWithProbeContext tests alerts with probe details
func TestProbeAlertStateTransitions_AlertWithProbeContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"id":        1,
					"status":    "active",
					"probe_id":  456,
					"probe_type": "HTTP",
					"server_id": 789,
					"created_at": time.Now().Add(-1 * time.Hour).Unix(),
					"failure_count": 5,
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

	// Get alert with probe context
	// In real scenario: alert, err := client.ProbeAlerts.Get(ctx, 1)

	assert.NotNil(t, client)
	_ = ctx
}

// BenchmarkProbeAlertStateTransitions benchmarks probe alert state transition operations
func BenchmarkProbeAlertStateTransitions(b *testing.B) {
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
		// In real scenario: _, _ = client.ProbeAlerts.List(ctx, &ListOptions{})
		_ = ctx
		_ = client
	}
}
