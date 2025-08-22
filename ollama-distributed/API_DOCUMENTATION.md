# Ollama Distributed API Documentation

## Overview

The Ollama Distributed API provides a comprehensive REST API for managing distributed AI model inference across multiple nodes. This API enables model distribution, cluster management, real-time monitoring, and WebSocket-based live updates.

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

The API uses JWT-based authentication for secure endpoints. Authentication is required for all protected endpoints.

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json
```

**Authentication Endpoints:**
- `POST /api/v1/auth/login` - Authenticate and receive JWT token
- `POST /api/v1/auth/refresh` - Refresh JWT token
- `POST /api/v1/auth/logout` - Logout and invalidate token

## Endpoints

### 🏥 Health & Status

#### GET /health
Returns the health status of the distributed system.

**Authentication:** None required

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:00:00Z",
  "uptime": "24h 15m 30s"
}
```

**Status Codes:**
- `200 OK` - System is healthy
- `503 Service Unavailable` - System is unhealthy

#### GET /version
Returns version information about the system.

**Authentication:** None required

**Response:**
```json
{
  "version": "1.0.0",
  "build_time": "2024-01-15T10:00:00Z",
  "git_commit": "abc123def456",
  "go_version": "go1.21.0",
  "platform": "linux/amd64",
  "api_version": "v1"
}
```

#### GET /cluster/status
Returns detailed cluster status information.

**Response:**
```json
{
  "node_id": "QmXxX...",
  "is_leader": true,
  "leader": "QmXxX...",
  "peers": 3,
  "status": "healthy"
}
```

### 🖥️ Node Management

#### GET /nodes
Returns all nodes in the distributed cluster.

**Authentication:** Required

**Response:**
```json
{
  "nodes": [
    {
      "id": "node-001",
      "address": "192.168.1.100:8080",
      "status": "online",
      "last_seen": "2024-01-15T10:00:00Z",
      "models": ["llama2:7b", "codellama:13b"],
      "capacity": {
        "cpu_cores": 8,
        "memory_gb": 32,
        "disk_gb": 500,
        "gpu_count": 1
      },
      "usage": {
        "cpu_usage": 0.45,
        "memory_usage": 0.67,
        "disk_usage": 0.23,
        "gpu_usage": 0.12
      }
    }
  ]
}
```

#### GET /nodes/:id
Returns information about a specific node.

**Parameters:**
- `id` (string): Node ID

**Response:**
```json
{
  "node": {
    "id": "node_001",
    "address": "192.168.1.100:8080",
    "status": "online",
    "models": ["llama2", "mistral"],
    "usage": {
      "cpu": 45.2,
      "memory": 67.8,
      "bandwidth": 12.5
    }
  }
}
```

#### POST /nodes/:id/drain
Drains a node (prevents new requests from being routed to it).

**Parameters:**
- `id` (string): Node ID

**Response:**
```json
{
  "message": "Node node_001 is being drained"
}
```

#### POST /nodes/:id/undrain
Removes drain status from a node.

**Parameters:**
- `id` (string): Node ID

**Response:**
```json
{
  "message": "Node node_001 is no longer draining"
}
```

### 🧠 Model Management

#### GET /models
Returns all models available in the distributed system.

**Authentication:** Required

**Response:**
```json
{
  "models": [
    {
      "name": "llama2:7b",
      "size": 3826793677,
      "digest": "sha256:fe938a131f40e6f6d40083c9f0f430a515233eb2edaa6d72eb85c50d64f2300e",
      "modified_at": "2024-01-15T10:00:00Z",
      "details": {
        "parent_model": "",
        "format": "gguf",
        "family": "llama",
        "families": ["llama"],
        "parameter_size": "7B",
        "quantization_level": "Q4_0"
      }
    }
  ]
}
```

#### GET /models/:name
Returns information about a specific model.

**Parameters:**
- `name` (string): Model name

**Response:**
```json
{
  "model": {
    "name": "llama2",
    "size": 7340032,
    "status": "available",
    "replicas": ["node_001", "node_002"],
    "inference_ready": true
  }
}
```

#### POST /models/:name/download
Initiates download of a model to the distributed system.

**Authentication:** Required

**Parameters:**
- `name` (string): Model name

**Response:**
```json
{
  "status": "downloading",
  "model": "llama2:7b",
  "message": "Model download started"
}
```

#### DELETE /models/:name
Removes a model from the distributed system.

**Parameters:**
- `name` (string): Model name

**Response:**
```json
{
  "message": "Successfully deleted model llama2",
  "request_id": "delete_llama2_1642694400",
  "node_id": "node_001"
}
```

### ⚡ Distribution Management

#### POST /distribution/auto-configure
Enables or disables automatic model distribution.

**Request Body:**
```json
{
  "enabled": true
}
```

**Response:**
```json
{
  "message": "Auto-distribution enabled",
  "status": "success"
}
```

### 🔄 Transfer Management

#### GET /transfers
Returns all active and recent model transfers.

**Response:**
```json
{
  "transfers": [
    {
      "id": "transfer_001",
      "model_name": "llama2",
      "type": "download",
      "status": "active",
      "progress": 45.2,
      "speed": 2415919, // bytes per second
      "eta": 120, // seconds
      "peer_id": "peer_001",
      "node_id": "node_001"
    }
  ]
}
```

#### GET /transfers/:id
Returns information about a specific transfer.

**Parameters:**
- `id` (string): Transfer ID

**Response:**
```json
{
  "transfer": {
    "id": "transfer_001",
    "model_name": "llama2",
    "type": "download",
    "status": "active",
    "progress": 45.2,
    "speed": 2415919,
    "eta": 120,
    "peer_id": "peer_001",
    "node_id": "node_001"
  }
}
```

### 📊 Metrics & Monitoring

#### GET /metrics
Returns comprehensive system metrics.

**Response:**
```json
{
  "node_id": "QmXxX...",
  "connected_peers": 3,
  "is_leader": true,
  "requests_processed": 1425,
  "models_loaded": 5,
  "cluster_size": 4,
  "scheduler_stats": {
    "total_requests": 1425,
    "completed_requests": 1398,
    "failed_requests": 27,
    "queued_requests": 0,
    "average_latency": 156.7
  },
  "uptime": "running",
  "status": "healthy",
  "cpu_usage": 35.2,
  "memory_usage": 62.8,
  "network_usage": 18.5,
  "requests_per_second": 23.75,
  "average_latency": 156,
  "active_connections": 3,
  "error_rate": 0.5
}
```

### 🔗 Cluster Operations

#### GET /cluster/leader
Returns the current cluster leader.

**Response:**
```json
{
  "leader": "QmXxX..."
}
```

#### POST /cluster/join
Joins a node to the cluster.

**Request Body:**
```json
{
  "node_id": "QmYyY...",
  "address": "192.168.1.101:8080"
}
```

**Response:**
```json
{
  "message": "Node joined cluster"
}
```

#### POST /cluster/leave
Removes a node from the cluster.

**Request Body:**
```json
{
  "node_id": "QmYyY..."
}
```

**Response:**
```json
{
  "message": "Node left cluster"
}
```

### 🤖 Inference Endpoints

#### POST /generate
Generates text using a distributed model.

**Authentication:** Required

**Request Body:**
```json
{
  "model": "llama2:7b",
  "prompt": "Hello, how are you?",
  "stream": false,
  "options": {},
  "context": []
}
```

**Response:**
```json
{
  "model": "llama2:7b",
  "response": "This is a mock response to your prompt: Hello, how are you?",
  "done": true,
  "context": [1, 2, 3, 4, 5],
  "created_at": "2024-01-15T10:00:00Z",
  "total_duration": 1500000000,
  "load_duration": 500000000,
  "prompt_eval_count": 25,
  "prompt_eval_duration": 200000000,
  "eval_count": 50,
  "eval_duration": 800000000
}
```

#### POST /chat
Performs chat completion using a distributed model.

**Authentication:** Required

**Request Body:**
```json
{
  "model": "llama2:7b",
  "messages": [
    {
      "role": "user",
      "content": "Hello!"
    }
  ],
  "stream": false,
  "options": {}
}
```

**Response:**
```json
{
  "model": "llama2:7b",
  "message": {
    "role": "assistant",
    "content": "This is a mock chat response."
  },
  "done": true,
  "created_at": "2024-01-15T10:00:00Z",
  "total_duration": 1200000000,
  "load_duration": 300000000,
  "prompt_eval_count": 20,
  "prompt_eval_duration": 150000000,
  "eval_count": 40,
  "eval_duration": 750000000
}
```

#### POST /embeddings
Generates embeddings using a distributed model.

**Authentication:** Required

**Request Body:**
```json
{
  "model": "llama2:7b",
  "prompt": "Hello world"
}
```

**Response:**
```json
{
  "model": "llama2:7b",
  "embedding": [0.001, 0.002, 0.003, ...]
}
```

**Note:** The embedding array contains 4096 float64 values representing the text embedding.

## 🔌 WebSocket API

### Connection

Connect to the WebSocket endpoint:

```
ws://localhost:8080/ws
```

**Authentication:** Required via query parameter or header

**Query Parameter:**
```
ws://localhost:8080/ws?token=<JWT_TOKEN>
```

**Or Header:**
```
Authorization: Bearer <JWT_TOKEN>
```

### Message Types

#### Client to Server

**Ping:**
```json
{
  "type": "ping"
}
```

**Subscribe to Channel:**
```json
{
  "type": "subscribe",
  "channel": "metrics"
}
```

**Unsubscribe from Channel:**
```json
{
  "type": "unsubscribe",
  "channel": "metrics"
}
```

#### Server to Client

**Pong:**
```json
{
  "type": "pong"
}
```

**Real-time Metrics:**
```json
{
  "type": "metrics",
  "data": {
    "cpu": 35.2,
    "memory": 62.8,
    "network": 18.5,
    "nodes": 4,
    "models": 5,
    "peers": 3
  }
}
```

**Cluster Status:**
```json
{
  "type": "cluster_status",
  "data": {
    "node_id": "QmXxX...",
    "is_leader": true,
    "leader": "QmXxX...",
    "peers": 3,
    "status": "healthy"
  }
}
```

**Model Events:**
```json
{
  "type": "model_event",
  "event_type": "download_started",
  "model_name": "llama2",
  "data": {
    "node_id": "node_001",
    "status": "downloading"
  },
  "timestamp": 1642694400
}
```

**Node Events:**
```json
{
  "type": "node_event",
  "event_type": "node_online",
  "node_id": "node_001",
  "data": {
    "address": "192.168.1.100:8080",
    "status": "online"
  },
  "timestamp": 1642694400
}
```

**Alerts:**
```json
{
  "type": "alert",
  "level": "warning",
  "message": "Node node_001 is running low on memory",
  "timestamp": 1642694400
}
```

**Subscription Acknowledgment:**
```json
{
  "type": "subscription_ack",
  "channel": "metrics",
  "status": "subscribed"
}
```

## 🚨 Error Handling

### HTTP Status Codes

- `200 OK` - Request successful
- `400 Bad Request` - Invalid request parameters
- `401 Unauthorized` - Authentication required
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error
- `503 Service Unavailable` - Service temporarily unavailable
- `408 Request Timeout` - Request timeout

### Error Response Format

```json
{
  "error": "Description of the error",
  "code": "ERROR_CODE",
  "details": {
    "field": "Additional error details"
  }
}
```

## 📈 Rate Limiting

The API implements rate limiting to prevent abuse:

- **General endpoints**: 100 requests per minute
- **Model operations**: 10 requests per minute
- **WebSocket connections**: 5 connections per IP

## 🔧 Configuration

### Environment Variables

- `OLLAMA_JWT_SECRET` - JWT signing secret
- `API_LISTEN` - API server listen address (default: "0.0.0.0:8080")
- `P2P_LISTEN` - P2P network listen address (default: "/ip4/0.0.0.0/tcp/9000")
- `RAFT_BIND_ADDR` - Raft consensus bind address (default: "0.0.0.0:7000")
- `LOG_LEVEL` - Logging level (default: "info")
- `NODE_ID` - Unique node identifier
- `BOOTSTRAP` - Whether this node is a bootstrap node (default: false)

### Configuration Files

Configuration is managed through YAML files in the `config/` directory:

- `config.yaml` - Main configuration
- `node.yaml` - Node-specific configuration

## 🧪 Testing

### Integration Test

Run the comprehensive integration test:

```bash
go run integration_test.go
```

### Manual Testing

Use curl to test endpoints:

```bash
# Health check
curl http://localhost:8080/api/v1/health

# Get cluster status
curl http://localhost:8080/api/v1/cluster/status

# Download a model
curl -X POST http://localhost:8080/api/v1/models/llama2/download

# Enable auto-distribution
curl -X POST -H "Content-Type: application/json" \
  -d '{"enabled": true}' \
  http://localhost:8080/api/v1/distribution/auto-configure
```

## 📝 Notes

- All timestamps are Unix timestamps in seconds
- Model sizes are in bytes
- Bandwidth is measured in bytes per second
- Latency is measured in milliseconds
- CPU, memory, and network usage are percentages (0-100)

## 🔮 Future Enhancements

- Authentication improvements (JWT tokens)
- More granular rate limiting
- Advanced model scheduling algorithms
- Enhanced metrics and monitoring
- Multi-tenancy support
- API versioning