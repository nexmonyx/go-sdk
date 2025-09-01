package nexmonyx

import (
	"encoding/json"
	"testing"
)

// TestHardwareDetailsMarshaling tests that the enhanced hardware structures marshal correctly to JSON
func TestHardwareDetailsMarshaling(t *testing.T) {
	// Create a comprehensive ServerDetailsUpdateRequest with hardware details
	req := NewServerDetailsUpdateRequest().
		WithBasicInfo("test-server", "192.168.1.100", "production", "datacenter-1", "web-server").
		WithSystemInfo("linux", "Ubuntu 22.04", "x86_64", "ABC123", "00:11:22:33:44:55").
		WithLegacyHardware("Intel Xeon", 2, 8, 32768, 1000000).
		WithCPUs([]ServerCPUInfo{
			{
				PhysicalID:       "0",
				Manufacturer:     "Intel",
				ModelName:        "Intel(R) Xeon(R) CPU E5-2680 v4 @ 2.40GHz",
				Family:           "6",
				Model:            "79",
				Stepping:         "1",
				Architecture:     "x86_64",
				SocketType:       "LGA2011",
				BaseSpeed:        2400.0,
				MaxSpeed:         3300.0,
				CurrentSpeed:     2400.0,
				SocketCount:      1,
				PhysicalCores:    14,
				LogicalCores:     28,
				L1Cache:          32,
				L2Cache:          256,
				L3Cache:          35840,
				Virtualization:   "VT-x",
				Usage:            45.2,
				Temperature:      62.5,
			},
		}).
		WithMemory(&ServerMemoryInfo{
			TotalSize:     34359738368, // 32GB
			AvailableSize: 16106127360, // 15GB
			UsedSize:      18253611008, // 17GB
			MemoryType:    "DDR4",
			Speed:         2400,
			ModuleCount:   4,
			ECCSupported:  true,
		}).
		WithNetworkInterfaces([]ServerNetworkInterfaceInfo{
			{
				Name:          "eth0",
				HardwareAddr:  "00:11:22:33:44:55",
				MTU:           1500,
				Flags:         "up|broadcast|running|multicast",
				Addrs:         "192.168.1.100/24",
				BytesReceived: 1024000000,
				BytesSent:     512000000,
				SpeedMbps:     1000,
				IsUp:          true,
				IsWireless:    false,
			},
			{
				Name:          "wlan0",
				HardwareAddr:  "aa:bb:cc:dd:ee:ff",
				MTU:           1500,
				Flags:         "up|broadcast|running|multicast",
				Addrs:         "10.0.0.50/24",
				BytesReceived: 50000000,
				BytesSent:     25000000,
				SpeedMbps:     300,
				IsUp:          true,
				IsWireless:    true,
			},
		}).
		WithDisks([]ServerDiskInfo{
			{
				Device:       "/dev/sda",
				DiskModel:    "Samsung SSD 970 EVO Plus",
				SerialNumber: "S4EWNX0R123456",
				Size:         1000204886016, // ~1TB
				Type:         "SSD",
				Vendor:       "Samsung",
			},
			{
				Device:       "/dev/sdb",
				DiskModel:    "Seagate Barracuda ST2000DM008",
				SerialNumber: "ZFL12345",
				Size:         2000398934016, // ~2TB
				Type:         "HDD",
				Vendor:       "Seagate",
			},
		})

	// Test JSON marshaling
	jsonData, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal ServerDetailsUpdateRequest to JSON: %v", err)
	}

	t.Logf("Marshaled JSON:\n%s", string(jsonData))

	// Test that we can unmarshal it back
	var unmarshaledReq ServerDetailsUpdateRequest
	err = json.Unmarshal(jsonData, &unmarshaledReq)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON back to ServerDetailsUpdateRequest: %v", err)
	}

	// Verify basic fields
	if unmarshaledReq.Hostname != "test-server" {
		t.Errorf("Expected hostname 'test-server', got '%s'", unmarshaledReq.Hostname)
	}

	if unmarshaledReq.MainIP != "192.168.1.100" {
		t.Errorf("Expected main_ip '192.168.1.100', got '%s'", unmarshaledReq.MainIP)
	}

	// Verify hardware details are present
	if !unmarshaledReq.HasHardwareDetails() {
		t.Error("Expected hardware details to be present")
	}

	if !unmarshaledReq.HasDisks() {
		t.Error("Expected disk information to be present")
	}

	// Verify CPU information
	if len(unmarshaledReq.Hardware.CPU) != 1 {
		t.Errorf("Expected 1 CPU, got %d", len(unmarshaledReq.Hardware.CPU))
	} else {
		cpu := unmarshaledReq.Hardware.CPU[0]
		if cpu.Manufacturer != "Intel" {
			t.Errorf("Expected CPU manufacturer 'Intel', got '%s'", cpu.Manufacturer)
		}
		if cpu.PhysicalCores != 14 {
			t.Errorf("Expected 14 physical cores, got %d", cpu.PhysicalCores)
		}
	}

	// Verify Memory information
	if unmarshaledReq.Hardware.Memory == nil {
		t.Error("Expected memory information to be present")
	} else {
		if unmarshaledReq.Hardware.Memory.TotalSize != 34359738368 {
			t.Errorf("Expected memory total size 34359738368, got %d", unmarshaledReq.Hardware.Memory.TotalSize)
		}
		if unmarshaledReq.Hardware.Memory.MemoryType != "DDR4" {
			t.Errorf("Expected memory type 'DDR4', got '%s'", unmarshaledReq.Hardware.Memory.MemoryType)
		}
	}

	// Verify Network interfaces
	if len(unmarshaledReq.Hardware.Network) != 2 {
		t.Errorf("Expected 2 network interfaces, got %d", len(unmarshaledReq.Hardware.Network))
	} else {
		eth0 := unmarshaledReq.Hardware.Network[0]
		if eth0.Name != "eth0" {
			t.Errorf("Expected interface name 'eth0', got '%s'", eth0.Name)
		}
		if eth0.SpeedMbps != 1000 {
			t.Errorf("Expected interface speed 1000 Mbps, got %d", eth0.SpeedMbps)
		}
	}

	// Verify Disk information
	if len(unmarshaledReq.Hardware.Disks) != 2 {
		t.Errorf("Expected 2 disks, got %d", len(unmarshaledReq.Hardware.Disks))
	} else {
		ssd := unmarshaledReq.Hardware.Disks[0]
		if ssd.Device != "/dev/sda" {
			t.Errorf("Expected disk device '/dev/sda', got '%s'", ssd.Device)
		}
		if ssd.Type != "SSD" {
			t.Errorf("Expected disk type 'SSD', got '%s'", ssd.Type)
		}
		if ssd.Vendor != "Samsung" {
			t.Errorf("Expected disk vendor 'Samsung', got '%s'", ssd.Vendor)
		}
	}
}

// TestAPICompatibilityStructure tests that the JSON structure matches API expectations
func TestAPICompatibilityStructure(t *testing.T) {
	// Create a minimal request with just disk information (our primary use case)
	req := NewServerDetailsUpdateRequest().
		WithDisks([]ServerDiskInfo{
			{
				Device:       "/dev/nvme0n1",
				DiskModel:    "WD Black SN750",
				SerialNumber: "WD-WX12345678901",
				Size:         512110190592, // ~512GB
				Type:         "NVMe",
				Vendor:       "Western Digital",
			},
		})

	jsonData, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Parse the JSON to verify structure
	var jsonStructure map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonStructure)
	if err != nil {
		t.Fatalf("Failed to parse JSON structure: %v", err)
	}

	// Verify the hardware section exists
	hardware, exists := jsonStructure["hardware"]
	if !exists {
		t.Error("Expected 'hardware' field in JSON structure")
	}

	// Verify hardware is an object
	hardwareObj, ok := hardware.(map[string]interface{})
	if !ok {
		t.Error("Expected 'hardware' to be an object")
	}

	// Verify disks array exists
	disks, exists := hardwareObj["disks"]
	if !exists {
		t.Error("Expected 'disks' field in hardware object")
	}

	// Verify disks is an array
	disksArray, ok := disks.([]interface{})
	if !ok {
		t.Error("Expected 'disks' to be an array")
	}

	if len(disksArray) != 1 {
		t.Errorf("Expected 1 disk, got %d", len(disksArray))
	}

	// Verify disk structure
	disk0, ok := disksArray[0].(map[string]interface{})
	if !ok {
		t.Error("Expected disk to be an object")
	}

	expectedFields := []string{"device", "disk_model", "serial_number", "size", "type", "vendor"}
	for _, field := range expectedFields {
		if _, exists := disk0[field]; !exists {
			t.Errorf("Expected disk field '%s' to be present", field)
		}
	}

	t.Logf("JSON structure test passed. Generated JSON: %s", string(jsonData))
}

// TestBuilderPattern tests the fluent builder pattern
func TestBuilderPattern(t *testing.T) {
	// Test that all methods return the same instance for chaining
	req1 := NewServerDetailsUpdateRequest()
	req2 := req1.WithBasicInfo("test", "1.1.1.1", "prod", "dc1", "web")
	req3 := req2.WithSystemInfo("linux", "ubuntu", "amd64", "serial", "mac")
	
	// All should be the same instance
	if req1 != req2 || req2 != req3 {
		t.Error("Builder methods should return the same instance for chaining")
	}

	// Test that hardware details are initialized properly
	req4 := req3.WithCPUs([]ServerCPUInfo{{Manufacturer: "Intel"}})
	if !req4.HasHardwareDetails() {
		t.Error("Expected hardware details to be present after adding CPUs")
	}

	req5 := req4.WithDisks([]ServerDiskInfo{{Device: "/dev/sda"}})
	if !req5.HasDisks() {
		t.Error("Expected disks to be present after adding disk information")
	}
}

// TestEmptyHardwareDetails tests behavior with empty hardware details
func TestEmptyHardwareDetails(t *testing.T) {
	req := NewServerDetailsUpdateRequest()
	
	if req.HasHardwareDetails() {
		t.Error("Expected no hardware details for new request")
	}

	if req.HasDisks() {
		t.Error("Expected no disks for new request")
	}

	// Adding empty hardware should still create the structure
	req.WithHardwareDetails(&HardwareDetails{})
	if !req.HasHardwareDetails() {
		t.Error("Expected hardware details to be present after adding empty hardware")
	}

	if req.HasDisks() {
		t.Error("Expected no disks for empty hardware details")
	}
}

// BenchmarkJSONMarshaling benchmarks the JSON marshaling performance
func BenchmarkJSONMarshaling(b *testing.B) {
	req := NewServerDetailsUpdateRequest().
		WithBasicInfo("test-server", "192.168.1.100", "production", "datacenter-1", "web-server").
		WithDisks([]ServerDiskInfo{
			{Device: "/dev/sda", DiskModel: "Test SSD", Size: 1000000000000, Type: "SSD", Vendor: "TestVendor"},
			{Device: "/dev/sdb", DiskModel: "Test HDD", Size: 2000000000000, Type: "HDD", Vendor: "TestVendor"},
		})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(req)
		if err != nil {
			b.Fatalf("Failed to marshal: %v", err)
		}
	}
}