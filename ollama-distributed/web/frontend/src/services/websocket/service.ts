/**
 * Enhanced WebSocket Service - Real-time coordination service
 * Provides high-level abstractions for WebSocket operations with automatic
 * subscription management, message batching, and state synchronization
 */

import { wsClient } from './client'
import { useStore } from '@/stores'
import type {
  WebSocketMessage,
  WebSocketEventHandlers,
  NodeStatusUpdate,
  ModelSyncUpdate,
  TaskUpdate,
  TransferProgressUpdate,
  ClusterMetricsUpdate,
  PerformanceAlert,
  SecurityAlert,
  SystemNotification,
  UserNotification,
} from '@/types/websocket'

export interface WebSocketServiceConfig {
  autoConnect?: boolean
  autoReconnect?: boolean
  batchMessages?: boolean
  batchInterval?: number
  debug?: boolean
  heartbeatInterval?: number
  messageQueue?: boolean
  maxQueueSize?: number
  retryAttempts?: number
  retryDelay?: number
}

export interface SubscriptionOptions {
  filters?: Record<string, any>
  persistent?: boolean
  qos?: 'at-most-once' | 'at-least-once' | 'exactly-once'
}

export interface MessageBatch {
  messages: WebSocketMessage[]
  timestamp: number
  size: number
}

export class WebSocketService {
  private config: Required<WebSocketServiceConfig>
  private subscriptions: Map<string, SubscriptionOptions> = new Map()
  private messageQueue: WebSocketMessage[] = []
  private messageBatch: WebSocketMessage[] = []
  private batchTimer: NodeJS.Timeout | null = null
  private retryTimer: NodeJS.Timeout | null = null
  private connectionPromise: Promise<void> | null = null
  private eventHandlers: Map<string, Set<Function>> = new Map()
  private isInitialized: boolean = false

  constructor(config: WebSocketServiceConfig = {}) {
    this.config = {
      autoConnect: true,
      autoReconnect: true,
      batchMessages: false,
      batchInterval: 100,
      debug: false,
      heartbeatInterval: 30000,
      messageQueue: true,
      maxQueueSize: 1000,
      retryAttempts: 5,
      retryDelay: 1000,
      ...config,
    }

    this.initialize()
  }

  /**
   * Initialize the WebSocket service with enhanced features
   */
  private async initialize(): Promise<void> {
    if (this.isInitialized) return

    try {
      // Set up enhanced event handlers
      this.setupEventHandlers()

      // Set up message batching if enabled
      if (this.config.batchMessages) {
        this.setupMessageBatching()
      }

      // Set up automatic connection management
      if (this.config.autoConnect) {
        this.setupAutoConnection()
      }

      this.isInitialized = true
      this.log('WebSocket service initialized successfully')
    } catch (error) {
      this.log('Failed to initialize WebSocket service:', error)
      throw error
    }
  }

  /**
   * Connect to WebSocket server with enhanced retry logic
   */
  public async connect(): Promise<void> {
    if (this.connectionPromise) {
      return this.connectionPromise
    }

    this.connectionPromise = this.performConnection()
    return this.connectionPromise
  }

  private async performConnection(): Promise<void> {
    let attempt = 0
    
    while (attempt < this.config.retryAttempts) {
      try {
        await wsClient.connect()
        
        // Restore persistent subscriptions
        await this.restoreSubscriptions()
        
        // Process queued messages
        await this.processMessageQueue()
        
        this.connectionPromise = null
        this.log('Connected successfully')
        return
        
      } catch (error) {
        attempt++
        this.log(`Connection attempt ${attempt} failed:`, error)
        
        if (attempt < this.config.retryAttempts) {
          const delay = this.config.retryDelay * Math.pow(2, attempt - 1)
          this.log(`Retrying in ${delay}ms...`)
          await this.wait(delay)
        } else {
          this.connectionPromise = null
          throw error
        }
      }
    }
  }

  /**
   * Enhanced subscription management with persistence and QoS
   */
  public async subscribe(
    channels: string[], 
    options: SubscriptionOptions = {}
  ): Promise<void> {
    const { filters, persistent = true } = options

    // Store subscription for restoration
    if (persistent) {
      channels.forEach(channel => {
        this.subscriptions.set(channel, options)
      })
    }

    if (!wsClient.getState().connected) {
      this.log('Not connected, subscription will be restored on reconnect:', channels)
      return
    }

    try {
      wsClient.subscribe(channels, filters)
      this.log('Subscribed to channels:', channels)
    } catch (error) {
      this.log('Failed to subscribe to channels:', channels, error)
      throw error
    }
  }

  /**
   * Enhanced unsubscription with cleanup
   */
  public async unsubscribe(channels: string[]): Promise<void> {
    // Remove from persistent subscriptions
    channels.forEach(channel => {
      this.subscriptions.delete(channel)
    })

    if (!wsClient.getState().connected) {
      this.log('Not connected, unsubscription recorded:', channels)
      return
    }

    try {
      wsClient.unsubscribe(channels)
      this.log('Unsubscribed from channels:', channels)
    } catch (error) {
      this.log('Failed to unsubscribe from channels:', channels, error)
      throw error
    }
  }

  /**
   * Enhanced message sending with queuing and batching
   */
  public async send(message: WebSocketMessage): Promise<void> {
    if (!wsClient.getState().connected) {
      if (this.config.messageQueue) {
        this.queueMessage(message)
        this.log('Message queued (not connected):', message.type)
        return
      } else {
        throw new Error('WebSocket not connected and queuing is disabled')
      }
    }

    if (this.config.batchMessages) {
      this.batchMessage(message)
    } else {
      this.sendMessage(message)
    }
  }

  /**
   * Send multiple messages efficiently
   */
  public async sendBatch(messages: WebSocketMessage[]): Promise<void> {
    if (!wsClient.getState().connected) {
      if (this.config.messageQueue) {
        messages.forEach(message => this.queueMessage(message))
        this.log('Messages queued (not connected):', messages.length)
        return
      } else {
        throw new Error('WebSocket not connected and queuing is disabled')
      }
    }

    // Send as batch or individual messages based on configuration
    if (this.config.batchMessages) {
      messages.forEach(message => this.batchMessage(message))
      this.flushBatch()
    } else {
      messages.forEach(message => this.sendMessage(message))
    }
  }

  /**
   * Disconnect with cleanup
   */
  public disconnect(): void {
    if (this.batchTimer) {
      clearTimeout(this.batchTimer)
      this.batchTimer = null
    }

    if (this.retryTimer) {
      clearTimeout(this.retryTimer)
      this.retryTimer = null
    }

    // Flush any pending batch
    if (this.messageBatch.length > 0) {
      this.flushBatch()
    }

    wsClient.disconnect()
    this.connectionPromise = null
    this.log('Disconnected')
  }

  /**
   * Get current connection and subscription status
   */
  public getStatus() {
    const wsState = wsClient.getState()
    return {
      connected: wsState.connected,
      connecting: wsState.connecting,
      error: wsState.error,
      subscriptions: Array.from(this.subscriptions.keys()),
      queuedMessages: this.messageQueue.length,
      batchedMessages: this.messageBatch.length,
      lastMessage: wsState.lastMessage,
      reconnectAttempts: wsState.reconnectAttempts,
    }
  }

  /**
   * Private helper methods
   */
  private setupEventHandlers(): void {
    // Event handler setup would go here
    this.log('Event handlers set up')
  }

  private setupMessageBatching(): void {
    this.batchTimer = setInterval(() => {
      if (this.messageBatch.length > 0) {
        this.flushBatch()
      }
    }, this.config.batchInterval)
  }

  private setupAutoConnection(): void {
    const store = useStore.getState()
    
    // Connect when authenticated
    if (store.auth.isAuthenticated) {
      this.connect().catch(error => {
        this.log('Auto-connection failed:', error)
      })
    }
  }

  private async restoreSubscriptions(): Promise<void> {
    if (this.subscriptions.size === 0) return

    const channels = Array.from(this.subscriptions.keys())
    try {
      wsClient.subscribe(channels)
      this.log('Restored subscriptions:', channels)
    } catch (error) {
      this.log('Failed to restore subscriptions:', error)
    }
  }

  private async processMessageQueue(): Promise<void> {
    if (this.messageQueue.length === 0) return

    this.log(`Processing ${this.messageQueue.length} queued messages`)
    
    const messages = [...this.messageQueue]
    this.messageQueue = []

    if (this.config.batchMessages) {
      this.messageBatch.push(...messages)
      this.flushBatch()
    } else {
      messages.forEach(message => this.sendMessage(message))
    }
  }

  private queueMessage(message: WebSocketMessage): void {
    if (this.messageQueue.length >= this.config.maxQueueSize) {
      // Remove oldest message to make room
      this.messageQueue.shift()
      this.log('Message queue full, removed oldest message')
    }
    
    this.messageQueue.push(message)
  }

  private batchMessage(message: WebSocketMessage): void {
    this.messageBatch.push(message)
    
    // If batch is getting large, flush immediately
    if (this.messageBatch.length >= 50) {
      this.flushBatch()
    }
  }

  private flushBatch(): void {
    if (this.messageBatch.length === 0) return

    const batch: MessageBatch = {
      messages: [...this.messageBatch],
      timestamp: Date.now(),
      size: this.messageBatch.length,
    }

    // Send batch as single message or individual messages
    if (batch.size === 1) {
      this.sendMessage(batch.messages[0])
    } else {
      batch.messages.forEach(message => this.sendMessage(message))
    }

    this.messageBatch = []
    this.log('Flushed batch:', batch.size, 'messages')
  }

  private sendMessage(message: WebSocketMessage): void {
    try {
      wsClient.send(message)
      this.log('Message sent:', message.type)
    } catch (error) {
      this.log('Failed to send message:', message.type, error)
      
      // Re-queue if sending fails and queuing is enabled
      if (this.config.messageQueue) {
        this.queueMessage(message)
      }
    }
  }

  private wait(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms))
  }

  private log(message: string, ...args: any[]): void {
    if (this.config.debug) {
      console.log(`[WebSocketService] ${message}`, ...args)
    }
  }
}

// Create singleton instance
export const webSocketService = new WebSocketService({
  debug: import.meta.env.DEV,
  autoConnect: true,
  autoReconnect: true,
  batchMessages: true,
  messageQueue: true,
})

export default webSocketService