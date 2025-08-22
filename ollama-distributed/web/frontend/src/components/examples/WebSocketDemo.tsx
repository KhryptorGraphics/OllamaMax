/**
 * WebSocket Integration Demo Component
 * Demonstrates the comprehensive WebSocket functionality including:
 * - Connection management with health monitoring
 * - Real-time subscriptions by topic
 * - Message sending with retry logic  
 * - Connection metrics and status
 * - Error handling and recovery
 */

import React, { useState, useEffect } from 'react'
import { Card, Button, Badge } from '@/design-system/components'
import { webSocketService } from '@/services/websocket/service'

interface DemoMessage {
  type: string
  payload: any
  timestamp: string
}

interface ConnectionMetrics {
  totalConnections: number
  messagesSent: number
  messagesReceived: number
  avgLatency: number
  totalErrors: number
  uptime: number
}

interface HealthCheck {
  status: 'healthy' | 'degraded' | 'unhealthy'
  latency: number
  lastCheck: string
}

export interface WebSocketDemoProps {
  className?: string
  showMetrics?: boolean
  autoConnect?: boolean
}

const WebSocketDemo: React.FC<WebSocketDemoProps> = ({
  className = '',
  showMetrics = true,
  autoConnect = true
}) => {
  const [connectionState, setConnectionState] = useState<'connected' | 'disconnected' | 'connecting'>('disconnected')
  const [isConnecting, setIsConnecting] = useState(false)
  const [receivedMessages, setReceivedMessages] = useState<DemoMessage[]>([])
  const [health, setHealth] = useState<HealthCheck | null>(null)
  const [metrics, setMetrics] = useState<ConnectionMetrics | null>(null)
  const [lastError, setLastError] = useState<string | null>(null)
  const [subscriptions, setSubscriptions] = useState<string[]>([])

  // Available demo topics
  const demoTopics = [
    'dashboard',
    'models', 
    'nodes',
    'monitoring',
    'notifications'
  ]

  useEffect(() => {
    if (autoConnect) {
      handleConnect()
    }

    // Cleanup on unmount
    return () => {
      webSocketService.disconnect()
    }
  }, [autoConnect])

  useEffect(() => {
    // Update connection state based on service status
    const updateConnectionState = () => {
      const status = webSocketService.getStatus()
      if (status.connected) {
        setConnectionState('connected')
        setIsConnecting(false)
      } else if (status.connecting) {
        setConnectionState('connecting')
        setIsConnecting(true)
      } else {
        setConnectionState('disconnected')
        setIsConnecting(false)
      }
    }

    const interval = setInterval(updateConnectionState, 1000)
    return () => clearInterval(interval)
  }, [])

  const handleConnect = async () => {
    try {
      setIsConnecting(true)
      setLastError(null)
      await webSocketService.connect()
      setConnectionState('connected')
      
      // Simulate health check
      setHealth({
        status: 'healthy',
        latency: Math.floor(Math.random() * 50) + 10,
        lastCheck: new Date().toISOString()
      })

      // Simulate metrics
      setMetrics({
        totalConnections: 1,
        messagesSent: 0,
        messagesReceived: 0,
        avgLatency: 25,
        totalErrors: 0,
        uptime: 0
      })
    } catch (error) {
      setLastError(error instanceof Error ? error.message : 'Connection failed')
      setConnectionState('disconnected')
    } finally {
      setIsConnecting(false)
    }
  }

  const handleDisconnect = () => {
    webSocketService.disconnect()
    setConnectionState('disconnected')
    setSubscriptions([])
    setHealth(null)
    setMetrics(null)
  }

  const handleSubscribe = async (topic: string) => {
    try {
      await webSocketService.subscribe([topic])
      setSubscriptions(prev => [...prev, topic])
      
      // Add a demo message for this subscription
      const demoMessage: DemoMessage = {
        type: `${topic}_update`,
        payload: { 
          status: 'subscribed',
          topic,
          timestamp: Date.now()
        },
        timestamp: new Date().toISOString()
      }
      
      setReceivedMessages(prev => [demoMessage, ...prev.slice(0, 9)])
      
      // Update metrics
      setMetrics(prev => prev ? {
        ...prev,
        messagesReceived: prev.messagesReceived + 1
      } : null)
    } catch (error) {
      setLastError(error instanceof Error ? error.message : 'Subscription failed')
    }
  }

  const handleUnsubscribe = async (topic: string) => {
    try {
      await webSocketService.unsubscribe([topic])
      setSubscriptions(prev => prev.filter(t => t !== topic))
    } catch (error) {
      setLastError(error instanceof Error ? error.message : 'Unsubscription failed')
    }
  }

  const handleSendMessage = async () => {
    try {
      const message = {
        type: 'ping',
        data: { timestamp: Date.now(), demo: true },
        timestamp: new Date().toISOString()
      }
      
      await webSocketService.send(message)
      
      // Update metrics
      setMetrics(prev => prev ? {
        ...prev,
        messagesSent: prev.messagesSent + 1
      } : null)
    } catch (error) {
      setLastError(error instanceof Error ? error.message : 'Send failed')
    }
  }

  const getConnectionStatusColor = () => {
    switch (connectionState) {
      case 'connected': return 'text-green-600'
      case 'connecting': return 'text-yellow-600'
      case 'disconnected': return 'text-red-600'
      default: return 'text-gray-600'
    }
  }

  const getHealthStatusColor = () => {
    if (!health) return 'text-gray-600'
    switch (health.status) {
      case 'healthy': return 'text-green-600'
      case 'degraded': return 'text-yellow-600'
      case 'unhealthy': return 'text-red-600'
      default: return 'text-gray-600'
    }
  }

  return (
    <div className={`p-6 max-w-4xl mx-auto space-y-6 ${className}`}>
      <Card className="p-6">
        <h2 className="text-2xl font-bold mb-4">WebSocket Integration Demo</h2>
        
        {/* Connection Status */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
          <div className="bg-gray-50 p-4 rounded-lg">
            <h3 className="font-semibold text-sm mb-2">Connection</h3>
            <p className={`font-mono text-sm ${getConnectionStatusColor()}`}>
              {connectionState.toUpperCase()}
            </p>
            {isConnecting && <p className="text-xs text-gray-500">Connecting...</p>}
          </div>
          
          <div className="bg-gray-50 p-4 rounded-lg">
            <h3 className="font-semibold text-sm mb-2">Health</h3>
            <p className={`font-mono text-sm ${getHealthStatusColor()}`}>
              {health?.status?.toUpperCase() || 'UNKNOWN'}
            </p>
            {health && (
              <p className="text-xs text-gray-500">{health.latency}ms</p>
            )}
          </div>
          
          <div className="bg-gray-50 p-4 rounded-lg">
            <h3 className="font-semibold text-sm mb-2">Subscriptions</h3>
            <p className="font-mono text-sm">{subscriptions.length}</p>
            <p className="text-xs text-gray-500">Active topics</p>
          </div>
          
          <div className="bg-gray-50 p-4 rounded-lg">
            <h3 className="font-semibold text-sm mb-2">Messages</h3>
            <p className="font-mono text-sm">{receivedMessages.length}</p>
            <p className="text-xs text-gray-500">Received</p>
          </div>
        </div>

        {/* Error Display */}
        {lastError && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
            <h3 className="font-semibold text-red-800 mb-2">Error</h3>
            <p className="text-red-600 text-sm">{lastError}</p>
            <Button 
              variant="outline" 
              size="sm" 
              onClick={() => setLastError(null)}
              className="mt-2"
            >
              Dismiss
            </Button>
          </div>
        )}

        {/* Connection Controls */}
        <div className="flex gap-4 mb-6">
          <Button 
            onClick={handleConnect}
            disabled={connectionState === 'connected' || isConnecting}
            variant="default"
          >
            {isConnecting ? 'Connecting...' : 'Connect'}
          </Button>
          
          <Button 
            onClick={handleDisconnect}
            disabled={connectionState === 'disconnected'}
            variant="outline"
          >
            Disconnect
          </Button>
          
          <Button 
            onClick={handleSendMessage}
            disabled={connectionState !== 'connected'}
            variant="outline"
          >
            Send Ping
          </Button>
        </div>

        {/* Topic Subscriptions */}
        <div className="mb-6">
          <h3 className="font-semibold mb-3">Topic Subscriptions</h3>
          <div className="flex flex-wrap gap-2">
            {demoTopics.map(topic => {
              const isSubscribed = subscriptions.includes(topic)
              return (
                <div key={topic} className="flex items-center gap-2">
                  <Badge 
                    variant={isSubscribed ? "default" : "secondary"}
                    className="cursor-pointer"
                    onClick={() => isSubscribed ? handleUnsubscribe(topic) : handleSubscribe(topic)}
                  >
                    {topic} {isSubscribed ? '✓' : '+'}
                  </Badge>
                </div>
              )
            })}
          </div>
        </div>

        {/* Connection Metrics */}
        {showMetrics && metrics && (
          <div className="bg-gray-50 rounded-lg p-4 mb-6">
            <h3 className="font-semibold mb-3">Connection Metrics</h3>
            <div className="grid grid-cols-2 md:grid-cols-3 gap-4 text-sm">
              <div>
                <span className="font-medium">Total Connections:</span>
                <span className="ml-2">{metrics.totalConnections}</span>
              </div>
              <div>
                <span className="font-medium">Messages Sent:</span>
                <span className="ml-2">{metrics.messagesSent}</span>
              </div>
              <div>
                <span className="font-medium">Messages Received:</span>
                <span className="ml-2">{metrics.messagesReceived}</span>
              </div>
              <div>
                <span className="font-medium">Avg Latency:</span>
                <span className="ml-2">{Math.round(metrics.avgLatency)}ms</span>
              </div>
              <div>
                <span className="font-medium">Total Errors:</span>
                <span className="ml-2">{metrics.totalErrors}</span>
              </div>
              <div>
                <span className="font-medium">Uptime:</span>
                <span className="ml-2">{Math.round(metrics.uptime / 1000)}s</span>
              </div>
            </div>
          </div>
        )}

        {/* Recent Messages */}
        {receivedMessages.length > 0 && (
          <div className="bg-gray-50 rounded-lg p-4">
            <h3 className="font-semibold mb-3">Recent Messages ({receivedMessages.length})</h3>
            <div className="space-y-2 max-h-64 overflow-y-auto">
              {receivedMessages.map((msg, index) => (
                <div key={index} className="bg-white p-3 rounded border text-sm">
                  <div className="flex justify-between items-start mb-1">
                    <span className="font-medium text-blue-600">{msg.type}</span>
                    <span className="text-gray-500 text-xs">
                      {new Date(msg.timestamp).toLocaleTimeString()}
                    </span>
                  </div>
                  <pre className="text-xs text-gray-700 overflow-x-auto">
                    {JSON.stringify(msg.payload, null, 2)}
                  </pre>
                </div>
              ))}
            </div>
          </div>
        )}
      </Card>
    </div>
  )
}

export default WebSocketDemo