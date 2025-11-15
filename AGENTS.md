# OllamaMax: Agent Development Guide

## 🚀 Project Overview

**OllamaMax** is an enterprise-grade distributed AI model platform that transforms single-node Ollama architecture into a horizontally scalable, fault-tolerant distributed system. This guide helps agents understand how to work effectively in this complex codebase.

## 🏗️ Architecture & Technologies

### Core Technologies
- **Backend**: Go 1.24.5+ (primary), Node.js 18+ (API gateway)
- **Frontend**: JavaScript/TypeScript, React
- **Database**: PostgreSQL (primary), SQLite (development), Redis (caching)
- **Infrastructure**: Docker, Kubernetes, Docker Compose
- **Monitoring**: Prometheus, Grafana, Jaeger, ELK Stack
- **Testing**: Jest, Playwright, Go testing framework

### Key Architecture Components

```
🌐 Load Balancer (NGINX)
    ↓
🎛️  API Gateway (Node.js:13100)
    ↓
🔗 Consensus Layer (Raft-based)
    ↓
🌐 P2P Network (libp2p)
    ↓
🤖 Model Distribution & Serving
    ↓
💾 Storage (PostgreSQL + Redis)
```

### Directory Structure
```
.
├── cmd/                    # Go command-line applications
├── pkg/                    # Go packages (core business logic)
│   ├── api/               # API handlers & middleware
│   ├── auth/              # Authentication & JWT
│   ├── consensus/         # Raft consensus implementation
│   ├── distributed/       # Distributed systems logic
│   ├── p2p/               # Peer-to-peer networking
│   ├── scheduler/         # Intelligent scheduling
│   └── security/          # Security components
├── src/                    # Node.js source code
│   ├── server.js          # Main API server
│   ├── routes/            # API route handlers
│   ├── services/          # Business logic services
│   └── middleware/        # Express middleware
├── api-server/            # Legacy API server
├── web-interface/         # Frontend web interface
├── tests/                 # Comprehensive test suites
├── scripts/               # Build, deployment, utility scripts
├── .bmad-core/            # Agent workflows & templates
├── .claude/               # Claude Flow integration
├── docker/                # Docker configurations
├── k8s/                   # Kubernetes manifests
├── monitoring/            # Monitoring stack configs
└── docs/                  # Documentation
```

## 🔧 Build, Test & Deploy Commands

### Development Setup
```bash
# Install dependencies
npm ci
go mod download

# Set up environment
cp .env.example .env
# Edit .env with proper credentials

# Quick start
npm run dev
```

### Build Commands
```bash
# Build Go binary
make build-app
# OR
go build -o bin/ollamamax ./cmd/ollama-distributed

# Build Node.js application
npm run build

# Build Docker image
docker build -t ollamamax:latest .
```

### Test Commands
```bash
# Unit tests (Go)
go test -v ./pkg/... ./internal/... ./cmd/...

# Unit tests (JavaScript)
npm run test

# Integration tests
npm run test:integration

# E2E tests
npm run test:e2e

# Performance tests
npm run test:performance

# Security tests
npm run test:security

# Comprehensive test suite
npm run test:all

# Coverage validation (90% threshold)
npm run test:coverage
```

### CI/CD Pipeline
```bash
# Local CI validation
npm run validate:ci

# Docker build & test
npm run build-and-test-docker

# Deployment validation
bash scripts/run-deployment-validation.sh
```

## 📋 Code Patterns & Conventions

### Go Code Patterns
- **Error Handling**: Use wrapped errors with `fmt.Errorf("context: %w", err)`
- **Logging**: Structured logging with `slog`
- **Configuration**: Centralized config with validation
- **Testing**: Table-driven tests, 90% coverage minimum
- **Documentation**: Comprehensive godoc comments

### JavaScript/Node.js Patterns
- **Error Handling**: Async/await with try-catch
- **Authentication**: JWT middleware pattern
- **API Design**: RESTful with OpenAPI specs
- **Testing**: Jest with 90% coverage minimum
- **Security**: Input validation, rate limiting, CORS

### Security Patterns
- **JWT Authentication**: RS256 with 2048-bit keys
- **Password Hashing**: bcrypt with 12 rounds
- **Rate Limiting**: Per-endpoint rate limiting
- **Input Validation**: Comprehensive validation middleware
- **Security Headers**: Helmet.js implementation

## 🧪 Testing Approach

### Test Pyramid
1. **Unit Tests** (90% coverage required)
   - Go: `pkg/*/test.go`
   - JavaScript: `*.test.js`

2. **Integration Tests**
   - API integration tests
   - Database integration tests
   - Service-to-service tests

3. **E2E Tests**
   - Playwright browser tests
   - Full user workflow validation
   - Cross-browser testing

4. **Performance Tests**
   - Load testing with k6
   - Stress testing
   - Benchmarking

### Test Organization
```
tests/
├── e2e/                   # Playwright E2E tests
├── integration/           # Integration test suites
├── ml/                    # Machine learning tests
├── deployment/            # Deployment validation
├── performance/           # Performance benchmarks
└── security/              # Security validation
```

### Key Testing Commands
```bash
# Run specific test types
npm run test:auth        # Authentication tests
npm run test:api         # API health tests
npm run test:load        # Load tests
npm run test:performance # Performance tests

# Coverage reporting
npm run test:coverage:report
npm run test:coverage:ci
```

## ⚙️ Configuration Management

### Environment Configuration
```bash
# Development
export NODE_ENV=development
export LOG_LEVEL=debug

# Staging  
export NODE_ENV=staging
export LOG_LEVEL=info

# Production
export NODE_ENV=production
export LOG_LEVEL=warn
export SSL_ENABLED=true
export TLS_ENABLED=true
```

### Key Configuration Files
- `.env` - Environment variables
- `config/*.yaml` - Application configuration
- `docker-compose.yml` - Docker services
- `k8s/*.yaml` - Kubernetes manifests
- `monitoring/*.yaml` - Monitoring configuration

### Configuration Validation
```bash
# Validate configuration
ollama-distributed validate --fix

# Check environment
npm run validate:config
```

## 🚀 Deployment & Operations

### Docker Deployment
```bash
# Build and run
docker-compose up -d

# Production deployment
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# Health checks
curl http://localhost:13100/health
```

### Kubernetes Deployment
```bash
# Deploy to Kubernetes
kubectl apply -f k8s/

# Helm deployment
helm install ollamamax ollamamax/ollamamax-cluster

# Monitoring deployment
kubectl apply -f monitoring/
```

### Monitoring Stack
```bash
# Start monitoring
docker-compose up -d prometheus grafana alertmanager

# Access dashboards
echo "Grafana: http://localhost:3001"
echo "Prometheus: http://localhost:9090"
```

## 🎯 Agent-Specific Patterns

### Claude Flow Integration
```bash
# Agent commands
npm run agents:spawn
npm run agents:orchestrate
npm run agents:status

# Smart agent integration
node src/agents/claude-flow-integration.js spawn
```

### BMAD Agent Workflows
- Located in `.bmad-core/`, `.bmad-creative-writing/`, `.bmad-infrastructure-devops/`
- YAML-based workflow definitions
- Template-driven documentation generation
- Checklist-based quality gates

### Agent Teams
```
web-bundles/teams/
├── team-all.txt           # All agents
├── team-fullstack.txt     # Fullstack team
├── team-ide-minimal.txt   # Minimal IDE team
└── team-no-ui.txt         # Backend-only team
```

## ⚠️ Gotchas & Non-Obvious Patterns

### Critical Security Requirements
1. **JWT Secrets**: Must be 32+ characters, never use defaults
2. **Database Passwords**: Required in production, 8+ characters minimum
3. **CORS Configuration**: Restrict origins in production
4. **SSL/TLS**: Required for production deployments
5. **Environment Variables**: Never commit secrets to version control

### Common Pitfalls
1. **Port Conflicts**: Uses non-standard ports (PostgreSQL: 15432, Redis: 16379)
2. **Coverage Thresholds**: 90% minimum for both Go and JavaScript
3. **Docker Networking**: Complex service dependencies
4. **Database Migrations**: Must run before starting services
5. **Memory Requirements**: High memory usage for AI models

### Performance Considerations
1. **Model Loading**: Large memory footprint for AI models
2. **Database Connections**: Connection pooling required
3. **Redis Caching**: Critical for performance
4. **Load Balancing**: Round-robin with health checks
5. **Monitoring Overhead**: Tracing and metrics impact

### Development Workflow
1. **Branch Strategy**: `main` (production), `develop` (staging), feature branches
2. **Testing**: All tests must pass before merge
3. **Code Review**: Required for all changes
4. **Documentation**: Auto-generated from code comments
5. **CI/CD**: Automated testing and deployment

## 🛠️ Tooling & Utilities

### Development Tools
```bash
# Linting
npm run lint
go vet ./...

# Formatting
gofmt -s -w .
npm run format

# Dependency management
go mod tidy
npm audit
```

### Monitoring Tools
```bash
# Performance monitoring
npm run test:performance:report

# Security scanning
npm run deploy:scan:vulnerabilities

# Coverage reporting
npm run coverage-report
```

### Debugging Tools
```bash
# Health checks
curl http://localhost:13100/health/live
curl http://localhost:13100/health/ready

# Metrics
curl http://localhost:13100/metrics

# Logs
docker-compose logs -f
```

## 📚 Key Files & Entry Points

### Main Entry Points
- `main.go` - Go backend entry point
- `src/server.js` - Node.js API server
- `cmd/ollama-distributed/main.go` - CLI application
- `web-interface/index.html` - Web interface

### Critical Configuration
- `.env.example` - Environment template
- `package.json` - Node.js configuration
- `go.mod` - Go module definition
- `docker-compose.yml` - Service orchestration
- `k8s/deployment.yaml` - Kubernetes deployment

### Documentation Sources
- `README.md` - Project overview
- `docs/` - Detailed documentation
- `claudedocs/` - Generated system analysis
- `COMPREHENSIVE_SYSTEM_REVIEW.md` - System assessment

This guide provides agents with the essential knowledge to navigate, develop, test, and deploy the OllamaMax platform effectively.