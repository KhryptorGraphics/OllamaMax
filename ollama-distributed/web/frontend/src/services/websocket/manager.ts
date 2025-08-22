/**
 * WebSocket Connection Manager
 * Provides unified lifecycle management, connection health monitoring,
 * and automatic recovery for WebSocket connections
 */

import { webSocketService } from './service'
import { storeIntegration } from './integration'
import { eventHandlers } from './handlers'
import { wsClient } from './client'
import { useStore } from '@/stores'
import type {
  WebSocketState,
  WebSocketMessage,
  ConnectionState,
} from '@/types/websocket'

export interface ConnectionManagerConfig {
  enableHealthMonitoring?: boolean
  enableAutoRecovery?: boolean
  enableConnectionMetrics?: boolean
  healthCheckInterval?: number
  connectionTimeout?: number
  maxReconnectAttempts?: number
  reconnectDelay?: number
  pingInterval?: number
  pongTimeout?: number
}

export interface ConnectionMetrics {
  connectTime: number
  disconnectTime: number | null
  totalConnections: number
  totalDisconnections: number
  totalErrors: number
  avgLatency: number
  messagesSent: number
  messagesReceived: number
  bytesTransferred: number
  uptime: number
  connectionQuality: 'excellent' | 'good' | 'poor' | 'offline'
}

export interface HealthCheck {
  timestamp: number
  latency: number
  status: 'healthy' | 'degraded' | 'unhealthy'
  errors: string[]
}

export class WebSocketConnectionManager {
  private config: Required<ConnectionManagerConfig>
  private metrics: ConnectionMetrics
  private healthChecks: HealthCheck[] = []
  private healthCheckTimer: NodeJS.Timeout | null = null
  private pingTimer: NodeJS.Timeout | null = null
  private pongTimer: NodeJS.Timeout | null = null
  private reconnectTimer: NodeJS.Timeout | null = null
  private lastPingTime: number = 0
  private isInitialized: boolean = false
  private connectionPromise: Promise<void> | null = null

  constructor(config: ConnectionManagerConfig = {}) {
    this.config = {
      enableHealthMonitoring: true,
      enableAutoRecovery: true,
      enableConnectionMetrics: true,
      healthCheckInterval: 30000, // 30 seconds
      connectionTimeout: 10000,   // 10 seconds
      maxReconnectAttempts: 5,
      reconnectDelay: 1000,
      pingInterval: 25000,        // 25 seconds
      pongTimeout: 5000,          // 5 seconds
      ...config,
    }

    this.metrics = this.initializeMetrics()
    this.initialize()
  }

  /**
   * Initialize the connection manager
   */
  private async initialize(): Promise<void> {
    if (this.isInitialized) return

    try {
      // Set up event handlers
      this.setupEventHandlers()

      // Start health monitoring if enabled
      if (this.config.enableHealthMonitoring) {
        this.startHealthMonitoring()
      }

      // Set up ping/pong for connection health
      this.setupPingPong()

      this.isInitialized = true
      console.log('[ConnectionManager] Initialized successfully')

    } catch (error) {
      console.error('[ConnectionManager] Failed to initialize:', error)
      throw error
    }
  }

  /**
   * Connect with enhanced error handling and recovery
   */
  public async connect(): Promise<void> {
    if (this.connectionPromise) {
      return this.connectionPromise
    }

    this.connectionPromise = this.performConnect()
    return this.connectionPromise
  }

  private async performConnect(): Promise<void> {
    const store = useStore.getState()

    try {
      // Check if user is authenticated
      if (!store.auth.isAuthenticated) {
        throw new Error('User not authenticated')
      }

      // Record connection attempt
      this.metrics.totalConnections++
      this.metrics.connectTime = Date.now()

      // Connect to WebSocket
      await webSocketService.connect()

      // Start ping/pong monitoring
      this.startPingPong()

      // Initialize store integration
      this.initializeStoreIntegration()

      console.log('[ConnectionManager] Connected successfully')

    } catch (error) {
      this.metrics.totalErrors++
      console.error('[ConnectionManager] Connection failed:', error)

      // Auto-recovery if enabled
      if (this.config.enableAutoRecovery) {
        this.scheduleReconnect()
      }

      throw error
    } finally {
      this.connectionPromise = null
    }
  }

  /**
   * Disconnect with cleanup
   */
  public async disconnect(): Promise<void> {
    console.log('[ConnectionManager] Disconnecting...')

    // Stop all timers
    this.stopHealthMonitoring()
    this.stopPingPong()
    this.clearReconnectTimer()

    // Record disconnection
    this.metrics.disconnectTime = Date.now()
    this.metrics.totalDisconnections++

    // Disconnect services
    webSocketService.disconnect()
    storeIntegration.cleanup()

    console.log('[ConnectionManager] Disconnected')
  }

  /**
   * Get current connection status
   */
  public getConnectionStatus(): {
    state: ConnectionState
    metrics: ConnectionMetrics
    health: HealthCheck | null
    isHealthy: boolean
  } {
    const wsState = wsClient.getState()
    const latestHealth = this.healthChecks[this.healthChecks.length - 1] || null

    return {
      state: this.getConnectionState(),
      metrics: { ...this.metrics },
      health: latestHealth,
      isHealthy: latestHealth?.status === 'healthy' || false,
    }
  }

  /**
   * Force a health check
   */
  public async performHealthCheck(): Promise<HealthCheck> {
    const startTime = Date.now()
    const errors: string[] = []

    try {
      // Send ping and wait for pong
      await this.sendPing()
      
      const latency = Date.now() - startTime
      const status = this.determineHealthStatus(latency, errors)

      const healthCheck: HealthCheck = {
        timestamp: Date.now(),
        latency,
        status,
        errors,
      }

      // Store health check (keep last 10)
      this.healthChecks.push(healthCheck)
      if (this.healthChecks.length > 10) {
        this.healthChecks.shift()
      }

      // Update metrics
      this.updateLatencyMetrics(latency)

      return healthCheck

    } catch (error) {
      errors.push((error as Error).message)
      
      const healthCheck: HealthCheck = {
        timestamp: Date.now(),
        latency: Date.now() - startTime,
        status: 'unhealthy',
        errors,
      }

      this.healthChecks.push(healthCheck)
      return healthCheck
    }
  }

  /**
   * Get connection quality based on recent metrics
   */
  public getConnectionQuality(): 'excellent' | 'good' | 'poor' | 'offline' {
    const wsState = wsClient.getState()
    
    if (!wsState.connected) {
      return 'offline'
    }

    const recentChecks = this.healthChecks.slice(-5)
    if (recentChecks.length === 0) {
      return 'good' // Default for new connections
    }

    const avgLatency = recentChecks.reduce((sum, check) => sum + check.latency, 0) / recentChecks.length
    const healthyChecks = recentChecks.filter(check => check.status === 'healthy').length
    const healthRatio = healthyChecks / recentChecks.length

    if (avgLatency < 100 && healthRatio >= 0.8) {
      return 'excellent'
    } else if (avgLatency < 500 && healthRatio >= 0.6) {
      return 'good'
    } else {
      return 'poor'
    }
  }

  /**
   * Handle connection recovery after network issues
   */
  public async recover(): Promise<void> {
    console.log('[ConnectionManager] Starting recovery...')

    try {
      // Disconnect first to clean up any stale connections
      await this.disconnect()

      // Wait a moment before reconnecting
      await new Promise(resolve => setTimeout(resolve, 1000))

      // Reconnect
      await this.connect()

      // Trigger state synchronization
      await storeIntegration.synchronizeState()

      console.log('[ConnectionManager] Recovery completed successfully')

    } catch (error) {
      console.error('[ConnectionManager] Recovery failed:', error)
      throw error
    }
  }

  /**
   * Get detailed connection metrics
   */
  public getMetrics(): ConnectionMetrics {
    // Update uptime
    if (this.metrics.connectTime > 0 && !this.metrics.disconnectTime) {
      this.metrics.uptime = Date.now() - this.metrics.connectTime
    }

    // Update connection quality
    this.metrics.connectionQuality = this.getConnectionQuality()

    return { ...this.metrics }
  }

  /**
   * Reset metrics
   */
  public resetMetrics(): void {
    this.metrics = this.initializeMetrics()
    this.healthChecks = []
  }

  /**
   * Subscribe to connection events
   */
  public onConnectionChange(callback: (isConnected: boolean) => void): () => void {
    const handleConnect = () => callback(true)
    const handleDisconnect = () => callback(false)

    wsClient.on('connect', handleConnect)
    wsClient.on('disconnect', handleDisconnect)

    return () => {
      wsClient.off('connect', handleConnect)
      wsClient.off('disconnect', handleDisconnect)
    }
  }

  /**
   * Private helper methods
   */
  private initializeMetrics(): ConnectionMetrics {
    return {
      connectTime: 0,
      disconnectTime: null,
      totalConnections: 0,
      totalDisconnections: 0,
      totalErrors: 0,
      avgLatency: 0,
      messagesSent: 0,
      messagesReceived: 0,
      bytesTransferred: 0,
      uptime: 0,
      connectionQuality: 'offline',
    }
  }

  private setupEventHandlers(): void {
    wsClient.on('connect', () => {
      console.log('[ConnectionManager] WebSocket connected')
      
      // Reset disconnect time
      this.metrics.disconnectTime = null
      
      // Start ping/pong
      this.startPingPong()
      
      // Notify store integration
      storeIntegration.handleConnectionChange(true)
    })

    wsClient.on('disconnect', () => {
      console.log('[ConnectionManager] WebSocket disconnected')
      
      // Record disconnect time
      this.metrics.disconnectTime = Date.now()
      this.metrics.totalDisconnections++
      
      // Stop ping/pong
      this.stopPingPong()
      
      // Notify store integration
      storeIntegration.handleConnectionChange(false)
      
      // Auto-recovery if enabled
      if (this.config.enableAutoRecovery) {
        this.scheduleReconnect()
      }
    })

    wsClient.on('error', (error: Error) => {
      console.error('[ConnectionManager] WebSocket error:', error)
      this.metrics.totalErrors++
    })

    wsClient.on('message', (message: WebSocketMessage) => {
      this.metrics.messagesReceived++
      
      // Handle pong messages
      if (message.type === 'pong') {
        this.handlePong(message.timestamp)
      }
    })
  }

  private startHealthMonitoring(): void {
    if (this.healthCheckTimer) {
      clearInterval(this.healthCheckTimer)
    }

    this.healthCheckTimer = setInterval(() => {
      if (wsClient.getState().connected) {
        this.performHealthCheck().catch(console.error)
      }
    }, this.config.healthCheckInterval)
  }

  private stopHealthMonitoring(): void {
    if (this.healthCheckTimer) {
      clearInterval(this.healthCheckTimer)
      this.healthCheckTimer = null
    }
  }

  private setupPingPong(): void {
    wsClient.on('pong', (data: any) => {
      this.handlePong(data.timestamp)
    })
  }

  private startPingPong(): void {
    if (this.pingTimer) {
      clearInterval(this.pingTimer)
    }

    this.pingTimer = setInterval(() => {
      if (wsClient.getState().connected) {
        this.sendPing().catch(console.error)
      }
    }, this.config.pingInterval)
  }

  private stopPingPong(): void {
    if (this.pingTimer) {
      clearInterval(this.pingTimer)
      this.pingTimer = null
    }
    
    if (this.pongTimer) {
      clearTimeout(this.pongTimer)
      this.pongTimer = null
    }
  }

  private async sendPing(): Promise<void> {
    return new Promise((resolve, reject) => {
      this.lastPingTime = Date.now()

      const ping: WebSocketMessage = {
        type: 'ping',
        payload: { timestamp: this.lastPingTime },
        timestamp: this.lastPingTime,
      }

      // Set timeout for pong response
      this.pongTimer = setTimeout(() => {
        reject(new Error('Pong timeout'))
      }, this.config.pongTimeout)

      // Send ping
      try {
        wsClient.send(ping)
        this.metrics.messagesSent++
        
        // Resolve immediately since we'll handle pong separately
        resolve()
      } catch (error) {
        if (this.pongTimer) {
          clearTimeout(this.pongTimer)
        }
        reject(error)
      }
    })
  }

  private handlePong(timestamp: number): void {
    if (this.pongTimer) {
      clearTimeout(this.pongTimer)
      this.pongTimer = null
    }

    const latency = Date.now() - timestamp
    this.updateLatencyMetrics(latency)
  }

  private updateLatencyMetrics(latency: number): void {
    // Simple moving average for latency
    const alpha = 0.1 // Smoothing factor
    this.metrics.avgLatency = this.metrics.avgLatency * (1 - alpha) + latency * alpha
  }

  private determineHealthStatus(latency: number, errors: string[]): 'healthy' | 'degraded' | 'unhealthy' {
    if (errors.length > 0) {
      return 'unhealthy'
    }
    
    if (latency > 1000) {
      return 'degraded'
    }
    
    return 'healthy'
  }

  private getConnectionState(): ConnectionState {
    const wsState = wsClient.getState()
    
    if (wsState.connected) return 'connected'
    if (wsState.connecting) return 'connecting'
    if (wsState.error) return 'error'
    return 'disconnected'
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
    }

    const wsState = wsClient.getState()
    const attempt = wsState.reconnectAttempts

    if (attempt >= this.config.maxReconnectAttempts) {
      console.log('[ConnectionManager] Max reconnect attempts reached')
      return
    }

    const delay = this.config.reconnectDelay * Math.pow(2, attempt)
    console.log(`[ConnectionManager] Scheduling reconnect in ${delay}ms (attempt ${attempt + 1})`)

    this.reconnectTimer = setTimeout(() => {
      this.connect().catch(error => {
        console.error('[ConnectionManager] Reconnect failed:', error)
      })
    }, delay)
  }

  private clearReconnectTimer(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
  }

  private initializeStoreIntegration(): void {
    // Set up page-specific integrations based on current route
    const currentPath = window.location.pathname
    
    if (currentPath.includes('/dashboard')) {
      storeIntegration.setupPageIntegration('dashboard')
    } else if (currentPath.includes('/models')) {
      storeIntegration.setupPageIntegration('models')
    } else if (currentPath.includes('/nodes')) {
      storeIntegration.setupPageIntegration('nodes')
    } else if (currentPath.includes('/monitoring')) {
      storeIntegration.setupPageIntegration('monitoring')
    }
  }

  /**
   * Cleanup method
   */
  public cleanup(): void {
    this.stopHealthMonitoring()
    this.stopPingPong()
    this.clearReconnectTimer()
    
    storeIntegration.cleanup()
    eventHandlers.cleanup()
  }
}

// Create singleton instance
export const connectionManager = new WebSocketConnectionManager({
  enableHealthMonitoring: true,
  enableAutoRecovery: true,
  enableConnectionMetrics: true,
  maxReconnectAttempts: 5,
})

export default connectionManager