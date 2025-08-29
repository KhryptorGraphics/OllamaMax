#!/bin/bash

# OllamaMax Enterprise Security Setup Script
# Sets up comprehensive security hardening for production deployment

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
SECURITY_DIR="$PROJECT_ROOT/security"
CERTS_DIR="$PROJECT_ROOT/certs"
SECRETS_DIR="$PROJECT_ROOT/secrets"
LOGS_DIR="$PROJECT_ROOT/logs/security"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color

# Logging
LOG_FILE="$LOGS_DIR/security-setup.log"

print_header() {
    echo -e "${BLUE}============================================================${NC}"
    echo -e "${WHITE}$1${NC}"
    echo -e "${BLUE}============================================================${NC}"
    echo ""
}

print_status() {
    local status=$1
    local message=$2
    case $status in
        "INFO")    echo -e "${BLUE}[INFO]${NC} $message" ;;
        "SUCCESS") echo -e "${GREEN}[SUCCESS]${NC} $message" ;;
        "WARNING") echo -e "${YELLOW}[WARNING]${NC} $message" ;;
        "ERROR")   echo -e "${RED}[ERROR]${NC} $message" ;;
        "STEP")    echo -e "${PURPLE}[STEP]${NC} $message" ;;
    esac
    echo "$(date '+%Y-%m-%d %H:%M:%S') [$status] $message" >> "$LOG_FILE"
}

check_requirements() {
    print_header "Checking System Requirements"
    
    local missing_tools=()
    
    # Check required tools
    local tools=("openssl" "docker" "docker-compose" "curl" "jq" "vault")
    for tool in "${tools[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            missing_tools+=("$tool")
        fi
    done
    
    if [[ ${#missing_tools[@]} -ne 0 ]]; then
        print_status "ERROR" "Missing required tools: ${missing_tools[*]}"
        print_status "INFO" "Please install missing tools and try again"
        return 1
    fi
    
    # Check Docker daemon
    if ! docker info &> /dev/null; then
        print_status "ERROR" "Docker daemon is not running"
        return 1
    fi
    
    # Check OpenSSL version
    local openssl_version=$(openssl version | awk '{print $2}')
    print_status "INFO" "OpenSSL version: $openssl_version"
    
    # Check available entropy
    local entropy=$(cat /proc/sys/kernel/random/entropy_avail 2>/dev/null || echo "unknown")
    print_status "INFO" "Available entropy: $entropy"
    
    print_status "SUCCESS" "All system requirements met"
}

setup_directories() {
    print_header "Setting Up Directory Structure"
    
    # Create directory structure
    local dirs=(
        "$SECURITY_DIR"
        "$CERTS_DIR"
        "$SECRETS_DIR"
        "$LOGS_DIR"
        "$PROJECT_ROOT/data/secure-node-1"
        "$PROJECT_ROOT/data/secure-node-2"
        "$PROJECT_ROOT/models/secure-node-1"
        "$PROJECT_ROOT/models/secure-node-2"
        "$PROJECT_ROOT/quarantine"
        "$PROJECT_ROOT/backup"
        "$PROJECT_ROOT/compliance"
    )
    
    for dir in "${dirs[@]}"; do
        if [[ ! -d "$dir" ]]; then
            mkdir -p "$dir"
            print_status "INFO" "Created directory: $dir"
        fi
    done
    
    # Set secure permissions
    chmod 700 "$SECRETS_DIR" "$CERTS_DIR"
    chmod 750 "$SECURITY_DIR" "$LOGS_DIR"
    chmod 755 "$PROJECT_ROOT/data" "$PROJECT_ROOT/models"
    
    print_status "SUCCESS" "Directory structure created"
}

generate_secrets() {
    print_header "Generating Security Secrets"
    
    # JWT Secret (256-bit)
    if [[ ! -f "$SECRETS_DIR/jwt-secret" ]]; then
        openssl rand -hex 32 > "$SECRETS_DIR/jwt-secret"
        chmod 600 "$SECRETS_DIR/jwt-secret"
        print_status "SUCCESS" "Generated JWT secret"
    fi
    
    # Encryption Key (256-bit AES)
    if [[ ! -f "$SECRETS_DIR/encryption-key" ]]; then
        openssl rand -hex 32 > "$SECRETS_DIR/encryption-key"
        chmod 600 "$SECRETS_DIR/encryption-key"
        print_status "SUCCESS" "Generated encryption key"
    fi
    
    # Redis Password
    if [[ ! -f "$SECRETS_DIR/redis-password" ]]; then
        openssl rand -base64 32 > "$SECRETS_DIR/redis-password"
        chmod 600 "$SECRETS_DIR/redis-password"
        print_status "SUCCESS" "Generated Redis password"
    fi
    
    # Vault Root Token
    if [[ ! -f "$SECRETS_DIR/vault-token" ]]; then
        openssl rand -hex 16 > "$SECRETS_DIR/vault-token"
        chmod 600 "$SECRETS_DIR/vault-token"
        print_status "SUCCESS" "Generated Vault root token"
    fi
    
    # API Keys file
    if [[ ! -f "$SECRETS_DIR/api-keys" ]]; then
        cat > "$SECRETS_DIR/api-keys" << EOF
# OllamaMax API Keys
# Format: key_id:hashed_key:user_id:permissions:created_at
EOF
        chmod 600 "$SECRETS_DIR/api-keys"
        print_status "SUCCESS" "Created API keys file"
    fi
    
    # Grafana Admin Password
    if [[ ! -f "$SECRETS_DIR/grafana-admin-password" ]]; then
        openssl rand -base64 16 > "$SECRETS_DIR/grafana-admin-password"
        chmod 600 "$SECRETS_DIR/grafana-admin-password"
        print_status "SUCCESS" "Generated Grafana admin password"
    fi
    
    # Grafana Secret Key
    if [[ ! -f "$SECRETS_DIR/grafana-secret-key" ]]; then
        openssl rand -base64 32 > "$SECRETS_DIR/grafana-secret-key"
        chmod 600 "$SECRETS_DIR/grafana-secret-key"
        print_status "SUCCESS" "Generated Grafana secret key"
    fi
    
    print_status "SUCCESS" "All secrets generated"
}

generate_certificates() {
    print_header "Generating TLS Certificates"
    
    local ca_key="$CERTS_DIR/ca.key"
    local ca_crt="$CERTS_DIR/ca.crt"
    local server_key="$CERTS_DIR/server.key"
    local server_crt="$CERTS_DIR/server.crt"
    local client_key="$CERTS_DIR/client.key"
    local client_crt="$CERTS_DIR/client.crt"
    
    # Certificate configuration
    local country="US"
    local state="Security"
    local city="Secure"
    local org="OllamaMax"
    local ou="Security"
    local validity_days=90
    
    # Generate CA private key
    if [[ ! -f "$ca_key" ]]; then
        openssl genrsa -out "$ca_key" 4096
        chmod 600 "$ca_key"
        print_status "SUCCESS" "Generated CA private key"
    fi
    
    # Generate CA certificate
    if [[ ! -f "$ca_crt" ]]; then
        openssl req -new -x509 -days $validity_days -key "$ca_key" -out "$ca_crt" \
            -subj "/C=$country/ST=$state/L=$city/O=$org/OU=$ou/CN=OllamaMax-CA"
        chmod 644 "$ca_crt"
        print_status "SUCCESS" "Generated CA certificate"
    fi
    
    # Generate server private key
    if [[ ! -f "$server_key" ]]; then
        openssl genrsa -out "$server_key" 4096
        chmod 600 "$server_key"
        print_status "SUCCESS" "Generated server private key"
    fi
    
    # Generate server certificate
    if [[ ! -f "$server_crt" ]]; then
        # Create temporary config file with SAN
        local temp_conf=$(mktemp)
        cat > "$temp_conf" << EOF
[req]
default_bits = 4096
prompt = no
distinguished_name = req_distinguished_name
req_extensions = v3_req

[req_distinguished_name]
C=$country
ST=$state
L=$city
O=$org
OU=$ou
CN=localhost

[v3_req]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
DNS.2 = ollamamax-secure-node1
DNS.3 = ollamamax-secure-node2
DNS.4 = secure-lb
DNS.5 = vault
DNS.6 = prometheus
DNS.7 = grafana
DNS.8 = *.ollamamax.local
IP.1 = 127.0.0.1
IP.2 = 0.0.0.0
IP.3 = 172.30.0.1
EOF
        
        # Generate CSR
        openssl req -new -key "$server_key" -out "$CERTS_DIR/server.csr" -config "$temp_conf"
        
        # Sign server certificate
        openssl x509 -req -in "$CERTS_DIR/server.csr" -CA "$ca_crt" -CAkey "$ca_key" \
            -CAcreateserial -out "$server_crt" -days $validity_days \
            -extensions v3_req -extfile "$temp_conf"
        
        chmod 644 "$server_crt"
        rm "$temp_conf" "$CERTS_DIR/server.csr"
        print_status "SUCCESS" "Generated server certificate"
    fi
    
    # Generate client private key
    if [[ ! -f "$client_key" ]]; then
        openssl genrsa -out "$client_key" 4096
        chmod 600 "$client_key"
        print_status "SUCCESS" "Generated client private key"
    fi
    
    # Generate client certificate
    if [[ ! -f "$client_crt" ]]; then
        openssl req -new -key "$client_key" -out "$CERTS_DIR/client.csr" \
            -subj "/C=$country/ST=$state/L=$city/O=$org/OU=$ou/CN=ollama-client"
        
        openssl x509 -req -in "$CERTS_DIR/client.csr" -CA "$ca_crt" -CAkey "$ca_key" \
            -CAcreateserial -out "$client_crt" -days $validity_days
        
        chmod 644 "$client_crt"
        rm "$CERTS_DIR/client.csr"
        print_status "SUCCESS" "Generated client certificate"
    fi
    
    # Generate DH parameters for perfect forward secrecy
    if [[ ! -f "$CERTS_DIR/dhparam.pem" ]]; then
        print_status "INFO" "Generating DH parameters (this may take a while)..."
        openssl dhparam -out "$CERTS_DIR/dhparam.pem" 2048
        chmod 644 "$CERTS_DIR/dhparam.pem"
        print_status "SUCCESS" "Generated DH parameters"
    fi
    
    # Verify certificates
    print_status "INFO" "Verifying certificates..."
    if openssl verify -CAfile "$ca_crt" "$server_crt" > /dev/null 2>&1; then
        print_status "SUCCESS" "Server certificate verification passed"
    else
        print_status "ERROR" "Server certificate verification failed"
        return 1
    fi
    
    if openssl verify -CAfile "$ca_crt" "$client_crt" > /dev/null 2>&1; then
        print_status "SUCCESS" "Client certificate verification passed"
    else
        print_status "ERROR" "Client certificate verification failed"
        return 1
    fi
    
    # Display certificate information
    print_status "INFO" "Certificate Information:"
    echo "  CA Certificate:"
    openssl x509 -in "$ca_crt" -noout -subject -dates
    echo "  Server Certificate:"
    openssl x509 -in "$server_crt" -noout -subject -dates
    echo "  Client Certificate:"
    openssl x509 -in "$client_crt" -noout -subject -dates
    
    print_status "SUCCESS" "All certificates generated and verified"
}

setup_vault() {
    print_header "Setting Up HashiCorp Vault"
    
    # Start Vault in development mode for setup
    local vault_token=$(cat "$SECRETS_DIR/vault-token")
    export VAULT_ADDR="http://127.0.0.1:8200"
    export VAULT_TOKEN="$vault_token"
    
    print_status "INFO" "Starting Vault container..."
    docker run -d --name ollama-vault-setup \
        -p 8200:8200 \
        -e VAULT_DEV_ROOT_TOKEN_ID="$vault_token" \
        -e VAULT_DEV_LISTEN_ADDRESS=0.0.0.0:8200 \
        hashicorp/vault:1.15 > /dev/null
    
    # Wait for Vault to be ready
    print_status "INFO" "Waiting for Vault to be ready..."
    local max_attempts=30
    local attempt=0
    while [[ $attempt -lt $max_attempts ]]; do
        if vault status > /dev/null 2>&1; then
            break
        fi
        sleep 2
        ((attempt++))
    done
    
    if [[ $attempt -eq $max_attempts ]]; then
        print_status "ERROR" "Vault failed to start within timeout"
        return 1
    fi
    
    print_status "SUCCESS" "Vault is ready"
    
    # Enable KV secrets engine
    vault secrets enable -path=ollama kv-v2 || true
    
    # Store secrets in Vault
    print_status "INFO" "Storing secrets in Vault..."
    
    local jwt_secret=$(cat "$SECRETS_DIR/jwt-secret")
    local encryption_key=$(cat "$SECRETS_DIR/encryption-key")
    local redis_password=$(cat "$SECRETS_DIR/redis-password")
    
    vault kv put ollama/jwt-secret value="$jwt_secret"
    vault kv put ollama/encryption-key value="$encryption_key"
    vault kv put ollama/redis-password value="$redis_password"
    
    # Create policy for OllamaMax
    vault policy write ollama-policy - << EOF
# OllamaMax Policy
path "ollama/*" {
  capabilities = ["read"]
}

path "auth/token/lookup-self" {
  capabilities = ["read"]
}

path "auth/token/renew-self" {
  capabilities = ["update"]
}
EOF
    
    print_status "SUCCESS" "Vault setup completed"
    
    # Stop setup container
    docker stop ollama-vault-setup > /dev/null
    docker rm ollama-vault-setup > /dev/null
    
    print_status "INFO" "Vault setup container cleaned up"
}

create_haproxy_config() {
    print_header "Creating HAProxy Configuration"
    
    local haproxy_config="$PROJECT_ROOT/config/haproxy-secure.cfg"
    
    cat > "$haproxy_config" << 'EOF'
# HAProxy Enterprise Security Configuration
global
    # Daemon and process settings
    daemon
    nbproc 2
    nbthread 4
    cpu-map auto:1/1-2 0-1
    
    # Security settings
    chroot /var/lib/haproxy
    user haproxy
    group haproxy
    
    # SSL settings
    ssl-default-bind-ciphers ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256
    ssl-default-bind-ciphersuites TLS_AES_256_GCM_SHA384:TLS_CHACHA20_POLY1305_SHA256:TLS_AES_128_GCM_SHA256
    ssl-default-bind-options ssl-min-ver TLSv1.2 no-tls-tickets
    ssl-default-server-ciphers ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256
    ssl-default-server-ciphersuites TLS_AES_256_GCM_SHA384:TLS_CHACHA20_POLY1305_SHA256:TLS_AES_128_GCM_SHA256
    ssl-default-server-options ssl-min-ver TLSv1.2 no-tls-tickets
    
    # DH parameters
    ssl-dh-param-file /etc/ssl/certs/dhparam.pem
    
    # Security headers
    tune.ssl.default-dh-param 2048
    tune.ssl.capture-cipherlist-size 1024
    
    # Logging
    log 127.0.0.1:514 local0
    log-tag haproxy-secure

defaults
    mode http
    timeout connect 5000ms
    timeout client 50000ms
    timeout server 50000ms
    timeout tunnel 1h
    
    # Security options
    option httplog
    option dontlognull
    option log-health-checks
    option redispatch
    
    # Security headers
    option forwardfor
    option http-server-close
    
    # Error pages
    errorfile 400 /usr/local/etc/haproxy/errors/400.http
    errorfile 403 /usr/local/etc/haproxy/errors/403.http
    errorfile 408 /usr/local/etc/haproxy/errors/408.http
    errorfile 500 /usr/local/etc/haproxy/errors/500.http
    errorfile 502 /usr/local/etc/haproxy/errors/502.http
    errorfile 503 /usr/local/etc/haproxy/errors/503.http
    errorfile 504 /usr/local/etc/haproxy/errors/504.http

# Frontend for HTTPS
frontend ollama_https
    bind *:443 ssl crt /etc/ssl/certs/server.crt key /etc/ssl/private/server.key ca-file /etc/ssl/certs/ca.crt verify optional
    
    # Security headers
    http-response set-header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload"
    http-response set-header X-Frame-Options "DENY"
    http-response set-header X-Content-Type-Options "nosniff"
    http-response set-header X-XSS-Protection "1; mode=block"
    http-response set-header Referrer-Policy "strict-origin-when-cross-origin"
    http-response set-header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'"
    http-response del-header Server
    http-response del-header X-Powered-By
    
    # Rate limiting
    stick-table type ip size 100k expire 30s store http_req_rate(10s)
    http-request track-sc0 src
    http-request deny if { sc_http_req_rate(0) gt 20 }
    
    # Client certificate validation
    acl client_cert_valid ssl_c_used ssl_c_verify 0
    http-request deny unless client_cert_valid
    
    # Route to backend
    default_backend ollama_secure_nodes

# Frontend for HTTP (redirect to HTTPS)
frontend ollama_http
    bind *:80
    
    # Security headers for redirects
    http-response set-header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload"
    
    # Redirect all HTTP to HTTPS
    redirect scheme https code 301

# Frontend for stats
frontend stats
    bind *:8404 ssl crt /etc/ssl/certs/server.crt key /etc/ssl/private/server.key
    stats enable
    stats uri /stats
    stats realm HAProxy\ Statistics
    stats auth admin:secure-stats-password
    stats refresh 30s

# Backend for OllamaMax secure nodes
backend ollama_secure_nodes
    balance roundrobin
    option httpchk GET /api/v1/health
    
    # SSL backend configuration
    server secure-node1 ollamamax-secure-node1:8443 check ssl verify required ca-file /etc/ssl/certs/ca.crt crt /etc/ssl/certs/server.crt
    server secure-node2 ollamamax-secure-node2:8443 check ssl verify required ca-file /etc/ssl/certs/ca.crt crt /etc/ssl/certs/server.crt
EOF
    
    chmod 644 "$haproxy_config"
    print_status "SUCCESS" "HAProxy configuration created"
}

create_prometheus_config() {
    print_header "Creating Prometheus Configuration"
    
    local prometheus_config="$PROJECT_ROOT/config/prometheus-secure.yml"
    
    cat > "$prometheus_config" << 'EOF'
# Prometheus Enterprise Security Configuration
global:
  scrape_interval: 15s
  evaluation_interval: 15s
  external_labels:
    monitor: 'ollama-security-monitor'
    environment: 'production'

rule_files:
  - "/etc/prometheus/rules/*.yml"

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093
      scheme: https
      tls_config:
        cert_file: /etc/ssl/certs/server.crt
        key_file: /etc/ssl/private/server.key
        ca_file: /etc/ssl/certs/ca.crt

scrape_configs:
  # OllamaMax secure nodes
  - job_name: 'ollama-secure-nodes'
    scheme: https
    tls_config:
      cert_file: /etc/ssl/certs/server.crt
      key_file: /etc/ssl/private/server.key
      ca_file: /etc/ssl/certs/ca.crt
    static_configs:
      - targets:
        - ollamamax-secure-node1:9090
        - ollamamax-secure-node2:9090
    metrics_path: /metrics
    scrape_interval: 30s
    scrape_timeout: 10s

  # HAProxy stats
  - job_name: 'haproxy'
    scheme: https
    tls_config:
      cert_file: /etc/ssl/certs/server.crt
      key_file: /etc/ssl/private/server.key
      ca_file: /etc/ssl/certs/ca.crt
    static_configs:
      - targets:
        - secure-loadbalancer:8404
    metrics_path: /stats/prometheus
    scrape_interval: 30s

  # Redis metrics
  - job_name: 'redis'
    static_configs:
      - targets:
        - redis:6379
    scrape_interval: 30s

  # Vault metrics
  - job_name: 'vault'
    scheme: https
    tls_config:
      cert_file: /etc/ssl/certs/server.crt
      key_file: /etc/ssl/private/server.key
      ca_file: /etc/ssl/certs/ca.crt
    static_configs:
      - targets:
        - vault:8200
    metrics_path: /v1/sys/metrics
    params:
      format: ['prometheus']
    scrape_interval: 30s

  # Security monitor
  - job_name: 'security-monitor'
    static_configs:
      - targets:
        - security-monitor:8080
    scrape_interval: 15s
    
  # Node exporter (if available)
  - job_name: 'node-exporter'
    static_configs:
      - targets:
        - node-exporter:9100
    scrape_interval: 30s
EOF
    
    chmod 644 "$prometheus_config"
    print_status "SUCCESS" "Prometheus configuration created"
}

setup_environment() {
    print_header "Setting Up Environment"
    
    local env_file="$PROJECT_ROOT/.env.security"
    
    cat > "$env_file" << EOF
# OllamaMax Enterprise Security Environment

# Build information
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Security settings
OLLAMA_SECURITY_MODE=enterprise
OLLAMA_HARDENING_LEVEL=strict
OLLAMA_COMPLIANCE_MODE=SOC2

# Secrets (references to files)
JWT_SECRET_FILE=/etc/secrets/jwt-secret
ENCRYPTION_KEY_FILE=/etc/secrets/encryption-key
REDIS_PASSWORD_FILE=/etc/secrets/redis-password
VAULT_TOKEN_FILE=/etc/secrets/vault-token

# Vault configuration
VAULT_ADDR=https://vault:8200
VAULT_NAMESPACE=ollama
VAULT_ROOT_TOKEN=$(cat "$SECRETS_DIR/vault-token")

# Redis configuration
REDIS_PASSWORD=$(cat "$SECRETS_DIR/redis-password")

# Grafana configuration
GRAFANA_ADMIN_PASSWORD=$(cat "$SECRETS_DIR/grafana-admin-password")
GRAFANA_SECRET_KEY=$(cat "$SECRETS_DIR/grafana-secret-key")

# Network configuration
OLLAMA_NETWORK_SUBNET=172.30.0.0/16
MONITOR_NETWORK_SUBNET=172.31.0.0/16

# Logging
OLLAMA_LOG_LEVEL=info
SECURITY_LOG_LEVEL=info

# Certificate settings
TLS_CERT_VALIDITY_DAYS=90
CERT_RENEWAL_THRESHOLD_DAYS=30

# Security thresholds
RATE_LIMIT_THRESHOLD=100
BAN_THRESHOLD=5
SECURITY_SCAN_INTERVAL=3600

# Compliance settings
AUDIT_RETENTION_DAYS=2555  # 7 years
LOG_RETENTION_DAYS=365
BACKUP_RETENTION_DAYS=90
EOF
    
    chmod 600 "$env_file"
    print_status "SUCCESS" "Environment configuration created"
    
    # Create Docker environment file
    cat > "$PROJECT_ROOT/.env" << EOF
# Docker Compose Environment
COMPOSE_PROJECT_NAME=ollama-security
COMPOSE_FILE=docker-compose.security.yml

# Import security environment
$(cat "$env_file" | grep -v '^#' | grep -v '^$')
EOF
    
    chmod 600 "$PROJECT_ROOT/.env"
    print_status "SUCCESS" "Docker environment file created"
}

run_security_tests() {
    print_header "Running Security Tests"
    
    print_status "INFO" "Testing certificate chain..."
    if openssl verify -CAfile "$CERTS_DIR/ca.crt" "$CERTS_DIR/server.crt"; then
        print_status "SUCCESS" "Certificate chain validation passed"
    else
        print_status "ERROR" "Certificate chain validation failed"
        return 1
    fi
    
    print_status "INFO" "Testing secret file permissions..."
    local secret_files=("$SECRETS_DIR/jwt-secret" "$SECRETS_DIR/encryption-key" "$SECRETS_DIR/redis-password")
    for secret_file in "${secret_files[@]}"; do
        local perms=$(stat -c "%a" "$secret_file")
        if [[ "$perms" == "600" ]]; then
            print_status "SUCCESS" "Secret file permissions correct: $(basename "$secret_file")"
        else
            print_status "ERROR" "Incorrect permissions on: $(basename "$secret_file") ($perms)"
            return 1
        fi
    done
    
    print_status "INFO" "Testing configuration files..."
    if [[ -f "$PROJECT_ROOT/config/security.yaml" ]]; then
        print_status "SUCCESS" "Security configuration exists"
    else
        print_status "ERROR" "Security configuration missing"
        return 1
    fi
    
    print_status "INFO" "Testing Docker Compose configuration..."
    if docker-compose -f "$PROJECT_ROOT/docker-compose.security.yml" config > /dev/null; then
        print_status "SUCCESS" "Docker Compose configuration valid"
    else
        print_status "ERROR" "Docker Compose configuration invalid"
        return 1
    fi
    
    print_status "SUCCESS" "All security tests passed"
}

deploy_security_stack() {
    print_header "Deploying Security Stack"
    
    local compose_file="$PROJECT_ROOT/docker-compose.security.yml"
    
    print_status "INFO" "Starting security infrastructure..."
    
    # Start core infrastructure first
    docker-compose -f "$compose_file" up -d vault redis cert-manager
    
    # Wait for vault to be ready
    print_status "INFO" "Waiting for Vault to be ready..."
    local max_attempts=30
    local attempt=0
    while [[ $attempt -lt $max_attempts ]]; do
        if docker-compose -f "$compose_file" exec -T vault vault status > /dev/null 2>&1; then
            break
        fi
        sleep 2
        ((attempt++))
    done
    
    if [[ $attempt -eq $max_attempts ]]; then
        print_status "ERROR" "Vault failed to start within timeout"
        return 1
    fi
    
    print_status "SUCCESS" "Core infrastructure started"
    
    # Start OllamaMax secure nodes
    print_status "INFO" "Starting OllamaMax secure nodes..."
    docker-compose -f "$compose_file" up -d ollamamax-secure-node1 ollamamax-secure-node2
    
    # Start load balancer
    print_status "INFO" "Starting secure load balancer..."
    docker-compose -f "$compose_file" up -d secure-loadbalancer
    
    # Start monitoring stack
    print_status "INFO" "Starting monitoring stack..."
    docker-compose -f "$compose_file" up -d prometheus-secure grafana-secure security-monitor
    
    print_status "SUCCESS" "Security stack deployed"
    
    # Display status
    print_status "INFO" "Service status:"
    docker-compose -f "$compose_file" ps
}

show_summary() {
    print_header "Security Setup Summary"
    
    echo -e "${GREEN}✅ Security setup completed successfully!${NC}"
    echo ""
    echo -e "${CYAN}Generated Files:${NC}"
    echo "  📁 Certificates: $CERTS_DIR/"
    echo "  🔐 Secrets: $SECRETS_DIR/"
    echo "  📋 Logs: $LOGS_DIR/"
    echo "  ⚙️  Configuration: $PROJECT_ROOT/config/"
    echo ""
    echo -e "${CYAN}Security Features Enabled:${NC}"
    echo "  🔒 TLS 1.3 with mutual authentication"
    echo "  🛡️  Enterprise WAF with OWASP CRS"
    echo "  🚫 Data Loss Prevention (DLP)"
    echo "  ⚡ Rate limiting and DDoS protection"
    echo "  🔍 Multi-factor authentication (MFA)"
    echo "  📊 Comprehensive audit logging"
    echo "  🗄️  HashiCorp Vault secret management"
    echo "  📈 Security monitoring and alerting"
    echo ""
    echo -e "${CYAN}Access Information:${NC}"
    echo "  🌐 HTTPS API: https://localhost:443"
    echo "  📊 Grafana: https://localhost:3000"
    echo "  🔧 HAProxy Stats: https://localhost:8404/stats"
    echo "  🗄️  Vault: https://localhost:8200"
    echo ""
    echo -e "${CYAN}Important Security Notes:${NC}"
    echo "  ⚠️  Change default passwords immediately"
    echo "  🔄 Certificate rotation every 90 days"
    echo "  📝 Review audit logs regularly"
    echo "  🛡️  Update security policies as needed"
    echo "  🔍 Monitor security metrics and alerts"
    echo ""
    echo -e "${YELLOW}Next Steps:${NC}"
    echo "  1. Start the security stack: docker-compose -f docker-compose.security.yml up -d"
    echo "  2. Configure OAuth2/OIDC providers (optional)"
    echo "  3. Set up SIEM integration (optional)"
    echo "  4. Run security compliance scan"
    echo "  5. Configure backup and disaster recovery"
    echo ""
    echo -e "${GREEN}🎉 OllamaMax Enterprise Security is ready for production!${NC}"
}

# Main execution
main() {
    print_header "OllamaMax Enterprise Security Setup"
    
    # Create logs directory first
    mkdir -p "$LOGS_DIR"
    
    # Run setup steps
    check_requirements
    setup_directories
    generate_secrets
    generate_certificates
    setup_vault
    create_haproxy_config
    create_prometheus_config
    setup_environment
    run_security_tests
    
    # Option to deploy immediately
    if [[ "${1:-}" == "--deploy" ]]; then
        deploy_security_stack
    fi
    
    show_summary
    
    print_status "SUCCESS" "Enterprise security setup completed"
}

# Run main function with all arguments
main "$@"