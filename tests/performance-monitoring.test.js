/**
 * OllamaMax Performance Monitoring and Metrics Collection
 * Continuous performance monitoring and alerting system
 */

import { test, expect } from '@playwright/test';

const WEB_INTERFACE_URL = 'http://localhost:13100/web-interface/index.html';

// Performance monitoring utilities
class PerformanceMonitor {
  constructor(page) {
    this.page = page;
    this.metrics = {
      fps: [],
      memorySnapshots: [],
      networkMetrics: [],
      userTiming: [],
      paintTimings: []
    };
  }
  
  async startMonitoring() {
    // Inject performance monitoring into the page
    await this.page.evaluate(() => {
      window.performanceMonitor = {
        metrics: {
          fps: [],
          memorySnapshots: [],
          userTiming: [],
          paintTimings: []
        },
        
        // FPS monitoring
        startFPSMonitoring() {
          let frameCount = 0;
          let lastTime = performance.now();
          
          const measureFPS = (currentTime) => {
            frameCount++;
            
            if (currentTime - lastTime >= 1000) {
              this.metrics.fps.push(frameCount);
              frameCount = 0;
              lastTime = currentTime;
            }
            
            if (this.monitoring) {
              requestAnimationFrame(measureFPS);
            }
          };
          
          this.monitoring = true;
          requestAnimationFrame(measureFPS);
        },
        
        stopFPSMonitoring() {
          this.monitoring = false;
        },
        
        // Memory monitoring
        takeMemorySnapshot() {
          if (performance.memory) {
            this.metrics.memorySnapshots.push({
              timestamp: Date.now(),
              used: performance.memory.usedJSHeapSize,
              total: performance.memory.totalJSHeapSize
            });
          }
        },
        
        // User timing marks
        markStart(name) {
          performance.mark(`${name}-start`);
        },
        
        markEnd(name) {
          performance.mark(`${name}-end`);
          performance.measure(name, `${name}-start`, `${name}-end`);
          
          const measure = performance.getEntriesByName(name, 'measure')[0];
          this.metrics.userTiming.push({
            name,
            duration: measure.duration,
            timestamp: Date.now()
          });
        },
        
        // Paint timing
        collectPaintTimings() {
          const paintEntries = performance.getEntriesByType('paint');
          paintEntries.forEach(entry => {
            this.metrics.paintTimings.push({
              name: entry.name,
              startTime: entry.startTime
            });
          });
        }
      };
    });
  }
  
  async collectMetrics() {
    return await this.page.evaluate(() => window.performanceMonitor.metrics);
  }
  
  async generateReport(metrics) {
    const report = {
      fps: {
        average: metrics.fps.reduce((a, b) => a + b, 0) / metrics.fps.length || 0,
        min: Math.min(...metrics.fps) || 0,
        max: Math.max(...metrics.fps) || 0
      },
      memory: {
        growth: metrics.memorySnapshots.length > 1 ? 
          metrics.memorySnapshots[metrics.memorySnapshots.length - 1].used - metrics.memorySnapshots[0].used : 0,
        peak: Math.max(...metrics.memorySnapshots.map(s => s.used)) || 0
      },
      userTiming: metrics.userTiming,
      paintTimings: metrics.paintTimings
    };
    
    return report;
  }
}

test.describe('Performance Monitoring Tests', () => {
  
  test('should monitor real-time performance metrics', async ({ page }) => {
    const monitor = new PerformanceMonitor(page);
    
    await page.goto(WEB_INTERFACE_URL);
    await monitor.startMonitoring();
    
    // Start performance monitoring
    await page.evaluate(() => {
      window.performanceMonitor.startFPSMonitoring();
      window.performanceMonitor.takeMemorySnapshot();
      window.performanceMonitor.collectPaintTimings();
    });
    
    // Simulate user interaction session
    const userActions = [
      { action: () => page.click('[data-tab="nodes"]'), name: 'nodes-tab-switch' },
      { action: () => page.click('[data-tab="models"]'), name: 'models-tab-switch' },
      { action: () => page.fill('#messageInput', 'Test message'), name: 'message-input' },
      { action: () => page.click('#sendButton'), name: 'send-message' },
      { action: () => page.click('[data-tab="settings"]'), name: 'settings-tab-switch' }
    ];
    
    for (const { action, name } of userActions) {
      await page.evaluate((actionName) => {
        window.performanceMonitor.markStart(actionName);
      }, name);
      
      await action();
      
      await page.evaluate((actionName) => {
        window.performanceMonitor.markEnd(actionName);
        window.performanceMonitor.takeMemorySnapshot();
      }, name);
      
      await page.waitForTimeout(500);
    }
    
    // Stop monitoring and collect results
    await page.evaluate(() => {
      window.performanceMonitor.stopFPSMonitoring();
    });
    
    const metrics = await monitor.collectMetrics();
    const report = await monitor.generateReport(metrics);
    
    console.log('\n=== REAL-TIME PERFORMANCE REPORT ===');
    console.log(`FPS - Average: ${report.fps.average.toFixed(1)}, Min: ${report.fps.min}, Max: ${report.fps.max}`);
    console.log(`Memory Growth: ${(report.memory.growth / 1024 / 1024).toFixed(2)} MB`);
    console.log(`Peak Memory: ${(report.memory.peak / 1024 / 1024).toFixed(2)} MB`);
    
    console.log('\nUser Action Performance:');
    report.userTiming.forEach(timing => {
      console.log(`  ${timing.name}: ${timing.duration.toFixed(2)}ms`);
    });
    
    console.log('\nPaint Timings:');
    report.paintTimings.forEach(paint => {
      console.log(`  ${paint.name}: ${paint.startTime.toFixed(2)}ms`);
    });
    console.log('=====================================\n');
    
    // Validate performance metrics
    expect(report.fps.average).toBeGreaterThan(30); // Maintain good frame rate
    expect(report.memory.growth).toBeLessThan(50 * 1024 * 1024); // Less than 50MB growth
    
    // All user actions should be under 500ms
    report.userTiming.forEach(timing => {
      expect(timing.duration).toBeLessThan(500);
    });
  });
  
  test('should validate WebSocket message processing efficiency', async ({ page }) => {
    await page.goto(WEB_INTERFACE_URL);
    
    // Wait for WebSocket connection
    await page.waitForFunction(() => 
      window.llamaClient?.ws?.readyState === WebSocket.OPEN
    );
    
    // Monitor WebSocket message processing
    const messageMetrics = await page.evaluate(() => {
      const metrics = {
        messageProcessingTimes: [],
        totalMessages: 0,
        errors: 0
      };
      
      // Override message handler to measure performance
      const originalHandler = window.llamaClient.handleMessage;
      window.llamaClient.handleMessage = function(data) {
        const start = performance.now();
        
        try {
          originalHandler.call(this, data);
          metrics.totalMessages++;
        } catch (error) {
          metrics.errors++;
          console.error('Message handling error:', error);
        }
        
        const processingTime = performance.now() - start;
        metrics.messageProcessingTimes.push(processingTime);
      };
      
      // Simulate various message types
      const messageTypes = [
        { type: 'node_update', nodes: [] },
        { type: 'metrics', latency: 150, node: 'test-node' },
        { type: 'stream_chunk', id: 'test', chunk: 'Hello', done: false },
        { type: 'response', content: 'Test response', node: 'test-node' }
      ];
      
      messageTypes.forEach((msg, index) => {
        setTimeout(() => {
          window.llamaClient.handleMessage(msg);
        }, index * 50);
      });
      
      return new Promise(resolve => {
        setTimeout(() => resolve(metrics), 1000);
      });
    });
    
    const avgProcessingTime = messageMetrics.messageProcessingTimes.reduce((a, b) => a + b, 0) / 
                             messageMetrics.messageProcessingTimes.length;
    
    console.log(`WebSocket message processing - Average: ${avgProcessingTime.toFixed(2)}ms`);
    console.log(`Messages processed: ${messageMetrics.totalMessages}, Errors: ${messageMetrics.errors}`);
    
    expect(avgProcessingTime).toBeLessThan(10); // Fast message processing
    expect(messageMetrics.errors).toBe(0); // No processing errors
  });
  
  test('should monitor and validate animation performance', async ({ page }) => {
    await page.goto(WEB_INTERFACE_URL);
    
    // Test animation performance during various operations
    const animationMetrics = await page.evaluate(() => {
      const metrics = {
        transitions: [],
        animations: [],
        repaints: []
      };
      
      // Monitor CSS transitions
      document.addEventListener('transitionstart', (e) => {
        const start = performance.now();
        const element = e.target;
        
        const checkTransitionEnd = () => {
          if (element.style.transition && !element.style.transition.includes('running')) {
            metrics.transitions.push({
              element: element.tagName,
              property: e.propertyName,
              duration: performance.now() - start
            });
          } else {
            requestAnimationFrame(checkTransitionEnd);
          }
        };
        
        requestAnimationFrame(checkTransitionEnd);
      });
      
      // Trigger various animations
      document.querySelectorAll('.tab-button').forEach((button, index) => {
        setTimeout(() => {
          button.dispatchEvent(new MouseEvent('mouseover'));
          setTimeout(() => {
            button.dispatchEvent(new MouseEvent('mouseout'));
          }, 200);
        }, index * 100);
      });
      
      return new Promise(resolve => {
        setTimeout(() => resolve(metrics), 2000);
      });
    });
    
    console.log('Animation Performance Metrics:');
    console.log(`Transitions tracked: ${animationMetrics.transitions.length}`);
    
    if (animationMetrics.transitions.length > 0) {
      const avgTransitionTime = animationMetrics.transitions.reduce((sum, t) => sum + t.duration, 0) / 
                               animationMetrics.transitions.length;
      console.log(`Average transition time: ${avgTransitionTime.toFixed(2)}ms`);
      
      expect(avgTransitionTime).toBeLessThan(300); // Smooth transitions
    }
  });
});