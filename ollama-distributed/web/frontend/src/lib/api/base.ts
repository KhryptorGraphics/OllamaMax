/**
 * @fileoverview Base API client with interceptors and error handling
 * @description Core HTTP client with retry logic, authentication, and request/response interceptors
 */

import { APIError, APIResponse, RequestConfig, APIClientConfig } from '../../types/api';
import { useAuthStore } from '../../store/auth';

export class APIError extends Error {
  constructor(
    message: string,
    public status?: number,
    public code?: string,
    public details?: Record<string, any>
  ) {
    super(message);
    this.name = 'APIError';
  }
}

export class BaseAPIClient {
  private baseURL: string;
  private defaultTimeout: number;
  private defaultRetries: number;
  private defaultRetryDelay: number;

  constructor(config: APIClientConfig) {
    this.baseURL = config.baseURL;
    this.defaultTimeout = config.timeout || 30000;
    this.defaultRetries = config.retries || 3;
    this.defaultRetryDelay = config.retryDelay || 1000;
  }

  /**
   * Make HTTP request with automatic retries and error handling
   */
  async request<T = any>(
    endpoint: string,
    options: RequestInit & RequestConfig = {}
  ): Promise<APIResponse<T>> {
    const {
      timeout = this.defaultTimeout,
      retries = this.defaultRetries,
      retryDelay = this.defaultRetryDelay,
      headers = {},
      signal,
      ...fetchOptions
    } = options;

    const url = `${this.baseURL}${endpoint}`;
    
    // Prepare headers with authentication
    const requestHeaders = await this.prepareHeaders(headers);
    
    // Create abort controller for timeout
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), timeout);
    
    // Combine signals if provided
    const requestSignal = signal ? this.combineSignals(signal, controller.signal) : controller.signal;

    const requestConfig: RequestInit = {
      ...fetchOptions,
      headers: requestHeaders,
      signal: requestSignal,
    };

    // Apply request interceptors
    const finalConfig = await this.applyRequestInterceptors(requestConfig);

    let lastError: Error;
    
    for (let attempt = 0; attempt <= retries; attempt++) {
      try {
        const response = await fetch(url, finalConfig);
        clearTimeout(timeoutId);

        // Apply response interceptors
        const interceptedResponse = await this.applyResponseInterceptors(response.clone());
        
        // Handle HTTP errors
        if (!interceptedResponse.ok) {
          const errorData = await this.parseErrorResponse(interceptedResponse);
          throw new APIError(
            errorData.error || `HTTP ${interceptedResponse.status}`,
            interceptedResponse.status,
            errorData.code,
            errorData.details
          );
        }

        // Parse successful response
        const data = await this.parseResponse<T>(interceptedResponse);
        return { data, timestamp: Date.now() };

      } catch (error) {
        lastError = error as Error;
        
        // Don't retry on authentication errors or client errors (4xx)
        if (error instanceof APIError && error.status && error.status >= 400 && error.status < 500) {
          break;
        }
        
        // Don't retry on the last attempt
        if (attempt === retries) {
          break;
        }
        
        // Wait before retrying with exponential backoff
        await this.delay(retryDelay * Math.pow(2, attempt));
      }
    }

    clearTimeout(timeoutId);
    throw lastError || new APIError('Request failed after retries');
  }

  /**
   * GET request helper
   */
  async get<T = any>(endpoint: string, config?: RequestConfig): Promise<APIResponse<T>> {
    return this.request<T>(endpoint, { ...config, method: 'GET' });
  }

  /**
   * POST request helper
   */
  async post<T = any>(
    endpoint: string,
    data?: any,
    config?: RequestConfig
  ): Promise<APIResponse<T>> {
    return this.request<T>(endpoint, {
      ...config,
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
      headers: {
        'Content-Type': 'application/json',
        ...config?.headers,
      },
    });
  }

  /**
   * PUT request helper
   */
  async put<T = any>(
    endpoint: string,
    data?: any,
    config?: RequestConfig
  ): Promise<APIResponse<T>> {
    return this.request<T>(endpoint, {
      ...config,
      method: 'PUT',
      body: data ? JSON.stringify(data) : undefined,
      headers: {
        'Content-Type': 'application/json',
        ...config?.headers,
      },
    });
  }

  /**
   * DELETE request helper
   */
  async delete<T = any>(endpoint: string, config?: RequestConfig): Promise<APIResponse<T>> {
    return this.request<T>(endpoint, { ...config, method: 'DELETE' });
  }

  /**
   * PATCH request helper
   */
  async patch<T = any>(
    endpoint: string,
    data?: any,
    config?: RequestConfig
  ): Promise<APIResponse<T>> {
    return this.request<T>(endpoint, {
      ...config,
      method: 'PATCH',
      body: data ? JSON.stringify(data) : undefined,
      headers: {
        'Content-Type': 'application/json',
        ...config?.headers,
      },
    });
  }

  /**
   * Streaming request for Server-Sent Events
   */
  async stream(
    endpoint: string,
    options: RequestInit & RequestConfig = {}
  ): Promise<ReadableStream> {
    const response = await this.request(endpoint, {
      ...options,
      headers: {
        Accept: 'text/event-stream',
        'Cache-Control': 'no-cache',
        ...options.headers,
      },
    });

    if (!response.data?.body) {
      throw new APIError('No stream available');
    }

    return response.data.body;
  }

  /**
   * Prepare headers with authentication token
   */
  private async prepareHeaders(headers: Record<string, string>): Promise<HeadersInit> {
    const authStore = useAuthStore.getState();
    const requestHeaders: Record<string, string> = {
      'X-Request-ID': this.generateRequestId(),
      ...headers,
    };

    // Add authorization header if token exists
    if (authStore.token) {
      requestHeaders.Authorization = `Bearer ${authStore.token}`;
    }

    return requestHeaders;
  }

  /**
   * Apply request interceptors
   */
  private async applyRequestInterceptors(config: RequestInit): Promise<RequestInit> {
    let finalConfig = config;
    
    // Apply auth token refresh if needed
    finalConfig = await this.handleTokenRefresh(finalConfig);
    
    return finalConfig;
  }

  /**
   * Apply response interceptors
   */
  private async applyResponseInterceptors(response: Response): Promise<Response> {
    // Handle token expiration
    if (response.status === 401) {
      await this.handleTokenExpiration();
    }
    
    return response;
  }

  /**
   * Handle token refresh logic
   */
  private async handleTokenRefresh(config: RequestInit): Promise<RequestInit> {
    const authStore = useAuthStore.getState();
    
    // Check if token needs refresh (implement token expiration check)
    if (authStore.token && this.isTokenExpiring(authStore.token)) {
      try {
        // Attempt to refresh token
        await this.refreshToken();
        
        // Update headers with new token
        const headers = new Headers(config.headers);
        const newToken = useAuthStore.getState().token;
        if (newToken) {
          headers.set('Authorization', `Bearer ${newToken}`);
        }
        
        return { ...config, headers };
      } catch (error) {
        console.warn('Token refresh failed:', error);
      }
    }
    
    return config;
  }

  /**
   * Handle token expiration
   */
  private async handleTokenExpiration(): Promise<void> {
    const authStore = useAuthStore.getState();
    
    try {
      await this.refreshToken();
    } catch (error) {
      // Refresh failed, clear auth state
      authStore.clear();
      
      // Redirect to login if in browser
      if (typeof window !== 'undefined') {
        window.location.href = '/auth/login';
      }
    }
  }

  /**
   * Refresh authentication token
   */
  private async refreshToken(): Promise<void> {
    const authStore = useAuthStore.getState();
    
    if (!authStore.refreshToken) {
      throw new APIError('No refresh token available');
    }

    const response = await fetch(`${this.baseURL}/api/v1/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ token: authStore.refreshToken }),
    });

    if (!response.ok) {
      throw new APIError('Token refresh failed');
    }

    const data = await response.json();
    authStore.setAuth({
      token: data.token,
      refreshToken: data.refresh_token,
      user: data.user,
    });
  }

  /**
   * Check if token is expiring soon
   */
  private isTokenExpiring(token: string): boolean {
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      const exp = payload.exp * 1000; // Convert to milliseconds
      const now = Date.now();
      const fiveMinutes = 5 * 60 * 1000;
      
      return exp - now < fiveMinutes;
    } catch {
      return false;
    }
  }

  /**
   * Parse successful response
   */
  private async parseResponse<T>(response: Response): Promise<T> {
    const contentType = response.headers.get('content-type');
    
    if (contentType?.includes('application/json')) {
      return response.json();
    }
    
    if (contentType?.includes('text/')) {
      return response.text() as unknown as T;
    }
    
    return response.blob() as unknown as T;
  }

  /**
   * Parse error response
   */
  private async parseErrorResponse(response: Response): Promise<APIError> {
    try {
      const data = await response.json();
      return {
        error: data.error || 'Unknown error',
        code: data.code,
        details: data.details,
      };
    } catch {
      return {
        error: `HTTP ${response.status}: ${response.statusText}`,
      };
    }
  }

  /**
   * Combine abort signals
   */
  private combineSignals(signal1: AbortSignal, signal2: AbortSignal): AbortSignal {
    const controller = new AbortController();
    
    const abort = () => controller.abort();
    
    if (signal1.aborted || signal2.aborted) {
      abort();
    } else {
      signal1.addEventListener('abort', abort);
      signal2.addEventListener('abort', abort);
    }
    
    return controller.signal;
  }

  /**
   * Generate unique request ID
   */
  private generateRequestId(): string {
    return `req_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  /**
   * Delay helper for retries
   */
  private delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}