import React, { useState, useEffect } from 'react'
import { 
  Globe2, 
  Network, 
  MapPin, 
  Server,
  Link,
  Shield,
  Users,
  Activity,
  Settings,
  Plus,
  Minus,
  RefreshCw,
  AlertTriangle,
  CheckCircle,
  XCircle,
  Clock,
  BarChart3,
  Zap,
  Database,
  Monitor
} from 'lucide-react'
import { Button, Card, Input, Alert, Badge, Switch } from '@/design-system'
import { useWebSocket } from '@/hooks/useWebSocket'
import { useAPI } from '@/hooks/useAPI'
import { formatBytes, formatNumber, formatDuration } from '@/utils/format'

interface FederatedCluster {
  id: string
  name: string
  region: string
  provider: 'aws' | 'gcp' | 'azure' | 'on-premise'
  status: 'active' | 'joining' | 'leaving' | 'disconnected' | 'error'
  role: 'coordinator' | 'member' | 'observer'
  location: {
    latitude: number
    longitude: number
    city: string
    country: string
  }
  resources: {
    nodes: number
    cpu: number
    memory: number
    storage: number
    gpu: number
  }
  services: {
    total: number
    healthy: number
    unhealthy: number
    external: number
  }
  network: {
    latency: number
    bandwidth: number
    connections: number
    security: 'tls' | 'mtls' | 'vpn'
  }
  governance: {
    policies: number
    compliance: number
    violations: number
  }
  lastHeartbeat: string
  joinedAt: string
  version: string
}

interface FederationConfig {
  name: string
  coordinatorCluster: string
  consensus: 'raft' | 'gossip' | 'byzantine'
  replicationFactor: number
  quorumSize: number
  syncInterval: number
  failoverTimeout: number
  encryption: boolean
  crossRegionTraffic: boolean
  loadBalancing: {
    algorithm: 'round_robin' | 'least_connections' | 'weighted' | 'geographic'
    weights: Record<string, number>
  }
  serviceDiscovery: {
    protocol: 'dns' | 'consul' | 'etcd'
    ttl: number
    retries: number
  }
  security: {
    mutualTLS: boolean
    certificateRotation: number
    accessControl: boolean
  }
}

interface ServiceMapping {
  id: string
  name: string
  namespace: string
  type: 'internal' | 'external' | 'cross-cluster'
  endpoints: {
    clusterId: string
    address: string
    port: number
    healthy: boolean
  }[]
  traffic: {
    requests: number
    errors: number
    latency: number
  }
  policies: {
    loadBalancing: string
    failover: boolean
    retries: number
  }
}

interface FederationMetrics {
  totalClusters: number
  activeClusters: number
  totalServices: number
  crossClusterServices: number
  totalTraffic: number
  avgLatency: number
  errorRate: number
  dataTransferred: number
}

export const FederationManagement: React.FC = () => {
  const [clusters, setClusters] = useState<FederatedCluster[]>([])
  const [selectedCluster, setSelectedCluster] = useState<FederatedCluster | null>(null)
  const [services, setServices] = useState<ServiceMapping[]>([])
  const [config, setConfig] = useState<FederationConfig | null>(null)
  const [metrics, setMetrics] = useState<FederationMetrics | null>(null)
  const [activeTab, setActiveTab] = useState<'topology' | 'services' | 'config' | 'governance'>('topology')
  const [showJoinModal, setShowJoinModal] = useState(false)
  const [joinRequest, setJoinRequest] = useState({
    clusterEndpoint: '',
    authToken: '',
    role: 'member' as const
  })

  const { data: federationData, mutate: refreshFederation } = useAPI('/api/federation/overview')
  const { data: servicesData } = useAPI('/api/federation/services')
  
  const ws = useWebSocket('ws://localhost:8080/ws/federation', {
    onMessage: (data) => {
      switch (data.type) {
        case 'cluster_update':
          setClusters(prev => {
            const index = prev.findIndex(c => c.id === data.cluster.id)
            if (index >= 0) {
              const updated = [...prev]
              updated[index] = { ...updated[index], ...data.cluster }
              return updated
            }
            return [...prev, data.cluster]
          })
          break
        case 'service_update':
          setServices(prev => {
            const index = prev.findIndex(s => s.id === data.service.id)
            if (index >= 0) {
              const updated = [...prev]
              updated[index] = { ...updated[index], ...data.service }
              return updated
            }
            return [...prev, data.service]
          })
          break
        case 'metrics_update':
          setMetrics(data.metrics)
          break
      }
    }
  })

  useEffect(() => {
    if (federationData) {
      setClusters(federationData.clusters || [])
      setConfig(federationData.config || null)
      setMetrics(federationData.metrics || null)
    }
    if (servicesData) {
      setServices(servicesData.services || [])
    }
  }, [federationData, servicesData])

  const joinFederation = async () => {
    try {
      const response = await fetch('/api/federation/join', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(joinRequest)
      })
      
      if (!response.ok) throw new Error('Failed to join federation')
      
      refreshFederation()
      setShowJoinModal(false)
      setJoinRequest({ clusterEndpoint: '', authToken: '', role: 'member' })
    } catch (error) {
      console.error('Failed to join federation:', error)
    }
  }

  const leaveCluster = async (clusterId: string) => {
    try {
      await fetch(`/api/federation/clusters/${clusterId}/leave`, {
        method: 'POST'
      })
      refreshFederation()
    } catch (error) {
      console.error('Failed to leave cluster:', error)
    }
  }

  const updateConfig = async (newConfig: Partial<FederationConfig>) => {
    try {
      await fetch('/api/federation/config', {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(newConfig)
      })
      refreshFederation()
    } catch (error) {
      console.error('Failed to update config:', error)
    }
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'active': return <CheckCircle className="w-4 h-4 text-green-500" />
      case 'joining': return <RefreshCw className="w-4 h-4 text-blue-500 animate-spin" />
      case 'leaving': return <RefreshCw className="w-4 h-4 text-orange-500 animate-spin" />
      case 'disconnected': return <XCircle className="w-4 h-4 text-red-500" />
      case 'error': return <AlertTriangle className="w-4 h-4 text-yellow-500" />
      default: return <Globe2 className="w-4 h-4" />
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'success'
      case 'joining': case 'leaving': return 'default'
      case 'disconnected': case 'error': return 'destructive'
      default: return 'secondary'
    }
  }

  const getRoleColor = (role: string) => {
    switch (role) {
      case 'coordinator': return 'default'
      case 'member': return 'secondary'
      case 'observer': return 'outline'
      default: return 'secondary'
    }
  }

  return (
    <div className="space-y-6">
      {/* Federation Overview */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <Card.Content className="p-4">
            <div className="flex items-center space-x-3">
              <Globe2 className="w-8 h-8 text-blue-500" />
              <div>
                <p className="text-sm text-muted-foreground">Total Clusters</p>
                <p className="text-2xl font-bold">{metrics?.totalClusters || 0}</p>
              </div>
            </div>
          </Card.Content>
        </Card>

        <Card>
          <Card.Content className="p-4">
            <div className="flex items-center space-x-3">
              <Activity className="w-8 h-8 text-green-500" />
              <div>
                <p className="text-sm text-muted-foreground">Active Clusters</p>
                <p className="text-2xl font-bold">{metrics?.activeClusters || 0}</p>
              </div>
            </div>
          </Card.Content>
        </Card>

        <Card>
          <Card.Content className="p-4">
            <div className="flex items-center space-x-3">
              <Network className="w-8 h-8 text-purple-500" />
              <div>
                <p className="text-sm text-muted-foreground">Cross-Cluster Services</p>
                <p className="text-2xl font-bold">{metrics?.crossClusterServices || 0}</p>
              </div>
            </div>
          </Card.Content>
        </Card>

        <Card>
          <Card.Content className="p-4">
            <div className="flex items-center space-x-3">
              <Zap className="w-8 h-8 text-orange-500" />
              <div>
                <p className="text-sm text-muted-foreground">Avg Latency</p>
                <p className="text-2xl font-bold">{metrics?.avgLatency || 0}ms</p>
              </div>
            </div>
          </Card.Content>
        </Card>
      </div>

      {/* Main Dashboard */}
      <Card>
        <Card.Header>
          <Card.Title>Federation Dashboard</Card.Title>
          <div className="flex space-x-2">
            <Button onClick={() => setShowJoinModal(true)}>
              <Plus className="w-4 h-4 mr-2" />
              Join Federation
            </Button>
            <Button variant="outline" onClick={() => refreshFederation()}>
              <RefreshCw className="w-4 h-4 mr-2" />
              Refresh
            </Button>
          </div>
        </Card.Header>

        <Card.Content>
          {/* Tab Navigation */}
          <div className="border-b mb-6">
            <nav className="flex space-x-4">
              {['topology', 'services', 'config', 'governance'].map((tab) => (
                <button
                  key={tab}
                  className={`py-2 px-1 border-b-2 text-sm font-medium capitalize ${
                    activeTab === tab
                      ? 'border-primary text-primary'
                      : 'border-transparent text-muted-foreground hover:text-foreground'
                  }`}
                  onClick={() => setActiveTab(tab as any)}
                >
                  {tab}
                </button>
              ))}
            </nav>
          </div>

          {/* Topology View */}
          {activeTab === 'topology' && (
            <div className="space-y-6">
              {/* Cluster Topology Map */}
              <div className="h-96 bg-gray-50 rounded-lg p-4 relative">
                <h4 className="text-sm font-medium mb-4">Global Cluster Topology</h4>
                
                {/* World Map Visualization */}
                <div className="absolute inset-4 bg-blue-50 rounded border-2 border-dashed border-blue-200 flex items-center justify-center">
                  <div className="text-center text-muted-foreground">
                    <Globe2 className="w-12 h-12 mx-auto mb-2" />
                    <p>Interactive world map with cluster locations</p>
                    <p className="text-xs">Cluster positions based on geographic coordinates</p>
                  </div>
                </div>

                {/* Cluster Nodes */}
                {clusters.map((cluster, index) => (
                  <div
                    key={cluster.id}
                    className={`absolute w-4 h-4 rounded-full cursor-pointer ${
                      cluster.status === 'active' ? 'bg-green-500' :
                      cluster.status === 'joining' || cluster.status === 'leaving' ? 'bg-blue-500' :
                      'bg-red-500'
                    }`}
                    style={{
                      left: `${20 + (index % 3) * 25}%`,
                      top: `${30 + Math.floor(index / 3) * 20}%`
                    }}
                    onClick={() => setSelectedCluster(cluster)}
                    title={`${cluster.name} (${cluster.region})`}
                  />
                ))}
              </div>

              {/* Cluster List */}
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
                {clusters.map((cluster) => (
                  <div
                    key={cluster.id}
                    className={`p-4 border rounded-lg cursor-pointer transition-colors ${
                      selectedCluster?.id === cluster.id ? 'border-primary bg-primary/5' : 'hover:bg-gray-50'
                    }`}
                    onClick={() => setSelectedCluster(cluster)}
                  >
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center space-x-2 mb-2">
                          {getStatusIcon(cluster.status)}
                          <h4 className="text-sm font-medium">{cluster.name}</h4>
                          <Badge variant={getRoleColor(cluster.role)}>
                            {cluster.role}
                          </Badge>
                        </div>
                        
                        <p className="text-xs text-muted-foreground mb-2">
                          {cluster.location.city}, {cluster.location.country} • {cluster.provider}
                        </p>
                        
                        <div className="grid grid-cols-2 gap-2 text-xs">
                          <div>Nodes: {cluster.resources.nodes}</div>
                          <div>Services: {cluster.services.total}</div>
                          <div>CPU: {cluster.resources.cpu} cores</div>
                          <div>Memory: {formatBytes(cluster.resources.memory * 1024 * 1024 * 1024)}</div>
                        </div>
                      </div>
                      
                      <div className="flex flex-col items-end space-y-1">
                        <Badge variant={getStatusColor(cluster.status)}>
                          {cluster.status}
                        </Badge>
                        <span className="text-xs text-muted-foreground">
                          {cluster.network.latency}ms
                        </span>
                        {cluster.status !== 'coordinator' && (
                          <Button
                            size="sm"
                            variant="destructive"
                            onClick={(e) => {
                              e.stopPropagation()
                              leaveCluster(cluster.id)
                            }}
                          >
                            <Minus className="w-3 h-3" />
                          </Button>
                        )}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Services View */}
          {activeTab === 'services' && (
            <div className="space-y-4">
              <div className="flex justify-between items-center">
                <h4 className="text-sm font-medium">Cross-Cluster Services ({services.length})</h4>
              </div>
              
              <div className="space-y-3">
                {services.map((service) => (
                  <div key={service.id} className="p-4 border rounded-lg">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center space-x-2 mb-2">
                          <Network className="w-4 h-4" />
                          <h4 className="text-sm font-medium">{service.name}</h4>
                          <Badge variant={service.type === 'cross-cluster' ? 'default' : 'secondary'}>
                            {service.type}
                          </Badge>
                        </div>
                        
                        <p className="text-xs text-muted-foreground mb-2">
                          Namespace: {service.namespace}
                        </p>
                        
                        <div className="grid grid-cols-3 gap-4 text-xs">
                          <div>
                            <span className="text-muted-foreground">Endpoints:</span>
                            <span className="ml-1">{service.endpoints.length}</span>
                          </div>
                          <div>
                            <span className="text-muted-foreground">Requests:</span>
                            <span className="ml-1">{formatNumber(service.traffic.requests)}</span>
                          </div>
                          <div>
                            <span className="text-muted-foreground">Latency:</span>
                            <span className="ml-1">{service.traffic.latency}ms</span>
                          </div>
                        </div>
                      </div>
                      
                      <div className="text-xs text-muted-foreground text-right">
                        <div>Error Rate: {((service.traffic.errors / service.traffic.requests) * 100).toFixed(2)}%</div>
                        <div className="mt-1">
                          {service.endpoints.filter(e => e.healthy).length}/{service.endpoints.length} healthy
                        </div>
                      </div>
                    </div>
                    
                    {/* Endpoints */}
                    <div className="mt-3 pt-3 border-t">
                      <div className="text-xs font-medium mb-2">Endpoints:</div>
                      <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
                        {service.endpoints.map((endpoint, index) => (
                          <div key={index} className="flex items-center justify-between text-xs p-2 bg-gray-50 rounded">
                            <span>{endpoint.address}:{endpoint.port}</span>
                            <div className="flex items-center space-x-2">
                              <Badge size="sm" variant={endpoint.healthy ? 'success' : 'destructive'}>
                                {endpoint.healthy ? 'healthy' : 'unhealthy'}
                              </Badge>
                              <span className="text-muted-foreground">
                                {clusters.find(c => c.id === endpoint.clusterId)?.name}
                              </span>
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  </div>
                ))}
                
                {services.length === 0 && (
                  <div className="text-center text-muted-foreground py-8">
                    No cross-cluster services found
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Configuration View */}
          {activeTab === 'config' && config && (
            <div className="space-y-6">
              <div>
                <h4 className="text-sm font-medium mb-4">Federation Configuration</h4>
                
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <Input
                    label="Federation Name"
                    value={config.name}
                    onChange={(e) => setConfig({ ...config, name: e.target.value })}
                  />
                  
                  <div>
                    <label className="block text-sm font-medium mb-2">Consensus Algorithm</label>
                    <select
                      className="w-full px-3 py-2 border rounded-md"
                      value={config.consensus}
                      onChange={(e) => setConfig({ ...config, consensus: e.target.value as any })}
                    >
                      <option value="raft">Raft</option>
                      <option value="gossip">Gossip</option>
                      <option value="byzantine">Byzantine</option>
                    </select>
                  </div>
                  
                  <Input
                    label="Replication Factor"
                    type="number"
                    value={config.replicationFactor}
                    onChange={(e) => setConfig({ ...config, replicationFactor: parseInt(e.target.value) })}
                    min={1}
                    max={10}
                  />
                  
                  <Input
                    label="Quorum Size"
                    type="number"
                    value={config.quorumSize}
                    onChange={(e) => setConfig({ ...config, quorumSize: parseInt(e.target.value) })}
                    min={1}
                    max={clusters.length}
                  />
                  
                  <Input
                    label="Sync Interval (seconds)"
                    type="number"
                    value={config.syncInterval}
                    onChange={(e) => setConfig({ ...config, syncInterval: parseInt(e.target.value) })}
                    min={1}
                    max={3600}
                  />
                  
                  <Input
                    label="Failover Timeout (seconds)"
                    type="number"
                    value={config.failoverTimeout}
                    onChange={(e) => setConfig({ ...config, failoverTimeout: parseInt(e.target.value) })}
                    min={5}
                    max={300}
                  />
                </div>
                
                <div className="mt-4 space-y-3">
                  <div className="flex justify-between items-center">
                    <span className="text-sm">Encryption Enabled</span>
                    <Switch
                      checked={config.encryption}
                      onCheckedChange={(checked) => setConfig({ ...config, encryption: checked })}
                    />
                  </div>
                  
                  <div className="flex justify-between items-center">
                    <span className="text-sm">Cross-Region Traffic</span>
                    <Switch
                      checked={config.crossRegionTraffic}
                      onCheckedChange={(checked) => setConfig({ ...config, crossRegionTraffic: checked })}
                    />
                  </div>
                  
                  <div className="flex justify-between items-center">
                    <span className="text-sm">Mutual TLS</span>
                    <Switch
                      checked={config.security.mutualTLS}
                      onCheckedChange={(checked) => setConfig({ 
                        ...config, 
                        security: { ...config.security, mutualTLS: checked }
                      })}
                    />
                  </div>
                </div>
                
                <div className="mt-6">
                  <Button onClick={() => updateConfig(config)}>
                    <Settings className="w-4 h-4 mr-2" />
                    Update Configuration
                  </Button>
                </div>
              </div>
            </div>
          )}

          {/* Governance View */}
          {activeTab === 'governance' && (
            <div className="space-y-6">
              <div>
                <h4 className="text-sm font-medium mb-4">Federation Governance</h4>
                
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
                  <Card>
                    <Card.Content className="p-4">
                      <div className="flex items-center space-x-3">
                        <Shield className="w-6 h-6 text-blue-500" />
                        <div>
                          <p className="text-sm text-muted-foreground">Active Policies</p>
                          <p className="text-xl font-bold">
                            {clusters.reduce((sum, c) => sum + c.governance.policies, 0)}
                          </p>
                        </div>
                      </div>
                    </Card.Content>
                  </Card>
                  
                  <Card>
                    <Card.Content className="p-4">
                      <div className="flex items-center space-x-3">
                        <CheckCircle className="w-6 h-6 text-green-500" />
                        <div>
                          <p className="text-sm text-muted-foreground">Compliance Score</p>
                          <p className="text-xl font-bold">
                            {Math.round(clusters.reduce((sum, c) => sum + c.governance.compliance, 0) / clusters.length || 0)}%
                          </p>
                        </div>
                      </div>
                    </Card.Content>
                  </Card>
                  
                  <Card>
                    <Card.Content className="p-4">
                      <div className="flex items-center space-x-3">
                        <AlertTriangle className="w-6 h-6 text-red-500" />
                        <div>
                          <p className="text-sm text-muted-foreground">Violations</p>
                          <p className="text-xl font-bold">
                            {clusters.reduce((sum, c) => sum + c.governance.violations, 0)}
                          </p>
                        </div>
                      </div>
                    </Card.Content>
                  </Card>
                </div>
                
                {/* Cluster Governance Details */}
                <div className="space-y-3">
                  {clusters.map((cluster) => (
                    <div key={cluster.id} className="p-4 border rounded-lg">
                      <div className="flex items-center justify-between mb-3">
                        <div className="flex items-center space-x-2">
                          <h4 className="text-sm font-medium">{cluster.name}</h4>
                          <Badge variant={getRoleColor(cluster.role)}>
                            {cluster.role}
                          </Badge>
                        </div>
                        <div className="text-xs text-muted-foreground">
                          Last updated: {new Date(cluster.lastHeartbeat).toLocaleString()}
                        </div>
                      </div>
                      
                      <div className="grid grid-cols-3 gap-4 text-sm">
                        <div>
                          <span className="text-muted-foreground">Policies:</span>
                          <span className="ml-2">{cluster.governance.policies}</span>
                        </div>
                        <div>
                          <span className="text-muted-foreground">Compliance:</span>
                          <span className="ml-2">{cluster.governance.compliance}%</span>
                        </div>
                        <div>
                          <span className="text-muted-foreground">Violations:</span>
                          <span className="ml-2 text-red-500">{cluster.governance.violations}</span>
                        </div>
                      </div>
                      
                      <div className="mt-2">
                        <div className="w-full bg-gray-200 rounded-full h-2">
                          <div
                            className="bg-green-600 h-2 rounded-full"
                            style={{ width: `${cluster.governance.compliance}%` }}
                          />
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          )}
        </Card.Content>
      </Card>

      {/* Cluster Details Sidebar */}
      {selectedCluster && (
        <Card>
          <Card.Header>
            <Card.Title>Cluster Details: {selectedCluster.name}</Card.Title>
          </Card.Header>
          
          <Card.Content className="space-y-4">
            <div>
              <div className="flex items-center space-x-2 mb-2">
                {getStatusIcon(selectedCluster.status)}
                <h3 className="font-semibold">{selectedCluster.name}</h3>
                <Badge variant={getStatusColor(selectedCluster.status)}>
                  {selectedCluster.status}
                </Badge>
                <Badge variant={getRoleColor(selectedCluster.role)}>
                  {selectedCluster.role}
                </Badge>
              </div>
              
              <p className="text-sm text-muted-foreground">
                {selectedCluster.location.city}, {selectedCluster.location.country} • {selectedCluster.provider}
              </p>
            </div>

            <div>
              <h4 className="text-sm font-medium mb-2">Resources</h4>
              <div className="grid grid-cols-2 gap-2 text-sm">
                <div>Nodes: {selectedCluster.resources.nodes}</div>
                <div>CPU: {selectedCluster.resources.cpu} cores</div>
                <div>Memory: {formatBytes(selectedCluster.resources.memory * 1024 * 1024 * 1024)}</div>
                <div>Storage: {formatBytes(selectedCluster.resources.storage * 1024 * 1024 * 1024)}</div>
                {selectedCluster.resources.gpu > 0 && (
                  <div>GPU: {selectedCluster.resources.gpu}</div>
                )}
              </div>
            </div>

            <div>
              <h4 className="text-sm font-medium mb-2">Services</h4>
              <div className="grid grid-cols-2 gap-2 text-sm">
                <div>Total: {selectedCluster.services.total}</div>
                <div>Healthy: {selectedCluster.services.healthy}</div>
                <div>Unhealthy: {selectedCluster.services.unhealthy}</div>
                <div>External: {selectedCluster.services.external}</div>
              </div>
            </div>

            <div>
              <h4 className="text-sm font-medium mb-2">Network</h4>
              <div className="grid grid-cols-2 gap-2 text-sm">
                <div>Latency: {selectedCluster.network.latency}ms</div>
                <div>Bandwidth: {formatBytes(selectedCluster.network.bandwidth * 1024 * 1024)}/s</div>
                <div>Connections: {selectedCluster.network.connections}</div>
                <div>Security: {selectedCluster.network.security.toUpperCase()}</div>
              </div>
            </div>

            <div className="text-xs text-muted-foreground">
              <div>Joined: {new Date(selectedCluster.joinedAt).toLocaleString()}</div>
              <div>Version: {selectedCluster.version}</div>
              <div>Last heartbeat: {new Date(selectedCluster.lastHeartbeat).toLocaleString()}</div>
            </div>
          </Card.Content>
        </Card>
      )}

      {/* Join Federation Modal */}
      {showJoinModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <Card className="w-full max-w-md mx-4">
            <Card.Header>
              <Card.Title>Join Federation</Card.Title>
            </Card.Header>
            
            <Card.Content className="space-y-4">
              <Input
                label="Cluster Endpoint"
                value={joinRequest.clusterEndpoint}
                onChange={(e) => setJoinRequest({ ...joinRequest, clusterEndpoint: e.target.value })}
                placeholder="https://cluster.example.com:8443"
              />

              <Input
                label="Authentication Token"
                type="password"
                value={joinRequest.authToken}
                onChange={(e) => setJoinRequest({ ...joinRequest, authToken: e.target.value })}
                placeholder="Enter authentication token"
              />

              <div>
                <label className="block text-sm font-medium mb-2">Role</label>
                <select
                  className="w-full px-3 py-2 border rounded-md"
                  value={joinRequest.role}
                  onChange={(e) => setJoinRequest({ ...joinRequest, role: e.target.value as any })}
                >
                  <option value="member">Member</option>
                  <option value="observer">Observer</option>
                </select>
              </div>
            </Card.Content>
            
            <Card.Footer className="flex justify-end space-x-2">
              <Button
                variant="outline"
                onClick={() => setShowJoinModal(false)}
              >
                Cancel
              </Button>
              <Button
                onClick={joinFederation}
                disabled={!joinRequest.clusterEndpoint || !joinRequest.authToken}
              >
                Join Federation
              </Button>
            </Card.Footer>
          </Card>
        </div>
      )}
    </div>
  )
}

export default FederationManagement