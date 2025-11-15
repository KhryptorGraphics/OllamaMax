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

// Import compression middleware
const compression = require('compression');
const zlib = require('zlib');

// Configure compression middleware with Brotli and Gzip support
app.use(compression({
  // Enable compression for all requests
  filter: (req, res) => {
    // Don't compress responses with this request header
    if (req.headers['x-no-compression']) {
      return false;
    }

    // Compress all other responses
    return compression.filter(req, res);
  },

  // Compression level (1-9, higher = better compression but slower)
  level: 6,

  // Minimum response size to compress (in bytes)
  threshold: 1024,

  // Memory level for compression (1-9)
  memLevel: 8,

  // Window size for compression
  windowBits: 15,

  // Compression strategy
  strategy: zlib.constants.Z_DEFAULT_STRATEGY
}));

// Request size limits and security middleware
app.use(express.json({
  limit: '10mb',  // Limit JSON payload size
  verify: (req, res, buf, encoding) => {
    // Log large requests for monitoring
    if (buf.length > 1024 * 1024) { // 1MB
      console.warn(`Large JSON request: ${buf.length} bytes from ${req.ip}`);
    }
  }
}));

app.use(express.urlencoded({
  extended: true,
  limit: '10mb',
  verify: (req, res, buf, encoding) => {
    if (buf.length > 1024 * 1024) {
      console.warn(`Large URL-encoded request: ${buf.length} bytes from ${req.ip}`);
    }
  }
}));

// CORS configuration with security considerations
app.use(cors({
  origin: process.env.NODE_ENV === 'production'
    ? ['https://ollamamax.com', 'https://api.ollamamax.com']
    : true,
  credentials: true,
  optionsSuccessStatus: 200,
  maxAge: 86400 // 24 hours
}));

// Configuration
const PORT = process.env.PORT || 13101;
const REDIS_HOST = process.env.REDIS_HOST || 'localhost';
const REDIS_PORT = process.env.REDIS_PORT || 6379;
const REDIS_PASSWORD = process.env.REDIS_PASSWORD || 'ollama_redis_pass';

// Request size limits configuration
const REQUEST_LIMITS = {
  JSON_LIMIT: process.env.MAX_JSON_SIZE || '10mb',
  URLENCODED_LIMIT: process.env.MAX_URLENCODED_SIZE || '10mb',
  WEBSOCKET_MESSAGE_LIMIT: parseInt(process.env.MAX_WS_MESSAGE_SIZE) || 1024 * 1024, // 1MB
  HTTP_BODY_LIMIT: parseInt(process.env.MAX_HTTP_BODY_SIZE) || 10 * 1024 * 1024, // 10MB
  INFERENCE_PROMPT_LIMIT: parseInt(process.env.MAX_PROMPT_SIZE) || 100 * 1024 // 100KB
};

// Request size validation middleware
const validateRequestSize = (req, res, next) => {
  const contentLength = parseInt(req.get('content-length') || '0');

  if (contentLength > REQUEST_LIMITS.HTTP_BODY_LIMIT) {
    console.warn(`Request too large: ${contentLength} bytes from ${req.ip} to ${req.path}`);
    return res.status(413).json({
      error: 'Request Entity Too Large',
      message: `Request body size (${contentLength} bytes) exceeds maximum allowed size (${REQUEST_LIMITS.HTTP_BODY_LIMIT} bytes)`,
      code: 'REQUEST_TOO_LARGE',
      maxSize: REQUEST_LIMITS.HTTP_BODY_LIMIT
    });
  }

  next();
};

// Add request size validation middleware
app.use(validateRequestSize);

// Error handling for request size limits
app.use((error, req, res, next) => {
  if (error.type === 'entity.too.large') {
    console.warn(`Request entity too large from ${req.ip} to ${req.path}: ${error.message}`);
    return res.status(413).json({
      error: 'Request Entity Too Large',
      message: 'Request body exceeds maximum allowed size',
      code: 'REQUEST_TOO_LARGE',
      details: error.message
    });
  }

  if (error.type === 'entity.parse.failed') {
    console.warn(`Request parse failed from ${req.ip} to ${req.path}: ${error.message}`);
    return res.status(400).json({
      error: 'Bad Request',
      message: 'Failed to parse request body',
      code: 'PARSE_ERROR',
      details: error.message
    });
  }

  next(error);
});

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

// Import health and metrics services
const HealthMonitor = require('../src/services/health-monitor');
const MetricsCollector = require('../src/services/metrics-collector');

// Initialize health monitor and metrics collector
const healthMonitor = new HealthMonitor(nodeRegistry, {
    checkInterval: 30000,  // 30 seconds
    healthTimeout: 5000,   // 5 seconds
    maxRetries: 3
});

const metricsCollector = new MetricsCollector({
    collectionInterval: 15000,  // 15 seconds
    retentionPeriod: 24 * 60 * 60 * 1000,  // 24 hours
    maxDataPoints: 1000
});

// Set up event listeners for health and metrics
healthMonitor.on('healthUpdate', (data) => {
    // Broadcast health updates to WebSocket clients
    const update = JSON.stringify({
        type: 'health_update',
        nodeId: data.nodeId,
        status: data.status,
        metrics: data.metrics,
        latency: data.latency
    });

    wss.clients.forEach(client => {
        if (client.readyState === WebSocket.OPEN) {
            client.send(update);
        }
    });

    // Record metrics
    metricsCollector.recordNodeMetrics({
        nodeId: data.nodeId,
        status: data.status,
        cpu: data.metrics.cpu,
        memory: data.metrics.memory,
        load: data.metrics.load,
        latency: data.latency
    });
});

healthMonitor.on('clusterHealth', (data) => {
    // Broadcast cluster health summary
    const update = JSON.stringify({
        type: 'cluster_health',
        ...data
    });

    wss.clients.forEach(client => {
        if (client.readyState === WebSocket.OPEN) {
            client.send(update);
        }
    });
});

metricsCollector.on('metricsCollected', (data) => {
    // Broadcast metrics to interested clients
    const update = JSON.stringify({
        type: 'metrics_update',
        ...data
    });

    wss.clients.forEach(client => {
        if (client.readyState === WebSocket.OPEN) {
            client.send(update);
        }
    });
});

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
            // Validate WebSocket message size
            if (message.length > REQUEST_LIMITS.WEBSOCKET_MESSAGE_LIMIT) {
                console.warn(`WebSocket message too large: ${message.length} bytes from ${ws._socket.remoteAddress}`);
                ws.send(JSON.stringify({
                    type: 'error',
                    error: 'Message Too Large',
                    message: `WebSocket message size (${message.length} bytes) exceeds maximum allowed size (${REQUEST_LIMITS.WEBSOCKET_MESSAGE_LIMIT} bytes)`,
                    code: 'MESSAGE_TOO_LARGE',
                    maxSize: REQUEST_LIMITS.WEBSOCKET_MESSAGE_LIMIT
                }));
                return;
            }

            const data = JSON.parse(message);

            // Validate inference prompt size if applicable
            if (data.type === 'inference' && data.content) {
                const promptSize = Buffer.byteLength(data.content, 'utf8');
                if (promptSize > REQUEST_LIMITS.INFERENCE_PROMPT_LIMIT) {
                    console.warn(`Inference prompt too large: ${promptSize} bytes from ${ws._socket.remoteAddress}`);
                    ws.send(JSON.stringify({
                        type: 'error',
                        error: 'Prompt Too Large',
                        message: `Inference prompt size (${promptSize} bytes) exceeds maximum allowed size (${REQUEST_LIMITS.INFERENCE_PROMPT_LIMIT} bytes)`,
                        code: 'PROMPT_TOO_LARGE',
                        maxSize: REQUEST_LIMITS.INFERENCE_PROMPT_LIMIT
                    }));
                    return;
                }
            }
            
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

    // Prepare routing options
    const routingOptions = {
      strategy: data.loadBalancing || data.routingMode || 'round-robin',
      sessionId: data.sessionId,
      modelName: data.model,
      taskType: data.taskType || 'inference',
      requiredCpu: data.requiredCpu,
      requiredMemory: data.requiredMemory,
      requiredGpu: data.requiredGpu,
      preferredRegion: data.preferredRegion,
      preferredZone: data.preferredZone,
      preferredNodeId: data.preferredNodeId,
      maxNodes: data.maxNodes,
      minNodes: data.minNodes
    };

    // Select node(s) using configured strategy
    const selectedNodes = nodeRegistry.selectNodeWithOptions(routingOptions);

    if (!selectedNodes || (Array.isArray(selectedNodes) && selectedNodes.length === 0)) {
        ws.send(JSON.stringify({
            type: 'error',
            message: 'No healthy nodes available'
        }));
        return;
    }

    // Handle broadcast mode (multiple nodes)
    if (Array.isArray(selectedNodes) && routingOptions.strategy === 'broadcast') {
        return handleBroadcastInference(ws, selectedNodes, data, startTime);
    }

    // Single node processing
    const node = Array.isArray(selectedNodes) ? selectedNodes[0] : selectedNodes;

    console.log(`Routing request to node: ${node.name} (strategy: ${routingOptions.strategy})`);

    // Store in Redis for distributed tracking
    await redis.set(`request:${data.timestamp}`, JSON.stringify({
        node: node.name,
        model: data.model,
        strategy: routingOptions.strategy,
        sessionId: data.sessionId,
        startTime
    }));

    // Send initial response
    ws.send(JSON.stringify({
        type: 'response',
        id: data.timestamp,
        node: node.name,
        strategy: routingOptions.strategy,
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

    // Record metrics in collector
    metricsCollector.recordRequest({
        modelName: data.model,
        nodeId: node.id,
        success: true,
        latency: latency,
        taskType: 'inference'
    });

    // Record in health monitor
    healthMonitor.recordRequest(true, latency);

    ws.send(JSON.stringify({
        type: 'metrics',
        latency,
        node: node.name,
        strategy: routingOptions.strategy
    }));
}

// Broadcast inference handling (parallel processing across multiple nodes)
async function handleBroadcastInference(ws, nodes, data, startTime) {
    console.log(`Broadcasting request to ${nodes.length} nodes: ${nodes.map(n => n.name).join(', ')}`);

    // Send initial response indicating broadcast mode
    ws.send(JSON.stringify({
        type: 'response',
        id: data.timestamp,
        nodes: nodes.map(n => n.name),
        strategy: 'broadcast',
        streaming: data.settings.streaming
    }));

    const promises = nodes.map(async (node, index) => {
        try {
            const nodeStartTime = Date.now();

            // Store individual node request in Redis
            await redis.set(`request:${data.timestamp}:node:${node.id}`, JSON.stringify({
                node: node.name,
                model: data.model,
                strategy: 'broadcast',
                nodeIndex: index,
                startTime: nodeStartTime
            }));

            let result;
            if (data.settings.streaming) {
                // For streaming, we'll collect the full response from each node
                result = await collectStreamingResponse(node, data);
            } else {
                result = await getSingleResponse(node, data);
            }

            const nodeLatency = Date.now() - nodeStartTime;

            // Send individual node result
            ws.send(JSON.stringify({
                type: 'broadcast_result',
                id: data.timestamp,
                nodeId: node.id,
                nodeName: node.name,
                nodeIndex: index,
                content: result.content,
                latency: nodeLatency,
                success: true
            }));

            return { node, result, latency: nodeLatency, success: true };

        } catch (error) {
            console.error(`Broadcast inference failed for node ${node.name}:`, error);

            ws.send(JSON.stringify({
                type: 'broadcast_result',
                id: data.timestamp,
                nodeId: node.id,
                nodeName: node.name,
                nodeIndex: index,
                error: error.message,
                success: false
            }));

            return { node, error: error.message, success: false };
        }
    });

    // Wait for all nodes to complete
    const results = await Promise.allSettled(promises);
    const totalLatency = Date.now() - startTime;

    // Send final broadcast summary
    const successfulResults = results.filter(r => r.status === 'fulfilled' && r.value.success);
    const failedResults = results.filter(r => r.status === 'rejected' || !r.value.success);

    ws.send(JSON.stringify({
        type: 'broadcast_complete',
        id: data.timestamp,
        totalNodes: nodes.length,
        successfulNodes: successfulResults.length,
        failedNodes: failedResults.length,
        totalLatency,
        averageLatency: successfulResults.length > 0
            ? successfulResults.reduce((sum, r) => sum + r.value.latency, 0) / successfulResults.length
            : 0
    }));
}

// Helper function to collect streaming response from a single node
async function collectStreamingResponse(node, data) {
    return new Promise((resolve, reject) => {
        const chunks = [];

        fetch(`${node.url}/api/generate`, {
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
        })
        .then(response => {
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}`);
            }

            const reader = response.body.getReader();
            const decoder = new TextDecoder();

            function readChunk() {
                reader.read().then(({ done, value }) => {
                    if (done) {
                        const fullContent = chunks.join('');
                        resolve({ content: fullContent });
                        return;
                    }

                    const chunk = decoder.decode(value);
                    const lines = chunk.split('\n').filter(line => line.trim());

                    for (const line of lines) {
                        try {
                            const parsed = JSON.parse(line);
                            if (parsed.response) {
                                chunks.push(parsed.response);
                            }
                        } catch (e) {
                            // Ignore parsing errors for individual chunks
                        }
                    }

                    readChunk();
                }).catch(reject);
            }

            readChunk();
        })
        .catch(reject);
    });
}

// Helper function to get single response from a node
async function getSingleResponse(node, data) {
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

    if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
    }

    const result = await response.json();
    return { content: result.response };
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

// Routing configuration endpoints
app.get('/api/routing/stats', (req, res) => {
    const stats = nodeRegistry.getRoutingStats();
    res.json(stats);
});

app.post('/api/routing/clear-affinity', (req, res) => {
    const { sessionId } = req.body;
    nodeRegistry.clearSessionAffinity(sessionId);

    res.json({
        success: true,
        message: sessionId ? `Cleared affinity for session ${sessionId}` : 'Cleared all session affinities'
    });
});

app.get('/api/routing/modes', (req, res) => {
    res.json({
        modes: [
            {
                id: 'round-robin',
                name: 'Round Robin',
                description: 'Distributes requests evenly across all healthy nodes',
                useCase: 'General load distribution'
            },
            {
                id: 'least-loaded',
                name: 'Least Loaded',
                description: 'Routes to the node with the lowest current load',
                useCase: 'Optimal resource utilization'
            },
            {
                id: 'fastest',
                name: 'Fastest Response',
                description: 'Routes to the node with the best response time',
                useCase: 'Latency-sensitive applications'
            },
            {
                id: 'auto',
                name: 'Auto (Intelligent)',
                description: 'Automatically selects the best routing strategy based on context',
                useCase: 'Adaptive routing for mixed workloads'
            },
            {
                id: 'single-node',
                name: 'Single Node',
                description: 'Routes all requests to a single best node',
                useCase: 'Consistency and resource consolidation'
            },
            {
                id: 'pinned',
                name: 'Pinned (Session Affinity)',
                description: 'Maintains session affinity by routing to the same node',
                useCase: 'Stateful conversations and consistency'
            },
            {
                id: 'broadcast',
                name: 'Broadcast (Parallel)',
                description: 'Sends requests to multiple nodes for parallel processing',
                useCase: 'Consensus, comparison, and redundancy'
            },
            {
                id: 'geographic',
                name: 'Geographic',
                description: 'Routes based on geographic proximity and region preferences',
                useCase: 'Latency optimization and data locality'
            },
            {
                id: 'resource-aware',
                name: 'Resource Aware',
                description: 'Routes based on specific resource requirements (CPU, memory, GPU)',
                useCase: 'Resource-intensive tasks and specialized workloads'
            }
        ]
    });
});

app.post('/api/routing/test', async (req, res) => {
    const { strategy, options = {} } = req.body;

    try {
        const selectedNodes = nodeRegistry.selectNodeWithOptions({
            strategy,
            ...options
        });

        if (!selectedNodes) {
            return res.status(404).json({ error: 'No nodes available' });
        }

        const result = Array.isArray(selectedNodes) ? selectedNodes : [selectedNodes];

        res.json({
            strategy,
            options,
            selectedNodes: result.map(node => ({
                id: node.id,
                name: node.name,
                status: node.status,
                load: node.health.load,
                cpu: node.health.cpu,
                memory: node.health.memory
            }))
        });
    } catch (error) {
        res.status(400).json({ error: error.message });
    }
});

// Model management endpoints
app.get('/api/models', (req, res) => {
    // Get models from all nodes
    const allNodes = nodeRegistry.getAllNodes();
    const modelMap = new Map();

    allNodes.forEach(node => {
        if (node.modelsLoaded && Array.isArray(node.modelsLoaded)) {
            node.modelsLoaded.forEach(modelName => {
                if (!modelMap.has(modelName)) {
                    modelMap.set(modelName, {
                        name: modelName,
                        nodes: [],
                        replicas: 0,
                        status: 'available'
                    });
                }

                const model = modelMap.get(modelName);
                model.nodes.push({
                    id: node.id,
                    name: node.name,
                    status: node.status,
                    load: node.load || 0
                });
                model.replicas++;
            });
        }
    });

    const models = Array.from(modelMap.values());

    res.json({
        models,
        totalModels: models.length,
        totalReplicas: models.reduce((sum, model) => sum + model.replicas, 0),
        healthyNodes: allNodes.filter(n => n.status === 'healthy').length,
        totalNodes: allNodes.length
    });
});

app.post('/api/models/:modelName/download', async (req, res) => {
    const { modelName } = req.params;
    const { targetNodeId, sourceNodeId } = req.body;

    try {
        // Find source node if not specified
        let sourceNode = null;
        if (sourceNodeId) {
            sourceNode = nodeRegistry.getNodes().find(n => n.id === sourceNodeId);
        } else {
            // Find a healthy node that has the model
            const nodesWithModel = nodeRegistry.getNodes().filter(node =>
                node.status === 'healthy' &&
                node.health.modelsLoaded &&
                node.health.modelsLoaded.includes(modelName)
            );
            sourceNode = nodesWithModel[0];
        }

        if (!sourceNode) {
            return res.status(404).json({
                error: `Model ${modelName} not found on any healthy nodes`
            });
        }

        // Find target node
        const targetNode = targetNodeId
            ? nodeRegistry.getNodes().find(n => n.id === targetNodeId)
            : nodeRegistry.selectNode('least-loaded');

        if (!targetNode) {
            return res.status(404).json({
                error: 'No suitable target node available'
            });
        }

        // Simulate model download/replication
        const downloadId = `download_${modelName}_${Date.now()}`;

        // Store download status in Redis
        await redis.set(`download:${downloadId}`, JSON.stringify({
            id: downloadId,
            modelName,
            sourceNode: sourceNode.name,
            targetNode: targetNode.name,
            status: 'in_progress',
            progress: 0,
            startTime: Date.now()
        }));

        // Simulate async download process
        setTimeout(async () => {
            try {
                // Update progress
                for (let progress = 10; progress <= 100; progress += 10) {
                    await redis.set(`download:${downloadId}`, JSON.stringify({
                        id: downloadId,
                        modelName,
                        sourceNode: sourceNode.name,
                        targetNode: targetNode.name,
                        status: 'in_progress',
                        progress,
                        startTime: Date.now()
                    }));

                    if (progress < 100) {
                        await new Promise(resolve => setTimeout(resolve, 1000));
                    }
                }

                // Mark as completed
                await redis.set(`download:${downloadId}`, JSON.stringify({
                    id: downloadId,
                    modelName,
                    sourceNode: sourceNode.name,
                    targetNode: targetNode.name,
                    status: 'completed',
                    progress: 100,
                    startTime: Date.now(),
                    completedAt: Date.now()
                }));

                // Add model to target node's loaded models
                if (!targetNode.health.modelsLoaded) {
                    targetNode.health.modelsLoaded = [];
                }
                if (!targetNode.health.modelsLoaded.includes(modelName)) {
                    targetNode.health.modelsLoaded.push(modelName);
                }

                console.log(`Model ${modelName} successfully downloaded to ${targetNode.name}`);

            } catch (error) {
                console.error('Download simulation error:', error);
                await redis.set(`download:${downloadId}`, JSON.stringify({
                    id: downloadId,
                    modelName,
                    sourceNode: sourceNode.name,
                    targetNode: targetNode.name,
                    status: 'failed',
                    error: error.message,
                    startTime: Date.now()
                }));
            }
        }, 100);

        res.json({
            downloadId,
            modelName,
            sourceNode: sourceNode.name,
            targetNode: targetNode.name,
            status: 'initiated'
        });

    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

app.get('/api/models/downloads/:downloadId', async (req, res) => {
    const { downloadId } = req.params;

    try {
        const downloadData = await redis.get(`download:${downloadId}`);
        if (!downloadData) {
            return res.status(404).json({ error: 'Download not found' });
        }

        res.json(JSON.parse(downloadData));
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

app.delete('/api/models/:modelName', async (req, res) => {
    const { modelName } = req.params;
    const { nodeId } = req.query;

    try {
        let deletedFrom = [];

        if (nodeId) {
            // Delete from specific node
            const node = nodeRegistry.getNodes().find(n => n.id === nodeId);
            if (node && node.health.modelsLoaded) {
                const index = node.health.modelsLoaded.indexOf(modelName);
                if (index > -1) {
                    node.health.modelsLoaded.splice(index, 1);
                    deletedFrom.push(node.name);
                }
            }
        } else {
            // Delete from all nodes
            nodeRegistry.getNodes().forEach(node => {
                if (node.health.modelsLoaded) {
                    const index = node.health.modelsLoaded.indexOf(modelName);
                    if (index > -1) {
                        node.health.modelsLoaded.splice(index, 1);
                        deletedFrom.push(node.name);
                    }
                }
            });
        }

        res.json({
            modelName,
            deletedFrom,
            success: deletedFrom.length > 0
        });

    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

app.get('/api/health', (req, res) => {
    const clusterHealth = healthMonitor.getClusterHealthSummary();
    const systemMetrics = healthMonitor.getSystemMetrics();

    res.json({
        status: clusterHealth.healthyNodes > 0 ? 'healthy' : 'unhealthy',
        cluster: clusterHealth,
        system: systemMetrics,
        queueLength: messageQueue.getLength(),
        uptime: process.uptime(),
        openrouter: {
            enabled: !!openRouterClient,
            models: openRouterClient ? ['alpindale/sonoma-sky-alpha'] : []
        }
    });
});

// Enhanced health endpoints
app.get('/api/health/detailed', (req, res) => {
    const clusterHealth = healthMonitor.getClusterHealthSummary();
    const systemMetrics = healthMonitor.getSystemMetrics();
    const metricsSnapshot = metricsCollector.getSnapshot();

    res.json({
        timestamp: Date.now(),
        cluster: clusterHealth,
        system: systemMetrics,
        metrics: metricsSnapshot,
        services: {
            healthMonitor: healthMonitor.isRunning,
            metricsCollector: metricsCollector.isCollecting,
            redis: redis.status === 'ready',
            messageQueue: messageQueue.getLength()
        }
    });
});

app.get('/api/health/nodes/:nodeId', (req, res) => {
    const { nodeId } = req.params;
    const { limit = 50 } = req.query;

    const healthHistory = healthMonitor.getNodeHealthHistory(nodeId, parseInt(limit));

    if (healthHistory.length === 0) {
        return res.status(404).json({ error: 'Node not found or no health data available' });
    }

    res.json({
        nodeId,
        historyCount: healthHistory.length,
        history: healthHistory
    });
});

// Metrics endpoints
app.get('/api/metrics', (req, res) => {
    const snapshot = metricsCollector.getSnapshot();
    const latencyStats = metricsCollector.getLatencyStats();

    res.json({
        ...snapshot,
        latency: latencyStats
    });
});

app.get('/api/metrics/prometheus', (req, res) => {
    res.set('Content-Type', 'text/plain; version=0.0.4; charset=utf-8');

    // Combine metrics from both health monitor and metrics collector
    const healthMetrics = healthMonitor.getPrometheusMetrics();
    const systemMetrics = metricsCollector.getPrometheusMetrics();

    res.send(`${healthMetrics}\n\n${systemMetrics}`);
});

app.get('/api/metrics/timeseries/:metric', (req, res) => {
    const { metric } = req.params;
    const { timeRange = 3600000 } = req.query; // Default 1 hour

    const data = metricsCollector.getTimeSeriesData(metric, parseInt(timeRange));

    res.json({
        metric,
        timeRange: parseInt(timeRange),
        dataPoints: data.length,
        data
    });
});

// Start health monitoring and metrics collection
app.get('/api/monitoring/start', (req, res) => {
    try {
        if (!healthMonitor.isRunning) {
            healthMonitor.start();
        }
        if (!metricsCollector.isCollecting) {
            metricsCollector.start();
        }

        res.json({
            success: true,
            message: 'Monitoring services started',
            services: {
                healthMonitor: healthMonitor.isRunning,
                metricsCollector: metricsCollector.isCollecting
            }
        });
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

app.get('/api/monitoring/stop', (req, res) => {
    try {
        healthMonitor.stop();
        metricsCollector.stop();

        res.json({
            success: true,
            message: 'Monitoring services stopped',
            services: {
                healthMonitor: healthMonitor.isRunning,
                metricsCollector: metricsCollector.isCollecting
            }
        });
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
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
    console.log(`Health endpoint: http://localhost:${PORT}/api/health`);
    console.log(`Metrics endpoint: http://localhost:${PORT}/api/metrics`);
    console.log(`Prometheus metrics: http://localhost:${PORT}/api/metrics/prometheus`);

    // Initialize default nodes
    initializeDefaultNodes();

    // Connect to Redis
    redis.on('connect', () => {
        console.log('Connected to Redis for distributed state management');
    });

    redis.on('error', (error) => {
        console.error('Redis connection error:', error);
    });

    // Start monitoring services after a short delay
    setTimeout(() => {
        console.log('Starting health monitoring and metrics collection...');
        try {
            healthMonitor.start();
            metricsCollector.start();
            console.log('✓ Health monitoring started');
            console.log('✓ Metrics collection started');
        } catch (error) {
            console.error('Failed to start monitoring services:', error);
        }
    }, 3000); // Wait 3 seconds for nodes to initialize
});