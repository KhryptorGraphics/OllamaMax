# Comprehensive Deployment Procedures

**Version:** 1.0
**Last Updated:** 2025
**References:** DOCKER_DEPLOYMENT.md, FINAL_DEPLOYMENT_GUIDE.md

## Table of Contents

1. [Pre-Deployment Checklist](#pre-deployment-checklist)
2. [Docker Deployment](#docker-deployment)
3. [Kubernetes Deployment](#kubernetes-deployment)
4. [Multi-Region Deployment](#multi-region-deployment)
5. [Monitoring Setup](#monitoring-setup)
6. [Security Configuration](#security-configuration)
7. [Post-Deployment Validation](#post-deployment-validation)
8. [Operational Procedures](#operational-procedures)

---

## Pre-Deployment Checklist

### System Requirements

- **Hardware:**
  - CPU: 8+ cores
  - RAM: 16GB+ (32GB recommended for GPU workloads)
  - Disk: 100GB+ free space
  - GPU: NVIDIA GPU with CUDA support (optional)

- **Software:**
  - Docker: 24.0+
  - Docker Compose: 2.20+
  - kubectl: 1.28+ (for Kubernetes deployments)
  - Helm: 3.12+ (optional)

### Dependency Installation

```bash
# Install Docker
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# Install Docker Compose
sudo apt-get update
sudo apt-get install docker-compose-plugin

# Install kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Verify installations
docker --version
docker compose version
kubectl version --client
```

### Configuration Validation

```bash
# Validate Docker Compose files
docker compose -f docker-compose.yml config
docker compose -f docker-compose.prod.yml config

# Validate Kubernetes manifests
kubectl apply --dry-run=client -f k8s/

# Run pre-deployment validation
bash scripts/validate-docker-deployment.sh
```

### Backup Procedures

```bash
# Backup existing data
docker compose exec postgres pg_dump -U postgres > backup-$(date +%Y%m%d).sql
docker compose exec redis redis-cli BGSAVE

# Backup configuration files
tar -czf config-backup-$(date +%Y%m%d).tar.gz \
  docker-compose*.yml \
  k8s/ \
  nginx/ \
  .env
```

### Rollback Plan Preparation

1. Document current versions
2. Create rollback scripts
3. Test rollback procedures
4. Prepare communication plan

---

## Docker Deployment

### Development Deployment

```bash
# 1. Clone repository
git clone https://github.com/yourusername/OllamaMax.git
cd OllamaMax

# 2. Configure environment
cp .env.example .env
nano .env  # Edit configuration

# 3. Start services
docker compose up -d

# 4. Verify deployment
docker compose ps
docker compose logs -f

# 5. Run health checks
bash ollama-distributed/scripts/health-check.sh
```

### Production Deployment

```bash
# 1. Pull latest production images
docker compose -f docker-compose.prod.yml pull

# 2. Start services with production configuration
docker compose -f docker-compose.prod.yml up -d

# 3. Wait for services to be ready
sleep 30

# 4. Run validation
bash scripts/validate-docker-deployment.sh

# 5. Test load balancing (ensure proxied endpoint is used)
bash scripts/test-load-balancing.sh http://localhost /api/health

# 6. Monitor logs
docker compose -f docker-compose.prod.yml logs -f
```

### GPU Deployment

```bash
# 1. Verify GPU is available
nvidia-smi

# 2. Deploy with GPU support
docker compose -f docker-compose.gpu.yml up -d

# 3. Verify GPU utilization
nvidia-smi -l 1
```

### Service Startup Order

The Docker Compose configuration handles dependency ordering automatically:

1. **Infrastructure Services** (startup first)
   - PostgreSQL
   - Redis
   - InfluxDB

2. **Core Services** (after infrastructure)
   - API server
   - Web frontend

3. **Ollama Nodes** (after core services)
   - Ollama-node-1
   - Ollama-node-2
   - Ollama-node-3

4. **Monitoring** (after all services)
   - Prometheus
   - Grafana
   - Nginx

### Troubleshooting Common Issues

**Issue: Port already in use**
```bash
# Find process using port
sudo lsof -i :8080
# Kill process or change port in configuration
```

**Issue: Services won't start**
```bash
# Check logs
docker compose logs <service-name>
# Restart specific service
docker compose restart <service-name>
```

**Issue: Out of memory**
```bash
# Check Docker resources
docker system df
# Clean up unused resources
docker system prune -a
```

---

## Kubernetes Deployment

### Cluster Preparation

```bash
# 1. Verify cluster access
kubectl cluster-info
kubectl get nodes

# 2. Create namespaces
kubectl create namespace ollamamax-redis
kubectl create namespace ollamamax-timeseries
kubectl create namespace ollamamax-ml
kubectl create namespace ollamamax-monitoring

# 3. Set up RBAC
kubectl apply -f k8s/rbac.yaml
```

### Manifest Application Order

```bash
# Sprint 1: Infrastructure
bash scripts/deploy-sprint1.sh

# This deploys:
# 1. Redis Cluster (StatefulSet)
# 2. Time Series Database (InfluxDB)
# 3. Monitoring Stack (Prometheus, Grafana)

# Sprint 2: ML Pipeline
bash scripts/deploy-sprint2.sh

# This deploys:
# 4. ML Pipeline Components
# 5. Feature Store
# 6. Model Registry
```

### Manual Step-by-Step Deployment

```bash
# 1. Apply Redis Cluster
kubectl apply -f k8s/redis-cluster.yaml

# Wait for StatefulSet
kubectl rollout status statefulset/redis-cluster -n ollamamax-redis

# Initialize Redis Cluster
kubectl exec -it redis-cluster-0 -n ollamamax-redis -- \
  redis-cli --cluster create \
  redis-cluster-0:6379 redis-cluster-1:6379 redis-cluster-2:6379 \
  --cluster-replicas 1 --cluster-yes

# 2. Apply Time Series DB
kubectl apply -f k8s/timeseries-db.yaml

# Wait for deployment
kubectl rollout status deployment/influxdb -n ollamamax-timeseries

# Initialize InfluxDB
kubectl exec -it $(kubectl get pod -n ollamamax-timeseries -l app=influxdb -o jsonpath='{.items[0].metadata.name}') \
  -n ollamamax-timeseries -- \
  influx setup --force \
  --username admin \
  --password changeme \
  --org ollamamax \
  --bucket metrics

# 3. Apply ML Pipeline
kubectl apply -f k8s/ml-pipeline.yaml

# Wait for all deployments
kubectl get deployments -n ollamamax-ml
kubectl rollout status deployment/agent-selection-model -n ollamamax-ml
kubectl rollout status deployment/predictive-scaling-system -n ollamamax-ml

# 4. Apply Monitoring Stack
kubectl apply -f k8s/monitoring-stack.yaml

# Wait for monitoring services
kubectl rollout status deployment/prometheus -n ollamamax-monitoring
kubectl rollout status deployment/grafana -n ollamamax-monitoring
```

### Service Exposure

```bash
# Expose services externally
kubectl apply -f k8s/ingress.yaml

# Get external IPs
kubectl get services --all-namespaces -o wide

# Port forwarding for local access
kubectl port-forward -n ollamamax-monitoring svc/grafana 3001:3000
kubectl port-forward -n ollamamax-monitoring svc/prometheus 9090:9090
```

---

## Multi-Region Deployment

### Region Selection Criteria

- **Latency Requirements:** <100ms for US, <150ms for EU
- **Data Residency:** Compliance with local regulations
- **Cost Optimization:** Balance between performance and cost
- **Disaster Recovery:** Geographic redundancy

### Cross-Region Replication Setup

```bash
# 1. Apply multi-region configuration
kubectl apply -f k8s/multi-region-deployment.yaml

# 2. Configure Redis cross-region replication
# Primary region (us-east-1)
kubectl exec -n ollamamax-multiregion redis-cluster-us-east-0 -- \
  redis-cli CONFIG SET replica-priority 100

# Secondary regions (us-west-2, eu-west-1)
kubectl exec -n ollamamax-multiregion redis-cluster-us-west-0 -- \
  redis-cli REPLICAOF redis-cluster-us-east-0.redis-cluster-us-east 6379

# 3. Test replication
kubectl exec -n ollamamax-multiregion redis-cluster-us-east-0 -- \
  redis-cli SET test-key "cross-region-test"

kubectl exec -n ollamamax-multiregion redis-cluster-us-west-0 -- \
  redis-cli GET test-key
```

### DNS Configuration

```bash
# Set up GeoDNS for regional routing
# Example with AWS Route 53:

# Create health checks
aws route53 create-health-check \
  --caller-reference $(date +%s) \
  --health-check-config \
    IPAddress=<us-east-ip>,Port=80,Type=HTTP,ResourcePath=/health

# Create geolocation routing
aws route53 change-resource-record-sets \
  --hosted-zone-id <zone-id> \
  --change-batch file://geo-routing.json
```

### Failover Testing

```bash
# Run multi-region simulation
bash scripts/simulate-multi-region.sh

# Monitor failover
watch -n 1 'kubectl get pods --all-namespaces | grep -E "us-|eu-"'
```

---

## Monitoring Setup

### Prometheus Configuration

```bash
# 1. Verify Prometheus is running
kubectl get pods -n ollamamax-monitoring -l app=prometheus

# 2. Access Prometheus UI
kubectl port-forward -n ollamamax-monitoring svc/prometheus 9090:9090

# 3. Verify targets
# Open http://localhost:9090/targets
# All targets should show "UP" status

# 4. Test queries
curl http://localhost:9090/api/v1/query?query=up
```

### Grafana Dashboard Import

```bash
# 1. Access Grafana
kubectl port-forward -n ollamamax-monitoring svc/grafana 3001:3000

# 2. Login (default: admin/admin)
# Open http://localhost:3001

# 3. Add Prometheus datasource
# Configuration > Data Sources > Add data source > Prometheus
# URL: http://prometheus.ollamamax-monitoring:9090

# 4. Import deployment dashboard
# Create > Import > Upload JSON file
# Select: monitoring/deployment-dashboard.json
```

### Alert Rule Configuration

```bash
# Edit alert rules
kubectl edit configmap prometheus-alerts -n ollamamax-monitoring

# Reload Prometheus configuration
kubectl exec -n ollamamax-monitoring \
  $(kubectl get pod -n ollamamax-monitoring -l app=prometheus -o jsonpath='{.items[0].metadata.name}') \
  -- kill -HUP 1
```

---

## Security Configuration

### TLS Certificate Installation

```bash
# 1. Generate certificates (Let's Encrypt)
certbot certonly --standalone -d ollamamax.io -d www.ollamamax.io

# 2. Copy certificates to nginx
mkdir -p nginx/ssl
cp /etc/letsencrypt/live/ollamamax.io/fullchain.pem nginx/ssl/
cp /etc/letsencrypt/live/ollamamax.io/privkey.pem nginx/ssl/
cp /etc/letsencrypt/live/ollamamax.io/chain.pem nginx/ssl/

# 3. Update nginx configuration
# Use nginx/nginx-production.conf

# 4. Reload nginx
docker compose restart nginx
```

### Secrets Management

```bash
# Docker Compose: Use .env file
cp .env.example .env
chmod 600 .env

# Kubernetes: Create secrets
kubectl create secret generic db-credentials \
  --from-literal=username=postgres \
  --from-literal=password=$(openssl rand -base64 32) \
  -n ollamamax

kubectl create secret generic api-keys \
  --from-literal=jwt-secret=$(openssl rand -base64 64) \
  -n ollamamax
```

### Authentication Setup

```bash
# Configure JWT authentication
# Edit .env file
JWT_SECRET=$(openssl rand -base64 64)
JWT_EXPIRATION=3600

# Configure OAuth2 (if using)
OAUTH_CLIENT_ID=your-client-id
OAUTH_CLIENT_SECRET=your-client-secret
OAUTH_CALLBACK_URL=https://ollamamax.io/auth/callback
```

---

## Post-Deployment Validation

### Health Check Execution

```bash
# Run comprehensive health checks
bash ollama-distributed/scripts/health-check.sh

# Expected output: All services should return "OK"
```

### Smoke Test Execution

```bash
# Run deployment smoke tests
go test ./ollama-distributed/tests/smoke/... -v

# Run deployment integration tests
go test ./tests/deployment/... -v
```

### Performance Validation

```bash
# Run load tests
k6 run load-test.js

# Expected results:
# - Response time P95 < 500ms
# - Error rate < 1%
# - Throughput > 1000 RPS
```

### Load Balancing Validation

```bash
# Run load balancing tests
# Syntax: bash scripts/test-load-balancing.sh <LB_URL> <LB_PATH>
# Default: http://localhost /api/health

# Test with default proxied endpoint (/api/health)
bash scripts/test-load-balancing.sh http://localhost

# Test with custom proxied endpoint
bash scripts/test-load-balancing.sh http://localhost /ollama/api/tags

# Test with root path (proxied to web_backend)
bash scripts/test-load-balancing.sh http://localhost /

# Expected results:
# - Backend distribution within 20% variance
# - X-Served-By header present (e.g., api:8080, ollama-node-1:11434)
# - All backends receiving traffic
# - No failed requests under load
```

**Important Notes:**
- The test script uses proxied endpoints that include the `X-Served-By` header set by Nginx
- The self-served `/health` and `/nginx-health` endpoints are non-proxied and do NOT include `X-Served-By`
- For accurate backend distribution analysis, always use proxied paths like `/api/health`, `/ollama/api/tags`, or `/`
- The script automatically trims CRLF from headers to ensure accurate key matching

### Security Validation

```bash
# Run security validation
bash scripts/validate-security.sh

# Expected: Security score > 80/100
```

---

## Operational Procedures

### Scaling Procedures

**Manual Scaling:**
```bash
# Docker Compose
docker compose up -d --scale ollama-node=5

# Kubernetes
kubectl scale deployment/ollamamax-api --replicas=5 -n ollamamax
```

**Auto-Scaling:**
```bash
# Apply HPA configuration
kubectl apply -f k8s/hpa-autoscaling.yaml

# Monitor auto-scaling
kubectl get hpa --all-namespaces -w
```

### Update Procedures

```bash
# 1. Pull latest images
docker compose pull

# 2. Perform rolling update
docker compose up -d --no-deps --build <service>

# 3. Verify update
docker compose ps
docker compose logs <service>

# 4. Run health checks
bash ollama-distributed/scripts/health-check.sh
```

### Backup Procedures

```bash
# Automated backup script
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)

# Database backup
docker compose exec -T postgres pg_dump -U postgres > backup-db-${DATE}.sql

# Redis backup
docker compose exec redis redis-cli BGSAVE
docker cp $(docker compose ps -q redis):/data/dump.rdb backup-redis-${DATE}.rdb

# Configuration backup
tar -czf backup-config-${DATE}.tar.gz docker-compose*.yml k8s/ nginx/ .env
```

### Monitoring Procedures

```bash
# View real-time metrics
docker stats

# View service logs
docker compose logs -f --tail=100

# Kubernetes logs
kubectl logs -f -n ollamamax <pod-name>

# Aggregate logs
kubectl logs -f -n ollamamax -l app=ollamamax-api
```

### Incident Response

1. **Detect:** Monitor alerts in Grafana/Prometheus
2. **Assess:** Check service status and logs
3. **Respond:** Execute rollback if necessary
4. **Resolve:** Fix root cause
5. **Document:** Update runbooks

---

## Support and Resources

- **Documentation:** `/docs` directory
- **Runbooks:** `/docs/runbooks` directory
- **Issue Tracker:** GitHub Issues
- **Monitoring:** Grafana dashboards
- **Logs:** Centralized logging system

---

**Note:** Always test deployment procedures in a staging environment before applying to production.
