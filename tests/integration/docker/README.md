# Mock API Server - Docker Container

This directory contains the Docker containerization of the mock API server for the Nexmonyx Go SDK integration testing framework.

## Overview

The mock API server is packaged as a standalone Docker container that can be:
- Run locally via `docker-compose`
- Deployed in CI/CD environments
- Used for development without the need to run integration tests directly

## Files

- `Dockerfile` - Multi-stage build for the mock API server
- `cmd/main.go` - Containerized entry point for the mock API server

## Building the Image

### Local Build

```bash
# From repository root
docker build -f tests/integration/docker/Dockerfile \
  -t nexmonyx-mock-api:latest .
```

### From Docker Compose

```bash
# Build and start
docker-compose up --build -d

# Just build
docker-compose build --no-cache
```

## Running the Container

### Standalone

```bash
docker run -p 8080:8080 \
  -e API_PORT=8080 \
  -e AUTH_TOKEN=test-token \
  nexmonyx-mock-api:latest
```

### With Environment Variables

```bash
docker run -p 8080:8080 \
  -e API_PORT=8080 \
  -e API_HOST=0.0.0.0 \
  -e AUTH_TOKEN=my-token \
  -e API_LOG_LEVEL=debug \
  nexmonyx-mock-api:latest
```

### With Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f mock-api

# Stop services
docker-compose down
```

## Configuration

### Environment Variables

| Variable | Default | Purpose |
|----------|---------|---------|
| `API_PORT` | `8080` | Server listening port |
| `API_HOST` | `0.0.0.0` | Server listening interface |
| `AUTH_TOKEN` | `test-token` | Bearer token for authentication |
| `API_LOG_LEVEL` | `info` | Logging level (debug/info/warn/error) |
| `FIXTURES_PATH` | `/app/tests/integration/fixtures` | Path to JSON fixtures |

### Health Checks

The container includes a health check endpoint:

```bash
# Direct request
curl http://localhost:8080/health

# Docker health status
docker ps --format "{{.Names}} {{.Status}}"
```

Readiness check (ensures service is operational):

```bash
curl http://localhost:8080/ready
```

## Development

### Making Changes

1. **Edit source code**:
   - Modify `cmd/main.go` for server logic
   - Update mock handlers as needed

2. **Rebuild Docker image**:
   ```bash
   docker-compose build --no-cache mock-api
   ```

3. **Restart service**:
   ```bash
   docker-compose restart mock-api
   ```

4. **View logs**:
   ```bash
   docker-compose logs -f mock-api
   ```

### Debugging

#### Access container shell

```bash
docker-compose exec mock-api sh
```

#### View environment variables

```bash
docker-compose exec mock-api env | sort
```

#### Test endpoints directly

```bash
# From host
curl -H "Authorization: Bearer test-token" \
     http://localhost:8080/api/v1/servers

# From inside container
docker-compose exec mock-api curl \
  -H "Authorization: Bearer test-token" \
  http://localhost:8080/api/v1/servers
```

#### Enable debug logging

```bash
# Update docker-compose.yml or run directly
docker run -p 8080:8080 \
  -e API_LOG_LEVEL=debug \
  nexmonyx-mock-api:latest
```

## Performance

### Image Size

The multi-stage build optimizes for minimal image size:
- Builder stage: Full Go toolchain
- Runtime stage: Alpine Linux base (~5MB)
- Final image: ~15-20MB

### Container Resources

Recommended resource allocation:

```yaml
# docker-compose.yml
services:
  mock-api:
    # ...
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 256M
        reservations:
          cpus: '0.25'
          memory: 128M
```

### Performance Tips

1. **Disable debug logging** in production/CI:
   ```bash
   -e API_LOG_LEVEL=warn
   ```

2. **Use Alpine base** for smaller images

3. **Run in parallel** for stress testing:
   ```bash
   for i in {1..5}; do
     docker run -p $((8080 + i)):8080 \
       nexmonyx-mock-api:latest &
   done
   ```

## Networking

### Docker Compose Network

Services communicate via `nexmonyx-test-network`:

```bash
# View network
docker network inspect nexmonyx-test-network

# Connect new container
docker run --network nexmonyx-test-network \
  -e INTEGRATION_TEST_API_URL=http://mock-api:8080 \
  nexmonyx-mock-api:latest
```

### Port Mapping

```yaml
# Local development - one service
ports:
  - "8080:8080"

# Multiple services
ports:
  - "8080:8080"  # Instance 1
  - "8081:8080"  # Instance 2
```

## Troubleshooting

### Container fails to start

```bash
# Check logs
docker-compose logs mock-api

# Inspect image
docker image inspect nexmonyx-mock-api:latest

# Rebuild from scratch
docker-compose down -v
docker-compose build --no-cache
docker-compose up -d
```

### Port conflicts

```bash
# Find what's using the port
lsof -i :8080

# Use alternate port
docker run -p 8081:8080 nexmonyx-mock-api:latest
```

### Authentication fails

```bash
# Verify token is set
docker-compose exec mock-api env | grep AUTH

# Try with correct token
curl -H "Authorization: Bearer test-token" \
     http://localhost:8080/api/v1/servers
```

### Slow response times

```bash
# Check container resources
docker stats mock-api

# View logs for errors
docker-compose logs -f mock-api

# Monitor network
docker-compose exec mock-api netstat -an
```

## Registry Deployment

### Building for Registry

```bash
# Tag for registry
docker tag nexmonyx-mock-api:latest \
  docker.io/yourusername/nexmonyx-mock-api:latest

# Push to registry
docker push docker.io/yourusername/nexmonyx-mock-api:latest
```

### Using from Registry

```yaml
# docker-compose.yml
services:
  mock-api:
    image: docker.io/yourusername/nexmonyx-mock-api:latest
```

## Reference

- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Multi-Stage Builds](https://docs.docker.com/build/building/multi-stage/)
- [Integration Testing Guide](../../docs/INTEGRATION_TESTING.md)

