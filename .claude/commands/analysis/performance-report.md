# performance-report

Generate comprehensive performance reports for swarm operations.

## MCP Tool Usage
```javascript
// Generate performance report
mcp__claude-flow__performance_report({
  "format": "summary",  // Options: "summary", "detailed", "json"
  "timeframe": "24h"    // Options: "24h", "7d", "30d"
})
```

## Report Options
- `format` - Report format ("summary", "detailed", "json")
- `timeframe` - Analysis period ("24h", "7d", "30d")

## Examples

### Generate summary report
```javascript
mcp__claude-flow__performance_report({
  "format": "summary",
  "timeframe": "24h"
})

// Returns:
{
  "avgExecutionTime": "2.3s",
  "totalTasks": 145,
  "successRate": "94%",
  "bottlenecks": [
    {"type": "coordination", "impact": "15%"}
  ]
}
```

### Detailed metrics report
```javascript
mcp__claude-flow__performance_report({
  "format": "detailed",
  "timeframe": "7d"
})

// Returns comprehensive metrics with trends
```

### JSON export format
```javascript
mcp__claude-flow__performance_report({
  "format": "json",
  "timeframe": "30d"
})

// Returns full JSON data for external processing
```
