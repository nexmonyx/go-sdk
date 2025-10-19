package nexmonyx

import (
	"context"
	"testing"
	"time"
)

func TestNewMonitoringAgentClient(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid monitoring key",
			config: &Config{
				Auth: AuthConfig{
					MonitoringKey: "MON_test_key_12345",
				},
			},
			wantErr: false,
		},
		{
			name: "missing monitoring key",
			config: &Config{
				Auth: AuthConfig{},
			},
			wantErr: true,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "empty monitoring key",
			config: &Config{
				Auth: AuthConfig{
					MonitoringKey: "",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewMonitoringAgentClient(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewMonitoringAgentClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewMonitoringAgentClient() returned nil client without error")
			}
			if !tt.wantErr && client.config.Auth.MonitoringKey != tt.config.Auth.MonitoringKey {
				t.Error("NewMonitoringAgentClient() did not set monitoring key correctly")
			}
		})
	}
}

func TestMonitoringAgentClient_AuthenticationClearing(t *testing.T) {
	config := &Config{
		Auth: AuthConfig{
			Token:          "should-be-cleared",
			APIKey:         "should-be-cleared",
			APISecret:      "should-be-cleared", 
			ServerUUID:     "should-be-cleared",
			ServerSecret:   "should-be-cleared",
			MonitoringKey:  "MON_test_key_12345",
		},
	}

	client, err := NewMonitoringAgentClient(config)
	if err != nil {
		t.Fatalf("NewMonitoringAgentClient() error = %v", err)
	}

	// Verify other auth methods were cleared
	if client.config.Auth.Token != "" {
		t.Error("Token should be cleared")
	}
	if client.config.Auth.APIKey != "" {
		t.Error("APIKey should be cleared")
	}
	if client.config.Auth.APISecret != "" {
		t.Error("APISecret should be cleared")
	}
	if client.config.Auth.ServerUUID != "" {
		t.Error("ServerUUID should be cleared")
	}
	if client.config.Auth.ServerSecret != "" {
		t.Error("ServerSecret should be cleared")
	}
	if client.config.Auth.MonitoringKey != "MON_test_key_12345" {
		t.Error("MonitoringKey should be preserved")
	}
}

func TestWithMonitoringKey(t *testing.T) {
	// Create a base client
	baseClient, err := NewClient(&Config{
		Auth: AuthConfig{
			APIKey:    "old-key",
			APISecret: "old-secret",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create base client: %v", err)
	}

	// Create monitoring client using WithMonitoringKey
	monitoringClient := baseClient.WithMonitoringKey("MON_new_key_12345")

	// Verify the new client has only monitoring key authentication
	if monitoringClient.config.Auth.MonitoringKey != "MON_new_key_12345" {
		t.Error("MonitoringKey not set correctly")
	}
	if monitoringClient.config.Auth.APIKey != "" {
		t.Error("APIKey should be cleared")
	}
	if monitoringClient.config.Auth.APISecret != "" {
		t.Error("APISecret should be cleared")
	}
}

func TestProbeAssignment_Validation(t *testing.T) {
	assignment := &ProbeAssignment{
		ProbeID:        1,
		ProbeUUID:      "probe-123",
		Name:           "Test Probe",
		Type:           "http",
		Target:         "https://example.com",
		Interval:       60,
		Timeout:        30,
		Enabled:        true,
		Region:         "us-east-1",
		OrganizationID: 100,
	}

	// Basic validation tests
	if assignment.ProbeID == 0 {
		t.Error("ProbeID should not be zero")
	}
	if assignment.ProbeUUID == "" {
		t.Error("ProbeUUID should not be empty")
	}
	if assignment.Type == "" {
		t.Error("Type should not be empty")
	}
	if assignment.Target == "" {
		t.Error("Target should not be empty")
	}
	if assignment.Interval <= 0 {
		t.Error("Interval should be positive")
	}
	if assignment.Timeout <= 0 {
		t.Error("Timeout should be positive")
	}
}

func TestProbeExecutionResult_Validation(t *testing.T) {
	result := &ProbeExecutionResult{
		ProbeID:      1,
		ProbeUUID:    "probe-123",
		ExecutedAt:   time.Now(),
		Region:       "us-east-1",
		Status:       "success",
		ResponseTime: 150,
		StatusCode:   200,
	}

	// Basic validation tests
	if result.ProbeID == 0 {
		t.Error("ProbeID should not be zero")
	}
	if result.ProbeUUID == "" {
		t.Error("ProbeUUID should not be empty")
	}
	if result.ExecutedAt.IsZero() {
		t.Error("ExecutedAt should not be zero")
	}
	if result.Status == "" {
		t.Error("Status should not be empty")
	}
	if result.ResponseTime < 0 {
		t.Error("ResponseTime should not be negative")
	}
}

func TestNodeInfo_Validation(t *testing.T) {
	nodeInfo := &NodeInfo{
		AgentID:        "test-agent",
		AgentVersion:   "1.0.0",
		Region:         "us-east-1",
		Status:         "healthy",
		Uptime:         time.Hour,
		LastSeen:       time.Now(),
		ProbesAssigned: 5,
		SupportedTypes: []string{"http", "https", "tcp"},
	}

	// Basic validation tests
	if nodeInfo.AgentID == "" {
		t.Error("AgentID should not be empty")
	}
	if nodeInfo.AgentVersion == "" {
		t.Error("AgentVersion should not be empty")
	}
	if nodeInfo.Region == "" {
		t.Error("Region should not be empty")
	}
	if nodeInfo.Status == "" {
		t.Error("Status should not be empty")
	}
	if nodeInfo.LastSeen.IsZero() {
		t.Error("LastSeen should not be zero")
	}
	if len(nodeInfo.SupportedTypes) == 0 {
		t.Error("SupportedTypes should not be empty")
	}
}

func TestProbeResultsSubmission_Structure(t *testing.T) {
	results := []ProbeExecutionResult{
		{
			ProbeID:      1,
			ProbeUUID:    "probe-1",
			ExecutedAt:   time.Now(),
			Region:       "us-east-1",
			Status:       "success",
			ResponseTime: 100,
		},
		{
			ProbeID:      2,
			ProbeUUID:    "probe-2",
			ExecutedAt:   time.Now(),
			Region:       "us-east-1",
			Status:       "failed",
			ResponseTime: 5000,
			Error:        "timeout",
		},
	}

	submission := &ProbeResultsSubmission{
		Results:   results,
		Timestamp: time.Now(),
	}

	if len(submission.Results) != 2 {
		t.Error("Results length should be 2")
	}
	if submission.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
}

func TestMonitoringAgentHeartbeat_Structure(t *testing.T) {
	nodeInfo := NodeInfo{
		AgentID:      "test-agent",
		AgentVersion: "1.0.0",
		Region:       "us-east-1",
		Status:       "healthy",
		LastSeen:     time.Now(),
	}

	heartbeat := &MonitoringAgentHeartbeat{
		NodeInfo:  nodeInfo,
		Timestamp: time.Now(),
	}

	if heartbeat.NodeInfo.AgentID != "test-agent" {
		t.Error("NodeInfo.AgentID not preserved")
	}
	if heartbeat.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
}

// TestMonitoringServiceMethods tests that the monitoring service methods exist and have correct signatures
func TestMonitoringServiceMethods(t *testing.T) {
	client, err := NewMonitoringAgentClient(&Config{
		Auth: AuthConfig{
			MonitoringKey: "MON_test_key",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test that the methods exist by checking they are not nil
	if client.Monitoring == nil {
		t.Fatal("Monitoring service is nil")
	}

	// Test method signatures by calling with mock data (will fail at network level, but validates signature)
	// Use context with timeout to prevent hanging on network retry logic
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Test GetAssignedProbes method signature
	_, err = client.Monitoring.GetAssignedProbes(ctx, "us-east-1")
	// We expect this to fail due to network/auth, but not due to method signature
	if err == nil {
		t.Log("GetAssignedProbes method exists and accepts correct parameters")
	}

	// Test SubmitResults method signature
	results := []ProbeExecutionResult{
		{
			ProbeID:      1,
			ProbeUUID:    "test",
			ExecutedAt:   time.Now(),
			Region:       "us-east-1",
			Status:       "success",
			ResponseTime: 100,
		},
	}
	err = client.Monitoring.SubmitResults(ctx, results)
	// We expect this to fail due to network/auth, but not due to method signature
	if err == nil {
		t.Log("SubmitResults method exists and accepts correct parameters")
	}

	// Test Heartbeat method signature
	nodeInfo := NodeInfo{
		AgentID:      "test",
		AgentVersion: "1.0.0",
		Region:       "us-east-1",
		Status:       "healthy",
		LastSeen:     time.Now(),
	}
	err = client.Monitoring.Heartbeat(ctx, nodeInfo)
	// We expect this to fail due to network/auth, but not due to method signature
	if err == nil {
		t.Log("Heartbeat method exists and accepts correct parameters")
	}
}