package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Submit method coverage tests

func TestIPMIService_Submit_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Internal server error",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	req := &IPMISubmitRequest{
		ServerUUID:  "test-uuid",
		CollectedAt: time.Now(),
		IPMI:        IPMIInfo{CollectionMethod: "ipmitool"},
	}

	resp, err := client.IPMI.Submit(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

// Type assertion failure for Submit happens during JSON unmarshaling, not in code

// Get method coverage tests

func TestIPMIService_Get_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Server not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	resp, err := client.IPMI.Get(context.Background(), "nonexistent-uuid")
	assert.Error(t, err)
	assert.Nil(t, resp)
}

// Type assertion failure for Get happens during JSON unmarshaling, not in code

// GetSensorData method coverage tests

func TestIPMIService_GetSensorData_EmptyList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/ipmi/test-uuid/sensors", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": []*IPMISensor{
				{
					ID:     "sensor1",
					Name:   "CPU Temp",
					Type:   "temperature",
					Status: "ok",
				},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	sensors, err := client.IPMI.GetSensorData(context.Background(), "test-uuid")
	assert.NoError(t, err)
	assert.NotNil(t, sensors)
	assert.Len(t, sensors, 1)
	assert.Equal(t, "sensor1", sensors[0].ID)
}

func TestIPMIService_GetSensorData_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Failed to get sensor data",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	sensors, err := client.IPMI.GetSensorData(context.Background(), "test-uuid")
	assert.Error(t, err)
	assert.Nil(t, sensors)
}

// ExecuteCommand method coverage tests

func TestIPMIService_ExecuteCommand_PowerStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/ipmi/test-uuid/execute", r.URL.Path)

		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.NoError(t, err)
		assert.Equal(t, "power", body["command"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": IPMICommandResult{
				Command:  "power",
				Output:   "Chassis Power is on",
				ExitCode: 0,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, err := client.IPMI.ExecuteCommand(context.Background(), "test-uuid", "power", []string{"status"})
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "power", result.Command)
	assert.Equal(t, 0, result.ExitCode)
}

func TestIPMIService_ExecuteCommand_InvalidCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Invalid command",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, err := client.IPMI.ExecuteCommand(context.Background(), "test-uuid", "invalid", []string{})
	assert.Error(t, err)
	assert.Nil(t, result)
}

// Type assertion failure for ExecuteCommand happens during JSON unmarshaling, not in code

// GetIPMI method coverage tests - with time range

func TestIPMIService_GetIPMI_WithTimeRange(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v2/ipmi/test-uuid", r.URL.Path)

		// Verify time range parameters
		assert.NotEmpty(t, r.URL.Query().Get("start"))
		assert.NotEmpty(t, r.URL.Query().Get("end"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": IPMIInfo{
				CollectionMethod: "ipmitool",
				IPMIVersion:      "2.0",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	timeRange := &TimeRange{
		Start: "2024-01-01T00:00:00Z",
		End:   "2024-01-02T00:00:00Z",
	}

	ipmi, err := client.IPMI.GetIPMI(context.Background(), "test-uuid", timeRange)
	assert.NoError(t, err)
	assert.NotNil(t, ipmi)
	assert.Equal(t, "ipmitool", ipmi.CollectionMethod)
}

func TestIPMIService_GetIPMI_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Internal server error",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	ipmi, err := client.IPMI.GetIPMI(context.Background(), "test-uuid", nil)
	assert.Error(t, err)
	assert.Nil(t, ipmi)
}

// Type assertion failure for GetIPMI happens during JSON unmarshaling, not in code

// GetLatestIPMI method coverage tests

func TestIPMIService_GetLatestIPMI_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "No IPMI data found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	ipmi, err := client.IPMI.GetLatestIPMI(context.Background(), "test-uuid")
	assert.Error(t, err)
	assert.Nil(t, ipmi)
}

// Type assertion failure for GetLatestIPMI happens during JSON unmarshaling, not in code

// ListIPMIHistory method coverage tests - with options

func TestIPMIService_ListIPMIHistory_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v2/ipmi/test-uuid/history", r.URL.Path)

		// Verify pagination parameters
		assert.NotEmpty(t, r.URL.Query().Get("page"))
		assert.NotEmpty(t, r.URL.Query().Get("limit"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": []*IPMIRecord{
				{
					ID:         1,
					ServerUUID: "test-uuid",
					IPMI: IPMIInfo{
						CollectionMethod: "ipmitool",
					},
				},
			},
			"meta": PaginationMeta{
				Page:       1,
				Limit:      10,
				TotalItems: 1,
				TotalPages: 1,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	opts := &IPMIListOptions{
		ListOptions: ListOptions{
			Page:  1,
			Limit: 10,
		},
	}

	records, meta, err := client.IPMI.ListIPMIHistory(context.Background(), "test-uuid", opts)
	assert.NoError(t, err)
	assert.NotNil(t, records)
	assert.NotNil(t, meta)
	assert.Len(t, records, 1)
	assert.Equal(t, 1, meta.Page)
	assert.Equal(t, 10, meta.Limit)
}

func TestIPMIService_ListIPMIHistory_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Failed to fetch history",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	records, meta, err := client.IPMI.ListIPMIHistory(context.Background(), "test-uuid", nil)
	assert.Error(t, err)
	assert.Nil(t, records)
	assert.Nil(t, meta)
}
