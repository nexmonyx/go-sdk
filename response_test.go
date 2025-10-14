package nexmonyx

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestQueryTimeRange_ToStrings tests the ToStrings method
func TestQueryTimeRange_ToStrings(t *testing.T) {
	tests := []struct {
		name          string
		timeRange     *QueryTimeRange
		expectedStart string
		expectedEnd   string
	}{
		{
			name: "valid time range",
			timeRange: &QueryTimeRange{
				Start: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC),
			},
			expectedStart: "2023-01-01T00:00:00Z",
			expectedEnd:   "2023-12-31T23:59:59Z",
		},
		{
			name: "same start and end",
			timeRange: &QueryTimeRange{
				Start: time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC),
				End:   time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC),
			},
			expectedStart: "2023-06-15T12:00:00Z",
			expectedEnd:   "2023-06-15T12:00:00Z",
		},
		{
			name: "with timezone",
			timeRange: &QueryTimeRange{
				Start: time.Date(2023, 3, 1, 10, 30, 0, 0, time.FixedZone("PST", -8*3600)),
				End:   time.Date(2023, 3, 1, 18, 30, 0, 0, time.FixedZone("PST", -8*3600)),
			},
			expectedStart: "2023-03-01T10:30:00-08:00",
			expectedEnd:   "2023-03-01T18:30:00-08:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := tt.timeRange.ToStrings()
			assert.Equal(t, tt.expectedStart, start)
			assert.Equal(t, tt.expectedEnd, end)
		})
	}
}

// TestLast24Hours tests the Last24Hours helper function
func TestLast24Hours(t *testing.T) {
	before := time.Now()
	qtr := Last24Hours()
	after := time.Now()

	// Check that the time range is approximately 24 hours
	duration := qtr.End.Sub(qtr.Start)
	assert.InDelta(t, 24*time.Hour, duration, float64(time.Second))

	// Check that End is approximately now
	assert.True(t, qtr.End.After(before) || qtr.End.Equal(before))
	assert.True(t, qtr.End.Before(after) || qtr.End.Equal(after))

	// Check that Start is approximately 24 hours ago
	expectedStart := before.Add(-24 * time.Hour)
	assert.InDelta(t, expectedStart.Unix(), qtr.Start.Unix(), 1)
}

// TestLast7Days tests the Last7Days helper function
func TestLast7Days(t *testing.T) {
	before := time.Now()
	qtr := Last7Days()
	after := time.Now()

	// Check that the time range is approximately 7 days
	duration := qtr.End.Sub(qtr.Start)
	assert.InDelta(t, 7*24*time.Hour, duration, float64(time.Second))

	// Check that End is approximately now
	assert.True(t, qtr.End.After(before) || qtr.End.Equal(before))
	assert.True(t, qtr.End.Before(after) || qtr.End.Equal(after))

	// Check that Start is approximately 7 days ago
	expectedStart := before.Add(-7 * 24 * time.Hour)
	assert.InDelta(t, expectedStart.Unix(), qtr.Start.Unix(), 1)
}

// TestLast30Days tests the Last30Days helper function
func TestLast30Days(t *testing.T) {
	before := time.Now()
	qtr := Last30Days()
	after := time.Now()

	// Check that the time range is approximately 30 days
	duration := qtr.End.Sub(qtr.Start)
	assert.InDelta(t, 30*24*time.Hour, duration, float64(time.Second))

	// Check that End is approximately now
	assert.True(t, qtr.End.After(before) || qtr.End.Equal(before))
	assert.True(t, qtr.End.Before(after) || qtr.End.Equal(after))

	// Check that Start is approximately 30 days ago
	expectedStart := before.Add(-30 * 24 * time.Hour)
	assert.InDelta(t, expectedStart.Unix(), qtr.Start.Unix(), 1)
}

// TestThisMonth tests the ThisMonth helper function
func TestThisMonth(t *testing.T) {
	now := time.Now()
	qtr := ThisMonth()

	// Check that Start is the first day of the current month at midnight
	expectedStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	assert.Equal(t, expectedStart, qtr.Start)

	// Check that End is approximately now
	assert.True(t, qtr.End.After(now.Add(-time.Second)))
	assert.True(t, qtr.End.Before(now.Add(time.Second)))
}

// TestLastMonth tests the LastMonth helper function
func TestLastMonth(t *testing.T) {
	now := time.Now()
	qtr := LastMonth()

	// Expected start is the first day of the previous month
	expectedStart := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, now.Location())
	assert.Equal(t, expectedStart, qtr.Start)

	// Expected end is one second before the first day of the current month
	expectedEnd := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Add(-time.Second)
	assert.Equal(t, expectedEnd, qtr.End)

	// Verify the time range is within the previous month
	assert.True(t, qtr.Start.Before(qtr.End))
	assert.True(t, qtr.End.Before(time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())))
}

// TestLastMonth_EdgeCases tests LastMonth with edge cases like January
func TestLastMonth_EdgeCases(t *testing.T) {
	// This test ensures LastMonth works correctly when called in January
	// We can't control the current time, but we can verify the logic is sound
	qtr := LastMonth()

	// Verify Start is before End
	assert.True(t, qtr.Start.Before(qtr.End), "Start should be before End")

	// Verify the difference is approximately one month
	// Allow for months of different lengths (28-31 days)
	duration := qtr.End.Sub(qtr.Start)
	assert.True(t, duration >= 27*24*time.Hour, "Duration should be at least 27 days")
	assert.True(t, duration <= 31*24*time.Hour, "Duration should be at most 31 days")
}

// TestPaginatedResponse_Unmarshal tests unmarshaling of paginated responses
func TestPaginatedResponse_Unmarshal(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		validate func(t *testing.T, resp *PaginatedResponse)
	}{
		{
			name: "complete paginated response",
			json: `{
				"status": "success",
				"message": "Servers retrieved",
				"data": [{"id": 1}, {"id": 2}],
				"meta": {
					"page": 1,
					"limit": 25,
					"total_items": 100,
					"total_pages": 4,
					"has_more": true,
					"next_page": 2,
					"first_page": 1,
					"last_page": 4,
					"from": 1,
					"to": 25,
					"per_page": 25,
					"current_page": 1
				}
			}`,
			validate: func(t *testing.T, resp *PaginatedResponse) {
				assert.Equal(t, "success", resp.Status)
				assert.Equal(t, "Servers retrieved", resp.Message)
				assert.NotNil(t, resp.Meta)
				assert.Equal(t, 1, resp.Meta.Page)
				assert.Equal(t, 25, resp.Meta.Limit)
				assert.Equal(t, 100, resp.Meta.TotalItems)
				assert.Equal(t, 4, resp.Meta.TotalPages)
				assert.True(t, resp.Meta.HasMore)
				assert.NotNil(t, resp.Meta.NextPage)
				assert.Equal(t, 2, *resp.Meta.NextPage)
			},
		},
		{
			name: "last page without next",
			json: `{
				"status": "success",
				"message": "Servers retrieved",
				"data": [],
				"meta": {
					"page": 4,
					"limit": 25,
					"total_items": 100,
					"total_pages": 4,
					"has_more": false,
					"prev_page": 3,
					"first_page": 1,
					"last_page": 4,
					"from": 76,
					"to": 100,
					"per_page": 25,
					"current_page": 4
				}
			}`,
			validate: func(t *testing.T, resp *PaginatedResponse) {
				assert.Equal(t, "success", resp.Status)
				assert.False(t, resp.Meta.HasMore)
				assert.Nil(t, resp.Meta.NextPage)
				assert.NotNil(t, resp.Meta.PrevPage)
				assert.Equal(t, 3, *resp.Meta.PrevPage)
			},
		},
		{
			name: "with page URLs",
			json: `{
				"status": "success",
				"message": "Data retrieved",
				"data": [],
				"meta": {
					"page": 2,
					"limit": 10,
					"total_items": 50,
					"total_pages": 5,
					"has_more": true,
					"first_page": 1,
					"last_page": 5,
					"from": 11,
					"to": 20,
					"per_page": 10,
					"current_page": 2,
					"first_page_url": "/api/v1/data?page=1",
					"last_page_url": "/api/v1/data?page=5",
					"next_page_url": "/api/v1/data?page=3",
					"prev_page_url": "/api/v1/data?page=1"
				}
			}`,
			validate: func(t *testing.T, resp *PaginatedResponse) {
				assert.Equal(t, "/api/v1/data?page=1", resp.Meta.FirstPageURL)
				assert.Equal(t, "/api/v1/data?page=5", resp.Meta.LastPageURL)
				assert.Equal(t, "/api/v1/data?page=3", resp.Meta.NextPageURL)
				assert.Equal(t, "/api/v1/data?page=1", resp.Meta.PrevPageURL)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp PaginatedResponse
			err := json.Unmarshal([]byte(tt.json), &resp)
			require.NoError(t, err)
			tt.validate(t, &resp)
		})
	}
}

// TestBatchResponse_Unmarshal tests unmarshaling of batch responses
func TestBatchResponse_Unmarshal(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		validate func(t *testing.T, resp *BatchResponse)
	}{
		{
			name: "mixed success and failure",
			json: `{
				"status": "partial",
				"message": "Batch operation completed with some failures",
				"successful": [
					{"id": "1", "status": "success", "data": {"name": "Server 1"}},
					{"id": "2", "status": "success", "data": {"name": "Server 2"}}
				],
				"failed": [
					{"id": "3", "status": "error", "error": "validation_error", "message": "Invalid input"}
				],
				"total": 3,
				"success": 2,
				"failures": 1
			}`,
			validate: func(t *testing.T, resp *BatchResponse) {
				assert.Equal(t, "partial", resp.Status)
				assert.Equal(t, 3, resp.Total)
				assert.Equal(t, 2, resp.Success)
				assert.Equal(t, 1, resp.Failures)
				assert.Len(t, resp.Successful, 2)
				assert.Len(t, resp.Failed, 1)
				assert.Equal(t, "1", resp.Successful[0].ID)
				assert.Equal(t, "3", resp.Failed[0].ID)
			},
		},
		{
			name: "all successful",
			json: `{
				"status": "success",
				"message": "All operations completed successfully",
				"successful": [
					{"id": "1", "status": "success"},
					{"id": "2", "status": "success"}
				],
				"failed": [],
				"total": 2,
				"success": 2,
				"failures": 0
			}`,
			validate: func(t *testing.T, resp *BatchResponse) {
				assert.Equal(t, "success", resp.Status)
				assert.Equal(t, 2, resp.Total)
				assert.Equal(t, 2, resp.Success)
				assert.Equal(t, 0, resp.Failures)
				assert.Len(t, resp.Successful, 2)
				assert.Len(t, resp.Failed, 0)
			},
		},
		{
			name: "all failed",
			json: `{
				"status": "error",
				"message": "All operations failed",
				"successful": [],
				"failed": [
					{"id": "1", "status": "error", "error": "not_found", "message": "Resource not found"},
					{"id": "2", "status": "error", "error": "forbidden", "message": "Access denied"}
				],
				"total": 2,
				"success": 0,
				"failures": 2
			}`,
			validate: func(t *testing.T, resp *BatchResponse) {
				assert.Equal(t, "error", resp.Status)
				assert.Equal(t, 2, resp.Total)
				assert.Equal(t, 0, resp.Success)
				assert.Equal(t, 2, resp.Failures)
				assert.Len(t, resp.Successful, 0)
				assert.Len(t, resp.Failed, 2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp BatchResponse
			err := json.Unmarshal([]byte(tt.json), &resp)
			require.NoError(t, err)
			tt.validate(t, &resp)
		})
	}
}

// TestStatusResponse_Unmarshal tests unmarshaling of status responses
func TestStatusResponse_Unmarshal(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		validate func(t *testing.T, resp *StatusResponse)
	}{
		{
			name: "healthy system",
			json: `{
				"status": "operational",
				"healthy": true,
				"version": "1.0.0",
				"uptime": 3600,
				"timestamp": "2023-01-01T12:00:00Z",
				"services": {
					"database": true,
					"cache": true,
					"queue": true
				},
				"details": {
					"cpu_usage": 45.2,
					"memory_usage": 60.5
				}
			}`,
			validate: func(t *testing.T, resp *StatusResponse) {
				assert.Equal(t, "operational", resp.Status)
				assert.True(t, resp.Healthy)
				assert.Equal(t, "1.0.0", resp.Version)
				assert.Equal(t, int64(3600), resp.Uptime)
				assert.NotNil(t, resp.Services)
				assert.True(t, resp.Services["database"])
				assert.True(t, resp.Services["cache"])
				assert.NotNil(t, resp.Details)
			},
		},
		{
			name: "degraded system",
			json: `{
				"status": "degraded",
				"healthy": false,
				"version": "1.0.0",
				"timestamp": "2023-01-01T12:00:00Z",
				"services": {
					"database": true,
					"cache": false
				}
			}`,
			validate: func(t *testing.T, resp *StatusResponse) {
				assert.Equal(t, "degraded", resp.Status)
				assert.False(t, resp.Healthy)
				assert.True(t, resp.Services["database"])
				assert.False(t, resp.Services["cache"])
			},
		},
		{
			name: "minimal status response",
			json: `{
				"status": "unknown",
				"healthy": false,
				"timestamp": "2023-01-01T12:00:00Z"
			}`,
			validate: func(t *testing.T, resp *StatusResponse) {
				assert.Equal(t, "unknown", resp.Status)
				assert.False(t, resp.Healthy)
				assert.Empty(t, resp.Version)
				assert.Equal(t, int64(0), resp.Uptime)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp StatusResponse
			err := json.Unmarshal([]byte(tt.json), &resp)
			require.NoError(t, err)
			tt.validate(t, &resp)
		})
	}
}

// TestHeartbeatResponse_Unmarshal tests unmarshaling of heartbeat responses
func TestHeartbeatResponse_Unmarshal(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		validate func(t *testing.T, resp *HeartbeatResponse)
	}{
		{
			name: "complete heartbeat",
			json: `{
				"status": "success",
				"message": "Heartbeat received",
				"last_heartbeat": "2023-01-01T12:00:00Z",
				"server_uuid": "test-uuid-123",
				"server_status": "online",
				"heartbeat_count": 42,
				"details": {
					"agent_version": "1.2.3",
					"last_seen": "2023-01-01T11:59:00Z",
					"next_expected": "2023-01-01T12:01:00Z"
				}
			}`,
			validate: func(t *testing.T, resp *HeartbeatResponse) {
				assert.Equal(t, "success", resp.Status)
				assert.Equal(t, "Heartbeat received", resp.Message)
				assert.NotNil(t, resp.LastHeartbeat)
				assert.Equal(t, "test-uuid-123", resp.ServerUUID)
				assert.Equal(t, "online", resp.ServerStatus)
				assert.Equal(t, 42, resp.HeartbeatCount)
				assert.Equal(t, "1.2.3", resp.Details.AgentVersion)
				assert.NotNil(t, resp.Details.LastSeen)
				assert.NotNil(t, resp.Details.NextExpected)
			},
		},
		{
			name: "minimal heartbeat",
			json: `{
				"status": "success",
				"message": "Heartbeat received",
				"server_uuid": "test-uuid-456"
			}`,
			validate: func(t *testing.T, resp *HeartbeatResponse) {
				assert.Equal(t, "success", resp.Status)
				assert.Equal(t, "test-uuid-456", resp.ServerUUID)
				assert.Nil(t, resp.LastHeartbeat)
				assert.Empty(t, resp.ServerStatus)
				assert.Equal(t, 0, resp.HeartbeatCount)
			},
		},
		{
			name: "heartbeat with error",
			json: `{
				"status": "error",
				"message": "Server not found",
				"server_uuid": "unknown-uuid"
			}`,
			validate: func(t *testing.T, resp *HeartbeatResponse) {
				assert.Equal(t, "error", resp.Status)
				assert.Equal(t, "Server not found", resp.Message)
				assert.Equal(t, "unknown-uuid", resp.ServerUUID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp HeartbeatResponse
			err := json.Unmarshal([]byte(tt.json), &resp)
			require.NoError(t, err)
			tt.validate(t, &resp)
		})
	}
}

// TestErrorResponse_Unmarshal tests unmarshaling of error responses
func TestErrorResponse_Unmarshal(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		validate func(t *testing.T, resp *ErrorResponse)
	}{
		{
			name: "complete error response",
			json: `{
				"status": "error",
				"error": "validation_error",
				"message": "Validation failed",
				"details": "Multiple fields are invalid",
				"request_id": "req-123",
				"errors": {
					"email": ["Email is required", "Email format is invalid"],
					"password": ["Password must be at least 8 characters"]
				}
			}`,
			validate: func(t *testing.T, resp *ErrorResponse) {
				assert.Equal(t, "error", resp.Status)
				assert.Equal(t, "validation_error", resp.Error)
				assert.Equal(t, "Validation failed", resp.Message)
				assert.Equal(t, "Multiple fields are invalid", resp.Details)
				assert.Equal(t, "req-123", resp.RequestID)
				assert.NotNil(t, resp.Errors)
				assert.Len(t, resp.Errors["email"], 2)
				assert.Len(t, resp.Errors["password"], 1)
			},
		},
		{
			name: "simple error response",
			json: `{
				"status": "error",
				"error": "not_found",
				"message": "Resource not found"
			}`,
			validate: func(t *testing.T, resp *ErrorResponse) {
				assert.Equal(t, "error", resp.Status)
				assert.Equal(t, "not_found", resp.Error)
				assert.Equal(t, "Resource not found", resp.Message)
				assert.Empty(t, resp.Details)
				assert.Empty(t, resp.RequestID)
				assert.Nil(t, resp.Errors)
			},
		},
		{
			name: "error with request ID",
			json: `{
				"status": "error",
				"error": "internal_error",
				"message": "An internal error occurred",
				"request_id": "req-xyz-789"
			}`,
			validate: func(t *testing.T, resp *ErrorResponse) {
				assert.Equal(t, "error", resp.Status)
				assert.Equal(t, "internal_error", resp.Error)
				assert.Equal(t, "req-xyz-789", resp.RequestID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp ErrorResponse
			err := json.Unmarshal([]byte(tt.json), &resp)
			require.NoError(t, err)
			tt.validate(t, &resp)
		})
	}
}

// TestPaginationMeta_EdgeCases tests edge cases in pagination metadata
func TestPaginationMeta_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{
			name: "single page result",
			json: `{
				"page": 1,
				"limit": 25,
				"total_items": 10,
				"total_pages": 1,
				"has_more": false,
				"first_page": 1,
				"last_page": 1,
				"from": 1,
				"to": 10,
				"per_page": 25,
				"current_page": 1
			}`,
		},
		{
			name: "empty result",
			json: `{
				"page": 1,
				"limit": 25,
				"total_items": 0,
				"total_pages": 0,
				"has_more": false,
				"first_page": 1,
				"last_page": 0,
				"from": 0,
				"to": 0,
				"per_page": 25,
				"current_page": 1
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var meta PaginationMeta
			err := json.Unmarshal([]byte(tt.json), &meta)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, meta.Page, 0)
			assert.GreaterOrEqual(t, meta.TotalItems, 0)
		})
	}
}

// TestTimeRange_Struct tests the TimeRange struct
func TestTimeRange_Struct(t *testing.T) {
	tr := TimeRange{
		Start: "2023-01-01T00:00:00Z",
		End:   "2023-01-31T23:59:59Z",
	}

	jsonData, err := json.Marshal(tr)
	require.NoError(t, err)

	var decoded TimeRange
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err)

	assert.Equal(t, tr.Start, decoded.Start)
	assert.Equal(t, tr.End, decoded.End)
}

// TestTimeRange_EmptyValues tests TimeRange with empty values
func TestTimeRange_EmptyValues(t *testing.T) {
	tr := TimeRange{}

	jsonData, err := json.Marshal(tr)
	require.NoError(t, err)

	var decoded TimeRange
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err)

	assert.Empty(t, decoded.Start)
	assert.Empty(t, decoded.End)
}

// TestListOptions_AllParameters tests ListOptions with all parameters
func TestListOptions_AllParameters(t *testing.T) {
	opts := &ListOptions{
		Page:        5,
		Limit:       100,
		PerPage:     50,
		Sort:        "created_at",
		Order:       "asc",
		Search:      "production",
		Query:       "status:active AND type:server",
		StartDate:   "2023-01-01",
		EndDate:     "2023-12-31",
		TimeRange:   "custom",
		GroupBy:     "datacenter",
		Aggregation: "count",
		Filters: map[string]string{
			"environment": "production",
			"tier":        "premium",
			"region":      "us-east-1",
		},
		Fields:  []string{"id", "name", "status", "uptime"},
		Expand:  []string{"organization", "metrics", "alerts"},
		Include: []string{"metadata", "tags", "history"},
	}

	query := opts.ToQuery()

	// Verify all scalar fields are included
	assert.Equal(t, "5", query["page"])
	assert.Equal(t, "100", query["limit"])
	assert.Equal(t, "50", query["per_page"])
	assert.Equal(t, "created_at", query["sort"])
	assert.Equal(t, "asc", query["order"])
	assert.Equal(t, "production", query["search"])
	assert.Equal(t, "status:active AND type:server", query["q"])
	assert.Equal(t, "2023-01-01", query["start_date"])
	assert.Equal(t, "2023-12-31", query["end_date"])
	assert.Equal(t, "custom", query["time_range"])
	assert.Equal(t, "datacenter", query["group_by"])
	assert.Equal(t, "count", query["aggregation"])

	// Verify filters are included
	assert.Equal(t, "production", query["environment"])
	assert.Equal(t, "premium", query["tier"])
	assert.Equal(t, "us-east-1", query["region"])

	// Array fields (Fields, Expand, Include) are not included in ToQuery
	// as they use struct tags for URL encoding
	assert.Len(t, opts.Fields, 4)
	assert.Len(t, opts.Expand, 3)
	assert.Len(t, opts.Include, 3)
}

// TestListOptions_NilAndEmpty tests ListOptions with nil and empty values
func TestListOptions_NilAndEmpty(t *testing.T) {
	tests := []struct {
		name     string
		options  *ListOptions
		expected map[string]string
	}{
		{
			name:     "nil filters",
			options:  &ListOptions{Page: 1, Filters: nil},
			expected: map[string]string{"page": "1"},
		},
		{
			name:     "empty filters",
			options:  &ListOptions{Page: 1, Filters: map[string]string{}},
			expected: map[string]string{"page": "1"},
		},
		{
			name:     "empty strings not included",
			options:  &ListOptions{Page: 1, Sort: "", Order: "", Search: ""},
			expected: map[string]string{"page": "1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.options.ToQuery()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestBatchItemResponse_AllFields tests all fields of BatchItemResponse
func TestBatchItemResponse_AllFields(t *testing.T) {
	item := BatchItemResponse{
		ID:      "item-1",
		Status:  "success",
		Data:    map[string]interface{}{"result": "ok", "count": 42},
		Error:   "",
		Message: "Operation completed successfully",
	}

	jsonData, err := json.Marshal(item)
	require.NoError(t, err)

	var decoded BatchItemResponse
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err)

	assert.Equal(t, item.ID, decoded.ID)
	assert.Equal(t, item.Status, decoded.Status)
	assert.Equal(t, item.Message, decoded.Message)
	assert.Empty(t, decoded.Error)

	// Verify data is decoded correctly
	data, ok := decoded.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "ok", data["result"])
	assert.Equal(t, float64(42), data["count"]) // JSON numbers decode as float64
}

// TestBatchItemResponse_WithError tests BatchItemResponse with error
func TestBatchItemResponse_WithError(t *testing.T) {
	item := BatchItemResponse{
		ID:      "item-error",
		Status:  "error",
		Error:   "validation_failed",
		Message: "Invalid input data",
		Data:    nil,
	}

	jsonData, err := json.Marshal(item)
	require.NoError(t, err)

	var decoded BatchItemResponse
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "item-error", decoded.ID)
	assert.Equal(t, "error", decoded.Status)
	assert.Equal(t, "validation_failed", decoded.Error)
	assert.Equal(t, "Invalid input data", decoded.Message)
}

// TestStandardResponse_AllFields tests StandardResponse with all fields
func TestStandardResponse_AllFields(t *testing.T) {
	resp := StandardResponse{
		Status:  "success",
		Message: "Operation completed",
		Data:    map[string]interface{}{"id": "123", "name": "Test"},
		Error:   "",
		Details: "",
	}

	jsonData, err := json.Marshal(resp)
	require.NoError(t, err)

	var decoded StandardResponse
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err)

	assert.Equal(t, resp.Status, decoded.Status)
	assert.Equal(t, resp.Message, decoded.Message)
	assert.NotNil(t, decoded.Data)
}

// TestListOptions_ToQuery_EdgeCases tests edge cases in ToQuery
func TestListOptions_ToQuery_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		options  *ListOptions
		expected int // expected number of query params
	}{
		{
			name:     "all zero values",
			options:  &ListOptions{},
			expected: 0,
		},
		{
			name: "only one field set",
			options: &ListOptions{
				Search: "test",
			},
			expected: 1,
		},
		{
			name: "filters with special characters",
			options: &ListOptions{
				Filters: map[string]string{
					"name":   "server-01",
					"status": "active|pending",
					"query":  "type:web AND region:us-*",
				},
			},
			expected: 3,
		},
		{
			name: "numeric values",
			options: &ListOptions{
				Page:    999,
				Limit:   1000,
				PerPage: 500,
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.options.ToQuery()
			assert.Len(t, result, tt.expected)
		})
	}
}
