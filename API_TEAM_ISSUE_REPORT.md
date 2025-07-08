# Critical API Authentication Inconsistency - Action Required

## Executive Summary

The Nexmonyx API has a critical authentication header inconsistency that is blocking all Linux agents from submitting metrics and heartbeats. The API expects different header formats across different endpoints, causing authentication failures.

## Issue Details

### Problem
- **v2 endpoints** expect headers WITHOUT the "X-" prefix: `Server-UUID`, `Server-Secret`
- **v1 endpoints** are returning 500 errors regardless of header format
- Error messages are misleading (e.g., "Missing Server-UUID or Server-Secret headers" when headers are present with X- prefix)
- This affects ALL agent operations: heartbeats, metrics submission, server updates

### Test Results

Using server credentials:
- UUID: `182193a4-0e57-4f14-9d21-a8d41f860e38`
- Secret: `2225e0a23fa038625e30d7696c961df7ed79ec9496b3e1f2748886e7bfbfb7f8`
- API: `https://api-dev.nexmonyx.com`

| Endpoint | X-Server-UUID Headers | Server-UUID Headers | Expected Result |
|----------|----------------------|---------------------|-----------------|
| POST /v1/servers/heartbeat | ❌ 500 Error | ❌ 500 Error | Should work with one format |
| PUT /v1/servers/{uuid}/details | ❌ 500 Error | ❌ 500 Error | Should work with one format |
| POST /v1/metrics/comprehensive | ❌ 500 Error | ❌ 500 Error | Should work with one format |
| POST /v2/metrics/comprehensive | ❌ 401 "Missing Server-UUID" | ✅ 200 Success | Working as expected |
| POST /v2/hardware/inventory | ❌ 500 Error | ❌ 500 Error | Should work with one format |

### SDK Changes Made

The Go SDK v1.1.2 has been updated to send headers WITHOUT the X- prefix:
```go
// Changed from:
restyClient.SetHeader("X-Server-UUID", config.Auth.ServerUUID)
restyClient.SetHeader("X-Server-Secret", config.Auth.ServerSecret)

// To:
restyClient.SetHeader("Server-UUID", config.Auth.ServerUUID)
restyClient.SetHeader("Server-Secret", config.Auth.ServerSecret)
```

This fixes v2 endpoints but v1 endpoints are still broken.

## Root Cause Analysis Required

Please investigate:

1. **Authentication Middleware**
   - Why do v2 endpoints expect `Server-UUID` while the SDK historically used `X-Server-UUID`?
   - Why are v1 endpoints returning 500 errors instead of 401?
   - Is there different authentication middleware for v1 vs v2 endpoints?

2. **Server Status**
   - Is server `182193a4-0e57-4f14-9d21-a8d41f860e38` active and valid?
   - Are there any flags or settings that might be causing the 500 errors?
   - Check server logs for the actual error when these requests fail

3. **Environment Differences**
   - Does production API have the same behavior as dev?
   - Are there different authentication requirements between environments?

## Immediate Actions Needed

### Option 1: Fix API to Accept Consistent Headers (Recommended)
Update all endpoints to accept BOTH header formats for backwards compatibility:
- Accept both `X-Server-UUID` and `Server-UUID`
- Accept both `X-Server-Secret` and `Server-Secret`

Example middleware update:
```python
def get_server_auth(headers):
    # Check both header formats
    server_uuid = headers.get('Server-UUID') or headers.get('X-Server-UUID')
    server_secret = headers.get('Server-Secret') or headers.get('X-Server-Secret')
    
    if not server_uuid or not server_secret:
        return None, "Missing Server-UUID or Server-Secret headers"
    
    return server_uuid, server_secret
```

### Option 2: Fix API to Use One Consistent Format
Choose ONE format and update all endpoints to use it consistently:
- Either use `X-Server-UUID` everywhere (industry standard for custom headers)
- Or use `Server-UUID` everywhere (current v2 behavior)

### Option 3: Fix the 500 Errors First
Before addressing header consistency, fix why v1 endpoints return 500:
1. Check API logs for the actual error
2. Verify the test server is properly configured
3. Test with a fresh server registration

## Test Scripts Provided

We've created test scripts in the SDK repository to help debug:

1. **test_headers.sh** - Tests specific endpoints with both header formats
2. **test_all_endpoints.sh** - Comprehensive test of all agent endpoints
3. **auth_debug.go** - Go utility to test authentication

Run these from the SDK directory:
```bash
# Test all endpoints
./test_all_endpoints.sh <SERVER_UUID> <SERVER_SECRET> <API_URL>

# Test specific endpoint
./test_headers.sh <SERVER_UUID> <SERVER_SECRET> <API_URL>
```

## Impact

- **Severity**: CRITICAL
- **Affected**: ALL Linux agents using SDK v1.1.1 or earlier (and partially v1.1.2)
- **Business Impact**: No monitoring data can be collected
- **User Reports**: Multiple agents failing with "Missing Server-UUID or Server-Secret headers"

## Success Criteria

1. All agent endpoints accept the same header format
2. Clear error messages that match the actual expected headers
3. No 500 errors for valid authentication attempts
4. Documentation updated to specify the correct header format

## Questions for API Team

1. What is the intended header format for server authentication?
2. Why do v1 and v2 endpoints have different authentication handling?
3. Can you provide the actual error logs for the 500 responses?
4. Is there a migration in progress from X- prefixed headers to non-prefixed?
5. Are there any server-side flags or settings affecting this test server?

## Timeline

This is blocking all agent deployments. Please investigate and provide an update within 24 hours.

---

**Contact**: SDK Team
**Date**: 2025-07-08
**SDK Version**: 1.1.2 (updated to remove X- prefix)
**Test Environment**: api-dev.nexmonyx.com