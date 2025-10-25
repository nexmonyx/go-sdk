package nexmonyx_test

import (
	"testing"

	"github.com/nexmonyx/go-sdk/v2"
	"github.com/stretchr/testify/assert"
)

// Additional tests to improve temperature_power_helpers.go coverage from 62.17% to 70%+

func TestTemperatureMetrics_GetSensorByID(t *testing.T) {
	metrics := nexmonyx.NewTemperatureMetrics()

	// Add some sensors
	sensor1 := nexmonyx.TemperatureSensorData{
		SensorID:    "cpu_0",
		SensorName:  "CPU Core 0",
		Temperature: 45.0,
		Type:        "cpu",
	}
	sensor2 := nexmonyx.TemperatureSensorData{
		SensorID:    "cpu_1",
		SensorName:  "CPU Core 1",
		Temperature: 47.0,
		Type:        "cpu",
	}

	metrics.AddTemperatureSensor(sensor1)
	metrics.AddTemperatureSensor(sensor2)

	t.Run("Found", func(t *testing.T) {
		found := metrics.GetSensorByID("cpu_0")
		assert.NotNil(t, found)
		assert.Equal(t, "cpu_0", found.SensorID)
		assert.Equal(t, 45.0, found.Temperature)
	})

	t.Run("NotFound", func(t *testing.T) {
		notFound := metrics.GetSensorByID("nonexistent")
		assert.Nil(t, notFound)
	})

	t.Run("EmptyMetrics", func(t *testing.T) {
		emptyMetrics := nexmonyx.NewTemperatureMetrics()
		result := emptyMetrics.GetSensorByID("any")
		assert.Nil(t, result)
	})
}

func TestTemperatureMetrics_GetSensorsByType(t *testing.T) {
	metrics := nexmonyx.NewTemperatureMetrics()

	// Add various sensor types
	metrics.AddTemperatureSensor(nexmonyx.TemperatureSensorData{
		SensorID: "cpu_0",
		Type:     "cpu",
	})
	metrics.AddTemperatureSensor(nexmonyx.TemperatureSensorData{
		SensorID: "cpu_1",
		Type:     "cpu",
	})
	metrics.AddTemperatureSensor(nexmonyx.TemperatureSensorData{
		SensorID: "disk_0",
		Type:     "disk",
	})
	metrics.AddTemperatureSensor(nexmonyx.TemperatureSensorData{
		SensorID: "system_0",
		Type:     "system",
	})

	t.Run("CPUSensors", func(t *testing.T) {
		cpuSensors := metrics.GetSensorsByType("cpu")
		assert.Len(t, cpuSensors, 2)
		for _, s := range cpuSensors {
			assert.Equal(t, "cpu", s.Type)
		}
	})

	t.Run("DiskSensors", func(t *testing.T) {
		diskSensors := metrics.GetSensorsByType("disk")
		assert.Len(t, diskSensors, 1)
		assert.Equal(t, "disk_0", diskSensors[0].SensorID)
	})

	t.Run("NoMatch", func(t *testing.T) {
		noMatch := metrics.GetSensorsByType("gpu")
		assert.Empty(t, noMatch)
	})

	t.Run("EmptyMetrics", func(t *testing.T) {
		emptyMetrics := nexmonyx.NewTemperatureMetrics()
		result := emptyMetrics.GetSensorsByType("cpu")
		assert.Empty(t, result)
	})
}

func TestPowerMetrics_GetPowerSupplyByID(t *testing.T) {
	metrics := nexmonyx.NewPowerMetrics()

	ps1 := nexmonyx.PowerSupplyMetrics{
		ID:         "ps1",
		Name:       "Power Supply 1",
		Status:     "ok",
		PowerWatts: 250.0,
	}
	ps2 := nexmonyx.PowerSupplyMetrics{
		ID:         "ps2",
		Name:       "Power Supply 2",
		Status:     "ok",
		PowerWatts: 230.0,
	}

	metrics.AddPowerSupply(ps1)
	metrics.AddPowerSupply(ps2)

	t.Run("Found", func(t *testing.T) {
		found := metrics.GetPowerSupplyByID("ps1")
		assert.NotNil(t, found)
		assert.Equal(t, "ps1", found.ID)
		assert.Equal(t, 250.0, found.PowerWatts)
	})

	t.Run("NotFound", func(t *testing.T) {
		notFound := metrics.GetPowerSupplyByID("ps3")
		assert.Nil(t, notFound)
	})

	t.Run("EmptyMetrics", func(t *testing.T) {
		emptyMetrics := nexmonyx.NewPowerMetrics()
		result := emptyMetrics.GetPowerSupplyByID("any")
		assert.Nil(t, result)
	})
}

func TestPowerMetrics_GetFailedPowerSupplies(t *testing.T) {
	metrics := nexmonyx.NewPowerMetrics()

	metrics.PowerSupplies = []nexmonyx.PowerSupplyMetrics{
		{ID: "ps1", Status: "ok", PowerWatts: 250.0},
		{ID: "ps2", Status: "failed", PowerWatts: 0.0},
		{ID: "ps3", Status: "critical", PowerWatts: 100.0},
		{ID: "ps4", Status: "warning", PowerWatts: 200.0},
		{ID: "ps5", Status: "ok", PowerWatts: 240.0},
	}

	t.Run("FailedAndCritical", func(t *testing.T) {
		failed := metrics.GetFailedPowerSupplies()
		assert.Len(t, failed, 2)
		ids := []string{failed[0].ID, failed[1].ID}
		assert.Contains(t, ids, "ps2")
		assert.Contains(t, ids, "ps3")
	})

	t.Run("NoFailures", func(t *testing.T) {
		healthyMetrics := nexmonyx.NewPowerMetrics()
		healthyMetrics.PowerSupplies = []nexmonyx.PowerSupplyMetrics{
			{ID: "ps1", Status: "ok"},
			{ID: "ps2", Status: "ok"},
		}
		failed := healthyMetrics.GetFailedPowerSupplies()
		assert.Empty(t, failed)
	})

	t.Run("EmptyMetrics", func(t *testing.T) {
		emptyMetrics := nexmonyx.NewPowerMetrics()
		failed := emptyMetrics.GetFailedPowerSupplies()
		assert.Empty(t, failed)
	})
}

func TestPowerMetrics_CalculateAverageEfficiency_EdgeCases(t *testing.T) {
	t.Run("AllZeroEfficiency", func(t *testing.T) {
		metrics := nexmonyx.NewPowerMetrics()
		metrics.PowerSupplies = []nexmonyx.PowerSupplyMetrics{
			{Efficiency: 0.0},
			{Efficiency: 0.0},
			{Efficiency: 0.0},
		}
		avg := metrics.CalculateAverageEfficiency()
		assert.Equal(t, 0.0, avg)
	})

	t.Run("MixedZeroAndNonZero", func(t *testing.T) {
		metrics := nexmonyx.NewPowerMetrics()
		metrics.PowerSupplies = []nexmonyx.PowerSupplyMetrics{
			{Efficiency: 95.0},
			{Efficiency: 0.0},
			{Efficiency: 93.0},
			{Efficiency: 0.0},
		}
		avg := metrics.CalculateAverageEfficiency()
		assert.InDelta(t, 94.0, avg, 0.01)
	})
}

func TestCreateSystemTemperatureSensor(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		sensor := nexmonyx.CreateSystemTemperatureSensor("Motherboard", 45.0)
		assert.Equal(t, "system_Motherboard", sensor.SensorID)
		assert.Equal(t, "Motherboard", sensor.SensorName)
		assert.Equal(t, 45.0, sensor.Temperature)
		assert.Equal(t, "ok", sensor.Status)
		assert.Equal(t, "system", sensor.Type)
		assert.Equal(t, 60.0, sensor.UpperWarning)
		assert.Equal(t, 75.0, sensor.UpperCritical)
	})

	t.Run("Warning", func(t *testing.T) {
		sensor := nexmonyx.CreateSystemTemperatureSensor("PCH", 65.0)
		assert.Equal(t, "warning", sensor.Status)
	})

	t.Run("Critical", func(t *testing.T) {
		sensor := nexmonyx.CreateSystemTemperatureSensor("Chipset", 80.0)
		assert.Equal(t, "critical", sensor.Status)
	})
}

func TestDetermineTemperatureStatus_EdgeCases(t *testing.T) {
	t.Run("ExactlyAtWarning", func(t *testing.T) {
		status := nexmonyx.DetermineTemperatureStatus(75.0, 75.0, 90.0)
		assert.Equal(t, "warning", status)
	})

	t.Run("ExactlyAtCritical", func(t *testing.T) {
		status := nexmonyx.DetermineTemperatureStatus(90.0, 75.0, 90.0)
		assert.Equal(t, "critical", status)
	})

	t.Run("BothThresholdsZero", func(t *testing.T) {
		status := nexmonyx.DetermineTemperatureStatus(100.0, 0.0, 0.0)
		assert.Equal(t, "ok", status)
	})

	t.Run("OnlyWarningSet", func(t *testing.T) {
		status := nexmonyx.DetermineTemperatureStatus(80.0, 75.0, 0.0)
		assert.Equal(t, "warning", status)
	})
}

func TestCreateCPUTemperatureSensor_EdgeCases(t *testing.T) {
	t.Run("HighCoreNumber", func(t *testing.T) {
		sensor := nexmonyx.CreateCPUTemperatureSensor(31, 50.0)
		assert.Equal(t, "cpu_core_31", sensor.SensorID)
		assert.Equal(t, "CPU Core 31", sensor.SensorName)
		assert.Equal(t, "processor", sensor.Location)
	})

	t.Run("CriticalTemp", func(t *testing.T) {
		sensor := nexmonyx.CreateCPUTemperatureSensor(0, 95.0)
		assert.Equal(t, "critical", sensor.Status)
	})

	t.Run("WarningTemp", func(t *testing.T) {
		sensor := nexmonyx.CreateCPUTemperatureSensor(0, 80.0)
		assert.Equal(t, "warning", sensor.Status)
	})
}

func TestCreateDiskTemperatureSensor_EdgeCases(t *testing.T) {
	t.Run("NVMeDrive", func(t *testing.T) {
		sensor := nexmonyx.CreateDiskTemperatureSensor("nvme0n1", 45.0)
		assert.Equal(t, "disk_nvme0n1", sensor.SensorID)
		assert.Equal(t, "Disk nvme0n1", sensor.SensorName)
		assert.Equal(t, "storage", sensor.Location)
		assert.Equal(t, "ok", sensor.Status)
	})

	t.Run("CriticalDiskTemp", func(t *testing.T) {
		sensor := nexmonyx.CreateDiskTemperatureSensor("sda", 65.0)
		assert.Equal(t, "critical", sensor.Status)
	})
}

func TestTemperatureMetrics_AddTemperatureSensor_NilSensors(t *testing.T) {
	metrics := &nexmonyx.TemperatureMetrics{}
	// Sensors is nil initially

	sensor := nexmonyx.TemperatureSensorData{
		SensorID:    "test",
		Temperature: 50.0,
	}

	metrics.AddTemperatureSensor(sensor)
	assert.NotNil(t, metrics.Sensors)
	assert.Len(t, metrics.Sensors, 1)
}

func TestPowerMetrics_AddPowerSupply_NilSupplies(t *testing.T) {
	metrics := &nexmonyx.PowerMetrics{}
	// PowerSupplies is nil initially

	ps := nexmonyx.PowerSupplyMetrics{
		ID:         "ps1",
		PowerWatts: 250.0,
	}

	metrics.AddPowerSupply(ps)
	assert.NotNil(t, metrics.PowerSupplies)
	assert.Len(t, metrics.PowerSupplies, 1)
	assert.Equal(t, 250.0, metrics.TotalPowerW)
}

func TestPowerMetrics_CalculateTotalPower(t *testing.T) {
	metrics := nexmonyx.NewPowerMetrics()

	metrics.PowerSupplies = []nexmonyx.PowerSupplyMetrics{
		{PowerWatts: 250.5},
		{PowerWatts: 230.2},
		{PowerWatts: 180.3},
	}

	total := metrics.CalculateTotalPower()
	assert.InDelta(t, 661.0, total, 0.1)
}

// TestTemperatureMetrics_GetMaxTemperature tests the GetMaxTemperature helper method
func TestTemperatureMetrics_GetMaxTemperature(t *testing.T) {
	tests := []struct {
		name             string
		sensors          []nexmonyx.TemperatureSensorData
		expectedTemp     float64
		expectedSensor   string
		emptyCollection  bool
	}{
		{
			name: "SingleSensor",
			sensors: []nexmonyx.TemperatureSensorData{
				{SensorID: "cpu_0", SensorName: "CPU Core 0", Temperature: 55.0},
			},
			expectedTemp:   55.0,
			expectedSensor: "CPU Core 0",
		},
		{
			name: "MultipleSensorsMaxAtBeginning",
			sensors: []nexmonyx.TemperatureSensorData{
				{SensorID: "cpu_0", SensorName: "CPU Core 0", Temperature: 85.0},
				{SensorID: "cpu_1", SensorName: "CPU Core 1", Temperature: 65.0},
				{SensorID: "disk_0", SensorName: "Disk sda", Temperature: 45.0},
			},
			expectedTemp:   85.0,
			expectedSensor: "CPU Core 0",
		},
		{
			name: "MultipleSensorsMaxInMiddle",
			sensors: []nexmonyx.TemperatureSensorData{
				{SensorID: "cpu_0", SensorName: "CPU Core 0", Temperature: 65.0},
				{SensorID: "cpu_1", SensorName: "CPU Core 1", Temperature: 92.0},
				{SensorID: "disk_0", SensorName: "Disk sda", Temperature: 45.0},
			},
			expectedTemp:   92.0,
			expectedSensor: "CPU Core 1",
		},
		{
			name: "MultipleSensorsMaxAtEnd",
			sensors: []nexmonyx.TemperatureSensorData{
				{SensorID: "cpu_0", SensorName: "CPU Core 0", Temperature: 65.0},
				{SensorID: "cpu_1", SensorName: "CPU Core 1", Temperature: 72.0},
				{SensorID: "disk_0", SensorName: "Disk sda", Temperature: 88.0},
			},
			expectedTemp:   88.0,
			expectedSensor: "Disk sda",
		},
		{
			name: "AllSensorsEqualTemp",
			sensors: []nexmonyx.TemperatureSensorData{
				{SensorID: "cpu_0", SensorName: "CPU Core 0", Temperature: 50.0},
				{SensorID: "cpu_1", SensorName: "CPU Core 1", Temperature: 50.0},
				{SensorID: "cpu_2", SensorName: "CPU Core 2", Temperature: 50.0},
			},
			expectedTemp:   50.0,
			expectedSensor: "CPU Core 0",
		},
		{
			name: "NegativeTemperatures",
			sensors: []nexmonyx.TemperatureSensorData{
				{SensorID: "external_0", SensorName: "External Sensor 1", Temperature: -10.0},
				{SensorID: "external_1", SensorName: "External Sensor 2", Temperature: -5.0},
				{SensorID: "external_2", SensorName: "External Sensor 3", Temperature: -15.0},
			},
			expectedTemp:   -5.0,
			expectedSensor: "External Sensor 2",
		},
		{
			name: "MixedPositiveNegativeTemperatures",
			sensors: []nexmonyx.TemperatureSensorData{
				{SensorID: "sensor_0", SensorName: "Sensor 0", Temperature: -10.0},
				{SensorID: "sensor_1", SensorName: "Sensor 1", Temperature: 0.0},
				{SensorID: "sensor_2", SensorName: "Sensor 2", Temperature: 25.0},
			},
			expectedTemp:   25.0,
			expectedSensor: "Sensor 2",
		},
		{
			name: "ZeroTemperatures",
			sensors: []nexmonyx.TemperatureSensorData{
				{SensorID: "sensor_0", SensorName: "Sensor 0", Temperature: 0.0},
				{SensorID: "sensor_1", SensorName: "Sensor 1", Temperature: 0.0},
			},
			expectedTemp:   0.0,
			expectedSensor: "Sensor 0",
		},
		{
			name: "VeryHighTemperatures",
			sensors: []nexmonyx.TemperatureSensorData{
				{SensorID: "cpu_0", SensorName: "CPU Core 0", Temperature: 95.5},
				{SensorID: "cpu_1", SensorName: "CPU Core 1", Temperature: 105.0},
				{SensorID: "cpu_2", SensorName: "CPU Core 2", Temperature: 98.2},
			},
			expectedTemp:   105.0,
			expectedSensor: "CPU Core 1",
		},
		{
			name: "ManySensors",
			sensors: []nexmonyx.TemperatureSensorData{
				{SensorID: "cpu_0", SensorName: "CPU Core 0", Temperature: 55.0},
				{SensorID: "cpu_1", SensorName: "CPU Core 1", Temperature: 58.0},
				{SensorID: "cpu_2", SensorName: "CPU Core 2", Temperature: 52.0},
				{SensorID: "cpu_3", SensorName: "CPU Core 3", Temperature: 60.0},
				{SensorID: "disk_0", SensorName: "Disk sda", Temperature: 42.0},
				{SensorID: "disk_1", SensorName: "Disk sdb", Temperature: 45.0},
				{SensorID: "system_0", SensorName: "Motherboard", Temperature: 48.0},
				{SensorID: "gpu_0", SensorName: "GPU", Temperature: 75.0},
			},
			expectedTemp:   75.0,
			expectedSensor: "GPU",
		},
		{
			name:            "EmptyMetrics",
			sensors:         []nexmonyx.TemperatureSensorData{},
			expectedTemp:    0.0,
			expectedSensor:  "",
			emptyCollection: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := nexmonyx.NewTemperatureMetrics()
			for _, sensor := range tt.sensors {
				metrics.AddTemperatureSensor(sensor)
			}

			maxTemp, maxSensorName := metrics.GetMaxTemperature()

			assert.Equal(t, tt.expectedTemp, maxTemp, "Temperature should match expected value")
			assert.Equal(t, tt.expectedSensor, maxSensorName, "Sensor name should match expected value")
		})
	}
}

// TestTemperatureMetrics_GetMaxTemperature_NilSensors tests edge case with nil sensors
func TestTemperatureMetrics_GetMaxTemperature_NilSensors(t *testing.T) {
	metrics := &nexmonyx.TemperatureMetrics{
		Sensors: nil,
	}

	maxTemp, maxSensorName := metrics.GetMaxTemperature()

	assert.Equal(t, 0.0, maxTemp)
	assert.Equal(t, "", maxSensorName)
}
