/**
 * Performance Reporting Dashboard
 * Real-time performance metrics visualization and monitoring
 */

import React, { useState, useEffect, useCallback, useMemo } from 'react'
import { Card } from '../../../design-system/components/Card/Card'
import { Button } from '../../../design-system/components/Button/Button'
import { Select } from '../../../design-system/components/Select/Select'
import { Chart } from '../../analytics/components/charts/Chart'
import { useWebSocket } from '../../../hooks/useWebSocket'
import { usePerformanceMonitor, PerformanceMonitor } from '../../../utils/performance'
import { EnhancedPerformanceMetrics, WebVitalsMetrics } from '../../analytics/types'
import {
  Activity,
  Zap,
  Clock,
  TrendingUp,
  TrendingDown,
  AlertCircle,
  CheckCircle,
  Cpu,
  Memory,
  Network,
  HardDrive,
  Monitor,
  Smartphone,
  Globe,
  RefreshCw,
  Download,
  Settings,
  Filter,
  Calendar
} from 'lucide-react'

interface PerformanceAlert {
  id: string
  type: 'warning' | 'error' | 'info'
  metric: string
  message: string
  value: number
  threshold: number
  timestamp: number
}

interface PerformanceThresholds {
  fcp: { good: number; needs_improvement: number }
  lcp: { good: number; needs_improvement: number }
  fid: { good: number; needs_improvement: number }
  cls: { good: number; needs_improvement: number }
  ttfb: { good: number; needs_improvement: number }
}

const DEFAULT_THRESHOLDS: PerformanceThresholds = {
  fcp: { good: 1800, needs_improvement: 3000 },
  lcp: { good: 2500, needs_improvement: 4000 },
  fid: { good: 100, needs_improvement: 300 },
  cls: { good: 0.1, needs_improvement: 0.25 },
  ttfb: { good: 800, needs_improvement: 1800 }
}

export const PerformanceDashboard: React.FC = () => {
  const [timeRange, setTimeRange] = useState('1h')
  const [deviceFilter, setDeviceFilter] = useState('all')
  const [browserFilter, setBrowserFilter] = useState('all')
  const [metrics, setMetrics] = useState<EnhancedPerformanceMetrics | null>(null)
  const [historicalData, setHistoricalData] = useState<any[]>([])
  const [alerts, setAlerts] = useState<PerformanceAlert[]>([])
  const [loading, setLoading] = useState(true)
  const [thresholds, setThresholds] = useState<PerformanceThresholds>(DEFAULT_THRESHOLDS)
  const [autoRefresh, setAutoRefresh] = useState(true)

  const { sendMessage, lastMessage, isConnected } = useWebSocket()
  const performanceMonitor = PerformanceMonitor.getInstance()

  // Subscribe to real-time performance updates
  useEffect(() => {
    if (lastMessage) {
      const data = JSON.parse(lastMessage.data)
      
      if (data.type === 'performance_metrics') {
        setMetrics(data.metrics)
        setLoading(false)
      } else if (data.type === 'performance_history') {
        setHistoricalData(data.history)
      } else if (data.type === 'performance_alerts') {
        setAlerts(data.alerts)
      }
    }
  }, [lastMessage])

  // Subscribe to performance updates
  useEffect(() => {
    sendMessage({ 
      action: 'subscribe_performance',
      timeRange,
      filters: { device: deviceFilter, browser: browserFilter }
    })
    
    return () => sendMessage({ action: 'unsubscribe_performance' })
  }, [timeRange, deviceFilter, browserFilter, sendMessage])

  // Auto-refresh data
  useEffect(() => {
    if (!autoRefresh) return

    const interval = setInterval(() => {
      sendMessage({ action: 'refresh_performance' })
    }, 30000) // Refresh every 30 seconds

    return () => clearInterval(interval)
  }, [autoRefresh, sendMessage])

  // Calculate performance scores
  const performanceScores = useMemo(() => {
    if (!metrics?.webVitals) return null

    const calculateScore = (value: number, metric: keyof PerformanceThresholds) => {
      const threshold = thresholds[metric]
      if (value <= threshold.good) return { score: 'good', color: 'green' }
      if (value <= threshold.needs_improvement) return { score: 'needs_improvement', color: 'yellow' }
      return { score: 'poor', color: 'red' }
    }

    return {
      fcp: calculateScore(metrics.webVitals.fcp, 'fcp'),
      lcp: calculateScore(metrics.webVitals.lcp, 'lcp'),
      fid: calculateScore(metrics.webVitals.fid, 'fid'),
      cls: calculateScore(metrics.webVitals.cls, 'cls'),
      ttfb: calculateScore(metrics.webVitals.ttfb, 'ttfb')
    }
  }, [metrics, thresholds])

  const formatMetric = useCallback((value: number, unit: string) => {
    if (unit === 'ms') return `${value.toFixed(0)}ms`
    if (unit === 's') return `${(value / 1000).toFixed(2)}s`
    if (unit === 'score') return value.toFixed(3)
    if (unit === 'MB') return `${(value / 1024 / 1024).toFixed(1)}MB`
    return value.toString()
  }, [])

  const exportData = useCallback(async (format: 'csv' | 'json' | 'pdf') => {
    try {
      const exportData = {
        metrics,
        historicalData,
        alerts,
        timeRange,
        filters: { device: deviceFilter, browser: browserFilter },
        generatedAt: new Date().toISOString()
      }

      if (format === 'json') {
        const blob = new Blob([JSON.stringify(exportData, null, 2)], {
          type: 'application/json'
        })
        const url = URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = `performance-report-${Date.now()}.json`
        a.click()
        URL.revokeObjectURL(url)
      } else {
        // Send to server for PDF/CSV processing
        sendMessage({ 
          action: 'export_performance',
          format,
          data: exportData
        })
      }
    } catch (error) {
      console.error('Export failed:', error)
    }
  }, [metrics, historicalData, alerts, timeRange, deviceFilter, browserFilter, sendMessage])

  const refreshData = useCallback(() => {
    setLoading(true)
    sendMessage({ action: 'refresh_performance' })
  }, [sendMessage])

  if (loading && !metrics) {
    return (
      <div className="flex items-center justify-center h-96">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold">Performance Dashboard</h1>
          <p className="text-gray-600 mt-1">Real-time performance monitoring and Web Vitals</p>
        </div>

        <div className="flex items-center gap-3">
          <div className="flex items-center gap-2">
            <div className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-500' : 'bg-red-500'}`} />
            <span className="text-sm text-gray-600">
              {isConnected ? 'Connected' : 'Disconnected'}
            </span>
          </div>
          
          <Button
            onClick={() => setAutoRefresh(!autoRefresh)}
            variant="outline"
            size="sm"
            className={autoRefresh ? 'bg-green-50 text-green-700' : ''}
          >
            <RefreshCw className={`w-4 h-4 mr-2 ${autoRefresh ? 'animate-spin' : ''}`} />
            Auto Refresh
          </Button>

          <Select
            value={timeRange}
            onChange={setTimeRange}
            className="w-32"
          >
            <option value="5m">5 Minutes</option>
            <option value="15m">15 Minutes</option>
            <option value="1h">1 Hour</option>
            <option value="6h">6 Hours</option>
            <option value="24h">24 Hours</option>
            <option value="7d">7 Days</option>
          </Select>

          <Button onClick={refreshData} variant="outline" size="sm">
            <RefreshCw className="w-4 h-4 mr-2" />
            Refresh
          </Button>

          <Button onClick={() => exportData('json')} variant="outline" size="sm">
            <Download className="w-4 h-4 mr-2" />
            Export
          </Button>
        </div>
      </div>

      {/* Alerts */}
      {alerts.length > 0 && (
        <Card className="border-l-4 border-l-red-500">
          <div className="flex items-start gap-3">
            <AlertCircle className="w-5 h-5 text-red-500 mt-0.5" />
            <div className="flex-1">
              <h3 className="font-medium text-red-900">Performance Alerts</h3>
              <div className="mt-2 space-y-1">
                {alerts.slice(0, 3).map(alert => (
                  <p key={alert.id} className="text-sm text-red-700">
                    {alert.message} ({formatMetric(alert.value, 'ms')} vs threshold {formatMetric(alert.threshold, 'ms')})
                  </p>
                ))}
                {alerts.length > 3 && (
                  <p className="text-sm text-red-600">+{alerts.length - 3} more alerts</p>
                )}
              </div>
            </div>
          </div>
        </Card>
      )}

      {/* Web Vitals Overview */}
      <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
        {metrics?.webVitals && Object.entries(metrics.webVitals).map(([key, value]) => {
          const score = performanceScores?.[key as keyof WebVitalsMetrics]
          const metricInfo = {
            fcp: { label: 'First Contentful Paint', icon: <Zap className="w-5 h-5" />, unit: 'ms' },
            lcp: { label: 'Largest Contentful Paint', icon: <Monitor className="w-5 h-5" />, unit: 'ms' },
            fid: { label: 'First Input Delay', icon: <Clock className="w-5 h-5" />, unit: 'ms' },
            cls: { label: 'Cumulative Layout Shift', icon: <Activity className="w-5 h-5" />, unit: 'score' },
            ttfb: { label: 'Time to First Byte', icon: <Network className="w-5 h-5" />, unit: 'ms' }
          }[key as keyof WebVitalsMetrics]

          if (!metricInfo) return null

          return (
            <Card key={key}>
              <div className="flex items-center justify-between mb-3">
                <div className={`p-2 rounded-lg ${
                  score?.color === 'green' ? 'bg-green-100 text-green-700' :
                  score?.color === 'yellow' ? 'bg-yellow-100 text-yellow-700' :
                  'bg-red-100 text-red-700'
                }`}>
                  {metricInfo.icon}
                </div>
                <div className={`px-2 py-1 rounded text-xs font-medium ${
                  score?.color === 'green' ? 'bg-green-100 text-green-700' :
                  score?.color === 'yellow' ? 'bg-yellow-100 text-yellow-700' :
                  'bg-red-100 text-red-700'
                }`}>
                  {score?.score.replace('_', ' ').toUpperCase()}
                </div>
              </div>
              
              <div className="space-y-1">
                <p className="text-2xl font-bold">
                  {formatMetric(value as number, metricInfo.unit)}
                </p>
                <p className="text-sm text-gray-600">{metricInfo.label}</p>
              </div>
            </Card>
          )
        })}
      </div>

      {/* Performance Trends */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <div className="p-4">
            <h3 className="text-lg font-semibold mb-4">Web Vitals Trends</h3>
            <Chart
              type="line"
              data={historicalData.filter(d => d.metric === 'webVitals')}
              config={{
                xKey: 'timestamp',
                yKey: ['fcp', 'lcp', 'ttfb'],
                smooth: true,
                showGrid: true,
                showLegend: true,
                colors: ['#3b82f6', '#10b981', '#f59e0b'],
                yAxisFormatter: (value: number) => `${value}ms`
              }}
              height={300}
            />
          </div>
        </Card>

        <Card>
          <div className="p-4">
            <h3 className="text-lg font-semibold mb-4">Resource Usage</h3>
            {metrics?.runtime && (
              <div className="space-y-4">
                <div>
                  <div className="flex justify-between items-center mb-2">
                    <div className="flex items-center gap-2">
                      <Memory className="w-4 h-4 text-blue-500" />
                      <span className="text-sm font-medium">Memory Usage</span>
                    </div>
                    <span className="text-sm text-gray-600">
                      {formatMetric(metrics.runtime.memory.used, 'MB')} / {formatMetric(metrics.runtime.memory.total, 'MB')}
                    </span>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-2">
                    <div 
                      className="bg-blue-500 h-2 rounded-full"
                      style={{ width: `${(metrics.runtime.memory.used / metrics.runtime.memory.total) * 100}%` }}
                    />
                  </div>
                </div>

                <div>
                  <div className="flex justify-between items-center mb-2">
                    <div className="flex items-center gap-2">
                      <Cpu className="w-4 h-4 text-green-500" />
                      <span className="text-sm font-medium">CPU Cores</span>
                    </div>
                    <span className="text-sm text-gray-600">
                      {metrics.runtime.cpu.cores} cores
                    </span>
                  </div>
                </div>

                {metrics.runtime.battery && (
                  <div>
                    <div className="flex justify-between items-center mb-2">
                      <div className="flex items-center gap-2">
                        <Smartphone className="w-4 h-4 text-orange-500" />
                        <span className="text-sm font-medium">Battery</span>
                      </div>
                      <span className="text-sm text-gray-600">
                        {(metrics.runtime.battery.level * 100).toFixed(0)}%
                        {metrics.runtime.battery.charging && ' (Charging)'}
                      </span>
                    </div>
                    <div className="w-full bg-gray-200 rounded-full h-2">
                      <div 
                        className={`h-2 rounded-full ${
                          metrics.runtime.battery.level > 0.5 ? 'bg-green-500' :
                          metrics.runtime.battery.level > 0.2 ? 'bg-yellow-500' : 'bg-red-500'
                        }`}
                        style={{ width: `${metrics.runtime.battery.level * 100}%` }}
                      />
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>
        </Card>
      </div>

      {/* Network Performance */}
      <Card>
        <div className="p-4">
          <h3 className="text-lg font-semibold mb-4">Network Performance</h3>
          {metrics?.network && (
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
              <div className="text-center">
                <div className="flex items-center justify-center mb-2">
                  <Globe className="w-5 h-5 text-blue-500" />
                </div>
                <p className="text-lg font-semibold">{metrics.network.effectiveType}</p>
                <p className="text-sm text-gray-600">Connection Type</p>
              </div>

              <div className="text-center">
                <div className="flex items-center justify-center mb-2">
                  <Network className="w-5 h-5 text-green-500" />
                </div>
                <p className="text-lg font-semibold">{metrics.network.rtt}ms</p>
                <p className="text-sm text-gray-600">Round Trip Time</p>
              </div>

              <div className="text-center">
                <div className="flex items-center justify-center mb-2">
                  <TrendingDown className="w-5 h-5 text-purple-500" />
                </div>
                <p className="text-lg font-semibold">{metrics.network.downlink} Mbps</p>
                <p className="text-sm text-gray-600">Download Speed</p>
              </div>

              <div className="text-center">
                <div className="flex items-center justify-center mb-2">
                  <HardDrive className={`w-5 h-5 ${metrics.network.saveData ? 'text-orange-500' : 'text-gray-400'}`} />
                </div>
                <p className="text-lg font-semibold">
                  {metrics.network.saveData ? 'ON' : 'OFF'}
                </p>
                <p className="text-sm text-gray-600">Data Saver</p>
              </div>
            </div>
          )}
        </div>
      </Card>

      {/* User Experience Metrics */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <div className="p-4">
            <h3 className="text-lg font-semibold mb-4">User Experience</h3>
            {metrics?.user && (
              <div className="space-y-4">
                <div className="flex justify-between items-center">
                  <span className="text-sm text-gray-600">Session Duration</span>
                  <span className="font-medium">
                    {Math.round(metrics.user.totalSessionTime / 1000 / 60)} minutes
                  </span>
                </div>

                <div className="flex justify-between items-center">
                  <span className="text-sm text-gray-600">Active Time</span>
                  <span className="font-medium">
                    {Math.round(metrics.user.activeTime / 1000 / 60)} minutes
                  </span>
                </div>

                <div className="flex justify-between items-center">
                  <span className="text-sm text-gray-600">Interactions</span>
                  <span className="font-medium">{metrics.user.interactions}</span>
                </div>

                <div className="flex justify-between items-center">
                  <span className="text-sm text-gray-600">Scroll Depth</span>
                  <span className="font-medium">{metrics.user.scrollDepth}%</span>
                </div>

                <div className="flex justify-between items-center">
                  <span className="text-sm text-gray-600">Bounce Rate</span>
                  <span className="font-medium">{(metrics.user.bounceRate * 100).toFixed(1)}%</span>
                </div>
              </div>
            )}
          </div>
        </Card>

        <Card>
          <div className="p-4">
            <h3 className="text-lg font-semibold mb-4">Performance Score</h3>
            <div className="text-center">
              <div className="relative w-32 h-32 mx-auto mb-4">
                <svg className="w-full h-full transform -rotate-90" viewBox="0 0 100 100">
                  <circle
                    cx="50"
                    cy="50"
                    r="45"
                    stroke="#e5e7eb"
                    strokeWidth="8"
                    fill="none"
                  />
                  <circle
                    cx="50"
                    cy="50"
                    r="45"
                    stroke="#3b82f6"
                    strokeWidth="8"
                    fill="none"
                    strokeLinecap="round"
                    strokeDasharray={`${85 * 2.83} ${100 * 2.83}`}
                    className="transition-all duration-300"
                  />
                </svg>
                <div className="absolute inset-0 flex items-center justify-center">
                  <span className="text-2xl font-bold">85</span>
                </div>
              </div>
              <p className="text-sm text-gray-600">Overall Performance Score</p>
            </div>
          </div>
        </Card>
      </div>
    </div>
  )
}

export default PerformanceDashboard