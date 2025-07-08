# API Authentication Header Inconsistency - Fix Required

## Current Situation

The SDK authentication issues have been partially resolved by fixing incorrect endpoint paths. However, testing revealed that the API has **inconsistent authentication header expectations** across different endpoints.

## Inconsistency Details

### Current Behavior

| Endpoint | Expected Headers | Working? |
|----------|-----------------|----------|
| `POST /v1/heartbeat` | `X-Server-UUID`, `X-Server-Secret` | ✅ Yes |
| `PUT /v1/server/{uuid}/details` | `Server-UUID`, `Server-Secret` (no X-) | ✅ Yes |
| `POST /v2/metrics/comprehensive` | `X-Server-UUID`, `X-Server-Secret` | ✅ Yes |
| `POST /v2/hardware/inventory` | Unknown (needs testing) | ❓ |

### Evidence

```bash
# Heartbeat - Only works with X- prefix
curl -X POST "https://api-dev.nexmonyx.com/v1/heartbeat" \
  -H "X-Server-UUID: 182193a4-0e57-4f14-9d21-a8d41f860e38" \
  -H "X-Server-Secret: xxx"
# ✅ 200 OK

curl -X POST "https://api-dev.nexmonyx.com/v1/heartbeat" \
  -H "Server-UUID: 182193a4-0e57-4f14-9d21-a8d41f860e38" \
  -H "Server-Secret: xxx"
# ❌ 401 Unauthorized

# Server Details - Only works WITHOUT X- prefix
curl -X PUT "https://api-dev.nexmonyx.com/v1/server/182193a4-0e57-4f14-9d21-a8d41f860e38/details" \
  -H "Server-UUID: 182193a4-0e57-4f14-9d21-a8d41f860e38" \
  -H "Server-Secret: xxx"
# ✅ 200 OK

curl -X PUT "https://api-dev.nexmonyx.com/v1/server/182193a4-0e57-4f14-9d21-a8d41f860e38/details" \
  -H "X-Server-UUID: 182193a4-0e57-4f14-9d21-a8d41f860e38" \
  -H "X-Server-Secret: xxx"
# ❌ 401 "Invalid or missing authentication headers"
```

## Root Cause

Different routes appear to use different authentication middleware or header parsing logic:
- Some routes check for `X-Server-UUID` and `X-Server-Secret`
- Other routes check for `Server-UUID` and `Server-Secret`
- The recent fix to accept both formats may not have been applied to all endpoints

## Recommended Fix

### Option 1: Update ALL Endpoints to Accept Both Formats (Preferred)

Update the server authentication middleware to check both header formats:

```go
// In ServerCredentialAuthMiddleware or equivalent
serverUUID := r.Header.Get("X-Server-UUID")
if serverUUID == "" {
    serverUUID = r.Header.Get("Server-UUID")
}

serverSecret := r.Header.Get("X-Server-Secret")
if serverSecret == "" {
    serverSecret = r.Header.Get("Server-Secret")
}

if serverUUID == "" || serverSecret == "" {
    // Return 401 with clear error message
    return c.JSON(401, map[string]interface{}{
        "error": "unauthorized",
        "message": "Missing authentication headers. Provide either X-Server-UUID/X-Server-Secret or Server-UUID/Server-Secret",
    })
}
```

### Option 2: Standardize on One Format

Choose ONE format and update all endpoints to use it consistently:
- **Industry Standard**: Use `X-` prefix for custom headers (`X-Server-UUID`)
- **Alternative**: Use non-prefixed headers everywhere (`Server-UUID`)

Then update all route handlers to use the same middleware.

## Affected Code Locations

Based on the routes, check these handlers:
1. `servers_communication.ServerHeartbeat` - Currently expects X- headers
2. `registration_server.UpdateServerInfo` - Currently expects non-X headers
3. `servers_admin.UpdateServerHeartbeat` - Unknown
4. Any other endpoints using server authentication

## Testing Plan

After implementing the fix:

1. Test all endpoints with BOTH header formats
2. Ensure consistent behavior across all server-authenticated endpoints
3. Update API documentation to clarify accepted formats

## Impact

- **SDK**: Currently uses `X-` prefix headers (v1.1.3)
- **Agents**: Mixed - some may use X-, some may not
- **Priority**: HIGH - This inconsistency causes confusion and errors

## Success Criteria

All server-authenticated endpoints should:
1. Accept the same header format(s)
2. Return consistent error messages
3. Work with the current SDK (v1.1.3) which sends X- headers

## Timeline

This should be fixed ASAP as it's causing authentication failures for agents trying to update server details.

---

**Reported by**: SDK Team  
**Date**: 2025-07-08  
**Test Server**: 182193a4-0e57-4f14-9d21-a8d41f860e38