import React, { useState, useEffect } from 'react'
import { 
  TrendingUp, 
  Brain, 
  Zap, 
  Target, 
  AlertTriangle,
  Clock,
  Activity,
  BarChart3,
  LineChart,
  PieChart,
  RefreshCw,
  Download,
  Filter,
  Calendar,
  Settings
} from 'lucide-react'
import { Button, Card, Input, Alert, Badge, Select } from '@/design-system'
import { useAPI } from '@/hooks/useAPI'
import { useWebSocket } from '@/hooks/useWebSocket'
import { formatBytes, formatNumber, formatDuration } from '@/utils/format'

interface Prediction {
  id: string
  type: 'resource_usage' | 'model_performance' | 'failure_prediction' | 'capacity_planning'
  title: string
  description: string
  confidence: number
  timeHorizon: string
  severity: 'low' | 'medium' | 'high' | 'critical'
  predictions: PredictionData[]
  recommendations: string[]
  createdAt: string
  updatedAt: string
}

interface PredictionData {
  timestamp: number
  predicted: number
  confidence: number
  actualValue?: number
  upperBound: number
  lowerBound: number
}

interface AnalyticsConfig {
  models: string[]
  timeRange: '1h' | '6h' | '24h' | '7d' | '30d'
  predictionHorizon: '1h' | '6h' | '24h' | '7d'
  refreshInterval: number
  enableRealTime: boolean
  algorithms: string[]
  thresholds: {
    [key: string]: number
  }
}

interface MetricTrend {
  metric: string
  current: number
  trend: 'up' | 'down' | 'stable'
  change: number
  predicted: number
  confidence: number
}

export const PredictiveAnalytics: React.FC = () => {
  const [predictions, setPredictions] = useState<Prediction[]>([])
  const [selectedPrediction, setSelectedPrediction] = useState<Prediction | null>(null)
  const [trends, setTrends] = useState<MetricTrend[]>([])
  const [config, setConfig] = useState<AnalyticsConfig>({
    models: [],
    timeRange: '24h',
    predictionHorizon: '6h',
    refreshInterval: 30000,
    enableRealTime: true,
    algorithms: ['linear_regression', 'arima', 'neural_network'],
    thresholds: {
      cpu_usage: 80,
      memory_usage: 85,
      disk_usage: 90,
      error_rate: 5,
      response_time: 1000
    }
  })
  const [isTraining, setIsTraining] = useState(false)
  const [lastUpdate, setLastUpdate] = useState<Date>(new Date())

  const { data: availableModels } = useAPI('/api/models')
  const { data: analyticsData, mutate: refreshAnalytics } = useAPI('/api/analytics/predictions', {
    refreshInterval: config.refreshInterval
  })

  const ws = useWebSocket('ws://localhost:8080/ws/analytics', {
    onMessage: (data) => {
      if (data.type === 'prediction_update') {
        setPredictions(prev => {
          const index = prev.findIndex(p => p.id === data.prediction.id)
          if (index >= 0) {
            const updated = [...prev]
            updated[index] = data.prediction
            return updated
          }
          return [...prev, data.prediction]
        })
      } else if (data.type === 'trends_update') {
        setTrends(data.trends)
      }
      setLastUpdate(new Date())
    }
  })

  useEffect(() => {
    if (analyticsData) {
      setPredictions(analyticsData.predictions || [])
      setTrends(analyticsData.trends || [])
    }
  }, [analyticsData])

  const generatePredictions = async () => {
    setIsTraining(true)
    try {
      const response = await fetch('/api/analytics/predictions/generate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(config)
      })
      
      if (!response.ok) throw new Error('Failed to generate predictions')
      
      refreshAnalytics()
    } catch (error) {
      console.error('Failed to generate predictions:', error)
    } finally {
      setIsTraining(false)
    }
  }

  const exportPredictions = () => {
    const data = {
      predictions,
      trends,
      config,
      exportedAt: new Date().toISOString()
    }
    
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `predictions_${new Date().toISOString().split('T')[0]}.json`
    a.click()
    URL.revokeObjectURL(url)
  }

  const getPredictionIcon = (type: string) => {
    switch (type) {
      case 'resource_usage': return <Activity className="w-4 h-4" />
      case 'model_performance': return <Target className="w-4 h-4" />
      case 'failure_prediction': return <AlertTriangle className="w-4 h-4" />
      case 'capacity_planning': return <BarChart3 className="w-4 h-4" />
      default: return <Brain className="w-4 h-4" />
    }
  }

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical': return 'destructive'
      case 'high': return 'warning'
      case 'medium': return 'default'
      case 'low': return 'secondary'
      default: return 'default'
    }
  }

  const getTrendIcon = (trend: string) => {
    switch (trend) {
      case 'up': return <TrendingUp className="w-4 h-4 text-green-500" />
      case 'down': return <TrendingUp className="w-4 h-4 text-red-500 rotate-180" />
      default: return <TrendingUp className="w-4 h-4 text-gray-500 rotate-90" />
    }
  }

  return (
    <div className="space-y-6">
      {/* Analytics Configuration */}
      <Card>
        <Card.Header>
          <Card.Title>
            <Brain className="w-5 h-5 inline mr-2" />
            Predictive Analytics Configuration
          </Card.Title>
          <Card.Description>
            Configure ML models and prediction parameters
          </Card.Description>
        </Card.Header>
        
        <Card.Content className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {/* Time Range */}
            <div>
              <label className="block text-sm font-medium mb-2">Analysis Time Range</label>
              <select
                className="w-full px-3 py-2 border rounded-md"
                value={config.timeRange}
                onChange={(e) => setConfig({ ...config, timeRange: e.target.value as any })}
              >
                <option value="1h">Last Hour</option>
                <option value="6h">Last 6 Hours</option>
                <option value="24h">Last 24 Hours</option>
                <option value="7d">Last 7 Days</option>
                <option value="30d">Last 30 Days</option>
              </select>
            </div>

            {/* Prediction Horizon */}
            <div>
              <label className="block text-sm font-medium mb-2">Prediction Horizon</label>
              <select
                className="w-full px-3 py-2 border rounded-md"
                value={config.predictionHorizon}
                onChange={(e) => setConfig({ ...config, predictionHorizon: e.target.value as any })}
              >
                <option value="1h">Next Hour</option>
                <option value="6h">Next 6 Hours</option>
                <option value="24h">Next 24 Hours</option>
                <option value="7d">Next 7 Days</option>
              </select>
            </div>

            {/* Refresh Interval */}
            <div>
              <label className="block text-sm font-medium mb-2">Refresh Interval</label>
              <select
                className="w-full px-3 py-2 border rounded-md"
                value={config.refreshInterval}
                onChange={(e) => setConfig({ ...config, refreshInterval: parseInt(e.target.value) })}
              >
                <option value={10000}>10 seconds</option>
                <option value={30000}>30 seconds</option>
                <option value={60000}>1 minute</option>
                <option value={300000}>5 minutes</option>
              </select>
            </div>
          </div>

          {/* Model Selection */}
          <div>
            <label className="block text-sm font-medium mb-2">Models to Analyze</label>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-2 max-h-32 overflow-y-auto">
              {availableModels?.map((model: any) => (
                <label key={model.id} className="flex items-center space-x-2">
                  <input
                    type="checkbox"
                    checked={config.models.includes(model.id)}
                    onChange={(e) => {
                      if (e.target.checked) {
                        setConfig({ ...config, models: [...config.models, model.id] })
                      } else {
                        setConfig({ ...config, models: config.models.filter(m => m !== model.id) })
                      }
                    }}
                  />
                  <span className="text-sm">{model.name}</span>
                </label>
              ))}
            </div>
          </div>

          {/* Controls */}
          <div className="flex justify-between items-center">
            <div className="flex items-center space-x-2">
              <label className="flex items-center space-x-2">
                <input
                  type="checkbox"
                  checked={config.enableRealTime}
                  onChange={(e) => setConfig({ ...config, enableRealTime: e.target.checked })}
                />
                <span className="text-sm">Real-time Updates</span>
              </label>
              <span className="text-xs text-muted-foreground">
                Last update: {lastUpdate.toLocaleTimeString()}
              </span>
            </div>
            
            <div className="flex space-x-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => refreshAnalytics()}
              >
                <RefreshCw className="w-4 h-4 mr-2" />
                Refresh
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={exportPredictions}
                disabled={predictions.length === 0}
              >
                <Download className="w-4 h-4 mr-2" />
                Export
              </Button>
              <Button
                onClick={generatePredictions}
                loading={isTraining}
                loadingText="Generating..."
                disabled={config.models.length === 0}
              >
                <Brain className="w-4 h-4 mr-2" />
                Generate Predictions
              </Button>
            </div>
          </div>
        </Card.Content>
      </Card>

      {/* Key Trends */}
      <Card>
        <Card.Header>
          <Card.Title>
            <TrendingUp className="w-5 h-5 inline mr-2" />
            Key Metric Trends
          </Card.Title>
        </Card.Header>
        
        <Card.Content>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {trends.map((trend) => (
              <div key={trend.metric} className="p-4 border rounded-lg">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm font-medium capitalize">
                    {trend.metric.replace('_', ' ')}
                  </span>
                  {getTrendIcon(trend.trend)}
                </div>
                
                <div className="space-y-1">
                  <div className="text-2xl font-bold">
                    {formatNumber(trend.current)}
                  </div>
                  <div className="text-sm text-muted-foreground">
                    {trend.trend === 'up' ? '+' : trend.trend === 'down' ? '-' : ''}
                    {Math.abs(trend.change).toFixed(1)}% from baseline
                  </div>
                  <div className="text-sm">
                    Predicted: {formatNumber(trend.predicted)}
                    <span className="text-muted-foreground ml-1">
                      ({(trend.confidence * 100).toFixed(0)}% confidence)
                    </span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </Card.Content>
      </Card>

      {/* Predictions List */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <Card.Header>
            <Card.Title>Active Predictions</Card.Title>
            <Card.Description>
              ML-generated predictions and recommendations
            </Card.Description>
          </Card.Header>
          
          <Card.Content>
            <div className="space-y-3 max-h-96 overflow-y-auto">
              {predictions.map((prediction) => (
                <div
                  key={prediction.id}
                  className={`p-3 border rounded-lg cursor-pointer transition-colors ${
                    selectedPrediction?.id === prediction.id ? 'border-primary bg-primary/5' : 'hover:bg-gray-50'
                  }`}
                  onClick={() => setSelectedPrediction(prediction)}
                >
                  <div className="flex items-start justify-between">
                    <div className="flex items-start space-x-2">
                      {getPredictionIcon(prediction.type)}
                      <div className="flex-1">
                        <h4 className="text-sm font-medium">{prediction.title}</h4>
                        <p className="text-xs text-muted-foreground line-clamp-2">
                          {prediction.description}
                        </p>
                      </div>
                    </div>
                    
                    <div className="flex flex-col items-end space-y-1">
                      <Badge variant={getSeverityColor(prediction.severity)}>
                        {prediction.severity}
                      </Badge>
                      <span className="text-xs text-muted-foreground">
                        {(prediction.confidence * 100).toFixed(0)}%
                      </span>
                    </div>
                  </div>
                  
                  <div className="mt-2 text-xs text-muted-foreground">
                    Horizon: {prediction.timeHorizon} • 
                    Updated: {new Date(prediction.updatedAt).toLocaleTimeString()}
                  </div>
                </div>
              ))}
              
              {predictions.length === 0 && (
                <div className="text-center text-muted-foreground py-8">
                  No predictions available. Generate new predictions to get started.
                </div>
              )}
            </div>
          </Card.Content>
        </Card>

        {/* Prediction Details */}
        <Card>
          <Card.Header>
            <Card.Title>Prediction Details</Card.Title>
          </Card.Header>
          
          <Card.Content>
            {selectedPrediction ? (
              <div className="space-y-4">
                {/* Prediction Header */}
                <div>
                  <div className="flex items-center space-x-2 mb-2">
                    {getPredictionIcon(selectedPrediction.type)}
                    <h3 className="font-semibold">{selectedPrediction.title}</h3>
                    <Badge variant={getSeverityColor(selectedPrediction.severity)}>
                      {selectedPrediction.severity}
                    </Badge>
                  </div>
                  <p className="text-sm text-muted-foreground">
                    {selectedPrediction.description}
                  </p>
                </div>

                {/* Prediction Chart */}
                <div className="h-48 bg-gray-50 rounded-lg p-4">
                  <h4 className="text-sm font-medium mb-2">Prediction Timeline</h4>
                  {/* Chart would be rendered here using a charting library */}
                  <div className="text-center text-muted-foreground mt-16">
                    Prediction chart visualization
                  </div>
                </div>

                {/* Recommendations */}
                <div>
                  <h4 className="text-sm font-medium mb-2">Recommendations</h4>
                  <ul className="space-y-1">
                    {selectedPrediction.recommendations.map((recommendation, index) => (
                      <li key={index} className="text-sm text-muted-foreground flex items-start">
                        <span className="w-1.5 h-1.5 bg-primary rounded-full mt-2 mr-2 flex-shrink-0" />
                        {recommendation}
                      </li>
                    ))}
                  </ul>
                </div>

                {/* Metadata */}
                <div className="border-t pt-4 space-y-2">
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Confidence:</span>
                    <span>{(selectedPrediction.confidence * 100).toFixed(1)}%</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Time Horizon:</span>
                    <span>{selectedPrediction.timeHorizon}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Created:</span>
                    <span>{new Date(selectedPrediction.createdAt).toLocaleString()}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Last Updated:</span>
                    <span>{new Date(selectedPrediction.updatedAt).toLocaleString()}</span>
                  </div>
                </div>
              </div>
            ) : (
              <div className="text-center text-muted-foreground py-8">
                Select a prediction to view details
              </div>
            )}
          </Card.Content>
        </Card>
      </div>

      {/* Alerts for Critical Predictions */}
      {predictions.filter(p => p.severity === 'critical').length > 0 && (
        <Alert variant="destructive">
          <AlertTriangle className="w-4 h-4" />
          <Alert.Title>Critical Predictions Detected</Alert.Title>
          <Alert.Description>
            {predictions.filter(p => p.severity === 'critical').length} critical prediction(s) require immediate attention.
            Review the recommendations and take appropriate action.
          </Alert.Description>
        </Alert>
      )}
    </div>
  )
}

export default PredictiveAnalytics