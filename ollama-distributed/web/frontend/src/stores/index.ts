// Central store for OllamaMax distributed system
import { create } from 'zustand'
import { subscribeWithSelector } from 'zustand/middleware'
import { immer } from 'zustand/middleware/immer'
import { devtools } from 'zustand/middleware'
import { persist } from 'zustand/middleware'
import type {
  GlobalState,
  AuthState,
  UIState,
  WebSocketState,
  PerformanceState,
  ClusterState,
  NodeState,
  ModelState,
  ModelInfo,
  ModelSyncStatus,
  PerformanceMetrics,
  PerformanceAlert,
  Notification,
  ApiResponse,
  PaginatedResponse,
} from '@/types'
import { authService } from '@/services/auth/authService'
import { wsClient } from '@/services/websocket/client'
import { apiClient } from '@/services/api/client'
import { debounce } from '@/utils/debounce'
import { retryWithBackoff } from '@/utils/retry'

// Auth store slice
interface AuthSlice {
  auth: AuthState
  login: (credentials: any) => Promise<void>
  logout: () => Promise<void>
  refreshToken: () => Promise<void>
  updateProfile: (updates: any) => Promise<void>
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
}

// Models store slice - Enhanced for Sprint C
interface ModelsSlice {
  models: ModelState
  fetchModels: () => Promise<void>
  fetchModelDetails: (modelName: string) => Promise<ModelInfo | null>
  deployModel: (modelName: string, nodes?: string[]) => Promise<void>
  undeployModel: (modelName: string, nodes?: string[]) => Promise<void>
  syncModel: (modelName: string, targetNodes?: string[]) => Promise<void>
  deleteModel: (modelName: string) => Promise<void>
  uploadModel: (file: File, metadata: any) => Promise<void>
  updateModelSync: (modelName: string, status: ModelSyncStatus) => void
  clearModelsError: () => void
  setModelsLoading: (loading: boolean) => void
  // Real-time optimization
  subscribeToModelUpdates: () => void
  unsubscribeFromModelUpdates: () => void
  // Caching and performance
  invalidateModelsCache: () => void
  getModelFromCache: (modelName: string) => ModelInfo | null
}

// Nodes store slice - Enhanced for Sprint C
interface NodesSlice {
  nodes: NodeState[]
  nodesLoading: boolean
  nodesError: string | null
  selectedNode: NodeState | null
  nodeMetrics: Record<string, any>
  fetchNodes: () => Promise<void>
  fetchNodeDetails: (nodeId: string) => Promise<NodeState | null>
  fetchNodeMetrics: (nodeId: string) => Promise<void>
  updateNodeStatus: (nodeId: string, updates: Partial<NodeState>) => void
  drainNode: (nodeId: string) => Promise<void>
  enableNode: (nodeId: string) => Promise<void>
  removeNode: (nodeId: string) => Promise<void>
  setSelectedNode: (node: NodeState | null) => void
  clearNodesError: () => void
  // Real-time optimization
  subscribeToNodeUpdates: () => void
  unsubscribeFromNodeUpdates: () => void
  // Performance optimization
  getNodeFromCache: (nodeId: string) => NodeState | null
  invalidateNodesCache: () => void
}

// Monitoring store slice - Enhanced for Sprint C
interface MonitoringSlice {
  monitoring: {
    metrics: PerformanceMetrics | null
    alerts: PerformanceAlert[]
    loading: boolean
    error: string | null
    lastUpdated: number | null
    autoRefresh: boolean
    refreshInterval: number
  }
  fetchMetrics: () => Promise<void>
  fetchAlerts: () => Promise<void>
  acknowledgeAlert: (alertId: string) => Promise<void>
  resolveAlert: (alertId: string) => Promise<void>
  createAlert: (alert: Omit<PerformanceAlert, 'id' | 'timestamp'>) => Promise<void>
  updateMetrics: (metrics: PerformanceMetrics) => void
  addAlert: (alert: PerformanceAlert) => void
  removeAlert: (alertId: string) => void
  setAutoRefresh: (enabled: boolean) => void
  setRefreshInterval: (interval: number) => void
  clearMonitoringError: () => void
  // Real-time optimization
  subscribeToMetricsUpdates: () => void
  unsubscribeFromMetricsUpdates: () => void
  startAutoRefresh: () => void
  stopAutoRefresh: () => void
  // Performance optimization
  getMetricsHistory: (duration: string) => PerformanceMetrics[]
  invalidateMetricsCache: () => void
}

// Dashboard store slice - Enhanced for Sprint C
interface DashboardSlice {
  dashboard: {
    summary: {
      totalNodes: number
      healthyNodes: number
      totalModels: number
      activeTasks: number
      systemHealth: 'healthy' | 'warning' | 'critical'
    }
    recentActivity: any[]
    quickActions: any[]
    widgets: any[]
    layout: any
    loading: boolean
    error: string | null
    lastUpdated: number | null
  }
  fetchDashboardData: () => Promise<void>
  fetchRecentActivity: () => Promise<void>
  updateDashboardSummary: (summary: any) => void
  addRecentActivity: (activity: any) => void
  updateWidgetLayout: (layout: any) => void
  setDashboardLoading: (loading: boolean) => void
  clearDashboardError: () => void
  // Real-time optimization
  subscribeToDashboardUpdates: () => void
  unsubscribeFromDashboardUpdates: () => void
  // Performance optimization
  refreshDashboardData: () => Promise<void>
  invalidateDashboardCache: () => void
}

// UI store slice
interface UISlice {
  ui: UIState
  toggleTheme: () => void
  setTheme: (theme: 'light' | 'dark' | 'auto') => void
  toggleSidebar: () => void
  setSidebarOpen: (open: boolean) => void
  setLoading: (loading: boolean) => void
  addNotification: (notification: Omit<Notification, 'id'>) => void
  removeNotification: (id: string) => void
  clearNotifications: () => void
  openModal: (component: string, props?: any) => void
  closeModal: () => void
}

// WebSocket store slice
interface WebSocketSlice {
  websocket: WebSocketState
  connect: () => Promise<void>
  disconnect: () => void
  subscribe: (channels: string[]) => void
  unsubscribe: (channels: string[]) => void
  updateConnectionState: (updates: Partial<WebSocketState>) => void
}

// Cluster store slice  
interface ClusterSlice {
  cluster: ClusterState
  nodes: any[]
  models: any[]
  tasks: any[]
  transfers: any[]
  metrics: any
  fetchNodes: () => Promise<void>
  fetchModels: () => Promise<void>
  fetchTasks: () => Promise<void>
  fetchTransfers: () => Promise<void>
  fetchMetrics: () => Promise<void>
  updateNodeStatus: (nodeId: string, status: any) => void
  updateModelSync: (modelName: string, status: any) => void
  updateTaskStatus: (taskId: string, status: any) => void
  updateTransferProgress: (transferId: string, progress: any) => void
}

// Performance store slice
interface PerformanceSlice {
  performance: PerformanceState
  updateMetrics: (metrics: any) => void
  addAlert: (alert: any) => void
  clearAlert: (id: string) => void
}

// Combined store type
type Store = AuthSlice & UISlice & WebSocketSlice & ClusterSlice & PerformanceSlice & ModelsSlice & NodesSlice & MonitoringSlice & DashboardSlice

// Request deduplication and caching
const requestCache = new Map<string, { data: any; timestamp: number; ttl: number }>()
const pendingRequests = new Map<string, Promise<any>>()
const CACHE_TTL = {
  models: 30000,    // 30 seconds
  nodes: 15000,     // 15 seconds 
  metrics: 10000,   // 10 seconds
  dashboard: 20000, // 20 seconds
}

// Debounced update functions
const debouncedUpdateMetrics = debounce((updateFn: (metrics: any) => void, metrics: any) => {
  updateFn(metrics)
}, 1000)

const debouncedUpdateModels = debounce((updateFn: (models: any) => void, models: any) => {
  updateFn(models)
}, 500)

// Cache utilities
function getCachedData<T>(key: string): T | null {
  const cached = requestCache.get(key)
  if (cached && Date.now() - cached.timestamp < cached.ttl) {
    return cached.data
  }
  requestCache.delete(key)
  return null
}

function setCachedData<T>(key: string, data: T, ttl: number): void {
  requestCache.set(key, { data, timestamp: Date.now(), ttl })
}

// Request deduplication utility
function dedupedRequest<T>(key: string, requestFn: () => Promise<T>): Promise<T> {
  if (pendingRequests.has(key)) {
    return pendingRequests.get(key) as Promise<T>
  }
  
  const promise = requestFn().finally(() => {
    pendingRequests.delete(key)
  })
  
  pendingRequests.set(key, promise)
  return promise
}

// Optimistic update utility
function optimisticUpdate<T>(
  updateFn: (data: T) => void,
  optimisticData: T,
  asyncAction: () => Promise<T>,
  revertFn?: (error: any) => void
): Promise<T> {
  // Apply optimistic update immediately
  updateFn(optimisticData)
  
  // Execute async action
  return asyncAction().catch((error) => {
    // Revert on error
    if (revertFn) {
      revertFn(error)
    }
    throw error
  })
}

// Create the store with middleware
export const useStore = create<Store>()(
  devtools(
    subscribeWithSelector(
      immer((set, get) => ({
        // Auth slice
        auth: {
          isAuthenticated: false,
          user: null,
          token: null,
          refreshToken: null,
          expiresAt: null,
          permissions: [],
          loading: false,
          error: null,
        },

        login: async (credentials) => {
          set((state) => {
            state.auth.loading = true
            state.auth.error = null
          })

          try {
            const user = await authService.login(credentials)
            const authState = authService.getCurrentState()
            
            set((state) => {
              state.auth = authState
            })
          } catch (error) {
            set((state) => {
              state.auth.loading = false
              state.auth.error = (error as Error).message
            })
            throw error
          }
        },

        logout: async () => {
          await authService.logout()
          set((state) => {
            state.auth = {
              isAuthenticated: false,
              user: null,
              token: null,
              refreshToken: null,
              expiresAt: null,
              permissions: [],
              loading: false,
              error: null,
            }
          })
        },

        refreshToken: async () => {
          try {
            await authService.refreshToken()
            const authState = authService.getCurrentState()
            set((state) => {
              state.auth = authState
            })
          } catch (error) {
            // Auto logout on refresh failure
            await get().logout()
            throw error
          }
        },

        updateProfile: async (updates) => {
          try {
            const user = await authService.updateProfile(updates)
            set((state) => {
              state.auth.user = user
            })
          } catch (error) {
            set((state) => {
              state.auth.error = (error as Error).message
            })
            throw error
          }
        },

        setLoading: (loading) => {
          set((state) => {
            state.auth.loading = loading
          })
        },

        setError: (error) => {
          set((state) => {
            state.auth.error = error
          })
        },

        // UI slice
        ui: {
          theme: (localStorage.getItem('theme') as 'light' | 'dark' | 'auto') || 'auto',
          sidebarOpen: window.innerWidth >= 1024, // Open on desktop by default
          loading: false,
          notifications: [],
          modal: {
            isOpen: false,
            component: undefined,
            props: undefined,
          },
        },

        // Notification counters for navigation badges
        notifications: {
          alerts: [],
          pendingTasks: 0,
          securityAlerts: 0,
        },

        toggleTheme: () => {
          set((state) => {
            const themes: Array<'light' | 'dark' | 'auto'> = ['light', 'dark', 'auto']
            const currentIndex = themes.indexOf(state.ui.theme)
            const nextTheme = themes[(currentIndex + 1) % themes.length]
            state.ui.theme = nextTheme
            localStorage.setItem('theme', nextTheme)
          })
        },

        setTheme: (theme) => {
          set((state) => {
            state.ui.theme = theme
            localStorage.setItem('theme', theme)
          })
        },

        toggleSidebar: () => {
          set((state) => {
            state.ui.sidebarOpen = !state.ui.sidebarOpen
          })
        },

        setSidebarOpen: (open) => {
          set((state) => {
            state.ui.sidebarOpen = open
          })
        },

        addNotification: (notification) => {
          set((state) => {
            const id = Math.random().toString(36).substr(2, 9)
            state.ui.notifications.push({
              ...notification,
              id,
              timestamp: Date.now(),
            })
          })
        },

        removeNotification: (id) => {
          set((state) => {
            state.ui.notifications = state.ui.notifications.filter(n => n.id !== id)
          })
        },

        clearNotifications: () => {
          set((state) => {
            state.ui.notifications = []
          })
        },

        openModal: (component, props) => {
          set((state) => {
            state.ui.modal = {
              isOpen: true,
              component,
              props,
            }
          })
        },

        closeModal: () => {
          set((state) => {
            state.ui.modal = {
              isOpen: false,
              component: undefined,
              props: undefined,
            }
          })
        },

        // WebSocket slice
        websocket: {
          connected: false,
          connecting: false,
          error: null,
          lastMessage: null,
          subscriptions: new Set(),
          reconnectAttempts: 0,
          maxReconnectAttempts: 5,
          reconnectInterval: 5000,
        },

        connect: async () => {
          set((state) => {
            state.websocket.connecting = true
            state.websocket.error = null
          })

          try {
            await wsClient.connect()
            set((state) => {
              state.websocket.connected = true
              state.websocket.connecting = false
            })
          } catch (error) {
            set((state) => {
              state.websocket.connecting = false
              state.websocket.error = (error as Error).message
            })
            throw error
          }
        },

        disconnect: () => {
          wsClient.disconnect()
          set((state) => {
            state.websocket.connected = false
            state.websocket.connecting = false
            state.websocket.subscriptions = new Set()
          })
        },

        subscribe: (channels) => {
          wsClient.subscribe(channels)
          set((state) => {
            channels.forEach(channel => {
              state.websocket.subscriptions.add(channel)
            })
          })
        },

        unsubscribe: (channels) => {
          wsClient.unsubscribe(channels)
          set((state) => {
            channels.forEach(channel => {
              state.websocket.subscriptions.delete(channel)
            })
          })
        },

        updateConnectionState: (updates) => {
          set((state) => {
            Object.assign(state.websocket, updates)
          })
        },

        // Cluster slice
        cluster: {
          status: 'unknown',
          leader: null,
          nodes: 0,
          healthyNodes: 0,
          loading: false,
          error: null,
          lastUpdated: null,
        },

        // Initial state will be populated by slice implementations below

        // Models slice implementation
        models: {
          models: [],
          loading: false,
          error: null,
          syncStatus: {},
        },

        fetchModels: async () => {
          const cacheKey = 'models:list'
          const cached = getCachedData<ModelInfo[]>(cacheKey)
          if (cached) {
            set((state) => {
              state.models.models = cached
            })
            return
          }

          return dedupedRequest(cacheKey, async () => {
            set((state) => {
              state.models.loading = true
              state.models.error = null
            })

            try {
              const response = await retryWithBackoff(() => apiClient.getModels())
              const models = response.data.data || []
              
              setCachedData(cacheKey, models, CACHE_TTL.models)
              
              set((state) => {
                state.models.models = models
                state.models.loading = false
              })
            } catch (error) {
              set((state) => {
                state.models.loading = false
                state.models.error = (error as Error).message
              })
              throw error
            }
          })
        },

        fetchModelDetails: async (modelName: string) => {
          const cacheKey = `models:details:${modelName}`
          const cached = getCachedData<ModelInfo>(cacheKey)
          if (cached) return cached

          try {
            const response = await apiClient.getModel(modelName)
            const model = response.data
            setCachedData(cacheKey, model, CACHE_TTL.models)
            return model
          } catch (error) {
            console.error('Failed to fetch model details:', error)
            return null
          }
        },

        deployModel: async (modelName: string, nodes?: string[]) => {
          return optimisticUpdate(
            (optimisticData) => {
              set((state) => {
                const model = state.models.models.find(m => m.name === modelName)
                if (model) {
                  model.syncStatus = optimisticData
                }
              })
            },
            { status: 'syncing', progress: 0, nodesTotal: nodes?.length || 0, nodesSynced: 0 } as ModelSyncStatus,
            async () => {
              const response = await apiClient.deployModel(modelName, { nodes })
              get().addNotification({
                type: 'success',
                title: 'Model Deployment',
                message: `Started deploying ${modelName}`,
              })
              return response.data
            },
            (error) => {
              set((state) => {
                const model = state.models.models.find(m => m.name === modelName)
                if (model) {
                  model.syncStatus = { status: 'failed', progress: 0, nodesTotal: 0, nodesSynced: 0, error: error.message }
                }
              })
              get().addNotification({
                type: 'error',
                title: 'Deployment Failed',
                message: `Failed to deploy ${modelName}: ${error.message}`,
              })
            }
          )
        },

        undeployModel: async (modelName: string, nodes?: string[]) => {
          try {
            await apiClient.undeployModel(modelName, { nodes })
            get().addNotification({
              type: 'success',
              title: 'Model Undeployment',
              message: `Started undeploying ${modelName}`,
            })
          } catch (error) {
            get().addNotification({
              type: 'error',
              title: 'Undeployment Failed',
              message: `Failed to undeploy ${modelName}: ${(error as Error).message}`,
            })
            throw error
          }
        },

        syncModel: async (modelName: string, targetNodes?: string[]) => {
          try {
            await apiClient.syncModel(modelName, targetNodes)
            get().addNotification({
              type: 'info',
              title: 'Model Sync',
              message: `Started syncing ${modelName}`,
            })
          } catch (error) {
            get().addNotification({
              type: 'error',
              title: 'Sync Failed',
              message: `Failed to sync ${modelName}: ${(error as Error).message}`,
            })
            throw error
          }
        },

        deleteModel: async (modelName: string) => {
          try {
            await apiClient.deleteModel(modelName)
            set((state) => {
              state.models.models = state.models.models.filter(m => m.name !== modelName)
            })
            get().invalidateModelsCache()
            get().addNotification({
              type: 'success',
              title: 'Model Deleted',
              message: `Successfully deleted ${modelName}`,
            })
          } catch (error) {
            get().addNotification({
              type: 'error',
              title: 'Delete Failed',
              message: `Failed to delete ${modelName}: ${(error as Error).message}`,
            })
            throw error
          }
        },

        uploadModel: async (file: File, metadata: any) => {
          const formData = new FormData()
          formData.append('file', file)
          formData.append('metadata', JSON.stringify(metadata))

          try {
            set((state) => {
              state.models.loading = true
            })

            await apiClient.uploadModel(formData)

            await get().fetchModels()
            get().addNotification({
              type: 'success',
              title: 'Model Uploaded',
              message: `Successfully uploaded ${file.name}`,
            })
          } catch (error) {
            set((state) => {
              state.models.loading = false
              state.models.error = (error as Error).message
            })
            get().addNotification({
              type: 'error',
              title: 'Upload Failed',
              message: `Failed to upload ${file.name}: ${(error as Error).message}`,
            })
            throw error
          }
        },

        updateModelSync: (modelName: string, status: ModelSyncStatus) => {
          debouncedUpdateModels((status) => {
            set((state) => {
              const model = state.models.models.find(m => m.name === modelName)
              if (model) {
                model.syncStatus = status
              }
              state.models.syncStatus[modelName] = status
            })
          }, status)
        },

        clearModelsError: () => {
          set((state) => {
            state.models.error = null
          })
        },

        setModelsLoading: (loading: boolean) => {
          set((state) => {
            state.models.loading = loading
          })
        },

        subscribeToModelUpdates: () => {
          get().subscribe(['models:*', 'model_sync:*'])
        },

        unsubscribeFromModelUpdates: () => {
          get().unsubscribe(['models:*', 'model_sync:*'])
        },

        invalidateModelsCache: () => {
          Array.from(requestCache.keys())
            .filter(key => key.startsWith('models:'))
            .forEach(key => requestCache.delete(key))
        },

        getModelFromCache: (modelName: string) => {
          return getCachedData<ModelInfo>(`models:details:${modelName}`)
        },

        // Nodes slice implementation
        nodes: [],
        nodesLoading: false,
        nodesError: null,
        selectedNode: null,
        nodeMetrics: {},

        fetchNodes: async () => {
          const cacheKey = 'nodes:list'
          const cached = getCachedData<NodeState[]>(cacheKey)
          if (cached) {
            set((state) => {
              state.nodes = cached
            })
            return
          }

          return dedupedRequest(cacheKey, async () => {
            set((state) => {
              state.nodesLoading = true
              state.nodesError = null
            })

            try {
              const response = await retryWithBackoff(() => apiClient.getNodes())
              const nodes = response.data || []
              
              setCachedData(cacheKey, nodes, CACHE_TTL.nodes)
              
              set((state) => {
                state.nodes = nodes
                state.nodesLoading = false
                state.cluster.nodes = nodes.length
                state.cluster.healthyNodes = nodes.filter(n => n.status === 'online').length
              })
            } catch (error) {
              set((state) => {
                state.nodesLoading = false
                state.nodesError = (error as Error).message
              })
              throw error
            }
          })
        },

        fetchNodeDetails: async (nodeId: string) => {
          const cacheKey = `nodes:details:${nodeId}`
          const cached = getCachedData<NodeState>(cacheKey)
          if (cached) return cached

          try {
            const response = await apiClient.getNode(nodeId)
            const node = response.data
            setCachedData(cacheKey, node, CACHE_TTL.nodes)
            return node
          } catch (error) {
            console.error('Failed to fetch node details:', error)
            return null
          }
        },

        fetchNodeMetrics: async (nodeId: string) => {
          try {
            const response = await apiClient.getNodeMetrics(nodeId)
            set((state) => {
              state.nodeMetrics[nodeId] = response.data
            })
          } catch (error) {
            console.error('Failed to fetch node metrics:', error)
          }
        },

        updateNodeStatus: (nodeId: string, updates: Partial<NodeState>) => {
          set((state) => {
            const nodeIndex = state.nodes.findIndex(n => n.id === nodeId)
            if (nodeIndex !== -1) {
              Object.assign(state.nodes[nodeIndex], updates)
            }
            
            // Update cluster state
            state.cluster.healthyNodes = state.nodes.filter(n => n.status === 'online').length
          })
        },

        drainNode: async (nodeId: string) => {
          return optimisticUpdate(
            (optimisticData) => {
              set((state) => {
                const node = state.nodes.find(n => n.id === nodeId)
                if (node) {
                  node.status = optimisticData.status
                }
              })
            },
            { status: 'draining' } as Partial<NodeState>,
            async () => {
              const response = await apiClient.drainNode(nodeId)
              get().addNotification({
                type: 'info',
                title: 'Node Draining',
                message: `Started draining node ${nodeId}`,
              })
              return response.data
            },
            (error) => {
              // Revert optimistic update
              get().fetchNodes()
              get().addNotification({
                type: 'error',
                title: 'Drain Failed',
                message: `Failed to drain node ${nodeId}: ${error.message}`,
              })
            }
          )
        },

        enableNode: async (nodeId: string) => {
          try {
            await apiClient.enableNode(nodeId)
            get().fetchNodes()
            get().addNotification({
              type: 'success',
              title: 'Node Enabled',
              message: `Successfully enabled node ${nodeId}`,
            })
          } catch (error) {
            get().addNotification({
              type: 'error',
              title: 'Enable Failed',
              message: `Failed to enable node ${nodeId}: ${(error as Error).message}`,
            })
            throw error
          }
        },

        removeNode: async (nodeId: string) => {
          try {
            await apiClient.removeNode(nodeId)
            set((state) => {
              state.nodes = state.nodes.filter(n => n.id !== nodeId)
              state.cluster.nodes = state.nodes.length
              state.cluster.healthyNodes = state.nodes.filter(n => n.status === 'online').length
            })
            get().addNotification({
              type: 'success',
              title: 'Node Removed',
              message: `Successfully removed node ${nodeId}`,
            })
          } catch (error) {
            get().addNotification({
              type: 'error',
              title: 'Remove Failed',
              message: `Failed to remove node ${nodeId}: ${(error as Error).message}`,
            })
            throw error
          }
        },

        setSelectedNode: (node: NodeState | null) => {
          set((state) => {
            state.selectedNode = node
          })
        },

        clearNodesError: () => {
          set((state) => {
            state.nodesError = null
          })
        },

        subscribeToNodeUpdates: () => {
          get().subscribe(['nodes:*', 'node_status:*'])
        },

        unsubscribeFromNodeUpdates: () => {
          get().unsubscribe(['nodes:*', 'node_status:*'])
        },

        getNodeFromCache: (nodeId: string) => {
          return getCachedData<NodeState>(`nodes:details:${nodeId}`)
        },

        invalidateNodesCache: () => {
          Array.from(requestCache.keys())
            .filter(key => key.startsWith('nodes:'))
            .forEach(key => requestCache.delete(key))
        },

        // Monitoring slice implementation
        monitoring: {
          metrics: null,
          alerts: [],
          loading: false,
          error: null,
          lastUpdated: null,
          autoRefresh: false,
          refreshInterval: 10000, // 10 seconds
        },

        fetchMetrics: async () => {
          const cacheKey = 'metrics:current'
          const cached = getCachedData<PerformanceMetrics>(cacheKey)
          if (cached) {
            set((state) => {
              state.monitoring.metrics = cached
              state.monitoring.lastUpdated = Date.now()
            })
            return
          }

          return dedupedRequest(cacheKey, async () => {
            set((state) => {
              state.monitoring.loading = true
              state.monitoring.error = null
            })

            try {
              const response = await retryWithBackoff(() => 
                apiClient.getClusterMetrics()
              )
              const metrics = response.data
              
              setCachedData(cacheKey, metrics, CACHE_TTL.metrics)
              
              set((state) => {
                state.monitoring.metrics = metrics
                state.monitoring.loading = false
                state.monitoring.lastUpdated = Date.now()
              })
            } catch (error) {
              set((state) => {
                state.monitoring.loading = false
                state.monitoring.error = (error as Error).message
              })
              throw error
            }
          })
        },

        fetchAlerts: async () => {
          try {
            const response = await apiClient.getAlerts()
            set((state) => {
              state.monitoring.alerts = response.data || []
            })
          } catch (error) {
            console.error('Failed to fetch alerts:', error)
          }
        },

        acknowledgeAlert: async (alertId: string) => {
          try {
            await apiClient.acknowledgeAlert(alertId)
            set((state) => {
              const alert = state.monitoring.alerts.find(a => a.id === alertId)
              if (alert) {
                alert.acknowledged = true
              }
            })
          } catch (error) {
            console.error('Failed to acknowledge alert:', error)
            throw error
          }
        },

        resolveAlert: async (alertId: string) => {
          try {
            await apiClient.resolveAlert(alertId)
            set((state) => {
              const alert = state.monitoring.alerts.find(a => a.id === alertId)
              if (alert) {
                alert.resolvedAt = Date.now()
              }
            })
          } catch (error) {
            console.error('Failed to resolve alert:', error)
            throw error
          }
        },

        createAlert: async (alert: Omit<PerformanceAlert, 'id' | 'timestamp'>) => {
          try {
            const response = await apiClient.createAlert(alert)
            set((state) => {
              state.monitoring.alerts.push(response.data)
            })
          } catch (error) {
            console.error('Failed to create alert:', error)
            throw error
          }
        },

        updateMetrics: (metrics: PerformanceMetrics) => {
          debouncedUpdateMetrics((metrics) => {
            set((state) => {
              state.monitoring.metrics = metrics
              state.monitoring.lastUpdated = Date.now()
            })
          }, metrics)
        },

        addAlert: (alert: PerformanceAlert) => {
          set((state) => {
            state.monitoring.alerts.unshift(alert)
            // Keep only latest 100 alerts
            if (state.monitoring.alerts.length > 100) {
              state.monitoring.alerts = state.monitoring.alerts.slice(0, 100)
            }
          })
        },

        removeAlert: (alertId: string) => {
          set((state) => {
            state.monitoring.alerts = state.monitoring.alerts.filter(a => a.id !== alertId)
          })
        },

        setAutoRefresh: (enabled: boolean) => {
          set((state) => {
            state.monitoring.autoRefresh = enabled
          })
          
          if (enabled) {
            get().startAutoRefresh()
          } else {
            get().stopAutoRefresh()
          }
        },

        setRefreshInterval: (interval: number) => {
          set((state) => {
            state.monitoring.refreshInterval = interval
          })
          
          // Restart auto-refresh with new interval
          if (get().monitoring.autoRefresh) {
            get().stopAutoRefresh()
            get().startAutoRefresh()
          }
        },

        clearMonitoringError: () => {
          set((state) => {
            state.monitoring.error = null
          })
        },

        subscribeToMetricsUpdates: () => {
          get().subscribe(['metrics:*', 'alerts:*', 'performance:*'])
        },

        unsubscribeFromMetricsUpdates: () => {
          get().unsubscribe(['metrics:*', 'alerts:*', 'performance:*'])
        },

        startAutoRefresh: (() => {
          let intervalId: NodeJS.Timeout
          
          return () => {
            if (intervalId) clearInterval(intervalId)
            
            const refresh = () => {
              const state = get()
              if (state.monitoring.autoRefresh && state.auth.isAuthenticated) {
                state.fetchMetrics().catch(console.error)
              }
            }
            
            intervalId = setInterval(refresh, get().monitoring.refreshInterval)
          }
        })(),

        stopAutoRefresh: (() => {
          let intervalId: NodeJS.Timeout
          
          return () => {
            if (intervalId) {
              clearInterval(intervalId)
              intervalId = undefined
            }
          }
        })(),

        getMetricsHistory: (duration: string) => {
          // Implementation for retrieving historical metrics
          // This would typically fetch from a time-series database
          return []
        },

        invalidateMetricsCache: () => {
          Array.from(requestCache.keys())
            .filter(key => key.startsWith('metrics:'))
            .forEach(key => requestCache.delete(key))
        },

        // Dashboard slice implementation
        dashboard: {
          summary: {
            totalNodes: 0,
            healthyNodes: 0,
            totalModels: 0,
            activeTasks: 0,
            systemHealth: 'healthy' as const,
          },
          recentActivity: [],
          quickActions: [],
          widgets: [],
          layout: null,
          loading: false,
          error: null,
          lastUpdated: null,
        },

        fetchDashboardData: async () => {
          const cacheKey = 'dashboard:summary'
          const cached = getCachedData<any>(cacheKey)
          if (cached) {
            set((state) => {
              state.dashboard.summary = cached
              state.dashboard.lastUpdated = Date.now()
            })
            return
          }

          return dedupedRequest(cacheKey, async () => {
            set((state) => {
              state.dashboard.loading = true
              state.dashboard.error = null
            })

            try {
              const [summaryRes, activityRes] = await Promise.all([
                apiClient.getDashboardSummary(),
                apiClient.getDashboardActivity(),
              ])
              
              const summary = summaryRes.data
              const activity = activityRes.data || []
              
              setCachedData(cacheKey, summary, CACHE_TTL.dashboard)
              
              set((state) => {
                state.dashboard.summary = summary
                state.dashboard.recentActivity = activity
                state.dashboard.loading = false
                state.dashboard.lastUpdated = Date.now()
              })
            } catch (error) {
              set((state) => {
                state.dashboard.loading = false
                state.dashboard.error = (error as Error).message
              })
              throw error
            }
          })
        },

        fetchRecentActivity: async () => {
          try {
            const response = await apiClient.getDashboardActivity()
            set((state) => {
              state.dashboard.recentActivity = response.data || []
            })
          } catch (error) {
            console.error('Failed to fetch recent activity:', error)
          }
        },

        updateDashboardSummary: (summary: any) => {
          set((state) => {
            Object.assign(state.dashboard.summary, summary)
            state.dashboard.lastUpdated = Date.now()
          })
        },

        addRecentActivity: (activity: any) => {
          set((state) => {
            state.dashboard.recentActivity.unshift(activity)
            // Keep only latest 50 activities
            if (state.dashboard.recentActivity.length > 50) {
              state.dashboard.recentActivity = state.dashboard.recentActivity.slice(0, 50)
            }
          })
        },

        updateWidgetLayout: (layout: any) => {
          set((state) => {
            state.dashboard.layout = layout
          })
        },

        setDashboardLoading: (loading: boolean) => {
          set((state) => {
            state.dashboard.loading = loading
          })
        },

        clearDashboardError: () => {
          set((state) => {
            state.dashboard.error = null
          })
        },

        subscribeToDashboardUpdates: () => {
          get().subscribe(['dashboard:*', 'system:*'])
        },

        unsubscribeFromDashboardUpdates: () => {
          get().unsubscribe(['dashboard:*', 'system:*'])
        },

        refreshDashboardData: async () => {
          get().invalidateDashboardCache()
          await Promise.all([
            get().fetchDashboardData(),
            get().fetchNodes(),
            get().fetchModels(),
          ])
        },

        invalidateDashboardCache: () => {
          Array.from(requestCache.keys())
            .filter(key => key.startsWith('dashboard:'))
            .forEach(key => requestCache.delete(key))
        },

        // Performance slice
        performance: {
          metrics: null,
          alerts: [],
          lastUpdated: null,
          loading: false,
          error: null,
        },

        // Metrics and alerts are handled by monitoring slice above

        clearAlert: (id) => {
          set((state) => {
            state.performance.alerts = state.performance.alerts.filter(a => a.id !== id)
          })
        },
      }))
    ),
    {
      name: 'ollama-store',
      partialize: (state) => ({
        ui: {
          theme: state.ui.theme,
          sidebarOpen: state.ui.sidebarOpen,
        },
        monitoring: {
          autoRefresh: state.monitoring.autoRefresh,
          refreshInterval: state.monitoring.refreshInterval,
        },
        dashboard: {
          layout: state.dashboard.layout,
        },
      }),
    }
  )
)

// Set up WebSocket event handlers
wsClient.setupEventHandlers({
  onConnect: () => {
    useStore.getState().updateConnectionState({ 
      connected: true, 
      connecting: false,
      error: null,
    })
  },
  
  onDisconnect: () => {
    useStore.getState().updateConnectionState({ 
      connected: false,
      connecting: false,
    })
  },
  
  onError: (error) => {
    useStore.getState().updateConnectionState({ 
      error: error.message,
      connecting: false,
    })
  },
  
  onNodeStatusUpdate: (data) => {
    useStore.getState().updateNodeStatus(data.nodeId, data)
  },
  
  onModelSyncUpdate: (data) => {
    useStore.getState().updateModelSync(data.modelName, {
      status: data.status,
      progress: data.progress,
      nodesTotal: data.nodesTotal || 0,
      nodesSynced: data.nodesSynced || 0,
      error: data.error,
    })
  },
  
  onClusterSummaryUpdate: (data) => {
    useStore.getState().updateDashboardSummary(data)
  },
  
  onActivityUpdate: (data) => {
    useStore.getState().addRecentActivity(data)
  },
  
  onTaskUpdate: (data) => {
    useStore.getState().updateTaskStatus(data.taskId, data)
  },
  
  onTransferProgress: (data) => {
    useStore.getState().updateTransferProgress(data.transferId, data.progress)
  },
  
  onClusterMetrics: (data) => {
    useStore.getState().updateMetrics(data)
  },
  
  onPerformanceAlert: (data) => {
    useStore.getState().addAlert(data)
  },
  
  onSecurityAlert: (data) => {
    useStore.getState().addAlert(data)
  },
  
  onSystemNotification: (data) => {
    useStore.getState().addNotification({
      type: 'info',
      title: data.title,
      message: data.message,
      duration: 10000,
    })
  },
  
  onUserNotification: (data) => {
    useStore.getState().addNotification({
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
})

// Subscribe to auth state changes
authService.subscribe((authState) => {
  useStore.setState((state) => {
    state.auth = authState
  })
})

// Auto-connect WebSocket when authenticated
useStore.subscribe(
  (state) => state.auth.isAuthenticated,
  (isAuthenticated, prevIsAuthenticated) => {
    if (isAuthenticated && !prevIsAuthenticated) {
      // User just logged in - connect WebSocket and fetch initial data
      const store = useStore.getState()
      store.connect().then(() => {
        // Subscribe to real-time updates
        store.subscribeToModelUpdates()
        store.subscribeToNodeUpdates()
        store.subscribeToMetricsUpdates()
        store.subscribeToDashboardUpdates()
        
        // Fetch initial data
        Promise.all([
          store.fetchDashboardData(),
          store.fetchNodes(),
          store.fetchModels(),
          store.fetchMetrics(),
        ]).catch(console.error)
      }).catch(console.error)
    } else if (!isAuthenticated && prevIsAuthenticated) {
      // User just logged out - disconnect WebSocket and clear data
      const store = useStore.getState()
      store.disconnect()
      store.stopAutoRefresh()
      
      // Clear sensitive data
      store.setState((state) => {
        state.models = {
          models: [],
          loading: false,
          error: null,
          syncStatus: {},
        }
        state.nodes = []
        state.monitoring = {
          ...state.monitoring,
          metrics: null,
          alerts: [],
        }
        state.dashboard = {
          ...state.dashboard,
          summary: {
            totalNodes: 0,
            healthyNodes: 0,
            totalModels: 0,
            activeTasks: 0,
            systemHealth: 'healthy',
          },
          recentActivity: [],
        }
      })
    }
  }
)

// Auto-refresh monitoring data when enabled
useStore.subscribe(
  (state) => [state.monitoring.autoRefresh, state.auth.isAuthenticated],
  ([autoRefresh, isAuthenticated], [prevAutoRefresh]) => {
    if (isAuthenticated && autoRefresh && !prevAutoRefresh) {
      useStore.getState().startAutoRefresh()
    } else if (!autoRefresh && prevAutoRefresh) {
      useStore.getState().stopAutoRefresh()
    }
  }
)

// Initialize auth state from stored data
const initialAuthState = authService.getCurrentState()
if (initialAuthState.isAuthenticated) {
  useStore.setState((state) => {
    state.auth = initialAuthState
  })
}

// Cleanup on page unload
if (typeof window !== 'undefined') {
  window.addEventListener('beforeunload', () => {
    const store = useStore.getState()
    store.stopAutoRefresh()
    store.disconnect()
  })
}

// Memoized selectors for performance
export const useModelsSelector = () => useStore((state) => state.models)
export const useNodesSelector = () => useStore((state) => state.nodes)
export const useMonitoringSelector = () => useStore((state) => state.monitoring)
export const useDashboardSelector = () => useStore((state) => state.dashboard)
export const useClusterSelector = () => useStore((state) => state.cluster)

// Derived selectors
export const useSystemHealthSelector = () => useStore((state) => {
  const { nodes } = state
  const healthyNodes = nodes.filter(n => n.status === 'online').length
  const totalNodes = nodes.length
  
  if (totalNodes === 0) return 'unknown'
  if (healthyNodes === totalNodes) return 'healthy'
  if (healthyNodes >= totalNodes * 0.7) return 'warning'
  return 'critical'
})

export const useModelsSummarySelector = () => useStore((state) => {
  const { models } = state.models
  return {
    total: models.length,
    syncing: models.filter(m => m.syncStatus.status === 'syncing').length,
    failed: models.filter(m => m.syncStatus.status === 'failed').length,
    synchronized: models.filter(m => m.syncStatus.status === 'synchronized').length,
  }
})

export const useActiveAlertsSelector = () => useStore((state) => {
  return state.monitoring.alerts.filter(alert => !alert.acknowledged && !alert.resolvedAt)
})

export default useStore