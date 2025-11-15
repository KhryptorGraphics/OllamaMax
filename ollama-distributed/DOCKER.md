# OllamaMax Docker Guide

This guide covers building and running OllamaMax using Docker with support for both production and development environments.

## Quick Start

### Production Build
```bash
# Build multi-architecture production image
npm run docker:build

# Build and push to registry
npm run build:docker:push

# Run with docker-compose
npm run docker:up
```

### Development Build
```bash
# Build development image with hot reload
npm run build:docker:dev

# Run development environment
npm run dev:docker

# Run development with rebuild
npm run dev:docker:build
```

## Docker Images

### Production Image Features
- **Multi-stage build** with Go and Node.js stages
- **Multi-architecture support** (AMD64, ARM64)
- **Minimal runtime** based on Alpine Linux
- **Security hardened** with non-root user
- **Health checks** for all exposed ports
- **Optimized size** with layer caching

### Development Image Features
- **Hot reload** for both Go and Node.js
- **Development tools** (Air, Delve debugger, linting)
- **Volume mounting** for live code changes
- **Debug ports** exposed for IDE integration
- **Comprehensive logging** with debug mode

## Build Commands

### Using Build Script
```bash
# Production build (default)
./scripts/build.sh

# Development build
./scripts/build.sh --dev

# Multi-architecture build
./scripts/build.sh --platforms linux/amd64,linux/arm64

# Build specific version
./scripts/build.sh --version v1.0.0

# Build and push
./scripts/build.sh --push

# Build with tests
./scripts/build.sh --test

# Dry run (show commands)
./scripts/build.sh --dry-run
```

### Using NPM Scripts
```bash
# Production builds
npm run docker:build              # Single architecture
npm run build:docker:multi        # Multi-architecture
npm run build:docker:push         # Build and push

# Development builds
npm run build:docker:dev          # Development image
npm run dev:docker                # Run dev environment
npm run dev:docker:build          # Build and run dev

# Testing
npm run test:docker               # Test built image
```

### Direct Docker Commands
```bash
# Production build
docker build -t ollamamax:latest .

# Development build
docker build -f Dockerfile.dev -t ollamamax:dev .

# Multi-architecture build
docker buildx build --platform linux/amd64,linux/arm64 -t ollamamax:latest .
```

## Running Containers

### Single Node
```bash
# Production
docker run -p 11434:11434 ollamamax:latest

# Development with volume mount
docker run -it --rm \
  -v $(pwd):/app \
  -p 11434:11434 \
  -p 2345:2345 \
  ollamamax:dev
```

### Multi-Node Cluster

#### Production
```bash
# Start full cluster
docker-compose up -d

# Scale nodes
docker-compose up -d --scale ollamamax-node2=2

# View logs
docker-compose logs -f ollamamax-node1
```

#### Development
```bash
# Start development cluster
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up

# Start with rebuild
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up --build

# Start specific services
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up ollamamax-node1 postgres
```

## Port Mapping

### Production Ports
- **11434, 11435, 11436**: API servers (HTTP)
- **12000, 12001, 12002**: P2P networking
- **13000, 13001, 13002**: Raft consensus
- **14090, 14091, 14092**: Metrics

### Development Ports
- **2345, 2346, 2347**: Debug ports (Delve)
- **3000, 3001, 3002**: Web UI development servers

## Environment Variables

### Production
```bash
NODE_ID=node-1                    # Unique node identifier
BOOTSTRAP=true                    # Bootstrap node flag
API_LISTEN=0.0.0.0:11434         # API server address
P2P_LISTEN=/ip4/0.0.0.0/tcp/12000 # P2P listen address
RAFT_BIND_ADDR=0.0.0.0:13102     # Raft bind address (updated to avoid conflicts)
LOG_LEVEL=info                    # Logging level
```

### Development
```bash
LOG_LEVEL=debug                   # Debug logging
GIN_MODE=debug                    # Gin debug mode
NODE_ENV=development              # Node environment
HOT_RELOAD=true                   # Enable hot reload
DEBUG_PORT=2345                   # Debugger port
```

## Volume Mounts

### Production Volumes
```yaml
volumes:
  - ollamamax-data:/app/data      # Application data
  - ollamamax-models:/app/models  # Model storage
  - ollamamax-logs:/app/logs      # Log files
```

### Development Volumes
```yaml
volumes:
  - .:/app:cached                 # Source code (cached)
  - go-mod-cache:/go/pkg/mod      # Go module cache
  - node-modules-cache:/app/web/node_modules  # Node modules cache
```

## Health Checks

### Production Health Check
```bash
# Container health
docker inspect ollamamax-node1 --format='{{.State.Health.Status}}'

# Manual check
curl -f http://localhost:11434/api/v1/health
```

### Development Health Check
```bash
# Quick health check
curl -f http://localhost:11434/api/v1/health

# Debug endpoint
curl http://localhost:11434/debug/vars
```

## Debugging

### Development Debugging
```bash
# Connect with Delve
dlv connect localhost:2345

# VS Code launch.json
{
  "name": "Connect to server",
  "type": "go",
  "request": "attach",
  "mode": "remote",
  "remotePath": "/app",
  "port": 2345,
  "host": "127.0.0.1"
}
```

### Log Analysis
```bash
# Follow logs
docker-compose logs -f ollamamax-node1

# Search logs
docker-compose logs ollamamax-node1 | grep ERROR

# Export logs
docker-compose logs --no-color ollamamax-node1 > node1.log
```

## Performance Optimization

### Build Optimization
- **Multi-stage builds** minimize final image size
- **Layer caching** with proper COPY order
- **Build arg caching** for dependencies
- **Static binary compilation** reduces dependencies

### Runtime Optimization
- **Alpine base image** for minimal footprint
- **Non-root user** for security
- **Health checks** for container orchestration
- **Resource limits** in docker-compose

## Image Size Comparison
```bash
# Check image sizes
docker images | grep ollamamax

# Expected sizes:
# ollamamax:latest    ~50-80MB  (production)
# ollamamax:dev       ~500MB    (development)
```

## Security Considerations

### Production Security
- Non-root user (`ollama:1000`)
- Minimal base image (Alpine)
- No shell access in production
- Secrets via environment variables
- Network policies via docker-compose

### Development Security
- Isolated network
- Volume mount restrictions
- Debug ports only on localhost
- Separate database credentials

## Troubleshooting

### Common Issues

1. **Build fails with "no space left"**
   ```bash
   # Clean Docker system
   docker system prune -a --volumes
   ```

2. **Multi-arch build fails**
   ```bash
   # Setup buildx
   docker buildx create --name multiarch --use
   docker buildx inspect --bootstrap
   ```

3. **Container exits immediately**
   ```bash
   # Check logs
   docker logs <container_id>
   
   # Run interactively
   docker run -it --entrypoint /bin/sh ollamamax:latest
   ```

4. **Hot reload not working**
   ```bash
   # Ensure volume mount is correct
   docker run -v $(pwd):/app ollamamax:dev
   
   # Check file permissions
   ls -la $(pwd)
   ```

5. **Health check fails**
   ```bash
   # Check if service is running
   docker exec <container_id> curl localhost:11434/api/v1/health
   
   # Check port binding
   docker port <container_id>
   ```

### Performance Issues

1. **Slow builds**
   - Use build cache: `--cache-from`
   - Enable BuildKit: `DOCKER_BUILDKIT=1`
   - Use .dockerignore

2. **High memory usage**
   - Set resource limits in docker-compose
   - Monitor with `docker stats`
   - Use multi-stage builds

## Best Practices

### Development Workflow
1. Use development docker-compose for local testing
2. Mount source code for hot reload
3. Use separate databases for development
4. Enable debug logging and ports
5. Test with production image before deployment

### Production Deployment
1. Use multi-architecture builds for ARM support
2. Implement proper health checks
3. Use secrets management
4. Set resource limits
5. Monitor with metrics endpoints

### CI/CD Integration
```yaml
# GitHub Actions example
- name: Build and push
  run: |
    ./scripts/build.sh --version ${{ github.ref_name }} --push
```

## Additional Resources

- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Multi-architecture Builds](https://docs.docker.com/buildx/working-with-buildx/)
- [OllamaMax Architecture](./docs/architecture.md)