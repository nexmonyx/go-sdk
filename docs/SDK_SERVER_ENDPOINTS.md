# SDK Server Endpoints Coverage

## API Endpoints → SDK Methods Mapping

| API Endpoint | HTTP Method | SDK Method | Authentication | Status |
|--------------|-------------|------------|----------------|---------|
| `/v1/server/:uuid/info` | PUT | `UpdateInfo(ctx, uuid, req)` | Server Creds | ✅ Added |
| `/v1/server/:uuid/details` | PUT | `UpdateDetails(ctx, uuid, req)` | Server Creds | ✅ Exists |
| `/v1/server/:uuid/heartbeat` | PUT | `UpdateHeartbeat(ctx, uuid)` | Server Creds | ✅ Added |
| `/v1/server/:uuid/details` | GET | `GetDetails(ctx, uuid)` | Server Creds | ✅ Added |
| `/v1/server/:uuid/full-details` | GET | `GetFullDetails(ctx, uuid)` | JWT + servers:read | ✅ Added |
| `/v1/server/:uuid/heartbeat` | GET | `GetHeartbeat(ctx, uuid)` | Server Creds | ✅ Added |
| `/v1/heartbeat` | POST | `Heartbeat(ctx)` | Server Creds | ✅ Exists |

## Usage Examples

### 1. Update Server Info
```go
details := &nexmonyx.ServerDetailsUpdateRequest{
    Hostname: "web-server-01",
    OS:       "Ubuntu 22.04",
    Kernel:   "5.15.0-88-generic",
    // ... other fields
}

// Update via /info endpoint
server, err := client.Servers.UpdateInfo(ctx, serverUUID, details)

// Or update via /details endpoint (same data)
server, err := client.Servers.UpdateDetails(ctx, serverUUID, details)
```

### 2. Get Server Details
```go
// Basic details (server auth)
server, err := client.Servers.GetDetails(ctx, serverUUID)

// Full details including CPU (requires JWT auth)
server, err := client.Servers.GetFullDetails(ctx, serverUUID)
```

### 3. Heartbeat Operations
```go
// Send heartbeat (authenticated server)
err := client.Servers.Heartbeat(ctx)

// Update heartbeat for specific server
err := client.Servers.UpdateHeartbeat(ctx, serverUUID)

// Get heartbeat info
heartbeat, err := client.Servers.GetHeartbeat(ctx, serverUUID)
if err == nil {
    fmt.Printf("Last heartbeat: %v\n", heartbeat.LastHeartbeat)
    fmt.Printf("Server status: %s\n", heartbeat.ServerStatus)
}
```

## Authentication Requirements

### Server Credential Auth
Most endpoints use server credentials (`X-Server-UUID` and `X-Server-Secret`):
```go
client, err := nexmonyx.NewClient(&nexmonyx.Config{
    Auth: nexmonyx.AuthConfig{
        ServerUUID:   "your-server-uuid",
        ServerSecret: "your-server-secret",
    },
})
```

### JWT Auth
The `GetFullDetails` endpoint requires JWT authentication:
```go
client, err := nexmonyx.NewClient(&nexmonyx.Config{
    Auth: nexmonyx.AuthConfig{
        Token: "your-jwt-token",
    },
})
```

## Notes

1. **API Consistency**: The API now accepts both `X-Server-UUID` and `Server-UUID` header formats
2. **Endpoint Paths**: All paths use singular `/v1/server/` (not `/v1/servers/`)
3. **Response Types**: All methods return `*Server` except heartbeat methods
4. **Error Handling**: All methods return appropriate error types defined in `errors.go`