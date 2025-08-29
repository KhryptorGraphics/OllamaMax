// API Service for OllamaMax Frontend
// Handles all backend communication with proper error handling and authentication

class APIService {
  constructor() {
    this.baseURL = this.getBaseURL();
    this.token = localStorage.getItem('auth_token');
    this.refreshToken = localStorage.getItem('refresh_token');
    this.wsConnection = null;
    this.wsReconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.reconnectDelay = 1000;
  }

  // Determine base URL based on environment
  getBaseURL() {
    if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
      return 'http://localhost:11434/api/v1';
    }
    return `${window.location.protocol}//${window.location.hostname}:11434/api/v1`;
  }

  // Set authentication token
  setToken(token, refreshToken = null) {
    this.token = token;
    this.refreshToken = refreshToken;
    localStorage.setItem('auth_token', token);
    if (refreshToken) {
      localStorage.setItem('refresh_token', refreshToken);
    }
  }

  // Clear authentication
  clearAuth() {
    this.token = null;
    this.refreshToken = null;
    localStorage.removeItem('auth_token');
    localStorage.removeItem('refresh_token');
  }

  // Get default headers
  getHeaders() {
    const headers = {
      'Content-Type': 'application/json',
    };
    
    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }
    
    return headers;
  }

  // Generic request method with error handling
  async request(endpoint, options = {}) {
    const url = `${this.baseURL}${endpoint}`;
    const config = {
      headers: this.getHeaders(),
      ...options,
    };

    try {
      const response = await fetch(url, config);
      
      // Handle authentication errors
      if (response.status === 401) {
        if (this.refreshToken) {
          const refreshed = await this.refreshAuthToken();
          if (refreshed) {
            // Retry the original request with new token
            config.headers = this.getHeaders();
            const retryResponse = await fetch(url, config);
            return this.handleResponse(retryResponse);
          }
        }
        this.clearAuth();
        throw new Error('Authentication required');
      }

      return this.handleResponse(response);
    } catch (error) {
      console.error(`API request failed: ${endpoint}`, error);
      throw error;
    }
  }

  // Handle response parsing
  async handleResponse(response) {
    const contentType = response.headers.get('content-type');
    
    if (!response.ok) {
      let errorMessage = `HTTP ${response.status}: ${response.statusText}`;
      
      if (contentType && contentType.includes('application/json')) {
        try {
          const errorData = await response.json();
          errorMessage = errorData.error || errorData.message || errorMessage;
        } catch (e) {
          // Ignore JSON parsing errors for error responses
        }
      }
      
      throw new Error(errorMessage);
    }

    if (contentType && contentType.includes('application/json')) {
      return await response.json();
    }
    
    return await response.text();
  }

  // Refresh authentication token
  async refreshAuthToken() {
    if (!this.refreshToken) return false;

    try {
      const response = await fetch(`${this.baseURL}/auth/refresh`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: this.refreshToken }),
      });

      if (response.ok) {
        const data = await response.json();
        this.setToken(data.access_token, data.refresh_token);
        return true;
      }
    } catch (error) {
      console.error('Token refresh failed:', error);
    }

    return false;
  }

  // Authentication endpoints
  async login(username, password) {
    const response = await this.request('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    });
    
    if (response.access_token) {
      this.setToken(response.access_token, response.refresh_token);
    }
    
    return response;
  }

  async register(userData) {
    return await this.request('/auth/register', {
      method: 'POST',
      body: JSON.stringify(userData),
    });
  }

  async logout() {
    try {
      await this.request('/auth/logout', { method: 'POST' });
    } catch (error) {
      console.error('Logout request failed:', error);
    } finally {
      this.clearAuth();
    }
  }

  // Health and status endpoints
  async getHealth() {
    return await this.request('/health');
  }

  async getReadiness() {
    return await this.request('/readiness');
  }

  async getClusterStatus() {
    return await this.request('/cluster/status');
  }

  // Node management
  async getNodes() {
    return await this.request('/nodes');
  }

  async getNode(nodeId) {
    return await this.request(`/nodes/${nodeId}`);
  }

  async updateNode(nodeId, data) {
    return await this.request(`/nodes/${nodeId}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  // Model management
  async getModels(page = 1, limit = 20) {
    return await this.request(`/models?page=${page}&limit=${limit}`);
  }

  async getModel(modelId) {
    return await this.request(`/models/${modelId}`);
  }

  async createModel(modelData) {
    return await this.request('/models', {
      method: 'POST',
      body: JSON.stringify(modelData),
    });
  }

  async deleteModel(modelId) {
    return await this.request(`/models/${modelId}`, {
      method: 'DELETE',
    });
  }

  // Inference endpoints
  async createInference(inferenceData) {
    return await this.request('/inference', {
      method: 'POST',
      body: JSON.stringify(inferenceData),
    });
  }

  async getInferenceStatus(requestId) {
    return await this.request(`/inference/${requestId}/status`);
  }

  // User management (admin only)
  async getUsers(page = 1, limit = 20) {
    return await this.request(`/admin/users?page=${page}&limit=${limit}`);
  }

  async createUser(userData) {
    return await this.request('/admin/users', {
      method: 'POST',
      body: JSON.stringify(userData),
    });
  }

  async updateUser(userId, userData) {
    return await this.request(`/admin/users/${userId}`, {
      method: 'PUT',
      body: JSON.stringify(userData),
    });
  }

  async deleteUser(userId) {
    return await this.request(`/admin/users/${userId}`, {
      method: 'DELETE',
    });
  }

  // Metrics endpoints
  async getMetrics() {
    return await this.request('/metrics');
  }

  async getSystemMetrics() {
    return await this.request('/admin/metrics/system');
  }

  // WebSocket connection for real-time updates
  connectWebSocket(onMessage, onError, onClose) {
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsURL = `${wsProtocol}//${window.location.hostname}:11434/api/v1/ws`;

    try {
      this.wsConnection = new WebSocket(wsURL);
      
      this.wsConnection.onopen = () => {
        console.log('WebSocket connected');
        this.wsReconnectAttempts = 0;
        
        // Send authentication if available
        if (this.token) {
          this.wsConnection.send(JSON.stringify({
            type: 'auth',
            token: this.token,
          }));
        }
      };

      this.wsConnection.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          if (onMessage) onMessage(data);
        } catch (error) {
          console.error('WebSocket message parsing error:', error);
        }
      };

      this.wsConnection.onerror = (error) => {
        console.error('WebSocket error:', error);
        if (onError) onError(error);
      };

      this.wsConnection.onclose = (event) => {
        console.log('WebSocket disconnected:', event.code, event.reason);
        if (onClose) onClose(event);
        
        // Attempt to reconnect if not a clean close
        if (event.code !== 1000 && this.wsReconnectAttempts < this.maxReconnectAttempts) {
          setTimeout(() => {
            this.wsReconnectAttempts++;
            console.log(`Attempting WebSocket reconnection (${this.wsReconnectAttempts}/${this.maxReconnectAttempts})`);
            this.connectWebSocket(onMessage, onError, onClose);
          }, this.reconnectDelay * this.wsReconnectAttempts);
        }
      };

    } catch (error) {
      console.error('WebSocket connection failed:', error);
      if (onError) onError(error);
    }
  }

  // Send WebSocket message
  sendWebSocketMessage(message) {
    if (this.wsConnection && this.wsConnection.readyState === WebSocket.OPEN) {
      this.wsConnection.send(JSON.stringify(message));
    } else {
      console.warn('WebSocket not connected');
    }
  }

  // Close WebSocket connection
  disconnectWebSocket() {
    if (this.wsConnection) {
      this.wsConnection.close(1000, 'Client disconnect');
      this.wsConnection = null;
    }
  }

  // File upload helper
  async uploadFile(endpoint, file, onProgress) {
    const formData = new FormData();
    formData.append('file', file);

    return new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest();
      
      xhr.upload.addEventListener('progress', (event) => {
        if (event.lengthComputable && onProgress) {
          const percentComplete = (event.loaded / event.total) * 100;
          onProgress(percentComplete);
        }
      });

      xhr.addEventListener('load', () => {
        if (xhr.status >= 200 && xhr.status < 300) {
          try {
            const response = JSON.parse(xhr.responseText);
            resolve(response);
          } catch (error) {
            resolve(xhr.responseText);
          }
        } else {
          reject(new Error(`Upload failed: ${xhr.status} ${xhr.statusText}`));
        }
      });

      xhr.addEventListener('error', () => {
        reject(new Error('Upload failed'));
      });

      xhr.open('POST', `${this.baseURL}${endpoint}`);
      
      // Add auth header if available
      if (this.token) {
        xhr.setRequestHeader('Authorization', `Bearer ${this.token}`);
      }
      
      xhr.send(formData);
    });
  }

  // Utility methods
  isAuthenticated() {
    return !!this.token;
  }

  getAuthToken() {
    return this.token;
  }

  // Database operations (admin only)
  async executeDatabaseQuery(query) {
    return await this.request('/admin/database/query', {
      method: 'POST',
      body: JSON.stringify({ query }),
    });
  }

  async getDatabaseSchema() {
    return await this.request('/admin/database/schema');
  }

  async getDatabaseStats() {
    return await this.request('/admin/database/stats');
  }
}

// Create and export singleton instance
const apiService = new APIService();
export default apiService;

// Also make it available globally for debugging
if (typeof window !== 'undefined') {
  window.apiService = apiService;
}
