#!/bin/bash

# Ollamacron Universal Installation Script
# Detects platform and runs appropriate installer

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Repository information
REPO_URL="https://raw.githubusercontent.com/ollama-distributed/ollamacron/main/deploy/install"

# Logging functions
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

# Detect operating system
detect_os() {
    case "$(uname -s)" in
        Linux*)
            OS="linux"
            ;;
        Darwin*)
            OS="macos"
            ;;
        CYGWIN*|MINGW*|MSYS*)
            OS="windows"
            ;;
        *)
            error "Unsupported operating system: $(uname -s)"
            exit 1
            ;;
    esac
    
    info "Detected OS: $OS"
}

# Show banner
show_banner() {
    cat << 'EOF'
   _____ _ _                                         
  |  _  | | |                                        
  | | | | | | __ _ _ __ ___   __ _  ___ _ __ ___  _ __  
  | | | | | |/ _` | '_ ` _ \ / _` |/ __| '__/ _ \| '_ \ 
  \ \_/ / | | (_| | | | | | | (_| | (__| | | (_) | | | |
   \___/|_|_|\__,_|_| |_| |_|\__,_|\___|_|  \___/|_| |_|
                                                      
  Distributed AI Inference Platform
  Universal Installation Script
  
EOF
}

# Download and run platform-specific installer
install_platform() {
    local script_name
    local script_url
    local temp_script
    
    case $OS in
        linux)
            script_name="install.sh"
            script_url="$REPO_URL/linux/$script_name"
            ;;
        macos)
            script_name="install.sh"
            script_url="$REPO_URL/macos/$script_name"
            ;;
        windows)
            error "Windows installation requires PowerShell"
            error "Please run: irm https://raw.githubusercontent.com/ollama-distributed/ollamacron/main/deploy/install/windows/install.ps1 | iex"
            exit 1
            ;;
        *)
            error "Unsupported platform: $OS"
            exit 1
            ;;
    esac
    
    log "Downloading $OS installer..."
    
    # Create temporary file
    temp_script=$(mktemp)
    
    # Download installer
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$script_url" -o "$temp_script"
    elif command -v wget >/dev/null 2>&1; then
        wget -q "$script_url" -O "$temp_script"
    else
        error "Neither curl nor wget found. Please install one of them."
        exit 1
    fi
    
    # Make executable
    chmod +x "$temp_script"
    
    log "Running $OS installer..."
    
    # Pass all arguments to the platform-specific installer
    "$temp_script" "$@"
    
    # Clean up
    rm -f "$temp_script"
}

# Check prerequisites
check_prerequisites() {
    # Check if we have internet connectivity
    if ! ping -c 1 google.com >/dev/null 2>&1; then
        error "No internet connectivity detected"
        exit 1
    fi
    
    # Check if running as root on Linux (required)
    if [[ "$OS" == "linux" && $EUID -ne 0 ]]; then
        error "This script must be run as root on Linux (use sudo)"
        exit 1
    fi
    
    # Check if running as root on macOS (not recommended)
    if [[ "$OS" == "macos" && $EUID -eq 0 ]]; then
        error "This script should not be run as root on macOS"
        exit 1
    fi
}

# Show help
show_help() {
    cat << EOF
Ollamacron Universal Installer

Usage: $0 [options]

Options:
  --uninstall    Uninstall Ollamacron
  --version      Show installer version
  --help         Show this help message

Environment variables:
  OLLAMACRON_VERSION    Version to install (default: latest)

Platform-specific installers:
  Linux:   curl -fsSL https://raw.githubusercontent.com/ollama-distributed/ollamacron/main/deploy/install/linux/install.sh | sudo bash
  macOS:   curl -fsSL https://raw.githubusercontent.com/ollama-distributed/ollamacron/main/deploy/install/macos/install.sh | bash
  Windows: irm https://raw.githubusercontent.com/ollama-distributed/ollamacron/main/deploy/install/windows/install.ps1 | iex

Examples:
  # Install latest version
  curl -fsSL https://raw.githubusercontent.com/ollama-distributed/ollamacron/main/deploy/install/install.sh | bash

  # Install specific version
  OLLAMACRON_VERSION=v1.0.0 curl -fsSL https://raw.githubusercontent.com/ollama-distributed/ollamacron/main/deploy/install/install.sh | bash

  # Uninstall
  curl -fsSL https://raw.githubusercontent.com/ollama-distributed/ollamacron/main/deploy/install/install.sh | bash -s -- --uninstall

For more information, visit: https://github.com/ollama-distributed/ollamacron
EOF
}

# Main function
main() {
    # Show banner
    show_banner
    
    # Parse arguments
    case ${1:-} in
        --version)
            echo "Ollamacron Universal Installer v1.0.0"
            exit 0
            ;;
        --help)
            show_help
            exit 0
            ;;
    esac
    
    # Detect platform
    detect_os
    
    # Check prerequisites
    check_prerequisites
    
    # Install for the detected platform
    install_platform "$@"
}

# Run main function
main "$@"