package nexmonyx

import (
	"context"
	"fmt"
)

// ExampleServerDetailsUpdateWithHardware demonstrates how to use the enhanced
// ServerDetailsUpdateRequest to send detailed hardware information to the API
func ExampleServerDetailsUpdateWithHardware() {
	// Create a client (replace with your actual configuration)
	client, err := NewClient(&Config{
		BaseURL: "https://api.nexmonyx.com",
		Auth: AuthConfig{
			ServerUUID:   "your-server-uuid",
			ServerSecret: "your-server-secret",
		},
	})
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		return
	}

	// Create a comprehensive server details update request
	req := NewServerDetailsUpdateRequest().
		// Basic server information
		WithBasicInfo("web-server-01", "192.168.1.100", "production", "datacenter-east", "web-server").
		// System information
		WithSystemInfo("linux", "Ubuntu 22.04.3 LTS", "x86_64", "SRV-001-ABC123", "00:50:56:c0:00:08").
		// Legacy hardware fields (for backward compatibility)
		WithLegacyHardware("Intel Xeon E5-2680 v4", 2, 28, 67108864, 2000000000000).
		// Enhanced hardware details
		WithCPUs([]ServerCPUInfo{
			{
				PhysicalID:       "0",
				Manufacturer:     "Intel",
				ModelName:        "Intel(R) Xeon(R) CPU E5-2680 v4 @ 2.40GHz",
				Family:           "6",
				Model:            "79",
				Architecture:     "x86_64",
				SocketType:       "LGA2011-3",
				BaseSpeed:        2400.0,
				MaxSpeed:         3300.0,
				PhysicalCores:    14,
				LogicalCores:     28,
				L1Cache:          32,
				L2Cache:          256,
				L3Cache:          35840,
				Virtualization:   "VT-x",
			},
			{
				PhysicalID:       "1",
				Manufacturer:     "Intel",
				ModelName:        "Intel(R) Xeon(R) CPU E5-2680 v4 @ 2.40GHz",
				Family:           "6",
				Model:            "79",
				Architecture:     "x86_64",
				SocketType:       "LGA2011-3",
				BaseSpeed:        2400.0,
				MaxSpeed:         3300.0,
				PhysicalCores:    14,
				LogicalCores:     28,
				L1Cache:          32,
				L2Cache:          256,
				L3Cache:          35840,
				Virtualization:   "VT-x",
			},
		}).
		WithMemory(&ServerMemoryInfo{
			TotalSize:     68719476736, // 64GB
			AvailableSize: 34359738368, // 32GB
			UsedSize:      34359738368, // 32GB
			MemoryType:    "DDR4",
			Speed:         2400,
			ModuleCount:   8,
			ECCSupported:  true,
		}).
		WithNetworkInterfaces([]ServerNetworkInterfaceInfo{
			{
				Name:          "eth0",
				HardwareAddr:  "00:50:56:c0:00:08",
				MTU:           1500,
				Flags:         "up|broadcast|running|multicast",
				Addrs:         "192.168.1.100/24",
				SpeedMbps:     1000,
				IsUp:          true,
				IsWireless:    false,
			},
			{
				Name:          "eth1",
				HardwareAddr:  "00:50:56:c0:00:09",
				MTU:           1500,
				Flags:         "up|broadcast|running|multicast",
				Addrs:         "10.0.1.100/24",
				SpeedMbps:     1000,
				IsUp:          true,
				IsWireless:    false,
			},
		}).
		WithDisks([]ServerDiskInfo{
			{
				Device:       "/dev/sda",
				DiskModel:    "Samsung SSD 980 PRO",
				SerialNumber: "S5P2NS0R123456",
				Size:         1000204886016, // ~1TB
				Type:         "NVMe",
				Vendor:       "Samsung",
			},
			{
				Device:       "/dev/sdb",
				DiskModel:    "WD Red Plus WD40EFPX",
				SerialNumber: "WD-WX12345678901",
				Size:         4000787030016, // ~4TB
				Type:         "HDD",
				Vendor:       "Western Digital",
			},
		})

	// Send the update to the API
	ctx := context.Background()
	server, err := client.Servers.UpdateDetails(ctx, "your-server-uuid", req)
	if err != nil {
		fmt.Printf("Failed to update server details: %v\n", err)
		return
	}

	fmt.Printf("Successfully updated server: %s (ID: %d)\n", server.Hostname, server.ID)
}

// ExampleDiskOnlyUpdate demonstrates how to send just disk information
// This is the primary use case for collecting individual disk metrics
func ExampleDiskOnlyUpdate() {
	client, err := NewClient(&Config{
		BaseURL: "https://api.nexmonyx.com",
		Auth: AuthConfig{
			ServerUUID:   "your-server-uuid",
			ServerSecret: "your-server-secret",
		},
		Debug: true, // Enable debug logging to see detailed hardware info in logs
	})
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		return
	}

	// Create a minimal request with just disk information
	req := NewServerDetailsUpdateRequest().
		WithDisks([]ServerDiskInfo{
			{
				Device:       "/dev/nvme0n1",
				DiskModel:    "Samsung SSD 980 PRO 1TB",
				SerialNumber: "S5P2NS0R789012",
				Size:         1000204886016,
				Type:         "NVMe",
				Vendor:       "Samsung",
			},
			{
				Device:       "/dev/nvme1n1",
				DiskModel:    "WD Black SN750 2TB",
				SerialNumber: "WD-WX98765432109",
				Size:         2000398934016,
				Type:         "NVMe",
				Vendor:       "Western Digital",
			},
		})

	// Check if the request has the required disk information
	if !req.HasDisks() {
		fmt.Println("No disk information to send")
		return
	}

	fmt.Printf("Sending disk information for %d disks\n", len(req.Hardware.Disks))

	// Send the update
	ctx := context.Background()
	server, err := client.Servers.UpdateDetails(ctx, "your-server-uuid", req)
	if err != nil {
		fmt.Printf("Failed to update server with disk details: %v\n", err)
		return
	}

	fmt.Printf("Successfully updated server with disk information: %s\n", server.Hostname)
}

// ExampleBackwardCompatibility demonstrates that legacy hardware fields still work
func ExampleBackwardCompatibility() {
	client, err := NewClient(&Config{
		BaseURL: "https://api.nexmonyx.com",
		Auth: AuthConfig{
			ServerUUID:   "your-server-uuid",
			ServerSecret: "your-server-secret",
		},
	})
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		return
	}

	// Create a request using only legacy fields
	req := NewServerDetailsUpdateRequest().
		WithBasicInfo("legacy-server", "192.168.1.200", "staging", "datacenter-west", "database").
		WithSystemInfo("linux", "CentOS 7", "x86_64", "SRV-002-XYZ789", "00:50:56:c0:00:10").
		WithLegacyHardware("AMD EPYC 7742", 1, 128, 134217728, 8000000000000) // 128GB RAM, 8TB storage

	// This request will work without any enhanced hardware details
	if req.HasHardwareDetails() {
		fmt.Println("Has enhanced hardware details")
	} else {
		fmt.Println("Using legacy hardware fields only")
	}

	ctx := context.Background()
	server, err := client.Servers.UpdateDetails(ctx, "your-server-uuid", req)
	if err != nil {
		fmt.Printf("Failed to update server with legacy details: %v\n", err)
		return
	}

	fmt.Printf("Successfully updated server using legacy fields: %s\n", server.Hostname)
}

// ExampleDynamicHardwareCollection shows how to programmatically build hardware information
func ExampleDynamicHardwareCollection() {
	client, err := NewClient(&Config{
		BaseURL: "https://api.nexmonyx.com",
		Auth: AuthConfig{
			ServerUUID:   "your-server-uuid",
			ServerSecret: "your-server-secret",
		},
	})
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		return
	}

	// Start with a basic request
	req := NewServerDetailsUpdateRequest().
		WithBasicInfo("dynamic-server", "192.168.1.300", "production", "datacenter-central", "compute")

	// Simulate collecting disk information from system
	disks := []ServerDiskInfo{}
	
	// Example: add disks based on detection logic
	detectedDisks := []string{"/dev/sda", "/dev/sdb", "/dev/nvme0n1"}
	for i, device := range detectedDisks {
		disk := ServerDiskInfo{
			Device: device,
		}
		
		// Simulate disk detection logic
		if device[:8] == "/dev/nvm" {
			disk.Type = "NVMe"
			disk.DiskModel = fmt.Sprintf("NVMe SSD %d", i+1)
			disk.Size = 1000204886016 // 1TB
			disk.Vendor = "Generic NVMe"
		} else {
			disk.Type = "SATA"
			disk.DiskModel = fmt.Sprintf("SATA Drive %d", i+1)
			disk.Size = 2000398934016 // 2TB
			disk.Vendor = "Generic SATA"
		}
		disk.SerialNumber = fmt.Sprintf("SN%06d", 100000+i)
		
		disks = append(disks, disk)
	}

	// Add the collected disks to the request
	req = req.WithDisks(disks)

	fmt.Printf("Collected %d disks dynamically\n", len(disks))

	// Send the update
	ctx := context.Background()
	server, err := client.Servers.UpdateDetails(ctx, "your-server-uuid", req)
	if err != nil {
		fmt.Printf("Failed to update server with dynamic hardware: %v\n", err)
		return
	}

	fmt.Printf("Successfully updated server with dynamically collected hardware: %s\n", server.Hostname)
}