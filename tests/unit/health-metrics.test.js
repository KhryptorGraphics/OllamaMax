const { describe, test, expect, beforeEach, afterEach, jest } = require('@jest/globals');
const EventEmitter = require('events');

// Mock node registry
class MockNodeRegistry {
  constructor() {
    this.nodes = new Map();
  }

  getNodes() {
    return Array.from(this.nodes.values());
  }

  getAllNodes() {
    return this.getNodes();
  }

  addNode(id, config) {
    const node = {
      id,
      name: config.name || `Node ${id}`,
      url: config.url || `http://localhost:${11434 + parseInt(id)}`,
      status: 'healthy',
      health: {
        load: Math.random() * 100,
        cpu: Math.random() * 100,
        memory: Math.random() * 100,
        latency: Math.random() * 1000,
        uptime: Date.now(),
        modelsLoaded: []
      },
      mock: config.mock || false
    };
    this.nodes.set(id, node);
    return node;
  }
}

// Mock HealthMonitor class
class MockHealthMonitor extends EventEmitter {
  constructor(nodeRegistry, config = {}) {
    super();
    this.nodeRegistry = nodeRegistry;
    this.config = {
      checkInterval: config.checkInterval || 1000, // Faster for testing
      healthTimeout: config.healthTimeout || 1000,
      maxRetries: config.maxRetries || 3,
      ...config
    };
    
    this.healthHistory = new Map();
    this.systemMetrics = {
      startTime: Date.now(),
      totalRequests: 0,
      successfulRequests: 0,
      failedRequests: 0,
      averageLatency: 0,
      lastUpdate: Date.now()
    };
    
    this.isRunning = false;
    this.healthCheckInterval = null;
  }

  start() {
    if (this.isRunning) return;
    
    this.isRunning = true;
    this.healthCheckInterval = setInterval(() => {
      this.performHealthChecks();
    }, this.config.checkInterval);
    
    this.emit('started');
  }

  stop() {
    if (!this.isRunning) return;
    
    this.isRunning = false;
    if (this.healthCheckInterval) {
      clearInterval(this.healthCheckInterval);
      this.healthCheckInterval = null;
    }
    
    this.emit('stopped');
  }

  async performHealthChecks() {
    const nodes = this.nodeRegistry.getNodes();
    const healthPromises = nodes.map(node => this.checkNodeHealth(node));
    
    try {
      const results = await Promise.allSettled(healthPromises);
      this.processHealthResults(nodes, results);
    } catch (error) {
      console.error('Health check batch failed:', error);
    }
  }

  async checkNodeHealth(node) {
    const startTime = Date.now();
    const healthData = {
      nodeId: node.id,
      nodeName: node.name,
      timestamp: startTime,
      status: 'healthy',
      latency: Math.floor(Math.random() * 100) + 10,
      error: null,
      metrics: {
        cpu: Math.random() * 100,
        memory: Math.random() * 100,
        load: Math.random() * 100,
        uptime: Date.now() - node.health.uptime
      }
    };

    // Simulate occasional failures
    if (Math.random() < 0.1) {
      healthData.status = 'unhealthy';
      healthData.error = 'Simulated failure';
    }

    this.updateNodeHealth(node, healthData);
    this.storeHealthHistory(node.id, healthData);
    
    return healthData;
  }

  updateNodeHealth(node, healthData) {
    node.health = {
      ...node.health,
      ...healthData.metrics,
      lastCheck: healthData.timestamp,
      latency: healthData.latency,
      status: healthData.status
    };
    
    node.status = healthData.status;
    
    this.emit('healthUpdate', {
      nodeId: node.id,
      nodeName: node.name,
      status: healthData.status,
      metrics: healthData.metrics,
      latency: healthData.latency
    });
  }

  storeHealthHistory(nodeId, healthData) {
    if (!this.healthHistory.has(nodeId)) {
      this.healthHistory.set(nodeId, []);
    }
    
    const history = this.healthHistory.get(nodeId);
    history.push(healthData);
    
    if (history.length > 100) {
      history.shift();
    }
  }

  processHealthResults(nodes, results) {
    let healthyCount = 0;
    let unhealthyCount = 0;
    let totalLatency = 0;
    let validLatencyCount = 0;

    results.forEach((result, index) => {
      if (result.status === 'fulfilled') {
        const healthData = result.value;
        if (healthData.status === 'healthy') {
          healthyCount++;
        } else {
          unhealthyCount++;
        }
        
        if (healthData.latency > 0) {
          totalLatency += healthData.latency;
          validLatencyCount++;
        }
      } else {
        unhealthyCount++;
      }
    });

    this.systemMetrics.lastUpdate = Date.now();
    if (validLatencyCount > 0) {
      this.systemMetrics.averageLatency = totalLatency / validLatencyCount;
    }

    this.emit('clusterHealth', {
      totalNodes: nodes.length,
      healthyNodes: healthyCount,
      unhealthyNodes: unhealthyCount,
      averageLatency: this.systemMetrics.averageLatency,
      timestamp: Date.now()
    });
  }

  getNodeHealthHistory(nodeId, limit = 50) {
    const history = this.healthHistory.get(nodeId) || [];
    return history.slice(-limit);
  }

  getClusterHealthSummary() {
    const nodes = this.nodeRegistry.getAllNodes();
    return {
      totalNodes: nodes.length,
      healthyNodes: nodes.filter(n => n.status === 'healthy').length,
      unhealthyNodes: nodes.filter(n => n.status !== 'healthy').length,
      unknownNodes: 0,
      averageLatency: this.systemMetrics.averageLatency,
      nodeDetails: nodes.map(node => ({
        id: node.id,
        name: node.name,
        status: node.status,
        lastCheck: node.health.lastCheck,
        latency: node.health.latency || 0,
        cpu: node.health.cpu || 0,
        memory: node.health.memory || 0,
        load: node.health.load || 0
      }))
    };
  }

  recordRequest(success = true, latency = 0) {
    this.systemMetrics.totalRequests++;
    if (success) {
      this.systemMetrics.successfulRequests++;
    } else {
      this.systemMetrics.failedRequests++;
    }

    if (latency > 0) {
      const alpha = 0.1;
      this.systemMetrics.averageLatency = 
        (alpha * latency) + ((1 - alpha) * this.systemMetrics.averageLatency);
    }
  }
}

describe('Health Monitor Tests', () => {
  let nodeRegistry;
  let healthMonitor;

  beforeEach(() => {
    nodeRegistry = new MockNodeRegistry();
    healthMonitor = new MockHealthMonitor(nodeRegistry, {
      checkInterval: 100, // Fast for testing
      healthTimeout: 500
    });

    // Add test nodes
    nodeRegistry.addNode('node1', { name: 'Test Node 1', mock: true });
    nodeRegistry.addNode('node2', { name: 'Test Node 2', mock: true });
    nodeRegistry.addNode('node3', { name: 'Test Node 3', mock: true });
  });

  afterEach(() => {
    if (healthMonitor.isRunning) {
      healthMonitor.stop();
    }
  });

  describe('Health Monitor Lifecycle', () => {
    test('should start and stop correctly', () => {
      expect(healthMonitor.isRunning).toBe(false);

      healthMonitor.start();
      expect(healthMonitor.isRunning).toBe(true);

      healthMonitor.stop();
      expect(healthMonitor.isRunning).toBe(false);
    });

    test('should emit started and stopped events', (done) => {
      let eventCount = 0;

      healthMonitor.on('started', () => {
        eventCount++;
        expect(healthMonitor.isRunning).toBe(true);
      });

      healthMonitor.on('stopped', () => {
        eventCount++;
        expect(healthMonitor.isRunning).toBe(false);
        expect(eventCount).toBe(2);
        done();
      });

      healthMonitor.start();
      setTimeout(() => healthMonitor.stop(), 50);
    });

    test('should not start twice', () => {
      healthMonitor.start();
      const firstInterval = healthMonitor.healthCheckInterval;

      healthMonitor.start(); // Try to start again
      expect(healthMonitor.healthCheckInterval).toBe(firstInterval);

      healthMonitor.stop();
    });
  });

  describe('Health Checks', () => {
    test('should perform health checks on all nodes', async () => {
      const nodes = nodeRegistry.getNodes();
      expect(nodes.length).toBe(3);

      await healthMonitor.performHealthChecks();

      // All nodes should have updated health data
      nodes.forEach(node => {
        expect(node.health.lastCheck).toBeDefined();
        expect(node.health.latency).toBeDefined();
        expect(['healthy', 'unhealthy']).toContain(node.status);
      });
    });

    test('should emit health update events', (done) => {
      let updateCount = 0;

      healthMonitor.on('healthUpdate', (data) => {
        updateCount++;
        expect(data).toHaveProperty('nodeId');
        expect(data).toHaveProperty('status');
        expect(data).toHaveProperty('metrics');
        expect(data).toHaveProperty('latency');

        if (updateCount === 3) { // All 3 nodes updated
          done();
        }
      });

      healthMonitor.performHealthChecks();
    });

    test('should emit cluster health events', (done) => {
      healthMonitor.on('clusterHealth', (data) => {
        expect(data).toHaveProperty('totalNodes');
        expect(data).toHaveProperty('healthyNodes');
        expect(data).toHaveProperty('unhealthyNodes');
        expect(data).toHaveProperty('averageLatency');
        expect(data.totalNodes).toBe(3);
        done();
      });

      healthMonitor.performHealthChecks();
    });
  });

  describe('Health History', () => {
    test('should store health history for nodes', async () => {
      await healthMonitor.performHealthChecks();

      const history = healthMonitor.getNodeHealthHistory('node1');
      expect(history.length).toBeGreaterThan(0);
      expect(history[0]).toHaveProperty('nodeId', 'node1');
      expect(history[0]).toHaveProperty('timestamp');
      expect(history[0]).toHaveProperty('status');
    });

    test('should limit history size', async () => {
      // Perform many health checks
      for (let i = 0; i < 150; i++) {
        await healthMonitor.checkNodeHealth(nodeRegistry.nodes.get('node1'));
      }

      const history = healthMonitor.getNodeHealthHistory('node1');
      expect(history.length).toBeLessThanOrEqual(100);
    });

    test('should return limited history', async () => {
      // Add some history
      for (let i = 0; i < 20; i++) {
        await healthMonitor.checkNodeHealth(nodeRegistry.nodes.get('node1'));
      }

      const limitedHistory = healthMonitor.getNodeHealthHistory('node1', 5);
      expect(limitedHistory.length).toBe(5);
    });
  });

  describe('Cluster Health Summary', () => {
    test('should provide cluster health summary', async () => {
      await healthMonitor.performHealthChecks();

      const summary = healthMonitor.getClusterHealthSummary();

      expect(summary).toHaveProperty('totalNodes', 3);
      expect(summary).toHaveProperty('healthyNodes');
      expect(summary).toHaveProperty('unhealthyNodes');
      expect(summary).toHaveProperty('nodeDetails');
      expect(summary.nodeDetails).toHaveLength(3);

      summary.nodeDetails.forEach(detail => {
        expect(detail).toHaveProperty('id');
        expect(detail).toHaveProperty('name');
        expect(detail).toHaveProperty('status');
      });
    });
  });

  describe('Request Metrics', () => {
    test('should record successful requests', () => {
      const initialSuccessful = healthMonitor.systemMetrics.successfulRequests;
      const initialTotal = healthMonitor.systemMetrics.totalRequests;

      healthMonitor.recordRequest(true, 100);

      expect(healthMonitor.systemMetrics.successfulRequests).toBe(initialSuccessful + 1);
      expect(healthMonitor.systemMetrics.totalRequests).toBe(initialTotal + 1);
      expect(healthMonitor.systemMetrics.averageLatency).toBeGreaterThan(0);
    });

    test('should record failed requests', () => {
      const initialFailed = healthMonitor.systemMetrics.failedRequests;
      const initialTotal = healthMonitor.systemMetrics.totalRequests;

      healthMonitor.recordRequest(false, 200);

      expect(healthMonitor.systemMetrics.failedRequests).toBe(initialFailed + 1);
      expect(healthMonitor.systemMetrics.totalRequests).toBe(initialTotal + 1);
    });

    test('should update average latency', () => {
      healthMonitor.recordRequest(true, 100);
      const firstLatency = healthMonitor.systemMetrics.averageLatency;

      healthMonitor.recordRequest(true, 200);
      const secondLatency = healthMonitor.systemMetrics.averageLatency;

      expect(secondLatency).toBeGreaterThan(firstLatency);
    });
  });

  describe('Automatic Health Monitoring', () => {
    test('should perform periodic health checks when started', (done) => {
      let healthUpdateCount = 0;

      healthMonitor.on('healthUpdate', () => {
        healthUpdateCount++;
        if (healthUpdateCount >= 6) { // 2 rounds of 3 nodes
          healthMonitor.stop();
          expect(healthUpdateCount).toBeGreaterThanOrEqual(6);
          done();
        }
      });

      healthMonitor.start();
    }, 10000); // Longer timeout for this test

    test('should stop periodic checks when stopped', (done) => {
      let healthUpdateCount = 0;

      healthMonitor.on('healthUpdate', () => {
        healthUpdateCount++;
      });

      healthMonitor.start();

      setTimeout(() => {
        const countAtStop = healthUpdateCount;
        healthMonitor.stop();

        setTimeout(() => {
          expect(healthUpdateCount).toBe(countAtStop); // Should not increase after stop
          done();
        }, 200);
      }, 300);
    });
  });

  describe('Error Handling', () => {
    test('should handle node health check failures gracefully', async () => {
      // Create a node that will fail health checks
      const failingNode = {
        id: 'failing-node',
        name: 'Failing Node',
        status: 'unknown',
        health: {}
      };

      // Mock the checkNodeHealth to throw an error for this node
      const originalCheck = healthMonitor.checkNodeHealth;
      healthMonitor.checkNodeHealth = jest.fn().mockImplementation((node) => {
        if (node.id === 'failing-node') {
          throw new Error('Health check failed');
        }
        return originalCheck.call(healthMonitor, node);
      });

      nodeRegistry.nodes.set('failing-node', failingNode);

      // Should not throw an error
      await expect(healthMonitor.performHealthChecks()).resolves.not.toThrow();

      // Restore original method
      healthMonitor.checkNodeHealth = originalCheck;
    });
  });
});
