import React from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/design-system/components/Card/Card'
import { Progress } from '@/design-system/components/Progress/Progress'
import type { Node } from '../NodesPage'

interface NodesListProps {
  nodes: Node[]
  viewMode: 'grid' | 'list' | 'topology'
  selectedNode: Node | null
  onNodeSelect: (node: Node) => void
  onNodeAction: (nodeId: string, action: string, params?: any) => void
}

const StatusBadge: React.FC<{ status: Node['status'] }> = ({ status }) => {
  const statusConfig = {
    online: { bg: 'bg-success/10', text: 'text-success', label: 'Online' },
    offline: { bg: 'bg-error/10', text: 'text-error', label: 'Offline' },
    draining: { bg: 'bg-warning/10', text: 'text-warning', label: 'Draining' },
    maintenance: { bg: 'bg-info/10', text: 'text-info', label: 'Maintenance' },
    error: { bg: 'bg-error/10', text: 'text-error', label: 'Error' },
    unknown: { bg: 'bg-muted/10', text: 'text-muted-foreground', label: 'Unknown' }
  }

  const config = statusConfig[status] || statusConfig.unknown

  return (
    <span className={`px-2 py-1 rounded-full text-xs font-medium ${config.bg} ${config.text}`}>
      {config.label}
    </span>
  )
}

const NodeCard: React.FC<{
  node: Node
  isSelected: boolean
  onSelect: () => void
  onAction: (action: string, params?: any) => void
}> = ({ node, isSelected, onSelect, onAction }) => {
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

  const formatUptime = (hours: number) => {
    if (hours < 24) return `${hours.toFixed(1)}h`
    const days = Math.floor(hours / 24)
    const remainingHours = hours % 24
    return `${days}d ${remainingHours.toFixed(0)}h`
  }

  const unresolvedAlerts = node.alerts.filter(alert => !alert.resolved)
  const criticalAlerts = unresolvedAlerts.filter(alert => alert.severity === 'critical')

  return (
    <Card
      variant={isSelected ? 'interactive' : 'default'}
      className={`cursor-pointer transition-all ${
        isSelected ? 'ring-2 ring-primary border-primary' : 'hover:shadow-md'
      }`}
      onClick={onSelect}
    >
      <CardHeader>
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <CardTitle className="text-lg flex items-center gap-2">
              {node.name}
              {node.capabilities.gpu_enabled && (
                <span className="text-xs bg-purple-100 text-purple-800 px-2 py-1 rounded">
                  GPU
                </span>
              )}
            </CardTitle>
            <p className="text-sm text-muted-foreground">{node.hostname}</p>
            <p className="text-xs text-muted-foreground">{node.ip_address}:{node.port}</p>
          </div>
          <div className="flex flex-col items-end gap-2">
            <StatusBadge status={node.status} />
            {unresolvedAlerts.length > 0 && (
              <div className="flex items-center gap-1">
                <svg className={`w-4 h-4 ${criticalAlerts.length > 0 ? 'text-error' : 'text-warning'}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.464 0L4.35 16.5c-.77.833.192 2.5 1.732 2.5z" />
                </svg>
                <span className="text-xs text-muted-foreground">{unresolvedAlerts.length}</span>
              </div>
            )}
          </div>
        </div>
      </CardHeader>

      <CardContent>
        <div className="space-y-4">
          {/* Health Score */}
          <div>
            <div className="flex justify-between items-center mb-1">
              <span className="text-sm text-muted-foreground">Health Score</span>
              <span className="text-sm font-medium">{node.health_score}%</span>
            </div>
            <Progress 
              value={node.health_score} 
              variant={node.health_score > 80 ? 'success' : node.health_score > 60 ? 'warning' : 'error'}
              size="sm"
            />
          </div>

          {/* Resource Usage */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <div className="flex justify-between items-center mb-1">
                <span className="text-xs text-muted-foreground">CPU</span>
                <span className="text-xs font-medium">{node.resources.cpu_usage}%</span>
              </div>
              <Progress 
                value={node.resources.cpu_usage} 
                variant={node.resources.cpu_usage > 80 ? 'error' : node.resources.cpu_usage > 60 ? 'warning' : 'primary'}
                size="xs"
              />
            </div>

            <div>
              <div className="flex justify-between items-center mb-1">
                <span className="text-xs text-muted-foreground">Memory</span>
                <span className="text-xs font-medium">{node.resources.memory_usage}%</span>
              </div>
              <Progress 
                value={node.resources.memory_usage} 
                variant={node.resources.memory_usage > 80 ? 'error' : node.resources.memory_usage > 60 ? 'warning' : 'primary'}
                size="xs"
              />
            </div>

            {node.capabilities.gpu_enabled && node.resources.gpu_usage !== undefined && (
              <div>
                <div className="flex justify-between items-center mb-1">
                  <span className="text-xs text-muted-foreground">GPU</span>
                  <span className="text-xs font-medium">{node.resources.gpu_usage}%</span>
                </div>
                <Progress 
                  value={node.resources.gpu_usage} 
                  variant={node.resources.gpu_usage > 80 ? 'error' : node.resources.gpu_usage > 60 ? 'warning' : 'secondary'}
                  size="xs"
                />
              </div>
            )}

            <div>
              <div className="flex justify-between items-center mb-1">
                <span className="text-xs text-muted-foreground">Disk</span>
                <span className="text-xs font-medium">{node.resources.disk_usage}%</span>
              </div>
              <Progress 
                value={node.resources.disk_usage} 
                variant={node.resources.disk_usage > 80 ? 'error' : node.resources.disk_usage > 60 ? 'warning' : 'primary'}
                size="xs"
              />
            </div>
          </div>

          {/* Performance Metrics */}
          <div className="grid grid-cols-2 gap-4 pt-2 border-t border-border">
            <div>
              <p className="text-xs text-muted-foreground">Active Requests</p>
              <p className="text-lg font-bold text-foreground">{node.resources.active_requests}</p>
            </div>
            <div>
              <p className="text-xs text-muted-foreground">Queue Size</p>
              <p className="text-lg font-bold text-foreground">{node.resources.queue_size}</p>
            </div>
            <div>
              <p className="text-xs text-muted-foreground">RPS</p>
              <p className="text-sm font-medium text-foreground">{node.performance.requests_per_second.toFixed(1)}</p>
            </div>
            <div>
              <p className="text-xs text-muted-foreground">Avg Response</p>
              <p className="text-sm font-medium text-foreground">{node.performance.average_response_time}ms</p>
            </div>
          </div>

          {/* Models */}
          <div>
            <p className="text-xs text-muted-foreground mb-2">Available Models</p>
            <div className="flex flex-wrap gap-1">
              {node.capabilities.models.slice(0, 3).map(model => (
                <span key={model} className="text-xs bg-muted text-muted-foreground px-2 py-1 rounded">
                  {model}
                </span>
              ))}
              {node.capabilities.models.length > 3 && (
                <span className="text-xs text-muted-foreground">
                  +{node.capabilities.models.length - 3} more
                </span>
              )}
            </div>
          </div>

          {/* Quick Actions */}
          <div className="flex gap-2 pt-2">
            {node.status === 'offline' && (
              <button
                onClick={(e) => {
                  e.stopPropagation()
                  onAction('start')
                }}
                className="flex-1 px-3 py-1.5 bg-success text-success-foreground rounded text-xs hover:bg-success/90 transition-colors"
              >
                Start
              </button>
            )}
            {node.status === 'online' && (
              <>
                <button
                  onClick={(e) => {
                    e.stopPropagation()
                    onAction('drain')
                  }}
                  className="flex-1 px-3 py-1.5 bg-warning text-warning-foreground rounded text-xs hover:bg-warning/90 transition-colors"
                >
                  Drain
                </button>
                <button
                  onClick={(e) => {
                    e.stopPropagation()
                    onAction('stop')
                  }}
                  className="flex-1 px-3 py-1.5 bg-error text-error-foreground rounded text-xs hover:bg-error/90 transition-colors"
                >
                  Stop
                </button>
              </>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

const NodeListItem: React.FC<{
  node: Node
  isSelected: boolean
  onSelect: () => void
  onAction: (action: string, params?: any) => void
}> = ({ node, isSelected, onSelect, onAction }) => {
  const unresolvedAlerts = node.alerts.filter(alert => !alert.resolved)
  const criticalAlerts = unresolvedAlerts.filter(alert => alert.severity === 'critical')

  return (
    <Card
      variant={isSelected ? 'interactive' : 'default'}
      className={`cursor-pointer transition-all ${
        isSelected ? 'ring-2 ring-primary border-primary' : 'hover:shadow-sm'
      }`}
      onClick={onSelect}
      padding="md"
    >
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4 flex-1">
          {/* Node Info */}
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2">
              <h3 className="font-medium text-foreground">{node.name}</h3>
              {node.capabilities.gpu_enabled && (
                <span className="text-xs bg-purple-100 text-purple-800 px-2 py-1 rounded">
                  GPU
                </span>
              )}
              <StatusBadge status={node.status} />
              {unresolvedAlerts.length > 0 && (
                <div className="flex items-center gap-1">
                  <svg className={`w-4 h-4 ${criticalAlerts.length > 0 ? 'text-error' : 'text-warning'}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.464 0L4.35 16.5c-.77.833.192 2.5 1.732 2.5z" />
                  </svg>
                  <span className="text-xs text-muted-foreground">{unresolvedAlerts.length}</span>
                </div>
              )}
            </div>
            <p className="text-sm text-muted-foreground">{node.hostname} • {node.ip_address}:{node.port}</p>
          </div>

          {/* Health Score */}
          <div className="flex items-center gap-2">
            <span className="text-sm text-muted-foreground">Health:</span>
            <span className="text-sm font-medium">{node.health_score}%</span>
            <div className="w-16">
              <Progress 
                value={node.health_score} 
                variant={node.health_score > 80 ? 'success' : node.health_score > 60 ? 'warning' : 'error'}
                size="xs"
              />
            </div>
          </div>

          {/* Resource Usage */}
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-1">
              <span className="text-xs text-muted-foreground">CPU:</span>
              <span className="text-xs font-medium">{node.resources.cpu_usage}%</span>
              <div className="w-12">
                <Progress 
                  value={node.resources.cpu_usage} 
                  variant={node.resources.cpu_usage > 80 ? 'error' : 'primary'}
                  size="xs"
                />
              </div>
            </div>

            <div className="flex items-center gap-1">
              <span className="text-xs text-muted-foreground">Memory:</span>
              <span className="text-xs font-medium">{node.resources.memory_usage}%</span>
              <div className="w-12">
                <Progress 
                  value={node.resources.memory_usage} 
                  variant={node.resources.memory_usage > 80 ? 'error' : 'primary'}
                  size="xs"
                />
              </div>
            </div>

            {node.capabilities.gpu_enabled && node.resources.gpu_usage !== undefined && (
              <div className="flex items-center gap-1">
                <span className="text-xs text-muted-foreground">GPU:</span>
                <span className="text-xs font-medium">{node.resources.gpu_usage}%</span>
                <div className="w-12">
                  <Progress 
                    value={node.resources.gpu_usage} 
                    variant={node.resources.gpu_usage > 80 ? 'error' : 'secondary'}
                    size="xs"
                  />
                </div>
              </div>
            )}
          </div>

          {/* Performance */}
          <div className="flex items-center gap-4 text-xs text-muted-foreground">
            <div>
              <span>Requests: </span>
              <span className="font-medium text-foreground">{node.resources.active_requests}</span>
            </div>
            <div>
              <span>RPS: </span>
              <span className="font-medium text-foreground">{node.performance.requests_per_second.toFixed(1)}</span>
            </div>
            <div>
              <span>Response: </span>
              <span className="font-medium text-foreground">{node.performance.average_response_time}ms</span>
            </div>
          </div>
        </div>

        {/* Actions */}
        <div className="flex gap-2">
          {node.status === 'offline' && (
            <button
              onClick={(e) => {
                e.stopPropagation()
                onAction('start')
              }}
              className="px-3 py-1.5 bg-success text-success-foreground rounded text-xs hover:bg-success/90 transition-colors"
            >
              Start
            </button>
          )}
          {node.status === 'online' && (
            <>
              <button
                onClick={(e) => {
                  e.stopPropagation()
                  onAction('drain')
                }}
                className="px-3 py-1.5 bg-warning text-warning-foreground rounded text-xs hover:bg-warning/90 transition-colors"
              >
                Drain
              </button>
              <button
                onClick={(e) => {
                  e.stopPropagation()
                  onAction('stop')
                }}
                className="px-3 py-1.5 bg-error text-error-foreground rounded text-xs hover:bg-error/90 transition-colors"
              >
                Stop
              </button>
            </>
          )}
        </div>
      </div>
    </Card>
  )
}

export const NodesList: React.FC<NodesListProps> = ({
  nodes,
  viewMode,
  selectedNode,
  onNodeSelect,
  onNodeAction
}) => {
  if (nodes.length === 0) {
    return (
      <Card>
        <CardContent padding="xl">
          <div className="text-center py-12">
            <svg className="w-16 h-16 text-muted-foreground mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2" />
            </svg>
            <h3 className="text-lg font-medium text-foreground mb-2">No Nodes Found</h3>
            <p className="text-muted-foreground mb-4">
              No nodes match your current filter criteria.
            </p>
            <button className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors">
              Add Your First Node
            </button>
          </div>
        </CardContent>
      </Card>
    )
  }

  if (viewMode === 'list') {
    return (
      <div className="space-y-3">
        {nodes.map(node => (
          <NodeListItem
            key={node.id}
            node={node}
            isSelected={selectedNode?.id === node.id}
            onSelect={() => onNodeSelect(node)}
            onAction={(action, params) => onNodeAction(node.id, action, params)}
          />
        ))}
      </div>
    )
  }

  // Grid view
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      {nodes.map(node => (
        <NodeCard
          key={node.id}
          node={node}
          isSelected={selectedNode?.id === node.id}
          onSelect={() => onNodeSelect(node)}
          onAction={(action, params) => onNodeAction(node.id, action, params)}
        />
      ))}
    </div>
  )
}

export default NodesList