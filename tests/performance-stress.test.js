/**
 * OllamaMax Stress Testing Suite
 * High-load scenarios and edge case performance validation
 */

import { test, expect, chromium } from '@playwright/test';

const STRESS_THRESHOLDS = {
  HIGH_LOAD_RESPONSE: 10000,    // 10 seconds under stress
  MEMORY_LEAK_THRESHOLD: 100,   // 100MB max growth
  CONCURRENT_CONNECTIONS: 10,   // Support 10+ concurrent users
  MESSAGE_BURST_RATE: 50,       // Handle 50 messages in quick succession
  NODE_COUNT_LIMIT: 50          // Support monitoring 50+ nodes
};

const WEB_INTERFACE_URL = 'http://localhost:13100/web-interface/index.html';

test.describe('Stress Testing Suite', () => {
  
  test.describe('High Load Scenarios', () => {
    
    test('should handle message burst without degradation', async ({ page }) => {
      await page.goto(WEB_INTERFACE_URL);
      
      // Wait for WebSocket connection
      await page.waitForFunction(() => 
        window.llamaClient?.ws?.readyState === WebSocket.OPEN
      );
      
      const burstStart = Date.now();
      
      // Send burst of messages
      for (let i = 0; i < STRESS_THRESHOLDS.MESSAGE_BURST_RATE; i++) {
        await page.evaluate((index) => {
          window.llamaClient.addMessage('user', `Burst message ${index}`, 'test-node');
        }, i);
        
        // Small delay to simulate rapid but realistic typing
        if (i % 10 === 0) await page.waitForTimeout(10);
      }
      
      const burstTime = Date.now() - burstStart;
      console.log(`Message burst handling time: ${burstTime}ms`);
      
      // UI should remain responsive
      expect(burstTime).toBeLessThan(2000);
      
      // Verify all messages are displayed
      const messageCount = await page.locator('.message.user').count();
      expect(messageCount).toBe(STRESS_THRESHOLDS.MESSAGE_BURST_RATE);
      
      // Scrolling should still be smooth
      const scrollStart = Date.now();
      await page.evaluate(() => {
        const messagesArea = document.getElementById('messagesArea');
        messagesArea.scrollTop = messagesArea.scrollHeight;
      });
      const scrollTime = Date.now() - scrollStart;
      
      expect(scrollTime).toBeLessThan(100);
    });
    
    test('should handle large node dataset efficiently', async ({ page }) => {
      await page.goto(WEB_INTERFACE_URL);
      
      // Navigate to nodes tab
      await page.click('[data-tab="nodes"]');
      
      const largeDatasetStart = Date.now();
      
      // Create large dataset of nodes
      await page.evaluate(() => {
        const largeNodeSet = Array.from({ length: 100 }, (_, i) => ({
          id: `stress-node-${i}`,
          name: `stress-llama-${i.toString().padStart(3, '0')}`,
          status: ['healthy', 'warning', 'error'][i % 3],
          systemInfo: {
            cpu: { 
              usage: Math.random() * 100, 
              cores: [4, 8, 16, 32][i % 4],
              model: `CPU Model ${i % 5}`
            },
            memory: { 
              usage: Math.random() * 100,
              total: [16, 32, 64, 128][i % 4] * 1024 * 1024 * 1024
            },
            disk: { usage: Math.random() * 100 },
            network: {
              rx: Math.random() * 1000 * 1024 * 1024,
              tx: Math.random() * 500 * 1024 * 1024
            }
          },
          ollamaInfo: {
            models: Array.from({ length: Math.floor(Math.random() * 5) + 1 }, (_, j) => ({
              name: `model-${j}-on-node-${i}`,
              size: Math.random() * 5000 * 1024 * 1024
            })),
            activeRequests: Math.floor(Math.random() * 10),
            queueLength: Math.floor(Math.random() * 20)
          },
          performanceHistory: {
            timestamps: Array.from({ length: 50 }, (_, j) => Date.now() - j * 60000),
            cpu: Array.from({ length: 50 }, () => Math.random() * 100),
            memory: Array.from({ length: 50 }, () => Math.random() * 100),
            responseTime: Array.from({ length: 50 }, () => 50 + Math.random() * 500)
          }
        }));
        
        window.llamaClient.detailedNodes = largeNodeSet;
        window.llamaClient.displayEnhancedNodes();
      });
      
      const renderTime = Date.now() - largeDatasetStart;
      console.log(`Large dataset (100 nodes) render time: ${renderTime}ms`);
      
      expect(renderTime).toBeLessThan(2000);
      
      // Test filtering performance
      const filterStart = Date.now();
      await page.fill('#nodeSearch', 'stress-llama-01');
      await page.waitForTimeout(300); // Debounce delay
      
      const filterTime = Date.now() - filterStart;
      console.log(`Large dataset filter time: ${filterTime}ms`);
      
      expect(filterTime).toBeLessThan(500);
    });
    
    test('should maintain performance during continuous updates', async ({ page }) => {
      await page.goto(WEB_INTERFACE_URL);
      
      // Start continuous update simulation
      const updateTimes = [];
      let updateCount = 0;
      
      await page.evaluate(() => {
        window.continuousUpdateTest = true;
        
        const updateInterval = setInterval(() => {
          if (!window.continuousUpdateTest) {
            clearInterval(updateInterval);
            return;
          }
          
          const start = performance.now();
          
          // Simulate node status updates
          if (window.llamaClient.nodes.length > 0) {
            window.llamaClient.nodes.forEach(node => {
              node.load = Math.random() * 100;
              node.memory = Math.random() * 100;
              node.requestsPerSecond = Math.floor(Math.random() * 20);
            });
            window.llamaClient.updateNodeDisplay();
          }
          
          const updateTime = performance.now() - start;
          window.updatePerformanceTimes = window.updatePerformanceTimes || [];
          window.updatePerformanceTimes.push(updateTime);
          
        }, 100); // Update every 100ms
      });
      
      // Let it run for 10 seconds
      await page.waitForTimeout(10000);
      
      // Stop continuous updates
      await page.evaluate(() => {
        window.continuousUpdateTest = false;
      });
      
      const updatePerformance = await page.evaluate(() => window.updatePerformanceTimes || []);
      const avgUpdateTime = updatePerformance.reduce((a, b) => a + b, 0) / updatePerformance.length;
      const maxUpdateTime = Math.max(...updatePerformance);
      
      console.log(`Continuous updates - Average: ${avgUpdateTime.toFixed(2)}ms, Max: ${maxUpdateTime.toFixed(2)}ms`);
      
      expect(avgUpdateTime).toBeLessThan(50); // Should update efficiently
      expect(maxUpdateTime).toBeLessThan(200); // No single update should take too long
    });
  });
  
  test.describe('Edge Case Performance', () => {
    
    test('should handle WebSocket reconnection efficiently', async ({ page }) => {
      await page.goto(WEB_INTERFACE_URL);
      
      // Wait for initial connection
      await page.waitForFunction(() => 
        window.llamaClient?.ws?.readyState === WebSocket.OPEN
      );
      
      // Simulate connection loss
      const reconnectStart = Date.now();
      
      await page.evaluate(() => {
        // Force close WebSocket to simulate connection loss
        window.llamaClient.ws.close();
      });
      
      // Wait for reconnection attempt
      await page.waitForFunction(() => 
        document.getElementById('connectionText').textContent.includes('Reconnecting') ||
        document.getElementById('connectionText').textContent.includes('Connected')
      );
      
      const reconnectTime = Date.now() - reconnectStart;
      console.log(`WebSocket reconnection detection time: ${reconnectTime}ms`);
      
      expect(reconnectTime).toBeLessThan(3000);
    });
    
    test('should handle rapid tab switching without performance degradation', async ({ page }) => {
      await page.goto(WEB_INTERFACE_URL);
      
      const tabs = ['chat', 'nodes', 'models', 'settings'];
      const switchTimes = [];
      
      // Rapid tab switching test
      for (let i = 0; i < 50; i++) {
        const tab = tabs[i % tabs.length];
        const switchStart = Date.now();
        
        await page.click(`[data-tab="${tab}"]`);
        await page.waitForSelector(`#${tab}Tab.active`);
        
        const switchTime = Date.now() - switchStart;
        switchTimes.push(switchTime);
        
        // No delay between switches to stress test
      }
      
      const avgSwitchTime = switchTimes.reduce((a, b) => a + b, 0) / switchTimes.length;
      const maxSwitchTime = Math.max(...switchTimes);
      
      console.log(`Rapid tab switching - Average: ${avgSwitchTime}ms, Max: ${maxSwitchTime}ms`);
      
      expect(avgSwitchTime).toBeLessThan(200);
      expect(maxSwitchTime).toBeLessThan(500);
    });
    
    test('should handle browser resize performance', async ({ page }) => {
      await page.goto(WEB_INTERFACE_URL);
      
      const resizeSizes = [
        { width: 1920, height: 1080 },
        { width: 1280, height: 720 },
        { width: 768, height: 1024 },
        { width: 375, height: 667 },
        { width: 1440, height: 900 }
      ];
      
      const resizeTimes = [];
      
      for (const size of resizeSizes) {
        const resizeStart = Date.now();
        
        await page.setViewportSize(size);
        
        // Wait for layout to stabilize
        await page.waitForTimeout(200);
        
        const resizeTime = Date.now() - resizeStart;
        resizeTimes.push(resizeTime);
        
        console.log(`Resize to ${size.width}x${size.height}: ${resizeTime}ms`);
      }
      
      const avgResizeTime = resizeTimes.reduce((a, b) => a + b, 0) / resizeTimes.length;
      expect(avgResizeTime).toBeLessThan(300);
    });
  });
  
  test.describe('Memory Leak Detection', () => {
    
    test('should not leak memory during extended usage', async ({ page }) => {
      await page.goto(WEB_INTERFACE_URL);
      
      // Get baseline memory
      const initialMemory = await page.evaluate(() => {
        return performance.memory ? performance.memory.usedJSHeapSize : 0;
      });
      
      // Simulate 10 minutes of usage
      for (let cycle = 0; cycle < 20; cycle++) {
        // Add and remove messages
        await page.evaluate(() => {
          for (let i = 0; i < 10; i++) {
            window.llamaClient.addMessage('user', `Cycle ${arguments[0]} message ${i}`, 'test-node');
          }
        }, cycle);
        
        // Switch tabs
        await page.click('[data-tab="nodes"]');
        await page.waitForTimeout(100);
        
        await page.click('[data-tab="models"]');
        await page.waitForTimeout(100);
        
        await page.click('[data-tab="chat"]');
        await page.waitForTimeout(100);
        
        // Force garbage collection every 5 cycles
        if (cycle % 5 === 0) {
          await page.evaluate(() => {
            if (window.gc) window.gc();
          });
          await page.waitForTimeout(100);
        }
      }
      
      // Final memory check
      const finalMemory = await page.evaluate(() => {
        return performance.memory ? performance.memory.usedJSHeapSize : 0;
      });
      
      const memoryGrowth = (finalMemory - initialMemory) / 1024 / 1024;
      console.log(`Memory growth during extended usage: ${memoryGrowth.toFixed(2)} MB`);
      
      expect(memoryGrowth).toBeLessThan(STRESS_THRESHOLDS.MEMORY_LEAK_THRESHOLD);
    });
    
    test('should clean up event listeners and timers', async ({ page }) => {
      await page.goto(WEB_INTERFACE_URL);
      
      // Count initial event listeners
      const initialEventListeners = await page.evaluate(() => {
        const getEventListeners = (element) => {
          return getEventListeners ? getEventListeners(element) : {};
        };
        
        let count = 0;
        document.querySelectorAll('*').forEach(el => {
          const listeners = getEventListeners(el);
          Object.keys(listeners).forEach(type => {
            count += listeners[type].length;
          });
        });
        return count;
      });
      
      // Simulate component creation and destruction cycles
      for (let i = 0; i < 10; i++) {
        await page.click('[data-tab="nodes"]');
        await page.waitForTimeout(200);
        
        // Simulate dynamic content loading
        await page.evaluate(() => {
          if (window.llamaClient.displayEnhancedNodes) {
            window.llamaClient.displayEnhancedNodes();
          }
        });
        
        await page.click('[data-tab="chat"]');
        await page.waitForTimeout(200);
      }
      
      // Check for timer leaks
      const activeTimers = await page.evaluate(() => {
        // Return count of active intervals/timeouts (if accessible)
        return window.activeTimers || 0;
      });
      
      console.log(`Active timers detected: ${activeTimers}`);
      
      // Should not accumulate excessive timers
      expect(activeTimers).toBeLessThan(20);
    });
  });
  
  test.describe('Concurrent User Simulation', () => {
    
    test('should handle multiple concurrent users', async () => {
      const browser = await chromium.launch();
      const contexts = [];
      const pages = [];
      const performanceMetrics = [];
      
      try {
        // Create multiple browser contexts to simulate concurrent users
        for (let i = 0; i < STRESS_THRESHOLDS.CONCURRENT_CONNECTIONS; i++) {
          const context = await browser.newContext();
          const page = await context.newPage();
          contexts.push(context);
          pages.push(page);
        }
        
        console.log(`Testing with ${STRESS_THRESHOLDS.CONCURRENT_CONNECTIONS} concurrent users`);
        
        const concurrentStart = Date.now();
        
        // Load interface simultaneously
        const loadPromises = pages.map(async (page, index) => {
          const pageStart = Date.now();
          await page.goto(WEB_INTERFACE_URL);
          await page.waitForSelector('#app');
          
          const pageLoadTime = Date.now() - pageStart;
          return { user: index, loadTime: pageLoadTime };
        });
        
        const loadResults = await Promise.all(loadPromises);
        const totalConcurrentTime = Date.now() - concurrentStart;
        
        console.log(`Concurrent load completed in: ${totalConcurrentTime}ms`);
        
        // Analyze individual load times
        const loadTimes = loadResults.map(r => r.loadTime);
        const avgLoadTime = loadTimes.reduce((a, b) => a + b, 0) / loadTimes.length;
        const maxLoadTime = Math.max(...loadTimes);
        
        console.log(`Load times - Average: ${avgLoadTime}ms, Max: ${maxLoadTime}ms`);
        
        expect(maxLoadTime).toBeLessThan(8000); // Even under load, should load within 8s
        expect(avgLoadTime).toBeLessThan(5000);
        
        // Test concurrent interaction
        const interactionPromises = pages.map(async (page, index) => {
          // Each user performs different actions
          const actions = [
            () => page.click('[data-tab="nodes"]'),
            () => page.click('[data-tab="models"]'),
            () => page.click('[data-tab="settings"]'),
            () => page.fill('#messageInput', `User ${index} message`)
          ];
          
          const actionStart = Date.now();
          await actions[index % actions.length]();
          return Date.now() - actionStart;
        });
        
        const interactionTimes = await Promise.all(interactionPromises);
        const avgInteractionTime = interactionTimes.reduce((a, b) => a + b, 0) / interactionTimes.length;
        
        console.log(`Concurrent interaction average time: ${avgInteractionTime}ms`);
        expect(avgInteractionTime).toBeLessThan(1000);
        
      } finally {
        // Cleanup
        for (const context of contexts) {
          await context.close();
        }
        await browser.close();
      }
    });
  });
  
  test.describe('Resource Exhaustion Tests', () => {
    
    test('should handle DOM manipulation stress', async ({ page }) => {
      await page.goto(WEB_INTERFACE_URL);
      
      const domStressStart = Date.now();
      
      // Stress test DOM operations
      await page.evaluate(() => {
        const messagesArea = document.getElementById('messagesArea');
        
        // Add large number of elements rapidly
        for (let i = 0; i < 1000; i++) {
          const messageEl = document.createElement('div');
          messageEl.className = 'message stress-test';
          messageEl.innerHTML = `
            <div class="message-header">
              <span>User</span>
              <span>node-${i % 3}</span>
              <span>${new Date().toLocaleTimeString()}</span>
            </div>
            <div class="message-content">Stress test message ${i} with some content</div>
          `;
          messagesArea.appendChild(messageEl);
          
          // Occasionally trigger layout
          if (i % 100 === 0) {
            messagesArea.scrollTop = messagesArea.scrollHeight;
          }
        }
      });
      
      const domStressTime = Date.now() - domStressStart;
      console.log(`DOM stress test (1000 elements) time: ${domStressTime}ms`);
      
      expect(domStressTime).toBeLessThan(3000);
      
      // Verify UI remains responsive
      await page.click('[data-tab="nodes"]');
      await page.waitForSelector('#nodesTab.active');
      
      // Clean up stress test elements
      await page.evaluate(() => {
        document.querySelectorAll('.message.stress-test').forEach(el => el.remove());
      });
    });
    
    test('should handle Canvas performance under load', async ({ page }) => {
      await page.goto(WEB_INTERFACE_URL);
      
      // Navigate to nodes tab to trigger canvas creation
      await page.click('[data-tab="nodes"]');
      await page.waitForTimeout(500);
      
      // Test canvas drawing performance
      const canvasPerformance = await page.evaluate(() => {
        const canvases = document.querySelectorAll('canvas');
        const renderTimes = [];
        
        canvases.forEach((canvas, index) => {
          const ctx = canvas.getContext('2d');
          const start = performance.now();
          
          // Simulate complex drawing operations
          for (let i = 0; i < 100; i++) {
            ctx.beginPath();
            ctx.moveTo(Math.random() * canvas.width, Math.random() * canvas.height);
            ctx.lineTo(Math.random() * canvas.width, Math.random() * canvas.height);
            ctx.strokeStyle = `hsl(${Math.random() * 360}, 50%, 50%)`;
            ctx.stroke();
          }
          
          const renderTime = performance.now() - start;
          renderTimes.push(renderTime);
        });
        
        return renderTimes;
      });
      
      if (canvasPerformance.length > 0) {
        const avgCanvasRenderTime = canvasPerformance.reduce((a, b) => a + b, 0) / canvasPerformance.length;
        console.log(`Canvas rendering performance - Average: ${avgCanvasRenderTime}ms`);
        
        expect(avgCanvasRenderTime).toBeLessThan(100);
      }
    });
  });
  
  test.describe('Network Performance Under Load', () => {
    
    test('should handle API request bursts efficiently', async ({ page }) => {
      const requests = [];
      
      // Monitor all API requests
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
      
      // Create burst of API-triggering actions
      const burstStart = Date.now();
      
      for (let i = 0; i < 20; i++) {
        await page.click('[data-tab="nodes"]');
        await page.waitForTimeout(50);
        await page.click('[data-tab="models"]');
        await page.waitForTimeout(50);
        await page.click('#refreshNodes');
        await page.waitForTimeout(50);
      }
      
      const burstTime = Date.now() - burstStart;
      console.log(`API burst test completed in: ${burstTime}ms`);
      console.log(`Total requests generated: ${requests.length}`);
      
      // Should handle burst without timeout
      expect(burstTime).toBeLessThan(10000);
      
      // Should implement request deduplication
      const uniqueRequests = new Set(requests.map(r => r.url));
      const deduplicationRatio = uniqueRequests.size / requests.length;
      console.log(`Request deduplication ratio: ${(deduplicationRatio * 100).toFixed(1)}%`);
      
      // Should have some level of deduplication
      expect(deduplicationRatio).toBeLessThan(0.8);
    });
    
    test('should handle offline/online transitions', async ({ page }) => {
      await page.goto(WEB_INTERFACE_URL);
      
      // Start in online mode
      await page.waitForFunction(() => 
        window.llamaClient?.ws?.readyState === WebSocket.OPEN
      );
      
      // Simulate going offline
      const offlineStart = Date.now();
      
      await page.context().setOffline(true);
      
      // Should detect offline status quickly
      await page.waitForFunction(() => 
        document.getElementById('connectionText').textContent.includes('error') ||
        document.getElementById('connectionText').textContent.includes('Disconnected')
      );
      
      const offlineDetectionTime = Date.now() - offlineStart;
      console.log(`Offline detection time: ${offlineDetectionTime}ms`);
      
      expect(offlineDetectionTime).toBeLessThan(5000);
      
      // Simulate going back online
      const onlineStart = Date.now();
      
      await page.context().setOffline(false);
      
      // Should reconnect quickly
      await page.waitForFunction(() => 
        document.getElementById('connectionText').textContent.includes('Connected') ||
        document.getElementById('connectionText').textContent.includes('Connecting')
      );
      
      const onlineDetectionTime = Date.now() - onlineStart;
      console.log(`Online detection time: ${onlineDetectionTime}ms`);
      
      expect(onlineDetectionTime).toBeLessThan(3000);
    });
  });
});

test.describe('Performance Regression Prevention', () => {
  
  test('should establish performance baseline benchmarks', async ({ page }) => {
    const baseline = {
      pageLoad: null,
      memoryFootprint: null,
      domComplexity: null,
      jsExecutionTime: null
    };
    
    // 1. Page Load Baseline
    const loadStart = Date.now();
    await page.goto(WEB_INTERFACE_URL, { waitUntil: 'networkidle' });
    baseline.pageLoad = Date.now() - loadStart;
    
    // 2. Memory Footprint Baseline
    baseline.memoryFootprint = await page.evaluate(() => {
      return performance.memory ? 
        Math.round(performance.memory.usedJSHeapSize / 1024 / 1024) : 0;
    });
    
    // 3. DOM Complexity Baseline
    baseline.domComplexity = await page.evaluate(() => {
      return {
        totalNodes: document.querySelectorAll('*').length,
        eventListeners: document.querySelectorAll('[onclick], [onmouseover]').length,
        canvasElements: document.querySelectorAll('canvas').length
      };
    });
    
    // 4. JavaScript Execution Baseline
    const jsStart = Date.now();
    await page.evaluate(() => {
      // Simulate typical JS operations
      window.llamaClient.updateNodeDisplay();
      if (window.llamaClient.loadModels) {
        // Mock load models operation
        window.llamaClient.updateModelSelector(['model1', 'model2', 'model3']);
      }
    });
    baseline.jsExecutionTime = Date.now() - jsStart;
    
    // Log baseline metrics
    console.log('\n=== PERFORMANCE BASELINE ESTABLISHED ===');
    console.log(`Page Load: ${baseline.pageLoad}ms`);
    console.log(`Memory Footprint: ${baseline.memoryFootprint}MB`);
    console.log(`DOM Complexity: ${baseline.domComplexity.totalNodes} nodes, ${baseline.domComplexity.eventListeners} listeners`);
    console.log(`JS Execution: ${baseline.jsExecutionTime}ms`);
    console.log('=====================================\n');
    
    // Store baseline for comparison in CI/CD
    await page.evaluate((baseline) => {
      localStorage.setItem('performanceBaseline', JSON.stringify(baseline));
    }, baseline);
    
    // Validate baseline meets performance requirements
    expect(baseline.pageLoad).toBeLessThan(PERFORMANCE_THRESHOLDS.PAGE_LOAD_TIME);
    expect(baseline.memoryFootprint).toBeLessThan(100); // Initial footprint under 100MB
    expect(baseline.jsExecutionTime).toBeLessThan(100);
  });
});