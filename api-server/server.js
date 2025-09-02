/**
 * Distributed Llama API Server
 * Manages WebSocket connections and load balancing across multiple Ollama nodes
 */

const WebSocket = require('ws');
const http = require('http');
const express = require('express');
const cors = require('cors');
const Redis = require('ioredis');

const app = express();
app.use(cors());
app.use(express.json());

// Configuration
const PORT = process.env.PORT || 13100;
const REDIS_HOST = process.env.REDIS_HOST || 'localhost';
const REDIS_PORT = process.env.REDIS_PORT || 6379;
const REDIS_PASSWORD = process.env.REDIS_PASSWORD || 'ollama_redis_pass';

// Initialize Redis for distributed state management
const redis = new Redis({
    host: REDIS_HOST,
    port: REDIS_PORT,
    password: REDIS_PASSWORD,
    retryStrategy: (times) => Math.min(times * 50, 2000),
    connectTimeout: 10000,
    maxRetriesPerRequest: 3
});

// Node registry
class NodeRegistry {
    constructor() {
        this.nodes = new Map();
        this.loadBalancer = new LoadBalancer();
    }

    addNode(id, config) {
        const node = {
            id,
            name: config.name,
            url: config.url,
            status: 'connecting',
            health: {
                load: 0,
                memory: 0,
                requestsPerSecond: 0,
                queue: 0,
                lastCheck: Date.now()
            },
            connection: null
        };
        
        this.nodes.set(id, node);
        this.connectToNode(node);
        return node;
    }

    connectToNode(node) {
        try {
            // Connect to Ollama instance
            const ollamaUrl = node.url.replace('http', 'ws');
            node.connection = new WebSocket(ollamaUrl);
            
            node.connection.on('open', () => {
                console.log(`Connected to node: ${node.name}`);
                node.status = 'healthy';
                this.updateNodeHealth(node);
            });
            
            node.connection.on('error', (error) => {
                console.error(`Node ${node.name} error:`, error);
                node.status = 'error';
            });
            
            node.connection.on('close', () => {
                console.log(`Node ${node.name} disconnected`);
                node.status = 'offline';
                // Attempt reconnection after delay
                setTimeout(() => this.connectToNode(node), 5000);
            });
        } catch (error) {
            console.error(`Failed to connect to node ${node.name}:`, error);
            node.status = 'error';
        }
    }

    async updateNodeHealth(node) {
        try {
            // Fetch node metrics from Ollama API
            const response = await fetch(`${node.url}/api/tags`);
            if (response.ok) {
                const data = await response.json();
                node.health.lastCheck = Date.now();
                
                // Update load metrics (simulated for now)
                node.health.load = Math.random() * 100;
                node.health.memory = Math.random() * 100;
                node.health.requestsPerSecond = Math.floor(Math.random() * 20);
                node.health.queue = Math.floor(Math.random() * 10);
                
                // Update status based on health
                if (node.health.load > 90 || node.health.memory > 90) {
                    node.status = 'warning';
                } else {
                    node.status = 'healthy';
                }
            }
        } catch (error) {
            console.error(`Health check failed for ${node.name}:`, error);
            node.status = 'error';
        }
    }

    removeNode(id) {
        const node = this.nodes.get(id);
        if (node) {
            if (node.connection) {
                node.connection.close();
            }
            this.nodes.delete(id);
        }
    }

    getHealthyNodes() {
        return Array.from(this.nodes.values()).filter(n => n.status === 'healthy');
    }

    getAllNodes() {
        return Array.from(this.nodes.values()).map(node => ({
            id: node.id,
            name: node.name,
            status: node.status,
            ...node.health
        }));
    }

    selectNode(strategy = 'round-robin') {
        const healthyNodes = this.getHealthyNodes();
        if (healthyNodes.length === 0) return null;
        
        return this.loadBalancer.select(healthyNodes, strategy);
    }
}

// Load Balancing Strategies
class LoadBalancer {
    constructor() {
        this.currentIndex = 0;
    }

    select(nodes, strategy) {
        switch(strategy) {
            case 'round-robin':
                return this.roundRobin(nodes);
            case 'least-loaded':
                return this.leastLoaded(nodes);
            case 'fastest':
                return this.fastest(nodes);
            default:
                return this.roundRobin(nodes);
        }
    }

    roundRobin(nodes) {
        const node = nodes[this.currentIndex % nodes.length];
        this.currentIndex++;
        return node;
    }

    leastLoaded(nodes) {
        return nodes.reduce((min, node) => 
            node.health.load < min.health.load ? node : min
        );
    }

    fastest(nodes) {
        return nodes.reduce((min, node) => 
            node.health.requestsPerSecond > min.health.requestsPerSecond ? node : min
        );
    }
}

// Message Queue for handling requests
class MessageQueue {
    constructor() {
        this.queue = [];
        this.processing = false;
    }

    add(message, callback) {
        this.queue.push({ message, callback });
        this.process();
    }

    async process() {
        if (this.processing || this.queue.length === 0) return;
        
        this.processing = true;
        
        while (this.queue.length > 0) {
            const { message, callback } = this.queue.shift();
            
            try {
                await callback(message);
            } catch (error) {
                console.error('Queue processing error:', error);
            }
            
            // Small delay between processing
            await new Promise(resolve => setTimeout(resolve, 100));
        }
        
        this.processing = false;
    }

    getLength() {
        return this.queue.length;
    }
}

// Initialize components
const nodeRegistry = new NodeRegistry();
const messageQueue = new MessageQueue();

// Create HTTP server
const server = http.createServer(app);

// Create WebSocket server
const wss = new WebSocket.Server({ server });

// WebSocket connection handler
wss.on('connection', (ws) => {
    console.log('New WebSocket client connected');
    
    // Send initial node status
    ws.send(JSON.stringify({
        type: 'node_update',
        nodes: nodeRegistry.getAllNodes()
    }));
    
    ws.on('message', async (message) => {
        try {
            const data = JSON.parse(message);
            
            switch(data.type) {
                case 'inference':
                    await handleInference(ws, data);
                    break;
                case 'add_node':
                    handleAddNode(ws, data);
                    break;
                case 'remove_node':
                    handleRemoveNode(ws, data);
                    break;
                case 'get_nodes':
                    handleGetNodes(ws);
                    break;
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
    });
});

// Inference handling
async function handleInference(ws, data) {
    const startTime = Date.now();
    
    // Select node using configured strategy
    const node = nodeRegistry.selectNode(data.loadBalancing || 'round-robin');
    
    if (!node) {
        ws.send(JSON.stringify({
            type: 'error',
            message: 'No healthy nodes available'
        }));
        return;
    }
    
    console.log(`Routing request to node: ${node.name}`);
    
    // Store in Redis for distributed tracking
    await redis.set(`request:${data.timestamp}`, JSON.stringify({
        node: node.name,
        model: data.model,
        startTime
    }));
    
    // Send initial response
    ws.send(JSON.stringify({
        type: 'response',
        id: data.timestamp,
        node: node.name,
        streaming: data.settings.streaming
    }));
    
    // Process inference
    if (data.settings.streaming) {
        await streamInference(ws, node, data);
    } else {
        await completeInference(ws, node, data);
    }
    
    // Update metrics
    const latency = Date.now() - startTime;
    ws.send(JSON.stringify({
        type: 'metrics',
        latency,
        node: node.name
    }));
}

async function streamInference(ws, node, data) {
    try {
        // Make request to Ollama API
        const response = await fetch(`${node.url}/api/generate`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                model: data.model || 'llama2',
                prompt: data.content,
                stream: true,
                options: {
                    temperature: data.settings.temperature,
                    num_predict: data.settings.maxTokens
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
                    
                    ws.send(JSON.stringify({
                        type: 'stream_chunk',
                        id: data.timestamp,
                        chunk: json.response,
                        done: json.done
                    }));
                    
                    if (json.done) {
                        return;
                    }
                } catch (e) {
                    // Ignore JSON parse errors for partial chunks
                }
            }
        }
    } catch (error) {
        console.error('Streaming error:', error);
        ws.send(JSON.stringify({
            type: 'error',
            message: `Streaming failed: ${error.message}`
        }));
    }
}

async function completeInference(ws, node, data) {
    try {
        const response = await fetch(`${node.url}/api/generate`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                model: data.model || 'llama2',
                prompt: data.content,
                stream: false,
                options: {
                    temperature: data.settings.temperature,
                    num_predict: data.settings.maxTokens
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

// Node management handlers
function handleAddNode(ws, data) {
    const nodeId = `node-${Date.now()}`;
    const node = nodeRegistry.addNode(nodeId, data.node);
    
    ws.send(JSON.stringify({
        type: 'node_added',
        node: {
            id: node.id,
            name: node.name,
            status: node.status
        }
    }));
    
    // Broadcast to all clients
    broadcastNodeUpdate();
}

function handleRemoveNode(ws, data) {
    nodeRegistry.removeNode(data.nodeId);
    
    ws.send(JSON.stringify({
        type: 'node_removed',
        nodeId: data.nodeId
    }));
    
    broadcastNodeUpdate();
}

function handleGetNodes(ws) {
    ws.send(JSON.stringify({
        type: 'node_update',
        nodes: nodeRegistry.getAllNodes()
    }));
}

function broadcastNodeUpdate() {
    const update = JSON.stringify({
        type: 'node_update',
        nodes: nodeRegistry.getAllNodes()
    });
    
    wss.clients.forEach(client => {
        if (client.readyState === WebSocket.OPEN) {
            client.send(update);
        }
    });
}

// REST API endpoints
app.get('/api/nodes', (req, res) => {
    res.json({
        nodes: nodeRegistry.getAllNodes(),
        queueLength: messageQueue.getLength()
    });
});

app.post('/api/nodes', (req, res) => {
    const nodeId = `node-${Date.now()}`;
    const node = nodeRegistry.addNode(nodeId, req.body);
    
    res.json({
        id: node.id,
        name: node.name,
        status: node.status
    });
    
    broadcastNodeUpdate();
});

app.delete('/api/nodes/:id', (req, res) => {
    nodeRegistry.removeNode(req.params.id);
    res.json({ success: true });
    broadcastNodeUpdate();
});

app.get('/api/health', (req, res) => {
    res.json({
        status: 'healthy',
        nodes: nodeRegistry.getHealthyNodes().length,
        totalNodes: nodeRegistry.nodes.size,
        queueLength: messageQueue.getLength(),
        uptime: process.uptime()
    });
});

// Periodic health checks
setInterval(() => {
    nodeRegistry.nodes.forEach(node => {
        nodeRegistry.updateNodeHealth(node);
    });
    broadcastNodeUpdate();
}, 5000);

// Initialize default nodes from environment
function initializeDefaultNodes() {
    const defaultNodes = [];
    
    // Add nodes from environment variables
    if (process.env.OLLAMA_PRIMARY) {
        defaultNodes.push({ name: 'ollama-primary', url: process.env.OLLAMA_PRIMARY });
    }
    if (process.env.OLLAMA_WORKER_2) {
        defaultNodes.push({ name: 'ollama-worker-2', url: process.env.OLLAMA_WORKER_2 });
    }
    if (process.env.OLLAMA_WORKER_3) {
        defaultNodes.push({ name: 'ollama-worker-3', url: process.env.OLLAMA_WORKER_3 });
    }
    
    // Fallback to default nodes if no environment variables set
    if (defaultNodes.length === 0) {
        defaultNodes.push(
            { name: 'ollama-primary', url: 'http://localhost:13000' },
            { name: 'ollama-worker-2', url: 'http://localhost:13001' },
            { name: 'ollama-worker-3', url: 'http://localhost:13002' }
        );
    }
    
    // Check if running in Docker Swarm
    if (process.env.DOCKER_SWARM === 'true') {
        // Use service discovery
        defaultNodes.push(
            { name: 'llama-swarm-1', url: 'http://ollama_1:11434' },
            { name: 'llama-swarm-2', url: 'http://ollama_2:11434' },
            { name: 'llama-swarm-3', url: 'http://ollama_3:11434' }
        );
    }
    
    console.log(`Initializing nodes:`, defaultNodes.map(n => `${n.name}: ${n.url}`));
    
    defaultNodes.forEach((config, index) => {
        const nodeId = `node-${index}`;
        nodeRegistry.addNode(nodeId, config);
    });
}

// Start server
server.listen(PORT, () => {
    console.log(`Distributed Llama API Server running on port ${PORT}`);
    console.log(`WebSocket endpoint: ws://localhost:${PORT}/chat`);
    console.log(`REST API: http://localhost:${PORT}/api`);
    
    // Initialize default nodes
    initializeDefaultNodes();
    
    // Connect to Redis
    redis.on('connect', () => {
        console.log('Connected to Redis for distributed state management');
    });
    
    redis.on('error', (error) => {
        console.error('Redis connection error:', error);
    });
});