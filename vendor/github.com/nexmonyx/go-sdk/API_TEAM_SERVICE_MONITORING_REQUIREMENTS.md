# API Backend Requirements for Systemd Service Monitoring

## Overview

The Linux Agent v1.4.0 has added comprehensive systemd service monitoring capabilities. The Go SDK has been updated to support these features. The API backend needs to be updated to accept, store, and process service monitoring data.

## Required API Changes

### 1. Comprehensive Metrics Endpoint Update

**Endpoint:** `POST /v1/metrics/comprehensive`

The `ComprehensiveMetricsRequest` payload now includes a new optional field:

```json
{
  "server_uuid": "string",
  "collected_at": "string",
  // ... existing fields ...
  "services": {
    "services": [
      {
        "name": "ssh.service",
        "state": "active",
        "sub_state": "running",
        "load_state": "loaded",
        "description": "OpenBSD Secure Shell server",
        "main_pid": 163754,
        "memory_current": 4308992,
        "cpu_usage_nsec": 890000000,
        "tasks_current": 1,
        "restart_count": 0,
        "active_since": "2024-01-13T10:30:45Z"
      }
    ],
    "metrics": [
      {
        "service_name": "ssh.service",
        "timestamp": "2024-01-13T15:45:30Z",
        "cpu_percent": 0.1,
        "memory_rss": 4308992,
        "process_count": 1,
        "thread_count": 1
      }
    ],
    "logs": {
      "ssh.service": [
        {
          "timestamp": "2024-01-13T15:44:12Z",
          "level": "info",
          "message": "Accepted publickey for user from 192.168.1.100",
          "fields": {
            "pid": "163754",
            "unit": "ssh.service"
          }
        }
      ]
    }
  }
}
```

### 2. Hardware Inventory Endpoint Update

**Endpoint:** `POST /v1/servers/{server_uuid}/hardware`

The `HardwareInventoryInfo` structure now includes:

```json
{
  "hardware": {
    // ... existing fields ...
    "services": {
      "services": [...],
      "metrics": [...],
      "logs": {...}
    }
  }
}
```

### 3. New Service-Specific Endpoints

#### a. Submit Service Data
**Endpoint:** `POST /v1/servers/{server_uuid}/services`
```json
{
  "server_uuid": "string",
  "services": {
    "services": [...],
    "metrics": [...],
    "logs": {...}
  }
}
```

#### b. Submit Service Metrics (Time-Series)
**Endpoint:** `POST /v1/servers/{server_uuid}/services/metrics`
```json
{
  "server_uuid": "string",
  "metrics": [
    {
      "service_name": "nginx.service",
      "timestamp": "2024-01-13T15:45:30Z",
      "cpu_percent": 2.5,
      "memory_rss": 134217728,
      "process_count": 4,
      "thread_count": 4
    }
  ]
}
```

#### c. Submit Service Logs
**Endpoint:** `POST /v1/servers/{server_uuid}/services/logs`
```json
{
  "server_uuid": "string",
  "logs": {
    "service_name": [
      {
        "timestamp": "2024-01-13T15:44:12Z",
        "level": "error",
        "message": "Connection refused",
        "fields": {...}
      }
    ]
  }
}
```

#### d. Get Server Services Status
**Endpoint:** `GET /v1/servers/{server_uuid}/services`
**Response:**
```json
{
  "server_uuid": "string",
  "last_updated": "2024-01-13T16:00:00Z",
  "services": [...],
  "summary": {
    "total": 75,
    "active": 68,
    "inactive": 5,
    "failed": 2,
    "state_counts": {
      "active": 68,
      "inactive": 5,
      "failed": 2
    }
  }
}
```

#### e. Get Service History
**Endpoint:** `GET /v1/servers/{server_uuid}/services/{service_name}/history`
**Query Parameters:** `start_date`, `end_date`, `limit`
**Response:**
```json
{
  "server_uuid": "string",
  "service_name": "nginx.service",
  "history": [
    {
      "timestamp": "2024-01-13T16:00:00Z",
      "state": "active",
      "sub_state": "running",
      "cpu_percent": 2.5,
      "memory_bytes": 134217728,
      "restart_count": 0
    }
  ]
}
```

#### f. Get Service Logs
**Endpoint:** `GET /v1/servers/{server_uuid}/services/{service_name}/logs`
**Query Parameters:** `start_date`, `end_date`, `limit`, `level`

#### g. Request Service Restart (Optional)
**Endpoint:** `POST /v1/servers/{server_uuid}/services/{service_name}/restart`

#### h. Get Failed Services (Organization-wide)
**Endpoint:** `GET /v1/organizations/{org_uuid}/services/failed`

#### i. Create Service Alert Rules
**Endpoint:** `POST /v1/organizations/{org_uuid}/alerts/services`
```json
{
  "name": "Critical Service Failed",
  "description": "Alert when critical services fail",
  "service_patterns": ["ssh*", "nginx*", "mysql*"],
  "conditions": ["state=failed", "restart_count>3"],
  "severity": "critical",
  "enabled": true
}
```

### 4. Database Schema Changes

#### Service State Table
```sql
CREATE TABLE service_states (
    id BIGSERIAL PRIMARY KEY,
    server_uuid UUID NOT NULL REFERENCES servers(uuid),
    service_name VARCHAR(255) NOT NULL,
    state VARCHAR(50) NOT NULL,
    sub_state VARCHAR(50),
    load_state VARCHAR(50),
    description TEXT,
    main_pid INTEGER,
    memory_current BIGINT,
    cpu_usage_nsec BIGINT,
    tasks_current INTEGER,
    restart_count INTEGER DEFAULT 0,
    active_since TIMESTAMPTZ,
    last_updated TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(server_uuid, service_name)
);

CREATE INDEX idx_service_states_server ON service_states(server_uuid);
CREATE INDEX idx_service_states_state ON service_states(state);
```

#### Service Metrics Table (Time-Series)
```sql
CREATE TABLE service_metrics (
    server_uuid UUID NOT NULL,
    service_name VARCHAR(255) NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    cpu_percent DECIMAL(5,2),
    memory_rss BIGINT,
    process_count INTEGER,
    thread_count INTEGER,
    PRIMARY KEY (server_uuid, service_name, timestamp)
);

-- Hypertable for TimescaleDB
SELECT create_hypertable('service_metrics', 'timestamp');

-- Continuous aggregate for hourly stats
CREATE MATERIALIZED VIEW service_metrics_hourly
WITH (timescaledb.continuous) AS
SELECT 
    server_uuid,
    service_name,
    time_bucket('1 hour', timestamp) AS hour,
    AVG(cpu_percent) as avg_cpu_percent,
    MAX(cpu_percent) as max_cpu_percent,
    AVG(memory_rss) as avg_memory_rss,
    MAX(memory_rss) as max_memory_rss
FROM service_metrics
GROUP BY server_uuid, service_name, hour;
```

#### Service Logs Table
```sql
CREATE TABLE service_logs (
    id BIGSERIAL PRIMARY KEY,
    server_uuid UUID NOT NULL,
    service_name VARCHAR(255) NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    level VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    fields JSONB,
    cursor VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_service_logs_server_service_time 
    ON service_logs(server_uuid, service_name, timestamp DESC);
CREATE INDEX idx_service_logs_level 
    ON service_logs(level) WHERE level IN ('error', 'err', 'critical');
CREATE UNIQUE INDEX idx_service_logs_cursor 
    ON service_logs(server_uuid, cursor) WHERE cursor IS NOT NULL;

-- Partition by time for better performance
CREATE TABLE service_logs_y2024m01 PARTITION OF service_logs
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
```

#### Service State History Table
```sql
CREATE TABLE service_state_history (
    id BIGSERIAL PRIMARY KEY,
    server_uuid UUID NOT NULL,
    service_name VARCHAR(255) NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    old_state VARCHAR(50),
    new_state VARCHAR(50),
    old_sub_state VARCHAR(50),
    new_sub_state VARCHAR(50),
    restart_count INTEGER,
    details JSONB
);

CREATE INDEX idx_service_state_history_lookup 
    ON service_state_history(server_uuid, service_name, timestamp DESC);
```

### 5. Alert Integration

#### Default Alert Rules
```yaml
- name: "Service Failed"
  condition: "state = 'failed'"
  severity: "critical"
  
- name: "Service Flapping"
  condition: "restart_count > 3 in 1 hour"
  severity: "warning"
  
- name: "High Service Memory"
  condition: "memory_rss > 2GB"
  severity: "warning"
  
- name: "Service CPU Spike"
  condition: "cpu_percent > 80 for 5 minutes"
  severity: "warning"
```

#### Metric Names for Alerting
- `service.state` - Service state (active/inactive/failed)
- `service.cpu_percent` - CPU usage percentage
- `service.memory_bytes` - Memory usage in bytes
- `service.restart_count` - Number of restarts
- `service.uptime_seconds` - Service uptime

### 6. Performance Considerations

1. **Data Volume**:
   - Servers may have 50-100+ services
   - Metrics collected every 60 seconds
   - Logs can be high volume for active services

2. **Optimization Strategies**:
   - Use bulk inserts for metrics
   - Implement log deduplication using cursor field
   - Use TimescaleDB continuous aggregates
   - Implement data retention policies:
     - Raw metrics: 7 days
     - Hourly aggregates: 30 days
     - Daily aggregates: 1 year
     - Logs: 3 days (configurable)

3. **Caching**:
   - Cache service state for 30 seconds
   - Cache service lists per server for 60 seconds

### 7. API Response Examples

#### Service Health Summary
```json
GET /v1/servers/{server_uuid}/services/health

{
  "server_uuid": "...",
  "timestamp": "2024-01-13T16:00:00Z",
  "health_score": 94,  // 0-100
  "issues": [
    {
      "service": "nginx.service",
      "issue": "failed",
      "severity": "critical"
    },
    {
      "service": "mysql.service",
      "issue": "high_restart_count",
      "severity": "warning",
      "details": {"restart_count": 5}
    }
  ]
}
```

#### Service Dependencies (Future Enhancement)
```json
GET /v1/servers/{server_uuid}/services/{service_name}/dependencies

{
  "service": "nginx.service",
  "depends_on": ["network.target", "remote-fs.target"],
  "required_by": ["multi-user.target"],
  "wants": ["php-fpm.service"],
  "conflicts": ["apache2.service"]
}
```

### 8. Migration and Rollout

1. **Phase 1**: Accept service data in comprehensive metrics
2. **Phase 2**: Implement dedicated service endpoints
3. **Phase 3**: Add alerting and historical queries
4. **Phase 4**: Advanced features (dependencies, predictions)

### 9. Security Considerations

1. **Log Sanitization**:
   - Remove sensitive data from logs (passwords, tokens)
   - Implement configurable log filters

2. **Access Control**:
   - Service restart requires admin permissions
   - Log access follows server access permissions

3. **Rate Limiting**:
   - Limit log submissions to prevent abuse
   - Implement per-service metrics limits

### 10. Testing Requirements

1. **Load Testing**:
   - 1000 servers × 75 services × metrics/minute
   - Log ingestion at 10K logs/second

2. **Integration Tests**:
   - Service state transitions
   - Alert triggering
   - Log deduplication

3. **Edge Cases**:
   - Services with special characters in names
   - Very long service descriptions
   - Rapid state changes

## SDK Implementation Reference

The Go SDK implementation includes:
- Models: `models.go` lines 1102-1152 (service data structures)
- Helpers: `service_monitoring_helpers.go` (utility functions)
- API Methods: `service_monitoring_api.go` (client methods)
- Examples: `examples/service_monitoring/main.go`
- Tests: `service_monitoring_test.go`

## Questions for API Team

1. Should service logs be stored in the main database or a separate log storage system?
2. What retention policies should apply to service metrics and logs?
3. Should we implement service dependency tracking in phase 1?
4. Do we need real-time WebSocket updates for service state changes?
5. Should service restart functionality be included, or leave it to external tools?
6. What level of log filtering/sanitization is required?
7. Should we aggregate metrics at ingestion or query time?