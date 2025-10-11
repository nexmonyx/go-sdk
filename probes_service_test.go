package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestProbesService_ListByOrganization tests the ListByOrganization method
func TestProbesService_ListByOrganization(t *testing.T) {
	tests := []struct {
		name           string
		orgID          uint
		responseProbes []*MonitoringProbe
		expectError    bool
	}{
		{
			name:  "successful list with multiple probes",
			orgID: 1,
			responseProbes: []*MonitoringProbe{
				{
					GormModel: GormModel{ID: 1},
					Name:      "test-probe-1",
					Type:      "http",
					Target:    "https://example.com",
					Enabled:   true,
					Interval:  60,
				},
				{
					GormModel: GormModel{ID: 2},
					Name:      "test-probe-2",
					Type:      "icmp",
					Target:    "8.8.8.8",
					Enabled:   true,
					Interval:  30,
				},
			},
			expectError: false,
		},
		{
			name:           "empty result for organization",
			orgID:          999,
			responseProbes: []*MonitoringProbe{},
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("Expected GET method, got %s", r.Method)
				}
				if !strings.Contains(r.URL.Path, "/v1/controllers/probes/list") {
					t.Errorf("Expected /v1/controllers/probes/list, got %s", r.URL.Path)
				}

				// Verify org_id query parameter
				orgID := r.URL.Query().Get("org_id")
				if orgID == "" {
					t.Error("Expected org_id query parameter")
				}

				response := struct {
					Status string             `json:"status"`
					Data   []*MonitoringProbe `json:"data"`
				}{
					Status: "success",
					Data:   tt.responseProbes,
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth: AuthConfig{
					MonitoringKey: "test-monitoring-key",
				},
			})
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			probes, err := client.Probes.ListByOrganization(context.Background(), tt.orgID)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("ListByOrganization failed: %v", err)
			}

			if len(probes) != len(tt.responseProbes) {
				t.Errorf("Expected %d probes, got %d", len(tt.responseProbes), len(probes))
			}

			for i, probe := range probes {
				if probe.Name != tt.responseProbes[i].Name {
					t.Errorf("Expected probe name '%s', got '%s'", tt.responseProbes[i].Name, probe.Name)
				}
			}
		})
	}
}

// TestProbesService_GetByUUID tests the GetByUUID method
func TestProbesService_GetByUUID(t *testing.T) {
	tests := []struct {
		name        string
		probeUUID   string
		probe       *MonitoringProbe
		expectError bool
	}{
		{
			name:      "successful get probe",
			probeUUID: "test-probe-uuid",
			probe: &MonitoringProbe{
				GormModel: GormModel{ID: 1},
				Name:      "test-probe",
				Type:      "http",
				Target:    "https://example.com",
				Enabled:   true,
				Interval:  60,
				Regions:   []string{"us-east-1", "eu-west-1"},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("Expected GET method, got %s", r.Method)
				}

				response := struct {
					Status string           `json:"status"`
					Data   *MonitoringProbe `json:"data"`
				}{
					Status: "success",
					Data:   tt.probe,
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth: AuthConfig{
					MonitoringKey: "test-monitoring-key",
				},
			})
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			probe, err := client.Probes.GetByUUID(context.Background(), tt.probeUUID)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("GetByUUID failed: %v", err)
			}

			if probe.Name != tt.probe.Name {
				t.Errorf("Expected probe name '%s', got '%s'", tt.probe.Name, probe.Name)
			}
			if probe.Type != tt.probe.Type {
				t.Errorf("Expected probe type '%s', got '%s'", tt.probe.Type, probe.Type)
			}
		})
	}
}

// TestProbesService_GetRegionalResults tests the GetRegionalResults method
func TestProbesService_GetRegionalResults(t *testing.T) {
	tests := []struct {
		name           string
		probeUUID      string
		regionalResult []RegionalResult
		expectError    bool
	}{
		{
			name:      "successful get regional results with multiple regions",
			probeUUID: "test-probe-uuid",
			regionalResult: []RegionalResult{
				{
					Region:       "us-east-1",
					Status:       "up",
					ResponseTime: 150,
					CheckedAt:    time.Now().Format(time.RFC3339),
				},
				{
					Region:       "eu-west-1",
					Status:       "up",
					ResponseTime: 200,
					CheckedAt:    time.Now().Format(time.RFC3339),
				},
				{
					Region:       "ap-south-1",
					Status:       "down",
					ResponseTime: 0,
					ErrorMessage: stringPtr("Connection timeout"),
					CheckedAt:    time.Now().Format(time.RFC3339),
				},
			},
			expectError: false,
		},
		{
			name:           "empty results",
			probeUUID:      "probe-with-no-results",
			regionalResult: []RegionalResult{},
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("Expected GET method, got %s", r.Method)
				}
				if !strings.Contains(r.URL.Path, "/v1/controllers/probes/"+tt.probeUUID+"/regional-results") {
					t.Errorf("Expected path to contain probe UUID, got %s", r.URL.Path)
				}

				response := struct {
					Status string           `json:"status"`
					Data   []RegionalResult `json:"data"`
				}{
					Status: "success",
					Data:   tt.regionalResult,
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth: AuthConfig{
					MonitoringKey: "test-monitoring-key",
				},
			})
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			results, err := client.Probes.GetRegionalResults(context.Background(), tt.probeUUID)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("GetRegionalResults failed: %v", err)
			}

			if len(results) != len(tt.regionalResult) {
				t.Errorf("Expected %d regional results, got %d", len(tt.regionalResult), len(results))
			}

			for i, result := range results {
				if result.Region != tt.regionalResult[i].Region {
					t.Errorf("Expected region '%s', got '%s'", tt.regionalResult[i].Region, result.Region)
				}
				if result.Status != tt.regionalResult[i].Status {
					t.Errorf("Expected status '%s', got '%s'", tt.regionalResult[i].Status, result.Status)
				}
			}
		})
	}
}

// TestProbesService_UpdateControllerStatus tests the UpdateControllerStatus method
func TestProbesService_UpdateControllerStatus(t *testing.T) {
	tests := []struct {
		name        string
		probeUUID   string
		status      string
		expectError bool
	}{
		{
			name:        "update status to up",
			probeUUID:   "test-probe-uuid",
			status:      "up",
			expectError: false,
		},
		{
			name:        "update status to down",
			probeUUID:   "test-probe-uuid",
			status:      "down",
			expectError: false,
		},
		{
			name:        "update status to degraded",
			probeUUID:   "test-probe-uuid",
			status:      "degraded",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "PUT" {
					t.Errorf("Expected PUT method, got %s", r.Method)
				}
				if !strings.Contains(r.URL.Path, "/v1/controllers/probes/"+tt.probeUUID+"/status") {
					t.Errorf("Expected path to contain probe UUID, got %s", r.URL.Path)
				}

				// Verify request body
				var body map[string]interface{}
				json.NewDecoder(r.Body).Decode(&body)
				if body["status"] != tt.status {
					t.Errorf("Expected status '%s', got '%s'", tt.status, body["status"])
				}

				response := struct {
					Status  string `json:"status"`
					Message string `json:"message"`
				}{
					Status:  "success",
					Message: "Status updated successfully",
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth: AuthConfig{
					MonitoringKey: "test-monitoring-key",
				},
			})
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			err = client.Probes.UpdateControllerStatus(context.Background(), tt.probeUUID, tt.status)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("UpdateControllerStatus failed: %v", err)
			}
		})
	}
}

// TestProbesService_GetProbeConfig tests the GetProbeConfig method
func TestProbesService_GetProbeConfig(t *testing.T) {
	tests := []struct {
		name              string
		probeUUID         string
		probe             *MonitoringProbe
		expectedConsensus string
		expectError       bool
	}{
		{
			name:      "probe with explicit consensus type",
			probeUUID: "test-probe-uuid-1",
			probe: &MonitoringProbe{
				GormModel: GormModel{ID: 1},
				Name:      "test-probe-1",
				Type:      "http",
				Target:    "https://example.com",
				Enabled:   true,
				Interval:  60,
				Timeout:   10,
				Regions:   []string{"us-east-1", "eu-west-1", "ap-south-1"},
				Config: map[string]interface{}{
					"consensus_type": "all",
				},
			},
			expectedConsensus: "all",
			expectError:       false,
		},
		{
			name:      "probe without consensus type (defaults to majority)",
			probeUUID: "test-probe-uuid-2",
			probe: &MonitoringProbe{
				GormModel: GormModel{ID: 2},
				Name:      "test-probe-2",
				Type:      "icmp",
				Target:    "8.8.8.8",
				Enabled:   true,
				Interval:  30,
				Timeout:   5,
				Regions:   []string{"us-east-1", "eu-west-1"},
				Config:    map[string]interface{}{},
			},
			expectedConsensus: "majority",
			expectError:       false,
		},
		{
			name:      "probe with any consensus type",
			probeUUID: "test-probe-uuid-3",
			probe: &MonitoringProbe{
				GormModel: GormModel{ID: 3},
				Name:      "test-probe-3",
				Type:      "tcp",
				Target:    "database.example.com:5432",
				Enabled:   true,
				Interval:  120,
				Timeout:   15,
				Regions:   []string{"us-east-1"},
				Config: map[string]interface{}{
					"consensus_type": "any",
				},
			},
			expectedConsensus: "any",
			expectError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("Expected GET method, got %s", r.Method)
				}

				response := struct {
					Status string           `json:"status"`
					Data   *MonitoringProbe `json:"data"`
				}{
					Status: "success",
					Data:   tt.probe,
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth: AuthConfig{
					MonitoringKey: "test-monitoring-key",
				},
			})
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			config, err := client.Probes.GetProbeConfig(context.Background(), tt.probeUUID)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("GetProbeConfig failed: %v", err)
			}

			if config.ProbeUUID != tt.probeUUID {
				t.Errorf("Expected probe UUID '%s', got '%s'", tt.probeUUID, config.ProbeUUID)
			}
			if config.ConsensusType != tt.expectedConsensus {
				t.Errorf("Expected consensus type '%s', got '%s'", tt.expectedConsensus, config.ConsensusType)
			}
			if config.Name != tt.probe.Name {
				t.Errorf("Expected name '%s', got '%s'", tt.probe.Name, config.Name)
			}
			if config.Type != tt.probe.Type {
				t.Errorf("Expected type '%s', got '%s'", tt.probe.Type, config.Type)
			}
			if len(config.Regions) != len(tt.probe.Regions) {
				t.Errorf("Expected %d regions, got %d", len(tt.probe.Regions), len(config.Regions))
			}
		})
	}
}

// TestProbesService_RecordConsensusResult tests the RecordConsensusResult method
func TestProbesService_RecordConsensusResult(t *testing.T) {
	tests := []struct {
		name        string
		result      *ConsensusResultRequest
		expectError bool
	}{
		{
			name: "successful consensus result recording",
			result: &ConsensusResultRequest{
				ProbeUUID:      "test-probe-uuid",
				OrganizationID: 1,
				ConsensusType:  "majority",
				GlobalStatus:   "up",
				RegionResults: []RegionalResult{
					{
						Region:       "us-east-1",
						Status:       "up",
						ResponseTime: 150,
						CheckedAt:    time.Now().Format(time.RFC3339),
					},
					{
						Region:       "eu-west-1",
						Status:       "up",
						ResponseTime: 200,
						CheckedAt:    time.Now().Format(time.RFC3339),
					},
				},
				TotalRegions:        2,
				UpRegions:           2,
				DownRegions:         0,
				DegradedRegions:     0,
				UnknownRegions:      0,
				ShouldAlert:         false,
				AverageResponseTime: 175,
				MinResponseTime:     150,
				MaxResponseTime:     200,
				UptimePercentage:    100.0,
			},
			expectError: false,
		},
		{
			name: "consensus result with down status",
			result: &ConsensusResultRequest{
				ProbeUUID:      "test-probe-uuid-2",
				OrganizationID: 1,
				ConsensusType:  "all",
				GlobalStatus:   "down",
				RegionResults: []RegionalResult{
					{
						Region:       "us-east-1",
						Status:       "up",
						ResponseTime: 150,
						CheckedAt:    time.Now().Format(time.RFC3339),
					},
					{
						Region:       "eu-west-1",
						Status:       "down",
						ResponseTime: 0,
						ErrorMessage: stringPtr("Connection timeout"),
						CheckedAt:    time.Now().Format(time.RFC3339),
					},
				},
				TotalRegions:        2,
				UpRegions:           1,
				DownRegions:         1,
				DegradedRegions:     0,
				UnknownRegions:      0,
				ShouldAlert:         true,
				AverageResponseTime: 75,
				MinResponseTime:     0,
				MaxResponseTime:     150,
				UptimePercentage:    50.0,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					t.Errorf("Expected POST method, got %s", r.Method)
				}
				if !strings.Contains(r.URL.Path, "/v1/controllers/probes/consensus-results") {
					t.Errorf("Expected path to contain consensus-results, got %s", r.URL.Path)
				}

				// Verify request body
				var body ConsensusResultRequest
				json.NewDecoder(r.Body).Decode(&body)
				if body.ProbeUUID != tt.result.ProbeUUID {
					t.Errorf("Expected probe UUID '%s', got '%s'", tt.result.ProbeUUID, body.ProbeUUID)
				}
				if body.GlobalStatus != tt.result.GlobalStatus {
					t.Errorf("Expected global status '%s', got '%s'", tt.result.GlobalStatus, body.GlobalStatus)
				}

				response := struct {
					Status  string `json:"status"`
					Message string `json:"message"`
				}{
					Status:  "success",
					Message: "Consensus result recorded successfully",
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth: AuthConfig{
					MonitoringKey: "test-monitoring-key",
				},
			})
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			err = client.Probes.RecordConsensusResult(context.Background(), tt.result)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("RecordConsensusResult failed: %v", err)
			}
		})
	}
}
