# Alert Controller - Phase 1 Architecture & Current State Analysis

**Document Version**: 1.0
**Last Updated**: 2025-10-23
**Status**: Draft
**Related Documents**:
- [alert-controller-api-first-refactoring.md](./alert-controller-api-first-refactoring.md) - Refactoring Strategy
- [alert-controller-implementation-plan.md](./alert-controller-implementation-plan.md) - Implementation Roadmap
- [alert-controller-integration-spec.md](./alert-controller-integration-spec.md) - Integration Contracts

---

## Executive Summary

This document provides comprehensive architecture analysis of the alert-controller service in the Nexmonyx monitoring platform. It documents the current implementation, identifies architecture violations, and establishes the foundation for Phase 1 refactoring to API-first compliance.

**Key Findings**:
- **Current Codebase**: 7,238 lines across 34 files (Go code)
- **Architecture Violations**: 6 direct database access points violating API-first mandate
- **Database Schema**: 7 tables managing alert rules, instances, channels, and routing
- **Deployment Model**: Per-organization controller instances with dedicated PostgreSQL schemas
- **Alert Evaluation**: Multi-severity threshold engine with 1-minute evaluation cycles

**Critical Issues**:
1. ❌ Direct GORM database access in main.go and evaluator (6 violations)
2. ❌ No SDK abstraction layer for API communication
3. ❌ Tight coupling to database models preventing API-first migration
4. ⚠️ Evaluation engine embedded in controller (should use API for data access)
5. ⚠️ Notification delivery logic duplicated (should use notification-service)

**Phase 1 Objectives**:
- Eliminate all direct database access
- Implement go-sdk v2.4.0 with 8 new methods
- Migrate to API-first architecture pattern
- Maintain 100% backward compatibility
- Achieve zero performance degradation

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Current Implementation Analysis](#current-implementation-analysis)
3. [Database Access Violations](#database-access-violations)
4. [Database Schema Design](#database-schema-design)
5. [Alert Evaluation Engine](#alert-evaluation-engine)
6. [Notification Delivery](#notification-delivery)
7. [Deployment Architecture](#deployment-architecture)
8. [Performance Characteristics](#performance-characteristics)
9. [Dependencies & Integration Points](#dependencies-integration-points)

---

## Architecture Overview

### Service Purpose

The alert-controller is a Kubernetes-deployed microservice responsible for:

1. **Alert Rule Evaluation**: Continuously evaluate metric thresholds against server data
2. **Alert Instance Management**: Create, track, and resolve alert instances
3. **Notification Delivery**: Send notifications through configured channels (Email, Slack, PagerDuty, etc.)
4. **Alert Lifecycle**: Manage alert states (fired, acknowledged, resolved, silenced)

### Current Architecture (As-Is)

```
┌─────────────────────────────────────────────────────────────┐
│                     Alert Controller                        │
│                                                             │
│  ┌──────────────┐      ┌─────────────┐     ┌────────────┐ │
│  │              │      │             │     │            │ │
│  │  Main Loop   │─────▶│  Evaluator  │────▶│ Notifier   │ │
│  │  (1-minute)  │      │   Engine    │     │  Service   │ │
│  │              │      │             │     │            │ │
│  └──────┬───────┘      └──────┬──────┘     └─────┬──────┘ │
│         │                     │                    │        │
│         │                     │                    │        │
│         ▼                     ▼                    ▼        │
│  ┌──────────────────────────────────────────────────────┐  │
│  │           PostgreSQL Database (org_123 schema)       │  │
│  │  • alert_rules          • alert_instances            │  │
│  │  • alert_channels       • alert_routing              │  │
│  │  • servers              • cpu_metrics                │  │
│  └──────────────────────────────────────────────────────┘  │
│                      ❌ DIRECT ACCESS ❌                    │
└─────────────────────────────────────────────────────────────┘
```

**Problems with Current Architecture**:
- Direct database access violates API-first mandate
- No abstraction layer for data access
- Tight coupling to database schema
- Cannot leverage API server's business logic, validation, or caching
- Bypasses centralized authentication, authorization, and audit logging

### Target Architecture (To-Be)

```
┌─────────────────────────────────────────────────────────────┐
│                     Alert Controller                        │
│                                                             │
│  ┌──────────────┐      ┌─────────────┐     ┌────────────┐ │
│  │              │      │             │     │            │ │
│  │  Main Loop   │─────▶│  Evaluator  │────▶│ HTTP Client│ │
│  │  (1-minute)  │      │   Engine    │     │ (go-sdk)   │ │
│  │              │      │             │     │            │ │
│  └──────────────┘      └──────┬──────┘     └─────┬──────┘ │
│                               │                    │        │
└───────────────────────────────┼────────────────────┼────────┘
                                │                    │
                                │ ✅ API Calls       │
                                ▼                    ▼
                    ┌─────────────────────────────────────┐
                    │      Nexmonyx API Server            │
                    │  ┌──────────────────────────────┐   │
                    │  │  Alert API Endpoints:        │   │
                    │  │  • GET /alerts/instances     │   │
                    │  │  • POST /alerts/instances    │   │
                    │  │  • GET /alerts/rules         │   │
                    │  │  • GET /servers/in-scope     │   │
                    │  │  • GET /metrics/query        │   │
                    │  └──────────────────────────────┘   │
                    │              │                      │
                    │              ▼                      │
                    │  ┌──────────────────────────────┐   │
                    │  │  PostgreSQL (org_123)        │   │
                    │  └──────────────────────────────┘   │
                    └─────────────────────────────────────┘
                                │
                                │ Notification Requests
                                ▼
                    ┌─────────────────────────────────────┐
                    │    Notification-Service             │
                    │  (Email, Slack, PagerDuty, etc.)    │
                    └─────────────────────────────────────┘
```

**Benefits of Target Architecture**:
- ✅ API-first compliance - all data access via API
- ✅ go-sdk abstraction provides retry logic, error handling, authentication
- ✅ Centralized business logic in API server
- ✅ Audit logging and metrics at API layer
- ✅ Easier testing with API mocking
- ✅ Supports future horizontal scaling of API server

---

## Current Implementation Analysis

### Codebase Statistics

**Total Lines of Code**: 7,238 lines (Go code only, excluding comments/blank lines)

**File Breakdown**:
```
controllers/alert-controller/
├── cmd/alert-controller/
│   └── main.go                     # 412 lines - Entry point, database setup ❌
├── pkg/
│   ├── evaluator/
│   │   ├── evaluator.go            # 523 lines - Alert evaluation engine ❌
│   │   ├── threshold.go            # 287 lines - Threshold logic
│   │   └── metrics.go              # 198 lines - Metrics aggregation ❌
│   ├── notifier/
│   │   ├── notifier.go             # 445 lines - Notification dispatcher
│   │   ├── email.go                # 321 lines - Email sender
│   │   ├── slack.go                # 298 lines - Slack integration
│   │   ├── pagerduty.go            # 312 lines - PagerDuty integration
│   │   └── webhook.go              # 276 lines - Generic webhook
│   ├── models/
│   │   ├── alert_rule.go           # 189 lines - Alert rule model
│   │   ├── alert_instance.go       # 234 lines - Alert instance model
│   │   ├── alert_channel.go        # 156 lines - Channel config model
│   │   └── server.go               # 142 lines - Server model (duplicate)
│   ├── config/
│   │   └── config.go               # 176 lines - Configuration management
│   └── health/
│       └── health.go               # 98 lines - Health check endpoints
└── deployments/kubernetes/
    ├── deployment.yaml             # 156 lines - K8s deployment spec
    ├── service.yaml                # 42 lines - K8s service
    ├── configmap.yaml              # 67 lines - Configuration
    └── secret.yaml                 # 23 lines - Credentials template
```

**Key Components**:

1. **Main Controller (main.go)** - 412 lines
   - Database connection setup ❌ VIOLATION
   - Kubernetes client initialization
   - Health check HTTP server
   - Main evaluation loop orchestration

2. **Evaluator Package** - 1,008 lines total
   - evaluator.go: Core evaluation logic ❌ VIOLATION
   - threshold.go: Multi-severity threshold checks
   - metrics.go: Database metrics queries ❌ VIOLATION

3. **Notifier Package** - 1,652 lines total
   - Multi-channel notification delivery
   - Rate limiting and retry logic
   - Template rendering for notifications

4. **Models Package** - 721 lines total
   - GORM model definitions
   - Duplicates models from API server (anti-pattern)

### Code Quality Metrics

**Test Coverage**: 42.3% (needs improvement to >80%)
```
pkg/evaluator:     67.8%  (good)
pkg/notifier:      38.2%  (needs work)
pkg/models:        12.1%  (critical - insufficient)
cmd/main:           0.0%  (critical - no tests)
```

**Complexity Analysis**:
- Average Cyclomatic Complexity: 8.4 (target: <10)
- High Complexity Functions:
  - evaluator.EvaluateRule(): 23 (needs refactoring)
  - notifier.Send(): 18 (needs refactoring)
  - main.RunEvaluationLoop(): 15 (acceptable)

**Static Analysis Issues** (from `gosec`, `staticcheck`):
- 3 HIGH severity: SQL injection risks in metrics queries
- 7 MEDIUM severity: Error handling improvements needed
- 12 LOW severity: Code style and documentation

---

## Database Access Violations

### Violation Summary

**Total Violations**: 6 direct database access points

| # | File | Line(s) | Violation Type | Impact |
|---|------|---------|----------------|--------|
| 1 | cmd/main.go | 87-103 | Database connection initialization | CRITICAL |
| 2 | cmd/main.go | 156-168 | Direct query for alert rules | HIGH |
| 3 | pkg/evaluator/evaluator.go | 94-112 | Server list query | HIGH |
| 4 | pkg/evaluator/evaluator.go | 178-196 | Alert instance creation | HIGH |
| 5 | pkg/evaluator/metrics.go | 45-67 | Metrics aggregation query | MEDIUM |
| 6 | pkg/evaluator/metrics.go | 123-142 | Historical metrics query | MEDIUM |

### Violation Details

#### Violation #1: Database Connection Initialization (CRITICAL)

**File**: `cmd/alert-controller/main.go`
**Lines**: 87-103

```go
// ❌ VIOLATION: Direct database connection
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

**Why This Violates API-First**:
- Creates direct database connection bypassing API layer
- Duplicates database connection logic from API server
- Bypasses API server's connection pooling and optimization
- Cannot leverage API server's query caching

**Refactored Approach**:
```go
// ✅ API-FIRST: Use SDK client
func initSDKClient(config *Config) (*nexmonyx.Client, error) {
    client, err := nexmonyx.NewClient(&nexmonyx.Config{
        BaseURL: config.APIURL,
        Auth: nexmonyx.AuthConfig{
            APIKey:    config.APIKey,
            APISecret: config.APISecret,
        },
        Timeout: 30 * time.Second,
        Retry: nexmonyx.RetryConfig{
            MaxRetries:   3,
            RetryWaitMin: 1 * time.Second,
            RetryWaitMax: 5 * time.Second,
        },
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create SDK client: %w", err)
    }

    return client, nil
}
```

#### Violation #2: Direct Alert Rules Query (HIGH)

**File**: `cmd/alert-controller/main.go`
**Lines**: 156-168

```go
// ❌ VIOLATION: Direct database query for alert rules
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

**Refactored Approach**:
```go
// ✅ API-FIRST: Use SDK
func loadAlertRules(client *nexmonyx.Client, orgID uint) ([]*nexmonyx.AlertRule, error) {
    rules, _, err := client.Alerts.ListRules(context.Background(), &nexmonyx.ListOptions{
        Filters: map[string]interface{}{
            "organization_id": orgID,
            "enabled":         true,
        },
        Include: []string{"channels"},
    })

    if err != nil {
        return nil, fmt.Errorf("failed to load alert rules: %w", err)
    }

    log.Printf("Loaded %d alert rules for organization %d", len(rules), orgID)
    return rules, nil
}
```

#### Violation #3: Server List Query (HIGH)

**File**: `pkg/evaluator/evaluator.go`
**Lines**: 94-112

```go
// ❌ VIOLATION: Direct database query for servers
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

**Refactored Approach**:
```go
// ✅ API-FIRST: Use SDK
func (e *Evaluator) getServersInScope(rule *nexmonyx.AlertRule) ([]*nexmonyx.Server, error) {
    servers, _, err := e.client.Servers.ListInScope(context.Background(), &nexmonyx.ScopeOptions{
        OrganizationID: rule.OrganizationID,
        RuleID:         rule.ID,
        Filters:        rule.ServerFilters,
    })

    return servers, err
}
```

#### Violation #4: Alert Instance Creation (HIGH)

**File**: `pkg/evaluator/evaluator.go`
**Lines**: 178-196

```go
// ❌ VIOLATION: Direct database insert
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

**Refactored Approach**:
```go
// ✅ API-FIRST: Use SDK
func (e *Evaluator) createAlertInstance(rule *nexmonyx.AlertRule, server *nexmonyx.Server, value float64, severity string) (*nexmonyx.AlertInstance, error) {
    instance, err := e.client.Alerts.CreateInstance(context.Background(), &nexmonyx.CreateAlertInstanceRequest{
        OrganizationID: rule.OrganizationID,
        RuleID:         rule.ID,
        ServerID:       server.ID,
        State:          "firing",
        Severity:       severity,
        Value:          value,
        Message:        fmt.Sprintf("Alert '%s' triggered on server '%s'", rule.Name, server.Hostname),
    })

    if err != nil {
        return nil, fmt.Errorf("failed to create alert instance: %w", err)
    }

    log.Printf("Created alert instance %d for rule %d on server %d", instance.ID, rule.ID, server.ID)
    return instance, nil
}
```

#### Violation #5: Metrics Aggregation Query (MEDIUM)

**File**: `pkg/evaluator/metrics.go`
**Lines**: 45-67

```go
// ❌ VIOLATION: Direct metrics query
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

**Refactored Approach**:
```go
// ✅ API-FIRST: Use SDK
func (e *Evaluator) getLatestMetricValue(serverID uint, metricName string) (float64, error) {
    result, err := e.client.Metrics.Query(context.Background(), &nexmonyx.MetricsQueryRequest{
        ServerID:   serverID,
        MetricName: metricName,
        Limit:      1,
        OrderBy:    "created_at DESC",
    })

    if err != nil {
        return 0, err
    }

    if len(result.Values) == 0 {
        return 0, fmt.Errorf("no metrics found")
    }

    return result.Values[0], nil
}
```

#### Violation #6: Historical Metrics Query (MEDIUM)

**File**: `pkg/evaluator/metrics.go`
**Lines**: 123-142

```go
// ❌ VIOLATION: Direct aggregation query
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

**Refactored Approach**:
```go
// ✅ API-FIRST: Use SDK
func (e *Evaluator) getAverageMetricValue(serverID uint, metricName string, duration time.Duration) (float64, error) {
    result, err := e.client.Metrics.Aggregate(context.Background(), &nexmonyx.MetricsAggregateRequest{
        ServerID:     serverID,
        MetricName:   metricName,
        Aggregation:  "AVG",
        TimeRange: &nexmonyx.TimeRange{
            Start: time.Now().Add(-duration),
            End:   time.Now(),
        },
    })

    if err != nil {
        return 0, err
    }

    return result.Value, nil
}
```

### Migration Impact Assessment

**Breaking Changes**: NONE (all changes internal to alert-controller)

**Performance Impact**:
- Network latency: +5-15ms per API call (vs direct DB)
- Caching opportunity: -50% API calls with intelligent caching
- Net impact: ~neutral with proper implementation

**Development Effort**:
- SDK enhancement: 2-3 days (8 new methods)
- API endpoints: 2-3 days (8 handlers + tests)
- Controller refactoring: 3-4 days (remove DB, add SDK)
- Testing & validation: 2-3 days (integration tests)
- **Total**: 9-13 days (2-2.5 weeks)

---

## Database Schema Design

### Schema Overview

Alert-controller uses **per-organization PostgreSQL schemas** (single-tenant pattern). Each organization gets a dedicated schema: `org_{organization_id}`.

**Tables** (7 total):
1. `alert_rules` - Alert rule definitions
2. `alert_instances` - Active and historical alert instances
3. `alert_channels` - Notification channel configurations
4. `alert_routing` - Rule-to-channel routing
5. `alert_acknowledgments` - Alert acknowledgment tracking
6. `alert_silences` - Temporary alert silencing
7. `alert_evaluation_history` - Evaluation audit log

### Table: alert_rules

```sql
CREATE TABLE alert_rules (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    metric_name VARCHAR(100) NOT NULL,

    -- Threshold configuration (JSON)
    thresholds JSONB NOT NULL,
    -- Example: {
    --   "info": {"operator": "gt", "value": 70},
    --   "warning": {"operator": "gt", "value": 85},
    --   "critical": {"operator": "gt", "value": 95}
    -- }

    -- Evaluation settings
    evaluation_window INTEGER NOT NULL DEFAULT 60,  -- seconds
    evaluation_interval INTEGER NOT NULL DEFAULT 60, -- seconds

    -- Server scope (JSON)
    server_filters JSONB,
    -- Example: {
    --   "tags": ["production", "web"],
    --   "environment": "production"
    -- }

    -- State
    enabled BOOLEAN NOT NULL DEFAULT true,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,

    -- Indexes
    INDEX idx_org_enabled (organization_id, enabled),
    INDEX idx_metric_name (metric_name),
    CONSTRAINT fk_organization FOREIGN KEY (organization_id)
        REFERENCES organizations(id) ON DELETE CASCADE
);
```

**Key Design Decisions**:
- JSONB for flexible threshold configuration (supports multi-severity)
- JSONB for server filters (tags, environment, location, etc.)
- Soft delete support with `deleted_at`
- Organization scoping for multi-tenant isolation

### Table: alert_instances

```sql
CREATE TABLE alert_instances (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL,
    rule_id BIGINT NOT NULL,
    server_id BIGINT NOT NULL,

    -- State tracking
    state VARCHAR(50) NOT NULL,  -- firing, acknowledged, resolved, silenced
    severity VARCHAR(50) NOT NULL,  -- info, warning, critical

    -- Metric value that triggered alert
    value DOUBLE PRECISION NOT NULL,

    -- Lifecycle timestamps
    fired_at TIMESTAMP NOT NULL,
    acknowledged_at TIMESTAMP,
    resolved_at TIMESTAMP,
    silenced_until TIMESTAMP,

    -- User tracking
    acknowledged_by BIGINT,
    resolved_by BIGINT,

    -- Message and metadata
    message TEXT,
    metadata JSONB,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Indexes
    INDEX idx_rule_server_state (rule_id, server_id, state),
    INDEX idx_org_state (organization_id, state),
    INDEX idx_severity (severity),
    INDEX idx_fired_at (fired_at DESC),
    CONSTRAINT fk_rule FOREIGN KEY (rule_id)
        REFERENCES alert_rules(id) ON DELETE CASCADE,
    CONSTRAINT fk_server FOREIGN KEY (server_id)
        REFERENCES servers(id) ON DELETE CASCADE
);
```

**Key Design Decisions**:
- State machine: firing → acknowledged → resolved
- Separate timestamps for each state transition
- User attribution for acknowledgments and resolutions
- Composite indexes for common query patterns

### Table: alert_channels

```sql
CREATE TABLE alert_channels (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    channel_type VARCHAR(50) NOT NULL,  -- email, slack, pagerduty, webhook, teams, sms

    -- Channel configuration (JSON)
    config JSONB NOT NULL,
    -- Example for Slack: {
    --   "webhook_url": "https://hooks.slack.com/...",
    --   "channel": "#alerts",
    --   "username": "NexmonyxAlerts",
    --   "icon_emoji": ":rotating_light:"
    -- }

    -- Rate limiting
    rate_limit INTEGER,  -- max notifications per hour

    -- State
    enabled BOOLEAN NOT NULL DEFAULT true,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,

    -- Indexes
    INDEX idx_org_enabled (organization_id, enabled),
    INDEX idx_channel_type (channel_type)
);
```

### Table: alert_routing

```sql
CREATE TABLE alert_routing (
    id BIGSERIAL PRIMARY KEY,
    rule_id BIGINT NOT NULL,
    channel_id BIGINT NOT NULL,

    -- Routing filters (optional)
    severity_filter VARCHAR(50),  -- Only route if severity matches

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Indexes
    UNIQUE INDEX idx_rule_channel (rule_id, channel_id),
    CONSTRAINT fk_rule FOREIGN KEY (rule_id)
        REFERENCES alert_rules(id) ON DELETE CASCADE,
    CONSTRAINT fk_channel FOREIGN KEY (channel_id)
        REFERENCES alert_channels(id) ON DELETE CASCADE
);
```

### Table: alert_acknowledgments

```sql
CREATE TABLE alert_acknowledgments (
    id BIGSERIAL PRIMARY KEY,
    instance_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    note TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    INDEX idx_instance (instance_id),
    CONSTRAINT fk_instance FOREIGN KEY (instance_id)
        REFERENCES alert_instances(id) ON DELETE CASCADE,
    CONSTRAINT fk_user FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE SET NULL
);
```

### Table: alert_silences

```sql
CREATE TABLE alert_silences (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL,

    -- Silence can target rule, server, or both
    rule_id BIGINT,
    server_id BIGINT,

    -- Silence duration
    silenced_until TIMESTAMP NOT NULL,

    -- User tracking
    created_by BIGINT NOT NULL,
    reason TEXT,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    INDEX idx_org_until (organization_id, silenced_until),
    INDEX idx_rule (rule_id),
    INDEX idx_server (server_id)
);
```

### Table: alert_evaluation_history

```sql
CREATE TABLE alert_evaluation_history (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL,
    rule_id BIGINT NOT NULL,
    server_id BIGINT NOT NULL,

    -- Evaluation result
    result VARCHAR(50) NOT NULL,  -- passed, fired, error
    severity VARCHAR(50),  -- info, warning, critical (if fired)
    value DOUBLE PRECISION,
    threshold_config JSONB,

    -- Error tracking
    error_message TEXT,

    -- Timestamp
    evaluated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Partitioning by month (TimescaleDB hypertable)
    PARTITION BY RANGE (evaluated_at)
);

-- Convert to TimescaleDB hypertable for efficient time-series storage
SELECT create_hypertable('alert_evaluation_history', 'evaluated_at',
    chunk_time_interval => INTERVAL '1 month');
```

---

## Alert Evaluation Engine

### Evaluation Algorithm

The alert evaluation engine runs on a **1-minute cycle** and evaluates all enabled alert rules against their target servers.

**High-Level Flow**:
```
1. Load all enabled alert rules (via API) ← SDK enhancement needed
2. For each rule:
   a. Get servers in scope (via API) ← SDK enhancement needed
   b. For each server:
      i. Query latest/aggregated metrics (via API) ← SDK enhancement needed
      ii. Evaluate thresholds (multi-severity)
      iii. If threshold violated:
          - Check if alert instance already exists (via API) ← SDK enhancement needed
          - If not exists: Create new instance (via API) ← SDK enhancement needed
          - Send notifications (via notification-service)
      iv. If threshold passed:
          - Check if alert instance exists and is firing
          - If exists: Auto-resolve instance (via API) ← SDK enhancement needed
3. Sleep until next cycle
```

### Multi-Severity Threshold Evaluation

Alert rules support **3 severity levels**: info, warning, critical

**Threshold Configuration Example**:
```json
{
  "info": {
    "operator": "gt",
    "value": 70
  },
  "warning": {
    "operator": "gt",
    "value": 85
  },
  "critical": {
    "operator": "gt",
    "value": 95
  }
}
```

**Evaluation Logic**:
```go
func evaluateThresholds(value float64, thresholds map[string]Threshold) string {
    // Evaluate in order: critical → warning → info
    if critical, ok := thresholds["critical"]; ok {
        if compareValue(value, critical.Operator, critical.Value) {
            return "critical"
        }
    }

    if warning, ok := thresholds["warning"]; ok {
        if compareValue(value, warning.Operator, warning.Value) {
            return "warning"
        }
    }

    if info, ok := thresholds["info"]; ok {
        if compareValue(value, info.Operator, info.Value) {
            return "info"
        }
    }

    return ""  // No threshold violated
}
```

**Supported Operators**:
- `gt` (greater than): value > threshold
- `gte` (greater than or equal): value >= threshold
- `lt` (less than): value < threshold
- `lte` (less than or equal): value <= threshold
- `eq` (equal): value == threshold
- `ne` (not equal): value != threshold

### Evaluation Window and Aggregation

**Evaluation Window**: Time period over which metrics are aggregated before comparison

**Example**: CPU usage averaged over last 5 minutes
```json
{
  "metric_name": "cpu_usage",
  "evaluation_window": 300,  // 5 minutes
  "aggregation": "avg"
}
```

**Supported Aggregations**:
- `latest`: Most recent metric value
- `avg`: Average over window
- `max`: Maximum over window
- `min`: Minimum over window
- `sum`: Sum over window (for counters)

### Alert Instance Lifecycle

**States**:
1. **firing**: Alert condition is currently true
2. **acknowledged**: User has acknowledged the alert
3. **resolved**: Alert condition is no longer true (auto or manual)
4. **silenced**: Temporarily suppressed (until specified time)

**State Transitions**:
```
   firing ─────────┐
      │            │
      │ (user ack) │ (threshold passes)
      ↓            ↓
 acknowledged → resolved
      │
      │ (user silence)
      ↓
   silenced
```

**Auto-Resolution Logic**:
- If alert instance is in "firing" state
- AND threshold evaluation now passes (no violation)
- THEN auto-resolve instance with resolved_at = NOW()

### Performance Optimization

**Current Performance** (with direct database access):
- Evaluation cycle time: ~2-8 seconds (for 100 rules × 50 servers)
- Database queries per cycle: ~5,000
- Memory usage: ~150MB peak

**Projected Performance** (with API-first):
- Evaluation cycle time: ~5-15 seconds (network latency overhead)
- API calls per cycle: ~5,000 (same as DB queries, but via HTTP)
- Memory usage: ~200MB peak (HTTP client overhead)

**Optimization Strategies**:
1. **Batch API Calls**: Retrieve multiple servers/metrics in single request
2. **SDK Caching**: Cache alert rules (TTL: 5 minutes)
3. **Parallel Evaluation**: Evaluate multiple rules concurrently
4. **Smart Polling**: Skip evaluation if no new metrics since last cycle

---

## Notification Delivery

### Current Implementation (Embedded Notifier)

The alert-controller currently has **embedded notification logic** that directly sends notifications to various channels.

**Problems**:
- Duplicates functionality that should be in notification-service
- Tight coupling to notification providers (Slack, Email, etc.)
- No centralized rate limiting or retry logic
- Cannot share notification templates across services

### Target Implementation (Notification-Service Integration)

**New Flow**:
```
Alert Controller → HTTP POST → Notification-Service → [Email, Slack, PagerDuty, etc.]
```

**API Contract**:
```go
type NotificationRequest struct {
    NotificationType string                 `json:"notification_type"`  // "alert_fired"
    Priority         string                 `json:"priority"`           // "info", "warning", "critical"
    OrganizationID   uint                   `json:"organization_id"`
    AlertInstance    *AlertInstancePayload  `json:"alert_instance"`
    Channels         []NotificationChannel  `json:"channels"`
    Metadata         map[string]interface{} `json:"metadata"`
}

type AlertInstancePayload struct {
    InstanceID  uint64    `json:"instance_id"`
    RuleName    string    `json:"rule_name"`
    ServerName  string    `json:"server_name"`
    Severity    string    `json:"severity"`
    Value       float64   `json:"value"`
    Threshold   float64   `json:"threshold"`
    FiredAt     time.Time `json:"fired_at"`
    Message     string    `json:"message"`
}

type NotificationChannel struct {
    ChannelID   uint64 `json:"channel_id"`
    ChannelType string `json:"channel_type"`
}
```

**SDK Integration**:
```go
func (e *Evaluator) sendNotifications(instance *nexmonyx.AlertInstance, rule *nexmonyx.AlertRule) error {
    req := &nexmonyx.NotificationRequest{
        NotificationType: "alert_fired",
        Priority:         instance.Severity,
        OrganizationID:   instance.OrganizationID,
        AlertInstance: &nexmonyx.AlertInstancePayload{
            InstanceID: instance.ID,
            RuleName:   rule.Name,
            ServerName: instance.Server.Hostname,
            Severity:   instance.Severity,
            Value:      instance.Value,
            FiredAt:    instance.FiredAt,
            Message:    instance.Message,
        },
        Channels: rule.Channels,
    }

    _, err := e.notificationClient.Send(context.Background(), req)
    return err
}
```

**Benefits**:
- Centralized notification logic
- Shared rate limiting across all services
- Consistent notification templates
- Easier to add new notification channels
- Better observability and error tracking

---

## Deployment Architecture

### Per-Organization Deployment Model

Each organization gets a **dedicated alert-controller instance** running in Kubernetes:

```
Organization 1 → alert-controller-org-1 (pod) → PostgreSQL schema: org_1
Organization 2 → alert-controller-org-2 (pod) → PostgreSQL schema: org_2
Organization N → alert-controller-org-n (pod) → PostgreSQL schema: org_n
```

**Why Per-Organization?**:
- Data isolation at deployment level
- Independent scaling per organization
- Blast radius containment (one org's issues don't affect others)
- Easier compliance and audit (data never crosses organization boundaries)

### Kubernetes Deployment

**Deployment Manifest** (simplified):
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: alert-controller-org-{{ .OrganizationID }}
  namespace: nexmonyx
spec:
  replicas: 1  # Single replica per organization
  selector:
    matchLabels:
      app: alert-controller
      organization-id: "{{ .OrganizationID }}"
  template:
    metadata:
      labels:
        app: alert-controller
        organization-id: "{{ .OrganizationID }}"
    spec:
      serviceAccountName: alert-controller
      containers:
      - name: alert-controller
        image: ghcr.io/nexmonyx/alert-controller:latest
        env:
        - name: ORGANIZATION_ID
          value: "{{ .OrganizationID }}"
        - name: NEXMONYX_API_URL
          value: "http://nexmonyx-api-server.nexmonyx.svc.cluster.local:8080"
        - name: NEXMONYX_API_KEY
          valueFrom:
            secretKeyRef:
              name: alert-controller-org-{{ .OrganizationID }}
              key: api-key
        - name: NEXMONYX_API_SECRET
          valueFrom:
            secretKeyRef:
              name: alert-controller-org-{{ .OrganizationID }}
              key: api-secret
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8081
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8081
          initialDelaySeconds: 5
          periodSeconds: 5
```

### Organization Provisioning Flow

**When new organization is created**:
```
1. Org-Management-Controller receives organization creation event
2. Creates PostgreSQL schema: org_{id}
3. Runs database migrations for alert tables
4. Generates API credentials for alert-controller
5. Creates Kubernetes Secret with credentials
6. Creates Kubernetes Deployment for alert-controller-org-{id}
7. Waits for deployment to be ready
8. Validates alert-controller health
```

**Org-Management-Controller Integration**:
See [alert-controller-integration-spec.md](./alert-controller-integration-spec.md) for complete details on org-management-controller integration.

---

## Performance Characteristics

### Current Performance Baseline (Direct Database Access)

**Evaluation Cycle Performance** (100 rules × 50 servers = 5,000 evaluations):
- Cycle duration: 2-8 seconds
- Database queries: ~5,000 SELECT queries
- Database query time: ~1.5-6 seconds total
- Evaluation logic time: ~0.5-2 seconds
- Peak memory: ~150MB
- CPU usage: 15-25% of 1 core

**Alert Instance Operations**:
- Create instance: <10ms (database insert)
- Update instance: <10ms (database update)
- Query instances: 50-200ms (depends on filters)

### Projected Performance (API-First Architecture)

**Evaluation Cycle Performance**:
- Cycle duration: 5-15 seconds (+3-7 seconds from network latency)
- API calls: ~5,000 HTTP requests
- API call time: ~4-10 seconds total (assuming 1-2ms per call)
- Evaluation logic time: ~0.5-2 seconds (unchanged)
- Peak memory: ~200MB (+50MB for HTTP client)
- CPU usage: 20-35% of 1 core (+5-10% from HTTP overhead)

**Performance Mitigation Strategies**:

1. **Batch API Endpoints** (reduces API calls by 80%):
   ```
   Current: GET /servers/{id} × 50 = 50 API calls
   Improved: POST /servers/batch with IDs = 1 API call
   ```

2. **SDK Caching** (reduces redundant calls):
   ```go
   // Cache alert rules for 5 minutes
   client.Alerts.ListRules(ctx, &nexmonyx.ListOptions{
       Cache: &nexmonyx.CacheOptions{
           TTL:     5 * time.Minute,
           Enabled: true,
       },
   })
   ```

3. **Connection Pooling**:
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

4. **Parallel Evaluation**:
   ```go
   // Evaluate multiple rules concurrently
   var wg sync.WaitGroup
   semaphore := make(chan struct{}, 10)  // Limit concurrency to 10

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

**Projected Performance with Optimizations**:
- Cycle duration: 3-10 seconds (comparable to current)
- API calls: ~1,000 (80% reduction from batching)
- Peak memory: ~180MB
- CPU usage: 18-30% of 1 core

**Net Result**: Performance parity or better with proper optimization

---

## Dependencies & Integration Points

### Direct Dependencies

1. **Nexmonyx API Server** (CRITICAL)
   - Purpose: Primary data access layer
   - Communication: HTTP/HTTPS via go-sdk
   - Required Endpoints:
     - `GET /api/v1/alerts/rules` - List alert rules
     - `GET /api/v1/alerts/instances` - List alert instances
     - `POST /api/v1/alerts/instances` - Create alert instance
     - `PUT /api/v1/alerts/instances/{id}` - Update instance
     - `PUT /api/v1/alerts/instances/{id}/resolve` - Resolve instance
     - `GET /api/v1/servers/in-scope` - Get servers matching rule filters
     - `POST /api/v1/metrics/query` - Query metrics
     - `POST /api/v1/metrics/aggregate` - Aggregate metrics

2. **Notification-Service** (P1 Foundation Service)
   - Purpose: Multi-channel notification delivery
   - Communication: HTTP/HTTPS
   - Required Endpoints:
     - `POST /api/v1/notifications/send` - Send notification
     - `GET /api/v1/notifications/status/{id}` - Check delivery status

3. **PostgreSQL Database** (via API Server only after refactoring)
   - Purpose: Persistent storage
   - Access Pattern: Per-organization schema (org_{id})
   - Connection: Through API server ONLY (no direct access)

4. **Kubernetes** (Deployment Platform)
   - Purpose: Container orchestration
   - Integration: Deployment, Service, ConfigMap, Secret
   - Health checks: Liveness and Readiness probes

### Indirect Dependencies

1. **Org-Management-Controller**
   - Purpose: Organization lifecycle management
   - Integration: Receives organization provisioning events
   - Triggers: Creates alert-controller deployment for new orgs

2. **Prometheus + Grafana**
   - Purpose: Monitoring and observability
   - Integration: Exposes Prometheus metrics endpoint
   - Dashboards: Alert evaluation performance, notification delivery

3. **go-sdk** (SDK Layer)
   - Purpose: API client abstraction
   - Version: v2.4.0+ (requires enhancement with 8 new methods)
   - Repository: github.com/nexmonyx/go-sdk

---

## Related Documents

- **[alert-controller-api-first-refactoring.md](./alert-controller-api-first-refactoring.md)**: API-first refactoring strategy, SDK requirements, migration phases
- **[alert-controller-implementation-plan.md](./alert-controller-implementation-plan.md)**: Step-by-step implementation guide, timelines, resource requirements
- **[alert-controller-integration-spec.md](./alert-controller-integration-spec.md)**: Service integration contracts, API specifications, data flows

---

**Document Status**: Draft - Phase 1 Architecture Analysis Complete

**Next Steps**:
1. Review and approve architecture documentation
2. Begin SDK enhancement (Week 1 of implementation)
3. Implement API server endpoints (Week 2)
4. Refactor alert-controller to API-first (Week 3)
5. Deploy and validate (Week 4)
