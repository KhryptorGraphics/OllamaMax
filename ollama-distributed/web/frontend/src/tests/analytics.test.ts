/**
 * Analytics System Tests
 * Comprehensive test suite for analytics and reporting functionality
 */

import { describe, it, expect, beforeEach, vi, Mock } from 'vitest'
import { AnalyticsService } from '../features/analytics/services/analyticsService'
import { DataAggregator, TimePeriod, AggregationType } from '../features/analytics/utils/dataAggregation'
import { ExportService } from '../features/reporting/services/exportService'
import { AnalyticsEvent, AnalyticsEventType } from '../features/analytics/types'

// Mock WebSocket
class MockWebSocket {
  static CONNECTING = 0
  static OPEN = 1
  static CLOSING = 2
  static CLOSED = 3

  readyState = MockWebSocket.OPEN
  onopen: ((event: Event) => void) | null = null
  onclose: ((event: CloseEvent) => void) | null = null
  onmessage: ((event: MessageEvent) => void) | null = null
  onerror: ((event: Event) => void) | null = null

  constructor(public url: string) {
    setTimeout(() => {
      if (this.onopen) {
        this.onopen(new Event('open'))
      }
    }, 0)
  }

  send(data: string) {
    console.log('WebSocket send:', data)
  }

  close() {
    this.readyState = MockWebSocket.CLOSED
  }
}

// Mock global objects
Object.defineProperty(window, 'WebSocket', {
  value: MockWebSocket
})

Object.defineProperty(window, 'location', {
  value: {
    href: 'https://example.com/test',
    protocol: 'https:'
  }
})

Object.defineProperty(navigator, 'userAgent', {
  value: 'Mozilla/5.0 (Test Browser)'
})

Object.defineProperty(navigator, 'hardwareConcurrency', {
  value: 8
})

// Mock performance API
Object.defineProperty(window, 'performance', {
  value: {
    now: vi.fn(() => Date.now()),
    timing: {},
    memory: {
      usedJSHeapSize: 10000000,
      totalJSHeapSize: 20000000,
      jsHeapSizeLimit: 100000000
    },
    getEntriesByType: vi.fn(() => []),
    getEntriesByName: vi.fn(() => [])
  }
})

// Mock localStorage
const mockLocalStorage = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn()
}
Object.defineProperty(window, 'localStorage', { value: mockLocalStorage })

describe('AnalyticsService', () => {
  let analyticsService: AnalyticsService

  beforeEach(() => {
    vi.clearAllMocks()
    analyticsService = AnalyticsService.getInstance()
  })

  it('should track page views correctly', () => {
    const trackSpy = vi.spyOn(analyticsService, 'track')
    
    analyticsService.trackPageView('/test-page', 'Test Page')
    
    expect(trackSpy).toHaveBeenCalledWith(
      'page_view',
      'navigation',
      'page_view',
      '/test-page',
      undefined,
      expect.objectContaining({
        title: 'Test Page',
        path: '/test-page'
      })
    )
  })

  it('should track user interactions', () => {
    const trackSpy = vi.spyOn(analyticsService, 'track')
    
    analyticsService.trackClick('login-button')
    
    expect(trackSpy).toHaveBeenCalledWith(
      'click',
      'interaction',
      'click',
      'login-button',
      undefined,
      expect.objectContaining({
        element: 'login-button'
      })
    )
  })

  it('should track API calls with performance metrics', () => {
    const trackSpy = vi.spyOn(analyticsService, 'track')
    
    analyticsService.trackApiCall('/api/users', 'GET', 200, 150, 1024)
    
    expect(trackSpy).toHaveBeenCalledWith(
      'api_call',
      'api',
      'GET',
      '/api/users',
      150,
      expect.objectContaining({
        endpoint: '/api/users',
        method: 'GET',
        status: 200,
        duration: 150,
        size: 1024,
        success: true
      })
    )
  })

  it('should track errors with proper context', () => {
    const trackSpy = vi.spyOn(analyticsService, 'track')
    const testError = new Error('Test error')
    
    analyticsService.trackError(testError, { component: 'TestComponent' })
    
    expect(trackSpy).toHaveBeenCalledWith(
      'error',
      'error',
      'javascript_error',
      'Test error',
      undefined,
      expect.objectContaining({
        message: 'Test error',
        component: 'TestComponent'
      })
    )
  })

  it('should export data in different formats', async () => {
    // Add some test events
    analyticsService.trackPageView('/page1')
    analyticsService.trackClick('button1')
    analyticsService.trackApiCall('/api/test', 'POST', 201, 200)

    const csvData = await analyticsService.exportData('csv')
    expect(typeof csvData).toBe('string')
    expect(csvData).toContain('id,type,category,action')

    const jsonData = await analyticsService.exportData('json')
    expect(typeof jsonData).toBe('string')
    const parsed = JSON.parse(jsonData)
    expect(Array.isArray(parsed)).toBe(true)
  })

  it('should handle privacy compliance', () => {
    analyticsService.setUserId('test-user-123')
    analyticsService.trackPageView('/private-page')
    
    analyticsService.anonymizeUser()
    
    // Verify user ID is removed
    const realtimeMetrics = analyticsService.getRealTimeMetrics()
    expect(realtimeMetrics.events.every(event => !event.userId)).toBe(true)
  })
})

describe('DataAggregator', () => {
  let aggregator: DataAggregator
  let testData: AnalyticsEvent[]

  beforeEach(() => {
    aggregator = DataAggregator.getInstance()
    testData = [
      {
        id: '1',
        type: 'page_view' as AnalyticsEventType,
        category: 'navigation',
        action: 'page_view',
        sessionId: 'session-1',
        timestamp: Date.now() - 3600000, // 1 hour ago
        context: {
          url: 'https://example.com/page1',
          referrer: '',
          userAgent: 'Test',
          viewport: { width: 1920, height: 1080 },
          device: { type: 'desktop', os: 'Test', browser: 'Test', version: '1.0' }
        }
      },
      {
        id: '2',
        type: 'click' as AnalyticsEventType,
        category: 'interaction',
        action: 'click',
        sessionId: 'session-1',
        timestamp: Date.now() - 1800000, // 30 minutes ago
        context: {
          url: 'https://example.com/page1',
          referrer: '',
          userAgent: 'Test',
          viewport: { width: 1920, height: 1080 },
          device: { type: 'desktop', os: 'Test', browser: 'Test', version: '1.0' }
        }
      },
      {
        id: '3',
        type: 'api_call' as AnalyticsEventType,
        category: 'api',
        action: 'GET',
        value: 150,
        sessionId: 'session-2',
        timestamp: Date.now() - 900000, // 15 minutes ago
        context: {
          url: 'https://example.com/page2',
          referrer: '',
          userAgent: 'Test',
          viewport: { width: 1920, height: 1080 },
          device: { type: 'mobile', os: 'Test', browser: 'Test', version: '1.0' }
        }
      }
    ]
  })

  it('should aggregate data by category', () => {
    const result = aggregator.aggregate(testData, {
      groupBy: ['category'],
      metrics: [
        { field: 'id', aggregation: 'count' as AggregationType }
      ]
    })

    expect(result.groups).toHaveLength(3)
    expect(result.groups.find(g => g.group_0 === 'navigation')?.id_count).toBe(1)
    expect(result.groups.find(g => g.group_0 === 'interaction')?.id_count).toBe(1)
    expect(result.groups.find(g => g.group_0 === 'api')?.id_count).toBe(1)
  })

  it('should calculate time series data', () => {
    const timeSeries = aggregator.aggregateTimeSeries(
      testData,
      'timestamp',
      'value',
      TimePeriod.HOUR
    )

    expect(Array.isArray(timeSeries)).toBe(true)
    expect(timeSeries.length).toBeGreaterThan(0)
    expect(timeSeries[0]).toHaveProperty('timestamp')
    expect(timeSeries[0]).toHaveProperty('value')
  })

  it('should analyze cohorts', () => {
    const cohortData = [
      { userId: 'user1', firstSeen: Date.now() - 86400000 * 30, returned: Date.now() - 86400000 * 25 },
      { userId: 'user1', firstSeen: Date.now() - 86400000 * 30, returned: Date.now() - 86400000 * 23 },
      { userId: 'user2', firstSeen: Date.now() - 86400000 * 30, returned: Date.now() - 86400000 * 20 },
      { userId: 'user3', firstSeen: Date.now() - 86400000 * 15, returned: Date.now() - 86400000 * 10 }
    ]

    const cohorts = aggregator.analyzeCohorts(
      cohortData,
      'firstSeen',
      'returned',
      'userId'
    )

    expect(Array.isArray(cohorts)).toBe(true)
    expect(cohorts.length).toBeGreaterThan(0)
    expect(cohorts[0]).toHaveProperty('cohort')
    expect(cohorts[0]).toHaveProperty('size')
    expect(cohorts[0]).toHaveProperty('retention')
  })

  it('should analyze funnels', () => {
    const funnelData = [
      { userId: 'user1', action: 'page_view', page: '/landing' },
      { userId: 'user1', action: 'click', element: 'signup-button' },
      { userId: 'user1', action: 'page_view', page: '/signup' },
      { userId: 'user1', action: 'form_submit', form: 'signup-form' },
      { userId: 'user2', action: 'page_view', page: '/landing' },
      { userId: 'user2', action: 'click', element: 'signup-button' },
      { userId: 'user3', action: 'page_view', page: '/landing' }
    ]

    const funnel = aggregator.analyzeFunnel(
      funnelData,
      [
        { name: 'Landing Page', condition: (item: any) => item.page === '/landing' },
        { name: 'Clicked Signup', condition: (item: any) => item.element === 'signup-button' },
        { name: 'Visited Signup Page', condition: (item: any) => item.page === '/signup' },
        { name: 'Completed Signup', condition: (item: any) => item.form === 'signup-form' }
      ],
      'userId'
    )

    expect(Array.isArray(funnel)).toBe(true)
    expect(funnel).toHaveLength(4)
    expect(funnel[0].step).toBe('Landing Page')
    expect(funnel[0].users).toBe(3)
    expect(funnel[3].step).toBe('Completed Signup')
    expect(funnel[3].users).toBe(1)
  })

  it('should calculate statistics', () => {
    const values = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
    const stats = aggregator.calculateStatistics(values)

    expect(stats.count).toBe(10)
    expect(stats.sum).toBe(55)
    expect(stats.mean).toBe(5.5)
    expect(stats.median).toBe(5.5)
    expect(stats.min).toBe(1)
    expect(stats.max).toBe(10)
    expect(stats.standardDeviation).toBeCloseTo(2.87, 2)
  })

  it('should detect anomalies', () => {
    const values = [1, 2, 3, 4, 5, 100, 6, 7, 8, 9] // 100 is an outlier
    const anomalies = aggregator.detectAnomalies(values, 'zscore', 2)

    expect(Array.isArray(anomalies)).toBe(true)
    expect(anomalies.length).toBeGreaterThan(0)
    expect(anomalies[0].value).toBe(100)
    expect(anomalies[0].score).toBeGreaterThan(2)
  })
})

describe('ExportService', () => {
  let exportService: ExportService

  beforeEach(() => {
    exportService = ExportService.getInstance()
    
    // Mock URL.createObjectURL
    global.URL.createObjectURL = vi.fn(() => 'blob:mock-url')
    global.URL.revokeObjectURL = vi.fn()
    
    // Mock DOM methods
    document.createElement = vi.fn((tag) => {
      if (tag === 'a') {
        return {
          href: '',
          download: '',
          click: vi.fn(),
          style: {}
        } as any
      }
      return {} as any
    })
    
    document.body.appendChild = vi.fn()
    document.body.removeChild = vi.fn()
  })

  it('should export data as JSON', async () => {
    const testData = {
      title: 'Test Report',
      data: [
        { id: 1, name: 'Test 1', value: 100 },
        { id: 2, name: 'Test 2', value: 200 }
      ]
    }

    const result = await exportService.exportReport(testData, 'json')

    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('filename')
    expect(result.filename).toMatch(/\.json$/)
    expect(result).toHaveProperty('downloadUrl')
    expect(result).toHaveProperty('size')
  })

  it('should export data as CSV', async () => {
    const testData = {
      title: 'Test Report',
      data: [
        { id: 1, name: 'Test 1', value: 100 },
        { id: 2, name: 'Test 2', value: 200 }
      ]
    }

    const result = await exportService.exportReport(testData, 'csv')

    expect(result).toHaveProperty('id')
    expect(result.filename).toMatch(/\.csv$/)
    expect(result.size).toBeGreaterThan(0)
  })

  it('should export data as PDF', async () => {
    const testData = {
      title: 'Test Report',
      data: [
        { id: 1, name: 'Test 1', value: 100 },
        { id: 2, name: 'Test 2', value: 200 }
      ],
      summary: {
        totalRecords: 2,
        totalValue: 300
      }
    }

    const result = await exportService.exportReport(testData, 'pdf', {
      format: 'pdf',
      includeCharts: true,
      includeData: true,
      compressed: false,
      branding: true
    })

    expect(result).toHaveProperty('id')
    expect(result.filename).toMatch(/\.pdf$/)
    expect(result.size).toBeGreaterThan(0)
  })

  it('should handle export options correctly', async () => {
    const testData = {
      title: 'Confidential Report',
      data: [{ id: 1, sensitive: 'data' }]
    }

    const result = await exportService.exportReport(testData, 'pdf', {
      format: 'pdf',
      includeCharts: false,
      includeData: true,
      compressed: true,
      password: 'secret123',
      watermark: 'CONFIDENTIAL',
      branding: false
    })

    expect(result.password).toBe(true)
  })
})

describe('Analytics Integration', () => {
  it('should handle end-to-end analytics flow', async () => {
    const analyticsService = AnalyticsService.getInstance()
    const aggregator = DataAggregator.getInstance()
    const exportService = ExportService.getInstance()

    // Track some events
    analyticsService.trackPageView('/dashboard')
    analyticsService.trackClick('export-button')
    analyticsService.trackApiCall('/api/data', 'GET', 200, 150)

    // Get real-time metrics
    const metrics = analyticsService.getRealTimeMetrics()
    expect(metrics).toHaveProperty('activeUsers')
    expect(metrics).toHaveProperty('events')
    expect(metrics.events.length).toBeGreaterThan(0)

    // Aggregate the data
    const aggregated = aggregator.aggregate(metrics.events, {
      groupBy: ['type'],
      metrics: [
        { field: 'timestamp', aggregation: 'count' as AggregationType }
      ]
    })

    expect(aggregated.groups.length).toBeGreaterThan(0)

    // Export the results
    const exportData = {
      title: 'Analytics Summary',
      data: aggregated.groups,
      summary: aggregated.totals
    }

    const exported = await exportService.exportReport(exportData, 'json')
    expect(exported.filename).toMatch(/analytics_summary.*\.json/)
  })
})