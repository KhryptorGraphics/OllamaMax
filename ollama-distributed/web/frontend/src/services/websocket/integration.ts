/**
 * WebSocket Store Integration Patterns
 * Provides patterns for connecting WebSocket events to store state management
 * with optimistic updates, rollbacks, and state synchronization
 */

import { useStore } from '@/stores'
import { webSocketService } from './service'
import { eventHandlers } from './handlers'
import type {
  WebSocketMessage,
  NodeStatusUpdate,
  ModelSyncUpdate,
  TaskUpdate,
  TransferProgressUpdate,
} from '@/types/websocket'

export interface IntegrationConfig {
  enableOptimisticUpdates?: boolean
  enableRollbacks?: boolean
  enableStateSynchronization?: boolean
  enableConflictResolution?: boolean
  enableOfflineSupport?: boolean
  syncInterval?: number
  maxRetries?: number
}

export interface OptimisticUpdate<T> {
  id: string
  type: string
  originalState: T
  optimisticState: T
  timestamp: number
  rollbackFn: () => void
}

export interface StateSync {
  lastSyncTimestamp: number
  pendingChanges: Map<string, any>
  conflictResolution: 'server-wins' | 'client-wins' | 'merge'
}

export class WebSocketStoreIntegration {
  private config: Required<IntegrationConfig>
  private optimisticUpdates: Map<string, OptimisticUpdate<any>> = new Map()
  private syncState: StateSync
  private syncTimer: NodeJS.Timeout | null = null
  private retryQueue: Map<string, { message: WebSocketMessage; retries: number }> = new Map()

  constructor(config: IntegrationConfig = {}) {
    this.config = {
      enableOptimisticUpdates: true,
      enableRollbacks: true,
      enableStateSynchronization: true,
      enableConflictResolution: true,
      enableOfflineSupport: true,
      syncInterval: 30000, // 30 seconds
      maxRetries: 3,
      ...config,
    }

    this.syncState = {
      lastSyncTimestamp: Date.now(),
      pendingChanges: new Map(),
      conflictResolution: 'server-wins',
    }

    this.initialize()
  }

  /**
   * Initialize integration patterns
   */
  private initialize(): void {
    this.setupStoreSubscriptions()
    this.setupWebSocketHandlers()
    this.setupStateSynchronization()
    this.setupOfflineSupport()
  }

  /**
   * Optimistic update pattern for immediate UI feedback
   */
  public async performOptimisticUpdate<T>(
    id: string,
    type: string,
    optimisticAction: () => void,
    serverAction: () => Promise<T>,
    rollbackAction: () => void
  ): Promise<T> {
    if (!this.config.enableOptimisticUpdates) {
      return serverAction()
    }

    const store = useStore.getState()
    const originalState = this.captureRelevantState(type)

    try {
      // Apply optimistic update immediately
      optimisticAction()

      // Track the optimistic update
      this.optimisticUpdates.set(id, {
        id,
        type,
        originalState,
        optimisticState: this.captureRelevantState(type),
        timestamp: Date.now(),
        rollbackFn: rollbackAction,
      })

      // Perform server action
      const result = await serverAction()

      // Remove from optimistic updates on success
      this.optimisticUpdates.delete(id)

      return result

    } catch (error) {
      // Rollback optimistic update on failure
      if (this.config.enableRollbacks) {
        this.rollbackOptimisticUpdate(id)
      }

      // Show error notification
      store.addNotification({
        type: 'error',
        title: 'Operation Failed',
        message: `Failed to ${type}: ${(error as Error).message}`,
        duration: 10000,
      })

      throw error
    }
  }

  /**
   * Rollback optimistic update
   */
  public rollbackOptimisticUpdate(id: string): void {
    const update = this.optimisticUpdates.get(id)
    if (update) {
      update.rollbackFn()
      this.optimisticUpdates.delete(id)
      console.log(`Rolled back optimistic update: ${id}`)
    }
  }

  /**
   * Handle connection state changes
   */
  public handleConnectionChange(isConnected: boolean): void {
    const store = useStore.getState()

    if (isConnected) {
      // Re-sync state when reconnected
      if (this.config.enableStateSynchronization) {
        this.synchronizeState()
      }

      // Process retry queue
      this.processRetryQueue()

      store.addNotification({
        type: 'success',
        title: 'Connection Restored',
        message: 'Real-time updates have been restored',
        duration: 3000,
      })
    } else {
      // Handle offline mode
      if (this.config.enableOfflineSupport) {
        this.enableOfflineMode()
      }

      store.addNotification({
        type: 'warning',
        title: 'Connection Lost',
        message: 'Operating in offline mode. Changes will sync when connection is restored.',
        duration: 5000,
      })
    }
  }

  /**
   * Synchronize local state with server state
   */
  public async synchronizeState(): Promise<void> {
    if (!this.config.enableStateSynchronization) return

    const store = useStore.getState()

    try {
      // Fetch latest state from server
      await Promise.all([
        store.fetchNodes(),
        store.fetchModels(),
        store.fetchMetrics(),
      ])

      // Update sync timestamp
      this.syncState.lastSyncTimestamp = Date.now()
      this.syncState.pendingChanges.clear()

      console.log('State synchronized successfully')

    } catch (error) {
      console.error('Failed to synchronize state:', error)
      
      // Retry synchronization
      setTimeout(() => this.synchronizeState(), 5000)
    }
  }

  /**
   * Handle message sending with retry logic
   */
  public async sendMessageWithRetry(
    message: WebSocketMessage,
    maxRetries: number = this.config.maxRetries
  ): Promise<void> {
    const messageId = message.id || `${message.type}-${Date.now()}`

    try {
      await webSocketService.send(message)
      
      // Remove from retry queue on success
      this.retryQueue.delete(messageId)
      
    } catch (error) {
      console.error('Failed to send message:', error)
      
      // Add to retry queue
      const retryEntry = this.retryQueue.get(messageId) || { message, retries: 0 }
      
      if (retryEntry.retries < maxRetries) {
        retryEntry.retries++
        this.retryQueue.set(messageId, retryEntry)
        
        // Schedule retry with exponential backoff
        const delay = 1000 * Math.pow(2, retryEntry.retries - 1)
        setTimeout(() => {
          this.sendMessageWithRetry(message, maxRetries)
        }, delay)
      } else {
        // Max retries exceeded
        this.retryQueue.delete(messageId)
        throw new Error(`Failed to send message after ${maxRetries} retries`)
      }
    }
  }

  /**
   * Private helper methods
   */
  private setupStoreSubscriptions(): void {
    // Subscribe to auth state changes
    useStore.subscribe(
      (state) => state.auth.isAuthenticated,
      (isAuthenticated) => {
        if (isAuthenticated) {
          webSocketService.connect()
        } else {
          webSocketService.disconnect()
          this.clearOptimisticUpdates()
        }
      }
    )
  }

  private setupWebSocketHandlers(): void {
    // Enhanced message handling with store integration
    webSocketService.subscribe(['*'], { persistent: true })
  }

  private setupStateSynchronization(): void {
    if (!this.config.enableStateSynchronization) return

    this.syncTimer = setInterval(() => {
      if (webSocketService.getStatus().connected) {
        this.synchronizeState()
      }
    }, this.config.syncInterval)
  }

  private setupOfflineSupport(): void {
    if (!this.config.enableOfflineSupport) return

    // Listen for online/offline events
    window.addEventListener('online', () => {
      this.handleConnectionChange(true)
    })

    window.addEventListener('offline', () => {
      this.handleConnectionChange(false)
    })
  }

  private captureRelevantState(type: string): any {
    const store = useStore.getState()

    switch (type) {
      case 'node':
        return { nodes: store.nodes }
      case 'model':
        return { models: store.models }
      case 'task':
        return { tasks: store.tasks }
      default:
        return {}
    }
  }

  private enableOfflineMode(): void {
    // Cache current state for offline usage
    const store = useStore.getState()
    const offlineState = {
      nodes: store.nodes,
      models: store.models,
      metrics: store.monitoring.metrics,
      timestamp: Date.now(),
    }

    localStorage.setItem('ollama-offline-state', JSON.stringify(offlineState))
  }

  private processRetryQueue(): void {
    this.retryQueue.forEach((entry, messageId) => {
      this.sendMessageWithRetry(entry.message)
    })
  }

  private clearOptimisticUpdates(): void {
    this.optimisticUpdates.clear()
  }

  /**
   * Cleanup method
   */
  public cleanup(): void {
    if (this.syncTimer) {
      clearInterval(this.syncTimer)
    }
    
    this.clearOptimisticUpdates()
    this.retryQueue.clear()
    eventHandlers.cleanup()
  }
}

// Create singleton instance
export const storeIntegration = new WebSocketStoreIntegration({
  enableOptimisticUpdates: true,
  enableRollbacks: true,
  enableStateSynchronization: true,
  enableConflictResolution: true,
  enableOfflineSupport: true,
})

export default storeIntegration