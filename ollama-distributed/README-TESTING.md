# OllamaMax Distributed Testing Guide

This guide provides comprehensive testing options for the OllamaMax Distributed system.

## Quick Start

### Option 1: Docker Compose Cluster (Recommended)

The easiest way to test the distributed system is using Docker Compose:

```bash
# Run the automated test suite
./scripts/test-cluster.sh

# Or start manually
docker-compose -f docker-compose.test.yml up --build -d

# Check cluster status
./scripts/test-api.sh http://localhost
```

### Option 2: Manual Node Testing

For testing on separate machines or VMs:

```bash
# On Node 1 (Bootstrap node)
./bin/ollama-distributed \
  --node-id=node-1 \
  --cluster-id=test-cluster \
  --bootstrap=true \
  --api-listen=0.0.0.0:8080 \
  --p2p-listen=/ip4/0.0.0.0/tcp/9090 \
  --consensus-bind=0.0.0.0:7070

# On Node 2 (Follower)
./bin/ollama-distributed \
  --node-id=node-2 \
  --cluster-id=test-cluster \
  --bootstrap-peers=<NODE1_IP>:7070 \
  --api-listen=0.0.0.0:8080 \
  --p2p-listen=/ip4/0.0.0.0/tcp/9090 \
  --p2p-bootstrap-peers=/ip4/<NODE1_IP>/tcp/9090 \
  --consensus-bind=0.0.0.0:7070

# On Node 3 (Follower)
./bin/ollama-distributed \
  --node-id=node-3 \
  --cluster-id=test-cluster \
  --bootstrap-peers=<NODE1_IP>:7070,<NODE2_IP>:7070 \
  --api-listen=0.0.0.0:8080 \
  --p2p-listen=/ip4/0.0.0.0/tcp/9090 \
  --p2p-bootstrap-peers=/ip4/<NODE1_IP>/tcp/9090,/ip4/<NODE2_IP>/tcp/9090 \
  --consensus-bind=0.0.0.0:7070
```

## Testing Components

### 1. Cluster Health Testing

```bash
# Test all nodes
for port in 8080 8081 8082; do
  curl -s http://localhost:$port/api/v1/health
done

# Test through load balancer
curl -s http://localhost/api/v1/health
```

### 2. API Endpoint Testing

```bash
# Run comprehensive API tests
./scripts/test-api.sh http://localhost:8080 --verbose

# Test specific endpoints
curl -s http://localhost:8080/api/v1/version | jq
curl -s http://localhost:8080/api/v1/cluster/status | jq
curl -s http://localhost:8080/api/v1/nodes | jq
curl -s http://localhost:8080/api/v1/models | jq
```

### 3. P2P Network Testing

```bash
# Check P2P connectivity (from inside containers)
docker exec ollama-node1 netstat -tlnp | grep 9090
docker exec ollama-node2 netstat -tlnp | grep 9090
docker exec ollama-node3 netstat -tlnp | grep 9090
```

### 4. Consensus Testing

```bash
# Check Raft consensus status
curl -s http://localhost:8080/api/v1/cluster/leader
curl -s http://localhost:8081/api/v1/cluster/leader
curl -s http://localhost:8082/api/v1/cluster/leader

# All should return the same leader
```

### 5. Load Balancing Testing

```bash
# Test load distribution
for i in {1..10}; do
  curl -s http://localhost/api/v1/health -w "Response time: %{time_total}s\n"
done
```

## Test Scenarios

### Scenario 1: Node Failure Recovery

```bash
# Stop one node
docker stop ollama-node2

# Test cluster still works
./scripts/test-api.sh http://localhost

# Restart node
docker start ollama-node2

# Verify it rejoins
sleep 30
./scripts/test-api.sh http://localhost:8081
```

### Scenario 2: Network Partition

```bash
# Simulate network partition
docker network disconnect ollama-cluster_ollama-cluster ollama-node3

# Test cluster behavior
./scripts/test-api.sh http://localhost

# Reconnect
docker network connect ollama-cluster_ollama-cluster ollama-node3
```

### Scenario 3: Load Testing

```bash
# Install Apache Bench (if not available)
# sudo apt-get install apache2-utils

# Run load test
ab -n 1000 -c 10 http://localhost/api/v1/health

# Or use the built-in load test
VERBOSE=true ./scripts/test-api.sh http://localhost
```

## Monitoring and Debugging

### View Logs

```bash
# All services
docker-compose -f docker-compose.test.yml logs -f

# Specific service
docker-compose -f docker-compose.test.yml logs -f ollama-node1

# Last 100 lines
docker-compose -f docker-compose.test.yml logs --tail=100
```

### Check Resource Usage

```bash
# Container stats
docker stats

# Detailed container info
docker inspect ollama-node1
```

### Debug Network Issues

```bash
# Check container networking
docker network ls
docker network inspect ollama-cluster_ollama-cluster

# Test connectivity between containers
docker exec ollama-node1 ping ollama-node2
docker exec ollama-node1 nc -zv ollama-node2 8080
```

## Cleanup

```bash
# Stop and remove all containers
docker-compose -f docker-compose.test.yml down -v

# Remove images (optional)
docker-compose -f docker-compose.test.yml down --rmi all -v

# Clean up everything
./scripts/test-cluster.sh cleanup
```

## Troubleshooting

### Common Issues

1. **Port conflicts**: Ensure ports 8080-8082, 9090-9092, 7070-7072 are available
2. **Docker memory**: Ensure Docker has sufficient memory allocated (4GB+ recommended)
3. **Consensus timeouts**: Normal in test environments without proper cluster setup
4. **P2P discovery**: May take time in containerized environments

### Debug Commands

```bash
# Check if binary works
./bin/ollama-distributed version

# Test configuration
./bin/ollama-distributed --help

# Check Docker build
docker build -t ollama-distributed .

# Manual container run
docker run -it --rm -p 8080:8080 ollama-distributed
```

## Performance Benchmarks

Expected performance in test environment:
- Health check response: < 10ms
- API response time: < 100ms
- Cluster formation: < 30s
- Node recovery: < 60s

## Next Steps

After successful testing:
1. Deploy to production environment
2. Configure proper TLS certificates
3. Set up monitoring and alerting
4. Configure backup and disaster recovery
5. Implement proper authentication and authorization
