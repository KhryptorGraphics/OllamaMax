#!/bin/bash

# Grafana Dashboard Validation Script
# Validates dashboard JSON files and provisioning configuration

set -e

echo "🔍 Validating Grafana Dashboard Implementation..."
echo

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

ERRORS=0
WARNINGS=0

# Check provisioning configuration
echo "📋 Checking provisioning configuration..."
if [ -f "monitoring/grafana/provisioning/dashboards/dashboard.yml" ]; then
    echo -e "${GREEN}✓${NC} Provisioning config exists"
    
    # Check path configuration
    if grep -q "path: /var/lib/grafana/dashboards" monitoring/grafana/provisioning/dashboards/dashboard.yml; then
        echo -e "${GREEN}✓${NC} Dashboard path correctly configured"
    else
        echo -e "${RED}✗${NC} Dashboard path not correctly set"
        ((ERRORS++))
    fi
else
    echo -e "${RED}✗${NC} Provisioning config missing"
    ((ERRORS++))
fi
echo

# Check dashboard directory
echo "📁 Checking dashboard directory..."
if [ -d "monitoring/grafana/dashboards" ]; then
    echo -e "${GREEN}✓${NC} Dashboard directory exists"
    
    DASHBOARD_COUNT=$(ls -1 monitoring/grafana/dashboards/*.json 2>/dev/null | wc -l)
    echo "   Found $DASHBOARD_COUNT dashboard file(s)"
else
    echo -e "${RED}✗${NC} Dashboard directory missing"
    ((ERRORS++))
fi
echo

# Validate required dashboard files
echo "🎨 Validating dashboard files..."

REQUIRED_DASHBOARDS=(
    "api-performance.json"
    "database-performance.json"
    "p2p-detailed.json"
)

for dashboard in "${REQUIRED_DASHBOARDS[@]}"; do
    DASHBOARD_PATH="monitoring/grafana/dashboards/$dashboard"
    
    if [ -f "$DASHBOARD_PATH" ]; then
        echo -e "${GREEN}✓${NC} $dashboard exists"
        
        # Validate JSON syntax
        if jq empty "$DASHBOARD_PATH" 2>/dev/null; then
            echo "   └─ JSON syntax valid"
            
            # Extract dashboard info
            TITLE=$(jq -r '.title' "$DASHBOARD_PATH")
            PANEL_COUNT=$(jq '.panels | length' "$DASHBOARD_PATH")
            DASH_UID=$(jq -r '.uid // "none"' "$DASHBOARD_PATH")
            
            echo "   └─ Title: $TITLE"
            echo "   └─ Panels: $PANEL_COUNT"
            echo "   └─ UID: $DASH_UID"
            
            # Check for required fields
            if [ "$PANEL_COUNT" -eq 0 ]; then
                echo -e "   ${YELLOW}⚠${NC} No panels defined"
                ((WARNINGS++))
            fi
            
            # Check for Prometheus datasource
            if jq -e '.panels[].datasource' "$DASHBOARD_PATH" | grep -q "Prometheus" 2>/dev/null; then
                echo "   └─ Prometheus datasource configured"
            else
                echo -e "   ${YELLOW}⚠${NC} Prometheus datasource may not be configured"
                ((WARNINGS++))
            fi
            
        else
            echo -e "   ${RED}✗${NC} Invalid JSON syntax"
            ((ERRORS++))
        fi
    else
        echo -e "${RED}✗${NC} $dashboard missing"
        ((ERRORS++))
    fi
    echo
done

# Check docker-compose volume mounts
echo "🐳 Checking Docker Compose configuration..."
if [ -f "docker-compose.yml" ]; then
    if grep -A10 "grafana:" docker-compose.yml | grep -q "grafana/dashboards"; then
        echo -e "${GREEN}✓${NC} Dashboard volume mount configured"
    else
        echo -e "${RED}✗${NC} Dashboard volume mount not found in docker-compose.yml"
        ((ERRORS++))
    fi
    
    if grep -A10 "grafana:" docker-compose.yml | grep -q "grafana/provisioning"; then
        echo -e "${GREEN}✓${NC} Provisioning volume mount configured"
    else
        echo -e "${RED}✗${NC} Provisioning volume mount not found in docker-compose.yml"
        ((ERRORS++))
    fi
else
    echo -e "${RED}✗${NC} docker-compose.yml not found"
    ((ERRORS++))
fi
echo

# Check metric queries
echo "📊 Analyzing dashboard metrics..."
for dashboard in "${REQUIRED_DASHBOARDS[@]}"; do
    DASHBOARD_PATH="monitoring/grafana/dashboards/$dashboard"
    
    if [ -f "$DASHBOARD_PATH" ]; then
        METRIC_COUNT=$(jq '[.panels[].targets[]?.expr // empty] | length' "$DASHBOARD_PATH")
        echo "   $dashboard: $METRIC_COUNT metric queries"
        
        # Extract unique metrics
        METRICS=$(jq -r '[.panels[].targets[]?.expr // empty] | unique[]' "$DASHBOARD_PATH" | grep -oE 'ollamamax_[a-z_]+' | sort -u)
        if [ -n "$METRICS" ]; then
            echo "   └─ Metrics used:"
            echo "$METRICS" | sed 's/^/      - /'
        fi
    fi
done
echo

# Summary
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    echo -e "${GREEN}✓ All checks passed!${NC}"
    echo
    echo "📋 Implementation Summary:"
    echo "   - Provisioning: Configured"
    echo "   - Dashboards: ${#REQUIRED_DASHBOARDS[@]} files validated"
    echo "   - Docker: Volume mounts configured"
    echo
    echo "🚀 Ready to deploy!"
    echo "   Run: docker-compose up -d grafana"
    echo "   Access: http://localhost:3001"
    exit 0
elif [ $ERRORS -eq 0 ]; then
    echo -e "${YELLOW}⚠ Validation completed with $WARNINGS warning(s)${NC}"
    exit 0
else
    echo -e "${RED}✗ Validation failed with $ERRORS error(s) and $WARNINGS warning(s)${NC}"
    exit 1
fi
