import React, { useState, useCallback, useEffect } from 'react'
import {
  Shield, AlertTriangle, CheckCircle, XCircle, Eye, Download,
  Calendar, Filter, Search, RefreshCw, TrendingUp, TrendingDown,
  Activity, Users, Lock, Key, Globe, Monitor, Smartphone, Mail
} from 'lucide-react'
import { apiClient } from '@/services/api/client'
import type { SecurityEvent } from '@/types/auth'

interface SecurityAuditPanelProps {
  className?: string
}

interface SecurityMetrics {
  totalEvents: number
  criticalEvents: number
  failedLogins: number
  successfulLogins: number
  mfaEvents: number
  passwordChanges: number
  suspiciousActivity: number
  blockedIPs: number
}

interface SecurityTrend {
  date: string
  events: number
  severity: 'low' | 'medium' | 'high' | 'critical'
}

interface AuditFilter {
  timeRange: '1h' | '24h' | '7d' | '30d' | 'custom'
  severity: 'all' | 'low' | 'medium' | 'high' | 'critical'
  eventType: 'all' | 'authentication' | 'authorization' | 'security' | 'system'
  search: string
  startDate?: string
  endDate?: string
}

export const SecurityAuditPanel: React.FC<SecurityAuditPanelProps> = ({ className = '' }) => {
  const [events, setEvents] = useState<SecurityEvent[]>([])
  const [metrics, setMetrics] = useState<SecurityMetrics | null>(null)
  const [trends, setTrends] = useState<SecurityTrend[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [filter, setFilter] = useState<AuditFilter>({
    timeRange: '24h',
    severity: 'all',
    eventType: 'all',
    search: ''
  })

  useEffect(() => {
    loadSecurityData()
  }, [filter])

  const loadSecurityData = useCallback(async () => {
    setIsLoading(true)
    setError(null)

    try {
      // Load security events
      const eventsResponse = await apiClient.request<SecurityEvent[]>('/admin/security/events', {
        method: 'GET',
        query: {
          timeRange: filter.timeRange,
          severity: filter.severity === 'all' ? undefined : filter.severity,
          type: filter.eventType === 'all' ? undefined : filter.eventType,
          search: filter.search || undefined,
          startDate: filter.startDate,
          endDate: filter.endDate
        }
      })

      if (eventsResponse.success) {
        setEvents(eventsResponse.data)
      }

      // Load security metrics
      const metricsResponse = await apiClient.request<SecurityMetrics>('/admin/security/metrics', {
        method: 'GET',
        query: { timeRange: filter.timeRange }
      })

      if (metricsResponse.success) {
        setMetrics(metricsResponse.data)
      }

      // Load security trends
      const trendsResponse = await apiClient.request<SecurityTrend[]>('/admin/security/trends', {
        method: 'GET',
        query: { timeRange: filter.timeRange }
      })

      if (trendsResponse.success) {
        setTrends(trendsResponse.data)
      }

    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load security data')
    } finally {
      setIsLoading(false)
    }
  }, [filter])

  const handleExport = useCallback(async () => {
    try {
      const response = await apiClient.request('/admin/security/export', {
        method: 'POST',
        body: {
          timeRange: filter.timeRange,
          severity: filter.severity === 'all' ? undefined : filter.severity,
          type: filter.eventType === 'all' ? undefined : filter.eventType,
          format: 'csv'
        }
      })

      if (response.success && response.data) {
        // Create and download file
        const blob = new Blob([response.data], { type: 'text/csv' })
        const url = window.URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = `security-audit-${new Date().toISOString().split('T')[0]}.csv`
        document.body.appendChild(a)
        a.click()
        window.URL.revokeObjectURL(url)
        document.body.removeChild(a)
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to export security data')
    }
  }, [filter])

  const getSeverityIcon = (severity: string) => {
    switch (severity) {
      case 'critical':
        return <XCircle className="w-4 h-4 text-red-600" />
      case 'high':
        return <AlertTriangle className="w-4 h-4 text-orange-600" />
      case 'medium':
        return <Eye className="w-4 h-4 text-yellow-600" />
      case 'low':
        return <CheckCircle className="w-4 h-4 text-blue-600" />
      default:
        return <Shield className="w-4 h-4 text-gray-400" />
    }
  }

  const getSeverityBadge = (severity: string) => {
    const baseClasses = "px-2 py-1 text-xs font-medium rounded-full"
    switch (severity) {
      case 'critical':
        return `${baseClasses} bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200`
      case 'high':
        return `${baseClasses} bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200`
      case 'medium':
        return `${baseClasses} bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200`
      case 'low':
        return `${baseClasses} bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200`
      default:
        return `${baseClasses} bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200`
    }
  }

  const getEventTypeIcon = (type: string) => {
    switch (type) {
      case 'login_success':
      case 'login_failure':
        return <Key className="w-4 h-4" />
      case 'mfa_enabled':
      case 'mfa_disabled':
        return <Smartphone className="w-4 h-4" />
      case 'password_change':
        return <Lock className="w-4 h-4" />
      case 'suspicious_activity':
        return <AlertTriangle className="w-4 h-4" />
      case 'account_locked':
        return <XCircle className="w-4 h-4" />
      default:
        return <Activity className="w-4 h-4" />
    }
  }

  const getMetricTrend = (current: number, previous: number) => {
    if (previous === 0) return null
    const change = ((current - previous) / previous) * 100
    return {
      value: Math.abs(change).toFixed(1),
      isPositive: change > 0,
      icon: change > 0 ? TrendingUp : TrendingDown
    }
  }

  const formatEventType = (type: string) => {
    return type.split('_').map(word => 
      word.charAt(0).toUpperCase() + word.slice(1)
    ).join(' ')
  }

  return (
    <div className={`w-full ${className}`}>
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700">
        {/* Header */}
        <div className="p-6 border-b border-gray-200 dark:border-gray-700">
          <div className="flex items-center justify-between">
            <div className="flex items-center">
              <Shield className="w-6 h-6 text-red-600 dark:text-red-400 mr-3" />
              <div>
                <h1 className="text-xl font-semibold text-gray-900 dark:text-white">
                  Security Audit Panel
                </h1>
                <p className="text-sm text-gray-600 dark:text-gray-400">
                  Monitor security events and analyze threats
                </p>
              </div>
            </div>
            <div className="flex space-x-2">
              <button
                onClick={loadSecurityData}
                disabled={isLoading}
                className="flex items-center px-3 py-2 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
              >
                <RefreshCw className={`w-4 h-4 mr-2 ${isLoading ? 'animate-spin' : ''}`} />
                Refresh
              </button>
              <button
                onClick={handleExport}
                className="flex items-center px-3 py-2 text-sm border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700"
              >
                <Download className="w-4 h-4 mr-2" />
                Export
              </button>
            </div>
          </div>
        </div>

        {/* Error Message */}
        {error && (
          <div className="mx-6 mt-4 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md">
            <div className="flex items-center">
              <AlertTriangle className="w-4 h-4 text-red-500 mr-2" />
              <span className="text-sm text-red-700 dark:text-red-400">{error}</span>
            </div>
          </div>
        )}

        {/* Security Metrics */}
        {metrics && (
          <div className="p-6 border-b border-gray-200 dark:border-gray-700">
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div className="bg-red-50 dark:bg-red-900/20 rounded-lg p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-red-600 dark:text-red-400">
                      Critical Events
                    </p>
                    <p className="text-2xl font-bold text-red-700 dark:text-red-300">
                      {metrics.criticalEvents}
                    </p>
                  </div>
                  <XCircle className="w-8 h-8 text-red-600 dark:text-red-400" />
                </div>
              </div>

              <div className="bg-orange-50 dark:bg-orange-900/20 rounded-lg p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-orange-600 dark:text-orange-400">
                      Failed Logins
                    </p>
                    <p className="text-2xl font-bold text-orange-700 dark:text-orange-300">
                      {metrics.failedLogins}
                    </p>
                  </div>
                  <Key className="w-8 h-8 text-orange-600 dark:text-orange-400" />
                </div>
              </div>

              <div className="bg-green-50 dark:bg-green-900/20 rounded-lg p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-green-600 dark:text-green-400">
                      Successful Logins
                    </p>
                    <p className="text-2xl font-bold text-green-700 dark:text-green-300">
                      {metrics.successfulLogins}
                    </p>
                  </div>
                  <CheckCircle className="w-8 h-8 text-green-600 dark:text-green-400" />
                </div>
              </div>

              <div className="bg-blue-50 dark:bg-blue-900/20 rounded-lg p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-blue-600 dark:text-blue-400">
                      MFA Events
                    </p>
                    <p className="text-2xl font-bold text-blue-700 dark:text-blue-300">
                      {metrics.mfaEvents}
                    </p>
                  </div>
                  <Smartphone className="w-8 h-8 text-blue-600 dark:text-blue-400" />
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Filters */}
        <div className="p-6 border-b border-gray-200 dark:border-gray-700">
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            {/* Time Range */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Time Range
              </label>
              <select
                value={filter.timeRange}
                onChange={(e) => setFilter(prev => ({ ...prev, timeRange: e.target.value as any }))}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
              >
                <option value="1h">Last Hour</option>
                <option value="24h">Last 24 Hours</option>
                <option value="7d">Last 7 Days</option>
                <option value="30d">Last 30 Days</option>
                <option value="custom">Custom Range</option>
              </select>
            </div>

            {/* Severity */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Severity
              </label>
              <select
                value={filter.severity}
                onChange={(e) => setFilter(prev => ({ ...prev, severity: e.target.value as any }))}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
              >
                <option value="all">All Severities</option>
                <option value="critical">Critical</option>
                <option value="high">High</option>
                <option value="medium">Medium</option>
                <option value="low">Low</option>
              </select>
            </div>

            {/* Event Type */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Event Type
              </label>
              <select
                value={filter.eventType}
                onChange={(e) => setFilter(prev => ({ ...prev, eventType: e.target.value as any }))}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
              >
                <option value="all">All Types</option>
                <option value="authentication">Authentication</option>
                <option value="authorization">Authorization</option>
                <option value="security">Security</option>
                <option value="system">System</option>
              </select>
            </div>

            {/* Search */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Search
              </label>
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
                <input
                  type="text"
                  placeholder="Search events..."
                  value={filter.search}
                  onChange={(e) => setFilter(prev => ({ ...prev, search: e.target.value }))}
                  className="w-full pl-10 pr-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                />
              </div>
            </div>
          </div>

          {/* Custom Date Range */}
          {filter.timeRange === 'custom' && (
            <div className="grid grid-cols-2 gap-4 mt-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Start Date
                </label>
                <input
                  type="datetime-local"
                  value={filter.startDate}
                  onChange={(e) => setFilter(prev => ({ ...prev, startDate: e.target.value }))}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  End Date
                </label>
                <input
                  type="datetime-local"
                  value={filter.endDate}
                  onChange={(e) => setFilter(prev => ({ ...prev, endDate: e.target.value }))}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                />
              </div>
            </div>
          )}
        </div>

        {/* Security Events */}
        <div className="overflow-x-auto">
          {isLoading ? (
            <div className="text-center py-12">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
            </div>
          ) : events.length === 0 ? (
            <div className="text-center py-12">
              <Shield className="w-12 h-12 text-gray-400 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">
                No Security Events
              </h3>
              <p className="text-gray-600 dark:text-gray-400">
                No security events found for the selected criteria
              </p>
            </div>
          ) : (
            <table className="w-full">
              <thead className="bg-gray-50 dark:bg-gray-700">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Event
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Severity
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    User
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    IP Address
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Location
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Timestamp
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
                {events.map((event) => (
                  <tr key={event.id} className="hover:bg-gray-50 dark:hover:bg-gray-700">
                    <td className="px-6 py-4">
                      <div className="flex items-center">
                        <div className="mr-3">
                          {getEventTypeIcon(event.type)}
                        </div>
                        <div>
                          <div className="text-sm font-medium text-gray-900 dark:text-white">
                            {formatEventType(event.type)}
                          </div>
                          <div className="text-sm text-gray-500 dark:text-gray-400">
                            {event.description}
                          </div>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center">
                        {getSeverityIcon(event.severity)}
                        <span className={`ml-2 ${getSeverityBadge(event.severity)}`}>
                          {event.severity}
                        </span>
                      </div>
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-900 dark:text-white">
                      {event.userId || 'System'}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-500 dark:text-gray-400">
                      {event.ipAddress}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-500 dark:text-gray-400">
                      {event.location ? `${event.location.city}, ${event.location.country}` : 'Unknown'}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-500 dark:text-gray-400">
                      {new Date(event.timestamp).toLocaleString()}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>

        {/* Pagination */}
        <div className="px-6 py-3 border-t border-gray-200 dark:border-gray-700">
          <div className="flex items-center justify-between">
            <div className="text-sm text-gray-700 dark:text-gray-300">
              Showing {events.length} security events
            </div>
            <div className="flex space-x-2">
              <button className="px-3 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded hover:bg-gray-50 dark:hover:bg-gray-700">
                Previous
              </button>
              <button className="px-3 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded hover:bg-gray-50 dark:hover:bg-gray-700">
                Next
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default SecurityAuditPanel