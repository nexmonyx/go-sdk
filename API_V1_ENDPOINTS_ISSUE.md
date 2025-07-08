# Critical: v1 Endpoints Returning 500 Errors

## Current Status

After the authentication header updates, we've discovered that:

1. **v2 endpoints** are working correctly WITH X- prefix headers (`X-Server-UUID`, `X-Server-Secret`)
2. **v1 endpoints** are returning 500 errors regardless of header format

## Test Results (2025-07-08 15:30 UTC)

| Endpoint | X-Server-UUID Headers | Server-UUID Headers | Notes |
|----------|----------------------|---------------------|-------|
| POST /v1/servers/heartbeat | ❌ 500 Error | ❌ 500 Error | Critical for agent health |
| PUT /v1/servers/{uuid}/details | ❌ 500 Error | ❌ 500 Error | Blocks server info updates |
| POST /v1/metrics/comprehensive | ❌ 500 Error | ❌ 500 Error | Legacy metrics endpoint |
| POST /v2/metrics/comprehensive | ✅ 200 Success | ❌ 500 Error* | Working with X- headers |
| POST /v2/hardware/inventory | ❌ 500 Error | ❌ 500 Error | Hardware inventory blocked |

*v2 metrics with non-X headers fails with duplicate key error, not auth error

## Evidence

### Working v2 Request:
```bash
curl -X POST "https://api-dev.nexmonyx.com/v2/metrics/comprehensive" \
  -H "X-Server-UUID: 182193a4-0e57-4f14-9d21-a8d41f860e38" \
  -H "X-Server-Secret: 2225e0a23fa038625e30d7696c961df7ed79ec9496b3e1f2748886e7bfbfb7f8" \
  -H "Content-Type: application/json" \
  -d '{"server_uuid": "...", "cpu": {...}}'

# Response: 200 OK - "Comprehensive metrics stored successfully"
```

### Failing v1 Request:
```bash
curl -X POST "https://api-dev.nexmonyx.com/v1/servers/heartbeat" \
  -H "X-Server-UUID: 182193a4-0e57-4f14-9d21-a8d41f860e38" \
  -H "X-Server-Secret: 2225e0a23fa038625e30d7696c961df7ed79ec9496b3e1f2748886e7bfbfb7f8" \
  -H "Content-Type: application/json" \
  -d '{}'

# Response: 500 - "An unexpected error occurred"
```

## Impact

This is **CRITICAL** because:
1. **Heartbeats are failing** - Servers can't report their health status
2. **Server updates are blocked** - Can't update server details/configuration
3. **Hardware inventory fails** - Can't submit hardware information

While v2 metrics submission works, the other critical agent functions are completely broken.

## Root Cause Investigation Needed

Please check:
1. **v1 endpoint error logs** - What's the actual error behind the 500s?
2. **Recent deployments** - Were v1 endpoints changed recently?
3. **Database/permissions** - Is there an issue with v1 endpoint data access?
4. **Routing/middleware** - Is v1 traffic being routed correctly?

## Temporary Workaround

For metrics only, agents can use v2 endpoint with X- headers:
```go
// SDK currently sends X- headers (as of v1.1.3)
client.Metrics.SubmitComprehensiveToTimescale(ctx, metrics)
```

But this doesn't solve heartbeat and server update failures.

## Questions

1. Why are ALL v1 endpoints failing with 500 errors?
2. Why does v2 work with X- headers but v1 doesn't work with any headers?
3. What changed between our initial testing and now?
4. Can you provide the actual error logs for these 500 responses?

## Test Server

Using the same test server as before:
- UUID: `182193a4-0e57-4f14-9d21-a8d41f860e38`
- Environment: `https://api-dev.nexmonyx.com`

## Urgency

This needs immediate attention as it's blocking all agent operations except v2 metrics submission.