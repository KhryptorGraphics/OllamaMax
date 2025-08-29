// WebSocket Service for Real-time Updates
// Handles WebSocket connections with automatic reconnection and event management

class WebSocketService {
  constructor() {
    this.connection = null;
    this.isConnected = false;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.reconnectDelay = 1000;
    this.eventListeners = new Map();
    this.messageQueue = [];
    this.heartbeatInterval = null;
    this.heartbeatTimeout = null;
    this.lastHeartbeat = null;
  }

  // Connect to WebSocket server
  connect(token = null) {
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsURL = `${wsProtocol}//${window.location.hostname}:11434/api/v1/ws`;

    try {
      this.connection = new WebSocket(wsURL);
      this.setupEventHandlers(token);
    } catch (error) {
      console.error('WebSocket connection failed:', error);
      this.handleConnectionError(error);
    }
  }

  // Setup WebSocket event handlers
  setupEventHandlers(token) {
    this.connection.onopen = () => {
      console.log('🔌 WebSocket connected');
      this.isConnected = true;
      this.reconnectAttempts = 0;
      
      // Send authentication if token provided
      if (token) {
        this.send({
          type: 'auth',
          token: token
        });
      }

      // Send any queued messages
      this.flushMessageQueue();
      
      // Start heartbeat
      this.startHeartbeat();
      
      // Emit connection event
      this.emit('connected');
    };

    this.connection.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        this.handleMessage(data);
      } catch (error) {
        console.error('WebSocket message parsing error:', error);
        this.emit('error', { type: 'parse_error', error });
      }
    };

    this.connection.onerror = (error) => {
      console.error('WebSocket error:', error);
      this.handleConnectionError(error);
    };

    this.connection.onclose = (event) => {
      console.log('🔌 WebSocket disconnected:', event.code, event.reason);
      this.isConnected = false;
      this.stopHeartbeat();
      
      this.emit('disconnected', {
        code: event.code,
        reason: event.reason,
        wasClean: event.wasClean
      });

      // Attempt reconnection if not a clean close
      if (event.code !== 1000 && this.reconnectAttempts < this.maxReconnectAttempts) {
        this.scheduleReconnect();
      }
    };
  }

  // Handle incoming messages
  handleMessage(data) {
    // Update last heartbeat time
    if (data.type === 'pong') {
      this.lastHeartbeat = Date.now();
      return;
    }

    // Handle authentication response
    if (data.type === 'auth_response') {
      if (data.success) {
        console.log('✅ WebSocket authenticated');
        this.emit('authenticated');
      } else {
        console.error('❌ WebSocket authentication failed:', data.error);
        this.emit('auth_failed', data.error);
      }
      return;
    }

    // Emit specific event based on message type
    this.emit(data.type, data);
    
    // Also emit generic message event
    this.emit('message', data);
  }

  // Handle connection errors
  handleConnectionError(error) {
    this.isConnected = false;
    this.emit('error', error);
  }

  // Schedule reconnection attempt
  scheduleReconnect() {
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts);
    
    setTimeout(() => {
      this.reconnectAttempts++;
      console.log(`🔄 Attempting WebSocket reconnection (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
      this.connect();
    }, delay);
  }

  // Send message to server
  send(message) {
    if (this.isConnected && this.connection.readyState === WebSocket.OPEN) {
      try {
        this.connection.send(JSON.stringify(message));
      } catch (error) {
        console.error('Failed to send WebSocket message:', error);
        this.queueMessage(message);
      }
    } else {
      this.queueMessage(message);
    }
  }

  // Queue message for later sending
  queueMessage(message) {
    this.messageQueue.push(message);
    
    // Limit queue size to prevent memory issues
    if (this.messageQueue.length > 100) {
      this.messageQueue.shift();
    }
  }

  // Send all queued messages
  flushMessageQueue() {
    while (this.messageQueue.length > 0) {
      const message = this.messageQueue.shift();
      this.send(message);
    }
  }

  // Start heartbeat mechanism
  startHeartbeat() {
    this.heartbeatInterval = setInterval(() => {
      if (this.isConnected) {
        this.send({ type: 'ping', timestamp: Date.now() });
        
        // Check if we received a pong recently
        if (this.lastHeartbeat && Date.now() - this.lastHeartbeat > 30000) {
          console.warn('⚠️ WebSocket heartbeat timeout');
          this.connection.close(1000, 'Heartbeat timeout');
        }
      }
    }, 15000); // Send ping every 15 seconds
  }

  // Stop heartbeat mechanism
  stopHeartbeat() {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
    
    if (this.heartbeatTimeout) {
      clearTimeout(this.heartbeatTimeout);
      this.heartbeatTimeout = null;
    }
  }

  // Event listener management
  on(event, callback) {
    if (!this.eventListeners.has(event)) {
      this.eventListeners.set(event, []);
    }
    this.eventListeners.get(event).push(callback);
  }

  off(event, callback) {
    if (this.eventListeners.has(event)) {
      const listeners = this.eventListeners.get(event);
      const index = listeners.indexOf(callback);
      if (index > -1) {
        listeners.splice(index, 1);
      }
    }
  }

  emit(event, data = null) {
    if (this.eventListeners.has(event)) {
      this.eventListeners.get(event).forEach(callback => {
        try {
          callback(data);
        } catch (error) {
          console.error(`Error in WebSocket event listener for '${event}':`, error);
        }
      });
    }
  }

  // Disconnect from server
  disconnect() {
    this.stopHeartbeat();
    
    if (this.connection) {
      this.connection.close(1000, 'Client disconnect');
      this.connection = null;
    }
    
    this.isConnected = false;
    this.reconnectAttempts = 0;
    this.messageQueue = [];
  }

  // Get connection status
  getStatus() {
    return {
      isConnected: this.isConnected,
      readyState: this.connection ? this.connection.readyState : WebSocket.CLOSED,
      reconnectAttempts: this.reconnectAttempts,
      queuedMessages: this.messageQueue.length,
      lastHeartbeat: this.lastHeartbeat
    };
  }

  // Subscribe to specific data streams
  subscribe(stream, callback) {
    this.on(stream, callback);
    
    // Send subscription request to server
    this.send({
      type: 'subscribe',
      stream: stream,
      timestamp: Date.now()
    });
  }

  // Unsubscribe from data streams
  unsubscribe(stream, callback = null) {
    if (callback) {
      this.off(stream, callback);
    } else {
      // Remove all listeners for this stream
      this.eventListeners.delete(stream);
    }
    
    // Send unsubscription request to server
    this.send({
      type: 'unsubscribe',
      stream: stream,
      timestamp: Date.now()
    });
  }

  // Request real-time metrics
  requestMetrics(types = ['system', 'cluster', 'models']) {
    this.send({
      type: 'request_metrics',
      metrics: types,
      timestamp: Date.now()
    });
  }

  // Request node status updates
  requestNodeUpdates() {
    this.send({
      type: 'request_node_updates',
      timestamp: Date.now()
    });
  }

  // Request inference status updates
  requestInferenceUpdates(requestId = null) {
    this.send({
      type: 'request_inference_updates',
      requestId: requestId,
      timestamp: Date.now()
    });
  }

  // Send user activity
  sendActivity(activity) {
    this.send({
      type: 'user_activity',
      activity: activity,
      timestamp: Date.now()
    });
  }
}

// Create and export singleton instance
const wsService = new WebSocketService();
export default wsService;

// Make it available globally for debugging
if (typeof window !== 'undefined') {
  window.wsService = wsService;
}
