/**
 * AlertsCard Component - Displays system alerts and notifications
 */

import React, { useState } from 'react'
import { Card, CardHeader, CardContent, CardTitle, CardDescription } from '@/design-system/components/Card/Card'
import { Button } from '@/design-system/components/Button/Button'
import { Badge } from '@/design-system/components/Badge/Badge'
import { Alert } from '@/design-system/components/Alert/Alert'
import { 
  AlertTriangle,
  XCircle,
  Info,
  CheckCircle,
  X,
  Clock,
  Eye,
  EyeOff,
  MoreVertical,
  Filter,
  Bell,
  BellOff
} from 'lucide-react'

interface SystemAlert {
  id: string
  type: 'error' | 'warning' | 'info'
  title: string
  message: string
  timestamp: string
  acknowledged: boolean
  source: string
}

interface AlertsCardProps {
  alerts: SystemAlert[]
  onAcknowledge: (alertId: string) => void
  maxItems?: number
}

const AlertsCard: React.FC<AlertsCardProps> = ({
  alerts,
  onAcknowledge,
  maxItems = 5
}) => {
  const [filter, setFilter] = useState<string>('all')
  const [showAcknowledged, setShowAcknowledged] = useState(false)

  const getAlertIcon = (type: string) => {
    switch (type) {
      case 'error':
        return <XCircle className="h-4 w-4 text-error-600" />
      case 'warning':
        return <AlertTriangle className="h-4 w-4 text-warning-600" />
      case 'info':
        return <Info className="h-4 w-4 text-info-600" />
      default:
        return <Info className="h-4 w-4" />
    }
  }

  const getAlertVariant = (type: string) => {
    switch (type) {
      case 'error':
        return 'destructive'
      case 'warning':
        return 'warning'
      case 'info':
        return 'default'
      default:
        return 'default'
    }
  }

  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffMins = Math.floor(diffMs / (1000 * 60))
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60))

    if (diffMins < 1) return 'Just now'
    if (diffMins < 60) return `${diffMins}m ago`
    return `${diffHours}h ago`
  }

  const filteredAlerts = alerts
    .filter(alert => {
      if (!showAcknowledged && alert.acknowledged) return false
      if (filter === 'all') return true
      return alert.type === filter
    })
    .slice(0, maxItems)

  const unacknowledgedCount = alerts.filter(alert => !alert.acknowledged).length

  const filterOptions = [
    { value: 'all', label: 'All' },
    { value: 'error', label: 'Errors' },
    { value: 'warning', label: 'Warnings' },
    { value: 'info', label: 'Info' }
  ]

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2">
              <Bell className="h-5 w-5" />
              Recent Alerts
              {unacknowledgedCount > 0 && (
                <Badge variant="destructive" className="text-xs">
                  {unacknowledgedCount}
                </Badge>
              )}
            </CardTitle>
            <CardDescription>System alerts and notifications</CardDescription>
          </div>
          
          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setShowAcknowledged(!showAcknowledged)}
            >
              {showAcknowledged ? (
                <EyeOff className="h-4 w-4" />
              ) : (
                <Eye className="h-4 w-4" />
              )}
            </Button>
          </div>
        </div>
      </CardHeader>

      <CardContent className="space-y-4">
        {/* Filter Controls */}
        <div className="flex flex-wrap gap-2">
          {filterOptions.map((option) => (
            <Button
              key={option.value}
              variant={filter === option.value ? 'default' : 'ghost'}
              size="sm"
              onClick={() => setFilter(option.value)}
              className="text-xs"
            >
              {option.label}
            </Button>
          ))}
          
          <Button
            variant={showAcknowledged ? 'default' : 'ghost'}
            size="sm"
            onClick={() => setShowAcknowledged(!showAcknowledged)}
            className="text-xs ml-auto"
          >
            {showAcknowledged ? 'Hide' : 'Show'} Acknowledged
          </Button>
        </div>

        {/* Alerts List */}
        <div className="space-y-3">
          {filteredAlerts.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <CheckCircle className="h-8 w-8 mx-auto mb-2 text-success-600" />
              <p className="text-sm">No alerts to display</p>
              <p className="text-xs mt-1">System is running smoothly</p>
            </div>
          ) : (
            filteredAlerts.map((alert) => (
              <div
                key={alert.id}
                className={`border rounded-lg p-3 transition-all ${
                  alert.acknowledged 
                    ? 'bg-muted/30 opacity-60' 
                    : 'bg-background hover:shadow-sm'
                }`}
              >
                <div className="flex items-start gap-3">
                  <div className="flex-shrink-0 mt-0.5">
                    {getAlertIcon(alert.type)}
                  </div>
                  
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between mb-1">
                      <h4 className="text-sm font-medium text-foreground">
                        {alert.title}
                      </h4>
                      <div className="flex items-center gap-2">
                        <Badge variant={getAlertVariant(alert.type) as any} className="text-xs">
                          {alert.type}
                        </Badge>
                        {alert.acknowledged && (
                          <Badge variant="secondary" className="text-xs">
                            Acknowledged
                          </Badge>
                        )}
                      </div>
                    </div>
                    
                    <p className="text-sm text-muted-foreground mb-2">
                      {alert.message}
                    </p>
                    
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-4 text-xs text-muted-foreground">
                        <div className="flex items-center gap-1">
                          <Clock className="h-3 w-3" />
                          {formatTimestamp(alert.timestamp)}
                        </div>
                        <div>
                          Source: {alert.source}
                        </div>
                      </div>
                      
                      {!alert.acknowledged && (
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => onAcknowledge(alert.id)}
                          className="text-xs"
                        >
                          <CheckCircle className="h-3 w-3 mr-1" />
                          Acknowledge
                        </Button>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            ))
          )}
        </div>

        {/* Alert Summary */}
        <div className="grid grid-cols-3 gap-3 pt-3 border-t border-border">
          <div className="text-center">
            <div className="text-sm font-semibold text-error-600">
              {alerts.filter(a => a.type === 'error' && !a.acknowledged).length}
            </div>
            <div className="text-xs text-muted-foreground">Errors</div>
          </div>
          <div className="text-center">
            <div className="text-sm font-semibold text-warning-600">
              {alerts.filter(a => a.type === 'warning' && !a.acknowledged).length}
            </div>
            <div className="text-xs text-muted-foreground">Warnings</div>
          </div>
          <div className="text-center">
            <div className="text-sm font-semibold text-info-600">
              {alerts.filter(a => a.type === 'info' && !a.acknowledged).length}
            </div>
            <div className="text-xs text-muted-foreground">Info</div>
          </div>
        </div>

        {/* Show more button */}
        {alerts.length > maxItems && (
          <div className="text-center pt-3 border-t border-border">
            <Button variant="ghost" size="sm" className="text-xs">
              View all alerts ({alerts.length})
            </Button>
          </div>
        )}

        {/* Alert Actions */}
        {unacknowledgedCount > 1 && (
          <div className="pt-3 border-t border-border">
            <Button
              variant="secondary"
              size="sm"
              className="w-full text-xs"
              onClick={() => {
                alerts.forEach(alert => {
                  if (!alert.acknowledged) {
                    onAcknowledge(alert.id)
                  }
                })
              }}
            >
              Acknowledge All ({unacknowledgedCount})
            </Button>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

export default AlertsCard