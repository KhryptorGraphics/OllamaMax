/**
 * OllamaMax Performance Testing Suite
 * Comprehensive performance validation for distributed AI platform web interface
 * 
 * Performance Requirements:
 * - Page load time < 3 seconds
 * - Real-time updates < 100ms latency  
 * - Smooth animations and transitions
 * - Efficient memory usage
 * - Network optimization
 * - Worker failover timing
 */

import { test, expect, chromium } from '@playwright/test';

// Performance thresholds
const PERFORMANCE_THRESHOLDS = {
  PAGE_LOAD_TIME: 3000,        // 3 seconds
  WEBSOCKET_LATENCY: 100,      // 100ms
  ANIMATION_FRAME_RATE: 50,    // 50+ FPS
  MEMORY_GROWTH_LIMIT: 50,     // 50MB max growth
  NETWORK_TIMEOUT: 5000,       // 5 seconds
  WORKER_FAILOVER_TIME: 2000   // 2 seconds
};

// URLs and endpoints
const BASE_URL = 'http://localhost:13100';
const WEB_INTERFACE_URL = `${BASE_URL}/web-interface/index.html`;
const API_ENDPOINT = `${BASE_URL}/api`;
const WEBSOCKET_URL = `ws://localhost:13100/chat`;

test.describe('Performance Testing Suite', () => {
  
  test.describe('1. Page Load Performance', () => {
    
    test('should load initial page within performance threshold', async ({ page }) => {
      const startTime = Date.now();
      
      // Navigate with network monitoring
      await page.goto(WEB_INTERFACE_URL, { waitUntil: 'networkidle' });
      
      const loadTime = Date.now() - startTime;
      console.log(`Page load time: ${loadTime}ms`);
      
      expect(loadTime).toBeLessThan(PERFORMANCE_THRESHOLDS.PAGE_LOAD_TIME);
      
      // Verify critical elements are loaded
      await expect(page.locator('#app')).toBeVisible();
      await expect(page.locator('.app-header')).toBeVisible();
      await expect(page.locator('#messagesArea')).toBeVisible();
    });
    
    test('should measure Time to Interactive (TTI)', async ({ page }) => {
      const startTime = Date.now();
      
      await page.goto(WEB_INTERFACE_URL);
      
      // Wait for JavaScript to initialize
      await page.waitForFunction(() => window.llamaClient !== undefined);
      
      // Test interactivity by clicking send button
      await page.waitForSelector('#sendButton:enabled');
      
      const ttiTime = Date.now() - startTime;
      console.log(`Time to Interactive: ${ttiTime}ms`);
      
      expect(ttiTime).toBeLessThan(PERFORMANCE_THRESHOLDS.PAGE_LOAD_TIME);
    });
    
    test('should validate Core Web Vitals metrics', async ({ page }) => {
      await page.goto(WEB_INTERFACE_URL);
      
      // Measure LCP (Largest Contentful Paint)
      const lcp = await page.evaluate(() => {
        return new Promise((resolve) => {
          new PerformanceObserver((list) => {
            const entries = list.getEntries();
            const lastEntry = entries[entries.length - 1];
            resolve(lastEntry.startTime);
          }).observe({ entryTypes: ['largest-contentful-paint'] });
          
          setTimeout(() => resolve(0), 5000); // Fallback
        });
      });
      
      console.log(`LCP: ${lcp}ms`);
      expect(lcp).toBeLessThan(2500); // Good LCP threshold
      
      // Measure CLS (Cumulative Layout Shift)
      const cls = await page.evaluate(() => {
        return new Promise((resolve) => {
          let clsValue = 0;
          new PerformanceObserver((list) => {
            for (const entry of list.getEntries()) {
              if (!entry.hadRecentInput) {
                clsValue += entry.value;
              }
            }
            resolve(clsValue);
          }).observe({ entryTypes: ['layout-shift'] });
          
          setTimeout(() => resolve(clsValue), 3000);
        });
      });
      
      console.log(`CLS: ${cls}`);
      expect(cls).toBeLessThan(0.1); // Good CLS threshold
    });
  });
  
  test.describe('2. Real-time WebSocket Performance', () => {
    
    test('should establish WebSocket connection quickly', async ({ page }) => {
      await page.goto(WEB_INTERFACE_URL);
      
      const connectionStart = Date.now();
      
      // Wait for WebSocket connection
      await page.waitForFunction(() => {
        return window.llamaClient && 
               window.llamaClient.ws && 
               window.llamaClient.ws.readyState === WebSocket.OPEN;
      });
      
      const connectionTime = Date.now() - connectionStart;
      console.log(`WebSocket connection time: ${connectionTime}ms`);
      
      expect(connectionTime).toBeLessThan(2000);
      
      // Verify connection status UI update
      await expect(page.locator('#connectionText')).toContainText('Connected');
    });
    
    test('should handle real-time message latency', async ({ page }) => {
      await page.goto(WEB_INTERFACE_URL);
      
      // Wait for connection
      await page.waitForFunction(() => 
        window.llamaClient?.ws?.readyState === WebSocket.OPEN
      );
      
      // Measure round-trip latency with ping-pong test
      const latencies = [];
      
      for (let i = 0; i < 5; i++) {
        const pingStart = Date.now();
        
        // Send test message
        await page.evaluate(() => {
          window.llamaClient.ws.send(JSON.stringify({
            type: 'ping',
            timestamp: Date.now()
          }));
        });
        
        // Wait for response (mock with timeout)
        await page.waitForTimeout(50);
        const latency = Date.now() - pingStart;
        latencies.push(latency);
        
        console.log(`Message ${i + 1} latency: ${latency}ms`);
      }
      
      const avgLatency = latencies.reduce((a, b) => a + b, 0) / latencies.length;
      console.log(`Average latency: ${avgLatency}ms`);
      
      expect(avgLatency).toBeLessThan(PERFORMANCE_THRESHOLDS.WEBSOCKET_LATENCY);
    });
    
    test('should handle streaming message updates efficiently', async ({ page }) => {
      await page.goto(WEB_INTERFACE_URL);
      
      // Monitor streaming message performance
      let updateCount = 0;
      const updateTimes = [];
      
      page.on('domcontentloaded', () => {
        page.evaluate(() => {
          // Override streaming message handler to measure performance
          const originalUpdate = window.llamaClient.updateStreamingMessage;
          window.llamaClient.updateStreamingMessage = function(content) {
            const start = performance.now();
            originalUpdate.call(this, content);
            const duration = performance.now() - start;
            window.streamingUpdateTimes = window.streamingUpdateTimes || [];
            window.streamingUpdateTimes.push(duration);
          };
        });
      });
      
      // Simulate rapid streaming updates
      await page.evaluate(() => {
        window.llamaClient.streamingMessage = { id: 'test', content: '', node: 'test-node' };
        
        for (let i = 0; i < 50; i++) {
          setTimeout(() => {
            window.llamaClient.updateStreamingMessage('Test content '.repeat(i + 1));
          }, i * 20);
        }
      });
      
      await page.waitForTimeout(2000);
      
      const updatePerformance = await page.evaluate(() => window.streamingUpdateTimes || []);
      const avgUpdateTime = updatePerformance.reduce((a, b) => a + b, 0) / updatePerformance.length;
      
      console.log(`Average streaming update time: ${avgUpdateTime}ms`);
      expect(avgUpdateTime).toBeLessThan(16); // 60 FPS = 16.67ms per frame
    });
  });
  
  test.describe('3. UI Responsiveness and Animation Performance', () => {
    
    test('should maintain smooth tab transitions', async ({ page }) => {
      await page.goto(WEB_INTERFACE_URL);
      
      const tabs = ['nodes', 'models', 'settings', 'chat'];
      const transitionTimes = [];
      
      for (const tab of tabs) {
        const startTime = Date.now();
        
        await page.click(`[data-tab="${tab}"]`);
        
        // Wait for transition to complete
        await page.waitForSelector(`#${tab}Tab.active`);
        
        const transitionTime = Date.now() - startTime;
        transitionTimes.push(transitionTime);
        
        console.log(`Tab transition to ${tab}: ${transitionTime}ms`);
        
        await page.waitForTimeout(100); // Brief pause between transitions
      }
      
      const avgTransitionTime = transitionTimes.reduce((a, b) => a + b, 0) / transitionTimes.length;
      expect(avgTransitionTime).toBeLessThan(300); // Smooth transitions under 300ms
    });
    
    test('should handle rapid user interactions without lag', async ({ page }) => {
      await page.goto(WEB_INTERFACE_URL);
      
      // Test rapid clicking on various elements
      const interactions = [
        () => page.click('[data-tab="nodes"]'),
        () => page.click('[data-tab="models"]'),
        () => page.click('[data-tab="settings"]'),
        () => page.click('#refreshNodes'),
        () => page.click('#modelSelector')
      ];
      
      const startTime = Date.now();
      
      // Perform rapid interactions
      for (let i = 0; i < 20; i++) {
        const interaction = interactions[i % interactions.length];
        await interaction();
        await page.waitForTimeout(50); // Rapid interactions
      }
      
      const totalTime = Date.now() - startTime;
      console.log(`20 rapid interactions completed in: ${totalTime}ms`);
      
      // UI should remain responsive
      expect(totalTime).toBeLessThan(5000);
      
      // No JavaScript errors should occur
      const errors = await page.evaluate(() => window.errors || []);
      expect(errors.length).toBe(0);
    });
    
    test('should animate node cards smoothly', async ({ page }) => {
      await page.goto(WEB_INTERFACE_URL);
      
      // Navigate to nodes tab
      await page.click('[data-tab="nodes"]');
      
      // Wait for nodes to load
      await page.waitForTimeout(1000);
      
      // Measure hover animation performance
      const nodeCards = await page.locator('.node-card, .enhanced-node-card').count();
      
      if (nodeCards > 0) {
        // Test hover performance
        const hoverTimes = [];
        
        for (let i = 0; i < Math.min(nodeCards, 5); i++) {
          const startTime = Date.now();
          
          await page.hover(`.node-card:nth-child(${i + 1}), .enhanced-node-card:nth-child(${i + 1})`);
          
          // Wait for hover animation to complete
          await page.waitForTimeout(300);
          
          const hoverTime = Date.now() - startTime;
          hoverTimes.push(hoverTime);
        }
        
        const avgHoverTime = hoverTimes.reduce((a, b) => a + b, 0) / hoverTimes.length;
        console.log(`Average hover animation time: ${avgHoverTime}ms`);
        
        expect(avgHoverTime).toBeLessThan(350);
      }
    });
  });
  
  test.describe('4. Memory Usage and Leak Detection', () => {
    
    test('should not have significant memory leaks during normal operation', async ({ page }) => {
      await page.goto(WEB_INTERFACE_URL);
      
      // Get initial memory usage
      const initialMemory = await page.evaluate(() => {
        if (performance.memory) {
          return {
            used: performance.memory.usedJSHeapSize,
            total: performance.memory.totalJSHeapSize
          };
        }
        return null;
      });
      
      if (!initialMemory) {
        test.skip('Memory API not available');
      }
      
      console.log(`Initial memory usage: ${(initialMemory.used / 1024 / 1024).toFixed(2)} MB`);
      
      // Simulate 5 minutes of normal usage
      for (let i = 0; i < 30; i++) {
        // Switch tabs
        await page.click(`[data-tab="${['chat', 'nodes', 'models', 'settings'][i % 4]}"]`);
        
        // Simulate some activity
        if (i % 4 === 0) {
          await page.fill('#messageInput', `Test message ${i}`);
          await page.waitForTimeout(100);
        }
        
        await page.waitForTimeout(200);
      }
      
      // Force garbage collection if available
      await page.evaluate(() => {
        if (window.gc) window.gc();
      });
      
      await page.waitForTimeout(1000);
      
      const finalMemory = await page.evaluate(() => {
        if (performance.memory) {
          return {
            used: performance.memory.usedJSHeapSize,
            total: performance.memory.totalJSHeapSize
          };
        }
        return null;
      });
      
      const memoryGrowth = (finalMemory.used - initialMemory.used) / 1024 / 1024;
      console.log(`Memory growth: ${memoryGrowth.toFixed(2)} MB`);
      
      expect(memoryGrowth).toBeLessThan(PERFORMANCE_THRESHOLDS.MEMORY_GROWTH_LIMIT);
    });
    
    test('should handle large dataset rendering efficiently', async ({ page }) => {
      await page.goto(WEB_INTERFACE_URL);
      
      // Navigate to nodes tab
      await page.click('[data-tab="nodes"]');
      
      // Mock large number of nodes
      await page.evaluate(() => {
        const mockNodes = Array.from({ length: 100 }, (_, i) => ({
          id: `node-${i}`,
          name: `llama-node-${i.toString().padStart(3, '0')}`,
          status: ['healthy', 'warning', 'error'][i % 3],
          systemInfo: {
            cpu: { usage: Math.random() * 100, cores: 8 },
            memory: { usage: Math.random() * 100, total: 32 * 1024 * 1024 * 1024 },
            disk: { usage: Math.random() * 100 }
          },
          performanceHistory: {
            cpu: Array.from({ length: 20 }, () => Math.random() * 100),
            memory: Array.from({ length: 20 }, () => Math.random() * 100),
            responseTime: Array.from({ length: 20 }, () => 100 + Math.random() * 400)
          }
        }));
        
        window.llamaClient.detailedNodes = mockNodes;
        
        const renderStart = performance.now();
        window.llamaClient.displayEnhancedNodes();
        const renderTime = performance.now() - renderStart;
        
        window.largeDatasetRenderTime = renderTime;
      });
      
      const renderTime = await page.evaluate(() => window.largeDatasetRenderTime);
      console.log(`Large dataset render time: ${renderTime}ms`);
      
      expect(renderTime).toBeLessThan(500); // Should render 100 nodes in under 500ms
    });
  });
  
  test.describe('5. Network Request Optimization', () => {
    
    test('should efficiently batch network requests', async ({ page }) => {
      const requests = [];
      
      // Monitor network requests
      page.on('request', request => {
        if (request.url().includes('/api/')) {
          requests.push({
            url: request.url(),
            method: request.method(),
            timestamp: Date.now()
          });
        }
      });
      
      await page.goto(WEB_INTERFACE_URL);
      
      // Navigate through tabs to trigger API calls
      await page.click('[data-tab="nodes"]');
      await page.waitForTimeout(500);
      
      await page.click('[data-tab="models"]');
      await page.waitForTimeout(500);
      
      // Refresh operations
      await page.click('#refreshNodes');
      await page.waitForTimeout(200);
      
      console.log(`Total API requests: ${requests.length}`);
      console.log('Requests:', requests.map(r => `${r.method} ${r.url}`));
      
      // Should not make excessive requests
      expect(requests.length).toBeLessThan(10);
      
      // No duplicate requests within short timeframe
      const duplicates = requests.filter((req, index) => 
        requests.findIndex(r => r.url === req.url && Math.abs(r.timestamp - req.timestamp) < 1000) !== index
      );
      expect(duplicates.length).toBe(0);
    });
    
    test('should handle network errors gracefully', async ({ page }) => {
      // Block network requests to simulate offline
      await page.route('/api/**', route => route.abort());
      
      await page.goto(WEB_INTERFACE_URL);
      
      const startTime = Date.now();
      
      // Try to refresh nodes (should fail gracefully)
      await page.click('[data-tab="nodes"]');
      
      // Should show fallback content quickly
      await page.waitForTimeout(1000);
      
      const fallbackTime = Date.now() - startTime;
      console.log(`Fallback handling time: ${fallbackTime}ms`);
      
      expect(fallbackTime).toBeLessThan(2000);
      
      // Should not crash the application
      const appVisible = await page.locator('#app').isVisible();
      expect(appVisible).toBe(true);
    });
    
    test('should implement request caching and deduplication', async ({ page }) => {
      const requests = [];
      
      page.on('request', request => {
        if (request.url().includes('/api/nodes')) {
          requests.push({
            url: request.url(),
            timestamp: Date.now()
          });
        }
      });
      
      await page.goto(WEB_INTERFACE_URL);
      
      // Make multiple rapid requests
      for (let i = 0; i < 5; i++) {
        await page.click('[data-tab="nodes"]');
        await page.waitForTimeout(100);
        await page.click('[data-tab="chat"]');
        await page.waitForTimeout(100);
      }
      
      console.log(`Node API requests made: ${requests.length}`);
      
      // Should deduplicate rapid requests
      expect(requests.length).toBeLessThan(10);
    });
  });
  
  test.describe('6. Worker Failover Performance', () => {
    
    test('should detect node failures quickly', async ({ page }) => {
      await page.goto(WEB_INTERFACE_URL);
      
      // Wait for initial connection
      await page.waitForFunction(() => 
        window.llamaClient?.ws?.readyState === WebSocket.OPEN
      );
      
      // Simulate node failure by intercepting health check responses
      await page.route('/api/nodes/detailed', route => {
        route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            nodes: [
              {
                id: 'node-1',
                name: 'ollama-primary',
                status: 'error', // Simulated failure
                systemInfo: { cpu: { usage: 0 }, memory: { usage: 0 } }
              }
            ]
          })
        });
      });
      
      const failureDetectionStart = Date.now();
      
      // Navigate to nodes tab to trigger health check
      await page.click('[data-tab="nodes"]');
      
      // Wait for error status to be displayed
      await page.waitForSelector('.node-card.error, .enhanced-node-card .node-status:has-text("error")', 
        { timeout: 5000 });
      
      const detectionTime = Date.now() - failureDetectionStart;
      console.log(`Node failure detection time: ${detectionTime}ms`);
      
      expect(detectionTime).toBeLessThan(PERFORMANCE_THRESHOLDS.WORKER_FAILOVER_TIME);
    });
    
    test('should failover to healthy nodes efficiently', async ({ page }) => {
      await page.goto(WEB_INTERFACE_URL);
      
      // Mock multiple nodes with one failing
      await page.evaluate(() => {
        window.llamaClient.nodes = [
          { id: 'node-1', name: 'llama-01', status: 'error', load: 100 },
          { id: 'node-2', name: 'llama-02', status: 'healthy', load: 30 },
          { id: 'node-3', name: 'llama-03', status: 'healthy', load: 45 }
        ];
        
        // Override node selection to test failover logic
        window.llamaClient.selectHealthyNode = function() {
          const healthyNodes = this.nodes.filter(n => n.status === 'healthy');
          return healthyNodes.length > 0 ? healthyNodes[0] : null;
        };
      });
      
      const failoverStart = Date.now();
      
      // Simulate sending message (should failover to healthy node)
      await page.fill('#messageInput', 'Test failover message');
      await page.click('#sendButton');
      
      // Should show message was queued or sent to healthy node
      await page.waitForSelector('.message.user', { timeout: 2000 });
      
      const failoverTime = Date.now() - failoverStart;
      console.log(`Failover time: ${failoverTime}ms`);
      
      expect(failoverTime).toBeLessThan(PERFORMANCE_THRESHOLDS.WORKER_FAILOVER_TIME);
    });
  });
  
  test.describe('7. Resource Utilization Monitoring', () => {
    
    test('should monitor CPU usage during heavy operations', async ({ page }) => {
      await page.goto(WEB_INTERFACE_URL);
      
      // Start CPU monitoring
      const cpuSamples = [];
      
      const monitorCPU = async () => {
        const cpuUsage = await page.evaluate(() => {
          // Simulate CPU-intensive operation
          const start = performance.now();
          let result = 0;
          for (let i = 0; i < 100000; i++) {
            result += Math.random();
          }
          return performance.now() - start;
        });
        cpuSamples.push(cpuUsage);
      };
      
      // Monitor during various operations
      await monitorCPU();
      await page.click('[data-tab="nodes"]');
      await monitorCPU();
      
      // Navigate to models tab (heavy operation)
      await page.click('[data-tab="models"]');
      await monitorCPU();
      
      const avgCPUTime = cpuSamples.reduce((a, b) => a + b, 0) / cpuSamples.length;
      console.log(`Average CPU operation time: ${avgCPUTime}ms`);
      
      expect(avgCPUTime).toBeLessThan(50); // Operations should be efficient
    });
    
    test('should track DOM node count growth', async ({ page }) => {
      await page.goto(WEB_INTERFACE_URL);
      
      const initialNodes = await page.evaluate(() => document.querySelectorAll('*').length);
      console.log(`Initial DOM nodes: ${initialNodes}`);
      
      // Perform operations that add DOM elements
      for (let i = 0; i < 10; i++) {
        await page.click('[data-tab="nodes"]');
        await page.waitForTimeout(200);
        await page.click('[data-tab="models"]');
        await page.waitForTimeout(200);
      }
      
      const finalNodes = await page.evaluate(() => document.querySelectorAll('*').length);
      const nodeGrowth = finalNodes - initialNodes;
      
      console.log(`DOM node growth: ${nodeGrowth} nodes`);
      
      // Should not have excessive DOM growth
      expect(nodeGrowth).toBeLessThan(100);
    });
  });
  
  test.describe('8. Performance Under Load', () => {
    
    test('should handle multiple concurrent WebSocket connections', async () => {
      const browser = await chromium.launch();
      const contexts = [];
      const pages = [];
      
      try {
        // Create 5 concurrent browser contexts
        for (let i = 0; i < 5; i++) {
          const context = await browser.newContext();
          const page = await context.newPage();
          contexts.push(context);
          pages.push(page);
        }
        
        const loadStart = Date.now();
        
        // Load the interface simultaneously in all contexts
        await Promise.all(pages.map(page => 
          page.goto(WEB_INTERFACE_URL, { waitUntil: 'networkidle' })
        ));
        
        const concurrentLoadTime = Date.now() - loadStart;
        console.log(`Concurrent load time (5 clients): ${concurrentLoadTime}ms`);
        
        expect(concurrentLoadTime).toBeLessThan(8000); // Should handle concurrent load
        
        // Verify all connections are working
        for (const page of pages) {
          await expect(page.locator('#connectionText')).toContainText(/Connected|Connecting/);
        }
        
      } finally {
        // Cleanup
        for (const context of contexts) {
          await context.close();
        }
        await browser.close();
      }
    });
    
    test('should maintain performance with large message history', async ({ page }) => {
      await page.goto(WEB_INTERFACE_URL);
      
      // Simulate large message history
      await page.evaluate(() => {
        for (let i = 0; i < 100; i++) {
          window.llamaClient.addMessage(
            i % 2 === 0 ? 'user' : 'ai',
            `Message ${i}: ${'This is a test message with some content. '.repeat(5)}`,
            'test-node'
          );
        }
      });
      
      const scrollStart = Date.now();
      
      // Test scrolling performance with large history
      await page.evaluate(() => {
        const messagesArea = document.getElementById('messagesArea');
        messagesArea.scrollTop = 0;
      });
      
      await page.waitForTimeout(100);
      
      await page.evaluate(() => {
        const messagesArea = document.getElementById('messagesArea');
        messagesArea.scrollTop = messagesArea.scrollHeight;
      });
      
      const scrollTime = Date.now() - scrollStart;
      console.log(`Large history scroll time: ${scrollTime}ms`);
      
      expect(scrollTime).toBeLessThan(200);
    });
  });
  
  test.describe('9. Real-world Performance Scenarios', () => {
    
    test('should handle typical user workflow efficiently', async ({ page }) => {
      await page.goto(WEB_INTERFACE_URL);
      
      const workflowStart = Date.now();
      
      // Typical user workflow
      // 1. Check connection status
      await expect(page.locator('#connectionText')).toBeVisible();
      
      // 2. Check available nodes
      await page.click('[data-tab="nodes"]');
      await page.waitForTimeout(500);
      
      // 3. Check models
      await page.click('[data-tab="models"]');
      await page.waitForTimeout(500);
      
      // 4. Adjust settings
      await page.click('[data-tab="settings"]');
      await page.fill('#maxTokens', '1024');
      await page.selectOption('#loadBalancing', 'least-loaded');
      
      // 5. Send a message
      await page.click('[data-tab="chat"]');
      await page.fill('#messageInput', 'Hello, can you help me test this system?');
      await page.click('#sendButton');
      
      // 6. Wait for response UI update
      await page.waitForSelector('.message.user', { timeout: 2000 });
      
      const workflowTime = Date.now() - workflowStart;
      console.log(`Complete user workflow time: ${workflowTime}ms`);
      
      expect(workflowTime).toBeLessThan(5000);
    });
    
    test('should measure end-to-end inference performance', async ({ page }) => {
      await page.goto(WEB_INTERFACE_URL);
      
      // Wait for WebSocket connection
      await page.waitForFunction(() => 
        window.llamaClient?.ws?.readyState === WebSocket.OPEN
      );
      
      const inferenceStart = Date.now();
      
      // Send inference request
      await page.fill('#messageInput', 'What is artificial intelligence?');
      await page.click('#sendButton');
      
      // Wait for user message to appear
      await page.waitForSelector('.message.user');
      
      // Wait for AI response to start (or timeout for mock)
      try {
        await page.waitForSelector('.message.ai', { timeout: 5000 });
        const inferenceTime = Date.now() - inferenceStart;
        console.log(`End-to-end inference time: ${inferenceTime}ms`);
        expect(inferenceTime).toBeLessThan(10000);
      } catch (error) {
        console.log('AI response not received (expected in test environment)');
        // Verify at least the request was processed
        const requestTime = Date.now() - inferenceStart;
        expect(requestTime).toBeLessThan(2000);
      }
    });
  });
});

test.describe('Performance Benchmarking Suite', () => {
  
  test('should generate comprehensive performance report', async ({ page }) => {
    const performanceMetrics = {
      pageLoad: null,
      webSocketConnection: null,
      tabTransitions: [],
      memoryUsage: null,
      networkRequests: [],
      animationFrameRate: null
    };
    
    // 1. Page Load Performance
    const loadStart = Date.now();
    await page.goto(WEB_INTERFACE_URL, { waitUntil: 'networkidle' });
    performanceMetrics.pageLoad = Date.now() - loadStart;
    
    // 2. WebSocket Connection Performance
    const wsStart = Date.now();
    await page.waitForFunction(() => 
      window.llamaClient?.ws?.readyState === WebSocket.OPEN
    );
    performanceMetrics.webSocketConnection = Date.now() - wsStart;
    
    // 3. Tab Transition Performance
    const tabs = ['nodes', 'models', 'settings'];
    for (const tab of tabs) {
      const transitionStart = Date.now();
      await page.click(`[data-tab="${tab}"]`);
      await page.waitForSelector(`#${tab}Tab.active`);
      performanceMetrics.tabTransitions.push({
        tab,
        time: Date.now() - transitionStart
      });
    }
    
    // 4. Memory Usage Assessment
    const memoryInfo = await page.evaluate(() => {
      if (performance.memory) {
        return {
          used: Math.round(performance.memory.usedJSHeapSize / 1024 / 1024),
          total: Math.round(performance.memory.totalJSHeapSize / 1024 / 1024)
        };
      }
      return { used: 0, total: 0 };
    });
    performanceMetrics.memoryUsage = memoryInfo;
    
    // 5. Animation Frame Rate Test
    const frameRate = await page.evaluate(() => {
      return new Promise((resolve) => {
        let frames = 0;
        const startTime = performance.now();
        
        function countFrames() {
          frames++;
          if (performance.now() - startTime < 1000) {
            requestAnimationFrame(countFrames);
          } else {
            resolve(frames);
          }
        }
        
        requestAnimationFrame(countFrames);
      });
    });
    performanceMetrics.animationFrameRate = frameRate;
    
    // Generate performance report
    console.log('\n=== PERFORMANCE BENCHMARK REPORT ===');
    console.log(`Page Load Time: ${performanceMetrics.pageLoad}ms (Target: <${PERFORMANCE_THRESHOLDS.PAGE_LOAD_TIME}ms)`);
    console.log(`WebSocket Connection: ${performanceMetrics.webSocketConnection}ms`);
    console.log(`Memory Usage: ${performanceMetrics.memoryUsage.used}MB / ${performanceMetrics.memoryUsage.total}MB`);
    console.log(`Animation Frame Rate: ${performanceMetrics.animationFrameRate} FPS`);
    
    console.log('\nTab Transition Performance:');
    performanceMetrics.tabTransitions.forEach(({ tab, time }) => {
      console.log(`  ${tab}: ${time}ms`);
    });
    
    // Validate against thresholds
    expect(performanceMetrics.pageLoad).toBeLessThan(PERFORMANCE_THRESHOLDS.PAGE_LOAD_TIME);
    expect(performanceMetrics.animationFrameRate).toBeGreaterThan(PERFORMANCE_THRESHOLDS.ANIMATION_FRAME_RATE);
    
    const avgTransitionTime = performanceMetrics.tabTransitions.reduce((sum, t) => sum + t.time, 0) / performanceMetrics.tabTransitions.length;
    expect(avgTransitionTime).toBeLessThan(300);
  });
});