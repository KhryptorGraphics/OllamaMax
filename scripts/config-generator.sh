#!/bin/bash

# OllamaMax Configuration Generator
# Generates different configuration profiles for various deployment scenarios

set -euo pipefail
IFS=$'\n\t'

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
PROFILE="development"
OUTPUT="config.yaml"
NODE_NAME="ollama-node-001"
API_PORT=8080
WEB_PORT=8081
P2P_PORT=4001
DATA_DIR="$HOME/.ollamamax/data"
SECURITY="basic"
GPU_ENABLED="false"
NODES=1

# Function to display help
show_help() {
    cat << EOF
🔧 OllamaMax Configuration Generator
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Generate optimized configuration files for different deployment scenarios.

Usage: $0 [OPTIONS]

Options:
  -h, --help              Show this help message
  -p, --profile PROFILE   Configuration profile to use
                          Options: development, production, edge, gpu, cluster
  -o, --output FILE       Output configuration file (default: config.yaml)
  -n, --node-name NAME    Node name (default: ollama-node-001)
  --api-port PORT         API server port (default: 8080)
  --web-port PORT         Web interface port (default: 8081)
  --p2p-port PORT         P2P network port (default: 4001)
  --data-dir PATH         Data directory path (default: ~/.ollamamax/data)
  --security LEVEL        Security level: basic, standard, enterprise (default: basic)
  --gpu                   Enable GPU support
  --nodes COUNT           Number of nodes for cluster profile (default: 1)

Available Profiles:
  development   - Local development with debug logging and minimal security
  production    - Production-ready with security, monitoring, and optimization
  edge          - Edge/IoT deployment with resource constraints
  gpu           - GPU-optimized for NVIDIA/AMD hardware acceleration
  cluster       - Multi-node cluster configuration

Examples:
  # Generate development configuration
  $0 --profile development

  # Generate production configuration with custom output
  $0 --profile production --output prod-config.yaml

  # Generate GPU-optimized configuration
  $0 --profile gpu --gpu --output gpu-config.yaml

  # Generate 3-node cluster configuration
  $0 --profile cluster --nodes 3 --output cluster-config.yaml

  # Generate enterprise security configuration
  $0 --profile production --security enterprise
EOF
}

# Input validation functions
validate_port() {
    local port=$1
    if ! [[ "$port" =~ ^[0-9]+$ ]] || [ "$port" -lt 1024 ] || [ "$port" -gt 65535 ]; then
        echo -e "${RED}❌ Invalid port: $port${NC}" >&2
        return 1
    fi
}

validate_node_name() {
    local name=$1
    if [[ ! "$name" =~ ^[a-zA-Z0-9_-]+$ ]]; then
        echo -e "${RED}❌ Invalid node name: $name (only alphanumeric, underscore, hyphen allowed)${NC}" >&2
        return 1
    fi
}

validate_secrets() {
    if [[ "${PROFILE}" == "production" ]]; then
        if [[ ! -v JWT_SECRET ]] || [[ -z "${JWT_SECRET}" ]]; then
            echo -e "${RED}❌ ERROR: JWT_SECRET environment variable required for production${NC}" >&2
            exit 1
        fi
        if [[ "${JWT_SECRET}" == *"development"* ]] || [[ "${JWT_SECRET}" == *"dev-secret"* ]]; then
            echo -e "${RED}❌ ERROR: Development secrets detected in production profile${NC}" >&2
            exit 1
        fi
    fi
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -p|--profile)
            PROFILE="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT="$2"
            shift 2
            ;;
        -n|--node-name)
            validate_node_name "$2"
            NODE_NAME="$2"
            shift 2
            ;;
        --api-port)
            validate_port "$2"
            API_PORT="$2"
            shift 2
            ;;
        --web-port)
            validate_port "$2"
            WEB_PORT="$2"
            shift 2
            ;;
        --p2p-port)
            validate_port "$2"
            P2P_PORT="$2"
            shift 2
            ;;
        --data-dir)
            DATA_DIR="$2"
            shift 2
            ;;
        --security)
            SECURITY="$2"
            shift 2
            ;;
        --gpu)
            GPU_ENABLED="true"
            shift
            ;;
        --nodes)
            NODES="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Function to generate development configuration
generate_development() {
    cat > "$OUTPUT" << EOF
# OllamaMax Development Configuration
# Generated on: $(date)
# Profile: development
# Features: Debug logging, single node, minimal security

node:
  id: "${NODE_NAME}"
  name: "Development Node"
  environment: "development"
  data_dir: "${DATA_DIR}"

api:
  host: "127.0.0.1"
  port: ${API_PORT}
  enable_tls: false
  cors:
    enabled: true
    origins: ["*"]
    credentials: false

web:
  enabled: true
  host: "127.0.0.1"
  port: ${WEB_PORT}
  enable_tls: false

p2p:
  enabled: true
  port: ${P2P_PORT}
  bootstrap_peers: []
  max_peers: 10
  enable_discovery: true

models:
  store_path: "${DATA_DIR}/models"
  max_cache_size: "10GB"
  auto_cleanup: true
  download_timeout: "30m"
  concurrency: 4

performance:
  max_workers: 4
  max_connections: 100
  request_timeout: "30s"
  gpu_enabled: ${GPU_ENABLED}

security:
  authentication: false
  encryption: false
  rate_limiting: false

logging:
  level: "debug"
  format: "text"
  output: "stdout"
  file: "${DATA_DIR}/logs/ollama.log"
  max_size: "100MB"
  max_backups: 5
  max_age: 7

monitoring:
  enabled: true
  metrics_port: 9090
  health_check_interval: "10s"

development:
  hot_reload: true
  verbose_errors: true
  enable_profiling: true
  debug_endpoints: true
EOF
    echo -e "${GREEN}✅ Development configuration generated: ${OUTPUT}${NC}"
}

# Function to generate production configuration
generate_production() {
    cat > "$OUTPUT" << EOF
# OllamaMax Production Configuration
# Generated on: $(date)
# Profile: production
# Features: Security hardening, monitoring, high availability

node:
  id: "${NODE_NAME}"
  name: "Production Node"
  environment: "production"
  data_dir: "${DATA_DIR}"

api:
  host: "0.0.0.0"
  port: ${API_PORT}
  enable_tls: true
  tls:
    cert_file: "/etc/ollamamax/certs/server.crt"
    key_file: "/etc/ollamamax/certs/server.key"
    ca_file: "/etc/ollamamax/certs/ca.crt"
    min_version: "TLS1.3"
  cors:
    enabled: true
    origins: ["https://app.example.com"]
    credentials: true
  max_request_size: "100MB"

web:
  enabled: true
  host: "0.0.0.0"
  port: ${WEB_PORT}
  enable_tls: true
  tls:
    cert_file: "/etc/ollamamax/certs/web.crt"
    key_file: "/etc/ollamamax/certs/web.key"

p2p:
  enabled: true
  port: ${P2P_PORT}
  bootstrap_peers:
    - "/dns4/node1.ollamamax.com/tcp/4001/p2p/QmNode1"
    - "/dns4/node2.ollamamax.com/tcp/4001/p2p/QmNode2"
  max_peers: 50
  enable_discovery: true
  enable_relay: true

models:
  store_path: "${DATA_DIR}/models"
  max_cache_size: "100GB"
  auto_cleanup: true
  download_timeout: "60m"
  concurrency: 8
  integrity_check: true
  encryption_at_rest: true

performance:
  max_workers: 16
  max_connections: 1000
  request_timeout: "60s"
  gpu_enabled: ${GPU_ENABLED}
  cache:
    enabled: true
    size: "10GB"
    ttl: "1h"

security:
  authentication: true
  auth_type: "jwt"
  jwt_secret: "\${JWT_SECRET:?JWT_SECRET environment variable required}"
  encryption: true
  rate_limiting: true
  rate_limit:
    requests_per_minute: 60
    burst: 100
  ip_whitelist: []
  audit_logging: true

logging:
  level: "info"
  format: "json"
  output: "file"
  file: "${DATA_DIR}/logs/ollama.log"
  max_size: "1GB"
  max_backups: 30
  max_age: 90
  syslog:
    enabled: true
    host: "syslog.example.com"
    port: 514

monitoring:
  enabled: true
  metrics_port: 9090
  health_check_interval: "30s"
  prometheus:
    enabled: true
    path: "/metrics"
  grafana:
    enabled: true
    dashboard_id: "ollama-prod"
  alerting:
    enabled: true
    webhook_url: "\${ALERT_WEBHOOK_URL}"

high_availability:
  enabled: true
  replication_factor: 3
  consistency_level: "quorum"
  failover_timeout: "30s"

backup:
  enabled: true
  schedule: "0 2 * * *"
  retention_days: 30
  s3:
    enabled: true
    bucket: "ollamamax-backups"
    region: "us-west-2"
EOF
    echo -e "${GREEN}✅ Production configuration generated: ${OUTPUT}${NC}"
}

# Function to generate edge configuration
generate_edge() {
    cat > "$OUTPUT" << EOF
# OllamaMax Edge/IoT Configuration
# Generated on: $(date)
# Profile: edge
# Features: Resource-constrained, lightweight, offline-capable

node:
  id: "${NODE_NAME}"
  name: "Edge Node"
  environment: "edge"
  data_dir: "${DATA_DIR}"

api:
  host: "127.0.0.1"
  port: ${API_PORT}
  enable_tls: false
  max_request_size: "10MB"

web:
  enabled: false

p2p:
  enabled: true
  port: ${P2P_PORT}
  bootstrap_peers: []
  max_peers: 5
  enable_discovery: false

models:
  store_path: "${DATA_DIR}/models"
  max_cache_size: "2GB"
  auto_cleanup: true
  download_timeout: "120m"
  concurrency: 1
  prefer_quantized: true
  max_model_size: "1GB"

performance:
  max_workers: 2
  max_connections: 10
  request_timeout: "120s"
  gpu_enabled: false
  memory_limit: "1GB"
  cpu_limit: 2

security:
  authentication: false
  encryption: false
  rate_limiting: false

logging:
  level: "warn"
  format: "text"
  output: "stdout"
  max_size: "10MB"
  max_backups: 1
  max_age: 7

monitoring:
  enabled: false

edge:
  offline_mode: true
  power_saving: true
  adaptive_quality: true
  local_caching: true
  compression: true
EOF
    echo -e "${GREEN}✅ Edge configuration generated: ${OUTPUT}${NC}"
}

# Function to generate GPU configuration
generate_gpu() {
    cat > "$OUTPUT" << EOF
# OllamaMax GPU-Optimized Configuration
# Generated on: $(date)
# Profile: gpu
# Features: GPU acceleration, optimized memory, parallel processing

node:
  id: "${NODE_NAME}"
  name: "GPU Node"
  environment: "gpu"
  data_dir: "${DATA_DIR}"

api:
  host: "0.0.0.0"
  port: ${API_PORT}
  enable_tls: false
  max_request_size: "500MB"

web:
  enabled: true
  host: "0.0.0.0"
  port: ${WEB_PORT}

p2p:
  enabled: true
  port: ${P2P_PORT}
  bootstrap_peers: []
  max_peers: 20

models:
  store_path: "${DATA_DIR}/models"
  max_cache_size: "200GB"
  auto_cleanup: false
  download_timeout: "60m"
  concurrency: 8
  prefer_fp16: true
  enable_quantization: false

performance:
  max_workers: 32
  max_connections: 500
  request_timeout: "300s"
  gpu_enabled: true
  gpu:
    device_ids: [0, 1, 2, 3]
    memory_fraction: 0.9
    allow_growth: true
    enable_mixed_precision: true
    enable_xla: true
    enable_tensor_cores: true
  batch_processing:
    enabled: true
    max_batch_size: 32
    batch_timeout: "100ms"

security:
  authentication: true
  encryption: false
  rate_limiting: true
  rate_limit:
    requests_per_minute: 120
    burst: 200

logging:
  level: "info"
  format: "json"
  output: "file"
  file: "${DATA_DIR}/logs/ollama-gpu.log"
  max_size: "500MB"
  max_backups: 10
  max_age: 30

monitoring:
  enabled: true
  metrics_port: 9090
  health_check_interval: "10s"
  gpu_monitoring:
    enabled: true
    nvidia_smi: true
    dcgm_exporter: true
    metrics:
      - "gpu_utilization"
      - "memory_utilization"
      - "temperature"
      - "power_draw"

gpu_optimization:
  cuda_graphs: true
  flash_attention: true
  compile_mode: "max-autotune"
  memory_pool: "cudaMallocAsync"
EOF
    echo -e "${GREEN}✅ GPU configuration generated: ${OUTPUT}${NC}"
}

# Function to generate cluster configuration
generate_cluster() {
    cat > "$OUTPUT" << EOF
# OllamaMax Cluster Configuration
# Generated on: $(date)
# Profile: cluster
# Features: Multi-node, consensus, distributed processing

cluster:
  name: "ollamamax-cluster"
  size: ${NODES}
  
nodes:
EOF
    
    # Generate node configurations
    for ((i=1; i<=$NODES; i++)); do
        cat >> "$OUTPUT" << EOF
  - id: "node-$i"
    name: "Cluster Node $i"
    address: "192.168.1.$((100+i))"
    api_port: $((API_PORT + i - 1))
    p2p_port: $((P2P_PORT + i - 1))
    role: $([ $i -eq 1 ] && echo "leader" || echo "follower")
EOF
    done
    
    cat >> "$OUTPUT" << EOF

consensus:
  algorithm: "raft"
  election_timeout: "5s"
  heartbeat_interval: "1s"
  snapshot_threshold: 1000
  log_entries_max: 10000

p2p:
  enabled: true
  discovery:
    method: "mdns"
    interval: "30s"
  gossip:
    enabled: true
    interval: "5s"
    fanout: 3

models:
  replication_factor: $([ $NODES -ge 3 ] && echo "3" || echo "$NODES")
  distribution_strategy: "consistent-hash"
  shard_size: "10GB"
  auto_rebalance: true

load_balancing:
  algorithm: "least-connections"
  health_check:
    enabled: true
    interval: "10s"
    timeout: "5s"
    unhealthy_threshold: 3
    healthy_threshold: 2

performance:
  max_workers_per_node: 8
  max_connections_per_node: 200
  request_timeout: "60s"
  distributed_cache:
    enabled: true
    size: "50GB"
    eviction_policy: "lru"

security:
  inter_node_encryption: true
  tls:
    enabled: true
    mutual_auth: true
    cert_rotation: "30d"
  authentication: true
  
logging:
  level: "info"
  format: "json"
  centralized:
    enabled: true
    endpoint: "logs.ollamamax.local"

monitoring:
  enabled: true
  metrics_aggregation: true
  distributed_tracing: true
  service_mesh:
    enabled: true
    provider: "linkerd"
EOF
    echo -e "${GREEN}✅ Cluster configuration generated: ${OUTPUT}${NC}"
}

# Validate configuration before generation
validate_secrets

# Generate configuration based on profile
echo -e "${BLUE}🔧 Generating ${PROFILE} configuration...${NC}"

case "$PROFILE" in
    development)
        generate_development
        ;;
    production)
        generate_production
        ;;
    edge)
        generate_edge
        ;;
    gpu)
        GPU_ENABLED="true"
        generate_gpu
        ;;
    cluster)
        generate_cluster
        ;;
    *)
        echo -e "${RED}❌ Unknown profile: ${PROFILE}${NC}"
        echo "Available profiles: development, production, edge, gpu, cluster"
        exit 1
        ;;
esac

# Validate the generated configuration
if [ -f "$OUTPUT" ]; then
    echo -e "${YELLOW}📝 Configuration saved to: ${OUTPUT}${NC}"
    echo -e "${BLUE}🔍 Validating configuration...${NC}"
    
    # Basic YAML validation
    if command -v python3 &> /dev/null; then
        python3 -c "import yaml; yaml.safe_load(open('$OUTPUT'))" 2>/dev/null
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✅ Configuration is valid YAML${NC}"
        else
            echo -e "${RED}❌ Configuration has YAML syntax errors${NC}"
            exit 1
        fi
    else
        echo -e "${YELLOW}⚠️  Python3 not found, skipping YAML validation${NC}"
    fi
    
    echo ""
    echo -e "${GREEN}🎉 Configuration generation complete!${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Review the configuration: cat ${OUTPUT}"
    echo "  2. Copy to OllamaMax directory: cp ${OUTPUT} ~/.ollamamax/"
    echo "  3. Start the node: ollama-distributed start --config ${OUTPUT}"
else
    echo -e "${RED}❌ Failed to generate configuration${NC}"
    exit 1
fi