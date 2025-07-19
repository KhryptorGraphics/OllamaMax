# Multi-stage Dockerfile for Ollamacron
# Optimized for production deployment with minimal image size

# Build stage
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X 'main.version=$(git describe --tags --always)' -X 'main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)'" \
    -o ollamacron \
    ./cmd/node

# Runtime stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    curl \
    jq \
    tzdata \
    && update-ca-certificates

# Create non-root user
RUN addgroup -g 1000 -S ollamacron && \
    adduser -u 1000 -S ollamacron -G ollamacron

# Create directories
RUN mkdir -p /app/data /app/logs /app/config && \
    chown -R ollamacron:ollamacron /app

# Copy binary from builder
COPY --from=builder /app/ollamacron /usr/local/bin/ollamacron

# Copy configuration
COPY --chown=ollamacron:ollamacron config/node.yaml /app/config/config.yaml

# Switch to non-root user
USER ollamacron

# Set working directory
WORKDIR /app

# Expose ports
EXPOSE 8080 9000 9090 8081

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8081/health || exit 1

# Environment variables
ENV OLLAMACRON_CONFIG="/app/config/config.yaml"
ENV OLLAMACRON_DATA_DIR="/app/data"
ENV OLLAMACRON_LOG_DIR="/app/logs"

# Set entrypoint
ENTRYPOINT ["ollamacron"]
CMD ["server", "--config", "/app/config/config.yaml"]