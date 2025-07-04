package nexmonyx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:   "default config",
			config: nil,
		},
		{
			name: "custom config",
			config: &Config{
				BaseURL: "https://custom.api.com",
				Auth: AuthConfig{
					Token: "test-token",
				},
				Timeout: 60 * time.Second,
			},
		},
		{
			name: "api key auth",
			config: &Config{
				Auth: AuthConfig{
					APIKey:    "test-key",
					APISecret: "test-secret",
				},
			},
		},
		{
			name: "server credentials",
			config: &Config{
				Auth: AuthConfig{
					ServerUUID:   "test-uuid",
					ServerSecret: "test-secret",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, client)
			assert.NotNil(t, client.Organizations)
			assert.NotNil(t, client.Servers)
			assert.NotNil(t, client.Users)
			assert.NotNil(t, client.Metrics)
			assert.NotNil(t, client.Monitoring)
			assert.NotNil(t, client.Billing)
			assert.NotNil(t, client.Settings)
			assert.NotNil(t, client.Alerts)
		})
	}
}

func TestClient_WithToken(t *testing.T) {
	client, err := NewClient(&Config{})
	require.NoError(t, err)

	newClient := client.WithToken("new-token")
	assert.NotEqual(t, client, newClient)
	assert.Equal(t, "new-token", newClient.config.Auth.Token)
	assert.Empty(t, newClient.config.Auth.APIKey)
}

func TestClient_WithAPIKey(t *testing.T) {
	client, err := NewClient(&Config{})
	require.NoError(t, err)

	newClient := client.WithAPIKey("key", "secret")
	assert.NotEqual(t, client, newClient)
	assert.Equal(t, "key", newClient.config.Auth.APIKey)
	assert.Equal(t, "secret", newClient.config.Auth.APISecret)
	assert.Empty(t, newClient.config.Auth.Token)
}

func TestClient_WithServerCredentials(t *testing.T) {
	client, err := NewClient(&Config{})
	require.NoError(t, err)

	newClient := client.WithServerCredentials("uuid", "secret")
	assert.NotEqual(t, client, newClient)
	assert.Equal(t, "uuid", newClient.config.Auth.ServerUUID)
	assert.Equal(t, "secret", newClient.config.Auth.ServerSecret)
	assert.Empty(t, newClient.config.Auth.Token)
}

func TestClient_Do(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/success":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "success", "data": {"message": "ok"}}`))
		case "/error":
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"status": "error", "error": "bad_request", "message": "Invalid request"}`))
		case "/not-found":
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"status": "error", "error": "not_found", "message": "Resource not found"}`))
		}
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("successful request", func(t *testing.T) {
		var result StandardResponse

		resp, err := client.Do(ctx, &Request{
			Method: "GET",
			Path:   "/success",
			Result: &result,
		})

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Debug: print the response body
		t.Logf("Response body: %s", resp.Body)

		assert.Equal(t, "success", result.Status)

		// Check data field
		if data, ok := result.Data.(map[string]interface{}); ok {
			assert.Equal(t, "ok", data["message"])
		} else {
			t.Fatal("Expected data to be a map")
		}
	})

	t.Run("error request", func(t *testing.T) {
		_, err := client.Do(ctx, &Request{
			Method: "GET",
			Path:   "/error",
		})

		assert.Error(t, err)
	})

	t.Run("not found request", func(t *testing.T) {
		_, err := client.Do(ctx, &Request{
			Method: "GET",
			Path:   "/not-found",
		})

		assert.Error(t, err)
	})
}

func TestListOptions_ToQuery(t *testing.T) {
	opts := &ListOptions{
		Page:   2,
		Limit:  50,
		Sort:   "name",
		Order:  "desc",
		Search: "test",
		Filters: map[string]string{
			"status": "active",
			"type":   "server",
		},
	}

	query := opts.ToQuery()

	expected := map[string]string{
		"page":   "2",
		"limit":  "50",
		"sort":   "name",
		"order":  "desc",
		"search": "test",
		"status": "active",
		"type":   "server",
	}

	assert.Equal(t, expected, query)
}

func TestCustomTime_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		expected string
	}{
		{
			name:     "RFC3339 format",
			input:    `"2023-01-01T12:00:00Z"`,
			expected: "2023-01-01T12:00:00Z",
		},
		{
			name:     "RFC3339 with milliseconds",
			input:    `"2023-01-01T12:00:00.000Z"`,
			expected: "2023-01-01T12:00:00Z",
		},
		{
			name:     "null value",
			input:    `"null"`,
			expected: "",
		},
		{
			name:     "empty string",
			input:    `""`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ct CustomTime
			err := ct.UnmarshalJSON([]byte(tt.input))

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			if tt.expected != "" {
				expected, _ := time.Parse(time.RFC3339, tt.expected)
				assert.True(t, ct.Time.Equal(expected))
			}
		})
	}
}

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		apiError APIError
		expected string
	}{
		{
			name: "with details",
			apiError: APIError{
				ErrorCode: "validation_error",
				Message:   "Invalid input",
				Details:   "Field 'email' is required",
			},
			expected: "validation_error: Invalid input (Field 'email' is required)",
		},
		{
			name: "without details",
			apiError: APIError{
				ErrorCode: "not_found",
				Message:   "Resource not found",
			},
			expected: "not_found: Resource not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.apiError.Error())
		})
	}
}

func TestStandardResponse_IsSuccess(t *testing.T) {
	tests := []struct {
		name     string
		response StandardResponse
		expected bool
	}{
		{
			name:     "success response",
			response: StandardResponse{Status: "success"},
			expected: true,
		},
		{
			name:     "error response",
			response: StandardResponse{Status: "error"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.response.IsSuccess())
		})
	}
}

func TestStandardResponse_GetError(t *testing.T) {
	tests := []struct {
		name     string
		response StandardResponse
		wantErr  bool
	}{
		{
			name:     "success response",
			response: StandardResponse{Status: "success"},
			wantErr:  false,
		},
		{
			name: "error response",
			response: StandardResponse{
				Status:  "error",
				Error:   "test_error",
				Message: "Test error message",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.response.GetError()
			if tt.wantErr {
				assert.Error(t, err)
				apiErr, ok := err.(*APIError)
				assert.True(t, ok)
				assert.Equal(t, "test_error", apiErr.ErrorType)
				assert.Equal(t, "Test error message", apiErr.Message)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_HealthCheck(t *testing.T) {
	tests := []struct {
		name       string
		serverFunc func(w http.ResponseWriter, r *http.Request)
		wantErr    bool
		errMsg     string
	}{
		{
			name: "healthy API",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/v1/healthz", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"status": "success",
					"data": {
						"status": "operational",
						"healthy": true,
						"version": "1.0.0",
						"timestamp": "2023-01-01T12:00:00Z"
					}
				}`))
			},
			wantErr: false,
		},
		{
			name: "unhealthy API",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/v1/healthz", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"status": "success",
					"data": {
						"status": "degraded",
						"healthy": false,
						"version": "1.0.0",
						"timestamp": "2023-01-01T12:00:00Z"
					}
				}`))
			},
			wantErr: true,
			errMsg:  "API is unhealthy: degraded",
		},
		{
			name: "API error",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/v1/healthz", r.URL.Path)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{
					"status": "error",
					"error": "internal_error",
					"message": "Internal server error"
				}`))
			},
			wantErr: true,
		},
		{
			name: "network error",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				// Simulate network error by closing connection
				hj, ok := w.(http.Hijacker)
				if ok {
					conn, _, _ := hj.Hijack()
					conn.Close()
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(tt.serverFunc))
			defer server.Close()

			// Create client
			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth: AuthConfig{
					Token: "test-token",
				},
			})
			require.NoError(t, err)

			// Test health check
			ctx := context.Background()
			err = client.HealthCheck(ctx)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_HealthCheck_Context(t *testing.T) {
	// Create a test server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": "success",
			"data": {
				"status": "operational",
				"healthy": true,
				"version": "1.0.0",
				"timestamp": "2023-01-01T12:00:00Z"
			}
		}`))
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err = client.HealthCheck(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")

	// Test with timeout context
	ctx, cancel = context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err = client.HealthCheck(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}
