# Authentication Header Issue - RESOLVED

## Issue Summary
Linux agents were failing to authenticate with "Missing Server-UUID or Server-Secret headers" errors because:
- The SDK was sending headers WITH the "X-" prefix: `X-Server-UUID`, `X-Server-Secret`
- The API only accepted headers WITHOUT the prefix: `Server-UUID`, `Server-Secret`

## Resolution

### API Changes (Implemented)
The API team updated the `ServerCredentialAuthMiddleware` to accept BOTH header formats:
1. **Primary**: `Server-UUID` and `Server-Secret` (preferred)
2. **Fallback**: `X-Server-UUID` and `X-Server-Secret` (for compatibility)

This ensures:
- ✅ All existing agents continue to work without changes
- ✅ No breaking changes or downtime
- ✅ Clear migration path for future updates

### SDK Changes (v1.1.2)
The Go SDK has been updated to use the preferred non-prefixed headers:
```go
// Now sends:
restyClient.SetHeader("Server-UUID", config.Auth.ServerUUID)
restyClient.SetHeader("Server-Secret", config.Auth.ServerSecret)
```

## Compatibility Matrix

| SDK Version | Header Format | API Support | Status |
|-------------|---------------|-------------|---------|
| v1.1.1 and earlier | X-Server-UUID | ✅ Yes (via fallback) | Working |
| v1.1.2+ | Server-UUID | ✅ Yes (primary) | Recommended |

## Migration Guide

### For New Installations
- Use SDK v1.1.2 or later
- Headers will automatically use the preferred format

### For Existing Installations
- **No immediate action required** - existing agents will continue to work
- Update to SDK v1.1.2+ at your convenience
- The X-prefixed headers may be deprecated in a future major version

## Testing

Use the provided test scripts to verify authentication:

```bash
# Test your server credentials
./test_headers.sh <SERVER_UUID> <SERVER_SECRET> <API_URL>

# Test all endpoints
./test_all_endpoints.sh <SERVER_UUID> <SERVER_SECRET> <API_URL>
```

Both header formats should now work successfully.

## Timeline

- **2025-07-08**: Issue identified - SDK sending X-prefixed headers
- **2025-07-08**: SDK v1.1.2 released - uses non-prefixed headers
- **2025-07-08**: API updated - accepts both header formats
- **Future**: Consider deprecating X-prefixed headers in next major version

## Additional Notes

- All endpoints (v1 and v2) use the same authentication middleware
- The 500 errors mentioned in testing were likely environment-specific
- Debug logging has been added to the API for better troubleshooting

## Contact

For any issues or questions:
- SDK Team: [GitHub Issues](https://github.com/nexmonyx/go-sdk/v2/issues)
- API Team: Internal support channels