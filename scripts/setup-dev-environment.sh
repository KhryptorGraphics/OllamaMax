#!/bin/bash

# Ollamamax Development Environment Setup Script
# This script sets up a complete development environment for Ollamamax

set -e

echo "🚀 Setting up Ollamamax Development Environment"
echo "=============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Check if running as root
if [ "$EUID" -eq 0 ]; then 
   print_error "Please do not run this script as root"
   exit 1
fi

# Detect OS
OS=""
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    OS="linux"
    DISTRO=$(lsb_release -si 2>/dev/null || echo "Unknown")
elif [[ "$OSTYPE" == "darwin"* ]]; then
    OS="macos"
elif [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" ]]; then
    OS="windows"
else
    print_error "Unsupported operating system: $OSTYPE"
    exit 1
fi

print_status "Detected OS: $OS $DISTRO"

# Step 1: Install system dependencies
echo ""
echo "📦 Installing System Dependencies..."
echo "-----------------------------------"

if [ "$OS" == "linux" ]; then
    # Update package manager
    sudo apt-get update || sudo yum update || sudo dnf update || true
    
    # Install build tools
    sudo apt-get install -y \
        build-essential \
        cmake \
        git \
        curl \
        wget \
        python3 \
        python3-pip \
        python3-venv \
        nodejs \
        npm \
        docker.io \
        docker-compose \
        redis-server \
        postgresql \
        postgresql-contrib \
        sqlite3 \
        jq \
        htop \
        tmux \
        vim \
        2>/dev/null || true
        
    print_status "Linux dependencies installed"
    
elif [ "$OS" == "macos" ]; then
    # Install Homebrew if not present
    if ! command -v brew &> /dev/null; then
        print_warning "Installing Homebrew..."
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    fi
    
    # Install dependencies
    brew install \
        cmake \
        git \
        node \
        redis \
        postgresql \
        sqlite \
        jq \
        htop \
        tmux \
        vim \
        docker \
        docker-compose \
        2>/dev/null || true
        
    print_status "macOS dependencies installed"
fi

# Step 2: Install Go (for P2P implementation)
echo ""
echo "🔧 Installing Go..."
echo "------------------"

GO_VERSION="1.21.5"
if ! command -v go &> /dev/null || [[ $(go version | awk '{print $3}') < "go$GO_VERSION" ]]; then
    print_warning "Installing Go $GO_VERSION..."
    
    if [ "$OS" == "linux" ]; then
        wget -q https://go.dev/dl/go$GO_VERSION.linux-amd64.tar.gz
        sudo rm -rf /usr/local/go
        sudo tar -C /usr/local -xzf go$GO_VERSION.linux-amd64.tar.gz
        rm go$GO_VERSION.linux-amd64.tar.gz
        
        # Add to PATH
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
        export PATH=$PATH:/usr/local/go/bin
        
    elif [ "$OS" == "macos" ]; then
        brew install go || brew upgrade go
    fi
    
    print_status "Go installed: $(go version)"
else
    print_status "Go already installed: $(go version)"
fi

# Step 3: Install Node.js dependencies
echo ""
echo "📦 Installing Node.js Dependencies..."
echo "------------------------------------"

# Check Node version
NODE_VERSION=$(node -v | cut -d'v' -f2 | cut -d'.' -f1)
if [ "$NODE_VERSION" -lt 18 ]; then
    print_warning "Node.js version is less than 18. Installing Node 18..."
    curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
    sudo apt-get install -y nodejs
fi

# Install global packages
npm install -g \
    typescript \
    ts-node \
    nodemon \
    pm2 \
    eslint \
    prettier \
    swagger-jsdoc \
    @swc/core \
    @swc/cli

print_status "Node.js dependencies installed"

# Step 4: Install Python dependencies
echo ""
echo "🐍 Setting up Python Environment..."
echo "----------------------------------"

# Create virtual environment
python3 -m venv venv
source venv/bin/activate

# Upgrade pip
pip install --upgrade pip

# Install Python packages
pip install \
    numpy \
    torch \
    transformers \
    fastapi \
    uvicorn \
    pytest \
    black \
    flake8 \
    mypy \
    requests \
    websockets \
    prometheus-client \
    redis \
    psycopg2-binary \
    sqlalchemy \
    alembic

print_status "Python environment configured"

# Step 5: Setup Docker
echo ""
echo "🐳 Configuring Docker..."
echo "-----------------------"

# Add user to docker group
if [ "$OS" == "linux" ]; then
    sudo usermod -aG docker $USER
    print_warning "You may need to log out and back in for Docker group changes to take effect"
fi

# Start Docker service
if [ "$OS" == "linux" ]; then
    sudo systemctl start docker || true
    sudo systemctl enable docker || true
elif [ "$OS" == "macos" ]; then
    open -a Docker || true
fi

# Pull required Docker images
docker pull redis:alpine || true
docker pull postgres:15-alpine || true
docker pull prom/prometheus:latest || true
docker pull grafana/grafana:latest || true

print_status "Docker configured"

# Step 6: Setup databases
echo ""
echo "💾 Setting up Databases..."
echo "-------------------------"

# Start PostgreSQL
if [ "$OS" == "linux" ]; then
    sudo systemctl start postgresql || true
    sudo systemctl enable postgresql || true
    
    # Create ollamamax database and user
    sudo -u postgres psql <<EOF 2>/dev/null || true
CREATE USER ollamamax WITH PASSWORD 'ollamamax_dev';
CREATE DATABASE ollamamax_db;
GRANT ALL PRIVILEGES ON DATABASE ollamamax_db TO ollamamax;
EOF
fi

# Start Redis
if [ "$OS" == "linux" ]; then
    sudo systemctl start redis-server || true
    sudo systemctl enable redis-server || true
elif [ "$OS" == "macos" ]; then
    brew services start redis || true
fi

print_status "Databases configured"

# Step 7: Install Ollamamax dependencies
echo ""
echo "📦 Installing Ollamamax Dependencies..."
echo "--------------------------------------"

# Install Node dependencies
npm install

# Install Go dependencies
go mod download 2>/dev/null || go mod init github.com/ollamamax/ollamamax

# Create required directories
mkdir -p \
    models \
    data \
    logs \
    cache \
    tmp \
    sprints \
    tests/unit \
    tests/integration \
    tests/e2e \
    scripts \
    docs/api \
    .swarm \
    .claude-flow

print_status "Ollamamax dependencies installed"

# Step 8: Download a small model for testing
echo ""
echo "🤖 Downloading Test Model..."
echo "---------------------------"

# Check if any models exist
if [ ! -d "models" ] || [ -z "$(ls -A models)" ]; then
    print_warning "Downloading TinyLlama for testing..."
    
    # Download TinyLlama (1.1B parameters, ~550MB)
    wget -q --show-progress \
        https://huggingface.co/TheBloke/TinyLlama-1.1B-Chat-v1.0-GGUF/resolve/main/tinyllama-1.1b-chat-v1.0.Q4_K_M.gguf \
        -O models/tinyllama-1.1b-chat-v1.0.Q4_K_M.gguf || true
        
    print_status "Test model downloaded"
else
    print_status "Models directory already contains files"
fi

# Step 9: Setup Git hooks
echo ""
echo "🔗 Setting up Git Hooks..."
echo "-------------------------"

# Create pre-commit hook
cat > .git/hooks/pre-commit <<'EOF'
#!/bin/bash
# Run linting and tests before commit

echo "Running pre-commit checks..."

# Run ESLint
npm run lint

# Run tests
npm test

# Check for console.log statements
if grep -r "console.log" --include="*.js" --include="*.ts" src/; then
    echo "Warning: console.log statements found in source code"
fi

echo "Pre-commit checks completed"
EOF

chmod +x .git/hooks/pre-commit

print_status "Git hooks configured"

# Step 10: Create environment file
echo ""
echo "⚙️ Creating Environment Configuration..."
echo "---------------------------------------"

if [ ! -f .env ]; then
    cat > .env <<EOF
# Ollamamax Development Environment Configuration
NODE_ENV=development

# Server Configuration
PORT=13000
HOST=0.0.0.0
API_VERSION=v1

# Database Configuration
POSTGRES_HOST=localhost
POSTGRES_PORT=13432
POSTGRES_USER=ollamamax
POSTGRES_PASSWORD=ollamamax_dev
POSTGRES_DB=ollamamax_db

REDIS_HOST=localhost
REDIS_PORT=13379
REDIS_PASSWORD=

# Model Configuration
MODEL_PATH=./models
DEFAULT_MODEL=tinyllama-1.1b-chat-v1.0.Q4_K_M.gguf
MAX_CONTEXT_LENGTH=2048
DEFAULT_BATCH_SIZE=1

# Security
JWT_SECRET=$(openssl rand -base64 32)
API_KEY=$(openssl rand -hex 32)
ENABLE_AUTH=true
ENABLE_RATE_LIMITING=true

# P2P Configuration
P2P_ENABLED=false
P2P_PORT=14000
P2P_BOOTSTRAP_NODES=

# Monitoring
ENABLE_METRICS=true
METRICS_PORT=13090
LOG_LEVEL=debug

# Resource Limits
MAX_MEMORY_GB=8
MAX_THREADS=4
GPU_ENABLED=false

# Development
HOT_RELOAD=true
VERBOSE_LOGGING=true
EOF
    print_status "Environment configuration created"
else
    print_status "Environment configuration already exists"
fi

# Step 11: Create VS Code configuration
echo ""
echo "📝 Creating VS Code Configuration..."
echo "-----------------------------------"

mkdir -p .vscode

cat > .vscode/launch.json <<'EOF'
{
  "version": "0.2.0",
  "configurations": [
    {
      "type": "node",
      "request": "launch",
      "name": "Launch Ollamamax",
      "skipFiles": ["<node_internals>/**"],
      "program": "${workspaceFolder}/main.js",
      "envFile": "${workspaceFolder}/.env",
      "console": "integratedTerminal"
    },
    {
      "type": "node",
      "request": "launch",
      "name": "Debug Tests",
      "skipFiles": ["<node_internals>/**"],
      "program": "${workspaceFolder}/node_modules/.bin/jest",
      "args": ["--runInBand"],
      "console": "integratedTerminal"
    }
  ]
}
EOF

cat > .vscode/settings.json <<'EOF'
{
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.fixAll.eslint": true
  },
  "eslint.validate": [
    "javascript",
    "javascriptreact",
    "typescript",
    "typescriptreact"
  ],
  "files.exclude": {
    "node_modules": true,
    ".git": true,
    "*.log": true
  },
  "search.exclude": {
    "node_modules": true,
    "models": true,
    "data": true,
    "logs": true
  }
}
EOF

print_status "VS Code configuration created"

# Step 12: Verify installation
echo ""
echo "✅ Verifying Installation..."
echo "---------------------------"

# Check all tools
TOOLS_OK=true

check_tool() {
    if command -v $1 &> /dev/null; then
        print_status "$1: $(command -v $1)"
    else
        print_error "$1: NOT FOUND"
        TOOLS_OK=false
    fi
}

check_tool node
check_tool npm
check_tool go
check_tool python3
check_tool docker
check_tool redis-cli
check_tool psql
check_tool git

# Final summary
echo ""
echo "🎉 Development Environment Setup Complete!"
echo "========================================="
echo ""
echo "Next steps:"
echo "1. Restart your terminal or run: source ~/.bashrc"
echo "2. Start the development server: npm run dev"
echo "3. View API documentation: http://localhost:13000/docs"
echo "4. Run tests: npm test"
echo ""

if [ "$TOOLS_OK" = false ]; then
    print_warning "Some tools were not found. Please install them manually."
fi

print_status "Happy coding! 🚀"