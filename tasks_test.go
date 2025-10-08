package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTasksService_CreateTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/tasks", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		var reqBody TaskConfiguration
		json.NewDecoder(r.Body).Decode(&reqBody)
		assert.Equal(t, "Generate Monthly Report", reqBody.Name)
		assert.Equal(t, "report_generation", reqBody.Type)
		assert.Equal(t, "high", reqBody.Priority)

		response := struct {
			Data    *Task  `json:"data"`
			Status  string `json:"status"`
			Message string `json:"message"`
		}{
			Data: &Task{
				ID:             1,
				OrganizationID: 100,
				Name:           reqBody.Name,
				Type:           reqBody.Type,
				Status:         "pending",
				Priority:       reqBody.Priority,
				Parameters:     reqBody.Parameters,
				Progress:       0,
				MaxRetries:     3,
				CurrentRetry:   0,
				ExecutionCount: 0,
				CreatedAt:      CustomTime{Time: time.Now()},
				UpdatedAt:      CustomTime{Time: time.Now()},
			},
			Status:  "success",
			Message: "Task created successfully",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	task, err := client.Tasks.CreateTask(context.Background(), &TaskConfiguration{
		Name:     "Generate Monthly Report",
		Type:     "report_generation",
		Priority: "high",
		Parameters: map[string]interface{}{
			"month": "January",
			"year":  2024,
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "Generate Monthly Report", task.Name)
	assert.Equal(t, "report_generation", task.Type)
	assert.Equal(t, "pending", task.Status)
	assert.Equal(t, "high", task.Priority)
}

func TestTasksService_ListTasks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/tasks", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "50", r.URL.Query().Get("limit"))
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		response := struct {
			Data []Task          `json:"data"`
			Meta *PaginationMeta `json:"meta"`
		}{
			Data: []Task{
				{
					ID:             1,
					OrganizationID: 100,
					Name:           "Daily Backup",
					Type:           "backup",
					Status:         "completed",
					Priority:       "high",
					Progress:       100,
					Schedule:       "0 2 * * *", // Daily at 2 AM
					ExecutionCount: 45,
					CreatedAt:      CustomTime{Time: time.Now()},
					UpdatedAt:      CustomTime{Time: time.Now()},
				},
				{
					ID:             2,
					OrganizationID: 100,
					Name:           "Weekly Report",
					Type:           "report_generation",
					Status:         "running",
					Priority:       "normal",
					Progress:       65,
					ExecutionCount: 12,
					CreatedAt:      CustomTime{Time: time.Now()},
					UpdatedAt:      CustomTime{Time: time.Now()},
				},
			},
			Meta: &PaginationMeta{
				Page:       1,
				PerPage:    50,
				TotalItems: 2,
				TotalPages: 1,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	tasks, meta, err := client.Tasks.ListTasks(context.Background(),
		&PaginationOptions{Page: 1, Limit: 50},
		nil)
	require.NoError(t, err)
	assert.Len(t, tasks, 2)
	assert.Equal(t, "Daily Backup", tasks[0].Name)
	assert.Equal(t, "completed", tasks[0].Status)
	assert.Equal(t, "0 2 * * *", tasks[0].Schedule)
	assert.Equal(t, "Weekly Report", tasks[1].Name)
	assert.Equal(t, "running", tasks[1].Status)
	assert.NotNil(t, meta)
	assert.Equal(t, 2, meta.TotalItems)
}

func TestTasksService_ListTasks_WithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/tasks", r.URL.Path)
		assert.Equal(t, "running", r.URL.Query().Get("status"))
		assert.Equal(t, "backup", r.URL.Query().Get("type"))
		assert.Equal(t, "high", r.URL.Query().Get("priority"))

		response := struct {
			Data []Task          `json:"data"`
			Meta *PaginationMeta `json:"meta"`
		}{
			Data: []Task{
				{
					ID:             5,
					OrganizationID: 100,
					Name:           "Critical Backup",
					Type:           "backup",
					Status:         "running",
					Priority:       "high",
					Progress:       45,
					CreatedAt:      CustomTime{Time: time.Now()},
					UpdatedAt:      CustomTime{Time: time.Now()},
				},
			},
			Meta: &PaginationMeta{
				Page:       1,
				PerPage:    20,
				TotalItems: 1,
				TotalPages: 1,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	tasks, meta, err := client.Tasks.ListTasks(context.Background(),
		nil,
		map[string]interface{}{
			"status":   "running",
			"type":     "backup",
			"priority": "high",
		})
	require.NoError(t, err)
	assert.Len(t, tasks, 1)
	assert.Equal(t, "running", tasks[0].Status)
	assert.Equal(t, "backup", tasks[0].Type)
	assert.Equal(t, "high", tasks[0].Priority)
	assert.NotNil(t, meta)
}

func TestTasksService_GetTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/tasks/123", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		scheduledAt := CustomTime{Time: time.Now().Add(1 * time.Hour)}
		startedAt := CustomTime{Time: time.Now()}
		completedAt := CustomTime{Time: time.Now().Add(5 * time.Minute)}
		nextExecution := CustomTime{Time: time.Now().Add(24 * time.Hour)}

		response := struct {
			Data    *Task  `json:"data"`
			Status  string `json:"status"`
			Message string `json:"message"`
		}{
			Data: &Task{
				ID:             123,
				OrganizationID: 100,
				Name:           "Data Cleanup",
				Type:           "cleanup",
				Status:         "completed",
				Priority:       "normal",
				Parameters: map[string]interface{}{
					"days_to_keep": 30,
					"table_name":   "audit_logs",
				},
				Result: map[string]interface{}{
					"rows_deleted": 15000,
					"duration_ms":  285000,
				},
				Progress:        100,
				Schedule:        "0 3 * * 0", // Weekly on Sunday at 3 AM
				ScheduledAt:     &scheduledAt,
				StartedAt:       &startedAt,
				CompletedAt:     &completedAt,
				ExecutionCount:  8,
				NextExecutionAt: &nextExecution,
				MaxRetries:      3,
				CurrentRetry:    0,
				TimeoutSeconds:  600,
				CreatedAt:       CustomTime{Time: time.Now()},
				UpdatedAt:       CustomTime{Time: time.Now()},
			},
			Status:  "success",
			Message: "Task retrieved successfully",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	task, err := client.Tasks.GetTask(context.Background(), 123)
	require.NoError(t, err)
	assert.Equal(t, uint(123), task.ID)
	assert.Equal(t, "Data Cleanup", task.Name)
	assert.Equal(t, "cleanup", task.Type)
	assert.Equal(t, "completed", task.Status)
	assert.Equal(t, 100, task.Progress)
	assert.Equal(t, "0 3 * * 0", task.Schedule)
	assert.NotNil(t, task.Parameters)
	assert.Equal(t, float64(30), task.Parameters["days_to_keep"])
	assert.NotNil(t, task.Result)
	assert.Equal(t, float64(15000), task.Result["rows_deleted"])
}

func TestTasksService_UpdateTaskStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/tasks/456/status", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		var reqBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&reqBody)
		assert.Equal(t, "completed", reqBody["status"])
		assert.NotNil(t, reqBody["result"])

		completedAt := CustomTime{Time: time.Now()}
		response := struct {
			Data    *Task  `json:"data"`
			Status  string `json:"status"`
			Message string `json:"message"`
		}{
			Data: &Task{
				ID:             456,
				OrganizationID: 100,
				Name:           "Export Report",
				Type:           "data_export",
				Status:         "completed",
				Priority:       "normal",
				Progress:       100,
				Result: map[string]interface{}{
					"file_size": 2048576,
					"rows":      5000,
				},
				CompletedAt: &completedAt,
				CreatedAt:   CustomTime{Time: time.Now()},
				UpdatedAt:   CustomTime{Time: time.Now()},
			},
			Status:  "success",
			Message: "Task status updated successfully",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	task, err := client.Tasks.UpdateTaskStatus(context.Background(),
		456,
		"completed",
		map[string]interface{}{
			"file_size": 2048576,
			"rows":      5000,
		})
	require.NoError(t, err)
	assert.Equal(t, "completed", task.Status)
	assert.Equal(t, 100, task.Progress)
	assert.NotNil(t, task.Result)
	assert.Equal(t, float64(2048576), task.Result["file_size"])
}

func TestTasksService_CancelTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/tasks/789/cancel", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		response := struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		}{
			Status:  "success",
			Message: "Task cancelled successfully",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    AuthConfig{Token: "test-token"},
	})
	require.NoError(t, err)

	err = client.Tasks.CancelTask(context.Background(), 789)
	require.NoError(t, err)
}

func TestTasksService_ErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		expectedError bool
	}{
		{
			name:          "Unauthorized",
			statusCode:    http.StatusUnauthorized,
			expectedError: true,
		},
		{
			name:          "Forbidden",
			statusCode:    http.StatusForbidden,
			expectedError: true,
		},
		{
			name:          "Not Found",
			statusCode:    http.StatusNotFound,
			expectedError: true,
		},
		{
			name:          "Internal Server Error",
			statusCode:    http.StatusInternalServerError,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(StandardResponse{
					Status:  "error",
					Message: "Error occurred",
				})
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			_, _, err = client.Tasks.ListTasks(context.Background(), nil, nil)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
