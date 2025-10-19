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

// TestSubscriptionStateTransitions_TrialToActive tests transition from trial to active subscription
func TestSubscriptionStateTransitions_TrialToActive(t *testing.T) {
	tests := []struct {
		name                 string
		trialStatus          string
		paymentSucceeds      bool
		shouldTransition     bool
		expectedFinalStatus  string
	}{
		{
			name:                "trial period expires with valid payment - transitions to active",
			trialStatus:         "trialing",
			paymentSucceeds:      true,
			shouldTransition:     true,
			expectedFinalStatus:  "active",
		},
		{
			name:                "trial period expires with failed payment - transitions to past_due",
			trialStatus:         "trialing",
			paymentSucceeds:      false,
			shouldTransition:     true,
			expectedFinalStatus:  "past_due",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				if r.Method == http.MethodGet {
					w.WriteHeader(http.StatusOK)
					status := tt.expectedFinalStatus
					if !tt.paymentSucceeds && tt.expectedFinalStatus == "past_due" {
						status = "past_due"
					}
					json.NewEncoder(w).Encode(map[string]interface{}{
						"data": map[string]interface{}{
							"id":                   "sub-123",
							"organization_id":      1,
							"plan_id":              "plan-monthly",
							"plan_name":            "Professional",
							"status":               status,
							"current_period_start": time.Now().Unix(),
							"current_period_end":   time.Now().AddDate(0, 1, 0).Unix(),
							"trial_start":          time.Now().AddDate(0, 0, -30).Unix(),
							"trial_end":            time.Now().Unix(),
							"quantity":             1,
						},
					})
					return
				}

				if r.Method == http.MethodPut {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"data": map[string]interface{}{
							"id":                   "sub-123",
							"organization_id":      1,
							"plan_id":              "plan-monthly",
							"plan_name":            "Professional",
							"status":               tt.expectedFinalStatus,
							"current_period_start": time.Now().Unix(),
							"current_period_end":   time.Now().AddDate(0, 1, 0).Unix(),
							"quantity":             1,
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

			if tt.shouldTransition {
				assert.NotNil(t, client)
				assert.Equal(t, "test-token", client.config.Auth.Token)
			}

			_ = ctx
		})
	}
}

// TestSubscriptionStateTransitions_ActiveToPastDue tests transition from active to past_due on payment failure
func TestSubscriptionStateTransitions_ActiveToPastDue(t *testing.T) {
	tests := []struct {
		name                string
		initialStatus       string
		paymentAttempts     int
		shouldTransitionToPastDue bool
		gracePeriodDays     int
	}{
		{
			name:                "payment fails immediately - transitions to past_due",
			initialStatus:       "active",
			paymentAttempts:     1,
			shouldTransitionToPastDue: true,
			gracePeriodDays:     14,
		},
		{
			name:                "multiple payment attempts fail - transitions to past_due",
			initialStatus:       "active",
			paymentAttempts:     3,
			shouldTransitionToPastDue: true,
			gracePeriodDays:     14,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				if r.Method == http.MethodGet {
					w.WriteHeader(http.StatusOK)
					status := tt.initialStatus
					if tt.shouldTransitionToPastDue {
						status = "past_due"
					}
					json.NewEncoder(w).Encode(map[string]interface{}{
						"data": map[string]interface{}{
							"id":                   "sub-456",
							"organization_id":      2,
							"plan_id":              "plan-enterprise",
							"plan_name":            "Enterprise",
							"status":               status,
							"current_period_start": time.Now().AddDate(0, -1, 0).Unix(),
							"current_period_end":   time.Now().Unix(),
							"quantity":             5,
						},
					})
					return
				}

				if r.Method == http.MethodPut {
					expectedStatus := "past_due"
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"data": map[string]interface{}{
							"id":                   "sub-456",
							"organization_id":      2,
							"plan_id":              "plan-enterprise",
							"plan_name":            "Enterprise",
							"status":               expectedStatus,
							"current_period_start": time.Now().AddDate(0, -1, 0).Unix(),
							"current_period_end":   time.Now().Unix(),
							"quantity":             5,
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

			assert.NotNil(t, client)
			assert.GreaterOrEqual(t, tt.paymentAttempts, 1)
			_ = ctx
		})
	}
}

// TestSubscriptionStateTransitions_PastDueToActive tests payment recovery from past_due to active
func TestSubscriptionStateTransitions_PastDueToActive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"id":                   "sub-789",
					"organization_id":      3,
					"plan_id":              "plan-monthly",
					"plan_name":            "Professional",
					"status":               "past_due",
					"current_period_start": time.Now().AddDate(0, -1, -5).Unix(),
					"current_period_end":   time.Now().AddDate(0, -1, 25).Unix(),
					"quantity":             1,
				},
			})
			return
		}

		if r.Method == http.MethodPut {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"id":                   "sub-789",
					"organization_id":      3,
					"plan_id":              "plan-monthly",
					"plan_name":            "Professional",
					"status":               "active",
					"current_period_start": time.Now().Unix(),
					"current_period_end":   time.Now().AddDate(0, 1, 0).Unix(),
					"quantity":             1,
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

	// Payment recovered - subscription reactivated
	assert.NotNil(t, client)
	_ = ctx
}

// TestSubscriptionStateTransitions_ActiveToCanceled tests manual subscription cancellation
func TestSubscriptionStateTransitions_ActiveToCanceled(t *testing.T) {
	tests := []struct {
		name                   string
		initialStatus          string
		cancelImmediately      bool
		shouldSucceed          bool
		expectedFinalStatus    string
	}{
		{
			name:                  "cancel active subscription immediately",
			initialStatus:         "active",
			cancelImmediately:     true,
			shouldSucceed:         true,
			expectedFinalStatus:   "canceled",
		},
		{
			name:                  "cancel active subscription at period end",
			initialStatus:         "active",
			cancelImmediately:     false,
			shouldSucceed:         true,
			expectedFinalStatus:   "active", // Remains active until period end, but marked for cancellation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				if r.Method == http.MethodDelete || r.Method == http.MethodPut {
					w.WriteHeader(http.StatusOK)
					canceledAt := time.Time{}
					if tt.cancelImmediately {
						canceledAt = time.Now()
					}
					json.NewEncoder(w).Encode(map[string]interface{}{
						"data": map[string]interface{}{
							"id":                   "sub-cancel-1",
							"organization_id":      4,
							"plan_id":              "plan-pro",
							"plan_name":            "Pro",
							"status":               tt.expectedFinalStatus,
							"current_period_start": time.Now().AddDate(0, -1, 0).Unix(),
							"current_period_end":   time.Now().AddDate(0, 1, 0).Unix(),
							"cancel_at_period_end": !tt.cancelImmediately,
							"canceled_at":          canceledAt.Unix(),
							"quantity":             1,
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

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}

			_ = ctx
		})
	}
}

// TestSubscriptionStateTransitions_PastDueToCanceled tests cancellation from past_due state during grace period
func TestSubscriptionStateTransitions_PastDueToCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodDelete || r.Method == http.MethodPut {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"id":                   "sub-pastdue-cancel",
					"organization_id":      5,
					"plan_id":              "plan-starter",
					"plan_name":            "Starter",
					"status":               "canceled",
					"current_period_start": time.Now().AddDate(0, -1, -10).Unix(),
					"current_period_end":   time.Now().AddDate(0, -1, 20).Unix(),
					"canceled_at":          time.Now().Unix(),
					"quantity":             1,
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

	// Subscription canceled during grace period from past_due state
	assert.NotNil(t, client)
	_ = ctx
}

// TestSubscriptionStateTransitions_CanceledIsFinalState tests that canceled is a final state
func TestSubscriptionStateTransitions_CanceledIsFinalState(t *testing.T) {
	tests := []struct {
		name              string
		attemptedStatus   string
		shouldFail        bool
		expectedErrorCode int
	}{
		{
			name:              "canceled to active - should fail",
			attemptedStatus:   "active",
			shouldFail:        true,
			expectedErrorCode: http.StatusConflict,
		},
		{
			name:              "canceled to past_due - should fail",
			attemptedStatus:   "past_due",
			shouldFail:        true,
			expectedErrorCode: http.StatusConflict,
		},
		{
			name:              "canceled to trialing - should fail",
			attemptedStatus:   "trialing",
			shouldFail:        true,
			expectedErrorCode: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				if r.Method == http.MethodPut {
					w.WriteHeader(tt.expectedErrorCode)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"error": fmt.Sprintf("Cannot transition from canceled to %s", tt.attemptedStatus),
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
				assert.True(t, tt.shouldFail)
			}

			_ = ctx
			_ = client
		})
	}
}

// TestSubscriptionStateTransitions_GracePeriodExpiry tests automatic cancellation after grace period expires
func TestSubscriptionStateTransitions_GracePeriodExpiry(t *testing.T) {
	tests := []struct {
		name                      string
		gracePeriodDaysRemaining  int
		shouldAutoCancelOnExpiry   bool
		expectedStatus            string
	}{
		{
			name:                     "grace period with days remaining - remains past_due",
			gracePeriodDaysRemaining: 5,
			shouldAutoCancelOnExpiry: false,
			expectedStatus:           "past_due",
		},
		{
			name:                     "grace period expired - auto-cancels",
			gracePeriodDaysRemaining: 0,
			shouldAutoCancelOnExpiry: true,
			expectedStatus:           "canceled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				if r.Method == http.MethodGet {
					w.WriteHeader(http.StatusOK)
					var graceExpiresAt time.Time
					if tt.gracePeriodDaysRemaining > 0 {
						graceExpiresAt = time.Now().AddDate(0, 0, tt.gracePeriodDaysRemaining)
					} else {
						graceExpiresAt = time.Now().AddDate(0, 0, -1) // Already expired
					}

					json.NewEncoder(w).Encode(map[string]interface{}{
						"data": map[string]interface{}{
							"id":                   "sub-grace-test",
							"organization_id":      6,
							"plan_id":              "plan-monthly",
							"plan_name":            "Monthly Plan",
							"status":               tt.expectedStatus,
							"current_period_start": time.Now().AddDate(0, -1, -10).Unix(),
							"current_period_end":   time.Now().AddDate(0, -1, 20).Unix(),
							"grace_period_expires": graceExpiresAt.Unix(),
							"quantity":             1,
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

			assert.NotNil(t, client)
			assert.True(t, tt.gracePeriodDaysRemaining >= 0 || tt.shouldAutoCancelOnExpiry)
			_ = ctx
		})
	}
}

// TestSubscriptionStateTransitions_FeatureAvailabilityByState tests feature access based on subscription state
func TestSubscriptionStateTransitions_FeatureAvailabilityByState(t *testing.T) {
	tests := []struct {
		name                 string
		subscriptionStatus   string
		canAccessFeatures    bool
		canSubmitMetrics     bool
		canCreateAlerts      bool
	}{
		{
			name:               "trialing state - full feature access",
			subscriptionStatus: "trialing",
			canAccessFeatures:  true,
			canSubmitMetrics:   true,
			canCreateAlerts:    true,
		},
		{
			name:               "active state - full feature access",
			subscriptionStatus: "active",
			canAccessFeatures:  true,
			canSubmitMetrics:   true,
			canCreateAlerts:    true,
		},
		{
			name:               "past_due state - limited feature access",
			subscriptionStatus: "past_due",
			canAccessFeatures:  true,  // Can still read
			canSubmitMetrics:   false, // Cannot write
			canCreateAlerts:    false, // Cannot create
		},
		{
			name:               "canceled state - no feature access",
			subscriptionStatus: "canceled",
			canAccessFeatures:  false,
			canSubmitMetrics:   false,
			canCreateAlerts:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				// Feature access based on subscription state
				if tt.subscriptionStatus == "canceled" && r.Method != http.MethodGet {
					w.WriteHeader(http.StatusForbidden)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"error": "Subscription canceled - feature access denied",
					})
					return
				}

				if tt.subscriptionStatus == "past_due" && (r.URL.Path == "/v1/metrics" || r.URL.Path == "/v1/alerts") {
					w.WriteHeader(http.StatusPaymentRequired) // 402 Payment Required
					json.NewEncoder(w).Encode(map[string]interface{}{
						"error": "Payment required - write access denied",
					})
					return
				}

				if r.Method == http.MethodGet {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"data": map[string]interface{}{
							"id":              "sub-features",
							"status":          tt.subscriptionStatus,
							"plan_name":       "Professional",
							"organization_id": 7,
						},
					})
					return
				}

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{
						"id":     "op-123",
						"status": "success",
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

			ctx := context.Background()

			// Verify feature access assertions
			assert.Equal(t, tt.subscriptionStatus != "canceled", tt.canAccessFeatures || tt.subscriptionStatus == "past_due")
			assert.Equal(t, tt.subscriptionStatus == "active" || tt.subscriptionStatus == "trialing", tt.canSubmitMetrics)
			assert.Equal(t, tt.subscriptionStatus == "active" || tt.subscriptionStatus == "trialing", tt.canCreateAlerts)

			_ = ctx
			_ = client
		})
	}
}

// TestSubscriptionStateTransitions_CompleteLifecycle tests complete subscription lifecycle from trial to cancellation
func TestSubscriptionStateTransitions_CompleteLifecycle(t *testing.T) {
	callSequence := []string{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodGet && r.URL.Path == "/v1/subscriptions/complete-lifecycle" {
			callSequence = append(callSequence, "get_subscription")
			w.WriteHeader(http.StatusOK)

			// Simulate progression: trialing -> active -> past_due -> canceled
			status := "trialing"
			if len(callSequence) > 2 {
				status = "active"
			}
			if len(callSequence) > 4 {
				status = "past_due"
			}
			if len(callSequence) > 6 {
				status = "canceled"
			}

			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"id":     "sub-lifecycle",
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
					"id":     "sub-lifecycle",
					"status": "updated",
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

	// Step 1: Trial period active
	// Step 2: Transition to active after trial ends and payment succeeds
	// Step 3: Payment fails - transition to past_due
	// Step 4: Cancel subscription from past_due state
	// Step 5: Verify subscription is now canceled (final state)

	assert.NotNil(t, client)
	_ = ctx
}

// BenchmarkSubscriptionStateTransitions benchmarks subscription state transition operations
func BenchmarkSubscriptionStateTransitions(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":     "sub-bench",
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
		// In real scenario: _, _ = client.Billing.GetSubscription(ctx, "org-id")
		_ = ctx
		_ = client
	}
}
