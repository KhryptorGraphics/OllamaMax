import { Page } from '@playwright/test';
import fs from 'fs/promises';
import path from 'path';

/**
 * Performance Testing Helper for OllamaMax Platform
 * 
 * Provides utilities for:
 * - Performance metrics collection
 * - Load time measurement
 * - Memory usage monitoring
 * - Network timing analysis
 * - Core Web Vitals tracking
 */

export interface PerformanceMetrics {
  loadTime: number;
  firstContentfulPaint: number;
  largestContentfulPaint: number;
  firstInputDelay: number;
  cumulativeLayoutShift: number;
  timeToInteractive: number;
  totalBlockingTime: number;
  memoryUsage?: {
    totalJSHeapSize: number;
    usedJSHeapSize: number;
    jsHeapSizeLimit: number;
  };
  networkTiming: {
    dns: number;
    tcp: number;
    request: number;
    response: number;
    domProcessing: number;
  };
  resourceCount: number;
  totalTransferSize: number;
}

export class PerformanceHelper {
  constructor(private page: Page) {}

  /**
   * Collect basic performance metrics
   */
  async collectMetrics(): Promise<Partial<PerformanceMetrics>> {
    return await this.page.evaluate(() => {
      const navigation = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
      const paint = performance.getEntriesByType('paint');
      
      const fcp = paint.find(p => p.name === 'first-contentful-paint')?.startTime || 0;
      const loadTime = navigation.loadEventEnd - navigation.fetchStart;
      
      return {
        loadTime: Math.round(loadTime),
        firstContentfulPaint: Math.round(fcp),
        resourceCount: performance.getEntriesByType('resource').length
      };
    });
  }

  /**
   * Collect comprehensive performance metrics including Core Web Vitals
   */
  async collectDetailedMetrics(): Promise<PerformanceMetrics> {
    const metrics = await this.page.evaluate(() => {
      return new Promise<PerformanceMetrics>((resolve) => {
        const navigation = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
        const paint = performance.getEntriesByType('paint');
        const resources = performance.getEntriesByType('resource');
        
        // Basic timing metrics
        const loadTime = navigation.loadEventEnd - navigation.fetchStart;
        const fcp = paint.find(p => p.name === 'first-contentful-paint')?.startTime || 0;
        
        // Network timing breakdown
        const networkTiming = {
          dns: navigation.domainLookupEnd - navigation.domainLookupStart,
          tcp: navigation.connectEnd - navigation.connectStart,
          request: navigation.responseStart - navigation.requestStart,
          response: navigation.responseEnd - navigation.responseStart,
          domProcessing: navigation.domContentLoadedEventEnd - navigation.responseEnd
        };
        
        // Resource analysis
        const totalTransferSize = resources.reduce((sum: number, resource: any) => {
          return sum + (resource.transferSize || 0);
        }, 0);
        
        // Memory usage (if available)
        const memoryInfo = (performance as any).memory;
        const memoryUsage = memoryInfo ? {
          totalJSHeapSize: memoryInfo.totalJSHeapSize,
          usedJSHeapSize: memoryInfo.usedJSHeapSize,
          jsHeapSizeLimit: memoryInfo.jsHeapSizeLimit
        } : undefined;
        
        // TTI estimation (simplified)
        const timeToInteractive = navigation.domInteractive - navigation.fetchStart;
        
        const baseMetrics = {
          loadTime: Math.round(loadTime),
          firstContentfulPaint: Math.round(fcp),
          largestContentfulPaint: 0, // Will be updated by LCP observer
          firstInputDelay: 0, // Will be updated by FID observer
          cumulativeLayoutShift: 0, // Will be updated by CLS observer
          timeToInteractive: Math.round(timeToInteractive),
          totalBlockingTime: 0, // Simplified - would need detailed calculation
          memoryUsage,
          networkTiming: {
            dns: Math.round(networkTiming.dns),
            tcp: Math.round(networkTiming.tcp),
            request: Math.round(networkTiming.request),
            response: Math.round(networkTiming.response),
            domProcessing: Math.round(networkTiming.domProcessing)
          },
          resourceCount: resources.length,
          totalTransferSize
        };
        
        // Set up Web Vitals observers
        let lcpValue = 0;
        let clsValue = 0;
        let fidValue = 0;
        let observersComplete = 0;
        const totalObservers = 3;
        
        const checkComplete = () => {
          observersComplete++;
          if (observersComplete >= totalObservers) {
            resolve({
              ...baseMetrics,
              largestContentfulPaint: Math.round(lcpValue),
              firstInputDelay: Math.round(fidValue),
              cumulativeLayoutShift: Math.round(clsValue * 1000) / 1000
            });
          }
        };
        
        // LCP Observer
        if ('PerformanceObserver' in window) {
          try {
            const lcpObserver = new PerformanceObserver((list) => {
              const entries = list.getEntries();
              if (entries.length > 0) {
                lcpValue = entries[entries.length - 1].startTime;
              }
            });
            lcpObserver.observe({ type: 'largest-contentful-paint', buffered: true });
            
            setTimeout(() => {
              lcpObserver.disconnect();
              checkComplete();
            }, 2000);
          } catch (e) {
            checkComplete();
          }
          
          // CLS Observer
          try {
            let clsScore = 0;
            const clsObserver = new PerformanceObserver((list) => {
              for (const entry of list.getEntries() as any[]) {
                if (entry.hadRecentInput) continue;
                clsScore += entry.value;
              }
              clsValue = clsScore;
            });
            clsObserver.observe({ type: 'layout-shift', buffered: true });
            
            setTimeout(() => {
              clsObserver.disconnect();
              checkComplete();
            }, 2000);
          } catch (e) {
            checkComplete();
          }
          
          // FID Observer
          try {
            const fidObserver = new PerformanceObserver((list) => {
              const entries = list.getEntries();
              if (entries.length > 0) {
                fidValue = (entries[0] as any).processingStart - entries[0].startTime;
              }
            });
            fidObserver.observe({ type: 'first-input', buffered: true });
            
            setTimeout(() => {
              fidObserver.disconnect();
              checkComplete();
            }, 2000);
          } catch (e) {
            checkComplete();
          }
        } else {
          // Fallback if PerformanceObserver is not available
          setTimeout(() => resolve(baseMetrics as PerformanceMetrics), 1000);
        }
      });
    });
    
    return metrics;
  }

  /**
   * Monitor memory usage over time
   */
  async monitorMemoryUsage(duration: number = 10000): Promise<Array<{ timestamp: number; memoryUsage: any }>> {
    const samples: Array<{ timestamp: number; memoryUsage: any }> = [];
    const interval = 1000; // Sample every second
    const endTime = Date.now() + duration;
    
    while (Date.now() < endTime) {
      const memoryUsage = await this.page.evaluate(() => {
        const memory = (performance as any).memory;
        return memory ? {
          totalJSHeapSize: memory.totalJSHeapSize,
          usedJSHeapSize: memory.usedJSHeapSize,
          jsHeapSizeLimit: memory.jsHeapSizeLimit
        } : null;
      });
      
      if (memoryUsage) {
        samples.push({
          timestamp: Date.now(),
          memoryUsage
        });
      }
      
      await this.page.waitForTimeout(interval);
    }
    
    return samples;
  }

  /**
   * Measure page load performance with multiple iterations
   */
  async measurePageLoadPerformance(url: string, iterations: number = 3): Promise<PerformanceMetrics[]> {
    const results: PerformanceMetrics[] = [];
    
    for (let i = 0; i < iterations; i++) {
      // Clear cache and cookies for clean test
      if (i === 0) {
        await this.page.context().clearCookies();
      }
      
      await this.page.goto(url, { waitUntil: 'networkidle' });
      const metrics = await this.collectDetailedMetrics();
      results.push(metrics);
      
      // Small delay between iterations
      await this.page.waitForTimeout(1000);
    }
    
    return results;
  }

  /**
   * Analyze resource loading performance
   */
  async analyzeResourcePerformance(): Promise<{
    slowestResources: Array<{ url: string; duration: number; size: number }>;
    totalResources: number;
    totalSize: number;
    averageLoadTime: number;
  }> {
    return await this.page.evaluate(() => {
      const resources = performance.getEntriesByType('resource') as PerformanceResourceTiming[];
      
      const resourceData = resources.map(resource => ({
        url: resource.name,
        duration: resource.responseEnd - resource.startTime,
        size: (resource as any).transferSize || 0
      }));
      
      const slowestResources = resourceData
        .sort((a, b) => b.duration - a.duration)
        .slice(0, 10);
      
      const totalSize = resourceData.reduce((sum, r) => sum + r.size, 0);
      const averageLoadTime = resourceData.length > 0 
        ? resourceData.reduce((sum, r) => sum + r.duration, 0) / resourceData.length 
        : 0;
      
      return {
        slowestResources,
        totalResources: resources.length,
        totalSize,
        averageLoadTime: Math.round(averageLoadTime)
      };
    });
  }

  /**
   * Test performance under different network conditions
   */
  async testNetworkConditions(): Promise<{ [condition: string]: PerformanceMetrics }> {
    const conditions = {
      'fast3g': { downloadThroughput: 1.5 * 1024 * 1024 / 8, uploadThroughput: 750 * 1024 / 8, latency: 40 },
      'slow3g': { downloadThroughput: 500 * 1024 / 8, uploadThroughput: 500 * 1024 / 8, latency: 400 },
      'offline': { downloadThroughput: 0, uploadThroughput: 0, latency: 0 }
    };
    
    const results: { [condition: string]: PerformanceMetrics } = {};
    
    for (const [name, condition] of Object.entries(conditions)) {
      if (name === 'offline') continue; // Skip offline test to avoid failures
      
      const client = await this.page.context().newCDPSession(this.page);
      
      // Set network conditions
      await client.send('Network.emulateNetworkConditions', {
        offline: false,
        downloadThroughput: condition.downloadThroughput,
        uploadThroughput: condition.uploadThroughput,
        latency: condition.latency
      });
      
      await this.page.reload({ waitUntil: 'networkidle' });
      const metrics = await this.collectDetailedMetrics();
      results[name] = metrics;
      
      // Reset network conditions
      await client.send('Network.emulateNetworkConditions', {
        offline: false,
        downloadThroughput: -1,
        uploadThroughput: -1,
        latency: 0
      });
      
      await client.detach();
    }
    
    return results;
  }

  /**
   * Save performance metrics to file
   */
  async saveMetrics(testName: string, metrics: PerformanceMetrics | PerformanceMetrics[], additionalData?: any): Promise<void> {
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
    const filename = `performance-${testName}-${timestamp}.json`;
    const filepath = path.join('reports', 'performance', filename);
    
    // Ensure directory exists
    await fs.mkdir(path.dirname(filepath), { recursive: true }).catch(() => {});
    
    const data = {
      testName,
      timestamp: new Date().toISOString(),
      userAgent: await this.page.evaluate(() => navigator.userAgent),
      viewport: await this.page.viewportSize(),
      url: this.page.url(),
      metrics,
      ...additionalData
    };
    
    await fs.writeFile(filepath, JSON.stringify(data, null, 2));
    console.log(`Performance metrics saved to: ${filepath}`);
  }

  /**
   * Generate performance report summary
   */
  static generateSummary(metricsArray: PerformanceMetrics[]): {
    averages: Partial<PerformanceMetrics>;
    minimums: Partial<PerformanceMetrics>;
    maximums: Partial<PerformanceMetrics>;
    medians: Partial<PerformanceMetrics>;
  } {
    if (metricsArray.length === 0) {
      throw new Error('No metrics provided for summary generation');
    }
    
    const keys = ['loadTime', 'firstContentfulPaint', 'largestContentfulPaint', 'firstInputDelay', 'cumulativeLayoutShift'] as const;
    
    const averages: any = {};
    const minimums: any = {};
    const maximums: any = {};
    const medians: any = {};
    
    for (const key of keys) {
      const values = metricsArray.map(m => m[key]).filter(v => v !== undefined && v > 0);
      
      if (values.length > 0) {
        averages[key] = Math.round(values.reduce((sum, val) => sum + val, 0) / values.length);
        minimums[key] = Math.min(...values);
        maximums[key] = Math.max(...values);
        
        const sorted = values.sort((a, b) => a - b);
        medians[key] = sorted[Math.floor(sorted.length / 2)];
      }
    }
    
    return { averages, minimums, maximums, medians };
  }
}