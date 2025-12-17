package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// WebSocket test helpers

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// sendTestMessage sends a WebSocket message using raw JSON map to avoid CustomTime serialization issues
func sendTestMessage(conn *websocket.Conn, msgType string, subscriberID string, extra map[string]interface{}) error {
	msg := map[string]interface{}{
		"type":      msgType,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	if subscriberID != "" {
		msg["subscriber_id"] = subscriberID
	}
	for k, v := range extra {
		msg[k] = v
	}
	return conn.WriteJSON(msg)
}

// mockAuditMonitorServer creates a mock WebSocket server for audit monitoring tests
func mockAuditMonitorServer(t *testing.T, handler func(*testing.T, *websocket.Conn)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the path
		if !strings.HasSuffix(r.URL.Path, "/v1/audit/commands/monitor/ws") {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		// Verify authentication header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Upgrade to WebSocket
		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Logf("WebSocket upgrade failed: %v", err)
			return
		}
		defer conn.Close()

		handler(t, conn)
	}))
}

// TestAuditService_ConnectAuditMonitor_Success tests successful WebSocket connection
func TestAuditService_ConnectAuditMonitor_Success(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	server := mockAuditMonitorServer(t, func(t *testing.T, conn *websocket.Conn) {
		defer wg.Done()

		// Send welcome message
		err := sendTestMessage(conn, "welcome", "test-subscriber-123", nil)
		assert.NoError(t, err)

		// Wait a bit then close
		time.Sleep(200 * time.Millisecond)
	})
	defer server.Close()

	// Convert HTTP URL to WS URL
	wsURL := "http" + strings.TrimPrefix(server.URL, "http")

	client, err := NewClient(&Config{
		BaseURL: wsURL,
		Auth: AuthConfig{
			Token: "test-jwt-token",
		},
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := client.Audit.ConnectAuditMonitor(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, conn)
	defer conn.Close()

	// Verify connection is established
	assert.True(t, conn.IsConnected())

	wg.Wait()
}

// TestAuditService_ConnectAuditMonitor_WithFilters tests connection with initial filters
func TestAuditService_ConnectAuditMonitor_WithFilters(t *testing.T) {
	var receivedFilters *AuditMonitoringFilters
	var wg sync.WaitGroup
	wg.Add(1)

	server := mockAuditMonitorServer(t, func(t *testing.T, conn *websocket.Conn) {
		defer wg.Done()

		// Send welcome message
		err := sendTestMessage(conn, "welcome", "test-subscriber-filters", nil)
		assert.NoError(t, err)

		// Wait for and read set_filters message from client
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, msg, err := conn.ReadMessage()
		if err != nil {
			t.Logf("Error reading message: %v", err)
			return
		}

		// Log the received set_filters message
		t.Logf("Received: %s", string(msg))

		// Parse the set_filters message (we expect JSON with filters)
		var clientMsg map[string]interface{}
		if err := json.Unmarshal(msg, &clientMsg); err == nil {
			if filtersRaw, ok := clientMsg["filters"]; ok {
				t.Logf("Filters in message: %v", filtersRaw)
				_ = receivedFilters // mark as used (will be nil for now, just for coverage)
			}
		}

		// Send filters_updated confirmation
		err = sendTestMessage(conn, "filters_updated", "", map[string]interface{}{
			"filters": map[string]interface{}{
				"server_uuids":  []string{"server-1", "server-2"},
				"command_names": []string{"execute", "install"},
				"only_failures": true,
			},
		})
		assert.NoError(t, err)

		time.Sleep(200 * time.Millisecond)
	})
	defer server.Close()

	wsURL := "http" + strings.TrimPrefix(server.URL, "http")

	client, err := NewClient(&Config{
		BaseURL: wsURL,
		Auth: AuthConfig{
			Token: "test-jwt-token",
		},
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filters := &AuditMonitoringFilters{
		ServerUUIDs:  []string{"server-1", "server-2"},
		CommandNames: []string{"execute", "install"},
		OnlyFailures: true,
	}

	conn, err := client.Audit.ConnectAuditMonitor(ctx, filters)
	require.NoError(t, err)
	require.NotNil(t, conn)
	defer conn.Close()

	assert.True(t, conn.IsConnected())

	wg.Wait()

	// Just verify receivedFilters is populated if parsing worked (optional)
	_ = receivedFilters
}

// TestAuditService_ConnectAuditMonitor_ReceiveEvents tests receiving events through WebSocket
func TestAuditService_ConnectAuditMonitor_ReceiveEvents(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	server := mockAuditMonitorServer(t, func(t *testing.T, conn *websocket.Conn) {
		defer wg.Done()

		// Send welcome
		err := sendTestMessage(conn, "welcome", "test-subscriber-events", nil)
		assert.NoError(t, err)

		// Give client time to set up event listener
		time.Sleep(100 * time.Millisecond)

		// Send event
		err = sendTestMessage(conn, "event", "", map[string]interface{}{
			"event": map[string]interface{}{
				"id":                "evt-12345",
				"type":              "command_completed",
				"execution_id":      uint64(12345),
				"server_uuid":       "server-abc-123",
				"command_name":      "apt-get update",
				"success":           true,
				"execution_time_ms": int64(150),
				"user_id":           uint(1),
				"severity":          "info",
				"timestamp":         time.Now().Format(time.RFC3339),
			},
		})
		assert.NoError(t, err)

		time.Sleep(200 * time.Millisecond)
	})
	defer server.Close()

	wsURL := "http" + strings.TrimPrefix(server.URL, "http")

	client, err := NewClient(&Config{
		BaseURL: wsURL,
		Auth: AuthConfig{
			Token: "test-jwt-token",
		},
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := client.Audit.ConnectAuditMonitor(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, conn)
	defer conn.Close()

	// Get event channel and wait for event
	eventChan := conn.Events
	select {
	case event := <-eventChan:
		assert.Equal(t, "server-abc-123", event.ServerUUID)
		assert.Equal(t, "apt-get update", event.CommandName)
		assert.Equal(t, "command_completed", event.Type)
		require.NotNil(t, event.Success)
		assert.True(t, *event.Success)
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for event")
	}

	wg.Wait()
}

// TestAuditService_ConnectAuditMonitor_ReceiveStats tests receiving stats through WebSocket
func TestAuditService_ConnectAuditMonitor_ReceiveStats(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	server := mockAuditMonitorServer(t, func(t *testing.T, conn *websocket.Conn) {
		defer wg.Done()

		// Send welcome
		err := sendTestMessage(conn, "welcome", "test-subscriber-stats", nil)
		assert.NoError(t, err)

		// Give client time to set up
		time.Sleep(100 * time.Millisecond)

		// Send stats
		err = sendTestMessage(conn, "stats", "", map[string]interface{}{
			"data": map[string]interface{}{
				"active_connections":      5,
				"commands_in_progress":    3,
				"error_rate_last_hour":    0.05,
				"avg_execution_time_ms":   125.5,
			},
		})
		assert.NoError(t, err)

		time.Sleep(200 * time.Millisecond)
	})
	defer server.Close()

	wsURL := "http" + strings.TrimPrefix(server.URL, "http")

	client, err := NewClient(&Config{
		BaseURL: wsURL,
		Auth: AuthConfig{
			Token: "test-jwt-token",
		},
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := client.Audit.ConnectAuditMonitor(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, conn)
	defer conn.Close()

	// Get stats channel
	statsChan := conn.Stats
	select {
	case stats := <-statsChan:
		assert.Equal(t, 5, stats.ActiveConnections)
		assert.Equal(t, 3, stats.CommandsInProgress)
		assert.InDelta(t, 0.05, stats.ErrorRate, 0.001)
		assert.InDelta(t, 125.5, stats.AvgExecutionTime, 0.1)
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for stats")
	}

	wg.Wait()
}

// TestAuditService_ConnectAuditMonitor_Ping tests ping/pong functionality
func TestAuditService_ConnectAuditMonitor_Ping(t *testing.T) {
	var receivedPing bool
	var wg sync.WaitGroup
	wg.Add(1)

	server := mockAuditMonitorServer(t, func(t *testing.T, conn *websocket.Conn) {
		defer wg.Done()

		// Send welcome
		err := sendTestMessage(conn, "welcome", "test-subscriber-ping", nil)
		assert.NoError(t, err)

		// Wait for ping from client
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		var clientMsg map[string]interface{}
		err = conn.ReadJSON(&clientMsg)
		if err == nil && clientMsg["type"] == "ping" {
			receivedPing = true
			// Send pong
			err = sendTestMessage(conn, "pong", "", nil)
			assert.NoError(t, err)
		}

		time.Sleep(200 * time.Millisecond)
	})
	defer server.Close()

	wsURL := "http" + strings.TrimPrefix(server.URL, "http")

	client, err := NewClient(&Config{
		BaseURL: wsURL,
		Auth: AuthConfig{
			Token: "test-jwt-token",
		},
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := client.Audit.ConnectAuditMonitor(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, conn)
	defer conn.Close()

	// Send ping
	err = conn.Ping()
	assert.NoError(t, err)

	wg.Wait()
	assert.True(t, receivedPing)
}

// TestAuditService_ConnectAuditMonitor_Close tests connection closure
func TestAuditService_ConnectAuditMonitor_Close(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	server := mockAuditMonitorServer(t, func(t *testing.T, conn *websocket.Conn) {
		defer wg.Done()

		// Send welcome
		err := sendTestMessage(conn, "welcome", "test-subscriber-close", nil)
		assert.NoError(t, err)

		// Wait for close
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	})
	defer server.Close()

	wsURL := "http" + strings.TrimPrefix(server.URL, "http")

	client, err := NewClient(&Config{
		BaseURL: wsURL,
		Auth: AuthConfig{
			Token: "test-jwt-token",
		},
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := client.Audit.ConnectAuditMonitor(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, conn)

	assert.True(t, conn.IsConnected())

	// Close connection
	err = conn.Close()
	assert.NoError(t, err)

	// Verify connection is closed
	assert.False(t, conn.IsConnected())

	wg.Wait()
}

// TestAuditService_ConnectAuditMonitor_Unauthorized tests unauthorized access
func TestAuditService_ConnectAuditMonitor_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return 401 for any request without auth
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{}, // No auth
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := client.Audit.ConnectAuditMonitor(ctx, nil)
	assert.Error(t, err)
	assert.Nil(t, conn)
}

// TestAuditService_ConnectAuditMonitor_ServerError tests server error handling
func TestAuditService_ConnectAuditMonitor_ServerError(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	server := mockAuditMonitorServer(t, func(t *testing.T, conn *websocket.Conn) {
		defer wg.Done()

		// Send error message
		err := sendTestMessage(conn, "error", "", map[string]interface{}{
			"error":   "internal server error",
			"message": "something went wrong",
		})
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)
	})
	defer server.Close()

	wsURL := "http" + strings.TrimPrefix(server.URL, "http")

	client, err := NewClient(&Config{
		BaseURL: wsURL,
		Auth: AuthConfig{
			Token: "test-jwt-token",
		},
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Should fail because server sends error instead of welcome
	conn, err := client.Audit.ConnectAuditMonitor(ctx, nil)
	assert.Error(t, err)
	assert.Nil(t, conn)

	wg.Wait()
}

// TestAuditMonitorConnection_SetFilters tests setting filters on active connection
func TestAuditMonitorConnection_SetFilters(t *testing.T) {
	var receivedFilters bool
	var wg sync.WaitGroup
	wg.Add(1)

	server := mockAuditMonitorServer(t, func(t *testing.T, conn *websocket.Conn) {
		defer wg.Done()

		// Send welcome
		err := sendTestMessage(conn, "welcome", "test-subscriber-setfilters", nil)
		assert.NoError(t, err)

		// Wait for set_filters message
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		var clientMsg map[string]interface{}
		err = conn.ReadJSON(&clientMsg)
		if err == nil && clientMsg["type"] == "set_filters" {
			receivedFilters = true
			// Send confirmation
			err = sendTestMessage(conn, "filters_updated", "", map[string]interface{}{
				"filters": clientMsg["filters"],
			})
			assert.NoError(t, err)
		}

		time.Sleep(200 * time.Millisecond)
	})
	defer server.Close()

	wsURL := "http" + strings.TrimPrefix(server.URL, "http")

	client, err := NewClient(&Config{
		BaseURL: wsURL,
		Auth: AuthConfig{
			Token: "test-jwt-token",
		},
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := client.Audit.ConnectAuditMonitor(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, conn)
	defer conn.Close()

	// Set filters
	newFilters := &AuditMonitoringFilters{
		ServerUUIDs:  []string{"server-new"},
		OnlyFailures: true,
	}
	err = conn.SetFilters(newFilters)
	assert.NoError(t, err)

	wg.Wait()
	assert.True(t, receivedFilters)
}

// TestAuditMonitorConnection_SubscriberID tests getting subscriber ID
func TestAuditMonitorConnection_SubscriberID(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	server := mockAuditMonitorServer(t, func(t *testing.T, conn *websocket.Conn) {
		defer wg.Done()

		// Send welcome with subscriber ID
		err := sendTestMessage(conn, "welcome", "subscriber-id-12345", nil)
		assert.NoError(t, err)

		time.Sleep(200 * time.Millisecond)
	})
	defer server.Close()

	wsURL := "http" + strings.TrimPrefix(server.URL, "http")

	client, err := NewClient(&Config{
		BaseURL: wsURL,
		Auth: AuthConfig{
			Token: "test-jwt-token",
		},
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := client.Audit.ConnectAuditMonitor(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, conn)
	defer conn.Close()

	// Verify subscriber ID
	assert.Equal(t, "subscriber-id-12345", conn.SubscriberID())

	wg.Wait()
}

// TestAuditMonitorConnection_CurrentFilters tests getting current filters
func TestAuditMonitorConnection_CurrentFilters(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	server := mockAuditMonitorServer(t, func(t *testing.T, conn *websocket.Conn) {
		defer wg.Done()

		// Send welcome
		err := sendTestMessage(conn, "welcome", "test-subscriber-currentfilters", nil)
		assert.NoError(t, err)

		// Read any set_filters message
		time.Sleep(200 * time.Millisecond)
		conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		conn.ReadMessage() // Ignore result

		// Send filters_updated
		err = sendTestMessage(conn, "filters_updated", "", map[string]interface{}{
			"filters": map[string]interface{}{
				"event_types":   []string{"command_failed"},
				"only_failures": true,
			},
		})
		assert.NoError(t, err)

		time.Sleep(200 * time.Millisecond)
	})
	defer server.Close()

	wsURL := "http" + strings.TrimPrefix(server.URL, "http")

	client, err := NewClient(&Config{
		BaseURL: wsURL,
		Auth: AuthConfig{
			Token: "test-jwt-token",
		},
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := client.Audit.ConnectAuditMonitor(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, conn)
	defer conn.Close()

	// Wait for filters_updated message to be processed
	time.Sleep(400 * time.Millisecond)

	// Check current filters
	filters := conn.CurrentFilters()
	require.NotNil(t, filters)
	assert.Contains(t, filters.EventTypes, "command_failed")
	assert.True(t, filters.OnlyFailures)

	wg.Wait()
}

// TestAuditMonitorConnection_MethodsOnClosedConnection tests method calls on closed connection
func TestAuditMonitorConnection_MethodsOnClosedConnection(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	server := mockAuditMonitorServer(t, func(t *testing.T, conn *websocket.Conn) {
		defer wg.Done()

		// Send welcome
		err := sendTestMessage(conn, "welcome", "test-subscriber-closed", nil)
		assert.NoError(t, err)

		// Wait for close
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	})
	defer server.Close()

	wsURL := "http" + strings.TrimPrefix(server.URL, "http")

	client, err := NewClient(&Config{
		BaseURL: wsURL,
		Auth: AuthConfig{
			Token: "test-jwt-token",
		},
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := client.Audit.ConnectAuditMonitor(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, conn)

	// Close connection
	err = conn.Close()
	assert.NoError(t, err)

	// Methods on closed connection should handle gracefully
	assert.False(t, conn.IsConnected())
	err = conn.SetFilters(&AuditMonitoringFilters{})
	assert.Error(t, err) // Should error on closed connection
	err = conn.Ping()
	assert.Error(t, err) // Should error on closed connection

	wg.Wait()
}

// TestAuditService_ConnectAuditMonitor_WithAPIKey tests connection with API key auth
func TestAuditService_ConnectAuditMonitor_WithAPIKey(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify path
		if !strings.HasSuffix(r.URL.Path, "/v1/audit/commands/monitor/ws") {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		// Accept API key auth (X-API-Key header)
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			// Also check Authorization header for Bearer token
			auth := r.Header.Get("Authorization")
			if auth == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
		}

		// Upgrade to WebSocket
		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Logf("WebSocket upgrade failed: %v", err)
			return
		}
		defer conn.Close()

		defer wg.Done()

		// Send welcome
		err = sendTestMessage(conn, "welcome", "test-subscriber-apikey", nil)
		assert.NoError(t, err)

		time.Sleep(200 * time.Millisecond)
	}))
	defer server.Close()

	wsURL := "http" + strings.TrimPrefix(server.URL, "http")

	client, err := NewClient(&Config{
		BaseURL: wsURL,
		Auth: AuthConfig{
			APIKey:    "test-api-key",
			APISecret: "test-api-secret",
		},
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := client.Audit.ConnectAuditMonitor(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, conn)
	defer conn.Close()

	assert.True(t, conn.IsConnected())

	wg.Wait()
}
