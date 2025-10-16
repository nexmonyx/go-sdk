# Integration Testing Framework

This directory contains the integration testing framework for the Nexmonyx Go SDK. Integration tests validate the complete SDK functionality against a mock API server that mimics the real Nexmonyx API.

## Overview

The integration testing framework provides:

- **Mock API Server**: Full HTTP server implementation that mimics Nexmonyx API endpoints
- **Test Fixtures**: Realistic sample data for servers, organizations, users, metrics, and alerts
- **Test Helpers**: Utilities for setup, teardown, assertions, and common operations
- **Example Tests**: Comprehensive test suite demonstrating integration testing patterns

## Directory Structure

```
tests/integration/
├── README.md                  # This file
├── integration_test.go        # TestMain entry point
├── mock_api_server.go         # Mock API server implementation
├── helpers.go                 # Test helper functions
├── servers_test.go            # Example: Server integration tests
└── fixtures/                  # Test data fixtures
    ├── servers.json
    ├── organizations.json
    ├── users.json
    ├── metrics.json
    └── alerts.json
```

## Running Integration Tests

### Basic Usage

Integration tests are **disabled by default** to avoid running in CI/CD without explicit configuration.

To run integration tests:

```bash
# Run all integration tests
INTEGRATION_TESTS=true go test -v ./tests/integration/...

# Run with short mode (skips long-running tests)
INTEGRATION_TESTS=true go test -short -v ./tests/integration/...

# Run specific test
INTEGRATION_TESTS=true go test -v -run TestServersIntegration ./tests/integration/
```

### Advanced Options

```bash
# Enable debug logging
INTEGRATION_TESTS=true INTEGRATION_TEST_DEBUG=true go test -v ./tests/integration/...

# Set custom timeout
INTEGRATION_TESTS=true INTEGRATION_TEST_TIMEOUT=60s go test -v ./tests/integration/...

# Run with race detector
INTEGRATION_TESTS=true go test -race -v ./tests/integration/...

# Run with coverage
INTEGRATION_TESTS=true go test -cover -coverprofile=integration-coverage.out ./tests/integration/...
```

## Writing Integration Tests

### Test Structure

Every integration test should follow this pattern:

```go
func TestMyFeature(t *testing.T) {
    skipIfShort(t)  // Skip in short mode

    // Setup test environment
    env := setupIntegrationTest(t)
    defer teardownIntegrationTest(t, env)

    // Run test logic
    t.Run("SubTest", func(t *testing.T) {
        result, err := env.Client.SomeService.SomeMethod(env.Ctx, ...)
        require.NoError(t, err)
        assert.NotNil(t, result)
        // ... more assertions
    })
}
```

### Available Helpers

#### Setup and Teardown

- `setupIntegrationTest(t)` - Creates test environment with mock API server
- `teardownIntegrationTest(t, env)` - Cleans up test environment
- `skipIfShort(t)` - Skips test if running in short mode

#### Test Data Creation

- `createTestServer(t, env, hostname)` - Creates a test server via API
- `createTestOrganization(t, env, name)` - Gets a test organization
- `createTestMetricsPayload(serverUUID)` - Creates sample metrics data

#### Assertions

- `assertServerEqual(t, expected, actual)` - Compares two servers
- `assertOrganizationEqual(t, expected, actual)` - Compares two organizations
- `assertAlertEqual(t, expected, actual)` - Compares two alerts
- `assertValidTimestamp(t, timestamp, fieldName)` - Validates timestamp
- `assertValidUUID(t, uuid, fieldName)` - Validates UUID format
- `assertPaginationValid(t, meta)` - Validates pagination metadata

#### Utilities

- `waitForCondition(t, condition, timeout, message)` - Waits for condition with timeout
- `retryOperation(t, operation, maxRetries, initialDelay)` - Retries operation with backoff
- `loadFixture(t, filename)` - Loads JSON fixture file
- `getTestTimeout()` - Returns configured test timeout

### Example Test

See `servers_test.go` for a comprehensive example covering:

- List servers with pagination
- Get server by UUID
- Create new server
- Update existing server
- Delete server
- Search servers
- Authentication testing
- Error handling

## Mock API Server

The mock API server (`mock_api_server.go`) provides a complete simulation of the Nexmonyx API:

### Features

- **Stateful**: Tracks created/updated/deleted resources during tests
- **Thread-safe**: Uses `sync.RWMutex` for concurrent access
- **Authentication**: Validates Bearer tokens (default: "test-token")
- **RESTful**: Implements standard REST patterns (GET, POST, PUT, DELETE)
- **Fixtures**: Loads realistic test data on startup
- **Error Handling**: Returns appropriate HTTP status codes

### Supported Endpoints

| Endpoint | Methods | Description |
|----------|---------|-------------|
| `/v2/servers` | GET, POST | List/create servers |
| `/v2/servers/{uuid}` | GET, PUT, DELETE | Get/update/delete server |
| `/v2/organizations` | GET | List organizations |
| `/v2/organizations/{uuid}` | GET | Get organization |
| `/v2/metrics/submit` | POST | Submit metrics |
| `/v2/metrics` | GET | Query metrics |
| `/v2/alerts` | GET | List alerts |
| `/v2/alerts/{uuid}` | GET | Get alert |
| `/v2/monitoring/probes` | GET | List monitoring probes |
| `/v2/system/health` | GET | Health check |
| `/v2/system/version` | GET | API version |

### Customization

```go
// Create mock server with custom configuration
mock := NewMockAPIServer(t)

// Disable authentication for specific tests
mock.DisableAuth()

// Set custom auth token
mock.SetAuthToken("custom-token")

// Access mock state for verification
mock.mu.RLock()
serverCount := len(mock.servers)
mock.mu.RUnlock()
```

## Test Fixtures

Fixtures are realistic sample data stored in `fixtures/` directory:

### servers.json
5 sample servers covering different environments (production, staging, development) and classifications (web, database, monitoring, worker).

### organizations.json
3 organizations with different subscription tiers (basic, professional, enterprise).

### users.json
3 users with different roles (admin, member, viewer).

### metrics.json
Sample metrics data including:
- Comprehensive metrics (CPU, memory, disk, network)
- Simple metrics (basic percentages)
- Historical time-series data

### alerts.json
3 alert configurations for different metric types (cpu, disk, memory) with varying severity levels.

## Best Practices

### 1. Use Subtests

Organize related tests with `t.Run()`:

```go
func TestServers(t *testing.T) {
    env := setupIntegrationTest(t)
    defer teardownIntegrationTest(t, env)

    t.Run("Create", func(t *testing.T) { /* ... */ })
    t.Run("Update", func(t *testing.T) { /* ... */ })
    t.Run("Delete", func(t *testing.T) { /* ... */ })
}
```

### 2. Clean Up After Tests

Always use `defer` to ensure cleanup:

```go
env := setupIntegrationTest(t)
defer teardownIntegrationTest(t, env)
```

### 3. Use Helper Functions

Prefer helper functions over repeating logic:

```go
// Good
server := createTestServer(t, env, "test-server")

// Less good
server, err := env.Client.Servers.Create(env.Ctx, &nexmonyx.ServerCreateRequest{...})
require.NoError(t, err)
```

### 4. Test Error Cases

Don't just test the happy path:

```go
t.Run("NotFound", func(t *testing.T) {
    _, err := env.Client.Servers.Get(env.Ctx, "non-existent")
    require.Error(t, err)
    _, isNotFound := err.(*nexmonyx.NotFoundError)
    assert.True(t, isNotFound)
})
```

### 5. Use Meaningful Assertions

Provide helpful assertion messages:

```go
// Good
require.NoError(t, err, "Failed to create server")

// Less good
require.NoError(t, err)
```

### 6. Skip Long Tests in Short Mode

Use `skipIfShort(t)` for tests that take significant time:

```go
func TestLongRunningOperation(t *testing.T) {
    skipIfShort(t)
    // ... test logic
}
```

## Running Against Dev API Server

The framework supports testing against both a **mock API server** (default) and a **real Nexmonyx development API server** for comprehensive validation.

### Setup for Dev Mode

#### 1. Get Dev API Access

- Obtain dev API URL from your team (e.g., `https://dev-api.nexmonyx.com`)
- Generate an API token with appropriate permissions for your user account

#### 2. Set Environment Variables

```bash
export INTEGRATION_TESTS=true
export INTEGRATION_TEST_MODE=dev
export INTEGRATION_TEST_API_URL=https://dev-api.nexmonyx.com
export INTEGRATION_TEST_AUTH_TOKEN=your-dev-api-token
export INTEGRATION_TEST_DEBUG=true  # Optional: Enable debug logging
export INTEGRATION_TEST_TIMEOUT=60s  # Optional: Custom timeout
```

#### 3. Run Tests Against Dev API

```bash
# Run all integration tests against dev API
INTEGRATION_TESTS=true \
INTEGRATION_TEST_MODE=dev \
INTEGRATION_TEST_API_URL=https://dev-api.nexmonyx.com \
INTEGRATION_TEST_AUTH_TOKEN=your-token \
go test -v ./tests/integration/...

# Run specific test against dev API
INTEGRATION_TESTS=true \
INTEGRATION_TEST_MODE=dev \
INTEGRATION_TEST_API_URL=https://dev-api.nexmonyx.com \
INTEGRATION_TEST_AUTH_TOKEN=your-token \
go test -v -run TestServerLifecycleWorkflow ./tests/integration/

# Run with coverage
INTEGRATION_TESTS=true \
INTEGRATION_TEST_MODE=dev \
INTEGRATION_TEST_API_URL=https://dev-api.nexmonyx.com \
INTEGRATION_TEST_AUTH_TOKEN=your-token \
go test -v -cover -coverprofile=integration-dev-coverage.out ./tests/integration/...
```

### Dev Mode vs Mock Mode

| Feature | Mock Mode | Dev Mode |
|---------|-----------|----------|
| **Speed** | Very fast (~2-5s per test) | Slower (~10-30s per test, network latency) |
| **Setup** | None required | Requires dev API access and credentials |
| **Reliability** | Always available | Depends on dev API uptime |
| **Data** | Fixture-based, predictable | Real database, may have existing data |
| **Isolation** | Complete test isolation | Shared dev environment (use test prefixes!) |
| **API Validation** | Simulated API responses | Real API responses from actual backend |
| **Use Case** | Development, CI/CD, fast iteration | Pre-release validation, API compatibility checks |
| **Cleanup** | Automatic (mock server shutdown) | Best-effort automatic cleanup |

### Best Practices for Dev Mode

#### 1. Use Test Prefixes

Always prefix test resource names with `test-`, `workflow-test-`, or `alert-test-` for easy identification and cleanup:

```go
server := &nexmonyx.Server{
    Hostname:       "test-my-feature-server",  // ✅ Good
    OrganizationID: 1,
    // ...
}

// ❌ Bad - no test prefix
server := &nexmonyx.Server{
    Hostname:       "production-web-01",  // Don't do this in tests!
    // ...
}
```

#### 2. Clean Up Resources

The framework attempts automatic cleanup, but you should:

- Verify tests clean up properly by checking dev API after test runs
- Manually delete orphaned resources if tests fail mid-execution
- Use the dev API UI or SDK to check for leftover `test-*` resources

#### 3. Handle Flakiness

Dev API may have network issues or rate limiting:

```go
// Use retry helpers for flaky operations
err := retryOperation(t, func() error {
    return env.Client.Servers.Create(env.Ctx, server)
}, 3, 1*time.Second)
```

#### 4. Don't Rely on State

Dev API state may change between test runs:

```go
// ✅ Good - create your own test data
server := createTestServer(t, env, "test-my-server")

// ❌ Bad - assuming a specific server exists
server, err := env.Client.Servers.GetByUUID(env.Ctx, "server-001")
```

#### 5. Check Quotas and Limits

Dev API may have:
- Rate limits (e.g., 100 requests/minute)
- Resource quotas (e.g., max 50 servers per organization)
- Test data retention policies

### Troubleshooting Dev Mode

#### Problem: Tests fail with connection errors

```
Error: Failed to create SDK client for dev API: dial tcp: lookup dev-api.nexmonyx.com: no such host
```

**Solution**: Verify dev API URL is correct and accessible from your network:
```bash
curl -I https://dev-api.nexmonyx.com/v2/system/health
```

#### Problem: Tests fail with authentication errors

```
Error: 401 Unauthorized: Invalid or expired token
```

**Solutions**:
1. Verify your API token is valid:
   ```bash
   curl -H "Authorization: Bearer your-token" \
        https://dev-api.nexmonyx.com/v2/users/me
   ```
2. Check token hasn't expired
3. Ensure token has required permissions (read/write access to servers, metrics, etc.)

#### Problem: Tests leave orphaned resources

```
Warning: Found 15 test servers in dev API after test run
```

**Solution**: Manually clean up using SDK or dev API UI:
```bash
# List all test servers
go run scripts/cleanup_test_resources.go --api-url=https://dev-api.nexmonyx.com \
    --token=your-token --prefix="test-"
```

Or use the SDK directly:
```go
// cleanup_script.go
servers, _, _ := client.Servers.List(ctx, &nexmonyx.ListOptions{Search: "test-"})
for _, server := range servers {
    client.Servers.Delete(ctx, server.ServerUUID)
}
```

#### Problem: Tests are very slow in dev mode

```
PASS: TestServerLifecycleWorkflow (45.23s)
```

**Solution**: This is expected due to network latency and real database operations. For fast iteration:
1. Use **mock mode** during development
2. Use **dev mode** only for pre-release validation or specific compatibility testing
3. Run specific tests instead of the full suite: `go test -run TestSpecificTest`

#### Problem: Rate limiting in dev mode

```
Error: 429 Too Many Requests: Rate limit exceeded
```

**Solution**:
1. Add delays between operations:
   ```go
   time.Sleep(100 * time.Millisecond)  // Throttle requests
   ```
2. Use `-p 1` flag to run tests sequentially:
   ```bash
   go test -p 1 -v ./tests/integration/...
   ```
3. Contact dev team to increase rate limits for your test account

### When to Use Each Mode

#### Use Mock Mode When:
- ✅ Developing new SDK features
- ✅ Writing new integration tests
- ✅ Running tests in CI/CD pipelines
- ✅ Fast iteration and debugging
- ✅ Testing error handling and edge cases

#### Use Dev Mode When:
- ✅ Validating SDK against real API (pre-release)
- ✅ Catching breaking API changes
- ✅ Testing API compatibility after backend updates
- ✅ Verifying end-to-end workflows with actual backend
- ✅ Investigating issues that only occur with real API

### Environment Variables Reference

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `INTEGRATION_TESTS` | Yes | `false` | Must be `true` to run integration tests |
| `INTEGRATION_TEST_MODE` | No | `mock` | Set to `dev` for dev API mode |
| `INTEGRATION_TEST_API_URL` | Dev mode only | N/A | Dev API base URL (e.g., `https://dev-api.nexmonyx.com`) |
| `INTEGRATION_TEST_AUTH_TOKEN` | Dev mode only | N/A | Your dev API JWT authentication token |
| `INTEGRATION_TEST_DEBUG` | No | `false` | Set to `true` for verbose HTTP logging |
| `INTEGRATION_TEST_TIMEOUT` | No | `30s` | Timeout for individual API calls |

## Troubleshooting

### Tests Not Running

**Problem**: Tests are skipped when running `go test`

**Solution**: Set `INTEGRATION_TESTS=true` environment variable:
```bash
INTEGRATION_TESTS=true go test -v ./tests/integration/...
```

### Authentication Errors

**Problem**: Getting 401 Unauthorized errors in tests

**Solution**: The mock server expects Bearer token "test-token" by default. Ensure your test client uses this token (handled by `setupIntegrationTest`).

### Port Conflicts

**Problem**: Mock server fails to start due to port conflicts

**Solution**: The mock server uses `httptest.NewServer()` which automatically finds an available port. If you still have issues, check for processes holding ports.

### Fixture Loading Errors

**Problem**: Tests fail with "fixture not found" errors

**Solution**: Ensure you're running tests from the repository root or the fixtures directory is in the correct location. The helper uses `runtime.Caller(0)` to find the fixtures directory relative to the test file.

## Contributing

When adding new integration tests:

1. Follow the existing test structure and patterns
2. Add appropriate fixtures if testing new resource types
3. Update this README if adding new helpers or patterns
4. Ensure tests clean up created resources
5. Test both success and error cases
6. Add meaningful test names and assertion messages

## References

- [Go Testing Package](https://pkg.go.dev/testing)
- [Testify Library](https://github.com/stretchr/testify)
- [httptest Package](https://pkg.go.dev/net/http/httptest)
- [Table-Driven Tests in Go](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
