import React, { useState, useEffect } from 'react'
import { 
  Brain, 
  Activity, 
  TrendingUp, 
  Zap, 
  Database,
  Play,
  Pause,
  RotateCcw,
  Download,
  Upload,
  Settings,
  AlertTriangle
} from 'lucide-react'
import { Button, Card, Input, Alert, Badge } from '@/design-system'
import { useWebSocket } from '@/hooks/useWebSocket'
import { useAPI } from '@/hooks/useAPI'
import { formatBytes, formatDuration } from '@/utils/format'

interface TrainingConfig {
  modelId: string
  dataset: string
  epochs: number
  batchSize: number
  learningRate: number
  optimizer: 'adam' | 'sgd' | 'rmsprop'
  validationSplit: number
  earlyStopping: boolean
  patience: number
  distributed: boolean
  nodes: string[]
}

interface TrainingMetrics {
  epoch: number
  loss: number
  accuracy: number
  valLoss: number
  valAccuracy: number
  learningRate: number
  timePerEpoch: number
  eta: number
}

interface TrainingJob {
  id: string
  modelId: string
  status: 'idle' | 'preparing' | 'training' | 'validating' | 'completed' | 'failed'
  progress: number
  startTime: number
  endTime?: number
  currentEpoch: number
  totalEpochs: number
  metrics: TrainingMetrics[]
  logs: string[]
  checkpoints: string[]
}

export const ModelTraining: React.FC = () => {
  const [config, setConfig] = useState<TrainingConfig>({
    modelId: '',
    dataset: '',
    epochs: 100,
    batchSize: 32,
    learningRate: 0.001,
    optimizer: 'adam',
    validationSplit: 0.2,
    earlyStopping: true,
    patience: 10,
    distributed: true,
    nodes: []
  })

  const [activeJob, setActiveJob] = useState<TrainingJob | null>(null)
  const [models, setModels] = useState<any[]>([])
  const [datasets, setDatasets] = useState<any[]>([])
  const [selectedNodes, setSelectedNodes] = useState<string[]>([])
  const [showAdvanced, setShowAdvanced] = useState(false)

  const { data: availableModels } = useAPI('/api/ml/models')
  const { data: availableDatasets } = useAPI('/api/ml/datasets')
  const { data: clusterNodes } = useAPI('/api/cluster/nodes')
  
  const ws = useWebSocket('ws://localhost:8080/ws/training', {
    onMessage: (data) => {
      if (data.type === 'training_update' && activeJob) {
        setActiveJob({
          ...activeJob,
          ...data.update
        })
      }
    }
  })

  useEffect(() => {
    if (availableModels) setModels(availableModels)
    if (availableDatasets) setDatasets(availableDatasets)
  }, [availableModels, availableDatasets])

  const startTraining = async () => {
    try {
      const response = await fetch('/api/ml/training/start', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(config)
      })
      
      const job = await response.json()
      setActiveJob(job)
      
      // Subscribe to training updates
      ws.send(JSON.stringify({
        type: 'subscribe_training',
        jobId: job.id
      }))
    } catch (error) {
      console.error('Failed to start training:', error)
    }
  }

  const pauseTraining = async () => {
    if (!activeJob) return
    
    await fetch(`/api/ml/training/${activeJob.id}/pause`, {
      method: 'POST'
    })
  }

  const resumeTraining = async () => {
    if (!activeJob) return
    
    await fetch(`/api/ml/training/${activeJob.id}/resume`, {
      method: 'POST'
    })
  }

  const stopTraining = async () => {
    if (!activeJob) return
    
    await fetch(`/api/ml/training/${activeJob.id}/stop`, {
      method: 'POST'
    })
    
    setActiveJob(null)
  }

  const downloadCheckpoint = (checkpoint: string) => {
    window.open(`/api/ml/training/checkpoints/${checkpoint}/download`, '_blank')
  }

  return (
    <div className="space-y-6">
      {/* Training Configuration */}
      <Card>
        <Card.Header>
          <Card.Title>
            <Brain className="w-5 h-5 inline mr-2" />
            Model Training Configuration
          </Card.Title>
          <Card.Description>
            Configure and start distributed model training
          </Card.Description>
        </Card.Header>
        
        <Card.Content className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Model Selection */}
            <div>
              <label className="block text-sm font-medium mb-2">Model</label>
              <select
                className="w-full px-3 py-2 border rounded-md"
                value={config.modelId}
                onChange={(e) => setConfig({ ...config, modelId: e.target.value })}
                disabled={activeJob !== null}
              >
                <option value="">Select a model...</option>
                {models.map(model => (
                  <option key={model.id} value={model.id}>
                    {model.name} ({model.architecture})
                  </option>
                ))}
              </select>
            </div>

            {/* Dataset Selection */}
            <div>
              <label className="block text-sm font-medium mb-2">Dataset</label>
              <select
                className="w-full px-3 py-2 border rounded-md"
                value={config.dataset}
                onChange={(e) => setConfig({ ...config, dataset: e.target.value })}
                disabled={activeJob !== null}
              >
                <option value="">Select a dataset...</option>
                {datasets.map(dataset => (
                  <option key={dataset.id} value={dataset.id}>
                    {dataset.name} ({formatBytes(dataset.size)})
                  </option>
                ))}
              </select>
            </div>

            {/* Epochs */}
            <Input
              label="Epochs"
              type="number"
              value={config.epochs}
              onChange={(e) => setConfig({ ...config, epochs: parseInt(e.target.value) })}
              disabled={activeJob !== null}
              min={1}
              max={1000}
            />

            {/* Batch Size */}
            <Input
              label="Batch Size"
              type="number"
              value={config.batchSize}
              onChange={(e) => setConfig({ ...config, batchSize: parseInt(e.target.value) })}
              disabled={activeJob !== null}
              min={1}
              max={512}
            />

            {/* Learning Rate */}
            <Input
              label="Learning Rate"
              type="number"
              step="0.0001"
              value={config.learningRate}
              onChange={(e) => setConfig({ ...config, learningRate: parseFloat(e.target.value) })}
              disabled={activeJob !== null}
              min={0.00001}
              max={1}
            />

            {/* Optimizer */}
            <div>
              <label className="block text-sm font-medium mb-2">Optimizer</label>
              <select
                className="w-full px-3 py-2 border rounded-md"
                value={config.optimizer}
                onChange={(e) => setConfig({ ...config, optimizer: e.target.value as any })}
                disabled={activeJob !== null}
              >
                <option value="adam">Adam</option>
                <option value="sgd">SGD</option>
                <option value="rmsprop">RMSprop</option>
              </select>
            </div>
          </div>

          {/* Advanced Settings */}
          <div className="border-t pt-4">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setShowAdvanced(!showAdvanced)}
              className="mb-4"
            >
              <Settings className="w-4 h-4 mr-2" />
              Advanced Settings
            </Button>

            {showAdvanced && (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <Input
                  label="Validation Split"
                  type="number"
                  step="0.1"
                  value={config.validationSplit}
                  onChange={(e) => setConfig({ ...config, validationSplit: parseFloat(e.target.value) })}
                  disabled={activeJob !== null}
                  min={0}
                  max={0.5}
                />

                <div>
                  <label className="flex items-center space-x-2">
                    <input
                      type="checkbox"
                      checked={config.earlyStopping}
                      onChange={(e) => setConfig({ ...config, earlyStopping: e.target.checked })}
                      disabled={activeJob !== null}
                    />
                    <span className="text-sm font-medium">Early Stopping</span>
                  </label>
                </div>

                {config.earlyStopping && (
                  <Input
                    label="Patience"
                    type="number"
                    value={config.patience}
                    onChange={(e) => setConfig({ ...config, patience: parseInt(e.target.value) })}
                    disabled={activeJob !== null}
                    min={1}
                    max={50}
                  />
                )}

                <div>
                  <label className="flex items-center space-x-2">
                    <input
                      type="checkbox"
                      checked={config.distributed}
                      onChange={(e) => setConfig({ ...config, distributed: e.target.checked })}
                      disabled={activeJob !== null}
                    />
                    <span className="text-sm font-medium">Distributed Training</span>
                  </label>
                </div>

                {config.distributed && (
                  <div>
                    <label className="block text-sm font-medium mb-2">Training Nodes</label>
                    <div className="space-y-2 max-h-40 overflow-y-auto">
                      {clusterNodes?.map((node: any) => (
                        <label key={node.id} className="flex items-center space-x-2">
                          <input
                            type="checkbox"
                            checked={config.nodes.includes(node.id)}
                            onChange={(e) => {
                              if (e.target.checked) {
                                setConfig({ ...config, nodes: [...config.nodes, node.id] })
                              } else {
                                setConfig({ ...config, nodes: config.nodes.filter(n => n !== node.id) })
                              }
                            }}
                            disabled={activeJob !== null}
                          />
                          <span className="text-sm">{node.name} ({node.gpus} GPUs)</span>
                        </label>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>

          {/* Training Controls */}
          <div className="flex justify-end space-x-2">
            {!activeJob && (
              <Button
                onClick={startTraining}
                disabled={!config.modelId || !config.dataset}
              >
                <Play className="w-4 h-4 mr-2" />
                Start Training
              </Button>
            )}
            
            {activeJob && activeJob.status === 'training' && (
              <>
                <Button variant="secondary" onClick={pauseTraining}>
                  <Pause className="w-4 h-4 mr-2" />
                  Pause
                </Button>
                <Button variant="destructive" onClick={stopTraining}>
                  Stop Training
                </Button>
              </>
            )}
            
            {activeJob && activeJob.status === 'paused' && (
              <>
                <Button onClick={resumeTraining}>
                  <Play className="w-4 h-4 mr-2" />
                  Resume
                </Button>
                <Button variant="destructive" onClick={stopTraining}>
                  Stop Training
                </Button>
              </>
            )}
          </div>
        </Card.Content>
      </Card>

      {/* Active Training Job */}
      {activeJob && (
        <Card>
          <Card.Header>
            <Card.Title>
              <Activity className="w-5 h-5 inline mr-2" />
              Training Progress
            </Card.Title>
            <div className="flex items-center space-x-2">
              <Badge variant={
                activeJob.status === 'completed' ? 'success' :
                activeJob.status === 'failed' ? 'destructive' :
                activeJob.status === 'training' ? 'default' :
                'secondary'
              }>
                {activeJob.status}
              </Badge>
              <span className="text-sm text-muted-foreground">
                Job ID: {activeJob.id}
              </span>
            </div>
          </Card.Header>
          
          <Card.Content className="space-y-4">
            {/* Progress Bar */}
            <div>
              <div className="flex justify-between text-sm mb-2">
                <span>Epoch {activeJob.currentEpoch} / {activeJob.totalEpochs}</span>
                <span>{Math.round(activeJob.progress)}%</span>
              </div>
              <div className="w-full bg-gray-200 rounded-full h-2">
                <div
                  className="bg-blue-600 h-2 rounded-full transition-all duration-300"
                  style={{ width: `${activeJob.progress}%` }}
                />
              </div>
            </div>

            {/* Current Metrics */}
            {activeJob.metrics.length > 0 && (
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                <div>
                  <p className="text-sm text-muted-foreground">Loss</p>
                  <p className="text-xl font-semibold">
                    {activeJob.metrics[activeJob.metrics.length - 1].loss.toFixed(4)}
                  </p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Accuracy</p>
                  <p className="text-xl font-semibold">
                    {(activeJob.metrics[activeJob.metrics.length - 1].accuracy * 100).toFixed(2)}%
                  </p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Val Loss</p>
                  <p className="text-xl font-semibold">
                    {activeJob.metrics[activeJob.metrics.length - 1].valLoss.toFixed(4)}
                  </p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Val Accuracy</p>
                  <p className="text-xl font-semibold">
                    {(activeJob.metrics[activeJob.metrics.length - 1].valAccuracy * 100).toFixed(2)}%
                  </p>
                </div>
              </div>
            )}

            {/* Training Time */}
            <div className="flex justify-between text-sm">
              <span>
                Elapsed: {formatDuration(Date.now() - activeJob.startTime)}
              </span>
              {activeJob.metrics.length > 0 && (
                <span>
                  ETA: {formatDuration(activeJob.metrics[activeJob.metrics.length - 1].eta * 1000)}
                </span>
              )}
            </div>

            {/* Checkpoints */}
            {activeJob.checkpoints.length > 0 && (
              <div>
                <h4 className="text-sm font-medium mb-2">Checkpoints</h4>
                <div className="space-y-1">
                  {activeJob.checkpoints.map((checkpoint, index) => (
                    <div key={checkpoint} className="flex justify-between items-center">
                      <span className="text-sm">Checkpoint {index + 1}</span>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => downloadCheckpoint(checkpoint)}
                      >
                        <Download className="w-4 h-4" />
                      </Button>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Training Logs */}
            <div>
              <h4 className="text-sm font-medium mb-2">Training Logs</h4>
              <div className="bg-gray-900 text-gray-100 p-3 rounded-md h-40 overflow-y-auto font-mono text-xs">
                {activeJob.logs.map((log, index) => (
                  <div key={index}>{log}</div>
                ))}
              </div>
            </div>
          </Card.Content>
        </Card>
      )}

      {/* Training History Charts */}
      {activeJob && activeJob.metrics.length > 0 && (
        <Card>
          <Card.Header>
            <Card.Title>
              <TrendingUp className="w-5 h-5 inline mr-2" />
              Training Metrics
            </Card.Title>
          </Card.Header>
          <Card.Content>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {/* Loss Chart */}
              <div className="h-64 bg-gray-50 rounded-lg p-4">
                <h4 className="text-sm font-medium mb-2">Loss</h4>
                {/* Chart would be rendered here using a charting library */}
                <div className="text-center text-muted-foreground mt-20">
                  Loss chart visualization
                </div>
              </div>

              {/* Accuracy Chart */}
              <div className="h-64 bg-gray-50 rounded-lg p-4">
                <h4 className="text-sm font-medium mb-2">Accuracy</h4>
                {/* Chart would be rendered here using a charting library */}
                <div className="text-center text-muted-foreground mt-20">
                  Accuracy chart visualization
                </div>
              </div>
            </div>
          </Card.Content>
        </Card>
      )}
    </div>
  )
}

export default ModelTraining