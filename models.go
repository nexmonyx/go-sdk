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
	ID             uint                   `json:"id"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	OrganizationID uint                   `json:"organization_id"`
	
	// Scope configuration
	ScopeType      string                 `json:"scope_type"` // organization, server, tag, group
	ScopeID        *uint                  `json:"scope_id,omitempty"`
	ScopeValue     string                 `json:"scope_value,omitempty"`
	
	// Metric configuration
	MetricName     string                 `json:"metric_name"`
	Aggregation    string                 `json:"aggregation"` // avg, sum, min, max, count
	
	// Conditions
	Conditions     AlertConditions        `json:"conditions"`
	
	// Notification settings
	ChannelIDs     []uint                 `json:"channel_ids"`
	
	// State
	Enabled        bool                   `json:"enabled"`
	LastEvaluated  *time.Time             `json:"last_evaluated,omitempty"`
	
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// AlertConditions represents the conditions for triggering an alert
type AlertConditions struct {
	TimeWindow int                `json:"time_window"` // in minutes
	Thresholds []AlertThreshold   `json:"thresholds"`
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

// MonitoringAgent represents a monitoring agent or probe
type MonitoringAgent struct {
	GormModel
	UUID           string `json:"uuid"`
	NodeUUID       string `json:"node_uuid,omitempty"`
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
	Type           string `json:"type"`
	OrganizationID *uint  `json:"organization_id,omitempty"`

	// Agent configuration
	Region        string                 `json:"region"`
	Location      string                 `json:"location,omitempty"`
	IPAddress     string                 `json:"ip_address,omitempty"`
	Provider      string                 `json:"provider,omitempty"`
	Version       string                 `json:"version"`
	Capabilities  []string               `json:"capabilities"`
	Config        map[string]interface{} `json:"config,omitempty"`
	Enabled       bool                   `json:"enabled"`
	MaxProbes     int                    `json:"max_probes,omitempty"`
	CurrentProbes int                    `json:"current_probes,omitempty"`

	// Agent status
	Status        string      `json:"status"`
	LastHeartbeat *CustomTime `json:"last_heartbeat,omitempty"`
	LastError     string      `json:"last_error,omitempty"`
	ErrorCount    int         `json:"error_count"`

	// Performance metrics
	ProbesExecuted  int64   `json:"probes_executed"`
	ProbesFailed    int64   `json:"probes_failed"`
	AvgResponseTime float64 `json:"avg_response_time"`
	Uptime          float64 `json:"uptime"`
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
	System          *SystemHardwareInfo  `json:"system,omitempty"`
	Motherboard     *MotherboardInfo     `json:"motherboard,omitempty"`
	CPUs            []CPUInfo            `json:"cpus,omitempty"`
	Memory          *MemoryInfo          `json:"memory,omitempty"`
	MemoryModules   []MemoryModuleInfo   `json:"memory_modules,omitempty"`
	Storage         []StorageDeviceInfo  `json:"storage,omitempty"`
	StorageDevices  []StorageDeviceInfo  `json:"storage_devices,omitempty"` // Alias for Storage
	Network         []NetworkCardInfo    `json:"network,omitempty"`
	NetworkCards    []NetworkCardInfo    `json:"network_cards,omitempty"` // Alias for Network
	GPUs            []GPUInfo            `json:"gpus,omitempty"`
	PowerSupplies   []PowerSupplyInfo    `json:"power_supplies,omitempty"`
	RAIDControllers []RAIDControllerInfo `json:"raid_controllers,omitempty"`
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
	Model         string `json:"model,omitempty"`
	Manufacturer  string `json:"manufacturer,omitempty"`
	SerialNumber  string `json:"serial_number,omitempty"`
	MaxPowerWatts int    `json:"max_power_watts,omitempty"`
	Type          string `json:"type,omitempty"`
	Status        string `json:"status,omitempty"`
	Efficiency    string `json:"efficiency,omitempty"`
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
	ServerUUID    string                 `json:"server_uuid"`
	CollectedAt   string                 `json:"collected_at"`
	SystemInfo    *SystemInfo            `json:"system_info,omitempty"`
	CPU           *CPUMetrics            `json:"cpu,omitempty"`
	Memory        *MemoryMetrics         `json:"memory,omitempty"`
	Disks         []DiskMetrics          `json:"disks,omitempty"`
	Network       []NetworkMetrics       `json:"network,omitempty"`
	Processes     []ProcessMetrics       `json:"processes,omitempty"`
	CustomMetrics map[string]interface{} `json:"custom_metrics,omitempty"`
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
