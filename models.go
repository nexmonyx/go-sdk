package nexmonyx

import (
	"encoding/json"
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
	// Hardware details
	CPUModel     string `json:"cpu_model,omitempty"`
	CPUCount     int    `json:"cpu_count,omitempty"`
	CPUCores     int    `json:"cpu_cores,omitempty"`
	MemoryTotal  uint64 `json:"memory_total,omitempty"`
	StorageTotal uint64 `json:"storage_total,omitempty"`
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
