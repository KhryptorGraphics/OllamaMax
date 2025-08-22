// Analytics Feature Export Index

// Components
export { default as AnalyticsDashboard } from './components/AnalyticsDashboard'
export { BusinessIntelligence } from './components/BusinessIntelligence'
export { PredictiveAnalytics } from './components/PredictiveAnalytics'

// Chart Components
export { Chart, LineChartComponent, AreaChartComponent, BarChartComponent, PieChartComponent, ScatterChartComponent } from './components/charts/Chart'

// Services
export { AnalyticsService, analyticsService } from './services/analyticsService'

// Hooks
export { default as useRealTimeAnalytics } from './hooks/useRealTimeAnalytics'
export { useAnalytics } from './services/analyticsService'

// Utilities
export { 
  DataAggregator, 
  dataAggregator, 
  aggregateData, 
  aggregateTimeSeries, 
  analyzeCohorts, 
  analyzeFunnel,
  calculateStatistics,
  detectAnomalies,
  TimePeriod
} from './utils/dataAggregation'

// Types
export type * from './types'