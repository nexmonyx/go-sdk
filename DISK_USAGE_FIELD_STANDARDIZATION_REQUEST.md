# CORRECTED: Disk Usage Field Mapping Analysis & Recommendation

**Date**: August 18, 2025 (Updated: August 19, 2025)  
**Priority**: Medium  
**Status**: Analysis Complete - Recommendation Provided

## Critical Issues Found in Original Request

After reviewing the Go SDK v2.5.1, several critical inaccuracies were found in the original standardization request that require correction before proceeding.

## Current State Analysis (CORRECTED)

### Actual SDK Field Names (3 Different Patterns Found):

1. **`DiskMetrics` struct** (`models.go:625`):
   ```go
   type DiskMetrics struct {
       UsagePercent float64 `json:"usage_percent"`  // NOT UsedPercent as originally stated
       // ... other fields
   }
   ```

2. **`TimescaleFilesystem` struct** (`models.go:763`):
   ```go
   type TimescaleFilesystem struct {
       UsedPercent float64 `json:"used_percent"`   // Different pattern: used_percent
       // ... other fields  
   }
   ```

3. **`DiskUsageAggregate` struct** (`models.go:637`):
   ```go
   type DiskUsageAggregate struct {
       UsedPercent float64 `json:"used_percent"`   // Also uses used_percent
       // ... other fields
   }
   ```

### Database Schema
- **Expected field**: `disk_usage_percent` (double precision)
- **Current value**: NULL (no mapping exists)

## Problem Summary
- **Agent sends**: `usage_percent` in disk metrics JSON
- **SDK has**: Both `usage_percent` AND `used_percent` field patterns
- **Database expects**: `disk_usage_percent`
- **API mapping**: Missing/broken

## Impact Assessment of Breaking Change Approach

### Structures Requiring Updates
If proceeding with standardization to `disk_usage_percent`:

1. **`DiskMetrics.UsagePercent`** → `DiskUsagePercent` (`models.go:625`)
2. **`TimescaleFilesystem.UsedPercent`** → `DiskUsagePercent` (`models.go:763`) 
3. **`DiskUsageAggregate.UsedPercent`** → `DiskUsagePercent` (`models.go:637`)
4. **All test files** referencing these fields
5. **All consuming applications** using these SDK structures

### Breaking Change Impact
- **Go code**: All references to `UsagePercent` field must change
- **JSON compatibility**: Changes from `usage_percent`/`used_percent` to `disk_usage_percent`
- **Agent coordination**: Linux agents must update simultaneously
- **Integration testing**: Comprehensive testing required across all components
- **Rollback complexity**: Difficult to rollback once deployed

## RECOMMENDATION: API Mapping Fix (Preferred)

Instead of the breaking change approach, we recommend **fixing the API mapping layer**:

### Approach
Update the Nexmonyx API to map existing field names to the database column:
- Map `usage_percent` → `disk_usage_percent` (from DiskMetrics)
- Map `used_percent` → `disk_usage_percent` (from TimescaleFilesystem)

### Benefits
- **No breaking changes** to SDK or agents
- **Backward compatible** with existing integrations
- **Lower deployment risk** - API-only changes
- **Immediate resolution** of NULL values issue
- **No agent coordination required**

### Implementation
Add field mapping logic in the comprehensive metrics API endpoint (`/v2/metrics/comprehensive`) to handle the field name translation during data processing.

## Alternative: SDK Standardization (High Risk)

If standardization is still preferred despite the risks:

### Required Changes
```go
// Update DiskMetrics (models.go:625)
type DiskMetrics struct {
    DiskUsagePercent float64 `json:"disk_usage_percent"`  // Changed from UsagePercent
    // ... other fields unchanged
}

// Update TimescaleFilesystem (models.go:763)  
type TimescaleFilesystem struct {
    DiskUsagePercent float64 `json:"disk_usage_percent"`  // Changed from UsedPercent
    // ... other fields unchanged
}

// Update DiskUsageAggregate (models.go:637)
type DiskUsageAggregate struct {
    DiskUsagePercent float64 `json:"disk_usage_percent"`  // Changed from UsedPercent  
    // ... other fields unchanged
}
```

### Deployment Requirements
1. **SDK v2.6.0** with updated field names
2. **API updates** to expect `disk_usage_percent` 
3. **Agent updates** to send `disk_usage_percent`
4. **Coordinated release** across all components
5. **Comprehensive integration testing**
6. **Rollback plan** in case of issues

## Recommendation to API Team

**Proceed with API mapping fix** rather than SDK standardization:

1. **Lower risk** solution
2. **Faster implementation** timeline  
3. **No coordination** dependencies
4. **Preserves** existing integrations
5. **Resolves** the immediate NULL values issue

The breaking change approach should only be considered if there are compelling architectural reasons that outweigh the significant deployment complexity and coordination requirements.