#!/bin/bash

# Ollamacron macOS Installation Script
# Automated installation for macOS (Intel and Apple Silicon)

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
CONFIG_DIR="/usr/local/etc/ollamacron"
DATA_DIR="/usr/local/var/lib/ollamacron"
LOG_DIR="/usr/local/var/log/ollamacron"
SERVICE_USER="$(whoami)"
LAUNCH_AGENT_DIR="$HOME/Library/LaunchAgents"

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

# Check if running as root (not recommended on macOS)
check_not_root() {
    if [[ $EUID -eq 0 ]]; then
        error "This script should not be run as root on macOS"
        error "Please run as a regular user with sudo privileges"
        exit 1
    fi
}

# Detect system architecture
detect_arch() {
    ARCH=$(uname -m)
    case $ARCH in
        x86_64)
            ARCH="amd64"
            ;;
        arm64)
            ARCH="arm64"
            ;;
        *)
            error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac
    
    info "Detected architecture: $ARCH"
}

# Check macOS version
check_macos_version() {
    MACOS_VERSION=$(sw_vers -productVersion)
    MACOS_MAJOR=$(echo "$MACOS_VERSION" | cut -d. -f1)
    MACOS_MINOR=$(echo "$MACOS_VERSION" | cut -d. -f2)
    
    if [[ $MACOS_MAJOR -lt 11 ]]; then
        error "macOS 11.0 (Big Sur) or later is required"
        exit 1
    fi
    
    info "macOS version: $MACOS_VERSION"
}

# Install Homebrew if not present
install_homebrew() {
    if ! command -v brew &> /dev/null; then
        log "Installing Homebrew..."
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
        
        # Add Homebrew to PATH
        if [[ "$ARCH" == "arm64" ]]; then
            echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zprofile
            eval "$(/opt/homebrew/bin/brew shellenv)"
        else
            echo 'eval "$(/usr/local/bin/brew shellenv)"' >> ~/.zprofile
            eval "$(/usr/local/bin/brew shellenv)"
        fi
        
        log "Homebrew installed successfully"
    else
        log "Homebrew already installed"
    fi
}

# Install dependencies
install_dependencies() {
    log "Installing dependencies..."
    
    # Update Homebrew
    brew update
    
    # Install required packages
    brew install curl wget jq go docker docker-compose
    
    # Install Docker Desktop if not present
    if ! command -v docker &> /dev/null; then
        log "Installing Docker Desktop..."
        brew install --cask docker
        
        warn "Docker Desktop installed. Please start Docker Desktop manually."
        warn "You can find it in Applications or use Spotlight search."
    fi
    
    log "Dependencies installed successfully"
}

# Download Ollamacron binary
download_ollamacron() {
    log "Downloading Ollamacron..."
    
    # Get latest release info
    if [[ "$OLLAMACRON_VERSION" == "latest" ]]; then
        RELEASE_URL="${GITHUB_API_URL}/releases/latest"
        DOWNLOAD_URL=$(curl -s $RELEASE_URL | jq -r ".assets[] | select(.name | contains(\"darwin-${ARCH}\")) | .browser_download_url")
    else
        DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download/${OLLAMACRON_VERSION}/ollamacron-darwin-${ARCH}"
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
    sudo mkdir -p "$INSTALL_DIR"
    sudo mv ollamacron "$INSTALL_DIR/ollamacron"
    
    log "Ollamacron binary installed to $INSTALL_DIR/ollamacron"
}

# Create directory structure
create_directories() {
    log "Creating directory structure..."
    
    # Create directories
    sudo mkdir -p "$CONFIG_DIR" "$DATA_DIR" "$LOG_DIR"
    
    # Set ownership and permissions
    sudo chown -R "$SERVICE_USER:staff" "$CONFIG_DIR" "$DATA_DIR" "$LOG_DIR"
    sudo chmod 755 "$CONFIG_DIR" "$DATA_DIR"
    sudo chmod 750 "$LOG_DIR"
    
    log "Directory structure created"
}

# Create default configuration
create_config() {
    log "Creating default configuration..."
    
    sudo tee "$CONFIG_DIR/config.yaml" > /dev/null << EOF
# Ollamacron Configuration
server:
  bind: "127.0.0.1:8080"
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
  bind: "127.0.0.1:9090"
  path: "/metrics"

health:
  enabled: true
  bind: "127.0.0.1:8081"
  path: "/health"
EOF
    
    sudo chown "$SERVICE_USER:staff" "$CONFIG_DIR/config.yaml"
    sudo chmod 644 "$CONFIG_DIR/config.yaml"
    
    log "Default configuration created"
}

# Create LaunchAgent for auto-start
create_launch_agent() {
    log "Creating LaunchAgent..."
    
    mkdir -p "$LAUNCH_AGENT_DIR"
    
    cat > "$LAUNCH_AGENT_DIR/com.ollamacron.agent.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.ollamacron.agent</string>
    <key>ProgramArguments</key>
    <array>
        <string>$INSTALL_DIR/ollamacron</string>
        <string>server</string>
        <string>--config</string>
        <string>$CONFIG_DIR/config.yaml</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>$LOG_DIR/ollamacron.out.log</string>
    <key>StandardErrorPath</key>
    <string>$LOG_DIR/ollamacron.err.log</string>
    <key>EnvironmentVariables</key>
    <dict>
        <key>PATH</key>
        <string>/usr/local/bin:/usr/bin:/bin</string>
        <key>OLLAMACRON_CONFIG</key>
        <string>$CONFIG_DIR/config.yaml</string>
        <key>OLLAMACRON_DATA_DIR</key>
        <string>$DATA_DIR</string>
        <key>OLLAMACRON_LOG_DIR</key>
        <string>$LOG_DIR</string>
    </dict>
    <key>ProcessType</key>
    <string>Interactive</string>
    <key>ThrottleInterval</key>
    <integer>10</integer>
</dict>
</plist>
EOF
    
    log "LaunchAgent created"
}

# Install auto-update mechanism
install_auto_update() {
    log "Installing auto-update mechanism..."
    
    # Create update script
    sudo tee /usr/local/bin/ollamacron-update << 'EOF'
#!/bin/bash
# Ollamacron Auto-Update Script for macOS

CURRENT_VERSION=$(/usr/local/bin/ollamacron version 2>/dev/null | grep -o 'v[0-9]\+\.[0-9]\+\.[0-9]\+' || echo "unknown")
LATEST_VERSION=$(curl -s https://api.github.com/repos/ollama-distributed/ollamacron/releases/latest | jq -r .tag_name)

if [[ "$CURRENT_VERSION" != "$LATEST_VERSION" ]]; then
    echo "Update available: $CURRENT_VERSION -> $LATEST_VERSION"
    
    # Download and install new version
    ARCH=$(uname -m)
    case $ARCH in
        x86_64) ARCH="amd64" ;;
        arm64) ARCH="arm64" ;;
    esac
    
    cd /tmp
    curl -L -o ollamacron-new "https://github.com/ollama-distributed/ollamacron/releases/download/$LATEST_VERSION/ollamacron-darwin-$ARCH"
    chmod +x ollamacron-new
    
    # Stop service
    launchctl unload ~/Library/LaunchAgents/com.ollamacron.agent.plist
    
    # Replace binary
    sudo mv ollamacron-new /usr/local/bin/ollamacron
    
    # Restart service
    launchctl load ~/Library/LaunchAgents/com.ollamacron.agent.plist
    
    echo "Updated to $LATEST_VERSION"
else
    echo "Already up to date: $CURRENT_VERSION"
fi
EOF
    
    sudo chmod +x /usr/local/bin/ollamacron-update
    
    # Create LaunchAgent for auto-update
    cat > "$LAUNCH_AGENT_DIR/com.ollamacron.update.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.ollamacron.update</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/ollamacron-update</string>
    </array>
    <key>StartCalendarInterval</key>
    <dict>
        <key>Hour</key>
        <integer>2</integer>
        <key>Minute</key>
        <integer>0</integer>
    </dict>
    <key>StandardOutPath</key>
    <string>$LOG_DIR/ollamacron-update.log</string>
    <key>StandardErrorPath</key>
    <string>$LOG_DIR/ollamacron-update.log</string>
</dict>
</plist>
EOF
    
    log "Auto-update mechanism installed"
}

# Configure macOS firewall
configure_firewall() {
    log "Configuring macOS firewall..."
    
    # Add firewall rules (requires user interaction)
    warn "Please allow Ollamacron through the firewall when prompted"
    
    # The firewall rules will be added automatically when the service starts
    # and macOS prompts the user for permission
    
    log "Firewall configuration noted"
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
    
    # Check LaunchAgent
    if [[ ! -f "$LAUNCH_AGENT_DIR/com.ollamacron.agent.plist" ]]; then
        error "LaunchAgent not found"
        exit 1
    fi
    
    log "Installation verification completed"
}

# Main installation function
main() {
    log "Starting Ollamacron installation..."
    
    # Check prerequisites
    check_not_root
    check_macos_version
    detect_arch
    
    # Installation steps
    install_homebrew
    install_dependencies
    create_directories
    download_ollamacron
    create_config
    create_launch_agent
    install_auto_update
    configure_firewall
    
    # Load and start LaunchAgent
    launchctl load "$LAUNCH_AGENT_DIR/com.ollamacron.agent.plist"
    launchctl load "$LAUNCH_AGENT_DIR/com.ollamacron.update.plist"
    
    # Verify installation
    verify_installation
    
    log "Ollamacron installation completed successfully!"
    echo
    echo -e "${GREEN}🎉 Ollamacron is now installed and running!${NC}"
    echo
    echo -e "${BLUE}Next steps:${NC}"
    echo "  1. Check service status: launchctl list com.ollamacron.agent"
    echo "  2. View logs: tail -f $LOG_DIR/ollamacron.out.log"
    echo "  3. Test API: curl http://localhost:8080/health"
    echo "  4. Edit config: $CONFIG_DIR/config.yaml"
    echo "  5. View metrics: curl http://localhost:9090/metrics"
    echo
    echo -e "${BLUE}Service management:${NC}"
    echo "  Start:   launchctl load $LAUNCH_AGENT_DIR/com.ollamacron.agent.plist"
    echo "  Stop:    launchctl unload $LAUNCH_AGENT_DIR/com.ollamacron.agent.plist"
    echo "  Restart: launchctl unload $LAUNCH_AGENT_DIR/com.ollamacron.agent.plist && launchctl load $LAUNCH_AGENT_DIR/com.ollamacron.agent.plist"
    echo
    echo -e "${BLUE}Documentation:${NC}"
    echo "  https://github.com/ollama-distributed/ollamacron/docs"
    echo
}

# Handle script arguments
case ${1:-} in
    --uninstall)
        log "Uninstalling Ollamacron..."
        launchctl unload "$LAUNCH_AGENT_DIR/com.ollamacron.agent.plist" 2>/dev/null || true
        launchctl unload "$LAUNCH_AGENT_DIR/com.ollamacron.update.plist" 2>/dev/null || true
        rm -f "$LAUNCH_AGENT_DIR/com.ollamacron.agent.plist"
        rm -f "$LAUNCH_AGENT_DIR/com.ollamacron.update.plist"
        sudo rm -f "$INSTALL_DIR/ollamacron"
        sudo rm -f /usr/local/bin/ollamacron-update
        sudo rm -rf "$CONFIG_DIR" "$DATA_DIR" "$LOG_DIR"
        log "Ollamacron uninstalled"
        ;;
    --version)
        echo "Ollamacron macOS Installer v1.0.0"
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
        echo "  CONFIG_DIR           Configuration directory (default: /usr/local/etc/ollamacron)"
        echo "  DATA_DIR             Data directory (default: /usr/local/var/lib/ollamacron)"
        echo
        ;;
    *)
        main "$@"
        ;;
esac