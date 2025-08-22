import React, { useState, useEffect } from 'react'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Select } from '@/components/ui/Select'
import { useWebSocket } from '@/hooks/useWebSocket'
import {
  BarChart, LineChart, ScatterChart, Treemap, Sankey,
  Bar, Line, Scatter, XAxis, YAxis, CartesianGrid,
  Tooltip, Legend, ResponsiveContainer, Cell
} from 'recharts'
import {
  TrendingUp, Target, DollarSign, Users, Package,
  AlertTriangle, CheckCircle, Clock, Zap, Globe
} from 'lucide-react'

interface KPI {
  id: string
  name: string
  value: number
  target: number
  unit: string
  trend: 'up' | 'down' | 'stable'
  percentChange: number
  status: 'on-track' | 'at-risk' | 'critical'
}

interface Forecast {
  metric: string
  current: number
  predicted: number
  confidence: number
  timeframe: string
}

interface Segment {
  name: string
  value: number
  growth: number
  share: number
}

export const BusinessIntelligence: React.FC = () => {
  const [kpis, setKPIs] = useState<KPI[]>([])
  const [forecasts, setForecasts] = useState<Forecast[]>([])
  const [segments, setSegments] = useState<Segment[]>([])
  const [selectedDimension, setSelectedDimension] = useState('revenue')
  const [timeframe, setTimeframe] = useState('quarter')

  const { sendMessage, lastMessage } = useWebSocket('/ws/business-intelligence')

  useEffect(() => {
    if (lastMessage) {
      const data = JSON.parse(lastMessage.data)
      if (data.type === 'kpis') setKPIs(data.kpis)
      if (data.type === 'forecasts') setForecasts(data.forecasts)
      if (data.type === 'segments') setSegments(data.segments)
    }
  }, [lastMessage])

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'on-track': return 'text-green-600'
      case 'at-risk': return 'text-yellow-600'
      case 'critical': return 'text-red-600'
      default: return 'text-gray-600'
    }
  }

  const getTrendIcon = (trend: string) => {
    if (trend === 'up') return '↑'
    if (trend === 'down') return '↓'
    return '→'
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold">Business Intelligence</h2>
        <Select value={timeframe} onChange={setTimeframe}>
          <option value="daily">Daily</option>
          <option value="weekly">Weekly</option>
          <option value="monthly">Monthly</option>
          <option value="quarter">Quarterly</option>
          <option value="yearly">Yearly</option>
        </Select>
      </div>

      {/* KPI Dashboard */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {kpis.map(kpi => (
          <Card key={kpi.id}>
            <div className="p-4">
              <div className="flex justify-between items-start mb-2">
                <p className="text-sm text-gray-600">{kpi.name}</p>
                <span className={`text-xs font-medium px-2 py-1 rounded ${
                  kpi.status === 'on-track' ? 'bg-green-100 text-green-700' :
                  kpi.status === 'at-risk' ? 'bg-yellow-100 text-yellow-700' :
                  'bg-red-100 text-red-700'
                }`}>
                  {kpi.status}
                </span>
              </div>
              <div className="flex items-baseline gap-2">
                <p className="text-2xl font-bold">
                  {kpi.value.toLocaleString()}{kpi.unit}
                </p>
                <span className={`text-sm ${
                  kpi.trend === 'up' ? 'text-green-600' : 
                  kpi.trend === 'down' ? 'text-red-600' : 
                  'text-gray-600'
                }`}>
                  {getTrendIcon(kpi.trend)} {Math.abs(kpi.percentChange)}%
                </span>
              </div>
              <div className="mt-2">
                <div className="flex justify-between text-xs text-gray-600 mb-1">
                  <span>Target: {kpi.target.toLocaleString()}{kpi.unit}</span>
                  <span>{((kpi.value / kpi.target) * 100).toFixed(0)}%</span>
                </div>
                <div className="w-full bg-gray-200 rounded-full h-2">
                  <div 
                    className={`h-2 rounded-full ${
                      kpi.status === 'on-track' ? 'bg-green-600' :
                      kpi.status === 'at-risk' ? 'bg-yellow-600' :
                      'bg-red-600'
                    }`}
                    style={{ width: `${Math.min((kpi.value / kpi.target) * 100, 100)}%` }}
                  />
                </div>
              </div>
            </div>
          </Card>
        ))}
      </div>

      {/* Predictive Analytics */}
      <Card>
        <div className="p-4">
          <h3 className="text-lg font-semibold mb-4">Predictive Forecasting</h3>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {forecasts.map(forecast => (
              <div key={forecast.metric} className="border rounded-lg p-4">
                <p className="text-sm text-gray-600 mb-2">{forecast.metric}</p>
                <div className="flex items-center justify-between mb-2">
                  <div>
                    <p className="text-xs text-gray-500">Current</p>
                    <p className="text-lg font-semibold">{forecast.current.toLocaleString()}</p>
                  </div>
                  <Zap className="w-4 h-4 text-blue-500" />
                  <div>
                    <p className="text-xs text-gray-500">Predicted</p>
                    <p className="text-lg font-semibold">{forecast.predicted.toLocaleString()}</p>
                  </div>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-xs text-gray-600">{forecast.timeframe}</span>
                  <span className="text-xs">
                    Confidence: {forecast.confidence}%
                  </span>
                </div>
                <div className="mt-2 w-full bg-gray-200 rounded-full h-1">
                  <div 
                    className="bg-blue-600 h-1 rounded-full"
                    style={{ width: `${forecast.confidence}%` }}
                  />
                </div>
              </div>
            ))}
          </div>
        </div>
      </Card>

      {/* Market Segmentation */}
      <Card>
        <div className="p-4">
          <h3 className="text-lg font-semibold mb-4">Market Segmentation</h3>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={segments}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="name" />
              <YAxis />
              <Tooltip />
              <Legend />
              <Bar dataKey="value" fill="#2563eb" />
              <Bar dataKey="growth" fill="#10b981" />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </Card>
    </div>
  )
}