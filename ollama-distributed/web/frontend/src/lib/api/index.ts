/**
 * @fileoverview Unified API client for OllamaMax
 * @description Main API client that combines all service clients
 */

import { BaseAPIClient } from './base';
import { AuthAPI } from './auth';
import { ClusterAPI } from './cluster';
import { ModelsAPI } from './models';
import { MonitoringAPI } from './monitoring';
import { SecurityAPI } from './security';
import { NotificationsAPI } from './notifications';
import { APIClientConfig } from '../../types/api';

// Default configuration
const DEFAULT_CONFIG: APIClientConfig = {
  baseURL: typeof window !== 'undefined' 
    ? `${window.location.protocol}//${window.location.host}`
    : 'http://localhost:8080',
  timeout: 30000,
  retries: 3,
  retryDelay: 1000,
};

/**
 * Main OllamaMax API client
 * Provides access to all backend services through a unified interface
 */
export class OllamaMaxAPI {
  public readonly auth: AuthAPI;
  public readonly cluster: ClusterAPI;
  public readonly models: ModelsAPI;
  public readonly monitoring: MonitoringAPI;
  public readonly security: SecurityAPI;
  public readonly notifications: NotificationsAPI;

  private config: APIClientConfig;

  constructor(config: Partial<APIClientConfig> = {}) {
    this.config = { ...DEFAULT_CONFIG, ...config };

    // Initialize all service clients
    this.auth = new AuthAPI(this.config);
    this.cluster = new ClusterAPI(this.config);
    this.models = new ModelsAPI(this.config);
    this.monitoring = new MonitoringAPI(this.config);
    this.security = new SecurityAPI(this.config);
    this.notifications = new NotificationsAPI(this.config);
  }

  /**
   * Update configuration for all clients
   */
  updateConfig(newConfig: Partial<APIClientConfig>): void {
    this.config = { ...this.config, ...newConfig };
    
    // Update all client configurations
    Object.values(this).forEach(client => {
      if (client instanceof BaseAPIClient) {
        // Note: Would need to add updateConfig method to BaseAPIClient
        console.warn('Config update not implemented for individual clients');
      }
    });
  }

  /**
   * Get current configuration
   */
  getConfig(): APIClientConfig {
    return { ...this.config };
  }

  /**
   * Test connection to backend
   */
  async testConnection(): Promise<{
    status: 'success' | 'error';
    latency: number;
    services: Record<string, boolean>;
  }> {
    const startTime = Date.now();
    const results: Record<string, boolean> = {};

    try {
      // Test health endpoint
      const health = await this.cluster.getHealth({ timeout: 5000 });
      results.cluster = health.status === 'healthy';
    } catch {
      results.cluster = false;
    }

    try {
      // Test auth service
      await this.auth.checkHealth({ timeout: 5000 });
      results.auth = true;
    } catch {
      results.auth = false;
    }

    try {
      // Test models service
      await this.models.getVersion({ timeout: 5000 });
      results.models = true;
    } catch {
      results.models = false;
    }

    const latency = Date.now() - startTime;
    const allServicesUp = Object.values(results).every(status => status);

    return {
      status: allServicesUp ? 'success' : 'error',
      latency,
      services: results
    };
  }

  /**
   * Get comprehensive system status
   */
  async getSystemStatus(): Promise<{
    cluster: any;
    models: any;
    performance: any;
    alerts: any;
    timestamp: number;
  }> {
    const [cluster, models, performance, alerts] = await Promise.allSettled([
      this.cluster.getStatus(),
      this.models.list(),
      this.monitoring.getPerformanceMetrics(),
      this.monitoring.getAlerts(false)
    ]);

    return {
      cluster: cluster.status === 'fulfilled' ? cluster.value : null,
      models: models.status === 'fulfilled' ? models.value : null,
      performance: performance.status === 'fulfilled' ? performance.value : null,
      alerts: alerts.status === 'fulfilled' ? alerts.value : null,
      timestamp: Date.now()
    };
  }

  /**
   * Initialize WebSocket connection for real-time updates
   */
  async connectWebSocket(token?: string): Promise<void> {
    return this.notifications.connectWebSocket(token);
  }

  /**
   * Disconnect WebSocket
   */
  disconnectWebSocket(): void {
    this.notifications.disconnectWebSocket();
  }

  /**
   * Subscribe to real-time updates
   */
  subscribeToUpdates(handlers: {
    onNotification?: (notification: any) => void;
    onClusterUpdate?: (status: any) => void;
    onModelUpdate?: (update: any) => void;
    onAlert?: (alert: any) => void;
  }): () => void {
    const unsubscribers: Array<() => void> = [];

    if (handlers.onNotification) {
      unsubscribers.push(this.notifications.subscribeToNotifications(handlers.onNotification));
    }

    if (handlers.onClusterUpdate) {
      unsubscribers.push(this.notifications.subscribeToClusterUpdates(handlers.onClusterUpdate));
    }

    if (handlers.onModelUpdate) {
      unsubscribers.push(this.notifications.subscribeToModelUpdates(handlers.onModelUpdate));
    }

    if (handlers.onAlert) {
      unsubscribers.push(this.notifications.subscribeToAlerts(handlers.onAlert));
    }

    // Return cleanup function
    return () => {
      unsubscribers.forEach(unsubscribe => unsubscribe());
    };
  }
}

// Create singleton instance
let apiClient: OllamaMaxAPI | null = null;

/**
 * Get or create the singleton API client instance
 */
export function getAPIClient(config?: Partial<APIClientConfig>): OllamaMaxAPI {
  if (!apiClient || config) {
    apiClient = new OllamaMaxAPI(config);
  }
  return apiClient;
}

/**
 * Create a new API client instance (non-singleton)
 */
export function createAPIClient(config?: Partial<APIClientConfig>): OllamaMaxAPI {
  return new OllamaMaxAPI(config);
}

// Export individual clients for direct use
export {
  AuthAPI,
  ClusterAPI,
  ModelsAPI,
  MonitoringAPI,
  SecurityAPI,
  NotificationsAPI,
  BaseAPIClient
};

// Export types
export type { APIClientConfig } from '../../types/api';

// Default export
export default OllamaMaxAPI;
