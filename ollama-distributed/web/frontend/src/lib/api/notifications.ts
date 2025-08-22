/**
 * @fileoverview Notifications and real-time updates API client
 * @description Handles notifications, WebSocket connections, and real-time updates
 */

import { BaseAPIClient } from './base';
import {
  Notification,
  WebSocketMessage,
  PaginatedResponse,
  APIResponse,
  RequestConfig,
} from '../../types/api';

export class NotificationsAPI extends BaseAPIClient {
  private wsConnection: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private messageHandlers: Map<string, (message: WebSocketMessage) => void> = new Map();

  /**
   * Get user notifications
   */
  async getNotifications(
    filters?: {
      read?: boolean;
      type?: 'info' | 'success' | 'warning' | 'error';
      limit?: number;
      offset?: number;
    },
    config?: RequestConfig
  ): Promise<PaginatedResponse<Notification>> {
    const params = new URLSearchParams();
    if (filters) {
      Object.entries(filters).forEach(([key, value]) => {
        if (value !== undefined) {
          params.append(key, value.toString());
        }
      });
    }

    const endpoint = `/api/notifications${params.toString() ? `?${params}` : ''}`;
    const response = await this.get<PaginatedResponse<Notification>>(endpoint, config);
    return response.data!;
  }

  /**
   * Get specific notification
   */
  async getNotification(notificationId: string, config?: RequestConfig): Promise<Notification> {
    const response = await this.get<Notification>(`/api/notifications/${notificationId}`, config);
    return response.data!;
  }

  /**
   * Create notification
   */
  async createNotification(
    notification: Omit<Notification, 'id' | 'timestamp' | 'read'>,
    config?: RequestConfig
  ): Promise<Notification> {
    const response = await this.post<Notification>('/api/notifications', notification, config);
    return response.data!;
  }

  /**
   * Mark notification as read
   */
  async markAsRead(notificationId: string, config?: RequestConfig): Promise<{ message: string }> {
    const response = await this.patch<{ message: string }>(
      `/api/notifications/${notificationId}/read`,
      undefined,
      config
    );
    return response.data!;
  }

  /**
   * Mark notification as unread
   */
  async markAsUnread(notificationId: string, config?: RequestConfig): Promise<{ message: string }> {
    const response = await this.patch<{ message: string }>(
      `/api/notifications/${notificationId}/unread`,
      undefined,
      config
    );
    return response.data!;
  }

  /**
   * Mark all notifications as read
   */
  async markAllAsRead(config?: RequestConfig): Promise<{ message: string; count: number }> {
    const response = await this.patch<{ message: string; count: number }>(
      '/api/notifications/mark-all-read',
      undefined,
      config
    );
    return response.data!;
  }

  /**
   * Delete notification
   */
  async deleteNotification(notificationId: string, config?: RequestConfig): Promise<{ message: string }> {
    const response = await this.delete<{ message: string }>(`/api/notifications/${notificationId}`, config);
    return response.data!;
  }

  /**
   * Delete all read notifications
   */
  async deleteAllRead(config?: RequestConfig): Promise<{ message: string; count: number }> {
    const response = await this.delete<{ message: string; count: number }>(
      '/api/notifications/read',
      config
    );
    return response.data!;
  }

  /**
   * Get notification statistics
   */
  async getNotificationStats(config?: RequestConfig): Promise<{
    total: number;
    unread: number;
    by_type: Record<string, number>;
    recent: number;
  }> {
    const response = await this.get<{
      total: number;
      unread: number;
      by_type: Record<string, number>;
      recent: number;
    }>('/api/notifications/stats', config);
    return response.data!;
  }

  /**
   * Get notification preferences
   */
  async getPreferences(config?: RequestConfig): Promise<{
    email_notifications: boolean;
    push_notifications: boolean;
    categories: Record<string, boolean>;
    quiet_hours: {
      enabled: boolean;
      start: string;
      end: string;
    };
  }> {
    const response = await this.get<{
      email_notifications: boolean;
      push_notifications: boolean;
      categories: Record<string, boolean>;
      quiet_hours: {
        enabled: boolean;
        start: string;
        end: string;
      };
    }>('/api/notifications/preferences', config);
    return response.data!;
  }

  /**
   * Update notification preferences
   */
  async updatePreferences(
    preferences: Partial<{
      email_notifications: boolean;
      push_notifications: boolean;
      categories: Record<string, boolean>;
      quiet_hours: {
        enabled: boolean;
        start: string;
        end: string;
      };
    }>,
    config?: RequestConfig
  ): Promise<{ message: string }> {
    const response = await this.put<{ message: string }>(
      '/api/notifications/preferences',
      preferences,
      config
    );
    return response.data!;
  }

  // WebSocket Methods

  /**
   * Connect to WebSocket for real-time notifications
   */
  async connectWebSocket(token?: string): Promise<void> {
    if (this.wsConnection?.readyState === WebSocket.OPEN) {
      return;
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws/notifications`;
    
    return new Promise((resolve, reject) => {
      this.wsConnection = new WebSocket(wsUrl);

      this.wsConnection.onopen = () => {
        console.log('WebSocket connected');
        this.reconnectAttempts = 0;
        
        // Send authentication if token provided
        if (token) {
          this.sendMessage('auth', { token });
        }
        
        resolve();
      };

      this.wsConnection.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data);
          this.handleWebSocketMessage(message);
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error);
        }
      };

      this.wsConnection.onclose = (event) => {
        console.log('WebSocket disconnected:', event.code, event.reason);
        this.attemptReconnect();
      };

      this.wsConnection.onerror = (error) => {
        console.error('WebSocket error:', error);
        reject(error);
      };
    });
  }

  /**
   * Disconnect WebSocket
   */
  disconnectWebSocket(): void {
    if (this.wsConnection) {
      this.wsConnection.close();
      this.wsConnection = null;
    }
    this.messageHandlers.clear();
  }

  /**
   * Send message through WebSocket
   */
  sendMessage(type: string, payload: any): void {
    if (this.wsConnection?.readyState === WebSocket.OPEN) {
      const message: WebSocketMessage = {
        type,
        payload,
        timestamp: new Date().toISOString(),
      };
      this.wsConnection.send(JSON.stringify(message));
    } else {
      console.warn('WebSocket not connected');
    }
  }

  /**
   * Subscribe to specific message types
   */
  subscribe(messageType: string, handler: (message: WebSocketMessage) => void): () => void {
    this.messageHandlers.set(messageType, handler);
    
    // Return unsubscribe function
    return () => {
      this.messageHandlers.delete(messageType);
    };
  }

  /**
   * Subscribe to notification updates
   */
  subscribeToNotifications(handler: (notification: Notification) => void): () => void {
    return this.subscribe('notification', (message) => {
      if (message.payload && typeof message.payload === 'object') {
        handler(message.payload as Notification);
      }
    });
  }

  /**
   * Subscribe to system alerts
   */
  subscribeToAlerts(handler: (alert: any) => void): () => void {
    return this.subscribe('alert', (message) => {
      if (message.payload) {
        handler(message.payload);
      }
    });
  }

  /**
   * Subscribe to cluster status updates
   */
  subscribeToClusterUpdates(handler: (status: any) => void): () => void {
    return this.subscribe('cluster_status', (message) => {
      if (message.payload) {
        handler(message.payload);
      }
    });
  }

  /**
   * Subscribe to model status updates
   */
  subscribeToModelUpdates(handler: (modelUpdate: any) => void): () => void {
    return this.subscribe('model_update', (message) => {
      if (message.payload) {
        handler(message.payload);
      }
    });
  }

  /**
   * Get WebSocket connection status
   */
  getConnectionStatus(): 'connecting' | 'open' | 'closing' | 'closed' {
    if (!this.wsConnection) return 'closed';
    
    switch (this.wsConnection.readyState) {
      case WebSocket.CONNECTING:
        return 'connecting';
      case WebSocket.OPEN:
        return 'open';
      case WebSocket.CLOSING:
        return 'closing';
      case WebSocket.CLOSED:
        return 'closed';
      default:
        return 'closed';
    }
  }

  // Private Methods

  private handleWebSocketMessage(message: WebSocketMessage): void {
    const handler = this.messageHandlers.get(message.type);
    if (handler) {
      handler(message);
    }

    // Emit general message event
    const generalHandler = this.messageHandlers.get('*');
    if (generalHandler) {
      generalHandler(message);
    }
  }

  private attemptReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnection attempts reached');
      return;
    }

    this.reconnectAttempts++;
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
    
    console.log(`Attempting to reconnect in ${delay}ms (attempt ${this.reconnectAttempts})`);
    
    setTimeout(() => {
      this.connectWebSocket();
    }, delay);
  }

  /**
   * Subscribe to all message types
   */
  subscribeToAll(handler: (message: WebSocketMessage) => void): () => void {
    return this.subscribe('*', handler);
  }

  /**
   * Send ping to keep connection alive
   */
  ping(): void {
    this.sendMessage('ping', { timestamp: Date.now() });
  }

  /**
   * Start keep-alive ping interval
   */
  startKeepAlive(intervalMs: number = 30000): () => void {
    const interval = setInterval(() => {
      if (this.getConnectionStatus() === 'open') {
        this.ping();
      }
    }, intervalMs);

    return () => clearInterval(interval);
  }
}