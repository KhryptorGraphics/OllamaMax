import { Page } from '@playwright/test';
import fs from 'fs/promises';
import path from 'path';

/**
 * Metrics Collection Helper for OllamaMax Platform
 * 
 * Provides utilities for:
 * - System resource monitoring
 * - Application metrics collection
 * - Performance benchmarking
 * - Resource utilization tracking
 * - Custom metrics aggregation
 */

export interface SystemMetrics {
  timestamp: number;
  memory: {
    used: number;
    total: number;
    percentage: number;
  };
  cpu?: {
    usage: number;
    processes: number;
  };
  network: {
    requestCount: number;
    totalBytes: number;
    activeConnections: number;
  };
  dom: {
    elementCount: number;
    scriptCount: number;
    styleSheetCount: number;
  };
  performance: {
    loadTime: number;
    renderTime: number;
    interactionReady: number;
  };
}

export interface ApplicationMetrics {
  modelMetrics?: {
    loadedModels: number;
    activeInferences: number;
    totalInferences: number;
    averageResponseTime: number;
  };
  distributedMetrics?: {
    activeNodes: number;
    totalNodes: number;
    loadDistribution: { [nodeId: string]: number };
    failoverEvents: number;
  };
  apiMetrics?: {
    totalRequests: number;
    successRate: number;
    errorRate: number;
    averageLatency: number;
  };
}

export class MetricsHelper {
  constructor(private page: Page) {}

  /**
   * Collect comprehensive system metrics
   */
  async collectSystemMetrics(): Promise<SystemMetrics> {
    const metrics = await this.page.evaluate(() => {
      const now = Date.now();
      
      // Memory information
      const memory = (performance as any).memory || {};
      const memoryUsed = memory.usedJSHeapSize || 0;
      const memoryTotal = memory.totalJSHeapSize || 0;
      const memoryPercentage = memoryTotal > 0 ? (memoryUsed / memoryTotal) * 100 : 0;
      
      // Network information
      const resources = performance.getEntriesByType('resource') as PerformanceResourceTiming[];
      const totalBytes = resources.reduce((sum, resource) => {
        return sum + ((resource as any).transferSize || 0);
      }, 0);
      
      // DOM information
      const elements = document.querySelectorAll('*').length;
      const scripts = document.scripts.length;
      const stylesheets = document.styleSheets.length;
      
      // Performance timing
      const navigation = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
      const loadTime = navigation ? navigation.loadEventEnd - navigation.fetchStart : 0;
      const renderTime = navigation ? navigation.domContentLoadedEventEnd - navigation.fetchStart : 0;
      const interactionTime = navigation ? navigation.domInteractive - navigation.fetchStart : 0;
      
      return {
        timestamp: now,
        memory: {
          used: memoryUsed,
          total: memoryTotal,
          percentage: Math.round(memoryPercentage * 100) / 100
        },
        network: {
          requestCount: resources.length,
          totalBytes: totalBytes,
          activeConnections: 0 // Placeholder - would need WebRTC or other APIs for actual count
        },
        dom: {
          elementCount: elements,
          scriptCount: scripts,
          styleSheetCount: stylesheets
        },
        performance: {
          loadTime: Math.round(loadTime),
          renderTime: Math.round(renderTime),
          interactionReady: Math.round(interactionTime)
        }
      };
    });
    
    return metrics as SystemMetrics;
  }

  /**
   * Monitor system metrics over time
   */
  async monitorMetrics(
    duration: number = 30000, 
    interval: number = 2000
  ): Promise<SystemMetrics[]> {
    const samples: SystemMetrics[] = [];
    const endTime = Date.now() + duration;
    
    console.log(`📊 Starting metrics monitoring for ${duration/1000} seconds...`);
    
    while (Date.now() < endTime) {
      const metrics = await this.collectSystemMetrics();
      samples.push(metrics);
      
      await this.page.waitForTimeout(interval);
    }
    
    console.log(`📊 Metrics monitoring completed. Collected ${samples.length} samples.`);
    return samples;
  }

  /**
   * Collect application-specific metrics
   */
  async collectApplicationMetrics(): Promise<ApplicationMetrics> {
    const metrics: ApplicationMetrics = {};
    
    // Try to collect model metrics (if available in the UI)
    const modelMetrics = await this.collectModelMetrics();
    if (modelMetrics) {
      metrics.modelMetrics = modelMetrics;
    }
    
    // Try to collect distributed system metrics
    const distributedMetrics = await this.collectDistributedMetrics();
    if (distributedMetrics) {
      metrics.distributedMetrics = distributedMetrics;
    }
    
    // Try to collect API metrics
    const apiMetrics = await this.collectAPIMetrics();
    if (apiMetrics) {
      metrics.apiMetrics = apiMetrics;
    }
    
    return metrics;
  }

  /**
   * Collect model-specific metrics
   */
  private async collectModelMetrics(): Promise<ApplicationMetrics['modelMetrics'] | null> {
    try {
      return await this.page.evaluate(() => {
        // Look for model information in the DOM
        const modelElements = document.querySelectorAll('.model, [data-testid="model"], .loaded-model');
        const loadedModels = modelElements.length;
        
        // Look for active inference indicators
        const activeInferences = document.querySelectorAll('.inference-active, .processing, .generating').length;
        
        // Try to extract metrics from data attributes or text content
        let totalInferences = 0;
        let responseTimeSum = 0;
        let responseTimeCount = 0;
        
        modelElements.forEach(element => {
          const inferenceCount = element.getAttribute('data-inference-count');
          const avgResponseTime = element.getAttribute('data-avg-response-time');
          
          if (inferenceCount) {
            totalInferences += parseInt(inferenceCount, 10) || 0;
          }
          
          if (avgResponseTime) {
            const time = parseFloat(avgResponseTime) || 0;
            responseTimeSum += time;
            responseTimeCount++;
          }
        });
        
        const averageResponseTime = responseTimeCount > 0 ? responseTimeSum / responseTimeCount : 0;
        
        return loadedModels > 0 ? {
          loadedModels,
          activeInferences,
          totalInferences,
          averageResponseTime: Math.round(averageResponseTime)
        } : null;
      });
    } catch (error) {
      return null;
    }
  }

  /**
   * Collect distributed system metrics
   */
  private async collectDistributedMetrics(): Promise<ApplicationMetrics['distributedMetrics'] | null> {
    try {
      return await this.page.evaluate(() => {
        // Look for node information
        const nodeElements = document.querySelectorAll('.node, [data-testid="node"], .cluster-node, .worker-node');
        const totalNodes = nodeElements.length;
        
        if (totalNodes === 0) return null;
        
        let activeNodes = 0;
        const loadDistribution: { [nodeId: string]: number } = {};
        let failoverEvents = 0;
        
        nodeElements.forEach((element, index) => {
          const text = element.textContent?.toLowerCase() || '';
          const nodeId = element.getAttribute('data-node-id') || `node-${index}`;
          
          // Check if node is active
          if (text.includes('active') || text.includes('online') || text.includes('healthy')) {
            activeNodes++;
          }
          
          // Extract load information if available
          const loadMatch = text.match(/load[:\s]*(\d+)/);
          if (loadMatch) {
            loadDistribution[nodeId] = parseInt(loadMatch[1], 10);
          }
          
          // Count failover events
          if (text.includes('failover') || text.includes('switched')) {
            failoverEvents++;
          }
        });
        
        return {
          activeNodes,
          totalNodes,
          loadDistribution,
          failoverEvents
        };
      });
    } catch (error) {
      return null;
    }
  }

  /**
   * Collect API metrics
   */
  private async collectAPIMetrics(): Promise<ApplicationMetrics['apiMetrics'] | null> {
    try {
      // Get network timing information
      const resources = await this.page.evaluate(() => {
        const resources = performance.getEntriesByType('resource') as PerformanceResourceTiming[];
        return resources
          .filter(resource => resource.name.includes('/api/'))
          .map(resource => ({
            url: resource.name,
            duration: resource.responseEnd - resource.startTime,
            status: (resource as any).responseStatus || 200 // Not always available
          }));
      });
      
      if (resources.length === 0) return null;
      
      const totalRequests = resources.length;
      const successfulRequests = resources.filter(r => !r.status || r.status < 400).length;
      const successRate = (successfulRequests / totalRequests) * 100;
      const errorRate = 100 - successRate;
      const averageLatency = resources.reduce((sum, r) => sum + r.duration, 0) / totalRequests;
      
      return {
        totalRequests,
        successRate: Math.round(successRate * 100) / 100,
        errorRate: Math.round(errorRate * 100) / 100,
        averageLatency: Math.round(averageLatency)
      };
    } catch (error) {
      return null;
    }
  }

  /**
   * Benchmark specific operations
   */
  async benchmarkOperation<T>(
    operation: () => Promise<T>,
    name: string,
    iterations: number = 10
  ): Promise<{
    name: string;
    iterations: number;
    totalTime: number;
    averageTime: number;
    minTime: number;
    maxTime: number;
    results: T[];
    memoryDelta: number;
  }> {
    console.log(`🏃 Benchmarking "${name}" with ${iterations} iterations...`);
    
    const results: T[] = [];
    const times: number[] = [];
    
    // Get initial memory
    const initialMemory = await this.page.evaluate(() => {
      const memory = (performance as any).memory;
      return memory ? memory.usedJSHeapSize : 0;
    });
    
    // Run iterations
    for (let i = 0; i < iterations; i++) {
      const startTime = Date.now();
      const result = await operation();
      const endTime = Date.now();
      
      results.push(result);
      times.push(endTime - startTime);
      
      // Small delay between iterations
      await this.page.waitForTimeout(100);
    }
    
    // Get final memory
    const finalMemory = await this.page.evaluate(() => {
      const memory = (performance as any).memory;
      return memory ? memory.usedJSHeapSize : 0;
    });
    
    const totalTime = times.reduce((sum, time) => sum + time, 0);
    const averageTime = totalTime / iterations;
    const minTime = Math.min(...times);
    const maxTime = Math.max(...times);
    const memoryDelta = finalMemory - initialMemory;
    
    const benchmark = {
      name,
      iterations,
      totalTime,
      averageTime: Math.round(averageTime * 100) / 100,
      minTime,
      maxTime,
      results,
      memoryDelta
    };
    
    console.log(`✅ Benchmark "${name}" completed:`, {
      avgTime: `${benchmark.averageTime}ms`,
      minTime: `${minTime}ms`,
      maxTime: `${maxTime}ms`,
      memoryDelta: `${Math.round(memoryDelta / 1024)}KB`
    });
    
    return benchmark;
  }

  /**
   * Save metrics to file
   */
  async saveMetrics(
    testName: string, 
    metrics: SystemMetrics | SystemMetrics[] | ApplicationMetrics,
    additionalData?: any
  ): Promise<void> {
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
    const filename = `metrics-${testName}-${timestamp}.json`;
    const filepath = path.join('reports', 'metrics', filename);
    
    // Ensure directory exists
    await fs.mkdir(path.dirname(filepath), { recursive: true }).catch(() => {});
    
    const data = {
      testName,
      timestamp: new Date().toISOString(),
      url: this.page.url(),
      userAgent: await this.page.evaluate(() => navigator.userAgent),
      viewport: await this.page.viewportSize(),
      metrics,
      ...additionalData
    };
    
    await fs.writeFile(filepath, JSON.stringify(data, null, 2));
    console.log(`📊 Metrics saved to: ${filepath}`);
  }

  /**
   * Generate metrics summary report
   */
  static generateMetricsSummary(samples: SystemMetrics[]): {
    summary: {
      duration: number;
      sampleCount: number;
      averages: Partial<SystemMetrics>;
      peaks: Partial<SystemMetrics>;
      trends: {
        memoryTrend: 'increasing' | 'decreasing' | 'stable';
        performanceTrend: 'improving' | 'degrading' | 'stable';
      };
    };
  } {
    if (samples.length === 0) {
      throw new Error('No samples provided for metrics summary');
    }
    
    const duration = samples[samples.length - 1].timestamp - samples[0].timestamp;
    
    // Calculate averages
    const averages = {
      memory: {
        used: Math.round(samples.reduce((sum, s) => sum + s.memory.used, 0) / samples.length),
        total: Math.round(samples.reduce((sum, s) => sum + s.memory.total, 0) / samples.length),
        percentage: Math.round((samples.reduce((sum, s) => sum + s.memory.percentage, 0) / samples.length) * 100) / 100
      },
      network: {
        requestCount: Math.round(samples.reduce((sum, s) => sum + s.network.requestCount, 0) / samples.length),
        totalBytes: Math.round(samples.reduce((sum, s) => sum + s.network.totalBytes, 0) / samples.length),
        activeConnections: Math.round(samples.reduce((sum, s) => sum + s.network.activeConnections, 0) / samples.length)
      },
      performance: {
        loadTime: Math.round(samples.reduce((sum, s) => sum + s.performance.loadTime, 0) / samples.length),
        renderTime: Math.round(samples.reduce((sum, s) => sum + s.performance.renderTime, 0) / samples.length),
        interactionReady: Math.round(samples.reduce((sum, s) => sum + s.performance.interactionReady, 0) / samples.length)
      }
    };
    
    // Find peaks
    const peaks = {
      memory: {
        used: Math.max(...samples.map(s => s.memory.used)),
        total: Math.max(...samples.map(s => s.memory.total)),
        percentage: Math.max(...samples.map(s => s.memory.percentage))
      },
      network: {
        requestCount: Math.max(...samples.map(s => s.network.requestCount)),
        totalBytes: Math.max(...samples.map(s => s.network.totalBytes)),
        activeConnections: Math.max(...samples.map(s => s.network.activeConnections))
      },
      performance: {
        loadTime: Math.max(...samples.map(s => s.performance.loadTime)),
        renderTime: Math.max(...samples.map(s => s.performance.renderTime)),
        interactionReady: Math.max(...samples.map(s => s.performance.interactionReady))
      }
    };
    
    // Analyze trends (simple linear regression)
    const memoryValues = samples.map(s => s.memory.percentage);
    const memoryTrend = MetricsHelper.analyzeTrend(memoryValues);
    
    const loadTimeValues = samples.map(s => s.performance.loadTime);
    const performanceTrend = MetricsHelper.analyzeTrend(loadTimeValues, true); // Inverted - lower is better
    
    return {
      summary: {
        duration,
        sampleCount: samples.length,
        averages: averages as Partial<SystemMetrics>,
        peaks: peaks as Partial<SystemMetrics>,
        trends: {
          memoryTrend,
          performanceTrend
        }
      }
    };
  }

  /**
   * Analyze trend in values
   */
  private static analyzeTrend(
    values: number[], 
    inverted: boolean = false
  ): 'increasing' | 'decreasing' | 'stable' {
    if (values.length < 2) return 'stable';
    
    const firstHalf = values.slice(0, Math.floor(values.length / 2));
    const secondHalf = values.slice(Math.floor(values.length / 2));
    
    const firstAvg = firstHalf.reduce((sum, v) => sum + v, 0) / firstHalf.length;
    const secondAvg = secondHalf.reduce((sum, v) => sum + v, 0) / secondHalf.length;
    
    const threshold = 0.05; // 5% change threshold
    const change = (secondAvg - firstAvg) / firstAvg;
    
    if (inverted) {
      if (change > threshold) return 'degrading';
      if (change < -threshold) return 'improving';
    } else {
      if (change > threshold) return 'increasing';
      if (change < -threshold) return 'decreasing';
    }
    
    return 'stable';
  }
}