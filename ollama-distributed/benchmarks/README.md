# Comprehensive Benchmarking Framework

This directory contains a comprehensive benchmarking framework for measuring and optimizing the performance of the ollama-distributed system.

## 🎯 Overview

The benchmarking framework provides:

- **Performance Baseline Establishment**: Measure current system performance across different cluster configurations
- **Comprehensive Benchmark Suite**: Test all major system components (consensus, P2P, API, memory, etc.)
- **Real-time Performance Monitoring**: Live dashboard with alerts and historical data
- **Regression Detection**: Automated detection of performance regressions
- **Detailed Reporting**: Generate comprehensive performance reports with actionable recommendations

## 📊 Benchmark Categories

### 1. Consensus Operations
- Single-threaded consensus performance
- Multi-threaded consensus scalability  
- High-concurrency consensus behavior
- Data size scaling impact

### 2. P2P Networking
- Peer discovery performance
- Message broadcasting efficiency
- Content routing optimization
- NAT traversal performance

### 3. Model Distribution
- Model download speeds
- Model replication efficiency
- Concurrent model access
- Cache performance

### 4. API Endpoints
- Health check response times
- Cluster status queries
- Node operation performance
- Model operation throughput

### 5. Memory Usage
- Memory allocation patterns
- Garbage collection efficiency
- Memory leak detection
- Cache optimization

### 6. Concurrent Operations
- Mixed workload performance
- Resource contention analysis
- Stress testing limits
- Deadlock detection

### 7. Fault Tolerance
- Node failure recovery time
- Network partition handling
- Leader election performance
- Data consistency verification

### 8. Load Balancing
- Load distribution efficiency
- Response time consistency
- Scalability factors
- Resource utilization balance

## 🚀 Quick Start

### Run Complete Benchmark Suite

```bash
# Run all benchmarks with comprehensive reporting
./benchmarks/run_benchmarks.sh

# Run specific benchmark category
go test -bench=BenchmarkConsensusOperations ./benchmarks/...

# Run with custom duration
go test -bench=BenchmarkComprehensivePerformance -benchtime=10m ./benchmarks/...
```

### Start Performance Dashboard

```bash
# Start real-time monitoring dashboard
go run ./benchmarks/performance_monitor.go

# Access dashboard at http://localhost:8080
```

### Run Regression Tests

```bash
# Check for performance regressions
go test -run=TestPerformanceRegression ./benchmarks/...
```

## 📈 Performance Targets

Based on our optimization goals:

### Throughput Targets
- **Single Node**: 100+ requests/second
- **3-Node Cluster**: 250+ requests/second (2.5x scaling)
- **5-Node Cluster**: 400+ requests/second (4x scaling)

### Latency Targets
- **P50 Latency**: ≤ 50ms
- **P95 Latency**: ≤ 200ms  
- **P99 Latency**: ≤ 500ms

### Resource Usage Targets
- **CPU Usage**: ≤ 80% under normal load
- **Memory Usage**: ≤ 2GB per node
- **Error Rate**: ≤ 1%

## 📋 Benchmark Configuration

### Configuration Files

- `benchmark_config.yaml` - Main benchmark configuration
- `performance_config.yaml` - Performance optimization settings

### Key Configuration Options

```yaml
execution:
  duration: "5m"              # Benchmark duration
  warmup_duration: "30s"      # Warmup phase
  concurrent_workers: 8       # Concurrent benchmark workers

system:
  cluster_sizes: [1, 3, 5, 7] # Cluster configurations to test
  model_sizes: [100, 500, 1000] # Model sizes in MB
  concurrency_levels: [1, 10, 50, 100] # Concurrency levels

categories:
  consensus: true
  p2p_networking: true
  model_distribution: true
  # ... other categories
```

## 📊 Results and Reports

### Output Locations

- `benchmarks/results/` - Raw benchmark results
- `benchmarks/reports/` - Generated reports and analysis
- Archives created as `benchmark_results_TIMESTAMP.tar.gz`

### Report Types

1. **Performance Summary**: Overview of all benchmark results
2. **Scalability Analysis**: Analysis of scaling behavior across cluster sizes
3. **Resource Utilization**: CPU, memory, network usage patterns
4. **Regression Analysis**: Comparison with baseline performance
5. **Recommendations**: Actionable optimization suggestions

### Sample Report Structure

```
# Performance Benchmark Report

## System Information
- OS: Linux x86_64
- CPU: 8 cores
- Memory: 16GB
- Go Version: 1.21

## Key Findings
- Throughput: 285 req/s (3-node cluster)
- P95 Latency: 180ms
- Memory Efficiency: 94%
- Success Rate: 99.2%

## Performance Analysis
### Consensus Operations
- Single-threaded: 1,200 ops/s
- Multi-threaded: 4,800 ops/s
- High-concurrency: 12,000 ops/s

### Scalability Metrics
- Linear scaling: 85%
- Efficiency ratio: 2.4

## Recommendations
1. Optimize memory allocation patterns
2. Implement connection pooling
3. Tune consensus batch sizes
```

## 🔍 Real-time Monitoring

### Performance Dashboard Features

- **Live Metrics**: Real-time throughput, latency, resource usage
- **Historical Charts**: Performance trends over time
- **Alert System**: Configurable performance alerts
- **WebSocket Updates**: Real-time data streaming

### Dashboard Metrics

#### System Metrics
- CPU Usage (%)
- Memory Usage (MB & %)
- Active Goroutines
- Network I/O (MB/s)

#### Performance Metrics  
- Requests/Second
- Operations/Second
- Average Latency (ms)
- Error Rate (%)

#### Alert Rules
- High CPU (>80%)
- High Memory (>85%)
- High Latency (>1s P95)
- High Error Rate (>5%)
- Low Throughput (<10 req/s)

## 🔧 Advanced Usage

### Custom Benchmarks

```go
func BenchmarkCustomOperation(b *testing.B) {
    logger := &TestLogger{t: b}
    config := DefaultBenchmarkConfig()
    
    runner := NewBenchmarkRunner(config, logger)
    ctx := context.Background()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        err := runner.benchmarkCustomOperation(ctx)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### Custom Alert Rules

```go
monitor.AddAlertRule(AlertRule{
    Name:      "custom_metric_high",
    Condition: AlertConditionGreaterThan,
    Threshold: 100.0,
    Duration:  time.Minute,
    Severity:  AlertSeverityWarning,
    Callback:  func(alert Alert) {
        // Custom alert handling
        fmt.Printf("Custom alert fired: %v\n", alert)
    },
})
```

### Programmatic Usage

```go
// Create and configure benchmark runner
config := &BenchmarkConfig{
    Duration:          2 * time.Minute,
    ClusterSizes:      []int{1, 3},
    Categories:        []string{"consensus", "api_endpoints"},
    OutputDir:         "./my-results",
}

runner := NewBenchmarkRunner(config, logger)

// Run benchmarks
ctx := context.Background()
err := runner.Run(ctx)
if err != nil {
    log.Fatal("Benchmark failed:", err)
}

// Access results
results := runner.GetResults()
fmt.Printf("Total tests: %d\n", results.Summary.TotalTests)
fmt.Printf("Success rate: %.1f%%\n", 
    float64(results.Summary.SuccessfulTests)/float64(results.Summary.TotalTests)*100)
```

## 📚 Integration

### CI/CD Integration

```yaml
# GitHub Actions example
- name: Run Performance Benchmarks
  run: |
    ./benchmarks/run_benchmarks.sh
    
- name: Check for Regressions
  run: |
    go test -run=TestPerformanceRegression ./benchmarks/...
```

### Monitoring Integration

The framework integrates with:
- **Prometheus**: Export metrics for long-term storage
- **Grafana**: Create custom dashboards
- **AlertManager**: Advanced alerting rules
- **CI/CD Pipelines**: Automated performance validation

## 🎯 Expected Performance Improvements

Based on our topology optimizations and benchmarking framework:

### Throughput Improvements
- **3x improvement** in overall system throughput
- **2.5x scaling efficiency** in 3-node clusters
- **4x scaling efficiency** in 5-node clusters

### Latency Reductions
- **35% reduction** in P95 latency
- **50% reduction** in P99 latency
- **Improved consistency** in response times

### Resource Efficiency
- **30% reduction** in memory usage per node
- **25% improvement** in CPU efficiency
- **40% better** network utilization

### Reliability Improvements
- **99.5% uptime** target achievement
- **<5 second** failure recovery times
- **Zero data loss** during node failures

## 🔧 Troubleshooting

### Common Issues

1. **High Memory Usage During Benchmarks**
   - Reduce concurrent workers in config
   - Increase warmup duration
   - Check for memory leaks

2. **Inconsistent Results**
   - Ensure stable system load
   - Run benchmarks multiple times
   - Check for background processes

3. **Dashboard Not Accessible**
   - Verify port 8080 is available
   - Check firewall settings
   - Ensure monitor is started

### Debug Mode

```bash
# Run with debug logging
go test -bench=BenchmarkAll -v ./benchmarks/...

# Enable detailed metrics
export BENCHMARK_DEBUG=1
./benchmarks/run_benchmarks.sh
```

## 🤝 Contributing

When adding new benchmarks:

1. Follow the existing naming conventions
2. Include proper error handling
3. Add documentation and examples
4. Update configuration files as needed
5. Test across different cluster sizes

## 📄 License

This benchmarking framework is part of the ollama-distributed project and follows the same license terms.