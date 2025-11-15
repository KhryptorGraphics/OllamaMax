# OllamaMax Production Deployment Guide

## Overview

This guide provides step-by-step instructions for deploying OllamaMax to production environments.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Pre-Deployment Checklist](#pre-deployment-checklist)
3. [Deployment Steps](#deployment-steps)
4. [Post-Deployment Verification](#post-deployment-verification)
5. [Monitoring Setup](#monitoring-setup)
6. [Troubleshooting](#troubleshooting)

---

## Prerequisites

### System Requirements

**Minimum:**
- CPU: 4 cores
- RAM: 8 GB
- Disk: 50 GB SSD
- OS: Ubuntu 20.04+ / Debian 11+ / RHEL 8+

**Recommended:**
- CPU: 8+ cores
- RAM: 16+ GB
- Disk: 100+ GB NVMe SSD
- OS: Ubuntu 22.04 LTS

### Software Requirements

- Node.js 18.x or higher
- npm 9.x or higher
- Docker 24.x or higher (optional)
- Docker Compose 2.x or higher (optional)
- Nginx or similar reverse proxy
- SSL/TLS certificates

---

## Pre-Deployment Checklist

### 1. Security Audit

```bash
# Run security audit
./scripts/security-audit.sh

# Fix any critical issues before proceeding
```

### 2. Environment Configuration

```bash
# Copy and customize environment file
cp .env.example .env.production

# Edit .env.production with production values
nano .env.production
```

**Critical settings to change:**

For a comprehensive list of all environment variables, please refer to `docs/ENVIRONMENT_VARIABLES.md` and `docs/COMPREHENSIVE_ENV_VAR_REFERENCE.md`.

```bash
# Server
NODE_ENV=production
PORT=13000

# Security - MUST CHANGE
JWT_SECRET=<generate-with-openssl-rand-base64-48>
JWT_SECRET_KEY=<same-as-above>

# Database
DB_PATH=./data/ollamamax.db

# SSL/TLS
SSL_ENABLED=true
SSL_CERT_PATH=/etc/letsencrypt/live/yourdomain.com/fullchain.pem
SSL_KEY_PATH=/etc/letsencrypt/live/yourdomain.com/privkey.pem

# CORS - Restrict to your domain
ALLOWED_ORIGINS=https://yourdomain.com,https://app.yourdomain.com

# Ollama Nodes
ENABLE_OLLAMA_DISCOVERY=true
OLLAMA_NODES=http://node1:11434,http://node2:11434,http://node3:11434

# Monitoring
ENABLE_METRICS=true
PROMETHEUS_ENABLED=true

# Email (for alerts)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_FROM=alerts@yourdomain.com
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
```

### 3. SSL/TLS Certificates

**Option A: Let's Encrypt (Recommended)**

```bash
# Install certbot
sudo apt install certbot

# Generate certificate
sudo certbot certonly --standalone -d api.yourdomain.com

# Certificates will be in:
# /etc/letsencrypt/live/api.yourdomain.com/fullchain.pem
# /etc/letsencrypt/live/api.yourdomain.com/privkey.pem

# Set up auto-renewal
sudo certbot renew --dry-run
```

**Option B: Self-Signed (Development/Testing)**

```bash
./scripts/setup-ssl.sh
```

### 4. Database Backup

```bash
# Create backup directory
mkdir -p backups

# Backup existing database (if upgrading)
cp data/ollamamax.db backups/ollamamax-$(date +%Y%m%d-%H%M%S).db
```

### 5. Load Testing

```bash
# Run load tests to verify capacity
node tests/load-test.js light health
node tests/load-test.js medium models
node tests/load-test.js heavy completion

# Review results in load-test-results/
```

---

## Deployment Steps

### Option 1: Direct Deployment

#### Step 1: Install Dependencies

```bash
# Install production dependencies
npm ci --production

# Or with dev dependencies for monitoring
npm ci
```

#### Step 2: Build Assets (if applicable)

```bash
# Build frontend assets
npm run build

# Optimize images
npm run optimize-assets
```

#### Step 3: Set Up Systemd Service

```bash
# Create systemd service file
sudo nano /etc/systemd/system/ollamamax.service
```

```ini
[Unit]
Description=OllamaMax API Server
After=network.target

[Service]
Type=simple
User=ollama
Group=ollama
WorkingDirectory=/opt/ollamamax
Environment=NODE_ENV=production
EnvironmentFile=/opt/ollamamax/.env.production
ExecStart=/usr/bin/node src/server.js
Restart=always
RestartSec=10
StandardOutput=append:/var/log/ollamamax/access.log
StandardError=append:/var/log/ollamamax/error.log

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/ollamamax/data /opt/ollamamax/logs

[Install]
WantedBy=multi-user.target
```

```bash
# Create log directory
sudo mkdir -p /var/log/ollamamax
sudo chown ollama:ollama /var/log/ollamamax

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable ollamamax
sudo systemctl start ollamamax

# Check status
sudo systemctl status ollamamax
```

#### Step 4: Configure Nginx Reverse Proxy

```bash
sudo nano /etc/nginx/sites-available/ollamamax
```

```nginx
# Rate limiting
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;
limit_req_zone $binary_remote_addr zone=auth_limit:10m rate=5r/m;

# Upstream
upstream ollamamax_backend {
    least_conn;
    server localhost:13000 max_fails=3 fail_timeout=30s;
    keepalive 32;
}

# HTTP to HTTPS redirect
server {
    listen 80;
    server_name api.yourdomain.com;
    return 301 https://$server_name$request_uri;
}

# HTTPS server
server {
    listen 443 ssl http2;
    server_name api.yourdomain.com;

    # SSL Configuration
    ssl_certificate /etc/letsencrypt/live/api.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.yourdomain.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    # Security Headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    # Logging
    access_log /var/log/nginx/ollamamax-access.log;
    error_log /var/log/nginx/ollamamax-error.log;

    # Rate limiting
    limit_req zone=api_limit burst=20 nodelay;

    # API endpoints
    location / {
        proxy_pass http://ollamamax_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
        
        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # Auth endpoints - stricter rate limiting
    location /auth/ {
        limit_req zone=auth_limit burst=5 nodelay;
        proxy_pass http://ollamamax_backend;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # WebSocket support
    location /chat {
        proxy_pass http://ollamamax_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # WebSocket timeouts
        proxy_connect_timeout 7d;
        proxy_send_timeout 7d;
        proxy_read_timeout 7d;
    }

    # Health check - no rate limiting
    location /health {
        proxy_pass http://ollamamax_backend;
        access_log off;
    }
}
```

```bash
# Enable site
sudo ln -s /etc/nginx/sites-available/ollamamax /etc/nginx/sites-enabled/

# Test configuration
sudo nginx -t

# Reload Nginx
sudo systemctl reload nginx
```

---

### Option 2: Developer All-in-One Docker Deployment

For a comprehensive local development environment that mirrors the production setup, you can use the main `docker-compose.yml` file. This setup includes the full monitoring stack.

```bash
# Start all services, including monitoring
docker-compose up -d
```

### Option 3: Docker Deployment

#### Step 1: Build Docker Image

```bash
# Build production image
docker build -t ollamamax:latest -f Dockerfile.production .

# Or use docker-compose
docker-compose -f docker-compose.production.yml build
```

#### Step 2: Deploy with Docker Compose

```bash
# Start services
docker-compose -f docker-compose.production.yml up -d

# Check logs
docker-compose -f docker-compose.production.yml logs -f

# Check status
docker-compose -f docker-compose.production.yml ps
```

### Option 3: Kubernetes Deployment

For production-grade deployments, Kubernetes is the recommended option. The repository includes comprehensive Kubernetes manifests for a full production setup.

#### Step 1: Apply Manifests

The main production manifest is located at `ollama-distributed/deploy/integration/production-deployment.yaml`. This manifest includes the deployment, services, autoscaling, and security policies.

```bash
# Apply the production manifest
kubectl apply -f ollama-distributed/deploy/integration/production-deployment.yaml
```

For a complete monitoring stack, you can also apply the monitoring manifest:

```bash
# Apply the monitoring stack
kubectl apply -f k8s/monitoring-stack.yaml
```

#### Step 2: Verify Deployment

```bash
# Check the status of the deployment
kubectl get pods -n ollama-system

# Check the services
kubectl get services -n ollama-system
```

---

## Frontend Development

To run the frontend locally for development, you'll need to have Node.js and npm installed. The frontend code is located in the `web-interface` directory.

### Step 1: Install Dependencies

```bash
cd web-interface
npm install
```

### Step 2: Configure Backend Connection

The frontend needs to know the URL of the backend API. You can configure this by setting the `API_BASE_URL` environment variable. For example, if you are running the backend locally using the "Developer All-in-One Docker Deployment", the API will be available at `http://localhost:13100`.

Create a `.env` file in the `web-interface` directory:

```
API_BASE_URL=http://localhost:13100
```

### Step 3: Run Development Server

```bash
npm run dev
```

This will start a local development server for the frontend, which will typically be available at `http://localhost:8080`.

---

## Post-Deployment Verification

### 1. Health Checks

```bash
# Basic health
curl https://api.yourdomain.com/health

# Readiness check
curl https://api.yourdomain.com/health/ready

# Liveness check
curl https://api.yourdomain.com/health/live
```

### 2. API Functionality

```bash
# List models
curl https://api.yourdomain.com/v1/models

# List nodes
curl https://api.yourdomain.com/api/nodes

# Test authentication
curl -X POST https://api.yourdomain.com/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password"}'
```

### 3. SSL/TLS Verification

```bash
# Check SSL certificate
openssl s_client -connect api.yourdomain.com:443 -servername api.yourdomain.com

# Test SSL Labs
# Visit: https://www.ssllabs.com/ssltest/analyze.html?d=api.yourdomain.com
```

### 4. Performance Testing

```bash
# Run light load test
node tests/load-test.js light health

# Expected: >1000 req/s, <100ms p99 latency
```

---

## Monitoring Setup

### 1. Start Monitoring Stack

```bash
# Start Prometheus, Grafana, and Alertmanager
docker-compose -f docker-compose.monitoring.yml up -d

# Access Grafana
# URL: http://your-server:3001
# Default credentials: admin/admin (change immediately)
```

### 2. Configure Alerts

```bash
# Edit alertmanager config
nano monitoring/alertmanager/config.yml

# Add your email/Slack webhook
# Restart alertmanager
docker-compose -f docker-compose.monitoring.yml restart alertmanager
```

### 3. Import Dashboards

1. Open Grafana (http://your-server:3001)
2. Go to Dashboards → Import
3. Import dashboards from `monitoring/grafana/dashboards/`

---

## Monitoring

The project includes a comprehensive monitoring stack. When using the `docker-compose.yml` or the Kubernetes manifests, the following services will be available:

*   **Prometheus:** [http://localhost:9090](http://localhost:9090)
*   **Grafana:** [http://localhost:3001](http://localhost:3001)
*   **Jaeger:** [http://localhost:16686](http://localhost:16686)
*   **Kibana:** [http://localhost:5601](http://localhost:5601)

### Health Endpoints

The application exposes the following health endpoints:

*   `/health`: Basic health check.
*   `/api/health`: API health check.

You can use these endpoints to monitor the status of the application.

---

## Troubleshooting

### Service Won't Start

```bash
# Check logs
sudo journalctl -u ollamamax -n 100 --no-pager

# Check for port conflicts
sudo lsof -i :13000

# Check file permissions
ls -la /opt/ollamamax/data
```

### High Memory Usage

```bash
# Check Node.js memory
ps aux | grep node

# Restart service
sudo systemctl restart ollamamax

# Consider increasing memory limit
# Edit systemd service: Environment=NODE_OPTIONS=--max-old-space-size=4096
```

### Database Locked

```bash
# Check for multiple processes
ps aux | grep ollamamax

# Kill stale processes
sudo systemctl stop ollamamax
sudo killall node
sudo systemctl start ollamamax
```

### SSL Certificate Issues

```bash
# Renew Let's Encrypt certificate
sudo certbot renew

# Check certificate expiration
openssl x509 -in /etc/letsencrypt/live/api.yourdomain.com/fullchain.pem -noout -dates
```

---

## Rollback Procedure

If deployment fails:

```bash
# Stop new version
sudo systemctl stop ollamamax

# Restore database backup
cp backups/ollamamax-YYYYMMDD-HHMMSS.db data/ollamamax.db

# Checkout previous version
git checkout <previous-tag>

# Reinstall dependencies
npm ci --production

# Start service
sudo systemctl start ollamamax

# Verify
curl https://api.yourdomain.com/health
```

---

## Maintenance

### Daily
- Monitor alerts and logs
- Check system resources

### Weekly
- Review security logs
- Check for dependency updates
- Review performance metrics

### Monthly
- Update dependencies: `npm update`
- Rotate logs
- Review and optimize database
- Test backup restoration

### Quarterly
- Security audit: `./scripts/security-audit.sh`
- Load testing
- Review and update documentation
- Disaster recovery drill

---

## Support

For deployment issues:
- Check logs: `/var/log/ollamamax/`
- Review documentation: `docs/`
- Contact support: support@ollamamax.com

---

**Last Updated:** November 1, 2025  
**Version:** 1.0.0

