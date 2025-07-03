# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This is the official **Nexmonyx Go SDK** - a comprehensive client library for the Nexmonyx API, a server monitoring and management platform. The SDK provides full coverage of the Nexmonyx API with type-safe operations, authentication support, and comprehensive error handling.

**Note**: This repository has been reorganized for better maintainability while maintaining full backwards compatibility. See `docs/REORGANIZATION_SUMMARY.md` for details.

## Development Commands

### Testing Commands
```bash
# Run all unit tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out

# Run tests with verbose output
go test -v ./...

# Run only unit tests (excluding integration)
go test -short ./...

# Run integration tests (requires environment setup)
export NEXMONYX_INTEGRATION_TESTS="true"
export NEXMONYX_API_URL="https://api.nexmonyx.com"
export NEXMONYX_AUTH_TOKEN="your-token"
go test -v -tags=integration -timeout 30m ./...
```

### Build Commands
```bash
# Build all packages
go build ./...

# Build with verbose output
go build -v ./...

# Install dependencies
go mod download

# Clean up module dependencies
go mod tidy

# Verify module dependencies
go mod verify
```

### Code Quality Commands
```bash
# Format code
go fmt ./...

# Run go vet for static analysis
go vet ./...

# Install and run golint
go install golang.org/x/lint/golint@latest
golint ./...

# Install and run staticcheck
go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck ./...

# Install and run gosec (security scanner)
go install github.com/securego/gosec/v2/cmd/gosec@latest
gosec ./...
```

### Complete Development Workflow
```bash
# 1. Dependency management
go mod tidy && go mod verify

# 2. Code formatting and analysis
go fmt ./... && go vet ./...

# 3. Linting (install tools if needed)
golint ./... && staticcheck ./...

# 4. Security scanning
gosec ./...

# 5. Testing with coverage
go test -v -race -coverprofile=coverage.out ./...

# 6. Build verification
go build -v ./...
```

## Architecture Overview

### Repository Structure

```
pkg/nexmonyx/               # Main SDK package
├── client.go               # Core client implementation
├── errors.go               # Error types and handling
├── models/                 # All data models
│   └── models.go           # Combined models file
├── services/               # Service implementations (27 files)
│   ├── organizations.go    # Organizations service
│   ├── servers.go          # Servers service
│   ├── monitoring.go       # Monitoring service
│   └── ...                 # Other services
├── helpers/                # Helper utilities
│   ├── response.go         # Response helpers
│   └── metrics_helpers.go  # Metrics helpers
└── hardware/               # Hardware-specific services
    ├── inventory.go        # Hardware inventory
    ├── ipmi.go            # IPMI management
    └── systemd.go         # Systemd management
```

### Core Components

**Main Client (`pkg/nexmonyx/client.go`)**
- `Client` struct is the main entry point for SDK operations
- Supports multiple authentication methods: JWT tokens, API keys, server credentials, monitoring keys
- Built on `github.com/go-resty/resty/v2` HTTP client with automatic retry logic and error handling
- Provides service-specific clients for different API domains

**Authentication Methods**
1. **JWT Token**: For user authentication (`Token`)
2. **API Key/Secret**: For service-to-service authentication (`APIKey`, `APISecret`)
3. **Server Credentials**: For agent authentication (`ServerUUID`, `ServerSecret`)
4. **Monitoring Key**: For monitoring agent authentication (`MonitoringKey`)

**Service Architecture**
The SDK is organized into specialized service clients for different API domains:

- **Organizations** - Organization management and membership
- **Servers** - Server registration, monitoring, and management
- **Users** - User profile and preference management
- **Metrics** - Metrics submission and querying
- **Monitoring** - Probes, regions, and monitoring infrastructure
- **Billing** - Subscription and billing management
- **Alerts** - Alert rules and notification channels
- **VMs** - Virtual machine and cloud provider management
- **Jobs/BackgroundJobs** - Background job and task management
- **APIKeys** - API key creation and management
- **System** - Health, version, and system status
- **Admin** - Administrative operations (with BillingOverrides subservice)

### Data Models (`pkg/nexmonyx/models/models.go`)

**Base Models**
- `GormModel` - Base model with ID, timestamps
- `BaseModel` - Base model with UUID, timestamps
- `CustomTime` - Custom time handling for various API date formats

**Core Domain Models**
- `Organization` - Organizations with subscription and monitoring details
- `User` - Users with RBAC roles and preferences
- `Server` - Monitored servers with hardware and metrics
- `Alert` - Alert rules and notification configuration
- `Probe` - Monitoring probes with regional execution
- `Subscription` - Billing subscriptions and invoices

**Monitoring Models**
- `ProbeRequest`, `ProbeResult`, `ProbeMetrics` - Monitoring probe lifecycle
- `MonitoringAgent`, `RegionalController` - Monitoring infrastructure
- `ControllerHeartbeat` - Controller status and health tracking

**Billing Models**
- `BillingOverride`, `OrganizationUsage` - Billing limits and usage tracking
- `BillingJob`, `UsageHistory` - Background billing operations

### Error Handling (`pkg/nexmonyx/errors.go`)

Structured error types for different HTTP status codes:
- `APIError` - General API errors with details
- `RateLimitError` - Rate limit exceeded (429)
- `ValidationError` - Request validation failures (400)
- `NotFoundError` - Resource not found (404) 
- `UnauthorizedError` - Authentication required (401)
- `ForbiddenError` - Insufficient permissions (403)

## Common Usage Patterns

### Client Initialization
```go
// JWT authentication
client, err := nexmonyx.NewClient(&nexmonyx.Config{
    BaseURL: "https://api.nexmonyx.com",
    Auth: nexmonyx.AuthConfig{Token: "jwt-token"},
})

// Server credentials (for agents)
client, err := nexmonyx.NewClient(&nexmonyx.Config{
    Auth: nexmonyx.AuthConfig{
        ServerUUID:   "server-uuid",
        ServerSecret: "server-secret",
    },
})
```

### Service Usage
```go
// List servers with pagination
servers, meta, err := client.Servers.List(ctx, &nexmonyx.ListOptions{
    Page: 1, Limit: 25, Search: "web-server",
})

// Submit metrics (agent use case)
err = client.Metrics.SubmitComprehensive(ctx, metricsData)

// Get organization details
org, err := client.Organizations.Get(ctx, "org-uuid")
```

### Error Handling
```go
_, err := client.Users.GetMe(ctx)
if err != nil {
    switch e := err.(type) {
    case *nexmonyx.UnauthorizedError:
        // Handle authentication failure
    case *nexmonyx.RateLimitError:
        // Handle rate limiting with retry
    case *nexmonyx.ValidationError:
        // Handle validation errors
    }
}
```

## Key Dependencies

- **Go 1.24** - Latest Go version
- **github.com/go-resty/resty/v2** - HTTP client library
- **github.com/stretchr/testify** - Testing framework

## Integration Testing

Integration tests require environment variables:
```bash
export NEXMONYX_INTEGRATION_TESTS="true"
export NEXMONYX_API_URL="https://api.nexmonyx.com"
export NEXMONYX_AUTH_TOKEN="your-jwt-token"
export NEXMONYX_SERVER_UUID="your-server-uuid"  # Optional for agent tests
export NEXMONYX_SERVER_SECRET="your-server-secret"  # Optional for agent tests
export NEXMONYX_DEBUG="true"  # Optional for debug logging
```

## File Organization

### Working with Services
- Service implementations are in `pkg/nexmonyx/services/`
- Each service file contains the service struct, request/response types, and methods
- Hardware-specific services are in `pkg/nexmonyx/hardware/`

### Working with Models
- All data models are consolidated in `pkg/nexmonyx/models/models.go`
- Models are organized by domain (organizations, servers, monitoring, billing)

### Working with Tests
- Unit tests are in `tests/unit/`
- Integration tests are in `tests/integration/`
- Examples are in `examples/basic/`

### Backwards Compatibility
- Main package file `nexmonyx.go` re-exports all types for compatibility
- Existing code continues to work without changes
- New development can optionally use new package structure

## Notes for Development

- All API operations require `context.Context` for cancellation support
- The SDK supports comprehensive pagination with `ListOptions`
- Rate limiting and retries are handled automatically
- Debug mode can be enabled for request/response logging
- The SDK follows standard Go error handling patterns
- All service methods return typed responses with metadata when applicable
- Repository has been reorganized for better maintainability (see `docs/REORGANIZATION_SUMMARY.md`)