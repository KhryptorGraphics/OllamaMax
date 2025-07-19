# Multi-stage Dockerfile for Ollamacron
# Builds a minimal production image with security best practices

# Build stage
FROM golang:1.21-alpine AS builder

# Set build arguments
ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
    -o ollamacron \
    ./cmd/ollamacron

# Production stage
FROM alpine:3.18

# Set metadata
LABEL maintainer="Ollama Team <team@ollama.ai>"
LABEL description="Ollamacron - Distributed Ollama Platform"
LABEL version="${VERSION}"

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    jq

# Create non-root user
RUN addgroup -g 1000 ollama && \
    adduser -D -s /bin/sh -u 1000 -G ollama ollama

# Create directories
RUN mkdir -p /app/data /app/models /app/cache /app/logs /app/config && \
    chown -R ollama:ollama /app

# Copy binary from builder
COPY --from=builder /build/ollamacron /usr/local/bin/ollamacron
RUN chmod +x /usr/local/bin/ollamacron

# Copy configuration
COPY --chown=ollama:ollama config/config.yaml /app/config/

# Set working directory
WORKDIR /app

# Switch to non-root user
USER ollama

# Expose ports
EXPOSE 11434 4001 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:11434/api/health || exit 1

# Default command
CMD ["ollamacron", "node", "--config", "/app/config/config.yaml"]