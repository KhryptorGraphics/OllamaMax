const { describe, test, expect, beforeAll, afterAll, beforeEach, afterEach } = require('@jest/globals');
const request = require('supertest');
const WebSocket = require('ws');

// Test configuration
const TEST_CONFIG = {
  API_SERVER: process.env.TEST_API_SERVER || 'http://localhost:13100',
  WS_SERVER: process.env.TEST_WS_SERVER || 'ws://localhost:13100',
  TIMEOUT: 30000
};

describe('Distributed Engine Integration Tests', () => {
  let wsClient;

  beforeAll(async () => {
    // Wait for server to be ready
    await new Promise(resolve => setTimeout(resolve, 2000));
  });

  afterAll(async () => {
    if (wsClient && wsClient.readyState === WebSocket.OPEN) {
      wsClient.close();
    }
  });

  beforeEach(() => {
    // Reset any test state
  });

  afterEach(() => {
    if (wsClient && wsClient.readyState === WebSocket.OPEN) {
      wsClient.close();
      wsClient = null;
    }
  });

  describe('Health and Monitoring', () => {
    test('should provide basic health status', async () => {
      const response = await request(TEST_CONFIG.API_SERVER)
        .get('/api/health')
        .expect(200);

      expect(response.body).toHaveProperty('status');
      expect(response.body).toHaveProperty('cluster');
      expect(response.body).toHaveProperty('system');
      expect(response.body.cluster).toHaveProperty('totalNodes');
      expect(response.body.cluster).toHaveProperty('healthyNodes');
    });

    test('should provide detailed health information', async () => {
      const response = await request(TEST_CONFIG.API_SERVER)
        .get('/api/health/detailed')
        .expect(200);

      expect(response.body).toHaveProperty('timestamp');
      expect(response.body).toHaveProperty('cluster');
      expect(response.body).toHaveProperty('system');
      expect(response.body).toHaveProperty('metrics');
      expect(response.body).toHaveProperty('services');
      
      expect(response.body.services).toHaveProperty('healthMonitor');
      expect(response.body.services).toHaveProperty('metricsCollector');
    });

    test('should provide Prometheus metrics', async () => {
      const response = await request(TEST_CONFIG.API_SERVER)
        .get('/api/metrics/prometheus')
        .expect(200);

      expect(response.headers['content-type']).toContain('text/plain');
      expect(response.text).toContain('ollama_nodes_total');
      expect(response.text).toContain('ollama_requests_total');
    });

    test('should provide metrics summary', async () => {
      const response = await request(TEST_CONFIG.API_SERVER)
        .get('/api/metrics')
        .expect(200);

      expect(response.body).toHaveProperty('requests');
      expect(response.body).toHaveProperty('nodes');
      expect(response.body).toHaveProperty('models');
      expect(response.body).toHaveProperty('system');
      expect(response.body).toHaveProperty('latency');
    });
  });

  describe('Node Management', () => {
    test('should list available nodes', async () => {
      const response = await request(TEST_CONFIG.API_SERVER)
        .get('/api/nodes')
        .expect(200);

      expect(response.body).toHaveProperty('nodes');
      expect(Array.isArray(response.body.nodes)).toBe(true);
      expect(response.body).toHaveProperty('queueLength');
    });

    test('should add new node', async () => {
      const newNode = {
        name: 'Test Node',
        url: 'http://localhost:11435',
        mock: true
      };

      const response = await request(TEST_CONFIG.API_SERVER)
        .post('/api/nodes')
        .send(newNode)
        .expect(200);

      expect(response.body).toHaveProperty('id');
      expect(response.body).toHaveProperty('name', 'Test Node');
      expect(response.body).toHaveProperty('status');
    });

    test('should remove node', async () => {
      // First add a node
      const newNode = {
        name: 'Temp Node',
        url: 'http://localhost:11436',
        mock: true
      };

      const addResponse = await request(TEST_CONFIG.API_SERVER)
        .post('/api/nodes')
        .send(newNode)
        .expect(200);

      const nodeId = addResponse.body.id;

      // Then remove it
      const removeResponse = await request(TEST_CONFIG.API_SERVER)
        .delete(`/api/nodes/${nodeId}`)
        .expect(200);

      expect(removeResponse.body).toHaveProperty('success', true);
    });
  });

  describe('Routing Configuration', () => {
    test('should provide available routing modes', async () => {
      const response = await request(TEST_CONFIG.API_SERVER)
        .get('/api/routing/modes')
        .expect(200);

      expect(response.body).toHaveProperty('modes');
      expect(Array.isArray(response.body.modes)).toBe(true);
      
      const modeIds = response.body.modes.map(m => m.id);
      expect(modeIds).toContain('round-robin');
      expect(modeIds).toContain('least-loaded');
      expect(modeIds).toContain('auto');
      expect(modeIds).toContain('pinned');
      expect(modeIds).toContain('broadcast');
    });

    test('should provide routing statistics', async () => {
      const response = await request(TEST_CONFIG.API_SERVER)
        .get('/api/routing/stats')
        .expect(200);

      expect(response.body).toHaveProperty('totalNodes');
      expect(response.body).toHaveProperty('healthyNodes');
      expect(response.body).toHaveProperty('loadBalancer');
    });

    test('should test routing strategies', async () => {
      const testRequest = {
        strategy: 'least-loaded',
        options: {
          requiredCpu: 50,
          requiredMemory: 30
        }
      };

      const response = await request(TEST_CONFIG.API_SERVER)
        .post('/api/routing/test')
        .send(testRequest)
        .expect(200);

      expect(response.body).toHaveProperty('strategy', 'least-loaded');
      expect(response.body).toHaveProperty('selectedNodes');
      expect(Array.isArray(response.body.selectedNodes)).toBe(true);
    });

    test('should clear session affinity', async () => {
      const clearRequest = {
        sessionId: 'test-session-123'
      };

      const response = await request(TEST_CONFIG.API_SERVER)
        .post('/api/routing/clear-affinity')
        .send(clearRequest)
        .expect(200);

      expect(response.body).toHaveProperty('success', true);
      expect(response.body).toHaveProperty('message');
    });
  });

  describe('Model Management', () => {
    test('should list available models', async () => {
      const response = await request(TEST_CONFIG.API_SERVER)
        .get('/api/models')
        .expect(200);

      expect(response.body).toHaveProperty('models');
      expect(response.body).toHaveProperty('totalModels');
      expect(response.body).toHaveProperty('totalReplicas');
      expect(response.body).toHaveProperty('healthyNodes');
      expect(response.body).toHaveProperty('totalNodes');
      expect(Array.isArray(response.body.models)).toBe(true);
    });

    test('should initiate model download', async () => {
      const downloadRequest = {
        targetNodeId: null, // Auto-select
        sourceNodeId: null  // Auto-select
      };

      const response = await request(TEST_CONFIG.API_SERVER)
        .post('/api/models/test-model/download')
        .send(downloadRequest)
        .expect(200);

      expect(response.body).toHaveProperty('downloadId');
      expect(response.body).toHaveProperty('modelName', 'test-model');
      expect(response.body).toHaveProperty('status', 'initiated');
    });

    test('should track download progress', async () => {
      // First initiate a download
      const downloadRequest = {
        targetNodeId: null,
        sourceNodeId: null
      };

      const downloadResponse = await request(TEST_CONFIG.API_SERVER)
        .post('/api/models/progress-test-model/download')
        .send(downloadRequest)
        .expect(200);

      const downloadId = downloadResponse.body.downloadId;

      // Wait a bit for progress
      await new Promise(resolve => setTimeout(resolve, 500));

      // Check progress
      const progressResponse = await request(TEST_CONFIG.API_SERVER)
        .get(`/api/models/downloads/${downloadId}`)
        .expect(200);

      expect(progressResponse.body).toHaveProperty('id', downloadId);
      expect(progressResponse.body).toHaveProperty('modelName', 'progress-test-model');
      expect(progressResponse.body).toHaveProperty('status');
      expect(progressResponse.body).toHaveProperty('progress');
    });
  });

  describe('WebSocket Communication', () => {
    test('should establish WebSocket connection', (done) => {
      wsClient = new WebSocket(`${TEST_CONFIG.WS_SERVER}/chat`);

      wsClient.on('open', () => {
        expect(wsClient.readyState).toBe(WebSocket.OPEN);
        done();
      });

      wsClient.on('error', (error) => {
        done(error);
      });
    });

    test('should receive node updates via WebSocket', (done) => {
      wsClient = new WebSocket(`${TEST_CONFIG.WS_SERVER}/chat`);

      wsClient.on('open', () => {
        // Should receive initial node update
      });

      wsClient.on('message', (data) => {
        const message = JSON.parse(data.toString());

        if (message.type === 'node_update') {
          expect(message).toHaveProperty('nodes');
          expect(Array.isArray(message.nodes)).toBe(true);
          done();
        }
      });

      wsClient.on('error', (error) => {
        done(error);
      });
    });

    test('should handle inference requests via WebSocket', (done) => {
      wsClient = new WebSocket(`${TEST_CONFIG.WS_SERVER}/chat`);

      let responseReceived = false;

      wsClient.on('open', () => {
        const inferenceRequest = {
          type: 'inference',
          model: 'test-model',
          content: 'Hello, world!',
          timestamp: Date.now(),
          settings: {
            temperature: 0.7,
            maxTokens: 100,
            streaming: false
          },
          routingMode: 'auto',
          sessionId: 'test-session-ws'
        };

        wsClient.send(JSON.stringify(inferenceRequest));
      });

      wsClient.on('message', (data) => {
        const message = JSON.parse(data.toString());

        if (message.type === 'response' && !responseReceived) {
          responseReceived = true;
          expect(message).toHaveProperty('id');
          expect(message).toHaveProperty('node');
          expect(message).toHaveProperty('strategy');
          done();
        } else if (message.type === 'error') {
          // Error is acceptable if no nodes are available
          expect(message).toHaveProperty('message');
          done();
        }
      });

      wsClient.on('error', (error) => {
        done(error);
      });
    });

    test('should handle broadcast inference requests', (done) => {
      wsClient = new WebSocket(`${TEST_CONFIG.WS_SERVER}/chat`);

      let broadcastResponseReceived = false;

      wsClient.on('open', () => {
        const broadcastRequest = {
          type: 'inference',
          model: 'test-model',
          content: 'Broadcast test',
          timestamp: Date.now(),
          settings: {
            temperature: 0.7,
            maxTokens: 50,
            streaming: false
          },
          routingMode: 'broadcast',
          maxNodes: 3,
          minNodes: 1
        };

        wsClient.send(JSON.stringify(broadcastRequest));
      });

      wsClient.on('message', (data) => {
        const message = JSON.parse(data.toString());

        if (message.type === 'response' && message.strategy === 'broadcast' && !broadcastResponseReceived) {
          broadcastResponseReceived = true;
          expect(message).toHaveProperty('nodes');
          expect(Array.isArray(message.nodes)).toBe(true);
          done();
        } else if (message.type === 'error') {
          // Error is acceptable if no nodes are available
          done();
        }
      });

      wsClient.on('error', (error) => {
        done(error);
      });
    });

    test('should receive health updates via WebSocket', (done) => {
      wsClient = new WebSocket(`${TEST_CONFIG.WS_SERVER}/chat`);

      let healthUpdateReceived = false;

      wsClient.on('message', (data) => {
        const message = JSON.parse(data.toString());

        if (message.type === 'health_update' && !healthUpdateReceived) {
          healthUpdateReceived = true;
          expect(message).toHaveProperty('nodeId');
          expect(message).toHaveProperty('status');
          expect(message).toHaveProperty('metrics');
          expect(message).toHaveProperty('latency');
          done();
        }
      });

      wsClient.on('error', (error) => {
        done(error);
      });

      // Timeout after 10 seconds if no health update received
      setTimeout(() => {
        if (!healthUpdateReceived) {
          done(new Error('No health update received within timeout'));
        }
      }, 10000);
    });

    test('should receive cluster health updates', (done) => {
      wsClient = new WebSocket(`${TEST_CONFIG.WS_SERVER}/chat`);

      let clusterHealthReceived = false;

      wsClient.on('message', (data) => {
        const message = JSON.parse(data.toString());

        if (message.type === 'cluster_health' && !clusterHealthReceived) {
          clusterHealthReceived = true;
          expect(message).toHaveProperty('totalNodes');
          expect(message).toHaveProperty('healthyNodes');
          expect(message).toHaveProperty('unhealthyNodes');
          expect(message).toHaveProperty('averageLatency');
          done();
        }
      });

      wsClient.on('error', (error) => {
        done(error);
      });

      // Timeout after 15 seconds
      setTimeout(() => {
        if (!clusterHealthReceived) {
          done(new Error('No cluster health update received within timeout'));
        }
      }, 15000);
    });
  });

  describe('Monitoring Services Control', () => {
    test('should start monitoring services', async () => {
      const response = await request(TEST_CONFIG.API_SERVER)
        .get('/api/monitoring/start')
        .expect(200);

      expect(response.body).toHaveProperty('success', true);
      expect(response.body).toHaveProperty('services');
      expect(response.body.services).toHaveProperty('healthMonitor');
      expect(response.body.services).toHaveProperty('metricsCollector');
    });

    test('should stop monitoring services', async () => {
      const response = await request(TEST_CONFIG.API_SERVER)
        .get('/api/monitoring/stop')
        .expect(200);

      expect(response.body).toHaveProperty('success', true);
      expect(response.body).toHaveProperty('services');
    });
  });

  describe('Time Series Metrics', () => {
    test('should provide time series data', async () => {
      const response = await request(TEST_CONFIG.API_SERVER)
        .get('/api/metrics/timeseries/requests')
        .query({ timeRange: 3600000 }) // 1 hour
        .expect(200);

      expect(response.body).toHaveProperty('metric', 'requests');
      expect(response.body).toHaveProperty('timeRange', 3600000);
      expect(response.body).toHaveProperty('dataPoints');
      expect(response.body).toHaveProperty('data');
      expect(Array.isArray(response.body.data)).toBe(true);
    });

    test('should handle invalid time series metrics', async () => {
      const response = await request(TEST_CONFIG.API_SERVER)
        .get('/api/metrics/timeseries/invalid-metric')
        .expect(200);

      expect(response.body).toHaveProperty('metric', 'invalid-metric');
      expect(response.body).toHaveProperty('data');
      expect(Array.isArray(response.body.data)).toBe(true);
      expect(response.body.data).toHaveLength(0);
    });
  });

  describe('Node Health History', () => {
    test('should provide node health history', async () => {
      // First get available nodes
      const nodesResponse = await request(TEST_CONFIG.API_SERVER)
        .get('/api/nodes')
        .expect(200);

      if (nodesResponse.body.nodes.length > 0) {
        const nodeId = nodesResponse.body.nodes[0].id;

        const historyResponse = await request(TEST_CONFIG.API_SERVER)
          .get(`/api/health/nodes/${nodeId}`)
          .query({ limit: 10 })
          .expect(200);

        expect(historyResponse.body).toHaveProperty('nodeId', nodeId);
        expect(historyResponse.body).toHaveProperty('historyCount');
        expect(historyResponse.body).toHaveProperty('history');
        expect(Array.isArray(historyResponse.body.history)).toBe(true);
      }
    });

    test('should handle non-existent node health history', async () => {
      const response = await request(TEST_CONFIG.API_SERVER)
        .get('/api/health/nodes/non-existent-node')
        .expect(404);

      expect(response.body).toHaveProperty('error');
    });
  });
}, TEST_CONFIG.TIMEOUT);
