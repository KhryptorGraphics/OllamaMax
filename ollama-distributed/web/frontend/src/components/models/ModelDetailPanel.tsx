import React, { useState } from 'react'
import { 
  X, 
  Copy, 
  Download, 
  Upload, 
  Trash2, 
  Edit, 
  Play, 
  Pause, 
  RefreshCw,
  Users, 
  HardDrive, 
  Cpu, 
  Calendar,
  GitBranch,
  BarChart3,
  Activity,
  CheckCircle,
  Clock,
  AlertTriangle,
  Database,
  FileText,
  Settings,
  Network,
  Monitor
} from 'lucide-react'
import { Card } from '@/design-system/components/Card/Card'
import { Badge } from '@/design-system/components/Badge/Badge'
import { Button, IconButton } from '@/design-system/components/Button/Button'
import { Progress } from '@/design-system/components/Progress/Progress'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/design-system/components/Tabs/Tabs'
import { cn } from '@/utils/cn'
import type { ModelInfo } from '@/types/api'

interface ModelDetailPanelProps {
  model: ModelInfo | null
  onClose: () => void
  onAction: (action: string, modelName: string) => void
  className?: string
}

export const ModelDetailPanel: React.FC<ModelDetailPanelProps> = ({
  model,
  onClose,
  onAction,
  className
}) => {
  const [activeTab, setActiveTab] = useState('overview')

  if (!model) return null

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
      <div className="space-y-2">
        <div className="flex items-center justify-between text-sm">
          <span>Sync Progress</span>
          <span>{Math.round(progress)}%</span>
        </div>
        <Progress value={progress} />
        <div className="flex items-center justify-between text-xs text-muted-foreground">
          <span>{syncStatus.synced_nodes} of {syncStatus.total_nodes} nodes</span>
          <span>ETA: 2m 15s</span>
        </div>
      </div>
    )
  }

  return (
    <div className={cn(
      'fixed inset-y-0 right-0 w-96 bg-background border-l border-border shadow-lg z-50',
      'transform transition-transform duration-300 ease-in-out',
      className
    )}>
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b border-border">
        <div className="flex-1 min-w-0">
          <h2 className="text-lg font-semibold truncate">{model.name}</h2>
          <p className="text-sm text-muted-foreground font-mono">
            {model.digest.slice(0, 16)}...
          </p>
        </div>
        <IconButton
          icon={<X className="w-4 h-4" />}
          variant="ghost"
          size="sm"
          aria-label="Close panel"
          onClick={onClose}
        />
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto">
        <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
          <TabsList className="grid w-full grid-cols-4 mx-4 mt-4">
            <TabsTrigger value="overview">Overview</TabsTrigger>
            <TabsTrigger value="distribution">Distribution</TabsTrigger>
            <TabsTrigger value="usage">Usage</TabsTrigger>
            <TabsTrigger value="versions">Versions</TabsTrigger>
          </TabsList>

          {/* Overview Tab */}
          <TabsContent value="overview" className="p-4 space-y-4">
            {/* Basic Info */}
            <Card className="p-4">
              <h3 className="font-medium mb-3 flex items-center">
                <Database className="w-4 h-4 mr-2" />
                Model Information
              </h3>
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Tag:</span>
                  <Badge variant="outline">{model.tag}</Badge>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Family:</span>
                  <span className="text-sm font-medium capitalize">{model.family}</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Format:</span>
                  <span className="text-sm font-medium">{model.format}</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Size:</span>
                  <span className="text-sm font-medium">{formatSize(model.size)}</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Parameters:</span>
                  <span className="text-sm font-medium">{model.parameter_size}</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Quantization:</span>
                  <span className="text-sm font-medium">{model.quantization_level}</span>
                </div>
              </div>
            </Card>

            {/* Status */}
            <Card className="p-4">
              <h3 className="font-medium mb-3 flex items-center">
                <Activity className="w-4 h-4 mr-2" />
                Sync Status
              </h3>
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Status:</span>
                  {getStatusBadge(model.sync_status?.status || 'pending')}
                </div>
                {getSyncProgress(model.sync_status)}
                {model.sync_status?.error && (
                  <div className="p-3 bg-destructive/10 border border-destructive/20 rounded-md">
                    <p className="text-sm text-destructive">{model.sync_status.error}</p>
                  </div>
                )}
              </div>
            </Card>

            {/* Timestamps */}
            <Card className="p-4">
              <h3 className="font-medium mb-3 flex items-center">
                <Calendar className="w-4 h-4 mr-2" />
                Timestamps
              </h3>
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Created:</span>
                  <span className="text-sm font-medium">
                    {new Date(model.created_at).toLocaleString()}
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Modified:</span>
                  <span className="text-sm font-medium">
                    {new Date(model.modified_at).toLocaleString()}
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Last Sync:</span>
                  <span className="text-sm font-medium">
                    {model.sync_status?.last_sync 
                      ? new Date(model.sync_status.last_sync).toLocaleString()
                      : 'Never'
                    }
                  </span>
                </div>
              </div>
            </Card>
          </TabsContent>

          {/* Distribution Tab */}
          <TabsContent value="distribution" className="p-4 space-y-4">
            <Card className="p-4">
              <h3 className="font-medium mb-3 flex items-center">
                <Network className="w-4 h-4 mr-2" />
                Node Distribution
              </h3>
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Strategy:</span>
                  <Badge variant="outline" className="capitalize">
                    {model.distribution?.strategy || 'replicated'}
                  </Badge>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Replicas:</span>
                  <span className="text-sm font-medium">{model.distribution?.replicas || 1}</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Total Nodes:</span>
                  <span className="text-sm font-medium">{model.sync_status?.total_nodes || 0}</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Synced Nodes:</span>
                  <span className="text-sm font-medium">{model.sync_status?.synced_nodes || 0}</span>
                </div>
              </div>
            </Card>

            {/* Node List */}
            <Card className="p-4">
              <h3 className="font-medium mb-3 flex items-center">
                <Monitor className="w-4 h-4 mr-2" />
                Active Nodes
              </h3>
              <div className="space-y-2">
                {model.sync_status?.nodes?.map((nodeId, index) => (
                  <div key={nodeId} className="flex items-center justify-between p-2 bg-muted/30 rounded">
                    <span className="text-sm font-mono">{nodeId}</span>
                    <Badge 
                      variant={model.sync_status?.failed_nodes?.includes(nodeId) ? 'destructive' : 'success'}
                      size="sm"
                    >
                      {model.sync_status?.failed_nodes?.includes(nodeId) ? 'Failed' : 'Active'}
                    </Badge>
                  </div>
                )) || (
                  <p className="text-sm text-muted-foreground">No nodes available</p>
                )}
              </div>
            </Card>
          </TabsContent>

          {/* Usage Tab */}
          <TabsContent value="usage" className="p-4 space-y-4">
            <Card className="p-4">
              <h3 className="font-medium mb-3 flex items-center">
                <BarChart3 className="w-4 h-4 mr-2" />
                Usage Statistics
              </h3>
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Total Requests:</span>
                  <span className="text-sm font-medium">1,234</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Success Rate:</span>
                  <span className="text-sm font-medium">98.5%</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Avg Response Time:</span>
                  <span className="text-sm font-medium">245ms</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Last Used:</span>
                  <span className="text-sm font-medium">2 hours ago</span>
                </div>
              </div>
            </Card>

            <Card className="p-4">
              <h3 className="font-medium mb-3">Usage Trend (24h)</h3>
              <div className="h-32 bg-muted/30 rounded-md flex items-center justify-center">
                <span className="text-sm text-muted-foreground">Chart placeholder</span>
              </div>
            </Card>
          </TabsContent>

          {/* Versions Tab */}
          <TabsContent value="versions" className="p-4 space-y-4">
            <Card className="p-4">
              <h3 className="font-medium mb-3 flex items-center">
                <GitBranch className="w-4 h-4 mr-2" />
                Version History
              </h3>
              <div className="space-y-3">
                <div className="flex items-center justify-between p-3 bg-primary/10 border border-primary/20 rounded-md">
                  <div>
                    <p className="text-sm font-medium">{model.tag} (Current)</p>
                    <p className="text-xs text-muted-foreground">{model.digest.slice(0, 16)}</p>
                  </div>
                  <Badge variant="success">Active</Badge>
                </div>
                
                {/* Mock previous versions */}
                <div className="flex items-center justify-between p-3 border border-border rounded-md">
                  <div>
                    <p className="text-sm font-medium">v1.0.0</p>
                    <p className="text-xs text-muted-foreground">sha256:abc123def456</p>
                  </div>
                  <Button variant="ghost" size="sm">Rollback</Button>
                </div>
              </div>
            </Card>
          </TabsContent>
        </Tabs>
      </div>

      {/* Actions */}
      <div className="p-4 border-t border-border">
        <div className="grid grid-cols-2 gap-2">
          <Button 
            variant="outline" 
            size="sm"
            leftIcon={<Download className="w-4 h-4" />}
            onClick={() => onAction('download', model.name)}
          >
            Download
          </Button>
          <Button 
            variant="outline" 
            size="sm"
            leftIcon={<Copy className="w-4 h-4" />}
            onClick={() => onAction('copy', model.name)}
          >
            Copy
          </Button>
          <Button 
            variant="outline" 
            size="sm"
            leftIcon={<RefreshCw className="w-4 h-4" />}
            onClick={() => onAction('sync', model.name)}
          >
            Sync
          </Button>
          <Button 
            variant="destructive" 
            size="sm"
            leftIcon={<Trash2 className="w-4 h-4" />}
            onClick={() => onAction('delete', model.name)}
          >
            Delete
          </Button>
        </div>
      </div>
    </div>
  )
}

export default ModelDetailPanel