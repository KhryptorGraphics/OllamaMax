import { useState, useEffect, useCallback, useRef } from 'react';
import { WebSocketManager, WebSocketConfig, WebSocketMessage, WebSocketMetrics, StreamConfig } from '../utils/websocket/WebSocketManager';

export interface WebSocketStreamConfig extends WebSocketConfig {
  autoConnect?: boolean;
  bufferSize?: number;
  flushInterval?: number;
  enableCompression?: boolean;
  enableBatching?: boolean;
  retryFailedMessages?: boolean;
}

export interface StreamSubscription {
  id: string;
  topic: string;
  active: boolean;
  messageCount: number;
  lastMessage?: Date;
}

export interface UseWebSocketStreamReturn {
  // Connection state
  isConnected: boolean;
  isConnecting: boolean;
  connectionError: string | null;
  metrics: WebSocketMetrics;
  
  // Connection management
  connect: () => Promise<void>;
  disconnect: () => void;
  reconnect: () => Promise<void>;
  
  // Message handling
  sendMessage: (type: string, data: any, correlationId?: string) => Promise<WebSocketMessage | void>;
  sendBatch: (messages: Array<{ type: string; data: any }>) => Promise<void>;
  
  // Subscription management
  subscriptions: StreamSubscription[];
  subscribe: (topic: string, callback: (message: WebSocketMessage) => void, filter?: (message: WebSocketMessage) => boolean) => string;
  unsubscribe: (subscriptionId: string) => void;
  unsubscribeFromTopic: (topic: string) => void;
  clearSubscriptions: () => void;
  
  // Real-time data streaming
  createDataStream: <T>(topic: string, transform?: (data: any) => T) => DataStream<T>;
  
  // Event handling
  onConnect: (callback: () => void) => () => void;
  onDisconnect: (callback: (reason?: string) => void) => () => void;
  onError: (callback: (error: Error) => void) => () => void;
  onReconnecting: (callback: (attempt: number) => void) => () => void;
  
  // Stream control
  pauseStream: () => void;
  resumeStream: () => void;
  isPaused: boolean;
}

export interface DataStream<T> {
  subscribe: (callback: (data: T) => void) => () => void;
  filter: (predicate: (data: T) => boolean) => DataStream<T>;
  map: <U>(transform: (data: T) => U) => DataStream<U>;
  buffer: (size: number) => DataStream<T[]>;
  throttle: (ms: number) => DataStream<T>;
  debounce: (ms: number) => DataStream<T>;
  latest: () => T | undefined;
  history: (limit?: number) => T[];
  close: () => void;
}

class DataStreamImpl<T> implements DataStream<T> {
  private subscribers: Set<(data: T) => void> = new Set();
  private dataHistory: T[] = [];
  private lastData: T | undefined;
  private filters: Array<(data: T) => boolean> = [];
  private mappers: Array<(data: any) => any> = [];
  private throttleTimer: NodeJS.Timeout | null = null;
  private debounceTimer: NodeJS.Timeout | null = null;
  private bufferData: T[] = [];
  private bufferSize = 0;
  private isActive = true;

  constructor(
    private wsManager: WebSocketManager,
    private topic: string,
    private transform?: (data: any) => T,
    private historyLimit = 100
  ) {
    this.setupSubscription();
  }

  private setupSubscription(): void {
    this.wsManager.subscribe(this.topic, (message) => {
      if (!this.isActive) return;

      try {
        let data = message.data;
        
        // Apply transform if provided
        if (this.transform) {
          data = this.transform(data);
        }
        
        // Apply filters
        if (this.filters.length > 0 && !this.filters.every(filter => filter(data))) {
          return;
        }
        
        // Apply mappers
        this.mappers.forEach(mapper => {
          data = mapper(data);
        });

        this.processData(data);
      } catch (error) {
        console.error('Error processing stream data:', error);
      }
    });
  }

  private processData(data: T): void {
    this.lastData = data;
    this.dataHistory.push(data);
    
    // Maintain history limit
    if (this.dataHistory.length > this.historyLimit) {
      this.dataHistory.shift();
    }

    if (this.bufferSize > 0) {
      this.bufferData.push(data);
      if (this.bufferData.length >= this.bufferSize) {
        this.emitToSubscribers(this.bufferData.slice() as any);
        this.bufferData = [];
      }
    } else {
      this.emitToSubscribers(data);
    }
  }

  private emitToSubscribers(data: T): void {
    this.subscribers.forEach(callback => {
      try {
        callback(data);
      } catch (error) {
        console.error('Error in stream subscriber:', error);
      }
    });
  }

  subscribe(callback: (data: T) => void): () => void {
    this.subscribers.add(callback);
    
    // Send latest data immediately if available
    if (this.lastData !== undefined) {
      try {
        callback(this.lastData);
      } catch (error) {
        console.error('Error in immediate callback:', error);
      }
    }
    
    return () => {
      this.subscribers.delete(callback);
    };
  }

  filter(predicate: (data: T) => boolean): DataStream<T> {
    const newStream = new DataStreamImpl<T>(this.wsManager, this.topic, this.transform, this.historyLimit);
    newStream.filters = [...this.filters, predicate];
    newStream.mappers = [...this.mappers];
    return newStream;
  }

  map<U>(transform: (data: T) => U): DataStream<U> {
    const newStream = new DataStreamImpl<U>(this.wsManager, this.topic, this.transform, this.historyLimit);
    newStream.filters = [...this.filters] as any;
    newStream.mappers = [...this.mappers, transform];
    return newStream;
  }

  buffer(size: number): DataStream<T[]> {
    const newStream = new DataStreamImpl<T[]>(this.wsManager, this.topic, this.transform, this.historyLimit);
    newStream.filters = [...this.filters] as any;
    newStream.mappers = [...this.mappers] as any;
    newStream.bufferSize = size;
    return newStream;
  }

  throttle(ms: number): DataStream<T> {
    const newStream = new DataStreamImpl<T>(this.wsManager, this.topic, this.transform, this.historyLimit);
    newStream.filters = [...this.filters];
    newStream.mappers = [...this.mappers];
    
    const originalEmit = newStream.emitToSubscribers.bind(newStream);
    newStream.emitToSubscribers = (data: T) => {
      if (newStream.throttleTimer) return;
      
      originalEmit(data);
      newStream.throttleTimer = setTimeout(() => {
        newStream.throttleTimer = null;
      }, ms);
    };
    
    return newStream;
  }

  debounce(ms: number): DataStream<T> {
    const newStream = new DataStreamImpl<T>(this.wsManager, this.topic, this.transform, this.historyLimit);
    newStream.filters = [...this.filters];
    newStream.mappers = [...this.mappers];
    
    const originalEmit = newStream.emitToSubscribers.bind(newStream);
    newStream.emitToSubscribers = (data: T) => {
      if (newStream.debounceTimer) {
        clearTimeout(newStream.debounceTimer);
      }
      
      newStream.debounceTimer = setTimeout(() => {
        originalEmit(data);
        newStream.debounceTimer = null;
      }, ms);
    };
    
    return newStream;
  }

  latest(): T | undefined {
    return this.lastData;
  }

  history(limit?: number): T[] {
    const historyLimit = limit || this.dataHistory.length;
    return this.dataHistory.slice(-historyLimit);
  }

  close(): void {
    this.isActive = false;
    this.subscribers.clear();
    
    if (this.throttleTimer) {
      clearTimeout(this.throttleTimer);
    }
    
    if (this.debounceTimer) {
      clearTimeout(this.debounceTimer);
    }
  }
}

export function useWebSocketStream(config: WebSocketStreamConfig): UseWebSocketStreamReturn {
  const [isConnected, setIsConnected] = useState(false);
  const [isConnecting, setIsConnecting] = useState(false);
  const [connectionError, setConnectionError] = useState<string | null>(null);
  const [metrics, setMetrics] = useState<WebSocketMetrics>({
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
  });
  const [subscriptions, setSubscriptions] = useState<StreamSubscription[]>([]);
  const [isPaused, setIsPaused] = useState(false);

  const wsManagerRef = useRef<WebSocketManager | null>(null);
  const eventCallbacksRef = useRef<{
    onConnect: Set<() => void>;
    onDisconnect: Set<(reason?: string) => void>;
    onError: Set<(error: Error) => void>;
    onReconnecting: Set<(attempt: number) => void>;
  }>({
    onConnect: new Set(),
    onDisconnect: new Set(),
    onError: new Set(),
    onReconnecting: new Set()
  });

  // Initialize WebSocket manager
  useEffect(() => {
    const streamConfig: StreamConfig = {
      bufferSize: config.bufferSize || 100,
      flushInterval: config.flushInterval || 1000,
      compression: config.enableCompression || false,
      batchMessages: config.enableBatching !== false,
      maxBatchSize: 10,
      retryFailedMessages: config.retryFailedMessages !== false,
      messageTimeout: 10000
    };

    wsManagerRef.current = new WebSocketManager(config, streamConfig);

    // Set up event listeners
    const wsManager = wsManagerRef.current;

    wsManager.on('connect', () => {
      setIsConnected(true);
      setIsConnecting(false);
      setConnectionError(null);
      eventCallbacksRef.current.onConnect.forEach(callback => callback());
    });

    wsManager.on('disconnect', (event) => {
      setIsConnected(false);
      setIsConnecting(false);
      eventCallbacksRef.current.onDisconnect.forEach(callback => 
        callback(event.data?.reason)
      );
    });

    wsManager.on('error', (error) => {
      setConnectionError(error.message);
      setIsConnecting(false);
      eventCallbacksRef.current.onError.forEach(callback => callback(error));
    });

    wsManager.on('reconnecting', (event) => {
      setIsConnecting(true);
      eventCallbacksRef.current.onReconnecting.forEach(callback => 
        callback(event.data?.attempt || 0)
      );
    });

    wsManager.on('reconnected', () => {
      setIsConnecting(false);
      setConnectionError(null);
    });

    wsManager.on('subscription-added', (event) => {
      setSubscriptions(prev => [
        ...prev,
        {
          id: event.data.subscriptionId,
          topic: event.data.topic,
          active: true,
          messageCount: 0,
          lastMessage: undefined
        }
      ]);
    });

    wsManager.on('subscription-removed', (event) => {
      if (event.data.all) {
        setSubscriptions([]);
      } else {
        setSubscriptions(prev => 
          prev.filter(sub => sub.id !== event.data.subscriptionId)
        );
      }
    });

    wsManager.on('message', () => {
      setMetrics(wsManager.getMetrics());
      
      // Update subscription message counts
      setSubscriptions(prev => 
        prev.map(sub => ({
          ...sub,
          messageCount: sub.messageCount + 1,
          lastMessage: new Date()
        }))
      );
    });

    // Auto-connect if enabled
    if (config.autoConnect !== false) {
      connect();
    }

    return () => {
      if (wsManagerRef.current) {
        wsManagerRef.current.destroy();
      }
    };
  }, []);

  const connect = useCallback(async (): Promise<void> => {
    if (!wsManagerRef.current || isConnecting || isConnected) return;

    setIsConnecting(true);
    setConnectionError(null);

    try {
      await wsManagerRef.current.connect();
    } catch (error) {
      setConnectionError(error instanceof Error ? error.message : 'Connection failed');
      setIsConnecting(false);
      throw error;
    }
  }, [isConnecting, isConnected]);

  const disconnect = useCallback((): void => {
    if (wsManagerRef.current) {
      wsManagerRef.current.disconnect();
    }
  }, []);

  const reconnect = useCallback(async (): Promise<void> => {
    disconnect();
    await new Promise(resolve => setTimeout(resolve, 1000)); // Brief delay
    await connect();
  }, [connect, disconnect]);

  const sendMessage = useCallback(async (type: string, data: any, correlationId?: string): Promise<WebSocketMessage | void> => {
    if (!wsManagerRef.current || isPaused) {
      throw new Error('WebSocket not available or stream is paused');
    }

    const message: WebSocketMessage = {
      id: `msg_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      type,
      data,
      timestamp: new Date(),
      correlationId
    };

    return wsManagerRef.current.send(message);
  }, [isPaused]);

  const sendBatch = useCallback(async (messages: Array<{ type: string; data: any }>): Promise<void> => {
    if (!wsManagerRef.current || isPaused) {
      throw new Error('WebSocket not available or stream is paused');
    }

    const batchMessage: WebSocketMessage = {
      id: `batch_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      type: 'batch',
      data: messages.map(msg => ({
        id: `msg_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
        type: msg.type,
        data: msg.data,
        timestamp: new Date()
      })),
      timestamp: new Date()
    };

    await wsManagerRef.current.send(batchMessage);
  }, [isPaused]);

  const subscribe = useCallback((
    topic: string, 
    callback: (message: WebSocketMessage) => void, 
    filter?: (message: WebSocketMessage) => boolean
  ): string => {
    if (!wsManagerRef.current) {
      throw new Error('WebSocket manager not initialized');
    }

    return wsManagerRef.current.subscribe(topic, callback, filter);
  }, []);

  const unsubscribe = useCallback((subscriptionId: string): void => {
    if (wsManagerRef.current) {
      wsManagerRef.current.unsubscribe(subscriptionId);
    }
  }, []);

  const unsubscribeFromTopic = useCallback((topic: string): void => {
    if (wsManagerRef.current) {
      wsManagerRef.current.unsubscribeFromTopic(topic);
    }
  }, []);

  const clearSubscriptions = useCallback((): void => {
    if (wsManagerRef.current) {
      wsManagerRef.current.clearSubscriptions();
    }
  }, []);

  const createDataStream = useCallback(<T>(topic: string, transform?: (data: any) => T): DataStream<T> => {
    if (!wsManagerRef.current) {
      throw new Error('WebSocket manager not initialized');
    }

    return new DataStreamImpl<T>(wsManagerRef.current, topic, transform);
  }, []);

  const onConnect = useCallback((callback: () => void): (() => void) => {
    eventCallbacksRef.current.onConnect.add(callback);
    return () => {
      eventCallbacksRef.current.onConnect.delete(callback);
    };
  }, []);

  const onDisconnect = useCallback((callback: (reason?: string) => void): (() => void) => {
    eventCallbacksRef.current.onDisconnect.add(callback);
    return () => {
      eventCallbacksRef.current.onDisconnect.delete(callback);
    };
  }, []);

  const onError = useCallback((callback: (error: Error) => void): (() => void) => {
    eventCallbacksRef.current.onError.add(callback);
    return () => {
      eventCallbacksRef.current.onError.delete(callback);
    };
  }, []);

  const onReconnecting = useCallback((callback: (attempt: number) => void): (() => void) => {
    eventCallbacksRef.current.onReconnecting.add(callback);
    return () => {
      eventCallbacksRef.current.onReconnecting.delete(callback);
    };
  }, []);

  const pauseStream = useCallback((): void => {
    setIsPaused(true);
  }, []);

  const resumeStream = useCallback((): void => {
    setIsPaused(false);
  }, []);

  return {
    isConnected,
    isConnecting,
    connectionError,
    metrics,
    connect,
    disconnect,
    reconnect,
    sendMessage,
    sendBatch,
    subscriptions,
    subscribe,
    unsubscribe,
    unsubscribeFromTopic,
    clearSubscriptions,
    createDataStream,
    onConnect,
    onDisconnect,
    onError,
    onReconnecting,
    pauseStream,
    resumeStream,
    isPaused
  };
}