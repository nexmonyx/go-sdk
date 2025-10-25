# Alert Controller - API-First Refactoring Strategy

**Document Version**: 1.0
**Last Updated**: 2025-10-23
**Status**: Draft
**Related Documents**:
- [alert-controller-p1-architecture.md](./alert-controller-p1-architecture.md) - Architecture & Current State
- [alert-controller-implementation-plan.md](./alert-controller-implementation-plan.md) - Implementation Roadmap
- [alert-controller-integration-spec.md](./alert-controller-integration-spec.md) - Integration Contracts

---

## Executive Summary

This document outlines the refactoring strategy to bring the alert-controller into compliance with Nexmonyx's **API-first architecture mandate**. The current implementation contains **6 direct database access violations** that bypass the API layer, creating architectural debt and maintenance challenges.

**Problem Statement**:
The alert-controller directly accesses PostgreSQL database using GORM, violating the architectural principle that **all services must communicate exclusively through the API layer using the official go-sdk**.

**Impact**:
- ❌ Bypasses API server's authentication, authorization, and audit logging
- ❌ Duplicates business logic between controller and API server
- ❌ Prevents leveraging API server's caching, rate limiting, and optimization
- ❌ Creates tight coupling to database schema
- ❌ Complicates testing and mocking
- ❌ Blocks future API server horizontal scaling

**Solution**:
- ✅ Remove all direct database access from alert-controller
- ✅ Enhance go-sdk v2.4.0 with 8 new methods for alert operations
- ✅ Implement corresponding API endpoints in API server
- ✅ Refactor alert-controller to use SDK exclusively
- ✅ Achieve 100% API-first compliance with zero breaking changes

**Timeline**: 4 weeks (phased implementation)
**Effort**: 18 senior engineer days + 4 DevOps days + 3 QA days

---

## Table of Contents

1. [Problem Statement](#problem-statement)
2. [Current State Analysis](#current-state-analysis)
3. [Target Architecture](#target-architecture)
4. [SDK Enhancement Requirements](#sdk-enhancement-requirements)
5. [API Server Enhancements](#api-server-enhancements)
6. [Controller Refactoring Strategy](#controller-refactoring-strategy)
7. [Migration Phases](#migration-phases)
8. [Performance Analysis](#performance-analysis)
9. [Risk Assessment](#risk-assessment)
10. [Success Criteria](#success-criteria)

---

## Problem Statement

### Architecture Mandate

**Nexmonyx Platform Architecture Principle**:
> All microservices, controllers, and agents MUST communicate exclusively through the Nexmonyx API Server using the official go-sdk. Direct database access is PROHIBITED except within the API server itself.

**Why This Matters**:

1. **Centralized Business Logic**: All validation, authorization, and data transformation happens in one place (API server)
2. **Audit Trail**: Every data access is logged through API endpoints for compliance
3. **Consistent Security**: API server enforces organization scoping, rate limiting, and authentication
4. **Scalability**: API server can be horizontally scaled without controller changes
5. **Testing**: Controllers can be tested against mock API server without database
6. **Caching**: API server provides intelligent caching, reducing database load

### Current Violations

The alert-controller currently violates this principle in **6 locations**:

| Location | Violation Type | Severity | Lines |
|----------|---------------|----------|-------|
| `cmd/main.go` | Database connection initialization | CRITICAL | 87-103 |
| `cmd/main.go` | Direct alert rules query | HIGH | 156-168 |
| `pkg/evaluator/evaluator.go` | Server list query | HIGH | 94-112 |
| `pkg/evaluator/evaluator.go` | Alert instance creation | HIGH | 178-196 |
| `pkg/evaluator/metrics.go` | Metrics aggregation query | MEDIUM | 45-67 |
| `pkg/evaluator/metrics.go` | Historical metrics query | MEDIUM | 123-142 |

**Evidence**: See [alert-controller-p1-architecture.md](./alert-controller-p1-architecture.md) Section 3 for detailed code analysis.

### Impact Assessment

**Technical Debt**:
- 4,127 lines of code affected (58% of total codebase)
- 34 function signatures require modification
- 156 unit tests need updating (database mocks → API mocks)

**Operational Risk**:
- Security bypass: Database access not subject to API-level authorization
- Compliance gap: No audit trail for data access
- Performance: Cannot leverage API server's query optimization and caching

**Future Constraints**:
- Blocks API server horizontal scaling (shared database connection pool)
- Prevents migration to cloud-managed API gateway
- Limits observability (no centralized API metrics)

---

## Current State Analysis

### Violation #1: Database Connection Initialization (CRITICAL)

**File**: `cmd/alert-controller/main.go:87-103`

**Current Code**:
```go
func initDatabase(config *Config) (*gorm.DB, error) {
    dsn := fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s search_path=%s sslmode=disable",
        config.DBHost,
        config.DBPort,
        config.DBUser,
        config.DBPassword,
        config.DBName,
        fmt.Sprintf("org_%d", config.OrganizationID),
    )

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Silent),
    })
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }

    return db, nil
}
```

**Problems**:
1. ❌ Creates direct PostgreSQL connection
2. ❌ Bypasses API server entirely
3. ❌ Duplicates connection pooling logic
4. ❌ Requires database credentials in controller
5. ❌ Cannot leverage API server's connection management

**Root Cause**: Alert-controller was initially developed before API-first mandate

**Blast Radius**: Entire controller depends on this database connection

---

### Violation #2: Direct Alert Rules Query (HIGH)

**File**: `cmd/alert-controller/main.go:156-168`

**Current Code**:
```go
func loadAlertRules(db *gorm.DB, orgID uint) ([]*models.AlertRule, error) {
    var rules []*models.AlertRule

    err := db.Where("organization_id = ? AND enabled = ?", orgID, true).
        Preload("Channels").
        Find(&rules).Error

    if err != nil {
        return nil, fmt.Errorf("failed to load alert rules: %w", err)
    }

    log.Printf("Loaded %d alert rules for organization %d", len(rules), orgID)
    return rules, nil
}
```

**Problems**:
1. ❌ Direct GORM query bypasses API
2. ❌ No API-level authorization check (trusts organization_id from config)
3. ❌ Preload logic duplicates API server's relationship loading
4. ❌ Cannot leverage API server's query optimization
5. ❌ No caching (API server could cache rules for 5 minutes)

**Business Logic Duplication**:
- API server already has `GET /api/v1/alerts/rules` endpoint
- Duplicates filtering, preloading, and response formatting

**Performance Impact**:
- Database query: ~50-200ms
- API call would be: ~60-250ms (similar + network latency)
- **With caching**: 5-10ms (90% reduction in database load)

---

### Violation #3: Server List Query (HIGH)

**File**: `pkg/evaluator/evaluator.go:94-112`

**Current Code**:
```go
func (e *Evaluator) getServersInScope(rule *models.AlertRule) ([]*models.Server, error) {
    var servers []*models.Server

    query := e.db.Where("organization_id = ?", rule.OrganizationID)

    // Apply server filters from rule
    if rule.ServerFilters != nil {
        if tags, ok := rule.ServerFilters["tags"].([]string); ok && len(tags) > 0 {
            query = query.Where("tags @> ?", pq.Array(tags))
        }
        if environment, ok := rule.ServerFilters["environment"].(string); ok {
            query = query.Where("environment = ?", environment)
        }
    }

    err := query.Find(&servers).Error
    return servers, err
}
```

**Problems**:
1. ❌ Direct database query with complex filtering
2. ❌ Duplicates server filtering logic from API server
3. ❌ PostgreSQL-specific syntax (`tags @>`) locks to specific database
4. ❌ No validation of filter syntax (API server validates)
5. ❌ Cannot leverage API server's server caching

**Business Logic Duplication**:
- API server's `ServerService.List()` already implements this filtering
- Tag-based filtering logic exists in API server
- Environment scoping exists in API server

**Missing Features**:
- API server supports additional filters: location, classification, status
- API server handles pagination for large server lists
- API server validates filter syntax and provides better error messages

---

### Violation #4: Alert Instance Creation (HIGH)

**File**: `pkg/evaluator/evaluator.go:178-196`

**Current Code**:
```go
func (e *Evaluator) createAlertInstance(rule *models.AlertRule, server *models.Server, value float64, severity string) error {
    instance := &models.AlertInstance{
        OrganizationID: rule.OrganizationID,
        RuleID:         rule.ID,
        ServerID:       server.ID,
        State:          "firing",
        Severity:       severity,
        Value:          value,
        FiredAt:        time.Now(),
        Message:        fmt.Sprintf("Alert '%s' triggered on server '%s'", rule.Name, server.Hostname),
    }

    if err := e.db.Create(instance).Error; err != nil {
        return fmt.Errorf("failed to create alert instance: %w", err)
    }

    log.Printf("Created alert instance %d for rule %d on server %d", instance.ID, rule.ID, server.ID)
    return nil
}
```

**Problems**:
1. ❌ Direct database INSERT bypasses API
2. ❌ No validation of severity values (API server validates)
3. ❌ No duplicate detection (API server should check for existing firing instance)
4. ❌ No audit logging (API server logs all instance creations)
5. ❌ Message formatting duplicates API server logic

**Missing Validations** (API server would provide):
- Verify rule exists and is enabled
- Verify server exists and belongs to organization
- Check for duplicate firing instances (same rule + server)
- Validate severity against rule's threshold configuration
- Rate limiting (prevent alert storms)

**Audit Trail Gap**:
- API server logs: WHO created instance, WHEN, from WHERE (IP/service)
- Direct DB insert: No attribution, no service tracking

---

### Violation #5: Metrics Aggregation Query (MEDIUM)

**File**: `pkg/evaluator/metrics.go:45-67`

**Current Code**:
```go
func (e *Evaluator) getLatestMetricValue(serverID uint, metricName string) (float64, error) {
    var result struct {
        Value float64
    }

    query := `
        SELECT
            CASE
                WHEN ? = 'cpu_usage' THEN cpu_usage
                WHEN ? = 'memory_usage' THEN memory_usage
                WHEN ? = 'disk_usage' THEN disk_usage
                ELSE NULL
            END as value
        FROM cpu_metrics
        WHERE server_id = ?
        ORDER BY created_at DESC
        LIMIT 1
    `

    err := e.db.Raw(query, metricName, metricName, metricName, serverID).Scan(&result).Error
    return result.Value, err
}
```

**Problems**:
1. ❌ Raw SQL query (fragile, database-specific)
2. ❌ Hardcoded metric names (inflexible)
3. ❌ No query caching (repeated calls for same server)
4. ❌ SQL injection risk if metricName ever becomes user input
5. ❌ Cannot leverage API server's TimescaleDB optimization

**Business Logic Duplication**:
- API server's `MetricsService.Query()` handles all metric types
- API server uses TimescaleDB continuous aggregates (faster)
- API server caches recent metrics (reduces database load)

**Performance Impact**:
- Current: 5-15ms per query × 5,000 servers = 25-75 seconds total
- API-first with caching: 2-5ms per call × 5,000 servers = 10-25 seconds (60% reduction)

---

### Violation #6: Historical Metrics Query (MEDIUM)

**File**: `pkg/evaluator/metrics.go:123-142`

**Current Code**:
```go
func (e *Evaluator) getAverageMetricValue(serverID uint, metricName string, duration time.Duration) (float64, error) {
    var result struct {
        AvgValue float64
    }

    query := `
        SELECT AVG(
            CASE
                WHEN ? = 'cpu_usage' THEN cpu_usage
                WHEN ? = 'memory_usage' THEN memory_usage
                ELSE NULL
            END
        ) as avg_value
        FROM cpu_metrics
        WHERE server_id = ? AND created_at >= ?
    `

    since := time.Now().Add(-duration)
    err := e.db.Raw(query, metricName, metricName, serverID, since).Scan(&result).Error

    return result.AvgValue, err
}
```

**Problems**:
1. ❌ Raw SQL with complex CASE logic
2. ❌ No support for other aggregations (MAX, MIN, SUM)
3. ❌ Inefficient for large time ranges (no continuous aggregates)
4. ❌ Cannot leverage TimescaleDB time-bucket optimization
5. ❌ No caching for expensive aggregations

**API Server Advantages**:
- Uses TimescaleDB's `time_bucket()` for efficient aggregation
- Supports continuous aggregates (pre-computed hourly/daily averages)
- Caches aggregation results (5-minute TTL)
- Supports all aggregation types: AVG, MAX, MIN, SUM, P50, P95, P99

**Performance Comparison**:
```
Raw query (current):     100-500ms (full table scan)
API with time_bucket:    20-50ms (TimescaleDB optimization)
API with cont. agg:      5-10ms (pre-computed aggregates)
```

---

### Summary of Violations

**Severity Breakdown**:
- 1 CRITICAL: Database connection initialization (blocks entire refactoring)
- 3 HIGH: Alert rules, servers, alert instances (core functionality)
- 2 MEDIUM: Metrics queries (performance impact)

**Total Code Impact**:
- 6 functions require complete rewrite
- 34 function signatures require modification
- 156 unit tests require updating
- 4,127 lines of code affected (58% of total)

**Dependencies**:
- Violation #1 must be fixed first (foundation)
- Violations #2-6 can be fixed in parallel after #1

---

## Target Architecture

### API-First Architecture Pattern

**Principle**: All data access flows through the API Server using the official go-sdk

```
┌─────────────────────────────────────────┐
│        Alert Controller                 │
│  ┌───────────────────────────────────┐  │
│  │     Business Logic Layer          │  │
│  │  • Alert evaluation engine        │  │
│  │  • Threshold checking             │  │
│  │  • Lifecycle management           │  │
│  └───────────┬───────────────────────┘  │
│              │                           │
│              ▼                           │
│  ┌───────────────────────────────────┐  │
│  │       go-sdk Client               │  │
│  │  • Authentication                 │  │
│  │  • Retry logic                    │  │
│  │  • Error handling                 │  │
│  │  • Request/response parsing       │  │
│  └───────────┬───────────────────────┘  │
└──────────────┼───────────────────────────┘
               │ HTTP/HTTPS
               │ (API Key + Secret)
               ▼
┌─────────────────────────────────────────┐
│        Nexmonyx API Server              │
│  ┌───────────────────────────────────┐  │
│  │     API Endpoints Layer           │  │
│  │  • Authentication middleware      │  │
│  │  • Authorization checks           │  │
│  │  • Request validation             │  │
│  │  • Response formatting            │  │
│  └───────────┬───────────────────────┘  │
│              ▼                           │
│  ┌───────────────────────────────────┐  │
│  │     Business Logic Layer          │  │
│  │  • AlertService                   │  │
│  │  • ServerService                  │  │
│  │  • MetricsService                 │  │
│  │  • Caching, optimization          │  │
│  └───────────┬───────────────────────┘  │
│              ▼                           │
│  ┌───────────────────────────────────┐  │
│  │     Data Access Layer (GORM)      │  │
│  │  • Database queries               │  │
│  │  • Transaction management         │  │
│  │  • Connection pooling             │  │
│  └───────────┬───────────────────────┘  │
└──────────────┼───────────────────────────┘
               │
               ▼
       ┌───────────────┐
       │   PostgreSQL  │
       │  (org_123)    │
       └───────────────┘
```

**Benefits**:
1. ✅ **Single Source of Truth**: All business logic in API server
2. ✅ **Security**: API server enforces authentication, authorization, rate limiting
3. ✅ **Observability**: All requests logged, metrics collected
4. ✅ **Scalability**: API server can scale horizontally without controller changes
5. ✅ **Testing**: Controllers testable with mock API server
6. ✅ **Caching**: API server provides intelligent query caching
7. ✅ **Consistency**: All services use same API contracts

### Component Responsibilities

**Alert Controller** (Client of API):
- Alert evaluation logic (threshold checking)
- Orchestration of evaluation cycles
- Notification delivery coordination
- **NO direct database access**
- **NO business logic that belongs in API**

**API Server** (Owner of Data):
- All database access (CRUD operations)
- Authentication and authorization
- Input validation and sanitization
- Business logic (duplicate detection, rate limiting)
- Query optimization and caching
- Audit logging

**go-sdk** (Abstraction Layer):
- HTTP client management
- Request/response serialization
- Retry logic with exponential backoff
- Error handling and type conversion
- Authentication header injection

---

## SDK Enhancement Requirements

### Required go-sdk v2.4.0 Methods

The go-sdk must be enhanced with **8 new methods** to support alert-controller's needs:

#### 1. ListInstances - List Alert Instances

```go
// ListInstances retrieves alert instances with filtering and pagination
func (s *AlertsService) ListInstances(ctx context.Context, opts *ListInstancesOptions) ([]*AlertInstance, *PaginationMeta, error)

type ListInstancesOptions struct {
    OrganizationID uint              `json:"organization_id"`
    RuleID         *uint             `json:"rule_id,omitempty"`
    ServerID       *uint             `json:"server_id,omitempty"`
    State          *string           `json:"state,omitempty"`  // firing, acknowledged, resolved, silenced
    Severity       *string           `json:"severity,omitempty"` // info, warning, critical
    FiredAfter     *time.Time        `json:"fired_after,omitempty"`
    FiredBefore    *time.Time        `json:"fired_before,omitempty"`
    Limit          int               `json:"limit,omitempty"`
    Offset         int               `json:"offset,omitempty"`
}
```

**Usage**:
```go
instances, meta, err := client.Alerts.ListInstances(ctx, &nexmonyx.ListInstancesOptions{
    OrganizationID: 123,
    State:          ptr("firing"),
    Limit:          100,
})
```

#### 2. GetInstance - Get Single Alert Instance

```go
// GetInstance retrieves a specific alert instance by ID
func (s *AlertsService) GetInstance(ctx context.Context, instanceID uint) (*AlertInstance, error)
```

**Usage**:
```go
instance, err := client.Alerts.GetInstance(ctx, 456)
```

#### 3. CreateInstance - Create Alert Instance

```go
// CreateInstance creates a new alert instance
func (s *AlertsService) CreateInstance(ctx context.Context, req *CreateAlertInstanceRequest) (*AlertInstance, error)

type CreateAlertInstanceRequest struct {
    OrganizationID uint      `json:"organization_id"`
    RuleID         uint      `json:"rule_id"`
    ServerID       uint      `json:"server_id"`
    State          string    `json:"state"`     // firing, acknowledged, resolved
    Severity       string    `json:"severity"`  // info, warning, critical
    Value          float64   `json:"value"`
    Message        string    `json:"message"`
    Metadata       map[string]interface{} `json:"metadata,omitempty"`
}
```

**Usage**:
```go
instance, err := client.Alerts.CreateInstance(ctx, &nexmonyx.CreateAlertInstanceRequest{
    OrganizationID: 123,
    RuleID:         1,
    ServerID:       5,
    State:          "firing",
    Severity:       "critical",
    Value:          98.5,
    Message:        "CPU usage exceeded 95%",
})
```

#### 4. UpdateInstance - Update Alert Instance

```go
// UpdateInstance updates an existing alert instance
func (s *AlertsService) UpdateInstance(ctx context.Context, instanceID uint, req *UpdateAlertInstanceRequest) (*AlertInstance, error)

type UpdateAlertInstanceRequest struct {
    State            *string    `json:"state,omitempty"`
    AcknowledgedBy   *uint      `json:"acknowledged_by,omitempty"`
    ResolvedBy       *uint      `json:"resolved_by,omitempty"`
    SilencedUntil    *time.Time `json:"silenced_until,omitempty"`
    Metadata         map[string]interface{} `json:"metadata,omitempty"`
}
```

**Usage**:
```go
instance, err := client.Alerts.UpdateInstance(ctx, 456, &nexmonyx.UpdateAlertInstanceRequest{
    State:          ptr("acknowledged"),
    AcknowledgedBy: ptr(uint(789)),
})
```

#### 5. ResolveInstance - Auto-Resolve Alert Instance

```go
// ResolveInstance marks an alert instance as resolved (auto or manual)
func (s *AlertsService) ResolveInstance(ctx context.Context, instanceID uint, resolvedBy *uint) (*AlertInstance, error)
```

**Usage**:
```go
// Auto-resolve (no user attribution)
instance, err := client.Alerts.ResolveInstance(ctx, 456, nil)

// Manual resolve (with user ID)
instance, err := client.Alerts.ResolveInstance(ctx, 456, ptr(uint(789)))
```

#### 6. ListInScope - Get Servers Matching Alert Rule Scope

```go
// ListInScope retrieves servers that match an alert rule's filters
func (s *ServersService) ListInScope(ctx context.Context, opts *ScopeOptions) ([]*Server, *PaginationMeta, error)

type ScopeOptions struct {
    OrganizationID uint                   `json:"organization_id"`
    RuleID         uint                   `json:"rule_id"`
    Filters        map[string]interface{} `json:"filters,omitempty"`
    Limit          int                    `json:"limit,omitempty"`
    Offset         int                    `json:"offset,omitempty"`
}
```

**Usage**:
```go
servers, meta, err := client.Servers.ListInScope(ctx, &nexmonyx.ScopeOptions{
    OrganizationID: 123,
    RuleID:         1,
    Filters: map[string]interface{}{
        "tags":        []string{"production", "web"},
        "environment": "production",
    },
    Limit: 100,
})
```

#### 7. Query - Query Metrics

```go
// Query retrieves metric values for a server
func (s *MetricsService) Query(ctx context.Context, req *MetricsQueryRequest) (*MetricsQueryResponse, error)

type MetricsQueryRequest struct {
    ServerID   uint       `json:"server_id"`
    MetricName string     `json:"metric_name"`
    TimeRange  *TimeRange `json:"time_range,omitempty"`
    Limit      int        `json:"limit,omitempty"`
    OrderBy    string     `json:"order_by,omitempty"`
}

type MetricsQueryResponse struct {
    ServerID   uint      `json:"server_id"`
    MetricName string    `json:"metric_name"`
    Values     []float64 `json:"values"`
    Timestamps []time.Time `json:"timestamps"`
    Count      int       `json:"count"`
}
```

**Usage**:
```go
result, err := client.Metrics.Query(ctx, &nexmonyx.MetricsQueryRequest{
    ServerID:   5,
    MetricName: "cpu_usage",
    Limit:      1,
    OrderBy:    "created_at DESC",
})

latestValue := result.Values[0]
```

#### 8. Aggregate - Aggregate Metrics Over Time

```go
// Aggregate computes aggregate metrics over a time range
func (s *MetricsService) Aggregate(ctx context.Context, req *MetricsAggregateRequest) (*MetricsAggregateResponse, error)

type MetricsAggregateRequest struct {
    ServerID     uint       `json:"server_id"`
    MetricName   string     `json:"metric_name"`
    Aggregation  string     `json:"aggregation"`  // AVG, MAX, MIN, SUM, P50, P95, P99
    TimeRange    *TimeRange `json:"time_range"`
}

type TimeRange struct {
    Start time.Time `json:"start"`
    End   time.Time `json:"end"`
}

type MetricsAggregateResponse struct {
    ServerID    uint      `json:"server_id"`
    MetricName  string    `json:"metric_name"`
    Aggregation string    `json:"aggregation"`
    Value       float64   `json:"value"`
    TimeRange   TimeRange `json:"time_range"`
}
```

**Usage**:
```go
result, err := client.Metrics.Aggregate(ctx, &nexmonyx.MetricsAggregateRequest{
    ServerID:    5,
    MetricName:  "cpu_usage",
    Aggregation: "AVG",
    TimeRange: &nexmonyx.TimeRange{
        Start: time.Now().Add(-5 * time.Minute),
        End:   time.Now(),
    },
})

avgCPU := result.Value
```

### SDK Implementation Requirements

**Authentication**:
- All methods must include API Key + Secret in headers
- Support for JWT tokens (user-initiated requests)
- Automatic token refresh for expired JWTs

**Retry Logic**:
- Exponential backoff: 1s, 2s, 4s, 8s, max 30s
- Retry on network errors and 5xx responses
- Max 3 retry attempts
- Configurable retry policy

**Error Handling**:
- Parse API error responses into typed errors
- Distinguish between client errors (4xx) and server errors (5xx)
- Include request ID in error messages for debugging

**Testing**:
- Unit tests with mock HTTP client (>80% coverage)
- Integration tests against mock API server
- Contract tests validating request/response schemas

---

## API Server Enhancements

### Required API Endpoints

The API server must implement **8 new endpoints** corresponding to the SDK methods:

#### 1. List Alert Instances

```
GET /api/v1/alerts/instances
Query Parameters:
  - rule_id (optional): Filter by rule ID
  - server_id (optional): Filter by server ID
  - state (optional): Filter by state (firing, acknowledged, resolved, silenced)
  - severity (optional): Filter by severity (info, warning, critical)
  - fired_after (optional): Filter by fired timestamp (RFC3339)
  - fired_before (optional): Filter by fired timestamp (RFC3339)
  - limit (optional): Pagination limit (default: 50, max: 100)
  - offset (optional): Pagination offset (default: 0)

Response:
{
  "status": "success",
  "data": [AlertInstance, ...],
  "meta": {
    "total": 150,
    "limit": 50,
    "offset": 0,
    "has_more": true
  }
}
```

**Handler**: `pkg/api/handlers/alerts/instances/list_instances.go`

**Authorization**: User must belong to organization, or use organization-scoped API key

#### 2. Get Alert Instance

```
GET /api/v1/alerts/instances/:id

Response:
{
  "status": "success",
  "data": AlertInstance
}
```

**Handler**: `pkg/api/handlers/alerts/instances/get_instance.go`

#### 3. Create Alert Instance

```
POST /api/v1/alerts/instances
Content-Type: application/json

Body:
{
  "organization_id": 123,
  "rule_id": 1,
  "server_id": 5,
  "state": "firing",
  "severity": "critical",
  "value": 98.5,
  "message": "CPU usage exceeded 95%",
  "metadata": {...}
}

Response:
{
  "status": "success",
  "data": AlertInstance
}
```

**Handler**: `pkg/api/handlers/alerts/instances/create_instance.go`

**Validations**:
- Verify rule exists and is enabled
- Verify server exists and belongs to organization
- Verify severity is valid (info, warning, critical)
- Check for duplicate firing instance (same rule + server + state=firing)
- Rate limit: Max 100 instances/minute per organization

#### 4. Update Alert Instance

```
PUT /api/v1/alerts/instances/:id
Content-Type: application/json

Body:
{
  "state": "acknowledged",
  "acknowledged_by": 789,
  "metadata": {...}
}

Response:
{
  "status": "success",
  "data": AlertInstance
}
```

**Handler**: `pkg/api/handlers/alerts/instances/update_instance.go`

**Validations**:
- Verify instance exists and belongs to organization
- Verify state transition is valid (firing → acknowledged → resolved)
- Set timestamp fields based on state (acknowledged_at, resolved_at)

#### 5. Resolve Alert Instance

```
PUT /api/v1/alerts/instances/:id/resolve
Content-Type: application/json

Body:
{
  "resolved_by": 789  // Optional: null for auto-resolve
}

Response:
{
  "status": "success",
  "data": AlertInstance
}
```

**Handler**: `pkg/api/handlers/alerts/instances/resolve_instance.go`

#### 6. List Servers In Scope

```
POST /api/v1/servers/in-scope
Content-Type: application/json

Body:
{
  "organization_id": 123,
  "rule_id": 1,
  "filters": {
    "tags": ["production", "web"],
    "environment": "production"
  },
  "limit": 100,
  "offset": 0
}

Response:
{
  "status": "success",
  "data": [Server, ...],
  "meta": {
    "total": 75,
    "limit": 100,
    "offset": 0
  }
}
```

**Handler**: `pkg/api/handlers/servers/management/list_in_scope.go`

**Logic**:
- Load alert rule to get server_filters
- Merge rule filters with request filters
- Apply tag matching (PostgreSQL `@>` operator or JSON filtering)
- Apply environment, location, classification filters
- Return paginated results

#### 7. Query Metrics

```
POST /api/v1/metrics/query
Content-Type: application/json

Body:
{
  "server_id": 5,
  "metric_name": "cpu_usage",
  "time_range": {
    "start": "2025-10-23T10:00:00Z",
    "end": "2025-10-23T11:00:00Z"
  },
  "limit": 100,
  "order_by": "created_at DESC"
}

Response:
{
  "status": "success",
  "data": {
    "server_id": 5,
    "metric_name": "cpu_usage",
    "values": [98.5, 97.2, 96.8, ...],
    "timestamps": ["2025-10-23T10:59:00Z", "2025-10-23T10:58:00Z", ...],
    "count": 60
  }
}
```

**Handler**: `pkg/api/handlers/metrics/query_metrics.go`

**Optimization**:
- Use TimescaleDB `time_bucket()` for efficient time-based queries
- Cache results for 1 minute (reduces load for repeated queries)
- Index on (server_id, created_at DESC)

#### 8. Aggregate Metrics

```
POST /api/v1/metrics/aggregate
Content-Type: application/json

Body:
{
  "server_id": 5,
  "metric_name": "cpu_usage",
  "aggregation": "AVG",
  "time_range": {
    "start": "2025-10-23T10:00:00Z",
    "end": "2025-10-23T11:00:00Z"
  }
}

Response:
{
  "status": "success",
  "data": {
    "server_id": 5,
    "metric_name": "cpu_usage",
    "aggregation": "AVG",
    "value": 87.3,
    "time_range": {
      "start": "2025-10-23T10:00:00Z",
      "end": "2025-10-23T11:00:00Z"
    }
  }
}
```

**Handler**: `pkg/api/handlers/metrics/aggregate_metrics.go`

**Optimizations**:
- Use TimescaleDB continuous aggregates when available
- Cache aggregation results for 5 minutes
- Support percentiles (P50, P95, P99) using PostgreSQL percentile functions

### Handler Architecture Standards

All handlers MUST follow the single-function-per-file pattern:

**Directory Structure**:
```
pkg/api/handlers/alerts/instances/
├── helpers.go                    # Logger, shared types
├── list_instances.go             # Single function: ListInstances
├── get_instance.go               # Single function: GetInstance
├── create_instance.go            # Single function: CreateInstance
├── update_instance.go            # Single function: UpdateInstance
└── resolve_instance.go           # Single function: ResolveInstance
```

**Logging Pattern** (MANDATORY):
```go
var instanceLogger = logging.NewLogger("handlers", "alerts.instances")

func CreateInstance(c *fiber.Ctx) error {
    if instanceLogger.ShouldTrace("CreateInstance") {
        instanceLogger.Trace(c, "CreateInstance", "Processing request")
    }

    // Business logic with Debug logging

    if instanceLogger.ShouldTrace("CreateInstance") {
        instanceLogger.Trace(c, "CreateInstance", "Operation completed successfully")
    }

    return utils.SendSuccessResponse(c, fiber.StatusCreated, instance, "Alert instance created")
}
```

**Error Handling** (MANDATORY):
```go
import nexerrors "github.com/nexmonyx/nexmonyx/pkg/errors"

// Bad Request
if validation fails {
    return nexerrors.BadRequest(c, "Invalid input", "Severity must be: info, warning, or critical")
}

// Not Found
if instance not found {
    return nexerrors.NotFound(c, "Alert instance not found", "No instance exists with the provided ID")
}

// Internal Error
if database error {
    return nexerrors.InternalServerError(c, "Failed to create alert instance", err.Error())
}
```

**Swagger Documentation** (MANDATORY):
```go
// CreateInstance creates a new alert instance
// @Summary Create alert instance
// @Description Creates a new alert instance when a threshold is violated
// @Tags Alerts
// @Accept json
// @Produce json
// @Param request body CreateAlertInstanceRequest true "Alert instance details"
// @Success 201 {object} utils.Response{data=models.AlertInstance}
// @Failure 400 {object} errors.ErrorResponse "Invalid request"
// @Failure 404 {object} errors.ErrorResponse "Rule or server not found"
// @Failure 500 {object} errors.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /v1/alerts/instances [post]
func CreateInstance(c *fiber.Ctx) error {
    // Implementation
}
```

---

## Controller Refactoring Strategy

### Refactoring Approach: Incremental Migration

**Strategy**: Refactor violations one at a time, maintaining backward compatibility throughout.

**Phases**:
1. **Phase 1**: SDK enhancement (Week 1)
2. **Phase 2**: API server enhancements (Week 2)
3. **Phase 3**: Controller refactoring (Week 3)
4. **Phase 4**: Validation and deployment (Week 4)

### Phase 1: SDK Enhancement (Week 1)

**Goal**: Implement 8 new SDK methods with >80% test coverage

**Tasks**:
1. Create SDK method signatures and types
2. Implement HTTP request/response logic
3. Add retry and error handling
4. Write unit tests with mock HTTP client
5. Write integration tests
6. Update SDK documentation
7. Tag release: v2.4.0

**Deliverables**:
- go-sdk v2.4.0 released
- 8 new methods fully tested
- Documentation updated

**Estimated Effort**: 2-3 days senior engineer

### Phase 2: API Server Enhancements (Week 2)

**Goal**: Implement 8 API endpoints with proper validation and error handling

**Tasks**:
1. Create handler files (single-function-per-file)
2. Implement business logic with proper logging
3. Add input validation
4. Implement error responses using nexerrors
5. Write unit tests for handlers
6. Add integration tests
7. Generate Swagger documentation
8. Update route definitions

**Deliverables**:
- 8 API endpoints deployed to dev environment
- >80% test coverage
- Swagger docs updated

**Estimated Effort**: 2-3 days senior engineer

### Phase 3: Controller Refactoring (Week 3)

**Goal**: Remove all database access, integrate SDK client

**Tasks**:
1. Remove database connection initialization
2. Add SDK client initialization
3. Refactor loadAlertRules() to use SDK
4. Refactor getServersInScope() to use SDK
5. Refactor createAlertInstance() to use SDK
6. Refactor metrics queries to use SDK
7. Update all tests to use mock API client
8. Remove duplicate models (use SDK types)
9. Update error handling

**Deliverables**:
- Alert-controller fully API-first compliant
- All tests passing with mock API
- Zero direct database access

**Estimated Effort**: 3-4 days senior engineer

### Phase 4: Validation & Deployment (Week 4)

**Goal**: Validate performance, deploy to production

**Tasks**:
1. Performance testing (baseline vs API-first)
2. Load testing (5,000 rule evaluations)
3. Integration testing (end-to-end alert flow)
4. Deploy to dev environment
5. Deploy to staging with canary rollout
6. Monitor performance and error rates
7. Deploy to production (phased rollout)
8. Final validation

**Deliverables**:
- Performance metrics documented
- Alert-controller deployed to production
- Zero regressions

**Estimated Effort**: 2-3 days senior engineer + 4 days DevOps + 3 days QA

---

## Migration Phases

### Week 1: SDK Enhancement

**Day 1-2: Method Implementation**
- Create SDK method signatures
- Implement HTTP request logic
- Add authentication headers
- Parse responses

**Day 3: Retry & Error Handling**
- Implement exponential backoff
- Add circuit breaker pattern
- Handle 4xx vs 5xx errors

**Day 4: Testing**
- Unit tests with mock HTTP client
- Integration tests
- Contract tests

**Day 5: Documentation & Release**
- Update README.md
- Add usage examples
- Tag v2.4.0
- Publish to GitHub

**Exit Criteria**:
- ✅ All 8 methods implemented
- ✅ >80% test coverage
- ✅ Integration tests passing
- ✅ v2.4.0 tagged and released

### Week 2: API Server Enhancements

**Day 1-2: Handler Implementation**
- Create handler directory structure
- Implement alert instances handlers (5 handlers)
- Implement servers in-scope handler
- Implement metrics handlers (2 handlers)

**Day 3: Validation & Business Logic**
- Add input validation
- Implement duplicate detection
- Add rate limiting
- Implement caching

**Day 4: Testing**
- Unit tests for handlers
- Integration tests
- Swagger documentation

**Day 5: Deployment to Dev**
- Deploy to dev environment
- Run smoke tests
- Validate API responses

**Exit Criteria**:
- ✅ All 8 endpoints implemented
- ✅ >80% test coverage
- ✅ Deployed to dev environment
- ✅ Swagger docs generated

### Week 3: Controller Refactoring

**Day 1: Foundation**
- Remove database connection code
- Add SDK client initialization
- Update configuration

**Day 2-3: Refactor Data Access**
- Refactor alert rules loading
- Refactor server scope queries
- Refactor alert instance creation
- Refactor metrics queries

**Day 4: Testing**
- Update all unit tests
- Add integration tests with mock API
- Fix any failing tests

**Day 5: Cleanup**
- Remove duplicate models
- Remove unused database code
- Update documentation

**Exit Criteria**:
- ✅ Zero direct database access
- ✅ All tests passing
- ✅ Code review approved
- ✅ Ready for deployment

### Week 4: Validation & Deployment

**Day 1: Performance Testing**
- Baseline performance measurement
- Load testing (5,000 evaluations)
- Latency analysis
- Memory profiling

**Day 2: Deploy to Dev**
- Deploy refactored controller to dev
- Run integration tests
- Monitor for errors

**Day 3: Deploy to Staging**
- Canary deployment (10% traffic)
- Monitor performance metrics
- Gradually increase traffic to 100%

**Day 4-5: Production Deployment**
- Deploy to production (phased rollout)
- Monitor closely for 24 hours
- Document any issues
- Final validation

**Exit Criteria**:
- ✅ Performance within acceptable range
- ✅ Deployed to production
- ✅ Zero critical issues
- ✅ Task #296 marked complete

---

## Performance Analysis

### Current Baseline (Direct Database Access)

**Evaluation Cycle Performance**:
- Cycle duration: 2-8 seconds (100 rules × 50 servers)
- Database queries: ~5,000 SELECT queries
- Peak memory: ~150MB
- CPU usage: 15-25%

**Per-Query Performance**:
- Alert rules query: 50-200ms
- Server scope query: 30-150ms
- Alert instance query: 20-100ms
- Metrics query: 5-15ms
- Metrics aggregate: 100-500ms

### Projected Performance (API-First)

**Without Optimizations**:
- Cycle duration: 8-20 seconds (+6-12 seconds from network latency)
- API calls: ~5,000 HTTP requests
- Peak memory: ~200MB
- CPU usage: 25-40%

**With Optimizations** (caching, batching, parallel requests):
- Cycle duration: 3-10 seconds (comparable to baseline)
- API calls: ~1,000 (80% reduction from batching)
- Peak memory: ~180MB
- CPU usage: 20-35%

### Optimization Strategies

**1. SDK Caching** (reduces redundant API calls):
```go
// Cache alert rules for 5 minutes
rules, _, err := client.Alerts.ListRules(ctx, &nexmonyx.ListOptions{
    Cache: &nexmonyx.CacheOptions{
        TTL:     5 * time.Minute,
        Enabled: true,
    },
})
```

**Impact**: 90% reduction in alert rules API calls

**2. Batch API Calls** (reduces HTTP overhead):
```go
// Batch server queries
servers, _, err := client.Servers.ListInScope(ctx, &nexmonyx.ScopeOptions{
    OrganizationID: orgID,
    RuleID:         ruleID,
    Limit:          500,  // Fetch all servers in one call
})
```

**Impact**: 95% reduction in server API calls

**3. Parallel Evaluation** (reduces total time):
```go
var wg sync.WaitGroup
semaphore := make(chan struct{}, 10)

for _, rule := range rules {
    wg.Add(1)
    semaphore <- struct{}{}
    go func(r *nexmonyx.AlertRule) {
        defer wg.Done()
        defer func() { <-semaphore }()
        evaluateRule(r)
    }(rule)
}
wg.Wait()
```

**Impact**: 70% reduction in cycle time (2-3x speedup)

**4. Connection Pooling** (reduces connection overhead):
```go
client := nexmonyx.NewClient(&nexmonyx.Config{
    HTTPClient: &http.Client{
        Transport: &http.Transport{
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 10,
            IdleConnTimeout:     90 * time.Second,
        },
    },
})
```

**Impact**: 40% reduction in network latency

### Performance Comparison

| Metric | Current | API-First (No Opt) | API-First (Optimized) | Change |
|--------|---------|-------------------|----------------------|--------|
| Cycle Time | 2-8s | 8-20s | 3-10s | +25% to -20% |
| Database Load | High | Low | Very Low | -60% |
| API Calls | 0 | ~5,000 | ~1,000 | N/A |
| Memory | 150MB | 200MB | 180MB | +20% |
| CPU | 15-25% | 25-40% | 20-35% | +10% |

**Conclusion**: With proper optimizations, API-first architecture achieves comparable or better performance while providing architectural benefits.

---

## Risk Assessment

### High Risks

**Risk #1: Performance Degradation**
- **Probability**: Medium
- **Impact**: High
- **Mitigation**:
  - Implement aggressive caching in SDK
  - Use batch API endpoints
  - Parallel evaluation of rules
  - Performance testing before production

**Risk #2: API Server Dependency**
- **Probability**: Low
- **Impact**: Critical
- **Mitigation**:
  - Implement circuit breaker pattern
  - Graceful degradation (skip evaluation cycles if API unavailable)
  - Health check monitoring
  - API server HA deployment (multiple replicas)

**Risk #3: Breaking Changes in Production**
- **Probability**: Low
- **Impact**: Critical
- **Mitigation**:
  - Comprehensive integration testing
  - Canary deployment to staging
  - Phased production rollout (10% → 50% → 100%)
  - Rollback procedure documented and tested

### Medium Risks

**Risk #4: SDK Bugs**
- **Probability**: Medium
- **Impact**: Medium
- **Mitigation**:
  - >80% test coverage requirement
  - Integration tests with real API
  - Code review by 2+ engineers

**Risk #5: Timeline Overrun**
- **Probability**: Medium
- **Impact**: Low
- **Mitigation**:
  - Conservative 4-week timeline estimate
  - Daily standup with progress tracking
  - Escalation path for blockers

### Low Risks

**Risk #6: Incomplete Documentation**
- **Probability**: Low
- **Impact**: Low
- **Mitigation**:
  - Documentation updates in each phase
  - Swagger auto-generation
  - Code review includes doc review

---

## Success Criteria

### Technical Success Criteria

1. ✅ **Zero Direct Database Access**
   - No GORM imports in alert-controller
   - All data access via go-sdk

2. ✅ **Test Coverage >80%**
   - SDK methods: >80%
   - API handlers: >80%
   - Controller refactored code: >80%

3. ✅ **Performance Parity**
   - Evaluation cycle time: ≤ 10 seconds (current max: 8s)
   - Memory usage: ≤ 200MB (current: 150MB)
   - CPU usage: ≤ 40% (current: 25%)

4. ✅ **Zero Breaking Changes**
   - Alert evaluation continues without interruption
   - All alert instances created successfully
   - Notifications delivered as before

5. ✅ **API-First Compliance**
   - Architecture review approval
   - Passes API-first validation checklist

### Business Success Criteria

1. ✅ **Zero Production Incidents**
   - No alert evaluation failures
   - No notification delivery failures
   - No data loss or corruption

2. ✅ **Deployment Success**
   - Deployed to dev, staging, production
   - Phased rollout completed
   - No rollbacks required

3. ✅ **Documentation Complete**
   - All 4 design documents completed
   - API documentation updated
   - Runbooks updated

### Validation Checklist

**Pre-Deployment**:
- [ ] All 8 SDK methods implemented and tested
- [ ] All 8 API endpoints deployed to dev
- [ ] Controller refactored with zero direct DB access
- [ ] Integration tests passing (>80% coverage)
- [ ] Performance tests show acceptable results
- [ ] Code review approved by 2+ engineers

**Post-Deployment (Dev)**:
- [ ] Alert evaluation cycles completing successfully
- [ ] Alert instances being created
- [ ] Notifications being delivered
- [ ] No error spikes in logs
- [ ] API response times acceptable (<500ms p95)

**Post-Deployment (Staging)**:
- [ ] Canary deployment successful (10% → 100%)
- [ ] 24-hour monitoring shows stable performance
- [ ] Integration tests passing in staging
- [ ] Load testing completed successfully

**Post-Deployment (Production)**:
- [ ] Phased rollout completed (10% → 50% → 100%)
- [ ] 48-hour monitoring shows stable performance
- [ ] Zero critical incidents
- [ ] Performance metrics within acceptable range
- [ ] Task #296 marked complete in TaskForge

---

**Document Status**: Draft - API-First Refactoring Strategy Complete

**Next Steps**:
1. Review and approve refactoring strategy
2. Begin SDK enhancement (Phase 1, Week 1)
3. Implement API endpoints (Phase 2, Week 2)
4. Refactor controller (Phase 3, Week 3)
5. Deploy and validate (Phase 4, Week 4)
