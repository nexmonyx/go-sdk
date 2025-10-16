# Multi-stage build for mock API server

# Stage 1: Build
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go module files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy application source
COPY . .

# Build the mock API server binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/mock-api \
    ./tests/integration/docker/cmd/main.go

# Stage 2: Runtime
FROM alpine:3.19

RUN apk add --no-cache ca-certificates curl

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/mock-api /app/mock-api

# Copy fixtures for embedded loading
COPY --from=builder /app/tests/integration/fixtures /app/tests/integration/fixtures

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /app

USER appuser

# Health check
HEALTHCHECK --interval=5s --timeout=3s --retries=5 --start-period=10s \
    CMD curl -f http://localhost:${API_PORT:-8080}/health || exit 1

# Expose API port
EXPOSE 8080

# Run mock API server
ENTRYPOINT ["/app/mock-api"]
