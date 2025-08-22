import React, { useEffect, useState } from 'react'
import { useWebSocket, useClusterStatus, useMetrics, useNotifications } from '../hooks/useWebSocket'
import { WebSocketStatus, WebSocketStatusHeader } from '../components/WebSocketStatus'
import { ClusterUpdateData, MetricsUpdateData } from '../types/websocket'
import { getAPIClient } from '../lib/api'
import { NodeInfo, ModelInfo, PerformanceMetrics } from '../types/api'

export function KPIWidget({ 
  title, 
  value, 
  unit, 
  isLive = false, 
  trend,
  className = '' 
}: { 
  title: string
  value: number
  unit?: string
  isLive?: boolean
  trend?: 'up' | 'down' | 'stable'
  className?: string
}) {
  const getTrendIcon = () => {
    switch (trend) {
      case 'up': return '📈'
      case 'down': return '📉'
      case 'stable': return '➡️'
      default: return ''
    }
  }

  return (
    <div className={`omx-v2 p-4 rounded border bg-white shadow transition-all duration-300 ${isLive ? 'ring-2 ring-green-200 bg-green-50' : ''} ${className}`}>
      <div className="flex items-center justify-between">
        <div className="text-sm text-slate-500">{title}</div>
        <div className="flex items-center gap-1">
          {trend && <span className="text-xs">{getTrendIcon()}</span>}
          {isLive && <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>}
        </div>
      </div>
      <div className="text-3xl font-semibold mt-1">
        {typeof value === 'number' ? value.toFixed(1) : value}{unit}
      </div>
    </div>
  )
}

export function ClusterStatusCard({ clusterData }: { clusterData?: ClusterUpdateData }) {
  if (!clusterData?.dashboard) {
    return (
      <div className="omx-v2 p-4 rounded border bg-gray-50 shadow">
        <div className="text-sm text-gray-500">Cluster Status</div>
        <div className="text-lg font-medium mt-1 text-gray-400">No data available</div>
      </div>
    )
  }

  const { clusterStatus, nodeCount, activeModels } = clusterData.dashboard

  return (
    <div className="omx-v2 p-4 rounded border bg-white shadow">
      <div className="flex items-center justify-between mb-3">
        <div className="text-sm text-slate-500">Cluster Status</div>
        <div className={`px-2 py-1 rounded-full text-xs font-medium ${
          clusterStatus.healthy 
            ? 'bg-green-100 text-green-700' 
            : 'bg-red-100 text-red-700'
        }`}>
          {clusterStatus.healthy ? '✅ Healthy' : '❌ Issues'}
        </div>
      </div>
      
      <div className="space-y-2">
        <div className="flex justify-between">
          <span className="text-sm text-gray-600">Nodes:</span>
          <span className="font-medium">{nodeCount}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-sm text-gray-600">Active Models:</span>
          <span className="font-medium">{activeModels}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-sm text-gray-600">Leader:</span>
          <span className={`font-medium ${clusterStatus.leader ? 'text-green-600' : 'text-orange-600'}`}>
            {clusterStatus.leader ? 'Yes' : 'No'}
          </span>
        </div>
        <div className="flex justify-between">
          <span className="text-sm text-gray-600">Consensus:</span>
          <span className={`font-medium ${clusterStatus.consensus ? 'text-green-600' : 'text-red-600'}`}>
            {clusterStatus.consensus ? 'Active' : 'Inactive'}
          </span>
        </div>
      </div>
    </div>
  )
}

export function NotificationPanel() {
  const { notifications, removeNotification, clearNotifications } = useNotifications()
  
  if (notifications.length === 0) return null
  
  return (
    <div className="fixed top-4 right-4 z-50 space-y-2 max-w-sm">
      <div className="flex justify-between items-center">
        <span className="text-sm font-medium text-gray-600">Notifications</span>
        <button 
          onClick={clearNotifications}
          className="text-xs text-blue-600 hover:text-blue-800"
        >
          Clear All
        </button>
      </div>
      
      {notifications.map((notification, index) => (
        <div 
          key={`${notification.timestamp}-${index}`}
          className="bg-white border border-gray-200 rounded-lg shadow-lg p-3 animate-slide-in"
        >
          <div className="flex justify-between items-start">
            <div className="flex-1">
              <div className="font-medium text-sm text-gray-900">
                {notification.notification.title}
              </div>
              <div className="text-xs text-gray-600 mt-1">
                {notification.notification.message}
              </div>
              <div className="text-xs text-gray-400 mt-1">
                {new Date(notification.timestamp).toLocaleTimeString()}
              </div>
            </div>
            <button 
              onClick={() => removeNotification(index)}
              className="text-gray-400 hover:text-gray-600 ml-2"
            >
              ✕
            </button>
          </div>
        </div>
      ))}
    </div>
  )
}

export default function Dashboard() {
  const [enabled, setEnabled] = useState(false)
  const [fallbackKpi, setFallbackKpi] = useState(0)
  const [nodes, setNodes] = useState<NodeInfo[]>([])
  const [models, setModels] = useState<ModelInfo[]>([])
  const [performanceMetrics, setPerformanceMetrics] = useState<PerformanceMetrics | null>(null)
  const [systemStatus, setSystemStatus] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  
  // WebSocket integration
  const { isConnected, connectionState, error } = useWebSocket()
  const { data: clusterData, lastUpdate: clusterLastUpdate } = useClusterStatus()
  const { data: metricsData, lastUpdate: metricsLastUpdate } = useMetrics()
  
  // API client
  const apiClient = getAPIClient()

  // Load initial data
  useEffect(() => {
    const loadData = async () => {
      setLoading(true)
      try {
        const [nodesData, modelsData, metricsData, statusData] = await Promise.allSettled([
          apiClient.cluster.getNodes(),
          apiClient.models.list(),
          apiClient.monitoring.getPerformanceMetrics(),
          apiClient.getSystemStatus()
        ])

        if (nodesData.status === 'fulfilled') setNodes(nodesData.value)
        if (modelsData.status === 'fulfilled') setModels(modelsData.value)
        if (metricsData.status === 'fulfilled') setPerformanceMetrics(metricsData.value)
        if (statusData.status === 'fulfilled') setSystemStatus(statusData.value)
      } catch (error) {
        console.error('Failed to load dashboard data:', error)
      } finally {
        setLoading(false)
      }
    }

    loadData()
  }, [])

  useEffect(() => {
    // Feature flag gate
    const flag = localStorage.getItem('V2_KPI_WIDGET') === '1' || (window as any).V2_KPI_WIDGET === true
    setEnabled(flag)

    let t: any
    if (flag && !isConnected) {
      // Fallback mock updates when WebSocket is not connected
      t = setInterval(() => {
        setFallbackKpi((v) => Math.max(0, Math.min(100, v + (Math.random() * 10 - 5))))
      }, 1000)
    }
    return () => t && clearInterval(t)
  }, [isConnected])

  // Refresh data periodically when not connected to WebSocket
  useEffect(() => {
    if (!isConnected) {
      const interval = setInterval(async () => {
        try {
          const [nodesData, metricsData] = await Promise.allSettled([
            apiClient.cluster.getNodes(),
            apiClient.monitoring.getPerformanceMetrics()
          ])

          if (nodesData.status === 'fulfilled') setNodes(nodesData.value)
          if (metricsData.status === 'fulfilled') setPerformanceMetrics(metricsData.value)
        } catch (error) {
          console.error('Failed to refresh data:', error)
        }
      }, 30000) // Refresh every 30 seconds when offline

      return () => clearInterval(interval)
    }
  }, [isConnected, apiClient])

  if (!enabled) return (
    <div className="omx-v2 min-h-screen bg-gray-50 flex items-center justify-center">
      <div className="text-center">
        <div className="text-slate-500 text-lg mb-4">Dashboard</div>
        <div className="text-sm text-slate-400 mb-4">
          Enable V2 dashboard by setting localStorage V2_KPI_WIDGET = '1'
        </div>
        <button 
          onClick={() => {
            localStorage.setItem('V2_KPI_WIDGET', '1')
            window.location.reload()
          }}
          className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors"
        >
          Enable V2 Dashboard
        </button>
      </div>
    </div>
  )

  // Calculate real-time metrics
  const healthyNodes = nodes.filter(node => node.health === 'healthy').length
  const totalNodes = nodes.length
  const clusterUtilization = isConnected && clusterData?.dashboard?.clusterStatus?.healthy 
    ? (clusterData.dashboard.nodeCount / Math.max(clusterData.dashboard.nodeCount, 1)) * 100
    : totalNodes > 0 ? (healthyNodes / totalNodes) * 100 : fallbackKpi

  const activeNodes = isConnected ? (clusterData?.dashboard?.nodeCount || 0) : totalNodes
  const activeModels = isConnected ? (clusterData?.dashboard?.activeModels || 0) : models.length
  const isClusterHealthy = isConnected 
    ? (clusterData?.dashboard?.clusterStatus?.healthy ?? false)
    : healthyNodes === totalNodes && totalNodes > 0

  const avgCpuUsage = nodes.length > 0 
    ? nodes.reduce((sum, node) => sum + node.resources.cpu_usage, 0) / nodes.length 
    : 0
    
  const avgMemoryUsage = nodes.length > 0
    ? nodes.reduce((sum, node) => sum + node.resources.memory_usage, 0) / nodes.length
    : 0

  const requestsPerSecond = performanceMetrics?.requests_per_second || 0
  const avgResponseTime = performanceMetrics?.average_response_time || 0

  return (
    <div className="omx-v2 min-h-screen bg-gray-50">
      {/* Header with WebSocket status */}
      <div className="bg-white border-b border-gray-200 px-6 py-4">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
            <p className="text-sm text-gray-600 mt-1">
              Distributed Ollama Cluster Monitoring
            </p>
          </div>
          <div className="flex items-center gap-4">
            <WebSocketStatusHeader />
            {clusterLastUpdate && (
              <div className="text-xs text-gray-500">
                Last update: {new Date(clusterLastUpdate).toLocaleTimeString()}
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Main dashboard content */}
      <div className="p-6">
        {/* Connection status warning */}
        {!isConnected && (
          <div className="mb-6 p-4 bg-yellow-50 border border-yellow-200 rounded-lg">
            <div className="flex items-center gap-2">
              <span className="text-yellow-600">⚠️</span>
              <div>
                <div className="font-medium text-yellow-800">Real-time updates unavailable</div>
                <div className="text-sm text-yellow-700">
                  WebSocket connection is {connectionState}. Showing cached or mock data.
                  {error && <span className="block">Error: {error.message}</span>}
                </div>
              </div>
            </div>
          </div>
        )}

        {/* KPI Grid */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-6 gap-4 mb-6">
          <KPIWidget 
            title="Cluster Health" 
            value={clusterUtilization} 
            unit="%" 
            isLive={isConnected}
            trend={clusterUtilization > 90 ? 'stable' : clusterUtilization > 75 ? 'up' : 'down'}
          />
          <KPIWidget 
            title="Active Nodes" 
            value={activeNodes} 
            isLive={isConnected}
          />
          <KPIWidget 
            title="Running Models" 
            value={activeModels} 
            isLive={isConnected}
          />
          <KPIWidget 
            title="CPU Usage" 
            value={avgCpuUsage} 
            unit="%" 
            isLive={isConnected}
            trend={avgCpuUsage > 80 ? 'up' : avgCpuUsage < 30 ? 'down' : 'stable'}
          />
          <KPIWidget 
            title="Memory Usage" 
            value={avgMemoryUsage} 
            unit="%" 
            isLive={isConnected}
            trend={avgMemoryUsage > 80 ? 'up' : avgMemoryUsage < 30 ? 'down' : 'stable'}
          />
          <KPIWidget 
            title="RPS" 
            value={requestsPerSecond} 
            isLive={isConnected}
            trend={requestsPerSecond > 100 ? 'up' : requestsPerSecond < 10 ? 'down' : 'stable'}
          />
        </div>

        {/* Detailed Status Cards */}
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-6 mb-6">
          <ClusterStatusCard clusterData={clusterData || undefined} />
          
          {/* Nodes Status Card */}
          <div className="omx-v2 p-4 rounded border bg-white shadow">
            <div className="text-sm text-slate-500 mb-3">Node Status</div>
            {loading ? (
              <div className="text-sm text-gray-500">Loading...</div>
            ) : (
              <div className="space-y-2">
                <div className="flex justify-between">
                  <span className="text-sm text-gray-600">Total Nodes:</span>
                  <span className="font-medium">{totalNodes}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-gray-600">Healthy:</span>
                  <span className="font-medium text-green-600">{healthyNodes}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-gray-600">Leader:</span>
                  <span className="font-medium">
                    {nodes.find(n => n.role === 'leader')?.id.slice(0, 8) || 'None'}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-gray-600">Avg CPU:</span>
                  <span className="font-medium">{avgCpuUsage.toFixed(1)}%</span>
                </div>
              </div>
            )}
          </div>
          
          {/* Models Status Card */}
          <div className="omx-v2 p-4 rounded border bg-white shadow">
            <div className="text-sm text-slate-500 mb-3">Models Overview</div>
            {loading ? (
              <div className="text-sm text-gray-500">Loading...</div>
            ) : (
              <div className="space-y-2">
                <div className="flex justify-between">
                  <span className="text-sm text-gray-600">Total Models:</span>
                  <span className="font-medium">{models.length}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-gray-600">Distributed:</span>
                  <span className="font-medium">
                    {models.filter(m => m.distribution.availability === 'full').length}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-gray-600">Available:</span>
                  <span className="font-medium text-green-600">
                    {models.filter(m => m.distribution.availability !== 'unavailable').length}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-gray-600">Avg Size:</span>
                  <span className="font-medium">
                    {models.length > 0 ? (models.reduce((sum, m) => sum + m.size, 0) / models.length / 1024 / 1024 / 1024).toFixed(1) + 'GB' : '0GB'}
                  </span>
                </div>
              </div>
            )}
          </div>
          
          {/* Performance Metrics Card */}
          <div className="omx-v2 p-4 rounded border bg-white shadow">
            <div className="text-sm text-slate-500 mb-3">Performance</div>
            {loading ? (
              <div className="text-sm text-gray-500">Loading...</div>
            ) : (
              <div className="space-y-2">
                <div className="flex justify-between">
                  <span className="text-sm text-gray-600">Requests/sec:</span>
                  <span className="font-medium">{requestsPerSecond.toFixed(1)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-gray-600">Avg Latency:</span>
                  <span className="font-medium">{avgResponseTime.toFixed(0)}ms</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-gray-600">Error Rate:</span>
                  <span className={`font-medium ${(performanceMetrics?.error_rate || 0) > 0.05 ? 'text-red-600' : 'text-green-600'}`}>
                    {((performanceMetrics?.error_rate || 0) * 100).toFixed(2)}%
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-gray-600">Connections:</span>
                  <span className="font-medium">{performanceMetrics?.active_connections || 0}</span>
                </div>
              </div>
            )}
          </div>
        </div>
        
        {/* WebSocket Status Section */}
        <div className="mb-6">
          <WebSocketStatus showDetails={true} />
        </div>

        {/* Development Info */}
        {process.env.NODE_ENV === 'development' && (
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <details>
              <summary className="font-medium text-blue-800 cursor-pointer">
                Development Debug Info
              </summary>
              <div className="mt-2 text-sm text-blue-700 space-y-1">
                <div>WebSocket State: {connectionState}</div>
                <div>Connected: {isConnected ? 'Yes' : 'No'}</div>
                <div>Loading: {loading ? 'Yes' : 'No'}</div>
                <div>Nodes Loaded: {nodes.length}</div>
                <div>Models Loaded: {models.length}</div>
                <div>Cluster Data: {clusterData ? 'Available' : 'None'}</div>
                <div>Metrics Data: {metricsData ? 'Available' : 'None'}</div>
                <div>Performance Data: {performanceMetrics ? 'Available' : 'None'}</div>
                <div>System Status: {systemStatus ? 'Available' : 'None'}</div>
                <div>Last Cluster Update: {clusterLastUpdate || 'Never'}</div>
                <div>Last Metrics Update: {metricsLastUpdate || 'Never'}</div>
              </div>
            </details>
          </div>
        )}
      </div>

      {/* Notification Panel */}
      <NotificationPanel />
    </div>
  )
}

