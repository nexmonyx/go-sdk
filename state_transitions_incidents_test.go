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

// TestIncidentStateTransitions_ValidTransitions tests valid state transitions for incidents
func TestIncidentStateTransitions_ValidTransitions(t *testing.T) {
	tests := []struct {
		name              string
		initialStatus     IncidentStatus
		targetStatus      IncidentStatus
		updateRequest     UpdateIncidentRequest
		shouldSucceed      bool
		expectedFinalStatus IncidentStatus
	}{
		{
			name:              "active to acknowledged - valid transition",
			initialStatus:     IncidentStatusActive,
			targetStatus:      IncidentStatusAcknowledged,
			updateRequest:     UpdateIncidentRequest{Status: IncidentStatusAcknowledged, Notes: "Acknowledged by team"},
			shouldSucceed:      true,
			expectedFinalStatus: IncidentStatusAcknowledged,
		},
		{
			name:              "active to resolved - valid direct resolution",
			initialStatus:     IncidentStatusActive,
			targetStatus:      IncidentStatusResolved,
			updateRequest:     UpdateIncidentRequest{Status: IncidentStatusResolved, Notes: "Issue resolved"},
			shouldSucceed:      true,
			expectedFinalStatus: IncidentStatusResolved,
		},
		{
			name:              "acknowledged to resolved - valid transition",
			initialStatus:     IncidentStatusAcknowledged,
			targetStatus:      IncidentStatusResolved,
			updateRequest:     UpdateIncidentRequest{Status: IncidentStatusResolved, Notes: "Fix deployed"},
			shouldSucceed:      true,
			expectedFinalStatus: IncidentStatusResolved,
		},
		{
			name:              "acknowledged to active - escalation from acknowledged",
			initialStatus:     IncidentStatusAcknowledged,
			targetStatus:      IncidentStatusActive,
			updateRequest:     UpdateIncidentRequest{Status: IncidentStatusActive, Notes: "Issue escalated"},
			shouldSucceed:      true,
			expectedFinalStatus: IncidentStatusActive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				// First request: Get existing incident with initialStatus
				if r.Method == http.MethodGet {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"data": map[string]interface{}{
							"id":     1,
							"status": string(tt.initialStatus),
							"title":  "Test Incident",
						},
					})
					return
				}

				// Update request: Return updated incident with targetStatus
				if r.Method == http.MethodPut {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"data": map[string]interface{}{
							"id":                1,
							"status":            string(tt.expectedFinalStatus),
							"title":             "Test Incident",
							"acknowledged_at":   time.Now().Unix(),
							"acknowledged_by":   1,
							"resolved_at":       time.Now().Unix(),
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

			// Update the incident
			result, err := client.Incidents.Update(ctx, 1, &tt.updateRequest)

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, string(tt.expectedFinalStatus), result.Status)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// TestIncidentStateTransitions_InvalidTransitions tests invalid state transitions that should be rejected
func TestIncidentStateTransitions_InvalidTransitions(t *testing.T) {
	tests := []struct {
		name              string
		initialStatus     IncidentStatus
		attemptedStatus   IncidentStatus
		shouldFail        bool
		expectedErrorCode int
	}{
		{
			name:              "resolved to active - should fail",
			initialStatus:     IncidentStatusResolved,
			attemptedStatus:   IncidentStatusActive,
			shouldFail:        true,
			expectedErrorCode: http.StatusConflict, // 409 Conflict
		},
		{
			name:              "resolved to acknowledged - should fail",
			initialStatus:     IncidentStatusResolved,
			attemptedStatus:   IncidentStatusAcknowledged,
			shouldFail:        true,
			expectedErrorCode: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				if r.Method == http.MethodPut {
					// Simulate invalid state transition rejection
					w.WriteHeader(tt.expectedErrorCode)
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

			// Attempt invalid transition
			updateReq := &UpdateIncidentRequest{
				Status: tt.attemptedStatus,
				Notes:  "Invalid transition attempt",
			}

			result, err := client.Incidents.Update(ctx, 1, updateReq)

			if tt.shouldFail {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

// TestIncidentStateTransitions_IdempotentUpdates tests idempotent updates to the same state
func TestIncidentStateTransitions_IdempotentUpdates(t *testing.T) {
	tests := []struct {
		name            string
		status          IncidentStatus
		description     string
	}{
		{
			name:            "acknowledge active incident twice",
			status:          IncidentStatusAcknowledged,
			description:     "Acknowledging already acknowledged incident should not fail",
		},
		{
			name:            "resolve resolved incident twice",
			status:          IncidentStatusResolved,
			description:     "Resolving already resolved incident should not fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				if r.Method == http.MethodPut {
					updateCount++
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"data": map[string]interface{}{
							"id":     1,
							"status": string(tt.status),
							"title":  "Test Incident",
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
			updateReq := &UpdateIncidentRequest{
				Status: tt.status,
				Notes:  "Idempotent update",
			}

			// First update
			result1, err := client.Incidents.Update(ctx, 1, updateReq)
			require.NoError(t, err)
			require.NotNil(t, result1)
			assert.Equal(t, string(tt.status), result1.Status)

			// Second identical update (should succeed idempotently)
			result2, err := client.Incidents.Update(ctx, 1, updateReq)
			require.NoError(t, err)
			require.NotNil(t, result2)
			assert.Equal(t, string(tt.status), result2.Status)

			// Both should succeed
			assert.Equal(t, 2, updateCount)
		})
	}
}

// TestIncidentStateTransitions_AcknowledgmentTracking tests that acknowledgment details are tracked
func TestIncidentStateTransitions_AcknowledgmentTracking(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodPut {
			w.WriteHeader(http.StatusOK)
			now := time.Now()
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"id":               1,
					"status":           "acknowledged",
					"title":            "Test Incident",
					"acknowledged_at":  now.Unix(),
					"acknowledged_by":  123,
					"acknowledgment_notes": "Team investigating",
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
	updateReq := &UpdateIncidentRequest{
		Status: IncidentStatusAcknowledged,
		Notes:  "Team investigating",
	}

	result, err := client.Incidents.Update(ctx, 1, updateReq)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify acknowledgment details are tracked
	assert.Equal(t, "acknowledged", result.Status)
	assert.NotNil(t, result.AcknowledgedAt)
	assert.NotNil(t, result.AcknowledgedBy)
}

// TestIncidentStateTransitions_ResolutionDetails tests resolution state tracking
func TestIncidentStateTransitions_ResolutionDetails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodPut {
			w.WriteHeader(http.StatusOK)
			now := time.Now()
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"id":              1,
					"status":          "resolved",
					"title":           "Test Incident",
					"resolved_at":     now.Unix(),
					"resolution_time": 3600, // 1 hour
					"root_cause":      "Configuration error fixed",
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
	updateReq := &UpdateIncidentRequest{
		Status: IncidentStatusResolved,
		Notes:  "Configuration error fixed",
	}

	result, err := client.Incidents.Update(ctx, 1, updateReq)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify resolution details are tracked
	assert.Equal(t, "resolved", result.Status)
	assert.NotNil(t, result.ResolvedAt)
}

// TestIncidentStateTransitions_WithContext tests state transitions with timeout contexts
func TestIncidentStateTransitions_WithContext(t *testing.T) {
	tests := []struct {
		name            string
		timeout         time.Duration
		shouldTimeout   bool
	}{
		{
			name:            "with sufficient timeout",
			timeout:         5 * time.Second,
			shouldTimeout:   false,
		},
		{
			name:            "with minimal timeout",
			timeout:         100 * time.Millisecond,
			shouldTimeout:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Simulate potential delay
				if tt.shouldTimeout {
					time.Sleep(200 * time.Millisecond)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{
						"id":     1,
						"status": "acknowledged",
						"title":  "Test Incident",
					},
				})
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL:    server.URL,
				Auth:       AuthConfig{Token: "test-token"},
				RetryCount: 0,
			})
			require.NoError(t, err)

			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			updateReq := &UpdateIncidentRequest{
				Status: IncidentStatusAcknowledged,
				Notes:  "Context timeout test",
			}

			result, err := client.Incidents.Update(ctx, 1, updateReq)

			if tt.shouldTimeout {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

// BenchmarkIncidentStateTransitions benchmarks incident state transition operations
func BenchmarkIncidentStateTransitions(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":     1,
				"status": "acknowledged",
				"title":  "Test Incident",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{
		BaseURL:    server.URL,
		Auth:       AuthConfig{Token: "test-token"},
		RetryCount: 0,
	})

	updateReq := &UpdateIncidentRequest{
		Status: IncidentStatusAcknowledged,
		Notes:  "Benchmark update",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		_, _ = client.Incidents.Update(ctx, 1, updateReq)
	}
}
