import React, { useState, useEffect } from 'react'
import { 
  Network, 
  Route, 
  Shield, 
  Activity,
  Zap,
  GitBranch,
  Target,
  BarChart3,
  Clock,
  AlertTriangle,
  CheckCircle,
  Settings,
  Play,
  Pause,
  RotateCcw,
  TrendingUp,
  Eye,
  Lock,
  Globe,
  Server,
  Database,
  RefreshCw,
  Filter,
  Search,
  Download,
  Upload
} from 'lucide-react'
import { Button, Card, Input, Alert, Badge, Switch, Select } from '@/design-system'
import { useWebSocket } from '@/hooks/useWebSocket'
import { useAPI } from '@/hooks/useAPI'
import { formatBytes, formatNumber, formatDuration } from '@/utils/format'

interface Service {
  id: string
  name: string
  namespace: string
  version: string
  status: 'healthy' | 'degraded' | 'down' | 'unknown'
  type: 'http' | 'grpc' | 'tcp' | 'redis' | 'database'
  endpoints: ServiceEndpoint[]
  replicas: {
    desired: number
    current: number
    ready: number
  }
  resources: {
    cpu: number
    memory: number
    requests: number
  }
  metrics: {
    requestRate: number
    errorRate: number
    latency: number
    successRate: number
  }
  dependencies: string[]
  dependents: string[]
  policies: string[]
  lastDeployed: string
}

interface ServiceEndpoint {
  id: string
  address: string
  port: number
  protocol: string
  healthy: boolean
  weight: number
  zone: string
}

interface TrafficPolicy {
  id: string
  name: string
  service: string
  type: 'load_balancing' | 'circuit_breaker' | 'retry' | 'timeout' | 'rate_limit'
  enabled: boolean
  config: {
    algorithm?: 'round_robin' | 'least_connections' | 'consistent_hash' | 'random'
    healthCheck?: {
      path: string
      interval: number
      timeout: number
      retries: number
    }
    circuitBreaker?: {
      errorThreshold: number
      requestVolumeThreshold: number
      sleepWindow: number
    }
    retry?: {
      attempts: number
      perTryTimeout: number
      retryOn: string[]
    }
    timeout?: {
      request: number
      idle: number
    }
    rateLimit?: {
      requestsPerUnit: number
      unit: 'second' | 'minute' | 'hour'
      burstSize: number
    }
  }
  createdAt: string
  updatedAt: string
}

interface Deployment {
  id: string
  service: string
  strategy: 'canary' | 'blue_green' | 'rolling' | 'recreate'
  status: 'pending' | 'in_progress' | 'completed' | 'failed' | 'rollback'
  progress: number
  versions: {
    current: string
    target: string
  }
  traffic: {
    current: number
    target: number
  }
  metrics: {
    errorRate: number
    latency: number
    throughput: number
  }
  config: {
    steps?: number
    interval?: number
    successThreshold?: number
    failureThreshold?: number
    autoRollback?: boolean
  }
  startTime: string
  endTime?: string
}

interface SecurityPolicy {
  id: string
  name: string
  type: 'authorization' | 'authentication' | 'mtls' | 'network'
  scope: {
    services: string[]
    namespaces: string[]
  }
  rules: {
    allow?: string[]
    deny?: string[]
    require?: string[]
  }
  enabled: boolean
  priority: number
  createdAt: string
}

interface MeshMetrics {
  services: {
    total: number
    healthy: number
    degraded: number
    down: number
  }
  traffic: {
    requestsPerSecond: number
    bytesPerSecond: number
    totalRequests: number
    totalBytes: number
  }
  performance: {
    avgLatency: number
    p99Latency: number
    errorRate: number
    successRate: number
  }
  security: {
    mtlsEnabled: number
    policiesActive: number
    violations: number
  }
}

export const ServiceMeshDashboard: React.FC = () => {
  const [services, setServices] = useState<Service[]>([])
  const [selectedService, setSelectedService] = useState<Service | null>(null)
  const [policies, setPolicies] = useState<TrafficPolicy[]>([])
  const [deployments, setDeployments] = useState<Deployment[]>([])
  const [securityPolicies, setSecurityPolicies] = useState<SecurityPolicy[]>([])
  const [metrics, setMetrics] = useState<MeshMetrics | null>(null)
  const [activeTab, setActiveTab] = useState<'topology' | 'traffic' | 'security' | 'deployments'>('topology')
  const [filterStatus, setFilterStatus] = useState<string>('all')
  const [filterNamespace, setFilterNamespace] = useState<string>('all')
  const [searchQuery, setSearchQuery] = useState('')
  const [showTopology, setShowTopology] = useState(true)

  const { data: meshData, mutate: refreshMesh } = useAPI('/api/service-mesh/overview')
  
  const ws = useWebSocket('ws://localhost:8080/ws/service-mesh', {
    onMessage: (data) => {
      switch (data.type) {
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
    if (meshData) {
      setServices(meshData.services || [])
      setPolicies(meshData.policies || [])
      setDeployments(meshData.deployments || [])
      setSecurityPolicies(meshData.securityPolicies || [])
      setMetrics(meshData.metrics || null)
    }
  }, [meshData])

  const createCanaryDeployment = async (serviceId: string, targetVersion: string) => {
    try {
      const deployment = {
        service: serviceId,
        strategy: 'canary',
        versions: {
          target: targetVersion
        },
        config: {
          steps: 5,
          interval: 300,
          successThreshold: 95,
          failureThreshold: 5,
          autoRollback: true
        }
      }

      const response = await fetch('/api/service-mesh/deployments', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(deployment)
      })
      
      if (!response.ok) throw new Error('Failed to create deployment')
      
      refreshMesh()
    } catch (error) {
      console.error('Failed to create deployment:', error)
    }
  }

  const togglePolicy = async (policyId: string, enabled: boolean) => {
    try {
      await fetch(`/api/service-mesh/policies/${policyId}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ enabled })
      })
      refreshMesh()
    } catch (error) {
      console.error('Failed to toggle policy:', error)
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'healthy': case 'completed': return 'success'
      case 'degraded': case 'in_progress': return 'warning'
      case 'down': case 'failed': return 'destructive'
      case 'pending': return 'default'
      default: return 'secondary'
    }
  }

  const getServiceIcon = (type: string) => {
    switch (type) {
      case 'http': return <Globe className="w-4 h-4" />
      case 'grpc': return <Zap className="w-4 h-4" />
      case 'tcp': return <Network className="w-4 h-4" />
      case 'redis': return <Database className="w-4 h-4" />
      case 'database': return <Server className="w-4 h-4" />
      default: return <Activity className="w-4 h-4" />
    }
  }

  const filteredServices = services.filter(service => {
    const statusMatch = filterStatus === 'all' || service.status === filterStatus
    const namespaceMatch = filterNamespace === 'all' || service.namespace === filterNamespace
    const searchMatch = searchQuery === '' || 
      service.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      service.namespace.toLowerCase().includes(searchQuery.toLowerCase())
    return statusMatch && namespaceMatch && searchMatch
  })

  const namespaces = [...new Set(services.map(s => s.namespace))]

  return (
    <div className="space-y-6">
      {/* Mesh Overview */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <Card.Content className="p-4">
            <div className="flex items-center space-x-3">
              <Network className="w-8 h-8 text-blue-500" />
              <div>
                <p className="text-sm text-muted-foreground">Total Services</p>
                <p className="text-2xl font-bold">{metrics?.services.total || 0}</p>
              </div>
            </div>
          </Card.Content>
        </Card>

        <Card>
          <Card.Content className="p-4">
            <div className="flex items-center space-x-3">
              <Activity className="w-8 h-8 text-green-500" />
              <div>
                <p className="text-sm text-muted-foreground">Request Rate</p>
                <p className="text-2xl font-bold">{formatNumber(metrics?.traffic.requestsPerSecond || 0)}/s</p>
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
                <p className="text-2xl font-bold">{metrics?.performance.avgLatency || 0}ms</p>
              </div>
            </div>
          </Card.Content>
        </Card>

        <Card>
          <Card.Content className="p-4">
            <div className="flex items-center space-x-3">
              <Shield className="w-8 h-8 text-purple-500" />
              <div>
                <p className="text-sm text-muted-foreground">Success Rate</p>
                <p className="text-2xl font-bold">{(metrics?.performance.successRate || 0).toFixed(1)}%</p>
              </div>
            </div>
          </Card.Content>
        </Card>
      </div>

      {/* Main Dashboard */}
      <Card>
        <Card.Header>
          <Card.Title>Service Mesh Dashboard</Card.Title>
          <div className="flex space-x-2">
            <Button variant="outline" onClick={() => refreshMesh()}>
              <RefreshCw className="w-4 h-4 mr-2" />
              Refresh
            </Button>
            <Button variant="outline">
              <Download className="w-4 h-4 mr-2" />
              Export Config
            </Button>
          </div>
        </Card.Header>

        <Card.Content>
          {/* Tab Navigation */}
          <div className="border-b mb-6">
            <nav className="flex space-x-4">
              {['topology', 'traffic', 'security', 'deployments'].map((tab) => (
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

          {/* Topology Tab */}
          {activeTab === 'topology' && (
            <div className="space-y-6">
              {/* Filters */}
              <div className="flex space-x-4">
                <div className="flex-1">
                  <Input
                    placeholder="Search services..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    leftIcon={<Search className="w-4 h-4" />}
                  />
                </div>
                <select
                  className="px-3 py-2 border rounded"
                  value={filterStatus}
                  onChange={(e) => setFilterStatus(e.target.value)}
                >
                  <option value="all">All Status</option>
                  <option value="healthy">Healthy</option>
                  <option value="degraded">Degraded</option>
                  <option value="down">Down</option>
                  <option value="unknown">Unknown</option>
                </select>
                <select
                  className="px-3 py-2 border rounded"
                  value={filterNamespace}
                  onChange={(e) => setFilterNamespace(e.target.value)}
                >
                  <option value="all">All Namespaces</option>
                  {namespaces.map(ns => (
                    <option key={ns} value={ns}>{ns}</option>
                  ))}
                </select>
                <Button
                  variant="outline"
                  onClick={() => setShowTopology(!showTopology)}
                >
                  {showTopology ? <Eye className="w-4 h-4 mr-2" /> : <Eye className="w-4 h-4 mr-2" />}
                  {showTopology ? 'Hide' : 'Show'} Topology
                </Button>
              </div>

              {/* Service Topology */}
              {showTopology && (
                <div className="h-96 bg-gray-50 rounded-lg p-4 relative">
                  <h4 className="text-sm font-medium mb-4">Service Topology</h4>
                  
                  {/* Interactive service mesh visualization */}
                  <div className="absolute inset-4 bg-blue-50 rounded border-2 border-dashed border-blue-200 flex items-center justify-center">
                    <div className="text-center text-muted-foreground">
                      <Network className="w-12 h-12 mx-auto mb-2" />
                      <p>Interactive service mesh topology</p>
                      <p className="text-xs">Services connected by traffic flows</p>
                    </div>
                  </div>

                  {/* Service nodes would be positioned here */}
                  {filteredServices.slice(0, 6).map((service, index) => (
                    <div
                      key={service.id}
                      className={`absolute w-20 h-20 rounded-lg border-2 cursor-pointer flex items-center justify-center ${
                        service.status === 'healthy' ? 'bg-green-100 border-green-500' :
                        service.status === 'degraded' ? 'bg-yellow-100 border-yellow-500' :
                        'bg-red-100 border-red-500'
                      }`}
                      style={{
                        left: `${15 + (index % 3) * 30}%`,
                        top: `${25 + Math.floor(index / 3) * 40}%`
                      }}
                      onClick={() => setSelectedService(service)}
                      title={service.name}
                    >
                      <div className="text-center">
                        {getServiceIcon(service.type)}
                        <div className="text-xs font-medium mt-1">{service.name.substring(0, 8)}</div>
                      </div>
                    </div>
                  ))}
                </div>
              )}

              {/* Services Grid */}
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {filteredServices.map((service) => (
                  <div
                    key={service.id}
                    className={`p-4 border rounded-lg cursor-pointer transition-colors ${
                      selectedService?.id === service.id ? 'border-primary bg-primary/5' : 'hover:bg-gray-50'
                    }`}
                    onClick={() => setSelectedService(service)}
                  >
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center space-x-2 mb-2">
                          {getServiceIcon(service.type)}
                          <h4 className="text-sm font-medium">{service.name}</h4>
                          <Badge variant={getStatusColor(service.status)}>
                            {service.status}
                          </Badge>
                        </div>
                        
                        <p className="text-xs text-muted-foreground mb-2">
                          {service.namespace} • v{service.version}
                        </p>
                        
                        <div className="grid grid-cols-2 gap-2 text-xs">
                          <div>Replicas: {service.replicas.ready}/{service.replicas.desired}</div>
                          <div>RPS: {formatNumber(service.metrics.requestRate)}</div>
                          <div>Latency: {service.metrics.latency}ms</div>
                          <div>Errors: {service.metrics.errorRate.toFixed(2)}%</div>
                        </div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Traffic Management Tab */}
          {activeTab === 'traffic' && (
            <div className="space-y-6">
              <div className="flex justify-between items-center">
                <h4 className="text-sm font-medium">Traffic Policies ({policies.length})</h4>
                <Button>
                  <Settings className="w-4 h-4 mr-2" />
                  Create Policy
                </Button>
              </div>
              
              <div className="space-y-4">
                {policies.map((policy) => (
                  <div key={policy.id} className="p-4 border rounded-lg">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center space-x-2 mb-2">
                          <Route className="w-4 h-4" />
                          <h4 className="text-sm font-medium">{policy.name}</h4>
                          <Badge variant="outline">{policy.type}</Badge>
                          <Badge variant={policy.enabled ? 'success' : 'secondary'}>
                            {policy.enabled ? 'enabled' : 'disabled'}
                          </Badge>
                        </div>
                        
                        <p className="text-xs text-muted-foreground mb-2">
                          Service: {policy.service}
                        </p>
                        
                        {/* Policy Configuration */}
                        <div className="mt-3 p-3 bg-gray-50 rounded text-xs">
                          {policy.type === 'load_balancing' && policy.config.algorithm && (
                            <div>Algorithm: {policy.config.algorithm}</div>
                          )}
                          {policy.type === 'circuit_breaker' && policy.config.circuitBreaker && (
                            <div>
                              Error Threshold: {policy.config.circuitBreaker.errorThreshold}% • 
                              Sleep Window: {policy.config.circuitBreaker.sleepWindow}s
                            </div>
                          )}
                          {policy.type === 'retry' && policy.config.retry && (
                            <div>
                              Max Attempts: {policy.config.retry.attempts} • 
                              Timeout: {policy.config.retry.perTryTimeout}ms
                            </div>
                          )}
                          {policy.type === 'rate_limit' && policy.config.rateLimit && (
                            <div>
                              Limit: {policy.config.rateLimit.requestsPerUnit}/{policy.config.rateLimit.unit} • 
                              Burst: {policy.config.rateLimit.burstSize}
                            </div>
                          )}
                        </div>
                      </div>
                      
                      <div className="flex items-center space-x-2">
                        <Switch
                          checked={policy.enabled}
                          onCheckedChange={(enabled) => togglePolicy(policy.id, enabled)}
                        />
                        <Button size="sm" variant="outline">
                          <Settings className="w-3 h-3" />
                        </Button>
                      </div>
                    </div>
                  </div>
                ))}
                
                {policies.length === 0 && (
                  <div className="text-center text-muted-foreground py-8">
                    No traffic policies configured
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Security Tab */}
          {activeTab === 'security' && (
            <div className="space-y-6">
              {/* Security Overview */}
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <Card>
                  <Card.Content className="p-4">
                    <div className="flex items-center space-x-3">
                      <Lock className="w-6 h-6 text-green-500" />
                      <div>
                        <p className="text-sm text-muted-foreground">mTLS Enabled</p>
                        <p className="text-xl font-bold">{metrics?.security.mtlsEnabled || 0}</p>
                      </div>
                    </div>
                  </Card.Content>
                </Card>
                
                <Card>
                  <Card.Content className="p-4">
                    <div className="flex items-center space-x-3">
                      <Shield className="w-6 h-6 text-blue-500" />
                      <div>
                        <p className="text-sm text-muted-foreground">Active Policies</p>
                        <p className="text-xl font-bold">{metrics?.security.policiesActive || 0}</p>
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
                        <p className="text-xl font-bold">{metrics?.security.violations || 0}</p>
                      </div>
                    </div>
                  </Card.Content>
                </Card>
              </div>

              {/* Security Policies */}
              <div>
                <div className="flex justify-between items-center mb-4">
                  <h4 className="text-sm font-medium">Security Policies ({securityPolicies.length})</h4>
                  <Button>
                    <Shield className="w-4 h-4 mr-2" />
                    Create Policy
                  </Button>
                </div>
                
                <div className="space-y-3">
                  {securityPolicies.map((policy) => (
                    <div key={policy.id} className="p-4 border rounded-lg">
                      <div className="flex items-start justify-between">
                        <div className="flex-1">
                          <div className="flex items-center space-x-2 mb-2">
                            <Shield className="w-4 h-4" />
                            <h4 className="text-sm font-medium">{policy.name}</h4>
                            <Badge variant="outline">{policy.type}</Badge>
                            <Badge variant={policy.enabled ? 'success' : 'secondary'}>
                              {policy.enabled ? 'enabled' : 'disabled'}
                            </Badge>
                          </div>
                          
                          <div className="text-xs text-muted-foreground mb-2">
                            <div>Services: {policy.scope.services.join(', ')}</div>
                            <div>Namespaces: {policy.scope.namespaces.join(', ')}</div>
                          </div>
                          
                          <div className="text-xs">
                            <div>Priority: {policy.priority}</div>
                            <div>Created: {new Date(policy.createdAt).toLocaleDateString()}</div>
                          </div>
                        </div>
                        
                        <div className="flex items-center space-x-2">
                          <Switch
                            checked={policy.enabled}
                            onCheckedChange={(enabled) => {
                              // Update security policy
                            }}
                          />
                          <Button size="sm" variant="outline">
                            <Settings className="w-3 h-3" />
                          </Button>
                        </div>
                      </div>
                    </div>
                  ))}
                  
                  {securityPolicies.length === 0 && (
                    <div className="text-center text-muted-foreground py-8">
                      No security policies configured
                    </div>
                  )}
                </div>
              </div>
            </div>
          )}

          {/* Deployments Tab */}
          {activeTab === 'deployments' && (
            <div className="space-y-6">
              <div className="flex justify-between items-center">
                <h4 className="text-sm font-medium">Active Deployments ({deployments.length})</h4>
                <Button>
                  <GitBranch className="w-4 h-4 mr-2" />
                  New Deployment
                </Button>
              </div>
              
              <div className="space-y-4">
                {deployments.map((deployment) => (
                  <div key={deployment.id} className="p-4 border rounded-lg">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center space-x-2 mb-2">
                          <GitBranch className="w-4 h-4" />
                          <h4 className="text-sm font-medium">{deployment.service}</h4>
                          <Badge variant="outline">{deployment.strategy}</Badge>
                          <Badge variant={getStatusColor(deployment.status)}>
                            {deployment.status}
                          </Badge>
                        </div>
                        
                        <div className="grid grid-cols-2 md:grid-cols-4 gap-2 text-xs mb-3">
                          <div>Current: v{deployment.versions.current}</div>
                          <div>Target: v{deployment.versions.target}</div>
                          <div>Traffic: {deployment.traffic.current}% → {deployment.traffic.target}%</div>
                          <div>Progress: {deployment.progress}%</div>
                        </div>
                        
                        {deployment.status === 'in_progress' && (
                          <div className="w-full bg-gray-200 rounded-full h-2 mb-2">
                            <div
                              className="bg-blue-600 h-2 rounded-full transition-all duration-300"
                              style={{ width: `${deployment.progress}%` }}
                            />
                          </div>
                        )}
                        
                        <div className="grid grid-cols-3 gap-2 text-xs">
                          <div>Error Rate: {deployment.metrics.errorRate.toFixed(2)}%</div>
                          <div>Latency: {deployment.metrics.latency}ms</div>
                          <div>Throughput: {formatNumber(deployment.metrics.throughput)}/s</div>
                        </div>
                      </div>
                      
                      <div className="flex items-center space-x-2">
                        {deployment.status === 'in_progress' && (
                          <>
                            <Button size="sm" variant="outline">
                              <Pause className="w-3 h-3" />
                            </Button>
                            <Button size="sm" variant="destructive">
                              <RotateCcw className="w-3 h-3" />
                            </Button>
                          </>
                        )}
                        {deployment.status === 'completed' && (
                          <Button size="sm" variant="outline">
                            <Target className="w-3 h-3" />
                          </Button>
                        )}
                      </div>
                    </div>
                  </div>
                ))}
                
                {deployments.length === 0 && (
                  <div className="text-center text-muted-foreground py-8">
                    No active deployments
                  </div>
                )}
              </div>
            </div>
          )}
        </Card.Content>
      </Card>

      {/* Service Details */}
      {selectedService && (
        <Card>
          <Card.Header>
            <Card.Title>Service Details: {selectedService.name}</Card.Title>
            <Button
              onClick={() => createCanaryDeployment(selectedService.id, `${selectedService.version}.1`)}
            >
              <GitBranch className="w-4 h-4 mr-2" />
              Deploy Canary
            </Button>
          </Card.Header>
          
          <Card.Content className="space-y-4">
            <div>
              <div className="flex items-center space-x-2 mb-2">
                {getServiceIcon(selectedService.type)}
                <h3 className="font-semibold">{selectedService.name}</h3>
                <Badge variant={getStatusColor(selectedService.status)}>
                  {selectedService.status}
                </Badge>
              </div>
              
              <p className="text-sm text-muted-foreground">
                {selectedService.namespace} • {selectedService.type} • v{selectedService.version}
              </p>
            </div>

            <div>
              <h4 className="text-sm font-medium mb-2">Performance Metrics</h4>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                <div>
                  <span className="text-muted-foreground">Request Rate:</span>
                  <span className="ml-2 font-medium">{formatNumber(selectedService.metrics.requestRate)}/s</span>
                </div>
                <div>
                  <span className="text-muted-foreground">Error Rate:</span>
                  <span className="ml-2 font-medium">{selectedService.metrics.errorRate.toFixed(2)}%</span>
                </div>
                <div>
                  <span className="text-muted-foreground">Latency:</span>
                  <span className="ml-2 font-medium">{selectedService.metrics.latency}ms</span>
                </div>
                <div>
                  <span className="text-muted-foreground">Success Rate:</span>
                  <span className="ml-2 font-medium">{selectedService.metrics.successRate.toFixed(1)}%</span>
                </div>
              </div>
            </div>

            <div>
              <h4 className="text-sm font-medium mb-2">Replicas</h4>
              <div className="flex items-center space-x-4 text-sm">
                <span>Desired: {selectedService.replicas.desired}</span>
                <span>Current: {selectedService.replicas.current}</span>
                <span>Ready: {selectedService.replicas.ready}</span>
              </div>
            </div>

            <div>
              <h4 className="text-sm font-medium mb-2">Endpoints ({selectedService.endpoints.length})</h4>
              <div className="space-y-2">
                {selectedService.endpoints.map((endpoint) => (
                  <div key={endpoint.id} className="flex items-center justify-between p-2 bg-gray-50 rounded text-sm">
                    <span>{endpoint.address}:{endpoint.port}</span>
                    <div className="flex items-center space-x-2">
                      <Badge variant={endpoint.healthy ? 'success' : 'destructive'} size="sm">
                        {endpoint.healthy ? 'healthy' : 'unhealthy'}
                      </Badge>
                      <span className="text-muted-foreground">Weight: {endpoint.weight}</span>
                      <span className="text-muted-foreground">{endpoint.zone}</span>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            <div>
              <h4 className="text-sm font-medium mb-2">Dependencies</h4>
              <div className="flex flex-wrap gap-1">
                {selectedService.dependencies.map((dep, index) => (
                  <Badge key={index} variant="outline" size="sm">
                    {dep}
                  </Badge>
                ))}
                {selectedService.dependencies.length === 0 && (
                  <span className="text-sm text-muted-foreground">No dependencies</span>
                )}
              </div>
            </div>

            <div>
              <h4 className="text-sm font-medium mb-2">Policies ({selectedService.policies.length})</h4>
              <div className="flex flex-wrap gap-1">
                {selectedService.policies.map((policy, index) => (
                  <Badge key={index} variant="default" size="sm">
                    {policy}
                  </Badge>
                ))}
                {selectedService.policies.length === 0 && (
                  <span className="text-sm text-muted-foreground">No policies applied</span>
                )}
              </div>
            </div>

            <div className="text-xs text-muted-foreground">
              Last deployed: {new Date(selectedService.lastDeployed).toLocaleString()}
            </div>
          </Card.Content>
        </Card>
      )}
    </div>
  )
}

export default ServiceMeshDashboard