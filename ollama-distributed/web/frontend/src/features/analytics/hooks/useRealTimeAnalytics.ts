/**
 * Real-time Analytics Hook
 * WebSocket-based real-time analytics data streaming and management
 */

import { useState, useEffect, useCallback, useRef } from 'react'
import { useWebSocket } from '../../../hooks/useWebSocket'
import { analyticsService } from '../services/analyticsService'
import { 
  RealTimeMetrics, 
  AnalyticsEvent, 
  EnhancedPerformanceMetrics,
  BusinessMetrics,
  ErrorEvent 
} from '../types'

interface RealTimeConfig {
  enablePerformanceTracking?: boolean
  enableErrorTracking?: boolean
  enableBusinessMetrics?: boolean
  updateInterval?: number
  bufferSize?: number
  autoConnect?: boolean
}

interface RealTimeState {
  isConnected: boolean
  metrics: RealTimeMetrics | null
  events: AnalyticsEvent[]
  errors: ErrorEvent[]
  lastUpdate: number | null
  connectionStatus: 'connecting' | 'connected' | 'disconnected' | 'error'
  stats: {
    totalEvents: number
    eventsPerSecond: number
    errorRate: number
    uptimeSeconds: number
  }
}

const DEFAULT_CONFIG: Required<RealTimeConfig> = {
  enablePerformanceTracking: true,
  enableErrorTracking: true,
  enableBusinessMetrics: false,
  updateInterval: 5000, // 5 seconds
  bufferSize: 1000,
  autoConnect: true
}

export function useRealTimeAnalytics(config: RealTimeConfig = {}) {
  const mergedConfig = { ...DEFAULT_CONFIG, ...config }
  
  const [state, setState] = useState<RealTimeState>({
    isConnected: false,
    metrics: null,
    events: [],
    errors: [],
    lastUpdate: null,
    connectionStatus: 'disconnected',
    stats: {
      totalEvents: 0,
      eventsPerSecond: 0,
      errorRate: 0,
      uptimeSeconds: 0
    }
  })

  const metricsBufferRef = useRef<AnalyticsEvent[]>([])
  const connectionStartRef = useRef<number | null>(null)
  const eventCounterRef = useRef(0)
  const errorCounterRef = useRef(0)

  const { 
    sendMessage, 
    lastMessage, 
    isConnected,
    connectionState,
    error 
  } = useWebSocket({
    url: '/ws/realtime-analytics',
    autoConnect: mergedConfig.autoConnect
  })

  // Handle incoming WebSocket messages
  useEffect(() => {
    if (!lastMessage) return

    try {
      const data = JSON.parse(lastMessage.data)
      
      switch (data.type) {
        case 'real_time_metrics':
          handleRealTimeMetrics(data.payload)
          break
          
        case 'analytics_event':
          handleAnalyticsEvent(data.payload)
          break
          
        case 'performance_update':
          handlePerformanceUpdate(data.payload)
          break
          
        case 'error_event':
          handleErrorEvent(data.payload)
          break
          
        case 'business_metrics':
          handleBusinessMetrics(data.payload)
          break
          
        case 'connection_stats':
          handleConnectionStats(data.payload)
          break
          
        default:
          console.warn('Unknown real-time message type:', data.type)
      }
    } catch (error) {
      console.error('Failed to parse real-time message:', error)
    }
  }, [lastMessage])

  // Connection state management
  useEffect(() => {
    setState(prev => ({
      ...prev,
      isConnected,
      connectionStatus: connectionState === 'CONNECTED' ? 'connected' :
                      connectionState === 'CONNECTING' ? 'connecting' :
                      connectionState === 'DISCONNECTED' ? 'disconnected' : 'error'
    }))

    if (isConnected && !connectionStartRef.current) {
      connectionStartRef.current = Date.now()
      
      // Subscribe to real-time updates
      sendMessage({
        type: 'subscribe',
        config: {
          performance: mergedConfig.enablePerformanceTracking,
          errors: mergedConfig.enableErrorTracking,
          business: mergedConfig.enableBusinessMetrics,
          interval: mergedConfig.updateInterval
        }
      })
    } else if (!isConnected) {
      connectionStartRef.current = null
    }
  }, [isConnected, connectionState, sendMessage])

  // Statistics calculation interval
  useEffect(() => {
    if (!isConnected) return

    const interval = setInterval(() => {
      const now = Date.now()
      const recentEvents = metricsBufferRef.current.filter(
        event => now - event.timestamp < 60000 // Last minute
      )
      
      const eventsPerSecond = recentEvents.length / 60
      const errorEvents = recentEvents.filter(event => event.type === 'error')
      const errorRate = recentEvents.length > 0 ? (errorEvents.length / recentEvents.length) * 100 : 0
      const uptimeSeconds = connectionStartRef.current 
        ? Math.floor((now - connectionStartRef.current) / 1000)
        : 0

      setState(prev => ({
        ...prev,
        stats: {
          totalEvents: eventCounterRef.current,
          eventsPerSecond,
          errorRate,
          uptimeSeconds
        }
      }))
    }, 5000)

    return () => clearInterval(interval)
  }, [isConnected])

  // Event handlers
  const handleRealTimeMetrics = useCallback((metrics: RealTimeMetrics) => {
    setState(prev => ({
      ...prev,
      metrics,
      lastUpdate: Date.now()
    }))
  }, [])

  const handleAnalyticsEvent = useCallback((event: AnalyticsEvent) => {
    eventCounterRef.current++
    
    // Add to buffer
    metricsBufferRef.current.push(event)
    
    // Maintain buffer size
    if (metricsBufferRef.current.length > mergedConfig.bufferSize) {
      metricsBufferRef.current = metricsBufferRef.current.slice(-mergedConfig.bufferSize)
    }

    setState(prev => ({
      ...prev,
      events: [event, ...prev.events].slice(0, 100) // Keep last 100 events in state
    }))

    // Process event through analytics service
    analyticsService.track(
      event.type,
      event.category,
      event.action,
      event.label,
      event.value,
      event.metadata
    )
  }, [mergedConfig.bufferSize])

  const handlePerformanceUpdate = useCallback((performance: EnhancedPerformanceMetrics) => {
    setState(prev => ({
      ...prev,
      metrics: prev.metrics ? {
        ...prev.metrics,
        performance
      } : null
    }))
  }, [])

  const handleErrorEvent = useCallback((error: ErrorEvent) => {
    errorCounterRef.current++
    
    setState(prev => ({
      ...prev,
      errors: [error, ...prev.errors].slice(0, 50) // Keep last 50 errors
    }))

    // Track error through analytics service
    analyticsService.trackError(error.message, {
      type: error.type,
      severity: error.severity,
      stack: error.stack,
      url: error.url
    })
  }, [])

  const handleBusinessMetrics = useCallback((metrics: BusinessMetrics) => {
    setState(prev => ({
      ...prev,
      metrics: prev.metrics ? {
        ...prev.metrics,
        timestamp: Date.now(),
        // Would merge business metrics here
        activeUsers: prev.metrics.activeUsers,
        pageViews: prev.metrics.pageViews,
        events: prev.metrics.events,
        performance: prev.metrics.performance,
        errors: prev.metrics.errors
      } : null
    }))
  }, [])

  const handleConnectionStats = useCallback((stats: any) => {
    setState(prev => ({
      ...prev,
      stats: { ...prev.stats, ...stats }
    }))
  }, [])

  // Action methods
  const subscribe = useCallback((eventTypes: string[]) => {
    if (!isConnected) return

    sendMessage({
      type: 'subscribe_events',
      eventTypes
    })
  }, [isConnected, sendMessage])

  const unsubscribe = useCallback((eventTypes: string[]) => {
    if (!isConnected) return

    sendMessage({
      type: 'unsubscribe_events',
      eventTypes
    })
  }, [isConnected, sendMessage])

  const requestSnapshot = useCallback(() => {
    if (!isConnected) return

    sendMessage({
      type: 'request_snapshot'
    })
  }, [isConnected, sendMessage])

  const clearBuffer = useCallback(() => {
    metricsBufferRef.current = []
    setState(prev => ({
      ...prev,
      events: [],
      errors: []
    }))
  }, [])

  const getEventsByType = useCallback((type: string) => {
    return state.events.filter(event => event.type === type)
  }, [state.events])

  const getEventsByCategory = useCallback((category: string) => {
    return state.events.filter(event => event.category === category)
  }, [state.events])

  const getMetricsInTimeRange = useCallback((startTime: number, endTime: number) => {
    return metricsBufferRef.current.filter(
      event => event.timestamp >= startTime && event.timestamp <= endTime
    )
  }, [])

  const exportBufferData = useCallback(() => {
    return {
      events: metricsBufferRef.current,
      stats: state.stats,
      connectionInfo: {
        isConnected: state.isConnected,
        connectionStatus: state.connectionStatus,
        lastUpdate: state.lastUpdate,
        uptime: state.stats.uptimeSeconds
      },
      timestamp: Date.now()
    }
  }, [state])

  // Performance tracking integration
  useEffect(() => {
    if (!mergedConfig.enablePerformanceTracking) return

    const performanceObserver = new PerformanceObserver((list) => {
      const entries = list.getEntries()
      entries.forEach(entry => {
        if (entry.entryType === 'navigation') {
          const navEntry = entry as PerformanceNavigationTiming
          handleAnalyticsEvent({
            id: `perf_${Date.now()}`,
            type: 'performance',
            category: 'navigation',
            action: 'page_load',
            sessionId: analyticsService.sessionId || '',
            timestamp: Date.now(),
            metadata: {
              loadTime: navEntry.loadEventEnd - navEntry.loadEventStart,
              domContentLoaded: navEntry.domContentLoadedEventEnd - navEntry.domContentLoadedEventStart,
              firstByte: navEntry.responseStart - navEntry.requestStart
            },
            context: {
              url: window.location.href,
              referrer: document.referrer,
              userAgent: navigator.userAgent,
              viewport: {
                width: window.innerWidth,
                height: window.innerHeight
              },
              device: {
                type: 'desktop', // Would detect actual device type
                os: 'unknown',
                browser: 'unknown',
                version: 'unknown'
              }
            }
          })
        }
      })
    })

    try {
      performanceObserver.observe({ entryTypes: ['navigation', 'paint', 'largest-contentful-paint'] })
    } catch (e) {
      console.warn('Performance Observer not supported')
    }

    return () => performanceObserver.disconnect()
  }, [mergedConfig.enablePerformanceTracking, handleAnalyticsEvent])

  // Error tracking integration
  useEffect(() => {
    if (!mergedConfig.enableErrorTracking) return

    const handleError = (event: ErrorEvent) => {
      handleErrorEvent({
        id: `error_${Date.now()}`,
        type: 'javascript',
        message: event.error?.message || event.message || 'Unknown error',
        stack: event.error?.stack,
        url: window.location.href,
        line: event.lineno,
        column: event.colno,
        sessionId: analyticsService.sessionId || '',
        timestamp: Date.now(),
        severity: 'medium',
        resolved: false,
        tags: ['javascript', 'runtime']
      })
    }

    const handleUnhandledRejection = (event: PromiseRejectionEvent) => {
      handleErrorEvent({
        id: `error_${Date.now()}`,
        type: 'javascript',
        message: event.reason?.message || 'Unhandled promise rejection',
        stack: event.reason?.stack,
        url: window.location.href,
        sessionId: analyticsService.sessionId || '',
        timestamp: Date.now(),
        severity: 'medium',
        resolved: false,
        tags: ['javascript', 'promise', 'unhandled']
      })
    }

    window.addEventListener('error', handleError)
    window.addEventListener('unhandledrejection', handleUnhandledRejection)

    return () => {
      window.removeEventListener('error', handleError)
      window.removeEventListener('unhandledrejection', handleUnhandledRejection)
    }
  }, [mergedConfig.enableErrorTracking, handleErrorEvent])

  return {
    // State
    ...state,
    config: mergedConfig,
    
    // Actions
    subscribe,
    unsubscribe,
    requestSnapshot,
    clearBuffer,
    
    // Query methods
    getEventsByType,
    getEventsByCategory,
    getMetricsInTimeRange,
    exportBufferData,
    
    // Connection control
    connect: () => sendMessage({ type: 'connect' }),
    disconnect: () => sendMessage({ type: 'disconnect' }),
    
    // Utility
    isHealthy: isConnected && !error && state.connectionStatus === 'connected'
  }
}

export default useRealTimeAnalytics