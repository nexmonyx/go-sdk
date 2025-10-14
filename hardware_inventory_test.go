package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHardwareInventoryService_Submit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v2/hardware/inventory", r.URL.Path)
		assert.Equal(t, "test-uuid", r.Header.Get("X-Server-UUID"))
		assert.Equal(t, "test-secret", r.Header.Get("X-Server-Secret"))

		var req HardwareInventoryRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "test-server-123", req.ServerUUID)
		assert.Equal(t, "dmidecode", req.Hardware.CollectionMethod)
		assert.Equal(t, "lshw", req.Hardware.DetectionTool)
		assert.Equal(t, "Dell Inc.", req.Hardware.Manufacturer)
		assert.Len(t, req.Hardware.CPUs, 2)

		response := map[string]interface{}{
			"data": HardwareInventorySubmitResponse{
				ServerUUID:       "test-server-123",
				Timestamp:        time.Now(),
				CollectionMethod: "dmidecode",
				DetectionTool:    "lshw",
				ComponentCounts: map[string]int{
					"cpus":            2,
					"memory_modules":  4,
					"storage_devices": 2,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			ServerUUID:   "test-uuid",
			ServerSecret: "test-secret",
		},
	})
	require.NoError(t, err)

	req := &HardwareInventoryRequest{
		ServerUUID:  "test-server-123",
		CollectedAt: time.Now(),
		Hardware: HardwareInventoryInfo{
			Manufacturer:     "Dell Inc.",
			Model:            "PowerEdge R740",
			SerialNumber:     "ABC123",
			CollectionMethod: "dmidecode",
			DetectionTool:    "lshw",
			CPUs: []CPUInfo{
				{
					Manufacturer: "Intel",
					Model:        "Xeon Gold 6230",
					Cores:        20,
					Threads:      40,
				},
				{
					Manufacturer: "Intel",
					Model:        "Xeon Gold 6230",
					Cores:        20,
					Threads:      40,
				},
			},
			Memory: &MemoryInfo{
				TotalCapacity:  137438953472, // 128GB
				AvailableSlots: 24,
				UsedSlots:      4,
				ECCSupported:   true,
			},
			MemoryModules: []MemoryModuleInfo{
				{
					Slot:         "DIMM_A1",
					Manufacturer: "Samsung",
					Size:         34359738368, // 32GB
					Type:         "DDR4",
					Speed:        2666,
				},
			},
			StorageDevices: []StorageDeviceInfo{
				{
					DeviceName:   "/dev/sda",
					Manufacturer: "Samsung",
					Model:        "SSD 860 EVO",
					Capacity:     500107862016, // ~500GB
					Type:         "SSD",
					Interface:    "SATA",
				},
			},
		},
	}

	resp, err := client.HardwareInventory.Submit(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test-server-123", resp.ServerUUID)
	assert.Equal(t, "dmidecode", resp.CollectionMethod)
	assert.Equal(t, 2, resp.ComponentCounts["cpus"])
}

func TestHardwareInventoryService_Get(t *testing.T) {
	serverUUID := "test-server-123"
	_ = time.Now()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/hardware-inventory/"+serverUUID, r.URL.Path)

		response := map[string]interface{}{
			"success": true,
			"data": HardwareInventoryInfo{
				Manufacturer:     "Dell Inc.",
				Model:            "PowerEdge R740",
				CollectionMethod: "dmidecode",
				DetectionTool:    "lshw",
				
				CPUs: []CPUInfo{
					{
						Manufacturer: "Intel",
						Model:        "Xeon Gold 6230",
						Cores:        20,
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	require.NoError(t, err)

	inventory, err := client.HardwareInventory.Get(context.Background(), serverUUID)
	require.NoError(t, err)
	assert.NotNil(t, inventory)
	assert.Equal(t, "Dell Inc.", inventory.Manufacturer)
	assert.Equal(t, "PowerEdge R740", inventory.Model)
	assert.Len(t, inventory.CPUs, 1)
}

func TestHardwareInventoryService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/hardware-inventory", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "25", r.URL.Query().Get("limit"))

		inventories := []*HardwareInventoryInfo{
			{
				Manufacturer: "Dell Inc.",
				Model:        "PowerEdge R740",
			},
			{
				Manufacturer: "HP",
				Model:        "ProLiant DL380",
			},
		}

		response := map[string]interface{}{
			"success": true,
			"data":    inventories,
			"meta": map[string]interface{}{
				"total":       2,
				"page":        1,
				"limit":       25,
				"total_pages": 1,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	require.NoError(t, err)

	inventories, meta, err := client.HardwareInventory.List(context.Background(), &ListOptions{
		Page:  1,
		Limit: 25,
	})
	require.NoError(t, err)
	assert.Len(t, inventories, 2)
	assert.NotNil(t, meta)
	assert.Equal(t, 1, meta.Page)
}

func TestHardwareInventoryService_GetHistory(t *testing.T) {
	serverUUID := "test-server-123"
	_ = time.Now()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/hardware-inventory/"+serverUUID+"/history", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("page"))

		inventories := []*HardwareInventoryInfo{
			{
				Manufacturer: "Dell Inc.",
				Model:        "PowerEdge R740",
				
			},
			{
				Manufacturer: "Dell Inc.",
				Model:        "PowerEdge R740",
				
			},
		}

		response := map[string]interface{}{
			"success": true,
			"data":    inventories,
			"meta": map[string]interface{}{
				"total":       2,
				"page":        1,
				"limit":       25,
				"total_pages": 1,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	require.NoError(t, err)

	inventories, meta, err := client.HardwareInventory.GetHistory(context.Background(), serverUUID, &ListOptions{
		Page: 1,
	})
	require.NoError(t, err)
	assert.Len(t, inventories, 2)
	assert.NotNil(t, meta)
}

func TestHardwareInventoryService_GetChanges(t *testing.T) {
	serverUUID := "test-server-123"
	now := time.Now()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/hardware-inventory/"+serverUUID+"/changes", r.URL.Path)
		assert.NotEmpty(t, r.URL.Query().Get("start"))
		assert.NotEmpty(t, r.URL.Query().Get("end"))

		changes := []HardwareChange{
			{
				ID:            1,
				ServerUUID:    serverUUID,
				ComponentType: "memory",
				ChangeType:    "added",
				NewValue:      "32GB DDR4 Module",
				ChangedAt:     &CustomTime{Time: now},
				Details:       "Added new memory module",
			},
			{
				ID:            2,
				ServerUUID:    serverUUID,
				ComponentType: "storage",
				ChangeType:    "removed",
				OldValue:      "500GB SSD",
				ChangedAt:     &CustomTime{Time: now.Add(-1 * time.Hour)},
				Details:       "Removed old SSD",
			},
		}

		response := map[string]interface{}{
			"success": true,
			"data":    changes,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	require.NoError(t, err)

	timeRange := &QueryTimeRange{
		Start: now.Add(-7 * 24 * time.Hour),
		End:   now,
	}

	changes, err := client.HardwareInventory.GetChanges(context.Background(), serverUUID, timeRange)
	require.NoError(t, err)
	assert.Len(t, changes, 2)
	assert.Equal(t, "memory", changes[0].ComponentType)
	assert.Equal(t, "added", changes[0].ChangeType)
}

func TestHardwareInventoryService_Search(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/hardware-inventory/search", r.URL.Path)

		var search HardwareSearch
		err := json.NewDecoder(r.Body).Decode(&search)
		require.NoError(t, err)
		assert.Equal(t, "Dell Inc.", search.Manufacturer)
		assert.Equal(t, "PowerEdge", search.Model)

		inventories := []*HardwareInventoryInfo{
			{
				Manufacturer: "Dell Inc.",
				Model:        "PowerEdge R740",
			},
			{
				Manufacturer: "Dell Inc.",
				Model:        "PowerEdge R640",
			},
		}

		response := map[string]interface{}{
			"success": true,
			"data":    inventories,
			"meta": map[string]interface{}{
				"total": 2,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	require.NoError(t, err)

	inventories, meta, err := client.HardwareInventory.Search(context.Background(), &HardwareSearch{
		Manufacturer: "Dell Inc.",
		Model:        "PowerEdge",
	})
	require.NoError(t, err)
	assert.Len(t, inventories, 2)
	assert.NotNil(t, meta)
}

func TestHardwareInventoryService_Export(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/hardware-inventory/export", r.URL.Path)

		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "csv", body["format"])

		csvData := "Server UUID,Manufacturer,Model\nserver-1,Dell Inc.,PowerEdge R740\n"
		w.Header().Set("Content-Type", "text/csv")
		w.Write([]byte(csvData))
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	require.NoError(t, err)

	data, err := client.HardwareInventory.Export(context.Background(), "csv", []string{"server-1", "server-2"})
	require.NoError(t, err)
	assert.NotNil(t, data)
	assert.Contains(t, string(data), "PowerEdge R740")
}

func TestHardwareInventoryService_GetHardwareInventory(t *testing.T) {
	serverUUID := "test-server-123"
	now := time.Now()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v2/hardware/inventory/"+serverUUID, r.URL.Path)
		assert.Equal(t, "2024-01-01T00:00:00Z", r.URL.Query().Get("start"))
		assert.Equal(t, "2024-01-31T23:59:59Z", r.URL.Query().Get("end"))

		response := map[string]interface{}{
			"data": []HardwareInventoryRecord{
				{
					ID:             1,
					ServerUUID:     serverUUID,
					OrganizationID: 1,
					Timestamp:      now,
					Hardware: HardwareInventoryInfo{
						Manufacturer: "Dell Inc.",
						Model:        "PowerEdge R740",
						CPUs: []CPUInfo{
							{
								Manufacturer: "Intel",
								Model:        "Xeon Gold 6230",
								Cores:        20,
							},
						},
					},
					CollectionMethod: "dmidecode",
					DetectionTool:    "lshw",
					CreatedAt:        now,
					UpdatedAt:        now,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	require.NoError(t, err)

	timeRange := &TimeRange{
		Start: "2024-01-01T00:00:00Z",
		End:   "2024-01-31T23:59:59Z",
	}

	records, err := client.HardwareInventory.GetHardwareInventory(context.Background(), serverUUID, timeRange)
	require.NoError(t, err)
	assert.NotNil(t, records)
	assert.Len(t, records, 1)
	assert.Equal(t, serverUUID, records[0].ServerUUID)
	assert.Equal(t, "Dell Inc.", records[0].Hardware.Manufacturer)
}

func TestHardwareInventoryService_GetLatestHardwareInventory(t *testing.T) {
	serverUUID := "test-server-123"
	now := time.Now()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v2/hardware/inventory/"+serverUUID+"/latest", r.URL.Path)

		response := map[string]interface{}{
			"data": HardwareInventoryRecord{
				ID:             1,
				ServerUUID:     serverUUID,
				OrganizationID: 1,
				Timestamp:      now,
				Hardware: HardwareInventoryInfo{
					Manufacturer: "HP",
					Model:        "ProLiant DL380 Gen10",
					CPUs: []CPUInfo{
						{
							Manufacturer: "Intel",
							Model:        "Xeon Silver 4210",
							Cores:        10,
							Threads:      20,
						},
					},
					Memory: &MemoryInfo{
						TotalCapacity:  68719476736, // 64GB
						AvailableSlots: 24,
						UsedSlots:      8,
						ECCSupported:   true,
					},
				},
				CollectionMethod: "dmidecode",
				DetectionTool:    "hwinfo",
				CreatedAt:        now,
				UpdatedAt:        now,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	require.NoError(t, err)

	record, err := client.HardwareInventory.GetLatestHardwareInventory(context.Background(), serverUUID)
	require.NoError(t, err)
	assert.NotNil(t, record)
	assert.Equal(t, serverUUID, record.ServerUUID)
	assert.Equal(t, "HP", record.Hardware.Manufacturer)
	assert.Equal(t, "ProLiant DL380 Gen10", record.Hardware.Model)
	assert.Len(t, record.Hardware.CPUs, 1)
}

func TestHardwareInventoryService_ListHardwareHistory(t *testing.T) {
	serverUUID := "test-server-123"
	now := time.Now()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v2/hardware/inventory/"+serverUUID+"/history", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "50", r.URL.Query().Get("limit"))
		assert.Equal(t, "2024-01-01T00:00:00Z", r.URL.Query().Get("start_time"))
		assert.Equal(t, "2024-01-31T23:59:59Z", r.URL.Query().Get("end_time"))

		response := map[string]interface{}{
			"data": []HardwareInventoryRecord{
				{
					ID:             1,
					ServerUUID:     serverUUID,
					OrganizationID: 1,
					Timestamp:      now.Add(-24 * time.Hour),
					Hardware: HardwareInventoryInfo{
						Manufacturer: "Dell Inc.",
						Model:        "PowerEdge R740",
					},
					CollectionMethod: "dmidecode",
					DetectionTool:    "lshw",
					CreatedAt:        now.Add(-24 * time.Hour),
					UpdatedAt:        now.Add(-24 * time.Hour),
				},
				{
					ID:             2,
					ServerUUID:     serverUUID,
					OrganizationID: 1,
					Timestamp:      now,
					Hardware: HardwareInventoryInfo{
						Manufacturer: "Dell Inc.",
						Model:        "PowerEdge R740",
						// Updated with additional GPU
						GPUs: []GPUInfo{
							{
								Manufacturer: "NVIDIA",
								Model:        "Tesla V100",
								MemorySize:   34359738368, // 32GB
							},
						},
					},
					CollectionMethod: "dmidecode",
					DetectionTool:    "lshw",
					CreatedAt:        now,
					UpdatedAt:        now,
				},
			},
			"meta": PaginationMeta{
				Page:       1,
				Limit:      50,
				TotalItems: 2,
				TotalPages: 1,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
	})
	require.NoError(t, err)

	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

	opts := &HardwareInventoryListOptions{
		ListOptions: ListOptions{
			Page:  1,
			Limit: 50,
		},
		StartTime: &startTime,
		EndTime:   &endTime,
	}

	records, meta, err := client.HardwareInventory.ListHardwareHistory(context.Background(), serverUUID, opts)
	require.NoError(t, err)
	assert.NotNil(t, records)
	assert.NotNil(t, meta)
	assert.Len(t, records, 2)
	assert.Equal(t, 1, meta.Page)
	assert.Equal(t, 50, meta.Limit)

	// Check that hardware was updated (GPU added)
	assert.Len(t, records[0].Hardware.GPUs, 0)
	assert.Len(t, records[1].Hardware.GPUs, 1)
}

func TestHardwareInventoryService_ServerNotFound(t *testing.T) {
	serverUUID := "nonexistent-server"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, serverUUID) {
			w.WriteHeader(http.StatusNotFound)
			response := APIError{
				Status:    "error",
				Message:   "Server not found",
				ErrorCode: "not_found",
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Token: "test-token",
		},
		// Disable retry for tests
		RetryCount: 0,
	})
	require.NoError(t, err)

	record, err := client.HardwareInventory.GetLatestHardwareInventory(context.Background(), serverUUID)
	assert.Error(t, err)
	assert.Nil(t, record)

	// The hardware_inventory.go now properly converts API errors to NotFoundError
	notFoundErr, ok := err.(*NotFoundError)
	assert.True(t, ok, "Expected NotFoundError but got: %T - %v", err, err)
	assert.Equal(t, "resource not found", notFoundErr.Message)
}

func TestHardwareInventoryService_ValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/hardware/inventory" {
			w.WriteHeader(http.StatusBadRequest)
			response := APIError{
				Status:    "error",
				Message:   "Validation failed",
				ErrorCode: "validation_error",
				Details:   "collection_method: Collection method is required; detection_tool: Detection tool is required",
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth: AuthConfig{
			ServerUUID:   "test-uuid",
			ServerSecret: "test-secret",
		},
	})
	require.NoError(t, err)

	req := &HardwareInventoryRequest{
		ServerUUID: "test-server-123",
		Hardware:   HardwareInventoryInfo{}, // Missing required fields
	}

	resp, err := client.HardwareInventory.Submit(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, resp)

	// The error might be wrapped, so check the error message
	assert.Contains(t, err.Error(), "validation_error")
	assert.Contains(t, err.Error(), "Validation failed")
}
