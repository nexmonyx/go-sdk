package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Schedule CRUD Operations Tests
// =============================================================================

func TestSchedulesService_CreateSchedule_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/schedules", r.URL.Path)

		var reqBody CreateScheduleRequest
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		require.NoError(t, err)
		assert.Equal(t, "Daily Backup", reqBody.Name)
		assert.Equal(t, "0 0 * * *", reqBody.CronExpression)
		assert.Equal(t, ScheduleTargetJob, reqBody.TargetType)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Schedule created successfully",
			"data": map[string]interface{}{
				"id":              1,
				"schedule_uuid":   "sched-uuid-123",
				"organization_id": 1,
				"name":            "Daily Backup",
				"cron_expression": "0 0 * * *",
				"timezone":        "UTC",
				"target_type":     "job",
				"target_config":   map[string]interface{}{"job_id": 123},
				"enabled":         true,
				"status":          "active",
				"created_at":      "2024-01-01T00:00:00Z",
				"updated_at":      "2024-01-01T00:00:00Z",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	enabled := true
	schedule, resp, err := client.Schedules.CreateSchedule(context.Background(), &CreateScheduleRequest{
		Name:           "Daily Backup",
		CronExpression: "0 0 * * *",
		TargetType:     ScheduleTargetJob,
		TargetConfig:   map[string]interface{}{"job_id": 123},
		Enabled:        &enabled,
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, schedule)
	assert.Equal(t, uint(1), schedule.ID)
	assert.Equal(t, "Daily Backup", schedule.Name)
	assert.Equal(t, "0 0 * * *", schedule.CronExpression)
	assert.Equal(t, ScheduleTargetJob, schedule.TargetType)
	assert.True(t, schedule.Enabled)
}

func TestSchedulesService_CreateSchedule_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/schedules", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Invalid cron expression",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	schedule, _, err := client.Schedules.CreateSchedule(context.Background(), &CreateScheduleRequest{
		Name:           "Bad Schedule",
		CronExpression: "invalid",
		TargetType:     ScheduleTargetJob,
		TargetConfig:   map[string]interface{}{},
	})

	assert.Error(t, err)
	assert.Nil(t, schedule)
}

func TestSchedulesService_ListSchedules_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/schedules", r.URL.Path)
		assert.Equal(t, "active", r.URL.Query().Get("status"))
		assert.Equal(t, "1", r.URL.Query().Get("page"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Schedules retrieved successfully",
			"data": []map[string]interface{}{
				{
					"id":              1,
					"schedule_uuid":   "sched-uuid-1",
					"organization_id": 1,
					"name":            "Schedule 1",
					"cron_expression": "0 0 * * *",
					"timezone":        "UTC",
					"target_type":     "job",
					"enabled":         true,
					"status":          "active",
					"created_at":      "2024-01-01T00:00:00Z",
					"updated_at":      "2024-01-01T00:00:00Z",
				},
				{
					"id":              2,
					"schedule_uuid":   "sched-uuid-2",
					"organization_id": 1,
					"name":            "Schedule 2",
					"cron_expression": "0 12 * * *",
					"timezone":        "UTC",
					"target_type":     "report",
					"enabled":         true,
					"status":          "active",
					"created_at":      "2024-01-01T00:00:00Z",
					"updated_at":      "2024-01-01T00:00:00Z",
				},
			},
			"meta": map[string]interface{}{
				"page":        1,
				"page_size":   10,
				"total":       2,
				"total_pages": 1,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, resp, err := client.Schedules.ListSchedules(context.Background(), &ListSchedulesOptions{
		Page:   1,
		Status: "active",
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, result)
	assert.Len(t, result.Schedules, 2)
	assert.Equal(t, "Schedule 1", result.Schedules[0].Name)
	assert.Equal(t, "Schedule 2", result.Schedules[1].Name)
}

func TestSchedulesService_ListSchedules_NilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/schedules", r.URL.Path)
		assert.Empty(t, r.URL.RawQuery)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Schedules retrieved successfully",
			"data":    []map[string]interface{}{},
			"meta": map[string]interface{}{
				"page":        1,
				"page_size":   10,
				"total":       0,
				"total_pages": 0,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, resp, err := client.Schedules.ListSchedules(context.Background(), nil)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, result)
}

func TestSchedulesService_ListSchedules_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Database connection failed",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, _, err := client.Schedules.ListSchedules(context.Background(), nil)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestSchedulesService_GetSchedule_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/schedules/123", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Schedule retrieved successfully",
			"data": map[string]interface{}{
				"id":              123,
				"schedule_uuid":   "sched-uuid-123",
				"organization_id": 1,
				"name":            "Daily Backup",
				"cron_expression": "0 0 * * *",
				"timezone":        "America/New_York",
				"target_type":     "job",
				"target_config":   map[string]interface{}{"job_id": 456},
				"enabled":         true,
				"max_retries":     3,
				"retry_policy":    "exponential",
				"timeout_minutes": 30,
				"status":          "active",
				"run_count":       100,
				"success_count":   95,
				"failure_count":   5,
				"created_at":      "2024-01-01T00:00:00Z",
				"updated_at":      "2024-01-01T00:00:00Z",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	schedule, resp, err := client.Schedules.GetSchedule(context.Background(), 123)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, schedule)
	assert.Equal(t, uint(123), schedule.ID)
	assert.Equal(t, "Daily Backup", schedule.Name)
	assert.Equal(t, "America/New_York", schedule.Timezone)
	assert.Equal(t, int64(100), schedule.RunCount)
	assert.Equal(t, int64(95), schedule.SuccessCount)
}

func TestSchedulesService_GetSchedule_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/schedules/999", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Schedule not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	schedule, _, err := client.Schedules.GetSchedule(context.Background(), 999)

	assert.Error(t, err)
	assert.Nil(t, schedule)
}

func TestSchedulesService_UpdateSchedule_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/schedules/123", r.URL.Path)

		var reqBody UpdateScheduleRequest
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		require.NoError(t, err)
		assert.NotNil(t, reqBody.Name)
		assert.Equal(t, "Updated Schedule", *reqBody.Name)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Schedule updated successfully",
			"data": map[string]interface{}{
				"id":              123,
				"schedule_uuid":   "sched-uuid-123",
				"organization_id": 1,
				"name":            "Updated Schedule",
				"cron_expression": "0 0 * * *",
				"timezone":        "UTC",
				"target_type":     "job",
				"enabled":         true,
				"status":          "active",
				"created_at":      "2024-01-01T00:00:00Z",
				"updated_at":      "2024-01-02T00:00:00Z",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	newName := "Updated Schedule"
	schedule, resp, err := client.Schedules.UpdateSchedule(context.Background(), 123, &UpdateScheduleRequest{
		Name: &newName,
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, schedule)
	assert.Equal(t, "Updated Schedule", schedule.Name)
}

func TestSchedulesService_UpdateSchedule_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Invalid cron expression",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	invalidCron := "invalid-cron"
	schedule, _, err := client.Schedules.UpdateSchedule(context.Background(), 123, &UpdateScheduleRequest{
		CronExpression: &invalidCron,
	})

	assert.Error(t, err)
	assert.Nil(t, schedule)
}

func TestSchedulesService_DeleteSchedule_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/schedules/123", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Schedule deleted successfully",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	resp, err := client.Schedules.DeleteSchedule(context.Background(), 123)

	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestSchedulesService_DeleteSchedule_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Schedule not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	_, err := client.Schedules.DeleteSchedule(context.Background(), 999)

	assert.Error(t, err)
}

// =============================================================================
// Schedule Control Operations Tests
// =============================================================================

func TestSchedulesService_EnableSchedule_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/schedules/123/enable", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Schedule enabled successfully",
			"data": map[string]interface{}{
				"id":              123,
				"schedule_uuid":   "sched-uuid-123",
				"organization_id": 1,
				"name":            "Daily Backup",
				"cron_expression": "0 0 * * *",
				"timezone":        "UTC",
				"target_type":     "job",
				"enabled":         true,
				"status":          "active",
				"created_at":      "2024-01-01T00:00:00Z",
				"updated_at":      "2024-01-01T00:00:00Z",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	schedule, resp, err := client.Schedules.EnableSchedule(context.Background(), 123)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, schedule)
	assert.True(t, schedule.Enabled)
}

func TestSchedulesService_EnableSchedule_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Schedule not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	schedule, _, err := client.Schedules.EnableSchedule(context.Background(), 999)

	assert.Error(t, err)
	assert.Nil(t, schedule)
}

func TestSchedulesService_DisableSchedule_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/schedules/123/disable", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Schedule disabled successfully",
			"data": map[string]interface{}{
				"id":              123,
				"schedule_uuid":   "sched-uuid-123",
				"organization_id": 1,
				"name":            "Daily Backup",
				"cron_expression": "0 0 * * *",
				"timezone":        "UTC",
				"target_type":     "job",
				"enabled":         false,
				"status":          "paused",
				"created_at":      "2024-01-01T00:00:00Z",
				"updated_at":      "2024-01-01T00:00:00Z",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	schedule, resp, err := client.Schedules.DisableSchedule(context.Background(), 123)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, schedule)
	assert.False(t, schedule.Enabled)
}

func TestSchedulesService_DisableSchedule_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Schedule not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	schedule, _, err := client.Schedules.DisableSchedule(context.Background(), 999)

	assert.Error(t, err)
	assert.Nil(t, schedule)
}

func TestSchedulesService_TriggerSchedule_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/schedules/123/trigger", r.URL.Path)

		var reqBody TriggerScheduleRequest
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		require.NoError(t, err)
		assert.Equal(t, "Manual test run", reqBody.Reason)
		assert.True(t, reqBody.SkipJitter)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Schedule triggered successfully",
			"data": map[string]interface{}{
				"id":               1,
				"execution_uuid":   "exec-uuid-123",
				"schedule_id":      123,
				"organization_id":  1,
				"trigger_time":     "2024-01-01T12:00:00Z",
				"scheduled_time":   "2024-01-01T12:00:00Z",
				"status":           "pending",
				"manual_trigger":   true,
				"retry_count":      0,
				"created_at":       "2024-01-01T12:00:00Z",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	execution, resp, err := client.Schedules.TriggerSchedule(context.Background(), 123, &TriggerScheduleRequest{
		SkipJitter: true,
		Reason:     "Manual test run",
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, execution)
	assert.Equal(t, uint(1), execution.ID)
	assert.True(t, execution.ManualTrigger)
	assert.Equal(t, ScheduleExecutionPending, execution.Status)
}

func TestSchedulesService_TriggerSchedule_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Schedule is disabled",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	execution, _, err := client.Schedules.TriggerSchedule(context.Background(), 123, &TriggerScheduleRequest{})

	assert.Error(t, err)
	assert.Nil(t, execution)
}

// =============================================================================
// Schedule History and Preview Tests
// =============================================================================

func TestSchedulesService_GetExecutions_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/schedules/123/executions", r.URL.Path)
		assert.Equal(t, "completed", r.URL.Query().Get("status"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Executions retrieved successfully",
			"data": []map[string]interface{}{
				{
					"id":              1,
					"execution_uuid":  "exec-uuid-1",
					"schedule_id":     123,
					"organization_id": 1,
					"trigger_time":    "2024-01-01T00:00:00Z",
					"scheduled_time":  "2024-01-01T00:00:00Z",
					"status":          "completed",
					"started_at":      "2024-01-01T00:00:01Z",
					"completed_at":    "2024-01-01T00:01:00Z",
					"duration_ms":     59000,
					"retry_count":     0,
					"manual_trigger":  false,
					"created_at":      "2024-01-01T00:00:00Z",
				},
				{
					"id":              2,
					"execution_uuid":  "exec-uuid-2",
					"schedule_id":     123,
					"organization_id": 1,
					"trigger_time":    "2024-01-02T00:00:00Z",
					"scheduled_time":  "2024-01-02T00:00:00Z",
					"status":          "completed",
					"started_at":      "2024-01-02T00:00:01Z",
					"completed_at":    "2024-01-02T00:00:30Z",
					"duration_ms":     29000,
					"retry_count":     0,
					"manual_trigger":  false,
					"created_at":      "2024-01-02T00:00:00Z",
				},
			},
			"meta": map[string]interface{}{
				"page":        1,
				"page_size":   10,
				"total":       2,
				"total_pages": 1,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, resp, err := client.Schedules.GetExecutions(context.Background(), 123, &ListExecutionsOptions{
		Status: "completed",
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, result)
	assert.Len(t, result.Executions, 2)
	assert.Equal(t, ScheduleExecutionCompleted, result.Executions[0].Status)
}

func TestSchedulesService_GetExecutions_NilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/schedules/123/executions", r.URL.Path)
		assert.Empty(t, r.URL.RawQuery)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Executions retrieved",
			"data":    []map[string]interface{}{},
			"meta": map[string]interface{}{
				"page":        1,
				"page_size":   10,
				"total":       0,
				"total_pages": 0,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, resp, err := client.Schedules.GetExecutions(context.Background(), 123, nil)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, result)
}

func TestSchedulesService_GetExecutions_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Schedule not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, _, err := client.Schedules.GetExecutions(context.Background(), 999, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestSchedulesService_GetStatistics_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/schedules/123/statistics", r.URL.Path)
		assert.Equal(t, "24h", r.URL.Query().Get("period"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Statistics retrieved successfully",
			"data": map[string]interface{}{
				"schedule_id":            123,
				"period":                 "24h",
				"total_executions":       24,
				"successful_executions":  22,
				"failed_executions":      2,
				"success_rate":           91.67,
				"avg_duration_ms":        45000,
				"min_duration_ms":        30000,
				"max_duration_ms":        60000,
				"last_execution_at":      "2024-01-01T23:00:00Z",
				"next_execution_at":      "2024-01-02T00:00:00Z",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	stats, resp, err := client.Schedules.GetStatistics(context.Background(), 123, "24h")

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, stats)
	assert.Equal(t, uint(123), stats.ScheduleID)
	assert.Equal(t, "24h", stats.Period)
	assert.Equal(t, int64(24), stats.TotalExecutions)
	assert.Equal(t, int64(22), stats.SuccessfulExecutions)
	assert.InDelta(t, 91.67, stats.SuccessRate, 0.01)
}

func TestSchedulesService_GetStatistics_EmptyPeriod(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/schedules/123/statistics", r.URL.Path)
		assert.Empty(t, r.URL.Query().Get("period"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Statistics retrieved",
			"data": map[string]interface{}{
				"schedule_id":            123,
				"period":                 "7d",
				"total_executions":       168,
				"successful_executions":  165,
				"failed_executions":      3,
				"success_rate":           98.21,
				"avg_duration_ms":        42000,
				"min_duration_ms":        25000,
				"max_duration_ms":        75000,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	stats, resp, err := client.Schedules.GetStatistics(context.Background(), 123, "")

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, stats)
}

func TestSchedulesService_GetStatistics_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Schedule not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	stats, _, err := client.Schedules.GetStatistics(context.Background(), 999, "24h")

	assert.Error(t, err)
	assert.Nil(t, stats)
}

func TestSchedulesService_GetNextRuns_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/schedules/123/next-runs", r.URL.Path)
		assert.Equal(t, "5", r.URL.Query().Get("count"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Next runs retrieved successfully",
			"data": map[string]interface{}{
				"schedule_id":     123,
				"cron_expression": "0 0 * * *",
				"timezone":        "UTC",
				"next_runs": []map[string]interface{}{
					{"run_time": "2024-01-02T00:00:00Z", "run_number": 1},
					{"run_time": "2024-01-03T00:00:00Z", "run_number": 2},
					{"run_time": "2024-01-04T00:00:00Z", "run_number": 3},
					{"run_time": "2024-01-05T00:00:00Z", "run_number": 4},
					{"run_time": "2024-01-06T00:00:00Z", "run_number": 5},
				},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, resp, err := client.Schedules.GetNextRuns(context.Background(), 123, 5)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, result)
	assert.Equal(t, uint(123), result.ScheduleID)
	assert.Equal(t, "0 0 * * *", result.CronExpression)
	assert.Len(t, result.NextRuns, 5)
	assert.Equal(t, 1, result.NextRuns[0].RunNumber)
}

func TestSchedulesService_GetNextRuns_ZeroCount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/schedules/123/next-runs", r.URL.Path)
		assert.Empty(t, r.URL.Query().Get("count"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Next runs retrieved",
			"data": map[string]interface{}{
				"schedule_id":     123,
				"cron_expression": "0 0 * * *",
				"timezone":        "UTC",
				"next_runs":       []map[string]interface{}{},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, resp, err := client.Schedules.GetNextRuns(context.Background(), 123, 0)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, result)
}

func TestSchedulesService_GetNextRuns_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Schedule not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, _, err := client.Schedules.GetNextRuns(context.Background(), 999, 5)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// =============================================================================
// Execution Management Tests
// =============================================================================

func TestSchedulesService_ExecutionCallback_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/schedules/123/executions/456/callback", r.URL.Path)

		var reqBody ExecutionCallbackRequest
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		require.NoError(t, err)
		assert.Equal(t, "completed", reqBody.Status)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Execution callback processed",
			"data": map[string]interface{}{
				"id":              456,
				"execution_uuid":  "exec-uuid-456",
				"schedule_id":     123,
				"organization_id": 1,
				"trigger_time":    "2024-01-01T00:00:00Z",
				"scheduled_time":  "2024-01-01T00:00:00Z",
				"status":          "completed",
				"started_at":      "2024-01-01T00:00:01Z",
				"completed_at":    "2024-01-01T00:01:00Z",
				"duration_ms":     59000,
				"retry_count":     0,
				"manual_trigger":  false,
				"created_at":      "2024-01-01T00:00:00Z",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	durationMs := 59000
	execution, resp, err := client.Schedules.ExecutionCallback(context.Background(), 123, 456, &ExecutionCallbackRequest{
		Status:     "completed",
		DurationMs: &durationMs,
		Result:     map[string]interface{}{"message": "Success"},
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, execution)
	assert.Equal(t, uint(456), execution.ID)
	assert.Equal(t, ScheduleExecutionCompleted, execution.Status)
}

func TestSchedulesService_ExecutionCallback_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Execution not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	execution, _, err := client.Schedules.ExecutionCallback(context.Background(), 123, 999, &ExecutionCallbackRequest{
		Status: "completed",
	})

	assert.Error(t, err)
	assert.Nil(t, execution)
}

func TestSchedulesService_UpdateExecution_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "/v1/schedules/123/executions/456", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Execution updated",
			"data": map[string]interface{}{
				"id":              456,
				"execution_uuid":  "exec-uuid-456",
				"schedule_id":     123,
				"organization_id": 1,
				"trigger_time":    "2024-01-01T00:00:00Z",
				"scheduled_time":  "2024-01-01T00:00:00Z",
				"status":          "failed",
				"error_message":   "Connection timeout",
				"retry_count":     2,
				"manual_trigger":  false,
				"created_at":      "2024-01-01T00:00:00Z",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	status := "failed"
	errMsg := "Connection timeout"
	execution, resp, err := client.Schedules.UpdateExecution(context.Background(), 123, 456, &UpdateExecutionRequest{
		Status:       &status,
		ErrorMessage: &errMsg,
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, execution)
	assert.Equal(t, ScheduleExecutionFailed, execution.Status)
}

func TestSchedulesService_UpdateExecution_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Execution not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	execution, _, err := client.Schedules.UpdateExecution(context.Background(), 123, 999, &UpdateExecutionRequest{})

	assert.Error(t, err)
	assert.Nil(t, execution)
}

func TestSchedulesService_GetExecution_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/schedules/123/executions/456", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Execution retrieved",
			"data": map[string]interface{}{
				"id":                   456,
				"execution_uuid":       "exec-uuid-456",
				"schedule_id":          123,
				"organization_id":      1,
				"trigger_time":         "2024-01-01T00:00:00Z",
				"scheduled_time":       "2024-01-01T00:00:00Z",
				"jitter_applied_ms":    500,
				"status":               "completed",
				"started_at":           "2024-01-01T00:00:01Z",
				"completed_at":         "2024-01-01T00:01:00Z",
				"duration_ms":          59000,
				"target_resource_id":   "job-123",
				"target_resource_type": "job",
				"retry_count":          0,
				"result":               map[string]interface{}{"records_processed": 1000},
				"manual_trigger":       false,
				"created_at":           "2024-01-01T00:00:00Z",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	execution, resp, err := client.Schedules.GetExecution(context.Background(), 123, 456)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, execution)
	assert.Equal(t, uint(456), execution.ID)
	assert.Equal(t, ScheduleExecutionCompleted, execution.Status)
	assert.Equal(t, 500, execution.JitterAppliedMs)
}

func TestSchedulesService_GetExecution_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Execution not found",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	execution, _, err := client.Schedules.GetExecution(context.Background(), 123, 999)

	assert.Error(t, err)
	assert.Nil(t, execution)
}

// =============================================================================
// Utility Operations Tests
// =============================================================================

func TestSchedulesService_ValidateCron_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/schedules/validate-cron", r.URL.Path)

		var reqBody ValidateCronRequest
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		require.NoError(t, err)
		assert.Equal(t, "0 0 * * *", reqBody.CronExpression)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Cron expression validated",
			"data": map[string]interface{}{
				"valid":       true,
				"expression":  "0 0 * * *",
				"description": "Every day at midnight",
				"next_runs": []string{
					"2024-01-02T00:00:00Z",
					"2024-01-03T00:00:00Z",
					"2024-01-04T00:00:00Z",
				},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, resp, err := client.Schedules.ValidateCron(context.Background(), &ValidateCronRequest{
		CronExpression: "0 0 * * *",
		PreviewCount:   3,
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, result)
	assert.True(t, result.Valid)
	assert.Equal(t, "0 0 * * *", result.Expression)
	assert.Equal(t, "Every day at midnight", result.Description)
	assert.Len(t, result.NextRuns, 3)
}

func TestSchedulesService_ValidateCron_Invalid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Cron expression validated",
			"data": map[string]interface{}{
				"valid":      false,
				"expression": "invalid-cron",
				"error":      "Invalid cron expression format",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	result, resp, err := client.Schedules.ValidateCron(context.Background(), &ValidateCronRequest{
		CronExpression: "invalid-cron",
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.NotEmpty(t, result.Error)
}

func TestSchedulesService_ValidateCron_Error(t *testing.T) {
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
	result, _, err := client.Schedules.ValidateCron(context.Background(), &ValidateCronRequest{
		CronExpression: "0 0 * * *",
	})

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestSchedulesService_ListTimezones_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/schedules/timezones", r.URL.Path)
		assert.Equal(t, "America", r.URL.Query().Get("region"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Timezones retrieved",
			"data": []map[string]interface{}{
				{"name": "America/New_York", "offset": "-05:00", "region": "America"},
				{"name": "America/Chicago", "offset": "-06:00", "region": "America"},
				{"name": "America/Los_Angeles", "offset": "-08:00", "region": "America"},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	timezones, resp, err := client.Schedules.ListTimezones(context.Background(), "America")

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, timezones, 3)
	assert.Equal(t, "America/New_York", timezones[0].Name)
	assert.Equal(t, "-05:00", timezones[0].Offset)
}

func TestSchedulesService_ListTimezones_NoRegion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/schedules/timezones", r.URL.Path)
		assert.Empty(t, r.URL.Query().Get("region"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Timezones retrieved",
			"data": []map[string]interface{}{
				{"name": "UTC", "offset": "+00:00"},
				{"name": "Europe/London", "offset": "+00:00", "region": "Europe"},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	timezones, resp, err := client.Schedules.ListTimezones(context.Background(), "")

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, timezones)
}

func TestSchedulesService_ListTimezones_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Failed to retrieve timezones",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	timezones, _, err := client.Schedules.ListTimezones(context.Background(), "")

	assert.Error(t, err)
	assert.Nil(t, timezones)
}

// =============================================================================
// Helper Methods Tests
// =============================================================================

func TestSchedule_IsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		schedule Schedule
		expected bool
	}{
		{
			name:     "enabled and active",
			schedule: Schedule{Enabled: true, Status: ScheduleStatusActive},
			expected: true,
		},
		{
			name:     "enabled but paused",
			schedule: Schedule{Enabled: true, Status: ScheduleStatusPaused},
			expected: false,
		},
		{
			name:     "disabled",
			schedule: Schedule{Enabled: false, Status: ScheduleStatusActive},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.schedule.IsEnabled())
		})
	}
}

func TestSchedule_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		schedule Schedule
		expected bool
	}{
		{
			name:     "active status",
			schedule: Schedule{Status: ScheduleStatusActive},
			expected: true,
		},
		{
			name:     "paused status",
			schedule: Schedule{Status: ScheduleStatusPaused},
			expected: false,
		},
		{
			name:     "error status",
			schedule: Schedule{Status: ScheduleStatusError},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.schedule.IsActive())
		})
	}
}

func TestSchedule_IsPaused(t *testing.T) {
	tests := []struct {
		name     string
		schedule Schedule
		expected bool
	}{
		{
			name:     "paused status",
			schedule: Schedule{Status: ScheduleStatusPaused},
			expected: true,
		},
		{
			name:     "active status",
			schedule: Schedule{Status: ScheduleStatusActive},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.schedule.IsPaused())
		})
	}
}

func TestSchedule_HasErrors(t *testing.T) {
	errorMsg := "Connection failed"
	tests := []struct {
		name     string
		schedule Schedule
		expected bool
	}{
		{
			name:     "error status",
			schedule: Schedule{Status: ScheduleStatusError},
			expected: true,
		},
		{
			name:     "has last run error",
			schedule: Schedule{Status: ScheduleStatusActive, LastRunError: &errorMsg},
			expected: true,
		},
		{
			name:     "no errors",
			schedule: Schedule{Status: ScheduleStatusActive},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.schedule.HasErrors())
		})
	}
}

func TestSchedule_GetSuccessRate(t *testing.T) {
	tests := []struct {
		name     string
		schedule Schedule
		expected float64
	}{
		{
			name:     "no runs",
			schedule: Schedule{RunCount: 0, SuccessCount: 0},
			expected: 0,
		},
		{
			name:     "all successful",
			schedule: Schedule{RunCount: 100, SuccessCount: 100},
			expected: 100,
		},
		{
			name:     "50% success",
			schedule: Schedule{RunCount: 100, SuccessCount: 50},
			expected: 50,
		},
		{
			name:     "90% success",
			schedule: Schedule{RunCount: 100, SuccessCount: 90},
			expected: 90,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.InDelta(t, tt.expected, tt.schedule.GetSuccessRate(), 0.01)
		})
	}
}

func TestScheduleExecution_IsComplete(t *testing.T) {
	tests := []struct {
		name      string
		execution ScheduleExecution
		expected  bool
	}{
		{
			name:      "completed",
			execution: ScheduleExecution{Status: ScheduleExecutionCompleted},
			expected:  true,
		},
		{
			name:      "failed",
			execution: ScheduleExecution{Status: ScheduleExecutionFailed},
			expected:  true,
		},
		{
			name:      "cancelled",
			execution: ScheduleExecution{Status: ScheduleExecutionCancelled},
			expected:  true,
		},
		{
			name:      "timed out",
			execution: ScheduleExecution{Status: ScheduleExecutionTimedOut},
			expected:  true,
		},
		{
			name:      "running",
			execution: ScheduleExecution{Status: ScheduleExecutionRunning},
			expected:  false,
		},
		{
			name:      "pending",
			execution: ScheduleExecution{Status: ScheduleExecutionPending},
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.execution.IsComplete())
		})
	}
}

func TestScheduleExecution_IsSuccessful(t *testing.T) {
	tests := []struct {
		name      string
		execution ScheduleExecution
		expected  bool
	}{
		{
			name:      "completed",
			execution: ScheduleExecution{Status: ScheduleExecutionCompleted},
			expected:  true,
		},
		{
			name:      "failed",
			execution: ScheduleExecution{Status: ScheduleExecutionFailed},
			expected:  false,
		},
		{
			name:      "running",
			execution: ScheduleExecution{Status: ScheduleExecutionRunning},
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.execution.IsSuccessful())
		})
	}
}

func TestScheduleExecution_IsFailed(t *testing.T) {
	tests := []struct {
		name      string
		execution ScheduleExecution
		expected  bool
	}{
		{
			name:      "failed",
			execution: ScheduleExecution{Status: ScheduleExecutionFailed},
			expected:  true,
		},
		{
			name:      "timed out",
			execution: ScheduleExecution{Status: ScheduleExecutionTimedOut},
			expected:  true,
		},
		{
			name:      "completed",
			execution: ScheduleExecution{Status: ScheduleExecutionCompleted},
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.execution.IsFailed())
		})
	}
}

func TestScheduleExecution_IsRunning(t *testing.T) {
	tests := []struct {
		name      string
		execution ScheduleExecution
		expected  bool
	}{
		{
			name:      "running",
			execution: ScheduleExecution{Status: ScheduleExecutionRunning},
			expected:  true,
		},
		{
			name:      "pending",
			execution: ScheduleExecution{Status: ScheduleExecutionPending},
			expected:  false,
		},
		{
			name:      "completed",
			execution: ScheduleExecution{Status: ScheduleExecutionCompleted},
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.execution.IsRunning())
		})
	}
}

// =============================================================================
// ToQuery Method Tests
// =============================================================================

func TestListSchedulesOptions_ToQuery(t *testing.T) {
	enabled := true
	opts := &ListSchedulesOptions{
		Page:       1,
		PageSize:   20,
		Status:     "active",
		TargetType: "job",
		Enabled:    &enabled,
		Search:     "backup",
	}

	query := opts.ToQuery()

	assert.Equal(t, "1", query["page"])
	assert.Equal(t, "20", query["page_size"])
	assert.Equal(t, "active", query["status"])
	assert.Equal(t, "job", query["target_type"])
	assert.Equal(t, "true", query["enabled"])
	assert.Equal(t, "backup", query["search"])
}

func TestListSchedulesOptions_ToQuery_Empty(t *testing.T) {
	opts := &ListSchedulesOptions{}
	query := opts.ToQuery()

	assert.Empty(t, query)
}

func TestListExecutionsOptions_ToQuery(t *testing.T) {
	opts := &ListExecutionsOptions{
		Page:     2,
		PageSize: 25,
		Status:   "failed",
	}

	query := opts.ToQuery()

	assert.Equal(t, "2", query["page"])
	assert.Equal(t, "25", query["page_size"])
	assert.Equal(t, "failed", query["status"])
}

func TestListExecutionsOptions_ToQuery_Empty(t *testing.T) {
	opts := &ListExecutionsOptions{}
	query := opts.ToQuery()

	assert.Empty(t, query)
}
