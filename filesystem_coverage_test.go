package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestFilesystemService_Submit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v2/metrics/filesystem", r.URL.Path)

		// Verify request body structure
		var submission FilesystemMetricsSubmission
		err := json.NewDecoder(r.Body).Decode(&submission)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, submission.ServerUUID)
		assert.NotZero(t, submission.Timestamp)
		assert.NotEmpty(t, submission.Filesystems)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Filesystem metrics submitted successfully",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	serverUUID := uuid.New()
	totalBytes := int64(1000000000)
	usedBytes := int64(500000000)
	availableBytes := int64(500000000)
	usagePercent := 50.0

	submission := &FilesystemMetricsSubmission{
		ServerUUID: serverUUID,
		Timestamp:  time.Now(),
		Filesystems: []FilesystemMetricsData{
			{
				FilesystemName: "root",
				FilesystemType: "ext4",
				TotalBytes:     &totalBytes,
				UsedBytes:      &usedBytes,
				AvailableBytes: &availableBytes,
				UsagePercent:   &usagePercent,
				OverallHealth:  "HEALTHY",
				WarningCount:   0,
				ErrorCount:     0,
			},
		},
	}

	err := client.Filesystem.Submit(context.Background(), submission)
	assert.NoError(t, err)
}

func TestFilesystemService_Submit_EmptyFilesystems(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v2/metrics/filesystem", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Filesystem metrics submitted successfully",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	submission := &FilesystemMetricsSubmission{
		ServerUUID:  uuid.New(),
		Timestamp:   time.Now(),
		Filesystems: []FilesystemMetricsData{},
	}

	err := client.Filesystem.Submit(context.Background(), submission)
	assert.NoError(t, err)
}

func TestFilesystemService_Submit_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Invalid filesystem metrics data",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	submission := &FilesystemMetricsSubmission{
		ServerUUID:  uuid.New(),
		Timestamp:   time.Now(),
		Filesystems: []FilesystemMetricsData{},
	}

	err := client.Filesystem.Submit(context.Background(), submission)
	assert.Error(t, err)
}

func TestFilesystemService_SubmitZFS(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v2/metrics/filesystem", r.URL.Path)

		// Verify request body structure
		var submission FilesystemMetricsSubmission
		err := json.NewDecoder(r.Body).Decode(&submission)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, submission.ServerUUID)
		assert.Len(t, submission.Filesystems, 2) // We're sending 2 ZFS pools

		// Verify ZFS-specific fields
		for _, fs := range submission.Filesystems {
			assert.Equal(t, "zfs", fs.FilesystemType)
			assert.NotNil(t, fs.ZFSPoolName)
			assert.NotNil(t, fs.ZFSPoolHealth)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "ZFS metrics submitted successfully",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	serverUUID := uuid.New()
	totalBytes := int64(2000000000000)
	usedBytes := int64(1000000000000)
	availableBytes := int64(1000000000000)
	usagePercent := 50.0
	healthStr := "ONLINE"
	stateStr := "ACTIVE"
	compressionRatio := 1.5
	dedupRatio := 1.2
	fragmentationPercent := 10.5
	allocatedBytes := int64(1200000000000)
	referencedBytes := int64(1000000000000)
	snapshotsCount := 5
	snapshotSizeBytes := int64(50000000000)
	scrubState := "completed"
	scrubPercent := 100.0
	readErrors := int64(0)
	writeErrors := int64(0)
	checksumErrors := int64(0)
	healthScore := 95.5

	zfsMetrics := []ZFSPoolMetrics{
		{
			PoolName:                "tank",
			TotalBytes:              &totalBytes,
			UsedBytes:               &usedBytes,
			AvailableBytes:          &availableBytes,
			UsagePercent:            &usagePercent,
			Health:                  &healthStr,
			State:                   &stateStr,
			CompressionRatio:        &compressionRatio,
			DedupRatio:              &dedupRatio,
			FragmentationPercent:    &fragmentationPercent,
			AllocatedBytes:          &allocatedBytes,
			ReferencedBytes:         &referencedBytes,
			SnapshotsCount:          &snapshotsCount,
			SnapshotSizeBytes:       &snapshotSizeBytes,
			ScrubState:              &scrubState,
			ScrubPercentComplete:    &scrubPercent,
			ReadErrors:              &readErrors,
			WriteErrors:             &writeErrors,
			ChecksumErrors:          &checksumErrors,
			OverallHealth:           "HEALTHY",
			HealthScore:             &healthScore,
			WarningCount:            0,
			ErrorCount:              0,
		},
		{
			PoolName:       "backup",
			TotalBytes:     &totalBytes,
			UsedBytes:      &usedBytes,
			AvailableBytes: &availableBytes,
			UsagePercent:   &usagePercent,
			Health:         &healthStr,
			State:          &stateStr,
			OverallHealth:  "HEALTHY",
			HealthScore:    &healthScore,
			WarningCount:   0,
			ErrorCount:     0,
		},
	}

	err := client.Filesystem.SubmitZFS(context.Background(), serverUUID, zfsMetrics)
	assert.NoError(t, err)
}

func TestFilesystemService_SubmitZFS_EmptyMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)

		// Verify request body has empty filesystems array
		var submission FilesystemMetricsSubmission
		err := json.NewDecoder(r.Body).Decode(&submission)
		assert.NoError(t, err)
		assert.Empty(t, submission.Filesystems)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "ZFS metrics submitted successfully",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	err := client.Filesystem.SubmitZFS(context.Background(), uuid.New(), []ZFSPoolMetrics{})
	assert.NoError(t, err)
}

func TestFilesystemService_SubmitZFS_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Failed to process ZFS metrics",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	serverUUID := uuid.New()
	totalBytes := int64(2000000000000)
	usedBytes := int64(1000000000000)
	availableBytes := int64(1000000000000)
	usagePercent := 50.0
	healthStr := "ONLINE"
	stateStr := "ACTIVE"
	healthScore := 95.5

	zfsMetrics := []ZFSPoolMetrics{
		{
			PoolName:       "tank",
			TotalBytes:     &totalBytes,
			UsedBytes:      &usedBytes,
			AvailableBytes: &availableBytes,
			UsagePercent:   &usagePercent,
			Health:         &healthStr,
			State:          &stateStr,
			OverallHealth:  "HEALTHY",
			HealthScore:    &healthScore,
			WarningCount:   0,
			ErrorCount:     0,
		},
	}

	err := client.Filesystem.SubmitZFS(context.Background(), serverUUID, zfsMetrics)
	assert.Error(t, err)
}

func TestFilesystemService_SubmitRAID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v2/metrics/filesystem", r.URL.Path)

		// Verify request body structure
		var submission FilesystemMetricsSubmission
		err := json.NewDecoder(r.Body).Decode(&submission)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, submission.ServerUUID)
		assert.Len(t, submission.Filesystems, 1)

		// Verify RAID-specific fields
		fs := submission.Filesystems[0]
		assert.Equal(t, "mdraid", fs.FilesystemType)
		assert.NotNil(t, fs.RAIDLevel)
		assert.NotNil(t, fs.RAIDDeviceName)
		assert.NotNil(t, fs.RAIDState)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "RAID metrics submitted successfully",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	serverUUID := uuid.New()
	totalBytes := int64(4000000000000)
	usedBytes := int64(2000000000000)
	availableBytes := int64(2000000000000)
	usagePercent := 50.0
	raidLevel := "raid5"
	raidState := "active"
	totalDevices := 4
	activeDevices := 4
	spareDevices := 0
	failedDevices := 0
	syncPercent := 100.0
	chunkSizeKB := 512
	healthScore := 98.0

	raidMetrics := []RAIDArrayMetrics{
		{
			DeviceName:     "/dev/md0",
			TotalBytes:     &totalBytes,
			UsedBytes:      &usedBytes,
			AvailableBytes: &availableBytes,
			UsagePercent:   &usagePercent,
			Level:          &raidLevel,
			State:          &raidState,
			TotalDevices:   &totalDevices,
			ActiveDevices:  &activeDevices,
			SpareDevices:   &spareDevices,
			FailedDevices:  &failedDevices,
			SyncPercent:    &syncPercent,
			ChunkSizeKB:    &chunkSizeKB,
			OverallHealth:  "HEALTHY",
			HealthScore:    &healthScore,
			WarningCount:   0,
			ErrorCount:     0,
		},
	}

	err := client.Filesystem.SubmitRAID(context.Background(), serverUUID, raidMetrics)
	assert.NoError(t, err)
}

func TestFilesystemService_SubmitRAID_EmptyMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)

		// Verify request body has empty filesystems array
		var submission FilesystemMetricsSubmission
		err := json.NewDecoder(r.Body).Decode(&submission)
		assert.NoError(t, err)
		assert.Empty(t, submission.Filesystems)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "RAID metrics submitted successfully",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	err := client.Filesystem.SubmitRAID(context.Background(), uuid.New(), []RAIDArrayMetrics{})
	assert.NoError(t, err)
}

func TestFilesystemService_SubmitRAID_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Invalid RAID metrics data",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	serverUUID := uuid.New()
	totalBytes := int64(4000000000000)
	usedBytes := int64(2000000000000)
	availableBytes := int64(2000000000000)
	usagePercent := 50.0
	raidLevel := "raid5"
	raidState := "active"
	healthScore := 98.0

	raidMetrics := []RAIDArrayMetrics{
		{
			DeviceName:     "/dev/md0",
			TotalBytes:     &totalBytes,
			UsedBytes:      &usedBytes,
			AvailableBytes: &availableBytes,
			UsagePercent:   &usagePercent,
			Level:          &raidLevel,
			State:          &raidState,
			OverallHealth:  "HEALTHY",
			HealthScore:    &healthScore,
			WarningCount:   0,
			ErrorCount:     0,
		},
	}

	err := client.Filesystem.SubmitRAID(context.Background(), serverUUID, raidMetrics)
	assert.Error(t, err)
}

func TestFilesystemService_SubmitLVM(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v2/metrics/filesystem", r.URL.Path)

		// Verify request body structure
		var submission FilesystemMetricsSubmission
		err := json.NewDecoder(r.Body).Decode(&submission)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, submission.ServerUUID)
		assert.Len(t, submission.Filesystems, 2)

		// Verify LVM-specific fields
		for _, fs := range submission.Filesystems {
			assert.Equal(t, "lvm", fs.FilesystemType)
			assert.NotNil(t, fs.LVMVGName)
			assert.NotNil(t, fs.LVMLVName)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "LVM metrics submitted successfully",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	serverUUID := uuid.New()
	mountPoint := "/mnt/data"
	devicePath := "/dev/vg0/lv0"
	totalBytes := int64(1000000000000)
	usedBytes := int64(500000000000)
	availableBytes := int64(500000000000)
	usagePercent := 50.0
	pvCount := 2
	lvCount := 3
	peSizeBytes := int64(4194304)
	totalPE := 238418
	freePE := 119209
	allocatedPE := 119209
	vgStatus := "available"
	lvStatus := "active"
	attributes := "-wi-ao----"
	healthScore := 95.0

	lvmMetrics := []LVMVolumeMetrics{
		{
			LogicalVolumeName:        "lv_data",
			VolumeGroupName:          "vg0",
			MountPoint:               &mountPoint,
			DevicePath:               &devicePath,
			TotalBytes:               &totalBytes,
			UsedBytes:                &usedBytes,
			AvailableBytes:           &availableBytes,
			UsagePercent:             &usagePercent,
			PhysicalVolumeCount:      &pvCount,
			LogicalVolumeCount:       &lvCount,
			PhysicalExtentSize:       &peSizeBytes,
			TotalPhysicalExtents:     &totalPE,
			FreePhysicalExtents:      &freePE,
			AllocatedPhysicalExtents: &allocatedPE,
			VolumeGroupStatus:        &vgStatus,
			LogicalVolumeStatus:      &lvStatus,
			Attributes:               &attributes,
			OverallHealth:            "HEALTHY",
			HealthScore:              &healthScore,
			WarningCount:             0,
			ErrorCount:               0,
		},
		{
			LogicalVolumeName:   "lv_backup",
			VolumeGroupName:     "vg0",
			MountPoint:          &mountPoint,
			TotalBytes:          &totalBytes,
			UsedBytes:           &usedBytes,
			AvailableBytes:      &availableBytes,
			UsagePercent:        &usagePercent,
			VolumeGroupStatus:   &vgStatus,
			LogicalVolumeStatus: &lvStatus,
			OverallHealth:       "HEALTHY",
			HealthScore:         &healthScore,
			WarningCount:        0,
			ErrorCount:          0,
		},
	}

	err := client.Filesystem.SubmitLVM(context.Background(), serverUUID, lvmMetrics)
	assert.NoError(t, err)
}

func TestFilesystemService_SubmitLVM_EmptyMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)

		// Verify request body has empty filesystems array
		var submission FilesystemMetricsSubmission
		err := json.NewDecoder(r.Body).Decode(&submission)
		assert.NoError(t, err)
		assert.Empty(t, submission.Filesystems)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "LVM metrics submitted successfully",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	err := client.Filesystem.SubmitLVM(context.Background(), uuid.New(), []LVMVolumeMetrics{})
	assert.NoError(t, err)
}

func TestFilesystemService_SubmitLVM_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Unauthorized",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	serverUUID := uuid.New()
	totalBytes := int64(1000000000000)
	usedBytes := int64(500000000000)
	availableBytes := int64(500000000000)
	usagePercent := 50.0
	vgStatus := "available"
	lvStatus := "active"
	healthScore := 95.0

	lvmMetrics := []LVMVolumeMetrics{
		{
			LogicalVolumeName:   "lv_data",
			VolumeGroupName:     "vg0",
			TotalBytes:          &totalBytes,
			UsedBytes:           &usedBytes,
			AvailableBytes:      &availableBytes,
			UsagePercent:        &usagePercent,
			VolumeGroupStatus:   &vgStatus,
			LogicalVolumeStatus: &lvStatus,
			OverallHealth:       "HEALTHY",
			HealthScore:         &healthScore,
			WarningCount:        0,
			ErrorCount:          0,
		},
	}

	err := client.Filesystem.SubmitLVM(context.Background(), serverUUID, lvmMetrics)
	assert.Error(t, err)
}
