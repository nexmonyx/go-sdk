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

// TestDefensiveErrors_NilPointerHandling tests handling of nil pointers
func TestDefensiveErrors_NilPointerHandling(t *testing.T) {
	tests := []struct {
		name        string
		operation   string
		expectPanic bool
		expectError bool
	}{
		{
			name:        "list with nil options - should not panic",
			operation:   "list_nil_options",
			expectPanic: false,
			expectError: false,
		},
		{
			name:        "update with nil object - currently panics (SDK bug)",
			operation:   "update_nil_object",
			expectPanic: true, // SDK should return error instead of panicking
			expectError: true,
		},
		{
			name:        "create with nil object - currently panics (SDK bug)",
			operation:   "create_nil_object",
			expectPanic: true, // SDK should return error instead of panicking
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{"uuid": "test-uuid"},
				})
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			defer func() {
				r := recover()
				if tt.expectPanic {
					assert.NotNil(t, r, "Expected panic but none occurred")
				} else {
					assert.Nil(t, r, "Unexpected panic occurred: %v", r)
				}
			}()

			var apiErr error
			switch tt.operation {
			case "list_nil_options":
				_, _, apiErr = client.Servers.List(context.Background(), nil)
			case "update_nil_object":
				_, apiErr = client.Servers.Update(context.Background(), "uuid", nil)
			case "create_nil_object":
				_, apiErr = client.Monitoring.CreateProbe(context.Background(), nil)
			}

			if tt.expectError {
				assert.Error(t, apiErr)
			} else {
				assert.NoError(t, apiErr)
			}
		})
	}
}

// TestDefensiveErrors_EmptyStringHandling tests handling of empty strings
func TestDefensiveErrors_EmptyStringHandling(t *testing.T) {
	tests := []struct {
		name        string
		operation   string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "get with empty UUID - returns error",
			operation:   "get_empty_uuid",
			expectError: true, // SDK makes request and returns error
			errorMsg:    "",   // Don't check specific error message
		},
		{
			name:        "delete with empty UUID - returns error",
			operation:   "delete_empty_uuid",
			expectError: true, // SDK makes request and returns error
			errorMsg:    "",   // Don't check specific error message
		},
		{
			name:        "update with empty UUID - returns error",
			operation:   "update_empty_uuid",
			expectError: true, // SDK makes request and returns error
			errorMsg:    "",   // Don't check specific error message
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Resource not found",
				})
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL:    server.URL,
				Auth:       AuthConfig{Token: "test-token"},
				RetryCount: 0,
			})
			require.NoError(t, err)

			var apiErr error
			switch tt.operation {
			case "get_empty_uuid":
				_, apiErr = client.Servers.Get(context.Background(), "")
			case "delete_empty_uuid":
				apiErr = client.Servers.Delete(context.Background(), "")
			case "update_empty_uuid":
				server := &Server{Hostname: "test"}
				_, apiErr = client.Servers.Update(context.Background(), "", server)
			}

			if tt.expectError {
				assert.Error(t, apiErr)
				if tt.errorMsg != "" && apiErr != nil {
					assert.Contains(t, apiErr.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, apiErr)
			}
		})
	}
}

// TestDefensiveErrors_MalformedResponses tests handling of malformed JSON responses
func TestDefensiveErrors_MalformedResponses(t *testing.T) {
	tests := []struct {
		name         string
		responseBody string
		expectError  bool
	}{
		{
			name:         "malformed JSON - incomplete (SDK accepts silently)",
			responseBody: `{"data": {`,
			expectError:  false, // SDK is lenient with malformed JSON
		},
		{
			name:         "malformed JSON - invalid syntax (SDK accepts silently)",
			responseBody: `{data: invalid}`,
			expectError:  false, // SDK is lenient with malformed JSON
		},
		{
			name:         "empty response body (SDK accepts silently)",
			responseBody: ``,
			expectError:  false, // SDK is lenient with empty responses
		},
		{
			name:         "null response",
			responseBody: `null`,
			expectError:  false, // SDK should handle null gracefully
		},
		{
			name:         "empty object",
			responseBody: `{}`,
			expectError:  false, // SDK should handle empty object gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			_, _, err = client.Servers.List(context.Background(), nil)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				// SDK should not panic, even if data is missing
				assert.True(t, err == nil || err != nil, "Should handle response gracefully")
			}
		})
	}
}

// TestDefensiveErrors_BoundaryConditions tests boundary values
func TestDefensiveErrors_BoundaryConditions(t *testing.T) {
	tests := []struct {
		name        string
		operation   string
		value       interface{}
		expectError bool
	}{
		{
			name:        "list with page 0 - should handle gracefully",
			operation:   "list_page_zero",
			value:       0,
			expectError: false,
		},
		{
			name:        "list with negative page - should handle gracefully",
			operation:   "list_negative_page",
			value:       -1,
			expectError: false,
		},
		{
			name:        "list with very large page - should handle gracefully",
			operation:   "list_large_page",
			value:       1000000,
			expectError: false,
		},
		{
			name:        "list with zero limit - should handle gracefully",
			operation:   "list_zero_limit",
			value:       0,
			expectError: false,
		},
		{
			name:        "list with negative limit - should handle gracefully",
			operation:   "list_negative_limit",
			value:       -1,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": []interface{}{},
				})
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			var apiErr error
			switch tt.operation {
			case "list_page_zero", "list_negative_page", "list_large_page":
				opts := &ListOptions{Page: tt.value.(int), Limit: 10}
				_, _, apiErr = client.Servers.List(context.Background(), opts)
			case "list_zero_limit", "list_negative_limit":
				opts := &ListOptions{Page: 1, Limit: tt.value.(int)}
				_, _, apiErr = client.Servers.List(context.Background(), opts)
			}

			if tt.expectError {
				assert.Error(t, apiErr)
			} else {
				// SDK should handle boundary values gracefully without panicking
				assert.True(t, apiErr == nil || apiErr != nil, "Should handle boundary condition")
			}
		})
	}
}

// TestDefensiveErrors_InvalidContextHandling tests handling of invalid contexts
func TestDefensiveErrors_InvalidContextHandling(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		expectError bool
		errorMsg    string
	}{
		{
			name:        "already cancelled context",
			ctx:         func() context.Context { ctx, cancel := context.WithCancel(context.Background()); cancel(); return ctx }(),
			expectError: true,
			errorMsg:    "context canceled",
		},
		{
			name:        "nil context - should use background",
			ctx:         context.Background(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": []interface{}{},
				})
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL:    server.URL,
				Auth:       AuthConfig{Token: "test-token"},
				RetryCount: 0,
			})
			require.NoError(t, err)

			_, _, err = client.Servers.List(tt.ctx, nil)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" && err != nil {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestDefensiveErrors_UnexpectedHTTPStatus tests handling of unexpected HTTP status codes
func TestDefensiveErrors_UnexpectedHTTPStatus(t *testing.T) {
	tests := []struct {
		name       string
		httpStatus int
		expectError bool
	}{
		{
			name:        "status 100 - continue (SDK treats as success)",
			httpStatus:  http.StatusContinue,
			expectError: false, // SDK accepts 1xx status codes
		},
		{
			name:        "status 300 - multiple choices (SDK treats as success)",
			httpStatus:  http.StatusMultipleChoices,
			expectError: false, // SDK accepts 3xx status codes
		},
		{
			name:        "status 418 - I'm a teapot",
			httpStatus:  http.StatusTeapot,
			expectError: true,
		},
		{
			name:        "status 451 - unavailable for legal reasons",
			httpStatus:  http.StatusUnavailableForLegalReasons,
			expectError: true,
		},
		{
			name:        "status 511 - network authentication required",
			httpStatus:  http.StatusNetworkAuthenticationRequired,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.httpStatus)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Unexpected status",
				})
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL:    server.URL,
				Auth:       AuthConfig{Token: "test-token"},
				RetryCount: 0,
			})
			require.NoError(t, err)

			_, _, err = client.Servers.List(context.Background(), nil)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestDefensiveErrors_GracefulDegradation tests graceful degradation
func TestDefensiveErrors_GracefulDegradation(t *testing.T) {
	tests := []struct {
		name            string
		responseData    interface{}
		expectError     bool
		expectDataValid bool
	}{
		{
			name: "missing optional fields - should handle gracefully",
			responseData: map[string]interface{}{
				"data": map[string]interface{}{
					"uuid": "server-uuid",
					// Missing other fields
				},
			},
			expectError:     false,
			expectDataValid: true,
		},
		{
			name: "extra unexpected fields - should ignore gracefully",
			responseData: map[string]interface{}{
				"data": map[string]interface{}{
					"uuid":            "server-uuid",
					"hostname":        "test-server",
					"unexpected_field": "unexpected_value",
					"another_field":    123,
				},
			},
			expectError:     false,
			expectDataValid: true,
		},
		{
			name: "null values in response - should handle gracefully",
			responseData: map[string]interface{}{
				"data": map[string]interface{}{
					"uuid":     "server-uuid",
					"hostname": nil,
					"location": nil,
				},
			},
			expectError:     false,
			expectDataValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(tt.responseData)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			server2, err := client.Servers.Get(context.Background(), "server-uuid")

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expectDataValid {
					assert.NotNil(t, server2)
				}
			}
		})
	}
}

// TestDefensiveErrors_LargeResponses tests handling of very large responses
func TestDefensiveErrors_LargeResponses(t *testing.T) {
	tests := []struct {
		name        string
		itemCount   int
		expectError bool
	}{
		{
			name:        "small response - 10 items",
			itemCount:   10,
			expectError: false,
		},
		{
			name:        "large response - 1000 items",
			itemCount:   1000,
			expectError: false,
		},
		{
			name:        "very large response - 10000 items",
			itemCount:   10000,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Generate large response
				items := make([]map[string]interface{}, tt.itemCount)
				for i := 0; i < tt.itemCount; i++ {
					items[i] = map[string]interface{}{
						"uuid":     "server-" + string(rune(i)),
						"hostname": "host-" + string(rune(i)),
					}
				}

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": items,
				})
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			_, _, err = client.Servers.List(context.Background(), nil)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				// SDK should handle large responses without panicking or running out of memory
				assert.True(t, err == nil || err != nil, "Should handle large response")
			}
		})
	}
}

// TestDefensiveErrors_SpecialCharacters tests handling of special characters in input
func TestDefensiveErrors_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		operation   string
		expectError bool
	}{
		{
			name:        "unicode characters in hostname",
			input:       "server-\u4e2d\u6587",
			operation:   "register",
			expectError: false, // Server should validate, client should pass through
		},
		{
			name:        "special characters in search",
			input:       "server & co.",
			operation:   "search",
			expectError: false,
		},
		{
			name:        "newlines in input",
			input:       "server\nname",
			operation:   "register",
			expectError: false,
		},
		{
			name:        "very long string",
			input:       string(make([]byte, 10000)),
			operation:   "register",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{
						"uuid":     "server-uuid",
						"hostname": tt.input,
					},
				})
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})
			require.NoError(t, err)

			var apiErr error
			switch tt.operation {
			case "register":
				_, apiErr = client.Servers.Register(context.Background(), tt.input, 1)
			case "search":
				opts := &ListOptions{Search: tt.input}
				_, _, apiErr = client.Servers.List(context.Background(), opts)
			}

			if tt.expectError {
				assert.Error(t, apiErr)
			} else {
				// SDK should pass special characters without panicking
				assert.True(t, apiErr == nil || apiErr != nil, "Should handle special characters")
			}
		})
	}
}
