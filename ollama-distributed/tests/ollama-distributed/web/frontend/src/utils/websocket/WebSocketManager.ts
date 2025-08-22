export interface WebSocketConfig {
  url: string;
  protocols?: string[];
  reconnectInterval?: number;
  maxReconnectAttempts?: number;
  heartbeatInterval?: number;
  debug?: boolean;
  auth?: {
    type: 'bearer' | 'api-key' | 'custom';
    token?: string;
    apiKey?: string;
    headers?: Record<string, string>;
  };
}

export interface WebSocketMessage {
  id: string;
  type: string;
  data: any;
  timestamp: Date;
  correlationId?: string;
}

export interface WebSocketSubscription {
  id: string;
  topic: string;
  callback: (message: WebSocketMessage) => void;
  filter?: (message: WebSocketMessage) => boolean;
}

export interface WebSocketMetrics {
  connected: boolean;
  connectionTime: Date | null;
  lastMessage: Date | null;
  messagesSent: number;
  messagesReceived: number;
  reconnectAttempts: number;
  bytesTransmitted: number;
  bytesReceived: number;
  averageLatency: number;
  errorCount: number;
  lastError: string | null;
}

export interface StreamConfig {
  bufferSize: number;
  flushInterval: number;
  compression: boolean;
  batchMessages: boolean;
  maxBatchSize: number;
  retryFailedMessages: boolean;
  messageTimeout: number;
}

export type WebSocketEventType = 
  | 'connect' 
  | 'disconnect' 
  | 'message' 
  | 'error' 
  | 'reconnecting' 
  | 'reconnected'
  | 'heartbeat'
  | 'subscription-added'
  | 'subscription-removed';

export type WebSocketEventCallback = (event: WebSocketEvent) => void;

export interface WebSocketEvent {
  type: WebSocketEventType;
  data?: any;
  error?: Error;
  timestamp: Date;
}

export class WebSocketManager {
  private ws: WebSocket | null = null;
  private config: WebSocketConfig;
  private subscriptions: Map<string, WebSocketSubscription> = new Map();
  private eventListeners: Map<WebSocketEventType, Set<WebSocketEventCallback>> = new Map();
  private messageQueue: WebSocketMessage[] = [];
  private reconnectAttempts = 0;
  private reconnectTimer: NodeJS.Timeout | null = null;
  private heartbeatTimer: NodeJS.Timeout | null = null;
  private metrics: WebSocketMetrics;
  private streamConfig: StreamConfig;
  private messageBuffer: WebSocketMessage[] = [];
  private flushTimer: NodeJS.Timeout | null = null;
  private pendingMessages: Map<string, { message: WebSocketMessage; resolve: Function; reject: Function; timeout: NodeJS.Timeout }> = new Map();

  constructor(config: WebSocketConfig, streamConfig?: Partial<StreamConfig>) {
    this.config = {
      reconnectInterval: 5000,
      maxReconnectAttempts: 10,
      heartbeatInterval: 30000,
      debug: false,
      ...config
    };

    this.streamConfig = {
      bufferSize: 100,
      flushInterval: 1000,
      compression: false,
      batchMessages: true,
      maxBatchSize: 10,
      retryFailedMessages: true,
      messageTimeout: 10000,
      ...streamConfig
    };

    this.metrics = {
      connected: false,
      connectionTime: null,
      lastMessage: null,
      messagesSent: 0,
      messagesReceived: 0,
      reconnectAttempts: 0,
      bytesTransmitted: 0,
      bytesReceived: 0,
      averageLatency: 0,
      errorCount: 0,
      lastError: null
    };

    this.initializeEventListeners();
  }

  private initializeEventListeners(): void {
    Object.values(['connect', 'disconnect', 'message', 'error', 'reconnecting', 'reconnected', 'heartbeat', 'subscription-added', 'subscription-removed'] as WebSocketEventType[]).forEach(type => {
      this.eventListeners.set(type, new Set());
    });
  }

  public connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
          resolve();
          return;
        }

        this.log('Connecting to WebSocket...', this.config.url);

        const url = this.buildConnectionUrl();
        this.ws = new WebSocket(url, this.config.protocols);

        this.ws.onopen = (event) => {
          this.log('WebSocket connected');
          this.metrics.connected = true;
          this.metrics.connectionTime = new Date();
          this.reconnectAttempts = 0;
          
          this.startHeartbeat();
          this.processQueuedMessages();
          this.emit('connect', { connected: true });
          
          resolve();
        };

        this.ws.onmessage = (event) => {
          this.handleMessage(event);
        };

        this.ws.onclose = (event) => {
          this.handleDisconnect(event);
        };

        this.ws.onerror = (event) => {
          this.handleError(new Error('WebSocket error'));
          reject(new Error('Failed to connect to WebSocket'));
        };

      } catch (error) {
        this.handleError(error as Error);
        reject(error);
      }
    });
  }

  private buildConnectionUrl(): string {
    let url = this.config.url;
    
    if (this.config.auth) {
      const params = new URLSearchParams();
      
      switch (this.config.auth.type) {
        case 'bearer':
          if (this.config.auth.token) {
            params.append('token', this.config.auth.token);
          }
          break;
        case 'api-key':
          if (this.config.auth.apiKey) {
            params.append('api_key', this.config.auth.apiKey);
          }
          break;
      }
      
      if (params.toString()) {
        url += (url.includes('?') ? '&' : '?') + params.toString();
      }
    }
    
    return url;
  }

  public disconnect(): void {
    this.log('Disconnecting WebSocket...');
    
    this.stopHeartbeat();
    this.stopReconnectTimer();
    this.stopFlushTimer();
    
    if (this.ws) {
      this.ws.close(1000, 'Client disconnect');
      this.ws = null;
    }
    
    this.metrics.connected = false;
    this.metrics.connectionTime = null;
    this.emit('disconnect', { reason: 'client_disconnect' });
  }

  private handleMessage(event: MessageEvent): void {
    try {
      this.metrics.messagesReceived++;
      this.metrics.bytesReceived += event.data.length;
      this.metrics.lastMessage = new Date();

      const message: WebSocketMessage = JSON.parse(event.data);
      message.timestamp = new Date();

      this.log('Received message:', message);

      // Handle heartbeat responses
      if (message.type === 'heartbeat') {
        this.emit('heartbeat', message.data);
        return;
      }

      // Handle message responses
      if (message.correlationId && this.pendingMessages.has(message.correlationId)) {
        const pending = this.pendingMessages.get(message.correlationId)!;
        clearTimeout(pending.timeout);
        this.pendingMessages.delete(message.correlationId);
        pending.resolve(message);
        return;
      }

      // Process subscriptions
      this.processSubscriptions(message);
      
      // Emit general message event
      this.emit('message', message);

    } catch (error) {
      this.handleError(new Error(`Failed to parse message: ${error}`));
    }
  }

  private processSubscriptions(message: WebSocketMessage): void {
    this.subscriptions.forEach((subscription) => {
      try {
        // Check if message matches subscription topic
        if (this.matchesTopic(message.type, subscription.topic)) {
          // Apply filter if provided
          if (!subscription.filter || subscription.filter(message)) {
            subscription.callback(message);
          }
        }
      } catch (error) {
        this.log('Error in subscription callback:', error);
        this.handleError(new Error(`Subscription callback error: ${error}`));
      }
    });
  }

  private matchesTopic(messageType: string, topic: string): boolean {
    // Support wildcard matching
    if (topic === '*') return true;
    if (topic.endsWith('*')) {
      return messageType.startsWith(topic.slice(0, -1));
    }
    return messageType === topic;
  }

  private handleDisconnect(event: CloseEvent): void {
    this.log('WebSocket disconnected:', event.code, event.reason);
    
    this.metrics.connected = false;
    this.stopHeartbeat();
    
    this.emit('disconnect', {
      code: event.code,
      reason: event.reason,
      wasClean: event.wasClean
    });

    // Attempt reconnection if not a clean close
    if (!event.wasClean && this.reconnectAttempts < this.config.maxReconnectAttempts!) {
      this.scheduleReconnect();
    }
  }

  private handleError(error: Error): void {
    this.log('WebSocket error:', error);
    
    this.metrics.errorCount++;
    this.metrics.lastError = error.message;
    
    this.emit('error', error);
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer) return;

    this.reconnectAttempts++;
    this.metrics.reconnectAttempts = this.reconnectAttempts;
    
    const delay = Math.min(
      this.config.reconnectInterval! * Math.pow(2, this.reconnectAttempts - 1),
      30000 // Max 30 seconds
    );

    this.log(`Scheduling reconnect attempt ${this.reconnectAttempts} in ${delay}ms`);
    
    this.emit('reconnecting', {
      attempt: this.reconnectAttempts,
      maxAttempts: this.config.maxReconnectAttempts,
      delay
    });

    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      this.connect().then(() => {
        this.emit('reconnected', { attempt: this.reconnectAttempts });
      }).catch((error) => {
        this.log('Reconnect failed:', error);
        if (this.reconnectAttempts < this.config.maxReconnectAttempts!) {
          this.scheduleReconnect();
        }
      });
    }, delay);
  }

  private stopReconnectTimer(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  }

  private startHeartbeat(): void {
    if (!this.config.heartbeatInterval) return;
    
    this.heartbeatTimer = setInterval(() => {
      this.sendHeartbeat();
    }, this.config.heartbeatInterval);
  }

  private stopHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
  }

  private sendHeartbeat(): void {
    this.send({
      id: this.generateId(),
      type: 'heartbeat',
      data: { timestamp: Date.now() },
      timestamp: new Date()
    });
  }

  public send(message: WebSocketMessage): Promise<WebSocketMessage | void> {
    return new Promise((resolve, reject) => {
      if (!this.isConnected()) {
        if (this.streamConfig.retryFailedMessages) {
          this.messageQueue.push(message);
          resolve();
        } else {
          reject(new Error('WebSocket not connected'));
        }
        return;
      }

      try {
        // Add to buffer for batching
        if (this.streamConfig.batchMessages && message.type !== 'heartbeat') {
          this.addToBuffer(message);
          
          // If expecting response, set up correlation tracking
          if (message.correlationId) {
            const timeout = setTimeout(() => {
              this.pendingMessages.delete(message.correlationId!);
              reject(new Error('Message timeout'));
            }, this.streamConfig.messageTimeout);

            this.pendingMessages.set(message.correlationId, {
              message,
              resolve,
              reject,
              timeout
            });
          } else {
            resolve();
          }
        } else {
          this.sendDirectly(message);
          resolve();
        }
      } catch (error) {
        reject(error);
      }
    });
  }

  private addToBuffer(message: WebSocketMessage): void {
    this.messageBuffer.push(message);
    
    // Flush if buffer is full
    if (this.messageBuffer.length >= this.streamConfig.bufferSize) {
      this.flushBuffer();
    } else if (!this.flushTimer) {
      // Schedule flush
      this.flushTimer = setTimeout(() => {
        this.flushBuffer();
      }, this.streamConfig.flushInterval);
    }
  }

  private flushBuffer(): void {
    if (this.messageBuffer.length === 0) return;
    
    this.stopFlushTimer();
    
    if (this.streamConfig.batchMessages && this.messageBuffer.length > 1) {
      // Send as batch
      const batchMessage: WebSocketMessage = {
        id: this.generateId(),
        type: 'batch',
        data: this.messageBuffer.splice(0, this.streamConfig.maxBatchSize),
        timestamp: new Date()
      };
      this.sendDirectly(batchMessage);
    } else {
      // Send individually
      while (this.messageBuffer.length > 0) {
        const message = this.messageBuffer.shift()!;
        this.sendDirectly(message);
      }
    }
  }

  private stopFlushTimer(): void {
    if (this.flushTimer) {
      clearTimeout(this.flushTimer);
      this.flushTimer = null;
    }
  }

  private sendDirectly(message: WebSocketMessage): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      throw new Error('WebSocket not connected');
    }

    const payload = JSON.stringify(message);
    this.ws.send(payload);
    
    this.metrics.messagesSent++;
    this.metrics.bytesTransmitted += payload.length;
    
    this.log('Sent message:', message);
  }

  private processQueuedMessages(): void {
    while (this.messageQueue.length > 0 && this.isConnected()) {
      const message = this.messageQueue.shift()!;
      this.send(message).catch((error) => {
        this.log('Failed to send queued message:', error);
      });
    }
  }

  public subscribe(topic: string, callback: (message: WebSocketMessage) => void, filter?: (message: WebSocketMessage) => boolean): string {
    const subscription: WebSocketSubscription = {
      id: this.generateId(),
      topic,
      callback,
      filter
    };

    this.subscriptions.set(subscription.id, subscription);
    this.emit('subscription-added', { subscriptionId: subscription.id, topic });
    
    this.log('Added subscription:', subscription.id, 'for topic:', topic);
    
    return subscription.id;
  }

  public unsubscribe(subscriptionId: string): boolean {
    const subscription = this.subscriptions.get(subscriptionId);
    if (subscription) {
      this.subscriptions.delete(subscriptionId);
      this.emit('subscription-removed', { subscriptionId, topic: subscription.topic });
      this.log('Removed subscription:', subscriptionId);
      return true;
    }
    return false;
  }

  public unsubscribeFromTopic(topic: string): number {
    let count = 0;
    this.subscriptions.forEach((subscription, id) => {
      if (subscription.topic === topic) {
        this.subscriptions.delete(id);
        this.emit('subscription-removed', { subscriptionId: id, topic });
        count++;
      }
    });
    this.log('Removed', count, 'subscriptions for topic:', topic);
    return count;
  }

  public on(event: WebSocketEventType, callback: WebSocketEventCallback): void {
    const listeners = this.eventListeners.get(event);
    if (listeners) {
      listeners.add(callback);
    }
  }

  public off(event: WebSocketEventType, callback: WebSocketEventCallback): void {
    const listeners = this.eventListeners.get(event);
    if (listeners) {
      listeners.delete(callback);
    }
  }

  private emit(event: WebSocketEventType, data?: any): void {
    const listeners = this.eventListeners.get(event);
    if (listeners) {
      const eventObj: WebSocketEvent = {
        type: event,
        data,
        timestamp: new Date()
      };
      
      listeners.forEach(callback => {
        try {
          callback(eventObj);
        } catch (error) {
          this.log('Error in event callback:', error);
        }
      });
    }
  }

  public isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN;
  }

  public getMetrics(): WebSocketMetrics {
    return { ...this.metrics };
  }

  public getSubscriptions(): WebSocketSubscription[] {
    return Array.from(this.subscriptions.values());
  }

  public clearSubscriptions(): void {
    this.subscriptions.clear();
    this.emit('subscription-removed', { all: true });
  }

  private generateId(): string {
    return `ws_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  private log(message: string, ...args: any[]): void {
    if (this.config.debug) {
      console.log(`[WebSocketManager] ${message}`, ...args);
    }
  }

  public destroy(): void {
    this.disconnect();
    this.clearSubscriptions();
    this.eventListeners.clear();
    this.messageQueue = [];
    this.messageBuffer = [];
    this.pendingMessages.clear();
  }
}