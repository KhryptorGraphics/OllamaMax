import React, { useState, useEffect } from 'react'
import { 
  X, 
  Download, 
  Search, 
  AlertCircle, 
  CheckCircle, 
  Clock,
  Loader2,
  HardDrive,
  Database,
  Tag,
  Info
} from 'lucide-react'
import { Card } from '@/design-system/components/Card/Card'
import { Badge } from '@/design-system/components/Badge/Badge'
import { Button } from '@/design-system/components/Button/Button'
import { Input } from '@/design-system/components/Input/Input'
import { Progress } from '@/design-system/components/Progress/Progress'
import { cn } from '@/utils/cn'
import { ModelsAPI } from '@/lib/api/models'
import type { DownloadProgress } from '@/types/api'

interface ModelPullDialogProps {
  isOpen: boolean
  onClose: () => void
  onSuccess: () => void
  className?: string
}

interface PopularModel {
  name: string
  description: string
  size: string
  downloads: number
  tags: string[]
  family: string
}

// Popular models for suggestions
const POPULAR_MODELS: PopularModel[] = [
  {
    name: 'llama2',
    description: 'Meta\'s Llama 2 model, great for general-purpose tasks',
    size: '3.8GB',
    downloads: 50000,
    tags: ['7b', '13b', '70b'],
    family: 'llama'
  },
  {
    name: 'mistral',
    description: 'Mistral 7B model, excellent for reasoning and coding',
    size: '4.1GB',
    downloads: 25000,
    tags: ['7b', 'instruct'],
    family: 'mistral'
  },
  {
    name: 'codellama',
    description: 'Code-specialized Llama model for programming tasks',
    size: '3.8GB',
    downloads: 30000,
    tags: ['7b', '13b', '34b'],
    family: 'codellama'
  },
  {
    name: 'vicuna',
    description: 'Fine-tuned LLaMA model trained on user conversations',
    size: '3.9GB',
    downloads: 15000,
    tags: ['7b', '13b', '33b'],
    family: 'vicuna'
  }
]

const modelsAPI = new ModelsAPI()

export const ModelPullDialog: React.FC<ModelPullDialogProps> = ({
  isOpen,
  onClose,
  onSuccess,
  className
}) => {
  const [modelName, setModelName] = useState('')
  const [selectedTag, setSelectedTag] = useState('')
  const [isDownloading, setIsDownloading] = useState(false)
  const [downloadProgress, setDownloadProgress] = useState<DownloadProgress | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [searchQuery, setSearchQuery] = useState('')

  // Filter popular models based on search
  const filteredModels = POPULAR_MODELS.filter(model =>
    model.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    model.description.toLowerCase().includes(searchQuery.toLowerCase()) ||
    model.family.toLowerCase().includes(searchQuery.toLowerCase())
  )

  // Reset state when dialog opens/closes
  useEffect(() => {
    if (!isOpen) {
      setModelName('')
      setSelectedTag('')
      setIsDownloading(false)
      setDownloadProgress(null)
      setError(null)
      setSearchQuery('')
    }
  }, [isOpen])

  const handlePull = async () => {
    if (!modelName.trim()) {
      setError('Please enter a model name')
      return
    }

    const fullModelName = selectedTag ? `${modelName}:${selectedTag}` : modelName

    try {
      setIsDownloading(true)
      setError(null)
      setDownloadProgress({ status: 'downloading', completed: 0, total: 0 })

      // Start streaming download
      const progressStream = modelsAPI.downloadStream(fullModelName)
      
      for await (const progress of progressStream) {
        setDownloadProgress(progress)
        
        if (progress.status === 'success') {
          onSuccess()
          onClose()
          return
        }
        
        if (progress.status === 'error') {
          throw new Error(progress.error || 'Download failed')
        }
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to download model')
    } finally {
      setIsDownloading(false)
    }
  }

  const handleCancel = () => {
    if (isDownloading) {
      // In a real implementation, you'd cancel the download stream here
      setIsDownloading(false)
      setDownloadProgress(null)
    }
    onClose()
  }

  const selectPopularModel = (model: PopularModel, tag?: string) => {
    setModelName(model.name)
    setSelectedTag(tag || '')
    setSearchQuery('')
  }

  const formatBytes = (bytes: number): string => {
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    if (bytes === 0) return '0 B'
    const i = Math.floor(Math.log(bytes) / Math.log(1024))
    return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i]
  }

  const getProgressPercentage = () => {
    if (!downloadProgress || downloadProgress.total === 0) return 0
    return (downloadProgress.completed / downloadProgress.total) * 100
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <Card className={cn('w-full max-w-2xl max-h-[90vh] overflow-hidden', className)}>
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-border">
          <div>
            <h2 className="text-xl font-semibold">Pull Model</h2>
            <p className="text-sm text-muted-foreground mt-1">
              Download a model from the registry
            </p>
          </div>
          <Button
            variant="ghost"
            size="sm"
            onClick={handleCancel}
            disabled={isDownloading}
          >
            <X className="w-4 h-4" />
          </Button>
        </div>

        <div className="p-6 space-y-6 overflow-y-auto max-h-[calc(90vh-180px)]">
          {/* Model Input */}
          <div className="space-y-4">
            <div>
              <label className="text-sm font-medium mb-2 block">
                Model Name <span className="text-destructive">*</span>
              </label>
              <Input
                value={modelName}
                onChange={(e) => setModelName(e.target.value)}
                placeholder="Enter model name (e.g., llama2, mistral)"
                disabled={isDownloading}
                error={error}
              />
            </div>

            <div>
              <label className="text-sm font-medium mb-2 block">
                Tag (Optional)
              </label>
              <Input
                value={selectedTag}
                onChange={(e) => setSelectedTag(e.target.value)}
                placeholder="Enter tag (e.g., 7b, 13b, latest)"
                disabled={isDownloading}
                helperText="Leave empty to use the latest tag"
              />
            </div>
          </div>

          {/* Download Progress */}
          {isDownloading && downloadProgress && (
            <Card className="p-4 bg-primary/5 border-primary/20">
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium">Downloading {modelName}...</span>
                  <span className="text-sm text-muted-foreground">
                    {Math.round(getProgressPercentage())}%
                  </span>
                </div>
                
                <Progress value={getProgressPercentage()} />
                
                <div className="flex items-center justify-between text-xs text-muted-foreground">
                  <span>
                    {formatBytes(downloadProgress.completed)} / {formatBytes(downloadProgress.total)}
                  </span>
                  <span>
                    {downloadProgress.status === 'downloading' && (
                      <span className="flex items-center">
                        <Loader2 className="w-3 h-3 mr-1 animate-spin" />
                        Downloading...
                      </span>
                    )}
                  </span>
                </div>
              </div>
            </Card>
          )}

          {/* Error Display */}
          {error && (
            <Card className="p-4 bg-destructive/10 border-destructive/20">
              <div className="flex items-center space-x-2 text-destructive">
                <AlertCircle className="w-4 h-4" />
                <span className="text-sm">{error}</span>
              </div>
            </Card>
          )}

          {/* Popular Models */}
          {!isDownloading && (
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <h3 className="text-lg font-medium">Popular Models</h3>
                <Input
                  type="search"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  placeholder="Search models..."
                  className="w-64"
                  leftIcon={<Search className="w-4 h-4" />}
                />
              </div>

              <div className="grid gap-4">
                {filteredModels.map((model) => (
                  <Card 
                    key={model.name} 
                    className="p-4 hover:shadow-md transition-shadow cursor-pointer"
                    onClick={() => selectPopularModel(model)}
                  >
                    <div className="space-y-3">
                      {/* Header */}
                      <div className="flex items-start justify-between">
                        <div className="flex-1">
                          <h4 className="font-medium">{model.name}</h4>
                          <p className="text-sm text-muted-foreground mt-1">
                            {model.description}
                          </p>
                        </div>
                        <Badge variant="outline" className="ml-2">
                          {model.family}
                        </Badge>
                      </div>

                      {/* Stats */}
                      <div className="flex items-center space-x-4 text-sm text-muted-foreground">
                        <div className="flex items-center space-x-1">
                          <HardDrive className="w-3 h-3" />
                          <span>{model.size}</span>
                        </div>
                        <div className="flex items-center space-x-1">
                          <Download className="w-3 h-3" />
                          <span>{model.downloads.toLocaleString()} downloads</span>
                        </div>
                      </div>

                      {/* Tags */}
                      <div className="flex items-center space-x-2">
                        <Tag className="w-3 h-3 text-muted-foreground" />
                        <div className="flex flex-wrap gap-1">
                          {model.tags.map((tag) => (
                            <Button
                              key={tag}
                              variant="ghost"
                              size="sm"
                              className="h-6 px-2 text-xs"
                              onClick={(e) => {
                                e.stopPropagation()
                                selectPopularModel(model, tag)
                              }}
                            >
                              {tag}
                            </Button>
                          ))}
                        </div>
                      </div>
                    </div>
                  </Card>
                ))}
              </div>

              {filteredModels.length === 0 && searchQuery && (
                <div className="text-center py-8">
                  <Database className="w-12 h-12 text-muted-foreground mx-auto mb-3" />
                  <p className="text-muted-foreground">No models found matching "{searchQuery}"</p>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between p-6 border-t border-border">
          <div className="flex items-center text-sm text-muted-foreground">
            <Info className="w-4 h-4 mr-1" />
            Models will be downloaded to all cluster nodes
          </div>
          
          <div className="flex items-center space-x-2">
            <Button
              variant="outline"
              onClick={handleCancel}
              disabled={isDownloading}
            >
              {isDownloading ? 'Cancel' : 'Close'}
            </Button>
            <Button
              onClick={handlePull}
              disabled={!modelName.trim() || isDownloading}
              loading={isDownloading}
              leftIcon={<Download className="w-4 h-4" />}
            >
              {isDownloading ? 'Downloading...' : 'Pull Model'}
            </Button>
          </div>
        </div>
      </Card>
    </div>
  )
}

export default ModelPullDialog