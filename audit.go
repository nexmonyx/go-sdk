package nexmonyx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// AuditService handles audit log operations and compliance tracking
type AuditService struct {
	client *Client
}

// GetAuditLogs retrieves audit logs with comprehensive filtering
// Authentication: JWT Token required
// Endpoint: GET /v1/audit/logs
// Parameters:
//   - opts: Optional pagination options
//   - filters: Optional filters (user_id, action, resource_type, start_date, end_date, severity)
// Returns: Array of AuditLog objects with pagination metadata
func (s *AuditService) GetAuditLogs(ctx context.Context, opts *PaginationOptions, filters map[string]interface{}) ([]AuditLog, *PaginationMeta, error) {
	var resp struct {
		Data []AuditLog      `json:"data"`
		Meta *PaginationMeta `json:"meta"`
	}

	queryParams := make(map[string]string)
	if opts != nil {
		if opts.Page > 0 {
			queryParams["page"] = fmt.Sprintf("%d", opts.Page)
		}
		if opts.Limit > 0 {
			queryParams["limit"] = fmt.Sprintf("%d", opts.Limit)
		}
	}

	// Add filter parameters
	if filters != nil {
		if userID, ok := filters["user_id"].(uint); ok && userID > 0 {
			queryParams["user_id"] = fmt.Sprintf("%d", userID)
		}
		if action, ok := filters["action"].(string); ok && action != "" {
			queryParams["action"] = action
		}
		if resourceType, ok := filters["resource_type"].(string); ok && resourceType != "" {
			queryParams["resource_type"] = resourceType
		}
		if resourceID, ok := filters["resource_id"].(string); ok && resourceID != "" {
			queryParams["resource_id"] = resourceID
		}
		if startDate, ok := filters["start_date"].(string); ok && startDate != "" {
			queryParams["start_date"] = startDate
		}
		if endDate, ok := filters["end_date"].(string); ok && endDate != "" {
			queryParams["end_date"] = endDate
		}
		if severity, ok := filters["severity"].(string); ok && severity != "" {
			queryParams["severity"] = severity
		}
		if ipAddress, ok := filters["ip_address"].(string); ok && ipAddress != "" {
			queryParams["ip_address"] = ipAddress
		}
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/audit/logs",
		Result: &resp,
	}
	if len(queryParams) > 0 {
		req.Query = queryParams
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return resp.Data, resp.Meta, nil
}

// GetAuditLog retrieves a specific audit log entry by ID
// Authentication: JWT Token required
// Endpoint: GET /v1/audit/logs/{id}
// Parameters:
//   - id: Audit log ID
// Returns: AuditLog object with full details
func (s *AuditService) GetAuditLog(ctx context.Context, id uint) (*AuditLog, error) {
	var resp struct {
		Data    *AuditLog `json:"data"`
		Status  string    `json:"status"`
		Message string    `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/audit/logs/%d", id),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// ExportAuditLogs exports audit logs in the specified format
// Authentication: JWT Token required
// Endpoint: POST /v1/audit/logs/export
// Parameters:
//   - format: Export format (csv, json, pdf)
//   - filters: Optional filters (same as GetAuditLogs)
// Returns: Exported audit logs as byte array
func (s *AuditService) ExportAuditLogs(ctx context.Context, format string, filters map[string]interface{}) ([]byte, error) {
	body := map[string]interface{}{
		"format": format,
	}

	// Add filters to request body
	if filters != nil {
		if userID, ok := filters["user_id"].(uint); ok && userID > 0 {
			body["user_id"] = userID
		}
		if action, ok := filters["action"].(string); ok && action != "" {
			body["action"] = action
		}
		if resourceType, ok := filters["resource_type"].(string); ok && resourceType != "" {
			body["resource_type"] = resourceType
		}
		if startDate, ok := filters["start_date"].(string); ok && startDate != "" {
			body["start_date"] = startDate
		}
		if endDate, ok := filters["end_date"].(string); ok && endDate != "" {
			body["end_date"] = endDate
		}
		if severity, ok := filters["severity"].(string); ok && severity != "" {
			body["severity"] = severity
		}
	}

	resp, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/audit/logs/export",
		Body:   body,
	})
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

// GetAuditStatistics retrieves comprehensive audit activity statistics
// Authentication: JWT Token required
// Endpoint: GET /v1/audit/statistics
// Parameters:
//   - startDate: Optional start date filter (ISO 8601 format)
//   - endDate: Optional end date filter (ISO 8601 format)
// Returns: AuditStatistics object with activity breakdown
func (s *AuditService) GetAuditStatistics(ctx context.Context, startDate string, endDate string) (*AuditStatistics, error) {
	var resp struct {
		Data    *AuditStatistics `json:"data"`
		Status  string           `json:"status"`
		Message string           `json:"message"`
	}

	queryParams := make(map[string]string)
	if startDate != "" {
		queryParams["start_date"] = startDate
	}
	if endDate != "" {
		queryParams["end_date"] = endDate
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/audit/statistics",
		Result: &resp,
	}
	if len(queryParams) > 0 {
		req.Query = queryParams
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// GetUserAuditHistory retrieves audit history for a specific user
// Authentication: JWT Token required
// Endpoint: GET /v1/audit/users/{userID}/history
// Parameters:
//   - userID: User ID
//   - opts: Optional pagination options
//   - startDate: Optional start date filter
//   - endDate: Optional end date filter
// Returns: Array of AuditLog objects with pagination metadata
func (s *AuditService) GetUserAuditHistory(ctx context.Context, userID uint, opts *PaginationOptions, startDate string, endDate string) ([]AuditLog, *PaginationMeta, error) {
	var resp struct {
		Data []AuditLog      `json:"data"`
		Meta *PaginationMeta `json:"meta"`
	}

	queryParams := make(map[string]string)
	if opts != nil {
		if opts.Page > 0 {
			queryParams["page"] = fmt.Sprintf("%d", opts.Page)
		}
		if opts.Limit > 0 {
			queryParams["limit"] = fmt.Sprintf("%d", opts.Limit)
		}
	}
	if startDate != "" {
		queryParams["start_date"] = startDate
	}
	if endDate != "" {
		queryParams["end_date"] = endDate
	}

	req := &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/audit/users/%d/history", userID),
		Result: &resp,
	}
	if len(queryParams) > 0 {
		req.Query = queryParams
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return resp.Data, resp.Meta, nil
}

// === Command Audit Methods ===

// RecordSecurityEvent logs a security event for command auditing
// Authentication: JWT Token required
// Endpoint: POST /v1/audit/commands/security-event
// Parameters:
//   - req: SecurityEventRequest with event details
//
// Returns: SecurityEventResponse with event ID and status
func (s *AuditService) RecordSecurityEvent(ctx context.Context, req *SecurityEventRequest) (*SecurityEventResponse, error) {
	var resp struct {
		Data    *SecurityEventResponse `json:"data"`
		Status  string                 `json:"status"`
		Message string                 `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/audit/commands/security-event",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// GetCommandAnalytics retrieves command execution analytics
// Authentication: JWT Token required
// Endpoint: GET /v1/audit/commands/analytics
// Parameters:
//   - filters: Optional CommandAnalyticsFilter for filtering results
//
// Returns: CommandAnalytics with execution stats, performance metrics, and trends
func (s *AuditService) GetCommandAnalytics(ctx context.Context, filters *CommandAnalyticsFilter) (*CommandAnalytics, error) {
	var resp struct {
		Data      *CommandAnalytics `json:"data,omitempty"`
		Analytics *CommandAnalytics `json:"analytics,omitempty"`
		Status    string            `json:"status"`
		Message   string            `json:"message"`
	}

	queryParams := make(map[string]string)
	if filters != nil {
		if len(filters.ServerUUIDs) > 0 {
			queryParams["server_uuids"] = strings.Join(filters.ServerUUIDs, ",")
		}
		if len(filters.CommandNames) > 0 {
			queryParams["command_names"] = strings.Join(filters.CommandNames, ",")
		}
		if filters.StartTime != "" {
			queryParams["start_time"] = filters.StartTime
		}
		if filters.EndTime != "" {
			queryParams["end_time"] = filters.EndTime
		}
		if filters.TimeRange != "" {
			queryParams["time_range"] = filters.TimeRange
		}
		if filters.Success != nil {
			queryParams["success"] = fmt.Sprintf("%t", *filters.Success)
		}
		if filters.MinDuration > 0 {
			queryParams["min_duration_ms"] = fmt.Sprintf("%d", filters.MinDuration)
		}
		if filters.MaxDuration > 0 {
			queryParams["max_duration_ms"] = fmt.Sprintf("%d", filters.MaxDuration)
		}
	}

	analyticsReq := &Request{
		Method: "GET",
		Path:   "/v1/audit/commands/analytics",
		Result: &resp,
	}
	if len(queryParams) > 0 {
		analyticsReq.Query = queryParams
	}

	_, err := s.client.Do(ctx, analyticsReq)
	if err != nil {
		return nil, err
	}

	// Handle different response structures
	if resp.Analytics != nil {
		return resp.Analytics, nil
	}
	return resp.Data, nil
}

// GetCommandHistory retrieves command execution history with filtering
// Authentication: JWT Token required
// Endpoint: GET /v1/audit/commands/history
// Parameters:
//   - filters: Optional CommandHistoryFilter for filtering results
//   - opts: Optional pagination options
//
// Returns: Array of CommandAuditEntry, summary statistics, and pagination metadata
func (s *AuditService) GetCommandHistory(ctx context.Context, filters *CommandHistoryFilter, opts *PaginationOptions) ([]CommandAuditEntry, *CommandAuditHistorySummary, *PaginationMeta, error) {
	var resp struct {
		Data       []CommandAuditEntry         `json:"data,omitempty"`
		Commands   []CommandAuditEntry         `json:"commands,omitempty"`
		Summary    *CommandAuditHistorySummary `json:"summary,omitempty"`
		Pagination *PaginationMeta             `json:"pagination,omitempty"`
		Meta       *PaginationMeta             `json:"meta,omitempty"`
	}

	queryParams := make(map[string]string)
	if opts != nil {
		if opts.Page > 0 {
			queryParams["page"] = fmt.Sprintf("%d", opts.Page)
		}
		if opts.Limit > 0 {
			queryParams["limit"] = fmt.Sprintf("%d", opts.Limit)
		}
	}
	if filters != nil {
		if len(filters.ServerUUIDs) > 0 {
			queryParams["server_uuids"] = strings.Join(filters.ServerUUIDs, ",")
		}
		if len(filters.CommandNames) > 0 {
			queryParams["command_names"] = strings.Join(filters.CommandNames, ",")
		}
		if filters.StartTime != "" {
			queryParams["start_time"] = filters.StartTime
		}
		if filters.EndTime != "" {
			queryParams["end_time"] = filters.EndTime
		}
		if filters.Success != nil {
			queryParams["success"] = fmt.Sprintf("%t", *filters.Success)
		}
		if filters.CorrelationID != "" {
			queryParams["correlation_id"] = filters.CorrelationID
		}
		if filters.IncludePayload {
			queryParams["include_payload"] = "true"
		}
		if filters.IncludeResponse {
			queryParams["include_response"] = "true"
		}
		if filters.SearchText != "" {
			queryParams["search"] = filters.SearchText
		}
		if filters.OrderBy != "" {
			queryParams["order_by"] = filters.OrderBy
		}
		if filters.OrderDirection != "" {
			queryParams["order_direction"] = filters.OrderDirection
		}
	}

	historyReq := &Request{
		Method: "GET",
		Path:   "/v1/audit/commands/history",
		Result: &resp,
	}
	if len(queryParams) > 0 {
		historyReq.Query = queryParams
	}

	_, err := s.client.Do(ctx, historyReq)
	if err != nil {
		return nil, nil, nil, err
	}

	// Handle different response structures
	data := resp.Data
	if len(resp.Commands) > 0 {
		data = resp.Commands
	}
	meta := resp.Meta
	if resp.Pagination != nil {
		meta = resp.Pagination
	}

	return data, resp.Summary, meta, nil
}

// ExportCommandAudit exports command audit logs in the specified format
// Authentication: JWT Token required
// Endpoint: GET /v1/audit/commands/export
// Parameters:
//   - filters: Optional CommandExportFilter for filtering results
//   - format: Export format (csv, json)
//
// Returns: Raw bytes of the exported file (CSV or JSON content)
func (s *AuditService) ExportCommandAudit(ctx context.Context, filters *CommandExportFilter, format string) ([]byte, error) {
	queryParams := make(map[string]string)
	queryParams["format"] = format

	if filters != nil {
		if len(filters.ServerUUIDs) > 0 {
			queryParams["server_uuids"] = strings.Join(filters.ServerUUIDs, ",")
		}
		if len(filters.CommandNames) > 0 {
			queryParams["command_names"] = strings.Join(filters.CommandNames, ",")
		}
		if filters.StartTime != "" {
			queryParams["start_time"] = filters.StartTime
		}
		if filters.EndTime != "" {
			queryParams["end_time"] = filters.EndTime
		}
		if filters.Success != nil {
			queryParams["success"] = fmt.Sprintf("%t", *filters.Success)
		}
		if filters.IncludePayload {
			queryParams["include_payload"] = "true"
		}
		if filters.IncludeResponse {
			queryParams["include_response"] = "true"
		}
		if filters.TimeZone != "" {
			queryParams["timezone"] = filters.TimeZone
		}
	}

	resp, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/audit/commands/export",
		Query:  queryParams,
	})
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

// === Real-Time Audit Monitoring (WebSocket) ===

// AuditMonitorConnection represents a WebSocket connection for real-time audit monitoring
type AuditMonitorConnection struct {
	conn         *websocket.Conn
	client       *Client
	mu           sync.RWMutex
	connected    bool
	subscriberID string
	filters      *AuditMonitoringFilters

	// Channels for events and errors
	Events chan *AuditMonitoringEvent
	Stats  chan *AuditMonitoringStats
	Errors chan error

	// Control channels
	done     chan struct{}
	stopOnce sync.Once
}

// ConnectAuditMonitor establishes a WebSocket connection for real-time audit event monitoring
// Authentication: JWT Token required
// Endpoint: GET /v1/audit/commands/monitor/ws (WebSocket upgrade)
// Parameters:
//   - ctx: Context for connection lifecycle
//   - filters: Optional initial filters for events
//
// Returns: AuditMonitorConnection for receiving events, or error if connection fails
//
// Usage:
//
//	conn, err := client.Audit.ConnectAuditMonitor(ctx, &nexmonyx.AuditMonitoringFilters{
//	    EventTypes: []string{"command_failed", "security_alert"},
//	})
//	if err != nil { return err }
//	defer conn.Close()
//
//	for event := range conn.Events {
//	    fmt.Printf("Event: %s - %s\n", event.Type, event.CommandName)
//	}
func (s *AuditService) ConnectAuditMonitor(ctx context.Context, filters *AuditMonitoringFilters) (*AuditMonitorConnection, error) {
	// Build WebSocket URL from client base URL
	baseURL := s.client.config.BaseURL
	wsURL := strings.Replace(baseURL, "https://", "wss://", 1)
	wsURL = strings.Replace(wsURL, "http://", "ws://", 1)
	wsURL = strings.TrimSuffix(wsURL, "/") + "/v1/audit/commands/monitor/ws"

	// Set up headers with authentication
	headers := http.Header{}
	if s.client.config.Auth.Token != "" {
		headers.Set("Authorization", "Bearer "+s.client.config.Auth.Token)
	} else if s.client.config.Auth.UnifiedAPIKey != "" {
		headers.Set("Authorization", "Bearer "+s.client.config.Auth.UnifiedAPIKey)
	} else if s.client.config.Auth.APIKey != "" {
		// Use X-API-Key header for API key authentication
		headers.Set("X-API-Key", s.client.config.Auth.APIKey)
		if s.client.config.Auth.APISecret != "" {
			headers.Set("X-API-Secret", s.client.config.Auth.APISecret)
		}
	}

	// Dial WebSocket connection
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, resp, err := dialer.DialContext(ctx, wsURL, headers)
	if err != nil {
		if resp != nil {
			return nil, fmt.Errorf("websocket dial failed with status %d: %w", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("websocket dial failed: %w", err)
	}

	// Create connection wrapper
	amc := &AuditMonitorConnection{
		conn:    conn,
		client:  s.client,
		filters: filters,
		Events:  make(chan *AuditMonitoringEvent, 100),
		Stats:   make(chan *AuditMonitoringStats, 10),
		Errors:  make(chan error, 10),
		done:    make(chan struct{}),
	}

	// Start message reader goroutine
	go amc.readLoop()

	// Wait for welcome message with timeout
	welcomeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	select {
	case <-welcomeCtx.Done():
		amc.Close()
		return nil, fmt.Errorf("timeout waiting for welcome message")
	case err := <-amc.Errors:
		amc.Close()
		return nil, fmt.Errorf("error during connection setup: %w", err)
	case <-time.After(100 * time.Millisecond):
		// Brief delay to allow welcome message processing
	}

	// Apply initial filters if provided
	if filters != nil {
		if err := amc.SetFilters(filters); err != nil {
			amc.Close()
			return nil, fmt.Errorf("failed to set initial filters: %w", err)
		}
	}

	amc.mu.Lock()
	amc.connected = true
	amc.mu.Unlock()

	return amc, nil
}

// readLoop continuously reads messages from the WebSocket connection
func (amc *AuditMonitorConnection) readLoop() {
	defer func() {
		amc.mu.Lock()
		amc.connected = false
		amc.mu.Unlock()
		close(amc.Events)
		close(amc.Stats)
	}()

	for {
		select {
		case <-amc.done:
			return
		default:
		}

		_, data, err := amc.conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				return
			}
			select {
			case amc.Errors <- err:
			default:
			}
			return
		}

		var msg AuditMonitorMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			select {
			case amc.Errors <- fmt.Errorf("failed to parse message: %w", err):
			default:
			}
			continue
		}

		amc.handleMessage(&msg)
	}
}

// handleMessage processes incoming WebSocket messages
func (amc *AuditMonitorConnection) handleMessage(msg *AuditMonitorMessage) {
	switch msg.Type {
	case "welcome":
		amc.mu.Lock()
		amc.subscriberID = msg.SubscriberID
		amc.mu.Unlock()

	case "stats":
		if msg.Data != nil {
			select {
			case amc.Stats <- msg.Data:
			default:
				// Stats channel full, drop oldest
			}
		}

	case "event":
		if msg.Event != nil {
			select {
			case amc.Events <- msg.Event:
			default:
				// Events channel full, drop oldest
			}
		}

	case "filters_updated":
		if msg.Filters != nil {
			amc.mu.Lock()
			amc.filters = msg.Filters
			amc.mu.Unlock()
		}

	case "pong":
		// Pong received, connection is alive

	case "error":
		select {
		case amc.Errors <- fmt.Errorf("server error: %s", msg.Error):
		default:
		}
	}
}

// SetFilters updates the event filters for the connection
func (amc *AuditMonitorConnection) SetFilters(filters *AuditMonitoringFilters) error {
	amc.mu.RLock()
	if !amc.connected && amc.conn == nil {
		amc.mu.RUnlock()
		return fmt.Errorf("connection not established")
	}
	amc.mu.RUnlock()

	msg := map[string]interface{}{
		"type":    "set_filters",
		"filters": filters,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal filters: %w", err)
	}

	amc.mu.Lock()
	defer amc.mu.Unlock()

	if err := amc.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("failed to send filters: %w", err)
	}

	return nil
}

// GetStats requests current monitoring statistics from the server
func (amc *AuditMonitorConnection) GetStats() error {
	amc.mu.RLock()
	if !amc.connected {
		amc.mu.RUnlock()
		return fmt.Errorf("connection not established")
	}
	amc.mu.RUnlock()

	msg := map[string]interface{}{
		"type": "get_stats",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	amc.mu.Lock()
	defer amc.mu.Unlock()

	if err := amc.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	return nil
}

// Ping sends a ping message to keep the connection alive
func (amc *AuditMonitorConnection) Ping() error {
	amc.mu.RLock()
	if !amc.connected {
		amc.mu.RUnlock()
		return fmt.Errorf("connection not established")
	}
	amc.mu.RUnlock()

	msg := map[string]interface{}{
		"type": "ping",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal ping: %w", err)
	}

	amc.mu.Lock()
	defer amc.mu.Unlock()

	if err := amc.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("failed to send ping: %w", err)
	}

	return nil
}

// IsConnected returns whether the connection is currently active
func (amc *AuditMonitorConnection) IsConnected() bool {
	amc.mu.RLock()
	defer amc.mu.RUnlock()
	return amc.connected
}

// SubscriberID returns the unique subscriber ID assigned by the server
func (amc *AuditMonitorConnection) SubscriberID() string {
	amc.mu.RLock()
	defer amc.mu.RUnlock()
	return amc.subscriberID
}

// CurrentFilters returns the currently active filters
func (amc *AuditMonitorConnection) CurrentFilters() *AuditMonitoringFilters {
	amc.mu.RLock()
	defer amc.mu.RUnlock()
	return amc.filters
}

// Close gracefully closes the WebSocket connection
func (amc *AuditMonitorConnection) Close() error {
	var closeErr error
	amc.stopOnce.Do(func() {
		close(amc.done)

		amc.mu.Lock()
		amc.connected = false
		if amc.conn != nil {
			// Send close message
			closeErr = amc.conn.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			)
			amc.conn.Close()
		}
		amc.mu.Unlock()
	})
	return closeErr
}
