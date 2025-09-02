# OllamaMax Project Overview

## Project Purpose
OllamaMax is an enterprise-grade distributed AI model platform that transforms single-node Ollama architecture into a horizontally scalable, fault-tolerant, high-performance distributed system for serving large language models.

## Tech Stack
- **Backend**: Node.js with Express.js API server (distributed-inference.js)
- **Frontend**: Vanilla HTML/CSS/JavaScript with WebSocket connections
- **Database/State**: Redis for distributed state management and queuing
- **Storage**: MinIO for distributed model storage
- **Monitoring**: Prometheus metrics, Grafana dashboards
- **Container**: Docker with Docker Compose for orchestration
- **P2P**: Built-in peer-to-peer model sharing and migration

## Key Architecture Components
1. **Distributed API Server** (api-server/distributed-inference.js) - Main coordination layer
2. **Web Interface** (web-interface/) - Frontend dashboard for management
3. **Enhanced Node Management** - Advanced monitoring and control
4. **P2P Model Migration** - Automatic model distribution between nodes
5. **Redis Coordination** - Distributed state and messaging
6. **WebSocket Real-time Communication** - Live updates and chat functionality

## Project Structure
- `/api-server/` - Node.js backend services
- `/web-interface/` - Frontend web dashboard
- `/docker-compose*.yml` - Container orchestration configs
- `/monitoring/` - Prometheus/Grafana monitoring stack
- `/docs/` - Comprehensive documentation
- `/scripts/` - Deployment and utility scripts