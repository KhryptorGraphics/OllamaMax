# Installation Guide

This comprehensive guide covers all installation methods for Ollama Distributed, from quick setup to custom builds.

## Prerequisites

### System Requirements

#### Minimum Requirements
| Component | Requirement |
|-----------|------------|
| **Operating System** | Linux (Ubuntu 18.04+), macOS (10.15+), Windows 10+ with WSL2 |
| **CPU** | 2 cores (x86_64 or ARM64) |
| **Memory** | 4GB RAM available |
| **Storage** | 20GB free disk space |
| **Network** | Internet connection for downloads |

#### Recommended for Production
| Component | Recommendation |
|-----------|---------------|
| **Operating System** | Linux (Ubuntu 20.04+ or CentOS 8+) |
| **CPU** | 8+ cores with AVX2 support |
| **Memory** | 16GB+ RAM (32GB for large models) |
| **Storage** | 100GB+ NVMe SSD |
| **Network** | Gigabit Ethernet, low latency |
| **GPU** | NVIDIA GPU with 8GB+ VRAM (optional) |

### Dependencies

Ollama Distributed will automatically install required dependencies, but you can install them manually:

#### Ubuntu/Debian
```bash
sudo apt update
sudo apt install -y curl wget tar gzip ca-certificates
```

#### CentOS/RHEL/Fedora
```bash
sudo dnf install -y curl wget tar gzip ca-certificates
# Or for older systems:
sudo yum install -y curl wget tar gzip ca-certificates
```

#### macOS
```bash
# Using Homebrew
brew install curl wget

# Or using MacPorts
sudo port install curl wget
```

## Installation Methods

### Method 1: Quick Install Script (Recommended)

The fastest way to get started:

```bash
# Download and run the installer
curl -fsSL https://install.ollamamax.com/install.sh | bash

# Or download first, then run (more secure)
curl -fsSL https://install.ollamamax.com/install.sh -o install.sh
chmod +x install.sh
./install.sh
```

#### What the installer does:
1. Detects your operating system and architecture
2. Downloads the appropriate binary
3. Installs to `/usr/local/bin/ollama-distributed`
4. Creates configuration directory `~/.ollama-distributed/`
5. Sets up systemd service (Linux) or launchd (macOS)
6. Adds shell completion

#### Installation options:
```bash
# Install to custom directory
curl -fsSL https://install.ollamamax.com/install.sh | bash -s -- --prefix=/opt/ollama

# Install specific version
curl -fsSL https://install.ollamamax.com/install.sh | bash -s -- --version=v0.2.0

# Skip service setup
curl -fsSL https://install.ollamamax.com/install.sh | bash -s -- --no-service

# Install with GPU support
curl -fsSL https://install.ollamamax.com/install.sh | bash -s -- --gpu
```

### Method 2: Package Managers

#### Homebrew (macOS/Linux)
```bash
# Add the tap
brew tap ollamamax/tap

# Install Ollama Distributed
brew install ollama-distributed

# Start the service
brew services start ollama-distributed
```

#### APT (Ubuntu/Debian)
```bash
# Add repository
curl -fsSL https://packages.ollamamax.com/gpg.key | sudo apt-key add -
echo "deb https://packages.ollamamax.com/apt stable main" | sudo tee /etc/apt/sources.list.d/ollamamax.list

# Install
sudo apt update
sudo apt install ollama-distributed
```

#### YUM/DNF (CentOS/RHEL/Fedora)
```bash
# Add repository
sudo tee /etc/yum.repos.d/ollamamax.repo <<EOF
[ollamamax]
name=OllamaMax Repository
baseurl=https://packages.ollamamax.com/rpm
enabled=1
gpgcheck=1
gpgkey=https://packages.ollamamax.com/gpg.key
EOF

# Install
sudo dnf install ollama-distributed
# Or: sudo yum install ollama-distributed
```

### Method 3: Docker

#### Quick Start with Docker
```bash
# Run single container
docker run -d --name ollama-distributed \
  -p 8081:8081 \
  -v ollama-data:/data \
  ollamamax/ollama-distributed:latest

# Or use Docker Compose
curl -O https://raw.githubusercontent.com/ollamamax/ollama-distributed/main/docker-compose.yml
docker-compose up -d
```

#### Docker Compose for Production
```yaml
version: '3.8'
services:
  ollama-node1:
    image: ollamamax/ollama-distributed:latest
    container_name: ollama-node1
    ports:
      - "8081:8081"
    environment:
      - OLLAMA_CLUSTER_NODE_ID=node1
      - OLLAMA_CLUSTER_BOOTSTRAP=true
    volumes:
      - node1-data:/data
    networks:
      - ollama-cluster

  ollama-node2:
    image: ollamamax/ollama-distributed:latest
    container_name: ollama-node2
    ports:
      - "8082:8081"
    environment:
      - OLLAMA_CLUSTER_NODE_ID=node2
      - OLLAMA_CLUSTER_BOOTSTRAP_NODES=ollama-node1:7946
    volumes:
      - node2-data:/data
    networks:
      - ollama-cluster
    depends_on:
      - ollama-node1

  ollama-node3:
    image: ollamamax/ollama-distributed:latest
    container_name: ollama-node3
    ports:
      - "8083:8081"
    environment:
      - OLLAMA_CLUSTER_NODE_ID=node3
      - OLLAMA_CLUSTER_BOOTSTRAP_NODES=ollama-node1:7946
    volumes:
      - node3-data:/data
    networks:
      - ollama-cluster
    depends_on:
      - ollama-node1

volumes:
  node1-data:
  node2-data:
  node3-data:

networks:
  ollama-cluster:
    driver: bridge
```

### Method 4: Binary Download

For manual installation:

```bash
# Download for your platform
# Linux x86_64
wget https://github.com/ollamamax/ollama-distributed/releases/latest/download/ollama-distributed-linux-amd64.tar.gz

# Linux ARM64
wget https://github.com/ollamamax/ollama-distributed/releases/latest/download/ollama-distributed-linux-arm64.tar.gz

# macOS Intel
wget https://github.com/ollamamax/ollama-distributed/releases/latest/download/ollama-distributed-darwin-amd64.tar.gz

# macOS Apple Silicon
wget https://github.com/ollamamax/ollama-distributed/releases/latest/download/ollama-distributed-darwin-arm64.tar.gz

# Extract and install
tar -xzf ollama-distributed-*.tar.gz
sudo mv ollama-distributed /usr/local/bin/
sudo chmod +x /usr/local/bin/ollama-distributed
```

### Method 5: Build from Source

For developers and advanced users:

```bash
# Prerequisites
# Go 1.21+, Node.js 18+, npm/yarn

# Clone repository
git clone https://github.com/ollamamax/ollama-distributed.git
cd ollama-distributed

# Build backend
go mod download
go build -o bin/ollama-distributed ./cmd/ollama-distributed

# Build web dashboard (optional)
cd web
npm install
npm run build
cd ..

# Install
sudo cp bin/ollama-distributed /usr/local/bin/
sudo chmod +x /usr/local/bin/ollama-distributed
```

#### Build Options
```bash
# Build with GPU support
CGO_ENABLED=1 go build -tags gpu -o bin/ollama-distributed ./cmd/ollama-distributed

# Build for different architecture
GOOS=linux GOARCH=arm64 go build -o bin/ollama-distributed-arm64 ./cmd/ollama-distributed

# Build with debug symbols
go build -gcflags="all=-N -l" -o bin/ollama-distributed-debug ./cmd/ollama-distributed
```

## Post-Installation Setup

### Verify Installation

```bash
# Check version
ollama-distributed version

# Validate installation
ollama-distributed validate --installation

# Test basic functionality
ollama-distributed quickstart --dry-run
```

### Shell Completion

Enable command completion for your shell:

#### Bash
```bash
# Add to ~/.bashrc
echo 'source <(ollama-distributed completion bash)' >> ~/.bashrc
source ~/.bashrc
```

#### Zsh
```bash
# Add to ~/.zshrc
echo 'source <(ollama-distributed completion zsh)' >> ~/.zshrc
source ~/.zshrc
```

#### Fish
```bash
# Add to ~/.config/fish/config.fish
echo 'ollama-distributed completion fish | source' >> ~/.config/fish/config.fish
```

### System Service Setup

#### Systemd (Linux)
```bash
# Create service file
sudo tee /etc/systemd/system/ollama-distributed.service <<EOF
[Unit]
Description=Ollama Distributed
After=network.target

[Service]
Type=simple
User=ollama
Group=ollama
ExecStart=/usr/local/bin/ollama-distributed start --config /etc/ollama-distributed/config.yaml
Restart=always
RestartSec=5
Environment=OLLAMA_HOST=0.0.0.0:8081

[Install]
WantedBy=multi-user.target
EOF

# Create user and directories
sudo useradd -r -s /bin/false ollama
sudo mkdir -p /etc/ollama-distributed /var/lib/ollama-distributed
sudo chown ollama:ollama /var/lib/ollama-distributed

# Enable and start
sudo systemctl daemon-reload
sudo systemctl enable ollama-distributed
sudo systemctl start ollama-distributed
```

#### Launchd (macOS)
```bash
# Create service file
sudo tee /Library/LaunchDaemons/com.ollamamax.distributed.plist <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.ollamamax.distributed</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/ollama-distributed</string>
        <string>start</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
</dict>
</plist>
EOF

# Load service
sudo launchctl load /Library/LaunchDaemons/com.ollamamax.distributed.plist
```

## Uninstallation

### Clean Removal
```bash
# Stop service
sudo systemctl stop ollama-distributed
sudo systemctl disable ollama-distributed

# Remove binary
sudo rm /usr/local/bin/ollama-distributed

# Remove configuration and data
rm -rf ~/.ollama-distributed
sudo rm -rf /etc/ollama-distributed /var/lib/ollama-distributed

# Remove service file
sudo rm /etc/systemd/system/ollama-distributed.service
sudo systemctl daemon-reload
```

### Package Manager Removal
```bash
# Homebrew
brew uninstall ollama-distributed
brew untap ollamamax/tap

# APT
sudo apt remove ollama-distributed
sudo rm /etc/apt/sources.list.d/ollamamax.list

# YUM/DNF
sudo dnf remove ollama-distributed
sudo rm /etc/yum.repos.d/ollamamax.repo
```

## Troubleshooting Installation

### Common Issues

#### Permission Denied
```bash
# Make binary executable
sudo chmod +x /usr/local/bin/ollama-distributed

# Fix ownership
sudo chown $USER:$USER ~/.ollama-distributed
```

#### Command Not Found
```bash
# Add to PATH
echo 'export PATH=$PATH:/usr/local/bin' >> ~/.bashrc
source ~/.bashrc

# Or create symlink
sudo ln -sf /usr/local/bin/ollama-distributed /usr/bin/ollama-distributed
```

#### Port Already in Use
```bash
# Find what's using the port
sudo lsof -i :8081

# Use different port
ollama-distributed start --port 8082
```

#### Validation Failures
```bash
# Run comprehensive validation
ollama-distributed validate --verbose

# Check system requirements
ollama-distributed validate --requirements

# Test network connectivity
ollama-distributed validate --connectivity
```

### Getting Help

If you encounter issues:

1. **Check logs**: `ollama-distributed logs --level debug`
2. **Run validation**: `ollama-distributed validate --all`
3. **Check GitHub issues**: [GitHub Issues](https://github.com/ollamamax/ollama-distributed/issues)
4. **Join Discord**: [Discord Community](https://discord.gg/ollamamax)

## Next Steps

After successful installation:

1. **[Quick Start Guide](../getting-started.md)** - Get your first cluster running
2. **[Configuration](../configuration.md)** - Customize your setup
3. **[Deployment Guide](../deployment/overview.md)** - Production deployment strategies