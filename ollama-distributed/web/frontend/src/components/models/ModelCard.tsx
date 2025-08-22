import React from 'react'
import { 
  Eye, 
  Copy, 
  Trash2, 
  Download, 
  Upload, 
  Play, 
  Pause, 
  Users, 
  HardDrive, 
  Calendar,
  CheckCircle,
  Clock,
  AlertTriangle,
  GitBranch
} from 'lucide-react'
import { Card } from '@/design-system/components/Card/Card'
import { Badge } from '@/design-system/components/Badge/Badge'
import { Button, IconButton } from '@/design-system/components/Button/Button'
import { Progress } from '@/design-system/components/Progress/Progress'
import { cn } from '@/utils/cn'
import type { ModelInfo } from '@/types/api'

interface ModelCardProps {
  model: ModelInfo
  onAction: (action: string, modelName: string) => void
  onSelect: (model: ModelInfo) => void
  isSelected?: boolean
  className?: string
}

export const ModelCard: React.FC<ModelCardProps> = ({
  model,
  onAction,
  onSelect,
  isSelected = false,
  className
}) => {
  const formatSize = (bytes: number): string => {
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    if (bytes === 0) return '0 B'
    const i = Math.floor(Math.log(bytes) / Math.log(1024))
    return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i]
  }

  const getStatusBadge = (status: string) => {
    const variants = {
      synchronized: 'success',
      syncing: 'warning',
      failed: 'destructive',
      pending: 'secondary'
    } as const

    const icons = {
      synchronized: CheckCircle,
      syncing: Clock,
      failed: AlertTriangle,
      pending: Clock
    }

    const Icon = icons[status as keyof typeof icons] || Clock

    return (
      <Badge variant={variants[status as keyof typeof variants] || 'secondary'}>
        <Icon className="w-3 h-3 mr-1" />
        {status.charAt(0).toUpperCase() + status.slice(1)}
      </Badge>
    )
  }

  const getSyncProgress = (syncStatus: any) => {
    if (!syncStatus || syncStatus.status !== 'syncing') return null
    
    const progress = (syncStatus.synced_nodes / syncStatus.total_nodes) * 100
    return (
      <div className="flex items-center space-x-2">
        <Progress value={progress} className="flex-1" />
        <span className="text-xs text-muted-foreground min-w-fit">
          {syncStatus.synced_nodes}/{syncStatus.total_nodes}
        </span>
      </div>
    )
  }

  return (
    <Card 
      className={cn(
        'p-4 hover:shadow-md transition-shadow cursor-pointer',
        isSelected && 'ring-2 ring-primary-500',
        className
      )}
      onClick={() => onSelect(model)}
    >
      {/* Header */}
      <div className="flex items-start justify-between mb-3">
        <div className="flex-1 min-w-0">
          <h3 className="font-semibold text-foreground truncate">{model.name}</h3>
          <p className="text-xs text-muted-foreground font-mono">
            {model.digest.slice(0, 12)}...
          </p>
        </div>
        <div className="flex items-center space-x-1 ml-2">
          <Badge variant="outline">{model.tag}</Badge>
        </div>
      </div>

      {/* Model Info */}
      <div className="space-y-3 mb-4">
        <div className="flex items-center justify-between text-sm">
          <span className="text-muted-foreground">Family:</span>
          <span className="font-medium capitalize">{model.family}</span>
        </div>

        <div className="flex items-center justify-between text-sm">
          <span className="text-muted-foreground">Size:</span>
          <span className="font-medium">{formatSize(model.size)}</span>
        </div>

        <div className="flex items-center justify-between text-sm">
          <span className="text-muted-foreground">Modified:</span>
          <span className="font-medium">
            {new Date(model.modified_at).toLocaleDateString()}
          </span>
        </div>
      </div>

      {/* Status */}
      <div className="space-y-2 mb-4">
        <div className="flex items-center justify-between">
          <span className="text-sm text-muted-foreground">Status:</span>
          {getStatusBadge(model.sync_status?.status || 'pending')}
        </div>
        
        {getSyncProgress(model.sync_status) && (
          <div className="space-y-1">
            <span className="text-xs text-muted-foreground">Sync Progress:</span>
            {getSyncProgress(model.sync_status)}
          </div>
        )}
      </div>

      {/* Distribution */}
      <div className="flex items-center space-x-2 mb-4 text-sm">
        <Users className="w-4 h-4 text-muted-foreground" />
        <span className="text-muted-foreground">Distribution:</span>
        <span className="font-medium">
          {model.sync_status?.synced_nodes || 0}/{model.sync_status?.total_nodes || 0} nodes
        </span>
      </div>

      {/* Actions */}
      <div className="flex items-center justify-end space-x-1 pt-3 border-t border-border">
        <IconButton
          icon={<Eye className="w-4 h-4" />}
          variant="ghost"
          size="sm"
          aria-label="View details"
          onClick={(e) => {
            e.stopPropagation()
            onSelect(model)
          }}
        />
        <IconButton
          icon={<Copy className="w-4 h-4" />}
          variant="ghost"
          size="sm"
          aria-label="Copy model"
          onClick={(e) => {
            e.stopPropagation()
            onAction('copy', model.name)
          }}
        />
        <IconButton
          icon={<Download className="w-4 h-4" />}
          variant="ghost"
          size="sm"
          aria-label="Download model"
          onClick={(e) => {
            e.stopPropagation()
            onAction('download', model.name)
          }}
        />
        <IconButton
          icon={<Trash2 className="w-4 h-4" />}
          variant="ghost"
          size="sm"
          aria-label="Delete model"
          onClick={(e) => {
            e.stopPropagation()
            onAction('delete', model.name)
          }}
        />
      </div>
    </Card>
  )
}

export default ModelCard