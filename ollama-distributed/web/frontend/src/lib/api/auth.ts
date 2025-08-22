/**
 * @fileoverview Authentication API client
 * @description Handles user authentication, session management, and API keys
 */

import { BaseAPIClient } from './base';
import {
  LoginRequest,
  LoginResponse,
  RegisterRequest,
  RefreshTokenRequest,
  User,
  APIKey,
  CreateAPIKeyRequest,
  CreateAPIKeyResponse,
  AdminStats,
  APIResponse,
  RequestConfig,
} from '../../types/api';

export class AuthAPI extends BaseAPIClient {
  /**
   * Authenticate user with username and password
   */
  async login(credentials: LoginRequest, config?: RequestConfig): Promise<LoginResponse> {
    const response = await this.post<LoginResponse>('/api/v1/login', credentials, config);
    return response.data!;
  }

  /**
   * Register new user account
   */
  async register(userData: RegisterRequest, config?: RequestConfig): Promise<{ user: User }> {
    const response = await this.post<{ user: User }>('/api/v1/register', userData, config);
    return response.data!;
  }

  /**
   * Refresh authentication token
   */
  async refreshToken(refreshData: RefreshTokenRequest, config?: RequestConfig): Promise<LoginResponse> {
    const response = await this.post<LoginResponse>('/api/v1/refresh', refreshData, config);
    return response.data!;
  }

  /**
   * Logout current user
   */
  async logout(config?: RequestConfig): Promise<{ message: string }> {
    const response = await this.post<{ message: string }>('/api/v1/user/logout', undefined, config);
    return response.data!;
  }

  /**
   * Get current user profile
   */
  async getProfile(config?: RequestConfig): Promise<{ user: User }> {
    const response = await this.get<{ user: User }>('/api/v1/user/profile', config);
    return response.data!;
  }

  /**
   * Update user profile
   */
  async updateProfile(
    updates: Partial<Pick<User, 'email' | 'metadata'>>,
    config?: RequestConfig
  ): Promise<{ user: User }> {
    const response = await this.put<{ user: User }>('/api/v1/user/profile', updates, config);
    return response.data!;
  }

  /**
   * Change user password
   */
  async changePassword(
    passwordData: { current_password: string; new_password: string },
    config?: RequestConfig
  ): Promise<{ message: string }> {
    const response = await this.post<{ message: string }>(
      '/api/v1/user/change-password',
      passwordData,
      config
    );
    return response.data!;
  }

  /**
   * Get user sessions
   */
  async getSessions(config?: RequestConfig): Promise<{ sessions: Array<any> }> {
    const response = await this.get<{ sessions: Array<any> }>('/api/v1/user/sessions', config);
    return response.data!;
  }

  /**
   * Revoke specific session
   */
  async revokeSession(sessionId: string, config?: RequestConfig): Promise<{ message: string }> {
    const response = await this.delete<{ message: string }>(
      `/api/v1/user/sessions/${sessionId}`,
      config
    );
    return response.data!;
  }

  /**
   * List user's API keys
   */
  async listAPIKeys(config?: RequestConfig): Promise<{ api_keys: APIKey[] }> {
    const response = await this.get<{ api_keys: APIKey[] }>('/api/v1/api-keys', config);
    return response.data!;
  }

  /**
   * Create new API key
   */
  async createAPIKey(
    keyData: CreateAPIKeyRequest,
    config?: RequestConfig
  ): Promise<CreateAPIKeyResponse> {
    const response = await this.post<CreateAPIKeyResponse>('/api/v1/api-keys', keyData, config);
    return response.data!;
  }

  /**
   * Revoke API key
   */
  async revokeAPIKey(keyId: string, config?: RequestConfig): Promise<{ message: string }> {
    const response = await this.delete<{ message: string }>(`/api/v1/api-keys/${keyId}`, config);
    return response.data!;
  }

  // Admin endpoints (require admin role)

  /**
   * List all users (admin only)
   */
  async listUsers(config?: RequestConfig): Promise<{ users: User[] }> {
    const response = await this.get<{ users: User[] }>('/api/v1/admin/users', config);
    return response.data!;
  }

  /**
   * Create user (admin only)
   */
  async createUser(userData: RegisterRequest, config?: RequestConfig): Promise<{ user: User }> {
    const response = await this.post<{ user: User }>('/api/v1/admin/users', userData, config);
    return response.data!;
  }

  /**
   * Get specific user (admin only)
   */
  async getUser(userId: string, config?: RequestConfig): Promise<{ user: User }> {
    const response = await this.get<{ user: User }>(`/api/v1/admin/users/${userId}`, config);
    return response.data!;
  }

  /**
   * Update user (admin only)
   */
  async updateUser(
    userId: string,
    updates: Partial<User>,
    config?: RequestConfig
  ): Promise<{ user: User }> {
    const response = await this.put<{ user: User }>(`/api/v1/admin/users/${userId}`, updates, config);
    return response.data!;
  }

  /**
   * Delete user (admin only)
   */
  async deleteUser(userId: string, config?: RequestConfig): Promise<{ message: string }> {
    const response = await this.delete<{ message: string }>(`/api/v1/admin/users/${userId}`, config);
    return response.data!;
  }

  /**
   * Reset user password (admin only)
   */
  async resetUserPassword(
    userId: string,
    config?: RequestConfig
  ): Promise<{ message: string; temporary_password: string }> {
    const response = await this.post<{ message: string; temporary_password: string }>(
      `/api/v1/admin/users/${userId}/reset-password`,
      undefined,
      config
    );
    return response.data!;
  }

  /**
   * Get authentication statistics (admin only)
   */
  async getAuthStats(config?: RequestConfig): Promise<{ stats: AdminStats }> {
    const response = await this.get<{ stats: AdminStats }>('/api/v1/admin/stats', config);
    return response.data!;
  }

  /**
   * Check authentication health
   */
  async checkHealth(config?: RequestConfig): Promise<{
    status: string;
    service: string;
    timestamp: number;
  }> {
    const response = await this.get<{
      status: string;
      service: string;
      timestamp: number;
    }>('/api/v1/health', config);
    return response.data!;
  }
}