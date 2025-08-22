/**
 * ActivityFeedCard Component - Displays real-time activity feed
 */

import React, { useState } from 'react'
import { Card, CardHeader, CardContent, CardTitle, CardDescription } from '@/design-system/components/Card/Card'
import { Badge } from '@/design-system/components/Badge/Badge'
import { Button } from '@/design-system/components/Button/Button'
import { 
  Activity,
  RefreshCw,
  Server,
  Database,
  Play,
  AlertTriangle,
  Settings,
  CheckCircle,
  XCircle,
  Clock,
  Filter,
  MoreVertical
} from 'lucide-react'

interface ActivityItem {
  id: string
  type: 'node' | 'model' | 'task' | 'alert' | 'system'
  title: string
  description: string
  timestamp: string
  severity: 'info' | 'warning' | 'error' | 'success'
  metadata?: Record<string, any>
}

interface ActivityFeedCardProps {
  activities: ActivityItem[]
  onRefresh: () => void
  maxItems?: number
}

const ActivityFeedCard: React.FC<ActivityFeedCardProps> = ({
  activities,
  onRefresh,
  maxItems = 10
}) => {
  const [filter, setFilter] = useState<string>('all')
  const [isRefreshing, setIsRefreshing] = useState(false)

  const handleRefresh = async () => {
    setIsRefreshing(true)
    await new Promise(resolve => setTimeout(resolve, 1000)) // Simulate refresh delay
    onRefresh()
    setIsRefreshing(false)
  }

  const getActivityIcon = (type: string) => {
    switch (type) {
      case 'node':
        return <Server className="h-4 w-4" />
      case 'model':
        return <Database className="h-4 w-4" />
      case 'task':
        return <Play className="h-4 w-4" />
      case 'alert':
        return <AlertTriangle className="h-4 w-4" />
      case 'system':
        return <Settings className="h-4 w-4" />
      default:
        return <Activity className="h-4 w-4" />
    }
  }

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'success':
        return 'text-success-600 bg-success-50 border-success-200'
      case 'warning':
        return 'text-warning-600 bg-warning-50 border-warning-200'
      case 'error':
        return 'text-error-600 bg-error-50 border-error-200'
      default:
        return 'text-info-600 bg-info-50 border-info-200'
    }
  }

  const getSeverityBadge = (severity: string) => {
    switch (severity) {
      case 'success':
        return 'secondary'
      case 'warning':
        return 'warning'
      case 'error':
        return 'destructive'
      default:
        return 'secondary'
    }
  }

  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffMins = Math.floor(diffMs / (1000 * 60))
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60))
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))

    if (diffMins < 1) return 'Just now'
    if (diffMins < 60) return `${diffMins}m ago`
    if (diffHours < 24) return `${diffHours}h ago`
    return `${diffDays}d ago`
  }

  const filteredActivities = activities
    .filter(activity => filter === 'all' || activity.type === filter)
    .slice(0, maxItems)

  const filterOptions = [
    { value: 'all', label: 'All', count: activities.length },
    { value: 'node', label: 'Nodes', count: activities.filter(a => a.type === 'node').length },
    { value: 'model', label: 'Models', count: activities.filter(a => a.type === 'model').length },
    { value: 'task', label: 'Tasks', count: activities.filter(a => a.type === 'task').length },
    { value: 'alert', label: 'Alerts', count: activities.filter(a => a.type === 'alert').length }
  ]

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2">
              <Activity className="h-5 w-5" />
              Activity Feed
            </CardTitle>
            <CardDescription>Real-time system events and updates</CardDescription>
          </div>
          <Button
            variant="ghost"
            size="sm"
            onClick={handleRefresh}
            disabled={isRefreshing}
          >
            <RefreshCw className={`h-4 w-4 ${isRefreshing ? 'animate-spin' : ''}`} />
          </Button>
        </div>
      </CardHeader>

      <CardContent className="space-y-4">
        {/* Filter Tabs */}
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
              {option.count > 0 && (
                <Badge variant="secondary" className="ml-1 text-xs">
                  {option.count}
                </Badge>
              )}
            </Button>
          ))}
        </div>

        {/* Activity List */}
        <div className="space-y-3">
          {filteredActivities.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <Activity className="h-8 w-8 mx-auto mb-2 opacity-50" />
              <p className="text-sm">No activities found</p>
            </div>
          ) : (
            filteredActivities.map((activity) => (
              <div
                key={activity.id}
                className="flex items-start gap-3 p-3 rounded-lg border border-border hover:bg-muted/50 transition-colors"
              >
                <div className={`p-2 rounded-lg ${getSeverityColor(activity.severity)}`}>
                  {getActivityIcon(activity.type)}
                </div>
                
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between mb-1">
                    <h4 className="text-sm font-medium text-foreground truncate">
                      {activity.title}
                    </h4>
                    <Badge variant={getSeverityBadge(activity.severity) as any} className="text-xs ml-2">
                      {activity.severity}
                    </Badge>
                  </div>
                  
                  <p className="text-sm text-muted-foreground line-clamp-2">
                    {activity.description}
                  </p>
                  
                  <div className="flex items-center gap-2 mt-2">
                    <Clock className="h-3 w-3 text-muted-foreground" />
                    <span className="text-xs text-muted-foreground">
                      {formatTimestamp(activity.timestamp)}
                    </span>
                    
                    <Badge variant="outline" className="text-xs ml-auto">
                      {activity.type}
                    </Badge>
                  </div>
                </div>
              </div>
            ))
          )}
        </div>

        {/* Show more button */}
        {activities.length > maxItems && (
          <div className="text-center pt-3 border-t border-border">
            <Button variant="ghost" size="sm" className="text-xs">
              View all activities ({activities.length})
            </Button>
          </div>
        )}

        {/* Activity stats */}
        <div className="grid grid-cols-4 gap-2 pt-3 border-t border-border">
          <div className="text-center">
            <div className="text-sm font-semibold text-success-600">
              {activities.filter(a => a.severity === 'success').length}
            </div>
            <div className="text-xs text-muted-foreground">Success</div>
          </div>
          <div className="text-center">
            <div className="text-sm font-semibold text-info-600">
              {activities.filter(a => a.severity === 'info').length}
            </div>
            <div className="text-xs text-muted-foreground">Info</div>
          </div>
          <div className="text-center">
            <div className="text-sm font-semibold text-warning-600">
              {activities.filter(a => a.severity === 'warning').length}
            </div>
            <div className="text-xs text-muted-foreground">Warning</div>
          </div>
          <div className="text-center">
            <div className="text-sm font-semibold text-error-600">
              {activities.filter(a => a.severity === 'error').length}
            </div>
            <div className="text-xs text-muted-foreground">Error</div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

export default ActivityFeedCard