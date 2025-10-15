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
		{
			name: "unified api key bearer token",
			config: &Config{
				Auth: AuthConfig{
					UnifiedAPIKey: "test-unified-key",
				},
			},
		},
		{
			name: "unified api key with secret",
			config: &Config{
				Auth: AuthConfig{
					UnifiedAPIKey: "test-unified-key",
					APIKeySecret:  "test-secret",
				},
			},
		},
		{
			name: "registration key",
			config: &Config{
				Auth: AuthConfig{
					RegistrationKey: "test-registration-key",
				},
			},
		},
		{
			name: "monitoring key",
			config: &Config{
				Auth: AuthConfig{
					MonitoringKey: "test-monitoring-key",
				},
			},
		},
		{
			name: "custom headers",
			config: &Config{
				Headers: map[string]string{
					"X-Custom-Header": "custom-value",
					"X-Another":       "another-value",
				},
			},
		},
		{
			name: "custom http client",
			config: &Config{
				HTTPClient: &http.Client{
					Timeout: 45 * time.Second,
				},
			},
		},
		{
			name: "debug mode enabled",
			config: &Config{
				Debug: true,
			},
		},
		{
			name: "custom retry configuration",
			config: &Config{
				RetryCount:    5,
				RetryWaitTime: 2 * time.Second,
				RetryMaxWait:  60 * time.Second,
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

			// Verify all service clients are initialized
			assert.NotNil(t, client.Organizations)
			assert.NotNil(t, client.Servers)
			assert.NotNil(t, client.Users)
			assert.NotNil(t, client.Metrics)
			assert.NotNil(t, client.Monitoring)
			assert.NotNil(t, client.Billing)
			assert.NotNil(t, client.BillingUsage)
			assert.NotNil(t, client.QuotaHistory)
			assert.NotNil(t, client.Settings)
			assert.NotNil(t, client.Alerts)
			assert.NotNil(t, client.ProbeAlerts)
			assert.NotNil(t, client.Admin)
			assert.NotNil(t, client.StatusPages)
			assert.NotNil(t, client.Providers)
			assert.NotNil(t, client.Jobs)
			assert.NotNil(t, client.BackgroundJobs)
			assert.NotNil(t, client.APIKeys)
			assert.NotNil(t, client.System)
			assert.NotNil(t, client.Terms)
			assert.NotNil(t, client.EmailQueue)
			assert.NotNil(t, client.Public)
			assert.NotNil(t, client.Distros)
			assert.NotNil(t, client.AgentDownload)
			assert.NotNil(t, client.Controllers)
			assert.NotNil(t, client.HardwareInventory)
			assert.NotNil(t, client.IPMI)
			assert.NotNil(t, client.Systemd)
			assert.NotNil(t, client.NetworkHardware)
			assert.NotNil(t, client.MonitoringDeployments)
			assert.NotNil(t, client.NamespaceDeployments)
			assert.NotNil(t, client.MonitoringAgentKeys)
			assert.NotNil(t, client.RemoteClusters)
			assert.NotNil(t, client.Health)
			assert.NotNil(t, client.ServiceMonitoring)
			assert.NotNil(t, client.Probes)
			assert.NotNil(t, client.Incidents)
			assert.NotNil(t, client.AgentVersions)
			assert.NotNil(t, client.DiskIO)
			assert.NotNil(t, client.SmartHealth)
			assert.NotNil(t, client.Filesystem)
			assert.NotNil(t, client.Tags)
			assert.NotNil(t, client.Analytics)
			assert.NotNil(t, client.ML)
			assert.NotNil(t, client.VMs)
			assert.NotNil(t, client.Reporting)
			assert.NotNil(t, client.ServerGroups)
			assert.NotNil(t, client.Search)
			assert.NotNil(t, client.Audit)
			assert.NotNil(t, client.Tasks)
			assert.NotNil(t, client.Clusters)
			assert.NotNil(t, client.Packages)
			assert.NotNil(t, client.Notifications)
			assert.NotNil(t, client.ProbeController)
			assert.NotNil(t, client.Database)

			// WebSocket service should be nil until explicitly initialized
			assert.Nil(t, client.WebSocket)

			// Verify config defaults are applied
			if tt.config == nil || tt.config.BaseURL == "" {
				assert.Equal(t, defaultBaseURL, client.config.BaseURL)
			}
			if tt.config == nil || tt.config.Timeout == 0 {
				assert.Equal(t, defaultTimeout, client.config.Timeout)
			}
			if tt.config == nil || tt.config.RetryCount == 0 {
				assert.Equal(t, 3, client.config.RetryCount)
			}
			if tt.config == nil || tt.config.RetryWaitTime == 0 {
				assert.Equal(t, 1*time.Second, client.config.RetryWaitTime)
			}
			if tt.config == nil || tt.config.RetryMaxWait == 0 {
				assert.Equal(t, 30*time.Second, client.config.RetryMaxWait)
			}
		})
	}
}

func TestClient_WithToken(t *testing.T) {
	client, err := NewClient(&Config{
		Auth: AuthConfig{
			APIKey:          "old-key",
			APISecret:       "old-secret",
			ServerUUID:      "old-uuid",
			MonitoringKey:   "old-monitoring",
			RegistrationKey: "old-registration",
		},
	})
	require.NoError(t, err)

	newClient := client.WithToken("new-token")
	assert.NotEqual(t, client, newClient)
	assert.Equal(t, "new-token", newClient.config.Auth.Token)

	// Verify all other auth methods are cleared
	assert.Empty(t, newClient.config.Auth.UnifiedAPIKey)
	assert.Empty(t, newClient.config.Auth.APIKeySecret)
	assert.Empty(t, newClient.config.Auth.APIKey)
	assert.Empty(t, newClient.config.Auth.APISecret)
	assert.Empty(t, newClient.config.Auth.ServerUUID)
	assert.Empty(t, newClient.config.Auth.ServerSecret)
	assert.Empty(t, newClient.config.Auth.MonitoringKey)
	assert.Empty(t, newClient.config.Auth.RegistrationKey)
}

func TestClient_WithUnifiedAPIKey(t *testing.T) {
	client, err := NewClient(&Config{})
	require.NoError(t, err)

	newClient := client.WithUnifiedAPIKey("unified-key")
	assert.NotEqual(t, client, newClient)
	assert.Equal(t, "unified-key", newClient.config.Auth.UnifiedAPIKey)

	// Verify all other auth methods are cleared
	assert.Empty(t, newClient.config.Auth.Token)
	assert.Empty(t, newClient.config.Auth.APIKeySecret)
	assert.Empty(t, newClient.config.Auth.APIKey)
	assert.Empty(t, newClient.config.Auth.APISecret)
	assert.Empty(t, newClient.config.Auth.ServerUUID)
	assert.Empty(t, newClient.config.Auth.ServerSecret)
	assert.Empty(t, newClient.config.Auth.MonitoringKey)
	assert.Empty(t, newClient.config.Auth.RegistrationKey)
}

func TestClient_WithUnifiedAPIKeyAndSecret(t *testing.T) {
	client, err := NewClient(&Config{})
	require.NoError(t, err)

	newClient := client.WithUnifiedAPIKeyAndSecret("unified-key", "unified-secret")
	assert.NotEqual(t, client, newClient)
	assert.Equal(t, "unified-key", newClient.config.Auth.UnifiedAPIKey)
	assert.Equal(t, "unified-secret", newClient.config.Auth.APIKeySecret)

	// Verify all other auth methods are cleared
	assert.Empty(t, newClient.config.Auth.Token)
	assert.Empty(t, newClient.config.Auth.APIKey)
	assert.Empty(t, newClient.config.Auth.APISecret)
	assert.Empty(t, newClient.config.Auth.ServerUUID)
	assert.Empty(t, newClient.config.Auth.ServerSecret)
	assert.Empty(t, newClient.config.Auth.MonitoringKey)
	assert.Empty(t, newClient.config.Auth.RegistrationKey)
}

func TestClient_WithRegistrationKey(t *testing.T) {
	client, err := NewClient(&Config{})
	require.NoError(t, err)

	newClient := client.WithRegistrationKey("registration-key")
	assert.NotEqual(t, client, newClient)
	assert.Equal(t, "registration-key", newClient.config.Auth.RegistrationKey)

	// Verify all other auth methods are cleared
	assert.Empty(t, newClient.config.Auth.Token)
	assert.Empty(t, newClient.config.Auth.UnifiedAPIKey)
	assert.Empty(t, newClient.config.Auth.APIKeySecret)
	assert.Empty(t, newClient.config.Auth.APIKey)
	assert.Empty(t, newClient.config.Auth.APISecret)
	assert.Empty(t, newClient.config.Auth.ServerUUID)
	assert.Empty(t, newClient.config.Auth.ServerSecret)
	assert.Empty(t, newClient.config.Auth.MonitoringKey)
}

func TestClient_WithAPIKey(t *testing.T) {
	client, err := NewClient(&Config{})
	require.NoError(t, err)

	newClient := client.WithAPIKey("key", "secret")
	assert.NotEqual(t, client, newClient)
	assert.Equal(t, "key", newClient.config.Auth.APIKey)
	assert.Equal(t, "secret", newClient.config.Auth.APISecret)

	// Verify all other auth methods are cleared
	assert.Empty(t, newClient.config.Auth.Token)
	assert.Empty(t, newClient.config.Auth.UnifiedAPIKey)
	assert.Empty(t, newClient.config.Auth.APIKeySecret)
	assert.Empty(t, newClient.config.Auth.ServerUUID)
	assert.Empty(t, newClient.config.Auth.ServerSecret)
	assert.Empty(t, newClient.config.Auth.MonitoringKey)
	assert.Empty(t, newClient.config.Auth.RegistrationKey)
}

func TestClient_WithServerCredentials(t *testing.T) {
	client, err := NewClient(&Config{})
	require.NoError(t, err)

	newClient := client.WithServerCredentials("uuid", "secret")
	assert.NotEqual(t, client, newClient)
	assert.Equal(t, "uuid", newClient.config.Auth.ServerUUID)
	assert.Equal(t, "secret", newClient.config.Auth.ServerSecret)

	// Verify all other auth methods are cleared
	assert.Empty(t, newClient.config.Auth.Token)
	assert.Empty(t, newClient.config.Auth.UnifiedAPIKey)
	assert.Empty(t, newClient.config.Auth.APIKeySecret)
	assert.Empty(t, newClient.config.Auth.APIKey)
	assert.Empty(t, newClient.config.Auth.APISecret)
	assert.Empty(t, newClient.config.Auth.MonitoringKey)
	assert.Empty(t, newClient.config.Auth.RegistrationKey)
}

func TestClient_WithMonitoringKey(t *testing.T) {
	client, err := NewClient(&Config{})
	require.NoError(t, err)

	newClient := client.WithMonitoringKey("monitoring-key")
	assert.NotEqual(t, client, newClient)
	assert.Equal(t, "monitoring-key", newClient.config.Auth.MonitoringKey)

	// Verify all other auth methods are cleared
	assert.Empty(t, newClient.config.Auth.Token)
	assert.Empty(t, newClient.config.Auth.UnifiedAPIKey)
	assert.Empty(t, newClient.config.Auth.APIKeySecret)
	assert.Empty(t, newClient.config.Auth.APIKey)
	assert.Empty(t, newClient.config.Auth.APISecret)
	assert.Empty(t, newClient.config.Auth.ServerUUID)
	assert.Empty(t, newClient.config.Auth.ServerSecret)
	assert.Empty(t, newClient.config.Auth.RegistrationKey)
}


func TestClient_Do(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/success":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "success", "data": {"message": "ok"}}`))
		case "/with-query":
			// Verify query parameters
			assert.Equal(t, "value1", r.URL.Query().Get("key1"))
			assert.Equal(t, "value2", r.URL.Query().Get("key2"))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "success"}`))
		case "/with-headers":
			// Verify custom headers
			assert.Equal(t, "custom-value", r.Header.Get("X-Custom-Header"))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "success"}`))
		case "/with-body":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "success"}`))
		case "/error":
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"status": "error", "message": "Invalid request"}`))
		case "/not-found":
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"status": "error", "message": "Resource not found"}`))
		case "/unauthorized":
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"message": "authentication required"}`))
		case "/forbidden":
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"message": "insufficient permissions"}`))
		case "/rate-limit":
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"message": "rate limit exceeded"}`))
		case "/internal-error":
			w.Header().Set("X-Request-ID", "req-123")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "internal server error"}`))
		case "/bad-gateway":
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte(`{"message": "bad gateway"}`))
		case "/service-unavailable":
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"message": "service unavailable"}`))
		case "/gateway-timeout":
			w.WriteHeader(http.StatusGatewayTimeout)
			w.Write([]byte(`{"message": "gateway timeout"}`))
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
		assert.Equal(t, "success", result.Status)

		if data, ok := result.Data.(map[string]interface{}); ok {
			assert.Equal(t, "ok", data["message"])
		}
	})

	t.Run("request with query parameters", func(t *testing.T) {
		_, err := client.Do(ctx, &Request{
			Method: "GET",
			Path:   "/with-query",
			Query: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		})

		assert.NoError(t, err)
	})

	t.Run("request with custom headers", func(t *testing.T) {
		_, err := client.Do(ctx, &Request{
			Method: "GET",
			Path:   "/with-headers",
			Headers: map[string]string{
				"X-Custom-Header": "custom-value",
			},
		})

		assert.NoError(t, err)
	})

	t.Run("request with body", func(t *testing.T) {
		body := map[string]interface{}{
			"key": "value",
		}

		_, err := client.Do(ctx, &Request{
			Method: "POST",
			Path:   "/with-body",
			Body:   body,
		})

		assert.NoError(t, err)
	})

	t.Run("bad request error", func(t *testing.T) {
		_, err := client.Do(ctx, &Request{
			Method: "GET",
			Path:   "/error",
		})

		assert.Error(t, err)
		validationErr, ok := err.(*ValidationError)
		require.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, validationErr.StatusCode)
	})

	t.Run("not found error", func(t *testing.T) {
		_, err := client.Do(ctx, &Request{
			Method: "GET",
			Path:   "/not-found",
		})

		assert.Error(t, err)
		_, ok := err.(*NotFoundError)
		require.True(t, ok)
	})

	t.Run("unauthorized error", func(t *testing.T) {
		_, err := client.Do(ctx, &Request{
			Method: "GET",
			Path:   "/unauthorized",
		})

		assert.Error(t, err)
		_, ok := err.(*UnauthorizedError)
		require.True(t, ok)
	})

	t.Run("forbidden error", func(t *testing.T) {
		_, err := client.Do(ctx, &Request{
			Method: "GET",
			Path:   "/forbidden",
		})

		assert.Error(t, err)
		_, ok := err.(*ForbiddenError)
		require.True(t, ok)
	})

	t.Run("rate limit error", func(t *testing.T) {
		_, err := client.Do(ctx, &Request{
			Method: "GET",
			Path:   "/rate-limit",
		})

		assert.Error(t, err)
		rateLimitErr, ok := err.(*RateLimitError)
		require.True(t, ok)
		assert.Equal(t, "60", rateLimitErr.RetryAfter)
	})

	t.Run("internal server error", func(t *testing.T) {
		_, err := client.Do(ctx, &Request{
			Method: "GET",
			Path:   "/internal-error",
		})

		assert.Error(t, err)
		internalErr, ok := err.(*InternalServerError)
		require.True(t, ok)
		assert.Equal(t, http.StatusInternalServerError, internalErr.StatusCode)
		assert.Equal(t, "req-123", internalErr.RequestID)
	})

	t.Run("bad gateway error", func(t *testing.T) {
		_, err := client.Do(ctx, &Request{
			Method: "GET",
			Path:   "/bad-gateway",
		})

		assert.Error(t, err)
		internalErr, ok := err.(*InternalServerError)
		require.True(t, ok)
		assert.Equal(t, http.StatusBadGateway, internalErr.StatusCode)
	})

	t.Run("service unavailable error", func(t *testing.T) {
		_, err := client.Do(ctx, &Request{
			Method: "GET",
			Path:   "/service-unavailable",
		})

		assert.Error(t, err)
		internalErr, ok := err.(*InternalServerError)
		require.True(t, ok)
		assert.Equal(t, http.StatusServiceUnavailable, internalErr.StatusCode)
	})

	t.Run("gateway timeout error", func(t *testing.T) {
		_, err := client.Do(ctx, &Request{
			Method: "GET",
			Path:   "/gateway-timeout",
		})

		assert.Error(t, err)
		internalErr, ok := err.(*InternalServerError)
		require.True(t, ok)
		assert.Equal(t, http.StatusGatewayTimeout, internalErr.StatusCode)
	})
}

func TestClient_Do_DebugMode(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "success"}`))
	}))
	defer server.Close()

	// Create client with debug mode enabled
	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Debug:   true,
		Auth: AuthConfig{
			ServerUUID:   "test-uuid",
			ServerSecret: "test-secret",
		},
	})
	require.NoError(t, err)

	ctx := context.Background()

	// Test request with debug logging
	_, err = client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/success",
		Body:   map[string]string{"test": "data"},
	})

	assert.NoError(t, err)
}

func TestClient_HandleError(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		responseBody   string
		headers        map[string]string
		expectedErrType interface{}
		checkError     func(t *testing.T, err error)
	}{
		{
			name:           "api error with details",
			statusCode:     http.StatusBadRequest,
			responseBody:   `{"error": "validation_error", "message": "Invalid input", "details": "Field required"}`,
			expectedErrType: &APIError{},
			checkError: func(t *testing.T, err error) {
				apiErr, ok := err.(*APIError)
				require.True(t, ok)
				assert.Equal(t, "validation_error", apiErr.ErrorType)
				assert.Equal(t, "Invalid input", apiErr.Message)
			},
		},
		{
			name:           "validation error",
			statusCode:     http.StatusBadRequest,
			responseBody:   `{"message": "validation failed"}`,
			expectedErrType: &ValidationError{},
		},
		{
			name:           "unauthorized with message",
			statusCode:     http.StatusUnauthorized,
			responseBody:   `{"message": "token expired"}`,
			expectedErrType: &UnauthorizedError{},
			checkError: func(t *testing.T, err error) {
				unAuthErr, ok := err.(*UnauthorizedError)
				require.True(t, ok)
				assert.Contains(t, unAuthErr.Message, "token expired")
			},
		},
		{
			name:           "unauthorized empty body",
			statusCode:     http.StatusUnauthorized,
			responseBody:   `{}`,
			expectedErrType: &UnauthorizedError{},
			checkError: func(t *testing.T, err error) {
				unAuthErr, ok := err.(*UnauthorizedError)
				require.True(t, ok)
				assert.Equal(t, "authentication required", unAuthErr.Message)
			},
		},
		{
			name:           "forbidden with message",
			statusCode:     http.StatusForbidden,
			responseBody:   `{"message": "access denied"}`,
			expectedErrType: &ForbiddenError{},
			checkError: func(t *testing.T, err error) {
				forbiddenErr, ok := err.(*ForbiddenError)
				require.True(t, ok)
				assert.Contains(t, forbiddenErr.Message, "access denied")
			},
		},
		{
			name:           "forbidden empty body",
			statusCode:     http.StatusForbidden,
			responseBody:   `{}`,
			expectedErrType: &ForbiddenError{},
			checkError: func(t *testing.T, err error) {
				forbiddenErr, ok := err.(*ForbiddenError)
				require.True(t, ok)
				assert.Equal(t, "insufficient permissions", forbiddenErr.Message)
			},
		},
		{
			name:           "rate limit with retry-after",
			statusCode:     http.StatusTooManyRequests,
			responseBody:   `{"message": "rate limit exceeded"}`,
			headers:        map[string]string{"Retry-After": "120"},
			expectedErrType: &RateLimitError{},
			checkError: func(t *testing.T, err error) {
				rateLimitErr, ok := err.(*RateLimitError)
				require.True(t, ok)
				assert.Equal(t, "120", rateLimitErr.RetryAfter)
			},
		},
		{
			name:           "generic error",
			statusCode:     http.StatusConflict,
			responseBody:   `{"message": "conflict"}`,
			expectedErrType: &APIError{},
			checkError: func(t *testing.T, err error) {
				apiErr, ok := err.(*APIError)
				require.True(t, ok)
				assert.Equal(t, "HTTP_409", apiErr.ErrorCode)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				for k, v := range tt.headers {
					w.Header().Set(k, v)
				}
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Debug:   true, // Enable debug to test debug logging in handleError
			})
			require.NoError(t, err)

			_, err = client.Do(context.Background(), &Request{
				Method: "GET",
				Path:   "/",
			})

			assert.Error(t, err)
			if tt.checkError != nil {
				tt.checkError(t, err)
			}
		})
	}
}

func TestClient_GetAuthMethod(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name: "JWT Token",
			config: &Config{
				Auth: AuthConfig{
					Token: "test-token",
				},
			},
			expected: "JWT Token",
		},
		{
			name: "Unified API Key (Bearer)",
			config: &Config{
				Auth: AuthConfig{
					UnifiedAPIKey: "test-key",
				},
			},
			expected: "Unified API Key (Bearer)",
		},
		{
			name: "Unified API Key (Key/Secret)",
			config: &Config{
				Auth: AuthConfig{
					UnifiedAPIKey: "test-key",
					APIKeySecret:  "test-secret",
				},
			},
			expected: "Unified API Key (Key/Secret)",
		},
		{
			name: "Registration Key",
			config: &Config{
				Auth: AuthConfig{
					RegistrationKey: "test-key",
				},
			},
			expected: "Registration Key",
		},
		{
			name: "API Key/Secret (Legacy)",
			config: &Config{
				Auth: AuthConfig{
					APIKey:    "test-key",
					APISecret: "test-secret",
				},
			},
			expected: "API Key/Secret (Legacy)",
		},
		{
			name: "Server Credentials",
			config: &Config{
				Auth: AuthConfig{
					ServerUUID:   "test-uuid",
					ServerSecret: "test-secret",
				},
			},
			expected: "Server Credentials",
		},
		{
			name: "Monitoring Key (Legacy)",
			config: &Config{
				Auth: AuthConfig{
					MonitoringKey: "test-key",
				},
			},
			expected: "Monitoring Key (Legacy)",
		},
		{
			name:     "No authentication",
			config:   &Config{},
			expected: "None",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)
			require.NoError(t, err)

			authMethod := client.getAuthMethod()
			assert.Equal(t, tt.expected, authMethod)
		})
	}
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
				require.True(t, ok)
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
			name: "healthy API with healthy flag",
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
			name: "healthy API with status only",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"status": "success",
					"data": {
						"status": "healthy",
						"version": "1.0.0"
					}
				}`))
			},
			wantErr: false,
		},
		{
			name: "healthy API with ok status",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"status": "success",
					"data": {
						"status": "ok",
						"version": "1.0.0"
					}
				}`))
			},
			wantErr: false,
		},
		{
			name: "unhealthy API",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
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
			name: "unhealthy API no status",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"status": "success",
					"data": {
						"healthy": false,
						"version": "1.0.0"
					}
				}`))
			},
			wantErr: true,
			errMsg:  "API is unhealthy",
		},
		{
			name: "API error",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{
					"status": "error",
					"error": "internal_error",
					"message": "Internal server error"
				}`))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverFunc))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL: server.URL,
				Auth: AuthConfig{
					Token: "test-token",
				},
			})
			require.NoError(t, err)

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
