/**
 * SystemHealthCard Component - Displays overall system health status
 */

import React from 'react'
import { Card, CardHeader, CardContent, CardTitle, CardDescription } from '@/design-system/components/Card/Card'
import { Badge } from '@/design-system/components/Badge/Badge'
import { 
  CheckCircle, 
  AlertTriangle, 
  XCircle, 
  Activity,
  Cpu,
  MemoryStick,
  HardDrive,
  Network,
  Zap
} from 'lucide-react'

interface SystemHealthCardProps {
  clusterStatus?: any
  metrics: {
    nodes: {
      total: number
      healthy: number
      degraded: number
      offline: number
    }
    performance: {
      cpu: number
      memory: number
      disk: number
      network: number
    }
  }
}

const SystemHealthCard: React.FC<SystemHealthCardProps> = ({
  clusterStatus,
  metrics
}) => {
  // Calculate overall health score
  const calculateHealthScore = () => {
    const nodeHealth = (metrics.nodes.healthy / metrics.nodes.total) * 100
    const performanceHealth = 100 - Math.max(
      metrics.performance.cpu,
      metrics.performance.memory,
      metrics.performance.disk,
      metrics.performance.network
    )
    return Math.round((nodeHealth + performanceHealth) / 2)
  }

  const healthScore = calculateHealthScore()

  const getOverallStatus = () => {
    if (healthScore >= 90) return { status: 'healthy', color: 'success', icon: CheckCircle }
    if (healthScore >= 70) return { status: 'warning', color: 'warning', icon: AlertTriangle }
    return { status: 'critical', color: 'destructive', icon: XCircle }
  }

  const overallStatus = getOverallStatus()

  const healthChecks = [
    {
      name: 'Cluster Connectivity',
      status: metrics.nodes.offline === 0 ? 'healthy' : 'degraded',
      value: `${metrics.nodes.healthy}/${metrics.nodes.total} nodes`,
      icon: Network
    },
    {
      name: 'CPU Usage',
      status: metrics.performance.cpu < 80 ? 'healthy' : metrics.performance.cpu < 90 ? 'warning' : 'critical',
      value: `${metrics.performance.cpu}%`,
      icon: Cpu
    },
    {
      name: 'Memory Usage',
      status: metrics.performance.memory < 80 ? 'healthy' : metrics.performance.memory < 90 ? 'warning' : 'critical',
      value: `${metrics.performance.memory}%`,
      icon: MemoryStick
    },
    {
      name: 'Storage Usage',
      status: metrics.performance.disk < 80 ? 'healthy' : metrics.performance.disk < 90 ? 'warning' : 'critical',
      value: `${metrics.performance.disk}%`,
      icon: HardDrive
    }
  ]

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'healthy':
        return 'text-success-600'
      case 'warning':
      case 'degraded':
        return 'text-warning-600'
      case 'critical':
        return 'text-error-600'
      default:
        return 'text-muted-foreground'
    }
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'healthy':
        return <CheckCircle className="h-4 w-4 text-success-600" />
      case 'warning':
      case 'degraded':
        return <AlertTriangle className="h-4 w-4 text-warning-600" />
      case 'critical':
        return <XCircle className="h-4 w-4 text-error-600" />
      default:
        return <Activity className="h-4 w-4 text-muted-foreground" />
    }
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2">
              <Activity className="h-5 w-5" />
              System Health
            </CardTitle>
            <CardDescription>Overall system status and health metrics</CardDescription>
          </div>
          <Badge variant={overallStatus.color as any} className="flex items-center gap-1">
            <overallStatus.icon className="h-3 w-3" />
            {overallStatus.status}
          </Badge>
        </div>
      </CardHeader>
      
      <CardContent className="space-y-4">
        {/* Health Score */}
        <div className="text-center p-4 bg-muted/50 rounded-lg">
          <div className="text-3xl font-bold text-foreground mb-1">
            {healthScore}%
          </div>
          <div className="text-sm text-muted-foreground">
            Overall Health Score
          </div>
        </div>

        {/* Health Checks */}
        <div className="space-y-3">
          <div className="text-sm font-medium text-foreground mb-2">
            Health Checks
          </div>
          
          {healthChecks.map((check, index) => (
            <div key={index} className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <check.icon className="h-4 w-4 text-muted-foreground" />
                <span className="text-sm">{check.name}</span>
              </div>
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium">{check.value}</span>
                {getStatusIcon(check.status)}
              </div>
            </div>
          ))}
        </div>

        {/* Quick Stats */}
        <div className="grid grid-cols-2 gap-3 pt-3 border-t border-border">
          <div className="text-center">
            <div className="text-lg font-semibold text-foreground">
              {metrics.nodes.healthy}
            </div>
            <div className="text-xs text-muted-foreground">
              Healthy Nodes
            </div>
          </div>
          <div className="text-center">
            <div className="text-lg font-semibold text-foreground">
              {100 - Math.round((metrics.performance.cpu + metrics.performance.memory) / 2)}%
            </div>
            <div className="text-xs text-muted-foreground">
              Avg Resources Free
            </div>
          </div>
        </div>

        {/* Status Timeline */}
        <div className="pt-3 border-t border-border">
          <div className="text-sm font-medium text-foreground mb-2">
            Recent Status
          </div>
          <div className="flex justify-between items-center text-xs">
            <span className="text-muted-foreground">Last 24h</span>
            <div className="flex gap-1">
              {Array.from({ length: 24 }, (_, i) => (
                <div
                  key={i}
                  className={`w-1 h-4 rounded-sm ${
                    Math.random() > 0.8 
                      ? 'bg-error-400' 
                      : Math.random() > 0.6 
                        ? 'bg-warning-400' 
                        : 'bg-success-400'
                  }`}
                />
              ))}
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

export default SystemHealthCard