#!/bin/bash

# OllamaMax Comprehensive Test Execution Script
# Usage: ./run-tests.sh [quick|api|ui|p2p|full|cross-browser]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
WEB_INTERFACE_URL="http://localhost:8080"
API_SERVER_URL="http://localhost:13100"
WORKERS=("localhost:13000" "localhost:13001" "localhost:13002")

echo -e "${BLUE}🚀 OllamaMax Comprehensive Testing Suite${NC}"
echo -e "${BLUE}=======================================${NC}"

# Function to check if a service is running
check_service() {
    local url=$1
    local name=$2
    
    if curl -s "$url" > /dev/null 2>&1; then
        echo -e "✅ ${GREEN}$name: Running${NC}"
        return 0
    else
        echo -e "❌ ${RED}$name: Not responding${NC}"
        return 1
    fi
}

# Function to run pre-flight checks
run_preflight_checks() {
    echo -e "\n${YELLOW}🔍 Running pre-flight checks...${NC}"
    
    local failed=0
    
    # Check web interface
    if ! check_service "$WEB_INTERFACE_URL" "Web Interface"; then
        failed=1
    fi
    
    # Check API server
    if ! check_service "$API_SERVER_URL/health" "API Server"; then
        failed=1
    fi
    
    # Check worker nodes
    for worker in "${WORKERS[@]}"; do
        if ! check_service "http://$worker/api/version" "Worker ($worker)"; then
            echo -e "⚠️  ${YELLOW}Worker $worker: Not responding (non-critical)${NC}"
        fi
    done
    
    if [ $failed -eq 1 ]; then
        echo -e "\n${RED}❌ Critical services are not running. Please start the required services.${NC}"
        echo -e "${YELLOW}Required services:${NC}"
        echo -e "  - Web Interface: python3 -m http.server 8080 (in web-interface/)"
        echo -e "  - API Server: node api-server/server.js (port 13100)"
        exit 1
    fi
    
    echo -e "\n${GREEN}✅ Pre-flight checks passed!${NC}\n"
}

# Function to run specific test suite
run_test_suite() {
    local test_type=$1
    local description=$2
    local command=$3
    
    echo -e "${BLUE}📋 Running $test_type Tests...${NC}"
    echo -e "📄 $description"
    echo -e "🚀 Command: $command\n"
    
    if eval "$command"; then
        echo -e "\n${GREEN}✅ $test_type tests passed!${NC}\n"
        return 0
    else
        echo -e "\n${RED}❌ $test_type tests failed!${NC}\n"
        return 1
    fi
}

# Main execution logic
main() {
    local test_scenario=${1:-"quick"}
    
    echo -e "🎯 Test Scenario: ${YELLOW}$test_scenario${NC}\n"
    
    # Run pre-flight checks
    run_preflight_checks
    
    local total_tests=0
    local passed_tests=0
    local failed_tests=0
    
    case $test_scenario in
        "quick")
            echo -e "${BLUE}🏃 Quick Smoke Test (5-10 minutes)${NC}"
            
            total_tests=2
            
            if run_test_suite "API Health" \
                "Validates API server and worker connectivity" \
                "npx playwright test tests/api-health-tests.js --project=chromium"; then
                ((passed_tests++))
            else
                ((failed_tests++))
            fi
            
            if run_test_suite "UI Interaction" \
                "Tests core UI functionality and navigation" \
                "npx playwright test tests/ui-interaction-tests.js --project=chromium"; then
                ((passed_tests++))
            else
                ((failed_tests++))
            fi
            ;;
            
        "api")
            echo -e "${BLUE}🔧 API Health Test${NC}"
            
            total_tests=1
            
            if run_test_suite "API Health" \
                "Comprehensive API and worker validation" \
                "npx playwright test tests/api-health-tests.js --project=chromium"; then
                ((passed_tests++))
            else
                ((failed_tests++))
            fi
            ;;
            
        "ui")
            echo -e "${BLUE}🎨 UI Interaction Test${NC}"
            
            total_tests=1
            
            if run_test_suite "UI Interaction" \
                "Complete UI element and interaction testing" \
                "npx playwright test tests/ui-interaction-tests.js --project=chromium"; then
                ((passed_tests++))
            else
                ((failed_tests++))
            fi
            ;;
            
        "p2p")
            echo -e "${BLUE}🔄 P2P Model Migration Test (CRITICAL)${NC}"
            
            total_tests=1
            
            if run_test_suite "P2P Model Migration" \
                "Tests P2P model migration controls and functionality" \
                "npx playwright test tests/p2p-model-migration-tests.js --project=chromium"; then
                ((passed_tests++))
            else
                ((failed_tests++))
            fi
            ;;
            
        "full")
            echo -e "${BLUE}🎯 Full Comprehensive Test (20-30 minutes)${NC}"
            
            total_tests=4
            
            if run_test_suite "API Health" \
                "System health and connectivity validation" \
                "npx playwright test tests/api-health-tests.js --project=chromium"; then
                ((passed_tests++))
            else
                ((failed_tests++))
            fi
            
            if run_test_suite "UI Interaction" \
                "Complete UI functionality testing" \
                "npx playwright test tests/ui-interaction-tests.js --project=chromium"; then
                ((passed_tests++))
            else
                ((failed_tests++))
            fi
            
            if run_test_suite "P2P Model Migration" \
                "P2P functionality and model management" \
                "npx playwright test tests/p2p-model-migration-tests.js --project=chromium"; then
                ((passed_tests++))
            else
                ((failed_tests++))
            fi
            
            if run_test_suite "Comprehensive Integration" \
                "End-to-end system validation" \
                "npx playwright test tests/comprehensive-test-strategy.js --project=chromium"; then
                ((passed_tests++))
            else
                ((failed_tests++))
            fi
            ;;
            
        "cross-browser")
            echo -e "${BLUE}🌐 Cross-Browser Compatibility Test${NC}"
            
            total_tests=3
            
            for browser in chromium firefox webkit; do
                if run_test_suite "UI ($browser)" \
                    "UI functionality in $browser browser" \
                    "npx playwright test tests/ui-interaction-tests.js --project=$browser"; then
                    ((passed_tests++))
                else
                    ((failed_tests++))
                fi
            done
            ;;
            
        *)
            echo -e "${RED}❌ Unknown test scenario: $test_scenario${NC}"
            echo -e "${YELLOW}Available scenarios: quick, api, ui, p2p, full, cross-browser${NC}"
            exit 1
            ;;
    esac
    
    # Display summary
    echo -e "${BLUE}📊 TEST EXECUTION SUMMARY${NC}"
    echo -e "${BLUE}========================${NC}"
    echo -e "Total Test Suites: $total_tests"
    echo -e "✅ Passed: ${GREEN}$passed_tests${NC}"
    echo -e "❌ Failed: ${RED}$failed_tests${NC}"
    
    local success_rate
    if [ $total_tests -eq 0 ]; then
        success_rate=0
    else
        success_rate=$((passed_tests * 100 / total_tests))
    fi
    
    echo -e "🎯 Success Rate: ${success_rate}%"
    
    # Generate test report
    echo -e "\n${YELLOW}📄 Generating test report...${NC}"
    if [ -d "test-results" ]; then
        echo -e "📊 HTML Report: file://$(pwd)/test-results/html-report/index.html"
        echo -e "📋 JSON Report: $(pwd)/test-results/results.json"
    fi
    
    # Final result
    if [ $failed_tests -eq 0 ]; then
        echo -e "\n${GREEN}🎉 All tests passed! System is ready for use.${NC}"
        exit 0
    else
        echo -e "\n${RED}⚠️  Some tests failed. Please review the results and fix issues.${NC}"
        exit 1
    fi
}

# Execute main function with arguments
main "$@"