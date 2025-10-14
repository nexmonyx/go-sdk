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

func TestSmartHealthService_Submit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v2/metrics/smart-health", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "SMART health metrics submitted successfully",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	submission := &SmartHealthMetricsSubmission{
		ServerUUID: uuid.New(),
		Timestamp:  time.Now(),
		Devices: []SmartHealthDeviceMetrics{
			{
				DeviceName:          "sda",
				DeviceModel:         "WD Blue",
				DeviceSerial:        "WD12345",
				DeviceType:          "ata",
				OverallHealthStatus: "PASSED",
				PredictedFailure:    false,
			},
		},
	}

	err := client.SmartHealth.Submit(context.Background(), submission)
	assert.NoError(t, err)
}

func TestSmartHealthService_Submit_CompleteHDDMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	temp := 45
	powerOn := int64(10000)
	cycles := int64(500)
	reallocated := int64(0)
	pending := int64(0)
	uncorrectable := int64(0)
	healthPct := 95.5
	confidence := 0.98
	deviceInterface := "SATA"
	firmware := "80.00A80"
	capacity := int64(2000000000000)
	spinRetry := int64(0)
	calibration := int64(0)
	headFlying := int64(10000)
	loadUnload := int64(5000)
	seekError := int64(10)

	submission := &SmartHealthMetricsSubmission{
		ServerUUID: uuid.New(),
		Timestamp:  time.Now(),
		Devices: []SmartHealthDeviceMetrics{
			{
				DeviceName:                  "sda",
				DeviceModel:                 "WD Red",
				DeviceSerial:                "WD-ABCD1234",
				DeviceType:                  "ata",
				DeviceInterface:             &deviceInterface,
				FirmwareVersion:             &firmware,
				CapacityBytes:               &capacity,
				OverallHealthStatus:         "PASSED",
				HealthPercentage:            &healthPct,
				PredictedFailure:            false,
				FailurePredictionConfidence: &confidence,
				TemperatureCelsius:          &temp,
				PowerOnHours:                &powerOn,
				PowerCycleCount:             &cycles,
				ReallocatedSectorsCount:     &reallocated,
				PendingSectorsCount:         &pending,
				UncorrectableErrorsCount:    &uncorrectable,
				SpinRetryCount:              &spinRetry,
				CalibrationRetryCount:       &calibration,
				HeadFlyingHours:             &headFlying,
				LoadUnloadCycles:            &loadUnload,
				SeekErrorRate:               &seekError,
			},
		},
	}

	err := client.SmartHealth.Submit(context.Background(), submission)
	assert.NoError(t, err)
}

func TestSmartHealthService_Submit_CompleteSSDMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	wearLevel := int64(5)
	programFail := int64(0)
	eraseFail := int64(0)
	lbasWritten := int64(1000000000)
	lbasRead := int64(5000000000)
	ssdLife := 95.0

	submission := &SmartHealthMetricsSubmission{
		ServerUUID: uuid.New(),
		Timestamp:  time.Now(),
		Devices: []SmartHealthDeviceMetrics{
			{
				DeviceName:           "sda",
				DeviceModel:          "Samsung 870 EVO",
				DeviceSerial:         "S123456789",
				DeviceType:           "ata",
				OverallHealthStatus:  "PASSED",
				PredictedFailure:     false,
				WearLevelingCount:    &wearLevel,
				ProgramFailCount:     &programFail,
				EraseFailCount:       &eraseFail,
				TotalLBAsWritten:     &lbasWritten,
				TotalLBAsRead:        &lbasRead,
				SSDLifeLeftPercent:   &ssdLife,
			},
		},
	}

	err := client.SmartHealth.Submit(context.Background(), submission)
	assert.NoError(t, err)
}

func TestSmartHealthService_Submit_CompleteNVMeMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	criticalWarning := 0
	compositeTemp := 40
	availableSpare := 100.0
	spareThreshold := 10.0
	percentUsed := 5.0
	dataUnitsRead := int64(500000)
	dataUnitsWritten := int64(250000)
	hostReads := int64(1000000)
	hostWrites := int64(500000)
	busyTime := int64(50000)
	powerCycles := int64(100)
	powerOnHours := int64(5000)
	unsafeShutdowns := int64(5)
	mediaErrors := int64(0)
	errorLogEntries := int64(0)

	submission := &SmartHealthMetricsSubmission{
		ServerUUID: uuid.New(),
		Timestamp:  time.Now(),
		Devices: []SmartHealthDeviceMetrics{
			{
				DeviceName:                             "nvme0n1",
				DeviceModel:                            "Samsung 970 EVO Plus",
				DeviceSerial:                           "S4EUNG0N123456",
				DeviceType:                             "nvme",
				OverallHealthStatus:                    "PASSED",
				PredictedFailure:                       false,
				NVMeCriticalWarning:                    &criticalWarning,
				NVMeCompositeTemperature:               &compositeTemp,
				NVMeAvailableSparePercent:              &availableSpare,
				NVMeAvailableSpareThresholdPercent:     &spareThreshold,
				NVMePercentageUsed:                     &percentUsed,
				NVMeDataUnitsRead:                      &dataUnitsRead,
				NVMeDataUnitsWritten:                   &dataUnitsWritten,
				NVMeHostReads:                          &hostReads,
				NVMeHostWrites:                         &hostWrites,
				NVMeControllerBusyTime:                 &busyTime,
				NVMePowerCycles:                        &powerCycles,
				NVMePowerOnHours:                       &powerOnHours,
				NVMeUnsafeShutdowns:                    &unsafeShutdowns,
				NVMeMediaErrors:                        &mediaErrors,
				NVMeErrorLogEntries:                    &errorLogEntries,
			},
		},
	}

	err := client.SmartHealth.Submit(context.Background(), submission)
	assert.NoError(t, err)
}

func TestSmartHealthService_Submit_WithWarningsAndAlerts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	submission := &SmartHealthMetricsSubmission{
		ServerUUID: uuid.New(),
		Timestamp:  time.Now(),
		Devices: []SmartHealthDeviceMetrics{
			{
				DeviceName:          "sda",
				DeviceModel:         "Seagate Barracuda",
				DeviceSerial:        "ZA123456",
				DeviceType:          "ata",
				OverallHealthStatus: "FAILED",
				PredictedFailure:    true,
				CriticalWarnings:    []string{"High temperature", "Reallocated sectors"},
				WarningMessages:     []string{"SMART test failed", "Pending sectors detected"},
			},
		},
	}

	err := client.SmartHealth.Submit(context.Background(), submission)
	assert.NoError(t, err)
}

func TestSmartHealthService_Submit_MultipleDevices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	submission := &SmartHealthMetricsSubmission{
		ServerUUID: uuid.New(),
		Timestamp:  time.Now(),
		Devices: []SmartHealthDeviceMetrics{
			{
				DeviceName:          "sda",
				DeviceModel:         "WD Blue",
				DeviceSerial:        "WD1",
				DeviceType:          "ata",
				OverallHealthStatus: "PASSED",
				PredictedFailure:    false,
			},
			{
				DeviceName:          "sdb",
				DeviceModel:         "Samsung SSD",
				DeviceSerial:        "S1",
				DeviceType:          "ata",
				OverallHealthStatus: "PASSED",
				PredictedFailure:    false,
			},
			{
				DeviceName:          "nvme0n1",
				DeviceModel:         "Intel Optane",
				DeviceSerial:        "INTEL1",
				DeviceType:          "nvme",
				OverallHealthStatus: "PASSED",
				PredictedFailure:    false,
			},
		},
	}

	err := client.SmartHealth.Submit(context.Background(), submission)
	assert.NoError(t, err)
}

func TestSmartHealthService_Submit_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Invalid SMART data",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})

	submission := &SmartHealthMetricsSubmission{
		ServerUUID: uuid.New(),
		Timestamp:  time.Now(),
		Devices:    []SmartHealthDeviceMetrics{},
	}

	err := client.SmartHealth.Submit(context.Background(), submission)
	assert.Error(t, err)
}

func TestSmartHealthService_Submit_NetworkError(t *testing.T) {
	client, _ := NewClient(&Config{BaseURL: "http://invalid-server:9999"})

	submission := &SmartHealthMetricsSubmission{
		ServerUUID: uuid.New(),
		Timestamp:  time.Now(),
		Devices: []SmartHealthDeviceMetrics{
			{DeviceName: "sda"},
		},
	}

	err := client.SmartHealth.Submit(context.Background(), submission)
	assert.Error(t, err)
}
