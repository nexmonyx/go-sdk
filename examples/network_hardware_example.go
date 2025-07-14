package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/nexmonyx/go-sdk"
)

func main() {
	// Create client with server authentication
	client, err := nexmonyx.NewClient(&nexmonyx.Config{
		BaseURL: os.Getenv("NEXMONYX_API_URL"), // Default: https://api.nexmonyx.com
		Auth: nexmonyx.AuthConfig{
			ServerUUID:   os.Getenv("SERVER_UUID"),
			ServerSecret: os.Getenv("SERVER_SECRET"),
		},
		Debug: true, // Enable debug logging
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Example network interface data
	interfaces := []nexmonyx.NetworkHardwareInfo{
		{
			InterfaceName:    "eth0",
			InterfaceType:    "physical",
			MacAddress:       "00:11:22:33:44:55",
			DriverName:       "e1000e",
			DriverVersion:    "3.2.6-k",
			FirmwareVersion:  "1.2.3",
			SpeedMbps:        1000,
			Duplex:           "full",
			MTU:              1500,
			LinkDetected:     true,
			OperationalState: "up",
			IPAddresses:      []string{"192.168.1.100/24"},
			IPv6Addresses:    []string{"fe80::211:22ff:fe33:4455/64"},
			RxBytes:          1234567890,
			TxBytes:          987654321,
			RxPackets:        1000000,
			TxPackets:        900000,
		},
		{
			InterfaceName:    "bond0",
			InterfaceType:    "bond",
			MacAddress:       "00:11:22:33:44:66",
			BondMode:         "802.3ad",
			BondSlaves:       []string{"eth1", "eth2"},
			BondActiveSlave:  "eth1",
			LACPRate:         "slow",
			SpeedMbps:        2000,
			LinkDetected:     true,
			OperationalState: "up",
			MTU:              1500,
		},
		{
			InterfaceName:      "br0",
			InterfaceType:      "bridge",
			MacAddress:         "00:11:22:33:44:77",
			BridgePorts:        []string{"bond0.100"},
			BridgeSTP:          true,
			BridgeForwardDelay: 1500,
			BridgeHelloTime:    200,
			BridgeMaxAge:       2000,
			LinkDetected:       true,
			OperationalState:   "up",
			IPAddresses:        []string{"10.0.0.1/24"},
		},
		{
			InterfaceName:    "bond0.100",
			InterfaceType:    "vlan",
			VlanID:           100,
			VlanParent:       "bond0",
			MacAddress:       "00:11:22:33:44:66",
			LinkDetected:     true,
			OperationalState: "up",
		},
	}

	// Submit network hardware information
	ctx := context.Background()
	serverUUID := os.Getenv("SERVER_UUID")

	// Check for required environment variables
	if serverUUID == "" {
		log.Fatal("SERVER_UUID environment variable is required")
	}

	serverSecret := os.Getenv("SERVER_SECRET")
	if serverSecret == "" {
		log.Fatal("SERVER_SECRET environment variable is required")
	}

	fmt.Printf("Starting network hardware submission for server: %s\n", serverUUID)
	fmt.Printf("Submitting %d network interfaces...\n", len(interfaces))

	resp, err := client.NetworkHardware.Submit(ctx, serverUUID, interfaces)
	if err != nil {
		fmt.Printf("Failed to submit network hardware: %v\n", err)
		fmt.Printf("Error type: %T\n", err)
		log.Fatalf("Submission failed")
	}

	fmt.Printf("\n=== Network hardware submitted successfully ===\n")
	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Message: %s\n", resp.Message)
	if resp.Data != nil {
		fmt.Printf("Data: %+v\n", resp.Data)
	}
}
