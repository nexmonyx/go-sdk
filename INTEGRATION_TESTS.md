# Integration Tests

This document describes how to run the integration tests for the Nexmonyx Go SDK.

## Overview

Integration tests validate the SDK against a real Nexmonyx API instance. They test actual HTTP requests, authentication, error handling, and data flow.

**Important**: Integration tests are **NOT** run by default when you run `go test`. They must be explicitly enabled using build tags.

## Prerequisites

1. Access to a Nexmonyx API instance (production, staging, or local)
2. Valid authentication credentials:
   - JWT token for user operations
   - OR Server UUID + Secret for agent operations
3. Go 1.24 or later

## Environment Variables

### Required

- `NEXMONYX_API_URL`: The base URL of the Nexmonyx API
  - Example: `https://api.nexmonyx.com`
  - Example: `http://localhost:8080` (for local development)

- `NEXMONYX_AUTH_TOKEN`: A valid JWT token for authentication
  - Get this from the Nexmonyx web interface or API

### Optional

- `NEXMONYX_SERVER_UUID`: Server UUID for agent-specific tests
  - Only required for `TestIntegration_ServerAgent`

- `NEXMONYX_SERVER_SECRET`: Server secret for agent-specific tests
  - Only required for `TestIntegration_ServerAgent`

- `NEXMONYX_DEBUG`: Set to `true` to enable debug logging
  - Useful for troubleshooting API issues

## Running Integration Tests

### Run All Integration Tests

```bash
export NEXMONYX_API_URL="https://api.nexmonyx.com"
export NEXMONYX_AUTH_TOKEN="your-jwt-token-here"

go test -tags=integration -v -timeout 30m
```

### Run Specific Integration Test

```bash
export NEXMONYX_API_URL="https://api.nexmonyx.com"
export NEXMONYX_AUTH_TOKEN="your-jwt-token-here"

go test -tags=integration -run TestIntegration_Servers -v
```

### Run with Debug Logging

```bash
export NEXMONYX_API_URL="https://api.nexmonyx.com"
export NEXMONYX_AUTH_TOKEN="your-jwt-token-here"
export NEXMONYX_DEBUG="true"

go test -tags=integration -v
```

### Run Agent Tests

```bash
export NEXMONYX_API_URL="https://api.nexmonyx.com"
export NEXMONYX_AUTH_TOKEN="your-jwt-token-here"
export NEXMONYX_SERVER_UUID="your-server-uuid"
export NEXMONYX_SERVER_SECRET="your-server-secret"

go test -tags=integration -run TestIntegration_ServerAgent -v
```

## Test Coverage

The integration test suite covers the following areas:

### 1. User Operations (`TestIntegration_UserProfile`)
- Get current user profile
- Update user preferences

### 2. Organizations (`TestIntegration_Organizations`)
- List organizations
- Get organization details
- Pagination

### 3. Servers (`TestIntegration_Servers`)
- List servers
- Get server details
- Search servers
- Server updates

### 4. Server Agent (`TestIntegration_ServerAgent`)
- Submit heartbeats
- Submit metrics
- Agent authentication

### 5. Alerts (`TestIntegration_Alerts`)
- List alerts
- Get alert details
- Alert rules

### 6. Monitoring Probes (`TestIntegration_Probes`)
- List probes
- List regions
- Probe management

### 7. Metrics Query (`TestIntegration_MetricsQuery`)
- Query historical metrics
- Time-range queries
- Metric types

### 8. Pagination (`TestIntegration_Pagination`)
- Multi-page results
- Page navigation
- Metadata validation

### 9. Error Handling (`TestIntegration_ErrorHandling`)
- Not found errors (404)
- Validation errors (400)
- Invalid parameters

### 10. Rate Limiting (`TestIntegration_RateLimiting`)
- Rate limit detection
- Rate limit headers
- Retry behavior

### 11. Context Management (`TestIntegration_ContextCancellation`)
- Context cancellation
- Timeouts
- Deadline exceeded

## Best Practices

### 1. Use Non-Production Environment

**Always use a test or staging environment** for integration tests. Never run integration tests against production unless:
- You have explicit permission
- The tests are read-only
- You understand the impact

### 2. Clean Up Resources

Integration tests should:
- Use read-only operations when possible
- Clean up any created resources
- Not interfere with other tests

### 3. Skip on Missing Credentials

Tests automatically skip if required credentials are not set:
```go
if apiURL == "" {
    t.Skip("NEXMONYX_API_URL not set - skipping integration tests")
}
```

### 4. Use Appropriate Timeouts

Integration tests use longer timeouts than unit tests:
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
```

## Troubleshooting

### Tests Skip Immediately

**Problem**: All tests show "SKIP" status

**Solution**: Make sure you're using the `-tags=integration` flag:
```bash
go test -tags=integration -v
```

### Authentication Errors

**Problem**: Tests fail with 401 Unauthorized

**Solutions**:
1. Verify your JWT token is valid and not expired
2. Check that NEXMONYX_AUTH_TOKEN is set correctly
3. Ensure the token has necessary permissions

### Connection Errors

**Problem**: Tests fail with connection refused or timeout

**Solutions**:
1. Verify NEXMONYX_API_URL is correct
2. Check network connectivity to the API
3. Ensure the API server is running
4. Check firewall rules

### Rate Limiting

**Problem**: Tests fail with 429 Too Many Requests

**Solutions**:
1. Wait a few minutes before running tests again
2. Use a test account with higher rate limits
3. Run tests sequentially instead of in parallel

### Debug Mode

Enable debug logging to see full request/response details:
```bash
export NEXMONYX_DEBUG="true"
go test -tags=integration -v
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Integration Tests

on:
  workflow_dispatch:  # Manual trigger only

jobs:
  integration:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'

      - name: Run Integration Tests
        env:
          NEXMONYX_API_URL: ${{ secrets.NEXMONYX_API_URL }}
          NEXMONYX_AUTH_TOKEN: ${{ secrets.NEXMONYX_AUTH_TOKEN }}
        run: go test -tags=integration -v -timeout 30m
```

### GitLab CI Example

```yaml
integration-tests:
  stage: test
  only:
    - schedules  # Run on schedule only
  script:
    - export NEXMONYX_API_URL="${NEXMONYX_API_URL}"
    - export NEXMONYX_AUTH_TOKEN="${NEXMONYX_AUTH_TOKEN}"
    - go test -tags=integration -v -timeout 30m
  timeout: 40m
```

## Safety Checklist

Before running integration tests:

- [ ] Confirmed using test/staging environment
- [ ] Verified credentials are valid
- [ ] Checked that tests won't modify critical data
- [ ] Set appropriate timeout values
- [ ] Reviewed test logs for any issues
- [ ] Ensured tests can be safely re-run

## Support

For issues with integration tests:
1. Check the [main README](README.md) for SDK documentation
2. Review the [CLAUDE.md](CLAUDE.md) file for development guidelines
3. Open an issue on the GitHub repository
4. Contact the Nexmonyx support team

## License

These integration tests are part of the Nexmonyx Go SDK and share the same license.
