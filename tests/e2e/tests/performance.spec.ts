import { test, expect } from '@playwright/test';
import { PerformanceHelper } from '../helpers/performance-helper';
import { LoadTestHelper } from '../helpers/load-test-helper';
import { MetricsHelper } from '../helpers/metrics-helper';

/**
 * Performance Testing Suite for OllamaMax Platform
 * 
 * Tests performance characteristics:
 * - Core Web Vitals measurement
 * - Memory usage patterns
 * - Network performance optimization
 * - Load testing scenarios
 * - Cross-device performance validation
 * - Resource consumption analysis
 */

test.describe('Performance Testing', () => {
  let performanceHelper: PerformanceHelper;
  let loadHelper: LoadTestHelper;
  let metricsHelper: MetricsHelper;

  test.beforeEach(async ({ page }) => {
    performanceHelper = new PerformanceHelper(page);
    loadHelper = new LoadTestHelper(page);
    metricsHelper = new MetricsHelper(page);
  });

  test('core web vitals measurement', async ({ page }) => {
    await page.goto('/', { waitUntil: 'networkidle' });
    
    const metrics = await performanceHelper.collectDetailedMetrics();
    
    // Core Web Vitals thresholds (Google recommendations)
    expect(metrics.firstContentfulPaint).toBeLessThan(2500); // 2.5s - Good
    expect(metrics.largestContentfulPaint).toBeLessThan(4000); // 4s - Needs improvement threshold
    expect(metrics.cumulativeLayoutShift).toBeLessThan(0.25); // 0.25 - Needs improvement threshold
    expect(metrics.firstInputDelay).toBeLessThan(300); // 300ms - Needs improvement threshold
    
    // Overall load time
    expect(metrics.loadTime).toBeLessThan(8000); // 8s maximum
    
    // Memory usage (if available)
    if (metrics.memoryUsage) {
      expect(metrics.memoryUsage.usedJSHeapSize).toBeLessThan(50 * 1024 * 1024); // 50MB
    }
    
    await performanceHelper.saveMetrics('core-web-vitals', metrics);
    
    console.log('Core Web Vitals Results:', {
      fcp: `${metrics.firstContentfulPaint}ms`,
      lcp: `${metrics.largestContentfulPaint}ms`,
      cls: metrics.cumulativeLayoutShift,
      fid: `${metrics.firstInputDelay}ms`,
      loadTime: `${metrics.loadTime}ms`
    });
  });

  test('memory usage monitoring', async ({ page }) => {
    await page.goto('/');
    
    // Monitor memory usage over time
    const memoryData = await performanceHelper.monitorMemoryUsage(10000); // 10 seconds
    
    expect(memoryData.length).toBeGreaterThan(5); // Should have multiple samples
    
    // Check for memory leaks (memory should not continuously increase)
    if (memoryData.length >= 3) {
      const firstThird = memoryData.slice(0, Math.floor(memoryData.length / 3));
      const lastThird = memoryData.slice(-Math.floor(memoryData.length / 3));
      
      const avgFirst = firstThird.reduce((sum, sample) => 
        sum + sample.memoryUsage.usedJSHeapSize, 0) / firstThird.length;
      const avgLast = lastThird.reduce((sum, sample) => 
        sum + sample.memoryUsage.usedJSHeapSize, 0) / lastThird.length;
      
      const memoryIncrease = (avgLast - avgFirst) / avgFirst;
      
      // Memory increase should not be more than 50% over 10 seconds
      expect(memoryIncrease).toBeLessThan(0.5);
      
      console.log(`Memory trend: ${(memoryIncrease * 100).toFixed(2)}% increase over monitoring period`);
    }
    
    await performanceHelper.saveMetrics('memory-monitoring', { memoryData });
  });

  test('network performance analysis', async ({ page }) => {
    await page.goto('/', { waitUntil: 'networkidle' });
    
    const resourceAnalysis = await performanceHelper.analyzeResourcePerformance();
    
    // Performance expectations
    expect(resourceAnalysis.averageLoadTime).toBeLessThan(2000); // 2s average
    expect(resourceAnalysis.totalSize).toBeLessThan(5 * 1024 * 1024); // 5MB total
    expect(resourceAnalysis.totalResources).toBeLessThan(100); // Reasonable resource count
    
    // Check for slow resources
    const slowResources = resourceAnalysis.slowestResources.filter(r => r.duration > 3000);
    
    if (slowResources.length > 0) {
      console.warn('Slow resources detected:', slowResources.map(r => ({
        url: r.url.split('/').pop(),
        duration: `${r.duration}ms`,
        size: `${(r.size / 1024).toFixed(1)}KB`
      })));
    }
    
    // Should not have more than 3 resources taking longer than 3 seconds
    expect(slowResources.length).toBeLessThan(3);
    
    console.log('Network Performance:', {
      totalResources: resourceAnalysis.totalResources,
      totalSize: `${(resourceAnalysis.totalSize / 1024).toFixed(1)}KB`,
      averageLoadTime: `${resourceAnalysis.averageLoadTime}ms`
    });
  });

  test('cross-device performance validation', async ({ page }) => {
    const devices = [
      { width: 375, height: 812, name: 'mobile' },
      { width: 768, height: 1024, name: 'tablet' },
      { width: 1920, height: 1080, name: 'desktop' }
    ];
    
    const devicePerformance = [];
    
    for (const device of devices) {
      await page.setViewportSize({ width: device.width, height: device.height });
      await page.reload({ waitUntil: 'networkidle' });
      
      const metrics = await performanceHelper.collectDetailedMetrics();
      
      devicePerformance.push({
        device: device.name,
        viewport: `${device.width}x${device.height}`,
        ...metrics
      });
      
      // Device-specific performance expectations
      if (device.name === 'mobile') {
        // Mobile should be optimized for slower connections
        expect(metrics.loadTime).toBeLessThan(12000); // 12s for mobile
        expect(metrics.firstContentfulPaint).toBeLessThan(4000); // 4s for mobile
      } else {
        expect(metrics.loadTime).toBeLessThan(8000); // 8s for tablet/desktop
        expect(metrics.firstContentfulPaint).toBeLessThan(3000); // 3s for tablet/desktop
      }
      
      console.log(`${device.name} performance:`, {
        loadTime: `${metrics.loadTime}ms`,
        fcp: `${metrics.firstContentfulPaint}ms`,
        lcp: `${metrics.largestContentfulPaint}ms`
      });
    }
    
    await performanceHelper.saveMetrics('cross-device-performance', devicePerformance);
  });

  test('concurrent user load simulation', async ({ page, request }) => {
    const loadTestResult = await loadHelper.runConcurrentAPITest({
      concurrentUsers: 10,
      requestsPerUser: 5,
      endpoint: '/api/v1/health',
      method: 'GET'
    });
    
    // Performance expectations under load
    expect(loadTestResult.successfulRequests).toBeGreaterThan(40); // At least 80% success rate
    expect(loadTestResult.averageResponseTime).toBeLessThan(2000); // 2s average response
    expect(loadTestResult.requestsPerSecond).toBeGreaterThan(5); // At least 5 RPS
    
    // Response time distribution
    expect(loadTestResult.responseTimePercentiles.p95).toBeLessThan(5000); // 95th percentile under 5s
    
    console.log('Load Test Results:', {
      totalRequests: loadTestResult.totalRequests,
      successRate: `${Math.round((loadTestResult.successfulRequests / loadTestResult.totalRequests) * 100)}%`,
      avgResponseTime: `${loadTestResult.averageResponseTime}ms`,
      rps: loadTestResult.requestsPerSecond,
      p95: `${loadTestResult.responseTimePercentiles.p95}ms`
    });
    
    await loadHelper.saveResults('concurrent-users', loadTestResult);
  });

  test('API endpoint performance benchmarking', async ({ page, request }) => {
    const endpoints = [
      { path: '/api/v1/health', name: 'health-check' },
      { path: '/api/v1/status', name: 'status-check' },
      { path: '/', name: 'main-page' }
    ];
    
    const benchmarkResults = [];
    
    for (const endpoint of endpoints) {
      const benchmark = await metricsHelper.benchmarkOperation(
        async () => {
          const response = await request.get(endpoint.path, { timeout: 10000 });
          return {
            status: response.status(),
            ok: response.ok(),
            headers: response.headers()
          };
        },
        endpoint.name,
        10 // 10 iterations
      );
      
      benchmarkResults.push(benchmark);
      
      // Performance expectations per endpoint
      expect(benchmark.averageTime).toBeLessThan(1500); // 1.5s max average
      expect(benchmark.maxTime).toBeLessThan(5000); // 5s max for any single request
      
      console.log(`${endpoint.name} benchmark:`, {
        avgTime: `${benchmark.averageTime}ms`,
        minTime: `${benchmark.minTime}ms`,
        maxTime: `${benchmark.maxTime}ms`,
        iterations: benchmark.iterations
      });
    }
    
    await metricsHelper.saveMetrics('api-benchmarks', benchmarkResults);
  });

  test('resource consumption under different network conditions', async ({ page }) => {
    const networkConditions = await performanceHelper.testNetworkConditions();
    
    for (const [condition, metrics] of Object.entries(networkConditions)) {
      console.log(`${condition} network performance:`, {
        loadTime: `${metrics.loadTime}ms`,
        fcp: `${metrics.firstContentfulPaint}ms`,
        totalSize: metrics.totalTransferSize ? `${(metrics.totalTransferSize / 1024).toFixed(1)}KB` : 'unknown'
      });
      
      // Network-specific expectations
      if (condition === 'fast3g') {
        expect(metrics.loadTime).toBeLessThan(15000); // 15s for fast 3G
      } else if (condition === 'slow3g') {
        expect(metrics.loadTime).toBeLessThan(30000); // 30s for slow 3G
      }
      
      // First Contentful Paint should be reasonable even on slow networks
      expect(metrics.firstContentfulPaint).toBeLessThan(10000); // 10s max
    }
    
    await performanceHelper.saveMetrics('network-conditions', networkConditions);
  });

  test('long-term performance monitoring', async ({ page }) => {
    // Extended monitoring session
    const monitoringResults = await metricsHelper.monitorMetrics(30000, 3000); // 30s, every 3s
    
    expect(monitoringResults.length).toBeGreaterThan(8); // Should have multiple samples
    
    // Analyze performance stability
    const summary = MetricsHelper.generateMetricsSummary(monitoringResults);
    
    // Performance should be relatively stable
    expect(summary.summary.trends.performanceTrend).not.toBe('degrading');
    
    // Memory should not continuously increase
    if (summary.summary.trends.memoryTrend === 'increasing') {
      console.warn('⚠️  Memory usage trending upward - potential memory leak');
    }
    
    console.log('Long-term monitoring summary:', {
      duration: `${summary.summary.duration / 1000}s`,
      samples: summary.summary.sampleCount,
      avgLoadTime: `${summary.summary.averages.performance?.loadTime || 'N/A'}ms`,
      memoryTrend: summary.summary.trends.memoryTrend,
      perfTrend: summary.summary.trends.performanceTrend
    });
    
    await metricsHelper.saveMetrics('long-term-monitoring', {
      samples: monitoringResults,
      summary
    });
  });

  test('performance regression detection', async ({ page }) => {
    // This test would typically compare against historical baselines
    // For now, we'll establish current performance as baseline
    
    const currentMetrics = await performanceHelper.measurePageLoadPerformance('/', 3);
    
    expect(currentMetrics.length).toBe(3);
    
    // Calculate performance statistics
    const summary = PerformanceHelper.generateSummary(currentMetrics);
    
    // Performance consistency checks
    const loadTimeVariation = summary.maximums.loadTime! - summary.minimums.loadTime!;
    const fcpVariation = summary.maximums.firstContentfulPaint! - summary.minimums.firstContentfulPaint!;
    
    // Performance should be consistent across runs
    expect(loadTimeVariation).toBeLessThan(5000); // Max 5s variation in load times
    expect(fcpVariation).toBeLessThan(2000); // Max 2s variation in FCP
    
    console.log('Performance consistency:', {
      loadTimeRange: `${summary.minimums.loadTime}-${summary.maximums.loadTime}ms`,
      fcpRange: `${summary.minimums.firstContentfulPaint}-${summary.maximums.firstContentfulPaint}ms`,
      avgLoadTime: `${summary.averages.loadTime}ms`,
      avgFCP: `${summary.averages.firstContentfulPaint}ms`
    });
    
    // Save as performance baseline for future regression testing
    await performanceHelper.saveMetrics('performance-baseline', {
      measurements: currentMetrics,
      summary,
      timestamp: new Date().toISOString(),
      baselineVersion: '1.0.0'
    });
  });
});