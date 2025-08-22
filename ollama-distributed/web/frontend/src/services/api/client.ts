// Enhanced API client with full backend integration
import type {
  ApiResponse,
  ApiClientConfig,
  RequestOptions,
  PaginationParams,
  HealthResponse,
  VersionInfo,
  NodeInfo,
  ModelInfo,
  TaskInfo,
  TransferInfo,
  ClusterInfo,
  ClusterMetrics,
} from '@/types/api'

export class ApiClient {
  private config: ApiClientConfig
  private abortControllers: Map<string, AbortController> = new Map()

  constructor(config: Partial<ApiClientConfig> = {}) {
    this.config = {
      baseUrl: config.baseUrl || '/api/v1',
      timeout: config.timeout || 30000,
      retries: config.retries || 3,
      headers: {
        'Content-Type': 'application/json',
        ...config.headers,
      },
      ...config,
    }
  }

  // Core request method with retry logic and error handling
  private async request<T = any>(
    endpoint: string,
    options: RequestOptions = {}
  ): Promise<ApiResponse<T>> {
    const {
      method = 'GET',
      headers = {},
      params = {},
      body,
      timeout = this.config.timeout,
      retries = this.config.retries,
      cache = false,
    } = options

    const url = new URL(`${this.config.baseUrl}${endpoint}`)
    
    // Add query parameters
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        url.searchParams.append(key, String(value))
      }
    })

    const requestId = Math.random().toString(36).substr(2, 9)
    const abortController = new AbortController()
    this.abortControllers.set(requestId, abortController)

    const requestOptions: RequestInit = {
      method,
      headers: {
        ...this.config.headers,
        ...headers,
      },
      signal: abortController.signal,
    }

    if (body && method !== 'GET') {
      requestOptions.body = typeof body === 'string' ? body : JSON.stringify(body)
    }

    // Add authentication if available
    if (this.config.auth) {
      const authHeader = this.getAuthHeader()
      if (authHeader) {
        (requestOptions.headers as Record<string, string>)[authHeader.name] = authHeader.value
      }
    }

    let lastError: Error | null = null
    
    for (let attempt = 0; attempt <= retries; attempt++) {
      try {
        const timeoutId = setTimeout(() => {
          abortController.abort()
        }, timeout)

        const response = await fetch(url.toString(), requestOptions)
        clearTimeout(timeoutId)

        const responseData = await response.json()

        if (!response.ok) {
          throw new Error(responseData.error?.message || `HTTP ${response.status}`)
        }

        this.abortControllers.delete(requestId)
        return responseData as ApiResponse<T>

      } catch (error) {
        lastError = error as Error
        
        if (error instanceof Error && error.name === 'AbortError') {
          throw new Error('Request timeout')
        }

        if (attempt === retries) {
          break
        }

        // Exponential backoff
        await this.delay(Math.pow(2, attempt) * 1000)
      }
    }

    this.abortControllers.delete(requestId)
    throw lastError || new Error('Request failed')
  }

  private getAuthHeader(): { name: string; value: string } | null {
    if (!this.config.auth) return null

    switch (this.config.auth.type) {
      case 'bearer':
        return { name: 'Authorization', value: `Bearer ${this.config.auth.credentials}` }
      case 'basic':
        return { name: 'Authorization', value: `Basic ${this.config.auth.credentials}` }
      case 'api_key':
        return { name: 'X-API-Key', value: this.config.auth.credentials }
      default:
        return null
    }
  }

  private delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms))
  }

  // Update configuration
  updateConfig(updates: Partial<ApiClientConfig>): void {
    this.config = { ...this.config, ...updates }
  }

  // Cancel all pending requests
  cancelAllRequests(): void {
    this.abortControllers.forEach(controller => controller.abort())
    this.abortControllers.clear()
  }

  // Health and System APIs
  async getHealth(): Promise<ApiResponse<HealthResponse>> {
    return this.request<HealthResponse>('/health')
  }

  async getVersion(): Promise<ApiResponse<VersionInfo>> {
    return this.request<VersionInfo>('/version')
  }

  // Node Management APIs
  async getNodes(params?: PaginationParams): Promise<ApiResponse<NodeInfo[]>> {
    return this.request<NodeInfo[]>('/nodes', { params })
  }

  async getNode(nodeId: string): Promise<ApiResponse<NodeInfo>> {
    return this.request<NodeInfo>(`/nodes/${nodeId}`)
  }

  async updateNode(nodeId: string, data: Partial<NodeInfo>): Promise<ApiResponse<NodeInfo>> {
    return this.request<NodeInfo>(`/nodes/${nodeId}`, {
      method: 'PUT',
      body: data,
    })
  }

  async removeNode(nodeId: string): Promise<ApiResponse<void>> {
    return this.request<void>(`/nodes/${nodeId}`, { method: 'DELETE' })
  }

  async drainNode(nodeId: string): Promise<ApiResponse<void>> {
    return this.request<void>(`/nodes/${nodeId}/drain`, { method: 'POST' })
  }

  // Model Management APIs
  async getModels(params?: PaginationParams): Promise<ApiResponse<ModelInfo[]>> {
    return this.request<ModelInfo[]>('/models', { params })
  }

  async getModel(modelName: string): Promise<ApiResponse<ModelInfo>> {
    return this.request<ModelInfo>(`/models/${encodeURIComponent(modelName)}`)
  }

  async pullModel(modelName: string, tag?: string): Promise<ApiResponse<void>> {
    return this.request<void>('/models/pull', {
      method: 'POST',
      body: { name: modelName, tag },
    })
  }

  async deleteModel(modelName: string, tag?: string): Promise<ApiResponse<void>> {
    const name = tag ? `${modelName}:${tag}` : modelName
    return this.request<void>(`/models/${encodeURIComponent(name)}`, {
      method: 'DELETE',
    })
  }

  async syncModel(modelName: string, nodes?: string[]): Promise<ApiResponse<void>> {
    return this.request<void>(`/models/${encodeURIComponent(modelName)}/sync`, {
      method: 'POST',
      body: { nodes },
    })
  }

  async getModelSyncStatus(modelName: string): Promise<ApiResponse<any>> {
    return this.request(`/models/${encodeURIComponent(modelName)}/sync/status`)
  }

  // Task Management APIs
  async getTasks(params?: PaginationParams): Promise<ApiResponse<TaskInfo[]>> {
    return this.request<TaskInfo[]>('/tasks', { params })
  }

  async getTask(taskId: string): Promise<ApiResponse<TaskInfo>> {
    return this.request<TaskInfo>(`/tasks/${taskId}`)
  }

  async cancelTask(taskId: string): Promise<ApiResponse<void>> {
    return this.request<void>(`/tasks/${taskId}/cancel`, { method: 'POST' })
  }

  async getTaskQueue(): Promise<ApiResponse<TaskInfo[]>> {
    return this.request<TaskInfo[]>('/tasks/queue')
  }

  // Transfer Management APIs
  async getTransfers(params?: PaginationParams): Promise<ApiResponse<TransferInfo[]>> {
    return this.request<TransferInfo[]>('/transfers', { params })
  }

  async getTransfer(transferId: string): Promise<ApiResponse<TransferInfo>> {
    return this.request<TransferInfo>(`/transfers/${transferId}`)
  }

  async cancelTransfer(transferId: string): Promise<ApiResponse<void>> {
    return this.request<void>(`/transfers/${transferId}`, { method: 'DELETE' })
  }

  // Cluster Management APIs
  async getClusterStatus(): Promise<ApiResponse<ClusterInfo>> {
    return this.request<ClusterInfo>('/cluster/status')
  }

  async getClusterLeader(): Promise<ApiResponse<{ leader: string }>> {
    return this.request<{ leader: string }>('/cluster/leader')
  }

  async getClusterMetrics(): Promise<ApiResponse<ClusterMetrics>> {
    return this.request<ClusterMetrics>('/metrics')
  }

  // Performance Monitoring APIs
  async getResourceMetrics(): Promise<ApiResponse<any>> {
    return this.request('/metrics/resources')
  }

  async getPerformanceMetrics(): Promise<ApiResponse<any>> {
    return this.request('/metrics/performance')
  }

  async getPerformanceReport(): Promise<ApiResponse<any>> {
    return this.request('/performance/report')
  }

  async getBottlenecks(): Promise<ApiResponse<any>> {
    return this.request('/performance/bottlenecks')
  }

  async getOptimizations(): Promise<ApiResponse<any>> {
    return this.request('/performance/optimizations')
  }

  // Security APIs
  async getSecurityStatus(): Promise<ApiResponse<any>> {
    return this.request('/security/status')
  }

  async getThreats(): Promise<ApiResponse<any>> {
    return this.request('/security/threats')
  }

  async getSecurityAlerts(): Promise<ApiResponse<any>> {
    return this.request('/security/alerts')
  }

  async getAuditLog(params?: PaginationParams): Promise<ApiResponse<any>> {
    return this.request('/security/audit', { params })
  }

  // Dashboard APIs - Sprint C specific endpoints
  async getDashboardSummary(): Promise<ApiResponse<any>> {
    return this.request('/dashboard/summary')
  }

  async getDashboardActivity(): Promise<ApiResponse<any[]>> {
    return this.request('/dashboard/activity')
  }

  // Alert Management APIs  
  async getAlerts(params?: PaginationParams): Promise<ApiResponse<any[]>> {
    return this.request('/alerts', { params })
  }

  async acknowledgeAlert(alertId: string): Promise<ApiResponse<void>> {
    return this.request<void>(`/alerts/${alertId}/acknowledge`, { method: 'POST' })
  }

  async resolveAlert(alertId: string): Promise<ApiResponse<void>> {
    return this.request<void>(`/alerts/${alertId}/resolve`, { method: 'POST' })
  }

  async createAlert(alert: any): Promise<ApiResponse<any>> {
    return this.request('/alerts', {
      method: 'POST',
      body: alert,
    })
  }

  // Model deployment specific APIs
  async deployModel(modelName: string, options?: { nodes?: string[] }): Promise<ApiResponse<void>> {
    return this.request<void>(`/models/${encodeURIComponent(modelName)}/deploy`, {
      method: 'POST',
      body: options,
    })
  }

  async undeployModel(modelName: string, options?: { nodes?: string[] }): Promise<ApiResponse<void>> {
    return this.request<void>(`/models/${encodeURIComponent(modelName)}/undeploy`, {
      method: 'POST',
      body: options,
    })
  }

  async uploadModel(formData: FormData): Promise<ApiResponse<any>> {
    return this.request('/models/upload', {
      method: 'POST',
      body: formData,
      headers: {}, // Let browser set Content-Type for multipart/form-data
    })
  }

  // Node operation specific APIs
  async enableNode(nodeId: string): Promise<ApiResponse<void>> {
    return this.request<void>(`/nodes/${nodeId}/enable`, { method: 'POST' })
  }

  async getNodeMetrics(nodeId: string): Promise<ApiResponse<any>> {
    return this.request(`/nodes/${nodeId}/metrics`)
  }

  // Inference API
  async runInference(data: {
    model: string
    prompt: string
    stream?: boolean
    options?: Record<string, any>
  }): Promise<ApiResponse<any>> {
    return this.request('/inference', {
      method: 'POST',
      body: data,
    })
  }

  // Streaming inference with Server-Sent Events
  async streamInference(
    data: {
      model: string
      prompt: string
      options?: Record<string, any>
    },
    onMessage: (chunk: any) => void,
    onError?: (error: Error) => void,
    onComplete?: () => void
  ): Promise<void> {
    const url = `${this.config.baseUrl}/inference/stream`
    const eventSource = new EventSource(url, {
      withCredentials: true,
    })

    // Send request data via POST to start streaming
    const response = await fetch(url, {
      method: 'POST',
      headers: this.config.headers,
      body: JSON.stringify(data),
    })

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`)
    }

    eventSource.onmessage = (event) => {
      try {
        const chunk = JSON.parse(event.data)
        onMessage(chunk)
      } catch (error) {
        onError?.(error as Error)
      }
    }

    eventSource.onerror = (error) => {
      eventSource.close()
      onError?.(new Error('Streaming error'))
    }

    eventSource.addEventListener('end', () => {
      eventSource.close()
      onComplete?.()
    })
  }
}

// Create singleton instance
export const apiClient = new ApiClient({
  baseUrl: import.meta.env.VITE_API_BASE_URL || '/api/v1',
})

export default apiClient