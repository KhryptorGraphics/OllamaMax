import React, { useState, useEffect, useMemo, useCallback } from 'react'
import { 
  Search, 
  Download, 
  Upload, 
  Trash2, 
  Play, 
  Pause, 
  Copy, 
  MoreHorizontal, 
  Filter,
  SortAsc,
  SortDesc,
  RefreshCw,
  Plus,
  Eye,
  Settings,
  AlertTriangle,
  CheckCircle,
  Clock,
  Activity,
  HardDrive,
  Cpu,
  GitBranch,
  Users,
  BarChart3,
  ArrowUpDown,
  ArrowUp,
  ArrowDown,
  Database,
  Grid,
  List,
  LayoutGrid
} from 'lucide-react'
import { 
  Table, 
  TableHeader, 
  TableBody, 
  TableRow, 
  TableHead, 
  TableCell,
  TablePagination,
  EmptyState
} from '@/design-system/components/Table/Table'
import { Button, IconButton } from '@/design-system/components/Button/Button'
import { Input } from '@/design-system/components/Input/Input'
import { Card } from '@/design-system/components/Card/Card'
import { Badge } from '@/design-system/components/Badge/Badge'
import { Progress } from '@/design-system/components/Progress/Progress'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/design-system/components/Tabs/Tabs'
import { SimpleSelect } from '@/components/models/SimpleSelect'
import { ModelCard } from '@/components/models/ModelCard'
import { ModelDetailPanel } from '@/components/models/ModelDetailPanel'
import { ModelPullDialog } from '@/components/models/ModelPullDialog'
import { cn } from '@/utils/cn'
import { useWebSocket } from '@/hooks/useWebSocket'
import { ModelsAPI } from '@/lib/api/models'
import type { ModelInfo } from '@/types/api'

// Initialize API client
const modelsAPI = new ModelsAPI()

// Filter and sort options
const FILTER_OPTIONS = {
  status: [
    { value: 'all', label: 'All Status' },
    { value: 'synchronized', label: 'Synchronized' },
    { value: 'syncing', label: 'Syncing' },
    { value: 'failed', label: 'Failed' },
    { value: 'pending', label: 'Pending' }
  ],
  family: [
    { value: 'all', label: 'All Families' },
    { value: 'llama', label: 'Llama' },
    { value: 'mistral', label: 'Mistral' },
    { value: 'codellama', label: 'CodeLlama' },
    { value: 'vicuna', label: 'Vicuna' },
    { value: 'other', label: 'Other' }
  ],
  size: [
    { value: 'all', label: 'All Sizes' },
    { value: 'small', label: 'Small (< 1GB)' },
    { value: 'medium', label: 'Medium (1-10GB)' },
    { value: 'large', label: 'Large (> 10GB)' }
  ]
}

const SORT_OPTIONS = [
  { value: 'name', label: 'Name' },
  { value: 'size', label: 'Size' },
  { value: 'created_at', label: 'Created' },
  { value: 'modified_at', label: 'Modified' },
  { value: 'usage', label: 'Usage' }
]

interface ModelFilters {
  search: string
  status: string
  family: string
  size: string
}

interface ModelSort {
  field: string
  direction: 'asc' | 'desc'
}

const ModelsPage: React.FC = () => {
  // State management
  const [models, setModels] = useState<ModelInfo[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedModels, setSelectedModels] = useState<Set<string>>(new Set())
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [viewMode, setViewMode] = useState<'list' | 'grid' | 'detailed'>('list')
  const [selectedModel, setSelectedModel] = useState<ModelInfo | null>(null)
  const [showDetailPanel, setShowDetailPanel] = useState(false)
  const [showPullDialog, setShowPullDialog] = useState(false)
  
  // Filters and sorting
  const [filters, setFilters] = useState<ModelFilters>({
    search: '',
    status: 'all',
    family: 'all',
    size: 'all'
  })
  const [sort, setSort] = useState<ModelSort>({
    field: 'name',
    direction: 'asc'
  })

  // WebSocket for real-time updates
  const { isConnected, subscribe, unsubscribe } = useWebSocket()

  // Load models data
  const loadModels = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const data = await modelsAPI.list()
      setModels(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load models')
    } finally {
      setLoading(false)
    }
  }, [])

  // Subscribe to WebSocket updates
  useEffect(() => {
    if (isConnected) {
      subscribe('model_updates', handleModelUpdate)
      subscribe('sync_status', handleSyncUpdate)
    }
    
    return () => {
      unsubscribe('model_updates')
      unsubscribe('sync_status')
    }
  }, [isConnected])

  // Real-time update handlers
  const handleModelUpdate = useCallback((data: any) => {
    setModels(prevModels => {
      const index = prevModels.findIndex(m => m.name === data.model)
      if (index >= 0) {
        const updated = [...prevModels]
        updated[index] = { ...updated[index], ...data }
        return updated
      }
      return prevModels
    })
  }, [])

  const handleSyncUpdate = useCallback((data: any) => {
    setModels(prevModels => 
      prevModels.map(model => 
        model.name === data.model 
          ? { ...model, sync_status: data.sync_status }
          : model
      )
    )
  }, [])

  // Filter and sort models
  const filteredModels = useMemo(() => {
    let filtered = models.filter(model => {
      // Search filter
      if (filters.search) {
        const searchLower = filters.search.toLowerCase()
        if (!model.name.toLowerCase().includes(searchLower) &&
            !model.family.toLowerCase().includes(searchLower) &&
            !model.tag.toLowerCase().includes(searchLower)) {
          return false
        }
      }

      // Status filter
      if (filters.status !== 'all' && model.sync_status?.status !== filters.status) {
        return false
      }

      // Family filter
      if (filters.family !== 'all' && model.family.toLowerCase() !== filters.family) {
        return false
      }

      // Size filter
      if (filters.size !== 'all') {
        const sizeGB = model.size / (1024 * 1024 * 1024)
        if (filters.size === 'small' && sizeGB >= 1) return false
        if (filters.size === 'medium' && (sizeGB < 1 || sizeGB > 10)) return false
        if (filters.size === 'large' && sizeGB <= 10) return false
      }

      return true
    })

    // Sort filtered results
    filtered.sort((a, b) => {
      let aValue: any, bValue: any

      switch (sort.field) {
        case 'name':
          aValue = a.name
          bValue = b.name
          break
        case 'size':
          aValue = a.size
          bValue = b.size
          break
        case 'created_at':
          aValue = new Date(a.created_at)
          bValue = new Date(b.created_at)
          break
        case 'modified_at':
          aValue = new Date(a.modified_at)
          bValue = new Date(b.modified_at)
          break
        default:
          aValue = a.name
          bValue = b.name
      }

      if (aValue < bValue) return sort.direction === 'asc' ? -1 : 1
      if (aValue > bValue) return sort.direction === 'asc' ? 1 : -1
      return 0
    })

    return filtered
  }, [models, filters, sort])

  // Pagination
  const totalPages = Math.ceil(filteredModels.length / pageSize)
  const paginatedModels = filteredModels.slice(
    (currentPage - 1) * pageSize,
    currentPage * pageSize
  )

  // Load data on mount
  useEffect(() => {
    loadModels()
  }, [loadModels])

  // Model operations
  const handleModelAction = async (action: string, modelName: string) => {
    try {
      switch (action) {
        case 'delete':
          await modelsAPI.delete(modelName)
          break
        case 'copy':
          // Handle copy action
          break
        case 'download':
          // Handle download action
          break
        default:
          break
      }
      await loadModels()
    } catch (err) {
      setError(err instanceof Error ? err.message : `Failed to ${action} model`)
    }
  }

  const handleBulkAction = async (action: string) => {
    if (selectedModels.size === 0) return

    try {
      for (const modelName of selectedModels) {
        await handleModelAction(action, modelName)
      }
      setSelectedModels(new Set())
    } catch (err) {
      setError(err instanceof Error ? err.message : `Failed to ${action} models`)
    }
  }

  // Helper functions
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
        <Progress value={progress} className="w-20" />
        <span className="text-xs text-muted-foreground">
          {syncStatus.synced_nodes}/{syncStatus.total_nodes}
        </span>
      </div>
    )
  }

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Models</h1>
          <p className="text-gray-600 dark:text-gray-400">
            Manage AI models across your distributed cluster
          </p>
        </div>
        
        <div className="flex items-center space-x-2">
          <Button 
            variant="outline" 
            leftIcon={<RefreshCw className="w-4 h-4" />}
            onClick={loadModels}
            loading={loading}
          >
            Refresh
          </Button>
          <Button 
            leftIcon={<Plus className="w-4 h-4" />}
            onClick={() => setShowPullDialog(true)}
          >
            Pull Model
          </Button>
        </div>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card className="p-4">
          <div className="flex items-center space-x-3">
            <Database className="w-8 h-8 text-primary-500" />
            <div>
              <p className="text-sm text-muted-foreground">Total Models</p>
              <p className="text-2xl font-bold">{models.length}</p>
            </div>
          </div>
        </Card>
        
        <Card className="p-4">
          <div className="flex items-center space-x-3">
            <CheckCircle className="w-8 h-8 text-green-500" />
            <div>
              <p className="text-sm text-muted-foreground">Synchronized</p>
              <p className="text-2xl font-bold">
                {models.filter(m => m.sync_status?.status === 'synchronized').length}
              </p>
            </div>
          </div>
        </Card>
        
        <Card className="p-4">
          <div className="flex items-center space-x-3">
            <Activity className="w-8 h-8 text-yellow-500" />
            <div>
              <p className="text-sm text-muted-foreground">Syncing</p>
              <p className="text-2xl font-bold">
                {models.filter(m => m.sync_status?.status === 'syncing').length}
              </p>
            </div>
          </div>
        </Card>
        
        <Card className="p-4">
          <div className="flex items-center space-x-3">
            <HardDrive className="w-8 h-8 text-blue-500" />
            <div>
              <p className="text-sm text-muted-foreground">Total Size</p>
              <p className="text-2xl font-bold">
                {formatSize(models.reduce((acc, m) => acc + m.size, 0))}
              </p>
            </div>
          </div>
        </Card>
      </div>

      {/* Filters and Controls */}
      <Card className="p-4">
        <div className="space-y-4">
          {/* Search and View Toggle */}
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4 flex-1">
              <Input
                type="search"
                placeholder="Search models..."
                value={filters.search}
                onChange={(e) => setFilters(prev => ({ ...prev, search: e.target.value }))}
                className="w-64"
                leftIcon={<Search className="w-4 h-4" />}
              />
              
              <SimpleSelect
                value={filters.status}
                onValueChange={(value) => setFilters(prev => ({ ...prev, status: value }))}
                options={FILTER_OPTIONS.status}
                className="w-40"
              />

              <SimpleSelect
                value={filters.family}
                onValueChange={(value) => setFilters(prev => ({ ...prev, family: value }))}
                options={FILTER_OPTIONS.family}
                className="w-40"
              />

              <SimpleSelect
                value={sort.field}
                onValueChange={(value) => setSort(prev => ({ ...prev, field: value }))}
                options={SORT_OPTIONS}
                className="w-40"
              />

              <IconButton
                icon={sort.direction === 'asc' ? <SortAsc className="w-4 h-4" /> : <SortDesc className="w-4 h-4" />}
                variant="outline"
                aria-label={`Sort ${sort.direction === 'asc' ? 'ascending' : 'descending'}`}
                onClick={() => setSort(prev => ({ 
                  ...prev, 
                  direction: prev.direction === 'asc' ? 'desc' : 'asc' 
                }))}
              />
            </div>

            {/* View Mode Toggle */}
            <div className="flex items-center space-x-1 border border-border rounded-md">
              <IconButton
                icon={<List className="w-4 h-4" />}
                variant={viewMode === 'list' ? 'primary' : 'ghost'}
                size="sm"
                aria-label="List view"
                onClick={() => setViewMode('list')}
              />
              <IconButton
                icon={<LayoutGrid className="w-4 h-4" />}
                variant={viewMode === 'grid' ? 'primary' : 'ghost'}
                size="sm"
                aria-label="Grid view"
                onClick={() => setViewMode('grid')}
              />
            </div>

            {/* Bulk Actions */}
            {selectedModels.size > 0 && (
              <div className="flex items-center space-x-2">
                <span className="text-sm text-muted-foreground">
                  {selectedModels.size} selected
                </span>
                <Button
                  variant="outline"
                  size="sm"
                  leftIcon={<Trash2 className="w-4 h-4" />}
                  onClick={() => handleBulkAction('delete')}
                >
                  Delete
                </Button>
              </div>
            )}
          </div>
        </div>
      </Card>

      {/* Models Content */}
      <Card>
        {error && (
          <div className="p-4 bg-destructive/10 border-b border-destructive/20">
            <div className="flex items-center space-x-2 text-destructive">
              <AlertTriangle className="w-4 h-4" />
              <span className="text-sm">{error}</span>
            </div>
          </div>
        )}

        {/* List View */}
        {viewMode === 'list' && (
          <Table
            selectionMode="multiple"
            selectedRows={selectedModels}
            onSelectionChange={setSelectedModels}
            isLoading={loading}
            stickyHeader
          >
            <TableHeader>
              <TableRow>
                <TableHead sortKey="name" sortable>Name</TableHead>
                <TableHead>Tag</TableHead>
                <TableHead sortKey="family" sortable>Family</TableHead>
                <TableHead sortKey="size" sortable>Size</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Distribution</TableHead>
                <TableHead sortKey="modified_at" sortable>Modified</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {paginatedModels.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={8}>
                    <EmptyState
                      title="No models found"
                      description="No models match your current filters. Try adjusting your search criteria."
                      action={
                        <Button 
                          variant="outline"
                          onClick={() => setFilters({ search: '', status: 'all', family: 'all', size: 'all' })}
                        >
                          Clear Filters
                        </Button>
                      }
                    />
                  </TableCell>
                </TableRow>
              ) : (
                paginatedModels.map((model) => (
                  <TableRow 
                    key={`${model.name}:${model.tag}`}
                    rowId={`${model.name}:${model.tag}`}
                    onRowClick={() => {
                      setSelectedModel(model)
                      setShowDetailPanel(true)
                    }}
                  >
                    <TableCell>
                      <div className="flex flex-col">
                        <span className="font-medium">{model.name}</span>
                        <span className="text-xs text-muted-foreground">{model.digest.slice(0, 12)}</span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline">{model.tag}</Badge>
                    </TableCell>
                    <TableCell>
                      <span className="capitalize">{model.family}</span>
                    </TableCell>
                    <TableCell>{formatSize(model.size)}</TableCell>
                    <TableCell>
                      <div className="space-y-1">
                        {getStatusBadge(model.sync_status?.status || 'pending')}
                        {getSyncProgress(model.sync_status)}
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center space-x-2">
                        <Users className="w-4 h-4 text-muted-foreground" />
                        <span className="text-sm">
                          {model.sync_status?.synced_nodes || 0}/{model.sync_status?.total_nodes || 0}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <span className="text-sm text-muted-foreground">
                        {new Date(model.modified_at).toLocaleDateString()}
                      </span>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center space-x-1">
                        <IconButton
                          icon={<Eye className="w-4 h-4" />}
                          variant="ghost"
                          size="sm"
                          aria-label="View details"
                          onClick={(e) => {
                            e.stopPropagation()
                            setSelectedModel(model)
                            setShowDetailPanel(true)
                          }}
                        />
                        <IconButton
                          icon={<Copy className="w-4 h-4" />}
                          variant="ghost"
                          size="sm"
                          aria-label="Copy model"
                          onClick={(e) => {
                            e.stopPropagation()
                            handleModelAction('copy', model.name)
                          }}
                        />
                        <IconButton
                          icon={<Trash2 className="w-4 h-4" />}
                          variant="ghost"
                          size="sm"
                          aria-label="Delete model"
                          onClick={(e) => {
                            e.stopPropagation()
                            handleModelAction('delete', model.name)
                          }}
                        />
                      </div>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        )}

        {/* Grid View */}
        {viewMode === 'grid' && (
          <div className="p-4">
            {loading ? (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
                {Array.from({ length: 8 }).map((_, index) => (
                  <div key={index} className="h-64 bg-muted/30 rounded-lg animate-pulse" />
                ))}
              </div>
            ) : paginatedModels.length === 0 ? (
              <EmptyState
                title="No models found"
                description="No models match your current filters. Try adjusting your search criteria."
                action={
                  <Button 
                    variant="outline"
                    onClick={() => setFilters({ search: '', status: 'all', family: 'all', size: 'all' })}
                  >
                    Clear Filters
                  </Button>
                }
              />
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
                {paginatedModels.map((model) => (
                  <ModelCard
                    key={`${model.name}:${model.tag}`}
                    model={model}
                    onAction={handleModelAction}
                    onSelect={(selectedModel) => {
                      setSelectedModel(selectedModel)
                      setShowDetailPanel(true)
                    }}
                    isSelected={selectedModels.has(`${model.name}:${model.tag}`)}
                  />
                ))}
              </div>
            )}
          </div>
        )}

        {/* Pagination */}
        {filteredModels.length > 0 && (
          <TablePagination
            currentPage={currentPage}
            totalPages={totalPages}
            pageSize={pageSize}
            totalItems={filteredModels.length}
            onPageChange={setCurrentPage}
            onPageSizeChange={setPageSize}
          />
        )}
      </Card>

      {/* Detail Panel */}
      {showDetailPanel && (
        <ModelDetailPanel
          model={selectedModel}
          onClose={() => {
            setShowDetailPanel(false)
            setSelectedModel(null)
          }}
          onAction={handleModelAction}
        />
      )}

      {/* Pull Dialog */}
      <ModelPullDialog
        isOpen={showPullDialog}
        onClose={() => setShowPullDialog(false)}
        onSuccess={() => {
          setShowPullDialog(false)
          loadModels()
        }}
      />
    </div>
  )
}

export default ModelsPage