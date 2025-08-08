#!/bin/bash

# OllamaMax Health Check Script
# Comprehensive health checking for deployment validation

set -e

echo "🏥 OllamaMax Health Check"
echo "========================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default configuration
ENVIRONMENT="staging"
TIMEOUT=300
API_URL=""
VERBOSE=false
CHECK_INTERVAL=5
MAX_RETRIES=60

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

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --environment)
                ENVIRONMENT="$2"
                shift 2
                ;;
            --timeout)
                TIMEOUT="$2"
                shift 2
                ;;
            --api-url)
                API_URL="$2"
                shift 2
                ;;
            --verbose)
                VERBOSE=true
                shift
                ;;
            --help)
                show_help
                exit 0
                ;;
            *)
                echo "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# Show help
show_help() {
    cat << EOF
OllamaMax Health Check Script

Usage: $0 [OPTIONS]

Options:
    --environment ENV    Target environment (staging, production)
    --timeout SECONDS    Maximum time to wait for health checks
    --api-url URL        Custom API URL to check
    --verbose            Enable verbose output
    --help               Show this help message

Examples:
    $0 --environment staging
    $0 --environment production --timeout 600
    $0 --api-url http://localhost:8080 --verbose
EOF
}

# Determine API URL based on environment
determine_api_url() {
    if [ -n "$API_URL" ]; then
        return
    fi

    case $ENVIRONMENT in
        staging)
            API_URL="https://staging.ollamamax.com"
            ;;
        production)
            API_URL="https://ollamamax.com"
            ;;
        local)
            API_URL="http://localhost:8080"
            ;;
        *)
            API_URL="http://localhost:8080"
            ;;
    esac
}

# Basic connectivity check
check_connectivity() {
    print_status "INFO" "Checking connectivity to $API_URL..."
    
    if curl -s --connect-timeout 10 "$API_URL" > /dev/null; then
        print_status "SUCCESS" "Connectivity check passed"
        return 0
    else
        print_status "ERROR" "Cannot connect to $API_URL"
        return 1
    fi
}

# Health endpoint check
check_health_endpoint() {
    print_status "INFO" "Checking health endpoint..."
    
    local health_url="$API_URL/health"
    local response
    local http_code
    
    response=$(curl -s -w "%{http_code}" "$health_url" 2>/dev/null)
    http_code="${response: -3}"
    response_body="${response%???}"
    
    if [ "$http_code" = "200" ]; then
        print_status "SUCCESS" "Health endpoint responding (HTTP $http_code)"
        
        if [ "$VERBOSE" = true ]; then
            echo "Response: $response_body"
        fi
        return 0
    else
        print_status "ERROR" "Health endpoint failed (HTTP $http_code)"
        if [ "$VERBOSE" = true ]; then
            echo "Response: $response_body"
        fi
        return 1
    fi
}

# API endpoints check
check_api_endpoints() {
    print_status "INFO" "Checking API endpoints..."
    
    local endpoints=(
        "/api/v1/proxy/status"
        "/api/v1/proxy/instances"
        "/api/v1/proxy/metrics"
    )
    
    local failed_endpoints=0
    
    for endpoint in "${endpoints[@]}"; do
        local url="$API_URL$endpoint"
        local http_code
        
        http_code=$(curl -s -o /dev/null -w "%{http_code}" "$url" 2>/dev/null)
        
        if [ "$http_code" = "200" ] || [ "$http_code" = "401" ]; then
            print_status "SUCCESS" "Endpoint $endpoint responding (HTTP $http_code)"
        else
            print_status "ERROR" "Endpoint $endpoint failed (HTTP $http_code)"
            ((failed_endpoints++))
        fi
    done
    
    if [ $failed_endpoints -eq 0 ]; then
        print_status "SUCCESS" "All API endpoints are healthy"
        return 0
    else
        print_status "ERROR" "$failed_endpoints API endpoints failed"
        return 1
    fi
}

# Database connectivity check
check_database() {
    print_status "INFO" "Checking database connectivity..."
    
    local db_health_url="$API_URL/api/v1/health/database"
    local response
    local http_code
    
    response=$(curl -s -w "%{http_code}" "$db_health_url" 2>/dev/null)
    http_code="${response: -3}"
    
    if [ "$http_code" = "200" ]; then
        print_status "SUCCESS" "Database connectivity check passed"
        return 0
    else
        print_status "WARNING" "Database health check not available or failed (HTTP $http_code)"
        return 0  # Non-critical for basic health check
    fi
}

# Performance check
check_performance() {
    print_status "INFO" "Checking response performance..."
    
    local start_time
    local end_time
    local duration
    
    start_time=$(date +%s%N)
    
    if curl -s --max-time 10 "$API_URL/health" > /dev/null; then
        end_time=$(date +%s%N)
        duration=$(( (end_time - start_time) / 1000000 ))  # Convert to milliseconds
        
        if [ $duration -lt 1000 ]; then
            print_status "SUCCESS" "Response time: ${duration}ms (excellent)"
        elif [ $duration -lt 3000 ]; then
            print_status "SUCCESS" "Response time: ${duration}ms (good)"
        elif [ $duration -lt 5000 ]; then
            print_status "WARNING" "Response time: ${duration}ms (slow)"
        else
            print_status "ERROR" "Response time: ${duration}ms (too slow)"
            return 1
        fi
        return 0
    else
        print_status "ERROR" "Performance check failed (timeout)"
        return 1
    fi
}

# Security headers check
check_security_headers() {
    print_status "INFO" "Checking security headers..."
    
    local headers
    headers=$(curl -s -I "$API_URL" 2>/dev/null)
    
    local security_headers=(
        "Strict-Transport-Security"
        "X-Content-Type-Options"
        "X-Frame-Options"
        "X-XSS-Protection"
    )
    
    local missing_headers=0
    
    for header in "${security_headers[@]}"; do
        if echo "$headers" | grep -qi "$header"; then
            print_status "SUCCESS" "Security header present: $header"
        else
            print_status "WARNING" "Security header missing: $header"
            ((missing_headers++))
        fi
    done
    
    if [ $missing_headers -eq 0 ]; then
        print_status "SUCCESS" "All security headers present"
        return 0
    else
        print_status "WARNING" "$missing_headers security headers missing"
        return 0  # Non-critical for basic health check
    fi
}

# Wait for service to be ready
wait_for_ready() {
    print_status "INFO" "Waiting for service to be ready (timeout: ${TIMEOUT}s)..."
    
    local retries=0
    local max_retries=$((TIMEOUT / CHECK_INTERVAL))
    
    while [ $retries -lt $max_retries ]; do
        if check_connectivity && check_health_endpoint; then
            print_status "SUCCESS" "Service is ready after $((retries * CHECK_INTERVAL))s"
            return 0
        fi
        
        ((retries++))
        if [ $retries -lt $max_retries ]; then
            if [ "$VERBOSE" = true ]; then
                print_status "INFO" "Retry $retries/$max_retries in ${CHECK_INTERVAL}s..."
            fi
            sleep $CHECK_INTERVAL
        fi
    done
    
    print_status "ERROR" "Service not ready after ${TIMEOUT}s timeout"
    return 1
}

# Comprehensive health check
run_comprehensive_check() {
    print_status "INFO" "Running comprehensive health check..."
    
    local checks=(
        "check_connectivity"
        "check_health_endpoint"
        "check_api_endpoints"
        "check_database"
        "check_performance"
        "check_security_headers"
    )
    
    local failed_checks=0
    local total_checks=${#checks[@]}
    
    for check in "${checks[@]}"; do
        if ! $check; then
            ((failed_checks++))
        fi
        echo ""
    done
    
    # Summary
    local passed_checks=$((total_checks - failed_checks))
    local success_rate=$((passed_checks * 100 / total_checks))
    
    echo "📊 Health Check Summary"
    echo "======================"
    echo "Passed: $passed_checks/$total_checks ($success_rate%)"
    echo "Failed: $failed_checks/$total_checks"
    echo "Environment: $ENVIRONMENT"
    echo "API URL: $API_URL"
    
    if [ $failed_checks -eq 0 ]; then
        print_status "SUCCESS" "All health checks passed! 🎉"
        return 0
    elif [ $success_rate -ge 80 ]; then
        print_status "WARNING" "Most health checks passed, but some issues detected"
        return 0
    else
        print_status "ERROR" "Multiple health checks failed"
        return 1
    fi
}

# Main function
main() {
    parse_args "$@"
    
    print_status "INFO" "Starting health check for environment: $ENVIRONMENT"
    
    determine_api_url
    
    print_status "INFO" "Target URL: $API_URL"
    print_status "INFO" "Timeout: ${TIMEOUT}s"
    
    # Wait for service to be ready first
    if ! wait_for_ready; then
        exit 1
    fi
    
    echo ""
    
    # Run comprehensive checks
    if run_comprehensive_check; then
        print_status "SUCCESS" "Health check completed successfully!"
        exit 0
    else
        print_status "ERROR" "Health check failed!"
        exit 1
    fi
}

# Run main function
main "$@"
