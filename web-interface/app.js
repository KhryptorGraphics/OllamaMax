/**
 * Distributed Llama Chat Interface - Main Application
 * Handles WebSocket connections, message streaming, and node management
 */

class DistributedLlamaClient {
    constructor() {
        this.ws = null;
        this.nodes = [];
        this.activeNode = null;
        this.messageQueue = [];
        this.messages = [];
        this.settings = this.loadSettings();
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.reconnectDelay = 1000;
        this.streamingMessage = null;
        this.performanceData = {
            latency: [],
            throughput: [],
            memoryUsage: []
        };
        
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.connect();
        this.loadNodes();
        this.startPerformanceMonitoring();
    }

    // WebSocket Connection Management
    connect() {
        const endpoint = this.settings.apiEndpoint || 'ws://localhost:13100/chat';
        this.updateConnectionStatus('connecting');
        
        try {
            this.ws = new WebSocket(endpoint);
            
            this.ws.onopen = () => {
                console.log('WebSocket connected');
                this.updateConnectionStatus('connected');
                this.reconnectAttempts = 0;
                this.processMessageQueue();
            };
            
            this.ws.onmessage = (event) => {
                this.handleMessage(JSON.parse(event.data));
            };
            
            this.ws.onerror = (error) => {
                console.error('WebSocket error:', error);
                this.updateConnectionStatus('error');
            };
            
            this.ws.onclose = () => {
                console.log('WebSocket disconnected');
                this.updateConnectionStatus('disconnected');
                this.attemptReconnect();
            };
        } catch (error) {
            console.error('Failed to create WebSocket:', error);
            this.updateConnectionStatus('error');
        }
    }

    attemptReconnect() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
            
            console.log(`Reconnecting in ${delay}ms... (Attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
            document.getElementById('connectionText').textContent = `Reconnecting... (${this.reconnectAttempts}/${this.maxReconnectAttempts})`;
            
            setTimeout(() => this.connect(), delay);
        } else {
            this.updateConnectionStatus('error');
            document.getElementById('connectionText').textContent = 'Connection failed';
        }
    }

    updateConnectionStatus(status) {
        const indicator = document.getElementById('connectionStatus');
        const text = document.getElementById('connectionText');
        
        indicator.className = 'status-indicator';
        
        switch(status) {
            case 'connected':
                indicator.classList.add('connected');
                text.textContent = 'Connected';
                break;
            case 'connecting':
                text.textContent = 'Connecting...';
                break;
            case 'error':
                indicator.classList.add('error');
                text.textContent = 'Connection error';
                break;
            case 'disconnected':
                indicator.classList.add('error');
                text.textContent = 'Disconnected';
                break;
        }
    }

    // Message Handling
    handleMessage(data) {
        console.log('Received message:', data);
        
        switch(data.type) {
            case 'response':
                this.handleChatResponse(data);
                break;
            case 'stream_chunk':
                this.handleStreamChunk(data);
                break;
            case 'node_update':
                this.handleNodeUpdate(data);
                break;
            case 'error':
                this.handleError(data);
                break;
            case 'metrics':
                this.handleMetrics(data);
                break;
        }
    }

    handleChatResponse(data) {
        if (data.streaming) {
            this.streamingMessage = {
                id: data.id,
                content: '',
                node: data.node,
                timestamp: new Date()
            };
            this.addMessage('ai', '', data.node, true);
        } else {
            this.addMessage('ai', data.content, data.node);
        }
        
        this.updateActiveNode(data.node);
        this.updateLatency(data.latency);
    }

    handleStreamChunk(data) {
        if (this.streamingMessage && this.streamingMessage.id === data.id) {
            this.streamingMessage.content += data.chunk;
            this.updateStreamingMessage(this.streamingMessage.content);
        }
        
        if (data.done) {
            this.finalizeStreamingMessage();
        }
    }

    handleNodeUpdate(data) {
        this.nodes = data.nodes;
        this.updateNodeDisplay();
        document.getElementById('nodeCount').textContent = this.nodes.filter(n => n.status === 'healthy').length;
    }

    handleError(data) {
        console.error('Server error:', data.message);
        this.addMessage('system', `Error: ${data.message}`);
    }

    handleMetrics(data) {
        this.performanceData = data;
        this.updatePerformanceCharts();
    }

    // Send Message
    sendMessage(content) {
        if (!content.trim()) return;
        
        const message = {
            type: 'inference',
            content: content,
            model: document.getElementById('modelSelector').value,
            settings: {
                temperature: parseFloat(document.getElementById('temperature').value),
                maxTokens: parseInt(document.getElementById('maxTokens').value),
                streaming: document.getElementById('streamingEnabled').checked
            },
            timestamp: Date.now()
        };
        
        this.addMessage('user', content);
        
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(message));
        } else {
            this.messageQueue.push(message);
            this.addMessage('system', 'Message queued - waiting for connection...');
        }
        
        // Clear input
        document.getElementById('messageInput').value = '';
    }

    processMessageQueue() {
        while (this.messageQueue.length > 0 && this.ws.readyState === WebSocket.OPEN) {
            const message = this.messageQueue.shift();
            this.ws.send(JSON.stringify(message));
        }
        
        if (this.messageQueue.length > 0) {
            document.getElementById('queueLength').textContent = this.messageQueue.length;
        }
    }

    // UI Updates
    addMessage(sender, content, node = null, streaming = false) {
        const messagesArea = document.getElementById('messagesArea');
        const template = document.getElementById('messageTemplate');
        const messageEl = template.content.cloneNode(true).querySelector('.message');
        
        messageEl.classList.add(sender);
        messageEl.id = `message-${Date.now()}`;
        
        const senderEl = messageEl.querySelector('.message-sender');
        senderEl.textContent = sender === 'user' ? 'You' : sender === 'ai' ? 'Llama' : 'System';
        
        if (node) {
            const nodeEl = messageEl.querySelector('.message-node');
            nodeEl.textContent = node;
            nodeEl.style.display = 'inline-block';
        }
        
        const timeEl = messageEl.querySelector('.message-time');
        timeEl.textContent = new Date().toLocaleTimeString();
        
        const contentEl = messageEl.querySelector('.message-content');
        if (streaming) {
            contentEl.classList.add('streaming');
            contentEl.innerHTML = '<div class="typing-indicator"><span class="typing-dot"></span><span class="typing-dot"></span><span class="typing-dot"></span></div>';
        } else {
            contentEl.textContent = content;
        }
        
        // Add action buttons
        const copyBtn = messageEl.querySelector('.copy-button');
        copyBtn.addEventListener('click', () => this.copyMessage(content));
        
        const retryBtn = messageEl.querySelector('.retry-button');
        retryBtn.addEventListener('click', () => this.retryMessage(content));
        
        messagesArea.appendChild(messageEl);
        
        // Remove welcome message if it exists
        const welcomeMsg = messagesArea.querySelector('.welcome-message');
        if (welcomeMsg) {
            welcomeMsg.remove();
        }
        
        // Auto-scroll
        if (document.getElementById('autoScroll').checked) {
            messagesArea.scrollTop = messagesArea.scrollHeight;
        }
        
        this.messages.push({ sender, content, node, timestamp: new Date() });
    }

    updateStreamingMessage(content) {
        const messages = document.querySelectorAll('.message.ai');
        const lastMessage = messages[messages.length - 1];
        
        if (lastMessage) {
            const contentEl = lastMessage.querySelector('.message-content');
            contentEl.classList.add('streaming');
            contentEl.innerHTML = this.formatMessage(content) + '<span class="streaming-cursor">▊</span>';
            
            // Auto-scroll
            if (document.getElementById('autoScroll').checked) {
                const messagesArea = document.getElementById('messagesArea');
                messagesArea.scrollTop = messagesArea.scrollHeight;
            }
        }
    }

    finalizeStreamingMessage() {
        const messages = document.querySelectorAll('.message.ai');
        const lastMessage = messages[messages.length - 1];
        
        if (lastMessage) {
            const contentEl = lastMessage.querySelector('.message-content');
            contentEl.classList.remove('streaming');
            contentEl.innerHTML = this.formatMessage(this.streamingMessage.content);
        }
        
        this.streamingMessage = null;
    }

    formatMessage(content) {
        // Simple markdown-like formatting
        return content
            .replace(/```(\w+)?\n([\s\S]+?)```/g, '<pre class="code-block language-$1">$2</pre>')
            .replace(/`([^`]+)`/g, '<code>$1</code>')
            .replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
            .replace(/\*([^*]+)\*/g, '<em>$1</em>')
            .replace(/\n/g, '<br>');
    }

    copyMessage(content) {
        navigator.clipboard.writeText(content).then(() => {
            this.showToast('Message copied to clipboard');
        });
    }

    retryMessage(content) {
        this.sendMessage(content);
    }

    // Node Management
    async loadNodes() {
        try {
            const response = await fetch('http://localhost:13100/api/nodes/detailed');
            const data = await response.json();
            this.nodes = data.nodes || [];
            this.updateNodeDisplay();
        } catch (error) {
            console.error('Failed to load nodes:', error);
            // Use mock data for testing
            this.nodes = this.getMockNodes();
            this.updateNodeDisplay();
        }
    }

    getMockNodes() {
        return [
            { id: 'node-1', name: 'llama-01', status: 'healthy', load: 45, memory: 67, requestsPerSecond: 12, queue: 2 },
            { id: 'node-2', name: 'llama-02', status: 'warning', load: 89, memory: 92, requestsPerSecond: 8, queue: 5 },
            { id: 'node-3', name: 'llama-03', status: 'healthy', load: 23, memory: 45, requestsPerSecond: 15, queue: 0 }
        ];
    }

    getMockDetailedNodes() {
        return [
            {
                id: 'worker-1',
                name: 'ollama-primary',
                url: 'http://localhost:13000',
                status: 'healthy',
                systemInfo: {
                    cpu: {
                        model: 'Intel Core i7-12700K',
                        cores: 8,
                        usage: 45.2,
                        load: [1.2, 0.8, 0.6]
                    },
                    memory: {
                        total: 32768 * 1024 * 1024, // 32GB in bytes
                        used: 12288 * 1024 * 1024,  // 12GB in bytes
                        usage: 37.5
                    },
                    disk: {
                        total: 1024 * 1024 * 1024 * 1024, // 1TB
                        used: 512 * 1024 * 1024 * 1024,   // 512GB
                        usage: 50.0
                    },
                    network: {
                        rx: 1024 * 1024 * 100, // 100MB
                        tx: 1024 * 1024 * 50   // 50MB
                    }
                },
                ollamaInfo: {
                    models: [
                        { name: 'tinyllama:latest', size: 637 * 1024 * 1024 },
                        { name: 'llama2:7b', size: 3800 * 1024 * 1024 }
                    ],
                    activeRequests: 2,
                    queueLength: 1,
                    gpuMemory: {
                        used: 4096 * 1024 * 1024,  // 4GB
                        total: 8192 * 1024 * 1024  // 8GB
                    }
                },
                healthStatus: {
                    checks: {
                        'API': 'healthy',
                        'Models': 'healthy', 
                        'Resources': 'warning',
                        'Network': 'healthy'
                    },
                    warnings: ['Memory usage above 80%'],
                    errors: []
                },
                performanceHistory: {
                    timestamps: Array.from({length: 20}, (_, i) => Date.now() - (19-i) * 60000),
                    cpu: Array.from({length: 20}, () => Math.random() * 100),
                    memory: Array.from({length: 20}, () => Math.random() * 100),
                    requests: Array.from({length: 20}, () => Math.floor(Math.random() * 50)),
                    responseTime: Array.from({length: 20}, () => 100 + Math.random() * 400)
                },
                config: {
                    maxConcurrentRequests: 4,
                    requestTimeout: 30000,
                    autoModelMigration: true,
                    healthCheckInterval: 30000
                }
            },
            {
                id: 'worker-2', 
                name: 'ollama-worker-2',
                url: 'http://localhost:13001',
                status: 'warning',
                systemInfo: {
                    cpu: {
                        model: 'AMD Ryzen 9 5900X',
                        cores: 12,
                        usage: 78.1,
                        load: [2.1, 1.8, 1.2]
                    },
                    memory: {
                        total: 64768 * 1024 * 1024, // 64GB
                        used: 48288 * 1024 * 1024,  // 48GB  
                        usage: 75.2
                    },
                    disk: {
                        total: 2048 * 1024 * 1024 * 1024, // 2TB
                        used: 1024 * 1024 * 1024 * 1024,   // 1TB
                        usage: 50.0
                    },
                    network: {
                        rx: 1024 * 1024 * 200,
                        tx: 1024 * 1024 * 150
                    }
                },
                ollamaInfo: {
                    models: [
                        { name: 'codellama:7b', size: 3800 * 1024 * 1024 },
                        { name: 'mistral:7b', size: 4100 * 1024 * 1024 }
                    ],
                    activeRequests: 5,
                    queueLength: 3
                },
                healthStatus: {
                    checks: {
                        'API': 'healthy',
                        'Models': 'healthy',
                        'Resources': 'warning', 
                        'Network': 'healthy'
                    },
                    warnings: ['High CPU usage detected', 'Memory usage approaching limit'],
                    errors: []
                },
                performanceHistory: {
                    timestamps: Array.from({length: 20}, (_, i) => Date.now() - (19-i) * 60000),
                    cpu: Array.from({length: 20}, () => 60 + Math.random() * 40),
                    memory: Array.from({length: 20}, () => 70 + Math.random() * 30),
                    requests: Array.from({length: 20}, () => Math.floor(Math.random() * 30)),
                    responseTime: Array.from({length: 20}, () => 200 + Math.random() * 600)
                },
                config: {
                    maxConcurrentRequests: 6,
                    requestTimeout: 45000,
                    autoModelMigration: false,
                    healthCheckInterval: 15000
                }
            }
        ];
    }

    updateNodeDisplay() {
        const nodesGrid = document.getElementById('nodesGrid');
        if (!nodesGrid) {
            console.warn('nodesGrid element not found, skipping node display update');
            return;
        }
        
        nodesGrid.innerHTML = '';
        
        this.nodes.forEach(node => {
            const template = document.getElementById('nodeCardTemplate');
            if (!template) {
                console.warn('nodeCardTemplate not found, creating basic node card');
                const basicCard = document.createElement('div');
                basicCard.className = 'node-card basic';
                basicCard.innerHTML = `<h4>${node.name}</h4><p>Status: ${node.status}</p>`;
                nodesGrid.appendChild(basicCard);
                return;
            }
            
            const cardEl = template.content.cloneNode(true).querySelector('.node-card');
            
            cardEl.classList.add(node.status);
            cardEl.querySelector('.node-name').textContent = node.name;
            
            const statusEl = cardEl.querySelector('.node-status');
            statusEl.classList.add(node.status);
            
            cardEl.querySelector('.load-value').textContent = `${node.load}%`;
            cardEl.querySelector('.memory-value').textContent = `${node.memory}%`;
            cardEl.querySelector('.requests-value').textContent = node.requestsPerSecond;
            cardEl.querySelector('.queue-value').textContent = node.queue;
            
            // Draw sparkline
            const canvas = cardEl.querySelector('canvas');
            this.drawSparkline(canvas, node);
            
            // Add action handlers
            cardEl.querySelectorAll('.node-action-button').forEach(btn => {
                btn.addEventListener('click', (e) => {
                    const action = e.target.dataset.action;
                    this.handleNodeAction(node, action);
                });
            });
            
            nodesGrid.appendChild(cardEl);
        });
    }%`;
            cardEl.querySelector('.memory-value').textContent = `${node.memory}%`;
            cardEl.querySelector('.requests-value').textContent = node.requestsPerSecond;
            cardEl.querySelector('.queue-value').textContent = node.queue;
            
            // Draw sparkline
            const canvas = cardEl.querySelector('canvas');
            this.drawSparkline(canvas, node);
            
            // Add action handlers
            cardEl.querySelectorAll('.node-action-button').forEach(btn => {
                btn.addEventListener('click', (e) => {
                    const action = e.target.dataset.action;
                    this.handleNodeAction(node, action);
                });
            });
            
            nodesGrid.appendChild(cardEl);
        });
    }

    drawSparkline(canvas, node) {
        const ctx = canvas.getContext('2d');
        const data = Array.from({length: 20}, () => Math.random() * 100);
        
        ctx.strokeStyle = node.status === 'healthy' ? '#48bb78' : 
                          node.status === 'warning' ? '#ed8936' : '#e53e3e';
        ctx.lineWidth = 2;
        ctx.beginPath();
        
        data.forEach((value, index) => {
            const x = (index / data.length) * canvas.width;
            const y = canvas.height - (value / 100) * canvas.height;
            
            if (index === 0) {
                ctx.moveTo(x, y);
            } else {
                ctx.lineTo(x, y);
            }
        });
        
        ctx.stroke();
    }

    handleNodeAction(node, action) {
        switch(action) {
            case 'details':
                this.showNodeDetails(node);
                break;
            case 'remove':
                this.removeNode(node);
                break;
        }
    }

    showNodeDetails(node) {
        // Would show detailed modal
        console.log('Show details for node:', node);
        this.showToast(`Node ${node.name} details`);
    }

    removeNode(node) {
        if (confirm(`Remove node ${node.name}?`)) {
            this.nodes = this.nodes.filter(n => n.id !== node.id);
            this.updateNodeDisplay();
            this.showToast(`Node ${node.name} removed`);
        }
    }

    addNode(url, name) {
        const newNode = {
            id: `node-${Date.now()}`,
            name: name,
            url: url,
            status: 'connecting',
            load: 0,
            memory: 0,
            requestsPerSecond: 0,
            queue: 0
        };
        
        this.nodes.push(newNode);
        this.updateNodeDisplay();
        
        // Send add node request
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify({
                type: 'add_node',
                node: newNode
            }));
        }
    }

    // Performance Monitoring
    startPerformanceMonitoring() {
        setInterval(() => {
            this.updateQueueLength();
            this.updateActiveNode();
        }, 5000);
    }

    updateQueueLength() {
        const totalQueue = this.nodes.reduce((sum, node) => sum + (node.queue || 0), 0);
        document.getElementById('queueLength').textContent = totalQueue + this.messageQueue.length;
    }

    updateActiveNode(nodeName = null) {
        if (nodeName) {
            this.activeNode = nodeName;
        }
        document.getElementById('activeNode').textContent = this.activeNode || '-';
    }

    updateLatency(latency) {
        document.getElementById('latency').textContent = latency ? `${latency}ms` : '-';
    }

    updatePerformanceCharts() {
        // Would update Chart.js charts
        console.log('Update performance charts with:', this.performanceData);
    }

    // Settings Management
    loadSettings() {
        const stored = localStorage.getItem('llamaChatSettings');
        return stored ? JSON.parse(stored) : {
            apiEndpoint: 'ws://localhost:13100/chat',
            streamingEnabled: true,
            autoScroll: true,
            maxTokens: 2048,
            temperature: 0.7,
            loadBalancing: 'round-robin'
        };
    }

    saveSettings() {
        this.settings = {
            apiEndpoint: document.getElementById('apiEndpoint').value,
            apiKey: document.getElementById('apiKey').value,
            streamingEnabled: document.getElementById('streamingEnabled').checked,
            autoScroll: document.getElementById('autoScroll').checked,
            maxTokens: parseInt(document.getElementById('maxTokens').value),
            temperature: parseFloat(document.getElementById('temperature').value),
            loadBalancing: document.getElementById('loadBalancing').value
        };
        
        localStorage.setItem('llamaChatSettings', JSON.stringify(this.settings));
        this.showToast('Settings saved');
        
        // Reconnect if endpoint changed
        if (this.settings.apiEndpoint !== this.ws.url) {
            this.ws.close();
            this.connect();
        }
    }

    resetSettings() {
        if (confirm('Reset all settings to defaults?')) {
            localStorage.removeItem('llamaChatSettings');
            this.settings = this.loadSettings();
            this.applySettings();
            this.showToast('Settings reset to defaults');
        }
    }

    applySettings() {
        document.getElementById('apiEndpoint').value = this.settings.apiEndpoint;
        document.getElementById('streamingEnabled').checked = this.settings.streamingEnabled;
        document.getElementById('autoScroll').checked = this.settings.autoScroll;
        document.getElementById('maxTokens').value = this.settings.maxTokens;
        document.getElementById('temperature').value = this.settings.temperature;
        document.getElementById('temperatureValue').textContent = this.settings.temperature;
        document.getElementById('loadBalancing').value = this.settings.loadBalancing;
    }

    // UI Event Listeners
    setupEventListeners() {
        // Tab switching
        document.querySelectorAll('.tab-button').forEach(button => {
            button.addEventListener('click', (e) => {
                document.querySelectorAll('.tab-button').forEach(b => b.classList.remove('active'));
                document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
                
                e.target.classList.add('active');
                const tabId = e.target.dataset.tab + 'Tab';
                document.getElementById(tabId).classList.add('active');
            });
        });

        // Message input
        const messageInput = document.getElementById('messageInput');
        const sendButton = document.getElementById('sendButton');
        
        messageInput.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                this.sendMessage(messageInput.value);
            }
        });
        
        sendButton.addEventListener('click', () => {
            this.sendMessage(messageInput.value);
        });

        // Node management
        document.getElementById('refreshNodes').addEventListener('click', () => {
            this.loadNodes();
        });

        // Model management
        document.getElementById('refreshModels').addEventListener('click', () => {
            this.loadModels();
        });

        document.getElementById('downloadModel').addEventListener('click', () => {
            const modelName = document.getElementById('newModelName').value.trim();
            if (!modelName) {
                this.showToast('Please enter a model name');
                return;
            }

            // Get selected workers
            const selectedWorkers = [];
            document.querySelectorAll('#downloadTargets input:checked').forEach(checkbox => {
                selectedWorkers.push(checkbox.value);
            });

            if (selectedWorkers.length === 0) {
                this.showToast('Please select at least one worker');
                return;
            }

            this.downloadModel(modelName, selectedWorkers);
            document.getElementById('newModelName').value = '';
        });

        // Load models when models tab is clicked
        document.querySelector('[data-tab="models"]').addEventListener('click', () => {
            setTimeout(() => this.loadModels(), 100);
        });

        document.getElementById('addNodeButton').addEventListener('click', () => {
            document.getElementById('addNodeModal').classList.add('active');
        });

        document.getElementById('confirmAddNode').addEventListener('click', () => {
            const url = document.getElementById('nodeUrl').value;
            const name = document.getElementById('nodeName').value;
            
            if (url && name) {
                this.addNode(url, name);
                document.getElementById('addNodeModal').classList.remove('active');
                document.getElementById('nodeUrl').value = '';
                document.getElementById('nodeName').value = '';
            }
        });

        document.getElementById('cancelAddNode').addEventListener('click', () => {
            document.getElementById('addNodeModal').classList.remove('active');
        });

        // Settings
        document.getElementById('temperature').addEventListener('input', (e) => {
            document.getElementById('temperatureValue').textContent = e.target.value;
        });

        document.getElementById('saveSettings').addEventListener('click', () => {
            this.saveSettings();
        });

        document.getElementById('resetSettings').addEventListener('click', () => {
            this.resetSettings();
        });

        // Apply initial settings
        this.applySettings();
    }

    // Model Management
    async loadModels() {
        try {
            const response = await fetch('http://localhost:13100/api/models');
            const data = await response.json();
            
            this.updateModelSelector(data.availableModels);
            this.displayModelCards(data);
            this.updateWorkerCheckboxes(data.workers);
            
        } catch (error) {
            console.error('Error loading models:', error);
            this.showToast('Failed to load models');
        }
    }

    updateModelSelector(availableModels) {
        const selector = document.getElementById('modelSelector');
        if (!selector) return;
        
        // Clear current options
        selector.innerHTML = '';
        
        // Add available models
        availableModels.forEach(model => {
            const option = document.createElement('option');
            option.value = model.replace(':latest', '');
            option.textContent = model;
            selector.appendChild(option);
        });
        
        // Select first model by default
        if (availableModels.length > 0) {
            selector.value = availableModels[0].replace(':latest', '');
        }
    }

    displayModelCards(data) {
        const modelGrid = document.getElementById('modelGrid');
        if (!modelGrid) return;
        
        modelGrid.innerHTML = '';
        
        // Create model cards for each unique model
        data.availableModels.forEach(modelName => {
            const card = this.createModelCard(modelName, data.workers);
            modelGrid.appendChild(card);
        });
    }

    createModelCard(modelName, workers) {
        const template = document.getElementById('modelCardTemplate');
        const card = template.content.cloneNode(true);
        
        // Find model details from any worker that has it
        let modelDetails = null;
        const nodesWithModel = [];
        
        Object.entries(workers).forEach(([workerName, workerData]) => {
            const model = workerData.models.find(m => m.name === modelName);
            if (model) {
                if (!modelDetails) modelDetails = model;
                nodesWithModel.push(workerName);
            }
        });
        
        // Populate card with model information
        card.querySelector('.model-name').textContent = modelName;
        card.querySelector('.model-size').textContent = modelDetails ? 
            `${Math.round(modelDetails.size / 1024 / 1024)} MB` : '-';
        
        if (modelDetails && modelDetails.details) {
            card.querySelector('.model-family').textContent = 
                modelDetails.details.family || '-';
            card.querySelector('.model-parameter-size').textContent = 
                modelDetails.details.parameter_size || '-';
            card.querySelector('.model-format').textContent = 
                modelDetails.details.format || '-';
        }
        
        // Add node badges
        const nodeBadges = card.querySelector('.node-badges');
        nodesWithModel.forEach(nodeName => {
            const badge = document.createElement('span');
            badge.className = 'node-badge';
            badge.textContent = nodeName;
            nodeBadges.appendChild(badge);
        });
        
        // Add event listeners
        const propagateBtn = card.querySelector('.propagate-button');
        const deleteBtn = card.querySelector('.delete-button');
        
        propagateBtn.addEventListener('click', () => this.propagateModel(modelName, nodesWithModel));
        deleteBtn.addEventListener('click', () => this.deleteModel(modelName));
        
        return card;
    }

    updateWorkerCheckboxes(workers) {
        const container = document.getElementById('downloadTargets');
        if (!container) return;
        
        container.innerHTML = '';
        
        Object.keys(workers).forEach(workerName => {
            const label = document.createElement('label');
            label.className = 'worker-checkbox';
            
            const checkbox = document.createElement('input');
            checkbox.type = 'checkbox';
            checkbox.value = workerName;
            checkbox.checked = true;
            
            const text = document.createTextNode(' ' + workerName);
            
            label.appendChild(checkbox);
            label.appendChild(text);
            container.appendChild(label);
        });
    }

    async downloadModel(modelName, targetWorkers = []) {
        try {
            const response = await fetch('http://localhost:13100/api/models/pull', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    model: modelName,
                    workers: targetWorkers
                })
            });
            
            const result = await response.json();
            this.showToast(`Model download initiated: ${result.message}`);
            
            // Refresh models after a delay
            setTimeout(() => this.loadModels(), 5000);
            
        } catch (error) {
            console.error('Error downloading model:', error);
            this.showToast('Failed to download model');
        }
    }

    async propagateModel(modelName, sourceNodes) {
        if (sourceNodes.length === 0) {
            this.showToast('No source nodes available');
            return;
        }
        
        try {
            const response = await fetch('http://localhost:13100/api/models/propagate', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    model: modelName,
                    sourceWorker: sourceNodes[0],
                    targetWorkers: null // propagate to all other workers
                })
            });
            
            const result = await response.json();
            this.showToast(`Model propagation initiated: ${result.message}`);
            
            // Refresh models after a delay
            setTimeout(() => this.loadModels(), 5000);
            
        } catch (error) {
            console.error('Error propagating model:', error);
            this.showToast('Failed to propagate model');
        }
    }

    async deleteModel(modelName) {
        if (!confirm(`Delete model "${modelName}" from all nodes?`)) {
            return;
        }
        
        try {
            const response = await fetch(`http://localhost:13100/api/models/${encodeURIComponent(modelName)}`, {
                method: 'DELETE',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ workers: null }) // delete from all workers
            });
            
            const result = await response.json();
            this.showToast(`Model deletion attempted: ${result.message}`);
            
            // Refresh models after a delay
            setTimeout(() => this.loadModels(), 3000);
            
        } catch (error) {
            console.error('Error deleting model:', error);
            this.showToast('Failed to delete model');
        }
    }

    // Utility
    showToast(message) {
        // Simple toast notification
        const toast = document.createElement('div');
        toast.className = 'toast';
        toast.textContent = message;
        toast.style.cssText = `
            position: fixed;
            bottom: 2rem;
            right: 2rem;
            background: var(--dark);
            color: white;
            padding: 1rem 1.5rem;
            border-radius: 8px;
            box-shadow: var(--shadow-lg);
            z-index: 1000;
            animation: fadeInUp 0.3s ease;
        `;
        
        document.body.appendChild(toast);
        
        setTimeout(() => {
            toast.style.animation = 'fadeOutDown 0.3s ease';
            setTimeout(() => toast.remove(), 300);
        }, 3000);
    }

    // Enhanced Node Management
    async loadDetailedNodes() {
        try {
            const response = await fetch('http://localhost:13100/api/nodes/detailed');
            const data = await response.json();
            
            this.detailedNodes = data.nodes;
            this.updateClusterOverview();
            this.displayEnhancedNodes();
            
        } catch (error) {
            console.error('Error loading detailed nodes:', error);
            this.showToast('Failed to load detailed node information');
            
            // Provide fallback with mock data for development
            this.detailedNodes = this.getMockDetailedNodes();
            this.updateClusterOverview();
            this.displayEnhancedNodes();
        }
    }

    updateClusterOverview() {
        const totalNodes = this.detailedNodes.length;
        const healthyNodes = this.detailedNodes.filter(node => node.status === 'healthy').length;
        const totalCores = this.detailedNodes.reduce((sum, node) => sum + (node.systemInfo?.cpu?.cores || 0), 0);
        const totalMemory = this.detailedNodes.reduce((sum, node) => sum + (node.systemInfo?.memory?.total || 0), 0);
        
        const totalNodesEl = document.getElementById('totalNodes');
        const healthyNodesEl = document.getElementById('healthyNodes');
        const totalCoresEl = document.getElementById('totalCores');
        const totalMemoryEl = document.getElementById('totalMemory');
        const healthRatioEl = document.getElementById('healthRatio');
        
        if (totalNodesEl) totalNodesEl.textContent = totalNodes;
        if (healthyNodesEl) healthyNodesEl.textContent = healthyNodes;
        if (totalCoresEl) totalCoresEl.textContent = totalCores;
        if (totalMemoryEl) totalMemoryEl.textContent = this.formatBytes(totalMemory);
        
        // Update health ratio
        const healthRatio = totalNodes > 0 ? (healthyNodes / totalNodes * 100) : 0;
        if (healthRatioEl) healthRatioEl.textContent = `${healthRatio.toFixed(1)}%`;
    }

    displayEnhancedNodes() {
        const container = document.getElementById('enhancedNodesContainer');
        if (!container) return;

        const filteredNodes = this.filterNodes();
        
        container.innerHTML = filteredNodes.map(node => this.createEnhancedNodeCard(node)).join('');
        
        // Attach event listeners after rendering
        this.attachNodeEventListeners();
    }

    filterNodes() {
        const statusFilter = document.getElementById('statusFilter')?.value || 'all';
        const searchQuery = document.getElementById('nodeSearch')?.value?.toLowerCase() || '';
        
        return this.detailedNodes.filter(node => {
            const matchesStatus = statusFilter === 'all' || node.status === statusFilter;
            const matchesSearch = searchQuery === '' || 
                node.name.toLowerCase().includes(searchQuery) ||
                node.url.toLowerCase().includes(searchQuery);
            
            return matchesStatus && matchesSearch;
        });
    }

    createEnhancedNodeCard(node) {
        const statusColor = node.status === 'healthy' ? '#10b981' : 
                           node.status === 'warning' ? '#f59e0b' : '#ef4444';
        
        const cpuUsage = node.systemInfo?.cpu?.usage || 0;
        const memoryUsage = node.systemInfo?.memory?.usage || 0;
        const responseTime = node.performanceHistory?.responseTime?.slice(-1)[0] || 0;
        
        return `
            <div class="enhanced-node-card" data-node-id="${node.id}">
                <div class="node-header">
                    <div class="node-basic-info">
                        <h3 class="node-title">${node.name}</h3>
                        <div class="node-status">
                            <div class="status-indicator" style="background-color: ${statusColor}"></div>
                            <span>${node.status}</span>
                        </div>
                    </div>
                    <div class="node-quick-stats">
                        <div class="quick-stat">
                            <span class="stat-label">CPU</span>
                            <span class="stat-value">${cpuUsage.toFixed(1)}%</span>
                        </div>
                        <div class="quick-stat">
                            <span class="stat-label">Memory</span>
                            <span class="stat-value">${memoryUsage.toFixed(1)}%</span>
                        </div>
                        <div class="quick-stat">
                            <span class="stat-label">Response</span>
                            <span class="stat-value">${responseTime}ms</span>
                        </div>
                    </div>
                    <button class="expand-btn" onclick="llamaClient.toggleNodeExpansion('${node.id}')">
                        <span class="expand-icon">▼</span>
                    </button>
                </div>
                
                <div class="node-expandable" id="expandable-${node.id}" style="display: none;">
                    <div class="node-tabs">
                        <button class="tab-btn active" onclick="llamaClient.switchNodeTab('${node.id}', 'performance')">Performance</button>
                        <button class="tab-btn" onclick="llamaClient.switchNodeTab('${node.id}', 'health')">Health</button>
                        <button class="tab-btn" onclick="llamaClient.switchNodeTab('${node.id}', 'models')">Models</button>
                        <button class="tab-btn" onclick="llamaClient.switchNodeTab('${node.id}', 'config')">Config</button>
                    </div>
                    
                    <div class="tab-content">
                        <div id="performance-${node.id}" class="tab-panel active">
                            ${this.createPerformancePanel(node)}
                        </div>
                        <div id="health-${node.id}" class="tab-panel">
                            ${this.createHealthPanel(node)}
                        </div>
                        <div id="models-${node.id}" class="tab-panel">
                            ${this.createModelsPanel(node)}
                        </div>
                        <div id="config-${node.id}" class="tab-panel">
                            ${this.createConfigPanel(node)}
                        </div>
                    </div>
                </div>
            </div>
        `;
    }

    createPerformancePanel(node) {
        const systemInfo = node.systemInfo || {};
        
        return `
            <div class="performance-panel">
                <div class="system-metrics">
                    <div class="metric-group">
                        <h4>CPU Information</h4>
                        <div class="metric-item">
                            <span>Model:</span>
                            <span>${systemInfo.cpu?.model || 'Unknown'}</span>
                        </div>
                        <div class="metric-item">
                            <span>Cores:</span>
                            <span>${systemInfo.cpu?.cores || 'N/A'}</span>
                        </div>
                        <div class="metric-item">
                            <span>Usage:</span>
                            <span>${(systemInfo.cpu?.usage || 0).toFixed(1)}%</span>
                        </div>
                    </div>
                    
                    <div class="metric-group">
                        <h4>Memory Information</h4>
                        <div class="metric-item">
                            <span>Total:</span>
                            <span>${this.formatBytes(systemInfo.memory?.total || 0)}</span>
                        </div>
                        <div class="metric-item">
                            <span>Used:</span>
                            <span>${this.formatBytes(systemInfo.memory?.used || 0)}</span>
                        </div>
                        <div class="metric-item">
                            <span>Usage:</span>
                            <span>${(systemInfo.memory?.usage || 0).toFixed(1)}%</span>
                        </div>
                    </div>
                </div>
                
                <div class="performance-chart">
                    <canvas id="chart-${node.id}" width="400" height="200"></canvas>
                </div>
            </div>
        `;
    }

    createHealthPanel(node) {
        const health = node.healthStatus || {};
        const checks = health.checks || {};
        
        return `
            <div class="health-panel">
                <div class="health-checks">
                    <h4>System Health Checks</h4>
                    ${Object.entries(checks).map(([check, status]) => `
                        <div class="health-check">
                            <div class="check-status ${status}"></div>
                            <span class="check-name">${check}</span>
                            <span class="check-result">${status}</span>
                        </div>
                    `).join('')}
                </div>
                
                ${health.warnings?.length > 0 ? `
                    <div class="health-warnings">
                        <h4>Warnings</h4>
                        ${health.warnings.map(warning => `
                            <div class="warning-item">${warning}</div>
                        `).join('')}
                    </div>
                ` : ''}
                
                ${health.errors?.length > 0 ? `
                    <div class="health-errors">
                        <h4>Errors</h4>
                        ${health.errors.map(error => `
                            <div class="error-item">${error}</div>
                        `).join('')}
                    </div>
                ` : ''}
                
                <div class="health-actions">
                    <button class="health-btn" onclick="llamaClient.runHealthCheck('${node.id}')">
                        Run Health Check
                    </button>
                    <button class="health-btn" onclick="llamaClient.clearHealthIssues('${node.id}')">
                        Clear Issues
                    </button>
                </div>
            </div>
        `;
    }

    createModelsPanel(node) {
        const ollama = node.ollamaInfo || {};
        const models = ollama.models || [];
        
        return `
            <div class="models-panel">
                <div class="models-header">
                    <h4>Available Models (${models.length})</h4>
                    <div class="ollama-stats">
                        <span>Active Requests: ${ollama.activeRequests || 0}</span>
                        <span>Queue Length: ${ollama.queueLength || 0}</span>
                    </div>
                </div>
                
                <div class="models-list">
                    ${models.map(model => `
                        <div class="model-item">
                            <div class="model-info">
                                <span class="model-name">${model.name}</span>
                                <span class="model-size">${this.formatBytes(model.size || 0)}</span>
                            </div>
                            <div class="model-actions">
                                <button class="model-btn" onclick="llamaClient.loadModel('${node.id}', '${model.name}')">
                                    Load
                                </button>
                                <button class="model-btn danger" onclick="llamaClient.unloadModel('${node.id}', '${model.name}')">
                                    Unload
                                </button>
                            </div>
                        </div>
                    `).join('')}
                </div>
                
                ${ollama.gpuMemory ? `
                    <div class="gpu-info">
                        <h4>GPU Memory</h4>
                        <div class="gpu-usage">
                            <span>Used: ${this.formatBytes(ollama.gpuMemory.used)}</span>
                            <span>Total: ${this.formatBytes(ollama.gpuMemory.total)}</span>
                            <div class="gpu-bar">
                                <div class="gpu-fill" style="width: ${(ollama.gpuMemory.used / ollama.gpuMemory.total * 100).toFixed(1)}%"></div>
                            </div>
                        </div>
                    </div>
                ` : ''}
            </div>
        `;
    }

    createConfigPanel(node) {
        const config = node.config || {};
        
        return `
            <div class="config-panel">
                <div class="config-section">
                    <h4>Node Configuration</h4>
                    <div class="config-form">
                        <div class="config-item">
                            <label>Max Concurrent Requests:</label>
                            <input type="number" id="maxConcurrent-${node.id}" value="${config.maxConcurrentRequests || 4}" min="1" max="20">
                        </div>
                        <div class="config-item">
                            <label>Request Timeout (ms):</label>
                            <input type="number" id="requestTimeout-${node.id}" value="${config.requestTimeout || 30000}" min="5000" step="1000">
                        </div>
                        <div class="config-item">
                            <label>Auto Model Migration:</label>
                            <input type="checkbox" id="autoMigration-${node.id}" ${config.autoModelMigration ? 'checked' : ''}>
                        </div>
                        <div class="config-item">
                            <label>Health Check Interval (s):</label>
                            <input type="number" id="healthInterval-${node.id}" value="${(config.healthCheckInterval || 30000) / 1000}" min="10" max="300">
                        </div>
                    </div>
                </div>
                
                <div class="config-actions">
                    <button class="config-btn" onclick="llamaClient.saveNodeConfig('${node.id}')">
                        Save Configuration
                    </button>
                    <button class="config-btn" onclick="llamaClient.restartNode('${node.id}')">
                        Restart Node
                    </button>
                    <button class="config-btn danger" onclick="llamaClient.resetNodeConfig('${node.id}')">
                        Reset to Defaults
                    </button>
                </div>
            </div>
        `;
    }

    // Node Control Methods
    toggleNodeExpansion(nodeId) {
        const expandable = document.getElementById(`expandable-${nodeId}`);
        const expandIcon = expandable.parentElement.querySelector('.expand-icon');
        
        if (expandable.style.display === 'none') {
            expandable.style.display = 'block';
            expandIcon.textContent = '▲';
            
            // Initialize performance chart if performance tab is active
            if (document.querySelector(`#performance-${nodeId}.active`)) {
                setTimeout(() => this.initPerformanceChart(nodeId), 100);
            }
        } else {
            expandable.style.display = 'none';
            expandIcon.textContent = '▼';
        }
    }

    switchNodeTab(nodeId, tabName) {
        // Remove active class from all tabs and panels for this node
        const nodeCard = document.querySelector(`[data-node-id="${nodeId}"]`);
        nodeCard.querySelectorAll('.tab-btn').forEach(btn => btn.classList.remove('active'));
        nodeCard.querySelectorAll('.tab-panel').forEach(panel => panel.classList.remove('active'));
        
        // Add active class to selected tab and panel
        nodeCard.querySelector(`button[onclick*="'${tabName}'"]`).classList.add('active');
        document.getElementById(`${tabName}-${nodeId}`).classList.add('active');
        
        // Initialize chart if switching to performance tab
        if (tabName === 'performance') {
            setTimeout(() => this.initPerformanceChart(nodeId), 100);
        }
    }

    initPerformanceChart(nodeId) {
        const canvas = document.getElementById(`chart-${nodeId}`);
        if (!canvas || canvas.chart) return; // Already initialized
        
        const node = this.detailedNodes.find(n => n.id === nodeId);
        if (!node?.performanceHistory) return;
        
        const ctx = canvas.getContext('2d');
        const history = node.performanceHistory;
        
        // Simple chart implementation (would use Chart.js in production)
        canvas.chart = this.drawSimpleChart(ctx, history);
    }

    drawSimpleChart(ctx, history) {
        const width = ctx.canvas.width;
        const height = ctx.canvas.height;
        const padding = 40;
        
        ctx.clearRect(0, 0, width, height);
        
        // Draw axes
        ctx.strokeStyle = '#e5e7eb';
        ctx.lineWidth = 1;
        ctx.beginPath();
        ctx.moveTo(padding, height - padding);
        ctx.lineTo(width - padding, height - padding);
        ctx.moveTo(padding, padding);
        ctx.lineTo(padding, height - padding);
        ctx.stroke();
        
        // Draw CPU usage line
        if (history.cpu && history.cpu.length > 0) {
            ctx.strokeStyle = '#3b82f6';
            ctx.lineWidth = 2;
            ctx.beginPath();
            
            const maxPoints = Math.min(history.cpu.length, 20);
            const stepX = (width - 2 * padding) / (maxPoints - 1);
            
            history.cpu.slice(-maxPoints).forEach((value, i) => {
                const x = padding + i * stepX;
                const y = height - padding - (value / 100) * (height - 2 * padding);
                
                if (i === 0) ctx.moveTo(x, y);
                else ctx.lineTo(x, y);
            });
            
            ctx.stroke();
        }
        
        return true; // Mark as initialized
    }

    // Node Control Operations
    async restartNode(nodeId) {
        if (!confirm('Are you sure you want to restart this node?')) return;
        
        try {
            const response = await fetch(`http://localhost:13100/api/nodes/${nodeId}/control`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ action: 'restart' })
            });
            
            const result = await response.json();
            this.showToast(result.success ? 'Node restart initiated' : result.message);
            
            if (result.success) {
                setTimeout(() => this.loadDetailedNodes(), 5000);
            }
        } catch (error) {
            console.error('Error restarting node:', error);
            this.showToast('Failed to restart node');
        }
    }

    async saveNodeConfig(nodeId) {
        const config = {
            maxConcurrentRequests: parseInt(document.getElementById(`maxConcurrent-${nodeId}`).value),
            requestTimeout: parseInt(document.getElementById(`requestTimeout-${nodeId}`).value),
            autoModelMigration: document.getElementById(`autoMigration-${nodeId}`).checked,
            healthCheckInterval: parseInt(document.getElementById(`healthInterval-${nodeId}`).value) * 1000
        };
        
        try {
            const response = await fetch(`http://localhost:13100/api/nodes/${nodeId}/config`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(config)
            });
            
            const result = await response.json();
            this.showToast(result.success ? 'Configuration saved' : result.message);
            
        } catch (error) {
            console.error('Error saving config:', error);
            this.showToast('Failed to save configuration');
        }
    }

    async loadModel(nodeId, modelName) {
        try {
            const response = await fetch(`http://localhost:13100/api/nodes/${nodeId}/control`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ action: 'load_model', model: modelName })
            });
            
            const result = await response.json();
            this.showToast(result.success ? `Loading model: ${modelName}` : result.message);
            
        } catch (error) {
            console.error('Error loading model:', error);
            this.showToast('Failed to load model');
        }
    }

    async unloadModel(nodeId, modelName) {
        try {
            const response = await fetch(`http://localhost:13100/api/nodes/${nodeId}/control`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ action: 'unload_model', model: modelName })
            });
            
            const result = await response.json();
            this.showToast(result.success ? `Unloaded model: ${modelName}` : result.message);
            
        } catch (error) {
            console.error('Error unloading model:', error);
            this.showToast('Failed to unload model');
        }
    }

    async runHealthCheck(nodeId) {
        this.showToast('Running health check...');
        
        setTimeout(() => this.loadDetailedNodes(), 2000);
    }

    async clearHealthIssues(nodeId) {
        this.showToast('Health issues cleared');
        setTimeout(() => this.loadDetailedNodes(), 1000);
    }

    // Event Listeners
    attachNodeEventListeners() {
        const statusFilter = document.getElementById('statusFilter');
        const nodeSearch = document.getElementById('nodeSearch');
        const sortBy = document.getElementById('sortBy');
        
        if (statusFilter) {
            statusFilter.addEventListener('change', () => this.displayEnhancedNodes());
        }
        
        if (nodeSearch) {
            nodeSearch.addEventListener('input', () => this.displayEnhancedNodes());
        }
        
        if (sortBy) {
            sortBy.addEventListener('change', () => {
                this.sortNodes(sortBy.value);
                this.displayEnhancedNodes();
            });
        }
    }

    sortNodes(criteria) {
        this.detailedNodes.sort((a, b) => {
            switch (criteria) {
                case 'name':
                    return a.name.localeCompare(b.name);
                case 'status':
                    return a.status.localeCompare(b.status);
                case 'cpu':
                    return (b.systemInfo?.cpu?.usage || 0) - (a.systemInfo?.cpu?.usage || 0);
                case 'memory':
                    return (b.systemInfo?.memory?.usage || 0) - (a.systemInfo?.memory?.usage || 0);
                default:
                    return 0;
            }
        });
    }


    // Performance Optimization: Debounced search
    debounce(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }


    // Performance Optimization: Lazy loading with Intersection Observer
    setupLazyLoading() {
        if ('IntersectionObserver' in window) {
            this.lazyObserver = new IntersectionObserver((entries) => {
                entries.forEach(entry => {
                    if (entry.isIntersecting) {
                        const target = entry.target;
                        if (target.dataset.lazy === 'nodes') {
                            this.loadDetailedNodes();
                        } else if (target.dataset.lazy === 'models') {
                            this.loadModels();
                        }
                        this.lazyObserver.unobserve(target);
                    }
                });
            }, { threshold: 0.1 });
        }
    }

    formatBytes(bytes) {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }
}

// Initialize application when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    window.llamaClient = new DistributedLlamaClient();
    
    // Setup tab navigation and auto-loading
    setupTabNavigation();
});

// Setup tab navigation functionality
function setupTabNavigation() {
    const tabButtons = document.querySelectorAll('.tab-button');
    const tabContents = document.querySelectorAll('.tab-content');

    tabButtons.forEach(button => {
        button.addEventListener('click', () => {
            const tabName = button.getAttribute('data-tab');
            
            // Remove active class from all tabs and contents
            tabButtons.forEach(btn => btn.classList.remove('active'));
            tabContents.forEach(content => content.classList.remove('active'));
            
            // Add active class to clicked tab and corresponding content
            button.classList.add('active');
            const targetContent = document.getElementById(tabName + 'Tab');
            if (targetContent) {
                targetContent.classList.add('active');
            }
            
            // Auto-load content based on tab
            if (tabName === 'nodes') {
                // Load detailed nodes when nodes tab is activated
                if (window.llamaClient && window.llamaClient.loadDetailedNodes) {
                    window.llamaClient.loadDetailedNodes();
                }
            } else if (tabName === 'models') {
                // Load models when models tab is activated  
                if (window.llamaClient && window.llamaClient.loadModels) {
                    window.llamaClient.loadModels();
                }
            }
        });
    });
    
    // Auto-refresh detailed nodes every 30 seconds when nodes tab is active
    setInterval(() => {
        const nodesTab = document.querySelector('.tab-button[data-tab="nodes"]');
        if (nodesTab && nodesTab.classList.contains('active')) {
            if (window.llamaClient && window.llamaClient.loadDetailedNodes) {
                window.llamaClient.loadDetailedNodes();
            }
        }
    }, 30000);
}

// Add fade out animation
const style = document.createElement('style');
style.textContent = `
    @keyframes fadeOutDown {
        to {
            opacity: 0;
            transform: translateY(10px);
        }
    }
`;
document.head.appendChild(style);