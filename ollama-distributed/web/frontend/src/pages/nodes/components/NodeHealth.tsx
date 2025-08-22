import React, { useState } from 'react'
import { Card, CardHeader, CardContent, CardTitle } from '@/design-system/components/Card/Card'
import { Progress } from '@/design-system/components/Progress/Progress'
import type { Node } from '../NodesPage'

interface NodeHealthProps {
  node: Node
}

interface HealthCheck {
  name: string
  status: 'healthy' | 'warning' | 'critical' | 'unknown'
  value: number | string
  threshold?: number
  unit?: string
  description: string
  lastChecked: string
}

export const NodeHealth: React.FC<NodeHealthProps> = ({ node }) => {
  const [showDetails, setShowDetails] = useState(false)

  // Calculate health checks based on node data
  const healthChecks: HealthCheck[] = [
    {
      name: 'Connectivity',
      status: node.status === 'online' ? 'healthy' : 'critical',
      value: node.status === 'online' ? 'Connected' : 'Disconnected',
      description: 'Node connectivity and availability',
      lastChecked: node.last_seen
    },
    {
      name: 'CPU Load',
      status: node.resources.cpu_usage > 90 ? 'critical' : node.resources.cpu_usage > 75 ? 'warning' : 'healthy',
      value: node.resources.cpu_usage,
      threshold: 80,
      unit: '%',
      description: 'Current CPU utilization',
      lastChecked: node.last_seen
    },
    {
      name: 'Memory Usage',
      status: node.resources.memory_usage > 85 ? 'critical' : node.resources.memory_usage > 70 ? 'warning' : 'healthy',
      value: node.resources.memory_usage,
      threshold: 80,
      unit: '%',
      description: 'Current memory utilization',
      lastChecked: node.last_seen
    },
    {
      name: 'Disk Space',
      status: node.resources.disk_usage > 90 ? 'critical' : node.resources.disk_usage > 80 ? 'warning' : 'healthy',
      value: node.resources.disk_usage,
      threshold: 85,
      unit: '%',
      description: 'Available disk space',
      lastChecked: node.last_seen
    },
    {
      name: 'Response Time',
      status: node.performance.average_response_time > 2000 ? 'critical' : node.performance.average_response_time > 1000 ? 'warning' : 'healthy',
      value: node.performance.average_response_time,
      threshold: 1500,
      unit: 'ms',
      description: 'Average API response time',
      lastChecked: node.last_seen
    },
    {
      name: 'Error Rate',
      status: node.performance.error_rate > 0.05 ? 'critical' : node.performance.error_rate > 0.02 ? 'warning' : 'healthy',
      value: (node.performance.error_rate * 100).toFixed(2),
      threshold: 3,
      unit: '%',
      description: 'Request error rate',
      lastChecked: node.last_seen
    },
    {
      name: 'Uptime',
      status: node.performance.uptime < 95 ? 'critical' : node.performance.uptime < 99 ? 'warning' : 'healthy',
      value: node.performance.uptime.toFixed(2),
      threshold: 99,
      unit: '%',
      description: 'Service uptime percentage',
      lastChecked: node.last_seen
    }
  ]

  // Add GPU health check if GPU is enabled
  if (node.capabilities.gpu_enabled && node.resources.gpu_usage !== undefined) {
    healthChecks.push({
      name: 'GPU Usage',
      status: node.resources.gpu_usage > 95 ? 'critical' : node.resources.gpu_usage > 85 ? 'warning' : 'healthy',
      value: node.resources.gpu_usage,
      threshold: 90,
      unit: '%',
      description: 'GPU utilization and memory',
      lastChecked: node.last_seen
    })
  }

  const healthyChecks = healthChecks.filter(check => check.status === 'healthy')
  const warningChecks = healthChecks.filter(check => check.status === 'warning')
  const criticalChecks = healthChecks.filter(check => check.status === 'critical')

  const getStatusIcon = (status: HealthCheck['status']) => {
    switch (status) {
      case 'healthy':
        return (
          <svg className="w-4 h-4 text-success" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        )
      case 'warning':
        return (
          <svg className="w-4 h-4 text-warning" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.464 0L4.35 16.5c-.77.833.192 2.5 1.732 2.5z" />
          </svg>
        )
      case 'critical':
        return (
          <svg className="w-4 h-4 text-error" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        )
      default:
        return (
          <svg className="w-4 h-4 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        )
    }
  }

  const getStatusColor = (status: HealthCheck['status']) => {
    switch (status) {
      case 'healthy': return 'text-success'
      case 'warning': return 'text-warning'
      case 'critical': return 'text-error'
      default: return 'text-muted-foreground'
    }
  }

  const formatDateTime = (isoString: string) => {
    return new Date(isoString).toLocaleString()
  }

  const overallHealthStatus = criticalChecks.length > 0 ? 'critical' : warningChecks.length > 0 ? 'warning' : 'healthy'

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <svg className="w-5 h-5 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
          </svg>
          Node Health
        </CardTitle>
      </CardHeader>

      <CardContent spacing="md">
        {/* Overall Health Score */}
        <div className="mb-6">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm font-medium text-foreground">Overall Health Score</span>
            <div className="flex items-center gap-2">
              {getStatusIcon(overallHealthStatus)}
              <span className={`text-lg font-bold ${getStatusColor(overallHealthStatus)}`}>
                {node.health_score}%
              </span>
            </div>
          </div>
          <Progress
            value={node.health_score}
            variant={node.health_score > 80 ? 'success' : node.health_score > 60 ? 'warning' : 'error'}
            size="lg"
          />
          <div className="flex justify-between text-xs text-muted-foreground mt-1">
            <span>Poor</span>
            <span>Good</span>
            <span>Excellent</span>
          </div>
        </div>

        {/* Health Summary */}
        <div className="grid grid-cols-3 gap-4 mb-6">
          <div className="text-center">
            <div className="flex items-center justify-center gap-1 mb-1">
              <svg className="w-4 h-4 text-success" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span className="text-lg font-bold text-success">{healthyChecks.length}</span>
            </div>
            <p className="text-xs text-muted-foreground">Healthy</p>
          </div>
          <div className="text-center">
            <div className="flex items-center justify-center gap-1 mb-1">
              <svg className="w-4 h-4 text-warning" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.464 0L4.35 16.5c-.77.833.192 2.5 1.732 2.5z" />
              </svg>
              <span className="text-lg font-bold text-warning">{warningChecks.length}</span>
            </div>
            <p className="text-xs text-muted-foreground">Warnings</p>
          </div>
          <div className="text-center">
            <div className="flex items-center justify-center gap-1 mb-1">
              <svg className="w-4 h-4 text-error" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span className="text-lg font-bold text-error">{criticalChecks.length}</span>
            </div>
            <p className="text-xs text-muted-foreground">Critical</p>
          </div>
        </div>

        {/* Critical Issues (if any) */}
        {criticalChecks.length > 0 && (
          <div className="mb-4 p-3 bg-error/10 border border-error/20 rounded-lg">
            <h4 className="text-sm font-medium text-error mb-2">Critical Issues</h4>
            <div className="space-y-2">
              {criticalChecks.map(check => (
                <div key={check.name} className="flex items-center justify-between text-sm">
                  <span className="text-foreground">{check.name}</span>
                  <span className="text-error font-medium">
                    {typeof check.value === 'number' ? check.value.toFixed(1) : check.value}
                    {check.unit}
                  </span>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Warning Issues (if any) */}
        {warningChecks.length > 0 && (
          <div className="mb-4 p-3 bg-warning/10 border border-warning/20 rounded-lg">
            <h4 className="text-sm font-medium text-warning mb-2">Warnings</h4>
            <div className="space-y-2">
              {warningChecks.map(check => (
                <div key={check.name} className="flex items-center justify-between text-sm">
                  <span className="text-foreground">{check.name}</span>
                  <span className="text-warning font-medium">
                    {typeof check.value === 'number' ? check.value.toFixed(1) : check.value}
                    {check.unit}
                  </span>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Health Check Details */}
        <div>
          <button
            onClick={() => setShowDetails(!showDetails)}
            className="flex items-center justify-between w-full text-sm font-medium text-foreground hover:text-primary transition-colors"
          >
            <span>Health Check Details</span>
            <svg className={`w-4 h-4 transition-transform ${showDetails ? 'rotate-180' : ''}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
            </svg>
          </button>
          
          {showDetails && (
            <div className="mt-3 space-y-3">
              {healthChecks.map(check => (
                <div key={check.name} className="p-3 border border-border rounded-lg">
                  <div className="flex items-center justify-between mb-2">
                    <div className="flex items-center gap-2">
                      {getStatusIcon(check.status)}
                      <span className="text-sm font-medium text-foreground">{check.name}</span>
                    </div>
                    <span className={`text-sm font-bold ${getStatusColor(check.status)}`}>
                      {typeof check.value === 'number' ? check.value.toFixed(1) : check.value}
                      {check.unit}
                    </span>
                  </div>
                  <p className="text-xs text-muted-foreground mb-2">{check.description}</p>
                  {check.threshold && typeof check.value === 'number' && (
                    <div className="mb-2">
                      <div className="flex justify-between text-xs text-muted-foreground mb-1">
                        <span>Threshold: {check.threshold}{check.unit}</span>
                        <span>Current: {check.value.toFixed(1)}{check.unit}</span>
                      </div>
                      <Progress
                        value={(check.value / (check.threshold * 1.2)) * 100} // Scale to show threshold
                        variant={check.status === 'critical' ? 'error' : check.status === 'warning' ? 'warning' : 'success'}
                        size="xs"
                      />
                    </div>
                  )}
                  <p className="text-xs text-muted-foreground">
                    Last checked: {formatDateTime(check.lastChecked)}
                  </p>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Last Health Check */}
        <div className="pt-4 border-t border-border text-center">
          <p className="text-xs text-muted-foreground">
            Last health check: {formatDateTime(node.last_seen)}
          </p>
          <button className="mt-2 px-3 py-1.5 bg-primary text-primary-foreground rounded text-xs hover:bg-primary/90 transition-colors">
            Run Health Check
          </button>
        </div>
      </CardContent>
    </Card>
  )
}

export default NodeHealth