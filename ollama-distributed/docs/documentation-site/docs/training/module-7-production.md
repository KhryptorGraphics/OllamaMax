# Module 7: Production Deployment and Scaling

> 🚨 **Critical Security Notice**: This module demonstrates production deployment patterns. **All examples require secure secret management, TLS certificates, and proper access controls before production use**. Review [Security Guidelines](../../../../SECURITY-GUIDELINES.md) thoroughly before deployment.

**Duration**: 20 minutes  
**Objective**: Master production deployment strategies, scaling approaches, high availability, monitoring, security hardening, and disaster recovery for enterprise OllamaMax Distributed environments

Welcome to Module 7 - the final and most advanced module! This is where you'll learn to deploy, secure, monitor, and scale OllamaMax Distributed in production environments with enterprise-grade reliability and performance.

## 🎯 What You'll Learn

By the end of this module, you will:
- ✅ Deploy OllamaMax using Docker containers and Kubernetes orchestration
- ✅ Implement load balancing and high availability patterns
- ✅ Execute horizontal and vertical scaling strategies
- ✅ Set up comprehensive monitoring with Prometheus and Grafana
- ✅ Implement security hardening and compliance measures
- ✅ Design backup and disaster recovery procedures
- ✅ Optimize performance at enterprise scale
- ✅ Troubleshoot production issues and implement SRE practices

## 📋 Prerequisites

Before starting this module, ensure you have:
- Completed Modules 1-6
- Docker and Docker Compose installed
- Kubernetes cluster access (minikube/kind for testing)
- Basic understanding of container orchestration
- Production environment considerations

## 🐳 Production Deployment Strategies

### Docker Containerization

#### Production Dockerfile

Create a multi-stage production-ready Docker image:

```dockerfile
# Production Dockerfile for OllamaMax Distributed
FROM golang:1.21-alpine AS builder

# Security: Create non-root user
RUN adduser -D -s /bin/sh ollamauser

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o ollama-distributed ./cmd/ollama-distributed

# Production stage
FROM alpine:3.18

# Security updates
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN adduser -D -s /bin/sh ollamauser

# Create necessary directories
RUN mkdir -p /app/data /app/logs /app/config && \
    chown -R ollamauser:ollamauser /app

# Copy binary and set permissions
COPY --from=builder /app/ollama-distributed /app/
COPY --from=builder /app/config/production.yaml /app/config/
RUN chmod +x /app/ollama-distributed

# Switch to non-root user
USER ollamauser

# Set working directory
WORKDIR /app

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ./ollama-distributed health --config=/app/config/production.yaml || exit 1

# Expose ports
EXPOSE 8080 8443 9090

# Set entrypoint
ENTRYPOINT ["./ollama-distributed"]
CMD ["server", "--config=/app/config/production.yaml"]
```

#### Docker Compose for Production

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  ollama-coordinator:
    build:
      context: .
      dockerfile: Dockerfile.prod
    image: ollamamax/distributed:latest
    container_name: ollama-coordinator
    restart: unless-stopped
    environment:
      - OLLAMA_ENV=production
      - OLLAMA_NODE_TYPE=coordinator
      - OLLAMA_LOG_LEVEL=info
      - OLLAMA_METRICS_ENABLED=true
    ports:
      - "8080:8080"
      - "8443:8443"
      - "9090:9090"
    volumes:
      - ./config/production.yaml:/app/config/production.yaml:ro
      - ollama-data:/app/data
      - ollama-logs:/app/logs
    networks:
      - ollama-network
    healthcheck:
      test: ["CMD", "./ollama-distributed", "health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    deploy:
      resources:
        limits:
          memory: 4G
          cpus: '2.0'
        reservations:
          memory: 2G
          cpus: '1.0'

  ollama-worker-1:
    build:
      context: .
      dockerfile: Dockerfile.prod
    image: ollamamax/distributed:latest
    container_name: ollama-worker-1
    restart: unless-stopped
    environment:
      - OLLAMA_ENV=production
      - OLLAMA_NODE_TYPE=worker
      - OLLAMA_COORDINATOR_URL=http://ollama-coordinator:8080
      - OLLAMA_LOG_LEVEL=info
      - OLLAMA_METRICS_ENABLED=true
    volumes:
      - ./config/production.yaml:/app/config/production.yaml:ro
      - ollama-worker1-data:/app/data
      - ollama-worker1-logs:/app/logs
    networks:
      - ollama-network
    depends_on:
      ollama-coordinator:
        condition: service_healthy
    deploy:
      resources:
        limits:
          memory: 8G
          cpus: '4.0'
        reservations:
          memory: 4G
          cpus: '2.0'

  ollama-worker-2:
    build:
      context: .
      dockerfile: Dockerfile.prod
    image: ollamamax/distributed:latest
    container_name: ollama-worker-2
    restart: unless-stopped
    environment:
      - OLLAMA_ENV=production
      - OLLAMA_NODE_TYPE=worker
      - OLLAMA_COORDINATOR_URL=http://ollama-coordinator:8080
      - OLLAMA_LOG_LEVEL=info
      - OLLAMA_METRICS_ENABLED=true
    volumes:
      - ./config/production.yaml:/app/config/production.yaml:ro
      - ollama-worker2-data:/app/data
      - ollama-worker2-logs:/app/logs
    networks:
      - ollama-network
    depends_on:
      ollama-coordinator:
        condition: service_healthy
    deploy:
      resources:
        limits:
          memory: 8G
          cpus: '4.0'
        reservations:
          memory: 4G
          cpus: '2.0'

  redis:
    image: redis:7.2-alpine
    container_name: ollama-redis
    restart: unless-stopped
    command: redis-server --appendonly yes --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis-data:/data
    networks:
      - ollama-network
    deploy:
      resources:
        limits:
          memory: 1G
          cpus: '0.5'

  nginx:
    image: nginx:alpine
    container_name: ollama-nginx
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./config/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
      - nginx-logs:/var/log/nginx
    networks:
      - ollama-network
    depends_on:
      - ollama-coordinator
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '0.5'

volumes:
  ollama-data:
  ollama-logs:
  ollama-worker1-data:
  ollama-worker1-logs:
  ollama-worker2-data:
  ollama-worker2-logs:
  redis-data:
  nginx-logs:

networks:
  ollama-network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16
```

### Kubernetes Deployment

#### Production Kubernetes Manifests

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: ollamamax
  labels:
    name: ollamamax
    environment: production
---
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: ollama-config
  namespace: ollamamax
data:
  production.yaml: |
    server:
      host: "0.0.0.0"
      port: 8080
      tls:
        enabled: true
        cert_file: "/etc/ssl/certs/tls.crt"
        key_file: "/etc/ssl/certs/tls.key"
      
    cluster:
      coordination:
        type: "kubernetes"
        discovery:
          method: "k8s-service"
          service_name: "ollama-coordinator"
          namespace: "ollamamax"
      
    metrics:
      enabled: true
      port: 9090
      path: "/metrics"
      
    logging:
      level: "info"
      format: "json"
      output: "stdout"
      
    performance:
      max_concurrent_requests: 100
      request_timeout: "30s"
      model_cache_size: "4GB"
      
    security:
      cors:
        enabled: true
        allowed_origins: ["https://*.yourdomain.com"]
      rate_limiting:
        enabled: true
        requests_per_minute: 1000
---
# k8s/secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: ollama-secrets
  namespace: ollamamax
type: Opaque
data:
  redis-password: <base64-encoded-password>
  jwt-secret: <base64-encoded-jwt-secret>
  api-key: <base64-encoded-api-key>
---
# k8s/pvc.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: ollama-data-pvc
  namespace: ollamamax
spec:
  accessModes:
    - ReadWriteMany
  storageClassName: fast-ssd
  resources:
    requests:
      storage: 100Gi
---
# k8s/coordinator-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ollama-coordinator
  namespace: ollamamax
  labels:
    app: ollama-coordinator
    component: coordinator
spec:
  replicas: 2
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: ollama-coordinator
  template:
    metadata:
      labels:
        app: ollama-coordinator
        component: coordinator
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9090"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: ollama-service-account
      securityContext:
        runAsNonRoot: true
        runAsUser: 1001
        fsGroup: 1001
      containers:
      - name: coordinator
        image: ollamamax/distributed:v1.0.0
        imagePullPolicy: Always
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 8443
          name: https
        - containerPort: 9090
          name: metrics
        env:
        - name: OLLAMA_NODE_TYPE
          value: "coordinator"
        - name: OLLAMA_ENV
          value: "production"
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: ollama-secrets
              key: redis-password
        resources:
          requests:
            memory: "2Gi"
            cpu: "1000m"
          limits:
            memory: "4Gi"
            cpu: "2000m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 2
        volumeMounts:
        - name: config
          mountPath: /app/config
          readOnly: true
        - name: data
          mountPath: /app/data
        - name: tls-certs
          mountPath: /etc/ssl/certs
          readOnly: true
      volumes:
      - name: config
        configMap:
          name: ollama-config
      - name: data
        persistentVolumeClaim:
          claimName: ollama-data-pvc
      - name: tls-certs
        secret:
          secretName: ollama-tls-secret
---
# k8s/worker-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ollama-worker
  namespace: ollamamax
  labels:
    app: ollama-worker
    component: worker
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 2
      maxUnavailable: 1
  selector:
    matchLabels:
      app: ollama-worker
  template:
    metadata:
      labels:
        app: ollama-worker
        component: worker
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9090"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: ollama-service-account
      securityContext:
        runAsNonRoot: true
        runAsUser: 1001
        fsGroup: 1001
      containers:
      - name: worker
        image: ollamamax/distributed:v1.0.0
        imagePullPolicy: Always
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 9090
          name: metrics
        env:
        - name: OLLAMA_NODE_TYPE
          value: "worker"
        - name: OLLAMA_ENV
          value: "production"
        - name: OLLAMA_COORDINATOR_URL
          value: "https://ollama-coordinator:8443"
        resources:
          requests:
            memory: "4Gi"
            cpu: "2000m"
            nvidia.com/gpu: 1
          limits:
            memory: "8Gi"
            cpu: "4000m"
            nvidia.com/gpu: 1
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 60
          periodSeconds: 15
          timeoutSeconds: 10
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5
          timeoutSeconds: 5
          failureThreshold: 3
        volumeMounts:
        - name: config
          mountPath: /app/config
          readOnly: true
        - name: models
          mountPath: /app/models
        - name: tmp
          mountPath: /tmp
      volumes:
      - name: config
        configMap:
          name: ollama-config
      - name: models
        persistentVolumeClaim:
          claimName: ollama-models-pvc
      - name: tmp
        emptyDir:
          sizeLimit: 10Gi
---
# k8s/services.yaml
apiVersion: v1
kind: Service
metadata:
  name: ollama-coordinator
  namespace: ollamamax
  labels:
    app: ollama-coordinator
spec:
  type: ClusterIP
  ports:
  - port: 8080
    targetPort: 8080
    protocol: TCP
    name: http
  - port: 8443
    targetPort: 8443
    protocol: TCP
    name: https
  - port: 9090
    targetPort: 9090
    protocol: TCP
    name: metrics
  selector:
    app: ollama-coordinator
---
apiVersion: v1
kind: Service
metadata:
  name: ollama-worker
  namespace: ollamamax
  labels:
    app: ollama-worker
spec:
  type: ClusterIP
  ports:
  - port: 8080
    targetPort: 8080
    protocol: TCP
    name: http
  - port: 9090
    targetPort: 9090
    protocol: TCP
    name: metrics
  selector:
    app: ollama-worker
---
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ollama-ingress
  namespace: ollamamax
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/use-regex: "true"
    nginx.ingress.kubernetes.io/rewrite-target: /$1
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/rate-limit-window: "1m"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  tls:
  - hosts:
    - api.ollamamax.yourdomain.com
    secretName: ollama-tls-secret
  rules:
  - host: api.ollamamax.yourdomain.com
    http:
      paths:
      - path: /(.*)
        pathType: Prefix
        backend:
          service:
            name: ollama-coordinator
            port:
              number: 8080
```

#### Horizontal Pod Autoscaler

```yaml
# k8s/hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: ollama-worker-hpa
  namespace: ollamamax
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: ollama-worker
  minReplicas: 3
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60
    scaleDown:
      stabilizationWindowSeconds: 600
      policies:
      - type: Percent
        value: 25
        periodSeconds: 60
```

## 🔄 Load Balancing and High Availability

### NGINX Load Balancer Configuration

```nginx
# config/nginx.conf
upstream ollama_coordinators {
    least_conn;
    server ollama-coordinator-1:8080 max_fails=3 fail_timeout=30s;
    server ollama-coordinator-2:8080 max_fails=3 fail_timeout=30s;
    keepalive 32;
}

upstream ollama_workers {
    least_conn;
    server ollama-worker-1:8080 max_fails=3 fail_timeout=30s;
    server ollama-worker-2:8080 max_fails=3 fail_timeout=30s;
    server ollama-worker-3:8080 max_fails=3 fail_timeout=30s;
    keepalive 64;
}

# Rate limiting
limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
limit_req_zone $binary_remote_addr zone=models:10m rate=5r/s;

server {
    listen 80;
    server_name api.ollamamax.yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name api.ollamamax.yourdomain.com;
    
    # SSL Configuration
    ssl_certificate /etc/nginx/ssl/tls.crt;
    ssl_certificate_key /etc/nginx/ssl/tls.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    
    # Security Headers
    add_header Strict-Transport-Security "max-age=63072000" always;
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    
    # Logging
    access_log /var/log/nginx/ollama_access.log combined;
    error_log /var/log/nginx/ollama_error.log warn;
    
    # Health check endpoint
    location /health {
        access_log off;
        proxy_pass http://ollama_coordinators;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_connect_timeout 5s;
        proxy_read_timeout 10s;
    }
    
    # API endpoints
    location /api/ {
        limit_req zone=api burst=20 nodelay;
        
        proxy_pass http://ollama_coordinators;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        proxy_connect_timeout 10s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
        
        # Connection keep-alive
        proxy_http_version 1.1;
        proxy_set_header Connection "";
    }
    
    # Model endpoints (higher timeout)
    location /models/ {
        limit_req zone=models burst=10 nodelay;
        
        proxy_pass http://ollama_workers;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        proxy_connect_timeout 30s;
        proxy_send_timeout 300s;
        proxy_read_timeout 300s;
        
        # Large file uploads
        client_max_body_size 10G;
        
        proxy_http_version 1.1;
        proxy_set_header Connection "";
    }
    
    # Metrics endpoint (internal only)
    location /metrics {
        allow 10.0.0.0/8;
        allow 172.16.0.0/12;
        allow 192.168.0.0/16;
        deny all;
        
        proxy_pass http://ollama_coordinators;
        proxy_set_header Host $host;
    }
}
```

### HAProxy Configuration

```haproxy
# config/haproxy.cfg
global
    maxconn 4096
    log stdout local0
    chroot /var/lib/haproxy
    stats socket /run/haproxy/admin.sock mode 660 level admin
    stats timeout 30s
    user haproxy
    group haproxy
    daemon

defaults
    mode http
    timeout connect 5000ms
    timeout client 50000ms
    timeout server 50000ms
    option httplog
    option dontlognull
    option http-server-close
    option forwardfor except 127.0.0.0/8
    option redispatch
    retries 3
    
    # Health check
    option httpchk GET /health
    http-check expect status 200

frontend ollama_frontend
    bind *:80
    bind *:443 ssl crt /etc/ssl/certs/ollama.pem
    redirect scheme https if !{ ssl_fc }
    
    # Rate limiting
    stick-table type ip size 100k expire 30s store http_req_rate(10s)
    http-request track-sc0 src
    http-request reject if { sc_http_req_rate(0) gt 20 }
    
    # Route to backends
    use_backend coordinators if { path_beg /api }
    use_backend workers if { path_beg /models }
    default_backend coordinators

backend coordinators
    balance roundrobin
    option httpchk GET /health
    
    server coordinator1 ollama-coordinator-1:8080 check
    server coordinator2 ollama-coordinator-2:8080 check

backend workers
    balance leastconn
    option httpchk GET /health
    
    server worker1 ollama-worker-1:8080 check
    server worker2 ollama-worker-2:8080 check
    server worker3 ollama-worker-3:8080 check

listen stats
    bind *:8404
    stats enable
    stats uri /stats
    stats refresh 30s
    stats admin if TRUE
```

## 📈 Scaling Strategies

### Horizontal Scaling Script

```bash
#!/bin/bash
# scripts/scale-cluster.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="${SCRIPT_DIR}/../config/scaling.conf"

# Load configuration
source "${CONFIG_FILE}"

# Logging function
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $*" | tee -a "${LOG_FILE}"
}

# Check current load
check_load() {
    local coordinator_url="$1"
    
    # Get current metrics
    local cpu_usage
    local memory_usage
    local request_rate
    local queue_length
    
    cpu_usage=$(curl -s "${coordinator_url}/metrics" | grep "cpu_usage" | awk '{print $2}')
    memory_usage=$(curl -s "${coordinator_url}/metrics" | grep "memory_usage" | awk '{print $2}')
    request_rate=$(curl -s "${coordinator_url}/metrics" | grep "request_rate" | awk '{print $2}')
    queue_length=$(curl -s "${coordinator_url}/metrics" | grep "queue_length" | awk '{print $2}')
    
    echo "${cpu_usage:-0} ${memory_usage:-0} ${request_rate:-0} ${queue_length:-0}"
}

# Scale up workers
scale_up() {
    local current_replicas="$1"
    local target_replicas="$2"
    
    log "Scaling up from ${current_replicas} to ${target_replicas} replicas"
    
    if [[ "${DEPLOYMENT_TYPE}" == "kubernetes" ]]; then
        kubectl scale deployment ollama-worker --replicas="${target_replicas}" -n ollamamax
        kubectl rollout status deployment/ollama-worker -n ollamamax --timeout=300s
    elif [[ "${DEPLOYMENT_TYPE}" == "docker" ]]; then
        docker-compose -f docker-compose.prod.yml up -d --scale ollama-worker="${target_replicas}"
    fi
    
    log "Scale up completed successfully"
}

# Scale down workers
scale_down() {
    local current_replicas="$1"
    local target_replicas="$2"
    
    log "Scaling down from ${current_replicas} to ${target_replicas} replicas"
    
    # Graceful shutdown: drain requests first
    if [[ "${DEPLOYMENT_TYPE}" == "kubernetes" ]]; then
        # Get pods to be terminated
        local pods_to_terminate
        pods_to_terminate=$(kubectl get pods -n ollamamax -l app=ollama-worker --sort-by=.metadata.creationTimestamp -o name | head -n $((current_replicas - target_replicas)))
        
        # Drain each pod
        for pod in ${pods_to_terminate}; do
            log "Draining ${pod}"
            kubectl annotate "${pod}" -n ollamamax drain="true"
            # Wait for requests to complete (up to 60 seconds)
            sleep 60
        done
        
        kubectl scale deployment ollama-worker --replicas="${target_replicas}" -n ollamamax
    elif [[ "${DEPLOYMENT_TYPE}" == "docker" ]]; then
        docker-compose -f docker-compose.prod.yml up -d --scale ollama-worker="${target_replicas}"
    fi
    
    log "Scale down completed successfully"
}

# Get current replica count
get_current_replicas() {
    if [[ "${DEPLOYMENT_TYPE}" == "kubernetes" ]]; then
        kubectl get deployment ollama-worker -n ollamamax -o jsonpath='{.spec.replicas}'
    elif [[ "${DEPLOYMENT_TYPE}" == "docker" ]]; then
        docker-compose -f docker-compose.prod.yml ps ollama-worker | grep -c "Up" || echo "0"
    fi
}

# Main scaling logic
main() {
    local coordinator_url="${COORDINATOR_URL}"
    local current_replicas
    local load_metrics
    local cpu_usage
    local memory_usage
    local request_rate
    local queue_length
    
    current_replicas=$(get_current_replicas)
    load_metrics=$(check_load "${coordinator_url}")
    read -r cpu_usage memory_usage request_rate queue_length <<< "${load_metrics}"
    
    log "Current metrics: CPU=${cpu_usage}%, Memory=${memory_usage}%, Rate=${request_rate}req/s, Queue=${queue_length}"
    
    # Scale up conditions
    if (( $(echo "${cpu_usage} > ${SCALE_UP_CPU_THRESHOLD}" | bc -l) )) || \
       (( $(echo "${memory_usage} > ${SCALE_UP_MEMORY_THRESHOLD}" | bc -l) )) || \
       (( $(echo "${queue_length} > ${SCALE_UP_QUEUE_THRESHOLD}" | bc -l) )); then
        
        if (( current_replicas < MAX_REPLICAS )); then
            local target_replicas=$((current_replicas + SCALE_UP_STEP))
            if (( target_replicas > MAX_REPLICAS )); then
                target_replicas=${MAX_REPLICAS}
            fi
            scale_up "${current_replicas}" "${target_replicas}"
        else
            log "Already at maximum replicas (${MAX_REPLICAS})"
        fi
        
    # Scale down conditions
    elif (( $(echo "${cpu_usage} < ${SCALE_DOWN_CPU_THRESHOLD}" | bc -l) )) && \
         (( $(echo "${memory_usage} < ${SCALE_DOWN_MEMORY_THRESHOLD}" | bc -l) )) && \
         (( $(echo "${queue_length} < ${SCALE_DOWN_QUEUE_THRESHOLD}" | bc -l) )); then
        
        if (( current_replicas > MIN_REPLICAS )); then
            local target_replicas=$((current_replicas - SCALE_DOWN_STEP))
            if (( target_replicas < MIN_REPLICAS )); then
                target_replicas=${MIN_REPLICAS}
            fi
            scale_down "${current_replicas}" "${target_replicas}"
        else
            log "Already at minimum replicas (${MIN_REPLICAS})"
        fi
    else
        log "No scaling action needed (replicas: ${current_replicas})"
    fi
}

# Configuration file template
create_scaling_config() {
    cat > "${CONFIG_FILE}" << 'EOF'
# Scaling Configuration
DEPLOYMENT_TYPE="kubernetes"  # or "docker"
COORDINATOR_URL="http://localhost:8080"

# Scaling thresholds
SCALE_UP_CPU_THRESHOLD=70.0
SCALE_UP_MEMORY_THRESHOLD=80.0
SCALE_UP_QUEUE_THRESHOLD=10

SCALE_DOWN_CPU_THRESHOLD=30.0
SCALE_DOWN_MEMORY_THRESHOLD=50.0
SCALE_DOWN_QUEUE_THRESHOLD=2

# Scaling parameters
MIN_REPLICAS=2
MAX_REPLICAS=20
SCALE_UP_STEP=2
SCALE_DOWN_STEP=1

# Logging
LOG_FILE="/var/log/ollama-scaling.log"
EOF
}

# Create config if it doesn't exist
if [[ ! -f "${CONFIG_FILE}" ]]; then
    create_scaling_config
    log "Created scaling configuration at ${CONFIG_FILE}"
fi

# Run main function
main "$@"
```

### Vertical Scaling Configuration

```yaml
# config/vertical-scaling.yaml
apiVersion: autoscaling/v1
kind: VerticalPodAutoscaler
metadata:
  name: ollama-worker-vpa
  namespace: ollamamax
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: ollama-worker
  updatePolicy:
    updateMode: "Auto"
  resourcePolicy:
    containerPolicies:
    - containerName: worker
      maxAllowed:
        cpu: "8"
        memory: "16Gi"
      minAllowed:
        cpu: "1"
        memory: "2Gi"
      controlledResources: ["cpu", "memory"]
```

## 📊 Monitoring and Alerting

### Prometheus Configuration

```yaml
# monitoring/prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "ollama_alerts.yml"

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093

scrape_configs:
  - job_name: 'ollama-coordinator'
    kubernetes_sd_configs:
      - role: pod
        namespaces:
          names:
          - ollamamax
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_label_app]
        action: keep
        regex: ollama-coordinator
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)

  - job_name: 'ollama-worker'
    kubernetes_sd_configs:
      - role: pod
        namespaces:
          names:
          - ollamamax
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_label_app]
        action: keep
        regex: ollama-worker
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_port]
        action: replace
        target_label: __address__
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2

  - job_name: 'kubernetes-nodes'
    kubernetes_sd_configs:
      - role: node
    relabel_configs:
      - action: labelmap
        regex: __meta_kubernetes_node_label_(.+)

  - job_name: 'nvidia-gpu'
    static_configs:
      - targets: ['dcgm-exporter:9400']
```

### Alert Rules

```yaml
# monitoring/ollama_alerts.yml
groups:
  - name: ollama.rules
    rules:
    - alert: OllamaCoordinatorDown
      expr: up{job="ollama-coordinator"} == 0
      for: 1m
      labels:
        severity: critical
      annotations:
        summary: "Ollama coordinator is down"
        description: "Ollama coordinator has been down for more than 1 minute."

    - alert: OllamaWorkerDown
      expr: up{job="ollama-worker"} == 0
      for: 2m
      labels:
        severity: warning
      annotations:
        summary: "Ollama worker is down"
        description: "Ollama worker {{ $labels.instance }} has been down for more than 2 minutes."

    - alert: HighCPUUsage
      expr: (100 - (avg by(instance) (irate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)) > 80
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: "High CPU usage detected"
        description: "CPU usage is above 80% for more than 5 minutes on {{ $labels.instance }}."

    - alert: HighMemoryUsage
      expr: (1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100 > 85
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: "High memory usage detected"
        description: "Memory usage is above 85% for more than 5 minutes on {{ $labels.instance }}."

    - alert: GPUUtilizationHigh
      expr: DCGM_FI_DEV_GPU_UTIL > 90
      for: 10m
      labels:
        severity: warning
      annotations:
        summary: "High GPU utilization"
        description: "GPU utilization is above 90% for more than 10 minutes on {{ $labels.gpu }}."

    - alert: RequestLatencyHigh
      expr: histogram_quantile(0.95, sum(rate(ollama_request_duration_seconds_bucket[5m])) by (le)) > 2
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: "High request latency"
        description: "95th percentile latency is above 2 seconds for more than 5 minutes."

    - alert: ErrorRateHigh
      expr: sum(rate(ollama_requests_total{status=~"5.."}[5m])) / sum(rate(ollama_requests_total[5m])) > 0.05
      for: 2m
      labels:
        severity: critical
      annotations:
        summary: "High error rate"
        description: "Error rate is above 5% for more than 2 minutes."

    - alert: DiskSpaceLow
      expr: (node_filesystem_avail_bytes / node_filesystem_size_bytes) * 100 < 10
      for: 5m
      labels:
        severity: critical
      annotations:
        summary: "Low disk space"
        description: "Disk space is below 10% on {{ $labels.instance }}:{{ $labels.mountpoint }}."

    - alert: ModelLoadFailure
      expr: increase(ollama_model_load_failures_total[5m]) > 0
      for: 0m
      labels:
        severity: critical
      annotations:
        summary: "Model load failure"
        description: "Model load failure detected on {{ $labels.instance }}."
```

### Grafana Dashboard

```json
{
  "dashboard": {
    "id": null,
    "title": "OllamaMax Distributed - Production Dashboard",
    "tags": ["ollama", "production"],
    "style": "dark",
    "timezone": "browser",
    "panels": [
      {
        "id": 1,
        "title": "System Overview",
        "type": "stat",
        "targets": [
          {
            "expr": "count(up{job=\"ollama-coordinator\"} == 1)",
            "legendFormat": "Coordinators Online"
          },
          {
            "expr": "count(up{job=\"ollama-worker\"} == 1)",
            "legendFormat": "Workers Online"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "color": {
              "mode": "thresholds"
            },
            "thresholds": {
              "steps": [
                {"color": "red", "value": 0},
                {"color": "yellow", "value": 1},
                {"color": "green", "value": 2}
              ]
            }
          }
        }
      },
      {
        "id": 2,
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "sum(rate(ollama_requests_total[5m]))",
            "legendFormat": "Requests/sec"
          }
        ]
      },
      {
        "id": 3,
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.50, sum(rate(ollama_request_duration_seconds_bucket[5m])) by (le))",
            "legendFormat": "50th percentile"
          },
          {
            "expr": "histogram_quantile(0.95, sum(rate(ollama_request_duration_seconds_bucket[5m])) by (le))",
            "legendFormat": "95th percentile"
          },
          {
            "expr": "histogram_quantile(0.99, sum(rate(ollama_request_duration_seconds_bucket[5m])) by (le))",
            "legendFormat": "99th percentile"
          }
        ]
      },
      {
        "id": 4,
        "title": "Resource Utilization",
        "type": "graph",
        "targets": [
          {
            "expr": "100 - (avg(irate(node_cpu_seconds_total{mode=\"idle\"}[5m])) * 100)",
            "legendFormat": "CPU Usage %"
          },
          {
            "expr": "(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100",
            "legendFormat": "Memory Usage %"
          }
        ]
      },
      {
        "id": 5,
        "title": "GPU Utilization",
        "type": "graph",
        "targets": [
          {
            "expr": "DCGM_FI_DEV_GPU_UTIL",
            "legendFormat": "GPU {{gpu}} Utilization"
          },
          {
            "expr": "DCGM_FI_DEV_MEM_COPY_UTIL",
            "legendFormat": "GPU {{gpu}} Memory"
          }
        ]
      }
    ],
    "time": {
      "from": "now-1h",
      "to": "now"
    },
    "refresh": "5s"
  }
}
```

## 🛡️ Security Hardening

### Security Configuration

```yaml
# config/security.yaml
security:
  # TLS Configuration
  tls:
    enabled: true
    min_version: "1.2"
    cert_file: "/etc/ssl/certs/tls.crt"
    key_file: "/etc/ssl/certs/tls.key"
    ca_file: "/etc/ssl/certs/ca.crt"
    client_auth_required: true
    
  # Authentication
  authentication:
    enabled: true
    type: "jwt"
    jwt:
      secret_key: "${JWT_SECRET_KEY}"
      expiration: "1h"
      refresh_expiration: "24h"
      issuer: "ollamamax"
      
  # Authorization
  authorization:
    enabled: true
    rbac:
      enabled: true
      roles:
        admin:
          permissions: ["*"]
        user:
          permissions: ["read", "execute"]
        readonly:
          permissions: ["read"]
          
  # Rate Limiting
  rate_limiting:
    enabled: true
    global_rate: "1000/minute"
    per_user_rate: "100/minute"
    burst_size: 50
    
  # CORS
  cors:
    enabled: true
    allowed_origins:
      - "https://app.yourdomain.com"
      - "https://admin.yourdomain.com"
    allowed_methods: ["GET", "POST", "PUT", "DELETE"]
    allowed_headers: ["Authorization", "Content-Type"]
    max_age: 86400
    
  # Content Security Policy
  csp:
    enabled: true
    policy: "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'"
    
  # Request Validation
  validation:
    max_request_size: "100MB"
    max_header_size: "1MB"
    timeout: "30s"
    
  # Audit Logging
  audit:
    enabled: true
    log_level: "info"
    include_request_body: false
    include_response_body: false
    sensitive_fields: ["password", "token", "key"]
```

### Network Security Policies

```yaml
# k8s/network-policy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: ollama-network-policy
  namespace: ollamamax
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
  
  ingress:
  # Allow ingress from nginx
  - from:
    - namespaceSelector:
        matchLabels:
          name: nginx-ingress
    ports:
    - protocol: TCP
      port: 8080
    - protocol: TCP
      port: 8443
      
  # Allow metrics scraping from prometheus
  - from:
    - namespaceSelector:
        matchLabels:
          name: monitoring
    ports:
    - protocol: TCP
      port: 9090
      
  # Allow inter-pod communication
  - from:
    - podSelector:
        matchLabels:
          app: ollama-coordinator
    - podSelector:
        matchLabels:
          app: ollama-worker
    ports:
    - protocol: TCP
      port: 8080
      
  egress:
  # Allow DNS resolution
  - to: []
    ports:
    - protocol: UDP
      port: 53
      
  # Allow HTTPS outbound (for model downloads)
  - to: []
    ports:
    - protocol: TCP
      port: 443
      
  # Allow communication to Redis
  - to:
    - podSelector:
        matchLabels:
          app: redis
    ports:
    - protocol: TCP
      port: 6379
```

### Pod Security Standards

```yaml
# k8s/pod-security-policy.yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: ollama-psp
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
    - ALL
  volumes:
    - 'configMap'
    - 'emptyDir'
    - 'projected'
    - 'secret'
    - 'downwardAPI'
    - 'persistentVolumeClaim'
  runAsUser:
    rule: 'MustRunAsNonRoot'
  seLinux:
    rule: 'RunAsAny'
  fsGroup:
    rule: 'RunAsAny'
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: ollama-psp-use
rules:
- apiGroups: ['policy']
  resources: ['podsecuritypolicies']
  verbs: ['use']
  resourceNames:
  - ollama-psp
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: ollama-psp-use
roleRef:
  kind: ClusterRole
  name: ollama-psp-use
  apiGroup: rbac.authorization.k8s.io
subjects:
- kind: ServiceAccount
  name: ollama-service-account
  namespace: ollamamax
```

## 💾 Backup and Disaster Recovery

### Backup Strategy

```bash
#!/bin/bash
# scripts/backup.sh

set -euo pipefail

BACKUP_DIR="${BACKUP_DIR:-/backup/ollama}"
RETENTION_DAYS="${RETENTION_DAYS:-30}"
S3_BUCKET="${S3_BUCKET:-ollama-backups}"
ENCRYPTION_KEY="${ENCRYPTION_KEY:-/etc/ssl/backup.key}"

# Logging
exec 1> >(logger -s -t ollama-backup)
exec 2>&1

log() {
    echo "[$(date -Iseconds)] $*"
}

# Create backup directory structure
prepare_backup() {
    local backup_date="$1"
    local backup_path="${BACKUP_DIR}/${backup_date}"
    
    mkdir -p "${backup_path}"/{config,data,models,logs}
    echo "${backup_path}"
}

# Backup configuration files
backup_config() {
    local backup_path="$1"
    
    log "Backing up configuration files..."
    
    if [[ "${DEPLOYMENT_TYPE}" == "kubernetes" ]]; then
        kubectl get configmaps -n ollamamax -o yaml > "${backup_path}/config/configmaps.yaml"
        kubectl get secrets -n ollamamax -o yaml > "${backup_path}/config/secrets.yaml"
        kubectl get deployments -n ollamamax -o yaml > "${backup_path}/config/deployments.yaml"
        kubectl get services -n ollamamax -o yaml > "${backup_path}/config/services.yaml"
    elif [[ "${DEPLOYMENT_TYPE}" == "docker" ]]; then
        cp docker-compose.prod.yml "${backup_path}/config/"
        cp -r config/ "${backup_path}/config/"
    fi
    
    log "Configuration backup completed"
}

# Backup persistent data
backup_data() {
    local backup_path="$1"
    
    log "Backing up persistent data..."
    
    # Database backup (Redis)
    if command -v redis-cli &> /dev/null; then
        redis-cli --rdb "${backup_path}/data/redis.rdb"
    fi
    
    # Application data backup
    if [[ -d "/app/data" ]]; then
        tar -czf "${backup_path}/data/app-data.tar.gz" -C /app/data .
    fi
    
    # Model files backup (only if changed)
    if [[ -d "/app/models" ]]; then
        find /app/models -type f -newer "${BACKUP_DIR}/.last-model-backup" 2>/dev/null | \
        tar -czf "${backup_path}/data/models-incremental.tar.gz" -T - 2>/dev/null || \
        tar -czf "${backup_path}/data/models-full.tar.gz" -C /app/models .
        
        touch "${BACKUP_DIR}/.last-model-backup"
    fi
    
    log "Data backup completed"
}

# Backup logs
backup_logs() {
    local backup_path="$1"
    
    log "Backing up logs..."
    
    # Application logs
    if [[ -d "/app/logs" ]]; then
        tar -czf "${backup_path}/logs/app-logs.tar.gz" -C /app/logs .
    fi
    
    # System logs
    journalctl -u ollama-distributed --since="24 hours ago" > "${backup_path}/logs/systemd.log" || true
    
    log "Logs backup completed"
}

# Encrypt backup
encrypt_backup() {
    local backup_path="$1"
    local encrypted_path="${backup_path}.tar.gz.enc"
    
    log "Encrypting backup..."
    
    tar -czf - -C "${BACKUP_DIR}" "$(basename "${backup_path}")" | \
    openssl enc -aes-256-cbc -salt -in - -out "${encrypted_path}" -pass file:"${ENCRYPTION_KEY}"
    
    rm -rf "${backup_path}"
    echo "${encrypted_path}"
}

# Upload to S3
upload_backup() {
    local encrypted_path="$1"
    local s3_key="backups/$(basename "${encrypted_path}")"
    
    log "Uploading backup to S3..."
    
    aws s3 cp "${encrypted_path}" "s3://${S3_BUCKET}/${s3_key}" \
        --storage-class STANDARD_IA \
        --server-side-encryption AES256
    
    # Verify upload
    aws s3api head-object --bucket "${S3_BUCKET}" --key "${s3_key}" > /dev/null
    
    log "Upload completed: s3://${S3_BUCKET}/${s3_key}"
}

# Cleanup old backups
cleanup_backups() {
    log "Cleaning up old backups..."
    
    # Local cleanup
    find "${BACKUP_DIR}" -name "*.enc" -mtime +${RETENTION_DAYS} -delete
    
    # S3 cleanup
    aws s3api list-objects-v2 --bucket "${S3_BUCKET}" --prefix "backups/" \
        --query "Contents[?LastModified<='$(date -d "${RETENTION_DAYS} days ago" -Iseconds)'].Key" \
        --output text | \
    xargs -I {} aws s3 rm "s3://${S3_BUCKET}/{}"
    
    log "Cleanup completed"
}

# Verify backup integrity
verify_backup() {
    local encrypted_path="$1"
    local temp_dir
    
    log "Verifying backup integrity..."
    
    temp_dir=$(mktemp -d)
    
    # Decrypt and extract
    openssl enc -aes-256-cbc -d -in "${encrypted_path}" -out "${temp_dir}/backup.tar.gz" -pass file:"${ENCRYPTION_KEY}"
    tar -tzf "${temp_dir}/backup.tar.gz" > /dev/null
    
    rm -rf "${temp_dir}"
    log "Backup verification successful"
}

# Main backup function
main() {
    local backup_date
    local backup_path
    local encrypted_path
    
    backup_date=$(date +%Y%m%d-%H%M%S)
    
    log "Starting backup process: ${backup_date}"
    
    # Prepare backup directory
    backup_path=$(prepare_backup "${backup_date}")
    
    # Perform backups
    backup_config "${backup_path}"
    backup_data "${backup_path}"
    backup_logs "${backup_path}"
    
    # Encrypt backup
    encrypted_path=$(encrypt_backup "${backup_path}")
    
    # Verify backup
    verify_backup "${encrypted_path}"
    
    # Upload to S3
    upload_backup "${encrypted_path}"
    
    # Cleanup
    cleanup_backups
    
    log "Backup process completed successfully"
}

# Configuration check
if [[ -z "${DEPLOYMENT_TYPE:-}" ]]; then
    log "Error: DEPLOYMENT_TYPE not set"
    exit 1
fi

if [[ ! -f "${ENCRYPTION_KEY}" ]]; then
    log "Error: Encryption key not found at ${ENCRYPTION_KEY}"
    exit 1
fi

# Execute main function
main "$@"
```

### Disaster Recovery Procedures

```bash
#!/bin/bash
# scripts/disaster-recovery.sh

set -euo pipefail

RECOVERY_MODE="${1:-full}"  # full, partial, config-only
BACKUP_SOURCE="${2:-s3}"   # s3, local
BACKUP_DATE="${3:-latest}" # YYYYMMDD-HHMMSS or latest

# Logging
exec 1> >(logger -s -t ollama-dr)
exec 2>&1

log() {
    echo "[$(date -Iseconds)] $*"
}

# Download backup from S3
download_backup() {
    local backup_file="$1"
    local local_path="${BACKUP_DIR}/${backup_file}"
    
    log "Downloading backup from S3..."
    
    aws s3 cp "s3://${S3_BUCKET}/backups/${backup_file}" "${local_path}"
    echo "${local_path}"
}

# Find latest backup
find_latest_backup() {
    if [[ "${BACKUP_SOURCE}" == "s3" ]]; then
        aws s3api list-objects-v2 --bucket "${S3_BUCKET}" --prefix "backups/" \
            --query "max_by(Contents, &LastModified).Key" --output text | \
            sed 's|backups/||'
    else
        ls -t "${BACKUP_DIR}"/*.enc | head -1 | xargs basename
    fi
}

# Decrypt and extract backup
extract_backup() {
    local encrypted_path="$1"
    local extract_dir="${BACKUP_DIR}/restore"
    
    log "Extracting backup..."
    
    rm -rf "${extract_dir}"
    mkdir -p "${extract_dir}"
    
    openssl enc -aes-256-cbc -d -in "${encrypted_path}" -out "${extract_dir}/backup.tar.gz" -pass file:"${ENCRYPTION_KEY}"
    tar -xzf "${extract_dir}/backup.tar.gz" -C "${extract_dir}"
    
    echo "${extract_dir}"
}

# Stop services
stop_services() {
    log "Stopping OllamaMax services..."
    
    if [[ "${DEPLOYMENT_TYPE}" == "kubernetes" ]]; then
        kubectl scale deployment ollama-coordinator --replicas=0 -n ollamamax
        kubectl scale deployment ollama-worker --replicas=0 -n ollamamax
        kubectl wait --for=delete pod -l app=ollama-coordinator -n ollamamax --timeout=300s
        kubectl wait --for=delete pod -l app=ollama-worker -n ollamamax --timeout=300s
    elif [[ "${DEPLOYMENT_TYPE}" == "docker" ]]; then
        docker-compose -f docker-compose.prod.yml stop
    fi
    
    log "Services stopped"
}

# Restore configuration
restore_config() {
    local restore_dir="$1"
    local config_dir="${restore_dir}/config"
    
    log "Restoring configuration..."
    
    if [[ "${DEPLOYMENT_TYPE}" == "kubernetes" ]]; then
        # Restore Kubernetes resources
        kubectl apply -f "${config_dir}/configmaps.yaml"
        kubectl apply -f "${config_dir}/secrets.yaml"
        kubectl apply -f "${config_dir}/deployments.yaml"
        kubectl apply -f "${config_dir}/services.yaml"
    elif [[ "${DEPLOYMENT_TYPE}" == "docker" ]]; then
        # Restore Docker configuration
        cp "${config_dir}/docker-compose.prod.yml" ./
        cp -r "${config_dir}/config" ./
    fi
    
    log "Configuration restored"
}

# Restore data
restore_data() {
    local restore_dir="$1"
    local data_dir="${restore_dir}/data"
    
    log "Restoring data..."
    
    # Restore Redis data
    if [[ -f "${data_dir}/redis.rdb" ]]; then
        if [[ "${DEPLOYMENT_TYPE}" == "kubernetes" ]]; then
            kubectl cp "${data_dir}/redis.rdb" ollamamax/redis-0:/data/dump.rdb
        else
            docker cp "${data_dir}/redis.rdb" ollama-redis:/data/dump.rdb
        fi
    fi
    
    # Restore application data
    if [[ -f "${data_dir}/app-data.tar.gz" ]]; then
        mkdir -p /app/data
        tar -xzf "${data_dir}/app-data.tar.gz" -C /app/data
    fi
    
    # Restore models
    if [[ -f "${data_dir}/models-full.tar.gz" ]]; then
        mkdir -p /app/models
        tar -xzf "${data_dir}/models-full.tar.gz" -C /app/models
    elif [[ -f "${data_dir}/models-incremental.tar.gz" ]]; then
        mkdir -p /app/models
        tar -xzf "${data_dir}/models-incremental.tar.gz" -C /app/models
    fi
    
    log "Data restored"
}

# Start services
start_services() {
    log "Starting OllamaMax services..."
    
    if [[ "${DEPLOYMENT_TYPE}" == "kubernetes" ]]; then
        kubectl scale deployment ollama-coordinator --replicas=2 -n ollamamax
        kubectl scale deployment ollama-worker --replicas=3 -n ollamamax
        kubectl rollout status deployment/ollama-coordinator -n ollamamax --timeout=300s
        kubectl rollout status deployment/ollama-worker -n ollamamax --timeout=300s
    elif [[ "${DEPLOYMENT_TYPE}" == "docker" ]]; then
        docker-compose -f docker-compose.prod.yml up -d
    fi
    
    log "Services started"
}

# Verify recovery
verify_recovery() {
    log "Verifying disaster recovery..."
    
    local coordinator_url
    if [[ "${DEPLOYMENT_TYPE}" == "kubernetes" ]]; then
        coordinator_url="http://$(kubectl get service ollama-coordinator -n ollamamax -o jsonpath='{.spec.clusterIP}'):8080"
    else
        coordinator_url="http://localhost:8080"
    fi
    
    # Wait for services to be ready
    local max_attempts=30
    local attempt=1
    
    while (( attempt <= max_attempts )); do
        if curl -f "${coordinator_url}/health" &> /dev/null; then
            log "Health check passed on attempt ${attempt}"
            break
        fi
        
        log "Health check failed, attempt ${attempt}/${max_attempts}"
        sleep 10
        ((attempt++))
    done
    
    if (( attempt > max_attempts )); then
        log "Error: Health check failed after ${max_attempts} attempts"
        return 1
    fi
    
    # Test API functionality
    if curl -f "${coordinator_url}/api/models" &> /dev/null; then
        log "API test passed"
    else
        log "Warning: API test failed"
    fi
    
    log "Disaster recovery verification completed"
}

# Main recovery function
main() {
    local backup_file
    local backup_path
    local restore_dir
    
    log "Starting disaster recovery: mode=${RECOVERY_MODE}, source=${BACKUP_SOURCE}, date=${BACKUP_DATE}"
    
    # Find backup file
    if [[ "${BACKUP_DATE}" == "latest" ]]; then
        backup_file=$(find_latest_backup)
    else
        backup_file="${BACKUP_DATE}.tar.gz.enc"
    fi
    
    log "Using backup file: ${backup_file}"
    
    # Download backup if needed
    if [[ "${BACKUP_SOURCE}" == "s3" ]]; then
        backup_path=$(download_backup "${backup_file}")
    else
        backup_path="${BACKUP_DIR}/${backup_file}"
    fi
    
    # Extract backup
    restore_dir=$(extract_backup "${backup_path}")
    
    # Stop services
    stop_services
    
    # Perform restore based on mode
    case "${RECOVERY_MODE}" in
        "full")
            restore_config "${restore_dir}"
            restore_data "${restore_dir}"
            ;;
        "partial")
            restore_data "${restore_dir}"
            ;;
        "config-only")
            restore_config "${restore_dir}"
            ;;
        *)
            log "Error: Invalid recovery mode: ${RECOVERY_MODE}"
            exit 1
            ;;
    esac
    
    # Start services
    start_services
    
    # Verify recovery
    verify_recovery
    
    # Cleanup
    rm -rf "${restore_dir}"
    
    log "Disaster recovery completed successfully"
}

# Execute main function
main "$@"
```

## 🎯 Module Assessment

### Practical Exercise: Production Deployment

Complete this hands-on assessment to demonstrate your production deployment skills:

#### Exercise 1: Container Deployment (25 points)

1. **Docker Production Setup**:
   ```bash
   # Build and deploy the production container
   docker build -t ollamamax/distributed:prod .
   docker-compose -f docker-compose.prod.yml up -d
   
   # Verify deployment
   docker ps
   docker logs ollama-coordinator
   ```

2. **Health Check Validation**:
   ```bash
   # Test health endpoints
   curl http://localhost:8080/health
   curl http://localhost:8080/ready
   curl http://localhost:9090/metrics
   ```

3. **Security Verification**:
   - Verify containers run as non-root
   - Check TLS certificate configuration
   - Test authentication endpoints

#### Exercise 2: Kubernetes Orchestration (25 points)

1. **Deploy to Kubernetes**:
   ```bash
   # Apply all manifests
   kubectl apply -f k8s/namespace.yaml
   kubectl apply -f k8s/configmap.yaml
   kubectl apply -f k8s/secret.yaml
   kubectl apply -f k8s/
   
   # Verify deployment
   kubectl get all -n ollamamax
   kubectl describe pods -n ollamamax
   ```

2. **Scaling Test**:
   ```bash
   # Test horizontal scaling
   kubectl scale deployment ollama-worker --replicas=5 -n ollamamax
   kubectl get pods -n ollamamax -w
   ```

3. **Rolling Update**:
   ```bash
   # Perform rolling update
   kubectl set image deployment/ollama-worker worker=ollamamax/distributed:v1.0.1 -n ollamamax
   kubectl rollout status deployment/ollama-worker -n ollamamax
   ```

#### Exercise 3: Monitoring Setup (25 points)

1. **Deploy Monitoring Stack**:
   ```bash
   # Deploy Prometheus and Grafana
   kubectl apply -f monitoring/prometheus.yml
   kubectl apply -f monitoring/grafana-dashboard.json
   ```

2. **Configure Alerts**:
   ```bash
   # Test alert rules
   kubectl apply -f monitoring/ollama_alerts.yml
   promtool check rules monitoring/ollama_alerts.yml
   ```

3. **Dashboard Verification**:
   - Access Grafana dashboard
   - Verify metrics collection
   - Test alert triggers

#### Exercise 4: Disaster Recovery Test (25 points)

1. **Backup Creation**:
   ```bash
   # Create production backup
   ./scripts/backup.sh
   ls -la /backup/ollama/
   ```

2. **Simulate Failure**:
   ```bash
   # Stop all services
   kubectl delete namespace ollamamax
   # or
   docker-compose -f docker-compose.prod.yml down -v
   ```

3. **Recovery Execution**:
   ```bash
   # Restore from backup
   ./scripts/disaster-recovery.sh full s3 latest
   ```

4. **Verification**:
   ```bash
   # Test full functionality
   curl http://localhost:8080/api/models
   curl -X POST http://localhost:8080/api/chat -d '{"model":"llama2","messages":[{"role":"user","content":"Hello"}]}'
   ```

### Assessment Checklist

- [ ] **Container Security**: Non-root user, minimal base image, security scanning
- [ ] **High Availability**: Multiple replicas, health checks, graceful shutdown
- [ ] **Scalability**: HPA configured, resource limits set, scaling tested
- [ ] **Monitoring**: Metrics exported, dashboards working, alerts configured
- [ ] **Security**: TLS enabled, authentication working, network policies applied
- [ ] **Backup/Recovery**: Automated backups, successful recovery test
- [ ] **Performance**: Load testing passed, optimization applied
- [ ] **Documentation**: Runbooks created, procedures documented

### Performance Benchmarks

Your production deployment should meet these benchmarks:

- **Availability**: 99.9% uptime (8.77 hours downtime/year)
- **Response Time**: 95th percentile < 500ms for API calls
- **Throughput**: Handle 1000+ concurrent requests
- **Recovery Time**: Complete disaster recovery < 30 minutes
- **Scaling Time**: Scale up/down operations < 2 minutes
- **Security**: Pass all security scans with no critical vulnerabilities

## 🎉 Congratulations!

You've completed the advanced production deployment module! You now have the expertise to:

✅ Deploy OllamaMax Distributed in production environments  
✅ Implement enterprise-grade security and monitoring  
✅ Handle scaling and high availability requirements  
✅ Manage backup and disaster recovery procedures  
✅ Optimize performance at scale  

### Next Steps

1. **Production Deployment**: Apply these skills to deploy in your environment
2. **SRE Practices**: Implement Site Reliability Engineering principles
3. **Advanced Monitoring**: Explore APM tools and distributed tracing
4. **Capacity Planning**: Develop long-term scaling strategies
5. **Security Auditing**: Regular security assessments and compliance checks

### Resources

- [Kubernetes Production Best Practices](https://kubernetes.io/docs/setup/best-practices/)
- [Docker Security Guide](https://docs.docker.com/engine/security/)
- [Prometheus Monitoring](https://prometheus.io/docs/practices/)
- [Site Reliability Engineering](https://sre.google/)

**Continue to**: [Training Program Overview](./README.md) | **Previous**: [Module 6: Advanced Configuration](./module-6-advanced-config.md)