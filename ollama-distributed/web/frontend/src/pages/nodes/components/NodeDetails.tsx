import React from 'react'
import { Card, CardHeader, CardContent, CardTitle, CardDescription } from '@/design-system/components/Card/Card'
import type { Node } from '../NodesPage'

interface NodeDetailsProps {
  node: Node
}

export const NodeDetails: React.FC<NodeDetailsProps> = ({ node }) => {
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

  const formatDateTime = (isoString: string) => {
    return new Date(isoString).toLocaleString()
  }

  const formatUptime = (percentage: number) => {
    const hours = (percentage / 100) * 24 * 365 // Approximate hours in a year
    if (hours < 24) return `${hours.toFixed(1)}h`
    const days = Math.floor(hours / 24)
    return `${days}d`
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <svg className="w-5 h-5 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z" />
            </svg>
          Node Details
        </CardTitle>
        <CardDescription>
          Detailed information about {node.name}
        </CardDescription>
      </CardHeader>

      <CardContent spacing="md">
        {/* Basic Information */}
        <div className="space-y-3">
          <div>
            <h4 className="font-medium text-foreground mb-2">Basic Information</h4>
            <div className="grid grid-cols-1 gap-2 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Name:</span>
                <span className="font-medium text-foreground">{node.name}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Hostname:</span>
                <span className="font-medium text-foreground">{node.hostname}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">IP Address:</span>
                <span className="font-medium text-foreground">{node.ip_address}:{node.port}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Version:</span>
                <span className="font-medium text-foreground">{node.version}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Last Seen:</span>
                <span className="font-medium text-foreground">{formatDateTime(node.last_seen)}</span>
              </div>
            </div>
          </div>

          {/* Hardware Capabilities */}
          <div className="pt-3 border-t border-border">
            <h4 className="font-medium text-foreground mb-2">Hardware Capabilities</h4>
            <div className="grid grid-cols-1 gap-2 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">CPU Cores:</span>
                <span className="font-medium text-foreground">{node.capabilities.cpu_cores}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">RAM:</span>
                <span className="font-medium text-foreground">{formatBytes(node.capabilities.ram_total * 1024 * 1024)}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Disk Space:</span>
                <span className="font-medium text-foreground">{formatBytes(node.capabilities.disk_space * 1024 * 1024)}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">GPU Enabled:</span>
                <span className={`font-medium ${node.capabilities.gpu_enabled ? 'text-success' : 'text-muted-foreground'}`}>
                  {node.capabilities.gpu_enabled ? 'Yes' : 'No'}
                </span>
              </div>
              {node.capabilities.gpu_enabled && (
                <div className="flex justify-between">
                  <span className="text-muted-foreground">GPU Memory:</span>
                  <span className="font-medium text-foreground">{formatBytes(node.capabilities.gpu_memory * 1024 * 1024)}</span>
                </div>
              )}
              <div className="flex justify-between">
                <span className="text-muted-foreground">Max Concurrent:</span>
                <span className="font-medium text-foreground">{node.capabilities.max_concurrent_requests}</span>
              </div>
            </div>
          </div>

          {/* Location Information */}
          {node.location && (
            <div className="pt-3 border-t border-border">
              <h4 className="font-medium text-foreground mb-2">Location</h4>
              <div className="grid grid-cols-1 gap-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Region:</span>
                  <span className="font-medium text-foreground">{node.location.region}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Datacenter:</span>
                  <span className="font-medium text-foreground">{node.location.datacenter}</span>
                </div>
                {node.location.rack && (
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Rack:</span>
                    <span className="font-medium text-foreground">{node.location.rack}</span>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Performance Stats */}
          <div className="pt-3 border-t border-border">
            <h4 className="font-medium text-foreground mb-2">Performance</h4>
            <div className="grid grid-cols-1 gap-2 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Requests/sec:</span>
                <span className="font-medium text-foreground">{node.performance.requests_per_second.toFixed(1)}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Avg Response:</span>
                <span className="font-medium text-foreground">{node.performance.average_response_time}ms</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Error Rate:</span>
                <span className={`font-medium ${node.performance.error_rate > 0.05 ? 'text-error' : 'text-success'}`}>
                  {(node.performance.error_rate * 100).toFixed(2)}%
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Uptime:</span>
                <span className="font-medium text-success">{node.performance.uptime.toFixed(2)}%</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Tokens/sec:</span>
                <span className="font-medium text-foreground">{node.performance.tokens_per_second.toFixed(1)}</span>
              </div>
            </div>
          </div>

          {/* Network I/O */}
          <div className="pt-3 border-t border-border">
            <h4 className="font-medium text-foreground mb-2">Network I/O</h4>
            <div className="grid grid-cols-1 gap-2 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Received:</span>
                <span className="font-medium text-foreground">{formatBytes(node.resources.network_io.rx_bytes)}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Transmitted:</span>
                <span className="font-medium text-foreground">{formatBytes(node.resources.network_io.tx_bytes)}</span>
              </div>
            </div>
          </div>

          {/* Available Models */}
          <div className="pt-3 border-t border-border">
            <h4 className="font-medium text-foreground mb-2">Available Models</h4>
            <div className="flex flex-wrap gap-1">
              {node.capabilities.models.map(model => (
                <span key={model} className="text-xs bg-muted text-muted-foreground px-2 py-1 rounded">
                  {model}
                </span>
              ))}
            </div>
          </div>

          {/* Maintenance Schedule */}
          {node.maintenance?.scheduled && (
            <div className="pt-3 border-t border-border">
              <h4 className="font-medium text-foreground mb-2">Scheduled Maintenance</h4>
              <div className="grid grid-cols-1 gap-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Start:</span>
                  <span className="font-medium text-foreground">
                    {node.maintenance.start_time ? formatDateTime(node.maintenance.start_time) : 'Not set'}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">End:</span>
                  <span className="font-medium text-foreground">
                    {node.maintenance.end_time ? formatDateTime(node.maintenance.end_time) : 'Not set'}
                  </span>
                </div>
                {node.maintenance.reason && (
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Reason:</span>
                    <span className="font-medium text-foreground">{node.maintenance.reason}</span>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Active Alerts */}
          {node.alerts.length > 0 && (
            <div className="pt-3 border-t border-border">
              <h4 className="font-medium text-foreground mb-2">Recent Alerts</h4>
              <div className="space-y-2">
                {node.alerts.slice(0, 3).map(alert => (
                  <div key={alert.id} className={`p-2 rounded-lg border-l-4 ${
                    alert.severity === 'critical' ? 'bg-error/10 border-error' :
                    alert.severity === 'high' ? 'bg-warning/10 border-warning' :
                    alert.severity === 'medium' ? 'bg-info/10 border-info' :
                    'bg-muted/10 border-muted'
                  }`}>
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <p className="text-sm font-medium text-foreground">{alert.message}</p>
                        <p className="text-xs text-muted-foreground">
                          {formatDateTime(alert.timestamp)}
                        </p>
                      </div>
                      <span className={`text-xs px-2 py-1 rounded ${
                        alert.resolved ? 'bg-success/10 text-success' : 'bg-error/10 text-error'
                      }`}>
                        {alert.resolved ? 'Resolved' : 'Active'}
                      </span>
                    </div>
                  </div>
                ))}
                {node.alerts.length > 3 && (
                  <p className="text-xs text-muted-foreground text-center">
                    +{node.alerts.length - 3} more alerts
                  </p>
                )}
              </div>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}

export default NodeDetails