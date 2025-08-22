/**
 * Dashboard Page - Main overview for OllamaMax Distributed System
 * Features real-time metrics, activity feed, and system health monitoring
 */

import React, { useState, useEffect } from 'react'
import { 
  ResponsiveContainer, 
  LineChart, 
  Line, 
  AreaChart, 
  Area, 
  BarChart, 
  Bar, 
  PieChart, 
  Pie, 
  Cell,
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  Legend 
} from 'recharts'
import {
  Activity,
  AlertTriangle,
  ArrowUp,
  ArrowDown,
  Cpu,
  Database,
  Download,
  HardDrive,
  MemoryStick,
  Monitor,
  Network,
  Play,
  Pause,
  RefreshCw,
  Server,
  Users,
  Zap,
  BarChart3,
  TrendingUp,
  TrendingDown,
  Clock,
  CheckCircle,
  XCircle,
  AlertCircle,
  Globe,
  Wifi,
  WifiOff
} from 'lucide-react'

// Component imports
import { Card, CardHeader, CardContent, CardFooter, CardTitle, CardDescription } from '@/design-system/components/Card/Card'
import { Badge } from '@/design-system/components/Badge/Badge'
import { Button } from '@/design-system/components/Button/Button'
import { Alert } from '@/design-system/components/Alert/Alert'

// Hook imports
import { useWebSocket, useClusterStatus, useMetrics, useNotifications } from '@/hooks/useWebSocket'
import { useTheme } from '@/theme/hooks/useTheme'

// Dashboard component imports
import { 
  MetricCard,
  SystemHealthCard,
  ActivityFeedCard,
  QuickActionsCard,
  AlertsCard
} from './components'
import ExportUtils from './components/ExportUtils'

// Type imports
import type { 
  ClusterInfo, 
  NodeInfo, 
  ModelInfo, 
  TaskInfo, 
  ClusterMetrics,
  ResourceUsage 
} from '@/types/api'

// Dashboard data interfaces
interface DashboardMetrics {
  nodes: {
    total: number
    healthy: number
    degraded: number
    offline: number
  }
  models: {
    total: number
    synced: number
    syncing: number
    failed: number
  }
  tasks: {
    total: number
    running: number
    completed: number
    failed: number
    pending: number
  }
  performance: {
    cpu: number
    memory: number
    disk: number
    network: number
    avgResponseTime: number
    throughput: number
  }
}

interface ActivityItem {
  id: string
  type: 'node' | 'model' | 'task' | 'alert' | 'system'
  title: string
  description: string
  timestamp: string
  severity: 'info' | 'warning' | 'error' | 'success'
  metadata?: Record<string, any>
}

interface SystemAlert {
  id: string
  type: 'error' | 'warning' | 'info'
  title: string
  message: string
  timestamp: string
  acknowledged: boolean
  source: string
}

const Dashboard: React.FC = () => {
  const { theme } = useTheme()
  const { isConnected, connectionState } = useWebSocket()
  const { data: clusterStatus } = useClusterStatus()
  const { data: metrics } = useMetrics()
  const { notifications } = useNotifications()
  
  // State management
  const [refreshInterval, setRefreshInterval] = useState(5000)
  const [autoRefresh, setAutoRefresh] = useState(true)
  const [timeRange, setTimeRange] = useState('1h')
  const [selectedView, setSelectedView] = useState('overview')
  
  // Mock data - replace with real API calls
  const [dashboardMetrics, setDashboardMetrics] = useState<DashboardMetrics>({
    nodes: { total: 5, healthy: 4, degraded: 1, offline: 0 },
    models: { total: 12, synced: 10, syncing: 1, failed: 1 },
    tasks: { total: 156, running: 8, completed: 142, failed: 3, pending: 3 },
    performance: { cpu: 68, memory: 74, disk: 45, network: 32, avgResponseTime: 125, throughput: 450 }
  })
  
  const [activityFeed, setActivityFeed] = useState<ActivityItem[]>([
    {
      id: '1',
      type: 'node',
      title: 'Node joined cluster',
      description: 'worker-node-03 successfully joined the cluster',
      timestamp: new Date(Date.now() - 5 * 60 * 1000).toISOString(),
      severity: 'success'
    },
    {
      id: '2',
      type: 'model',
      title: 'Model sync completed',
      description: 'llama3-8b synchronized across 4 nodes',
      timestamp: new Date(Date.now() - 12 * 60 * 1000).toISOString(),
      severity: 'success'
    },
    {
      id: '3',
      type: 'alert',
      title: 'High memory usage',
      description: 'Node worker-node-01 memory usage at 89%',
      timestamp: new Date(Date.now() - 18 * 60 * 1000).toISOString(),
      severity: 'warning'
    },
    {
      id: '4',
      type: 'task',
      title: 'Inference task completed',
      description: 'Task #1247 completed in 2.3s',
      timestamp: new Date(Date.now() - 25 * 60 * 1000).toISOString(),
      severity: 'info'
    }
  ])
  
  const [systemAlerts, setSystemAlerts] = useState<SystemAlert[]>([
    {
      id: '1',
      type: 'warning',
      title: 'High Memory Usage',
      message: 'Node worker-node-01 memory usage is at 89%. Consider adding more nodes.',
      timestamp: new Date(Date.now() - 18 * 60 * 1000).toISOString(),
      acknowledged: false,
      source: 'monitoring'
    },
    {
      id: '2',
      type: 'error',
      title: 'Model Sync Failed',
      message: 'Failed to sync llama3-70b to worker-node-04. Retrying...',
      timestamp: new Date(Date.now() - 35 * 60 * 1000).toISOString(),
      acknowledged: false,
      source: 'sync'
    }
  ])
  
  // Performance chart data
  const performanceData = [
    { time: '00:00', cpu: 45, memory: 67, network: 23, tasks: 5 },
    { time: '00:05', cpu: 52, memory: 71, network: 28, tasks: 7 },
    { time: '00:10', cpu: 48, memory: 69, network: 31, tasks: 6 },
    { time: '00:15', cpu: 68, memory: 74, network: 32, tasks: 8 },
    { time: '00:20', cpu: 65, memory: 76, network: 29, tasks: 9 },
    { time: '00:25', cpu: 71, memory: 78, network: 35, tasks: 12 }
  ]
  
  // Node distribution data
  const nodeDistributionData = [
    { name: 'Healthy', value: dashboardMetrics.nodes.healthy, color: '#22c55e' },
    { name: 'Degraded', value: dashboardMetrics.nodes.degraded, color: '#f59e0b' },
    { name: 'Offline', value: dashboardMetrics.nodes.offline, color: '#ef4444' }
  ]
  
  // Task throughput data
  const throughputData = [
    { time: '00:00', completed: 12, failed: 1 },
    { time: '00:05', completed: 15, failed: 0 },
    { time: '00:10', completed: 18, failed: 2 },
    { time: '00:15', completed: 22, failed: 1 },
    { time: '00:20', completed: 28, failed: 0 },
    { time: '00:25', completed: 31, failed: 1 }
  ]
  
  // Auto-refresh effect
  useEffect(() => {
    if (!autoRefresh) return
    
    const interval = setInterval(() => {
      // Simulate real-time updates
      setDashboardMetrics(prev => ({
        ...prev,
        performance: {
          ...prev.performance,
          cpu: Math.max(0, Math.min(100, prev.performance.cpu + (Math.random() - 0.5) * 10)),
          memory: Math.max(0, Math.min(100, prev.performance.memory + (Math.random() - 0.5) * 5)),
          network: Math.max(0, Math.min(100, prev.performance.network + (Math.random() - 0.5) * 15))
        }
      }))
    }, refreshInterval)
    
    return () => clearInterval(interval)
  }, [autoRefresh, refreshInterval])
  
  // Helper functions
  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp)
    return new Intl.RelativeTimeFormat('en', { numeric: 'auto' }).format(
      Math.round((date.getTime() - Date.now()) / (1000 * 60)),
      'minute'
    )
  }
  
  const getStatusColor = (value: number, thresholds = { good: 70, warning: 85 }) => {
    if (value <= thresholds.good) return 'text-success-600'
    if (value <= thresholds.warning) return 'text-warning-600'
    return 'text-error-600'
  }
  
  const getTrendIcon = (current: number, previous: number) => {
    if (current > previous) return <ArrowUp className="h-4 w-4 text-success-600" />
    if (current < previous) return <ArrowDown className="h-4 w-4 text-error-600" />
    return <div className="h-4 w-4" />
  }
  
  const acknowledgeAlert = (alertId: string) => {
    setSystemAlerts(prev => 
      prev.map(alert => 
        alert.id === alertId ? { ...alert, acknowledged: true } : alert
      )
    )
  }
  
  return (
    <div className="min-h-screen bg-background p-4 md:p-6 space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div>
          <h1 className="text-3xl font-bold text-foreground">
            OllamaMax Dashboard
          </h1>
          <p className="text-muted-foreground mt-1">
            Distributed AI Platform Overview
          </p>
        </div>
        
        <div className="flex items-center gap-3">
          {/* Connection Status */}
          <div className="flex items-center gap-2">
            {isConnected ? (
              <>
                <Wifi className="h-4 w-4 text-success-600" />
                <span className="text-sm text-success-600">Connected</span>
              </>
            ) : (
              <>
                <WifiOff className="h-4 w-4 text-error-600" />
                <span className="text-sm text-error-600">Disconnected</span>
              </>
            )}
          </div>
          
          {/* Auto-refresh toggle */}
          <Button
            variant={autoRefresh ? 'default' : 'secondary'}
            size="sm"
            onClick={() => setAutoRefresh(!autoRefresh)}
          >
            <RefreshCw className={`h-4 w-4 mr-2 ${autoRefresh ? 'animate-spin' : ''}`} />
            Auto-refresh
          </Button>
          
          {/* Time range selector */}
          <select 
            value={timeRange} 
            onChange={(e) => setTimeRange(e.target.value)}
            className="px-3 py-2 border border-border rounded-md bg-background text-foreground text-sm"
          >
            <option value="15m">15 minutes</option>
            <option value="1h">1 hour</option>
            <option value="6h">6 hours</option>
            <option value="24h">24 hours</option>
          </select>
          
          {/* Export button */}
          <ExportUtils
            data={{
              metrics: dashboardMetrics,
              activities: activityFeed,
              alerts: systemAlerts,
              performanceData: performanceData,
              timestamp: new Date().toISOString()
            }}
            filename="ollama-dashboard"
            className="hidden sm:flex"
          />
        </div>
      </div>
      
      {/* Key Metrics Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <MetricCard
          title="Active Nodes"
          value={dashboardMetrics.nodes.healthy}
          total={dashboardMetrics.nodes.total}
          icon={<Server className="h-5 w-5" />}
          trend={getTrendIcon(dashboardMetrics.nodes.healthy, 3)}
          status="healthy"
          subtitle={`${dashboardMetrics.nodes.degraded} degraded`}
        />
        
        <MetricCard
          title="Models"
          value={dashboardMetrics.models.synced}
          total={dashboardMetrics.models.total}
          icon={<Database className="h-5 w-5" />}
          trend={getTrendIcon(dashboardMetrics.models.synced, 9)}
          status="healthy"
          subtitle={`${dashboardMetrics.models.syncing} syncing`}
        />
        
        <MetricCard
          title="Active Tasks"
          value={dashboardMetrics.tasks.running}
          total={dashboardMetrics.tasks.total}
          icon={<Play className="h-5 w-5" />}
          trend={getTrendIcon(dashboardMetrics.tasks.running, 6)}
          status="warning"
          subtitle={`${dashboardMetrics.tasks.pending} pending`}
        />
        
        <MetricCard
          title="Avg Response Time"
          value={`${dashboardMetrics.performance.avgResponseTime}ms`}
          icon={<Clock className="h-5 w-5" />}
          trend={getTrendIcon(125, 140)}
          status="healthy"
          subtitle="Last 5 minutes"
        />
      </div>
      
      {/* Main Content Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Resource Utilization Charts */}
        <div className="lg:col-span-2 space-y-6">
          {/* System Performance Chart */}
          <Card>
            <CardHeader>
              <div className="flex justify-between items-center">
                <div>
                  <CardTitle>System Performance</CardTitle>
                  <CardDescription>Real-time resource utilization</CardDescription>
                </div>
                <div className="flex gap-2">
                  <Badge variant={dashboardMetrics.performance.cpu > 80 ? 'destructive' : 'secondary'}>
                    CPU: {dashboardMetrics.performance.cpu}%
                  </Badge>
                  <Badge variant={dashboardMetrics.performance.memory > 80 ? 'destructive' : 'secondary'}>
                    Memory: {dashboardMetrics.performance.memory}%
                  </Badge>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <div className="h-80">
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={performanceData}>
                    <CartesianGrid strokeDasharray="3 3" stroke={theme === 'dark' ? '#374151' : '#e5e7eb'} />
                    <XAxis 
                      dataKey="time" 
                      stroke={theme === 'dark' ? '#9ca3af' : '#6b7280'}
                      fontSize={12}
                    />
                    <YAxis 
                      stroke={theme === 'dark' ? '#9ca3af' : '#6b7280'}
                      fontSize={12}
                    />
                    <Tooltip 
                      contentStyle={{
                        backgroundColor: theme === 'dark' ? '#1f2937' : '#ffffff',
                        border: `1px solid ${theme === 'dark' ? '#374151' : '#e5e7eb'}`,
                        borderRadius: '8px'
                      }}
                    />
                    <Legend />
                    <Line 
                      type="monotone" 
                      dataKey="cpu" 
                      stroke="#3b82f6" 
                      strokeWidth={2}
                      name="CPU %"
                    />
                    <Line 
                      type="monotone" 
                      dataKey="memory" 
                      stroke="#10b981" 
                      strokeWidth={2}
                      name="Memory %"
                    />
                    <Line 
                      type="monotone" 
                      dataKey="network" 
                      stroke="#f59e0b" 
                      strokeWidth={2}
                      name="Network %"
                    />
                  </LineChart>
                </ResponsiveContainer>
              </div>
            </CardContent>
          </Card>
          
          {/* Task Throughput Chart */}
          <Card>
            <CardHeader>
              <CardTitle>Task Throughput</CardTitle>
              <CardDescription>Completed vs failed tasks over time</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="h-64">
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={throughputData}>
                    <CartesianGrid strokeDasharray="3 3" stroke={theme === 'dark' ? '#374151' : '#e5e7eb'} />
                    <XAxis 
                      dataKey="time" 
                      stroke={theme === 'dark' ? '#9ca3af' : '#6b7280'}
                      fontSize={12}
                    />
                    <YAxis 
                      stroke={theme === 'dark' ? '#9ca3af' : '#6b7280'}
                      fontSize={12}
                    />
                    <Tooltip 
                      contentStyle={{
                        backgroundColor: theme === 'dark' ? '#1f2937' : '#ffffff',
                        border: `1px solid ${theme === 'dark' ? '#374151' : '#e5e7eb'}`,
                        borderRadius: '8px'
                      }}
                    />
                    <Area 
                      type="monotone" 
                      dataKey="completed" 
                      stackId="1" 
                      stroke="#22c55e" 
                      fill="#22c55e" 
                      fillOpacity={0.6}
                      name="Completed"
                    />
                    <Area 
                      type="monotone" 
                      dataKey="failed" 
                      stackId="1" 
                      stroke="#ef4444" 
                      fill="#ef4444" 
                      fillOpacity={0.6}
                      name="Failed"
                    />
                  </AreaChart>
                </ResponsiveContainer>
              </div>
            </CardContent>
          </Card>
        </div>
        
        {/* Sidebar */}
        <div className="space-y-6">
          {/* System Health Overview */}
          <SystemHealthCard 
            clusterStatus={clusterStatus}
            metrics={dashboardMetrics}
          />
          
          {/* Node Distribution */}
          <Card>
            <CardHeader>
              <CardTitle>Node Distribution</CardTitle>
              <CardDescription>Cluster node status breakdown</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="h-48">
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie
                      data={nodeDistributionData}
                      cx="50%"
                      cy="50%"
                      outerRadius={60}
                      dataKey="value"
                      label={({ name, value }) => `${name}: ${value}`}
                    >
                      {nodeDistributionData.map((entry, index) => (
                        <Cell key={`cell-${index}`} fill={entry.color} />
                      ))}
                    </Pie>
                    <Tooltip />
                  </PieChart>
                </ResponsiveContainer>
              </div>
              
              <div className="mt-4 space-y-2">
                {nodeDistributionData.map((item, index) => (
                  <div key={index} className="flex items-center justify-between text-sm">
                    <div className="flex items-center gap-2">
                      <div 
                        className="w-3 h-3 rounded-full" 
                        style={{ backgroundColor: item.color }}
                      />
                      <span>{item.name}</span>
                    </div>
                    <span className="font-medium">{item.value}</span>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
          
          {/* Quick Actions */}
          <QuickActionsCard />
        </div>
      </div>
      
      {/* Bottom Section */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Activity Feed */}
        <ActivityFeedCard 
          activities={activityFeed}
          onRefresh={() => setActivityFeed([...activityFeed])}
        />
        
        {/* Recent Alerts */}
        <AlertsCard 
          alerts={systemAlerts}
          onAcknowledge={acknowledgeAlert}
        />
      </div>
    </div>
  )
}

export default Dashboard