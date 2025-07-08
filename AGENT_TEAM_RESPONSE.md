# Authentication Issue RESOLVED - SDK v1.1.2 Released

## Great News! 🎉

The authentication issue blocking Linux agents has been completely resolved through coordinated fixes in both the SDK and API.

## What Happened

### Root Cause
- SDK was sending headers as `X-Server-UUID` and `X-Server-Secret`
- API expected headers as `Server-UUID` and `Server-Secret` (without X- prefix)
- This mismatch caused "Missing Server-UUID or Server-Secret headers" errors

### Solution Implemented

1. **SDK v1.1.2** (Released Now)
   - Updated to send headers without X- prefix
   - Available at: https://github.com/nexmonyx/go-sdk/releases/tag/v1.1.2
   
2. **API Update** (Already Deployed)
   - Now accepts BOTH header formats for backwards compatibility
   - No breaking changes - all existing agents continue to work

## For Your Linux Agents

### Existing Agents (Using SDK v1.1.1 or earlier)
✅ **No immediate action required** - They will continue to work thanks to API backwards compatibility

### New Deployments
✅ **Use SDK v1.1.2** - Automatically uses the correct header format

### Upgrading Existing Agents
When convenient, update your agent's go.mod:
```bash
go get github.com/nexmonyx/go-sdk@v1.1.2
go mod tidy
```

## Testing Your Agents

We've added debug capabilities to help verify authentication:

### 1. Enable Debug Mode
```go
sdkConfig := &nexmonyx.Config{
    BaseURL: config.CFG.Endpoint,
    Auth: nexmonyx.AuthConfig{
        ServerUUID:   config.CFG.ServerUUID,
        ServerSecret: config.CFG.ServerSecret,
    },
    Debug: true,  // Shows headers being sent
}
```

### 2. Use Test Scripts
```bash
# Download and run from SDK repo
./test_headers.sh <SERVER_UUID> <SERVER_SECRET> https://api-dev.nexmonyx.com
```

### 3. Verify in Agent Logs
With SDK v1.1.2, you should see successful:
- ✅ Heartbeats
- ✅ Metrics submission
- ✅ Server detail updates
- ✅ Hardware inventory submission

## What This Means

1. **All agent operations now work correctly**
2. **Zero downtime** - Existing agents kept working during the fix
3. **Future-proof** - Clear migration path established
4. **Better debugging** - Enhanced tools for troubleshooting

## Timeline Recap

- **2025-07-08 09:00**: Issue reported - agents failing authentication
- **2025-07-08 14:30**: Root cause identified - header format mismatch
- **2025-07-08 15:00**: SDK v1.1.2 released with fix
- **2025-07-08 15:30**: API updated to accept both formats
- **2025-07-08 16:00**: Issue fully resolved

## Support

If you encounter any issues:
1. Ensure you're using SDK v1.1.2 or later
2. Enable debug mode to see authentication headers
3. Check the [Authentication Resolution Guide](https://github.com/nexmonyx/go-sdk/blob/main/AUTHENTICATION_RESOLUTION.md)
4. Open an issue at https://github.com/nexmonyx/go-sdk/issues

## Thank You

Thank you for your patience while we resolved this issue. The coordination between the SDK and API teams ensured a smooth resolution with zero downtime for existing deployments.

Your Linux agents should now be fully operational! 🚀

---

**SDK Team**  
**Release**: [v1.1.2](https://github.com/nexmonyx/go-sdk/releases/tag/v1.1.2)  
**Date**: 2025-07-08