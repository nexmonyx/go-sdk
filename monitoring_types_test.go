package nexmonyx

import (
	"encoding/json"
	"testing"
	"time"
)

func TestProbeRequestMarshaling(t *testing.T) {
	probe := &ProbeRequest{
		Name:        "test-probe",
		Description: "Test probe description",
		Type:        "http",
		Scope:       "public",
		Target:      "https://example.com",
		Interval:    60,
		Timeout:     30,
		Enabled:     true,
		Config: &ProbeConfig{
			Method:             stringPtr("GET"),
			ExpectedStatusCode: intPtr(200),
			FollowRedirects:    boolPtr(true),
		},
		Regions:        []string{"us-east-1", "eu-west-1"},
		AlertThreshold: 3,
		AlertEnabled:   true,
	}

	// Test marshaling
	data, err := json.Marshal(probe)
	if err != nil {
		t.Fatalf("Failed to marshal ProbeRequest: %v", err)
	}

	// Test unmarshaling
	var unmarshaled ProbeRequest
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal ProbeRequest: %v", err)
	}

	// Verify critical fields
	if unmarshaled.Name != probe.Name {
		t.Errorf("Expected name %s, got %s", probe.Name, unmarshaled.Name)
	}
	if unmarshaled.Type != probe.Type {
		t.Errorf("Expected type %s, got %s", probe.Type, unmarshaled.Type)
	}
	if unmarshaled.Target != probe.Target {
		t.Errorf("Expected target %s, got %s", probe.Target, unmarshaled.Target)
	}
	if len(unmarshaled.Regions) != len(probe.Regions) {
		t.Errorf("Expected %d regions, got %d", len(probe.Regions), len(unmarshaled.Regions))
	}
}

func TestProbeResultMarshaling(t *testing.T) {
	now := time.Now()
	result := &ProbeResult{
		ProbeID:      1,
		ProbeUUID:    "test-uuid-123",
		Region:       "us-east-1",
		Status:       "up",
		ResponseTime: 150,
		ExecutedAt:   &CustomTime{Time: now},
		Details: &ProbeResultDetails{
			StatusCode:   intPtr(200),
			ResponseSize: intPtr(1024),
			ContentMatch: boolPtr(true),
		},
	}

	// Test marshaling
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal ProbeResult: %v", err)
	}

	// Test unmarshaling
	var unmarshaled ProbeResult
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal ProbeResult: %v", err)
	}

	// Verify critical fields
	if unmarshaled.ProbeUUID != result.ProbeUUID {
		t.Errorf("Expected probe UUID %s, got %s", result.ProbeUUID, unmarshaled.ProbeUUID)
	}
	if unmarshaled.Status != result.Status {
		t.Errorf("Expected status %s, got %s", result.Status, unmarshaled.Status)
	}
	if unmarshaled.ResponseTime != result.ResponseTime {
		t.Errorf("Expected response time %d, got %d", result.ResponseTime, unmarshaled.ResponseTime)
	}
}

func TestMonitoringAgentMarshaling(t *testing.T) {
	agent := &MonitoringAgent{
		NodeUUID:       "node-uuid-456",
		Name:           "test-agent",
		Description:    "Test monitoring agent",
		OrganizationID: uintPtr(1),
		Type:           "private",
		Status:         "active",
		Region:         "us-west-2",
		Location:       "California",
		IPAddress:      "192.168.1.100",
		Enabled:        true,
		MaxProbes:      100,
		CurrentProbes:  25,
	}

	// Test marshaling
	data, err := json.Marshal(agent)
	if err != nil {
		t.Fatalf("Failed to marshal MonitoringAgent: %v", err)
	}

	// Test unmarshaling
	var unmarshaled MonitoringAgent
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal MonitoringAgent: %v", err)
	}

	// Verify critical fields
	if unmarshaled.NodeUUID != agent.NodeUUID {
		t.Errorf("Expected node UUID %s, got %s", agent.NodeUUID, unmarshaled.NodeUUID)
	}
	if unmarshaled.Name != agent.Name {
		t.Errorf("Expected name %s, got %s", agent.Name, unmarshaled.Name)
	}
	if unmarshaled.Status != agent.Status {
		t.Errorf("Expected status %s, got %s", agent.Status, unmarshaled.Status)
	}
}

func TestMonitoringDeploymentMarshaling(t *testing.T) {
	deployment := &MonitoringDeployment{
		ID:             1, // Add ID field from the actual struct
		OrganizationID: 1,
		// OrganizationUUID: "org-uuid-789", // Remove field not in actual struct
		Region:         "us-east-1",
		NamespaceName:  "nexmonyx-dev", // Use NamespaceName from actual struct
		DeploymentName: "monitoring-agent",
		// ClusterName:      "dev-cluster", // Remove field not in actual struct
		Status: "active",
		// HealthCheckCount: 10, // Remove field not in actual struct
		ErrorCount: 0,
		// Add other fields from the actual struct with zero values or appropriate test values if needed
		CurrentVersion: "v1.0.0",
		TargetVersion:  "v1.0.0",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Test marshaling
	data, err := json.Marshal(deployment)
	if err != nil {
		t.Fatalf("Failed to marshal MonitoringDeployment: %v", err)
	}

	// Test unmarshaling
	var unmarshaled MonitoringDeployment
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal MonitoringDeployment: %v", err)
	}

	// Verify critical fields
	// Remove assertions for fields not in the actual struct
	// if unmarshaled.OrganizationUUID != deployment.OrganizationUUID {
	// 	t.Errorf("Expected org UUID %s, got %s", deployment.OrganizationUUID, unmarshaled.OrganizationUUID)
	// }
	// if unmarshaled.Environment != deployment.Environment {
	// 	t.Errorf("Expected environment %s, got %s", deployment.Environment, unmarshaled.Environment)
	// }
	if unmarshaled.Status != deployment.Status {
		t.Errorf("Expected status %s, got %s", deployment.Status, unmarshaled.Status)
	}
	if unmarshaled.NamespaceName != deployment.NamespaceName {
		t.Errorf("Expected namespace %s, got %s", deployment.NamespaceName, unmarshaled.NamespaceName)
	}
}

func TestProbeAlertChannelMarshaling(t *testing.T) {
	channel := &ProbeAlertChannel{
		ProbeID: 1,
		Type:    "email",
		Name:    "Email Alert",
		Enabled: true,
		Config: &AlertConfig{
			Recipients: []string{"admin@example.com", "ops@example.com"},
		},
	}

	// Test marshaling
	data, err := json.Marshal(channel)
	if err != nil {
		t.Fatalf("Failed to marshal ProbeAlertChannel: %v", err)
	}

	// Test unmarshaling
	var unmarshaled ProbeAlertChannel
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal ProbeAlertChannel: %v", err)
	}

	// Verify critical fields
	if unmarshaled.Type != channel.Type {
		t.Errorf("Expected type %s, got %s", channel.Type, unmarshaled.Type)
	}
	if unmarshaled.Name != channel.Name {
		t.Errorf("Expected name %s, got %s", channel.Name, unmarshaled.Name)
	}
	if len(unmarshaled.Config.Recipients) != len(channel.Config.Recipients) {
		t.Errorf("Expected %d recipients, got %d", len(channel.Config.Recipients), len(unmarshaled.Config.Recipients))
	}
}

// Helper functions for test pointers
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

func uintPtr(u uint) *uint {
	return &u
}
