/**
 * WebSocket connection status indicator component
 * Shows real-time connection state with visual feedback
 */

import { useWebSocket } from '@/services/websocket'

interface WebSocketStatusProps {
  showDetails?: boolean
  className?: string
}

export function WebSocketStatus({ showDetails = false, className = '' }: WebSocketStatusProps) {
  const { isConnected, connectionState, error } = useWebSocket()
  
  // Get status color and icon based on connection state
  const getStatusInfo = () => {
    switch (connectionState) {
      case 'connected':
        return {
          color: 'text-green-600',
          bgColor: 'bg-green-100',
          borderColor: 'border-green-300',
          icon: '🟢',
          text: 'Connected',
          description: 'Real-time updates active'
        }
      case 'connecting':
        return {
          color: 'text-yellow-600',
          bgColor: 'bg-yellow-100',
          borderColor: 'border-yellow-300',
          icon: '🟡',
          text: 'Connecting',
          description: 'Establishing connection...'
        }
      case 'reconnecting':
        return {
          color: 'text-orange-600',
          bgColor: 'bg-orange-100',
          borderColor: 'border-orange-300',
          icon: '🔄',
          text: 'Reconnecting',
          description: 'Attempting to reconnect...'
        }
      case 'error':
        return {
          color: 'text-red-600',
          bgColor: 'bg-red-100',
          borderColor: 'border-red-300',
          icon: '🔴',
          text: 'Error',
          description: error?.message || 'Connection error'
        }
      case 'disconnected':
      default:
        return {
          color: 'text-gray-600',
          bgColor: 'bg-gray-100',
          borderColor: 'border-gray-300',
          icon: '⚫',
          text: 'Disconnected',
          description: 'No real-time updates'
        }
    }
  }
  
  const statusInfo = getStatusInfo()
  
  // Get status information
  
  // Simple status indicator
  if (!showDetails) {
    return (
      <div className={`inline-flex items-center gap-2 ${className}`}>
        <span 
          className="text-sm"
          title={`WebSocket: ${statusInfo.text} - ${statusInfo.description}`}
        >
          {statusInfo.icon}
        </span>
        <span className={`text-xs font-medium ${statusInfo.color}`}>
          {statusInfo.text}
        </span>
      </div>
    )
  }
  
  // Detailed status panel
  return (
    <div className={`rounded-lg border p-4 ${statusInfo.bgColor} ${statusInfo.borderColor} ${className}`}>
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <span className="text-lg">{statusInfo.icon}</span>
          <div>
            <h3 className={`font-semibold ${statusInfo.color}`}>
              WebSocket Connection
            </h3>
            <p className="text-sm text-gray-600">
              {statusInfo.description}
            </p>
          </div>
        </div>
        
        {/* Connection pulse animation for active state */}
        {isConnected && (
          <div className="relative">
            <div className="w-3 h-3 bg-green-500 rounded-full animate-pulse"></div>
            <div className="absolute inset-0 w-3 h-3 bg-green-500 rounded-full animate-ping opacity-75"></div>
          </div>
        )}
      </div>
      
      {/* Connection details */}
      {showDetails && (
        <div className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <span className="font-medium text-gray-700">State:</span>
            <span className={`ml-2 ${statusInfo.color}`}>
              {statusInfo.text}
            </span>
          </div>
          
          {isConnected && (
            <div>
              <span className="font-medium text-gray-700">Status:</span>
              <span className="ml-2 text-green-600">
                Real-time updates active
              </span>
            </div>
          )}
          
          {error && (
            <div className="col-span-2">
              <span className="font-medium text-gray-700">Error:</span>
              <span className="ml-2 text-red-600 text-xs">
                {error.message}
              </span>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

/**
 * Compact status badge for navigation bars
 */
export function WebSocketStatusBadge({ className = '' }: { className?: string }) {
  const { isConnected, connectionState } = useWebSocket()
  
  const getStatusColor = () => {
    switch (connectionState) {
      case 'connected':
        return 'bg-green-500'
      case 'connecting':
      case 'reconnecting':
        return 'bg-yellow-500'
      case 'error':
        return 'bg-red-500'
      default:
        return 'bg-gray-400'
    }
  }
  
  return (
    <div className={`relative inline-block ${className}`}>
      <div 
        className={`w-2 h-2 rounded-full ${getStatusColor()}`}
        title={`WebSocket: ${connectionState}`}
      >
        {isConnected && (
          <div className="absolute inset-0 w-2 h-2 bg-green-500 rounded-full animate-ping opacity-75"></div>
        )}
      </div>
    </div>
  )
}

/**
 * WebSocket status for headers/navigation
 */
export function WebSocketStatusHeader({ className = '' }: { className?: string }) {
  const { isConnected, connectionState, error } = useWebSocket()
  
  const getStatusText = () => {
    switch (connectionState) {
      case 'connected':
        return 'Live'
      case 'connecting':
        return 'Connecting...'
      case 'reconnecting':
        return 'Reconnecting...'
      case 'error':
        return 'Offline'
      default:
        return 'Disconnected'
    }
  }
  
  const getStatusClass = () => {
    switch (connectionState) {
      case 'connected':
        return 'text-green-600 bg-green-50'
      case 'connecting':
      case 'reconnecting':
        return 'text-yellow-600 bg-yellow-50'
      case 'error':
        return 'text-red-600 bg-red-50'
      default:
        return 'text-gray-600 bg-gray-50'
    }
  }
  
  return (
    <div className={`inline-flex items-center gap-2 px-2 py-1 rounded-full text-xs font-medium ${getStatusClass()} ${className}`}>
      <WebSocketStatusBadge />
      <span>{getStatusText()}</span>
    </div>
  )
}

export default WebSocketStatus