# Debug Guide for Nexmonyx SDK

This guide explains how to use the comprehensive debug logging features added in v1.1.5 to troubleshoot heartbeat and server information operations.

## Enabling Debug Mode

To enable debug logging, set `Debug: true` in your client configuration:

```go
client, err := nexmonyx.NewClient(&nexmonyx.Config{
    BaseURL: "https://api.nexmonyx.com",
    Auth: nexmonyx.AuthConfig{
        ServerUUID:   "your-server-uuid",
        ServerSecret: "your-server-secret",
    },
    Debug: true,  // Enable debug logging
})
```

## Debug Output Examples

### Heartbeat Operations

#### Basic Heartbeat
```
[DEBUG] Heartbeat: Starting heartbeat request
[DEBUG] Heartbeat: Endpoint: POST /v1/heartbeat
[DEBUG] Heartbeat: Using server UUID: 182193a4-0e57-4f14-9d21-a8d41f860e38
[DEBUG] Request: POST /v1/heartbeat
[DEBUG] Headers being sent:
[DEBUG]   X-Server-UUID: 182193a4-0e57-4f14-9d21-a8d41f860e38
[DEBUG]   X-Server-Secret: [REDACTED]
[DEBUG]   Content-Type: [application/json]
[DEBUG]   User-Agent: [nexmonyx-go-sdk/1.1.5]
[DEBUG] Heartbeat: Request successful
[DEBUG] Heartbeat: Response status: success
[DEBUG] Heartbeat: Response message: Heartbeat received
[DEBUG] Heartbeat: HTTP Status Code: 200
```

#### Heartbeat with Version
```
[DEBUG] HeartbeatWithVersion: Starting heartbeat request with version
[DEBUG] HeartbeatWithVersion: Endpoint: POST /v1/heartbeat
[DEBUG] HeartbeatWithVersion: Agent version: v1.0.0
[DEBUG] HeartbeatWithVersion: Using server UUID: 182193a4-0e57-4f14-9d21-a8d41f860e38
[DEBUG] HeartbeatWithVersion: Request body: map[agent_version:v1.0.0]
[DEBUG] Request has body (type: map[string]string)
```

### Server Update Operations

#### Update Details
```
[DEBUG] UpdateDetails: Starting server details update
[DEBUG] UpdateDetails: Endpoint: PUT /v1/server/182193a4-0e57-4f14-9d21-a8d41f860e38/details
[DEBUG] UpdateDetails: Server UUID: 182193a4-0e57-4f14-9d21-a8d41f860e38
[DEBUG] UpdateDetails: Request data:
[DEBUG]   Hostname: web-server-01
[DEBUG]   OS: Ubuntu 22.04 LTS
[DEBUG]   Kernel: 5.15.0-88-generic
[DEBUG]   Architecture: x86_64
[DEBUG]   CPUModel: Intel(R) Core(TM) i7-9700K CPU @ 3.60GHz
[DEBUG]   CPUCores: 8
[DEBUG]   MemoryTotalMB: 16384
[DEBUG]   DiskTotalGB: 500
[DEBUG] UpdateDetails: Using authentication - Server UUID: 182193a4-0e57-4f14-9d21-a8d41f860e38
[DEBUG] Request has body (type: *nexmonyx.ServerDetailsUpdateRequest)
[DEBUG] UpdateDetails: Request successful
[DEBUG] UpdateDetails: Response status: success
[DEBUG] UpdateDetails: Response message: Server information updated successfully
[DEBUG] UpdateDetails: HTTP Status Code: 200
[DEBUG] UpdateDetails: Server ID: 123
[DEBUG] UpdateDetails: Server UUID: 182193a4-0e57-4f14-9d21-a8d41f860e38
[DEBUG] UpdateDetails: Server Hostname: web-server-01
```

### Error Scenarios

#### Authentication Failure
```
[DEBUG] Heartbeat: Starting heartbeat request
[DEBUG] Heartbeat: Endpoint: POST /v1/heartbeat
[DEBUG] Heartbeat: Using server UUID: invalid-uuid
[DEBUG] Error Response: Status=401
[DEBUG] Error Body: {"error":"unauthorized","message":"Invalid or missing authentication headers"}
[DEBUG] Heartbeat: Request failed with error: authentication required
```

#### Network Error
```
[DEBUG] UpdateDetails: Starting server details update
[DEBUG] UpdateDetails: Endpoint: PUT /v1/server/182193a4-0e57-4f14-9d21-a8d41f860e38/details
[DEBUG] UpdateDetails: Request failed with error: request failed: Post "https://api.nexmonyx.com/v1/server/...": dial tcp: lookup api.nexmonyx.com: no such host
```

## Using the Debug Example

A comprehensive debug example is provided in `examples/debug_heartbeat/main.go`:

```bash
# Test heartbeat
go run examples/debug_heartbeat/main.go \
  -uuid YOUR_SERVER_UUID \
  -secret YOUR_SERVER_SECRET \
  -api https://api-dev.nexmonyx.com \
  -test heartbeat

# Test server details update
go run examples/debug_heartbeat/main.go \
  -uuid YOUR_SERVER_UUID \
  -secret YOUR_SERVER_SECRET \
  -test update-details

# Test get heartbeat
go run examples/debug_heartbeat/main.go \
  -uuid YOUR_SERVER_UUID \
  -secret YOUR_SERVER_SECRET \
  -test get-heartbeat
```

## Interpreting Debug Output

### What to Look For

1. **Endpoint Path**: Verify the correct endpoint is being called
2. **Authentication Headers**: Ensure X-Server-UUID and X-Server-Secret are present
3. **Request Body**: For updates, verify all fields are populated correctly
4. **HTTP Status Code**: 
   - 200/201: Success
   - 401: Authentication issue
   - 404: Wrong endpoint or resource not found
   - 500: Server error
5. **Response Status**: Should be "success" for successful operations
6. **Error Messages**: Detailed error information for troubleshooting

### Common Issues

1. **"Missing Server-UUID or Server-Secret headers"**
   - Check that server credentials are correctly configured
   - Verify headers are being sent (shown in debug output)

2. **"Unhandled route" errors (500)**
   - Endpoint path may be incorrect
   - Check the debug output for the exact path being called

3. **Empty responses**
   - May indicate network issues or timeout
   - Check for connection errors in debug output

## Debug Logging Best Practices

1. **Enable Only When Needed**: Debug mode generates verbose output - use only for troubleshooting
2. **Secure Sensitive Data**: The SDK automatically redacts secrets in debug output
3. **Capture Full Output**: When reporting issues, include the complete debug log
4. **Test in Isolation**: Use the debug example to test specific operations
5. **Compare Working vs Failing**: Debug output from successful operations helps identify issues

## Reporting Issues

When reporting SDK issues, please include:
1. SDK version (shown in User-Agent header)
2. Complete debug output
3. Expected vs actual behavior
4. Any error messages
5. Network environment (proxy, firewall, etc.)

The comprehensive debug logging in v1.1.5 makes it much easier to identify and resolve authentication, network, and API integration issues.