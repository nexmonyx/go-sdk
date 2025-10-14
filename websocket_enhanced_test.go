package nexmonyx

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Test configuration setter methods
func TestWebSocketService_SetTimeout(t *testing.T) {
	client, _ := NewClient(&Config{
		BaseURL: "https://api.example.com",
		Auth:    AuthConfig{ServerUUID: "test-uuid", ServerSecret: "test-secret"},
	})

	ws, err := client.NewWebSocketService()
	assert.NoError(t, err)
	assert.NotNil(t, ws)

	// Test setting custom timeout
	customTimeout := 45 * time.Second
	ws.SetTimeout(customTimeout)
	assert.Equal(t, customTimeout, ws.timeout)
}

func TestWebSocketService_SetReconnectDelay(t *testing.T) {
	client, _ := NewClient(&Config{
		BaseURL: "https://api.example.com",
		Auth:    AuthConfig{ServerUUID: "test-uuid", ServerSecret: "test-secret"},
	})

	ws, err := client.NewWebSocketService()
	assert.NoError(t, err)
	assert.NotNil(t, ws)

	// Test setting custom reconnect delay
	customDelay := 10 * time.Second
	ws.SetReconnectDelay(customDelay)
	assert.Equal(t, customDelay, ws.reconnectDelay)
}

func TestWebSocketService_SetMaxReconnects(t *testing.T) {
	client, _ := NewClient(&Config{
		BaseURL: "https://api.example.com",
		Auth:    AuthConfig{ServerUUID: "test-uuid", ServerSecret: "test-secret"},
	})

	ws, err := client.NewWebSocketService()
	assert.NoError(t, err)
	assert.NotNil(t, ws)

	// Test setting max reconnects
	maxReconnects := 10
	ws.SetMaxReconnects(maxReconnects)
	assert.Equal(t, maxReconnects, ws.maxReconnects)
}

// Test event handler setters
func TestWebSocketService_OnConnect(t *testing.T) {
	client, _ := NewClient(&Config{
		BaseURL: "https://api.example.com",
		Auth:    AuthConfig{ServerUUID: "test-uuid", ServerSecret: "test-secret"},
	})

	ws, err := client.NewWebSocketService()
	assert.NoError(t, err)
	assert.NotNil(t, ws)

	// Test setting onConnect callback
	called := false
	ws.OnConnect(func() {
		called = true
	})

	assert.NotNil(t, ws.onConnect)
	// Trigger the callback to verify it was set
	if ws.onConnect != nil {
		ws.onConnect()
	}
	assert.True(t, called)
}

func TestWebSocketService_OnDisconnect(t *testing.T) {
	client, _ := NewClient(&Config{
		BaseURL: "https://api.example.com",
		Auth:    AuthConfig{ServerUUID: "test-uuid", ServerSecret: "test-secret"},
	})

	ws, err := client.NewWebSocketService()
	assert.NoError(t, err)
	assert.NotNil(t, ws)

	// Test setting onDisconnect callback
	ws.OnDisconnect(func(err error) {
		// Callback set successfully
	})

	assert.NotNil(t, ws.onDisconnect)
}

func TestWebSocketService_OnMessage(t *testing.T) {
	client, _ := NewClient(&Config{
		BaseURL: "https://api.example.com",
		Auth:    AuthConfig{ServerUUID: "test-uuid", ServerSecret: "test-secret"},
	})

	ws, err := client.NewWebSocketService()
	assert.NoError(t, err)
	assert.NotNil(t, ws)

	// Test setting onMessage callback
	ws.OnMessage(func(msg *WSMessage) {
		// Callback set successfully
	})

	assert.NotNil(t, ws.onMessage)
}

// Test IsConnected method
func TestWebSocketService_IsConnected_NotConnected(t *testing.T) {
	client, _ := NewClient(&Config{
		BaseURL: "https://api.example.com",
		Auth:    AuthConfig{ServerUUID: "test-uuid", ServerSecret: "test-secret"},
	})

	ws, err := client.NewWebSocketService()
	assert.NoError(t, err)
	assert.NotNil(t, ws)

	// Should not be connected initially
	assert.False(t, ws.IsConnected())
}

// Test URL building for different protocols
func TestWebSocketService_BuildWebSocketURL(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		expectedURL string
	}{
		{
			name:        "HTTPS to WSS",
			baseURL:     "https://api.example.com",
			expectedURL: "wss://api.example.com/v1/agent/websocket",
		},
		{
			name:        "HTTP to WS",
			baseURL:     "http://localhost:8080",
			expectedURL: "ws://localhost:8080/v1/agent/websocket",
		},
		{
			name:        "Already WSS",
			baseURL:     "wss://api.example.com",
			expectedURL: "wss://api.example.com/v1/agent/websocket",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _ := NewClient(&Config{
				BaseURL: tt.baseURL,
				Auth:    AuthConfig{ServerUUID: "test-uuid", ServerSecret: "test-secret"},
			})

			ws, err := client.NewWebSocketService()
			assert.NoError(t, err)
			assert.NotNil(t, ws)

			url := ws.buildWebSocketURL()
			assert.Equal(t, tt.expectedURL, url)
		})
	}
}

// Test command methods with nil requests
func TestWebSocketService_ForceCollection_NilRequest(t *testing.T) {
	client, _ := NewClient(&Config{
		BaseURL: "https://api.example.com",
		Auth:    AuthConfig{ServerUUID: "test-uuid", ServerSecret: "test-secret"},
	})

	ws, err := client.NewWebSocketService()
	assert.NoError(t, err)

	// Should fail because not connected
	ctx := context.Background()
	_, err = ws.ForceCollection(ctx, "server-uuid", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestWebSocketService_GracefulRestart_NilRequest(t *testing.T) {
	client, _ := NewClient(&Config{
		BaseURL: "https://api.example.com",
		Auth:    AuthConfig{ServerUUID: "test-uuid", ServerSecret: "test-secret"},
	})

	ws, err := client.NewWebSocketService()
	assert.NoError(t, err)

	// Should fail because not connected
	ctx := context.Background()
	_, err = ws.GracefulRestart(ctx, "server-uuid", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestWebSocketService_RunCollection_NilRequest(t *testing.T) {
	client, _ := NewClient(&Config{
		BaseURL: "https://api.example.com",
		Auth:    AuthConfig{ServerUUID: "test-uuid", ServerSecret: "test-secret"},
	})

	ws, err := client.NewWebSocketService()
	assert.NoError(t, err)

	// Should fail because not connected
	ctx := context.Background()
	_, err = ws.RunCollection(ctx, "server-uuid", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestWebSocketService_UpdateAgent_NilRequest(t *testing.T) {
	client, _ := NewClient(&Config{
		BaseURL: "https://api.example.com",
		Auth:    AuthConfig{ServerUUID: "test-uuid", ServerSecret: "test-secret"},
	})

	ws, err := client.NewWebSocketService()
	assert.NoError(t, err)

	// Should fail because not connected
	ctx := context.Background()
	_, err = ws.UpdateAgent(ctx, "server-uuid", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestWebSocketService_RestartAgent_NilRequest(t *testing.T) {
	client, _ := NewClient(&Config{
		BaseURL: "https://api.example.com",
		Auth:    AuthConfig{ServerUUID: "test-uuid", ServerSecret: "test-secret"},
	})

	ws, err := client.NewWebSocketService()
	assert.NoError(t, err)

	// Should fail because not connected
	ctx := context.Background()
	_, err = ws.RestartAgent(ctx, "server-uuid", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

// Test generateCorrelationID
func TestGenerateCorrelationID(t *testing.T) {
	id1 := generateCorrelationID()
	assert.NotEmpty(t, id1)
	assert.Contains(t, id1, "sdk-")

	// Generate another ID and ensure they're different
	time.Sleep(1 * time.Millisecond)
	id2 := generateCorrelationID()
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2, "Correlation IDs should be unique")
}

// Test Disconnect when not connected
func TestWebSocketService_Disconnect_NotConnected(t *testing.T) {
	client, _ := NewClient(&Config{
		BaseURL: "https://api.example.com",
		Auth:    AuthConfig{ServerUUID: "test-uuid", ServerSecret: "test-secret"},
	})

	ws, err := client.NewWebSocketService()
	assert.NoError(t, err)

	// Disconnect when not connected should not error
	err = ws.Disconnect()
	assert.NoError(t, err)
}

// Test NewWebSocketService without credentials
func TestWebSocketService_NewWebSocketService_NoCredentials(t *testing.T) {
	tests := []struct {
		name     string
		auth     AuthConfig
		wantErr  bool
		errorMsg string
	}{
		{
			name:     "Missing server UUID",
			auth:     AuthConfig{ServerSecret: "secret"},
			wantErr:  true,
			errorMsg: "WebSocket service requires server authentication credentials",
		},
		{
			name:     "Missing server secret",
			auth:     AuthConfig{ServerUUID: "uuid"},
			wantErr:  true,
			errorMsg: "WebSocket service requires server authentication credentials",
		},
		{
			name:     "Missing both",
			auth:     AuthConfig{},
			wantErr:  true,
			errorMsg: "WebSocket service requires server authentication credentials",
		},
		{
			name:     "Valid credentials",
			auth:     AuthConfig{ServerUUID: "uuid", ServerSecret: "secret"},
			wantErr:  false,
			errorMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _ := NewClient(&Config{
				BaseURL: "https://api.example.com",
				Auth:    tt.auth,
			})

			ws, err := client.NewWebSocketService()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, ws)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, ws)
			}
		})
	}
}
