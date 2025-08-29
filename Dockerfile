# Build stage
FROM golang:1.21-alpine AS builder

# Set up build environment
WORKDIR /app
RUN apk add --no-cache git ca-certificates

# Copy go modules files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/ollamamax .

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 ollama && \
    adduser -D -s /bin/sh -u 1001 -G ollama ollama

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/bin/ollamamax .
COPY --from=builder /app/config.yaml ./config.yaml

# Change ownership
RUN chown -R ollama:ollama /app

# Switch to non-root user
USER ollama

# Expose ports (using ports above 11111 as requested)
EXPOSE 11434

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:11434/health || exit 1

# Set environment variables
ENV OLLAMA_HOST=0.0.0.0
ENV OLLAMA_PORT=11434

# Run the application
CMD ["./ollamamax"]