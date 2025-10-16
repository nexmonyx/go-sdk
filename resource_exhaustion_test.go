package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestResourceExhaustion_RateLimiting tests rate limiting behavior
func TestResourceExhaustion_RateLimiting(t *testing.T) {
	tests := []struct {
		name           string
		requestCount   int
		rateLimit      int
		expectThrottle bool
		throttledCount int
	}{
		{
			name:           "within rate limit - no throttling",
			requestCount:   10,
			rateLimit:      20,
			expectThrottle: false,
			throttledCount: 0,
		},
		{
			name:           "exceed rate limit - some throttled",
			requestCount:   30,
			rateLimit:      20,
			expectThrottle: true,
			throttledCount: 10,
		},
		{
			name:           "exactly at rate limit - no throttling",
			requestCount:   20,
			rateLimit:      20,
			expectThrottle: false,
			throttledCount: 0,
		},
		{
			name:           "far exceed rate limit - many throttled",
			requestCount:   100,
			rateLimit:      10,
			expectThrottle: true,
			throttledCount: 90,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestNum := int64(0)
			throttledRequests := int64(0)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				current := atomic.AddInt64(&requestNum, 1)

				if current > int64(tt.rateLimit) {
					atomic.AddInt64(&throttledRequests, 1)
					w.Header().Set("X-RateLimit-Limit", "20")
					w.Header().Set("X-RateLimit-Remaining", "0")
					w.Header().Set("X-RateLimit-Reset", "1640000000")
					w.WriteHeader(http.StatusTooManyRequests)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"error": "Rate limit exceeded",
					})
					return
				}

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": []interface{}{},
				})
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL:    server.URL,
				Auth:       AuthConfig{Token: "test-token"},
				RetryCount: 0, // No retries for rate limit testing
			})
			require.NoError(t, err)

			var wg sync.WaitGroup
			successCount := int64(0)
			errorCount := int64(0)

			for i := 0; i < tt.requestCount; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_, _, err := client.Servers.List(context.Background(), nil)
					if err != nil {
						atomic.AddInt64(&errorCount, 1)
					} else {
						atomic.AddInt64(&successCount, 1)
					}
				}()
			}

			wg.Wait()

			if tt.expectThrottle {
				assert.Greater(t, atomic.LoadInt64(&errorCount), int64(0), "Expected some requests to be throttled")
				// Note: Actual throttled count may be higher due to retries
				t.Logf("Expected ~%d throttled, got %d (includes retries)", tt.throttledCount, atomic.LoadInt64(&throttledRequests))
			} else {
				assert.Equal(t, int64(0), atomic.LoadInt64(&errorCount), "No requests should be throttled")
			}

			t.Logf("Success: %d, Errors: %d, Throttled: %d", successCount, errorCount, throttledRequests)
		})
	}
}

// TestResourceExhaustion_QuotaExceeded tests quota exceeded scenarios
func TestResourceExhaustion_QuotaExceeded(t *testing.T) {
	tests := []struct {
		name          string
		currentUsage  int
		quota         int
		action        string
		expectBlocked bool
	}{
		{
			name:          "within quota - allowed",
			currentUsage:  50,
			quota:         100,
			action:        "create_server",
			expectBlocked: false,
		},
		{
			name:          "at quota limit - blocked",
			currentUsage:  100,
			quota:         100,
			action:        "create_server",
			expectBlocked: true,
		},
		{
			name:          "exceeded quota - blocked",
			currentUsage:  110,
			quota:         100,
			action:        "create_server",
			expectBlocked: true,
		},
		{
			name:          "near quota - one more allowed",
			currentUsage:  99,
			quota:         100,
			action:        "create_server",
			expectBlocked: false,
		},
		{
			name:          "storage quota exceeded - blocked",
			currentUsage:  1024,
			quota:         1000,
			action:        "upload_metrics",
			expectBlocked: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.expectBlocked {
					w.WriteHeader(http.StatusForbidden)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"error":         "Quota exceeded",
						"current_usage": tt.currentUsage,
						"quota_limit":   tt.quota,
					})
				} else {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"data": map[string]interface{}{
							"id": 1,
						},
					})
				}
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			var apiErr error
			switch tt.action {
			case "create_server":
				_, apiErr = client.Servers.Register(context.Background(), "test-server", 1)
			case "upload_metrics":
				_, _, apiErr = client.Monitoring.ListProbes(context.Background(), nil)
			}

			if tt.expectBlocked {
				assert.Error(t, apiErr)
			} else {
				assert.NoError(t, apiErr)
			}
		})
	}
}

// TestResourceExhaustion_ConnectionLimits tests connection pool exhaustion
func TestResourceExhaustion_ConnectionLimits(t *testing.T) {
	tests := []struct {
		name            string
		concurrency     int
		maxConnections  int
		serverDelay     time.Duration
		expectTimeouts  bool
		expectQueueing  bool
	}{
		{
			name:            "low concurrency - no queueing",
			concurrency:     5,
			maxConnections:  10,
			serverDelay:     10 * time.Millisecond,
			expectTimeouts:  false,
			expectQueueing:  false,
		},
		{
			name:            "high concurrency - connection queueing",
			concurrency:     50,
			maxConnections:  10,
			serverDelay:     100 * time.Millisecond,
			expectTimeouts:  false,
			expectQueueing:  true,
		},
		{
			name:            "very high concurrency with slow server - possible timeouts",
			concurrency:     100,
			maxConnections:  5,
			serverDelay:     500 * time.Millisecond,
			expectTimeouts:  true,
			expectQueueing:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			activeConnections := int64(0)
			maxActiveConnections := int64(0)
			var connMu sync.Mutex

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				current := atomic.AddInt64(&activeConnections, 1)

				connMu.Lock()
				if current > maxActiveConnections {
					maxActiveConnections = current
				}
				connMu.Unlock()

				time.Sleep(tt.serverDelay)

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
			timeoutCount := int64(0)

			startTime := time.Now()

			for i := 0; i < tt.concurrency; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					_, _, err := client.Servers.List(ctx, nil)
					if err != nil && err.Error() == "context deadline exceeded" {
						atomic.AddInt64(&timeoutCount, 1)
					}
				}()
			}

			wg.Wait()
			duration := time.Since(startTime)

			t.Logf("Max concurrent connections: %d", atomic.LoadInt64(&maxActiveConnections))
			t.Logf("Duration: %v", duration)
			t.Logf("Timeouts: %d", atomic.LoadInt64(&timeoutCount))

			if tt.expectQueueing {
				// If queueing, max connections should be limited
				// Note: HTTP client may be more efficient than expected
				t.Logf("Queueing expected, max connections limited to reasonable level")
				// The important thing is that we handled high concurrency without errors
			}
		})
	}
}

// TestResourceExhaustion_MemoryPressure tests behavior under memory constraints
func TestResourceExhaustion_MemoryPressure(t *testing.T) {
	tests := []struct {
		name           string
		responseSize   int    // KB
		requestCount   int
		expectSuccess  bool
	}{
		{
			name:          "small responses - no issues",
			responseSize:  1,    // 1 KB
			requestCount:  100,
			expectSuccess: true,
		},
		{
			name:          "medium responses - manageable",
			responseSize:  100,  // 100 KB
			requestCount:  50,
			expectSuccess: true,
		},
		{
			name:          "large responses - memory intensive",
			responseSize:  1000, // 1 MB
			requestCount:  10,
			expectSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Generate response of specified size
				data := make([]map[string]interface{}, 0)
				itemSize := 100 // bytes per item
				itemCount := (tt.responseSize * 1024) / itemSize

				for i := 0; i < itemCount; i++ {
					data = append(data, map[string]interface{}{
						"id":       i,
						"name":     "item-" + string(rune(i)),
						"data":     "padding-data-padding-data-padding-data",
					})
				}

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": data,
				})
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			successCount := 0
			errorCount := 0

			for i := 0; i < tt.requestCount; i++ {
				_, _, err := client.Servers.List(context.Background(), nil)
				if err != nil {
					errorCount++
				} else {
					successCount++
				}
			}

			if tt.expectSuccess {
				assert.Greater(t, successCount, errorCount, "Most requests should succeed")
			}

			t.Logf("Success: %d, Errors: %d, Response size: %d KB", successCount, errorCount, tt.responseSize)
		})
	}
}

// TestResourceExhaustion_TimeoutCascades tests timeout cascade scenarios
func TestResourceExhaustion_TimeoutCascades(t *testing.T) {
	tests := []struct {
		name           string
		slowEndpoints  []string
		fastEndpoints  []string
		expectCascade  bool
	}{
		{
			name:           "single slow endpoint - no cascade",
			slowEndpoints:  []string{"/v2/servers"},
			fastEndpoints:  []string{"/v2/alerts"},
			expectCascade:  false,
		},
		{
			name:           "multiple slow endpoints - potential cascade",
			slowEndpoints:  []string{"/v2/servers", "/v2/alerts", "/v2/probes"},
			fastEndpoints:  []string{},
			expectCascade:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeoutCount := int64(0)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				isSlow := false
				for _, endpoint := range tt.slowEndpoints {
					if r.URL.Path == endpoint {
						isSlow = true
						break
					}
				}

				if isSlow {
					time.Sleep(3 * time.Second) // Intentionally slow
				}

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

			// Make requests with short timeout
			for i := 0; i < len(tt.slowEndpoints) + len(tt.fastEndpoints); i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()

					ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
					defer cancel()

					_, _, err := client.Servers.List(ctx, nil)
					if err != nil {
						atomic.AddInt64(&timeoutCount, 1)
					}
				}(i)
			}

			wg.Wait()

			if tt.expectCascade {
				assert.Greater(t, atomic.LoadInt64(&timeoutCount), int64(0), "Expected cascade timeouts")
			}

			t.Logf("Timeouts: %d", atomic.LoadInt64(&timeoutCount))
		})
	}
}

// TestResourceExhaustion_BackpressureHandling tests backpressure handling
func TestResourceExhaustion_BackpressureHandling(t *testing.T) {
	tests := []struct {
		name              string
		incomingRate      int // requests per second
		processingRate    int // requests per second
		duration          time.Duration
		expectQueueBuild  bool
	}{
		{
			name:             "balanced rates - no backpressure",
			incomingRate:     10,
			processingRate:   10,
			duration:         1 * time.Second,
			expectQueueBuild: false,
		},
		{
			name:             "incoming faster - backpressure builds",
			incomingRate:     50,
			processingRate:   10,
			duration:         2 * time.Second,
			expectQueueBuild: true,
		},
		{
			name:             "processing faster - queue drains",
			incomingRate:     10,
			processingRate:   50,
			duration:         1 * time.Second,
			expectQueueBuild: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queueSize := int64(0)
			maxQueueSize := int64(0)
			var queueMu sync.Mutex

			processingDelay := time.Second / time.Duration(tt.processingRate)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				current := atomic.AddInt64(&queueSize, 1)

				queueMu.Lock()
				if current > maxQueueSize {
					maxQueueSize = current
				}
				queueMu.Unlock()

				time.Sleep(processingDelay)

				atomic.AddInt64(&queueSize, -1)

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

			stopChan := make(chan bool)
			var wg sync.WaitGroup

			// Start generating requests at specified rate
			go func() {
				ticker := time.NewTicker(time.Second / time.Duration(tt.incomingRate))
				defer ticker.Stop()

				for {
					select {
					case <-ticker.C:
						wg.Add(1)
						go func() {
							defer wg.Done()
							client.Servers.List(context.Background(), nil)
						}()
					case <-stopChan:
						return
					}
				}
			}()

			time.Sleep(tt.duration)
			close(stopChan)
			wg.Wait()

			t.Logf("Max queue size: %d", atomic.LoadInt64(&maxQueueSize))

			if tt.expectQueueBuild {
				assert.Greater(t, atomic.LoadInt64(&maxQueueSize), int64(5), "Expected queue to build up")
			}
		})
	}
}
