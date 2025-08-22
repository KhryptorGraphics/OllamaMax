import React, { useEffect, useState } from 'react'
import { Card, Button } from '@/design-system/components'
import { webSocketService } from '@/services/websocket/service'
import { storeIntegration } from '@/services/websocket/integration'

interface Metric {
  label: string
  value: string
  trend: 'up' | 'down' | 'stable'
}

export function DashboardWithWebSocket() {
  const [metrics, setMetrics] = useState<Metric[]>([
    { label: 'Active Nodes', value: '12', trend: 'stable' },
    { label: 'Models Synced', value: '45', trend: 'up' },
    { label: 'Tasks Running', value: '8', trend: 'down' },
    { label: 'Avg Response', value: '120ms', trend: 'stable' }
  ])

  const [isConnected, setIsConnected] = useState(false)

  useEffect(() => {
    const initializeWebSocket = async () => {
      try {
        await webSocketService.connect()
        await webSocketService.subscribeToPage('dashboard')
        setIsConnected(true)
      } catch (error) {
        console.error('Failed to connect WebSocket:', error)
      }
    }

    initializeWebSocket()

    return () => {
      webSocketService.disconnect()
    }
  }, [])

  const refreshData = async () => {
    await storeIntegration.performOptimisticUpdate(
      'refresh-dashboard',
      'dashboard',
      () => {
        // Optimistic update
        setMetrics(prev => prev.map(m => ({ ...m, value: 'Loading...' })))
      },
      async () => {
        // Actual refresh simulation
        await new Promise(resolve => setTimeout(resolve, 1000))
        setMetrics([
          { label: 'Active Nodes', value: '13', trend: 'up' },
          { label: 'Models Synced', value: '47', trend: 'up' },
          { label: 'Tasks Running', value: '6', trend: 'down' },
          { label: 'Avg Response', value: '115ms', trend: 'up' }
        ])
      },
      () => {
        // Rollback on error
        setMetrics([
          { label: 'Active Nodes', value: '12', trend: 'stable' },
          { label: 'Models Synced', value: '45', trend: 'up' },
          { label: 'Tasks Running', value: '8', trend: 'down' },
          { label: 'Avg Response', value: '120ms', trend: 'stable' }
        ])
      }
    )
  }

  return (
    <div className="p-6 space-y-6">
      <Card className="p-4">
        <div className="flex items-center justify-between mb-4">
          <h1 className="text-2xl font-bold">Dashboard with WebSocket</h1>
          <div className="flex items-center gap-4">
            <div className={`flex items-center gap-2 px-3 py-1 rounded-full text-sm ${
              isConnected 
                ? 'bg-green-100 text-green-800' 
                : 'bg-red-100 text-red-800'
            }`}>
              <div className={`w-2 h-2 rounded-full ${
                isConnected ? 'bg-green-500' : 'bg-red-500'
              }`} />
              {isConnected ? 'Connected' : 'Disconnected'}
            </div>
            <Button onClick={refreshData} variant="outline" size="sm">
              Refresh
            </Button>
          </div>
        </div>
        
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {metrics.map((metric) => (
            <Card key={metric.label} className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-gray-600">{metric.label}</p>
                  <p className="text-2xl font-bold">{metric.value}</p>
                </div>
                <div className={`text-sm ${
                  metric.trend === 'up' ? 'text-green-600' : 
                  metric.trend === 'down' ? 'text-red-600' : 
                  'text-gray-600'
                }`}>
                  {metric.trend === 'up' ? '↗' : 
                   metric.trend === 'down' ? '↘' : 
                   '→'}
                </div>
              </div>
            </Card>
          ))}
        </div>
      </Card>
    </div>
  )
}