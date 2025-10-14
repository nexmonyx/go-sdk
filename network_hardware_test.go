package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNetworkHardwareService_Submit_WithDebug tests the Submit method with debug mode enabled
func TestNetworkHardwareService_Submit_WithDebug(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Send success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(StandardResponse{
			Status:  "success",
			Message: "Network hardware information submitted successfully",
		})
	}))
	defer server.Close()

	// Create client with debug enabled
	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
		Debug:   true,
	})
	require.NoError(t, err)

	// Submit with various interface types to cover debug logging
	interfaces := []NetworkHardwareInfo{
		{
			InterfaceName:    "eth0",
			InterfaceType:    "physical",
			MacAddress:       "00:11:22:33:44:55",
			SpeedMbps:        1000,
			OperationalState: "up",
			LinkDetected:     true,
			IPAddresses:      []string{"192.168.1.100"},
			RxBytes:          1024000,
			TxBytes:          512000,
		},
		{
			InterfaceName: "bond0",
			InterfaceType: "bond",
			BondMode:      "802.3ad",
			BondSlaves:    []string{"eth1", "eth2"},
		},
		{
			InterfaceName: "eth0.100",
			InterfaceType: "vlan",
			VlanID:        100,
			VlanParent:    "eth0",
		},
		{
			InterfaceName: "br0",
			InterfaceType: "bridge",
			BridgePorts:   []string{"eth3", "eth4"},
			BridgeSTP:     true,
		},
	}

	result, err := client.NetworkHardware.Submit(context.Background(), "server-uuid-123", interfaces)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "success", result.Status)
}

// TestNetworkHardwareService_Submit_WithDebug_Error tests error handling with debug mode
func TestNetworkHardwareService_Submit_WithDebug_Error(t *testing.T) {
	// Create mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(StandardResponse{
			Status:  "error",
			Message: "Server not found",
			Error:   "SERVER_NOT_FOUND",
		})
	}))
	defer server.Close()

	// Create client with debug enabled
	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
		Debug:   true,
	})
	require.NoError(t, err)

	interfaces := []NetworkHardwareInfo{
		{
			InterfaceName: "eth0",
			InterfaceType: "physical",
		},
	}

	result, err := client.NetworkHardware.Submit(context.Background(), "nonexistent-server", interfaces)
	assert.Error(t, err)
	assert.Nil(t, result)
}

// TestNetworkHardwareService_Submit_WithDebug_EmptyUUID tests empty UUID with debug mode
func TestNetworkHardwareService_Submit_WithDebug_EmptyUUID(t *testing.T) {
	client, err := NewClient(&Config{
		BaseURL: "http://localhost",
		Auth:    AuthConfig{Token: "test-token"},
		Debug:   true,
	})
	require.NoError(t, err)

	interfaces := []NetworkHardwareInfo{
		{
			InterfaceName: "eth0",
			InterfaceType: "physical",
		},
	}

	result, err := client.NetworkHardware.Submit(context.Background(), "", interfaces)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "server UUID is required")
}

// TestNetworkHardwareService_Submit tests the Submit method
func TestNetworkHardwareService_Submit(t *testing.T) {
	tests := []struct {
		name       string
		serverUUID string
		interfaces []NetworkHardwareInfo
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *StandardResponse)
	}{
		{
			name:       "successful submission with physical interface",
			serverUUID: "server-uuid-123",
			interfaces: []NetworkHardwareInfo{
				{
					InterfaceName:       "eth0",
					InterfaceType:       "physical",
					MacAddress:          "00:11:22:33:44:55",
					Manufacturer:        "Intel",
					DeviceName:          "Ethernet Controller",
					DriverName:          "e1000e",
					DriverVersion:       "3.2.6",
					SpeedMbps:           1000,
					Duplex:              "full",
					LinkDetected:        true,
					OperationalState:    "up",
					AdministrativeState: "up",
					MTU:                 1500,
					IPAddresses:         []string{"192.168.1.100"},
					RxBytes:             1024000,
					TxBytes:             512000,
					RxPackets:           1000,
					TxPackets:           500,
				},
			},
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Network hardware information submitted successfully",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, resp *StandardResponse) {
				assert.NotNil(t, resp)
				assert.Equal(t, "success", resp.Status)
			},
		},
		{
			name:       "successful submission with bond interface",
			serverUUID: "server-uuid-456",
			interfaces: []NetworkHardwareInfo{
				{
					InterfaceName:    "bond0",
					InterfaceType:    "bond",
					MacAddress:       "00:aa:bb:cc:dd:ee",
					BondMode:         "802.3ad",
					BondSlaves:       []string{"eth0", "eth1"},
					BondPrimary:      "eth0",
					BondActiveSlave:  "eth0",
					LACPRate:         "fast",
					XmitHashPolicy:   "layer3+4",
					OperationalState: "up",
					IPAddresses:      []string{"10.0.0.10"},
					SpeedMbps:        2000,
				},
			},
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Bond interface submitted successfully",
			},
			wantErr: false,
		},
		{
			name:       "successful submission with VLAN interface",
			serverUUID: "server-uuid-789",
			interfaces: []NetworkHardwareInfo{
				{
					InterfaceName:    "eth0.100",
					InterfaceType:    "vlan",
					MacAddress:       "00:ff:ee:dd:cc:bb",
					VlanID:           100,
					VlanParent:       "eth0",
					NativeVlan:       1,
					AllowedVlans:     []int{100, 200, 300},
					OperationalState: "up",
					IPAddresses:      []string{"172.16.100.10"},
					SpeedMbps:        1000,
				},
			},
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "VLAN interface submitted successfully",
			},
			wantErr: false,
		},
		{
			name:       "successful submission with bridge interface",
			serverUUID: "server-uuid-abc",
			interfaces: []NetworkHardwareInfo{
				{
					InterfaceName:      "br0",
					InterfaceType:      "bridge",
					MacAddress:         "00:12:34:56:78:9a",
					BridgePorts:        []string{"eth0", "eth1"},
					BridgeSTP:          true,
					BridgeForwardDelay: 15,
					BridgeHelloTime:    2,
					BridgeMaxAge:       20,
					BridgePriority:     32768,
					OperationalState:   "up",
					IPAddresses:        []string{"192.168.100.1"},
				},
			},
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Bridge interface submitted successfully",
			},
			wantErr: false,
		},
		{
			name:       "successful submission with wireless interface",
			serverUUID: "server-uuid-def",
			interfaces: []NetworkHardwareInfo{
				{
					InterfaceName:        "wlan0",
					InterfaceType:        "wireless",
					MacAddress:           "a1:b2:c3:d4:e5:f6",
					IsWireless:           true,
					WirelessMode:         "Managed",
					WirelessProtocol:     "802.11ac",
					WirelessFrequencyMHz: 5180.0,
					WirelessChannel:      36,
					WirelessSSID:         "MyNetwork",
					WirelessBSSID:        "00:11:22:33:44:55",
					WirelessEncryption:   "WPA2",
					SignalStrengthDBM:    -55.0,
					LinkQualityPercent:   85.0,
					NoiseLevelDBM:        -90.0,
					OperationalState:     "up",
					IPAddresses:          []string{"192.168.0.50"},
					SpeedMbps:            867,
				},
			},
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Wireless interface submitted successfully",
			},
			wantErr: false,
		},
		{
			name:       "successful submission with multiple interfaces",
			serverUUID: "server-uuid-multi",
			interfaces: []NetworkHardwareInfo{
				{
					InterfaceName:    "eth0",
					InterfaceType:    "physical",
					MacAddress:       "00:11:22:33:44:55",
					OperationalState: "up",
					IPAddresses:      []string{"192.168.1.10"},
					SpeedMbps:        1000,
				},
				{
					InterfaceName:    "eth1",
					InterfaceType:    "physical",
					MacAddress:       "00:11:22:33:44:66",
					OperationalState: "up",
					SpeedMbps:        1000,
				},
				{
					InterfaceName:    "eth0.100",
					InterfaceType:    "vlan",
					MacAddress:       "00:11:22:33:44:77",
					VlanID:           100,
					VlanParent:       "eth0",
					OperationalState: "up",
					IPAddresses:      []string{"10.0.100.10"},
				},
			},
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Multiple interfaces submitted successfully",
			},
			wantErr: false,
		},
		{
			name:       "empty server UUID error",
			serverUUID: "",
			interfaces: []NetworkHardwareInfo{
				{
					InterfaceName: "eth0",
					InterfaceType: "physical",
				},
			},
			mockStatus: http.StatusBadRequest,
			mockBody:   StandardResponse{},
			wantErr:    true,
		},
		{
			name:       "server not found",
			serverUUID: "nonexistent-server",
			interfaces: []NetworkHardwareInfo{
				{
					InterfaceName: "eth0",
					InterfaceType: "physical",
				},
			},
			mockStatus: http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Server not found",
			},
			wantErr: true,
		},
		{
			name:       "unauthorized",
			serverUUID: "server-uuid-123",
			interfaces: []NetworkHardwareInfo{
				{
					InterfaceName: "eth0",
					InterfaceType: "physical",
				},
			},
			mockStatus: http.StatusUnauthorized,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Unauthorized",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			serverUUID: "server-uuid-123",
			interfaces: []NetworkHardwareInfo{
				{
					InterfaceName: "eth0",
					InterfaceType: "physical",
				},
			},
			mockStatus: http.StatusInternalServerError,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip server creation for empty UUID test (client-side validation)
			if tt.serverUUID == "" {
				client, err := NewClient(&Config{
					BaseURL: "http://localhost",
					Auth:    AuthConfig{Token: "test-token"},
				})
				require.NoError(t, err)

				result, err := client.NetworkHardware.Submit(context.Background(), tt.serverUUID, tt.interfaces)
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "server UUID is required")
				return
			}

			// Create mock server for other tests
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method and path
				assert.Equal(t, "POST", r.Method)
				assert.Contains(t, r.URL.Path, "/v2/server/")
				assert.Contains(t, r.URL.Path, "/hardware/network")
				assert.Contains(t, r.URL.Path, tt.serverUUID)

				// Verify request body
				var receivedRequest NetworkHardwareRequest
				err := json.NewDecoder(r.Body).Decode(&receivedRequest)
				require.NoError(t, err)
				assert.Equal(t, tt.serverUUID, receivedRequest.ServerUUID)
				assert.Equal(t, len(tt.interfaces), len(receivedRequest.Interfaces))

				// Verify interface details for single interface tests
				if len(tt.interfaces) == 1 {
					assert.Equal(t, tt.interfaces[0].InterfaceName, receivedRequest.Interfaces[0].InterfaceName)
					assert.Equal(t, tt.interfaces[0].InterfaceType, receivedRequest.Interfaces[0].InterfaceType)
				}

				// Send mock response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			// Create client
			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			// Call Submit
			result, err := client.NetworkHardware.Submit(context.Background(), tt.serverUUID, tt.interfaces)

			// Check error
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestNetworkHardwareInfo_JSON tests JSON marshaling and unmarshaling
func TestNetworkHardwareInfo_JSON(t *testing.T) {
	info := &NetworkHardwareInfo{
		InterfaceName:       "eth0",
		InterfaceType:       "physical",
		MacAddress:          "00:11:22:33:44:55",
		Manufacturer:        "Intel",
		DeviceName:          "Ethernet Controller",
		DriverName:          "e1000e",
		DriverVersion:       "3.2.6",
		SpeedMbps:           1000,
		MaxSpeedMbps:        10000,
		SupportedSpeeds:     []int{100, 1000, 10000},
		Duplex:              "full",
		AutoNegotiation:     true,
		LinkDetected:        true,
		CarrierStatus:       true,
		OperationalState:    "up",
		AdministrativeState: "up",
		MTU:                 1500,
		IPAddresses:         []string{"192.168.1.100", "10.0.0.100"},
		IPv6Addresses:       []string{"fe80::1"},
		SubnetMasks:         []string{"255.255.255.0"},
		GatewayAddresses:    []string{"192.168.1.1"},
		DNSServers:          []string{"8.8.8.8", "8.8.4.4"},
		RxBytes:             1024000,
		TxBytes:             512000,
		RxPackets:           1000,
		TxPackets:           500,
		RxErrors:            0,
		TxErrors:            0,
		Status:              "active",
	}

	// Marshal to JSON
	data, err := json.Marshal(info)
	require.NoError(t, err)
	assert.Contains(t, string(data), "eth0")
	assert.Contains(t, string(data), "Intel")

	// Unmarshal from JSON
	var decoded NetworkHardwareInfo
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, info.InterfaceName, decoded.InterfaceName)
	assert.Equal(t, info.InterfaceType, decoded.InterfaceType)
	assert.Equal(t, info.MacAddress, decoded.MacAddress)
	assert.Equal(t, info.Manufacturer, decoded.Manufacturer)
	assert.Equal(t, info.SpeedMbps, decoded.SpeedMbps)
	assert.Equal(t, info.LinkDetected, decoded.LinkDetected)
	assert.Equal(t, len(info.IPAddresses), len(decoded.IPAddresses))
}

// TestNetworkHardwareInfo_BondInterface tests bond-specific fields
func TestNetworkHardwareInfo_BondInterface(t *testing.T) {
	info := &NetworkHardwareInfo{
		InterfaceName:   "bond0",
		InterfaceType:   "bond",
		BondMode:        "802.3ad",
		BondSlaves:      []string{"eth0", "eth1", "eth2", "eth3"},
		BondPrimary:     "eth0",
		BondActiveSlave: "eth0",
		LACPRate:        "fast",
		XmitHashPolicy:  "layer3+4",
		SpeedMbps:       4000,
	}

	// Marshal to JSON
	data, err := json.Marshal(info)
	require.NoError(t, err)
	assert.Contains(t, string(data), "bond0")
	assert.Contains(t, string(data), "802.3ad")

	// Unmarshal from JSON
	var decoded NetworkHardwareInfo
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "bond0", decoded.InterfaceName)
	assert.Equal(t, "bond", decoded.InterfaceType)
	assert.Equal(t, "802.3ad", decoded.BondMode)
	assert.Equal(t, 4, len(decoded.BondSlaves))
	assert.Equal(t, "eth0", decoded.BondPrimary)
}

// TestNetworkHardwareInfo_VLANInterface tests VLAN-specific fields
func TestNetworkHardwareInfo_VLANInterface(t *testing.T) {
	info := &NetworkHardwareInfo{
		InterfaceName: "eth0.100",
		InterfaceType: "vlan",
		VlanID:        100,
		VlanParent:    "eth0",
		NativeVlan:    1,
		AllowedVlans:  []int{100, 200, 300},
	}

	// Marshal to JSON
	data, err := json.Marshal(info)
	require.NoError(t, err)
	assert.Contains(t, string(data), "eth0.100")
	assert.Contains(t, string(data), "vlan")

	// Unmarshal from JSON
	var decoded NetworkHardwareInfo
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "eth0.100", decoded.InterfaceName)
	assert.Equal(t, "vlan", decoded.InterfaceType)
	assert.Equal(t, 100, decoded.VlanID)
	assert.Equal(t, "eth0", decoded.VlanParent)
	assert.Equal(t, 3, len(decoded.AllowedVlans))
}

// TestNetworkHardwareInfo_BridgeInterface tests bridge-specific fields
func TestNetworkHardwareInfo_BridgeInterface(t *testing.T) {
	info := &NetworkHardwareInfo{
		InterfaceName:      "br0",
		InterfaceType:      "bridge",
		BridgePorts:        []string{"eth0", "eth1"},
		BridgeSTP:          true,
		BridgeForwardDelay: 15,
		BridgeHelloTime:    2,
		BridgeMaxAge:       20,
		BridgePriority:     32768,
	}

	// Marshal to JSON
	data, err := json.Marshal(info)
	require.NoError(t, err)
	assert.Contains(t, string(data), "br0")
	assert.Contains(t, string(data), "bridge")

	// Unmarshal from JSON
	var decoded NetworkHardwareInfo
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "br0", decoded.InterfaceName)
	assert.Equal(t, "bridge", decoded.InterfaceType)
	assert.True(t, decoded.BridgeSTP)
	assert.Equal(t, 2, len(decoded.BridgePorts))
	assert.Equal(t, 32768, decoded.BridgePriority)
}

// TestNetworkHardwareInfo_WirelessInterface tests wireless-specific fields
func TestNetworkHardwareInfo_WirelessInterface(t *testing.T) {
	info := &NetworkHardwareInfo{
		InterfaceName:        "wlan0",
		InterfaceType:        "wireless",
		IsWireless:           true,
		WirelessMode:         "Managed",
		WirelessProtocol:     "802.11ac",
		WirelessFrequencyMHz: 5180.0,
		WirelessChannel:      36,
		WirelessSSID:         "MyNetwork",
		WirelessBSSID:        "00:11:22:33:44:55",
		WirelessEncryption:   "WPA2",
		SignalStrengthDBM:    -55.0,
		LinkQualityPercent:   85.0,
		NoiseLevelDBM:        -90.0,
	}

	// Marshal to JSON
	data, err := json.Marshal(info)
	require.NoError(t, err)
	assert.Contains(t, string(data), "wlan0")
	assert.Contains(t, string(data), "wireless")
	assert.Contains(t, string(data), "MyNetwork")

	// Unmarshal from JSON
	var decoded NetworkHardwareInfo
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "wlan0", decoded.InterfaceName)
	assert.Equal(t, "wireless", decoded.InterfaceType)
	assert.True(t, decoded.IsWireless)
	assert.Equal(t, "MyNetwork", decoded.WirelessSSID)
	assert.Equal(t, 36, decoded.WirelessChannel)
	assert.Equal(t, -55.0, decoded.SignalStrengthDBM)
}

// TestNetworkHardwareRequest_JSON tests request marshaling
func TestNetworkHardwareRequest_JSON(t *testing.T) {
	req := &NetworkHardwareRequest{
		ServerUUID: "server-uuid-123",
		Interfaces: []NetworkHardwareInfo{
			{
				InterfaceName:    "eth0",
				InterfaceType:    "physical",
				MacAddress:       "00:11:22:33:44:55",
				OperationalState: "up",
			},
			{
				InterfaceName:    "eth1",
				InterfaceType:    "physical",
				MacAddress:       "00:11:22:33:44:66",
				OperationalState: "down",
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(req)
	require.NoError(t, err)
	assert.Contains(t, string(data), "server-uuid-123")
	assert.Contains(t, string(data), "interfaces")

	// Unmarshal from JSON
	var decoded NetworkHardwareRequest
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "server-uuid-123", decoded.ServerUUID)
	assert.Equal(t, 2, len(decoded.Interfaces))
	assert.Equal(t, "eth0", decoded.Interfaces[0].InterfaceName)
	assert.Equal(t, "eth1", decoded.Interfaces[1].InterfaceName)
}
