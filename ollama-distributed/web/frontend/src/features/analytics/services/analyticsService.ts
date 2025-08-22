/**
 * Enhanced Analytics Service
 * Provides comprehensive event tracking, performance monitoring, and reporting
 */

import { 
  AnalyticsEvent, 
  EnhancedPerformanceMetrics, 
  BusinessMetrics,
  RealTimeMetrics,
  ErrorEvent,
  DeviceInfo,
  EventContext,
  AnalyticsEventType 
} from '../types'
import { PerformanceMonitor } from '../../../utils/performance'

export class AnalyticsService {
  private static instance: AnalyticsService
  private events: AnalyticsEvent[] = []
  private sessionId: string
  private userId?: string
  private deviceInfo: DeviceInfo
  private performanceMonitor: PerformanceMonitor
  private websocket: WebSocket | null = null
  private batchSize = 50
  private batchInterval = 30000 // 30 seconds
  private eventQueue: AnalyticsEvent[] = []
  private config = {
    enableRealTime: true,
    enablePerformanceTracking: true,
    enableErrorTracking: true,
    enableUserTracking: true,
    batchEvents: true,
    endpoint: '/api/analytics/events',
    websocketUrl: `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/ws/analytics`
  }

  private constructor() {
    this.sessionId = this.generateSessionId()
    this.deviceInfo = this.detectDevice()
    this.performanceMonitor = PerformanceMonitor.getInstance()
    
    this.initializeTracking()
    this.setupEventListeners()
    this.startBatchProcessor()
  }

  static getInstance(): AnalyticsService {
    if (!AnalyticsService.instance) {
      AnalyticsService.instance = new AnalyticsService()
    }
    return AnalyticsService.instance
  }

  // Configuration
  configure(config: Partial<typeof this.config>): void {
    this.config = { ...this.config, ...config }
    
    if (config.enableRealTime && !this.websocket) {
      this.initializeWebSocket()
    } else if (!config.enableRealTime && this.websocket) {
      this.websocket.close()
      this.websocket = null
    }
  }

  setUserId(userId: string): void {
    this.userId = userId
  }

  // Event Tracking
  track(
    type: AnalyticsEventType,
    category: string,
    action: string,
    label?: string,
    value?: number,
    metadata?: Record<string, any>
  ): void {
    const event: AnalyticsEvent = {
      id: this.generateEventId(),
      type,
      category,
      action,
      label,
      value,
      userId: this.userId,
      sessionId: this.sessionId,
      timestamp: Date.now(),
      metadata,
      context: this.getEventContext()
    }

    this.events.push(event)
    this.eventQueue.push(event)

    if (this.config.enableRealTime && this.websocket) {
      this.sendRealTimeEvent(event)
    }

    // Store in local storage for offline capability
    this.persistEvent(event)
  }

  // Page View Tracking
  trackPageView(path: string, title?: string, referrer?: string): void {
    this.track('page_view', 'navigation', 'page_view', path, undefined, {
      title: title || document.title,
      referrer: referrer || document.referrer,
      path
    })
  }

  // Click Tracking
  trackClick(element: string, category = 'interaction'): void {
    this.track('click', category, 'click', element, undefined, {
      element,
      timestamp: Date.now()
    })
  }

  // Form Tracking
  trackFormSubmit(formName: string, success: boolean, fields?: string[]): void {
    this.track('form_submit', 'form', success ? 'submit_success' : 'submit_error', formName, undefined, {
      formName,
      success,
      fields,
      timestamp: Date.now()
    })
  }

  // API Call Tracking
  trackApiCall(
    endpoint: string,
    method: string,
    status: number,
    duration: number,
    size?: number
  ): void {
    this.track('api_call', 'api', method, endpoint, duration, {
      endpoint,
      method,
      status,
      duration,
      size,
      success: status >= 200 && status < 400
    })
  }

  // Error Tracking
  trackError(error: Error | string, context?: Record<string, any>): void {
    const errorEvent: ErrorEvent = {
      id: this.generateEventId(),
      type: 'javascript',
      message: typeof error === 'string' ? error : error.message,
      stack: typeof error === 'object' ? error.stack : undefined,
      url: window.location.href,
      userId: this.userId,
      sessionId: this.sessionId,
      timestamp: Date.now(),
      severity: 'medium',
      resolved: false,
      tags: []
    }

    this.track('error', 'error', 'javascript_error', errorEvent.message, undefined, {
      ...errorEvent,
      ...context
    })
  }

  // Performance Tracking
  trackPerformance(): void {
    if (!this.config.enablePerformanceTracking) return

    const metrics = this.performanceMonitor.getAllMetrics()
    const componentMetrics = this.performanceMonitor.getComponentMetrics()
    const networkMetrics = this.performanceMonitor.getNetworkMetrics()

    this.track('performance', 'performance', 'web_vitals', undefined, undefined, {
      webVitals: {
        fcp: metrics.fcp,
        lcp: metrics.lcp,
        fid: metrics.fid,
        cls: metrics.cls,
        ttfb: metrics.ttfb
      },
      memory: metrics.memory,
      components: componentMetrics,
      network: networkMetrics
    })
  }

  // User Interaction Tracking
  trackUserInteraction(interactionType: string, target: string, duration?: number): void {
    this.track('user_interaction', 'engagement', interactionType, target, duration, {
      interactionType,
      target,
      duration,
      timestamp: Date.now()
    })
  }

  // Custom Event Tracking
  trackCustomEvent(eventName: string, properties?: Record<string, any>): void {
    this.track('custom', 'custom', eventName, undefined, undefined, properties)
  }

  // Real-time Metrics
  getRealTimeMetrics(): RealTimeMetrics {
    const recentEvents = this.events.filter(
      event => event.timestamp > Date.now() - 60000 // Last minute
    )

    return {
      activeUsers: this.getActiveUsersCount(),
      pageViews: recentEvents.filter(e => e.type === 'page_view').length,
      events: recentEvents,
      performance: this.getEnhancedPerformanceMetrics(),
      errors: this.getRecentErrors(),
      timestamp: Date.now()
    }
  }

  // Business Metrics (placeholder for server-side calculation)
  async getBusinessMetrics(dateRange: { start: number; end: number }): Promise<BusinessMetrics> {
    const response = await fetch('/api/analytics/business-metrics', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ dateRange, sessionId: this.sessionId })
    })
    
    return response.json()
  }

  // Data Export
  async exportData(format: 'json' | 'csv' | 'excel', dateRange?: { start: number; end: number }) {
    const data = dateRange 
      ? this.events.filter(e => e.timestamp >= dateRange.start && e.timestamp <= dateRange.end)
      : this.events

    if (format === 'json') {
      return JSON.stringify(data, null, 2)
    }

    if (format === 'csv') {
      return this.convertToCSV(data)
    }

    // For Excel, send to server for processing
    const response = await fetch('/api/analytics/export', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ data, format })
    })

    return response.blob()
  }

  // Privacy and Compliance
  anonymizeUser(): void {
    this.userId = undefined
    this.events = this.events.map(event => ({
      ...event,
      userId: undefined,
      context: {
        ...event.context,
        location: undefined
      }
    }))
  }

  deleteUserData(userId: string): void {
    this.events = this.events.filter(event => event.userId !== userId)
    this.clearPersistedEvents()
  }

  // Data Retention
  cleanupOldData(retentionDays = 90): void {
    const cutoffDate = Date.now() - (retentionDays * 24 * 60 * 60 * 1000)
    this.events = this.events.filter(event => event.timestamp > cutoffDate)
  }

  // Private Methods
  private generateSessionId(): string {
    return `session_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  }

  private generateEventId(): string {
    return `event_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  }

  private detectDevice(): DeviceInfo {
    const userAgent = navigator.userAgent
    
    return {
      type: this.getDeviceType(),
      os: this.getOS(userAgent),
      browser: this.getBrowser(userAgent),
      version: this.getBrowserVersion(userAgent)
    }
  }

  private getDeviceType(): 'desktop' | 'tablet' | 'mobile' {
    const width = window.innerWidth
    if (width < 768) return 'mobile'
    if (width < 1024) return 'tablet'
    return 'desktop'
  }

  private getOS(userAgent: string): string {
    if (userAgent.includes('Windows')) return 'Windows'
    if (userAgent.includes('Mac')) return 'macOS'
    if (userAgent.includes('Linux')) return 'Linux'
    if (userAgent.includes('Android')) return 'Android'
    if (userAgent.includes('iOS')) return 'iOS'
    return 'Unknown'
  }

  private getBrowser(userAgent: string): string {
    if (userAgent.includes('Chrome')) return 'Chrome'
    if (userAgent.includes('Firefox')) return 'Firefox'
    if (userAgent.includes('Safari')) return 'Safari'
    if (userAgent.includes('Edge')) return 'Edge'
    return 'Unknown'
  }

  private getBrowserVersion(userAgent: string): string {
    const matches = userAgent.match(/(?:Chrome|Firefox|Safari|Edge)\/(\d+\.\d+)/i)
    return matches ? matches[1] : 'Unknown'
  }

  private getEventContext(): EventContext {
    return {
      url: window.location.href,
      referrer: document.referrer,
      userAgent: navigator.userAgent,
      viewport: {
        width: window.innerWidth,
        height: window.innerHeight
      },
      device: this.deviceInfo
    }
  }

  private initializeTracking(): void {
    if (this.config.enablePerformanceTracking) {
      this.performanceMonitor.initializeWebVitals()
      
      // Track performance metrics every 30 seconds
      setInterval(() => {
        this.trackPerformance()
      }, 30000)
    }
  }

  private setupEventListeners(): void {
    // Page visibility change
    document.addEventListener('visibilitychange', () => {
      this.track('user_interaction', 'engagement', 
        document.hidden ? 'page_hidden' : 'page_visible'
      )
    })

    // Unload event
    window.addEventListener('beforeunload', () => {
      this.flushEvents()
    })

    // Error tracking
    if (this.config.enableErrorTracking) {
      window.addEventListener('error', (event) => {
        this.trackError(event.error || event.message, {
          filename: event.filename,
          lineno: event.lineno,
          colno: event.colno
        })
      })

      window.addEventListener('unhandledrejection', (event) => {
        this.trackError(event.reason, {
          type: 'unhandled_promise_rejection'
        })
      })
    }
  }

  private initializeWebSocket(): void {
    if (!this.config.enableRealTime) return

    try {
      this.websocket = new WebSocket(this.config.websocketUrl)
      
      this.websocket.onopen = () => {
        console.log('Analytics WebSocket connected')
      }

      this.websocket.onerror = (error) => {
        console.error('Analytics WebSocket error:', error)
      }

      this.websocket.onclose = () => {
        console.log('Analytics WebSocket disconnected')
        // Attempt reconnection after 5 seconds
        setTimeout(() => {
          this.initializeWebSocket()
        }, 5000)
      }
    } catch (error) {
      console.error('Failed to initialize analytics WebSocket:', error)
    }
  }

  private sendRealTimeEvent(event: AnalyticsEvent): void {
    if (this.websocket && this.websocket.readyState === WebSocket.OPEN) {
      this.websocket.send(JSON.stringify({
        type: 'analytics_event',
        data: event
      }))
    }
  }

  private startBatchProcessor(): void {
    if (!this.config.batchEvents) return

    setInterval(() => {
      this.flushEvents()
    }, this.batchInterval)
  }

  private async flushEvents(): Promise<void> {
    if (this.eventQueue.length === 0) return

    const events = [...this.eventQueue]
    this.eventQueue = []

    try {
      await fetch(this.config.endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ events, sessionId: this.sessionId }),
        keepalive: true
      })
    } catch (error) {
      console.error('Failed to send analytics events:', error)
      // Re-queue events for retry
      this.eventQueue.unshift(...events)
    }
  }

  private persistEvent(event: AnalyticsEvent): void {
    try {
      const stored = localStorage.getItem('analytics_events') || '[]'
      const events: AnalyticsEvent[] = JSON.parse(stored)
      events.push(event)
      
      // Keep only last 1000 events in localStorage
      const trimmed = events.slice(-1000)
      localStorage.setItem('analytics_events', JSON.stringify(trimmed))
    } catch (error) {
      console.warn('Failed to persist analytics event:', error)
    }
  }

  private clearPersistedEvents(): void {
    try {
      localStorage.removeItem('analytics_events')
    } catch (error) {
      console.warn('Failed to clear persisted events:', error)
    }
  }

  private getActiveUsersCount(): number {
    // This would typically be provided by the server via WebSocket
    return 1 // Current user
  }

  private getEnhancedPerformanceMetrics(): EnhancedPerformanceMetrics {
    const webVitals = this.performanceMonitor.getAllMetrics()
    
    return {
      webVitals: {
        fcp: webVitals.fcp,
        lcp: webVitals.lcp,
        fid: webVitals.fid,
        cls: webVitals.cls,
        ttfb: webVitals.ttfb
      },
      runtime: {
        memory: {
          used: webVitals.memory,
          total: (performance as any).memory?.totalJSHeapSize || 0,
          limit: (performance as any).memory?.jsHeapSizeLimit || 0,
          peak: (performance as any).memory?.totalJSHeapSize || 0
        },
        cpu: {
          usage: 0, // Would need server-side calculation
          cores: navigator.hardwareConcurrency || 1,
          speed: 0 // Not available in browser
        }
      },
      network: {
        effectiveType: (navigator as any).connection?.effectiveType || 'unknown',
        rtt: (navigator as any).connection?.rtt || 0,
        downlink: (navigator as any).connection?.downlink || 0,
        saveData: (navigator as any).connection?.saveData || false,
        requests: []
      },
      resources: {
        scripts: [],
        stylesheets: [],
        images: [],
        fonts: [],
        other: []
      },
      user: {
        totalSessionTime: Date.now() - this.getSessionStart(),
        activeTime: 0,
        idleTime: 0,
        interactions: this.events.filter(e => e.type === 'user_interaction').length,
        scrollDepth: 0,
        bounceRate: 0
      }
    }
  }

  private getRecentErrors(): ErrorEvent[] {
    return this.events
      .filter(e => e.type === 'error' && e.timestamp > Date.now() - 3600000) // Last hour
      .map(e => ({
        id: e.id,
        type: 'javascript' as const,
        message: e.metadata?.message || 'Unknown error',
        url: e.context.url,
        userId: e.userId,
        sessionId: e.sessionId,
        timestamp: e.timestamp,
        severity: 'medium' as const,
        resolved: false,
        tags: []
      }))
  }

  private getSessionStart(): number {
    const sessionStart = sessionStorage.getItem('analytics_session_start')
    if (sessionStart) {
      return parseInt(sessionStart)
    }
    
    const now = Date.now()
    sessionStorage.setItem('analytics_session_start', now.toString())
    return now
  }

  private convertToCSV(events: AnalyticsEvent[]): string {
    const headers = ['id', 'type', 'category', 'action', 'label', 'value', 'userId', 'sessionId', 'timestamp', 'url']
    const rows = [headers.join(',')]

    events.forEach(event => {
      const row = [
        event.id,
        event.type,
        event.category,
        event.action,
        event.label || '',
        event.value || '',
        event.userId || '',
        event.sessionId,
        event.timestamp,
        event.context.url
      ]
      rows.push(row.map(field => `"${field}"`).join(','))
    })

    return rows.join('\n')
  }
}

// Export singleton instance
export const analyticsService = AnalyticsService.getInstance()

// React hook for analytics
export function useAnalytics() {
  const service = AnalyticsService.getInstance()

  return {
    track: service.track.bind(service),
    trackPageView: service.trackPageView.bind(service),
    trackClick: service.trackClick.bind(service),
    trackFormSubmit: service.trackFormSubmit.bind(service),
    trackApiCall: service.trackApiCall.bind(service),
    trackError: service.trackError.bind(service),
    trackCustomEvent: service.trackCustomEvent.bind(service),
    getRealTimeMetrics: service.getRealTimeMetrics.bind(service),
    exportData: service.exportData.bind(service),
    configure: service.configure.bind(service),
    setUserId: service.setUserId.bind(service)
  }
}