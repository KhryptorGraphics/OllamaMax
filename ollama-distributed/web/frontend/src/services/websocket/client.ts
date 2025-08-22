// Enhanced WebSocket client for real-time communication
import type {
  WebSocketClient as IWebSocketClient,
  WebSocketConfig,
  WebSocketState,
  WebSocketMessage,
  WebSocketEventHandlers,
  SubscriptionMessage,
} from '@/types/websocket'

export class WebSocketClient implements IWebSocketClient {
  private ws: WebSocket | null = null
  private config: WebSocketConfig
  private state: WebSocketState
  private eventHandlers: Map<string, Set<Function>> = new Map()
  private heartbeatTimer: number | null = null
  private reconnectTimer: number | null = null
  private subscriptions: Set<string> = new Set()

  constructor(config: Partial<WebSocketConfig> = {}) {
    this.config = {
      url: config.url || this.getDefaultUrl(),
      protocols: config.protocols,
      reconnect: config.reconnect ?? true,
      maxReconnectAttempts: config.maxReconnectAttempts ?? 5,
      reconnectInterval: config.reconnectInterval ?? 5000,
      heartbeatInterval: config.heartbeatInterval ?? 30000,
      timeout: config.timeout ?? 10000,
      debug: config.debug ?? false,
    }

    this.state = {
      connected: false,
      connecting: false,
      error: null,
      lastMessage: null,
      subscriptions: new Set(),
      reconnectAttempts: 0,
      maxReconnectAttempts: this.config.maxReconnectAttempts,
      reconnectInterval: this.config.reconnectInterval,
    }
  }

  private getDefaultUrl(): string {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    return `${protocol}//${host}/ws`
  }

  private log(message: string, ...args: any[]): void {
    if (this.config.debug) {
      console.log(`[WebSocket] ${message}`, ...args)
    }
  }

  private emit(event: string, data?: any): void {
    const handlers = this.eventHandlers.get(event)
    if (handlers) {
      handlers.forEach(handler => {
        try {
          handler(data)
        } catch (error) {
          console.error(`Error in WebSocket event handler for ${event}:`, error)
        }
      })
    }
  }

  private startHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer)
    }

    this.heartbeatTimer = window.setInterval(() => {
      if (this.state.connected && this.ws?.readyState === WebSocket.OPEN) {
        this.send({
          type: 'ping',
          payload: {},
          timestamp: Date.now(),
        })
      }
    }, this.config.heartbeatInterval)
  }

  private stopHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer)
      this.heartbeatTimer = null
    }
  }

  private handleMessage(event: MessageEvent): void {
    try {
      const message: WebSocketMessage = JSON.parse(event.data)
      
      this.state.lastMessage = message
      this.log('Received message:', message)

      // Handle system messages
      switch (message.type) {
        case 'welcome':
          this.log('Connected to server')
          break
        
        case 'pong':
          this.log('Received pong')
          break
        
        case 'subscription_confirmed':
          this.subscriptions.add(message.payload.channel)
          this.state.subscriptions = new Set(this.subscriptions)
          break
        
        case 'subscription_error':
          this.log('Subscription error:', message.payload)
          break
        
        case 'error':
          this.state.error = message.payload.message || 'WebSocket error'
          this.emit('error', new Error(this.state.error))
          break
      }

      // Emit specific event
      this.emit(message.type, message.payload)
      
      // Emit generic message event
      this.emit('message', message)
      
    } catch (error) {
      console.error('Failed to parse WebSocket message:', error)
      this.emit('error', error)
    }
  }

  private handleOpen(): void {
    this.log('WebSocket connected')
    
    this.state.connected = true
    this.state.connecting = false
    this.state.error = null
    this.state.reconnectAttempts = 0
    
    this.startHeartbeat()
    this.emit('connect')
    
    // Resubscribe to previous channels
    if (this.subscriptions.size > 0) {
      this.subscribe(Array.from(this.subscriptions))
    }
  }

  private handleClose(event: CloseEvent): void {
    this.log('WebSocket disconnected:', event.code, event.reason)
    
    this.state.connected = false
    this.state.connecting = false
    this.stopHeartbeat()
    
    this.emit('disconnect', event.reason || 'Connection closed')
    
    if (this.config.reconnect && this.state.reconnectAttempts < this.config.maxReconnectAttempts) {
      this.scheduleReconnect()
    }
  }

  private handleError(event: Event): void {
    this.log('WebSocket error:', event)
    
    this.state.error = 'Connection error'
    this.emit('error', new Error(this.state.error))
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
    }

    const delay = this.config.reconnectInterval * Math.pow(2, this.state.reconnectAttempts)
    this.log(`Reconnecting in ${delay}ms (attempt ${this.state.reconnectAttempts + 1}/${this.config.maxReconnectAttempts})`)

    this.reconnectTimer = window.setTimeout(() => {
      this.state.reconnectAttempts++
      this.connect()
    }, delay)
  }

  async connect(): Promise<void> {
    if (this.state.connected || this.state.connecting) {
      return
    }

    this.log('Connecting to WebSocket:', this.config.url)
    
    this.state.connecting = true
    this.state.error = null

    try {
      this.ws = new WebSocket(this.config.url, this.config.protocols)
      
      this.ws.addEventListener('open', () => this.handleOpen())
      this.ws.addEventListener('close', (event) => this.handleClose(event))
      this.ws.addEventListener('error', (event) => this.handleError(event))
      this.ws.addEventListener('message', (event) => this.handleMessage(event))
      
      // Wait for connection or timeout
      await new Promise<void>((resolve, reject) => {
        const timeout = setTimeout(() => {
          reject(new Error('Connection timeout'))
        }, this.config.timeout)

        const onConnect = () => {
          clearTimeout(timeout)
          this.off('connect', onConnect)
          this.off('error', onError)
          resolve()
        }

        const onError = (error: Error) => {
          clearTimeout(timeout)
          this.off('connect', onConnect)
          this.off('error', onError)
          reject(error)
        }

        this.on('connect', onConnect)
        this.on('error', onError)
      })
      
    } catch (error) {
      this.state.connecting = false
      this.state.error = (error as Error).message
      throw error
    }
  }

  disconnect(): void {
    this.log('Disconnecting WebSocket')
    
    this.config.reconnect = false
    this.stopHeartbeat()
    
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    
    if (this.ws) {
      this.ws.close(1000, 'Client disconnect')
      this.ws = null
    }
    
    this.state.connected = false
    this.state.connecting = false
    this.subscriptions.clear()
    this.state.subscriptions = new Set()
  }

  subscribe(channels: string[], filters?: Record<string, any>): void {
    if (!this.state.connected) {
      this.log('Cannot subscribe: not connected')
      return
    }

    const message: SubscriptionMessage = {
      action: 'subscribe',
      channels,
      filters,
    }

    this.send({
      type: 'subscribe',
      payload: message,
      timestamp: Date.now(),
    })

    // Optimistically add to subscriptions
    channels.forEach(channel => this.subscriptions.add(channel))
    this.state.subscriptions = new Set(this.subscriptions)
    
    this.log('Subscribed to channels:', channels)
  }

  unsubscribe(channels: string[]): void {
    if (!this.state.connected) {
      this.log('Cannot unsubscribe: not connected')
      return
    }

    const message: SubscriptionMessage = {
      action: 'unsubscribe',
      channels,
    }

    this.send({
      type: 'unsubscribe',
      payload: message,
      timestamp: Date.now(),
    })

    // Remove from subscriptions
    channels.forEach(channel => this.subscriptions.delete(channel))
    this.state.subscriptions = new Set(this.subscriptions)
    
    this.log('Unsubscribed from channels:', channels)
  }

  send(message: WebSocketMessage): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      this.log('Cannot send message: not connected')
      return
    }

    try {
      const data = JSON.stringify(message)
      this.ws.send(data)
      this.log('Sent message:', message)
    } catch (error) {
      console.error('Failed to send WebSocket message:', error)
      this.emit('error', error)
    }
  }

  on(event: string, handler: (data: any) => void): void {
    if (!this.eventHandlers.has(event)) {
      this.eventHandlers.set(event, new Set())
    }
    this.eventHandlers.get(event)!.add(handler)
  }

  off(event: string, handler?: (data: any) => void): void {
    const handlers = this.eventHandlers.get(event)
    if (!handlers) return

    if (handler) {
      handlers.delete(handler)
    } else {
      handlers.clear()
    }
  }

  getState(): WebSocketState {
    return { ...this.state }
  }

  // Convenience methods for common subscriptions
  subscribeToNodeUpdates(): void {
    this.subscribe(['node_status'])
  }

  subscribeToModelSync(): void {
    this.subscribe(['model_sync'])
  }

  subscribeToTaskUpdates(): void {
    this.subscribe(['task_updates'])
  }

  subscribeToClusterMetrics(): void {
    this.subscribe(['cluster_metrics'])
  }

  subscribeToAlerts(): void {
    this.subscribe(['performance_alerts', 'security_alerts'])
  }

  subscribeToNotifications(userId?: string): void {
    const channels = ['system_notifications']
    if (userId) {
      channels.push(`user_notifications:${userId}`)
    }
    this.subscribe(channels)
  }

  // Setup common event handlers
  setupEventHandlers(handlers: Partial<WebSocketEventHandlers>): void {
    Object.entries(handlers).forEach(([event, handler]) => {
      if (handler) {
        const eventName = event.replace(/^on/, '').toLowerCase().replace(/([A-Z])/g, '_$1')
        this.on(eventName, handler)
      }
    })
  }
}

// Create singleton instance with enhanced configuration
export const wsClient = new WebSocketClient({
  url: import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws',
  debug: import.meta.env.DEV,
  reconnect: true,
  maxReconnectAttempts: 5,
  reconnectInterval: 5000,
  heartbeatInterval: 30000,
  timeout: 10000,
})

export default wsClient