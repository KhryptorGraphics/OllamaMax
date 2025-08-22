/**
 * Alerts Panel Component
 * Displays active alerts and alert history with management features
 */

import React, { useState, useMemo } from 'react'
import {
  AlertTriangle,
  AlertCircle,
  Info,
  Check,
  X,
  Eye,
  EyeOff,
  Filter,
  Clock,
  User,
  MoreVertical,
  RefreshCw
} from 'lucide-react'
import { format, formatDistanceToNow } from 'date-fns'
import {
  MonitoringAlert,
  AlertSeverity,
  AlertCategory,
  AlertType
} from '../../types/monitoring'

interface AlertsPanelProps {
  alerts: MonitoringAlert[]
  onAcknowledge: (alertId: string) => void
  onResolve: (alertId: string) => void
  onRefresh?: () => void
  className?: string
}

interface AlertFilters {
  severity?: AlertSeverity
  category?: AlertCategory
  acknowledged?: boolean
  resolved?: boolean
  search?: string
}

const SEVERITY_COLORS = {
  info: {
    bg: 'bg-blue-50',
    border: 'border-blue-200',
    text: 'text-blue-800',
    icon: 'text-blue-500'
  },
  warning: {
    bg: 'bg-yellow-50',
    border: 'border-yellow-200',
    text: 'text-yellow-800',
    icon: 'text-yellow-500'
  },
  error: {
    bg: 'bg-red-50',
    border: 'border-red-200',
    text: 'text-red-800',
    icon: 'text-red-500'
  },
  critical: {
    bg: 'bg-red-100',
    border: 'border-red-300',
    text: 'text-red-900',
    icon: 'text-red-600'
  }
}

const getSeverityIcon = (severity: AlertSeverity) => {
  const iconClass = `w-5 h-5 ${SEVERITY_COLORS[severity].icon}`
  
  switch (severity) {
    case 'critical':
      return <AlertTriangle className={iconClass} />
    case 'error':
      return <AlertCircle className={iconClass} />
    case 'warning':
      return <AlertTriangle className={iconClass} />
    default:
      return <Info className={iconClass} />
  }
}

const AlertCard: React.FC<{
  alert: MonitoringAlert
  onAcknowledge: (alertId: string) => void
  onResolve: (alertId: string) => void
}> = ({ alert, onAcknowledge, onResolve }) => {
  const [showDetails, setShowDetails] = useState(false)
  const [showActions, setShowActions] = useState(false)
  
  const colors = SEVERITY_COLORS[alert.severity]
  const isResolved = !!alert.resolvedAt
  const isAcknowledged = alert.acknowledged
  
  const formatTimestamp = (timestamp: number) => {
    return format(new Date(timestamp), 'MMM dd, HH:mm:ss')
  }
  
  const getRelativeTime = (timestamp: number) => {
    return formatDistanceToNow(new Date(timestamp), { addSuffix: true })
  }
  
  return (
    <div className={`
      p-4 rounded-lg border-2 transition-all duration-200
      ${colors.bg} ${colors.border}
      ${isResolved ? 'opacity-60' : ''}
    `}>
      <div className="flex items-start justify-between">
        <div className="flex items-start space-x-3 flex-1">
          {getSeverityIcon(alert.severity)}
          
          <div className="flex-1 min-w-0">
            <div className="flex items-center space-x-2 mb-1">
              <h4 className={`font-medium ${colors.text}`}>
                {alert.message}
              </h4>
              {isAcknowledged && (
                <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800">
                  <Check className="w-3 h-3 mr-1" />
                  Acknowledged
                </span>
              )}
              {isResolved && (
                <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
                  <Check className="w-3 h-3 mr-1" />
                  Resolved
                </span>
              )}
            </div>
            
            <div className="flex flex-wrap items-center gap-2 text-sm text-gray-600 mb-2">
              <span className="inline-flex items-center">
                <Clock className="w-4 h-4 mr-1" />
                {getRelativeTime(alert.timestamp)}
              </span>
              <span>•</span>
              <span className="capitalize">{alert.source}</span>
              <span>•</span>
              <span className="capitalize">{alert.category}</span>
              {alert.threshold && (
                <>
                  <span>•</span>
                  <span>
                    {alert.currentValue.toFixed(2)} / {alert.threshold.toFixed(2)}
                  </span>
                </>
              )}
            </div>
            
            {alert.description && showDetails && (
              <p className="text-sm text-gray-700 mb-2">{alert.description}</p>
            )}
            
            {alert.tags.length > 0 && (
              <div className="flex flex-wrap gap-1 mb-2">
                {alert.tags.map((tag, index) => (
                  <span
                    key={index}
                    className="inline-flex items-center px-2 py-1 rounded-md text-xs font-medium bg-gray-100 text-gray-800"
                  >
                    {tag}
                  </span>
                ))}
              </div>
            )}
            
            {showDetails && (
              <div className="space-y-2 text-sm text-gray-600">
                <div>
                  <strong>Alert ID:</strong> {alert.id}
                </div>
                <div>
                  <strong>Created:</strong> {formatTimestamp(alert.timestamp)}
                </div>
                {alert.acknowledgedAt && (
                  <div>
                    <strong>Acknowledged:</strong> {formatTimestamp(alert.acknowledgedAt)}
                    {alert.acknowledgedBy && (
                      <span className="ml-1">by {alert.acknowledgedBy}</span>
                    )}
                  </div>
                )}
                {alert.resolvedAt && (
                  <div>
                    <strong>Resolved:</strong> {formatTimestamp(alert.resolvedAt)}
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
        
        <div className="flex items-center space-x-2">
          <button
            onClick={() => setShowDetails(!showDetails)}
            className="p-1 text-gray-400 hover:text-gray-600 transition-colors"
            title={showDetails ? 'Hide details' : 'Show details'}
          >
            {showDetails ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
          </button>
          
          <div className="relative">
            <button
              onClick={() => setShowActions(!showActions)}
              className="p-1 text-gray-400 hover:text-gray-600 transition-colors"
              title="Actions"
            >
              <MoreVertical className="w-4 h-4" />
            </button>
            
            {showActions && (
              <div className="absolute right-0 top-8 bg-white border border-gray-200 rounded-md shadow-lg z-10 min-w-[120px]">
                {!isAcknowledged && !isResolved && (
                  <button
                    onClick={() => {
                      onAcknowledge(alert.id)
                      setShowActions(false)
                    }}
                    className="block w-full text-left px-3 py-2 text-sm text-gray-700 hover:bg-gray-50"
                  >
                    Acknowledge
                  </button>
                )}
                
                {!isResolved && (
                  <button
                    onClick={() => {
                      onResolve(alert.id)
                      setShowActions(false)
                    }}
                    className="block w-full text-left px-3 py-2 text-sm text-gray-700 hover:bg-gray-50"
                  >
                    Resolve
                  </button>
                )}
                
                {alert.actions.map((action, index) => (
                  <button
                    key={index}
                    onClick={() => {
                      // Handle custom action
                      console.log('Custom action:', action)
                      setShowActions(false)
                    }}
                    className="block w-full text-left px-3 py-2 text-sm text-gray-700 hover:bg-gray-50"
                  >
                    {action.label}
                  </button>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

const FilterBar: React.FC<{
  filters: AlertFilters
  onFiltersChange: (filters: AlertFilters) => void
  alertCount: number
}> = ({ filters, onFiltersChange, alertCount }) => {
  const [showFilters, setShowFilters] = useState(false)
  
  return (
    <div className="bg-white border border-gray-200 rounded-lg p-4 mb-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <h3 className="text-lg font-medium text-gray-900">
            Alerts ({alertCount})
          </h3>
          
          <button
            onClick={() => setShowFilters(!showFilters)}
            className="inline-flex items-center px-3 py-1 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
          >
            <Filter className="w-4 h-4 mr-2" />
            Filters
          </button>
        </div>
        
        <div className="flex items-center space-x-2">
          {/* Quick filter buttons */}
          <button
            onClick={() => onFiltersChange({ ...filters, severity: 'critical' })}
            className={`px-3 py-1 rounded-md text-sm font-medium ${
              filters.severity === 'critical'
                ? 'bg-red-100 text-red-800'
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
            }`}
          >
            Critical
          </button>
          
          <button
            onClick={() => onFiltersChange({ ...filters, acknowledged: false })}
            className={`px-3 py-1 rounded-md text-sm font-medium ${
              filters.acknowledged === false
                ? 'bg-yellow-100 text-yellow-800'
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
            }`}
          >
            Unacknowledged
          </button>
          
          <button
            onClick={() => onFiltersChange({})}
            className="px-3 py-1 rounded-md text-sm font-medium bg-gray-100 text-gray-700 hover:bg-gray-200"
          >
            Clear
          </button>
        </div>
      </div>
      
      {showFilters && (
        <div className="mt-4 grid grid-cols-1 md:grid-cols-3 lg:grid-cols-5 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Severity
            </label>
            <select
              value={filters.severity || ''}
              onChange={(e) => onFiltersChange({
                ...filters,
                severity: e.target.value as AlertSeverity || undefined
              })}
              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
            >
              <option value="">All</option>
              <option value="critical">Critical</option>
              <option value="error">Error</option>
              <option value="warning">Warning</option>
              <option value="info">Info</option>
            </select>
          </div>
          
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Category
            </label>
            <select
              value={filters.category || ''}
              onChange={(e) => onFiltersChange({
                ...filters,
                category: e.target.value as AlertCategory || undefined
              })}
              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
            >
              <option value="">All</option>
              <option value="system">System</option>
              <option value="cluster">Cluster</option>
              <option value="model">Model</option>
              <option value="network">Network</option>
              <option value="security">Security</option>
              <option value="performance">Performance</option>
            </select>
          </div>
          
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Status
            </label>
            <select
              value={
                filters.acknowledged === true ? 'acknowledged' :
                filters.acknowledged === false ? 'unacknowledged' :
                filters.resolved === true ? 'resolved' :
                filters.resolved === false ? 'active' : ''
              }
              onChange={(e) => {
                const value = e.target.value
                const newFilters = { ...filters }
                delete newFilters.acknowledged
                delete newFilters.resolved
                
                if (value === 'acknowledged') newFilters.acknowledged = true
                else if (value === 'unacknowledged') newFilters.acknowledged = false
                else if (value === 'resolved') newFilters.resolved = true
                else if (value === 'active') newFilters.resolved = false
                
                onFiltersChange(newFilters)
              }}
              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
            >
              <option value="">All</option>
              <option value="active">Active</option>
              <option value="acknowledged">Acknowledged</option>
              <option value="resolved">Resolved</option>
              <option value="unacknowledged">Unacknowledged</option>
            </select>
          </div>
          
          <div className="md:col-span-2">
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Search
            </label>
            <input
              type="text"
              value={filters.search || ''}
              onChange={(e) => onFiltersChange({ ...filters, search: e.target.value })}
              placeholder="Search alerts..."
              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
            />
          </div>
        </div>
      )}
    </div>
  )
}

export const AlertsPanel: React.FC<AlertsPanelProps> = ({
  alerts,
  onAcknowledge,
  onResolve,
  onRefresh,
  className = ''
}) => {
  const [filters, setFilters] = useState<AlertFilters>({})
  
  const filteredAlerts = useMemo(() => {
    return alerts.filter(alert => {
      if (filters.severity && alert.severity !== filters.severity) return false
      if (filters.category && alert.category !== filters.category) return false
      if (filters.acknowledged !== undefined && alert.acknowledged !== filters.acknowledged) return false
      if (filters.resolved !== undefined && !!alert.resolvedAt !== filters.resolved) return false
      if (filters.search) {
        const search = filters.search.toLowerCase()
        return (
          alert.message.toLowerCase().includes(search) ||
          alert.source.toLowerCase().includes(search) ||
          alert.metric.toLowerCase().includes(search)
        )
      }
      return true
    })
  }, [alerts, filters])
  
  // Sort alerts by severity and timestamp
  const sortedAlerts = useMemo(() => {
    const severityOrder = { critical: 0, error: 1, warning: 2, info: 3 }
    
    return [...filteredAlerts].sort((a, b) => {
      // Unresolved alerts first
      if (!!a.resolvedAt !== !!b.resolvedAt) {
        return !!a.resolvedAt ? 1 : -1
      }
      
      // Then by severity
      const severityDiff = severityOrder[a.severity] - severityOrder[b.severity]
      if (severityDiff !== 0) return severityDiff
      
      // Finally by timestamp (newest first)
      return b.timestamp - a.timestamp
    })
  }, [filteredAlerts])
  
  return (
    <div className={`space-y-4 ${className}`}>
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold text-gray-900">Monitoring Alerts</h2>
        {onRefresh && (
          <button
            onClick={onRefresh}
            className="inline-flex items-center px-3 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
          >
            <RefreshCw className="w-4 h-4 mr-2" />
            Refresh
          </button>
        )}
      </div>
      
      <FilterBar
        filters={filters}
        onFiltersChange={setFilters}
        alertCount={filteredAlerts.length}
      />
      
      {sortedAlerts.length === 0 ? (
        <div className="text-center py-8 bg-white border border-gray-200 rounded-lg">
          <AlertCircle className="w-12 h-12 text-gray-400 mx-auto mb-4" />
          <div className="text-gray-500">
            {alerts.length === 0 ? 'No alerts found' : 'No alerts match your filters'}
          </div>
        </div>
      ) : (
        <div className="space-y-3">
          {sortedAlerts.map(alert => (
            <AlertCard
              key={alert.id}
              alert={alert}
              onAcknowledge={onAcknowledge}
              onResolve={onResolve}
            />
          ))}
        </div>
      )}
    </div>
  )
}