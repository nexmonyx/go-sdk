# API Backend Requirements for Temperature and Power Monitoring

## Overview

The Linux Agent v1.3.0 has added comprehensive temperature and power supply monitoring capabilities. The Go SDK has been updated to support these new features. The API backend needs to be updated to accept, store, and process this new monitoring data.

## Required API Changes

### 1. Comprehensive Metrics Endpoint Update

**Endpoint:** `POST /v1/metrics/comprehensive`

The `ComprehensiveMetricsRequest` payload now includes two new optional fields:

```json
{
  "server_uuid": "string",
  "collected_at": "string",
  // ... existing fields ...
  "temperature": {
    "sensors": [
      {
        "sensor_id": "cpu_core_0",
        "sensor_name": "CPU Core 0",
        "temperature": 45.5,
        "status": "ok",  // ok, warning, critical
        "type": "cpu",   // cpu, system, disk, gpu, etc.
        "location": "processor",
        "upper_warning": 75.0,
        "upper_critical": 90.0
      }
    ]
  },
  "power": {
    "power_supplies": [
      {
        "id": "ps1",
        "name": "Power Supply 1",
        "status": "ok",  // ok, warning, critical, failed
        "power_watts": 196.0,
        "max_power_watts": 750.0,
        "voltage": 124.0,
        "current": 1.6,
        "efficiency": 94.5,
        "temperature": 38.0
      }
    ],
    "total_power_watts": 392.0
  }
}
```

### 2. Hardware Inventory Endpoint Update

**Endpoint:** `POST /v1/servers/{server_uuid}/hardware` (or similar)

The `HardwareInventoryInfo` structure now includes:

```json
{
  "hardware": {
    // ... existing fields ...
    "temperature_sensors": [
      {
        "sensor_id": "coretemp_package",
        "sensor_name": "CPU Package Temperature",
        "type": "cpu",
        "location": "processor",
        "max_temp": 100.0,
        "min_temp": 0.0
      }
    ],
    "power_supplies": [
      {
        // ... existing fields ...
        "current_power_watts": 196.0,
        "voltage": 124.0,
        "current": 1.6,
        "temperature": 38.0,
        "fan_speed": 2400,
        "input_voltage": 120.0,
        "output_voltage": 12.0
      }
    ]
  }
}
```

### 3. Database Schema Changes

#### Temperature Metrics Table (Time-Series)
```sql
CREATE TABLE temperature_metrics (
    server_uuid UUID NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    sensor_id VARCHAR(255) NOT NULL,
    sensor_name VARCHAR(255),
    temperature DOUBLE PRECISION NOT NULL,
    status VARCHAR(20),
    type VARCHAR(50),
    location VARCHAR(255),
    upper_warning DOUBLE PRECISION,
    upper_critical DOUBLE PRECISION,
    PRIMARY KEY (server_uuid, timestamp, sensor_id)
);

-- Hypertable for TimescaleDB
SELECT create_hypertable('temperature_metrics', 'timestamp');
```

#### Power Metrics Table (Time-Series)
```sql
CREATE TABLE power_metrics (
    server_uuid UUID NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    power_supply_id VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    status VARCHAR(20),
    power_watts DOUBLE PRECISION,
    max_power_watts DOUBLE PRECISION,
    voltage DOUBLE PRECISION,
    current DOUBLE PRECISION,
    efficiency DOUBLE PRECISION,
    temperature DOUBLE PRECISION,
    PRIMARY KEY (server_uuid, timestamp, power_supply_id)
);

-- Hypertable for TimescaleDB
SELECT create_hypertable('power_metrics', 'timestamp');
```

#### Hardware Inventory Updates
```sql
-- Add to existing hardware inventory tables or create new ones
CREATE TABLE hardware_temperature_sensors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    server_uuid UUID NOT NULL REFERENCES servers(uuid),
    sensor_id VARCHAR(255) NOT NULL,
    sensor_name VARCHAR(255),
    type VARCHAR(50),
    location VARCHAR(255),
    max_temp DOUBLE PRECISION,
    min_temp DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(server_uuid, sensor_id)
);

-- Update power_supplies table to include new monitoring fields
ALTER TABLE hardware_power_supplies
ADD COLUMN current_power_watts DOUBLE PRECISION,
ADD COLUMN voltage DOUBLE PRECISION,
ADD COLUMN current DOUBLE PRECISION,
ADD COLUMN temperature DOUBLE PRECISION,
ADD COLUMN fan_speed INTEGER,
ADD COLUMN input_voltage DOUBLE PRECISION,
ADD COLUMN output_voltage DOUBLE PRECISION;
```

### 4. Alert Integration Considerations

The new temperature and power metrics should integrate with the existing alerting system:

1. **Default Alert Rules** (to be added):
   - CPU Temperature Warning: > 75°C
   - CPU Temperature Critical: > 90°C
   - System Temperature Warning: > 60°C
   - Disk Temperature Warning: > 50°C
   - Power Supply Failure: status = "failed" or "critical"
   - High Power Consumption: total_power_watts > threshold

2. **Metric Names for Alerting**:
   - `temperature.cpu.max` - Maximum CPU temperature
   - `temperature.system.max` - Maximum system temperature
   - `temperature.disk.max` - Maximum disk temperature
   - `power.total_watts` - Total power consumption
   - `power.supply.status` - Power supply health status

### 5. API Response Considerations

1. **Aggregation Endpoints** - Add temperature/power to existing aggregated metrics:
   ```json
   GET /v1/servers/{server_uuid}/metrics/aggregated
   {
     // ... existing aggregations ...
     "temperature": {
       "max_temp": 68.0,
       "max_sensor": "GPU",
       "avg_cpu_temp": 45.5,
       "sensor_count": 12,
       "warnings": 0,
       "critical": 0
     },
     "power": {
       "total_watts": 392.0,
       "supply_count": 2,
       "failed_supplies": 0,
       "avg_efficiency": 94.35
     }
   }
   ```

2. **Historical Data Queries**:
   ```
   GET /v1/servers/{server_uuid}/metrics/temperature?start=2024-01-01&end=2024-01-02&sensor_id=cpu_core_0
   GET /v1/servers/{server_uuid}/metrics/power?start=2024-01-01&end=2024-01-02&aggregation=hourly
   ```

### 6. Validation Requirements

1. **Temperature Validation**:
   - Temperature values should be reasonable (-50°C to 150°C)
   - Status must be one of: ok, warning, critical
   - If thresholds are provided, validate temperature against them

2. **Power Validation**:
   - Power values must be non-negative
   - Efficiency should be between 0-100%
   - Status must be one of: ok, warning, critical, failed
   - Voltage/current values should be reasonable

### 7. Migration Considerations

1. All new fields are optional to maintain backward compatibility
2. Existing agents without temperature/power monitoring will continue to work
3. The API should gracefully handle missing temperature/power data
4. Consider adding feature flags for gradual rollout

### 8. Performance Considerations

1. **Data Retention**:
   - Temperature/power metrics generate significant time-series data
   - Consider retention policies (e.g., raw data for 7 days, hourly aggregates for 30 days)
   - Use TimescaleDB continuous aggregates for performance

2. **Bulk Insert Optimization**:
   - Metrics may include dozens of temperature sensors
   - Optimize for bulk inserts in time-series tables
   - Consider batching strategies

### 9. Example Implementation Priority

1. **Phase 1** (MVP):
   - Update comprehensive metrics endpoint to accept new fields
   - Store temperature/power data in time-series tables
   - Basic validation

2. **Phase 2**:
   - Hardware inventory integration
   - Alert rule integration
   - Basic aggregation endpoints

3. **Phase 3**:
   - Historical data queries
   - Advanced aggregations
   - Dashboard integration

## Testing Requirements

1. **Unit Tests**:
   - Validate all new data structures
   - Test threshold calculations
   - Verify status determination logic

2. **Integration Tests**:
   - Submit metrics with various temperature/power configurations
   - Test with missing optional fields
   - Verify time-series data storage
   - Test alert triggering

3. **Load Tests**:
   - Test with servers having many sensors (50+ temperature sensors)
   - High-frequency metric submissions
   - Concurrent metric submissions from multiple agents

## Example Test Data

```json
{
  "temperature": {
    "sensors": [
      {"sensor_id": "cpu_package", "sensor_name": "CPU Package", "temperature": 45.0, "status": "ok", "type": "cpu"},
      {"sensor_id": "cpu_core_0", "sensor_name": "CPU Core 0", "temperature": 43.0, "status": "ok", "type": "cpu"},
      {"sensor_id": "inlet_temp", "sensor_name": "Inlet Temp", "temperature": 26.0, "status": "ok", "type": "system"},
      {"sensor_id": "disk_sda", "sensor_name": "Disk /dev/sda", "temperature": 31.0, "status": "ok", "type": "disk"}
    ]
  },
  "power": {
    "power_supplies": [
      {
        "id": "ps1",
        "name": "Power Supply 1",
        "status": "ok",
        "power_watts": 196.0,
        "max_power_watts": 750.0,
        "voltage": 124.0,
        "current": 1.6,
        "efficiency": 94.5
      }
    ],
    "total_power_watts": 392.0
  }
}
```

## Questions for API Team

1. Should we create new dedicated endpoints for temperature/power metrics or update existing ones?
2. What retention policy should we use for time-series temperature/power data?
3. Should we automatically create default alert rules for temperature/power thresholds?
4. Do we need real-time WebSocket updates for critical temperature/power events?
5. Should we add temperature/power data to the existing health check endpoints?

## SDK Reference

The Go SDK implementation can be found at:
- Models: `models.go` lines 1044-1088 (new data structures)
- Helpers: `temperature_power_helpers.go` (utility functions)
- Examples: `examples/temperature_power/main.go` (usage examples)
- Tests: `temperature_power_test.go` (test cases)