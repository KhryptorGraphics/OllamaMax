/**
 * Distributed Llama API Server
 * Manages WebSocket connections and load balancing across multiple Ollama nodes
 */

const WebSocket = require('ws');
const http = require('http');
const express = require('express');
const cors = require('cors');
const swaggerUi = require('swagger-ui-express');
const Redis = require('ioredis');
// const { OpenRouterClient } = require('../pkg/openrouter/client.js'); // Temporarily disabled due to ES module issues
const AuthSystem = require('./auth-system.js');

const app = express();
app.use(cors());
app.use(express.json());

// Configuration
const PORT = process.env.PORT || 13000;
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

// Initialize Authentication System
const auth = new AuthSystem();

// OpenAPI Specification
const openapiSpec = {
  openapi: "3.0.0",
  info: {
    title: "OllamaMax Distributed API",
    version: "1.0.0",
    description: "Enterprise-grade distributed AI inference platform with WebSocket support and load balancing"
  },
  servers: [
    {
      url: `http://localhost:${PORT}`,
      description: "Local development server"
    }
  ],
  paths: {
    "/health": {
      get: {
        summary: "Health Check",
        description: "Get system health status",
        responses: {
          "200": {
            description: "System is healthy",
            content: {
              "application/json": {
                schema: {
                  type: "object",
                  properties: {
                    status: { type: "string" },
                    nodes: { type: "number" },
                    totalNodes: { type: "number" },
                    queueLength: { type: "number" },
                    uptime: { type: "number" }
                  }
                }
              }
            }
          }
        }
      }
    },
    "/api/nodes": {
      get: {
        summary: "Get Nodes Status",
        description: "Get status of all cluster nodes",
        responses: {
          "200": {
            description: "List of nodes",
            content: {
              "application/json": {
                schema: {
                  type: "object",
                  properties: {
                    nodes: {
                      type: "array",
                      items: {
                        type: "object",
                        properties: {
                          id: { type: "string" },
                          name: { type: "string" },
                          status: { type: "string" },
                          load: { type: "number" },
                          memory: { type: "number" },
                          requestsPerSecond: { type: "number" },
                          queue: { type: "number" },
                          lastCheck: { type: "number" }
                        }
                      }
                    },
                    queueLength: { type: "number" }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
};

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

// Initialize OpenRouter integration
let openRouterClient = null;
/*
if (process.env.OPENROUTER_API_KEY) {
    openRouterClient = new OpenRouterClient({
        apiKey: process.env.OPENROUTER_API_KEY,
        baseURL: 'https://openrouter.ai/api/v1',
        timeout: 300000, // 5 minutes
        maxRetries: 3,
        retryDelay: 2000
    });
    console.log('OpenRouter integration enabled with Sonoma Sky Alpha model');
}
*/

// Create HTTP server
const server = http.createServer(app);

// =================== AUTHENTICATION ROUTES ===================

// Middleware to authenticate requests
async function authenticateToken(req, res, next) {
    const authHeader = req.headers['authorization'];
    const token = authHeader && authHeader.split(' ')[1];

    if (token == null) {
        return res.sendStatus(401);
    }

    try {
        const user = await auth.validateSession(token);
        if (!user) {
            return res.sendStatus(403);
        }
        req.user = user;
        next();
    } catch (error) {
        return res.sendStatus(403);
    }
}

// User Registration
app.post('/api/auth/register', async (req, res) => {
    try {
        const { username, email, password } = req.body;
        
        if (!username || !email || !password) {
            return res.status(400).json({
                success: false,
                message: 'Username, email, and password are required'
            });
        }

        if (password.length < 6) {
            return res.status(400).json({
                success: false,
                message: 'Password must be at least 6 characters long'
            });
        }

        const result = await auth.registerUser(username, email, password);
        
        // Send verification email
        await auth.sendVerificationEmail(email, result.verificationToken, username);
        
        res.json({
            success: true,
            message: 'User registered successfully! Please check your email for verification.',
            userId: result.userId
        });

    } catch (error) {
        console.error('Registration error:', error.message);
        res.status(400).json({
            success: false,
            message: error.message
        });
    }
});

// User Login
app.post('/api/auth/login', async (req, res) => {
    try {
        const { email, password } = req.body;
        
        if (!email || !password) {
            return res.status(400).json({
                success: false,
                message: 'Email and password are required'
            });
        }

        const result = await auth.loginUser(email, password);
        
        res.json({
            success: true,
            message: 'Login successful',
            token: result.token,
            user: result.user
        });

    } catch (error) {
        console.error('Login error:', error.message);
        res.status(401).json({
            success: false,
            message: error.message
        });
    }
});

// Email Verification
app.get('/api/verify-email', async (req, res) => {
    try {
        const { token } = req.query;
        
        if (!token) {
            return res.status(400).send(`
                <html><body style="font-family: Arial, sans-serif; text-align: center; padding: 50px;">
                    <h2 style="color: #dc2626;">❌ Invalid Verification Link</h2>
                    <p>The verification token is missing or invalid.</p>
                    <a href="/" style="color: #2563eb;">Return to OllamaMax</a>
                </body></html>
            `);
        }

        const result = await auth.verifyEmail(token);
        
        res.send(`
            <html><body style="font-family: Arial, sans-serif; text-align: center; padding: 50px;">
                <div style="max-width: 600px; margin: 0 auto;">
                    <h1 style="color: #2563eb;">🦙 OllamaMax</h1>
                    <h2 style="color: #059669;">✅ Email Verified Successfully!</h2>
                    <p style="color: #4b5563; line-height: 1.6;">
                        Welcome to OllamaMax! Your account (${result.email}) has been verified.<br>
                        You can now log in to access the distributed AI platform.
                    </p>
                    <a href="/" style="background: #2563eb; color: white; padding: 15px 30px; text-decoration: none; border-radius: 5px; font-weight: bold; display: inline-block; margin-top: 20px;">
                        Go to OllamaMax Dashboard
                    </a>
                </div>
            </body></html>
        `);

    } catch (error) {
        console.error('Email verification error:', error.message);
        res.status(400).send(`
            <html><body style="font-family: Arial, sans-serif; text-align: center; padding: 50px;">
                <h2 style="color: #dc2626;">❌ Verification Failed</h2>
                <p>${error.message}</p>
                <a href="/" style="color: #2563eb;">Return to OllamaMax</a>
            </body></html>
        `);
    }
});

// Get Current User
app.get('/api/auth/user', authenticateToken, (req, res) => {
    res.json({
        success: true,
        user: req.user
    });
});

// Logout
app.post('/api/auth/logout', authenticateToken, (req, res) => {
    // In a more robust implementation, you'd invalidate the token in the database
    res.json({
        success: true,
        message: 'Logged out successfully'
    });
});

// Admin route to list all users (for testing)
app.get('/api/auth/users', async (req, res) => {
    try {
        const users = await auth.getAllUsers();
        res.json({
            success: true,
            users: users
        });
    } catch (error) {
        res.status(500).json({
            success: false,
            message: error.message
        });
    }
});

// Test email route
app.post('/api/auth/test-email', async (req, res) => {
    try {
        const { email } = req.body;
        if (!email) {
            return res.status(400).json({
                success: false,
                message: 'Email address required'
            });
        }

        await auth.sendTestEmail(email);
        
        res.json({
            success: true,
            message: `Test email sent to ${email}`
        });
    } catch (error) {
        console.error('Test email error:', error.message);
        res.status(500).json({
            success: false,
            message: error.message
        });
    }
});

// =================== END AUTHENTICATION ROUTES ===================

// =================== OPENAPI DOCUMENTATION ROUTES ===================

// Serve OpenAPI specification
app.get('/openapi.json', (req, res) => {
  res.setHeader('Content-Type', 'application/json');
  res.send(JSON.stringify(openapiSpec, null, 2));
});

// Serve Swagger UI documentation
app.use('/docs', swaggerUi.serve, swaggerUi.setup(openapiSpec, {
  customCss: `
    .swagger-ui .topbar { display: none }
    .swagger-ui .info .title { color: #2563eb }
  `,
  customSiteTitle: "OllamaMax API Documentation",
  swaggerOptions: {
    persistAuthorization: true,
    displayRequestDuration: true,
    docExpansion: 'list',
    filter: true,
    showExtensions: true,
    showCommonExtensions: true
  }
}));

// =================== END OPENAPI DOCUMENTATION ROUTES ===================

// Create WebSocket server with noServer option for path gating
const wss = new WebSocket.Server({ noServer: true });

// WebSocket upgrade handler with path gating (SECURITY: only allow /chat path)
server.on('upgrade', (request, socket, head) => {
    const { pathname } = new URL(request.url, `http://${request.headers.host}`);

    // Only allow WebSocket connections on /chat path
    if (pathname === '/chat') {
        wss.handleUpgrade(request, socket, head, (ws) => {
            wss.emit('connection', ws, request);
        });
    } else {
        console.log(`WebSocket connection rejected: invalid path ${pathname}`);
        socket.destroy();
    }
});

// WebSocket connection handler
wss.on('connection', (ws, request) => {
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
    
    // Check if request should use OpenRouter (Sonoma Sky Alpha)
    if (data.model === 'sonoma-sky-alpha' && openRouterClient) {
        return handleOpenRouterInference(ws, data, startTime);
    }
    
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
        uptime: process.uptime(),
        openrouter: {
            enabled: !!openRouterClient,
            models: openRouterClient ? ['alpindale/sonoma-sky-alpha'] : []
        }
    });
});

// OpenRouter-specific endpoints
app.get('/api/openrouter/models', async (req, res) => {
    if (!openRouterClient) {
        return res.status(503).json({ error: 'OpenRouter not configured' });
    }
    
    try {
        const models = await openRouterClient.getModels();
        res.json({ models });
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

app.post('/api/openrouter/chat', async (req, res) => {
    if (!openRouterClient) {
        return res.status(503).json({ error: 'OpenRouter not configured' });
    }
    
    try {
        const response = await openRouterClient.chatCompletion(req.body);
        res.json(response);
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
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

// OpenRouter inference handler
async function handleOpenRouterInference(ws, data, startTime) {
    try {
        console.log('Routing request to OpenRouter Sonoma Sky Alpha model');
        
        // Prepare OpenRouter request
        const openRouterRequest = {
            model: 'alpindale/sonoma-sky-alpha',
            messages: [
                { role: 'user', content: data.content }
            ],
            max_tokens: data.settings.maxTokens || 4096,
            temperature: data.settings.temperature || 0.7,
            top_p: data.settings.topP || 0.9,
            stream: data.settings.streaming || false
        };
        
        // Send initial response
        ws.send(JSON.stringify({
            type: 'response',
            id: data.timestamp,
            node: 'openrouter-sonoma-sky-alpha',
            streaming: data.settings.streaming
        }));
        
        // Store in Redis for distributed tracking
        await redis.set(`request:${data.timestamp}`, JSON.stringify({
            node: 'openrouter-sonoma-sky-alpha',
            model: 'alpindale/sonoma-sky-alpha',
            startTime
        }));
        
        if (data.settings.streaming) {
            // Note: OpenRouter streaming would require different implementation
            // For now, treat as complete response
            const response = await openRouterClient.chatCompletion(openRouterRequest);
            
            ws.send(JSON.stringify({
                type: 'stream_chunk',
                id: data.timestamp,
                chunk: response.choices[0].message.content,
                done: true
            }));
        } else {
            const response = await openRouterClient.chatCompletion(openRouterRequest);
            
            ws.send(JSON.stringify({
                type: 'response',
                id: data.timestamp,
                content: response.choices[0].message.content,
                node: 'openrouter-sonoma-sky-alpha',
                streaming: false,
                usage: response.usage
            }));
        }
        
        // Update metrics
        const latency = Date.now() - startTime;
        ws.send(JSON.stringify({
            type: 'metrics',
            latency,
            node: 'openrouter-sonoma-sky-alpha',
            model: 'alpindale/sonoma-sky-alpha'
        }));
        
    } catch (error) {
        console.error('OpenRouter inference error:', error);
        ws.send(JSON.stringify({
            type: 'error',
            message: `OpenRouter inference failed: ${error.message}`
        }));
    }
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