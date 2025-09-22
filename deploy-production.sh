#!/bin/bash

# OllamaMax Production Deployment Script
# Deploys essential services for production

set -e

echo "================================================"
echo "   OllamaMax Production Deployment"
echo "================================================"

# Check Docker Swarm
if ! docker info --format '{{.Swarm.LocalNodeState}}' | grep -q "active"; then
    echo "Error: Docker Swarm is not active"
    exit 1
fi

# Create secrets if they don't exist
echo "Setting up secrets..."
docker secret ls | grep -q postgres_password || echo "postgres123" | docker secret create postgres_password -
docker secret ls | grep -q jwt_secret || echo "jwt-secret-production" | docker secret create jwt_secret -
docker secret ls | grep -q encryption_key || echo "encryption-key-prod" | docker secret create encryption_key -
docker secret ls | grep -q grafana_password || echo "admin123" | docker secret create grafana_password -

# Create necessary directories
echo "Creating directories..."
mkdir -p nginx data logs models memory monitoring/grafana/provisioning/datasources

# Create minimal nginx config
cat > nginx/nginx.conf << 'EOF'
events {
    worker_connections 1024;
}

http {
    upstream api {
        server ollamamax-api:3000;
    }

    server {
        listen 80;
        location / {
            proxy_pass http://api;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
        }
        
        location /health {
            return 200 "OK\n";
            add_header Content-Type text/plain;
        }
    }
}
EOF

# Create minimal compose file for production
cat > docker-compose.prod.yml << 'EOF'
version: '3.8'

services:
  redis:
    image: redis:7-alpine
    deploy:
      replicas: 1
    networks:
      - prod-net
    volumes:
      - redis-data:/data
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
      - prod-net
    volumes:
      - postgres-data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ollamamax"]
      interval: 10s

  ollamamax-api:
    image: ollamamax:latest
    deploy:
      replicas: 3
      update_config:
        parallelism: 1
        delay: 10s
      restart_policy:
        condition: on-failure
    environment:
      NODE_ENV: production
      PORT: 3000
      REDIS_URL: redis://redis:6379
      DATABASE_URL: postgresql://ollamamax:postgres123@postgres:5432/ollamamax
    secrets:
      - jwt_secret
      - encryption_key
    networks:
      - prod-net
    volumes:
      - ./data:/app/data
      - ./logs:/app/logs
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:3000/health"]
      interval: 30s
    depends_on:
      - redis
      - postgres

  swarm-worker:
    image: ollamamax-swarm:latest
    deploy:
      replicas: 2
    environment:
      NODE_ENV: production
      WORKER_TYPE: swarm
      REDIS_URL: redis://redis:6379
    networks:
      - prod-net
    volumes:
      - ./memory:/app/memory
      - ./logs:/app/logs

  nginx:
    image: nginx:alpine
    deploy:
      replicas: 1
    ports:
      - "80:80"
    networks:
      - prod-net
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost/health"]
      interval: 10s
    depends_on:
      - ollamamax-api

  prometheus:
    image: prom/prometheus:latest
    deploy:
      replicas: 1
    ports:
      - "9090:9090"
    networks:
      - prod-net
    volumes:
      - prometheus-data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'

  grafana:
    image: grafana/grafana:latest
    deploy:
      replicas: 1
    ports:
      - "3001:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD__FILE=/run/secrets/grafana_password
    secrets:
      - grafana_password
    networks:
      - prod-net
    volumes:
      - grafana-data:/var/lib/grafana

networks:
  prod-net:
    driver: overlay
    attachable: true

volumes:
  redis-data:
  postgres-data:
  prometheus-data:
  grafana-data:

secrets:
  postgres_password:
    external: true
  jwt_secret:
    external: true
  encryption_key:
    external: true
  grafana_password:
    external: true
EOF

# Deploy the stack
echo "Deploying production stack..."
docker stack deploy -c docker-compose.prod.yml ollamamax

# Wait for services
echo "Waiting for services to start (30 seconds)..."
sleep 30

# Check status
echo ""
echo "Service Status:"
docker service ls | grep ollamamax

echo ""
echo "================================================"
echo "   Production Deployment Complete!"
echo "================================================"
echo ""
echo "Access Points:"
echo "  - Application: http://localhost"
echo "  - Prometheus: http://localhost:9090"
echo "  - Grafana: http://localhost:3001 (admin/admin123)"
echo ""
echo "Services:"
docker service ls | grep ollamamax | awk '{print "  - "$2": "$4" replicas"}'
echo ""
echo "To check logs: docker service logs <service_name>"
echo "To scale: docker service scale ollamamax_ollamamax-api=5"
echo "To remove: docker stack rm ollamamax"
echo ""