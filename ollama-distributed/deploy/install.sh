#!/bin/bash

# Ollamacron Installation Script
# Installs Ollamacron as a system service

set -e

# Configuration
SERVICE_USER="ollama"
SERVICE_GROUP="ollama"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/ollamacron"
DATA_DIR="/var/lib/ollamacron"
LOG_DIR="/var/log/ollamacron"
SYSTEMD_DIR="/etc/systemd/system"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "This script must be run as root"
        exit 1
    fi
}

check_dependencies() {
    log_info "Checking dependencies..."
    
    # Check if systemd is available
    if ! command -v systemctl &> /dev/null; then
        log_error "systemd is required but not found"
        exit 1
    fi
    
    # Check if required directories exist
    if [[ ! -d "/etc/systemd/system" ]]; then
        log_error "systemd directory not found"
        exit 1
    fi
    
    log_info "Dependencies check passed"
}

create_user() {
    log_info "Creating service user..."
    
    if ! id "$SERVICE_USER" &>/dev/null; then
        useradd --system --shell /bin/false --home-dir "$DATA_DIR" --create-home "$SERVICE_USER"
        log_info "Created user: $SERVICE_USER"
    else
        log_info "User already exists: $SERVICE_USER"
    fi
}

create_directories() {
    log_info "Creating directories..."
    
    # Create directories
    mkdir -p "$CONFIG_DIR"
    mkdir -p "$DATA_DIR"
    mkdir -p "$LOG_DIR"
    
    # Set permissions
    chown -R "$SERVICE_USER:$SERVICE_GROUP" "$DATA_DIR"
    chown -R "$SERVICE_USER:$SERVICE_GROUP" "$LOG_DIR"
    chmod 755 "$CONFIG_DIR"
    chmod 750 "$DATA_DIR"
    chmod 750 "$LOG_DIR"
    
    log_info "Directories created and configured"
}

install_binary() {
    log_info "Installing binary..."
    
    # Find the binary
    if [[ -f "./build/ollamacron" ]]; then
        BINARY_PATH="./build/ollamacron"
    elif [[ -f "./ollamacron" ]]; then
        BINARY_PATH="./ollamacron"
    else
        log_error "Ollamacron binary not found. Please build it first."
        exit 1
    fi
    
    # Install binary
    cp "$BINARY_PATH" "$INSTALL_DIR/ollamacron"
    chmod +x "$INSTALL_DIR/ollamacron"
    
    log_info "Binary installed to: $INSTALL_DIR/ollamacron"
}

install_config() {
    log_info "Installing configuration..."
    
    if [[ -f "./config/config.yaml" ]]; then
        cp "./config/config.yaml" "$CONFIG_DIR/config.yaml"
        chmod 644 "$CONFIG_DIR/config.yaml"
        log_info "Configuration installed to: $CONFIG_DIR/config.yaml"
    else
        log_warn "Configuration file not found. Generating default config..."
        "$INSTALL_DIR/ollamacron" config generate "$CONFIG_DIR/config.yaml"
        chmod 644 "$CONFIG_DIR/config.yaml"
        log_info "Default configuration generated"
    fi
}

install_systemd_service() {
    log_info "Installing systemd service..."
    
    local service_type="${1:-node}"
    local service_file="ollamacron"
    
    if [[ "$service_type" == "coordinator" ]]; then
        service_file="ollamacron-coordinator"
    fi
    
    # Install service file
    if [[ -f "./deploy/systemd/${service_file}.service" ]]; then
        cp "./deploy/systemd/${service_file}.service" "$SYSTEMD_DIR/"
        chmod 644 "$SYSTEMD_DIR/${service_file}.service"
        log_info "Service file installed: $SYSTEMD_DIR/${service_file}.service"
    else
        log_error "Service file not found: ./deploy/systemd/${service_file}.service"
        exit 1
    fi
    
    # Reload systemd
    systemctl daemon-reload
    log_info "Systemd configuration reloaded"
}

enable_service() {
    local service_type="${1:-node}"
    local service_name="ollamacron"
    
    if [[ "$service_type" == "coordinator" ]]; then
        service_name="ollamacron-coordinator"
    fi
    
    log_info "Enabling service: $service_name"
    systemctl enable "$service_name"
    log_info "Service enabled: $service_name"
}

start_service() {
    local service_type="${1:-node}"
    local service_name="ollamacron"
    
    if [[ "$service_type" == "coordinator" ]]; then
        service_name="ollamacron-coordinator"
    fi
    
    log_info "Starting service: $service_name"
    systemctl start "$service_name"
    
    # Check status
    if systemctl is-active --quiet "$service_name"; then
        log_info "Service started successfully: $service_name"
    else
        log_error "Failed to start service: $service_name"
        systemctl status "$service_name"
        exit 1
    fi
}

show_status() {
    local service_type="${1:-node}"
    local service_name="ollamacron"
    
    if [[ "$service_type" == "coordinator" ]]; then
        service_name="ollamacron-coordinator"
    fi
    
    log_info "Service status:"
    systemctl status "$service_name" --no-pager
}

print_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --type TYPE     Service type (node|coordinator) [default: node]"
    echo "  --enable        Enable service after installation"
    echo "  --start         Start service after installation"
    echo "  --help          Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 --type node --enable --start"
    echo "  $0 --type coordinator --enable --start"
}

# Main installation function
main() {
    local service_type="node"
    local enable_service_flag=false
    local start_service_flag=false
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --type)
                service_type="$2"
                shift 2
                ;;
            --enable)
                enable_service_flag=true
                shift
                ;;
            --start)
                start_service_flag=true
                shift
                ;;
            --help)
                print_usage
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                print_usage
                exit 1
                ;;
        esac
    done
    
    # Validate service type
    if [[ "$service_type" != "node" && "$service_type" != "coordinator" ]]; then
        log_error "Invalid service type: $service_type"
        print_usage
        exit 1
    fi
    
    log_info "Starting Ollamacron installation..."
    log_info "Service type: $service_type"
    
    # Run installation steps
    check_root
    check_dependencies
    create_user
    create_directories
    install_binary
    install_config
    install_systemd_service "$service_type"
    
    # Optional steps
    if [[ "$enable_service_flag" == true ]]; then
        enable_service "$service_type"
    fi
    
    if [[ "$start_service_flag" == true ]]; then
        start_service "$service_type"
        sleep 2
        show_status "$service_type"
    fi
    
    log_info "Installation completed successfully!"
    log_info ""
    log_info "Configuration: $CONFIG_DIR/config.yaml"
    log_info "Data directory: $DATA_DIR"
    log_info "Log directory: $LOG_DIR"
    log_info ""
    log_info "Service management:"
    log_info "  Enable:  systemctl enable ollamacron"
    log_info "  Start:   systemctl start ollamacron"
    log_info "  Status:  systemctl status ollamacron"
    log_info "  Logs:    journalctl -u ollamacron -f"
}

# Run main function
main "$@"