/**
 * @fileoverview Cluster management API client
 * @description Handles cluster status, node management, and distributed operations
 */

import { BaseAPIClient } from './base';
import {
  NodeInfo,
  ClusterStatus,
  HealthResponse,
  LeaderInfo,
  APIResponse,
  RequestConfig,
} from '../../types/api';

export class ClusterAPI extends BaseAPIClient {
  /**
   * Get cluster health status
   */
  async getHealth(config?: RequestConfig): Promise<HealthResponse> {
    const response = await this.get<HealthResponse>('/health', config);
    return response.data!;
  }

  /**
   * Get all cluster nodes
   */
  async getNodes(config?: RequestConfig): Promise<NodeInfo[]> {
    const response = await this.get<{ nodes: NodeInfo[] }>('/api/distributed/nodes', config);
    return response.data!.nodes;
  }

  /**
   * Get cluster status and statistics
   */
  async getStatus(config?: RequestConfig): Promise<ClusterStatus & Record<string, any>> {
    const response = await this.get<ClusterStatus & Record<string, any>>(
      '/api/distributed/status',
      config
    );
    return response.data!;
  }

  /**
   * Get current cluster leader information
   */
  async getLeader(config?: RequestConfig): Promise<LeaderInfo> {
    const status = await this.getStatus(config);
    const nodes = await this.getNodes(config);
    
    const leaderNode = nodes.find(node => node.role === 'leader');
    if (!leaderNode) {
      throw new Error('No leader found in cluster');
    }

    return {
      node_id: leaderNode.id,
      address: leaderNode.address,
      elected_at: '', // This would need to be added to the backend
      term: 0, // This would need to be added to the backend
    };
  }

  /**
   * Get specific node information
   */
  async getNode(nodeId: string, config?: RequestConfig): Promise<NodeInfo> {
    const nodes = await this.getNodes(config);
    const node = nodes.find(n => n.id === nodeId);
    
    if (!node) {
      throw new Error(`Node ${nodeId} not found`);
    }
    
    return node;
  }

  /**
   * Drain a node (prepare for maintenance)
   */
  async drainNode(nodeId: string, config?: RequestConfig): Promise<{ message: string }> {
    const response = await this.post<{ message: string }>(
      `/admin/nodes/${nodeId}/drain`,
      undefined,
      {
        ...config,
        headers: {
          Authorization: `Bearer ${process.env.OLLAMA_ADMIN_TOKEN}`,
          ...config?.headers,
        },
      }
    );
    return response.data!;
  }

  /**
   * Remove drain status from a node
   */
  async undrainNode(nodeId: string, config?: RequestConfig): Promise<{ message: string }> {
    const response = await this.delete<{ message: string }>(
      `/admin/nodes/${nodeId}/drain`,
      {
        ...config,
        headers: {
          Authorization: `Bearer ${process.env.OLLAMA_ADMIN_TOKEN}`,
          ...config?.headers,
        },
      }
    );
    return response.data!;
  }

  /**
   * Force cluster rebalancing
   */
  async rebalance(config?: RequestConfig): Promise<{ message: string }> {
    const response = await this.post<{ message: string }>('/api/distributed/rebalance', undefined, config);
    return response.data!;
  }

  /**
   * Migrate model between nodes
   */
  async migrateModel(
    modelName: string,
    fromNode: string,
    toNode: string,
    config?: RequestConfig
  ): Promise<{ message: string }> {
    const response = await this.post<{ message: string }>(
      '/api/distributed/migrate',
      {
        model_name: modelName,
        from_node: fromNode,
        to_node: toNode,
      },
      config
    );
    return response.data!;
  }

  /**
   * Get cluster metrics
   */
  async getMetrics(config?: RequestConfig): Promise<Record<string, any>> {
    const response = await this.get<Record<string, any>>('/metrics', config);
    return response.data!;
  }

  /**
   * Set cluster mode (admin only)
   */
  async setMode(
    mode: 'distributed' | 'local',
    config?: RequestConfig
  ): Promise<{ message: string }> {
    const response = await this.post<{ message: string }>(
      '/admin/mode',
      { mode },
      {
        ...config,
        headers: {
          Authorization: `Bearer ${process.env.OLLAMA_ADMIN_TOKEN}`,
          ...config?.headers,
        },
      }
    );
    return response.data!;
  }

  /**
   * Set fallback mode (admin only)
   */
  async setFallbackMode(
    enabled: boolean,
    config?: RequestConfig
  ): Promise<{ message: string }> {
    const response = await this.post<{ message: string }>(
      '/admin/fallback',
      { enabled },
      {
        ...config,
        headers: {
          Authorization: `Bearer ${process.env.OLLAMA_ADMIN_TOKEN}`,
          ...config?.headers,
        },
      }
    );
    return response.data!;
  }

  /**
   * Force rebalance (admin only)
   */
  async forceRebalance(config?: RequestConfig): Promise<{ message: string }> {
    const response = await this.post<{ message: string }>(
      '/admin/rebalance',
      undefined,
      {
        ...config,
        headers: {
          Authorization: `Bearer ${process.env.OLLAMA_ADMIN_TOKEN}`,
          ...config?.headers,
        },
      }
    );
    return response.data!;
  }

  /**
   * Get detailed cluster statistics (admin only)
   */
  async getStats(config?: RequestConfig): Promise<Record<string, any>> {
    const response = await this.get<Record<string, any>>(
      '/admin/stats',
      {
        ...config,
        headers: {
          Authorization: `Bearer ${process.env.OLLAMA_ADMIN_TOKEN}`,
          ...config?.headers,
        },
      }
    );
    return response.data!;
  }

  /**
   * Monitor cluster status in real-time
   */
  async *monitorStatus(intervalMs: number = 5000): AsyncGenerator<ClusterStatus & Record<string, any>> {
    while (true) {
      try {
        const status = await this.getStatus();
        yield status;
        await new Promise(resolve => setTimeout(resolve, intervalMs));
      } catch (error) {
        console.error('Error monitoring cluster status:', error);
        await new Promise(resolve => setTimeout(resolve, intervalMs * 2));
      }
    }
  }
}