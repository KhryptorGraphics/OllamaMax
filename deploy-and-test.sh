#!/bin/bash

# OllamaMax - Complete Deployment and Testing Script
# This script deploys the application and runs comprehensive tests

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

log_success() {
    echo -e "${GREEN}✓${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

log_error() {
    echo -e "${RED}✗${NC} $1"
}

log_header() {
    echo ""
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo ""
}

# Check if running from project root
if [ ! -f "package.json" ]; then
    log_error "Please run this script from the project root directory"
    exit 1
fi

log_header "OllamaMax Deployment & Testing"

# Step 1: Check dependencies
log_info "Checking dependencies..."

if ! command -v node &> /dev/null; then
    log_error "Node.js is not installed"
    exit 1
fi
log_success "Node.js $(node --version) found"

if ! command -v npm &> /dev/null; then
    log_error "npm is not installed"
    exit 1
fi
log_success "npm $(npm --version) found"

# Step 2: Install dependencies
log_info "Installing dependencies..."
npm install --silent
log_success "Dependencies installed"

# Step 3: Check environment variables
log_info "Checking environment configuration..."

if [ ! -f ".env" ]; then
    log_warning ".env file not found, creating from .env.example..."
    if [ -f ".env.example" ]; then
        cp .env.example .env
        log_success "Created .env from .env.example"
    else
        log_error ".env.example not found"
        exit 1
    fi
else
    log_success ".env file exists"
fi

# Step 4: Create data directory
log_info "Setting up data directory..."
mkdir -p data
log_success "Data directory ready"

# Step 5: Kill any existing server on port 13000
log_info "Checking for existing server..."
if lsof -Pi :13000 -sTCP:LISTEN -t >/dev/null 2>&1 ; then
    log_warning "Port 13000 is in use, stopping existing server..."
    kill $(lsof -t -i:13000) 2>/dev/null || true
    sleep 2
    log_success "Existing server stopped"
else
    log_success "Port 13000 is available"
fi

# Step 6: Start the server
log_header "Starting Server"

log_info "Starting OllamaMax server..."
npm start > logs/server.log 2>&1 &
SERVER_PID=$!
log_success "Server started (PID: $SERVER_PID)"

# Step 7: Wait for server to be ready
log_info "Waiting for server to be ready..."
MAX_WAIT=30
WAIT_COUNT=0

while [ $WAIT_COUNT -lt $MAX_WAIT ]; do
    if curl -s http://localhost:13000/health > /dev/null 2>&1; then
        log_success "Server is ready!"
        break
    fi
    sleep 1
    WAIT_COUNT=$((WAIT_COUNT + 1))
    echo -n "."
done

echo ""

if [ $WAIT_COUNT -eq $MAX_WAIT ]; then
    log_error "Server failed to start within ${MAX_WAIT} seconds"
    log_info "Check logs/server.log for details"
    exit 1
fi

# Step 8: Run health checks
log_header "Running Health Checks"

# Health check
log_info "Testing /health endpoint..."
HEALTH_RESPONSE=$(curl -s http://localhost:13000/health)
if echo "$HEALTH_RESPONSE" | grep -q "healthy"; then
    log_success "Health check passed"
else
    log_error "Health check failed"
    echo "$HEALTH_RESPONSE"
fi

# Liveness check
log_info "Testing /health/live endpoint..."
if curl -s http://localhost:13000/health/live | grep -q "alive"; then
    log_success "Liveness check passed"
else
    log_warning "Liveness check failed"
fi

# Readiness check
log_info "Testing /health/ready endpoint..."
READY_RESPONSE=$(curl -s -w "%{http_code}" http://localhost:13000/health/ready -o /dev/null)
if [ "$READY_RESPONSE" = "200" ]; then
    log_success "Readiness check passed"
else
    log_warning "Readiness check returned $READY_RESPONSE (may need more time to initialize)"
fi

# Step 9: Test API endpoints
log_header "Testing API Endpoints"

# Test models endpoint
log_info "Testing /v1/models endpoint..."
if curl -s http://localhost:13000/v1/models | grep -q "llama"; then
    log_success "Models endpoint working"
else
    log_warning "Models endpoint may have issues"
fi

# Test nodes endpoint
log_info "Testing /api/nodes endpoint..."
NODES_RESPONSE=$(curl -s http://localhost:13000/api/nodes)
NODE_COUNT=$(echo "$NODES_RESPONSE" | grep -o '"id"' | wc -l)
if [ "$NODE_COUNT" -gt 0 ]; then
    log_success "Nodes endpoint working ($NODE_COUNT nodes found)"
else
    log_warning "No nodes found"
fi

# Test documentation
log_info "Testing /docs endpoint..."
if curl -s http://localhost:13000/docs | grep -q "swagger"; then
    log_success "Documentation endpoint working"
else
    log_warning "Documentation endpoint may have issues"
fi

# Step 10: Display server information
log_header "Server Information"

echo -e "${GREEN}🚀 OllamaMax is running!${NC}"
echo ""
echo -e "${BLUE}Server URLs:${NC}"
echo -e "  📍 Main API:        ${GREEN}http://localhost:13000${NC}"
echo -e "  📚 Documentation:   ${GREEN}http://localhost:13000/docs${NC}"
echo -e "  ❤️  Health Check:   ${GREEN}http://localhost:13000/health${NC}"
echo -e "  🔑 Authentication:  ${GREEN}http://localhost:13000/auth${NC}"
echo -e "  🤖 OpenAI API:      ${GREEN}http://localhost:13000/v1${NC}"
echo -e "  📊 Metrics:         ${GREEN}http://localhost:13000/metrics${NC}"
echo -e "  🔌 WebSocket:       ${GREEN}ws://localhost:13000/chat${NC}"
echo -e "  🖥️  Nodes API:       ${GREEN}http://localhost:13000/api/nodes${NC}"
echo ""
echo -e "${BLUE}Web Interface:${NC}"
echo -e "  🌐 Chat Interface:  ${GREEN}http://localhost:13000/index.html${NC}"
echo -e "  🔐 Auth Page:       ${GREEN}http://localhost:13000/auth.html${NC}"
echo ""
echo -e "${BLUE}Server Process:${NC}"
echo -e "  PID: ${GREEN}$SERVER_PID${NC}"
echo -e "  Logs: ${GREEN}logs/server.log${NC}"
echo ""
echo -e "${YELLOW}To stop the server:${NC}"
echo -e "  kill $SERVER_PID"
echo ""

# Step 11: Open web interface (optional)
if command -v xdg-open &> /dev/null; then
    log_info "Opening web interface in browser..."
    xdg-open http://localhost:13000/index.html 2>/dev/null &
elif command -v open &> /dev/null; then
    log_info "Opening web interface in browser..."
    open http://localhost:13000/index.html 2>/dev/null &
fi

# Step 12: Offer to run tests
echo ""
read -p "Would you like to run comprehensive tests? (y/n) " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    log_header "Running Comprehensive Tests"
    node tests/comprehensive-userflow-test.js
fi

log_success "Deployment complete!"
echo ""
echo -e "${GREEN}✨ OllamaMax is ready to use!${NC}"
echo ""

