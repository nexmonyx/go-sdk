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

func TestDiskIOService_Submit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v2/metrics/disk-io", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Disk I/O metrics submitted successfully",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	serverUUID := uuid.New()
	submission := &DiskIOMetricsSubmission{
		ServerUUID: serverUUID,
		Timestamp:  time.Now(),
		Devices: []DiskIODeviceMetrics{
			{
				DeviceName:      "sda",
				ReadsCompleted:  1000,
				ReadsMerged:     100,
				SectorsRead:     50000,
				ReadTimeMs:      1500,
				WritesCompleted: 500,
				WritesMerged:    50,
				SectorsWritten:  25000,
				WriteTimeMs:     800,
				IOInProgress:    2,
				IOTimeMs:        2000,
				WeightedIOTimeMs: 2300,
			},
		},
	}

	err := client.DiskIO.Submit(context.Background(), submission)
	assert.NoError(t, err)
}

func TestDiskIOService_Submit_WithOptionalFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Metrics submitted",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	readBytesPerSec := int64(1024000)
	writeBytesPerSec := int64(512000)
	readOpsPerSec := 50.5
	writeOpsPerSec := 25.3
	utilizationPercent := 75.8
	deviceType := "ssd"
	deviceSize := int64(500000000000)
	deviceModel := "Samsung 970 EVO"
	deviceSerial := "S1234567890"

	submission := &DiskIOMetricsSubmission{
		ServerUUID: uuid.New(),
		Timestamp:  time.Now(),
		Devices: []DiskIODeviceMetrics{
			{
				DeviceName:         "nvme0n1",
				ReadsCompleted:     5000,
				WritesCompleted:    2500,
				ReadBytesPerSec:    &readBytesPerSec,
				WriteBytesPerSec:   &writeBytesPerSec,
				ReadOpsPerSec:      &readOpsPerSec,
				WriteOpsPerSec:     &writeOpsPerSec,
				UtilizationPercent: &utilizationPercent,
				DeviceType:         &deviceType,
				DeviceSize:         &deviceSize,
				DeviceModel:        &deviceModel,
				DeviceSerial:       &deviceSerial,
			},
		},
	}

	err := client.DiskIO.Submit(context.Background(), submission)
	assert.NoError(t, err)
}

func TestDiskIOService_Submit_MultipleDevices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	submission := &DiskIOMetricsSubmission{
		ServerUUID: uuid.New(),
		Timestamp:  time.Now(),
		Devices: []DiskIODeviceMetrics{
			{
				DeviceName:      "sda",
				ReadsCompleted:  1000,
				WritesCompleted: 500,
			},
			{
				DeviceName:      "sdb",
				ReadsCompleted:  2000,
				WritesCompleted: 1000,
			},
			{
				DeviceName:      "nvme0n1",
				ReadsCompleted:  5000,
				WritesCompleted: 2500,
			},
		},
	}

	err := client.DiskIO.Submit(context.Background(), submission)
	assert.NoError(t, err)
}

func TestDiskIOService_Submit_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Invalid metrics data",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	submission := &DiskIOMetricsSubmission{
		ServerUUID: uuid.New(),
		Timestamp:  time.Now(),
		Devices:    []DiskIODeviceMetrics{},
	}

	err := client.DiskIO.Submit(context.Background(), submission)
	assert.Error(t, err)
}

func TestDiskIOService_Submit_NetworkError(t *testing.T) {
	client, _ := NewClient(&Config{BaseURL: "http://invalid-server:9999"})

	submission := &DiskIOMetricsSubmission{
		ServerUUID: uuid.New(),
		Timestamp:  time.Now(),
		Devices: []DiskIODeviceMetrics{
			{DeviceName: "sda"},
		},
	}

	err := client.DiskIO.Submit(context.Background(), submission)
	assert.Error(t, err)
}
