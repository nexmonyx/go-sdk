package nexmonyx

import (
	"testing"
)

// TestNewTemperatureMetrics tests the temperature metrics constructor
func TestNewTemperatureMetrics(t *testing.T) {
	metrics := NewTemperatureMetrics()

	if metrics == nil {
		t.Fatal("NewTemperatureMetrics() returned nil")
	}

	if metrics.Sensors == nil {
		t.Error("Sensors slice should be initialized")
	}

	if len(metrics.Sensors) != 0 {
		t.Errorf("Expected empty Sensors slice, got length %d", len(metrics.Sensors))
	}
}

// TestAddTemperatureSensor tests adding temperature sensors
func TestAddTemperatureSensor(t *testing.T) {
	tests := []struct {
		name          string
		initialState  *TemperatureMetrics
		sensors       []TemperatureSensorData
		expectedCount int
	}{
		{
			name:         "add to new metrics",
			initialState: NewTemperatureMetrics(),
			sensors: []TemperatureSensorData{
				{SensorID: "cpu_0", SensorName: "CPU Core 0", Temperature: 65.5},
			},
			expectedCount: 1,
		},
		{
			name:         "add multiple sensors",
			initialState: NewTemperatureMetrics(),
			sensors: []TemperatureSensorData{
				{SensorID: "cpu_0", SensorName: "CPU Core 0", Temperature: 65.5},
				{SensorID: "cpu_1", SensorName: "CPU Core 1", Temperature: 67.2},
				{SensorID: "disk_sda", SensorName: "Disk sda", Temperature: 45.0},
			},
			expectedCount: 3,
		},
		{
			name:          "add to nil sensors slice",
			initialState:  &TemperatureMetrics{Sensors: nil},
			sensors:       []TemperatureSensorData{{SensorID: "test", Temperature: 50.0}},
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, sensor := range tt.sensors {
				tt.initialState.AddTemperatureSensor(sensor)
			}

			if len(tt.initialState.Sensors) != tt.expectedCount {
				t.Errorf("Expected %d sensors, got %d", tt.expectedCount, len(tt.initialState.Sensors))
			}
		})
	}
}

// TestGetSensorByID tests sensor lookup by ID
func TestGetSensorByID(t *testing.T) {
	metrics := NewTemperatureMetrics()
	sensor1 := TemperatureSensorData{SensorID: "cpu_0", SensorName: "CPU Core 0", Temperature: 65.5}
	sensor2 := TemperatureSensorData{SensorID: "cpu_1", SensorName: "CPU Core 1", Temperature: 67.2}
	sensor3 := TemperatureSensorData{SensorID: "disk_sda", SensorName: "Disk sda", Temperature: 45.0}

	metrics.AddTemperatureSensor(sensor1)
	metrics.AddTemperatureSensor(sensor2)
	metrics.AddTemperatureSensor(sensor3)

	tests := []struct {
		name     string
		sensorID string
		wantNil  bool
		wantTemp float64
	}{
		{
			name:     "find existing sensor",
			sensorID: "cpu_0",
			wantNil:  false,
			wantTemp: 65.5,
		},
		{
			name:     "find another existing sensor",
			sensorID: "disk_sda",
			wantNil:  false,
			wantTemp: 45.0,
		},
		{
			name:     "non-existent sensor",
			sensorID: "nonexistent",
			wantNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := metrics.GetSensorByID(tt.sensorID)

			if tt.wantNil {
				if result != nil {
					t.Errorf("Expected nil, got sensor: %+v", result)
				}
			} else {
				if result == nil {
					t.Fatal("Expected sensor, got nil")
				}
				if result.Temperature != tt.wantTemp {
					t.Errorf("Expected temperature %f, got %f", tt.wantTemp, result.Temperature)
				}
			}
		})
	}
}

// TestGetSensorsByType tests filtering sensors by type
func TestGetSensorsByType(t *testing.T) {
	metrics := NewTemperatureMetrics()
	metrics.AddTemperatureSensor(TemperatureSensorData{SensorID: "cpu_0", Type: "cpu", Temperature: 65.5})
	metrics.AddTemperatureSensor(TemperatureSensorData{SensorID: "cpu_1", Type: "cpu", Temperature: 67.2})
	metrics.AddTemperatureSensor(TemperatureSensorData{SensorID: "disk_sda", Type: "disk", Temperature: 45.0})
	metrics.AddTemperatureSensor(TemperatureSensorData{SensorID: "system_1", Type: "system", Temperature: 55.0})

	tests := []struct {
		name          string
		sensorType    string
		expectedCount int
	}{
		{
			name:          "filter cpu sensors",
			sensorType:    "cpu",
			expectedCount: 2,
		},
		{
			name:          "filter disk sensors",
			sensorType:    "disk",
			expectedCount: 1,
		},
		{
			name:          "filter system sensors",
			sensorType:    "system",
			expectedCount: 1,
		},
		{
			name:          "non-existent type",
			sensorType:    "gpu",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := metrics.GetSensorsByType(tt.sensorType)

			if len(result) != tt.expectedCount {
				t.Errorf("Expected %d sensors of type %s, got %d", tt.expectedCount, tt.sensorType, len(result))
			}

			// Verify all returned sensors are of the correct type
			for _, sensor := range result {
				if sensor.Type != tt.sensorType {
					t.Errorf("Expected type %s, got %s", tt.sensorType, sensor.Type)
				}
			}
		})
	}
}

// TestGetMaxTemperature tests finding maximum temperature
func TestGetMaxTemperature(t *testing.T) {
	tests := []struct {
		name             string
		sensors          []TemperatureSensorData
		expectedTemp     float64
		expectedSensor   string
	}{
		{
			name:           "empty sensors",
			sensors:        []TemperatureSensorData{},
			expectedTemp:   0,
			expectedSensor: "",
		},
		{
			name: "single sensor",
			sensors: []TemperatureSensorData{
				{SensorName: "CPU Core 0", Temperature: 65.5},
			},
			expectedTemp:   65.5,
			expectedSensor: "CPU Core 0",
		},
		{
			name: "multiple sensors - max in middle",
			sensors: []TemperatureSensorData{
				{SensorName: "CPU Core 0", Temperature: 65.5},
				{SensorName: "CPU Core 1", Temperature: 85.2},
				{SensorName: "Disk sda", Temperature: 45.0},
			},
			expectedTemp:   85.2,
			expectedSensor: "CPU Core 1",
		},
		{
			name: "multiple sensors - max at end",
			sensors: []TemperatureSensorData{
				{SensorName: "CPU Core 0", Temperature: 65.5},
				{SensorName: "Disk sda", Temperature: 45.0},
				{SensorName: "CPU Core 1", Temperature: 90.0},
			},
			expectedTemp:   90.0,
			expectedSensor: "CPU Core 1",
		},
		{
			name: "identical temperatures",
			sensors: []TemperatureSensorData{
				{SensorName: "Sensor 1", Temperature: 50.0},
				{SensorName: "Sensor 2", Temperature: 50.0},
			},
			expectedTemp:   50.0,
			expectedSensor: "Sensor 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := NewTemperatureMetrics()
			for _, sensor := range tt.sensors {
				metrics.AddTemperatureSensor(sensor)
			}

			temp, sensorName := metrics.GetMaxTemperature()

			if temp != tt.expectedTemp {
				t.Errorf("Expected max temperature %f, got %f", tt.expectedTemp, temp)
			}
			if sensorName != tt.expectedSensor {
				t.Errorf("Expected sensor name %s, got %s", tt.expectedSensor, sensorName)
			}
		})
	}
}

// TestNewPowerMetrics tests the power metrics constructor
func TestNewPowerMetrics(t *testing.T) {
	metrics := NewPowerMetrics()

	if metrics == nil {
		t.Fatal("NewPowerMetrics() returned nil")
	}

	if metrics.PowerSupplies == nil {
		t.Error("PowerSupplies slice should be initialized")
	}

	if len(metrics.PowerSupplies) != 0 {
		t.Errorf("Expected empty PowerSupplies slice, got length %d", len(metrics.PowerSupplies))
	}

	if metrics.TotalPowerW != 0 {
		t.Errorf("Expected TotalPowerW to be 0, got %f", metrics.TotalPowerW)
	}
}

// TestAddPowerSupply tests adding power supplies
func TestAddPowerSupply(t *testing.T) {
	tests := []struct {
		name          string
		supplies      []PowerSupplyMetrics
		expectedCount int
		expectedTotal float64
	}{
		{
			name: "add single power supply",
			supplies: []PowerSupplyMetrics{
				{ID: "ps1", PowerWatts: 250.5},
			},
			expectedCount: 1,
			expectedTotal: 250.5,
		},
		{
			name: "add multiple power supplies",
			supplies: []PowerSupplyMetrics{
				{ID: "ps1", PowerWatts: 250.5},
				{ID: "ps2", PowerWatts: 300.0},
				{ID: "ps3", PowerWatts: 150.0},
			},
			expectedCount: 3,
			expectedTotal: 700.5,
		},
		{
			name: "add zero power supply",
			supplies: []PowerSupplyMetrics{
				{ID: "ps1", PowerWatts: 0},
			},
			expectedCount: 1,
			expectedTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := NewPowerMetrics()

			for _, ps := range tt.supplies {
				metrics.AddPowerSupply(ps)
			}

			if len(metrics.PowerSupplies) != tt.expectedCount {
				t.Errorf("Expected %d power supplies, got %d", tt.expectedCount, len(metrics.PowerSupplies))
			}

			if metrics.TotalPowerW != tt.expectedTotal {
				t.Errorf("Expected total power %f, got %f", tt.expectedTotal, metrics.TotalPowerW)
			}
		})
	}
}

// TestCalculateTotalPower tests power calculation
func TestCalculateTotalPower(t *testing.T) {
	tests := []struct {
		name          string
		supplies      []PowerSupplyMetrics
		expectedTotal float64
	}{
		{
			name:          "empty supplies",
			supplies:      []PowerSupplyMetrics{},
			expectedTotal: 0,
		},
		{
			name: "single supply",
			supplies: []PowerSupplyMetrics{
				{PowerWatts: 250.5},
			},
			expectedTotal: 250.5,
		},
		{
			name: "multiple supplies",
			supplies: []PowerSupplyMetrics{
				{PowerWatts: 250.5},
				{PowerWatts: 300.0},
				{PowerWatts: 150.75},
			},
			expectedTotal: 701.25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := &PowerMetrics{PowerSupplies: tt.supplies}
			total := metrics.CalculateTotalPower()

			if total != tt.expectedTotal {
				t.Errorf("Expected total power %f, got %f", tt.expectedTotal, total)
			}
		})
	}
}

// TestGetPowerSupplyByID tests power supply lookup
func TestGetPowerSupplyByID(t *testing.T) {
	metrics := NewPowerMetrics()
	ps1 := PowerSupplyMetrics{ID: "ps1", PowerWatts: 250.5, Status: "ok"}
	ps2 := PowerSupplyMetrics{ID: "ps2", PowerWatts: 300.0, Status: "ok"}
	ps3 := PowerSupplyMetrics{ID: "ps3", PowerWatts: 150.0, Status: "critical"}

	metrics.AddPowerSupply(ps1)
	metrics.AddPowerSupply(ps2)
	metrics.AddPowerSupply(ps3)

	tests := []struct {
		name       string
		id         string
		wantNil    bool
		wantPower  float64
		wantStatus string
	}{
		{
			name:       "find existing supply",
			id:         "ps1",
			wantNil:    false,
			wantPower:  250.5,
			wantStatus: "ok",
		},
		{
			name:       "find critical supply",
			id:         "ps3",
			wantNil:    false,
			wantPower:  150.0,
			wantStatus: "critical",
		},
		{
			name:    "non-existent supply",
			id:      "nonexistent",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := metrics.GetPowerSupplyByID(tt.id)

			if tt.wantNil {
				if result != nil {
					t.Errorf("Expected nil, got: %+v", result)
				}
			} else {
				if result == nil {
					t.Fatal("Expected power supply, got nil")
				}
				if result.PowerWatts != tt.wantPower {
					t.Errorf("Expected power %f, got %f", tt.wantPower, result.PowerWatts)
				}
				if result.Status != tt.wantStatus {
					t.Errorf("Expected status %s, got %s", tt.wantStatus, result.Status)
				}
			}
		})
	}
}

// TestGetFailedPowerSupplies tests filtering failed power supplies
func TestGetFailedPowerSupplies(t *testing.T) {
	tests := []struct {
		name          string
		supplies      []PowerSupplyMetrics
		expectedCount int
	}{
		{
			name:          "no supplies",
			supplies:      []PowerSupplyMetrics{},
			expectedCount: 0,
		},
		{
			name: "all ok supplies",
			supplies: []PowerSupplyMetrics{
				{ID: "ps1", Status: "ok"},
				{ID: "ps2", Status: "ok"},
			},
			expectedCount: 0,
		},
		{
			name: "mixed status supplies",
			supplies: []PowerSupplyMetrics{
				{ID: "ps1", Status: "ok"},
				{ID: "ps2", Status: "failed"},
				{ID: "ps3", Status: "critical"},
				{ID: "ps4", Status: "ok"},
			},
			expectedCount: 2,
		},
		{
			name: "all failed supplies",
			supplies: []PowerSupplyMetrics{
				{ID: "ps1", Status: "failed"},
				{ID: "ps2", Status: "critical"},
			},
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := &PowerMetrics{PowerSupplies: tt.supplies}
			failed := metrics.GetFailedPowerSupplies()

			if len(failed) != tt.expectedCount {
				t.Errorf("Expected %d failed supplies, got %d", tt.expectedCount, len(failed))
			}

			// Verify all returned supplies are failed or critical
			for _, ps := range failed {
				if ps.Status != "failed" && ps.Status != "critical" {
					t.Errorf("Expected failed/critical status, got %s", ps.Status)
				}
			}
		})
	}
}

// TestCalculateAverageEfficiency tests efficiency calculation
func TestCalculateAverageEfficiency(t *testing.T) {
	tests := []struct {
		name               string
		supplies           []PowerSupplyMetrics
		expectedEfficiency float64
	}{
		{
			name:               "empty supplies",
			supplies:           []PowerSupplyMetrics{},
			expectedEfficiency: 0,
		},
		{
			name: "single supply",
			supplies: []PowerSupplyMetrics{
				{Efficiency: 85.5},
			},
			expectedEfficiency: 85.5,
		},
		{
			name: "multiple supplies",
			supplies: []PowerSupplyMetrics{
				{Efficiency: 85.0},
				{Efficiency: 90.0},
				{Efficiency: 87.5},
			},
			expectedEfficiency: 87.5,
		},
		{
			name: "supplies with zero efficiency",
			supplies: []PowerSupplyMetrics{
				{Efficiency: 85.0},
				{Efficiency: 0},
				{Efficiency: 90.0},
			},
			expectedEfficiency: 87.5,
		},
		{
			name: "all zero efficiency",
			supplies: []PowerSupplyMetrics{
				{Efficiency: 0},
				{Efficiency: 0},
			},
			expectedEfficiency: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := &PowerMetrics{PowerSupplies: tt.supplies}
			avg := metrics.CalculateAverageEfficiency()

			if avg != tt.expectedEfficiency {
				t.Errorf("Expected efficiency %f, got %f", tt.expectedEfficiency, avg)
			}
		})
	}
}

// TestDetermineTemperatureStatus tests temperature status determination
func TestDetermineTemperatureStatus(t *testing.T) {
	tests := []struct {
		name           string
		temp           float64
		warning        float64
		critical       float64
		expectedStatus string
	}{
		{
			name:           "ok temperature",
			temp:           50.0,
			warning:        75.0,
			critical:       90.0,
			expectedStatus: "ok",
		},
		{
			name:           "warning temperature",
			temp:           80.0,
			warning:        75.0,
			critical:       90.0,
			expectedStatus: "warning",
		},
		{
			name:           "critical temperature",
			temp:           95.0,
			warning:        75.0,
			critical:       90.0,
			expectedStatus: "critical",
		},
		{
			name:           "at warning threshold",
			temp:           75.0,
			warning:        75.0,
			critical:       90.0,
			expectedStatus: "warning",
		},
		{
			name:           "at critical threshold",
			temp:           90.0,
			warning:        75.0,
			critical:       90.0,
			expectedStatus: "critical",
		},
		{
			name:           "zero thresholds",
			temp:           50.0,
			warning:        0,
			critical:       0,
			expectedStatus: "ok",
		},
		{
			name:           "negative thresholds",
			temp:           50.0,
			warning:        -10.0,
			critical:       -5.0,
			expectedStatus: "ok",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := DetermineTemperatureStatus(tt.temp, tt.warning, tt.critical)

			if status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, status)
			}
		})
	}
}

// TestCreateCPUTemperatureSensor tests CPU sensor creation
func TestCreateCPUTemperatureSensor(t *testing.T) {
	tests := []struct {
		name           string
		coreID         int
		temp           float64
		expectedID     string
		expectedName   string
		expectedStatus string
		expectedType   string
	}{
		{
			name:           "normal temperature",
			coreID:         0,
			temp:           65.0,
			expectedID:     "cpu_core_0",
			expectedName:   "CPU Core 0",
			expectedStatus: "ok",
			expectedType:   "cpu",
		},
		{
			name:           "warning temperature",
			coreID:         1,
			temp:           80.0,
			expectedID:     "cpu_core_1",
			expectedName:   "CPU Core 1",
			expectedStatus: "warning",
			expectedType:   "cpu",
		},
		{
			name:           "critical temperature",
			coreID:         2,
			temp:           95.0,
			expectedID:     "cpu_core_2",
			expectedName:   "CPU Core 2",
			expectedStatus: "critical",
			expectedType:   "cpu",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sensor := CreateCPUTemperatureSensor(tt.coreID, tt.temp)

			if sensor.SensorID != tt.expectedID {
				t.Errorf("Expected ID %s, got %s", tt.expectedID, sensor.SensorID)
			}
			if sensor.SensorName != tt.expectedName {
				t.Errorf("Expected name %s, got %s", tt.expectedName, sensor.SensorName)
			}
			if sensor.Temperature != tt.temp {
				t.Errorf("Expected temperature %f, got %f", tt.temp, sensor.Temperature)
			}
			if sensor.Status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, sensor.Status)
			}
			if sensor.Type != tt.expectedType {
				t.Errorf("Expected type %s, got %s", tt.expectedType, sensor.Type)
			}
			if sensor.UpperWarning != 75.0 {
				t.Errorf("Expected UpperWarning 75.0, got %f", sensor.UpperWarning)
			}
			if sensor.UpperCritical != 90.0 {
				t.Errorf("Expected UpperCritical 90.0, got %f", sensor.UpperCritical)
			}
		})
	}
}

// TestCreateDiskTemperatureSensor tests disk sensor creation
func TestCreateDiskTemperatureSensor(t *testing.T) {
	tests := []struct {
		name           string
		device         string
		temp           float64
		expectedID     string
		expectedName   string
		expectedStatus string
	}{
		{
			name:           "normal disk temperature",
			device:         "sda",
			temp:           40.0,
			expectedID:     "disk_sda",
			expectedName:   "Disk sda",
			expectedStatus: "ok",
		},
		{
			name:           "warning disk temperature",
			device:         "nvme0n1",
			temp:           55.0,
			expectedID:     "disk_nvme0n1",
			expectedName:   "Disk nvme0n1",
			expectedStatus: "warning",
		},
		{
			name:           "critical disk temperature",
			device:         "sdb",
			temp:           65.0,
			expectedID:     "disk_sdb",
			expectedName:   "Disk sdb",
			expectedStatus: "critical",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sensor := CreateDiskTemperatureSensor(tt.device, tt.temp)

			if sensor.SensorID != tt.expectedID {
				t.Errorf("Expected ID %s, got %s", tt.expectedID, sensor.SensorID)
			}
			if sensor.SensorName != tt.expectedName {
				t.Errorf("Expected name %s, got %s", tt.expectedName, sensor.SensorName)
			}
			if sensor.Temperature != tt.temp {
				t.Errorf("Expected temperature %f, got %f", tt.temp, sensor.Temperature)
			}
			if sensor.Status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, sensor.Status)
			}
			if sensor.Type != "disk" {
				t.Errorf("Expected type disk, got %s", sensor.Type)
			}
			if sensor.UpperWarning != 50.0 {
				t.Errorf("Expected UpperWarning 50.0, got %f", sensor.UpperWarning)
			}
			if sensor.UpperCritical != 60.0 {
				t.Errorf("Expected UpperCritical 60.0, got %f", sensor.UpperCritical)
			}
		})
	}
}

// TestCreateSystemTemperatureSensor tests system sensor creation
func TestCreateSystemTemperatureSensor(t *testing.T) {
	tests := []struct {
		name           string
		sensorName     string
		temp           float64
		expectedID     string
		expectedStatus string
	}{
		{
			name:           "normal system temperature",
			sensorName:     "Ambient",
			temp:           50.0,
			expectedID:     "system_Ambient",
			expectedStatus: "ok",
		},
		{
			name:           "warning system temperature",
			sensorName:     "Chassis",
			temp:           65.0,
			expectedID:     "system_Chassis",
			expectedStatus: "warning",
		},
		{
			name:           "critical system temperature",
			sensorName:     "Motherboard",
			temp:           80.0,
			expectedID:     "system_Motherboard",
			expectedStatus: "critical",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sensor := CreateSystemTemperatureSensor(tt.sensorName, tt.temp)

			if sensor.SensorID != tt.expectedID {
				t.Errorf("Expected ID %s, got %s", tt.expectedID, sensor.SensorID)
			}
			if sensor.SensorName != tt.sensorName {
				t.Errorf("Expected name %s, got %s", tt.sensorName, sensor.SensorName)
			}
			if sensor.Temperature != tt.temp {
				t.Errorf("Expected temperature %f, got %f", tt.temp, sensor.Temperature)
			}
			if sensor.Status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, sensor.Status)
			}
			if sensor.Type != "system" {
				t.Errorf("Expected type system, got %s", sensor.Type)
			}
			if sensor.UpperWarning != 60.0 {
				t.Errorf("Expected UpperWarning 60.0, got %f", sensor.UpperWarning)
			}
			if sensor.UpperCritical != 75.0 {
				t.Errorf("Expected UpperCritical 75.0, got %f", sensor.UpperCritical)
			}
		})
	}
}

// Benchmark tests for performance validation
func BenchmarkAddTemperatureSensor(b *testing.B) {
	metrics := NewTemperatureMetrics()
	sensor := TemperatureSensorData{SensorID: "bench", Temperature: 65.0}

	for i := 0; i < b.N; i++ {
		metrics.AddTemperatureSensor(sensor)
	}
}

func BenchmarkGetMaxTemperature(b *testing.B) {
	metrics := NewTemperatureMetrics()
	for i := 0; i < 100; i++ {
		metrics.AddTemperatureSensor(TemperatureSensorData{
			SensorID:    string(rune(i)),
			Temperature: float64(i),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = metrics.GetMaxTemperature()
	}
}

func BenchmarkCalculateTotalPower(b *testing.B) {
	metrics := NewPowerMetrics()
	for i := 0; i < 10; i++ {
		metrics.AddPowerSupply(PowerSupplyMetrics{
			ID:         string(rune(i)),
			PowerWatts: float64(i * 100),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = metrics.CalculateTotalPower()
	}
}

func BenchmarkCalculateAverageEfficiency(b *testing.B) {
	metrics := NewPowerMetrics()
	for i := 0; i < 10; i++ {
		metrics.AddPowerSupply(PowerSupplyMetrics{
			ID:         string(rune(i)),
			Efficiency: 80.0 + float64(i),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = metrics.CalculateAverageEfficiency()
	}
}
