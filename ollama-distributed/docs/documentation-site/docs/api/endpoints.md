# API Endpoints

Complete reference for all available API endpoints based on actual implementation.

## Core API Endpoints

### Health Check
```http
GET /health
```

Returns system health status with service information.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-08-28T01:20:00Z",
  "version": "1.0.0",
  "node_id": "12D3KooW...",
  "services": {
    "p2p": true,
    "p2p_peers": 2,
    "consensus": true,
    "consensus_leader": false,
    "scheduler": true,
    "available_nodes": 3
  }
}
```

### Version Information
```http
GET /api/version
GET /api/v1/version
```

Returns API version and build information.

**Response:**
```json
{
  "version": "1.0.0",
  "build_date": "2024-01-01",
  "git_commit": "unknown",
  "go_version": "1.21+"
}
```

## Ollama-Compatible API

### Generate Text
```http
POST /api/generate
```

Generate text completion using a specified model.

**Request:**
```json
{
  "model": "llama2",
  "prompt": "Explain quantum computing",
  "stream": false
}
```

**Response:**
```json
{
  "model": "llama2",
  "response": "This is a placeholder response. Distributed inference not yet implemented.",
  "done": true
}
```

### Chat Completion
```http
POST /api/chat
```

Chat-style completions with conversation context.

**Request:**
```json
{
  "model": "llama2",
  "messages": [
    {"role": "user", "content": "Hello!"}
  ],
  "stream": false,
  "options": {}
}
```

**Response:**
```json
{
  "model": "llama2",
  "message": {
    "role": "assistant",
    "content": "This is a placeholder response. Distributed chat inference not yet implemented."
  },
  "done": true
}
```

### List Models
```http
GET /api/tags
```

Returns all available models.

**Response:**
```json
{
  "models": [
    {
      "name": "llama2:7b",
      "status": "available",
      "size": "3.8GB"
    }
  ]
}
```

### Pull Model
```http
POST /api/pull
```

Download a model to the cluster.

**Request:**
```json
{
  "name": "llama2:7b",
  "stream": false
}
```

**Response:**
```json
{
  "status": "pulling model manifest",
  "digest": "sha256:...",
  "total": 3825819519
}
```

### Delete Model
```http
DELETE /api/delete
```

Remove a model from the cluster.

**Request:**
```json
{
  "name": "llama2:7b"
}
```

### Show Model
```http
POST /api/show
```

Get detailed model information.

**Request:**
```json
{
  "name": "llama2:7b"
}
```

### Create Model
```http
POST /api/create
```

Create a model from a Modelfile.

### Copy Model
```http
POST /api/copy
```

Copy a model.

### Push Model
```http
POST /api/push
```

Push a model to a registry.

### Embeddings
```http
POST /api/embed
POST /api/embeddings
```

Generate embeddings for input text.

**Request:**
```json
{
  "model": "all-minilm",
  "prompt": "The sky is blue"
}
```

**Response:**
```json
{
  "embeddings": [[0.1, 0.2, 0.3, ...]]
}
```

## Distributed-Specific API

### Cluster Status
```http
GET /api/distributed/status
GET /api/v1/cluster/status
```

Get comprehensive cluster status.

**Response:**
```json
{
  "distributed_mode": true,
  "fallback_mode": true,
  "cluster_size": 3,
  "active_nodes": ["node1", "node2", "node3"],
  "scheduler_stats": {},
  "runner_stats": {},
  "integration_stats": {}
}
```

### List Nodes
```http
GET /api/distributed/nodes
GET /api/v1/nodes
```

List all cluster nodes and their status.

**Response:**
```json
{
  "nodes": [
    {
      "id": "node1",
      "status": "active",
      "address": "10.0.1.10:11434",
      "models": ["llama2:7b"],
      "resources": {
        "cpu": 0.15,
        "memory": 0.25,
        "disk": 0.20
      }
    }
  ]
}
```

### Distributed Models
```http
GET /api/distributed/models
```

List models distributed across the cluster.

**Response:**
```json
{
  "models": [
    {
      "name": "llama2:7b",
      "replicas": 2,
      "nodes": ["node1", "node2"],
      "total_size": "3.8GB",
      "status": "ready"
    }
  ]
}
```

### Model Replicas
```http
GET /api/distributed/models/{name}/replicas
```

Get replica information for a specific model.

### System Metrics
```http
GET /api/distributed/metrics
GET /metrics
```

Get system metrics (Prometheus format also available).

**Response:**
```json
{
  "timestamp": "2025-08-28T01:20:00Z",
  "node_id": "12D3KooW...",
  "connected_peers": 2,
  "is_leader": false,
  "requests_processed": 0,
  "models_loaded": 0,
  "nodes_total": 3,
  "nodes_online": 3,
  "uptime": 3600,
  "cpu_usage": 0.0,
  "memory_usage": 0.0,
  "network_usage": 0.0,
  "websocket_connections": 5
}
```

### Active Requests
```http
GET /api/distributed/requests
```

List currently active inference requests.

### Replication Status
```http
GET /api/distributed/replication/status
```

Get model replication status across nodes.

### Rebalance Models
```http
POST /api/distributed/rebalance
```

Trigger model rebalancing across nodes.

### Migrate Model
```http
POST /api/distributed/migrate
```

Migrate a model between nodes.

**Request:**
```json
{
  "model_name": "llama2:7b",
  "from_node": "node1",
  "to_node": "node2"
}
```

## OpenAI-Compatible API

### Chat Completions
```http
POST /v1/chat/completions
```

OpenAI-compatible chat completions.

**Request:**
```json
{
  "model": "llama2",
  "messages": [
    {"role": "user", "content": "Hello!"}
  ],
  "temperature": 0.7,
  "max_tokens": 150
}
```

### Completions
```http
POST /v1/completions
```

OpenAI-compatible text completions.

### Embeddings
```http
POST /v1/embeddings
```

OpenAI-compatible embeddings.

### List Models
```http
GET /v1/models
```

OpenAI-compatible model listing.

### Get Model
```http
GET /v1/models/{model}
```

Get specific model information.

## Node Management API

### Get Node
```http
GET /api/v1/nodes/{id}
```

Get detailed information about a specific node.

### Drain Node
```http
POST /api/v1/nodes/{id}/drain
```

Mark a node for draining (no new tasks).

### Undrain Node
```http
POST /api/v1/nodes/{id}/undrain
```

Remove drain status from a node.

## Model Management API

### Download Model
```http
POST /api/v1/models/{name}/download
```

Initiate model download to a specific node.

### Get Model
```http
GET /api/v1/models/{name}
```

Get detailed model information.

### Delete Model
```http
DELETE /api/v1/models/{name}
```

Remove model from the cluster.

## Transfer Management API

### List Transfers
```http
GET /api/v1/transfers
```

Get all active model transfers.

### Get Transfer
```http
GET /api/v1/transfers/{id}
```

Get specific transfer details.

### Cancel Transfer
```http
DELETE /api/v1/transfers/{id}
```

Cancel an active transfer.

## Admin API

### Set Mode
```http
POST /admin/mode
```

Switch between distributed and local mode.

**Request:**
```json
{
  "mode": "distributed"
}
```

**Headers:**
```
Authorization: Bearer <admin_token>
```

### Set Fallback
```http
POST /admin/fallback
```

Enable/disable fallback mode.

### Force Rebalance
```http
POST /admin/rebalance
```

Force immediate model rebalancing.

### Get Statistics
```http
GET /admin/stats
```

Get detailed system statistics.

## WebSocket API

### Connect
```
ws://localhost:11434/ws
```

Connect to real-time WebSocket API for:
- Live metrics updates
- Model transfer progress
- Cluster event notifications
- Request status updates

## HTTP Response Codes

| Code | Description | Usage |
|------|-------------|-------|
| 200 | OK | Successful request |
| 201 | Created | Resource created successfully |
| 400 | Bad Request | Invalid request parameters |
| 401 | Unauthorized | Authentication required |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Resource not found |
| 409 | Conflict | Resource conflict |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Server error |
| 503 | Service Unavailable | Service temporarily unavailable |

## Error Response Format

All error responses follow this format:

```json
{
  "error": "Detailed error message",
  "code": "ERROR_CODE",
  "timestamp": "2025-08-28T01:20:00Z",
  "request_id": "req_123456789"
}
```

## Authentication

Most endpoints support Bearer token authentication:

```http
Authorization: Bearer <your_token>
```

Admin endpoints require valid admin token set via `OLLAMA_ADMIN_TOKEN` environment variable.

## Rate Limiting

API endpoints are rate limited:
- Standard endpoints: 1000 requests/hour
- Inference endpoints: 100 requests/minute
- Admin endpoints: 60 requests/minute

Rate limit headers are included in responses:
```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
```