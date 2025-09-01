package nexmonyx

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// DiskIOService handles disk I/O metrics operations
type DiskIOService struct {
	client *Client
}

// DiskIOMetricsSubmission represents the payload for submitting disk I/O metrics
type DiskIOMetricsSubmission struct {
	ServerUUID uuid.UUID                `json:"server_uuid"`
	Timestamp  time.Time                `json:"timestamp"`
	Devices    []DiskIODeviceMetrics    `json:"devices"`
}

// DiskIODeviceMetrics represents I/O metrics for a single device
type DiskIODeviceMetrics struct {
	DeviceName string `json:"device_name"`
	
	// Core I/O counters
	ReadsCompleted  int64 `json:"reads_completed"`
	ReadsMerged     int64 `json:"reads_merged"`
	SectorsRead     int64 `json:"sectors_read"`
	ReadTimeMs      int64 `json:"read_time_ms"`
	
	WritesCompleted int64 `json:"writes_completed"`
	WritesMerged    int64 `json:"writes_merged"`
	SectorsWritten  int64 `json:"sectors_written"`
	WriteTimeMs     int64 `json:"write_time_ms"`
	
	IOInProgress     int   `json:"io_in_progress"`
	IOTimeMs         int64 `json:"io_time_ms"`
	WeightedIOTimeMs int64 `json:"weighted_io_time_ms"`
	
	// Optional calculated metrics
	ReadBytesPerSec    *int64   `json:"read_bytes_per_sec,omitempty"`
	WriteBytesPerSec   *int64   `json:"write_bytes_per_sec,omitempty"`
	ReadOpsPerSec      *float64 `json:"read_ops_per_sec,omitempty"`
	WriteOpsPerSec     *float64 `json:"write_ops_per_sec,omitempty"`
	UtilizationPercent *float64 `json:"utilization_percent,omitempty"`
	
	// Device metadata
	DeviceType   *string `json:"device_type,omitempty"`
	DeviceSize   *int64  `json:"device_size_bytes,omitempty"`
	DeviceModel  *string `json:"device_model,omitempty"`
	DeviceSerial *string `json:"device_serial,omitempty"`
}

// Submit submits disk I/O metrics to the API
func (s *DiskIOService) Submit(ctx context.Context, submission *DiskIOMetricsSubmission) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v2/metrics/disk-io",
		Body:   submission,
		Result: &resp,
	})
	return err
}