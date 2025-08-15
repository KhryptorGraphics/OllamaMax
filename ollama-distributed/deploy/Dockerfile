# Multi-stage build for OllamaMax Distributed
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make build-base

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ollama-distributed ./cmd/distributed

# Production stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates curl jq

# Create app user
RUN addgroup -g 1001 -S ollama && \
    adduser -u 1001 -S ollama -G ollama

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/ollama-distributed .

# Copy configuration files
COPY --from=builder /app/config ./config
COPY --from=builder /app/deploy/entrypoint.sh ./entrypoint.sh

# Create necessary directories
RUN mkdir -p /app/data /app/models /app/logs && \
    chown -R ollama:ollama /app

# Make entrypoint executable
RUN chmod +x ./entrypoint.sh

# Switch to app user
USER ollama

# Expose ports
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Set entrypoint
ENTRYPOINT ["./entrypoint.sh"]
