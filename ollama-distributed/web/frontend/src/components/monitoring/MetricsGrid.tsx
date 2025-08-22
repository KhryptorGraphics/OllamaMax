/**
 * Metrics Grid Component
 * Displays multiple metric cards in a responsive grid layout
 */

import React from 'react'
import { TrendingUp, TrendingDown, Minus, AlertTriangle, Check, Activity } from 'lucide-react'
import { MetricValue, SystemMetrics, ClusterMetrics } from '../../types/monitoring'

interface MetricsGridProps {
  systemMetrics?: SystemMetrics
  clusterMetrics?: ClusterMetrics
  className?: string
}

interface MetricCardProps {
  title: string
  metric: MetricValue
  icon?: React.ReactNode
  color?: 'blue' | 'green' | 'yellow' | 'red' | 'purple'
  size?: 'sm' | 'md' | 'lg'
}

const MetricCard: React.FC<MetricCardProps> = ({
  title,
  metric,
  icon,
  color = 'blue',
  size = 'md'
}) => {
  const getTrendIcon = () => {
    switch (metric.trend) {
      case 'up':
        return <TrendingUp className="w-4 h-4 text-green-500" />
      case 'down':
        return <TrendingDown className="w-4 h-4 text-red-500" />
      default:
        return <Minus className="w-4 h-4 text-gray-500" />
    }
  }
  
  const getStatusColor = () => {
    if (!metric.threshold) return color
    
    const usage = (metric.current / metric.threshold) * 100
    if (usage >= 90) return 'red'
    if (usage >= 75) return 'yellow'
    return 'green'
  }
  
  const statusColor = getStatusColor()
  
  const colorClasses = {
    blue: 'border-blue-200 bg-blue-50',
    green: 'border-green-200 bg-green-50',
    yellow: 'border-yellow-200 bg-yellow-50',
    red: 'border-red-200 bg-red-50',
    purple: 'border-purple-200 bg-purple-50'
  }
  
  const sizeClasses = {
    sm: 'p-3',
    md: 'p-4',
    lg: 'p-6'
  }
  
  const formatValue = (value: number) => {
    if (metric.unit === '%') {
      return `${value.toFixed(1)}%`
    }
    if (metric.unit === 'bytes') {
      const units = ['B', 'KB', 'MB', 'GB', 'TB']
      let unitIndex = 0
      let size = value
      
      while (size >= 1024 && unitIndex < units.length - 1) {
        size /= 1024
        unitIndex++
      }
      
      return `${size.toFixed(1)} ${units[unitIndex]}`
    }
    if (metric.unit === 'ms') {
      return `${value.toFixed(0)}ms`
    }
    
    return `${value.toFixed(2)} ${metric.unit}`
  }
  
  return (
    <div className={`
      rounded-lg border-2 transition-all duration-200 hover:shadow-md
      ${colorClasses[statusColor]} ${sizeClasses[size]}
    `}>
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center space-x-2">
          {icon}
          <h3 className="text-sm font-medium text-gray-700">{title}</h3>
        </div>
        {getTrendIcon()}
      </div>
      
      <div className="space-y-2">
        <div className="text-2xl font-bold text-gray-900">
          {formatValue(metric.current)}
        </div>
        
        <div className="flex justify-between text-xs text-gray-600">
          <span>Avg: {formatValue(metric.average)}</span>
          <span>Peak: {formatValue(metric.peak)}</span>
        </div>
        
        {metric.threshold && (
          <div className="mt-2">
            <div className="flex justify-between text-xs text-gray-500 mb-1">
              <span>Usage</span>
              <span>{((metric.current / metric.threshold) * 100).toFixed(1)}%</span>
            </div>
            <div className="w-full bg-gray-200 rounded-full h-2">
              <div
                className={`h-2 rounded-full transition-all duration-300 ${
                  statusColor === 'red' ? 'bg-red-500' :
                  statusColor === 'yellow' ? 'bg-yellow-500' :
                  'bg-green-500'
                }`}
                style={{
                  width: `${Math.min((metric.current / metric.threshold) * 100, 100)}%`
                }}
              />
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

const NetworkCard: React.FC<{ networkMetrics: any }> = ({ networkMetrics }) => {
  return (
    <div className="col-span-1 md:col-span-2 p-4 bg-white rounded-lg border-2 border-gray-200">
      <div className="flex items-center space-x-2 mb-4">
        <Activity className="w-5 h-5 text-blue-500" />
        <h3 className="text-lg font-semibold text-gray-900">Network</h3>
      </div>
      
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div className="text-center">
          <div className="text-sm text-gray-500">Throughput In</div>
          <div className="text-lg font-bold text-green-600">
            {(networkMetrics.throughput.in / 1024 / 1024).toFixed(1)} MB/s
          </div>
        </div>
        
        <div className="text-center">
          <div className="text-sm text-gray-500">Throughput Out</div>
          <div className="text-lg font-bold text-blue-600">
            {(networkMetrics.throughput.out / 1024 / 1024).toFixed(1)} MB/s
          </div>
        </div>
        
        <div className="text-center">
          <div className="text-sm text-gray-500">Avg Latency</div>
          <div className="text-lg font-bold text-purple-600">
            {networkMetrics.latency.avg.toFixed(1)}ms
          </div>
        </div>
        
        <div className="text-center">
          <div className="text-sm text-gray-500">Errors</div>
          <div className={`text-lg font-bold ${
            networkMetrics.errors > 0 ? 'text-red-600' : 'text-green-600'
          }`}>
            {networkMetrics.errors}
          </div>
        </div>
      </div>
    </div>
  )
}

const ClusterStatusCard: React.FC<{ clusterMetrics: ClusterMetrics }> = ({ clusterMetrics }) => {
  const getHealthColor = () => {
    switch (clusterMetrics.health.status) {
      case 'healthy': return 'text-green-600'
      case 'degraded': return 'text-yellow-600'
      case 'critical': return 'text-red-600'
      default: return 'text-gray-600'
    }
  }
  
  const getHealthIcon = () => {
    switch (clusterMetrics.health.status) {
      case 'healthy': return <Check className="w-5 h-5 text-green-500" />
      case 'degraded': return <AlertTriangle className="w-5 h-5 text-yellow-500" />
      case 'critical': return <AlertTriangle className="w-5 h-5 text-red-500" />
      default: return <Minus className="w-5 h-5 text-gray-500" />
    }
  }
  
  return (
    <div className="col-span-1 md:col-span-3 p-4 bg-white rounded-lg border-2 border-gray-200">
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center space-x-2">
          {getHealthIcon()}
          <h3 className="text-lg font-semibold text-gray-900">Cluster Status</h3>
        </div>
        <span className={`text-sm font-medium ${getHealthColor()}`}>
          {clusterMetrics.health.status.toUpperCase()}
        </span>
      </div>
      
      <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
        <div className="text-center">
          <div className="text-sm text-gray-500">Active Nodes</div>
          <div className="text-xl font-bold text-green-600">
            {clusterMetrics.activeNodes}/{clusterMetrics.totalNodes}
          </div>
        </div>
        
        <div className="text-center">
          <div className="text-sm text-gray-500">Models</div>
          <div className="text-xl font-bold text-blue-600">
            {clusterMetrics.totalModels}
          </div>
        </div>
        
        <div className="text-center">
          <div className="text-sm text-gray-500">Active Tasks</div>
          <div className="text-xl font-bold text-purple-600">
            {clusterMetrics.activeTasks}
          </div>
        </div>
        
        <div className="text-center">
          <div className="text-sm text-gray-500">Avg Response</div>
          <div className="text-xl font-bold text-orange-600">
            {clusterMetrics.averageResponseTime.toFixed(0)}ms
          </div>
        </div>
        
        <div className="text-center">
          <div className="text-sm text-gray-500">Error Rate</div>
          <div className={`text-xl font-bold ${
            clusterMetrics.errorRate > 5 ? 'text-red-600' : 
            clusterMetrics.errorRate > 1 ? 'text-yellow-600' : 'text-green-600'
          }`}>
            {clusterMetrics.errorRate.toFixed(2)}%
          </div>
        </div>
      </div>
      
      {clusterMetrics.health.issues.length > 0 && (
        <div className="mt-4 p-3 bg-yellow-50 border border-yellow-200 rounded-md">
          <div className="text-sm font-medium text-yellow-800 mb-1">Issues:</div>
          <ul className="text-sm text-yellow-700 space-y-1">
            {clusterMetrics.health.issues.map((issue, index) => (
              <li key={index}>• {issue}</li>
            ))}
          </ul>
        </div>
      )}
    </div>
  )
}

export const MetricsGrid: React.FC<MetricsGridProps> = ({
  systemMetrics,
  clusterMetrics,
  className = ''
}) => {
  if (!systemMetrics && !clusterMetrics) {
    return (
      <div className={`text-center py-8 ${className}`}>
        <div className="text-gray-500">No metrics data available</div>
      </div>
    )
  }
  
  return (
    <div className={`space-y-6 ${className}`}>
      {/* System Metrics */}
      {systemMetrics && (
        <div>
          <h2 className="text-lg font-semibold text-gray-900 mb-4">System Metrics</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <MetricCard
              title="CPU Usage"
              metric={systemMetrics.cpu}
              icon={<Activity className="w-4 h-4 text-blue-500" />}
              color="blue"
            />
            <MetricCard
              title="Memory Usage"
              metric={systemMetrics.memory}
              icon={<Activity className="w-4 h-4 text-green-500" />}
              color="green"
            />
            <MetricCard
              title="Disk Usage"
              metric={systemMetrics.disk}
              icon={<Activity className="w-4 h-4 text-purple-500" />}
              color="purple"
            />
            {systemMetrics.temperature && (
              <MetricCard
                title="Temperature"
                metric={systemMetrics.temperature}
                icon={<Activity className="w-4 h-4 text-red-500" />}
                color="red"
              />
            )}
          </div>
          
          {/* Network Card */}
          <div className="mt-4">
            <NetworkCard networkMetrics={systemMetrics.network} />
          </div>
        </div>
      )}
      
      {/* Cluster Metrics */}
      {clusterMetrics && (
        <div>
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Cluster Status</h2>
          <ClusterStatusCard clusterMetrics={clusterMetrics} />
        </div>
      )}
    </div>
  )
}