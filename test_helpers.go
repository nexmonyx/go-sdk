package nexmonyx

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// TestResponseHelper provides utilities for creating realistic mock HTTP responses
type TestResponseHelper struct {
	requestID      string
	rateLimitTotal int
	rateLimitUsed  int
}

// NewTestResponseHelper creates a new test response helper with default values
func NewTestResponseHelper() *TestResponseHelper {
	return &TestResponseHelper{
		requestID:      uuid.New().String(),
		rateLimitTotal: 1000,
		rateLimitUsed:  10,
	}
}

// WithRequestID sets a custom request ID
func (h *TestResponseHelper) WithRequestID(requestID string) *TestResponseHelper {
	h.requestID = requestID
	return h
}

// WithRateLimit sets rate limit headers
func (h *TestResponseHelper) WithRateLimit(total, used int) *TestResponseHelper {
	h.rateLimitTotal = total
	h.rateLimitUsed = used
	return h
}

// AddRealisticHeaders adds production-like headers to the response
func (h *TestResponseHelper) AddRealisticHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", h.requestID)
	w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", h.rateLimitTotal))
	w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", h.rateLimitTotal-h.rateLimitUsed))
	w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Hour).Unix()))
	w.Header().Set("X-Response-Time", "42ms")
	w.Header().Set("Server", "Nexmonyx/2.0")
}

// WriteSuccessResponse writes a success response with realistic headers
func (h *TestResponseHelper) WriteSuccessResponse(w http.ResponseWriter, data interface{}) error {
	h.AddRealisticHeaders(w)
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(StandardResponse{
		Status:  "success",
		Message: "Operation completed successfully",
		Data:    data,
	})
}

// WriteErrorResponse writes an error response with realistic headers
func (h *TestResponseHelper) WriteErrorResponse(w http.ResponseWriter, statusCode int, errorType, message, details string) error {
	h.AddRealisticHeaders(w)
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(ErrorResponse{
		Status:    "error",
		Error:     errorType,
		Message:   message,
		Details:   details,
		RequestID: h.requestID,
	})
}

// WriteValidationError writes a validation error with field-specific errors
func (h *TestResponseHelper) WriteValidationError(w http.ResponseWriter, fieldErrors map[string][]string) error {
	h.AddRealisticHeaders(w)
	w.WriteHeader(http.StatusBadRequest)
	return json.NewEncoder(w).Encode(ErrorResponse{
		Status:    "error",
		Error:     "validation_error",
		Message:   "Validation failed for one or more fields",
		RequestID: h.requestID,
		Errors:    fieldErrors,
	})
}

// WriteRateLimitError writes a rate limit exceeded error with retry-after header
func (h *TestResponseHelper) WriteRateLimitError(w http.ResponseWriter) error {
	h.AddRealisticHeaders(w)
	w.Header().Set("Retry-After", "60")
	w.WriteHeader(http.StatusTooManyRequests)
	return json.NewEncoder(w).Encode(ErrorResponse{
		Status:    "error",
		Error:     "rate_limit_exceeded",
		Message:   "API rate limit exceeded. Please retry after 60 seconds.",
		RequestID: h.requestID,
	})
}

// WritePaginatedResponse writes a paginated list response with metadata
func (h *TestResponseHelper) WritePaginatedResponse(w http.ResponseWriter, data interface{}, page, limit, totalItems int) error {
	h.AddRealisticHeaders(w)
	w.WriteHeader(http.StatusOK)

	totalPages := (totalItems + limit - 1) / limit
	hasMore := page < totalPages
	from := (page-1)*limit + 1
	to := from + limit - 1
	if to > totalItems {
		to = totalItems
	}

	var nextPage, prevPage *int
	if hasMore {
		next := page + 1
		nextPage = &next
	}
	if page > 1 {
		prev := page - 1
		prevPage = &prev
	}

	meta := &PaginationMeta{
		Page:        page,
		Limit:       limit,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		HasMore:     hasMore,
		NextPage:    nextPage,
		PrevPage:    prevPage,
		FirstPage:   1,
		LastPage:    totalPages,
		From:        from,
		To:          to,
		PerPage:     limit,
		CurrentPage: page,
	}

	return json.NewEncoder(w).Encode(PaginatedResponse{
		Status:  "success",
		Message: "Data retrieved successfully",
		Data:    data,
		Meta:    meta,
	})
}

// TestDataGenerator provides utilities for generating realistic test data
type TestDataGenerator struct {
	baseTime time.Time
}

// NewTestDataGenerator creates a new test data generator
func NewTestDataGenerator() *TestDataGenerator {
	return &TestDataGenerator{
		baseTime: time.Now().Add(-24 * time.Hour),
	}
}

// GenerateTimestamp generates a timestamp offset from the base time
func (g *TestDataGenerator) GenerateTimestamp(offsetHours int) time.Time {
	return g.baseTime.Add(time.Duration(offsetHours) * time.Hour)
}

// GenerateUUID generates a test UUID
func (g *TestDataGenerator) GenerateUUID() string {
	return uuid.New().String()
}

// GenerateServerData generates realistic server test data
func (g *TestDataGenerator) GenerateServerData(id int, name string) map[string]interface{} {
	return map[string]interface{}{
		"id":               id,
		"uuid":             g.GenerateUUID(),
		"name":             name,
		"hostname":         fmt.Sprintf("%s.example.com", name),
		"organization_id":  1,
		"status":           "active",
		"os":               "Ubuntu 22.04",
		"cpu_count":        8,
		"memory_bytes":     16000000000,
		"disk_bytes":       500000000000,
		"ip_address":       fmt.Sprintf("192.168.1.%d", id),
		"created_at":       g.GenerateTimestamp(-720).Format(time.RFC3339),
		"updated_at":       g.GenerateTimestamp(-1).Format(time.RFC3339),
		"last_seen":        g.GenerateTimestamp(0).Format(time.RFC3339),
		"agent_version":    "2.5.1",
		"monitoring_enabled": true,
	}
}

// GenerateOrganizationData generates realistic organization test data
func (g *TestDataGenerator) GenerateOrganizationData(id int, name string) map[string]interface{} {
	return map[string]interface{}{
		"id":             id,
		"uuid":           g.GenerateUUID(),
		"name":           name,
		"description":    fmt.Sprintf("Test organization %s", name),
		"status":         "active",
		"created_at":     g.GenerateTimestamp(-2160).Format(time.RFC3339),
		"updated_at":     g.GenerateTimestamp(-1).Format(time.RFC3339),
		"server_count":   25,
		"user_count":     10,
		"subscription": map[string]interface{}{
			"plan":       "professional",
			"status":     "active",
			"expires_at": g.GenerateTimestamp(720).Format(time.RFC3339),
		},
	}
}

// GenerateTaskData generates realistic task test data
func (g *TestDataGenerator) GenerateTaskData(id int, name string) map[string]interface{} {
	return map[string]interface{}{
		"id":              id,
		"name":            name,
		"description":     fmt.Sprintf("Test task: %s", name),
		"status":          "todo",
		"priority":        "medium",
		"project_id":      1,
		"feature_id":      nil,
		"sprint_id":       nil,
		"estimated_hours": 4.0,
		"actual_hours":    nil,
		"created_at":      g.GenerateTimestamp(-48).Format(time.RFC3339),
		"updated_at":      g.GenerateTimestamp(-1).Format(time.RFC3339),
		"assignee":        "user@example.com",
	}
}

// ErrorScenarios provides predefined error response scenarios
type ErrorScenarios struct {
	helper *TestResponseHelper
}

// NewErrorScenarios creates a new error scenarios helper
func NewErrorScenarios() *ErrorScenarios {
	return &ErrorScenarios{
		helper: NewTestResponseHelper(),
	}
}

// WriteUnauthorizedError writes a 401 unauthorized error
func (e *ErrorScenarios) WriteUnauthorizedError(w http.ResponseWriter) error {
	return e.helper.WriteErrorResponse(w, http.StatusUnauthorized, "unauthorized",
		"Authentication required. Please provide valid credentials.",
		"Token expired or invalid")
}

// WriteForbiddenError writes a 403 forbidden error
func (e *ErrorScenarios) WriteForbiddenError(w http.ResponseWriter) error {
	return e.helper.WriteErrorResponse(w, http.StatusForbidden, "forbidden",
		"You do not have permission to perform this action.",
		"Requires admin role")
}

// WriteNotFoundError writes a 404 not found error
func (e *ErrorScenarios) WriteNotFoundError(w http.ResponseWriter, resourceType string) error {
	return e.helper.WriteErrorResponse(w, http.StatusNotFound, "not_found",
		fmt.Sprintf("%s not found", resourceType),
		"The requested resource does not exist or has been deleted")
}

// WriteConflictError writes a 409 conflict error
func (e *ErrorScenarios) WriteConflictError(w http.ResponseWriter, message string) error {
	return e.helper.WriteErrorResponse(w, http.StatusConflict, "conflict",
		message,
		"Resource already exists with the same unique identifiers")
}

// WriteServerError writes a 500 internal server error
func (e *ErrorScenarios) WriteServerError(w http.ResponseWriter) error {
	return e.helper.WriteErrorResponse(w, http.StatusInternalServerError, "internal_error",
		"An unexpected error occurred. Please try again later.",
		"Database connection timeout")
}

// WriteServiceUnavailableError writes a 503 service unavailable error
func (e *ErrorScenarios) WriteServiceUnavailableError(w http.ResponseWriter) error {
	w.Header().Set("Retry-After", "300")
	return e.helper.WriteErrorResponse(w, http.StatusServiceUnavailable, "service_unavailable",
		"Service temporarily unavailable. Please retry after 5 minutes.",
		"System maintenance in progress")
}
