import React, { useState, useEffect } from 'react'
import { 
  Users, 
  Building, 
  Shield, 
  Settings, 
  Key,
  Globe,
  Database,
  Cpu,
  HardDrive,
  Network,
  Lock,
  UserPlus,
  Edit,
  Trash2,
  Eye,
  BarChart3,
  Activity,
  Clock,
  AlertTriangle,
  CheckCircle
} from 'lucide-react'
import { Button, Card, Input, Alert, Badge, Switch, Select } from '@/design-system'
import { useAPI } from '@/hooks/useAPI'
import { formatBytes, formatNumber, formatCurrency } from '@/utils/format'

interface Tenant {
  id: string
  name: string
  displayName: string
  description: string
  status: 'active' | 'suspended' | 'inactive'
  tier: 'free' | 'basic' | 'premium' | 'enterprise'
  owner: {
    id: string
    name: string
    email: string
  }
  limits: {
    users: number
    models: number
    requests: number
    storage: number // GB
    compute: number // CPU hours
    bandwidth: number // GB
  }
  usage: {
    users: number
    models: number
    requests: number
    storage: number
    compute: number
    bandwidth: number
  }
  billing: {
    plan: string
    cost: number
    currency: string
    billingCycle: 'monthly' | 'yearly'
    lastPayment: string
    nextBilling: string
  }
  security: {
    ssoEnabled: boolean
    mfaRequired: boolean
    ipWhitelist: string[]
    dataRetention: number // days
    encryptionLevel: 'standard' | 'advanced'
  }
  isolation: {
    type: 'shared' | 'dedicated' | 'hybrid'
    namespace: string
    resources: {
      cpu: number
      memory: number
      storage: number
    }
    network: {
      subnet: string
      isolation: boolean
    }
  }
  createdAt: string
  updatedAt: string
  lastActivity: string
}

interface User {
  id: string
  tenantId: string
  name: string
  email: string
  role: 'admin' | 'user' | 'viewer'
  status: 'active' | 'invited' | 'suspended'
  permissions: string[]
  lastLogin: string
  createdAt: string
}

interface TenantMetrics {
  totalTenants: number
  activeTenants: number
  totalUsers: number
  totalRevenue: number
  resourceUtilization: {
    cpu: number
    memory: number
    storage: number
  }
  topTenants: {
    id: string
    name: string
    usage: number
    cost: number
  }[]
}

export const TenantManagement: React.FC = () => {
  const [tenants, setTenants] = useState<Tenant[]>([])
  const [selectedTenant, setSelectedTenant] = useState<Tenant | null>(null)
  const [users, setUsers] = useState<User[]>([])
  const [metrics, setMetrics] = useState<TenantMetrics | null>(null)
  const [activeTab, setActiveTab] = useState<'overview' | 'users' | 'security' | 'billing' | 'resources'>('overview')
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [showUserModal, setShowUserModal] = useState(false)
  const [newTenant, setNewTenant] = useState({
    name: '',
    displayName: '',
    description: '',
    tier: 'basic' as const,
    ownerEmail: ''
  })
  const [newUser, setNewUser] = useState({
    name: '',
    email: '',
    role: 'user' as const
  })

  const { data: tenantsData, mutate: refreshTenants } = useAPI('/api/admin/tenants')
  const { data: metricsData } = useAPI('/api/admin/metrics')
  const { data: tenantUsers, mutate: refreshUsers } = useAPI(
    selectedTenant ? `/api/admin/tenants/${selectedTenant.id}/users` : null
  )

  useEffect(() => {
    if (tenantsData) {
      setTenants(tenantsData.tenants || [])
    }
    if (metricsData) {
      setMetrics(metricsData)
    }
  }, [tenantsData, metricsData])

  useEffect(() => {
    if (tenantUsers) {
      setUsers(tenantUsers.users || [])
    }
  }, [tenantUsers])

  const createTenant = async () => {
    try {
      const response = await fetch('/api/admin/tenants', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(newTenant)
      })
      
      if (!response.ok) throw new Error('Failed to create tenant')
      
      await refreshTenants()
      setShowCreateModal(false)
      setNewTenant({
        name: '',
        displayName: '',
        description: '',
        tier: 'basic',
        ownerEmail: ''
      })
    } catch (error) {
      console.error('Failed to create tenant:', error)
    }
  }

  const updateTenantStatus = async (tenantId: string, status: Tenant['status']) => {
    try {
      await fetch(`/api/admin/tenants/${tenantId}/status`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ status })
      })
      await refreshTenants()
    } catch (error) {
      console.error('Failed to update tenant status:', error)
    }
  }

  const inviteUser = async () => {
    if (!selectedTenant) return
    
    try {
      const response = await fetch(`/api/admin/tenants/${selectedTenant.id}/users/invite`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(newUser)
      })
      
      if (!response.ok) throw new Error('Failed to invite user')
      
      await refreshUsers()
      setShowUserModal(false)
      setNewUser({ name: '', email: '', role: 'user' })
    } catch (error) {
      console.error('Failed to invite user:', error)
    }
  }

  const updateUserRole = async (userId: string, role: User['role']) => {
    if (!selectedTenant) return
    
    try {
      await fetch(`/api/admin/tenants/${selectedTenant.id}/users/${userId}/role`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ role })
      })
      await refreshUsers()
    } catch (error) {
      console.error('Failed to update user role:', error)
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'success'
      case 'suspended': return 'warning'
      case 'inactive': return 'destructive'
      case 'invited': return 'default'
      default: return 'secondary'
    }
  }

  const getTierColor = (tier: string) => {
    switch (tier) {
      case 'enterprise': return 'default'
      case 'premium': return 'warning'
      case 'basic': return 'secondary'
      case 'free': return 'outline'
      default: return 'secondary'
    }
  }

  const getUsagePercentage = (used: number, limit: number) => {
    return limit > 0 ? Math.min((used / limit) * 100, 100) : 0
  }

  const getUsageColor = (percentage: number) => {
    if (percentage >= 90) return 'bg-red-500'
    if (percentage >= 75) return 'bg-yellow-500'
    if (percentage >= 50) return 'bg-blue-500'
    return 'bg-green-500'
  }

  return (
    <div className="space-y-6">
      {/* Metrics Overview */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <Card.Content className="p-4">
            <div className="flex items-center space-x-3">
              <Building className="w-8 h-8 text-blue-500" />
              <div>
                <p className="text-sm text-muted-foreground">Total Tenants</p>
                <p className="text-2xl font-bold">{metrics?.totalTenants || 0}</p>
              </div>
            </div>
          </Card.Content>
        </Card>

        <Card>
          <Card.Content className="p-4">
            <div className="flex items-center space-x-3">
              <Activity className="w-8 h-8 text-green-500" />
              <div>
                <p className="text-sm text-muted-foreground">Active Tenants</p>
                <p className="text-2xl font-bold">{metrics?.activeTenants || 0}</p>
              </div>
            </div>
          </Card.Content>
        </Card>

        <Card>
          <Card.Content className="p-4">
            <div className="flex items-center space-x-3">
              <Users className="w-8 h-8 text-purple-500" />
              <div>
                <p className="text-sm text-muted-foreground">Total Users</p>
                <p className="text-2xl font-bold">{formatNumber(metrics?.totalUsers || 0)}</p>
              </div>
            </div>
          </Card.Content>
        </Card>

        <Card>
          <Card.Content className="p-4">
            <div className="flex items-center space-x-3">
              <BarChart3 className="w-8 h-8 text-orange-500" />
              <div>
                <p className="text-sm text-muted-foreground">Monthly Revenue</p>
                <p className="text-2xl font-bold">{formatCurrency(metrics?.totalRevenue || 0)}</p>
              </div>
            </div>
          </Card.Content>
        </Card>
      </div>

      {/* Tenant Management */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Tenants List */}
        <Card>
          <Card.Header>
            <Card.Title>Tenants</Card.Title>
            <Button onClick={() => setShowCreateModal(true)}>
              <Building className="w-4 h-4 mr-2" />
              Create Tenant
            </Button>
          </Card.Header>
          
          <Card.Content>
            <div className="space-y-3 max-h-96 overflow-y-auto">
              {tenants.map((tenant) => (
                <div
                  key={tenant.id}
                  className={`p-3 border rounded-lg cursor-pointer transition-colors ${
                    selectedTenant?.id === tenant.id ? 'border-primary bg-primary/5' : 'hover:bg-gray-50'
                  }`}
                  onClick={() => setSelectedTenant(tenant)}
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <h4 className="text-sm font-medium">{tenant.displayName}</h4>
                      <p className="text-xs text-muted-foreground">{tenant.name}</p>
                      <div className="flex items-center space-x-2 mt-1">
                        <Badge variant={getStatusColor(tenant.status)}>
                          {tenant.status}
                        </Badge>
                        <Badge variant={getTierColor(tenant.tier)}>
                          {tenant.tier}
                        </Badge>
                      </div>
                    </div>
                    
                    <div className="text-xs text-muted-foreground text-right">
                      <div>{tenant.usage.users}/{tenant.limits.users} users</div>
                      <div>{tenant.usage.models} models</div>
                      <div>{formatNumber(tenant.usage.requests)} requests</div>
                    </div>
                  </div>
                  
                  <div className="mt-2 text-xs text-muted-foreground">
                    Owner: {tenant.owner.name} • Last activity: {new Date(tenant.lastActivity).toLocaleDateString()}
                  </div>
                </div>
              ))}
              
              {tenants.length === 0 && (
                <div className="text-center text-muted-foreground py-8">
                  No tenants found
                </div>
              )}
            </div>
          </Card.Content>
        </Card>

        {/* Tenant Details */}
        <Card>
          <Card.Header>
            <Card.Title>Tenant Details</Card.Title>
            {selectedTenant && (
              <div className="flex space-x-2">
                <select
                  className="px-3 py-1 border rounded text-sm"
                  value={selectedTenant.status}
                  onChange={(e) => updateTenantStatus(selectedTenant.id, e.target.value as Tenant['status'])}
                >
                  <option value="active">Active</option>
                  <option value="suspended">Suspended</option>
                  <option value="inactive">Inactive</option>
                </select>
              </div>
            )}
          </Card.Header>
          
          <Card.Content>
            {selectedTenant ? (
              <div className="space-y-4">
                {/* Tenant Info */}
                <div>
                  <div className="flex items-center space-x-2 mb-2">
                    <Building className="w-5 h-5" />
                    <h3 className="font-semibold">{selectedTenant.displayName}</h3>
                    <Badge variant={getStatusColor(selectedTenant.status)}>
                      {selectedTenant.status}
                    </Badge>
                    <Badge variant={getTierColor(selectedTenant.tier)}>
                      {selectedTenant.tier}
                    </Badge>
                  </div>
                  <p className="text-sm text-muted-foreground mb-2">{selectedTenant.description}</p>
                  <p className="text-sm">
                    <span className="text-muted-foreground">Owner:</span> {selectedTenant.owner.name} ({selectedTenant.owner.email})
                  </p>
                </div>

                {/* Tab Navigation */}
                <div className="border-b">
                  <nav className="flex space-x-4">
                    {['overview', 'users', 'security', 'billing', 'resources'].map((tab) => (
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

                {/* Tab Content */}
                {activeTab === 'overview' && (
                  <div className="space-y-4">
                    {/* Usage Overview */}
                    <div>
                      <h4 className="text-sm font-medium mb-2">Resource Usage</h4>
                      <div className="space-y-2">
                        <div>
                          <div className="flex justify-between text-sm mb-1">
                            <span>Users</span>
                            <span>{selectedTenant.usage.users} / {selectedTenant.limits.users}</span>
                          </div>
                          <div className="w-full bg-gray-200 rounded-full h-2">
                            <div
                              className={`h-2 rounded-full ${getUsageColor(getUsagePercentage(selectedTenant.usage.users, selectedTenant.limits.users))}`}
                              style={{ width: `${getUsagePercentage(selectedTenant.usage.users, selectedTenant.limits.users)}%` }}
                            />
                          </div>
                        </div>
                        
                        <div>
                          <div className="flex justify-between text-sm mb-1">
                            <span>Storage</span>
                            <span>{formatBytes(selectedTenant.usage.storage * 1024 * 1024 * 1024)} / {formatBytes(selectedTenant.limits.storage * 1024 * 1024 * 1024)}</span>
                          </div>
                          <div className="w-full bg-gray-200 rounded-full h-2">
                            <div
                              className={`h-2 rounded-full ${getUsageColor(getUsagePercentage(selectedTenant.usage.storage, selectedTenant.limits.storage))}`}
                              style={{ width: `${getUsagePercentage(selectedTenant.usage.storage, selectedTenant.limits.storage)}%` }}
                            />
                          </div>
                        </div>
                        
                        <div>
                          <div className="flex justify-between text-sm mb-1">
                            <span>Requests</span>
                            <span>{formatNumber(selectedTenant.usage.requests)} / {formatNumber(selectedTenant.limits.requests)}</span>
                          </div>
                          <div className="w-full bg-gray-200 rounded-full h-2">
                            <div
                              className={`h-2 rounded-full ${getUsageColor(getUsagePercentage(selectedTenant.usage.requests, selectedTenant.limits.requests))}`}
                              style={{ width: `${getUsagePercentage(selectedTenant.usage.requests, selectedTenant.limits.requests)}%` }}
                            />
                          </div>
                        </div>
                      </div>
                    </div>

                    {/* Key Metrics */}
                    <div className="grid grid-cols-2 gap-4 text-sm">
                      <div>
                        <span className="text-muted-foreground">Models:</span>
                        <span className="ml-2">{selectedTenant.usage.models}</span>
                      </div>
                      <div>
                        <span className="text-muted-foreground">Compute Hours:</span>
                        <span className="ml-2">{formatNumber(selectedTenant.usage.compute)}</span>
                      </div>
                      <div>
                        <span className="text-muted-foreground">Bandwidth:</span>
                        <span className="ml-2">{formatBytes(selectedTenant.usage.bandwidth * 1024 * 1024 * 1024)}</span>
                      </div>
                      <div>
                        <span className="text-muted-foreground">Created:</span>
                        <span className="ml-2">{new Date(selectedTenant.createdAt).toLocaleDateString()}</span>
                      </div>
                    </div>
                  </div>
                )}

                {activeTab === 'users' && (
                  <div className="space-y-4">
                    <div className="flex justify-between items-center">
                      <h4 className="text-sm font-medium">Users ({users.length})</h4>
                      <Button size="sm" onClick={() => setShowUserModal(true)}>
                        <UserPlus className="w-4 h-4 mr-2" />
                        Invite User
                      </Button>
                    </div>
                    
                    <div className="space-y-2 max-h-64 overflow-y-auto">
                      {users.map((user) => (
                        <div key={user.id} className="flex items-center justify-between p-2 border rounded">
                          <div>
                            <div className="text-sm font-medium">{user.name}</div>
                            <div className="text-xs text-muted-foreground">{user.email}</div>
                          </div>
                          <div className="flex items-center space-x-2">
                            <Badge variant={getStatusColor(user.status)}>
                              {user.status}
                            </Badge>
                            <select
                              className="text-xs border rounded px-2 py-1"
                              value={user.role}
                              onChange={(e) => updateUserRole(user.id, e.target.value as User['role'])}
                            >
                              <option value="viewer">Viewer</option>
                              <option value="user">User</option>
                              <option value="admin">Admin</option>
                            </select>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                )}

                {activeTab === 'security' && (
                  <div className="space-y-4">
                    <h4 className="text-sm font-medium">Security Settings</h4>
                    
                    <div className="space-y-3">
                      <div className="flex justify-between items-center">
                        <span className="text-sm">SSO Enabled</span>
                        <Switch checked={selectedTenant.security.ssoEnabled} />
                      </div>
                      
                      <div className="flex justify-between items-center">
                        <span className="text-sm">MFA Required</span>
                        <Switch checked={selectedTenant.security.mfaRequired} />
                      </div>
                      
                      <div>
                        <span className="text-sm">Encryption Level:</span>
                        <span className="ml-2 capitalize">{selectedTenant.security.encryptionLevel}</span>
                      </div>
                      
                      <div>
                        <span className="text-sm">Data Retention:</span>
                        <span className="ml-2">{selectedTenant.security.dataRetention} days</span>
                      </div>
                      
                      <div>
                        <div className="text-sm mb-1">IP Whitelist:</div>
                        <div className="text-xs text-muted-foreground">
                          {selectedTenant.security.ipWhitelist.length > 0 
                            ? selectedTenant.security.ipWhitelist.join(', ')
                            : 'No restrictions'
                          }
                        </div>
                      </div>
                    </div>
                  </div>
                )}

                {activeTab === 'billing' && (
                  <div className="space-y-4">
                    <h4 className="text-sm font-medium">Billing Information</h4>
                    
                    <div className="space-y-2 text-sm">
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">Plan:</span>
                        <span>{selectedTenant.billing.plan}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">Cost:</span>
                        <span>{formatCurrency(selectedTenant.billing.cost)} / {selectedTenant.billing.billingCycle}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">Last Payment:</span>
                        <span>{new Date(selectedTenant.billing.lastPayment).toLocaleDateString()}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">Next Billing:</span>
                        <span>{new Date(selectedTenant.billing.nextBilling).toLocaleDateString()}</span>
                      </div>
                    </div>
                  </div>
                )}

                {activeTab === 'resources' && (
                  <div className="space-y-4">
                    <h4 className="text-sm font-medium">Resource Isolation</h4>
                    
                    <div className="space-y-3">
                      <div>
                        <span className="text-sm">Isolation Type:</span>
                        <span className="ml-2 capitalize">{selectedTenant.isolation.type}</span>
                      </div>
                      
                      <div>
                        <span className="text-sm">Namespace:</span>
                        <span className="ml-2 font-mono text-xs">{selectedTenant.isolation.namespace}</span>
                      </div>
                      
                      <div>
                        <div className="text-sm mb-2">Allocated Resources:</div>
                        <div className="grid grid-cols-3 gap-2 text-xs">
                          <div>
                            <div className="text-muted-foreground">CPU</div>
                            <div>{selectedTenant.isolation.resources.cpu} cores</div>
                          </div>
                          <div>
                            <div className="text-muted-foreground">Memory</div>
                            <div>{selectedTenant.isolation.resources.memory} GB</div>
                          </div>
                          <div>
                            <div className="text-muted-foreground">Storage</div>
                            <div>{selectedTenant.isolation.resources.storage} GB</div>
                          </div>
                        </div>
                      </div>
                      
                      <div>
                        <div className="text-sm mb-1">Network:</div>
                        <div className="text-xs">
                          <div>Subnet: {selectedTenant.isolation.network.subnet}</div>
                          <div>Isolated: {selectedTenant.isolation.network.isolation ? 'Yes' : 'No'}</div>
                        </div>
                      </div>
                    </div>
                  </div>
                )}
              </div>
            ) : (
              <div className="text-center text-muted-foreground py-8">
                Select a tenant to view details
              </div>
            )}
          </Card.Content>
        </Card>
      </div>

      {/* Create Tenant Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <Card className="w-full max-w-md mx-4">
            <Card.Header>
              <Card.Title>Create New Tenant</Card.Title>
            </Card.Header>
            
            <Card.Content className="space-y-4">
              <Input
                label="Tenant Name"
                value={newTenant.name}
                onChange={(e) => setNewTenant({ ...newTenant, name: e.target.value })}
                placeholder="acme-corp"
                helperText="Unique identifier (lowercase, no spaces)"
              />

              <Input
                label="Display Name"
                value={newTenant.displayName}
                onChange={(e) => setNewTenant({ ...newTenant, displayName: e.target.value })}
                placeholder="Acme Corporation"
              />

              <Input
                label="Description"
                value={newTenant.description}
                onChange={(e) => setNewTenant({ ...newTenant, description: e.target.value })}
                placeholder="Description of the tenant"
              />

              <div>
                <label className="block text-sm font-medium mb-2">Tier</label>
                <select
                  className="w-full px-3 py-2 border rounded-md"
                  value={newTenant.tier}
                  onChange={(e) => setNewTenant({ ...newTenant, tier: e.target.value as any })}
                >
                  <option value="free">Free</option>
                  <option value="basic">Basic</option>
                  <option value="premium">Premium</option>
                  <option value="enterprise">Enterprise</option>
                </select>
              </div>

              <Input
                label="Owner Email"
                type="email"
                value={newTenant.ownerEmail}
                onChange={(e) => setNewTenant({ ...newTenant, ownerEmail: e.target.value })}
                placeholder="admin@acme-corp.com"
              />
            </Card.Content>
            
            <Card.Footer className="flex justify-end space-x-2">
              <Button
                variant="outline"
                onClick={() => setShowCreateModal(false)}
              >
                Cancel
              </Button>
              <Button
                onClick={createTenant}
                disabled={!newTenant.name || !newTenant.displayName || !newTenant.ownerEmail}
              >
                Create Tenant
              </Button>
            </Card.Footer>
          </Card>
        </div>
      )}

      {/* Invite User Modal */}
      {showUserModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <Card className="w-full max-w-md mx-4">
            <Card.Header>
              <Card.Title>Invite User</Card.Title>
            </Card.Header>
            
            <Card.Content className="space-y-4">
              <Input
                label="Name"
                value={newUser.name}
                onChange={(e) => setNewUser({ ...newUser, name: e.target.value })}
                placeholder="John Doe"
              />

              <Input
                label="Email"
                type="email"
                value={newUser.email}
                onChange={(e) => setNewUser({ ...newUser, email: e.target.value })}
                placeholder="john@example.com"
              />

              <div>
                <label className="block text-sm font-medium mb-2">Role</label>
                <select
                  className="w-full px-3 py-2 border rounded-md"
                  value={newUser.role}
                  onChange={(e) => setNewUser({ ...newUser, role: e.target.value as any })}
                >
                  <option value="viewer">Viewer</option>
                  <option value="user">User</option>
                  <option value="admin">Admin</option>
                </select>
              </div>
            </Card.Content>
            
            <Card.Footer className="flex justify-end space-x-2">
              <Button
                variant="outline"
                onClick={() => setShowUserModal(false)}
              >
                Cancel
              </Button>
              <Button
                onClick={inviteUser}
                disabled={!newUser.name || !newUser.email}
              >
                Send Invitation
              </Button>
            </Card.Footer>
          </Card>
        </div>
      )}
    </div>
  )
}

export default TenantManagement