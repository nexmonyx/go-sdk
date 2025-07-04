package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestIPMIService_Submit(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/v2/ipmi/data" {
			t.Errorf("Expected path /v2/ipmi/data, got %s", r.URL.Path)
		}

		// Verify headers
		if r.Header.Get("X-Server-UUID") == "" {
			t.Error("Missing X-Server-UUID header")
		}
		if r.Header.Get("X-Server-Secret") == "" {
			t.Error("Missing X-Server-Secret header")
		}

		// Parse request body
		var req IPMISubmitRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		// Verify required fields
		if req.IPMI.CollectionMethod == "" {
			t.Error("Missing collection method")
		}

		// Send response
		response := map[string]interface{}{
			"success": true,
			"data": IPMISubmitResponse{
				ServerUUID:       req.ServerUUID,
				Timestamp:        time.Now(),
				CollectionMethod: req.IPMI.CollectionMethod,
				IPMIVersion:      req.IPMI.IPMIVersion,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			ServerUUID:   "test-server-uuid",
			ServerSecret: "test-server-secret",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Create test IPMI data
	ipmiInfo := IPMIInfo{
		CollectionMethod: "ipmitool",
		IPMIVersion:      "2.0",
		BMC: &BMCInfo{
			DeviceID:         "0x20",
			FirmwareRevision: "2.53",
			ManufacturerName: "Dell Inc.",
			ProductName:      "iDRAC9",
		},
		Sensors: []IPMISensorInfo{
			{
				SensorID:   "temp1",
				SensorName: "CPU1 Temp",
				SensorType: "temperature",
				Value:      45.0,
				Unit:       "degrees C",
				Status:     "ok",
			},
		},
		PowerInfo: &IPMIPowerInfo{
			PowerConsumption: 250.5,
			PowerCapacity:    750.0,
			PowerState:       "on",
		},
		Fans: []IPMIFanInfo{
			{
				FanID:        "fan1",
				FanName:      "System Fan 1",
				Speed:        3000,
				SpeedPercent: 50.0,
				Status:       "ok",
			},
		},
		Temperatures: []IPMITemperatureInfo{
			{
				SensorID:     "cpu1_temp",
				SensorName:   "CPU1 Temperature",
				Temperature:  45.0,
				Status:       "ok",
				UpperWarning: func() *float64 { v := 75.0; return &v }(),
			},
		},
		SystemHealth: &IPMISystemHealth{
			OverallStatus:  "ok",
			PowerStatus:    "ok",
			ThermalStatus:  "ok",
			FanStatus:      "ok",
			VoltageStatus:  "ok",
			CriticalEvents: 0,
			WarningEvents:  2,
			HealthScore:    95,
		},
	}

	// Submit IPMI data
	req := &IPMISubmitRequest{
		ServerUUID:  "test-server-uuid",
		CollectedAt: time.Now(),
		IPMI:        ipmiInfo,
	}

	resp, err := client.IPMI.Submit(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to submit IPMI data: %v", err)
	}

	// Verify response
	if resp.ServerUUID != req.ServerUUID {
		t.Errorf("Expected server UUID %s, got %s", req.ServerUUID, resp.ServerUUID)
	}
	if resp.CollectionMethod != ipmiInfo.CollectionMethod {
		t.Errorf("Expected collection method %s, got %s", ipmiInfo.CollectionMethod, resp.CollectionMethod)
	}
}

func TestIPMIService_GetIPMI(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		expectedPath := "/v2/ipmi/test-server-uuid"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Send response
		response := map[string]interface{}{
			"success": true,
			"data": IPMIInfo{
				CollectionMethod: "ipmitool",
				IPMIVersion:      "2.0",
				BMC: &BMCInfo{
					DeviceID:         "0x20",
					FirmwareRevision: "2.53",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Get IPMI data
	ipmi, err := client.IPMI.GetIPMI(context.Background(), "test-server-uuid", nil)
	if err != nil {
		t.Fatalf("Failed to get IPMI data: %v", err)
	}

	// Verify response
	if ipmi.CollectionMethod != "ipmitool" {
		t.Errorf("Expected collection method ipmitool, got %s", ipmi.CollectionMethod)
	}
	if ipmi.IPMIVersion != "2.0" {
		t.Errorf("Expected IPMI version 2.0, got %s", ipmi.IPMIVersion)
	}
}

func TestIPMIService_GetLatestIPMI(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		expectedPath := "/v2/ipmi/test-server-uuid/latest"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Send response
		response := map[string]interface{}{
			"success": true,
			"data": IPMIInfo{
				CollectionMethod: "ipmitool",
				SystemHealth: &IPMISystemHealth{
					OverallStatus: "ok",
					HealthScore:   100,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Get latest IPMI data
	ipmi, err := client.IPMI.GetLatestIPMI(context.Background(), "test-server-uuid")
	if err != nil {
		t.Fatalf("Failed to get latest IPMI data: %v", err)
	}

	// Verify response
	if ipmi.SystemHealth == nil {
		t.Fatal("Expected system health data")
	}
	if ipmi.SystemHealth.OverallStatus != "ok" {
		t.Errorf("Expected overall status ok, got %s", ipmi.SystemHealth.OverallStatus)
	}
	if ipmi.SystemHealth.HealthScore != 100 {
		t.Errorf("Expected health score 100, got %d", ipmi.SystemHealth.HealthScore)
	}
}

func TestIPMIService_ListIPMIHistory(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		expectedPath := "/v2/ipmi/test-server-uuid/history"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Send response
		response := map[string]interface{}{
			"success": true,
			"data": []IPMIInfo{
				{
					CollectionMethod: "ipmitool",
					IPMIVersion:      "2.0",
				},
				{
					CollectionMethod: "ipmi-sensors",
					IPMIVersion:      "2.0",
				},
			},
			"meta": &PaginationMeta{
				Page:       1,
				Limit:      20,
				TotalItems: 2,
				TotalPages: 1,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// List IPMI history
	history, meta, err := client.IPMI.ListIPMIHistory(context.Background(), "test-server-uuid", nil)
	if err != nil {
		t.Fatalf("Failed to list IPMI history: %v", err)
	}

	// Verify response
	if len(history) != 2 {
		t.Errorf("Expected 2 history items, got %d", len(history))
	}
	if meta.TotalItems != 2 {
		t.Errorf("Expected total 2, got %d", meta.TotalItems)
	}
}
