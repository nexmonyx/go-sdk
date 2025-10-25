package nexmonyx

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAggregateDiskUsage(t *testing.T) {
	tests := []struct {
		name     string
		disks    []DiskMetrics
		expected *DiskUsageAggregate
	}{
		{
			name: "single filesystem",
			disks: []DiskMetrics{
				{
					Device:       "/dev/sda1",
					Mountpoint:   "/",
					Filesystem:   "ext4",
					TotalBytes:   1000000000,
					UsedBytes:    750000000,
					FreeBytes:    250000000,
					UsagePercent: 75.0,
				},
			},
			expected: &DiskUsageAggregate{
				TotalBytes:      1000000000,
				UsedBytes:       750000000,
				FreeBytes:       250000000,
				UsedPercent:     75.0,
				FilesystemCount: 1,
				LargestMount:    "/",
				CriticalMounts:  []string{},
			},
		},
		{
			name: "multiple filesystems with critical mount",
			disks: []DiskMetrics{
				{
					Device:       "/dev/sda1",
					Mountpoint:   "/",
					Filesystem:   "ext4",
					TotalBytes:   1000000000,
					UsedBytes:    750000000,
					FreeBytes:    250000000,
					UsagePercent: 75.0,
				},
				{
					Device:       "/dev/sda2",
					Mountpoint:   "/var/log",
					Filesystem:   "ext4",
					TotalBytes:   500000000,
					UsedBytes:    475000000,
					FreeBytes:    25000000,
					UsagePercent: 95.0,
				},
				{
					Device:       "/dev/sda3",
					Mountpoint:   "/home",
					Filesystem:   "xfs",
					TotalBytes:   2000000000,
					UsedBytes:    1000000000,
					FreeBytes:    1000000000,
					UsagePercent: 50.0,
				},
			},
			expected: &DiskUsageAggregate{
				TotalBytes:      3500000000,
				UsedBytes:       2225000000,
				FreeBytes:       1275000000,
				UsedPercent:     63.57142857142857, // 2225000000 / 3500000000 * 100
				FilesystemCount: 3,
				LargestMount:    "/home",
				CriticalMounts:  []string{"/var/log"},
			},
		},
		{
			name: "mixed filesystems with exclusions",
			disks: []DiskMetrics{
				{
					Device:       "/dev/sda1",
					Mountpoint:   "/",
					Filesystem:   "ext4",
					TotalBytes:   1000000000,
					UsedBytes:    750000000,
					FreeBytes:    250000000,
					UsagePercent: 75.0,
				},
				{
					Device:       "tmpfs",
					Mountpoint:   "/tmp",
					Filesystem:   "tmpfs",
					TotalBytes:   100000000,
					UsedBytes:    50000000,
					FreeBytes:    50000000,
					UsagePercent: 50.0,
				},
				{
					Device:       "proc",
					Mountpoint:   "/proc",
					Filesystem:   "proc",
					TotalBytes:   0,
					UsedBytes:    0,
					FreeBytes:    0,
					UsagePercent: 0,
				},
			},
			expected: &DiskUsageAggregate{
				TotalBytes:      1000000000,
				UsedBytes:       750000000,
				FreeBytes:       250000000,
				UsedPercent:     75.0,
				FilesystemCount: 1,
				LargestMount:    "/",
				CriticalMounts:  []string{},
			},
		},
		{
			name:  "empty disk list",
			disks: []DiskMetrics{},
			expected: &DiskUsageAggregate{
				TotalBytes:      0,
				UsedBytes:       0,
				FreeBytes:       0,
				UsedPercent:     0,
				FilesystemCount: 0,
				LargestMount:    "",
				CriticalMounts:  []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AggregateDiskUsage(tt.disks)

			assert.Equal(t, tt.expected.TotalBytes, result.TotalBytes)
			assert.Equal(t, tt.expected.UsedBytes, result.UsedBytes)
			assert.Equal(t, tt.expected.FreeBytes, result.FreeBytes)
			assert.InDelta(t, tt.expected.UsedPercent, result.UsedPercent, 0.001)
			assert.Equal(t, tt.expected.FilesystemCount, result.FilesystemCount)
			assert.Equal(t, tt.expected.LargestMount, result.LargestMount)
			assert.Equal(t, tt.expected.CriticalMounts, result.CriticalMounts)

			// Verify CalculatedAt is a valid RFC3339 timestamp
			_, err := time.Parse(time.RFC3339, result.CalculatedAt)
			assert.NoError(t, err)
		})
	}
}

func TestShouldIncludeInAggregation(t *testing.T) {
	tests := []struct {
		name     string
		disk     DiskMetrics
		expected bool
	}{
		{
			name: "ext4 filesystem should be included",
			disk: DiskMetrics{
				Device:     "/dev/sda1",
				Mountpoint: "/",
				Filesystem: "ext4",
			},
			expected: true,
		},
		{
			name: "tmpfs should be excluded",
			disk: DiskMetrics{
				Device:     "tmpfs",
				Mountpoint: "/tmp",
				Filesystem: "tmpfs",
			},
			expected: false,
		},
		{
			name: "proc filesystem should be excluded",
			disk: DiskMetrics{
				Device:     "proc",
				Mountpoint: "/proc",
				Filesystem: "proc",
			},
			expected: false,
		},
		{
			name: "docker overlay should be excluded by mount prefix",
			disk: DiskMetrics{
				Device:     "overlay",
				Mountpoint: "/var/lib/docker/overlay2/abc123",
				Filesystem: "overlay",
			},
			expected: false,
		},
		{
			name: "snap mount should be excluded",
			disk: DiskMetrics{
				Device:     "/dev/loop0",
				Mountpoint: "/snap/core/123",
				Filesystem: "squashfs",
			},
			expected: false,
		},
		{
			name: "xfs filesystem should be included",
			disk: DiskMetrics{
				Device:     "/dev/sdb1",
				Mountpoint: "/home",
				Filesystem: "xfs",
			},
			expected: true,
		},
		{
			name: "nfs filesystem should be included",
			disk: DiskMetrics{
				Device:     "server:/export",
				Mountpoint: "/mnt/nfs",
				Filesystem: "nfs4",
			},
			expected: true,
		},
		{
			name: "unknown filesystem should be excluded",
			disk: DiskMetrics{
				Device:     "/dev/sdc1",
				Mountpoint: "/mnt/unknown",
				Filesystem: "unknownfs",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldIncludeInAggregation(tt.disk)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAggregateDiskUsageFromRequest(t *testing.T) {
	t.Run("valid request with disks", func(t *testing.T) {
		request := &ComprehensiveMetricsRequest{
			Disks: []DiskMetrics{
				{
					Device:       "/dev/sda1",
					Mountpoint:   "/",
					Filesystem:   "ext4",
					TotalBytes:   1000000000,
					UsedBytes:    750000000,
					FreeBytes:    250000000,
					UsagePercent: 75.0,
				},
			},
		}

		result := AggregateDiskUsageFromRequest(request)

		assert.Equal(t, uint64(1000000000), result.TotalBytes)
		assert.Equal(t, uint64(750000000), result.UsedBytes)
		assert.Equal(t, uint64(250000000), result.FreeBytes)
		assert.Equal(t, 1, result.FilesystemCount)
	})

	t.Run("nil request", func(t *testing.T) {
		result := AggregateDiskUsageFromRequest(nil)

		assert.Equal(t, uint64(0), result.TotalBytes)
		assert.Equal(t, uint64(0), result.UsedBytes)
		assert.Equal(t, uint64(0), result.FreeBytes)
		assert.Equal(t, 0, result.FilesystemCount)
		assert.Equal(t, "", result.LargestMount)
		assert.Equal(t, []string{}, result.CriticalMounts)
	})

	t.Run("request with nil disks", func(t *testing.T) {
		request := &ComprehensiveMetricsRequest{
			Disks: nil,
		}

		result := AggregateDiskUsageFromRequest(request)

		assert.Equal(t, uint64(0), result.TotalBytes)
		assert.Equal(t, uint64(0), result.UsedBytes)
		assert.Equal(t, uint64(0), result.FreeBytes)
		assert.Equal(t, 0, result.FilesystemCount)
	})
}

func TestValidateDiskUsageAggregate(t *testing.T) {
	t.Run("valid aggregate", func(t *testing.T) {
		aggregate := &DiskUsageAggregate{
			TotalBytes:      1000000000,
			UsedBytes:       750000000,
			FreeBytes:       250000000,
			UsedPercent:     75.0,
			FilesystemCount: 1,
			LargestMount:    "/",
			CriticalMounts:  []string{},
			CalculatedAt:    time.Now().UTC().Format(time.RFC3339),
		}

		assert.True(t, ValidateDiskUsageAggregate(aggregate))
	})

	t.Run("nil aggregate", func(t *testing.T) {
		assert.False(t, ValidateDiskUsageAggregate(nil))
	})

	t.Run("inconsistent total bytes", func(t *testing.T) {
		aggregate := &DiskUsageAggregate{
			TotalBytes:      1000000000,
			UsedBytes:       750000000,
			FreeBytes:       300000000, // Should be 250000000
			UsedPercent:     75.0,
			FilesystemCount: 1,
			CalculatedAt:    time.Now().UTC().Format(time.RFC3339),
		}

		assert.False(t, ValidateDiskUsageAggregate(aggregate))
	})

	t.Run("invalid usage percentage", func(t *testing.T) {
		aggregate := &DiskUsageAggregate{
			TotalBytes:      1000000000,
			UsedBytes:       750000000,
			FreeBytes:       250000000,
			UsedPercent:     105.0, // Invalid percentage > 100
			FilesystemCount: 1,
			CalculatedAt:    time.Now().UTC().Format(time.RFC3339),
		}

		assert.False(t, ValidateDiskUsageAggregate(aggregate))
	})

	t.Run("negative filesystem count", func(t *testing.T) {
		aggregate := &DiskUsageAggregate{
			TotalBytes:      1000000000,
			UsedBytes:       750000000,
			FreeBytes:       250000000,
			UsedPercent:     75.0,
			FilesystemCount: -1, // Invalid negative count
			CalculatedAt:    time.Now().UTC().Format(time.RFC3339),
		}

		assert.False(t, ValidateDiskUsageAggregate(aggregate))
	})

	t.Run("invalid timestamp format", func(t *testing.T) {
		aggregate := &DiskUsageAggregate{
			TotalBytes:      1000000000,
			UsedBytes:       750000000,
			FreeBytes:       250000000,
			UsedPercent:     75.0,
			FilesystemCount: 1,
			CalculatedAt:    "invalid-timestamp",
		}

		assert.False(t, ValidateDiskUsageAggregate(aggregate))
	})

	t.Run("inconsistent usage percentage calculation", func(t *testing.T) {
		aggregate := &DiskUsageAggregate{
			TotalBytes:      1000000000,
			UsedBytes:       750000000,
			FreeBytes:       250000000,
			UsedPercent:     50.0, // Should be 75.0, off by more than tolerance (1%)
			FilesystemCount: 1,
			CalculatedAt:    time.Now().UTC().Format(time.RFC3339),
		}

		assert.False(t, ValidateDiskUsageAggregate(aggregate))
	})

	t.Run("empty timestamp allowed", func(t *testing.T) {
		aggregate := &DiskUsageAggregate{
			TotalBytes:      1000000000,
			UsedBytes:       750000000,
			FreeBytes:       250000000,
			UsedPercent:     75.0,
			FilesystemCount: 1,
			CalculatedAt:    "", // Empty timestamp should be allowed
		}

		assert.True(t, ValidateDiskUsageAggregate(aggregate))
	})

	t.Run("negative usage percentage", func(t *testing.T) {
		aggregate := &DiskUsageAggregate{
			TotalBytes:      1000000000,
			UsedBytes:       750000000,
			FreeBytes:       250000000,
			UsedPercent:     -5.0, // Invalid negative percentage
			FilesystemCount: 1,
			CalculatedAt:    time.Now().UTC().Format(time.RFC3339),
		}

		assert.False(t, ValidateDiskUsageAggregate(aggregate))
	})

	t.Run("zero total bytes with zero usage", func(t *testing.T) {
		aggregate := &DiskUsageAggregate{
			TotalBytes:      0,
			UsedBytes:       0,
			FreeBytes:       0,
			UsedPercent:     0,
			FilesystemCount: 0,
			CalculatedAt:    time.Now().UTC().Format(time.RFC3339),
		}

		assert.True(t, ValidateDiskUsageAggregate(aggregate))
	})

	t.Run("zero total bytes with invalid percentage over 100", func(t *testing.T) {
		// This test covers the range check that happens when TotalBytes is 0
		// (which skips the percentage consistency check)
		aggregate := &DiskUsageAggregate{
			TotalBytes:      0,
			UsedBytes:       0,
			FreeBytes:       0,
			UsedPercent:     150.0, // Invalid: > 100
			FilesystemCount: 0,
			CalculatedAt:    time.Now().UTC().Format(time.RFC3339),
		}

		assert.False(t, ValidateDiskUsageAggregate(aggregate))
	})

	t.Run("zero total bytes with invalid negative percentage", func(t *testing.T) {
		// This test covers the range check for negative percentages when TotalBytes is 0
		aggregate := &DiskUsageAggregate{
			TotalBytes:      0,
			UsedBytes:       0,
			FreeBytes:       0,
			UsedPercent:     -10.0, // Invalid: < 0
			FilesystemCount: 0,
			CalculatedAt:    time.Now().UTC().Format(time.RFC3339),
		}

		assert.False(t, ValidateDiskUsageAggregate(aggregate))
	})
}

func TestComprehensiveMetricsRequestSerialization(t *testing.T) {
	t.Run("serialization with DiskUsageAggregate", func(t *testing.T) {
		request := &ComprehensiveMetricsRequest{
			ServerUUID:  "test-server-uuid",
			CollectedAt: time.Now().UTC().Format(time.RFC3339),
			Disks: []DiskMetrics{
				{
					Device:       "/dev/sda1",
					Mountpoint:   "/",
					Filesystem:   "ext4",
					TotalBytes:   1000000000,
					UsedBytes:    750000000,
					FreeBytes:    250000000,
					UsagePercent: 75.0,
				},
			},
			DiskUsageAggregate: &DiskUsageAggregate{
				TotalBytes:      1000000000,
				UsedBytes:       750000000,
				FreeBytes:       250000000,
				UsedPercent:     75.0,
				FilesystemCount: 1,
				LargestMount:    "/",
				CriticalMounts:  []string{},
				CalculatedAt:    time.Now().UTC().Format(time.RFC3339),
			},
		}

		// Test JSON serialization
		jsonData, err := json.Marshal(request)
		require.NoError(t, err)
		assert.Contains(t, string(jsonData), "disk_usage_aggregate")

		// Test JSON deserialization
		var deserialized ComprehensiveMetricsRequest
		err = json.Unmarshal(jsonData, &deserialized)
		require.NoError(t, err)

		assert.Equal(t, request.ServerUUID, deserialized.ServerUUID)
		assert.Equal(t, request.CollectedAt, deserialized.CollectedAt)
		assert.Len(t, deserialized.Disks, 1)
		assert.NotNil(t, deserialized.DiskUsageAggregate)
		assert.Equal(t, request.DiskUsageAggregate.TotalBytes, deserialized.DiskUsageAggregate.TotalBytes)
		assert.Equal(t, request.DiskUsageAggregate.UsedBytes, deserialized.DiskUsageAggregate.UsedBytes)
		assert.Equal(t, request.DiskUsageAggregate.FilesystemCount, deserialized.DiskUsageAggregate.FilesystemCount)
	})

	t.Run("serialization without DiskUsageAggregate (backward compatibility)", func(t *testing.T) {
		request := &ComprehensiveMetricsRequest{
			ServerUUID:  "test-server-uuid",
			CollectedAt: time.Now().UTC().Format(time.RFC3339),
			Disks: []DiskMetrics{
				{
					Device:       "/dev/sda1",
					Mountpoint:   "/",
					Filesystem:   "ext4",
					TotalBytes:   1000000000,
					UsedBytes:    750000000,
					FreeBytes:    250000000,
					UsagePercent: 75.0,
				},
			},
			// DiskUsageAggregate is nil (not provided)
		}

		// Test JSON serialization
		jsonData, err := json.Marshal(request)
		require.NoError(t, err)
		
		// Should not contain disk_usage_aggregate field since it's nil and has omitempty
		assert.NotContains(t, string(jsonData), "disk_usage_aggregate")

		// Test JSON deserialization
		var deserialized ComprehensiveMetricsRequest
		err = json.Unmarshal(jsonData, &deserialized)
		require.NoError(t, err)

		assert.Equal(t, request.ServerUUID, deserialized.ServerUUID)
		assert.Equal(t, request.CollectedAt, deserialized.CollectedAt)
		assert.Len(t, deserialized.Disks, 1)
		assert.Nil(t, deserialized.DiskUsageAggregate)
	})

	t.Run("deserialization of old format (without DiskUsageAggregate)", func(t *testing.T) {
		// Simulate old JSON format without disk_usage_aggregate field
		oldFormatJSON := `{
			"server_uuid": "test-server-uuid",
			"collected_at": "2023-01-01T00:00:00Z",
			"disks": [
				{
					"device": "/dev/sda1",
					"mountpoint": "/",
					"filesystem": "ext4",
					"total_bytes": 1000000000,
					"used_bytes": 750000000,
					"free_bytes": 250000000,
					"usage_percent": 75.0
				}
			]
		}`

		var deserialized ComprehensiveMetricsRequest
		err := json.Unmarshal([]byte(oldFormatJSON), &deserialized)
		require.NoError(t, err)

		assert.Equal(t, "test-server-uuid", deserialized.ServerUUID)
		assert.Equal(t, "2023-01-01T00:00:00Z", deserialized.CollectedAt)
		assert.Len(t, deserialized.Disks, 1)
		assert.Nil(t, deserialized.DiskUsageAggregate) // Should be nil for old format
	})
}

func TestDiskUsageAggregateEdgeCases(t *testing.T) {
	t.Run("zero-sized filesystems", func(t *testing.T) {
		disks := []DiskMetrics{
			{
				Device:       "/dev/sda1",
				Mountpoint:   "/empty",
				Filesystem:   "ext4",
				TotalBytes:   0,
				UsedBytes:    0,
				FreeBytes:    0,
				UsagePercent: 0,
			},
		}

		result := AggregateDiskUsage(disks)
		assert.Equal(t, uint64(0), result.TotalBytes)
		assert.Equal(t, float64(0), result.UsedPercent)
		assert.Equal(t, 0, result.FilesystemCount) // Zero-sized disks are skipped, so count should be 0
	})

	t.Run("large filesystem sizes", func(t *testing.T) {
		disks := []DiskMetrics{
			{
				Device:       "/dev/sda1",
				Mountpoint:   "/large",
				Filesystem:   "ext4",
				TotalBytes:   9223372036854775807, // Max int64
				UsedBytes:    4611686018427387903, // ~50%
				FreeBytes:    4611686018427387904,
				UsagePercent: 50.0,
			},
		}

		result := AggregateDiskUsage(disks)
		assert.Equal(t, uint64(9223372036854775807), result.TotalBytes)
		assert.InDelta(t, 50.0, result.UsedPercent, 0.1)
		assert.Equal(t, 1, result.FilesystemCount)
	})

	t.Run("multiple critical mounts", func(t *testing.T) {
		disks := []DiskMetrics{
			{
				Device:       "/dev/sda1",
				Mountpoint:   "/critical1",
				Filesystem:   "ext4",
				TotalBytes:   1000000000,
				UsedBytes:    950000000,
				FreeBytes:    50000000,
				UsagePercent: 95.0,
			},
			{
				Device:       "/dev/sda2",
				Mountpoint:   "/critical2",
				Filesystem:   "ext4",
				TotalBytes:   1000000000,
				UsedBytes:    920000000,
				FreeBytes:    80000000,
				UsagePercent: 92.0,
			},
		}

		result := AggregateDiskUsage(disks)
		assert.Len(t, result.CriticalMounts, 2)
		assert.Contains(t, result.CriticalMounts, "/critical1")
		assert.Contains(t, result.CriticalMounts, "/critical2")
	})

	t.Run("negative byte values are treated as zero", func(t *testing.T) {
		disks := []DiskMetrics{
			{
				Device:       "/dev/sda1",
				Mountpoint:   "/",
				Filesystem:   "ext4",
				TotalBytes:   1000000000,
				UsedBytes:    -100000, // Negative value - should be treated as 0
				FreeBytes:    1000000000,
				UsagePercent: 0,
			},
		}

		result := AggregateDiskUsage(disks)
		assert.Equal(t, uint64(1000000000), result.TotalBytes)
		assert.Equal(t, uint64(0), result.UsedBytes) // Negative treated as zero
		assert.Equal(t, 1, result.FilesystemCount)
	})

	t.Run("all filesystems excluded by type", func(t *testing.T) {
		disks := []DiskMetrics{
			{
				Device:       "tmpfs",
				Mountpoint:   "/tmp",
				Filesystem:   "tmpfs",
				TotalBytes:   1000000000,
				UsedBytes:    500000000,
				FreeBytes:    500000000,
				UsagePercent: 50.0,
			},
			{
				Device:       "proc",
				Mountpoint:   "/proc",
				Filesystem:   "proc",
				TotalBytes:   0,
				UsedBytes:    0,
				FreeBytes:    0,
				UsagePercent: 0,
			},
		}

		result := AggregateDiskUsage(disks)
		assert.Equal(t, uint64(0), result.TotalBytes)
		assert.Equal(t, 0, result.FilesystemCount)
		assert.Equal(t, "", result.LargestMount)
		assert.Empty(t, result.CriticalMounts)
	})

	t.Run("all filesystems excluded by mount prefix", func(t *testing.T) {
		disks := []DiskMetrics{
			{
				Device:       "overlay",
				Mountpoint:   "/var/lib/docker/overlay2/abc123",
				Filesystem:   "overlay",
				TotalBytes:   5000000000,
				UsedBytes:    3000000000,
				FreeBytes:    2000000000,
				UsagePercent: 60.0,
			},
			{
				Device:       "/dev/loop0",
				Mountpoint:   "/snap/core/123",
				Filesystem:   "squashfs",
				TotalBytes:   200000000,
				UsedBytes:    200000000,
				FreeBytes:    0,
				UsagePercent: 100.0,
			},
		}

		result := AggregateDiskUsage(disks)
		assert.Equal(t, uint64(0), result.TotalBytes)
		assert.Equal(t, 0, result.FilesystemCount)
		assert.Empty(t, result.CriticalMounts)
	})

	t.Run("mixed included filesystem types", func(t *testing.T) {
		disks := []DiskMetrics{
			{
				Device:       "/dev/sda1",
				Mountpoint:   "/",
				Filesystem:   "ext4",
				TotalBytes:   1000000000,
				UsedBytes:    400000000,
				FreeBytes:    600000000,
				UsagePercent: 40.0,
			},
			{
				Device:       "/dev/sdb1",
				Mountpoint:   "/data",
				Filesystem:   "xfs",
				TotalBytes:   2000000000,
				UsedBytes:    800000000,
				FreeBytes:    1200000000,
				UsagePercent: 40.0,
			},
			{
				Device:       "/dev/sdc1",
				Mountpoint:   "/backup",
				Filesystem:   "btrfs",
				TotalBytes:   5000000000,
				UsedBytes:    2500000000,
				FreeBytes:    2500000000,
				UsagePercent: 50.0,
			},
			{
				Device:       "nfs-server:/export",
				Mountpoint:   "/mnt/nfs",
				Filesystem:   "nfs4",
				TotalBytes:   10000000000,
				UsedBytes:    3000000000,
				FreeBytes:    7000000000,
				UsagePercent: 30.0,
			},
		}

		result := AggregateDiskUsage(disks)
		assert.Equal(t, uint64(18000000000), result.TotalBytes)
		assert.Equal(t, uint64(6700000000), result.UsedBytes)
		assert.Equal(t, 4, result.FilesystemCount)
		assert.Equal(t, "/mnt/nfs", result.LargestMount) // Largest by capacity
		assert.InDelta(t, 37.22, result.UsedPercent, 0.1)
	})

	t.Run("critical threshold exactly at 90 percent", func(t *testing.T) {
		disks := []DiskMetrics{
			{
				Device:       "/dev/sda1",
				Mountpoint:   "/exactly90",
				Filesystem:   "ext4",
				TotalBytes:   1000000000,
				UsedBytes:    900000000,
				FreeBytes:    100000000,
				UsagePercent: 90.0,
			},
		}

		result := AggregateDiskUsage(disks)
		assert.Empty(t, result.CriticalMounts) // 90.0 is not > 90.0, so not critical
	})

	t.Run("critical threshold just above 90 percent", func(t *testing.T) {
		disks := []DiskMetrics{
			{
				Device:       "/dev/sda1",
				Mountpoint:   "/justover90",
				Filesystem:   "ext4",
				TotalBytes:   1000000000,
				UsedBytes:    901000000,
				FreeBytes:    99000000,
				UsagePercent: 90.1,
			},
		}

		result := AggregateDiskUsage(disks)
		assert.Len(t, result.CriticalMounts, 1)
		assert.Contains(t, result.CriticalMounts, "/justover90")
	})
}