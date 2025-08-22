import React, { useState, useCallback, useEffect } from 'react'
import { 
  Globe, Shield, Github, Mail, Building, Key, ExternalLink,
  AlertCircle, CheckCircle, Settings, Plus, Trash2, Edit,
  Loader2, RefreshCw
} from 'lucide-react'
import { apiClient } from '@/services/api/client'
import type { SSOProvider, SSOConfig, OAuthState } from '@/types/auth'

interface SSOProvidersProps {
  onProviderSelect?: (provider: SSOProvider) => void
  showManagement?: boolean
  className?: string
}

interface ProviderTemplate {
  type: 'oauth2' | 'saml' | 'ldap'
  name: string
  icon: React.ReactNode
  description: string
  configFields: Array<{
    key: string
    label: string
    type: 'text' | 'url' | 'password' | 'number'
    required: boolean
    description?: string
  }>
}

export const SSOProviders: React.FC<SSOProvidersProps> = ({
  onProviderSelect,
  showManagement = false,
  className = ''
}) => {
  const [providers, setProviders] = useState<SSOProvider[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [showAddModal, setShowAddModal] = useState(false)
  const [editingProvider, setEditingProvider] = useState<SSOProvider | null>(null)
  const [testingProvider, setTestingProvider] = useState<string | null>(null)
  const [newProvider, setNewProvider] = useState<Partial<SSOProvider>>({
    name: '',
    type: 'oauth2',
    enabled: true,
    config: {}
  })

  const providerTemplates: ProviderTemplate[] = [
    {
      type: 'oauth2',
      name: 'Google',
      icon: <Globe className="w-5 h-5" />,
      description: 'Google OAuth 2.0 authentication',
      configFields: [
        { key: 'clientId', label: 'Client ID', type: 'text', required: true },
        { key: 'clientSecret', label: 'Client Secret', type: 'password', required: true },
        { key: 'authUrl', label: 'Authorization URL', type: 'url', required: true },
        { key: 'tokenUrl', label: 'Token URL', type: 'url', required: true },
        { key: 'userInfoUrl', label: 'User Info URL', type: 'url', required: true },
        { key: 'scopes', label: 'Scopes', type: 'text', required: false, description: 'Comma-separated list' }
      ]
    },
    {
      type: 'oauth2',
      name: 'GitHub',
      icon: <Github className="w-5 h-5" />,
      description: 'GitHub OAuth authentication',
      configFields: [
        { key: 'clientId', label: 'Client ID', type: 'text', required: true },
        { key: 'clientSecret', label: 'Client Secret', type: 'password', required: true },
        { key: 'authUrl', label: 'Authorization URL', type: 'url', required: true },
        { key: 'tokenUrl', label: 'Token URL', type: 'url', required: true },
        { key: 'userInfoUrl', label: 'User Info URL', type: 'url', required: true }
      ]
    },
    {
      type: 'oauth2',
      name: 'Microsoft',
      icon: <Building className="w-5 h-5" />,
      description: 'Microsoft Azure AD authentication',
      configFields: [
        { key: 'clientId', label: 'Application ID', type: 'text', required: true },
        { key: 'clientSecret', label: 'Client Secret', type: 'password', required: true },
        { key: 'tenantId', label: 'Tenant ID', type: 'text', required: true },
        { key: 'authUrl', label: 'Authorization URL', type: 'url', required: true },
        { key: 'tokenUrl', label: 'Token URL', type: 'url', required: true }
      ]
    },
    {
      type: 'saml',
      name: 'SAML 2.0',
      icon: <Shield className="w-5 h-5" />,
      description: 'SAML 2.0 single sign-on',
      configFields: [
        { key: 'ssoUrl', label: 'SSO URL', type: 'url', required: true },
        { key: 'entityId', label: 'Entity ID', type: 'text', required: true },
        { key: 'certificate', label: 'X.509 Certificate', type: 'text', required: true }
      ]
    },
    {
      type: 'ldap',
      name: 'LDAP/AD',
      icon: <Building className="w-5 h-5" />,
      description: 'LDAP or Active Directory',
      configFields: [
        { key: 'host', label: 'LDAP Host', type: 'text', required: true },
        { key: 'port', label: 'Port', type: 'number', required: true },
        { key: 'baseDN', label: 'Base DN', type: 'text', required: true },
        { key: 'userFilter', label: 'User Filter', type: 'text', required: true },
        { key: 'bindDN', label: 'Bind DN', type: 'text', required: false },
        { key: 'bindPassword', label: 'Bind Password', type: 'password', required: false }
      ]
    }
  ]

  useEffect(() => {
    loadProviders()
  }, [])

  const loadProviders = useCallback(async () => {
    setIsLoading(true)
    try {
      const response = await apiClient.request<SSOProvider[]>('/auth/sso/providers')
      if (response.success) {
        setProviders(response.data)
      }
    } catch (err) {
      setError('Failed to load SSO providers')
    } finally {
      setIsLoading(false)
    }
  }, [])

  const handleProviderAuth = useCallback(async (provider: SSOProvider) => {
    if (!provider.enabled) return

    try {
      // Generate OAuth state for security
      const state: OAuthState = {
        provider: provider.id,
        redirectUrl: window.location.href,
        nonce: crypto.getRandomValues(new Uint32Array(1))[0].toString(16)
      }

      // Store state in session storage
      sessionStorage.setItem('oauth_state', JSON.stringify(state))

      // Get authorization URL from server
      const response = await apiClient.request<{ authUrl: string }>(`/auth/sso/${provider.id}/authorize`, {
        method: 'POST',
        body: { state: state.nonce }
      })

      if (response.success) {
        // Redirect to provider
        window.location.href = response.data.authUrl
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to initiate SSO authentication')
    }
  }, [])

  const handleCreateProvider = useCallback(async () => {
    if (!newProvider.name || !newProvider.type) {
      setError('Provider name and type are required')
      return
    }

    try {
      const response = await apiClient.request<SSOProvider>('/auth/sso/providers', {
        method: 'POST',
        body: newProvider
      })

      if (response.success) {
        setProviders(prev => [...prev, response.data])
        setShowAddModal(false)
        setNewProvider({ name: '', type: 'oauth2', enabled: true, config: {} })
        setError(null)
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create provider')
    }
  }, [newProvider])

  const handleUpdateProvider = useCallback(async (providerId: string, updates: Partial<SSOProvider>) => {
    try {
      const response = await apiClient.request<SSOProvider>(`/auth/sso/providers/${providerId}`, {
        method: 'PUT',
        body: updates
      })

      if (response.success) {
        setProviders(prev => prev.map(p => p.id === providerId ? response.data : p))
        setEditingProvider(null)
        setError(null)
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update provider')
    }
  }, [])

  const handleDeleteProvider = useCallback(async (providerId: string) => {
    if (!confirm('Are you sure you want to delete this SSO provider?')) return

    try {
      await apiClient.request(`/auth/sso/providers/${providerId}`, {
        method: 'DELETE'
      })

      setProviders(prev => prev.filter(p => p.id !== providerId))
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete provider')
    }
  }, [])

  const handleTestProvider = useCallback(async (providerId: string) => {
    setTestingProvider(providerId)
    try {
      const response = await apiClient.request<{ status: string; message: string }>(`/auth/sso/providers/${providerId}/test`, {
        method: 'POST'
      })

      if (response.success) {
        if (response.data.status === 'success') {
          alert('Provider test successful!')
        } else {
          setError(`Provider test failed: ${response.data.message}`)
        }
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to test provider')
    } finally {
      setTestingProvider(null)
    }
  }, [])

  const getProviderIcon = (provider: SSOProvider) => {
    const template = providerTemplates.find(t => t.name.toLowerCase() === provider.name.toLowerCase())
    if (template) return template.icon

    switch (provider.type) {
      case 'oauth2':
        return <Key className="w-5 h-5" />
      case 'saml':
        return <Shield className="w-5 h-5" />
      case 'ldap':
        return <Building className="w-5 h-5" />
      default:
        return <Globe className="w-5 h-5" />
    }
  }

  if (!showManagement) {
    // Simple provider list for login
    return (
      <div className={`space-y-3 ${className}`}>
        {isLoading ? (
          <div className="text-center py-4">
            <Loader2 className="w-6 h-6 animate-spin mx-auto text-blue-600" />
          </div>
        ) : (
          providers.filter(p => p.enabled).map((provider) => (
            <button
              key={provider.id}
              onClick={() => handleProviderAuth(provider)}
              className="w-full flex items-center justify-center px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm bg-white dark:bg-gray-700 text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-600 transition-colors"
            >
              {getProviderIcon(provider)}
              <span className="ml-2">Continue with {provider.name}</span>
              <ExternalLink className="w-4 h-4 ml-2" />
            </button>
          ))
        )}

        {error && (
          <div className="p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md">
            <div className="flex items-center">
              <AlertCircle className="w-4 h-4 text-red-500 mr-2" />
              <span className="text-sm text-red-700 dark:text-red-400">{error}</span>
            </div>
          </div>
        )}
      </div>
    )
  }

  // Full management interface
  return (
    <div className={`w-full ${className}`}>
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700">
        {/* Header */}
        <div className="p-6 border-b border-gray-200 dark:border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-xl font-semibold text-gray-900 dark:text-white">
                SSO Providers
              </h2>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                Configure single sign-on authentication providers
              </p>
            </div>
            <button
              onClick={() => setShowAddModal(true)}
              className="flex items-center px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
            >
              <Plus className="w-4 h-4 mr-2" />
              Add Provider
            </button>
          </div>
        </div>

        {/* Error Message */}
        {error && (
          <div className="mx-6 mt-4 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md">
            <div className="flex items-center">
              <AlertCircle className="w-4 h-4 text-red-500 mr-2" />
              <span className="text-sm text-red-700 dark:text-red-400">{error}</span>
            </div>
          </div>
        )}

        {/* Providers List */}
        <div className="p-6">
          {isLoading ? (
            <div className="text-center py-12">
              <Loader2 className="w-8 h-8 animate-spin mx-auto text-blue-600" />
            </div>
          ) : providers.length === 0 ? (
            <div className="text-center py-12">
              <Shield className="w-12 h-12 text-gray-400 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">
                No SSO Providers
              </h3>
              <p className="text-gray-600 dark:text-gray-400 mb-4">
                Add an SSO provider to enable single sign-on authentication
              </p>
              <button
                onClick={() => setShowAddModal(true)}
                className="flex items-center mx-auto px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
              >
                <Plus className="w-4 h-4 mr-2" />
                Add Your First Provider
              </button>
            </div>
          ) : (
            <div className="space-y-4">
              {providers.map((provider) => (
                <div
                  key={provider.id}
                  className="border border-gray-200 dark:border-gray-600 rounded-lg p-4"
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center">
                      <div className={`
                        p-2 rounded-lg mr-4
                        ${provider.enabled 
                          ? 'bg-green-100 dark:bg-green-900 text-green-600 dark:text-green-400'
                          : 'bg-gray-100 dark:bg-gray-700 text-gray-400'
                        }
                      `}>
                        {getProviderIcon(provider)}
                      </div>
                      <div>
                        <div className="flex items-center">
                          <h3 className="text-lg font-medium text-gray-900 dark:text-white">
                            {provider.name}
                          </h3>
                          <span className={`
                            ml-3 px-2 py-1 text-xs rounded-full
                            ${provider.enabled
                              ? 'bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200'
                              : 'bg-gray-100 dark:bg-gray-700 text-gray-800 dark:text-gray-200'
                            }
                          `}>
                            {provider.enabled ? 'Enabled' : 'Disabled'}
                          </span>
                        </div>
                        <p className="text-sm text-gray-600 dark:text-gray-400">
                          {provider.type.toUpperCase()} • Last updated: {new Date().toLocaleDateString()}
                        </p>
                      </div>
                    </div>

                    <div className="flex items-center space-x-2">
                      <button
                        onClick={() => handleTestProvider(provider.id)}
                        disabled={testingProvider === provider.id || !provider.enabled}
                        className="p-2 text-blue-600 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded disabled:opacity-50"
                        title="Test provider"
                      >
                        {testingProvider === provider.id ? (
                          <Loader2 className="w-4 h-4 animate-spin" />
                        ) : (
                          <RefreshCw className="w-4 h-4" />
                        )}
                      </button>
                      <button
                        onClick={() => setEditingProvider(provider)}
                        className="p-2 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-700 rounded"
                        title="Edit provider"
                      >
                        <Edit className="w-4 h-4" />
                      </button>
                      <button
                        onClick={() => handleDeleteProvider(provider.id)}
                        className="p-2 text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 rounded"
                        title="Delete provider"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  </div>

                  {/* Provider Configuration Summary */}
                  <div className="mt-4 grid grid-cols-2 gap-4 text-sm">
                    {provider.type === 'oauth2' && (
                      <>
                        <div>
                          <span className="text-gray-500 dark:text-gray-400">Client ID:</span>
                          <span className="ml-2 text-gray-900 dark:text-white">
                            {provider.config.clientId ? '••••••••' : 'Not configured'}
                          </span>
                        </div>
                        <div>
                          <span className="text-gray-500 dark:text-gray-400">Auth URL:</span>
                          <span className="ml-2 text-gray-900 dark:text-white">
                            {provider.config.authUrl || 'Not configured'}
                          </span>
                        </div>
                      </>
                    )}
                    {provider.type === 'saml' && (
                      <>
                        <div>
                          <span className="text-gray-500 dark:text-gray-400">SSO URL:</span>
                          <span className="ml-2 text-gray-900 dark:text-white">
                            {provider.config.ssoUrl || 'Not configured'}
                          </span>
                        </div>
                        <div>
                          <span className="text-gray-500 dark:text-gray-400">Entity ID:</span>
                          <span className="ml-2 text-gray-900 dark:text-white">
                            {provider.config.entityId || 'Not configured'}
                          </span>
                        </div>
                      </>
                    )}
                    {provider.type === 'ldap' && (
                      <>
                        <div>
                          <span className="text-gray-500 dark:text-gray-400">Host:</span>
                          <span className="ml-2 text-gray-900 dark:text-white">
                            {provider.config.host || 'Not configured'}
                          </span>
                        </div>
                        <div>
                          <span className="text-gray-500 dark:text-gray-400">Base DN:</span>
                          <span className="ml-2 text-gray-900 dark:text-white">
                            {provider.config.baseDN || 'Not configured'}
                          </span>
                        </div>
                      </>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Add Provider Modal */}
      {showAddModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg w-full max-w-2xl mx-4 max-h-[90vh] overflow-y-auto">
            <div className="p-6 border-b border-gray-200 dark:border-gray-700">
              <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
                Add SSO Provider
              </h2>
            </div>
            <div className="p-6">
              {/* Provider Template Selection */}
              <div className="mb-6">
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">
                  Choose Provider Template
                </label>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  {providerTemplates.map((template) => (
                    <button
                      key={template.name}
                      onClick={() => setNewProvider(prev => ({
                        ...prev,
                        name: template.name,
                        type: template.type,
                        config: {}
                      }))}
                      className={`
                        p-3 border rounded-lg text-left transition-colors
                        ${newProvider.name === template.name
                          ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
                          : 'border-gray-200 dark:border-gray-600 hover:border-gray-300'
                        }
                      `}
                    >
                      <div className="flex items-center mb-2">
                        {template.icon}
                        <span className="ml-2 font-medium text-gray-900 dark:text-white">
                          {template.name}
                        </span>
                      </div>
                      <p className="text-sm text-gray-600 dark:text-gray-400">
                        {template.description}
                      </p>
                    </button>
                  ))}
                </div>
              </div>

              {/* Configuration Fields */}
              {newProvider.name && (
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                      Provider Name
                    </label>
                    <input
                      type="text"
                      value={newProvider.name}
                      onChange={(e) => setNewProvider(prev => ({ ...prev, name: e.target.value }))}
                      className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                    />
                  </div>

                  {providerTemplates
                    .find(t => t.name === newProvider.name)
                    ?.configFields.map((field) => (
                      <div key={field.key}>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                          {field.label} {field.required && <span className="text-red-500">*</span>}
                        </label>
                        <input
                          type={field.type}
                          value={newProvider.config?.[field.key] || ''}
                          onChange={(e) => setNewProvider(prev => ({
                            ...prev,
                            config: {
                              ...prev.config,
                              [field.key]: field.type === 'number' ? parseInt(e.target.value) : e.target.value
                            }
                          }))}
                          className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                          placeholder={field.description}
                        />
                        {field.description && (
                          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                            {field.description}
                          </p>
                        )}
                      </div>
                    ))}

                  <div className="flex items-center">
                    <input
                      type="checkbox"
                      id="enabled"
                      checked={newProvider.enabled}
                      onChange={(e) => setNewProvider(prev => ({ ...prev, enabled: e.target.checked }))}
                      className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                    />
                    <label htmlFor="enabled" className="ml-2 text-sm text-gray-700 dark:text-gray-300">
                      Enable this provider
                    </label>
                  </div>
                </div>
              )}
            </div>
            <div className="p-6 border-t border-gray-200 dark:border-gray-700 flex justify-end space-x-3">
              <button
                onClick={() => {
                  setShowAddModal(false)
                  setNewProvider({ name: '', type: 'oauth2', enabled: true, config: {} })
                }}
                className="px-4 py-2 text-sm text-gray-700 dark:text-gray-300 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700"
              >
                Cancel
              </button>
              <button
                onClick={handleCreateProvider}
                disabled={!newProvider.name}
                className="px-4 py-2 text-sm text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:opacity-50"
              >
                Create Provider
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default SSOProviders