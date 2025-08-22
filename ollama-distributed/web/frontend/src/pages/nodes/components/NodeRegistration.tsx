import React, { useState } from 'react'
import { Card, CardHeader, CardContent, CardTitle, CardDescription } from '@/design-system/components/Card/Card'
import type { Node } from '../NodesPage'

interface NodeRegistrationProps {
  onClose: () => void
  onNodeRegistered: (node: Partial<Node>) => void
}

interface RegistrationForm {
  name: string
  hostname: string
  ip_address: string
  port: number
  version: string
  location: {
    region: string
    datacenter: string
    rack: string
  }
  capabilities: {
    gpu_enabled: boolean
    gpu_memory: number
    cpu_cores: number
    ram_total: number
    disk_space: number
    max_concurrent_requests: number
  }
  auth: {
    ssh_key: string
    api_token: string
    username: string
  }
  connection: {
    ssl_enabled: boolean
    verify_ssl: boolean
    timeout: number
  }
}

export const NodeRegistration: React.FC<NodeRegistrationProps> = ({ onClose, onNodeRegistered }) => {
  const [step, setStep] = useState<'basic' | 'hardware' | 'location' | 'auth' | 'review'>('basic')
  const [isRegistering, setIsRegistering] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [testConnection, setTestConnection] = useState<'idle' | 'testing' | 'success' | 'failed'>('idle')
  
  const [form, setForm] = useState<RegistrationForm>({
    name: '',
    hostname: '',
    ip_address: '',
    port: 11434,
    version: '',
    location: {
      region: '',
      datacenter: '',
      rack: ''
    },
    capabilities: {
      gpu_enabled: false,
      gpu_memory: 0,
      cpu_cores: 1,
      ram_total: 1024,
      disk_space: 10240,
      max_concurrent_requests: 4
    },
    auth: {
      ssh_key: '',
      api_token: '',
      username: 'ollama'
    },
    connection: {
      ssl_enabled: true,
      verify_ssl: true,
      timeout: 30
    }
  })

  const updateForm = (path: string, value: any) => {
    setForm(prev => {
      const newForm = { ...prev }
      const keys = path.split('.')
      let current: any = newForm
      
      for (let i = 0; i < keys.length - 1; i++) {
        if (!(keys[i] in current)) {
          current[keys[i]] = {}
        }
        current = current[keys[i]]
      }
      
      current[keys[keys.length - 1]] = value
      return newForm
    })
  }

  const handleTestConnection = async () => {
    setTestConnection('testing')
    setError(null)
    
    try {
      // Simulate connection test
      await new Promise(resolve => setTimeout(resolve, 2000))
      
      // Simulate random success/failure for demo
      if (Math.random() > 0.3) {
        setTestConnection('success')
        
        // Auto-fill some capabilities if connection successful
        if (!form.version) {
          updateForm('version', '0.1.20')
        }
        if (form.capabilities.cpu_cores === 1) {
          updateForm('capabilities.cpu_cores', 8)
          updateForm('capabilities.ram_total', 16384)
          updateForm('capabilities.disk_space', 102400)
        }
      } else {
        setTestConnection('failed')
        setError('Failed to connect to node. Please check the hostname/IP and port.')
      }
    } catch (err) {
      setTestConnection('failed')
      setError('Connection test failed. Please verify the node is accessible.')
    }
  }

  const handleRegister = async () => {
    setIsRegistering(true)
    setError(null)
    
    try {
      // Simulate registration process
      await new Promise(resolve => setTimeout(resolve, 3000))
      
      const newNode: Partial<Node> = {
        id: `node-${Date.now()}`,
        name: form.name,
        hostname: form.hostname,
        ip_address: form.ip_address,
        port: form.port,
        status: 'online',
        health_score: 95,
        last_seen: new Date().toISOString(),
        version: form.version,
        location: {
          region: form.location.region,
          datacenter: form.location.datacenter,
          rack: form.location.rack || undefined
        },
        capabilities: {
          models: ['llama2:7b'], // Default model
          ...form.capabilities
        },
        resources: {
          cpu_usage: Math.random() * 30 + 10,
          memory_usage: Math.random() * 40 + 20,
          gpu_usage: form.capabilities.gpu_enabled ? Math.random() * 50 + 10 : undefined,
          disk_usage: Math.random() * 30 + 15,
          network_io: {
            rx_bytes: 0,
            tx_bytes: 0
          },
          active_requests: 0,
          queue_size: 0
        },
        performance: {
          requests_per_second: 0,
          average_response_time: 0,
          error_rate: 0,
          uptime: 100,
          tokens_per_second: 0
        },
        alerts: []
      }
      
      onNodeRegistered(newNode)
    } catch (err) {
      setError('Failed to register node. Please try again.')
      setIsRegistering(false)
    }
  }

  const isStepValid = () => {
    switch (step) {
      case 'basic':
        return form.name && form.hostname && form.ip_address && form.port
      case 'hardware':
        return form.capabilities.cpu_cores > 0 && form.capabilities.ram_total > 0
      case 'location':
        return form.location.region && form.location.datacenter
      case 'auth':
        return form.auth.username
      case 'review':
        return true
      default:
        return false
    }
  }

  const steps = [
    { id: 'basic', label: 'Basic Info', icon: '📡' },
    { id: 'hardware', label: 'Hardware', icon: '🖥️' },
    { id: 'location', label: 'Location', icon: '📍' },
    { id: 'auth', label: 'Authentication', icon: '🔐' },
    { id: 'review', label: 'Review', icon: '✅' }
  ]

  const currentStepIndex = steps.findIndex(s => s.id === step)

  const renderBasicStep = () => (
    <div className="space-y-4">
      <div>
        <label className="block text-sm font-medium text-foreground mb-1">
          Node Name *
        </label>
        <input
          type="text"
          value={form.name}
          onChange={(e) => updateForm('name', e.target.value)}
          placeholder="e.g., gpu-node-01"
          className="w-full px-3 py-2 border border-border rounded-lg bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-foreground mb-1">
          Hostname *
        </label>
        <input
          type="text"
          value={form.hostname}
          onChange={(e) => updateForm('hostname', e.target.value)}
          placeholder="e.g., gpu-01.ollama.local"
          className="w-full px-3 py-2 border border-border rounded-lg bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
        />
      </div>

      <div className="grid grid-cols-3 gap-3">
        <div className="col-span-2">
          <label className="block text-sm font-medium text-foreground mb-1">
            IP Address *
          </label>
          <input
            type="text"
            value={form.ip_address}
            onChange={(e) => updateForm('ip_address', e.target.value)}
            placeholder="192.168.1.10"
            className="w-full px-3 py-2 border border-border rounded-lg bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-foreground mb-1">
            Port *
          </label>
          <input
            type="number"
            value={form.port}
            onChange={(e) => updateForm('port', parseInt(e.target.value) || 11434)}
            className="w-full px-3 py-2 border border-border rounded-lg bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
          />
        </div>
      </div>

      <div>
        <label className="block text-sm font-medium text-foreground mb-1">
          Ollama Version
        </label>
        <input
          type="text"
          value={form.version}
          onChange={(e) => updateForm('version', e.target.value)}
          placeholder="Will be detected automatically"
          className="w-full px-3 py-2 border border-border rounded-lg bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
        />
      </div>

      {form.ip_address && form.port && (
        <div className="pt-4 border-t border-border">
          <button
            onClick={handleTestConnection}
            disabled={testConnection === 'testing'}
            className={`w-full px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
              testConnection === 'testing' 
                ? 'bg-muted text-muted-foreground cursor-wait'
                : testConnection === 'success'
                ? 'bg-success text-success-foreground'
                : testConnection === 'failed'
                ? 'bg-error text-error-foreground'
                : 'bg-primary text-primary-foreground hover:bg-primary/90'
            }`}
          >
            {testConnection === 'testing' && (
              <svg className="animate-spin w-4 h-4 inline mr-2" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"/>
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/>
              </svg>
            )}
            {testConnection === 'testing' ? 'Testing Connection...' :
             testConnection === 'success' ? 'Connection Successful ✓' :
             testConnection === 'failed' ? 'Connection Failed ✗' :
             'Test Connection'}
          </button>
          
          {testConnection === 'success' && (
            <p className="text-xs text-success mt-2">
              Node is reachable and responding to requests.
            </p>
          )}
        </div>
      )}
    </div>
  )

  const renderHardwareStep = () => (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-foreground mb-1">
            CPU Cores *
          </label>
          <input
            type="number"
            min="1"
            max="128"
            value={form.capabilities.cpu_cores}
            onChange={(e) => updateForm('capabilities.cpu_cores', parseInt(e.target.value) || 1)}
            className="w-full px-3 py-2 border border-border rounded-lg bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-foreground mb-1">
            RAM (MB) *
          </label>
          <input
            type="number"
            min="512"
            value={form.capabilities.ram_total}
            onChange={(e) => updateForm('capabilities.ram_total', parseInt(e.target.value) || 1024)}
            className="w-full px-3 py-2 border border-border rounded-lg bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
          />
        </div>
      </div>

      <div>
        <label className="block text-sm font-medium text-foreground mb-1">
          Disk Space (MB) *
        </label>
        <input
          type="number"
          min="1024"
          value={form.capabilities.disk_space}
          onChange={(e) => updateForm('capabilities.disk_space', parseInt(e.target.value) || 10240)}
          className="w-full px-3 py-2 border border-border rounded-lg bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-foreground mb-1">
          Max Concurrent Requests
        </label>
        <input
          type="number"
          min="1"
          max="32"
          value={form.capabilities.max_concurrent_requests}
          onChange={(e) => updateForm('capabilities.max_concurrent_requests', parseInt(e.target.value) || 4)}
          className="w-full px-3 py-2 border border-border rounded-lg bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
        />
      </div>

      <div className="pt-4 border-t border-border">
        <div className="flex items-center gap-2 mb-3">
          <input
            type="checkbox"
            id="gpu_enabled"
            checked={form.capabilities.gpu_enabled}
            onChange={(e) => updateForm('capabilities.gpu_enabled', e.target.checked)}
            className="w-4 h-4 text-primary border-border rounded focus:ring-primary"
          />
          <label htmlFor="gpu_enabled" className="text-sm font-medium text-foreground">
            GPU Enabled
          </label>
        </div>

        {form.capabilities.gpu_enabled && (
          <div>
            <label className="block text-sm font-medium text-foreground mb-1">
              GPU Memory (MB)
            </label>
            <input
              type="number"
              min="0"
              value={form.capabilities.gpu_memory}
              onChange={(e) => updateForm('capabilities.gpu_memory', parseInt(e.target.value) || 0)}
              className="w-full px-3 py-2 border border-border rounded-lg bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
            />
          </div>
        )}
      </div>
    </div>
  )

  const renderLocationStep = () => (
    <div className="space-y-4">
      <div>
        <label className="block text-sm font-medium text-foreground mb-1">
          Region *
        </label>
        <select
          value={form.location.region}
          onChange={(e) => updateForm('location.region', e.target.value)}
          className="w-full px-3 py-2 border border-border rounded-lg bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
        >
          <option value="">Select Region</option>
          <option value="us-west-1">US West 1</option>
          <option value="us-west-2">US West 2</option>
          <option value="us-east-1">US East 1</option>
          <option value="us-east-2">US East 2</option>
          <option value="eu-west-1">EU West 1</option>
          <option value="eu-central-1">EU Central 1</option>
          <option value="ap-southeast-1">AP Southeast 1</option>
          <option value="ap-northeast-1">AP Northeast 1</option>
        </select>
      </div>

      <div>
        <label className="block text-sm font-medium text-foreground mb-1">
          Datacenter *
        </label>
        <input
          type="text"
          value={form.location.datacenter}
          onChange={(e) => updateForm('location.datacenter', e.target.value)}
          placeholder="e.g., dc-01, main-dc, edge-dc-01"
          className="w-full px-3 py-2 border border-border rounded-lg bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-foreground mb-1">
          Rack (Optional)
        </label>
        <input
          type="text"
          value={form.location.rack}
          onChange={(e) => updateForm('location.rack', e.target.value)}
          placeholder="e.g., rack-a1, r01, cabinet-12"
          className="w-full px-3 py-2 border border-border rounded-lg bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
        />
      </div>
    </div>
  )

  const renderAuthStep = () => (
    <div className="space-y-4">
      <div>
        <label className="block text-sm font-medium text-foreground mb-1">
          Username *
        </label>
        <input
          type="text"
          value={form.auth.username}
          onChange={(e) => updateForm('auth.username', e.target.value)}
          placeholder="ollama"
          className="w-full px-3 py-2 border border-border rounded-lg bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-foreground mb-1">
          API Token (Optional)
        </label>
        <input
          type="password"
          value={form.auth.api_token}
          onChange={(e) => updateForm('auth.api_token', e.target.value)}
          placeholder="Enter API token if required"
          className="w-full px-3 py-2 border border-border rounded-lg bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-foreground mb-1">
          SSH Public Key (Optional)
        </label>
        <textarea
          value={form.auth.ssh_key}
          onChange={(e) => updateForm('auth.ssh_key', e.target.value)}
          placeholder="ssh-rsa AAAAB3NzaC1yc2E..."
          rows={3}
          className="w-full px-3 py-2 border border-border rounded-lg bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
        />
      </div>

      <div className="space-y-3 pt-4 border-t border-border">
        <h4 className="text-sm font-medium text-foreground">Connection Settings</h4>
        
        <div className="flex items-center gap-2">
          <input
            type="checkbox"
            id="ssl_enabled"
            checked={form.connection.ssl_enabled}
            onChange={(e) => updateForm('connection.ssl_enabled', e.target.checked)}
            className="w-4 h-4 text-primary border-border rounded focus:ring-primary"
          />
          <label htmlFor="ssl_enabled" className="text-sm text-foreground">
            Enable SSL/TLS
          </label>
        </div>

        {form.connection.ssl_enabled && (
          <div className="flex items-center gap-2 ml-6">
            <input
              type="checkbox"
              id="verify_ssl"
              checked={form.connection.verify_ssl}
              onChange={(e) => updateForm('connection.verify_ssl', e.target.checked)}
              className="w-4 h-4 text-primary border-border rounded focus:ring-primary"
            />
            <label htmlFor="verify_ssl" className="text-sm text-foreground">
              Verify SSL certificates
            </label>
          </div>
        )}

        <div>
          <label className="block text-sm font-medium text-foreground mb-1">
            Connection Timeout (seconds)
          </label>
          <input
            type="number"
            min="5"
            max="300"
            value={form.connection.timeout}
            onChange={(e) => updateForm('connection.timeout', parseInt(e.target.value) || 30)}
            className="w-full px-3 py-2 border border-border rounded-lg bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
          />
        </div>
      </div>
    </div>
  )

  const renderReviewStep = () => (
    <div className="space-y-4">
      <div className="p-4 bg-muted/50 rounded-lg">
        <h4 className="text-sm font-medium text-foreground mb-3">Node Configuration</h4>
        <div className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <span className="text-muted-foreground">Name:</span>
            <span className="font-medium ml-2">{form.name}</span>
          </div>
          <div>
            <span className="text-muted-foreground">Hostname:</span>
            <span className="font-medium ml-2">{form.hostname}</span>
          </div>
          <div>
            <span className="text-muted-foreground">Address:</span>
            <span className="font-medium ml-2">{form.ip_address}:{form.port}</span>
          </div>
          <div>
            <span className="text-muted-foreground">Version:</span>
            <span className="font-medium ml-2">{form.version || 'Auto-detect'}</span>
          </div>
        </div>
      </div>

      <div className="p-4 bg-muted/50 rounded-lg">
        <h4 className="text-sm font-medium text-foreground mb-3">Hardware Specifications</h4>
        <div className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <span className="text-muted-foreground">CPU Cores:</span>
            <span className="font-medium ml-2">{form.capabilities.cpu_cores}</span>
          </div>
          <div>
            <span className="text-muted-foreground">RAM:</span>
            <span className="font-medium ml-2">{(form.capabilities.ram_total / 1024).toFixed(1)} GB</span>
          </div>
          <div>
            <span className="text-muted-foreground">Disk:</span>
            <span className="font-medium ml-2">{(form.capabilities.disk_space / 1024).toFixed(1)} GB</span>
          </div>
          <div>
            <span className="text-muted-foreground">GPU:</span>
            <span className="font-medium ml-2">
              {form.capabilities.gpu_enabled 
                ? `Enabled (${(form.capabilities.gpu_memory / 1024).toFixed(1)} GB)` 
                : 'Disabled'
              }
            </span>
          </div>
          <div>
            <span className="text-muted-foreground">Max Requests:</span>
            <span className="font-medium ml-2">{form.capabilities.max_concurrent_requests}</span>
          </div>
        </div>
      </div>

      <div className="p-4 bg-muted/50 rounded-lg">
        <h4 className="text-sm font-medium text-foreground mb-3">Location & Access</h4>
        <div className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <span className="text-muted-foreground">Region:</span>
            <span className="font-medium ml-2">{form.location.region}</span>
          </div>
          <div>
            <span className="text-muted-foreground">Datacenter:</span>
            <span className="font-medium ml-2">{form.location.datacenter}</span>
          </div>
          {form.location.rack && (
            <div>
              <span className="text-muted-foreground">Rack:</span>
              <span className="font-medium ml-2">{form.location.rack}</span>
            </div>
          )}
          <div>
            <span className="text-muted-foreground">Username:</span>
            <span className="font-medium ml-2">{form.auth.username}</span>
          </div>
          <div>
            <span className="text-muted-foreground">SSL:</span>
            <span className="font-medium ml-2">{form.connection.ssl_enabled ? 'Enabled' : 'Disabled'}</span>
          </div>
          <div>
            <span className="text-muted-foreground">Timeout:</span>
            <span className="font-medium ml-2">{form.connection.timeout}s</span>
          </div>
        </div>
      </div>
    </div>
  )

  const renderStepContent = () => {
    switch (step) {
      case 'basic': return renderBasicStep()
      case 'hardware': return renderHardwareStep()
      case 'location': return renderLocationStep()
      case 'auth': return renderAuthStep()
      case 'review': return renderReviewStep()
      default: return null
    }
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <Card className="w-full max-w-2xl max-h-[90vh] overflow-y-auto">
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Register New Node</CardTitle>
              <CardDescription>
                Add a new Ollama node to your distributed cluster
              </CardDescription>
            </div>
            <button
              onClick={onClose}
              className="p-1 hover:bg-muted rounded-lg transition-colors"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          {/* Progress Steps */}
          <div className="flex items-center justify-between mt-6">
            {steps.map((stepInfo, index) => (
              <div key={stepInfo.id} className="flex items-center">
                <div className={`flex items-center justify-center w-8 h-8 rounded-full text-sm font-medium transition-colors ${
                  index <= currentStepIndex 
                    ? 'bg-primary text-primary-foreground' 
                    : 'bg-muted text-muted-foreground'
                }`}>
                  {stepInfo.icon}
                </div>
                <div className="ml-2 hidden sm:block">
                  <div className={`text-xs font-medium ${
                    index <= currentStepIndex ? 'text-foreground' : 'text-muted-foreground'
                  }`}>
                    {stepInfo.label}
                  </div>
                </div>
                {index < steps.length - 1 && (
                  <div className={`w-8 h-0.5 mx-4 transition-colors ${
                    index < currentStepIndex ? 'bg-primary' : 'bg-muted'
                  }`} />
                )}
              </div>
            ))}
          </div>
        </CardHeader>

        <CardContent>
          {error && (
            <div className="mb-4 p-3 bg-error/10 border border-error/20 rounded-lg">
              <p className="text-sm text-error">{error}</p>
            </div>
          )}

          {renderStepContent()}

          <div className="flex justify-between mt-8 pt-6 border-t border-border">
            <button
              onClick={() => {
                const currentIndex = steps.findIndex(s => s.id === step)
                if (currentIndex > 0) {
                  setStep(steps[currentIndex - 1].id as any)
                } else {
                  onClose()
                }
              }}
              className="px-4 py-2 bg-muted text-muted-foreground rounded-lg hover:bg-muted/80 transition-colors"
            >
              {currentStepIndex === 0 ? 'Cancel' : 'Previous'}
            </button>

            {step === 'review' ? (
              <button
                onClick={handleRegister}
                disabled={isRegistering}
                className={`px-6 py-2 rounded-lg text-sm font-medium transition-colors ${
                  isRegistering 
                    ? 'bg-muted text-muted-foreground cursor-wait'
                    : 'bg-primary text-primary-foreground hover:bg-primary/90'
                }`}
              >
                {isRegistering ? (
                  <>
                    <svg className="animate-spin w-4 h-4 inline mr-2" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"/>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/>
                    </svg>
                    Registering Node...
                  </>
                ) : (
                  'Register Node'
                )}
              </button>
            ) : (
              <button
                onClick={() => {
                  const currentIndex = steps.findIndex(s => s.id === step)
                  setStep(steps[currentIndex + 1].id as any)
                }}
                disabled={!isStepValid()}
                className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors disabled:bg-muted disabled:text-muted-foreground disabled:cursor-not-allowed"
              >
                Next
              </button>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

export default NodeRegistration