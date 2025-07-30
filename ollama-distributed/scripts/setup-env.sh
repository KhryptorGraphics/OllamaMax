#!/bin/bash

# Environment Setup Script for OllamaMax
# This script sets up environment-specific configurations

set -euo pipefail

# Color output functions
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Default values
ENVIRONMENT="development"
NODE_NAME=""
REGION="us-west-2"
ZONE="us-west-2a"
GENERATE_CERTS=false
OUTPUT_FILE=""

# Usage function
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Setup environment-specific configuration for OllamaMax.

OPTIONS:
    -e, --environment ENV    Environment (development|staging|production) [default: development]
    -n, --node-name NAME     Node name [default: auto-generated]
    -r, --region REGION      AWS region [default: us-west-2]
    -z, --zone ZONE          Availability zone [default: us-west-2a]
    -c, --generate-certs     Generate self-signed TLS certificates
    -o, --output FILE        Output environment file [default: .env]
    -h, --help               Show this help message

EXAMPLES:
    # Setup development environment
    $0 --environment development

    # Setup production environment with TLS certificates
    $0 --environment production --generate-certs

    # Setup staging environment with custom node name
    $0 --environment staging --node-name my-staging-node

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -e|--environment)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -n|--node-name)
            NODE_NAME="$2"
            shift 2
            ;;
        -r|--region)
            REGION="$2"
            shift 2
            ;;
        -z|--zone)
            ZONE="$2"
            shift 2
            ;;
        -c|--generate-certs)
            GENERATE_CERTS=true
            shift
            ;;
        -o|--output)
            OUTPUT_FILE="$2"
            shift 2
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            error "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Validate environment
if [[ ! "$ENVIRONMENT" =~ ^(development|staging|production)$ ]]; then
    error "Invalid environment: $ENVIRONMENT. Must be one of: development, staging, production"
    exit 1
fi

# Set default output file if not specified
if [[ -z "$OUTPUT_FILE" ]]; then
    OUTPUT_FILE=".env.${ENVIRONMENT}"
fi

# Generate node name if not provided
if [[ -z "$NODE_NAME" ]]; then
    HOSTNAME=$(hostname -s 2>/dev/null || echo "unknown")
    NODE_NAME="ollama-${ENVIRONMENT}-${HOSTNAME}"
fi

info "Setting up environment: $ENVIRONMENT"
info "Node name: $NODE_NAME"
info "Region: $REGION"
info "Zone: $ZONE"

# Create directories
info "Creating necessary directories..."
mkdir -p config
mkdir -p data/${ENVIRONMENT}
mkdir -p logs/${ENVIRONMENT}
mkdir -p certs/${ENVIRONMENT}

# Generate environment configuration
info "Generating environment configuration..."

cat > "$OUTPUT_FILE" << EOF
# OllamaMax Environment Configuration
# Environment: $ENVIRONMENT
# Generated: $(date -u +"%Y-%m-%d %H:%M:%S UTC")

# Node Configuration
export OLLAMA_ENVIRONMENT="$ENVIRONMENT"
export OLLAMA_NODE_ID=""  # Auto-generated if empty
export OLLAMA_NODE_NAME="$NODE_NAME"
export OLLAMA_NODE_REGION="$REGION"
export OLLAMA_NODE_ZONE="$ZONE"

# API Configuration
export OLLAMA_API_LISTEN="0.0.0.0:11434"
export OLLAMA_TLS_ENABLED="$([ "$ENVIRONMENT" = "development" ] && echo "false" || echo "true")"
export OLLAMA_TLS_CERT_FILE="/etc/tls/${ENVIRONMENT}/server.crt"
export OLLAMA_TLS_KEY_FILE="/etc/tls/${ENVIRONMENT}/server.key"

# Authentication Configuration  
export OLLAMA_AUTH_ENABLED="$([ "$ENVIRONMENT" = "development" ] && echo "false" || echo "true")"
export OLLAMA_AUTH_METHOD="jwt"
export OLLAMA_JWT_SECRET="$(openssl rand -base64 32)"

# Metrics Configuration
export OLLAMA_METRICS_ENABLED="true"
export OLLAMA_METRICS_LISTEN="0.0.0.0:9090"

# Logging Configuration
export OLLAMA_LOG_LEVEL="$([ "$ENVIRONMENT" = "development" ] && echo "debug" || [ "$ENVIRONMENT" = "staging" ] && echo "info" || echo "warn")"
export OLLAMA_LOG_FORMAT="$([ "$ENVIRONMENT" = "development" ] && echo "console" || echo "json")"
export OLLAMA_LOG_OUTPUT="$([ "$ENVIRONMENT" = "development" ] && echo "stdout" || echo "file")"
export OLLAMA_LOG_FILE="/var/log/ollama/${ENVIRONMENT}/application.log"

# Storage Configuration
export OLLAMA_DATA_DIR="$([ "$ENVIRONMENT" = "development" ] && echo "./data/dev" || echo "/var/lib/ollama/${ENVIRONMENT}")"
export OLLAMA_MODEL_DIR="$([ "$ENVIRONMENT" = "development" ] && echo "./data/dev/models" || echo "/var/lib/ollama/${ENVIRONMENT}/models")"
export OLLAMA_CACHE_DIR="$([ "$ENVIRONMENT" = "development" ] && echo "./data/dev/cache" || echo "/var/cache/ollama/${ENVIRONMENT}")"

# Consensus Configuration
export OLLAMA_CONSENSUS_BOOTSTRAP="$([ "$ENVIRONMENT" = "development" ] && echo "true" || echo "false")"
export OLLAMA_CONSENSUS_BIND_ADDR="0.0.0.0:7000"
export OLLAMA_CONSENSUS_ADVERTISE_ADDR=""  # Auto-detected if empty

# Database Configuration (if applicable)
export DATABASE_URL="postgresql://ollama:password@localhost:5432/ollama_${ENVIRONMENT}"
export REDIS_URL="redis://localhost:6379/0"

# Monitoring Configuration
export PROMETHEUS_URL="http://localhost:9090"
export GRAFANA_URL="http://localhost:3000"
export JAEGER_URL="http://localhost:14268/api/traces"

# Cloud Provider Configuration (AWS example)
export AWS_REGION="$REGION"
export AWS_DEFAULT_REGION="$REGION"

# Kubernetes Configuration (if applicable)
export KUBE_NAMESPACE="ollama-${ENVIRONMENT}"
export KUBE_SERVICE_ACCOUNT="ollama-${ENVIRONMENT}"

EOF

success "Environment configuration written to: $OUTPUT_FILE"

# Generate TLS certificates if requested
if [[ "$GENERATE_CERTS" = true ]]; then
    info "Generating self-signed TLS certificates..."
    
    CERT_DIR="certs/${ENVIRONMENT}"
    
    # Generate CA key and certificate
    openssl genrsa -out "${CERT_DIR}/ca.key" 4096
    openssl req -new -x509 -days 365 -key "${CERT_DIR}/ca.key" -out "${CERT_DIR}/ca.crt" \
        -subj "/C=US/ST=California/L=San Francisco/O=OllamaMax/CN=OllamaMax CA"
    
    # Generate server key and certificate signing request
    openssl genrsa -out "${CERT_DIR}/server.key" 4096
    openssl req -new -key "${CERT_DIR}/server.key" -out "${CERT_DIR}/server.csr" \
        -subj "/C=US/ST=California/L=San Francisco/O=OllamaMax/CN=${NODE_NAME}"
    
    # Generate server certificate
    openssl x509 -req -days 365 -in "${CERT_DIR}/server.csr" \
        -CA "${CERT_DIR}/ca.crt" -CAkey "${CERT_DIR}/ca.key" -CAcreateserial \
        -out "${CERT_DIR}/server.crt"
    
    # Generate client key and certificate (for mutual TLS)
    openssl genrsa -out "${CERT_DIR}/client.key" 4096
    openssl req -new -key "${CERT_DIR}/client.key" -out "${CERT_DIR}/client.csr" \
        -subj "/C=US/ST=California/L=San Francisco/O=OllamaMax/CN=client"
    
    openssl x509 -req -days 365 -in "${CERT_DIR}/client.csr" \
        -CA "${CERT_DIR}/ca.crt" -CAkey "${CERT_DIR}/ca.key" -CAcreateserial \
        -out "${CERT_DIR}/client.crt"
    
    # Set proper permissions
    chmod 600 "${CERT_DIR}"/*.key
    chmod 644 "${CERT_DIR}"/*.crt
    
    # Clean up CSR files
    rm -f "${CERT_DIR}"/*.csr
    
    success "TLS certificates generated in: $CERT_DIR"
    
    # Update environment file with correct cert paths
    sed -i "s|/etc/tls/${ENVIRONMENT}/|$(pwd)/${CERT_DIR}/|g" "$OUTPUT_FILE"
fi

# Create systemd service file for production
if [[ "$ENVIRONMENT" = "production" ]]; then
    info "Creating systemd service file..."
    
    cat > "ollama-distributed.service" << EOF
[Unit]
Description=OllamaMax Distributed AI Platform
Documentation=https://github.com/khryptorgraphics/ollamamax
After=network.target

[Service]
Type=simple
User=ollama
Group=ollama
WorkingDirectory=/opt/ollama
EnvironmentFile=/etc/ollama/environment
ExecStart=/usr/local/bin/ollama-distributed start
ExecReload=/bin/kill -HUP \$MAINPID
KillMode=mixed
KillSignal=SIGTERM
TimeoutStopSec=30
Restart=on-failure
RestartSec=5
StartLimitInterval=60s
StartLimitBurst=3

# Security settings
NoNewPrivileges=true
PrivateTmp=true
PrivateDevices=true
ProtectHome=true
ProtectSystem=strict
ReadWritePaths=/var/lib/ollama /var/log/ollama /var/cache/ollama

# Resource limits
LimitNOFILE=65536
LimitNPROC=32768

[Install]
WantedBy=multi-user.target
EOF
    
    success "Systemd service file created: ollama-distributed.service"
    info "To install: sudo cp ollama-distributed.service /etc/systemd/system/"
    info "To enable: sudo systemctl enable ollama-distributed"
fi

# Create Docker Compose file for development
if [[ "$ENVIRONMENT" = "development" ]]; then
    info "Creating Docker Compose file for development..."
    
    cat > "docker-compose.${ENVIRONMENT}.yml" << EOF
version: '3.8'

services:
  ollama-node-1:
    build: .
    ports:
      - "11434:11434"
      - "8080:8080"
      - "9090:9090"
    environment:
      - OLLAMA_ENVIRONMENT=development
      - OLLAMA_NODE_NAME=ollama-dev-node-1
      - OLLAMA_API_LISTEN=0.0.0.0:11434
      - OLLAMA_CONSENSUS_BOOTSTRAP=true
    volumes:
      - ./data/dev/node1:/var/lib/ollama
      - ./logs/dev/node1:/var/log/ollama
      - ./config:/etc/ollama
    networks:
      - ollama-network

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9091:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
    networks:
      - ollama-network

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana-storage:/var/lib/grafana
      - ./monitoring/grafana/dashboards:/etc/grafana/provisioning/dashboards
    networks:
      - ollama-network

networks:
  ollama-network:
    driver: bridge

volumes:
  grafana-storage:
EOF
    
    success "Docker Compose file created: docker-compose.${ENVIRONMENT}.yml"
fi

# Create helpful scripts
info "Creating helper scripts..."

# Environment loader script
cat > "load-env.sh" << EOF
#!/bin/bash
# Load environment variables for OllamaMax

if [[ -f "$OUTPUT_FILE" ]]; then
    source "$OUTPUT_FILE"
    echo "Environment $ENVIRONMENT loaded from $OUTPUT_FILE"
else
    echo "Error: Environment file $OUTPUT_FILE not found"
    exit 1
fi
EOF

chmod +x load-env.sh

# Configuration validator script
cat > "validate-config.sh" << EOF
#!/bin/bash
# Validate OllamaMax configuration

set -euo pipefail

echo "Validating OllamaMax configuration..."

# Load environment
source "$OUTPUT_FILE"

# Check required directories
echo "Checking directories..."
mkdir -p "\$OLLAMA_DATA_DIR" "\$OLLAMA_MODEL_DIR" "\$OLLAMA_CACHE_DIR"

# Check TLS certificates if enabled
if [[ "\$OLLAMA_TLS_ENABLED" = "true" ]]; then
    echo "Checking TLS certificates..."
    if [[ ! -f "\$OLLAMA_TLS_CERT_FILE" ]] || [[ ! -f "\$OLLAMA_TLS_KEY_FILE" ]]; then
        echo "Warning: TLS enabled but certificates not found"
        echo "Generate certificates with: $0 --generate-certs"
    else
        echo "TLS certificates found"
    fi
fi

# Check JWT secret
if [[ "\$OLLAMA_AUTH_ENABLED" = "true" ]] && [[ -z "\$OLLAMA_JWT_SECRET" ]]; then
    echo "Warning: Authentication enabled but JWT secret not set"
fi

echo "Configuration validation complete"
EOF

chmod +x validate-config.sh

success "Helper scripts created: load-env.sh, validate-config.sh"

# Print next steps
info "Next steps:"
echo "1. Source the environment: source $OUTPUT_FILE"
echo "2. Validate configuration: ./validate-config.sh"
echo "3. Build the application: make build"
echo "4. Start the service: ./ollama-distributed start"

if [[ "$ENVIRONMENT" = "development" ]]; then
    echo "5. Or use Docker Compose: docker-compose -f docker-compose.${ENVIRONMENT}.yml up"
fi

success "Environment setup complete for: $ENVIRONMENT"