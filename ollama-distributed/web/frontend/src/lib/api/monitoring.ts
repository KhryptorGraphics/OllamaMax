/**
 * @fileoverview Monitoring and metrics API client
 * @description Handles system metrics, alerts, and performance monitoring
 */

import { BaseAPIClient } from './base';
import {
  MetricsData,
  Alert,
  SystemStats,
  PerformanceMetrics,
  APIResponse,
  RequestConfig,
} from '../../types/api';

export class MonitoringAPI extends BaseAPIClient {
  /**
   * Get current system metrics
   */
  async getMetrics(nodeId?: string, config?: RequestConfig): Promise<MetricsData> {
    const endpoint = nodeId ? `/metrics?node_id=${nodeId}` : '/metrics';
    const response = await this.get<MetricsData>(endpoint, config);
    return response.data!;
  }

  /**
   * Get system statistics
   */
  async getStats(config?: RequestConfig): Promise<SystemStats> {
    const response = await this.get<SystemStats>('/admin/stats', {
      ...config,
      headers: {
        Authorization: `Bearer ${process.env.OLLAMA_ADMIN_TOKEN}`,
        ...config?.headers,
      },
    });
    return response.data!;
  }

  /**
   * Get performance metrics
   */
  async getPerformanceMetrics(
    timeRange?: '1h' | '24h' | '7d' | '30d',
    config?: RequestConfig
  ): Promise<PerformanceMetrics> {
    const params = timeRange ? `?range=${timeRange}` : '';
    const response = await this.get<PerformanceMetrics>(`/api/metrics/performance${params}`, config);
    return response.data!;
  }

  /**
   * Get historical metrics
   */
  async getHistoricalMetrics(
    startTime: string,
    endTime: string,
    resolution?: '1m' | '5m' | '1h' | '1d',
    config?: RequestConfig
  ): Promise<MetricsData[]> {
    const params = new URLSearchParams({
      start: startTime,
      end: endTime,
      ...(resolution && { resolution }),
    });
    
    const response = await this.get<{ metrics: MetricsData[] }>(
      `/api/metrics/historical?${params}`,
      config
    );
    return response.data!.metrics;
  }

  /**
   * Get active alerts
   */
  async getAlerts(resolved?: boolean, config?: RequestConfig): Promise<Alert[]> {
    const params = resolved !== undefined ? `?resolved=${resolved}` : '';
    const response = await this.get<{ alerts: Alert[] }>(`/api/alerts${params}`, config);
    return response.data!.alerts;
  }

  /**
   * Create new alert
   */
  async createAlert(alert: Omit<Alert, 'id' | 'timestamp'>, config?: RequestConfig): Promise<Alert> {
    const response = await this.post<Alert>('/api/alerts', alert, config);
    return response.data!;
  }

  /**
   * Resolve alert
   */
  async resolveAlert(alertId: string, config?: RequestConfig): Promise<{ message: string }> {
    const response = await this.patch<{ message: string }>(
      `/api/alerts/${alertId}/resolve`,
      undefined,
      config
    );
    return response.data!;
  }

  /**
   * Delete alert
   */
  async deleteAlert(alertId: string, config?: RequestConfig): Promise<{ message: string }> {
    const response = await this.delete<{ message: string }>(`/api/alerts/${alertId}`, config);
    return response.data!;
  }

  /**
   * Get node-specific metrics
   */
  async getNodeMetrics(nodeId: string, config?: RequestConfig): Promise<{
    cpu_usage: number;
    memory_usage: number;
    disk_usage: number;
    gpu_usage?: number;
    network_io: { in: number; out: number };
    model_count: number;
    active_requests: number;
  }> {
    const response = await this.get<{
      cpu_usage: number;
      memory_usage: number;
      disk_usage: number;
      gpu_usage?: number;
      network_io: { in: number; out: number };
      model_count: number;
      active_requests: number;
    }>(`/api/nodes/${nodeId}/metrics`, config);
    return response.data!;
  }

  /**
   * Get cluster-wide performance overview
   */
  async getClusterPerformance(config?: RequestConfig): Promise<{
    total_requests: number;
    average_response_time: number;
    error_rate: number;
    throughput: number;
    active_models: number;
    healthy_nodes: number;
    total_nodes: number;
  }> {
    const response = await this.get<{
      total_requests: number;
      average_response_time: number;
      error_rate: number;
      throughput: number;
      active_models: number;
      healthy_nodes: number;
      total_nodes: number;
    }>('/api/cluster/performance', config);
    return response.data!;
  }

  /**
   * Get model performance metrics
   */
  async getModelMetrics(modelName: string, config?: RequestConfig): Promise<{
    inference_count: number;
    average_latency: number;
    tokens_per_second: number;
    memory_usage: number;
    error_rate: number;
    last_used: string;
  }> {
    const response = await this.get<{
      inference_count: number;
      average_latency: number;
      tokens_per_second: number;
      memory_usage: number;
      error_rate: number;
      last_used: string;
    }>(`/api/models/${encodeURIComponent(modelName)}/metrics`, config);
    return response.data!;
  }

  /**
   * Export metrics data
   */
  async exportMetrics(
    format: 'json' | 'csv' | 'prometheus',
    startTime?: string,
    endTime?: string,
    config?: RequestConfig
  ): Promise<string | object> {
    const params = new URLSearchParams({ format });
    if (startTime) params.append('start', startTime);
    if (endTime) params.append('end', endTime);

    const response = await this.get<string | object>(
      `/api/metrics/export?${params}`,
      config
    );
    return response.data!;
  }

  /**
   * Get system health check
   */
  async getHealthCheck(config?: RequestConfig): Promise<{
    status: 'healthy' | 'degraded' | 'critical';
    checks: Record<string, {
      status: 'pass' | 'fail' | 'warn';
      message?: string;
      duration: number;
    }>;
    timestamp: number;
  }> {
    const response = await this.get<{
      status: 'healthy' | 'degraded' | 'critical';
      checks: Record<string, {
        status: 'pass' | 'fail' | 'warn';
        message?: string;
        duration: number;
      }>;
      timestamp: number;
    }>('/health/detailed', config);
    return response.data!;
  }

  /**
   * Get resource utilization trends
   */
  async getResourceTrends(
    resource: 'cpu' | 'memory' | 'disk' | 'network',
    period: '1h' | '24h' | '7d',
    config?: RequestConfig
  ): Promise<Array<{ timestamp: number; value: number }>> {
    const response = await this.get<{ trends: Array<{ timestamp: number; value: number }> }>(
      `/api/metrics/trends/${resource}?period=${period}`,
      config
    );
    return response.data!.trends;
  }

  /**
   * Set alert threshold
   */
  async setAlertThreshold(
    metric: string,
    threshold: number,
    condition: 'gt' | 'lt' | 'eq',
    config?: RequestConfig
  ): Promise<{ message: string }> {
    const response = await this.post<{ message: string }>(
      '/api/alerts/thresholds',
      { metric, threshold, condition },
      config
    );
    return response.data!;
  }

  /**
   * Stream real-time metrics
   */
  async *streamMetrics(config?: RequestConfig): AsyncGenerator<MetricsData> {
    const stream = await this.stream('/api/metrics/stream', {
      ...config,
      headers: {
        Accept: 'text/event-stream',
        'Cache-Control': 'no-cache',
        ...config?.headers,
      },
    });

    const reader = stream.getReader();
    const decoder = new TextDecoder();

    try {
      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        const chunk = decoder.decode(value);
        const lines = chunk.split('\n').filter(line => line.trim());

        for (const line of lines) {
          if (line.startsWith('data: ')) {
            try {
              const data: MetricsData = JSON.parse(line.substring(6));
              yield data;
            } catch (error) {
              console.warn('Failed to parse metrics data:', line);
            }
          }
        }
      }
    } finally {
      reader.releaseLock();
    }
  }

  /**
   * Monitor alerts in real-time
   */
  async *monitorAlerts(config?: RequestConfig): AsyncGenerator<Alert> {
    while (true) {
      try {
        const alerts = await this.getAlerts(false, config);
        const newAlerts = alerts.filter(alert => 
          Date.now() - new Date(alert.timestamp).getTime() < 60000 // Last minute
        );
        
        for (const alert of newAlerts) {
          yield alert;
        }
        
        await new Promise(resolve => setTimeout(resolve, 10000)); // Check every 10 seconds
      } catch (error) {
        console.error('Error monitoring alerts:', error);
        await new Promise(resolve => setTimeout(resolve, 30000)); // Wait longer on error
      }
    }
  }

  /**
   * Get dashboard metrics summary
   */
  async getDashboardMetrics(config?: RequestConfig): Promise<{
    cluster: {
      status: string;
      nodes: number;
      healthy_nodes: number;
      total_models: number;
    };
    performance: {
      requests_per_second: number;
      average_latency: number;
      error_rate: number;
    };
    resources: {
      cpu_usage: number;
      memory_usage: number;
      disk_usage: number;
    };
    alerts: {
      critical: number;
      warning: number;
      total: number;
    };
  }> {
    const response = await this.get<{
      cluster: {
        status: string;
        nodes: number;
        healthy_nodes: number;
        total_models: number;
      };
      performance: {
        requests_per_second: number;
        average_latency: number;
        error_rate: number;
      };
      resources: {
        cpu_usage: number;
        memory_usage: number;
        disk_usage: number;
      };
      alerts: {
        critical: number;
        warning: number;
        total: number;
      };
    }>('/api/dashboard/metrics', config);
    return response.data!;
  }
}