# SDK Builder Image
FROM golang:1.24.4-alpine AS sdk-builder

# Install git and necessary tools
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /sdk

# Copy SDK files
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build a dummy binary to ensure all dependencies are downloaded
RUN go build -v ./...

# Final SDK image with just the source code
FROM alpine:latest AS sdk-source

# Copy SDK source files
COPY --from=sdk-builder /sdk /sdk/nexmonyx

# This image will be used as a build context in the API Dockerfile