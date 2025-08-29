# WebSocket API

Real-time communication using WebSocket connections.

## Connection

Connect to the WebSocket endpoint:

```javascript
const ws = new WebSocket('wss://api.ollama-distributed.example.com/ws');

ws.onopen = () => {
  // Send authentication
  ws.send(JSON.stringify({
    type: 'auth',
    token: 'your-jwt-token'
  }));
};
```

## Message Types

### Authentication
```json
{
  "type": "auth",
  "token": "jwt-token-here"
}
```

### Subscribe to Events
```json
{
  "type": "subscribe",
  "channels": ["cluster", "models", "inference"]
}
```

### Real-time Generation
```json
{
  "type": "generate",
  "model": "llama2",
  "prompt": "Tell me a story",
  "stream": true
}
```

## Event Types

### Cluster Events
- `node_joined`: New node added to cluster
- `node_left`: Node removed from cluster
- `node_status_changed`: Node status update

### Model Events
- `model_loaded`: Model loaded on node
- `model_unloaded`: Model unloaded from node
- `model_replicated`: Model replicated to new node

### Inference Events
- `inference_started`: New inference request
- `inference_completed`: Inference finished
- `inference_failed`: Inference error

## Example Implementation

```javascript
const ws = new WebSocket('wss://api.ollama-distributed.example.com/ws');

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  
  switch (message.type) {
    case 'auth_success':
      console.log('Authenticated successfully');
      ws.send(JSON.stringify({
        type: 'subscribe',
        channels: ['cluster']
      }));
      break;
      
    case 'cluster_event':
      console.log('Cluster event:', message.data);
      break;
      
    case 'generation_chunk':
      process.stdout.write(message.data.text);
      break;
  }
};
```