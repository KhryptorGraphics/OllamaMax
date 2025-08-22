import React, { useState } from 'react'
import { Card, CardHeader, CardContent, CardTitle } from '@/design-system/components/Card/Card'
import type { Node } from '../NodesPage'

interface NodeActionsProps {
  node: Node
  onAction: (action: string, params?: any) => void
}

interface MaintenanceForm {
  reason: string
  startTime: string
  endTime: string
  drainFirst: boolean
}

export const NodeActions: React.FC<NodeActionsProps> = ({ node, onAction }) => {
  const [showMaintenanceForm, setShowMaintenanceForm] = useState(false)
  const [maintenanceForm, setMaintenanceForm] = useState<MaintenanceForm>({
    reason: '',
    startTime: '',
    endTime: '',
    drainFirst: true
  })
  const [isExecuting, setIsExecuting] = useState<string | null>(null)
  const [showTerminal, setShowTerminal] = useState(false)

  const handleAction = async (action: string, params?: any) => {
    setIsExecuting(action)
    try {
      await onAction(action, params)
      
      // Reset maintenance form if successfully scheduled
      if (action === 'maintenance') {
        setShowMaintenanceForm(false)
        setMaintenanceForm({
          reason: '',
          startTime: '',
          endTime: '',
          drainFirst: true
        })
      }
    } catch (error) {
      console.error('Action failed:', error)
    } finally {
      setIsExecuting(null)
    }
  }

  const handleMaintenanceSubmit = () => {
    if (!maintenanceForm.reason || !maintenanceForm.startTime || !maintenanceForm.endTime) {
      return
    }

    handleAction('maintenance', {
      scheduled: true,
      start_time: new Date(maintenanceForm.startTime).toISOString(),
      end_time: new Date(maintenanceForm.endTime).toISOString(),
      reason: maintenanceForm.reason,
      drain_first: maintenanceForm.drainFirst
    })
  }

  const isActionDisabled = (action: string) => {
    if (isExecuting) return true
    
    switch (action) {
      case 'start':
        return node.status !== 'offline'
      case 'stop':
      case 'restart':
        return node.status === 'offline'
      case 'drain':
        return node.status !== 'online'
      case 'maintenance':
        return node.status === 'offline'
      default:
        return false
    }
  }

  const getActionButtonClass = (action: string, variant: 'primary' | 'secondary' | 'warning' | 'error' = 'primary') => {
    const baseClasses = 'w-full px-4 py-2 rounded-lg text-sm font-medium transition-all duration-200 flex items-center justify-center gap-2'
    const disabled = isActionDisabled(action)
    const executing = isExecuting === action

    if (disabled) {
      return `${baseClasses} bg-muted text-muted-foreground cursor-not-allowed`
    }

    if (executing) {
      return `${baseClasses} bg-muted text-muted-foreground cursor-wait`
    }

    const variants = {
      primary: 'bg-primary text-primary-foreground hover:bg-primary/90',
      secondary: 'bg-secondary text-secondary-foreground hover:bg-secondary/90',
      warning: 'bg-warning text-warning-foreground hover:bg-warning/90',
      error: 'bg-error text-error-foreground hover:bg-error/90'
    }

    return `${baseClasses} ${variants[variant]}`
  }

  const ActionIcon: React.FC<{ action: string }> = ({ action }) => {
    if (isExecuting === action) {
      return (
        <svg className="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"/>
          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/>
        </svg>
      )
    }

    switch (action) {
      case 'start':
        return (
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M14.828 14.828a4 4 0 01-5.656 0M9 10h1m4 0h1m-6 4h8m-9 6h10a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
        )
      case 'stop':
        return (
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 10h6v4H9z" />
          </svg>
        )
      case 'restart':
        return (
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
        )
      case 'drain':
        return (
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 17h8m0 0V9m0 8l-8-8-4 4-6-6" />
          </svg>
        )
      case 'maintenance':
        return (
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
        )
      default:
        return null
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <svg className="w-5 h-5 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 100 4m0-4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 100 4m0-4v2m0-6V4" />
          </svg>
          Node Actions
        </CardTitle>
      </CardHeader>

      <CardContent spacing="md">
        <div className="space-y-3">
          {/* Primary Actions */}
          <div className="space-y-2">
            <h4 className="text-sm font-medium text-foreground">Primary Actions</h4>
            
            {node.status === 'offline' && (
              <button
                onClick={() => handleAction('start')}
                disabled={isActionDisabled('start')}
                className={getActionButtonClass('start', 'primary')}
              >
                <ActionIcon action="start" />
                Start Node
              </button>
            )}

            {node.status === 'online' && (
              <>
                <button
                  onClick={() => handleAction('restart')}
                  disabled={isActionDisabled('restart')}
                  className={getActionButtonClass('restart', 'warning')}
                >
                  <ActionIcon action="restart" />
                  Restart Node
                </button>
                
                <button
                  onClick={() => handleAction('drain')}
                  disabled={isActionDisabled('drain')}
                  className={getActionButtonClass('drain', 'warning')}
                >
                  <ActionIcon action="drain" />
                  Drain Node
                </button>
              </>
            )}

            {(node.status === 'online' || node.status === 'draining') && (
              <button
                onClick={() => handleAction('stop')}
                disabled={isActionDisabled('stop')}
                className={getActionButtonClass('stop', 'error')}
              >
                <ActionIcon action="stop" />
                Stop Node
              </button>
            )}
          </div>

          {/* Maintenance Actions */}
          <div className="pt-3 border-t border-border">
            <h4 className="text-sm font-medium text-foreground mb-2">Maintenance</h4>
            
            {!showMaintenanceForm ? (
              <button
                onClick={() => setShowMaintenanceForm(true)}
                disabled={isActionDisabled('maintenance')}
                className={getActionButtonClass('maintenance', 'secondary')}
              >
                <ActionIcon action="maintenance" />
                Schedule Maintenance
              </button>
            ) : (
              <div className="space-y-3 p-3 border border-border rounded-lg bg-muted/50">
                <div>
                  <label className="text-xs text-muted-foreground block mb-1">Reason</label>
                  <input
                    type="text"
                    value={maintenanceForm.reason}
                    onChange={(e) => setMaintenanceForm(prev => ({ ...prev, reason: e.target.value }))}
                    placeholder="e.g., OS security updates"
                    className="w-full px-3 py-2 text-sm border border-border rounded bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                  />
                </div>
                
                <div className="grid grid-cols-2 gap-2">
                  <div>
                    <label className="text-xs text-muted-foreground block mb-1">Start Time</label>
                    <input
                      type="datetime-local"
                      value={maintenanceForm.startTime}
                      onChange={(e) => setMaintenanceForm(prev => ({ ...prev, startTime: e.target.value }))}
                      className="w-full px-3 py-2 text-sm border border-border rounded bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                    />
                  </div>
                  
                  <div>
                    <label className="text-xs text-muted-foreground block mb-1">End Time</label>
                    <input
                      type="datetime-local"
                      value={maintenanceForm.endTime}
                      onChange={(e) => setMaintenanceForm(prev => ({ ...prev, endTime: e.target.value }))}
                      className="w-full px-3 py-2 text-sm border border-border rounded bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                    />
                  </div>
                </div>
                
                <div className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    id="drainFirst"
                    checked={maintenanceForm.drainFirst}
                    onChange={(e) => setMaintenanceForm(prev => ({ ...prev, drainFirst: e.target.checked }))}
                    className="w-4 h-4 text-primary border-border rounded focus:ring-primary"
                  />
                  <label htmlFor="drainFirst" className="text-xs text-muted-foreground">
                    Drain requests before maintenance
                  </label>
                </div>
                
                <div className="flex gap-2">
                  <button
                    onClick={handleMaintenanceSubmit}
                    disabled={!maintenanceForm.reason || !maintenanceForm.startTime || !maintenanceForm.endTime}
                    className="flex-1 px-3 py-2 bg-primary text-primary-foreground rounded text-xs hover:bg-primary/90 transition-colors disabled:bg-muted disabled:text-muted-foreground"
                  >
                    Schedule
                  </button>
                  <button
                    onClick={() => setShowMaintenanceForm(false)}
                    className="flex-1 px-3 py-2 bg-muted text-muted-foreground rounded text-xs hover:bg-muted/80 transition-colors"
                  >
                    Cancel
                  </button>
                </div>
              </div>
            )}

            {/* Show current maintenance if scheduled */}
            {node.maintenance?.scheduled && (
              <div className="mt-3 p-3 border border-info/20 bg-info/10 rounded-lg">
                <h5 className="text-sm font-medium text-info mb-2">Scheduled Maintenance</h5>
                <div className="text-xs text-muted-foreground space-y-1">
                  <p><strong>Reason:</strong> {node.maintenance.reason}</p>
                  <p><strong>Start:</strong> {node.maintenance.start_time ? new Date(node.maintenance.start_time).toLocaleString() : 'Not set'}</p>
                  <p><strong>End:</strong> {node.maintenance.end_time ? new Date(node.maintenance.end_time).toLocaleString() : 'Not set'}</p>
                </div>
                <button
                  onClick={() => handleAction('cancel-maintenance')}
                  className="mt-2 px-3 py-1 bg-error text-error-foreground rounded text-xs hover:bg-error/90 transition-colors"
                >
                  Cancel Maintenance
                </button>
              </div>
            )}
          </div>

          {/* Advanced Actions */}
          <div className="pt-3 border-t border-border">
            <h4 className="text-sm font-medium text-foreground mb-2">Advanced</h4>
            
            <div className="space-y-2">
              <button
                onClick={() => setShowTerminal(!showTerminal)}
                className="w-full px-4 py-2 bg-muted text-muted-foreground rounded-lg text-sm hover:bg-muted/80 transition-colors flex items-center justify-center gap-2"
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v14a2 2 0 002 2z" />
                </svg>
                SSH Terminal
              </button>
              
              <button
                onClick={() => handleAction('update')}
                disabled={node.status === 'offline'}
                className={getActionButtonClass('update', 'secondary')}
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M9 19l3 3m0 0l3-3m-3 3V10" />
                </svg>
                Update Node
              </button>
              
              <button
                onClick={() => handleAction('logs')}
                className="w-full px-4 py-2 bg-muted text-muted-foreground rounded-lg text-sm hover:bg-muted/80 transition-colors flex items-center justify-center gap-2"
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                View Logs
              </button>
            </div>
          </div>

          {/* Danger Zone */}
          <div className="pt-3 border-t border-error/20">
            <h4 className="text-sm font-medium text-error mb-2">Danger Zone</h4>
            <button
              onClick={() => handleAction('remove')}
              disabled={node.status !== 'offline'}
              className="w-full px-4 py-2 bg-error/10 text-error border border-error/20 rounded-lg text-sm hover:bg-error/20 transition-colors flex items-center justify-center gap-2 disabled:bg-muted disabled:text-muted-foreground disabled:border-muted"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
              Remove Node
            </button>
            <p className="text-xs text-muted-foreground mt-1 text-center">
              Node must be offline to remove
            </p>
          </div>
        </div>

        {/* Terminal Modal */}
        {showTerminal && (
          <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
            <div className="bg-background border border-border rounded-lg w-full max-w-4xl h-3/4 flex flex-col">
              <div className="flex items-center justify-between p-4 border-b border-border">
                <h3 className="text-lg font-medium text-foreground">SSH Terminal - {node.name}</h3>
                <button
                  onClick={() => setShowTerminal(false)}
                  className="p-1 hover:bg-muted rounded"
                >
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
              <div className="flex-1 p-4 bg-black text-green-400 font-mono text-sm overflow-auto">
                <div className="mb-2">$ ssh ollama@{node.ip_address}</div>
                <div className="mb-2">Welcome to {node.hostname}</div>
                <div className="mb-2">Last login: {new Date().toLocaleString()}</div>
                <div className="flex items-center">
                  <span className="text-green-400">ollama@{node.hostname}:~$ </span>
                  <span className="bg-green-400 w-2 h-4 animate-pulse ml-1"></span>
                </div>
              </div>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

export default NodeActions