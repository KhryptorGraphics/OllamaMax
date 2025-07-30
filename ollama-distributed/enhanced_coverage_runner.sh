#!/bin/bash

# Enhanced Coverage Test Runner for Ollama Distributed System
# This script runs comprehensive test coverage analysis with detailed reporting

set -e

# Configuration
PROJECT_NAME="ollama-distributed"
COVERAGE_TARGET=80
MIN_COVERAGE=60
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
ARTIFACTS_DIR="test-artifacts/enhanced_coverage_${TIMESTAMP}"
REPORTS_DIR="${ARTIFACTS_DIR}/reports"
LOGS_DIR="${ARTIFACTS_DIR}/logs"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Create directories
mkdir -p "$ARTIFACTS_DIR" "$REPORTS_DIR" "$LOGS_DIR"

echo -e "${BLUE}🧪 Enhanced Coverage Analysis for ${PROJECT_NAME}${NC}"
echo -e "${BLUE}================================================${NC}"
echo "Timestamp: $(date)"
echo "Target Coverage: ${COVERAGE_TARGET}%"
echo "Minimum Coverage: ${MIN_COVERAGE}%"
echo "Artifacts Directory: ${ARTIFACTS_DIR}"
echo ""

# Function to print section headers
print_section() {
    echo -e "${PURPLE}📋 $1${NC}"
    echo "----------------------------------------"
}

# Function to run tests with coverage
run_coverage_tests() {
    local test_type="$1"
    local package_pattern="$2"
    local output_file="$3"
    local extra_flags="$4"
    
    print_section "Running ${test_type} Tests with Coverage"
    
    echo "Package Pattern: ${package_pattern}"
    echo "Output File: ${output_file}"
    echo "Extra Flags: ${extra_flags}"
    echo ""
    
    if go test ${extra_flags} -coverprofile="${output_file}" -covermode=atomic ${package_pattern} 2>&1 | tee "${LOGS_DIR}/${test_type}_test.log"; then
        echo -e "${GREEN}✅ ${test_type} tests completed successfully${NC}"
        
        # Generate coverage report
        if [ -f "${output_file}" ]; then
            go tool cover -html="${output_file}" -o "${REPORTS_DIR}/${test_type}_coverage.html"
            
            # Extract coverage percentage
            local coverage=$(go tool cover -func="${output_file}" | grep "total:" | awk '{print $3}' | sed 's/%//')
            echo -e "${CYAN}📊 ${test_type} Coverage: ${coverage}%${NC}"
            
            # Store coverage for summary
            echo "${test_type}:${coverage}" >> "${ARTIFACTS_DIR}/coverage_summary.txt"
        fi
    else
        echo -e "${RED}❌ ${test_type} tests failed${NC}"
        return 1
    fi
    echo ""
}

# Function to analyze coverage by package
analyze_package_coverage() {
    print_section "Package-Level Coverage Analysis"
    
    local combined_profile="${ARTIFACTS_DIR}/combined_coverage.out"
    
    # Combine all coverage profiles
    echo "mode: atomic" > "$combined_profile"
    for profile in "${ARTIFACTS_DIR}"/*.out; do
        if [ -f "$profile" ] && [ "$(basename "$profile")" != "combined_coverage.out" ]; then
            tail -n +2 "$profile" >> "$combined_profile"
        fi
    done
    
    if [ -f "$combined_profile" ]; then
        echo "📊 Coverage by Package:"
        echo "======================="
        
        # Generate detailed package coverage report
        go tool cover -func="$combined_profile" | grep -v "total:" | while read line; do
            if [[ $line =~ ^([^[:space:]]+)[[:space:]]+[^[:space:]]+[[:space:]]+([0-9.]+)% ]]; then
                package=${BASH_REMATCH[1]}
                coverage=${BASH_REMATCH[2]}
                
                if (( $(echo "$coverage >= $COVERAGE_TARGET" | bc -l) )); then
                    echo -e "${GREEN}✅ $package: ${coverage}%${NC}"
                elif (( $(echo "$coverage >= $MIN_COVERAGE" | bc -l) )); then
                    echo -e "${YELLOW}⚠️  $package: ${coverage}%${NC}"
                else
                    echo -e "${RED}❌ $package: ${coverage}%${NC}"
                fi
            fi
        done | tee "${REPORTS_DIR}/package_coverage.txt"
        
        # Generate HTML report for combined coverage
        go tool cover -html="$combined_profile" -o "${REPORTS_DIR}/combined_coverage.html"
        
        # Calculate overall coverage
        local total_coverage=$(go tool cover -func="$combined_profile" | grep "total:" | awk '{print $3}' | sed 's/%//')
        echo ""
        echo -e "${CYAN}🎯 Overall Coverage: ${total_coverage}%${NC}"
        echo "overall:${total_coverage}" >> "${ARTIFACTS_DIR}/coverage_summary.txt"
    fi
    echo ""
}

# Function to run mutation testing
run_mutation_testing() {
    print_section "Mutation Testing Analysis"
    
    # Check if go-mutesting is available
    if ! command -v go-mutesting &> /dev/null; then
        echo -e "${YELLOW}⚠️  go-mutesting not found. Installing...${NC}"
        go install github.com/zimmski/go-mutesting/cmd/go-mutesting@latest
    fi
    
    if command -v go-mutesting &> /dev/null; then
        echo "Running mutation testing on core packages..."
        
        # Run mutation testing on critical packages
        local mutation_packages=(
            "./pkg/consensus"
            "./pkg/p2p"
            "./internal/auth"
        )
        
        for package in "${mutation_packages[@]}"; do
            if [ -d "$package" ]; then
                echo "Testing mutations in: $package"
                local package_name=$(basename "$package")
                
                go-mutesting --disable "branch/if,branch/else" "$package" 2>&1 | tee "${LOGS_DIR}/mutation_${package_name}.log" || true
            fi
        done
        
        echo -e "${GREEN}✅ Mutation testing completed${NC}"
    else
        echo -e "${YELLOW}⚠️  Skipping mutation testing (go-mutesting not available)${NC}"
    fi
    echo ""
}

# Function to run property-based tests
run_property_tests() {
    print_section "Property-Based Testing"
    
    # Check if gopter is available in any test files
    if grep -r "gopter" . --include="*.go" &> /dev/null; then
        echo "Running property-based tests..."
        go test -tags=property -v ./... 2>&1 | tee "${LOGS_DIR}/property_tests.log" || true
        echo -e "${GREEN}✅ Property-based tests completed${NC}"
    else
        echo -e "${YELLOW}⚠️  No property-based tests found${NC}"
        echo "Consider adding property-based tests using github.com/leanovate/gopter"
    fi
    echo ""
}

# Function to analyze test quality
analyze_test_quality() {
    print_section "Test Quality Analysis"
    
    local quality_report="${REPORTS_DIR}/test_quality.txt"
    
    echo "📊 Test Quality Metrics" > "$quality_report"
    echo "======================" >> "$quality_report"
    echo "" >> "$quality_report"
    
    # Count test files and functions
    local test_files=$(find . -name "*_test.go" | wc -l)
    local test_functions=$(grep -r "func Test" . --include="*_test.go" | wc -l)
    local benchmark_functions=$(grep -r "func Benchmark" . --include="*_test.go" | wc -l)
    local example_functions=$(grep -r "func Example" . --include="*_test.go" | wc -l)
    
    echo "Test Files: $test_files" >> "$quality_report"
    echo "Test Functions: $test_functions" >> "$quality_report"
    echo "Benchmark Functions: $benchmark_functions" >> "$quality_report"
    echo "Example Functions: $example_functions" >> "$quality_report"
    echo "" >> "$quality_report"
    
    # Analyze test patterns
    echo "Test Patterns Analysis:" >> "$quality_report"
    echo "----------------------" >> "$quality_report"
    
    local table_tests=$(grep -r "tests := \[\]struct" . --include="*_test.go" | wc -l)
    local subtests=$(grep -r "t\.Run(" . --include="*_test.go" | wc -l)
    local parallel_tests=$(grep -r "t\.Parallel()" . --include="*_test.go" | wc -l)
    local cleanup_tests=$(grep -r "t\.Cleanup(" . --include="*_test.go" | wc -l)
    
    echo "Table-driven tests: $table_tests" >> "$quality_report"
    echo "Subtests: $subtests" >> "$quality_report"
    echo "Parallel tests: $parallel_tests" >> "$quality_report"
    echo "Tests with cleanup: $cleanup_tests" >> "$quality_report"
    echo "" >> "$quality_report"
    
    # Test coverage by type
    echo "Test Coverage by Type:" >> "$quality_report"
    echo "---------------------" >> "$quality_report"
    
    local unit_test_coverage=$(grep -r "unit" "${ARTIFACTS_DIR}/coverage_summary.txt" | cut -d: -f2 || echo "0")
    local integration_test_coverage=$(grep -r "integration" "${ARTIFACTS_DIR}/coverage_summary.txt" | cut -d: -f2 || echo "0")
    local e2e_test_coverage=$(grep -r "e2e" "${ARTIFACTS_DIR}/coverage_summary.txt" | cut -d: -f2 || echo "0")
    
    echo "Unit Tests: ${unit_test_coverage}%" >> "$quality_report"
    echo "Integration Tests: ${integration_test_coverage}%" >> "$quality_report"
    echo "E2E Tests: ${e2e_test_coverage}%" >> "$quality_report"
    echo "" >> "$quality_report"
    
    # Display quality report
    cat "$quality_report"
    echo ""
}

# Function to check for common test issues
check_test_issues() {
    print_section "Test Issues Analysis"
    
    local issues_report="${REPORTS_DIR}/test_issues.txt"
    
    echo "🔍 Potential Test Issues" > "$issues_report"
    echo "========================" >> "$issues_report"
    echo "" >> "$issues_report"
    
    # Check for missing error checks in tests
    echo "Missing error checks in tests:" >> "$issues_report"
    grep -rn "_, err :=" . --include="*_test.go" | grep -v "assert\|require\|if err" | head -10 >> "$issues_report" || echo "None found" >> "$issues_report"
    echo "" >> "$issues_report"
    
    # Check for tests without assertions
    echo "Tests without assertions:" >> "$issues_report"
    find . -name "*_test.go" -exec grep -L "assert\|require\|Error\|Fatal" {} \; | head -10 >> "$issues_report" || echo "None found" >> "$issues_report"
    echo "" >> "$issues_report"
    
    # Check for very long test functions
    echo "Very long test functions (>100 lines):" >> "$issues_report"
    find . -name "*_test.go" -exec awk '/^func Test/ {start=NR; name=$0} /^}$/ && start {if (NR-start > 100) print FILENAME":"start":"name}' {} \; >> "$issues_report" || echo "None found" >> "$issues_report"
    echo "" >> "$issues_report"
    
    # Check for hardcoded values that might indicate brittle tests
    echo "Potential hardcoded values in tests:" >> "$issues_report"
    grep -rn "localhost:808\|127\.0\.0\.1:808\|hardcoded\|magic.*number" . --include="*_test.go" | head -10 >> "$issues_report" || echo "None found" >> "$issues_report"
    echo "" >> "$issues_report"
    
    # Display issues
    cat "$issues_report"
    echo ""
}

# Function to generate recommendations
generate_recommendations() {
    print_section "Coverage Improvement Recommendations"
    
    local recommendations="${REPORTS_DIR}/recommendations.txt"
    
    echo "📋 Coverage Improvement Recommendations" > "$recommendations"
    echo "=======================================" >> "$recommendations"
    echo "" >> "$recommendations"
    
    # Analyze low coverage packages
    if [ -f "${REPORTS_DIR}/package_coverage.txt" ]; then
        echo "Low Coverage Packages (< ${MIN_COVERAGE}%):" >> "$recommendations"
        grep "❌" "${REPORTS_DIR}/package_coverage.txt" | while read line; do
            package=$(echo "$line" | sed 's/❌ \([^:]*\):.*/\1/')
            echo "- $package: Add unit tests for core functions" >> "$recommendations"
        done
        echo "" >> "$recommendations"
    fi
    
    # General recommendations
    echo "General Recommendations:" >> "$recommendations"
    echo "- Add integration tests for component interactions" >> "$recommendations"
    echo "- Implement property-based tests for algorithms" >> "$recommendations"
    echo "- Add benchmark tests for performance-critical paths" >> "$recommendations"
    echo "- Consider chaos engineering tests for resilience" >> "$recommendations"
    echo "- Implement contract tests for API endpoints" >> "$recommendations"
    echo "- Add mutation testing for critical business logic" >> "$recommendations"
    echo "" >> "$recommendations"
    
    # Test type recommendations
    echo "Test Type Recommendations:" >> "$recommendations"
    echo "- Unit Tests: Focus on isolated component testing" >> "$recommendations"
    echo "- Integration Tests: Test component interactions" >> "$recommendations"
    echo "- E2E Tests: Test complete user workflows" >> "$recommendations"
    echo "- Performance Tests: Benchmark critical operations" >> "$recommendations"
    echo "- Security Tests: Validate authentication and authorization" >> "$recommendations"
    echo "- Chaos Tests: Test system resilience under failures" >> "$recommendations"
    echo "" >> "$recommendations"
    
    cat "$recommendations"
    echo ""
}

# Function to generate final summary
generate_final_summary() {
    print_section "Final Coverage Summary"
    
    local summary="${REPORTS_DIR}/final_summary.txt"
    
    echo "🎯 Final Coverage Summary" > "$summary"
    echo "=========================" >> "$summary"
    echo "Timestamp: $(date)" >> "$summary"
    echo "Project: ${PROJECT_NAME}" >> "$summary"
    echo "" >> "$summary"
    
    if [ -f "${ARTIFACTS_DIR}/coverage_summary.txt" ]; then
        while IFS=: read -r test_type coverage; do
            if [ "$test_type" = "overall" ]; then
                echo "Overall Coverage: ${coverage}%" >> "$summary"
                
                if (( $(echo "$coverage >= $COVERAGE_TARGET" | bc -l) )); then
                    echo "Status: 🎉 EXCELLENT (Target Achieved)" >> "$summary"
                elif (( $(echo "$coverage >= $MIN_COVERAGE" | bc -l) )); then
                    echo "Status: ✅ GOOD (Above Minimum)" >> "$summary"
                else
                    echo "Status: ⚠️  NEEDS IMPROVEMENT (Below Minimum)" >> "$summary"
                fi
            else
                echo "${test_type^} Coverage: ${coverage}%" >> "$summary"
            fi
        done < "${ARTIFACTS_DIR}/coverage_summary.txt"
    fi
    
    echo "" >> "$summary"
    echo "Artifacts Location: ${ARTIFACTS_DIR}" >> "$summary"
    echo "HTML Reports: ${REPORTS_DIR}/" >> "$summary"
    echo "Logs: ${LOGS_DIR}/" >> "$summary"
    echo "" >> "$summary"
    
    cat "$summary"
    
    # Also display summary to console
    echo ""
    echo -e "${BLUE}📊 Enhanced Coverage Analysis Complete${NC}"
    echo -e "${BLUE}=====================================${NC}"
    cat "$summary"
    echo ""
}

# Main execution
main() {
    # Clean previous coverage files
    find . -name "*.out" -type f -delete 2>/dev/null || true
    
    # Run different types of coverage tests
    echo -e "${CYAN}🧪 Running Enhanced Coverage Tests...${NC}"
    echo ""
    
    # Unit Tests Coverage
    run_coverage_tests "unit" "./pkg/... ./internal/..." "${ARTIFACTS_DIR}/unit_coverage.out" "-race -timeout=5m"
    
    # Integration Tests Coverage  
    run_coverage_tests "integration" "./tests/integration/..." "${ARTIFACTS_DIR}/integration_coverage.out" "-race -timeout=10m"
    
    # E2E Tests Coverage (if not in short mode)
    if [ "${RUN_E2E:-true}" = "true" ]; then
        run_coverage_tests "e2e" "./tests/e2e/..." "${ARTIFACTS_DIR}/e2e_coverage.out" "-race -timeout=15m"
    else
        echo -e "${YELLOW}⚠️  Skipping E2E tests (RUN_E2E=false)${NC}"
    fi
    
    # Security Tests Coverage
    run_coverage_tests "security" "./tests/security/..." "${ARTIFACTS_DIR}/security_coverage.out" "-race -timeout=10m"
    
    # Performance Tests Coverage (benchmarks)
    if [ "${RUN_BENCHMARKS:-false}" = "true" ]; then
        run_coverage_tests "performance" "./tests/performance/..." "${ARTIFACTS_DIR}/performance_coverage.out" "-bench=. -benchmem -timeout=20m"
    else
        echo -e "${YELLOW}⚠️  Skipping performance tests (RUN_BENCHMARKS=false)${NC}"
    fi
    
    # Chaos Tests Coverage (if enabled)
    if [ "${RUN_CHAOS:-false}" = "true" ]; then
        run_coverage_tests "chaos" "./tests/chaos/..." "${ARTIFACTS_DIR}/chaos_coverage.out" "-race -timeout=15m"
    else
        echo -e "${YELLOW}⚠️  Skipping chaos tests (RUN_CHAOS=false)${NC}"
    fi
    
    # Analyze coverage
    analyze_package_coverage
    
    # Advanced analysis
    run_mutation_testing
    run_property_tests
    analyze_test_quality
    check_test_issues
    generate_recommendations
    
    # Final summary
    generate_final_summary
    
    echo -e "${GREEN}✅ Enhanced coverage analysis completed successfully!${NC}"
    echo -e "${CYAN}📁 All artifacts saved to: ${ARTIFACTS_DIR}${NC}"
    echo -e "${CYAN}🌐 View HTML reports at: ${REPORTS_DIR}/combined_coverage.html${NC}"
}

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "Enhanced Coverage Test Runner"
        echo ""
        echo "Usage: $0 [options]"
        echo ""
        echo "Options:"
        echo "  --help, -h          Show this help message"
        echo "  --quick            Run only unit and integration tests"
        echo "  --full             Run all test types including E2E and chaos"
        echo ""
        echo "Environment Variables:"
        echo "  RUN_E2E=false      Skip E2E tests"
        echo "  RUN_BENCHMARKS=true   Run performance benchmarks"
        echo "  RUN_CHAOS=true     Run chaos engineering tests"
        echo "  COVERAGE_TARGET=80 Set coverage target percentage"
        echo ""
        exit 0
        ;;
    --quick)
        export RUN_E2E=false
        export RUN_BENCHMARKS=false
        export RUN_CHAOS=false
        echo -e "${YELLOW}🏃‍♂️ Quick mode: Running unit and integration tests only${NC}"
        ;;
    --full)
        export RUN_E2E=true
        export RUN_BENCHMARKS=true
        export RUN_CHAOS=true
        echo -e "${BLUE}🚀 Full mode: Running all test types${NC}"
        ;;
esac

# Run main function
main "$@"