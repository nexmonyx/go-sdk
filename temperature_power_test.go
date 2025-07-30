package nexmonyx_test

import (
	"testing"

	"github.com/nexmonyx/go-sdk/v2"
	"github.com/stretchr/testify/assert"
)

func TestTemperatureMetrics(t *testing.T) {
	t.Run("NewTemperatureMetrics", func(t *testing.T) {
		metrics := nexmonyx.NewTemperatureMetrics()
		assert.NotNil(t, metrics)
		assert.NotNil(t, metrics.Sensors)
		assert.Equal(t, 0, len(metrics.Sensors))
	})

	t.Run("AddTemperatureSensor", func(t *testing.T) {
		metrics := nexmonyx.NewTemperatureMetrics()
		
		sensor := nexmonyx.TemperatureSensorData{
			SensorID:    "test_sensor",
			SensorName:  "Test Sensor",
			Temperature: 45.5,
			Status:      "ok",
			Type:        "cpu",
		}
		
		metrics.AddTemperatureSensor(sensor)
		assert.Equal(t, 1, len(metrics.Sensors))
		assert.Equal(t, "test_sensor", metrics.Sensors[0].SensorID)
		assert.Equal(t, 45.5, metrics.Sensors[0].Temperature)
	})

	t.Run("GetMaxTemperature", func(t *testing.T) {
		metrics := nexmonyx.NewTemperatureMetrics()
		
		// Test with empty sensors
		maxTemp, maxSensor := metrics.GetMaxTemperature()
		assert.Equal(t, 0.0, maxTemp)
		assert.Equal(t, "", maxSensor)
		
		// Add sensors
		metrics.AddTemperatureSensor(nexmonyx.TemperatureSensorData{
			SensorName:  "CPU Core 0",
			Temperature: 45.0,
		})
		metrics.AddTemperatureSensor(nexmonyx.TemperatureSensorData{
			SensorName:  "GPU",
			Temperature: 68.0,
		})
		
		maxTemp, maxSensor = metrics.GetMaxTemperature()
		assert.Equal(t, 68.0, maxTemp)
		assert.Equal(t, "GPU", maxSensor)
	})
}

func TestPowerMetrics(t *testing.T) {
	t.Run("NewPowerMetrics", func(t *testing.T) {
		metrics := nexmonyx.NewPowerMetrics()
		assert.NotNil(t, metrics)
		assert.NotNil(t, metrics.PowerSupplies)
		assert.Equal(t, 0, len(metrics.PowerSupplies))
		assert.Equal(t, 0.0, metrics.TotalPowerW)
	})

	t.Run("AddPowerSupply", func(t *testing.T) {
		metrics := nexmonyx.NewPowerMetrics()
		
		ps := nexmonyx.PowerSupplyMetrics{
			ID:            "ps1",
			Name:          "Power Supply 1",
			Status:        "ok",
			PowerWatts:    250.0,
			MaxPowerWatts: 750.0,
		}
		
		metrics.AddPowerSupply(ps)
		assert.Equal(t, 1, len(metrics.PowerSupplies))
		assert.Equal(t, 250.0, metrics.TotalPowerW)
	})

	t.Run("CalculateAverageEfficiency", func(t *testing.T) {
		metrics := nexmonyx.NewPowerMetrics()
		
		// Test with no power supplies
		avg := metrics.CalculateAverageEfficiency()
		assert.Equal(t, 0.0, avg)
		
		// Test with power supplies having efficiency values
		metrics.PowerSupplies = []nexmonyx.PowerSupplyMetrics{
			{Efficiency: 94.5},
			{Efficiency: 92.0},
			{Efficiency: 93.5},
		}
		
		avg = metrics.CalculateAverageEfficiency()
		assert.InDelta(t, 93.333, avg, 0.01)
	})
}

func TestTemperatureHelpers(t *testing.T) {
	t.Run("DetermineTemperatureStatus", func(t *testing.T) {
		tests := []struct {
			name     string
			temp     float64
			warning  float64
			critical float64
			expected string
		}{
			{"Normal", 45.0, 75.0, 90.0, "ok"},
			{"Warning", 76.0, 75.0, 90.0, "warning"},
			{"Critical", 91.0, 75.0, 90.0, "critical"},
			{"No Thresholds", 100.0, 0, 0, "ok"},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				status := nexmonyx.DetermineTemperatureStatus(tt.temp, tt.warning, tt.critical)
				assert.Equal(t, tt.expected, status)
			})
		}
	})

	t.Run("CreateCPUTemperatureSensor", func(t *testing.T) {
		sensor := nexmonyx.CreateCPUTemperatureSensor(0, 65.0)
		
		assert.Equal(t, "cpu_core_0", sensor.SensorID)
		assert.Equal(t, "CPU Core 0", sensor.SensorName)
		assert.Equal(t, 65.0, sensor.Temperature)
		assert.Equal(t, "ok", sensor.Status)
		assert.Equal(t, "cpu", sensor.Type)
	})

	t.Run("CreateDiskTemperatureSensor", func(t *testing.T) {
		sensor := nexmonyx.CreateDiskTemperatureSensor("/dev/sda", 55.0)
		
		assert.Equal(t, "disk_/dev/sda", sensor.SensorID)
		assert.Equal(t, "Disk /dev/sda", sensor.SensorName)
		assert.Equal(t, 55.0, sensor.Temperature)
		assert.Equal(t, "warning", sensor.Status)
		assert.Equal(t, "disk", sensor.Type)
	})
}