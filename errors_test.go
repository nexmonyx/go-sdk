package nexmonyx

import (
	"encoding/json"
	"errors"
	"testing"
)

// TestAPIError tests the APIError type and its Error() method
func TestAPIError(t *testing.T) {
	tests := []struct {
		name     string
		err      *APIError
		expected string
	}{
		{
			name: "full error with details",
			err: &APIError{
				Status:    "error",
				ErrorType: "validation_error",
				ErrorCode: "INVALID_INPUT",
				Message:   "Invalid input provided",
				Details:   "Field 'email' is required",
				RequestID: "req-123",
			},
			expected: "INVALID_INPUT: Invalid input provided (Field 'email' is required)",
		},
		{
			name: "error with code but no details",
			err: &APIError{
				ErrorCode: "NOT_FOUND",
				Message:   "Resource not found",
			},
			expected: "NOT_FOUND: Resource not found",
		},
		{
			name: "error with message only",
			err: &APIError{
				Message: "Something went wrong",
			},
			expected: "Something went wrong",
		},
		{
			name: "empty error",
			err: &APIError{
				Message: "",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("APIError.Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestAPIErrorJSON tests JSON marshaling/unmarshaling of APIError
func TestAPIErrorJSON(t *testing.T) {
	original := &APIError{
		Status:    "error",
		ErrorType: "validation_error",
		ErrorCode: "INVALID_INPUT",
		Message:   "Invalid input",
		Details:   "Missing field",
		RequestID: "req-456",
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal APIError: %v", err)
	}

	// Unmarshal back
	var decoded APIError
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal APIError: %v", err)
	}

	// Verify fields
	if decoded.Status != original.Status {
		t.Errorf("Status = %q, want %q", decoded.Status, original.Status)
	}
	if decoded.ErrorType != original.ErrorType {
		t.Errorf("ErrorType = %q, want %q", decoded.ErrorType, original.ErrorType)
	}
	if decoded.ErrorCode != original.ErrorCode {
		t.Errorf("ErrorCode = %q, want %q", decoded.ErrorCode, original.ErrorCode)
	}
	if decoded.Message != original.Message {
		t.Errorf("Message = %q, want %q", decoded.Message, original.Message)
	}
	if decoded.Details != original.Details {
		t.Errorf("Details = %q, want %q", decoded.Details, original.Details)
	}
	if decoded.RequestID != original.RequestID {
		t.Errorf("RequestID = %q, want %q", decoded.RequestID, original.RequestID)
	}
}

// TestRateLimitError tests the RateLimitError type
func TestRateLimitError(t *testing.T) {
	tests := []struct {
		name     string
		err      *RateLimitError
		expected string
	}{
		{
			name: "with retry after",
			err: &RateLimitError{
				RetryAfter: "60",
				Message:    "Rate limit exceeded",
				Limit:      100,
				Remaining:  0,
				Reset:      1234567890,
			},
			expected: "Rate limit exceeded (retry after: 60)",
		},
		{
			name: "without retry after",
			err: &RateLimitError{
				Message:   "Rate limit exceeded",
				Limit:     100,
				Remaining: 0,
			},
			expected: "Rate limit exceeded",
		},
		{
			name: "empty message",
			err: &RateLimitError{
				Message: "",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("RateLimitError.Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestValidationError tests the ValidationError type
func TestValidationError(t *testing.T) {
	tests := []struct {
		name     string
		err      *ValidationError
		expected string
	}{
		{
			name: "with field errors",
			err: &ValidationError{
				StatusCode: 400,
				Message:    "Validation failed",
				Errors: map[string][]string{
					"email": {"is required", "must be valid"},
					"age":   {"must be positive"},
				},
			},
			expected: "validation error: Validation failed",
		},
		{
			name: "without field errors",
			err: &ValidationError{
				StatusCode: 400,
				Message:    "Validation failed",
			},
			expected: "Validation failed",
		},
		{
			name: "empty errors map",
			err: &ValidationError{
				Message: "Invalid request",
				Errors:  map[string][]string{},
			},
			expected: "Invalid request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("ValidationError.Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestValidationErrorJSON tests JSON marshaling of ValidationError
func TestValidationErrorJSON(t *testing.T) {
	original := &ValidationError{
		StatusCode: 400,
		Message:    "Validation failed",
		Errors: map[string][]string{
			"email": {"is required"},
			"age":   {"must be positive"},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal ValidationError: %v", err)
	}

	var decoded ValidationError
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal ValidationError: %v", err)
	}

	if len(decoded.Errors) != len(original.Errors) {
		t.Errorf("Errors length = %d, want %d", len(decoded.Errors), len(original.Errors))
	}
}

// TestNotFoundError tests the NotFoundError type
func TestNotFoundError(t *testing.T) {
	tests := []struct {
		name     string
		err      *NotFoundError
		expected string
	}{
		{
			name: "with resource and ID",
			err: &NotFoundError{
				Resource: "server",
				ID:       "srv-123",
				Message:  "",
			},
			expected: "server with ID srv-123 not found",
		},
		{
			name: "with custom message",
			err: &NotFoundError{
				Message: "The requested resource was not found",
			},
			expected: "The requested resource was not found",
		},
		{
			name: "default message",
			err: &NotFoundError{
				Resource: "",
				ID:       "",
				Message:  "",
			},
			expected: "resource not found",
		},
		{
			name: "only resource set",
			err: &NotFoundError{
				Resource: "server",
				ID:       "",
				Message:  "",
			},
			expected: "resource not found",
		},
		{
			name: "only ID set",
			err: &NotFoundError{
				Resource: "",
				ID:       "123",
				Message:  "",
			},
			expected: "resource not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("NotFoundError.Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestUnauthorizedError tests the UnauthorizedError type
func TestUnauthorizedError(t *testing.T) {
	tests := []struct {
		name     string
		err      *UnauthorizedError
		expected string
	}{
		{
			name: "with custom message",
			err: &UnauthorizedError{
				Message: "Invalid credentials",
			},
			expected: "Invalid credentials",
		},
		{
			name: "default message",
			err: &UnauthorizedError{
				Message: "",
			},
			expected: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("UnauthorizedError.Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestForbiddenError tests the ForbiddenError type
func TestForbiddenError(t *testing.T) {
	tests := []struct {
		name     string
		err      *ForbiddenError
		expected string
	}{
		{
			name: "with resource and action",
			err: &ForbiddenError{
				Resource: "servers",
				Action:   "delete",
				Message:  "",
			},
			expected: "forbidden: cannot delete servers",
		},
		{
			name: "with custom message",
			err: &ForbiddenError{
				Message: "You don't have permission",
			},
			expected: "You don't have permission",
		},
		{
			name: "default message",
			err: &ForbiddenError{
				Resource: "",
				Action:   "",
				Message:  "",
			},
			expected: "forbidden",
		},
		{
			name: "only resource set",
			err: &ForbiddenError{
				Resource: "servers",
				Action:   "",
				Message:  "",
			},
			expected: "forbidden",
		},
		{
			name: "only action set",
			err: &ForbiddenError{
				Resource: "",
				Action:   "delete",
				Message:  "",
			},
			expected: "forbidden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("ForbiddenError.Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestInternalServerError tests the InternalServerError type
func TestInternalServerError(t *testing.T) {
	tests := []struct {
		name     string
		err      *InternalServerError
		expected string
	}{
		{
			name: "with request ID",
			err: &InternalServerError{
				StatusCode: 500,
				Message:    "Internal server error",
				RequestID:  "req-789",
			},
			expected: "Internal server error (request ID: req-789)",
		},
		{
			name: "without request ID",
			err: &InternalServerError{
				StatusCode: 500,
				Message:    "Internal server error",
			},
			expected: "Internal server error",
		},
		{
			name: "empty message",
			err: &InternalServerError{
				Message: "",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("InternalServerError.Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestConflictError tests the ConflictError type
func TestConflictError(t *testing.T) {
	tests := []struct {
		name     string
		err      *ConflictError
		expected string
	}{
		{
			name: "with resource",
			err: &ConflictError{
				Resource: "email",
				Message:  "",
			},
			expected: "conflict: email already exists",
		},
		{
			name: "with custom message",
			err: &ConflictError{
				Message: "Resource already exists",
			},
			expected: "Resource already exists",
		},
		{
			name: "empty resource and message",
			err: &ConflictError{
				Resource: "",
				Message:  "",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("ConflictError.Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestServiceUnavailableError tests the ServiceUnavailableError type
func TestServiceUnavailableError(t *testing.T) {
	tests := []struct {
		name     string
		err      *ServiceUnavailableError
		expected string
	}{
		{
			name: "with retry time",
			err: &ServiceUnavailableError{
				Message:   "Service temporarily unavailable",
				RetryTime: 60,
			},
			expected: "Service temporarily unavailable (retry in 60 seconds)",
		},
		{
			name: "without retry time",
			err: &ServiceUnavailableError{
				Message:   "Service unavailable",
				RetryTime: 0,
			},
			expected: "Service unavailable",
		},
		{
			name: "negative retry time",
			err: &ServiceUnavailableError{
				Message:   "Service unavailable",
				RetryTime: -1,
			},
			expected: "Service unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("ServiceUnavailableError.Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestIsNotFound tests the IsNotFound helper function
func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "NotFoundError",
			err:      &NotFoundError{Message: "not found"},
			expected: true,
		},
		{
			name:     "different error type",
			err:      &UnauthorizedError{Message: "unauthorized"},
			expected: false,
		},
		{
			name:     "generic error",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFound(tt.err); got != tt.expected {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestIsRateLimit tests the IsRateLimit helper function
func TestIsRateLimit(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "RateLimitError",
			err:      &RateLimitError{Message: "rate limited"},
			expected: true,
		},
		{
			name:     "different error type",
			err:      &NotFoundError{Message: "not found"},
			expected: false,
		},
		{
			name:     "generic error",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRateLimit(tt.err); got != tt.expected {
				t.Errorf("IsRateLimit() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestIsUnauthorized tests the IsUnauthorized helper function
func TestIsUnauthorized(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "UnauthorizedError",
			err:      &UnauthorizedError{Message: "unauthorized"},
			expected: true,
		},
		{
			name:     "different error type",
			err:      &ForbiddenError{Message: "forbidden"},
			expected: false,
		},
		{
			name:     "generic error",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUnauthorized(tt.err); got != tt.expected {
				t.Errorf("IsUnauthorized() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestIsForbidden tests the IsForbidden helper function
func TestIsForbidden(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "ForbiddenError",
			err:      &ForbiddenError{Message: "forbidden"},
			expected: true,
		},
		{
			name:     "different error type",
			err:      &UnauthorizedError{Message: "unauthorized"},
			expected: false,
		},
		{
			name:     "generic error",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsForbidden(tt.err); got != tt.expected {
				t.Errorf("IsForbidden() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestIsValidation tests the IsValidation helper function
func TestIsValidation(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "ValidationError",
			err:      &ValidationError{Message: "validation failed"},
			expected: true,
		},
		{
			name:     "different error type",
			err:      &NotFoundError{Message: "not found"},
			expected: false,
		},
		{
			name:     "generic error",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidation(tt.err); got != tt.expected {
				t.Errorf("IsValidation() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestIsConflict tests the IsConflict helper function
func TestIsConflict(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "ConflictError",
			err:      &ConflictError{Message: "conflict"},
			expected: true,
		},
		{
			name:     "different error type",
			err:      &NotFoundError{Message: "not found"},
			expected: false,
		},
		{
			name:     "generic error",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsConflict(tt.err); got != tt.expected {
				t.Errorf("IsConflict() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestIsServerError tests the IsServerError helper function
func TestIsServerError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "InternalServerError",
			err:      &InternalServerError{Message: "server error"},
			expected: true,
		},
		{
			name:     "ServiceUnavailableError",
			err:      &ServiceUnavailableError{Message: "unavailable"},
			expected: true,
		},
		{
			name:     "client error",
			err:      &NotFoundError{Message: "not found"},
			expected: false,
		},
		{
			name:     "generic error",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsServerError(tt.err); got != tt.expected {
				t.Errorf("IsServerError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestErrorInterfaceCompliance tests that all error types implement error interface
func TestErrorInterfaceCompliance(t *testing.T) {
	var _ error = &APIError{}
	var _ error = &RateLimitError{}
	var _ error = &ValidationError{}
	var _ error = &NotFoundError{}
	var _ error = &UnauthorizedError{}
	var _ error = &ForbiddenError{}
	var _ error = &InternalServerError{}
	var _ error = &ConflictError{}
	var _ error = &ServiceUnavailableError{}
}

// TestErrUnexpectedResponse tests the common error variable
func TestErrUnexpectedResponse(t *testing.T) {
	if ErrUnexpectedResponse == nil {
		t.Error("ErrUnexpectedResponse should not be nil")
	}

	expectedMsg := "unexpected response format from API"
	if ErrUnexpectedResponse.Error() != expectedMsg {
		t.Errorf("ErrUnexpectedResponse.Error() = %q, want %q", ErrUnexpectedResponse.Error(), expectedMsg)
	}
}

// TestErrorTypesAsValues tests that error types can be used as values
func TestErrorTypesAsValues(t *testing.T) {
	// Test that we can create and use error values
	notFoundErr := &NotFoundError{Resource: "test", ID: "123"}
	rateLimitErr := &RateLimitError{Message: "limited"}
	validationErr := &ValidationError{Message: "invalid"}

	// Test that we can call Error() on them
	if notFoundErr.Error() == "" {
		t.Error("NotFoundError.Error() should not be empty")
	}
	if rateLimitErr.Error() == "" {
		t.Error("RateLimitError.Error() should not be empty")
	}
	if validationErr.Error() == "" {
		t.Error("ValidationError.Error() should not be empty")
	}

	// Test that we can use them in error type assertions
	var err error
	err = notFoundErr
	if !IsNotFound(err) {
		t.Error("IsNotFound should return true for NotFoundError")
	}

	err = rateLimitErr
	if !IsRateLimit(err) {
		t.Error("IsRateLimit should return true for RateLimitError")
	}

	err = validationErr
	if !IsValidation(err) {
		t.Error("IsValidation should return true for ValidationError")
	}
}
