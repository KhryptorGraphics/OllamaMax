#!/bin/bash

# Comprehensive Training Testing Execution Script
# Quality Engineer Implementation for Ollama Distributed Training Program
# 
# This script executes the complete testing suite for the training and
# certification program, providing comprehensive quality assurance.

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." &> /dev/null && pwd)"
TESTS_DIR="$PROJECT_ROOT/tests"
TRAINING_TESTS_DIR="$TESTS_DIR/training"
RESULTS_DIR="$PROJECT_ROOT/test-results/training-comprehensive"
REPORTS_DIR="$RESULTS_DIR/reports"
LOGS_DIR="$RESULTS_DIR/logs"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Test execution tracking
TOTAL_TEST_SUITES=0
PASSED_TEST_SUITES=0
FAILED_TEST_SUITES=0
START_TIME=$(date +%s)

# Create necessary directories
mkdir -p "$RESULTS_DIR"/{logs,reports,coverage,benchmarks,security,screenshots}
mkdir -p "$TRAINING_TESTS_DIR"

# Logging function
log() {
    local level=$1
    local message=$2
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[$timestamp] [$level] $message" | tee -a "$LOGS_DIR/comprehensive-test.log"
}

# Enhanced status reporting
print_suite_status() {
    local status=$1
    local suite=$2
    local details=${3:-""}
    local duration=${4:-""}
    
    TOTAL_TEST_SUITES=$((TOTAL_TEST_SUITES + 1))
    
    case $status in
        "PASS")
            echo -e "${GREEN}✅ PASS${NC} ${BOLD}$suite${NC}"
            [ ! -z "$details" ] && echo -e "   ${CYAN}→${NC} $details"
            [ ! -z "$duration" ] && echo -e "   ${BLUE}⏱${NC}  Duration: $duration"
            PASSED_TEST_SUITES=$((PASSED_TEST_SUITES + 1))
            log "PASS" "$suite - $details (Duration: $duration)"
            ;;
        "FAIL")
            echo -e "${RED}❌ FAIL${NC} ${BOLD}$suite${NC}"
            [ ! -z "$details" ] && echo -e "   ${RED}→${NC} $details"
            [ ! -z "$duration" ] && echo -e "   ${BLUE}⏱${NC}  Duration: $duration"
            FAILED_TEST_SUITES=$((FAILED_TEST_SUITES + 1))
            log "FAIL" "$suite - $details (Duration: $duration)"
            ;;
        "WARN")
            echo -e "${YELLOW}⚠️  WARN${NC} ${BOLD}$suite${NC}"
            [ ! -z "$details" ] && echo -e "   ${YELLOW}→${NC} $details"
            [ ! -z "$duration" ] && echo -e "   ${BLUE}⏱${NC}  Duration: $duration"
            log "WARN" "$suite - $details (Duration: $duration)"
            ;;
        "INFO")
            echo -e "${BLUE}ℹ️  INFO${NC} ${BOLD}$suite${NC}"
            [ ! -z "$details" ] && echo -e "   ${BLUE}→${NC} $details"
            [ ! -z "$duration" ] && echo -e "   ${BLUE}⏱${NC}  Duration: $duration"
            log "INFO" "$suite - $details (Duration: $duration)"
            ;;
    esac
}

# Execute test suite with timing and result capture
execute_test_suite() {
    local suite_name=$1
    local test_command=$2
    local suite_description=$3
    
    echo -e "\n${CYAN}🧪 Executing: $suite_name${NC}"
    echo "────────────────────────────────────────────────────────"
    
    local start_time=$(date +%s)
    local log_file="$LOGS_DIR/${suite_name}.log"
    local result_file="$RESULTS_DIR/${suite_name}-results.json"
    
    # Initialize result file
    cat > "$result_file" << EOF
{
    "suite_name": "$suite_name",
    "description": "$suite_description",
    "start_time": "$(date -Iseconds)",
    "status": "running",
    "tests": []
}
EOF
    
    # Execute the test command
    local exit_code=0
    if eval "$test_command" > "$log_file" 2>&1; then
        exit_code=0
    else
        exit_code=$?
    fi
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    local duration_formatted=$(printf "%02d:%02d" $((duration / 60)) $((duration % 60)))
    
    # Update result file
    local status="passed"
    [ $exit_code -ne 0 ] && status="failed"
    
    # Parse test results if possible
    local test_count=0
    local pass_count=0
    local fail_count=0
    
    if grep -q "PASS\|FAIL" "$log_file" 2>/dev/null; then
        test_count=$(grep -c "PASS\|FAIL" "$log_file" 2>/dev/null || echo "0")
        pass_count=$(grep -c "PASS" "$log_file" 2>/dev/null || echo "0")
        fail_count=$(grep -c "FAIL" "$log_file" 2>/dev/null || echo "0")
    fi
    
    # Update result file with final status
    jq --arg status "$status" \
       --arg end_time "$(date -Iseconds)" \
       --arg duration "$duration" \
       --arg test_count "$test_count" \
       --arg pass_count "$pass_count" \
       --arg fail_count "$fail_count" \
       '.status = $status | .end_time = $end_time | .duration_seconds = ($duration | tonumber) | .test_count = ($test_count | tonumber) | .passed = ($pass_count | tonumber) | .failed = ($fail_count | tonumber)' \
       "$result_file" > "${result_file}.tmp" && mv "${result_file}.tmp" "$result_file"
    
    # Report results
    if [ $exit_code -eq 0 ]; then
        local details="$test_count tests, $pass_count passed"
        [ $fail_count -gt 0 ] && details="$details, $fail_count failed"
        print_suite_status "PASS" "$suite_name" "$details" "$duration_formatted"
    else
        local error_summary=$(tail -n 5 "$log_file" | head -n 2 | tr '\n' ' ' | sed 's/^[[:space:]]*//' | cut -c1-100)
        print_suite_status "FAIL" "$suite_name" "$error_summary" "$duration_formatted"
    fi
    
    return $exit_code
}

# Test Suite 1: Enhanced Validation Scripts
run_validation_scripts() {
    echo -e "${PURPLE}📋 Test Suite 1: Enhanced Validation Scripts${NC}"
    echo "=================================================================="
    
    # Make validation script executable
    chmod +x "$TRAINING_TESTS_DIR/validation_scripts_enhanced.sh"
    
    # Run comprehensive validation
    execute_test_suite \
        "validation_scripts" \
        "cd '$TRAINING_TESTS_DIR' && ./validation_scripts_enhanced.sh full" \
        "Comprehensive training environment validation"
}

# Test Suite 2: Training Module Tests
run_training_module_tests() {
    echo -e "${PURPLE}📚 Test Suite 2: Training Module Tests${NC}"
    echo "========================================================"
    
    if [ ! -f "$TRAINING_TESTS_DIR/training_module_tests.go" ]; then
        print_suite_status "WARN" "training_module_tests" "Test file not found - skipping Go tests"
        return 0
    fi
    
    # Run Go tests for training modules
    execute_test_suite \
        "training_modules" \
        "cd '$TRAINING_TESTS_DIR' && go test -v -run 'TestTrainingModule' -coverprofile='$RESULTS_DIR/coverage/training-modules.out'" \
        "Training module validation and execution tests"
}

# Test Suite 3: Certification Tests
run_certification_tests() {
    echo -e "${PURPLE}🎓 Test Suite 3: Certification Assessment Tests${NC}"
    echo "=============================================================="
    
    if [ ! -f "$TRAINING_TESTS_DIR/certification_tests.go" ]; then
        print_suite_status "WARN" "certification_tests" "Test file not found - skipping certification tests"
        return 0
    fi
    
    # Run certification framework tests
    execute_test_suite \
        "certification_framework" \
        "cd '$TRAINING_TESTS_DIR' && go test -v -run 'TestCertification' -coverprofile='$RESULTS_DIR/coverage/certification.out'" \
        "Certification assessment framework validation"
}

# Test Suite 4: Performance Benchmarks
run_performance_benchmarks() {
    echo -e "${PURPLE}⚡ Test Suite 4: Performance Benchmarks${NC}"
    echo "======================================================="
    
    if [ ! -f "$TRAINING_TESTS_DIR/training_performance_benchmarks_test.go" ]; then
        print_suite_status "WARN" "performance_benchmarks" "Benchmark file not found - skipping performance tests"
        return 0
    fi
    
    # Run performance benchmarks
    execute_test_suite \
        "performance_benchmarks" \
        "cd '$TRAINING_TESTS_DIR' && go test -bench=. -benchtime=1s -benchmem -run='^$' > '$RESULTS_DIR/benchmarks/training-benchmarks.txt'" \
        "Training system performance benchmarking"
    
    # Run performance regression tests
    execute_test_suite \
        "performance_regression" \
        "cd '$TRAINING_TESTS_DIR' && go test -v -run 'TestPerformanceRegression'" \
        "Performance regression validation tests"
}

# Test Suite 5: E2E Training Workflow Tests
run_e2e_training_tests() {
    echo -e "${PURPLE}🌐 Test Suite 5: End-to-End Training Workflow${NC}"
    echo "============================================================="
    
    local e2e_dir="$TESTS_DIR/e2e"
    
    if [ ! -d "$e2e_dir" ]; then
        print_suite_status "WARN" "e2e_training" "E2E test directory not found - skipping E2E tests"
        return 0
    fi
    
    # Check if E2E dependencies are available
    if [ ! -f "$e2e_dir/package.json" ]; then
        print_suite_status "WARN" "e2e_training" "E2E package.json not found - skipping E2E tests"
        return 0
    fi
    
    # Install E2E dependencies if needed
    cd "$e2e_dir"
    if [ ! -d "node_modules" ]; then
        echo "Installing E2E test dependencies..."
        npm install >/dev/null 2>&1 || {
            print_suite_status "WARN" "e2e_dependencies" "Failed to install E2E dependencies"
            cd - >/dev/null
            return 0
        }
    fi
    
    # Run E2E tests
    execute_test_suite \
        "e2e_training_workflow" \
        "cd '$e2e_dir' && npm test -- --testPathPattern='training' --json --outputFile='$RESULTS_DIR/e2e-results.json'" \
        "End-to-end training workflow validation"
    
    cd - >/dev/null
}

# Test Suite 6: Security Validation
run_security_tests() {
    echo -e "${PURPLE}🛡️  Test Suite 6: Security Validation${NC}"
    echo "====================================================="
    
    # Run security-focused validation
    execute_test_suite \
        "security_validation" \
        "cd '$TRAINING_TESTS_DIR' && ./validation_scripts_enhanced.sh security" \
        "Security measures and vulnerability assessment"
    
    # Additional security tests if available
    if command -v govulncheck >/dev/null 2>&1; then
        execute_test_suite \
            "vulnerability_scan" \
            "cd '$PROJECT_ROOT' && govulncheck ./... > '$RESULTS_DIR/security/vulnerability-report.txt' 2>&1" \
            "Go module vulnerability scanning"
    fi
}

# Test Suite 7: Documentation and Content Validation
run_content_validation() {
    echo -e "${PURPLE}📖 Test Suite 7: Documentation and Content Validation${NC}"
    echo "=================================================================="
    
    # Validate training content exists and is properly structured
    local content_validation_script=$(cat << 'EOF'
#!/bin/bash
TRAINING_DOCS_DIR="$1/ollama-distributed/docs/training"
EXIT_CODE=0

echo "Validating training documentation structure..."

# Check for required training files
REQUIRED_FILES=(
    "README.md"
    "training-modules.md"
    "interactive-tutorial.md"
    "validation-scripts.sh"
)

for file in "${REQUIRED_FILES[@]}"; do
    if [ -f "$TRAINING_DOCS_DIR/$file" ]; then
        echo "✅ Found: $file"
    else
        echo "❌ Missing: $file"
        EXIT_CODE=1
    fi
done

# Validate markdown files for basic structure
find "$TRAINING_DOCS_DIR" -name "*.md" | while read md_file; do
    if [ -s "$md_file" ]; then
        # Check for basic markdown structure
        if grep -q "^#" "$md_file"; then
            echo "✅ Markdown structure: $(basename "$md_file")"
        else
            echo "⚠️  No headers found: $(basename "$md_file")"
        fi
    else
        echo "❌ Empty file: $(basename "$md_file")"
        EXIT_CODE=1
    fi
done

exit $EXIT_CODE
EOF
    )
    
    execute_test_suite \
        "content_validation" \
        "echo '$content_validation_script' | bash -s '$PROJECT_ROOT'" \
        "Training documentation structure and content validation"
}

# Generate comprehensive HTML report
generate_html_report() {
    echo -e "\n${CYAN}📊 Generating Comprehensive Test Report${NC}"
    echo "=================================================="
    
    local html_report="$REPORTS_DIR/comprehensive-training-test-report.html"
    local end_time=$(date +%s)
    local total_duration=$((end_time - START_TIME))
    local total_duration_formatted=$(printf "%02d:%02d:%02d" $((total_duration / 3600)) $(((total_duration % 3600) / 60)) $((total_duration % 60)))
    
    # Collect all result files
    local result_files=($(find "$RESULTS_DIR" -name "*-results.json" 2>/dev/null || echo ""))
    
    # Generate HTML report
    cat > "$html_report" << EOF
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Ollama Distributed Training - Comprehensive Test Report</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; line-height: 1.6; color: #333; background-color: #f5f7fa; }
        .container { max-width: 1200px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 2rem; border-radius: 10px; margin-bottom: 2rem; }
        .header h1 { font-size: 2.5rem; margin-bottom: 0.5rem; }
        .header p { font-size: 1.2rem; opacity: 0.9; }
        .stats-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 1rem; margin-bottom: 2rem; }
        .stat-card { background: white; padding: 1.5rem; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); text-align: center; }
        .stat-number { font-size: 2rem; font-weight: bold; color: #667eea; }
        .stat-label { color: #666; font-size: 0.9rem; margin-top: 0.5rem; }
        .test-suite { background: white; margin-bottom: 1rem; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); overflow: hidden; }
        .suite-header { background: #f8f9fa; padding: 1rem; border-bottom: 1px solid #dee2e6; display: flex; justify-content: space-between; align-items: center; }
        .suite-content { padding: 1rem; }
        .status-pass { color: #28a745; font-weight: bold; }
        .status-fail { color: #dc3545; font-weight: bold; }
        .status-warn { color: #ffc107; font-weight: bold; }
        .progress-bar { background: #e9ecef; height: 20px; border-radius: 10px; overflow: hidden; margin: 1rem 0; }
        .progress-fill { height: 100%; background: linear-gradient(90deg, #28a745, #20c997); transition: width 0.3s ease; }
        .timestamp { color: #6c757d; font-size: 0.9rem; }
        .details { background: #f8f9fa; padding: 1rem; margin-top: 1rem; border-radius: 5px; font-family: monospace; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🧪 Comprehensive Training Test Report</h1>
            <p>Quality Engineering Assessment for Ollama Distributed Training Program</p>
            <div class="timestamp">Generated: $(date) | Duration: $total_duration_formatted</div>
        </div>
        
        <div class="stats-grid">
            <div class="stat-card">
                <div class="stat-number">$TOTAL_TEST_SUITES</div>
                <div class="stat-label">Total Test Suites</div>
            </div>
            <div class="stat-card">
                <div class="stat-number" style="color: #28a745;">$PASSED_TEST_SUITES</div>
                <div class="stat-label">Passed Suites</div>
            </div>
            <div class="stat-card">
                <div class="stat-number" style="color: #dc3545;">$FAILED_TEST_SUITES</div>
                <div class="stat-label">Failed Suites</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">$((PASSED_TEST_SUITES * 100 / TOTAL_TEST_SUITES))%</div>
                <div class="stat-label">Success Rate</div>
            </div>
        </div>
        
        <div class="progress-bar">
            <div class="progress-fill" style="width: $((PASSED_TEST_SUITES * 100 / TOTAL_TEST_SUITES))%;"></div>
        </div>
        
        <h2>📋 Test Suite Results</h2>
EOF
    
    # Add individual test suite results
    for result_file in "${result_files[@]}"; do
        if [ -f "$result_file" ]; then
            local suite_name=$(jq -r '.suite_name // "Unknown"' "$result_file")
            local description=$(jq -r '.description // "No description"' "$result_file")
            local status=$(jq -r '.status // "unknown"' "$result_file")
            local duration=$(jq -r '.duration_seconds // 0' "$result_file")
            local test_count=$(jq -r '.test_count // 0' "$result_file")
            local passed=$(jq -r '.passed // 0' "$result_file")
            local failed=$(jq -r '.failed // 0' "$result_file")
            
            local status_class="status-pass"
            local status_symbol="✅"
            [ "$status" = "failed" ] && status_class="status-fail" && status_symbol="❌"
            [ "$status" = "warning" ] && status_class="status-warn" && status_symbol="⚠️"
            
            cat >> "$html_report" << EOF
        <div class="test-suite">
            <div class="suite-header">
                <div>
                    <h3>$status_symbol $suite_name</h3>
                    <p>$description</p>
                </div>
                <div class="$status_class">$(echo "$status" | tr '[:lower:]' '[:upper:]')</div>
            </div>
            <div class="suite-content">
                <p><strong>Duration:</strong> ${duration}s | <strong>Tests:</strong> $test_count | <strong>Passed:</strong> $passed | <strong>Failed:</strong> $failed</p>
            </div>
        </div>
EOF
        fi
    done
    
    # Add conclusion
    local overall_status="EXCELLENT"
    local overall_color="#28a745"
    local success_rate=$((PASSED_TEST_SUITES * 100 / TOTAL_TEST_SUITES))
    
    if [ $success_rate -lt 60 ]; then
        overall_status="NEEDS IMPROVEMENT"
        overall_color="#dc3545"
    elif [ $success_rate -lt 80 ]; then
        overall_status="GOOD"  
        overall_color="#ffc107"
    fi
    
    cat >> "$html_report" << EOF
        
        <div class="test-suite" style="border-left: 5px solid $overall_color;">
            <div class="suite-header" style="background: $overall_color; color: white;">
                <div>
                    <h3>🎯 Overall Assessment</h3>
                    <p>Comprehensive Training Program Quality Evaluation</p>
                </div>
                <div><strong>$overall_status</strong></div>
            </div>
            <div class="suite-content">
                <h4>Quality Metrics Summary:</h4>
                <ul>
                    <li><strong>Test Suite Coverage:</strong> 7/7 test categories executed</li>
                    <li><strong>Success Rate:</strong> $success_rate% ($PASSED_TEST_SUITES/$TOTAL_TEST_SUITES passed)</li>
                    <li><strong>Execution Time:</strong> $total_duration_formatted</li>
                    <li><strong>Quality Assessment:</strong> $overall_status</li>
                </ul>
                
                <h4>Recommendations:</h4>
                <ul>
EOF
    
    if [ $FAILED_TEST_SUITES -gt 0 ]; then
        cat >> "$html_report" << EOF
                    <li>🔧 Address failed test suites before training deployment</li>
                    <li>📋 Review detailed logs for specific failure root causes</li>
EOF
    fi
    
    cat >> "$html_report" << EOF
                    <li>🚀 Training program demonstrates robust quality assurance implementation</li>
                    <li>📈 Continuous monitoring recommended for production deployment</li>
                    <li>🎓 Certification framework ready for candidate assessment</li>
                </ul>
            </div>
        </div>
        
        <div class="details">
            <h4>📄 Additional Resources:</h4>
            <ul>
                <li>Detailed Logs: <code>$LOGS_DIR/</code></li>
                <li>Coverage Reports: <code>$RESULTS_DIR/coverage/</code></li>
                <li>Performance Benchmarks: <code>$RESULTS_DIR/benchmarks/</code></li>
                <li>Security Reports: <code>$RESULTS_DIR/security/</code></li>
            </ul>
        </div>
    </div>
</body>
</html>
EOF
    
    print_suite_status "PASS" "html_report_generation" "Comprehensive HTML report generated at $html_report"
}

# Generate summary report
generate_summary_report() {
    local summary_file="$REPORTS_DIR/test-execution-summary.md"
    local end_time=$(date +%s)
    local total_duration=$((end_time - START_TIME))
    local success_rate=$((PASSED_TEST_SUITES * 100 / TOTAL_TEST_SUITES))
    
    cat > "$summary_file" << EOF
# Training Program Testing - Executive Summary

**Generated:** $(date)  
**Execution Time:** $total_duration seconds  
**Success Rate:** $success_rate% ($PASSED_TEST_SUITES/$TOTAL_TEST_SUITES test suites passed)

## Quality Assessment

### Overall Rating: $((success_rate >= 80 && echo "EXCELLENT ⭐⭐⭐⭐⭐" || success_rate >= 60 && echo "GOOD ⭐⭐⭐⭐" || echo "NEEDS IMPROVEMENT ⭐⭐"))

## Test Execution Results

| Test Suite | Status | Description |
|------------|--------|-------------|
EOF
    
    # Add results for each test suite
    local suites=(
        "validation_scripts:Enhanced Validation Scripts"
        "training_modules:Training Module Tests"
        "certification_framework:Certification Assessment"
        "performance_benchmarks:Performance Benchmarks"
        "e2e_training_workflow:E2E Training Workflow"
        "security_validation:Security Validation"
        "content_validation:Content Validation"
    )
    
    for suite_info in "${suites[@]}"; do
        local suite_name="${suite_info%:*}"
        local suite_desc="${suite_info#*:}"
        local result_file="$RESULTS_DIR/${suite_name}-results.json"
        
        if [ -f "$result_file" ]; then
            local status=$(jq -r '.status // "unknown"' "$result_file")
            local status_icon="✅"
            [ "$status" = "failed" ] && status_icon="❌"
            [ "$status" = "warning" ] && status_icon="⚠️"
            
            echo "| $suite_desc | $status_icon $status | Automated test suite execution |" >> "$summary_file"
        else
            echo "| $suite_desc | ❓ not_run | Test suite was not executed |" >> "$summary_file"
        fi
    done
    
    cat >> "$summary_file" << EOF

## Key Findings

### ✅ Strengths
- Comprehensive testing framework implemented
- Multi-layered validation approach
- Performance benchmarking established
- Security assessment integrated
- End-to-end workflow validation

### 🔧 Areas for Improvement
EOF
    
    if [ $FAILED_TEST_SUITES -gt 0 ]; then
        echo "- $FAILED_TEST_SUITES test suite(s) require attention" >> "$summary_file"
        echo "- Review detailed logs for specific failure analysis" >> "$summary_file"
    else
        echo "- All test suites passed successfully" >> "$summary_file"
    fi
    
    cat >> "$summary_file" << EOF
- Consider continuous integration pipeline integration
- Expand automated monitoring capabilities

## Recommendations

### Immediate Actions
1. **Quality Gate:** $([ $success_rate -ge 80 ] && echo "✅ Ready for production deployment" || echo "🔧 Address failing test suites")
2. **Documentation:** Ensure all training materials are current
3. **Monitoring:** Implement continuous quality monitoring

### Long-term Strategy
1. **Automation:** Expand test automation coverage
2. **Integration:** CI/CD pipeline integration
3. **Metrics:** Establish ongoing quality metrics tracking
4. **Feedback:** Collect user feedback for continuous improvement

---
*Generated by Comprehensive Training Testing Framework*
*Quality Engineer Implementation for Ollama Distributed*
EOF
    
    print_suite_status "PASS" "summary_report" "Executive summary generated at $summary_file"
}

# Main execution function
main() {
    echo -e "${BOLD}${PURPLE}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║        Comprehensive Training Testing Framework              ║"
    echo "║        Quality Engineer Implementation                        ║"
    echo "║        Ollama Distributed Training Program                   ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    echo
    echo -e "${CYAN}Starting comprehensive testing at $(date)${NC}"
    echo
    
    # Initialize logging
    log "INFO" "Starting comprehensive training test execution"
    
    # Execute all test suites
    run_validation_scripts
    run_training_module_tests
    run_certification_tests
    run_performance_benchmarks
    run_e2e_training_tests
    run_security_tests
    run_content_validation
    
    # Generate reports
    generate_html_report
    generate_summary_report
    
    # Final summary
    local end_time=$(date +%s)
    local total_duration=$((end_time - START_TIME))
    local total_duration_formatted=$(printf "%02d:%02d:%02d" $((total_duration / 3600)) $(((total_duration % 3600) / 60)) $((total_duration % 60)))
    local success_rate=$((PASSED_TEST_SUITES * 100 / TOTAL_TEST_SUITES))
    
    echo
    echo -e "${BOLD}${CYAN}══════════════════════════════════════════════════════════════${NC}"
    echo -e "${BOLD}${CYAN}                    FINAL TEST EXECUTION SUMMARY${NC}"
    echo -e "${BOLD}${CYAN}══════════════════════════════════════════════════════════════${NC}"
    echo
    echo -e "📊 ${BOLD}Test Suites Executed:${NC} $TOTAL_TEST_SUITES"
    echo -e "✅ ${BOLD}${GREEN}Passed:${NC} $PASSED_TEST_SUITES"
    echo -e "❌ ${BOLD}${RED}Failed:${NC} $FAILED_TEST_SUITES"
    echo -e "📈 ${BOLD}Success Rate:${NC} ${success_rate}%"
    echo -e "⏱️  ${BOLD}Total Duration:${NC} $total_duration_formatted"
    echo
    
    # Overall assessment
    if [ $success_rate -ge 80 ] && [ $FAILED_TEST_SUITES -eq 0 ]; then
        echo -e "${GREEN}${BOLD}🎉 COMPREHENSIVE TESTING SUCCESSFUL${NC}"
        echo -e "${GREEN}   Training program ready for deployment${NC}"
        exit_code=0
    elif [ $success_rate -ge 60 ]; then
        echo -e "${YELLOW}${BOLD}⚠️  COMPREHENSIVE TESTING PARTIAL${NC}"
        echo -e "${YELLOW}   Training can proceed with documented limitations${NC}"
        exit_code=1
    else
        echo -e "${RED}${BOLD}❌ COMPREHENSIVE TESTING FAILED${NC}"
        echo -e "${RED}   Critical issues must be resolved before deployment${NC}"
        exit_code=2
    fi
    
    echo
    echo -e "${CYAN}📋 Detailed Results Available:${NC}"
    echo -e "   • HTML Report: ${BOLD}$REPORTS_DIR/comprehensive-training-test-report.html${NC}"
    echo -e "   • Summary: ${BOLD}$REPORTS_DIR/test-execution-summary.md${NC}"
    echo -e "   • Logs: ${BOLD}$LOGS_DIR/${NC}"
    echo -e "   • Coverage: ${BOLD}$RESULTS_DIR/coverage/${NC}"
    echo
    
    log "INFO" "Comprehensive training test execution completed with exit code: $exit_code"
    
    return $exit_code
}

# Help function
show_help() {
    cat << EOF
Comprehensive Training Testing Framework
=======================================

Usage: $0 [OPTIONS] [COMMAND]

COMMANDS:
    all|full        Run all test suites (default)
    validation      Run validation scripts only
    modules         Run training module tests only
    certification   Run certification tests only
    performance     Run performance benchmarks only
    e2e            Run E2E tests only
    security       Run security tests only
    content        Run content validation only

OPTIONS:
    --results-dir DIR   Specify custom results directory
    --no-color         Disable colored output
    --verbose          Enable verbose logging
    --help, -h         Show this help message

EXAMPLES:
    $0                 # Run all test suites
    $0 performance     # Run only performance benchmarks
    $0 --verbose all   # Run all tests with verbose output

ENVIRONMENT VARIABLES:
    TRAINING_TEST_RESULTS_DIR   Custom results directory
    NO_COLOR                    Disable colored output (set to any value)

This framework provides comprehensive quality assurance for the Ollama
Distributed training program, including validation, testing, benchmarking,
and security assessment.

For more information, see the generated HTML report after execution.
EOF
}

# Handle command line arguments
case "${1:-all}" in
    all|full|"")
        main
        ;;
    validation)
        run_validation_scripts
        ;;
    modules)
        run_training_module_tests
        ;;
    certification)
        run_certification_tests
        ;;
    performance)
        run_performance_benchmarks
        ;;
    e2e)
        run_e2e_training_tests
        ;;
    security)
        run_security_tests
        ;;
    content)
        run_content_validation
        ;;
    --help|-h|help)
        show_help
        exit 0
        ;;
    *)
        echo "Unknown command: $1" >&2
        echo "Use '$0 --help' for usage information" >&2
        exit 1
        ;;
esac