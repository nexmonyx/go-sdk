package nexmonyx

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketServiceImpl provides WebSocket communication capabilities for sending commands to agents
type WebSocketServiceImpl struct {
	client *Client
	
	// Connection management
	conn       *websocket.Conn
	connected  bool
	connecting bool
	mu         sync.RWMutex
	
	// Message correlation
	pendingResponses map[string]chan *WSCommandResponse
	responseMutex    sync.RWMutex
	
	// Configuration
	timeout         time.Duration
	reconnectDelay  time.Duration
	maxReconnects   int
	
	// Context for connection management
	ctx    context.Context
	cancel context.CancelFunc
	
	// Message handlers
	onConnect    func()
	onDisconnect func(error)
	onMessage    func(*WSMessage)
}

// WebSocket message types matching the API WebSocket manager
const (
	WSTypeAuth            = "auth"
	WSTypeAuthResponse    = "auth_response"
	WSTypePing            = "ping"
	WSTypePong            = "pong"
	WSTypeCommand         = "command"
	WSTypeCommandResponse = "command_response"
	WSTypeRequest         = "request"
	WSTypeRequestResponse = "request_response"
	WSTypeUpdateProgress  = "update_progress"
	WSTypeError           = "error"

	// WSProtocolVersion is the WebSocket protocol version
	WSProtocolVersion = "1.0"
)

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type      string          `json:"type"`
	ID        string          `json:"id,omitempty"`
	Timestamp int64           `json:"timestamp,omitempty"`
	Priority  int             `json:"priority,omitempty"` // 0=normal, 1=high, 2=urgent
	Payload   json.RawMessage `json:"payload,omitempty"`
}

// WSAuthPayload represents authentication payload
type WSAuthPayload struct {
	ServerUUID      string   `json:"server_uuid"`
	ServerSecret    string   `json:"server_secret"`
	AgentVersion    string   `json:"agent_version"`
	ProtocolVersion string   `json:"protocol_version"`         // WebSocket protocol version (e.g., "1.0")
	Capabilities    []string `json:"capabilities"`
	OrganizationID  int      `json:"organization_id,omitempty"` // Optional organization ID
}

// WSAuthResponsePayload represents authentication response
type WSAuthResponsePayload struct {
	Status            string `json:"status"`
	SessionID         string `json:"session_id"`
	HeartbeatInterval int    `json:"heartbeat_interval"`
	Error             string `json:"error,omitempty"`
}

// WSCommandPayload represents a command message
type WSCommandPayload struct {
	Command string          `json:"command"`
	Payload json.RawMessage `json:"payload"`
}

// WSCommandResponse represents the response from a command execution
type WSCommandResponse struct {
	Success  bool                   `json:"success"`
	Data     json.RawMessage        `json:"data,omitempty"`
	Error    string                 `json:"error,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Command-specific payload structures for the 8 system commands
type CollectionRequest struct {
	CollectorTypes []string `json:"collector_types,omitempty"`
	Force          bool     `json:"force,omitempty"`
	Comprehensive  bool     `json:"comprehensive,omitempty"`
	Timeout        int      `json:"timeout,omitempty"`
}

type UpdateRequest struct {
	Version     string `json:"version,omitempty"`
	Force       bool   `json:"force,omitempty"`
	Immediate   bool   `json:"immediate,omitempty"`
	DownloadURL string `json:"download_url,omitempty"`
}

type RestartRequest struct {
	Delay    int    `json:"delay,omitempty"`
	Reason   string `json:"reason,omitempty"`
	Graceful bool   `json:"graceful,omitempty"`
}

// NewWebSocketService creates a new WebSocket service instance
func (c *Client) NewWebSocketService() (*WebSocketServiceImpl, error) {
	if c.config.Auth.ServerUUID == "" || c.config.Auth.ServerSecret == "" {
		return nil, fmt.Errorf("WebSocket service requires server authentication credentials")
	}

	ctx, cancel := context.WithCancel(context.Background())
	
	ws := &WebSocketServiceImpl{
		client:           c,
		pendingResponses: make(map[string]chan *WSCommandResponse),
		timeout:          30 * time.Second,
		reconnectDelay:   5 * time.Second,
		maxReconnects:    5,
		ctx:              ctx,
		cancel:           cancel,
	}

	return ws, nil
}

// Connect establishes a WebSocket connection to the API
func (ws *WebSocketServiceImpl) Connect() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.connected {
		return nil
	}

	if ws.connecting {
		return fmt.Errorf("connection already in progress")
	}

	ws.connecting = true
	defer func() { ws.connecting = false }()

	// Build WebSocket URL
	wsURL := ws.buildWebSocketURL()

	// Create WebSocket connection
	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	ws.conn = conn
	ws.connected = true

	// Authenticate
	if err := ws.authenticate(); err != nil {
		ws.conn.Close()
		ws.conn = nil
		ws.connected = false
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Start message handling
	go ws.handleMessages()
	go ws.pingHandler()

	if ws.onConnect != nil {
		ws.onConnect()
	}

	return nil
}

// Disconnect closes the WebSocket connection
func (ws *WebSocketServiceImpl) Disconnect() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if !ws.connected || ws.conn == nil {
		return nil
	}

	// Cancel context to stop goroutines
	ws.cancel()

	// Close connection
	err := ws.conn.Close()
	ws.conn = nil
	ws.connected = false

	// Clean up pending responses
	ws.responseMutex.Lock()
	for id, ch := range ws.pendingResponses {
		select {
		case ch <- &WSCommandResponse{
			Success: false,
			Error:   "connection closed",
		}:
		default:
		}
		close(ch)
		delete(ws.pendingResponses, id)
	}
	ws.responseMutex.Unlock()

	if ws.onDisconnect != nil {
		ws.onDisconnect(err)
	}

	return err
}

// IsConnected returns whether the WebSocket connection is active
func (ws *WebSocketServiceImpl) IsConnected() bool {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.connected
}

// =============================================================================
// System Command Methods
// =============================================================================

// RunCollection triggers a metrics collection run on the agent
func (ws *WebSocketServiceImpl) RunCollection(ctx context.Context, serverUUID string, req *CollectionRequest) (*WSCommandResponse, error) {
	if req == nil {
		req = &CollectionRequest{}
	}
	return ws.sendCommand(ctx, serverUUID, "run_collection", req)
}

// ForceCollection forces an immediate comprehensive metrics collection
func (ws *WebSocketServiceImpl) ForceCollection(ctx context.Context, serverUUID string, req *CollectionRequest) (*WSCommandResponse, error) {
	if req == nil {
		req = &CollectionRequest{Force: true, Comprehensive: true}
	} else {
		req.Force = true
		req.Comprehensive = true
	}
	return ws.sendCommand(ctx, serverUUID, "force_collection", req)
}

// UpdateAgent triggers an agent update
func (ws *WebSocketServiceImpl) UpdateAgent(ctx context.Context, serverUUID string, req *UpdateRequest) (*WSCommandResponse, error) {
	if req == nil {
		req = &UpdateRequest{}
	}
	return ws.sendCommand(ctx, serverUUID, "update_agent", req)
}

// CheckUpdates checks for available agent updates
func (ws *WebSocketServiceImpl) CheckUpdates(ctx context.Context, serverUUID string) (*WSCommandResponse, error) {
	return ws.sendCommand(ctx, serverUUID, "check_updates", map[string]interface{}{})
}

// RestartAgent restarts the agent service
func (ws *WebSocketServiceImpl) RestartAgent(ctx context.Context, serverUUID string, req *RestartRequest) (*WSCommandResponse, error) {
	if req == nil {
		req = &RestartRequest{}
	}
	return ws.sendCommand(ctx, serverUUID, "restart_agent", req)
}

// GracefulRestart performs a graceful restart of the agent
func (ws *WebSocketServiceImpl) GracefulRestart(ctx context.Context, serverUUID string, req *RestartRequest) (*WSCommandResponse, error) {
	if req == nil {
		req = &RestartRequest{Graceful: true}
	} else {
		req.Graceful = true
	}
	return ws.sendCommand(ctx, serverUUID, "graceful_restart", req)
}

// AgentHealth requests agent health status
func (ws *WebSocketServiceImpl) AgentHealth(ctx context.Context, serverUUID string) (*WSCommandResponse, error) {
	return ws.sendCommand(ctx, serverUUID, "agent_health", map[string]interface{}{})
}

// SystemStatus requests system status information
func (ws *WebSocketServiceImpl) SystemStatus(ctx context.Context, serverUUID string) (*WSCommandResponse, error) {
	return ws.sendCommand(ctx, serverUUID, "system_status", map[string]interface{}{})
}

// =============================================================================
// Event Handlers
// =============================================================================

// OnConnect sets the connection callback
func (ws *WebSocketServiceImpl) OnConnect(fn func()) {
	ws.onConnect = fn
}

// OnDisconnect sets the disconnection callback
func (ws *WebSocketServiceImpl) OnDisconnect(fn func(error)) {
	ws.onDisconnect = fn
}

// OnMessage sets the message callback for handling non-command messages
func (ws *WebSocketServiceImpl) OnMessage(fn func(*WSMessage)) {
	ws.onMessage = fn
}

// SetTimeout configures the command timeout duration
func (ws *WebSocketServiceImpl) SetTimeout(timeout time.Duration) {
	ws.timeout = timeout
}

// SetReconnectDelay configures the delay between reconnection attempts
func (ws *WebSocketServiceImpl) SetReconnectDelay(delay time.Duration) {
	ws.reconnectDelay = delay
}

// SetMaxReconnects configures the maximum number of reconnection attempts
func (ws *WebSocketServiceImpl) SetMaxReconnects(max int) {
	ws.maxReconnects = max
}

// =============================================================================
// Private Methods
// =============================================================================

// buildWebSocketURL constructs the WebSocket URL from the base URL
func (ws *WebSocketServiceImpl) buildWebSocketURL() string {
	baseURL := ws.client.config.BaseURL
	if len(baseURL) > 4 && baseURL[:4] == "http" {
		if baseURL[:5] == "https" {
			baseURL = "wss" + baseURL[5:]
		} else {
			baseURL = "ws" + baseURL[4:]
		}
	}
	return fmt.Sprintf("%s/v1/agent/websocket", baseURL)
}

// authenticate sends authentication message to the WebSocket server
func (ws *WebSocketServiceImpl) authenticate() error {
	authPayload := WSAuthPayload{
		ServerUUID:      ws.client.config.Auth.ServerUUID,
		ServerSecret:    ws.client.config.Auth.ServerSecret,
		AgentVersion:    "sdk-" + Version,
		ProtocolVersion: WSProtocolVersion, // "1.0"
		Capabilities:    []string{"commands", "responses"},
		// OrganizationID is omitted (0 value) - determined server-side
	}

	payloadBytes, err := json.Marshal(authPayload)
	if err != nil {
		return err
	}

	msg := WSMessage{
		Type:      WSTypeAuth,
		Timestamp: time.Now().Unix(),
		Payload:   payloadBytes,
	}

	if err := ws.conn.WriteJSON(msg); err != nil {
		return err
	}

	// Wait for auth response
	ws.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	var response WSMessage
	if err := ws.conn.ReadJSON(&response); err != nil {
		return err
	}

	if response.Type != WSTypeAuthResponse {
		return fmt.Errorf("expected auth response, got %s", response.Type)
	}

	var authResp WSAuthResponsePayload
	if err := json.Unmarshal(response.Payload, &authResp); err != nil {
		return err
	}

	if authResp.Status != "success" {
		return fmt.Errorf("authentication failed: %s", authResp.Error)
	}

	// Clear read deadline
	ws.conn.SetReadDeadline(time.Time{})
	return nil
}

// sendCommand sends a command and waits for response with correlation ID
func (ws *WebSocketServiceImpl) sendCommand(ctx context.Context, serverUUID, command string, payload interface{}) (*WSCommandResponse, error) {
	if !ws.connected {
		return nil, fmt.Errorf("not connected to WebSocket")
	}

	// Generate correlation ID
	correlationID := generateCorrelationID()

	// Create response channel
	responseChan := make(chan *WSCommandResponse, 1)
	
	// Store pending response
	ws.responseMutex.Lock()
	ws.pendingResponses[correlationID] = responseChan
	ws.responseMutex.Unlock()

	// Clean up on exit
	defer func() {
		ws.responseMutex.Lock()
		delete(ws.pendingResponses, correlationID)
		ws.responseMutex.Unlock()
		close(responseChan)
	}()

	// Marshal command payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Create command payload
	commandPayload := WSCommandPayload{
		Command: command,
		Payload: payloadBytes,
	}

	cmdPayloadBytes, err := json.Marshal(commandPayload)
	if err != nil {
		return nil, err
	}

	// Send command message
	msg := WSMessage{
		Type:      WSTypeCommand,
		ID:        correlationID,
		Timestamp: time.Now().Unix(),
		Payload:   cmdPayloadBytes,
	}

	ws.mu.RLock()
	if ws.conn == nil {
		ws.mu.RUnlock()
		return nil, fmt.Errorf("connection closed")
	}
	err = ws.conn.WriteJSON(msg)
	ws.mu.RUnlock()

	if err != nil {
		return nil, err
	}

	// Wait for response
	select {
	case response := <-responseChan:
		return response, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(ws.timeout):
		return nil, fmt.Errorf("command timeout after %v", ws.timeout)
	}
}

// handleMessages processes incoming WebSocket messages
func (ws *WebSocketServiceImpl) handleMessages() {
	defer func() {
		ws.mu.Lock()
		ws.connected = false
		ws.mu.Unlock()
	}()

	for {
		select {
		case <-ws.ctx.Done():
			return
		default:
			var msg WSMessage
			if err := ws.conn.ReadJSON(&msg); err != nil {
				if ws.onDisconnect != nil && ws.connected {
					ws.onDisconnect(err)
				}
				return
			}

			switch msg.Type {
			case WSTypeCommandResponse:
				ws.handleCommandResponse(&msg)
			case WSTypePing:
				ws.sendPong()
			case WSTypePong:
				// Pong received, connection is alive
			default:
				if ws.onMessage != nil {
					ws.onMessage(&msg)
				}
			}
		}
	}
}

// handleCommandResponse processes command response messages
func (ws *WebSocketServiceImpl) handleCommandResponse(msg *WSMessage) {
	if msg.ID == "" {
		return
	}

	ws.responseMutex.RLock()
	responseChan, exists := ws.pendingResponses[msg.ID]
	ws.responseMutex.RUnlock()

	if !exists {
		return
	}

	var response WSCommandResponse
	if err := json.Unmarshal(msg.Payload, &response); err != nil {
		response = WSCommandResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse response: %v", err),
		}
	}

	select {
	case responseChan <- &response:
	default:
		// Channel is full or closed
	}
}

// pingHandler sends periodic ping messages
func (ws *WebSocketServiceImpl) pingHandler() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ws.ctx.Done():
			return
		case <-ticker.C:
			if err := ws.sendPing(); err != nil {
				return
			}
		}
	}
}

// sendPing sends a ping message
func (ws *WebSocketServiceImpl) sendPing() error {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	if !ws.connected || ws.conn == nil {
		return fmt.Errorf("not connected")
	}

	msg := WSMessage{
		Type:      WSTypePing,
		Timestamp: time.Now().Unix(),
	}

	return ws.conn.WriteJSON(msg)
}

// sendPong sends a pong message
func (ws *WebSocketServiceImpl) sendPong() error {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	if !ws.connected || ws.conn == nil {
		return fmt.Errorf("not connected")
	}

	msg := WSMessage{
		Type:      WSTypePong,
		Timestamp: time.Now().Unix(),
	}

	return ws.conn.WriteJSON(msg)
}

// generateCorrelationID generates a unique correlation ID for commands
func generateCorrelationID() string {
	return fmt.Sprintf("sdk-%d-%d", time.Now().UnixNano(), time.Now().Nanosecond()%1000)
}