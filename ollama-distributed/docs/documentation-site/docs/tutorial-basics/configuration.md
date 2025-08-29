# Configuration Guide

Learn how to configure Ollama Distributed based on the actual implementation and configuration structure.

## Configuration Overview

Ollama Distributed uses a comprehensive configuration system that supports:

- **YAML configuration files** for structured settings
- **Environment variables** for deployment flexibility  
- **Command-line flags** for runtime overrides
- **Interactive wizard** for guided setup

### Configuration Priority

Settings are applied in this order (highest to lowest priority):

1. Command-line flags
2. Environment variables  
3. Configuration files (`~/.ollamamax/config.yaml`)
4. Default values

## Quick Configuration

### Interactive Setup Wizard

The easiest way to configure Ollama Distributed:

```bash
# Launch interactive configuration wizard
ollama-distributed setup

# Follow the prompts to configure:
# - Node name (default: ollama-node)
# - API port (default: 8080)  
# - Web port (default: 8081)
# - GPU support (y/N)
```

### QuickStart Configuration

Generate a basic configuration automatically:

```bash
# Generate default configuration with quickstart
ollama-distributed quickstart

# This creates ~/.ollamamax/quickstart-config.yaml
```

### Configuration Scripts

Use the provided configuration generator:

```bash
# Generate development configuration
./scripts/config-generator.sh --profile development

# Generate production configuration  
./scripts/config-generator.sh --profile production --security

# Generate GPU-optimized configuration
./scripts/config-generator.sh --profile gpu

# Generate cluster configuration
./scripts/config-generator.sh --profile cluster --nodes 3
```

## Configuration File Structure

The main configuration file is located at `~/.ollamamax/config.yaml`:

```yaml
# Node Configuration
node:
  id: "quickstart-node"           # Unique node identifier
  name: "quickstart-node"         # Human-readable name  
  data_dir: "~/.ollamamax/data"   # Data directory path
  log_level: "info"              # Log level (debug, info, warn, error)
  environment: "development"      # Environment (development, production)

# API Server Configuration
api:
  host: "0.0.0.0"               # Bind address
  port: 8080                    # HTTP/API port
  enable_tls: false             # Enable TLS encryption
  cert_file: ""                 # TLS certificate file (if TLS enabled)
  key_file: ""                  # TLS private key file (if TLS enabled)  
  max_request_size: 104857600   # Max request size in bytes (100MB)
  timeout:
    read: "30s"                 # Read timeout
    write: "30s"                # Write timeout
    idle: "120s"                # Idle timeout

# Web Interface Configuration  
web:
  enabled: true                 # Enable web dashboard
  host: "0.0.0.0"              # Web interface bind address
  port: 8081                    # Web interface port
  enable_tls: false             # Enable TLS for web interface

# P2P Network Configuration
p2p:
  enabled: true                 # Enable P2P networking
  listen_port: 4001             # P2P listen port
  bootstrap_peers: []           # Bootstrap peer addresses
  dial_timeout: "10s"           # Peer connection timeout
  max_connections: 100          # Maximum P2P connections

# Model Management Configuration
models:
  store_path: "~/.ollamamax/data/models"  # Model storage directory
  max_cache_size: "10GB"        # Maximum cache size
  auto_cleanup: true            # Automatically cleanup unused models
  download_timeout: "30m"       # Model download timeout
  concurrency: 3                # Concurrent downloads allowed
  isolated_storage: false       # Use isolated storage per node
  migration_source: ""          # Source for model migration

# Performance Configuration
performance:
  max_concurrency: 4            # Max concurrent inference requests
  memory_limit: "8GB"           # Memory limit for processes
  gpu_enabled: false            # Enable GPU acceleration (auto-detected)
  worker_pool_size: 4           # Worker pool size
  queue_size: 100               # Request queue size
  gc_percent: 100               # Go garbage collection percentage
  cooperative_mode: false       # Enable cooperative scheduling

# Authentication Configuration
auth:
  enabled: false                # Enable authentication
  method: "jwt"                 # Authentication method
  jwt_secret: ""                # JWT secret key
  token_expiry: "24h"          # Token expiration time
  admin_users: []               # List of admin users

# Security Configuration  
security:
  enable_cors: true             # Enable CORS
  cors_origins: ["*"]           # Allowed CORS origins
  rate_limiting:
    enabled: false              # Enable rate limiting
    requests_per_minute: 60     # Requests per minute limit
  encryption:
    at_rest: false              # Enable encryption at rest
    in_transit: true            # Enable encryption in transit

# Database Configuration
database:
  driver: "sqlite"              # Database driver (sqlite, postgres, mysql)
  connection_string: "./data/ollama.db"  # Database connection string
  max_connections: 25           # Maximum database connections
  max_idle_connections: 5       # Maximum idle connections
  connection_lifetime: "1h"     # Connection lifetime

# Consensus Configuration (for clustering)
consensus:
  enabled: false                # Enable Raft consensus
  node_id: ""                   # Consensus node ID
  raft_dir: "./data/raft"       # Raft data directory
  bind_addr: "127.0.0.1:7000"  # Raft bind address
  bootstrap: false              # Bootstrap new cluster
  join_addr: ""                 # Address to join existing cluster

# Monitoring Configuration
monitoring:
  enabled: true                 # Enable monitoring
  prometheus:
    enabled: true               # Enable Prometheus metrics
    bind_addr: "127.0.0.1:9090" # Prometheus bind address
    path: "/metrics"            # Metrics endpoint path
  health_checks:
    enabled: true               # Enable health checks
    interval: "30s"             # Health check interval
    timeout: "5s"               # Health check timeout

# Logging Configuration
logging:
  level: "info"                 # Log level
  format: "json"                # Log format (json, text)
  output: "stdout"              # Output destination (stdout, stderr, file)
  file: "./logs/ollama.log"     # Log file path (if output=file)
  max_size: "100MB"            # Max log file size
  max_backups: 3                # Max backup files to keep
  max_age: "30d"               # Max age of log files

# Resource Limits
resource_limits:
  cpu_limit: "0"                # CPU limit (0 = unlimited)
  memory_limit: "0"             # Memory limit (0 = unlimited)  
  disk_limit: "0"               # Disk usage limit (0 = unlimited)
  network_limit: "0"            # Network bandwidth limit (0 = unlimited)

# Optional Migration Configuration
migration:
  enabled: false                # Enable migration from existing Ollama
  source_path: ""               # Source Ollama data path
  preserve_models: true         # Preserve existing models
  backup_original: true         # Backup original data

# Optional Coexistence Configuration  
coexistence:
  enabled: false                # Enable coexistence with existing Ollama
  port_offset: 1000             # Port offset for coexistence
  data_isolation: true          # Isolate data from original Ollama
```

## Configuration Profiles

### Development Profile

Optimized for local development:

```bash
# Generate development configuration
ollama-distributed config generate --profile dev

# Key settings:
# - Single node setup
# - Debug logging enabled
# - Development-friendly ports
# - Local file storage
# - No authentication
```

Example development config:
```yaml
node:
  id: "dev-node"
  
server:
  port: 8081
  
cluster:
  bootstrap: true
  
logging:
  level: "debug"
  
security:
  authentication:
    enabled: false
```

### Production Profile

Enterprise-ready production configuration:

```bash
# Generate production configuration
ollama-distributed config generate --profile production

# Key settings:
# - Multi-node cluster ready
# - TLS encryption enabled
# - Authentication required
# - Performance optimized
# - Comprehensive monitoring
```

Example production config:
```yaml
server:
  tls:
    enabled: true
    cert_file: "/etc/ssl/ollama/server.crt"
    key_file: "/etc/ssl/ollama/server.key"

cluster:
  bootstrap_expect: 3
  encryption_key: "${CLUSTER_ENCRYPTION_KEY}"

security:
  authentication:
    enabled: true
    method: "jwt"
    secret: "${JWT_SECRET}"

monitoring:
  enabled: true
  prometheus:
    enabled: true

logging:
  level: "info"
  output: "file"
  file: "/var/log/ollama/distributed.log"
```

### GPU Profile

Optimized for GPU-accelerated inference:

```bash
# Generate GPU configuration
ollama-distributed config generate --profile gpu

# Key settings:
# - GPU acceleration enabled
# - Optimized memory allocation
# - CUDA/ROCm support
# - Higher concurrent limits
```

Example GPU config:
```yaml
inference:
  max_concurrent: 20
  gpu:
    enabled: true
    devices: [0, 1, 2, 3]
    memory_fraction: 0.9
  cpu:
    optimization: "balanced"

models:
  cache_size: "20GB"
```

### Cluster Profile

Multi-node cluster configuration:

```bash
# Generate cluster configuration for 3 nodes
ollama-distributed config generate --profile cluster --nodes 3

# Generates configs for:
# - node-1 (bootstrap)
# - node-2 (follower)
# - node-3 (follower)
```

## Environment Variables

All configuration options can be set via environment variables using the `OLLAMA_` prefix:

```bash
# Server configuration
export OLLAMA_SERVER_HOST="0.0.0.0"
export OLLAMA_SERVER_PORT="8081"

# Cluster configuration
export OLLAMA_CLUSTER_NODE_ID="production-node-1"
export OLLAMA_CLUSTER_BOOTSTRAP="true"
export OLLAMA_CLUSTER_ENCRYPTION_KEY="your-32-byte-key-here"

# Model configuration
export OLLAMA_MODELS_STORE_PATH="/data/models"
export OLLAMA_MODELS_CACHE_SIZE="50GB"

# Inference configuration
export OLLAMA_INFERENCE_MAX_CONCURRENT="20"
export OLLAMA_INFERENCE_GPU_ENABLED="true"

# Security configuration
export OLLAMA_SECURITY_JWT_SECRET="your-jwt-secret-here"
export OLLAMA_SECURITY_AUTHENTICATION_ENABLED="true"

# Database configuration
export OLLAMA_DATABASE_DRIVER="postgres"
export OLLAMA_DATABASE_CONNECTION="postgres://user:pass@localhost/ollama"
```

## Command-Line Configuration

Override any setting using command-line flags:

```bash
# Basic server settings
ollama-distributed start --port 8082 --host 127.0.0.1

# Cluster settings
ollama-distributed start --cluster.node-id custom-node --cluster.bootstrap

# Model settings
ollama-distributed start --models.store-path /custom/path --models.cache-size 20GB

# Performance settings
ollama-distributed start --inference.max-concurrent 50 --inference.gpu.enabled

# Multiple overrides
ollama-distributed start \
  --port 8082 \
  --cluster.node-id production-1 \
  --models.cache-size 30GB \
  --inference.max-concurrent 25 \
  --logging.level debug
```

## Advanced Configuration

### Custom Model Registries

Configure multiple model registries:

```yaml
models:
  registries:
    - name: "ollama"
      url: "https://ollama.ai/library"
      auth: false
    - name: "huggingface"
      url: "https://huggingface.co"
      auth: true
      token: "${HF_TOKEN}"
    - name: "private"
      url: "https://models.company.com"
      auth: true
      username: "${REGISTRY_USER}"
      password: "${REGISTRY_PASS}"
```

### Load Balancing Configuration

Configure sophisticated load balancing:

```yaml
load_balancer:
  enabled: true
  strategy: "least_connections"  # round_robin, least_connections, weighted
  health_check:
    enabled: true
    interval: "10s"
    timeout: "5s"
    path: "/health"
  circuit_breaker:
    enabled: true
    threshold: 5
    timeout: "30s"
```

### Custom Resource Limits

Fine-tune resource allocation:

```yaml
resources:
  cpu:
    limit: "4.0"                # CPU cores
    reservation: "2.0"          # Reserved cores
  memory:
    limit: "16GB"               # Memory limit
    reservation: "8GB"          # Reserved memory
  disk:
    limit: "100GB"              # Disk space limit
    iops_limit: 1000           # IOPS limit
  network:
    bandwidth_limit: "1Gbps"    # Network bandwidth limit
```

### High Availability Configuration

Configure for maximum uptime:

```yaml
high_availability:
  enabled: true
  backup:
    enabled: true
    interval: "1h"
    retention: "7d"
    location: "s3://backup-bucket/ollama/"
  disaster_recovery:
    enabled: true
    replication_sites: 2
    recovery_time_objective: "15m"
    recovery_point_objective: "5m"
```

## Configuration Validation

### Validate Configuration

```bash
# Validate current configuration
ollama-distributed config validate

# Validate specific config file
ollama-distributed config validate --config /path/to/config.yaml

# Validate and show warnings
ollama-distributed config validate --warnings

# Validate against specific profile
ollama-distributed config validate --profile production
```

### Configuration Dry Run

Test configuration without starting:

```bash
# Dry run with current config
ollama-distributed start --dry-run

# Dry run with specific config
ollama-distributed start --config /path/to/config.yaml --dry-run

# Show what would be different
ollama-distributed config diff --profile production
```

## Configuration Management

### Export Configuration

```bash
# Export current configuration
ollama-distributed config export --format yaml > current-config.yaml

# Export with defaults included
ollama-distributed config export --include-defaults --format yaml

# Export specific sections
ollama-distributed config export --section cluster --format json
```

### Import Configuration

```bash
# Import configuration
ollama-distributed config import --file imported-config.yaml

# Merge with existing configuration
ollama-distributed config import --file new-settings.yaml --merge

# Import and validate
ollama-distributed config import --file config.yaml --validate
```

### Configuration Templates

Generate configuration templates:

```bash
# Generate template for specific use case
ollama-distributed config template --profile production --nodes 5

# Generate Docker Compose template
ollama-distributed config template --format docker-compose --nodes 3

# Generate Kubernetes template
ollama-distributed config template --format kubernetes --namespace ollama
```

## Troubleshooting Configuration

### Common Configuration Issues

#### Invalid Configuration Format
```bash
# Check configuration syntax
ollama-distributed config validate --syntax

# Show configuration parsing errors
ollama-distributed config validate --verbose
```

#### Port Conflicts
```bash
# Check port availability
ollama-distributed config validate --ports

# Find alternative ports
ollama-distributed config suggest --ports
```

#### Resource Constraints
```bash
# Check resource requirements
ollama-distributed config validate --resources

# Show resource recommendations
ollama-distributed config recommend --target production
```

#### Network Configuration Issues
```bash
# Test network connectivity
ollama-distributed config validate --network

# Show network configuration
ollama-distributed config show --section network
```

### Configuration Debugging

Enable configuration debugging:

```bash
# Start with configuration debugging
ollama-distributed start --debug-config

# Show effective configuration
ollama-distributed config effective

# Trace configuration loading
ollama-distributed config trace
```

## Best Practices

### Security Best Practices

1. **Always use TLS in production**
2. **Enable authentication and authorization**
3. **Use environment variables for secrets**
4. **Rotate encryption keys regularly**
5. **Enable audit logging**

### Performance Best Practices

1. **Configure appropriate cache sizes**
2. **Tune concurrent request limits**
3. **Use GPU acceleration when available**
4. **Monitor resource utilization**
5. **Implement proper load balancing**

### Operational Best Practices

1. **Use configuration profiles**
2. **Validate configurations before deployment**
3. **Use version control for configurations**
4. **Document custom configurations**
5. **Regular configuration backups**

## Next Steps

- **[Deployment Guide](../deployment/overview.md)** - Deploy your configured cluster
- **[Security Guide](../security.md)** - Secure your cluster
- **[Monitoring Guide](../monitoring.md)** - Set up monitoring
- **[API Reference](../api/overview.md)** - Explore API configuration options