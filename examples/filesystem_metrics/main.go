package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	nexmonyx "github.com/nexmonyx/go-sdk/v2"
)

func main() {
	// Create client
	client, err := nexmonyx.NewClient(&nexmonyx.Config{
		BaseURL: "https://api-dev.nexmonyx.com",
		Auth: nexmonyx.AuthConfig{
			ServerUUID:   os.Getenv("SERVER_UUID"),
			ServerSecret: os.Getenv("SERVER_SECRET"),
		},
		Debug: true,
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	serverUUID, err := uuid.Parse(os.Getenv("SERVER_UUID"))
	if err != nil {
		log.Fatalf("Invalid SERVER_UUID: %v", err)
	}

	// Example 1: Submit ZFS pool metrics
	fmt.Println("=== Testing ZFS Pool Metrics ===")
	err = testZFSMetrics(client, serverUUID)
	if err != nil {
		log.Printf("ZFS test failed: %v", err)
	} else {
		fmt.Println("✅ ZFS metrics submitted successfully")
	}

	// Example 2: Submit RAID array metrics
	fmt.Println("\n=== Testing RAID Array Metrics ===")
	err = testRAIDMetrics(client, serverUUID)
	if err != nil {
		log.Printf("RAID test failed: %v", err)
	} else {
		fmt.Println("✅ RAID metrics submitted successfully")
	}

	// Example 3: Submit LVM volume metrics
	fmt.Println("\n=== Testing LVM Volume Metrics ===")
	err = testLVMMetrics(client, serverUUID)
	if err != nil {
		log.Printf("LVM test failed: %v", err)
	} else {
		fmt.Println("✅ LVM metrics submitted successfully")
	}

	// Example 4: Submit general filesystem metrics
	fmt.Println("\n=== Testing General Filesystem Metrics ===")
	err = testGeneralFilesystemMetrics(client, serverUUID)
	if err != nil {
		log.Printf("General filesystem test failed: %v", err)
	} else {
		fmt.Println("✅ General filesystem metrics submitted successfully")
	}
}

func testZFSMetrics(client *nexmonyx.Client, serverUUID uuid.UUID) error {
	zfsMetrics := []nexmonyx.ZFSPoolMetrics{
		{
			PoolName:                "zfspool",
			TotalBytes:              ptrInt64(159175803469824),
			UsedBytes:               ptrInt64(95505482081894),
			AvailableBytes:          ptrInt64(63670321387930),
			UsagePercent:            ptrFloat64(60.0),
			Health:                  ptrString("ONLINE"),
			State:                   ptrString("ACTIVE"),
			CompressionRatio:        ptrFloat64(1.45),
			DedupRatio:              ptrFloat64(1.0),
			FragmentationPercent:    ptrFloat64(15.2),
			AllocatedBytes:          ptrInt64(95505482081894),
			ReferencedBytes:         ptrInt64(89234567890123),
			SnapshotsCount:          ptrInt(12),
			SnapshotSizeBytes:       ptrInt64(6270914191771),
			ScrubState:              ptrString("completed"),
			ScrubPercentComplete:    ptrFloat64(100.0),
			ReadErrors:              ptrInt64(0),
			WriteErrors:             ptrInt64(0),
			ChecksumErrors:          ptrInt64(0),
			OverallHealth:           "HEALTHY",
			HealthScore:             ptrFloat64(98.5),
			WarningCount:            0,
			ErrorCount:              0,
		},
	}

	return client.Filesystem.SubmitZFS(context.Background(), serverUUID, zfsMetrics)
}

func testRAIDMetrics(client *nexmonyx.Client, serverUUID uuid.UUID) error {
	raidMetrics := []nexmonyx.RAIDArrayMetrics{
		{
			DeviceName:     "/dev/md0",
			TotalBytes:     ptrInt64(2000398934016),
			UsedBytes:      ptrInt64(1200239360409),
			AvailableBytes: ptrInt64(800159573607),
			UsagePercent:   ptrFloat64(60.0),
			Level:          ptrString("raid1"),
			State:          ptrString("clean"),
			TotalDevices:   ptrInt(2),
			ActiveDevices:  ptrInt(2),
			SpareDevices:   ptrInt(0),
			FailedDevices:  ptrInt(0),
			ChunkSizeKB:    ptrInt(512),
			OverallHealth:  "HEALTHY",
			HealthScore:    ptrFloat64(100.0),
			WarningCount:   0,
			ErrorCount:     0,
		},
	}

	return client.Filesystem.SubmitRAID(context.Background(), serverUUID, raidMetrics)
}

func testLVMMetrics(client *nexmonyx.Client, serverUUID uuid.UUID) error {
	lvmMetrics := []nexmonyx.LVMVolumeMetrics{
		{
			LogicalVolumeName:         "root",
			VolumeGroupName:          "vg0",
			DevicePath:               ptrString("/dev/mapper/vg0-root"),
			TotalBytes:               ptrInt64(107374182400),
			UsedBytes:                ptrInt64(75161927680),
			AvailableBytes:           ptrInt64(32212254720),
			UsagePercent:             ptrFloat64(70.0),
			PhysicalVolumeCount:      ptrInt(1),
			LogicalVolumeCount:       ptrInt(3),
			PhysicalExtentSize:       ptrInt64(4194304),
			TotalPhysicalExtents:     ptrInt(25600),
			FreePhysicalExtents:      ptrInt(7680),
			AllocatedPhysicalExtents: ptrInt(17920),
			VolumeGroupStatus:        ptrString("available"),
			LogicalVolumeStatus:      ptrString("available"),
			Attributes:               ptrString("-wi-ao----"),
			OverallHealth:            "HEALTHY",
			HealthScore:              ptrFloat64(95.0),
			WarningCount:             0,
			ErrorCount:               0,
		},
	}

	return client.Filesystem.SubmitLVM(context.Background(), serverUUID, lvmMetrics)
}

func testGeneralFilesystemMetrics(client *nexmonyx.Client, serverUUID uuid.UUID) error {
	submission := &nexmonyx.FilesystemMetricsSubmission{
		ServerUUID: serverUUID,
		Timestamp:  time.Now(),
		Filesystems: []nexmonyx.FilesystemMetricsData{
			{
				FilesystemName:   "/dev/sda1",
				FilesystemType:   "ext4",
				MountPoint:       ptrString("/boot"),
				TotalBytes:       ptrInt64(1073741824),
				UsedBytes:        ptrInt64(536870912),
				AvailableBytes:   ptrInt64(536870912),
				UsagePercent:     ptrFloat64(50.0),
				ReadOpsPerSec:    ptrFloat64(10.5),
				WriteOpsPerSec:   ptrFloat64(5.2),
				ReadBytesPerSec:  ptrInt64(2097152),
				WriteBytesPerSec: ptrInt64(1048576),
				OverallHealth:    "HEALTHY",
				HealthScore:      ptrFloat64(100.0),
				WarningCount:     0,
				ErrorCount:       0,
			},
		},
	}

	return client.Filesystem.Submit(context.Background(), submission)
}

// Helper functions for creating pointers
func ptrString(s string) *string { return &s }
func ptrInt(i int) *int { return &i }
func ptrInt64(i int64) *int64 { return &i }
func ptrFloat64(f float64) *float64 { return &f }