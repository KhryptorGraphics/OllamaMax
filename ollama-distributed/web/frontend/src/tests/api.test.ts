/**
 * @fileoverview API client integration tests
 * @description Integration tests for API client endpoints and error handling
 */

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { OllamaMaxAPI, getAPIClient } from '../lib/api';
import { APIError } from '../lib/api/base';

// Mock fetch
const mockFetch = vi.fn();
global.fetch = mockFetch;

describe('OllamaMaxAPI', () => {
  let api: OllamaMaxAPI;

  beforeEach(() => {
    api = new OllamaMaxAPI({
      baseURL: 'http://localhost:8080',
      timeout: 5000,
      retries: 1,
      retryDelay: 100
    });
    
    vi.clearAllMocks();
  });

  afterEach(() => {
    api.disconnectWebSocket();
  });

  describe('Initialization', () => {
    it('should initialize with default config', () => {
      const defaultApi = new OllamaMaxAPI();
      const config = defaultApi.getConfig();
      
      expect(config.timeout).toBe(30000);
      expect(config.retries).toBe(3);
      expect(config.retryDelay).toBe(1000);
    });

    it('should initialize with custom config', () => {
      const config = api.getConfig();
      
      expect(config.baseURL).toBe('http://localhost:8080');
      expect(config.timeout).toBe(5000);
      expect(config.retries).toBe(1);
    });

    it('should have all service clients', () => {
      expect(api.auth).toBeDefined();
      expect(api.cluster).toBeDefined();
      expect(api.models).toBeDefined();
      expect(api.monitoring).toBeDefined();
      expect(api.security).toBeDefined();
      expect(api.notifications).toBeDefined();
    });
  });

  describe('Connection Testing', () => {
    it('should test connection successfully', async () => {
      // Mock successful responses
      mockFetch
        .mockResolvedValueOnce(createMockResponse({ status: 'healthy' })) // cluster health
        .mockResolvedValueOnce(createMockResponse({ status: 'ok' })) // auth health
        .mockResolvedValueOnce(createMockResponse({ version: '1.0.0' })); // models version

      const result = await api.testConnection();

      expect(result.status).toBe('success');
      expect(result.latency).toBeGreaterThan(0);
      expect(result.services.cluster).toBe(true);
      expect(result.services.auth).toBe(true);
      expect(result.services.models).toBe(true);
    });

    it('should handle connection failures', async () => {
      // Mock failed responses
      mockFetch
        .mockRejectedValueOnce(new Error('Network error'))
        .mockRejectedValueOnce(new Error('Network error'))
        .mockRejectedValueOnce(new Error('Network error'));

      const result = await api.testConnection();

      expect(result.status).toBe('error');
      expect(result.services.cluster).toBe(false);
      expect(result.services.auth).toBe(false);
      expect(result.services.models).toBe(false);
    });

    it('should handle partial failures', async () => {
      // Mock mixed responses
      mockFetch
        .mockResolvedValueOnce(createMockResponse({ status: 'healthy' })) // cluster success
        .mockRejectedValueOnce(new Error('Auth service down')) // auth failure
        .mockResolvedValueOnce(createMockResponse({ version: '1.0.0' })); // models success

      const result = await api.testConnection();

      expect(result.status).toBe('error');
      expect(result.services.cluster).toBe(true);
      expect(result.services.auth).toBe(false);
      expect(result.services.models).toBe(true);
    });
  });

  describe('System Status', () => {
    it('should get comprehensive system status', async () => {
      const mockClusterStatus = { size: 3, active_nodes: 3, leader: 'node1' };
      const mockModels = [{ name: 'model1', size: 1000 }];
      const mockPerformance = { requests_per_second: 100, average_response_time: 50 };
      const mockAlerts = [{ id: '1', type: 'warning', message: 'High CPU usage' }];

      mockFetch
        .mockResolvedValueOnce(createMockResponse(mockClusterStatus))
        .mockResolvedValueOnce(createMockResponse({ models: mockModels }))
        .mockResolvedValueOnce(createMockResponse(mockPerformance))
        .mockResolvedValueOnce(createMockResponse({ alerts: mockAlerts }));

      const status = await api.getSystemStatus();

      expect(status.cluster).toEqual(mockClusterStatus);
      expect(status.models).toEqual(mockModels);
      expect(status.performance).toEqual(mockPerformance);
      expect(status.alerts).toEqual(mockAlerts);
      expect(status.timestamp).toBeGreaterThan(Date.now() - 1000);
    });

    it('should handle partial system status failures', async () => {
      mockFetch
        .mockResolvedValueOnce(createMockResponse({ size: 3 })) // cluster success
        .mockRejectedValueOnce(new Error('Models service down')) // models failure
        .mockResolvedValueOnce(createMockResponse({ rps: 100 })) // performance success
        .mockRejectedValueOnce(new Error('Alerts service down')); // alerts failure

      const status = await api.getSystemStatus();

      expect(status.cluster).toEqual({ size: 3 });
      expect(status.models).toBeNull();
      expect(status.performance).toEqual({ rps: 100 });
      expect(status.alerts).toBeNull();
    });
  });

  describe('Service Client Integration', () => {
    describe('Cluster API', () => {
      it('should get cluster health', async () => {
        const mockHealth = { 
          status: 'healthy', 
          distributed: true, 
          cluster_size: 3 
        };

        mockFetch.mockResolvedValueOnce(createMockResponse(mockHealth));

        const health = await api.cluster.getHealth();
        expect(health).toEqual(mockHealth);
        expect(mockFetch).toHaveBeenCalledWith(
          'http://localhost:8080/health',
          expect.objectContaining({ method: 'GET' })
        );
      });

      it('should get cluster nodes', async () => {
        const mockNodes = [
          { id: 'node1', status: 'active', role: 'leader' },
          { id: 'node2', status: 'active', role: 'follower' }
        ];

        mockFetch.mockResolvedValueOnce(createMockResponse({ nodes: mockNodes }));

        const nodes = await api.cluster.getNodes();
        expect(nodes).toEqual(mockNodes);
      });
    });

    describe('Models API', () => {
      it('should list models', async () => {
        const mockModels = [
          { name: 'llama2', size: 4000000000, distribution: { availability: 'full' } },
          { name: 'codellama', size: 7000000000, distribution: { availability: 'partial' } }
        ];

        mockFetch.mockResolvedValueOnce(createMockResponse({ models: mockModels }));

        const models = await api.models.list();
        expect(models).toEqual(mockModels);
      });

      it('should get specific model', async () => {
        const mockModel = { 
          name: 'llama2', 
          size: 4000000000,
          metrics: { inference_count: 100 }
        };

        mockFetch.mockResolvedValueOnce(createMockResponse(mockModel));

        const model = await api.models.get('llama2');
        expect(model).toEqual(mockModel);
      });
    });

    describe('Monitoring API', () => {
      it('should get performance metrics', async () => {
        const mockMetrics = {
          requests_per_second: 150,
          average_response_time: 75,
          error_rate: 0.02,
          active_connections: 50
        };

        mockFetch.mockResolvedValueOnce(createMockResponse(mockMetrics));

        const metrics = await api.monitoring.getPerformanceMetrics();
        expect(metrics).toEqual(mockMetrics);
      });

      it('should get alerts', async () => {
        const mockAlerts = [
          { id: '1', type: 'warning', message: 'High CPU usage', resolved: false },
          { id: '2', type: 'error', message: 'Node disconnected', resolved: false }
        ];

        mockFetch.mockResolvedValueOnce(createMockResponse({ alerts: mockAlerts }));

        const alerts = await api.monitoring.getAlerts();
        expect(alerts).toEqual(mockAlerts);
      });
    });
  });

  describe('Error Handling', () => {
    it('should handle HTTP errors', async () => {
      mockFetch.mockResolvedValueOnce(createMockResponse(
        { error: 'Not found' }, 
        { status: 404, ok: false }
      ));

      await expect(api.cluster.getHealth()).rejects.toThrow('Not found');
    });

    it('should handle network errors', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Network error'));

      await expect(api.cluster.getHealth()).rejects.toThrow('Network error');
    });

    it('should retry failed requests', async () => {
      // First call fails, second succeeds
      mockFetch
        .mockRejectedValueOnce(new Error('Network error'))
        .mockResolvedValueOnce(createMockResponse({ status: 'healthy' }));

      const health = await api.cluster.getHealth();
      expect(health).toEqual({ status: 'healthy' });
      expect(mockFetch).toHaveBeenCalledTimes(2);
    });

    it('should not retry 4xx errors', async () => {
      mockFetch.mockResolvedValueOnce(createMockResponse(
        { error: 'Unauthorized' }, 
        { status: 401, ok: false }
      ));

      await expect(api.auth.getProfile()).rejects.toThrow();
      expect(mockFetch).toHaveBeenCalledTimes(1); // No retry
    });
  });

  describe('WebSocket Integration', () => {
    it('should connect WebSocket', async () => {
      // Mock WebSocket constructor
      const mockWs = {
        readyState: WebSocket.OPEN,
        close: vi.fn(),
        send: vi.fn()
      };

      (global as any).WebSocket = vi.fn(() => mockWs);

      await api.connectWebSocket('test-token');
      
      expect(api.notifications.getConnectionStatus()).toBe('open');
    });

    it('should subscribe to updates', () => {
      const handlers = {
        onNotification: vi.fn(),
        onClusterUpdate: vi.fn(),
        onModelUpdate: vi.fn(),
        onAlert: vi.fn()
      };

      const unsubscribe = api.subscribeToUpdates(handlers);
      
      expect(typeof unsubscribe).toBe('function');
      
      // Test unsubscribe
      unsubscribe();
    });
  });

  describe('Singleton Pattern', () => {
    it('should return same instance from getAPIClient', () => {
      const client1 = getAPIClient();
      const client2 = getAPIClient();
      
      expect(client1).toBe(client2);
    });

    it('should create new instance with different config', () => {
      const client1 = getAPIClient();
      const client2 = getAPIClient({ baseURL: 'http://different:8080' });
      
      expect(client1).not.toBe(client2);
    });
  });
});

describe('API Error Handling', () => {
  let api: OllamaMaxAPI;

  beforeEach(() => {
    api = new OllamaMaxAPI({ 
      baseURL: 'http://localhost:8080',
      retries: 0 // No retries for these tests
    });
  });

  it('should handle timeout errors', async () => {
    // Mock slow response
    mockFetch.mockImplementation(() => 
      new Promise(resolve => setTimeout(resolve, 10000))
    );

    await expect(api.cluster.getHealth({ timeout: 100 })).rejects.toThrow();
  });

  it('should handle JSON parsing errors', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.reject(new Error('Invalid JSON'))
    });

    await expect(api.cluster.getHealth()).rejects.toThrow();
  });

  it('should handle authentication errors', async () => {
    mockFetch.mockResolvedValueOnce(createMockResponse(
      { error: 'Token expired' },
      { status: 401, ok: false }
    ));

    await expect(api.auth.getProfile()).rejects.toThrow('Token expired');
  });
});

// Helper functions
function createMockResponse(data: any, options: { status?: number; ok?: boolean } = {}) {
  return {
    ok: options.ok ?? true,
    status: options.status ?? 200,
    json: () => Promise.resolve(data),
    text: () => Promise.resolve(JSON.stringify(data)),
    blob: () => Promise.resolve(new Blob([JSON.stringify(data)])),
    headers: new Headers({
      'content-type': 'application/json'
    })
  };
}