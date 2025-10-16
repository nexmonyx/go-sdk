# Integration Testing Guide

Complete guide to setting up and running integration tests for the Nexmonyx Go SDK using Docker Compose and mock API services.

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Local Development](#local-development)
- [Environment Configuration](#environment-configuration)
- [Docker Services](#docker-services)
- [Running Tests](#running-tests)
- [CI/CD Integration](#cicd-integration)
- [Troubleshooting](#troubleshooting)
- [Advanced Usage](#advanced-usage)

## Overview

The integration testing framework for the Nexmonyx Go SDK provides:

- **Mock API Server**: In-memory HTTP server mimicking Nexmonyx API endpoints
- **Docker Compose**: Containerized environment for consistent testing across machines
- **Test Fixtures**: Realistic sample data for all major entities
- **Automated Setup**: Scripts to configure environment and manage resources
- **CI/CD Ready**: Pre-configured GitHub Actions workflows

### Key Benefits

✅ **No External Dependencies** - Tests run with mock API, no real credentials needed
✅ **Fast Execution** - In-memory mock server, no network latency
✅ **Consistent Results** - Same environment locally and in CI/CD
✅ **Easy Cleanup** - Automated resource management scripts
✅ **Production-Ready** - Optional real API testing for validation

## Quick Start

### 1. Initial Setup (One-time)

```bash
# From repository root
./scripts/setup-test-env.sh
```

This will:
- Validate Docker installation
- Create `.env` file from template
- Verify environment variables
- Display next steps

### 2. Start Services

```bash
docker-compose up -d
```

Verify services are running:
```bash
docker-compose ps
```

Expected output:
```
CONTAINER ID   IMAGE              STATUS              PORTS
...           nexmonyx-mock-api  Up 2 seconds        0.0.0.0:8080->8080/tcp
```

### 3. Run Tests

```bash
# Load environment variables
source .env

# Run all integration tests
INTEGRATION_TESTS=true go test -v ./tests/integration/...

# Run specific test
INTEGRATION_TESTS=true go test -v -run TestServerLifecycle ./tests/integration/
```

### 4. Stop Services

```bash
docker-compose down
```

## Local Development

### Development Workflow

1. **Start environment** (first time or after cleanup):
   ```bash
   ./scripts/setup-test-env.sh
   docker-compose up -d
   ```

2. **Develop and test** (iteratively):
   ```bash
   source .env
   go test -v ./tests/integration/...
   ```

3. **Check service logs** (if tests fail):
   ```bash
   docker-compose logs -f mock-api
   ```

4. **Access service directly** (for debugging):
   ```bash
   curl -H "Authorization: Bearer test-token" http://localhost:8080/api/v1/servers
   ```

5. **Stop environment** (when done):
   ```bash
   docker-compose down
   ```

### Development Environment Variables

Key variables to customize in `.env`:

```bash
# Enable integration tests
INTEGRATION_TESTS=true

# Use mock server (local)
INTEGRATION_TEST_MODE=mock
INTEGRATION_TEST_API_URL=http://localhost:8080

# Authentication token
INTEGRATION_TEST_AUTH_TOKEN=test-token

# Debug logging
INTEGRATION_TEST_DEBUG=true

# Custom timeout for long tests
INTEGRATION_TEST_TIMEOUT=60s
```

## Environment Configuration

### `.env` File Template

The `.env.example` file provides a complete template with:

- **Integration Test Settings**: Mode, URL, token, debug, timeout
- **Docker Service Configuration**: Ports, credentials, log levels
- **Database Settings**: For optional PostgreSQL testing
- **CI/CD Configuration**: Comments for GitHub Secrets

### Creating Custom `.env`

```bash
# Copy template
cp .env.example .env

# Edit for your environment
nano .env

# Verify configuration
cat .env | grep -v "^#" | grep -v "^$"
```

### Environment Variables Reference

| Variable | Default | Description |
|----------|---------|-------------|
| `INTEGRATION_TESTS` | `false` | Enable integration tests |
| `INTEGRATION_TEST_MODE` | `mock` | Test mode: "mock" or "dev" |
| `INTEGRATION_TEST_API_URL` | `http://localhost:8080` | API endpoint URL |
| `INTEGRATION_TEST_AUTH_TOKEN` | `test-token` | Authentication token |
| `INTEGRATION_TEST_DEBUG` | `false` | Enable debug logging |
| `INTEGRATION_TEST_TIMEOUT` | `30s` | Test timeout duration |
| `API_PORT` | `8080` | Mock API server port |
| `API_HOST` | `0.0.0.0` | Mock API server host |

## Docker Services

### Mock API Server

**Container**: `nexmonyx-mock-api`
**Port**: 8080
**Health Check**: `/health` endpoint

#### Starting the Service

```bash
# Start all services
docker-compose up -d

# Start specific service
docker-compose up -d mock-api

# View logs
docker-compose logs -f mock-api
```

#### Service Health

```bash
# Check service status
docker-compose ps

# Direct health check
curl http://localhost:8080/health

# Detailed status
docker-compose exec mock-api curl http://localhost:8080/health
```

### Optional Services

The `docker-compose.yml` includes commented-out optional services:

- **PostgreSQL with TimescaleDB**: For persistence testing
- **Redis**: For caching/session testing

To enable these services, uncomment them in `docker-compose.yml` and rebuild:

```bash
# Edit docker-compose.yml
nano docker-compose.yml

# Restart with new services
docker-compose up -d
```

## Running Tests

### Basic Test Execution

```bash
# Run all integration tests
source .env
INTEGRATION_TESTS=true go test -v ./tests/integration/...

# Run tests without verbose output
INTEGRATION_TESTS=true go test ./tests/integration/...

# Run tests with race detector
INTEGRATION_TESTS=true go test -race ./tests/integration/...
```

### Running Specific Tests

```bash
# Run by test name pattern
INTEGRATION_TESTS=true go test -run TestServer ./tests/integration/

# Run by full test name
INTEGRATION_TESTS=true go test -run TestServerLifecycleWorkflow ./tests/integration/

# Run workflow tests only
INTEGRATION_TESTS=true go test -run ".*Workflow" ./tests/integration/
```

### Coverage Reporting

```bash
# Run tests with coverage
INTEGRATION_TESTS=true go test -cover -coverprofile=integration-coverage.out ./tests/integration/...

# View coverage report
go tool cover -html=integration-coverage.out

# Coverage summary
go tool cover -func=integration-coverage.out
```

### Performance Testing

```bash
# Run with short mode (skips long tests)
INTEGRATION_TESTS=true go test -short ./tests/integration/...

# Run with custom timeout
INTEGRATION_TESTS=true go test -timeout 60s ./tests/integration/...

# Benchmark specific tests
INTEGRATION_TESTS=true go test -bench=. ./tests/integration/...
```

## CI/CD Integration

### GitHub Actions Workflow

The `.github/workflows/integration-tests.yml` file:

1. **Builds Docker image** of mock API server
2. **Starts services** as containers
3. **Waits for health** checks to pass
4. **Runs integration tests** with mock API
5. **Collects coverage** reports
6. **Comments** results on pull requests
7. **Uploads artifacts** for inspection

### Workflow Configuration

Services run automatically in CI/CD:

```yaml
services:
  mock-api:
    image: nexmonyx/mock-api:latest
    ports:
      - 8080:8080
    healthcheck:
      test: curl -f http://localhost:8080/health
      interval: 5s
      timeout: 3s
      retries: 5
```

### GitHub Secrets (Optional)

For testing against real dev API:

1. Go to: Repository → Settings → Secrets and variables → Actions
2. Create these secrets:
   - `NEXMONYX_API_URL` - Dev API URL
   - `NEXMONYX_AUTH_TOKEN` - JWT token
   - `CODECOV_TOKEN` - For coverage reports

3. Use in workflows:
   ```yaml
   env:
     INTEGRATION_TEST_MODE: dev
     INTEGRATION_TEST_API_URL: ${{ secrets.NEXMONYX_API_URL }}
     INTEGRATION_TEST_AUTH_TOKEN: ${{ secrets.NEXMONYX_AUTH_TOKEN }}
   ```

## Troubleshooting

### Docker Issues

#### Port already in use
```bash
# Find process using port 8080
lsof -i :8080

# Stop the process or use different port
# Edit docker-compose.yml ports: "8081:8080"
```

#### Container fails to start
```bash
# Check logs
docker-compose logs mock-api

# Rebuild image
docker-compose build --no-cache mock-api

# Start fresh
docker-compose down -v
docker-compose up -d
```

#### Network issues
```bash
# Inspect network
docker network inspect nexmonyx-test-network

# Recreate network
docker-compose down -v
docker-compose up -d
```

### Test Issues

#### Tests hang or timeout
```bash
# Increase timeout
INTEGRATION_TESTS=true go test -timeout 120s ./tests/integration/...

# Check if mock API is running
curl http://localhost:8080/health

# View mock API logs
docker-compose logs mock-api
```

#### Tests fail with 401/403 errors
```bash
# Verify auth token in .env
grep INTEGRATION_TEST_AUTH_TOKEN .env

# Check API requires auth
curl -i http://localhost:8080/api/v1/servers

# Add token to request
curl -H "Authorization: Bearer test-token" \
     http://localhost:8080/api/v1/servers
```

#### Database connection errors
```bash
# Verify PostgreSQL service is enabled in docker-compose.yml
docker-compose ps

# Check database is ready
docker-compose exec postgres pg_isready -U test -d nexmonyx_test

# View database logs
docker-compose logs postgres
```

### Performance Issues

#### Slow test execution
```bash
# Run tests in parallel
go test -parallel 4 ./tests/integration/...

# Reduce verbosity
INTEGRATION_TESTS=true go test ./tests/integration/...

# Check system resources
docker stats
```

## Advanced Usage

### Real API Testing

To test against real Nexmonyx dev API:

1. **Obtain credentials**:
   ```bash
   # Get JWT token from dev environment
   echo "Token: your-jwt-token"
   ```

2. **Configure `.env`**:
   ```bash
   INTEGRATION_TEST_MODE=dev
   INTEGRATION_TEST_API_URL=https://dev-api.nexmonyx.com
   INTEGRATION_TEST_AUTH_TOKEN=your-jwt-token
   ```

3. **Run tests**:
   ```bash
   source .env
   INTEGRATION_TESTS=true go test -v ./tests/integration/...
   ```

⚠️ **Warning**: Real API testing creates actual resources. Use dedicated test accounts and clean up afterward.

### Database Testing

To use PostgreSQL for persistence testing:

1. **Uncomment PostgreSQL** in `docker-compose.yml`
2. **Create schema fixtures** in `tests/integration/fixtures/schema.sql`
3. **Update tests** to use database connections
4. **Run tests** with database enabled

```bash
# Verify database is ready
docker-compose exec postgres pg_isready -U test -d nexmonyx_test

# Connect to database
docker-compose exec postgres psql -U test -d nexmonyx_test
```

### Debugging Test Failures

```bash
# Enable debug logging
INTEGRATION_TEST_DEBUG=true INTEGRATION_TESTS=true go test -v ./tests/integration/

# Run single test with debug
INTEGRATION_TEST_DEBUG=true INTEGRATION_TESTS=true \
  go test -run TestServerLifecycle -v ./tests/integration/

# Shell into mock API container
docker-compose exec mock-api sh

# View complete request/response
curl -v -H "Authorization: Bearer test-token" \
     http://localhost:8080/api/v1/servers | jq
```

## Reference

- **[Integration Test README](../tests/integration/README.md)** - Test framework details
- **[Docker Documentation](../tests/integration/docker/README.md)** - Docker-specific info
- **[Testing Standards](./TESTING.md)** - SDK testing best practices
- **[Docker Compose Docs](https://docs.docker.com/compose/)** - Compose reference

