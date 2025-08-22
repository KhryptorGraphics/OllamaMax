// API-related types for the OllamaMax distributed system

export interface BaseApiResponse {
  timestamp: string
  requestId?: string
  version?: string
}

export interface SuccessResponse<T = any> extends BaseApiResponse {
  success: true
  data: T
}

export interface ErrorResponse extends BaseApiResponse {
  success: false
  error: {
    code: string
    message: string
    details?: any
    stack?: string
  }
}

export type ApiResponse<T = any> = SuccessResponse<T> | ErrorResponse

// Health and Status APIs
export interface HealthResponse {
  status: 'healthy' | 'degraded' | 'unhealthy'
  checks: HealthCheck[]
  uptime: number
  version: string
}

export interface HealthCheck {
  name: string
  status: 'pass' | 'fail' | 'warn'
  details?: any
  timestamp: string
  duration?: number
}

export interface VersionInfo {
  version: string
  gitCommit: string
  buildTime: string
  goVersion: string
  platform: string
}

// Node Management APIs
export interface NodeInfo {
  id: string
  address: string
  port: number
  status: NodeStatus
  role: NodeRole
  region?: string
  datacenter?: string
  metadata: Record<string, any>
  health: NodeHealth
  capabilities: NodeCapabilities
  resources: NodeResources
  lastSeen: string
  joinedAt: string
}

export type NodeStatus = 'online' | 'offline' | 'draining' | 'maintenance'
export type NodeRole = 'coordinator' | 'worker' | 'hybrid'

export interface NodeHealth {
  cpu: number
  memory: number
  disk: number
  network: number
  load: number[]
  errors: number
  warnings: number
}

export interface NodeCapabilities {
  maxConcurrentTasks: number
  supportedModels: string[]
  gpu: boolean
  memory: number
  storage: number
}

export interface NodeResources {
  cpu: ResourceUsage
  memory: ResourceUsage
  disk: ResourceUsage
  network: ResourceUsage
  gpu?: ResourceUsage
}

export interface ResourceUsage {
  used: number
  total: number
  percentage: number
  trend?: 'up' | 'down' | 'stable'
}

// Model Management APIs
export interface ModelInfo {
  name: string
  tag: string
  size: number
  format: string
  family: string
  parent?: string
  parameter_size: string
  quantization_level: string
  digest: string
  created_at: string
  modified_at: string
  details: ModelDetails
  sync_status: ModelSyncStatus
  distribution: ModelDistribution
}

export interface ModelDetails {
  format: string
  family: string
  families?: string[]
  parameter_size: string
  quantization_level: string
  template?: string
  system?: string
  license?: string
}

export interface ModelSyncStatus {
  status: 'synchronized' | 'syncing' | 'failed' | 'pending'
  nodes: string[]
  total_nodes: number
  synced_nodes: number
  failed_nodes: string[]
  last_sync: string
  error?: string
}

export interface ModelDistribution {
  strategy: 'replicated' | 'sharded' | 'hybrid'
  replicas: number
  shards?: number
  placement: ModelPlacement[]
}

export interface ModelPlacement {
  node_id: string
  shard?: number
  replica?: number
  status: 'active' | 'inactive' | 'syncing'
  size: number
  checksum: string
}

// Task and Transfer APIs
export interface TaskInfo {
  id: string
  type: TaskType
  status: TaskStatus
  model: string
  node_id: string
  created_at: string
  started_at?: string
  completed_at?: string
  input: any
  output?: any
  error?: string
  metadata: Record<string, any>
  metrics: TaskMetrics
}

export type TaskType = 'inference' | 'model_sync' | 'health_check' | 'maintenance'
export type TaskStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled'

export interface TaskMetrics {
  duration?: number
  tokens?: number
  memory_used?: number
  cpu_time?: number
  queue_time?: number
}

export interface TransferInfo {
  id: string
  type: 'model_sync' | 'model_pull' | 'backup' | 'migration'
  status: TransferStatus
  source: string
  destination: string
  model?: string
  progress: TransferProgress
  created_at: string
  started_at?: string
  completed_at?: string
  error?: string
}

export type TransferStatus = 'pending' | 'active' | 'completed' | 'failed' | 'cancelled'

export interface TransferProgress {
  bytes_transferred: number
  total_bytes: number
  percentage: number
  rate: number
  eta?: number
  checksum_verified?: boolean
}

// Cluster Management APIs
export interface ClusterInfo {
  id: string
  name: string
  status: ClusterStatus
  leader: string
  nodes: number
  healthy_nodes: number
  total_capacity: ClusterCapacity
  used_capacity: ClusterCapacity
  version: string
  created_at: string
  metadata: Record<string, any>
}

export type ClusterStatus = 'healthy' | 'degraded' | 'unhealthy' | 'maintenance'

export interface ClusterCapacity {
  cpu: number
  memory: number
  storage: number
  models: number
  concurrent_tasks: number
}

export interface ClusterMetrics {
  timestamp: string
  nodes: NodeMetrics[]
  cluster: ClusterAggregateMetrics
  models: ModelMetrics[]
  tasks: TaskQueueMetrics
}

export interface NodeMetrics {
  node_id: string
  cpu: number
  memory: number
  disk: number
  network: NetworkMetrics
  tasks: number
  errors: number
}

export interface NetworkMetrics {
  bytes_in: number
  bytes_out: number
  packets_in: number
  packets_out: number
  errors: number
}

export interface ClusterAggregateMetrics {
  total_nodes: number
  healthy_nodes: number
  total_tasks: number
  active_tasks: number
  completed_tasks: number
  failed_tasks: number
  avg_response_time: number
  total_models: number
  total_storage: number
}

export interface ModelMetrics {
  model: string
  usage_count: number
  avg_response_time: number
  error_rate: number
  last_used: string
}

export interface TaskQueueMetrics {
  pending: number
  running: number
  completed: number
  failed: number
  avg_queue_time: number
  avg_execution_time: number
}

// API Client Configuration
export interface ApiClientConfig {
  baseUrl: string
  timeout: number
  retries: number
  headers: Record<string, string>
  auth?: {
    type: 'bearer' | 'basic' | 'api_key'
    credentials: string
  }
}

// Request/Response utilities
export interface RequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'
  headers?: Record<string, string>
  params?: Record<string, any>
  body?: any
  timeout?: number
  retries?: number
  cache?: boolean
}

export interface PaginationParams {
  page?: number
  limit?: number
  sort?: string
  order?: 'asc' | 'desc'
  filter?: Record<string, any>
}

export interface PaginationMeta {
  page: number
  limit: number
  total: number
  totalPages: number
  hasNextPage: boolean
  hasPrevPage: boolean
}

// Download Progress for streaming operations
export interface DownloadProgress {
  status: 'downloading' | 'extracting' | 'verifying' | 'success' | 'error'
  completed: number
  total: number
  digest?: string
  error?: string
}

// Pull Request for model downloads
export interface PullRequest {
  name: string
  insecure?: boolean
  stream?: boolean
}

// Generate Request for inference
export interface GenerateRequest {
  model: string
  prompt: string
  suffix?: string
  images?: string[]
  system?: string
  template?: string
  context?: number[]
  stream?: boolean
  raw?: boolean
  format?: string
  options?: Record<string, any>
}

// Chat Request for conversation
export interface ChatRequest {
  model: string
  messages: Array<{
    role: 'system' | 'user' | 'assistant'
    content: string
  }>
  stream?: boolean
  format?: string
  options?: Record<string, any>
}

// Sync Status for distributed operations
export interface SyncStatus {
  syncing: boolean
  progress: number
  total_nodes: number
  completed_nodes: number
  failed_nodes: string[]
  error?: string
}

// Request Configuration
export interface RequestConfig {
  headers?: Record<string, string>
  timeout?: number
  retries?: number
  signal?: AbortSignal
}