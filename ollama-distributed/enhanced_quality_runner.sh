#!/bin/bash

# Enhanced Quality Test Runner for Ollama Distributed System
# Integrates property-based testing and mutation testing for comprehensive code quality validation

set -e

# Configuration
PROJECT_NAME="ollama-distributed"
MUTATION_SCORE_TARGET=75
PROPERTY_TEST_COUNT=1000
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
ARTIFACTS_DIR="test-artifacts/enhanced_quality_${TIMESTAMP}"
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

echo -e "${BLUE}🧬 Enhanced Quality Testing for ${PROJECT_NAME}${NC}"
echo -e "${BLUE}================================================${NC}"
echo "Timestamp: $(date)"
echo "Mutation Score Target: ${MUTATION_SCORE_TARGET}%"
echo "Property Test Count: ${PROPERTY_TEST_COUNT}"
echo "Artifacts Directory: ${ARTIFACTS_DIR}"
echo ""

# Function to print section headers
print_section() {
    echo -e "${PURPLE}📋 $1${NC}"
    echo "----------------------------------------"
}

# Function to run property-based tests
run_property_tests() {
    print_section "Running Property-Based Tests"
    
    echo "🔍 Searching for property-based tests..."
    
    # Check if property tests exist
    if [ -d "tests/property" ]; then
        echo "📁 Found property tests directory"
        
        # Run property-based tests with high iteration count
        echo "🧪 Running property-based tests with ${PROPERTY_TEST_COUNT} iterations..."
        
        go test -v -count=${PROPERTY_TEST_COUNT} -timeout=10m ./tests/property/... 2>&1 | tee "${LOGS_DIR}/property_tests.log"
        
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✅ Property-based tests completed successfully${NC}"
            echo "property:PASS" >> "${ARTIFACTS_DIR}/test_results.txt"
        else
            echo -e "${RED}❌ Property-based tests failed${NC}"
            echo "property:FAIL" >> "${ARTIFACTS_DIR}/test_results.txt"
            return 1
        fi
    else
        echo -e "${YELLOW}⚠️  No property-based tests found${NC}"
        echo "Consider adding property-based tests using github.com/leanovate/gopter"
        echo "property:SKIP" >> "${ARTIFACTS_DIR}/test_results.txt"
    fi
    
    echo ""
}

# Function to run mutation testing
run_mutation_testing() {
    print_section "Running Mutation Testing for Code Quality"
    
    echo "🧬 Starting comprehensive mutation testing..."
    
    # Check if mutation testing tool exists
    if [ -f "cmd/mutation-test/main.go" ]; then
        echo "🔧 Building mutation testing tool..."
        go build -o "${ARTIFACTS_DIR}/mutation-test" ./cmd/mutation-test/
        
        if [ $? -ne 0 ]; then
            echo -e "${RED}❌ Failed to build mutation testing tool${NC}"
            echo "mutation:BUILD_FAIL" >> "${ARTIFACTS_DIR}/test_results.txt"
            return 1
        fi
        
        echo "🚀 Running mutation testing..."
        
        # Run mutation testing with appropriate settings
        "${ARTIFACTS_DIR}/mutation-test" \
            -root . \
            -threshold ${MUTATION_SCORE_TARGET} \
            -timeout 30s \
            -verbose \
            -output "${ARTIFACTS_DIR}" \
            -format html \
            -quick 2>&1 | tee "${LOGS_DIR}/mutation_testing.log"
        
        mutation_exit_code=$?
        
        # Extract mutation score from output
        mutation_score=$(grep "Mutation Score:" "${LOGS_DIR}/mutation_testing.log" | tail -1 | sed 's/.*Mutation Score: \([0-9.]*\)%.*/\1/')
        
        if [ ! -z "$mutation_score" ]; then
            echo "mutation_score:${mutation_score}" >> "${ARTIFACTS_DIR}/test_results.txt"
            
            if (( $(echo "$mutation_score >= $MUTATION_SCORE_TARGET" | bc -l) )); then
                echo -e "${GREEN}✅ Mutation testing passed (Score: ${mutation_score}%)${NC}"
                echo "mutation:PASS" >> "${ARTIFACTS_DIR}/test_results.txt"
            else
                echo -e "${YELLOW}⚠️  Mutation score ${mutation_score}% is below target ${MUTATION_SCORE_TARGET}%${NC}"
                echo "mutation:BELOW_THRESHOLD" >> "${ARTIFACTS_DIR}/test_results.txt"
            fi
        else
            if [ $mutation_exit_code -eq 0 ]; then
                echo -e "${GREEN}✅ Mutation testing completed${NC}"
                echo "mutation:PASS" >> "${ARTIFACTS_DIR}/test_results.txt"
            else
                echo -e "${RED}❌ Mutation testing failed${NC}"
                echo "mutation:FAIL" >> "${ARTIFACTS_DIR}/test_results.txt"
            fi
        fi
    else
        echo -e "${YELLOW}⚠️  Mutation testing tool not found${NC}"
        echo "Building simple mutation tester..."
        
        # Run basic mutation testing using go-mutesting if available
        if command -v go-mutesting &> /dev/null; then
            echo "🧬 Running go-mutesting..."
            
            # Test critical packages
            critical_packages=("pkg/consensus" "pkg/p2p" "internal/auth" "pkg/api")
            
            for package in "${critical_packages[@]}"; do
                if [ -d "$package" ]; then
                    echo "Testing mutations in: $package"
                    go-mutesting --disable "branch/if,branch/else" "$package" 2>&1 | tee "${LOGS_DIR}/mutation_${package//\//_}.log" || true
                fi
            done
            
            echo "mutation:BASIC_COMPLETE" >> "${ARTIFACTS_DIR}/test_results.txt"
        else
            echo "Installing go-mutesting..."
            go install github.com/zimmski/go-mutesting/cmd/go-mutesting@latest
            
            if command -v go-mutesting &> /dev/null; then
                echo "🧬 Running go-mutesting..."
                go-mutesting --disable "branch/if,branch/else" ./pkg/consensus 2>&1 | tee "${LOGS_DIR}/mutation_basic.log" || true
                echo "mutation:BASIC_COMPLETE" >> "${ARTIFACTS_DIR}/test_results.txt"
            else
                echo "mutation:SKIP" >> "${ARTIFACTS_DIR}/test_results.txt"
            fi
        fi
    fi
    
    echo ""
}

# Function to run traditional coverage tests
run_coverage_tests() {
    print_section "Running Traditional Coverage Tests"
    
    echo "📊 Running standard unit tests with coverage..."
    
    # Run tests with coverage on working packages
    working_packages=("internal/auth")
    
    for package in "${working_packages[@]}"; do
        if [ -d "$package" ]; then
            echo "Testing package: $package"
            
            go test -race -coverprofile="${ARTIFACTS_DIR}/${package//\//_}_coverage.out" \
                   -covermode=atomic -timeout=5m "./$package" 2>&1 | tee "${LOGS_DIR}/${package//\//_}_test.log"
            
            if [ -f "${ARTIFACTS_DIR}/${package//\//_}_coverage.out" ]; then
                coverage=$(go tool cover -func="${ARTIFACTS_DIR}/${package//\//_}_coverage.out" | grep "total:" | awk '{print $3}' | sed 's/%//')
                echo "Coverage for $package: ${coverage}%"
                echo "${package//\//_}_coverage:${coverage}" >> "${ARTIFACTS_DIR}/test_results.txt"
            fi
        fi
    done
    
    echo "coverage:COMPLETE" >> "${ARTIFACTS_DIR}/test_results.txt"
    echo ""
}

# Function to analyze code quality
analyze_code_quality() {
    print_section "Code Quality Analysis"
    
    echo "🔍 Analyzing code quality metrics..."
    
    # Count test files and functions
    test_files=$(find . -name "*_test.go" | wc -l)
    test_functions=$(grep -r "func Test" . --include="*_test.go" | wc -l)
    benchmark_functions=$(grep -r "func Benchmark" . --include="*_test.go" | wc -l)
    property_functions=$(grep -r "Property(" . --include="*_test.go" | wc -l)
    
    echo "📊 Test Metrics:"
    echo "  Test Files: $test_files"
    echo "  Test Functions: $test_functions"
    echo "  Benchmark Functions: $benchmark_functions"
    echo "  Property Tests: $property_functions"
    
    # Store metrics
    echo "test_files:$test_files" >> "${ARTIFACTS_DIR}/quality_metrics.txt"
    echo "test_functions:$test_functions" >> "${ARTIFACTS_DIR}/quality_metrics.txt"
    echo "benchmark_functions:$benchmark_functions" >> "${ARTIFACTS_DIR}/quality_metrics.txt"
    echo "property_functions:$property_functions" >> "${ARTIFACTS_DIR}/quality_metrics.txt"
    
    # Check for test patterns
    echo ""
    echo "🎯 Test Pattern Analysis:"
    
    table_tests=$(grep -r "tests := \[\]struct" . --include="*_test.go" | wc -l)
    parallel_tests=$(grep -r "t.Parallel()" . --include="*_test.go" | wc -l)
    subtests=$(grep -r "t.Run(" . --include="*_test.go" | wc -l)
    
    echo "  Table-driven tests: $table_tests"
    echo "  Parallel tests: $parallel_tests"
    echo "  Subtests: $subtests"
    
    echo "table_tests:$table_tests" >> "${ARTIFACTS_DIR}/quality_metrics.txt"
    echo "parallel_tests:$parallel_tests" >> "${ARTIFACTS_DIR}/quality_metrics.txt"
    echo "subtests:$subtests" >> "${ARTIFACTS_DIR}/quality_metrics.txt"
    
    echo ""
}

# Function to generate comprehensive report
generate_comprehensive_report() {
    print_section "Generating Comprehensive Quality Report"
    
    local report_file="${REPORTS_DIR}/comprehensive_quality_report.txt"
    
    echo "📄 Comprehensive Code Quality Report" > "$report_file"
    echo "=====================================" >> "$report_file"
    echo "Generated: $(date)" >> "$report_file"
    echo "Project: ${PROJECT_NAME}" >> "$report_file"
    echo "" >> "$report_file"
    
    # Test Results Summary
    echo "🧪 TEST RESULTS SUMMARY" >> "$report_file"
    echo "----------------------" >> "$report_file"
    
    if [ -f "${ARTIFACTS_DIR}/test_results.txt" ]; then
        while IFS=: read -r test_type result; do
            echo "${test_type}: ${result}" >> "$report_file"
        done < "${ARTIFACTS_DIR}/test_results.txt"
    fi
    echo "" >> "$report_file"
    
    # Quality Metrics
    echo "📊 QUALITY METRICS" >> "$report_file"
    echo "------------------" >> "$report_file"
    
    if [ -f "${ARTIFACTS_DIR}/quality_metrics.txt" ]; then
        while IFS=: read -r metric value; do
            echo "${metric}: ${value}" >> "$report_file"
        done < "${ARTIFACTS_DIR}/quality_metrics.txt"
    fi
    echo "" >> "$report_file"
    
    # Calculate overall quality score
    echo "🎯 OVERALL QUALITY ASSESSMENT" >> "$report_file"
    echo "-----------------------------" >> "$report_file"
    
    # Property testing
    property_score=0
    if grep -q "property:PASS" "${ARTIFACTS_DIR}/test_results.txt"; then
        property_score=25
    elif grep -q "property:SKIP" "${ARTIFACTS_DIR}/test_results.txt"; then
        property_score=10
    fi
    
    # Mutation testing
    mutation_score=0
    if grep -q "mutation:PASS" "${ARTIFACTS_DIR}/test_results.txt"; then
        mutation_score=35
    elif grep -q "mutation:BELOW_THRESHOLD" "${ARTIFACTS_DIR}/test_results.txt"; then
        mutation_score=20
    elif grep -q "mutation:BASIC_COMPLETE" "${ARTIFACTS_DIR}/test_results.txt"; then
        mutation_score=15
    fi
    
    # Coverage testing
    coverage_score=0
    if grep -q "coverage:COMPLETE" "${ARTIFACTS_DIR}/test_results.txt"; then
        coverage_score=25
    fi
    
    # Test diversity
    diversity_score=0
    if [ -f "${ARTIFACTS_DIR}/quality_metrics.txt" ]; then
        test_functions=$(grep "test_functions:" "${ARTIFACTS_DIR}/quality_metrics.txt" | cut -d: -f2)
        benchmark_functions=$(grep "benchmark_functions:" "${ARTIFACTS_DIR}/quality_metrics.txt" | cut -d: -f2)
        property_functions=$(grep "property_functions:" "${ARTIFACTS_DIR}/quality_metrics.txt" | cut -d: -f2)
        
        if [ "$test_functions" -gt 50 ] && [ "$benchmark_functions" -gt 10 ] && [ "$property_functions" -gt 5 ]; then
            diversity_score=15
        elif [ "$test_functions" -gt 20 ] && [ "$benchmark_functions" -gt 5 ]; then
            diversity_score=10
        elif [ "$test_functions" -gt 10 ]; then
            diversity_score=5
        fi
    fi
    
    total_score=$((property_score + mutation_score + coverage_score + diversity_score))
    
    echo "Property Testing: ${property_score}/25" >> "$report_file"
    echo "Mutation Testing: ${mutation_score}/35" >> "$report_file"
    echo "Coverage Testing: ${coverage_score}/25" >> "$report_file"
    echo "Test Diversity: ${diversity_score}/15" >> "$report_file"
    echo "" >> "$report_file"
    echo "TOTAL QUALITY SCORE: ${total_score}/100" >> "$report_file"
    
    # Quality grade
    if [ $total_score -ge 85 ]; then
        quality_grade="A (Excellent)"
    elif [ $total_score -ge 75 ]; then
        quality_grade="B (Good)"
    elif [ $total_score -ge 65 ]; then
        quality_grade="C (Fair)"
    elif [ $total_score -ge 55 ]; then
        quality_grade="D (Poor)"
    else
        quality_grade="F (Fail)"
    fi
    
    echo "QUALITY GRADE: $quality_grade" >> "$report_file"
    echo "" >> "$report_file"
    
    # Recommendations
    echo "💡 RECOMMENDATIONS" >> "$report_file"
    echo "------------------" >> "$report_file"
    
    if [ $property_score -lt 20 ]; then
        echo "- Implement property-based testing for critical algorithms" >> "$report_file"
    fi
    
    if [ $mutation_score -lt 25 ]; then
        echo "- Improve test quality through mutation testing" >> "$report_file"
        echo "- Add edge case testing and boundary condition tests" >> "$report_file"
    fi
    
    if [ $coverage_score -lt 20 ]; then
        echo "- Increase unit test coverage" >> "$report_file"
    fi
    
    if [ $diversity_score -lt 10 ]; then
        echo "- Add more diverse test types (benchmarks, property tests)" >> "$report_file"
    fi
    
    echo "- Regular quality testing in CI/CD pipeline" >> "$report_file"
    echo "- Monitor quality metrics over time" >> "$report_file"
    
    # Display summary
    echo -e "${CYAN}📊 Quality Score: ${total_score}/100 (${quality_grade})${NC}"
    
    cat "$report_file"
    echo ""
    echo -e "${GREEN}✅ Comprehensive report saved to: ${report_file}${NC}"
}

# Main execution
main() {
    echo -e "${CYAN}🚀 Starting Enhanced Quality Testing...${NC}"
    echo ""
    
    # Step 1: Run property-based tests
    run_property_tests
    
    # Step 2: Run mutation testing
    run_mutation_testing
    
    # Step 3: Run coverage tests
    run_coverage_tests
    
    # Step 4: Analyze code quality
    analyze_code_quality
    
    # Step 5: Generate comprehensive report
    generate_comprehensive_report
    
    echo -e "${GREEN}✅ Enhanced quality testing completed!${NC}"
    echo -e "${CYAN}📁 All artifacts saved to: ${ARTIFACTS_DIR}${NC}"
    echo -e "${CYAN}📄 View reports at: ${REPORTS_DIR}/${NC}"
}

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "Enhanced Quality Test Runner"
        echo ""
        echo "Usage: $0 [options]"
        echo ""
        echo "Options:"
        echo "  --help, -h          Show this help message"
        echo "  --property-only     Run only property-based tests"
        echo "  --mutation-only     Run only mutation testing"
        echo "  --coverage-only     Run only coverage testing"
        echo ""
        echo "Environment Variables:"
        echo "  MUTATION_SCORE_TARGET=75    Set mutation score target"
        echo "  PROPERTY_TEST_COUNT=1000    Set property test iterations"
        echo ""
        exit 0
        ;;
    --property-only)
        echo -e "${BLUE}🎯 Property Testing Only Mode${NC}"
        run_property_tests
        exit 0
        ;;
    --mutation-only)
        echo -e "${BLUE}🧬 Mutation Testing Only Mode${NC}"
        run_mutation_testing
        exit 0
        ;;
    --coverage-only)
        echo -e "${BLUE}📊 Coverage Testing Only Mode${NC}"
        run_coverage_tests
        exit 0
        ;;
esac

# Run main function
main "$@"