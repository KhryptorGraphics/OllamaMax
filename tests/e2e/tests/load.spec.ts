import { test, expect } from '@playwright/test';
import { LoadTestHelper } from '../helpers/load-test-helper';
import { MetricsHelper } from '../helpers/metrics-helper';

/**
 * Load Testing Suite for OllamaMax Platform
 * 
 * Tests system behavior under various load conditions:
 * - Concurrent user scenarios
 * - High-throughput API testing
 * - WebSocket stress testing
 * - Resource exhaustion scenarios
 * - Scalability validation
 * - Performance under pressure
 */

test.describe('Load Testing', () => {
  let loadHelper: LoadTestHelper;
  let metricsHelper: MetricsHelper;

  test.beforeEach(async ({ page }) => {
    loadHelper = new LoadTestHelper(page);
    metricsHelper = new MetricsHelper(page);
  });

  test('basic load test - 10 concurrent users', async ({ page, request }) => {
    const loadTest = await loadHelper.runConcurrentAPITest({
      concurrentUsers: 10,
      requestsPerUser: 20,
      endpoint: '/api/v1/health',
      method: 'GET',
      rampUpTime: 2000 // 2 second ramp-up
    });

    // Basic performance expectations
    expect(loadTest.totalRequests).toBe(200); // 10 users × 20 requests
    expect(loadTest.successfulRequests).toBeGreaterThan(180); // >90% success rate
    expect(loadTest.averageResponseTime).toBeLessThan(2000); // <2s average
    expect(loadTest.requestsPerSecond).toBeGreaterThan(10); // >10 RPS

    // Response time distribution
    expect(loadTest.responseTimePercentiles.p95).toBeLessThan(5000); // 95th percentile <5s
    expect(loadTest.responseTimePercentiles.p99).toBeLessThan(10000); // 99th percentile <10s

    console.log('Basic Load Test Results:', {
      successRate: `${Math.round((loadTest.successfulRequests / loadTest.totalRequests) * 100)}%`,
      avgResponseTime: `${loadTest.averageResponseTime}ms`,
      rps: loadTest.requestsPerSecond,
      p95: `${loadTest.responseTimePercentiles.p95}ms`
    });

    await loadHelper.saveResults('basic-load-test', loadTest);
  });

  test('moderate load test - 50 concurrent users', async ({ page, request }) => {
    const loadTest = await loadHelper.runConcurrentAPITest({
      concurrentUsers: 50,
      requestsPerUser: 10,
      endpoint: '/api/v1/health',
      method: 'GET',
      rampUpTime: 10000 // 10 second ramp-up
    });

    // Moderate load expectations
    expect(loadTest.totalRequests).toBe(500);
    expect(loadTest.successfulRequests).toBeGreaterThan(450); // >90% success rate
    expect(loadTest.averageResponseTime).toBeLessThan(5000); // <5s average under load
    expect(loadTest.requestsPerSecond).toBeGreaterThan(20); // >20 RPS

    // Check for errors
    const errorRate = (loadTest.failedRequests / loadTest.totalRequests) * 100;
    expect(errorRate).toBeLessThan(10); // <10% error rate

    console.log('Moderate Load Test Results:', {
      totalRequests: loadTest.totalRequests,
      successRate: `${Math.round((loadTest.successfulRequests / loadTest.totalRequests) * 100)}%`,
      errorRate: `${errorRate.toFixed(2)}%`,
      avgResponseTime: `${loadTest.averageResponseTime}ms`,
      maxResponseTime: `${loadTest.maxResponseTime}ms`,
      rps: loadTest.requestsPerSecond
    });

    await loadHelper.saveResults('moderate-load-test', loadTest);
  });

  test('stress test - 100 concurrent users', async ({ page, request }) => {
    const loadTest = await loadHelper.runConcurrentAPITest({
      concurrentUsers: 100,
      requestsPerUser: 5,
      endpoint: '/api/v1/health',
      method: 'GET',
      rampUpTime: 20000 // 20 second ramp-up
    });

    // Stress test expectations (more lenient)
    expect(loadTest.totalRequests).toBe(500);
    expect(loadTest.successfulRequests).toBeGreaterThan(400); // >80% success rate under stress
    expect(loadTest.averageResponseTime).toBeLessThan(10000); // <10s average under stress

    // System should not completely fail
    const successRate = (loadTest.successfulRequests / loadTest.totalRequests) * 100;
    expect(successRate).toBeGreaterThan(70); // At least 70% success rate

    console.log('Stress Test Results:', {
      concurrentUsers: 100,
      successRate: `${successRate.toFixed(1)}%`,
      avgResponseTime: `${loadTest.averageResponseTime}ms`,
      maxResponseTime: `${loadTest.maxResponseTime}ms`,
      rps: loadTest.requestsPerSecond
    });

    if (loadTest.errors.length > 0) {
      console.log('Errors encountered:', loadTest.errors.slice(0, 5)); // Show first 5 errors
    }

    await loadHelper.saveResults('stress-test', loadTest);
  });

  test('API endpoint endurance test', async ({ page, request }) => {
    // Test different API endpoints under sustained load
    const endpoints = [
      { path: '/api/v1/health', name: 'health' },
      { path: '/api/v1/status', name: 'status' },
      { path: '/', name: 'root' }
    ];

    const enduranceResults = [];

    for (const endpoint of endpoints) {
      console.log(`Testing endpoint: ${endpoint.path}`);
      
      const loadTest = await loadHelper.runConcurrentAPITest({
        concurrentUsers: 20,
        requestsPerUser: 15,
        endpoint: endpoint.path,
        method: 'GET',
        rampUpTime: 5000
      });

      enduranceResults.push({
        endpoint: endpoint.name,
        ...loadTest
      });

      // Endpoint-specific expectations
      expect(loadTest.successfulRequests).toBeGreaterThan(250); // >80% success
      expect(loadTest.averageResponseTime).toBeLessThan(3000); // <3s average

      console.log(`${endpoint.name} results:`, {
        successRate: `${Math.round((loadTest.successfulRequests / loadTest.totalRequests) * 100)}%`,
        avgTime: `${loadTest.averageResponseTime}ms`,
        rps: loadTest.requestsPerSecond
      });
    }

    await loadHelper.saveResults('endurance-test', enduranceResults);
  });

  test('WebSocket load testing', async ({ page }) => {
    const wsLoadTest = await loadHelper.testWebSocketLoad({
      concurrentConnections: 25,
      messagesPerConnection: 10,
      messageInterval: 500 // Send message every 500ms
    });

    // WebSocket performance expectations
    expect(wsLoadTest.successfulRequests).toBeGreaterThan(200); // >80% messages successful
    expect(wsLoadTest.averageResponseTime).toBeLessThan(1000); // <1s average response

    const successRate = (wsLoadTest.successfulRequests / wsLoadTest.totalRequests) * 100;
    expect(successRate).toBeGreaterThan(75); // At least 75% success rate

    console.log('WebSocket Load Test Results:', {
      connections: 25,
      totalMessages: wsLoadTest.totalRequests,
      successRate: `${successRate.toFixed(1)}%`,
      avgResponseTime: `${wsLoadTest.averageResponseTime}ms`,
      messagesPerSecond: wsLoadTest.requestsPerSecond
    });

    await loadHelper.saveResults('websocket-load-test', wsLoadTest);
  });

  test('payload size impact testing', async ({ page, request }) => {
    const payloadSizes = [
      100,    // 100 bytes - small
      1000,   // 1KB - medium
      10000,  // 10KB - large
      100000  // 100KB - very large
    ];

    const payloadResults = await loadHelper.testPayloadSizes('/api/v1/health', payloadSizes);

    expect(payloadResults.length).toBe(payloadSizes.length);

    // Analyze how payload size affects performance
    for (let i = 0; i < payloadResults.length; i++) {
      const result = payloadResults[i];
      const payloadSize = payloadSizes[i];

      console.log(`Payload ${payloadSize} bytes:`, {
        avgResponseTime: `${result.averageResponseTime}ms`,
        successRate: `${Math.round((result.successfulRequests / result.totalRequests) * 100)}%`,
        rps: result.requestsPerSecond
      });

      // Larger payloads should still be handled reasonably
      expect(result.averageResponseTime).toBeLessThan(5000 + (payloadSize / 1000)); // Scale with size
      expect(result.successfulRequests).toBeGreaterThan(40); // >80% success (5 users × 10 requests × 0.8)
    }

    await loadHelper.saveResults('payload-size-impact', payloadResults);
  });

  test('resource monitoring under load', async ({ page }) => {
    // Start resource monitoring
    const monitoringPromise = loadHelper.monitorResourceUsage(30000); // 30 seconds

    // Apply load while monitoring
    await page.goto('/');
    
    const loadTest = await loadHelper.runConcurrentAPITest({
      concurrentUsers: 30,
      requestsPerUser: 8,
      endpoint: '/api/v1/health',
      method: 'GET'
    });

    // Wait for monitoring to complete
    const resourceData = await monitoringPromise;

    expect(resourceData.length).toBeGreaterThan(10); // Should have multiple samples

    // Analyze resource usage patterns
    const memoryUsages = resourceData
      .map(sample => sample.memory?.usedJSHeapSize || 0)
      .filter(usage => usage > 0);

    if (memoryUsages.length > 0) {
      const maxMemory = Math.max(...memoryUsages);
      const avgMemory = memoryUsages.reduce((sum, val) => sum + val, 0) / memoryUsages.length;

      console.log('Resource Usage Under Load:', {
        loadTestSuccess: `${Math.round((loadTest.successfulRequests / loadTest.totalRequests) * 100)}%`,
        maxMemoryUsage: `${(maxMemory / 1024 / 1024).toFixed(1)}MB`,
        avgMemoryUsage: `${(avgMemory / 1024 / 1024).toFixed(1)}MB`,
        monitoringSamples: resourceData.length
      });

      // Memory usage should not exceed reasonable limits
      expect(maxMemory).toBeLessThan(200 * 1024 * 1024); // 200MB max
    }

    await loadHelper.saveResults('resource-monitoring-under-load', {
      loadTest,
      resourceData
    });
  });

  test('gradual load increase test', async ({ page, request }) => {
    // Test how the system handles gradually increasing load
    const loadLevels = [5, 10, 20, 30];
    const results = [];

    for (const userCount of loadLevels) {
      console.log(`Testing with ${userCount} concurrent users...`);

      const loadTest = await loadHelper.runConcurrentAPITest({
        concurrentUsers: userCount,
        requestsPerUser: 10,
        endpoint: '/api/v1/health',
        method: 'GET',
        rampUpTime: 3000
      });

      results.push({
        concurrentUsers: userCount,
        ...loadTest
      });

      const successRate = (loadTest.successfulRequests / loadTest.totalRequests) * 100;
      
      console.log(`${userCount} users:`, {
        successRate: `${successRate.toFixed(1)}%`,
        avgResponseTime: `${loadTest.averageResponseTime}ms`,
        rps: loadTest.requestsPerSecond
      });

      // Each level should maintain reasonable performance
      expect(successRate).toBeGreaterThan(85); // >85% success rate
      expect(loadTest.averageResponseTime).toBeLessThan(3000 + (userCount * 50)); // Scale with load

      // Small delay between load levels
      await page.waitForTimeout(5000);
    }

    // Analyze performance degradation pattern
    const responseTimeTrend = results.map(r => r.averageResponseTime);
    const throughputTrend = results.map(r => r.requestsPerSecond);

    console.log('Load Scaling Analysis:', {
      responseTimeTrend: responseTimeTrend.map(t => `${t}ms`),
      throughputTrend: throughputTrend.map(t => `${t}rps`),
      scalingFactor: (throughputTrend[throughputTrend.length - 1] / throughputTrend[0]).toFixed(2)
    });

    await loadHelper.saveResults('gradual-load-increase', results);
  });

  test('peak load recovery test', async ({ page, request }) => {
    // Test system recovery after peak load
    
    // Phase 1: Apply high load
    console.log('Phase 1: Applying peak load...');
    const peakLoadTest = await loadHelper.runConcurrentAPITest({
      concurrentUsers: 80,
      requestsPerUser: 5,
      endpoint: '/api/v1/health',
      method: 'GET'
    });

    // Phase 2: Cool down period
    console.log('Phase 2: Cool down period...');
    await page.waitForTimeout(10000); // 10 second recovery

    // Phase 3: Test normal load to verify recovery
    console.log('Phase 3: Testing recovery with normal load...');
    const recoveryTest = await loadHelper.runConcurrentAPITest({
      concurrentUsers: 10,
      requestsPerUser: 10,
      endpoint: '/api/v1/health',
      method: 'GET'
    });

    // Recovery expectations
    const recoverySuccessRate = (recoveryTest.successfulRequests / recoveryTest.totalRequests) * 100;
    expect(recoverySuccessRate).toBeGreaterThan(90); // Should recover to >90%
    expect(recoveryTest.averageResponseTime).toBeLessThan(2000); // Should be back to normal

    const peakSuccessRate = (peakLoadTest.successfulRequests / peakLoadTest.totalRequests) * 100;

    console.log('Peak Load Recovery Test:', {
      peakLoadSuccess: `${peakSuccessRate.toFixed(1)}%`,
      peakAvgTime: `${peakLoadTest.averageResponseTime}ms`,
      recoverySuccess: `${recoverySuccessRate.toFixed(1)}%`,
      recoveryAvgTime: `${recoveryTest.averageResponseTime}ms`,
      recoveryImprovement: `${((peakLoadTest.averageResponseTime - recoveryTest.averageResponseTime) / peakLoadTest.averageResponseTime * 100).toFixed(1)}%`
    });

    await loadHelper.saveResults('peak-load-recovery', {
      peakLoad: peakLoadTest,
      recovery: recoveryTest
    });
  });
});