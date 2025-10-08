package nexmonyx

import (
	"fmt"
)

// APIError represents an error response from the Nexmonyx API
type APIError struct {
	Status    string `json:"status"`
	ErrorType string `json:"error"`
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
	Details   string `json:"details,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.ErrorCode, e.Message, e.Details)
	}
	if e.ErrorCode != "" {
		return fmt.Sprintf("%s: %s", e.ErrorCode, e.Message)
	}
	return e.Message
}

// RateLimitError represents a rate limit error
type RateLimitError struct {
	RetryAfter string
	Message    string
	Limit      int
	Remaining  int
	Reset      int64
}

// Error implements the error interface
func (e *RateLimitError) Error() string {
	if e.RetryAfter != "" {
		return fmt.Sprintf("%s (retry after: %s)", e.Message, e.RetryAfter)
	}
	return e.Message
}

// ValidationError represents a validation error
type ValidationError struct {
	StatusCode int
	Message    string
	Errors     map[string][]string `json:"errors,omitempty"`
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	if len(e.Errors) > 0 {
		return fmt.Sprintf("validation error: %s", e.Message)
	}
	return e.Message
}

// NotFoundError represents a 404 error
type NotFoundError struct {
	Resource string
	ID       string
	Message  string
}

// Error implements the error interface
func (e *NotFoundError) Error() string {
	if e.Resource != "" && e.ID != "" {
		return fmt.Sprintf("%s with ID %s not found", e.Resource, e.ID)
	}
	if e.Message != "" {
		return e.Message
	}
	return "resource not found"
}

// UnauthorizedError represents a 401 error
type UnauthorizedError struct {
	Message string
}

// Error implements the error interface
func (e *UnauthorizedError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "unauthorized"
}

// ForbiddenError represents a 403 error
type ForbiddenError struct {
	Resource string
	Action   string
	Message  string
}

// Error implements the error interface
func (e *ForbiddenError) Error() string {
	if e.Resource != "" && e.Action != "" {
		return fmt.Sprintf("forbidden: cannot %s %s", e.Action, e.Resource)
	}
	if e.Message != "" {
		return e.Message
	}
	return "forbidden"
}

// InternalServerError represents a 500 error
type InternalServerError struct {
	StatusCode int
	Message    string
	RequestID  string
}

// Error implements the error interface
func (e *InternalServerError) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("%s (request ID: %s)", e.Message, e.RequestID)
	}
	return e.Message
}

// ConflictError represents a 409 error
type ConflictError struct {
	Resource string
	Message  string
}

// Error implements the error interface
func (e *ConflictError) Error() string {
	if e.Resource != "" {
		return fmt.Sprintf("conflict: %s already exists", e.Resource)
	}
	return e.Message
}

// ServiceUnavailableError represents a 503 error
type ServiceUnavailableError struct {
	Message   string
	RetryTime int
}

// Error implements the error interface
func (e *ServiceUnavailableError) Error() string {
	if e.RetryTime > 0 {
		return fmt.Sprintf("%s (retry in %d seconds)", e.Message, e.RetryTime)
	}
	return e.Message
}

// IsNotFound returns true if the error is a NotFoundError
func IsNotFound(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}

// IsRateLimit returns true if the error is a RateLimitError
func IsRateLimit(err error) bool {
	_, ok := err.(*RateLimitError)
	return ok
}

// IsUnauthorized returns true if the error is an UnauthorizedError
func IsUnauthorized(err error) bool {
	_, ok := err.(*UnauthorizedError)
	return ok
}

// IsForbidden returns true if the error is a ForbiddenError
func IsForbidden(err error) bool {
	_, ok := err.(*ForbiddenError)
	return ok
}

// IsValidation returns true if the error is a ValidationError
func IsValidation(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

// IsConflict returns true if the error is a ConflictError
func IsConflict(err error) bool {
	_, ok := err.(*ConflictError)
	return ok
}

// IsServerError returns true if the error is a server error (5xx)
func IsServerError(err error) bool {
	_, ok := err.(*InternalServerError)
	if ok {
		return true
	}
	_, ok = err.(*ServiceUnavailableError)
	return ok
}

// Common error variables
var (
	// ErrUnexpectedResponse is returned when the API returns an unexpected response format
	ErrUnexpectedResponse = fmt.Errorf("unexpected response format from API")
)
