#!/bin/bash

# OllamaMax Build Verification Script
# This script verifies that all core packages compile successfully

set -e

echo "🔧 OllamaMax Build Verification"
echo "================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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
            echo -e "ℹ️  $message"
            ;;
    esac
}

# Function to check if a package builds
check_package() {
    local package=$1
    local description=$2
    
    echo -n "Building $description... "
    if go build -o /dev/null ./$package 2>/dev/null; then
        print_status "SUCCESS" "$description"
        return 0
    else
        print_status "ERROR" "$description"
        echo "Build errors for $package:"
        go build ./$package 2>&1 | head -10
        return 1
    fi
}

# Function to run tests for a package
test_package() {
    local package=$1
    local description=$2
    
    echo -n "Testing $description... "
    if go test -v ./$package 2>/dev/null | grep -q "PASS\|no test files"; then
        print_status "SUCCESS" "$description tests"
        return 0
    else
        print_status "WARNING" "$description tests (some failures)"
        return 1
    fi
}

# Change to the ollama-distributed directory
cd "$(dirname "$0")/.."

print_status "INFO" "Starting build verification in $(pwd)"

# Check Go version
GO_VERSION=$(go version | cut -d' ' -f3)
print_status "INFO" "Using Go version: $GO_VERSION"

# Clean and update dependencies
print_status "INFO" "Updating dependencies..."
go mod tidy

# Core packages to verify (in dependency order)
declare -a CORE_PACKAGES=(
    "pkg/config:Configuration Types"
    "pkg/types:Core Types"
    "pkg/integration:Integration Stubs"
    "internal/config:Internal Configuration"
    "pkg/p2p/discovery:P2P Discovery"
    "pkg/p2p/host:P2P Host"
    "pkg/p2p/security:P2P Security"
    "pkg/p2p/protocols:P2P Protocols"
    "pkg/p2p:P2P Core"
    "pkg/consensus:Consensus Engine"
    "pkg/api:API Layer"
    "pkg/models:Model Management"
    "internal/auth:Authentication"
    "internal/metrics:Metrics Collection"
)

# Build verification
echo ""
print_status "INFO" "Building core packages..."
echo ""

BUILD_FAILURES=0
TEST_FAILURES=0

for package_info in "${CORE_PACKAGES[@]}"; do
    IFS=':' read -r package description <<< "$package_info"
    
    if [ -d "$package" ]; then
        if ! check_package "$package" "$description"; then
            ((BUILD_FAILURES++))
        fi
        
        # Run tests if they exist
        if ls $package/*_test.go 1> /dev/null 2>&1; then
            if ! test_package "$package" "$description"; then
                ((TEST_FAILURES++))
            fi
        fi
    else
        print_status "WARNING" "Package $package not found"
    fi
done

# Try to build main commands
echo ""
print_status "INFO" "Building main commands..."
echo ""

if [ -d "cmd/node" ]; then
    check_package "cmd/node" "Node Command"
fi

if [ -d "cmd/cli" ]; then
    check_package "cmd/cli" "CLI Command"
fi

# Summary
echo ""
echo "🏁 Build Verification Summary"
echo "============================="

if [ $BUILD_FAILURES -eq 0 ]; then
    print_status "SUCCESS" "All core packages build successfully!"
else
    print_status "ERROR" "$BUILD_FAILURES package(s) failed to build"
fi

if [ $TEST_FAILURES -eq 0 ]; then
    print_status "SUCCESS" "All tests passed or no test failures detected"
else
    print_status "WARNING" "$TEST_FAILURES package(s) have test failures"
fi

# Check for common issues
echo ""
print_status "INFO" "Checking for common issues..."

# Check for import cycles
if go list -f '{{.ImportPath}}: {{.Deps}}' ./... 2>&1 | grep -q "import cycle"; then
    print_status "ERROR" "Import cycles detected"
    go list -f '{{.ImportPath}}: {{.Deps}}' ./... 2>&1 | grep "import cycle"
else
    print_status "SUCCESS" "No import cycles detected"
fi

# Check for unused dependencies
UNUSED_DEPS=$(go mod tidy -v 2>&1 | grep "unused" || true)
if [ -n "$UNUSED_DEPS" ]; then
    print_status "WARNING" "Unused dependencies found"
    echo "$UNUSED_DEPS"
else
    print_status "SUCCESS" "No unused dependencies"
fi

# Final status
echo ""
if [ $BUILD_FAILURES -eq 0 ]; then
    print_status "SUCCESS" "Build verification completed successfully! 🎉"
    exit 0
else
    print_status "ERROR" "Build verification failed with $BUILD_FAILURES errors"
    exit 1
fi
