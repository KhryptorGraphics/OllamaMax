/**
 * @fileoverview Security and audit API client
 * @description Handles security events, audit logs, and compliance monitoring
 */

import { BaseAPIClient } from './base';
import {
  SecurityEvent,
  AuditLog,
  PaginatedResponse,
  APIResponse,
  RequestConfig,
} from '../../types/api';

export class SecurityAPI extends BaseAPIClient {
  /**
   * Get security events with optional filtering
   */
  async getSecurityEvents(
    filters?: {
      type?: 'authentication' | 'authorization' | 'access' | 'security';
      severity?: 'low' | 'medium' | 'high' | 'critical';
      user_id?: string;
      start_time?: string;
      end_time?: string;
      limit?: number;
      offset?: number;
    },
    config?: RequestConfig
  ): Promise<PaginatedResponse<SecurityEvent>> {
    const params = new URLSearchParams();
    if (filters) {
      Object.entries(filters).forEach(([key, value]) => {
        if (value !== undefined) {
          params.append(key, value.toString());
        }
      });
    }

    const endpoint = `/api/security/events${params.toString() ? `?${params}` : ''}`;
    const response = await this.get<PaginatedResponse<SecurityEvent>>(endpoint, config);
    return response.data!;
  }

  /**
   * Get specific security event
   */
  async getSecurityEvent(eventId: string, config?: RequestConfig): Promise<SecurityEvent> {
    const response = await this.get<SecurityEvent>(`/api/security/events/${eventId}`, config);
    return response.data!;
  }

  /**
   * Create security event
   */
  async createSecurityEvent(
    event: Omit<SecurityEvent, 'id' | 'timestamp'>,
    config?: RequestConfig
  ): Promise<SecurityEvent> {
    const response = await this.post<SecurityEvent>('/api/security/events', event, config);
    return response.data!;
  }

  /**
   * Get audit logs with filtering and pagination
   */
  async getAuditLogs(
    filters?: {
      action?: string;
      resource?: string;
      user_id?: string;
      success?: boolean;
      start_time?: string;
      end_time?: string;
      limit?: number;
      offset?: number;
    },
    config?: RequestConfig
  ): Promise<PaginatedResponse<AuditLog>> {
    const params = new URLSearchParams();
    if (filters) {
      Object.entries(filters).forEach(([key, value]) => {
        if (value !== undefined) {
          params.append(key, value.toString());
        }
      });
    }

    const endpoint = `/api/security/audit${params.toString() ? `?${params}` : ''}`;
    const response = await this.get<PaginatedResponse<AuditLog>>(endpoint, config);
    return response.data!;
  }

  /**
   * Get specific audit log entry
   */
  async getAuditLog(logId: string, config?: RequestConfig): Promise<AuditLog> {
    const response = await this.get<AuditLog>(`/api/security/audit/${logId}`, config);
    return response.data!;
  }

  /**
   * Create audit log entry
   */
  async createAuditLog(
    log: Omit<AuditLog, 'id' | 'timestamp'>,
    config?: RequestConfig
  ): Promise<AuditLog> {
    const response = await this.post<AuditLog>('/api/security/audit', log, config);
    return response.data!;
  }

  /**
   * Get security statistics
   */
  async getSecurityStats(
    period?: '24h' | '7d' | '30d',
    config?: RequestConfig
  ): Promise<{
    total_events: number;
    critical_events: number;
    failed_logins: number;
    successful_logins: number;
    api_key_usage: number;
    blocked_ips: number;
    event_breakdown: Record<string, number>;
    severity_breakdown: Record<string, number>;
  }> {
    const params = period ? `?period=${period}` : '';
    const response = await this.get<{
      total_events: number;
      critical_events: number;
      failed_logins: number;
      successful_logins: number;
      api_key_usage: number;
      blocked_ips: number;
      event_breakdown: Record<string, number>;
      severity_breakdown: Record<string, number>;
    }>(`/api/security/stats${params}`, config);
    return response.data!;
  }

  /**
   * Get failed login attempts
   */
  async getFailedLogins(
    timeRange?: '1h' | '24h' | '7d',
    config?: RequestConfig
  ): Promise<Array<{
    ip_address: string;
    username: string;
    timestamp: string;
    user_agent: string;
    attempts: number;
  }>> {
    const params = timeRange ? `?range=${timeRange}` : '';
    const response = await this.get<{ failed_logins: Array<{
      ip_address: string;
      username: string;
      timestamp: string;
      user_agent: string;
      attempts: number;
    }> }>(`/api/security/failed-logins${params}`, config);
    return response.data!.failed_logins;
  }

  /**
   * Get blocked IP addresses
   */
  async getBlockedIPs(config?: RequestConfig): Promise<Array<{
    ip_address: string;
    reason: string;
    blocked_at: string;
    expires_at?: string;
    attempts: number;
  }>> {
    const response = await this.get<{ blocked_ips: Array<{
      ip_address: string;
      reason: string;
      blocked_at: string;
      expires_at?: string;
      attempts: number;
    }> }>('/api/security/blocked-ips', config);
    return response.data!.blocked_ips;
  }

  /**
   * Block IP address
   */
  async blockIP(
    ipAddress: string,
    reason: string,
    duration?: number,
    config?: RequestConfig
  ): Promise<{ message: string }> {
    const response = await this.post<{ message: string }>(
      '/api/security/block-ip',
      { ip_address: ipAddress, reason, duration },
      config
    );
    return response.data!;
  }

  /**
   * Unblock IP address
   */
  async unblockIP(ipAddress: string, config?: RequestConfig): Promise<{ message: string }> {
    const response = await this.delete<{ message: string }>(
      `/api/security/block-ip/${encodeURIComponent(ipAddress)}`,
      config
    );
    return response.data!;
  }

  /**
   * Get suspicious activities
   */
  async getSuspiciousActivities(
    config?: RequestConfig
  ): Promise<Array<{
    type: string;
    description: string;
    user_id?: string;
    ip_address: string;
    timestamp: string;
    risk_score: number;
    details: Record<string, any>;
  }>> {
    const response = await this.get<{ activities: Array<{
      type: string;
      description: string;
      user_id?: string;
      ip_address: string;
      timestamp: string;
      risk_score: number;
      details: Record<string, any>;
    }> }>('/api/security/suspicious-activities', config);
    return response.data!.activities;
  }

  /**
   * Run security scan
   */
  async runSecurityScan(
    scanType: 'vulnerability' | 'permissions' | 'configuration' | 'full',
    config?: RequestConfig
  ): Promise<{
    scan_id: string;
    status: 'started' | 'running' | 'completed' | 'failed';
    started_at: string;
  }> {
    const response = await this.post<{
      scan_id: string;
      status: 'started' | 'running' | 'completed' | 'failed';
      started_at: string;
    }>('/api/security/scan', { type: scanType }, config);
    return response.data!;
  }

  /**
   * Get security scan results
   */
  async getScanResults(scanId: string, config?: RequestConfig): Promise<{
    scan_id: string;
    type: string;
    status: 'running' | 'completed' | 'failed';
    started_at: string;
    completed_at?: string;
    results: {
      vulnerabilities: Array<{
        id: string;
        severity: 'low' | 'medium' | 'high' | 'critical';
        title: string;
        description: string;
        affected_component: string;
        remediation: string;
      }>;
      misconfigurations: Array<{
        component: string;
        issue: string;
        severity: 'low' | 'medium' | 'high';
        recommendation: string;
      }>;
      permissions: Array<{
        user_id: string;
        excessive_permissions: string[];
        recommendations: string[];
      }>;
    };
    summary: {
      total_issues: number;
      critical_issues: number;
      high_issues: number;
      medium_issues: number;
      low_issues: number;
    };
  }> {
    const response = await this.get<{
      scan_id: string;
      type: string;
      status: 'running' | 'completed' | 'failed';
      started_at: string;
      completed_at?: string;
      results: {
        vulnerabilities: Array<{
          id: string;
          severity: 'low' | 'medium' | 'high' | 'critical';
          title: string;
          description: string;
          affected_component: string;
          remediation: string;
        }>;
        misconfigurations: Array<{
          component: string;
          issue: string;
          severity: 'low' | 'medium' | 'high';
          recommendation: string;
        }>;
        permissions: Array<{
          user_id: string;
          excessive_permissions: string[];
          recommendations: string[];
        }>;
      };
      summary: {
        total_issues: number;
        critical_issues: number;
        high_issues: number;
        medium_issues: number;
        low_issues: number;
      };
    }>(`/api/security/scan/${scanId}/results`, config);
    return response.data!;
  }

  /**
   * Get compliance status
   */
  async getComplianceStatus(
    framework?: 'SOC2' | 'GDPR' | 'HIPAA' | 'PCI-DSS',
    config?: RequestConfig
  ): Promise<{
    framework: string;
    status: 'compliant' | 'non-compliant' | 'partial';
    score: number;
    last_assessment: string;
    controls: Array<{
      id: string;
      name: string;
      status: 'pass' | 'fail' | 'not-applicable';
      evidence?: string;
      remediation?: string;
    }>;
    recommendations: string[];
  }> {
    const params = framework ? `?framework=${framework}` : '';
    const response = await this.get<{
      framework: string;
      status: 'compliant' | 'non-compliant' | 'partial';
      score: number;
      last_assessment: string;
      controls: Array<{
        id: string;
        name: string;
        status: 'pass' | 'fail' | 'not-applicable';
        evidence?: string;
        remediation?: string;
      }>;
      recommendations: string[];
    }>(`/api/security/compliance${params}`, config);
    return response.data!;
  }

  /**
   * Export security data
   */
  async exportSecurityData(
    type: 'events' | 'audit' | 'compliance',
    format: 'json' | 'csv' | 'pdf',
    filters?: Record<string, any>,
    config?: RequestConfig
  ): Promise<Blob | string> {
    const params = new URLSearchParams({ type, format });
    if (filters) {
      Object.entries(filters).forEach(([key, value]) => {
        if (value !== undefined) {
          params.append(key, value.toString());
        }
      });
    }

    const response = await this.get<Blob | string>(`/api/security/export?${params}`, {
      ...config,
      headers: {
        Accept: format === 'json' ? 'application/json' : 
                format === 'csv' ? 'text/csv' : 'application/pdf',
        ...config?.headers,
      },
    });
    return response.data!;
  }

  /**
   * Monitor security events in real-time
   */
  async *monitorSecurityEvents(config?: RequestConfig): AsyncGenerator<SecurityEvent> {
    const stream = await this.stream('/api/security/events/stream', {
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
              const event: SecurityEvent = JSON.parse(line.substring(6));
              yield event;
            } catch (error) {
              console.warn('Failed to parse security event:', line);
            }
          }
        }
      }
    } finally {
      reader.releaseLock();
    }
  }

  /**
   * Get security recommendations
   */
  async getSecurityRecommendations(config?: RequestConfig): Promise<Array<{
    id: string;
    category: 'authentication' | 'authorization' | 'encryption' | 'monitoring' | 'configuration';
    priority: 'low' | 'medium' | 'high' | 'critical';
    title: string;
    description: string;
    action_required: string;
    impact: string;
    effort: 'low' | 'medium' | 'high';
  }>> {
    const response = await this.get<{ recommendations: Array<{
      id: string;
      category: 'authentication' | 'authorization' | 'encryption' | 'monitoring' | 'configuration';
      priority: 'low' | 'medium' | 'high' | 'critical';
      title: string;
      description: string;
      action_required: string;
      impact: string;
      effort: 'low' | 'medium' | 'high';
    }> }>('/api/security/recommendations', config);
    return response.data!.recommendations;
  }
}