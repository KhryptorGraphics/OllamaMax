/**
 * QuickActionsCard Component - Provides quick access to common operations
 */

import React, { useState } from 'react'
import { Card, CardHeader, CardContent, CardTitle, CardDescription } from '@/design-system/components/Card/Card'
import { Button } from '@/design-system/components/Button/Button'
import { Badge } from '@/design-system/components/Badge/Badge'
import { 
  Zap,
  Plus,
  Download,
  Upload,
  Trash2,
  Pause,
  Play,
  RotateCcw,
  Settings,
  Database,
  Server,
  RefreshCw,
  AlertTriangle,
  CheckCircle,
  ExternalLink
} from 'lucide-react'

interface QuickAction {
  id: string
  title: string
  description: string
  icon: React.ReactNode
  variant: 'default' | 'secondary' | 'destructive' | 'ghost'
  disabled?: boolean
  badge?: {
    text: string
    variant: 'default' | 'secondary' | 'destructive' | 'warning'
  }
  onClick: () => void
}

const QuickActionsCard: React.FC = () => {
  const [isExecuting, setIsExecuting] = useState<string | null>(null)

  const executeAction = async (actionId: string, action: () => void) => {
    setIsExecuting(actionId)
    try {
      // Simulate async operation
      await new Promise(resolve => setTimeout(resolve, 1500))
      action()
    } catch (error) {
      console.error('Action failed:', error)
    } finally {
      setIsExecuting(null)
    }
  }

  const quickActions: QuickAction[] = [
    {
      id: 'add-node',
      title: 'Add Node',
      description: 'Add a new worker node to the cluster',
      icon: <Plus className="h-4 w-4" />,
      variant: 'default',
      onClick: () => console.log('Add node')
    },
    {
      id: 'sync-models',
      title: 'Sync Models',
      description: 'Synchronize models across all nodes',
      icon: <RefreshCw className="h-4 w-4" />,
      variant: 'secondary',
      badge: { text: '2 pending', variant: 'warning' },
      onClick: () => console.log('Sync models')
    },
    {
      id: 'download-model',
      title: 'Download Model',
      description: 'Download a new model from registry',
      icon: <Download className="h-4 w-4" />,
      variant: 'secondary',
      onClick: () => console.log('Download model')
    },
    {
      id: 'backup-data',
      title: 'Backup Data',
      description: 'Create a backup of cluster data',
      icon: <Database className="h-4 w-4" />,
      variant: 'secondary',
      onClick: () => console.log('Backup data')
    },
    {
      id: 'restart-services',
      title: 'Restart Services',
      description: 'Restart all cluster services',
      icon: <RotateCcw className="h-4 w-4" />,
      variant: 'destructive',
      badge: { text: 'High Impact', variant: 'destructive' },
      onClick: () => console.log('Restart services')
    },
    {
      id: 'cluster-settings',
      title: 'Settings',
      description: 'Configure cluster settings',
      icon: <Settings className="h-4 w-4" />,
      variant: 'ghost',
      onClick: () => console.log('Open settings')
    }
  ]

  const systemActions = [
    {
      id: 'health-check',
      title: 'Run Health Check',
      description: 'Perform comprehensive system health check',
      icon: <CheckCircle className="h-4 w-4" />,
      variant: 'secondary' as const,
      onClick: () => console.log('Health check')
    },
    {
      id: 'view-logs',
      title: 'View Logs',
      description: 'Open system logs dashboard',
      icon: <ExternalLink className="h-4 w-4" />,
      variant: 'ghost' as const,
      onClick: () => console.log('View logs')
    }
  ]

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Zap className="h-5 w-5" />
          Quick Actions
        </CardTitle>
        <CardDescription>Common operations and shortcuts</CardDescription>
      </CardHeader>

      <CardContent className="space-y-6">
        {/* Primary Actions */}
        <div className="space-y-3">
          <h4 className="text-sm font-medium text-foreground">
            Primary Actions
          </h4>
          <div className="grid gap-2">
            {quickActions.map((action) => (
              <Button
                key={action.id}
                variant={action.variant}
                size="sm"
                disabled={action.disabled || isExecuting === action.id}
                onClick={() => executeAction(action.id, action.onClick)}
                className="w-full justify-start h-auto p-3"
              >
                <div className="flex items-center gap-3 w-full">
                  <div className="flex-shrink-0">
                    {isExecuting === action.id ? (
                      <RefreshCw className="h-4 w-4 animate-spin" />
                    ) : (
                      action.icon
                    )}
                  </div>
                  
                  <div className="flex-1 text-left">
                    <div className="flex items-center gap-2">
                      <span className="font-medium text-sm">
                        {action.title}
                      </span>
                      {action.badge && (
                        <Badge variant={action.badge.variant} className="text-xs">
                          {action.badge.text}
                        </Badge>
                      )}
                    </div>
                    <p className="text-xs text-muted-foreground mt-0.5">
                      {action.description}
                    </p>
                  </div>
                </div>
              </Button>
            ))}
          </div>
        </div>

        {/* System Actions */}
        <div className="space-y-3">
          <h4 className="text-sm font-medium text-foreground">
            System Actions
          </h4>
          <div className="grid gap-2">
            {systemActions.map((action) => (
              <Button
                key={action.id}
                variant={action.variant}
                size="sm"
                disabled={isExecuting === action.id}
                onClick={() => executeAction(action.id, action.onClick)}
                className="w-full justify-start h-auto p-3"
              >
                <div className="flex items-center gap-3 w-full">
                  <div className="flex-shrink-0">
                    {isExecuting === action.id ? (
                      <RefreshCw className="h-4 w-4 animate-spin" />
                    ) : (
                      action.icon
                    )}
                  </div>
                  
                  <div className="flex-1 text-left">
                    <span className="font-medium text-sm">
                      {action.title}
                    </span>
                    <p className="text-xs text-muted-foreground mt-0.5">
                      {action.description}
                    </p>
                  </div>
                </div>
              </Button>
            ))}
          </div>
        </div>

        {/* Quick Stats */}
        <div className="pt-3 border-t border-border">
          <div className="grid grid-cols-2 gap-3 text-center">
            <div>
              <div className="text-lg font-semibold text-foreground">12</div>
              <div className="text-xs text-muted-foreground">Actions Today</div>
            </div>
            <div>
              <div className="text-lg font-semibold text-foreground">3m</div>
              <div className="text-xs text-muted-foreground">Avg Duration</div>
            </div>
          </div>
        </div>

        {/* Emergency Actions */}
        <div className="pt-3 border-t border-border">
          <div className="flex items-center gap-2 mb-2">
            <AlertTriangle className="h-4 w-4 text-warning-600" />
            <span className="text-sm font-medium text-foreground">Emergency</span>
          </div>
          
          <Button
            variant="destructive"
            size="sm"
            className="w-full"
            disabled={isExecuting === 'emergency-stop'}
            onClick={() => executeAction('emergency-stop', () => console.log('Emergency stop'))}
          >
            {isExecuting === 'emergency-stop' ? (
              <RefreshCw className="h-4 w-4 animate-spin mr-2" />
            ) : (
              <Pause className="h-4 w-4 mr-2" />
            )}
            Emergency Stop
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}

export default QuickActionsCard