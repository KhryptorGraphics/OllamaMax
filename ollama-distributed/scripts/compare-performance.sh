#!/bin/bash

# Performance Comparison Script
# Compares current performance against baseline and detects regressions

set -e

echo "🔍 Performance Comparison Analysis"
echo "=================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default configuration
CURRENT_FILE=""
BASELINE_FILE=""
THRESHOLD=10
OUTPUT_FILE=""
VERBOSE=false

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
            --current)
                CURRENT_FILE="$2"
                shift 2
                ;;
            --baseline)
                BASELINE_FILE="$2"
                shift 2
                ;;
            --threshold)
                THRESHOLD="$2"
                shift 2
                ;;
            --output)
                OUTPUT_FILE="$2"
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
Performance Comparison Script

Usage: $0 --current FILE --baseline FILE [OPTIONS]

Options:
    --current FILE       Current performance report file [REQUIRED]
    --baseline FILE      Baseline performance report file [REQUIRED]
    --threshold PERCENT  Regression threshold percentage (default: 10)
    --output FILE        Output file for comparison results
    --verbose            Enable verbose output
    --help               Show this help message

Examples:
    $0 --current current.json --baseline baseline.json
    $0 --current current.json --baseline baseline.json --threshold 5 --output comparison.json
EOF
}

# Validate arguments
validate_args() {
    if [ -z "$CURRENT_FILE" ]; then
        print_status "ERROR" "Current performance file is required. Use --current FILE"
        exit 1
    fi

    if [ -z "$BASELINE_FILE" ]; then
        print_status "ERROR" "Baseline performance file is required. Use --baseline FILE"
        exit 1
    fi

    if [ ! -f "$CURRENT_FILE" ]; then
        print_status "ERROR" "Current performance file does not exist: $CURRENT_FILE"
        exit 1
    fi

    if [ ! -f "$BASELINE_FILE" ]; then
        print_status "ERROR" "Baseline performance file does not exist: $BASELINE_FILE"
        exit 1
    fi

    if [ -z "$OUTPUT_FILE" ]; then
        OUTPUT_FILE="./performance-comparison-$(date +%Y%m%d-%H%M%S).json"
    fi
}

# Compare performance data
compare_performance() {
    print_status "INFO" "Comparing performance data..."
    print_status "INFO" "Current: $CURRENT_FILE"
    print_status "INFO" "Baseline: $BASELINE_FILE"
    print_status "INFO" "Threshold: ${THRESHOLD}%"

    # Use Python for detailed comparison
    python3 << 'EOF'
import json
import sys
import os
from datetime import datetime

# Get environment variables
current_file = os.environ.get('CURRENT_FILE')
baseline_file = os.environ.get('BASELINE_FILE')
threshold = float(os.environ.get('THRESHOLD', '10'))
output_file = os.environ.get('OUTPUT_FILE')
verbose = os.environ.get('VERBOSE', 'false').lower() == 'true'

def load_performance_data(filename):
    """Load performance data from JSON file"""
    try:
        with open(filename, 'r') as f:
            return json.load(f)
    except Exception as e:
        print(f"Error loading {filename}: {e}")
        return None

def compare_benchmarks(current_benchmarks, baseline_benchmarks):
    """Compare benchmark results"""
    comparisons = []
    regressions = []
    improvements = []
    
    # Create lookup for baseline benchmarks
    baseline_lookup = {b['name']: b for b in baseline_benchmarks}
    
    for current_bench in current_benchmarks:
        name = current_bench['name']
        
        if name not in baseline_lookup:
            if verbose:
                print(f"New benchmark found: {name}")
            continue
            
        baseline_bench = baseline_lookup[name]
        
        # Compare ns/op (lower is better)
        current_ns = current_bench.get('ns_per_op', 0)
        baseline_ns = baseline_bench.get('ns_per_op', 0)
        
        if baseline_ns == 0:
            continue
            
        # Calculate percentage change
        change_percent = ((current_ns - baseline_ns) / baseline_ns) * 100
        
        comparison = {
            'benchmark': name,
            'current_ns_per_op': current_ns,
            'baseline_ns_per_op': baseline_ns,
            'change_percent': round(change_percent, 2),
            'change_type': 'regression' if change_percent > 0 else 'improvement'
        }
        
        comparisons.append(comparison)
        
        # Check for regressions (performance got worse)
        if change_percent > threshold:
            regressions.append({
                'benchmark': name,
                'change': round(change_percent, 2)
            })
        elif change_percent < -threshold:  # Significant improvement
            improvements.append({
                'benchmark': name,
                'change': round(abs(change_percent), 2)
            })
    
    return comparisons, regressions, improvements

def analyze_performance_metrics(current_metrics, baseline_metrics):
    """Analyze performance metrics"""
    metric_comparisons = []
    
    # Create lookup for baseline metrics
    baseline_lookup = {m['metric']: m for m in baseline_metrics}
    
    for current_metric in current_metrics:
        metric_name = current_metric['metric']
        
        if metric_name not in baseline_lookup:
            continue
            
        baseline_metric = baseline_lookup[metric_name]
        
        # Try to compare numeric values
        try:
            current_value = float(current_metric['value'])
            baseline_value = float(baseline_metric['value'])
            
            if baseline_value != 0:
                change_percent = ((current_value - baseline_value) / baseline_value) * 100
                
                metric_comparisons.append({
                    'metric': metric_name,
                    'current_value': current_value,
                    'baseline_value': baseline_value,
                    'change_percent': round(change_percent, 2),
                    'unit': current_metric.get('unit', '')
                })
        except ValueError:
            # Non-numeric values, skip comparison
            continue
    
    return metric_comparisons

# Load performance data
print("Loading performance data...")
current_data = load_performance_data(current_file)
baseline_data = load_performance_data(baseline_file)

if not current_data or not baseline_data:
    print("Failed to load performance data")
    sys.exit(1)

# Extract benchmark data
current_benchmarks = current_data.get('benchmarks', [])
baseline_benchmarks = baseline_data.get('benchmarks', [])

current_metrics = current_data.get('performance_metrics', [])
baseline_metrics = baseline_data.get('performance_metrics', [])

print(f"Comparing {len(current_benchmarks)} current benchmarks against {len(baseline_benchmarks)} baseline benchmarks")

# Compare benchmarks
comparisons, regressions, improvements = compare_benchmarks(current_benchmarks, baseline_benchmarks)

# Analyze performance metrics
metric_comparisons = analyze_performance_metrics(current_metrics, baseline_metrics)

# Calculate summary statistics
worst_regression = max([r['change'] for r in regressions], default=0)
best_improvement = max([i['change'] for i in improvements], default=0)

regression_detected = len(regressions) > 0

# Create comparison report
report = {
    'metadata': {
        'generated_at': datetime.utcnow().isoformat() + 'Z',
        'current_file': current_file,
        'baseline_file': baseline_file,
        'threshold': threshold
    },
    'summary': {
        'total_comparisons': len(comparisons),
        'regressions_found': len(regressions),
        'improvements_found': len(improvements),
        'regression_detected': regression_detected,
        'worst_regression': worst_regression,
        'best_improvement': best_improvement
    },
    'regressions': regressions,
    'improvements': improvements,
    'detailed_comparisons': comparisons,
    'metric_comparisons': metric_comparisons
}

# Save comparison report
with open(output_file, 'w') as f:
    json.dump(report, f, indent=2)

print(f"Comparison report saved to: {output_file}")

# Print summary
print("\n📊 Performance Comparison Summary:")
print(f"  Total comparisons: {len(comparisons)}")
print(f"  Regressions found: {len(regressions)}")
print(f"  Improvements found: {len(improvements)}")
print(f"  Regression threshold: {threshold}%")

if regression_detected:
    print(f"\n❌ Performance regressions detected!")
    print(f"  Worst regression: {worst_regression}%")
    print("  Regressions:")
    for reg in regressions[:5]:  # Show top 5
        print(f"    - {reg['benchmark']}: {reg['change']}% slower")
    
    # Create regression marker file
    with open('./performance-results/regression-detected', 'w') as f:
        f.write(f"Regression detected: {worst_regression}% > {threshold}%\n")
else:
    print(f"\n✅ No performance regressions detected")
    if best_improvement > 0:
        print(f"  Best improvement: {best_improvement}%")

# Verbose output
if verbose:
    print("\n🔍 Detailed Comparisons:")
    for comp in comparisons[:10]:  # Show top 10
        print(f"  {comp['benchmark']}: {comp['change_percent']:+.2f}%")

EOF

    print_status "SUCCESS" "Performance comparison completed"
}

# Generate comparison summary
generate_summary() {
    print_status "INFO" "Generating comparison summary..."

    # Extract key information from comparison report
    if [ -f "$OUTPUT_FILE" ]; then
        python3 << 'EOF'
import json
import os

output_file = os.environ.get('OUTPUT_FILE')

with open(output_file, 'r') as f:
    report = json.load(f)

summary = report['summary']
regressions = report['regressions']
improvements = report['improvements']

print("\n" + "="*50)
print("📊 PERFORMANCE COMPARISON SUMMARY")
print("="*50)
print(f"Total Comparisons: {summary['total_comparisons']}")
print(f"Regressions Found: {summary['regressions_found']}")
print(f"Improvements Found: {summary['improvements_found']}")
print(f"Threshold: {report['metadata']['threshold']}%")

if summary['regression_detected']:
    print(f"\n❌ REGRESSION DETECTED!")
    print(f"Worst Regression: {summary['worst_regression']}%")
    print("\nTop Regressions:")
    for i, reg in enumerate(regressions[:3], 1):
        print(f"  {i}. {reg['benchmark']}: {reg['change']}% slower")
else:
    print(f"\n✅ NO REGRESSIONS DETECTED")
    if summary['best_improvement'] > 0:
        print(f"Best Improvement: {summary['best_improvement']}%")
        print("\nTop Improvements:")
        for i, imp in enumerate(improvements[:3], 1):
            print(f"  {i}. {imp['benchmark']}: {imp['change']}% faster")

print("="*50)
EOF
    fi
}

# Main function
main() {
    parse_args "$@"
    validate_args
    
    print_status "INFO" "Starting performance comparison analysis"
    
    compare_performance
    generate_summary
    
    # Check if regression was detected
    if [ -f "./performance-results/regression-detected" ]; then
        print_status "ERROR" "Performance regression detected above threshold!"
        print_status "INFO" "Review the comparison report: $OUTPUT_FILE"
        exit 1
    else
        print_status "SUCCESS" "Performance comparison completed successfully! 📊"
        print_status "INFO" "Comparison report: $OUTPUT_FILE"
        exit 0
    fi
}

# Export environment variables for Python scripts
export CURRENT_FILE
export BASELINE_FILE
export THRESHOLD
export OUTPUT_FILE
export VERBOSE

# Run main function
main "$@"
