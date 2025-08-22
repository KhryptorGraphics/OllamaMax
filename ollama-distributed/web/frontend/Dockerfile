# Multi-stage production build for Ollama Distributed Frontend
# Optimized for performance and security

# Stage 1: Dependencies
FROM node:20-alpine AS deps
WORKDIR /app

# Install dependencies with cache mount
COPY package*.json ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci --only=production

# Stage 2: Builder
FROM node:20-alpine AS builder
WORKDIR /app

# Copy dependencies from deps stage
COPY --from=deps /app/node_modules ./node_modules
COPY package*.json ./

# Install dev dependencies
RUN --mount=type=cache,target=/root/.npm \
    npm ci

# Copy source code
COPY . .

# Build arguments for environment-specific builds
ARG NODE_ENV=production
ARG API_URL=http://localhost:8080
ARG WS_URL=ws://localhost:8080

# Set environment variables for build
ENV NODE_ENV=${NODE_ENV}
ENV VITE_API_URL=${API_URL}
ENV VITE_WS_URL=${WS_URL}

# Build the application
RUN npm run build && \
    npm run test:unit && \
    npm run lint

# Stage 3: Runtime
FROM nginx:alpine AS runtime

# Install curl for health checks
RUN apk add --no-cache curl

# Security: Create non-root user
RUN addgroup -g 1001 -S nodejs && \
    adduser -S nodejs -u 1001

# Copy custom nginx configuration
COPY --chown=nodejs:nodejs nginx.conf /etc/nginx/nginx.conf
COPY --chown=nodejs:nodejs nginx-security.conf /etc/nginx/conf.d/security.conf

# Copy built application from builder stage
COPY --from=builder --chown=nodejs:nodejs /app/dist /usr/share/nginx/html

# Copy health check script
COPY --chown=nodejs:nodejs healthcheck.sh /usr/local/bin/healthcheck.sh
RUN chmod +x /usr/local/bin/healthcheck.sh

# Create cache directories
RUN mkdir -p /var/cache/nginx /var/run/nginx && \
    chown -R nodejs:nodejs /var/cache/nginx /var/run/nginx /var/log/nginx

# Security headers and optimizations
RUN echo "server_tokens off;" > /etc/nginx/conf.d/server_tokens.conf && \
    echo "client_max_body_size 10M;" > /etc/nginx/conf.d/client_max_body_size.conf

# Switch to non-root user
USER nodejs

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD /usr/local/bin/healthcheck.sh || exit 1

# Start nginx
CMD ["nginx", "-g", "daemon off;"]