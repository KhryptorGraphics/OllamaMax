/**
 * WebSocket Service Module - Comprehensive Real-time Communication
 * 
 * This module provides a complete WebSocket integration system with:
 * - Enhanced connection management with automatic reconnection
 * - Topic-based subscriptions for different page types  
 * - Message routing and event handling
 * - Connection state management and health monitoring
 * - Error recovery and retry logic
 * - Store integration with optimistic updates
 * - Offline support and state synchronization
 */

// Core WebSocket client
export { WebSocketClient, wsClient } from './client'

// Enhanced service layer
export { 
  WebSocketService, 
  webSocketService,
  type WebSocketServiceConfig,
  type SubscriptionOptions,
  type MessageBatch,
} from './service'

// Real-time event handlers
export {
  RealTimeEventHandlers,
  eventHandlers,
  type EventHandlerConfig,
} from './handlers'

// Store integration patterns
export {
  WebSocketStoreIntegration,
  storeIntegration,
  type IntegrationConfig,
  type OptimisticUpdate,
  type StateSync,
} from './integration'

// Connection lifecycle manager
export {
  WebSocketConnectionManager,
  connectionManager,
  type ConnectionManagerConfig,
  type ConnectionMetrics,
  type HealthCheck,
} from './manager'

// Enhanced React hook
export {
  useWebSocket,
  type UseWebSocketOptions,
  type UseWebSocketReturn,
  type WebSocketTopic,
} from '@/hooks/useWebSocket'

// Re-export types for convenience
export type {
  WebSocketState,
  WebSocketMessage,
  WebSocketMessageType,
  WebSocketEventHandlers,
  ConnectionState,
  NodeStatusUpdate,
  ModelSyncUpdate,
  TaskUpdate,
  TransferProgressUpdate,
  ClusterMetricsUpdate,
  PerformanceAlert,
  SecurityAlert,
  SystemNotification,
  UserNotification,
  SubscriptionMessage,
  SubscriptionResponse,
  WebSocketConfig,
  WebSocketClient as IWebSocketClient,
  WebSocketEventHandler,
} from '@/types/websocket'

/**
 * Quick Start Guide:
 * 
 * 1. Basic Usage:
 *    ```typescript
 *    import { useWebSocket } from '@/services/websocket'
 *    
 *    const { isConnected, subscribe, send } = useWebSocket({
 *      topics: ['dashboard', 'models'],
 *      autoConnect: true,
 *    })
 *    ```
 * 
 * 2. Advanced Service Usage:
 *    ```typescript
 *    import { webSocketService, connectionManager } from '@/services/websocket'
 *    
 *    // Connect and subscribe to specific events
 *    await webSocketService.connect()
 *    await webSocketService.subscribeToPage('dashboard')
 *    
 *    // Monitor connection health
 *    const status = connectionManager.getConnectionStatus()
 *    ```
 * 
 * 3. Store Integration:
 *    ```typescript
 *    import { storeIntegration } from '@/services/websocket'
 *    
 *    // Optimistic updates with automatic rollback
 *    await storeIntegration.performOptimisticUpdate(
 *      'deploy-model',
 *      'model',
 *      () => updateUIOptimistically(),
 *      () => deployModelToServer(),
 *      () => rollbackUI()
 *    )
 *    ```
 * 
 * 4. Event Handling:
 *    ```typescript
 *    import { eventHandlers } from '@/services/websocket'
 *    
 *    // Custom event handling with store integration
 *    eventHandlers.handleNodeStatusUpdate(nodeData)
 *    ```
 */

/**
 * Utility functions for common WebSocket operations
 */

// Connection utilities
export const connectWithRetry = async (maxAttempts: number = 3): Promise<void> => {
  let attempt = 0
  
  while (attempt < maxAttempts) {
    try {
      await webSocketService.connect()
      return
    } catch (error) {
      attempt++
      if (attempt >= maxAttempts) {
        throw error
      }
      
      // Exponential backoff
      const delay = 1000 * Math.pow(2, attempt - 1)
      await new Promise(resolve => setTimeout(resolve, delay))
    }
  }
}

// Subscription utilities
export const subscribeToMultipleTopics = async (topics: WebSocketTopic[]): Promise<void> => {
  for (const topic of topics) {
    await webSocketService.subscribeToPage(topic as any)
  }
}

// Health check utilities
export const waitForHealthyConnection = async (timeout: number = 30000): Promise<void> => {
  const startTime = Date.now()
  
  while (Date.now() - startTime < timeout) {
    const status = connectionManager.getConnectionStatus()
    
    if (status.isHealthy && status.state === 'connected') {
      return
    }
    
    await new Promise(resolve => setTimeout(resolve, 1000))
  }
  
  throw new Error('Connection did not become healthy within timeout')
}

// Batch operation utilities
export const sendBatchMessages = async (messages: WebSocketMessage[]): Promise<void> => {
  return webSocketService.sendBatch(messages)
}

// Error handling utilities
export const handleConnectionError = (error: Error): void => {
  console.error('[WebSocket] Connection error:', error)
  
  // Attempt recovery
  connectionManager.recover().catch(recoveryError => {
    console.error('[WebSocket] Recovery failed:', recoveryError)
  })
}

/**
 * Default export for easy integration
 */
export default {
  // Services
  client: wsClient,
  service: webSocketService,
  connectionManager,
  storeIntegration,
  eventHandlers,
  
  // Utilities
  connectWithRetry,
  subscribeToMultipleTopics,
  waitForHealthyConnection,
  sendBatchMessages,
  handleConnectionError,
  
  // Hook
  useWebSocket,
}