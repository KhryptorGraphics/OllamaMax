/**
 * Monitoring and metrics types
 */

export interface MonitoringState {
  metrics: MonitoringMetrics | null
  alerts: MonitoringAlert[]
  logs: LogEntry[]
  thresholds: AlertThreshold[]
  dashboardConfig: DashboardConfig
  loading: boolean
  error: string | null
  lastUpdated: number | null
}

export interface MonitoringMetrics {
  timestamp: string
  system: SystemMetrics
  cluster: ClusterMetrics
  models: ModelMetrics[]
  endpoints: EndpointMetrics[]
  custom: CustomMetric[]
}

export interface SystemMetrics {
  cpu: MetricValue
  memory: MetricValue
  disk: MetricValue
  network: NetworkMetrics
  load: number[]
  uptime: number
  temperature?: MetricValue
  processes: number
}

export interface MetricValue {
  current: number
  average: number
  peak: number
  minimum: number
  trend: 'up' | 'down' | 'stable'
  threshold?: number
  unit: string
  history: TimeSeriesPoint[]
}

export interface TimeSeriesPoint {
  timestamp: string
  value: number
  label?: string
}

export interface NetworkMetrics {
  throughput: {
    in: number
    out: number
  }
  latency: {
    min: number
    max: number
    avg: number
    p95: number
    p99: number
  }
  errors: number
  packets: {
    in: number
    out: number
    dropped: number
  }
  connections: {
    active: number
    total: number
    failed: number
  }
}

export interface ClusterMetrics {
  totalNodes: number
  activeNodes: number
  totalModels: number
  activeTasks: number
  averageResponseTime: number
  throughput: number
  errorRate: number
  replicationFactor: number
  syncStatus: 'synced' | 'syncing' | 'error'
  health: ClusterHealth
}

export interface ClusterHealth {
  status: 'healthy' | 'degraded' | 'critical'
  score: number
  issues: string[]
  lastCheck: string
}

export interface ModelMetrics {
  name: string
  requests: number
  averageResponseTime: number
  errorRate: number
  throughput: number
  lastUsed: string
  memoryUsage: number
  cpuUsage: number
  accuracy?: number
  version: string
}

export interface EndpointMetrics {
  path: string
  method: string
  requests: number
  averageResponseTime: number
  errorRate: number
  statusCodes: Record<string, number>
  lastAccessed: string
  throughput: number
}

export interface CustomMetric {
  id: string
  name: string
  value: number
  unit: string
  category: string
  description: string
  history: TimeSeriesPoint[]
  threshold?: AlertThreshold
}

export interface MonitoringAlert {
  id: string
  type: AlertType
  severity: AlertSeverity
  message: string
  description?: string
  source: string
  metric: string
  threshold: number
  currentValue: number
  timestamp: number
  acknowledged: boolean
  acknowledgedBy?: string
  acknowledgedAt?: number
  resolvedAt?: number
  actions: AlertAction[]
  tags: string[]
  category: AlertCategory
}

export type AlertSeverity = 'info' | 'warning' | 'error' | 'critical'

export type AlertCategory = 
  | 'system'
  | 'cluster'
  | 'model'
  | 'network'
  | 'security'
  | 'performance'
  | 'custom'

export type AlertType = 
  | 'cpu_high'
  | 'memory_high'
  | 'disk_full'
  | 'disk_low'
  | 'network_slow'
  | 'network_error'
  | 'response_slow'
  | 'error_rate_high'
  | 'throughput_low'
  | 'node_down'
  | 'node_degraded'
  | 'model_error'
  | 'model_timeout'
  | 'security_breach'
  | 'auth_failure'
  | 'custom_threshold'

export interface AlertAction {
  id: string
  label: string
  action: string
  params?: Record<string, any>
  confirmRequired?: boolean
}

export interface AlertThreshold {
  id: string
  name: string
  metric: string
  operator: ThresholdOperator
  value: number
  severity: AlertSeverity
  enabled: boolean
  conditions: ThresholdCondition[]
  actions: AlertAction[]
  cooldown: number
  createdAt: string
  updatedAt: string
}

export type ThresholdOperator = 'gt' | 'gte' | 'lt' | 'lte' | 'eq' | 'neq'

export interface ThresholdCondition {
  field: string
  operator: ThresholdOperator
  value: any
  duration?: number
}

export interface LogEntry {
  id: string
  timestamp: string
  level: LogLevel
  message: string
  source: string
  category: string
  metadata?: Record<string, any>
  tags: string[]
  correlation_id?: string
}

export type LogLevel = 'debug' | 'info' | 'warn' | 'error' | 'fatal'

export interface DashboardConfig {
  widgets: DashboardWidget[]
  layout: DashboardLayout
  timeRange: TimeRange
  refreshInterval: number
  autoRefresh: boolean
  theme: 'light' | 'dark' | 'auto'
}

export interface DashboardWidget {
  id: string
  type: WidgetType
  title: string
  position: WidgetPosition
  size: WidgetSize
  config: WidgetConfig
  visible: boolean
}

export type WidgetType = 
  | 'metric'
  | 'chart'
  | 'table'
  | 'alert_list'
  | 'log_viewer'
  | 'status'
  | 'gauge'
  | 'heatmap'
  | 'custom'

export interface WidgetPosition {
  x: number
  y: number
}

export interface WidgetSize {
  width: number
  height: number
}

export interface WidgetConfig {
  metric?: string
  chartType?: ChartType
  timeRange?: TimeRange
  aggregation?: AggregationType
  filters?: Record<string, any>
  colors?: string[]
  showLegend?: boolean
  showGrid?: boolean
  [key: string]: any
}

export type ChartType = 
  | 'line'
  | 'area'
  | 'bar'
  | 'pie'
  | 'scatter'
  | 'heatmap'
  | 'gauge'

export type AggregationType = 'avg' | 'sum' | 'min' | 'max' | 'count' | 'p50' | 'p95' | 'p99'

export interface DashboardLayout {
  columns: number
  padding: number
  margin: number
  compact: boolean
}

export interface TimeRange {
  start: string | number
  end: string | number
  preset?: TimeRangePreset
}

export type TimeRangePreset = 
  | '5m'
  | '15m'
  | '30m'
  | '1h'
  | '3h'
  | '6h'
  | '12h'
  | '24h'
  | '7d'
  | '30d'
  | 'custom'

export interface MetricFilter {
  field: string
  operator: 'eq' | 'neq' | 'gt' | 'gte' | 'lt' | 'lte' | 'contains' | 'regex'
  value: any
}

export interface ExportOptions {
  format: 'csv' | 'json' | 'pdf' | 'png'
  timeRange: TimeRange
  metrics: string[]
  includeAlerts?: boolean
  includeLogs?: boolean
}

export interface CorrelationMatrix {
  metrics: string[]
  correlations: number[][]
  timestamp: string
}

export interface AnomalyDetection {
  metric: string
  anomalies: AnomalyPoint[]
  confidence: number
  algorithm: string
  parameters: Record<string, any>
}

export interface AnomalyPoint {
  timestamp: string
  value: number
  expected: number
  deviation: number
  severity: number
}