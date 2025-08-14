# Production Dockerfile for OllamaMax
# Multi-stage build optimized for production deployment

# Build stage
FROM golang:1.24.5 AS builder

# Set build arguments
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT

# No additional build dependencies installed; using stock golang image

# Set working directory
ENV GOTOOLCHAIN=local
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies with retry and timeout handling
RUN go env -w GOPROXY=https://proxy.golang.org,direct && \
    go env -w GOSUMDB=sum.golang.org && \
    go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s \
    -X 'main.version=${VERSION}' \
    -X 'main.buildTime=${BUILD_TIME}' \
    -X 'main.gitCommit=${GIT_COMMIT}'" \
    -a -installsuffix cgo \
    -o ollama-distributed \
    ./cmd/node

# Verify the binary disabled (no 'version' command available)

# Runtime stage
FROM alpine:3.19 AS runtime

# Minimize runtime dependencies to avoid flaky apk fetches in CI/build environments
# No additional packages installed in runtime image

# Create non-root user for security
RUN addgroup -g 1000 ollama && \
    adduser -D -s /bin/sh -u 1000 -G ollama ollama

# Create application directories
RUN mkdir -p /app/data /app/models /app/cache /app/logs /app/config && \
    chown -R ollama:ollama /app

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/ollama-distributed /usr/local/bin/ollama-distributed

# Copy web assets
COPY --from=builder /app/web /app/web
COPY --from=builder /app/pkg/web/static /app/pkg/web/static

# Copy default configuration
COPY --from=builder /app/config/production.yaml /app/config/config.yaml

# Set proper permissions
RUN chmod +x /usr/local/bin/ollama-distributed && \
    chown -R ollama:ollama /app

# Switch to non-root user
USER ollama

# Set environment variables
ENV OLLAMA_DATA_DIR=/app/data
ENV OLLAMA_MODELS_DIR=/app/models
ENV OLLAMA_CACHE_DIR=/app/cache
ENV OLLAMA_LOG_DIR=/app/logs
ENV OLLAMA_CONFIG_FILE=/app/config/config.yaml
ENV OLLAMA_LOG_LEVEL=info
ENV OLLAMA_LOG_FORMAT=json
ENV GIN_MODE=release

# Expose ports
EXPOSE 8080 8081 9000 9090

# HEALTHCHECK disabled in runtime image to avoid dependency on curl
# HEALTHCHECK NONE

# Create startup script with proper newlines
RUN cat > /app/start.sh << 'EOF'
#!/bin/sh
set -e

# Print startup information
echo "Starting OllamaMax Distributed System"
ollama-distributed version || true
echo "Config: $OLLAMA_CONFIG_FILE"
echo "Data Dir: $OLLAMA_DATA_DIR"
echo "Log Level: $OLLAMA_LOG_LEVEL"
echo ""

# Ensure directories exist
mkdir -p "$OLLAMA_DATA_DIR" "$OLLAMA_MODELS_DIR" "$OLLAMA_CACHE_DIR" "$OLLAMA_LOG_DIR"

# Start the application
exec ollama-distributed start --config "$OLLAMA_CONFIG_FILE" "$@"
EOF
RUN chmod +x /app/start.sh

# Default command
CMD ["/app/start.sh"]

# Labels for metadata
LABEL maintainer="OllamaMax Team <team@ollamamax.com>"
LABEL version="${VERSION}"
LABEL description="OllamaMax - Enterprise Distributed AI Platform"
LABEL org.opencontainers.image.title="OllamaMax"
LABEL org.opencontainers.image.description="Enterprise-grade distributed AI platform based on Ollama"
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.vendor="OllamaMax"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.source="https://github.com/KhryptorGraphics/OllamaMax"
LABEL org.opencontainers.image.documentation="https://github.com/KhryptorGraphics/OllamaMax/blob/main/README.md"
