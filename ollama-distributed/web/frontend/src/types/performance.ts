// Performance monitoring types

export interface PerformanceState {
  metrics: PerformanceMetrics | null
  alerts: PerformanceAlert[]
  lastUpdated: number | null
  loading: boolean
  error: string | null
}

export interface PerformanceMetrics {
  timestamp: string
  system: SystemMetrics
  cluster: ClusterMetrics
  models: ModelMetrics[]
  endpoints: EndpointMetrics[]
}

export interface SystemMetrics {
  cpu: MetricValue
  memory: MetricValue
  disk: MetricValue
  network: NetworkMetrics
  load: number[]
  uptime: number
}

export interface MetricValue {
  current: number
  average: number
  peak: number
  trend: 'up' | 'down' | 'stable'
  threshold?: number
  unit: string
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
  }
  errors: number
  packets: {
    in: number
    out: number
    dropped: number
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
}

export interface ModelMetrics {
  name: string
  requests: number
  averageResponseTime: number
  errorRate: number
  throughput: number
  lastUsed: string
}

export interface EndpointMetrics {
  path: string
  method: string
  requests: number
  averageResponseTime: number
  errorRate: number
  statusCodes: Record<string, number>
}

export interface PerformanceAlert {
  id: string
  type: AlertType
  severity: 'low' | 'medium' | 'high' | 'critical'
  message: string
  source: string
  metric: string
  threshold: number
  currentValue: number
  timestamp: number
  acknowledged: boolean
  resolvedAt?: number
}

export type AlertType = 
  | 'cpu_high'
  | 'memory_high'
  | 'disk_full'
  | 'network_slow'
  | 'response_slow'
  | 'error_rate_high'
  | 'throughput_low'
  | 'node_down'
  | 'model_error'