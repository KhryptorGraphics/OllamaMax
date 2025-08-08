#!/usr/bin/env python3

"""
Performance Charts Generation Script
Generates visual charts for performance comparison and analysis
"""

import json
import argparse
import os
import sys
from datetime import datetime
import matplotlib.pyplot as plt
import matplotlib.dates as mdates
import seaborn as sns
import pandas as pd
import numpy as np

# Set style for better-looking charts
plt.style.use('seaborn-v0_8')
sns.set_palette("husl")

def load_performance_data(filename):
    """Load performance data from JSON file"""
    try:
        with open(filename, 'r') as f:
            return json.load(f)
    except Exception as e:
        print(f"Error loading {filename}: {e}")
        return None

def create_benchmark_comparison_chart(current_data, baseline_data, output_dir):
    """Create benchmark comparison chart"""
    print("📊 Creating benchmark comparison chart...")
    
    current_benchmarks = current_data.get('benchmarks', [])
    baseline_benchmarks = baseline_data.get('benchmarks', [])
    
    # Create lookup for baseline benchmarks
    baseline_lookup = {b['name']: b for b in baseline_benchmarks}
    
    # Prepare data for comparison
    benchmark_names = []
    current_values = []
    baseline_values = []
    improvements = []
    
    for current_bench in current_benchmarks:
        name = current_bench['name']
        if name in baseline_lookup:
            baseline_bench = baseline_lookup[name]
            
            current_ns = current_bench.get('ns_per_op', 0)
            baseline_ns = baseline_bench.get('ns_per_op', 0)
            
            if baseline_ns > 0 and current_ns > 0:
                # Calculate improvement percentage (negative means regression)
                improvement = ((baseline_ns - current_ns) / baseline_ns) * 100
                
                benchmark_names.append(name.replace('Benchmark', ''))
                current_values.append(current_ns)
                baseline_values.append(baseline_ns)
                improvements.append(improvement)
    
    if not benchmark_names:
        print("⚠️ No matching benchmarks found for comparison")
        return
    
    # Create comparison chart
    fig, (ax1, ax2) = plt.subplots(2, 1, figsize=(14, 10))
    
    # Chart 1: Performance comparison (ns/op)
    x = np.arange(len(benchmark_names))
    width = 0.35
    
    bars1 = ax1.bar(x - width/2, baseline_values, width, label='Baseline', alpha=0.8)
    bars2 = ax1.bar(x + width/2, current_values, width, label='Current', alpha=0.8)
    
    ax1.set_xlabel('Benchmarks')
    ax1.set_ylabel('Nanoseconds per Operation')
    ax1.set_title('Performance Comparison: Baseline vs Current')
    ax1.set_xticks(x)
    ax1.set_xticklabels(benchmark_names, rotation=45, ha='right')
    ax1.legend()
    ax1.grid(True, alpha=0.3)
    
    # Add value labels on bars
    for bar in bars1:
        height = bar.get_height()
        ax1.annotate(f'{int(height)}',
                    xy=(bar.get_x() + bar.get_width() / 2, height),
                    xytext=(0, 3),
                    textcoords="offset points",
                    ha='center', va='bottom', fontsize=8)
    
    for bar in bars2:
        height = bar.get_height()
        ax1.annotate(f'{int(height)}',
                    xy=(bar.get_x() + bar.get_width() / 2, height),
                    xytext=(0, 3),
                    textcoords="offset points",
                    ha='center', va='bottom', fontsize=8)
    
    # Chart 2: Improvement percentage
    colors = ['green' if imp > 0 else 'red' for imp in improvements]
    bars3 = ax2.bar(x, improvements, color=colors, alpha=0.7)
    
    ax2.set_xlabel('Benchmarks')
    ax2.set_ylabel('Improvement (%)')
    ax2.set_title('Performance Change: Positive = Improvement, Negative = Regression')
    ax2.set_xticks(x)
    ax2.set_xticklabels(benchmark_names, rotation=45, ha='right')
    ax2.axhline(y=0, color='black', linestyle='-', alpha=0.3)
    ax2.grid(True, alpha=0.3)
    
    # Add percentage labels
    for bar, imp in zip(bars3, improvements):
        height = bar.get_height()
        ax2.annotate(f'{imp:.1f}%',
                    xy=(bar.get_x() + bar.get_width() / 2, height),
                    xytext=(0, 3 if height >= 0 else -15),
                    textcoords="offset points",
                    ha='center', va='bottom' if height >= 0 else 'top',
                    fontsize=8, fontweight='bold')
    
    plt.tight_layout()
    plt.savefig(os.path.join(output_dir, 'benchmark_comparison.png'), dpi=300, bbox_inches='tight')
    plt.close()
    
    print("✅ Benchmark comparison chart saved")

def create_performance_trend_chart(current_data, baseline_data, output_dir):
    """Create performance trend chart"""
    print("📈 Creating performance trend chart...")
    
    # Extract performance statistics
    current_stats = current_data.get('performance_stats', {})
    baseline_stats = baseline_data.get('performance_stats', {})
    
    metrics = ['ns_per_op', 'allocs_per_op']
    stats_types = ['mean', 'median', 'min', 'max']
    
    fig, axes = plt.subplots(2, 2, figsize=(15, 10))
    axes = axes.flatten()
    
    for i, metric in enumerate(metrics):
        current_metric_data = current_stats.get(metric, {})
        baseline_metric_data = baseline_stats.get(metric, {})
        
        if not current_metric_data or not baseline_metric_data:
            continue
        
        # Prepare data
        categories = []
        current_values = []
        baseline_values = []
        
        for stat_type in stats_types:
            if stat_type in current_metric_data and stat_type in baseline_metric_data:
                categories.append(stat_type.title())
                current_values.append(current_metric_data[stat_type])
                baseline_values.append(baseline_metric_data[stat_type])
        
        if categories:
            x = np.arange(len(categories))
            width = 0.35
            
            ax = axes[i]
            bars1 = ax.bar(x - width/2, baseline_values, width, label='Baseline', alpha=0.8)
            bars2 = ax.bar(x + width/2, current_values, width, label='Current', alpha=0.8)
            
            ax.set_xlabel('Statistics')
            ax.set_ylabel(metric.replace('_', ' ').title())
            ax.set_title(f'{metric.replace("_", " ").title()} Statistics Comparison')
            ax.set_xticks(x)
            ax.set_xticklabels(categories)
            ax.legend()
            ax.grid(True, alpha=0.3)
    
    # Memory allocation trend
    if len(axes) > 2:
        ax = axes[2]
        current_benchmarks = current_data.get('benchmarks', [])
        baseline_benchmarks = baseline_data.get('benchmarks', [])
        
        # Extract memory allocation data
        current_allocs = [b.get('allocs_per_op', 0) for b in current_benchmarks if b.get('allocs_per_op', 0) > 0]
        baseline_allocs = [b.get('allocs_per_op', 0) for b in baseline_benchmarks if b.get('allocs_per_op', 0) > 0]
        
        if current_allocs and baseline_allocs:
            ax.hist(baseline_allocs, bins=20, alpha=0.7, label='Baseline', density=True)
            ax.hist(current_allocs, bins=20, alpha=0.7, label='Current', density=True)
            ax.set_xlabel('Allocations per Operation')
            ax.set_ylabel('Density')
            ax.set_title('Memory Allocation Distribution')
            ax.legend()
            ax.grid(True, alpha=0.3)
    
    # Performance score comparison
    if len(axes) > 3:
        ax = axes[3]
        
        # Calculate overall performance scores
        current_score = calculate_performance_score(current_data)
        baseline_score = calculate_performance_score(baseline_data)
        
        scores = [baseline_score, current_score]
        labels = ['Baseline', 'Current']
        colors = ['lightblue', 'lightgreen' if current_score >= baseline_score else 'lightcoral']
        
        bars = ax.bar(labels, scores, color=colors, alpha=0.8)
        ax.set_ylabel('Performance Score')
        ax.set_title('Overall Performance Score')
        ax.grid(True, alpha=0.3)
        
        # Add score labels
        for bar, score in zip(bars, scores):
            height = bar.get_height()
            ax.annotate(f'{score:.2f}',
                       xy=(bar.get_x() + bar.get_width() / 2, height),
                       xytext=(0, 3),
                       textcoords="offset points",
                       ha='center', va='bottom', fontweight='bold')
    
    plt.tight_layout()
    plt.savefig(os.path.join(output_dir, 'performance_trends.png'), dpi=300, bbox_inches='tight')
    plt.close()
    
    print("✅ Performance trend chart saved")

def calculate_performance_score(data):
    """Calculate overall performance score"""
    benchmarks = data.get('benchmarks', [])
    if not benchmarks:
        return 0.0
    
    # Simple scoring based on ns/op (lower is better)
    ns_values = [b.get('ns_per_op', 0) for b in benchmarks if b.get('ns_per_op', 0) > 0]
    if not ns_values:
        return 0.0
    
    # Normalize and invert (higher score = better performance)
    avg_ns = sum(ns_values) / len(ns_values)
    score = max(0, 100 - (avg_ns / 1000))  # Arbitrary scaling
    return min(100, score)

def create_resource_usage_chart(current_data, baseline_data, output_dir):
    """Create resource usage comparison chart"""
    print("💾 Creating resource usage chart...")
    
    # Extract resource usage metrics
    current_metrics = current_data.get('performance_metrics', [])
    baseline_metrics = baseline_data.get('performance_metrics', [])
    
    # Create lookup for metrics
    current_lookup = {m['metric']: m for m in current_metrics}
    baseline_lookup = {m['metric']: m for m in baseline_metrics}
    
    # Resource metrics to compare
    resource_metrics = ['CPU', 'Memory', 'Network']
    
    fig, ax = plt.subplots(figsize=(10, 6))
    
    current_values = []
    baseline_values = []
    metric_names = []
    
    for metric in resource_metrics:
        if metric in current_lookup and metric in baseline_lookup:
            try:
                current_val = float(current_lookup[metric]['value'])
                baseline_val = float(baseline_lookup[metric]['value'])
                
                current_values.append(current_val)
                baseline_values.append(baseline_val)
                metric_names.append(metric)
            except ValueError:
                continue
    
    if metric_names:
        x = np.arange(len(metric_names))
        width = 0.35
        
        bars1 = ax.bar(x - width/2, baseline_values, width, label='Baseline', alpha=0.8)
        bars2 = ax.bar(x + width/2, current_values, width, label='Current', alpha=0.8)
        
        ax.set_xlabel('Resource Metrics')
        ax.set_ylabel('Usage')
        ax.set_title('Resource Usage Comparison')
        ax.set_xticks(x)
        ax.set_xticklabels(metric_names)
        ax.legend()
        ax.grid(True, alpha=0.3)
        
        plt.tight_layout()
        plt.savefig(os.path.join(output_dir, 'resource_usage.png'), dpi=300, bbox_inches='tight')
        plt.close()
        
        print("✅ Resource usage chart saved")
    else:
        print("⚠️ No resource usage metrics found for comparison")

def create_summary_report(current_data, baseline_data, output_dir):
    """Create summary performance report"""
    print("📋 Creating summary report...")
    
    # Calculate key metrics
    current_benchmarks = current_data.get('benchmarks', [])
    baseline_benchmarks = baseline_data.get('benchmarks', [])
    
    total_benchmarks = len(current_benchmarks)
    regressions = 0
    improvements = 0
    
    baseline_lookup = {b['name']: b for b in baseline_benchmarks}
    
    for current_bench in current_benchmarks:
        name = current_bench['name']
        if name in baseline_lookup:
            baseline_bench = baseline_lookup[name]
            
            current_ns = current_bench.get('ns_per_op', 0)
            baseline_ns = baseline_bench.get('ns_per_op', 0)
            
            if baseline_ns > 0 and current_ns > 0:
                change = ((current_ns - baseline_ns) / baseline_ns) * 100
                if change > 5:  # 5% threshold
                    regressions += 1
                elif change < -5:
                    improvements += 1
    
    # Create summary chart
    fig, ((ax1, ax2), (ax3, ax4)) = plt.subplots(2, 2, figsize=(12, 8))
    
    # Pie chart of benchmark results
    labels = ['Regressions', 'Improvements', 'No Change']
    sizes = [regressions, improvements, total_benchmarks - regressions - improvements]
    colors = ['red', 'green', 'gray']
    
    ax1.pie(sizes, labels=labels, colors=colors, autopct='%1.1f%%', startangle=90)
    ax1.set_title('Benchmark Results Distribution')
    
    # Performance score comparison
    current_score = calculate_performance_score(current_data)
    baseline_score = calculate_performance_score(baseline_data)
    
    scores = [baseline_score, current_score]
    labels = ['Baseline', 'Current']
    colors = ['lightblue', 'lightgreen' if current_score >= baseline_score else 'lightcoral']
    
    bars = ax2.bar(labels, scores, color=colors)
    ax2.set_ylabel('Performance Score')
    ax2.set_title('Overall Performance Score')
    ax2.set_ylim(0, 100)
    
    # Add score labels
    for bar, score in zip(bars, scores):
        height = bar.get_height()
        ax2.annotate(f'{score:.1f}',
                    xy=(bar.get_x() + bar.get_width() / 2, height),
                    xytext=(0, 3),
                    textcoords="offset points",
                    ha='center', va='bottom', fontweight='bold')
    
    # Summary statistics
    ax3.text(0.1, 0.9, f'Total Benchmarks: {total_benchmarks}', transform=ax3.transAxes, fontsize=12)
    ax3.text(0.1, 0.8, f'Regressions: {regressions}', transform=ax3.transAxes, fontsize=12, color='red')
    ax3.text(0.1, 0.7, f'Improvements: {improvements}', transform=ax3.transAxes, fontsize=12, color='green')
    ax3.text(0.1, 0.6, f'Performance Score: {current_score:.1f}', transform=ax3.transAxes, fontsize=12)
    ax3.text(0.1, 0.5, f'Score Change: {current_score - baseline_score:+.1f}', transform=ax3.transAxes, fontsize=12)
    ax3.set_xlim(0, 1)
    ax3.set_ylim(0, 1)
    ax3.set_title('Performance Summary')
    ax3.axis('off')
    
    # Timestamp and metadata
    ax4.text(0.1, 0.9, f'Generated: {datetime.now().strftime("%Y-%m-%d %H:%M:%S")}', transform=ax4.transAxes, fontsize=10)
    ax4.text(0.1, 0.8, f'Baseline Benchmarks: {len(baseline_benchmarks)}', transform=ax4.transAxes, fontsize=10)
    ax4.text(0.1, 0.7, f'Current Benchmarks: {len(current_benchmarks)}', transform=ax4.transAxes, fontsize=10)
    ax4.set_xlim(0, 1)
    ax4.set_ylim(0, 1)
    ax4.set_title('Report Metadata')
    ax4.axis('off')
    
    plt.tight_layout()
    plt.savefig(os.path.join(output_dir, 'performance_summary.png'), dpi=300, bbox_inches='tight')
    plt.close()
    
    print("✅ Summary report saved")

def main():
    parser = argparse.ArgumentParser(description='Generate performance comparison charts')
    parser.add_argument('--current', required=True, help='Current performance report JSON file')
    parser.add_argument('--baseline', required=True, help='Baseline performance report JSON file')
    parser.add_argument('--output', required=True, help='Output directory for charts')
    
    args = parser.parse_args()
    
    # Create output directory
    os.makedirs(args.output, exist_ok=True)
    
    # Load performance data
    print("📊 Loading performance data...")
    current_data = load_performance_data(args.current)
    baseline_data = load_performance_data(args.baseline)
    
    if not current_data or not baseline_data:
        print("❌ Failed to load performance data")
        sys.exit(1)
    
    print(f"✅ Loaded current data: {len(current_data.get('benchmarks', []))} benchmarks")
    print(f"✅ Loaded baseline data: {len(baseline_data.get('benchmarks', []))} benchmarks")
    
    # Generate charts
    create_benchmark_comparison_chart(current_data, baseline_data, args.output)
    create_performance_trend_chart(current_data, baseline_data, args.output)
    create_resource_usage_chart(current_data, baseline_data, args.output)
    create_summary_report(current_data, baseline_data, args.output)
    
    print(f"\n🎉 Performance charts generated successfully!")
    print(f"📁 Charts saved to: {args.output}")
    print("📊 Generated charts:")
    print("  - benchmark_comparison.png")
    print("  - performance_trends.png")
    print("  - resource_usage.png")
    print("  - performance_summary.png")

if __name__ == '__main__':
    main()
