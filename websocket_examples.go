package nexmonyx

import (
	"context"
	"fmt"
	"log"
	"time"
)

// Example: Basic WebSocket connection and command execution
func ExampleWebSocketService_basic() {
	// Create client with server credentials
	config := &Config{
		BaseURL: "https://api.nexmonyx.com",
		Auth: AuthConfig{
			ServerUUID:   "your-server-uuid",
			ServerSecret: "your-server-secret",
		},
	}

	client, err := NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Initialize WebSocket service
	wsService, err := client.NewWebSocketService()
	if err != nil {
		log.Fatalf("Failed to create WebSocket service: %v", err)
	}

	// Set up event handlers
	wsService.OnConnect(func() {
		fmt.Println("WebSocket connected successfully")
	})

	wsService.OnDisconnect(func(err error) {
		fmt.Printf("WebSocket disconnected: %v\n", err)
	})

	// Connect to WebSocket
	if err := wsService.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer wsService.Disconnect()

	// Send a command
	ctx := context.Background()
	serverUUID := "target-server-uuid"

	response, err := wsService.AgentHealth(ctx, serverUUID)
	if err != nil {
		log.Fatalf("Command failed: %v", err)
	}

	if response.Success {
		fmt.Printf("Agent health check successful: %s\n", string(response.Data))
	} else {
		fmt.Printf("Agent health check failed: %s\n", response.Error)
	}
}

// Example: Running metrics collection
func ExampleWebSocketService_RunCollection() {
	config := &Config{
		BaseURL: "https://api.nexmonyx.com",
		Auth: AuthConfig{
			ServerUUID:   "your-server-uuid",
			ServerSecret: "your-server-secret",
		},
	}

	client, err := NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	wsService, err := client.NewWebSocketService()
	if err != nil {
		log.Fatalf("Failed to create WebSocket service: %v", err)
	}

	if err := wsService.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer wsService.Disconnect()

	// Run collection with specific parameters
	collectionReq := &CollectionRequest{
		CollectorTypes: []string{"cpu", "memory", "network"},
		Comprehensive:  false,
		Timeout:        30,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	response, err := wsService.RunCollection(ctx, "target-server-uuid", collectionReq)
	if err != nil {
		log.Fatalf("Collection failed: %v", err)
	}

	fmt.Printf("Collection result: success=%v, data=%s\n", 
		response.Success, string(response.Data))
}

// Example: Force collection (immediate comprehensive metrics)
func ExampleWebSocketService_ForceCollection() {
	config := &Config{
		BaseURL: "https://api.nexmonyx.com",
		Auth: AuthConfig{
			ServerUUID:   "your-server-uuid",
			ServerSecret: "your-server-secret",
		},
	}

	client, err := NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	wsService, err := client.NewWebSocketService()
	if err != nil {
		log.Fatalf("Failed to create WebSocket service: %v", err)
	}

	if err := wsService.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer wsService.Disconnect()

	// Force comprehensive collection
	collectionReq := &CollectionRequest{
		CollectorTypes: []string{"all"}, // Collect all available metrics
		Timeout:        60,              // Allow more time for comprehensive collection
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	response, err := wsService.ForceCollection(ctx, "target-server-uuid", collectionReq)
	if err != nil {
		log.Fatalf("Force collection failed: %v", err)
	}

	fmt.Printf("Force collection result: success=%v\n", response.Success)
	if response.Metadata != nil {
		if execTime, ok := response.Metadata["execution_time_ms"].(float64); ok {
			fmt.Printf("Execution time: %.0fms\n", execTime)
		}
	}
}

// Example: Agent update
func ExampleWebSocketService_UpdateAgent() {
	config := &Config{
		BaseURL: "https://api.nexmonyx.com",
		Auth: AuthConfig{
			ServerUUID:   "your-server-uuid",
			ServerSecret: "your-server-secret",
		},
	}

	client, err := NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	wsService, err := client.NewWebSocketService()
	if err != nil {
		log.Fatalf("Failed to create WebSocket service: %v", err)
	}

	if err := wsService.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer wsService.Disconnect()

	// Update agent to specific version
	updateReq := &UpdateRequest{
		Version:   "2.1.5",
		Force:     false,
		Immediate: false,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	response, err := wsService.UpdateAgent(ctx, "target-server-uuid", updateReq)
	if err != nil {
		log.Fatalf("Agent update failed: %v", err)
	}

	fmt.Printf("Agent update result: success=%v\n", response.Success)
	if !response.Success {
		fmt.Printf("Update error: %s\n", response.Error)
	}
}

// Example: Graceful agent restart
func ExampleWebSocketService_GracefulRestart() {
	config := &Config{
		BaseURL: "https://api.nexmonyx.com",
		Auth: AuthConfig{
			ServerUUID:   "your-server-uuid",
			ServerSecret: "your-server-secret",
		},
	}

	client, err := NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	wsService, err := client.NewWebSocketService()
	if err != nil {
		log.Fatalf("Failed to create WebSocket service: %v", err)
	}

	if err := wsService.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer wsService.Disconnect()

	// Graceful restart with delay
	restartReq := &RestartRequest{
		Delay:  5,                         // 5 second delay
		Reason: "Scheduled maintenance",   // Reason for restart
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := wsService.GracefulRestart(ctx, "target-server-uuid", restartReq)
	if err != nil {
		log.Fatalf("Graceful restart failed: %v", err)
	}

	fmt.Printf("Graceful restart initiated: success=%v\n", response.Success)
}

// Example: System status check
func ExampleWebSocketService_SystemStatus() {
	config := &Config{
		BaseURL: "https://api.nexmonyx.com",
		Auth: AuthConfig{
			ServerUUID:   "your-server-uuid",
			ServerSecret: "your-server-secret",
		},
	}

	client, err := NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	wsService, err := client.NewWebSocketService()
	if err != nil {
		log.Fatalf("Failed to create WebSocket service: %v", err)
	}

	if err := wsService.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer wsService.Disconnect()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := wsService.SystemStatus(ctx, "target-server-uuid")
	if err != nil {
		log.Fatalf("System status failed: %v", err)
	}

	if response.Success {
		fmt.Printf("System status: %s\n", string(response.Data))
	} else {
		fmt.Printf("System status error: %s\n", response.Error)
	}
}

// Example: Batch operations with multiple servers
func ExampleWebSocketService_batch() {
	config := &Config{
		BaseURL: "https://api.nexmonyx.com",
		Auth: AuthConfig{
			ServerUUID:   "your-server-uuid",
			ServerSecret: "your-server-secret",
		},
	}

	client, err := NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	wsService, err := client.NewWebSocketService()
	if err != nil {
		log.Fatalf("Failed to create WebSocket service: %v", err)
	}

	if err := wsService.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer wsService.Disconnect()

	// List of target servers
	servers := []string{
		"server-uuid-1",
		"server-uuid-2",
		"server-uuid-3",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Run health checks on all servers
	results := make(map[string]*WSCommandResponse)
	for _, serverUUID := range servers {
		response, err := wsService.AgentHealth(ctx, serverUUID)
		if err != nil {
			fmt.Printf("Health check failed for %s: %v\n", serverUUID, err)
			continue
		}
		results[serverUUID] = response
	}

	// Print results
	for serverUUID, response := range results {
		status := "FAILED"
		if response.Success {
			status = "OK"
		}
		fmt.Printf("Server %s: %s\n", serverUUID, status)
	}
}

// Example: Advanced usage with custom message handling
func ExampleWebSocketService_advanced() {
	config := &Config{
		BaseURL: "https://api.nexmonyx.com",
		Auth: AuthConfig{
			ServerUUID:   "your-server-uuid",
			ServerSecret: "your-server-secret",
		},
		Debug: true, // Enable debug logging
	}

	client, err := NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	wsService, err := client.NewWebSocketService()
	if err != nil {
		log.Fatalf("Failed to create WebSocket service: %v", err)
	}

	// Set up comprehensive event handlers
	wsService.OnConnect(func() {
		fmt.Println("🔌 WebSocket connected successfully")
	})

	wsService.OnDisconnect(func(err error) {
		if err != nil {
			fmt.Printf("🔌 WebSocket disconnected with error: %v\n", err)
		} else {
			fmt.Println("🔌 WebSocket disconnected gracefully")
		}
	})

	wsService.OnMessage(func(msg *WSMessage) {
		switch msg.Type {
		case WSTypeUpdateProgress:
			fmt.Printf("📦 Update progress: %s\n", string(msg.Payload))
		case WSTypeError:
			fmt.Printf("❌ WebSocket error: %s\n", string(msg.Payload))
		default:
			fmt.Printf("📨 Received message: type=%s, id=%s\n", msg.Type, msg.ID)
		}
	})

	if err := wsService.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer wsService.Disconnect()

	// Keep connection alive and handle messages
	fmt.Println("WebSocket service is running. Press Ctrl+C to exit.")
	
	// In a real application, you would handle shutdown signals properly
	time.Sleep(60 * time.Second)
}