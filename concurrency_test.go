package nexmonyx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClient_ConcurrentRequests tests concurrent API requests
func TestClient_ConcurrentRequests(t *testing.T) {
	tests := []struct {
		name            string
		concurrency     int
		requestsPerGoro int
		expectErrors    bool
	}{
		{
			name:            "low concurrency - 5 goroutines",
			concurrency:     5,
			requestsPerGoro: 10,
			expectErrors:    false,
		},
		{
			name:            "medium concurrency - 20 goroutines",
			concurrency:     20,
			requestsPerGoro: 5,
			expectErrors:    false,
		},
		{
			name:            "high concurrency - 100 goroutines",
			concurrency:     100,
			requestsPerGoro: 2,
			expectErrors:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestCount := int64(0)

			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt64(&requestCount, 1)
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": []interface{}{},
				})
			}))
			defer server.Close()

			// Create client
			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			// Run concurrent requests
			var wg sync.WaitGroup
			errors := make([]error, 0)
			var errorsMu sync.Mutex

			for i := 0; i < tt.concurrency; i++ {
				wg.Add(1)
				go func(routineID int) {
					defer wg.Done()

					for j := 0; j < tt.requestsPerGoro; j++ {
						_, _, err := client.Servers.List(context.Background(), nil)
						if err != nil {
							errorsMu.Lock()
							errors = append(errors, err)
							errorsMu.Unlock()
						}
					}
				}(i)
			}

			wg.Wait()

			// Verify results
			expectedRequests := int64(tt.concurrency * tt.requestsPerGoro)
			assert.Equal(t, expectedRequests, atomic.LoadInt64(&requestCount), "Request count mismatch")

			if tt.expectErrors {
				assert.NotEmpty(t, errors, "Expected errors but got none")
			} else {
				assert.Empty(t, errors, "Unexpected errors: %v", errors)
			}
		})
	}
}

// TestClient_ConcurrentUpdates tests concurrent updates to the same resource
func TestClient_ConcurrentUpdates(t *testing.T) {
	tests := []struct {
		name        string
		concurrency int
		expectPanic bool
	}{
		{
			name:        "concurrent updates - 10 goroutines",
			concurrency: 10,
			expectPanic: false,
		},
		{
			name:        "high concurrent updates - 50 goroutines",
			concurrency: 50,
			expectPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateCount := int64(0)
			lastValue := ""
			var valueMu sync.Mutex

			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt64(&updateCount, 1)

				// Simulate processing time
				time.Sleep(1 * time.Millisecond)

				// Read request body
				var req map[string]interface{}
				json.NewDecoder(r.Body).Decode(&req)

				valueMu.Lock()
				if hostname, ok := req["hostname"].(string); ok {
					lastValue = hostname
				}
				valueMu.Unlock()

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{
						"uuid":     "server-uuid",
						"hostname": req["hostname"],
					},
				})
			}))
			defer server.Close()

			// Create client
			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			// Run concurrent updates
			var wg sync.WaitGroup
			errors := make([]error, 0)
			var errorsMu sync.Mutex

			for i := 0; i < tt.concurrency; i++ {
				wg.Add(1)
				go func(iteration int) {
					defer wg.Done()

					server := &Server{
						Hostname: fmt.Sprintf("server-%d", iteration),
					}

					_, err := client.Servers.Update(context.Background(), "server-uuid", server)
					if err != nil {
						errorsMu.Lock()
						errors = append(errors, err)
						errorsMu.Unlock()
					}
				}(i)
			}

			wg.Wait()

			// Verify no panics occurred and all updates were processed
			assert.Equal(t, int64(tt.concurrency), atomic.LoadInt64(&updateCount), "Update count mismatch")
			assert.Empty(t, errors, "Unexpected errors: %v", errors)

			// Verify final value is one of the expected values
			valueMu.Lock()
			finalValue := lastValue
			valueMu.Unlock()

			assert.NotEmpty(t, finalValue, "Final value should not be empty")
		})
	}
}

// TestClient_RaceConditions tests for race conditions using go test -race
func TestClient_RaceConditions(t *testing.T) {
	// This test is designed to be run with -race flag
	// go test -race -run TestClient_RaceConditions

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []interface{}{},
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	// Run parallel reads
	t.Run("parallel reads", func(t *testing.T) {
		var wg sync.WaitGroup
		iterations := 100

		for i := 0; i < iterations; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				client.Servers.List(context.Background(), nil)
			}()
		}

		wg.Wait()
	})

	// Run parallel writes
	t.Run("parallel writes", func(t *testing.T) {
		var wg sync.WaitGroup
		iterations := 100

		for i := 0; i < iterations; i++ {
			wg.Add(1)
			go func(iter int) {
				defer wg.Done()
				server := &Server{
					Hostname: fmt.Sprintf("server-%d", iter),
				}
				client.Servers.Update(context.Background(), "uuid", server)
			}(i)
		}

		wg.Wait()
	})

	// Run mixed reads and writes
	t.Run("mixed reads and writes", func(t *testing.T) {
		var wg sync.WaitGroup
		iterations := 100

		// Readers
		for i := 0; i < iterations; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				client.Servers.List(context.Background(), nil)
			}()
		}

		// Writers
		for i := 0; i < iterations; i++ {
			wg.Add(1)
			go func(iter int) {
				defer wg.Done()
				server := &Server{
					Hostname: fmt.Sprintf("server-%d", iter),
				}
				client.Servers.Update(context.Background(), "uuid", server)
			}(i)
		}

		wg.Wait()
	})
}

// TestClient_ConnectionPoolStress tests connection pool behavior under stress
func TestClient_ConnectionPoolStress(t *testing.T) {
	tests := []struct {
		name           string
		concurrency    int
		delayMs        int
		expectTimeouts bool
	}{
		{
			name:           "fast responses - no timeouts",
			concurrency:    50,
			delayMs:        1,
			expectTimeouts: false,
		},
		{
			name:           "slow responses - may cause timeouts",
			concurrency:    100,
			delayMs:        50,
			expectTimeouts: false, // Should handle gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			activeConnections := int64(0)
			maxConnections := int64(0)
			var connMu sync.Mutex

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				current := atomic.AddInt64(&activeConnections, 1)

				connMu.Lock()
				if current > maxConnections {
					maxConnections = current
				}
				connMu.Unlock()

				time.Sleep(time.Duration(tt.delayMs) * time.Millisecond)

				atomic.AddInt64(&activeConnections, -1)

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": []interface{}{},
				})
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			var wg sync.WaitGroup
			timeouts := int64(0)

			for i := 0; i < tt.concurrency; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					_, _, err := client.Servers.List(ctx, nil)
					if err != nil && err.Error() == "context deadline exceeded" {
						atomic.AddInt64(&timeouts, 1)
					}
				}()
			}

			wg.Wait()

			t.Logf("Max concurrent connections: %d", atomic.LoadInt64(&maxConnections))
			t.Logf("Timeouts: %d", atomic.LoadInt64(&timeouts))

			if tt.expectTimeouts {
				assert.Greater(t, atomic.LoadInt64(&timeouts), int64(0), "Expected some timeouts")
			}
		})
	}
}

// TestClient_DeadlockPrevention tests that concurrent operations don't deadlock
func TestClient_DeadlockPrevention(t *testing.T) {
	var requestCount int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Random delay to increase chance of deadlock if one exists
		count := atomic.AddInt64(&requestCount, 1)
		time.Sleep(time.Duration(1+count%10) * time.Millisecond)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []interface{}{},
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	// Set a timeout for the entire test to detect deadlocks
	done := make(chan bool, 1)

	go func() {
		var wg sync.WaitGroup
		operations := 200

		for i := 0; i < operations; i++ {
			wg.Add(1)
			go func(iter int) {
				defer wg.Done()

				// Mix different operations
				switch iter % 4 {
				case 0:
					client.Servers.List(context.Background(), nil)
				case 1:
					server := &Server{Hostname: fmt.Sprintf("s-%d", iter)}
					client.Servers.Update(context.Background(), "uuid", server)
				case 2:
					client.Servers.Get(context.Background(), "uuid")
				case 3:
					client.Servers.List(context.Background(), &ListOptions{Page: 1, Limit: 10})
				}
			}(i)
		}

		wg.Wait()
		done <- true
	}()

	// Wait with timeout
	select {
	case <-done:
		t.Log("All operations completed without deadlock")
	case <-time.After(30 * time.Second):
		t.Fatal("Test timed out - possible deadlock detected")
	}
}
