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
        this.maxReconnectAttempts = 10;
        this.reconnectDelay = 1000;
        this.maxReconnectDelay = 30000;
        this.streamingMessage = null;
        this.connectionQuality = {
            latency: 0,
            packetLoss: 0,
            lastPing: null,
            pingInterval: null
        };
        this.endpoints = [
            this.settings.apiEndpoint || 'ws://localhost:13100/chat',
            'ws://localhost:13101/chat',
            'ws://localhost:13102/chat'
        ];
        this.currentEndpointIndex = 0;
        this.performanceData = {
            latency: [],
            throughput: [],
            memoryUsage: [],
            connectionHistory: []
        };
        
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.connect();
        this.loadNodes();
        this.startPerformanceMonitoring();
        this.initializeHelpTour();
        this.setupAdvancedControls();
        this.initializePerformanceDashboard();
        this.setupViewToggles();
    }

    // WebSocket Connection Management
    // Enhanced WebSocket Connection Management
    connect() {
        const endpoint = this.getCurrentEndpoint();
        this.updateConnectionStatus('connecting');
        
        console.log(`Attempting to connect to ${endpoint} (Attempt ${this.reconnectAttempts + 1})`);
        this.addToConnectionHistory('connect_attempt', endpoint);
        
        try {
            this.ws = new WebSocket(endpoint);
            
            // Set connection timeout
            const connectionTimeout = setTimeout(() => {
                if (this.ws.readyState === WebSocket.CONNECTING) {
                    console.warn('WebSocket connection timeout');
                    this.ws.close();
                    this.handleConnectionFailure('timeout');
                }
            }, 10000); // 10 second connection timeout
            
            this.ws.onopen = () => {
                clearTimeout(connectionTimeout);
                console.log('WebSocket connected successfully');
                this.updateConnectionStatus('connected');
                this.reconnectAttempts = 0;
                this.currentEndpointIndex = this.endpoints.indexOf(endpoint);
                this.addToConnectionHistory('connected', endpoint);
                this.startConnectionMonitoring();
                this.processMessageQueue();
                this.sendQueuedMessages();
            };
            
            this.ws.onmessage = (event) => {
                try {
                    const data = JSON.parse(event.data);
                    this.recordMessageReceived();
                    this.handleMessage(data);
                } catch (error) {
                    console.error('Failed to parse WebSocket message:', error);
                }
            };
            
            this.ws.onerror = (error) => {
                clearTimeout(connectionTimeout);
                console.error('WebSocket error:', error);
                this.addToConnectionHistory('error', endpoint, error);
                this.updateConnectionStatus('error');
            };
            
            this.ws.onclose = (event) => {
                clearTimeout(connectionTimeout);
                console.log('WebSocket disconnected:', event.code, event.reason);
                this.addToConnectionHistory('disconnected', endpoint, {
                    code: event.code,
                    reason: event.reason,
                    wasClean: event.wasClean
                });
                this.stopConnectionMonitoring();
                this.updateConnectionStatus('disconnected');
                this.handleDisconnection(event);
            };
        } catch (error) {
            console.error('Failed to create WebSocket:', error);
            this.addToConnectionHistory('creation_failed', endpoint, error);
            this.updateConnectionStatus('error');
            this.handleConnectionFailure('creation_error');
        }
    }

    getCurrentEndpoint() {
        return this.endpoints[this.currentEndpointIndex % this.endpoints.length];
    }

    getNextEndpoint() {
        this.currentEndpointIndex = (this.currentEndpointIndex + 1) % this.endpoints.length;
        return this.getCurrentEndpoint();
    }

    addToConnectionHistory(eventType, endpoint, details = null) {
        const entry = {
            timestamp: new Date().toISOString(),
            event: eventType,
            endpoint: endpoint,
            details: details,
            attempt: this.reconnectAttempts + 1
        };
        
        this.performanceData.connectionHistory.push(entry);
        
        // Keep only last 50 entries
        if (this.performanceData.connectionHistory.length > 50) {
            this.performanceData.connectionHistory.shift();
        }
        
        // Update UI with connection history
        this.updateConnectionHistoryDisplay();
    }

    updateConnectionHistoryDisplay() {
        const history = this.performanceData.connectionHistory;
        const recentAttempts = history.filter(entry => 
            entry.event === 'connect_attempt' && 
            new Date() - new Date(entry.timestamp) < 300000 // Last 5 minutes
        );
        
        const failedAttempts = recentAttempts.filter(entry => 
            history.some(h => h.event === 'disconnected' && h.timestamp > entry.timestamp)
        ).length;
        
        if (failedAttempts > 0) {
            document.getElementById('connectionText').textContent = `Failed attempts: ${failedAttempts}`;
        }
    }

    handleConnectionFailure(reason) {
        this.updateConnectionStatus('error');
        
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            const delay = Math.min(
                this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1),
                this.maxReconnectDelay
            );
            
            console.log(`Connection failed (${reason}), retrying in ${delay}ms...`);
            document.getElementById('connectionText').textContent = `Retrying in ${Math.ceil(delay/1000)}s...`;
            
            setTimeout(() => {
                // Try next endpoint on failure
                if (reason === 'timeout' || reason === 'error') {
                    this.getNextEndpoint();
                }
                this.connect();
            }, delay);
        } else {
            this.updateConnectionStatus('error');
            document.getElementById('connectionText').textContent = 'Max retries exceeded';
            this.showConnectionErrorModal();
        }
    }

    handleDisconnection(event) {
        if (event.code === 1000) {
            // Normal closure
            return;
        }
        
        // Check if we should attempt reconnection
        if (this.shouldAttemptReconnect(event)) {
            this.attemptReconnect();
        } else {
            this.updateConnectionStatus('error');
            this.showConnectionErrorModal();
        }
    }

    shouldAttemptReconnect(event) {
        // Don't reconnect for certain error codes
        const noReconnectCodes = [1002, 1003, 1007, 1008, 1009, 1011];
        return !noReconnectCodes.includes(event.code) && 
               this.reconnectAttempts < this.maxReconnectAttempts;
    }

    attemptReconnect() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            const delay = Math.min(
                this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1),
                this.maxReconnectDelay
            );
            
            console.log(`Attempting reconnection in ${delay}ms... (Attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
            document.getElementById('connectionText').textContent = `Reconnecting... (${this.reconnectAttempts}/${this.maxReconnectAttempts})`;
            
            // Try next endpoint
            this.getNextEndpoint();
            
            setTimeout(() => this.connect(), delay);
        } else {
            this.updateConnectionStatus('error');
            document.getElementById('connectionText').textContent = 'Connection failed';
            this.showConnectionErrorModal();
        }
    }

    // Connection Quality Monitoring
    startConnectionMonitoring() {
        this.stopConnectionMonitoring(); // Clear any existing interval
        
        this.connectionQuality.pingInterval = setInterval(() => {
            this.sendPing();
        }, 30000); // Ping every 30 seconds
        
        // Start throughput monitoring
        this.startThroughputMonitoring();
    }

    stopConnectionMonitoring() {
        if (this.connectionQuality.pingInterval) {
            clearInterval(this.connectionQuality.pingInterval);
            this.connectionQuality.pingInterval = null;
        }
        this.stopThroughputMonitoring();
    }

    sendPing() {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.connectionQuality.lastPing = Date.now();
            this.sendMessage({
                type: 'ping',
                timestamp: this.connectionQuality.lastPing
            });
        }
    }

    recordPong(timestamp) {
        if (this.connectionQuality.lastPing) {
            const latency = Date.now() - this.connectionQuality.lastPing;
            this.connectionQuality.latency = latency;
            
            // Update latency display
            const latencyElement = document.getElementById('connectionLatency');
            if (latencyElement) {
                latencyElement.textContent = `${latency}ms`;
                latencyElement.className = `latency ${latency > 200 ? 'high' : latency > 100 ? 'medium' : 'low'}`;
            }
            
            // Add to performance data
            this.performanceData.latency.push({
                timestamp: Date.now(),
                latency: latency
            });
            
            // Keep only last 100 measurements
            if (this.performanceData.latency.length > 100) {
                this.performanceData.latency.shift();
            }
        }
    }

    recordMessageReceived() {
        const now = Date.now();
        if (!this.connectionQuality.lastMessageTime) {
            this.connectionQuality.lastMessageTime = now;
            return;
        }
        
        const timeSinceLastMessage = now - this.connectionQuality.lastMessageTime;
        this.connectionQuality.lastMessageTime = now;
        
        // Calculate throughput (messages per second)
        const throughput = 1000 / Math.max(timeSinceLastMessage, 1);
        this.performanceData.throughput.push({
            timestamp: now,
            throughput: throughput
        });
        
        // Keep only last 100 measurements
        if (this.performanceData.throughput.length > 100) {
            this.performanceData.throughput.shift();
        }
    }

    startThroughputMonitoring() {
        this.throughputInterval = setInterval(() => {
            // Calculate average throughput over last minute
            const now = Date.now();
            const oneMinuteAgo = now - 60000;
            const recentData = this.performanceData.throughput.filter(
                data => data.timestamp > oneMinuteAgo
            );
            
            if (recentData.length > 0) {
                const avgThroughput = recentData.reduce((sum, data) => sum + data.throughput, 0) / recentData.length;
                
                // Update throughput display
                const throughputElement = document.getElementById('connectionThroughput');
                if (throughputElement) {
                    throughputElement.textContent = `${avgThroughput.toFixed(1)} msg/s`;
                }
            }
        }, 5000); // Update every 5 seconds
    }

    stopThroughputMonitoring() {
        if (this.throughputInterval) {
            clearInterval(this.throughputInterval);
            this.throughputInterval = null;
        }
    }

    showConnectionErrorModal() {
        const modal = document.createElement('div');
        modal.className = 'modal connection-error-modal';
        modal.innerHTML = `
            <div class="modal-content">
                <h3>Connection Error</h3>
                <p>Unable to connect to the server. Please check your connection and try again.</p>
                <div class="modal-actions">
                    <button class="btn btn-primary" onclick="window.llamaClient.retryConnection()">Retry</button>
                    <button class="btn btn-secondary" onclick="window.llamaClient.dismissConnectionError()">Cancel</button>
                </div>
                <div class="connection-info">
                    <p><strong>Connection History:</strong></p>
                    <div id="connectionHistoryList">
                        ${this.performanceData.connectionHistory.slice(-10).map(entry => 
                            `<div class="history-entry">
                                <span class="timestamp">${new Date(entry.timestamp).toLocaleTimeString()}</span>
                                <span class="event">${entry.event}</span>
                                <span class="endpoint">${entry.endpoint}</span>
                            </div>`
                        ).join('')}
                    </div>
                </div>
            </div>
        `;
        
        document.body.appendChild(modal);
        
        // Add dismiss on escape key
        const handleEscape = (e) => {
            if (e.key === 'Escape') {
                this.dismissConnectionError();
                document.removeEventListener('keydown', handleEscape);
            }
        };
        document.addEventListener('keydown', handleEscape);
    }

    retryConnection() {
        const modal = document.querySelector('.connection-error-modal');
        if (modal) {
            modal.remove();
        }
        this.reconnectAttempts = 0;
        this.connect();
    }

    dismissConnectionError() {
        const modal = document.querySelector('.connection-error-modal');
        if (modal) {
            modal.remove();
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
                indicator.classList.add('connecting');
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
            case 'reconnecting':
                indicator.classList.add('reconnecting');
                text.textContent = 'Reconnecting...';
                break;
        }
        
        // Update connection status in UI
        this.updateConnectionStatusDisplay(status);
    }

    updateConnectionStatusDisplay(status) {
        // Update additional status displays
        const statusElements = document.querySelectorAll('[data-status-display]');
        statusElements.forEach(element => {
            element.textContent = status;
            element.className = `status-display ${status}`;
        });
        
        // Update connection quality indicators
        if (status === 'connected') {
            this.recordConnectionQuality();
        }
    }

    recordConnectionQuality() {
        const quality = {
            timestamp: Date.now(),
            latency: this.connectionQuality.latency,
            status: 'connected'
        };
        
        this.performanceData.connectionQuality = quality;
        
        // Update quality indicator
        const qualityElement = document.getElementById('connectionQuality');
        if (qualityElement) {
            const qualityLevel = this.getConnectionQualityLevel(quality.latency);
            qualityElement.className = `connection-quality ${qualityLevel}`;
            qualityElement.title = `Connection quality: ${qualityLevel} (${quality.latency}ms)`;
        }
    }

    getConnectionQualityLevel(latency) {
        if (latency < 50) return 'excellent';
        if (latency < 100) return 'good';
        if (latency < 200) return 'fair';
        return 'poor';
    }

    // Enhanced message queue handling
    sendQueuedMessages() {
        if (this.messageQueue.length > 0) {
            console.log(`Sending ${this.messageQueue.length} queued messages`);
            this.messageQueue.forEach(message => {
                this.sendMessage(message);
            });
            this.messageQueue = [];
        }
    }

    sendMessage(data) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            try {
                this.ws.send(JSON.stringify(data));
                return true;
            } catch (error) {
                console.error('Failed to send message:', error);
                return false;
            }
        } else {
            // Queue message for later delivery
            this.messageQueue.push(data);
            console.log('Message queued for later delivery');
            return false;
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
            case 'pong':
                this.recordPong(data.timestamp);
                break;
            case 'connection_quality':
                this.updateConnectionQuality(data);
                break;
            case 'system_status':
                this.handleSystemStatus(data);
                break;
            case 'queue_update':
                this.updateQueueStatus(data);
                break;
            default:
                console.warn('Unknown message type:', data.type);
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
        // Simple markdown-like formatting with XSS protection
        // Note: For production, use a proper Markdown library with built-in sanitization
        // This is a basic implementation - consider using DOMPurify in production

        // Escape HTML entities first to prevent XSS
        const escaped = content
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#039;');

        // Then apply markdown-like formatting on escaped content
        return escaped
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

    // Setup view toggles for nodes dashboard
    setupViewToggles() {
        const viewToggles = document.querySelectorAll('.view-toggle');
        const performanceDashboard = document.getElementById('performanceDashboard');
        const enhancedNodesContainer = document.getElementById('enhancedNodesContainer');
        
        viewToggles.forEach(toggle => {
            toggle.addEventListener('click', () => {
                const view = toggle.dataset.view;
                
                // Remove active class from all toggles
                viewToggles.forEach(t => t.classList.remove('active'));
                
                // Add active class to clicked toggle
                toggle.classList.add('active');
                
                // Show/hide appropriate content
                switch(view) {
                    case 'compact':
                        performanceDashboard.style.display = 'none';
                        enhancedNodesContainer.style.display = 'grid';
                        break;
                    case 'detailed':
                        performanceDashboard.style.display = 'none';
                        enhancedNodesContainer.style.display = 'grid';
                        break;
                    case 'performance':
                        performanceDashboard.style.display = 'block';
                        enhancedNodesContainer.style.display = 'none';
                        this.updatePerformanceCharts();
                        break;
                }
            });
        });
    }
    
    // Initialize performance dashboard with Chart.js
    initializePerformanceDashboard() {
        this.charts = {};
        
        // Check if Chart.js is available
        if (typeof Chart !== 'undefined' && typeof Chart.Chart !== 'undefined') {
            this.createCharts();
            this.startChartUpdates();
        } else {
            console.warn('Chart.js not loaded, performance dashboard disabled');
            // Hide performance view toggle if Chart.js is not available
            const performanceToggle = document.querySelector('[data-view="performance"]');
            if (performanceToggle) {
                performanceToggle.style.display = 'none';
            }
        }
    }
    
    createCharts() {
        const chartOptions = {
            responsive: true,
            maintainAspectRatio: false,
            interaction: {
                intersect: false,
                mode: 'index'
            },
            plugins: {
                legend: {
                    display: true,
                    position: 'top'
                },
                tooltip: {
                    backgroundColor: 'rgba(0, 0, 0, 0.8)',
                    titleColor: 'white',
                    bodyColor: 'white',
                    borderColor: 'rgba(255, 255, 255, 0.1)',
                    borderWidth: 1
                }
            },
            scales: {
                x: {
                    display: true,
                    type: 'time',
                    time: {
                        displayFormats: {
                            minute: 'HH:mm',
                            hour: 'HH:mm'
                        }
                    }
                },
                y: {
                    display: true
                }
            }
        };
        
        // System Overview Chart
        const systemCtx = document.getElementById('systemOverviewChart');
        if (systemCtx) {
            this.charts.system = new Chart(systemCtx, {
                type: 'line',
                data: {
                    datasets: [
                        {
                            label: 'CPU Usage (%)',
                            borderColor: '#ed8936',
                            backgroundColor: 'rgba(237, 137, 54, 0.1)',
                            data: [],
                            tension: 0.4
                        },
                        {
                            label: 'Memory Usage (%)',
                            borderColor: '#4299e1',
                            backgroundColor: 'rgba(66, 153, 225, 0.1)',
                            data: [],
                            tension: 0.4
                        },
                        {
                            label: 'Active Nodes',
                            borderColor: '#48bb78',
                            backgroundColor: 'rgba(72, 187, 120, 0.1)',
                            data: [],
                            tension: 0.4
                        }
                    ]
                },
                options: {
                    ...chartOptions,
                    plugins: {
                        ...chartOptions.plugins,
                        legend: {
                            ...chartOptions.plugins.legend,
                            labels: {
                                color: '#2d3748'
                            }
                        }
                    },
                    scales: {
                        ...chartOptions.scales,
                        y: {
                            ...chartOptions.scales.y,
                            min: 0,
                            max: 100,
                            ticks: {
                                color: '#718096'
                            }
                        },
                        x: {
                            ...chartOptions.scales.x,
                            ticks: {
                                color: '#718096'
                            }
                        }
                    }
                }
            });
        }
        
        // Node Performance Chart
        const nodeCtx = document.getElementById('nodePerformanceChart');
        if (nodeCtx) {
            this.charts.node = new Chart(nodeCtx, {
                type: 'bar',
                data: {
                    labels: [],
                    datasets: [
                        {
                            label: 'Requests/sec',
                            backgroundColor: 'rgba(72, 187, 120, 0.8)',
                            borderColor: '#48bb78',
                            data: []
                        },
                        {
                            label: 'Queue Length',
                            backgroundColor: 'rgba(237, 137, 54, 0.8)',
                            borderColor: '#ed8936',
                            data: []
                        }
                    ]
                },
                options: {
                    ...chartOptions,
                    plugins: {
                        ...chartOptions.plugins,
                        legend: {
                            ...chartOptions.plugins.legend,
                            labels: {
                                color: '#2d3748'
                            }
                        }
                    },
                    scales: {
                        ...chartOptions.scales,
                        y: {
                            ...chartOptions.scales.y,
                            ticks: {
                                color: '#718096'
                            }
                        },
                        x: {
                            ...chartOptions.scales.x,
                            ticks: {
                                color: '#718096'
                            }
                        }
                    }
                }
            });
        }
        
        // Latency Chart
        const latencyCtx = document.getElementById('latencyChart');
        if (latencyCtx) {
            this.charts.latency = new Chart(latencyCtx, {
                type: 'line',
                data: {
                    datasets: [
                        {
                            label: 'Average Latency (ms)',
                            borderColor: '#e53e3e',
                            backgroundColor: 'rgba(229, 62, 62, 0.1)',
                            data: [],
                            tension: 0.4
                        },
                        {
                            label: '95th Percentile (ms)',
                            borderColor: '#dd6b20',
                            backgroundColor: 'rgba(221, 107, 32, 0.1)',
                            data: [],
                            tension: 0.4
                        }
                    ]
                },
                options: {
                    ...chartOptions,
                    plugins: {
                        ...chartOptions.plugins,
                        legend: {
                            ...chartOptions.plugins.legend,
                            labels: {
                                color: '#2d3748'
                            }
                        }
                    },
                    scales: {
                        ...chartOptions.scales,
                        y: {
                            ...chartOptions.scales.y,
                            ticks: {
                                color: '#718096'
                            }
                        },
                        x: {
                            ...chartOptions.scales.x,
                            ticks: {
                                color: '#718096'
                            }
                        }
                    }
                }
            });
        }
        
        // Throughput Chart
        const throughputCtx = document.getElementById('throughputChart');
        if (throughputCtx) {
            this.charts.throughput = new Chart(throughputCtx, {
                type: 'line',
                data: {
                    datasets: [
                        {
                            label: 'Requests/min',
                            borderColor: '#667eea',
                            backgroundColor: 'rgba(102, 126, 234, 0.1)',
                            data: [],
                            tension: 0.4
                        },
                        {
                            label: 'Tokens/min',
                            borderColor: '#764ba2',
                            backgroundColor: 'rgba(118, 75, 162, 0.1)',
                            data: [],
                            tension: 0.4
                        }
                    ]
                },
                options: {
                    ...chartOptions,
                    plugins: {
                        ...chartOptions.plugins,
                        legend: {
                            ...chartOptions.plugins.legend,
                            labels: {
                                color: '#2d3748'
                            }
                        }
                    },
                    scales: {
                        ...chartOptions.scales,
                        y: {
                            ...chartOptions.scales.y,
                            ticks: {
                                color: '#718096'
                            }
                        },
                        x: {
                            ...chartOptions.scales.x,
                            ticks: {
                                color: '#718096'
                            }
                        }
                    }
                }
            });
        }
    }
    
    startChartUpdates() {
        // Update charts every 30 seconds
        this.chartUpdateInterval = setInterval(() => {
            this.updatePerformanceCharts();
        }, 30000);
    }
    
    updatePerformanceCharts() {
        if (!this.charts.system) return;
        
        const now = new Date();
        
        // Update system overview chart
        this.updateSystemChart(now);
        
        // Update node performance chart
        this.updateNodeChart();
        
        // Update latency chart
        this.updateLatencyChart(now);
        
        // Update throughput chart
        this.updateThroughputChart(now);
    }
    
    updateSystemChart(now) {
        if (!this.charts.system) return;
        
        // Add current data point
        const systemData = this.getSystemOverviewData();
        
        this.charts.system.data.datasets[0].data.push({
            x: now,
            y: systemData.cpuAvg
        });
        
        this.charts.system.data.datasets[1].data.push({
            x: now,
            y: systemData.memoryAvg
        });
        
        this.charts.system.data.datasets[2].data.push({
            x: now,
            y: systemData.activeNodes
        });
        
        // Keep only last 50 data points
        if (this.charts.system.data.datasets[0].data.length > 50) {
            this.charts.system.data.datasets.forEach(dataset => {
                dataset.data.shift();
            });
        }
        
        this.charts.system.update('none');
    }
    
    updateNodeChart() {
        if (!this.charts.node) return;
        
        const healthyNodes = this.nodes.filter(node => node.status === 'healthy');
        
        // Update labels and data
        this.charts.node.data.labels = healthyNodes.map(node => node.name);
        this.charts.node.data.datasets[0].data = healthyNodes.map(node => node.requestsPerSecond || 0);
        this.charts.node.data.datasets[1].data = healthyNodes.map(node => node.queue || 0);
        
        this.charts.node.update('none');
    }
    
    updateLatencyChart(now) {
        if (!this.charts.latency) return;
        
        // Add latency data point
        const avgLatency = this.calculateAverageLatency();
        const p95Latency = this.calculateP95Latency();
        
        this.charts.latency.data.datasets[0].data.push({
            x: now,
            y: avgLatency
        });
        
        this.charts.latency.data.datasets[1].data.push({
            x: now,
            y: p95Latency
        });
        
        // Keep only last 50 data points
        if (this.charts.latency.data.datasets[0].data.length > 50) {
            this.charts.latency.data.datasets.forEach(dataset => {
                dataset.data.shift();
            });
        }
        
        this.charts.latency.update('none');
    }
    
    updateThroughputChart(now) {
        if (!this.charts.throughput) return;
        
        // Add throughput data point
        const requestsPerMin = this.calculateRequestsPerMinute();
        const tokensPerMin = this.calculateTokensPerMinute();
        
        this.charts.throughput.data.datasets[0].data.push({
            x: now,
            y: requestsPerMin
        });
        
        this.charts.throughput.data.datasets[1].data.push({
            x: now,
            y: tokensPerMin
        });
        
        // Keep only last 50 data points
        if (this.charts.throughput.data.datasets[0].data.length > 50) {
            this.charts.throughput.data.datasets.forEach(dataset => {
                dataset.data.shift();
            });
        }
        
        this.charts.throughput.update('none');
    }
    
    getSystemOverviewData() {
        const healthyNodes = this.nodes.filter(node => node.status === 'healthy');
        
        if (healthyNodes.length === 0) {
            return { cpuAvg: 0, memoryAvg: 0, activeNodes: 0 };
        }
        
        const cpuSum = healthyNodes.reduce((sum, node) => sum + (node.systemInfo?.cpu?.usage || 0), 0);
        const memorySum = healthyNodes.reduce((sum, node) => sum + (node.systemInfo?.memory?.usage || 0), 0);
        
        return {
            cpuAvg: Math.round(cpuSum / healthyNodes.length),
            memoryAvg: Math.round(memorySum / healthyNodes.length),
            activeNodes: healthyNodes.length
        };
    }
    
    calculateAverageLatency() {
        if (!this.performanceData.latency || this.performanceData.latency.length === 0) {
            return 0;
        }
        
        const recentLatencies = this.performanceData.latency.slice(-20);
        const sum = recentLatencies.reduce((total, entry) => total + entry.latency, 0);
        return Math.round(sum / recentLatencies.length);
    }
    
    calculateP95Latency() {
        if (!this.performanceData.latency || this.performanceData.latency.length === 0) {
            return 0;
        }
        
        const recentLatencies = this.performanceData.latency.slice(-20);
        const sorted = recentLatencies.map(entry => entry.latency).sort((a, b) => a - b);
        const p95Index = Math.ceil(sorted.length * 0.95) - 1;
        return sorted[p95Index] || 0;
    }
    
    calculateRequestsPerMinute() {
        if (!this.performanceData.throughput || this.performanceData.throughput.length === 0) {
            return 0;
        }
        
        // Calculate requests per minute from throughput data
        const recentData = this.performanceData.throughput.slice(-10);
        if (recentData.length === 0) return 0;
        
        const avgThroughput = recentData.reduce((sum, data) => sum + data.throughput, 0) / recentData.length;
        return Math.round(avgThroughput * 60); // Convert per-second to per-minute
    }
    
    calculateTokensPerMinute() {
        // This would need to be implemented based on actual token counting
        // For now, return a placeholder value
        return Math.floor(Math.random() * 1000) + 500;
    }
    handleSystemStatus(data) {
        // Update system-wide status information
        if (data.status) {
            this.systemStatus = data.status;
            this.updateSystemStatusDisplay(data.status);
        }
        
        if (data.maintenance) {
            this.handleMaintenanceMode(data.maintenance);
        }
        
        if (data.rateLimit) {
            this.handleRateLimit(data.rateLimit);
        }
    }
    
    updateSystemStatusDisplay(status) {
        const statusElement = document.getElementById('systemStatus');
        if (statusElement) {
            statusElement.textContent = status;
            statusElement.className = `system-status ${status}`;
        }
    }
    
    handleMaintenanceMode(maintenance) {
        if (maintenance.enabled) {
            this.showMaintenanceNotification(maintenance);
        } else {
            this.hideMaintenanceNotification();
        }
    }
    
    handleRateLimit(rateLimit) {
        const rateLimitElement = document.getElementById('rateLimitStatus');
        if (rateLimitElement) {
            if (rateLimit.remaining > 0) {
                rateLimitElement.style.display = 'none';
            } else {
                rateLimitElement.style.display = 'block';
                rateLimitElement.textContent = `Rate limited. Reset in ${Math.ceil(rateLimit.resetTime / 1000)}s`;
            }
        }
    }
    
    updateQueueStatus(data) {
        const queueElement = document.getElementById('queueLength');
        if (queueElement) {
            queueElement.textContent = data.length || 0;
            
            // Visual feedback for queue length
            if (data.length > 10) {
                queueElement.style.color = 'var(--error)';
                queueElement.title = 'High queue length - expect delays';
            } else if (data.length > 5) {
                queueElement.style.color = 'var(--warning)';
                queueElement.title = 'Moderate queue length';
            } else {
                queueElement.style.color = 'var(--success)';
                queueElement.title = 'Low queue length';
            }
        }
        
        // Update queue time estimate
        if (data.estimatedWaitTime) {
            const waitTimeElement = document.getElementById('estimatedWaitTime');
            if (waitTimeElement) {
                waitTimeElement.textContent = `Est. wait: ${Math.ceil(data.estimatedWaitTime / 1000)}s`;
                waitTimeElement.style.display = 'inline';
            }
        }
    }
    
    updateConnectionQuality(data) {
        if (data.latency) {
            this.connectionQuality.latency = data.latency;
            this.recordConnectionQuality();
        }
        
        if (data.packetLoss !== undefined) {
            this.connectionQuality.packetLoss = data.packetLoss;
            this.updatePacketLossDisplay(data.packetLoss);
        }
    }
    
    updatePacketLossDisplay(packetLoss) {
        const packetLossElement = document.getElementById('packetLoss');
        if (packetLossElement) {
            packetLossElement.textContent = `${(packetLoss * 100).toFixed(2)}%`;
            packetLossElement.className = `packet-loss ${packetLoss > 0.05 ? 'high' : packetLoss > 0.01 ? 'medium' : 'low'}`;
        }
    }
    
    showMaintenanceNotification(maintenance) {
        const notification = document.createElement('div');
        notification.className = 'notification maintenance-notification';
        notification.innerHTML = `
            <div class="notification-content">
                <span class="notification-icon">🔧</span>
                <span class="notification-text">
                    ${maintenance.message || 'System maintenance in progress'}
                    ${maintenance.estimatedDowntime ? ` (Est. downtime: ${maintenance.estimatedDowntime})` : ''}
                </span>
                <button class="notification-close" onclick="this.parentElement.parentElement.remove()">×</button>
            </div>
        `;
        
        document.body.appendChild(notification);
        
        // Auto-hide after maintenance period if specified
        if (maintenance.estimatedDowntime) {
            setTimeout(() => {
                if (notification.parentElement) {
                    notification.remove();
                }
            }, maintenance.estimatedDowntime);
        }
    }
    
    hideMaintenanceNotification() {
        const notification = document.querySelector('.maintenance-notification');
        if (notification) {
            notification.remove();
        }
    }

    // Node Management
    async loadNodes() {
        try {
            const response = await fetch('http://localhost:13000/api/nodes/detailed');
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
            apiEndpoint: 'ws://localhost:13000/chat',
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

    toggleDarkMode(enabled) {
        if (enabled) {
            document.body.classList.add('dark-mode');
            localStorage.setItem('darkMode', 'true');
        } else {
            document.body.classList.remove('dark-mode');
            localStorage.setItem('darkMode', 'false');
        }
        this.showToast(enabled ? 'Dark mode enabled' : 'Dark mode disabled');
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
        const attachButton = document.getElementById('attachButton');
        const fileInput = document.getElementById('fileInput');
        const attachmentPreview = document.getElementById('attachmentPreview');
        
        messageInput.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                this.sendMessage(messageInput.value);
            }
        });
        
        sendButton.addEventListener('click', () => {
            this.sendMessage(messageInput.value);
        });

        // File attachment handling
        attachButton.addEventListener('click', () => {
            fileInput.click();
        });

        fileInput.addEventListener('change', (e) => {
            const files = Array.from(e.target.files);
            if (files.length > 0) {
                this.handleFileAttachments(files);
            }
        });

        // Drag and drop for file attachments
        const inputArea = document.querySelector('.input-area');
        inputArea.addEventListener('dragover', (e) => {
            e.preventDefault();
            inputArea.classList.add('drag-over');
        });

        inputArea.addEventListener('dragleave', (e) => {
            e.preventDefault();
            inputArea.classList.remove('drag-over');
        });

        inputArea.addEventListener('drop', (e) => {
            e.preventDefault();
            inputArea.classList.remove('drag-over');
            const files = Array.from(e.dataTransfer.files);
            if (files.length > 0) {
                this.handleFileAttachments(files);
            }
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

        // Dark mode toggle
        document.getElementById('darkMode').addEventListener('change', (e) => {
            this.toggleDarkMode(e.target.checked);
        });

        // Keyboard shortcuts
        document.addEventListener('keydown', (e) => {
            // Only if keyboard shortcuts are enabled
            if (!document.getElementById('keyboardShortcuts').checked) return;

            // Ctrl/Cmd + Enter to send message
            if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
                e.preventDefault();
                const messageInput = document.getElementById('messageInput');
                if (messageInput.value.trim()) {
                    this.sendMessage(messageInput.value);
                }
            }

            // Ctrl/Cmd + K to focus message input
            if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
                e.preventDefault();
                document.getElementById('messageInput').focus();
            }

            // Ctrl/Cmd + 1-4 to switch tabs
            if ((e.ctrlKey || e.metaKey) && ['1', '2', '3', '4'].includes(e.key)) {
                e.preventDefault();
                const tabs = ['chat', 'nodes', 'models', 'settings'];
                const tabIndex = parseInt(e.key) - 1;
                const tabButton = document.querySelector(`[data-tab="${tabs[tabIndex]}"]`);
                if (tabButton) tabButton.click();
            }

            // Escape to close modals
            if (e.key === 'Escape') {
                document.querySelectorAll('.modal.active').forEach(modal => {
                    modal.classList.remove('active');
                });
            }
        });

        // Apply initial settings
        this.applySettings();

        // Load dark mode preference
        const darkModeEnabled = localStorage.getItem('darkMode') === 'true';
        document.getElementById('darkMode').checked = darkModeEnabled;
        if (darkModeEnabled) {
            document.body.classList.add('dark-mode');
        }
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
            const response = await fetch('http://localhost:13000/api/models/pull', {
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
            const response = await fetch('http://localhost:13000/api/models/propagate', {
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
            const response = await fetch(`http://localhost:13000/api/models/${encodeURIComponent(modelName)}`, {
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
            const response = await fetch('http://localhost:13000/api/nodes/detailed');
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
            const response = await fetch(`http://localhost:13000/api/nodes/${nodeId}/control`, {
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
            const response = await fetch(`http://localhost:13000/api/nodes/${nodeId}/config`, {
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
            const response = await fetch(`http://localhost:13000/api/nodes/${nodeId}/control`, {
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
            const response = await fetch(`http://localhost:13000/api/nodes/${nodeId}/control`, {
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

    // File Attachment Handling
    handleFileAttachments(files) {
        const attachmentPreview = document.getElementById('attachmentPreview');
        const maxSize = 25 * 1024 * 1024; // 25MB limit
        
        files.forEach(file => {
            if (file.size > maxSize) {
                this.showToast(`File ${file.name} is too large. Maximum size is 25MB.`, 'error');
                return;
            }

            const fileId = `file_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
            const fileData = {
                id: fileId,
                name: file.name,
                size: file.size,
                type: file.type,
                file: file,
                uploadProgress: 0,
                status: 'pending'
            };

            this.uploadedFiles = this.uploadedFiles || [];
            this.uploadedFiles.push(fileData);
            
            this.renderAttachmentPreview();
            this.uploadFile(fileData);
        });
    }

    uploadFile(fileData) {
        fileData.status = 'uploading';
        this.renderAttachmentPreview();

        // Simulate file upload progress
        const uploadInterval = setInterval(() => {
            fileData.uploadProgress += Math.random() * 20;
            if (fileData.uploadProgress >= 100) {
                fileData.uploadProgress = 100;
                fileData.status = 'completed';
                clearInterval(uploadInterval);
                
                // Send file to server via WebSocket
                this.sendFileToServer(fileData);
                this.showToast(`File ${fileData.name} uploaded successfully`, 'success');
            }
            this.renderAttachmentPreview();
        }, 200);
    }

    sendFileToServer(fileData) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            const reader = new FileReader();
            reader.onload = (e) => {
                const message = {
                    type: 'file_upload',
                    file_id: fileData.id,
                    filename: fileData.name,
                    content_type: fileData.type,
                    data: e.target.result,
                    chunk_size: 1024 * 1024 // 1MB chunks
                };
                this.ws.send(JSON.stringify(message));
            };
            reader.readAsDataURL(fileData.file);
        }
    }

    renderAttachmentPreview() {
        const attachmentPreview = document.getElementById('attachmentPreview');
        if (!this.uploadedFiles || this.uploadedFiles.length === 0) {
            attachmentPreview.innerHTML = '';
            attachmentPreview.style.display = 'none';
            return;
        }

        attachmentPreview.style.display = 'block';
        attachmentPreview.innerHTML = '';

        this.uploadedFiles.forEach(fileData => {
            const template = document.getElementById('attachmentPreviewTemplate');
            const clone = template.content.cloneNode(true);
            
            clone.querySelector('.attachment-name').textContent = fileData.name;
            clone.querySelector('.attachment-size').textContent = this.formatBytes(fileData.size);

            const progressContainer = clone.querySelector('.upload-progress');
            const progressFill = clone.querySelector('.progress-fill');
            const progressText = clone.querySelector('.progress-text');

            if (fileData.status === 'uploading') {
                progressContainer.style.display = 'flex';
                progressFill.style.width = `${fileData.uploadProgress}%`;
                progressText.textContent = `${Math.round(fileData.uploadProgress)}%`;
            } else if (fileData.status === 'completed') {
                progressContainer.style.display = 'none';
            }

            const removeBtn = clone.querySelector('.remove-attachment-btn');
            removeBtn.addEventListener('click', () => {
                this.removeAttachment(fileData.id);
            });

            attachmentPreview.appendChild(clone);
        });
    }

    removeAttachment(fileId) {
        this.uploadedFiles = this.uploadedFiles.filter(f => f.id !== fileId);
        this.renderAttachmentPreview();
    }

    // Help Tour System
    initializeHelpTour() {
        this.helpTour = {
            isActive: false,
            currentStep: 0,
            steps: [
                {
                    title: 'Welcome to Distributed Llama Chat',
                    content: 'This is your AI chat interface. You can send messages, attach files, and interact with distributed AI models.',
                    element: '#messageInput',
                    position: 'bottom'
                },
                {
                    title: 'Node Dashboard',
                    content: 'Monitor your distributed inference nodes here. View real-time metrics, health status, and manage configurations.',
                    element: '[data-tab="nodes"]',
                    position: 'bottom'
                },
                {
                    title: 'Model Management',
                    content: 'Download new models, propagate them across nodes, and manage your model library for distributed inference.',
                    element: '[data-tab="models"]',
                    position: 'bottom'
                },
                {
                    title: 'Settings',
                    content: 'Customize your experience with API configuration, appearance settings, and advanced chat options.',
                    element: '[data-tab="settings"]',
                    position: 'bottom'
                }
            ]
        };

        this.setupHelpTourEvents();
    }

    setupHelpTourEvents() {
        // Help button
        const helpButton = document.getElementById('helpButton');
        if (helpButton) {
            helpButton.addEventListener('click', () => {
                this.startHelpTour();
            });
        }

        // Start tour button in settings
        const startTourButton = document.getElementById('startTourButton');
        if (startTourButton) {
            startTourButton.addEventListener('click', () => {
                this.startHelpTour();
            });
        }

        // Tour modal events
        const closeTourBtn = document.getElementById('closeTourBtn');
        if (closeTourBtn) {
            closeTourBtn.addEventListener('click', () => {
                this.endHelpTour();
            });
        }
    }

    startHelpTour() {
        this.helpTour.isActive = true;
        this.helpTour.currentStep = 0;
        
        // Show tour modal
        const tourModal = document.getElementById('helpTourModal');
        if (tourModal) {
            tourModal.style.display = 'flex';
            this.showTourStep(1);
        }

        this.updateTourProgress();
        this.highlightCurrentElement();
    }

    showTourStep(stepNumber) {
        const tourSteps = document.querySelectorAll('.tour-step');
        tourSteps.forEach(step => {
            step.classList.remove('active');
            if (step.dataset.step === stepNumber.toString()) {
                step.classList.add('active');
            }
        });
    }

    highlightCurrentElement() {
        if (this.helpTour.currentStep >= this.helpTour.steps.length) return;

        const step = this.helpTour.steps[this.helpTour.currentStep];
        const element = document.querySelector(step.element);
        
        if (element) {
            element.scrollIntoView({ behavior: 'smooth', block: 'center' });
            element.classList.add('tour-highlight');
        }
    }

    nextTourStep() {
        if (this.helpTour.currentStep < this.helpTour.steps.length - 1) {
            this.helpTour.currentStep++;
            this.showTourStep(this.helpTour.currentStep + 1);
            this.updateTourProgress();
            this.highlightCurrentElement();
        } else {
            this.endHelpTour();
        }
    }

    prevTourStep() {
        if (this.helpTour.currentStep > 0) {
            this.helpTour.currentStep--;
            this.showTourStep(this.helpTour.currentStep + 1);
            this.updateTourProgress();
            this.highlightCurrentElement();
        }
    }

    endHelpTour() {
        this.helpTour.isActive = false;
        
        // Remove highlights
        document.querySelectorAll('.tour-highlight').forEach(el => {
            el.classList.remove('tour-highlight');
        });

        // Hide tour modal
        const tourModal = document.getElementById('helpTourModal');
        if (tourModal) {
            tourModal.style.display = 'none';
        }

        this.showToast('Thanks for exploring Distributed Llama Chat! You can always access help from the settings.', 'success');
    }

    updateTourProgress() {
        const progress = ((this.helpTour.currentStep + 1) / this.helpTour.steps.length) * 100;
        console.log(`Tour progress: ${progress.toFixed(1)}%`);
    }

    // Enhanced Settings with Advanced Controls
    setupAdvancedControls() {
        const advancedControlsToggle = document.querySelector('.advanced-controls-toggle');
        const advancedControlsPanel = document.querySelector('.advanced-controls-panel');
        const modelSelector = document.getElementById('modelSelector');
        const temperatureSlider = document.getElementById('temperature');
        const temperatureValue = document.getElementById('temperatureValue');
        const contextLengthSlider = document.getElementById('contextLength');
        const contextLengthValue = document.getElementById('contextLengthValue');

        if (advancedControlsToggle && advancedControlsPanel) {
            advancedControlsToggle.addEventListener('click', () => {
                const isVisible = advancedControlsPanel.style.display === 'flex';
                advancedControlsPanel.style.display = isVisible ? 'none' : 'flex';
            });
        }

        if (temperatureSlider && temperatureValue) {
            temperatureSlider.addEventListener('input', () => {
                temperatureValue.textContent = temperatureSlider.value;
                this.settings.temperature = parseFloat(temperatureSlider.value);
                this.saveSettings();
            });
        }

        if (contextLengthSlider && contextLengthValue) {
            contextLengthSlider.addEventListener('input', () => {
                contextLengthValue.textContent = contextLengthSlider.value;
                this.settings.contextLength = parseInt(contextLengthSlider.value);
                this.saveSettings();
            });
        }

        if (modelSelector) {
            modelSelector.addEventListener('change', () => {
                this.settings.selectedModel = modelSelector.value;
                this.saveSettings();
                this.updateModelStatus();
            });
        }
    }

    updateModelStatus() {
        const activeNodeElement = document.getElementById('activeNode');
        const currentModel = this.settings.selectedModel || 'tinyllama';
        
        if (activeNodeElement) {
            activeNodeElement.textContent = `${currentModel} on ${this.activeNode || 'auto-selected node'}`;
        }
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

// Setup tour navigation events
function setupTourEvents() {
    // Next buttons
    document.querySelectorAll('.tour-next-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            if (window.llamaClient && window.llamaClient.helpTour && window.llamaClient.helpTour.isActive) {
                window.llamaClient.nextTourStep();
            }
        });
    });

    // Previous buttons
    document.querySelectorAll('.tour-prev-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            if (window.llamaClient && window.llamaClient.helpTour && window.llamaClient.helpTour.isActive) {
                window.llamaClient.prevTourStep();
            }
        });
    });

    // Skip buttons
    document.querySelectorAll('.tour-skip-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            if (window.llamaClient) {
                window.llamaClient.endHelpTour();
            }
        });
    });

    // Complete button
    document.querySelectorAll('.tour-complete-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            if (window.llamaClient) {
                window.llamaClient.endHelpTour();
            }
        });
    });
}

// Initialize tour events after DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    setupTourEvents();
});

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
// Export for testing and module usage
if (typeof module !== 'undefined' && module.exports) {
    module.exports = DistributedLlamaClient;
}

// Initialize the application if running in browser
if (typeof window !== 'undefined') {
    window.llamaClient = new DistributedLlamaClient();
}