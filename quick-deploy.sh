#!/bin/bash

# Quick Docker Swarm Deployment for Testing
# Uses pre-built images for faster deployment

set -e

echo "================================================"
echo "   OllamaMax Quick Deployment (Testing)"
echo "================================================"

# Check if Docker Swarm is active
if ! docker info --format '{{.Swarm.LocalNodeState}}' | grep -q "active"; then
    echo "Initializing Docker Swarm..."
    docker swarm init
fi

# Create minimal secrets
echo "Creating secrets..."
echo "password123" | docker secret create postgres_password - 2>/dev/null || true
echo "jwt-secret-123" | docker secret create jwt_secret - 2>/dev/null || true
echo "encryption-key-123" | docker secret create encryption_key - 2>/dev/null || true
echo "admin" | docker secret create grafana_password - 2>/dev/null || true

# Create directories
mkdir -p nginx monitoring/grafana data logs

# Create simple nginx config
cat > nginx/default.conf << 'EOF'
server {
    listen 80;
    location /health {
        return 200 "healthy\n";
        add_header Content-Type text/plain;
    }
}
EOF

# Create minimal docker-compose for testing
cat > docker-compose.test.yml << 'EOF'
version: '3.8'

services:
  redis:
    image: redis:7-alpine
    deploy:
      replicas: 1
    networks:
      - test-net
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s

  postgres:
    image: postgres:15-alpine
    deploy:
      replicas: 1
    environment:
      POSTGRES_DB: ollamamax
      POSTGRES_USER: ollamamax
      POSTGRES_PASSWORD_FILE: /run/secrets/postgres_password
    secrets:
      - postgres_password
    networks:
      - test-net
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ollamamax"]
      interval: 10s

  nginx:
    image: nginx:alpine
    deploy:
      replicas: 1
    ports:
      - "8080:80"
    networks:
      - test-net
    volumes:
      - ./nginx/default.conf:/etc/nginx/conf.d/default.conf:ro
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost/health"]
      interval: 10s

networks:
  test-net:
    driver: overlay
    attachable: true

secrets:
  postgres_password:
    external: true
EOF

# Deploy the test stack
echo "Deploying test stack..."
docker stack deploy -c docker-compose.test.yml ollamamax-test

# Wait for services
echo "Waiting for services to start..."
sleep 15

# Check service status
echo ""
echo "Service Status:"
docker service ls | grep ollamamax-test

echo ""
echo "================================================"
echo "   Quick Deployment Complete!"
echo "================================================"
echo ""
echo "Test URLs:"
echo "  - Nginx Health: http://localhost:8080/health"
echo ""
echo "To check services: docker service ls"
echo "To remove: docker stack rm ollamamax-test"
echo ""