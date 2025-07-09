package nexmonyx

import (
	"context"
	"fmt"
	"time"
)

// NetworkHardwareService handles network hardware operations
type NetworkHardwareService struct {
	client *Client
}

// NetworkHardwareRequest represents the request to submit network hardware information
type NetworkHardwareRequest struct {
	ServerUUID string                `json:"server_uuid"`
	Interfaces []NetworkHardwareInfo `json:"interfaces"`
}

// NetworkHardwareInfo represents comprehensive network hardware data
type NetworkHardwareInfo struct {
	// Interface identification
	InterfaceName   string `json:"interface_name"`
	InterfaceType   string `json:"interface_type,omitempty"`
	InterfaceAlias  string `json:"interface_alias,omitempty"`
	MacAddress      string `json:"mac_address,omitempty"`
	HardwareAddress string `json:"hardware_address,omitempty"`

	// Physical hardware specifications
	Manufacturer    string `json:"manufacturer,omitempty"`
	DeviceName      string `json:"device_name,omitempty"`
	DriverName      string `json:"driver_name,omitempty"`
	DriverVersion   string `json:"driver_version,omitempty"`
	FirmwareVersion string `json:"firmware_version,omitempty"`
	PCISlot         string `json:"pci_slot,omitempty"`
	BusInfo         string `json:"bus_info,omitempty"`

	// Physical port specifications
	PortType       string   `json:"port_type,omitempty"`
	ConnectorType  string   `json:"connector_type,omitempty"`
	SupportedPorts []string `json:"supported_ports,omitempty"`

	// Speed and duplex capabilities
	SpeedMbps       int    `json:"speed_mbps,omitempty"`
	MaxSpeedMbps    int    `json:"max_speed_mbps,omitempty"`
	SupportedSpeeds []int  `json:"supported_speeds,omitempty"`
	Duplex          string `json:"duplex,omitempty"`
	AutoNegotiation bool   `json:"auto_negotiation,omitempty"`

	// Link status and configuration
	LinkDetected        bool   `json:"link_detected,omitempty"`
	CarrierStatus       bool   `json:"carrier_status,omitempty"`
	OperationalState    string `json:"operational_state,omitempty"`
	AdministrativeState string `json:"administrative_state,omitempty"`
	MTU                 int    `json:"mtu,omitempty"`

	// Network configuration
	IPAddresses      []string `json:"ip_addresses,omitempty"`
	IPv6Addresses    []string `json:"ipv6_addresses,omitempty"`
	SubnetMasks      []string `json:"subnet_masks,omitempty"`
	GatewayAddresses []string `json:"gateway_addresses,omitempty"`
	DNSServers       []string `json:"dns_servers,omitempty"`
	Domains          []string `json:"domains,omitempty"`

	// VLAN configuration
	VlanID       int   `json:"vlan_id,omitempty"`
	VlanParent   string `json:"vlan_parent,omitempty"`
	NativeVlan   int   `json:"native_vlan,omitempty"`
	AllowedVlans []int `json:"allowed_vlans,omitempty"`

	// Bonding/Teaming configuration
	BondMode        string   `json:"bond_mode,omitempty"`
	BondMaster      string   `json:"bond_master,omitempty"`
	BondSlaves      []string `json:"bond_slaves,omitempty"`
	BondPrimary     string   `json:"bond_primary,omitempty"`
	BondActiveSlave string   `json:"bond_active_slave,omitempty"`
	LACPRate        string   `json:"lacp_rate,omitempty"`
	XmitHashPolicy  string   `json:"xmit_hash_policy,omitempty"`

	// Bridge configuration
	BridgeMaster       string   `json:"bridge_master,omitempty"`
	BridgePorts        []string `json:"bridge_ports,omitempty"`
	BridgeSTP          bool     `json:"bridge_stp,omitempty"`
	BridgeForwardDelay int      `json:"bridge_forward_delay,omitempty"`
	BridgeHelloTime    int      `json:"bridge_hello_time,omitempty"`
	BridgeMaxAge       int      `json:"bridge_max_age,omitempty"`
	BridgePriority     int      `json:"bridge_priority,omitempty"`

	// Wake-on-LAN configuration
	WOLEnabled bool     `json:"wol_enabled,omitempty"`
	WOLModes   []string `json:"wol_modes,omitempty"`

	// Power management
	PowerManagement         bool `json:"power_management,omitempty"`
	EnergyEfficientEthernet bool `json:"energy_efficient_ethernet,omitempty"`

	// Statistics and metrics
	RxBytes       int64 `json:"rx_bytes,omitempty"`
	TxBytes       int64 `json:"tx_bytes,omitempty"`
	RxPackets     int64 `json:"rx_packets,omitempty"`
	TxPackets     int64 `json:"tx_packets,omitempty"`
	RxErrors      int64 `json:"rx_errors,omitempty"`
	TxErrors      int64 `json:"tx_errors,omitempty"`
	RxDropped     int64 `json:"rx_dropped,omitempty"`
	TxDropped     int64 `json:"tx_dropped,omitempty"`
	RxFifoErrors  int64 `json:"rx_fifo_errors,omitempty"`
	TxFifoErrors  int64 `json:"tx_fifo_errors,omitempty"`
	RxFrameErrors int64 `json:"rx_frame_errors,omitempty"`
	RxCRCErrors   int64 `json:"rx_crc_errors,omitempty"`
	Collisions    int64 `json:"collisions,omitempty"`

	// Advanced statistics
	Multicast       int64 `json:"multicast,omitempty"`
	RxLengthErrors  int64 `json:"rx_length_errors,omitempty"`
	RxOverErrors    int64 `json:"rx_over_errors,omitempty"`
	TxAbortedErrors int64 `json:"tx_aborted_errors,omitempty"`
	TxCarrierErrors int64 `json:"tx_carrier_errors,omitempty"`
	TxWindowErrors  int64 `json:"tx_window_errors,omitempty"`
	RxCompressed    int64 `json:"rx_compressed,omitempty"`
	TxCompressed    int64 `json:"tx_compressed,omitempty"`

	// Quality metrics
	SignalStrengthDBM  float64 `json:"signal_strength_dbm,omitempty"`
	LinkQualityPercent float64 `json:"link_quality_percent,omitempty"`
	NoiseLevelDBM      float64 `json:"noise_level_dbm,omitempty"`

	// Wireless specific
	IsWireless           bool    `json:"is_wireless,omitempty"`
	WirelessMode         string  `json:"wireless_mode,omitempty"`
	WirelessProtocol     string  `json:"wireless_protocol,omitempty"`
	WirelessFrequencyMHz float64 `json:"wireless_frequency_mhz,omitempty"`
	WirelessChannel      int     `json:"wireless_channel,omitempty"`
	WirelessSSID         string  `json:"wireless_ssid,omitempty"`
	WirelessBSSID        string  `json:"wireless_bssid,omitempty"`
	WirelessEncryption   string  `json:"wireless_encryption,omitempty"`

	// Status and health
	Status        string    `json:"status,omitempty"`
	LastSeen      time.Time `json:"last_seen,omitempty"`
	UptimeSeconds int64     `json:"uptime_seconds,omitempty"`
}

// Submit submits network hardware information for a server
func (s *NetworkHardwareService) Submit(ctx context.Context, serverUUID string, interfaces []NetworkHardwareInfo) (*StandardResponse, error) {
	if serverUUID == "" {
		return nil, fmt.Errorf("server UUID is required")
	}

	req := NetworkHardwareRequest{
		ServerUUID: serverUUID,
		Interfaces: interfaces,
	}

	var resp StandardResponse
	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v2/server/%s/hardware/network", serverUUID),
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return &resp, nil
}