# Web Interface Dockerfile for OllamaMax
# Lightweight container serving the web interface

FROM nginx:alpine

# Install curl for health checks
RUN apk --no-cache add curl

# Create nginx user and directories
RUN addgroup -g 1001 webapp && \
    adduser -D -s /bin/sh -u 1001 -G webapp webapp

# Copy web interface files
COPY web-interface/ /app/web/
COPY nginx/webapp.conf /etc/nginx/conf.d/default.conf

# Create log directory and set permissions
RUN mkdir -p /var/log/nginx /var/cache/nginx && \
    chown -R webapp:webapp /app/web /var/log/nginx /var/cache/nginx /etc/nginx/conf.d

# Switch to non-root user
USER webapp

# Expose web port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/ || exit 1

# Start nginx
CMD ["nginx", "-g", "daemon off;"]