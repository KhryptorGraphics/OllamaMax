# Docker Deployment Guide - Enhanced Ollama Distributed Inference

This guide covers deploying the enhanced Ollama distributed inference system with the comprehensive node management dashboard using Docker.

## 🚀 Quick Start

### Development Deployment
```bash
# Quick development setup (no GPU required)
./deploy-docker.sh --dev

# Access the web interface at: http://localhost:13100
```

### Production Deployment
```bash
# Full production stack with GPU support
./deploy-docker.sh --prod --pull-models

# Access the enhanced dashboard at: http://localhost:13100
```

## 📋 Prerequisites

### Required
- Docker 20.10+
- Docker Compose 2.0+
- 8GB+ RAM
- 20GB+ disk space

### For Production (GPU Support)
- NVIDIA Docker Runtime
- NVIDIA GPU with 8GB+ VRAM
- CUDA 12.1+ compatible drivers

## 🏗️ Architecture Overview

The Docker deployment creates a distributed system with:

### Core Services
- **Distributed API Server** (`distributed-api`): Enhanced Node.js API with web interface
- **Ollama Primary** (`ollama-primary`): Primary inference node
- **Ollama Worker 2** (`ollama-worker-2`): Secondary inference node  
- **Ollama Worker 3** (`ollama-worker-3`): Tertiary inference node
- **Redis** (`redis`): Distributed state management
- **MinIO** (`minio`): Distributed model storage

### Monitoring Stack (Production)
- **Prometheus** (`prometheus`): Metrics collection
- **Grafana** (`grafana`): Visualization dashboard

## 🔧 Deployment Options

### 1. Development Mode
**File**: `docker-compose.dev.yml`

**Features**:
- Lightweight setup
- No GPU requirements
- File system mounts for development
- Connects to external Ollama instances

**Usage**:
```bash
./deploy-docker.sh --dev
```

**Ports**:
- `13100`: Web Interface & API
- `13101`: Redis
- `13190`: MinIO API
- `13191`: MinIO Console

### 2. Production Mode
**File**: `docker-compose.distributed.yml`

**Features**:
- Full distributed stack
- GPU acceleration
- High availability
- Comprehensive monitoring

**Usage**:
```bash
./deploy-docker.sh --prod
```

**Ports**:
- `13100`: Web Interface & API
- `13000`: Ollama Primary
- `13001`: Ollama Worker 2
- `13002`: Ollama Worker 3
- `13001`: Redis
- `13090`: MinIO API
- `13091`: MinIO Console
- `13092`: Prometheus
- `13093`: Grafana

## 🚦 Service Management

### Start Services
```bash
# Development
docker-compose -f docker-compose.dev.yml up -d

# Production
docker-compose -f docker-compose.distributed.yml up -d
```

### Stop Services
```bash
# Development
docker-compose -f docker-compose.dev.yml down

# Production
docker-compose -f docker-compose.distributed.yml down
```

### View Logs
```bash
# All services
docker-compose -f docker-compose.distributed.yml logs -f

# Specific service
docker-compose -f docker-compose.distributed.yml logs -f distributed-api
```

### Rebuild Images
```bash
./deploy-docker.sh --rebuild --dev   # Development
./deploy-docker.sh --rebuild --prod  # Production
```

## 🔍 Health Monitoring

### Service Health Checks
All services include comprehensive health checks:

```bash
# Check service status
docker-compose -f docker-compose.distributed.yml ps

# Check specific service health
docker inspect ollama-primary | jq '.[0].State.Health'
```

### Available Endpoints
- API Health: `http://localhost:13100/api/health`
- Ollama Primary: `http://localhost:13000/api/version`
- Redis: `redis://localhost:13001` (password protected)
- MinIO: `http://localhost:13090/minio/health/live`

## 📊 Enhanced Dashboard Features

### Web Interface (`http://localhost:13100`)
- **Chat Interface**: Multi-model distributed chat
- **Enhanced Nodes Tab**: 
  - Detailed node performance metrics
  - Real-time system monitoring
  - Interactive node controls
  - Configuration management
- **Models Tab**: 
  - Cross-node model management
  - P2P model propagation
  - Download and deployment tools
- **Settings Tab**: Configuration and preferences

### Node Management Features
- **Performance Monitoring**: CPU, memory, GPU usage
- **Health Diagnostics**: Multi-point health checks
- **Lifecycle Controls**: Start, stop, restart nodes
- **Model Operations**: Load, unload, migrate models
- **Configuration Management**: Real-time settings updates

## 🗄️ Data Persistence

### Volumes
- `ollama_primary_models`: Primary node models
- `ollama_worker2_models`: Worker 2 models  
- `ollama_worker3_models`: Worker 3 models
- `redis_distributed_data`: Redis state data
- `minio_data`: MinIO object storage
- `api_logs`: API server logs
- `prometheus_data`: Metrics data
- `grafana_data`: Dashboard configurations

### Backup Commands
```bash
# Backup all volumes
docker run --rm -v ollama_primary_models:/data -v $(pwd):/backup alpine \
    tar czf /backup/models-backup.tar.gz /data

# Restore volumes
docker run --rm -v ollama_primary_models:/data -v $(pwd):/backup alpine \
    tar xzf /backup/models-backup.tar.gz -C /
```

## 🛠️ Configuration

### Environment Variables
Key environment variables for customization:

```bash
# API Server
NODE_ENV=production
PORT=13100
REDIS_HOST=redis
OLLAMA_PRIMARY=http://ollama-primary:11434

# Ollama Nodes
OLLAMA_HOST=0.0.0.0
CUDA_VISIBLE_DEVICES=0  # GPU assignment

# Security
MINIO_ROOT_USER=ollama
MINIO_ROOT_PASSWORD=ollama_minio_pass
GF_SECURITY_ADMIN_PASSWORD=ollama_grafana_pass
```

### Custom Configuration
1. Copy and modify `docker-compose.distributed.yml`
2. Update environment variables as needed
3. Add custom volume mounts
4. Adjust resource limits

## 🔒 Security Considerations

### Production Security
- Change all default passwords
- Use proper SSL certificates
- Configure firewall rules
- Enable authentication on all services
- Regular security updates

### Network Security
- Services communicate on isolated Docker network
- Only necessary ports exposed
- Redis password protected
- MinIO with authentication

## 🐛 Troubleshooting

### Common Issues

**1. GPU Not Detected**
```bash
# Check NVIDIA runtime
docker info | grep nvidia

# Test GPU access
docker run --gpus all nvidia/cuda:12.1-base-ubuntu22.04 nvidia-smi
```

**2. Service Won't Start**
```bash
# Check logs
docker-compose -f docker-compose.distributed.yml logs service-name

# Check resource usage
docker stats
```

**3. Model Loading Fails**
```bash
# Check Ollama service status
curl http://localhost:13000/api/version

# Manually pull model
docker exec ollama-primary ollama pull tinyllama:latest
```

**4. API Connection Issues**
```bash
# Test API endpoints
curl http://localhost:13100/api/health
curl http://localhost:13100/api/nodes/detailed

# Check network connectivity
docker network ls
docker network inspect ollama_distributed
```

### Log Locations
- API Server: `docker-compose logs distributed-api`
- Ollama Nodes: `docker-compose logs ollama-primary`
- Redis: `docker-compose logs redis`
- System: `/var/log/docker`

## 📈 Performance Optimization

### Resource Allocation
```yaml
# In docker-compose.yml
services:
  ollama-primary:
    deploy:
      resources:
        reservations:
          memory: 4G
          devices:
            - driver: nvidia
              count: 1
              capabilities: [gpu]
        limits:
          memory: 8G
```

### Scaling
```bash
# Scale worker nodes
docker-compose -f docker-compose.distributed.yml up -d --scale ollama-worker-2=3

# Load balancing automatically handled by API server
```

## 🔄 Updates and Maintenance

### Update Deployment
```bash
# Pull latest images
docker-compose -f docker-compose.distributed.yml pull

# Restart with new images
docker-compose -f docker-compose.distributed.yml up -d

# Or use deployment script
./deploy-docker.sh --rebuild --prod
```

### Maintenance Tasks
```bash
# Clean unused images
docker image prune -a

# Clean unused volumes
docker volume prune

# Full cleanup
docker system prune -a --volumes
```

## 🆘 Support

For issues and support:
1. Check service logs first
2. Verify system requirements
3. Test individual components
4. Check network connectivity
5. Review configuration files

The enhanced Docker deployment provides a robust, scalable foundation for the distributed Ollama inference system with comprehensive monitoring and management capabilities.