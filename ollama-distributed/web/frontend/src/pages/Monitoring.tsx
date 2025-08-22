/**
 * Monitoring Page
 * Comprehensive monitoring dashboard with real-time metrics, alerts, and logs
 */

import React, { useState, useEffect, useMemo, useCallback } from 'react'
import {
  Activity,
  TrendingUp,
  AlertTriangle,
  Download,
  RefreshCw,
  Settings,
  Calendar,
  BarChart3,
  Monitor,
  Bell,
  FileText,
  Filter,
  Clock,
  Maximize2,
  Minimize2
} from 'lucide-react'
import { format, subHours, subDays } from 'date-fns'
import { LoadingSpinner } from '../components/common/LoadingSpinner'
import { MetricsGrid } from '../components/monitoring/MetricsGrid'
import { TimeSeriesChart, MultiSeriesChart } from '../components/monitoring/TimeSeriesChart'
import { AlertsPanel } from '../components/monitoring/AlertsPanel'
import { LogViewer } from '../components/monitoring/LogViewer'
import { ThresholdConfig } from '../components/monitoring/ThresholdConfig'
import { useMonitoring } from '../hooks/useMonitoring'
import {
  TimeRange,
  TimeRangePreset,
  ExportOptions,
  DashboardWidget,
  WidgetType,
  ChartType,
  CorrelationMatrix
} from '../types/monitoring'

interface MonitoringPageProps {
  className?: string
}

const TIME_RANGE_PRESETS: { value: TimeRangePreset; label: string }[] = [
  { value: '5m', label: '5 minutes' },
  { value: '15m', label: '15 minutes' },
  { value: '30m', label: '30 minutes' },
  { value: '1h', label: '1 hour' },
  { value: '3h', label: '3 hours' },
  { value: '6h', label: '6 hours' },
  { value: '12h', label: '12 hours' },
  { value: '24h', label: '24 hours' },
  { value: '7d', label: '7 days' },
  { value: '30d', label: '30 days' }
]

const getTimeRangeFromPreset = (preset: TimeRangePreset): TimeRange => {
  const now = new Date()
  
  switch (preset) {
    case '5m': return { start: subHours(now, 0).getTime() - 5 * 60 * 1000, end: now.getTime(), preset }
    case '15m': return { start: subHours(now, 0).getTime() - 15 * 60 * 1000, end: now.getTime(), preset }
    case '30m': return { start: subHours(now, 0).getTime() - 30 * 60 * 1000, end: now.getTime(), preset }
    case '1h': return { start: subHours(now, 1).getTime(), end: now.getTime(), preset }
    case '3h': return { start: subHours(now, 3).getTime(), end: now.getTime(), preset }
    case '6h': return { start: subHours(now, 6).getTime(), end: now.getTime(), preset }
    case '12h': return { start: subHours(now, 12).getTime(), end: now.getTime(), preset }
    case '24h': return { start: subHours(now, 24).getTime(), end: now.getTime(), preset }
    case '7d': return { start: subDays(now, 7).getTime(), end: now.getTime(), preset }
    case '30d': return { start: subDays(now, 30).getTime(), end: now.getTime(), preset }
    default: return { start: subHours(now, 1).getTime(), end: now.getTime(), preset: '1h' }
  }
}

const TabButton: React.FC<{
  active: boolean
  onClick: () => void
  children: React.ReactNode
  icon?: React.ReactNode
}> = ({ active, onClick, children, icon }) => (
  <button
    onClick={onClick}
    className={`
      inline-flex items-center px-4 py-2 text-sm font-medium border-b-2 transition-colors
      ${active
        ? 'border-blue-500 text-blue-600 bg-blue-50'
        : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
      }
    `}
  >
    {icon && <span className="mr-2">{icon}</span>}
    {children}
  </button>
)

const SummaryCard: React.FC<{
  title: string
  value: string | number
  change?: number
  icon?: React.ReactNode
  color?: 'blue' | 'green' | 'yellow' | 'red'
}> = ({ title, value, change, icon, color = 'blue' }) => {
  const colorClasses = {
    blue: 'bg-blue-50 text-blue-700 border-blue-200',
    green: 'bg-green-50 text-green-700 border-green-200',
    yellow: 'bg-yellow-50 text-yellow-700 border-yellow-200',
    red: 'bg-red-50 text-red-700 border-red-200'
  }
  
  return (
    <div className={`p-4 rounded-lg border ${colorClasses[color]}`}>
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center space-x-2">
          {icon}
          <h3 className="text-sm font-medium">{title}</h3>
        </div>
        {change !== undefined && (
          <div className={`text-xs ${change >= 0 ? 'text-green-600' : 'text-red-600'}`}>
            {change >= 0 ? '+' : ''}{change.toFixed(1)}%
          </div>
        )}
      </div>
      <div className="text-2xl font-bold">{value}</div>
    </div>
  )
}

const DashboardControls: React.FC<{
  timeRange: TimeRange
  onTimeRangeChange: (range: TimeRange) => void
  isRealTime: boolean
  onRealTimeToggle: () => void
  onRefresh: () => void
  onExport: () => void
  autoRefresh: boolean
  onAutoRefreshToggle: () => void
}> = ({
  timeRange,
  onTimeRangeChange,
  isRealTime,
  onRealTimeToggle,
  onRefresh,
  onExport,
  autoRefresh,
  onAutoRefreshToggle
}) => {
  const [showCustomRange, setShowCustomRange] = useState(false)
  const [customStart, setCustomStart] = useState('')
  const [customEnd, setCustomEnd] = useState('')
  
  const handlePresetChange = (preset: TimeRangePreset) => {
    const range = getTimeRangeFromPreset(preset)
    onTimeRangeChange(range)
    setShowCustomRange(false)
  }
  
  const handleCustomRangeApply = () => {
    if (customStart && customEnd) {
      onTimeRangeChange({
        start: new Date(customStart).getTime(),
        end: new Date(customEnd).getTime(),
        preset: 'custom'
      })
      setShowCustomRange(false)
    }
  }
  
  return (
    <div className="bg-white border border-gray-200 rounded-lg p-4 mb-6">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div className="flex items-center space-x-4">
          <div className="flex items-center space-x-2">
            <Calendar className="w-4 h-4 text-gray-500" />
            <span className="text-sm font-medium text-gray-700">Time Range:</span>
          </div>
          
          <select
            value={timeRange.preset || 'custom'}
            onChange={(e) => {
              const preset = e.target.value as TimeRangePreset
              if (preset === 'custom') {
                setShowCustomRange(true)
              } else {
                handlePresetChange(preset)
              }
            }}
            className="border border-gray-300 rounded-md px-3 py-1 text-sm"
          >
            {TIME_RANGE_PRESETS.map(preset => (
              <option key={preset.value} value={preset.value}>
                {preset.label}
              </option>
            ))}
            <option value="custom">Custom Range</option>
          </select>
          
          {showCustomRange && (
            <div className="flex items-center space-x-2">
              <input
                type="datetime-local"
                value={customStart}
                onChange={(e) => setCustomStart(e.target.value)}
                className="border border-gray-300 rounded-md px-2 py-1 text-sm"
              />
              <span className="text-gray-500">to</span>
              <input
                type="datetime-local"
                value={customEnd}
                onChange={(e) => setCustomEnd(e.target.value)}
                className="border border-gray-300 rounded-md px-2 py-1 text-sm"
              />
              <button
                onClick={handleCustomRangeApply}
                className="px-3 py-1 bg-blue-600 text-white rounded-md text-sm hover:bg-blue-700"
              >
                Apply
              </button>
            </div>
          )}
        </div>
        
        <div className="flex items-center space-x-2">
          <button
            onClick={onRealTimeToggle}
            className={`inline-flex items-center px-3 py-1 rounded-md text-sm font-medium transition-colors ${
              isRealTime
                ? 'bg-green-100 text-green-800'
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
            }`}
          >
            <div className={`w-2 h-2 rounded-full mr-2 ${
              isRealTime ? 'bg-green-500 animate-pulse' : 'bg-gray-400'
            }`} />
            {isRealTime ? 'Live' : 'Static'}
          </button>
          
          <button
            onClick={onAutoRefreshToggle}
            className={`inline-flex items-center px-3 py-1 rounded-md text-sm font-medium transition-colors ${
              autoRefresh
                ? 'bg-blue-100 text-blue-800'
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
            }`}
          >
            <RefreshCw className={`w-4 h-4 mr-2 ${autoRefresh ? 'animate-spin' : ''}`} />
            Auto Refresh
          </button>
          
          <button
            onClick={onRefresh}
            className="inline-flex items-center px-3 py-1 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
          >
            <RefreshCw className="w-4 h-4 mr-2" />
            Refresh
          </button>
          
          <button
            onClick={onExport}
            className="inline-flex items-center px-3 py-1 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
          >
            <Download className="w-4 h-4 mr-2" />
            Export
          </button>
        </div>
      </div>
    </div>
  )
}

const PerformanceCharts: React.FC<{
  metrics: any
  timeRange: TimeRange
}> = ({ metrics, timeRange }) => {
  if (!metrics) return null
  
  // Mock time series data - in a real app, this would come from the API
  const generateMockTimeSeries = (baseValue: number, points: number = 20) => {
    const data = []
    const now = new Date()
    const interval = (timeRange.end as number - (timeRange.start as number)) / points
    
    for (let i = 0; i < points; i++) {
      const timestamp = new Date((timeRange.start as number) + i * interval).toISOString()
      const variance = (Math.random() - 0.5) * baseValue * 0.3
      data.push({
        timestamp,
        value: Math.max(0, baseValue + variance)
      })
    }
    
    return data
  }
  
  const cpuData = generateMockTimeSeries(metrics.system?.cpu?.current || 45)
  const memoryData = generateMockTimeSeries(metrics.system?.memory?.current || 65)
  const networkData = generateMockTimeSeries(metrics.system?.network?.latency?.avg || 25)
  
  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <TimeSeriesChart
        data={cpuData}
        metric={metrics.system?.cpu}
        title="CPU Usage"
        chartType="area"
        colors={['#3B82F6']}
        valueFormat={(value) => `${value.toFixed(1)}%`}
      />
      
      <TimeSeriesChart
        data={memoryData}
        metric={metrics.system?.memory}
        title="Memory Usage"
        chartType="area"
        colors={['#10B981']}
        valueFormat={(value) => `${value.toFixed(1)}%`}
      />
      
      <TimeSeriesChart
        data={networkData}
        title="Network Latency"
        chartType="line"
        colors={['#F59E0B']}
        valueFormat={(value) => `${value.toFixed(1)}ms`}
      />
      
      <MultiSeriesChart
        series={[
          { name: 'CPU', data: cpuData, color: '#3B82F6' },
          { name: 'Memory', data: memoryData, color: '#10B981' }
        ]}
        title="System Resources"
        chartType="line"
        valueFormat={(value) => `${value.toFixed(1)}%`}
      />
    </div>
  )
}

export const Monitoring: React.FC<MonitoringPageProps> = ({ className = '' }) => {
  const [activeTab, setActiveTab] = useState<'overview' | 'metrics' | 'alerts' | 'logs' | 'thresholds'>('overview')
  const [timeRange, setTimeRange] = useState<TimeRange>(getTimeRangeFromPreset('1h'))
  const [isRealTime, setIsRealTime] = useState(true)
  const [isFullscreen, setIsFullscreen] = useState(false)
  
  const {
    state,
    metrics,
    alerts,
    activeAlerts,
    criticalAlerts,
    logs,
    filteredLogs,
    thresholds,
    refreshMetrics,
    acknowledgeAlert,
    resolveAlert,
    createThreshold,
    updateThreshold,
    deleteThreshold,
    exportData,
    updateDashboardConfig,
    isConnected,
    lastUpdate,
    error
  } = useMonitoring({
    autoRefresh: true,
    refreshInterval: 30000,
    enableRealTime: isRealTime
  })
  
  // Update dashboard time range
  useEffect(() => {
    updateDashboardConfig({ timeRange })
  }, [timeRange, updateDashboardConfig])
  
  // Calculate summary statistics
  const summaryStats = useMemo(() => {
    if (!metrics) return null
    
    return {
      systemHealth: metrics.system ? Math.round((
        (100 - metrics.system.cpu.current) +
        (100 - metrics.system.memory.current) +
        (100 - metrics.system.disk.current)
      ) / 3) : 0,
      activeNodes: metrics.cluster?.activeNodes || 0,
      totalNodes: metrics.cluster?.totalNodes || 0,
      responseTime: metrics.cluster?.averageResponseTime || 0,
      errorRate: metrics.cluster?.errorRate || 0,
      throughput: metrics.cluster?.throughput || 0
    }
  }, [metrics])
  
  const handleExport = useCallback(() => {
    const exportOptions: ExportOptions = {
      format: 'json',
      timeRange,
      metrics: ['system.cpu', 'system.memory', 'system.disk', 'cluster.response_time'],
      includeAlerts: true,
      includeLogs: true
    }
    
    exportData(exportOptions)
  }, [timeRange, exportData])
  
  const handleAutoRefreshToggle = useCallback(() => {
    updateDashboardConfig({ 
      autoRefresh: !state.dashboardConfig.autoRefresh 
    })
  }, [state.dashboardConfig.autoRefresh, updateDashboardConfig])
  
  if (state.loading && !metrics) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <LoadingSpinner size="lg" tip="Loading monitoring data..." />
      </div>
    )
  }
  
  return (
    <div className={`space-y-6 ${className} ${isFullscreen ? 'fixed inset-0 z-50 bg-white overflow-auto p-6' : ''}`}>
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <h1 className="text-2xl font-bold text-gray-900">System Monitoring</h1>
          
          {lastUpdate && (
            <div className="flex items-center space-x-2 text-sm text-gray-500">
              <Clock className="w-4 h-4" />
              <span>Last updated: {format(new Date(lastUpdate), 'HH:mm:ss')}</span>
            </div>
          )}
          
          {!isConnected && (
            <div className="flex items-center space-x-2 text-sm text-red-600">
              <AlertTriangle className="w-4 h-4" />
              <span>Connection lost</span>
            </div>
          )}
        </div>
        
        <div className="flex items-center space-x-2">
          <button
            onClick={() => setIsFullscreen(!isFullscreen)}
            className="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-50 rounded-md transition-colors"
            title={isFullscreen ? 'Exit fullscreen' : 'Enter fullscreen'}
          >
            {isFullscreen ? <Minimize2 className="w-4 h-4" /> : <Maximize2 className="w-4 h-4" />}
          </button>
        </div>
      </div>
      
      {/* Error Display */}
      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <div className="flex items-center space-x-2">
            <AlertTriangle className="w-5 h-5 text-red-500" />
            <div className="text-red-800">
              <strong>Error:</strong> {error}
            </div>
          </div>
        </div>
      )}
      
      {/* Dashboard Controls */}
      <DashboardControls
        timeRange={timeRange}
        onTimeRangeChange={setTimeRange}
        isRealTime={isRealTime}
        onRealTimeToggle={() => setIsRealTime(!isRealTime)}
        onRefresh={refreshMetrics}
        onExport={handleExport}
        autoRefresh={state.dashboardConfig.autoRefresh}
        onAutoRefreshToggle={handleAutoRefreshToggle}
      />
      
      {/* Summary Cards */}
      {summaryStats && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
          <SummaryCard
            title="System Health"
            value={`${summaryStats.systemHealth}%`}
            icon={<Activity className="w-4 h-4" />}
            color={summaryStats.systemHealth > 80 ? 'green' : summaryStats.systemHealth > 60 ? 'yellow' : 'red'}
          />
          
          <SummaryCard
            title="Active Nodes"
            value={`${summaryStats.activeNodes}/${summaryStats.totalNodes}`}
            icon={<Monitor className="w-4 h-4" />}
            color={summaryStats.activeNodes === summaryStats.totalNodes ? 'green' : 'yellow'}
          />
          
          <SummaryCard
            title="Response Time"
            value={`${summaryStats.responseTime.toFixed(0)}ms`}
            icon={<Clock className="w-4 h-4" />}
            color={summaryStats.responseTime < 100 ? 'green' : summaryStats.responseTime < 500 ? 'yellow' : 'red'}
          />
          
          <SummaryCard
            title="Error Rate"
            value={`${summaryStats.errorRate.toFixed(2)}%`}
            icon={<AlertTriangle className="w-4 h-4" />}
            color={summaryStats.errorRate < 1 ? 'green' : summaryStats.errorRate < 5 ? 'yellow' : 'red'}
          />
          
          <SummaryCard
            title="Throughput"
            value={`${summaryStats.throughput.toFixed(0)} req/s`}
            icon={<TrendingUp className="w-4 h-4" />}
            color="blue"
          />
        </div>
      )}
      
      {/* Critical Alerts Banner */}
      {criticalAlerts.length > 0 && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <div className="flex items-center space-x-2">
            <AlertTriangle className="w-5 h-5 text-red-500" />
            <div className="flex-1">
              <h3 className="text-red-800 font-medium">
                {criticalAlerts.length} Critical Alert{criticalAlerts.length > 1 ? 's' : ''}
              </h3>
              <div className="text-red-700 text-sm">
                {criticalAlerts.slice(0, 3).map(alert => alert.message).join(', ')}
                {criticalAlerts.length > 3 && ` and ${criticalAlerts.length - 3} more...`}
              </div>
            </div>
            <button
              onClick={() => setActiveTab('alerts')}
              className="px-3 py-1 bg-red-600 text-white rounded-md text-sm hover:bg-red-700"
            >
              View Alerts
            </button>
          </div>
        </div>
      )}
      
      {/* Tab Navigation */}
      <div className="border-b border-gray-200">
        <nav className="flex space-x-8">
          <TabButton
            active={activeTab === 'overview'}
            onClick={() => setActiveTab('overview')}
            icon={<BarChart3 className="w-4 h-4" />}
          >
            Overview
          </TabButton>
          
          <TabButton
            active={activeTab === 'metrics'}
            onClick={() => setActiveTab('metrics')}
            icon={<Activity className="w-4 h-4" />}
          >
            Metrics
          </TabButton>
          
          <TabButton
            active={activeTab === 'alerts'}
            onClick={() => setActiveTab('alerts')}
            icon={<Bell className="w-4 h-4" />}
          >
            Alerts ({activeAlerts.length})
          </TabButton>
          
          <TabButton
            active={activeTab === 'logs'}
            onClick={() => setActiveTab('logs')}
            icon={<FileText className="w-4 h-4" />}
          >
            Logs ({filteredLogs.length})
          </TabButton>
          
          <TabButton
            active={activeTab === 'thresholds'}
            onClick={() => setActiveTab('thresholds')}
            icon={<Settings className="w-4 h-4" />}
          >
            Thresholds ({thresholds.length})
          </TabButton>
        </nav>
      </div>
      
      {/* Tab Content */}
      <div className="min-h-[600px]">
        {activeTab === 'overview' && (
          <div className="space-y-6">
            <MetricsGrid
              systemMetrics={metrics?.system}
              clusterMetrics={metrics?.cluster}
            />
            
            <PerformanceCharts
              metrics={metrics}
              timeRange={timeRange}
            />
          </div>
        )}
        
        {activeTab === 'metrics' && (
          <div className="space-y-6">
            <MetricsGrid
              systemMetrics={metrics?.system}
              clusterMetrics={metrics?.cluster}
            />
            
            <PerformanceCharts
              metrics={metrics}
              timeRange={timeRange}
            />
          </div>
        )}
        
        {activeTab === 'alerts' && (
          <AlertsPanel
            alerts={alerts}
            onAcknowledge={acknowledgeAlert}
            onResolve={resolveAlert}
            onRefresh={refreshMetrics}
          />
        )}
        
        {activeTab === 'logs' && (
          <LogViewer
            logs={logs}
            onRefresh={refreshMetrics}
            onExport={(logs) => {
              const exportOptions: ExportOptions = {
                format: 'json',
                timeRange,
                metrics: [],
                includeLogs: true
              }
              exportData(exportOptions)
            }}
            realTime={isRealTime}
            autoScroll={isRealTime}
          />
        )}
        
        {activeTab === 'thresholds' && (
          <ThresholdConfig
            thresholds={thresholds}
            onCreateThreshold={createThreshold}
            onUpdateThreshold={updateThreshold}
            onDeleteThreshold={deleteThreshold}
          />
        )}
      </div>
    </div>
  )
}