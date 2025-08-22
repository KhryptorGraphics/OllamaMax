import React, { useState, useEffect } from 'react'
import { 
  Shield, 
  Key, 
  Lock,
  Users,
  Monitor,
  Globe,
  AlertTriangle,
  CheckCircle,
  XCircle,
  Clock,
  Eye,
  Settings,
  RefreshCw,
  Plus,
  Edit,
  Trash2,
  Download,
  Upload,
  Search,
  Filter,
  Activity,
  Zap,
  Database,
  Network
} from 'lucide-react'
import { Button, Card, Input, Alert, Badge, Switch, Select } from '@/design-system'
import { useWebSocket } from '@/hooks/useWebSocket'
import { useAPI } from '@/hooks/useAPI'
import { formatBytes, formatNumber, formatDuration } from '@/utils/format'

interface Identity {
  id: string
  type: 'user' | 'service' | 'device' | 'application'
  name: string
  email?: string
  status: 'verified' | 'pending' | 'revoked' | 'expired'
  lastVerified: string
  riskScore: number
  location: string
  deviceInfo?: {
    type: string
    os: string
    browser?: string
    fingerprint: string
  }
  permissions: string[]
  policies: string[]
  certificates: string[]
  mfaEnabled: boolean
  lastActivity: string
  failedAttempts: number
}

interface NetworkPolicy {
  id: string
  name: string
  description: string
  source: {
    type: 'identity' | 'service' | 'subnet' | 'any'
    values: string[]
  }
  destination: {
    type: 'service' | 'subnet' | 'port' | 'any'
    values: string[]
  }
  action: 'allow' | 'deny' | 'audit'
  protocol: 'tcp' | 'udp' | 'icmp' | 'any'
  ports: string[]
  conditions: {
    timeRange?: string
    location?: string[]
    riskScore?: number
  }
  enabled: boolean
  priority: number
  createdAt: string
  lastModified: string
  hitCount: number
}

interface Certificate {
  id: string
  subject: string
  issuer: string
  serialNumber: string
  algorithm: string
  keySize: number
  validFrom: string
  validTo: string
  status: 'valid' | 'expiring' | 'expired' | 'revoked'
  type: 'identity' | 'service' | 'ca' | 'intermediate'
  usage: string[]
  autoRenew: boolean
  renewBefore: number // days
  lastRotated: string
  fingerprint: string
}

interface SecurityEvent {
  id: string
  timestamp: string
  type: 'authentication' | 'authorization' | 'policy_violation' | 'certificate' | 'anomaly'
  severity: 'low' | 'medium' | 'high' | 'critical'
  source: {
    identity: string
    ip: string
    location: string
  }
  target: {
    resource: string
    action: string
  }
  result: 'success' | 'failure' | 'blocked'
  details: string
  riskFactors: string[]
  mitigations: string[]
}

interface SecurityMetrics {
  identitiesTotal: number
  identitiesVerified: number
  certificatesActive: number
  certificatesExpiring: number
  policiesActive: number
  eventsToday: number
  threatsBlocked: number
  avgRiskScore: number
  complianceScore: number
}

export const ZeroTrustDashboard: React.FC = () => {
  const [identities, setIdentities] = useState<Identity[]>([])
  const [policies, setPolicies] = useState<NetworkPolicy[]>([])
  const [certificates, setCertificates] = useState<Certificate[]>([])
  const [events, setEvents] = useState<SecurityEvent[]>([])
  const [metrics, setMetrics] = useState<SecurityMetrics | null>(null)
  const [selectedIdentity, setSelectedIdentity] = useState<Identity | null>(null)
  const [activeTab, setActiveTab] = useState<'identities' | 'policies' | 'certificates' | 'events'>('identities')
  const [filterType, setFilterType] = useState<string>('all')
  const [filterStatus, setFilterStatus] = useState<string>('all')
  const [searchQuery, setSearchQuery] = useState('')
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [showPolicyModal, setShowPolicyModal] = useState(false)

  const { data: securityData, mutate: refreshSecurity } = useAPI('/api/security/zero-trust')
  
  const ws = useWebSocket('ws://localhost:8080/ws/security', {
    onMessage: (data) => {
      switch (data.type) {
        case 'identity_update':
          setIdentities(prev => {
            const index = prev.findIndex(i => i.id === data.identity.id)
            if (index >= 0) {
              const updated = [...prev]
              updated[index] = { ...updated[index], ...data.identity }
              return updated
            }
            return [...prev, data.identity]
          })
          break
        case 'security_event':
          setEvents(prev => [data.event, ...prev.slice(0, 99)])
          break
        case 'certificate_update':
          setCertificates(prev => {
            const index = prev.findIndex(c => c.id === data.certificate.id)
            if (index >= 0) {
              const updated = [...prev]
              updated[index] = { ...updated[index], ...data.certificate }
              return updated
            }
            return [...prev, data.certificate]
          })
          break
        case 'metrics_update':
          setMetrics(data.metrics)
          break
      }
    }
  })

  useEffect(() => {
    if (securityData) {
      setIdentities(securityData.identities || [])
      setPolicies(securityData.policies || [])
      setCertificates(securityData.certificates || [])
      setEvents(securityData.events || [])
      setMetrics(securityData.metrics || null)
    }
  }, [securityData])

  const verifyIdentity = async (identityId: string) => {
    try {
      await fetch(`/api/security/identities/${identityId}/verify`, {
        method: 'POST'
      })
      refreshSecurity()
    } catch (error) {
      console.error('Failed to verify identity:', error)
    }
  }

  const revokeIdentity = async (identityId: string) => {
    try {
      await fetch(`/api/security/identities/${identityId}/revoke`, {
        method: 'POST'
      })
      refreshSecurity()
    } catch (error) {
      console.error('Failed to revoke identity:', error)
    }
  }

  const createPolicy = async (policy: Omit<NetworkPolicy, 'id' | 'createdAt' | 'lastModified' | 'hitCount'>) => {
    try {
      const response = await fetch('/api/security/policies', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(policy)
      })
      
      if (!response.ok) throw new Error('Failed to create policy')
      
      refreshSecurity()
      setShowPolicyModal(false)
    } catch (error) {
      console.error('Failed to create policy:', error)
    }
  }

  const renewCertificate = async (certificateId: string) => {
    try {
      await fetch(`/api/security/certificates/${certificateId}/renew`, {
        method: 'POST'
      })
      refreshSecurity()
    } catch (error) {
      console.error('Failed to renew certificate:', error)
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'verified': case 'valid': case 'success': return 'success'
      case 'pending': case 'expiring': return 'warning'
      case 'revoked': case 'expired': case 'failure': case 'blocked': return 'destructive'
      default: return 'secondary'
    }
  }

  const getRiskColor = (score: number) => {
    if (score >= 80) return 'text-red-500'
    if (score >= 60) return 'text-orange-500'
    if (score >= 40) return 'text-yellow-500'
    return 'text-green-500'
  }

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical': return 'destructive'
      case 'high': return 'warning'
      case 'medium': return 'default'
      case 'low': return 'secondary'
      default: return 'secondary'
    }
  }

  const filteredIdentities = identities.filter(identity => {
    const typeMatch = filterType === 'all' || identity.type === filterType
    const statusMatch = filterStatus === 'all' || identity.status === filterStatus
    const searchMatch = searchQuery === '' || 
      identity.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      identity.email?.toLowerCase().includes(searchQuery.toLowerCase())
    return typeMatch && statusMatch && searchMatch
  })

  return (
    <div className="space-y-6">
      {/* Security Overview */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <Card.Content className="p-4">
            <div className="flex items-center space-x-3">
              <Shield className="w-8 h-8 text-blue-500" />
              <div>
                <p className="text-sm text-muted-foreground">Verified Identities</p>
                <p className="text-2xl font-bold">{metrics?.identitiesVerified || 0}</p>
              </div>
            </div>
          </Card.Content>
        </Card>

        <Card>
          <Card.Content className="p-4">
            <div className="flex items-center space-x-3">
              <Key className="w-8 h-8 text-green-500" />
              <div>
                <p className="text-sm text-muted-foreground">Active Certificates</p>
                <p className="text-2xl font-bold">{metrics?.certificatesActive || 0}</p>
              </div>
            </div>
          </Card.Content>
        </Card>

        <Card>
          <Card.Content className="p-4">
            <div className="flex items-center space-x-3">
              <AlertTriangle className="w-8 h-8 text-red-500" />
              <div>
                <p className="text-sm text-muted-foreground">Threats Blocked</p>
                <p className="text-2xl font-bold">{metrics?.threatsBlocked || 0}</p>
              </div>
            </div>
          </Card.Content>
        </Card>

        <Card>
          <Card.Content className="p-4">
            <div className="flex items-center space-x-3">
              <Activity className="w-8 h-8 text-purple-500" />
              <div>
                <p className="text-sm text-muted-foreground">Compliance Score</p>
                <p className="text-2xl font-bold">{metrics?.complianceScore || 0}%</p>
              </div>
            </div>
          </Card.Content>
        </Card>
      </div>

      {/* Main Dashboard */}
      <Card>
        <Card.Header>
          <Card.Title>Zero Trust Security Dashboard</Card.Title>
          <div className="flex space-x-2">
            <Button onClick={() => setShowCreateModal(true)}>
              <Plus className="w-4 h-4 mr-2" />
              Add Identity
            </Button>
            <Button variant="outline" onClick={() => refreshSecurity()}>
              <RefreshCw className="w-4 h-4 mr-2" />
              Refresh
            </Button>
          </div>
        </Card.Header>

        <Card.Content>
          {/* Tab Navigation */}
          <div className="border-b mb-6">
            <nav className="flex space-x-4">
              {['identities', 'policies', 'certificates', 'events'].map((tab) => (
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

          {/* Identities Tab */}
          {activeTab === 'identities' && (
            <div className="space-y-4">
              {/* Filters */}
              <div className="flex space-x-4">
                <div className="flex-1">
                  <Input
                    placeholder="Search identities..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    leftIcon={<Search className="w-4 h-4" />}
                  />
                </div>
                <select
                  className="px-3 py-2 border rounded"
                  value={filterType}
                  onChange={(e) => setFilterType(e.target.value)}
                >
                  <option value="all">All Types</option>
                  <option value="user">Users</option>
                  <option value="service">Services</option>
                  <option value="device">Devices</option>
                  <option value="application">Applications</option>
                </select>
                <select
                  className="px-3 py-2 border rounded"
                  value={filterStatus}
                  onChange={(e) => setFilterStatus(e.target.value)}
                >
                  <option value="all">All Status</option>
                  <option value="verified">Verified</option>
                  <option value="pending">Pending</option>
                  <option value="revoked">Revoked</option>
                  <option value="expired">Expired</option>
                </select>
              </div>

              {/* Identities List */}
              <div className="space-y-3">
                {filteredIdentities.map((identity) => (
                  <div
                    key={identity.id}
                    className={`p-4 border rounded-lg cursor-pointer transition-colors ${
                      selectedIdentity?.id === identity.id ? 'border-primary bg-primary/5' : 'hover:bg-gray-50'
                    }`}
                    onClick={() => setSelectedIdentity(identity)}
                  >
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center space-x-2 mb-2">
                          <Users className="w-4 h-4" />
                          <h4 className="text-sm font-medium">{identity.name}</h4>
                          <Badge variant={getStatusColor(identity.status)}>
                            {identity.status}
                          </Badge>
                          <Badge variant="outline">{identity.type}</Badge>
                          {identity.mfaEnabled && (
                            <Badge variant="success" size="sm">MFA</Badge>
                          )}
                        </div>
                        
                        {identity.email && (
                          <p className="text-xs text-muted-foreground mb-2">{identity.email}</p>
                        )}
                        
                        <div className="flex items-center space-x-4 text-xs">
                          <span>Location: {identity.location}</span>
                          <span>Last activity: {new Date(identity.lastActivity).toLocaleString()}</span>
                          <span className={getRiskColor(identity.riskScore)}>
                            Risk: {identity.riskScore}%
                          </span>
                        </div>
                      </div>
                      
                      <div className="flex items-center space-x-2">
                        {identity.failedAttempts > 0 && (
                          <Badge variant="warning" size="sm">
                            {identity.failedAttempts} failed
                          </Badge>
                        )}
                        <div className="flex space-x-1">
                          {identity.status === 'pending' && (
                            <Button
                              size="sm"
                              onClick={(e) => {
                                e.stopPropagation()
                                verifyIdentity(identity.id)
                              }}
                            >
                              <CheckCircle className="w-3 h-3" />
                            </Button>
                          )}
                          {identity.status === 'verified' && (
                            <Button
                              size="sm"
                              variant="destructive"
                              onClick={(e) => {
                                e.stopPropagation()
                                revokeIdentity(identity.id)
                              }}
                            >
                              <XCircle className="w-3 h-3" />
                            </Button>
                          )}
                        </div>
                      </div>
                    </div>
                  </div>
                ))}
                
                {filteredIdentities.length === 0 && (
                  <div className="text-center text-muted-foreground py-8">
                    No identities match the current filters
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Policies Tab */}
          {activeTab === 'policies' && (
            <div className="space-y-4">
              <div className="flex justify-between items-center">
                <h4 className="text-sm font-medium">Network Policies ({policies.length})</h4>
                <Button onClick={() => setShowPolicyModal(true)}>
                  <Plus className="w-4 h-4 mr-2" />
                  Create Policy
                </Button>
              </div>
              
              <div className="space-y-3">
                {policies.map((policy) => (
                  <div key={policy.id} className="p-4 border rounded-lg">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center space-x-2 mb-2">
                          <Network className="w-4 h-4" />
                          <h4 className="text-sm font-medium">{policy.name}</h4>
                          <Badge variant={policy.action === 'allow' ? 'success' : policy.action === 'deny' ? 'destructive' : 'default'}>
                            {policy.action}
                          </Badge>
                          <Badge variant={policy.enabled ? 'success' : 'secondary'}>
                            {policy.enabled ? 'enabled' : 'disabled'}
                          </Badge>
                        </div>
                        
                        <p className="text-xs text-muted-foreground mb-2">{policy.description}</p>
                        
                        <div className="grid grid-cols-2 md:grid-cols-4 gap-2 text-xs">
                          <div>
                            <span className="text-muted-foreground">Source:</span>
                            <span className="ml-1">{policy.source.type}</span>
                          </div>
                          <div>
                            <span className="text-muted-foreground">Destination:</span>
                            <span className="ml-1">{policy.destination.type}</span>
                          </div>
                          <div>
                            <span className="text-muted-foreground">Protocol:</span>
                            <span className="ml-1">{policy.protocol.toUpperCase()}</span>
                          </div>
                          <div>
                            <span className="text-muted-foreground">Hits:</span>
                            <span className="ml-1">{formatNumber(policy.hitCount)}</span>
                          </div>
                        </div>
                      </div>
                      
                      <div className="flex items-center space-x-2">
                        <span className="text-xs text-muted-foreground">
                          Priority: {policy.priority}
                        </span>
                        <Switch
                          checked={policy.enabled}
                          onCheckedChange={(checked) => {
                            // Update policy enabled state
                          }}
                        />
                      </div>
                    </div>
                  </div>
                ))}
                
                {policies.length === 0 && (
                  <div className="text-center text-muted-foreground py-8">
                    No network policies configured
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Certificates Tab */}
          {activeTab === 'certificates' && (
            <div className="space-y-4">
              <div className="flex justify-between items-center">
                <h4 className="text-sm font-medium">Certificates ({certificates.length})</h4>
                <div className="flex space-x-2">
                  <Button variant="outline" size="sm">
                    <Upload className="w-4 h-4 mr-2" />
                    Import
                  </Button>
                  <Button size="sm">
                    <Plus className="w-4 h-4 mr-2" />
                    Generate
                  </Button>
                </div>
              </div>
              
              <div className="space-y-3">
                {certificates.map((cert) => (
                  <div key={cert.id} className="p-4 border rounded-lg">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center space-x-2 mb-2">
                          <Key className="w-4 h-4" />
                          <h4 className="text-sm font-medium">{cert.subject}</h4>
                          <Badge variant={getStatusColor(cert.status)}>
                            {cert.status}
                          </Badge>
                          <Badge variant="outline">{cert.type}</Badge>
                          {cert.autoRenew && (
                            <Badge variant="default" size="sm">Auto-renew</Badge>
                          )}
                        </div>
                        
                        <div className="grid grid-cols-2 md:grid-cols-4 gap-2 text-xs mb-2">
                          <div>
                            <span className="text-muted-foreground">Algorithm:</span>
                            <span className="ml-1">{cert.algorithm}</span>
                          </div>
                          <div>
                            <span className="text-muted-foreground">Key Size:</span>
                            <span className="ml-1">{cert.keySize} bits</span>
                          </div>
                          <div>
                            <span className="text-muted-foreground">Valid From:</span>
                            <span className="ml-1">{new Date(cert.validFrom).toLocaleDateString()}</span>
                          </div>
                          <div>
                            <span className="text-muted-foreground">Valid To:</span>
                            <span className="ml-1">{new Date(cert.validTo).toLocaleDateString()}</span>
                          </div>
                        </div>
                        
                        <div className="text-xs text-muted-foreground">
                          <div>Issuer: {cert.issuer}</div>
                          <div>Serial: {cert.serialNumber}</div>
                          <div>Fingerprint: {cert.fingerprint.substring(0, 32)}...</div>
                        </div>
                      </div>
                      
                      <div className="flex items-center space-x-2">
                        {cert.status === 'expiring' && (
                          <Button
                            size="sm"
                            onClick={() => renewCertificate(cert.id)}
                          >
                            <RefreshCw className="w-3 h-3 mr-1" />
                            Renew
                          </Button>
                        )}
                        <Button size="sm" variant="outline">
                          <Download className="w-3 h-3" />
                        </Button>
                      </div>
                    </div>
                    
                    {/* Certificate expiry progress */}
                    {cert.status === 'expiring' && (
                      <div className="mt-3">
                        <div className="flex justify-between text-xs mb-1">
                          <span>Expires in {Math.floor((new Date(cert.validTo).getTime() - Date.now()) / (1000 * 60 * 60 * 24))} days</span>
                          <span>Auto-renew: {cert.renewBefore} days before</span>
                        </div>
                        <div className="w-full bg-gray-200 rounded-full h-2">
                          <div className="bg-orange-500 h-2 rounded-full" style={{ width: '25%' }} />
                        </div>
                      </div>
                    )}
                  </div>
                ))}
                
                {certificates.length === 0 && (
                  <div className="text-center text-muted-foreground py-8">
                    No certificates found
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Events Tab */}
          {activeTab === 'events' && (
            <div className="space-y-4">
              <div className="flex justify-between items-center">
                <h4 className="text-sm font-medium">Security Events ({events.length})</h4>
                <div className="flex space-x-2">
                  <select className="px-3 py-1 border rounded text-sm">
                    <option value="all">All Types</option>
                    <option value="authentication">Authentication</option>
                    <option value="authorization">Authorization</option>
                    <option value="policy_violation">Policy Violation</option>
                    <option value="certificate">Certificate</option>
                    <option value="anomaly">Anomaly</option>
                  </select>
                  <select className="px-3 py-1 border rounded text-sm">
                    <option value="all">All Severity</option>
                    <option value="critical">Critical</option>
                    <option value="high">High</option>
                    <option value="medium">Medium</option>
                    <option value="low">Low</option>
                  </select>
                </div>
              </div>
              
              <div className="space-y-3 max-h-96 overflow-y-auto">
                {events.map((event) => (
                  <div key={event.id} className="p-4 border rounded-lg">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center space-x-2 mb-2">
                          <Activity className="w-4 h-4" />
                          <span className="text-sm font-medium capitalize">{event.type.replace('_', ' ')}</span>
                          <Badge variant={getSeverityColor(event.severity)}>
                            {event.severity}
                          </Badge>
                          <Badge variant={getStatusColor(event.result)}>
                            {event.result}
                          </Badge>
                        </div>
                        
                        <p className="text-sm mb-2">{event.details}</p>
                        
                        <div className="grid grid-cols-2 gap-2 text-xs">
                          <div>
                            <span className="text-muted-foreground">Source:</span>
                            <span className="ml-1">{event.source.identity} ({event.source.ip})</span>
                          </div>
                          <div>
                            <span className="text-muted-foreground">Target:</span>
                            <span className="ml-1">{event.target.resource}</span>
                          </div>
                          <div>
                            <span className="text-muted-foreground">Location:</span>
                            <span className="ml-1">{event.source.location}</span>
                          </div>
                          <div>
                            <span className="text-muted-foreground">Action:</span>
                            <span className="ml-1">{event.target.action}</span>
                          </div>
                        </div>
                        
                        {event.riskFactors.length > 0 && (
                          <div className="mt-2">
                            <span className="text-xs text-muted-foreground">Risk Factors: </span>
                            {event.riskFactors.map((factor, index) => (
                              <Badge key={index} variant="warning" size="sm" className="mr-1">
                                {factor}
                              </Badge>
                            ))}
                          </div>
                        )}
                      </div>
                      
                      <div className="text-xs text-muted-foreground text-right">
                        <div>{new Date(event.timestamp).toLocaleString()}</div>
                      </div>
                    </div>
                  </div>
                ))}
                
                {events.length === 0 && (
                  <div className="text-center text-muted-foreground py-8">
                    No security events found
                  </div>
                )}
              </div>
            </div>
          )}
        </Card.Content>
      </Card>

      {/* Identity Details Sidebar */}
      {selectedIdentity && (
        <Card>
          <Card.Header>
            <Card.Title>Identity Details: {selectedIdentity.name}</Card.Title>
          </Card.Header>
          
          <Card.Content className="space-y-4">
            <div>
              <div className="flex items-center space-x-2 mb-2">
                <Users className="w-4 h-4" />
                <h3 className="font-semibold">{selectedIdentity.name}</h3>
                <Badge variant={getStatusColor(selectedIdentity.status)}>
                  {selectedIdentity.status}
                </Badge>
              </div>
              
              {selectedIdentity.email && (
                <p className="text-sm text-muted-foreground">{selectedIdentity.email}</p>
              )}
            </div>

            <div>
              <h4 className="text-sm font-medium mb-2">Risk Assessment</h4>
              <div className="flex items-center space-x-2">
                <span className="text-sm">Risk Score:</span>
                <span className={`font-semibold ${getRiskColor(selectedIdentity.riskScore)}`}>
                  {selectedIdentity.riskScore}%
                </span>
              </div>
              <div className="w-full bg-gray-200 rounded-full h-2 mt-1">
                <div
                  className={`h-2 rounded-full ${
                    selectedIdentity.riskScore >= 80 ? 'bg-red-500' :
                    selectedIdentity.riskScore >= 60 ? 'bg-orange-500' :
                    selectedIdentity.riskScore >= 40 ? 'bg-yellow-500' : 'bg-green-500'
                  }`}
                  style={{ width: `${selectedIdentity.riskScore}%` }}
                />
              </div>
            </div>

            {selectedIdentity.deviceInfo && (
              <div>
                <h4 className="text-sm font-medium mb-2">Device Information</h4>
                <div className="text-sm space-y-1">
                  <div>Type: {selectedIdentity.deviceInfo.type}</div>
                  <div>OS: {selectedIdentity.deviceInfo.os}</div>
                  {selectedIdentity.deviceInfo.browser && (
                    <div>Browser: {selectedIdentity.deviceInfo.browser}</div>
                  )}
                  <div className="text-xs text-muted-foreground">
                    Fingerprint: {selectedIdentity.deviceInfo.fingerprint.substring(0, 16)}...
                  </div>
                </div>
              </div>
            )}

            <div>
              <h4 className="text-sm font-medium mb-2">Permissions ({selectedIdentity.permissions.length})</h4>
              <div className="flex flex-wrap gap-1">
                {selectedIdentity.permissions.map((permission, index) => (
                  <Badge key={index} variant="outline" size="sm">
                    {permission}
                  </Badge>
                ))}
              </div>
            </div>

            <div>
              <h4 className="text-sm font-medium mb-2">Policies ({selectedIdentity.policies.length})</h4>
              <div className="flex flex-wrap gap-1">
                {selectedIdentity.policies.map((policy, index) => (
                  <Badge key={index} variant="default" size="sm">
                    {policy}
                  </Badge>
                ))}
              </div>
            </div>

            <div className="text-xs text-muted-foreground space-y-1">
              <div>Last verified: {new Date(selectedIdentity.lastVerified).toLocaleString()}</div>
              <div>Last activity: {new Date(selectedIdentity.lastActivity).toLocaleString()}</div>
              <div>Location: {selectedIdentity.location}</div>
              <div>MFA enabled: {selectedIdentity.mfaEnabled ? 'Yes' : 'No'}</div>
              <div>Failed attempts: {selectedIdentity.failedAttempts}</div>
            </div>
          </Card.Content>
        </Card>
      )}
    </div>
  )
}

export default ZeroTrustDashboard