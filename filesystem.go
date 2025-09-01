package nexmonyx

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// FilesystemService handles advanced filesystem metrics operations
type FilesystemService struct {
	client *Client
}

// FilesystemMetricsSubmission represents the payload for submitting filesystem metrics
type FilesystemMetricsSubmission struct {
	ServerUUID  uuid.UUID                     `json:"server_uuid"`
	Timestamp   time.Time                     `json:"timestamp"`
	Filesystems []FilesystemMetricsData       `json:"filesystems"`
}

// FilesystemMetricsData represents filesystem metrics for a single filesystem
type FilesystemMetricsData struct {
	FilesystemName string  `json:"filesystem_name"`
	FilesystemType string  `json:"filesystem_type"` // 'zfs', 'lvm', 'mdraid', 'btrfs', 'ext4', 'xfs', 'ntfs'
	MountPoint     *string `json:"mount_point,omitempty"`
	DevicePath     *string `json:"device_path,omitempty"`

	// Storage capacity metrics
	TotalBytes     *int64   `json:"total_bytes,omitempty"`
	UsedBytes      *int64   `json:"used_bytes,omitempty"`
	AvailableBytes *int64   `json:"available_bytes,omitempty"`
	ReservedBytes  *int64   `json:"reserved_bytes,omitempty"`
	UsagePercent   *float64 `json:"usage_percent,omitempty"`

	// ZFS-specific metrics
	ZFSPoolName             *string    `json:"zfs_pool_name,omitempty"`
	ZFSDatasetName          *string    `json:"zfs_dataset_name,omitempty"`
	ZFSPoolHealth           *string    `json:"zfs_pool_health,omitempty"` // 'ONLINE', 'DEGRADED', 'FAULTED', 'OFFLINE'
	ZFSPoolState            *string    `json:"zfs_pool_state,omitempty"`  // 'ACTIVE', 'EXPORTED', 'DESTROYED', 'SPARE'
	ZFSCompressionRatio     *float64   `json:"zfs_compression_ratio,omitempty"`
	ZFSDedupRatio           *float64   `json:"zfs_dedup_ratio,omitempty"`
	ZFSFragmentationPercent *float64   `json:"zfs_fragmentation_percent,omitempty"`
	ZFSAllocatedBytes       *int64     `json:"zfs_allocated_bytes,omitempty"`
	ZFSReferencedBytes      *int64     `json:"zfs_referenced_bytes,omitempty"`
	ZFSSnapshotsCount       *int       `json:"zfs_snapshots_count,omitempty"`
	ZFSSnapshotSizeBytes    *int64     `json:"zfs_snapshot_size_bytes,omitempty"`

	// ZFS pool operations
	ZFSScrubState                    *string    `json:"zfs_scrub_state,omitempty"`
	ZFSScrubPercentComplete          *float64   `json:"zfs_scrub_percent_complete,omitempty"`
	ZFSScrubLastRun                  *time.Time `json:"zfs_scrub_last_run,omitempty"`
	ZFSResilverState                 *string    `json:"zfs_resilver_state,omitempty"`
	ZFSResilverPercentComplete       *float64   `json:"zfs_resilver_percent_complete,omitempty"`
	ZFSResilverEstimatedCompletion   *time.Time `json:"zfs_resilver_estimated_completion,omitempty"`

	// ZFS error counters
	ZFSReadErrors     *int64 `json:"zfs_read_errors,omitempty"`
	ZFSWriteErrors    *int64 `json:"zfs_write_errors,omitempty"`
	ZFSChecksumErrors *int64 `json:"zfs_checksum_errors,omitempty"`

	// LVM-specific metrics
	LVMVGName      *string `json:"lvm_vg_name,omitempty"`
	LVMLVName      *string `json:"lvm_lv_name,omitempty"`
	LVMPVCount     *int    `json:"lvm_pv_count,omitempty"`
	LVMLVCount     *int    `json:"lvm_lv_count,omitempty"`
	LVMPESizeBytes *int64  `json:"lvm_pe_size_bytes,omitempty"`
	LVMTotalPE     *int    `json:"lvm_total_pe,omitempty"`
	LVMFreePE      *int    `json:"lvm_free_pe,omitempty"`
	LVMAllocatedPE *int    `json:"lvm_allocated_pe,omitempty"`
	LVMVGStatus    *string `json:"lvm_vg_status,omitempty"`
	LVMLVStatus    *string `json:"lvm_lv_status,omitempty"`
	LVMAttributes  *string `json:"lvm_attributes,omitempty"`

	// RAID-specific metrics
	RAIDLevel         *string  `json:"raid_level,omitempty"` // 'raid0', 'raid1', 'raid4', 'raid5', 'raid6', 'raid10'
	RAIDDeviceName    *string  `json:"raid_device_name,omitempty"`
	RAIDState         *string  `json:"raid_state,omitempty"` // 'active', 'inactive', 'clean', 'degraded'
	RAIDTotalDevices  *int     `json:"raid_total_devices,omitempty"`
	RAIDActiveDevices *int     `json:"raid_active_devices,omitempty"`
	RAIDSpareDevices  *int     `json:"raid_spare_devices,omitempty"`
	RAIDFailedDevices *int     `json:"raid_failed_devices,omitempty"`
	RAIDSyncPercent   *float64 `json:"raid_sync_percent,omitempty"`
	RAIDReshapePercent *float64 `json:"raid_reshape_percent,omitempty"`
	RAIDChunkSizeKB   *int     `json:"raid_chunk_size_kb,omitempty"`

	// BTRFS-specific metrics
	BTRFSFilesystemUUID  *string  `json:"btrfs_filesystem_uuid,omitempty"`
	BTRFSTotalDevices    *int     `json:"btrfs_total_devices,omitempty"`
	BTRFSRAIDType        *string  `json:"btrfs_raid_type,omitempty"`
	BTRFSAllocationRatio *float64 `json:"btrfs_allocation_ratio,omitempty"`
	BTRFSDataRatio       *float64 `json:"btrfs_data_ratio,omitempty"`
	BTRFSMetadataRatio   *float64 `json:"btrfs_metadata_ratio,omitempty"`

	// Performance metrics
	ReadOpsPerSec     *float64 `json:"read_ops_per_sec,omitempty"`
	WriteOpsPerSec    *float64 `json:"write_ops_per_sec,omitempty"`
	ReadBytesPerSec   *int64   `json:"read_bytes_per_sec,omitempty"`
	WriteBytesPerSec  *int64   `json:"write_bytes_per_sec,omitempty"`
	AvgReadLatencyMs  *float64 `json:"avg_read_latency_ms,omitempty"`
	AvgWriteLatencyMs *float64 `json:"avg_write_latency_ms,omitempty"`
	QueueDepth        *float64 `json:"queue_depth,omitempty"`

	// Health and status indicators
	OverallHealth string   `json:"overall_health"` // 'HEALTHY', 'WARNING', 'CRITICAL', 'UNKNOWN'
	HealthScore   *float64 `json:"health_score,omitempty"`
	WarningCount  int      `json:"warning_count"`
	ErrorCount    int      `json:"error_count"`

	// Alerts and warnings
	CriticalAlerts  []string `json:"critical_alerts,omitempty"`
	WarningMessages []string `json:"warning_messages,omitempty"`

	// Raw metrics for extensibility
	RawMetrics map[string]interface{} `json:"raw_metrics,omitempty"`
}

// Submit submits filesystem metrics to the API
func (s *FilesystemService) Submit(ctx context.Context, submission *FilesystemMetricsSubmission) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v2/metrics/filesystem",
		Body:   submission,
		Result: &resp,
	})
	return err
}

// SubmitZFS is a convenience method for submitting ZFS-specific metrics
func (s *FilesystemService) SubmitZFS(ctx context.Context, serverUUID uuid.UUID, zfsMetrics []ZFSPoolMetrics) error {
	submission := &FilesystemMetricsSubmission{
		ServerUUID:  serverUUID,
		Timestamp:   time.Now(),
		Filesystems: make([]FilesystemMetricsData, 0, len(zfsMetrics)),
	}

	for _, pool := range zfsMetrics {
		filesystemData := FilesystemMetricsData{
			FilesystemName:          pool.PoolName,
			FilesystemType:          "zfs",
			MountPoint:              pool.MountPoint,
			TotalBytes:              pool.TotalBytes,
			UsedBytes:               pool.UsedBytes,
			AvailableBytes:          pool.AvailableBytes,
			UsagePercent:            pool.UsagePercent,
			ZFSPoolName:             &pool.PoolName,
			ZFSPoolHealth:           pool.Health,
			ZFSPoolState:            pool.State,
			ZFSCompressionRatio:     pool.CompressionRatio,
			ZFSDedupRatio:           pool.DedupRatio,
			ZFSFragmentationPercent: pool.FragmentationPercent,
			ZFSAllocatedBytes:       pool.AllocatedBytes,
			ZFSReferencedBytes:      pool.ReferencedBytes,
			ZFSSnapshotsCount:       pool.SnapshotsCount,
			ZFSSnapshotSizeBytes:    pool.SnapshotSizeBytes,
			ZFSScrubState:           pool.ScrubState,
			ZFSScrubPercentComplete: pool.ScrubPercentComplete,
			ZFSReadErrors:           pool.ReadErrors,
			ZFSWriteErrors:          pool.WriteErrors,
			ZFSChecksumErrors:       pool.ChecksumErrors,
			OverallHealth:           pool.OverallHealth,
			HealthScore:             pool.HealthScore,
			WarningCount:            pool.WarningCount,
			ErrorCount:              pool.ErrorCount,
		}
		submission.Filesystems = append(submission.Filesystems, filesystemData)
	}

	return s.Submit(ctx, submission)
}

// SubmitRAID is a convenience method for submitting RAID-specific metrics
func (s *FilesystemService) SubmitRAID(ctx context.Context, serverUUID uuid.UUID, raidMetrics []RAIDArrayMetrics) error {
	submission := &FilesystemMetricsSubmission{
		ServerUUID:  serverUUID,
		Timestamp:   time.Now(),
		Filesystems: make([]FilesystemMetricsData, 0, len(raidMetrics)),
	}

	for _, raid := range raidMetrics {
		filesystemData := FilesystemMetricsData{
			FilesystemName:    raid.DeviceName,
			FilesystemType:    "mdraid",
			MountPoint:        raid.MountPoint,
			TotalBytes:        raid.TotalBytes,
			UsedBytes:         raid.UsedBytes,
			AvailableBytes:    raid.AvailableBytes,
			UsagePercent:      raid.UsagePercent,
			RAIDLevel:         raid.Level,
			RAIDDeviceName:    &raid.DeviceName,
			RAIDState:         raid.State,
			RAIDTotalDevices:  raid.TotalDevices,
			RAIDActiveDevices: raid.ActiveDevices,
			RAIDSpareDevices:  raid.SpareDevices,
			RAIDFailedDevices: raid.FailedDevices,
			RAIDSyncPercent:   raid.SyncPercent,
			RAIDChunkSizeKB:   raid.ChunkSizeKB,
			OverallHealth:     raid.OverallHealth,
			HealthScore:       raid.HealthScore,
			WarningCount:      raid.WarningCount,
			ErrorCount:        raid.ErrorCount,
		}
		submission.Filesystems = append(submission.Filesystems, filesystemData)
	}

	return s.Submit(ctx, submission)
}

// SubmitLVM is a convenience method for submitting LVM-specific metrics
func (s *FilesystemService) SubmitLVM(ctx context.Context, serverUUID uuid.UUID, lvmMetrics []LVMVolumeMetrics) error {
	submission := &FilesystemMetricsSubmission{
		ServerUUID:  serverUUID,
		Timestamp:   time.Now(),
		Filesystems: make([]FilesystemMetricsData, 0, len(lvmMetrics)),
	}

	for _, lvm := range lvmMetrics {
		filesystemData := FilesystemMetricsData{
			FilesystemName: lvm.LogicalVolumeName,
			FilesystemType: "lvm",
			MountPoint:     lvm.MountPoint,
			DevicePath:     lvm.DevicePath,
			TotalBytes:     lvm.TotalBytes,
			UsedBytes:      lvm.UsedBytes,
			AvailableBytes: lvm.AvailableBytes,
			UsagePercent:   lvm.UsagePercent,
			LVMVGName:      &lvm.VolumeGroupName,
			LVMLVName:      &lvm.LogicalVolumeName,
			LVMPVCount:     lvm.PhysicalVolumeCount,
			LVMLVCount:     lvm.LogicalVolumeCount,
			LVMPESizeBytes: lvm.PhysicalExtentSize,
			LVMTotalPE:     lvm.TotalPhysicalExtents,
			LVMFreePE:      lvm.FreePhysicalExtents,
			LVMAllocatedPE: lvm.AllocatedPhysicalExtents,
			LVMVGStatus:    lvm.VolumeGroupStatus,
			LVMLVStatus:    lvm.LogicalVolumeStatus,
			LVMAttributes:  lvm.Attributes,
			OverallHealth:  lvm.OverallHealth,
			HealthScore:    lvm.HealthScore,
			WarningCount:   lvm.WarningCount,
			ErrorCount:     lvm.ErrorCount,
		}
		submission.Filesystems = append(submission.Filesystems, filesystemData)
	}

	return s.Submit(ctx, submission)
}

// Convenience types for specific storage technologies
type ZFSPoolMetrics struct {
	PoolName                string
	MountPoint              *string
	TotalBytes              *int64
	UsedBytes               *int64
	AvailableBytes          *int64
	UsagePercent            *float64
	Health                  *string
	State                   *string
	CompressionRatio        *float64
	DedupRatio              *float64
	FragmentationPercent    *float64
	AllocatedBytes          *int64
	ReferencedBytes         *int64
	SnapshotsCount          *int
	SnapshotSizeBytes       *int64
	ScrubState              *string
	ScrubPercentComplete    *float64
	ReadErrors              *int64
	WriteErrors             *int64
	ChecksumErrors          *int64
	OverallHealth           string
	HealthScore             *float64
	WarningCount            int
	ErrorCount              int
}

type RAIDArrayMetrics struct {
	DeviceName    string
	MountPoint    *string
	TotalBytes    *int64
	UsedBytes     *int64
	AvailableBytes *int64
	UsagePercent  *float64
	Level         *string
	State         *string
	TotalDevices  *int
	ActiveDevices *int
	SpareDevices  *int
	FailedDevices *int
	SyncPercent   *float64
	ChunkSizeKB   *int
	OverallHealth string
	HealthScore   *float64
	WarningCount  int
	ErrorCount    int
}

type LVMVolumeMetrics struct {
	LogicalVolumeName         string
	VolumeGroupName          string
	MountPoint               *string
	DevicePath               *string
	TotalBytes               *int64
	UsedBytes                *int64
	AvailableBytes           *int64
	UsagePercent             *float64
	PhysicalVolumeCount      *int
	LogicalVolumeCount       *int
	PhysicalExtentSize       *int64
	TotalPhysicalExtents     *int
	FreePhysicalExtents      *int
	AllocatedPhysicalExtents *int
	VolumeGroupStatus        *string
	LogicalVolumeStatus      *string
	Attributes               *string
	OverallHealth            string
	HealthScore              *float64
	WarningCount             int
	ErrorCount               int
}