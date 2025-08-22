import React, { useState, useEffect } from 'react'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Select } from '@/components/ui/Select'
import { Input } from '@/components/ui/Input'
import { useWebSocket } from '@/hooks/useWebSocket'
import {
  BarChart, LineChart, PieChart, Area, Bar, Line, Pie, 
  CartesianGrid, XAxis, YAxis, Tooltip, Legend, ResponsiveContainer
} from 'recharts'
import { 
  TrendingUp, Users, Activity, Server, Cpu, Clock,
  FileText, Download, Calendar, Filter, Settings, Eye
} from 'lucide-react'

interface AnalyticsMetrics {
  users: {
    total: number
    active: number
    new: number
    retention: number
  }
  usage: {
    apiCalls: number
    dataProcessed: number
    modelInferences: number
    averageLatency: number
  }
  performance: {
    uptime: number
    responseTime: number
    errorRate: number
    throughput: number
  }
  resources: {
    cpuUsage: number
    memoryUsage: number
    storageUsage: number
    networkBandwidth: number
  }
  costs: {
    compute: number
    storage: number
    network: number
    total: number
  }
}

interface TimeSeriesData {
  timestamp: string
  value: number
  metric: string
}

interface Report {
  id: string
  name: string
  type: 'usage' | 'performance' | 'compliance' | 'executive'
  frequency: 'daily' | 'weekly' | 'monthly' | 'quarterly'
  recipients: string[]
  lastGenerated: string
  status: 'scheduled' | 'generating' | 'ready' | 'failed'
}

export const AnalyticsDashboard: React.FC = () => {
  const [timeRange, setTimeRange] = useState('7d')
  const [metrics, setMetrics] = useState<AnalyticsMetrics>({
    users: { total: 0, active: 0, new: 0, retention: 0 },
    usage: { apiCalls: 0, dataProcessed: 0, modelInferences: 0, averageLatency: 0 },
    performance: { uptime: 0, responseTime: 0, errorRate: 0, throughput: 0 },
    resources: { cpuUsage: 0, memoryUsage: 0, storageUsage: 0, networkBandwidth: 0 },
    costs: { compute: 0, storage: 0, network: 0, total: 0 }
  })
  const [timeSeriesData, setTimeSeriesData] = useState<TimeSeriesData[]>([])
  const [reports, setReports] = useState<Report[]>([])
  const [selectedMetric, setSelectedMetric] = useState('apiCalls')
  const [exportFormat, setExportFormat] = useState('csv')

  const { sendMessage, lastMessage } = useWebSocket('/ws/analytics')

  useEffect(() => {
    if (lastMessage) {
      const data = JSON.parse(lastMessage.data)
      if (data.type === 'metrics') {
        setMetrics(data.metrics)
      } else if (data.type === 'timeseries') {
        setTimeSeriesData(data.data)
      } else if (data.type === 'reports') {
        setReports(data.reports)
      }
    }
  }, [lastMessage])

  useEffect(() => {
    sendMessage({ action: 'subscribe', timeRange })
    return () => sendMessage({ action: 'unsubscribe' })
  }, [timeRange])

  const generateReport = (type: string) => {
    sendMessage({ action: 'generate_report', type, timeRange })
  }

  const exportData = () => {
    const data = {
      metrics,
      timeSeriesData,
      timestamp: new Date().toISOString(),
      format: exportFormat
    }
    
    if (exportFormat === 'csv') {
      const csv = convertToCSV(data)
      downloadFile(csv, 'analytics-export.csv', 'text/csv')
    } else if (exportFormat === 'json') {
      downloadFile(JSON.stringify(data, null, 2), 'analytics-export.json', 'application/json')
    } else {
      generatePDFReport(data)
    }
  }

  const convertToCSV = (data: any) => {
    const rows = []
    rows.push(['Metric', 'Value', 'Timestamp'])
    
    Object.entries(metrics).forEach(([category, values]) => {
      Object.entries(values as any).forEach(([key, value]) => {
        rows.push([`${category}.${key}`, value, new Date().toISOString()])
      })
    })
    
    return rows.map(row => row.join(',')).join('\n')
  }

  const downloadFile = (content: string, filename: string, type: string) => {
    const blob = new Blob([content], { type })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    a.click()
    URL.revokeObjectURL(url)
  }

  const generatePDFReport = async (data: any) => {
    sendMessage({ action: 'generate_pdf', data })
  }

  const formatNumber = (num: number) => {
    if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`
    if (num >= 1000) return `${(num / 1000).toFixed(1)}K`
    return num.toFixed(0)
  }

  const formatBytes = (bytes: number) => {
    if (bytes >= 1073741824) return `${(bytes / 1073741824).toFixed(2)} GB`
    if (bytes >= 1048576) return `${(bytes / 1048576).toFixed(2)} MB`
    if (bytes >= 1024) return `${(bytes / 1024).toFixed(2)} KB`
    return `${bytes} B`
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-bold">Analytics & Reporting</h1>
        <div className="flex gap-4">
          <Select value={timeRange} onChange={setTimeRange}>
            <option value="1h">Last Hour</option>
            <option value="24h">Last 24 Hours</option>
            <option value="7d">Last 7 Days</option>
            <option value="30d">Last 30 Days</option>
            <option value="90d">Last 90 Days</option>
          </Select>
          <Button onClick={exportData}>
            <Download className="w-4 h-4 mr-2" />
            Export
          </Button>
        </div>
      </div>

      {/* Key Metrics Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600">Active Users</p>
              <p className="text-2xl font-bold">{formatNumber(metrics.users.active)}</p>
              <p className="text-xs text-green-600">+{metrics.users.new} new</p>
            </div>
            <Users className="w-8 h-8 text-blue-500" />
          </div>
        </Card>

        <Card>
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600">API Calls</p>
              <p className="text-2xl font-bold">{formatNumber(metrics.usage.apiCalls)}</p>
              <p className="text-xs text-gray-500">{metrics.usage.averageLatency}ms avg</p>
            </div>
            <Activity className="w-8 h-8 text-green-500" />
          </div>
        </Card>

        <Card>
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600">System Uptime</p>
              <p className="text-2xl font-bold">{metrics.performance.uptime.toFixed(2)}%</p>
              <p className="text-xs text-red-600">{metrics.performance.errorRate.toFixed(2)}% errors</p>
            </div>
            <Server className="w-8 h-8 text-purple-500" />
          </div>
        </Card>

        <Card>
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600">Total Cost</p>
              <p className="text-2xl font-bold">${metrics.costs.total.toFixed(2)}</p>
              <p className="text-xs text-gray-500">This period</p>
            </div>
            <TrendingUp className="w-8 h-8 text-orange-500" />
          </div>
        </Card>
      </div>

      {/* Usage Trends Chart */}
      <Card>
        <div className="p-4">
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-xl font-semibold">Usage Trends</h2>
            <Select value={selectedMetric} onChange={setSelectedMetric}>
              <option value="apiCalls">API Calls</option>
              <option value="dataProcessed">Data Processed</option>
              <option value="modelInferences">Model Inferences</option>
              <option value="responseTime">Response Time</option>
            </Select>
          </div>
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={timeSeriesData.filter(d => d.metric === selectedMetric)}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="timestamp" />
              <YAxis />
              <Tooltip />
              <Legend />
              <Line type="monotone" dataKey="value" stroke="#2563eb" />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </Card>

      {/* Resource Usage */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <div className="p-4">
            <h2 className="text-xl font-semibold mb-4">Resource Utilization</h2>
            <div className="space-y-4">
              <div>
                <div className="flex justify-between mb-1">
                  <span className="text-sm">CPU Usage</span>
                  <span className="text-sm font-medium">{metrics.resources.cpuUsage}%</span>
                </div>
                <div className="w-full bg-gray-200 rounded-full h-2">
                  <div 
                    className="bg-blue-600 h-2 rounded-full"
                    style={{ width: `${metrics.resources.cpuUsage}%` }}
                  />
                </div>
              </div>

              <div>
                <div className="flex justify-between mb-1">
                  <span className="text-sm">Memory Usage</span>
                  <span className="text-sm font-medium">{metrics.resources.memoryUsage}%</span>
                </div>
                <div className="w-full bg-gray-200 rounded-full h-2">
                  <div 
                    className="bg-green-600 h-2 rounded-full"
                    style={{ width: `${metrics.resources.memoryUsage}%` }}
                  />
                </div>
              </div>

              <div>
                <div className="flex justify-between mb-1">
                  <span className="text-sm">Storage Usage</span>
                  <span className="text-sm font-medium">{metrics.resources.storageUsage}%</span>
                </div>
                <div className="w-full bg-gray-200 rounded-full h-2">
                  <div 
                    className="bg-purple-600 h-2 rounded-full"
                    style={{ width: `${metrics.resources.storageUsage}%` }}
                  />
                </div>
              </div>

              <div>
                <div className="flex justify-between mb-1">
                  <span className="text-sm">Network Bandwidth</span>
                  <span className="text-sm font-medium">{formatBytes(metrics.resources.networkBandwidth)}/s</span>
                </div>
                <div className="w-full bg-gray-200 rounded-full h-2">
                  <div 
                    className="bg-orange-600 h-2 rounded-full"
                    style={{ width: `${Math.min(metrics.resources.networkBandwidth / 1000000000 * 100, 100)}%` }}
                  />
                </div>
              </div>
            </div>
          </div>
        </Card>

        <Card>
          <div className="p-4">
            <h2 className="text-xl font-semibold mb-4">Cost Breakdown</h2>
            <ResponsiveContainer width="100%" height={250}>
              <PieChart>
                <Pie
                  data={[
                    { name: 'Compute', value: metrics.costs.compute },
                    { name: 'Storage', value: metrics.costs.storage },
                    { name: 'Network', value: metrics.costs.network }
                  ]}
                  cx="50%"
                  cy="50%"
                  labelLine={false}
                  label={(entry) => `${entry.name}: $${entry.value.toFixed(2)}`}
                  outerRadius={80}
                  fill="#8884d8"
                  dataKey="value"
                >
                  {[
                    { fill: '#2563eb' },
                    { fill: '#10b981' },
                    { fill: '#f59e0b' }
                  ].map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={entry.fill} />
                  ))}
                </Pie>
                <Tooltip />
              </PieChart>
            </ResponsiveContainer>
          </div>
        </Card>
      </div>

      {/* Scheduled Reports */}
      <Card>
        <div className="p-4">
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-xl font-semibold">Scheduled Reports</h2>
            <Button variant="outline" size="sm">
              <Settings className="w-4 h-4 mr-2" />
              Configure
            </Button>
          </div>
          <div className="space-y-3">
            {reports.map(report => (
              <div key={report.id} className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                <div className="flex items-center gap-3">
                  <FileText className="w-5 h-5 text-gray-600" />
                  <div>
                    <p className="font-medium">{report.name}</p>
                    <p className="text-sm text-gray-600">
                      {report.type} • {report.frequency} • {report.recipients.length} recipients
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  {report.status === 'ready' && (
                    <Button variant="ghost" size="sm">
                      <Download className="w-4 h-4" />
                    </Button>
                  )}
                  <span className={`px-2 py-1 text-xs rounded-full ${
                    report.status === 'ready' ? 'bg-green-100 text-green-700' :
                    report.status === 'generating' ? 'bg-blue-100 text-blue-700' :
                    report.status === 'failed' ? 'bg-red-100 text-red-700' :
                    'bg-gray-100 text-gray-700'
                  }`}>
                    {report.status}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </div>
      </Card>

      {/* Quick Actions */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Button 
          variant="outline" 
          className="h-auto py-4"
          onClick={() => generateReport('usage')}
        >
          <div className="text-center">
            <Activity className="w-8 h-8 mx-auto mb-2" />
            <p className="font-medium">Usage Report</p>
            <p className="text-xs text-gray-600">Generate detailed usage analytics</p>
          </div>
        </Button>

        <Button 
          variant="outline" 
          className="h-auto py-4"
          onClick={() => generateReport('compliance')}
        >
          <div className="text-center">
            <FileText className="w-8 h-8 mx-auto mb-2" />
            <p className="font-medium">Compliance Report</p>
            <p className="text-xs text-gray-600">GDPR, SOC2, HIPAA compliance</p>
          </div>
        </Button>

        <Button 
          variant="outline" 
          className="h-auto py-4"
          onClick={() => generateReport('executive')}
        >
          <div className="text-center">
            <Eye className="w-8 h-8 mx-auto mb-2" />
            <p className="font-medium">Executive Summary</p>
            <p className="text-xs text-gray-600">High-level insights and KPIs</p>
          </div>
        </Button>
      </div>
    </div>
  )
}

const Cell = ({ fill }: { fill: string }) => null