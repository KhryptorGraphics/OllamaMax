/**
 * Simplified Distributed Llama API Server
 * Works with existing Ollama deployment on port 13000
 */

const WebSocket = require('ws');
const http = require('http');
const express = require('express');
const cors = require('cors');

const app = express();
app.use(cors());
app.use(express.json());

// Configuration - Use a different port since Ollama is on 13000
const PORT = process.env.PORT || 13100;
const OLLAMA_URL = process.env.OLLAMA_URL || 'http://localhost:13000';

// Simple in-memory storage for demo
const nodes = [
    { id: 'node-1', name: 'ollama-primary', url: OLLAMA_URL, status: 'healthy', load: 45, memory: 67, requestsPerSecond: 12, queue: 0 }
];

// Create HTTP server
const server = http.createServer(app);

// Create WebSocket server
const wss = new WebSocket.Server({ 
    server,
    path: '/chat'
});

// WebSocket connection handler
wss.on('connection', (ws) => {
    console.log('New WebSocket client connected');
    
    // Send initial node status
    ws.send(JSON.stringify({
        type: 'node_update',
        nodes: nodes
    }));
    
    ws.on('message', async (message) => {
        try {
            const data = JSON.parse(message);
            console.log('Received message:', data.type);
            
            switch(data.type) {
                case 'inference':
                    await handleInference(ws, data);
                    break;
                case 'get_nodes':
                    ws.send(JSON.stringify({
                        type: 'node_update',
                        nodes: nodes
                    }));
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
    
    ws.on('error', (error) => {
        console.error('WebSocket error:', error);
    });
});

// Inference handling
async function handleInference(ws, data) {
    const startTime = Date.now();
    const node = nodes[0]; // Use primary node
    
    console.log(`Processing inference request for model: ${data.model || 'default'}`);
    
    // Send initial response
    ws.send(JSON.stringify({
        type: 'response',
        id: data.timestamp,
        node: node.name,
        streaming: data.settings?.streaming || false
    }));
    
    try {
        // For demo, send a mock response since Ollama might not have models loaded
        if (data.settings?.streaming) {
            // Simulate streaming
            const response = "I'm a distributed Llama inference system. This is a demo response showing that the WebSocket connection is working properly. ";
            const words = response.split(' ');
            
            for (let i = 0; i < words.length; i++) {
                ws.send(JSON.stringify({
                    type: 'stream_chunk',
                    id: data.timestamp,
                    chunk: words[i] + ' ',
                    done: i === words.length - 1
                }));
                await new Promise(resolve => setTimeout(resolve, 100));
            }
        } else {
            // Send complete response
            ws.send(JSON.stringify({
                type: 'response',
                id: data.timestamp,
                content: "I'm a distributed Llama inference system. This is a demo response.",
                node: node.name,
                streaming: false
            }));
        }
        
        // Send metrics
        const latency = Date.now() - startTime;
        ws.send(JSON.stringify({
            type: 'metrics',
            latency,
            node: node.name
        }));
        
    } catch (error) {
        console.error('Inference error:', error);
        ws.send(JSON.stringify({
            type: 'error',
            message: `Inference failed: ${error.message}`
        }));
    }
}

// REST API endpoints
app.get('/api/health', (req, res) => {
    res.json({
        status: 'healthy',
        nodes: nodes.filter(n => n.status === 'healthy').length,
        totalNodes: nodes.length,
        queueLength: 0,
        uptime: process.uptime()
    });
});

app.get('/api/nodes', (req, res) => {
    res.json({
        nodes: nodes,
        queueLength: 0
    });
});

// Start server
server.listen(PORT, () => {
    console.log(`\n🦙 Distributed Llama API Server`);
    console.log(`================================`);
    console.log(`WebSocket endpoint: ws://localhost:${PORT}/chat`);
    console.log(`REST API: http://localhost:${PORT}/api`);
    console.log(`Health check: http://localhost:${PORT}/api/health`);
    console.log(`\nChat UI should connect to: ws://localhost:${PORT}/chat`);
});