# Ollamacron Deployment System

A comprehensive deployment system for the Ollamacron distributed AI inference platform.

## Overview

Ollamacron is a distributed system that enables AI model inference across multiple nodes using P2P networking. This deployment system provides automated installation, configuration, and orchestration capabilities for various environments.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Ollamacron Deployment                       │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐ │
│  │ Installation│  │ Configuration│  │ Orchestration│  │Monitoring│ │
│  │   Scripts   │  │ Management   │  │  (K8s/Docker)│  │ & Alerts│ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

## Quick Start

### Single Node Deployment
```bash
# Linux/macOS
curl -fsSL https://raw.githubusercontent.com/ollama-distributed/main/deploy/install/install.sh | sh

# Windows (PowerShell)
irm https://raw.githubusercontent.com/ollama-distributed/main/deploy/install/install.ps1 | iex
```

### Multi-Node Cluster
```bash
# Using Docker Compose
docker-compose -f deploy/docker/compose/cluster.yml up -d

# Using Kubernetes
helm install ollamacron deploy/kubernetes/helm/ollamacron
```

## Directory Structure

```
deploy/
├── install/                 # Installation scripts
│   ├── linux/              # Linux-specific installers
│   ├── macos/              # macOS-specific installers
│   └── windows/            # Windows-specific installers
├── docker/                 # Docker deployment files
│   ├── multi-stage/        # Multi-stage Dockerfiles
│   └── compose/            # Docker Compose configurations
├── kubernetes/             # Kubernetes deployment files
│   ├── helm/               # Helm charts
│   └── manifests/          # Raw Kubernetes manifests
├── monitoring/             # Monitoring and observability
│   ├── prometheus/         # Prometheus configuration
│   └── grafana/            # Grafana dashboards
├── config/                 # Configuration management
│   ├── defaults/           # Default configurations
│   └── environments/       # Environment-specific configs
└── scripts/               # Utility scripts
```

## Features

### ✅ Installation & Setup
- **Multi-platform support**: Linux, macOS, Windows
- **Automated dependency management**: Go, Docker, Kubernetes tools
- **Service configuration**: Systemd, launchd, Windows services
- **Auto-update mechanisms**: Self-updating deployment system

### ✅ Configuration Management
- **Environment-specific configs**: Development, staging, production
- **Configuration validation**: Schema validation and health checks
- **Runtime configuration updates**: Hot-reload capabilities
- **Secrets management**: Encrypted configuration storage

### ✅ Container & Orchestration
- **Docker support**: Multi-stage builds, optimized images
- **Docker Compose**: Multi-node development environments
- **Kubernetes**: Production-ready Helm charts
- **Service mesh**: Istio integration for advanced networking

### ✅ Monitoring & Observability
- **Metrics collection**: Prometheus integration
- **Visualization**: Grafana dashboards
- **Distributed tracing**: Jaeger integration
- **Log aggregation**: ELK stack integration

### ✅ High Availability
- **Load balancing**: HAProxy, nginx configurations
- **Health checks**: Comprehensive health endpoints
- **Auto-scaling**: Horizontal Pod Autoscaler (HPA)
- **Backup & restore**: Automated backup strategies

## Deployment Modes

### 1. Development Mode
- Single node deployment
- Local development environment
- Hot-reload capabilities
- Debug logging enabled

### 2. Production Mode
- Multi-node cluster
- High availability setup
- Performance optimizations
- Security hardening

### 3. Cloud-Native Mode
- Kubernetes-native deployment
- Service mesh integration
- Cloud provider integrations
- Automated scaling

## Security Features

- **TLS/mTLS**: End-to-end encryption
- **RBAC**: Role-based access control
- **Network policies**: Kubernetes network segmentation
- **Secrets management**: Kubernetes secrets, HashiCorp Vault
- **Security scanning**: Container and dependency scanning

## Performance Optimization

- **Resource management**: CPU/memory limits and requests
- **Caching**: Redis integration for distributed caching
- **Database optimization**: PostgreSQL clustering
- **Network optimization**: Custom CNI configurations

## Monitoring & Alerting

- **System metrics**: CPU, memory, disk, network
- **Application metrics**: Request latency, throughput, errors
- **Business metrics**: Model inference stats, P2P network health
- **Alerting**: PagerDuty, Slack integrations

## Supported Platforms

| Platform | Architecture | Status |
|----------|--------------|--------|
| Linux    | x86_64       | ✅     |
| Linux    | ARM64        | ✅     |
| macOS    | x86_64       | ✅     |
| macOS    | ARM64 (M1)   | ✅     |
| Windows  | x86_64       | ✅     |
| Windows  | ARM64        | 🚧     |

## Getting Started

1. **Choose your deployment method**:
   - Single node: Use installation scripts
   - Multi-node: Use Docker Compose or Kubernetes
   - Cloud: Use Helm charts with cloud provider integration

2. **Configure your environment**:
   - Copy and customize configuration files
   - Set up secrets and credentials
   - Configure monitoring and alerting

3. **Deploy**:
   - Run installation scripts or deployment commands
   - Verify deployment health
   - Configure monitoring dashboards

4. **Scale and manage**:
   - Monitor performance metrics
   - Scale based on demand
   - Update configurations as needed

## Support

- **Documentation**: [docs/](../docs/)
- **Issues**: [GitHub Issues](https://github.com/ollama-distributed/issues)
- **Community**: [Discord](https://discord.gg/ollama-distributed)
- **Commercial Support**: [enterprise@ollama-distributed.com](mailto:enterprise@ollama-distributed.com)

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines on contributing to the deployment system.

## License

Apache License 2.0 - see [LICENSE](../LICENSE) for details.