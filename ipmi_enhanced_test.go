package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Additional tests to improve ipmi.go coverage from 42.39% to 70%+

func TestIPMIService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/ipmi/test-server-uuid", r.URL.Path)

		response := map[string]interface{}{
			"success": true,
			"data": IPMIData{
				BMCInfo: &BMCInfo{
					Version:          "2.0",
					Manufacturer:     "Dell Inc.",
					Firmware:         "3.45",
					IPAddress:        "192.168.1.100",
					MACAddress:       "00:11:22:33:44:55",
					DeviceID:         "0x20",
					FirmwareRevision: "2.53",
					ManufacturerName: "Dell Inc.",
					ProductName:      "iDRAC9",
				},
				ChassisStatus: &ChassisStatus{
					PowerState:        "on",
					ChassisIntrusion:  false,
					FrontPanelLockout: false,
					DriveFault:        false,
					CoolingFault:      false,
				},
				PowerStatus: &PowerStatus{
					PowerOn:          true,
					PowerConsumption: 250.5,
					PowerCapacity:    750.0,
				},
				SystemHealth: "ok",
				FanStatus: []FanStatus{
					{
						Name:    "Fan1",
						RPM:     3000,
						Status:  "ok",
						Percent: 50.0,
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	require.NoError(t, err)

	ipmiData, err := client.IPMI.Get(context.Background(), "test-server-uuid")
	require.NoError(t, err)
	assert.NotNil(t, ipmiData)
	assert.NotNil(t, ipmiData.BMCInfo)
	assert.Equal(t, "Dell Inc.", ipmiData.BMCInfo.Manufacturer)
	assert.Equal(t, "ok", ipmiData.SystemHealth)
	assert.Len(t, ipmiData.FanStatus, 1)
}

func TestIPMIService_GetSensorData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/ipmi/test-server-uuid/sensors", r.URL.Path)

		sensors := []*IPMISensor{
			{
				ID:          "cpu1_temp",
				Name:        "CPU1 Temperature",
				Type:        "temperature",
				Reading:     45.0,
				Unit:        "degrees C",
				Status:      "ok",
				LowerBound:  0.0,
				UpperBound:  90.0,
				Description: "CPU core temperature",
			},
			{
				ID:         "fan1_speed",
				Name:       "System Fan 1",
				Type:       "fan",
				Reading:    3000.0,
				Unit:       "RPM",
				Status:     "ok",
				LowerBound: 1000.0,
				UpperBound: 5000.0,
			},
			{
				ID:         "voltage_12v",
				Name:       "+12V Rail",
				Type:       "voltage",
				Reading:    12.1,
				Unit:       "volts",
				Status:     "ok",
				LowerBound: 11.5,
				UpperBound: 12.5,
			},
		}

		response := map[string]interface{}{
			"success": true,
			"data":    sensors,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	require.NoError(t, err)

	sensors, err := client.IPMI.GetSensorData(context.Background(), "test-server-uuid")
	require.NoError(t, err)
	assert.Len(t, sensors, 3)
	assert.Equal(t, "cpu1_temp", sensors[0].ID)
	assert.Equal(t, "temperature", sensors[0].Type)
	assert.Equal(t, 45.0, sensors[0].Reading)
}

func TestIPMIService_ExecuteCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/ipmi/test-server-uuid/execute", r.URL.Path)

		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "chassis power status", body["command"])

		result := &IPMICommandResult{
			Command:    "chassis power status",
			Output:     "Chassis Power is on",
			ExitCode:   0,
			ExecutedAt: time.Now(),
		}

		response := map[string]interface{}{
			"success": true,
			"data":    result,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	require.NoError(t, err)

	result, err := client.IPMI.ExecuteCommand(context.Background(), "test-server-uuid", "chassis power status", []string{})
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "chassis power status", result.Command)
	assert.Equal(t, "Chassis Power is on", result.Output)
	assert.Equal(t, 0, result.ExitCode)
}

func TestIPMIService_ExecuteCommand_WithArgs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)

		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "sensor reading", body["command"])

		args, ok := body["args"].([]interface{})
		require.True(t, ok)
		assert.Len(t, args, 2)

		result := &IPMICommandResult{
			Command:    "sensor reading",
			Output:     "CPU1 Temp: 45.0 degrees C",
			ExitCode:   0,
			ExecutedAt: time.Now(),
		}

		response := map[string]interface{}{
			"success": true,
			"data":    result,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	require.NoError(t, err)

	result, err := client.IPMI.ExecuteCommand(context.Background(), "test-server-uuid", "sensor reading", []string{"cpu1", "temp"})
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result.Output, "45.0 degrees C")
}

func TestIPMIService_ExecuteCommand_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)

		result := &IPMICommandResult{
			Command:    "invalid command",
			Output:     "",
			Error:      "Command not found",
			ExitCode:   1,
			ExecutedAt: time.Now(),
		}

		response := map[string]interface{}{
			"success": true,
			"data":    result,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	require.NoError(t, err)

	result, err := client.IPMI.ExecuteCommand(context.Background(), "test-server-uuid", "invalid command", nil)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.ExitCode)
	assert.Equal(t, "Command not found", result.Error)
}
