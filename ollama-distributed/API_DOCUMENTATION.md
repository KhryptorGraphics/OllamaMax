# Ollama Distributed API Documentation

## Overview

The Ollama Distributed API provides a comprehensive REST API for managing distributed AI model inference across multiple nodes. This API enables model distribution, cluster management, real-time monitoring, and WebSocket-based live updates.

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

Currently, the API uses simple token-based authentication for admin endpoints:

```
Authorization: Bearer <OLLAMA_ADMIN_TOKEN>
```

## Endpoints

### 🏥 Health & Status

#### GET /health
Returns the health status of the distributed system.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": 1642694400,
  "services": {
    "p2p": "healthy",
    "consensus": "healthy",
    "scheduler": "healthy"
  }
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

**Response:**
```json
{
  "nodes": {
    "node_001": {
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

**Response:**
```json
{
  "models": {
    "llama2": {
      "name": "llama2",
      "size": 7340032,
      "status": "available",
      "replicas": ["node_001", "node_002"],
      "inference_ready": true
    }
  }
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

**Parameters:**
- `name` (string): Model name

**Response:**
```json
{
  "message": "Download started for model llama2 on node node_001",
  "node_id": "node_001"
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

**Request Body:**
```json
{
  "model": "llama2",
  "prompt": "Hello, how are you?",
  "stream": false
}
```

**Response:**
```json
{
  "response": "Generated response text",
  "model": "llama2",
  "node_id": "node_001"
}
```

#### POST /chat
Performs chat completion using a distributed model.

**Request Body:**
```json
{
  "model": "llama2",
  "messages": [
    {
      "role": "user",
      "content": "Hello!"
    }
  ],
  "stream": false
}
```

**Response:**
```json
{
  "message": {
    "role": "assistant",
    "content": "Hello! How can I help you today?"
  }
}
```

#### POST /embeddings
Generates embeddings using a distributed model.

**Request Body:**
```json
{
  "model": "llama2",
  "input": "Hello world"
}
```

**Response:**
```json
{
  "embeddings": [0.1, 0.2, 0.3, 0.4, 0.5]
}
```

## 🔌 WebSocket API

### Connection

Connect to the WebSocket endpoint:

```
ws://localhost:8080/api/v1/ws
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

- `OLLAMA_ADMIN_TOKEN` - Admin authentication token
- `OLLAMA_LISTEN_ADDRESS` - Server listen address (default: ":8080")
- `OLLAMA_DISTRIBUTED_MODE` - Enable distributed mode (default: true)
- `OLLAMA_FALLBACK_MODE` - Enable fallback mode (default: true)

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