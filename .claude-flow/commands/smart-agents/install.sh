#!/bin/bash

# Smart Agents Swarm Installation Script
# Sets up the hive-mind swarm system with all dependencies

set -e  # Exit on any error

echo "🚀 Installing Smart Agents Hive-Mind Swarm System..."

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   print_error "This script should not be run as root. Please run as a regular user."
   exit 1
fi

# Detect OS
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    OS="linux"
elif [[ "$OSTYPE" == "darwin"* ]]; then
    OS="macos"
else
    print_error "Unsupported OS: $OSTYPE"
    exit 1
fi

print_status "Detected OS: $OS"

# Check Node.js version
check_nodejs() {
    if command -v node &> /dev/null; then
        NODE_VERSION=$(node -v | cut -d'v' -f2)
        NODE_MAJOR=$(echo $NODE_VERSION | cut -d'.' -f1)
        
        if [[ $NODE_MAJOR -ge 16 ]]; then
            print_success "Node.js version $NODE_VERSION is supported"
        else
            print_error "Node.js version $NODE_VERSION is not supported. Please install Node.js 16 or higher."
            exit 1
        fi
    else
        print_error "Node.js is not installed. Please install Node.js 16 or higher."
        exit 1
    fi
}

# Check npm
check_npm() {
    if command -v npm &> /dev/null; then
        NPM_VERSION=$(npm -v)
        print_success "npm version $NPM_VERSION found"
    else
        print_error "npm is not installed. Please install npm."
        exit 1
    fi
}

# Create directory structure
setup_directory_structure() {
    print_status "Setting up directory structure..."
    
    BASE_DIR="/home/$USER/ollamamax/.claude-flow"
    
    # Create required directories
    mkdir -p "$BASE_DIR/memory"
    mkdir -p "$BASE_DIR/metrics"
    mkdir -p "$BASE_DIR/logs"
    mkdir -p "$BASE_DIR/agents/config"
    
    print_success "Directory structure created"
}

# Install package dependencies
install_dependencies() {
    print_status "Installing package dependencies..."
    
    SMART_AGENTS_DIR="/home/$USER/ollamamax/.claude-flow/commands/smart-agents"
    
    if [[ -f "$SMART_AGENTS_DIR/package.json" ]]; then
        cd "$SMART_AGENTS_DIR"
        npm install
        print_success "Dependencies installed successfully"
    else
        print_error "package.json not found in $SMART_AGENTS_DIR"
        exit 1
    fi
}

# Set executable permissions
set_permissions() {
    print_status "Setting executable permissions..."
    
    SMART_AGENTS_DIR="/home/$USER/ollamamax/.claude-flow/commands/smart-agents"
    
    chmod +x "$SMART_AGENTS_DIR/index.js"
    chmod +x "$SMART_AGENTS_DIR/install.sh"
    
    print_success "Permissions set successfully"
}

# Create global symlink
create_global_symlink() {
    print_status "Creating global symlink..."
    
    SMART_AGENTS_DIR="/home/$USER/ollamamax/.claude-flow/commands/smart-agents"
    SYMLINK_PATH="/usr/local/bin/smart-agents"
    
    # Check if symlink already exists
    if [[ -L "$SYMLINK_PATH" ]]; then
        print_warning "Symlink already exists at $SYMLINK_PATH"
        sudo rm "$SYMLINK_PATH"
    fi
    
    # Create symlink
    if sudo ln -s "$SMART_AGENTS_DIR/index.js" "$SYMLINK_PATH"; then
        print_success "Global symlink created at $SYMLINK_PATH"
    else
        print_warning "Failed to create global symlink. You can still run the command locally."
        print_status "To run locally: node $SMART_AGENTS_DIR/index.js"
    fi
}

# Initialize neural memory system
initialize_neural_memory() {
    print_status "Initializing neural memory system..."
    
    MEMORY_DIR="/home/$USER/ollamamax/.claude-flow/memory"
    
    # Create initial memory files if they don't exist
    if [[ ! -f "$MEMORY_DIR/neural-memory.json" ]]; then
        echo '{}' > "$MEMORY_DIR/neural-memory.json"
    fi
    
    if [[ ! -f "$MEMORY_DIR/performance-patterns.json" ]]; then
        echo '{}' > "$MEMORY_DIR/performance-patterns.json"
    fi
    
    if [[ ! -f "$MEMORY_DIR/adaptation-rules.json" ]]; then
        echo '{}' > "$MEMORY_DIR/adaptation-rules.json"
    fi
    
    print_success "Neural memory system initialized"
}

# Create configuration file
create_config() {
    print_status "Creating configuration file..."
    
    CONFIG_DIR="/home/$USER/ollamamax/.claude-flow/agents/config"
    CONFIG_FILE="$CONFIG_DIR/swarm-config.json"
    
    cat > "$CONFIG_FILE" << 'EOF'
{
  "swarm": {
    "maxAgents": 25,
    "minAgents": 8,
    "defaultTimeout": 120000,
    "scalingEnabled": true,
    "neuralLearningEnabled": true
  },
  "agents": {
    "concurrencyLimit": 15,
    "retryAttempts": 2,
    "healthCheckInterval": 10000
  },
  "neural": {
    "learningRate": 0.1,
    "memoryRetention": 1000,
    "adaptationThreshold": 0.8,
    "patternConfidenceThreshold": 0.7
  },
  "metrics": {
    "collectionInterval": 5000,
    "retentionDays": 30,
    "enableDetailedLogging": true
  },
  "sparc": {
    "enabled": true,
    "phaseTimeout": 300000,
    "parallelExecution": true
  }
}
EOF
    
    print_success "Configuration file created at $CONFIG_FILE"
}

# Verify installation
verify_installation() {
    print_status "Verifying installation..."
    
    # Test if command works
    if command -v smart-agents &> /dev/null; then
        print_success "Global command 'smart-agents' is available"
    else
        print_warning "Global command not available, testing local execution..."
        LOCAL_CMD="/home/$USER/ollamamax/.claude-flow/commands/smart-agents/index.js"
        if node "$LOCAL_CMD" status &> /dev/null; then
            print_success "Local execution works correctly"
        else
            print_error "Installation verification failed"
            exit 1
        fi
    fi
}

# Create desktop shortcut (optional)
create_desktop_shortcut() {
    if [[ "$OS" == "linux" ]] && [[ -d "$HOME/Desktop" ]]; then
        print_status "Creating desktop shortcut..."
        
        DESKTOP_FILE="$HOME/Desktop/smart-agents.desktop"
        
        cat > "$DESKTOP_FILE" << EOF
[Desktop Entry]
Version=1.0
Type=Application
Name=Smart Agents Swarm
Comment=Launch Smart Agents Hive-Mind Swarm
Exec=/usr/local/bin/smart-agents
Icon=applications-development
Terminal=true
Categories=Development;
EOF
        
        chmod +x "$DESKTOP_FILE"
        print_success "Desktop shortcut created"
    fi
}

# Performance optimization
optimize_system() {
    print_status "Applying system optimizations..."
    
    # Increase Node.js memory limit if needed
    NODE_OPTIONS_FILE="/home/$USER/.bashrc"
    
    if ! grep -q "NODE_OPTIONS" "$NODE_OPTIONS_FILE"; then
        echo 'export NODE_OPTIONS="--max-old-space-size=4096"' >> "$NODE_OPTIONS_FILE"
        print_success "Node.js memory limit optimized"
    fi
    
    # Create log rotation config
    LOGROTATE_CONFIG="/home/$USER/ollamamax/.claude-flow/logrotate.conf"
    
    cat > "$LOGROTATE_CONFIG" << 'EOF'
/home/$USER/ollamamax/.claude-flow/logs/*.log {
    daily
    missingok
    rotate 7
    compress
    delaycompress
    notifempty
    copytruncate
}
EOF
    
    print_success "Log rotation configured"
}

# Main installation process
main() {
    echo "=========================================="
    echo "Smart Agents Hive-Mind Swarm Installer"
    echo "=========================================="
    echo ""
    
    # Run installation steps
    check_nodejs
    check_npm
    setup_directory_structure
    install_dependencies
    set_permissions
    initialize_neural_memory
    create_config
    create_global_symlink
    create_desktop_shortcut
    optimize_system
    verify_installation
    
    echo ""
    echo "=========================================="
    print_success "Installation completed successfully!"
    echo "=========================================="
    echo ""
    
    # Display usage instructions
    echo -e "${BLUE}Usage Examples:${NC}"
    echo "  smart-agents execute \"build a distributed system\""
    echo "  smart-agents status"
    echo "  smart-agents metrics"
    echo "  smart-agents train"
    echo "  smart-agents scale 15"
    echo ""
    
    echo -e "${BLUE}Configuration:${NC}"
    echo "  Config file: ~/.claude-flow/agents/config/swarm-config.json"
    echo "  Memory data: ~/.claude-flow/memory/"
    echo "  Metrics: ~/.claude-flow/metrics/"
    echo "  Logs: ~/.claude-flow/logs/"
    echo ""
    
    echo -e "${BLUE}Getting Started:${NC}"
    echo "1. Try: smart-agents status"
    echo "2. Execute a task: smart-agents execute \"your task here\""
    echo "3. Monitor performance: smart-agents metrics"
    echo ""
    
    echo -e "${GREEN}🎉 Smart Agents Swarm is ready to unleash massively parallel AI development!${NC}"
}

# Run main installation
main "$@"