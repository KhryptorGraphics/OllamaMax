import React, { useState, useEffect } from 'react'
import { 
  Smartphone, 
  Wifi, 
  MapPin, 
  Zap, 
  Globe,
  Cpu,
  HardDrive,
  Network,
  Clock,
  Signal,
  Battery,
  Thermometer,
  Activity,
  Settings,
  Play,
  Pause,
  RefreshCw,
  AlertTriangle,
  CheckCircle,
  XCircle
} from 'lucide-react'
import { Button, Card, Input, Alert, Badge, Switch } from '@/design-system'
import { useWebSocket } from '@/hooks/useWebSocket'
import { useAPI } from '@/hooks/useAPI'
import { formatBytes, formatNumber, formatDuration } from '@/utils/format'

interface EdgeNode {
  id: string
  name: string
  location: {
    latitude: number
    longitude: number
    city: string
    country: string
  }
  status: 'online' | 'offline' | 'syncing' | 'error'
  type: 'mobile' | 'iot' | 'edge_server' | 'gateway'
  capabilities: {
    cpu: number // cores
    memory: number // GB
    storage: number // GB
    gpu: boolean
    tpu: boolean
  }
  metrics: {
    cpuUsage: number
    memoryUsage: number
    storageUsage: number
    networkLatency: number
    batteryLevel?: number
    temperature: number
    uptime: number
  }
  models: string[]
  lastSeen: string
  syncStatus: {
    lastSync: string
    pendingUpdates: number
    failedSyncs: number
  }
  config: {
    autoSync: boolean
    syncInterval: number
    maxStorageUsage: number
    offlineMode: boolean
    compressionEnabled: boolean
  }
}

interface EdgeDeployment {
  id: string
  name: string
  modelId: string
  targetNodes: string[]
  status: 'pending' | 'deploying' | 'deployed' | 'failed'
  progress: number
  createdAt: string
  deployedAt?: string
  config: {
    replicationFactor: number
    autoFailover: boolean
    loadBalancing: boolean
    compressionLevel: number
  }
}

interface EdgeMetrics {
  totalNodes: number
  onlineNodes: number
  totalRequests: number
  avgLatency: number
  dataTransferred: number
  syncSuccess: number
  syncFailures: number
}

export const EdgeComputing: React.FC = () => {
  const [nodes, setNodes] = useState<EdgeNode[]>([])
  const [selectedNode, setSelectedNode] = useState<EdgeNode | null>(null)
  const [deployments, setDeployments] = useState<EdgeDeployment[]>([])
  const [metrics, setMetrics] = useState<EdgeMetrics | null>(null)
  const [filterType, setFilterType] = useState<string>('all')
  const [filterStatus, setFilterStatus] = useState<string>('all')
  const [showDeployModal, setShowDeployModal] = useState(false)
  const [newDeployment, setNewDeployment] = useState({
    name: '',
    modelId: '',
    targetNodes: [] as string[],
    replicationFactor: 2,
    autoFailover: true,
    loadBalancing: true,
    compressionLevel: 5
  })

  const { data: edgeData, mutate: refreshEdgeData } = useAPI('/api/edge/overview')
  const { data: availableModels } = useAPI('/api/models')

  const ws = useWebSocket('ws://localhost:8080/ws/edge', {
    onMessage: (data) => {
      switch (data.type) {
        case 'node_update':
          setNodes(prev => {
            const index = prev.findIndex(n => n.id === data.node.id)
            if (index >= 0) {
              const updated = [...prev]
              updated[index] = { ...updated[index], ...data.node }
              return updated
            }
            return [...prev, data.node]
          })
          break
        case 'deployment_update':
          setDeployments(prev => {
            const index = prev.findIndex(d => d.id === data.deployment.id)
            if (index >= 0) {
              const updated = [...prev]
              updated[index] = { ...updated[index], ...data.deployment }
              return updated
            }
            return [...prev, data.deployment]
          })
          break
        case 'metrics_update':
          setMetrics(data.metrics)
          break
      }
    }
  })

  useEffect(() => {
    if (edgeData) {
      setNodes(edgeData.nodes || [])
      setDeployments(edgeData.deployments || [])
      setMetrics(edgeData.metrics || null)
    }
  }, [edgeData])

  const filteredNodes = nodes.filter(node => {
    const typeMatch = filterType === 'all' || node.type === filterType
    const statusMatch = filterStatus === 'all' || node.status === filterStatus
    return typeMatch && statusMatch
  })

  const deployModel = async () => {
    try {
      const response = await fetch('/api/edge/deploy', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(newDeployment)
      })
      
      if (!response.ok) throw new Error('Failed to deploy model')
      
      const deployment = await response.json()
      setDeployments(prev => [...prev, deployment])
      setShowDeployModal(false)
      setNewDeployment({
        name: '',
        modelId: '',
        targetNodes: [],
        replicationFactor: 2,
        autoFailover: true,
        loadBalancing: true,
        compressionLevel: 5
      })
    } catch (error) {
      console.error('Failed to deploy model:', error)
    }
  }

  const syncNode = async (nodeId: string) => {
    try {
      await fetch(`/api/edge/nodes/${nodeId}/sync`, {
        method: 'POST'
      })
      refreshEdgeData()
    } catch (error) {
      console.error('Failed to sync node:', error)
    }
  }

  const updateNodeConfig = async (nodeId: string, config: Partial<EdgeNode['config']>) => {
    try {
      await fetch(`/api/edge/nodes/${nodeId}/config`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(config)
      })
      refreshEdgeData()
    } catch (error) {
      console.error('Failed to update node config:', error)
    }
  }

  const getNodeIcon = (type: string) => {
    switch (type) {
      case 'mobile': return <Smartphone className="w-4 h-4" />
      case 'iot': return <Cpu className="w-4 h-4" />
      case 'edge_server': return <HardDrive className="w-4 h-4" />
      case 'gateway': return <Network className="w-4 h-4" />
      default: return <Globe className="w-4 h-4" />
    }
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'online': return <CheckCircle className="w-4 h-4 text-green-500" />
      case 'offline': return <XCircle className="w-4 h-4 text-red-500" />
      case 'syncing': return <RefreshCw className="w-4 h-4 text-blue-500 animate-spin" />
      case 'error': return <AlertTriangle className="w-4 h-4 text-yellow-500" />
      default: return <Globe className="w-4 h-4" />
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'online': return 'success'
      case 'offline': return 'destructive'
      case 'syncing': return 'default'
      case 'error': return 'warning'
      default: return 'secondary'
    }
  }

  return (
    <div className="space-y-6">
      {/* Edge Overview */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <Card.Content className="p-4">
            <div className="flex items-center space-x-3">
              <Globe className="w-8 h-8 text-blue-500" />
              <div>
                <p className="text-sm text-muted-foreground">Total Nodes</p>
                <p className="text-2xl font-bold">{metrics?.totalNodes || 0}</p>
              </div>
            </div>
          </Card.Content>
        </Card>

        <Card>
          <Card.Content className="p-4">
            <div className="flex items-center space-x-3">
              <Activity className="w-8 h-8 text-green-500" />
              <div>
                <p className="text-sm text-muted-foreground">Online Nodes</p>
                <p className="text-2xl font-bold">{metrics?.onlineNodes || 0}</p>
              </div>
            </div>
          </Card.Content>
        </Card>

        <Card>
          <Card.Content className="p-4">
            <div className="flex items-center space-x-3">
              <Zap className="w-8 h-8 text-purple-500" />
              <div>
                <p className="text-sm text-muted-foreground">Requests</p>
                <p className="text-2xl font-bold">{formatNumber(metrics?.totalRequests || 0)}</p>
              </div>
            </div>
          </Card.Content>
        </Card>

        <Card>
          <Card.Content className="p-4">
            <div className="flex items-center space-x-3">
              <Clock className="w-8 h-8 text-orange-500" />
              <div>
                <p className="text-sm text-muted-foreground">Avg Latency</p>
                <p className="text-2xl font-bold">{metrics?.avgLatency || 0}ms</p>
              </div>
            </div>
          </Card.Content>
        </Card>
      </div>

      {/* Edge Nodes Management */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Nodes List */}
        <Card>
          <Card.Header>
            <Card.Title>Edge Nodes</Card.Title>
            <div className="flex space-x-2">
              <select
                className="px-3 py-1 border rounded text-sm"
                value={filterType}
                onChange={(e) => setFilterType(e.target.value)}
              >
                <option value="all">All Types</option>
                <option value="mobile">Mobile</option>
                <option value="iot">IoT</option>
                <option value="edge_server">Edge Server</option>
                <option value="gateway">Gateway</option>
              </select>
              <select
                className="px-3 py-1 border rounded text-sm"
                value={filterStatus}
                onChange={(e) => setFilterStatus(e.target.value)}
              >
                <option value="all">All Status</option>
                <option value="online">Online</option>
                <option value="offline">Offline</option>
                <option value="syncing">Syncing</option>
                <option value="error">Error</option>
              </select>
            </div>
          </Card.Header>
          
          <Card.Content>
            <div className="space-y-3 max-h-96 overflow-y-auto">
              {filteredNodes.map((node) => (
                <div
                  key={node.id}
                  className={`p-3 border rounded-lg cursor-pointer transition-colors ${
                    selectedNode?.id === node.id ? 'border-primary bg-primary/5' : 'hover:bg-gray-50'
                  }`}
                  onClick={() => setSelectedNode(node)}
                >
                  <div className="flex items-start justify-between">
                    <div className="flex items-start space-x-3">
                      {getNodeIcon(node.type)}
                      <div className="flex-1">
                        <h4 className="text-sm font-medium">{node.name}</h4>
                        <p className="text-xs text-muted-foreground">
                          {node.location.city}, {node.location.country}
                        </p>
                        <div className="flex items-center space-x-2 mt-1">
                          <Badge variant={getStatusColor(node.status)}>
                            {node.status}
                          </Badge>
                          <span className="text-xs text-muted-foreground">
                            {node.models.length} models
                          </span>
                        </div>
                      </div>
                    </div>
                    
                    <div className="flex flex-col items-end space-y-1">
                      {getStatusIcon(node.status)}
                      <div className="text-xs text-muted-foreground text-right">
                        <div>CPU: {node.metrics.cpuUsage}%</div>
                        <div>Mem: {node.metrics.memoryUsage}%</div>
                        {node.metrics.batteryLevel && (
                          <div>Bat: {node.metrics.batteryLevel}%</div>
                        )}
                      </div>
                    </div>
                  </div>
                  
                  <div className="mt-2 text-xs text-muted-foreground">
                    Last seen: {new Date(node.lastSeen).toLocaleString()}
                  </div>
                </div>
              ))}
              
              {filteredNodes.length === 0 && (
                <div className="text-center text-muted-foreground py-8">
                  No nodes match the current filters
                </div>
              )}
            </div>
          </Card.Content>
        </Card>

        {/* Node Details */}
        <Card>
          <Card.Header>
            <Card.Title>Node Details</Card.Title>
            {selectedNode && (
              <Button
                size="sm"
                onClick={() => syncNode(selectedNode.id)}
                disabled={selectedNode.status === 'syncing'}
              >
                <RefreshCw className={`w-4 h-4 mr-2 ${
                  selectedNode.status === 'syncing' ? 'animate-spin' : ''
                }`} />
                Sync
              </Button>
            )}
          </Card.Header>
          
          <Card.Content>
            {selectedNode ? (
              <div className="space-y-4">
                {/* Node Info */}
                <div>
                  <div className="flex items-center space-x-2 mb-2">
                    {getNodeIcon(selectedNode.type)}
                    <h3 className="font-semibold">{selectedNode.name}</h3>
                    <Badge variant={getStatusColor(selectedNode.status)}>
                      {selectedNode.status}
                    </Badge>
                  </div>
                  <p className="text-sm text-muted-foreground">
                    {selectedNode.location.city}, {selectedNode.location.country}
                  </p>
                </div>

                {/* Capabilities */}
                <div>
                  <h4 className="text-sm font-medium mb-2">Capabilities</h4>
                  <div className="grid grid-cols-2 gap-2 text-sm">
                    <div>CPU: {selectedNode.capabilities.cpu} cores</div>
                    <div>Memory: {selectedNode.capabilities.memory} GB</div>
                    <div>Storage: {formatBytes(selectedNode.capabilities.storage * 1024 * 1024 * 1024)}</div>
                    <div className="flex space-x-2">
                      {selectedNode.capabilities.gpu && <Badge size="sm">GPU</Badge>}
                      {selectedNode.capabilities.tpu && <Badge size="sm">TPU</Badge>}
                    </div>
                  </div>
                </div>

                {/* Metrics */}
                <div>
                  <h4 className="text-sm font-medium mb-2">Current Metrics</h4>
                  <div className="space-y-2">
                    <div className="flex justify-between">
                      <span className="text-sm">CPU Usage:</span>
                      <span className="text-sm">{selectedNode.metrics.cpuUsage}%</span>
                    </div>
                    <div className="w-full bg-gray-200 rounded-full h-2">
                      <div
                        className="bg-blue-600 h-2 rounded-full"
                        style={{ width: `${selectedNode.metrics.cpuUsage}%` }}
                      />
                    </div>
                    
                    <div className="flex justify-between">
                      <span className="text-sm">Memory Usage:</span>
                      <span className="text-sm">{selectedNode.metrics.memoryUsage}%</span>
                    </div>
                    <div className="w-full bg-gray-200 rounded-full h-2">
                      <div
                        className="bg-green-600 h-2 rounded-full"
                        style={{ width: `${selectedNode.metrics.memoryUsage}%` }}
                      />
                    </div>
                    
                    <div className="flex justify-between">
                      <span className="text-sm">Storage Usage:</span>
                      <span className="text-sm">{selectedNode.metrics.storageUsage}%</span>
                    </div>
                    <div className="w-full bg-gray-200 rounded-full h-2">
                      <div
                        className="bg-orange-600 h-2 rounded-full"
                        style={{ width: `${selectedNode.metrics.storageUsage}%` }}
                      />
                    </div>

                    {selectedNode.metrics.batteryLevel && (
                      <>
                        <div className="flex justify-between">
                          <span className="text-sm">Battery Level:</span>
                          <span className="text-sm">{selectedNode.metrics.batteryLevel}%</span>
                        </div>
                        <div className="w-full bg-gray-200 rounded-full h-2">
                          <div
                            className="bg-purple-600 h-2 rounded-full"
                            style={{ width: `${selectedNode.metrics.batteryLevel}%` }}
                          />
                        </div>
                      </>
                    )}
                  </div>
                </div>

                {/* Additional Metrics */}
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <span className="text-muted-foreground">Latency:</span>
                    <span className="ml-2">{selectedNode.metrics.networkLatency}ms</span>
                  </div>
                  <div>
                    <span className="text-muted-foreground">Temperature:</span>
                    <span className="ml-2">{selectedNode.metrics.temperature}°C</span>
                  </div>
                  <div>
                    <span className="text-muted-foreground">Uptime:</span>
                    <span className="ml-2">{formatDuration(selectedNode.metrics.uptime * 1000)}</span>
                  </div>
                  <div>
                    <span className="text-muted-foreground">Models:</span>
                    <span className="ml-2">{selectedNode.models.length}</span>
                  </div>
                </div>

                {/* Configuration */}
                <div>
                  <h4 className="text-sm font-medium mb-2">Configuration</h4>
                  <div className="space-y-2">
                    <div className="flex justify-between items-center">
                      <span className="text-sm">Auto Sync:</span>
                      <Switch
                        checked={selectedNode.config.autoSync}
                        onCheckedChange={(checked) => 
                          updateNodeConfig(selectedNode.id, { autoSync: checked })
                        }
                      />
                    </div>
                    <div className="flex justify-between items-center">
                      <span className="text-sm">Offline Mode:</span>
                      <Switch
                        checked={selectedNode.config.offlineMode}
                        onCheckedChange={(checked) => 
                          updateNodeConfig(selectedNode.id, { offlineMode: checked })
                        }
                      />
                    </div>
                    <div className="flex justify-between items-center">
                      <span className="text-sm">Compression:</span>
                      <Switch
                        checked={selectedNode.config.compressionEnabled}
                        onCheckedChange={(checked) => 
                          updateNodeConfig(selectedNode.id, { compressionEnabled: checked })
                        }
                      />
                    </div>
                  </div>
                </div>

                {/* Sync Status */}
                <div>
                  <h4 className="text-sm font-medium mb-2">Sync Status</h4>
                  <div className="space-y-1 text-sm">
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Last Sync:</span>
                      <span>{new Date(selectedNode.syncStatus.lastSync).toLocaleString()}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Pending Updates:</span>
                      <span>{selectedNode.syncStatus.pendingUpdates}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Failed Syncs:</span>
                      <span>{selectedNode.syncStatus.failedSyncs}</span>
                    </div>
                  </div>
                </div>
              </div>
            ) : (
              <div className="text-center text-muted-foreground py-8">
                Select a node to view details
              </div>
            )}
          </Card.Content>
        </Card>
      </div>

      {/* Deployments */}
      <Card>
        <Card.Header>
          <Card.Title>Edge Deployments</Card.Title>
          <Button onClick={() => setShowDeployModal(true)}>
            <Play className="w-4 h-4 mr-2" />
            Deploy Model
          </Button>
        </Card.Header>
        
        <Card.Content>
          <div className="space-y-3">
            {deployments.map((deployment) => (
              <div key={deployment.id} className="p-3 border rounded-lg">
                <div className="flex items-center justify-between">
                  <div>
                    <h4 className="text-sm font-medium">{deployment.name}</h4>
                    <p className="text-xs text-muted-foreground">
                      Model: {deployment.modelId} • Nodes: {deployment.targetNodes.length}
                    </p>
                  </div>
                  
                  <div className="flex items-center space-x-2">
                    <Badge variant={
                      deployment.status === 'deployed' ? 'success' :
                      deployment.status === 'failed' ? 'destructive' :
                      deployment.status === 'deploying' ? 'default' :
                      'secondary'
                    }>
                      {deployment.status}
                    </Badge>
                    <span className="text-sm">{deployment.progress}%</span>
                  </div>
                </div>
                
                {deployment.status === 'deploying' && (
                  <div className="mt-2 w-full bg-gray-200 rounded-full h-2">
                    <div
                      className="bg-blue-600 h-2 rounded-full transition-all duration-300"
                      style={{ width: `${deployment.progress}%` }}
                    />
                  </div>
                )}
              </div>
            ))}
            
            {deployments.length === 0 && (
              <div className="text-center text-muted-foreground py-8">
                No deployments found. Deploy a model to get started.
              </div>
            )}
          </div>
        </Card.Content>
      </Card>

      {/* Deploy Modal */}
      {showDeployModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <Card className="w-full max-w-md mx-4">
            <Card.Header>
              <Card.Title>Deploy Model to Edge</Card.Title>
            </Card.Header>
            
            <Card.Content className="space-y-4">
              <Input
                label="Deployment Name"
                value={newDeployment.name}
                onChange={(e) => setNewDeployment({ ...newDeployment, name: e.target.value })}
                placeholder="Enter deployment name"
              />

              <div>
                <label className="block text-sm font-medium mb-2">Model</label>
                <select
                  className="w-full px-3 py-2 border rounded-md"
                  value={newDeployment.modelId}
                  onChange={(e) => setNewDeployment({ ...newDeployment, modelId: e.target.value })}
                >
                  <option value="">Select a model...</option>
                  {availableModels?.map((model: any) => (
                    <option key={model.id} value={model.id}>
                      {model.name}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium mb-2">Target Nodes</label>
                <div className="max-h-32 overflow-y-auto space-y-1">
                  {nodes.filter(n => n.status === 'online').map((node) => (
                    <label key={node.id} className="flex items-center space-x-2">
                      <input
                        type="checkbox"
                        checked={newDeployment.targetNodes.includes(node.id)}
                        onChange={(e) => {
                          if (e.target.checked) {
                            setNewDeployment({
                              ...newDeployment,
                              targetNodes: [...newDeployment.targetNodes, node.id]
                            })
                          } else {
                            setNewDeployment({
                              ...newDeployment,
                              targetNodes: newDeployment.targetNodes.filter(n => n !== node.id)
                            })
                          }
                        }}
                      />
                      <span className="text-sm">{node.name}</span>
                    </label>
                  ))}
                </div>
              </div>

              <Input
                label="Replication Factor"
                type="number"
                value={newDeployment.replicationFactor}
                onChange={(e) => setNewDeployment({ 
                  ...newDeployment, 
                  replicationFactor: parseInt(e.target.value) 
                })}
                min={1}
                max={newDeployment.targetNodes.length}
              />

              <div className="space-y-2">
                <label className="flex items-center space-x-2">
                  <input
                    type="checkbox"
                    checked={newDeployment.autoFailover}
                    onChange={(e) => setNewDeployment({ 
                      ...newDeployment, 
                      autoFailover: e.target.checked 
                    })}
                  />
                  <span className="text-sm">Auto Failover</span>
                </label>
                
                <label className="flex items-center space-x-2">
                  <input
                    type="checkbox"
                    checked={newDeployment.loadBalancing}
                    onChange={(e) => setNewDeployment({ 
                      ...newDeployment, 
                      loadBalancing: e.target.checked 
                    })}
                  />
                  <span className="text-sm">Load Balancing</span>
                </label>
              </div>
            </Card.Content>
            
            <Card.Footer className="flex justify-end space-x-2">
              <Button
                variant="outline"
                onClick={() => setShowDeployModal(false)}
              >
                Cancel
              </Button>
              <Button
                onClick={deployModel}
                disabled={!newDeployment.name || !newDeployment.modelId || newDeployment.targetNodes.length === 0}
              >
                Deploy
              </Button>
            </Card.Footer>
          </Card>
        </div>
      )}
    </div>
  )
}

export default EdgeComputing