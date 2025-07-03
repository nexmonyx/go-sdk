package nexmonyx

import (
	"context"
	"fmt"
	"time"
)

// Submit submits IPMI data for a server
func (s *IPMIService) Submit(ctx context.Context, request *IPMISubmitRequest) (*IPMISubmitResponse, error) {
	var resp map[string]IPMISubmitResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v2/ipmi/data",
		Body:   request,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if data, ok := resp["data"]; ok {
		return &data, nil
	}
	return nil, fmt.Errorf("unexpected response format")
}

// Get retrieves IPMI data for a server
func (s *IPMIService) Get(ctx context.Context, serverUUID string) (*IPMIData, error) {
	var resp StandardResponse
	resp.Data = &IPMIData{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/ipmi/%s", serverUUID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if data, ok := resp.Data.(*IPMIData); ok {
		return data, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetSensorData retrieves IPMI sensor data for a server
func (s *IPMIService) GetSensorData(ctx context.Context, serverUUID string) ([]*IPMISensor, error) {
	var resp StandardResponse
	var sensors []*IPMISensor
	resp.Data = &sensors

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/ipmi/%s/sensors", serverUUID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return sensors, nil
}

// ExecuteCommand executes an IPMI command on a server
func (s *IPMIService) ExecuteCommand(ctx context.Context, serverUUID string, command string, args []string) (*IPMICommandResult, error) {
	var resp StandardResponse
	resp.Data = &IPMICommandResult{}

	body := map[string]interface{}{
		"command": command,
		"args":    args,
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/ipmi/%s/execute", serverUUID),
		Body:   body,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if result, ok := resp.Data.(*IPMICommandResult); ok {
		return result, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetIPMI retrieves IPMI data for a server within a time range
func (s *IPMIService) GetIPMI(ctx context.Context, serverUUID string, timeRange *TimeRange) (*IPMIInfo, error) {
	var resp StandardResponse
	resp.Data = &IPMIInfo{}

	query := make(map[string]string)
	if timeRange != nil {
		query["start"] = timeRange.Start
		query["end"] = timeRange.End
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v2/ipmi/%s", serverUUID),
		Query:  query,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if data, ok := resp.Data.(*IPMIInfo); ok {
		return data, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetLatestIPMI retrieves the latest IPMI data for a server
func (s *IPMIService) GetLatestIPMI(ctx context.Context, serverUUID string) (*IPMIInfo, error) {
	var resp StandardResponse
	resp.Data = &IPMIInfo{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v2/ipmi/%s/latest", serverUUID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if data, ok := resp.Data.(*IPMIInfo); ok {
		return data, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// ListIPMIHistory retrieves IPMI history for a server
func (s *IPMIService) ListIPMIHistory(ctx context.Context, serverUUID string, opts *IPMIListOptions) ([]*IPMIRecord, *PaginationMeta, error) {
	var resp PaginatedResponse
	var records []*IPMIRecord
	resp.Data = &records

	req := &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v2/ipmi/%s/history", serverUUID),
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return records, resp.Meta, nil
}

// IPMISubmitRequest represents a request to submit IPMI data
type IPMISubmitRequest struct {
	ServerUUID  string       `json:"server_uuid"`
	CollectedAt time.Time    `json:"collected_at"`
	IPMI        IPMIInfo     `json:"ipmi"`
	IPMIData    *IPMIData    `json:"ipmi_data,omitempty"`
	Sensors     []IPMISensor `json:"sensors,omitempty"`
	Events      []IPMIEvent  `json:"events,omitempty"`
}

// IPMISubmitResponse represents the response from submitting IPMI data
type IPMISubmitResponse struct {
	ServerUUID       string    `json:"server_uuid"`
	Timestamp        time.Time `json:"timestamp"`
	DataSaved        bool      `json:"data_saved"`
	SensorCount      int       `json:"sensor_count"`
	EventCount       int       `json:"event_count"`
	CollectionMethod string    `json:"collection_method,omitempty"`
	IPMIVersion      string    `json:"ipmi_version,omitempty"`
}

// IPMIData represents IPMI data for a server
type IPMIData struct {
	BMCInfo       *BMCInfo               `json:"bmc_info,omitempty"`
	ChassisStatus *ChassisStatus         `json:"chassis_status,omitempty"`
	PowerStatus   *PowerStatus           `json:"power_status,omitempty"`
	SystemHealth  string                 `json:"system_health"`
	FanStatus     []FanStatus            `json:"fan_status,omitempty"`
	Temperatures  []TemperatureSensor    `json:"temperatures,omitempty"`
	PowerSupplies []PowerSupplyStatus    `json:"power_supplies,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// BMCInfo represents BMC (Baseboard Management Controller) information
type BMCInfo struct {
	Version          string `json:"version,omitempty"`
	Manufacturer     string `json:"manufacturer,omitempty"`
	Firmware         string `json:"firmware,omitempty"`
	IPAddress        string `json:"ip_address,omitempty"`
	MACAddress       string `json:"mac_address,omitempty"`
	DeviceID         string `json:"device_id,omitempty"`
	FirmwareRevision string `json:"firmware_revision,omitempty"`
	ManufacturerName string `json:"manufacturer_name,omitempty"`
	ProductName      string `json:"product_name,omitempty"`
}

// ChassisStatus represents chassis status
type ChassisStatus struct {
	PowerState        string `json:"power_state"`
	ChassisIntrusion  bool   `json:"chassis_intrusion"`
	FrontPanelLockout bool   `json:"front_panel_lockout"`
	DriveFault        bool   `json:"drive_fault"`
	CoolingFault      bool   `json:"cooling_fault"`
}

// PowerStatus represents power status
type PowerStatus struct {
	PowerOn          bool    `json:"power_on"`
	PowerConsumption float64 `json:"power_consumption"` // Watts
	PowerCapacity    float64 `json:"power_capacity"`    // Watts
}

// FanStatus represents fan status
type FanStatus struct {
	Name    string  `json:"name"`
	RPM     int     `json:"rpm"`
	Status  string  `json:"status"`
	Percent float64 `json:"percent"`
}

// TemperatureSensor represents a temperature sensor
type TemperatureSensor struct {
	Name      string  `json:"name"`
	Reading   float64 `json:"reading"` // Celsius
	Status    string  `json:"status"`
	Threshold float64 `json:"threshold"` // Celsius
	Critical  float64 `json:"critical"`  // Celsius
}

// PowerSupplyStatus represents power supply status
type PowerSupplyStatus struct {
	Name        string  `json:"name"`
	Status      string  `json:"status"`
	Present     bool    `json:"present"`
	PowerOutput float64 `json:"power_output"` // Watts
	Voltage     float64 `json:"voltage"`      // Volts
	Current     float64 `json:"current"`      // Amps
}

// IPMISensor represents an IPMI sensor
type IPMISensor struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Reading     float64 `json:"reading"`
	Unit        string  `json:"unit"`
	Status      string  `json:"status"`
	LowerBound  float64 `json:"lower_bound,omitempty"`
	UpperBound  float64 `json:"upper_bound,omitempty"`
	Description string  `json:"description,omitempty"`
}

// IPMIEvent represents an IPMI event
type IPMIEvent struct {
	ID          string    `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	SensorName  string    `json:"sensor_name"`
	SensorType  string    `json:"sensor_type"`
	EventType   string    `json:"event_type"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	RawData     string    `json:"raw_data,omitempty"`
}

// IPMICommandResult represents the result of an IPMI command
type IPMICommandResult struct {
	Command    string    `json:"command"`
	Output     string    `json:"output"`
	Error      string    `json:"error,omitempty"`
	ExitCode   int       `json:"exit_code"`
	ExecutedAt time.Time `json:"executed_at"`
}

// IPMIInfo represents IPMI information
type IPMIInfo struct {
	CollectionMethod string                `json:"collection_method"`
	IPMIVersion      string                `json:"ipmi_version"`
	BMC              *BMCInfo              `json:"bmc,omitempty"`
	Sensors          []IPMISensorInfo      `json:"sensors,omitempty"`
	PowerInfo        *IPMIPowerInfo        `json:"power_info,omitempty"`
	Fans             []IPMIFanInfo         `json:"fans,omitempty"`
	Temperatures     []IPMITemperatureInfo `json:"temperatures,omitempty"`
	SystemHealth     *IPMISystemHealth     `json:"system_health,omitempty"`
}

// IPMISensorInfo represents IPMI sensor information
type IPMISensorInfo struct {
	SensorID   string  `json:"sensor_id"`
	SensorName string  `json:"sensor_name"`
	SensorType string  `json:"sensor_type"`
	Value      float64 `json:"value"`
	Unit       string  `json:"unit"`
	Status     string  `json:"status"`
}

// IPMIPowerInfo represents IPMI power information
type IPMIPowerInfo struct {
	PowerConsumption float64 `json:"power_consumption"`
	PowerCapacity    float64 `json:"power_capacity"`
	PowerState       string  `json:"power_state"`
}

// IPMIFanInfo represents IPMI fan information
type IPMIFanInfo struct {
	FanID        string  `json:"fan_id"`
	FanName      string  `json:"fan_name"`
	Speed        int     `json:"speed"`
	SpeedPercent float64 `json:"speed_percent"`
	Status       string  `json:"status"`
}

// IPMITemperatureInfo represents IPMI temperature information
type IPMITemperatureInfo struct {
	SensorID     string   `json:"sensor_id"`
	SensorName   string   `json:"sensor_name"`
	Temperature  float64  `json:"temperature"`
	Status       string   `json:"status"`
	UpperWarning *float64 `json:"upper_warning,omitempty"`
}

// IPMISystemHealth represents IPMI system health
type IPMISystemHealth struct {
	OverallStatus  string `json:"overall_status"`
	PowerStatus    string `json:"power_status"`
	ThermalStatus  string `json:"thermal_status"`
	FanStatus      string `json:"fan_status"`
	VoltageStatus  string `json:"voltage_status"`
	CriticalEvents int    `json:"critical_events"`
	WarningEvents  int    `json:"warning_events"`
	HealthScore    int    `json:"health_score"`
}

// IPMIRecord represents an IPMI record with metadata
type IPMIRecord struct {
	ID               uint      `json:"id"`
	ServerUUID       string    `json:"server_uuid"`
	OrganizationID   uint      `json:"organization_id"`
	CollectedAt      time.Time `json:"collected_at"`
	CollectionMethod string    `json:"collection_method,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	IPMI             IPMIInfo  `json:"ipmi"`
}

// IPMIListOptions represents options for listing IPMI data
type IPMIListOptions struct {
	ListOptions
	StartTime *time.Time `url:"start_time,omitempty"`
	EndTime   *time.Time `url:"end_time,omitempty"`
}
