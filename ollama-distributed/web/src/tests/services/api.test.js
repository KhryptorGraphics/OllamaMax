// API Service Tests
// Comprehensive tests for the API service integration

import apiService from '../../services/api.js';

// Mock fetch for testing
global.fetch = jest.fn();

describe('API Service', () => {
  beforeEach(() => {
    fetch.mockClear();
    localStorage.clear();
  });

  describe('Authentication', () => {
    test('should set and get auth token', () => {
      const token = 'test-token';
      const refreshToken = 'test-refresh-token';
      
      apiService.setToken(token, refreshToken);
      
      expect(apiService.getAuthToken()).toBe(token);
      expect(localStorage.getItem('auth_token')).toBe(token);
      expect(localStorage.getItem('refresh_token')).toBe(refreshToken);
    });

    test('should clear authentication', () => {
      apiService.setToken('test-token', 'test-refresh');
      apiService.clearAuth();
      
      expect(apiService.getAuthToken()).toBeNull();
      expect(localStorage.getItem('auth_token')).toBeNull();
      expect(localStorage.getItem('refresh_token')).toBeNull();
    });

    test('should login successfully', async () => {
      const mockResponse = {
        access_token: 'new-token',
        refresh_token: 'new-refresh',
        user: { id: 1, username: 'testuser' }
      };

      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse
      });

      const result = await apiService.login('testuser', 'password');
      
      expect(fetch).toHaveBeenCalledWith(
        expect.stringContaining('/auth/login'),
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify({ username: 'testuser', password: 'password' })
        })
      );
      
      expect(result).toEqual(mockResponse);
      expect(apiService.getAuthToken()).toBe('new-token');
    });

    test('should handle login failure', async () => {
      fetch.mockResolvedValueOnce({
        ok: false,
        status: 401,
        statusText: 'Unauthorized'
      });

      await expect(apiService.login('testuser', 'wrongpassword'))
        .rejects.toThrow('HTTP 401: Unauthorized');
    });

    test('should refresh token automatically', async () => {
      // Set initial tokens
      apiService.setToken('expired-token', 'valid-refresh');

      // Mock 401 response for initial request
      fetch.mockResolvedValueOnce({
        ok: false,
        status: 401
      });

      // Mock successful refresh
      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          access_token: 'new-token',
          refresh_token: 'new-refresh'
        })
      });

      // Mock successful retry
      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ data: 'success' })
      });

      const result = await apiService.request('/test-endpoint');
      
      expect(fetch).toHaveBeenCalledTimes(3); // Original, refresh, retry
      expect(result).toEqual({ data: 'success' });
    });
  });

  describe('API Requests', () => {
    beforeEach(() => {
      apiService.setToken('valid-token');
    });

    test('should make GET request with auth headers', async () => {
      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ data: 'test' })
      });

      await apiService.request('/test');
      
      expect(fetch).toHaveBeenCalledWith(
        expect.stringContaining('/test'),
        expect.objectContaining({
          headers: expect.objectContaining({
            'Authorization': 'Bearer valid-token',
            'Content-Type': 'application/json'
          })
        })
      );
    });

    test('should handle JSON responses', async () => {
      const mockData = { users: [{ id: 1, name: 'Test' }] };
      
      fetch.mockResolvedValueOnce({
        ok: true,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => mockData
      });

      const result = await apiService.request('/users');
      expect(result).toEqual(mockData);
    });

    test('should handle text responses', async () => {
      fetch.mockResolvedValueOnce({
        ok: true,
        headers: new Map([['content-type', 'text/plain']]),
        text: async () => 'plain text response'
      });

      const result = await apiService.request('/health');
      expect(result).toBe('plain text response');
    });

    test('should handle network errors', async () => {
      fetch.mockRejectedValueOnce(new Error('Network error'));

      await expect(apiService.request('/test'))
        .rejects.toThrow('Network error');
    });
  });

  describe('Resource Endpoints', () => {
    beforeEach(() => {
      apiService.setToken('valid-token');
    });

    test('should get health status', async () => {
      const mockHealth = { status: 'healthy', uptime: '99.9%' };
      
      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockHealth
      });

      const result = await apiService.getHealth();
      
      expect(fetch).toHaveBeenCalledWith(
        expect.stringContaining('/health'),
        expect.any(Object)
      );
      expect(result).toEqual(mockHealth);
    });

    test('should get cluster status', async () => {
      const mockStatus = { nodes: 3, leader: 'node-1' };
      
      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockStatus
      });

      const result = await apiService.getClusterStatus();
      expect(result).toEqual(mockStatus);
    });

    test('should get nodes list', async () => {
      const mockNodes = { nodes: [{ id: 'node-1', status: 'online' }] };
      
      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockNodes
      });

      const result = await apiService.getNodes();
      expect(result).toEqual(mockNodes);
    });

    test('should get models with pagination', async () => {
      const mockModels = { 
        models: [{ id: 'model-1', name: 'llama2' }],
        total: 1,
        page: 1
      };
      
      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockModels
      });

      const result = await apiService.getModels(1, 10);
      
      expect(fetch).toHaveBeenCalledWith(
        expect.stringContaining('/models?page=1&limit=10'),
        expect.any(Object)
      );
      expect(result).toEqual(mockModels);
    });

    test('should create model', async () => {
      const modelData = { name: 'new-model', size: '7b' };
      const mockResponse = { id: 'model-123', ...modelData };
      
      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse
      });

      const result = await apiService.createModel(modelData);
      
      expect(fetch).toHaveBeenCalledWith(
        expect.stringContaining('/models'),
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify(modelData)
        })
      );
      expect(result).toEqual(mockResponse);
    });

    test('should handle admin endpoints with proper permissions', async () => {
      const mockUsers = { users: [{ id: 1, username: 'admin' }] };
      
      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockUsers
      });

      const result = await apiService.getUsers(1, 20);
      
      expect(fetch).toHaveBeenCalledWith(
        expect.stringContaining('/admin/users?page=1&limit=20'),
        expect.any(Object)
      );
      expect(result).toEqual(mockUsers);
    });
  });

  describe('File Upload', () => {
    test('should upload file with progress tracking', async () => {
      const mockFile = new File(['test content'], 'test.txt', { type: 'text/plain' });
      const mockProgress = jest.fn();
      
      // Mock XMLHttpRequest
      const mockXHR = {
        upload: { addEventListener: jest.fn() },
        addEventListener: jest.fn(),
        open: jest.fn(),
        setRequestHeader: jest.fn(),
        send: jest.fn(),
        status: 200,
        responseText: JSON.stringify({ success: true, file_id: 'file-123' })
      };

      global.XMLHttpRequest = jest.fn(() => mockXHR);

      const uploadPromise = apiService.uploadFile('/upload', mockFile, mockProgress);
      
      // Simulate successful upload
      const loadHandler = mockXHR.addEventListener.mock.calls.find(
        call => call[0] === 'load'
      )[1];
      loadHandler();

      const result = await uploadPromise;
      
      expect(mockXHR.open).toHaveBeenCalledWith('POST', expect.stringContaining('/upload'));
      expect(result).toEqual({ success: true, file_id: 'file-123' });
    });
  });

  describe('Error Handling', () => {
    test('should handle 404 errors', async () => {
      fetch.mockResolvedValueOnce({
        ok: false,
        status: 404,
        statusText: 'Not Found'
      });

      await expect(apiService.request('/nonexistent'))
        .rejects.toThrow('HTTP 404: Not Found');
    });

    test('should handle server errors with JSON error messages', async () => {
      fetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ error: 'Internal server error', code: 'SERVER_ERROR' })
      });

      await expect(apiService.request('/error'))
        .rejects.toThrow('Internal server error');
    });

    test('should handle malformed JSON responses', async () => {
      fetch.mockResolvedValueOnce({
        ok: true,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => { throw new Error('Invalid JSON'); }
      });

      await expect(apiService.request('/malformed'))
        .rejects.toThrow('Invalid JSON');
    });
  });

  describe('Base URL Detection', () => {
    test('should use localhost URL in development', () => {
      // Mock localhost environment
      Object.defineProperty(window, 'location', {
        value: { hostname: 'localhost', protocol: 'http:' },
        writable: true
      });

      const service = new (apiService.constructor)();
      expect(service.getBaseURL()).toBe('http://localhost:11434/api/v1');
    });

    test('should use current host in production', () => {
      // Mock production environment
      Object.defineProperty(window, 'location', {
        value: { hostname: 'example.com', protocol: 'https:' },
        writable: true
      });

      const service = new (apiService.constructor)();
      expect(service.getBaseURL()).toBe('https://example.com:11434/api/v1');
    });
  });
});
