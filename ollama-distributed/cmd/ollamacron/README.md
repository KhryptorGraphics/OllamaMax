# Ollamacron - Distributed Ollama Platform

Ollamacron is the main executable for the distributed Ollama platform. It provides a comprehensive command-line interface for managing distributed Ollama nodes, coordinators, and standalone instances.

## Features

- **Multiple Operation Modes**: Node, Coordinator, and Standalone
- **Comprehensive CLI**: Rich command-line interface with subcommands
- **Production-Ready**: Built for enterprise deployment
- **Security**: Built-in encryption, authentication, and audit logging
- **Monitoring**: Integrated metrics and health checking
- **Configuration**: Flexible YAML-based configuration system
- **Service Integration**: Systemd service files and Docker support

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Coordinator   │    │      Node       │    │   Standalone    │
│                 │    │                 │    │                 │
│ • Consensus     │    │ • P2P Network   │    │ • Local Only    │
│ • Coordination  │    │ • Model Sync    │    │ • No Clustering │
│ • Scheduling    │    │ • Load Balance  │    │ • Simple Deploy │
│ • Monitoring    │    │ • Fault Tolerant│    │ • Quick Start   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Installation

### Building from Source

```bash
# Clone the repository
git clone https://github.com/ollama/ollama-distributed.git
cd ollama-distributed/cmd/ollamacron

# Build the binary
make build

# Or build with specific version
make build VERSION=1.0.0
```

### Cross-Platform Builds

```bash
# Build for all supported platforms
make cross-compile VERSION=1.0.0

# Build for specific platform
GOOS=linux GOARCH=amd64 make build
```

### System Installation

```bash
# Install binary and systemd service
make install

# Or use the installation script
../../deploy/install.sh --type node --enable --start
```

## Usage

### Command Structure

```
ollamacron [global-options] <command> [command-options]
```

### Global Options

- `--config`: Configuration file path
- `--log-level`: Log level (debug, info, warn, error)
- `--log-format`: Log format (json, console)
- `--debug`: Enable debug mode

### Commands

#### Node Mode

Start as a distributed node that can join a cluster:

```bash
ollamacron node \
  --listen "0.0.0.0:11434" \
  --p2p-listen "/ip4/0.0.0.0/tcp/4001" \
  --bootstrap "coordinator:4001" \
  --data-dir "./data" \
  --node-name "worker-1"
```

#### Coordinator Mode

Start as a cluster coordinator:

```bash
ollamacron coordinator \
  --listen "0.0.0.0:11434" \
  --p2p-listen "/ip4/0.0.0.0/tcp/4001" \
  --consensus-listen "0.0.0.0:7000" \
  --bootstrap true \
  --data-dir "./data"
```

#### Standalone Mode

Start in standalone mode (no clustering):

```bash
ollamacron standalone \
  --listen "0.0.0.0:11434" \
  --data-dir "./data" \
  --model-dir "./models"
```

#### Configuration Management

```bash
# Generate default configuration
ollamacron config generate config.yaml

# Validate configuration
ollamacron config validate --config config.yaml
```

#### Status and Monitoring

```bash
# Show node status
ollamacron status

# Check system health
ollamacron health

# Show metrics
ollamacron metrics
```

#### Version Information

```bash
ollamacron version
```

## Configuration

Ollamacron uses YAML configuration files. The configuration can be placed in:

- `./config.yaml`
- `./config/config.yaml`
- `$HOME/.ollamacron/config.yaml`
- `/etc/ollamacron/config.yaml`

### Example Configuration

```yaml
# Node configuration
node:
  id: "node-1"
  name: "ollama-node"
  region: "us-west-2"
  environment: "production"

# API server configuration
api:
  listen: "0.0.0.0:11434"
  timeout: "30s"
  
# P2P networking
p2p:
  listen: "/ip4/0.0.0.0/tcp/4001"
  bootstrap: ["coordinator:4001"]
  enable_dht: true

# Storage configuration
storage:
  data_dir: "./data"
  model_dir: "./models"
  cache_dir: "./cache"

# Security configuration
security:
  auth:
    enabled: true
    method: "jwt"
  encryption:
    algorithm: "AES-256-GCM"
```

## Environment Variables

All configuration options can be overridden using environment variables:

```bash
export OLLAMACRON_LOG_LEVEL=debug
export OLLAMACRON_API_LISTEN=0.0.0.0:11434
export OLLAMACRON_P2P_LISTEN=/ip4/0.0.0.0/tcp/4001
export OLLAMACRON_STORAGE_DATA_DIR=/var/lib/ollamacron
```

## Service Management

### Systemd Service

```bash
# Enable and start service
sudo systemctl enable ollamacron
sudo systemctl start ollamacron

# Check status
sudo systemctl status ollamacron

# View logs
sudo journalctl -u ollamacron -f
```

### Docker Deployment

```bash
# Build Docker image
make docker-build VERSION=1.0.0

# Run single container
docker run -d \
  --name ollamacron \
  -p 11434:11434 \
  -p 4001:4001 \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  ollamacron:1.0.0

# Run with docker-compose
cd ../../deploy/docker
docker-compose up -d
```

## Development

### Building

```bash
# Install dependencies
make deps

# Run checks
make check

# Build for development
make dev-build

# Run development version
make dev-run
```

### Testing

```bash
# Run tests
make test

# Run benchmarks
make bench

# Security scan
make security

# Vulnerability check
make vuln
```

### Make Targets

- `build`: Build the binary
- `clean`: Clean build artifacts
- `test`: Run tests
- `check`: Run all checks
- `install`: Install binary and service
- `docker`: Build and run Docker container
- `release`: Create release packages
- `cross-compile`: Build for all platforms

## Monitoring

Ollamacron provides comprehensive monitoring through:

### Metrics

- **HTTP Metrics**: Request count, duration, status codes
- **P2P Metrics**: Peer count, message rates, bandwidth
- **Consensus Metrics**: Leader election, log replication
- **Storage Metrics**: Disk usage, model cache statistics
- **System Metrics**: CPU, memory, network usage

### Health Checks

- **API Health**: `/api/health` endpoint
- **P2P Health**: Peer connectivity status
- **Consensus Health**: Cluster consensus state
- **Storage Health**: Disk space and model availability

### Logging

- **Structured Logging**: JSON format with structured fields
- **Log Levels**: Debug, Info, Warn, Error
- **Audit Logging**: Security and administrative events
- **Rotation**: Automatic log file rotation

## Security

### Authentication

- **JWT**: JSON Web Token authentication
- **OAuth**: OAuth 2.0 integration
- **x509**: Certificate-based authentication

### Encryption

- **TLS**: Transport Layer Security for all communications
- **AES-256-GCM**: Advanced Encryption Standard
- **Key Management**: Secure key generation and rotation

### Network Security

- **Firewall**: Built-in IP filtering and rules
- **Rate Limiting**: Request rate limiting
- **CORS**: Cross-Origin Resource Sharing controls

## Deployment Scenarios

### Single Node Deployment

```bash
# Start standalone instance
ollamacron standalone --config config.yaml
```

### Multi-Node Cluster

```bash
# Start coordinator
ollamacron coordinator --bootstrap true

# Start nodes
ollamacron node --bootstrap coordinator:4001
ollamacron node --bootstrap coordinator:4001
```

### Production Deployment

```bash
# Install system service
sudo ./deploy/install.sh --type coordinator --enable --start

# Scale cluster
sudo ./deploy/install.sh --type node --enable --start
```

## Troubleshooting

### Common Issues

1. **Port Conflicts**: Ensure ports 11434, 4001, 7000, 8080, 9090 are available
2. **Firewall**: Open required ports in firewall
3. **Permissions**: Ensure proper file permissions for data directory
4. **Bootstrap**: Verify coordinator is running before starting nodes

### Debug Mode

```bash
# Enable debug logging
ollamacron node --debug --log-level debug

# View detailed logs
journalctl -u ollamacron -f
```

### Health Checks

```bash
# Check API health
curl http://localhost:11434/api/health

# Check metrics
curl http://localhost:9090/metrics

# Check node status
ollamacron status
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make check`
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

- **Documentation**: [https://docs.ollama.ai/distributed](https://docs.ollama.ai/distributed)
- **Issues**: [https://github.com/ollama/ollama-distributed/issues](https://github.com/ollama/ollama-distributed/issues)
- **Discord**: [https://discord.gg/ollama](https://discord.gg/ollama)
- **Email**: [support@ollama.ai](mailto:support@ollama.ai)