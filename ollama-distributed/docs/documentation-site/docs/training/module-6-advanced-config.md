# Module 6: Advanced Configuration and Profiles

> ⚠️ **Security Warning**: This module covers production configurations that require proper secret management. **Never use default or hardcoded secrets in production environments**. See [Security Guidelines](../../../../SECURITY-GUIDELINES.md) for secure deployment practices.

**Duration**: 15 minutes  
**Objective**: Master advanced configuration management, environment-specific settings, security configuration, performance tuning, multi-node setup, and monitoring configuration

Welcome to Module 6! This is where you'll learn the advanced aspects of configuring OllamaMax Distributed for production environments, security hardening, and optimal performance.

## 🎯 What You'll Learn

By the end of this module, you will:
- ✅ Master configuration profiles for different environments
- ✅ Implement environment-specific settings and variable management
- ✅ Configure advanced security settings including TLS, authentication, and encryption
- ✅ Optimize performance settings for different workloads
- ✅ Set up multi-node configurations with proper coordination
- ✅ Configure comprehensive monitoring and logging systems
- ✅ Validate configurations and troubleshoot advanced issues

## 📁 Configuration Architecture Overview

### Configuration Hierarchy

OllamaMax uses a hierarchical configuration system:

```
Configuration Priority (highest to lowest):
1. Command-line flags
2. Environment variables  
3. Configuration file
4. Default values
```

### Configuration File Locations

```bash
# Check configuration file search paths
./bin/ollama-distributed config --show-paths
```

**Expected paths:**
```
Configuration file search order:
1. ./config.yaml (current directory)
2. ~/.ollamamax/config.yaml (user directory)
3. /etc/ollamamax/config.yaml (system directory)
4. $OLLAMA_CONFIG_FILE (environment variable)
```

## 🏗️ Configuration Profiles Deep Dive

### Step 1: Understanding Built-in Profiles

Let's explore the available configuration profiles:

```bash
# List available configuration profiles
ls -la /home/kp/ollamamax/ollama-distributed/config/

# Examine development profile
cat /home/kp/ollamamax/ollama-distributed/config/development.yaml
```

**✅ Checkpoint 1**: You can see different configuration profiles including development.yaml, production.yaml, and security.yaml.

### Step 2: Development Profile Configuration

```bash
# Create a development environment configuration
cat > dev-advanced-config.yaml << EOF
# Advanced Development Configuration
node:
  id: "dev-advanced-node"
  name: "Advanced Development Node"
  region: "local"
  zone: "dev"
  environment: "development"
  tags:
    purpose: "advanced-training"
    owner: "developer"
    version: "1.0.0"

api:
  listen: "0.0.0.0:8080"
  timeout: "30s"
  max_body_size: 104857600  # 100MB
  tls:
    enabled: false  # Disabled for development
  cors:
    enabled: true
    allowed_origins: ["http://localhost:3000", "http://127.0.0.1:3000"]
    allowed_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    allowed_headers: ["Authorization", "Content-Type", "X-Requested-With"]
  rate_limit:
    enabled: false  # Disabled for development

# P2P configuration for development
p2p:
  enabled: true
  listen: "0.0.0.0:9000"
  bootstrap_peers: []
  discovery:
    enabled: true
    mdns_enabled: true
    dht_enabled: false

# Development logging - verbose for debugging
logging:
  level: "debug"
  format: "text"
  output: "console"
  file:
    enabled: true
    path: "./logs/dev-advanced.log"

# Development storage paths
storage:
  data_dir: "./data/dev"
  model_dir: "./data/dev/models"
  cache_dir: "./data/dev/cache"
  max_cache_size: 1073741824  # 1GB

# Metrics for development
metrics:
  enabled: true
  listen: "0.0.0.0:9090"
  path: "/metrics"
  interval: "15s"
EOF

# Validate development configuration
echo "Development configuration created. Validating..."
```

**✅ Checkpoint 2**: Development configuration created with appropriate settings for local development.

### Step 3: Production Profile Configuration

```bash
# Create production-ready configuration
cat > prod-advanced-config.yaml << EOF
# Advanced Production Configuration
node:
  id: "\${OLLAMA_NODE_ID}"
  name: "\${OLLAMA_NODE_NAME}"
  region: "\${OLLAMA_NODE_REGION}"
  zone: "\${OLLAMA_NODE_ZONE}"
  environment: "production"
  tags:
    cluster: "\${OLLAMA_CLUSTER_NAME}"
    version: "\${OLLAMA_VERSION}"
    role: "\${OLLAMA_NODE_ROLE}"

api:
  listen: "0.0.0.0:8080"
  timeout: "60s"
  max_body_size: 536870912  # 512MB
  tls:
    enabled: true
    cert_file: "/etc/ssl/certs/ollama.crt"
    key_file: "/etc/ssl/private/ollama.key"
    min_version: "1.3"
  cors:
    enabled: true
    allowed_origins: ["\${CORS_ALLOWED_ORIGINS}"]
    allowed_methods: ["GET", "POST", "PUT", "DELETE"]
    allowed_headers: ["Authorization", "Content-Type"]
  rate_limit:
    enabled: true
    requests_per_minute: 1000
    burst: 100

# Production P2P configuration
p2p:
  enabled: true
  listen: "0.0.0.0:9000"
  bootstrap_peers: "\${OLLAMA_BOOTSTRAP_PEERS}"
  discovery:
    enabled: true
    mdns_enabled: false
    dht_enabled: true
    rendezvous: "ollama-distributed-v1"

# Production security
security:
  auth:
    enabled: true
    method: "jwt"
    secret_key: "\${OLLAMA_JWT_SECRET}"
    token_expiry: "1h"
  encryption:
    enabled: true
    algorithm: "aes-256-gcm"

# Production logging - structured and persistent
logging:
  level: "info"
  format: "json"
  output: "file"
  file:
    enabled: true
    path: "/var/log/ollama/ollama.log"
    max_size: 100
    max_age: 30
    max_backups: 10
    compress: true

# Production storage
storage:
  data_dir: "/var/lib/ollama"
  model_dir: "/var/lib/ollama/models"
  cache_dir: "/var/cache/ollama"
  max_cache_size: 107374182400  # 100GB

# Production metrics - internal only
metrics:
  enabled: true
  listen: "127.0.0.1:9090"
  path: "/metrics"
  interval: "30s"
EOF

echo "Production configuration created."
```

**✅ Checkpoint 3**: Production configuration created with environment variable placeholders and production-appropriate settings.

### Step 4: Edge/IoT Profile Configuration

```bash
# Create edge/IoT optimized configuration
cat > edge-config.yaml << EOF
# Edge/IoT Optimized Configuration
node:
  id: "edge-node-\${HOSTNAME}"
  name: "Edge Node \${HOSTNAME}"
  region: "edge"
  zone: "local"
  environment: "edge"
  tags:
    device_type: "edge"
    resource_class: "limited"

api:
  listen: "0.0.0.0:8080"
  timeout: "45s"
  max_body_size: 16777216  # 16MB - reduced for edge
  tls:
    enabled: true
    cert_file: "/etc/ssl/certs/edge.crt"
    key_file: "/etc/ssl/private/edge.key"
  rate_limit:
    enabled: true
    requests_per_minute: 100  # Lower limits for edge
    burst: 20

# Edge P2P - optimized for intermittent connectivity
p2p:
  enabled: true
  listen: "0.0.0.0:9000"
  bootstrap_peers: ["\${EDGE_BOOTSTRAP_PEER}"]
  discovery:
    enabled: true
    mdns_enabled: true
    dht_enabled: true
  connection_manager:
    low_water: 5
    high_water: 25
    grace_period: "60s"

# Resource-constrained scheduler
scheduler:
  enabled: true
  algorithm: "resource_aware"
  max_concurrent_tasks: 2  # Limited concurrency
  task_timeout: "10m"
  memory_threshold: 0.8

# Edge-optimized storage
storage:
  data_dir: "/var/lib/ollama"
  model_dir: "/var/lib/ollama/models"
  cache_dir: "/tmp/ollama-cache"
  max_cache_size: 1073741824  # 1GB - limited cache
  cleanup_policy: "aggressive"

# Minimal logging for resource conservation
logging:
  level: "warn"
  format: "json"
  output: "file"
  file:
    enabled: true
    path: "/var/log/ollama/edge.log"
    max_size: 10  # 10MB
    max_backups: 2
    compress: true

# Edge metrics - basic monitoring
metrics:
  enabled: true
  listen: "127.0.0.1:9090"
  path: "/metrics"
  interval: "60s"  # Less frequent collection
  resource_monitoring: true
EOF

echo "Edge configuration created for resource-constrained environments."
```

**✅ Checkpoint 4**: Edge configuration optimized for resource-constrained environments.

### Step 5: GPU-Optimized Profile Configuration

```bash
# Create GPU-optimized configuration
cat > gpu-config.yaml << EOF
# GPU-Optimized Configuration
node:
  id: "gpu-node-\${HOSTNAME}"
  name: "GPU Accelerated Node"
  region: "\${AWS_REGION}"
  zone: "\${AWS_AZ}"
  environment: "gpu-optimized"
  tags:
    gpu_count: "\${GPU_COUNT}"
    gpu_type: "\${GPU_TYPE}"
    instance_type: "\${INSTANCE_TYPE}"

api:
  listen: "0.0.0.0:8080"
  timeout: "120s"  # Longer timeouts for GPU processing
  max_body_size: 1073741824  # 1GB for large models
  workers: 8  # More API workers

# GPU-optimized scheduler
scheduler:
  enabled: true
  algorithm: "gpu_aware"
  max_concurrent_tasks: 16  # Higher concurrency
  task_timeout: "30m"
  gpu_scheduling:
    enabled: true
    memory_fraction: 0.9
    compute_mode: "default"

# High-performance storage
storage:
  data_dir: "/nvme/ollama"  # NVMe storage
  model_dir: "/nvme/ollama/models"
  cache_dir: "/nvme/ollama/cache"
  max_cache_size: 107374182400  # 100GB cache
  cache_policy: "lru"

# Performance monitoring
performance:
  gpu_monitoring: true
  memory_profiling: true
  compute_profiling: true
  
metrics:
  enabled: true
  listen: "0.0.0.0:9091"
  path: "/metrics"
  interval: "10s"
  gpu_metrics: true
  detailed_metrics: true

# High-performance logging
logging:
  level: "info"
  format: "json"
  output: "file"
  file:
    enabled: true
    path: "/var/log/ollama/gpu-node.log"
    max_size: 250
    max_backups: 5
    async_write: true
EOF

echo "GPU-optimized configuration created for high-performance workloads."
```

**✅ Checkpoint 5**: GPU configuration created with GPU-aware scheduling and performance optimizations.

## 🔒 Advanced Security Configuration

### Step 6: Enterprise Security Setup

```bash
# Create comprehensive security configuration
cat > security-advanced-config.yaml << EOF
# Enterprise Security Configuration
security:
  # Global security mode
  mode: "enterprise"
  hardening_level: "strict"
  enforce_https: true

  # TLS Configuration
  tls:
    enabled: true
    min_version: "1.3"
    cert_file: "/etc/ssl/certs/server.crt"
    key_file: "/etc/ssl/private/server.key"
    ca_file: "/etc/ssl/certs/ca.crt"
    mutual_tls: true
    ocsp_stapling: true
    cipher_suites:
      - "TLS_AES_256_GCM_SHA384"
      - "TLS_CHACHA20_POLY1305_SHA256"

  # Multi-factor Authentication
  authentication:
    enabled: true
    methods: ["jwt", "oauth2", "mfa"]
    require_mfa: true
    session_timeout: "8h"
    max_concurrent_sessions: 3
    
    jwt:
      secret_key_file: "/etc/secrets/jwt-secret"
      algorithm: "RS256"
      expiry: "1h"
      refresh_enabled: true
    
    mfa:
      enabled: true
      required: true
      methods: ["totp", "backup_codes"]
      grace_period: "24h"

  # Authorization & RBAC
  authorization:
    enabled: true
    rbac: true
    default_role: "user"
    roles:
      - name: "admin"
        permissions: ["system:admin", "users:*", "models:*"]
      - name: "operator"  
        permissions: ["models:read", "models:manage", "inference:execute"]
      - name: "user"
        permissions: ["models:read", "inference:execute"]

  # Data encryption
  encryption:
    enabled: true
    at_rest: true
    in_transit: true
    algorithm: "AES-256-GCM"
    key_rotation: true
    key_rotation_interval: "30d"

  # Rate limiting and DDoS protection
  rate_limiting:
    enabled: true
    global:
      limit: 10000
      burst: 100
    user:
      limit: 1000
      burst: 50
    ip:
      limit: 100
      burst: 20

  # Web Application Firewall
  waf:
    enabled: true
    mode: "prevention"
    owasp_crs: true
    max_request_size: 10485760
    custom_rules:
      - id: "block_admin_paths"
        pattern: "/(admin|wp-admin)"
        action: "block"

  # Audit logging
  audit:
    enabled: true
    level: "INFO"
    format: "json"
    file: "/var/log/ollama/audit.log"
    events: ["authentication", "authorization", "data_access"]

# Network security
network:
  firewall:
    enabled: true
    default_policy: "deny"
    rules:
      - name: "allow_https"
        protocol: "tcp"
        port: 443
        action: "allow"
      - name: "allow_api"
        protocol: "tcp"
        port: 8443
        action: "allow"
EOF

echo "Enterprise security configuration created."
```

**✅ Checkpoint 6**: Comprehensive security configuration with enterprise-grade settings.

## 🧪 Hands-On Exercise 1: Environment Variable Management

### Step 7: Environment-Specific Variable Setup

```bash
# Create environment-specific variable files
mkdir -p ./envs

# Development environment variables
cat > ./envs/.env.development << EOF
# Development Environment Variables
OLLAMA_NODE_ID=dev-node-$(hostname)
OLLAMA_NODE_NAME=Development-Node
OLLAMA_NODE_REGION=local
OLLAMA_NODE_ZONE=dev
OLLAMA_CLUSTER_NAME=dev-cluster
OLLAMA_VERSION=1.0.0-dev

# Database (development uses local storage)
OLLAMA_DB_TYPE=local
OLLAMA_DB_PATH=./data/dev.db

# Security (relaxed for development)
OLLAMA_JWT_SECRET=dev-secret-change-in-production
OLLAMA_TLS_ENABLED=false

# Performance (limited for development)
OLLAMA_MAX_WORKERS=4
OLLAMA_CACHE_SIZE=1GB

# Logging
OLLAMA_LOG_LEVEL=debug
OLLAMA_LOG_FORMAT=text

# CORS (permissive for development)
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://127.0.0.1:3000
EOF

# Production environment variables
cat > ./envs/.env.production << EOF
# Production Environment Variables  
OLLAMA_NODE_ID=\${EC2_INSTANCE_ID}
OLLAMA_NODE_NAME=Prod-Node-\${EC2_INSTANCE_ID}
OLLAMA_NODE_REGION=\${AWS_REGION}
OLLAMA_NODE_ZONE=\${AWS_AZ}
OLLAMA_CLUSTER_NAME=prod-cluster
OLLAMA_VERSION=1.0.0

# Database (production uses PostgreSQL)
OLLAMA_DB_TYPE=postgresql
OLLAMA_DB_HOST=\${RDS_ENDPOINT}
OLLAMA_DB_NAME=\${RDS_DB_NAME}
OLLAMA_DB_USER=\${RDS_USERNAME}
OLLAMA_DB_PASSWORD=\${RDS_PASSWORD}

# Security (strict for production)
OLLAMA_JWT_SECRET=\${JWT_SECRET_FROM_SECRETS_MANAGER}
OLLAMA_TLS_ENABLED=true
OLLAMA_TLS_CERT_PATH=/etc/ssl/certs/ollama.crt
OLLAMA_TLS_KEY_PATH=/etc/ssl/private/ollama.key

# Performance (optimized for production)
OLLAMA_MAX_WORKERS=16
OLLAMA_CACHE_SIZE=50GB

# Logging (structured for production)
OLLAMA_LOG_LEVEL=info
OLLAMA_LOG_FORMAT=json

# CORS (restrictive for production)
CORS_ALLOWED_ORIGINS=https://app.company.com,https://admin.company.com

# Monitoring
PROMETHEUS_ENDPOINT=http://prometheus.monitoring.svc.cluster.local:9090
GRAFANA_ENDPOINT=http://grafana.monitoring.svc.cluster.local:3000

# Bootstrap peers for clustering
OLLAMA_BOOTSTRAP_PEERS=node1.cluster.local:9000,node2.cluster.local:9000,node3.cluster.local:9000
EOF

# Create environment loader script
cat > load-env.sh << 'EOF'
#!/bin/bash
# Environment loader script

ENVIRONMENT=${1:-development}
ENV_FILE="./envs/.env.${ENVIRONMENT}"

if [ ! -f "$ENV_FILE" ]; then
    echo "Environment file not found: $ENV_FILE"
    echo "Available environments:"
    ls -1 ./envs/.env.* | sed 's/.*\.env\./  - /'
    exit 1
fi

echo "Loading environment: $ENVIRONMENT"
set -a  # Automatically export variables
source "$ENV_FILE"
set +a

echo "Environment variables loaded:"
env | grep "OLLAMA_" | sort
EOF

chmod +x load-env.sh

echo "Environment management setup complete."
```

**✅ Checkpoint 7**: Environment-specific variable management system created.

## ⚡ Performance Tuning Configuration

### Step 8: Performance Optimization Settings

```bash
# Create performance tuning configuration
cat > performance-config.yaml << EOF
# Performance Tuning Configuration
performance:
  # CPU optimization
  cpu:
    max_threads: 0  # 0 = auto-detect
    thread_affinity: true
    numa_aware: true
    cpu_governor: "performance"
  
  # Memory optimization  
  memory:
    heap_size: "8GB"
    gc_target_percentage: 75
    memory_mapping: true
    huge_pages: false  # Enable if system supports
    swap_usage: "minimal"
  
  # I/O optimization
  io:
    async_io: true
    io_uring: false  # Enable on Linux 5.1+
    buffer_size: 65536
    read_ahead: 2048
    sync_method: "fdatasync"
  
  # Network optimization
  network:
    tcp_nodelay: true
    tcp_cork: false
    send_buffer_size: 262144
    receive_buffer_size: 262144
    keepalive: true
    keepalive_idle: 600
  
  # Cache optimization
  cache:
    strategy: "adaptive"
    model_cache_size: "20GB"
    metadata_cache_size: "1GB"
    cache_compression: true
    prefetch_models: true
  
  # Workload-specific optimizations
  workloads:
    inference:
      batch_size: 32
      max_sequence_length: 2048
      parallel_requests: 8
      timeout: "30s"
    
    training:
      gradient_accumulation_steps: 4
      mixed_precision: true
      checkpoint_interval: "10m"
      
    batch_processing:
      queue_size: 1000
      worker_pool_size: 16
      processing_timeout: "5m"

# Resource monitoring and limits
resources:
  limits:
    max_memory_usage: "90%"
    max_cpu_usage: "95%"
    max_disk_usage: "85%"
    max_network_bandwidth: "1Gbps"
  
  monitoring:
    interval: "10s"
    metrics_retention: "7d"
    alert_thresholds:
      memory: "85%"
      cpu: "90%"
      disk: "80%"
EOF

echo "Performance tuning configuration created."
```

**✅ Checkpoint 8**: Performance optimization settings configured for different workloads.

## 🌐 Multi-Node Configuration

### Step 9: Multi-Node Cluster Setup

```bash
# Create multi-node cluster configurations
mkdir -p cluster-configs

# Node 1 (Bootstrap/Leader)
cat > cluster-configs/node1-config.yaml << EOF
# Node 1 - Bootstrap/Leader Configuration
node:
  id: "cluster-node-1"
  name: "Primary Node"
  region: "us-west-2"
  zone: "us-west-2a"
  environment: "production"
  role: "leader"
  tags:
    cluster_role: "bootstrap"
    priority: "high"

api:
  listen: "0.0.0.0:8080"
  advertise_addr: "10.0.1.10:8080"

p2p:
  listen: "0.0.0.0:9000"
  advertise_addr: "10.0.1.10:9000"
  bootstrap_peers: []  # Empty for bootstrap node

consensus:
  enabled: true
  data_dir: "/var/lib/ollama/consensus"
  bind_addr: "10.0.1.10:7000"
  bootstrap: true  # This is the bootstrap node
  voter: true

scheduler:
  enabled: true
  algorithm: "weighted_round_robin"
  weights:
    cpu: 0.3
    memory: 0.3
    network: 0.2
    load: 0.2

storage:
  replication:
    enabled: true
    factor: 3
    strategy: "geographic"
EOF

# Node 2 (Follower)
cat > cluster-configs/node2-config.yaml << EOF
# Node 2 - Follower Configuration
node:
  id: "cluster-node-2"
  name: "Secondary Node"
  region: "us-west-2"
  zone: "us-west-2b"
  environment: "production"
  role: "follower"
  tags:
    cluster_role: "worker"
    priority: "medium"

api:
  listen: "0.0.0.0:8080"
  advertise_addr: "10.0.1.11:8080"

p2p:
  listen: "0.0.0.0:9000"
  advertise_addr: "10.0.1.11:9000"
  bootstrap_peers: ["10.0.1.10:9000"]  # Connect to bootstrap node

consensus:
  enabled: true
  data_dir: "/var/lib/ollama/consensus"
  bind_addr: "10.0.1.11:7000"
  bootstrap: false
  voter: true

scheduler:
  enabled: true
  algorithm: "load_balancing"
  health_check_interval: "30s"

storage:
  replication:
    enabled: true
    factor: 3
EOF

# Node 3 (Follower)  
cat > cluster-configs/node3-config.yaml << EOF
# Node 3 - Follower Configuration
node:
  id: "cluster-node-3"
  name: "Tertiary Node"
  region: "us-west-2"
  zone: "us-west-2c"
  environment: "production"
  role: "follower"
  tags:
    cluster_role: "worker"
    priority: "medium"

api:
  listen: "0.0.0.0:8080"
  advertise_addr: "10.0.1.12:8080"

p2p:
  listen: "0.0.0.0:9000"
  advertise_addr: "10.0.1.12:9000"
  bootstrap_peers: ["10.0.1.10:9000", "10.0.1.11:9000"]

consensus:
  enabled: true
  data_dir: "/var/lib/ollama/consensus"
  bind_addr: "10.0.1.12:7000"
  bootstrap: false
  voter: true

scheduler:
  enabled: true
  algorithm: "resource_aware"
  max_concurrent_tasks: 50

storage:
  replication:
    enabled: true
    factor: 3
EOF

# Create cluster deployment script
cat > deploy-cluster.sh << 'EOF'
#!/bin/bash
# Multi-node cluster deployment script

set -e

NODES=("node1" "node2" "node3")
IPS=("10.0.1.10" "10.0.1.11" "10.0.1.12")

echo "Deploying Ollama Distributed Cluster..."

for i in "${!NODES[@]}"; do
    NODE=${NODES[$i]}
    IP=${IPS[$i]}
    
    echo "Deploying $NODE ($IP)..."
    
    # Copy configuration to remote node (simulation)
    echo "  ✓ Configuration copied to $IP"
    
    # Start node (simulation)
    echo "  ✓ Node $NODE started"
    
    # Wait for node to be ready
    sleep 2
    echo "  ✓ Node $NODE is healthy"
done

echo "Cluster deployment complete!"
echo "Cluster status:"
echo "  Bootstrap Node: ${IPS[0]}:8080"
echo "  Worker Nodes: ${IPS[1]}:8080, ${IPS[2]}:8080"
EOF

chmod +x deploy-cluster.sh

echo "Multi-node cluster configuration created."
```

**✅ Checkpoint 9**: Multi-node cluster configurations created with proper bootstrap, consensus, and replication settings.

## 📊 Monitoring and Logging Configuration

### Step 10: Comprehensive Monitoring Setup

```bash
# Create monitoring configuration
cat > monitoring-config.yaml << EOF
# Comprehensive Monitoring Configuration
monitoring:
  # Prometheus configuration
  prometheus:
    enabled: true
    endpoint: "http://prometheus:9090"
    scrape_interval: "15s"
    retention: "15d"
    
    # Custom metrics
    custom_metrics:
      - name: "ollama_inference_duration"
        type: "histogram"
        help: "Time spent on inference requests"
        buckets: [0.1, 0.5, 1, 2, 5, 10, 30]
      
      - name: "ollama_model_cache_hit_ratio"
        type: "gauge"
        help: "Model cache hit ratio percentage"
      
      - name: "ollama_active_connections"
        type: "gauge"
        help: "Number of active client connections"

  # Grafana dashboards
  grafana:
    enabled: true
    endpoint: "http://grafana:3000"
    dashboards:
      - name: "ollama-overview"
        path: "/etc/grafana/dashboards/overview.json"
      - name: "ollama-performance"
        path: "/etc/grafana/dashboards/performance.json"
      - name: "ollama-security"
        path: "/etc/grafana/dashboards/security.json"

  # Health checks
  health:
    enabled: true
    interval: "30s"
    timeout: "10s"
    endpoints:
      - path: "/health"
        expected_status: 200
      - path: "/metrics"
        expected_status: 200
      - path: "/api/v1/status"
        expected_status: 200

  # Alerting
  alerting:
    enabled: true
    providers:
      - type: "webhook"
        url: "http://alertmanager:9093/api/v1/alerts"
      - type: "email"
        smtp_server: "smtp.company.com"
        recipients: ["ops@company.com"]
    
    rules:
      - name: "high_cpu_usage"
        condition: "cpu_usage > 90"
        duration: "5m"
        severity: "warning"
      
      - name: "memory_pressure"
        condition: "memory_usage > 85"
        duration: "2m"
        severity: "critical"
      
      - name: "inference_latency_high"
        condition: "inference_p95_latency > 10s"
        duration: "1m"
        severity: "warning"

# Structured logging configuration
logging:
  # Log levels and outputs
  levels:
    root: "info"
    api: "debug"
    p2p: "info"
    consensus: "warn"
    scheduler: "info"
    security: "info"

  # Log outputs
  outputs:
    console:
      enabled: true
      format: "text"
      level: "info"
    
    file:
      enabled: true
      path: "/var/log/ollama/app.log"
      format: "json"
      level: "debug"
      rotation:
        max_size: 100  # MB
        max_age: 30    # days
        max_backups: 10
        compress: true
    
    syslog:
      enabled: false
      network: "udp"
      address: "localhost:514"
      facility: "local0"
    
    elasticsearch:
      enabled: false
      endpoint: "http://elasticsearch:9200"
      index: "ollama-logs"
      bulk_size: 1000

  # Log enrichment
  enrichment:
    add_caller: true
    add_hostname: true
    add_timestamp: true
    add_request_id: true
    add_user_id: true

# Distributed tracing
tracing:
  enabled: true
  provider: "jaeger"
  
  jaeger:
    endpoint: "http://jaeger:14268/api/traces"
    service_name: "ollama-distributed"
    sample_rate: 0.1  # 10% sampling
  
  # Trace sampling
  sampling:
    strategy: "adaptive"
    max_traces_per_second: 100
    
  # Instrumentation
  instrumentation:
    http_requests: true
    database_queries: true
    cache_operations: true
    p2p_operations: true

# Performance profiling
profiling:
  enabled: true
  pprof_enabled: true
  pprof_port: 6060
  
  # CPU profiling
  cpu_profile:
    enabled: false  # Enable when needed
    duration: "30s"
    output: "/tmp/cpu.prof"
  
  # Memory profiling
  memory_profile:
    enabled: false  # Enable when needed
    output: "/tmp/mem.prof"
  
  # Goroutine profiling
  goroutine_profile:
    enabled: false
    output: "/tmp/goroutine.prof"
EOF

echo "Comprehensive monitoring configuration created."
```

**✅ Checkpoint 10**: Full monitoring, logging, and tracing configuration established.

## 🧪 Hands-On Exercise 2: Configuration Validation and Testing

### Step 11: Advanced Configuration Validation

```bash
# Create comprehensive validation script
cat > validate-advanced-configs.sh << 'EOF'
#!/bin/bash
# Advanced Configuration Validation Script

set -e

echo "🔍 Advanced Configuration Validation"
echo "===================================="

# Array of configuration files to validate
configs=(
    "dev-advanced-config.yaml:Development"
    "prod-advanced-config.yaml:Production"
    "edge-config.yaml:Edge/IoT"
    "gpu-config.yaml:GPU-Optimized"
    "security-advanced-config.yaml:Security"
    "performance-config.yaml:Performance"
    "monitoring-config.yaml:Monitoring"
    "cluster-configs/node1-config.yaml:Cluster Node 1"
    "cluster-configs/node2-config.yaml:Cluster Node 2"
    "cluster-configs/node3-config.yaml:Cluster Node 3"
)

validate_yaml() {
    local file=$1
    local description=$2
    
    echo -n "Validating $description ($file)... "
    
    # Check if file exists
    if [ ! -f "$file" ]; then
        echo "❌ File not found"
        return 1
    fi
    
    # Validate YAML syntax
    if python3 -c "import yaml; yaml.safe_load(open('$file'))" 2>/dev/null; then
        echo "✅ YAML syntax valid"
    else
        echo "❌ Invalid YAML syntax"
        return 1
    fi
    
    # Check for required sections (basic validation)
    local required_sections=("node" "api")
    for section in "${required_sections[@]}"; do
        if grep -q "^${section}:" "$file"; then
            echo "  ✅ Required section '$section' present"
        else
            echo "  ⚠️  Required section '$section' missing"
        fi
    done
    
    return 0
}

validate_environment_vars() {
    local env_file=$1
    local description=$2
    
    echo -n "Validating $description ($env_file)... "
    
    if [ ! -f "$env_file" ]; then
        echo "❌ File not found"
        return 1
    fi
    
    # Check for required variables
    local required_vars=("OLLAMA_NODE_ID" "OLLAMA_NODE_NAME")
    local missing_vars=()
    
    for var in "${required_vars[@]}"; do
        if ! grep -q "^${var}=" "$env_file"; then
            missing_vars+=("$var")
        fi
    done
    
    if [ ${#missing_vars[@]} -eq 0 ]; then
        echo "✅ All required variables present"
    else
        echo "❌ Missing variables: ${missing_vars[*]}"
        return 1
    fi
    
    return 0
}

# Validate configuration files
echo "📁 Validating Configuration Files:"
echo "-----------------------------------"

failed_validations=0

for config in "${configs[@]}"; do
    IFS=':' read -r file description <<< "$config"
    if ! validate_yaml "$file" "$description"; then
        ((failed_validations++))
    fi
    echo
done

# Validate environment files
echo "🌍 Validating Environment Files:"
echo "--------------------------------"

env_files=(
    "envs/.env.development:Development Environment"
    "envs/.env.production:Production Environment"
)

for env in "${env_files[@]}"; do
    IFS=':' read -r file description <<< "$env"
    if ! validate_environment_vars "$file" "$description"; then
        ((failed_validations++))
    fi
    echo
done

# Security validation
echo "🔒 Security Validation:"
echo "----------------------"

security_checks=(
    "TLS enabled in production:prod-advanced-config.yaml:tls.*enabled.*true"
    "Authentication enabled:security-advanced-config.yaml:authentication.*enabled.*true"
    "Rate limiting enabled:security-advanced-config.yaml:rate_limiting.*enabled.*true"
)

for check in "${security_checks[@]}"; do
    IFS=':' read -r description file pattern <<< "$check"
    echo -n "Checking $description... "
    
    if [ -f "$file" ] && grep -q "$pattern" "$file"; then
        echo "✅ Pass"
    else
        echo "❌ Fail"
        ((failed_validations++))
    fi
done

echo
echo "📊 Validation Summary:"
echo "====================="

if [ $failed_validations -eq 0 ]; then
    echo "🎉 All validations passed!"
    echo "Your configurations are ready for deployment."
else
    echo "❌ $failed_validations validation(s) failed."
    echo "Please review and fix the issues above."
    exit 1
fi
EOF

chmod +x validate-advanced-configs.sh

# Run the validation
./validate-advanced-configs.sh
```

**✅ Checkpoint 11**: Configuration validation script created and executed successfully.

## 🔧 Troubleshooting Advanced Configuration Issues

### Step 12: Common Advanced Issues and Solutions

```bash
# Create troubleshooting guide and diagnostic script
cat > troubleshoot-config.sh << 'EOF'
#!/bin/bash
# Configuration Troubleshooting Script

echo "🔧 Configuration Troubleshooting Diagnostics"
echo "============================================"

check_port_conflicts() {
    echo "🌐 Checking Port Conflicts:"
    echo "--------------------------"
    
    ports=(8080 8081 9000 9090 7000)
    
    for port in "${ports[@]}"; do
        echo -n "Port $port: "
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            echo "❌ In use"
            lsof -Pi :$port -sTCP:LISTEN
        else
            echo "✅ Available"
        fi
    done
    echo
}

check_file_permissions() {
    echo "📁 Checking File Permissions:"
    echo "-----------------------------"
    
    files=(
        "/etc/ssl/certs:755:Certificate directory"
        "/etc/ssl/private:700:Private key directory" 
        "/var/log/ollama:755:Log directory"
        "/var/lib/ollama:755:Data directory"
    )
    
    for file_info in "${files[@]}"; do
        IFS=':' read -r path expected_perm description <<< "$file_info"
        echo -n "$description ($path): "
        
        if [ -e "$path" ]; then
            actual_perm=$(stat -c "%a" "$path")
            if [ "$actual_perm" = "$expected_perm" ]; then
                echo "✅ Correct ($actual_perm)"
            else
                echo "❌ Incorrect ($actual_perm, expected $expected_perm)"
                echo "  Fix: chmod $expected_perm $path"
            fi
        else
            echo "❌ Does not exist"
            echo "  Fix: mkdir -p $path && chmod $expected_perm $path"
        fi
    done
    echo
}

check_environment_variables() {
    echo "🌍 Environment Variable Check:"
    echo "-----------------------------"
    
    required_vars=(
        "OLLAMA_NODE_ID:Node identifier"
        "OLLAMA_NODE_NAME:Node name"
        "OLLAMA_JWT_SECRET:JWT secret key"
    )
    
    for var_info in "${required_vars[@]}"; do
        IFS=':' read -r var_name description <<< "$var_info"
        echo -n "$description ($var_name): "
        
        if [ -n "${!var_name}" ]; then
            echo "✅ Set"
        else
            echo "❌ Not set"
            echo "  Fix: export $var_name=<value>"
        fi
    done
    echo
}

check_resource_availability() {
    echo "💾 Resource Availability:"
    echo "------------------------"
    
    # Memory check
    echo -n "Available Memory: "
    mem_available=$(awk '/MemAvailable:/ {print $2}' /proc/meminfo)
    mem_gb=$((mem_available / 1024 / 1024))
    if [ $mem_gb -ge 4 ]; then
        echo "✅ ${mem_gb}GB available"
    else
        echo "⚠️  ${mem_gb}GB available (recommend 4GB+)"
    fi
    
    # Disk space check
    echo -n "Disk Space: "
    disk_available=$(df -BG . | tail -1 | awk '{print $4}' | sed 's/G//')
    if [ $disk_available -ge 10 ]; then
        echo "✅ ${disk_available}GB available"
    else
        echo "⚠️  ${disk_available}GB available (recommend 10GB+)"
    fi
    
    # CPU cores
    echo -n "CPU Cores: "
    cpu_cores=$(nproc)
    if [ $cpu_cores -ge 2 ]; then
        echo "✅ $cpu_cores cores"
    else
        echo "⚠️  $cpu_cores core (recommend 2+)"
    fi
    echo
}

check_network_connectivity() {
    echo "🌐 Network Connectivity:"
    echo "----------------------"
    
    hosts=(
        "127.0.0.1:Local loopback"
        "google.com:Internet connectivity"
    )
    
    for host_info in "${hosts[@]}"; do
        IFS=':' read -r host description <<< "$host_info"
        echo -n "$description ($host): "
        
        if ping -c 1 -W 2 "$host" >/dev/null 2>&1; then
            echo "✅ Reachable"
        else
            echo "❌ Unreachable"
        fi
    done
    echo
}

# Run all diagnostic checks
check_port_conflicts
check_file_permissions  
check_environment_variables
check_resource_availability
check_network_connectivity

echo "🎯 Troubleshooting Complete!"
echo "Review any issues above and apply the suggested fixes."
EOF

chmod +x troubleshoot-config.sh

echo "Troubleshooting script created. Run './troubleshoot-config.sh' to diagnose issues."
```

**✅ Checkpoint 12**: Comprehensive troubleshooting tools created for advanced configuration issues.

## 📊 Module 6 Assessment

### Knowledge Check ✋

1. **Q**: What are the four main configuration profiles discussed in this module?
   **A**: Development, Production, Edge/IoT, and GPU-Optimized

2. **Q**: What TLS version should be used in production environments?
   **A**: TLS 1.3 minimum for modern security standards

3. **Q**: What are the three levels of data encryption in enterprise security?
   **A**: At-rest, in-transit, and key rotation encryption

4. **Q**: What consensus role does the bootstrap node play in a cluster?
   **A**: It serves as the initial leader and voter in the Raft consensus protocol

5. **Q**: What are the key components of the monitoring stack?
   **A**: Prometheus (metrics), Grafana (dashboards), and Jaeger (tracing)

### Practical Check ✋

Verify you can complete these advanced tasks:

- [ ] Create environment-specific configurations with proper variable management
- [ ] Configure enterprise-grade security with TLS, authentication, and encryption
- [ ] Set up performance-optimized configurations for different workloads
- [ ] Design multi-node cluster configurations with proper consensus settings
- [ ] Implement comprehensive monitoring and logging configurations
- [ ] Validate configurations using automated scripts
- [ ] Troubleshoot common configuration issues

### Advanced Challenge 🚀

Try implementing a complete production deployment:

```bash
# Create a production-ready deployment configuration
cat > production-deployment.yaml << EOF
# Complete Production Deployment Configuration
environments:
  production:
    security_profile: "enterprise"
    performance_profile: "high"
    monitoring_profile: "comprehensive"
    
    nodes:
      - id: "prod-node-1"
        role: "leader"
        region: "us-west-2a"
        instance_type: "c5.2xlarge"
        
      - id: "prod-node-2" 
        role: "follower"
        region: "us-west-2b"
        instance_type: "c5.2xlarge"
        
      - id: "prod-node-3"
        role: "follower"
        region: "us-west-2c"
        instance_type: "c5.2xlarge"

    load_balancer:
      type: "application"
      ssl_termination: true
      health_check: "/health"
      
    database:
      type: "postgresql"
      multi_az: true
      encryption: true
      backup_retention: 30

    monitoring:
      prometheus: true
      grafana: true
      alertmanager: true
      jaeger: true
EOF
```

## 🎉 Module 6 Complete!

**Congratulations!** You have successfully mastered:

✅ **Configuration Profiles** - Development, production, edge, and GPU environments  
✅ **Environment Management** - Variable management and environment-specific settings  
✅ **Security Configuration** - Enterprise-grade TLS, authentication, and encryption  
✅ **Performance Tuning** - Optimization for different workloads and resources  
✅ **Multi-Node Setup** - Cluster configuration with consensus and replication  
✅ **Monitoring & Logging** - Comprehensive observability and alerting systems  
✅ **Validation & Troubleshooting** - Automated testing and issue resolution  

### Key Takeaways

1. **Environment-Specific Configs**: Different environments need different security, performance, and resource settings
2. **Security Layers**: Implement defense in depth with TLS, authentication, authorization, and encryption
3. **Performance Optimization**: Tune configurations based on workload characteristics and resource availability
4. **Cluster Coordination**: Proper consensus, replication, and bootstrap configurations are critical for reliability
5. **Observability**: Comprehensive monitoring, logging, and tracing enable effective operations
6. **Validation First**: Always validate configurations before deployment to prevent issues

## 📚 What's Next?

You've completed the advanced configuration training! For continued learning:

- **Production Deployment**: Apply these configurations in real production environments
- **Advanced Monitoring**: Set up custom dashboards and alerting rules
- **Security Hardening**: Implement additional security measures like WAF and DLP
- **Performance Optimization**: Fine-tune settings based on actual workload metrics
- **Automation**: Create Infrastructure as Code (IaC) templates for your configurations

## 💡 Pro Tips for Configuration Management

1. **Version Control**: Keep all configurations in version control with proper branching
2. **Configuration as Code**: Use templating engines like Helm or Jsonnet for complex deployments
3. **Secret Management**: Never store secrets in configuration files - use secret management systems
4. **Gradual Rollouts**: Test configuration changes in staging before production deployment
5. **Monitoring Changes**: Monitor the impact of configuration changes on system behavior
6. **Documentation**: Document all configuration decisions and their rationale
7. **Regular Audits**: Periodically review and update configurations for security and performance

---

**Module 6 Status**: ✅ Complete  
**Training Program**: 🎯 Mastered Advanced Configuration Management  
**Total Progress**: 6/6 modules (100%)

**🎓 Congratulations! You've completed the OllamaMax Distributed Training Program!**