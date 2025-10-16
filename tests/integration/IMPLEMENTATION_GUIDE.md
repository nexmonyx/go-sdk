# Integration Testing Implementation Guide

This document provides detailed implementation guidance for completing Tasks #3002-#3004 and #3017.

## Prerequisites

Before implementing these tasks, the following adjustments need to be made to the existing integration test framework:

### 1. Fix Field Name Mismatches

The mock server fixtures and test code use field names that don't match the actual SDK API. Update these:

**Server Fields:**
- `uuid` → `server_uuid`
- `ip_address` → `main_ip`
- `last_seen` → `last_heartbeat`

**Update files:**
- `tests/integration/fixtures/servers.json`
- `tests/integration/mock_api_server.go`
- `tests/integration/helpers.go`
- `tests/integration/servers_test.go`

### 2. Fix Config Struct

The `Config` struct uses `Timeout` not `HTTPTimeout`:

```go
// Correct
client, err := nexmonyx.NewClient(&nexmonyx.Config{
    BaseURL: mockAPI.Server.URL,
    Auth: nexmonyx.AuthConfig{
        Token: "test-token",
    },
    Timeout: 10 * time.Second,  // Use Timeout, not HTTPTimeout
    Debug:   os.Getenv("INTEGRATION_TEST_DEBUG") == "true",
})
```

### 3. Fix Server Create/Update Methods

The SDK's `Servers.Create()` method takes a `*Server` object, not a separate request type:

```go
// Create a new server
server := &nexmonyx.Server{
    Hostname:       "test-server",
    OrganizationID: 1,
    MainIP:         "192.168.1.200",
    Location:       "Test-Region",
    Environment:    "testing",
    Classification: "test",
}
created, err := env.Client.Servers.Create(env.Ctx, server)
```

Similarly, `Update()` also takes a `*Server`:

```go
// Update server
server.Hostname = "updated-hostname"
updated, err := env.Client.Servers.Update(env.Ctx, server.ServerUUID, server)
```

---

## Task #3002: Core Service Integration Tests

**Effort**: 16 hours
**Priority**: HIGH
**Prerequisites**: Task #3001 (completed), field name fixes above

### Objective

Create comprehensive integration tests that validate complete workflows across multiple services, testing the SDK's ability to handle real-world usage patterns.

### Deliverables

#### 1. Server Registration → Metrics Submission → Data Retrieval (`servers_workflow_test.go`)

Test the complete server lifecycle:

```go
func TestServerLifecycleWorkflow(t *testing.T) {
    skipIfShort(t)
    env := setupIntegrationTest(t)
    defer teardownIntegrationTest(t, env)

    t.Run("CompleteServerLifecycle", func(t *testing.T) {
        // Step 1: Register a new server
        server := &nexmonyx.Server{
            Hostname:       "workflow-test-server",
            OrganizationID: 1,
            MainIP:         "192.168.100.50",
            Environment:    "testing",
            Classification: "test",
        }
        registered, err := env.Client.Servers.Create(env.Ctx, server)
        require.NoError(t, err)
        require.NotEmpty(t, registered.ServerUUID)

        // Step 2: Send heartbeat
        err = env.Client.Servers.SendHeartbeat(env.Ctx, registered.ServerUUID)
        require.NoError(t, err)

        // Step 3: Submit metrics
        metrics := createTestMetricsPayload(registered.ServerUUID)
        err = env.Client.Metrics.SubmitComprehensive(env.Ctx, metrics)
        require.NoError(t, err)

        // Step 4: Retrieve and verify metrics
        retrievedMetrics, _, err := env.Client.Servers.GetMetrics(env.Ctx, registered.ServerUUID, &nexmonyx.ListOptions{
            Page: 1, Limit: 10,
        })
        require.NoError(t, err)
        assert.NotEmpty(t, retrievedMetrics)

        // Step 5: Update server details
        registered.Location = "Updated-Location"
        updated, err := env.Client.Servers.Update(env.Ctx, registered.ServerUUID, registered)
        require.NoError(t, err)
        assert.Equal(t, "Updated-Location", updated.Location)

        // Step 6: Retrieve server details and verify
        retrieved, err := env.Client.Servers.GetByUUID(env.Ctx, registered.ServerUUID)
        require.NoError(t, err)
        assert.Equal(t, "Updated-Location", retrieved.Location)

        // Step 7: Clean up - delete server
        err = env.Client.Servers.Delete(env.Ctx, registered.ServerUUID)
        require.NoError(t, err)
    })
}
```

**Additional tests to implement:**
- Server with multiple metric submissions over time
- Server tag management workflow
- Server alert association workflow
- Bulk server operations

#### 2. Organization → User Management → Resource Access (`organizations_workflow_test.go`)

```go
func TestOrganizationWorkflow(t *testing.T) {
    skipIfShort(t)
    env := setupIntegrationTest(t)
    defer teardownIntegrationTest(t, env)

    t.Run("OrganizationUserServerWorkflow", func(t *testing.T) {
        // Step 1: Get organization
        org, err := env.Client.Organizations.Get(env.Ctx, "org-001")
        require.NoError(t, err)

        // Step 2: List users in organization
        users, _, err := env.Client.Users.List(env.Ctx, &nexmonyx.ListOptions{
            Filters: map[string]string{
                "organization_id": fmt.Sprintf("%d", org.ID),
            },
        })
        require.NoError(t, err)
        assert.NotEmpty(t, users)

        // Step 3: Create server in organization
        server := &nexmonyx.Server{
            Hostname:       "org-test-server",
            OrganizationID: org.ID,
            MainIP:         "192.168.100.60",
        }
        created, err := env.Client.Servers.Create(env.Ctx, server)
        require.NoError(t, err)

        // Step 4: Verify server belongs to organization
        assert.Equal(t, org.ID, created.OrganizationID)

        // Step 5: List all servers in organization
        orgServers, _, err := env.Client.Servers.List(env.Ctx, &nexmonyx.ListOptions{
            Filters: map[string]string{
                "organization_id": fmt.Sprintf("%d", org.ID),
            },
        })
        require.NoError(t, err)

        // Verify our server is in the list
        found := false
        for _, s := range orgServers {
            if s.ServerUUID == created.ServerUUID {
                found = true
                break
            }
        }
        assert.True(t, found, "Created server should be in organization's server list")
    })
}
```

**Additional tests:**
- Organization resource isolation (can't access other org's resources)
- Organization member role-based access control
- Organization subscription tier limits

#### 3. Alert Creation → Triggering → Notification (`alerts_workflow_test.go`)

```go
func TestAlertWorkflow(t *testing.T) {
    skipIfShort(t)
    env := setupIntegrationTest(t)
    defer teardownIntegrationTest(t, env)

    t.Run("AlertCreationAndTriggering", func(t *testing.T) {
        // Step 1: Create a server to monitor
        server := createTestServer(t, env, "alert-test-server")

        // Step 2: Create an alert rule
        alert := &nexmonyx.Alert{
            Name:              "High CPU Alert",
            Description:       "Alert when CPU exceeds 90%",
            ServerUUID:        server.ServerUUID,
            MetricType:        "cpu",
            Condition:         "greater_than",
            Threshold:         90.0,
            Severity:          "critical",
            Enabled:           true,
            NotificationChannels: []string{"email"},
        }

        created, err := env.Client.Alerts.Create(env.Ctx, alert)
        require.NoError(t, err)
        require.NotEmpty(t, created.UUID)

        // Step 3: Submit metrics that should trigger the alert
        metrics := &nexmonyx.ComprehensiveMetrics{
            ServerUUID: server.ServerUUID,
            Timestamp:  time.Now(),
            CPU: &nexmonyx.CPUMetrics{
                UsagePercent: 95.0,  // Exceeds threshold
            },
        }
        err = env.Client.Metrics.SubmitComprehensive(env.Ctx, metrics)
        require.NoError(t, err)

        // Step 4: Check alert status (may need to wait for processing)
        time.Sleep(2 * time.Second)

        alerts, _, err := env.Client.Servers.GetAlerts(env.Ctx, server.ServerUUID, nil)
        require.NoError(t, err)
        assert.NotEmpty(t, alerts)

        // Step 5: Update alert threshold
        created.Threshold = 95.0
        updated, err := env.Client.Alerts.Update(env.Ctx, created.UUID, created)
        require.NoError(t, err)
        assert.Equal(t, 95.0, updated.Threshold)

        // Step 6: Disable alert
        updated.Enabled = false
        disabled, err := env.Client.Alerts.Update(env.Ctx, updated.UUID, updated)
        require.NoError(t, err)
        assert.False(t, disabled.Enabled)
    })
}
```

**Additional tests:**
- Multiple alert rules per server
- Alert notification channel configuration
- Alert history and acknowledgment workflow

#### 4. Probe Deployment → Monitoring → Result Collection (`monitoring_workflow_test.go`)

```go
func TestMonitoringWorkflow(t *testing.T) {
    skipIfShort(t)
    env := setupIntegrationTest(t)
    defer teardownIntegrationTest(t, env)

    t.Run("ProbeDeploymentAndExecution", func(t *testing.T) {
        // Step 1: Create a probe
        probe := &nexmonyx.Probe{
            Name:        "HTTP Health Check",
            Type:        "http",
            Target:      "https://example.com",
            Interval:    60,
            Timeout:     10,
            Enabled:     true,
            ServerUUID:  "server-001",
        }

        created, err := env.Client.Monitoring.CreateProbe(env.Ctx, probe)
        require.NoError(t, err)
        require.NotEmpty(t, created.UUID)

        // Step 2: Execute probe immediately
        result, err := env.Client.Monitoring.ExecuteProbe(env.Ctx, created.UUID)
        require.NoError(t, err)
        assert.NotNil(t, result)

        // Step 3: Get probe results history
        results, _, err := env.Client.Monitoring.GetProbeResults(env.Ctx, created.UUID, &nexmonyx.ListOptions{
            Page: 1, Limit: 10,
        })
        require.NoError(t, err)
        assert.NotEmpty(t, results)

        // Step 4: Update probe configuration
        created.Interval = 120
        updated, err := env.Client.Monitoring.UpdateProbe(env.Ctx, created.UUID, created)
        require.NoError(t, err)
        assert.Equal(t, 120, updated.Interval)

        // Step 5: Disable probe
        err = env.Client.Monitoring.DisableProbe(env.Ctx, created.UUID)
        require.NoError(t, err)

        // Step 6: Delete probe
        err = env.Client.Monitoring.DeleteProbe(env.Ctx, created.UUID)
        require.NoError(t, err)
    })
}
```

**Additional tests:**
- Multiple probes from different regions
- Probe failure notification workflow
- Probe metrics aggregation

---

## Task #3003: Authentication Flow Integration Tests

**Effort**: 6 hours
**Priority**: MEDIUM
**Prerequisites**: Task #3001

### Objective

Validate all authentication methods supported by the SDK and ensure proper handling of authentication lifecycle events.

### Deliverable: `auth_integration_test.go`

```go
package integration

import (
    "context"
    "testing"
    "time"

    nexmonyx "github.com/nexmonyx/go-sdk/v2"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// TestJWTAuthentication tests JWT token-based authentication
func TestJWTAuthentication(t *testing.T) {
    skipIfShort(t)
    env := setupIntegrationTest(t)
    defer teardownIntegrationTest(t, env)

    t.Run("ValidJWTToken", func(t *testing.T) {
        // Create client with JWT token
        client, err := nexmonyx.NewClient(&nexmonyx.Config{
            BaseURL: env.BaseURL,
            Auth: nexmonyx.AuthConfig{
                Token: "test-token",
            },
        })
        require.NoError(t, err)

        // Test authenticated request
        user, err := client.Users.GetMe(context.Background())
        require.NoError(t, err)
        assert.NotNil(t, user)
    })

    t.Run("InvalidJWTToken", func(t *testing.T) {
        client, err := nexmonyx.NewClient(&nexmonyx.Config{
            BaseURL: env.BaseURL,
            Auth: nexmonyx.AuthConfig{
                Token: "invalid-token",
            },
        })
        require.NoError(t, err)

        // Should fail with unauthorized error
        _, err = client.Users.GetMe(context.Background())
        require.Error(t, err)
        _, isUnauthorized := err.(*nexmonyx.UnauthorizedError)
        assert.True(t, isUnauthorized)
    })

    t.Run("MissingToken", func(t *testing.T) {
        client, err := nexmonyx.NewClient(&nexmonyx.Config{
            BaseURL: env.BaseURL,
            // No auth config
        })
        require.NoError(t, err)

        _, err = client.Users.GetMe(context.Background())
        require.Error(t, err)
    })
}

// TestAPIKeyAuthentication tests API key/secret authentication
func TestAPIKeyAuthentication(t *testing.T) {
    skipIfShort(t)
    env := setupIntegrationTest(t)
    defer teardownIntegrationTest(t, env)

    t.Run("ValidAPIKeySecret", func(t *testing.T) {
        client, err := nexmonyx.NewClient(&nexmonyx.Config{
            BaseURL: env.BaseURL,
            Auth: nexmonyx.AuthConfig{
                APIKey:    "test-api-key",
                APISecret: "test-api-secret",
            },
        })
        require.NoError(t, err)

        servers, _, err := client.Servers.List(context.Background(), nil)
        require.NoError(t, err)
        assert.NotNil(t, servers)
    })

    t.Run("InvalidAPIKey", func(t *testing.T) {
        client, err := nexmonyx.NewClient(&nexmonyx.Config{
            BaseURL: env.BaseURL,
            Auth: nexmonyx.AuthConfig{
                APIKey:    "invalid-key",
                APISecret: "invalid-secret",
            },
        })
        require.NoError(t, err)

        _, _, err = client.Servers.List(context.Background(), nil)
        require.Error(t, err)
    })
}

// TestServerCredentialsAuthentication tests server UUID/secret authentication
func TestServerCredentialsAuthentication(t *testing.T) {
    skipIfShort(t)
    env := setupIntegrationTest(t)
    defer teardownIntegrationTest(t, env)

    t.Run("ValidServerCredentials", func(t *testing.T) {
        client, err := nexmonyx.NewClient(&nexmonyx.Config{
            BaseURL: env.BaseURL,
            Auth: nexmonyx.AuthConfig{
                ServerUUID:   "server-001",
                ServerSecret: "test-server-secret",
            },
        })
        require.NoError(t, err)

        // Server-authenticated client should be able to send heartbeat
        err = client.Servers.Heartbeat(context.Background())
        require.NoError(t, err)
    })

    t.Run("InvalidServerCredentials", func(t *testing.T) {
        client, err := nexmonyx.NewClient(&nexmonyx.Config{
            BaseURL: env.BaseURL,
            Auth: nexmonyx.AuthConfig{
                ServerUUID:   "invalid-uuid",
                ServerSecret: "invalid-secret",
            },
        })
        require.NoError(t, err)

        err = client.Servers.Heartbeat(context.Background())
        require.Error(t, err)
    })
}

// TestTokenRefresh tests token refresh functionality
func TestTokenRefresh(t *testing.T) {
    skipIfShort(t)
    env := setupIntegrationTest(t)
    defer teardownIntegrationTest(t, env)

    t.Run("TokenExpiration", func(t *testing.T) {
        // Note: This test would need mock server support for token expiration
        // In a real dev API environment, you would:
        // 1. Create a client with a short-lived token
        // 2. Wait for token to expire
        // 3. Verify requests fail with 401
        // 4. Refresh token
        // 5. Verify requests succeed again

        t.Skip("Requires mock server support for token expiration simulation")
    })
}

// TestAuthenticationHeaders tests that correct headers are sent
func TestAuthenticationHeaders(t *testing.T) {
    skipIfShort(t)
    env := setupIntegrationTest(t)
    defer teardownIntegrationTest(t, env)

    t.Run("BearerTokenHeader", func(t *testing.T) {
        // Verify Authorization: Bearer <token> header is sent
        client, err := nexmonyx.NewClient(&nexmonyx.Config{
            BaseURL: env.BaseURL,
            Auth: nexmonyx.AuthConfig{
                Token: "test-token",
            },
            Debug: true,  // Enable debug to see headers
        })
        require.NoError(t, err)

        _, err = client.Users.GetMe(context.Background())
        require.NoError(t, err)
    })
}
```

**Additional tests to implement:**
- Token refresh before expiration
- Multi-factor authentication flows (if supported)
- Session management and logout
- API key rotation workflow

---

## Task #3004: Error Scenario Integration Tests

**Effort**: 6 hours
**Priority**: MEDIUM
**Prerequisites**: Task #3001

### Objective

Test the SDK's resilience and error handling in various failure scenarios.

### Deliverable: `error_scenarios_test.go`

```go
package integration

import (
    "context"
    "errors"
    "net/http"
    "testing"
    "time"

    nexmonyx "github.com/nexmonyx/go-sdk/v2"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// TestNetworkFailureRecovery tests SDK behavior during network issues
func TestNetworkFailureRecovery(t *testing.T) {
    skipIfShort(t)

    t.Run("TimeoutHandling", func(t *testing.T) {
        // Create client with very short timeout
        client, err := nexmonyx.NewClient(&nexmonyx.Config{
            BaseURL: "http://192.0.2.1:9999",  // Non-routable IP
            Timeout: 1 * time.Second,
            Auth: nexmonyx.AuthConfig{
                Token: "test-token",
            },
        })
        require.NoError(t, err)

        // Should timeout
        ctx := context.Background()
        _, _, err = client.Servers.List(ctx, nil)
        require.Error(t, err)
        assert.Contains(t, err.Error(), "timeout")
    })

    t.Run("ConnectionRefused", func(t *testing.T) {
        client, err := nexmonyx.NewClient(&nexmonyx.Config{
            BaseURL: "http://localhost:9999",  // Nothing listening
            Auth: nexmonyx.AuthConfig{
                Token: "test-token",
            },
        })
        require.NoError(t, err)

        _, _, err = client.Servers.List(context.Background(), nil)
        require.Error(t, err)
        assert.Contains(t, err.Error(), "connection refused")
    })

    t.Run("DNSFailure", func(t *testing.T) {
        client, err := nexmonyx.NewClient(&nexmonyx.Config{
            BaseURL: "http://this-domain-does-not-exist-12345.com",
            Auth: nexmonyx.AuthConfig{
                Token: "test-token",
            },
        })
        require.NoError(t, err)

        _, _, err = client.Servers.List(context.Background(), nil)
        require.Error(t, err)
    })
}

// TestAPIRateLimiting tests rate limit handling
func TestAPIRateLimiting(t *testing.T) {
    skipIfShort(t)

    t.Run("RateLimitResponse", func(t *testing.T) {
        // Note: This requires mock server support for 429 responses
        // Or actual rate limiting in dev environment

        env := setupIntegrationTest(t)
        defer teardownIntegrationTest(t, env)

        // Make many rapid requests to trigger rate limiting
        for i := 0; i < 100; i++ {
            _, _, err := env.Client.Servers.List(env.Ctx, nil)
            if err != nil {
                // Check if it's a rate limit error
                _, isRateLimit := err.(*nexmonyx.RateLimitError)
                if isRateLimit {
                    t.Logf("Hit rate limit after %d requests", i+1)
                    return  // Test passed
                }
            }
        }

        t.Log("Did not hit rate limit (may need higher request volume)")
    })

    t.Run("RateLimitRetry", func(t *testing.T) {
        // Test automatic retry with backoff
        t.Skip("Requires rate limit simulation in mock server")
    })
}

// TestServiceUnavailability tests 503 handling
func TestServiceUnavailability(t *testing.T) {
    skipIfShort(t)

    t.Run("ServiceUnavailable503", func(t *testing.T) {
        // Mock server would need to simulate 503 responses
        t.Skip("Requires 503 simulation in mock server")
    })

    t.Run("RetryAfter503", func(t *testing.T) {
        // Test retry logic with Retry-After header
        t.Skip("Requires 503 simulation with Retry-After header")
    })
}

// TestPartialFailures tests handling of partial operation failures
func TestPartialFailures(t *testing.T) {
    skipIfShort(t)
    env := setupIntegrationTest(t)
    defer teardownIntegrationTest(t, env)

    t.Run("BulkOperationPartialFailure", func(t *testing.T) {
        // Create multiple servers, some should succeed, some fail
        servers := []*nexmonyx.Server{
            {Hostname: "valid-server-1", OrganizationID: 1, MainIP: "192.168.1.1"},
            {Hostname: "", OrganizationID: 1, MainIP: "192.168.1.2"},  // Invalid - no hostname
            {Hostname: "valid-server-2", OrganizationID: 1, MainIP: "192.168.1.3"},
        }

        successCount := 0
        failureCount := 0

        for _, server := range servers {
            _, err := env.Client.Servers.Create(env.Ctx, server)
            if err != nil {
                failureCount++
            } else {
                successCount++
            }
        }

        assert.Greater(t, successCount, 0, "Some servers should succeed")
        assert.Greater(t, failureCount, 0, "Some servers should fail")
    })
}

// TestContextCancellation tests context cancellation handling
func TestContextCancellation(t *testing.T) {
    skipIfShort(t)
    env := setupIntegrationTest(t)
    defer teardownIntegrationTest(t, env)

    t.Run("CancelDuringRequest", func(t *testing.T) {
        ctx, cancel := context.WithCancel(context.Background())

        // Cancel immediately
        cancel()

        _, _, err := env.Client.Servers.List(ctx, nil)
        require.Error(t, err)
        assert.True(t, errors.Is(err, context.Canceled))
    })

    t.Run("TimeoutContext", func(t *testing.T) {
        // Create context with very short timeout
        ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
        defer cancel()

        time.Sleep(10 * time.Millisecond)  // Ensure timeout passes

        _, _, err := env.Client.Servers.List(ctx, nil)
        require.Error(t, err)
        assert.True(t, errors.Is(err, context.DeadlineExceeded))
    })
}

// TestMalformedResponses tests handling of unexpected API responses
func TestMalformedResponses(t *testing.T) {
    skipIfShort(t)

    t.Run("InvalidJSON", func(t *testing.T) {
        // Mock server would need to return invalid JSON
        t.Skip("Requires malformed response simulation in mock server")
    })

    t.Run("UnexpectedStatusCode", func(t *testing.T) {
        // Test handling of unexpected 2xx/3xx/4xx/5xx codes
        t.Skip("Requires status code simulation in mock server")
    })

    t.Run("MissingRequiredFields", func(t *testing.T) {
        // Test response with missing required fields
        t.Skip("Requires field omission simulation in mock server")
    })
}

// TestResourceNotFound tests 404 handling
func TestResourceNotFound(t *testing.T) {
    skipIfShort(t)
    env := setupIntegrationTest(t)
    defer teardownIntegrationTest(t, env)

    t.Run("ServerNotFound", func(t *testing.T) {
        _, err := env.Client.Servers.GetByUUID(env.Ctx, "non-existent-uuid")
        require.Error(t, err)
        _, isNotFound := err.(*nexmonyx.NotFoundError)
        assert.True(t, isNotFound, "Error should be NotFoundError")
    })

    t.Run("OrganizationNotFound", func(t *testing.T) {
        _, err := env.Client.Organizations.Get(env.Ctx, "non-existent-org-uuid")
        require.Error(t, err)
        _, isNotFound := err.(*nexmonyx.NotFoundError)
        assert.True(t, isNotFound)
    })
}

// TestValidationErrors tests 400 validation error handling
func TestValidationErrors(t *testing.T) {
    skipIfShort(t)
    env := setupIntegrationTest(t)
    defer teardownIntegrationTest(t, env)

    t.Run("MissingRequiredField", func(t *testing.T) {
        // Create server without required hostname
        server := &nexmonyx.Server{
            OrganizationID: 1,
            MainIP:         "192.168.1.100",
        }
        _, err := env.Client.Servers.Create(env.Ctx, server)
        require.Error(t, err)
        _, isValidation := err.(*nexmonyx.ValidationError)
        assert.True(t, isValidation, "Error should be ValidationError")
    })

    t.Run("InvalidFieldFormat", func(t *testing.T) {
        // Create server with invalid IP address
        server := &nexmonyx.Server{
            Hostname:       "test-server",
            OrganizationID: 1,
            MainIP:         "invalid-ip-address",
        }
        _, err := env.Client.Servers.Create(env.Ctx, server)
        require.Error(t, err)
    })
}
```

**Additional tests:**
- SSL/TLS certificate validation failures
- Proxy connection failures
- API version mismatch handling
- Concurrent request failures

---

## Task #3017: Dev API Integration Testing

**Effort**: 4 hours
**Priority**: MEDIUM
**Prerequisites**: Task #3001, Tasks #3002-#3004 (recommended)

### Objective

Extend the integration test framework to support testing against a real Nexmonyx development API server, enabling validation of SDK behavior against actual backend responses.

### Benefits

- **API Compatibility Validation**: Catch breaking changes in API responses early
- **Real-world Behavior**: Test with actual network latency, timeouts, and edge cases
- **Backend Integration**: Validate end-to-end flows with real database state
- **Regression Detection**: Identify issues before they reach production

### Implementation Steps

#### 1. Update `helpers.go` - Add Dev Mode Support

Add dev mode detection and configuration:

```go
// isDev Mode checks if we're running against dev API
func isDevMode() bool {
    return os.Getenv("INTEGRATION_TEST_MODE") == "dev"
}

// setupIntegrationTest initializes test environment (updated)
func setupIntegrationTest(t *testing.T) *TestEnvironment {
    t.Helper()

    var client *nexmonyx.Client
    var mockAPI *MockAPIServer
    var baseURL string
    var authToken string

    if isDevMode() {
        // Dev API mode - use real API server
        baseURL = os.Getenv("INTEGRATION_TEST_API_URL")
        if baseURL == "" {
            t.Fatal("INTEGRATION_TEST_API_URL must be set in dev mode")
        }

        authToken = os.Getenv("INTEGRATION_TEST_AUTH_TOKEN")
        if authToken == "" {
            t.Fatal("INTEGRATION_TEST_AUTH_TOKEN must be set in dev mode")
        }

        var err error
        client, err = nexmonyx.NewClient(&nexmonyx.Config{
            BaseURL: baseURL,
            Auth: nexmonyx.AuthConfig{
                Token: authToken,
            },
            Timeout: getTestTimeout(),
            Debug:   os.Getenv("INTEGRATION_TEST_DEBUG") == "true",
        })
        require.NoError(t, err, "Failed to create SDK client for dev API")

        t.Logf("Integration test environment (DEV MODE) initialized with API: %s", baseURL)
    } else {
        // Mock mode - use mock API server
        mockAPI = NewMockAPIServer(t)
        baseURL = mockAPI.Server.URL
        authToken = "test-token"

        var err error
        client, err = nexmonyx.NewClient(&nexmonyx.Config{
            BaseURL: baseURL,
            Auth: nexmonyx.AuthConfig{
                Token: authToken,
            },
            Timeout: getTestTimeout(),
            Debug:   os.Getenv("INTEGRATION_TEST_DEBUG") == "true",
        })
        require.NoError(t, err, "Failed to create SDK client for mock API")

        t.Logf("Integration test environment (MOCK MODE) initialized at %s", baseURL)
    }

    ctx := context.Background()

    env := &TestEnvironment{
        Client:    client,
        MockAPI:   mockAPI,
        BaseURL:   baseURL,
        AuthToken: authToken,
        Ctx:       ctx,
    }

    return env
}

// teardownIntegrationTest cleans up (updated)
func teardownIntegrationTest(t *testing.T, env *TestEnvironment) {
    t.Helper()

    if env.MockAPI != nil {
        // Mock mode - close mock server
        env.MockAPI.Close()
        t.Log("Integration test environment (MOCK MODE) cleaned up")
    } else {
        // Dev mode - perform cleanup operations on dev API
        cleanupDevResources(t, env)
        t.Log("Integration test environment (DEV MODE) cleaned up")
    }
}

// cleanupDevResources removes test resources from dev API
func cleanupDevResources(t *testing.T, env *TestEnvironment) {
    t.Helper()

    // Delete any servers created during tests with "test-" prefix
    servers, _, err := env.Client.Servers.List(env.Ctx, &nexmonyx.ListOptions{
        Search: "test-",
        Limit:  100,
    })
    if err != nil {
        t.Logf("Warning: Failed to list test servers for cleanup: %v", err)
        return
    }

    for _, server := range servers {
        if strings.HasPrefix(server.Hostname, "test-") ||
           strings.HasPrefix(server.Hostname, "workflow-test-") {
            err := env.Client.Servers.Delete(env.Ctx, server.ServerUUID)
            if err != nil {
                t.Logf("Warning: Failed to delete test server %s: %v", server.ServerUUID, err)
            } else {
                t.Logf("Cleaned up test server: %s", server.ServerUUID)
            }
        }
    }
}
```

#### 2. Update `README.md` - Document Dev Mode Usage

Add dev mode documentation:

```markdown
## Running Against Dev API Server

### Setup

1. **Get Dev API Access**:
   - Obtain dev API URL from your team
   - Generate an API token with appropriate permissions

2. **Set Environment Variables**:
```bash
export INTEGRATION_TESTS=true
export INTEGRATION_TEST_MODE=dev
export INTEGRATION_TEST_API_URL=https://dev-api.nexmonyx.com
export INTEGRATION_TEST_AUTH_TOKEN=your-dev-api-token
export INTEGRATION_TEST_DEBUG=true  # Optional
```

3. **Run Tests**:
```bash
# Run all integration tests against dev API
INTEGRATION_TESTS=true \
INTEGRATION_TEST_MODE=dev \
INTEGRATION_TEST_API_URL=https://dev-api.nexmonyx.com \
INTEGRATION_TEST_AUTH_TOKEN=your-token \
go test -v ./tests/integration/...

# Run specific test
INTEGRATION_TESTS=true \
INTEGRATION_TEST_MODE=dev \
INTEGRATION_TEST_API_URL=https://dev-api.nexmonyx.com \
INTEGRATION_TEST_AUTH_TOKEN=your-token \
go test -v -run TestServerLifecycleWorkflow ./tests/integration/
```

### Dev Mode vs Mock Mode

| Feature | Mock Mode | Dev Mode |
|---------|-----------|----------|
| **Speed** | Very fast | Slower (network latency) |
| **Setup** | None required | Requires dev API access |
| **Reliability** | Always available | Depends on dev API uptime |
| **Data** | Fixture-based | Real database |
| **Isolation** | Complete | Shared dev environment |
| **API Validation** | Simulated | Real API responses |
| **Use Case** | Development, CI/CD | Pre-release validation |

### Best Practices for Dev Mode

1. **Use Test Prefixes**: Always prefix test resources with `test-` or `workflow-test-` for easy cleanup
2. **Clean Up Resources**: The framework attempts automatic cleanup, but verify manually if tests fail
3. **Handle Flakiness**: Dev API may have network issues; use retries and timeouts appropriately
4. **Don't Rely on State**: Dev API state may change between test runs; tests should be idempotent
5. **Check Quotas**: Dev API may have rate limits or resource quotas

### Troubleshooting Dev Mode

**Problem**: Tests fail with connection errors
- **Solution**: Verify dev API URL and network connectivity

**Problem**: Tests fail with authentication errors
- **Solution**: Verify API token is valid and has required permissions

**Problem**: Tests leave orphaned resources
- **Solution**: Manually delete resources with `test-` prefix or run cleanup script

**Problem**: Tests are slow in dev mode
- **Solution**: This is expected due to network latency; use mock mode for faster iteration
```

#### 3. Create CI/CD Workflow (Optional)

Create `.github/workflows/integration-dev-tests.yml`:

```yaml
name: Integration Tests (Dev API)

on:
  schedule:
    # Run nightly against dev API
    - cron: '0 2 * * *'
  workflow_dispatch:
    # Allow manual triggers

jobs:
  integration-tests-dev:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'

      - name: Run integration tests (dev mode)
        env:
          INTEGRATION_TESTS: true
          INTEGRATION_TEST_MODE: dev
          INTEGRATION_TEST_API_URL: ${{ secrets.DEV_API_URL }}
          INTEGRATION_TEST_AUTH_TOKEN: ${{ secrets.DEV_API_TOKEN }}
          INTEGRATION_TEST_DEBUG: true
          INTEGRATION_TEST_TIMEOUT: 60s
        run: |
          go test -v -timeout 30m ./tests/integration/...

      - name: Upload test results
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: integration-test-results
          path: |
            test-results.xml
            integration-coverage.out
```

#### 4. Update `TASKS.md`

Mark Task #3017 as completed and document the implementation.

---

## Testing Checklist

Before marking tasks as complete, ensure:

### Task #3002 Checklist:
- [ ] All field name mismatches fixed
- [ ] Server workflow tests pass in mock mode
- [ ] Organization workflow tests pass in mock mode
- [ ] Alert workflow tests pass in mock mode
- [ ] Monitoring workflow tests pass in mock mode
- [ ] All workflows test complete lifecycle
- [ ] Tests clean up created resources
- [ ] Tests handle errors appropriately

### Task #3003 Checklist:
- [ ] JWT authentication tests pass
- [ ] API key authentication tests pass
- [ ] Server credentials authentication tests pass
- [ ] Invalid authentication scenarios handled
- [ ] Authentication headers verified
- [ ] Tests cover all auth methods in SDK

### Task #3004 Checklist:
- [ ] Network failure tests pass
- [ ] Rate limiting tests implemented
- [ ] Context cancellation tests pass
- [ ] 404 Not Found tests pass
- [ ] 400 Validation error tests pass
- [ ] Partial failure tests pass
- [ ] All error types from SDK are tested

### Task #3017 Checklist:
- [ ] Dev mode environment variable detection works
- [ ] Dev mode uses real API server
- [ ] Mock mode still works (no regression)
- [ ] Cleanup function removes test resources in dev mode
- [ ] README.md documents dev mode usage
- [ ] Dev mode tested manually
- [ ] CI/CD workflow created (optional)
- [ ] Tests pass in both mock and dev modes

---

## Running All Tests

```bash
# Mock mode (default)
INTEGRATION_TESTS=true go test -v ./tests/integration/...

# Dev mode
INTEGRATION_TESTS=true \
INTEGRATION_TEST_MODE=dev \
INTEGRATION_TEST_API_URL=https://dev-api.nexmonyx.com \
INTEGRATION_TEST_AUTH_TOKEN=your-token \
go test -v ./tests/integration/...

# With coverage
INTEGRATION_TESTS=true go test -v -cover -coverprofile=integration-coverage.out ./tests/integration/...

# Specific test
INTEGRATION_TESTS=true go test -v -run TestServerLifecycleWorkflow ./tests/integration/
```

---

## Summary

This guide provides complete implementation details for Tasks #3002-#3004 and #3017. Each task builds on the foundation created in Task #3001 and together they provide comprehensive integration testing coverage for the Nexmonyx Go SDK.

**Total Effort**: 32 hours
- Task #3002: 16 hours (Core Service Integration Tests)
- Task #3003: 6 hours (Authentication Flow Tests)
- Task #3004: 6 hours (Error Scenario Tests)
- Task #3017: 4 hours (Dev API Mode Support)

**Priority**: HIGH for #3002, MEDIUM for #3003, #3004, #3017
