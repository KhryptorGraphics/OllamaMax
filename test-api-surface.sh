#!/bin/bash

# OllamaMax API Surface Validation Script
# Tests the configured API endpoints and verifies connectivity

set -e

echo "🚀 OllamaMax API Surface Validation"
echo "==================================="
echo

# Configuration
API_BASE="http://localhost:13100"
WS_ENDPOINT="ws://localhost:13100/chat"
GO_API_BASE="http://localhost:8080"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

success=0
total=0

test_endpoint() {
    local name="$1"
    local url="$2"
    local expected_status="$3"
    
    total=$((total + 1))
    
    echo -n "Testing $name... "
    
    if command -v curl >/dev/null 2>&1; then
        response=$(curl -s -w "%{http_code}" -o /dev/null "$url" --connect-timeout 5)
        
        if [ "$response" = "$expected_status" ]; then
            echo -e "${GREEN}✓ PASS${NC}"
            success=$((success + 1))
        else
            echo -e "${RED}✗ FAIL (status: $response)${NC}"
        fi
    else
        echo -e "${YELLOW}⚠ SKIP (curl not available)${NC}"
    fi
}

test_websocket() {
    local name="$1"
    local ws_url="$2"
    
    total=$((total + 1))
    
    echo -n "Testing $name... "
    
    if command -v wscat >/dev/null 2>&1; then
        # Test WebSocket connection (timeout after 3 seconds)
        if timeout 3 wscat -c "$ws_url" </dev/null >/dev/null 2>&1; then
            echo -e "${GREEN}✓ PASS${NC}"
            success=$((success + 1))
        else
            echo -e "${RED}✗ FAIL${NC}"
        fi
    else
        echo -e "${YELLOW}⚠ SKIP (wscat not available)${NC}"
    fi
}

echo "📋 Testing Node.js API Endpoints"
echo "--------------------------------"
test_endpoint "API Root" "$API_BASE/" "200"
test_endpoint "Health Check" "$API_BASE/health" "200"
test_endpoint "Liveness Probe" "$API_BASE/health/live" "200"
test_endpoint "Readiness Probe" "$API_BASE/health/ready" "200"
test_endpoint "Authentication" "$API_BASE/auth" "404"  # Should exist but return 404 for GET
test_endpoint "OpenAPI Spec" "$API_BASE/openapi.json" "200"
test_endpoint "Metrics" "$API_BASE/metrics" "200"
test_endpoint "Documentation" "$API_BASE/docs" "200"

echo
echo "📋 Testing WebSocket Connection"
echo "------------------------------"
test_websocket "WebSocket Chat" "$WS_ENDPOINT"

echo
echo "📋 Testing Go Backend API"
echo "-------------------------"
test_endpoint "Go API Health" "$GO_API_BASE/health" "200"

echo
echo "📋 Testing Ollama Engine"
echo "------------------------"
test_endpoint "Ollama Health" "http://localhost:11434/api/health" "200"

echo
echo "📊 Validation Summary"
echo "===================="
echo "Passed: $success/$total tests"

if [ $success -eq $total ]; then
    echo -e "${GREEN}🎉 All tests passed! API surface is working correctly.${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed. Please check the services are running.${NC}"
    echo
    echo "🔧 Troubleshooting steps:"
    echo "1. Start the Node.js API: npm start"
    echo "2. Start the Go backend: go run ./cmd/node/main.go"
    echo "3. Start Ollama: ollama serve"
    echo "4. Verify environment configuration in .env file"
    exit 1
fi