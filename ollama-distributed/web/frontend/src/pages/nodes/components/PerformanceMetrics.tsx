import React, { useState, useMemo } from 'react'
import { Card, CardHeader, CardContent, CardTitle } from '@/design-system/components/Card/Card'
import { Progress } from '@/design-system/components/Progress/Progress'
import type { Node } from '../NodesPage'

interface PerformanceMetricsProps {
  nodes: Node[]
}

export const PerformanceMetrics: React.FC<PerformanceMetricsProps> = ({ nodes }) => {
  const [metricType, setMetricType] = useState<'overview' | 'throughput' | 'latency' | 'errors' | 'resources'>('overview')
  const [timeRange, setTimeRange] = useState<'1h' | '6h' | '24h' | '7d'>('24h')

  // Calculate aggregate metrics
  const metrics = useMemo(() => {
    const onlineNodes = nodes.filter(node => node.status === 'online')
    const totalNodes = nodes.length
    
    if (onlineNodes.length === 0) {
      return {
        totalRequests: 0,
        avgResponseTime: 0,
        totalErrors: 0,
        errorRate: 0,
        totalThroughput: 0,
        avgHealthScore: 0,
        totalCapacity: 0,
        utilizedCapacity: 0,
        avgCpuUsage: 0,
        avgMemoryUsage: 0,
        avgGpuUsage: 0,
        totalUptime: 0,
        networkThroughput: { rx: 0, tx: 0 },
        capacityUtilization: 0
      }
    }

    const totalRequests = onlineNodes.reduce((sum, node) => sum + node.resources.active_requests, 0)
    const avgResponseTime = onlineNodes.reduce((sum, node) => sum + node.performance.average_response_time, 0) / onlineNodes.length
    const totalThroughput = onlineNodes.reduce((sum, node) => sum + node.performance.requests_per_second, 0)
    const avgHealthScore = nodes.reduce((sum, node) => sum + node.health_score, 0) / totalNodes
    const totalCapacity = nodes.reduce((sum, node) => sum + node.capabilities.max_concurrent_requests, 0)
    const utilizedCapacity = nodes.reduce((sum, node) => sum + node.resources.active_requests, 0)
    const avgCpuUsage = onlineNodes.reduce((sum, node) => sum + node.resources.cpu_usage, 0) / onlineNodes.length
    const avgMemoryUsage = onlineNodes.reduce((sum, node) => sum + node.resources.memory_usage, 0) / onlineNodes.length
    const avgGpuUsage = onlineNodes.filter(n => n.capabilities.gpu_enabled && n.resources.gpu_usage !== undefined)
      .reduce((sum, node) => sum + (node.resources.gpu_usage || 0), 0) / onlineNodes.filter(n => n.capabilities.gpu_enabled).length || 0
    const totalUptime = onlineNodes.reduce((sum, node) => sum + node.performance.uptime, 0) / onlineNodes.length
    const networkThroughput = onlineNodes.reduce((acc, node) => ({
      rx: acc.rx + node.resources.network_io.rx_bytes,
      tx: acc.tx + node.resources.network_io.tx_bytes
    }), { rx: 0, tx: 0 })

    const totalErrors = onlineNodes.reduce((sum, node) => {
      const errorCount = Math.round(node.performance.requests_per_second * node.performance.error_rate * 3600) // Errors per hour
      return sum + errorCount
    }, 0)
    
    const errorRate = onlineNodes.reduce((sum, node) => sum + node.performance.error_rate, 0) / onlineNodes.length

    return {
      totalRequests,
      avgResponseTime,
      totalErrors,
      errorRate,
      totalThroughput,
      avgHealthScore,
      totalCapacity,
      utilizedCapacity,
      avgCpuUsage,
      avgMemoryUsage,
      avgGpuUsage,
      totalUptime,
      networkThroughput,
      capacityUtilization: totalCapacity > 0 ? (utilizedCapacity / totalCapacity) * 100 : 0
    }
  }, [nodes])

  const formatBytes = (bytes: number) => {
    const units = ['B', 'KB', 'MB', 'GB', 'TB']
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

  const getTopPerformers = (metric: keyof Node['performance'] | keyof Node['resources']) => {
    return [...nodes]
      .filter(node => node.status === 'online')
      .sort((a, b) => {
        if (metric in a.performance) {
          const aVal = a.performance[metric as keyof Node['performance']] as number
          const bVal = b.performance[metric as keyof Node['performance']] as number
          return bVal - aVal
        } else {
          const aVal = a.resources[metric as keyof Node['resources']] as number
          const bVal = b.resources[metric as keyof Node['resources']] as number
          return bVal - aVal
        }
      })
      .slice(0, 3)
  }

  const getBottomPerformers = (metric: keyof Node['performance']) => {
    return [...nodes]
      .filter(node => node.status === 'online')
      .sort((a, b) => {
        const aVal = a.performance[metric] as number
        const bVal = b.performance[metric] as number
        return aVal - bVal
      })
      .slice(0, 3)
  }

  const renderOverviewMetrics = () => (
    <div className="space-y-4">
      {/* Key Performance Indicators */}
      <div className="grid grid-cols-2 gap-4">
        <div className="p-3 bg-primary/10 rounded-lg">
          <p className="text-xs text-muted-foreground">Total Throughput</p>
          <p className="text-lg font-bold text-primary">{metrics.totalThroughput.toFixed(1)}</p>
          <p className="text-xs text-muted-foreground">requests/sec</p>
        </div>
        <div className="p-3 bg-success/10 rounded-lg">
          <p className="text-xs text-muted-foreground">Avg Response Time</p>
          <p className="text-lg font-bold text-success">{metrics.avgResponseTime.toFixed(0)}</p>
          <p className="text-xs text-muted-foreground">milliseconds</p>
        </div>
        <div className="p-3 bg-warning/10 rounded-lg">
          <p className="text-xs text-muted-foreground">Error Rate</p>
          <p className="text-lg font-bold text-warning">{(metrics.errorRate * 100).toFixed(2)}</p>
          <p className="text-xs text-muted-foreground">percent</p>
        </div>
        <div className="p-3 bg-info/10 rounded-lg">
          <p className="text-xs text-muted-foreground">Capacity Usage</p>
          <p className="text-lg font-bold text-info">{metrics.capacityUtilization.toFixed(1)}</p>
          <p className="text-xs text-muted-foreground">percent</p>
        </div>
      </div>

      {/* Capacity Utilization */}
      <div>
        <div className="flex justify-between items-center mb-2">
          <span className="text-sm font-medium text-foreground">Cluster Capacity</span>
          <span className="text-sm text-muted-foreground">
            {metrics.utilizedCapacity} / {metrics.totalCapacity} requests
          </span>
        </div>
        <Progress
          value={metrics.capacityUtilization}
          variant={metrics.capacityUtilization > 80 ? 'warning' : 'primary'}
          size="md"
        />
      </div>

      {/* System Health */}
      <div>
        <div className="flex justify-between items-center mb-2">
          <span className="text-sm font-medium text-foreground">Average Health Score</span>
          <span className="text-sm font-medium text-foreground">{metrics.avgHealthScore.toFixed(1)}%</span>
        </div>
        <Progress
          value={metrics.avgHealthScore}
          variant={metrics.avgHealthScore > 80 ? 'success' : metrics.avgHealthScore > 60 ? 'warning' : 'error'}
          size="md"
        />
      </div>

      {/* Resource Utilization */}
      <div className="space-y-2">
        <h4 className="text-sm font-medium text-foreground">Resource Utilization</h4>
        <div className="space-y-2">
          <div>
            <div className="flex justify-between text-xs mb-1">
              <span className="text-muted-foreground">CPU</span>
              <span className="text-foreground">{metrics.avgCpuUsage.toFixed(1)}%</span>
            </div>
            <Progress value={metrics.avgCpuUsage} variant="primary" size="xs" />
          </div>
          <div>
            <div className="flex justify-between text-xs mb-1">
              <span className="text-muted-foreground">Memory</span>
              <span className="text-foreground">{metrics.avgMemoryUsage.toFixed(1)}%</span>
            </div>
            <Progress value={metrics.avgMemoryUsage} variant="secondary" size="xs" />
          </div>
          {metrics.avgGpuUsage > 0 && (
            <div>
              <div className="flex justify-between text-xs mb-1">
                <span className="text-muted-foreground">GPU</span>
                <span className="text-foreground">{metrics.avgGpuUsage.toFixed(1)}%</span>
              </div>
              <Progress value={metrics.avgGpuUsage} variant="info" size="xs" />
            </div>
          )}
        </div>
      </div>
    </div>
  )

  const renderThroughputMetrics = () => (
    <div className="space-y-4">
      <div className="p-3 bg-primary/10 rounded-lg text-center">
        <p className="text-xs text-muted-foreground">Total Cluster Throughput</p>
        <p className="text-2xl font-bold text-primary">{metrics.totalThroughput.toFixed(1)}</p>
        <p className="text-xs text-muted-foreground">requests per second</p>
      </div>

      <div>
        <h4 className="text-sm font-medium text-foreground mb-2">Top Performers (RPS)</h4>
        <div className="space-y-2">
          {getTopPerformers('requests_per_second').map((node, index) => (
            <div key={node.id} className="flex items-center justify-between p-2 bg-muted/50 rounded">
              <div className="flex items-center gap-2">
                <span className="text-xs bg-primary text-primary-foreground w-5 h-5 rounded-full flex items-center justify-center">
                  {index + 1}
                </span>
                <span className="text-sm font-medium">{node.name}</span>
              </div>
              <span className="text-sm text-primary font-bold">{node.performance.requests_per_second.toFixed(1)}</span>
            </div>
          ))}
        </div>
      </div>

      <div>
        <h4 className="text-sm font-medium text-foreground mb-2">Network Throughput</h4>
        <div className="grid grid-cols-2 gap-4">
          <div className="p-3 bg-success/10 rounded-lg">
            <p className="text-xs text-muted-foreground">Inbound</p>
            <p className="text-lg font-bold text-success">{formatBytesPerSecond(metrics.networkThroughput.rx / 3600)}</p>
          </div>
          <div className="p-3 bg-info/10 rounded-lg">
            <p className="text-xs text-muted-foreground">Outbound</p>
            <p className="text-lg font-bold text-info">{formatBytesPerSecond(metrics.networkThroughput.tx / 3600)}</p>
          </div>
        </div>
      </div>

      <div>
        <h4 className="text-sm font-medium text-foreground mb-2">Token Generation Rate</h4>
        <div className="space-y-2">
          {nodes.filter(n => n.status === 'online').map(node => (
            <div key={node.id} className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">{node.name}</span>
              <span className="font-medium">{node.performance.tokens_per_second.toFixed(1)} tok/s</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  )

  const renderLatencyMetrics = () => (
    <div className="space-y-4">
      <div className="p-3 bg-warning/10 rounded-lg text-center">
        <p className="text-xs text-muted-foreground">Average Response Time</p>
        <p className="text-2xl font-bold text-warning">{metrics.avgResponseTime.toFixed(0)}</p>
        <p className="text-xs text-muted-foreground">milliseconds</p>
      </div>

      <div>
        <h4 className="text-sm font-medium text-foreground mb-2">Best Response Times</h4>
        <div className="space-y-2">
          {getBottomPerformers('average_response_time').map((node, index) => (
            <div key={node.id} className="flex items-center justify-between p-2 bg-muted/50 rounded">
              <div className="flex items-center gap-2">
                <span className="text-xs bg-success text-success-foreground w-5 h-5 rounded-full flex items-center justify-center">
                  {index + 1}
                </span>
                <span className="text-sm font-medium">{node.name}</span>
              </div>
              <span className="text-sm text-success font-bold">{node.performance.average_response_time}ms</span>
            </div>
          ))}
        </div>
      </div>

      <div>
        <h4 className="text-sm font-medium text-foreground mb-2">Response Time Distribution</h4>
        <div className="space-y-2">
          {nodes.filter(n => n.status === 'online').map(node => {
            const responseTime = node.performance.average_response_time
            const maxTime = Math.max(...nodes.map(n => n.performance.average_response_time))
            const percentage = (responseTime / maxTime) * 100
            
            return (
              <div key={node.id}>
                <div className="flex justify-between text-xs mb-1">
                  <span className="text-muted-foreground">{node.name}</span>
                  <span className="text-foreground">{responseTime}ms</span>
                </div>
                <Progress 
                  value={percentage} 
                  variant={responseTime < 1000 ? 'success' : responseTime < 2000 ? 'warning' : 'error'} 
                  size="xs" 
                />
              </div>
            )
          })}
        </div>
      </div>

      <div>
        <h4 className="text-sm font-medium text-foreground mb-2">Uptime Statistics</h4>
        <div className="space-y-2">
          {nodes.map(node => (
            <div key={node.id} className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">{node.name}</span>
              <span className={`font-medium ${node.performance.uptime > 99 ? 'text-success' : node.performance.uptime > 95 ? 'text-warning' : 'text-error'}`}>
                {node.performance.uptime.toFixed(2)}%
              </span>
            </div>
          ))}
        </div>
      </div>
    </div>
  )

  const renderErrorMetrics = () => (
    <div className="space-y-4">
      <div className="p-3 bg-error/10 rounded-lg text-center">
        <p className="text-xs text-muted-foreground">Cluster Error Rate</p>
        <p className="text-2xl font-bold text-error">{(metrics.errorRate * 100).toFixed(3)}</p>
        <p className="text-xs text-muted-foreground">percent</p>
      </div>

      <div>
        <h4 className="text-sm font-medium text-foreground mb-2">Error Rates by Node</h4>
        <div className="space-y-2">
          {[...nodes]
            .filter(n => n.status === 'online')
            .sort((a, b) => b.performance.error_rate - a.performance.error_rate)
            .map(node => {
              const errorRate = node.performance.error_rate * 100
              return (
                <div key={node.id}>
                  <div className="flex justify-between text-xs mb-1">
                    <span className="text-muted-foreground">{node.name}</span>
                    <span className={`font-medium ${errorRate < 1 ? 'text-success' : errorRate < 5 ? 'text-warning' : 'text-error'}`}>
                      {errorRate.toFixed(3)}%
                    </span>
                  </div>
                  <Progress 
                    value={Math.min(100, errorRate * 20)} // Scale for visualization
                    variant={errorRate < 1 ? 'success' : errorRate < 5 ? 'warning' : 'error'}
                    size="xs" 
                  />
                </div>
              )
            })}
        </div>
      </div>

      <div>
        <h4 className="text-sm font-medium text-foreground mb-2">Estimated Errors ({timeRange})</h4>
        <div className="space-y-2">
          {nodes.filter(n => n.status === 'online').map(node => {
            const hoursMultiplier = timeRange === '1h' ? 1 : timeRange === '6h' ? 6 : timeRange === '24h' ? 24 : 168
            const estimatedErrors = Math.round(node.performance.requests_per_second * node.performance.error_rate * 3600 * hoursMultiplier)
            
            return (
              <div key={node.id} className="flex items-center justify-between text-sm">
                <span className="text-muted-foreground">{node.name}</span>
                <span className={`font-medium ${estimatedErrors < 10 ? 'text-success' : estimatedErrors < 100 ? 'text-warning' : 'text-error'}`}>
                  {estimatedErrors}
                </span>
              </div>
            )
          })}
        </div>
      </div>

      <div>
        <h4 className="text-sm font-medium text-foreground mb-2">Active Alerts</h4>
        <div className="space-y-2">
          {nodes.filter(n => n.alerts.some(a => !a.resolved)).map(node => {
            const activeAlerts = node.alerts.filter(a => !a.resolved)
            const criticalAlerts = activeAlerts.filter(a => a.severity === 'critical')
            
            return (
              <div key={node.id} className="p-2 bg-muted/50 rounded">
                <div className="flex items-center justify-between mb-1">
                  <span className="text-sm font-medium">{node.name}</span>
                  <span className="text-xs text-muted-foreground">{activeAlerts.length} alerts</span>
                </div>
                {criticalAlerts.length > 0 && (
                  <div className="text-xs text-error">
                    {criticalAlerts.length} critical alert{criticalAlerts.length !== 1 ? 's' : ''}
                  </div>
                )}
              </div>
            )
          })}
          {nodes.filter(n => n.alerts.some(a => !a.resolved)).length === 0 && (
            <div className="text-center py-4 text-muted-foreground text-sm">
              No active alerts
            </div>
          )}
        </div>
      </div>
    </div>
  )

  const renderResourceMetrics = () => (
    <div className="space-y-4">
      <div className="grid grid-cols-3 gap-3">
        <div className="p-3 bg-primary/10 rounded-lg text-center">
          <p className="text-xs text-muted-foreground">Avg CPU</p>
          <p className="text-lg font-bold text-primary">{metrics.avgCpuUsage.toFixed(1)}%</p>
        </div>
        <div className="p-3 bg-secondary/10 rounded-lg text-center">
          <p className="text-xs text-muted-foreground">Avg Memory</p>
          <p className="text-lg font-bold text-secondary">{metrics.avgMemoryUsage.toFixed(1)}%</p>
        </div>
        {metrics.avgGpuUsage > 0 && (
          <div className="p-3 bg-info/10 rounded-lg text-center">
            <p className="text-xs text-muted-foreground">Avg GPU</p>
            <p className="text-lg font-bold text-info">{metrics.avgGpuUsage.toFixed(1)}%</p>
          </div>
        )}
      </div>

      <div>
        <h4 className="text-sm font-medium text-foreground mb-2">CPU Utilization</h4>
        <div className="space-y-2">
          {getTopPerformers('cpu_usage').map(node => (
            <div key={node.id}>
              <div className="flex justify-between text-xs mb-1">
                <span className="text-muted-foreground">{node.name}</span>
                <span className="text-foreground">{node.resources.cpu_usage.toFixed(1)}%</span>
              </div>
              <Progress 
                value={node.resources.cpu_usage} 
                variant={node.resources.cpu_usage > 80 ? 'error' : node.resources.cpu_usage > 60 ? 'warning' : 'primary'} 
                size="xs" 
              />
            </div>
          ))}
        </div>
      </div>

      <div>
        <h4 className="text-sm font-medium text-foreground mb-2">Memory Utilization</h4>
        <div className="space-y-2">
          {getTopPerformers('memory_usage').map(node => (
            <div key={node.id}>
              <div className="flex justify-between text-xs mb-1">
                <span className="text-muted-foreground">{node.name}</span>
                <span className="text-foreground">{node.resources.memory_usage.toFixed(1)}%</span>
              </div>
              <Progress 
                value={node.resources.memory_usage} 
                variant={node.resources.memory_usage > 80 ? 'error' : node.resources.memory_usage > 60 ? 'warning' : 'secondary'} 
                size="xs" 
              />
            </div>
          ))}
        </div>
      </div>

      <div>
        <h4 className="text-sm font-medium text-foreground mb-2">Active Workload</h4>
        <div className="space-y-2">
          {nodes.filter(n => n.status === 'online').map(node => (
            <div key={node.id} className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">{node.name}</span>
              <div className="text-right">
                <div className="font-medium">{node.resources.active_requests}/{node.capabilities.max_concurrent_requests}</div>
                <div className="text-xs text-muted-foreground">
                  {node.resources.queue_size > 0 ? `+${node.resources.queue_size} queued` : 'no queue'}
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )

  const renderContent = () => {
    switch (metricType) {
      case 'throughput': return renderThroughputMetrics()
      case 'latency': return renderLatencyMetrics()
      case 'errors': return renderErrorMetrics()
      case 'resources': return renderResourceMetrics()
      default: return renderOverviewMetrics()
    }
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            <svg className="w-5 h-5 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v4a2 2 0 00-2 2v6a2 2 0 00-2 2zm-2-4h2m-2-8h2m-2 4h2" />
            </svg>
            Performance
          </CardTitle>
          
          <div className="flex items-center gap-2">
            <select
              value={timeRange}
              onChange={(e) => setTimeRange(e.target.value as any)}
              className="text-xs px-2 py-1 border border-border rounded bg-background text-foreground"
            >
              <option value="1h">1H</option>
              <option value="6h">6H</option>
              <option value="24h">24H</option>
              <option value="7d">7D</option>
            </select>
          </div>
        </div>
        
        <div className="flex flex-wrap gap-1 mt-3">
          {[
            { key: 'overview', label: 'Overview' },
            { key: 'throughput', label: 'Throughput' },
            { key: 'latency', label: 'Latency' },
            { key: 'errors', label: 'Errors' },
            { key: 'resources', label: 'Resources' }
          ].map(({ key, label }) => (
            <button
              key={key}
              onClick={() => setMetricType(key as any)}
              className={`text-xs px-2 py-1 rounded transition-colors ${
                metricType === key 
                  ? 'bg-primary text-primary-foreground' 
                  : 'bg-muted text-muted-foreground hover:bg-muted/80'
              }`}
            >
              {label}
            </button>
          ))}
        </div>
      </CardHeader>

      <CardContent>
        {renderContent()}
      </CardContent>
    </Card>
  )
}

export default PerformanceMetrics