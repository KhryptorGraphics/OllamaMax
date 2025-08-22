/**
 * Sprint C Store Usage Examples
 * Demonstrates how to use the enhanced Zustand store in dashboard, models, nodes, and monitoring pages
 */

import { useStore, useModelsSelector, useNodesSelector, useMonitoringSelector, useDashboardSelector, useSystemHealthSelector, useModelsSummarySelector, useActiveAlertsSelector } from './index'

// ============================================================================
// Dashboard Page Examples
// ============================================================================

export const DashboardPageExamples = {
  /**
   * Load dashboard data with caching and real-time updates
   */
  loadDashboardData: () => {
    const { fetchDashboardData, subscribeToDashboardUpdates } = useStore.getState()
    
    // Fetch initial data (cached if available)
    fetchDashboardData()
    
    // Subscribe to real-time updates
    subscribeToDashboardUpdates()
  },

  /**
   * Get dashboard summary with derived state
   */
  useDashboardSummary: () => {
    const dashboard = useDashboardSelector()
    const systemHealth = useSystemHealthSelector()
    const modelsSummary = useModelsSummarySelector()
    const activeAlerts = useActiveAlertsSelector()

    return {
      summary: dashboard.summary,
      systemHealth,
      modelsSummary,
      activeAlertsCount: activeAlerts.length,
      lastUpdated: dashboard.lastUpdated,
      isLoading: dashboard.loading,
      error: dashboard.error,
    }
  },

  /**
   * Refresh dashboard data manually
   */
  refreshDashboard: async () => {
    const { refreshDashboardData } = useStore.getState()
    await refreshDashboardData()
  },
}

// ============================================================================
// Models Page Examples  
// ============================================================================

export const ModelsPageExamples = {
  /**
   * Load models with real-time sync status updates
   */
  loadModels: () => {
    const { fetchModels, subscribeToModelUpdates } = useStore.getState()
    
    // Fetch models (cached if available)
    fetchModels()
    
    // Subscribe to sync status updates
    subscribeToModelUpdates()
  },

  /**
   * Deploy model with optimistic updates
   */
  deployModel: async (modelName: string, targetNodes?: string[]) => {
    const { deployModel } = useStore.getState()
    
    try {
      // This will show immediate UI feedback before API call completes
      await deployModel(modelName, targetNodes)
      console.log(`Model ${modelName} deployment started`)
    } catch (error) {
      console.error(`Failed to deploy ${modelName}:`, error)
      // UI will automatically revert on error
    }
  },

  /**
   * Upload new model with progress tracking
   */
  uploadModel: async (file: File, metadata: any) => {
    const { uploadModel, setModelsLoading } = useStore.getState()
    
    try {
      setModelsLoading(true)
      await uploadModel(file, metadata)
      console.log(`Model ${file.name} uploaded successfully`)
    } catch (error) {
      console.error(`Failed to upload ${file.name}:`, error)
    }
  },

  /**
   * Get models state with derived data
   */
  useModelsState: () => {
    const models = useModelsSelector()
    const modelsSummary = useModelsSummarySelector()

    return {
      models: models.models,
      summary: modelsSummary,
      syncStatus: models.syncStatus,
      isLoading: models.loading,
      error: models.error,
    }
  },
}

// ============================================================================
// Nodes Page Examples
// ============================================================================

export const NodesPageExamples = {
  /**
   * Load nodes with real-time status updates
   */
  loadNodes: () => {
    const { fetchNodes, subscribeToNodeUpdates } = useStore.getState()
    
    // Fetch nodes (cached if available)
    fetchNodes()
    
    // Subscribe to status updates
    subscribeToNodeUpdates()
  },

  /**
   * Drain node with optimistic updates
   */
  drainNode: async (nodeId: string) => {
    const { drainNode } = useStore.getState()
    
    try {
      // Shows immediate UI feedback
      await drainNode(nodeId)
      console.log(`Node ${nodeId} draining started`)
    } catch (error) {
      console.error(`Failed to drain node ${nodeId}:`, error)
      // UI reverts automatically
    }
  },

  /**
   * Get detailed node information
   */
  getNodeDetails: async (nodeId: string) => {
    const { fetchNodeDetails, fetchNodeMetrics, getNodeFromCache } = useStore.getState()
    
    // Try cache first
    let node = getNodeFromCache(nodeId)
    if (!node) {
      node = await fetchNodeDetails(nodeId)
    }
    
    // Fetch current metrics
    await fetchNodeMetrics(nodeId)
    
    return node
  },

  /**
   * Use nodes state with system health
   */
  useNodesState: () => {
    const nodes = useNodesSelector()
    const systemHealth = useSystemHealthSelector()

    return {
      nodes,
      systemHealth,
      totalNodes: nodes.length,
      healthyNodes: nodes.filter(n => n.status === 'online').length,
      isLoading: useStore((state) => state.nodesLoading),
      error: useStore((state) => state.nodesError),
    }
  },
}

// ============================================================================
// Monitoring Page Examples
// ============================================================================

export const MonitoringPageExamples = {
  /**
   * Start real-time monitoring
   */
  startMonitoring: () => {
    const { 
      fetchMetrics, 
      fetchAlerts, 
      subscribeToMetricsUpdates, 
      setAutoRefresh 
    } = useStore.getState()
    
    // Fetch initial data
    Promise.all([fetchMetrics(), fetchAlerts()])
    
    // Enable real-time updates
    subscribeToMetricsUpdates()
    setAutoRefresh(true)
  },

  /**
   * Stop monitoring and cleanup
   */
  stopMonitoring: () => {
    const { 
      unsubscribeFromMetricsUpdates, 
      setAutoRefresh, 
      stopAutoRefresh 
    } = useStore.getState()
    
    unsubscribeFromMetricsUpdates()
    setAutoRefresh(false)
    stopAutoRefresh()
  },

  /**
   * Handle alerts
   */
  handleAlert: async (alertId: string, action: 'acknowledge' | 'resolve') => {
    const { acknowledgeAlert, resolveAlert } = useStore.getState()
    
    try {
      if (action === 'acknowledge') {
        await acknowledgeAlert(alertId)
      } else {
        await resolveAlert(alertId)
      }
      console.log(`Alert ${alertId} ${action}d successfully`)
    } catch (error) {
      console.error(`Failed to ${action} alert:`, error)
    }
  },

  /**
   * Use monitoring state with alerts
   */
  useMonitoringState: () => {
    const monitoring = useMonitoringSelector()
    const activeAlerts = useActiveAlertsSelector()

    return {
      metrics: monitoring.metrics,
      alerts: monitoring.alerts,
      activeAlerts,
      criticalAlerts: activeAlerts.filter(a => a.severity === 'critical'),
      autoRefresh: monitoring.autoRefresh,
      refreshInterval: monitoring.refreshInterval,
      lastUpdated: monitoring.lastUpdated,
      isLoading: monitoring.loading,
      error: monitoring.error,
    }
  },

  /**
   * Configure monitoring settings
   */
  configureMonitoring: (settings: { autoRefresh?: boolean; refreshInterval?: number }) => {
    const { setAutoRefresh, setRefreshInterval } = useStore.getState()
    
    if (settings.autoRefresh !== undefined) {
      setAutoRefresh(settings.autoRefresh)
    }
    
    if (settings.refreshInterval !== undefined) {
      setRefreshInterval(settings.refreshInterval)
    }
  },
}

// ============================================================================
// Cross-Component Integration Examples
// ============================================================================

export const IntegrationExamples = {
  /**
   * Initialize app data on authentication
   */
  initializeAppData: async () => {
    const store = useStore.getState()
    
    if (!store.auth.isAuthenticated) {
      console.warn('User not authenticated')
      return
    }

    try {
      // Load all initial data in parallel
      await Promise.all([
        store.fetchDashboardData(),
        store.fetchNodes(),
        store.fetchModels(),
        store.fetchMetrics(),
      ])

      // Start real-time subscriptions
      store.subscribeToModelUpdates()
      store.subscribeToNodeUpdates()
      store.subscribeToMetricsUpdates()
      store.subscribeToDashboardUpdates()

      console.log('App data initialized successfully')
    } catch (error) {
      console.error('Failed to initialize app data:', error)
    }
  },

  /**
   * Cleanup on logout
   */
  cleanupOnLogout: () => {
    const store = useStore.getState()
    
    // Stop all real-time subscriptions
    store.unsubscribeFromModelUpdates()
    store.unsubscribeFromNodeUpdates()
    store.unsubscribeFromMetricsUpdates()
    store.unsubscribeFromDashboardUpdates()
    
    // Stop auto-refresh
    store.stopAutoRefresh()
    
    // Clear caches
    store.invalidateModelsCache()
    store.invalidateNodesCache()
    store.invalidateMetricsCache()
    store.invalidateDashboardCache()
    
    console.log('App cleanup completed')
  },

  /**
   * Get comprehensive system status
   */
  getSystemStatus: () => {
    const dashboard = useDashboardSelector()
    const systemHealth = useSystemHealthSelector()
    const modelsSummary = useModelsSummarySelector()
    const activeAlerts = useActiveAlertsSelector()
    const monitoring = useMonitoringSelector()

    return {
      overall: {
        health: systemHealth,
        status: dashboard.summary.systemHealth,
        lastUpdated: Math.max(
          dashboard.lastUpdated || 0,
          monitoring.lastUpdated || 0
        ),
      },
      nodes: {
        total: dashboard.summary.totalNodes,
        healthy: dashboard.summary.healthyNodes,
        percentage: dashboard.summary.totalNodes > 0 
          ? Math.round((dashboard.summary.healthyNodes / dashboard.summary.totalNodes) * 100)
          : 0,
      },
      models: modelsSummary,
      alerts: {
        total: activeAlerts.length,
        critical: activeAlerts.filter(a => a.severity === 'critical').length,
        high: activeAlerts.filter(a => a.severity === 'high').length,
      },
      performance: monitoring.metrics ? {
        cpu: monitoring.metrics.system.cpu.current,
        memory: monitoring.metrics.system.memory.current,
        network: monitoring.metrics.system.network.throughput,
      } : null,
    }
  },

  /**
   * Handle errors globally
   */
  handleGlobalError: (error: Error, context: string) => {
    const { addNotification } = useStore.getState()
    
    console.error(`[${context}] Error:`, error)
    
    addNotification({
      type: 'error',
      title: 'System Error',
      message: `An error occurred in ${context}: ${error.message}`,
      duration: 10000,
    })
  },
}

// ============================================================================
// Performance Monitoring Examples
// ============================================================================

export const PerformanceExamples = {
  /**
   * Monitor store performance
   */
  monitorStorePerformance: () => {
    const startTime = performance.now()
    
    // Subscribe to state changes
    return useStore.subscribe((state, prevState) => {
      const endTime = performance.now()
      const duration = endTime - startTime
      
      if (duration > 100) { // Log slow updates
        console.warn(`Slow store update detected: ${duration}ms`)
      }
    })
  },

  /**
   * Batch operations for better performance
   */
  batchOperations: async (operations: (() => Promise<void>)[]) => {
    const { setDashboardLoading } = useStore.getState()
    
    try {
      setDashboardLoading(true)
      
      // Execute operations in parallel
      await Promise.all(operations.map(op => op()))
      
      console.log(`Batch operation completed: ${operations.length} operations`)
    } catch (error) {
      console.error('Batch operation failed:', error)
    } finally {
      setDashboardLoading(false)
    }
  },
}

export default {
  DashboardPageExamples,
  ModelsPageExamples,
  NodesPageExamples,
  MonitoringPageExamples,
  IntegrationExamples,
  PerformanceExamples,
}