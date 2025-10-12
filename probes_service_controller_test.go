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

// TestProbesService_GetActiveProbes tests the GetActiveProbes method
func TestProbesService_GetActiveProbes(t *testing.T) {
	tests := []struct {
		name           string
		orgID          uint
		allProbes      []*MonitoringProbe
		expectedActive int
		expectError    bool
	}{
		{
			name:  "filters enabled probes only",
			orgID: 1,
			allProbes: []*MonitoringProbe{
				{
					GormModel: GormModel{ID: 1},
					Name:      "active-probe-1",
					Type:      "http",
					Target:    "https://example.com",
					Enabled:   true,
					Interval:  60,
				},
				{
					GormModel: GormModel{ID: 2},
					Name:      "disabled-probe",
					Type:      "icmp",
					Target:    "8.8.8.8",
					Enabled:   false, // Disabled
					Interval:  30,
				},
				{
					GormModel: GormModel{ID: 3},
					Name:      "active-probe-2",
					Type:      "tcp",
					Target:    "localhost:443",
					Enabled:   true,
					Interval:  120,
				},
			},
			expectedActive: 2, // Only 2 enabled probes
			expectError:    false,
		},
		{
			name:  "all probes disabled",
			orgID: 2,
			allProbes: []*MonitoringProbe{
				{
					GormModel: GormModel{ID: 4},
					Name:      "disabled-probe-1",
					Type:      "http",
					Target:    "https://test.com",
					Enabled:   false,
					Interval:  60,
				},
				{
					GormModel: GormModel{ID: 5},
					Name:      "disabled-probe-2",
					Type:      "icmp",
					Target:    "1.1.1.1",
					Enabled:   false,
					Interval:  30,
				},
			},
			expectedActive: 0, // No enabled probes
			expectError:    false,
		},
		{
			name:  "all probes enabled",
			orgID: 3,
			allProbes: []*MonitoringProbe{
				{
					GormModel: GormModel{ID: 6},
					Name:      "active-probe-3",
					Type:      "http",
					Target:    "https://google.com",
					Enabled:   true,
					Interval:  60,
				},
				{
					GormModel: GormModel{ID: 7},
					Name:      "active-probe-4",
					Type:      "https",
					Target:    "https://cloudflare.com",
					Enabled:   true,
					Interval:  120,
				},
			},
			expectedActive: 2, // All enabled
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

				response := struct {
					Status string             `json:"status"`
					Data   []*MonitoringProbe `json:"data"`
				}{
					Status: "success",
					Data:   tt.allProbes,
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

			activeProbes, err := client.Probes.GetActiveProbes(context.Background(), tt.orgID)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("GetActiveProbes failed: %v", err)
			}

			if len(activeProbes) != tt.expectedActive {
				t.Errorf("Expected %d active probes, got %d", tt.expectedActive, len(activeProbes))
			}

			// Verify all returned probes are enabled
			for _, probe := range activeProbes {
				if !probe.Enabled {
					t.Errorf("GetActiveProbes returned disabled probe: %s", probe.Name)
				}
			}
		})
	}
}

// TestProbesService_SubmitResult tests the SubmitResult method
func TestProbesService_SubmitResult(t *testing.T) {
	tests := []struct {
		name        string
		result      *ProbeExecutionResult
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful result submission",
			result: &ProbeExecutionResult{
				ProbeID:      1,
				ProbeUUID:    "test-probe-uuid",
				ExecutedAt:   time.Now(),
				Region:       "us-east-1",
				Status:       "success",
				ResponseTime: 150,
				StatusCode:   200,
			},
			expectError: false,
		},
		{
			name: "result with error message",
			result: &ProbeExecutionResult{
				ProbeID:      2,
				ProbeUUID:    "test-probe-uuid-2",
				ExecutedAt:   time.Now(),
				Region:       "eu-west-1",
				Status:       "failed",
				ResponseTime: 0,
				StatusCode:   0,
				Error:        "Connection timeout",
			},
			expectError: false,
		},
		{
			name:        "nil result",
			result:      nil,
			expectError: true,
			errorMsg:    "probe result cannot be nil",
		},
		{
			name: "result without probe UUID",
			result: &ProbeExecutionResult{
				ProbeID:      3,
				ProbeUUID:    "", // Missing UUID
				ExecutedAt:   time.Now(),
				Region:       "ap-south-1",
				Status:       "success",
				ResponseTime: 200,
				StatusCode:   200,
			},
			expectError: true,
			errorMsg:    "probe UUID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					t.Errorf("Expected POST method, got %s", r.Method)
				}
				if !strings.Contains(r.URL.Path, "/v1/monitoring/results") {
					t.Errorf("Expected path /v1/monitoring/results, got %s", r.URL.Path)
				}

				// Verify request body structure (wrapped in ProbeResultsSubmission)
				var body struct {
					Results []ProbeExecutionResult `json:"results"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("Failed to decode request body: %v", err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				if len(body.Results) != 1 {
					t.Errorf("Expected 1 result in array, got %d", len(body.Results))
				}

				response := struct {
					Status  string `json:"status"`
					Message string `json:"message"`
				}{
					Status:  "success",
					Message: "Results submitted successfully",
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

			err = client.Probes.SubmitResult(context.Background(), tt.result)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.errorMsg, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("SubmitResult failed: %v", err)
			}
		})
	}
}
