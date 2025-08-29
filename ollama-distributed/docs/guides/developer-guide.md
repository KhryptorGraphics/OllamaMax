# Developer Guide - Ollama Distributed

A comprehensive guide for developers working with the Ollama Distributed platform, covering architecture, APIs, SDKs, and integration patterns.

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Development Setup](#development-setup)
3. [API Integration](#api-integration)
4. [SDK Usage](#sdk-usage)
5. [Plugin Development](#plugin-development)
6. [Contributing Guidelines](#contributing-guidelines)
7. [Testing & Debugging](#testing--debugging)

## Architecture Overview

### System Architecture

Ollama Distributed follows a microservices architecture with the following core components:

```
┌─────────────────────────────────────────────────────────────┐
│                    Client Applications                       │
├─────────────────────────────────────────────────────────────┤
│                      Load Balancer                          │
├─────────────────────────────────────────────────────────────┤
│  API Gateway  │  WebSocket  │  Web UI  │  CLI Interface   │
├─────────────────────────────────────────────────────────────┤
│           Distributed Coordination Layer                    │
│  ├─ Consensus Engine (Raft)  ├─ Service Discovery         │
│  ├─ Leader Election          ├─ Configuration Management   │
├─────────────────────────────────────────────────────────────┤
│                    Core Services                            │
│  ├─ Model Manager    ├─ Scheduler       ├─ Auth Service    │
│  ├─ P2P Network     ├─ Storage Engine  ├─ Monitoring      │
├─────────────────────────────────────────────────────────────┤
│                   Node Infrastructure                       │
│  ├─ Inference Engine  ├─ Model Storage  ├─ Metrics        │
│  ├─ Health Checker    ├─ Log Aggregator ├─ Security       │
└─────────────────────────────────────────────────────────────┘
```

### Key Design Principles

- **Scalability**: Linear horizontal scaling to 10,000+ nodes
- **Fault Tolerance**: < 30s recovery time with 99.9% availability
- **Consistency**: Strong consistency for critical operations via Raft consensus
- **Performance**: Sub-100ms inference latency with intelligent load balancing
- **Security**: Zero-trust architecture with comprehensive authentication/authorization

### Component Deep Dive

#### 1. P2P Network Layer (`pkg/p2p/`)
```go
// Core networking interface
type NetworkManager interface {
    Connect(peerID string, address string) error
    Disconnect(peerID string) error
    Broadcast(message Message) error
    SendTo(peerID string, message Message) error
    Subscribe(messageType MessageType, handler MessageHandler)
}

// Message types
type MessageType int
const (
    ModelSyncMessage MessageType = iota
    ConsensusMessage
    HealthCheckMessage
    ConfigUpdateMessage
)
```

#### 2. Consensus Engine (`pkg/consensus/`)
```go
// Raft-based consensus for distributed coordination
type ConsensusEngine interface {
    Start() error
    Stop() error
    IsLeader() bool
    GetLeader() string
    ProposeChange(change StateChange) error
    GetState() ClusterState
}

// State machine for cluster operations
type ClusterState struct {
    Nodes      map[string]NodeInfo
    Models     map[string]ModelInfo
    Config     ClusterConfig
    Timestamp  time.Time
}
```

#### 3. Model Distribution (`pkg/models/`)
```go
// Model management and distribution
type ModelManager interface {
    Download(name string, opts DownloadOptions) error
    Distribute(name string, nodes []string) error
    Remove(name string) error
    GetStatus(name string) ModelStatus
    List() []ModelInfo
}

// P2P model transfer
type ModelTransfer interface {
    Transfer(modelName string, fromNode, toNode string) error
    GetProgress(transferID string) TransferProgress
    Cancel(transferID string) error
}
```

## Development Setup

### Prerequisites

```bash
# Required tools
go version >= 1.19
docker version >= 20.10
docker-compose version >= 2.0
git version >= 2.30

# Optional tools for development
golangci-lint   # Code linting
gotestsum      # Enhanced test output
air            # Live reload during development
delve          # Debugging
```

### Local Development Environment

#### 1. Clone and Setup
```bash
# Clone repository
git clone https://github.com/ollama/ollama-distributed.git
cd ollama-distributed

# Install dependencies
go mod download

# Set up development configuration
cp .env.example .env
cp config/node.yaml.example config/node.yaml

# Initialize development database
make dev-setup
```

#### 2. Build and Run
```bash
# Build all components
make build

# Run development cluster (3 nodes)
make dev-cluster

# Run single node for testing
go run cmd/distributed-ollama/main.go --config config/dev.yaml

# Run with live reload
air -c .air.toml
```

#### 3. Development Docker Environment
```bash
# Start development stack
docker-compose -f docker-compose.dev.yml up -d

# View logs
docker-compose -f docker-compose.dev.yml logs -f

# Rebuild and restart service
docker-compose -f docker-compose.dev.yml up -d --build ollama-node
```

### IDE Configuration

#### VS Code Setup
```json
// .vscode/settings.json
{
    "go.toolsManagement.checkForUpdates": "local",
    "go.useLanguageServer": true,
    "go.lintOnSave": "workspace",
    "go.formatTool": "gofmt",
    "go.testTimeout": "30s",
    "go.buildTags": "dev,integration"
}

// .vscode/launch.json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch Node",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/cmd/distributed-ollama",
            "args": ["--config", "config/dev.yaml", "--debug"],
            "cwd": "${workspaceFolder}"
        }
    ]
}
```

#### GoLand/IntelliJ Setup
```xml
<!-- Enable Go modules and build tags -->
<component name="GoModuleSettings">
  <option name="buildTags" value="dev,integration" />
  <option name="vendoringMode" value="MODULE" />
</component>
```

## API Integration

### REST API Usage

#### Authentication
```go
// API client with authentication
type APIClient struct {
    baseURL string
    token   string
    client  *http.Client
}

func NewAPIClient(baseURL, token string) *APIClient {
    return &APIClient{
        baseURL: baseURL,
        token:   token,
        client:  &http.Client{Timeout: 30 * time.Second},
    }
}

func (c *APIClient) makeRequest(method, endpoint string, body interface{}) (*http.Response, error) {
    var reqBody io.Reader
    if body != nil {
        jsonBody, err := json.Marshal(body)
        if err != nil {
            return nil, err
        }
        reqBody = bytes.NewBuffer(jsonBody)
    }
    
    req, err := http.NewRequest(method, c.baseURL+endpoint, reqBody)
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Authorization", "Bearer "+c.token)
    req.Header.Set("Content-Type", "application/json")
    
    return c.client.Do(req)
}
```

#### Model Management Examples
```go
// Download a model
func (c *APIClient) DownloadModel(modelName string) error {
    resp, err := c.makeRequest("POST", "/api/v1/models/"+modelName+"/download", nil)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("download failed: %s", resp.Status)
    }
    
    return nil
}

// List models
func (c *APIClient) ListModels() ([]ModelInfo, error) {
    resp, err := c.makeRequest("GET", "/api/v1/models", nil)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var response struct {
        Models map[string]ModelInfo `json:"models"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return nil, err
    }
    
    var models []ModelInfo
    for _, model := range response.Models {
        models = append(models, model)
    }
    
    return models, nil
}

// Generate text
func (c *APIClient) GenerateText(model, prompt string, stream bool) (*GenerateResponse, error) {
    request := GenerateRequest{
        Model:  model,
        Prompt: prompt,
        Stream: stream,
    }
    
    resp, err := c.makeRequest("POST", "/api/v1/generate", request)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var response GenerateResponse
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return nil, err
    }
    
    return &response, nil
}
```

### WebSocket Integration

#### Real-time Updates
```go
// WebSocket client for real-time updates
type WSClient struct {
    conn        *websocket.Conn
    subscribers map[string][]chan interface{}
    mutex       sync.RWMutex
}

func NewWSClient(url string) (*WSClient, error) {
    conn, _, err := websocket.DefaultDialer.Dial(url, nil)
    if err != nil {
        return nil, err
    }
    
    client := &WSClient{
        conn:        conn,
        subscribers: make(map[string][]chan interface{}),
    }
    
    go client.listen()
    return client, nil
}

func (c *WSClient) Subscribe(channel string) <-chan interface{} {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    ch := make(chan interface{}, 100)
    c.subscribers[channel] = append(c.subscribers[channel], ch)
    
    // Send subscription message
    msg := map[string]interface{}{
        "type":    "subscribe",
        "channel": channel,
    }
    c.conn.WriteJSON(msg)
    
    return ch
}

func (c *WSClient) listen() {
    for {
        var message map[string]interface{}
        if err := c.conn.ReadJSON(&message); err != nil {
            log.Printf("WebSocket read error: %v", err)
            break
        }
        
        msgType, ok := message["type"].(string)
        if !ok {
            continue
        }
        
        c.mutex.RLock()
        subscribers := c.subscribers[msgType]
        c.mutex.RUnlock()
        
        for _, ch := range subscribers {
            select {
            case ch <- message:
            default:
                // Channel full, skip
            }
        }
    }
}
```

#### Usage Example
```go
// Monitor cluster metrics in real-time
func MonitorMetrics() {
    ws, err := NewWSClient("ws://localhost:8080/api/v1/ws")
    if err != nil {
        log.Fatal(err)
    }
    
    metricsCh := ws.Subscribe("metrics")
    alertsCh := ws.Subscribe("alerts")
    
    for {
        select {
        case metric := <-metricsCh:
            handleMetrics(metric)
        case alert := <-alertsCh:
            handleAlert(alert)
        }
    }
}

func handleMetrics(data interface{}) {
    metrics, ok := data.(map[string]interface{})
    if !ok {
        return
    }
    
    if cpuUsage, ok := metrics["cpu_usage"].(float64); ok {
        if cpuUsage > 80 {
            log.Printf("High CPU usage detected: %.2f%%", cpuUsage)
        }
    }
}
```

## SDK Usage

### Go SDK

#### Installation
```bash
go get github.com/ollama/ollama-distributed/sdk/go
```

#### Basic Usage
```go
import (
    "context"
    "log"
    
    "github.com/ollama/ollama-distributed/sdk/go"
)

func main() {
    // Initialize client
    client := ollama.NewClient(&ollama.Config{
        BaseURL: "http://localhost:8080",
        APIKey:  "your-api-key",
        Timeout: 30 * time.Second,
    })
    
    ctx := context.Background()
    
    // Download model
    if err := client.Models.Download(ctx, "llama2"); err != nil {
        log.Fatal(err)
    }
    
    // Generate text
    response, err := client.Generate(ctx, &ollama.GenerateRequest{
        Model:  "llama2",
        Prompt: "Explain artificial intelligence in simple terms",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Generated text:", response.Response)
    
    // Stream generation
    stream, err := client.GenerateStream(ctx, &ollama.GenerateRequest{
        Model:  "llama2",
        Prompt: "Write a short story",
        Stream: true,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    for chunk := range stream {
        if chunk.Error != nil {
            log.Printf("Stream error: %v", chunk.Error)
            break
        }
        fmt.Print(chunk.Response)
    }
}
```

#### Advanced Features
```go
// Cluster management
cluster := client.Cluster

// Get cluster status
status, err := cluster.Status(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Leader: %s, Nodes: %d\n", status.Leader, len(status.Nodes))

// Node management
nodes := client.Nodes

// List all nodes
nodeList, err := nodes.List(ctx)
if err != nil {
    log.Fatal(err)
}

for id, node := range nodeList {
    fmt.Printf("Node %s: %s (CPU: %.1f%%, Memory: %.1f%%)\n",
        id, node.Status, node.Usage.CPU, node.Usage.Memory)
}

// Drain a node
if err := nodes.Drain(ctx, "node-001"); err != nil {
    log.Fatal(err)
}
```

### Python SDK

#### Installation
```bash
pip install ollama-distributed
```

#### Basic Usage
```python
from ollama_distributed import Client, GenerateRequest
import asyncio

async def main():
    # Initialize client
    client = Client(
        base_url="http://localhost:8080",
        api_key="your-api-key"
    )
    
    # Download model
    await client.models.download("llama2")
    
    # Generate text
    response = await client.generate(GenerateRequest(
        model="llama2",
        prompt="Explain quantum computing"
    ))
    
    print("Generated text:", response.response)
    
    # Stream generation
    async for chunk in client.generate_stream(GenerateRequest(
        model="llama2",
        prompt="Write a poem",
        stream=True
    )):
        print(chunk.response, end="", flush=True)

if __name__ == "__main__":
    asyncio.run(main())
```

#### WebSocket Support
```python
from ollama_distributed import WSClient
import asyncio

async def monitor_cluster():
    async with WSClient("ws://localhost:8080/api/v1/ws") as ws:
        # Subscribe to metrics
        await ws.subscribe("metrics")
        await ws.subscribe("alerts")
        
        async for message in ws.listen():
            if message.type == "metrics":
                handle_metrics(message.data)
            elif message.type == "alerts":
                handle_alert(message.data)

def handle_metrics(metrics):
    cpu_usage = metrics.get("cpu_usage", 0)
    memory_usage = metrics.get("memory_usage", 0)
    
    print(f"CPU: {cpu_usage:.1f}%, Memory: {memory_usage:.1f}%")
    
    if cpu_usage > 80:
        print("⚠️  High CPU usage detected!")

def handle_alert(alert):
    level = alert.get("level", "info")
    message = alert.get("message", "")
    
    icon = {"error": "❌", "warning": "⚠️", "info": "ℹ️"}.get(level, "📢")
    print(f"{icon} {level.upper()}: {message}")

if __name__ == "__main__":
    asyncio.run(monitor_cluster())
```

### JavaScript/Node.js SDK

#### Installation
```bash
npm install @ollama/distributed-client
```

#### Basic Usage
```javascript
const { OllamaClient } = require('@ollama/distributed-client');

async function main() {
    const client = new OllamaClient({
        baseURL: 'http://localhost:8080',
        apiKey: 'your-api-key'
    });
    
    // Download model
    await client.models.download('llama2');
    
    // Generate text
    const response = await client.generate({
        model: 'llama2',
        prompt: 'Explain machine learning'
    });
    
    console.log('Generated text:', response.response);
    
    // Stream generation
    const stream = await client.generateStream({
        model: 'llama2',
        prompt: 'Write a story',
        stream: true
    });
    
    for await (const chunk of stream) {
        process.stdout.write(chunk.response);
    }
}

main().catch(console.error);
```

#### WebSocket Support
```javascript
const { WSClient } = require('@ollama/distributed-client');

async function monitorCluster() {
    const ws = new WSClient('ws://localhost:8080/api/v1/ws');
    
    await ws.connect();
    
    // Subscribe to channels
    await ws.subscribe('metrics');
    await ws.subscribe('alerts');
    
    ws.on('metrics', (data) => {
        console.log(`CPU: ${data.cpu_usage}%, Memory: ${data.memory_usage}%`);
        
        if (data.cpu_usage > 80) {
            console.log('⚠️  High CPU usage detected!');
        }
    });
    
    ws.on('alerts', (alert) => {
        const icons = { error: '❌', warning: '⚠️', info: 'ℹ️' };
        const icon = icons[alert.level] || '📢';
        console.log(`${icon} ${alert.level.toUpperCase()}: ${alert.message}`);
    });
    
    // Keep alive
    process.on('SIGINT', async () => {
        await ws.disconnect();
        process.exit(0);
    });
}

monitorCluster().catch(console.error);
```

## Plugin Development

### Plugin Architecture

Ollama Distributed supports a plugin system for extending functionality:

#### Plugin Interface
```go
// Plugin defines the interface all plugins must implement
type Plugin interface {
    Name() string
    Version() string
    Description() string
    
    // Lifecycle methods
    Initialize(ctx context.Context, config map[string]interface{}) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    
    // Health check
    Health() error
}

// Hooks for different plugin types
type ModelPlugin interface {
    Plugin
    PreProcess(request *InferenceRequest) error
    PostProcess(response *InferenceResponse) error
}

type MonitoringPlugin interface {
    Plugin
    CollectMetrics() (map[string]interface{}, error)
    HandleAlert(alert Alert) error
}

type AuthPlugin interface {
    Plugin
    Authenticate(credentials map[string]string) (*User, error)
    Authorize(user *User, resource string, action string) error
}
```

#### Example Plugin
```go
// example-plugin/main.go
package main

import (
    "context"
    "log"
    
    "github.com/ollama/ollama-distributed/pkg/plugin"
)

type LoggingPlugin struct {
    config map[string]interface{}
}

func (p *LoggingPlugin) Name() string        { return "logging-plugin" }
func (p *LoggingPlugin) Version() string     { return "1.0.0" }
func (p *LoggingPlugin) Description() string { return "Enhanced logging plugin" }

func (p *LoggingPlugin) Initialize(ctx context.Context, config map[string]interface{}) error {
    p.config = config
    return nil
}

func (p *LoggingPlugin) Start(ctx context.Context) error {
    log.Println("Logging plugin started")
    return nil
}

func (p *LoggingPlugin) Stop(ctx context.Context) error {
    log.Println("Logging plugin stopped")
    return nil
}

func (p *LoggingPlugin) Health() error {
    return nil
}

func (p *LoggingPlugin) PreProcess(request *plugin.InferenceRequest) error {
    log.Printf("Processing request for model: %s", request.Model)
    return nil
}

func (p *LoggingPlugin) PostProcess(response *plugin.InferenceResponse) error {
    log.Printf("Response generated in %v", response.Duration)
    return nil
}

// Plugin registration
func main() {
    plugin.RegisterModelPlugin(&LoggingPlugin{})
}
```

#### Plugin Configuration
```yaml
# config/plugins.yaml
plugins:
  - name: logging-plugin
    path: ./plugins/logging-plugin.so
    config:
      log_level: info
      log_format: json
      
  - name: auth-plugin
    path: ./plugins/auth-plugin.so
    config:
      provider: oauth2
      client_id: "your-client-id"
      
  - name: metrics-plugin
    path: ./plugins/metrics-plugin.so
    config:
      interval: 30s
      exporters: [prometheus, statsd]
```

### Building Plugins
```bash
# Build plugin as shared library
go build -buildmode=plugin -o plugins/logging-plugin.so example-plugin/main.go

# Install plugin
./ollama-distributed plugin install ./plugins/logging-plugin.so

# List installed plugins
./ollama-distributed plugin list

# Enable/disable plugins
./ollama-distributed plugin enable logging-plugin
./ollama-distributed plugin disable logging-plugin
```

## Contributing Guidelines

### Code Style & Standards

#### Go Code Standards
```go
// Use proper package naming
package models // Not package model_manager

// Follow Go naming conventions
func GetUserByID(id string) (*User, error) // Not getUserById

// Proper error handling
func ProcessModel(name string) error {
    if name == "" {
        return fmt.Errorf("model name cannot be empty")
    }
    
    model, err := loadModel(name)
    if err != nil {
        return fmt.Errorf("failed to load model %s: %w", name, err)
    }
    
    return processModel(model)
}

// Use context for cancellation
func DownloadModel(ctx context.Context, name string) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // Continue processing
    }
    
    // Implementation
    return nil
}
```

#### Documentation Standards
```go
// Package documentation
// Package models provides model management and distribution functionality
// for the Ollama Distributed system.
//
// Key features:
//   - Model downloading and caching
//   - P2P model distribution
//   - Version management
//   - Health monitoring
package models

// Function documentation with examples
// DownloadModel downloads a model from the registry and stores it locally.
// It returns an error if the download fails or if the model already exists.
//
// Example:
//   err := DownloadModel(ctx, "llama2")
//   if err != nil {
//       log.Fatal(err)
//   }
func DownloadModel(ctx context.Context, name string) error {
    // Implementation
}
```

### Pull Request Process

#### 1. Development Workflow
```bash
# Create feature branch
git checkout -b feature/model-caching

# Make changes and commit
git add .
git commit -m "feat: implement model caching for improved performance

- Add LRU cache for frequently accessed models
- Implement cache eviction policy
- Add metrics for cache hit/miss rates
- Update documentation

Closes #123"

# Push and create PR
git push origin feature/model-caching
gh pr create --title "Implement model caching" --body "Description..."
```

#### 2. PR Requirements
- ✅ **Tests**: All new code must have tests (minimum 80% coverage)
- ✅ **Documentation**: Update relevant documentation
- ✅ **Linting**: Pass all linting and formatting checks
- ✅ **Security**: No security vulnerabilities introduced
- ✅ **Performance**: No significant performance regressions
- ✅ **Backward Compatibility**: Maintain API compatibility

#### 3. Review Process
1. **Automated Checks**: CI/CD pipeline runs all tests and checks
2. **Code Review**: At least 2 maintainer approvals required
3. **Testing**: Manual testing for complex features
4. **Documentation Review**: Technical writing review if needed
5. **Merge**: Squash and merge after all approvals

### Testing Requirements

#### Unit Tests
```go
// models/model_test.go
func TestDownloadModel(t *testing.T) {
    tests := []struct {
        name    string
        model   string
        want    error
        setup   func()
        cleanup func()
    }{
        {
            name:  "successful download",
            model: "llama2",
            want:  nil,
            setup: func() {
                // Setup test environment
            },
            cleanup: func() {
                // Cleanup test data
            },
        },
        {
            name:  "invalid model name",
            model: "",
            want:  ErrInvalidModelName,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if tt.setup != nil {
                tt.setup()
            }
            defer func() {
                if tt.cleanup != nil {
                    tt.cleanup()
                }
            }()
            
            got := DownloadModel(context.Background(), tt.model)
            if !errors.Is(got, tt.want) {
                t.Errorf("DownloadModel() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

#### Integration Tests
```go
// tests/integration/cluster_test.go
func TestClusterFormation(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    // Start test cluster
    cluster := NewTestCluster(t, 3)
    defer cluster.Cleanup()
    
    // Wait for cluster formation
    require.Eventually(t, func() bool {
        return cluster.IsFormed()
    }, 30*time.Second, time.Second)
    
    // Test leader election
    leader := cluster.GetLeader()
    require.NotEmpty(t, leader)
    
    // Test model distribution
    err := cluster.DownloadModel("test-model")
    require.NoError(t, err)
    
    // Verify model is distributed
    require.Eventually(t, func() bool {
        return cluster.ModelDistributed("test-model")
    }, 60*time.Second, 5*time.Second)
}
```

## Testing & Debugging

### Local Testing

#### Running Tests
```bash
# Run all tests
make test

# Run specific test package
go test ./pkg/models/...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run integration tests
make test-integration

# Run performance benchmarks
go test -bench=. ./pkg/performance/...

# Run race detection
go test -race ./...
```

#### Test Categories
```bash
# Unit tests (fast)
go test -short ./...

# Integration tests (slower)
go test -tags=integration ./tests/integration/...

# End-to-end tests (slowest)
go test -tags=e2e ./tests/e2e/...

# Performance tests
go test -bench=. -benchmem ./pkg/performance/...
```

### Debugging

#### Debugging with Delve
```bash
# Debug main application
dlv debug cmd/distributed-ollama/main.go -- --config config/dev.yaml

# Debug specific test
dlv test ./pkg/models -- -test.run TestDownloadModel

# Attach to running process
dlv attach $(pgrep ollama-distributed)

# Remote debugging
dlv debug --headless --listen=:2345 --api-version=2 cmd/distributed-ollama/main.go
```

#### Debug Configuration
```go
// Enable debug logging
import "github.com/ollama/ollama-distributed/pkg/logger"

func main() {
    logger.SetLevel(logger.DebugLevel)
    logger.EnableDebugMode()
    
    // Your application code
}
```

#### Profiling
```go
// Enable profiling
import _ "net/http/pprof"

func main() {
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
    
    // Your application code
}
```

```bash
# CPU profiling
go tool pprof http://localhost:6060/debug/pprof/profile

# Memory profiling
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine profiling
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

### Performance Testing

#### Load Testing
```bash
# Install load testing tools
go install github.com/grafana/k6@latest

# Run load tests
k6 run tests/load/inference-test.js

# Distributed load testing
k6 run --out cloud tests/load/cluster-test.js
```

#### Load Test Example
```javascript
// tests/load/inference-test.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
    vus: 100, // Virtual users
    duration: '5m',
    thresholds: {
        http_req_duration: ['p(95)<500'], // 95% of requests under 500ms
        http_req_failed: ['rate<0.01'],   // Error rate under 1%
    },
};

export default function() {
    const payload = JSON.stringify({
        model: 'llama2',
        prompt: 'Explain artificial intelligence',
    });
    
    const response = http.post('http://localhost:8080/api/v1/generate', payload, {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer ' + __ENV.API_KEY,
        },
    });
    
    check(response, {
        'status is 200': (r) => r.status === 200,
        'response time < 500ms': (r) => r.timings.duration < 500,
        'has response': (r) => r.json().response !== undefined,
    });
    
    sleep(1);
}
```

---

## Next Steps

- **Advanced Architecture**: [Architecture Deep Dive](./architecture-guide.md)
- **API Reference**: [Complete API Documentation](../api/reference.md)
- **Security Implementation**: [Security Guide](./security-guide.md)
- **Performance Optimization**: [Performance Guide](./performance-guide.md)
- **Production Deployment**: [Operations Guide](./operations-guide.md)

For questions and support, join our developer community on [Discord](https://discord.gg/ollama) or check our [GitHub Discussions](https://github.com/ollama/ollama-distributed/discussions).