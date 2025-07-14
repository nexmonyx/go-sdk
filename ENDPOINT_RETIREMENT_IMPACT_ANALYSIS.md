# Endpoint Retirement Impact Analysis

## Overview

The proposed retirement of the following server management endpoints will have **MAJOR IMPACT** on the Nexmonyx Go SDK:

### Endpoints Being Retired:
```
GET    /v1/server                  (Get all servers)
POST   /v1/server                  (Create a new server)
GET    /v1/server/:uuid            (Get a server by UUID)
PUT    /v1/server/:uuid            (Update a server by UUID)
PATCH  /v1/server/:uuid            (Partially update a server by UUID)
DELETE /v1/server/:uuid            (Delete a server by UUID)
GET    /v1/servers/:uuid/disk-summary (Get server disk usage summary)
```

## Critical Impact Assessment

### 🔴 **HIGH IMPACT - BREAKING CHANGES**

The SDK has **extensive usage** of these retired endpoints. Nearly all core server management functionality will be broken.

## Affected SDK Methods and Files

### 1. **servers.go** - Core Server Management (21 affected methods)

#### **Direct Matches to Retired Endpoints:**

| SDK Method | Current Endpoint | Retirement Status | Line |
|------------|-----------------|-------------------|------|
| `Get()` | `GET /v1/server/{id}` | ❌ **RETIRED** | 15 |
| `GetByUUID()` | `GET /v1/server/uuid/{uuid}` | ❌ **RETIRED** | 35 |
| `Update()` | `PUT /v1/server/{id}` | ❌ **RETIRED** | 100 |
| `Delete()` | `DELETE /v1/server/{id}` | ❌ **RETIRED** | 120 |
| `Register()` | `POST /v1/server/register` | ❌ **RETIRED** | 138 |
| `RegisterWithKeyFull()` | `POST /v1/server/register` | ❌ **RETIRED** | 296 |
| `UpdateServer()` | `PUT /v1/server/{uuid}` | ❌ **RETIRED** | 394 |

#### **Additional /v1/server Endpoints Also at Risk:**

| SDK Method | Current Endpoint | Risk Level | Line |
|------------|-----------------|------------|------|
| `UpdateHeartbeat()` | `PUT /v1/server/{uuid}/heartbeat` | 🟡 **HIGH RISK** | 158 |
| `GetMetrics()` | `GET /v1/server/{id}/metrics` | 🟡 **HIGH RISK** | 172 |
| `GetAlerts()` | `GET /v1/server/{id}/alerts` | 🟡 **HIGH RISK** | 196 |
| `UpdateTags()` | `PUT /v1/server/{id}/tags` | 🟡 **HIGH RISK** | 223 |
| `ExecuteCommand()` | `POST /v1/server/{id}/execute` | 🟡 **HIGH RISK** | 249 |
| `GetSystemInfo()` | `GET /v1/server/{id}/system-info` | 🟡 **HIGH RISK** | 267 |
| `UpdateDetails()` | `PUT /v1/server/{uuid}/details` | 🟡 **HIGH RISK** | 410 |
| `UpdateInfo()` | `PUT /v1/server/{uuid}/info` | 🟡 **HIGH RISK** | 470 |
| `GetDetails()` | `GET /v1/server/{uuid}/details` | 🟡 **HIGH RISK** | 535 |
| `GetFullDetails()` | `GET /v1/server/{uuid}/full-details` | 🟡 **HIGH RISK** | 556 |
| `GetHeartbeat()` | `GET /v1/server/{uuid}/heartbeat` | 🟡 **HIGH RISK** | 606 |

#### **Safe Endpoints (Currently using /v1/servers):**

| SDK Method | Current Endpoint | Status | Line |
|------------|-----------------|--------|------|
| `List()` | `GET /v1/servers` | ✅ **SAFE** | 56 |
| `Create()` | `POST /v1/servers` | ✅ **SAFE** | 79 |

### 2. **auth_debug.go** - Authentication Testing

| Function | Current Endpoint | Impact | Line |
|----------|-----------------|--------|------|
| `testAuthRequest()` | `POST /v1/server/heartbeat` | ❌ **BROKEN** | 95 |

### 3. **service_monitoring_api.go** - Service Monitoring (7 endpoints)

| SDK Method | Current Endpoint | Status | Line |
|------------|-----------------|--------|------|
| `SubmitServiceData()` | `POST /v1/servers/{uuid}/services` | ✅ **SAFE** | 72 |
| `SubmitServiceMetrics()` | `POST /v1/servers/{uuid}/services/metrics` | ✅ **SAFE** | 88 |
| `SubmitServiceLogs()` | `POST /v1/servers/{uuid}/services/logs` | ✅ **SAFE** | 104 |
| `GetServerServices()` | `GET /v1/servers/{uuid}/services` | ✅ **SAFE** | 115 |
| `GetServiceHistory()` | `GET /v1/servers/{uuid}/services/{name}/history` | ✅ **SAFE** | 132 |
| `RestartService()` | `POST /v1/servers/{uuid}/services/{name}/restart` | ✅ **SAFE** | 155 |
| `GetServiceLogs()` | `GET /v1/servers/{uuid}/services/{name}/logs` | ✅ **SAFE** | 165 |

### 4. **Missing Implementation**

| Retired Endpoint | Implementation Status |
|------------------|----------------------|
| `GET /v1/servers/:uuid/disk-summary` | ❌ **NOT IMPLEMENTED** in SDK |

## Required Changes

### **Immediate Actions Required:**

1. **Map Replacement Endpoints**: Determine what the new endpoints will be
   - Are they moving to `/v1/servers/{uuid}` pattern?
   - Are they moving to a completely different path?
   - Are they being consolidated into other endpoints?

2. **Update SDK Methods**: All affected methods need endpoint changes

3. **Update Authentication Testing**: Fix `auth_debug.go` heartbeat endpoint

4. **Add Missing Functionality**: Implement disk summary endpoint if replacement exists

### **Likely Migration Patterns:**

Based on the existing `/v1/servers` endpoints that are safe, the migration will likely be:

```
OLD: GET    /v1/server/{uuid}           → NEW: GET    /v1/servers/{uuid}
OLD: PUT    /v1/server/{uuid}           → NEW: PUT    /v1/servers/{uuid}
OLD: DELETE /v1/server/{uuid}           → NEW: DELETE /v1/servers/{uuid}
OLD: GET    /v1/server                  → NEW: GET    /v1/servers (already exists)
OLD: POST   /v1/server                  → NEW: POST   /v1/servers (already exists)
```

For sub-endpoints:
```
OLD: PUT /v1/server/{uuid}/heartbeat    → NEW: PUT /v1/servers/{uuid}/heartbeat
OLD: GET /v1/server/{uuid}/details      → NEW: GET /v1/servers/{uuid}/details
OLD: PUT /v1/server/{uuid}/details      → NEW: PUT /v1/servers/{uuid}/details
etc.
```

### **SDK Update Strategy:**

1. **Create Migration Constants**:
```go
const (
    // Deprecated endpoints (to be removed)
    deprecatedServerPath = "/v1/server"
    
    // New endpoints
    serverPath = "/v1/servers"
)
```

2. **Update All Methods Systematically**:
```go
// OLD
Path: fmt.Sprintf("/v1/server/%s", id)

// NEW  
Path: fmt.Sprintf("/v1/servers/%s", id)
```

3. **Add Backwards Compatibility Warning**:
```go
// Deprecated: Use GetByUUID instead
func (s *ServersService) Get(ctx context.Context, id string) (*Server, error) {
    // Add deprecation warning
    return s.GetByUUID(ctx, id)
}
```

### **Testing Impact:**

All server-related tests will need updates:
- Unit tests in `servers_test.go`
- Integration tests
- Example code
- Documentation

### **Breaking Change Assessment:**

This is a **MAJOR BREAKING CHANGE** that affects:
- ✅ 7+ core server management methods
- ✅ 14+ additional server sub-endpoint methods  
- ✅ Authentication testing functionality
- ✅ All client code using server management

### **Rollout Recommendations:**

1. **Phase 1**: Implement new endpoints alongside old ones
2. **Phase 2**: Update SDK to use new endpoints with deprecation warnings
3. **Phase 3**: Remove old endpoint support after grace period
4. **Version**: This requires a **MAJOR** version bump (v2.x.x)

## Files Requiring Updates

1. **servers.go** - 21 methods need endpoint changes
2. **auth_debug.go** - 1 endpoint needs update
3. **Documentation** - All server endpoint docs need updates
4. **Examples** - All server-related examples need updates
5. **Tests** - All server tests need endpoint updates
6. **CHANGELOG.md** - Document breaking changes

## Risk Mitigation

1. **Dual Support Period**: Support both old and new endpoints during transition
2. **Clear Migration Guide**: Provide step-by-step migration instructions
3. **Automated Migration Tools**: Consider providing migration scripts
4. **Extended Deprecation Period**: Give users time to migrate
5. **Communication**: Clear advance notice to all SDK users

## Conclusion

This endpoint retirement represents a **CRITICAL BREAKING CHANGE** that affects nearly every core server management function in the SDK. A comprehensive migration plan with careful versioning and backwards compatibility considerations is essential.