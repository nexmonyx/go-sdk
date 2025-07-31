# Unified API Keys Example

This example demonstrates how to use the new unified API key system in the Nexmonyx Go SDK.

## Overview

The unified API key system consolidates all API key types into a single, consistent interface while maintaining backward compatibility with existing code. It supports multiple key types and authentication methods.

## Key Types

- **User Keys** (`user`): Standard user-created keys with configurable capabilities
- **Admin Keys** (`admin`): Administrative keys with elevated permissions
- **Monitoring Agent Keys** (`monitoring_agent`): Keys for monitoring agents (public/private)
- **System Keys** (`system`): System-to-system communication keys
- **Public Agent Keys** (`public_agent`): Keys for public monitoring agents
- **Registration Keys** (`registration`): Keys specifically for server registration
- **Organization Monitoring Keys** (`org_monitoring`): Organization-level monitoring keys

## Authentication Methods

The unified system supports different authentication methods based on key type:

1. **Bearer Token**: Used for monitoring agents and some system keys
2. **Key/Secret Headers**: Used for user and admin keys (`Access-Key`/`Access-Secret`)
3. **Registration Headers**: Used for registration keys (`X-Registration-Key`)

## Usage Examples

### Creating API Keys

```go
// Create admin client
adminClient, err := nexmonyx.NewClient(&nexmonyx.Config{
    Auth: nexmonyx.AuthConfig{
        Token: "your-admin-jwt-token",
    },
})

// Create a user API key
userKeyReq := nexmonyx.NewUserAPIKey(
    "My API Key",
    "For accessing servers and metrics",
    []string{nexmonyx.CapabilityServersRead, nexmonyx.CapabilityMetricsRead},
)
userKey, err := adminClient.APIKeys.CreateUnified(ctx, userKeyReq)

// Create a monitoring agent key
agentKeyReq := nexmonyx.NewMonitoringAgentKey(
    "Production Agent",
    "Agent for production monitoring",
    "production",
    "private",
    "us-east-1",
    []string{"public", "private"},
)
agentKey, err := adminClient.APIKeys.CreateUnified(ctx, agentKeyReq)

// Create a registration key
regKeyReq := nexmonyx.NewRegistrationKey(
    "Server Registration",
    "For registering new servers",
    organizationID,
)
regKey, err := adminClient.APIKeys.AdminCreateUnified(ctx, regKeyReq)
```

### Using API Keys for Authentication

```go
// Using a user key with key/secret authentication
userClient := client.WithUnifiedAPIKeyAndSecret(userKey.KeyValue, userKey.Secret)

// Using a monitoring agent key with bearer token
agentClient := client.WithUnifiedAPIKey(agentKey.FullToken)

// Using a registration key
regClient := client.WithRegistrationKey(regKey.FullToken)
```

### Server Registration with Unified Keys

```go
// Register a server using a registration key
serverReq := &nexmonyx.ServerCreateRequest{
    Hostname:    "web-server-01",
    MainIP:      "10.0.1.100",
    OS:          "Linux",
    Environment: "production",
}

server, err := regClient.Servers.RegisterWithKey(ctx, regKey.FullToken, serverReq)
```

### Key Management and Validation

```go
// List API keys with filtering
keys, meta, err := adminClient.APIKeys.AdminListUnified(ctx, &nexmonyx.ListUnifiedAPIKeysOptions{
    Type:   nexmonyx.APIKeyTypeUser,
    Status: nexmonyx.APIKeyStatusActive,
})

// Validate key capabilities
for _, key := range keys {
    if key.IsActive() && key.HasCapability(nexmonyx.CapabilityServersRead) {
        fmt.Printf("Key %s can read servers\n", key.Name)
    }
}
```

## Capabilities

The system uses fine-grained capabilities instead of broad scopes:

- `servers:read`, `servers:write`, `servers:register`, `servers:delete`, `servers:*`
- `monitoring:read`, `monitoring:write`, `monitoring:execute`, `monitoring:*`
- `probes:read`, `probes:write`, `probes:execute`, `probes:*`
- `metrics:read`, `metrics:write`, `metrics:submit`, `metrics:*`
- `organization:read`, `organization:write`, `organization:*`
- `admin:read`, `admin:write`, `admin:*`
- `*` (wildcard for full access)

## Backward Compatibility

The unified system maintains full backward compatibility:

```go
// Legacy API key usage continues to work
apiKey := &nexmonyx.APIKey{
    Name:   "Legacy Key",
    Scopes: []string{"servers:read"},
}
created, err := client.APIKeys.Create(ctx, apiKey)

// Legacy authentication methods still work
legacyClient := client.WithAPIKey("key", "secret")
```

## Migration Path

1. **Immediate**: Start using unified key creation methods for new keys
2. **Gradual**: Replace legacy authentication with unified methods
3. **Long-term**: Migrate existing keys to unified system as they're renewed

## Running the Example

```bash
export NEXMONYX_ADMIN_TOKEN="your-admin-jwt-token"
go run examples/unified_api_keys/main.go
```

## Environment Variables

- `NEXMONYX_ADMIN_TOKEN`: JWT token with admin privileges for creating API keys
- `NEXMONYX_API_URL`: API base URL (defaults to https://api.nexmonyx.com)
- `NEXMONYX_DEBUG`: Enable debug logging (set to "true")