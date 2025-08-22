import { useState, useEffect, useCallback, useRef } from 'react'
import { wsClient } from '@/services/websocket/client'
import { useStore } from '@/stores'
import type {
  WebSocketState,
  WebSocketMessage,
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
} from '@/types/websocket'

interface UseWebSocketOptions {
  autoConnect?: boolean
  autoReconnect?: boolean
  debug?: boolean
  subscriptions?: string[]
  eventHandlers?: Partial<WebSocketEventHandlers>
  topics?: WebSocketTopic[]
}

export type WebSocketTopic = 
  | 'dashboard'
  | 'models' 
  | 'nodes'
  | 'monitoring'
  | 'tasks'
  | 'transfers'
  | 'alerts'
  | 'notifications'

interface UseWebSocketReturn {
  // Connection state
  connectionState: ConnectionState
  isConnected: boolean
  isConnecting: boolean
  error: string | null
  lastMessage: WebSocketMessage | null
  
  // Connection management
  connect: () => Promise<void>
  disconnect: () => void
  reconnect: () => Promise<void>
  
  // Subscription management
  subscribe: (channels: string[], filters?: Record<string, any>) => void
  unsubscribe: (channels: string[]) => void
  subscribeToTopic: (topic: WebSocketTopic) => void
  unsubscribeFromTopic: (topic: WebSocketTopic) => void
  
  // Message handling
  send: (message: WebSocketMessage) => void
  on: (event: string, handler: (data: any) => void) => void
  off: (event: string, handler?: (data: any) => void) => void
  
  // State information
  getSubscriptions: () => Set<string>
  getConnectionInfo: () => WebSocketState
}

// Topic-based subscription mapping
const TOPIC_CHANNELS: Record<WebSocketTopic, string[]> = {
  dashboard: ['dashboard_summary', 'system_notifications', 'activity_feed'],
  models: ['model_sync_update', 'model_deployment', 'model_status'],
  nodes: ['node_status_update', 'node_health', 'cluster_topology'],
  monitoring: ['cluster_metrics', 'performance_alerts', 'system_metrics'],
  tasks: ['task_update', 'task_progress', 'task_completion'],
  transfers: ['transfer_progress', 'transfer_status', 'bandwidth_metrics'],
  alerts: ['performance_alert', 'security_alert', 'system_alert'],
  notifications: ['user_notification', 'system_notification', 'broadcast_message'],
}

export const useWebSocket = (options: UseWebSocketOptions = {}): UseWebSocketReturn => {
  const {
    autoConnect = true,
    autoReconnect = true,
    debug = false,
    subscriptions = [],
    eventHandlers = {},
    topics = [],
  } = options

  // State management
  const [connectionState, setConnectionState] = useState<ConnectionState>('disconnected')
  const [lastMessage, setLastMessage] = useState<WebSocketMessage | null>(null)
  const subscribedChannels = useRef<Set<string>>(new Set())
  const eventListeners = useRef<Map<string, Set<Function>>>(new Map())
  const messageQueue = useRef<WebSocketMessage[]>([])
  const reconnectTimer = useRef<NodeJS.Timeout | null>(null)
  const heartbeatTimer = useRef<NodeJS.Timeout | null>(null)
  
  // Store integration
  const {
    auth,
    updateConnectionState,
    updateNodeStatus,
    updateModelSync,
    updateTaskStatus,
    updateTransferProgress,
    updateMetrics,
    addAlert,
    addNotification,
    addRecentActivity,
  } = useStore()

  // Enhanced connection management with retry logic
  const connect = useCallback(async (): Promise<void> => {
    if (connectionState === 'connected' || connectionState === 'connecting') {
      return
    }

    try {
      setConnectionState('connecting')
      updateConnectionState({ connecting: true, error: null })
      
      await wsClient.connect()
      
      setConnectionState('connected')
      updateConnectionState({ 
        connected: true, 
        connecting: false, 
        error: null,
        reconnectAttempts: 0 
      })
      
      // Process queued messages
      if (messageQueue.current.length > 0) {
        messageQueue.current.forEach(message => wsClient.send(message))
        messageQueue.current = []
      }
      
      // Resubscribe to channels
      if (subscribedChannels.current.size > 0) {
        wsClient.subscribe(Array.from(subscribedChannels.current))
      }
      
      // Subscribe to initial topics
      topics.forEach(topic => subscribeToTopic(topic))
      
      // Subscribe to initial channels
      if (subscriptions.length > 0) {
        subscribe(subscriptions)
      }
      
      // Start heartbeat
      startHeartbeat()
      
      if (debug) {
        console.log('[WebSocket] Connected successfully')
      }
      
    } catch (error) {
      setConnectionState('error')
      const errorMessage = (error as Error).message
      updateConnectionState({ 
        connecting: false, 
        error: errorMessage 
      })
      
      if (debug) {
        console.error('[WebSocket] Connection failed:', errorMessage)
      }
      
      // Auto-reconnect if enabled
      if (autoReconnect) {
        scheduleReconnect()
      }
      
      throw error
    }
  }, [connectionState, autoReconnect, debug, topics, subscriptions])

  const disconnect = useCallback((): void => {
    if (reconnectTimer.current) {
      clearTimeout(reconnectTimer.current)
      reconnectTimer.current = null
    }
    
    if (heartbeatTimer.current) {
      clearInterval(heartbeatTimer.current)
      heartbeatTimer.current = null
    }
    
    wsClient.disconnect()
    setConnectionState('disconnected')
    updateConnectionState({ 
      connected: false, 
      connecting: false,
      subscriptions: new Set()
    })
    subscribedChannels.current.clear()
    
    if (debug) {
      console.log('[WebSocket] Disconnected')
    }
  }, [debug])

  const reconnect = useCallback(async (): Promise<void> => {
    disconnect()
    await new Promise(resolve => setTimeout(resolve, 1000))
    return connect()
  }, [connect, disconnect])

  // Enhanced subscription management with topic support
  const subscribe = useCallback((channels: string[], filters?: Record<string, any>): void => {
    if (connectionState !== 'connected') {
      channels.forEach(channel => subscribedChannels.current.add(channel))
      return
    }
    
    wsClient.subscribe(channels, filters)
    channels.forEach(channel => subscribedChannels.current.add(channel))
    updateConnectionState({ 
      subscriptions: new Set(subscribedChannels.current) 
    })
    
    if (debug) {
      console.log('[WebSocket] Subscribed to channels:', channels)
    }
  }, [connectionState, debug])

  const unsubscribe = useCallback((channels: string[]): void => {
    if (connectionState === 'connected') {
      wsClient.unsubscribe(channels)
    }
    
    channels.forEach(channel => subscribedChannels.current.delete(channel))
    updateConnectionState({ 
      subscriptions: new Set(subscribedChannels.current) 
    })
    
    if (debug) {
      console.log('[WebSocket] Unsubscribed from channels:', channels)
    }
  }, [connectionState, debug])

  const subscribeToTopic = useCallback((topic: WebSocketTopic): void => {
    const channels = TOPIC_CHANNELS[topic] || []
    subscribe(channels)
    
    if (debug) {
      console.log(`[WebSocket] Subscribed to topic '${topic}':`, channels)
    }
  }, [subscribe, debug])

  const unsubscribeFromTopic = useCallback((topic: WebSocketTopic): void => {
    const channels = TOPIC_CHANNELS[topic] || []
    unsubscribe(channels)
    
    if (debug) {
      console.log(`[WebSocket] Unsubscribed from topic '${topic}':`, channels)
    }
  }, [unsubscribe, debug])

  // Enhanced message handling with queuing
  const send = useCallback((message: WebSocketMessage): void => {
    if (connectionState !== 'connected') {
      // Queue message for when connection is restored
      messageQueue.current.push(message)
      
      if (debug) {
        console.log('[WebSocket] Message queued (not connected):', message.type)
      }
      return
    }
    
    wsClient.send(message)
    
    if (debug) {
      console.log('[WebSocket] Message sent:', message.type)
    }
  }, [connectionState, debug])

  // Event handler management
  const on = useCallback((event: string, handler: (data: any) => void): void => {
    if (!eventListeners.current.has(event)) {
      eventListeners.current.set(event, new Set())
    }
    eventListeners.current.get(event)!.add(handler)
    wsClient.on(event, handler)
  }, [])

  const off = useCallback((event: string, handler?: (data: any) => void): void => {
    const handlers = eventListeners.current.get(event)
    if (handlers) {
      if (handler) {
        handlers.delete(handler)
        wsClient.off(event, handler)
      } else {
        handlers.clear()
        wsClient.off(event)
      }
    }
  }, [])

  // Utility functions
  const getSubscriptions = useCallback((): Set<string> => {
    return new Set(subscribedChannels.current)
  }, [])

  const getConnectionInfo = useCallback((): WebSocketState => {
    return wsClient.getState()
  }, [])

  // Heartbeat mechanism
  const startHeartbeat = useCallback((): void => {
    if (heartbeatTimer.current) {
      clearInterval(heartbeatTimer.current)
    }
    
    heartbeatTimer.current = setInterval(() => {
      if (connectionState === 'connected') {
        send({
          type: 'ping',
          payload: { timestamp: Date.now() },
          timestamp: Date.now(),
        })
      }
    }, 30000) // 30 seconds
  }, [connectionState, send])

  // Reconnection with exponential backoff
  const scheduleReconnect = useCallback((): void => {
    if (reconnectTimer.current) {
      clearTimeout(reconnectTimer.current)
    }
    
    const state = wsClient.getState()
    const delay = Math.min(1000 * Math.pow(2, state.reconnectAttempts), 30000)
    
    if (debug) {
      console.log(`[WebSocket] Reconnecting in ${delay}ms (attempt ${state.reconnectAttempts + 1})`) 
    }
    
    reconnectTimer.current = setTimeout(() => {
      updateConnectionState({ 
        reconnectAttempts: state.reconnectAttempts + 1 
      })
      connect().catch(console.error)
    }, delay)
  }, [connect, debug])

  // Set up WebSocket event handlers
  useEffect(() => {
    const handleConnect = () => {
      setConnectionState('connected')
      updateConnectionState({ 
        connected: true, 
        connecting: false, 
        error: null,
        reconnectAttempts: 0
      })
    }

    const handleDisconnect = (reason?: string) => {
      setConnectionState('disconnected')
      updateConnectionState({ 
        connected: false, 
        connecting: false 
      })
      
      if (autoReconnect && auth.isAuthenticated) {
        scheduleReconnect()
      }
    }

    const handleError = (error: Error) => {
      setConnectionState('error')
      updateConnectionState({ 
        error: error.message,
        connecting: false 
      })
    }

    const handleMessage = (message: WebSocketMessage) => {
      setLastMessage(message)
      updateConnectionState({ lastMessage: message })
      
      // Route messages to store actions
      routeMessageToStore(message)
    }

    // Enhanced event handlers with store integration
    const enhancedHandlers: Partial<WebSocketEventHandlers> = {
      onConnect: handleConnect,
      onDisconnect: handleDisconnect,
      onError: handleError,
      onMessage: handleMessage,
      
      onNodeStatusUpdate: (data: NodeStatusUpdate) => {
        updateNodeStatus(data.nodeId, {
          status: data.status,
          health: data.health,
          lastSeen: data.timestamp,
        })
      },
      
      onModelSyncUpdate: (data: ModelSyncUpdate) => {
        updateModelSync(data.modelName, {
          status: data.status,
          progress: data.progress,
          nodesTotal: 1,
          nodesSynced: data.status === 'completed' ? 1 : 0,
          error: data.error,
        })
      },
      
      onTaskUpdate: (data: TaskUpdate) => {
        updateTaskStatus(data.taskId, data)
      },
      
      onTransferProgress: (data: TransferProgressUpdate) => {
        updateTransferProgress(data.transferId, data.progress)
      },
      
      onClusterMetrics: (data: ClusterMetricsUpdate) => {
        updateMetrics(data)
      },
      
      onPerformanceAlert: (data: PerformanceAlert) => {
        addAlert(data)
        addNotification({
          type: data.severity === 'critical' ? 'error' : 'warning',
          title: 'Performance Alert',
          message: data.message,
          duration: data.severity === 'critical' ? 0 : 10000,
        })
      },
      
      onSecurityAlert: (data: SecurityAlert) => {
        addAlert(data)
        addNotification({
          type: 'error',
          title: 'Security Alert',
          message: data.message,
          duration: 0, // Never auto-dismiss security alerts
        })
      },
      
      onSystemNotification: (data: SystemNotification) => {
        addNotification({
          type: data.severity === 'error' ? 'error' : data.severity === 'warning' ? 'warning' : 'info',
          title: data.title,
          message: data.message,
          duration: data.severity === 'error' ? 0 : 10000,
        })
      },
      
      onUserNotification: (data: UserNotification) => {
        addNotification({
          type: 'info',
          title: data.title,
          message: data.message,
          duration: 15000,
          actions: data.actionUrl ? [{
            label: data.actionLabel || 'View',
            action: () => window.open(data.actionUrl, '_blank'),
          }] : undefined,
        })
      },
      
      // Merge with user-provided handlers
      ...eventHandlers,
    }

    // Set up all event handlers
    wsClient.setupEventHandlers(enhancedHandlers)

    return () => {
      // Cleanup event handlers
      Object.keys(enhancedHandlers).forEach(event => {
        const handler = enhancedHandlers[event as keyof WebSocketEventHandlers]
        if (handler) {
          const eventName = event.replace(/^on/, '').toLowerCase().replace(/([A-Z])/g, '_$1')
          wsClient.off(eventName, handler)
        }
      })
    }
  }, [auth.isAuthenticated, autoReconnect, eventHandlers])

  // Message routing to store actions
  const routeMessageToStore = useCallback((message: WebSocketMessage) => {
    switch (message.type) {
      case 'node_status_update':
        const nodeData = message.payload as NodeStatusUpdate
        updateNodeStatus(nodeData.nodeId, {
          status: nodeData.status,
          health: nodeData.health,
          lastSeen: nodeData.timestamp,
        })
        break
        
      case 'model_sync_update':
        const modelData = message.payload as ModelSyncUpdate
        updateModelSync(modelData.modelName, {
          status: modelData.status,
          progress: modelData.progress,
          nodesTotal: 1,
          nodesSynced: modelData.status === 'completed' ? 1 : 0,
          error: modelData.error,
        })
        break
        
      case 'cluster_metrics':
        updateMetrics(message.payload as ClusterMetricsUpdate)
        break
        
      case 'performance_alert':
      case 'security_alert':
        addAlert(message.payload)
        break
        
      case 'system_notification':
      case 'user_notification':
        const notification = message.payload
        addNotification({
          type: 'info',
          title: notification.title || 'Notification',
          message: notification.message,
          duration: 10000,
        })
        break
        
      default:
        if (debug) {
          console.log('[WebSocket] Unhandled message type:', message.type)
        }
    }
  }, [debug, updateNodeStatus, updateModelSync, updateMetrics, addAlert, addNotification])

  // Auto-connect when authenticated
  useEffect(() => {
    if (autoConnect && auth.isAuthenticated && connectionState === 'disconnected') {
      connect().catch(console.error)
    } else if (!auth.isAuthenticated && connectionState !== 'disconnected') {
      disconnect()
    }
  }, [auth.isAuthenticated, autoConnect, connectionState, connect, disconnect])

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (reconnectTimer.current) {
        clearTimeout(reconnectTimer.current)
      }
      if (heartbeatTimer.current) {
        clearInterval(heartbeatTimer.current)
      }
      disconnect()
    }
  }, [])

  return {
    // Connection state
    connectionState,
    isConnected: connectionState === 'connected',
    isConnecting: connectionState === 'connecting',
    error: wsClient.getState().error,
    lastMessage,
    
    // Connection management
    connect,
    disconnect,
    reconnect,
    
    // Subscription management
    subscribe,
    unsubscribe,
    subscribeToTopic,
    unsubscribeFromTopic,
    
    // Message handling
    send,
    on,
    off,
    
    // State information
    getSubscriptions,
    getConnectionInfo,
  }
}

export default useWebSocket