#!/bin/bash

# OllamaMax Docker Swarm Deployment Script
# This script deploys the complete OllamaMax stack to Docker Swarm

set -e

echo "================================================"
echo "   OllamaMax Docker Swarm Deployment"
echo "================================================"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[!]${NC} $1"
}

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed. Please install Docker first."
    exit 1
fi

# Check if running in Swarm mode
if ! docker info --format '{{.Swarm.LocalNodeState}}' | grep -q "active"; then
    print_warning "Docker Swarm is not initialized. Initializing now..."
    docker swarm init --advertise-addr $(hostname -I | awk '{print $1}')
    print_status "Docker Swarm initialized"
else
    print_status "Docker Swarm is active"
fi

# Create required directories
print_status "Creating required directories..."
mkdir -p data logs models memory \
    nginx/ssl monitoring/grafana/provisioning \
    monitoring/grafana/provisioning/dashboards \
    monitoring/grafana/provisioning/datasources

# Generate SSL certificates if they don't exist
if [ ! -f nginx/ssl/cert.pem ]; then
    print_status "Generating self-signed SSL certificates..."
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout nginx/ssl/key.pem \
        -out nginx/ssl/cert.pem \
        -subj "/C=US/ST=State/L=City/O=OllamaMax/CN=localhost" 2>/dev/null
fi

# Create Docker secrets
print_status "Creating Docker secrets..."

# Generate random passwords if not provided
POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-$(openssl rand -base64 32)}
JWT_SECRET=${JWT_SECRET:-$(openssl rand -base64 64)}
ENCRYPTION_KEY=${ENCRYPTION_KEY:-$(openssl rand -base64 32)}
GRAFANA_PASSWORD=${GRAFANA_PASSWORD:-$(openssl rand -base64 32)}

# Remove existing secrets if they exist
docker secret rm postgres_password jwt_secret encryption_key grafana_password 2>/dev/null || true

# Create new secrets
echo "$POSTGRES_PASSWORD" | docker secret create postgres_password -
echo "$JWT_SECRET" | docker secret create jwt_secret -
echo "$ENCRYPTION_KEY" | docker secret create encryption_key -
echo "$GRAFANA_PASSWORD" | docker secret create grafana_password -

print_status "Secrets created successfully"

# Create nginx configuration
print_status "Creating nginx configuration..."
cat > nginx/nginx.conf << 'EOF'
events {
    worker_connections 1024;
}

http {
    upstream ollamamax_api {
        least_conn;
        server ollamamax-api:3000 max_fails=3 fail_timeout=30s;
    }

    server {
        listen 80;
        server_name _;
        return 301 https://$host$request_uri;
    }

    server {
        listen 443 ssl http2;
        server_name _;

        ssl_certificate /etc/nginx/ssl/cert.pem;
        ssl_certificate_key /etc/nginx/ssl/key.pem;
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers HIGH:!aNULL:!MD5;

        location / {
            proxy_pass http://ollamamax_api;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            
            # WebSocket support
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection "upgrade";
        }

        location /health {
            access_log off;
            return 200 "healthy\n";
            add_header Content-Type text/plain;
        }
    }
}
EOF

# Create Prometheus configuration
print_status "Creating Prometheus configuration..."
cat > monitoring/prometheus.yml << 'EOF'
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'ollamamax-api'
    dns_sd_configs:
      - names:
          - 'tasks.ollamamax-api'
        type: 'A'
        port: 3000

  - job_name: 'swarm-workers'
    dns_sd_configs:
      - names:
          - 'tasks.swarm-worker'
        type: 'A'
        port: 8080

  - job_name: 'ml-workers'
    dns_sd_configs:
      - names:
          - 'tasks.ml-worker'
        type: 'A'
        port: 8081

  - job_name: 'redis'
    static_configs:
      - targets: ['redis-master:6379']

  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres:5432']
EOF

# Create Grafana datasource configuration
print_status "Creating Grafana configuration..."
cat > monitoring/grafana/provisioning/datasources/prometheus.yml << 'EOF'
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: true
EOF

# Build Docker images
print_status "Building Docker images..."
docker build -t ollamamax:latest -f Dockerfile .
docker build -t ollamamax-swarm:latest -f Dockerfile.swarm .
docker build -t ollamamax-ml:latest -f Dockerfile.ml .

# Deploy the stack
print_status "Deploying OllamaMax stack to Docker Swarm..."
docker stack deploy -c docker-compose.swarm.yml ollamamax

# Wait for services to be ready
print_status "Waiting for services to start..."
sleep 10

# Check service status
print_status "Checking service status..."
docker service ls | grep ollamamax

# Get service logs
print_status "Recent logs from services:"
docker service logs ollamamax_ollamamax-api --tail 20

# Display access information
echo ""
echo "================================================"
echo "   Deployment Complete!"
echo "================================================"
echo ""
echo "Access URLs:"
echo "  - Application: https://localhost"
echo "  - Prometheus: http://localhost:9090"
echo "  - Grafana: http://localhost:3001 (admin/check secret)"
echo "  - Kibana: http://localhost:5601"
echo "  - Portainer: http://localhost:9000"
echo ""
echo "Passwords saved in Docker secrets:"
echo "  - Postgres: postgres_password"
echo "  - JWT: jwt_secret"
echo "  - Encryption: encryption_key"
echo "  - Grafana: grafana_password"
echo ""
echo "To view service status:"
echo "  docker service ls"
echo ""
echo "To view service logs:"
echo "  docker service logs <service_name>"
echo ""
echo "To scale a service:"
echo "  docker service scale ollamamax_ollamamax-api=5"
echo ""
echo "To remove the stack:"
echo "  docker stack rm ollamamax"
echo ""
print_status "Deployment script completed successfully!"