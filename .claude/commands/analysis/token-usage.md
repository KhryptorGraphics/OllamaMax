# token-usage

Analyze token usage patterns and optimize for efficiency.

## MCP Tool Usage
```javascript
// Analyze token usage
mcp__claude-flow__token_usage({
  "operation": "analysis",
  "timeframe": "24h"  // Options: 1h, 24h, 7d, 30d
})
```

## Analysis Options
- `operation` - Type of analysis ("analysis", "session", "agent")
- `timeframe` - Analysis period ("1h", "24h", "7d", "30d")

## Examples

### Last 24 hours token usage
```javascript
mcp__claude-flow__token_usage({
  "operation": "analysis",
  "timeframe": "24h"
})

// Returns:
{
  "totalTokens": 45632,
  "byOperation": {
    "search": 12450,
    "analysis": 18320,
    "generation": 14862
  },
  "efficiency": "312 tokens/operation"
}
```

### By agent breakdown
```javascript
mcp__claude-flow__token_usage({
  "operation": "agent",
  "timeframe": "24h"
})

// Returns agent-specific metrics
```

### Session analysis
```javascript
mcp__claude-flow__token_usage({
  "operation": "session",
  "timeframe": "7d"
})

// Returns session-based token metrics
```
