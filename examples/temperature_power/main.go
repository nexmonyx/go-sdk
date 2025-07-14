package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nexmonyx/go-sdk"
)

func main() {
	// Initialize the SDK client with server credentials
	client, err := nexmonyx.NewClient(&nexmonyx.Config{
		BaseURL: "https://api.nexmonyx.com",
		Auth: nexmonyx.AuthConfig{
			ServerUUID:   "your-server-uuid",
			ServerSecret: "your-server-secret",
		},
		Debug: true,
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Example 1: Submit temperature and power metrics as part of comprehensive metrics
	submitTemperatureAndPowerMetrics(ctx, client)

	// Example 2: Update hardware inventory with temperature sensors
	updateHardwareInventoryWithSensors(ctx, client)

	// Example 3: Submit real-time temperature alerts
	monitorTemperatureThresholds(ctx, client)
}

func submitTemperatureAndPowerMetrics(ctx context.Context, client *nexmonyx.Client) {
	fmt.Println("=== Submitting Temperature and Power Metrics ===")

	// Create temperature metrics
	tempMetrics := nexmonyx.NewTemperatureMetrics()

	// Add CPU temperature sensors
	for i := 0; i < 4; i++ {
		temp := 45.0 + float64(i)*2.0 // Simulate varying temperatures
		sensor := nexmonyx.CreateCPUTemperatureSensor(i, temp)
		tempMetrics.AddTemperatureSensor(sensor)
	}

	// Add system temperature sensors
	tempMetrics.AddTemperatureSensor(nexmonyx.TemperatureSensorData{
		SensorID:      "inlet_temp",
		SensorName:    "Inlet Temperature",
		Temperature:   26.0,
		Status:        "ok",
		Type:          "system",
		Location:      "chassis_front",
		UpperWarning:  35.0,
		UpperCritical: 40.0,
	})

	tempMetrics.AddTemperatureSensor(nexmonyx.TemperatureSensorData{
		SensorID:      "exhaust_temp",
		SensorName:    "Exhaust Temperature",
		Temperature:   46.0,
		Status:        "ok",
		Type:          "system",
		Location:      "chassis_rear",
		UpperWarning:  55.0,
		UpperCritical: 65.0,
	})

	// Add disk temperature sensors
	disks := []string{"/dev/sda", "/dev/sdb", "/dev/sdc"}
	for _, disk := range disks {
		temp := 30.0 + float64(len(disk))*1.5 // Simulate varying temperatures
		sensor := nexmonyx.CreateDiskTemperatureSensor(disk, temp)
		tempMetrics.AddTemperatureSensor(sensor)
	}

	// Create power metrics
	powerMetrics := nexmonyx.NewPowerMetrics()

	// Add power supply metrics
	powerMetrics.AddPowerSupply(nexmonyx.PowerSupplyMetrics{
		ID:            "ps1",
		Name:          "Power Supply 1",
		Status:        "ok",
		PowerWatts:    196.0,
		MaxPowerWatts: 750.0,
		Voltage:       124.0,
		Current:       1.6,
		Efficiency:    94.5,
		Temperature:   38.0,
	})

	powerMetrics.AddPowerSupply(nexmonyx.PowerSupplyMetrics{
		ID:            "ps2",
		Name:          "Power Supply 2",
		Status:        "ok",
		PowerWatts:    196.0,
		MaxPowerWatts: 750.0,
		Voltage:       122.0,
		Current:       1.6,
		Efficiency:    94.2,
		Temperature:   39.0,
	})

	// Display summary
	maxTemp, maxSensor := tempMetrics.GetMaxTemperature()
	fmt.Printf("Max Temperature: %.1f°C (Sensor: %s)\n", maxTemp, maxSensor)
	fmt.Printf("Total Power Draw: %.1fW\n", powerMetrics.TotalPowerW)
	fmt.Printf("Average PSU Efficiency: %.1f%%\n", powerMetrics.CalculateAverageEfficiency())

	// Create comprehensive metrics request
	metricsRequest := &nexmonyx.ComprehensiveMetricsRequest{
		ServerUUID:  "your-server-uuid",
		CollectedAt: time.Now().UTC().Format(time.RFC3339),
		Temperature: tempMetrics,
		Power:       powerMetrics,
		// Include other metrics as needed
		SystemInfo: &nexmonyx.SystemInfo{
			Hostname:      "server-01",
			OS:            "Linux",
			OSVersion:     "Ubuntu 22.04",
			Kernel:        "Linux",
			KernelVersion: "5.15.0-88-generic",
			Architecture:  "x86_64",
			Uptime:        86400,
			BootTime:      time.Now().Add(-24 * time.Hour).Unix(),
			Processes:     150,
		},
	}

	// Submit the metrics
	err := client.Metrics.SubmitComprehensive(ctx, metricsRequest)
	if err != nil {
		log.Printf("Failed to submit metrics: %v", err)
	} else {
		fmt.Println("Successfully submitted temperature and power metrics")
	}
}

func updateHardwareInventoryWithSensors(ctx context.Context, client *nexmonyx.Client) {
	fmt.Println("\n=== Updating Hardware Inventory with Temperature Sensors ===")

	// Create temperature sensor inventory
	tempSensors := []nexmonyx.TemperatureSensorInfo{
		{
			SensorID:   "coretemp_package",
			SensorName: "CPU Package Temperature",
			Type:       "cpu",
			Location:   "processor",
			MaxTemp:    100.0,
			MinTemp:    0.0,
		},
		{
			SensorID:   "pch_temp",
			SensorName: "PCH Temperature",
			Type:       "chipset",
			Location:   "motherboard",
			MaxTemp:    90.0,
			MinTemp:    0.0,
		},
		{
			SensorID:   "dimm_a1_temp",
			SensorName: "DIMM A1 Temperature",
			Type:       "memory",
			Location:   "memory_slot_a1",
			MaxTemp:    85.0,
			MinTemp:    0.0,
		},
		{
			SensorID:   "nvme0_temp",
			SensorName: "NVMe SSD 0 Temperature",
			Type:       "storage",
			Location:   "m2_slot_0",
			MaxTemp:    70.0,
			MinTemp:    0.0,
		},
	}

	// Create enhanced power supply info
	powerSupplies := []nexmonyx.PowerSupplyInfo{
		{
			Model:             "Dell EPP-750AB A",
			Manufacturer:      "Dell",
			SerialNumber:      "CN-0VDTVR-17973-123-4567",
			MaxPowerWatts:     750,
			Type:              "AC",
			Status:            "ok",
			Efficiency:        "80 Plus Platinum",
			CurrentPowerWatts: 196.0,
			Voltage:           124.0,
			Current:           1.6,
			Temperature:       38.0,
			FanSpeed:          2400,
			InputVoltage:      120.0,
			OutputVoltage:     12.0,
		},
		{
			Model:             "Dell EPP-750AB A",
			Manufacturer:      "Dell",
			SerialNumber:      "CN-0VDTVR-17973-123-4568",
			MaxPowerWatts:     750,
			Type:              "AC",
			Status:            "ok",
			Efficiency:        "80 Plus Platinum",
			CurrentPowerWatts: 196.0,
			Voltage:           122.0,
			Current:           1.6,
			Temperature:       39.0,
			FanSpeed:          2450,
			InputVoltage:      120.0,
			OutputVoltage:     12.0,
		},
	}

	// Create hardware inventory request
	// Note: In a real implementation, you would submit this to the appropriate endpoint
	inventoryRequest := &nexmonyx.HardwareInventoryRequest{
		ServerUUID:  "your-server-uuid",
		CollectedAt: time.Now(),
		Hardware: nexmonyx.HardwareInventoryInfo{
			TemperatureSensors: tempSensors,
			PowerSupplies:     powerSupplies,
			// Include other hardware components as needed
			System: &nexmonyx.SystemHardwareInfo{
				Manufacturer: "Dell Inc.",
				ProductName:  "PowerEdge R720xd",
				Version:      "01",
				SerialNumber: "ABC1234",
			},
		},
		CollectionMethod: "agent",
	}

	// Note: You'll need to use the appropriate API endpoint for hardware inventory updates
	// This is a placeholder for the actual API call
	fmt.Printf("Hardware inventory prepared with %d temperature sensors and %d power supplies\n",
		len(inventoryRequest.Hardware.TemperatureSensors), 
		len(inventoryRequest.Hardware.PowerSupplies))
}

func monitorTemperatureThresholds(ctx context.Context, client *nexmonyx.Client) {
	fmt.Println("\n=== Monitoring Temperature Thresholds ===")

	// Simulate temperature monitoring loop
	for i := 0; i < 3; i++ {
		tempMetrics := nexmonyx.NewTemperatureMetrics()

		// Simulate a temperature spike
		cpuTemp := 45.0
		if i == 1 {
			cpuTemp = 82.0 // Warning threshold
		} else if i == 2 {
			cpuTemp = 92.0 // Critical threshold
		}

		sensor := nexmonyx.CreateCPUTemperatureSensor(0, cpuTemp)
		tempMetrics.AddTemperatureSensor(sensor)

		// Check status
		fmt.Printf("Iteration %d: CPU Temperature: %.1f°C - Status: %s\n",
			i+1, cpuTemp, sensor.Status)

		// In a real implementation, you might send alerts or notifications
		// when temperature exceeds thresholds
		if sensor.Status == "warning" {
			fmt.Println("  ⚠️  WARNING: CPU temperature is high!")
		} else if sensor.Status == "critical" {
			fmt.Println("  🚨 CRITICAL: CPU temperature is critical!")
		}

		time.Sleep(2 * time.Second)
	}
}