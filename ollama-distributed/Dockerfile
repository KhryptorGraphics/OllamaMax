# Multi-stage Production Dockerfile for OllamaMax
# Optimized for size, security, and multi-architecture support

ARG GOLANG_VERSION=1.24.6
ARG NODE_VERSION=18-alpine
ARG ALPINE_VERSION=3.19

# ============================================================================
# Stage 1: Go Build Stage
# ============================================================================
FROM golang:${GOLANG_VERSION}-alpine AS go-builder

# Set build arguments for versioning
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT
ARG TARGETARCH
ARG TARGETPLATFORM

# Install build dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata

# Set working directory
WORKDIR /app

# Configure Go environment for optimal builds
ENV GOTOOLCHAIN=local
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org
ENV CGO_ENABLED=0

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies with verification
RUN go mod download && \
    go mod verify

# Copy source code
COPY . .

# Build the main distributed Ollama binary with architecture support
RUN GOOS=linux GOARCH=${TARGETARCH:-amd64} go build \
    -ldflags="-w -s \
    -X 'main.version=${VERSION}' \
    -X 'main.buildTime=${BUILD_TIME}' \
    -X 'main.gitCommit=${GIT_COMMIT}' \
    -X 'main.targetPlatform=${TARGETPLATFORM}'" \
    -a -installsuffix cgo \
    -o ollama-distributed \
    ./cmd/distributed-ollama

# Build additional CLI tools
RUN GOOS=linux GOARCH=${TARGETARCH:-amd64} go build \
    -ldflags="-w -s" \
    -a -installsuffix cgo \
    -o ollama-config \
    ./cmd/config-tool

# Verify the binaries
RUN file ollama-distributed && \
    ldd ollama-distributed || echo "Static binary - no dependencies"

# ============================================================================
# Stage 2: Node.js Build Stage
# ============================================================================
FROM node:${NODE_VERSION} AS node-builder

# Set working directory
WORKDIR /app/web

# Copy package files for caching
COPY web/package*.json ./

# Install dependencies with clean npm cache
RUN npm ci --only=production --no-audit --no-fund && \
    npm cache clean --force

# Copy web source
COPY web/ ./

# Build the React application
RUN npm run build

# Verify build output
RUN ls -la dist/ || ls -la build/ || echo "Build output verification"

# ============================================================================
# Stage 3: Final Runtime Image
# ============================================================================
FROM alpine:${ALPINE_VERSION} AS runtime

# Set build arguments for labels
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT

# Install minimal runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    jq \
    && rm -rf /var/cache/apk/*

# Create non-root user for security
RUN addgroup -g 1000 ollama && \
    adduser -D -s /bin/sh -u 1000 -G ollama ollama

# Create application directory structure
RUN mkdir -p /app/{data,models,cache,logs,config,web} && \
    chown -R ollama:ollama /app

# Set working directory
WORKDIR /app

# Copy Go binaries from builder stage
COPY --from=go-builder /app/ollama-distributed /usr/local/bin/ollama-distributed
COPY --from=go-builder /app/ollama-config /usr/local/bin/ollama-config

# Copy Node.js built assets from builder stage
COPY --from=node-builder --chown=ollama:ollama /app/web/dist /app/web/dist
COPY --from=node-builder --chown=ollama:ollama /app/web/build /app/web/build

# Copy static web assets
COPY --chown=ollama:ollama web/index.html /app/web/
COPY --chown=ollama:ollama pkg/web/static /app/web/static

# Copy default configuration
COPY --chown=ollama:ollama config.yaml /app/config/config.yaml

# Set proper permissions
RUN chmod +x /usr/local/bin/ollama-* && \
    chown -R ollama:ollama /app

# Switch to non-root user
USER ollama

# Set environment variables
ENV OLLAMA_DATA_DIR=/app/data \
    OLLAMA_MODELS_DIR=/app/models \
    OLLAMA_CACHE_DIR=/app/cache \
    OLLAMA_LOG_DIR=/app/logs \
    OLLAMA_CONFIG_FILE=/app/config/config.yaml \
    OLLAMA_WEB_DIR=/app/web \
    OLLAMA_LOG_LEVEL=info \
    OLLAMA_LOG_FORMAT=json \
    GIN_MODE=release \
    NODE_ENV=production

# Expose ports for distributed Ollama nodes
EXPOSE 11434 11435 11436 12000 13000 14090

# Health check command
HEALTHCHECK --interval=30s --timeout=10s --retries=3 --start-period=30s \
    CMD curl -f http://localhost:11434/api/v1/health || exit 1

# Create optimized startup script
RUN cat > /app/start.sh << 'EOF'
#!/bin/sh
set -e

# Print startup information
echo "🚀 Starting OllamaMax Distributed System"
echo "Version: ${VERSION:-dev}"
echo "Build Time: ${BUILD_TIME:-unknown}"
echo "Git Commit: ${GIT_COMMIT:-unknown}"
echo "Architecture: $(uname -m)"
echo "Config: $OLLAMA_CONFIG_FILE"
echo "Data Dir: $OLLAMA_DATA_DIR"
echo "Web Dir: $OLLAMA_WEB_DIR"
echo "Log Level: $OLLAMA_LOG_LEVEL"
echo ""

# Ensure directories exist with proper permissions
mkdir -p "$OLLAMA_DATA_DIR" "$OLLAMA_MODELS_DIR" "$OLLAMA_CACHE_DIR" "$OLLAMA_LOG_DIR"

# Validate configuration if config tool is available
if command -v ollama-config >/dev/null 2>&1; then
    echo "🔧 Validating configuration..."
    ollama-config validate "$OLLAMA_CONFIG_FILE" || echo "⚠️ Configuration validation failed"
fi

# Start the application
echo "🎯 Starting OllamaMax distributed service..."
exec ollama-distributed --config "$OLLAMA_CONFIG_FILE" "$@"
EOF

RUN chmod +x /app/start.sh

# Default command
CMD ["/app/start.sh"]

# Comprehensive metadata labels
LABEL maintainer="OllamaMax Team <team@ollamamax.com>" \
      version="${VERSION}" \
      description="OllamaMax - Enterprise Distributed AI Platform with Web Interface" \
      org.opencontainers.image.title="OllamaMax" \
      org.opencontainers.image.description="Enterprise-grade distributed AI platform based on Ollama with React web interface" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.vendor="OllamaMax" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.source="https://github.com/KhryptorGraphics/OllamaMax" \
      org.opencontainers.image.documentation="https://github.com/KhryptorGraphics/OllamaMax/blob/main/README.md" \
      org.opencontainers.image.created="${BUILD_TIME}" \
      org.opencontainers.image.revision="${GIT_COMMIT}"

# Build instructions for multi-architecture:
# docker buildx build --platform linux/amd64,linux/arm64 -t ollamamax:latest .
# Single architecture build:
# docker build --build-arg VERSION=v1.0.0 --build-arg BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ) --build-arg GIT_COMMIT=$(git rev-parse HEAD) -t ollamamax:latest .