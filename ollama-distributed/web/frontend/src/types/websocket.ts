// WebSocket types for real-time communication

export interface WebSocketState {
  connected: boolean
  connecting: boolean
  error: string | null
  lastMessage: WebSocketMessage | null
  subscriptions: Set<string>
  reconnectAttempts: number
  maxReconnectAttempts: number
  reconnectInterval: number
}

export type ConnectionState = 'connecting' | 'connected' | 'disconnected' | 'error'

export interface WebSocketMessage<T = any> {
  type: WebSocketMessageType
  payload: T
  timestamp: number
  id?: string
  correlationId?: string
}

export type WebSocketMessageType =
  | 'welcome'
  | 'pong'
  | 'error'
  | 'subscription_confirmed'
  | 'subscription_error'
  | 'node_status_update'
  | 'model_sync_update'
  | 'task_update'
  | 'transfer_progress'
  | 'cluster_metrics'
  | 'performance_alert'
  | 'security_alert'
  | 'system_notification'
  | 'user_notification'

// Subscription messages
export interface SubscriptionMessage {
  action: 'subscribe' | 'unsubscribe'
  channels: string[]
  filters?: Record<string, any>
}

export interface SubscriptionResponse {
  channel: string
  status: 'subscribed' | 'unsubscribed' | 'error'
  error?: string
}

// Real-time data updates
export interface NodeStatusUpdate {
  nodeId: string
  status: 'online' | 'offline' | 'draining' | 'maintenance'
  health: {
    cpu: number
    memory: number
    disk: number
    network: number
  }
  timestamp: string
}

export interface ModelSyncUpdate {
  modelName: string
  nodeId: string
  status: 'syncing' | 'completed' | 'failed'
  progress: number
  bytesTransferred: number
  totalBytes: number
  error?: string
  timestamp: string
}

export interface TaskUpdate {
  taskId: string
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled'
  progress?: number
  result?: any
  error?: string
  metrics?: {
    duration?: number
    tokensProcessed?: number
    memoryUsed?: number
  }
  timestamp: string
}

export interface TransferProgressUpdate {
  transferId: string
  status: 'pending' | 'active' | 'completed' | 'failed' | 'cancelled'
  progress: {
    bytesTransferred: number
    totalBytes: number
    percentage: number
    rate: number
    eta?: number
  }
  timestamp: string
}

export interface ClusterMetricsUpdate {
  timestamp: string
  nodes: {
    [nodeId: string]: {
      cpu: number
      memory: number
      disk: number
      tasks: number
      status: string
    }
  }
  cluster: {
    totalNodes: number
    healthyNodes: number
    totalTasks: number
    activeTasks: number
    avgResponseTime: number
  }
}

export interface PerformanceAlert {
  id: string
  type: 'cpu_high' | 'memory_high' | 'disk_full' | 'slow_response' | 'error_rate_high'
  severity: 'low' | 'medium' | 'high' | 'critical'
  message: string
  nodeId?: string
  modelName?: string
  threshold: number
  currentValue: number
  timestamp: string
  duration?: number
}

export interface SecurityAlert {
  id: string
  type: 'unauthorized_access' | 'brute_force' | 'suspicious_activity' | 'policy_violation'
  severity: 'low' | 'medium' | 'high' | 'critical'
  message: string
  userId?: string
  ipAddress?: string
  location?: string
  timestamp: string
  details: Record<string, any>
}

export interface SystemNotification {
  id: string
  type: 'maintenance' | 'update' | 'outage' | 'feature' | 'deprecation'
  title: string
  message: string
  severity: 'info' | 'warning' | 'error'
  scheduledFor?: string
  duration?: number
  affectedServices?: string[]
  timestamp: string
}

export interface UserNotification {
  id: string
  userId: string
  type: 'task_completed' | 'task_failed' | 'model_synced' | 'quota_exceeded' | 'permission_changed'
  title: string
  message: string
  read: boolean
  actionUrl?: string
  actionLabel?: string
  timestamp: string
  expiresAt?: string
}

// WebSocket configuration
export interface WebSocketConfig {
  url: string
  protocols?: string[]
  reconnect: boolean
  maxReconnectAttempts: number
  reconnectInterval: number
  heartbeatInterval: number
  timeout: number
  debug: boolean
}

// WebSocket client interface
export interface WebSocketClient {
  connect(): Promise<void>
  disconnect(): void
  subscribe(channels: string[], filters?: Record<string, any>): void
  unsubscribe(channels: string[]): void
  send(message: WebSocketMessage): void
  on(event: string, handler: (data: any) => void): void
  off(event: string, handler?: (data: any) => void): void
  getState(): WebSocketState
}

// Event handlers
export type WebSocketEventHandler<T = any> = (data: T) => void

export interface WebSocketEventHandlers {
  onConnect?: () => void
  onDisconnect?: (reason: string) => void
  onError?: (error: Error) => void
  onMessage?: WebSocketEventHandler<WebSocketMessage>
  onNodeStatusUpdate?: WebSocketEventHandler<NodeStatusUpdate>
  onModelSyncUpdate?: WebSocketEventHandler<ModelSyncUpdate>
  onTaskUpdate?: WebSocketEventHandler<TaskUpdate>
  onTransferProgress?: WebSocketEventHandler<TransferProgressUpdate>
  onClusterMetrics?: WebSocketEventHandler<ClusterMetricsUpdate>
  onPerformanceAlert?: WebSocketEventHandler<PerformanceAlert>
  onSecurityAlert?: WebSocketEventHandler<SecurityAlert>
  onSystemNotification?: WebSocketEventHandler<SystemNotification>
  onUserNotification?: WebSocketEventHandler<UserNotification>
}