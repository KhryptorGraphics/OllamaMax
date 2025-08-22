/**
 * WebSocket event handlers for real-time updates
 * Handles incoming messages and updates the global store appropriately
 */

import { useStore } from '@/stores'
import type {
  NodeStatusUpdate,
  ModelSyncUpdate,
  TaskUpdate,
  TransferProgressUpdate,
  ClusterMetricsUpdate,
  PerformanceAlert,
  SecurityAlert,
  SystemNotification,
  UserNotification,
} from '@/types/websocket'

// Configuration for event handlers
interface EventHandlerConfig {
  enableDebugLogging: boolean
  enableOptimisticUpdates: boolean
  enableNotifications: boolean
  enableMetrics: boolean
  batchUpdates: boolean
  updateThrottleMs: number
}

/**
 * Real-time event handlers with optimistic updates and smart batching
 */
export class RealTimeEventHandlers {
  private config: EventHandlerConfig
  private updateTimers = new Map<string, NodeJS.Timeout>()
  private pendingUpdates = new Map<string, () => void>()

  constructor(config: Partial<EventHandlerConfig> = {}) {
    this.config = {
      enableDebugLogging: false,
      enableOptimisticUpdates: true,
      enableNotifications: true,
      enableMetrics: true,
      batchUpdates: true,
      updateThrottleMs: 500,
      ...config,
    }
  }

  /**
   * Handle node status updates with optimistic updates and batching
   */
  public handleNodeStatusUpdate = (data: NodeStatusUpdate): void => {
    const store = useStore.getState()

    try {
      // Apply immediate update for better UX
      if (this.config.enableOptimisticUpdates) {
        store.updateNodeStatus(data.nodeId, {
          status: data.status,
          health: data.health,
          lastSeen: data.timestamp,
        })
      }

      // Batch additional updates if enabled
      if (this.config.batchUpdates) {
        this.batchUpdate(`node:${data.nodeId}`, () => {
          this.processNodeUpdate(data)
        })
      } else {
        this.processNodeUpdate(data)
      }

      // Update cluster health metrics
      this.updateClusterHealth()

      // Log for debugging
      this.log('Node status updated:', data.nodeId, data.status)

    } catch (error) {
      this.log('Error handling node status update:', error)
      this.showErrorNotification('Failed to update node status', error)
    }
  }

  /**
   * Handle model sync updates with progress tracking
   */
  public handleModelSyncUpdate = (data: ModelSyncUpdate): void => {
    const store = useStore.getState()

    try {
      // Update model sync status
      store.updateModelSync(data.modelName, {
        status: data.status,
        progress: data.progress,
        nodesTotal: 1,
        nodesSynced: data.status === 'completed' ? 1 : 0,
        error: data.error,
      })

      // Show notifications for important status changes
      if (this.config.enableNotifications) {
        this.handleModelSyncNotifications(data)
      }

      // Update metrics if enabled
      if (this.config.enableMetrics) {
        this.updateModelMetrics(data)
      }

      this.log('Model sync updated:', data.modelName, data.status, `${data.progress}%`)

    } catch (error) {
      this.log('Error handling model sync update:', error)
      this.showErrorNotification('Failed to update model sync', error)
    }
  }

  /**
   * Handle task updates with progress tracking and notifications
   */
  public handleTaskUpdate = (data: TaskUpdate): void => {
    const store = useStore.getState()

    try {
      // Update task status
      store.updateTaskStatus(data.taskId, data)

      // Handle task completion/failure notifications
      if (this.config.enableNotifications) {
        this.handleTaskNotifications(data)
      }

      this.log('Task updated:', data.taskId, data.status)

    } catch (error) {
      this.log('Error handling task update:', error)
      this.showErrorNotification('Failed to update task', error)
    }
  }

  /**
   * Handle transfer progress with real-time updates
   */
  public handleTransferProgress = (data: TransferProgressUpdate): void => {
    const store = useStore.getState()

    try {
      // Update transfer progress
      store.updateTransferProgress(data.transferId, data.progress)

      // Show completion notification
      if (data.status === 'completed' && this.config.enableNotifications) {
        store.addNotification({
          type: 'success',
          title: 'Transfer Complete',
          message: `Transfer ${data.transferId} completed successfully`,
          duration: 5000,
        })
      }

      this.log('Transfer progress:', data.transferId, `${data.progress.percentage}%`)

    } catch (error) {
      this.log('Error handling transfer progress:', error)
      this.showErrorNotification('Failed to update transfer progress', error)
    }
  }

  /**
   * Handle cluster metrics with real-time performance monitoring
   */
  public handleClusterMetrics = (data: ClusterMetricsUpdate): void => {
    const store = useStore.getState()

    try {
      // Update cluster metrics
      store.updateMetrics(data)

      // Check for performance thresholds
      if (this.config.enableMetrics) {
        this.checkPerformanceThresholds(data)
      }

      this.log('Cluster metrics updated')

    } catch (error) {
      this.log('Error handling cluster metrics:', error)
      this.showErrorNotification('Failed to update cluster metrics', error)
    }
  }

  /**
   * Handle performance alerts with automatic severity-based actions
   */
  public handlePerformanceAlert = (data: PerformanceAlert): void => {
    const store = useStore.getState()

    try {
      // Add alert to store
      store.addAlert(data)

      // Show notification based on severity
      if (this.config.enableNotifications) {
        const notificationType = this.getNotificationTypeForSeverity(data.severity)
        const duration = data.severity === 'critical' ? 0 : 10000

        store.addNotification({
          type: notificationType,
          title: 'Performance Alert',
          message: data.message,
          duration,
        })
      }

      this.log('Performance alert:', data.type, data.severity)

    } catch (error) {
      this.log('Error handling performance alert:', error)
      this.showErrorNotification('Failed to process performance alert', error)
    }
  }

  /**
   * Handle security alerts with immediate high-priority notifications
   */
  public handleSecurityAlert = (data: SecurityAlert): void => {
    const store = useStore.getState()

    try {
      // Add alert to store
      store.addAlert(data)

      // Always show notification for security alerts
      if (this.config.enableNotifications) {
        store.addNotification({
          type: 'error',
          title: 'Security Alert',
          message: data.message,
          duration: 0, // Never auto-dismiss security alerts
        })
      }

      // Log security event
      console.warn('[SECURITY ALERT]', data.type, data.severity, data.message)

      this.log('Security alert:', data.type, data.severity)

    } catch (error) {
      this.log('Error handling security alert:', error)
      this.showErrorNotification('Failed to process security alert', error)
    }
  }

  /**
   * Handle system notifications with appropriate routing
   */
  public handleSystemNotification = (data: SystemNotification): void => {
    const store = useStore.getState()

    try {
      const notificationType = this.getNotificationTypeForSeverity(data.severity)
      const duration = data.severity === 'error' ? 0 : 10000

      store.addNotification({
        type: notificationType,
        title: data.title,
        message: data.message,
        duration,
      })

      this.log('System notification:', data.type, data.severity)

    } catch (error) {
      this.log('Error handling system notification:', error)
      this.showErrorNotification('Failed to process system notification', error)
    }
  }

  /**
   * Handle user-specific notifications
   */
  public handleUserNotification = (data: UserNotification): void => {
    const store = useStore.getState()

    try {
      store.addNotification({
        type: 'info',
        title: data.title,
        message: data.message,
        duration: 15000,
      })

      this.log('User notification:', data.type)

    } catch (error) {
      this.log('Error handling user notification:', error)
      this.showErrorNotification('Failed to process user notification', error)
    }
  }

  /**
   * Private helper methods
   */
  private processNodeUpdate(data: NodeStatusUpdate): void {
    // Additional processing for node updates
    if (data.status === 'offline') {
      // Check if this affects any running tasks
      this.checkAffectedTasks(data.nodeId)
    }
  }

  private handleModelSyncNotifications(data: ModelSyncUpdate): void {
    const store = useStore.getState()

    if (data.status === 'completed') {
      store.addNotification({
        type: 'success',
        title: 'Model Sync Complete',
        message: `${data.modelName} has been successfully synchronized`,
        duration: 5000,
      })
    } else if (data.status === 'failed') {
      store.addNotification({
        type: 'error',
        title: 'Model Sync Failed',
        message: `Failed to sync ${data.modelName}: ${data.error}`,
        duration: 10000,
      })
    }
  }

  private handleTaskNotifications(data: TaskUpdate): void {
    const store = useStore.getState()

    if (data.status === 'completed') {
      store.addNotification({
        type: 'success',
        title: 'Task Completed',
        message: `Task ${data.taskId} completed successfully`,
        duration: 5000,
      })
    } else if (data.status === 'failed') {
      store.addNotification({
        type: 'error',
        title: 'Task Failed',
        message: `Task ${data.taskId} failed: ${data.error}`,
        duration: 10000,
      })
    }
  }

  private updateClusterHealth(): void {
    // Implementation stub
    this.log('Updating cluster health')
  }

  private updateModelMetrics(data: ModelSyncUpdate): void {
    // Track model sync performance metrics
    this.log('Model sync metrics:', data.modelName)
  }

  private checkPerformanceThresholds(data: ClusterMetricsUpdate): void {
    // Check various performance thresholds
    this.log('Checking performance thresholds')
  }

  private checkAffectedTasks(nodeId: string): void {
    // Implementation to check and handle tasks affected by node going offline
    this.log('Checking tasks affected by node offline:', nodeId)
  }

  private batchUpdate(key: string, updateFn: () => void): void {
    // Clear existing timer for this key
    if (this.updateTimers.has(key)) {
      clearTimeout(this.updateTimers.get(key)!)
    }

    // Set new timer
    const timer = setTimeout(() => {
      updateFn()
      this.updateTimers.delete(key)
    }, this.config.updateThrottleMs)

    this.updateTimers.set(key, timer)
  }

  private getNotificationTypeForSeverity(severity: string): 'success' | 'info' | 'warning' | 'error' {
    switch (severity) {
      case 'critical':
      case 'error':
        return 'error'
      case 'high':
      case 'warning':
        return 'warning'
      case 'medium':
      case 'info':
        return 'info'
      default:
        return 'info'
    }
  }

  private showErrorNotification(title: string, error: any): void {
    if (this.config.enableNotifications) {
      const store = useStore.getState()
      store.addNotification({
        type: 'error',
        title,
        message: error.message || 'An unexpected error occurred',
        duration: 10000,
      })
    }
  }

  private log(message: string, ...args: any[]): void {
    if (this.config.enableDebugLogging) {
      console.log(`[EventHandlers] ${message}`, ...args)
    }
  }

  /**
   * Cleanup method to clear timers
   */
  public cleanup(): void {
    this.updateTimers.forEach(timer => clearTimeout(timer))
    this.updateTimers.clear()
    this.pendingUpdates.clear()
  }
}

// Create singleton instance
export const eventHandlers = new RealTimeEventHandlers({
  enableDebugLogging: import.meta.env.DEV,
  enableOptimisticUpdates: true,
  enableNotifications: true,
  enableMetrics: true,
  batchUpdates: true,
})

export default eventHandlers