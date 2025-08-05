#!/bin/bash

# OllamaMax Build Environment Setup Script
# This script configures the Go build environment to resolve common issues

set -e

echo "🔧 OllamaMax Build Environment Setup"
echo "===================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    case $status in
        "SUCCESS")
            echo -e "${GREEN}✅ $message${NC}"
            ;;
        "ERROR")
            echo -e "${RED}❌ $message${NC}"
            ;;
        "WARNING")
            echo -e "${YELLOW}⚠️  $message${NC}"
            ;;
        "INFO")
            echo -e "${BLUE}ℹ️  $message${NC}"
            ;;
    esac
}

# Function to set environment variable
set_env_var() {
    local var_name=$1
    local var_value=$2
    export $var_name="$var_value"
    echo "export $var_name=\"$var_value\"" >> ~/.bashrc
    print_status "SUCCESS" "Set $var_name=$var_value"
}

print_status "INFO" "Starting build environment setup..."

# Step 1: Configure Go Proxy Settings
print_status "INFO" "Configuring Go proxy settings..."

# Try direct mode first (bypass proxy issues)
set_env_var "GOPROXY" "direct"
set_env_var "GOSUMDB" "off"

# Set reasonable timeouts
set_env_var "GOTIMEOUT" "30s"

# Step 2: Configure Module Cache
print_status "INFO" "Configuring module cache..."

# Create temporary cache directories
TEMP_GOMODCACHE="/tmp/gomodcache"
TEMP_GOCACHE="/tmp/gocache"

mkdir -p "$TEMP_GOMODCACHE"
mkdir -p "$TEMP_GOCACHE"

set_env_var "GOMODCACHE" "$TEMP_GOMODCACHE"
set_env_var "GOCACHE" "$TEMP_GOCACHE"

# Step 3: Configure Git for Go modules
print_status "INFO" "Configuring Git for Go modules..."

# Configure Git to use HTTPS instead of SSH
git config --global url."https://github.com/".insteadOf "git@github.com:"
git config --global url."https://".insteadOf "git://"

print_status "SUCCESS" "Git configured for HTTPS"

# Step 4: Set Go build flags
print_status "INFO" "Setting Go build flags..."

set_env_var "GOFLAGS" "-mod=readonly"

# Step 5: Create build verification script
print_status "INFO" "Creating build verification script..."

cat > scripts/verify-go-env.sh << 'EOF'
#!/bin/bash

echo "🔍 Go Environment Verification"
echo "============================="

# Test 1: Go version
echo -n "Testing Go version... "
if command -v go >/dev/null 2>&1; then
    echo "✅ $(go version)"
else
    echo "❌ Go not found"
    exit 1
fi

# Test 2: Environment variables
echo "Go environment:"
echo "  GOPROXY: $GOPROXY"
echo "  GOSUMDB: $GOSUMDB"
echo "  GOMODCACHE: $GOMODCACHE"
echo "  GOCACHE: $GOCACHE"

# Test 3: Simple Go program
echo -n "Testing simple Go program... "
cat > /tmp/test.go << 'GOEOF'
package main
import "fmt"
func main() { fmt.Println("Go environment working!") }
GOEOF

if timeout 10s go run /tmp/test.go >/dev/null 2>&1; then
    echo "✅ Success"
else
    echo "❌ Failed"
fi

rm -f /tmp/test.go

echo ""
echo "✅ Environment verification complete"
EOF

chmod +x scripts/verify-go-env.sh

print_status "SUCCESS" "Created verification script"

# Step 6: Create alternative build methods
print_status "INFO" "Creating alternative build methods..."

# Create Docker build option
cat > Dockerfile.build << 'EOF'
FROM golang:1.21-alpine AS builder

# Install dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o ollama-distributed ./cmd/node

# Create minimal runtime image
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/ollama-distributed .
CMD ["./ollama-distributed"]
EOF

print_status "SUCCESS" "Created Docker build file"

# Create vendor build option
cat > scripts/build-with-vendor.sh << 'EOF'
#!/bin/bash

echo "🏗️ Building with vendor dependencies"
echo "==================================="

# Create vendor directory
echo "Creating vendor directory..."
go mod vendor

# Build using vendor
echo "Building with vendor..."
go build -mod=vendor -o ollama-distributed ./cmd/node

echo "✅ Build complete: ./ollama-distributed"
EOF

chmod +x scripts/build-with-vendor.sh

print_status "SUCCESS" "Created vendor build script"

# Step 7: Create environment reset script
cat > scripts/reset-go-env.sh << 'EOF'
#!/bin/bash

echo "🔄 Resetting Go Environment"
echo "=========================="

# Clear module cache
echo "Clearing module cache..."
rm -rf $(go env GOMODCACHE) 2>/dev/null || true
rm -rf $(go env GOCACHE) 2>/dev/null || true

# Reset environment variables
unset GOPROXY
unset GOSUMDB
unset GOMODCACHE
unset GOCACHE
unset GOTIMEOUT
unset GOFLAGS

echo "✅ Environment reset complete"
echo "Run setup-build-env.sh to reconfigure"
EOF

chmod +x scripts/reset-go-env.sh

print_status "SUCCESS" "Created environment reset script"

# Step 8: Final verification
print_status "INFO" "Running final verification..."

# Source the new environment
source ~/.bashrc

print_status "SUCCESS" "Build environment setup complete!"

echo ""
echo "📋 Next Steps:"
echo "1. Run: source ~/.bashrc"
echo "2. Test: ./scripts/verify-go-env.sh"
echo "3. Build: go build ./cmd/node"
echo "4. Alternative: ./scripts/build-with-vendor.sh"
echo "5. Docker: docker build -f Dockerfile.build -t ollama-distributed ."
echo ""
echo "🚨 If issues persist:"
echo "- Check BUILD_ENVIRONMENT_FIX.md for detailed troubleshooting"
echo "- Use Docker build as fallback"
echo "- Run ./scripts/reset-go-env.sh and try again"
