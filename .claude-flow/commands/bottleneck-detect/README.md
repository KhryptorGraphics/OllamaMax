# Bottleneck Detection - Performance Analysis & Optimization

Analyze performance bottlenecks in swarm operations and automatically apply optimizations to improve system efficiency.

## Overview

The Bottleneck Detection system continuously monitors swarm performance across four key areas:
- **Communication** - Message latency, coordination overhead
- **Processing** - Agent utilization, task completion times
- **Memory** - Cache performance, pattern loading
- **Network** - API latency, service timeouts

## Features

- **Real-time Analysis** - Monitor performance across configurable time ranges
- **Intelligent Detection** - Identify bottlenecks with configurable impact thresholds
- **Automatic Optimization** - Apply fixes automatically with `--fix` flag
- **Comprehensive Reporting** - Detailed analysis with actionable recommendations
- **Export Capabilities** - Save analysis results for further review

## Installation

```bash
# Install dependencies
npm install

# Make executable
chmod +x index.js

# Optional: Install globally
npm install -g .
```

## Usage

### Basic Commands

```bash
# Basic bottleneck detection
./index.js

# Analyze specific swarm
./index.js --swarm-id swarm-123

# Last 24 hours with export
./index.js -t 24h -e bottlenecks.json

# Auto-fix detected issues
./index.js --fix --threshold 15
```

### Command Options

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--swarm-id` | `-s` | Analyze specific swarm | current |
| `--time-range` | `-t` | Analysis period (1h, 24h, 7d, all) | 1h |
| `--threshold` | | Bottleneck threshold percentage | 20 |
| `--export` | `-e` | Export analysis to file | none |
| `--fix` | | Apply automatic optimizations | false |
| `--help` | `-h` | Show help message | false |

## Analysis Categories

### Communication Bottlenecks
- **Message Latency** - Delays in inter-agent communication
- **Coordination Overhead** - Resource consumption for coordination
- **Response Times** - Agent response delays
- **Queue Sizes** - Message queue backlogs

### Processing Bottlenecks
- **Agent Utilization** - Underutilized or overloaded agents
- **Task Completion** - Slow task execution times
- **Queue Wait Times** - Tasks waiting for assignment
- **Parallel Efficiency** - Suboptimal parallel execution

### Memory Bottlenecks
- **Cache Hit Rates** - Poor cache performance
- **Pattern Loading** - Slow neural pattern access
- **Memory Access** - High memory latency
- **Storage I/O** - Disk access delays

### Network Bottlenecks
- **API Latency** - Slow external API calls
- **MCP Communication** - Model Context Protocol delays
- **Service Timeouts** - External service issues
- **Request Limits** - Concurrent request bottlenecks

## Automatic Optimizations

### Communication Fixes
- **Message Batching** - Combine multiple messages for efficiency
- **Topology Switching** - Change to more efficient communication patterns
- **Priority Routing** - Prioritize critical messages

### Processing Fixes
- **Concurrency Tuning** - Adjust agent counts and parallel execution
- **Load Balancing** - Better distribute workload across agents
- **Task Prioritization** - Optimize task scheduling

### Memory Fixes
- **Smart Caching** - Implement intelligent caching strategies
- **Memory Pooling** - Optimize memory allocation
- **Pattern Preloading** - Load frequently used patterns in advance

### Network Fixes
- **Connection Pooling** - Reuse connections for efficiency
- **Request Batching** - Combine API requests
- **Timeout Tuning** - Optimize timeout settings

## Output Examples

### Basic Analysis
```
🔍 Bottleneck Analysis Report
━━━━━━━━━━━━━━━━━━━━━━━━━━━

📊 Summary
├── Time Range: Last 1 hour
├── Agents Analyzed: 7
├── Tasks Processed: 115
└── Critical Issues: 0

⚠️ Warning Bottlenecks
1. Message Latency (46% impact)
   └── Message latency averaging 2286ms
2. Low Cache Hit Rate (28% impact)
   └── Cache hit rate only 72%

💡 Recommendations
1. Enable message batching (est. 25% improvement)
2. Enable smart caching with preloading (est. 35% improvement)

📈 Performance Metrics
├── Communication Efficiency: 54%
├── Processing Efficiency: 77%
├── Memory Efficiency: 72%
└── Overall Score: 69%
```

### With Auto-Fix Applied
```
🔧 Optimization Results
━━━━━━━━━━━━━━━━━━━━━━━━
✅ Enable smart message batching
   └── Improvement: 27%
✅ Enable smart caching with LRU strategy
   └── Improvement: 34%

📊 Applied 2/2 optimizations
🚀 Estimated total improvement: 61%
```

## Performance Impact

Typical improvements after bottleneck resolution:

| Category | Improvement Range |
|----------|------------------|
| Communication | 30-50% faster message delivery |
| Processing | 20-40% reduced task completion time |
| Memory | 40-60% fewer cache misses |
| Network | 25-45% reduced API latency |
| Overall | 25-45% performance improvement |

## Export Format

Analysis results can be exported to JSON format:

```json
{
  "timestamp": "2024-01-29T10:30:00.000Z",
  "analysis": {
    "summary": {
      "timeRange": "Last 1 hour",
      "agentsAnalyzed": 7,
      "tasksProcessed": 115,
      "criticalIssues": 0
    },
    "criticalBottlenecks": [],
    "warningBottlenecks": [...],
    "recommendations": [...],
    "metrics": {...}
  }
}
```

## Integration

### Claude Code Integration
```javascript
mcp__claude-flow__bottleneck_detect { 
  timeRange: "1h",
  threshold: 20,
  autoFix: false
}
```

### Programmatic Usage
```javascript
const BottleneckDetectCLI = require('./index');

const cli = new BottleneckDetectCLI();
const result = await cli.run(['--threshold', '15', '--fix']);
```

## Troubleshooting

### Common Issues

1. **No Metrics Available**
   - Ensure swarm is running and generating metrics
   - Check time range - use longer periods for more data

2. **High False Positives**
   - Increase threshold percentage
   - Use longer time ranges for better averages

3. **Optimization Failures**
   - Check system permissions
   - Verify swarm configuration allows modifications

### Debug Mode
```bash
DEBUG=bottleneck-detect ./index.js
```

## Best Practices

1. **Regular Monitoring** - Run analysis periodically to catch issues early
2. **Baseline Establishment** - Record normal performance metrics
3. **Gradual Optimization** - Apply fixes incrementally and monitor results
4. **Export Analysis** - Keep records for trend analysis
5. **Threshold Tuning** - Adjust thresholds based on your performance requirements

## Contributing

1. Fork the repository
2. Create feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit pull request

## License

MIT License - see LICENSE file for details.
