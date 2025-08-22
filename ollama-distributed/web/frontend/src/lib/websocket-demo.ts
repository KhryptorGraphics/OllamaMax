/**
 * WebSocket demo and testing utilities
 * Provides mock data generation and connection testing for development
 */

import { WSMessage, MessageTypes, ClusterUpdateData, NodeUpdateData, MetricsUpdateData, NotificationData } from '../types/websocket'

export class WebSocketDemo {
  private mockDataInterval: NodeJS.Timeout | null = null
  private isRunning = false

  constructor(private onMessage: (message: WSMessage) => void) {}

  start(): void {
    if (this.isRunning) return
    
    this.isRunning = true
    console.log('[WebSocket Demo] Starting mock data generation')

    // Send initial cluster data
    this.sendMockClusterUpdate()
    
    // Start periodic updates
    this.mockDataInterval = setInterval(() => {
      const updateType = Math.random()
      
      if (updateType < 0.4) {
        this.sendMockClusterUpdate()
      } else if (updateType < 0.6) {
        this.sendMockNodeUpdate()
      } else if (updateType < 0.8) {
        this.sendMockMetricsUpdate()
      } else {
        this.sendMockNotification()
      }
    }, 3000)
  }

  stop(): void {
    if (!this.isRunning) return
    
    this.isRunning = false
    
    if (this.mockDataInterval) {
      clearInterval(this.mockDataInterval)
      this.mockDataInterval = null
    }
    
    console.log('[WebSocket Demo] Stopped mock data generation')
  }

  private sendMockClusterUpdate(): void {
    const nodeCount = Math.floor(Math.random() * 5) + 1
    const activeModels = Math.floor(Math.random() * 8) + 1
    const isHealthy = Math.random() > 0.1 // 90% chance of being healthy
    const isLeader = Math.random() > 0.5

    const data: ClusterUpdateData = {
      dashboard: {
        clusterStatus: {
          healthy: isHealthy,
          size: nodeCount,
          leader: isLeader,
          consensus: true
        },
        nodeCount,
        activeModels,
        timestamp: new Date().toISOString()
      },
      timestamp: new Date().toISOString()
    }

    this.onMessage({
      type: MessageTypes.CLUSTER_UPDATE,
      data,
      timestamp: new Date().toISOString()
    })
  }

  private sendMockNodeUpdate(): void {
    const nodeId = `node-${Math.floor(Math.random() * 5) + 1}`
    const statuses = ['healthy', 'busy', 'offline', 'starting']
    const status = statuses[Math.floor(Math.random() * statuses.length)]

    const data: NodeUpdateData = {
      node_id: nodeId,
      status,
      timestamp: new Date().toISOString()
    }

    this.onMessage({
      type: MessageTypes.NODE_UPDATE,
      data,
      timestamp: new Date().toISOString()
    })
  }

  private sendMockMetricsUpdate(): void {
    const data: MetricsUpdateData = {
      metrics: {
        cpu_usage: Math.random() * 100,
        memory_usage: Math.random() * 100,
        network_throughput: Math.random() * 1000,
        active_connections: Math.floor(Math.random() * 50),
        request_rate: Math.random() * 100,
        response_time: Math.random() * 500
      },
      timestamp: new Date().toISOString()
    }

    this.onMessage({
      type: MessageTypes.METRICS_UPDATE,
      data,
      timestamp: new Date().toISOString()
    })
  }

  private sendMockNotification(): void {
    const notifications = [
      { type: 'info' as const, title: 'System Info', message: 'Cluster is operating normally' },
      { type: 'success' as const, title: 'Model Loaded', message: 'New model successfully loaded on node-2' },
      { type: 'warning' as const, title: 'High CPU Usage', message: 'Node-3 experiencing high CPU usage (85%)' },
      { type: 'error' as const, title: 'Connection Failed', message: 'Failed to connect to node-4' }
    ]

    const notification = notifications[Math.floor(Math.random() * notifications.length)]

    const data: NotificationData = {
      notification: {
        id: `notif-${Date.now()}`,
        ...notification,
        timestamp: new Date().toISOString()
      },
      timestamp: new Date().toISOString()
    }

    this.onMessage({
      type: MessageTypes.NOTIFICATION,
      data,
      timestamp: new Date().toISOString()
    })
  }
}

/**
 * Create a mock WebSocket server for testing
 */
export function createMockWebSocket(url: string): WebSocket {
  let readyState = WebSocket.CONNECTING
  
  const mockWS = {
    get readyState() { return readyState },
    url,
    send: () => {},
    close: () => {},
    onopen: null as ((event: Event) => void) | null,
    onmessage: null as ((event: MessageEvent) => void) | null,
    onclose: null as ((event: CloseEvent) => void) | null,
    onerror: null as ((event: Event) => void) | null,
    addEventListener: () => {},
    removeEventListener: () => {},
    dispatchEvent: () => true,
    CONNECTING: WebSocket.CONNECTING,
    OPEN: WebSocket.OPEN,
    CLOSING: WebSocket.CLOSING,
    CLOSED: WebSocket.CLOSED,
    binaryType: 'blob' as BinaryType,
    bufferedAmount: 0,
    extensions: '',
    protocol: ''
  }

  // Simulate connection opening
  setTimeout(() => {
    readyState = WebSocket.OPEN
    if (mockWS.onopen) {
      mockWS.onopen(new Event('open'))
    }
  }, 100)

  return mockWS as unknown as WebSocket
}

/**
 * Test WebSocket connection with real backend
 */
export function testWebSocketConnection(url: string): Promise<boolean> {
  return new Promise((resolve) => {
    const ws = new WebSocket(url)
    let resolved = false

    const timeout = setTimeout(() => {
      if (!resolved) {
        resolved = true
        ws.close()
        resolve(false)
      }
    }, 5000)

    ws.onopen = () => {
      if (!resolved) {
        resolved = true
        clearTimeout(timeout)
        ws.close()
        resolve(true)
      }
    }

    ws.onerror = () => {
      if (!resolved) {
        resolved = true
        clearTimeout(timeout)
        resolve(false)
      }
    }

    ws.onclose = () => {
      if (!resolved) {
        resolved = true
        clearTimeout(timeout)
        resolve(false)
      }
    }
  })
}

/**
 * Get WebSocket URL based on current environment
 */
export function getWebSocketURL(): string {
  if (process.env.NODE_ENV === 'development') {
    // Development - try localhost first, fallback to current host
    const devHost = process.env.VITE_WS_HOST || 'localhost:8080'
    return `ws://${devHost}/ws`
  }
  
  // Production - use current host with appropriate protocol
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${protocol}//${window.location.host}/ws`
}

/**
 * Enable demo mode for development
 */
export function enableDemoMode(): void {
  if (typeof window !== 'undefined') {
    (window as any).WEBSOCKET_DEMO_MODE = true
    console.log('[WebSocket Demo] Demo mode enabled - using mock data')
  }
}

/**
 * Check if demo mode is enabled
 */
export function isDemoMode(): boolean {
  return typeof window !== 'undefined' && (window as any).WEBSOCKET_DEMO_MODE === true
}