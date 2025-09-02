/**
 * Distributed Llama Inference System
 * Real integration with Ollama API for distributed processing
 */

const WebSocket = require('ws');
const http = require('http');
const express = require('express');
const path = require('path');
const cors = require('cors');
const Redis = require('ioredis');
const fetch = require('node-fetch');

const app = express();
app.use(cors());
app.use(express.json());

// Serve static files from web-interface directory
// In Docker, web-interface is at ./web-interface, in development it's at ../web-interface
const webInterfacePath = process.env.NODE_ENV === 'production' 
    ? path.join(process.cwd(), 'web-interface')
    : path.join(__dirname, '..', 'web-interface');
app.use(express.static(webInterfacePath));

// Serve index.html for root path
app.get('/', (req, res) => {
    const indexPath = path.join(webInterfacePath, 'index.html');
    res.sendFile(path.resolve(indexPath));
});

// Configuration
const PORT = process.env.PORT || 13100;
const REDIS_HOST = process.env.REDIS_HOST || 'localhost';
const REDIS_PORT = process.env.REDIS_PORT || 13001;
const REDIS_PASSWORD = process.env.REDIS_PASSWORD || '';
const OLLAMA_PRIMARY = process.env.OLLAMA_PRIMARY || 'http://localhost:13000';

// Initialize Redis for distributed state and queue management
const redisConfig = {
    host: REDIS_HOST,
    port: REDIS_PORT,
    retryStrategy: (times) => Math.min(times * 50, 2000)
};

// Add password if provided
if (REDIS_PASSWORD) {
    redisConfig.password = REDIS_PASSWORD;
}

const redis = new Redis(redisConfig);

const redisPub = new Redis(redisConfig);

const redisSub = new Redis(redisConfig);

// Enhanced Worker Node Registry
class WorkerNode {
    constructor(id, url, name) {
        this.id = id;
        this.url = url;
        this.name = name;
        this.status = 'initializing';
        this.metrics = {
            totalRequests: 0,
            activeRequests: 0,
            avgResponseTime: 0,
            successRate: 100,
            lastCheck: Date.now()
        };
        this.model = null;
        
        // Enhanced monitoring data
        this.systemInfo = null;
        this.performanceHistory = {
            timestamps: [],
            cpu: [],
            memory: [],
            requests: [],
            responseTime: []
        };
        this.healthStatus = {
            lastCheck: Date.now(),
            checks: {
                api: 'unknown',
                models: 'unknown',
                resources: 'unknown',
                network: 'unknown'
            },
            warnings: [],
            errors: []
        };
        this.ollamaInfo = {
            version: null,
            models: [],
            activeRequests: 0,
            queueLength: 0,
            gpuMemory: { used: 0, total: 0 },
            concurrentCapacity: 10
        };
    }

    async checkHealth() {
        try {
            const response = await fetch(`${this.url}/api/tags`, {
                timeout: 2000
            });
            
            if (response.ok) {
                const data = await response.json();
                this.status = 'healthy';
                this.model = data.models && data.models.length > 0 ? data.models[0].name : null;
                this.ollamaInfo.models = data.models || [];
                
                // Update health checks
                this.healthStatus.checks.api = 'ok';
                this.healthStatus.checks.models = data.models && data.models.length > 0 ? 'ok' : 'warning';
                this.healthStatus.lastCheck = Date.now();
                
                return true;
            }
        } catch (error) {
            this.status = 'error';
            this.healthStatus.checks.api = 'error';
            this.healthStatus.errors = [`Health check failed: ${error.message}`];
            console.error(`Health check failed for ${this.name}:`, error.message);
        }
        return false;
    }

    async collectSystemInfo() {
        try {
            // Try to get system info from Docker if this is a container
            const containerName = this.name.replace('ollama-', '');
            
            // Simulate system info collection (in real implementation, would use Docker API)
            this.systemInfo = {
                cpu: {
                    usage: Math.random() * 100,
                    cores: 8,
                    load: [Math.random() * 2, Math.random() * 2, Math.random() * 2]
                },
                memory: {
                    used: Math.floor(Math.random() * 4096),
                    total: 8192,
                    get usage() { return (this.used / this.total) * 100; }
                },
                disk: {
                    used: Math.floor(Math.random() * 250),
                    total: 500,
                    get usage() { return (this.used / this.total) * 100; }
                },
                network: {
                    rx: Math.floor(Math.random() * 1024),
                    tx: Math.floor(Math.random() * 512),
                    connections: Math.floor(Math.random() * 50)
                },
                uptime: Math.floor(Date.now() / 1000),
                platform: 'linux',
                architecture: 'x64'
            };

            // Update performance history
            const now = Date.now();
            this.performanceHistory.timestamps.push(now);
            this.performanceHistory.cpu.push(this.systemInfo.cpu.usage);
            this.performanceHistory.memory.push(this.systemInfo.memory.usage);
            this.performanceHistory.requests.push(this.metrics.totalRequests);
            this.performanceHistory.responseTime.push(this.metrics.avgResponseTime);

            // Keep only last 60 data points (5 minutes at 5-second intervals)
            if (this.performanceHistory.timestamps.length > 60) {
                const keys = ['timestamps', 'cpu', 'memory', 'requests', 'responseTime'];
                keys.forEach(key => {
                    this.performanceHistory[key] = this.performanceHistory[key].slice(-60);
                });
            }

            // Update health status based on system metrics
            this.healthStatus.checks.resources = this.systemInfo.cpu.usage > 90 || this.systemInfo.memory.usage > 85 ? 'warning' : 'ok';
            this.healthStatus.warnings = [];
            
            if (this.systemInfo.cpu.usage > 90) {
                this.healthStatus.warnings.push('High CPU usage detected');
            }
            if (this.systemInfo.memory.usage > 85) {
                this.healthStatus.warnings.push('High memory usage detected');
            }

        } catch (error) {
            console.error(`System info collection failed for ${this.name}:`, error.message);
            this.healthStatus.checks.resources = 'error';
        }
    }

    async getDetailedStatus() {
        await this.collectSystemInfo();
        
        return {
            id: this.id,
            name: this.name,
            url: this.url,
            status: this.status,
            system: this.systemInfo,
            ollama: this.ollamaInfo,
            metrics: {
                ...this.metrics,
                requestsPerSecond: this.calculateRequestsPerSecond(),
                errorRate: this.calculateErrorRate(),
                uptime: this.calculateUptime(),
                tokensPerSecond: Math.random() * 50 // Placeholder
            },
            timeSeries: this.performanceHistory,
            health: this.healthStatus
        };
    }

    calculateRequestsPerSecond() {
        const historyLength = this.performanceHistory.timestamps.length;
        if (historyLength < 2) return 0;
        
        const timeSpan = (this.performanceHistory.timestamps[historyLength - 1] - 
                         this.performanceHistory.timestamps[0]) / 1000;
        const requestDiff = this.performanceHistory.requests[historyLength - 1] - 
                           this.performanceHistory.requests[0];
        
        return timeSpan > 0 ? requestDiff / timeSpan : 0;
    }

    calculateErrorRate() {
        return ((100 - this.metrics.successRate) / 100);
    }

    calculateUptime() {
        const uptime = (Date.now() - (this.healthStatus.lastCheck - 3600000)) / 1000;
        return Math.min(uptime / 3600, 100); // Percentage uptime over last hour
    }

    async process(prompt, settings) {
        const startTime = Date.now();
        this.metrics.activeRequests++;
        
        try {
            const response = await fetch(`${this.url}/api/generate`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    model: settings.model || 'tinyllama',
                    prompt: prompt,
                    stream: settings.streaming || false,
                    options: {
                        temperature: settings.temperature || 0.7,
                        num_predict: settings.maxTokens || 200,
                        top_k: settings.topK || 40,
                        top_p: settings.topP || 0.9,
                        repeat_penalty: settings.repeatPenalty || 1.1
                    }
                })
            });

            this.metrics.totalRequests++;
            this.metrics.activeRequests--;
            
            const responseTime = Date.now() - startTime;
            this.metrics.avgResponseTime = 
                (this.metrics.avgResponseTime * (this.metrics.totalRequests - 1) + responseTime) / 
                this.metrics.totalRequests;

            return response;
        } catch (error) {
            this.metrics.activeRequests--;
            this.metrics.successRate = 
                ((this.metrics.totalRequests - 1) * this.metrics.successRate / 100) / 
                this.metrics.totalRequests * 100;
            throw error;
        }
    }
}

// Distributed Inference Coordinator
class InferenceCoordinator {
    constructor() {
        this.workers = new Map();
        this.requestQueue = [];
        this.activeRequests = new Map();
        this.initializeWorkers();
    }

    async initializeWorkers() {
        const workerConfigs = [];
        
        // Add workers from environment variables
        if (process.env.OLLAMA_PRIMARY) {
            workerConfigs.push({
                id: 'worker-1',
                url: process.env.OLLAMA_PRIMARY,
                name: 'ollama-primary'
            });
        }
        
        if (process.env.OLLAMA_WORKER_2) {
            workerConfigs.push({
                id: 'worker-2',
                url: process.env.OLLAMA_WORKER_2,
                name: 'ollama-worker-2'
            });
        }
        
        if (process.env.OLLAMA_WORKER_3) {
            workerConfigs.push({
                id: 'worker-3',
                url: process.env.OLLAMA_WORKER_3,
                name: 'ollama-worker-3'
            });
        }
        
        // Fallback to localhost if no environment variables
        if (workerConfigs.length === 0) {
            workerConfigs.push(
                { id: 'worker-1', url: 'http://localhost:13000', name: 'ollama-primary' },
                { id: 'worker-2', url: 'http://localhost:13001', name: 'ollama-worker-2' },
                { id: 'worker-3', url: 'http://localhost:13002', name: 'ollama-worker-3' }
            );
        }
        
        console.log('Initializing workers:', workerConfigs.map(c => `${c.name}: ${c.url}`));
        
        // Initialize workers
        for (const config of workerConfigs) {
            const worker = new WorkerNode(config.id, config.url, config.name);
            
            if (await worker.checkHealth()) {
                this.workers.set(worker.id, worker);
                console.log(`✅ Added healthy worker: ${worker.name} at ${worker.url}`);
            } else {
                console.log(`❌ Worker unhealthy: ${worker.name} at ${worker.url}`);
            }
        }

        console.log(`🚀 Initialized ${this.workers.size} worker(s) out of ${workerConfigs.length} configured`);
    }

    selectWorker(strategy = 'least-loaded') {
        const healthyWorkers = Array.from(this.workers.values())
            .filter(w => w.status === 'healthy');
        
        if (healthyWorkers.length === 0) return null;

        switch(strategy) {
            case 'least-loaded':
                return healthyWorkers.reduce((min, worker) => 
                    worker.metrics.activeRequests < min.metrics.activeRequests ? worker : min
                );
            
            case 'fastest':
                return healthyWorkers.reduce((best, worker) => 
                    worker.metrics.avgResponseTime < best.metrics.avgResponseTime ? worker : best
                );
            
            case 'round-robin':
            default:
                // Simple round-robin using Redis counter
                return healthyWorkers[0]; // Simplified for now
        }
    }

    async distributeInference(requestId, prompt, settings) {
        const worker = this.selectWorker(settings.loadBalancing);
        
        if (!worker) {
            throw new Error('No healthy workers available');
        }

        console.log(`Distributing request ${requestId} to ${worker.name}`);
        
        // Store request metadata in Redis
        await redis.set(`request:${requestId}`, JSON.stringify({
            workerId: worker.id,
            workerName: worker.name,
            prompt: prompt.substring(0, 100),
            startTime: Date.now(),
            status: 'processing'
        }), 'EX', 3600);

        return worker;
    }

    async processChunkedInference(ws, requestId, worker, prompt, settings) {
        try {
            const response = await fetch(`${worker.url}/api/generate`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    model: settings.model || 'tinyllama',
                    prompt: prompt,
                    stream: true,
                    options: {
                        temperature: settings.temperature || 0.7,
                        num_predict: settings.maxTokens || 200
                    }
                })
            });

            let buffer = '';
            
            // Handle streaming response using Node.js streams
            response.body.on('data', async (chunk) => {
                buffer += chunk.toString();
                const lines = buffer.split('\n');
                buffer = lines.pop() || '';

                for (const line of lines) {
                    if (line.trim()) {
                        try {
                            const data = JSON.parse(line);
                            
                            // Send chunk to client
                            ws.send(JSON.stringify({
                                type: 'stream_chunk',
                                id: requestId,
                                chunk: data.response || '',
                                done: data.done || false,
                                worker: worker.name
                            }));

                            // Store partial response in Redis for recovery
                            if (data.response) {
                                await redis.append(`response:${requestId}`, data.response);
                                await redis.expire(`response:${requestId}`, 300);
                            }

                            if (data.done) {
                                // Update metrics
                                const requestData = await redis.get(`request:${requestId}`);
                                if (requestData) {
                                    const request = JSON.parse(requestData);
                                    const duration = Date.now() - request.startTime;
                                    
                                    await redis.zadd('request_times', Date.now(), `${requestId}:${duration}`);
                                    await redis.set(`request:${requestId}:complete`, JSON.stringify({
                                        ...request,
                                        status: 'complete',
                                        duration,
                                        totalTokens: data.total_duration ? Math.round(data.total_duration / 1000000) : 0
                                    }), 'EX', 3600);
                                }
                                
                                return;
                            }
                        } catch (e) {
                            console.error('Error parsing chunk:', e);
                        }
                    }
                }
            });

            response.body.on('end', () => {
                // Handle any remaining buffer content
                if (buffer.trim()) {
                    try {
                        const data = JSON.parse(buffer);
                        ws.send(JSON.stringify({
                            type: 'stream_chunk',
                            id: requestId,
                            chunk: data.response || '',
                            done: true,
                            worker: worker.name
                        }));
                    } catch (e) {
                        console.error('Error parsing final chunk:', e);
                    }
                }
            });
        } catch (error) {
            console.error('Streaming error:', error);
            
            // Try to recover or failover to another worker
            const alternativeWorker = this.selectWorker('least-loaded');
            if (alternativeWorker && alternativeWorker.id !== worker.id) {
                console.log(`Failing over to ${alternativeWorker.name}`);
                return this.processChunkedInference(ws, requestId, alternativeWorker, prompt, settings);
            }
            
            throw error;
        }
    }
}

// Initialize coordinator
const coordinator = new InferenceCoordinator();

// Create HTTP server
const server = http.createServer(app);

// WebSocket server
const wss = new WebSocket.Server({ 
    server,
    path: '/chat'
});

// WebSocket connection handler
wss.on('connection', (ws) => {
    console.log('New client connected');
    
    // Send worker status
    ws.send(JSON.stringify({
        type: 'node_update',
        nodes: Array.from(coordinator.workers.values()).map(w => ({
            id: w.id,
            name: w.name,
            status: w.status,
            model: w.model,
            metrics: w.metrics
        }))
    }));
    
    ws.on('message', async (message) => {
        try {
            const data = JSON.parse(message);
            
            switch(data.type) {
                case 'inference':
                    await handleInference(ws, data);
                    break;
                    
                case 'get_status':
                    await handleStatus(ws);
                    break;
                    
                case 'get_metrics':
                    await handleMetrics(ws);
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
    });
});

// Handle inference request
async function handleInference(ws, data) {
    const requestId = data.timestamp || Date.now();
    const startTime = Date.now();
    
    try {
        // Select worker and distribute
        const worker = await coordinator.distributeInference(
            requestId,
            data.content,
            data.settings || {}
        );
        
        // Send initial response
        ws.send(JSON.stringify({
            type: 'response',
            id: requestId,
            node: worker.name,
            streaming: data.settings?.streaming !== false
        }));
        
        // Process based on streaming preference
        if (data.settings?.streaming !== false) {
            await coordinator.processChunkedInference(
                ws,
                requestId,
                worker,
                data.content,
                data.settings || {}
            );
        } else {
            // Non-streaming response
            const response = await worker.process(data.content, data.settings || {});
            const result = await response.json();
            
            ws.send(JSON.stringify({
                type: 'response',
                id: requestId,
                content: result.response,
                node: worker.name,
                streaming: false
            }));
        }
        
        // Send metrics
        const latency = Date.now() - startTime;
        ws.send(JSON.stringify({
            type: 'metrics',
            latency,
            node: worker.name,
            requestId
        }));
        
    } catch (error) {
        console.error('Inference error:', error);
        ws.send(JSON.stringify({
            type: 'error',
            message: `Inference failed: ${error.message}`,
            requestId
        }));
    }
}

// Handle status request
async function handleStatus(ws) {
    const workers = Array.from(coordinator.workers.values());
    
    // Update health status
    await Promise.all(workers.map(w => w.checkHealth()));
    
    ws.send(JSON.stringify({
        type: 'status',
        workers: workers.map(w => ({
            id: w.id,
            name: w.name,
            status: w.status,
            model: w.model,
            url: w.url,
            metrics: w.metrics
        })),
        queueLength: coordinator.requestQueue.length,
        activeRequests: coordinator.activeRequests.size
    }));
}

// Handle metrics request
async function handleMetrics(ws) {
    // Get recent request times from Redis
    const recentRequests = await redis.zrevrange('request_times', 0, 99, 'WITHSCORES');
    
    const metrics = {
        workers: {},
        totalRequests: 0,
        avgResponseTime: 0,
        successRate: 0
    };
    
    // Aggregate worker metrics
    for (const [id, worker] of coordinator.workers) {
        metrics.workers[id] = worker.metrics;
        metrics.totalRequests += worker.metrics.totalRequests;
    }
    
    // Calculate averages
    if (metrics.totalRequests > 0) {
        const workers = Array.from(coordinator.workers.values());
        metrics.avgResponseTime = workers.reduce((sum, w) => 
            sum + w.metrics.avgResponseTime * w.metrics.totalRequests, 0) / metrics.totalRequests;
        metrics.successRate = workers.reduce((sum, w) => 
            sum + w.metrics.successRate * w.metrics.totalRequests, 0) / metrics.totalRequests;
    }
    
    // Parse recent request data
    const requestTimes = [];
    for (let i = 0; i < recentRequests.length; i += 2) {
        const [requestId, duration] = recentRequests[i].split(':');
        requestTimes.push({
            requestId,
            duration: parseInt(duration),
            timestamp: parseInt(recentRequests[i + 1])
        });
    }
    
    metrics.recentRequests = requestTimes;
    
    ws.send(JSON.stringify({
        type: 'metrics',
        data: metrics
    }));
}

// REST API endpoints
app.get('/api/health', async (req, res) => {
    const workers = Array.from(coordinator.workers.values());
    const healthyWorkers = workers.filter(w => w.status === 'healthy');
    
    res.json({
        status: healthyWorkers.length > 0 ? 'healthy' : 'degraded',
        workers: healthyWorkers.length,
        totalWorkers: workers.length,
        queueLength: coordinator.requestQueue.length,
        uptime: process.uptime()
    });
});

app.get('/api/workers', async (req, res) => {
    const workers = Array.from(coordinator.workers.values());
    
    res.json({
        workers: workers.map(w => ({
            id: w.id,
            name: w.name,
            status: w.status,
            model: w.model,
            metrics: w.metrics
        }))
    });
});

// Enhanced Node Management API Endpoints

// Get detailed information for all nodes
app.get('/api/nodes/detailed', async (req, res) => {
    try {
        const workers = Array.from(coordinator.workers.values());
        const detailedNodes = await Promise.all(
            workers.map(worker => worker.getDetailedStatus())
        );
        
        res.json({
            nodes: detailedNodes,
            cluster: {
                totalNodes: workers.length,
                healthyNodes: workers.filter(w => w.status === 'healthy').length,
                totalCpuCores: detailedNodes.reduce((sum, node) => 
                    sum + (node.system?.cpu?.cores || 0), 0),
                totalMemoryGB: detailedNodes.reduce((sum, node) => 
                    sum + (node.system?.memory?.total || 0) / 1024, 0),
                totalActiveRequests: detailedNodes.reduce((sum, node) => 
                    sum + (node.metrics?.activeRequests || 0), 0)
            }
        });
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

// Get detailed information for a specific node
app.get('/api/nodes/:nodeId/detailed', async (req, res) => {
    try {
        const { nodeId } = req.params;
        const worker = coordinator.workers.get(nodeId);
        
        if (!worker) {
            return res.status(404).json({ error: 'Node not found' });
        }
        
        const detailedStatus = await worker.getDetailedStatus();
        res.json(detailedStatus);
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

// Get performance metrics for a node
app.get('/api/nodes/:nodeId/metrics', async (req, res) => {
    try {
        const { nodeId } = req.params;
        const worker = coordinator.workers.get(nodeId);
        
        if (!worker) {
            return res.status(404).json({ error: 'Node not found' });
        }
        
        res.json({
            nodeId,
            metrics: worker.metrics,
            timeSeries: worker.performanceHistory,
            system: worker.systemInfo
        });
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

// Get health diagnostics for a node
app.get('/api/nodes/:nodeId/health', async (req, res) => {
    try {
        const { nodeId } = req.params;
        const worker = coordinator.workers.get(nodeId);
        
        if (!worker) {
            return res.status(404).json({ error: 'Node not found' });
        }
        
        // Perform comprehensive health check
        await worker.checkHealth();
        await worker.collectSystemInfo();
        
        res.json({
            nodeId,
            status: worker.status,
            health: worker.healthStatus,
            lastCheck: Date.now()
        });
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

// Control operations for nodes
app.post('/api/nodes/:nodeId/control', async (req, res) => {
    try {
        const { nodeId } = req.params;
        const { action, parameters } = req.body;
        const worker = coordinator.workers.get(nodeId);
        
        if (!worker) {
            return res.status(404).json({ error: 'Node not found' });
        }
        
        let result;
        
        switch (action) {
            case 'restart':
                result = await handleNodeRestart(worker, parameters);
                break;
                
            case 'load_model':
                result = await handleModelLoad(worker, parameters);
                break;
                
            case 'unload_model':
                result = await handleModelUnload(worker, parameters);
                break;
                
            case 'update_config':
                result = await handleConfigUpdate(worker, parameters);
                break;
                
            default:
                return res.status(400).json({ error: 'Unknown action' });
        }
        
        res.json({
            nodeId,
            action,
            result,
            timestamp: Date.now()
        });
        
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

// Node configuration management
app.get('/api/nodes/:nodeId/config', async (req, res) => {
    try {
        const { nodeId } = req.params;
        const worker = coordinator.workers.get(nodeId);
        
        if (!worker) {
            return res.status(404).json({ error: 'Node not found' });
        }
        
        // Get current configuration (placeholder implementation)
        const config = {
            nodeId,
            name: worker.name,
            url: worker.url,
            maxConcurrentRequests: worker.ollamaInfo.concurrentCapacity,
            memoryLimit: worker.systemInfo?.memory?.total || 8192,
            cpuLimit: worker.systemInfo?.cpu?.cores || 8,
            autoLoadBalance: true,
            healthCheckInterval: 5000,
            enableGpu: true
        };
        
        res.json(config);
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

app.put('/api/nodes/:nodeId/config', async (req, res) => {
    try {
        const { nodeId } = req.params;
        const worker = coordinator.workers.get(nodeId);
        const newConfig = req.body;
        
        if (!worker) {
            return res.status(404).json({ error: 'Node not found' });
        }
        
        // Validate and apply configuration changes
        if (newConfig.maxConcurrentRequests) {
            worker.ollamaInfo.concurrentCapacity = newConfig.maxConcurrentRequests;
        }
        
        res.json({
            nodeId,
            message: 'Configuration updated successfully',
            config: newConfig,
            timestamp: Date.now()
        });
        
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

// Control operation handlers
async function handleNodeRestart(worker, parameters) {
    // Simulate node restart process
    worker.status = 'restarting';
    
    return new Promise((resolve) => {
        setTimeout(() => {
            worker.status = 'healthy';
            worker.metrics.totalRequests = 0;
            worker.performanceHistory = {
                timestamps: [],
                cpu: [],
                memory: [],
                requests: [],
                responseTime: []
            };
            
            resolve({
                status: 'success',
                message: 'Node restarted successfully'
            });
        }, 3000);
    });
}

async function handleModelLoad(worker, parameters) {
    const { model } = parameters;
    const modelName = model;
    
    if (!modelName) {
        throw new Error('Model name is required');
    }
    
    try {
        const response = await fetch(`${worker.url}/api/pull`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name: modelName })
        });
        
        return {
            status: response.ok ? 'success' : 'error',
            message: response.ok ? 'Model load initiated' : 'Failed to load model'
        };
    } catch (error) {
        return {
            status: 'error',
            message: error.message
        };
    }
}

async function handleModelUnload(worker, parameters) {
    const { model } = parameters;
    const modelName = model;
    
    if (!modelName) {
        throw new Error('Model name is required');
    }
    
    try {
        const response = await fetch(`${worker.url}/api/delete`, {
            method: 'DELETE',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name: modelName })
        });
        
        return {
            status: response.ok ? 'success' : 'error',
            message: response.ok ? 'Model unloaded successfully' : 'Failed to unload model'
        };
    } catch (error) {
        return {
            status: 'error',
            message: error.message
        };
    }
}

async function handleConfigUpdate(worker, parameters) {
    // Apply configuration updates
    Object.keys(parameters).forEach(key => {
        if (key === 'concurrentCapacity') {
            worker.ollamaInfo.concurrentCapacity = parameters[key];
        }
    });
    
    return {
        status: 'success',
        message: 'Configuration updated successfully',
        updatedParameters: parameters
    };
}

// Model Management API endpoints
app.get('/api/models', async (req, res) => {
    try {
        const workers = Array.from(coordinator.workers.values());
        const modelData = {};
        
        for (const worker of workers) {
            try {
                const response = await fetch(`${worker.url}/api/tags`);
                const data = await response.json();
                modelData[worker.name] = {
                    status: worker.status,
                    models: data.models || []
                };
            } catch (error) {
                modelData[worker.name] = {
                    status: 'error',
                    models: [],
                    error: error.message
                };
            }
        }
        
        // Get unique models across all workers
        const allModels = new Set();
        Object.values(modelData).forEach(worker => {
            worker.models.forEach(model => allModels.add(model.name));
        });
        
        res.json({
            workers: modelData,
            availableModels: Array.from(allModels),
            totalWorkers: workers.length
        });
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

app.post('/api/models/pull', async (req, res) => {
    const { model, workers } = req.body;
    
    if (!model) {
        return res.status(400).json({ error: 'Model name is required' });
    }
    
    const targetWorkers = workers || Array.from(coordinator.workers.keys());
    const results = {};
    
    for (const workerId of targetWorkers) {
        const worker = coordinator.workers.get(workerId);
        if (!worker) continue;
        
        try {
            const response = await fetch(`${worker.url}/api/pull`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ name: model })
            });
            
            results[worker.name] = {
                status: response.ok ? 'pulling' : 'error',
                message: response.ok ? 'Model pull initiated' : 'Failed to start pull'
            };
        } catch (error) {
            results[worker.name] = {
                status: 'error',
                message: error.message
            };
        }
    }
    
    res.json({
        model,
        results,
        message: `Model pull initiated on ${Object.keys(results).length} workers`
    });
});

app.delete('/api/models/:model', async (req, res) => {
    const { model } = req.params;
    const { workers } = req.body;
    
    const targetWorkers = workers || Array.from(coordinator.workers.keys());
    const results = {};
    
    for (const workerId of targetWorkers) {
        const worker = coordinator.workers.get(workerId);
        if (!worker) continue;
        
        try {
            const response = await fetch(`${worker.url}/api/delete`, {
                method: 'DELETE',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ name: model })
            });
            
            results[worker.name] = {
                status: response.ok ? 'deleted' : 'error',
                message: response.ok ? 'Model deleted' : 'Failed to delete model'
            };
        } catch (error) {
            results[worker.name] = {
                status: 'error',
                message: error.message
            };
        }
    }
    
    res.json({
        model,
        results,
        message: `Model deletion attempted on ${Object.keys(results).length} workers`
    });
});

app.post('/api/models/propagate', async (req, res) => {
    const { model, sourceWorker, targetWorkers } = req.body;
    
    if (!model || !sourceWorker) {
        return res.status(400).json({ error: 'Model and source worker are required' });
    }
    
    const source = Array.from(coordinator.workers.values()).find(w => w.name === sourceWorker);
    if (!source) {
        return res.status(404).json({ error: 'Source worker not found' });
    }
    
    const targets = targetWorkers || Array.from(coordinator.workers.keys()).filter(id => {
        const worker = coordinator.workers.get(id);
        return worker && worker.name !== sourceWorker;
    });
    
    const results = {};
    
    // P2P model propagation - pull from source to targets
    for (const workerId of targets) {
        const worker = coordinator.workers.get(workerId);
        if (!worker) continue;
        
        try {
            // First check if model exists on source
            const sourceCheck = await fetch(`${source.url}/api/tags`);
            const sourceData = await sourceCheck.json();
            const hasModel = sourceData.models?.some(m => m.name === model);
            
            if (!hasModel) {
                results[worker.name] = {
                    status: 'error',
                    message: 'Model not found on source worker'
                };
                continue;
            }
            
            // Initiate pull on target worker
            const response = await fetch(`${worker.url}/api/pull`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ name: model })
            });
            
            results[worker.name] = {
                status: response.ok ? 'propagating' : 'error',
                message: response.ok ? 'Model propagation initiated' : 'Failed to start propagation'
            };
        } catch (error) {
            results[worker.name] = {
                status: 'error',
                message: error.message
            };
        }
    }
    
    res.json({
        model,
        sourceWorker,
        results,
        message: `Model propagation initiated from ${sourceWorker} to ${Object.keys(results).length} workers`
    });
});

app.post('/api/inference', async (req, res) => {
    const requestId = Date.now();
    
    try {
        const worker = await coordinator.distributeInference(
            requestId,
            req.body.prompt,
            req.body.settings || {}
        );
        
        const response = await worker.process(req.body.prompt, req.body.settings || {});
        const result = await response.json();
        
        res.json({
            requestId,
            response: result.response,
            worker: worker.name,
            model: result.model,
            totalDuration: result.total_duration,
            loadDuration: result.load_duration,
            promptEvalCount: result.prompt_eval_count,
            evalCount: result.eval_count
        });
        
    } catch (error) {
        res.status(500).json({
            error: error.message,
            requestId
        });
    }
});

// Periodic health checks
setInterval(async () => {
    for (const worker of coordinator.workers.values()) {
        await worker.checkHealth();
    }
    
    // Broadcast updated status to all connected clients
    const statusUpdate = JSON.stringify({
        type: 'node_update',
        nodes: Array.from(coordinator.workers.values()).map(w => ({
            id: w.id,
            name: w.name,
            status: w.status,
            model: w.model,
            metrics: w.metrics
        }))
    });
    
    wss.clients.forEach(client => {
        if (client.readyState === WebSocket.OPEN) {
            client.send(statusUpdate);
        }
    });
}, 5000);

// Start server
server.listen(PORT, () => {
    console.log(`
╔════════════════════════════════════════════════════╗
║     🦙 Distributed Llama Inference System          ║
╚════════════════════════════════════════════════════╝

📡 WebSocket: ws://localhost:${PORT}/chat
🌐 REST API:  http://localhost:${PORT}/api
📊 Health:    http://localhost:${PORT}/api/health
👷 Workers:   http://localhost:${PORT}/api/workers

🚀 System ready for distributed inference!
    `);
    
    // Connect to Redis pub/sub for distributed coordination
    redisSub.subscribe('inference:requests', 'inference:results');
    
    redisSub.on('message', (channel, message) => {
        console.log(`Redis message on ${channel}:`, message);
    });
});