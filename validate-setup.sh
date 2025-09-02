#!/bin/bash

# OllamaMax Setup Validation Script
# Checks if all required services and dependencies are ready for testing

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔍 OllamaMax Setup Validation${NC}"
echo -e "${BLUE}=============================${NC}\n"

# Function to check command availability
check_command() {
    local cmd=$1
    local name=$2
    
    if command -v "$cmd" &> /dev/null; then
        echo -e "✅ ${GREEN}$name: Available${NC}"
        return 0
    else
        echo -e "❌ ${RED}$name: Not found${NC}"
        return 1
    fi
}

# Function to check service health
check_service() {
    local url=$1
    local name=$2
    local timeout=${3:-5}
    
    if timeout "$timeout" curl -s "$url" > /dev/null 2>&1; then
        echo -e "✅ ${GREEN}$name: Running${NC}"
        
        # Try to get additional info
        response=$(timeout "$timeout" curl -s "$url" 2>/dev/null || echo "")
        if [[ -n "$response" && ${#response} -gt 10 ]]; then
            preview=$(echo "$response" | head -c 50)
            echo -e "   ${YELLOW}Response preview: ${preview}...${NC}"
        fi
        
        return 0
    else
        echo -e "❌ ${RED}$name: Not responding${NC}"
        return 1
    fi
}

# Function to check port availability
check_port() {
    local host=$1
    local port=$2
    local name=$3
    
    if nc -z "$host" "$port" 2>/dev/null; then
        echo -e "✅ ${GREEN}$name (${host}:${port}): Port open${NC}"
        return 0
    else
        echo -e "❌ ${RED}$name (${host}:${port}): Port closed${NC}"
        return 1
    fi
}

# Check system dependencies
echo -e "${BLUE}📦 Checking System Dependencies...${NC}"

failed_deps=0

if ! check_command "node" "Node.js"; then
    ((failed_deps++))
    echo -e "   ${YELLOW}Install: https://nodejs.org/${NC}"
fi

if ! check_command "npm" "NPM"; then
    ((failed_deps++))
    echo -e "   ${YELLOW}Should come with Node.js${NC}"
fi

if ! check_command "python3" "Python 3"; then
    ((failed_deps++))
    echo -e "   ${YELLOW}Install: https://python.org/${NC}"
fi

if ! check_command "curl" "cURL"; then
    ((failed_deps++))
    echo -e "   ${YELLOW}Install with package manager${NC}"
fi

if ! check_command "nc" "Netcat"; then
    ((failed_deps++))
    echo -e "   ${YELLOW}Install: netcat or netcat-openbsd${NC}"
fi

# Check Playwright installation
echo -e "\n${BLUE}🎭 Checking Playwright Setup...${NC}"

if [ -d "node_modules/@playwright/test" ]; then
    echo -e "✅ ${GREEN}Playwright: Installed${NC}"
    
    # Check if browsers are installed
    if npx playwright --version > /dev/null 2>&1; then
        version=$(npx playwright --version 2>/dev/null || echo "unknown")
        echo -e "   ${YELLOW}Version: $version${NC}"
    fi
    
    # Check browser installations
    browsers_dir="$HOME/.cache/ms-playwright"
    if [ -d "$browsers_dir" ]; then
        echo -e "✅ ${GREEN}Playwright Browsers: Likely installed${NC}"
        browser_count=$(find "$browsers_dir" -name "chrome*" -o -name "firefox*" -o -name "webkit*" 2>/dev/null | wc -l)
        echo -e "   ${YELLOW}Browser directories found: $browser_count${NC}"
    else
        echo -e "⚠️  ${YELLOW}Playwright Browsers: May need installation${NC}"
        echo -e "   ${YELLOW}Run: npx playwright install${NC}"
    fi
else
    echo -e "❌ ${RED}Playwright: Not installed${NC}"
    echo -e "   ${YELLOW}Run: npm install @playwright/test${NC}"
    ((failed_deps++))
fi

# Check project structure
echo -e "\n${BLUE}📁 Checking Project Structure...${NC}"

required_files=(
    "web-interface/index.html"
    "web-interface/app.js"
    "tests/api-health-tests.js"
    "tests/ui-interaction-tests.js"
    "tests/p2p-model-migration-tests.js"
    "tests/comprehensive-test-strategy.js"
    "playwright.config.js"
    "package.json"
)

missing_files=0

for file in "${required_files[@]}"; do
    if [ -f "$file" ]; then
        echo -e "✅ ${GREEN}$file: Exists${NC}"
    else
        echo -e "❌ ${RED}$file: Missing${NC}"
        ((missing_files++))
    fi
done

# Check test results directory
if [ ! -d "test-results" ]; then
    echo -e "📁 ${YELLOW}Creating test-results directory...${NC}"
    mkdir -p test-results
fi

# Check service availability
echo -e "\n${BLUE}🌐 Checking Service Availability...${NC}"

failed_services=0

# Web Interface
if ! check_service "http://localhost:8080" "Web Interface" 5; then
    ((failed_services++))
    echo -e "   ${YELLOW}Start with: cd web-interface && python3 -m http.server 8080${NC}"
fi

# API Server
if ! check_service "http://localhost:13100/health" "API Server" 5; then
    ((failed_services++))
    echo -e "   ${YELLOW}Start with: node api-server/server.js${NC}"
fi

# Worker Nodes
workers=("13000" "13001" "13002")
healthy_workers=0

for port in "${workers[@]}"; do
    if check_port "localhost" "$port" "Worker $port"; then
        ((healthy_workers++))
        
        # Try to get version info
        if timeout 3 curl -s "http://localhost:$port/api/version" > /dev/null 2>&1; then
            echo -e "   ${GREEN}Worker $port: API responding${NC}"
        fi
    else
        echo -e "   ${YELLOW}Worker $port: Offline (non-critical)${NC}"
    fi
done

echo -e "   ${BLUE}Healthy Workers: $healthy_workers/3${NC}"

# Generate summary
echo -e "\n${BLUE}📊 VALIDATION SUMMARY${NC}"
echo -e "${BLUE}====================${NC}"

total_issues=$((failed_deps + missing_files + failed_services))

if [ $total_issues -eq 0 ]; then
    echo -e "✅ ${GREEN}System Status: Ready for Testing${NC}"
    echo -e "🚀 ${GREEN}All dependencies and services are available${NC}"
    echo -e "🎯 ${GREEN}You can now run comprehensive tests${NC}"
    
    echo -e "\n${YELLOW}Quick Start Commands:${NC}"
    echo -e "  ./run-tests.sh quick          # 5-minute smoke test"
    echo -e "  ./run-tests.sh p2p            # Test P2P migration"
    echo -e "  ./run-tests.sh full           # Complete test suite"
    echo -e "  npm run test:show-report      # View test results"
    
    exit 0
else
    echo -e "⚠️  ${YELLOW}System Status: Needs Attention${NC}"
    echo -e "📋 ${YELLOW}Issues found: $total_issues${NC}"
    
    if [ $failed_deps -gt 0 ]; then
        echo -e "   - $failed_deps missing system dependencies"
    fi
    
    if [ $missing_files -gt 0 ]; then
        echo -e "   - $missing_files missing project files"
    fi
    
    if [ $failed_services -gt 0 ]; then
        echo -e "   - $failed_services critical services not running"
    fi
    
    echo -e "\n${YELLOW}Next Steps:${NC}"
    echo -e "1. Fix the issues listed above"
    echo -e "2. Run this validation script again"
    echo -e "3. Once ready, run: ./run-tests.sh quick"
    
    exit 1
fi