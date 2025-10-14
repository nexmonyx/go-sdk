package nexmonyx

import (
	"strings"
	"time"
)

// AggregateDiskUsage calculates overall disk usage from individual disk metrics
// This function applies filtering rules to exclude virtual/temporary filesystems
// and returns aggregated statistics across all valid filesystems.
func AggregateDiskUsage(disks []DiskMetrics) *DiskUsageAggregate {
	var total, used, free uint64
	var largestMount string
	var largestSize uint64
	validCount := 0
	
	// Initialize empty slice to avoid nil
	criticalMounts := make([]string, 0)

	for _, disk := range disks {
		// Apply filtering rules to exclude virtual/temporary filesystems
		if !shouldIncludeInAggregation(disk) {
			continue
		}

		// Safely convert int64 to uint64 for aggregation
		// Negative values indicate invalid data and are treated as zero
		totalBytes := SafeInt64ToUint64OrZero(disk.TotalBytes)
		usedBytes := SafeInt64ToUint64OrZero(disk.UsedBytes)
		freeBytes := SafeInt64ToUint64OrZero(disk.FreeBytes)

		// Skip disks with invalid/zero total size
		if totalBytes == 0 {
			continue
		}

		total += totalBytes
		used += usedBytes
		free += freeBytes
		validCount++

		// Track largest mount by total capacity
		if totalBytes > largestSize {
			largestSize = totalBytes
			largestMount = disk.Mountpoint
		}

		// Track critical mounts (>90% full)
		if disk.UsagePercent > 90.0 {
			criticalMounts = append(criticalMounts, disk.Mountpoint)
		}
	}

	// Calculate overall usage percentage
	var usedPercent float64
	if total > 0 {
		usedPercent = float64(used) / float64(total) * 100.0
	}

	return &DiskUsageAggregate{
		TotalBytes:      total,
		UsedBytes:       used,
		FreeBytes:       free,
		UsedPercent:     usedPercent,
		FilesystemCount: validCount,
		LargestMount:    largestMount,
		CriticalMounts:  criticalMounts,
		CalculatedAt:    time.Now().UTC().Format(time.RFC3339),
	}
}

// shouldIncludeInAggregation determines if a filesystem should be included in aggregation
// This function filters out virtual, temporary, and system filesystems that should not
// be included in disk usage calculations.
func shouldIncludeInAggregation(disk DiskMetrics) bool {
	// INCLUDE these filesystem types
	includedTypes := map[string]bool{
		"ext4": true, "ext3": true, "ext2": true,
		"xfs": true, "btrfs": true, "zfs": true,
		"ntfs": true, "apfs": true, "hfs+": true,
		"reiserfs": true, "jfs": true,
		"nfs": true, "nfs4": true, "cifs": true, "smb": true,
		"glusterfs": true, "lustre": true,
		"overlay": true, "aufs": true,
	}

	// EXCLUDE these filesystem types
	excludedTypes := map[string]bool{
		"tmpfs": true, "devtmpfs": true, "sysfs": true,
		"proc": true, "devpts": true, "debugfs": true,
		"tracefs": true, "securityfs": true, "cgroup": true,
		"cgroup2": true, "pstore": true, "bpf": true,
		"autofs": true, "mqueue": true, "hugetlbfs": true,
		"fusectl": true, "configfs": true, "ramfs": true,
		"rpc_pipefs": true, "fuse.gvfsd-fuse": true,
		"fuse.portal": true, "efivarfs": true, "binfmt_misc": true,
	}

	// EXCLUDE these mount point patterns
	excludedMountPrefixes := []string{
		"/tmp", "/var/tmp", "/dev/shm",
		"/sys/", "/proc/", "/dev/",
		"/boot/efi", "/boot/EFI",
		"/var/lib/docker/overlay2/",
		"/var/lib/kubelet/pods/",
		"/snap/",
		"/run/systemd/",
	}

	// Check if filesystem type should be excluded
	if excludedTypes[disk.Filesystem] {
		return false
	}

	// Check if mount point should be excluded
	for _, prefix := range excludedMountPrefixes {
		if strings.HasPrefix(disk.Mountpoint, prefix) {
			return false
		}
	}

	// Check if filesystem type should be included
	if includedTypes[disk.Filesystem] {
		return true
	}

	// Default to exclude unknown filesystem types
	return false
}

// AggregateDiskUsageFromRequest is a convenience function that extracts disk metrics
// from a ComprehensiveMetricsRequest and calculates the aggregated disk usage.
// This is useful when you want to calculate aggregation from an existing metrics request.
func AggregateDiskUsageFromRequest(request *ComprehensiveMetricsRequest) *DiskUsageAggregate {
	if request == nil || request.Disks == nil {
		return &DiskUsageAggregate{
			TotalBytes:      0,
			UsedBytes:       0,
			FreeBytes:       0,
			UsedPercent:     0,
			FilesystemCount: 0,
			LargestMount:    "",
			CriticalMounts:  make([]string, 0),
			CalculatedAt:    time.Now().UTC().Format(time.RFC3339),
		}
	}

	return AggregateDiskUsage(request.Disks)
}

// ValidateDiskUsageAggregate validates that the DiskUsageAggregate struct contains
// consistent and valid data. Returns true if valid, false otherwise.
func ValidateDiskUsageAggregate(aggregate *DiskUsageAggregate) bool {
	if aggregate == nil {
		return false
	}

	// Check that TotalBytes = UsedBytes + FreeBytes (with some tolerance for rounding)
	expectedTotal := aggregate.UsedBytes + aggregate.FreeBytes
	tolerance := uint64(1024 * 1024) // 1MB tolerance for rounding differences

	if aggregate.TotalBytes > expectedTotal+tolerance || expectedTotal > aggregate.TotalBytes+tolerance {
		return false
	}

	// Check that usage percentage is consistent
	if aggregate.TotalBytes > 0 {
		expectedPercent := float64(aggregate.UsedBytes) / float64(aggregate.TotalBytes) * 100.0
		tolerance := 1.0 // 1% tolerance for rounding
		if expectedPercent > aggregate.UsedPercent+tolerance || aggregate.UsedPercent > expectedPercent+tolerance {
			return false
		}
	}

	// Check that usage percentage is within valid range
	if aggregate.UsedPercent < 0 || aggregate.UsedPercent > 100 {
		return false
	}

	// Check that filesystem count is non-negative
	if aggregate.FilesystemCount < 0 {
		return false
	}

	// Validate timestamp format
	if aggregate.CalculatedAt != "" {
		if _, err := time.Parse(time.RFC3339, aggregate.CalculatedAt); err != nil {
			return false
		}
	}

	return true
}