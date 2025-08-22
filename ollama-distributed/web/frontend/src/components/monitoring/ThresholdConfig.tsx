/**
 * Threshold Configuration Component
 * Manages alert thresholds and notification settings
 */

import React, { useState } from 'react'
import {
  Plus,
  Edit,
  Trash2,
  Save,
  X,
  AlertTriangle,
  Settings,
  Bell,
  BellOff,
  Clock
} from 'lucide-react'
import {
  AlertThreshold,
  AlertSeverity,
  ThresholdOperator,
  ThresholdCondition,
  AlertAction
} from '../../types/monitoring'

interface ThresholdConfigProps {
  thresholds: AlertThreshold[]
  onCreateThreshold: (threshold: Omit<AlertThreshold, 'id' | 'createdAt' | 'updatedAt'>) => void
  onUpdateThreshold: (id: string, threshold: Partial<AlertThreshold>) => void
  onDeleteThreshold: (id: string) => void
  className?: string
}

interface ThresholdFormData {
  name: string
  metric: string
  operator: ThresholdOperator
  value: number
  severity: AlertSeverity
  enabled: boolean
  conditions: ThresholdCondition[]
  actions: AlertAction[]
  cooldown: number
}

const OPERATORS: { value: ThresholdOperator; label: string }[] = [
  { value: 'gt', label: 'Greater than' },
  { value: 'gte', label: 'Greater than or equal' },
  { value: 'lt', label: 'Less than' },
  { value: 'lte', label: 'Less than or equal' },
  { value: 'eq', label: 'Equal to' },
  { value: 'neq', label: 'Not equal to' }
]

const SEVERITIES: AlertSeverity[] = ['info', 'warning', 'error', 'critical']

const SEVERITY_COLORS = {
  info: 'bg-blue-100 text-blue-800',
  warning: 'bg-yellow-100 text-yellow-800',
  error: 'bg-red-100 text-red-800',
  critical: 'bg-red-200 text-red-900'
}

const AVAILABLE_METRICS = [
  { value: 'system.cpu.usage', label: 'CPU Usage (%)' },
  { value: 'system.memory.usage', label: 'Memory Usage (%)' },
  { value: 'system.disk.usage', label: 'Disk Usage (%)' },
  { value: 'system.network.latency', label: 'Network Latency (ms)' },
  { value: 'cluster.response_time', label: 'Response Time (ms)' },
  { value: 'cluster.error_rate', label: 'Error Rate (%)' },
  { value: 'cluster.throughput', label: 'Throughput (req/s)' },
  { value: 'model.response_time', label: 'Model Response Time (ms)' },
  { value: 'model.error_rate', label: 'Model Error Rate (%)' }
]

const ThresholdForm: React.FC<{
  threshold?: AlertThreshold
  onSave: (data: ThresholdFormData) => void
  onCancel: () => void
}> = ({ threshold, onSave, onCancel }) => {
  const [formData, setFormData] = useState<ThresholdFormData>({
    name: threshold?.name || '',
    metric: threshold?.metric || '',
    operator: threshold?.operator || 'gt',
    value: threshold?.value || 0,
    severity: threshold?.severity || 'warning',
    enabled: threshold?.enabled ?? true,
    conditions: threshold?.conditions || [],
    actions: threshold?.actions || [],
    cooldown: threshold?.cooldown || 300
  })
  
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [errors, setErrors] = useState<Record<string, string>>({})
  
  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {}
    
    if (!formData.name.trim()) {
      newErrors.name = 'Name is required'
    }
    
    if (!formData.metric) {
      newErrors.metric = 'Metric is required'
    }
    
    if (formData.value === undefined || formData.value === null) {
      newErrors.value = 'Value is required'
    }
    
    if (formData.cooldown < 0) {
      newErrors.cooldown = 'Cooldown must be non-negative'
    }
    
    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }
  
  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    
    if (validateForm()) {
      onSave(formData)
    }
  }
  
  const addCondition = () => {
    setFormData(prev => ({
      ...prev,
      conditions: [
        ...prev.conditions,
        { field: '', operator: 'eq', value: '' }
      ]
    }))
  }
  
  const updateCondition = (index: number, condition: Partial<ThresholdCondition>) => {
    setFormData(prev => ({
      ...prev,
      conditions: prev.conditions.map((c, i) => 
        i === index ? { ...c, ...condition } : c
      )
    }))
  }
  
  const removeCondition = (index: number) => {
    setFormData(prev => ({
      ...prev,
      conditions: prev.conditions.filter((_, i) => i !== index)
    }))
  }
  
  const addAction = () => {
    setFormData(prev => ({
      ...prev,
      actions: [
        ...prev.actions,
        { id: '', label: '', action: '', confirmRequired: false }
      ]
    }))
  }
  
  const updateAction = (index: number, action: Partial<AlertAction>) => {
    setFormData(prev => ({
      ...prev,
      actions: prev.actions.map((a, i) => 
        i === index ? { ...a, ...action } : a
      )
    }))
  }
  
  const removeAction = (index: number) => {
    setFormData(prev => ({
      ...prev,
      actions: prev.actions.filter((_, i) => i !== index)
    }))
  }
  
  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {/* Name */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Threshold Name *
          </label>
          <input
            type="text"
            value={formData.name}
            onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
            className={`w-full border rounded-md px-3 py-2 text-sm ${
              errors.name ? 'border-red-300' : 'border-gray-300'
            }`}
            placeholder="e.g., High CPU Usage"
          />
          {errors.name && (
            <p className="text-red-600 text-xs mt-1">{errors.name}</p>
          )}
        </div>
        
        {/* Metric */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Metric *
          </label>
          <select
            value={formData.metric}
            onChange={(e) => setFormData(prev => ({ ...prev, metric: e.target.value }))}
            className={`w-full border rounded-md px-3 py-2 text-sm ${
              errors.metric ? 'border-red-300' : 'border-gray-300'
            }`}
          >
            <option value="">Select a metric</option>
            {AVAILABLE_METRICS.map(metric => (
              <option key={metric.value} value={metric.value}>
                {metric.label}
              </option>
            ))}
          </select>
          {errors.metric && (
            <p className="text-red-600 text-xs mt-1">{errors.metric}</p>
          )}
        </div>
        
        {/* Operator */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Condition
          </label>
          <select
            value={formData.operator}
            onChange={(e) => setFormData(prev => ({ ...prev, operator: e.target.value as ThresholdOperator }))}
            className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
          >
            {OPERATORS.map(op => (
              <option key={op.value} value={op.value}>
                {op.label}
              </option>
            ))}
          </select>
        </div>
        
        {/* Value */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Threshold Value *
          </label>
          <input
            type="number"
            step="0.01"
            value={formData.value}
            onChange={(e) => setFormData(prev => ({ ...prev, value: parseFloat(e.target.value) }))}
            className={`w-full border rounded-md px-3 py-2 text-sm ${
              errors.value ? 'border-red-300' : 'border-gray-300'
            }`}
            placeholder="0.00"
          />
          {errors.value && (
            <p className="text-red-600 text-xs mt-1">{errors.value}</p>
          )}
        </div>
        
        {/* Severity */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Severity
          </label>
          <select
            value={formData.severity}
            onChange={(e) => setFormData(prev => ({ ...prev, severity: e.target.value as AlertSeverity }))}
            className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
          >
            {SEVERITIES.map(severity => (
              <option key={severity} value={severity} className="capitalize">
                {severity}
              </option>
            ))}
          </select>
        </div>
        
        {/* Cooldown */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Cooldown (seconds)
          </label>
          <input
            type="number"
            min="0"
            value={formData.cooldown}
            onChange={(e) => setFormData(prev => ({ ...prev, cooldown: parseInt(e.target.value) }))}
            className={`w-full border rounded-md px-3 py-2 text-sm ${
              errors.cooldown ? 'border-red-300' : 'border-gray-300'
            }`}
            placeholder="300"
          />
          {errors.cooldown && (
            <p className="text-red-600 text-xs mt-1">{errors.cooldown}</p>
          )}
        </div>
      </div>
      
      {/* Enabled Toggle */}
      <div className="flex items-center space-x-2">
        <input
          type="checkbox"
          id="enabled"
          checked={formData.enabled}
          onChange={(e) => setFormData(prev => ({ ...prev, enabled: e.target.checked }))}
          className="rounded border-gray-300"
        />
        <label htmlFor="enabled" className="text-sm font-medium text-gray-700">
          Enable this threshold
        </label>
      </div>
      
      {/* Advanced Settings */}
      <div>
        <button
          type="button"
          onClick={() => setShowAdvanced(!showAdvanced)}
          className="text-sm text-blue-600 hover:text-blue-800"
        >
          {showAdvanced ? 'Hide' : 'Show'} Advanced Settings
        </button>
        
        {showAdvanced && (
          <div className="mt-4 space-y-4 p-4 bg-gray-50 rounded-md">
            {/* Additional Conditions */}
            <div>
              <div className="flex items-center justify-between mb-2">
                <label className="block text-sm font-medium text-gray-700">
                  Additional Conditions
                </label>
                <button
                  type="button"
                  onClick={addCondition}
                  className="text-sm text-blue-600 hover:text-blue-800"
                >
                  Add Condition
                </button>
              </div>
              
              {formData.conditions.map((condition, index) => (
                <div key={index} className="grid grid-cols-4 gap-2 mb-2">
                  <input
                    type="text"
                    value={condition.field}
                    onChange={(e) => updateCondition(index, { field: e.target.value })}
                    placeholder="Field"
                    className="border border-gray-300 rounded-md px-2 py-1 text-sm"
                  />
                  <select
                    value={condition.operator}
                    onChange={(e) => updateCondition(index, { operator: e.target.value as ThresholdOperator })}
                    className="border border-gray-300 rounded-md px-2 py-1 text-sm"
                  >
                    {OPERATORS.map(op => (
                      <option key={op.value} value={op.value}>
                        {op.label}
                      </option>
                    ))}
                  </select>
                  <input
                    type="text"
                    value={condition.value}
                    onChange={(e) => updateCondition(index, { value: e.target.value })}
                    placeholder="Value"
                    className="border border-gray-300 rounded-md px-2 py-1 text-sm"
                  />
                  <button
                    type="button"
                    onClick={() => removeCondition(index)}
                    className="text-red-600 hover:text-red-800"
                  >
                    <X className="w-4 h-4" />
                  </button>
                </div>
              ))}
            </div>
            
            {/* Custom Actions */}
            <div>
              <div className="flex items-center justify-between mb-2">
                <label className="block text-sm font-medium text-gray-700">
                  Custom Actions
                </label>
                <button
                  type="button"
                  onClick={addAction}
                  className="text-sm text-blue-600 hover:text-blue-800"
                >
                  Add Action
                </button>
              </div>
              
              {formData.actions.map((action, index) => (
                <div key={index} className="grid grid-cols-4 gap-2 mb-2">
                  <input
                    type="text"
                    value={action.label}
                    onChange={(e) => updateAction(index, { label: e.target.value })}
                    placeholder="Action Label"
                    className="border border-gray-300 rounded-md px-2 py-1 text-sm"
                  />
                  <input
                    type="text"
                    value={action.action}
                    onChange={(e) => updateAction(index, { action: e.target.value })}
                    placeholder="Action ID"
                    className="border border-gray-300 rounded-md px-2 py-1 text-sm"
                  />
                  <label className="flex items-center text-sm">
                    <input
                      type="checkbox"
                      checked={action.confirmRequired || false}
                      onChange={(e) => updateAction(index, { confirmRequired: e.target.checked })}
                      className="mr-1"
                    />
                    Confirm?
                  </label>
                  <button
                    type="button"
                    onClick={() => removeAction(index)}
                    className="text-red-600 hover:text-red-800"
                  >
                    <X className="w-4 h-4" />
                  </button>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
      
      {/* Form Actions */}
      <div className="flex items-center justify-end space-x-3 pt-4 border-t border-gray-200">
        <button
          type="button"
          onClick={onCancel}
          className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
        >
          Cancel
        </button>
        <button
          type="submit"
          className="px-4 py-2 border border-transparent rounded-md text-sm font-medium text-white bg-blue-600 hover:bg-blue-700"
        >
          <Save className="w-4 h-4 mr-2 inline" />
          {threshold ? 'Update' : 'Create'} Threshold
        </button>
      </div>
    </form>
  )
}

const ThresholdCard: React.FC<{
  threshold: AlertThreshold
  onEdit: () => void
  onDelete: () => void
  onToggle: () => void
}> = ({ threshold, onEdit, onDelete, onToggle }) => {
  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString()
  }
  
  const getMetricLabel = (metric: string) => {
    const found = AVAILABLE_METRICS.find(m => m.value === metric)
    return found ? found.label : metric
  }
  
  const getOperatorLabel = (operator: ThresholdOperator) => {
    const found = OPERATORS.find(op => op.value === operator)
    return found ? found.label : operator
  }
  
  return (
    <div className="bg-white border border-gray-200 rounded-lg p-4">
      <div className="flex items-start justify-between mb-3">
        <div className="flex-1">
          <div className="flex items-center space-x-2 mb-1">
            <h3 className="text-lg font-medium text-gray-900">{threshold.name}</h3>
            <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${SEVERITY_COLORS[threshold.severity]}`}>
              {threshold.severity}
            </span>
            {threshold.enabled ? (
              <Bell className="w-4 h-4 text-green-500" title="Enabled" />
            ) : (
              <BellOff className="w-4 h-4 text-gray-400" title="Disabled" />
            )}
          </div>
          
          <div className="text-sm text-gray-600 space-y-1">
            <div>
              <strong>Metric:</strong> {getMetricLabel(threshold.metric)}
            </div>
            <div>
              <strong>Condition:</strong> {getOperatorLabel(threshold.operator)} {threshold.value}
            </div>
            <div className="flex items-center space-x-4">
              <span>
                <Clock className="w-4 h-4 inline mr-1" />
                <strong>Cooldown:</strong> {threshold.cooldown}s
              </span>
              <span>
                <strong>Created:</strong> {formatDate(threshold.createdAt)}
              </span>
            </div>
          </div>
          
          {threshold.conditions.length > 0 && (
            <div className="mt-2 text-xs text-gray-500">
              +{threshold.conditions.length} additional condition{threshold.conditions.length > 1 ? 's' : ''}
            </div>
          )}
          
          {threshold.actions.length > 0 && (
            <div className="mt-2 text-xs text-gray-500">
              {threshold.actions.length} custom action{threshold.actions.length > 1 ? 's' : ''}
            </div>
          )}
        </div>
        
        <div className="flex items-center space-x-1 ml-4">
          <button
            onClick={onToggle}
            className={`p-2 rounded-md transition-colors ${
              threshold.enabled
                ? 'text-green-600 hover:bg-green-50'
                : 'text-gray-400 hover:bg-gray-50'
            }`}
            title={threshold.enabled ? 'Disable' : 'Enable'}
          >
            {threshold.enabled ? <Bell className="w-4 h-4" /> : <BellOff className="w-4 h-4" />}
          </button>
          
          <button
            onClick={onEdit}
            className="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-50 rounded-md transition-colors"
            title="Edit"
          >
            <Edit className="w-4 h-4" />
          </button>
          
          <button
            onClick={onDelete}
            className="p-2 text-red-400 hover:text-red-600 hover:bg-red-50 rounded-md transition-colors"
            title="Delete"
          >
            <Trash2 className="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>
  )
}

export const ThresholdConfig: React.FC<ThresholdConfigProps> = ({
  thresholds,
  onCreateThreshold,
  onUpdateThreshold,
  onDeleteThreshold,
  className = ''
}) => {
  const [showForm, setShowForm] = useState(false)
  const [editingThreshold, setEditingThreshold] = useState<AlertThreshold | null>(null)
  
  const handleSave = (data: ThresholdFormData) => {
    if (editingThreshold) {
      onUpdateThreshold(editingThreshold.id, data)
    } else {
      onCreateThreshold(data)
    }
    
    setShowForm(false)
    setEditingThreshold(null)
  }
  
  const handleEdit = (threshold: AlertThreshold) => {
    setEditingThreshold(threshold)
    setShowForm(true)
  }
  
  const handleDelete = (threshold: AlertThreshold) => {
    if (confirm(`Are you sure you want to delete the threshold "${threshold.name}"?`)) {
      onDeleteThreshold(threshold.id)
    }
  }
  
  const handleToggle = (threshold: AlertThreshold) => {
    onUpdateThreshold(threshold.id, { enabled: !threshold.enabled })
  }
  
  const handleCancel = () => {
    setShowForm(false)
    setEditingThreshold(null)
  }
  
  return (
    <div className={`space-y-6 ${className}`}>
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold text-gray-900">Alert Thresholds</h2>
        {!showForm && (
          <button
            onClick={() => setShowForm(true)}
            className="inline-flex items-center px-4 py-2 border border-transparent rounded-md text-sm font-medium text-white bg-blue-600 hover:bg-blue-700"
          >
            <Plus className="w-4 h-4 mr-2" />
            New Threshold
          </button>
        )}
      </div>
      
      {showForm && (
        <div className="bg-white border border-gray-200 rounded-lg p-6">
          <h3 className="text-lg font-medium text-gray-900 mb-4">
            {editingThreshold ? 'Edit' : 'Create'} Threshold
          </h3>
          <ThresholdForm
            threshold={editingThreshold || undefined}
            onSave={handleSave}
            onCancel={handleCancel}
          />
        </div>
      )}
      
      <div className="space-y-4">
        {thresholds.length === 0 ? (
          <div className="text-center py-8 bg-white border border-gray-200 rounded-lg">
            <Settings className="w-12 h-12 text-gray-400 mx-auto mb-4" />
            <div className="text-gray-500">No thresholds configured</div>
            <div className="text-gray-400 text-sm">Create a threshold to get started with alerts</div>
          </div>
        ) : (
          thresholds.map(threshold => (
            <ThresholdCard
              key={threshold.id}
              threshold={threshold}
              onEdit={() => handleEdit(threshold)}
              onDelete={() => handleDelete(threshold)}
              onToggle={() => handleToggle(threshold)}
            />
          ))
        )}
      </div>
    </div>
  )
}