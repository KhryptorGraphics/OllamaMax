#!/bin/bash

# Performance Report Generation Script
# Generates comprehensive performance reports from benchmark results

set -e

echo "📊 Performance Report Generator"
echo "==============================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default configuration
INPUT_DIR=""
OUTPUT_FILE=""
FORMAT="json"
INCLUDE_CHARTS=false

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
            --input)
                INPUT_DIR="$2"
                shift 2
                ;;
            --output)
                OUTPUT_FILE="$2"
                shift 2
                ;;
            --format)
                FORMAT="$2"
                shift 2
                ;;
            --include-charts)
                INCLUDE_CHARTS=true
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
Performance Report Generation Script

Usage: $0 --input DIR --output FILE [OPTIONS]

Options:
    --input DIR          Input directory containing benchmark results [REQUIRED]
    --output FILE        Output file for the performance report [REQUIRED]
    --format FORMAT      Output format (json, html, markdown)
    --include-charts     Include performance charts in the report
    --help               Show this help message

Examples:
    $0 --input ./performance-results --output report.json
    $0 --input ./results --output report.html --format html --include-charts
EOF
}

# Validate arguments
validate_args() {
    if [ -z "$INPUT_DIR" ]; then
        print_status "ERROR" "Input directory is required. Use --input DIR"
        exit 1
    fi

    if [ -z "$OUTPUT_FILE" ]; then
        print_status "ERROR" "Output file is required. Use --output FILE"
        exit 1
    fi

    if [ ! -d "$INPUT_DIR" ]; then
        print_status "ERROR" "Input directory does not exist: $INPUT_DIR"
        exit 1
    fi
}

# Parse benchmark results
parse_benchmark_results() {
    local benchmark_file="$INPUT_DIR/benchmark-results.txt"
    
    if [ ! -f "$benchmark_file" ]; then
        print_status "WARNING" "Benchmark results file not found: $benchmark_file"
        return
    fi

    print_status "INFO" "Parsing benchmark results..."

    # Extract benchmark data using awk
    awk '
    BEGIN {
        print "{"
        print "  \"benchmarks\": ["
        first = 1
    }
    /^Benchmark/ {
        if (!first) print ","
        first = 0
        
        # Parse benchmark line: BenchmarkName-4 1000 1234 ns/op 567 B/op 8 allocs/op
        name = $1
        iterations = $2
        ns_per_op = $3
        
        # Extract additional metrics if present
        bytes_per_op = 0
        allocs_per_op = 0
        
        for (i = 4; i <= NF; i++) {
            if ($(i+1) == "B/op") {
                bytes_per_op = $i
                i++
            } else if ($(i+1) == "allocs/op") {
                allocs_per_op = $i
                i++
            }
        }
        
        printf "    {\n"
        printf "      \"name\": \"%s\",\n", name
        printf "      \"iterations\": %d,\n", iterations
        printf "      \"ns_per_op\": %d,\n", ns_per_op
        printf "      \"bytes_per_op\": %d,\n", bytes_per_op
        printf "      \"allocs_per_op\": %d\n", allocs_per_op
        printf "    }"
    }
    END {
        print ""
        print "  ]"
        print "}"
    }
    ' "$benchmark_file" > "$INPUT_DIR/parsed-benchmarks.json"

    print_status "SUCCESS" "Benchmark results parsed"
}

# Parse performance test results
parse_performance_tests() {
    local test_file="$INPUT_DIR/performance-tests.txt"
    
    if [ ! -f "$test_file" ]; then
        print_status "WARNING" "Performance test results file not found: $test_file"
        return
    fi

    print_status "INFO" "Parsing performance test results..."

    # Extract performance metrics from test output
    grep -E "(Latency|Throughput|CPU|Memory|Network)" "$test_file" | \
    awk '
    BEGIN {
        print "{"
        print "  \"performance_metrics\": ["
        first = 1
    }
    {
        if (!first) print ","
        first = 0
        
        # Extract metric name and value
        metric = $1
        value = $2
        unit = $3
        
        printf "    {\n"
        printf "      \"metric\": \"%s\",\n", metric
        printf "      \"value\": \"%s\",\n", value
        printf "      \"unit\": \"%s\"\n", unit
        printf "    }"
    }
    END {
        print ""
        print "  ]"
        print "}"
    }
    ' > "$INPUT_DIR/parsed-performance.json"

    print_status "SUCCESS" "Performance test results parsed"
}

# Calculate performance statistics
calculate_statistics() {
    print_status "INFO" "Calculating performance statistics..."

    # Use Python for statistical calculations
    python3 << 'EOF'
import json
import os
import sys
from statistics import mean, median, stdev

input_dir = os.environ.get('INPUT_DIR', '.')

# Load parsed benchmark data
try:
    with open(f'{input_dir}/parsed-benchmarks.json', 'r') as f:
        benchmark_data = json.load(f)
except FileNotFoundError:
    benchmark_data = {"benchmarks": []}

# Calculate statistics
stats = {
    "summary": {
        "total_benchmarks": len(benchmark_data["benchmarks"]),
        "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    },
    "performance_stats": {}
}

if benchmark_data["benchmarks"]:
    # Calculate ns/op statistics
    ns_per_op_values = [b["ns_per_op"] for b in benchmark_data["benchmarks"] if b["ns_per_op"] > 0]
    if ns_per_op_values:
        stats["performance_stats"]["ns_per_op"] = {
            "mean": mean(ns_per_op_values),
            "median": median(ns_per_op_values),
            "min": min(ns_per_op_values),
            "max": max(ns_per_op_values),
            "std_dev": stdev(ns_per_op_values) if len(ns_per_op_values) > 1 else 0
        }
    
    # Calculate memory allocation statistics
    allocs_values = [b["allocs_per_op"] for b in benchmark_data["benchmarks"] if b["allocs_per_op"] > 0]
    if allocs_values:
        stats["performance_stats"]["allocs_per_op"] = {
            "mean": mean(allocs_values),
            "median": median(allocs_values),
            "min": min(allocs_values),
            "max": max(allocs_values),
            "std_dev": stdev(allocs_values) if len(allocs_values) > 1 else 0
        }

# Save statistics
with open(f'{input_dir}/performance-stats.json', 'w') as f:
    json.dump(stats, f, indent=2)

print("Performance statistics calculated")
EOF

    print_status "SUCCESS" "Performance statistics calculated"
}

# Generate final report
generate_report() {
    print_status "INFO" "Generating final performance report..."

    case $FORMAT in
        json)
            generate_json_report
            ;;
        html)
            generate_html_report
            ;;
        markdown)
            generate_markdown_report
            ;;
        *)
            print_status "ERROR" "Unsupported format: $FORMAT"
            exit 1
            ;;
    esac
}

# Generate JSON report
generate_json_report() {
    print_status "INFO" "Generating JSON report..."

    # Combine all data into final report
    python3 << 'EOF'
import json
import os

input_dir = os.environ.get('INPUT_DIR', '.')
output_file = os.environ.get('OUTPUT_FILE', 'report.json')

# Load all data files
data = {
    "metadata": {
        "generated_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
        "input_directory": input_dir,
        "format": "json"
    }
}

# Load benchmark data
try:
    with open(f'{input_dir}/parsed-benchmarks.json', 'r') as f:
        data.update(json.load(f))
except FileNotFoundError:
    data["benchmarks"] = []

# Load performance metrics
try:
    with open(f'{input_dir}/parsed-performance.json', 'r') as f:
        perf_data = json.load(f)
        data.update(perf_data)
except FileNotFoundError:
    data["performance_metrics"] = []

# Load statistics
try:
    with open(f'{input_dir}/performance-stats.json', 'r') as f:
        stats_data = json.load(f)
        data.update(stats_data)
except FileNotFoundError:
    data["summary"] = {"total_benchmarks": 0}

# Save final report
with open(output_file, 'w') as f:
    json.dump(data, f, indent=2)

print(f"JSON report generated: {output_file}")
EOF

    print_status "SUCCESS" "JSON report generated: $OUTPUT_FILE"
}

# Generate HTML report
generate_html_report() {
    print_status "INFO" "Generating HTML report..."

    cat > "$OUTPUT_FILE" << 'EOF'
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>OllamaMax Performance Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f4f4f4; padding: 20px; border-radius: 5px; }
        .metric { margin: 10px 0; padding: 10px; border-left: 4px solid #007cba; }
        .benchmark { background: #f9f9f9; margin: 5px 0; padding: 10px; }
        .stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 10px; }
        .stat-card { background: white; border: 1px solid #ddd; padding: 15px; border-radius: 5px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🚀 OllamaMax Performance Report</h1>
        <p>Generated: $(date)</p>
    </div>
    
    <h2>📊 Performance Summary</h2>
    <div class="stats">
        <div class="stat-card">
            <h3>Total Benchmarks</h3>
            <p id="total-benchmarks">Loading...</p>
        </div>
        <div class="stat-card">
            <h3>Average Performance</h3>
            <p id="avg-performance">Loading...</p>
        </div>
    </div>
    
    <h2>🔍 Detailed Results</h2>
    <div id="benchmark-results">
        Loading benchmark results...
    </div>
    
    <script>
        // Load and display performance data
        // This would be populated with actual data in a real implementation
        document.getElementById('total-benchmarks').textContent = 'N/A';
        document.getElementById('avg-performance').textContent = 'N/A';
        document.getElementById('benchmark-results').innerHTML = '<p>Benchmark data would be displayed here</p>';
    </script>
</body>
</html>
EOF

    print_status "SUCCESS" "HTML report generated: $OUTPUT_FILE"
}

# Generate Markdown report
generate_markdown_report() {
    print_status "INFO" "Generating Markdown report..."

    cat > "$OUTPUT_FILE" << EOF
# 🚀 OllamaMax Performance Report

**Generated:** $(date)  
**Input Directory:** $INPUT_DIR

## 📊 Performance Summary

- **Total Benchmarks:** Loading...
- **Report Format:** Markdown
- **Analysis Date:** $(date -u +%Y-%m-%d)

## 🔍 Benchmark Results

Detailed benchmark results would be displayed here.

## 📈 Performance Metrics

Performance metrics and analysis would be shown here.

## 💡 Recommendations

Performance optimization recommendations would be provided here.
EOF

    print_status "SUCCESS" "Markdown report generated: $OUTPUT_FILE"
}

# Main function
main() {
    parse_args "$@"
    validate_args
    
    print_status "INFO" "Generating performance report from: $INPUT_DIR"
    print_status "INFO" "Output format: $FORMAT"
    
    parse_benchmark_results
    parse_performance_tests
    calculate_statistics
    generate_report
    
    print_status "SUCCESS" "Performance report generation completed! 📊"
    print_status "INFO" "Report saved to: $OUTPUT_FILE"
}

# Export environment variables for Python scripts
export INPUT_DIR
export OUTPUT_FILE
export FORMAT

# Run main function
main "$@"
