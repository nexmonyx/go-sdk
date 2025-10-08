package nexmonyx

import (
	"context"
	"fmt"
)

// TasksService handles task management, scheduling, and workflow automation
type TasksService struct {
	client *Client
}

// CreateTask creates a new task with specified configuration
// Authentication: JWT Token required
// Endpoint: POST /v1/tasks
// Parameters:
//   - config: Task configuration including type, parameters, and scheduling
// Returns: Created Task object
func (s *TasksService) CreateTask(ctx context.Context, config *TaskConfiguration) (*Task, error) {
	var resp struct {
		Data    *Task  `json:"data"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/tasks",
		Body:   config,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// ListTasks retrieves a list of tasks with filtering and pagination
// Authentication: JWT Token required
// Endpoint: GET /v1/tasks
// Parameters:
//   - opts: Optional pagination options
//   - filters: Optional filters (status, type, priority, scheduled_after, scheduled_before)
// Returns: Array of Task objects with pagination metadata
func (s *TasksService) ListTasks(ctx context.Context, opts *PaginationOptions, filters map[string]interface{}) ([]Task, *PaginationMeta, error) {
	var resp struct {
		Data []Task          `json:"data"`
		Meta *PaginationMeta `json:"meta"`
	}

	queryParams := make(map[string]string)
	if opts != nil {
		if opts.Page > 0 {
			queryParams["page"] = fmt.Sprintf("%d", opts.Page)
		}
		if opts.Limit > 0 {
			queryParams["limit"] = fmt.Sprintf("%d", opts.Limit)
		}
	}

	// Add filter parameters
	if filters != nil {
		if status, ok := filters["status"].(string); ok && status != "" {
			queryParams["status"] = status
		}
		if taskType, ok := filters["type"].(string); ok && taskType != "" {
			queryParams["type"] = taskType
		}
		if priority, ok := filters["priority"].(string); ok && priority != "" {
			queryParams["priority"] = priority
		}
		if scheduledAfter, ok := filters["scheduled_after"].(string); ok && scheduledAfter != "" {
			queryParams["scheduled_after"] = scheduledAfter
		}
		if scheduledBefore, ok := filters["scheduled_before"].(string); ok && scheduledBefore != "" {
			queryParams["scheduled_before"] = scheduledBefore
		}
	}

	req := &Request{
		Method: "GET",
		Path:   "/v1/tasks",
		Result: &resp,
	}
	if len(queryParams) > 0 {
		req.Query = queryParams
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return resp.Data, resp.Meta, nil
}

// GetTask retrieves a specific task by ID
// Authentication: JWT Token required
// Endpoint: GET /v1/tasks/{id}
// Parameters:
//   - taskID: Task ID
// Returns: Task object with full details including execution history
func (s *TasksService) GetTask(ctx context.Context, taskID uint) (*Task, error) {
	var resp struct {
		Data    *Task  `json:"data"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/tasks/%d", taskID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// UpdateTaskStatus updates the status of a task
// Authentication: JWT Token required
// Endpoint: PUT /v1/tasks/{id}/status
// Parameters:
//   - taskID: Task ID
//   - status: New status (pending, running, completed, failed, cancelled)
//   - result: Optional result data for completed/failed tasks
// Returns: Updated Task object
func (s *TasksService) UpdateTaskStatus(ctx context.Context, taskID uint, status string, result map[string]interface{}) (*Task, error) {
	var resp struct {
		Data    *Task  `json:"data"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	body := map[string]interface{}{
		"status": status,
	}
	if result != nil {
		body["result"] = result
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/tasks/%d/status", taskID),
		Body:   body,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// CancelTask cancels a pending or running task
// Authentication: JWT Token required
// Endpoint: POST /v1/tasks/{id}/cancel
// Parameters:
//   - taskID: Task ID
// Returns: Error if cancellation fails
func (s *TasksService) CancelTask(ctx context.Context, taskID uint) error {
	var resp struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/tasks/%d/cancel", taskID),
		Result: &resp,
	})
	if err != nil {
		return err
	}

	return nil
}
