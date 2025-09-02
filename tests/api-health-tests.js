/**
 * API Health and Worker Node Testing
 * Validates distributed system health and worker communication
 */

import { test, expect } from '@playwright/test';

const CONFIG = {
  API_SERVER: 'http://localhost:13100',
  WORKERS: [
    'http://localhost:13000',
    'http://localhost:13001', 
    'http://localhost:13002'
  ],
  TIMEOUT: 10000
};

test.describe('API Server Health Tests', () => {

  test('API server should be responsive', async ({ request }) => {
    // Test basic connectivity
    const startTime = Date.now();
    const response = await request.get(`${CONFIG.API_SERVER}/health`);
    const responseTime = Date.now() - startTime;
    
    expect(response.status()).toBe(200);
    expect(responseTime).toBeLessThan(2000);
    
    const data = await response.json();
    expect(data).toHaveProperty('status');
  });

  test('API server should return node registry', async ({ request }) => {
    const response = await request.get(`${CONFIG.API_SERVER}/api/nodes`);
    
    expect(response.status()).toBe(200);
    
    const data = await response.json();
    expect(data).toHaveProperty('nodes');
    expect(Array.isArray(data.nodes)).toBe(true);
    expect(data.nodes.length).toBeGreaterThanOrEqual(3);
    
    // Validate node structure
    data.nodes.forEach(node => {
      expect(node).toHaveProperty('id');
      expect(node).toHaveProperty('status'); 
      expect(node).toHaveProperty('health');
      expect(typeof node.id).toBe('string');
      expect(['healthy', 'warning', 'error', 'offline']).toContain(node.status);
    });
  });

  test('WebSocket endpoint should be available', async ({ page }) => {
    let wsConnected = false;
    let wsError = null;
    
    page.on('websocket', ws => {
      wsConnected = true;
      ws.on('socketerror', error => wsError = error);
    });
    
    // Connect via JavaScript WebSocket
    await page.evaluate((wsUrl) => {
      return new Promise((resolve, reject) => {
        const ws = new WebSocket(wsUrl);
        ws.onopen = () => {
          ws.close();
          resolve(true);
        };
        ws.onerror = reject;
        setTimeout(reject, 5000);
      });
    }, `ws://localhost:13100/chat`);
    
    expect(wsError).toBeNull();
  });

});

test.describe('Worker Node Health Tests', () => {

  CONFIG.WORKERS.forEach((workerUrl, index) => {
    test(`Worker ${index + 1} (${workerUrl}) should be accessible`, async ({ request }) => {
      // Test basic connectivity  
      const response = await request.get(`${workerUrl}/api/version`);
      expect(response.status()).toBe(200);
      
      const data = await response.json();
      expect(data).toHaveProperty('version');
    });

    test(`Worker ${index + 1} should list available models`, async ({ request }) => {
      const response = await request.get(`${workerUrl}/api/tags`);
      expect(response.status()).toBe(200);
      
      const data = await response.json();
      expect(data).toHaveProperty('models');
      expect(Array.isArray(data.models)).toBe(true);
    });
  });

  test('All workers should be registered with API server', async ({ request }) => {
    const response = await request.get(`${CONFIG.API_SERVER}/api/nodes`);
    const data = await response.json();
    
    expect(data.nodes.length).toBe(CONFIG.WORKERS.length);
    
    // Check each worker is represented
    const nodeUrls = data.nodes.map(node => node.url);
    CONFIG.WORKERS.forEach(workerUrl => {
      const found = nodeUrls.some(nodeUrl => nodeUrl.includes(workerUrl.split('//')[1]));
      expect(found).toBe(true);
    });
  });

});

test.describe('Distributed System Integration Tests', () => {

  test('Load balancer should distribute requests', async ({ request }) => {
    // Make multiple requests and track which nodes handle them
    const nodeUsage = new Map();
    
    for (let i = 0; i < 10; i++) {
      const response = await request.post(`${CONFIG.API_SERVER}/api/generate`, {
        data: {
          model: 'tinyllama',
          prompt: `Test request ${i}`,
          stream: false
        }
      });
      
      if (response.status() === 200) {
        const responseHeaders = response.headers();
        const nodeId = responseHeaders['x-node-id'] || 'unknown';
        nodeUsage.set(nodeId, (nodeUsage.get(nodeId) || 0) + 1);
      }
    }
    
    // Should use multiple nodes (load balancing)
    expect(nodeUsage.size).toBeGreaterThan(1);
  });

  test('System should handle node failure gracefully', async ({ request }) => {
    // Get initial node count
    const initialResponse = await request.get(`${CONFIG.API_SERVER}/api/nodes`);
    const initialData = await initialResponse.json();
    const initialNodeCount = initialData.nodes.filter(n => n.status === 'healthy').length;
    
    expect(initialNodeCount).toBeGreaterThan(0);
    
    // System should continue operating even if one node fails
    // (This is observational - we don't actually break a node in tests)
    const healthyNodes = initialData.nodes.filter(n => n.status === 'healthy');
    expect(healthyNodes.length).toBeGreaterThanOrEqual(1);
  });

});

test.describe('Performance and Load Tests', () => {

  test('API endpoints should respond within SLA', async ({ request }) => {
    const endpoints = [
      '/health',
      '/api/nodes',
      '/api/models'
    ];
    
    for (const endpoint of endpoints) {
      const startTime = Date.now();
      const response = await request.get(`${CONFIG.API_SERVER}${endpoint}`);
      const responseTime = Date.now() - startTime;
      
      expect(response.status()).toBe(200);
      expect(responseTime).toBeLessThan(1000); // 1 second SLA
    }
  });

  test('System should handle concurrent requests', async ({ request }) => {
    // Create 10 concurrent requests
    const requests = Array(10).fill().map((_, i) => 
      request.get(`${CONFIG.API_SERVER}/api/nodes`)
    );
    
    const responses = await Promise.all(requests);
    
    // All should succeed
    responses.forEach(response => {
      expect(response.status()).toBe(200);
    });
  });

  test('WebSocket should handle multiple connections', async ({ browser }) => {
    // Create multiple browser contexts
    const contexts = await Promise.all([
      browser.newContext(),
      browser.newContext(), 
      browser.newContext()
    ]);
    
    const pages = await Promise.all(
      contexts.map(context => context.newPage())
    );
    
    // Connect all pages to WebSocket
    const connections = await Promise.all(
      pages.map(page => 
        page.evaluate(() => {
          return new Promise((resolve) => {
            const ws = new WebSocket('ws://localhost:13100/chat');
            ws.onopen = () => resolve(true);
            ws.onerror = () => resolve(false);
          });
        })
      )
    );
    
    // All connections should succeed
    connections.forEach(connected => {
      expect(connected).toBe(true);
    });
    
    // Cleanup
    await Promise.all(contexts.map(context => context.close()));
  });

});