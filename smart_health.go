package nexmonyx

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// SmartHealthService handles SMART health metrics operations
type SmartHealthService struct {
	client *Client
}

// SmartHealthMetricsSubmission represents the payload for submitting SMART health metrics
type SmartHealthMetricsSubmission struct {
	ServerUUID uuid.UUID                   `json:"server_uuid"`
	Timestamp  time.Time                   `json:"timestamp"`
	Devices    []SmartHealthDeviceMetrics  `json:"devices"`
}

// SmartHealthDeviceMetrics represents SMART health data for a single device
type SmartHealthDeviceMetrics struct {
	DeviceName      string  `json:"device_name"`
	DeviceModel     string  `json:"device_model"`
	DeviceSerial    string  `json:"device_serial"`
	DeviceType      string  `json:"device_type"` // 'ata', 'scsi', 'nvme', 'usb'
	DeviceInterface *string `json:"device_interface,omitempty"`
	FirmwareVersion *string `json:"firmware_version,omitempty"`
	CapacityBytes   *int64  `json:"capacity_bytes,omitempty"`

	// Overall health status
	OverallHealthStatus         string   `json:"overall_health_status"` // 'PASSED', 'FAILED', 'UNKNOWN'
	HealthPercentage           *float64 `json:"health_percentage,omitempty"`
	PredictedFailure           bool     `json:"predicted_failure"`
	FailurePredictionConfidence *float64 `json:"failure_prediction_confidence,omitempty"`

	// Critical SMART attributes (common across drive types)
	TemperatureCelsius      *int   `json:"temperature_celsius,omitempty"`
	PowerOnHours           *int64 `json:"power_on_hours,omitempty"`
	PowerCycleCount        *int64 `json:"power_cycle_count,omitempty"`
	ReallocatedSectorsCount *int64 `json:"reallocated_sectors_count,omitempty"`
	PendingSectorsCount    *int64 `json:"pending_sectors_count,omitempty"`
	UncorrectableErrorsCount *int64 `json:"uncorrectable_errors_count,omitempty"`

	// SSD-specific attributes
	WearLevelingCount    *int64   `json:"wear_leveling_count,omitempty"`
	ProgramFailCount     *int64   `json:"program_fail_count,omitempty"`
	EraseFailCount       *int64   `json:"erase_fail_count,omitempty"`
	TotalLBAsWritten     *int64   `json:"total_lbas_written,omitempty"`
	TotalLBAsRead        *int64   `json:"total_lbas_read,omitempty"`
	SSDLifeLeftPercent   *float64 `json:"ssd_life_left_percent,omitempty"`

	// HDD-specific attributes
	SpinRetryCount           *int64 `json:"spin_retry_count,omitempty"`
	CalibrationRetryCount    *int64 `json:"calibration_retry_count,omitempty"`
	HeadFlyingHours         *int64 `json:"head_flying_hours,omitempty"`
	LoadUnloadCycles        *int64 `json:"load_unload_cycles,omitempty"`
	SeekErrorRate           *int64 `json:"seek_error_rate,omitempty"`

	// NVMe-specific attributes
	NVMeCriticalWarning                  *int     `json:"nvme_critical_warning,omitempty"`
	NVMeCompositeTemperature            *int     `json:"nvme_composite_temperature,omitempty"`
	NVMeAvailableSparePercent           *float64 `json:"nvme_available_spare_percent,omitempty"`
	NVMeAvailableSpareThresholdPercent  *float64 `json:"nvme_available_spare_threshold_percent,omitempty"`
	NVMePercentageUsed                  *float64 `json:"nvme_percentage_used,omitempty"`
	NVMeDataUnitsRead                   *int64   `json:"nvme_data_units_read,omitempty"`
	NVMeDataUnitsWritten                *int64   `json:"nvme_data_units_written,omitempty"`
	NVMeHostReads                       *int64   `json:"nvme_host_reads,omitempty"`
	NVMeHostWrites                      *int64   `json:"nvme_host_writes,omitempty"`
	NVMeControllerBusyTime              *int64   `json:"nvme_controller_busy_time,omitempty"`
	NVMePowerCycles                     *int64   `json:"nvme_power_cycles,omitempty"`
	NVMePowerOnHours                    *int64   `json:"nvme_power_on_hours,omitempty"`
	NVMeUnsafeShutdowns                 *int64   `json:"nvme_unsafe_shutdowns,omitempty"`
	NVMeMediaErrors                     *int64   `json:"nvme_media_errors,omitempty"`
	NVMeErrorLogEntries                 *int64   `json:"nvme_error_log_entries,omitempty"`

	// Self-test results
	LastSelfTestResult               *string    `json:"last_self_test_result,omitempty"`
	LastSelfTestTimestamp           *time.Time `json:"last_self_test_timestamp,omitempty"`
	ShortSelfTestPollingMinutes     *int       `json:"short_self_test_polling_minutes,omitempty"`
	ExtendedSelfTestPollingMinutes  *int       `json:"extended_self_test_polling_minutes,omitempty"`

	// Performance degradation indicators
	ReadErrorRate          *int64 `json:"read_error_rate,omitempty"`
	ThroughputPerformance  *int64 `json:"throughput_performance,omitempty"`
	SeekTimePerformance    *int64 `json:"seek_time_performance,omitempty"`

	// Raw SMART data (JSON for extensibility)
	RawSmartAttributes map[string]interface{} `json:"raw_smart_attributes,omitempty"`

	// Health alerts and warnings
	CriticalWarnings []string `json:"critical_warnings,omitempty"`
	WarningMessages  []string `json:"warning_messages,omitempty"`
}

// Submit submits SMART health metrics to the API
func (s *SmartHealthService) Submit(ctx context.Context, submission *SmartHealthMetricsSubmission) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v2/metrics/smart-health",
		Body:   submission,
		Result: &resp,
	})
	return err
}