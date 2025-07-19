#!/bin/bash

# Ollamacron Linux Installation Script
# Automated installation for Ubuntu/Debian, CentOS/RHEL, and Arch Linux

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
OLLAMACRON_VERSION="latest"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/ollamacron"
DATA_DIR="/var/lib/ollamacron"
LOG_DIR="/var/log/ollamacron"
SERVICE_USER="ollamacron"
SERVICE_GROUP="ollamacron"

# GitHub repository
GITHUB_REPO="ollama-distributed/ollamacron"
GITHUB_API_URL="https://api.github.com/repos/${GITHUB_REPO}"

# Logging function
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}" >&2
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

info() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] INFO: $1${NC}"
}

# Check if running as root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        error "This script must be run as root (use sudo)"
        exit 1
    fi
}

# Detect Linux distribution
detect_os() {
    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        OS=$ID
        VERSION=$VERSION_ID
    elif type lsb_release >/dev/null 2>&1; then
        OS=$(lsb_release -si | tr '[:upper:]' '[:lower:]')
        VERSION=$(lsb_release -sr)
    elif [[ -f /etc/redhat-release ]]; then
        OS="rhel"
        VERSION=$(grep -oE '[0-9]+\.[0-9]+' /etc/redhat-release)
    else
        error "Cannot detect Linux distribution"
        exit 1
    fi
    
    info "Detected OS: $OS $VERSION"
}

# Detect system architecture
detect_arch() {
    ARCH=$(uname -m)
    case $ARCH in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64)
            ARCH="arm64"
            ;;
        armv7l)
            ARCH="armv7"
            ;;
        *)
            error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac
    
    info "Detected architecture: $ARCH"
}

# Install dependencies based on distribution
install_dependencies() {
    log "Installing dependencies..."
    
    case $OS in
        ubuntu|debian)
            apt-get update
            apt-get install -y curl wget jq systemd docker.io docker-compose
            ;;
        centos|rhel|fedora)
            if command -v dnf &> /dev/null; then
                dnf install -y curl wget jq systemd docker docker-compose
            else
                yum install -y curl wget jq systemd docker docker-compose
            fi
            ;;
        arch)
            pacman -Sy --noconfirm curl wget jq systemd docker docker-compose
            ;;
        *)
            error "Unsupported distribution: $OS"
            exit 1
            ;;
    esac
    
    # Install Go if not present
    if ! command -v go &> /dev/null; then
        log "Installing Go..."
        install_go
    fi
    
    # Start and enable Docker
    systemctl start docker
    systemctl enable docker
    
    log "Dependencies installed successfully"
}

# Install Go programming language
install_go() {
    GO_VERSION="1.21.5"
    GO_TARBALL="go${GO_VERSION}.linux-${ARCH}.tar.gz"
    
    log "Installing Go $GO_VERSION..."
    
    # Download and extract Go
    cd /tmp
    wget -q "https://golang.org/dl/${GO_TARBALL}"
    tar -C /usr/local -xzf "${GO_TARBALL}"
    
    # Add Go to PATH
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    echo 'export GOPATH=/usr/local/go' >> /etc/profile
    
    # Create symlink for immediate use
    ln -sf /usr/local/go/bin/go /usr/local/bin/go
    
    log "Go installed successfully"
}

# Download Ollamacron binary
download_ollamacron() {
    log "Downloading Ollamacron..."
    
    # Get latest release info
    if [[ "$OLLAMACRON_VERSION" == "latest" ]]; then
        RELEASE_URL="${GITHUB_API_URL}/releases/latest"
        DOWNLOAD_URL=$(curl -s $RELEASE_URL | jq -r ".assets[] | select(.name | contains(\"linux-${ARCH}\")) | .browser_download_url")
    else
        DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download/${OLLAMACRON_VERSION}/ollamacron-linux-${ARCH}"
    fi
    
    if [[ -z "$DOWNLOAD_URL" ]]; then
        error "Could not find download URL for Ollamacron"
        exit 1
    fi
    
    # Download binary
    cd /tmp
    curl -L -o ollamacron "$DOWNLOAD_URL"
    chmod +x ollamacron
    
    # Install binary
    mv ollamacron "$INSTALL_DIR/ollamacron"
    
    log "Ollamacron binary installed to $INSTALL_DIR/ollamacron"
}

# Create system user and group
create_user() {
    log "Creating system user and group..."
    
    # Create group if it doesn't exist
    if ! getent group "$SERVICE_GROUP" &>/dev/null; then
        groupadd --system "$SERVICE_GROUP"
        log "Created group: $SERVICE_GROUP"
    fi
    
    # Create user if it doesn't exist
    if ! id "$SERVICE_USER" &>/dev/null; then
        useradd --system --gid "$SERVICE_GROUP" --no-create-home --shell /bin/false "$SERVICE_USER"
        log "Created user: $SERVICE_USER"
    fi
}

# Create directory structure
create_directories() {
    log "Creating directory structure..."
    
    # Create directories
    mkdir -p "$CONFIG_DIR" "$DATA_DIR" "$LOG_DIR"
    
    # Set ownership and permissions
    chown -R "$SERVICE_USER:$SERVICE_GROUP" "$CONFIG_DIR" "$DATA_DIR" "$LOG_DIR"
    chmod 755 "$CONFIG_DIR" "$DATA_DIR"
    chmod 750 "$LOG_DIR"
    
    log "Directory structure created"
}

# Create default configuration
create_config() {
    log "Creating default configuration..."
    
    cat > "$CONFIG_DIR/config.yaml" << EOF
# Ollamacron Configuration
server:
  bind: "0.0.0.0:8080"
  tls:
    enabled: false
    cert_file: ""
    key_file: ""

p2p:
  enabled: true
  listen_addr: "/ip4/0.0.0.0/tcp/9000"
  bootstrap_peers: []
  discovery:
    enabled: true
    rendezvous: "ollamacron-v1"

models:
  cache_dir: "$DATA_DIR/models"
  auto_pull: true
  sync_interval: "5m"

logging:
  level: "info"
  format: "json"
  output: "$LOG_DIR/ollamacron.log"

metrics:
  enabled: true
  bind: "0.0.0.0:9090"
  path: "/metrics"

health:
  enabled: true
  bind: "0.0.0.0:8081"
  path: "/health"
EOF
    
    chown "$SERVICE_USER:$SERVICE_GROUP" "$CONFIG_DIR/config.yaml"
    chmod 644 "$CONFIG_DIR/config.yaml"
    
    log "Default configuration created"
}

# Create systemd service
create_service() {
    log "Creating systemd service..."
    
    cat > /etc/systemd/system/ollamacron.service << EOF
[Unit]
Description=Ollamacron Distributed AI Inference Service
Documentation=https://github.com/ollama-distributed/ollamacron
After=network.target
Wants=network.target

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_GROUP
ExecStart=$INSTALL_DIR/ollamacron server --config $CONFIG_DIR/config.yaml
ExecReload=/bin/kill -HUP \$MAINPID
KillMode=mixed
Restart=always
RestartSec=5s

# Security settings
NoNewPrivileges=true
PrivateTmp=true
PrivateDevices=true
ProtectHome=true
ProtectSystem=strict
ReadWritePaths=$DATA_DIR $LOG_DIR
CapabilityBoundingSet=CAP_NET_BIND_SERVICE
AmbientCapabilities=CAP_NET_BIND_SERVICE

# Resource limits
LimitNOFILE=65536
LimitNPROC=65536

# Environment
Environment=PATH=/usr/local/bin:/usr/bin:/bin
Environment=OLLAMACRON_CONFIG=$CONFIG_DIR/config.yaml
Environment=OLLAMACRON_DATA_DIR=$DATA_DIR
Environment=OLLAMACRON_LOG_DIR=$LOG_DIR

[Install]
WantedBy=multi-user.target
EOF
    
    # Reload systemd
    systemctl daemon-reload
    
    log "Systemd service created"
}

# Install auto-update cron job
install_auto_update() {
    log "Installing auto-update mechanism..."
    
    cat > /etc/cron.d/ollamacron-update << EOF
# Ollamacron Auto-Update
# Check for updates daily at 2 AM
0 2 * * * root /usr/local/bin/ollamacron update --check-only && /usr/local/bin/ollamacron update --auto-restart
EOF
    
    log "Auto-update mechanism installed"
}

# Configure firewall
configure_firewall() {
    log "Configuring firewall..."
    
    # Check if ufw is available
    if command -v ufw &> /dev/null; then
        ufw --force enable
        ufw allow 8080/tcp  # API server
        ufw allow 9000/tcp  # P2P networking
        ufw allow 9090/tcp  # Metrics
        ufw allow 8081/tcp  # Health checks
        log "UFW firewall configured"
    # Check if firewalld is available
    elif command -v firewall-cmd &> /dev/null; then
        systemctl start firewalld
        systemctl enable firewalld
        firewall-cmd --permanent --add-port=8080/tcp
        firewall-cmd --permanent --add-port=9000/tcp
        firewall-cmd --permanent --add-port=9090/tcp
        firewall-cmd --permanent --add-port=8081/tcp
        firewall-cmd --reload
        log "Firewalld configured"
    else
        warn "No firewall detected. Please manually configure firewall rules."
    fi
}

# Verify installation
verify_installation() {
    log "Verifying installation..."
    
    # Check binary
    if [[ ! -f "$INSTALL_DIR/ollamacron" ]]; then
        error "Ollamacron binary not found"
        exit 1
    fi
    
    # Check version
    VERSION_OUTPUT=$($INSTALL_DIR/ollamacron version 2>/dev/null || echo "unknown")
    info "Installed version: $VERSION_OUTPUT"
    
    # Check service
    if ! systemctl is-enabled ollamacron &>/dev/null; then
        error "Ollamacron service not enabled"
        exit 1
    fi
    
    log "Installation verification completed"
}

# Main installation function
main() {
    log "Starting Ollamacron installation..."
    
    # Check prerequisites
    check_root
    detect_os
    detect_arch
    
    # Installation steps
    install_dependencies
    create_user
    create_directories
    download_ollamacron
    create_config
    create_service
    install_auto_update
    configure_firewall
    
    # Enable and start service
    systemctl enable ollamacron
    systemctl start ollamacron
    
    # Verify installation
    verify_installation
    
    log "Ollamacron installation completed successfully!"
    echo
    echo -e "${GREEN}🎉 Ollamacron is now installed and running!${NC}"
    echo
    echo -e "${BLUE}Next steps:${NC}"
    echo "  1. Check service status: systemctl status ollamacron"
    echo "  2. View logs: journalctl -u ollamacron -f"
    echo "  3. Test API: curl http://localhost:8080/health"
    echo "  4. Edit config: $CONFIG_DIR/config.yaml"
    echo "  5. View metrics: curl http://localhost:9090/metrics"
    echo
    echo -e "${BLUE}Documentation:${NC}"
    echo "  https://github.com/ollama-distributed/ollamacron/docs"
    echo
}

# Handle script arguments
case ${1:-} in
    --uninstall)
        log "Uninstalling Ollamacron..."
        systemctl stop ollamacron || true
        systemctl disable ollamacron || true
        rm -f /etc/systemd/system/ollamacron.service
        rm -f /etc/cron.d/ollamacron-update
        rm -f "$INSTALL_DIR/ollamacron"
        rm -rf "$CONFIG_DIR" "$DATA_DIR" "$LOG_DIR"
        userdel "$SERVICE_USER" || true
        groupdel "$SERVICE_GROUP" || true
        systemctl daemon-reload
        log "Ollamacron uninstalled"
        ;;
    --version)
        echo "Ollamacron Linux Installer v1.0.0"
        ;;
    --help)
        echo "Usage: $0 [options]"
        echo
        echo "Options:"
        echo "  --uninstall    Uninstall Ollamacron"
        echo "  --version      Show installer version"
        echo "  --help         Show this help message"
        echo
        echo "Environment variables:"
        echo "  OLLAMACRON_VERSION    Version to install (default: latest)"
        echo "  INSTALL_DIR          Installation directory (default: /usr/local/bin)"
        echo "  CONFIG_DIR           Configuration directory (default: /etc/ollamacron)"
        echo "  DATA_DIR             Data directory (default: /var/lib/ollamacron)"
        echo
        ;;
    *)
        main "$@"
        ;;
esac