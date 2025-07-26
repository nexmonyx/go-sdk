# Monitoring Agent Examples

This directory contains examples demonstrating how to use the Nexmonyx Go SDK to build monitoring agents.

## Prerequisites

1. **Monitoring Key**: You need a monitoring key (starting with `MON_`) from your Nexmonyx organization
2. **Go 1.24+**: These examples require Go 1.24 or later
3. **Network Access**: The monitoring agent needs to reach the Nexmonyx API

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `NEXMONYX_MONITORING_KEY` | Yes | - | Your monitoring key (MON_...) |
| `NEXMONYX_API_ENDPOINT` | No | `https://api.nexmonyx.com` | API server URL |
| `NEXMONYX_REGION` | No | `us-east-1` | Region for probe assignments |
| `NEXMONYX_AGENT_ID` | No | Auto-generated | Unique agent identifier |
| `DEBUG` | No | `false` | Enable debug logging |

## Examples

### 1. Basic Agent (`basic_agent.go`)

A simple monitoring agent that demonstrates the core functionality:

- Authentication with MON_ key
- Fetching assigned probes for a region
- Sending heartbeats with node information
- Submitting probe execution results

**Usage:**

```bash
export NEXMONYX_MONITORING_KEY="MON_your_monitoring_key_here"
export NEXMONYX_REGION="us-east-1"
go run basic_agent.go
```

**Features:**
- Simple probe simulation
- Basic error handling
- Health check validation
- Mock probe execution results

### 2. Advanced Agent (`advanced_agent.go`)

A more sophisticated monitoring agent with production-ready features:

- Concurrent probe execution
- Automatic probe refresh
- Graceful shutdown handling
- Statistics tracking
- Proper error handling and retries

**Usage:**

```bash
export NEXMONYX_MONITORING_KEY="MON_your_monitoring_key_here"
export NEXMONYX_REGION="us-east-1"
export NEXMONYX_AGENT_ID="my-production-agent"
go run advanced_agent.go
```

**Features:**
- Multi-threaded probe execution
- Real-time statistics
- Periodic heartbeats (30s intervals)
- Probe refresh (5 minute intervals)
- Signal handling for graceful shutdown
- Different probe type simulations (HTTP, TCP, ICMP)

## Building and Running

### Build Examples

```bash
# Build basic agent
go build -o basic_agent basic_agent.go

# Build advanced agent
go build -o advanced_agent advanced_agent.go
```

### Run with Docker

Create a `Dockerfile`:

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o monitoring_agent advanced_agent.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/monitoring_agent .
CMD ["./monitoring_agent"]
```

Build and run:

```bash
docker build -t nexmonyx-monitoring-agent .
docker run -e NEXMONYX_MONITORING_KEY=MON_your_key_here nexmonyx-monitoring-agent
```

## SDK Features Used

### Authentication

```go
// Create client with monitoring key authentication
client, err := nexmonyx.NewMonitoringAgentClient(&nexmonyx.Config{
    BaseURL: "https://api.nexmonyx.com",
    Auth: nexmonyx.AuthConfig{
        MonitoringKey: "MON_your_monitoring_key_here",
    },
})
```

### Probe Management

```go
// Get probes assigned to this region
probes, err := client.Monitoring.GetAssignedProbes(ctx, "us-east-1")

// Submit probe execution results
results := []nexmonyx.ProbeExecutionResult{...}
err = client.Monitoring.SubmitResults(ctx, results)
```

### Heartbeats

```go
// Send heartbeat with node information
nodeInfo := nexmonyx.NodeInfo{
    AgentID:      "my-agent",
    AgentVersion: "1.0.0",
    Region:       "us-east-1",
    Status:       "healthy",
    // ... more fields
}
err = client.Monitoring.Heartbeat(ctx, nodeInfo)
```

## Error Handling

The SDK provides structured error types:

```go
if err := client.HealthCheck(ctx); err != nil {
    switch e := err.(type) {
    case *nexmonyx.UnauthorizedError:
        log.Fatal("Invalid monitoring key")
    case *nexmonyx.RateLimitError:
        log.Printf("Rate limited, retry after: %s", e.RetryAfter)
    default:
        log.Printf("API error: %v", err)
    }
}
```

## Testing

### Unit Tests

```bash
go test ./...
```

### Integration Tests

Set up environment variables and run:

```bash
export NEXMONYX_MONITORING_KEY="MON_test_key"
export NEXMONYX_API_ENDPOINT="https://api-dev.nexmonyx.com"
go test -tags=integration ./...
```

## Production Deployment

### Systemd Service

Create `/etc/systemd/system/nexmonyx-monitoring-agent.service`:

```ini
[Unit]
Description=Nexmonyx Monitoring Agent
After=network.target

[Service]
Type=simple
User=nexmonyx
WorkingDirectory=/opt/nexmonyx
ExecStart=/opt/nexmonyx/monitoring_agent
Restart=always
RestartSec=30
Environment=NEXMONYX_MONITORING_KEY=MON_your_key_here
Environment=NEXMONYX_REGION=us-east-1

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl enable nexmonyx-monitoring-agent
sudo systemctl start nexmonyx-monitoring-agent
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nexmonyx-monitoring-agent
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nexmonyx-monitoring-agent
  template:
    metadata:
      labels:
        app: nexmonyx-monitoring-agent
    spec:
      containers:
      - name: monitoring-agent
        image: nexmonyx/monitoring-agent:latest
        env:
        - name: NEXMONYX_MONITORING_KEY
          valueFrom:
            secretKeyRef:
              name: nexmonyx-secret
              key: monitoring-key
        - name: NEXMONYX_REGION
          value: "us-east-1"
```

## Troubleshooting

### Common Issues

1. **Authentication Errors**
   - Verify monitoring key is correct and starts with `MON_`
   - Check if key has been revoked or expired
   - Ensure proper region assignment

2. **Network Issues**
   - Verify connectivity to API endpoint
   - Check firewall rules
   - Test with `curl` or `wget`

3. **No Probes Assigned**
   - Check if probes are configured for your region
   - Verify organization has active probes
   - Check probe assignment rules

### Debug Mode

Enable debug mode to see detailed request/response information:

```bash
export DEBUG=true
go run advanced_agent.go
```

### Monitoring Agent Health

The agent sends heartbeats every 30 seconds. Monitor the logs for:

- Successful heartbeats
- Probe execution statistics
- Error messages
- Performance metrics

## Support

For issues and questions:

1. Check the [SDK documentation](../README.md)
2. Review the [API documentation](https://docs.nexmonyx.com)
3. Contact support at support@nexmonyx.com