/**
 * WebSocket Service for Real-time Communication
 * Handles chat, node management, and real-time updates
 */

const WebSocket = require('ws');
const NodeRegistry = require('./node-registry');
const MessageQueue = require('./message-queue');

class WebSocketService {
  constructor(server) {
    this.wss = new WebSocket.Server({ noServer: true });
    this.server = server;
    this.nodeRegistry = new NodeRegistry();
    this.messageQueue = new MessageQueue();
    this.clients = new Set();
    
    this.setupUpgradeHandler();
    this.setupConnectionHandler();
    this.initializeMockNodes();
  }

  setupUpgradeHandler() {
    this.server.on('upgrade', (request, socket, head) => {
      const { pathname } = new URL(request.url, `http://${request.headers.host}`);

      // Only allow WebSocket connections on /chat path (SECURITY)
      if (pathname === '/chat') {
        this.wss.handleUpgrade(request, socket, head, (ws) => {
          this.wss.emit('connection', ws, request);
        });
      } else {
        console.log(`WebSocket connection rejected: invalid path ${pathname}`);
        socket.destroy();
      }
    });
  }

  setupConnectionHandler() {
    this.wss.on('connection', (ws, request) => {
      console.log('New WebSocket client connected');
      this.clients.add(ws);
      
      // Send initial node status
      ws.send(JSON.stringify({
        type: 'node_update',
        nodes: this.nodeRegistry.getAllNodes()
      }));
      
      ws.on('message', async (message) => {
        try {
          const data = JSON.parse(message);
          
          switch(data.type) {
            case 'inference':
              await this.handleInference(ws, data);
              break;
            case 'add_node':
              this.handleAddNode(ws, data);
              break;
            case 'remove_node':
              this.handleRemoveNode(ws, data);
              break;
            case 'get_nodes':
              this.handleGetNodes(ws);
              break;
            default:
              console.log('Unknown message type:', data.type);
          }
        } catch (error) {
          console.error('Message handling error:', error);
          ws.send(JSON.stringify({
            type: 'error',
            message: error.message
          }));
        }
      });
      
      ws.on('close', () => {
        console.log('Client disconnected');
        this.clients.delete(ws);
      });
      
      ws.on('error', (error) => {
        console.error('WebSocket error:', error);
        this.clients.delete(ws);
      });
    });
  }

  async handleInference(ws, data) {
    const startTime = Date.now();
    
    // Select node using configured strategy
    const node = this.nodeRegistry.selectNode(data.loadBalancing || 'round-robin');
    
    if (!node) {
      ws.send(JSON.stringify({
        type: 'error',
        message: 'No healthy nodes available'
      }));
      return;
    }
    
    console.log(`Routing request to node: ${node.name}`);
    
    // Send initial response
    ws.send(JSON.stringify({
      type: 'response',
      id: data.timestamp,
      node: node.name,
      streaming: data.settings?.streaming || false
    }));
    
    // Process inference
    if (data.settings?.streaming) {
      await this.streamInference(ws, node, data);
    } else {
      await this.completeInference(ws, node, data);
    }
    
    // Update metrics
    const latency = Date.now() - startTime;
    ws.send(JSON.stringify({
      type: 'metrics',
      latency,
      node: node.name
    }));
  }

  async streamInference(ws, node, data) {
    try {
      const response = await fetch(`${node.url}/api/generate`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          model: data.model || 'llama2',
          prompt: data.content,
          stream: true,
          options: {
            temperature: data.settings?.temperature || 0.7,
            num_predict: data.settings?.maxTokens || 2048
          }
        })
      });
      
      const reader = response.body.getReader();
      const decoder = new TextDecoder();
      
      while (true) {
        const { done, value } = await reader.read();
        if (done) break;
        
        const chunk = decoder.decode(value);
        const lines = chunk.split('\n').filter(line => line.trim());
        
        for (const line of lines) {
          try {
            const json = JSON.parse(line);

            if (json.response) {
              ws.send(JSON.stringify({
                type: 'stream_chunk',
                id: data.timestamp,
                content: json.response,
                done: json.done || false
              }));
            }
          } catch (e) {
            console.error('Error parsing stream chunk:', e);
          }
        }
      }
    } catch (error) {
      console.error('Streaming inference error:', error);
      ws.send(JSON.stringify({
        type: 'error',
        message: `Streaming failed: ${error.message}`
      }));
    }
  }

  async completeInference(ws, node, data) {
    try {
      const response = await fetch(`${node.url}/api/generate`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          model: data.model || 'llama2',
          prompt: data.content,
          stream: false,
          options: {
            temperature: data.settings?.temperature || 0.7,
            num_predict: data.settings?.maxTokens || 2048
          }
        })
      });

      const result = await response.json();

      ws.send(JSON.stringify({
        type: 'response',
        id: data.timestamp,
        content: result.response,
        node: node.name,
        streaming: false
      }));
    } catch (error) {
      console.error('Inference error:', error);
      ws.send(JSON.stringify({
        type: 'error',
        message: `Inference failed: ${error.message}`
      }));
    }
  }

  handleAddNode(ws, data) {
    const nodeId = `node-${Date.now()}`;
    const node = this.nodeRegistry.addNode(nodeId, data.node);

    ws.send(JSON.stringify({
      type: 'node_added',
      node: {
        id: node.id,
        name: node.name,
        status: node.status
      }
    }));

    this.broadcastNodeUpdate();
  }

  handleRemoveNode(ws, data) {
    this.nodeRegistry.removeNode(data.nodeId);

    ws.send(JSON.stringify({
      type: 'node_removed',
      nodeId: data.nodeId
    }));

    this.broadcastNodeUpdate();
  }

  handleGetNodes(ws) {
    ws.send(JSON.stringify({
      type: 'node_update',
      nodes: this.nodeRegistry.getAllNodes()
    }));
  }

  broadcastNodeUpdate() {
    const update = JSON.stringify({
      type: 'node_update',
      nodes: this.nodeRegistry.getAllNodes()
    });

    this.clients.forEach(client => {
      if (client.readyState === WebSocket.OPEN) {
        client.send(update);
      }
    });
  }

  initializeMockNodes() {
    // Initialize mock nodes if enabled
    const enableMockNodes = process.env.ENABLE_MOCK_NODES === 'true';
    const mockNodesCount = parseInt(process.env.MOCK_NODES_COUNT || '3');

    console.log(`Mock nodes check: ENABLE_MOCK_NODES=${process.env.ENABLE_MOCK_NODES}, enableMockNodes=${enableMockNodes}, count=${mockNodesCount}`);

    if (enableMockNodes) {
      console.log(`🤖 Initializing ${mockNodesCount} mock nodes for development...`);

      for (let i = 0; i < mockNodesCount; i++) {
        const nodeId = `mock-node-${i}`;
        const node = this.nodeRegistry.addNode(nodeId, {
          name: `Mock Node ${i + 1}`,
          url: `http://localhost:${11434 + i}`,
          mock: true
        });
        console.log(`  ✓ Added mock node: ${node.name} (${node.id})`);
      }
    } else {
      console.log('Mock nodes disabled');
    }
  }

  getNodeRegistry() {
    return this.nodeRegistry;
  }

  getMessageQueue() {
    return this.messageQueue;
  }
}

module.exports = WebSocketService;

