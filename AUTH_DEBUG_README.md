# Authentication Debugging Tools for Nexmonyx SDK

This document describes the debugging tools added to help diagnose the authentication header issue where agents are receiving "Missing Server-UUID or Server-Secret headers" errors.

## Summary of Changes

### 1. Enhanced Debug Logging in `client.go`

Added debug logging to the `Do()` method to show all headers being sent:
- Shows request method and path
- Lists all headers (with sensitive values redacted)
- Captures error response details including status code, body, and headers

To enable debug logging:
```go
client, err := nexmonyx.NewClient(&nexmonyx.Config{
    BaseURL: "https://api.nexmonyx.com",
    Auth: nexmonyx.AuthConfig{
        ServerUUID:   "your-uuid",
        ServerSecret: "your-secret",
    },
    Debug: true,  // Enable debug mode
})
```

### 2. Authentication Debug Method

Added `DebugAuthHeaders()` method to test both header formats:
```go
err := client.DebugAuthHeaders(context.Background())
```

This method:
- Tests headers WITH 'X-' prefix (X-Server-UUID, X-Server-Secret)
- Tests headers WITHOUT 'X-' prefix (Server-UUID, Server-Secret)
- Reports which format the API accepts
- Provides clear conclusions about what needs to be fixed

### 3. Command-Line Debug Tool

Created `examples/auth_debug/main.go` for testing authentication:

```bash
# Basic usage
go run examples/auth_debug/main.go -uuid <SERVER_UUID> -secret <SERVER_SECRET>

# With debug output
go run examples/auth_debug/main.go -uuid <SERVER_UUID> -secret <SERVER_SECRET> -debug

# Test both header formats
go run examples/auth_debug/main.go -uuid <SERVER_UUID> -secret <SERVER_SECRET> -test-auth

# Use different API endpoint
go run examples/auth_debug/main.go -uuid <SERVER_UUID> -secret <SERVER_SECRET> -api https://api-dev.nexmonyx.com
```

### 4. Shell Script for Direct Testing

Created `test_headers.sh` to test authentication using curl:

```bash
# Using command line arguments
./test_headers.sh <SERVER_UUID> <SERVER_SECRET> [API_URL]

# Using environment variables
export NEXMONYX_SERVER_UUID="your-uuid"
export NEXMONYX_SERVER_SECRET="your-secret"
./test_headers.sh
```

## How to Use These Tools

### For SDK Users/Agent Developers:

1. **Enable Debug Mode in Your Agent:**
   ```go
   sdkConfig := &nexmonyx.Config{
       BaseURL: config.CFG.Endpoint,
       Auth: nexmonyx.AuthConfig{
           ServerUUID:   config.CFG.ServerUUID,
           ServerSecret: config.CFG.ServerSecret,
       },
       Debug: true,  // Add this line
   }
   ```

2. **Run the Authentication Test:**
   ```bash
   # Test with your server credentials
   ./test_headers.sh 182193a4-0e57-4f14-9d21-a8d41f860e38 your-secret https://api-dev.nexmonyx.com
   ```

3. **Use the Debug Tool:**
   ```bash
   go run examples/auth_debug/main.go \
     -uuid 182193a4-0e57-4f14-9d21-a8d41f860e38 \
     -secret your-secret \
     -api https://api-dev.nexmonyx.com \
     -test-auth
   ```

### For API Team:

The test results will clearly show:
- Which header format your API expects
- The exact error message returned for failed authentication
- Whether the issue is with header names or something else

## Expected Outcomes

Based on the error message "Missing Server-UUID or Server-Secret headers", we suspect the API expects headers WITHOUT the 'X-' prefix, while the SDK is sending them WITH the prefix.

The tools will confirm:
1. ✅ If only "Server-UUID" works → SDK needs update
2. ✅ If only "X-Server-UUID" works → API error message is misleading
3. ✅ If both work → Issue is elsewhere
4. ❌ If neither works → Credentials or other auth issue

## Next Steps

After running these tests:

1. **If SDK needs update:** We'll change header names from "X-Server-UUID" to "Server-UUID"
2. **If API needs update:** API team should fix error messages or accept both formats
3. **If neither work:** Investigation needed on API authentication middleware

## Version Info

These debugging tools are included in SDK v1.1.1+ and can be used immediately without any SDK updates.