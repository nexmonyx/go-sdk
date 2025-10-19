package nexmonyx

import (
	"encoding/json"
	"fmt"
	"time"
)

// CustomTime handles custom time parsing for API responses
type CustomTime struct {
	time.Time
}

// UnmarshalJSON implements json.Unmarshaler
func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	s := string(b)
	s = s[1 : len(s)-1] // Remove quotes

	if s == "null" || s == "" {
		return nil
	}

	// Try multiple formats
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
	}

	var err error
	for _, format := range formats {
		ct.Time, err = time.Parse(format, s)
		if err == nil {
			return nil
		}
	}

	return err
}

// MarshalJSON implements json.Marshaler
func (ct CustomTime) MarshalJSON() ([]byte, error) {
	if ct.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(ct.Time.Format(time.RFC3339))
}

// GormModel is the base model for all entities
type GormModel struct {
	ID        uint        `json:"id"`
	CreatedAt *CustomTime `json:"created_at,omitempty"`
	UpdatedAt *CustomTime `json:"updated_at,omitempty"`
	DeletedAt *CustomTime `json:"deleted_at,omitempty"`
}

// BaseModel is the base model for entities with UUID
type BaseModel struct {
	UUID      string      `json:"uuid"`
	CreatedAt *CustomTime `json:"created_at,omitempty"`
	UpdatedAt *CustomTime `json:"updated_at,omitempty"`
}

// ResponseMeta represents metadata in API responses (alias for PaginationMeta)
type ResponseMeta = PaginationMeta

// PaginationOptions represents common pagination parameters
type PaginationOptions struct {
	Page  int `json:"page,omitempty"`
	Limit int `json:"limit,omitempty"`
}

// Organization represents an organization
type Organization struct {
	GormModel
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Industry    string `json:"industry,omitempty"`
	Website     string `json:"website,omitempty"`
	Size        string `json:"size,omitempty"`
	Country     string `json:"country,omitempty"`
	TimeZone    string `json:"timezone,omitempty"`

	// Billing and subscription
	StripeCustomerID   string      `json:"stripe_customer_id,omitempty"`
	SubscriptionID     string      `json:"subscription_id,omitempty"`
	SubscriptionStatus string      `json:"subscription_status,omitempty"`
	SubscriptionPlan   string      `json:"subscription_plan,omitempty"`
	TrialEndsAt        *CustomTime `json:"trial_ends_at,omitempty"`

	// Features and limits
	MaxServers        int  `json:"max_servers"`
	MaxUsers          int  `json:"max_users"`
	MaxProbes         int  `json:"max_probes"`
	DataRetentionDays int  `json:"data_retention_days"`
	AlertsEnabled     bool `json:"alerts_enabled"`
	MonitoringEnabled bool `json:"monitoring_enabled"`

	// Relationships
	Users            []User   `json:"users,omitempty"`
	Servers          []Server `json:"servers,omitempty"`
	BillingContact   *User    `json:"billing_contact,omitempty"`
	TechnicalContact *User    `json:"technical_contact,omitempty"`

	// Settings and preferences
	Settings map[string]interface{} `json:"settings,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Tags     []string               `json:"tags,omitempty"`
}

// User represents a user
type User struct {
	GormModel
	Email          string `json:"email"`
	FirstName      string `json:"first_name,omitempty"`
	LastName       string `json:"last_name,omitempty"`
	DisplayName    string `json:"display_name,omitempty"`
	ProfilePicture string `json:"profile_picture,omitempty"`
	PhoneNumber    string `json:"phone_number,omitempty"`

	// Authentication
	Auth0ID          string      `json:"auth0_id,omitempty"`
	LastLogin        *CustomTime `json:"last_login,omitempty"`
	EmailVerified    bool        `json:"email_verified"`
	TwoFactorEnabled bool        `json:"two_factor_enabled"`

	// Organization and permissions
	OrganizationID uint          `json:"organization_id,omitempty"`
	Organization   *Organization `json:"organization,omitempty"`
	Role           string        `json:"role,omitempty"`
	Permissions    []string      `json:"permissions,omitempty"`
	IsActive       bool          `json:"is_active"`
	IsAdmin        bool          `json:"is_admin"`

	// Preferences
	Timezone          string                 `json:"timezone,omitempty"`
	Language          string                 `json:"language,omitempty"`
	NotificationPrefs map[string]bool        `json:"notification_prefs,omitempty"`
	UIPreferences     map[string]interface{} `json:"ui_preferences,omitempty"`
}

// Server represents a monitored server
type Server struct {
	GormModel
	ServerUUID     string        `json:"server_uuid"`
	ServerSecret   string        `json:"server_secret,omitempty"`
	Hostname       string        `json:"hostname"`
	FQDN           string        `json:"fqdn,omitempty"`
	OrganizationID uint          `json:"organization_id"`
	Organization   *Organization `json:"organization,omitempty"`

	// System information
	OS              string  `json:"os,omitempty"`
	OSVersion       string  `json:"os_version,omitempty"`
	OSArch          string  `json:"os_arch,omitempty"`
	KernelVersion   string  `json:"kernel_version,omitempty"`
	CPUArchitecture string  `json:"cpu_architecture,omitempty"`
	CPUModel        string  `json:"cpu_model,omitempty"`
	CPUCores        int     `json:"cpu_cores"`
	TotalMemoryGB   float64 `json:"total_memory_gb"`
	TotalDiskGB     float64 `json:"total_disk_gb"`

	// Network information
	MainIP            string   `json:"main_ip,omitempty"`
	IPv6Address       string   `json:"ipv6_address,omitempty"`
	NetworkInterfaces []string `json:"network_interfaces,omitempty"`

	// Location and classification
	Environment    string `json:"environment,omitempty"`
	Location       string `json:"location,omitempty"`
	DataCenter     string `json:"data_center,omitempty"`
	Rack           string `json:"rack,omitempty"`
	Classification string `json:"classification,omitempty"`

	// Monitoring and status
	LastHeartbeat     *CustomTime `json:"last_heartbeat,omitempty"`
	Status            string      `json:"status,omitempty"`
	AgentVersion      string      `json:"agent_version,omitempty"`
	MonitoringEnabled bool        `json:"monitoring_enabled"`
	AlertsEnabled     bool        `json:"alerts_enabled"`

	// Cloud/provider information
	Provider         string                 `json:"provider,omitempty"`
	ProviderID       string                 `json:"provider_id,omitempty"`
	InstanceType     string                 `json:"instance_type,omitempty"`
	Region           string                 `json:"region,omitempty"`
	AvailabilityZone string                 `json:"availability_zone,omitempty"`
	ProviderMetadata map[string]interface{} `json:"provider_metadata,omitempty"`

	// Tags and metadata
	Tags         []string               `json:"tags,omitempty"`
	Labels       map[string]string      `json:"labels,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
}

// ServerCreateRequest represents a request to create/register a new server
type ServerCreateRequest struct {
	Hostname       string `json:"hostname"`
	MainIP         string `json:"main_ip"`
	OS             string `json:"os"`
	OSVersion      string `json:"os_version"`
	OSArch         string `json:"os_arch"`
	SerialNumber   string `json:"serial_number"`
	MacAddress     string `json:"mac_address"`
	Environment    string `json:"environment,omitempty"`
	Location       string `json:"location,omitempty"`
	Classification string `json:"classification,omitempty"`
	HardwareIP     string `json:"hardware_ip,omitempty"`
	HardwareType   string `json:"hardware_type,omitempty"`
}

// ServerRegistrationResponse represents the response from server registration
type ServerRegistrationResponse struct {
	Server       *Server `json:"server"`
	ServerUUID   string  `json:"server_uuid"`
	ServerSecret string  `json:"server_secret"`
}

// ServerUpdateRequest represents a request to update server information
type ServerUpdateRequest struct {
	Hostname       string `json:"hostname,omitempty"`
	MainIP         string `json:"main_ip,omitempty"`
	Environment    string `json:"environment,omitempty"`
	Location       string `json:"location,omitempty"`
	Classification string `json:"classification,omitempty"`
}

// HardwareDetails represents detailed hardware information for server updates
type HardwareDetails struct {
	CPU     []ServerCPUInfo              `json:"cpu,omitempty"`
	Memory  *ServerMemoryInfo            `json:"memory,omitempty"`
	Network []ServerNetworkInterfaceInfo `json:"network,omitempty"`
	Disks   []ServerDiskInfo             `json:"disks,omitempty"`
}

// ServerCPUInfo represents CPU hardware information for server update requests
type ServerCPUInfo struct {
	PhysicalID       string  `json:"physical_id,omitempty"`
	Manufacturer     string  `json:"manufacturer,omitempty"`
	ModelName        string  `json:"model_name,omitempty"`
	Family           string  `json:"family,omitempty"`
	Model            string  `json:"model,omitempty"`
	Stepping         string  `json:"stepping,omitempty"`
	Microcode        string  `json:"microcode,omitempty"`
	Architecture     string  `json:"architecture,omitempty"`
	SocketType       string  `json:"socket_type,omitempty"`
	BaseSpeed        float64 `json:"base_speed,omitempty"`
	MaxSpeed         float64 `json:"max_speed,omitempty"`
	CurrentSpeed     float64 `json:"current_speed,omitempty"`
	BusSpeed         float64 `json:"bus_speed,omitempty"`
	SocketCount      int     `json:"socket_count,omitempty"`
	PhysicalCores    int     `json:"physical_cores,omitempty"`
	LogicalCores     int     `json:"logical_cores,omitempty"`
	L1Cache          int     `json:"l1_cache,omitempty"`
	L2Cache          int     `json:"l2_cache,omitempty"`
	L3Cache          int     `json:"l3_cache,omitempty"`
	Flags            string  `json:"flags,omitempty"`
	Virtualization   string  `json:"virtualization,omitempty"`
	PowerFeatures    string  `json:"power_features,omitempty"`
	Usage            float64 `json:"usage,omitempty"`
	Temperature      float64 `json:"temperature,omitempty"`
	PowerConsumption float64 `json:"power_consumption,omitempty"`
}

// ServerMemoryInfo represents memory hardware information for server update requests
type ServerMemoryInfo struct {
	TotalSize     uint64 `json:"total_size,omitempty"`
	AvailableSize uint64 `json:"available_size,omitempty"`
	UsedSize      uint64 `json:"used_size,omitempty"`
	MemoryType    string `json:"memory_type,omitempty"`
	Speed         int    `json:"speed,omitempty"`
	ModuleCount   int    `json:"module_count,omitempty"`
	ECCSupported  bool   `json:"ecc_supported,omitempty"`
}

// ServerNetworkInterfaceInfo represents network interface hardware information for server update requests
type ServerNetworkInterfaceInfo struct {
	Name          string `json:"name,omitempty"`
	HardwareAddr  string `json:"hardware_addr,omitempty"`
	MTU           int    `json:"mtu,omitempty"`
	Flags         string `json:"flags,omitempty"`
	Addrs         string `json:"addrs,omitempty"`
	BytesReceived uint64 `json:"bytes_received,omitempty"`
	BytesSent     uint64 `json:"bytes_sent,omitempty"`
	SpeedMbps     int    `json:"speed_mbps,omitempty"`
	IsUp          bool   `json:"is_up,omitempty"`
	IsWireless    bool   `json:"is_wireless,omitempty"`
}

// ServerDiskInfo represents disk hardware information for server update requests
type ServerDiskInfo struct {
	Device       string `json:"device,omitempty"`
	DiskModel    string `json:"disk_model,omitempty"`
	SerialNumber string `json:"serial_number,omitempty"`
	Size         int64  `json:"size,omitempty"`
	Type         string `json:"type,omitempty"`
	Vendor       string `json:"vendor,omitempty"`
}

// ServerDetailsUpdateRequest represents a request to update detailed server information
type ServerDetailsUpdateRequest struct {
	// Basic server information
	Hostname       string `json:"hostname,omitempty"`
	MainIP         string `json:"main_ip,omitempty"`
	Environment    string `json:"environment,omitempty"`
	Location       string `json:"location,omitempty"`
	Classification string `json:"classification,omitempty"`
	// System information
	OS           string `json:"os,omitempty"`
	OSVersion    string `json:"os_version,omitempty"`
	OSArch       string `json:"os_arch,omitempty"`
	SerialNumber string `json:"serial_number,omitempty"`
	MacAddress   string `json:"mac_address,omitempty"`
	// Hardware details (legacy fields for backward compatibility)
	CPUModel     string `json:"cpu_model,omitempty"`
	CPUCount     int    `json:"cpu_count,omitempty"`
	CPUCores     int    `json:"cpu_cores,omitempty"`
	MemoryTotal  uint64 `json:"memory_total,omitempty"`
	StorageTotal uint64 `json:"storage_total,omitempty"`
	// Enhanced hardware details (optional)
	Hardware *HardwareDetails `json:"hardware,omitempty"`
}

// Alert represents an alert
type Alert struct {
	GormModel
	Name           string  `json:"name"`
	Description    string  `json:"description,omitempty"`
	OrganizationID uint    `json:"organization_id"`
	ServerID       *uint   `json:"server_id,omitempty"`
	Server         *Server `json:"server,omitempty"`

	// Alert configuration
	Type       string  `json:"type"`
	MetricName string  `json:"metric_name"`
	Condition  string  `json:"condition"`
	Threshold  float64 `json:"threshold"`
	Duration   int     `json:"duration"`
	Frequency  int     `json:"frequency"`

	// Alert state
	Enabled       bool        `json:"enabled"`
	Status        string      `json:"status"`
	LastTriggered *CustomTime `json:"last_triggered,omitempty"`
	LastResolved  *CustomTime `json:"last_resolved,omitempty"`
	TriggerCount  int         `json:"trigger_count"`

	// Notification settings
	Severity        string   `json:"severity"`
	Channels        []string `json:"channels"`
	Recipients      []string `json:"recipients"`
	NotifyOnResolve bool     `json:"notify_on_resolve"`

	// Actions and metadata
	Actions  []AlertAction          `json:"actions,omitempty"`
	Tags     []string               `json:"tags,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// AlertAction represents an action to take when an alert triggers
type AlertAction struct {
	Type      string                 `json:"type"`
	Config    map[string]interface{} `json:"config"`
	OnTrigger bool                   `json:"on_trigger"`
	OnResolve bool                   `json:"on_resolve"`
}

// AlertChannel represents a notification channel configuration
type AlertChannel struct {
	ID             uint                   `json:"id"`
	Name           string                 `json:"name"`
	Type           string                 `json:"type"` // email, slack, pagerduty, webhook
	Configuration  map[string]interface{} `json:"configuration"`
	Enabled        bool                   `json:"enabled"`
	OrganizationID uint                   `json:"organization_id"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// AlertRule represents a more detailed alert rule configuration
type AlertRule struct {
	ID             uint   `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	OrganizationID uint   `json:"organization_id"`

	// Scope configuration
	ScopeType  string `json:"scope_type"` // organization, server, tag, group
	ScopeID    *uint  `json:"scope_id,omitempty"`
	ScopeValue string `json:"scope_value,omitempty"`

	// Metric configuration
	MetricName  string `json:"metric_name"`
	Aggregation string `json:"aggregation"` // avg, sum, min, max, count

	// Conditions
	Conditions AlertConditions `json:"conditions"`

	// Notification settings
	ChannelIDs []uint `json:"channel_ids"`

	// State
	Enabled       bool       `json:"enabled"`
	LastEvaluated *time.Time `json:"last_evaluated,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AlertConditions represents the conditions for triggering an alert
type AlertConditions struct {
	TimeWindow int              `json:"time_window"` // in minutes
	Thresholds []AlertThreshold `json:"thresholds"`
}

// AlertThreshold represents a threshold configuration
type AlertThreshold struct {
	Value    float64 `json:"value"`
	Operator string  `json:"operator"` // >, >=, <, <=, ==, !=
	Duration int     `json:"duration"` // in minutes (how long condition must be true)
	Severity string  `json:"severity"` // critical, warning, info
}

// Metric represents a metric data point
type Metric struct {
	ServerID   uint                   `json:"server_id"`
	ServerUUID string                 `json:"server_uuid"`
	Timestamp  time.Time              `json:"timestamp"`
	Name       string                 `json:"name"`
	Value      float64                `json:"value"`
	Unit       string                 `json:"unit,omitempty"`
	Tags       map[string]string      `json:"tags,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// MonitoringAgent represents a monitoring agent
type MonitoringAgent struct {
	GormModel
	UUID           string                 `json:"uuid"`
	Name           string                 `json:"name"`
	Status         string                 `json:"status"`
	Version        string                 `json:"version"`
	OrganizationID uint                   `json:"organization_id"`
	ServerUUID     string                 `json:"server_uuid,omitempty"`
	Configuration  map[string]interface{} `json:"configuration,omitempty"`
	LastHeartbeat  *CustomTime            `json:"last_heartbeat,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// ProbeTestResult represents the result of a probe test
type ProbeTestResult struct {
	GormModel
	ProbeID    uint        `json:"probe_id"`
	ProbeUUID  string      `json:"probe_uuid,omitempty"`
	AgentID    uint        `json:"agent_id,omitempty"`
	ExecutedAt *CustomTime `json:"executed_at"`
	Target     string      `json:"target,omitempty"`
	Type       string      `json:"type,omitempty"`

	// Result data
	Status       string `json:"status"`
	ResponseTime int    `json:"response_time"`
	StatusCode   int    `json:"status_code,omitempty"`
	ResponseBody string `json:"response_body,omitempty"`
	Error        string `json:"error,omitempty"`

	// Additional metrics
	DNSTime       int `json:"dns_time,omitempty"`
	ConnectTime   int `json:"connect_time,omitempty"`
	TLSTime       int `json:"tls_time,omitempty"`
	FirstByteTime int `json:"first_byte_time,omitempty"`
	TotalTime     int `json:"total_time,omitempty"`

	// Metadata
	Region   string                 `json:"region,omitempty"`
	Location string                 `json:"location,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// HardwareInventoryRequest represents a request to submit hardware inventory
type HardwareInventoryRequest struct {
	ServerUUID       string                `json:"server_uuid"`
	CollectedAt      time.Time             `json:"collected_at"`
	CollectionMethod string                `json:"collection_method,omitempty"`
	Hardware         HardwareInventoryInfo `json:"hardware"`
}

// HardwareInventoryInfo contains detailed hardware information
type HardwareInventoryInfo struct {
	// Legacy flat fields for backward compatibility
	Manufacturer     string                 `json:"manufacturer,omitempty"`
	Model            string                 `json:"model,omitempty"`
	SerialNumber     string                 `json:"serial_number,omitempty"`
	CollectionMethod string                 `json:"collection_method,omitempty"`
	DetectionTool    string                 `json:"detection_tool,omitempty"`
	AdditionalInfo   map[string]interface{} `json:"additional_info,omitempty"`

	// Structured hardware components
	System              *SystemHardwareInfo    `json:"system,omitempty"`
	Motherboard         *MotherboardInfo       `json:"motherboard,omitempty"`
	CPUs                []CPUInfo              `json:"cpus,omitempty"`
	Memory              *MemoryInfo            `json:"memory,omitempty"`
	MemoryModules       []MemoryModuleInfo     `json:"memory_modules,omitempty"`
	Storage             []StorageDeviceInfo    `json:"storage,omitempty"`
	StorageDevices      []StorageDeviceInfo    `json:"storage_devices,omitempty"` // Alias for Storage
	Network             []NetworkCardInfo      `json:"network,omitempty"`
	NetworkCards        []NetworkCardInfo      `json:"network_cards,omitempty"` // Alias for Network
	GPUs                []GPUInfo              `json:"gpus,omitempty"`
	PowerSupplies       []PowerSupplyInfo      `json:"power_supplies,omitempty"`
	RAIDControllers     []RAIDControllerInfo   `json:"raid_controllers,omitempty"`
	TemperatureSensors  []TemperatureSensorInfo `json:"temperature_sensors,omitempty"`
	Services            *ServiceInfo           `json:"services,omitempty"`
}

// SystemHardwareInfo represents system-level hardware information
type SystemHardwareInfo struct {
	Manufacturer string `json:"manufacturer,omitempty"`
	ProductName  string `json:"product_name,omitempty"`
	Version      string `json:"version,omitempty"`
	SerialNumber string `json:"serial_number,omitempty"`
	UUID         string `json:"uuid,omitempty"`
	SKU          string `json:"sku,omitempty"`
}

// MotherboardInfo represents motherboard information
type MotherboardInfo struct {
	Manufacturer string    `json:"manufacturer,omitempty"`
	ProductName  string    `json:"product_name,omitempty"`
	Version      string    `json:"version,omitempty"`
	SerialNumber string    `json:"serial_number,omitempty"`
	AssetTag     string    `json:"asset_tag,omitempty"`
	BIOS         *BIOSInfo `json:"bios,omitempty"`
}

// BIOSInfo represents BIOS information
type BIOSInfo struct {
	Vendor      string `json:"vendor,omitempty"`
	Version     string `json:"version,omitempty"`
	ReleaseDate string `json:"release_date,omitempty"`
	Revision    string `json:"revision,omitempty"`
}

// CPUInfo represents CPU information
type CPUInfo struct {
	Manufacturer string  `json:"manufacturer,omitempty"`
	Model        string  `json:"model,omitempty"`
	Architecture string  `json:"architecture,omitempty"`
	Cores        int     `json:"cores,omitempty"`
	Threads      int     `json:"threads,omitempty"`
	BaseSpeedMHz float64 `json:"base_speed_mhz,omitempty"`
	MaxSpeedMHz  float64 `json:"max_speed_mhz,omitempty"`
	CacheSizeKB  int     `json:"cache_size_kb,omitempty"`
	Socket       string  `json:"socket,omitempty"`
}

// MemoryInfo represents memory information
type MemoryInfo struct {
	TotalCapacity      int64              `json:"total_capacity,omitempty"` // bytes
	TotalSizeGB        float64            `json:"total_size_gb,omitempty"`
	TotalSlots         int                `json:"total_slots,omitempty"`
	AvailableSlots     int                `json:"available_slots,omitempty"`
	UsedSlots          int                `json:"used_slots,omitempty"`
	MaxCapacityGB      float64            `json:"max_capacity_gb,omitempty"`
	MaxCapacityPerSlot int64              `json:"max_capacity_per_slot,omitempty"`
	SupportedTypes     []string           `json:"supported_types,omitempty"`
	SupportedSpeeds    []int              `json:"supported_speeds,omitempty"`
	ECCSupported       bool               `json:"ecc_supported,omitempty"`
	Modules            []MemoryModuleInfo `json:"modules,omitempty"`
}

// MemoryModuleInfo represents individual memory module information
type MemoryModuleInfo struct {
	Size         int64   `json:"size,omitempty"` // bytes
	SizeGB       float64 `json:"size_gb,omitempty"`
	Type         string  `json:"type,omitempty"`
	Speed        int     `json:"speed,omitempty"` // MHz
	SpeedMHz     int     `json:"speed_mhz,omitempty"`
	Manufacturer string  `json:"manufacturer,omitempty"`
	SerialNumber string  `json:"serial_number,omitempty"`
	PartNumber   string  `json:"part_number,omitempty"`
	Slot         string  `json:"slot,omitempty"`
	FormFactor   string  `json:"form_factor,omitempty"`
	ECC          bool    `json:"ecc,omitempty"`
	Registered   bool    `json:"registered,omitempty"`
}

// StorageDeviceInfo represents storage device information
type StorageDeviceInfo struct {
	DeviceName      string  `json:"device_name,omitempty"`
	Model           string  `json:"model,omitempty"`
	Vendor          string  `json:"vendor,omitempty"`
	Manufacturer    string  `json:"manufacturer,omitempty"`
	SerialNumber    string  `json:"serial_number,omitempty"`
	Capacity        int64   `json:"capacity,omitempty"` // bytes
	SizeGB          float64 `json:"size_gb,omitempty"`
	Type            string  `json:"type,omitempty"`
	Interface       string  `json:"interface,omitempty"`
	FirmwareVersion string  `json:"firmware_version,omitempty"`
	SmartStatus     string  `json:"smart_status,omitempty"`
	Health          string  `json:"health,omitempty"`
	Temperature     int     `json:"temperature,omitempty"`
	PowerOnHours    int64   `json:"power_on_hours,omitempty"`
	WriteEndurance  float64 `json:"write_endurance,omitempty"`
	FormFactor      string  `json:"form_factor,omitempty"`
}

// NetworkCardInfo represents network card information
type NetworkCardInfo struct {
	Model         string   `json:"model,omitempty"`
	Vendor        string   `json:"vendor,omitempty"`
	MACAddress    string   `json:"mac_address,omitempty"`
	SpeedMbps     int      `json:"speed_mbps,omitempty"`
	PortCount     int      `json:"port_count,omitempty"`
	Capabilities  []string `json:"capabilities,omitempty"`
	Driver        string   `json:"driver,omitempty"`
	DriverVersion string   `json:"driver_version,omitempty"`
}

// GPUInfo represents GPU information
type GPUInfo struct {
	Model         string  `json:"model,omitempty"`
	Vendor        string  `json:"vendor,omitempty"`
	Manufacturer  string  `json:"manufacturer,omitempty"`
	MemoryGB      float64 `json:"memory_gb,omitempty"`
	MemorySize    int64   `json:"memory_size,omitempty"`
	Driver        string  `json:"driver,omitempty"`
	DriverVersion string  `json:"driver_version,omitempty"`
	BusID         string  `json:"bus_id,omitempty"`
	Temperature   int     `json:"temperature,omitempty"`
}

// PowerSupplyInfo represents power supply information
type PowerSupplyInfo struct {
	Model             string  `json:"model,omitempty"`
	Manufacturer      string  `json:"manufacturer,omitempty"`
	SerialNumber      string  `json:"serial_number,omitempty"`
	MaxPowerWatts     int     `json:"max_power_watts,omitempty"`
	Type              string  `json:"type,omitempty"`
	Status            string  `json:"status,omitempty"`
	Efficiency        string  `json:"efficiency,omitempty"`
	CurrentPowerWatts float64 `json:"current_power_watts,omitempty"`
	Voltage           float64 `json:"voltage,omitempty"`
	Current           float64 `json:"current,omitempty"`
	Temperature       float64 `json:"temperature,omitempty"`
	FanSpeed          int     `json:"fan_speed,omitempty"`
	InputVoltage      float64 `json:"input_voltage,omitempty"`
	OutputVoltage     float64 `json:"output_voltage,omitempty"`
}

// SystemInfo represents system information for metrics
type SystemInfo struct {
	Hostname        string `json:"hostname"`
	OS              string `json:"os"`
	OSVersion       string `json:"os_version"`
	Kernel          string `json:"kernel"`
	KernelVersion   string `json:"kernel_version,omitempty"`
	Architecture    string `json:"architecture"`
	CPUArchitecture string `json:"cpu_architecture,omitempty"`
	Uptime          int64  `json:"uptime"`
	BootTime        int64  `json:"boot_time"`
	Processes       int    `json:"processes"`
	UsersLoggedIn   int    `json:"users_logged_in"`
	Platform        string `json:"platform,omitempty"`
	PlatformFamily  string `json:"platform_family,omitempty"`
}

// CPUMetrics represents CPU metrics
type CPUMetrics struct {
	UsagePercent  float64   `json:"usage_percent"`
	LoadAverage1  float64   `json:"load_average_1"`
	LoadAverage5  float64   `json:"load_average_5"`
	LoadAverage15 float64   `json:"load_average_15"`
	UserPercent   float64   `json:"user_percent"`
	SystemPercent float64   `json:"system_percent"`
	IdlePercent   float64   `json:"idle_percent"`
	IOWaitPercent float64   `json:"iowait_percent"`
	StealPercent  float64   `json:"steal_percent"`
	CoreCount     int       `json:"core_count"`
	ThreadCount   int       `json:"thread_count"`
	PerCoreUsage  []float64 `json:"per_core_usage,omitempty"`
}

// MemoryMetrics represents memory metrics
type MemoryMetrics struct {
	TotalBytes       int64   `json:"total_bytes"`
	UsedBytes        int64   `json:"used_bytes"`
	FreeBytes        int64   `json:"free_bytes"`
	AvailableBytes   int64   `json:"available_bytes"`
	UsagePercent     float64 `json:"usage_percent"`
	BuffersBytes     int64   `json:"buffers_bytes"`
	CachedBytes      int64   `json:"cached_bytes"`
	SwapTotalBytes   int64   `json:"swap_total_bytes"`
	SwapUsedBytes    int64   `json:"swap_used_bytes"`
	SwapFreeBytes    int64   `json:"swap_free_bytes"`
	SwapUsagePercent float64 `json:"swap_usage_percent"`
}

// DiskMetrics represents disk metrics
type DiskMetrics struct {
	Device             string  `json:"device"`
	Mountpoint         string  `json:"mountpoint"`
	Filesystem         string  `json:"filesystem"`
	TotalBytes         int64   `json:"total_bytes"`
	UsedBytes          int64   `json:"used_bytes"`
	FreeBytes          int64   `json:"free_bytes"`
	UsagePercent       float64 `json:"usage_percent"`
	InodesTotal        int64   `json:"inodes_total"`
	InodesUsed         int64   `json:"inodes_used"`
	InodesFree         int64   `json:"inodes_free"`
	InodesUsagePercent float64 `json:"inodes_usage_percent"`
}

// DiskUsageAggregate represents aggregated disk usage summary across all filesystems
type DiskUsageAggregate struct {
	TotalBytes      uint64   `json:"total_bytes"`      // Total bytes across all filesystems
	UsedBytes       uint64   `json:"used_bytes"`       // Used bytes across all filesystems
	FreeBytes       uint64   `json:"free_bytes"`       // Free bytes across all filesystems
	UsedPercent     float64  `json:"used_percent"`     // Overall usage percentage
	FilesystemCount int      `json:"filesystem_count"` // Number of filesystems included in aggregation
	LargestMount    string   `json:"largest_mount"`    // Mount point with largest capacity
	CriticalMounts  []string `json:"critical_mounts"`  // Mount points >90% full
	CalculatedAt    string   `json:"calculated_at"`    // ISO 8601 timestamp when aggregation was calculated
}

// NetworkMetrics represents network metrics
type NetworkMetrics struct {
	Interface   string `json:"interface"`
	BytesRecv   int64  `json:"bytes_recv"`
	BytesSent   int64  `json:"bytes_sent"`
	PacketsRecv int64  `json:"packets_recv"`
	PacketsSent int64  `json:"packets_sent"`
	ErrorsIn    int64  `json:"errors_in"`
	ErrorsOut   int64  `json:"errors_out"`
	DropsIn     int64  `json:"drops_in"`
	DropsOut    int64  `json:"drops_out"`
}

// ProcessMetrics represents process metrics
type ProcessMetrics struct {
	PID           int     `json:"pid"`
	Name          string  `json:"name"`
	Username      string  `json:"username"`
	State         string  `json:"state"`
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryPercent float64 `json:"memory_percent"`
	MemoryRSS     int64   `json:"memory_rss"`
	MemoryVMS     int64   `json:"memory_vms"`
	CreateTime    int64   `json:"create_time"`
	OpenFiles     int     `json:"open_files"`
	NumThreads    int     `json:"num_threads"`
}

// ComprehensiveMetricsRequest represents a comprehensive metrics submission
type ComprehensiveMetricsRequest struct {
	ServerUUID         string                 `json:"server_uuid"`
	CollectedAt        string                 `json:"collected_at"`
	SystemInfo         *SystemInfo            `json:"system_info,omitempty"`
	CPU                *CPUMetrics            `json:"cpu,omitempty"`
	Memory             *MemoryMetrics         `json:"memory,omitempty"`
	Disks              []DiskMetrics          `json:"disks,omitempty"`
	DiskUsageAggregate *DiskUsageAggregate    `json:"disk_usage_aggregate,omitempty"`
	Network            []NetworkMetrics       `json:"network,omitempty"`
	Processes          []ProcessMetrics       `json:"processes,omitempty"`
	Temperature        *TemperatureMetrics    `json:"temperature,omitempty"`
	Power              *PowerMetrics          `json:"power,omitempty"`
	Services           *ServiceInfo           `json:"services,omitempty"`
	CustomMetrics      map[string]interface{} `json:"custom_metrics,omitempty"`
}

// TimescaleDiskMetrics represents disk metrics for Timescale
type TimescaleDiskMetrics struct {
	Devices []TimescaleDiskDevice `json:"devices"`
}

// TimescaleDiskDevice represents individual disk device metrics
type TimescaleDiskDevice struct {
	Name                  string  `json:"name"`
	ReadCount             uint64  `json:"read_count"`
	WriteCount            uint64  `json:"write_count"`
	ReadBytes             uint64  `json:"read_bytes"`
	WriteBytes            uint64  `json:"write_bytes"`
	ReadTime              uint64  `json:"read_time"`
	WriteTime             uint64  `json:"write_time"`
	IoTime                uint64  `json:"io_time"`
	Size                  uint64  `json:"size"`
	ReadsPerSec           float64 `json:"reads_per_sec"`
	WritesPerSec          float64 `json:"writes_per_sec"`
	DiscardsPerSec        float64 `json:"discards_per_sec"`
	FlushesPerSec         float64 `json:"flushes_per_sec"`
	ReadKBPerSec          float64 `json:"read_kb_per_sec"`
	WriteKBPerSec         float64 `json:"write_kb_per_sec"`
	DiscardKBPerSec       float64 `json:"discard_kb_per_sec"`
	ReadMergePercent      float64 `json:"read_merge_percent"`
	WriteMergePercent     float64 `json:"write_merge_percent"`
	DiscardMergePercent   float64 `json:"discard_merge_percent"`
	AvgReadRequestSize    float64 `json:"avg_read_request_size"`
	AvgWriteRequestSize   float64 `json:"avg_write_request_size"`
	AvgDiscardRequestSize float64 `json:"avg_discard_request_size"`
	AvgReadWait           float64 `json:"avg_read_wait"`
	AvgWriteWait          float64 `json:"avg_write_wait"`
	AvgDiscardWait        float64 `json:"avg_discard_wait"`
	AvgFlushWait          float64 `json:"avg_flush_wait"`
	AvgQueueSize          float64 `json:"avg_queue_size"`
	Utilization           float64 `json:"utilization"`
	QueueDepth            uint64  `json:"queue_depth"`
}

// TimescaleNetworkMetrics represents network metrics for Timescale
type TimescaleNetworkMetrics struct {
	Interfaces []TimescaleNetworkInterface `json:"interfaces"`
}

// TimescaleNetworkInterface represents individual network interface metrics
type TimescaleNetworkInterface struct {
	Name        string  `json:"name"`
	BytesSent   uint64  `json:"bytes_sent"`
	BytesRecv   uint64  `json:"bytes_recv"`
	PacketsSent uint64  `json:"packets_sent"`
	PacketsRecv uint64  `json:"packets_recv"`
	Errin       uint64  `json:"errin"`
	Errout      uint64  `json:"errout"`
	Dropin      uint64  `json:"dropin"`
	Dropout     uint64  `json:"dropout"`
	Speed       uint64  `json:"speed"`
	Mtu         uint64  `json:"mtu"`
	State       string  `json:"state"`
	RxRateKbps  float64 `json:"rx_rate_kbps"`
	TxRateKbps  float64 `json:"tx_rate_kbps"`
}

// TimescaleFilesystemMetrics represents filesystem metrics for Timescale
type TimescaleFilesystemMetrics struct {
	Filesystems []TimescaleFilesystem `json:"filesystems"`
}

// TimescaleFilesystem represents individual filesystem metrics
type TimescaleFilesystem struct {
	Device      string  `json:"device"`
	Mountpoint  string  `json:"mountpoint"`
	Fstype      string  `json:"fstype"`
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
	InodesTotal uint64  `json:"inodes_total"`
	InodesUsed  uint64  `json:"inodes_used"`
	InodesFree  uint64  `json:"inodes_free"`
}

// TimescaleHostInfo represents host system information
type TimescaleHostInfo struct {
	Hostname       string `json:"hostname"`
	Uptime         uint64 `json:"uptime"`
	BootTime       uint64 `json:"boot_time"`
	Procs          uint32 `json:"procs"`
	OS             string `json:"os"`
	Platform       string `json:"platform"`
	PlatformFamily string `json:"platform_family"`
}

// RAIDControllerInfo represents RAID controller information
type RAIDControllerInfo struct {
	Manufacturer    string             `json:"manufacturer,omitempty"`
	Model           string             `json:"model,omitempty"`
	FirmwareVersion string             `json:"firmware_version,omitempty"`
	CacheSize       int64              `json:"cache_size,omitempty"`
	BatteryBackup   bool               `json:"battery_backup,omitempty"`
	BatteryStatus   string             `json:"battery_status,omitempty"`
	RAIDLevels      []string           `json:"raid_levels,omitempty"`
	LogicalDrives   []RAIDLogicalDrive `json:"logical_drives,omitempty"`
}

// RAIDLogicalDrive represents a logical drive on a RAID controller
type RAIDLogicalDrive struct {
	ID        string `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	RAIDLevel string `json:"raid_level,omitempty"`
	Size      int64  `json:"size,omitempty"`
	Status    string `json:"status,omitempty"`
	DiskCount int    `json:"disk_count,omitempty"`
}

// HardwareInventorySubmitResponse represents the response from submitting hardware inventory
type HardwareInventorySubmitResponse struct {
	ServerUUID       string         `json:"server_uuid"`
	Timestamp        time.Time      `json:"timestamp"`
	CollectionMethod string         `json:"collection_method,omitempty"`
	DetectionTool    string         `json:"detection_tool,omitempty"`
	ComponentCounts  map[string]int `json:"component_counts,omitempty"`
}

// HardwareInventoryRecord represents a hardware inventory record with metadata
type HardwareInventoryRecord struct {
	ID               uint                  `json:"id"`
	ServerUUID       string                `json:"server_uuid"`
	OrganizationID   uint                  `json:"organization_id"`
	CollectedAt      time.Time             `json:"collected_at"`
	CollectionMethod string                `json:"collection_method,omitempty"`
	DetectionTool    string                `json:"detection_tool,omitempty"`
	CreatedAt        time.Time             `json:"created_at"`
	UpdatedAt        time.Time             `json:"updated_at"`
	Hardware         HardwareInventoryInfo `json:"hardware"`
	Timestamp        time.Time             `json:"timestamp"`
}

// HardwareInventoryListOptions represents options for listing hardware inventory
type HardwareInventoryListOptions struct {
	ListOptions
	StartTime *time.Time `url:"start_time,omitempty"`
	EndTime   *time.Time `url:"end_time,omitempty"`
}

// ToQuery converts HardwareInventoryListOptions to query parameters
func (o *HardwareInventoryListOptions) ToQuery() map[string]string {
	params := o.ListOptions.ToQuery()
	if o.StartTime != nil {
		params["start_time"] = o.StartTime.Format(time.RFC3339)
	}
	if o.EndTime != nil {
		params["end_time"] = o.EndTime.Format(time.RFC3339)
	}
	return params
}

// Controller health and heartbeat types
type ControllerHealthInfo struct {
	Status        string                 `json:"status"`
	Version       string                 `json:"version"`
	Uptime        time.Duration          `json:"uptime"`
	LastHeartbeat *CustomTime            `json:"last_heartbeat,omitempty"`
	ResourceUsage *ResourceUsageInfo     `json:"resource_usage,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`

	// Additional fields needed by monitoring-controller:
	ServiceComponents map[string]string `json:"service_components,omitempty"`
	RegionHealth      map[string]string `json:"region_health,omitempty"`
	LastErrors        []string          `json:"last_errors,omitempty"`
}

type ControllerHeartbeatRequest struct {
	ControllerID  string                `json:"controller_id"`
	Status        string                `json:"status"`
	Version       string                `json:"version"`
	Health        *ControllerHealthInfo `json:"health"`
	ResourceUsage *ResourceUsageInfo    `json:"resource_usage,omitempty"`
	Timestamp     time.Time             `json:"timestamp"`

	// Additional fields needed by monitoring-controller:
	ControllerName    string                 `json:"controller_name"`
	ControllerType    string                 `json:"controller_type"`
	HeartbeatInterval int                    `json:"heartbeat_interval"`
	IsLeader          bool                   `json:"is_leader"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	HealthDetails     *ControllerHealthInfo  `json:"health_details,omitempty"`
}

type ResourceUsageInfo struct {
	CPUUsage       float64 `json:"cpu_usage"`
	MemoryUsage    int64   `json:"memory_usage"`
	MemoryLimit    int64   `json:"memory_limit,omitempty"`
	DiskUsage      int64   `json:"disk_usage,omitempty"`
	NetworkRxBytes int64   `json:"network_rx_bytes,omitempty"`
	NetworkTxBytes int64   `json:"network_tx_bytes,omitempty"`
}

// Regional monitoring types
type MonitoringRegion struct {
	GormModel
	Code        string                 `json:"code"`
	Name        string                 `json:"name"`
	Status      RegionStatus           `json:"status"`
	Location    string                 `json:"location,omitempty"`
	Description string                 `json:"description,omitempty"`
	Enabled     bool                   `json:"enabled"`
	Priority    int                    `json:"priority"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type RegionStatus string

const (
	RegionStatusActive      RegionStatus = "active"
	RegionStatusInactive    RegionStatus = "inactive"
	RegionStatusMaintenance RegionStatus = "maintenance"
)

// Remote cluster types
type RemoteCluster struct {
	GormModel
	Name         string                 `json:"name"`
	Endpoint     string                 `json:"endpoint"`
	Region       string                 `json:"region"`
	Status       string                 `json:"status"`
	Version      string                 `json:"version,omitempty"`
	Capabilities []string               `json:"capabilities,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	LastSeen     *CustomTime            `json:"last_seen,omitempty"`
}

// ControllerHeartbeat represents a controller heartbeat record
type ControllerHeartbeat struct {
	GormModel
	ControllerID  string                `json:"controller_id"`
	Status        string                `json:"status"`
	Version       string                `json:"version"`
	Health        *ControllerHealthInfo `json:"health"`
	ResourceUsage *ResourceUsageInfo    `json:"resource_usage,omitempty"`
	Timestamp     time.Time             `json:"timestamp"`
}

// NamespaceDeployment represents a deployment within a namespace
type NamespaceDeployment struct {
	GormModel
	Name           string                 `json:"name"`
	Namespace      string                 `json:"namespace"`
	OrganizationID uint                   `json:"organization_id"`
	Status         string                 `json:"status"`
	AgentVersion   string                 `json:"agent_version"`
	Configuration  map[string]interface{} `json:"configuration,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	LastUpdated    *CustomTime            `json:"last_updated,omitempty"`
}

// ProbeCreateRequest represents a request to create a probe
type ProbeCreateRequest struct {
	Name           string                 `json:"name"`
	Type           string                 `json:"type"`
	Target         string                 `json:"target"`
	Configuration  map[string]interface{} `json:"configuration,omitempty"`
	Interval       int                    `json:"interval"`
	Timeout        int                    `json:"timeout"`
	OrganizationID uint                   `json:"organization_id"`
	RegionCode     string                 `json:"region_code,omitempty"`
	Enabled        bool                   `json:"enabled"`
}

// ProbeUpdateRequest represents a request to update a probe
type ProbeUpdateRequest struct {
	Name          *string                `json:"name,omitempty"`
	Type          *string                `json:"type,omitempty"`
	Target        *string                `json:"target,omitempty"`
	Configuration map[string]interface{} `json:"configuration,omitempty"`
	Interval      *int                   `json:"interval,omitempty"`
	Timeout       *int                   `json:"timeout,omitempty"`
	RegionCode    *string                `json:"region_code,omitempty"`
	Enabled       *bool                  `json:"enabled,omitempty"`
}

// ProbeMetricsOptions represents options for retrieving probe metrics
type ProbeMetricsOptions struct {
	ProbeUUID   string     `json:"probe_uuid"`
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	Granularity string     `json:"granularity,omitempty"` // minute, hour, day
	Aggregation string     `json:"aggregation,omitempty"` // avg, min, max, sum
}

// AggregatedMetricsRequest represents a request to submit aggregated metrics
type AggregatedMetricsRequest struct {
	ServerUUID  string               `json:"server_uuid"`
	CollectedAt string               `json:"collected_at"`
	Disk        *DiskAggregation     `json:"disk,omitempty"`
	CPU         *CPUAggregation      `json:"cpu,omitempty"`
	Memory      *MemoryAggregation   `json:"memory,omitempty"`
	Network     *NetworkAggregation  `json:"network,omitempty"`
	GPU         *GPUAggregation      `json:"gpu,omitempty"`
}

// DiskAggregation represents aggregated disk metrics
type DiskAggregation struct {
	TotalBytes      uint64   `json:"total_bytes"`
	UsedBytes       uint64   `json:"used_bytes"`
	FreeBytes       uint64   `json:"free_bytes"`
	UsedPercent     float64  `json:"used_percent"`
	FilesystemCount int      `json:"filesystem_count"`
	LargestMount    string   `json:"largest_mount"`
	CriticalMounts  []string `json:"critical_mounts"`
	CalculatedAt    string   `json:"calculated_at"`
}

// CPUAggregation represents aggregated CPU metrics
type CPUAggregation struct {
	UsagePercent   float64 `json:"usage_percent"`
	LoadAverage1   float64 `json:"load_average_1"`
	LoadAverage5   float64 `json:"load_average_5"`
	LoadAverage15  float64 `json:"load_average_15"`
	CoreCount      int     `json:"core_count"`
	MaxCorePercent float64 `json:"max_core_percent"`
	MinCorePercent float64 `json:"min_core_percent"`
	StealPercent   float64 `json:"steal_percent"`
	IOWaitPercent  float64 `json:"iowait_percent"`
	CalculatedAt   string  `json:"calculated_at"`
}

// MemoryAggregation represents aggregated memory metrics
type MemoryAggregation struct {
	TotalBytes      uint64  `json:"total_bytes"`
	UsedBytes       uint64  `json:"used_bytes"`
	FreeBytes       uint64  `json:"free_bytes"`
	AvailableBytes  uint64  `json:"available_bytes"`
	UsedPercent     float64 `json:"used_percent"`
	SwapTotalBytes  uint64  `json:"swap_total_bytes"`
	SwapUsedBytes   uint64  `json:"swap_used_bytes"`
	SwapUsedPercent float64 `json:"swap_used_percent"`
	CacheBytes      uint64  `json:"cache_bytes"`
	BufferBytes     uint64  `json:"buffer_bytes"`
	CalculatedAt    string  `json:"calculated_at"`
}

// NetworkAggregation represents aggregated network metrics
type NetworkAggregation struct {
	TotalBandwidthBPS uint64  `json:"total_bandwidth_bps"`
	IngressBPS        uint64  `json:"ingress_bps"`
	EgressBPS         uint64  `json:"egress_bps"`
	TotalInterfaces   int     `json:"total_interfaces"`
	ActiveInterfaces  int     `json:"active_interfaces"`
	ErrorRate         float64 `json:"error_rate"`
	DropRate          float64 `json:"drop_rate"`
	PrimaryInterface  string  `json:"primary_interface"`
	CalculatedAt      string  `json:"calculated_at"`
}

// GPUAggregation represents aggregated GPU metrics
type GPUAggregation struct {
	TotalGPUs         int     `json:"total_gpus"`
	AvgUsagePercent   float64 `json:"avg_usage_percent"`
	MaxUsagePercent   float64 `json:"max_usage_percent"`
	TotalMemoryBytes  uint64  `json:"total_memory_bytes"`
	UsedMemoryBytes   uint64  `json:"used_memory_bytes"`
	MemoryUsedPercent float64 `json:"memory_used_percent"`
	AvgTemperature    float64 `json:"avg_temperature"`
	MaxTemperature    float64 `json:"max_temperature"`
	PowerUsageWatts   float64 `json:"power_usage_watts"`
	CalculatedAt      string  `json:"calculated_at"`
}

// TemperatureMetrics represents temperature sensor data
type TemperatureMetrics struct {
	Sensors []TemperatureSensorData `json:"sensors,omitempty"`
}

// TemperatureSensorData represents individual temperature sensor data
type TemperatureSensorData struct {
	SensorID      string  `json:"sensor_id"`
	SensorName    string  `json:"sensor_name"`
	Temperature   float64 `json:"temperature"`      // in Celsius
	Status        string  `json:"status"`           // ok, warning, critical
	Type          string  `json:"type,omitempty"`   // cpu, system, disk, gpu, etc.
	Location      string  `json:"location,omitempty"`
	UpperWarning  float64 `json:"upper_warning,omitempty"`
	UpperCritical float64 `json:"upper_critical,omitempty"`
}

// PowerMetrics represents power supply data
type PowerMetrics struct {
	PowerSupplies []PowerSupplyMetrics `json:"power_supplies,omitempty"`
	TotalPowerW   float64             `json:"total_power_watts,omitempty"`
}

// PowerSupplyMetrics represents individual power supply metrics
type PowerSupplyMetrics struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Status        string  `json:"status"`        // ok, warning, critical, failed
	PowerWatts    float64 `json:"power_watts"`   // current power draw
	MaxPowerWatts float64 `json:"max_power_watts"`
	Voltage       float64 `json:"voltage,omitempty"`
	Current       float64 `json:"current,omitempty"`
	Efficiency    float64 `json:"efficiency,omitempty"` // percentage
	Temperature   float64 `json:"temperature,omitempty"`
}

// TemperatureSensorInfo represents temperature sensor information for hardware inventory
type TemperatureSensorInfo struct {
	SensorID      string  `json:"sensor_id"`
	SensorName    string  `json:"sensor_name"`
	Type          string  `json:"type"`
	Location      string  `json:"location,omitempty"`
	MaxTemp       float64 `json:"max_temp,omitempty"`
	MinTemp       float64 `json:"min_temp,omitempty"`
}

// ServiceInfo contains service monitoring data
type ServiceInfo struct {
	Services []*ServiceMonitoringInfo      `json:"services,omitempty"`
	Metrics  []*ServiceMetrics             `json:"metrics,omitempty"`
	Logs     map[string][]ServiceLogEntry  `json:"logs,omitempty"`
}

// ServiceMonitoringInfo represents service monitoring data
type ServiceMonitoringInfo struct {
	Name          string     `json:"name"`
	State         string     `json:"state"`         // active, inactive, failed
	SubState      string     `json:"sub_state"`     // running, dead, exited, etc.
	LoadState     string     `json:"load_state"`    // loaded, not-found, masked
	Description   string     `json:"description"`
	MainPID       int        `json:"main_pid"`
	MemoryCurrent uint64     `json:"memory_current"`
	CPUUsageNSec  uint64     `json:"cpu_usage_nsec"`
	TasksCurrent  uint64     `json:"tasks_current"`
	RestartCount  int        `json:"restart_count"`
	ActiveSince   *time.Time `json:"active_since,omitempty"`
}

// ServiceMetrics represents service resource metrics
type ServiceMetrics struct {
	ServiceName  string    `json:"service_name"`
	Timestamp    time.Time `json:"timestamp"`
	CPUPercent   float64   `json:"cpu_percent"`
	MemoryRSS    uint64    `json:"memory_rss"`
	ProcessCount int       `json:"process_count"`
	ThreadCount  int       `json:"thread_count"`
}

// ServiceLogEntry represents a service log entry
type ServiceLogEntry struct {
	Timestamp time.Time         `json:"timestamp"`
	Level     string            `json:"level"`
	Message   string            `json:"message"`
	Fields    map[string]string `json:"fields,omitempty"`
}

// ServiceMonitoringConfig represents configuration for service monitoring
type ServiceMonitoringConfig struct {
	Enabled         bool     `json:"enabled"`          // Enable service monitoring
	IncludeServices []string `json:"include_services"` // Specific services to monitor
	ExcludeServices []string `json:"exclude_services"` // Services to exclude
	IncludePatterns []string `json:"include_patterns"` // Patterns like "nginx*", "apache*"
	ExcludePatterns []string `json:"exclude_patterns"` // Patterns like "*.scope", "*.slice"
	CollectMetrics  bool     `json:"collect_metrics"`  // Collect resource metrics
	CollectLogs     bool     `json:"collect_logs"`     // Collect recent logs
	LogLines        int      `json:"log_lines"`        // Number of recent log lines
	MetricsInterval string   `json:"metrics_interval"` // How often to collect metrics
	LogStateFile    string   `json:"log_state_file"`   // Path to log state file
}

// Type alias for backward compatibility
type Probe = MonitoringProbe

// IncidentSeverity represents the severity level of an incident
type IncidentSeverity string

const (
	// IncidentSeverityCritical indicates a critical incident requiring immediate attention
	IncidentSeverityCritical IncidentSeverity = "critical"
	// IncidentSeverityWarning indicates a warning-level incident
	IncidentSeverityWarning IncidentSeverity = "warning"
	// IncidentSeverityInfo indicates an informational incident
	IncidentSeverityInfo IncidentSeverity = "info"
)

// IncidentStatus represents the current status of an incident
type IncidentStatus string

const (
	// IncidentStatusActive indicates the incident is active and needs attention
	IncidentStatusActive IncidentStatus = "active"
	// IncidentStatusResolved indicates the incident has been resolved
	IncidentStatusResolved IncidentStatus = "resolved"
	// IncidentStatusAcknowledged indicates the incident has been acknowledged
	IncidentStatusAcknowledged IncidentStatus = "acknowledged"
)

// IncidentSource represents the source that created the incident
type IncidentSource string

const (
	// IncidentSourceProbe indicates the incident was created from a probe failure
	IncidentSourceProbe IncidentSource = "probe"
	// IncidentSourceAlert indicates the incident was created from an alert
	IncidentSourceAlert IncidentSource = "alert"
	// IncidentSourceManual indicates the incident was created manually
	IncidentSourceManual IncidentSource = "manual"
)

// IncidentEventType represents the type of event in an incident timeline
type IncidentEventType string

const (
	// IncidentEventTypeCreated indicates the incident was created
	IncidentEventTypeCreated IncidentEventType = "created"
	// IncidentEventTypeUpdated indicates the incident was updated
	IncidentEventTypeUpdated IncidentEventType = "updated"
	// IncidentEventTypeResolved indicates the incident was resolved
	IncidentEventTypeResolved IncidentEventType = "resolved"
	// IncidentEventTypeAcknowledged indicates the incident was acknowledged
	IncidentEventTypeAcknowledged IncidentEventType = "acknowledged"
	// IncidentEventTypeEscalated indicates the incident was escalated
	IncidentEventTypeEscalated IncidentEventType = "escalated"
	// IncidentEventTypeAssigned indicates the incident was assigned to someone
	IncidentEventTypeAssigned IncidentEventType = "assigned"
	// IncidentEventTypeComment indicates a comment was added to the incident
	IncidentEventTypeComment IncidentEventType = "comment"
)

// AffectedResource represents a resource affected by an incident
type AffectedResource struct {
	Type string `json:"type"` // "server", "probe", "service"
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// Incident represents an incident in the system
type Incident struct {
	GormModel
	OrganizationID    uint                  `json:"organization_id"`
	Organization      *Organization         `json:"organization,omitempty"`
	Title             string                `json:"title"`
	Description       string                `json:"description"`
	Severity          IncidentSeverity      `json:"severity"`
	Status            IncidentStatus        `json:"status"`
	Source            IncidentSource        `json:"source"`
	SourceID          *uint                 `json:"source_id,omitempty"`
	AffectedResources []AffectedResource    `json:"affected_resources"`
	StartedAt         *CustomTime           `json:"started_at"`
	ResolvedAt        *CustomTime           `json:"resolved_at,omitempty"`
	Events            []IncidentEvent       `json:"events,omitempty"`
}

// IncidentEvent represents an event in an incident timeline
type IncidentEvent struct {
	GormModel
	IncidentID uint              `json:"incident_id"`
	Incident   *Incident         `json:"incident,omitempty"`
	EventType  IncidentEventType `json:"event_type"`
	Message    string            `json:"message"`
	CreatedBy  *uint             `json:"created_by,omitempty"`
	User       *User             `json:"user,omitempty"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// CreateIncidentRequest represents a request to create an incident
type CreateIncidentRequest struct {
	Title             string             `json:"title"`
	Description       string             `json:"description"`
	Severity          IncidentSeverity   `json:"severity"`
	ServerID          *uint              `json:"server_id,omitempty"`
	ProbeID           *uint              `json:"probe_id,omitempty"`
	Tags              []string           `json:"tags,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateIncidentRequest represents a request to update an incident
type UpdateIncidentRequest struct {
	Title       string                 `json:"title,omitempty"`
	Description string                 `json:"description,omitempty"`
	Severity    IncidentSeverity       `json:"severity,omitempty"`
	Status      IncidentStatus         `json:"status,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// IncidentListOptions represents options for listing incidents
type IncidentListOptions struct {
	ListOptions
	Status   string `url:"status,omitempty"`
	Severity string `url:"severity,omitempty"`
	ServerID uint   `url:"server_id,omitempty"`
	ProbeID  uint   `url:"probe_id,omitempty"`
	Sort     string `url:"sort,omitempty"`
}

// IncidentStats represents incident statistics
type IncidentStats struct {
	TotalCount     int            `json:"total_count"`
	ActiveCount    int            `json:"active_count"`
	BySeverity     map[string]int `json:"by_severity"`
	MTTRMinutes    float64        `json:"mttr_minutes"`
	RecentCount    int64          `json:"recent_count"`
	RecentResolved int64          `json:"recent_resolved"`
	RecentMTTR     float64        `json:"recent_mttr"`
}

// =============================================================================
// Unified API Key System
// =============================================================================

// APIKeyType represents the type of API key
type APIKeyType string

const (
	// APIKeyTypeUser represents user-created keys for personal access
	APIKeyTypeUser APIKeyType = "user"
	// APIKeyTypeAdmin represents admin keys with elevated permissions
	APIKeyTypeAdmin APIKeyType = "admin"
	// APIKeyTypeMonitoringAgent represents keys for monitoring agents
	APIKeyTypeMonitoringAgent APIKeyType = "monitoring_agent"
	// APIKeyTypeSystem represents keys for system-to-system communication
	APIKeyTypeSystem APIKeyType = "system"
	// APIKeyTypePublicAgent represents keys for public monitoring agents
	APIKeyTypePublicAgent APIKeyType = "public_agent"
	// APIKeyTypeRegistration represents keys for server registration
	APIKeyTypeRegistration APIKeyType = "registration"
	// APIKeyTypeOrgMonitoring represents organization-level monitoring keys
	APIKeyTypeOrgMonitoring APIKeyType = "org_monitoring"
)

// APIKeyStatus represents the status of an API key
type APIKeyStatus string

const (
	// APIKeyStatusActive indicates the key is active and can be used
	APIKeyStatusActive APIKeyStatus = "active"
	// APIKeyStatusRevoked indicates the key has been revoked
	APIKeyStatusRevoked APIKeyStatus = "revoked"
	// APIKeyStatusExpired indicates the key has expired
	APIKeyStatusExpired APIKeyStatus = "expired"
	// APIKeyStatusPending indicates the key is pending activation
	APIKeyStatusPending APIKeyStatus = "pending"
)

// UnifiedAPIKey represents a unified API key that supports all key types and capabilities
type UnifiedAPIKey struct {
	GormModel

	// Basic identification
	KeyID       string `json:"key_id"`                // Unique identifier for the key
	Name        string `json:"name"`                  // Human-readable name
	Description string `json:"description,omitempty"` // Optional description
	KeyPrefix   string `json:"key_prefix"`            // First few characters for display

	// Key data (only returned on creation for security)
	Key       string `json:"key,omitempty"`        // The actual key (only on creation)
	Secret    string `json:"secret,omitempty"`     // The secret part (only on creation)
	FullToken string `json:"full_token,omitempty"` // Complete token (only on creation)

	// Type and permissions
	Type         APIKeyType `json:"type"`              // Type of API key
	Capabilities []string   `json:"capabilities"`      // Fine-grained permissions
	Scopes       []string   `json:"scopes,omitempty"`  // Legacy scopes for backward compatibility

	// Ownership and organization
	OrganizationID uint          `json:"organization_id"`
	Organization   *Organization `json:"organization,omitempty"`
	UserID         *uint         `json:"user_id,omitempty"`
	User           *User         `json:"user,omitempty"`

	// Monitoring-specific fields (for monitoring agent keys)
	RemoteClusterID    *uint          `json:"remote_cluster_id,omitempty"`
	RemoteCluster      *RemoteCluster `json:"remote_cluster,omitempty"`
	NamespaceName      string         `json:"namespace_name,omitempty"`
	AgentType          string         `json:"agent_type,omitempty"`          // public, private
	RegionCode         string         `json:"region_code,omitempty"`
	AllowedProbeScopes []string       `json:"allowed_probe_scopes,omitempty"`

	// Status and usage tracking
	Status     APIKeyStatus `json:"status"`                  // Current status
	ExpiresAt  *CustomTime  `json:"expires_at,omitempty"`    // Optional expiration
	LastUsedAt *CustomTime  `json:"last_used_at,omitempty"`  // Last usage timestamp
	LastUsedIP string       `json:"last_used_ip,omitempty"`  // Last used IP address
	UsageCount int          `json:"usage_count"`             // Number of times used

	// Security and rate limiting
	RateLimitPerHour int      `json:"rate_limit_per_hour,omitempty"` // Requests per hour limit
	AllowedIPs       []string `json:"allowed_ips,omitempty"`          // IP whitelist

	// Metadata and tagging
	Tags     []string               `json:"tags,omitempty"`     // Tags for organization
	Metadata map[string]interface{} `json:"metadata,omitempty"` // Custom metadata
}

// IsActive returns true if the API key is active and not expired
func (k *UnifiedAPIKey) IsActive() bool {
	if k.Status != APIKeyStatusActive {
		return false
	}
	if k.ExpiresAt != nil && k.ExpiresAt.Before(time.Now()) {
		return false
	}
	return true
}

// IsExpired returns true if the API key has expired
func (k *UnifiedAPIKey) IsExpired() bool {
	return k.ExpiresAt != nil && k.ExpiresAt.Before(time.Now())
}

// IsRevoked returns true if the API key has been revoked
func (k *UnifiedAPIKey) IsRevoked() bool {
	return k.Status == APIKeyStatusRevoked
}

// HasCapability checks if the API key has the specified capability
func (k *UnifiedAPIKey) HasCapability(capability string) bool {
	for _, cap := range k.Capabilities {
		if cap == capability || cap == "*" {
			return true
		}
	}
	return false
}

// HasScope checks if the API key has the specified scope (for backward compatibility)
func (k *UnifiedAPIKey) HasScope(scope string) bool {
	for _, s := range k.Scopes {
		if s == scope || s == "*" {
			return true
		}
	}
	return false
}

// IsMonitoringAgent returns true if this is a monitoring agent key
func (k *UnifiedAPIKey) IsMonitoringAgent() bool {
	return k.Type == APIKeyTypeMonitoringAgent || k.Type == APIKeyTypePublicAgent
}

// IsRegistrationKey returns true if this is a registration key
func (k *UnifiedAPIKey) IsRegistrationKey() bool {
	return k.Type == APIKeyTypeRegistration
}

// CanRegisterServers returns true if this key can register servers
func (k *UnifiedAPIKey) CanRegisterServers() bool {
	return k.IsRegistrationKey() || k.HasCapability("servers:register") || k.HasCapability("servers:*") || k.HasCapability("*")
}

// CanAccessOrganization returns true if this key can access the specified organization
func (k *UnifiedAPIKey) CanAccessOrganization(orgID uint) bool {
	return k.OrganizationID == orgID || k.Type == APIKeyTypeSystem || k.Type == APIKeyTypeAdmin
}

// IsPublicAgent returns true if this is a public monitoring agent key
func (k *UnifiedAPIKey) IsPublicAgent() bool {
	return k.Type == APIKeyTypePublicAgent || (k.Type == APIKeyTypeMonitoringAgent && k.AgentType == "public")
}

// IsPrivateAgent returns true if this is a private monitoring agent key
func (k *UnifiedAPIKey) IsPrivateAgent() bool {
	return k.Type == APIKeyTypeMonitoringAgent && k.AgentType == "private"
}

// GetAuthenticationMethod returns the preferred authentication method for this key type
func (k *UnifiedAPIKey) GetAuthenticationMethod() string {
	switch k.Type {
	case APIKeyTypeMonitoringAgent, APIKeyTypePublicAgent:
		return "bearer" // Use Bearer token
	case APIKeyTypeRegistration:
		return "headers" // Use X-Registration-Key header
	default:
		return "headers" // Use Access-Key/Access-Secret headers
	}
}

// CreateUnifiedAPIKeyRequest represents a request to create a unified API key
type CreateUnifiedAPIKeyRequest struct {
	Name               string            `json:"name"`
	Description        string            `json:"description,omitempty"`
	Type               APIKeyType        `json:"type"`
	Capabilities       []string          `json:"capabilities,omitempty"`
	OrganizationID     uint              `json:"organization_id,omitempty"`     // Only for admin creation
	ExpiresAt          *CustomTime       `json:"expires_at,omitempty"`
	RateLimitPerHour   int               `json:"rate_limit_per_hour,omitempty"`
	AllowedIPs         []string          `json:"allowed_ips,omitempty"`
	Tags               []string          `json:"tags,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`

	// Monitoring-specific fields
	RemoteClusterID    *uint    `json:"remote_cluster_id,omitempty"`
	NamespaceName      string   `json:"namespace_name,omitempty"`
	AgentType          string   `json:"agent_type,omitempty"`          // public, private
	RegionCode         string   `json:"region_code,omitempty"`
	AllowedProbeScopes []string `json:"allowed_probe_scopes,omitempty"`
}

// CreateUnifiedAPIKeyResponse represents the response when creating a unified API key
type CreateUnifiedAPIKeyResponse struct {
	Key       *UnifiedAPIKey `json:"key"`
	KeyID     string         `json:"key_id"`
	KeyValue  string         `json:"key_value"`         // The actual key
	Secret    string         `json:"secret,omitempty"`  // Secret if using key/secret auth
	FullToken string         `json:"full_token"`        // Complete token for bearer auth
}

// UpdateUnifiedAPIKeyRequest represents a request to update a unified API key
type UpdateUnifiedAPIKeyRequest struct {
	Name             *string           `json:"name,omitempty"`
	Description      *string           `json:"description,omitempty"`
	Capabilities     []string          `json:"capabilities,omitempty"`
	Status           APIKeyStatus      `json:"status,omitempty"`
	ExpiresAt        *CustomTime       `json:"expires_at,omitempty"`
	RateLimitPerHour *int              `json:"rate_limit_per_hour,omitempty"`
	AllowedIPs       []string          `json:"allowed_ips,omitempty"`
	Tags             []string          `json:"tags,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// ListUnifiedAPIKeysOptions represents options for listing unified API keys
type ListUnifiedAPIKeysOptions struct {
	ListOptions
	Type         APIKeyType   `url:"type,omitempty"`
	Status       APIKeyStatus `url:"status,omitempty"`
	UserID       uint         `url:"user_id,omitempty"`
	AgentType    string       `url:"agent_type,omitempty"`
	RegionCode   string       `url:"region_code,omitempty"`
	Namespace    string       `url:"namespace,omitempty"`
	Capability   string       `url:"capability,omitempty"`
	Tag          string       `url:"tag,omitempty"`
}

// Backward compatibility type alias
// This allows existing code to continue working while gradually migrating to UnifiedAPIKey
type APIKey = UnifiedAPIKey

// Legacy constructor functions for backward compatibility
// These create UnifiedAPIKey instances with appropriate types

// NewUserAPIKey creates a new user API key request
func NewUserAPIKey(name, description string, capabilities []string) *CreateUnifiedAPIKeyRequest {
	return &CreateUnifiedAPIKeyRequest{
		Name:         name,
		Description:  description,
		Type:         APIKeyTypeUser,
		Capabilities: capabilities,
	}
}

// NewAdminAPIKey creates a new admin API key request
func NewAdminAPIKey(name, description string, capabilities []string, orgID uint) *CreateUnifiedAPIKeyRequest {
	return &CreateUnifiedAPIKeyRequest{
		Name:           name,
		Description:    description,
		Type:           APIKeyTypeAdmin,
		Capabilities:   capabilities,
		OrganizationID: orgID,
	}
}

// NewMonitoringAgentKey creates a new monitoring agent key request
func NewMonitoringAgentKey(name, description, namespace, agentType, regionCode string, allowedScopes []string) *CreateUnifiedAPIKeyRequest {
	return &CreateUnifiedAPIKeyRequest{
		Name:               name,
		Description:        description,
		Type:               APIKeyTypeMonitoringAgent,
		NamespaceName:      namespace,
		AgentType:          agentType,
		RegionCode:         regionCode,
		AllowedProbeScopes: allowedScopes,
		Capabilities:       []string{"monitoring:execute", "probes:execute"},
	}
}

// NewRegistrationKey creates a new registration key request
func NewRegistrationKey(name, description string, orgID uint) *CreateUnifiedAPIKeyRequest {
	return &CreateUnifiedAPIKeyRequest{
		Name:           name,
		Description:    description,
		Type:           APIKeyTypeRegistration,
		OrganizationID: orgID,
		Capabilities:   []string{"servers:register", "servers:update"},
	}
}

// NewServerDetailsUpdateRequest creates a new server details update request with hardware information
func NewServerDetailsUpdateRequest() *ServerDetailsUpdateRequest {
	return &ServerDetailsUpdateRequest{}
}

// WithBasicInfo sets basic server information
func (r *ServerDetailsUpdateRequest) WithBasicInfo(hostname, mainIP, environment, location, classification string) *ServerDetailsUpdateRequest {
	r.Hostname = hostname
	r.MainIP = mainIP
	r.Environment = environment
	r.Location = location
	r.Classification = classification
	return r
}

// WithSystemInfo sets system information
func (r *ServerDetailsUpdateRequest) WithSystemInfo(os, osVersion, osArch, serialNumber, macAddress string) *ServerDetailsUpdateRequest {
	r.OS = os
	r.OSVersion = osVersion
	r.OSArch = osArch
	r.SerialNumber = serialNumber
	r.MacAddress = macAddress
	return r
}

// WithLegacyHardware sets legacy hardware fields for backward compatibility
func (r *ServerDetailsUpdateRequest) WithLegacyHardware(cpuModel string, cpuCount, cpuCores int, memoryTotal, storageTotal uint64) *ServerDetailsUpdateRequest {
	r.CPUModel = cpuModel
	r.CPUCount = cpuCount
	r.CPUCores = cpuCores
	r.MemoryTotal = memoryTotal
	r.StorageTotal = storageTotal
	return r
}

// WithHardwareDetails sets detailed hardware information
func (r *ServerDetailsUpdateRequest) WithHardwareDetails(hardware *HardwareDetails) *ServerDetailsUpdateRequest {
	r.Hardware = hardware
	return r
}

// WithCPUs adds CPU information to hardware details
func (r *ServerDetailsUpdateRequest) WithCPUs(cpus []ServerCPUInfo) *ServerDetailsUpdateRequest {
	if r.Hardware == nil {
		r.Hardware = &HardwareDetails{}
	}
	r.Hardware.CPU = cpus
	return r
}

// WithMemory adds memory information to hardware details
func (r *ServerDetailsUpdateRequest) WithMemory(memory *ServerMemoryInfo) *ServerDetailsUpdateRequest {
	if r.Hardware == nil {
		r.Hardware = &HardwareDetails{}
	}
	r.Hardware.Memory = memory
	return r
}

// WithNetworkInterfaces adds network interface information to hardware details
func (r *ServerDetailsUpdateRequest) WithNetworkInterfaces(interfaces []ServerNetworkInterfaceInfo) *ServerDetailsUpdateRequest {
	if r.Hardware == nil {
		r.Hardware = &HardwareDetails{}
	}
	r.Hardware.Network = interfaces
	return r
}

// WithDisks adds disk information to hardware details
func (r *ServerDetailsUpdateRequest) WithDisks(disks []ServerDiskInfo) *ServerDetailsUpdateRequest {
	if r.Hardware == nil {
		r.Hardware = &HardwareDetails{}
	}
	r.Hardware.Disks = disks
	return r
}

// HasHardwareDetails returns true if the request contains enhanced hardware details
func (r *ServerDetailsUpdateRequest) HasHardwareDetails() bool {
	return r.Hardware != nil
}

// HasDisks returns true if the request contains disk information
func (r *ServerDetailsUpdateRequest) HasDisks() bool {
	return r.Hardware != nil && len(r.Hardware.Disks) > 0
}

// Standard capability constants
const (
	// Server capabilities
	CapabilityServersRead     = "servers:read"
	CapabilityServersWrite    = "servers:write"
	CapabilityServersRegister = "servers:register"
	CapabilityServersDelete   = "servers:delete"
	CapabilityServersAll      = "servers:*"

	// Monitoring capabilities
	CapabilityMonitoringRead    = "monitoring:read"
	CapabilityMonitoringWrite   = "monitoring:write"
	CapabilityMonitoringExecute = "monitoring:execute"
	CapabilityMonitoringAll     = "monitoring:*"

	// Probe capabilities
	CapabilityProbesRead    = "probes:read"
	CapabilityProbesWrite   = "probes:write"
	CapabilityProbesExecute = "probes:execute"
	CapabilityProbesAll     = "probes:*"

	// Metrics capabilities
	CapabilityMetricsRead   = "metrics:read"
	CapabilityMetricsWrite  = "metrics:write"
	CapabilityMetricsSubmit = "metrics:submit"
	CapabilityMetricsAll    = "metrics:*"

	// Organization capabilities
	CapabilityOrganizationRead  = "organization:read"
	CapabilityOrganizationWrite = "organization:write"
	CapabilityOrganizationAll   = "organization:*"

	// Admin capabilities
	CapabilityAdminRead  = "admin:read"
	CapabilityAdminWrite = "admin:write"
	CapabilityAdminAll   = "admin:*"

	// Wildcard capability (full access)
	CapabilityAll = "*"
)

// AgentVersion represents an agent version
type AgentVersion struct {
	GormModel
	Version             string                 `json:"version"`
	Environment         string                 `json:"environment,omitempty"`
	Platform            string                 `json:"platform"`
	Architectures       []string               `json:"architectures,omitempty"`
	DownloadURLs        map[string]string      `json:"download_urls,omitempty"`
	UpdaterURLs         map[string]string      `json:"updater_urls,omitempty"`
	ReleaseNotes        string                 `json:"release_notes,omitempty"`
	MinimumAPIVersion   string                 `json:"minimum_api_version,omitempty"`
	ReleaseDate         *CustomTime            `json:"release_date,omitempty"`
	IsStable            bool                   `json:"is_stable"`
	IsPrerelease        bool                   `json:"is_prerelease"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

// AgentVersionRequest represents a request to register a new agent version
type AgentVersionRequest struct {
	Version           string                 `json:"version"`
	Environment       string                 `json:"environment,omitempty"`
	Platform          string                 `json:"platform"`
	Architectures     []string               `json:"architectures,omitempty"`
	DownloadURLs      map[string]string      `json:"download_urls,omitempty"`
	UpdaterURLs       map[string]string      `json:"updater_urls,omitempty"`
	ReleaseNotes      string                 `json:"release_notes,omitempty"`
	MinimumAPIVersion string                 `json:"minimum_api_version,omitempty"`
	IsStable          *bool                  `json:"is_stable,omitempty"`
	IsPrerelease      *bool                  `json:"is_prerelease,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// AgentBinaryRequest represents a request to add a binary for an agent version
type AgentBinaryRequest struct {
	Platform     string `json:"platform"`
	Architecture string `json:"architecture"`
	DownloadURL  string `json:"download_url"`
	FileHash     string `json:"file_hash"`
}

// OrganizationUsageMetrics represents the current usage snapshot for billing purposes
// This stores the most recent usage metrics for each organization
type OrganizationUsageMetrics struct {
	ID               uint64          `json:"id"`
	OrganizationID   uint            `json:"organization_id"`
	ActiveAgentCount int             `json:"active_agent_count"`
	TotalAgentCount  int             `json:"total_agent_count"`
	FeaturesEnabled  interface{}     `json:"features_enabled"` // JSON object
	RetentionDays    int             `json:"retention_days"`
	StorageUsedBytes int64           `json:"storage_used_bytes"`
	StorageUsedGB    float64         `json:"storage_used_gb"`
	CollectedAt      *CustomTime     `json:"collected_at"`
	CreatedAt        *CustomTime     `json:"created_at"`
	UpdatedAt        *CustomTime     `json:"updated_at"`
}

// UsageMetricsHistory stores historical usage metrics for billing and analytics
// This is a TimescaleDB hypertable optimized for time-series data
type UsageMetricsHistory struct {
	ID               uint64      `json:"id"`
	OrganizationID   uint        `json:"organization_id"`
	ActiveAgentCount int         `json:"active_agent_count"`
	TotalAgentCount  int         `json:"total_agent_count"`
	FeaturesEnabled  interface{} `json:"features_enabled"` // JSON object
	RetentionDays    int         `json:"retention_days"`
	StorageUsedGB    float64     `json:"storage_used_gb"`
	CollectedAt      *CustomTime `json:"collected_at"`
	CreatedAt        *CustomTime `json:"created_at"`
}

// UsageSummary represents aggregated usage statistics over a time period
type UsageSummary struct {
	OrganizationID        uint            `json:"organization_id"`
	StartDate             *CustomTime     `json:"start_date"`
	EndDate               *CustomTime     `json:"end_date"`
	AverageAgentCount     float64         `json:"average_agent_count"`
	MaxAgentCount         int             `json:"max_agent_count"`
	AverageStorageGB      float64         `json:"average_storage_gb"`
	MaxStorageGB          float64         `json:"max_storage_gb"`
	FeaturesEnabled       map[string]bool `json:"features_enabled"`
	RetentionDays         int             `json:"retention_days"`
	TotalDataPoints       int             `json:"total_data_points"`
	BillingRecommendation string          `json:"billing_recommendation,omitempty"`
}

// OrganizationUsageOverview represents a summary of usage across all organizations (admin)
type OrganizationUsageOverview struct {
	TotalOrganizations int                          `json:"total_organizations"`
	TotalActiveAgents  int                          `json:"total_active_agents"`
	TotalStorageGB     float64                      `json:"total_storage_gb"`
	Organizations      []OrganizationUsageMetrics   `json:"organizations"`
}

// UsageMetricsRecordRequest represents the request body for recording usage metrics
// Used by org-management-controller to submit usage metrics to the API
type UsageMetricsRecordRequest struct {
	OrganizationID   uint            `json:"organization_id"`
	ActiveAgentCount int             `json:"active_agent_count"`
	TotalAgentCount  int             `json:"total_agent_count"`
	FeaturesEnabled  map[string]bool `json:"features_enabled"`
	RetentionDays    int             `json:"retention_days"`
	StorageUsedBytes int64           `json:"storage_used_bytes"`
	CollectedAt      time.Time       `json:"collected_at"`
}

// AgentCountsResponse represents agent count statistics for billing
type AgentCountsResponse struct {
	OrganizationID uint `json:"organization_id"`
	ActiveCount    int  `json:"active_count"`
	TotalCount     int  `json:"total_count"`
}

// StorageUsageResponse represents storage usage statistics for billing
type StorageUsageResponse struct {
	OrganizationID uint    `json:"organization_id"`
	StorageBytes   int64   `json:"storage_bytes"`
	StorageGB      float64 `json:"storage_gb"`
}

// ============================================================================
// Tag Management
// ============================================================================

// Tag represents a tag in the system
type Tag struct {
	ID             uint       `json:"id"`
	OrganizationID uint       `json:"organization_id"`
	Namespace      string     `json:"namespace"`
	Key            string     `json:"key"`
	Value          string     `json:"value"`
	Source         string     `json:"source"`
	Description    string     `json:"description,omitempty"`
	CreatedByID    *uint      `json:"created_by_id,omitempty"`
	CreatedByEmail string     `json:"created_by_email,omitempty"`
	CreatedAt      CustomTime `json:"created_at"`
	UpdatedAt      CustomTime `json:"updated_at"`
	ServerCount    int64      `json:"server_count"`
}

// TagCreateRequest represents the request structure for creating a tag
type TagCreateRequest struct {
	Namespace   string `json:"namespace"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description,omitempty"`
}

// TagAssignRequest represents the request structure for assigning tags to a server
type TagAssignRequest struct {
	TagIDs []uint `json:"tag_ids"`
}

// TagAssignmentResult represents the result of a tag assignment operation
type TagAssignmentResult struct {
	Assigned        int `json:"assigned"`
	AlreadyAssigned int `json:"already_assigned"`
	Total           int `json:"total"`
}

// ServerTag represents a tag assigned to a server
type ServerTag struct {
	ID              uint       `json:"id"`
	TagID           uint       `json:"tag_id"`
	Namespace       string     `json:"namespace"`
	Key             string     `json:"key"`
	Value           string     `json:"value"`
	Source          string     `json:"source"`
	Description     string     `json:"description,omitempty"`
	AssignedAt      CustomTime `json:"assigned_at"`
	AssignedByEmail string     `json:"assigned_by_email,omitempty"`
	ConfidenceScore *float64   `json:"confidence_score,omitempty"`
}

// TagListOptions represents filtering and pagination options for listing tags
type TagListOptions struct {
	Namespace string // Filter by namespace
	Source    string // Filter by source (automatic, manual, inherited)
	Key       string // Filter by key pattern (partial match)
	Page      int    // Page number (default: 1)
	Limit     int    // Items per page (default: 50)
}

// ToQuery converts TagListOptions to a query parameter map
func (o *TagListOptions) ToQuery() map[string]string {
	query := make(map[string]string)

	if o.Namespace != "" {
		query["namespace"] = o.Namespace
	}
	if o.Source != "" {
		query["source"] = o.Source
	}
	if o.Key != "" {
		query["key"] = o.Key
	}
	if o.Page > 0 {
		query["page"] = fmt.Sprintf("%d", o.Page)
	}
	if o.Limit > 0 {
		query["limit"] = fmt.Sprintf("%d", o.Limit)
	}

	return query
}

// TagNamespace represents a tag namespace definition
type TagNamespace struct {
	ID               uint       `json:"id"`
	Namespace        string     `json:"namespace"`
	ParentNamespace  string     `json:"parent_namespace,omitempty"`
	Type             string     `json:"type"`
	Description      string     `json:"description"`
	KeyPattern       string     `json:"key_pattern,omitempty"`
	ValuePattern     string     `json:"value_pattern,omitempty"`
	AllowedValues    []string   `json:"allowed_values,omitempty"`
	RequiresApproval bool       `json:"requires_approval"`
	IsActive         bool       `json:"is_active"`
	CreatedByID      *uint      `json:"created_by_id,omitempty"`
	CreatedByEmail   string     `json:"created_by_email,omitempty"`
	CreatedAt        CustomTime `json:"created_at"`
	UpdatedAt        CustomTime `json:"updated_at"`
}

// TagNamespaceCreateRequest represents the request structure for creating a namespace
type TagNamespaceCreateRequest struct {
	Namespace        string   `json:"namespace"`
	ParentNamespace  string   `json:"parent_namespace,omitempty"`
	Type             string   `json:"type,omitempty"`
	Description      string   `json:"description,omitempty"`
	KeyPattern       string   `json:"key_pattern,omitempty"`
	ValuePattern     string   `json:"value_pattern,omitempty"`
	AllowedValues    []string `json:"allowed_values,omitempty"`
	RequiresApproval bool     `json:"requires_approval,omitempty"`
}

// TagNamespacePermissionRequest represents the request structure for setting namespace permissions
type TagNamespacePermissionRequest struct {
	UserID     *uint  `json:"user_id,omitempty"`
	RoleName   string `json:"role_name,omitempty"`
	CanCreate  bool   `json:"can_create"`
	CanRead    bool   `json:"can_read"`
	CanUpdate  bool   `json:"can_update"`
	CanDelete  bool   `json:"can_delete"`
	CanApprove bool   `json:"can_approve"`
}

// TagNamespaceListOptions represents filtering options for listing namespaces
type TagNamespaceListOptions struct {
	Type      string // Filter by namespace type
	Parent    string // Filter by parent namespace
	Active    *bool  // Filter by active status (nil = all, true = active only, false = inactive only)
	Search    string // Search in namespace name and description
	Hierarchy bool   // Return hierarchical structure
}

// ToQuery converts TagNamespaceListOptions to a query parameter map
func (o *TagNamespaceListOptions) ToQuery() map[string]string {
	query := make(map[string]string)

	if o.Type != "" {
		query["type"] = o.Type
	}
	if o.Parent != "" {
		query["parent"] = o.Parent
	}
	if o.Active != nil {
		if *o.Active {
			query["active"] = "true"
		} else {
			query["active"] = "false"
		}
	}
	if o.Search != "" {
		query["search"] = o.Search
	}
	if o.Hierarchy {
		query["hierarchy"] = "true"
	}

	return query
}

// ============================================================================
// Tag Inheritance Models
// ============================================================================

// Type aliases for inheritance enums
type InheritanceSource string
type InheritanceTarget string

// TagInheritanceRule represents an inheritance rule for automatic tag propagation
type TagInheritanceRule struct {
	ID             uint              `json:"id"`
	OrganizationID uint              `json:"organization_id"`
	Name           string            `json:"name"`
	Description    string            `json:"description,omitempty"`
	SourceType     InheritanceSource `json:"source_type"`
	TargetType     InheritanceTarget `json:"target_type"`
	Namespace      string            `json:"namespace,omitempty"`
	KeyPattern     string            `json:"key_pattern,omitempty"`
	ValuePattern   string            `json:"value_pattern,omitempty"`
	Conditions     string            `json:"conditions,omitempty"`
	Enabled        bool              `json:"enabled"`
	Priority       int               `json:"priority"`
	CreatedBy      *UserInfo         `json:"created_by,omitempty"`
	LastRunAt      *string           `json:"last_run_at,omitempty"`
	LastRunStatus  string            `json:"last_run_status,omitempty"`
	ProcessedCount int               `json:"processed_count"`
	CreatedAt      CustomTime        `json:"created_at"`
	UpdatedAt      CustomTime        `json:"updated_at"`
}

// TagInheritanceRuleCreateRequest represents a request to create an inheritance rule
type TagInheritanceRuleCreateRequest struct {
	Name         string            `json:"name"`
	Description  string            `json:"description,omitempty"`
	SourceType   InheritanceSource `json:"source_type"`
	TargetType   InheritanceTarget `json:"target_type"`
	Namespace    string            `json:"namespace,omitempty"`
	KeyPattern   string            `json:"key_pattern,omitempty"`
	ValuePattern string            `json:"value_pattern,omitempty"`
	Conditions   string            `json:"conditions,omitempty"`
	Enabled      bool              `json:"enabled"`
	Priority     int               `json:"priority"`
}

// OrganizationTag represents a tag set at the organization level
type OrganizationTag struct {
	ID             uint       `json:"id"`
	OrganizationID uint       `json:"organization_id"`
	Tag            TagInfo    `json:"tag"`
	InheritToAll   bool       `json:"inherit_to_all"`
	InheritRules   string     `json:"inherit_rules,omitempty"`
	CreatedBy      *UserInfo  `json:"created_by,omitempty"`
	CreatedAt      CustomTime `json:"created_at"`
	UpdatedAt      CustomTime `json:"updated_at"`
}

// OrganizationTagRequest represents a request to set an organization tag
type OrganizationTagRequest struct {
	TagID        uint   `json:"tag_id"`
	InheritToAll bool   `json:"inherit_to_all"`
	InheritRules string `json:"inherit_rules,omitempty"`
}

// ServerParentRelationship represents a parent-child relationship between servers
type ServerParentRelationship struct {
	ID             uint       `json:"id"`
	OrganizationID uint       `json:"organization_id"`
	ParentServer   ServerInfo `json:"parent_server"`
	ChildServer    ServerInfo `json:"child_server"`
	RelationType   string     `json:"relation_type"`
	InheritTags    bool       `json:"inherit_tags"`
	CreatedBy      *UserInfo  `json:"created_by,omitempty"`
	CreatedAt      CustomTime `json:"created_at"`
	UpdatedAt      CustomTime `json:"updated_at"`
}

// ServerRelationshipRequest represents a request to create a server relationship
type ServerRelationshipRequest struct {
	ParentServerID string `json:"parent_server_id"`
	ChildServerID  string `json:"child_server_id"`
	RelationType   string `json:"relation_type"`
	InheritTags    bool   `json:"inherit_tags"`
}

// TagInfo represents basic tag information
type TagInfo struct {
	ID          uint   `json:"id"`
	Namespace   string `json:"namespace"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description,omitempty"`
}

// UserInfo represents basic user information
type UserInfo struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
}

// ServerInfo represents basic server information
type ServerInfo struct {
	ID         uint   `json:"id"`
	ServerUUID string `json:"server_uuid"`
	Name       string `json:"name"`
}

// OrganizationTagListOptions provides filtering options for organization tags
type OrganizationTagListOptions struct {
	InheritOnly bool
}

func (o *OrganizationTagListOptions) ToQuery() map[string]string {
	query := make(map[string]string)
	if o.InheritOnly {
		query["inherit_only"] = "true"
	}
	return query
}

// ServerRelationshipListOptions provides filtering options for server relationships
type ServerRelationshipListOptions struct {
	ServerID     string
	RelationType string
	InheritOnly  bool
}

func (o *ServerRelationshipListOptions) ToQuery() map[string]string {
	query := make(map[string]string)
	if o.ServerID != "" {
		query["server_id"] = o.ServerID
	}
	if o.RelationType != "" {
		query["relation_type"] = o.RelationType
	}
	if o.InheritOnly {
		query["inherit_only"] = "true"
	}
	return query
}

// ============================================================================
// Tag History Models
// ============================================================================

// TagHistoryResponse represents a single tag history entry
type TagHistoryResponse struct {
	ID            string                `json:"id"`
	Action        string                `json:"action"`
	Tag           TagHistoryTag         `json:"tag"`
	PreviousValue *string               `json:"previous_value,omitempty"`
	Timestamp     string                `json:"timestamp"`
	User          *TagHistoryUser       `json:"user,omitempty"`
	Metadata      TagHistoryMetadataRes `json:"metadata"`
}

// TagHistoryTag represents tag information in history response
type TagHistoryTag struct {
	ID        uint    `json:"id,omitempty"`
	Key       string  `json:"key"`
	Value     string  `json:"value"`
	Namespace *string `json:"namespace,omitempty"`
}

// TagHistoryUser represents user information in history response
type TagHistoryUser struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// TagHistoryMetadataRes represents metadata in history response
type TagHistoryMetadataRes struct {
	Source     string   `json:"source"`
	Reason     *string  `json:"reason,omitempty"`
	Confidence *float64 `json:"confidence,omitempty"`
}

// TagHistorySummary represents aggregated statistics for tag history
type TagHistorySummary struct {
	TotalChanges       int                 `json:"total_changes"`
	ChangesByAction    map[string]int      `json:"changes_by_action"`
	ChangesByNamespace map[string]int      `json:"changes_by_namespace"`
	MostActiveUsers    []ActiveUser        `json:"most_active_users"`
	RecentActivity     RecentActivityStats `json:"recent_activity"`
}

// ActiveUser represents a user with their change count
type ActiveUser struct {
	UserID      uint   `json:"user_id"`
	Name        string `json:"name"`
	ChangeCount int    `json:"change_count"`
}

// RecentActivityStats represents recent activity statistics
type RecentActivityStats struct {
	Last24h int `json:"last_24h"`
	Last7d  int `json:"last_7d"`
	Last30d int `json:"last_30d"`
}

// TagHistoryQueryParams represents query parameters for filtering tag history
type TagHistoryQueryParams struct {
	Action    string `json:"action,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Source    string `json:"source,omitempty"`
	TagID     uint   `json:"tag_id,omitempty"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
	Page      int    `json:"page,omitempty"`
	Limit     int    `json:"limit,omitempty"`
}

func (o *TagHistoryQueryParams) ToQuery() map[string]string {
	query := make(map[string]string)
	if o.Action != "" {
		query["action"] = o.Action
	}
	if o.Namespace != "" {
		query["namespace"] = o.Namespace
	}
	if o.Source != "" {
		query["source"] = o.Source
	}
	if o.TagID > 0 {
		query["tag_id"] = fmt.Sprintf("%d", o.TagID)
	}
	if o.StartDate != "" {
		query["start_date"] = o.StartDate
	}
	if o.EndDate != "" {
		query["end_date"] = o.EndDate
	}
	if o.Page > 0 {
		query["page"] = fmt.Sprintf("%d", o.Page)
	}
	if o.Limit > 0 {
		query["limit"] = fmt.Sprintf("%d", o.Limit)
	}
	return query
}

// ============================================================================
// Bulk Tag Operation Models
// ============================================================================

// BulkTagCreateRequest represents a request to create multiple tags
type BulkTagCreateRequest struct {
	Tags []BulkTagCreateItem `json:"tags"`
}

// BulkTagCreateItem represents a single tag in bulk create
type BulkTagCreateItem struct {
	Namespace   string `json:"namespace"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description,omitempty"`
}

// BulkTagCreateResult represents the result of bulk tag creation
type BulkTagCreateResult struct {
	Created      []Tag    `json:"created"`
	Skipped      []string `json:"skipped"`
	CreatedCount int      `json:"created_count"`
	SkippedCount int      `json:"skipped_count"`
}

// BulkTagAssignRequest represents a request to assign tags to multiple servers
type BulkTagAssignRequest struct {
	ServerIDs []string `json:"server_ids"`
	TagIDs    []uint   `json:"tag_ids"`
}

// BulkTagAssignResult represents the result of bulk tag assignment
type BulkTagAssignResult struct {
	Assigned int `json:"assigned"`
	Skipped  int `json:"skipped"`
	Total    int `json:"total"`
}

// BulkGroupAssignRequest represents a request to assign servers to multiple groups
type BulkGroupAssignRequest struct {
	ServerIDs []string `json:"server_ids"`
	GroupIDs  []uint   `json:"group_ids"`
}

// BulkGroupAssignResult represents the result of bulk group assignment
type BulkGroupAssignResult struct {
	Assigned int `json:"assigned"`
	Skipped  int `json:"skipped"`
	Total    int `json:"total"`
}

// ============================================================================
// Tag Detection Rule Models
// ============================================================================

// TagDetectionRule represents a rule for automatic tag assignment
type TagDetectionRule struct {
	ID             uint            `json:"id"`
	OrganizationID uint            `json:"organization_id"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	Namespace      string          `json:"namespace"`
	TagKey         string          `json:"tag_key"`
	TagValue       string          `json:"tag_value"`
	Conditions     json.RawMessage `json:"conditions"`
	Priority       int             `json:"priority"`
	Confidence     float64         `json:"confidence"`
	Enabled        bool            `json:"enabled"`
	CreatedByID    *uint           `json:"created_by_id,omitempty"`
	CreatedByEmail string          `json:"created_by_email,omitempty"`
	CreatedAt      string          `json:"created_at"`
	UpdatedAt      string          `json:"updated_at"`
}

// TagDetectionRuleListOptions provides filtering options for listing tag detection rules
type TagDetectionRuleListOptions struct {
	Enabled   *bool
	Namespace string
	Page      int
	Limit     int
}

func (o *TagDetectionRuleListOptions) ToQuery() map[string]string {
	query := make(map[string]string)
	if o.Enabled != nil {
		if *o.Enabled {
			query["enabled"] = "true"
		} else {
			query["enabled"] = "false"
		}
	}
	if o.Namespace != "" {
		query["namespace"] = o.Namespace
	}
	if o.Page > 0 {
		query["page"] = fmt.Sprintf("%d", o.Page)
	}
	if o.Limit > 0 {
		query["limit"] = fmt.Sprintf("%d", o.Limit)
	}
	return query
}

// DefaultRulesCreateResult represents the result of creating default rules
type DefaultRulesCreateResult struct {
	CreatedCount int `json:"created_count"`
}

// EvaluateRulesRequest represents a request to evaluate tag detection rules
type EvaluateRulesRequest struct {
	ServerIDs  []string `json:"server_ids,omitempty"`
	AllServers bool     `json:"all_servers,omitempty"`
}

// EvaluateRulesResult represents the result of rule evaluation
type EvaluateRulesResult struct {
	ProcessingCount int `json:"processing_count"`
}

// ============================================================================
// Analytics Models
// ============================================================================

// AI Analytics Models

// AICapabilities represents available AI analytics features
type AICapabilities struct {
	AnomalyDetection    bool     `json:"anomaly_detection"`
	PredictiveAnalytics bool     `json:"predictive_analytics"`
	RootCauseAnalysis   bool     `json:"root_cause_analysis"`
	CapacityPlanning    bool     `json:"capacity_planning"`
	AvailableModels     []string `json:"available_models"`
	Status              string   `json:"status"` // "operational", "degraded", "unavailable"
}

// AIAnalysisRequest represents a request for AI-powered metric analysis
type AIAnalysisRequest struct {
	ServerUUIDs  []string               `json:"server_uuids,omitempty"`
	MetricTypes  []string               `json:"metric_types,omitempty"` // ["cpu", "memory", "disk", "network"]
	TimeRange    TimeRange              `json:"time_range"`
	AnalysisType string                 `json:"analysis_type"` // "anomaly", "prediction", "root_cause", "capacity"
	Context      map[string]interface{} `json:"context,omitempty"`
}

// AIAnalysisResult represents the result of AI-powered analysis
type AIAnalysisResult struct {
	AnalysisID  string                 `json:"analysis_id"`
	Timestamp   CustomTime             `json:"timestamp"`
	Insights    []AIInsight            `json:"insights"`
	Anomalies   []AIAnomaly            `json:"anomalies,omitempty"`
	Predictions []AIPrediction         `json:"predictions,omitempty"`
	Confidence  float64                `json:"confidence"` // 0.0 to 1.0
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// AIInsight represents an AI-generated insight
type AIInsight struct {
	Type        string   `json:"type"` // "recommendation", "warning", "info"
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Severity    string   `json:"severity"` // "low", "medium", "high", "critical"
	Confidence  float64  `json:"confidence"`
	Actions     []string `json:"actions,omitempty"`
}

// AIAnomaly represents a detected anomaly
type AIAnomaly struct {
	MetricType    string     `json:"metric_type"`
	ServerUUID    string     `json:"server_uuid"`
	DetectedAt    CustomTime `json:"detected_at"`
	Severity      string     `json:"severity"`
	ExpectedValue float64    `json:"expected_value"`
	ActualValue   float64    `json:"actual_value"`
	Deviation     float64    `json:"deviation"`
	Description   string     `json:"description"`
}

// AIPrediction represents a prediction for future metrics
type AIPrediction struct {
	MetricType  string     `json:"metric_type"`
	ServerUUID  string     `json:"server_uuid,omitempty"`
	PredictedAt CustomTime `json:"predicted_at"`
	Value       float64    `json:"value"`
	Confidence  float64    `json:"confidence"`
	UpperBound  float64    `json:"upper_bound,omitempty"`
	LowerBound  float64    `json:"lower_bound,omitempty"`
}

// AIServiceStatus represents the health status of AI services
type AIServiceStatus struct {
	Status          string     `json:"status"` // "operational", "degraded", "unavailable"
	LastCheck       CustomTime `json:"last_check"`
	ModelsAvailable int        `json:"models_available"`
	QueueLength     int        `json:"queue_length"`
	AverageLatency  float64    `json:"average_latency_ms"`
	Uptime          float64    `json:"uptime_percentage"`
}

// Hardware Analytics Models

// HardwareTrends represents historical hardware trends for a server
type HardwareTrends struct {
	ServerUUID   string                 `json:"server_uuid"`
	StartTime    CustomTime             `json:"start_time"`
	EndTime      CustomTime             `json:"end_time"`
	CPUTrend     MetricTrendData        `json:"cpu_trend,omitempty"`
	MemoryTrend  MetricTrendData        `json:"memory_trend,omitempty"`
	DiskTrend    MetricTrendData        `json:"disk_trend,omitempty"`
	NetworkTrend MetricTrendData        `json:"network_trend,omitempty"`
	Summary      map[string]interface{} `json:"summary,omitempty"`
}

// MetricTrendData represents trend data for a specific metric
type MetricTrendData struct {
	Average    float64      `json:"average"`
	Minimum    float64      `json:"minimum"`
	Maximum    float64      `json:"maximum"`
	TrendLine  []TrendPoint `json:"trend_line"`
	Growth     float64      `json:"growth_percentage"`
	Volatility float64      `json:"volatility"`
}

// TrendPoint represents a single point in a trend line
type TrendPoint struct {
	Timestamp CustomTime `json:"timestamp"`
	Value     float64    `json:"value"`
}

// HardwareHealth represents current hardware health score and diagnostics
type HardwareHealth struct {
	ServerUUID      string             `json:"server_uuid"`
	OverallScore    int                `json:"overall_score"` // 0-100
	LastCheck       CustomTime         `json:"last_check"`
	ComponentScores ComponentHealthMap `json:"component_scores"`
	Issues          []HealthIssue      `json:"issues,omitempty"`
	Recommendations []string           `json:"recommendations,omitempty"`
}

// ComponentHealthMap represents health scores for individual components
type ComponentHealthMap struct {
	CPU     int `json:"cpu"`
	Memory  int `json:"memory"`
	Disk    int `json:"disk"`
	Network int `json:"network"`
}

// HealthIssue represents a detected health issue
type HealthIssue struct {
	Component   string `json:"component"`
	Severity    string `json:"severity"` // "info", "warning", "critical"
	Description string `json:"description"`
	Impact      string `json:"impact,omitempty"`
}

// HardwarePrediction represents predictive analytics for hardware failures
type HardwarePrediction struct {
	ServerUUID           string                `json:"server_uuid"`
	PredictionHorizon    int                   `json:"prediction_horizon_days"`
	FailureProbability   float64               `json:"failure_probability"`
	ComponentPredictions []ComponentPrediction `json:"component_predictions"`
	RecommendedActions   []string              `json:"recommended_actions"`
	ConfidenceLevel      float64               `json:"confidence_level"`
}

// ComponentPrediction represents failure prediction for a specific component
type ComponentPrediction struct {
	Component          string     `json:"component"`
	FailureProbability float64    `json:"failure_probability"`
	EstimatedFailure   CustomTime `json:"estimated_failure,omitempty"`
	WarningLevel       string     `json:"warning_level"` // "none", "low", "medium", "high"
}

// Fleet Analytics Models

// FleetOverview represents organization-wide fleet statistics
type FleetOverview struct {
	TotalServers        int                 `json:"total_servers"`
	ActiveServers       int                 `json:"active_servers"`
	InactiveServers     int                 `json:"inactive_servers"`
	HealthDistribution  HealthDistribution  `json:"health_distribution"`
	ResourceUtilization ResourceUtilization `json:"resource_utilization"`
	TopIssues           []FleetIssue        `json:"top_issues,omitempty"`
	LastUpdated         CustomTime          `json:"last_updated"`
}

// HealthDistribution represents distribution of server health scores
type HealthDistribution struct {
	Healthy  int `json:"healthy"`  // Score >= 80
	Warning  int `json:"warning"`  // Score 50-79
	Critical int `json:"critical"` // Score < 50
	Unknown  int `json:"unknown"`  // No data
}

// ResourceUtilization represents aggregate resource utilization
type ResourceUtilization struct {
	AverageCPU    float64 `json:"average_cpu"`
	AverageMemory float64 `json:"average_memory"`
	AverageDisk   float64 `json:"average_disk"`
	TotalStorage  float64 `json:"total_storage_gb"`
}

// FleetIssue represents a common issue across the fleet
type FleetIssue struct {
	Type            string `json:"type"`
	Description     string `json:"description"`
	Severity        string `json:"severity"`
	AffectedServers int    `json:"affected_servers"`
}

// OrganizationDashboard represents comprehensive dashboard data
type OrganizationDashboard struct {
	FleetOverview            FleetOverview     `json:"fleet_overview"`
	RecentAlerts             []DashboardAlert  `json:"recent_alerts"`
	TrendingMetrics          []TrendingMetric  `json:"trending_metrics"`
	CapacityForecasts        []CapacityForecast `json:"capacity_forecasts,omitempty"`
	TopPerformingServers     []ServerSummary   `json:"top_performing_servers,omitempty"`
	BottomPerformingServers  []ServerSummary   `json:"bottom_performing_servers,omitempty"`
	LastUpdated              CustomTime        `json:"last_updated"`
}

// DashboardAlert represents an alert summary for the dashboard
type DashboardAlert struct {
	AlertID    uint       `json:"alert_id"`
	Severity   string     `json:"severity"`
	Title      string     `json:"title"`
	ServerUUID string     `json:"server_uuid,omitempty"`
	ServerName string     `json:"server_name,omitempty"`
	Timestamp  CustomTime `json:"timestamp"`
}

// TrendingMetric represents a metric with trend information
type TrendingMetric struct {
	MetricType string  `json:"metric_type"`
	Value      float64 `json:"value"`
	Change     float64 `json:"change_percentage"`
	Trend      string  `json:"trend"` // "up", "down", "stable"
}

// CapacityForecast represents capacity planning forecast
type CapacityForecast struct {
	ResourceType         string     `json:"resource_type"` // "cpu", "memory", "storage"
	CurrentUtilization   float64    `json:"current_utilization"`
	ForecastedDate       CustomTime `json:"forecasted_exhaustion_date,omitempty"`
	DaysUntilExhaustion  int        `json:"days_until_exhaustion"`
	Recommendation       string     `json:"recommendation"`
}

// ServerSummary represents a summary of server performance
type ServerSummary struct {
	UUID        string  `json:"uuid"`
	Hostname    string  `json:"hostname"`
	HealthScore int     `json:"health_score"`
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
}

// Advanced Analytics Models

// CorrelationAnalysisRequest represents a request for correlation analysis
type CorrelationAnalysisRequest struct {
	ServerUUIDs []string  `json:"server_uuids,omitempty"`
	MetricTypes []string  `json:"metric_types"` // Metrics to correlate
	TimeRange   TimeRange `json:"time_range"`
	Method      string    `json:"method,omitempty"` // "pearson", "spearman", "kendall"
}

// CorrelationResult represents the result of correlation analysis
type CorrelationResult struct {
	Correlations     []MetricCorrelation `json:"correlations"`
	Matrix           [][]float64         `json:"matrix,omitempty"` // Correlation matrix
	MetricLabels     []string            `json:"metric_labels"`
	SignificantPairs []CorrelationPair   `json:"significant_pairs,omitempty"`
	AnalyzedAt       CustomTime          `json:"analyzed_at"`
}

// MetricCorrelation represents correlation between two metrics
type MetricCorrelation struct {
	Metric1     string  `json:"metric1"`
	Metric2     string  `json:"metric2"`
	Coefficient float64 `json:"coefficient"` // -1.0 to 1.0
	Strength    string  `json:"strength"`    // "weak", "moderate", "strong"
	PValue      float64 `json:"p_value,omitempty"`
}

// CorrelationPair represents a significant correlation pair
type CorrelationPair struct {
	Metrics      []string `json:"metrics"`
	Relationship string   `json:"relationship"` // "positive", "negative"
	Insight      string   `json:"insight,omitempty"`
}

// DependencyGraph represents infrastructure dependency relationships
type DependencyGraph struct {
	Nodes         []DependencyNode `json:"nodes"`
	Edges         []DependencyEdge `json:"edges"`
	CriticalPaths [][]string       `json:"critical_paths,omitempty"`
	GeneratedAt   CustomTime       `json:"generated_at"`
}

// DependencyNode represents a node in the dependency graph
type DependencyNode struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"` // "server", "service", "database", "load_balancer"
	Name     string                 `json:"name"`
	Status   string                 `json:"status"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// DependencyEdge represents an edge between nodes
type DependencyEdge struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Type   string `json:"type"`          // "depends_on", "provides_to", "connected_to"
	Weight int    `json:"weight,omitempty"` // Importance/traffic weight
}

// ============================================================================
// Machine Learning Models
// ============================================================================

// TagSuggestion represents an ML-generated tag suggestion for a server
type TagSuggestion struct {
	ID           uint                   `json:"id"`
	ServerID     uint                   `json:"server_id"`
	ServerUUID   string                 `json:"server_uuid"`
	PredictionID string                 `json:"prediction_id"`
	TagKey       string                 `json:"tag_key"`
	TagValue     string                 `json:"tag_value"`
	Namespace    string                 `json:"namespace,omitempty"`
	Confidence   float64                `json:"confidence"`    // 0.0 to 1.0
	Reason       string                 `json:"reason"`        // Explanation for suggestion
	Applied      bool                   `json:"applied"`       // Whether suggestion was applied
	Rejected     bool                   `json:"rejected"`      // Whether suggestion was rejected
	Feedback     string                 `json:"feedback,omitempty"` // User feedback if rejected
	CreatedAt    CustomTime             `json:"created_at"`
	UpdatedAt    CustomTime             `json:"updated_at"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"` // Additional context
}

// TagPrediction represents an ML prediction for tag assignment
type TagPrediction struct {
	TagKey     string                 `json:"tag_key"`
	TagValue   string                 `json:"tag_value"`
	Confidence float64                `json:"confidence"`
	Reasoning  string                 `json:"reasoning"`
	Features   map[string]interface{} `json:"features,omitempty"` // Features used for prediction
}

// GroupSuggestion represents an ML-generated server grouping suggestion
type GroupSuggestion struct {
	ID              uint       `json:"id"`
	OrganizationID  uint       `json:"organization_id"`
	GroupName       string     `json:"group_name"`
	Description     string     `json:"description"`
	ServerIDs       []uint     `json:"server_ids"`
	ServerUUIDs     []string   `json:"server_uuids"`
	Confidence      float64    `json:"confidence"`      // 0.0 to 1.0
	Reason          string     `json:"reason"`          // Why these servers should be grouped
	Criteria        []string   `json:"criteria"`        // Criteria used (location, tags, specs, etc.)
	Accepted        bool       `json:"accepted"`        // Whether suggestion was accepted
	Rejected        bool       `json:"rejected"`        // Whether suggestion was rejected
	CreatedGroupID  *uint      `json:"created_group_id,omitempty"` // ID of created group if accepted
	CreatedAt       CustomTime `json:"created_at"`
	UpdatedAt       CustomTime `json:"updated_at"`
	EstimatedBenefit string    `json:"estimated_benefit,omitempty"` // Estimated benefits of grouping
}

// MLModel represents a machine learning model configuration and status
type MLModel struct {
	ID              uint       `json:"id"`
	Name            string     `json:"name"`
	ModelType       string     `json:"model_type"`       // "tag_prediction", "group_suggestion", etc.
	Version         string     `json:"version"`
	Status          string     `json:"status"`           // "active", "inactive", "training", "deprecated"
	Enabled         bool       `json:"enabled"`
	Accuracy        float64    `json:"accuracy,omitempty"` // Model accuracy (0.0 to 1.0)
	Precision       float64    `json:"precision,omitempty"`
	Recall          float64    `json:"recall,omitempty"`
	F1Score         float64    `json:"f1_score,omitempty"`
	TrainedAt       *CustomTime `json:"trained_at,omitempty"`
	TrainingDataSize int        `json:"training_data_size,omitempty"`
	CreatedAt       CustomTime `json:"created_at"`
	UpdatedAt       CustomTime `json:"updated_at"`
	Description     string     `json:"description,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// ModelPerformance represents detailed performance metrics for an ML model
type ModelPerformance struct {
	ModelID          uint                   `json:"model_id,omitempty"`
	ModelType        string                 `json:"model_type,omitempty"`
	Accuracy         float64                `json:"accuracy"`
	Precision        float64                `json:"precision"`
	Recall           float64                `json:"recall"`
	F1Score          float64                `json:"f1_score"`
	ConfusionMatrix  map[string]int         `json:"confusion_matrix,omitempty"`
	PredictionsCount int                    `json:"predictions_count,omitempty"`
	CorrectCount     int                    `json:"correct_count,omitempty"`
	IncorrectCount   int                    `json:"incorrect_count,omitempty"`
	AverageConfidence float64               `json:"average_confidence,omitempty"`
	EvaluatedAt      *CustomTime            `json:"evaluated_at,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// TrainingJob represents a machine learning model training job
type TrainingJob struct {
	ID              uint       `json:"id"`
	ModelID         uint       `json:"model_id"`
	ModelType       string     `json:"model_type"`
	Status          string     `json:"status"`         // "pending", "running", "completed", "failed"
	Progress        int        `json:"progress"`       // 0-100
	TrainingDataSize int       `json:"training_data_size,omitempty"`
	ValidationDataSize int     `json:"validation_data_size,omitempty"`
	StartedAt       *CustomTime `json:"started_at,omitempty"`
	CompletedAt     *CustomTime `json:"completed_at,omitempty"`
	Duration        int        `json:"duration,omitempty"`  // Duration in seconds
	ErrorMessage    string     `json:"error_message,omitempty"`
	Metrics         *ModelPerformance `json:"metrics,omitempty"` // Performance metrics after training
	CreatedAt       CustomTime `json:"created_at"`
	UpdatedAt       CustomTime `json:"updated_at"`
	Parameters      map[string]interface{} `json:"parameters,omitempty"` // Training parameters
	Logs            []string   `json:"logs,omitempty"` // Training logs
}

// TrainingJobStatus represents the current status of a training job
type TrainingJobStatus struct {
	JobID       uint       `json:"job_id"`
	Status      string     `json:"status"`
	Progress    int        `json:"progress"`
	Message     string     `json:"message,omitempty"`
	UpdatedAt   CustomTime `json:"updated_at"`
}

// ============================================================================
// Virtual Machine Models
// ============================================================================

// VirtualMachine represents a virtual machine instance
type VirtualMachine struct {
	ID             uint       `json:"id"`
	OrganizationID uint       `json:"organization_id"`
	Name           string     `json:"name"`
	Description    string     `json:"description,omitempty"`
	Status         string     `json:"status"` // "running", "stopped", "paused", "error"

	// Resource specifications
	CPUCores   int    `json:"cpu_cores"`
	MemoryMB   int    `json:"memory_mb"`
	StorageGB  int    `json:"storage_gb"`

	// Network configuration
	IPAddress      string   `json:"ip_address,omitempty"`
	MACAddress     string   `json:"mac_address,omitempty"`

	// Host information
	HostServerID   *uint  `json:"host_server_id,omitempty"`
	HostServerUUID string `json:"host_server_uuid,omitempty"`

	// Operating system
	OSType    string `json:"os_type,omitempty"`    // "linux", "windows", "other"
	OSVersion string `json:"os_version,omitempty"`

	// Lifecycle timestamps
	CreatedAt   CustomTime  `json:"created_at"`
	UpdatedAt   CustomTime  `json:"updated_at"`
	StartedAt   *CustomTime `json:"started_at,omitempty"`
	StoppedAt   *CustomTime `json:"stopped_at,omitempty"`

	// Metadata
	Tags     []string               `json:"tags,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// VMConfiguration represents virtual machine configuration for creation/updates
type VMConfiguration struct {
	Name           string                 `json:"name"`
	Description    string                 `json:"description,omitempty"`

	// Required resources
	CPUCores   int `json:"cpu_cores"`
	MemoryMB   int `json:"memory_mb"`
	StorageGB  int `json:"storage_gb"`

	// Optional configuration
	OSType         string   `json:"os_type,omitempty"`
	OSVersion      string   `json:"os_version,omitempty"`
	HostServerID   *uint    `json:"host_server_id,omitempty"`
	HostServerUUID string   `json:"host_server_uuid,omitempty"`

	// Network settings
	NetworkConfig map[string]interface{} `json:"network_config,omitempty"`

	// Additional settings
	Tags     []string               `json:"tags,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// VMStatus represents the current status of a virtual machine
type VMStatus struct {
	ID        uint       `json:"id"`
	VMID      uint       `json:"vm_id"`
	Status    string     `json:"status"`    // "running", "stopped", "paused", "error"
	Health    string     `json:"health"`    // "healthy", "degraded", "unhealthy"

	// Resource usage
	CPUUsagePercent    float64 `json:"cpu_usage_percent,omitempty"`
	MemoryUsageMB      int     `json:"memory_usage_mb,omitempty"`
	MemoryUsagePercent float64 `json:"memory_usage_percent,omitempty"`
	DiskUsageGB        int     `json:"disk_usage_gb,omitempty"`
	DiskUsagePercent   float64 `json:"disk_usage_percent,omitempty"`

	// Network statistics
	NetworkInMBps  float64 `json:"network_in_mbps,omitempty"`
	NetworkOutMBps float64 `json:"network_out_mbps,omitempty"`

	// Status details
	Message   string     `json:"message,omitempty"`
	UpdatedAt CustomTime `json:"updated_at"`
}

// VMOperation represents an asynchronous virtual machine operation
type VMOperation struct {
	ID           uint       `json:"id"`
	VMID         uint       `json:"vm_id"`
	OperationType string    `json:"operation_type"` // "start", "stop", "restart", "delete", "create"
	Status       string     `json:"status"`         // "pending", "in_progress", "completed", "failed"
	Progress     int        `json:"progress"`       // 0-100

	// Operation details
	RequestedBy string     `json:"requested_by,omitempty"`
	Message     string     `json:"message,omitempty"`
	ErrorDetails string    `json:"error_details,omitempty"`

	// Timestamps
	CreatedAt   CustomTime  `json:"created_at"`
	StartedAt   *CustomTime `json:"started_at,omitempty"`
	CompletedAt *CustomTime `json:"completed_at,omitempty"`

	// Result
	Result map[string]interface{} `json:"result,omitempty"`
}

// Report represents a generated report
type Report struct {
	ID             uint                   `json:"id"`
	OrganizationID uint                   `json:"organization_id"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description,omitempty"`
	ReportType     string                 `json:"report_type"` // "usage", "performance", "compliance", "billing"
	Status         string                 `json:"status"`      // "pending", "generating", "completed", "failed"
	Format         string                 `json:"format"`      // "pdf", "csv", "json", "html"
	FileURL        string                 `json:"file_url,omitempty"`
	FilePath       string                 `json:"file_path,omitempty"`
	FileSize       int64                  `json:"file_size,omitempty"`
	Configuration  *ReportConfiguration   `json:"configuration,omitempty"`
	CreatedBy      uint                   `json:"created_by,omitempty"`
	CreatedAt      CustomTime             `json:"created_at"`
	StartedAt      *CustomTime            `json:"started_at,omitempty"`
	CompletedAt    *CustomTime            `json:"completed_at,omitempty"`
	ExpiresAt      *CustomTime            `json:"expires_at,omitempty"`
	ErrorMessage   string                 `json:"error_message,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// ReportConfiguration defines report generation parameters
type ReportConfiguration struct {
	ReportType   string                 `json:"report_type"` // "usage", "performance", "compliance", "billing"
	Format       string                 `json:"format"`      // "pdf", "csv", "json", "html"
	Name         string                 `json:"name,omitempty"`
	Description  string                 `json:"description,omitempty"`
	TimeRange    *ReportTimeRange       `json:"time_range,omitempty"`
	Filters      *ReportFilter          `json:"filters,omitempty"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
	IncludeSections []string            `json:"include_sections,omitempty"`
	Delivery     *ReportDeliveryOptions `json:"delivery,omitempty"`
}

// ReportTimeRange defines the time period for report data
type ReportTimeRange struct {
	StartDate string `json:"start_date"` // ISO 8601 format
	EndDate   string `json:"end_date"`   // ISO 8601 format
	Preset    string `json:"preset,omitempty"` // "last_7_days", "last_30_days", "last_month", "last_quarter", "ytd"
}

// ReportFilter defines filtering criteria for report data
type ReportFilter struct {
	ServerIDs      []uint   `json:"server_ids,omitempty"`
	ServerUUIDs    []string `json:"server_uuids,omitempty"`
	ServerTags     []string `json:"server_tags,omitempty"`
	Locations      []string `json:"locations,omitempty"`
	Environments   []string `json:"environments,omitempty"`
	MetricTypes    []string `json:"metric_types,omitempty"`
	AlertTypes     []string `json:"alert_types,omitempty"`
	Severity       []string `json:"severity,omitempty"`
	IncludeInactive bool    `json:"include_inactive,omitempty"`
}

// ReportDeliveryOptions defines how reports should be delivered
type ReportDeliveryOptions struct {
	EmailRecipients []string `json:"email_recipients,omitempty"`
	EmailSubject    string   `json:"email_subject,omitempty"`
	EmailBody       string   `json:"email_body,omitempty"`
	WebhookURL      string   `json:"webhook_url,omitempty"`
	AutoDelete      bool     `json:"auto_delete,omitempty"`
	RetentionDays   int      `json:"retention_days,omitempty"`
}

// ReportSchedule represents a scheduled report
type ReportSchedule struct {
	ID             uint                   `json:"id"`
	OrganizationID uint                   `json:"organization_id"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description,omitempty"`
	Configuration  *ReportConfiguration   `json:"configuration"`
	Schedule       string                 `json:"schedule"` // Cron expression
	Enabled        bool                   `json:"enabled"`
	NextRunAt      *CustomTime            `json:"next_run_at,omitempty"`
	LastRunAt      *CustomTime            `json:"last_run_at,omitempty"`
	LastReportID   *uint                  `json:"last_report_id,omitempty"`
	CreatedBy      uint                   `json:"created_by,omitempty"`
	CreatedAt      CustomTime             `json:"created_at"`
	UpdatedAt      CustomTime             `json:"updated_at"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// ReportStatus represents the current status of a report generation process
type ReportStatus struct {
	ReportID   uint    `json:"report_id"`
	Status     string  `json:"status"` // "pending", "generating", "completed", "failed"
	Progress   int     `json:"progress"` // 0-100
	Message    string  `json:"message,omitempty"`
	Error      string  `json:"error,omitempty"`
	EstimatedCompletionTime *CustomTime `json:"estimated_completion_time,omitempty"`
}

// ServerGroup represents a logical group of servers
type ServerGroup struct {
	ID             uint                   `json:"id"`
	OrganizationID uint                   `json:"organization_id"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description,omitempty"`
	ServerCount    int                    `json:"server_count"`
	Tags           []string               `json:"tags,omitempty"`
	CreatedBy      uint                   `json:"created_by,omitempty"`
	CreatedAt      CustomTime             `json:"created_at"`
	UpdatedAt      CustomTime             `json:"updated_at"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// ServerGroupMembership represents a server's membership in a group
type ServerGroupMembership struct {
	GroupID      uint       `json:"group_id"`
	GroupName    string     `json:"group_name,omitempty"`
	ServerID     uint       `json:"server_id"`
	ServerUUID   string     `json:"server_uuid"`
	ServerName   string     `json:"server_name,omitempty"`
	ServerStatus string     `json:"server_status,omitempty"`
	AddedAt      CustomTime `json:"added_at"`
	AddedBy      uint       `json:"added_by,omitempty"`
}

// SearchResult represents a server search result with relevance scoring
type SearchResult struct {
	ServerID       uint                   `json:"server_id"`
	ServerUUID     string                 `json:"server_uuid"`
	ServerName     string                 `json:"server_name"`
	Hostname       string                 `json:"hostname,omitempty"`
	OrganizationID uint                   `json:"organization_id"`
	Location       string                 `json:"location,omitempty"`
	Environment    string                 `json:"environment,omitempty"`
	Classification string                 `json:"classification,omitempty"`
	Status         string                 `json:"status"`
	IPAddresses    []string               `json:"ip_addresses,omitempty"`
	Tags           []string               `json:"tags,omitempty"`
	RelevanceScore float64                `json:"relevance_score"`
	MatchedFields  []string               `json:"matched_fields,omitempty"`
	LastSeenAt     *CustomTime            `json:"last_seen_at,omitempty"`
	CreatedAt      CustomTime             `json:"created_at"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// TagSearchResult represents a tag search result with usage information
type TagSearchResult struct {
	TagID          uint                   `json:"tag_id"`
	TagName        string                 `json:"tag_name"`
	TagType        string                 `json:"tag_type"`        // "manual", "auto", "system"
	Scope          string                 `json:"scope"`           // "organization", "user", "server"
	Description    string                 `json:"description,omitempty"`
	Color          string                 `json:"color,omitempty"`
	UsageCount     int                    `json:"usage_count"`     // Number of resources using this tag
	ServerCount    int                    `json:"server_count"`    // Number of servers with this tag
	RelevanceScore float64                `json:"relevance_score"`
	MatchedFields  []string               `json:"matched_fields,omitempty"`
	CreatedAt      CustomTime             `json:"created_at"`
	UpdatedAt      CustomTime             `json:"updated_at"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// TagStatistics represents comprehensive statistics about tag usage
type TagStatistics struct {
	TotalTags       int                    `json:"total_tags"`
	ManualTags      int                    `json:"manual_tags"`
	AutoTags        int                    `json:"auto_tags"`
	SystemTags      int                    `json:"system_tags"`
	TagsByScope     map[string]int         `json:"tags_by_scope"`     // Breakdown by scope
	MostUsedTags    []TagUsageStats        `json:"most_used_tags"`    // Top 10 most used tags
	RecentlyCreated []TagSearchResult      `json:"recently_created"`  // Recently created tags
	UnusedTags      int                    `json:"unused_tags"`       // Tags with no usage
	AveragePerServer float64               `json:"average_per_server"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// TagUsageStats represents usage statistics for a specific tag
type TagUsageStats struct {
	TagID       uint       `json:"tag_id"`
	TagName     string     `json:"tag_name"`
	TagType     string     `json:"tag_type"`
	UsageCount  int        `json:"usage_count"`
	ServerCount int        `json:"server_count"`
	LastUsedAt  CustomTime `json:"last_used_at"`
}

// ============================================================================
// Audit Models
// ============================================================================

// AuditLog represents a single audit log entry tracking system activity
type AuditLog struct {
	ID               uint                   `json:"id"`
	OrganizationID   uint                   `json:"organization_id"`
	UserID           *uint                  `json:"user_id,omitempty"`
	UserEmail        string                 `json:"user_email,omitempty"`
	UserName         string                 `json:"user_name,omitempty"`
	Action           string                 `json:"action"`           // create, update, delete, login, logout, etc.
	ResourceType     string                 `json:"resource_type"`    // server, user, organization, alert, etc.
	ResourceID       string                 `json:"resource_id,omitempty"`
	ResourceName     string                 `json:"resource_name,omitempty"`
	Description      string                 `json:"description"`
	IPAddress        string                 `json:"ip_address,omitempty"`
	UserAgent        string                 `json:"user_agent,omitempty"`
	Severity         string                 `json:"severity"`         // info, warning, critical
	Status           string                 `json:"status"`           // success, failure, pending
	Changes          map[string]interface{} `json:"changes,omitempty"`          // Before/after values
	RequestID        string                 `json:"request_id,omitempty"`
	SessionID        string                 `json:"session_id,omitempty"`
	Location         string                 `json:"location,omitempty"`         // Geographic location
	DeviceType       string                 `json:"device_type,omitempty"`      // desktop, mobile, tablet
	ErrorMessage     string                 `json:"error_message,omitempty"`
	DurationMs       int                    `json:"duration_ms,omitempty"`      // Operation duration
	ComplianceFlags  []string               `json:"compliance_flags,omitempty"` // GDPR, HIPAA, SOC2, etc.
	CreatedAt        CustomTime             `json:"created_at"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// AuditStatistics represents comprehensive audit activity statistics
type AuditStatistics struct {
	TotalLogs           int                    `json:"total_logs"`
	TotalUsers          int                    `json:"total_users"`            // Unique users with activity
	TotalActions        int                    `json:"total_actions"`          // Distinct action types
	ActionBreakdown     map[string]int         `json:"action_breakdown"`       // Count by action type
	ResourceBreakdown   map[string]int         `json:"resource_breakdown"`     // Count by resource type
	SeverityBreakdown   map[string]int         `json:"severity_breakdown"`     // Count by severity
	StatusBreakdown     map[string]int         `json:"status_breakdown"`       // Count by status
	TopUsers            []AuditUserActivity    `json:"top_users"`              // Most active users
	TopActions          []AuditActionCount     `json:"top_actions"`            // Most common actions
	TopResources        []AuditResourceCount   `json:"top_resources"`          // Most accessed resources
	FailedAttempts      int                    `json:"failed_attempts"`        // Failed operations
	CriticalEvents      int                    `json:"critical_events"`        // Critical severity events
	ComplianceBreakdown map[string]int         `json:"compliance_breakdown"`   // Count by compliance flag
	AverageDurationMs   float64                `json:"average_duration_ms"`    // Average operation duration
	TimeRange           AuditTimeRange         `json:"time_range"`             // Statistics time range
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

// AuditUserActivity represents audit activity for a specific user
type AuditUserActivity struct {
	UserID          uint       `json:"user_id"`
	UserEmail       string     `json:"user_email"`
	UserName        string     `json:"user_name,omitempty"`
	ActionCount     int        `json:"action_count"`
	FailedAttempts  int        `json:"failed_attempts"`
	LastActivity    CustomTime `json:"last_activity"`
	TopActions      []string   `json:"top_actions,omitempty"`
}

// AuditActionCount represents count of a specific action type
type AuditActionCount struct {
	Action      string `json:"action"`
	Count       int    `json:"count"`
	SuccessRate float64 `json:"success_rate"` // Percentage of successful operations
}

// AuditResourceCount represents access count for a specific resource
type AuditResourceCount struct {
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id,omitempty"`
	ResourceName string `json:"resource_name,omitempty"`
	AccessCount  int    `json:"access_count"`
}

// AuditTimeRange represents the time range for audit statistics
type AuditTimeRange struct {
	StartDate  CustomTime `json:"start_date"`
	EndDate    CustomTime `json:"end_date"`
	DurationMs int64      `json:"duration_ms"` // Range duration in milliseconds
}

// ============================================================================
// Task Models
// ============================================================================

// Task represents a background task or scheduled job
type Task struct {
	ID               uint                   `json:"id"`
	OrganizationID   uint                   `json:"organization_id"`
	Name             string                 `json:"name"`
	Type             string                 `json:"type"`             // report_generation, data_export, cleanup, notification, etc.
	Status           string                 `json:"status"`           // pending, running, completed, failed, cancelled
	Priority         string                 `json:"priority"`         // low, normal, high, critical
	Parameters       map[string]interface{} `json:"parameters,omitempty"`
	Result           map[string]interface{} `json:"result,omitempty"` // Result data for completed tasks
	ErrorMessage     string                 `json:"error_message,omitempty"`
	Progress         int                    `json:"progress"`          // 0-100 percentage
	Schedule         string                 `json:"schedule,omitempty"` // Cron expression for recurring tasks
	ScheduledAt      *CustomTime            `json:"scheduled_at,omitempty"`
	StartedAt        *CustomTime            `json:"started_at,omitempty"`
	CompletedAt      *CustomTime            `json:"completed_at,omitempty"`
	ExecutionCount   int                    `json:"execution_count"`    // Number of times executed
	LastExecutionID  *uint                  `json:"last_execution_id,omitempty"`
	NextExecutionAt  *CustomTime            `json:"next_execution_at,omitempty"` // For recurring tasks
	MaxRetries       int                    `json:"max_retries"`
	CurrentRetry     int                    `json:"current_retry"`
	TimeoutSeconds   int                    `json:"timeout_seconds,omitempty"`
	CreatedBy        uint                   `json:"created_by,omitempty"`
	CreatedAt        CustomTime             `json:"created_at"`
	UpdatedAt        CustomTime             `json:"updated_at"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// TaskConfiguration represents parameters for creating a new task
type TaskConfiguration struct {
	Name           string                 `json:"name"`
	Type           string                 `json:"type"`
	Priority       string                 `json:"priority,omitempty"`        // Default: normal
	Parameters     map[string]interface{} `json:"parameters,omitempty"`
	Schedule       string                 `json:"schedule,omitempty"`        // Cron expression
	ScheduledAt    *CustomTime            `json:"scheduled_at,omitempty"`    // One-time scheduled task
	MaxRetries     int                    `json:"max_retries,omitempty"`     // Default: 3
	TimeoutSeconds int                    `json:"timeout_seconds,omitempty"` // Default: 300
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// TaskExecution represents a single execution of a task
type TaskExecution struct {
	ID             uint                   `json:"id"`
	TaskID         uint                   `json:"task_id"`
	Status         string                 `json:"status"`        // running, completed, failed
	Progress       int                    `json:"progress"`      // 0-100 percentage
	Result         map[string]interface{} `json:"result,omitempty"`
	ErrorMessage   string                 `json:"error_message,omitempty"`
	StartedAt      CustomTime             `json:"started_at"`
	CompletedAt    *CustomTime            `json:"completed_at,omitempty"`
	DurationMs     int                    `json:"duration_ms,omitempty"`
	RetryCount     int                    `json:"retry_count"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// TaskStatistics represents aggregated statistics about task execution
type TaskStatistics struct {
	TotalTasks          int                  `json:"total_tasks"`
	PendingTasks        int                  `json:"pending_tasks"`
	RunningTasks        int                  `json:"running_tasks"`
	CompletedTasks      int                  `json:"completed_tasks"`
	FailedTasks         int                  `json:"failed_tasks"`
	CancelledTasks      int                  `json:"cancelled_tasks"`
	TasksByType         map[string]int       `json:"tasks_by_type"`
	TasksByPriority     map[string]int       `json:"tasks_by_priority"`
	AverageDurationMs   float64              `json:"average_duration_ms"`
	SuccessRate         float64              `json:"success_rate"`          // Percentage
	TotalExecutions     int                  `json:"total_executions"`
	FailedExecutions    int                  `json:"failed_executions"`
	AverageRetries      float64              `json:"average_retries"`
	LongestRunningTask  *TaskSummary         `json:"longest_running_task,omitempty"`
	MostRecentFailure   *TaskSummary         `json:"most_recent_failure,omitempty"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

// TaskSummary represents a simplified task summary for statistics
type TaskSummary struct {
	ID           uint       `json:"id"`
	Name         string     `json:"name"`
	Type         string     `json:"type"`
	Status       string     `json:"status"`
	DurationMs   int        `json:"duration_ms,omitempty"`
	ErrorMessage string     `json:"error_message,omitempty"`
	CreatedAt    CustomTime `json:"created_at"`
}

// ============================================================================
// Notification Models
// ============================================================================

// NotificationRequest represents a request to send a notification
type NotificationRequest struct {
	OrganizationID uint                   `json:"organization_id"`
	ChannelIDs     []uint                 `json:"channel_ids,omitempty"`
	ChannelTypes   []string               `json:"channel_types,omitempty"`
	Subject        string                 `json:"subject"`
	Content        string                 `json:"content"`
	ContentType    string                 `json:"content_type,omitempty"` // "text" or "html"
	Recipients     []string               `json:"recipients,omitempty"`
	Priority       NotificationPriority   `json:"priority,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	AlertID        *uint                  `json:"alert_id,omitempty"`
	ProbeID        *uint                  `json:"probe_id,omitempty"`
	ScheduledAt    *CustomTime            `json:"scheduled_at,omitempty"`
	ExpiresAt      *CustomTime            `json:"expires_at,omitempty"`
	MaxRetries     *int                   `json:"max_retries,omitempty"`
	RetryDelay     *int                   `json:"retry_delay_minutes,omitempty"`
}

// BatchNotificationRequest represents a request to send multiple notifications
type BatchNotificationRequest struct {
	Notifications []NotificationRequest `json:"notifications"`
}

// NotificationPriority represents the priority level of a notification
type NotificationPriority string

const (
	NotificationPriorityLow      NotificationPriority = "low"
	NotificationPriorityNormal   NotificationPriority = "normal"
	NotificationPriorityHigh     NotificationPriority = "high"
	NotificationPriorityCritical NotificationPriority = "critical"
)

// String returns the string representation of NotificationPriority
func (np NotificationPriority) String() string {
	return string(np)
}

// NotificationResponse represents the response to a notification request
type NotificationResponse struct {
	ID             uint              `json:"id"`
	OrganizationID uint              `json:"organization_id"`
	Status         string            `json:"status"`
	ChannelsUsed   []ChannelUsageInfo `json:"channels_used"`
	CreatedAt      CustomTime        `json:"created_at"`
	ScheduledAt    *CustomTime       `json:"scheduled_at,omitempty"`
	SentAt         *CustomTime       `json:"sent_at,omitempty"`
	Message        string            `json:"message,omitempty"`
}

// BatchNotificationResponse represents the response to a batch notification request
type BatchNotificationResponse struct {
	TotalRequested int                    `json:"total_requested"`
	TotalAccepted  int                    `json:"total_accepted"`
	TotalRejected  int                    `json:"total_rejected"`
	Results        []NotificationResponse `json:"results"`
	Errors         []string               `json:"errors,omitempty"`
}

// ChannelUsageInfo provides information about how a channel was used for a notification
type ChannelUsageInfo struct {
	ChannelID   uint   `json:"channel_id"`
	ChannelName string `json:"channel_name"`
	ChannelType string `json:"channel_type"`
	Status      string `json:"status"`
	Recipient   string `json:"recipient"`
	Error       string `json:"error,omitempty"`
}

// NotificationStatusRequest represents a request to get notification status
type NotificationStatusRequest struct {
	NotificationIDs []uint `json:"notification_ids"`
}

// NotificationStatusResponse represents the response to a notification status request
type NotificationStatusResponse struct {
	Notifications []NotificationStatusInfo `json:"notifications"`
}

// NotificationStatusInfo provides detailed status information for a notification
type NotificationStatusInfo struct {
	ID              uint                   `json:"id"`
	OrganizationID  uint                   `json:"organization_id"`
	Status          string                 `json:"status"`
	Subject         string                 `json:"subject"`
	CreatedAt       CustomTime             `json:"created_at"`
	ScheduledAt     *CustomTime            `json:"scheduled_at,omitempty"`
	SentAt          *CustomTime            `json:"sent_at,omitempty"`
	DeliveredAt     *CustomTime            `json:"delivered_at,omitempty"`
	FailedAt        *CustomTime            `json:"failed_at,omitempty"`
	RetryCount      int                    `json:"retry_count"`
	NextRetryAt     *CustomTime            `json:"next_retry_at,omitempty"`
	ErrorMessage    string                 `json:"error_message,omitempty"`
	ChannelDelivery []ChannelDeliveryInfo  `json:"channel_delivery"`
}

// ChannelDeliveryInfo provides delivery information for a specific channel
type ChannelDeliveryInfo struct {
	ChannelID      uint        `json:"channel_id"`
	ChannelName    string      `json:"channel_name"`
	ChannelType    string      `json:"channel_type"`
	Status         string      `json:"status"`
	Recipient      string      `json:"recipient"`
	SentAt         *CustomTime `json:"sent_at,omitempty"`
	DeliveredAt    *CustomTime `json:"delivered_at,omitempty"`
	FailedAt       *CustomTime `json:"failed_at,omitempty"`
	RetryCount     int         `json:"retry_count"`
	NextRetryAt    *CustomTime `json:"next_retry_at,omitempty"`
	ErrorMessage   string      `json:"error_message,omitempty"`
	ExternalID     string      `json:"external_id,omitempty"`
	ProviderStatus string      `json:"provider_status,omitempty"`
}

// ============================================================================
// Cluster Models
// ============================================================================

// Cluster represents a remote Kubernetes cluster for deployment and monitoring
type Cluster struct {
	ID            uint         `json:"id"`
	Name          string       `json:"name"`                // Unique cluster name
	APIServerURL  string       `json:"api_server_url"`      // Kubernetes API server URL
	Token         string       `json:"token"`               // Service account token for authentication
	CACert        string       `json:"ca_cert,omitempty"`   // CA certificate for secure connection
	Status        string       `json:"status"`              // online, offline, error, unknown
	LastChecked   *CustomTime  `json:"last_checked,omitempty"`        // Last health check time
	LastConnected *CustomTime  `json:"last_connected,omitempty"`      // Last successful connection time
	ErrorMessage  string       `json:"error_message,omitempty"` // Error details if connection failed
	NodeCount     int          `json:"node_count"`          // Number of nodes in cluster
	PodCount      int          `json:"pod_count"`           // Number of pods running
	IsActive      bool         `json:"is_active"`           // Whether cluster monitoring is active
	CreatedAt     CustomTime   `json:"created_at"`
	UpdatedAt     CustomTime   `json:"updated_at"`
}

// ClusterCreateRequest represents a request to create a new cluster
type ClusterCreateRequest struct {
	Name         string `json:"name"`                   // Unique cluster name (required)
	APIServerURL string `json:"api_server_url"`         // Kubernetes API server URL (required)
	Token        string `json:"token"`                  // Service account token (required)
	CACert       string `json:"ca_cert,omitempty"`      // CA certificate (optional)
	IsActive     *bool  `json:"is_active,omitempty"`    // Enable/disable monitoring (default: true)
}

// ClusterUpdateRequest represents a request to update an existing cluster
type ClusterUpdateRequest struct {
	Name         *string `json:"name,omitempty"`          // Updated cluster name
	APIServerURL *string `json:"api_server_url,omitempty"` // Updated API server URL
	Token        *string `json:"token,omitempty"`         // Updated service account token
	CACert       *string `json:"ca_cert,omitempty"`       // Updated CA certificate
	IsActive     *bool   `json:"is_active,omitempty"`     // Enable/disable monitoring
}

// ClusterStatistics provides aggregate statistics across all monitored clusters
type ClusterStatistics struct {
	TotalClusters      int                  `json:"total_clusters"`       // Total number of clusters
	OnlineClusters     int                  `json:"online_clusters"`      // Clusters currently online
	OfflineClusters    int                  `json:"offline_clusters"`     // Clusters currently offline
	ErrorClusters      int                  `json:"error_clusters"`       // Clusters with errors
	TotalNodes         int                  `json:"total_nodes"`          // Total nodes across all clusters
	TotalPods          int                  `json:"total_pods"`           // Total pods across all clusters
	AverageNodeCount   float64              `json:"average_node_count"`   // Average nodes per cluster
	AveragePodCount    float64              `json:"average_pod_count"`    // Average pods per cluster
	ClustersByStatus   map[string]int       `json:"clusters_by_status"`   // Count grouped by status
	LastCheckTime      CustomTime           `json:"last_check_time"`      // Most recent health check
}

// ============================================================================
// Package/Tier Models
// ============================================================================

// OrganizationPackage represents an organization's subscription package and limits
type OrganizationPackage struct {
	ID                    uint64     `json:"id"`
	OrganizationID        uint       `json:"organization_id"`
	OrganizationUUID      string     `json:"organization_uuid"`
	PackageTier           string     `json:"package_tier"`                // starter, professional, enterprise
	MaxProbes             int        `json:"max_probes"`                  // Maximum number of probes allowed
	MaxRegions            int        `json:"max_regions"`                 // Maximum number of regions for probes
	MinFrequency          int        `json:"min_frequency"`               // Minimum probe check frequency (seconds)
	ProbeFrequencySeconds int        `json:"probe_frequency_seconds"`     // Default probe frequency (seconds)
	MaxAlertChannels      int        `json:"max_alert_channels"`          // Maximum alert notification channels
	MaxStatusPages        int        `json:"max_status_pages"`            // Maximum public status pages
	AllowedProbeTypes     []string   `json:"allowed_probe_types"`         // HTTP, ICMP, TCP, DNS, etc.
	Features              []string   `json:"features"`                    // Enabled features for this package
	SelectedRegions       []string   `json:"selected_regions,omitempty"`  // Regions selected for probes
	Active                bool       `json:"active"`                      // Whether package is currently active
	SubscriptionStatus    string     `json:"subscription_status"`         // active, canceled, past_due, etc.
	CurrentPeriodStart    CustomTime `json:"current_period_start"`        // Billing period start
	CurrentPeriodEnd      CustomTime `json:"current_period_end"`          // Billing period end
	CancelAtPeriodEnd     bool       `json:"cancel_at_period_end"`        // Whether to cancel at period end
	TrialEndsAt           *CustomTime `json:"trial_ends_at,omitempty"`    // Trial expiration date
	CreatedAt             CustomTime `json:"created_at"`
	UpdatedAt             CustomTime `json:"updated_at"`
}

// PackageUpgradeRequest represents a request to upgrade organization package tier
type PackageUpgradeRequest struct {
	NewTier         string                 `json:"new_tier"`                    // Target tier: starter, professional, enterprise
	PaymentMethodID *string                `json:"payment_method_id,omitempty"` // Stripe payment method ID (optional)
	BillingEmail    *string                `json:"billing_email,omitempty"`     // Billing contact email (optional)
	Metadata        map[string]interface{} `json:"metadata,omitempty"`          // Additional upgrade metadata
}

// ProbeConfigValidationRequest represents a request to validate probe configuration against package limits
type ProbeConfigValidationRequest struct {
	ProbeType        string   `json:"probe_type"`                   // HTTP, ICMP, TCP, DNS, etc.
	Frequency        int      `json:"frequency"`                    // Check frequency in seconds
	Regions          []string `json:"regions"`                      // Regions for probe execution
	AdditionalProbes *int     `json:"additional_probes,omitempty"`  // Number of new probes being created (optional)
}

// ProbeConfigValidationResult represents the result of probe configuration validation
type ProbeConfigValidationResult struct {
	Valid              bool     `json:"valid"`                           // Whether configuration is valid
	ProbeTypeAllowed   bool     `json:"probe_type_allowed"`              // Whether probe type is allowed
	FrequencyAllowed   bool     `json:"frequency_allowed"`               // Whether frequency is allowed
	RegionsAllowed     bool     `json:"regions_allowed"`                 // Whether number of regions is allowed
	ProbeCountAllowed  bool     `json:"probe_count_allowed"`             // Whether probe count is within limits
	Violations         []string `json:"violations,omitempty"`            // List of limit violations
	CurrentProbeCount  int      `json:"current_probe_count"`             // Current number of probes
	MaxProbes          int      `json:"max_probes"`                      // Maximum probes allowed
	MinFrequency       int      `json:"min_frequency"`                   // Minimum frequency allowed (seconds)
	MaxRegions         int      `json:"max_regions"`                     // Maximum regions allowed
	AllowedProbeTypes  []string `json:"allowed_probe_types,omitempty"`   // List of allowed probe types
	UpgradeSuggestion  string   `json:"upgrade_suggestion,omitempty"`    // Suggested tier for meeting requirements
}

// ============================================================
// Quota History Types
// ============================================================

// QuotaUsageRecordRequest represents a batch of quota usage records to store
type QuotaUsageRecordRequest struct {
	Records []QuotaUsageRecord `json:"records"`
}

// QuotaUsageRecord represents a single quota usage data point
type QuotaUsageRecord struct {
	OrganizationID uint      `json:"organization_id"`
	ResourceType   string    `json:"resource_type"` // cpu, memory, storage, pods, services, configmaps, secrets, persistentvolumeclaims
	UsedAmount     int64     `json:"used_amount"`
	HardLimit      int64     `json:"hard_limit"`
	CollectedAt    time.Time `json:"collected_at"`
}

// QuotaUsageHistory represents a historical quota usage record
type QuotaUsageHistory struct {
	ID                 uint64    `json:"id"`
	OrganizationID     uint      `json:"organization_id"`
	ResourceType       string    `json:"resource_type"`
	UsedAmount         int64     `json:"used_amount"`
	HardLimit          int64     `json:"hard_limit"`
	UtilizationPercent float64   `json:"utilization_percent"`
	CollectedAt        time.Time `json:"collected_at"`
	CreatedAt          time.Time `json:"created_at"`
}

// AverageUtilizationResponse represents average utilization statistics
type AverageUtilizationResponse struct {
	OrganizationID     uint    `json:"organization_id"`
	ResourceType       string  `json:"resource_type"`
	AverageUtilization float64 `json:"average_utilization"`
	AverageUsedAmount  float64 `json:"average_used_amount"`
	AverageHardLimit   float64 `json:"average_hard_limit"`
	StartDate          string  `json:"start_date"`
	EndDate            string  `json:"end_date"`
	SampleCount        int     `json:"sample_count"`
}

// DailyAggregateResponse represents daily aggregated quota statistics
type DailyAggregateResponse struct {
	Date               string  `json:"date"`
	AverageUtilization float64 `json:"average_utilization"`
	MaxUtilization     float64 `json:"max_utilization"`
	MinUtilization     float64 `json:"min_utilization"`
	AverageUsedAmount  float64 `json:"average_used_amount"`
	SampleCount        int     `json:"sample_count"`
}

// ResourceSummaryResponse represents summary statistics for a single resource type
type ResourceSummaryResponse struct {
	ResourceType       string  `json:"resource_type"`
	CurrentUtilization float64 `json:"current_utilization"`
	AverageUtilization float64 `json:"average_utilization"`
	PeakUtilization    float64 `json:"peak_utilization"`
	CurrentUsedAmount  int64   `json:"current_used_amount"`
	CurrentHardLimit   int64   `json:"current_hard_limit"`
	SampleCount        int     `json:"sample_count"`
}

// UsageTrendResponse represents trend analysis results
type UsageTrendResponse struct {
	OrganizationID uint    `json:"organization_id"`
	ResourceType   string  `json:"resource_type"`
	TrendSlope     float64 `json:"trend_slope"`
	TrendDirection string  `json:"trend_direction"` // increasing, decreasing, stable
	CurrentValue   float64 `json:"current_value"`
	PredictedValue float64 `json:"predicted_value"`
	DaysAnalyzed   int     `json:"days_analyzed"`
	SampleCount    int     `json:"sample_count"`
	StartDate      string  `json:"start_date"`
	EndDate        string  `json:"end_date"`
}

// UsagePattern represents a detected usage pattern
type UsagePattern struct {
	PatternType      string  `json:"pattern_type"` // high_utilization, rapid_growth, high_volatility, near_limit
	Description      string  `json:"description"`
	Severity         string  `json:"severity"` // info, warning, critical
	AffectedResource string  `json:"affected_resource"`
	DetectedValue    float64 `json:"detected_value"`
	ThresholdValue   float64 `json:"threshold_value"`
	Recommendation   string  `json:"recommendation"`
}

// UsagePatternsResponse contains all detected patterns
type UsagePatternsResponse struct {
	OrganizationID uint           `json:"organization_id"`
	AnalysisDate   string         `json:"analysis_date"`
	Patterns       []UsagePattern `json:"patterns"`
	PatternCount   int            `json:"pattern_count"`
}

// Organization-Level Health Check Types (Task 307)

// CreateHealthCheckDefinitionRequest represents a request to create an organization-level health check definition
type CreateHealthCheckDefinitionRequest struct {
	CheckType       string                 `json:"check_type" validate:"required"`
	CheckName       string                 `json:"check_name" validate:"required"`
	Description     string                 `json:"description,omitempty"`
	Enabled         bool                   `json:"enabled"`
	IntervalSeconds int                    `json:"interval_seconds" validate:"min=10"`
	TimeoutSeconds  int                    `json:"timeout_seconds" validate:"min=1"`
	TargetName      string                 `json:"target_name" validate:"required"`
	TargetConfig    map[string]interface{} `json:"target_config"`
	Thresholds      map[string]interface{} `json:"thresholds"`
}

// HealthCheckDefinitionResponse represents an organization-level health check definition
type HealthCheckDefinitionResponse struct {
	ID              uint64                 `json:"id"`
	OrganizationID  uint64                 `json:"organization_id"`
	CheckType       string                 `json:"check_type"`
	CheckName       string                 `json:"check_name"`
	Description     string                 `json:"description"`
	Enabled         bool                   `json:"enabled"`
	IntervalSeconds int                    `json:"interval_seconds"`
	TimeoutSeconds  int                    `json:"timeout_seconds"`
	TargetName      string                 `json:"target_name"`
	TargetConfig    map[string]interface{} `json:"target_config"`
	Thresholds      map[string]interface{} `json:"thresholds"`
	CreatedAt       string                 `json:"created_at"`
	UpdatedAt       string                 `json:"updated_at"`
}

// ListHealthCheckDefinitionsResponse represents a list of health check definitions
type ListHealthCheckDefinitionsResponse struct {
	Status      string                         `json:"status"`
	Message     string                         `json:"message"`
	Definitions []HealthCheckDefinitionResponse `json:"definitions"`
	Total       int64                          `json:"total"`
}

// SubmitHealthCheckResultRequest represents a request to submit a health check result
type SubmitHealthCheckResultRequest struct {
	DefinitionID        uint64                 `json:"definition_id" validate:"required"`
	Status              string                 `json:"status" validate:"required"`
	Score               int                    `json:"score" validate:"min=0,max=100"`
	ResponseTimeMs      int64                  `json:"response_time_ms,omitempty"`
	Message             string                 `json:"message,omitempty"`
	Details             map[string]interface{} `json:"details,omitempty"`
	ConsecutiveFailures int                    `json:"consecutive_failures,omitempty"`
}

// HealthCheckResultResponse represents a health check result
type HealthCheckResultResponse struct {
	ID                  uint64                 `json:"id"`
	OrganizationID      uint64                 `json:"organization_id"`
	DefinitionID        uint64                 `json:"definition_id"`
	Status              string                 `json:"status"`
	Score               int                    `json:"score"`
	ResponseTimeMs      int64                  `json:"response_time_ms"`
	Message             string                 `json:"message"`
	Details             map[string]interface{} `json:"details"`
	ExecutionTimeMs     int64                  `json:"execution_time_ms"`
	ConsecutiveFailures int                    `json:"consecutive_failures"`
	CreatedAt           string                 `json:"created_at"`
}

// HealthStatusAggregateResponse represents aggregated health status for an organization
type HealthStatusAggregateResponse struct {
	Status              string `json:"status"`
	Message             string `json:"message"`
	OrganizationID      uint64 `json:"organization_id"`
	DatabaseStatus      string `json:"database_status"`
	DatabaseScore       int    `json:"database_score"`
	APIStatus           string `json:"api_status"`
	APIScore            int    `json:"api_score"`
	ResourceStatus      string `json:"resource_status"`
	ResourceScore       int    `json:"resource_score"`
	MicroserviceStatus  string `json:"microservice_status"`
	MicroserviceScore   int    `json:"microservice_score"`
	OverallStatus       string `json:"overall_status"`
	OverallScore        int    `json:"overall_score"`
	HealthyCheckCount   int    `json:"healthy_check_count"`
	WarningCheckCount   int    `json:"warning_check_count"`
	CriticalCheckCount  int    `json:"critical_check_count"`
	UptimePercent       float64 `json:"uptime_percent"`
}

// HealthAlertResponse represents a health alert
type HealthAlertResponse struct {
	ID             uint64  `json:"id"`
	OrganizationID uint64  `json:"organization_id"`
	DefinitionID   uint64  `json:"definition_id"`
	Status         string  `json:"status"`
	Title          string  `json:"title"`
	Description    string  `json:"description"`
	Severity       string  `json:"severity"`
	Acknowledged   bool    `json:"acknowledged"`
	AcknowledgedAt *string `json:"acknowledged_at,omitempty"`
	ResolvedAt     *string `json:"resolved_at,omitempty"`
	CreatedAt      string  `json:"created_at"`
}

// ListHealthAlertsResponse represents a list of health alerts
type ListHealthAlertsResponse struct {
	Status  string                  `json:"status"`
	Message string                  `json:"message"`
	Alerts  []HealthAlertResponse   `json:"alerts"`
	Total   int64                   `json:"total"`
}
