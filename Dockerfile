# Multi-stage Node.js Dockerfile for OllamaMax with Authentication System
# Build stage
FROM node:20-alpine AS builder

# Set up build environment
WORKDIR /app
RUN apk add --no-cache git ca-certificates python3 make g++

# Copy package files
COPY package*.json ./
COPY api-server/package*.json ./api-server/

# Install all dependencies (including dev dependencies for build)
RUN npm install --only=production
RUN cd api-server && npm install --only=production

# Copy source code
COPY . .

# Create production build if needed
RUN npm run build 2>/dev/null || echo "No build script found, continuing..."

# Runtime stage
FROM node:20-alpine

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata sqlite curl

# Create non-root user
RUN addgroup -g 1001 ollama && \
    adduser -D -s /bin/sh -u 1001 -G ollama ollama

# Set working directory
WORKDIR /app

# Copy application from builder stage
COPY --from=builder /app/package*.json ./
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/api-server ./api-server
COPY --from=builder /app/web-interface ./web-interface
COPY --from=builder /app/src ./src
COPY --from=builder /app/*.js ./
COPY --from=builder /app/*.json ./
COPY --from=builder /app/*.md ./

# Create directories for data persistence
RUN mkdir -p /app/data /app/logs /app/uploads

# Create startup script as root
RUN printf '#!/bin/sh\ncd /app/api-server && node server.js\n' > /app/start.sh && \
    chmod +x /app/start.sh

# Change ownership after creating all files
RUN chown -R ollama:ollama /app

# Switch to non-root user
USER ollama

# Expose ports
EXPOSE 13100 3000 8080

# Health check for the API server
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:13100/api/health || exit 1

# Set environment variables
ENV NODE_ENV=production
ENV PORT=13100
ENV API_PORT=13100
ENV WEB_PORT=3000
ENV SQLITE_DB_PATH=/app/data/users.db

# Run the application
CMD ["/app/start.sh"]