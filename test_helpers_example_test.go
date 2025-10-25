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

// TestRealisticMockServer_SuccessResponse demonstrates using enhanced mock responses
func TestRealisticMockServer_SuccessResponse(t *testing.T) {
	helper := NewTestResponseHelper()
	dataGen := NewTestDataGenerator()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/servers/1", r.URL.Path)

		// Generate realistic server data
		serverData := dataGen.GenerateServerData(1, "web-01")

		// Write response with realistic headers
		helper.WriteSuccessResponse(w, serverData)
	}))
	defer server.Close()

	client, err := NewClient(&Config{BaseURL: server.URL})
	require.NoError(t, err)

	// Make request
	serverObj, err := client.Servers.Get(context.Background(), "1")
	assert.NoError(t, err)
	assert.NotNil(t, serverObj)
	assert.Equal(t, "web-01", serverObj.Hostname)
}

// TestRealisticMockServer_PaginatedResponse demonstrates pagination with metadata
func TestRealisticMockServer_PaginatedResponse(t *testing.T) {
	helper := NewTestResponseHelper()
	dataGen := NewTestDataGenerator()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/servers", r.URL.Path)

		// Generate realistic server list
		servers := []interface{}{
			dataGen.GenerateServerData(1, "web-01"),
			dataGen.GenerateServerData(2, "web-02"),
			dataGen.GenerateServerData(3, "db-01"),
		}

		// Write paginated response
		helper.WritePaginatedResponse(w, servers, 1, 25, 3)
	}))
	defer server.Close()

	client, err := NewClient(&Config{BaseURL: server.URL})
	require.NoError(t, err)

	// Make request
	servers, meta, err := client.Servers.List(context.Background(), &ListOptions{Page: 1, Limit: 25})
	assert.NoError(t, err)
	assert.NotNil(t, servers)
	assert.NotNil(t, meta)
	assert.Len(t, servers, 3)
	assert.Equal(t, 3, meta.TotalItems)
	assert.Equal(t, 1, meta.Page)
	assert.False(t, meta.HasMore)
}

// TestRealisticMockServer_ValidationError demonstrates field-specific validation errors
func TestRealisticMockServer_ValidationError(t *testing.T) {
	helper := NewTestResponseHelper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)

		// Return validation error with field-specific messages
		helper.WriteValidationError(w, map[string][]string{
			"name":  {"Name is required", "Name must be at least 3 characters"},
			"email": {"Email format is invalid"},
		})
	}))
	defer server.Close()

	client, err := NewClient(&Config{BaseURL: server.URL})
	require.NoError(t, err)

	// Make request that triggers validation error
	_, err = client.Organizations.Create(context.Background(), &Organization{Name: "ab"})
	assert.Error(t, err)

	// Verify error contains validation details
	apiErr, ok := err.(*ValidationError)
	assert.True(t, ok, "Expected ValidationError")
	if ok {
		assert.Contains(t, apiErr.Message, "Validation failed")
	}
}

// TestRealisticMockServer_RateLimitError demonstrates rate limiting
func TestRealisticMockServer_RateLimitError(t *testing.T) {
	helper := NewTestResponseHelper().WithRateLimit(1000, 999)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate rate limit exceeded
		helper.WriteRateLimitError(w)
	}))
	defer server.Close()

	client, err := NewClient(&Config{BaseURL: server.URL})
	require.NoError(t, err)

	// Make request that hits rate limit
	_, _, err = client.Servers.List(context.Background(), nil)
	assert.Error(t, err)

	// Verify error is rate limit error
	rateLimitErr, ok := err.(*RateLimitError)
	assert.True(t, ok, "Expected RateLimitError")
	if ok {
		assert.Contains(t, rateLimitErr.Message, "rate limit")
	}
}

// TestErrorScenarios_AllTypes demonstrates all error scenarios
func TestErrorScenarios_AllTypes(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(w http.ResponseWriter)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "unauthorized",
			setupMock: func(w http.ResponseWriter) {
				NewErrorScenarios().WriteUnauthorizedError(w)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name: "forbidden",
			setupMock: func(w http.ResponseWriter) {
				NewErrorScenarios().WriteForbiddenError(w)
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "forbidden",
		},
		{
			name: "not_found",
			setupMock: func(w http.ResponseWriter) {
				NewErrorScenarios().WriteNotFoundError(w, "Server")
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "not_found",
		},
		{
			name: "conflict",
			setupMock: func(w http.ResponseWriter) {
				NewErrorScenarios().WriteConflictError(w, "Server name already exists")
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "conflict",
		},
		{
			name: "server_error",
			setupMock: func(w http.ResponseWriter) {
				NewErrorScenarios().WriteServerError(w)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "internal_error",
		},
		{
			name: "service_unavailable",
			setupMock: func(w http.ResponseWriter) {
				NewErrorScenarios().WriteServiceUnavailableError(w)
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedError:  "service_unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tt.setupMock(w)
			}))
			defer server.Close()

			client, err := NewClient(&Config{BaseURL: server.URL})
			require.NoError(t, err)

			// Make request
			_, err = client.Servers.Get(context.Background(), "1")
			assert.Error(t, err)

			// Verify error type
			apiErr, ok := err.(*APIError)
			if ok {
				assert.Equal(t, tt.expectedError, apiErr.ErrorType)
			}
		})
	}
}

// TestRealisticHeaders_VerifyPresence ensures all realistic headers are present
func TestRealisticHeaders_VerifyPresence(t *testing.T) {
	helper := NewTestResponseHelper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		helper.AddRealisticHeaders(w)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	}))
	defer server.Close()

	// Make HTTP request directly to inspect headers
	resp, err := http.Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify all realistic headers are present
	assert.NotEmpty(t, resp.Header.Get("X-Request-ID"), "X-Request-ID header should be present")
	assert.NotEmpty(t, resp.Header.Get("X-RateLimit-Limit"), "X-RateLimit-Limit header should be present")
	assert.NotEmpty(t, resp.Header.Get("X-RateLimit-Remaining"), "X-RateLimit-Remaining header should be present")
	assert.NotEmpty(t, resp.Header.Get("X-RateLimit-Reset"), "X-RateLimit-Reset header should be present")
	assert.Equal(t, "42ms", resp.Header.Get("X-Response-Time"))
	assert.Equal(t, "Nexmonyx/2.0", resp.Header.Get("Server"))
}

// TestDataGenerator_RealisticTimestamps verifies timestamp generation
func TestDataGenerator_RealisticTimestamps(t *testing.T) {
	gen := NewTestDataGenerator()

	// Generate timestamps at different offsets
	past := gen.GenerateTimestamp(-24)
	present := gen.GenerateTimestamp(0)
	future := gen.GenerateTimestamp(24)

	// Verify ordering
	assert.True(t, past.Before(present), "Past timestamp should be before present")
	assert.True(t, present.Before(future), "Present timestamp should be before future")

	// Verify 24-hour offset
	expectedDiff := 24 * time.Hour
	assert.Equal(t, expectedDiff, present.Sub(past))
	assert.Equal(t, expectedDiff, future.Sub(present))
}

// TestPaginationMeta_Calculations verifies pagination metadata calculations
func TestPaginationMeta_Calculations(t *testing.T) {
	tests := []struct {
		name              string
		page              int
		limit             int
		totalItems        int
		expectedTotalPages int
		expectedHasMore    bool
		expectedFrom       int
		expectedTo         int
	}{
		{
			name:              "first page with more pages",
			page:              1,
			limit:             25,
			totalItems:        100,
			expectedTotalPages: 4,
			expectedHasMore:    true,
			expectedFrom:       1,
			expectedTo:         25,
		},
		{
			name:              "middle page",
			page:              2,
			limit:             25,
			totalItems:        100,
			expectedTotalPages: 4,
			expectedHasMore:    true,
			expectedFrom:       26,
			expectedTo:         50,
		},
		{
			name:              "last page partial",
			page:              4,
			limit:             25,
			totalItems:        100,
			expectedTotalPages: 4,
			expectedHasMore:    false,
			expectedFrom:       76,
			expectedTo:         100,
		},
		{
			name:              "single page",
			page:              1,
			limit:             25,
			totalItems:        10,
			expectedTotalPages: 1,
			expectedHasMore:    false,
			expectedFrom:       1,
			expectedTo:         10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helper := NewTestResponseHelper()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				helper.WritePaginatedResponse(w, []interface{}{}, tt.page, tt.limit, tt.totalItems)
			}))
			defer server.Close()

			// Make HTTP request and parse response
			resp, err := http.Get(server.URL)
			require.NoError(t, err)
			defer resp.Body.Close()

			var result PaginatedResponse
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			// Verify pagination metadata
			assert.Equal(t, tt.page, result.Meta.Page)
			assert.Equal(t, tt.limit, result.Meta.Limit)
			assert.Equal(t, tt.totalItems, result.Meta.TotalItems)
			assert.Equal(t, tt.expectedTotalPages, result.Meta.TotalPages)
			assert.Equal(t, tt.expectedHasMore, result.Meta.HasMore)
			assert.Equal(t, tt.expectedFrom, result.Meta.From)
			assert.Equal(t, tt.expectedTo, result.Meta.To)
		})
	}
}
