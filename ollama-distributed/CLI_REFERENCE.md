# OllamaMax CLI Reference

Quick reference guide for OllamaMax distributed system command-line interface.

## 🚀 Node Management

### Start Node
```bash
./ollama-distributed start [options]

Options:
  --config string     Configuration file path
  --peers string      Comma-separated list of peer addresses
  --port int          API server port (default: 8080)
  --node-id string    Unique node identifier
```

### Node Status
```bash
./ollama-distributed status

# Shows:
# - Node ID and status
# - Cluster membership
# - Resource usage
# - Network connectivity
```

### Join Cluster
```bash
./ollama-distributed join --peers node1:8080,node2:8080

# Joins existing cluster by connecting to specified peers
```

## 🎛️ Proxy Management

### Proxy Status
```bash
# Basic status
./ollama-distributed proxy status

# JSON output
./ollama-distributed proxy status --json

# Custom API URL
./ollama-distributed proxy status --api-url http://localhost:9999

# Output includes:
# - Proxy status (running/stopped)
# - Instance count
# - Healthy instances
# - Load balancer status
```

### Instance Management
```bash
# List instances
./ollama-distributed proxy instances

# JSON output
./ollama-distributed proxy instances --json

# Custom API URL
./ollama-distributed proxy instances --api-url http://node2:8080

# Output includes:
# - Instance ID and Node ID
# - Endpoint URL
# - Health status
# - Request/error counts
# - Last seen timestamp
```

### Performance Metrics
```bash
# Current metrics
./ollama-distributed proxy metrics

# JSON output
./ollama-distributed proxy metrics --json

# Real-time monitoring
./ollama-distributed proxy metrics --watch

# Custom update interval (seconds)
./ollama-distributed proxy metrics --watch --interval 10

# Custom API URL
./ollama-distributed proxy metrics --api-url http://node3:8080

# Output includes:
# - Total/successful/failed requests
# - Average latency
# - Requests per second
# - Load balancing metrics
# - Per-instance metrics
```

## 🔧 Common Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--json` | Output in JSON format | false |
| `--api-url` | API server URL | http://localhost:8080 |
| `--watch` | Real-time monitoring | false |
| `--interval` | Update interval (seconds) | 5 |
| `--config` | Configuration file path | - |
| `--peers` | Peer addresses | - |

## 📋 Usage Examples

### Basic Cluster Setup
```bash
# Node 1 (bootstrap)
./ollama-distributed start --port 8080

# Node 2 (join cluster)
./ollama-distributed start --port 8081 --peers localhost:8080

# Node 3 (join cluster)
./ollama-distributed start --port 8082 --peers localhost:8080,localhost:8081
```

### Monitoring Workflow
```bash
# Check cluster status
./ollama-distributed status

# Monitor proxy health
./ollama-distributed proxy status

# List all instances
./ollama-distributed proxy instances

# Watch metrics in real-time
./ollama-distributed proxy metrics --watch
```

### Scripting and Automation
```bash
# Health check script
if ./ollama-distributed proxy status --json | jq -e '.status == "running"' > /dev/null; then
  echo "Proxy is healthy"
else
  echo "Proxy is unhealthy"
  exit 1
fi

# Instance count monitoring
HEALTHY_COUNT=$(./ollama-distributed proxy instances --json | jq '[.instances[] | select(.status=="healthy")] | length')
echo "Healthy instances: $HEALTHY_COUNT"

# Metrics export
./ollama-distributed proxy metrics --json > /tmp/metrics-$(date +%s).json
```

### Troubleshooting
```bash
# Check specific node
./ollama-distributed proxy status --api-url http://problematic-node:8080

# Monitor failed requests
./ollama-distributed proxy metrics --json | jq '.failed_requests'

# List unhealthy instances
./ollama-distributed proxy instances --json | jq '.instances[] | select(.status!="healthy")'

# Watch for changes
watch -n 2 './ollama-distributed proxy instances'
```

## 🚨 Error Handling

### Common Error Messages

| Error | Cause | Solution |
|-------|-------|----------|
| `connection refused` | API server not running | Start the node with `./ollama-distributed start` |
| `HTTP 503` | Proxy not initialized | Wait for proxy to start or check logs |
| `timeout` | Network issues | Check network connectivity and firewall |
| `invalid JSON` | Malformed response | Check API server logs for errors |

### Debug Commands
```bash
# Verbose output
./ollama-distributed start --log-level debug

# Check logs
tail -f /var/log/ollama-distributed.log

# Network connectivity test
curl http://localhost:8080/api/v1/proxy/status
```

## 📚 Additional Resources

- [Main Documentation](../README.md)
- [Configuration Guide](CONFIG.md)
- [API Reference](API.md)
- [Troubleshooting Guide](TROUBLESHOOTING.md)

---

**Quick Help**: Run any command with `--help` for detailed usage information.

```bash
./ollama-distributed --help
./ollama-distributed proxy --help
./ollama-distributed proxy status --help
```
