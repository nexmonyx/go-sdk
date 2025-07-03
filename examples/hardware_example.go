package nexmonyx_test

import (
	"context"
	"fmt"
	"log"
	"time"

	nexmonyx "github.com/nexmonyx/go-sdk"
)

// DemoHardwareInventorySubmit demonstrates how to submit hardware inventory data
func DemoHardwareInventorySubmit() {
	// Create client with server credentials (used by agents)
	client, err := nexmonyx.NewClient(&nexmonyx.Config{
		BaseURL: "https://api.nexmonyx.com",
		Auth: nexmonyx.AuthConfig{
			ServerUUID:   "your-server-uuid",
			ServerSecret: "your-server-secret",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Prepare hardware inventory data
	inventory := &nexmonyx.HardwareInventoryRequest{
		ServerUUID:      "your-server-uuid",
		CollectedAt:     time.Now(),
		CollectionMethod: "dmidecode",
		Hardware: nexmonyx.HardwareInventoryInfo{
			System: &nexmonyx.SystemHardwareInfo{
				Manufacturer: "Dell Inc.",
				ProductName:  "PowerEdge R740",
				SerialNumber: "ABC123DEF",
				UUID:         "4C4C4544-0056-3810-8056-B8C04F303932",
			},

			// Motherboard
			Motherboard: &nexmonyx.MotherboardInfo{
				Manufacturer: "Dell Inc.",
				ProductName:  "0Y7WYT",
				SerialNumber: "MB123456",
				BIOS: &nexmonyx.BIOSInfo{
					Vendor:      "Dell Inc.",
					Version:     "2.12.2",
					ReleaseDate: "05/07/2021",
				},
			},

			// CPUs
			CPUs: []nexmonyx.CPUInfo{
				{
					Manufacturer:  "Intel",
					Model:         "Xeon Gold 6230",
					Cores:         20,
					Threads:       40,
					BaseSpeedMHz:  2100,
					MaxSpeedMHz:   3900,
					CacheSizeKB:   28160, // 27.5MB
					Architecture:  "x86_64",
					Socket:        "Socket 0",
				},
				{
					Manufacturer:  "Intel",
					Model:         "Xeon Gold 6230",
					Cores:         20,
					Threads:       40,
					BaseSpeedMHz:  2100,
					MaxSpeedMHz:   3900,
					CacheSizeKB:   28160,
					Architecture:  "x86_64",
					Socket:        "Socket 1",
				},
			},

			// Memory
			Memory: &nexmonyx.MemoryInfo{
				TotalSizeGB:   256,
				TotalSlots:    24,
				UsedSlots:     8,
				MaxCapacityGB: 1536, // 64GB per slot * 24 slots
				Modules: []nexmonyx.MemoryModuleInfo{
					{
						Slot:         "DIMM_A1",
						Manufacturer: "Samsung",
						PartNumber:   "M393A4K40CB2-CTD",
						SizeGB:       32,
						Type:         "DDR4",
						SpeedMHz:     2666,
						FormFactor:   "DIMM",
						SerialNumber: "00CE0123",
					},
					// Add more memory modules as needed
				},
			},

			// Storage
			Storage: []nexmonyx.StorageDeviceInfo{
				{
					Model:           "SSD 860 EVO",
					Vendor:          "Samsung",
					SerialNumber:    "S3Y2NB0K123456",
					SizeGB:          931.5, // ~1TB
					Type:            "SSD",
					Interface:       "SATA",
					FirmwareVersion: "RVT04B6Q",
					SmartStatus:     "Good",
					PowerOnHours:    8760, // 1 year
				},
				{
					Model:           "970 EVO Plus",
					Vendor:          "Samsung",
					SerialNumber:    "S4EVNF0M123456",
					SizeGB:          465.8, // ~500GB
					Type:            "NVMe",
					Interface:       "NVMe",
					FirmwareVersion: "2B2QEXM7",
					SmartStatus:     "Good",
					Temperature:     42,
				},
			},

			// Network Cards
			Network: []nexmonyx.NetworkCardInfo{
				{
					Model:         "Ethernet Controller X710 for 10GbE SFP+",
					Vendor:        "Intel Corporation",
					MACAddress:    "00:1B:21:AB:CD:EF",
					SpeedMbps:     10000, // 10Gbps
					PortCount:     4,
					Driver:        "i40e",
					DriverVersion: "2.17.15",
				},
			},

			// GPUs (if present)
			GPUs: []nexmonyx.GPUInfo{
				{
					Model:         "Tesla V100-PCIE-32GB",
					Vendor:        "NVIDIA",
					MemoryGB:      32,
					Driver:        "nvidia",
					DriverVersion: "525.60.13",
					BusID:         "0000:3b:00.0",
					Temperature:   42,
				},
			},

			// Power Supplies
			PowerSupplies: []nexmonyx.PowerSupplyInfo{
				{
					Model:         "EPP-1100-3AEEA",
					Manufacturer:  "Dell",
					SerialNumber:  "CN1797231B00EL",
					MaxPowerWatts: 1100,
					Type:          "AC",
					Status:        "OK",
					Efficiency:    "94% (Titanium)",
				},
				{
					Model:         "EPP-1100-3AEEA",
					Manufacturer:  "Dell",
					SerialNumber:  "CN1797231B00EM",
					MaxPowerWatts: 1100,
					Type:          "AC",
					Status:        "OK",
					Efficiency:    "94% (Titanium)",
				},
			},
		},
	}

	// Submit the hardware inventory
	response, err := client.HardwareInventory.Submit(context.Background(), inventory)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Hardware inventory submitted successfully!\n")
	fmt.Printf("Server UUID: %s\n", response.ServerUUID)
	fmt.Printf("Timestamp: %s\n", response.Timestamp)
	fmt.Printf("Component counts:\n")
	for component, count := range response.ComponentCounts {
		fmt.Printf("  %s: %d\n", component, count)
	}
}

// DemoHardwareInventoryGet demonstrates how to get hardware inventory
func DemoHardwareInventoryGet() {
	// Create client with JWT token (for UI/admin access)
	client, err := nexmonyx.NewClient(&nexmonyx.Config{
		BaseURL: "https://api.nexmonyx.com",
		Auth: nexmonyx.AuthConfig{
			Token: "your-jwt-token",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Get the hardware inventory for a server
	serverUUID := "your-server-uuid"
	inventory, err := client.HardwareInventory.Get(context.Background(), serverUUID)
	if err != nil {
		log.Fatal(err)
	}

	// Display system information
	if inventory.System != nil {
		fmt.Printf("Server: %s %s\n", inventory.System.Manufacturer, inventory.System.ProductName)
		fmt.Printf("Serial Number: %s\n", inventory.System.SerialNumber)
	}

	// Display CPU information
	fmt.Printf("\nCPUs (%d):\n", len(inventory.CPUs))
	for i, cpu := range inventory.CPUs {
		fmt.Printf("  CPU %d: %s %s (%d cores, %d threads)\n",
			i+1, cpu.Manufacturer, cpu.Model, cpu.Cores, cpu.Threads)
	}

	// Display memory information
	if inventory.Memory != nil {
		fmt.Printf("\nMemory:\n")
		fmt.Printf("  Total Capacity: %.2f GB\n", inventory.Memory.TotalSizeGB)
		fmt.Printf("  Slots Used: %d/%d\n", inventory.Memory.UsedSlots, inventory.Memory.TotalSlots)
	}

	// Display storage devices
	fmt.Printf("\nStorage Devices (%d):\n", len(inventory.Storage))
	for _, storage := range inventory.Storage {
		fmt.Printf("  %s %s (%.2f GB, %s)\n",
			storage.Vendor, storage.Model, storage.SizeGB, storage.Type)
	}

	// Display GPUs if present
	if len(inventory.GPUs) > 0 {
		fmt.Printf("\nGPUs (%d):\n", len(inventory.GPUs))
		for _, gpu := range inventory.GPUs {
			fmt.Printf("  %s %s (%.0f GB VRAM)\n",
				gpu.Vendor, gpu.Model, gpu.MemoryGB)
		}
	}
}

// DemoHardwareInventoryList demonstrates how to list hardware inventory history
func DemoHardwareInventoryList() {
	// Create client
	client, err := nexmonyx.NewClient(&nexmonyx.Config{
		BaseURL: "https://api.nexmonyx.com",
		Auth: nexmonyx.AuthConfig{
			Token: "your-jwt-token",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// List hardware inventory with options
	opts := &nexmonyx.ListOptions{
		Page:      1,
		Limit:     100,
		StartDate: time.Now().AddDate(0, 0, -30).Format(time.RFC3339),
		EndDate:   time.Now().Format(time.RFC3339),
	}

	inventories, meta, err := client.HardwareInventory.List(context.Background(), opts)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d hardware inventory records (page %d of %d)\n",
		len(inventories), meta.Page, meta.TotalPages)

	// Display inventory summaries
	for _, inventory := range inventories {
		if inventory.System != nil {
			fmt.Printf("\nSystem: %s %s\n", inventory.System.Manufacturer, inventory.System.ProductName)
		}
		fmt.Printf("  CPUs: %d\n", len(inventory.CPUs))
		if inventory.Memory != nil {
			fmt.Printf("  Memory: %.2f GB\n", inventory.Memory.TotalSizeGB)
		}
		fmt.Printf("  Storage Devices: %d\n", len(inventory.Storage))
		fmt.Printf("  GPUs: %d\n", len(inventory.GPUs))
	}
}

// DemoHardwareInventorySearch demonstrates how to search hardware
func DemoHardwareInventorySearch() {
	// Create client
	client, err := nexmonyx.NewClient(&nexmonyx.Config{
		BaseURL: "https://api.nexmonyx.com",
		Auth: nexmonyx.AuthConfig{
			Token: "your-jwt-token",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Search for specific hardware
	search := &nexmonyx.HardwareSearch{
		Manufacturer:  "Dell",
		Model:         "PowerEdge",
		ComponentType: "server",
	}

	inventories, meta, err := client.HardwareInventory.Search(context.Background(), search)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d matching systems (page %d of %d)\n",
		len(inventories), meta.Page, meta.TotalPages)

	// Display search results
	for _, inventory := range inventories {
		if inventory.System != nil {
			fmt.Printf("\nSystem: %s %s (S/N: %s)\n",
				inventory.System.Manufacturer,
				inventory.System.ProductName,
				inventory.System.SerialNumber)
		}
	}
}