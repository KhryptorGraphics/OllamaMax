# OllamaMax API Port Configuration

## Port Assignment Summary

This document outlines the resolved port conflicts and standardized API surface for OllamaMax.

### Service Port Assignments

| Service | Port | Purpose | Configuration File |
|---------|------|---------|-------------------|
| **Node.js API Gateway** | **13100** | Primary API server (OpenAI compatible) | `src/server.js`, `.env` |
| **Node.js WebSocket Server** | **13101** | WebSocket management & node coordination | `api-server/server.js` |
| **Node.js Fallback** | **13102** | Additional fallback endpoint | Frontend config |
| **Go Backend API** | **8080** | Distributed inference API | `ollama-distributed/internal/config/config.go` |
| **Go Backend Web** | **8081** | Web interface (if separate) | `ollama-distributed/cmd/node/quickstart.go` |
| **Ollama Engine** | **11434** | Core Ollama model serving | Docker config |
| **PostgreSQL** | **5432** | User database | Docker config |
| **Redis** | **6379** | Caching & sessions | Docker config |
| **Grafana** | **3001** | Monitoring dashboards | `docker-compose.yml` |
| **Prometheus** | **9090** | Metrics collection | `docker-compose.yml` |

### Environment Configuration

#### Primary API Configuration (`.env`)
```bash
# Server Ports - Primary API Gateway
PORT=13100
API_PORT=13100
GO_API_PORT=11434

# Ollama Configuration - Use Go backend as primary
ENABLE_OLLAMA_DISCOVERY=true
ENABLE_LOCAL_OLLAMA=false
OLLAMA_NODES=http://localhost:11434
```

#### Frontend Configuration
```javascript
// Primary API endpoints
this.endpoints = [
  'ws://localhost:13100/chat',    // Primary Node.js API
  'ws://localhost:13101/chat',    // WebSocket server
  'ws://localhost:13102/chat'     // Fallback
];

// REST API base URL
const API_BASE = 'http://localhost:13100';
```

### Conflict Resolution

#### Previous Conflicts
1. **Port 13000**: Used by both `src/server.js` and `api-server/server.js`
2. **Port 11434**: Used by Go backend and Ollama engine
3. **Port 13000**: Also used by Grafana in custom docker compose

#### Resolved Conflicts
1. ✅ Node.js API moved to **13100**
2. ✅ Node.js WebSocket server moved to **13101**
3. ✅ Go backend API moved to **8080**
4. ✅ Ollama engine remains on **11434**
5. ✅ Grafana moved to **3001**

### API Surface Standardization

#### Primary API Endpoints (Node.js)
- **Base URL**: `http://localhost:13100`
- **WebSocket**: `ws://localhost:13100/chat`
- **Authentication**: `http://localhost:13100/auth`
- **OpenAI API**: `http://localhost:13100/v1`
- **Health**: `http://localhost:13100/health`
- **Metrics**: `http://localhost:13100/metrics`
- **Documentation**: `http://localhost:13100/docs`

#### Secondary API Endpoints (Go Backend)
- **Base URL**: `http://localhost:8080`
- **API**: `http://localhost:8080/api/v1`
- **Health**: `http://localhost:8080/health`
- **Metrics**: `http://localhost:8080/metrics`

### Frontend Integration

#### Connection Strategy
1. **Primary**: Try `ws://localhost:13100/chat` (Node.js API)
2. **Fallback**: Try `ws://localhost:13101/chat` (WebSocket server)
3. **Alternative**: Try `ws://localhost:13102/chat`

#### API Calls
- All REST API calls use `http://localhost:13100`
- WebSocket connections use the endpoint strategy above
- Authentication tokens work across all services

### Docker Configuration

#### Production (`docker-compose.yml`)
```yaml
ollamamax-api:
  ports:
    - "13100:13100"  # API server
    - "3000:3000"    # Web interface

ollama:
  ports:
    - "11434:11434"  # Ollama engine

grafana:
  ports:
    - "3001:3000"    # Monitoring (avoid conflict)
```

#### Development
- Node.js servers use environment variables for port configuration
- Go backend uses port 8080 for API, 8081 for web
- All services can run simultaneously without conflicts

### Testing Configuration

#### Unit Tests
- Updated to use `ws://localhost:13100/chat`
- API base URLs updated to `http://localhost:13100`

#### Integration Tests
- Test multiple endpoint connections
- Verify fallback behavior
- Validate cross-service authentication

### Migration Guide

#### For Developers
1. Update local `.env` file with new port assignments
2. Restart all services
3. Clear browser cache for frontend changes
4. Update any hardcoded URLs in custom scripts

#### For Production
1. Update docker compose files
2. Update load balancer configurations
3. Update monitoring/alerting configurations
4. Update API documentation

### Monitoring & Observability

#### Health Checks
- Node.js API: `http://localhost:13100/health/ready`
- Go Backend: `http://localhost:8080/health/ready`
- Ollama: `http://localhost:11434/api/tags`

#### Metrics Endpoints
- Node.js: `http://localhost:13100/metrics`
- Go Backend: `http://localhost:8080/metrics`
- Prometheus: `http://localhost:9090`

### Security Considerations

#### CORS Configuration
- Node.js API allows origins: `localhost:13100`, `localhost:3000`
- Go backend allows origins: `localhost:13100`, `localhost:8080`
- Update allowed origins based on deployment domain

#### Authentication
- JWT tokens generated by Node.js API work across all services
- API keys work with both Node.js and Go backends
- Session management centralized through Node.js API

### Troubleshooting

#### Common Issues
1. **Port already in use**: Check for conflicting processes
2. **Connection refused**: Verify services are running
3. **Authentication errors**: Check JWT secret synchronization
4. **WebSocket errors**: Verify endpoint configuration

#### Debug Commands
```bash
# Check port usage
lsof -i :13100
lsof -i :13101
lsof -i :8080

# Test API connectivity
curl http://localhost:13100/health
curl http://localhost:8080/health

# Test WebSocket connection
wscat -c ws://localhost:13100/chat
```

### Future Considerations

#### Scaling Options
1. **Horizontal scaling**: Load balance across multiple Node.js instances
2. **Service separation**: Deploy Go backend as separate microservice
3. **Database scaling**: Consider read replicas for high availability

#### Port Evolution
- Consider using environment-based port ranges
- Implement service discovery for dynamic port allocation
- Standardize on cloud-native port conventions

---

**Last Updated**: November 15, 2025
**Version**: 1.0.0
**Status**: ✅ Configuration Complete