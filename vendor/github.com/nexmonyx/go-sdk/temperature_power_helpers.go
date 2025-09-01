package nexmonyx

import "fmt"

// NewTemperatureMetrics creates a new TemperatureMetrics instance
func NewTemperatureMetrics() *TemperatureMetrics {
	return &TemperatureMetrics{
		Sensors: make([]TemperatureSensorData, 0),
	}
}

// AddTemperatureSensor adds a temperature sensor reading
func (t *TemperatureMetrics) AddTemperatureSensor(sensor TemperatureSensorData) {
	if t.Sensors == nil {
		t.Sensors = make([]TemperatureSensorData, 0)
	}
	t.Sensors = append(t.Sensors, sensor)
}

// GetSensorByID returns a temperature sensor by its ID
func (t *TemperatureMetrics) GetSensorByID(sensorID string) *TemperatureSensorData {
	for i := range t.Sensors {
		if t.Sensors[i].SensorID == sensorID {
			return &t.Sensors[i]
		}
	}
	return nil
}

// GetSensorsByType returns all temperature sensors of a specific type
func (t *TemperatureMetrics) GetSensorsByType(sensorType string) []TemperatureSensorData {
	var sensors []TemperatureSensorData
	for _, sensor := range t.Sensors {
		if sensor.Type == sensorType {
			sensors = append(sensors, sensor)
		}
	}
	return sensors
}

// GetMaxTemperature returns the maximum temperature across all sensors
func (t *TemperatureMetrics) GetMaxTemperature() (float64, string) {
	if len(t.Sensors) == 0 {
		return 0, ""
	}
	
	maxTemp := t.Sensors[0].Temperature
	maxSensorName := t.Sensors[0].SensorName
	
	for _, sensor := range t.Sensors[1:] {
		if sensor.Temperature > maxTemp {
			maxTemp = sensor.Temperature
			maxSensorName = sensor.SensorName
		}
	}
	
	return maxTemp, maxSensorName
}

// NewPowerMetrics creates a new PowerMetrics instance
func NewPowerMetrics() *PowerMetrics {
	return &PowerMetrics{
		PowerSupplies: make([]PowerSupplyMetrics, 0),
	}
}

// AddPowerSupply adds power supply metrics
func (p *PowerMetrics) AddPowerSupply(ps PowerSupplyMetrics) {
	if p.PowerSupplies == nil {
		p.PowerSupplies = make([]PowerSupplyMetrics, 0)
	}
	p.PowerSupplies = append(p.PowerSupplies, ps)
	
	// Update total power
	p.TotalPowerW = p.CalculateTotalPower()
}

// CalculateTotalPower calculates the total power consumption across all power supplies
func (p *PowerMetrics) CalculateTotalPower() float64 {
	total := 0.0
	for _, ps := range p.PowerSupplies {
		total += ps.PowerWatts
	}
	return total
}

// GetPowerSupplyByID returns a power supply by its ID
func (p *PowerMetrics) GetPowerSupplyByID(id string) *PowerSupplyMetrics {
	for i := range p.PowerSupplies {
		if p.PowerSupplies[i].ID == id {
			return &p.PowerSupplies[i]
		}
	}
	return nil
}

// GetFailedPowerSupplies returns all power supplies with failed status
func (p *PowerMetrics) GetFailedPowerSupplies() []PowerSupplyMetrics {
	var failed []PowerSupplyMetrics
	for _, ps := range p.PowerSupplies {
		if ps.Status == "failed" || ps.Status == "critical" {
			failed = append(failed, ps)
		}
	}
	return failed
}

// CalculateAverageEfficiency calculates the average efficiency across all power supplies
func (p *PowerMetrics) CalculateAverageEfficiency() float64 {
	if len(p.PowerSupplies) == 0 {
		return 0
	}
	
	totalEfficiency := 0.0
	count := 0
	
	for _, ps := range p.PowerSupplies {
		if ps.Efficiency > 0 {
			totalEfficiency += ps.Efficiency
			count++
		}
	}
	
	if count == 0 {
		return 0
	}
	
	return totalEfficiency / float64(count)
}

// DetermineTemperatureStatus determines the status based on temperature and thresholds
func DetermineTemperatureStatus(temp, warning, critical float64) string {
	if temp >= critical && critical > 0 {
		return "critical"
	} else if temp >= warning && warning > 0 {
		return "warning"
	}
	return "ok"
}

// CreateCPUTemperatureSensor creates a temperature sensor for CPU
func CreateCPUTemperatureSensor(coreID int, temp float64) TemperatureSensorData {
	sensorID := fmt.Sprintf("cpu_core_%d", coreID)
	return TemperatureSensorData{
		SensorID:      sensorID,
		SensorName:    fmt.Sprintf("CPU Core %d", coreID),
		Temperature:   temp,
		Status:        DetermineTemperatureStatus(temp, 75.0, 90.0),
		Type:          "cpu",
		Location:      "processor",
		UpperWarning:  75.0,
		UpperCritical: 90.0,
	}
}

// CreateDiskTemperatureSensor creates a temperature sensor for disk
func CreateDiskTemperatureSensor(device string, temp float64) TemperatureSensorData {
	return TemperatureSensorData{
		SensorID:      "disk_" + device,
		SensorName:    "Disk " + device,
		Temperature:   temp,
		Status:        DetermineTemperatureStatus(temp, 50.0, 60.0),
		Type:          "disk",
		Location:      "storage",
		UpperWarning:  50.0,
		UpperCritical: 60.0,
	}
}

// CreateSystemTemperatureSensor creates a generic system temperature sensor
func CreateSystemTemperatureSensor(name string, temp float64) TemperatureSensorData {
	return TemperatureSensorData{
		SensorID:      "system_" + name,
		SensorName:    name,
		Temperature:   temp,
		Status:        DetermineTemperatureStatus(temp, 60.0, 75.0),
		Type:          "system",
		UpperWarning:  60.0,
		UpperCritical: 75.0,
	}
}