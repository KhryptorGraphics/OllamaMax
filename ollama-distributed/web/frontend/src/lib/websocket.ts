/**
 * WebSocket client implementation with auto-reconnection and topic subscriptions
 * Provides reliable real-time communication with the distributed Ollama backend
 */

import {
  WSMessage,
  MessageTypes,
  ConnectionState,
  WSClientConfig,
  WSEventListener,
  WSErrorListener,
  WSStateListener,
  WSSubscription
} from '../types/websocket'

export class WebSocketClient {
  private ws: WebSocket | null = null
  private config: Required<WSClientConfig>
  private reconnectAttempts = 0
  private reconnectTimer: NodeJS.Timeout | null = null
  private heartbeatTimer: NodeJS.Timeout | null = null
  private pingTimer: NodeJS.Timeout | null = null
  
  // Event listeners
  private subscribers = new Map<string, Set<WSEventListener>>()
  private errorListeners = new Set<WSErrorListener>()
  private stateListeners = new Set<WSStateListener>()
  private subscriptions = new Map<string, WSSubscription>()
  
  // State management
  private currentState: ConnectionState = ConnectionState.DISCONNECTED
  private isManualDisconnect = false
  private lastPingTime = 0
  private serverTimeOffset = 0

  constructor(config: WSClientConfig) {
    this.config = {
      reconnectAttempts: 10,
      reconnectInterval: 5000,
      heartbeatInterval: 30000,
      debug: false,
      ...config
    }
    
    if (this.config.debug) {
      console.log('[WebSocket] Client initialized with config:', this.config)
    }
  }

  /**
   * Connect to WebSocket server
   */
  async connect(): Promise<void> {
    if (this.ws?.readyState === WebSocket.OPEN) {
      if (this.config.debug) console.log('[WebSocket] Already connected')
      return
    }

    this.isManualDisconnect = false
    this.setState(ConnectionState.CONNECTING)

    return new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(this.config.url)
        
        this.ws.onopen = () => {
          if (this.config.debug) console.log('[WebSocket] Connected')
          this.setState(ConnectionState.CONNECTED)
          this.reconnectAttempts = 0
          this.setupHeartbeat()
          this.resubscribeToTopics()
          resolve()
        }

        this.ws.onmessage = (event) => {
          this.handleMessage(event.data)
        }

        this.ws.onclose = (event) => {
          if (this.config.debug) {
            console.log('[WebSocket] Disconnected:', event.code, event.reason)
          }
          this.cleanup()
          
          if (!this.isManualDisconnect && this.reconnectAttempts < this.config.reconnectAttempts) {
            this.setState(ConnectionState.RECONNECTING)
            this.scheduleReconnect()
          } else {
            this.setState(ConnectionState.DISCONNECTED)
          }
        }

        this.ws.onerror = (error) => {
          if (this.config.debug) console.error('[WebSocket] Error:', error)
          this.setState(ConnectionState.ERROR)
          this.emitError(new Error('WebSocket connection error'))
          reject(error)
        }

      } catch (error) {
        this.setState(ConnectionState.ERROR)
        this.emitError(error as Error)
        reject(error)
      }
    })
  }

  /**
   * Disconnect from WebSocket server
   */
  disconnect(): void {
    this.isManualDisconnect = true
    this.cleanup()
    
    if (this.ws) {
      this.ws.close(1000, 'Manual disconnect')
      this.ws = null
    }
    
    this.setState(ConnectionState.DISCONNECTED)
    if (this.config.debug) console.log('[WebSocket] Manually disconnected')
  }

  /**
   * Subscribe to a topic for real-time updates
   */
  subscribe(topic: string, callback: WSEventListener): () => void {
    const subscriptionId = `${topic}-${Date.now()}-${Math.random()}`
    
    // Add to local subscribers
    if (!this.subscribers.has(topic)) {
      this.subscribers.set(topic, new Set())
    }
    this.subscribers.get(topic)!.add(callback)
    
    // Store subscription for reconnection
    this.subscriptions.set(subscriptionId, { topic, callback, id: subscriptionId })
    
    // Send subscription to server if connected
    if (this.isConnected()) {
      this.send({
        type: MessageTypes.SUBSCRIBE,
        data: { topic },
        timestamp: new Date().toISOString()
      })
    }
    
    if (this.config.debug) console.log(`[WebSocket] Subscribed to topic: ${topic}`)
    
    // Return unsubscribe function
    return () => this.unsubscribe(topic, callback)
  }

  /**
   * Unsubscribe from a topic
   */
  unsubscribe(topic: string, callback?: WSEventListener): void {
    if (callback) {
      // Remove specific callback
      const topicSubscribers = this.subscribers.get(topic)
      if (topicSubscribers) {
        topicSubscribers.delete(callback)
        if (topicSubscribers.size === 0) {
          this.subscribers.delete(topic)
        }
      }
      
      // Remove from stored subscriptions
      for (const [id, sub] of this.subscriptions.entries()) {
        if (sub.topic === topic && sub.callback === callback) {
          this.subscriptions.delete(id)
          break
        }
      }
    } else {
      // Remove all callbacks for topic
      this.subscribers.delete(topic)
      
      // Remove all stored subscriptions for topic
      for (const [id, sub] of this.subscriptions.entries()) {
        if (sub.topic === topic) {
          this.subscriptions.delete(id)
        }
      }
    }
    
    // Send unsubscription to server if connected
    if (this.isConnected()) {
      this.send({
        type: MessageTypes.UNSUBSCRIBE,
        data: { topic },
        timestamp: new Date().toISOString()
      })
    }
    
    if (this.config.debug) console.log(`[WebSocket] Unsubscribed from topic: ${topic}`)
  }

  /**
   * Send a message to the server
   */
  send(message: WSMessage): void {
    if (!this.isConnected()) {
      if (this.config.debug) console.warn('[WebSocket] Cannot send message - not connected')
      return
    }
    
    try {
      this.ws!.send(JSON.stringify(message))
      if (this.config.debug) console.log('[WebSocket] Sent message:', message.type)
    } catch (error) {
      if (this.config.debug) console.error('[WebSocket] Failed to send message:', error)
      this.emitError(error as Error)
    }
  }

  /**
   * Add error listener
   */
  onError(listener: WSErrorListener): () => void {
    this.errorListeners.add(listener)
    return () => this.errorListeners.delete(listener)
  }

  /**
   * Add state change listener
   */
  onStateChange(listener: WSStateListener): () => void {
    this.stateListeners.add(listener)
    return () => this.stateListeners.delete(listener)
  }

  /**
   * Get current connection state
   */
  getState(): ConnectionState {
    return this.currentState
  }

  /**
   * Check if connected
   */
  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN
  }

  /**
   * Get server time offset for sync
   */
  getServerTimeOffset(): number {
    return this.serverTimeOffset
  }

  /**
   * Get connection statistics
   */
  getStats() {
    return {
      state: this.currentState,
      reconnectAttempts: this.reconnectAttempts,
      subscriptions: this.subscriptions.size,
      serverTimeOffset: this.serverTimeOffset,
      lastPingTime: this.lastPingTime
    }
  }

  // Private methods

  private handleMessage(data: string): void {
    try {
      const message: WSMessage = JSON.parse(data)
      
      if (this.config.debug) {
        console.log('[WebSocket] Received message:', message.type, message.data)
      }
      
      // Handle special message types
      switch (message.type) {
        case MessageTypes.WELCOME:
          if (this.config.debug) console.log('[WebSocket] Welcome received')
          break
          
        case MessageTypes.HEARTBEAT:
          this.handleHeartbeat(message)
          break
          
        case MessageTypes.PONG:
          this.handlePong(message)
          break
          
        default:
          // Emit to topic subscribers
          this.emitToSubscribers(message.type, message.data)
          break
      }
      
    } catch (error) {
      if (this.config.debug) console.error('[WebSocket] Failed to parse message:', error)
      this.emitError(new Error('Failed to parse WebSocket message'))
    }
  }

  private handleHeartbeat(message: WSMessage): void {
    const data = message.data
    if (data?.server_time) {
      const clientTime = Date.now() / 1000
      this.serverTimeOffset = data.server_time - clientTime
    }
  }

  private handlePong(message: WSMessage): void {
    if (this.lastPingTime > 0) {
      const latency = Date.now() - this.lastPingTime
      if (this.config.debug) console.log(`[WebSocket] Ping latency: ${latency}ms`)
    }
  }

  private emitToSubscribers(topic: string, data: any): void {
    const subscribers = this.subscribers.get(topic)
    if (subscribers) {
      subscribers.forEach(callback => {
        try {
          callback(data)
        } catch (error) {
          if (this.config.debug) {
            console.error(`[WebSocket] Error in subscriber callback for ${topic}:`, error)
          }
        }
      })
    }
  }

  private emitError(error: Error): void {
    this.errorListeners.forEach(listener => {
      try {
        listener(error)
      } catch (err) {
        console.error('[WebSocket] Error in error listener:', err)
      }
    })
  }

  private setState(newState: ConnectionState): void {
    if (this.currentState !== newState) {
      this.currentState = newState
      this.stateListeners.forEach(listener => {
        try {
          listener(newState)
        } catch (error) {
          console.error('[WebSocket] Error in state listener:', error)
        }
      })
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
    }
    
    this.reconnectAttempts++
    const delay = Math.min(
      this.config.reconnectInterval * Math.pow(1.5, this.reconnectAttempts - 1),
      30000 // Max 30 seconds
    )
    
    if (this.config.debug) {
      console.log(`[WebSocket] Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`)
    }
    
    this.reconnectTimer = setTimeout(() => {
      this.connect().catch(error => {
        if (this.config.debug) console.error('[WebSocket] Reconnection failed:', error)
      })
    }, delay)
  }

  private setupHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer)
    }
    
    this.heartbeatTimer = setInterval(() => {
      if (this.isConnected()) {
        this.lastPingTime = Date.now()
        this.send({
          type: MessageTypes.PING,
          data: { timestamp: this.lastPingTime },
          timestamp: new Date().toISOString()
        })
      }
    }, this.config.heartbeatInterval)
  }

  private resubscribeToTopics(): void {
    // Re-send all subscriptions after reconnection
    for (const subscription of this.subscriptions.values()) {
      this.send({
        type: MessageTypes.SUBSCRIBE,
        data: { topic: subscription.topic },
        timestamp: new Date().toISOString()
      })
    }
    
    if (this.config.debug && this.subscriptions.size > 0) {
      console.log(`[WebSocket] Resubscribed to ${this.subscriptions.size} topics`)
    }
  }

  private cleanup(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer)
      this.heartbeatTimer = null
    }
    
    if (this.pingTimer) {
      clearTimeout(this.pingTimer)
      this.pingTimer = null
    }
  }
}

// Create a singleton instance
let wsClient: WebSocketClient | null = null

export function createWebSocketClient(config: WSClientConfig): WebSocketClient {
  if (wsClient) {
    wsClient.disconnect()
  }
  wsClient = new WebSocketClient(config)
  return wsClient
}

export function getWebSocketClient(): WebSocketClient | null {
  return wsClient
}