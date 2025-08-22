import React, { useState, useEffect } from 'react'
import { Card, CardHeader, CardContent, CardTitle } from '@/design-system/components/Card/Card'
import { Progress } from '@/design-system/components/Progress/Progress'
import type { Node } from '../NodesPage'

interface ResourceMonitorProps {
  node: Node
}

interface ResourceHistory {
  timestamp: string
  cpu: number
  memory: number
  gpu?: number
  disk: number
  network_rx: number
  network_tx: number
}

export const ResourceMonitor: React.FC<ResourceMonitorProps> = ({ node }) => {
  const [history, setHistory] = useState<ResourceHistory[]>([])
  const [timeRange, setTimeRange] = useState<'1h' | '6h' | '24h'>('1h')

  // Simulate resource history data
  useEffect(() => {
    const generateHistory = () => {
      const now = Date.now()
      const points = timeRange === '1h' ? 60 : timeRange === '6h' ? 72 : 144
      const interval = timeRange === '1h' ? 60000 : timeRange === '6h' ? 300000 : 600000

      const newHistory: ResourceHistory[] = []
      
      for (let i = points; i >= 0; i--) {
        const timestamp = new Date(now - i * interval).toISOString()
        
        // Generate realistic fluctuating data around current values
        const cpuBase = node.resources.cpu_usage
        const memoryBase = node.resources.memory_usage
        const diskBase = node.resources.disk_usage
        const gpuBase = node.resources.gpu_usage || 0
        
        newHistory.push({
          timestamp,
          cpu: Math.max(0, Math.min(100, cpuBase + (Math.random() - 0.5) * 20)),
          memory: Math.max(0, Math.min(100, memoryBase + (Math.random() - 0.5) * 10)),
          gpu: node.capabilities.gpu_enabled ? Math.max(0, Math.min(100, gpuBase + (Math.random() - 0.5) * 30)) : undefined,
          disk: Math.max(0, Math.min(100, diskBase + (Math.random() - 0.5) * 5)),
          network_rx: Math.random() * 1000000,
          network_tx: Math.random() * 2000000
        })
      }
      
      setHistory(newHistory)
    }

    generateHistory()
    const interval = setInterval(generateHistory, 30000) // Update every 30 seconds

    return () => clearInterval(interval)
  }, [node, timeRange])

  const formatBytes = (bytes: number) => {
    const units = ['B', 'KB', 'MB', 'GB']
    let size = bytes
    let unitIndex = 0
    
    while (size >= 1024 && unitIndex < units.length - 1) {
      size /= 1024
      unitIndex++
    }
    
    return `${size.toFixed(1)} ${units[unitIndex]}`
  }

  const formatBytesPerSecond = (bytes: number) => {
    return `${formatBytes(bytes)}/s`
  }

  const getResourceStatus = (value: number) => {
    if (value > 90) return { variant: 'error' as const, status: 'Critical' }
    if (value > 80) return { variant: 'warning' as const, status: 'High' }
    if (value > 60) return { variant: 'info' as const, status: 'Medium' }
    return { variant: 'success' as const, status: 'Normal' }
  }

  const latestData = history[history.length - 1]
  const networkRxRate = latestData ? latestData.network_rx : 0
  const networkTxRate = latestData ? latestData.network_tx : 0

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            <svg className="w-5 h-5 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v4a2 2 0 00-2 2v6a2 2 0 00-2 2zm-2-4h2m-2-8h2m-2 4h2" />
            </svg>
            Resource Monitor
          </CardTitle>
          <select
            value={timeRange}
            onChange={(e) => setTimeRange(e.target.value as any)}
            className="text-xs px-2 py-1 border border-border rounded bg-background text-foreground"
          >
            <option value="1h">1 Hour</option>
            <option value="6h">6 Hours</option>
            <option value="24h">24 Hours</option>
          </select>
        </div>
      </CardHeader>

      <CardContent spacing="md">
        {/* Current Resource Usage */}
        <div className="space-y-4">
          {/* CPU Usage */}
          <div>
            <div className="flex items-center justify-between mb-2">
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium text-foreground">CPU Usage</span>
                <span className={`text-xs px-2 py-1 rounded ${
                  getResourceStatus(node.resources.cpu_usage).variant === 'error' ? 'bg-error/10 text-error' :
                  getResourceStatus(node.resources.cpu_usage).variant === 'warning' ? 'bg-warning/10 text-warning' :
                  getResourceStatus(node.resources.cpu_usage).variant === 'info' ? 'bg-info/10 text-info' :
                  'bg-success/10 text-success'
                }`}>
                  {getResourceStatus(node.resources.cpu_usage).status}
                </span>
              </div>
              <span className="text-sm font-bold text-foreground">{node.resources.cpu_usage.toFixed(1)}%</span>
            </div>
            <Progress
              value={node.resources.cpu_usage}
              variant={getResourceStatus(node.resources.cpu_usage).variant}
              size="md"
            />
            <div className="flex justify-between text-xs text-muted-foreground mt-1">
              <span>{node.capabilities.cpu_cores} cores available</span>
              <span>~{(node.resources.cpu_usage * node.capabilities.cpu_cores / 100).toFixed(1)} cores used</span>
            </div>
          </div>

          {/* Memory Usage */}
          <div>
            <div className="flex items-center justify-between mb-2">
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium text-foreground">Memory Usage</span>
                <span className={`text-xs px-2 py-1 rounded ${
                  getResourceStatus(node.resources.memory_usage).variant === 'error' ? 'bg-error/10 text-error' :
                  getResourceStatus(node.resources.memory_usage).variant === 'warning' ? 'bg-warning/10 text-warning' :
                  getResourceStatus(node.resources.memory_usage).variant === 'info' ? 'bg-info/10 text-info' :
                  'bg-success/10 text-success'
                }`}>
                  {getResourceStatus(node.resources.memory_usage).status}
                </span>
              </div>
              <span className="text-sm font-bold text-foreground">{node.resources.memory_usage.toFixed(1)}%</span>
            </div>
            <Progress
              value={node.resources.memory_usage}
              variant={getResourceStatus(node.resources.memory_usage).variant}
              size="md"
            />
            <div className="flex justify-between text-xs text-muted-foreground mt-1">
              <span>{formatBytes(node.capabilities.ram_total * 1024 * 1024)} total</span>
              <span>{formatBytes(node.capabilities.ram_total * 1024 * 1024 * node.resources.memory_usage / 100)} used</span>
            </div>
          </div>

          {/* GPU Usage (if available) */}
          {node.capabilities.gpu_enabled && node.resources.gpu_usage !== undefined && (
            <div>
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium text-foreground">GPU Usage</span>
                  <span className={`text-xs px-2 py-1 rounded ${
                    getResourceStatus(node.resources.gpu_usage).variant === 'error' ? 'bg-error/10 text-error' :
                    getResourceStatus(node.resources.gpu_usage).variant === 'warning' ? 'bg-warning/10 text-warning' :
                    getResourceStatus(node.resources.gpu_usage).variant === 'info' ? 'bg-info/10 text-info' :
                    'bg-success/10 text-success'
                  }`}>
                    {getResourceStatus(node.resources.gpu_usage).status}
                  </span>
                </div>
                <span className="text-sm font-bold text-foreground">{node.resources.gpu_usage.toFixed(1)}%</span>
              </div>
              <Progress
                value={node.resources.gpu_usage}
                variant={getResourceStatus(node.resources.gpu_usage).variant}
                size="md"
              />
              <div className="flex justify-between text-xs text-muted-foreground mt-1">
                <span>{formatBytes(node.capabilities.gpu_memory * 1024 * 1024)} GPU memory</span>
                <span>{formatBytes(node.capabilities.gpu_memory * 1024 * 1024 * node.resources.gpu_usage / 100)} used</span>
              </div>
            </div>
          )}

          {/* Disk Usage */}
          <div>
            <div className="flex items-center justify-between mb-2">
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium text-foreground">Disk Usage</span>
                <span className={`text-xs px-2 py-1 rounded ${
                  getResourceStatus(node.resources.disk_usage).variant === 'error' ? 'bg-error/10 text-error' :
                  getResourceStatus(node.resources.disk_usage).variant === 'warning' ? 'bg-warning/10 text-warning' :
                  getResourceStatus(node.resources.disk_usage).variant === 'info' ? 'bg-info/10 text-info' :
                  'bg-success/10 text-success'
                }`}>
                  {getResourceStatus(node.resources.disk_usage).status}
                </span>
              </div>
              <span className="text-sm font-bold text-foreground">{node.resources.disk_usage.toFixed(1)}%</span>
            </div>
            <Progress
              value={node.resources.disk_usage}
              variant={getResourceStatus(node.resources.disk_usage).variant}
              size="md"
            />
            <div className="flex justify-between text-xs text-muted-foreground mt-1">
              <span>{formatBytes(node.capabilities.disk_space * 1024 * 1024)} total</span>
              <span>{formatBytes(node.capabilities.disk_space * 1024 * 1024 * node.resources.disk_usage / 100)} used</span>
            </div>
          </div>

          {/* Network I/O */}
          <div className="pt-4 border-t border-border">
            <h4 className="text-sm font-medium text-foreground mb-3">Network I/O</h4>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <div className="flex items-center gap-2 mb-1">
                  <svg className="w-4 h-4 text-success" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4l8 8-8 8M4 12h16" />
                  </svg>
                  <span className="text-xs text-muted-foreground">Inbound</span>
                </div>
                <p className="text-sm font-bold text-foreground">{formatBytesPerSecond(networkRxRate)}</p>
                <p className="text-xs text-muted-foreground">Total: {formatBytes(node.resources.network_io.rx_bytes)}</p>
              </div>
              <div>
                <div className="flex items-center gap-2 mb-1">
                  <svg className="w-4 h-4 text-info" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 12H4m16 0l-8-8m8 8l-8 8" />
                  </svg>
                  <span className="text-xs text-muted-foreground">Outbound</span>
                </div>
                <p className="text-sm font-bold text-foreground">{formatBytesPerSecond(networkTxRate)}</p>
                <p className="text-xs text-muted-foreground">Total: {formatBytes(node.resources.network_io.tx_bytes)}</p>
              </div>
            </div>
          </div>

          {/* Active Workload */}
          <div className="pt-4 border-t border-border">
            <h4 className="text-sm font-medium text-foreground mb-3">Active Workload</h4>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <p className="text-xs text-muted-foreground">Active Requests</p>
                <p className="text-2xl font-bold text-foreground">{node.resources.active_requests}</p>
                <p className="text-xs text-muted-foreground">of {node.capabilities.max_concurrent_requests} max</p>
              </div>
              <div>
                <p className="text-xs text-muted-foreground">Queue Size</p>
                <p className="text-2xl font-bold text-foreground">{node.resources.queue_size}</p>
                <p className="text-xs text-muted-foreground">pending requests</p>
              </div>
            </div>
            
            {/* Workload Progress Bar */}
            <div className="mt-3">
              <div className="flex justify-between text-xs text-muted-foreground mb-1">
                <span>Capacity Usage</span>
                <span>{((node.resources.active_requests / node.capabilities.max_concurrent_requests) * 100).toFixed(1)}%</span>
              </div>
              <Progress
                value={(node.resources.active_requests / node.capabilities.max_concurrent_requests) * 100}
                variant={
                  (node.resources.active_requests / node.capabilities.max_concurrent_requests) > 0.9 ? 'error' :
                  (node.resources.active_requests / node.capabilities.max_concurrent_requests) > 0.7 ? 'warning' :
                  'success'
                }
                size="sm"
              />
            </div>
          </div>

          {/* Mini Resource Trend Chart (Text-based) */}
          {history.length > 0 && (
            <div className="pt-4 border-t border-border">
              <h4 className="text-sm font-medium text-foreground mb-3">Resource Trends ({timeRange})</h4>
              <div className="space-y-2 text-xs">
                <div className="flex items-center justify-between">
                  <span className="text-muted-foreground">CPU:</span>
                  <div className="flex items-center gap-2">
                    <div className="w-20 bg-muted rounded-full h-1">
                      <div 
                        className="h-1 bg-primary rounded-full transition-all"
                        style={{ width: `${Math.min(100, Math.max(0, history[history.length - 1]?.cpu || 0))}%` }}
                      />
                    </div>
                    <span className="text-foreground font-medium min-w-[3rem]">
                      {(history[history.length - 1]?.cpu || 0).toFixed(1)}%
                    </span>
                  </div>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-muted-foreground">Memory:</span>
                  <div className="flex items-center gap-2">
                    <div className="w-20 bg-muted rounded-full h-1">
                      <div 
                        className="h-1 bg-secondary rounded-full transition-all"
                        style={{ width: `${Math.min(100, Math.max(0, history[history.length - 1]?.memory || 0))}%` }}
                      />
                    </div>
                    <span className="text-foreground font-medium min-w-[3rem]">
                      {(history[history.length - 1]?.memory || 0).toFixed(1)}%
                    </span>
                  </div>
                </div>
                {node.capabilities.gpu_enabled && (
                  <div className="flex items-center justify-between">
                    <span className="text-muted-foreground">GPU:</span>
                    <div className="flex items-center gap-2">
                      <div className="w-20 bg-muted rounded-full h-1">
                        <div 
                          className="h-1 bg-info rounded-full transition-all"
                          style={{ width: `${Math.min(100, Math.max(0, history[history.length - 1]?.gpu || 0))}%` }}
                        />
                      </div>
                      <span className="text-foreground font-medium min-w-[3rem]">
                        {(history[history.length - 1]?.gpu || 0).toFixed(1)}%
                      </span>
                    </div>
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}

export default ResourceMonitor