#!/bin/bash

# GitHub Swarm Installation Script
# Installs and configures the GitHub Swarm command

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
INSTALL_DIR="$HOME/.claude-flow/commands/github-swarm"
BIN_DIR="$HOME/.local/bin"
COMMAND_NAME="github-swarm"

echo -e "${BLUE}🐙 GitHub Swarm Installation${NC}"
echo "=================================="

# Check prerequisites
echo -e "${YELLOW}📋 Checking prerequisites...${NC}"

# Check Node.js
if ! command -v node &> /dev/null; then
    echo -e "${RED}❌ Node.js is required but not installed${NC}"
    echo "Please install Node.js 14+ from https://nodejs.org/"
    exit 1
fi

NODE_VERSION=$(node --version | cut -d'v' -f2 | cut -d'.' -f1)
if [ "$NODE_VERSION" -lt 14 ]; then
    echo -e "${RED}❌ Node.js 14+ is required (found v$NODE_VERSION)${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Node.js $(node --version) found${NC}"

# Check npm
if ! command -v npm &> /dev/null; then
    echo -e "${RED}❌ npm is required but not installed${NC}"
    exit 1
fi

echo -e "${GREEN}✅ npm $(npm --version) found${NC}"

# Create directories
echo -e "${YELLOW}📁 Creating directories...${NC}"
mkdir -p "$INSTALL_DIR"
mkdir -p "$BIN_DIR"
mkdir -p "$INSTALL_DIR/agents"

# Copy files
echo -e "${YELLOW}📄 Installing files...${NC}"

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Copy main files
cp "$SCRIPT_DIR/index.js" "$INSTALL_DIR/"
cp "$SCRIPT_DIR/github-swarm-manager.js" "$INSTALL_DIR/"
cp "$SCRIPT_DIR/github-api-integration.js" "$INSTALL_DIR/"
cp "$SCRIPT_DIR/package.json" "$INSTALL_DIR/"
cp "$SCRIPT_DIR/README.md" "$INSTALL_DIR/"
cp "$SCRIPT_DIR/test-runner.js" "$INSTALL_DIR/"

# Copy agent files if they exist
if [ -d "$SCRIPT_DIR/agents" ]; then
    cp -r "$SCRIPT_DIR/agents/"* "$INSTALL_DIR/agents/" 2>/dev/null || true
fi

# Make main script executable
chmod +x "$INSTALL_DIR/index.js"

# Install dependencies
echo -e "${YELLOW}📦 Installing dependencies...${NC}"
cd "$INSTALL_DIR"
npm install --production

# Create symlink in bin directory
echo -e "${YELLOW}🔗 Creating command symlink...${NC}"
ln -sf "$INSTALL_DIR/index.js" "$BIN_DIR/$COMMAND_NAME"

# Check if bin directory is in PATH
if [[ ":$PATH:" != *":$BIN_DIR:"* ]]; then
    echo -e "${YELLOW}⚠️  Adding $BIN_DIR to PATH...${NC}"
    
    # Add to appropriate shell config
    if [ -f "$HOME/.bashrc" ]; then
        echo "export PATH=\"\$PATH:$BIN_DIR\"" >> "$HOME/.bashrc"
        echo -e "${GREEN}✅ Added to ~/.bashrc${NC}"
    fi
    
    if [ -f "$HOME/.zshrc" ]; then
        echo "export PATH=\"\$PATH:$BIN_DIR\"" >> "$HOME/.zshrc"
        echo -e "${GREEN}✅ Added to ~/.zshrc${NC}"
    fi
    
    # Export for current session
    export PATH="$PATH:$BIN_DIR"
fi

# Test installation
echo -e "${YELLOW}🧪 Testing installation...${NC}"
if command -v "$COMMAND_NAME" &> /dev/null; then
    echo -e "${GREEN}✅ Command '$COMMAND_NAME' is available${NC}"
else
    echo -e "${RED}❌ Command '$COMMAND_NAME' not found in PATH${NC}"
    echo "You may need to restart your shell or run: export PATH=\"\$PATH:$BIN_DIR\""
fi

# Run basic validation
echo -e "${YELLOW}🔍 Running validation tests...${NC}"
cd "$INSTALL_DIR"
if node test-runner.js > /dev/null 2>&1; then
    echo -e "${GREEN}✅ All tests passed${NC}"
else
    echo -e "${YELLOW}⚠️  Some tests failed (this is normal without GitHub token)${NC}"
fi

# Setup instructions
echo -e "\n${GREEN}🎉 Installation completed successfully!${NC}"
echo "=================================="
echo -e "${BLUE}Next steps:${NC}"
echo "1. Set up your GitHub token:"
echo "   export GITHUB_TOKEN=\"your_github_token\""
echo ""
echo "2. Test the installation:"
echo "   $COMMAND_NAME --help"
echo ""
echo "3. Run your first swarm:"
echo "   $COMMAND_NAME -r owner/repo -f maintenance"
echo ""
echo -e "${BLUE}Documentation:${NC}"
echo "- README: $INSTALL_DIR/README.md"
echo "- Tests: $COMMAND_NAME test"
echo ""
echo -e "${BLUE}GitHub Token Setup:${NC}"
echo "1. Go to GitHub Settings → Developer settings → Personal access tokens"
echo "2. Generate new token with 'repo' and 'read:org' scopes"
echo "3. Export the token: export GITHUB_TOKEN=\"your_token\""

# Check for GitHub token
if [ -z "$GITHUB_TOKEN" ]; then
    echo ""
    echo -e "${YELLOW}⚠️  No GITHUB_TOKEN found in environment${NC}"
    echo "Some features will be limited without GitHub API access"
fi

echo ""
echo -e "${GREEN}Installation complete! 🚀${NC}"
