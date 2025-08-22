/**
 * React hook for monitoring data management
 * Provides comprehensive monitoring state and real-time updates
 */

import { useState, useEffect, useCallback, useMemo } from 'react'
import { useWebSocket } from './useWebSocket'
import {
  MonitoringState,
  MonitoringMetrics,
  MonitoringAlert,
  LogEntry,
  AlertThreshold,
  DashboardConfig,
  TimeRange,
  MetricFilter,
  ExportOptions,
  AlertSeverity,
  LogLevel
} from '../types/monitoring'

interface UseMonitoringConfig {
  autoRefresh?: boolean
  refreshInterval?: number
  maxLogEntries?: number
  maxAlerts?: number
  enableRealTime?: boolean
}

interface UseMonitoringReturn {
  // State
  state: MonitoringState
  
  // Metrics
  metrics: MonitoringMetrics | null
  filteredMetrics: MonitoringMetrics | null
  
  // Alerts
  alerts: MonitoringAlert[]
  activeAlerts: MonitoringAlert[]
  criticalAlerts: MonitoringAlert[]
  
  // Logs
  logs: LogEntry[]
  filteredLogs: LogEntry[]
  
  // Configuration
  thresholds: AlertThreshold[]
  dashboardConfig: DashboardConfig
  
  // Actions
  refreshMetrics: () => Promise<void>
  acknowledgeAlert: (alertId: string) => Promise<void>
  resolveAlert: (alertId: string) => Promise<void>
  createThreshold: (threshold: Omit<AlertThreshold, 'id' | 'createdAt' | 'updatedAt'>) => Promise<void>
  updateThreshold: (id: string, threshold: Partial<AlertThreshold>) => Promise<void>
  deleteThreshold: (id: string) => Promise<void>
  exportData: (options: ExportOptions) => Promise<void>
  
  // Filtering
  setTimeRange: (range: TimeRange) => void
  setMetricFilters: (filters: MetricFilter[]) => void
  setLogFilters: (filters: { level?: LogLevel; source?: string; search?: string }) => void
  setAlertFilters: (filters: { severity?: AlertSeverity; acknowledged?: boolean; category?: string }) => void
  
  // Dashboard
  updateDashboardConfig: (config: Partial<DashboardConfig>) => void
  
  // Real-time status
  isConnected: boolean
  lastUpdate: number | null
  error: string | null
}

const DEFAULT_CONFIG: Required<UseMonitoringConfig> = {
  autoRefresh: true,
  refreshInterval: 30000, // 30 seconds
  maxLogEntries: 1000,
  maxAlerts: 100,
  enableRealTime: true
}

const DEFAULT_DASHBOARD_CONFIG: DashboardConfig = {
  widgets: [],
  layout: {
    columns: 12,
    padding: 8,
    margin: 16,
    compact: false
  },
  timeRange: {
    start: -3600000, // Last hour
    end: 0,
    preset: '1h'
  },
  refreshInterval: 30000,
  autoRefresh: true,
  theme: 'auto'
}

export function useMonitoring(config: UseMonitoringConfig = {}): UseMonitoringReturn {
  const mergedConfig = { ...DEFAULT_CONFIG, ...config }
  
  // Core state
  const [state, setState] = useState<MonitoringState>({
    metrics: null,
    alerts: [],
    logs: [],
    thresholds: [],
    dashboardConfig: DEFAULT_DASHBOARD_CONFIG,
    loading: false,
    error: null,
    lastUpdated: null
  })
  
  // Filtering state
  const [timeRange, setTimeRange] = useState<TimeRange>(DEFAULT_DASHBOARD_CONFIG.timeRange)
  const [metricFilters, setMetricFilters] = useState<MetricFilter[]>([])
  const [logFilters, setLogFilters] = useState<{
    level?: LogLevel
    source?: string
    search?: string
  }>({})
  const [alertFilters, setAlertFilters] = useState<{
    severity?: AlertSeverity
    acknowledged?: boolean
    category?: string
  }>({})
  
  // WebSocket for real-time updates
  const {
    isConnected,
    subscribe
  } = useWebSocket({
    autoConnect: mergedConfig.enableRealTime
  })
  
  // Fetch initial data
  const fetchMetrics = useCallback(async () => {
    setState(prev => ({ ...prev, loading: true, error: null }))
    
    try {
      const response = await fetch('/api/v1/metrics', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json'
        }
      })
      
      if (!response.ok) {
        throw new Error(`Failed to fetch metrics: ${response.statusText}`)
      }
      
      const metrics: MonitoringMetrics = await response.json()
      
      setState(prev => ({
        ...prev,
        metrics,
        loading: false,
        lastUpdated: Date.now()
      }))
    } catch (error) {
      setState(prev => ({
        ...prev,
        loading: false,
        error: error instanceof Error ? error.message : 'Unknown error'
      }))
    }
  }, [])
  
  const fetchAlerts = useCallback(async () => {
    try {
      const response = await fetch('/api/v1/alerts')
      if (!response.ok) throw new Error('Failed to fetch alerts')
      
      const alerts: MonitoringAlert[] = await response.json()
      
      setState(prev => ({
        ...prev,
        alerts: alerts.slice(0, mergedConfig.maxAlerts)
      }))
    } catch (error) {
      console.error('Failed to fetch alerts:', error)
    }
  }, [mergedConfig.maxAlerts])
  
  const fetchLogs = useCallback(async () => {
    try {
      const response = await fetch('/api/v1/logs')
      if (!response.ok) throw new Error('Failed to fetch logs')
      
      const logs: LogEntry[] = await response.json()
      
      setState(prev => ({
        ...prev,
        logs: logs.slice(0, mergedConfig.maxLogEntries)
      }))
    } catch (error) {
      console.error('Failed to fetch logs:', error)
    }
  }, [mergedConfig.maxLogEntries])
  
  const fetchThresholds = useCallback(async () => {
    try {
      const response = await fetch('/api/v1/thresholds')
      if (!response.ok) throw new Error('Failed to fetch thresholds')
      
      const thresholds: AlertThreshold[] = await response.json()
      
      setState(prev => ({ ...prev, thresholds }))
    } catch (error) {
      console.error('Failed to fetch thresholds:', error)
    }
  }, [])
  
  // Refresh all data
  const refreshMetrics = useCallback(async () => {
    await Promise.all([
      fetchMetrics(),
      fetchAlerts(),
      fetchLogs(),
      fetchThresholds()
    ])
  }, [fetchMetrics, fetchAlerts, fetchLogs, fetchThresholds])
  
  // Alert management
  const acknowledgeAlert = useCallback(async (alertId: string) => {
    try {
      const response = await fetch(`/api/v1/alerts/${alertId}/acknowledge`, {
        method: 'POST'
      })
      if (!response.ok) throw new Error('Failed to acknowledge alert')
      
      setState(prev => ({
        ...prev,
        alerts: prev.alerts.map(alert =>
          alert.id === alertId
            ? { ...alert, acknowledged: true, acknowledgedAt: Date.now() }
            : alert
        )
      }))
    } catch (error) {
      console.error('Failed to acknowledge alert:', error)
    }
  }, [])
  
  const resolveAlert = useCallback(async (alertId: string) => {
    try {
      const response = await fetch(`/api/v1/alerts/${alertId}/resolve`, {
        method: 'POST'
      })
      if (!response.ok) throw new Error('Failed to resolve alert')
      
      setState(prev => ({
        ...prev,
        alerts: prev.alerts.map(alert =>
          alert.id === alertId
            ? { ...alert, resolvedAt: Date.now() }
            : alert
        )
      }))
    } catch (error) {
      console.error('Failed to resolve alert:', error)
    }
  }, [])
  
  // Threshold management
  const createThreshold = useCallback(async (threshold: Omit<AlertThreshold, 'id' | 'createdAt' | 'updatedAt'>) => {
    try {
      const response = await fetch('/api/v1/thresholds', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(threshold)
      })
      if (!response.ok) throw new Error('Failed to create threshold')
      
      const newThreshold: AlertThreshold = await response.json()
      
      setState(prev => ({
        ...prev,
        thresholds: [...prev.thresholds, newThreshold]
      }))
    } catch (error) {
      console.error('Failed to create threshold:', error)
    }
  }, [])
  
  const updateThreshold = useCallback(async (id: string, threshold: Partial<AlertThreshold>) => {
    try {
      const response = await fetch(`/api/v1/thresholds/${id}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(threshold)
      })
      if (!response.ok) throw new Error('Failed to update threshold')
      
      const updatedThreshold: AlertThreshold = await response.json()
      
      setState(prev => ({
        ...prev,
        thresholds: prev.thresholds.map(t =>
          t.id === id ? updatedThreshold : t
        )
      }))
    } catch (error) {
      console.error('Failed to update threshold:', error)
    }
  }, [])
  
  const deleteThreshold = useCallback(async (id: string) => {
    try {
      const response = await fetch(`/api/v1/thresholds/${id}`, {
        method: 'DELETE'
      })
      if (!response.ok) throw new Error('Failed to delete threshold')
      
      setState(prev => ({
        ...prev,
        thresholds: prev.thresholds.filter(t => t.id !== id)
      }))
    } catch (error) {
      console.error('Failed to delete threshold:', error)
    }
  }, [])
  
  // Export functionality
  const exportData = useCallback(async (options: ExportOptions) => {
    try {
      const response = await fetch('/api/v1/export', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(options)
      })
      if (!response.ok) throw new Error('Failed to export data')
      
      const blob = await response.blob()
      const url = URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = `monitoring-export-${Date.now()}.${options.format}`
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      URL.revokeObjectURL(url)
    } catch (error) {
      console.error('Failed to export data:', error)
    }
  }, [])
  
  // Dashboard configuration
  const updateDashboardConfig = useCallback((config: Partial<DashboardConfig>) => {
    setState(prev => ({
      ...prev,
      dashboardConfig: { ...prev.dashboardConfig, ...config }
    }))
  }, [])
  
  // Filtered data computed values
  const filteredMetrics = useMemo(() => {
    if (!state.metrics || metricFilters.length === 0) return state.metrics
    
    // Apply metric filters (implementation would depend on specific filtering logic)
    return state.metrics
  }, [state.metrics, metricFilters])
  
  const filteredLogs = useMemo(() => {
    let filtered = [...state.logs]
    
    if (logFilters.level) {
      filtered = filtered.filter(log => log.level === logFilters.level)
    }
    
    if (logFilters.source) {
      filtered = filtered.filter(log => log.source.includes(logFilters.source!))
    }
    
    if (logFilters.search) {
      const search = logFilters.search.toLowerCase()
      filtered = filtered.filter(log =>
        log.message.toLowerCase().includes(search) ||
        log.source.toLowerCase().includes(search)
      )
    }
    
    return filtered
  }, [state.logs, logFilters])
  
  const activeAlerts = useMemo(() => {
    return state.alerts.filter(alert => !alert.resolvedAt)
  }, [state.alerts])
  
  const criticalAlerts = useMemo(() => {
    return activeAlerts.filter(alert => alert.severity === 'critical')
  }, [activeAlerts])
  
  // Auto-refresh setup
  useEffect(() => {
    if (!mergedConfig.autoRefresh) return
    
    const interval = setInterval(refreshMetrics, mergedConfig.refreshInterval)
    return () => clearInterval(interval)
  }, [mergedConfig.autoRefresh, mergedConfig.refreshInterval, refreshMetrics])
  
  // Real-time updates setup
  useEffect(() => {
    if (!mergedConfig.enableRealTime) return
    
    const unsubscribeMetrics = subscribe('metrics', (data) => {
      setState(prev => ({
        ...prev,
        metrics: data,
        lastUpdated: Date.now()
      }))
    })
    
    const unsubscribeAlerts = subscribe('alerts', (data) => {
      setState(prev => ({
        ...prev,
        alerts: data.slice(0, mergedConfig.maxAlerts)
      }))
    })
    
    const unsubscribeLogs = subscribe('logs', (data) => {
      setState(prev => ({
        ...prev,
        logs: [...data, ...prev.logs].slice(0, mergedConfig.maxLogEntries)
      }))
    })
    
    return () => {
      unsubscribeMetrics()
      unsubscribeAlerts()
      unsubscribeLogs()
    }
  }, [mergedConfig.enableRealTime, mergedConfig.maxAlerts, mergedConfig.maxLogEntries, subscribe])
  
  // Initial data fetch
  useEffect(() => {
    refreshMetrics()
  }, [])
  
  return {
    // State
    state,
    
    // Metrics
    metrics: state.metrics,
    filteredMetrics,
    
    // Alerts
    alerts: state.alerts,
    activeAlerts,
    criticalAlerts,
    
    // Logs
    logs: state.logs,
    filteredLogs,
    
    // Configuration
    thresholds: state.thresholds,
    dashboardConfig: state.dashboardConfig,
    
    // Actions
    refreshMetrics,
    acknowledgeAlert,
    resolveAlert,
    createThreshold,
    updateThreshold,
    deleteThreshold,
    exportData,
    
    // Filtering
    setTimeRange,
    setMetricFilters,
    setLogFilters,
    setAlertFilters,
    
    // Dashboard
    updateDashboardConfig,
    
    // Real-time status
    isConnected,
    lastUpdate: state.lastUpdated,
    error: state.error
  }
}