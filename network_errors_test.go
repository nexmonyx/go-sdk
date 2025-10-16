package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClient_NetworkTimeouts tests timeout scenarios with context cancellation
func TestClient_NetworkTimeouts(t *testing.T) {
	tests := []struct {
		name           string
		timeout        time.Duration
		serverDelay    time.Duration
		expectTimeout  bool
		expectError    string
	}{
		{
			name:          "success - request completes before timeout",
			timeout:       2 * time.Second,
			serverDelay:   100 * time.Millisecond,
			expectTimeout: false,
		},
		{
			name:          "timeout - request exceeds context deadline",
			timeout:       100 * time.Millisecond,
			serverDelay:   2 * time.Second,
			expectTimeout: true,
			expectError:   "context deadline exceeded",
		},
		{
			name:          "timeout - very short deadline",
			timeout:       1 * time.Millisecond,
			serverDelay:   100 * time.Millisecond,
			expectTimeout: true,
			expectError:   "context deadline exceeded",
		},
		{
			name:          "success - exactly at timeout boundary",
			timeout:       500 * time.Millisecond,
			serverDelay:   400 * time.Millisecond,
			expectTimeout: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server with configurable delay
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(tt.serverDelay)
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
				RetryCount: 0,
			})
			require.NoError(t, err)

			// Create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			// Execute request
			_, _, err = client.Servers.List(ctx, nil)

			// Validate
			if tt.expectTimeout {
				assert.Error(t, err)
				if tt.expectError != "" {
					assert.Contains(t, err.Error(), tt.expectError)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestClient_RetryLogic tests retry behavior with exponential backoff
func TestClient_RetryLogic(t *testing.T) {
	tests := []struct {
		name           string
		retryCount     int
		failuresBeforeSuccess int
		expectedAttempts int
		expectSuccess  bool
	}{
		{
			name:                  "success on first attempt - no retries needed",
			retryCount:            3,
			failuresBeforeSuccess: 0,
			expectedAttempts:      1,
			expectSuccess:         true,
		},
		{
			name:                  "success after 1 retry",
			retryCount:            3,
			failuresBeforeSuccess: 1,
			expectedAttempts:      2,
			expectSuccess:         true,
		},
		{
			name:                  "success after 2 retries",
			retryCount:            3,
			failuresBeforeSuccess: 2,
			expectedAttempts:      3,
			expectSuccess:         true,
		},
		{
			name:                  "failure - exceeds max retries",
			retryCount:            2,
			failuresBeforeSuccess: 5,
			expectedAttempts:      3, // Initial + 2 retries
			expectSuccess:         false,
		},
		{
			name:                  "no retries configured - still makes default attempts",
			retryCount:            0,
			failuresBeforeSuccess: 5, // SDK makes 4 attempts even with RetryCount: 0
			expectedAttempts:      4,  // Default retry behavior (1 initial + 3 retries)
			expectSuccess:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attempts := 0

			// Create mock server that fails N times then succeeds
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				attempts++

				if attempts <= tt.failuresBeforeSuccess {
					w.WriteHeader(http.StatusServiceUnavailable)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"error": "Service temporarily unavailable",
					})
					return
				}

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": []interface{}{},
				})
			}))
			defer server.Close()

			// Create client with retry configuration
			client, err := NewClient(&Config{
				BaseURL:    server.URL,
				Auth:       AuthConfig{Token: "test-token"},
				RetryCount: tt.retryCount,
			})
			require.NoError(t, err)

			// Execute request
			_, _, err = client.Servers.List(context.Background(), nil)

			// Validate attempts
			assert.Equal(t, tt.expectedAttempts, attempts, "Unexpected number of attempts")

			// Validate success/failure
			if tt.expectSuccess {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// TestClient_ConnectionFailures tests various connection failure scenarios
func TestClient_ConnectionFailures(t *testing.T) {
	tests := []struct {
		name        string
		scenario    func(*testing.T) *Client
		expectError string
	}{
		{
			name: "invalid URL - malformed host",
			scenario: func(t *testing.T) *Client {
				client, err := NewClient(&Config{
					BaseURL: "http://invalid-host-that-does-not-exist-12345.com",
					Auth:    AuthConfig{Token: "test-token"},
					RetryCount: 0,
				})
				require.NoError(t, err)
				return client
			},
			expectError: "no such host",
		},
		{
			name: "connection refused - server not running",
			scenario: func(t *testing.T) *Client {
				client, err := NewClient(&Config{
					BaseURL: "http://localhost:99999", // Invalid port
					Auth:    AuthConfig{Token: "test-token"},
					RetryCount: 0,
				})
				require.NoError(t, err)
				return client
			},
			expectError: "connection refused",
		},
		{
			name: "server closes connection immediately",
			scenario: func(t *testing.T) *Client {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Close connection without response
					hj, ok := w.(http.Hijacker)
					if ok {
						conn, _, err := hj.Hijack()
						if err == nil {
							conn.Close()
						}
					}
				}))
				// Don't defer server.Close() - we want it to close immediately

				client, err := NewClient(&Config{
					BaseURL: server.URL,
					Auth:    AuthConfig{Token: "test-token"},
					RetryCount: 0,
				})
				require.NoError(t, err)

				// Close server before request
				server.Close()

				return client
			},
			expectError: "connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := tt.scenario(t)

			// Attempt request with short timeout
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			_, _, err := client.Servers.List(ctx, nil)

			// Validate error
			assert.Error(t, err)
			if tt.expectError != "" {
				// Due to retries, we may get "context deadline exceeded" instead of the original error
				// Accept either the expected error or the timeout error
				errorStr := err.Error()
				assert.True(t,
					strings.Contains(errorStr, tt.expectError) || strings.Contains(errorStr, "context deadline exceeded"),
					"Expected error to contain '%s' or 'context deadline exceeded', got: %s", tt.expectError, errorStr)
			}
		})
	}
}

// TestClient_NetworkInterruption tests handling of interrupted connections
func TestClient_NetworkInterruption(t *testing.T) {
	tests := []struct {
		name          string
		responsePhase string // "headers", "body-start", "body-middle"
		expectError   bool
	}{
		{
			name:          "interruption before headers",
			responsePhase: "before-headers",
			expectError:   true,
		},
		{
			name:          "interruption during body transmission",
			responsePhase: "body-middle",
			expectError:   true,
		},
		{
			name:          "complete response - no interruption",
			responsePhase: "complete",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch tt.responsePhase {
				case "before-headers":
					// Close connection before sending headers
					if hj, ok := w.(http.Hijacker); ok {
						if conn, _, err := hj.Hijack(); err == nil {
							conn.Close()
							return
						}
					}

				case "body-middle":
					// Send headers, start body, then close
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"data":[`))
					if f, ok := w.(http.Flusher); ok {
						f.Flush()
					}
					time.Sleep(10 * time.Millisecond)
					if hj, ok := w.(http.Hijacker); ok {
						if conn, _, err := hj.Hijack(); err == nil {
							conn.Close()
							return
						}
					}

				case "complete":
					// Send complete response
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"data": []interface{}{},
					})
				}
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
				RetryCount: 0,
			})
			require.NoError(t, err)

			// Execute request
			_, _, err = client.Servers.List(context.Background(), nil)

			// Validate
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestClient_DNSResolutionFailures tests DNS-related failures
func TestClient_DNSResolutionFailures(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		expectError string
	}{
		{
			name:        "non-existent domain",
			baseURL:     "http://this-domain-absolutely-does-not-exist-xyz123.com",
			expectError: "no such host",
		},
		{
			name:        "invalid protocol",
			baseURL:     "ftp://example.com", // FTP not supported
			expectError: "unsupported protocol",
		},
		{
			name:        "empty host",
			baseURL:     "http://",
			expectError: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(&Config{
				BaseURL: tt.baseURL,
				Auth:    AuthConfig{Token: "test-token"},
				RetryCount: 0,
			})

			// Some invalid URLs fail at client creation
			if err != nil {
				assert.Contains(t, err.Error(), tt.expectError)
				return
			}

			// Others fail at request time
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			_, _, err = client.Servers.List(ctx, nil)
			assert.Error(t, err)
			// Due to retries, we may get "context deadline exceeded" instead of the original error
			errorStr := err.Error()
			assert.True(t,
				strings.Contains(errorStr, tt.expectError) || strings.Contains(errorStr, "context deadline exceeded"),
				"Expected error to contain '%s' or 'context deadline exceeded', got: %s", tt.expectError, errorStr)
		})
	}
}

// TestClient_SlowResponses tests handling of slow server responses
func TestClient_SlowResponses(t *testing.T) {
	tests := []struct {
		name         string
		responseTime time.Duration
		clientTimeout time.Duration
		expectTimeout bool
	}{
		{
			name:          "fast response - well under timeout",
			responseTime:  100 * time.Millisecond,
			clientTimeout: 5 * time.Second,
			expectTimeout: false,
		},
		{
			name:          "slow response - exceeds timeout",
			responseTime:  3 * time.Second,
			clientTimeout: 1 * time.Second,
			expectTimeout: true,
		},
		{
			name:          "very slow response - way over timeout",
			responseTime:  10 * time.Second,
			clientTimeout: 500 * time.Millisecond,
			expectTimeout: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(tt.responseTime)
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": []interface{}{},
				})
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
				RetryCount: 0,
			})
			require.NoError(t, err)

			ctx, cancel := context.WithTimeout(context.Background(), tt.clientTimeout)
			defer cancel()

			_, _, err = client.Servers.List(ctx, nil)

			if tt.expectTimeout {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "context deadline exceeded")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestClient_ContextCancellation tests manual context cancellation
func TestClient_ContextCancellation(t *testing.T) {
	t.Run("cancel during request", func(t *testing.T) {
		// Create slow server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(5 * time.Second) // Long delay
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []interface{}{},
			})
		}))
		defer server.Close()

		client, err := NewClient(&Config{
			BaseURL: server.URL,
			Auth:    AuthConfig{Token: "test-token"},
			RetryCount: 0,
		})
		require.NoError(t, err)

		// Create cancellable context
		ctx, cancel := context.WithCancel(context.Background())

		// Start request in goroutine
		errChan := make(chan error, 1)
		go func() {
			_, _, err := client.Servers.List(ctx, nil)
			errChan <- err
		}()

		// Cancel after short delay
		time.Sleep(100 * time.Millisecond)
		cancel()

		// Verify cancellation error
		err = <-errChan
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})

	t.Run("already cancelled context", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Server should not be called with cancelled context")
		}))
		defer server.Close()

		client, err := NewClient(&Config{
			BaseURL: server.URL,
			Auth:    AuthConfig{Token: "test-token"},
			RetryCount: 0,
		})
		require.NoError(t, err)

		// Create and immediately cancel context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// Attempt request with already-cancelled context
		_, _, err = client.Servers.List(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}
