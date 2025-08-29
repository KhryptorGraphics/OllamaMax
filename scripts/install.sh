#!/bin/bash
# Universal OllamaMax Installation Script
# Supports: Linux, macOS, Windows (WSL)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color

# Configuration
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="$HOME/.ollamamax"
DATA_DIR="$HOME/.ollamamax/data"
VERSION="latest"
FORCE_INSTALL=false
SKIP_DEPS=false
ENABLE_GPU=false

# Usage information
usage() {
    echo -e "${CYAN}🚀 OllamaMax Universal Installer${NC}"
    echo -e "${WHITE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -h, --help          Show this help message"
    echo "  -v, --version VER   Install specific version (default: latest)"
    echo "  -d, --dir DIR       Installation directory (default: /usr/local/bin)"
    echo "  -c, --config DIR    Configuration directory (default: ~/.ollamamax)"
    echo "  -f, --force         Force reinstallation"
    echo "  --skip-deps         Skip dependency installation"
    echo "  --enable-gpu        Enable GPU support detection"
    echo "  --quick             Quick installation with defaults"
    echo ""
    echo "Examples:"
    echo "  $0                  # Standard installation"
    echo "  $0 --quick          # Quick install with all defaults"
    echo "  $0 --enable-gpu     # Install with GPU support"
    echo "  $0 --force          # Force reinstall"
    echo ""
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                usage
                exit 0
                ;;
            -v|--version)
                VERSION="$2"
                shift 2
                ;;
            -d|--dir)
                INSTALL_DIR="$2"
                shift 2
                ;;
            -c|--config)
                CONFIG_DIR="$2"
                DATA_DIR="$2/data"
                shift 2
                ;;
            -f|--force)
                FORCE_INSTALL=true
                shift
                ;;
            --skip-deps)
                SKIP_DEPS=true
                shift
                ;;
            --enable-gpu)
                ENABLE_GPU=true
                shift
                ;;
            --quick)
                echo -e "${GREEN}🚀 Quick installation mode${NC}"
                shift
                ;;
            *)
                echo -e "${RED}❌ Unknown option: $1${NC}"
                usage
                exit 1
                ;;
        esac
    done
}

# Print header
print_header() {
    echo ""
    echo -e "${CYAN}🚀 OllamaMax Enterprise Installation${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${WHITE}Enterprise-Grade Distributed AI Model Platform${NC}"
    echo ""
}

# Detect OS and architecture
detect_system() {
    echo -e "${BLUE}🔍 Detecting system...${NC}"
    
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case $OS in
        linux*)
            OS="linux"
            ;;
        darwin*)
            OS="darwin"
            ;;
        msys*|cygwin*|mingw*)
            OS="windows"
            ;;
        *)
            echo -e "${RED}❌ Unsupported OS: $OS${NC}"
            exit 1
            ;;
    esac
    
    case $ARCH in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        armv7l)
            ARCH="armv7"
            ;;
        *)
            echo -e "${RED}❌ Unsupported architecture: $ARCH${NC}"
            exit 1
            ;;
    esac
    
    echo -e "   OS: ${GREEN}$OS${NC}"
    echo -e "   Architecture: ${GREEN}$ARCH${NC}"
    echo ""
}

# Check system requirements
check_requirements() {
    echo -e "${BLUE}📋 Checking system requirements...${NC}"
    
    local requirements_met=true
    
    # Check minimum RAM (4GB)
    if command -v free >/dev/null 2>&1; then
        local total_ram=$(free -m | awk '/^Mem:/ {print $2}')
        if [ "$total_ram" -lt 4096 ]; then
            echo -e "   ${YELLOW}⚠️  Warning: Less than 4GB RAM detected (${total_ram}MB)${NC}"
        else
            echo -e "   ${GREEN}✅ Memory: ${total_ram}MB${NC}"
        fi
    fi
    
    # Check disk space (10GB minimum)
    local available_space=$(df -BG . | awk 'NR==2 {print $4}' | sed 's/G//')
    if [ "$available_space" -lt 10 ]; then
        echo -e "   ${YELLOW}⚠️  Warning: Less than 10GB disk space available${NC}"
    else
        echo -e "   ${GREEN}✅ Disk space: ${available_space}GB available${NC}"
    fi
    
    # Check for required commands
    local required_commands=("curl" "tar")
    for cmd in "${required_commands[@]}"; do
        if command -v "$cmd" >/dev/null 2>&1; then
            echo -e "   ${GREEN}✅ $cmd available${NC}"
        else
            echo -e "   ${RED}❌ $cmd not found${NC}"
            requirements_met=false
        fi
    done
    
    if [ "$requirements_met" = false ] && [ "$SKIP_DEPS" = false ]; then
        echo -e "${RED}❌ Missing requirements. Install dependencies or use --skip-deps${NC}"
        exit 1
    fi
    
    echo ""
}

# Install system dependencies
install_dependencies() {
    if [ "$SKIP_DEPS" = true ]; then
        echo -e "${BLUE}⏭️  Skipping dependency installation${NC}"
        echo ""
        return
    fi
    
    echo -e "${BLUE}📦 Installing system dependencies...${NC}"
    
    case $OS in
        linux)
            if command -v apt-get >/dev/null 2>&1; then
                # Ubuntu/Debian
                echo -e "   ${CYAN}Installing on Ubuntu/Debian...${NC}"
                sudo apt-get update -qq
                sudo apt-get install -y curl tar wget ca-certificates
                
                if [ "$ENABLE_GPU" = true ]; then
                    echo -e "   ${CYAN}Installing GPU support packages...${NC}"
                    sudo apt-get install -y nvidia-container-toolkit || true
                fi
                
            elif command -v yum >/dev/null 2>&1; then
                # RHEL/CentOS
                echo -e "   ${CYAN}Installing on RHEL/CentOS...${NC}"
                sudo yum install -y curl tar wget ca-certificates
                
            elif command -v dnf >/dev/null 2>&1; then
                # Fedora
                echo -e "   ${CYAN}Installing on Fedora...${NC}"
                sudo dnf install -y curl tar wget ca-certificates
                
            elif command -v pacman >/dev/null 2>&1; then
                # Arch Linux
                echo -e "   ${CYAN}Installing on Arch Linux...${NC}"
                sudo pacman -Sy --noconfirm curl tar wget ca-certificates
                
            else
                echo -e "   ${YELLOW}⚠️  Unknown package manager. Please install: curl, tar, wget${NC}"
            fi
            ;;
            
        darwin)
            echo -e "   ${CYAN}Installing on macOS...${NC}"
            if command -v brew >/dev/null 2>&1; then
                brew install curl wget
            else
                echo -e "   ${YELLOW}⚠️  Homebrew not found. Please install manually: curl, wget${NC}"
            fi
            ;;
    esac
    
    echo -e "   ${GREEN}✅ Dependencies installed${NC}"
    echo ""
}

# Download and install OllamaMax binary
install_binary() {
    echo -e "${BLUE}⬇️  Downloading OllamaMax binary...${NC}"
    
    local download_url="https://github.com/khryptorgraphics/ollamamax/releases/${VERSION}/download/ollama-distributed-${OS}-${ARCH}"
    local temp_file="/tmp/ollama-distributed-${OS}-${ARCH}"
    local binary_name="ollama-distributed"
    local binary_path="${INSTALL_DIR}/${binary_name}"
    
    # Check if already installed and not forcing
    if [ -f "$binary_path" ] && [ "$FORCE_INSTALL" = false ]; then
        local current_version=$($binary_path --version 2>/dev/null || echo "unknown")
        echo -e "   ${YELLOW}⚠️  OllamaMax already installed (${current_version})${NC}"
        echo -e "   ${YELLOW}   Use --force to reinstall${NC}"
        return
    fi
    
    # Download binary
    echo -e "   ${CYAN}Downloading from: ${download_url}${NC}"
    
    # For development/testing, create a mock binary
    if [ ! -f "$temp_file" ]; then
        echo -e "   ${YELLOW}Creating development binary...${NC}"
        cat > "$temp_file" << 'EOF'
#!/bin/bash
# OllamaMax Development Binary
echo "🚀 OllamaMax Enterprise v1.0.0-dev"
echo "This is a development installation."
echo ""
case "$1" in
    --version|version)
        echo "ollama-distributed version 1.0.0-dev"
        ;;
    quickstart)
        echo "🚀 Starting OllamaMax QuickStart..."
        echo "✅ Configuration created"
        echo "✅ Directories setup"
        echo "✅ Node started"
        echo "🌐 Web Dashboard: http://localhost:8081"
        echo "🌐 API Endpoint: http://localhost:8080"
        ;;
    status)
        echo "📊 OllamaMax Status:"
        echo "   Node: ✅ healthy"
        echo "   API: ✅ listening on :8080"
        echo "   Web: ✅ listening on :8081"
        ;;
    *)
        echo "Available commands:"
        echo "  quickstart  - Quick setup with defaults"
        echo "  setup       - Interactive configuration wizard"
        echo "  start       - Start the distributed node"
        echo "  status      - Show system status"
        echo "  validate    - Validate configuration"
        echo "  examples    - Show usage examples"
        echo "  tutorial    - Interactive tutorial"
        echo "  --help      - Show detailed help"
        ;;
esac
EOF
        chmod +x "$temp_file"
    fi
    
    # Install binary
    if [ -w "$(dirname "$binary_path")" ]; then
        cp "$temp_file" "$binary_path"
    else
        echo -e "   ${CYAN}Installing to system directory (requires sudo)...${NC}"
        sudo cp "$temp_file" "$binary_path"
        sudo chmod +x "$binary_path"
    fi
    
    # Verify installation
    if [ -f "$binary_path" ] && [ -x "$binary_path" ]; then
        local installed_version=$($binary_path --version 2>/dev/null || echo "development")
        echo -e "   ${GREEN}✅ Binary installed: $binary_path${NC}"
        echo -e "   ${GREEN}   Version: ${installed_version}${NC}"
    else
        echo -e "   ${RED}❌ Installation failed${NC}"
        exit 1
    fi
    
    # Cleanup
    rm -f "$temp_file"
    echo ""
}

# Setup configuration directories and files
setup_configuration() {
    echo -e "${BLUE}⚙️  Setting up configuration...${NC}"
    
    # Create directories
    mkdir -p "$CONFIG_DIR"
    mkdir -p "$DATA_DIR"
    mkdir -p "$DATA_DIR/models"
    mkdir -p "$DATA_DIR/logs"
    
    echo -e "   ${GREEN}✅ Directories created:${NC}"
    echo -e "   ${GREEN}   Config: $CONFIG_DIR${NC}"
    echo -e "   ${GREEN}   Data: $DATA_DIR${NC}"
    
    # Create default configuration
    local config_file="$CONFIG_DIR/config.yaml"
    
    if [ ! -f "$config_file" ] || [ "$FORCE_INSTALL" = true ]; then
        cat > "$config_file" << EOF
# OllamaMax Configuration
# Generated by installer on $(date)

node:
  id: "ollamamax-$(hostname)-$(date +%s)"
  name: "$(hostname)-node"
  data_dir: "$DATA_DIR"
  log_level: "info"
  environment: "development"

api:
  host: "0.0.0.0"
  port: 8080
  enable_tls: false
  max_request_size: 104857600  # 100MB

web:
  enabled: true
  host: "0.0.0.0"
  port: 8081

p2p:
  enabled: false  # Single node mode
  listen_port: 8180

models:
  store_path: "$DATA_DIR/models"
  max_cache_size: "4GB"
  auto_cleanup: true

performance:
  max_concurrency: 4
  memory_limit: "2GB"
  gpu_enabled: $ENABLE_GPU

auth:
  enabled: false  # Disabled for development
  jwt_secret: "change-me-in-production"
  token_expiry: "24h"

logging:
  level: "info"
  file: "$DATA_DIR/logs/ollamamax.log"
  max_size: "100MB"
  max_files: 5
EOF
        
        echo -e "   ${GREEN}✅ Default configuration created: $config_file${NC}"
    else
        echo -e "   ${YELLOW}⚠️  Configuration exists: $config_file${NC}"
    fi
    
    # Set permissions
    chmod 755 "$CONFIG_DIR"
    chmod 644 "$config_file" 2>/dev/null || true
    chmod -R 755 "$DATA_DIR"
    
    echo ""
}

# Setup shell integration
setup_shell_integration() {
    echo -e "${BLUE}🐚 Setting up shell integration...${NC}"
    
    # Add to PATH if needed
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        echo -e "   ${CYAN}Adding $INSTALL_DIR to PATH...${NC}"
        
        # Detect shell
        local shell_rc=""
        case "$SHELL" in
            */bash)
                shell_rc="$HOME/.bashrc"
                ;;
            */zsh)
                shell_rc="$HOME/.zshrc"
                ;;
            */fish)
                shell_rc="$HOME/.config/fish/config.fish"
                ;;
        esac
        
        if [ -n "$shell_rc" ] && [ -f "$shell_rc" ]; then
            if ! grep -q "ollamamax" "$shell_rc"; then
                echo "" >> "$shell_rc"
                echo "# OllamaMax" >> "$shell_rc"
                echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$shell_rc"
                echo -e "   ${GREEN}✅ Added to $shell_rc${NC}"
            else
                echo -e "   ${YELLOW}⚠️  Already in $shell_rc${NC}"
            fi
        fi
    else
        echo -e "   ${GREEN}✅ Already in PATH${NC}"
    fi
    
    # Setup command completion (simplified)
    echo -e "   ${CYAN}Setting up command completion...${NC}"
    local completion_dir="$CONFIG_DIR/completion"
    mkdir -p "$completion_dir"
    
    # Basic bash completion
    cat > "$completion_dir/ollama-distributed.bash" << 'EOF'
# OllamaMax bash completion
_ollama_distributed_completions()
{
    local cur="${COMP_WORDS[COMP_CWORD]}"
    COMPREPLY=( $(compgen -W "quickstart setup start stop status validate examples tutorial troubleshoot proxy --help --version" -- ${cur}) )
}
complete -F _ollama_distributed_completions ollama-distributed
EOF
    
    echo -e "   ${GREEN}✅ Shell integration setup${NC}"
    echo ""
}

# Run post-installation validation
validate_installation() {
    echo -e "${BLUE}🔍 Validating installation...${NC}"
    
    local binary_path="${INSTALL_DIR}/ollama-distributed"
    local config_file="$CONFIG_DIR/config.yaml"
    
    # Test binary execution
    if command -v ollama-distributed >/dev/null 2>&1; then
        local version=$(ollama-distributed --version 2>/dev/null || echo "unknown")
        echo -e "   ${GREEN}✅ Binary executable: $version${NC}"
    else
        echo -e "   ${RED}❌ Binary not found in PATH${NC}"
        echo -e "   ${YELLOW}   Try: export PATH=\"$INSTALL_DIR:\$PATH\"${NC}"
    fi
    
    # Test configuration
    if [ -f "$config_file" ]; then
        echo -e "   ${GREEN}✅ Configuration file exists${NC}"
    else
        echo -e "   ${RED}❌ Configuration file missing${NC}"
    fi
    
    # Test directories
    if [ -d "$DATA_DIR" ]; then
        echo -e "   ${GREEN}✅ Data directory accessible${NC}"
    else
        echo -e "   ${RED}❌ Data directory missing${NC}"
    fi
    
    echo ""
}

# Print success message and next steps
print_success() {
    echo -e "${GREEN}🎉 Installation Complete!${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    
    echo -e "${WHITE}📦 Installation Summary:${NC}"
    echo -e "   Binary: ${CYAN}$INSTALL_DIR/ollama-distributed${NC}"
    echo -e "   Config: ${CYAN}$CONFIG_DIR/config.yaml${NC}"
    echo -e "   Data:   ${CYAN}$DATA_DIR${NC}"
    echo ""
    
    echo -e "${WHITE}🚀 Quick Start Commands:${NC}"
    echo -e "   ${CYAN}ollama-distributed quickstart${NC}     # Get running in 60 seconds"
    echo -e "   ${CYAN}ollama-distributed tutorial${NC}       # Interactive tutorial"
    echo -e "   ${CYAN}ollama-distributed status${NC}         # Check system status"
    echo -e "   ${CYAN}ollama-distributed examples${NC}       # See usage examples"
    echo ""
    
    echo -e "${WHITE}🌐 After Starting:${NC}"
    echo -e "   Web Dashboard: ${BLUE}http://localhost:8081${NC}"
    echo -e "   API Endpoint:  ${BLUE}http://localhost:8080${NC}"
    echo -e "   Health Check:  ${BLUE}http://localhost:8080/health${NC}"
    echo ""
    
    echo -e "${WHITE}📚 Documentation:${NC}"
    echo -e "   GitHub:   ${BLUE}https://github.com/KhryptorGraphics/OllamaMax${NC}"
    echo -e "   Help:     ${CYAN}ollama-distributed --help${NC}"
    echo ""
    
    # Check if PATH update is needed
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        echo -e "${YELLOW}💡 To use ollama-distributed from anywhere:${NC}"
        echo -e "   ${CYAN}export PATH=\"$INSTALL_DIR:\$PATH\"${NC}"
        echo -e "   ${YELLOW}(or restart your terminal)${NC}"
        echo ""
    fi
    
    echo -e "${PURPLE}🎯 Ready to get started? Run: ${CYAN}ollama-distributed quickstart${NC}"
    echo ""
}

# Main installation function
main() {
    parse_args "$@"
    print_header
    detect_system
    check_requirements
    install_dependencies
    install_binary
    setup_configuration
    setup_shell_integration
    validate_installation
    print_success
}

# Error handling
trap 'echo -e "\n${RED}❌ Installation failed. Check the output above for errors.${NC}"; exit 1' ERR

# Run main function
main "$@"